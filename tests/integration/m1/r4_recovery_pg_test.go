//go:build np_m1_pg

package m1_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"law-oa-go/internal/security"
)

// Exercise an unchanged legacy JSON echo. UseNumber keeps the test client lossless.
func TestM1R4LegacyMetadataEcho(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	c, d, err := security.ProtectIdentityNumber("SYNTHETIC-FIXTURE")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest,metadata)
 VALUES(?,'Fixture peer','LEGAL_PERSON','opponent','SOCIAL_CREDIT_CODE',?,?,'{"opaque":9007199254740993,"extra":{"note":"keep","weights":[9007199254740995]}}'::jsonb)`, f.intakeID, c, d).Error; err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/case-intakes/" + f.intakeID
	w := send("GET", path, nil)
	var detail struct {
		Data struct {
			Parties  []map[string]any
			Revision string
		}
	}
	dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	dec.UseNumber()
	if err := dec.Decode(&detail); err != nil || w.Code != 200 || len(detail.Data.Parties) != 1 {
		t.Fatal("fixture GET", w.Code, err)
	}
	party := detail.Data.Parties[0]
	meta := party["metadata"].(map[string]any)
	if n, ok := meta["opaque"].(json.Number); !ok || n.String() != "9007199254740993" {
		t.Fatalf("GET must preserve the JSON number; got %v", meta["opaque"])
	}
	row := party
	put := send("PUT", path, map[string]any{"expected_revision": detail.Data.Revision, "parties": []any{row}})
	var after string
	if err := db.Raw("SELECT metadata->>'opaque' FROM case_intake_parties WHERE intake_id=?", f.intakeID).Scan(&after).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("seed=9007199254740993 returned=%v PUT=%d persisted=%s", meta["opaque"], put.Code, after)
	if put.Code != 200 || after != "9007199254740993" {
		t.Fatal("unchanged legacy echo did not preserve metadata")
	}
	// The browser parses JSON numbers as doubles before stringifying the full form.
	w = send("GET", path, nil)
	var browser struct {
		Data struct {
			Revision string
			Parties  []map[string]any
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &browser); w.Code != 200 || err != nil || len(browser.Data.Parties) != 1 {
		t.Fatal("browser GET", w.Code, err)
	}
	echo := browser.Data.Parties[0]["metadata"].(map[string]any)
	echo["extra"].(map[string]any)["note"] = "new note"
	echo["party_type"] = "updated role"
	row = browser.Data.Parties[0]
	put = send("PUT", path, map[string]any{"expected_revision": browser.Data.Revision, "metadata": map[string]any{"precision": json.Number("9007199254740992")}, "parties": []any{row}})
	var ok bool
	if err := db.Raw(`SELECT metadata->>'opaque'='9007199254740993' AND metadata#>>'{extra,weights,0}'='9007199254740995'
        AND metadata#>>'{extra,note}'='new note' AND (SELECT metadata->>'precision' FROM case_intakes WHERE id=?)='9007199254740993'
 AND metadata->>'party_type'='updated role' FROM case_intake_parties WHERE intake_id=?`, f.intakeID, f.intakeID).Scan(&ok).Error; err != nil {
		t.Fatal(err)
	}
	if put.Code != 200 || !ok {
		t.Fatalf("legacy numeric echo plus legitimate edit: status=%d preserved=%t", put.Code, ok)
	}
	rev, _, _, _ := conflictPGIntakeState(t, db, f.intakeID)
	echo["opaque"] = json.Number("9007199254740995")
	put = send("PUT", path, map[string]any{"expected_revision": rev, "parties": []any{row}})
	if err := db.Raw("SELECT metadata->>'opaque' FROM case_intake_parties WHERE intake_id=?", f.intakeID).Scan(&after).Error; err != nil || put.Code != 200 || after != "9007199254740995" {
		t.Fatalf("explicit lossless edit: status=%d after=%s err=%v", put.Code, after, err)
	}
}

func TestM1R4NarrowSameIdentity(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	number := "SYNTHETIC123456"
	c, d, err := security.ProtectIdentityNumber(security.NormalizeIdentityNumber("SOCIAL_CREDIT_CODE", number))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest) VALUES(?,'Restored peer','LEGAL_PERSON','opposing_party','SOCIAL_CREDIT_CODE',?,?)", f.intakeID, c, d).Error; err != nil {
		t.Fatal(err)
	}
	w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, map[string]any{"expected_revision": f.revision, "opponent_identity_update": map[string]string{"entity_name": "Restored peer", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": number}})
	var eq struct{ Cipher, Digest bool }
	if err := db.Raw("SELECT identity_number_ciphertext=? AS cipher, identity_number_digest=? AS digest FROM case_intake_parties WHERE intake_id=?", c, d, f.intakeID).Scan(&eq).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("HTTP=%d same cipher=%t same digest=%t", w.Code, eq.Cipher, eq.Digest)
	if w.Code != 200 || !eq.Cipher || !eq.Digest {
		t.Fatal("re-entered identical identity changed its protected pair")
	}
}
