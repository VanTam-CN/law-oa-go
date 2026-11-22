package handlers

import (
	"encoding/json"
	"testing"
)

// PostgreSQL returns a jsonb metadata column to a GORM map scan as raw JSON
// text, and SQLite stores the same column as TEXT, so the practical read
// shape is a string, not a decoded map. These cases pin the column-level
// entry point: decode-then-redact keeps legal business values at their
// decoded precision, while unknown or undecodable shapes close with a fixed
// placeholder instead of echoing the raw bytes back to the client.
func TestAggregateMetadataStringFormIsRedacted(t *testing.T) {
	row := map[string]interface{}{
		"id":       "intake-redact",
		"title":    "脱敏回归草稿",
		"metadata": `{"source":"batch01_real_api","dispute_amount":1234567.89,"identity_number":"91310000MA1FL0000X","nested":{"identity_number_digest":"abc","note":"保留"}}`,
	}
	sanitized := sanitizeAggregateRow("case_intakes", row)
	metadata, ok := sanitized["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata must decode into a redacted object, got %T: %v", sanitized["metadata"], sanitized["metadata"])
	}
	if metadata["source"] != "batch01_real_api" {
		t.Fatalf("legal metadata keys must survive, got %v", metadata)
	}
	if wire, err := json.Marshal(metadata["dispute_amount"]); err != nil || string(wire) != "1234567.89" {
		t.Fatalf("business value precision must be preserved, got %T %v", metadata["dispute_amount"], metadata["dispute_amount"])
	}
	if _, leak := metadata["identity_number"]; leak {
		t.Fatalf("identity_number must be dropped at the top level: %v", metadata)
	}
	nested, _ := metadata["nested"].(map[string]interface{})
	if nested == nil {
		t.Fatalf("nested object must survive redaction: %v", metadata)
	}
	if _, leak := nested["identity_number_digest"]; leak {
		t.Fatalf("nested digest must be dropped: %v", nested)
	}
	if nested["note"] != "保留" {
		t.Fatalf("nested business value must survive, got %v", nested["note"])
	}
}

func TestAggregateMetadataUndecodableAndByteFormsClose(t *testing.T) {
	garbage := sanitizeAggregateRow("case_intakes", map[string]interface{}{
		"metadata": "not-json{",
	})
	if garbage["metadata"] != "已脱敏" {
		t.Fatalf("undecodable metadata must close with the placeholder, got %v", garbage["metadata"])
	}
	decoded := sanitizeAggregateRow("case_intakes", map[string]interface{}{
		"metadata": []byte(`{"identity_number":"secret","keep":1}`),
	})
	metadata, ok := decoded["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("[]byte metadata must decode and redact, got %T", decoded["metadata"])
	}
	if _, leak := metadata["identity_number"]; leak {
		t.Fatalf("[]byte form must drop identity_number: %v", metadata)
	}
	if wire, err := json.Marshal(metadata["keep"]); err != nil || string(wire) != "1" {
		t.Fatalf("[]byte form must keep business values, got %v", metadata["keep"])
	}
	if got := sanitizeAggregateRow("case_intakes", map[string]interface{}{"metadata": nil})["metadata"]; got != nil {
		t.Fatalf("nil metadata must stay nil, got %v", got)
	}
}
