package handlers

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// projectApprovalForViewer keeps the applicant's workflow data visible while
// removing historical conflict evidence from approval responses. The browser
// cannot be the disclosure boundary because the same response is also used by
// exports and other clients.
func projectApprovalForViewer(approval *models.ApprovalRequest, role string) *models.ApprovalRequest {
	if approval == nil {
		return approval
	}

	var projected models.ApprovalRequest
	raw, err := json.Marshal(projectApprovalResponseForViewer(approval, role))
	if err != nil || json.Unmarshal(raw, &projected) != nil {
		return nil
	}
	return &projected
}

// projectApprovalResponseForViewer is the shared success-response boundary for
// approval DTOs, lists and integrated results. It copies the payload, preserves
// JSON-string fields, and never changes persisted snapshots or caller objects.
func projectApprovalResponseForViewer(value interface{}, role string) interface{} {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil // Never fall back to an unprojected sensitive response.
	}
	var decoded interface{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&decoded) != nil {
		return nil
	}
	decoded = sanitizeSubmissionSecrets(decoded)
	if services.IsConflictReviewRole(role) {
		return decoded
	}
	return redactApprovalConflictPayload(decoded)
}

// projectApprovalSnapshotForViewer applies the same disclosure rule to the
// immutable snapshot endpoint. Conflict officers and authorized business
// management roles retain the frozen evidence needed for independent review.
func projectApprovalSnapshotForViewer(value interface{}, role string) interface{} {
	value, _ = decodeApprovalJSON(value)
	return projectApprovalResponseForViewer(value, role)
}

// redactApprovalConflictPayload recursively projects only the conflict
// evidence branches. Current-application fields remain available to the
// applicant so the workflow remains usable without exposing historical work.
func redactApprovalConflictPayload(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		projected := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			normalized := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
			switch normalized {
			case "metadata":
				if raw, ok := item.(string); ok && strings.TrimSpace(raw) != "" && !json.Valid([]byte(raw)) {
					projected[key] = "冲突证据详情仅向获授权独立冲突核查人展示。"
				} else {
					projected[key] = redactApprovalConflictPayload(item)
				}
			case "conflictcases", "conflict_cases":
				projected[key] = redactApprovalConflictCases(item)
			case "conflictresult", "conflict_result":
				projected[key] = redactApprovalConflictResult(item)
			case "checkresult", "check_result":
				projected[key] = redactApprovalConflictResult(item)
			case "evidence", "matchevidence", "match_evidence":
				projected[key] = redactApprovalEvidence(item)
			case "conflictrecord", "conflict_record":
				projected[key] = redactApprovalConflictRecord(item)
			default:
				projected[key] = redactApprovalConflictPayload(item)
			}
		}
		return projected
	case []interface{}:
		projected := make([]interface{}, len(typed))
		for index, item := range typed {
			projected[index] = redactApprovalConflictPayload(item)
		}
		return projected
	case []byte:
		var decoded interface{}
		if json.Unmarshal(typed, &decoded) == nil {
			return redactApprovalConflictPayload(decoded)
		}
		return string(typed)
	case string:
		var decoded interface{}
		if json.Unmarshal([]byte(typed), &decoded) == nil {
			switch decoded.(type) {
			case map[string]interface{}, []interface{}:
				projected := redactApprovalConflictPayload(decoded)
				if !reflect.DeepEqual(decoded, projected) {
					return jsonStringValue(projected)
				}
			}
		}
		return typed
	default:
		return value
	}
}

func redactApprovalConflictRecord(value interface{}) interface{} {
	decoded, wasJSON := decodeApprovalJSON(value)
	record, ok := decoded.(map[string]interface{})
	if !ok {
		return decoded
	}
	projected := redactApprovalConflictPayload(record).(map[string]interface{})
	for _, key := range []string{"riskLevel", "risk_level"} {
		if _, exists := projected[key]; exists {
			projected[key] = "REVIEW_REQUIRED"
		}
	}
	for _, key := range []string{"hasConflict", "has_conflict"} {
		if _, exists := projected[key]; exists {
			projected[key] = true
		}
	}
	if wasJSON {
		return jsonStringValue(projected)
	}
	return projected
}

