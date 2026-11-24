//go:build np_m1_pg

package m1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gorm.io/gorm"
	"law-oa-go/internal/repositories"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/config"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/router"
	"law-oa-go/internal/security"
)

func nativePartyRouter(t *testing.T, db *gorm.DB, f conflictPGFixture) (http.Handler, string) {
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
	return g, token
}

func nativePartyRequest(t *testing.T, db *gorm.DB, f conflictPGFixture) func(string, string, any) *httptest.ResponseRecorder {
	t.Helper()
	g, token := nativePartyRouter(t, db, f)
	return func(method, path string, input any) *httptest.ResponseRecorder {
		t.Helper()
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		if method == "POST" && strings.HasSuffix(path, "/facts-confirmation") {
			rev, _, _, _ := conflictPGIntakeState(t, db, f.intakeID)
			req.Header.Set("X-Intake-Revision", rev)
		}
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
	exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,metadata)
		SELECT ?, 'synthetic-' || n, 'LEGAL_PERSON', CASE WHEN n=1 THEN 'client' ELSE 'related_party' END, jsonb_build_object('opaque',n)
		FROM generate_series(1,23) AS n`, f.intakeID)
	exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,aliases,metadata,identity_type,identity_number_ciphertext,identity_number_digest)
		VALUES(?, 'Test opponent','LEGAL_PERSON','opposing_party','original alias','{"opaque":9007199254740993}','SOCIAL_CREDIT_CODE','old cipher','old digest')`, f.intakeID)
	exec(`INSERT INTO case_materials(intake_id,name,metadata) SELECT ?, 'material-' || n, jsonb_build_object('n',n) FROM generate_series(1,2) AS n`, f.intakeID)
	// Compare the entire stored rows, including IDs, unknown JSON and timestamps.
	snapshot := func() string {
		t.Helper()
		var value string
		err := db.Raw(`SELECT jsonb_build_array(
			(SELECT jsonb_agg(CASE WHEN party_role='opposing_party' THEN to_jsonb(p)-'aliases'-'identity_type'-'identity_number_ciphertext'-'identity_number_digest' ELSE to_jsonb(p) END ORDER BY id) FROM case_intake_parties p WHERE intake_id=?),
			(SELECT jsonb_agg(to_jsonb(m) ORDER BY id) FROM case_materials m WHERE intake_id=?)
		)::text`, f.intakeID, f.intakeID).Scan(&value).Error
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	detail := send("GET", "/api/v1/case-intakes/"+f.intakeID, nil)
	var projection struct {
		Data struct {
			Parties  []map[string]any
			Opponent map[string]any `json:"opponent_party"`
		}
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &projection); err != nil || detail.Code != 200 || len(projection.Data.Parties) != 20 || projection.Data.Opponent["entity_name"] != "Test opponent" {
		t.Fatal("missing unique opponent beyond capped detail", detail.Code, err)
	}
	for _, key := range []string{"identity_number", "identity_number_ciphertext", "identity_number_digest"} {
		if projection.Data.Opponent[key] != nil {
			t.Fatal("protected field in opponent projection", key)
		}
	}
	before := snapshot()
	const identity = "TEST-NEW-IDENTITY"
	w := send(http.MethodPut, "/api/v1/case-intakes/"+f.intakeID, map[string]any{
		"description": "reviewed native draft", "expected_revision": f.revision,
		"opponent_identity_update": map[string]string{"entity_name": "Test opponent", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": identity, "aliases": "restored alias"},
	})
	if w.Code != 200 {
		t.Fatalf("native save returned %d", w.Code)
	}
	var gotAlias string
	if err := db.Table("case_intake_parties").Select("aliases").Where("intake_id=? AND party_role='opposing_party'", f.intakeID).Scan(&gotAlias).Error; err != nil || gotAlias != "restored alias" {
		t.Fatal("restored alias was not persisted", err)
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

func legacyPartyEcho(t *testing.T, db *gorm.DB, f conflictPGFixture) ([]map[string]string, func() string) {
	t.Helper()
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
		if err := db.Raw(`SELECT jsonb_agg(CASE WHEN party_role='opposing_party' THEN to_jsonb(p)-'aliases' ELSE to_jsonb(p) END ORDER BY id)::text FROM case_intake_parties p WHERE intake_id=?`, f.intakeID).Scan(&value).Error; err != nil {
			t.Fatal(err)
		}
		return value
	}
	return parties, snapshot
}

func TestNativeFullSaveOnlyChangesExplicitAlias(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	parties, snapshot := legacyPartyEcho(t, db, f)
	var cname string
	if err := db.Table("clients").Select("name").Where("id=?", f.clientID).Scan(&cname).Error; err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"INITIAL-CLIENT-ID", "CORRECTED-CLIENT-ID"} {
		cipher, digest, err := security.ProtectIdentityNumber(v)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("UPDATE clients SET identity_type='SOCIAL_CREDIT_CODE', identity_number_ciphertext=?, identity_number_digest=? WHERE id=?", cipher, digest, f.clientID).Error; err != nil {
			t.Fatal(err)
		}
		if v == "INITIAL-CLIENT-ID" {
			if err := db.Exec("INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest) VALUES(?,?,'LEGAL_PERSON','client','SOCIAL_CREDIT_CODE',?,?)", f.intakeID, cname, cipher, digest).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	parties = append(parties, map[string]string{"entity_name": cname, "entity_type": "LEGAL_PERSON", "party_role": "client"})
	before := snapshot()
	revision := f.revision
	for _, want := range []string{"new alias", ""} {
		parties[0]["aliases"] = want
		body := map[string]any{"expected_revision": revision, "parties": parties}
		if want == "" {
			body["client_id"] = f.clientID
		}
		w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, body)
		if w.Code != 200 {
			t.Fatalf("full save returned %d", w.Code)
		}
		var alias string
		if err := db.Raw("SELECT aliases FROM case_intake_parties WHERE intake_id=? AND party_role='opposing_party'", f.intakeID).Scan(&alias).Error; err != nil || alias != want {
			t.Fatal("explicit alias was not applied", err)
		}
		if snapshot() != before {
			t.Fatal("an unrelated party fact or protected identity changed")
		}
		revision, _, _, _ = conflictPGIntakeState(t, db, f.intakeID)
	}
}

func TestNativeLegacyFullEchoPreservesRows(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	parties, snapshot := legacyPartyEcho(t, db, f)
	if err := db.Exec("UPDATE case_intake_parties SET relation_depth=3 WHERE intake_id=? AND party_role='related_party'", f.intakeID).Error; err != nil {
		t.Fatal(err)
	}
	beforeParties := snapshot()
	if err := db.Exec("INSERT INTO case_materials(intake_id,name,metadata) VALUES(?,'sample','{}')", f.intakeID).Error; err != nil {
		t.Fatal(err)
	}
	var before, after string
	materialSnapshot := func(dest *string) {
		if err := db.Raw("SELECT jsonb_agg(to_jsonb(m) ORDER BY id)::text FROM case_materials m WHERE intake_id=?", f.intakeID).Scan(dest).Error; err != nil {
			t.Fatal(err)
		}
	}
	materialSnapshot(&before)
	var wire []map[string]any
	for i, p := range parties {
		m := map[string]any{}
		for k, v := range p {
			m[k] = v
		}
		m["relation_depth"] = 1
		if i == 0 {
			m["relation_depth"] = 0
			m["aliases"] = "changed"
		}
		wire = append(wire, m)
	}
	w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, map[string]any{"expected_revision": f.revision, "parties": wire, "materials": []any{map[string]any{"name": "sample", "metadata": map[string]any{}}}})
	if w.Code != 200 {
		t.Fatalf("HTTP %d", w.Code)
	}
	materialSnapshot(&after)
	var depth int
	if err := db.Table("case_intake_parties").Select("relation_depth").Where("intake_id=? AND party_role='opposing_party'", f.intakeID).Scan(&depth).Error; err != nil {
		t.Fatal(err)
	}
	if depth != 1 || before != after || beforeParties != snapshot() {
		t.Fatalf("legacy echo changed stored rows: depth=%d material_rows_preserved=%v", depth, before == after)
	}
	rev, _, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	bad := send("PUT", "/api/v1/case-intakes/"+f.intakeID, map[string]any{"expected_revision": rev, "description": "must roll back", "materials": []any{map[string]any{"name": "different"}}})
	if bad.Code < 400 || bad.Code >= 500 {
		t.Fatalf("unsupported material edit: HTTP %d", bad.Code)
	}
	for _, props := range []map[string]any{{"materials": nil}, {"materials": []any{}}, {"material_changes": []any{}}, {"materials": []any{map[string]any{"name": "sample", "unknown": true}}}} {
		props["expected_revision"] = rev
		props["title"] = "must roll back"
		if w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, props); w.Code < 400 || w.Code >= 500 {
			t.Fatalf("invalid material change accepted: HTTP %d", w.Code)
		}
	}
	endRev, _, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	materialSnapshot(&after)
	if endRev != rev || before != after || beforeParties != snapshot() {
		t.Fatal("rejected material edit changed the draft")
	}
}

