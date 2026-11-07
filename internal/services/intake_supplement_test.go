package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"law-oa-go/internal/models"
)

func supplementTestInput() map[string]interface{} {
	return map[string]interface{}{"schema_version": 1, "supplement_requests": []map[string]interface{}{
		{"category": "field_correction", "field_key": "title", "description": "请补正案件名称", "required": true},
		{"category": "matter_explanation", "field_key": "description", "description": "请说明委托范围", "required": true},
	}}
}
func TestIntakeSupplementNormalization(t *testing.T) {
	raw, err := NormalizeSupplementRequirements(supplementTestInput(), "请补正两项")
	require.NoError(t, err)
	items, err := ReadSupplementRequests(&models.ApprovalRecord{ID: "r1", ImposedRequirements: raw})
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.NotEmpty(t, items[0].ItemID)
	require.NotEqual(t, items[0].ItemID, items[1].ItemID)
	again, err := ReadSupplementRequests(&models.ApprovalRecord{ID: "r1", ImposedRequirements: raw})
	require.NoError(t, err)
	require.Equal(t, items, again)
	_, err = normalizeSupplementResponses(items, []models.SupplementResponse{{ItemID: items[0].ItemID, Response: "已修正"}}, false)
	require.NoError(t, err)
	_, err = normalizeSupplementResponses(items, []models.SupplementResponse{{ItemID: items[0].ItemID, Response: "已修正"}}, true)
	require.Error(t, err)
	for _, responses := range [][]models.SupplementResponse{
		{{ItemID: "forged", Response: "完成"}},
		{{ItemID: items[0].ItemID, Response: "完成"}, {ItemID: items[0].ItemID, Response: "完成"}},
		{{ItemID: items[0].ItemID, Response: strings.Repeat("字", 2001)}},
	} {
		_, err := normalizeSupplementResponses(items, responses, false)
		require.Error(t, err)
	}
}
func TestIntakeSupplementRejectUncontrolledKeys(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(map[string]interface{})
	}{
		{"unknown envelope", func(m map[string]interface{}) { m["approved"] = true }},
		{"spoofed ID", func(m map[string]interface{}) {
			m["supplement_requests"].([]map[string]interface{})[0]["item_id"] = "forged"
		}},
		{"unknown item field", func(m map[string]interface{}) {
			m["supplement_requests"].([]map[string]interface{})[0]["verified"] = true
		}},
		{"arbitrary path", func(m map[string]interface{}) {
			m["supplement_requests"].([]map[string]interface{})[0]["field_key"] = "metadata.conflict_check_id"
		}},
		{"upload", func(m map[string]interface{}) {
			m["supplement_requests"].([]map[string]interface{})[0]["category"] = "upload"
		}},
		{"HTML", func(m map[string]interface{}) {
			m["supplement_requests"].([]map[string]interface{})[0]["description"] = "<script>hi</script>"
		}},
		{"empty list", func(m map[string]interface{}) { m["supplement_requests"] = []interface{}{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := supplementTestInput()
			tc.change(m)
			_, err := NormalizeSupplementRequirements(m, "说明")
			require.Error(t, err)
		})
	}
	_, err := NormalizeSupplementRequirements(supplementTestInput(), "")
	require.Error(t, err)
	m := supplementTestInput()
	m["supplement_requests"].([]map[string]interface{})[0]["category"] = "identity_verification"
	m["supplement_requests"].([]map[string]interface{})[0]["description"] = "输入不应进入普通身份清单"
	raw, err := NormalizeSupplementRequirements(m, "说明")
	require.NoError(t, err)
	require.NotContains(t, raw, "输入不应")
}
func TestIntakeSupplementLegacyReadOnlyAndFingerprint(t *testing.T) {
	record := &models.ApprovalRecord{ID: "ret", ImposedRequirements: "{}", DecisionReason: "旧意见"}
	items, err := ReadSupplementRequests(record)
	require.NoError(t, err)
	require.Equal(t, "legacy-ret", items[0].ItemID)
	require.Equal(t, "legacy_text", items[0].Source)
	require.Equal(t, "{}", record.ImposedRequirements)
	req := &models.ResubmitApprovalRequest{ExpectedIntakeRevision: "v1", ExpectedSubmissionSnapshotID: "s1", ReturnRecordID: "r1", RevisionNote: " 修改说明 ", Responses: []models.SupplementResponse{{ItemID: "b", Response: " 二 "}, {ItemID: "a", Response: "一"}}}
	one, err := resubmissionFingerprint(req)
	require.NoError(t, err)
	req.Responses = []models.SupplementResponse{{ItemID: "a", Response: "一"}, {ItemID: "b", Response: "二"}}
	two, err := resubmissionFingerprint(req)
	require.NoError(t, err)
	require.Equal(t, one, two)
	raw, err := json.Marshal(req)
	require.NoError(t, err)
	require.Contains(t, string(raw), "supplement_responses")
	req.RevisionNote = "不同"
	three, err := resubmissionFingerprint(req)
	require.NoError(t, err)
	require.NotEqual(t, one, three)
}
