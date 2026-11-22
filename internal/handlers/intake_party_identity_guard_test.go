package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func newPartyIdentityGuardDB(t *testing.T) *gorm.DB {
	t.Helper()
	_, db := newDraftRevisionDB(t)
	seedMetadata := `{"draft_revision":"guard-rev-1"}`
	if err := db.Exec(`INSERT INTO case_intakes (id, title, case_type, client_id, created_by, status, metadata) VALUES ('intake-guarded', '被保护草稿', 'civil', 33, '7', 'draft', ?)`, seedMetadata).Error; err != nil {
		t.Fatalf("seed intake: %v", err)
	}
	if err := db.Exec(`INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest, aliases) VALUES ('intake-guarded', '对手公司', 'LEGAL_PERSON', 'opponent', 'SOCIAL_CREDIT_CODE', 'old-ciphertext', 'old-digest', '对手别名')`).Error; err != nil {
		t.Fatalf("seed party: %v", err)
	}
	return db
}

func partyIdentityGuardRequest(t *testing.T, r *gin.Engine, body string) (int, map[string]interface{}) {
	t.Helper()
	// The revision guard fronts every update now: contract-aware callers
	// must name the revision they loaded, so these probes carry the seeded
	// token instead of bypassing the lost-update protection.
	body = strings.Replace(body, "{", `{"expected_revision":"guard-rev-1",`, 1)
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
	h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
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
	h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
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
	h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
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

// TestUpdateIntakeDraftIdentityInheritanceRequiresExactNameAndRole locks the
// precise-match contract of the inheritance lookup: a stored protected
// identity may only be inherited when BOTH the stored entity name AND the
// stored party role equal the saved row. The previous name|role concatenation
// degraded to a logical OR on MySQL, so a payload whose name matched one old
// row and role matched another could inherit the wrong party's protected
// identity.
func TestUpdateIntakeDraftIdentityInheritanceRequiresExactNameAndRole(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("i", 32))

	seedGuarantorParty := func(t *testing.T, db *gorm.DB) {
		t.Helper()
		if err := db.Exec("INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('intake-guarded', '担保公司', 'LEGAL_PERSON', 'guarantor', 'SOCIAL_CREDIT_CODE', 'guarantor-ciphertext', 'guarantor-digest')").Error; err != nil {
			t.Fatalf("seed guarantor party: %v", err)
		}
	}
	newGuardScenario := func(t *testing.T) (*gorm.DB, *gin.Engine) {
		t.Helper()
		db := newPartyIdentityGuardDB(t)
		seedGuarantorParty(t, db)
		gin.SetMode(gin.TestMode)
		h := NewDemoAggregateHandler(db)
		h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", c.GetHeader("X-Test-Role"))
			c.Set("user_id", c.GetHeader("X-Test-User-Id"))
			c.Next()
		})
		r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)
		return db, r
	}
	partyRowCount := func(t *testing.T, db *gorm.DB) int64 {
		t.Helper()
		var count int64
		if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-guarded").Count(&count).Error; err != nil {
			t.Fatalf("count parties: %v", err)
		}
		return count
	}
	partyCiphertextByName := func(t *testing.T, db *gorm.DB, name string) string {
		t.Helper()
		row := map[string]interface{}{}
		if err := db.Table("case_intake_parties").Where("intake_id = ? AND entity_name = ?", "intake-guarded", name).Take(&row).Error; err != nil {
			t.Fatalf("load party %s: %v", name, err)
		}
		return fmt.Sprint(row["identity_number_ciphertext"])
	}

	t.Run("name from one row and role from another must not inherit", func(t *testing.T) {
		db, r := newGuardScenario(t)
		code, payload := partyIdentityGuardRequest(t, r, "{\n  \"parties\": [{\"entity_name\": \"担保公司\", \"entity_type\": \"LEGAL_PERSON\", \"party_role\": \"opponent\", \"identity_type\": \"SOCIAL_CREDIT_CODE\", \"identity_number\": \"\"}]\n}")
		if code != http.StatusConflict {
			t.Fatalf("cross-match inherit status=%d payload=%v", code, payload)
		}
		if payload["error"] != "INTAKE_PARTY_IDENTITY_REQUIRED" {
			t.Fatalf("error code=%v want INTAKE_PARTY_IDENTITY_REQUIRED", payload["error"])
		}
		if got := partyRowCount(t, db); got != 2 {
			t.Fatalf("failed save must not mutate parties, rows=%d", got)
		}
		if got := partyCiphertextByName(t, db, "担保公司"); got != "guarantor-ciphertext" {
			t.Fatalf("guarantor identity must be untouched, got %q", got)
		}
	})

	t.Run("same name with different role must not inherit", func(t *testing.T) {
		db, r := newGuardScenario(t)
		code, payload := partyIdentityGuardRequest(t, r, "{\n  \"parties\": [{\"entity_name\": \"对手公司\", \"entity_type\": \"LEGAL_PERSON\", \"party_role\": \"guarantor\", \"identity_type\": \"SOCIAL_CREDIT_CODE\", \"identity_number\": \"\"}]\n}")
		if code != http.StatusConflict {
			t.Fatalf("role-only match inherit status=%d payload=%v", code, payload)
		}
		if payload["error"] != "INTAKE_PARTY_IDENTITY_REQUIRED" {
			t.Fatalf("error code=%v want INTAKE_PARTY_IDENTITY_REQUIRED", payload["error"])
		}
		if got := partyCiphertextByName(t, db, "对手公司"); got != "old-ciphertext" {
			t.Fatalf("opponent identity must be untouched, got %q", got)
		}
	})

	t.Run("exact name and role inherits its own protected identity", func(t *testing.T) {
		db, r := newGuardScenario(t)
		code, payload := partyIdentityGuardRequest(t, r, "{\n  \"parties\": [{\"entity_name\": \"对手公司\", \"entity_type\": \"LEGAL_PERSON\", \"party_role\": \"opponent\", \"identity_type\": \"SOCIAL_CREDIT_CODE\", \"identity_number\": \"\"}]\n}")
		if code != http.StatusOK {
			t.Fatalf("exact-match inherit status=%d payload=%v", code, payload)
		}
		if got := partyRowCount(t, db); got != 1 {
			t.Fatalf("save replaces parties with payload rows, rows=%d", got)
		}
		if got := partyCiphertextByName(t, db, "对手公司"); got != "old-ciphertext" {
			t.Fatalf("protected identity not inherited from exact match: %q", got)
		}
	})
}

