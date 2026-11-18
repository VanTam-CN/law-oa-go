package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// The frontend persists a related party's business role inside the party row
// metadata (party_type) and restores drafts from it. If the row serializer
// ever drops or stringifies that field incorrectly, every restored role
// degrades back to a pending state, so the round trip must keep it intact.
func TestPrepareCaseIntakePartyRowPreservesRoleMetadata(t *testing.T) {
	t.Setenv("SUBJECT_DATA_KEY", strings.Repeat("g", 32))
	now := time.Now()
	row, err := prepareCaseIntakePartyRow(map[string]interface{}{
		"entity_name":     "Guarantor Co",
		"entity_type":     "LEGAL_PERSON",
		"party_role":      "related_party",
		"relation_depth":  2,
		"metadata":        map[string]interface{}{"party_type": "保证人"},
		"identity_type":   "SOCIAL_CREDIT_CODE",
		"identity_number": "91310000MA1FL0000X",
	}, "intake-1", now)
	if err != nil {
		t.Fatalf("prepareCaseIntakePartyRow returned error: %v", err)
	}
	metadata, ok := row["metadata"].(string)
	if !ok {
		t.Fatalf("party metadata must serialize to a JSON string, got %T", row["metadata"])
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(metadata), &decoded); err != nil {
		t.Fatalf("party metadata is not decodable JSON: %v", err)
	}
	if decoded["party_type"] != "保证人" {
		t.Fatalf("party_type lost during serialization: %v", decoded)
	}
	if row["relation_depth"] != 2 {
		t.Fatalf("relation_depth should stay a passthrough column, got %v", row["relation_depth"])
	}
	if strings.Contains(fmt.Sprintf("%v", row), "91310000MA1FL0000X") {
		t.Fatalf("plaintext identity must not remain in the row map: %v", row)
	}
}
