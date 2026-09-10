package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newIntakeFactsTestContext(t *testing.T, intakeID string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/case-intakes/"+intakeID+"/facts-confirmation", nil)
	c.Params = gin.Params{{Key: "id", Value: intakeID}}
	c.Set("user_id", uint(39))
	c.Set("role", "lawyer")
	return c, w
}

func newIntakeFactsTestHandler(t *testing.T) *DemoAggregateHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, title TEXT, case_type TEXT, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT)").Error; err != nil {
		t.Fatalf("create case_intakes: %v", err)
	}
	seed := func(id, status string) {
		t.Helper()
		if err := db.Exec("INSERT INTO case_intakes (id, title, case_type, client_id, created_by, status, metadata) VALUES (?, '测试接案', 'civil', NULL, '39', ?, '{}')", id, status).Error; err != nil {
			t.Fatalf("seed intake %s: %v", id, err)
		}
	}
	seed("intake-draft", "draft")
	seed("intake-confirmed", "lawyer_facts_confirmed")
	return NewDemoAggregateHandler(db)
}

// A draft without a client must fail with a readable validation error.
// It used to return an empty HTTP 200 because the NULL client_id short-
// circuited past authorizeIntakeClient without writing any response.
func TestConfirmIntakeFactsRejectsMissingClientWithReadableError(t *testing.T) {
	handler := newIntakeFactsTestHandler(t)
	c, w := newIntakeFactsTestContext(t, "intake-draft")

	handler.ConfirmIntakeFacts(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", w.Code, w.Body.String())
	}
	var payload struct {
		Success bool `json:"success"`
		Error   struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Success {
		t.Fatalf("success = true, want false")
	}
	if payload.Error.Code == "" || !strings.Contains(payload.Error.Message, "客户") {
		t.Fatalf("readable client error expected; body = %s", w.Body.String())
	}
}

// The conflict-check entry point must also reject a client-less intake
// through the same authorization boundary instead of proceeding silently.
func TestStartIntakeConflictCheckRejectsMissingClientWithReadableError(t *testing.T) {
	handler := newIntakeFactsTestHandler(t)
	c, w := newIntakeFactsTestContext(t, "intake-confirmed")

	handler.StartIntakeConflictCheck(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) == "" {
		t.Fatalf("empty response body; a validation failure must stay readable")
	}
}
