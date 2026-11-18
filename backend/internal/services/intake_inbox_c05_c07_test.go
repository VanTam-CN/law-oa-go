package services

// N4 (C05-C07) acceptance coverage: role-direct todo objects, permission
// projection on reassignment, workflow completion/deletion guards, and
// snooze semantics that never change the business state.

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func newHandoffWorldDB(t *testing.T) (*gorm.DB, *ApprovalService) {
	t.Helper()
	db := seedHandoffWorld(t)
	return db, NewApprovalService(db)
}

// C05 (approval role): the submission handoff carries the approval UUID in
// source_ref (no uint/UUID guessing) and is workflow-owned.
func TestC05_ApprovalHandoffSourceRefIsApprovalUUID(t *testing.T) {
	db, svc := newHandoffWorldDB(t)
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-C05-%d", now), RequestNumber: fmt.Sprintf("RN-C05-%d", now),
		Title: "C05 审批直达", Type: "CASE_CREATION", Content: "c05",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲", Status: models.ApprovalStatusDraft,
		WorkflowType: "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.Create(approval).Error)

	_, err := svc.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)

	items := openTasks(t, db, "u-h-admin")
	require.Len(t, items, 1)
	require.Equal(t, approval.ID, items[0].SourceRef)
	require.Equal(t, "workflow", items[0].CompletionMode)

	var approvalCount int64
	require.NoError(t, db.Model(&models.ApprovalRequest{}).
		Where("status = ?", models.ApprovalStatusSubmitted).Count(&approvalCount).Error)
	require.Equal(t, int64(1), approvalCount)
}

// C05 (reviewer role): an appointment opens exactly one reviewer handoff;
// a repeated appointment stays at one (idempotent key).
func TestC05_ReviewerAppointmentOpensOneHandoff(t *testing.T) {
	db := reviewerTestDB(t)
	seedReviewerCheck(t, db)
	actor := AuthActor{UserID: 20, Role: "conflict_officer"}

	assignment, err := AssignConflictReviewer(context.Background(), db, actor, "check-reviewer",
		ConflictReviewerAssignmentInput{ReviewerID: 20, RecusalDeclared: true, IndependenceReason: "独立"})
	require.NoError(t, err)
	require.NotEmpty(t, assignment.ID)

	items := openTasks(t, db, "officer")
	require.Len(t, items, 1)
	require.Equal(t, "check-reviewer", items[0].SourceRef)
	require.Equal(t, "workflow", items[0].CompletionMode)

	_, err = AssignConflictReviewer(context.Background(), db, actor, "check-reviewer",
		ConflictReviewerAssignmentInput{ReviewerID: 20, RecusalDeclared: true})
	require.NoError(t, err)
	require.Len(t, openTasks(t, db, "officer"), 1)
}

// C06 (reassignment): moving the reviewer closes the old reviewer's open
// handoff and opens the new reviewer's handoff in the same transaction.
func TestC06_ReviewerReassignmentClosesOldHandoff(t *testing.T) {
	db := reviewerTestDB(t)
	seedReviewerCheck(t, db)
	require.NoError(t, db.Create(&models.User{
		ID: 21, Username: "compliance", Name: "合规复核人", Email: "compliance@example.test",
		Password: "x", Role: "compliance", Status: "active",
	}).Error)
	actor := AuthActor{UserID: 20, Role: "conflict_officer"}

	_, err := AssignConflictReviewer(context.Background(), db, actor, "check-reviewer",
		ConflictReviewerAssignmentInput{ReviewerID: 20, RecusalDeclared: true, IndependenceReason: "独立"})
	require.NoError(t, err)
	require.Len(t, openTasks(t, db, "officer"), 1)

	_, err = AssignConflictReviewer(context.Background(), db, actor, "check-reviewer",
		ConflictReviewerAssignmentInput{ReviewerID: 21, RecusalDeclared: true, IndependenceReason: "转交合规"})
	require.NoError(t, err)

	oldItems := openTasks(t, db, "officer")
	require.Len(t, oldItems, 0)
	var oldClosed int64
	require.NoError(t, db.Table("inbox_items").
		Where("workflow_task_key = ? AND is_completed = ?", "conflict_review:check-reviewer:20", true).
		Count(&oldClosed).Error)
	require.Equal(t, int64(1), oldClosed)
	require.Len(t, openTasks(t, db, "compliance"), 1)
}

