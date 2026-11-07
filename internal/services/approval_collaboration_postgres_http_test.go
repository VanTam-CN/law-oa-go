package services_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// Start from a reviewed intake with no approval. Unlike the decision fixture,
// this exercises the actual HTTP claim before the initial transaction/freeze.
func TestApprovalCollaborationPostgresHTTPInitialSubmission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, injectFailure := range []bool{false, true} {
		name := "first_submit_then_duplicate"
		if injectFailure {
			name = "snapshot_failure_then_original_retry"
		}
		t.Run(name, func(t *testing.T) {
			db, payload, state := services.ApprovalCollaborationPostgresInitialHTTPFixture(t)
			ordinary := handlers.NewApprovalHandler(db)
			caseService := services.NewCaseService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db))
			caseService.SetSubjectRecheckService(services.NewSubjectRecheckService(db, nil))
			integratedService := services.NewApprovalConflictIntegrationService(ordinary.GetApprovalService(), nil, caseService, repositories.NewIntegrationRepository(db), db)
			handler := handlers.NewIntegrationHandler(integratedService, nil, db)
			handler.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint(11))
				c.Set("role", "lawyer")
				c.Set("username", "虚构申请律师")
				c.Next()
			})
			router.POST("/integration/approvals/with-conflict", handler.CreateApprovalWithConflict)
			raw, err := json.Marshal(payload)
			require.NoError(t, err)
			post := func() *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodPost, "/integration/approvals/with-conflict", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				return recorder
			}
			var claims, faultHits atomic.Int32
			const claimObserver = "collab-pg-http-real-claim"
			require.NoError(t, db.Callback().Update().After("gorm:update").Register(claimObserver, func(tx *gorm.DB) {
				if fields, ok := tx.Statement.Dest.(map[string]interface{}); ok && tx.Error == nil && tx.RowsAffected == 1 && tx.Statement.Table == "case_intakes" && fields["status"] == "approval_submitting" {
					claims.Add(1)
				}
			}))
			t.Cleanup(func() { require.NoError(t, db.Callback().Update().Remove(claimObserver)) })
			intakeState := func() string {
				var value string
				// Claim acquisition/release legitimately refreshes updated_at;
				// compare every other intake column, including status and token.
				require.NoError(t, db.Raw("SELECT (to_jsonb(i) - 'updated_at')::text FROM case_intakes i WHERE id = ?", payload.Metadata["intake_id"]).Row().Scan(&value))
				return value
			}
			initialIntake := intakeState()
			if injectFailure {
				const faultName = "collab-pg-http-fail-initial-freeze"
				require.NoError(t, db.Callback().Create().Before("gorm:create").Register(faultName, func(tx *gorm.DB) {
					if tx.Statement.Table != "approval_snapshots" {
						return
					}
					if _, inTx := tx.Statement.ConnPool.(gorm.TxCommitter); !inTx {
						tx.AddError(errors.New("snapshot write outside transaction"))
						return
					}
					faultHits.Add(1)
					tx.AddError(errors.New("injected-http-initial-snapshot-failure"))
				}))
				t.Cleanup(func() { _ = db.Callback().Create().Remove(faultName) })
				before := state()
				failed := post()
				require.GreaterOrEqual(t, failed.Code, 400, failed.Body.String())
				require.Positive(t, faultHits.Load(), "must reach actual freeze write, not fail an earlier gate: %s", failed.Body.String())
				require.Positive(t, claims.Load(), "handler must acquire a real database claim")
				after := state()
				delete(before, "case_intakes")
				delete(after, "case_intakes")
				require.Equal(t, before, after, "failed HTTP submit leaves no approval/snapshot/slot/inbox artifacts")
				require.Equal(t, initialIntake, intakeState(), "failed request restores the intake claim and preserves all facts")
				require.NoError(t, db.Callback().Create().Remove(faultName))
			}
			created := post()
			require.Equal(t, http.StatusOK, created.Code, created.Body.String())
			require.Positive(t, claims.Load(), "HTTP flow must not bypass approval_submitting claim")
			require.Equal(t, initialIntake, intakeState(), "successful request releases the matching claim")
			var approval models.ApprovalRequest
			require.NoError(t, db.First(&approval).Error)
			require.Equal(t, models.ApprovalStatusSubmitted, approval.Status)
			require.NotNil(t, approval.ActiveSubmissionSnapshotID)
			for table, expected := range map[string]int64{"approval_requests": 1, "approval_snapshots": 1, "approval_case_creation_tracking": 1, "approval_records": 0, "cases": 0} {
				var count int64
				require.NoError(t, db.Table(table).Count(&count).Error)
				require.Equal(t, expected, count, table)
			}
			var snapshot models.ApprovalSnapshot
			require.NoError(t, db.First(&snapshot, "id = ?", *approval.ActiveSubmissionSnapshotID).Error)
			require.Equal(t, approval.ID, snapshot.ApprovalRequestID)
			require.Equal(t, 1, snapshot.SourceVersion)
			var tasks int64
			require.NoError(t, db.Model(&models.InboxItem{}).Where("workflow_purpose = 'decide_approval' AND is_completed = false").Count(&tasks).Error)
			require.EqualValues(t, 1, tasks)
			beforeDuplicate := state()
			duplicate := post()
			require.Equal(t, http.StatusConflict, duplicate.Code, duplicate.Body.String())
			require.Contains(t, duplicate.Body.String(), "INTAKE_APPROVAL_ALREADY_IN_PROGRESS")
			require.Equal(t, beforeDuplicate, state(), "duplicate HTTP submit must do no writes")
			t.Logf("initial HTTP actual_claims=%d snapshot_fault_hits=%d approval=1 snapshot=1 slot=1 decision_task=1 duplicate=409 no_writes=true", claims.Load(), faultHits.Load())
		})
	}
}

