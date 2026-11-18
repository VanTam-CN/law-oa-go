package handlers

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

// The continuation endpoint must serve the server-verified decision through
// the same object boundary as the workbench, reject stale revisions, and
// never leak identity digests, ciphertext, or plaintext identity material.

func seedContinuationHandlerWorld(t *testing.T) (*DemoAggregateHandler, *gin.Engine, string, string) {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("c", 32))
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/continuation_handler.db")), &gorm.Config{})
	require.NoError(t, err)
	// models.Case backs the ethical-wall authorization query; production
	// always has the table, and an empty fixture keeps the client readable.
	require.NoError(t, db.AutoMigrate(&models.Client{}, &models.ApprovalRequest{}, &models.Case{}))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, client_id INTEGER, created_by TEXT, status TEXT, metadata TEXT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, relation_depth INTEGER, identity_type TEXT, identity_number_ciphertext TEXT, identity_number_digest TEXT, aliases TEXT, metadata TEXT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE conflict_check_records (check_id TEXT PRIMARY KEY, client_id TEXT, client_name TEXT, check_status TEXT, search_parameters TEXT, check_result TEXT, user_id INTEGER, check_time DATETIME, created_at DATETIME, updated_at DATETIME)").Error)
	require.NoError(t, db.Exec("CREATE TABLE conflict_reviews (id INTEGER PRIMARY KEY AUTOINCREMENT, check_id TEXT, decision TEXT, notes TEXT, reviewer_id INTEGER, reviewer_name TEXT, evidence_hash TEXT, created_at DATETIME)").Error)

	ciphertext, digest, err := security.ProtectIdentityNumber(security.NormalizeIdentityNumber(string(models.IdentityTypeSocialCredit), "91310000MA1FL0000X"))
	require.NoError(t, err)
	require.NoError(t, db.Create(&models.Client{
		Name:                     "星河智联科技有限公司",
		Type:                     "企业",
		Email:                    "cont-client@example.com",
		IdentityType:             models.IdentityTypeSocialCredit,
		IdentityNumberCiphertext: ciphertext,
		IdentityNumberDigest:     digest,
		CreatedBy:                7,
	}).Error)
	partyCiphertext, partyDigest, err := security.ProtectIdentityNumber(security.NormalizeIdentityNumber(string(models.IdentityTypeSocialCredit), "91310000MA1FL0001X"))
	require.NoError(t, err)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, relation_depth, identity_type, identity_number_ciphertext, identity_number_digest, aliases, metadata) "+
			"VALUES ('intake-cont', '对手公司', 'COMPANY', 'OPPOSING_PARTY', 0, 'SOCIAL_CREDIT_CODE', ?, ?, '', '{\"party_type\":\"OPPOSING_PARTY\"}')",
		partyCiphertext, partyDigest).Error)
	intakeMeta := "{\"draft_revision\":\"rev-1\",\"conflict_check_id\":\"CCT_CONT_H1\",\"lawyer_facts_version\":\"FV-1\",\"conflict_check_facts_version\":\"FV-1\"}"
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, client_id, created_by, status, metadata) VALUES ('intake-cont', 1, '7', 'conflict_ready', ?)",
		intakeMeta).Error)
	// The frozen payload must carry the COMPLETE subject tuple the check
	// actually audited: client name/type plus the opposing party's full
	// identity anchor. A name-only party never proves identity.
	frozen := fmt.Sprintf("{\"intakeId\":\"intake-cont\",\"clientName\":\"星河智联科技有限公司\",\"clientType\":\"COMPANY\",\"clientIdentifiers\":{\"unified_social_credit_code\":\"hmac-sha256:%s\"},\"parties\":[{\"name\":\"对手公司\",\"role\":\"OPPOSING_PARTY\",\"entityType\":\"COMPANY\",\"identifiers\":{\"social_credit_code\":\"hmac-sha256:%s\"}}]}", digest, partyDigest)
	coverage := "{\"decision\":{\"coverageStatus\":\"COMPLETE\"}}"
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_check_records (check_id, client_id, client_name, check_status, search_parameters, check_result, user_id, check_time, created_at, updated_at) "+
			"VALUES ('CCT_CONT_H1', '1', '星河智联科技有限公司', 'COMPLETED', ?, ?, 7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		frozen, coverage).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_reviews (check_id, decision, notes, reviewer_id, reviewer_name, evidence_hash, created_at) "+
			"VALUES ('CCT_CONT_H1', 'no_conflict', '独立复核通过', 42, '核查人', 'hash', CURRENT_TIMESTAMP)").Error)

	h := NewDemoAggregateHandler(db)
	h.SetAuthorizationService(services.NewAuthorizationService(
		repositories.NewCaseRepository(db),
		repositories.NewClientRepository(db),
		repositories.NewUserRepository(db),
		repositories.NewDocumentRepository(db),
	))
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", c.GetHeader("X-Test-Role"))
		c.Set("user_id", c.GetHeader("X-Test-User-Id"))
		c.Next()
	})
	r.GET("/api/v1/case-intakes/:id/continuation", h.IntakeContinuation)
	return h, r, digest, partyDigest
}

