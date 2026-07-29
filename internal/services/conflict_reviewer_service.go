package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

// ConflictReviewerError is returned for a fail-closed reviewer decision. The
// code is safe to expose to the UI; the message explains the next permitted
// workflow without disclosing protected historical matter details.
type ConflictReviewerError struct {
	Code    string
	Message string
}

func (e *ConflictReviewerError) Error() string {
	if e == nil {
		return "冲突复核权限校验失败"
	}
	return e.Message
}

func newConflictReviewerError(code, message string) error {
	return &ConflictReviewerError{Code: code, Message: message}
}

// ValidateConflictReviewer is the server-side independence gate shared by
// direct conflict conclusions and subject-revision approvals. Role checks,
// appointment checks, account status, applicant/owner checks, and direct
// management recusal are all evaluated from database state.
func ValidateConflictReviewer(ctx context.Context, db *gorm.DB, checkID string, caseID, reviewerID uint, claimedRole string) error {
	if db == nil {
		return newConflictReviewerError("REVIEWER_GATE_UNAVAILABLE", "独立复核服务未初始化，已阻止提交")
	}
	checkID = strings.TrimSpace(checkID)
	if checkID == "" || reviewerID == 0 {
		return newConflictReviewerError("REVIEWER_REQUIRED", "冲突检测记录和复核人不能为空")
	}

	var reviewer models.User
	if err := db.WithContext(ctx).
		Select("id", "role", "status", "manager_id").
		Where("id = ? AND deleted_at IS NULL", reviewerID).
		First(&reviewer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return newConflictReviewerError("REVIEWER_NOT_FOUND", "独立复核账号不存在")
		}
		return fmt.Errorf("读取独立复核账号失败: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(reviewer.Status), "active") {
		return newConflictReviewerError("REVIEWER_INACTIVE", "独立复核账号未处于启用状态")
	}
	if !IsConflictReviewRole(reviewer.Role) {
		return newConflictReviewerError("REVIEWER_ROLE_FORBIDDEN", "该账号不是获授权的业务冲突复核角色")
	}
	if claimedRole != "" && !strings.EqualFold(strings.TrimSpace(claimedRole), strings.TrimSpace(reviewer.Role)) {
		return newConflictReviewerError("REVIEWER_ROLE_STALE", "当前登录身份与复核账号权限不一致，请重新登录")
	}

	var record models.ConflictCheckRecord
	if err := db.WithContext(ctx).Where("check_id = ?", checkID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return newConflictReviewerError("CHECK_NOT_FOUND", "冲突检测记录不存在")
		}
		return fmt.Errorf("读取冲突检测记录失败: %w", err)
	}
	if record.UserID == reviewerID {
		return newConflictReviewerError("REVIEWER_CONFLICTED", "申请律师不得复核本人发起的冲突检测")
	}

	subject, contextErr := ResolveConflictSubjectContext(ctx, db, &record)
	if contextErr != nil {
		return fmt.Errorf("读取冲突检测案件上下文失败: %w", contextErr)
	}
	if subject.CaseID == 0 && subject.IntakeID == "" {
		return newConflictReviewerError("CASE_CONTEXT_REQUIRED", "冲突检测缺少可验证的案件或接案上下文，不能指定或提交复核")
	}
	if caseID > 0 && subject.CaseID > 0 && caseID != subject.CaseID {
		return newConflictReviewerError("CASE_CONTEXT_MISMATCH", "复核案件上下文与检测记录不一致，已阻止复核")
	}
	if caseID == 0 {
		caseID = subject.CaseID
	}
	involvedIDs := []uint{record.UserID}
	if caseID > 0 {
		var caseModel models.Case
		if err := db.WithContext(ctx).Select("id", "lawyer_id", "created_by").Where("id = ? AND deleted_at IS NULL", caseID).First(&caseModel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newConflictReviewerError("CASE_CONTEXT_UNAVAILABLE", "无法确认冲突检测对应的案件负责人，已阻止复核")
			}
			return fmt.Errorf("读取案件复核关系失败: %w", err)
		}
		if caseModel.LawyerID > 0 {
			involvedIDs = append(involvedIDs, caseModel.LawyerID)
		}
		if createdBy, err := strconv.ParseUint(strings.TrimSpace(caseModel.CreatedBy), 10, 32); err == nil && createdBy > 0 {
			involvedIDs = append(involvedIDs, uint(createdBy))
		}
	} else if subject.IntakeID != "" {
		if ownerID, err := conflictIntakeOwnerID(ctx, db, subject.IntakeID); err != nil {
			return fmt.Errorf("读取接案复核关系失败: %w", err)
		} else if ownerID > 0 {
			involvedIDs = append(involvedIDs, ownerID)
		}
	}
	if hasDirectManagementConflict(db.WithContext(ctx), reviewer, involvedIDs) {
		return newConflictReviewerError("REVIEWER_CONFLICTED", "复核人与申请人或承办律师存在直接管理关系，必须回避")
	}

	var assignments []models.ConflictReviewerAssignment
	if err := db.WithContext(ctx).
		Where("check_id = ? AND reviewer_id = ? AND status = ?", checkID, reviewerID, models.ConflictReviewerAssignmentActive).
		Where("effective_from IS NULL OR effective_from <= ?", time.Now()).
		Where("effective_to IS NULL OR effective_to > ?", time.Now()).
		Order("created_at DESC").Find(&assignments).Error; err != nil {
		return fmt.Errorf("读取冲突复核指定失败: %w", err)
	}
	if len(assignments) == 0 {
		return newConflictReviewerError("REVIEWER_ASSIGNMENT_REQUIRED", "该冲突检测尚未指定独立复核人，请由冲突核查岗或管理合伙人先指定")
	}
	if !assignments[0].RecusalDeclared {
		return newConflictReviewerError("REVIEWER_RECUSAL_REQUIRED", "复核人尚未完成回避与独立性声明，不能提交结论")
	}
	return nil
}

