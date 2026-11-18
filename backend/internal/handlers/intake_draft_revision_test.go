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

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

var (
	draftRevDigestA string
	draftRevDigestB string
)

func newDraftRevisionDB(t *testing.T) (*DemoAggregateHandler, *gorm.DB) {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("r", 32))
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&models.Client{}); err != nil {
		t.Fatalf("migrate client: %v", err)
	}
	ddl := []string{
		"CREATE TABLE case_intakes (id TEXT PRIMARY KEY, intake_code TEXT, title TEXT, case_type TEXT, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT, created_at DATETIME, updated_at DATETIME)",
		"CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, relation_depth INTEGER, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT, metadata TEXT, created_at DATETIME)",
		"CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, role TEXT, status TEXT, department TEXT)",
		"CREATE TABLE cases (id INTEGER PRIMARY KEY AUTOINCREMENT, client_id INTEGER, lawyer_id TEXT, created_by TEXT, deleted_at DATETIME, ethical_wall_enabled BOOLEAN DEFAULT FALSE)",
		"CREATE TABLE case_ethical_wall_whitelist (id INTEGER PRIMARY KEY AUTOINCREMENT, case_id INTEGER, user_id TEXT)",
	}
	for _, stmt := range ddl {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table: %v", err)
		}
	}
	seedClient := func(id uint, name, clientType string, identityType models.IdentityType, identityNumber string) models.Client {
		t.Helper()
		ciphertext, digest, err := security.ProtectIdentityNumber(security.NormalizeIdentityNumber(string(identityType), identityNumber))
		if err != nil {
			t.Fatalf("protect identity: %v", err)
		}
		client := models.Client{
			Name:                     name,
			Type:                     clientType,
			Email:                    fmt.Sprintf("client-%d@example.com", id),
			IdentityType:             identityType,
			IdentityNumberCiphertext: ciphertext,
			IdentityNumberDigest:     digest,
			CreatedBy:                7,
		}
		client.ID = id
		if err := db.Create(&client).Error; err != nil {
			t.Fatalf("seed client %d: %v", id, err)
		}
		return client
	}
	clientA := seedClient(33, "星河智联科技有限公司", "企业", models.IdentityTypeSocialCredit, "91310000MA1FL0000X")
	clientB := seedClient(999, "林individual", "个人", models.IdentityTypeIDCard, "310110199001010011")
	draftRevDigestA = clientA.IdentityDigestValue()
	draftRevDigestB = clientB.IdentityDigestValue()
	seeds := []string{
		"INSERT INTO users (id, name, role, status, department) VALUES (7, '律师甲', 'lawyer', 'active', '争议解决')",
		"INSERT INTO case_intakes (id, intake_code, title, case_type, client_id, created_by, status, metadata) VALUES ('intake-rev', 'INT-900', '版本守护草稿', 'civil', 33, '7', 'draft', '{}')",
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('intake-rev', '星河智联科技有限公司', 'LEGAL_PERSON', 'client', 'SOCIAL_CREDIT_CODE', '" + clientA.IdentityNumberCiphertext + "', '" + clientA.IdentityNumberDigest + "')",
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, identity_type, identity_number_ciphertext, identity_number_digest) VALUES ('intake-rev', '对手公司', 'LEGAL_PERSON', 'opposing_party', 'SOCIAL_CREDIT_CODE', 'opp-enc', 'opp-digest')",
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

func newDraftRevisionRouter(h *DemoAggregateHandler) *gin.Engine {
	t := &testing.T{}
	_ = t
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.GET("/api/v1/case-intakes", h.ListIntakeDrafts)
	r.GET("/api/v1/case-intakes/:id", h.IntakeWorkbench)
	r.POST("/api/v1/case-intakes/:id/facts-confirmation", h.ConfirmIntakeFacts)
	r.PUT("/api/v1/case-intakes/:id", h.UpdateCaseIntake)
	return r
}

