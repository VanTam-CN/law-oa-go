package services_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func backendReviewHTTPRouter(db *gorm.DB, actor uint, role string) *gin.Engine {
	auth := services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil)
	ordinary := handlers.NewApprovalHandler(db)
	ordinary.SetAuthorizationService(auth)
	caseService := services.NewCaseService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db))
	caseService.SetSubjectRecheckService(services.NewSubjectRecheckService(db, nil))
	service := services.NewApprovalConflictIntegrationService(ordinary.GetApprovalService(), nil, caseService, repositories.NewIntegrationRepository(db), db)
	ordinary.SetIntegratedDecisionService(service)
	integrated := handlers.NewIntegrationHandler(service, nil, db)
	integrated.SetAuthorizationService(auth)
	aggregate := handlers.NewDemoAggregateHandler(db)
	aggregate.SetAuthorizationService(auth)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", actor)
		c.Set("role", role)
		c.Set("username", "虚构回归用户")
		c.Next()
	})
	r.POST("/approvals/:id/decision", ordinary.ProcessApprovalDecision)
	r.POST("/integration/approvals/:id/decision", integrated.ProcessApprovalWithConflict)
	r.POST("/integration/approvals/:id/retry", integrated.RetryFailedIntegration)
	r.POST("/conflict/tasks/:task_id/approval", aggregate.CreateConflictApproval)
	r.GET("/approvals/:id/snapshot", aggregate.GetApprovalSnapshot)
	r.GET("/approvals/:id", ordinary.GetApproval)
	r.POST("/approvals/:id/resubmit", ordinary.ResubmitApproval)
	r.PUT("/case-intakes/:id", aggregate.UpdateCaseIntake)
	return r
}

func backendReviewHTTP(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func TestBackendFinalReviewPostgresHTTPLegacyBoundaries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range []string{"ordinary", "integrated", "retry", "create_old_ticket"} {
		t.Run(entry, func(t *testing.T) {
			db, a, checkID, state := services.ApprovalBackendFinalReviewLegacyHTTPFixture(t, entry == "retry")
			actor, role := uint(22), "partner"
			path, body := "/approvals/"+a.ID+"/decision", `{"decision":"approve","decision_reason":"确认审批结论"}`
			code := "INTAKE_SUBMISSION_REQUIRED"
			switch entry {
			case "integrated":
				path = "/integration/approvals/" + a.ID + "/decision"
			case "retry":
				actor, role = 11, "lawyer"
				path, body = "/integration/approvals/"+a.ID+"/retry", `{"retry_type":"case_creation"}`
			case "create_old_ticket":
				actor, role = 11, "lawyer"
				path, body, code = "/conflict/tasks/"+checkID+"/approval", `{}`, "INTAKE_CONFLICT_REVIEW_REQUIRED"
			}
			r := backendReviewHTTPRouter(db, actor, role)
			before := state()
			response := backendReviewHTTP(r, http.MethodPost, path, body)
			require.Equal(t, 409, response.Code, response.Body.String())
			require.Contains(t, response.Body.String(), code)
			require.Equal(t, before, state(), "actual HTTP rejection must not write")
		})
	}
}

