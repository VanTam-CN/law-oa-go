//go:build np_m1_pg

package m1_test

import (
	"encoding/json"
	"testing"

	"law-oa-go/internal/security"
)

// Parse the actual GET with the same binary64 precision as a legacy browser.
// Read exact expected values independently from the pre-save PG snapshot.
func TestM1R7NestedArrayOrigin(t *testing.T) {
	db := isolatedDB(t)
	for _, entry := range []string{"intake", "party"} {
		for _, op := range []string{"drop_append", "drop_lossless", "drop_agree", "decimal_append", "echo", "append_one", "append_distinct_suffixes",
			"swap_hidden", "swap_hidden_text", "mixed_prefix", "mixed_suffix", "nested_hidden_echo", "nested_hidden_edit", "nested_hidden_sibling",
			"rotate_append", "rotate_exact", "cross_objects", "swap_visible_siblings", "cut_hidden", "cut_visible", "delete_hidden", "cross_hidden_child", "delete_with_empty_peer"} {
			t.Run(entry+"/"+op, func(t *testing.T) {
				original := "[[9007199254740993],[9007199254740992,2]]"
				switch op {
				case "nested_hidden_echo", "nested_hidden_edit", "nested_hidden_sibling":
					original = `[{"tag":"a","external_identifier":"outer","child":{"label":"k","external_identifier":"inner"}},1]`
				case "rotate_append":
					original = `[[9007199254740993,1],[9007199254740993,2],0]`
				case "rotate_exact":
					original = `[[11,1],[11,2],0]`
				case "cross_hidden_child":
					original = `[[0,{"tag":"a","child":{"external_identifier":"h"}}],[1,{"tag":"b"}]]`
				case "delete_with_empty_peer":
					original = `[[0,{"tag":"a","child":{"external_identifier":"h"}}],[1,{"tag":"b","child":{}}]]`
				case "cut_visible":
					original = `[[0,{"tag":"a"}],[1,{"tag":"b"}]]`
				case "cross_objects", "cut_hidden", "delete_hidden":
					original = `[[0,{"tag":"a","external_identifier":"h"}],[1,{"tag":"b"}]]`
				case "swap_visible_siblings":
					original = `[{"tag":"a","external_identifier":"outer","child":{"external_identifier":"inner"}},{"tag":"b"},{"tag":"c"}]`
				case "swap_hidden":
					original = `[9007199254740993,{"tag":"a","external_identifier":9007199254740993},{"tag":"b"}]`
				case "swap_hidden_text":
					original = `[0,{"tag":"a","external_identifier":"synthetic-hidden"},{"tag":"b"}]`
				case "mixed_prefix":
					original = `[{"fixed":9007199254740992},[9007199254740993,1],[9007199254740993,2]]`
				case "mixed_suffix":
					original = `[[9007199254740993,1],[9007199254740993,2],{"fixed":9007199254740992}]`
				case "drop_agree":
					original = "[[9007199254740993,1],[9007199254740993,2]]"
				case "decimal_append":
					original = "[[9007199254740993],[0.0000001],[0.0000001]]"
				case "drop_lossless":
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
				case "rotate_append", "rotate_exact":
					tail := append(a[0].([]any), 7)
					meta["numbers"] = append(a[1:], tail)
					expectedSQL = "jsonb_build_array(a->1,a->2,(a->0)||'[7]'::jsonb)"
				case "cut_hidden", "cut_visible", "delete_hidden", "delete_with_empty_peer":
					left, right := a[0].([]any), a[1].([]any)
					a[0] = left[:1]
					expectedSQL = "jsonb_build_array((a->0)-1,a->1)"
					if op != "delete_hidden" && op != "delete_with_empty_peer" {
						a[1] = append(right, left[1])
						expectedSQL = "jsonb_build_array((a->0)-1,(a->1)||jsonb_build_array(a#>'{0,1}'))"
					}
				case "cross_objects", "cross_hidden_child":
					left, right := a[0].([]any), a[1].([]any)
					left[1], right[1] = right[1], left[1]
				case "swap_visible_siblings":
					a[1], a[2] = a[2], a[1]
					expectedSQL = "jsonb_build_array(a->0,a->2,a->1)"
				case "nested_hidden_edit":
					a[0].(map[string]any)["tag"] = "changed"
					expectedSQL = `jsonb_set(a,'{0,tag}','"changed"'::jsonb)`
				case "nested_hidden_sibling":
					a[1] = 2
					expectedSQL = "jsonb_set(a,'{1}','2'::jsonb)"
				case "swap_hidden", "swap_hidden_text":
					if _, visible := a[1].(map[string]any)["external_identifier"]; visible {
						t.Fatal("GET exposed the hidden object field")
					}
					a[1], a[2] = a[2], a[1]
					expectedSQL = "a"
				case "mixed_prefix":
					a[1], a[2] = a[2], []any{7, 9}
					expectedSQL = "(a - 1) || '[[7,9]]'::jsonb"
				case "drop_append":
					meta["numbers"] = append(a[1:], []any{float64(9007199254740992), 2, 9})
					expectedSQL = "(a - 0) || '[[9007199254740992,2,9]]'::jsonb"
				case "decimal_append":
					a[2] = append(a[2].([]any), 7)
					expectedSQL = "jsonb_set(a,'{2}',(a->2)||'[7]'::jsonb)"
				case "drop_agree", "drop_lossless", "mixed_suffix":
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
				if op == "cross_hidden_child" || op == "cut_hidden" || op == "rotate_append" || op == "cross_objects" || (op == "nested_hidden_edit" && put.Code == 409) || op == "drop_append" || op == "swap_hidden" || op == "swap_hidden_text" {
					var rejected struct{ Error string }
					if put.Code != 409 || json.Unmarshal(put.Body.Bytes(), &rejected) != nil || rejected.Error != "INTAKE_METADATA_ARRAY_AMBIGUOUS" || !sameRows || !sameRevision {
						t.Fatal("ambiguous edit must leave both rows and the revision untouched")
					}
					return
				}
				if put.Code == 200 && actual == want && otherUnchanged && !sameRevision {
					return
				}
				t.Fatal("unexpected save outcome")
			})
		}
	}
}
