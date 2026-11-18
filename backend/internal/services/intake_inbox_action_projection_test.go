package services

// N5 (C05 extension): server-generated WorkflowAction projection. The
// client maps only whitelisted action codes; unresolvable objects stay
// visible but blocked instead of guessing a target.

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

// handoffWorldDBHandle bundles the sqlite world reused from the N4 tests.
type handoffWorldDBHandle struct {
	gorm     *gorm.DB
	approval *ApprovalService
}

func newProjectionWorld(t *testing.T) *handoffWorldDBHandle {
	t.Helper()
	db := seedHandoffWorld(t)
	return &handoffWorldDBHandle{gorm: db, approval: NewApprovalService(db)}
}

func mustProjectOne(t *testing.T, db *handoffWorldDBHandle, itemID, userID uint) map[string]interface{} {
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

func requireAction(t *testing.T, raw map[string]interface{}) map[string]interface{} {
	t.Helper()
	action, ok := raw["workflow_action"].(map[string]interface{})
	require.True(t, ok, "workflow_action missing: %v", raw)
	return action
}

// C05: an approver decision todo projects a whitelisted decide_approval
// action with the approval UUID and a cycle-anchored version.
func TestN5ActionApproverDecisionProjection(t *testing.T) {
	db := newProjectionWorld(t)
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-N5A-%d", now), RequestNumber: fmt.Sprintf("RN-N5A-%d", now),
		Title: "N5 审批直达", Type: "CASE_CREATION", Content: "n5",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲", Status: models.ApprovalStatusDraft,
		WorkflowType: "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.gorm.Create(approval).Error)
	_, err := db.approval.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)

	items := openTasks(t, db.gorm, "u-h-admin")
	require.Len(t, items, 1)
	raw := mustProjectOne(t, db, items[0].ID, items[0].UserID)
	action := requireAction(t, raw)
	require.Equal(t, "decide_approval", action["action_code"])
	require.Equal(t, "approval", action["resource_type"])
	require.Equal(t, approval.ID, action["resource_id"])
	require.Equal(t, "approval:"+approval.ID+":c0", action["action_version"])
	require.Equal(t, "workflow", action["completion_mode"])
}

// C05: the applicant revision todo projects supplement with the intake
// ID from the approval metadata; without the binding the action blocks
// instead of generating a guessed /case/create URL.
func TestN5ActionSupplementProjectionUsesIntakeBinding(t *testing.T) {
	db := newProjectionWorld(t)
	now := time.Now().UnixNano()
	intakeID := fmt.Sprintf("INT-N5-%d", now)
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-N5B-%d", now), RequestNumber: fmt.Sprintf("RN-N5B-%d", now),
		Title: "N5 补充直达", Type: "CASE_CREATION", Content: "n5",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲",
		Status:            models.ApprovalStatusNeedsRevision,
		CurrentApproverID: "u-h-admin",
		WorkflowType:      "CASE_APPROVAL",
		Metadata:          fmt.Sprintf("{\"intake_id\":\"%s\"}", intakeID),
	}
	require.NoError(t, db.gorm.Create(approval).Error)
	rec := &models.ApprovalRecord{
		ID: fmt.Sprintf("REC-N5-%d", now), ApprovalRequestID: approval.ID,
		Stage: "stage-1", StageOrder: 1, ApproverID: "u-h-admin",
		Decision: models.ApprovalDecisionRequestChanges, DecisionReason: "补充说明",
		ImposedRequirements: "{\"schema_version\":1,\"items\":[]}",
		FollowUpActions:     "[]", ApprovalDate: time.Now(),
		SupportingDocuments: "[]", EvidenceReferences: "[]", Status: "active",
	}
	require.NoError(t, db.gorm.Create(rec).Error)
	require.NoError(t, db.approval.openApplicantRevisionTask(db.gorm, approval))

	items := openTasks(t, db.gorm, "u-h-lawyer")
	require.Len(t, items, 1)
	raw := mustProjectOne(t, db, items[0].ID, items[0].UserID)
	action := requireAction(t, raw)
	require.Equal(t, "supplement", action["action_code"])
	require.Equal(t, "intake", action["resource_type"])
	require.Equal(t, intakeID, action["resource_id"])
	require.Equal(t, approval.ID, action["related_approval_id"])
}

