package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// Reuse the isolated PG fixture and real business services. The C2 check and
// independent review are persisted fixture inputs; resubmission itself is real.
func inboxB2RevisedFixture(t *testing.T) *approvalCollaborationPostgresFixture {
	t.Helper()
	f := newApprovalCollaborationPostgresFixture(t)
	original := f.reload(t, f.approval.ID)
	var first models.ApprovalSnapshot
	require.NoError(t, f.db.First(&first, "id = ?", *original.ActiveSubmissionSnapshotID).Error)
	req := f.returnForRevision(t)
	f.reviseReviewedSubject(t, req)
	updated, err := NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
	require.NoError(t, err)
	require.NotEqual(t, original.ActiveSubmissionSnapshotID, updated.ActiveSubmissionSnapshotID)
	durable := f.reload(t, f.approval.ID)
	require.Equal(t, f.checkID, durable.ConflictCheckID)
	require.Equal(t, original.Metadata, durable.Metadata, "resubmit must preserve historical metadata bytes")
	require.NotEqual(t, f.checkID, metadataString(parseMetadata(durable.Metadata), "conflict_check_id"))
	var preserved models.ApprovalSnapshot
	require.NoError(t, f.db.First(&preserved, "id = ?", first.ID).Error)
	require.Equal(t, first.SnapshotData, preserved.SnapshotData)
	t.Logf("B2 real resubmit approval=%s intake=%s C1=%s C2=%s active=%s historical_metadata_unchanged=true", durable.ID, f.intakeID, original.ConflictCheckID, f.checkID, *durable.ActiveSubmissionSnapshotID)
	return f
}

// Export only test data for the external test package to exercise real HTTP
// inbox and home handlers without a services -> handlers import cycle.
func InboxB2PostgresHTTPFixture(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	f := inboxB2RevisedFixture(t)
	return f.db, f.approval.ID
}

func TestInboxB2PostgresRevisedSubmission(t *testing.T) {
	f := inboxB2RevisedFixture(t)
	svc := NewInboxService(repositories.NewInboxRepository(f.db), repositories.NewUserRepository(f.db))
	assertBucket := func(user uint, bucket, code string) *InboxItemResponse {
		t.Helper()
		before := f.state(t)
		page, err := svc.ListInboxItems(context.Background(), user, &ListInboxItemsRequest{WorkflowBucket: bucket})
		require.NoError(t, err)
		require.Len(t, page.Items, 1, "authorized new round must remain visible")
		require.EqualValues(t, 1, page.Pagination.Total)
		require.Equal(t, code, page.Items[0].WorkflowAction.ActionCode)
		stats, err := svc.GetInboxStats(context.Background(), user)
		require.NoError(t, err)
		require.Equal(t, page.Stats, stats)
		if bucket == "needs_action" {
			require.EqualValues(t, 1, stats.NeedsAction)
		} else {
			require.EqualValues(t, 1, stats.Waiting)
		}
		detail, err := svc.GetInboxItemByID(context.Background(), page.Items[0].ID, user)
		require.NoError(t, err)
		require.Equal(t, page.Items[0].WorkflowAction, detail.WorkflowAction)
		require.Equal(t, f.state(t), before, "list/count/detail must never rewrite historical facts")
		return detail
	}
	assertBucket(22, "needs_action", "decide_approval")
	assertBucket(11, "waiting", "view_only")
	// A second return must use the C2 round too, including the explicit approval ID.
	_, err := NewApprovalService(f.db).ProcessApprovalDecision("22", f.approval.ID, f.decision(t, f.approval.ID, models.ApprovalDecisionRequestChanges))
	require.NoError(t, err)
	supplement := assertBucket(11, "needs_action", "supplement")
	require.Equal(t, f.intakeID, supplement.WorkflowAction.ResourceID)
	require.Equal(t, f.approval.ID, supplement.WorkflowAction.ApprovalID)
	var returned models.ApprovalRecord
	require.NoError(t, f.db.Where("approval_request_id = ? AND decision = ?", f.approval.ID, models.ApprovalDecisionRequestChanges).Order("created_at DESC, id DESC").First(&returned).Error)
	items, err := ReadSupplementRequests(&returned)
	require.NoError(t, err)
	current := f.reload(t, f.approval.ID)
	req := &models.ResubmitApprovalRequest{ExpectedIntakeRevision: "draft-r2", ExpectedSubmissionSnapshotID: *current.ActiveSubmissionSnapshotID,
		ReturnRecordID: returned.ID, RevisionNote: "补充C2事项说明", Responses: []models.SupplementResponse{{ItemID: items[0].ItemID, Response: "已补充C2说明"}}}
	_, err = NewApprovalService(f.db).ResubmitIntakeApproval("11", f.approval.ID, req)
	require.NoError(t, err)
	assertBucket(22, "needs_action", "decide_approval")
	_, err = f.integrated().ProcessApprovalWithConflict(context.Background(), "22", "虚构审批合伙人", f.approval.ID, f.decision(t, f.approval.ID, models.ApprovalDecisionApprove))
	require.NoError(t, err)
	caseID := f.assertOneCompleteCase(t, f.approval.ID)
	var result models.InboxItem
	require.NoError(t, f.db.Where("workflow_group = ? AND workflow_purpose = ?", "approval:"+f.approval.ID, "result").First(&result).Error)
	detail, err := svc.GetInboxItemByID(context.Background(), result.ID, 11)
	require.NoError(t, err)
	require.Equal(t, "case_created", detail.WorkflowAction.WorkflowStatus)
	require.EqualValues(t, 1, f.count(t, "inbox_items", "workflow_group = ? AND workflow_purpose = ?", "approval:"+f.approval.ID, "result"))
	t.Logf("B2 list/stats/get current-round agreement=true second_return_supplement=true case=%s result=%d", caseID, result.ID)
}
