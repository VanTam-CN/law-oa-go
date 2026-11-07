package services_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// Reuse the actual handlers and the existing private-schema PG fixture. Only
// authentication is injected; authorization, projection and writes are real.
func inboxC06Router(db *gorm.DB, actor uint, role string) *gin.Engine {
	r := backendReviewHTTPRouter(db, actor, role)
	inbox := handlers.NewInboxHandler(services.NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db)))
	home := handlers.NewDemoAggregateHandler(db)
	home.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
	r.GET("/inbox", inbox.ListInboxItems)
	r.GET("/inbox/stats", inbox.GetInboxStats)
	r.GET("/inbox/:id", inbox.GetInboxItem)
	r.GET("/home", home.CommandCenter)
	return r
}

func inboxC06Decode(t *testing.T, body []byte, target interface{}) {
	t.Helper()
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	require.True(t, envelope.Success, string(body))
	require.NoError(t, json.Unmarshal(envelope.Data, target))
}

func TestInboxC06PostgresHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("assistant_controlled_write", func(t *testing.T) {
		db, approval, request, intakeID, _, _, state := services.ApprovalBackendFinalReviewPendingClientFixture(t)
		require.NoError(t, db.Create(&models.User{ID: 44, Username: "c06-assistant", Email: "c06-assistant@example.invalid", Password: "fixture-only", Role: "assistant", Status: "active"}).Error)
		router := inboxC06Router(db, 44, "assistant")
		var item models.InboxItem
		require.NoError(t, db.Where("user_id = ? AND workflow_purpose = ?", 11, "supplement").First(&item).Error)
		before := state()
		for _, path := range []string{"/inbox", "/inbox/stats", "/home"} {
			w := backendReviewHTTP(router, http.MethodGet, path, "")
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
			require.NotContains(t, w.Body.String(), approval.ID)
			require.NotContains(t, w.Body.String(), intakeID)
			require.NotContains(t, w.Body.String(), approval.Title)
			if path == "/inbox" {
				var page services.ListInboxItemsResponse
				inboxC06Decode(t, w.Body.Bytes(), &page)
				require.Empty(t, page.Items)
				require.Zero(t, page.Pagination.Total)
			}
		}
		w := backendReviewHTTP(router, http.MethodGet, fmt.Sprintf("/inbox/%d", item.ID), "")
		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		payload, err := json.Marshal(map[string]interface{}{"approval_id": approval.ID, "return_record_id": request.ReturnRecordID, "expected_revision": request.ExpectedIntakeRevision, "supplement_responses": request.Responses})
		require.NoError(t, err)
		raw, err := json.Marshal(request)
		require.NoError(t, err)
		for _, endpoint := range []struct{ method, path, body string }{
			{http.MethodPut, "/case-intakes/" + intakeID, string(payload)},
			{http.MethodPost, "/approvals/" + approval.ID + "/resubmit", string(raw)},
		} {
			denied := backendReviewHTTP(router, endpoint.method, endpoint.path, endpoint.body)
			require.Equal(t, http.StatusForbidden, denied.Code, denied.Body.String())
			for _, secret := range []string{approval.Title, request.ReturnRecordID, "310110199001010011", "310110199001010044"} {
				require.NotContains(t, denied.Body.String(), secret)
			}
			require.Equal(t, before, state(), "assistant HTTP denial must not alter the 12 workflow tables")
			t.Logf("C06 assistant method=%s HTTP=%d business_tables_unchanged=true", endpoint.method, denied.Code)
		}
		// Same valid payload succeeds for the actual applicant, excluding stale
		// request/setup errors as the explanation for the assistant rejection.
		owner := inboxC06Router(db, 11, "lawyer")
		saved := backendReviewHTTP(owner, http.MethodPut, "/case-intakes/"+intakeID, string(payload))
		require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())
		var response struct {
			Revision string `json:"revision"`
		}
		inboxC06Decode(t, saved.Body.Bytes(), &response)
		require.NotEmpty(t, response.Revision)
		request.ExpectedIntakeRevision = response.Revision
		raw, err = json.Marshal(request)
		require.NoError(t, err)
		submitted := backendReviewHTTP(owner, http.MethodPost, "/approvals/"+approval.ID+"/resubmit", string(raw))
		require.Equal(t, http.StatusOK, submitted.Code, submitted.Body.String())
		var persisted models.ApprovalRequest
		require.NoError(t, db.First(&persisted, "id = ?", approval.ID).Error)
		require.Equal(t, models.ApprovalStatusResubmitted, persisted.Status)
		require.NotEqual(t, approval.ActiveSubmissionSnapshotID, persisted.ActiveSubmissionSnapshotID)
		t.Log("C06 assistant List/GET/home denied target; applicant same-payload save=200 resubmit=200 new_active=true")
	})

	t.Run("conflict_officer_revoked_context", func(t *testing.T) {
		db, _, caseID, state := services.ApprovalBackendFinalReviewSnapshotFixture(t)
		var matter models.Case
		require.NoError(t, db.First(&matter, caseID).Error)
		// New pending check: do not delete/rewrite the fixture's completed review
		// or immutable approval snapshot to manufacture an actionable task.
		now := time.Now().UTC().Truncate(time.Microsecond)
		checkID := "C06-" + uuid.NewString()
		check := models.ConflictCheckRecord{CheckID: checkID, ClientID: fmt.Sprint(matter.ClientID), ClientName: "C06_PRIVATE_CLIENT", CaseName: "C06_PRIVATE_TITLE", UserID: 11, CheckStatus: "COMPLETED", CheckTime: now,
			SearchParameters: models.JSON{"subjectCaseId": fmt.Sprint(caseID)}, CheckResult: models.JSON{"checkId": checkID, "decision": map[string]interface{}{"coverageStatus": "COMPLETE"}}}
		require.NoError(t, db.Create(&check).Error)
		assignment := models.ConflictReviewerAssignment{ID: uuid.NewString(), CheckID: checkID, CaseID: &caseID, ReviewerID: 33, AssignedBy: 22, Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, EffectiveFrom: &now, CreatedAt: now, UpdatedAt: now}
		require.NoError(t, db.Create(&assignment).Error)
		require.NoError(t, db.Model(&matter).Update("title", "C06_PRIVATE_TITLE").Error)
		require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return services.ReconcileConflictInboxInTx(tx, checkID) }))
		var item models.InboxItem
		require.NoError(t, db.Where("user_id = ? AND source_ref = ? AND workflow_purpose = ?", 33, checkID, "review_conflict").First(&item).Error)
		secrets := []string{"C06_PRIVATE_TITLE", "C06_PRIVATE_CLIENT", "C06_PRIVATE_IDENTITY", "C06_PRIVATE_REVIEW_NOTE"}
		require.NoError(t, db.Model(&item).Updates(map[string]interface{}{"title": secrets[0], "content": secrets[2] + " " + secrets[3]}).Error)
		router := inboxC06Router(db, 33, "conflict_officer")
		readPhase := func(label string, visible bool) {
			t.Helper()
			before := state()
			var page services.ListInboxItemsResponse
			var stats services.InboxStatsResponse
			var home struct {
				Todos []services.InboxItemResponse `json:"todo_items"`
				Stats services.InboxStatsResponse  `json:"inbox_stats"`
			}
			for _, endpoint := range []struct {
				path   string
				target interface{}
			}{
				{"/inbox?page=1&page_size=1", &page}, {"/inbox/stats", &stats}, {"/home", &home},
			} {
				w := backendReviewHTTP(router, http.MethodGet, endpoint.path, "")
				require.Equal(t, http.StatusOK, w.Code, w.Body.String())
				inboxC06Decode(t, w.Body.Bytes(), endpoint.target)
				if !visible {
					for _, secret := range append(secrets, checkID) {
						require.NotContains(t, w.Body.String(), secret, endpoint.path)
					}
				}
			}
			require.Equal(t, page.Stats, &stats)
			require.Equal(t, stats, home.Stats)
			detail := backendReviewHTTP(router, http.MethodGet, fmt.Sprintf("/inbox/%d", item.ID), "")
			if visible {
				require.Equal(t, http.StatusOK, detail.Code, detail.Body.String())
				var one services.InboxItemResponse
				inboxC06Decode(t, detail.Body.Bytes(), &one)
				require.Len(t, page.Items, 1)
				require.Len(t, home.Todos, 1)
				require.EqualValues(t, 1, page.Pagination.Total)
				require.EqualValues(t, 1, stats.NeedsAction)
				require.Equal(t, item.ID, one.ID)
				require.Equal(t, "review_conflict", one.WorkflowAction.ActionCode)
				require.Equal(t, checkID, one.WorkflowAction.ResourceID)
				require.Equal(t, one.WorkflowAction, page.Items[0].WorkflowAction)
				require.Equal(t, one.WorkflowAction, home.Todos[0].WorkflowAction)
				for _, secret := range []string{secrets[0], secrets[2], secrets[3]} {
					require.Contains(t, detail.Body.String(), secret, "positive exposure control")
				}
			} else {
				require.Empty(t, page.Items)
				require.Empty(t, home.Todos)
				require.Zero(t, page.Pagination.Total)
				require.Zero(t, stats.Total)
				require.Zero(t, stats.NeedsAction)
				require.Equal(t, http.StatusNotFound, detail.Code, detail.Body.String())
				for _, secret := range append(secrets, checkID) {
					require.NotContains(t, detail.Body.String(), secret)
				}
			}
			require.Equal(t, before, state(), "reads must not close tasks or rewrite workflow/history")
			var stored models.InboxItem
			require.NoError(t, db.First(&stored, item.ID).Error)
			require.False(t, stored.IsCompleted, "revoked visibility must not be achieved by completing the task")
			t.Logf("C06 officer phase=%s total=%d needs_action=%d home=%d GET=%d same_task=%d read_only=true", label, page.Pagination.Total, stats.NeedsAction, len(home.Todos), detail.Code, item.ID)
		}
		readPhase("authorized", true)
		require.NoError(t, db.Model(&matter).Update("ethical_wall_enabled", true).Error)
		readPhase("object_permission_revoked", false)
		require.NoError(t, db.Create(&models.CaseEthicalWallWhitelist{CaseID: caseID, UserID: 33, GrantedBy: 22}).Error)
		readPhase("permission_restored", true)
	})
}
