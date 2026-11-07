package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func submissionText(v interface{}) string {
	if v == nil {
		return ""
	}
	if b, ok := v.([]byte); ok {
		return strings.TrimSpace(string(b))
	}
	return strings.TrimSpace(fmt.Sprint(v))
}
func submissionMap(v interface{}) map[string]interface{} {
	var raw []byte
	switch value := v.(type) {
	case string:
		raw = []byte(value)
	case []byte:
		raw = value
	default:
		raw, _ = json.Marshal(v)
	}
	var result map[string]interface{}
	_ = json.Unmarshal(raw, &result)
	return result
}
func submissionMetadata(raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}, nil
	}
	var result map[string]interface{}
	if json.Unmarshal([]byte(raw), &result) != nil || result == nil {
		return nil, snapshotUnresolved("审批元数据无法解析")
	}
	return result, nil
}
func submissionRevision(id string, metadata map[string]interface{}) string {
	if s := submissionText(metadata["draft_revision"]); s != "" {
		return s
	}
	return "legacy-" + id
}
func submissionWithoutEnvelope(metadata map[string]interface{}) map[string]interface{} {
	result := submissionMap(metadata)
	for _, key := range []string{"approval_snapshot", "approval_snapshot_version", "approval_snapshot_created_at"} {
		delete(result, key)
	}
	return result
}

// ResolveApprovalSubmission is the sole authority resolver. No read writes to
// approval_requests or snapshots. New rounds MUST follow the explicit pointer.
// Legacy intake rows require agreement of embedded and persisted evidence;
// timestamps cannot resolve conflicting historical sources.
func ResolveApprovalSubmission(db *gorm.DB, approval *models.ApprovalRequest) (map[string]interface{}, error) {
	if approval == nil {
		return nil, snapshotUnresolved("审批不存在")
	}
	metadata, err := submissionMetadata(approval.Metadata)
	if err != nil {
		return nil, err
	}
	intakeID := submissionText(metadata["intake_id"])
	if approval.ActiveSubmissionSnapshotID != nil {
		if db == nil || strings.TrimSpace(*approval.ActiveSubmissionSnapshotID) == "" {
			return nil, snapshotUnresolved("当前提交快照指针无效")
		}
		var row models.ApprovalSnapshot
		if err := db.Where("id = ? AND approval_request_id = ? AND snapshot_type = ?", *approval.ActiveSubmissionSnapshotID, approval.ID, models.IntakeSubmissionSnapshotType).First(&row).Error; err != nil {
			return nil, fmt.Errorf("%w: %v", snapshotUnresolved("无法读取当前提交快照"), err)
		}
		snapshot := submissionMap(row.SnapshotData)
		frozen := submissionMap(snapshot["metadata"])
		if snapshot == nil || frozen == nil || submissionText(snapshot["snapshot_id"]) != row.ID || submissionText(snapshot["approval_request_id"]) != approval.ID || submissionText(snapshot["source_version"]) != strconv.Itoa(row.SourceVersion) || row.SourceVersion < 1 || submissionText(snapshot["intake_id"]) == "" || submissionText(snapshot["intake_id"]) != submissionText(frozen["intake_id"]) || (intakeID != "" && intakeID != submissionText(frozen["intake_id"])) || submissionText(frozen["conflict_check_id"]) == "" || submissionText(snapshot["conflict_check_id"]) != submissionText(frozen["conflict_check_id"]) {
			return nil, snapshotUnresolved("当前提交快照的审批、接案或轮次绑定不完整")
		}
		return snapshot, nil
	}
	embedded := submissionMap(metadata["approval_snapshot"])
	// Non-intake legacy approvals may have one persisted conflict snapshot.
	// Missing legacy tables are compatible only outside the intake protocol.
	if db == nil || (intakeID == "" && !db.Migrator().HasTable(&models.ApprovalSnapshot{})) {
		if intakeID != "" {
			return nil, snapshotUnresolved("接案快照服务未初始化")
		}
		if embedded != nil {
			return embedded, nil
		}
		return map[string]interface{}{"metadata": metadata}, nil
	}
	var rows []models.ApprovalSnapshot
	if err := db.Where("approval_request_id = ?", approval.ID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("读取历史提交快照失败: %w", err)
	}
	if len(rows) == 0 {
		if intakeID == "" {
			if embedded != nil {
				return embedded, nil
			}
			return map[string]interface{}{"metadata": metadata}, nil
		}
		return nil, snapshotUnresolved("缺少可核验的历史持久快照")
	}
	var authority map[string]interface{}
	for _, row := range rows {
		if row.SnapshotType == models.IntakeSubmissionSnapshotType {
			return nil, snapshotUnresolved("存在新轮次但当前提交指针缺失")
		}
		candidate := submissionMap(row.SnapshotData)
		if candidate == nil {
			return nil, snapshotUnresolved("历史快照无法解析")
		}
		if embedded != nil && !reflect.DeepEqual(embedded, candidate) {
			return nil, snapshotUnresolved("历史内嵌快照与持久快照不一致")
		}
		if authority != nil && !reflect.DeepEqual(authority, candidate) {
			return nil, snapshotUnresolved("多个历史快照来源冲突")
		}
		frozen := submissionMap(candidate["metadata"])
		if intakeID != "" && (frozen == nil || !reflect.DeepEqual(submissionWithoutEnvelope(frozen), submissionWithoutEnvelope(metadata))) {
			return nil, snapshotUnresolved("历史审批元数据与持久快照事实不一致")
		}
		authority = candidate
	}
	result := submissionMap(authority)
	// Use the persisted UUID as the compatibility version token without changing
	// the embedded legacy SNAP_* identifier in either stored source.
	result["snapshot_id"] = rows[0].ID
	result["source_version"] = rows[0].SourceVersion
	result["legacy_compatible"] = true
	return result, nil
}

