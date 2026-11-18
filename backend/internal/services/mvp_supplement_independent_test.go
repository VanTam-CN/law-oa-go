package services

import (
	"encoding/json"
	"strings"
	"testing"

	"law-oa-go/internal/models"
)

// A mixed checklist must retain the same material anchor when an internal
// caller passes the validated typed items through the public contract again.
func TestIndependentMixedChecklistRetainsTypedMaterialAnchor(t *testing.T) {
	items := []SupplementItem{
		{ItemID: "summary-fix", Category: "field_correction", FieldKey: "case_summary", Description: "补充演示争议焦点", Required: true},
		{ItemID: "contract-fix", Category: "material_request", FieldKey: "material", MaterialID: 200, Description: "更换演示合同扫描件", Required: true},
	}
	validated, err := ValidateSupplementItems(map[string]interface{}{"schema_version": 1, "items": items})
	if err != nil {
		t.Fatalf("mixed typed checklist rejected: %v", err)
	}
	if len(validated) != 2 || validated[1].MaterialID != 200 {
		t.Fatalf("original material anchor lost: %+v", validated)
	}
}

func TestIndependentRoundDiffReportsSameCountPartyChangeWithoutIdentityLeak(t *testing.T) {
	db := seedContinuationWorld(t)
	beforeDigest, afterDigest := strings.Repeat("c", 64), strings.Repeat("d", 64)
	previous := map[string]interface{}{
		"client_id": "9",
		"structured_facts": map[string]interface{}{
			"title": "虚构争议", "case_type": "commercial", "priority": "medium", "description": "虚构摘要",
			"parties": []map[string]interface{}{{"entity_name": "虚构原对手", "entity_type": "COMPANY", "party_role": "opposing_party", "identity_type": "SOCIAL_CREDIT_CODE", "identity_digest": beforeDigest}},
		},
		"material_fingerprints": []interface{}{},
	}
	previousJSON, err := json.Marshal(previous)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO approval_snapshots (id, approval_request_id, snapshot_type, snapshot_data, source_version) VALUES (?, ?, ?, ?, ?)",
		"independent-diff-previous", "independent-diff-approval", models.ApprovalSnapshotTypeIntakeSubmission, string(previousJSON), 1).Error; err != nil {
		t.Fatal(err)
	}
	current := &frozenResubmitFacts{ClientID: "9", Facts: map[string]interface{}{
		"title": "虚构争议", "case_type": "commercial", "priority": "medium", "description": "虚构摘要",
		"parties": []map[string]interface{}{{"entity_name": "虚构新对手", "entity_type": "COMPANY", "party_role": "opposing_party", "identity_type": "SOCIAL_CREDIT_CODE", "identity_digest": afterDigest}},
	}}
	diff, err := buildIntakeRoundDiffTx(db, "independent-diff-previous", current, nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(diff)
	if err != nil {
		t.Fatal(err)
	}
	body := string(encoded)
	if !strings.Contains(body, "虚构原对手") || !strings.Contains(body, "虚构新对手") {
		t.Fatalf("same count is not unchanged: actual old/new parties must be visible: %s", body)
	}
	if strings.Contains(body, `"field":"parties_count"`) {
		t.Fatalf("same party count must not be reported as a count change: %s", body)
	}
	for _, forbidden := range []string{beforeDigest, afterDigest, "identity_digest", "ciphertext"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("public round diff leaked protected identity material: %s", forbidden)
		}
	}
}
