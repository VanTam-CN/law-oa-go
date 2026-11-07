package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// All lifecycle entry points lock the approval before reading its current
// version. The transaction-bound service also binds assignment and record I/O.
func (s *ApprovalService) approvalTransaction(id string, run func(*ApprovalService, *models.ApprovalRequest) error) error {
	return s.approvalRepo.DB().Transaction(func(tx *gorm.DB) error {
		var approval models.ApprovalRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&approval, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrApprovalRecordNotFound
			}
			return err
		}
		return run(NewApprovalService(tx), &approval)
	})
}

func reconcileApprovalTransition(tx *gorm.DB, approval *models.ApprovalRequest) error {
	// Legacy embedded fixtures have no inbox. A protocol-enabled approval must
	// never commit a transition without its required workflow handoff schema.
	if !tx.Migrator().HasTable("inbox_items") && approval.ActiveSubmissionSnapshotID == nil {
		return nil
	}
	var persisted models.ApprovalRequest
	if err := tx.First(&persisted, "id = ?", approval.ID).Error; err != nil {
		return err
	}
	return ReconcileApprovalInboxInTx(tx, &persisted)
}

func (s *ApprovalService) SubmitApproval(userID, id string) (*models.ApprovalRequest, error) {
	var result *models.ApprovalRequest
	err := s.approvalTransaction(id, func(bound *ApprovalService, locked *models.ApprovalRequest) error {
		if locked.ActiveSubmissionSnapshotID != nil {
			return &ApprovalConflictGateError{Code: "INTAKE_REVISION_ROUTE_REQUIRED", Message: "已有提交快照，请从接案补充页面重新提交"}
		}
		var err error
		result, err = bound.submitApprovalInTx(userID, id)
		if err != nil {
			return err
		}
		return reconcileApprovalTransition(bound.approvalRepo.DB(), result)
	})
	return result, err
}

func (s *ApprovalService) ProcessApprovalDecision(userID, id string, req *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	var result *models.ApprovalRequest
	err := s.approvalTransaction(id, func(bound *ApprovalService, _ *models.ApprovalRequest) error {
		var err error
		result, err = bound.processApprovalDecisionInTx(userID, id, req)
		if err != nil {
			return err
		}
		return reconcileApprovalTransition(bound.approvalRepo.DB(), result)
	})
	return result, err
}

// ResubmitApproval preserves the old internal call shape for non-versioned
// approvals. HTTP intake resubmissions use the explicit versioned DTO below.
func (s *ApprovalService) ResubmitApproval(userID, id, note string) (*models.ApprovalRequest, error) {
	var result *models.ApprovalRequest
	err := s.approvalTransaction(id, func(bound *ApprovalService, _ *models.ApprovalRequest) error {
		var err error
		result, err = bound.resubmitApprovalInTx(userID, id, note)
		if err != nil {
			return err
		}
		return reconcileApprovalTransition(bound.approvalRepo.DB(), result)
	})
	return result, err
}

func (s *ApprovalService) ResubmitIntakeApproval(userID, id string, req *models.ResubmitApprovalRequest) (*models.ApprovalRequest, error) {
	var result *models.ApprovalRequest
	err := s.approvalTransaction(id, func(bound *ApprovalService, approval *models.ApprovalRequest) error {
		if approval.ApplicantID != userID {
			return ErrApprovalForbidden
		}
		if err := AuthorizeApprovalResubmission(bound.approvalRepo.DB(), userID, approval); err != nil {
			return err
		}
		tx := bound.approvalRepo.DB()
		replayed, err := ValidateResubmissionReplayInTx(tx, approval, req)
		if err != nil {
			return err
		}
		if replayed {
			result, err = bound.GetApproval(userID, id)
			return err
		}
		if approval.Status != models.ApprovalStatusNeedsRevision {
			return &ApprovalConflictGateError{Code: "INTAKE_RESUBMIT_STATE_CONFLICT", Message: "当前审批不在待补充状态，请刷新审批详情"}
		}
		if err := FreezeIntakeSubmissionInTx(tx, approval, userID, req); err != nil {
			return err
		}
		if err := bound.terminalBinding.rebindAfterResubmitInTx(tx, approval); err != nil {
			if asUniqueConstraintError(err) != nil {
				return &ApprovalConflictGateError{Code: "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", Message: "该接案已有进行中的审批，不能重新提交"}
			}
			return err
		}
		if err := bound.assigner.AssignApprover(approval); err != nil {
			return err
		}
		approval.Status = models.ApprovalStatusResubmitted
		approval.UpdatedBy = userID
		now := time.Now().UTC()
		approval.SubmissionDate = &now
		if err := tx.Omit("Records", "Notifications").Save(approval).Error; err != nil {
			return err
		}
		if err := reconcileApprovalTransition(tx, approval); err != nil {
			return err
		}
		result, err = bound.GetApproval(userID, id)
		return err
	})
	return result, err
}

func (s *ApprovalService) CancelApproval(userID, id string) error {
	return s.approvalTransaction(id, func(bound *ApprovalService, _ *models.ApprovalRequest) error {
		if err := bound.cancelApprovalInTx(userID, id); err != nil {
			return err
		}
		approval, err := bound.approvalRepo.FindByID(id)
		if err != nil {
			return err
		}
		return reconcileApprovalTransition(bound.approvalRepo.DB(), approval)
	})
}

