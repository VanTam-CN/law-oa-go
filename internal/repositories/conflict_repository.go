package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

// BasicConflictRepository 基础冲突检测数据仓库接口
type BasicConflictRepository interface {
	// 保存冲突检测记录
	SaveCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error
	// 获取单条冲突检测记录
	GetCheckRecord(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error)
	// 更新冲突检测记录
	UpdateCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error
	// 获取冲突检测历史
	GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error)
	// 获取冲突案例
	GetConflictCases(ctx context.Context, params *ConflictSearchParams) ([]*models.ConflictCase, error)
	// 获取潜在冲突案例（从主案件表）
	GetPotentialConflicts(ctx context.Context, clientID string, lawyerID uint, otherParties []string, since time.Time) ([]*models.ConflictCase, error)
	// 获取客户关系
	GetClientRelations(ctx context.Context, clientID string) ([]*models.ClientRelation, error)
	// 保存冲突案例
	SaveConflictCases(ctx context.Context, cases []*models.ConflictCase) error
	// 获取冲突规则
	GetConflictRules(ctx context.Context, activeOnly bool) ([]*models.ConflictRule, error)
	// 保存冲突规则
	SaveConflictRule(ctx context.Context, rule *models.ConflictRule) error
	// 更新冲突规则
	UpdateConflictRule(ctx context.Context, rule *models.ConflictRule) error
	// 获取MCP标准
	GetMCPStandards(ctx context.Context, activeOnly bool) (*models.MCPStandards, error)
	// 保存MCP标准
	SaveMCPStandards(ctx context.Context, standards *models.MCPStandards) error
	// 获取统计信息
	GetConflictStats(ctx context.Context, clientID string) (*ConflictStats, error)
}

// ConflictSubjectAssociation is the server-side link between the submitted
// case context and the immutable conflict-check record. The association is
// only accepted with an explicit case ID; title-only matching is deliberately
// unsupported.
type ConflictSubjectAssociation struct {
	CheckID           string
	SubjectCaseID     string
	SubjectCaseNumber string
	IntakeID          string
	ClientID          string
	CoverageStatus    string
	CheckedAt         time.Time
}

// ConflictSubjectLinker is optional on BasicConflictRepository so existing
// repository test doubles remain source-compatible. Production repositories
// implement it to persist the case/intake association transactionally.
type ConflictSubjectLinker interface {
	LinkConflictCheckToCase(ctx context.Context, association ConflictSubjectAssociation) error
}

// ConflictP0EvidenceWriter is optional so existing test doubles and legacy
// repositories remain source-compatible. The production repository implements
// it and writes the normalized P0 subject/evidence tables after each check.
type ConflictP0EvidenceWriter interface {
	SaveConflictP0Evidence(ctx context.Context, checkID string, subjects []models.ConflictNormalizedSubject, response *models.ConflictCheckResponse) error
}

// ConflictP0SubjectIndexHit is a match returned by the firm-wide immutable
// subject index. It contains only the source snapshot needed to construct a
// redacted conflict result; the API layer still applies ethical-wall policy.
type ConflictP0SubjectIndexHit struct {
	Version        models.ConflictSubjectVersion
	Identifiers    []models.ConflictSubjectIdentifier
	MatchType      string
	RuleCode       string
	RequestedParty string
	MatchedName    string
	MatchSource    string
	RelationType   string
}

// ConflictP0SubjectIndexer is optional for legacy non-production test doubles.
// A production repository implements both synchronization and read access.
type ConflictP0SubjectIndexer interface {
	SyncConflictSubjectIndex(ctx context.Context) error
	SearchConflictSubjectIndex(ctx context.Context, subjects []models.ConflictNormalizedSubject, clientID string) ([]ConflictP0SubjectIndexHit, error)
}

// ConflictP0SubjectIndexReconciler is the explicit operator path for a
// historical index build. Per-check lazy sync is not sufficient evidence for
// production coverage, so readiness requires a completed reconciliation run.
type ConflictP0SubjectIndexReconciler interface {
	ReconcileConflictSubjectIndex(ctx context.Context, actorID uint, evidenceReference string, apply bool) ([]models.ConflictIndexBuildRun, error)
}

