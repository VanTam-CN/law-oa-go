package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newIntakeDraftListRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.GET("/api/v1/case-intakes", h.ListIntakeDrafts)
	return r
}

func newIntakeDraftListDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE case_intakes (
		id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT,
		client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT,
		created_at DATETIME, updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	seed := []struct{ id, code, title, createdBy, status, updatedAt string }{
		{"i-1", "INT-001", "张三买卖合同纠纷", "7", "draft", "2026-09-10 08:00:00"},
		{"i-2", "INT-002", "李四劳动争议", "7", "draft", "2026-09-09 08:00:00"},
		{"i-3", "INT-003", "王五侵权案", "8", "draft", "2026-09-10 09:00:00"},
		{"i-4", "INT-004", "赵六咨询", "7", "lawyer_facts_confirmed", "2026-09-10 07:00:00"},
	}
	for _, s := range seed {
		if err := db.Exec(
			`INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, status, metadata, created_at, updated_at)
			 VALUES (?, ?, ?, 'civil', 3, ?, ?, '{}', ?, ?)`,
			s.id, s.code, s.title, s.createdBy, s.status, s.updatedAt, s.updatedAt,
		).Error; err != nil {
			t.Fatalf("seed %s: %v", s.id, err)
		}
	}
	return db
}

func listIntakeDrafts(t *testing.T, r *gin.Engine, role, userID, query string) (int, gin.H) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/case-intakes"+query, nil)
	req.Header.Set("X-Test-Role", role)
	req.Header.Set("X-Test-User-Id", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload struct {
		Data gin.H `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload.Data
}

func TestListIntakeDraftsScopesAndFields(t *testing.T) {
	r := newIntakeDraftListRouter(newIntakeDraftListDB(t))

	code, data := listIntakeDrafts(t, r, "lawyer", "7", "")
	if code != http.StatusOK {
		t.Fatalf("owner list: status=%d want 200", code)
	}
	items, _ := data["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("owner list: items=%d want 2 (own drafts only, confirmed intake excluded)", len(items))
	}
	for _, raw := range items {
		row, _ := raw.(map[string]interface{})
		for _, forbidden := range []string{"metadata", "parties", "materials"} {
			if _, exists := row[forbidden]; exists {
				t.Fatalf("owner list: forbidden field %q leaked", forbidden)
			}
		}
	}

	code, data = listIntakeDrafts(t, r, "director", "9", "")
	if code != http.StatusOK {
		t.Fatalf("management list: status=%d want 200", code)
	}
	items, _ = data["items"].([]interface{})
	if len(items) != 3 {
		t.Fatalf("management list: items=%d want 3", len(items))
	}
	code, _ = listIntakeDrafts(t, r, "director", "9", "?page=2&page_size=2")
	if code != http.StatusOK {
		t.Fatalf("pagination: status=%d want 200", code)
	}
	code, _ = listIntakeDrafts(t, r, "director", "9", "?page_size=999")
	if code != http.StatusOK {
		t.Fatalf("oversized page_size: status=%d want 200", code)
	}
}
