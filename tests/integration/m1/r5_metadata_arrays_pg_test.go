//go:build np_m1_pg

package m1_test

import (
	"encoding/json"
	"testing"
)

// A browser edit may move rounded values. A refusal must roll back the whole
// draft; an accepted edit must put each exact number in the intended position.
func TestM1R5ArrayEditPreservesOriginalNumbers(t *testing.T) {
	db := isolatedDB(t)
	for _, tc := range []struct {
		name, stored, wire, expected string
		mustAccept                   bool
	}{
		{"delete_first_then_append_collision", "[9007199254740993,9007199254740992,1]", "[9007199254740992,1,7]", "[9007199254740992,1,7]", false},
		{"nested_pure_append", "[[9007199254740993,2]]", "[[9007199254740992,2,7]]", "[[9007199254740993,2,7]]", true},
		{"delete_first_reverse_collision", "[9007199254740992,9007199254740993,1]", "[9007199254740992,1,7]", "[9007199254740993,1,7]", false},
		{"unchanged_collision_echo", "[9007199254740993,9007199254740992,1]", "[9007199254740992,9007199254740992,1]", "[9007199254740993,9007199254740992,1]", true},
		{"append", "[9007199254740993,2]", "[9007199254740992,2,7]", "[9007199254740993,2,7]", true},
		{"overflow_echo_append", "[1e400]", "[null,7]", "[1e400,7]", true},
		{"prepend", "[9007199254740993,2]", "[7,9007199254740992,2]", "[7,9007199254740993,2]", false},
		{"reorder", "[9007199254740993,2]", "[2,9007199254740992]", "[2,9007199254740993]", false},
		{"remove_first", "[2,9007199254740993]", "[9007199254740992]", "[9007199254740993]", false},
		{"safe_sibling_edit", "[9007199254740993,2]", "[9007199254740992,3]", "[9007199254740993,3]", true},
		{"ordinary_decimal_reorder", "[1.00,2]", "[2,1]", "[2,1]", true},
		{"ordinary_insert", "[1,2]", "[1,7,2]", "[1,7,2]", true},
		{"explicit_lossless_reorder", "[36028797018963969,36028797018963971,0]", "[0,36028797018963969,36028797018963971]", "[0,36028797018963969,36028797018963971]", true},
		{"explicit_lossless_edit", "[9007199254740993]", "[9007199254740995]", "[9007199254740995]", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := seedConflictPGFixture(t, db, false, "")
			send := nativePartyRequest(t, db, f)
			if err := db.Exec("UPDATE case_intakes SET metadata=metadata||jsonb_build_object('numbers',?::jsonb) WHERE id=?", tc.stored, f.intakeID).Error; err != nil {
				t.Fatal(err)
			}
			var before string
			if err := db.Raw("SELECT to_jsonb(i)::text FROM case_intakes i WHERE id=?", f.intakeID).Scan(&before).Error; err != nil {
				t.Fatal(err)
			}
			w := send("PUT", "/api/v1/case-intakes/"+f.intakeID, map[string]any{"expected_revision": f.revision, "title": "synthetic edit", "metadata": map[string]any{"numbers": json.RawMessage(tc.wire)}})
			var result struct {
				Exact, Unchanged, SameRevision bool
				Actual                         string
			}
			if err := db.Raw(`SELECT metadata->'numbers'=?::jsonb AS exact, metadata->'numbers' AS actual,
                to_jsonb(i)=?::jsonb AS unchanged,
                metadata->>'draft_revision'=(?::jsonb#>>'{metadata,draft_revision}') AS same_revision
                FROM case_intakes i WHERE id=?`, tc.expected, before, before, f.intakeID).Scan(&result).Error; err != nil {
				t.Fatal(err)
			}
			t.Logf("HTTP=%d stored=%s exact=%t same_row=%t same_revision=%t", w.Code, result.Actual, result.Exact, result.Unchanged, result.SameRevision)
			if w.Code == 200 && result.Exact && !result.SameRevision {
				return
			}
			if !tc.mustAccept && w.Code == 409 && result.Unchanged && result.SameRevision {
				return
			}
			t.Fatalf("status=%d exact=%v unchanged=%v", w.Code, result.Exact, result.Unchanged)
		})
	}
}
