package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"law-oa-go/internal/models"
)

// Structured supplement requests (plan §6.1, acceptance C08).
// Storage rides on ApprovalRecord.ImposedRequirements JSON; the shape and
// whitelists here are the contract. Client input never reaches the page as
// HTML, links, or arbitrary field paths.

const supplementSchemaVersion = 1

// supplementMaxItems keeps the list scannable in the revision UI; plan allows
// 1-20 with server-side enforcement taking the stricter of the two limits.
const supplementMaxItems = 20

// supplementMaxDescriptionLen mirrors the plan's 1000-char guidance.
const supplementMaxDescriptionLen = 1000

// supplementCategories is this round's whitelist. material_request joins
// it with a server-side material_id ownership check: the returned item
// must point at a real case_materials row of the approval's own intake
// (contract 2026-09-30, user item 3 + acceptance sample).
var supplementCategories = map[string]bool{
	"field_correction":      true,
	"identity_verification": true,
	"matter_explanation":    true,
	"material_request":      true,
}

// supplementFieldKeys maps to inputs this round actually renders: client,
// parties, case summary/type/area, responsible lawyer (plan §6.1 field_key).
var supplementFieldKeys = map[string]bool{
	"client_name":        true,
	"client_contact":     true,
	"case_title":         true,
	"case_summary":       true,
	"case_type":          true,
	"case_area":          true,
	"responsible_lawyer": true,
	"party":              true,
	// material anchors a material_request item to one case_materials row
	// of the approval's own intake (contract 2026-09-30).
	"material":           true,
}

// supplementItemIDPattern prevents path traversal and arbitrary nesting:
// keys are flat identifiers, not document paths.
var supplementItemIDPattern = regexp.MustCompile("^[a-zA-Z0-9_-]{1,64}$")

// dangerousMarkup catches HTML/script blocks and URLs in free-text
// descriptions; the revision UI renders descriptions as plain text.
var dangerousMarkup = regexp.MustCompile("(?i)<[a-z!/]|https?://|javascript:")

// identityLeakPattern: identity items must name the subject, never quote the
// number or hit details (plan §6.1: 受限命中详情不进普通清单).
var identityLeakPattern = regexp.MustCompile("[0-9Xx]{6,}")

var supplementFieldKeyPattern = regexp.MustCompile("^[a-z][a-z0-9_]{0,63}$")

// SupplementItem is one entry of a structured return request.
type SupplementItem struct {
	ItemID       string `json:"item_id"`
	Category     string `json:"category"`
	FieldKey     string `json:"field_key"`
	PartyLocalID string `json:"party_local_id,omitempty"`
	// MaterialID is set only for material_request items; it must reference
	// a case_materials row of the approval's own intake. Ownership is
	// re-validated against the live intake mapping at decision time, so a
	// stale or cross-application id cannot ride through persistence.
	MaterialID  uint   `json:"material_id,omitempty"`
	Description  string `json:"description"`
	Required     bool   `json:"required"`
	Source       string `json:"source,omitempty"` // "" | legacy_text
}

