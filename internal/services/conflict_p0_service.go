package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var allowedConflictReviewDecisions = map[string]bool{
	"no_conflict":              true,
	"confirmed_conflict":       true,
	"false_positive":           true,
	"insufficient_information": true,
	"waiver_requested":         true,
}

type ConflictReviewService interface {
	ReviewConflict(ctx context.Context, checkID, decision, notes string, reviewerID uint, reviewerName string, nextReviewAt *time.Time) (*models.ConflictReview, error)
	GetConflictReview(ctx context.Context, checkID string) (*models.ConflictReview, error)
}

func normalizeLegalSubjectName(name, entityType string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(
		"（", "(", "）", ")", "【", "", "】", "", "[", "", "]", "",
		"·", "", "•", "", "-", "", "_", "", ".", "", "，", "", ",", "",
		" ", "", "　", "", "\t", "", "\n", "",
	)
	name = replacer.Replace(name)
	if strings.EqualFold(entityType, "PERSON") {
		return name
	}
	for _, suffix := range []string{
		"股份有限公司", "有限责任公司", "集团有限公司", "有限公司", "律师事务所",
		"集团", "控股", "公司", "律所",
	} {
		name = strings.TrimSuffix(name, suffix)
	}
	return strings.TrimSpace(name)
}

const (
	conflictSubjectRoleClient        = "CLIENT"
	conflictSubjectRoleOpposingParty = "OPPOSING_PARTY"
	conflictSubjectRoleRelatedParty  = "RELATED_PARTY"

	conflictSubjectAuthorityOtherParty    = 1
	conflictSubjectAuthorityExplicitParty = 2
	conflictSubjectAuthorityClient        = 3
)

func normalizeConflictSubjectRole(role string) string {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case conflictSubjectRoleClient:
		return conflictSubjectRoleClient
	case conflictSubjectRoleOpposingParty, "OPPOSING", "ADVERSE":
		return conflictSubjectRoleOpposingParty
	case conflictSubjectRoleRelatedParty:
		return conflictSubjectRoleRelatedParty
	default:
		return conflictSubjectRoleRelatedParty
	}
}

func normalizeConflictSubjectEntityType(entityType string) string {
	switch strings.ToUpper(strings.TrimSpace(entityType)) {
	case "PERSON":
		return "PERSON"
	case "COMPANY":
		return "COMPANY"
	default:
		return "ANY"
	}
}

func conflictSubjectRolePriority(role string) int {
	switch role {
	case conflictSubjectRoleClient:
		return 3
	case conflictSubjectRoleOpposingParty:
		return 2
	default:
		return 1
	}
}

