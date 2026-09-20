package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// submitFailingApprovalService mimics a real CreateApproval success followed
// by a SubmitApproval failure (for example: no eligible approver found).
type submitFailingApprovalService struct {
	ApprovalServiceInterface
	created int
}

func (s *submitFailingApprovalService) CreateApproval(userID, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error) {
	s.created++
	return &models.ApprovalRequest{ID: fmt.Sprintf("APR-BIND-FAIL-%d", s.created)}, nil
}

func (s *submitFailingApprovalService) SubmitApproval(userID, approvalID string) (*models.ApprovalRequest, error) {
	return nil, fmt.Errorf("提交审批失败：分配审批人失败，无可用审批人")
}

// insertRecordingRepo records every successful CreateCaseCreationTracking so
// the test can prove the retry actually inserted its own bind row.
type insertRecordingRepo struct {
	*repositories.IntegrationRepository
	insertedBindIDs []string
}

func (r *insertRecordingRepo) CreateCaseCreationTracking(ctx context.Context, tracking *repositories.CaseCreationTracking) error {
	if err := r.IntegrationRepository.CreateCaseCreationTracking(ctx, tracking); err != nil {
		return err
	}
	r.insertedBindIDs = append(r.insertedBindIDs, tracking.ID)
	return nil
}

// TestIntakeBindCompensationOnSubmitFailure pins the compensating-delete
// contract on the failed-submit path: the INTAKE_BIND row written before
// SubmitApproval must be removed so the intake retry slot is released. With
// the UNIQUE index on intake_id this is what keeps a transient submit
// failure from permanently 409-locking the intake.
func TestIntakeBindCompensationOnSubmitFailure(t *testing.T) {
	db := setupApprovalCaseTxDB(t)
	repo := &insertRecordingRepo{IntegrationRepository: repositories.NewIntegrationRepository(db)}
	svc := &approvalConflictIntegrationService{
		approvalService: &submitFailingApprovalService{},
		integrationRepo: repo,
		db:              db,
	}
	intakeID := "ITK-COMP-1"

	res, err := svc.CreateIntegratedApproval(context.Background(), "u1", "申请人", &IntegrationRequest{
		Type:           "CASE_CREATION",
		Title:          "补偿路径验证",
		Content:        "approval body",
		ApplicantName:  "申请人",
		DepartmentName: "业务部",
		WorkflowType:   "standard",
		Metadata:       map[string]interface{}{"intake_id": intakeID},
	})
	require.Nil(t, res)
	require.Error(t, err)
	require.Contains(t, err.Error(), "提交集成审批申请失败")

	// The bind row for the failed draft is gone.
	var bindCount int64
	require.NoError(t, db.Model(&repositories.CaseCreationTracking{}).
		Where("intake_id = ?", intakeID).Count(&bindCount).Error)
	require.Zero(t, bindCount, "compensating delete must remove the INTAKE_BIND row after a failed submit")
	require.Error(t, db.First(&repositories.CaseCreationTracking{}, "id = ?", "INTAKE_BIND_APR-BIND-FAIL-1").Error,
		gorm.ErrRecordNotFound)

	// The slot is released: a legitimate retry's bind INSERT succeeds past
	// the UNIQUE index (if compensation had not run, this insert would have
	// failed with the INTAKE_APPROVAL_ALREADY_IN_PROGRESS gate instead).
	res2, err2 := svc.CreateIntegratedApproval(context.Background(), "u1", "申请人", &IntegrationRequest{
		Type:           "CASE_CREATION",
		Title:          "补偿后重试绑定",
		Content:        "approval body",
		ApplicantName:  "申请人",
		DepartmentName: "业务部",
		WorkflowType:   "standard",
		Metadata:       map[string]interface{}{"intake_id": intakeID},
	})
	require.Nil(t, res2)
	require.Contains(t, err2.Error(), "提交集成审批申请失败")
	require.NotContains(t, err2.Error(), "INTAKE_APPROVAL_ALREADY_IN_PROGRESS",
		"the retry must get past the bind insert, proving compensation released the slot")
	require.Equal(t, []string{"INTAKE_BIND_APR-BIND-FAIL-1", "INTAKE_BIND_APR-BIND-FAIL-2"}, repo.insertedBindIDs,
		"both attempts must have inserted their own bind row")

	// The retry's own bind row is also compensated away after its submit
	// fails, so no orphan bind remains for the intake.
	var bindCountAfterRetry int64
	require.NoError(t, db.Model(&repositories.CaseCreationTracking{}).
		Where("intake_id = ?", intakeID).Count(&bindCountAfterRetry).Error)
	require.Zero(t, bindCountAfterRetry)
}
