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
		{"append", "[9007199254740993,2]", "[9007199254740992,2,7]", "[9007199254740993,2,7]", true},
		{"prepend", "[9007199254740993,2]", "[7,9007199254740992,2]", "[7,9007199254740993,2]", false},
		{"reorder", "[9007199254740993,2]", "[2,9007199254740992]", "[2,9007199254740993]", false},
		{"remove_first", "[2,9007199254740993]", "[9007199254740992]", "[9007199254740993]", false},
		{"ordinary_insert", "[1,2]", "[1,7,2]", "[1,7,2]", true},
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
			var result struct{ Exact, Unchanged bool }
			if err := db.Raw("SELECT metadata->'numbers'=?::jsonb AS exact,to_jsonb(i)=?::jsonb AS unchanged FROM case_intakes i WHERE id=?", tc.expected, before, f.intakeID).Scan(&result).Error; err != nil {
				t.Fatal(err)
			}
			if w.Code == 200 && result.Exact {
				return
			}
			if !tc.mustAccept && w.Code == 409 && result.Unchanged {
				return
			}
			t.Fatalf("status=%d exact=%v unchanged=%v", w.Code, result.Exact, result.Unchanged)
		})
	}
}
