package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

type fakeConflictDetectionService struct {
	received *models.ConflictCheckRequest
}

func (f *fakeConflictDetectionService) PerformConflictCheck(_ context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	f.received = request
	return &models.ConflictCheckResponse{
		CheckID:       "check-test",
		HasConflict:   false,
		ConflictCases: []*models.ConflictCase{},
		RiskAssessment: &models.RiskAssessment{
			OverallRisk:      "LOW",
			RiskScore:        0,
			RiskReason:       "未发现冲突",
			RequiresApproval: false,
			RiskFactors:      []string{},
			Mitigation:       []string{},
		},
		CheckStatistics: &models.CheckStatistics{},
		Recommendations: []string{},
		CheckTime:       time.Now(),
		Duration:        1,
	}, nil
}

func (f *fakeConflictDetectionService) GetCheckHistory(context.Context, string, int) ([]*models.ConflictCheckRecord, error) {
	return nil, nil
}

func (f *fakeConflictDetectionService) GetConflictStats(context.Context, string) (*repositories.ConflictStats, error) {
	return &repositories.ConflictStats{}, nil
}

func performConflictHandlerRequest(handler *ConflictHandlerSimple, role string, jwtUserID uint, requestedUserID string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/conflict/check", func(c *gin.Context) {
		c.Set("user_id", jwtUserID)
		c.Set("role", role)
		handler.CheckConflict(c)
	})

	body := map[string]interface{}{
		"clientId":     "1",
		"clientName":   "示例客户",
		"clientType":   "COMPANY",
		"otherParties": []string{"示例对方"},
		"parties": []map[string]string{
			{"role": "CLIENT", "name": "示例客户", "entityType": "COMPANY"},
			{"role": "OPPOSING_PARTY", "name": "示例对方", "entityType": "COMPANY"},
		},
		"caseName":    "示例案件",
		"caseType":    "commercial",
		"searchYears": 5,
		"searchDepth": "STANDARD",
		"userId":      requestedUserID,
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/conflict/check", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestConflictCheckRejectsForgedLawyerIDForLawyer(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)

	recorder := performConflictHandlerRequest(handler, "lawyer", 7, "9")

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received != nil {
		t.Fatal("conflict service should not be called for forged lawyer id")
	}
}

func TestConflictCheckAllowsCurrentLawyerScope(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)

	recorder := performConflictHandlerRequest(handler, "lawyer", 7, "7")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received == nil || service.received.UserID != "7" {
		t.Fatalf("expected service to receive current lawyer id 7, got %#v", service.received)
	}
	if len(service.received.Parties) != 2 {
		t.Fatalf("expected role-aware parties to be preserved, got %#v", service.received.Parties)
	}
}

func TestConflictCheckAllowsAdminToCheckOtherLawyer(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)

	recorder := performConflictHandlerRequest(handler, "admin", 1, "9")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received == nil || service.received.UserID != "9" {
		t.Fatalf("expected admin delegated lawyer id 9, got %#v", service.received)
	}
}
