package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

type fakeConflictDetectionService struct {
	received *models.ConflictCheckRequest
}

type fakeAsyncConflictService struct {
	ownerID       uint
	subjectCaseID string
	result        models.JSON
}

type fakeConflictReviewService struct {
	called bool
}

func (f *fakeConflictReviewService) ReviewConflict(_ context.Context, checkID, decision, notes string, reviewerID uint, reviewerName string, nextReviewAt *time.Time) (*models.ConflictReview, error) {
	f.called = true
	return &models.ConflictReview{
		CheckID: checkID, Decision: decision, Notes: notes, ReviewerID: reviewerID,
		ReviewerName: reviewerName, NextReviewAt: nextReviewAt,
	}, nil
}

func (f *fakeConflictReviewService) GetConflictReview(context.Context, string) (*models.ConflictReview, error) {
	return nil, nil
}

func (f *fakeAsyncConflictService) CreateTask(context.Context, *models.ConflictCheckRequest) (*services.ConflictCheckTaskResponse, error) {
	return nil, nil
}

func (f *fakeAsyncConflictService) GetTask(context.Context, string) (*services.ConflictCheckTaskResponse, error) {
	return &services.ConflictCheckTaskResponse{TaskID: "task-1", Status: services.ConflictTaskStatusCompleted, OwnerID: f.ownerID, SubjectCaseID: f.subjectCaseID}, nil
}

func (f *fakeAsyncConflictService) GetTaskResult(context.Context, string) (*services.ConflictCheckTaskResultResponse, error) {
	return &services.ConflictCheckTaskResultResponse{Task: &services.ConflictCheckTaskResponse{TaskID: "task-1", Status: services.ConflictTaskStatusCompleted, OwnerID: f.ownerID, SubjectCaseID: f.subjectCaseID}, Result: f.result}, nil
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
	return performConflictHandlerRequestWithSubject(handler, role, jwtUserID, requestedUserID, "")
}

func performConflictHandlerRequestWithSubject(handler *ConflictHandlerSimple, role string, jwtUserID uint, requestedUserID, subjectCaseID string) *httptest.ResponseRecorder {
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
	if subjectCaseID != "" {
		body["subjectCaseId"] = subjectCaseID
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/conflict/check", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func newConflictContextAuthorization(t *testing.T) (*services.AuthorizationService, *gorm.DB) {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("c", 32))
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "conflict-context.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{}))
	require.NoError(t, db.Create(&models.Client{ID: 1, Name: "演练客户", Type: "企业", IdentityType: models.IdentityTypeSocialCredit, IdentityNumber: "91310000TESTCLIENT1", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.Case{
		ID: 99, CaseNumber: "CASE-CONTEXT-99", Title: "演练案件", ClientID: 1, LawyerID: 7,
		CaseType: "commercial", Status: "active", CreatedBy: "7",
	}).Error)
	return services.NewAuthorizationService(
		repositories.NewCaseRepository(db), repositories.NewClientRepository(db),
		repositories.NewUserRepository(db), repositories.NewDocumentRepository(db),
	), db
}

func TestConflictCheckRejectsForgedLawyerIDForLawyer(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)
	authz, db := newConflictContextAuthorization(t)
	handler.SetAuthorizationService(authz)
	handler.SetDatabase(db)

	recorder := performConflictHandlerRequestWithSubject(handler, "lawyer", 7, "9", "99")

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
	authz, db := newConflictContextAuthorization(t)
	handler.SetAuthorizationService(authz)
	handler.SetDatabase(db)

	recorder := performConflictHandlerRequestWithSubject(handler, "lawyer", 7, "7", "99")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received == nil || service.received.UserID != "7" {
		t.Fatalf("expected service to receive current lawyer id 7, got %#v", service.received)
	}
	if service.received.SearchYears != 0 {
		t.Fatalf("expected full-history search to be enforced, got %d years", service.received.SearchYears)
	}
	if service.received.ClientName != "演练客户" || service.received.CaseName != "演练案件" || service.received.CaseType != "commercial" {
		t.Fatalf("expected canonical case subject labels, got %#v", service.received)
	}
	if service.received.ClientIdentifiers["unified_social_credit_code"] != "91310000TESTCLIENT1" {
		t.Fatalf("expected enterprise client strong identity, got %#v", service.received.ClientIdentifiers)
	}
	if len(service.received.Parties) != 2 {
		t.Fatalf("expected role-aware parties to be preserved, got %#v", service.received.Parties)
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("存在受隔离记录")) {
		t.Fatalf("no-hit response must not claim an isolated hit: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("检索已完成")) {
		t.Fatalf("expected no-hit response to explain independent review requirement: %s", recorder.Body.String())
	}
}

func TestConflictCheckRejectsAssistant(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)

	recorder := performConflictHandlerRequest(handler, "assistant", 7, "7")
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected assistant to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received != nil {
		t.Fatal("conflict service should not be called for an assistant")
	}
}

func TestConflictCheckRejectsIntakeAssistantAlias(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)
	recorder := performConflictHandlerRequest(handler, "intake_assistant", 7, "7")
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected intake assistant to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received != nil {
		t.Fatal("conflict service should not be called for an intake assistant")
	}
}