// ValidateSupplementItems parses and whitelists the raw ImposedRequirements
// payload for a request_changes decision. It never mutates client data
// beyond trimming; every stored item gets a server-generated item_id.
func ValidateSupplementItems(raw map[string]interface{}) ([]SupplementItem, error) {
	if raw == nil {
		return nil, fmt.Errorf("补充项清单不能为空")
	}
	switch v := raw["schema_version"].(type) {
	case float64:
		if int(v) != supplementSchemaVersion {
			return nil, fmt.Errorf("补充项 schema_version 不受支持")
		}
	case int:
		if v != supplementSchemaVersion {
			return nil, fmt.Errorf("补充项 schema_version 不受支持")
		}
	default:
		return nil, fmt.Errorf("补充项 schema_version 不受支持")
	}
	var rawItems []interface{}
	switch v := raw["items"].(type) {
	case []interface{}:
		rawItems = v
	case []map[string]interface{}:
		// Convenience shape for in-process callers building items directly.
		for _, it := range v {
			rawItems = append(rawItems, it)
		}
		case []SupplementItem:
			// In-process callers may pass typed items (e.g. normalize-then-
			// validate); accept and re-validate to keep one whitelist code path.
			for _, it := range v {
				rawItems = append(rawItems, map[string]interface{}{
					"item_id": it.ItemID, "category": it.Category, "field_key": it.FieldKey,
					"material_id":    it.MaterialID,
					"party_local_id": it.PartyLocalID,
					"description":    it.Description,
					"required":       it.Required,
					"source":         it.Source,
				})
			}
		}
	if len(rawItems) == 0 {
		return nil, fmt.Errorf("退回至少需要一项补充项")
	}
	if len(rawItems) > supplementMaxItems {
		return nil, fmt.Errorf("补充项不能超过 %d 项", supplementMaxItems)
	}
	items := make([]SupplementItem, 0, len(rawItems))
	seen := map[string]bool{}
	for i, rawItem := range rawItems {
		obj, ok := rawItem.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("第 %d 项格式错误", i+1)
		}
		item, err := validateSupplementItem(obj, i)
		if err != nil {
			return nil, err
		}
		if seen[item.ItemID] {
			return nil, fmt.Errorf("item_id 重复: %s", item.ItemID)
		}
		seen[item.ItemID] = true
		items = append(items, item)
	}
	return items, nil
}

func validateSupplementItem(obj map[string]interface{}, idx int) (SupplementItem, error) {
	category := strings.TrimSpace(str(obj["category"]))
	if !supplementCategories[category] {
		return SupplementItem{}, fmt.Errorf("第 %d 项 category 不在支持范围", idx+1)
	}
	fieldKey := strings.TrimSpace(str(obj["field_key"]))
	if !supplementFieldKeyPattern.MatchString(fieldKey) || !supplementFieldKeys[fieldKey] {
		return SupplementItem{}, fmt.Errorf("第 %d 项 field_key 不在支持范围", idx+1)
	}
	// material_request items must anchor to a real material id and use the
	// material field key; other categories must not carry the key.
	materialID := materialIDFromItem(obj)
	if category == "material_request" {
		if fieldKey != "material" {
			return SupplementItem{}, fmt.Errorf("第 %d 项 material_request 必须使用 field_key=material", idx+1)
		}
		if materialID == 0 {
			return SupplementItem{}, fmt.Errorf("第 %d 项 material_request 缺少 material_id", idx+1)
		}
	} else if fieldKey == "material" {
		return SupplementItem{}, fmt.Errorf("第 %d 项只有 material_request 才能使用 field_key=material", idx+1)
	}
	description := strings.TrimSpace(str(obj["description"]))
	if description == "" {
		return SupplementItem{}, fmt.Errorf("第 %d 项说明不能为空", idx+1)
	}
	if len([]rune(description)) > supplementMaxDescriptionLen {
		return SupplementItem{}, fmt.Errorf("第 %d 项说明超过 %d 字上限", idx+1, supplementMaxDescriptionLen)
	}
	if dangerousMarkup.MatchString(description) {
		return SupplementItem{}, fmt.Errorf("第 %d 项说明不能包含网页标记或外链", idx+1)
	}
	partyLocalID := strings.TrimSpace(str(obj["party_local_id"]))
	if category == "identity_verification" {
		// Identity items point at a stable party localId and never leak digits.
		if partyLocalID == "" {
			return SupplementItem{}, fmt.Errorf("身份核验项必须提供 party_local_id")
		}
		if !supplementItemIDPattern.MatchString(partyLocalID) {
			return SupplementItem{}, fmt.Errorf("party_local_id 格式不受支持")
		}
		if identityLeakPattern.MatchString(description) {
			return SupplementItem{}, fmt.Errorf("第 %d 项身份核验项说明不能包含证件号或命中详情", idx+1)
		}
	}
	source := strings.TrimSpace(str(obj["source"]))
	itemID := strings.TrimSpace(str(obj["item_id"]))
	if itemID == "" {
		itemID = generateUUID()
	}
	return SupplementItem{
		ItemID:       itemID,
		Category:     category,
		FieldKey:     fieldKey,
		PartyLocalID: partyLocalID,
		MaterialID:   materialID,
		Description:  description,
		Required:     true,
		Source:       source,
	}, nil
}