// ConflictReviewerAssignmentInput is intentionally small. Assignment
// authority, target role, and independence are validated on the server.
type ConflictReviewerAssignmentInput struct {
	ReviewerID         uint       `json:"reviewer_id" binding:"required"`
	CaseID             *uint      `json:"case_id,omitempty"`
	DelegateForID      *uint      `json:"delegate_for_id,omitempty"`
	RecusalDeclared    bool       `json:"recusal_declared"`
	IndependenceReason string     `json:"independence_reason"`
	SLADueAt           *time.Time `json:"sla_due_at,omitempty"`
	EffectiveFrom      *time.Time `json:"effective_from,omitempty"`
	EffectiveTo        *time.Time `json:"effective_to,omitempty"`
}

// AssignConflictReviewer creates an auditable, non-destructive appointment.
// Existing active assignments are returned unchanged, making retries safe.
func AssignConflictReviewer(ctx context.Context, db *gorm.DB, actor AuthActor, checkID string, input ConflictReviewerAssignmentInput) (*models.ConflictReviewerAssignment, error) {
	if db == nil {
		return nil, newConflictReviewerError("REVIEWER_GATE_UNAVAILABLE", "独立复核服务未初始化")
	}
	if actor.UserID == 0 || (!IsConflictReviewRole(actor.Role) && !IsBusinessMatterManagementRole(actor.Role)) {
		return nil, newConflictReviewerError("REVIEWER_ASSIGNMENT_FORBIDDEN", "只有冲突核查岗或业务管理合伙人可以指定复核人")
	}
	checkID = strings.TrimSpace(checkID)
	if checkID == "" || input.ReviewerID == 0 {
		return nil, newConflictReviewerError("REVIEWER_ASSIGNMENT_REQUIRED", "冲突检测记录和复核人不能为空")
	}
	if !input.RecusalDeclared {
		return nil, newConflictReviewerError("REVIEWER_RECUSAL_REQUIRED", "指定复核人前必须完成回避与独立性声明")
	}

	var result models.ConflictReviewerAssignment
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record models.ConflictCheckRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("check_id = ?", checkID).First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newConflictReviewerError("CHECK_NOT_FOUND", "冲突检测记录不存在")
			}
			return err
		}
		var reviewer models.User
		if err := tx.Select("id", "role", "status", "manager_id").Where("id = ? AND deleted_at IS NULL", input.ReviewerID).First(&reviewer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newConflictReviewerError("REVIEWER_NOT_FOUND", "指定的复核账号不存在")
			}
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(reviewer.Status), "active") || !IsConflictReviewRole(reviewer.Role) {
			return newConflictReviewerError("REVIEWER_ROLE_FORBIDDEN", "指定对象不是启用中的业务冲突复核角色")
		}
		if record.UserID == input.ReviewerID {
			return newConflictReviewerError("REVIEWER_CONFLICTED", "申请律师不得被指定为本人检测记录的复核人")
		}
		subject, contextErr := ResolveConflictSubjectContext(ctx, tx, &record)
		if contextErr != nil {
			return fmt.Errorf("读取冲突检测案件上下文失败: %w", contextErr)
		}
		if subject.CaseID == 0 && subject.IntakeID == "" {
			return newConflictReviewerError("CASE_CONTEXT_REQUIRED", "冲突检测缺少可验证的案件或接案上下文，不能指定复核人")
		}
		caseID := subject.CaseID
		if input.CaseID != nil {
			if subject.CaseID == 0 || *input.CaseID != subject.CaseID {
				return newConflictReviewerError("CASE_CONTEXT_MISMATCH", "指定的案件上下文与检测记录不一致，已阻止指定复核人")
			}
			caseID = *input.CaseID
		}
		involvedIDs := []uint{record.UserID}
		if caseID > 0 {
			var caseModel models.Case
			if err := tx.Select("id", "lawyer_id", "created_by").Where("id = ? AND deleted_at IS NULL", caseID).First(&caseModel).Error; err != nil {
				return newConflictReviewerError("CASE_CONTEXT_UNAVAILABLE", "无法确认案件负责人，已阻止指定复核人")
			}
			if caseModel.LawyerID > 0 {
				involvedIDs = append(involvedIDs, caseModel.LawyerID)
			}
			if createdBy, err := strconv.ParseUint(strings.TrimSpace(caseModel.CreatedBy), 10, 32); err == nil && createdBy > 0 {
				involvedIDs = append(involvedIDs, uint(createdBy))
			}
		} else if subject.IntakeID != "" {
			if ownerID, err := conflictIntakeOwnerID(ctx, tx, subject.IntakeID); err != nil {
				return fmt.Errorf("读取接案复核关系失败: %w", err)
			} else if ownerID > 0 {
				involvedIDs = append(involvedIDs, ownerID)
			}
		}
		if hasDirectManagementConflict(tx, reviewer, involvedIDs) {
			return newConflictReviewerError("REVIEWER_CONFLICTED", "指定复核人与申请人或承办律师存在直接管理关系，必须回避")
		}

		var existing models.ConflictReviewerAssignment
		if err := tx.Where("check_id = ? AND reviewer_id = ? AND status = ?", checkID, input.ReviewerID, models.ConflictReviewerAssignmentActive).Order("created_at DESC").First(&existing).Error; err == nil {
			result = existing
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()
		effectiveFrom := input.EffectiveFrom
		if effectiveFrom == nil {
			effectiveFrom = &now
		}
		result = models.ConflictReviewerAssignment{
			ID: uuid.NewString(), CheckID: checkID, CaseID: input.CaseID,
			ReviewerID: input.ReviewerID, DelegateForID: input.DelegateForID, AssignedBy: actor.UserID,
			Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true,
			IndependenceReason: strings.TrimSpace(input.IndependenceReason), SLADueAt: input.SLADueAt,
			EffectiveFrom: effectiveFrom, EffectiveTo: input.EffectiveTo, CreatedAt: now, UpdatedAt: now,
		}
		if result.CaseID == nil && caseID > 0 {
			result.CaseID = &caseID
		}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		payload := map[string]interface{}{
			"check_id": checkID, "reviewer_id": input.ReviewerID, "delegate_for_id": input.DelegateForID,
			"recusal_declared": true, "independence_reason": result.IndependenceReason,
		}
		raw, _ := json.Marshal(payload)
		sum := sha256.Sum256(raw)
		actorID := actor.UserID
		return tx.Create(&models.ComplianceAuditEvent{
			ID: uuid.NewString(), ActorID: &actorID, ActorRole: actor.Role,
			EventType: "CONFLICT_REVIEWER_ASSIGNED", ObjectType: "CONFLICT_REVIEWER_ASSIGNMENT", ObjectID: result.ID,
			Payload: string(raw), IntegrityHash: hex.EncodeToString(sum[:]), CreatedAt: now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func GetActiveConflictReviewerAssignment(ctx context.Context, db *gorm.DB, checkID string) (*models.ConflictReviewerAssignment, error) {
	if db == nil {
		return nil, newConflictReviewerError("REVIEWER_GATE_UNAVAILABLE", "独立复核服务未初始化")
	}
	var assignment models.ConflictReviewerAssignment
	err := db.WithContext(ctx).
		Where("check_id = ? AND status = ?", strings.TrimSpace(checkID), models.ConflictReviewerAssignmentActive).
		Where("effective_from IS NULL OR effective_from <= ?", time.Now()).
		Where("effective_to IS NULL OR effective_to > ?", time.Now()).
		Order("created_at DESC").First(&assignment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取冲突复核指定失败: %w", err)
	}
	return &assignment, nil
}

func hasDirectManagementConflict(db *gorm.DB, reviewer models.User, involvedIDs []uint) bool {
	seen := make(map[uint]struct{}, len(involvedIDs))
	for _, id := range involvedIDs {
		if id == 0 {
			continue
		}
		if id == reviewer.ID {
			return true
		}
		seen[id] = struct{}{}
	}
	if reviewer.ManagerID != nil {
		if _, ok := seen[*reviewer.ManagerID]; ok {
			return true
		}
	}
	if len(seen) == 0 {
		return false
	}
	var involved []models.User
	if err := db.Select("id", "manager_id").Where("id IN ? AND deleted_at IS NULL", mapKeys(seen)).Find(&involved).Error; err != nil {
		// A failed independence query must fail closed. The caller cannot
		// distinguish this boolean-only helper, so treat the unknown relation as
		// conflicted rather than allowing a review through.
		return true
	}
	for _, user := range involved {
		if user.ManagerID != nil && *user.ManagerID == reviewer.ID {
			return true
		}
	}
	return false
}

func mapKeys(values map[uint]struct{}) []uint {
	keys := make([]uint, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func conflictIntakeOwnerID(ctx context.Context, db *gorm.DB, intakeID string) (uint, error) {
	if db == nil || strings.TrimSpace(intakeID) == "" {
		return 0, nil
	}
	var intake map[string]interface{}
	if err := db.WithContext(ctx).Table("case_intakes").Select("created_by").Where("id = ?", strings.TrimSpace(intakeID)).Take(&intake).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, newConflictReviewerError("CASE_CONTEXT_UNAVAILABLE", "无法确认接案负责人，已阻止复核")
		}
		return 0, err
	}
	ownerID, _ := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(intake["created_by"])), 10, 32)
	return uint(ownerID), nil
}

func subjectCaseIDFromSearchParameters(parameters models.JSON) uint {
	if len(parameters) == 0 {
		return 0
	}
	var values struct {
		SubjectCaseID      string `json:"subjectCaseId"`
		SubjectCaseIDSnake string `json:"subject_case_id"`
	}
	raw, err := json.Marshal(parameters)
	if err != nil || json.Unmarshal(raw, &values) != nil {
		return 0
	}
	caseIDValue := strings.TrimSpace(values.SubjectCaseID)
	if caseIDValue == "" {
		caseIDValue = strings.TrimSpace(values.SubjectCaseIDSnake)
	}
	caseID, err := strconv.ParseUint(caseIDValue, 10, 32)
	if err != nil {
		return 0
	}
	return uint(caseID)
}

func recordedConflictCaseID(ctx context.Context, db *gorm.DB, record *models.ConflictCheckRecord) (uint, error) {
	if record == nil {
		return 0, nil
	}
	if caseID := subjectCaseIDFromSearchParameters(record.SearchParameters); caseID > 0 {
		return caseID, nil
	}
	if db == nil || strings.TrimSpace(record.CheckID) == "" {
		return 0, nil
	}
	var caseIDs []uint
	if err := db.WithContext(ctx).
		Table("cases").
		Where("conflict_check_id = ? AND deleted_at IS NULL", record.CheckID).
		Limit(2).
		Pluck("id", &caseIDs).Error; err != nil {
		return 0, err
	}
	if len(caseIDs) != 1 {
		return 0, nil
	}
	return caseIDs[0], nil
}
