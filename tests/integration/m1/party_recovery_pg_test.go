//go:build np_m1_pg

package m1_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
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

func nativePartyRequest(t *testing.T, db *gorm.DB, f conflictPGFixture) func(string, string, any) *httptest.ResponseRecorder {
	t.Helper()
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
	return func(method, path string, input any) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		g.ServeHTTP(w, req)
		return w
	}
}

func TestNativeRecoveredOpponentPreservesPartyAndMaterialRows(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
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
	w := send(http.MethodPut, "/api/v1/case-intakes/"+f.intakeID, map[string]any{
		"description": "reviewed native draft", "expected_revision": f.revision,
		"opponent_identity_update": map[string]string{"entity_name": "Test opponent", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": identity},
	})
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

func TestNativeCreateRejectsNormalizedEmptyIdentityAtomically(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	title := "empty-identity-" + f.intakeID
	party := func(suffix, number string) map[string]string {
		return map[string]string{"entity_name": title + suffix, "party_role": "related_party", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": number}
	}
	w := send("POST", "/api/v1/case-intakes", map[string]any{"title": title, "case_type": "civil", "client_id": f.clientID, "parties": []any{party("-valid", "FIXTURE1234"), party("-empty", "- - -")}, "materials": []any{map[string]string{"name": title}}})
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "INTAKE_PARTY_IDENTITY_REQUIRED") {
		t.Fatalf("wanted identity rejection, got %d", w.Code)
	}
	var rows int
	if err := db.Raw("SELECT (SELECT count(*) FROM case_intakes WHERE title=?) + (SELECT count(*) FROM case_intake_parties WHERE entity_name IN (?,?)) + (SELECT count(*) FROM case_materials WHERE name=?)", title, title+"-valid", title+"-empty", title).Scan(&rows).Error; err != nil || rows != 0 {
		t.Fatalf("rollback: rows=%d err=%v", rows, err)
	}
}

func TestNativeFullSaveOnlyChangesExplicitAlias(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	var parties []map[string]string
	for i, role := range []string{"opposing_party", "related_party", "related_party"} {
		name := fmt.Sprintf("party-%d", i)
		if err := db.Exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,entity_id,party_role,relation_depth,aliases,metadata,identity_type,identity_number_ciphertext,identity_number_digest)
   VALUES(?,?,'LEGAL_PERSON',?,?,1,'old alias','{"party_type":"guarantor","opaque":9007199254740993}'::jsonb,'SOCIAL_CREDIT_CODE',?,?)`, f.intakeID, name, 100+i, role, "cipher-"+name, "digest-"+name).Error; err != nil {
			t.Fatal(err)
		}
		parties = append(parties, map[string]string{"entity_name": name, "entity_type": "LEGAL_PERSON", "party_role": role})
	}
	snapshot := func() string {
		t.Helper()
		var value string
		if err := db.Raw(`SELECT jsonb_agg(jsonb_build_array(entity_name,entity_type,entity_id,party_role,relation_depth,metadata,identity_type,identity_number_ciphertext,identity_number_digest,CASE WHEN party_role='opposing_party' THEN '' ELSE aliases END) ORDER BY entity_name)::text FROM case_intake_parties WHERE intake_id=?`, f.intakeID).Scan(&value).Error; err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := snapshot()
	parties[0]["aliases"] = "new alias"
	w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, map[string]any{"expected_revision": f.revision, "parties": parties})
	if w.Code != 200 {
		t.Fatalf("full save returned %d", w.Code)
	}
	var alias string
	if err := db.Raw("SELECT aliases FROM case_intake_parties WHERE intake_id=? AND party_role='opposing_party'", f.intakeID).Scan(&alias).Error; err != nil || alias != "new alias" {
		t.Fatal("explicit alias was not applied", err)
	}
	if snapshot() != before {
		t.Fatal("an unrelated party fact or protected identity changed")
	}
}