// ApplyApprovalSubmission projects authority only in memory. Save callers must
// preserve original Metadata or explicitly omit it from their update.
func ApplyApprovalSubmission(db *gorm.DB, approval *models.ApprovalRequest) error {
	snapshot, err := ResolveApprovalSubmission(db, approval)
	if err != nil {
		return err
	}
	original, err := submissionMetadata(approval.Metadata)
	if err != nil {
		return err
	}
	if approval.ActiveSubmissionSnapshotID == nil {
		if submissionText(original["intake_id"]) != "" {
			projectSubmissionConflict(approval, snapshot)
		}
		return nil
	}
	metadata := submissionMap(snapshot["metadata"])
	if metadata == nil {
		return snapshotUnresolved("提交快照缺少成案事实")
	}
	metadata["approval_snapshot"] = snapshot
	metadata["approval_snapshot_version"] = snapshot["source_version"]
	metadata["approval_snapshot_created_at"] = snapshot["snapshot_at"]
	raw, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	approval.Metadata = string(raw)
	projectSubmissionConflict(approval, snapshot)
	return nil
}
func projectSubmissionConflict(approval *models.ApprovalRequest, snapshot map[string]interface{}) {
	metadata := submissionMap(snapshot["metadata"])
	approval.ConflictCheckID = submissionText(metadata["conflict_check_id"])
	record := submissionMap(metadata["conflict_record"])
	approval.ConflictRiskLevel = submissionText(record["risk_level"])
	if stamp, err := time.Parse(time.RFC3339Nano, submissionText(record["check_time"])); err == nil {
		approval.ConflictCheckTime = &stamp
	}
	if result, ok := metadata["conflict_result"]; ok {
		raw, _ := json.Marshal(result)
		approval.ConflictResult = string(raw)
	}
}

type submissionIntakeRow struct {
	ID          string
	IntakeCode  string
	ClientID    string
	Title       string
	CaseType    string
	Status      string
	Priority    string
	Description string
	Metadata    string
	CreatedBy   string
}