func TestConflictTaskRejectsIntakeContextOutsideControlledEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service, &fakeAsyncConflictService{ownerID: 7})
	router := gin.New()
	router.POST("/conflict/tasks", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		handler.CreateConflictTask(c)
	})
	body := `{"intakeId":"intake-1","clientId":"1","clientName":"示例客户","caseName":"示例案件","caseType":"commercial"}`
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/conflict/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected generic intake conflict task to be rejected, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received != nil {
		t.Fatal("generic task endpoint must not start an intake-bound conflict check")
	}
}

func TestConflictCheckAllowsConflictOfficerToCheckOtherLawyer(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)
	authz, db := newConflictContextAuthorization(t)
	handler.SetAuthorizationService(authz)
	handler.SetDatabase(db)

	recorder := performConflictHandlerRequestWithSubject(handler, "conflict_officer", 1, "7", "99")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received == nil || service.received.UserID != "7" {
		t.Fatalf("expected case lawyer id 7, got %#v", service.received)
	}
}

func TestConflictCheckRejectsClientDifferentFromCase(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)
	authz, db := newConflictContextAuthorization(t)
	handler.SetAuthorizationService(authz)
	handler.SetDatabase(db)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/conflict/check", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		handler.CheckConflict(c)
	})
	body := `{"clientId":"2","clientName":"伪造客户","clientType":"COMPANY","caseName":"伪造案件","caseType":"commercial","userId":"7","subjectCaseId":"99"}`
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/conflict/check", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict || !bytes.Contains(recorder.Body.Bytes(), []byte("CONFLICT_CLIENT_SCOPE_MISMATCH")) {
		t.Fatalf("expected client binding mismatch, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if service.received != nil {
		t.Fatal("conflict service should not run for a mismatched case client")
	}
}

func TestConflictCheckRequiresCaseContext(t *testing.T) {
	service := &fakeConflictDetectionService{}
	handler := NewConflictHandlerSimple(service)

	recorder := performConflictHandlerRequest(handler, "lawyer", 7, "7")

	require.Equal(t, http.StatusConflict, recorder.Code)
	require.Contains(t, recorder.Body.String(), "CASE_CONTEXT_REQUIRED")
	require.Nil(t, service.received)
}

func TestLegacyConflictHistoryFailsClosedForReviewerWithoutMatterContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewConflictHandlerSimple(&fakeConflictDetectionService{})
	handler.SetAuthorizationService(&services.AuthorizationService{})
	router := gin.New()
	router.GET("/conflict/history", func(c *gin.Context) {
		c.Set("user_id", uint(8))
		c.Set("role", "conflict_officer")
		handler.GetCheckHistory(c)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/conflict/history?clientId=1", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected legacy reviewer history to fail closed, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte("CONFLICT_CLIENT_HISTORY_UNAVAILABLE")) {
		t.Fatalf("expected controlled unavailable response, got %s", recorder.Body.String())
	}
}

