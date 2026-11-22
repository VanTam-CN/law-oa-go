//go:build np_m1_pg

package m1_test

import (
	"context"
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
			if err := db.Exec("INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest,aliases,metadata) VALUES(?,'Sample opponent','LEGAL_PERSON',?,'SOCIAL_CREDIT_CODE',?,?,'Sample Alias','{\"preserve_me\":true}')", f.intakeID, role, c, d).Error; err != nil {
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
			w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, map[string]any{"expected_revision": f.revision, "parties": []any{map[string]string{"entity_name": "Sample opponent", "party_role": "opposing_party", "entity_type": "LEGAL_PERSON", "identity_type": "SOCIAL_CREDIT_CODE", "identity_number": "SYNTHETIC-OPPONENT", "aliases": "New Alias"}}})
			after, kept := read()
			t.Logf("http=%d same_row=%v opaque_preserved=%v", w.Code, before == after, kept)
			var equal struct{ Cipher, Digest bool }
			if err := db.Raw("SELECT identity_number_ciphertext=? AS cipher,identity_number_digest=? AS digest FROM case_intake_parties WHERE intake_id=?", c, d, f.intakeID).Scan(&equal).Error; err != nil {
				t.Fatal(err)
			}
			t.Logf("same_cipher=%v same_digest=%v", equal.Cipher, equal.Digest)
			if w.Code != 200 || after != before || !kept || !equal.Digest || !equal.Cipher {
				t.Fatal("same identity echo changed guarded columns")
			}
		})
	}
}