// C05: missing intake binding blocks the supplement action.
func TestN5ActionSupplementWithoutIntakeBlocksAction(t *testing.T) {
	db := newProjectionWorld(t)
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-N5C-%d", now), RequestNumber: fmt.Sprintf("RN-N5C-%d", now),
		Title: "N5 无绑定", Type: "CASE_CREATION", Content: "n5",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲",
		Status:            models.ApprovalStatusNeedsRevision,
		CurrentApproverID: "u-h-admin",
		WorkflowType:      "CASE_APPROVAL", Metadata: "{}",
	}
	require.NoError(t, db.gorm.Create(approval).Error)
	rec := &models.ApprovalRecord{
		ID: fmt.Sprintf("REC-N5C-%d", now), ApprovalRequestID: approval.ID,
		Stage: "stage-1", StageOrder: 1, ApproverID: "u-h-admin",
		Decision: models.ApprovalDecisionRequestChanges, DecisionReason: "补充说明",
		ImposedRequirements: "{\"schema_version\":1,\"items\":[]}",
		FollowUpActions:     "[]", ApprovalDate: time.Now(),
		SupportingDocuments: "[]", EvidenceReferences: "[]", Status: "active",
	}
	require.NoError(t, db.gorm.Create(rec).Error)
	require.NoError(t, db.approval.openApplicantRevisionTask(db.gorm, approval))

	items := openTasks(t, db.gorm, "u-h-lawyer")
	require.Len(t, items, 1)
	raw := mustProjectOne(t, db, items[0].ID, items[0].UserID)
	action := requireAction(t, raw)
	require.Equal(t, "view_only", action["action_code"])
	require.Equal(t, "legacy_unresolved", action["blocked_code"])
	require.Empty(t, action["related_approval_id"], "blocked supplement must not project an approval link")
}

// Defect 7 regression: the revision todo owner is the applicant even when
// the approval was transferred to another approver; the projection must
// keep supplement routable (no blocked_code) instead of view_only.
func TestN5ActionSupplementAfterTransferStaysRoutable(t *testing.T) {
	db := newProjectionWorld(t)
	now := time.Now().UnixNano()
	intakeID := fmt.Sprintf("INT-N5T-%d", now)
	approval := &models.ApprovalRequest{
		ID: fmt.Sprintf("APR-N5T-%d", now), RequestNumber: fmt.Sprintf("RN-N5T-%d", now),
		Title: "N5 转交后补充", Type: "CASE_CREATION", Content: "n5",
		ApplicantID: "u-h-lawyer", ApplicantName: "律师甲",
		Status:            models.ApprovalStatusNeedsRevision,
		CurrentApproverID: "u-h-admin",
		WorkflowType:      "CASE_APPROVAL",
		Metadata:          fmt.Sprintf(`{"intake_id":"%s"}`, intakeID),
	}
	require.NoError(t, db.gorm.Create(approval).Error)
	rec := &models.ApprovalRecord{
		ID: fmt.Sprintf("REC-N5T-%d", now), ApprovalRequestID: approval.ID,
		Stage: "stage-1", StageOrder: 1, ApproverID: "u-h-admin",
		Decision: models.ApprovalDecisionRequestChanges, DecisionReason: "补充说明",
		ImposedRequirements: `{"schema_version":1,"items":[]}`,
		FollowUpActions:     "[]", ApprovalDate: time.Now(),
		SupportingDocuments: "[]", EvidenceReferences: "[]", Status: "active",
	}
	require.NoError(t, db.gorm.Create(rec).Error)
	require.NoError(t, db.approval.openApplicantRevisionTask(db.gorm, approval))

	items := openTasks(t, db.gorm, "u-h-lawyer")
	require.Len(t, items, 1)
	raw := mustProjectOne(t, db, items[0].ID, items[0].UserID)
	action := requireAction(t, raw)
	require.Equal(t, "supplement", action["action_code"])
	require.Empty(t, action["blocked_code"], "transferred revision todo must stay routable")
	require.Equal(t, intakeID, action["resource_id"])
	require.Equal(t, approval.ID, action["related_approval_id"])
}