// C06 (approval reassignment): the old approver's todo closes and the new
// approver's todo opens with the same cycle, in one transaction.
func TestC06_ApprovalReassignmentMovesHandoff(t *testing.T) {
	db, svc := newHandoffWorldDB(t)
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-C06-%d", now), RequestNumber: fmt.Sprintf("RN-C06-%d", now),
		Title: "C06 转交", Type: "CASE_CREATION", Content: "c06",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲", Status: models.ApprovalStatusDraft,
		WorkflowType: "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.Create(approval).Error)
	_, err := svc.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)

	// The submission stored the numeric approver id (user 104 = u-h-admin);
	// decisions authenticate on that string identity.
	_, err = svc.ProcessApprovalDecision("104", approval.ID, &models.ApprovalDecisionRequest{
		Decision: models.ApprovalDecisionReassign, NextApproverID: "103",
	})
	require.NoError(t, err)

	// The c0 admin handoff is closed; the new approver (user 103) holds the
	// open c0 handoff.
	var closedOld int64
	require.NoError(t, db.Table("inbox_items").
		Where("workflow_task_key LIKE ? AND is_completed = ?",
			fmt.Sprintf("approval:approval_decision:%s:c0:%%", approval.ID), true).
		Count(&closedOld).Error)
	require.Equal(t, int64(1), closedOld)

	newKey := fmt.Sprintf("approval:approval_decision:%s:c0:103", approval.ID)
	open := taskKey(t, db, newKey)
	require.NotNil(t, open)
	require.False(t, open.IsCompleted)
}

// C07 (guards): workflow todos cannot be completed manually, cannot be
// deleted, and the API-visible projection carries the workflow completion
// mode so clients cannot fake a manual one.
func TestC07_WorkflowGuardsCompletionDeletionProjection(t *testing.T) {
	db, svc := newHandoffWorldDB(t)
	inbox := NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db))
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-C07-%d", now), RequestNumber: fmt.Sprintf("RN-C07-%d", now),
		Title: "C07 守卫", Type: "CASE_CREATION", Content: "c07",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲", Status: models.ApprovalStatusDraft,
		WorkflowType: "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.Create(approval).Error)
	_, err := svc.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)

	items := openTasks(t, db, "u-h-admin")
	require.Len(t, items, 1)
	task := items[0]

	// Manual completion of a workflow item is refused.
	require.ErrorIs(t, inbox.MarkAsCompleted(context.Background(), task.ID, task.UserID), ErrWorkflowItemManualComplete)

	// Deletion is unavailable for inbox items (global guard).
	err = inbox.DeleteInboxItem(context.Background(), task.UserID, task.ID)
	require.Error(t, err)
	require.NotErrorIs(t, err, gorm.ErrRecordNotFound)

	// The projection marks the item workflow-owned; completion mode cannot
	// be forged by the client.
	var n int64
	require.NoError(t, db.Table("inbox_items").
		Where("id = ? AND completion_mode = 'workflow'", task.ID).Count(&n).Error)
	require.Equal(t, int64(1), n)
	var raw map[string]interface{}
	proj, err := inbox.GetInboxItemByID(context.Background(), task.ID, task.UserID)
	require.NoError(t, err)
	data, err := json.Marshal(proj)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &raw))
	require.Equal(t, "workflow", raw["completion_mode"])
}

// C07 (snooze): snoozing a workflow todo only delays visibility; the
// business state stays untouched and the item remains open and
// workflow-owned underneath.
func TestC07_SnoozeDoesNotChangeBusinessState(t *testing.T) {
	db, svc := newHandoffWorldDB(t)
	inbox := NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db))
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-C07S-%d", now), RequestNumber: fmt.Sprintf("RN-C07S-%d", now),
		Title: "C07 延后", Type: "CASE_CREATION", Content: "c07",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲", Status: models.ApprovalStatusDraft,
		WorkflowType: "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.Create(approval).Error)
	_, err := svc.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)

	items := openTasks(t, db, "u-h-admin")
	require.Len(t, items, 1)
	task := items[0]
	until := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04:05")
	require.NoError(t, inbox.SnoozeInboxItem(context.Background(), task.ID, task.UserID, &SnoozeInboxItemRequest{Until: until}))

	var approvalCount int64
	require.NoError(t, db.Model(&models.ApprovalRequest{}).
		Where("id = ? AND status = ?", approval.ID, models.ApprovalStatusSubmitted).Count(&approvalCount).Error)
	require.Equal(t, int64(1), approvalCount)

	var refreshed models.InboxItem
	require.NoError(t, db.Where("id = ?", task.ID).Take(&refreshed).Error)
	require.False(t, refreshed.IsCompleted)
	require.Equal(t, "workflow", refreshed.CompletionMode)
}