// materialIDFromItem reads the material_id anchor of a material_request
// item. JSON decoding yields float64; in-process callers may pass typed
// integers, and both shapes coerce without silently accepting junk.
func materialIDFromItem(obj map[string]interface{}) uint {
	switch v := obj["material_id"].(type) {
	case float64:
		if v > 0 && v == float64(uint(v)) {
			return uint(v)
		}
	case int:
		if v > 0 {
			return uint(v)
		}
	case uint:
		if v > 0 {
			return v
		}
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0
		}
		n, err := strconv.ParseUint(s, 10, 32)
		if err == nil && n > 0 {
			return uint(n)
		}
	}
	return 0
}

// NormalizeSupplementRequest prepares the decision payload for a structured
// return. A request_changes decision requires a written reason. When the
// client supplied no structured items, the reason becomes one legacy_text
// item so old callers keep working and the revision UI has something to
// anchor replies to; the original reason text is preserved verbatim.
func NormalizeSupplementRequest(req *models.ApprovalDecisionRequest) error {
	if req == nil || req.Decision != models.ApprovalDecisionRequestChanges {
		return fmt.Errorf("仅退回决定需要补充项清单")
	}
	if strings.TrimSpace(req.DecisionReason) == "" {
		return fmt.Errorf("退回必须填写审批意见")
	}
	if req.ImposedRequirements == nil {
		req.ImposedRequirements = map[string]interface{}{
			"schema_version": float64(supplementSchemaVersion),
			"items":          []interface{}{},
		}
	}
	items, err := ValidateSupplementItems(req.ImposedRequirements)
	if err != nil {
		// No structured items: fall back to a single legacy_text item so old
		// callers keep working and the revision UI can anchor replies. The
		// original reason text is preserved verbatim.
		rawItems, _ := req.ImposedRequirements["items"].([]interface{})
		if len(rawItems) != 0 {
			return err
		}
		legacy := legacySupplementPayload(req.DecisionReason)
		parsed, verr := ValidateSupplementItems(legacy)
		if verr != nil {
			return verr
		}
		parsed[0].Source = "legacy_text"
		req.ImposedRequirements = map[string]interface{}{
			"schema_version": float64(supplementSchemaVersion),
			"items": []interface{}{map[string]interface{}{
				"item_id":     parsed[0].ItemID,
				"category":    parsed[0].Category,
				"field_key":   parsed[0].FieldKey,
				"description": parsed[0].Description,
				"required":    true,
				"source":      "legacy_text",
			}},
		}
		return nil
	}
	// Write the normalized items back so persistence stores the
	// server-generated item_id, not the client's raw input.
	req.ImposedRequirements = map[string]interface{}{
		"schema_version": float64(supplementSchemaVersion),
		"items":          items,
	}
	return nil
}

// legacySupplementPayload wraps a text-only return into the normalized shape.
func legacySupplementPayload(reason string) map[string]interface{} {
	return map[string]interface{}{
		"schema_version": float64(supplementSchemaVersion),
		"items": []interface{}{
			map[string]interface{}{
				"item_id":     generateUUID(),
				"category":    "matter_explanation",
				"field_key":   "case_summary",
				"description": reason,
				"required":    true,
				"source":      "legacy_text",
			},
		},
	}
}

// MarshalSupplementItems serializes validated items for the record column.
func MarshalSupplementItems(items []SupplementItem) (string, error) {
	payload := map[string]interface{}{
		"schema_version": supplementSchemaVersion,
		"items":          items,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// str is a nil-safe string coercion for JSON-decoded values.
func str(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// SupplementRequestError marks contract rejections (HTTP 400) separately from
// internal failures so handlers can map it without string matching.
type SupplementRequestError struct{ Message string }

func (e *SupplementRequestError) Error() string { return e.Message }