func TestLegacyConflictHistoryFiltersSharedClientByMatterContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "history-context.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{}))
	require.NoError(t, db.Create(&models.User{ID: 7, Username: "lawyer-a", Name: "律师 A", Email: "lawyer-a@example.test", Role: "lawyer", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.User{ID: 9, Username: "lawyer-b", Name: "律师 B", Email: "lawyer-b@example.test", Role: "lawyer", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.Client{ID: 1, Name: "共享客户", Type: "企业", Status: "active"}).Error)
	require.NoError(t, db.Create(&[]models.Case{
		{ID: 99, CaseNumber: "CASE-A-99", Title: "A 可见案件", ClientID: 1, LawyerID: 7, CaseType: "commercial", Status: "active", CreatedBy: "7"},
		{ID: 100, CaseNumber: "CASE-B-100", Title: "B 其他案件", ClientID: 1, LawyerID: 9, CaseType: "commercial", Status: "active", CreatedBy: "9"},
	}).Error)

	authz := services.NewAuthorizationService(
		repositories.NewCaseRepository(db), repositories.NewClientRepository(db),
		repositories.NewUserRepository(db), repositories.NewDocumentRepository(db),
	)
	handler := NewConflictHandlerSimple(&fakeConflictDetectionService{})
	handler.SetAuthorizationService(authz)
	handler.SetDatabase(db)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodGet, "/conflict/history?clientId=1", nil)
	ctx.Set("user_id", uint(7))
	ctx.Set("role", "lawyer")
	history := []*models.ConflictCheckRecord{
		{CheckID: "check-a", ClientID: "1", UserID: 7, SearchParameters: models.JSON{"subjectCaseId": "99"}},
		{CheckID: "check-b", ClientID: "1", UserID: 9, SearchParameters: models.JSON{"subjectCaseId": "100"}},
		{CheckID: "check-legacy", ClientID: "1", UserID: 7, SearchParameters: models.JSON{}},
	}

	visible := handler.filterConflictHistoryByContext(ctx, history)
	if len(visible) != 1 || visible[0].CheckID != "check-a" {
		t.Fatalf("expected only actor-visible matter history, got %#v", visible)
	}
}

func TestConflictTaskReadEnforcesObjectOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewConflictHandlerSimple(&fakeConflictDetectionService{})
	handler.asyncConflictService = &fakeAsyncConflictService{ownerID: 9}
	router := gin.New()
	router.GET("/conflict/tasks/:task_id", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		handler.GetConflictTaskStatus(c)
	})
	router.GET("/admin/conflict/tasks/:task_id", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("role", "director")
		handler.GetConflictTaskStatus(c)
	})
	router.GET("/conflict/tasks/:task_id/result", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		handler.GetConflictTaskResult(c)
	})

	lawyer := httptest.NewRecorder()
	router.ServeHTTP(lawyer, httptest.NewRequest(http.MethodGet, "/conflict/tasks/task-1", nil))
	if lawyer.Code != http.StatusForbidden {
		t.Fatalf("expected cross-lawyer task read to be forbidden, got %d: %s", lawyer.Code, lawyer.Body.String())
	}
	admin := httptest.NewRecorder()
	router.ServeHTTP(admin, httptest.NewRequest(http.MethodGet, "/admin/conflict/tasks/task-1", nil))
	if admin.Code != http.StatusForbidden {
		t.Fatalf("management must not read a task without verifiable case context, got %d: %s", admin.Code, admin.Body.String())
	}
	result := httptest.NewRecorder()
	router.ServeHTTP(result, httptest.NewRequest(http.MethodGet, "/conflict/tasks/task-1/result", nil))
	if result.Code != http.StatusForbidden {
		t.Fatalf("expected cross-lawyer result read to be forbidden, got %d: %s", result.Code, result.Body.String())
	}
}

