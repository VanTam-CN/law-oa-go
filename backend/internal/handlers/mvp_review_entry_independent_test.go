package handlers

import "testing"

// A complete machine no-match result cannot eliminate the independent
// reviewer workflow required before this intake can become a case.
func TestIndependentMachineClearIntakeKeepsReviewWorkflow(t *testing.T) {
	decision := map[string]interface{}{"status": "CLEAR", "coverageStatus": "COMPLETE", "evidenceCount": 0}
	result := map[string]interface{}{
		"decision": decision,
		"riskAssessment": map[string]interface{}{"overallRisk": "LOW", "requiresApproval": false},
	}
	row := map[string]interface{}{
		"search_parameters": `{"intakeId":"independent-intake"}`,
		"risk_level": "LOW", "has_conflict": false,
	}
	if conflictApprovalNotRequired(row, result, decision, 0) {
		t.Fatal("intake machine CLEAR must still allow the independent review workflow")
	}
}
