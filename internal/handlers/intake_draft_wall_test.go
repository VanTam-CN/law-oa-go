package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// newIntakeWallDB seeds one lawyer-owned intake bound to client 33. Client 33
// already has a walled case, so every management-role read of the intake row
// must pass the ethical-wall boundary before any payload is returned.
func newIntakeWallDB(t *testing.T) (*DemoAggregateHandler, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&models.Client{}); err != nil {
		t.Fatalf("migrate client: %v", err)
	}
	ddl := []string{
		"CREATE TABLE case_intakes (id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT, created_at DATETIME, updated_at DATETIME)",
		"CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, metadata TEXT, created_at DATETIME)",
		"CREATE TABLE case_materials (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, name TEXT, metadata TEXT, created_at DATETIME)",
		"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, role TEXT, status TEXT, department TEXT, deleted_at DATETIME)",
		"CREATE TABLE cases (id INTEGER PRIMARY KEY AUTOINCREMENT, client_id INTEGER, lawyer_id TEXT, created_by TEXT, deleted_at DATETIME, ethical_wall_enabled BOOLEAN DEFAULT FALSE)",
		"CREATE TABLE case_ethical_wall_whitelist (id INTEGER PRIMARY KEY AUTOINCREMENT, case_id INTEGER, user_id TEXT)",
		"CREATE TABLE approval_requests (id TEXT PRIMARY KEY, type TEXT, status TEXT, metadata TEXT, deleted_at DATETIME)",
		"CREATE TABLE approval_case_creation_tracking (id TEXT PRIMARY KEY, approval_request_id TEXT, intake_id TEXT, creation_status TEXT)",
	}
	for _, stmt := range ddl {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	wallClient := models.Client{Name: "星河智联科技有限公司", Type: "企业", ContactPerson: "陈律师", CreatedBy: 7}
	wallClient.ID = 33
	if err := db.Create(&wallClient).Error; err != nil {
		t.Fatalf("seed client: %v", err)
	}
	seeds := []string{
		"INSERT INTO users (id, name, role, status, department) VALUES (7, '律师甲', 'lawyer', 'active', '争议解决')",
		"INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, status, metadata, updated_at) VALUES ('intake-wall', 'INT-901', '隔离墙草稿', 'civil', 33, '7', 'draft', '{\"draft_revision\":\"rev-wall\"}', '2026-09-10 08:00:00')",
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role) VALUES ('intake-wall', '星河智联科技有限公司', 'LEGAL_PERSON', 'client')",
		"INSERT INTO case_materials (intake_id, name) VALUES ('intake-wall', '委托材料.pdf')",
		"INSERT INTO cases (client_id, lawyer_id, created_by, ethical_wall_enabled) VALUES (33, '7', '7', 1)",
	}
	for _, stmt := range seeds {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	h := NewDemoAggregateHandler(db)
	h.SetAuthorizationService(services.NewAuthorizationService(
		repositories.NewCaseRepository(db),
		repositories.NewClientRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewDocumentRepository(db),
	))
	return h, db
}

func newIntakeWallRouter(h *DemoAggregateHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.GET("/api/v1/case-intakes", h.ListIntakeDrafts)
	r.GET("/api/v1/case-intakes/:id", h.IntakeWorkbench)
	return r
}