func TestConflictReviewEnforcesEthicalWallBeforeWritingConclusion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "review-wall.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{}, &models.CaseEthicalWallWhitelist{}); err != nil {
		t.Fatalf("migrate review fixture: %v", err)
	}
	if err := db.Create(&[]models.User{
		{ID: 7, Username: "applicant", Name: "申请律师", Email: "applicant@example.test", Password: "fixture", Role: "lawyer", Status: "active"},
		{ID: 8, Username: "reviewer", Name: "冲突核查人", Email: "reviewer@example.test", Password: "fixture", Role: "conflict_officer", Status: "active"},
	}).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if err := db.Create(&models.Client{ID: 1, Name: "受保护客户", Type: "企业", Status: "active"}).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	if err := db.Create(&models.Case{
		ID: 99, CaseNumber: "CASE-WALL-99", Title: "受保护案件", ClientID: 1, LawyerID: 7,
		CaseType: "commercial", Status: "active", CreatedBy: "7", EthicalWallEnabled: true,
	}).Error; err != nil {
		t.Fatalf("seed case: %v", err)
	}
	authz := services.NewAuthorizationService(
		repositories.NewCaseRepository(db), repositories.NewClientRepository(db),
		repositories.NewUserRepository(db), repositories.NewDocumentRepository(db),
	)
	reviewService := &fakeConflictReviewService{}
	handler := NewConflictHandlerSimple(&fakeConflictDetectionService{}, &fakeAsyncConflictService{ownerID: 7, subjectCaseID: "99"})
	handler.authz = authz
	handler.conflictReviewService = reviewService
	router := gin.New()
	router.POST("/conflict/tasks/:task_id/review", func(c *gin.Context) {
		c.Set("user_id", uint(8))
		c.Set("role", "conflict_officer")
		c.Set("username", "reviewer")
		handler.ReviewConflict(c)
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/conflict/tasks/task-1/review", bytes.NewBufferString(`{"decision":"no_conflict","notes":"独立复核"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected ethically walled task review to be forbidden, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if reviewService.called {
		t.Fatal("review conclusion must not be written when the task context is hidden by an ethical wall")
	}
}

func TestConflictTaskResultRedactsHistoricalMatterForLawyer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	result := models.JSON{
		"checkId":  "task-1",
		"decision": map[string]interface{}{"status": "BLOCKED", "recommendation": "已确认存在冲突"},
		"conflictCases": []interface{}{map[string]interface{}{
			"restricted": true,
			"caseId":     "99", "caseNo": "SECRET-99", "caseName": "历史并购案件",
			"conflictType": "直接利益冲突", "riskLevel": "CRITICAL", "ruleCode": "DIRECT_ADVERSE_CURRENT_CLIENT",
			"description": "历史案件的案情摘要", "conflictDetails": "策略和证据",
			"evidence": []interface{}{map[string]interface{}{
				"sourceCaseId": "99", "sourceCaseNumber": "SECRET-99", "sourceCaseName": "历史并购案件",
				"summary": "历史证据摘要", "matchedEntity": "历史客户", "matchType": "EXACT",
			}},
		}},
	}
	handler := NewConflictHandlerSimple(&fakeConflictDetectionService{})
	handler.asyncConflictService = &fakeAsyncConflictService{ownerID: 7, result: result}
	router := gin.New()
	router.GET("/lawyer/conflict/tasks/:task_id/result", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		handler.GetConflictTaskResult(c)
	})
	router.GET("/officer/conflict/tasks/:task_id/result", func(c *gin.Context) {
		c.Set("user_id", uint(8))
		c.Set("role", "conflict_officer")
		handler.GetConflictTaskResult(c)
	})

	lawyer := httptest.NewRecorder()
	router.ServeHTTP(lawyer, httptest.NewRequest(http.MethodGet, "/lawyer/conflict/tasks/task-1/result", nil))
	if lawyer.Code != http.StatusOK {
		t.Fatalf("expected lawyer result response, got %d: %s", lawyer.Code, lawyer.Body.String())
	}
	for _, prohibited := range []string{"SECRET-99", "历史并购案件", "历史案件的案情摘要", "策略和证据", "历史证据摘要", "历史客户", "直接利益冲突", "CRITICAL", "DIRECT_ADVERSE_CURRENT_CLIENT"} {
		if bytes.Contains(lawyer.Body.Bytes(), []byte(prohibited)) {
			t.Fatalf("lawyer result leaked %q: %s", prohibited, lawyer.Body.String())
		}
	}
	if !bytes.Contains(lawyer.Body.Bytes(), []byte("受限历史事项")) {
		t.Fatalf("expected restricted disclosure marker: %s", lawyer.Body.String())
	}
	if !bytes.Contains(lawyer.Body.Bytes(), []byte("REVIEW_REQUIRED")) || bytes.Contains(lawyer.Body.Bytes(), []byte("已确认存在冲突")) {
		t.Fatalf("expected legacy machine conclusion to be projected as review-required: %s", lawyer.Body.String())
	}

	officer := httptest.NewRecorder()
	router.ServeHTTP(officer, httptest.NewRequest(http.MethodGet, "/officer/conflict/tasks/task-1/result", nil))
	if officer.Code != http.StatusForbidden {
		t.Fatalf("conflict officer must not see a task without verifiable case context: %d %s", officer.Code, officer.Body.String())
	}
}

