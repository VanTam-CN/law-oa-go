package services_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// Both real HTTP decision routes must advance an intermediate node without
// attempting case creation. Case writes are observed/faulted, never replaced.
func TestApprovalMultistagePostgresHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range []string{"ordinary", "integrated"} {
		for _, scenario := range []struct{ fault, withConfig bool }{{}, {fault: true}, {withConfig: true}} {
			fault := scenario.fault
			name := entry + "/complete"
			if fault {
				name = entry + "/final_case_failure_retry"
			}
			if scenario.withConfig {
				name = entry + "/config_direct_two_stages"
			}
			t.Run(name, func(t *testing.T) {
				db := ApprovalBootstrapPostgresDatabase(t)
				payload, state := services.ApprovalBootstrapPostgresHTTPFixture(t, db)
				var configIntake struct {
					Title, Description, CaseType, Priority string
					ClientID                               uint
				}
				if scenario.withConfig {
					require.NoError(t, db.Table("case_intakes").Where("id = ?", payload.Metadata["intake_id"]).Take(&configIntake).Error)
					// Match the browser's top-level and metadata config, retaining the
					// real initial submission path that creates the NULL-intake row.
					payload.CaseCreationConfig = map[string]interface{}{"title": configIntake.Title, "description": configIntake.Description,
						"client_id": configIntake.ClientID, "case_type": configIntake.CaseType, "priority": configIntake.Priority,
						"lawyer_id": 11, "billing_method": "hourly", "is_major_risk": false}
					payload.Metadata["case_creation_config"] = payload.CaseCreationConfig
				}
				require.NoError(t, db.Create(&models.User{ID: 44, Username: "pg-compliance", Name: "虚构合规审批人", Email: "compliance@example.invalid", Password: "fixture-only", Role: "conflict_officer", Status: "active"}).Error)
				require.NoError(t, db.Model(&models.ApprovalWorkflow{}).Where("workflow_type = ?", "COLLAB_PG").Update("stages", `[{"stage_name":"partner_review","stage_order":1,"approver_id":"22","required":true},{"stage_name":"compliance","stage_order":2,"approver_id":"44","required":true}]`).Error)
				ordinary := handlers.NewApprovalHandler(db)
				caseService := services.NewCaseService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db))
				caseService.SetSubjectRecheckService(services.NewSubjectRecheckService(db, nil))
				integrated := services.NewApprovalConflictIntegrationService(ordinary.GetApprovalService(), nil, caseService, repositories.NewIntegrationRepository(db), db)
				creator := handlers.NewIntegrationHandler(integrated, nil, db)
				creator.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
				applicant := backendReviewHTTPRouter(db, 11, "lawyer")
				applicant.POST("/integration/approvals/with-conflict", creator.CreateApprovalWithConflict)
				applicant.GET("/integration/approvals/:id/status", creator.GetApprovalIntegrationStatus)
				raw, err := json.Marshal(payload)
				require.NoError(t, err)
				created := backendReviewHTTP(applicant, http.MethodPost, "/integration/approvals/with-conflict", string(raw))
				require.Equal(t, 200, created.Code, created.Body.String())
				var approval models.ApprovalRequest
				require.NoError(t, db.First(&approval).Error)
				id := approval.ID
				assertStatus := func(wantStatus, wantCaseID string) {
					t.Helper()
					before := state()
					response := backendReviewHTTP(applicant, http.MethodGet, "/integration/approvals/"+id+"/status", "")
					require.Equal(t, 200, response.Code, response.Body.String())
					var envelope struct {
						Data services.IntegrationStatus `json:"data"`
					}
					require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
					require.NotNil(t, envelope.Data.CaseCreation)
					require.Equal(t, wantStatus, envelope.Data.CaseCreation.Status)
					require.Equal(t, wantCaseID, envelope.Data.CaseCreation.CaseID)
					require.Equal(t, before, state(), "GetIntegrationStatus must remain read-only")
					t.Logf("GetIntegrationStatus HTTP=200 status=%s case=%s read_only=true", wantStatus, wantCaseID)
				}
				var initialRows []repositories.CaseCreationTracking
				var canonicalID, genericID string
				expectedTrackingCount := int64(1)
				if scenario.withConfig {
					expectedTrackingCount = 2
					require.NoError(t, db.Where("approval_request_id = ?", id).Order("created_at ASC").Find(&initialRows).Error)
					require.Len(t, initialRows, 2)
					require.NotNil(t, initialRows[0].IntakeID)
					require.Equal(t, payload.Metadata["intake_id"], *initialRows[0].IntakeID)
					require.Nil(t, initialRows[1].IntakeID)
					require.True(t, initialRows[0].CreatedAt.Before(initialRows[1].CreatedAt), "canonical is older on direct initial submission")
					canonicalID, genericID = initialRows[0].ID, initialRows[1].ID
					require.Equal(t, "INTAKE_BIND_"+id, canonicalID)
					require.Contains(t, genericID, "CASE_TRACK_"+id+"_")
					for _, row := range initialRows {
						require.Equal(t, "pending", row.CreationStatus)
					}
					latest, err := repositories.NewIntegrationRepository(db).GetLatestCaseCreationTracking(t.Context(), id)
					require.NoError(t, err)
					require.NotNil(t, latest)
					require.Equal(t, genericID, latest.ID, "latest getter selects the newer generic row, not the canonical slot")
					assertStatus("pending", "")
					t.Logf("direct shape: canonical=%s older=true generic=%s NULL-intake=true tracking=2", canonicalID, genericID)
				}
				path := "/approvals/" + id + "/decision"
				if entry == "integrated" {
					path = "/integration" + path
				}
				reload := func() models.ApprovalRequest {
					t.Helper()
					var a models.ApprovalRequest
					require.NoError(t, db.First(&a, "id = ?", id).Error)
					return a
				}
				body := func() string {
					t.Helper()
					a := reload()
					require.NotNil(t, a.ActiveSubmissionSnapshotID)
					raw, err := json.Marshal(models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove, DecisionReason: "核对本节点事实", ExpectedSubmissionSnapshotID: *a.ActiveSubmissionSnapshotID, ExpectedActionVersion: a.UpdatedAt.UTC().Format(time.RFC3339Nano)})
					require.NoError(t, err)
					return string(raw)
				}
				tracking := func() repositories.CaseCreationTracking {
					t.Helper()
					var row repositories.CaseCreationTracking
					require.NoError(t, db.Where("approval_request_id = ?", id).First(&row).Error)
					return row
				}
				count := func(table string) int64 {
					t.Helper()
					var n int64
					require.NoError(t, db.Table(table).Count(&n).Error)
					return n
				}
				var caseWrites atomic.Int32
				var failCase atomic.Bool
				failCase.Store(fault)
				const callback = "multistage-case-write"
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
					if tx.Statement.Table != "cases" {
						return
					}
					caseWrites.Add(1)
					if _, ok := tx.Statement.ConnPool.(gorm.TxCommitter); !ok {
						tx.AddError(errors.New("case write outside transaction"))
						return
					}
					if failCase.Load() {
						tx.AddError(errors.New("injected-final-case-write"))
					}
				}))
				t.Cleanup(func() { _ = db.Callback().Create().Remove(callback) })
				partner := backendReviewHTTPRouter(db, 22, "partner")
				// Keep existing preflight: an unavailable owner cannot advance even
				// the first node, and cannot leave a decision or failed tracking.
				require.NoError(t, db.Model(&models.User{}).Where("id = 11").Update("status", "inactive").Error)
				before := state()
				blocked := backendReviewHTTP(partner, http.MethodPost, path, body())
				require.Equal(t, 409, blocked.Code, blocked.Body.String())
				require.Contains(t, blocked.Body.String(), "承办律师不存在或已停用")
				require.Equal(t, before, state())
				require.Zero(t, caseWrites.Load())
				require.NoError(t, db.Model(&models.User{}).Where("id = 11").Update("status", "active").Error)
				initialTracking := tracking()
				initialTrackingRows := state()["approval_case_creation_tracking"]
				require.EqualValues(t, expectedTrackingCount, count("approval_case_creation_tracking"))
				firstBody := body()
				first := backendReviewHTTP(partner, http.MethodPost, path, firstBody)
				a := reload()
				t.Logf("first node HTTP=%d approval=%s next=%s approver=%s tracking=%s tracking_rows=%d cases=%d", first.Code, a.Status, a.CurrentStage, a.CurrentApproverID, tracking().CreationStatus, count("approval_case_creation_tracking"), count("cases"))
				require.Equal(t, 200, first.Code, first.Body.String())
				require.Equal(t, models.ApprovalStatusUnderReview, a.Status)
				require.Equal(t, "compliance", a.CurrentStage)
				require.Equal(t, "44", a.CurrentApproverID)
				require.Contains(t, first.Body.String(), `"current_stage":"compliance"`)
				require.Equal(t, initialTracking, tracking(), "intermediate approval must not touch case tracking")
				require.EqualValues(t, expectedTrackingCount, count("approval_case_creation_tracking"))
				require.Equal(t, initialTrackingRows, state()["approval_case_creation_tracking"], "no extra failed NULL-intake tracking row may appear")
				if scenario.withConfig {
					assertStatus("pending", "")
				}
				require.Zero(t, caseWrites.Load())
				require.Zero(t, count("cases"))
				require.EqualValues(t, 1, count("approval_records"))
				var tasks []models.InboxItem
				require.NoError(t, db.Where("workflow_purpose = 'decide_approval' AND is_completed = false").Find(&tasks).Error)
				require.Len(t, tasks, 1)
				require.EqualValues(t, 44, tasks[0].UserID)
				before = state()
				stale := backendReviewHTTP(partner, http.MethodPost, path, firstBody)
				require.Equal(t, 409, stale.Code, stale.Body.String())
				require.Equal(t, before, state())
				compliance := backendReviewHTTPRouter(db, 44, "conflict_officer")
				final := backendReviewHTTP(compliance, http.MethodPost, path, body())
				require.EqualValues(t, 1, caseWrites.Load(), "only final approval reaches real case INSERT")
				if fault {
					failureStatus := 500
					if entry == "ordinary" {
						failureStatus = 409
					} // Existing handler failure contract.
					require.Equal(t, failureStatus, final.Code, final.Body.String())
					require.Contains(t, final.Body.String(), "injected-final-case-write")
					a = reload()
					require.Equal(t, models.ApprovalStatusApproved, a.Status)
					require.False(t, a.CaseCreated)
					require.Empty(t, a.CreatedCaseID)
					require.Equal(t, "failed", tracking().CreationStatus)
					require.Zero(t, count("cases"))
					require.Zero(t, count("case_parties"))
					require.EqualValues(t, 2, count("approval_records"))
					failCase.Store(false)
					retry := backendReviewHTTP(applicant, http.MethodPost, "/integration/approvals/"+id+"/retry", `{"retry_type":"case_creation"}`)
					require.Equal(t, 200, retry.Code, retry.Body.String())
				} else {
					require.Equal(t, 200, final.Code, final.Body.String())
				}
				a = reload()
				require.Equal(t, models.ApprovalStatusApproved, a.Status)
				require.True(t, a.CaseCreated)
				require.NotEmpty(t, a.CreatedCaseID)
				require.Equal(t, "completed", tracking().CreationStatus)
				require.EqualValues(t, 1, count("cases"))
				require.EqualValues(t, 2, count("case_parties"))
				require.EqualValues(t, 2, count("approval_records"))
				caseID := a.CreatedCaseID
				assertCompleteConfigCase := func() {
					t.Helper()
					if !scenario.withConfig {
						return
					}
					require.EqualValues(t, 2, count("approval_case_creation_tracking"))
					var canonical, generic repositories.CaseCreationTracking
					require.NoError(t, db.First(&canonical, "id = ?", canonicalID).Error)
					require.NoError(t, db.First(&generic, "id = ?", genericID).Error)
					require.Equal(t, initialRows[0], canonical, "direct approval keeps its original canonical slot intact")
					require.Nil(t, generic.IntakeID)
					require.Equal(t, "completed", generic.CreationStatus)
					require.NotNil(t, generic.CaseID)
					require.Equal(t, caseID, *generic.CaseID)
					var matter models.Case
					require.NoError(t, db.First(&matter, "id = ?", caseID).Error)
					require.Equal(t, configIntake.Title, matter.Title)
					require.Equal(t, configIntake.ClientID, matter.ClientID)
					require.EqualValues(t, 11, matter.LawyerID)
					require.Equal(t, payload.Metadata["conflict_check_id"], matter.ConflictCheckID)
					var client models.Client
					require.NoError(t, db.First(&client, configIntake.ClientID).Error)
					var opposingDigest string
					require.NoError(t, db.Table("case_intake_parties").Where("intake_id = ?", payload.Metadata["intake_id"]).Pluck("identity_number_digest", &opposingDigest).Error)
					type party struct{ Digest, PartyType string }
					var parties []party
					require.NoError(t, db.Raw(`SELECT e.identity_number_digest AS digest, p.party_type FROM case_parties p JOIN entities e ON e.id = p.entity_id WHERE p.case_id = ? AND p.deleted_at IS NULL AND e.deleted_at IS NULL ORDER BY p.party_type`, matter.ID).Scan(&parties).Error)
					require.Equal(t, []party{{client.IdentityNumberDigest, "CLIENT"}, {opposingDigest, "OPPOSING"}}, parties)
					require.NotEmpty(t, client.IdentityNumberDigest)
					require.NotEmpty(t, opposingDigest)
					assertStatus("completed", caseID)
					require.EqualValues(t, 1, count("approval_snapshots"), "direct two-stage flow never returned or resubmitted")
				}
				assertCompleteConfigCase()
				replay := backendReviewHTTP(compliance, http.MethodPost, path, body())
				require.Equal(t, 200, replay.Code, replay.Body.String())
				require.Equal(t, caseID, reload().CreatedCaseID)
				require.EqualValues(t, 1, count("cases"))
				require.EqualValues(t, 2, count("approval_records"))
				assertCompleteConfigCase()
				t.Logf("two nodes complete fault=%v records=2 cases=1 parties=2 same-case replay=true", fault)
			})
		}
	}
}
