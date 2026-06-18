package services

import (
	"context"
	"testing"
	"time"

	"law-oa-go/internal/models"
)

func TestAssessRiskKeepsHighRiskApprovalRequirement(t *testing.T) {
	assessor := NewRiskAssessor(nil, nil)
	oldConflict := &models.ConflictCase{
		ID:           "conflict-high-old",
		CaseID:       "case-1",
		CaseName:     "历史高风险案件",
		ConflictType: "代理冲突",
		RiskLevel:    "HIGH",
		CreatedAt:    time.Now().AddDate(-12, 0, 0),
	}

	result, err := assessor.AssessRisk(context.Background(), []*models.ConflictCase{oldConflict}, nil)
	if err != nil {
		t.Fatalf("AssessRisk failed: %v", err)
	}

	if result.OverallRisk != "HIGH" {
		t.Fatalf("expected HIGH to be preserved, got %s", result.OverallRisk)
	}
	if !result.RequiresApproval {
		t.Fatal("HIGH conflict must require approval even when score is reduced by weights or time decay")
	}
}

func TestAssessRiskKeepsCriticalRiskApprovalRequirement(t *testing.T) {
	assessor := NewRiskAssessor(nil, nil)
	conflict := &models.ConflictCase{
		ID:           "conflict-critical",
		CaseID:       "case-2",
		CaseName:     "严重冲突案件",
		ConflictType: "对方当事人冲突",
		RiskLevel:    "CRITICAL",
		CreatedAt:    time.Now(),
	}

	result, err := assessor.AssessRisk(context.Background(), []*models.ConflictCase{conflict}, nil)
	if err != nil {
		t.Fatalf("AssessRisk failed: %v", err)
	}

	if result.OverallRisk != "CRITICAL" {
		t.Fatalf("expected CRITICAL to be preserved, got %s", result.OverallRisk)
	}
	if !result.RequiresApproval {
		t.Fatal("CRITICAL conflict must require approval")
	}
}
