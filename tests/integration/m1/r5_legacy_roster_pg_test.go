//go:build np_m1_pg

package m1_test

import (
	"encoding/json"
	"testing"

	"law-oa-go/internal/security"
)

func TestM1R5TruncatedLegacyRoster(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	p := "/api/v1/case-intakes/" + f.intakeID
	c, d, e := security.ProtectIdentityNumber("SYNTHETIC-ROSTER")
	if e != nil {
		t.Fatal(e)
	}
	if e = db.Exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest,metadata)
 SELECT ?, 'Fixture-'||n,'LEGAL_PERSON','related_party','SOCIAL_CREDIT_CODE',?,?,'{}'::jsonb FROM generate_series(1,21) n`, f.intakeID, c, d).Error; e != nil {
		t.Fatal(e)
	}
	var detail struct {
		Data struct {
			Revision string
			Parties  []map[string]any
		}
	}
	w := send("GET", p, nil)
	if e = json.Unmarshal(w.Body.Bytes(), &detail); e != nil || w.Code != 200 || len(detail.Data.Parties) != 20 {
		t.Fatal("detail", w.Code, e)
	}
	snapshot := func() string {
		t.Helper()
		var state string
		if err := db.Raw(`SELECT jsonb_build_array(to_jsonb(i),(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM case_intake_parties p WHERE intake_id=i.id))::text
            FROM case_intakes i WHERE i.id=?`, f.intakeID).Scan(&state).Error; err != nil {
			t.Fatal(err)
		}
		return state
	}
	before := snapshot()
	w = send("PUT", p, map[string]any{"expected_revision": detail.Data.Revision, "parties": detail.Data.Parties})
	if w.Code != 409 || snapshot() != before {
		t.Fatalf("truncated roster must reject atomically, HTTP %d", w.Code)
	}
	// A caller that really supplies every stored identity can still update this draft.
	wire := append(detail.Data.Parties, map[string]any{"entity_name": "Fixture-21", "entity_type": "LEGAL_PERSON", "party_role": "related_party"})
	w = send("PUT", p, map[string]any{"expected_revision": detail.Data.Revision, "parties": wire})
	var count int64
	if err := db.Table("case_intake_parties").Where("intake_id=?", f.intakeID).Count(&count).Error; err != nil || w.Code != 200 || count != 21 {
		t.Fatalf("full roster HTTP %d, count %d, err=%v", w.Code, count, err)
	}

}