func TestUpdateIntakeDraftIdentityInheritanceFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saveWithEmptyIdentity := func(t *testing.T, seedSQL string) (int, map[string]interface{}, *gorm.DB) {
		t.Helper()
		t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("j", 32))
		db := newPartyIdentityGuardDB(t)
		if seedSQL != "" {
			if err := db.Exec(seedSQL).Error; err != nil {
				t.Fatalf("seed scenario party: %v", err)
			}
		}
		h := NewDemoAggregateHandler(db)
		h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("role", c.GetHeader("X-Test-Role"))
			c.Set("user_id", c.GetHeader("X-Test-User-Id"))
			c.Next()
		})
		r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)
		code, payload := partyIdentityGuardRequest(t, r, `{"parties":[{"entity_name":"对手公司","entity_type":"LEGAL_PERSON","party_role":"opponent","identity_type":"SOCIAL_CREDIT_CODE","identity_number":""}]}`)
		return code, payload, db
	}
	opponentCount := func(t *testing.T, db *gorm.DB) int64 {
		t.Helper()
		var count int64
		if err := db.Table("case_intake_parties").Where("intake_id = ? AND entity_name = ?", "intake-guarded", "对手公司").Count(&count).Error; err != nil {
			t.Fatalf("count opponents: %v", err)
		}
		return count
	}

	t.Run("handler duplicate exact matches are rejected", func(t *testing.T) {
		code, payload, db := saveWithEmptyIdentity(t, "INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('intake-guarded', '对手公司', 'LEGAL_PERSON', 'opponent', 'SOCIAL_CREDIT_CODE', 'duplicate-ciphertext', 'duplicate-digest')")
		if code != http.StatusConflict || payload["error"] != "INTAKE_PARTY_IDENTITY_REQUIRED" {
			t.Fatalf("identity guard status=%d payload=%v", code, payload)
		}
		if got := opponentCount(t, db); got != 2 {
			t.Fatalf("failed save must not mutate parties, rows=%d", got)
		}
	})

	t.Run("handler mismatched entity type is rejected", func(t *testing.T) {
		code, payload, db := saveWithEmptyIdentity(t, "UPDATE case_intake_parties SET entity_type = 'INDIVIDUAL' WHERE entity_name = '对手公司'")
		if code != http.StatusConflict || payload["error"] != "INTAKE_PARTY_IDENTITY_REQUIRED" {
			t.Fatalf("identity guard status=%d payload=%v", code, payload)
		}
		if got := opponentCount(t, db); got != 1 {
			t.Fatalf("failed save must not mutate parties, rows=%d", got)
		}
	})

	t.Run("handler digest missing is rejected", func(t *testing.T) {
		code, payload, db := saveWithEmptyIdentity(t, "UPDATE case_intake_parties SET identity_number_digest = '' WHERE entity_name = '对手公司'")
		if code != http.StatusConflict || payload["error"] != "INTAKE_PARTY_IDENTITY_REQUIRED" {
			t.Fatalf("identity guard status=%d payload=%v", code, payload)
		}
		if got := opponentCount(t, db); got != 1 {
			t.Fatalf("failed save must not mutate parties, rows=%d", got)
		}
	})

	emptyIdentityParty := func() map[string]interface{} {
		return map[string]interface{}{
			"entity_name":     "对手公司",
			"entity_type":     "LEGAL_PERSON",
			"party_role":      "opponent",
			"identity_type":   "SOCIAL_CREDIT_CODE",
			"identity_number": "",
		}
	}
	helperUpdateRow := func(t *testing.T, db *gorm.DB, party map[string]interface{}) (map[string]interface{}, error) {
		t.Helper()
		h := NewDemoAggregateHandler(db)
		h.SetAuthorizationService(services.NewAuthorizationService(repositories.NewCaseRepository(db), repositories.NewClientRepository(db), repositories.NewUserRepository(db), repositories.NewDocumentRepository(db)))
		tx := db.Begin()
		if tx.Error != nil {
			t.Fatalf("begin tx: %v", tx.Error)
		}
		defer tx.Rollback()
		return h.prepareCaseIntakePartyUpdateRow(tx, "intake-guarded", party, time.Now(), true)
	}

	t.Run("helper inherits only a single fully guarded exact match", func(t *testing.T) {
		db := newPartyIdentityGuardDB(t)
		row, err := helperUpdateRow(t, db, emptyIdentityParty())
		if err != nil {
			t.Fatalf("exact match inherit error: %v", err)
		}
		if fmt.Sprint(row["identity_type"]) != "SOCIAL_CREDIT_CODE" ||
			fmt.Sprint(row["identity_number_ciphertext"]) != "old-ciphertext" ||
			fmt.Sprint(row["identity_number_digest"]) != "old-digest" ||
			fmt.Sprint(row["aliases"]) != "对手别名" {
			t.Fatalf("guarded identity not inherited: %v", row)
		}
	})

	helperRejectScenarios := []struct {
		name    string
		seedSQL string
	}{
		{name: "helper duplicate exact matches fail closed", seedSQL: "INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('intake-guarded', '对手公司', 'LEGAL_PERSON', 'opponent', 'SOCIAL_CREDIT_CODE', 'duplicate-ciphertext', 'duplicate-digest')"},
		{name: "helper mismatched entity type fails closed", seedSQL: "UPDATE case_intake_parties SET entity_type = 'INDIVIDUAL' WHERE entity_name = '对手公司'"},
		{name: "helper missing digest fails closed", seedSQL: "UPDATE case_intake_parties SET identity_number_digest = '' WHERE entity_name = '对手公司'"},
	}
	for _, scenario := range helperRejectScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			db := newPartyIdentityGuardDB(t)
			if scenario.seedSQL != "" {
				if err := db.Exec(scenario.seedSQL).Error; err != nil {
					t.Fatalf("seed scenario: %v", err)
				}
			}
			_, err := helperUpdateRow(t, db, emptyIdentityParty())
			var workflowErr *services.SubjectWorkflowError
			if !errors.As(err, &workflowErr) || workflowErr.Code != "INTAKE_PARTY_IDENTITY_REQUIRED" {
				t.Fatalf("fail-closed error=%v want INTAKE_PARTY_IDENTITY_REQUIRED", err)
			}
			var count int64
			if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-guarded").Count(&count).Error; err != nil {
				t.Fatalf("count parties: %v", err)
			}
			if count == 0 {
				t.Fatalf("failed inheritance must keep prior rows")
			}
		})
	}
}
