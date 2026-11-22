//go:build np_m1_pg

package m1_test

import (
	"context"
	"encoding/json"
	"fmt"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
	"testing"
)

func TestM1R3IndexIncomingAlias(t *testing.T) {
	db := isolatedDB(t)
	hist := seedConflictPGFixture(t, db, false, "")
	active := seedConflictPGFixture(t, db, false, "")
	name := "Historical " + hist.intakeID
	if err := db.Table("clients").Where("id=?", hist.clientID).Update("name", name).Error; err != nil {
		t.Fatal(err)
	}
	idx := repositories.NewConflictRepository(db, nil).(repositories.ConflictP0SubjectIndexer)
	ctx := context.Background()
	if err := idx.SyncConflictSubjectIndex(ctx); err != nil {
		t.Fatal(err)
	}
	search := func(primary string, alias []string) bool {
		t.Helper()
		hits, err := idx.SearchConflictSubjectIndex(ctx, []models.ConflictNormalizedSubject{{OriginalName: primary, EntityType: "COMPANY", Role: "OPPOSING_PARTY", Aliases: alias}}, active.clientID)
		if err != nil {
			t.Fatal(err)
		}
		for _, h := range hits {
			if h.Version.ClientID == hist.clientID {
				return true
			}
		}
		return false
	}
	if !search(name, nil) {
		t.Fatal("positive control failed")
	}
	if search("Independent Current Brand", nil) {
		t.Fatal("negative control failed")
	}
	if !search("Independent Current Brand", []string{name}) {
		t.Fatal("main name matched, incoming alias alone failed to select same historical client")
	}
}
func TestM1R3SameIdentityEcho(t *testing.T) {
	for _, role := range []string{"opponent", "opposing_party"} {
		t.Run(role, func(t *testing.T) {
			db := isolatedDB(t)
			f := seedConflictPGFixture(t, db, false, "")
			send := nativePartyRequest(t, db, f)
			c, d, err := security.ProtectIdentityNumber(security.NormalizeIdentityNumber("SOCIAL_CREDIT_CODE", "SYNTHETIC-OPPONENT"))
			if err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest,aliases,metadata) VALUES(?,'Sample opponent','LEGAL_PERSON',?,'SOCIAL_CREDIT_CODE',?,?,'Sample Alias','{\"preserve_me\":9007199254740993}')", f.intakeID, role, c, d).Error; err != nil {
				t.Fatal(err)
			}
			read := func() (int, bool) {
				var r struct {
					ID   int
					Kept bool
				}
				if err := db.Raw("SELECT id, (metadata->>'preserve_me') IS NOT NULL AS kept FROM case_intake_parties WHERE intake_id=?", f.intakeID).Scan(&r).Error; err != nil {
					t.Fatal(err)
				}
				return r.ID, r.Kept
			}
			before, _ := read()
			// A dependent row makes a delete/reinsert with the same ID observable.
			pin := fmt.Sprintf("m1_test_party_pin_%d", before)
			if err := db.Exec("CREATE TABLE " + pin + " (party bigint REFERENCES case_intake_parties(id))").Error; err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := db.Exec("DROP TABLE " + pin).Error; err != nil {
					t.Error(err)
				}
			})
			if err := db.Exec("INSERT INTO "+pin+" VALUES (?)", before).Error; err != nil {
				t.Fatal(err)
			}
			snapshot := func() string {
				t.Helper()
				var raw string
				if err := db.Raw("SELECT (to_jsonb(p)-'aliases')::text FROM case_intake_parties p WHERE id=?", before).Scan(&raw).Error; err != nil {
					t.Fatal(err)
				}
				return raw
			}
			original := snapshot()
			w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, map[string]any{"expected_revision": f.revision, "parties": []any{map[string]string{"entity_name": "Sample opponent", "party_role": "opposing_party", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": "SYNTHETIC-OPPONENT", "aliases": "New Alias"}}})
			after, kept := read()
			t.Logf("http=%d same_row=%v opaque_preserved=%v", w.Code, before == after, kept)
			var equal struct{ Cipher, Digest, Alias bool }
			if err := db.Raw("SELECT identity_number_ciphertext=? AS cipher,identity_number_digest=? AS digest, aliases=? AS alias FROM case_intake_parties WHERE intake_id=?", c, d, "New Alias", f.intakeID).Scan(&equal).Error; err != nil {
				t.Fatal(err)
			}
			t.Logf("same_cipher=%v same_digest=%v", equal.Cipher, equal.Digest)
			if w.Code != 200 || after != before || !kept || !equal.Digest || !equal.Cipher || !equal.Alias || snapshot() != original {
				t.Fatal("same identity echo changed guarded columns")
			}
		})
	}
}

