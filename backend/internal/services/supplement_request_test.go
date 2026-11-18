package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"law-oa-go/internal/models"
)

// N3 / C08 contract: structured supplement requests (plan §6.1).
// Items are server-validated against whitelists; item_id is generated
// server-side and stays stable once persisted. Legacy text-only returns are
// normalized to one legacy_text item without fabricating structure.

func rawSupplementItems(items ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"schema_version": 1,
		"items":          items,
	}
}

func TestValidateSupplementItems_HappyPath(t *testing.T) {
	raw := rawSupplementItems(
		map[string]interface{}{
			"category": "field_correction", "field_key": "case_title",
			"description": "案件名称与委托合同不一致，请核对后修改",
		},
		map[string]interface{}{
			"category": "identity_verification", "field_key": "party",
			"party_local_id": "party-1", "description": "需核验对方当事人张三的身份",
		},
	)
	got, err := ValidateSupplementItems(raw)
	require.NoError(t, err)
	require.Len(t, got, 2)
	seen := map[string]bool{}
	for _, item := range got {
		require.NotEmpty(t, item.ItemID, "server must generate stable item_id")
		require.False(t, seen[item.ItemID], "item_id must be unique")
		seen[item.ItemID] = true
	}
}

// Typed items re-entering ValidateSupplementItems must keep every anchor:
// material_request items keep their original material_id, and party-scoped
// corrections keep party_local_id plus their whitelist source tag.
func TestValidateSupplementItems_TypedRoundTripRetainsAnchors(t *testing.T) {
	typed := []SupplementItem{
		{
			ItemID:      "item-material",
			Category:    "material_request",
			FieldKey:    "material",
			MaterialID:  42,
			Description: "更换委托合同扫描件",
			Required:    true,
		},
		{
			ItemID:       "item-party",
			Category:     "identity_verification",
			FieldKey:     "party",
			PartyLocalID: "party-9",
			Description:  "补核对方当事人身份",
			Required:     true,
			Source:       "legacy_text",
		},
	}
	validated, err := ValidateSupplementItems(map[string]interface{}{
		"schema_version": 1,
		"items":          typed,
	})
	require.NoError(t, err)
	require.Len(t, validated, 2)
	require.Equal(t, uint(42), validated[0].MaterialID, "material anchor must survive re-validation")
	require.Equal(t, "party-9", validated[1].PartyLocalID, "party anchor must survive re-validation")
	require.Equal(t, "legacy_text", validated[1].Source, "source tag must survive re-validation")
}

func TestValidateSupplementItems_Rejections(t *testing.T) {
	cases := []struct {
		name string
		raw  map[string]interface{}
	}{
		{"empty items", map[string]interface{}{"schema_version": 1, "items": []interface{}{}}},
		{"missing items key", map[string]interface{}{"schema_version": 1}},
		{"too many items", func() map[string]interface{} {
			items := []interface{}{}
			for i := 0; i < 21; i++ {
				items = append(items, map[string]interface{}{
					"category": "field_correction", "field_key": "case_title",
					"description": "x",
				})
			}
			return map[string]interface{}{"schema_version": 1, "items": items}
		}()},
		{"bad category", rawSupplementItems(map[string]interface{}{
			"category": "material_upload", "field_key": "case_title", "description": "x"})},
		{"arbitrary field path", rawSupplementItems(map[string]interface{}{
			"category": "field_correction", "field_key": "metadata.approval_snapshot", "description": "x"})},
		{"html description", rawSupplementItems(map[string]interface{}{
			"category": "field_correction", "field_key": "case_title",
			"description": "请修改 <script>alert(1)</script>"})},
		{"external link", rawSupplementItems(map[string]interface{}{
			"category": "field_correction", "field_key": "case_title",
			"description": "参考 https://evil.example.com/x 修改"})},
		{"oversize description", rawSupplementItems(map[string]interface{}{
			"category": "field_correction", "field_key": "case_title",
			"description": strings.Repeat("长", 1001)})},
		{"identity item leaks digits", rawSupplementItems(map[string]interface{}{
			"category": "identity_verification", "field_key": "party",
			"party_local_id": "party-1",
			"description":    "需核验身份 110101199001011234"})},
		{"party item without local id", rawSupplementItems(map[string]interface{}{
			"category": "identity_verification", "field_key": "party",
			"description": "需核验对方当事人身份"})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateSupplementItems(tc.raw)
			require.Error(t, err, "expected rejection")
		})
	}
}

func TestNormalizeSupplementRequest_LegacyTextBecomesOneItem(t *testing.T) {
	req := &models.ApprovalDecisionRequest{
		Decision:       models.ApprovalDecisionRequestChanges,
		DecisionReason: "请补充案件摘要并核对客户联系方式",
	}
	require.NoError(t, NormalizeSupplementRequest(req))
	got, err := ValidateSupplementItems(req.ImposedRequirements)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "legacy_text", got[0].Source)
}

func TestNormalizeSupplementRequest_EmptyReasonRejected(t *testing.T) {
	req := &models.ApprovalDecisionRequest{
		Decision: models.ApprovalDecisionRequestChanges,
	}
	require.Error(t, NormalizeSupplementRequest(req))
}