func lockSubmissionIntake(tx *gorm.DB, intakeID, userID string) (*submissionIntakeRow, map[string]interface{}, error) {
	var intake submissionIntakeRow
	if err := tx.Table("case_intakes").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", intakeID).Take(&intake).Error; err != nil {
		return nil, nil, err
	}
	if intake.CreatedBy != userID {
		return nil, nil, intakeCollaborationError("FORBIDDEN", "只有申请律师可以补正和提交该接案", 403)
	}
	var actor models.User
	if err := tx.Select("id, role, status").Where("id = ? AND deleted_at IS NULL", userID).First(&actor).Error; err != nil {
		return nil, nil, err
	}
	if !strings.EqualFold(actor.Status, "active") || (actor.Role != "lawyer" && !IsBusinessMatterManagementRole(actor.Role)) {
		return nil, nil, intakeCollaborationError("FORBIDDEN", "当前账号没有接案提交权限", 403)
	}
	metadata, err := submissionMetadata(intake.Metadata)
	return &intake, metadata, err
}

// ValidateResubmissionReplayInTx runs after the caller locks/authorizes approval,
// before changing status, rebinding tracking or invoking Freeze. A true result
// returns the already-created round; do not issue another transition or inbox write.
func ValidateResubmissionReplayInTx(tx *gorm.DB, approval *models.ApprovalRequest, req *models.ResubmitApprovalRequest) (bool, error) {
	if req == nil {
		return false, nil
	}
	if tx == nil || approval == nil {
		return false, errors.New("重提事务未初始化")
	}
	if strings.TrimSpace(req.ReturnRecordID) == "" {
		return false, supplementInvalid("重提必须指定退回记录")
	}
	fingerprint, err := resubmissionFingerprint(req)
	if err != nil {
		return false, err
	}
	var rows []models.ApprovalSnapshot
	if err := tx.Where("approval_request_id = ? AND snapshot_type = ?", approval.ID, models.IntakeSubmissionSnapshotType).Find(&rows).Error; err != nil {
		return false, err
	}
	for _, row := range rows {
		snapshot := submissionMap(row.SnapshotData)
		if submissionText(snapshot["return_record_id"]) != strings.TrimSpace(req.ReturnRecordID) {
			continue
		}
		if approval.ActiveSubmissionSnapshotID == nil || *approval.ActiveSubmissionSnapshotID != row.ID || submissionText(snapshot["resubmission_fingerprint"]) != fingerprint {
			return false, submissionStale("该退回轮次已提交其他版本，请重新加载")
		}
		if _, err := ResolveApprovalSubmission(tx, approval); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
func resubmissionFingerprint(req *models.ResubmitApprovalRequest) (string, error) {
	normalized := *req
	normalized.ExpectedIntakeRevision = strings.TrimSpace(req.ExpectedIntakeRevision)
	normalized.ExpectedSubmissionSnapshotID = strings.TrimSpace(req.ExpectedSubmissionSnapshotID)
	normalized.ReturnRecordID = strings.TrimSpace(req.ReturnRecordID)
	normalized.RevisionNote = strings.TrimSpace(req.RevisionNote)
	if normalized.ExpectedIntakeRevision == "" || normalized.ExpectedSubmissionSnapshotID == "" || normalized.ReturnRecordID == "" || !plainSupplementText(normalized.RevisionNote, 5000) {
		return "", supplementInvalid("重提必须提供草稿版本、提交版本、退回记录及修改说明")
	}
	items := make([]models.SupplementRequest, len(req.Responses))
	for i, r := range req.Responses {
		items[i].ItemID = strings.TrimSpace(r.ItemID)
	}
	responses, err := normalizeSupplementResponses(items, req.Responses, false)
	if err != nil {
		return "", err
	}
	normalized.Responses = responses
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:]), nil
}

