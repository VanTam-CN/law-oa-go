package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func inboxRevisionClientChange(t *testing.T) (*gorm.DB, *InboxService, *models.ApprovalRequest) {
	t.Helper()
	db, svc, a := seedInboxWorkflow(t)
	a.Status = models.ApprovalStatusNeedsRevision
	require.NoError(t, db.Save(a).Error)
	inboxReconcile(t, db, a)
	require.NoError(t, db.Create(&models.Client{ID: 200, Name: "当前补正客户", Type: "个人", Email: "revision-200@test.invalid", CreatedBy: 1}).Error)
	require.NoError(t, db.Table("case_intakes").Where("id = ?", "intake-uuid").Update("client_id", 200).Error)
	return db, svc, a
}

func TestInboxR1OnlyActiveRevisionMayChangeClient(t *testing.T) {
	for _, variant := range []string{"needs_revision", "submitted", "resubmitted", "under_review", "approved", "legacy", "wrong_intake_owner", "wrong_check_owner", "wrong_check_intake"} {
		t.Run(variant, func(t *testing.T) {
			db, svc, a := inboxRevisionClientChange(t)
			switch variant {
			case "legacy":
				require.NoError(t, db.Model(a).Update("active_submission_snapshot_id", nil).Error)
			case "wrong_intake_owner":
				require.NoError(t, db.Table("case_intakes").Where("id = ?", "intake-uuid").Update("created_by", "4").Error)
			case "wrong_check_owner":
				require.NoError(t, db.Model(&models.ConflictCheckRecord{}).Where("check_id = ?", "check-uuid").Update("user_id", 4).Error)
			case "wrong_check_intake":
				require.NoError(t, db.Model(&models.ConflictCheckRecord{}).Where("check_id = ?", "check-uuid").Update("search_parameters", models.JSON{"intakeId": "unrelated"}).Error)
			case "needs_revision":
			default:
				require.NoError(t, db.Model(a).Update("status", variant).Error)
			}
			page, err := svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{})
			require.NoError(t, err)
			if variant == "needs_revision" {
				require.Len(t, page.Items, 1)
				require.Equal(t, "supplement", page.Items[0].WorkflowAction.ActionCode)
			} else {
				require.Empty(t, page.Items, "the correction exception must not widen other states or bindings")
				require.Zero(t, page.Stats.Total)
			}
		})
	}
}

func TestInboxR1FrozenReferencesRemainAuthorized(t *testing.T) {
	for _, reference := range []string{"root_client", "metadata_client", "frozen_record_client", "config_client", "root_case", "frozen_case", "invalid_client", "missing_client", "missing_case", "missing_check"} {
		t.Run(reference, func(t *testing.T) {
			db, svc, a := inboxRevisionClientChange(t)
			require.NoError(t, db.Create(&models.Client{ID: 300, Name: "独立冻结对象", Type: "个人", Email: "revision-300@test.invalid", CreatedBy: 4}).Error)
			require.NoError(t, db.Create(&models.Case{ID: 30, CaseNumber: "FROZEN-WALL", Title: "冻结对象", ClientID: 300, LawyerID: 4, CreatedBy: "4", EthicalWallEnabled: true}).Error)
			var row models.ApprovalSnapshot
			require.NoError(t, db.First(&row, "id = ?", *a.ActiveSubmissionSnapshotID).Error)
			snapshot := submissionMap(row.SnapshotData)
			metadata := snapshot["metadata"].(map[string]interface{})
			switch reference {
			case "root_client":
				snapshot["client_id"] = 300
			case "metadata_client":
				metadata["client"] = map[string]interface{}{"id": 300}
			case "frozen_record_client":
				metadata["conflict_record"] = map[string]interface{}{"client_id": 300}
			case "config_client":
				snapshot["case_creation_config"] = map[string]interface{}{"client_id": 300}
			case "root_case":
				snapshot["subject_case_id"] = 30
			case "frozen_case":
				metadata["conflict_record"] = map[string]interface{}{"search_parameters": map[string]interface{}{"subjectCaseId": "30"}}
			case "invalid_client":
				metadata["client_id"] = "not-an-id"
			case "missing_client":
				metadata["client_id"] = 999
			case "missing_case":
				snapshot["case_id"] = 999
			case "missing_check":
				metadata["conflict_record"] = map[string]interface{}{"check_id": "absent"}
			}
			raw, err := json.Marshal(snapshot)
			require.NoError(t, err)
			// Deliberately inconsistent, isolated fixture; no production history edit.
			require.NoError(t, db.Table("approval_snapshots").Where("id = ?", row.ID).Update("snapshot_data", string(raw)).Error)
			page, err := svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{})
			require.NoError(t, err)
			require.Empty(t, page.Items)
			require.Zero(t, page.Stats.Total)
		})
	}
}

func TestInboxR1BatchAuthorizationAcrossRevisions(t *testing.T) {
	db, svc, a := inboxRevisionClientChange(t)
	queries, snapshots, clients := 0, 0, 0
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("inbox:r1-batch", func(tx *gorm.DB) {
		queries++
		if tx.Statement.Table == "approval_snapshots" {
			snapshots++
		}
		if tx.Statement.Table == "clients" {
			clients++
		}
	}))
	_, err := svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{})
	require.NoError(t, err)
	baseline := queries
	for i := 0; i < 75; i++ {
		clone := *a
		clone.ID, clone.RequestNumber = fmt.Sprintf("revision-%d", i), fmt.Sprintf("REV-%d", i)
		pointer := fmt.Sprintf("revision-snapshot-%d", i)
		clone.ActiveSubmissionSnapshotID = &pointer
		require.NoError(t, db.Create(&clone).Error)
		seedInboxSubmission(t, db, &clone)
		inboxReconcile(t, db, &clone)
	}
	queries, snapshots, clients = 0, 0, 0
	page, err := svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{PageSize: 100, WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 76)
	require.EqualValues(t, 76, page.Stats.NeedsAction)
	require.Equal(t, baseline, queries)
	require.Equal(t, 1, snapshots)
	require.Equal(t, 1, clients)
	t.Logf("R1 76 revisions: total_queries=%d one_revision_queries=%d snapshot_queries=%d client_queries=%d", queries, baseline, snapshots, clients)
}
