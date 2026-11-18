package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"law-oa-go/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	ConflictScopeActive      = "ACTIVE"
	ConflictScopeInactive    = "INACTIVE"
	ConflictCoverageComplete = "COMPLETE"
	ConflictCoverageLimited  = "COVERAGE_LIMITED"
)

// These four scopes are the minimum evidence set for a production conflict
// check. A single "all archives" row is not enough: each source family must
// be reconciled and approved independently so a missing relationship or
// subject registry cannot be hidden by a broad label.
var requiredConflictScopeTypes = []string{
	"CASE_ARCHIVE",
	"CLIENT_ARCHIVE",
	"SUBJECT_REGISTRY",
	"RELATION_ARCHIVE",
}

type ConflictSearchScopeInput struct {
	ID                           string     `json:"id"`
	ScopeType                    string     `json:"scope_type"`
	Status                       string     `json:"status"`
	CoverageStatus               string     `json:"coverage_status"`
	SourceVersion                string     `json:"source_version"`
	EvidenceReference            string     `json:"evidence_reference"`
	CoveredFrom                  *time.Time `json:"covered_from"`
	CoveredTo                    *time.Time `json:"covered_to"`
	MissingSources               []string   `json:"missing_sources"`
	IndexRunID                   string     `json:"index_run_id"`
	SourceOfTruth                bool       `json:"source_of_truth"`
	SyncMode                     string     `json:"sync_mode"`
	MaxSyncLagMinutes            int        `json:"max_sync_lag_minutes"`
	LastSuccessfulSyncAt         *time.Time `json:"last_successful_sync_at"`
	MinimumFieldCoverageBPS      int        `json:"minimum_field_coverage_bps"`
	MeasuredFieldCoverageBPS     int        `json:"measured_field_coverage_bps"`
	MaximumDuplicateRateBPS      int        `json:"maximum_duplicate_rate_bps"`
	MeasuredDuplicateRateBPS     int        `json:"measured_duplicate_rate_bps"`
	QualityOwnerID               *uint      `json:"quality_owner_id"`
	QualityReviewedAt            *time.Time `json:"quality_reviewed_at"`
	MaxQualityReviewAgeDays      int        `json:"max_quality_review_age_days"`
	FailureAlertReference        string     `json:"failure_alert_reference"`
	CorrectionProcedureReference string     `json:"correction_procedure_reference"`
}

type ConflictScopeService struct {
	db *gorm.DB
}

func NewConflictScopeService(db *gorm.DB) *ConflictScopeService {
	return &ConflictScopeService{db: db}
}

func (s *ConflictScopeService) List(ctx context.Context) ([]models.ConflictSearchScope, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("冲突档案覆盖服务未初始化")
	}
	var scopes []models.ConflictSearchScope
	if err := s.db.WithContext(ctx).Order("scope_type ASC").Find(&scopes).Error; err != nil {
		return nil, err
	}
	return scopes, nil
}

