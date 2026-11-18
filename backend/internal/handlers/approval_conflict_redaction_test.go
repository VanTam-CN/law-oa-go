package handlers

import (
	"strings"
	"testing"

	"law-oa-go/internal/models"
)

func TestProjectApprovalForViewerRedactsHistoricalConflictEvidence(t *testing.T) {
	approval := &models.ApprovalRequest{
		Metadata:       `{"conflict_cases":[{"case_no":"CASE-SECRET-001","case_name":"历史并购项目","conflict_type":"结构化主体候选待核实","risk_level":"低风险","description":"策略摘要"}],"conflict_result":{"conflictCases":[{"caseNo":"CASE-SECRET-001","caseName":"历史并购项目","description":"策略摘要"}],"evidence":[{"sourceCaseNumber":"CASE-SECRET-001","summary":"内部工作底稿"}]}}`,
		ConflictResult: `{"conflictCases":[{"caseNo":"CASE-SECRET-001","caseName":"历史并购项目"}]}`,
	}

	projected := projectApprovalForViewer(approval, "lawyer")
	serialized := projected.Metadata + projected.ConflictResult
	if strings.Contains(serialized, "CASE-SECRET-001") || strings.Contains(serialized, "历史并购项目") {
		t.Fatalf("普通律师响应泄露历史冲突证据: %s", serialized)
	}
	if !strings.Contains(serialized, "受限历史事项") || !strings.Contains(serialized, "存在受隔离记录") {
		t.Fatalf("普通律师响应缺少受限提示: %s", serialized)
	}
	if strings.Contains(serialized, "结构化主体候选待核实") || strings.Contains(serialized, "低风险") {
		t.Fatalf("普通律师响应泄露冲突类型或风险等级: %s", serialized)
	}
}

func TestProjectApprovalForViewerKeepsEvidenceForConflictReviewer(t *testing.T) {
	approval := &models.ApprovalRequest{
		Metadata: `{"conflict_cases":[{"case_no":"CASE-SECRET-001","case_name":"历史并购项目"}]}`,
	}

	projected := projectApprovalForViewer(approval, "conflict_officer")
	if projected.Metadata != approval.Metadata {
		t.Fatalf("独立冲突核查人的冻结证据不应被脱敏: got %s", projected.Metadata)
	}
}

func TestProjectApprovalSnapshotForViewerRedactsJSONSnapshot(t *testing.T) {
	raw := `{"conflict_result":{"conflictCases":[{"caseNo":"CASE-SECRET-001","caseName":"历史并购项目"}]},"case_creation_config":{"title":"当前申请案件"}}`

	projected := projectApprovalSnapshotForViewer(raw, "lawyer")
	serialized := jsonStringValue(projected)
	if strings.Contains(serialized, "CASE-SECRET-001") || strings.Contains(serialized, "历史并购项目") {
		t.Fatalf("审批快照泄露历史冲突证据: %s", serialized)
	}
	if !strings.Contains(serialized, "当前申请案件") || !strings.Contains(serialized, "受限历史事项") {
		t.Fatalf("审批快照未保留当前申请或受限提示: %s", serialized)
	}
}

func TestProjectApprovalForViewerRedactsNestedConflictRecordResult(t *testing.T) {
	approval := &models.ApprovalRequest{
		Metadata: `{"conflict_record":{"risk_level":"LOW","has_conflict":false,"check_result":{"decision":{"status":"REVIEW_REQUIRED"},"riskAssessment":{"overallRisk":"LOW"},"conflictCases":[{"caseNo":"CASE-SECRET-002","caseName":"历史诉讼项目"}]}}}`,
	}

	projected := projectApprovalForViewer(approval, "lawyer")
	serialized := projected.Metadata
	if strings.Contains(serialized, "CASE-SECRET-002") || strings.Contains(serialized, "历史诉讼项目") {
		t.Fatalf("普通律师响应泄露嵌套历史冲突证据: %s", serialized)
	}
	if strings.Contains(serialized, `"risk_level":"LOW"`) || strings.Contains(serialized, `"overallRisk":"LOW"`) {
		t.Fatalf("普通律师响应保留了误导性的低风险结论: %s", serialized)
	}
	if !strings.Contains(serialized, "REVIEW_REQUIRED") || !strings.Contains(serialized, "受限历史事项") {
		t.Fatalf("普通律师响应缺少范围受限复核状态: %s", serialized)
	}
}