func draftRevisionRequest(t *testing.T, r *gin.Engine, method, path, body, revisionHeader string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-Role", "lawyer")
	req.Header.Set("X-Test-User-Id", "7")
	if revisionHeader != "" {
		req.Header.Set("X-Intake-Revision", revisionHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload
}

func stampDraftRevision(t *testing.T, db *gorm.DB, revision string) {
	t.Helper()
	metadata := "{\"draft_revision\":\"" + revision + "\"}"
	if err := db.Table("case_intakes").Where("id = ?", "intake-rev").Update("metadata", metadata).Error; err != nil {
		t.Fatalf("stamp revision: %v", err)
	}
}

func draftRevisionServerState(t *testing.T, db *gorm.DB) (string, string, string, string) {
	t.Helper()
	row := map[string]interface{}{}
	if err := db.Table("case_intakes").Where("id = ?", "intake-rev").Take(&row).Error; err != nil {
		t.Fatalf("reload intake: %v", err)
	}
	var parties []map[string]interface{}
	if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-rev").Order("id").Find(&parties).Error; err != nil {
		t.Fatalf("reload parties: %v", err)
	}
	clientDigest := ""
	opponentDigest := ""
	for _, party := range parties {
		switch fmt.Sprint(party["party_role"]) {
		case "client":
			clientDigest = fmt.Sprint(party["identity_number_digest"])
		case "opposing_party":
			opponentDigest = fmt.Sprint(party["identity_number_digest"])
		}
	}
	return intakeDraftRevision(row["metadata"]), clientDigest, opponentDigest, fmt.Sprint(row["title"])
}

func TestUpdateIntakeRejectsMissingExpectedRevisionWithoutMutation(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)

	code, payload := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", "{\"title\":\"缺版本保存\"}", "")
	if code != http.StatusConflict {
		t.Fatalf("status=%d want 409 body=%v", code, payload)
	}
	errObj, _ := payload["error"].(map[string]interface{})
	if errObj == nil || errObj["code"] != "INTAKE_DRAFT_REVISION_REQUIRED" {
		t.Fatalf("want INTAKE_DRAFT_REVISION_REQUIRED, got %v", payload)
	}
	rev, clientDigest, opponentDigest, title := draftRevisionServerState(t, db)
	if title != "版本守护草稿" || clientDigest != draftRevDigestA || opponentDigest != "opp-digest" || rev != "" {
		t.Fatalf("server mutated on missing revision: rev=%q client=%q opponent=%q title=%q", rev, clientDigest, opponentDigest, title)
	}
}

func TestUpdateIntakeRejectsStaleRevisionWithoutOverwrite(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")

	codeA, savedA := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", "{\"expected_revision\":\"rev-1\",\"title\":\"A页改标题\"}", "")
	if codeA != http.StatusOK {
		t.Fatalf("tab A save: status=%d payload=%v", codeA, savedA)
	}
	dataA, _ := savedA["data"].(map[string]interface{})
	rev2, _ := dataA["revision"].(string)
	if rev2 == "" || rev2 == "rev-1" {
		t.Fatalf("tab A must return a fresh revision, got %q", rev2)
	}

	codeB, conflictB := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", "{\"expected_revision\":\"rev-1\",\"description\":\"B页旧版本描述\"}", "")
	if codeB != http.StatusConflict {
		t.Fatalf("tab B stale save: status=%d want 409 payload=%v", codeB, conflictB)
	}
	errObjB, _ := conflictB["error"].(map[string]interface{})
	if errObjB == nil || errObjB["code"] != "INTAKE_DRAFT_REVISION_CONFLICT" {
		t.Fatalf("want INTAKE_DRAFT_REVISION_CONFLICT, got %v", conflictB)
	}
	ctxB, _ := errObjB["context"].(map[string]interface{})
	if ctxB == nil || ctxB["current_revision"] != rev2 {
		t.Fatalf("conflict must carry current revision %q, got %v", rev2, conflictB)
	}

	rev, clientDigest, opponentDigest, title := draftRevisionServerState(t, db)
	if title != "A页改标题" {
		t.Fatalf("A title must survive, got %q", title)
	}
	if clientDigest != draftRevDigestA || opponentDigest != "opp-digest" {
		t.Fatalf("parties must survive conflict: client=%q opponent=%q", clientDigest, opponentDigest)
	}
	if rev != rev2 {
		t.Fatalf("revision unchanged expected %q got %q", rev2, rev)
	}
}