func TestNativeRecheckAfterSameMatterIsReconfirmed(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, true, "")
	if err := db.Exec("INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role) VALUES(?, 'Synthetic opponent', 'LEGAL_PERSON','opposing_party')", f.intakeID).Error; err != nil {
		t.Fatal(err)
	}
	send := nativePartyRequest(t, db, f)
	ctx := context.Background()
	link := func(id, rev string) error {
		insertConflictPGRecord(t, db, f, id, rev)
		return repositories.LinkConflictCheckToCase(ctx, db, repositories.ConflictSubjectAssociation{CheckID: id, SubjectCaseID: f.caseID, IntakeID: f.intakeID, IntakeRevision: rev, ClientID: f.clientID})
	}
	first := "m1a-" + f.intakeID
	if err := link(first, f.revision); err != nil {
		t.Fatal(err)
	}
	rev, _, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	path := "/api/v1/case-intakes/" + f.intakeID
	if w := send("PUT", path, map[string]any{"expected_revision": rev, "description": "revised synthetic summary"}); w.Code != 200 {
		t.Fatalf("native edit: %d", w.Code)
	}
	if w := send("POST", path+"/facts-confirmation", map[string]any{}); w.Code != 200 {
		t.Fatalf("native confirm: %d", w.Code)
	}
	rev, status, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	if err := link("m1b-"+f.intakeID, rev); err != nil {
		t.Fatalf("second completion after native edit and reconfirm: original pointer retained=%t; intake=%s; err=%v", conflictPGCaseCheckID(t, db, f.caseID) == first, status, err)
	}
	end, state, check, _ := conflictPGIntakeState(t, db, f.intakeID)
	if end == rev || state != "conflict_ready" || check != "m1b-"+f.intakeID || conflictPGCaseCheckID(t, db, f.caseID) != check {
		t.Fatal("recheck pointers or version inconsistent")
	}
	stale := repositories.ConflictSubjectAssociation{CheckID: first, SubjectCaseID: f.caseID, IntakeID: f.intakeID, IntakeRevision: f.revision, ClientID: f.clientID}
	if err := repositories.LinkConflictCheckToCase(ctx, db, stale); !errors.Is(err, repositories.ErrIntakeRevisionConflict) {
		t.Fatal("old delivery accepted", err)
	}
}
