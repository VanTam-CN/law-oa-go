package services

import (
	"testing"

	"law-oa-go/internal/models"
)

func TestNormalizeCaseCreationConfigDefaultsEmptyBillingMethod(t *testing.T) {
	config := map[string]interface{}{
		"title":          "D3多主体全链样本",
		"client_id":      1,
		"lawyer_id":      1,
		"case_type":      "commercial",
		"priority":       "medium",
		"billing_method": "",
	}

	approval := &models.ApprovalRequest{Title: "新建案件审批 - D3多主体全链样本", Type: "case_creation"}
	caseData := normalizeCaseCreationConfig(config, approval, "1")

	if got, ok := caseData["billing_method"].(string); !ok || got != "hourly" {
		t.Fatalf("billing_method = %v(%T), want hourly", caseData["billing_method"], caseData["billing_method"])
	}
}

func TestNormalizeCaseCreationConfigPreservesExplicitBillingMethod(t *testing.T) {
	config := map[string]interface{}{
		"title":          "显式计费方式样本",
		"billing_method": "fixed",
	}

	approval := &models.ApprovalRequest{Title: "新建案件审批 - 显式计费方式样本", Type: "case_creation"}
	caseData := normalizeCaseCreationConfig(config, approval, "1")

	if got := caseData["billing_method"]; got != "fixed" {
		t.Fatalf("billing_method = %v, want fixed (must not override explicit value)", got)
	}
}
