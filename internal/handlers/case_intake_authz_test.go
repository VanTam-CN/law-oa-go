package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCaseIntakeAuthzTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/intake-authz.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE case_intakes (
		id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT, description TEXT,
		priority TEXT, status TEXT, metadata TEXT, created_by TEXT, idempotency_key TEXT,
		created_at DATETIME, updated_at DATETIME
	)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func caseIntakeAuthzRequest(t *testing.T, method, path, role string, userID uint, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", userID)
	c.Set("role", role)
	return c, recorder
}

func TestCreateCaseIntakeRejectsUnauthorizedRoleWithoutWriting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newCaseIntakeAuthzTestDB(t)
	handler := NewDemoAggregateHandler(db)

	for _, role := range []string{"user", "intern", "finance", "admin", "super_admin", "conflict_officer", "management"} {
		c, recorder := caseIntakeAuthzRequest(t, http.MethodPost, "/api/v1/case-intakes", role, 41,
			`{"title":"越权接案","parties":[{"name":"受保护主体"}]}`)
		handler.CreateCaseIntake(c)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected role %s to receive 403, got %d: %s", role, recorder.Code, recorder.Body.String())
		}
	}
	var count int64
	if err := db.Table("case_intakes").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("unauthorized intake create wrote %d rows", count)
	}
}

func TestCreateCaseIntakeAuthorizationPrecedesParsingAndIdempotency(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newCaseIntakeAuthzTestDB(t)
	if err := db.Exec(`INSERT INTO case_intakes (id, intake_code, title, status, created_by, idempotency_key)
		VALUES ('existing-intake', 'INT-EXISTING', '已有接案', 'draft', '41', 'same-key')`).Error; err != nil {
		t.Fatal(err)
	}
	handler := NewDemoAggregateHandler(db)
	c, recorder := caseIntakeAuthzRequest(t, http.MethodPost, "/api/v1/case-intakes", "user", 41, "not-json")
	c.Request.Header.Set("Idempotency-Key", "same-key")
	handler.CreateCaseIntake(c)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected authorization to precede malformed JSON and idempotency lookup, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if bytes.Contains(recorder.Body.Bytes(), []byte("existing-intake")) {
		t.Fatalf("unauthorized request re-exposed an existing intake: %s", recorder.Body.String())
	}
}

func TestCreateCaseIntakeAllowsLawyerAndKeepsAssistantDraftBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newCaseIntakeAuthzTestDB(t)
	handler := NewDemoAggregateHandler(db)

	lawyerContext, lawyerRecorder := caseIntakeAuthzRequest(t, http.MethodPost, "/api/v1/case-intakes", "lawyer", 7,
		`{"title":"律师接案","case_type":"commercial"}`)
	handler.CreateCaseIntake(lawyerContext)
	if lawyerRecorder.Code != http.StatusCreated {
		t.Fatalf("expected lawyer create to succeed, got %d: %s", lawyerRecorder.Code, lawyerRecorder.Body.String())
	}

	assistantContext, assistantRecorder := caseIntakeAuthzRequest(t, http.MethodPost, "/api/v1/case-intakes", "assistant", 8,
		`{"title":"助理草稿","client_id":99}`)
	handler.CreateCaseIntake(assistantContext)
	if assistantRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected assistant identity boundary to remain enforced, got %d: %s", assistantRecorder.Code, assistantRecorder.Body.String())
	}

	assistantContext, assistantRecorder = caseIntakeAuthzRequest(t, http.MethodPost, "/api/v1/case-intakes", "intake_assistant", 8,
		`{"title":"助理草稿","description":"待律师确认"}`)
	handler.CreateCaseIntake(assistantContext)
	if assistantRecorder.Code != http.StatusCreated {
		t.Fatalf("expected safe assistant draft to succeed, got %d: %s", assistantRecorder.Code, assistantRecorder.Body.String())
	}
	var statuses []string
	if err := db.Table("case_intakes").Order("created_at").Pluck("status", &statuses).Error; err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 2 || statuses[1] != "assistant_draft" {
		t.Fatalf("unexpected intake statuses: %v", statuses)
	}
}