func (s *ConflictScopeService) Upsert(ctx context.Context, actor AuthActor, input ConflictSearchScopeInput) (*models.ConflictSearchScope, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("冲突档案覆盖服务未初始化")
	}
	if !IsConflictReviewRole(actor.Role) || actor.UserID == 0 {
		return nil, NewSubjectWorkflowError("SCOPE_ADMIN_REQUIRED", "只有独立冲突核查人或获授权管理合伙人可以维护档案覆盖确认")
	}
	input.ID = strings.TrimSpace(input.ID)
	input.ScopeType = strings.ToUpper(strings.TrimSpace(input.ScopeType))
	input.Status = strings.ToUpper(strings.TrimSpace(input.Status))
	input.CoverageStatus = strings.ToUpper(strings.TrimSpace(input.CoverageStatus))
	input.SourceVersion = strings.TrimSpace(input.SourceVersion)
	input.EvidenceReference = strings.TrimSpace(input.EvidenceReference)
	input.IndexRunID = strings.TrimSpace(input.IndexRunID)
	input.SyncMode = strings.ToUpper(strings.TrimSpace(input.SyncMode))
	input.FailureAlertReference = strings.TrimSpace(input.FailureAlertReference)
	input.CorrectionProcedureReference = strings.TrimSpace(input.CorrectionProcedureReference)
	if input.ID == "" || input.ScopeType == "" {
		return nil, NewSubjectWorkflowError("SCOPE_INPUT_REQUIRED", "档案来源编号和来源类型不能为空")
	}
	if len(input.ID) > 100 || len(input.ScopeType) > 50 || len(input.SourceVersion) > 120 || len(input.EvidenceReference) > 2000 {
		return nil, NewSubjectWorkflowError("SCOPE_INPUT_INVALID", "档案覆盖字段长度超出限制")
	}
	if input.Status == "" {
		input.Status = ConflictScopeActive
	}
	if input.Status != ConflictScopeActive && input.Status != ConflictScopeInactive {
		return nil, NewSubjectWorkflowError("SCOPE_STATUS_INVALID", "档案来源状态不受支持")
	}
	if input.CoverageStatus == "" {
		input.CoverageStatus = ConflictCoverageLimited
	}
	if input.CoverageStatus != ConflictCoverageComplete && input.CoverageStatus != ConflictCoverageLimited {
		return nil, NewSubjectWorkflowError("SCOPE_COVERAGE_INVALID", "档案覆盖状态不受支持")
	}
	missingSources := normalizeScopeSources(input.MissingSources)
	if input.CoverageStatus == ConflictCoverageComplete {
		if input.Status != ConflictScopeActive {
			return nil, NewSubjectWorkflowError("SCOPE_ACTIVE_REQUIRED", "标记为完整覆盖前，请先把来源状态设为当前有效")
		}
		if input.SourceVersion == "" {
			return nil, NewSubjectWorkflowError("SCOPE_SOURCE_VERSION_REQUIRED", "标记为完整覆盖前，必须填写导入对账报告中的来源数据版本")
		}
		if input.EvidenceReference == "" {
			return nil, NewSubjectWorkflowError("SCOPE_EVIDENCE_REQUIRED", "标记为完整覆盖前，必须填写可核验的核对凭证引用")
		}
		if input.IndexRunID == "" {
			return nil, NewSubjectWorkflowError("SCOPE_INDEX_RUN_REQUIRED", "标记为完整覆盖前，必须填写导入对账报告中的索引构建运行编号")
		}
		if len(missingSources) > 0 {
			return nil, NewSubjectWorkflowError("SCOPE_MISSING_SOURCE_CONFLICT", fmt.Sprintf("仍登记有未纳入资料：%s；覆盖结论只能选择覆盖受限", strings.Join(missingSources, "、")))
		}
		if input.CoveredFrom == nil || input.CoveredTo == nil || input.CoveredFrom.After(*input.CoveredTo) {
			return nil, NewSubjectWorkflowError("SCOPE_COVERAGE_RANGE_REQUIRED", "完整覆盖必须明确覆盖起止时间")
		}
		if err := s.validateIndexRun(ctx, input.IndexRunID, input.ScopeType, input.SourceVersion, input.EvidenceReference); err != nil {
			return nil, err
		}
		if err := validateScopeQualityInput(input, time.Now()); err != nil {
			return nil, err
		}
	}

	var scope models.ConflictSearchScope
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", input.ID).First(&scope).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			scope = models.ConflictSearchScope{ID: input.ID, CreatedAt: time.Now()}
		} else if err != nil {
			return err
		}
		if input.Status == ConflictScopeActive {
			var duplicateCount int64
			if err := tx.Model(&models.ConflictSearchScope{}).
				Where("scope_type = ? AND status = ? AND id <> ?", input.ScopeType, ConflictScopeActive, input.ID).
				Count(&duplicateCount).Error; err != nil {
				return err
			}
			if duplicateCount > 0 {
				return NewSubjectWorkflowError("SCOPE_DUPLICATE_ACTIVE", "同一档案来源类型只能有一个 ACTIVE 配置，请先停用旧版本")
			}
		}
		fromStatus := scope.CoverageStatus
		scope.ScopeType = input.ScopeType
		scope.Status = input.Status
		scope.CoverageStatus = input.CoverageStatus
		scope.SourceVersion = input.SourceVersion
		scope.EvidenceReference = input.EvidenceReference
		scope.CoveredFrom = input.CoveredFrom
		scope.CoveredTo = input.CoveredTo
		scope.MissingSources = scopeSourcesJSON(missingSources)
		scope.IndexRunID = input.IndexRunID
		scope.SourceOfTruth = input.SourceOfTruth
		scope.SyncMode = input.SyncMode
		scope.MaxSyncLagMinutes = input.MaxSyncLagMinutes
		scope.LastSuccessfulSyncAt = input.LastSuccessfulSyncAt
		scope.MinimumFieldCoverageBPS = input.MinimumFieldCoverageBPS
		scope.MeasuredFieldCoverageBPS = input.MeasuredFieldCoverageBPS
		scope.MaximumDuplicateRateBPS = input.MaximumDuplicateRateBPS
		scope.MeasuredDuplicateRateBPS = input.MeasuredDuplicateRateBPS
		scope.QualityOwnerID = input.QualityOwnerID
		scope.QualityReviewedAt = input.QualityReviewedAt
		scope.MaxQualityReviewAgeDays = input.MaxQualityReviewAgeDays
		scope.FailureAlertReference = input.FailureAlertReference
		scope.CorrectionProcedureReference = input.CorrectionProcedureReference
		scope.ApprovedBy = nil
		scope.ApprovedAt = nil
		if input.CoverageStatus == ConflictCoverageComplete {
			now := time.Now()
			scope.ApprovedBy = &actor.UserID
			scope.ApprovedAt = &now
		}
		scope.UpdatedAt = time.Now()
		if scope.CreatedAt.IsZero() {
			scope.CreatedAt = time.Now()
		}
		if err := tx.Save(&scope).Error; err != nil {
			return err
		}
		payload := map[string]interface{}{
			"scope_type":                     scope.ScopeType,
			"status":                         scope.Status,
			"coverage_status":                scope.CoverageStatus,
			"source_version":                 scope.SourceVersion,
			"evidence_reference":             scope.EvidenceReference,
			"index_run_id":                   scope.IndexRunID,
			"missing_sources":                missingSources,
			"source_of_truth":                scope.SourceOfTruth,
			"sync_mode":                      scope.SyncMode,
			"max_sync_lag_minutes":           scope.MaxSyncLagMinutes,
			"last_successful_sync_at":        scope.LastSuccessfulSyncAt,
			"minimum_field_coverage_bps":     scope.MinimumFieldCoverageBPS,
			"measured_field_coverage_bps":    scope.MeasuredFieldCoverageBPS,
			"maximum_duplicate_rate_bps":     scope.MaximumDuplicateRateBPS,
			"measured_duplicate_rate_bps":    scope.MeasuredDuplicateRateBPS,
			"quality_owner_id":               scope.QualityOwnerID,
			"quality_reviewed_at":            scope.QualityReviewedAt,
			"max_quality_review_age_days":    scope.MaxQualityReviewAgeDays,
			"failure_alert_reference":        scope.FailureAlertReference,
			"correction_procedure_reference": scope.CorrectionProcedureReference,
		}
		raw, _ := json.Marshal(payload)
		sum := sha256.Sum256(raw)
		actorID := actor.UserID
		audit := &models.ComplianceAuditEvent{
			ID: uuid.NewString(), ActorID: &actorID, ActorRole: actor.Role,
			EventType: "CONFLICT_SCOPE_UPDATED", ObjectType: "CONFLICT_SEARCH_SCOPE", ObjectID: scope.ID,
			FromState: fromStatus, ToState: scope.CoverageStatus, Payload: string(raw),
			IntegrityHash: hex.EncodeToString(sum[:]), CreatedAt: time.Now(),
		}
		return tx.Create(audit).Error
	})
	if err != nil {
		return nil, err
	}
	return &scope, nil
}

