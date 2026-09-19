package handlers

import "testing"

// The intake approval gate compares req metadata against the intake record.
// A missing metadata key must read as empty, not the literal "<nil>", so the
// fallback to conflict_result.checkId actually runs. Before this fix a payload
// without top-level conflict_check_id always failed with
// INTAKE_CONFLICT_LINK_MISMATCH even when the nested result carried the exact
// same check ID.
func TestIntegrationMetadataStringMissingKeyIsEmpty(t *testing.T) {
	if got := integrationMetadataString(nil, "conflict_check_id"); got != "" {
		t.Fatalf("nil metadata: got %q, want empty", got)
	}
	metadata := map[string]interface{}{"other": "x"}
	if got := integrationMetadataString(metadata, "conflict_check_id"); got != "" {
		t.Fatalf("missing key: got %q, want empty", got)
	}
	metadata["conflict_check_id"] = nil
	if got := integrationMetadataString(metadata, "conflict_check_id"); got != "" {
		t.Fatalf("nil value: got %q, want empty", got)
	}
	metadata["conflict_check_id"] = " CCT_test "
	if got := integrationMetadataString(metadata, "conflict_check_id"); got != "CCT_test" {
		t.Fatalf("string value: got %q, want trimmed CCT_test", got)
	}
}

// The gate falls back to the nested conflict_result.checkId when the payload
// omits conflict_check_id. Both sources must resolve to the same ID.
func TestIntakeApprovalCheckIDFallback(t *testing.T) {
	result := map[string]interface{}{"checkId": "CCT_abc123"}
	metadata := map[string]interface{}{"intake_id": "intake-1", "conflict_result": result}
	checkID := integrationMetadataString(metadata, "conflict_check_id")
	if checkID == "" {
		if nested, ok := metadata["conflict_result"].(map[string]interface{}); ok {
			checkID = firstIntegrationMetadataString(nested, "checkId", "check_id", "id")
		}
	}
	if checkID != "CCT_abc123" {
		t.Fatalf("fallback checkID = %q, want CCT_abc123", checkID)
	}
	intake := map[string]interface{}{"metadata": map[string]interface{}{"conflict_check_id": "CCT_abc123"}}
	if integrationMetadataString(integrationJSONMap(intake["metadata"]), "conflict_check_id") != checkID {
		t.Fatalf("intake metadata check id mismatch against %q", checkID)
	}
}
