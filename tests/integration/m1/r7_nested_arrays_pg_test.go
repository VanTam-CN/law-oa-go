//go:build np_m1_pg

package m1_test

import (
	"encoding/json"
	"strings"
	"testing"

	"law-oa-go/internal/security"
)

// Parse the actual GET with the same binary64 precision as a legacy browser.
// Read exact expected values independently from the pre-save PG snapshot.
func TestM1R7NestedArrayOrigin(t *testing.T) {
	db := isolatedDB(t)
	for _, entry := range []string{"intake", "party"} {
		for _, op := range []string{"drop_append", "drop_lossless", "echo", "append_one", "append_distinct_suffixes"} {
			t.Run(entry+"/"+op, func(t *testing.T) {
				original := "[[9007199254740993],[9007199254740992,2]]"
				if op == "drop_lossless" {
					original = "[[9007199254740993],[9007199254740989,2]]"
				}
				f := seedConflictPGFixture(t, db, false, "")
				send := nativePartyRequest(t, db, f)
				path := "/api/v1/case-intakes/" + f.intakeID
				cipher, digest, err := security.ProtectIdentityNumber("SYNTHETIC-R7")
				if err != nil {
					t.Fatal(err)
				}
				if err = db.Exec(`INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role,identity_type,identity_number_ciphertext,identity_number_digest,metadata)
     VALUES (?,'Synthetic peer','LEGAL_PERSON','opponent','SOCIAL_CREDIT_CODE',?,?,jsonb_build_object('numbers',?::jsonb))`, f.intakeID, cipher, digest, original).Error; err != nil {
					t.Fatal(err)
				}
				if err = db.Exec("UPDATE case_intakes SET metadata=metadata||jsonb_build_object('numbers',?::jsonb) WHERE id=?", original, f.intakeID).Error; err != nil {
					t.Fatal(err)
				}
				type snapshot struct{ Rows, Revision, Intake, Party string }
				read := func() snapshot {
					t.Helper()
					var s snapshot
					if err := db.Raw(`SELECT jsonb_build_array(to_jsonb(i),to_jsonb(p))::text AS rows,
      i.metadata->>'draft_revision' AS revision, i.metadata->'numbers' AS intake, p.metadata->'numbers' AS party
      FROM case_intakes i JOIN case_intake_parties p ON p.intake_id=i.id WHERE i.id=?`, f.intakeID).Scan(&s).Error; err != nil || s.Revision == "" {
						t.Fatalf("snapshot: %v", err)
					}
					return s
				}
				before := read()
				get := send("GET", path, nil)
				var detail struct {
					Data struct {
						Revision string
						Intake   map[string]any
						Parties  []map[string]any
					}
				}
				if err := json.Unmarshal(get.Body.Bytes(), &detail); err != nil || get.Code != 200 || len(detail.Data.Parties) != 1 || detail.Data.Revision != before.Revision {
					t.Fatalf("GET %d: %v", get.Code, err)
				}
				row := detail.Data.Intake
				prior := before.Intake
				if entry == "party" {
					row = detail.Data.Parties[0]
					prior = before.Party
				}
				meta := row["metadata"].(map[string]any)
				a := meta["numbers"].([]any)
				// Compute the expected result in PG from its snapshot, before PUT.
				expectedSQL := "a"
				switch op {
				case "drop_append":
					meta["numbers"] = append(a[1:], []any{float64(9007199254740992), 2, 9})
					expectedSQL = "(a - 0) || '[[9007199254740992,2,9]]'::jsonb"
				case "drop_lossless":
					meta["numbers"] = append(a[1:], []any{7, 9})
					expectedSQL = "(a - 0) || '[[7,9]]'::jsonb"
				case "append_one":
					a[0] = append(a[0].([]any), 7)
					expectedSQL = "jsonb_set(a,'{0}',(a->0)||'[7]'::jsonb)"
				case "append_distinct_suffixes":
					a[0] = append(a[0].([]any), 7)
					a[1] = append(a[1].([]any), 9)
					expectedSQL = "jsonb_build_array((a->0)||'[7]'::jsonb,(a->1)||'[9]'::jsonb)"
				}
				var want string
				if err := db.Raw("SELECT ("+expectedSQL+")::text FROM (SELECT ?::jsonb a) s", prior).Scan(&want).Error; err != nil {
					t.Fatal(err)
				}
				body := map[string]any{"expected_revision": before.Revision, "title": "synthetic sibling edit"}
				if entry == "party" {
					body["parties"] = detail.Data.Parties
				} else {
					body["metadata"] = map[string]any{"numbers": meta["numbers"]}
				}
				put := send("PUT", path, body)
				after := read()
				actual, otherUnchanged := after.Intake, after.Party == before.Party
				if entry == "party" {
					actual, otherUnchanged = after.Party, after.Intake == before.Intake
				}
				sameRows, sameRevision := after.Rows == before.Rows, after.Revision == before.Revision
				t.Logf("HTTP=%d prior=%s expected=%s actual=%s same_revision=%t same_rows=%t", put.Code, prior, want, actual, sameRevision, sameRows)
				if put.Code == 200 && actual == want && otherUnchanged && !sameRevision {
					return
				}
				if op == "drop_append" && put.Code == 409 && sameRows && sameRevision && strings.Contains(put.Body.String(), "INTAKE_METADATA_ARRAY_AMBIGUOUS") {
					return
				}
				t.Fatal("unexpected save outcome")
			})
		}
	}
}
