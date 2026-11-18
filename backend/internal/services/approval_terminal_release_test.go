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

// Terminal-state intake-bind lifecycle: the bind row guarding the UNIQUE
// intake_id index must be released when an approval ends its current
// lifecycle (rejected / needs-revision / cancelled) and re-armed when the
// SAME approval is revived via ResubmitApproval. Without this contract the
// intake stays locked behind INTAKE_APPROVAL_ALREADY_IN_PROGRESS forever
// (probe-reproduced 409 on 58a91ed, both reject and cancel paths).

func setupTerminalBindDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/terminal_bind.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ApprovalRequest{},
		&models.User{},
		&repositories.CaseCreationTracking{},
	))
	return db
}

func newTerminalApproval(t *testing.T, db *gorm.DB, status string) *models.ApprovalRequest {
	t.Helper()
	approval := &models.ApprovalRequest{
		ID:                fmt.Sprintf("APR-TS-%d", time.Now().UnixNano()),
		RequestNumber:     fmt.Sprintf("RN-TS-%d", time.Now().UnixNano()),
		Title:             "terminal bind lifecycle",
		Type:              "CASE_CREATION",
		Content:           "terminal-state bind lifecycle probe",
		ApplicantID:       "u-lawyer",
		ApplicantName:     "lawyer-a",
		DepartmentName:    "corporate",
		Status:            status,
		WorkflowType:      "CASE_APPROVAL",
		CurrentApproverID: "u-director",
		Metadata:          "{\"intake_id\":\"INT-TS-1\"}",
	}
	require.NoError(t, db.Create(approval).Error)
	return approval
}

func seedTerminalBind(t *testing.T, db *gorm.DB, approvalID string) {
	t.Helper()
	repo := repositories.NewIntegrationRepository(db)
	intake := "INT-TS-1"
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
		CreatedBy:           "u-lawyer",
	}))
}

func bindCountFor(db *gorm.DB, t *testing.T, approvalID string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&repositories.CaseCreationTracking{}).
		Where("approval_request_id = ?", approvalID).
		Count(&n).Error)
	return n
}

func TestTerminalBindingRelease_Contract(t *testing.T) {
	cases := []struct{ name, status string }{
		{"rejected releases bind", models.ApprovalStatusRejected},
		{"needs-revision releases bind", models.ApprovalStatusNeedsRevision},
		{"cancelled releases bind", models.ApprovalStatusCancelled},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTerminalBindDB(t)
			approval := newTerminalApproval(t, db, models.ApprovalStatusSubmitted)
			seedTerminalBind(t, db, approval.ID)
			releaser := newTerminalBindingReleaser(db)

			approval.Status = tc.status
			releaser.release(context.Background(), approval.ID)

			require.Equal(t, int64(0), bindCountFor(db, t, approval.ID),
				"terminal state must release the intake bind row")
		})
	}
}

func TestTerminalBindingRelease_NilSafe(t *testing.T) {
	var nilReleaser *terminalBindingReleaser
	require.NotPanics(t, func() {
		nilReleaser.release(context.Background(), "APR-anything")
	})
}

func TestResubmitRebind_RearmsUniqueSlot(t *testing.T) {
	db := setupTerminalBindDB(t)
	approval := newTerminalApproval(t, db, models.ApprovalStatusRejected)
	releaser := newTerminalBindingReleaser(db)

	approval.Status = models.ApprovalStatusResubmitted
	approval.UpdatedBy = "u-lawyer"
	releaser.rebindAfterResubmit(context.Background(), approval)

	require.Equal(t, int64(1), bindCountFor(db, t, approval.ID),
		"revived approval must own its intake bind row again")
}

func TestResubmitRebind_NoIntakeMetadata_Noop(t *testing.T) {
	db := setupTerminalBindDB(t)
	approval := newTerminalApproval(t, db, models.ApprovalStatusRejected)
	approval.Metadata = "{}"
	require.NoError(t, db.Save(approval).Error)
	releaser := newTerminalBindingReleaser(db)

	approval.Status = models.ApprovalStatusResubmitted
	approval.UpdatedBy = "u-lawyer"
	releaser.rebindAfterResubmit(context.Background(), approval)

	require.Equal(t, int64(0), bindCountFor(db, t, approval.ID))
}