// FreezeIntakeSubmissionInTx assumes the approval row is already locked (or was
// just inserted by this transaction). It acquires intake second; the caller then
// claims tracking. It never commits, updates old snapshots, saves approval, or
// changes its status. A returned error MUST abort the enclosing transaction.
func FreezeIntakeSubmissionInTx(tx *gorm.DB, approval *models.ApprovalRequest, userID string, req *models.ResubmitApprovalRequest) error {
	if approval == nil {
		return snapshotUnresolved("审批不存在")
	}
	oldMetadata, err := submissionMetadata(approval.Metadata)
	if err != nil {
		return err
	}
	intakeID := submissionText(oldMetadata["intake_id"])
	boundIntakeID, err := ResolveApprovalIntakeID(tx, approval)
	if err != nil {
		return err
	}
	if boundIntakeID != "" && intakeID != boundIntakeID {
		return snapshotUnresolved("审批接案关联缺失或与持久证据不一致")
	}
	if intakeID == "" && approval.ActiveSubmissionSnapshotID == nil {
		if req != nil {
			return snapshotUnresolved("原审批缺少可核验的接案关联")
		}
		return nil
	}
	if tx == nil {
		return errors.New("提交事务未初始化")
	}
	if approval.ApplicantID != userID {
		return intakeCollaborationError("FORBIDDEN", "只有申请人可以提交原审批", 403)
	}
	if approval.CaseCreated || approval.CreatedCaseID != "" {
		return submissionStale("已成案申请不能重新提交")
	}
	var previous map[string]interface{}
	round := 1
	var items []models.SupplementRequest
	var responses []models.SupplementResponse
	fingerprint := ""
	if req != nil {
		// This also makes accidental repeated Freeze calls harmless after a success.
		replay, err := ValidateResubmissionReplayInTx(tx, approval, req)
		if err != nil {
			return err
		}
		if replay {
			return nil
		}
		previous, err = ResolveApprovalSubmission(tx, approval)
		if err != nil {
			return err
		}
		if submissionText(previous["snapshot_id"]) != strings.TrimSpace(req.ExpectedSubmissionSnapshotID) {
			return submissionStale("提交版本已变化，请重新加载")
		}
		intakeID = submissionText(submissionMap(previous["metadata"])["intake_id"])
		fingerprint, err = resubmissionFingerprint(req)
		if err != nil {
			return err
		}
		previousVersion, _ := strconv.Atoi(submissionText(previous["source_version"]))
		round = previousVersion + 1
	} else {
		if approval.ActiveSubmissionSnapshotID != nil {
			return submissionStale("该审批已冻结提交版本，请使用原审批重提入口")
		}
		// Initial submissions must not silently turn an old returned approval into round 1.
		if approval.Status != models.ApprovalStatusDraft && approval.Status != models.ApprovalStatusSubmitted {
			return submissionStale("当前状态不允许首次冻结")
		}
		var existing int64
		if err := tx.Model(&models.ApprovalSnapshot{}).Where("approval_request_id = ? AND snapshot_type = ?", approval.ID, models.IntakeSubmissionSnapshotType).Count(&existing).Error; err != nil {
			return err
		}
		if existing != 0 {
			return snapshotUnresolved("已有提交轮次但当前指针缺失")
		}
	}
	if intakeID == "" {
		return snapshotUnresolved("提交缺少接案关联")
	}
	intake, metadata, err := lockSubmissionIntake(tx, intakeID, userID)
	if err != nil {
		return err
	}
	if err := rejectExistingIntakeCase(tx, intakeID); err != nil {
		return err
	}
	if req != nil {
		if strings.TrimSpace(req.ExpectedIntakeRevision) != submissionRevision(intakeID, metadata) {
			return intakeCollaborationError("INTAKE_DRAFT_REVISION_CONFLICT", "草稿已更新，请刷新后重新提交", 409)
		}
		record, err := currentIntakeReturn(tx, approval, req.ReturnRecordID)
		if err != nil {
			return err
		}
		items, err = ReadSupplementRequests(record)
		if err != nil {
			return err
		}
		responses, err = normalizeSupplementResponses(items, req.Responses, true)
		if err != nil {
			return err
		}
	}
	if intake.Status != "conflict_ready" || intake.Title == "" || intake.CaseType == "" || intake.ClientID == "" || submissionText(metadata["lawyer_facts_version"]) == "" || submissionText(metadata["lawyer_facts_confirmed_by"]) != userID || submissionText(metadata["lawyer_facts_confirmed_at"]) == "" {
		return intakeCollaborationError("INTAKE_FACTS_STALE", "请确认当前接案事实并完成正式冲突检测", 409)
	}
	checkID := submissionText(metadata["conflict_check_id"])
	if checkID == "" {
		return intakeCollaborationError("CONFLICT_CHECK_LINK_REQUIRED", "接案缺少当前正式检测", 409)
	}
	var check models.ConflictCheckRecord
	if err := tx.Where("check_id = ?", checkID).First(&check).Error; err != nil {
		return err
	}
	if SubjectIntakeIDFromSearchParameters(check.SearchParameters) != intakeID || check.ClientID != intake.ClientID || !strings.EqualFold(check.CheckStatus, "COMPLETED") || check.CaseName != intake.Title || check.CaseType != intake.CaseType {
		return intakeCollaborationError("INTAKE_CONFLICT_LINK_MISMATCH", "当前检测与接案事实不一致", 409)
	}
	if err := EnsureConflictRecordMatchesCurrentIntakeFacts(tx, intakeID, checkID, check.CheckTime); err != nil {
		return err
	}
	lawyerID := submissionText(metadata["lawyer_id"])
	if lawyerID == "" {
		lawyerID = userID
	}
	if strconv.FormatUint(uint64(check.UserID), 10) != lawyerID {
		return intakeCollaborationError("SUBJECT_LAWYER_MISMATCH", "当前承办律师与正式检测不一致", 409)
	}
	var response models.ConflictCheckResponse
	raw, err := json.Marshal(check.CheckResult)
	if err != nil {
		return err
	}
	if json.Unmarshal(raw, &response) != nil || response.Decision == nil || !conflictRecordCoverageComplete(&check, &response) || response.CheckID != checkID {
		return intakeCollaborationError("CONFLICT_COVERAGE_INCOMPLETE", "正式检测结果或覆盖范围不完整", 409)
	}
	var reviews []models.ConflictReview
	if err := tx.Where("check_id = ?", checkID).Order("created_at DESC, id DESC").Limit(2).Find(&reviews).Error; err != nil {
		return err
	}
	if len(reviews) != 1 {
		return intakeCollaborationError("CONFLICT_REVIEW_REQUIRED", "需要唯一有效的独立人工复核结论", 409)
	}
	review := reviews[0]
	if review.Decision != "no_conflict" && review.Decision != "false_positive" {
		return intakeCollaborationError("CONFLICT_REVIEW_REQUIRED", "当前独立复核结论不允许提交接案审批", 409)
	}
	if review.EvidenceHash == "" || review.EvidenceHash != conflictEvidenceHash(&response) || review.CreatedAt.Before(check.CheckTime) || (review.NextReviewAt != nil && !review.NextReviewAt.After(time.Now())) {
		return intakeCollaborationError("CONFLICT_REVIEW_STALE", "独立复核证据已失效，请重新检测及复核", 409)
	}
	if err := ValidateConflictReviewer(tx.Statement.Context, tx, checkID, 0, review.ReviewerID, ""); err != nil {
		return err
	}
	assignment, err := GetActiveConflictReviewerAssignment(tx.Statement.Context, tx, checkID)
	if err != nil {
		return err
	}
	if assignment == nil || assignment.ReviewerID != review.ReviewerID || assignment.RevokedAt != nil || (assignment.CaseID != nil && *assignment.CaseID != subjectCaseIDFromConflictParameters(check.SearchParameters)) || (assignment.EffectiveFrom != nil && review.CreatedAt.Before(*assignment.EffectiveFrom)) || review.CreatedAt.Before(assignment.CreatedAt) {
		return intakeCollaborationError("CONFLICT_REVIEW_ASSIGNMENT_STALE", "独立复核结论与当前有效指派不一致", 409)
	}
	var client map[string]interface{}
	if err := tx.Table("clients").Where("id = ? AND deleted_at IS NULL", intake.ClientID).Take(&client).Error; err != nil {
		return err
	}
	clientDigest := submissionText(client["identity_number_digest"])
	if clientDigest == "" {
		clientDigest = submissionText(client["id_card_digest"])
	}
	reviewedClient := clientPartyFromRecord(&check)
	if reviewedClient == nil || clientDigest == "" || submissionIdentityDigest(reviewedClient.Identifiers) != clientDigest || submissionText(client["name"]) != check.ClientName {
		return intakeCollaborationError("INTAKE_CLIENT_IDENTITY_STALE", "客户主档案与正式检测的身份不一致，请重新检测", 409)
	}
	parties, err := buildReviewedParties(&check)
	if err != nil {
		return err
	}
	if len(parties) < 2 {
		return intakeCollaborationError("INTAKE_PARTIES_REQUIRED", "正式检测缺少完整客户及当事人", 409)
	}
	var partyRows []map[string]interface{}
	if err := tx.Table("case_intake_parties").Where("intake_id = ?", intakeID).Order("id ASC").Find(&partyRows).Error; err != nil {
		return err
	}
	if err := validateSubmissionPartyRows(partyRows, &check); err != nil {
		return err
	}
	if err := validateSupplementTargets(items, partyRows); err != nil {
		return err
	}
	var materials []map[string]interface{}
	if err := tx.Table("case_materials").Where("intake_id = ?", intakeID).Order("id ASC").Find(&materials).Error; err != nil {
		return err
	}
	for _, material := range materials {
		if raw, ok := material["metadata"]; ok {
			material["metadata"] = submissionMap(raw)
		}
	}
	// All matter facts originate in the locked intake and immutable check. Never
	// copy arbitrary old approval metadata, even if it looks like case config.
	config := map[string]interface{}{"title": intake.Title, "case_type": intake.CaseType, "client_id": intake.ClientID, "lawyer_id": lawyerID, "description": intake.Description, "priority": intake.Priority}
	for _, key := range []string{"business_area", "sub_area", "dispute_amount", "source_channel", "source_contact", "investment_agreement_date", "dispute_date", "breach_date", "proposed_filing_date", "jurisdiction", "billing_method", "fee_base", "contingency_rate", "minimum_fee", "case_tags"} {
		if value, ok := metadata[key]; ok {
			config[key] = value
		}
	}
	frozen := map[string]interface{}{"intake_id": intakeID, "intake_code": intake.IntakeCode, "client": map[string]interface{}{"id": intake.ClientID, "name": check.ClientName}, "case_creation_config": config, "parties": parties, "intake_parties": partyRows, "materials": materials, "conflict_check_id": checkID, "conflict_result": check.CheckResult, "conflict_record": map[string]interface{}{"check_id": checkID, "check_time": check.CheckTime, "risk_level": check.RiskLevel, "client_id": check.ClientID, "user_id": check.UserID, "search_parameters": check.SearchParameters}, "lawyer_facts_version": metadata["lawyer_facts_version"], "lawyer_facts_confirmed_at": metadata["lawyer_facts_confirmed_at"], "lawyer_facts_confirmed_by": metadata["lawyer_facts_confirmed_by"], "integration_type": "both"}
	now := time.Now().UTC()
	id := uuid.NewString()
	snapshot := map[string]interface{}{"schema_version": 1, "snapshot_id": id, "snapshot_at": now.Format(time.RFC3339Nano), "snapshot_type": models.IntakeSubmissionSnapshotType, "source_version": round, "approval_request_id": approval.ID, "intake_id": intakeID, "intake_revision": submissionRevision(intakeID, metadata), "lawyer_facts_version": metadata["lawyer_facts_version"], "conflict_check_id": checkID, "conflict_review": map[string]interface{}{"id": review.ID, "check_id": review.CheckID, "decision": review.Decision, "reviewer_id": review.ReviewerID, "evidence_hash": review.EvidenceHash, "created_at": review.CreatedAt, "next_review_at": review.NextReviewAt}, "conflict_reviewer_assignment": map[string]interface{}{"id": assignment.ID, "check_id": assignment.CheckID, "reviewer_id": assignment.ReviewerID, "case_id": assignment.CaseID, "recusal_declared": assignment.RecusalDeclared, "created_at": assignment.CreatedAt, "effective_from": assignment.EffectiveFrom, "effective_to": assignment.EffectiveTo}, "metadata": frozen, "case_creation_config": config, "submitted_by": userID, "applicant": map[string]interface{}{"id": userID, "name": approval.ApplicantName}, "supplement_requests": items, "supplement_responses": responses}
	if req != nil {
		snapshot["previous_snapshot_id"] = previous["snapshot_id"]
		snapshot["return_record_id"] = req.ReturnRecordID
		snapshot["revision_note"] = strings.TrimSpace(req.RevisionNote)
		snapshot["resubmission_fingerprint"] = fingerprint
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	row := models.ApprovalSnapshot{ID: id, ApprovalRequestID: approval.ID, SnapshotType: models.IntakeSubmissionSnapshotType, SnapshotData: string(encoded), SourceVersion: round, CreatedAt: now}
	if err := tx.Create(&row).Error; err != nil {
		return err
	}
	approval.ActiveSubmissionSnapshotID = &id
	projectSubmissionConflict(approval, submissionMap(snapshot))
	return nil
}

func validateSupplementTargets(items []models.SupplementRequest, rows []map[string]interface{}) error {
	ids := map[string]bool{}
	for _, row := range rows {
		metadata := submissionMap(row["metadata"])
		for _, key := range []string{"localId", "local_id"} {
			if id := submissionText(metadata[key]); id != "" {
				ids[id] = true
			}
		}
	}
	for _, item := range items {
		if item.PartyLocalID != "" && !ids[item.PartyLocalID] {
			return supplementInvalid("补充项对应的当事人已不存在，请审批人重新确认补充范围")
		}
	}
	return nil
}

// Check the live identity/role set against the recorded materialization source.
// Ciphertext is never copied into an approval snapshot.
func validateSubmissionPartyRows(rows []map[string]interface{}, check *models.ConflictCheckRecord) error {
	checked := casePartiesFromSearchParameters(check.SearchParameters)
	live := map[string]int{}
	frozen := map[string]int{}
	for _, party := range checked {
		digest := submissionIdentityDigest(party.Identifiers)
		if digest == "" {
			return intakeCollaborationError("INTAKE_PARTY_IDENTITY_REQUIRED", "正式检测缺少受保护的当事人身份", 409)
		}
		role := strings.ToUpper(strings.TrimSpace(party.Role))
		if repositories.NormalizeConflictPartyRole(role) == "" {
			return intakeCollaborationError("INTAKE_PARTY_ROLE_INVALID", "当事人角色无法用于正式成案", 409)
		}
		frozen[party.Name+"\x00"+role+"\x00"+digest]++
	}
	for _, row := range rows {
		delete(row, "identity_number_ciphertext")
		metadata := submissionMap(row["metadata"])
		row["metadata"] = metadata
		role := strings.ToUpper(submissionText(row["party_role"]))
		if role == "CLIENT" {
			continue
		}
		if role == "RELATED_PARTY" && submissionText(metadata["party_type"]) != "" {
			role = strings.ToUpper(submissionText(metadata["party_type"]))
		}
		digest := submissionText(row["identity_number_digest"])
		if digest == "" {
			return intakeCollaborationError("INTAKE_PARTY_IDENTITY_REQUIRED", "当事人身份尚未核验", 409)
		}
		if repositories.NormalizeConflictPartyRole(role) == "" {
			return intakeCollaborationError("INTAKE_PARTY_ROLE_INVALID", "当事人角色无法用于正式成案", 409)
		}
		live[submissionText(row["entity_name"])+"\x00"+role+"\x00"+digest]++
	}
	if len(live) == 0 || !reflect.DeepEqual(live, frozen) {
		return intakeCollaborationError("INTAKE_FACTS_STALE", "接案当事人身份或角色与正式检测不一致，请重新检测", 409)
	}
	return nil
}

func submissionIdentityDigest(identifiers map[string]string) string {
	for _, key := range []string{"id_card", "social_credit_code", "unified_social_credit_code", "passport", "business_license", "organization_code"} {
		if value := identifiers[key]; strings.HasPrefix(value, "hmac-sha256:") {
			return strings.TrimPrefix(value, "hmac-sha256:")
		}
	}
	return ""
}
