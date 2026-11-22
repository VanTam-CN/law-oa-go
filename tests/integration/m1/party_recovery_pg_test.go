//go:build np_m1_pg

package m1_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/config"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/router"
	"law-oa-go/internal/security"
)

func TestNativeRecoveredOpponentPreservesPartyAndMaterialRows(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("t", 32))
	t.Setenv("JWT_SECRET", "synthetic-only-for-the-isolated-m1-test")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	middleware.InitJWT(cfg)
	g := gin.New()
	if err := router.Init(g, db, nil); err != nil {
		t.Fatal(err)
	}
	token, _, err := middleware.GenerateToken(f.userID, "fixture", "lawyer")
	if err != nil {
		t.Fatal(err)
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		if err := db.Exec(sql, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,aliases,metadata,identity_type,identity_number_ciphertext,identity_number_digest)
		VALUES(?, 'Test opponent','LEGAL_PERSON','opposing_party','original alias','{"opaque":9007199254740993}','SOCIAL_CREDIT_CODE','old cipher','old digest')`, f.intakeID)
	exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,metadata)
		SELECT ?, 'synthetic-' || n, 'LEGAL_PERSON', CASE WHEN n=1 THEN 'client' ELSE 'related_party' END, jsonb_build_object('opaque',n)
		FROM generate_series(1,23) AS n`, f.intakeID)
	exec(`INSERT INTO case_materials(intake_id,name,metadata) SELECT ?, 'material-' || n, jsonb_build_object('n',n) FROM generate_series(1,2) AS n`, f.intakeID)
	// Compare the entire stored rows, including IDs, unknown JSON and timestamps.
	snapshot := func() string {
		t.Helper()
		var value string
		err := db.Raw(`SELECT jsonb_build_array(
			(SELECT jsonb_agg(CASE WHEN party_role='opposing_party' THEN to_jsonb(p)-'identity_type'-'identity_number_ciphertext'-'identity_number_digest' ELSE to_jsonb(p) END ORDER BY id) FROM case_intake_parties p WHERE intake_id=?),
			(SELECT jsonb_agg(to_jsonb(m) ORDER BY id) FROM case_materials m WHERE intake_id=?)
		)::text`, f.intakeID, f.intakeID).Scan(&value).Error
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := snapshot()
	const identity = "TEST-NEW-IDENTITY"
	body, err := json.Marshal(map[string]any{
		"description": "reviewed native draft", "expected_revision": f.revision,
		"opponent_identity_update": map[string]string{"entity_name": "Test opponent", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": identity},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/case-intakes/"+f.intakeID, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("native save returned %d", w.Code)
	}
	var target struct{ Cipher, Digest string }
	if err := db.Raw("SELECT identity_number_ciphertext AS cipher,identity_number_digest AS digest FROM case_intake_parties WHERE intake_id=? AND party_role='opposing_party'", f.intakeID).Scan(&target).Error; err != nil {
		t.Fatal(err)
	}
	want, err := security.IdentityDigest(security.NormalizeIdentityNumber("SOCIAL_CREDIT_CODE", identity))
	if err != nil || target.Digest != want || target.Cipher == "" || target.Cipher == "old cipher" {
		t.Fatal("new protected identity did not reach the selected opponent")
	}
	if snapshot() != before {
		t.Fatal("unselected party, material or unrelated opponent fields changed")
	}
	version, state, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	if version == "" || version == f.revision || state != "draft" {
		t.Fatal("identity correction must invalidate confirmation")
	}
}