func intakeWallRequest(t *testing.T, r *gin.Engine, role, userID string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/case-intakes", nil)
	req.Header.Set("X-Test-Role", role)
	req.Header.Set("X-Test-User-Id", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload
}

// TestIntakeDraftListMasksWalledClientFromManagement keeps the recovery list
// behind the same ethical-wall boundary as the client and conflict APIs: a
// management role must not use the draft projection to read a walled client
// context. The wall also binds the owning lawyer until a whitelist row
// grants the matter back.
func TestIntakeDraftListMasksWalledClientFromManagement(t *testing.T) {
	h, db := newIntakeWallDB(t)
	r := newIntakeWallRouter(h)

	code, payload := intakeWallRequest(t, r, "partner", "1")
	if code != http.StatusOK {
		t.Fatalf("list status=%d payload=%v", code, payload)
	}
	data, _ := payload["data"].(map[string]interface{})
	if total, _ := data["total"].(float64); total != 0 {
		t.Fatalf("walled draft must be hidden from management, got %v", data)
	}

	count := int64(0)
	if err := db.Table("case_intakes").Where("id = ?", "intake-wall").Count(&count).Error; err != nil {
		t.Fatalf("reload intake: %v", err)
	}
	if count != 1 {
		t.Fatalf("read must not delete the draft, count=%d", count)
	}

	code, payload = intakeWallRequest(t, r, "lawyer", "7")
	if code != http.StatusOK {
		t.Fatalf("lawyer list status=%d payload=%v", code, payload)
	}
	data, _ = payload["data"].(map[string]interface{})
	if total, _ := data["total"].(float64); total != 0 {
		t.Fatalf("wall must bind the owning lawyer too before whitelisting, got %v", data)
	}

	if err := db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id) SELECT id, '7' FROM cases WHERE client_id = 33").Error; err != nil {
		t.Fatalf("whitelist lawyer: %v", err)
	}
	code, payload = intakeWallRequest(t, r, "lawyer", "7")
	if code != http.StatusOK {
		t.Fatalf("lawyer list status=%d payload=%v", code, payload)
	}
	data, _ = payload["data"].(map[string]interface{})
	if total, _ := data["total"].(float64); total != 1 {
		t.Fatalf("owning lawyer must keep the draft recoverable, got %v", data)
	}
}

// TestIntakeDraftDetailForbidsWalledClientFromManagement proves the same
// boundary on the detail alias: the management role hits the ethical-wall
// authorization before any payload exists, and the wall row alone unblocks it.
func TestIntakeDraftDetailForbidsWalledClientFromManagement(t *testing.T) {
	h, db := newIntakeWallDB(t)
	r := newIntakeWallRouter(h)
	gin.SetMode(gin.TestMode)
	detail := func(role, userID string) (int, map[string]interface{}) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/case-intakes/intake-wall", nil)
		req.Header.Set("X-Test-Role", role)
		req.Header.Set("X-Test-User-Id", userID)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var payload map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &payload)
		return w.Code, payload
	}

	code, payload := detail("partner", "1")
	if code != http.StatusForbidden {
		t.Fatalf("management detail must be 403 behind the wall, got %d %v", code, payload)
	}

	code, payload = detail("lawyer", "7")
	if code != http.StatusForbidden {
		t.Fatalf("owning lawyer detail must be 403 behind the wall too, got %d %v", code, payload)
	}

	if err := db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id) SELECT id, '1' FROM cases WHERE client_id = 33").Error; err != nil {
		t.Fatalf("whitelist management: %v", err)
	}
	if err := db.Exec("INSERT INTO case_ethical_wall_whitelist (case_id, user_id) SELECT id, '7' FROM cases WHERE client_id = 33").Error; err != nil {
		t.Fatalf("whitelist lawyer: %v", err)
	}
	code, payload = detail("partner", "1")
	if code != http.StatusOK {
		t.Fatalf("whitelisted management detail status=%d payload=%v", code, payload)
	}
	data, _ := payload["data"].(map[string]interface{})
	client, _ := data["client"].(map[string]interface{})
	if got := fmt.Sprint(client["contact_person"]); got != "陈**" {
		t.Fatalf("management must see a masked contact person, got %v", client["contact_person"])
	}
	if data["revision"] != "rev-wall" {
		t.Fatalf("restore detail must carry the revision token, got %v", data["revision"])
	}

	code, payload = detail("lawyer", "7")
	if code != http.StatusOK {
		t.Fatalf("whitelisted lawyer detail status=%d payload=%v", code, payload)
	}
}