func redactApprovalConflictCases(value interface{}) interface{} {
	decoded, wasJSON := decodeApprovalJSON(value)
	items, ok := decoded.([]interface{})
	if !ok {
		return decoded
	}
	projected := make([]interface{}, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			projected = append(projected, redactConflictQueueCase(row))
			continue
		}
		projected = append(projected, redactApprovalConflictPayload(item))
	}
	if wasJSON {
		return jsonStringValue(projected)
	}
	return projected
}

func redactApprovalConflictResult(value interface{}) interface{} {
	if raw, ok := value.(string); ok && strings.TrimSpace(raw) != "" && !json.Valid([]byte(raw)) {
		return "冲突证据详情仅向获授权独立冲突核查人展示。"
	}
	decoded, wasJSON := decodeApprovalJSON(value)
	if result, ok := decoded.(map[string]interface{}); ok {
		projected := redactConflictQueueCheckResult(result)
		if wasJSON {
			return jsonStringValue(projected)
		}
		return projected
	}
	return decoded
}

func redactApprovalEvidence(value interface{}) interface{} {
	decoded, wasJSON := decodeApprovalJSON(value)
	items, ok := decoded.([]interface{})
	if !ok {
		if row, rowOK := decoded.(map[string]interface{}); rowOK {
			projected := redactConflictQueueEvidence(row)
			if wasJSON {
				return jsonStringValue(projected)
			}
			return projected
		}
		return decoded
	}
	projected := make([]interface{}, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			projected = append(projected, redactConflictQueueEvidence(row))
		} else {
			projected = append(projected, redactApprovalConflictPayload(item))
		}
	}
	if wasJSON {
		return jsonStringValue(projected)
	}
	return projected
}

func decodeApprovalJSON(value interface{}) (interface{}, bool) {
	switch typed := value.(type) {
	case []byte:
		var decoded interface{}
		if json.Unmarshal(typed, &decoded) == nil {
			return decoded, true
		}
	case string:
		var decoded interface{}
		if json.Unmarshal([]byte(typed), &decoded) == nil {
			return decoded, true
		}
	}
	return value, false
}

// Protected identifiers remain on the server, including for authorized reviewers.
// The frozen audit uses digests/ciphertext; the UI needs only masked identities.
func sanitizeSubmissionSecrets(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "").Replace(key))
			switch normalized {
			case "identitynumberciphertext", "identitynumberdigest", "idcardciphertext", "idcarddigest", "ciphertext", "digest", "identitynumber", "idcard",
				"identifiers", "clientidentifiers", "passport", "passportnumber", "socialcreditcode", "unifiedsocialcreditcode", "businesslicense", "organizationcode":
				continue
			}
			result[key] = sanitizeSubmissionSecrets(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for i, item := range typed {
			result[i] = sanitizeSubmissionSecrets(item)
		}
		return result
	case []byte:
		var decoded interface{}
		if json.Unmarshal(typed, &decoded) == nil {
			return sanitizeSubmissionSecrets(decoded)
		}
		return string(typed)
	case string:
		if strings.Contains(strings.ToLower(typed), "hmac-sha256:") {
			// Decode JSON before replacing scalar HMAC values so unrelated fields
			// inside metadata/search_parameters remain available.
			var decoded interface{}
			if json.Unmarshal([]byte(typed), &decoded) != nil {
				return "受保护标识"
			}
		}
		var decoded interface{}
		if json.Unmarshal([]byte(typed), &decoded) == nil {
			switch decoded.(type) {
			case map[string]interface{}, []interface{}, string:
				projected := sanitizeSubmissionSecrets(decoded)
				if !reflect.DeepEqual(decoded, projected) {
					return jsonStringValue(projected)
				}
			}
		}
		return typed
	default:
		return value
	}
}
