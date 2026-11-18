package services

// C06/C07 retest against the plan originals (§11 C06/C07, §5.1/§5.4):
//
// C06 forgery -- a self-created inbox item may name source_type=approval,
// but the request DTO carries no source_ref/completion_mode, the DB
// default pins completion_mode=manual, and the action projection only
// renders business actions for workflow rows. Prove with an EXISTING
// approval object that the self-created row yields no object action (no
// navigation/read path) while the legit handoff in the same world still
// projects decide_approval.
//
// C07/Update -- UpdateInboxItemRequest only exposes title/content/
// priority/due_date/due_date_type (plan-converged reminder fields). The
// test pins that a workflow handoff keeps its completion_mode,
// workflow_task_key, and source_ref through an owner edit instead of
// inventing extra product rules beyond the plan.

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func c06UserIDByName(t *testing.T, db *handoffWorldDBHandle, username string) uint {
	t.Helper()
	var id int
	require.NoError(t, db.gorm.Raw("SELECT id FROM users WHERE username = ?", username).Scan(&id).Error)
	require.NotZero(t, id, "missing test user %s", username)
	return uint(id)
}

func c06ProjectOne(t *testing.T, db *handoffWorldDBHandle, itemID, userID uint) map[string]interface{} {
	t.Helper()
	svc := NewInboxService(repositories.NewInboxRepository(db.gorm), repositories.NewUserRepository(db.gorm))
	resp, err := svc.GetInboxItemByID(context.Background(), itemID, userID)
	require.NoError(t, err)
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	return raw
}

// C06: self-creating an approval-source todo against an EXISTING approval
// object must not mint an object action or a workflow-mode row.
func TestC06_SelfCreateApprovalSourceNeverGrantsObjectAction(t *testing.T) {
	db := newProjectionWorld(t)
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-C06S-%d", now), RequestNumber: fmt.Sprintf("RN-C06S-%d", now),
		Title: "C06 自建对照", Type: "CASE_CREATION", Content: "c06",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲", Status: models.ApprovalStatusDraft,
		WorkflowType: "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.gorm.Create(approval).Error)
	_, err := db.approval.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)

	// Legit handoff projection for contrast: the machinery works in this
	// world and the approver sees the whitelisted decide_approval action.
	realItems := openTasks(t, db.gorm, "u-h-admin")
	require.Len(t, realItems, 1)
	realRaw := c06ProjectOne(t, db, realItems[0].ID, realItems[0].UserID)
	realAction, ok := realRaw["workflow_action"].(map[string]interface{})
	require.True(t, ok, "legit handoff must project an action")
	require.Equal(t, "decide_approval", realAction["action_code"])
	require.Equal(t, "workflow", realAction["completion_mode"])

	// The applicant (who can already read the object through the approval
	// flow) forges an inbox row naming that same existing object.
	inbox := NewInboxService(repositories.NewInboxRepository(db.gorm), repositories.NewUserRepository(db.gorm))
	created, err := inbox.CreateInboxItem(context.Background(), c06UserIDByName(t, db, "u-h-lawyer"), &CreateInboxItemRequest{
		SourceType: "approval",
		// The numeric DTO cannot carry the approval UUID at all; point the
		// forged row at an ID that provably exists (the legit handoff's
		// inbox row) to prove existence alone unlocks nothing.
		SourceID: realItems[0].ID,
		Title:    "C06-自建待办-试图伪造对象动作",
		Priority: "high",
	})
	require.NoError(t, err)

	// Response: manual mode, no source_ref, no object action.
	data, err := json.Marshal(created)
	require.NoError(t, err)
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	require.Equal(t, "manual", raw["completion_mode"])
	// source_ref is omitempty: an unset ref serializes as an absent key
	// (JSON null here), which is exactly the no-binding contract.
	require.Nil(t, raw["source_ref"])
	require.Nil(t, raw["workflow_action"])

	// Re-projection after persistence stays actionless.
	reRaw := c06ProjectOne(t, db, created.ID, created.UserID)
	require.Nil(t, reRaw["workflow_action"])
	require.Equal(t, "manual", reRaw["completion_mode"])

	// DB row changes match the contract: manual mode, no ref, no key, and
	// the workflow-mode population is still exactly the legit handoff.
	var row models.InboxItem
	require.NoError(t, db.gorm.First(&row, created.ID).Error)
	require.Equal(t, "manual", row.CompletionMode)
	require.Empty(t, row.SourceRef)
	require.Empty(t, row.WorkflowTaskKey)
	var workflowRows int64
	require.NoError(t, db.gorm.Table("inbox_items").
		Where("completion_mode = ?", "workflow").Count(&workflowRows).Error)
	require.Equal(t, int64(1), workflowRows)

	// The forged row is still the owner's personal reminder: manual
	// completion succeeds (C07 keeps ordinary todos completable).
	require.NoError(t, inbox.MarkAsCompleted(context.Background(), created.ID, created.UserID))
}

// C07: updating a workflow handoff is confined to the reminder fields in
// UpdateInboxItemRequest; the workflow ownership fields never move.
func TestC07_UpdateRequestCannotTouchWorkflowFields(t *testing.T) {
	db := newProjectionWorld(t)
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-C06U-%d", now), RequestNumber: fmt.Sprintf("RN-C06U-%d", now),
		Title: "C07 更新对照", Type: "CASE_CREATION", Content: "c07",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲", Status: models.ApprovalStatusDraft,
		WorkflowType: "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.gorm.Create(approval).Error)
	_, err := db.approval.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)

	items := openTasks(t, db.gorm, "u-h-admin")
	require.Len(t, items, 1)
	item := items[0]

	inbox := NewInboxService(repositories.NewInboxRepository(db.gorm), repositories.NewUserRepository(db.gorm))
	updated, err := inbox.UpdateInboxItem(context.Background(), item.ID, item.UserID, &UpdateInboxItemRequest{
		Title:       "被所有者改名的 workflow 待办",
		Content:     "更新内容",
		Priority:    "low",
		DueDate:     func() *string { s := "2030-01-01"; return &s }(),
		DueDateType: "hearing",
	})
	require.NoError(t, err)

	// Response keeps the workflow identity fields.
	data, err := json.Marshal(updated)
	require.NoError(t, err)
	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))
	require.Equal(t, "workflow", raw["completion_mode"])
	require.Equal(t, item.WorkflowTaskKey, raw["workflow_task_key"])
	require.Equal(t, approval.ID, raw["source_ref"])
	action, ok := raw["workflow_action"].(map[string]interface{})
	require.True(t, ok, "workflow action must survive the edit")
	require.Equal(t, "decide_approval", action["action_code"])
	require.Equal(t, "workflow", action["completion_mode"])

	// DB row: only reminder fields changed; ownership fields intact.
	var row models.InboxItem
	require.NoError(t, db.gorm.First(&row, item.ID).Error)
	require.Equal(t, "workflow", row.CompletionMode)
	require.Equal(t, item.WorkflowTaskKey, row.WorkflowTaskKey)
	require.Equal(t, approval.ID, row.SourceRef)
	require.Equal(t, "被所有者改名的 workflow 待办", row.Title)
	require.Equal(t, "low", row.Priority)
	require.False(t, row.IsCompleted)
}
