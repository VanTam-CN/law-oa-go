package services_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

func TestApprovalSnapshotModePostgresHTTP(t *testing.T) {
	for _, mode := range []string{"ordinary_leave", "ordinary_with_legacy_snapshot", "legacy_check_binding", "legacy_tracking_binding", "legacy_snapshot_binding"} {
		t.Run(mode, func(t *testing.T) {
			db, legacy, checkID, state := services.ApprovalBackendFinalReviewLegacyHTTPFixture(t, false)
			a := legacy
			if mode != "legacy_check_binding" {
				var err error
				svc := services.NewApprovalService(db)
				a, err = svc.CreateApproval("11", "虚构申请律师", &models.CreateApprovalRequest{
					Title: "普通请假审批", Type: "leave", Content: "虚构请假申请", WorkflowType: "COLLAB_PG",
				})
				require.NoError(t, err)
				a, err = svc.SubmitApproval("11", a.ID)
				require.NoError(t, err)
			}
			if mode == "legacy_tracking_binding" || mode == "legacy_snapshot_binding" {
				var intakeID string
				require.NoError(t, db.Raw("SELECT search_parameters->>'intakeId' FROM conflict_check_records WHERE check_id = ?", checkID).Scan(&intakeID).Error)
				require.NotEmpty(t, intakeID)
				// The durable binding alone must identify an intake approval even
				// when its top-level metadata and type have no intake indication.
				if mode == "legacy_tracking_binding" {
					require.NoError(t, db.Exec(`INSERT INTO approval_case_creation_tracking(id,approval_request_id,intake_id,created_by) VALUES (?,?,?,'11')`, "mode-slot", a.ID, intakeID).Error)
				} else {
					require.NoError(t, db.Create(&models.ApprovalSnapshot{ApprovalRequestID: a.ID, SnapshotType: "conflict_approval", SourceVersion: 1,
						SnapshotData: fmt.Sprintf(`{"metadata":{"intake_id":%q}}`, intakeID)}).Error)
				}
			}
			if mode == "ordinary_with_legacy_snapshot" {
				require.NoError(t, db.Create(&models.ApprovalSnapshot{ApprovalRequestID: a.ID, SnapshotType: "legacy", SourceVersion: 1, SnapshotData: `{"metadata":{}}`}).Error)
			}
			r := backendReviewHTTPRouter(db, 22, "partner")
			before := state()
			detail := backendReviewHTTP(r, http.MethodGet, "/approvals/"+a.ID, "")
			require.Equal(t, 200, detail.Code, detail.Body.String())
			var envelope struct {
				Data struct {
					Required *bool `json:"submission_snapshot_required"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(detail.Body.Bytes(), &envelope))
			require.NotNil(t, envelope.Data.Required, "detail must expose an explicit server-resolved mode")
			require.Equal(t, mode != "ordinary_leave", *envelope.Data.Required)
			if mode == "ordinary_leave" {
				snapshot := backendReviewHTTP(r, http.MethodGet, "/approvals/"+a.ID+"/snapshot", "")
				require.Equal(t, 404, snapshot.Code, "ordinary leave legitimately has no frozen snapshot: %s", snapshot.Body.String())
				const fault = "snapshot-mode-read-failure"
				require.NoError(t, db.Callback().Query().Before("gorm:query").Register(fault, func(tx *gorm.DB) {
					if tx.Statement.Table == "approval_snapshots" {
						tx.AddError(errors.New("injected snapshot read failure"))
					}
				}))
				failed := backendReviewHTTP(r, http.MethodGet, "/approvals/"+a.ID, "")
				require.NoError(t, db.Callback().Query().Remove(fault))
				require.Equal(t, 500, failed.Code, failed.Body.String())
				require.NotContains(t, failed.Body.String(), `"submission_snapshot_required":false`)
			}
			require.Equal(t, before, state(), "mode projection must never persist or claim a slot")
			decision := backendReviewHTTP(r, http.MethodPost, "/approvals/"+a.ID+"/decision", `{"decision":"approve","decision_reason":"确认普通审批"}`)
			if mode == "ordinary_leave" || mode == "ordinary_with_legacy_snapshot" {
				require.Equal(t, 200, decision.Code, decision.Body.String())
				var saved models.ApprovalRequest
				require.NoError(t, db.First(&saved, "id = ?", a.ID).Error)
				require.Equal(t, models.ApprovalStatusApproved, saved.Status)
			} else {
				require.Equal(t, 409, decision.Code, decision.Body.String())
				require.Equal(t, before, state(), "legacy intake without an active pointer/slot remains fail closed")
			}
		})
	}
}

func TestApprovalSnapshotModePostgresCaseCreationAndEmbeddedMode(t *testing.T) {
	db, legacy, _, state := services.ApprovalBackendFinalReviewLegacyHTTPFixture(t, false)
	before := state()
	for _, a := range []models.ApprovalRequest{
		{ID: "mode-not-persisted", Type: "case_creation", Metadata: "{}"},
		{ID: "mode-not-persisted", Type: "leave", Metadata: `{"approval_snapshot":{"metadata":{}}}`},
	} {
		required, err := services.ApprovalRequiresSubmissionSnapshot(db, &a)
		require.NoError(t, err)
		require.True(t, required, a.Type)
	}
	// A persisted check binding must remain authoritative even if all editable
	// type/metadata hints claim this is ordinary business.
	legacy.Type, legacy.Metadata = "leave", "{}"
	required, err := services.ApprovalRequiresSubmissionSnapshot(db, legacy)
	require.NoError(t, err)
	require.True(t, required)
	require.Equal(t, before, state())
}
