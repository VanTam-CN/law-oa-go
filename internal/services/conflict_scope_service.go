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
	ID                string     `json:"id"`
	ScopeType         string     `json:"scope_type"`
	Status            string     `json:"status"`
	CoverageStatus    string     `json:"coverage_status"`
	SourceVersion     string     `json:"source_version"`
	EvidenceReference string     `json:"evidence_reference"`
	CoveredFrom       *time.Time `json:"covered_from"`
	CoveredTo         *time.Time `json:"covered_to"`
	MissingSources    []string   `json:"missing_sources"`
	IndexRunID        string     `json:"index_run_id"`
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
		if input.Status != ConflictScopeActive || input.SourceVersion == "" || input.EvidenceReference == "" || input.IndexRunID == "" || len(missingSources) > 0 {
			return nil, NewSubjectWorkflowError("SCOPE_COMPLETION_EVIDENCE_REQUIRED", "只有绑定索引构建凭证、有版本、有核对凭证且没有登记缺口的有效档案来源才能标记为完整覆盖")
		}
		if input.CoveredFrom == nil || input.CoveredTo == nil || input.CoveredFrom.After(*input.CoveredTo) {
			return nil, NewSubjectWorkflowError("SCOPE_COVERAGE_RANGE_REQUIRED", "完整覆盖必须明确覆盖起止时间")
		}
		if err := s.validateIndexRun(ctx, input.IndexRunID, input.ScopeType, input.SourceVersion, input.EvidenceReference); err != nil {
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
			"scope_type":         scope.ScopeType,
			"status":             scope.Status,
			"coverage_status":    scope.CoverageStatus,
			"source_version":     scope.SourceVersion,
			"evidence_reference": scope.EvidenceReference,
			"index_run_id":       scope.IndexRunID,
			"missing_sources":    missingSources,
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
		strings.TrimSpace(scope.IndexRunID) != ""
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
