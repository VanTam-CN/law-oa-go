package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// Author tests for the approval binding and phase projection only. Protected
// identity and conflict evidence still pass through the existing full gate.
func seedAuthorContinuationApproval(t *testing.T, db *gorm.DB, intakeID, status, metadata string) models.ApprovalRequest {
	t.Helper()
	active := "author-active-intake-round"
	approval := models.ApprovalRequest{
		ID: "author-intake-approval", RequestNumber: "author-intake-number", Title: "虚构接案审批",
		Type: "CASE_CREATION", Status: status, Metadata: metadata, ActiveSubmissionSnapshotID: &active,
	}
	require.NoError(t, db.Create(&approval).Error)
	require.NoError(t, db.Exec("INSERT INTO approval_snapshots (id, approval_request_id, snapshot_type, snapshot_data, source_version) VALUES (?, ?, ?, ?, ?)",
		active, approval.ID, models.ApprovalSnapshotTypeIntakeSubmission, `{"intake_id":"`+intakeID+`"}`, 1).Error)
	return approval
}

func TestAuthorContinuationApprovalJSONMatchesExactIntake(t *testing.T) {
	db := seedContinuationWorld(t)
	intakeID := "INT-AUTHOR-JSON"
	seedContinuationIntake(t, db, intakeID, continuationSeedOptions{withReview: true})
	approval := seedAuthorContinuationApproval(t, db, intakeID, models.ApprovalStatusNeedsRevision,
		"{\n  \"intake_id\": \""+intakeID+"\"\n}")
	// A newer row with the intake token under another key or nested object
	// must not hide the original approval or claim the intake as its own.
	foreign := models.ApprovalRequest{
		ID: "author-foreign-approval", RequestNumber: "author-foreign-number", Title: "虚构其他申请",
		Type: "CASE_CREATION", Status: models.ApprovalStatusSubmitted,
		Metadata: `{"intake_id":"` + intakeID + `-OTHER","other":{"intake_id":"` + intakeID + `"}}`,
	}
	require.NoError(t, db.Create(&foreign).Error)
	decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation(intakeID)
	require.NoError(t, err)
	require.NotNil(t, decision.Approval)
	require.Equal(t, approval.ID, decision.Approval.ApprovalID)
	require.Equal(t, *approval.ActiveSubmissionSnapshotID, decision.Approval.SnapshotID)
	require.Equal(t, 1, decision.Approval.Round)
	require.True(t, decision.CanSubmit)
}

func TestAuthorContinuationApprovalPhaseReasonsPreserveReuse(t *testing.T) {
	for _, tc := range []struct {
		status string
		reason string
	}{
		{models.ApprovalStatusNeedsRevision, ""},
		{models.ApprovalStatusSubmitted, "APPROVAL_IN_PROGRESS"},
		{models.ApprovalStatusUnderReview, "APPROVAL_IN_PROGRESS"},
		{models.ApprovalStatusResubmitted, "APPROVAL_IN_PROGRESS"},
		{models.ApprovalStatusApproved, "APPROVAL_CLOSED"},
		{models.ApprovalStatusRejected, "APPROVAL_CLOSED"},
		{models.ApprovalStatusCancelled, "APPROVAL_CLOSED"},
		{models.ApprovalStatusExpired, "APPROVAL_CLOSED"},
		{models.ApprovalStatusDraft, "APPROVAL_STATE_UNKNOWN"},
		{"unknown", "APPROVAL_STATE_UNKNOWN"},
		{"unrecognized_phase", "APPROVAL_STATE_UNKNOWN"},
	} {
		t.Run(tc.status, func(t *testing.T) {
			db := seedContinuationWorld(t)
			intakeID := "INT-AUTHOR-PHASE"
			seedContinuationIntake(t, db, intakeID, continuationSeedOptions{withReview: true})
			seedAuthorContinuationApproval(t, db, intakeID, tc.status, `{"intake_id":"`+intakeID+`"}`)
			decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation(intakeID)
			require.NoError(t, err)
			require.True(t, decision.CanReuseIdentity)
			require.True(t, decision.CanReuseCheck)
			require.False(t, decision.RecheckRequired)
			require.Equal(t, tc.reason == "", decision.CanSubmit)
			if tc.reason != "" {
				require.Contains(t, decision.Reasons, tc.reason)
			} else {
				require.Empty(t, decision.Reasons)
			}
		})
	}
}

func TestAuthorContinuationApprovalSnapshotReadErrorFailsClosed(t *testing.T) {
	db := seedContinuationWorld(t)
	intakeID := "INT-AUTHOR-READ-ERROR"
	seedContinuationIntake(t, db, intakeID, continuationSeedOptions{withReview: true})
	seedAuthorContinuationApproval(t, db, intakeID, models.ApprovalStatusNeedsRevision, `{"intake_id":"`+intakeID+`"}`)
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("author_snapshot_read_failure", func(tx *gorm.DB) {
		if tx.Statement.Table == "approval_snapshots" {
			tx.AddError(errors.New("author snapshot read failure"))
		}
	}))
	decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation(intakeID)
	require.ErrorContains(t, err, "author snapshot read failure")
	require.Nil(t, decision)
}

func TestAuthorContinuationApprovalInvalidJSONFailsClosed(t *testing.T) {
	db := seedContinuationWorld(t)
	intakeID := "INT-AUTHOR-INVALID-JSON"
	seedContinuationIntake(t, db, intakeID, continuationSeedOptions{withReview: true})
	seedAuthorContinuationApproval(t, db, intakeID, models.ApprovalStatusNeedsRevision, "{broken-json")
	decision, err := NewIntakeContinuationService(db).EvaluateIntakeContinuation(intakeID)
	require.Error(t, err)
	require.Nil(t, decision)
}