func TestIntakeWorkbenchRejectsCrossUserRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newCaseIntakeAuthzTestDB(t)
	if err := db.Exec(`INSERT INTO case_intakes (id, intake_code, title, status, created_by)
		VALUES ('intake-owned', 'INT-OWNED', '仅所有者可见', 'draft', '7')`).Error; err != nil {
		t.Fatal(err)
	}
	handler := NewDemoAggregateHandler(db)
	c, recorder := caseIntakeAuthzRequest(t, http.MethodGet, "/api/v1/cases/intake-workbench/intake-owned", "lawyer", 8, "")
	c.Params = gin.Params{{Key: "id", Value: "intake-owned"}}
	handler.IntakeWorkbench(c)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected cross-user intake read to receive 403, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var response map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["data"] != nil {
		t.Fatalf("cross-user response unexpectedly exposed data: %s", recorder.Body.String())
	}
}

func TestIntakeWorkbenchRejectsHistoricalUserOwnerAndOmitsTeam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newCaseIntakeAuthzTestDB(t)
	if err := db.Exec(`INSERT INTO case_intakes (id, intake_code, title, status, created_by)
		VALUES ('legacy-user-intake', 'INT-LEGACY', '历史草稿', 'draft', '41')`).Error; err != nil {
		t.Fatal(err)
	}
	handler := NewDemoAggregateHandler(db)
	userContext, userRecorder := caseIntakeAuthzRequest(t, http.MethodGet, "/api/v1/cases/intake-workbench/legacy-user-intake", "user", 41, "")
	userContext.Params = gin.Params{{Key: "id", Value: "legacy-user-intake"}}
	handler.IntakeWorkbench(userContext)
	if userRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected historical user owner to receive 403, got %d: %s", userRecorder.Code, userRecorder.Body.String())
	}

	lawyerContext, lawyerRecorder := caseIntakeAuthzRequest(t, http.MethodGet, "/api/v1/cases/intake-workbench/legacy-user-intake", "lawyer", 41, "")
	lawyerContext.Params = gin.Params{{Key: "id", Value: "legacy-user-intake"}}
	handler.IntakeWorkbench(lawyerContext)
	if lawyerRecorder.Code != http.StatusOK {
		t.Fatalf("expected lawyer owner to receive 200, got %d: %s", lawyerRecorder.Code, lawyerRecorder.Body.String())
	}
	var response map[string]interface{}
	if err := json.Unmarshal(lawyerRecorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if _, exists := response["team"]; exists {
		t.Fatalf("workbench unexpectedly returned team data: %s", lawyerRecorder.Body.String())
	}
}

func TestUpdateCaseIntakeRejectsUnauthorizedBeforeParsingOrLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newCaseIntakeAuthzTestDB(t)
	if err := db.Exec(`INSERT INTO case_intakes (id, intake_code, title, status, created_by)
		VALUES ('legacy-update', 'INT-UPDATE', '原始标题', 'draft', '41')`).Error; err != nil {
		t.Fatal(err)
	}
	handler := NewDemoAggregateHandler(db)
	c, recorder := caseIntakeAuthzRequest(t, http.MethodPut, "/api/v1/case-intakes/legacy-update", "user", 41, "not-json")
	c.Params = gin.Params{{Key: "id", Value: "legacy-update"}}
	handler.UpdateCaseIntake(c)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected unauthorized update to receive 403 before parsing, got %d: %s", recorder.Code, recorder.Body.String())
	}
	var title string
	if err := db.Table("case_intakes").Where("id = ?", "legacy-update").Pluck("title", &title).Error; err != nil {
		t.Fatal(err)
	}
	if title != "原始标题" {
		t.Fatalf("unauthorized update changed the historical intake: %q", title)
	}
}