// Exercise actual HTTP error mapping, not a test router that synthesizes 409.
func TestApprovalCollaborationPostgresHTTPStaleDecision(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range []string{"ordinary", "integrated"} {
		for _, stale := range []string{"snapshot", "node"} {
			t.Run(entry+"_"+stale, func(t *testing.T) {
				db, old, current, state := services.ApprovalCollaborationPostgresHTTPFixture(t)
				auth := services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil)
				ordinary := handlers.NewApprovalHandler(db)
				ordinary.SetAuthorizationService(auth)
				caseService := services.NewCaseService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db))
				caseService.SetSubjectRecheckService(services.NewSubjectRecheckService(db, nil))
				integratedService := services.NewApprovalConflictIntegrationService(ordinary.GetApprovalService(), nil, caseService, repositories.NewIntegrationRepository(db), db)
				integrated := handlers.NewIntegrationHandler(integratedService, nil, db)
				integrated.SetAuthorizationService(auth)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					c.Set("user_id", uint(22))
					c.Set("role", "partner")
					c.Set("username", "虚构审批合伙人")
					c.Next()
				})
				if entry == "ordinary" {
					router.POST("/approvals/:id/decision", ordinary.ProcessApprovalDecision)
				} else {
					router.POST("/approvals/:id/decision", integrated.ProcessApprovalWithConflict)
				}
				payload := models.ApprovalDecisionRequest{Decision: models.ApprovalDecisionApprove, DecisionReason: "旧页面迟到审批", ExpectedSubmissionSnapshotID: *current.ActiveSubmissionSnapshotID, ExpectedActionVersion: current.ActionVersion}
				code := "APPROVAL_SUBMISSION_VERSION_CONFLICT"
				if stale == "snapshot" {
					payload.ExpectedSubmissionSnapshotID = *old.ActiveSubmissionSnapshotID
				} else {
					payload.ExpectedActionVersion = old.ActionVersion
					code = "APPROVAL_ACTION_VERSION_CONFLICT"
				}
				raw, err := json.Marshal(payload)
				require.NoError(t, err)
				before := state()
				req := httptest.NewRequest(http.MethodPost, "/approvals/"+current.ID+"/decision", bytes.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				require.Equal(t, http.StatusConflict, recorder.Code, recorder.Body.String())
				require.Contains(t, recorder.Body.String(), code)
				require.Equal(t, before, state(), "HTTP 409 must leave every business row unchanged")
				t.Logf("C11 actual HTTP entry=%s stale=%s status=409 code=%s writes=0", entry, stale, code)
			})
		}
	}
}
