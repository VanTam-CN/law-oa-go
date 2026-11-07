package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func TestInboxActiveSnapshotBindingFailClosed(t *testing.T) {
	for _, fault := range []string{"missing", "blank_pointer", "foreign_approval", "wrong_type", "invalid_json", "snapshot_id", "approval_id", "round", "zero_round", "intake_id", "original_intake", "check_id", "task_id", "unauthorized_check"} {
		t.Run(fault, func(t *testing.T) {
			db, svc, a := seedInboxWorkflow(t)
			inboxReconcile(t, db, a)
			var item models.InboxItem
			require.NoError(t, db.Where("user_id = ?", 2).First(&item).Error)
			var row models.ApprovalSnapshot
			require.NoError(t, db.First(&row, "id = ?", *a.ActiveSubmissionSnapshotID).Error)
			snapshot := submissionMap(row.SnapshotData)
			frozen := snapshot["metadata"].(map[string]interface{})
			updates := map[string]interface{}{}
			switch fault {
			case "missing", "blank_pointer":
				pointer := "absent-snapshot"
				if fault == "blank_pointer" {
					pointer = ""
				}
				require.NoError(t, db.Model(a).Update("active_submission_snapshot_id", pointer).Error)
			case "foreign_approval":
				updates["approval_request_id"] = "different-approval"
			case "wrong_type":
				updates["snapshot_type"] = "conflict_approval"
			case "invalid_json":
				updates["snapshot_data"] = "{"
			case "snapshot_id":
				snapshot["snapshot_id"] = "wrong"
			case "approval_id":
				snapshot["approval_request_id"] = "wrong"
			case "round":
				snapshot["source_version"] = 2
			case "zero_round":
				updates["source_version"] = 0
				snapshot["source_version"] = 0
			case "intake_id":
				snapshot["intake_id"] = "wrong"
			case "original_intake":
				require.NoError(t, db.Model(a).Update("metadata", `{"intake_id":"wrong"}`).Error)
			case "check_id":
				snapshot["conflict_check_id"] = "wrong"
			case "task_id":
				frozen["conflict_task_id"] = "wrong"
			case "unauthorized_check":
				snapshot["conflict_check_id"] = "unrelated-check"
				frozen["conflict_check_id"] = "unrelated-check"
				require.NoError(t, db.Create(&models.ConflictCheckRecord{CheckID: "unrelated-check", UserID: 4, ClientID: "999", CheckStatus: "COMPLETED", SearchParameters: models.JSON{"intakeId": "unrelated-intake"}}).Error)
			}
			if fault != "invalid_json" {
				raw, err := json.Marshal(snapshot)
				require.NoError(t, err)
				updates["snapshot_data"] = string(raw)
			}
			// Corrupt only a disposable SQLite fixture, bypassing append-only model
			// hooks so the reader is tested against independently inconsistent data.
			require.NoError(t, db.Table("approval_snapshots").Where("id = ?", row.ID).Updates(updates).Error)
			page, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{})
			require.NoError(t, err)
			require.Empty(t, page.Items)
			require.Zero(t, page.Pagination.Total)
			stats, err := svc.GetInboxStats(context.Background(), 2)
			require.NoError(t, err)
			require.Zero(t, stats.Total)
			require.Equal(t, stats, page.Stats)
			_, err = svc.GetInboxItemByID(context.Background(), item.ID, 2)
			require.ErrorIs(t, err, repositories.ErrInboxItemNotFound)
		})
	}
}

func TestInboxActiveSnapshotSchemaErrorIsNotEmptySuccess(t *testing.T) {
	db, svc, a := seedInboxWorkflow(t)
	inboxReconcile(t, db, a)
	require.NoError(t, db.Migrator().DropTable(&models.ApprovalSnapshot{}))
	_, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{})
	require.ErrorContains(t, err, "approval_snapshots")
	_, err = svc.GetInboxStats(context.Background(), 2)
	require.ErrorContains(t, err, "approval_snapshots")
}

func TestInboxLegacyMetadataConsistencyStillRequired(t *testing.T) {
	db, svc, a := seedInboxWorkflow(t)
	require.NoError(t, db.Model(a).Updates(map[string]interface{}{"active_submission_snapshot_id": nil, "metadata": `{"intake_id":"intake-uuid","conflict_check_id":"other-check"}`}).Error)
	inboxReconcile(t, db, a)
	page, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{})
	require.NoError(t, err)
	require.Empty(t, page.Items)
	require.Zero(t, page.Stats.Total)
}

func TestInboxActiveSnapshotsBatchAcrossApprovals(t *testing.T) {
	db, svc, a := seedInboxWorkflow(t)
	inboxReconcile(t, db, a)
	queries, snapshotQueries := 0, 0
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("inbox:active-batch", func(tx *gorm.DB) {
		queries++
		if tx.Statement.Table == "approval_snapshots" {
			snapshotQueries++
		}
	}))
	_, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{})
	require.NoError(t, err)
	baseline := queries
	require.Equal(t, 1, snapshotQueries)
	for i := 0; i < 75; i++ {
		clone := *a
		clone.ID = fmt.Sprintf("approval-batch-%d", i)
		clone.RequestNumber = fmt.Sprintf("APR-BATCH-%d", i)
		pointer := fmt.Sprintf("snapshot-batch-%d", i)
		clone.ActiveSubmissionSnapshotID = &pointer
		require.NoError(t, db.Create(&clone).Error)
		seedInboxSubmission(t, db, &clone)
		inboxReconcile(t, db, &clone)
	}
	queries, snapshotQueries = 0, 0
	page, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{PageSize: 100, WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 76)
	require.EqualValues(t, 76, page.Stats.NeedsAction)
	require.Equal(t, 1, snapshotQueries, "76 distinct active snapshots must use one batch query")
	require.Equal(t, baseline, queries, "query count must not grow per approval or snapshot")
	t.Logf("approval_count=76 snapshot_queries=%d total_queries=%d one_approval_queries=%d", snapshotQueries, queries, baseline)
}
