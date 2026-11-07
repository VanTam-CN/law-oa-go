package services_test

import (
	"encoding/json"
	"net/http"
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

func TestApprovalBootstrapPostgresFirstHTTPSubmission(t *testing.T) {
	assertApprovalBootstrapFirstHTTPSubmission(t, ApprovalBootstrapPostgresDatabase(t))
}

func assertApprovalBootstrapFirstHTTPSubmission(t *testing.T, db *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	payload, state := services.ApprovalBootstrapPostgresHTTPFixture(t, db)
	ordinary := handlers.NewApprovalHandler(db)
	caseService := services.NewCaseService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db))
	caseService.SetSubjectRecheckService(services.NewSubjectRecheckService(db, nil))
	integrated := services.NewApprovalConflictIntegrationService(ordinary.GetApprovalService(), nil, caseService, repositories.NewIntegrationRepository(db), db)
	h := handlers.NewIntegrationHandler(integrated, nil, db)
	h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), nil))
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(11))
		c.Set("role", "lawyer")
		c.Set("username", "虚构申请律师")
		c.Next()
	})
	r.POST("/integration/approvals/with-conflict", h.CreateApprovalWithConflict)
	var claims atomic.Int32
	const observer = "bootstrap-observe-real-claim"
	require.NoError(t, db.Callback().Update().After("gorm:update").Register(observer, func(tx *gorm.DB) {
		if fields, ok := tx.Statement.Dest.(map[string]interface{}); ok && tx.Error == nil && tx.RowsAffected == 1 && tx.Statement.Table == "case_intakes" && fields["status"] == "approval_submitting" {
			claims.Add(1)
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Update().Remove(observer) })
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	created := backendReviewHTTP(r, http.MethodPost, "/integration/approvals/with-conflict", string(raw))
	require.Equal(t, 200, created.Code, "fresh production bootstrap first POST: %s", created.Body.String())
	require.EqualValues(t, 1, claims.Load(), "must reach the real database claim")
	var approval models.ApprovalRequest
	require.NoError(t, db.First(&approval).Error)
	require.Equal(t, models.ApprovalStatusSubmitted, approval.Status)
	require.NotNil(t, approval.ActiveSubmissionSnapshotID)
	detail := backendReviewHTTP(backendReviewHTTPRouter(db, 22, "partner"), http.MethodGet, "/approvals/"+approval.ID, "")
	require.Equal(t, 200, detail.Code, detail.Body.String())
	require.Contains(t, detail.Body.String(), `"submission_snapshot_required":true`)
	for table, expected := range map[string]int64{"approval_requests": 1, "approval_snapshots": 1, "approval_case_creation_tracking": 1, "approval_records": 0, "cases": 0} {
		var count int64
		require.NoError(t, db.Table(table).Count(&count).Error)
		require.Equal(t, expected, count, table)
	}
	var taskCount int64
	require.NoError(t, db.Model(&models.InboxItem{}).Where("workflow_purpose = 'decide_approval' AND is_completed = false").Count(&taskCount).Error)
	require.EqualValues(t, 1, taskCount)
	var intake struct {
		Status             string
		ApprovalClaimToken *string
	}
	require.NoError(t, db.Table("case_intakes").Where("id = ?", payload.Metadata["intake_id"]).Take(&intake).Error)
	require.Equal(t, "conflict_ready", intake.Status)
	require.True(t, intake.ApprovalClaimToken == nil || *intake.ApprovalClaimToken == "")
	before := state()
	duplicate := backendReviewHTTP(r, http.MethodPost, "/integration/approvals/with-conflict", string(raw))
	require.Equal(t, 409, duplicate.Code, duplicate.Body.String())
	require.Equal(t, before, state(), "duplicate must preserve all business tables")
	t.Log("fresh bootstrap -> real HTTP claim -> frozen snapshot + unique slot + decision task; duplicate=409; cases=0")
}