func TestConflictTaskResultShowsLimitedIdentityAndTypeWithoutHistoricalCaseDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	result := models.JSON{
		"checkId": "task-visible-candidate",
		"conflictCases": []interface{}{map[string]interface{}{
			"caseId": "99", "caseNo": "SECRET-99", "caseName": "历史并购案件",
			"conflictType": "对方当事人直接冲突", "riskLevel": "CRITICAL",
			"description": "历史案件的案情摘要", "conflictDetails": "策略和证据",
			"evidence": []interface{}{map[string]interface{}{
				"sourceCaseId": "99", "sourceCaseNumber": "SECRET-99", "sourceCaseName": "历史并购案件",
				"summary": "历史证据摘要", "matchedEntity": "历史客户", "matchType": "EXACT",
			}},
		}},
	}
	handler := NewConflictHandlerSimple(&fakeConflictDetectionService{})
	handler.asyncConflictService = &fakeAsyncConflictService{ownerID: 7, result: result}
	router := gin.New()
	router.GET("/lawyer/conflict/tasks/:task_id/result", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		handler.GetConflictTaskResult(c)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/lawyer/conflict/tasks/task-visible-candidate/result", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected lawyer result response, got %d: %s", recorder.Code, recorder.Body.String())
	}
	for _, expected := range []string{"历史客户", "对方当事人直接冲突", "REVIEW_REQUIRED"} {
		if !bytes.Contains(recorder.Body.Bytes(), []byte(expected)) {
			t.Fatalf("expected limited disclosure %q: %s", expected, recorder.Body.String())
		}
	}
	for _, prohibited := range []string{"SECRET-99", "历史并购案件", "历史案件的案情摘要", "策略和证据", "历史证据摘要", "CRITICAL", "EXACT"} {
		if bytes.Contains(recorder.Body.Bytes(), []byte(prohibited)) {
			t.Fatalf("lawyer result leaked historical detail %q: %s", prohibited, recorder.Body.String())
		}
	}
}

