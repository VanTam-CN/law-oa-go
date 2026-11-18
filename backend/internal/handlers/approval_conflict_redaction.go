package handlers

import (
	"encoding/json"
	"strings"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// projectApprovalForViewer keeps the applicant's workflow data visible while
// removing historical conflict evidence from approval responses. The browser
// cannot be the disclosure boundary because the same response is also used by
// exports and other clients.
func projectApprovalForViewer(approval *models.ApprovalRequest, role string) *models.ApprovalRequest {
	if approval == nil || services.IsConflictReviewRole(role) {
		return approval
	}

	projected := *approval
	projected.Metadata = redactApprovalJSON(projected.Metadata)
	projected.ConflictResult = redactApprovalJSON(projected.ConflictResult)
	return &projected
}

// projectApprovalSnapshotForViewer applies the same disclosure rule to the
// immutable snapshot endpoint. Conflict officers and authorized business
// management roles retain the frozen evidence needed for independent review.
func projectApprovalSnapshotForViewer(value interface{}, role string) interface{} {
	if services.IsConflictReviewRole(role) {
		return value
	}
	return redactApprovalConflictPayload(value)
}

func redactApprovalJSON(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return raw
	}
	var value interface{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return "冲突证据详情仅向获授权独立冲突核查人展示。"
	}
	return jsonStringValue(redactApprovalConflictPayload(value))
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
			return redactApprovalConflictPayload(decoded)
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
