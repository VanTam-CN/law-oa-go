package services_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func TestInboxR1PostgresClientEditBeforeResubmitHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	f := services.NewInboxRevisionPostgresFixture(t)
	db, a := f.DB, f.Approval
	router := backendReviewHTTPRouter(db, 11, "lawyer")
	inbox := handlers.NewInboxHandler(services.NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db)))
	home := handlers.NewDemoAggregateHandler(db)
	home.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
	router.GET("/inbox", inbox.ListInboxItems)
	router.GET("/inbox/stats", inbox.GetInboxStats)
	router.GET("/inbox/:id", inbox.GetInboxItem)
	router.GET("/home", home.CommandCenter)
	var item models.InboxItem
	require.NoError(t, db.Where("user_id = ? AND workflow_purpose = ?", 11, "supplement").First(&item).Error)
	var frozen models.ApprovalSnapshot
	require.NoError(t, db.First(&frozen, "id = ?", *a.ActiveSubmissionSnapshotID).Error)
	changed := backendReviewHTTP(router, http.MethodPut, "/case-intakes/"+f.IntakeID, fmt.Sprintf(`{"client_id":%d,"expected_revision":"draft-r1"}`, f.NewClientID))
	require.Equal(t, 200, changed.Code, changed.Body.String())
	var intake struct {
		ClientID uint
		Status   string
	}
	require.NoError(t, db.Table("case_intakes").Where("id = ?", f.IntakeID).Take(&intake).Error)
	require.Equal(t, f.NewClientID, intake.ClientID)
	require.Equal(t, "draft", intake.Status)
	var unchanged models.ApprovalRequest
	require.NoError(t, db.First(&unchanged, "id = ?", a.ID).Error)
	require.Equal(t, a.Metadata, unchanged.Metadata)
	require.Equal(t, a.ActiveSubmissionSnapshotID, unchanged.ActiveSubmissionSnapshotID)
	require.Equal(t, models.ApprovalStatusNeedsRevision, unchanged.Status)
	t.Logf("R1 real PUT client A=%d -> B=%d; intake=draft approval=needs_revision old_active_unchanged=true", f.OldClientID, f.NewClientID)
	decode := func(path string, value interface{}) {
		t.Helper()
		w := backendReviewHTTP(router, http.MethodGet, path, "")
		require.Equal(t, 200, w.Code, w.Body.String())
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
		require.NoError(t, json.Unmarshal(envelope.Data, value))
	}
	assertVisible := func(label string, visible bool) {
		t.Helper()
		before := f.State()
		for _, path := range []string{"/approvals/" + a.ID, "/approvals/" + a.ID + "/snapshot", "/approvals/" + a.ID + "/snapshot?snapshot_id=" + *a.ActiveSubmissionSnapshotID} {
			w := backendReviewHTTP(router, http.MethodGet, path, "")
			status := http.StatusForbidden
			if visible {
				status = http.StatusOK
			}
			require.Equal(t, status, w.Code, "shared revision authorization disagrees: %s", w.Body.String())
		}
		var page services.ListInboxItemsResponse
		decode("/inbox?page=1&page_size=20&workflow_bucket=needs_action", &page)
		var stats services.InboxStatsResponse
		decode("/inbox/stats", &stats)
		require.Equal(t, page.Stats, &stats)
		var dashboard struct {
			Todos []services.InboxItemResponse `json:"todo_items"`
			Stats services.InboxStatsResponse  `json:"inbox_stats"`
		}
		decode("/home", &dashboard)
		require.Equal(t, stats, dashboard.Stats)
		if visible {
			require.Len(t, page.Items, 1, label)
			require.EqualValues(t, 1, page.Pagination.Total)
			require.EqualValues(t, 1, stats.NeedsAction)
			require.Equal(t, item.ID, page.Items[0].ID)
			var detail services.InboxItemResponse
			decode(fmt.Sprintf("/inbox/%d", item.ID), &detail)
			require.Equal(t, "supplement", detail.WorkflowAction.ActionCode)
			require.Equal(t, a.ID, detail.WorkflowAction.ApprovalID)
			require.Equal(t, f.IntakeID, detail.WorkflowAction.ResourceID)
			require.Equal(t, detail.WorkflowAction, page.Items[0].WorkflowAction)
			require.Len(t, dashboard.Todos, 1)
			require.Equal(t, detail.WorkflowAction, dashboard.Todos[0].WorkflowAction)
		} else {
			require.Empty(t, page.Items, label)
			require.Zero(t, page.Pagination.Total)
			require.Zero(t, stats.Total)
			require.Empty(t, dashboard.Todos)
			w := backendReviewHTTP(router, http.MethodGet, fmt.Sprintf("/inbox/%d", item.ID), "")
			require.Equal(t, 404, w.Code, w.Body.String())
		}
		require.Equal(t, before, f.State(), "reads must not rewrite history or workflow state")
		t.Logf("R1 %s visible=%v list_total=%d needs_action=%d home=%d read_only=true", label, visible, page.Pagination.Total, stats.NeedsAction, len(dashboard.Todos))
	}
	assertVisible("client_saved_before_resubmit", true)
	for _, client := range []struct {
		id    uint
		label string
	}{{f.OldClientID, "frozen_client"}, {f.NewClientID, "editable_client"}} {
		wall := &models.Case{CaseNumber: "R1-WALL-" + client.label, Title: "R1授权样本", ClientID: client.id, LawyerID: 11, CreatedBy: "11", EthicalWallEnabled: true}
		require.NoError(t, db.Create(wall).Error)
		assertVisible(client.label+"_revoked", false)
		require.NoError(t, db.Create(&models.CaseEthicalWallWhitelist{CaseID: wall.ID, UserID: 11, GrantedBy: 22}).Error)
		assertVisible(client.label+"_restored", true)
	}
	request := f.PrepareReviewedSubmission()
	assertVisible("new_review_ready_before_resubmit", true)
	raw, err := json.Marshal(request)
	require.NoError(t, err)
	resubmitted := backendReviewHTTP(router, http.MethodPost, "/approvals/"+a.ID+"/resubmit", string(raw))
	require.Equal(t, 200, resubmitted.Code, resubmitted.Body.String())
	var current models.ApprovalRequest
	require.NoError(t, db.First(&current, "id = ?", a.ID).Error)
	require.NotEqual(t, a.ActiveSubmissionSnapshotID, current.ActiveSubmissionSnapshotID)
	partner := backendReviewHTTPRouter(db, 22, "partner")
	// The two synthetic walls remain effective; explicitly grant the deciding
	// partner both sides before its real approval operation.
	var walls []models.Case
	require.NoError(t, db.Where("case_number LIKE ?", "R1-WALL-%").Find(&walls).Error)
	for _, wall := range walls {
		require.NoError(t, db.Create(&models.CaseEthicalWallWhitelist{CaseID: wall.ID, UserID: 22, GrantedBy: 11}).Error)
	}
	decision, err := json.Marshal(map[string]interface{}{"decision": "approve", "decision_reason": "核对补正轮次", "expected_submission_snapshot_id": current.ActiveSubmissionSnapshotID, "expected_action_version": current.UpdatedAt.UTC().Format(time.RFC3339Nano)})
	require.NoError(t, err)
	approved := backendReviewHTTP(partner, http.MethodPost, "/approvals/"+a.ID+"/decision", string(decision))
	require.Equal(t, 200, approved.Code, approved.Body.String())
	require.NoError(t, db.First(&current, "id = ?", a.ID).Error)
	require.True(t, current.CaseCreated)
	require.NotEmpty(t, current.CreatedCaseID)
	var preserved models.ApprovalSnapshot
	require.NoError(t, db.First(&preserved, "id = ?", frozen.ID).Error)
	require.Equal(t, frozen.SnapshotData, preserved.SnapshotData)
	require.Equal(t, a.Metadata, current.Metadata)
	t.Logf("R1 actual HTTP resubmit and approve succeeded new_active=%s case=%s historical_bytes_unchanged=true", *current.ActiveSubmissionSnapshotID, current.CreatedCaseID)
}