func TestUpdateIntakeSwitchClientUsesNewMasterIdentity(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")

	body := "{\"expected_revision\":\"rev-1\",\"client_id\":999,\"parties\":[" +
		"{\"entity_name\":\"林individual\",\"entity_type\":\"INDIVIDUAL\",\"party_role\":\"client\"}," +
		"{\"entity_name\":\"对手公司\",\"entity_type\":\"LEGAL_PERSON\",\"party_role\":\"opposing_party\",\"identity_type\":\"SOCIAL_CREDIT_CODE\",\"identity_number\":\"91310000MA1FL0000X\"}]}"
	code, payload := draftRevisionRequest(t, r, http.MethodPut, "/api/v1/case-intakes/intake-rev", body, "")
	if code != http.StatusOK {
		t.Fatalf("client switch save: status=%d payload=%v", code, payload)
	}
	var parties []map[string]interface{}
	if err := db.Table("case_intake_parties").Where("intake_id = ?", "intake-rev").Order("id").Find(&parties).Error; err != nil {
		t.Fatalf("reload parties: %v", err)
	}
	clientDigest := ""
	for _, party := range parties {
		if fmt.Sprint(party["party_role"]) == "client" {
			clientDigest = fmt.Sprint(party["identity_number_digest"])
			if fmt.Sprint(party["entity_name"]) != "林individual" || fmt.Sprint(party["entity_type"]) != "INDIVIDUAL" {
				t.Fatalf("client party must reflect new client, got %v", party)
			}
		}
	}
	if clientDigest != draftRevDigestB {
		t.Fatalf("client party must carry NEW client master identity, got %q want %q", clientDigest, draftRevDigestB)
	}
}

func TestConfirmIntakeFactsRotatesRevisionAndRejectsStaleHeader(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")

	code, payload := draftRevisionRequest(t, r, http.MethodPost, "/api/v1/case-intakes/intake-rev/facts-confirmation", "{}", "rev-1")
	if code != http.StatusOK {
		t.Fatalf("confirm with fresh header: status=%d payload=%v", code, payload)
	}
	data, _ := payload["data"].(map[string]interface{})
	confirmedRev, _ := data["revision"].(string)
	if confirmedRev == "" || confirmedRev == "rev-1" {
		t.Fatalf("confirmation must rotate revision, got %q", confirmedRev)
	}
	rev, _, _, _ := draftRevisionServerState(t, db)
	if rev != confirmedRev {
		t.Fatalf("stored revision %q must equal response %q", rev, confirmedRev)
	}

	staleCode, stalePayload := draftRevisionRequest(t, r, http.MethodPost, "/api/v1/case-intakes/intake-rev/facts-confirmation", "{}", "rev-1")
	if staleCode != http.StatusConflict {
		t.Fatalf("stale header confirm: status=%d want 409 payload=%v", staleCode, stalePayload)
	}
	revAfter, clientDigest, opponentDigest, _ := draftRevisionServerState(t, db)
	if revAfter != confirmedRev || clientDigest != draftRevDigestA || opponentDigest != "opp-digest" {
		t.Fatalf("rejected stale confirmation must not mutate state: rev=%q client=%q opponent=%q", revAfter, clientDigest, opponentDigest)
	}
}

func TestDraftRestoreReturnsRevisionAndProtectedIdentityStatus(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")

	code, payload := draftRevisionRequest(t, r, http.MethodGet, "/api/v1/case-intakes/intake-rev", "", "")
	if code != http.StatusOK {
		t.Fatalf("restore read: status=%d payload=%v", code, payload)
	}
	data, _ := payload["data"].(map[string]interface{})
	if data == nil || data["revision"] != "rev-1" {
		t.Fatalf("restore must return revision, got %v", payload)
	}
	parties, _ := data["parties"].([]interface{})
	clientStatus := ""
	for _, raw := range parties {
		party, _ := raw.(map[string]interface{})
		if fmt.Sprint(party["party_role"]) == "client" {
			if _, leak := party["identity_number_ciphertext"]; leak {
				t.Fatalf("ciphertext must not leak: %v", party)
			}
			if _, leak := party["identity_number_digest"]; leak {
				t.Fatalf("digest must not leak: %v", party)
			}
			clientStatus = fmt.Sprint(party["identity_status"])
		}
	}
	if clientStatus != "已登记（受保护）" {
		t.Fatalf("client party must carry protected status, got %q", clientStatus)
	}
}

func TestDraftListCarriesRevisionTokens(t *testing.T) {
	h, db := newDraftRevisionDB(t)
	r := newDraftRevisionRouter(h)
	stampDraftRevision(t, db, "rev-1")

	code, payload := draftRevisionRequest(t, r, http.MethodGet, "/api/v1/case-intakes", "", "")
	if code != http.StatusOK {
		t.Fatalf("list: status=%d payload=%v", code, payload)
	}
	data, _ := payload["data"].(map[string]interface{})
	items, _ := data["items"].([]interface{})
	found := false
	for _, raw := range items {
		row, _ := raw.(map[string]interface{})
		if fmt.Sprint(row["id"]) == "intake-rev" {
			found = true
			if row["revision"] != "rev-1" {
				t.Fatalf("list row must carry revision, got %v", row)
			}
		}
	}
	if !found {
		t.Fatalf("draft row missing from list: %v", items)
	}
}