// ConflictSearchParams 冲突案例搜索参数
type ConflictSearchParams struct {
	ClientID  string    `json:"clientId"`
	CaseType  string    `json:"caseType"`
	RiskLevel string    `json:"riskLevel"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	Page      int       `json:"page"`
	PageSize  int       `json:"pageSize"`
}

// ConflictStats 冲突检测统计
type ConflictStats struct {
	TotalChecks     int64     `json:"totalChecks"`
	ConflictChecks  int64     `json:"conflictChecks"`
	HighRiskChecks  int64     `json:"highRiskChecks"`
	AverageDuration float64   `json:"averageDuration"`
	LastCheckTime   time.Time `json:"lastCheckTime"`
}

// conflictRepository 冲突检测数据仓库实现
type conflictRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewConflictRepository 创建新的冲突检测数据仓库
func NewConflictRepository(db *gorm.DB, redis *redis.Client) BasicConflictRepository {
	return &conflictRepository{
		db:    db,
		redis: redis,
	}
}

// LinkConflictCheckToCase persists the explicit subject-case association and
// the intake status in one transaction. A client mismatch, stale case number,
// or missing case is a hard error because silently attaching a result to the
// wrong matter would be worse than leaving it unlinked.
func (r *conflictRepository) LinkConflictCheckToCase(ctx context.Context, association ConflictSubjectAssociation) error {
	association.CheckID = strings.TrimSpace(association.CheckID)
	association.SubjectCaseID = strings.TrimSpace(association.SubjectCaseID)
	association.SubjectCaseNumber = strings.TrimSpace(association.SubjectCaseNumber)
	association.IntakeID = strings.TrimSpace(association.IntakeID)
	association.ClientID = strings.TrimSpace(association.ClientID)
	if association.CheckID == "" {
		return errors.New("冲突检测记录不能为空")
	}
	if association.SubjectCaseID == "" && association.IntakeID == "" {
		return nil
	}
	if association.ClientID == "" {
		return errors.New("案件关联失败：客户绑定不能为空")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record struct {
			ClientID         string
			SearchParameters string
		}
		if err := tx.Table("conflict_check_records").Select("client_id, search_parameters").Where("check_id = ?", association.CheckID).Take(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("案件关联失败：冲突检测记录不存在")
			}
			return fmt.Errorf("读取冲突检测记录失败: %w", err)
		}
		if strings.TrimSpace(record.ClientID) == "" || record.ClientID != association.ClientID {
			return fmt.Errorf("案件关联失败：检测记录客户与关联客户不一致")
		}
		var context struct {
			ClientID           string `json:"clientId"`
			ClientIDSnake      string `json:"client_id"`
			SubjectCaseID      string `json:"subjectCaseId"`
			SubjectCaseIDSnake string `json:"subject_case_id"`
			IntakeID           string `json:"intakeId"`
			IntakeIDSnake      string `json:"intake_id"`
		}
		if raw := strings.TrimSpace(record.SearchParameters); raw != "" {
			if err := json.Unmarshal([]byte(raw), &context); err != nil {
				return fmt.Errorf("案件关联失败：检测记录上下文无效")
			}
		}
		if recordedClientID := firstAssociationValue(context.ClientID, context.ClientIDSnake); recordedClientID != "" && recordedClientID != association.ClientID {
			return fmt.Errorf("案件关联失败：检测记录上下文客户不一致")
		}
		if association.SubjectCaseID != "" {
			recordedCaseID := firstAssociationValue(context.SubjectCaseID, context.SubjectCaseIDSnake)
			if recordedCaseID == "" || recordedCaseID != association.SubjectCaseID {
				return fmt.Errorf("案件关联失败：案件 ID 与检测记录上下文不一致")
			}
		}
		if association.IntakeID != "" {
			recordedIntakeID := firstAssociationValue(context.IntakeID, context.IntakeIDSnake)
			if recordedIntakeID == "" || recordedIntakeID != association.IntakeID {
				return fmt.Errorf("接案关联失败：接案 ID 与检测记录上下文不一致")
			}
		}

		if association.SubjectCaseID != "" {
			caseID, err := strconv.ParseUint(association.SubjectCaseID, 10, 32)
			if err != nil || caseID == 0 {
				return fmt.Errorf("案件关联失败：案件 ID 无效")
			}

			var target struct {
				ID              uint
				CaseNumber      string
				ClientID        uint
				ConflictCheckID string
			}
			if err := tx.Table("cases").Select("id, case_number, client_id, conflict_check_id").Where("id = ? AND deleted_at IS NULL", caseID).Take(&target).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("案件关联失败：案件不存在或已归档")
				}
				return fmt.Errorf("读取案件关联失败: %w", err)
			}
			if association.SubjectCaseNumber != "" && association.SubjectCaseNumber != target.CaseNumber {
				return fmt.Errorf("案件关联失败：案件编号与案件 ID 不一致")
			}
			if association.ClientID != "" {
				expectedClientID, err := strconv.ParseUint(association.ClientID, 10, 32)
				if err != nil || expectedClientID == 0 || uint(expectedClientID) != target.ClientID {
					return fmt.Errorf("案件关联失败：案件客户与检测客户不一致")
				}
			}
			if existingCheckID := strings.TrimSpace(target.ConflictCheckID); existingCheckID != "" && existingCheckID != association.CheckID {
				return fmt.Errorf("案件关联失败：案件已绑定其他冲突检测记录")
			}

			updates := map[string]interface{}{"conflict_check_id": association.CheckID}
			if coverageStatus := strings.TrimSpace(association.CoverageStatus); coverageStatus != "" {
				updates["conflict_coverage_status"] = coverageStatus
			}
			if err := tx.Table("cases").Where("id = ?", target.ID).Updates(updates).Error; err != nil {
				return fmt.Errorf("保存案件冲突检测关联失败: %w", err)
			}
		}

		if association.IntakeID != "" {
			var intake map[string]interface{}
			if err := tx.Table("case_intakes").Where("id = ?", association.IntakeID).Take(&intake).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("接案关联失败：接案记录不存在")
				}
				return fmt.Errorf("读取接案关联失败: %w", err)
			}
			intakeClientID := strings.TrimSpace(fmt.Sprint(intake["client_id"]))
			if intakeClientID == "" || intakeClientID != association.ClientID {
				return fmt.Errorf("接案关联失败：接案客户与检测客户不一致")
			}
			metadata := decodeJSONMap(intake["metadata"])
			if existingCheckID := firstAssociationValue(associationMapValue(metadata, "conflict_check_id"), associationMapValue(metadata, "conflictCheckId")); existingCheckID != "" && existingCheckID != association.CheckID {
				return fmt.Errorf("接案关联失败：接案记录已绑定其他冲突检测记录")
			}
			metadata["conflict_check_id"] = association.CheckID
			if coverageStatus := strings.TrimSpace(association.CoverageStatus); coverageStatus != "" {
				metadata["conflict_coverage_status"] = coverageStatus
			}
			if !association.CheckedAt.IsZero() {
				metadata["conflict_checked_at"] = association.CheckedAt
			}
			if association.SubjectCaseID != "" {
				metadata["subject_case_id"] = association.SubjectCaseID
			}
			if association.SubjectCaseNumber != "" {
				metadata["subject_case_number"] = association.SubjectCaseNumber
			}
			if err := tx.Table("case_intakes").Where("id = ?", association.IntakeID).Updates(map[string]interface{}{
				"status":     "conflict_ready",
				"metadata":   encodeJSONMap(metadata),
				"updated_at": time.Now(),
			}).Error; err != nil {
				return fmt.Errorf("保存接案冲突检测关联失败: %w", err)
			}
		}
		return nil
	})
}

func firstAssociationValue(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" && value != "0" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func associationMapValue(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(values[key]))
}

// LinkConflictCheckToCase is used by handlers that execute the synchronous
// intake endpoint rather than the asynchronous task service.
func LinkConflictCheckToCase(ctx context.Context, db *gorm.DB, association ConflictSubjectAssociation) error {
	if db == nil {
		return errors.New("冲突检测案件关联服务未初始化")
	}
	return (&conflictRepository{db: db}).LinkConflictCheckToCase(ctx, association)
}

func decodeJSONMap(value interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	if values, ok := value.(map[string]interface{}); ok {
		return values
	}
	var raw []byte
	switch typed := value.(type) {
	case []byte:
		raw = typed
	case string:
		raw = []byte(typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return map[string]interface{}{}
		}
		raw = encoded
	}
	var values map[string]interface{}
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return map[string]interface{}{}
	}
	return values
}

func encodeJSONMap(values map[string]interface{}) string {
	encoded, err := json.Marshal(values)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

// SaveCheckRecord 保存冲突检测记录
func (r *conflictRepository) SaveCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error {
	if record == nil || strings.TrimSpace(record.CheckID) == "" {
		return errors.New("冲突检测记录缺少不可变检测编号")
	}
	record.UpdatedAt = time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = record.UpdatedAt
	}

	// A check ID identifies one evidence snapshot. Updating an existing row on
	// retry could silently replace the subjects, search scope, or result that
	// a reviewer is expected to sign off. State transitions use the explicit
	// workflow update path instead.
	err := r.db.WithContext(ctx).Create(record).Error
	if err != nil {
		return fmt.Errorf("保存冲突检测记录失败: %w", err)
	}

	return nil
}

// SaveConflictP0Evidence persists the immutable subject snapshot used by the
// check and its normalized evidence rows. The write is idempotent by stable
// content IDs, so a request retry cannot duplicate compliance evidence.
func (r *conflictRepository) SaveConflictP0Evidence(ctx context.Context, checkID string, subjects []models.ConflictNormalizedSubject, response *models.ConflictCheckResponse) error {
	checkID = strings.TrimSpace(checkID)
	if checkID == "" || response == nil {
		return errors.New("P0 冲突证据缺少检测编号或结果")
	}
	if !r.db.Migrator().HasTable((&models.ConflictSubjectVersion{}).TableName()) ||
		!r.db.Migrator().HasTable((&models.ConflictSubjectIdentifier{}).TableName()) ||
		!r.db.Migrator().HasTable((&models.ConflictMatchEvidenceV2{}).TableName()) {
		return errors.New("P0 冲突证据表未部署")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		subjectIDs := make(map[string]string, len(subjects))
		for _, subject := range subjects {
			if strings.TrimSpace(subject.NormalizedName) == "" {
				continue
			}
			subjectKey := fmt.Sprintf("%s:%s:%s", checkID, subject.Role, subject.NormalizedName)
			versionID := stableConflictP0ID("SV", subjectKey)
			snapshot := map[string]interface{}{
				"originalName":   subject.OriginalName,
				"normalizedName": subject.NormalizedName,
				"role":           subject.Role,
				"entityType":     subject.EntityType,
				"aliases":        subject.Aliases,
			}
			snapshotJSON, err := json.Marshal(snapshot)
			if err != nil {
				return fmt.Errorf("序列化主体版本快照失败: %w", err)
			}
			aliasJSON, err := json.Marshal(subject.Aliases)
			if err != nil {
				return fmt.Errorf("序列化主体别名快照失败: %w", err)
			}
			version := &models.ConflictSubjectVersion{
				ID: versionID, SubjectKey: subjectKey, SourceType: "CONFLICT_CHECK_SUBJECT", SourceID: checkID,
				SubjectRole: subject.Role, SubjectType: subject.EntityType, OriginalName: subject.OriginalName,
				NormalizedName: subject.NormalizedName, AliasSnapshot: string(aliasJSON), SourceVersion: checkID,
				VersionNumber: 1, Verification: "SUBMITTED", Snapshot: string(snapshotJSON), CreatedAt: time.Now(),
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(version).Error; err != nil {
				return fmt.Errorf("保存主体版本快照失败: %w", err)
			}
			subjectIDs[strings.ToLower(subject.OriginalName)] = versionID
			subjectIDs[strings.ToLower(subject.NormalizedName)] = versionID
			for _, alias := range subject.Aliases {
				subjectIDs[strings.ToLower(strings.TrimSpace(alias))] = versionID
			}

			for identifierType, identifierValue := range subject.Identifiers {
				identifierValue = strings.TrimSpace(identifierValue)
				if identifierValue == "" {
					continue
				}
				ciphertext, digest, err := security.ProtectIdentityNumber(identifierValue)
				if err != nil {
					return fmt.Errorf("保护主体身份标识失败: %w", err)
				}
				identifierID := stableConflictP0ID("SI", versionID+":"+strings.ToLower(strings.TrimSpace(identifierType))+":"+digest)
				identifier := &models.ConflictSubjectIdentifier{
					ID: identifierID, SubjectVersionID: versionID, IdentifierType: strings.ToUpper(strings.TrimSpace(identifierType)),
					Digest: digest, Ciphertext: ciphertext, MaskedValue: security.MaskIdentityNumber(identifierValue),
					Verification: "SUBMITTED", SourceReference: "CONFLICT_CHECK:" + checkID, CreatedAt: time.Now(),
				}
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(identifier).Error; err != nil {
					return fmt.Errorf("保存主体身份索引失败: %w", err)
				}
			}
		}

		for _, conflict := range response.ConflictCases {
			if conflict == nil {
				continue
			}
			for _, evidence := range conflict.Evidence {
				evidenceJSON, err := json.Marshal(evidence)
				if err != nil {
					return fmt.Errorf("序列化冲突证据失败: %w", err)
				}
				hash := sha256.Sum256(evidenceJSON)
				evidenceID := stableConflictP0ID("ME", checkID+":"+evidence.EvidenceID+":"+evidence.SourceCaseID)
				subjectVersionID := subjectIDs[strings.ToLower(strings.TrimSpace(evidence.RequestedParty))]
				row := &models.ConflictMatchEvidenceV2{
					ID: evidenceID, CheckID: checkID, SubjectVersionID: subjectVersionID,
					MatchType: evidence.MatchType, SourceType: evidence.SourceType, SourceObjectID: evidence.SourceCaseID,
					Restricted: evidence.Restricted, EvidenceSnapshot: string(evidenceJSON),
					EvidenceHash: fmt.Sprintf("%x", hash[:]), CreatedAt: time.Now(),
				}
				if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error; err != nil {
					return fmt.Errorf("保存冲突命中证据失败: %w", err)
				}
			}
		}
		return nil
	})
}

func stableConflictP0ID(prefix, value string) string {
	sum := sha256.Sum256([]byte(value))
	return prefix + "_" + fmt.Sprintf("%x", sum[:])
}

// GetCheckRecord 获取单条冲突检测记录
func (r *conflictRepository) GetCheckRecord(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error) {
	var record models.ConflictCheckRecord
	err := r.db.WithContext(ctx).Where("check_id = ?", checkID).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("获取冲突检测记录失败: %w", err)
	}
	return &record, nil
}

// UpdateCheckRecord 更新冲突检测记录
func (r *conflictRepository) UpdateCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("更新冲突检测记录失败: %w", err)
	}
	return nil
}

// GetCheckHistory 获取冲突检测历史
func (r *conflictRepository) GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error) {
	var records []*models.ConflictCheckRecord

	query := r.db.WithContext(ctx).Where("client_id = ?", clientID).
		Order("check_time DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("获取冲突检测历史失败: %w", err)
	}

	return records, nil
}

// GetConflictCases 获取冲突案例
func (r *conflictRepository) GetConflictCases(ctx context.Context, params *ConflictSearchParams) ([]*models.ConflictCase, error) {
	var cases []*models.ConflictCase

	query := r.db.WithContext(ctx).Model(&models.ConflictCase{})

	if params.ClientID != "" {
		query = query.Where("client_id = ?", params.ClientID)
	}
	if params.CaseType != "" {
		query = query.Where("case_type = ?", params.CaseType)
	}
	if params.RiskLevel != "" {
		query = query.Where("risk_level = ?", params.RiskLevel)
	}
	if !params.StartDate.IsZero() {
		query = query.Where("created_at >= ?", params.StartDate)
	}
	if !params.EndDate.IsZero() {
		query = query.Where("created_at <= ?", params.EndDate)
	}

	// 分页
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	// 按时间倒序
	query = query.Order("created_at DESC")

	if err := query.Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("获取冲突案例失败: %w", err)
	}

	return cases, nil
}

// GetPotentialConflicts 获取潜在冲突案例（从主案件表）
func (r *conflictRepository) GetPotentialConflicts(ctx context.Context, clientID string, lawyerID uint, otherParties []string, since time.Time) ([]*models.ConflictCase, error) {
	var conflictCases []*models.ConflictCase
	seenCaseIDs := make(map[string]struct{})

	log.Printf("🔍 查询潜在冲突: clientID=%s, lawyerID=%d", clientID, lawyerID)

	// 🔧 修复：需要将字符串clientID转换为uint，同时保持原逻辑
	var clientIDUint uint
	if _, err := fmt.Sscanf(clientID, "%d", &clientIDUint); err != nil {
		// Returning an empty result here would turn malformed subject input into
		// a false clean check. The caller must fail closed instead.
		return nil, fmt.Errorf("客户ID格式错误: %s", clientID)
	}

	sinceFilter := ""
	// Conflict review is a firm-level duty. The lawyer ID remains part of the
	// request contract for compatibility and is used by the caller for
	// ownership/audit, but it must not narrow the historical archive search.
	args := []interface{}{clientIDUint}
	if !since.IsZero() {
		sinceFilter = "AND c.created_at >= ?"
		args = append(args, since)
	}

	// 查询全所历史案件，不能按当前发起律师缩小冲突检索范围。
	query := fmt.Sprintf(`
		SELECT
			c.id as case_id,
			c.case_number,
			c.title as case_name,
			c.case_type,
			c.description,
			c.client_id,
			cl.name as client_name,
			cl.type as client_type,
			u.name as lawyer_name,
			c.created_at,
			c.lawyer_id
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.client_id != ?
		AND c.deleted_at IS NULL
		%s
		ORDER BY c.created_at DESC
	`, sinceFilter)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询潜在冲突案例失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var caseModel models.Case
		var clientName, clientType, lawyerName string
		var foundLawyerID uint

		err := rows.Scan(
			&caseModel.ID,
			&caseModel.CaseNumber,
			&caseModel.Title,
			&caseModel.CaseType,
			&caseModel.Description,
			&caseModel.ClientID,
			&clientName,
			&clientType,
			&lawyerName,
			&caseModel.CreatedAt,
			&foundLawyerID,
		)
		if err != nil {
			log.Printf("⚠️ 扫描案件数据失败: %v", err)
			continue
		}

		matchKind, matchedParty := strongestOpposingPartyMatch(clientName, otherParties)
		if matchKind == PartyNoMatch {
			continue
		}

		conflictType := "名称相似待核实"
		riskLevel := "MEDIUM"
		description := fmt.Sprintf("当前对方当事人 %q 与律师 %s 的历史客户 %q 名称相似，需核实是否为同一主体", matchedParty, lawyerName, clientName)
		conflictDetails := "名称候选命中，不足以单独认定利益冲突"
		if matchKind == PartyExactNormalizedMatch {
			conflictType = "对方当事人直接冲突"
			riskLevel = "CRITICAL"
			description = fmt.Sprintf("当前对方当事人 %q 与律师 %s 的历史客户 %q 规范化名称完全一致", matchedParty, lawyerName, clientName)
			conflictDetails = "对方当事人与承办律师历史客户规范化名称完全命中"
		}

		// 创建冲突案例对象
		conflictCase := &models.ConflictCase{
			ID:              fmt.Sprintf("case_%d", caseModel.ID),
			CaseID:          fmt.Sprintf("%d", caseModel.ID),
			CaseName:        caseModel.Title,
			CaseNo:          caseModel.CaseNumber,
			CaseType:        caseModel.CaseType,
			Description:     description,
			ClientID:        fmt.Sprintf("%d", caseModel.ClientID),
			RiskLevel:       riskLevel,
			ConflictType:    conflictType,
			ConflictDetails: conflictDetails,
			CaseStatus:      "active",
			CreatedAt:       caseModel.CreatedAt,
		}
		conflictCase.OpposingParties = otherParties

		conflictCases = append(conflictCases, conflictCase)
		seenCaseIDs[conflictCase.CaseID] = struct{}{}
		log.Printf("✅ 创建冲突案例: %s", conflictCase.ID)
	}

	// 如果提供了其他当事人信息，也查询全所相关案件。
	if len(otherParties) > 0 {
		for _, party := range otherParties {
			// 查询包含对方当事人名称的案件描述
			partySinceFilter := ""
			partyArgs := []interface{}{"%" + party + "%", "%" + party + "%"}
			if !since.IsZero() {
				partySinceFilter = "AND c.created_at >= ?"
				partyArgs = append([]interface{}{since}, partyArgs...)
			}

			partyQuery := fmt.Sprintf(`
				SELECT
					c.id as case_id,
					c.case_number,
					c.title as case_name,
					c.case_type,
					c.description,
					c.client_id,
					cl.name as client_name,
					cl.type as client_type,
					u.name as lawyer_name,
					c.created_at,
					c.lawyer_id
				FROM cases c
				JOIN clients cl ON c.client_id = cl.id
				JOIN users u ON c.lawyer_id = u.id
				WHERE c.deleted_at IS NULL
				%s
				AND (LOWER(c.title) LIKE LOWER(?) OR LOWER(c.description) LIKE LOWER(?))
				ORDER BY c.created_at DESC
			`, partySinceFilter)

			partyRows, err := r.db.WithContext(ctx).Raw(partyQuery, partyArgs...).Rows()
			if err != nil {
				return nil, fmt.Errorf("查询对方当事人冲突失败 (party=%q): %w", party, err)
			}

			for partyRows.Next() {
				var caseModel models.Case
				var clientName, clientType, lawyerName string
				var foundLawyerID uint

				err := partyRows.Scan(
					&caseModel.ID,
					&caseModel.CaseNumber,
					&caseModel.Title,
					&caseModel.CaseType,
					&caseModel.Description,
					&caseModel.ClientID,
					&clientName,
					&clientType,
					&lawyerName,
					&caseModel.CreatedAt,
					&foundLawyerID,
				)
				if err != nil {
					partyRows.Close()
					return nil, fmt.Errorf("扫描对方当事人冲突行失败 (party=%q): %w", party, err)
				}

				// 如果是同一个律师的案件，跳过（已经在上面查过了）
				if foundLawyerID == lawyerID {
					continue
				}
				caseID := fmt.Sprintf("%d", caseModel.ID)
				if _, exists := seenCaseIDs[caseID]; exists {
					continue
				}

				// 创建冲突案例对象
				conflictCase := &models.ConflictCase{
					ID:              fmt.Sprintf("case_%d", caseModel.ID),
					CaseID:          fmt.Sprintf("%d", caseModel.ID),
					CaseName:        caseModel.Title,
					CaseNo:          caseModel.CaseNumber,
					CaseType:        caseModel.CaseType,
					Description:     fmt.Sprintf("对方名称 %q 出现在历史案件标题或摘要中，需核实其当事人角色", party),
					ClientID:        fmt.Sprintf("%d", caseModel.ClientID),
					RiskLevel:       "MEDIUM",
					ConflictType:    "文本提及待核实",
					ConflictDetails: "非结构化文本命中，不能单独认定对方当事人冲突",
					CaseStatus:      "active",
					CreatedAt:       caseModel.CreatedAt,
				}

				conflictCases = append(conflictCases, conflictCase)
				seenCaseIDs[caseID] = struct{}{}
			}
			partyRows.Close()
		}
	}

	log.Printf("🎯 冲突检测完成: 找到 %d 个潜在冲突案例", len(conflictCases))
	return conflictCases, nil
}

func strongestOpposingPartyMatch(clientName string, otherParties []string) (PartyMatchKind, string) {
	strongest := PartyNoMatch
	matchedParty := ""
	for _, party := range otherParties {
		kind := classifyDirectOpposingPartyMatch(clientName, party)
		if kind > strongest {
			strongest = kind
			matchedParty = strings.TrimSpace(party)
		}
	}
	return strongest, matchedParty
}

// PartyMatchKind 当事人名称匹配强度三态分类。
//
// 设计底层逻辑：一个 bool 承载不了"完全相等 vs 包含/简称"两种语义，
// 后者会误把短子串升级为 CRITICAL。分类后由 caller 按风险等级分流：
//   - PartyExactNormalizedMatch：可直接判 CRITICAL
//   - PartyCandidateMatch：仅作为 MEDIUM 候选，必须人工复核
//   - PartyNoMatch：不构成命中
type PartyMatchKind int

const (
	PartyNoMatch PartyMatchKind = iota
	PartyCandidateMatch
	PartyExactNormalizedMatch
)

// isDirectOpposingPartyClient 仅在存在 Exact 规范化相等时返回 true。
// 调用方需要区分 Candidate 时，请直接使用 classifyDirectOpposingPartyMatch。
//
// 旧实现用 strings.Contains 就返回 true → "华" 会命中 "华为技术有限公司"
// → 错误升级为 CRITICAL → 拒绝接案。新行为：仅 Exact 直接命中，候选返回 false。
func isDirectOpposingPartyClient(clientName string, otherParties []string) bool {
	for _, party := range otherParties {
		if classifyDirectOpposingPartyMatch(clientName, party) == PartyExactNormalizedMatch {
			return true
		}
	}
	return false
}

// classifyDirectOpposingPartyMatch 把单对当事人名称分成三态：
//   - Exact：去除大小写、空白、公司后缀后完全相等
//   - Candidate：单向/双向包含——只能作为候选
//   - NoMatch：完全无关，或任一输入过短/为空
//
// 短子串防护：任一侧规范化后长度 < 2 直接判 NoMatch。
func classifyDirectOpposingPartyMatch(clientName, party string) PartyMatchKind {
	cn := normalizeConflictPartyName(clientName)
	pn := normalizeConflictPartyName(party)
	if !isMeaningfulConflictName(cn) || !isMeaningfulConflictName(pn) {
		return PartyNoMatch
	}
	if cn == pn {
		return PartyExactNormalizedMatch
	}
	if strings.Contains(cn, pn) || strings.Contains(pn, cn) {
		return PartyCandidateMatch
	}
	return PartyNoMatch
}

// isMeaningfulConflictName 规范化后是否仍有有效辨识内容。
// 太短（<2 字符）视为无意义，避免单字"华"命中"华为技术有限公司"。
func isMeaningfulConflictName(name string) bool {
	if len(strings.TrimSpace(name)) < 2 {
		return false
	}
	return true
}

func normalizeConflictPartyName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, suffix := range []string{"有限公司", "股份有限公司", "集团", "控股", "公司", "律所", "律师事务所", " ", "　"} {
		name = strings.ReplaceAll(name, suffix, "")
	}
	return name
}

// GetClientRelations 获取客户关系
func (r *conflictRepository) GetClientRelations(ctx context.Context, clientID string) ([]*models.ClientRelation, error) {
	var relations []*models.ClientRelation

	if err := r.db.WithContext(ctx).
		Where("client_id = ? AND active = ?", clientID, true).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("获取客户关系失败: %w", err)
	}

	return relations, nil
}

// SaveConflictCases 保存冲突案例
func (r *conflictRepository) SaveConflictCases(ctx context.Context, cases []*models.ConflictCase) error {
	if len(cases) == 0 {
		return nil
	}

	// 批量插入
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).CreateInBatches(cases, 100).Error; err != nil {
		return fmt.Errorf("批量保存冲突案例失败: %w", err)
	}

	return nil
}

// GetConflictRules 获取冲突规则
func (r *conflictRepository) GetConflictRules(ctx context.Context, activeOnly bool) ([]*models.ConflictRule, error) {
	var rules []*models.ConflictRule

	query := r.db.WithContext(ctx).Model(&models.ConflictRule{})
	if activeOnly {
		query = query.Where("active = ?", true)
	}

	// 按优先级排序
	query = query.Order("priority DESC")

	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("获取冲突规则失败: %w", err)
	}

	return rules, nil
}

// SaveConflictRule 保存冲突规则
func (r *conflictRepository) SaveConflictRule(ctx context.Context, rule *models.ConflictRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("保存冲突规则失败: %w", err)
	}

	// 清除缓存
	if r.redis != nil {
		r.redis.Del(ctx, "conflict:rules:active")
	}

	return nil
}

// UpdateConflictRule 更新冲突规则
func (r *conflictRepository) UpdateConflictRule(ctx context.Context, rule *models.ConflictRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(rule).Error; err != nil {
		return fmt.Errorf("更新冲突规则失败: %w", err)
	}

	// 清除缓存
	if r.redis != nil {
		r.redis.Del(ctx, "conflict:rules:active")
	}

	return nil
}

// GetMCPStandards 获取MCP标准
func (r *conflictRepository) GetMCPStandards(ctx context.Context, activeOnly bool) (*models.MCPStandards, error) {
	var standards models.MCPStandards

	query := r.db.WithContext(ctx).Model(&models.MCPStandards{})
	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("last_updated DESC").First(&standards).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMCPStandardsNotFound
		}
		return nil, fmt.Errorf("获取MCP标准失败: %w", err)
	}

	return &standards, nil
}

// SaveMCPStandards 保存MCP标准
func (r *conflictRepository) SaveMCPStandards(ctx context.Context, standards *models.MCPStandards) error {
	if err := r.db.WithContext(ctx).Save(standards).Error; err != nil {
		return fmt.Errorf("保存MCP标准失败: %w", err)
	}

	// 清除缓存
	if r.redis != nil {
		r.redis.Del(ctx, "conflict:mcp_standards:active")
	}

	return nil
}

// GetConflictStats 获取统计信息
func (r *conflictRepository) GetConflictStats(ctx context.Context, clientID string) (*ConflictStats, error) {
	stats := &ConflictStats{}

	// 基础查询
	query := r.db.WithContext(ctx).Model(&models.ConflictCheckRecord{})
	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}

	// 总检查次数
	if err := query.Count(&stats.TotalChecks).Error; err != nil {
		return nil, fmt.Errorf("获取总检查次数失败: %w", err)
	}

	// 有冲突的检查次数
	if err := query.Where("has_conflict = ?", true).Count(&stats.ConflictChecks).Error; err != nil {
		return nil, fmt.Errorf("获取冲突检查次数失败: %w", err)
	}

	// 高风险检查次数
	if err := query.Where("risk_level IN ?", []string{"HIGH", "CRITICAL"}).Count(&stats.HighRiskChecks).Error; err != nil {
		return nil, fmt.Errorf("获取高风险检查次数失败: %w", err)
	}

	// 平均持续时间
	var avgDuration float64
	if err := query.Select("AVG(duration)").Scan(&avgDuration).Error; err != nil {
		return nil, fmt.Errorf("获取平均持续时间失败: %w", err)
	}
	stats.AverageDuration = avgDuration

	// 最后检查时间
	var lastCheck time.Time
	if err := query.Select("MAX(check_time)").Scan(&lastCheck).Error; err != nil {
		return nil, fmt.Errorf("获取最后检查时间失败: %w", err)
	}
	stats.LastCheckTime = lastCheck

	return stats, nil
}

// 自定义错误
var (
	ErrMCPStandardsNotFound = errors.New("MCP标准未找到")
)