func TestBackendFinalReviewPostgresHistoricalSnapshotClientRevoked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, actor := range []struct {
		id   uint
		role string
	}{{11, "lawyer"}, {22, "partner"}, {33, "conflict_officer"}} {
		t.Run(actor.role, func(t *testing.T) {
			db, approval, oldSnapshot, oldCase, state := services.ApprovalBackendFinalReviewHistoricalClientFixture(t)
			if actor.id == 33 {
				require.NoError(t, db.Model(approval).Update("current_approver_id", "33").Error)
			}
			r := backendReviewHTTPRouter(db, actor.id, actor.role)
			path := "/approvals/" + approval.ID + "/snapshot"
			get := func(query string) *httptest.ResponseRecorder {
				return backendReviewHTTP(r, http.MethodGet, path+query, "")
			}
			require.Equal(t, 200, get("?snapshot_id="+oldSnapshot).Code)
			require.NoError(t, db.Model(&models.Case{}).Where("id = ?", oldCase).Update("ethical_wall_enabled", true).Error)
			before := state()
			current := get("")
			require.Equal(t, 200, current.Code, current.Body.String())
			var envelope struct {
				Data struct {
					History []struct {
						ID string `json:"id"`
					} `json:"submission_history"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(current.Body.Bytes(), &envelope))
			require.Len(t, envelope.Data.History, 1)
			require.Equal(t, *approval.ActiveSubmissionSnapshotID, envelope.Data.History[0].ID)
			require.NotContains(t, current.Body.String(), oldSnapshot, "revoked old ID must not leak through history metadata")
			denied := get("?snapshot_id=" + oldSnapshot)
			require.Equal(t, 403, denied.Code, denied.Body.String())
			require.NotContains(t, denied.Body.String(), "snapshot")
			require.Equal(t, before, state(), "reading history must never rewrite stored snapshots or approval metadata")
			require.NoError(t, db.Create(&models.CaseEthicalWallWhitelist{CaseID: oldCase, UserID: actor.id, GrantedBy: 22}).Error)
			restored := get("?snapshot_id=" + oldSnapshot)
			require.Equal(t, 200, restored.Code, restored.Body.String())
		})
	}
}

func TestBackendFinalReviewPostgresSnapshotRevokedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, actor := range []struct {
		id   uint
		role string
	}{{11, "lawyer"}, {22, "partner"}, {33, "conflict_officer"}} {
		t.Run(actor.role, func(t *testing.T) {
			db, approval, caseID, state := services.ApprovalBackendFinalReviewSnapshotFixture(t)
			if actor.id == 33 {
				require.NoError(t, db.Model(approval).Update("current_approver_id", "33").Error)
			}
			h := handlers.NewDemoAggregateHandler(db)
			h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
			r := gin.New()
			r.Use(func(c *gin.Context) { c.Set("user_id", actor.id); c.Set("role", actor.role); c.Next() })
			r.GET("/approvals/:id/snapshot", h.GetApprovalSnapshot)
			get := func(query string) *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/approvals/"+approval.ID+"/snapshot"+query, nil))
				return w
			}
			require.Equal(t, 200, get("").Code)
			require.NoError(t, db.Model(&models.Case{}).Where("id = ?", caseID).Update("ethical_wall_enabled", true).Error)
			before := state()
			for _, query := range []string{"", "?snapshot_id=" + *approval.ActiveSubmissionSnapshotID} {
				response := get(query)
				require.Equal(t, 403, response.Code, "revoked context must override applicant/approver identity: %s", response.Body.String())
				require.NotContains(t, response.Body.String(), "submission_history")
			}
			require.Equal(t, before, state())
		})
	}
}

// The approval still points to A/C1 while the editable intake selects B/C2.
// Detail and snapshot must remain usable before resubmission, and neither
// old frozen access nor current intake access can substitute for the other.
func TestBackendFinalReviewPostgresPendingClientChangeHTTP(t *testing.T) {
	for _, role := range []string{"lawyer", "partner"} {
		t.Run(role, func(t *testing.T) {
			db, a, req, intakeID, oldClient, newClient, state := services.ApprovalBackendFinalReviewPendingClientFixture(t)
			actor := uint(11)
			if role == "partner" {
				actor = 22
			}
			r := backendReviewHTTPRouter(db, actor, role)
			reads := []string{"/approvals/" + a.ID, "/approvals/" + a.ID + "/snapshot", "/approvals/" + a.ID + "/snapshot?snapshot_id=" + *a.ActiveSubmissionSnapshotID}
			before := state()
			for _, path := range reads {
				response := backendReviewHTTP(r, http.MethodGet, path, "")
				require.Equal(t, 200, response.Code, "pending client change: %s", response.Body.String())
				if path == reads[0] {
					require.Contains(t, response.Body.String(), req.ReturnRecordID)
					require.Contains(t, response.Body.String(), "supplement_requests")
				}
			}
			require.Equal(t, before, state(), "detail/snapshot reads must be read-only")
			for _, clientID := range []uint{oldClient, newClient} {
				wall := &models.Case{CaseNumber: fmt.Sprintf("PENDING-WALL-%d", clientID), Title: "补正权限样本", ClientID: clientID, LawyerID: 11, CreatedBy: "11", EthicalWallEnabled: true}
				require.NoError(t, db.Create(wall).Error)
				before = state()
				for _, path := range reads {
					response := backendReviewHTTP(r, http.MethodGet, path, "")
					require.Equal(t, 403, response.Code, response.Body.String())
				}
				if role == "lawyer" {
					payload, _ := json.Marshal(map[string]interface{}{"approval_id": a.ID, "return_record_id": req.ReturnRecordID, "expected_revision": req.ExpectedIntakeRevision, "supplement_responses": req.Responses})
					response := backendReviewHTTP(r, http.MethodPut, "/case-intakes/"+intakeID, string(payload))
					require.Equal(t, 403, response.Code, response.Body.String())
					raw, _ := json.Marshal(req)
					response = backendReviewHTTP(r, http.MethodPost, "/approvals/"+a.ID+"/resubmit", string(raw))
					require.Equal(t, 403, response.Code, response.Body.String())
				}
				require.Equal(t, before, state(), "revocation rejection must not write")
				require.NoError(t, db.Unscoped().Delete(wall).Error)
			}
			if role == "lawyer" {
				payload, _ := json.Marshal(map[string]interface{}{"approval_id": a.ID, "return_record_id": req.ReturnRecordID, "expected_revision": req.ExpectedIntakeRevision, "supplement_responses": req.Responses})
				saved := backendReviewHTTP(r, http.MethodPut, "/case-intakes/"+intakeID, string(payload))
				require.Equal(t, 200, saved.Code, saved.Body.String())
				var envelope struct {
					Data struct {
						Revision string `json:"revision"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(saved.Body.Bytes(), &envelope))
				require.NotEmpty(t, envelope.Data.Revision)
				req.ExpectedIntakeRevision = envelope.Data.Revision
				raw, _ := json.Marshal(req)
				submitted := backendReviewHTTP(r, http.MethodPost, "/approvals/"+a.ID+"/resubmit", string(raw))
				require.Equal(t, 200, submitted.Code, submitted.Body.String())
				var persisted models.ApprovalRequest
				require.NoError(t, db.First(&persisted, "id = ?", a.ID).Error)
				require.Equal(t, models.ApprovalStatusResubmitted, persisted.Status)
				require.NotEqual(t, a.ActiveSubmissionSnapshotID, persisted.ActiveSubmissionSnapshotID)
				for _, path := range reads[:2] {
					require.Equal(t, 200, backendReviewHTTP(r, http.MethodGet, path, "").Code)
				}
			}
		})
	}
}