func validateApprovalDecisionVersion(approval *models.ApprovalRequest, req *models.ApprovalDecisionRequest) error {
	if req == nil {
		return fmt.Errorf("缺少审批决定")
	}
	if approval.ActiveSubmissionSnapshotID != nil && (req.ExpectedSubmissionSnapshotID != *approval.ActiveSubmissionSnapshotID || req.ExpectedActionVersion == "") {
		return &ApprovalConflictGateError{Code: "APPROVAL_SUBMISSION_VERSION_CONFLICT", Message: "审批材料或轮次已经变化，请刷新后重新审查"}
	}
	if req.ExpectedActionVersion != "" && req.ExpectedActionVersion != approval.UpdatedAt.UTC().Format(time.RFC3339Nano) {
		return &ApprovalConflictGateError{Code: "APPROVAL_ACTION_VERSION_CONFLICT", Message: "审批节点已经变化，请刷新后重新处理"}
	}
	return nil
}

func releaseIntakeBindingInTx(tx *gorm.DB, id string) error {
	if !tx.Migrator().HasTable("approval_case_creation_tracking") {
		return nil
	}
	return tx.Where("approval_request_id = ? AND intake_id IS NOT NULL", id).Delete(&repositories.CaseCreationTracking{}).Error
}

// Shared by both decision APIs; a remembered assignment is insufficient when
// access to the underlying business subject has since been revoked.
func authorizeApprovalSubject(db *gorm.DB, userID string, approval *models.ApprovalRequest) error {
	metadata := parseMetadata(approval.Metadata)
	checkID := strings.TrimSpace(approval.ConflictCheckID)
	if checkID == "" {
		checkID = conflictCheckIDFromMetadata(metadata)
	}
	controlled := approval.ActiveSubmissionSnapshotID != nil || checkID != "" || approval.Type == "case_creation" || approval.Type == "conflict_approval" || approval.Type == "conflict" || approval.Type == "waiver"
	if !controlled {
		return nil
	}
	if checkID == "" {
		return &ApprovalConflictGateError{Code: "APPROVAL_CONTEXT_UNAVAILABLE", Message: "审批缺少可验证的业务上下文"}
	}
	var user models.User
	if err := db.First(&user, "id = ?", userID).Error; err != nil {
		return ErrApprovalForbidden
	}
	var record models.ConflictCheckRecord
	if err := db.Where("check_id = ?", checkID).First(&record).Error; err != nil {
		return &ApprovalConflictGateError{Code: "APPROVAL_CONTEXT_UNAVAILABLE", Message: "审批关联的检测记录不可用"}
	}
	ctx := context.Background()
	subject, err := ResolveConflictSubjectContext(ctx, db, &record)
	if err != nil {
		return err
	}
	auth := NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil)
	actor := AuthActor{UserID: user.ID, Role: user.Role}
	var allowed bool
	if subject.CaseID > 0 {
		allowed, err = auth.CanReadConflictContext(ctx, actor, subject.CaseID)
	} else if subject.IntakeID != "" {
		allowed, err = auth.CanReadConflictIntakeContext(ctx, actor, subject.IntakeID, record.UserID, record.ClientID)
	}
	if err != nil {
		return err
	}
	if !allowed {
		return ErrApprovalForbidden
	}
	return nil
}

func reconcileConflictTransition(tx *gorm.DB, checkID string) error {
	if !tx.Migrator().HasTable("inbox_items") {
		return nil
	}
	return ReconcileConflictInboxInTx(tx, checkID)
}

// ReconcileFormalConflictInbox is passed to the repository commit boundary;
// keeping the callback here avoids a repository-to-service import cycle.
func ReconcileFormalConflictInbox(tx *gorm.DB, checkID string) error {
	return reconcileConflictTransition(tx, checkID)
}

// A snapshot is immutable, but a later intake edit revokes its eligibility for
// a new approval decision. Serialize with draft writes and check the facts token.
func validateCurrentIntakeSubmission(tx *gorm.DB, approval *models.ApprovalRequest) error {
	if err := requireIntakeCaseAuthority(tx, approval, false); err != nil {
		return err
	}
	if approval.ActiveSubmissionSnapshotID == nil {
		return nil
	}
	snapshot, err := ResolveApprovalSubmission(tx, approval)
	if err != nil {
		return err
	}
	frozen := submissionMap(snapshot["metadata"])
	intakeID := submissionText(frozen["intake_id"])
	var intake struct {
		Status   string
		Metadata string
	}
	if err := tx.Table("case_intakes").Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", intakeID).Take(&intake).Error; err != nil {
		return err
	}
	live := parseMetadata(intake.Metadata)
	if intake.Status != "conflict_ready" || submissionText(live["lawyer_facts_version"]) == "" || submissionText(live["lawyer_facts_version"]) != submissionText(frozen["lawyer_facts_version"]) || submissionText(live["conflict_check_id"]) != submissionText(frozen["conflict_check_id"]) {
		return &ApprovalConflictGateError{Code: "INTAKE_FACTS_STALE", Message: "接案事实或检测已更新，请退回补充并重新提交当前材料"}
	}
	gate := &approvalConflictIntegrationService{db: tx, integrationRepo: repositories.NewIntegrationRepository(tx)}
	_, err = gate.validateConflictApprovalGate(tx.Statement.Context, approval)
	return err
}