func TestM1R3NativeCheckUsesEditedAlias(t *testing.T) {
	db := isolatedDB(t)
	f := seedConflictPGFixture(t, db, true, "")
	hist := seedConflictPGFixture(t, db, false, "")
	send := nativePartyRequest(t, db, f)
	name := "Archive " + hist.intakeID
	exec := func(q string, args ...any) {
		t.Helper()
		if err := db.Exec(q, args...).Error; err != nil {
			t.Fatal(err)
		}
	}
	exec("UPDATE clients SET name=? WHERE id=?", name, hist.clientID)
	cipher, digest, err := security.ProtectIdentityNumber("CLIENT-" + f.intakeID)
	if err != nil {
		t.Fatal(err)
	}
	exec("UPDATE clients SET identity_type='SOCIAL_CREDIT_CODE',identity_number_ciphertext=?,identity_number_digest=? WHERE id=?", cipher, digest, f.clientID)
	exec("UPDATE case_intakes SET metadata=metadata||jsonb_build_object('subject_case_id',?::text) WHERE id=?", f.caseID, f.intakeID)
	cipher, digest, err = security.ProtectIdentityNumber("OPPONENT-" + f.intakeID)
	if err != nil {
		t.Fatal(err)
	}
	exec("INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest,aliases) VALUES(?,'Separate current brand','LEGAL_PERSON','opposing_party','SOCIAL_CREDIT_CODE',?,?,?)", f.intakeID, cipher, digest, "Synthetic other alias、"+name)
	path := "/api/v1/case-intakes/" + f.intakeID
	for _, expected := range []bool{true, false} {
		if !expected {
			rev, _, _, _ := conflictPGIntakeState(t, db, f.intakeID)
			if w := send("PUT", path, map[string]any{"expected_revision": rev, "opponent_identity_update": map[string]string{"entity_name": "Separate current brand", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": "OPPONENT-" + f.intakeID, "aliases": ""}}); w.Code != 200 {
				t.Fatalf("clear alias: HTTP %d", w.Code)
			}
			if w := send("POST", path+"/facts-confirmation", map[string]any{}); w.Code != 200 {
				t.Fatalf("confirm: HTTP %d", w.Code)
			}
		}
		w := send("POST", path+"/conflict-check", map[string]any{})
		var res struct {
			Data struct{ Result models.ConflictCheckResponse }
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); w.Code != 200 || err != nil {
			t.Fatalf("native check HTTP %d, %v", w.Code, err)
		}
		// The lawyer receives a redacted result. Follow the returned check ID
		// to the persisted evidence that the independent reviewer reads.
		var raw string
		if err := db.Table("conflict_check_records").Select("check_result").Where("check_id = ?", res.Data.Result.CheckID).Scan(&raw).Error; err != nil {
			t.Fatal(err)
		}
		var stored models.ConflictCheckResponse
		if err := json.Unmarshal([]byte(raw), &stored); err != nil {
			t.Fatal(err)
		}
		var hits int
		for _, c := range stored.ConflictCases {
			if c.ClientID == hist.clientID {
				for _, e := range c.Evidence {
					if e.RequestedParty == name && e.MatchType == "CANDIDATE" {
						hits++
					}
				}
			}
		}
		if (hits > 0) != expected {
			t.Errorf("expected alias candidate=%v, matched evidence=%d", expected, hits)
		}
	}
}