func continuationRequest(t *testing.T, r *gin.Engine, path, revisionHeader, actorID string) (int, map[string]interface{}, string) {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	req.Header.Set("X-Test-Role", "lawyer")
	req.Header.Set("X-Test-User-Id", actorID)
	if revisionHeader != "" {
		req.Header.Set("X-Intake-Revision", revisionHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var payload map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload, w.Body.String()
}

func TestIntakeContinuationHandlerReturnsVerifiedDecision(t *testing.T) {
	_, r, digest, partyDigest := seedContinuationHandlerWorld(t)
	code, payload, body := continuationRequest(t, r, "/api/v1/case-intakes/intake-cont/continuation", "rev-1", "7")
	require.Equal(t, 200, code)
	data := payload["data"].(map[string]interface{})
	require.Equal(t, true, data["can_reuse_identity"])
	require.Equal(t, true, data["can_reuse_check"])
	require.Equal(t, true, data["can_submit"])
	require.Equal(t, false, data["recheck_required"])
	require.NotContains(t, body, "identity_digest")
	require.NotContains(t, body, "ciphertext")
	require.NotContains(t, body, "91310000MA1FL0000X")
	require.NotContains(t, body, "91310000MA1FL0001X")
	require.NotContains(t, body, "hmac-sha256:")
	require.NotContains(t, body, digest)
	require.NotContains(t, body, partyDigest)
}

func TestIntakeContinuationHandlerRejectsStaleRevision(t *testing.T) {
	_, r, _, _ := seedContinuationHandlerWorld(t)
	code, payload, _ := continuationRequest(t, r, "/api/v1/case-intakes/intake-cont/continuation", "rev-stale", "7")
	require.Equal(t, 409, code)
	errorBody, ok := payload["error"].(map[string]interface{})
	require.True(t, ok, "conflict responses use the APIErrorWithContext envelope: %v", payload)
	require.Equal(t, "INTAKE_DRAFT_REVISION_CONFLICT", errorBody["code"])
}

func TestIntakeContinuationHandlerForbidsOtherLawyer(t *testing.T) {
	_, r, _, _ := seedContinuationHandlerWorld(t)
	code, _, _ := continuationRequest(t, r, "/api/v1/case-intakes/intake-cont/continuation", "", "99")
	require.Equal(t, 403, code)
}

func TestIntakeContinuationHandlerMissingIntake(t *testing.T) {
	_, r, _, _ := seedContinuationHandlerWorld(t)
	code, _, _ := continuationRequest(t, r, "/api/v1/case-intakes/intake-missing/continuation", "", "7")
	require.Equal(t, 404, code)
}