func TestConflictTaskResultHidesRestrictedSubjectFromLawyer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	result := models.JSON{
		"checkId": "task-restricted",
		"conflictCases": []interface{}{map[string]interface{}{
			"restricted": true, "caseId": "100", "caseNo": "WALL-100", "caseName": "受伦理墙保护事项",
			"conflictType": "直接利益冲突", "evidence": []interface{}{map[string]interface{}{
				"matchedEntity": "受保护历史客户", "sourceCaseName": "受伦理墙保护事项", "summary": "受保护摘要",
			}},
		}},
	}
	handler := NewConflictHandlerSimple(&fakeConflictDetectionService{})
	handler.asyncConflictService = &fakeAsyncConflictService{ownerID: 7, result: result}
	router := gin.New()
	router.GET("/lawyer/conflict/tasks/:task_id/result", func(c *gin.Context) {
		c.Set("user_id", uint(7))
		c.Set("role", "lawyer")
		handler.GetConflictTaskResult(c)
	})
	router.GET("/officer/conflict/tasks/:task_id/result", func(c *gin.Context) {
		c.Set("user_id", uint(8))
		c.Set("role", "conflict_officer")
		handler.GetConflictTaskResult(c)
	})

	lawyer := httptest.NewRecorder()
	router.ServeHTTP(lawyer, httptest.NewRequest(http.MethodGet, "/lawyer/conflict/tasks/task-restricted/result", nil))
	if lawyer.Code != http.StatusOK {
		t.Fatalf("expected lawyer result response, got %d: %s", lawyer.Code, lawyer.Body.String())
	}
	for _, prohibited := range []string{"WALL-100", "受伦理墙保护事项", "受保护历史客户", "受保护摘要"} {
		if bytes.Contains(lawyer.Body.Bytes(), []byte(prohibited)) {
			t.Fatalf("lawyer result leaked restricted value %q: %s", prohibited, lawyer.Body.String())
		}
	}
	if !bytes.Contains(lawyer.Body.Bytes(), []byte("受限历史记录")) {
		t.Fatalf("expected restricted history marker: %s", lawyer.Body.String())
	}

	officer := httptest.NewRecorder()
	router.ServeHTTP(officer, httptest.NewRequest(http.MethodGet, "/officer/conflict/tasks/task-restricted/result", nil))
	if officer.Code != http.StatusForbidden {
		t.Fatalf("conflict officer must not see a task without verifiable case context: %d %s", officer.Code, officer.Body.String())
	}
}

func TestSanitizeCSVCellNeutralizesFormulaPrefixes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "equals", in: "=HYPERLINK(\"http://example.test\")", want: "'=HYPERLINK(\"http://example.test\")"},
		{name: "plus after whitespace", in: " \t+cmd", want: "' \t+cmd"},
		{name: "minus", in: "-10", want: "'-10"},
		{name: "at", in: "@SUM(A1:A2)", want: "'@SUM(A1:A2)"},
		{name: "plain text", in: "普通客户", want: "普通客户"},
		{name: "empty", in: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeCSVCell(tt.in); got != tt.want {
				t.Fatalf("sanitizeCSVCell(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func newTrackedConflictApprovalDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "blocked-approval.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	statements := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY, name TEXT, username TEXT, department TEXT,
			role TEXT, status TEXT, deleted_at DATETIME
		)`,
		`CREATE TABLE conflict_check_records (
			check_id TEXT PRIMARY KEY, user_id INTEGER, case_name TEXT, client_id TEXT,
			client_name TEXT, case_type TEXT, check_status TEXT, risk_level TEXT,
			has_conflict BOOLEAN, check_result TEXT, search_parameters TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE approval_requests (
			id TEXT PRIMARY KEY, request_number TEXT UNIQUE, title TEXT, type TEXT, category TEXT,
			content TEXT, applicant_id TEXT, applicant_name TEXT, applicant_title TEXT,
			department_id TEXT, department_name TEXT, urgency TEXT, priority TEXT, status TEXT,
			submission_date DATETIME, current_stage TEXT, current_approver_id TEXT,
			current_approver_name TEXT, workflow_type TEXT, workflow_config TEXT,
			attachments TEXT, metadata TEXT, created_by TEXT, created_at DATETIME, updated_at DATETIME,
			deleted_at DATETIME, conflict_check_id TEXT, conflict_risk_level TEXT,
			conflict_check_time DATETIME, conflict_result TEXT
		)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create fixture table: %v", err)
		}
	}
	if err := db.Exec(`INSERT INTO users (id, name, username, department, role, status) VALUES
		(34, '当前律师', 'lawyer-a', '争议解决部', 'lawyer', 'active'),
		(33, '合规负责人', 'compliance-a', '合规部', 'compliance', 'active')`).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	return db
}

