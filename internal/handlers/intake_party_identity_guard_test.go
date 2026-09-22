package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPartyIdentityGuardDB(t *testing.T) *gorm.DB {
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
	if err := db.Exec(`CREATE TABLE case_intakes (id TEXT PRIMARY KEY, title TEXT, case_type TEXT, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT, updated_at DATETIME)`).Error; err != nil {
		t.Fatalf("create case_intakes: %v", err)
	}
	if err := db.Exec(`CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, relation_depth INTEGER, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT, metadata TEXT, created_at DATETIME)`).Error; err != nil {
		t.Fatalf("create case_intake_parties: %v", err)
	}
	if err := db.Exec(`INSERT INTO case_intakes (id, title, case_type, client_id, created_by, status, metadata) VALUES ('intake-guarded', '被保护草稿', 'civil', 3, '7', 'draft', '{}')`).Error; err != nil {
		t.Fatalf("seed intake: %v", err)
	}
	if err := db.Exec(`INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, aliases) VALUES ('intake-guarded', '对手公司', 'LEGAL_PERSON', 'opponent', 'SOCIAL_CREDIT_CODE', 'old-ciphertext', 'old-digest', '对手别名')`).Error; err != nil {
		t.Fatalf("seed party: %v", err)
	}
	return db
}

func partyIdentityGuardRequest(t *testing.T, r *gin.Engine, body string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/case-intakes/intake-guarded", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "lawyer")
	req.Header.Set("X-Test-User-Id", "7")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload
}

func storedPartyIdentity(t *testing.T, db *gorm.DB) (string, string, string) {
	t.Helper()
	row := map[string]interface{}{}
	if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-guarded").Take(&row).Error; err != nil {
		t.Fatalf("reload party: %v", err)
	}
	return fmt.Sprint(row["identity_type"]), fmt.Sprint(row["identity_number_ciphertext"]), fmt.Sprint(row["identity_number_digest"])
}

func TestUpdateIntakeDraftInheritsProtectedIdentityWhenInputEmpty(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("g", 32))
	db := newPartyIdentityGuardDB(t)
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)

	code, payload := partyIdentityGuardRequest(t, r, `{"parties":[{"entity_name":"对手公司","entity_type":"LEGAL_PERSON","party_role":"opponent","identity_type":"SOCIAL_CREDIT_CODE","identity_number":""}]}`)
	if code != http.StatusOK {
		t.Fatalf("inherit save status=%d payload=%v", code, payload)
	}
	identityType, ciphertext, digest := storedPartyIdentity(t, db)
	if identityType != "SOCIAL_CREDIT_CODE" || ciphertext != "old-ciphertext" || digest != "old-digest" {
		t.Fatalf("protected identity not inherited: type=%q ciphertext=%q digest=%q", identityType, ciphertext, digest)
	}
}

func TestUpdateIntakeDraftKeepsNewIdentityWhenProvided(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("h", 32))
	db := newPartyIdentityGuardDB(t)
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)

	code, payload := partyIdentityGuardRequest(t, r, `{"parties":[{"entity_name":"对手公司","entity_type":"LEGAL_PERSON","party_role":"opponent","identity_type":"SOCIAL_CREDIT_CODE","identity_number":"91310000MA1FL00X0B"}]}`)
	if code != http.StatusOK {
		t.Fatalf("re-identified save status=%d payload=%v", code, payload)
	}
	_, ciphertext, _ := storedPartyIdentity(t, db)
	if ciphertext == "old-ciphertext" || strings.TrimSpace(ciphertext) == "" {
		t.Fatalf("provided identity was not stored: %q", ciphertext)
	}
}

func TestUpdateIntakeDraftStillRequiresIdentityForNewParty(t *testing.T) {
	db := newPartyIdentityGuardDB(t)
	gin.SetMode(gin.TestMode)
	h := NewDemoAggregateHandler(db)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)

	code, payload := partyIdentityGuardRequest(t, r, `{"parties":[{"entity_name":"新被告","entity_type":"LEGAL_PERSON","party_role":"opponent","identity_type":"SOCIAL_CREDIT_CODE","identity_number":""}]}`)
	if code != http.StatusConflict {
		t.Fatalf("new party without identity status=%d payload=%v", code, payload)
	}
	if payload["error"] != "INTAKE_PARTY_IDENTITY_REQUIRED" {
		t.Fatalf("error code=%v want INTAKE_PARTY_IDENTITY_REQUIRED", payload["error"])
	}
}