func normalizeConflictSubjects(request *models.ConflictCheckRequest) []models.ConflictNormalizedSubject {
	subjects := make([]models.ConflictNormalizedSubject, 0, len(request.Parties)+1)
	identityIndex := map[string]int{}
	subjectAuthority := make([]int, 0, len(request.Parties)+1)
	appendSubject := func(name, role, entityType string, identifiers map[string]string, aliases []string, authority int) {
		role = normalizeConflictSubjectRole(role)
		entityType = normalizeConflictSubjectEntityType(entityType)
		normalized := normalizeLegalSubjectName(name, entityType)
		if normalized == "" {
			return
		}

		cleanIdentifiers := map[string]string{}
		identityKeys := []string{"name:" + normalized}
		for kind, value := range identifiers {
			cleanKind := strings.ToLower(strings.TrimSpace(kind))
			cleanValue := strings.TrimSpace(value)
			if cleanKind != "" && cleanValue != "" {
				cleanIdentifiers[cleanKind] = cleanValue
				identityKeys = append(identityKeys, "id:"+cleanKind+":"+strings.ToLower(cleanValue))
			}
		}
		cleanAliases := make([]string, 0, len(aliases))
		seenAliases := map[string]struct{}{}
		for _, alias := range aliases {
			if cleaned := normalizeLegalSubjectName(alias, entityType); cleaned != "" && cleaned != normalized {
				if _, exists := seenAliases[cleaned]; exists {
					continue
				}
				seenAliases[cleaned] = struct{}{}
				cleanAliases = append(cleanAliases, cleaned)
				identityKeys = append(identityKeys, "name:"+cleaned)
			}
		}

		existingIndex := -1
		for _, key := range identityKeys {
			if index, exists := identityIndex[key]; exists {
				existingIndex = index
				break
			}
		}
		if existingIndex >= 0 {
			existing := &subjects[existingIndex]
			if authority > subjectAuthority[existingIndex] ||
				(authority == subjectAuthority[existingIndex] && conflictSubjectRolePriority(role) > conflictSubjectRolePriority(existing.Role)) {
				existing.Role = role
				subjectAuthority[existingIndex] = authority
			}
			if existing.EntityType == "ANY" && entityType != "ANY" {
				existing.EntityType = entityType
			}
			for kind, value := range cleanIdentifiers {
				if _, exists := existing.Identifiers[kind]; !exists {
					existing.Identifiers[kind] = value
				}
			}
			for _, alias := range cleanAliases {
				if alias != existing.NormalizedName && !containsString(existing.Aliases, alias) {
					existing.Aliases = append(existing.Aliases, alias)
				}
			}
			for _, key := range identityKeys {
				identityIndex[key] = existingIndex
			}
			return
		}

		index := len(subjects)
		subjects = append(subjects, models.ConflictNormalizedSubject{
			OriginalName: strings.TrimSpace(name), NormalizedName: normalized, Role: role,
			EntityType: entityType, Identifiers: cleanIdentifiers, Aliases: cleanAliases,
		})
		subjectAuthority = append(subjectAuthority, authority)
		for _, key := range identityKeys {
			identityIndex[key] = index
		}
	}
	appendSubject(request.ClientName, conflictSubjectRoleClient, request.ClientType, request.ClientIdentifiers, request.ClientAliases, conflictSubjectAuthorityClient)
	for _, party := range request.Parties {
		appendSubject(party.Name, party.Role, party.EntityType, party.Identifiers, party.Aliases, conflictSubjectAuthorityExplicitParty)
	}
	for _, party := range request.OtherParties {
		appendSubject(party, conflictSubjectRoleOpposingParty, "ANY", nil, nil, conflictSubjectAuthorityOtherParty)
	}
	return subjects
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// conflictCoverageStatus is intentionally fail-closed. A conflict result is
// not eligible for a final "can continue" disposition until every active
// archive scope has been explicitly marked COMPLETE by the law firm. Missing
// scope metadata, an unavailable table, or any incomplete source is therefore
// surfaced as COVERAGE_LIMITED rather than silently treated as a clean search.
func (s *conflictDetectionService) conflictCoverageStatus(ctx context.Context) string {
	if s.caseRepo == nil || s.caseRepo.GetDB() == nil {
		return "COVERAGE_LIMITED"
	}
	if err := NewConflictScopeService(s.caseRepo.GetDB()).ValidateComplete(ctx); err != nil {
		return "COVERAGE_LIMITED"
	}
	return "COMPLETE"
}

func (s *conflictDetectionService) applyP0ConflictPolicy(ctx context.Context, request *models.ConflictCheckRequest, conflicts []*models.ConflictCase) (*models.ConflictDecisionSummary, *models.RiskAssessment) {
	ruleCodes := map[string]struct{}{}
	restrictedCount := 0
	hasExactNameCandidate := false
	evidenceCount := 0
	coverageStatus := s.conflictCoverageStatus(ctx)

	for _, conflict := range conflicts {
		if conflict == nil {
			continue
		}
		s.ensureConflictEvidence(conflict)
		if s.redactEthicalWallEvidence(ctx, request.ActorUserID, conflict) {
			restrictedCount++
		}
		if conflict.RuleCode == "" {
			conflict.RuleCode = "P0_REVIEW_REQUIRED"
		}
		ruleCodes[conflict.RuleCode] = struct{}{}
		evidenceCount += len(conflict.Evidence)
		// The current MVP has no authoritative, historical identity registry.
		// Name evidence, including a normalized exact name, therefore pauses the
		// intake for verification instead of confirming legal-entity identity.
		conflict.RequiresManualReview = true
		if conflict.MatchType == "EXACT" && (conflict.RuleCode == "DIRECT_ADVERSE_CURRENT_CLIENT" || conflict.RuleCode == "STRUCTURED_IDENTITY_EXACT") {
			hasExactNameCandidate = true
			conflict.RiskLevel = "HIGH"
		} else {
			conflict.RiskLevel = "MEDIUM"
		}
	}

	codes := make([]string, 0, len(ruleCodes))
	for code := range ruleCodes {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	decision := &models.ConflictDecisionSummary{
		Status: "REVIEW_REQUIRED", Recommendation: "未发现系统已导入档案中的匹配记录；档案覆盖范围和主体身份仍须由独立核查人确认，确认前不得视为无冲突。",
		RuleCodes: codes, EvidenceCount: evidenceCount, RestrictedCount: restrictedCount, RequiresManualReview: true,
		CoverageStatus: coverageStatus,
		CoverageNotice: "本次检索覆盖系统中已导入的内部案件、客户关系和结构化主体信息；权威档案来源、未导入历史档案及关联信息完整性尚未由律所配置确认。",
	}
	if coverageStatus == "COMPLETE" {
		decision.CoverageNotice = "本次检索覆盖律所已登记的全部有效冲突检索来源；仍需独立人工复核后形成处置结论。"
	}
	assessment := &models.RiskAssessment{
		OverallRisk: "LOW", RiskScore: 0, RiskReason: "未发现系统已导入档案中的匹配记录，待确认档案覆盖范围",
		RequiresApproval: true, ApprovalLevel: "CONFLICT_OFFICER", RiskFactors: []string{"权威档案覆盖范围待确认"},
		Mitigation: []string{"由独立核查人确认档案来源、主体身份标识和关联方信息"},
	}
	if len(conflicts) > 0 {
		decision.Status = "REVIEW_REQUIRED"
		decision.Recommendation = "发现名称、关联关系或文本线索，复核完成前不得将其视为已确认冲突或无冲突。"
		decision.RequiresManualReview = true
		assessment = &models.RiskAssessment{
			OverallRisk: "MEDIUM", RiskScore: 60, RiskReason: "存在待核实主体或关系证据，必须人工复核",
			RequiresApproval: true, ApprovalLevel: "COMPLIANCE", RiskFactors: []string{"存在待核实证据"},
			Mitigation: []string{"由合规人员核实主体身份、历史角色和保密信息接触范围"},
		}
	}
	if hasExactNameCandidate {
		assessment = &models.RiskAssessment{
			OverallRisk: "HIGH", RiskScore: 75, RiskReason: "名称规范化后完全一致，主体身份和历史委托范围必须由独立核查人确认",
			RequiresApproval: true, ApprovalLevel: "CONFLICT_OFFICER", RiskFactors: []string{"名称规范化后完全一致"},
			Mitigation: []string{"暂停接案提交", "核验统一社会信用代码或身份证件号", "核实历史代理范围和保密信息接触情况"},
		}
	}
	return decision, assessment
}

func (s *conflictDetectionService) ensureConflictEvidence(conflict *models.ConflictCase) {
	if len(conflict.Evidence) > 0 {
		return
	}
	matchType := conflict.MatchType
	if matchType == "" {
		switch {
		case strings.Contains(conflict.ConflictType, "直接冲突"):
			matchType = "EXACT"
		case strings.Contains(conflict.ConflictType, "关系"):
			matchType = "RELATION"
		case strings.Contains(conflict.ConflictType, "相似"):
			matchType = "CANDIDATE"
		default:
			matchType = "TEXT"
		}
	}
	conflict.MatchType = matchType
	if conflict.RuleCode == "" {
		if matchType == "EXACT" {
			conflict.RuleCode = "DIRECT_ADVERSE_CURRENT_CLIENT"
		} else {
			conflict.RuleCode = "SUBJECT_CANDIDATE_REVIEW"
		}
	}
	requestedParty := ""
	if len(conflict.OpposingParties) > 0 {
		requestedParty = conflict.OpposingParties[0]
	}
	conflict.Evidence = []models.ConflictEvidence{{
		EvidenceID: fmt.Sprintf("EV-%s", conflict.ID), RuleCode: conflict.RuleCode, MatchType: matchType,
		SourceType: "INTERNAL_CASE", RequestedParty: requestedParty, MatchedEntity: requestedParty,
		PartyRole: "OPPOSING_PARTY", HistoricalRole: "CLIENT", SourceCaseID: conflict.CaseID,
		SourceCaseNumber: conflict.CaseNo, SourceCaseName: conflict.CaseName, SourceUpdatedAt: conflict.CreatedAt,
		Summary: conflict.Description,
	}}
}

func (s *conflictDetectionService) redactEthicalWallEvidence(ctx context.Context, actorUserID uint, conflict *models.ConflictCase) bool {
	caseID, err := strconv.ParseUint(strings.TrimSpace(conflict.CaseID), 10, 32)
	if err != nil || caseID == 0 {
		return false
	}
	db := s.caseRepo.GetDB().WithContext(ctx)
	var enabled bool
	if err := db.Table("cases").Select("ethical_wall_enabled").Where("id = ?", uint(caseID)).Scan(&enabled).Error; err != nil || !enabled {
		return false
	}
	var count int64
	if actorUserID > 0 {
		_ = db.Table("case_ethical_wall_whitelist").Where("case_id = ? AND user_id = ?", uint(caseID), actorUserID).Count(&count).Error
	}
	if count > 0 {
		return false
	}
	// Preserve the authoritative evidence in the audit record. Rendering-time
	// projection decides whether the current viewer can see it; redacting here
	// would irreversibly remove the facts needed by an independent reviewer.
	conflict.Restricted = true
	for i := range conflict.Evidence {
		conflict.Evidence[i].Restricted = true
	}
	return true
}

func conflictEvidenceHash(response *models.ConflictCheckResponse) string {
	type hashEvidence struct {
		Rule, Match, Case, Party, Summary string
		Restricted                        bool
	}
	items := make([]hashEvidence, 0)
	for _, conflict := range response.ConflictCases {
		if conflict == nil {
			continue
		}
		for _, evidence := range conflict.Evidence {
			items = append(items, hashEvidence{evidence.RuleCode, evidence.MatchType, evidence.SourceCaseNumber, evidence.RequestedParty, evidence.Summary, evidence.Restricted})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		left, _ := json.Marshal(items[i])
		right, _ := json.Marshal(items[j])
		return string(left) < string(right)
	})
	raw, _ := json.Marshal(items)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (s *conflictDetectionService) ReviewConflict(ctx context.Context, checkID, decision, notes string, reviewerID uint, reviewerName string, nextReviewAt *time.Time) (*models.ConflictReview, error) {
	decision = strings.ToLower(strings.TrimSpace(decision))
	if !allowedConflictReviewDecisions[decision] {
		return nil, fmt.Errorf("不支持的复核结论")
	}
	if strings.TrimSpace(notes) == "" {
		return nil, fmt.Errorf("复核意见不能为空")
	}
	db := s.caseRepo.GetDB()
	if db == nil {
		return nil, &ConflictReviewerError{Code: "REVIEWER_GATE_UNAVAILABLE", Message: "冲突复核服务未初始化，已阻止提交"}
	}
	var review models.ConflictReview
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record models.ConflictCheckRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("check_id = ?", checkID).First(&record).Error; err != nil {
			return err
		}
		if record.UserID == reviewerID {
			return fmt.Errorf("申请律师或案件负责人不得自行复核本人的冲突检测")
		}
		if strings.ToUpper(strings.TrimSpace(record.CheckStatus)) != "COMPLETED" {
			return &ConflictReviewerError{Code: "REVIEW_REQUIRED", Message: "冲突检测尚未完成，不能提交人工复核结论"}
		}
		if err := ValidateConflictReviewer(ctx, tx, checkID, 0, reviewerID, ""); err != nil {
			return err
		}
		var response models.ConflictCheckResponse
		raw, _ := json.Marshal(record.CheckResult)
		if err := json.Unmarshal(raw, &response); err != nil {
			return newConflictReviewerError("REVIEW_EVIDENCE_INVALID", "冲突检测证据无法解析，已阻止形成复核结论")
		}
		if response.Decision == nil || !strings.EqualFold(strings.TrimSpace(response.Decision.CoverageStatus), "COMPLETE") {
			return newConflictReviewerError("COVERAGE_LIMITED", "冲突检索范围尚未完成律所确认，不能形成无冲突或豁免复核结论")
		}
		var existing models.ConflictReview
		if err := tx.Where("check_id = ?", checkID).Order("created_at DESC").First(&existing).Error; err == nil {
			return &ConflictReviewerError{Code: "REVIEW_ALREADY_EXISTS", Message: "该检测已有复核结论，不能重复提交；如证据有新版本，请重新运行冲突检测"}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		review = models.ConflictReview{
			CheckID: checkID, Decision: decision, Notes: strings.TrimSpace(notes), ReviewerID: reviewerID,
			ReviewerName: reviewerName, EvidenceHash: conflictEvidenceHash(&response), NextReviewAt: nextReviewAt, CreatedAt: time.Now(),
		}
		if err := tx.Create(&review).Error; err != nil {
			return fmt.Errorf("保存冲突复核失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// The original check evidence is immutable. The review is stored in its
	// own append-only row and is read through GetConflictReview; do not rewrite
	// the frozen check-result JSON merely to denormalize the review summary.
	return &review, nil
}

func (s *conflictDetectionService) GetConflictReview(ctx context.Context, checkID string) (*models.ConflictReview, error) {
	var reviews []models.ConflictReview
	err := s.caseRepo.GetDB().WithContext(ctx).
		Where("check_id = ?", checkID).
		Order("created_at DESC").
		Limit(1).
		Find(&reviews).Error
	if err != nil {
		return nil, err
	}
	if len(reviews) == 0 {
		return nil, nil
	}
	return &reviews[0], nil
}