func requestTrackedConflictApproval(t *testing.T, db *gorm.DB, checkID string) (int, map[string]interface{}) {
	t.Helper()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/conflict/tasks/"+checkID+"/approval", bytes.NewBufferString(`{}`))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Params = gin.Params{{Key: "task_id", Value: checkID}}
	context.Set("user_id", uint(34))
	context.Set("role", "lawyer")
	context.Set("username", "lawyer-a")
	NewDemoAggregateHandler(db).CreateConflictApproval(context)
	var body map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode approval response: %v; body=%s", err, recorder.Body.String())
	}
	return recorder.Code, body
}

func TestBlockedConflictApprovalLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		existingStatus string
		wantStatus     int
		wantCode       string
		wantReused     bool
	}{
		{name: "blocked without approval creates first independent review", wantStatus: http.StatusOK},
		{name: "blocked with terminal approval is rejected", existingStatus: "approved", wantStatus: http.StatusConflict, wantCode: "CONFLICT_APPROVAL_FINAL"},
		{name: "blocked with active approval is reused", existingStatus: "under_review", wantStatus: http.StatusOK, wantReused: true},
	}

	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := newTrackedConflictApprovalDB(t)
			checkID := fmt.Sprintf("BLOCKED-%d", index)
			checkResult := `{"decision":{"status":"BLOCKED","requiresManualReview":true,"evidenceCount":1},"riskAssessment":{"overallRisk":"CRITICAL","requiresApproval":true}}`
			if err := db.Exec(`INSERT INTO conflict_check_records
				(check_id, user_id, case_name, client_id, client_name, case_type, check_status,
				 risk_level, has_conflict, check_result, search_parameters, created_at, updated_at)
				VALUES (?, 34, '待复核案件', 'CLIENT-1', '委托方甲', 'commercial', 'COMPLETED',
				 'CRITICAL', 1, ?, '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, checkID, checkResult).Error; err != nil {
				t.Fatalf("seed blocked conflict: %v", err)
			}
			if test.existingStatus != "" {
				if err := db.Exec(`INSERT INTO approval_requests
					(id, request_number, type, applicant_id, current_approver_id, current_approver_name,
					 status, conflict_check_id, metadata, created_at)
					VALUES ('APP-EXISTING', 'APR-EXISTING', 'conflict_approval', '34', '33',
					 '合规负责人', ?, ?, '{}', CURRENT_TIMESTAMP)`, test.existingStatus, checkID).Error; err != nil {
					t.Fatalf("seed existing approval: %v", err)
				}
			}

			status, body := requestTrackedConflictApproval(t, db, checkID)
			if status != test.wantStatus {
				t.Fatalf("status=%d, want %d; body=%v", status, test.wantStatus, body)
			}
			if test.wantCode != "" {
				errorBody, _ := body["error"].(map[string]interface{})
				if code := fmt.Sprint(errorBody["code"]); code != test.wantCode {
					t.Fatalf("error code=%q, want %q; body=%v", code, test.wantCode, body)
				}
			}
			data, _ := body["data"].(map[string]interface{})
			if reused, _ := data["reused"].(bool); reused != test.wantReused {
				t.Fatalf("reused=%v, want %v; body=%v", reused, test.wantReused, body)
			}
			var count int64
			if err := db.Table("approval_requests").Where("conflict_check_id = ?", checkID).Count(&count).Error; err != nil {
				t.Fatalf("count approvals: %v", err)
			}
			if count != 1 {
				t.Fatalf("approval rows=%d, want 1", count)
			}
		})
	}
}

func TestConflictApprovalDoesNotTreatLimitedCoverageAsClear(t *testing.T) {
	row := map[string]interface{}{"risk_level": "LOW", "has_conflict": false}
	checkResult := map[string]interface{}{}
	decision := map[string]interface{}{"status": "CLEAR", "coverageStatus": "COVERAGE_LIMITED"}
	if conflictApprovalNotRequired(row, checkResult, decision, 0) {
		t.Fatal("limited archive coverage must not bypass independent conflict approval")
	}
	decision["coverageStatus"] = "COMPLETE"
	if !conflictApprovalNotRequired(row, checkResult, decision, 0) {
		t.Fatal("complete low-risk clear result should retain the no-approval fast path")
	}
}