func normalizeScopeSources(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func scopeSourcesJSON(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func (s *ConflictScopeService) ValidateComplete(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("冲突档案覆盖服务未初始化")
	}
	var scopes []models.ConflictSearchScope
	if err := s.db.WithContext(ctx).Where("status = ?", ConflictScopeActive).Find(&scopes).Error; err != nil {
		return err
	}
	if len(scopes) == 0 {
		return fmt.Errorf("冲突档案覆盖未完成: active=0 complete=0")
	}
	for _, scope := range scopes {
		if !scopeHasCompleteMetadata(scope) {
			return fmt.Errorf("冲突档案覆盖未完成: scope=%s", scope.ID)
		}
		if err := s.validateIndexRun(ctx, scope.IndexRunID, scope.ScopeType, scope.SourceVersion, scope.EvidenceReference); err != nil {
			return err
		}
	}
	for _, scopeType := range requiredConflictScopeTypes {
		found := false
		for _, scope := range scopes {
			if scope.ScopeType == scopeType && scopeHasCompleteMetadata(scope) {
				if err := s.validateIndexRun(ctx, scope.IndexRunID, scope.ScopeType, scope.SourceVersion, scope.EvidenceReference); err == nil {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("冲突档案覆盖缺少必需来源: %s", scopeType)
		}
	}
	return nil
}

func scopeHasCompleteMetadata(scope models.ConflictSearchScope) bool {
	return scope.Status == ConflictScopeActive && scope.CoverageStatus == ConflictCoverageComplete &&
		strings.TrimSpace(scope.SourceVersion) != "" && strings.TrimSpace(scope.EvidenceReference) != "" &&
		scope.CoveredFrom != nil && scope.CoveredTo != nil && !scope.CoveredFrom.After(*scope.CoveredTo) &&
		(strings.TrimSpace(scope.MissingSources) == "" || strings.TrimSpace(scope.MissingSources) == "[]") &&
		strings.TrimSpace(scope.IndexRunID) != "" && scopeQualityIsCurrent(scope, time.Now())
}

func validateScopeQualityInput(input ConflictSearchScopeInput, now time.Time) error {
	if !input.SourceOfTruth {
		return NewSubjectWorkflowError("SCOPE_SOURCE_OF_TRUTH_REQUIRED", "完整覆盖必须明确该来源是经律所确认的权威来源")
	}
	if !map[string]bool{"REALTIME": true, "BATCH": true, "MANUAL_IMPORT": true}[input.SyncMode] {
		return NewSubjectWorkflowError("SCOPE_SYNC_MODE_REQUIRED", "完整覆盖必须记录同步方式：REALTIME、BATCH 或 MANUAL_IMPORT")
	}
	if input.MaxSyncLagMinutes <= 0 || input.LastSuccessfulSyncAt == nil ||
		input.LastSuccessfulSyncAt.After(now.Add(5*time.Minute)) ||
		now.After(input.LastSuccessfulSyncAt.Add(time.Duration(input.MaxSyncLagMinutes)*time.Minute)) {
		return NewSubjectWorkflowError("SCOPE_SOURCE_STALE", "权威数据源未同步或已超过律所批准的最大允许延迟")
	}
	if input.MinimumFieldCoverageBPS <= 0 || input.MinimumFieldCoverageBPS > 10000 ||
		input.MeasuredFieldCoverageBPS < input.MinimumFieldCoverageBPS || input.MeasuredFieldCoverageBPS > 10000 {
		return NewSubjectWorkflowError("SCOPE_FIELD_COVERAGE_INSUFFICIENT", "权威数据源字段覆盖率未达到律所批准阈值")
	}
	if input.MaximumDuplicateRateBPS < 0 || input.MaximumDuplicateRateBPS > 10000 ||
		input.MeasuredDuplicateRateBPS < 0 || input.MeasuredDuplicateRateBPS > input.MaximumDuplicateRateBPS {
		return NewSubjectWorkflowError("SCOPE_DUPLICATE_RATE_EXCEEDED", "权威数据源重复率超过律所批准阈值")
	}
	if input.QualityOwnerID == nil || *input.QualityOwnerID == 0 || input.QualityReviewedAt == nil ||
		input.MaxQualityReviewAgeDays <= 0 || input.QualityReviewedAt.After(now.Add(5*time.Minute)) ||
		now.After(input.QualityReviewedAt.AddDate(0, 0, input.MaxQualityReviewAgeDays)) {
		return NewSubjectWorkflowError("SCOPE_QUALITY_REVIEW_REQUIRED", "完整覆盖必须记录数据质量责任人和最近核对时间")
	}
	if input.FailureAlertReference == "" || input.CorrectionProcedureReference == "" {
		return NewSubjectWorkflowError("SCOPE_QUALITY_PROCEDURE_REQUIRED", "完整覆盖必须引用同步失败告警和人工更正流程")
	}
	return nil
}

func scopeQualityIsCurrent(scope models.ConflictSearchScope, now time.Time) bool {
	if !scope.SourceOfTruth || scope.MaxSyncLagMinutes <= 0 || scope.LastSuccessfulSyncAt == nil ||
		scope.LastSuccessfulSyncAt.After(now.Add(5*time.Minute)) ||
		now.After(scope.LastSuccessfulSyncAt.Add(time.Duration(scope.MaxSyncLagMinutes)*time.Minute)) {
		return false
	}
	if !map[string]bool{"REALTIME": true, "BATCH": true, "MANUAL_IMPORT": true}[strings.ToUpper(strings.TrimSpace(scope.SyncMode))] {
		return false
	}
	if scope.MinimumFieldCoverageBPS <= 0 || scope.MinimumFieldCoverageBPS > 10000 ||
		scope.MeasuredFieldCoverageBPS < scope.MinimumFieldCoverageBPS || scope.MeasuredFieldCoverageBPS > 10000 {
		return false
	}
	if scope.MaximumDuplicateRateBPS < 0 || scope.MaximumDuplicateRateBPS > 10000 ||
		scope.MeasuredDuplicateRateBPS < 0 || scope.MeasuredDuplicateRateBPS > scope.MaximumDuplicateRateBPS {
		return false
	}
	return scope.QualityOwnerID != nil && *scope.QualityOwnerID > 0 && scope.QualityReviewedAt != nil &&
		scope.MaxQualityReviewAgeDays > 0 && !scope.QualityReviewedAt.After(now.Add(5*time.Minute)) &&
		!now.After(scope.QualityReviewedAt.AddDate(0, 0, scope.MaxQualityReviewAgeDays)) &&
		strings.TrimSpace(scope.FailureAlertReference) != "" && strings.TrimSpace(scope.CorrectionProcedureReference) != ""
}

func (s *ConflictScopeService) validateIndexRun(ctx context.Context, runID, scopeType, sourceVersion, evidenceReference string) error {
	var run models.ConflictIndexBuildRun
	if err := s.db.WithContext(ctx).Where("id = ?", runID).First(&run).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewSubjectWorkflowError("SCOPE_INDEX_RUN_REQUIRED", "完整覆盖必须绑定已完成的索引对账运行记录")
		}
		return fmt.Errorf("读取索引对账运行记录失败: %w", err)
	}
	if run.Status != models.ConflictIndexBuildComplete || run.ScopeType != scopeType || run.SourceVersion != sourceVersion ||
		run.MissingRecordCount != 0 || run.IndexedRecordCount < run.SourceRecordCount ||
		strings.TrimSpace(run.ReconciliationHash) == "" || strings.TrimSpace(run.EvidenceReference) == "" {
		return NewSubjectWorkflowError("SCOPE_INDEX_RUN_INVALID", "索引对账运行记录未完成、存在缺口或与当前档案覆盖配置不一致")
	}
	if strings.TrimSpace(evidenceReference) != strings.TrimSpace(run.EvidenceReference) {
		return NewSubjectWorkflowError("SCOPE_INDEX_RUN_EVIDENCE_MISMATCH", "档案覆盖凭证与索引对账运行记录不一致")
	}
	return nil
}
