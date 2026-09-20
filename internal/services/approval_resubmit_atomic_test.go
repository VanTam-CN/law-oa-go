package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// Regression tests for audit-25f1d4b P1: reviving a terminal approval must
// be atomic with its intake rebind, and the case-creation transaction must
// refuse to create a case when the intake slot belongs to a competing
// approval. Together they close the two-case-per-intake window.

func seedAtomicResubmitWorld(t *testing.T) (*gorm.DB, *repositories.IntegrationRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/resubmit_atomic.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ApprovalRequest{},
		&models.User{},
		&models.ApprovalWorkflow{},
		&repositories.CaseCreationTracking{},
	))
	// Rejected -> resubmitted re-assigns the approver; the default path needs
	// an active admin who is not the applicant.
	require.NoError(t, db.Create(&models.User{
		Username: fmt.Sprintf("admin-%d", time.Now().UnixNano()),
		Name:     "审批管理员",
		Email:    fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()),
		Password: "x",
		Role:     "admin",
		Status:   "active",
	}).Error)
	return db, repositories.NewIntegrationRepository(db)
}

func newAtomicResubmitApproval(t *testing.T, db *gorm.DB, intakeID string) *models.ApprovalRequest {
	t.Helper()
	approval := &models.ApprovalRequest{
		ID:                fmt.Sprintf("APR-RA-%d", time.Now().UnixNano()),
		RequestNumber:     fmt.Sprintf("RN-RA-%d", time.Now().UnixNano()),
		Title:             "atomic resubmit guard",
		Type:              "CASE_CREATION",
		Content:           "atomic revival + intake slot guard",
		ApplicantID:       "u-lawyer",
		ApplicantName:     "lawyer-a",
		DepartmentName:    "corporate",
		Status:            models.ApprovalStatusRejected,
		WorkflowType:      "CASE_APPROVAL",
		CurrentApproverID: "u-director",
		Metadata:          fmt.Sprintf(`{"intake_id":"%s"}`, intakeID),
	}
	require.NoError(t, db.Create(approval).Error)
	return approval
}

func seedAtomicBind(t *testing.T, db *gorm.DB, approvalID, intakeID string) {
	t.Helper()
	repo := repositories.NewIntegrationRepository(db)
	intake := intakeID
	require.NoError(t, repo.CreateCaseCreationTracking(context.Background(), &repositories.CaseCreationTracking{
		ID:                  fmt.Sprintf("INTAKE_BIND_%s", approvalID),
		ApprovalRequestID:   approvalID,
		IntakeID:            &intake,
		CreationStatus:      "pending",
		ErrorDetails:        "{}",
		DataMapping:         "{}",
		MappedFields:        "{}",
		UnmappedFields:      "{}",
		AppliedConditions:   "{}",
		ImposedRequirements: "{}",
		WorkflowActions:     "[]",
		CreatedBy:           approvalID,
	}))
}

func atomicBindCount(t *testing.T, db *gorm.DB, where string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&repositories.CaseCreationTracking{}).Where(where, args...).Count(&n).Error)
	return n
}

// When a competing submission owns the intake slot, resubmitting the
// terminal approval must be refused with the 409 contract AND must not
// flip the approval to resubmitted: status persistence and the rebind are
// one transaction.
func TestResubmitAtomic_CompetingSlotRollsBackStatus(t *testing.T) {
	db, _ := seedAtomicResubmitWorld(t)
	intake := "INT-RA-1"
	approval := newAtomicResubmitApproval(t, db, intake)
	// The competitor (B) claimed the UNIQUE slot while A sat rejected.
	seedAtomicBind(t, db, "APR-competitor-B", intake)

	svc := NewApprovalService(db)
	_, err := svc.ResubmitApproval("u-lawyer", approval.ID, "revive attempt")
	require.Error(t, err)
	var gateErr *ApprovalConflictGateError
	require.ErrorAs(t, err, &gateErr)
	require.Equal(t, "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", gateErr.Code)

	var reloaded models.ApprovalRequest
	require.NoError(t, db.First(&reloaded, "id = ?", approval.ID).Error)
	require.Equal(t, models.ApprovalStatusRejected, reloaded.Status,
		"status must roll back with the failed rebind")
	require.Equal(t, int64(1), atomicBindCount(t, db, "approval_request_id = ?", "APR-competitor-B"),
		"competitor keeps the slot")
	require.Equal(t, int64(0), atomicBindCount(t, db, "approval_request_id = ?", approval.ID))
}

// The same revival on an intake whose slot is free (or already self-owned)
// must succeed: rebind + resubmitted status commit together.
func TestResubmitAtomic_FreeSlotCommitsBoth(t *testing.T) {
	db, _ := seedAtomicResubmitWorld(t)
	intake := "INT-RA-2"
	approval := newAtomicResubmitApproval(t, db, intake)

	svc := NewApprovalService(db)
	updated, err := svc.ResubmitApproval("u-lawyer", approval.ID, "revive")
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)
	require.Equal(t, int64(1), atomicBindCount(t, db, "approval_request_id = ?", approval.ID),
		"the revived approval owns its slot again")
}

// Case-creation guard: a competing bind row on the same intake blocks the
// whole transaction — no case, no parties, no tracking completion.
func TestCaseTxGuard_CompetingIntakeSlotBlocked(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	intake := "INT-TX-GUARD-1"
	seedAtomicBind(t, f.db, "APR-TX-competitor", intake)
	// Point the approval at the contested intake.
	f.approval.Metadata = fmt.Sprintf(`{"conflict_check_id":"CHK-TX-1","intake_id":"%s"}`, intake)
	require.NoError(t, f.db.Save(f.approval).Error)

	_, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.Error(t, err)
	var workflowErr *SubjectWorkflowError
	require.ErrorAs(t, err, &workflowErr)
	require.Equal(t, "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", workflowErr.Code)

	var caseCount int64
	require.NoError(t, f.db.Model(&models.Case{}).Count(&caseCount).Error)
	require.EqualValues(t, 0, caseCount, "the contested case must roll back")
}

// The guard must not regress legitimate creation: when the bind row belongs
// to the same approval the case still lands.
func TestCaseTxGuard_SelfOwnedSlotPasses(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	intake := "INT-TX-GUARD-2"
	seedAtomicBind(t, f.db, "APR-TX-1", intake)
	f.approval.Metadata = fmt.Sprintf(`{"conflict_check_id":"CHK-TX-1","intake_id":"%s"}`, intake)
	require.NoError(t, f.db.Save(f.approval).Error)

	result, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "completed", result.Status)
}

// Legacy rows created before the bind invariant have no intake metadata and
// no bind row: creation must keep passing (000081 does not backfill).
func TestCaseTxGuard_LegacyNoIntakeMetadataPasses(t *testing.T) {
	f := seedApprovalCaseWorld(t)
	// f.approval.Metadata keeps only conflict_check_id (no intake_id).
	result, err := f.svc.AutoCreateCaseFromApproval(context.Background(), "APR-TX-1", caseTxData())
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "completed", result.Status)
}
