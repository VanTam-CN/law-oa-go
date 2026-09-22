package models

import "testing"

func TestConflictCheckRequestAcceptsStoredCaseTypeLabels(t *testing.T) {
	for _, caseType := range []string{"建设工程", "商事诉讼", "劳动争议", "construction", "commercial"} {
		req := &ConflictCheckRequest{
			ClientID:     "1",
			ClientName:   "虚构客户",
			ClientType:   "COMPANY",
			OtherParties: []string{"虚构对方"},
			CaseName:     "虚构案件",
			CaseType:     caseType,
			SearchDepth:  "STANDARD",
			UserID:       "1",
		}
		if err := req.Validate(); err != nil {
			t.Fatalf("stored case type %q should be accepted: %v", caseType, err)
		}
	}
}

func TestConflictCheckRequestRejectsUnknownCaseType(t *testing.T) {
	req := &ConflictCheckRequest{
		ClientID:     "1",
		ClientName:   "虚构客户",
		ClientType:   "COMPANY",
		OtherParties: []string{"虚构对方"},
		CaseName:     "虚构案件",
		CaseType:     "未知类型",
		SearchDepth:  "STANDARD",
		UserID:       "1",
	}
	if err := req.Validate(); err == nil {
		t.Fatal("unknown case types must be rejected")
	}
}

// Table-driven guard: every case-type code the intake UI can save must pass
// the formal conflict-check validation. Regression for the 2026-09-09 pilot
// finding where civil_litigation (and by contract ma) returned 500.
func TestConflictCheckRequestAcceptsEverySharedCaseType(t *testing.T) {
	for _, caseType := range ValidConflictCaseTypes {
		req := &ConflictCheckRequest{
			ClientID:     "1",
			ClientName:   "虚构客户",
			ClientType:   "COMPANY",
			OtherParties: []string{"虚构对方"},
			CaseName:     "虚构案件",
			CaseType:     caseType,
			SearchDepth:  "STANDARD",
			UserID:       "1",
		}
		if err := req.Validate(); err != nil {
			t.Fatalf("shared case type %q must pass formal validation: %v", caseType, err)
		}
	}
}

func TestConflictCheckRequestAcceptsIntakeUIOptionCodes(t *testing.T) {
	intakeUICodes := []string{
		"commercial", "civil", "civil_litigation", "construction", "labor",
		"intellectual", "criminal", "administrative", "financial", "ma",
	}
	for _, code := range intakeUICodes {
		if !IsValidConflictCaseType(code) {
			t.Fatalf("intake UI code %q missing from shared contract", code)
		}
	}
}

func TestValidConflictCaseTypesHasNoDuplicates(t *testing.T) {
	seen := map[string]int{}
	for _, code := range ValidConflictCaseTypes {
		seen[code]++
	}
	for code, count := range seen {
		if count > 1 {
			t.Fatalf("case type %q appears %d times in shared contract", code, count)
		}
	}
}
