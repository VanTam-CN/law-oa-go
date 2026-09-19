package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

type stubConflictDetectionService struct{}

func (s *stubConflictDetectionService) PerformConflictCheck(ctx context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	return &models.ConflictCheckResponse{}, nil
}

type stubProcessApprovalService struct {
	services.ApprovalConflictIntegrationService
	err error
}

func (s *stubProcessApprovalService) ProcessApprovalWithConflict(ctx context.Context, userID, userName string, approvalID string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	return nil, s.err
}

func newProcessApprovalTestServer(t *testing.T, svc services.ApprovalConflictIntegrationService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewIntegrationHandler(svc, &stubConflictDetectionService{})
	router.POST("/api/v1/integration/approvals/:id/decision", func(c *gin.Context) {
		c.Set("user_id", "4")
		c.Set("username", "管理合伙人")
		handler.ProcessApprovalWithConflict(c)
	})
	return router
}

func postApprovalDecision(t *testing.T, router *gin.Engine, approvalID string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(map[string]interface{}{
		"decision":        "approve",
		"decision_reason": "同意并成案",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/integration/approvals/"+approvalID+"/decision", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// The decision endpoint must keep the canonical lookup contract: a missing
// approval is 404, an object-level access denial is 403, and only true
// infrastructure failures stay 500. Before the mapping the not-found branch
// surfaced as 500 with the raw wrapped message.
func TestProcessApprovalWithConflictLookupErrors(t *testing.T) {
	router := newProcessApprovalTestServer(t, &stubProcessApprovalService{err: gorm.ErrRecordNotFound})
	rec := postApprovalDecision(t, router, "missing-approval")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing approval: got %d %s, want 404", rec.Code, rec.Body.String())
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if payload.Error != "APPROVAL_NOT_FOUND" {
		t.Fatalf("missing approval error = %q, want APPROVAL_NOT_FOUND", payload.Error)
	}

	router = newProcessApprovalTestServer(t, &stubProcessApprovalService{err: services.ErrApprovalForbidden})
	rec = postApprovalDecision(t, router, "someone-elses-approval")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("forbidden approval: got %d %s, want 403", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if payload.Error != "APPROVAL_ACCESS_DENIED" {
		t.Fatalf("forbidden error = %q, want APPROVAL_ACCESS_DENIED", payload.Error)
	}
}

// A replay whose live case parties no longer match the reviewed snapshot is a
// business gate (CASE_CREATION_INCOMPLETE), not an infrastructure fault: the
// decision endpoint must publish it as 409 with the canonical code instead of
// the generic 500 wrapper.
func TestProcessApprovalWithConflictSubjectWorkflowError(t *testing.T) {
	gateErr := services.NewSubjectWorkflowError("CASE_CREATION_INCOMPLETE", "正式案件的主体身份或角色集合与批准快照不一致，已阻止假完成")
	router := newProcessApprovalTestServer(t, &stubProcessApprovalService{err: gateErr})
	rec := postApprovalDecision(t, router, "18d6a9a182d49308")
	if rec.Code != http.StatusConflict {
		t.Fatalf("case creation incomplete: got %d %s, want 409", rec.Code, rec.Body.String())
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if payload.Error != "CASE_CREATION_INCOMPLETE" {
		t.Fatalf("case creation incomplete error = %q, want CASE_CREATION_INCOMPLETE", payload.Error)
	}
}
