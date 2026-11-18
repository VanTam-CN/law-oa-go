package repositories

import "testing"

// The formal-case party whitelist must accept every reviewed intake role the
// conflict workflow can produce (witness, expert, guarantor) without blocking
// case creation, while still rejecting truly unknown roles fail-closed.
func TestNormalizeConflictPartyRoleMappings(t *testing.T) {
	cases := map[string]string{
		"client":         "CLIENT",
		"co_client":      "CLIENT",
		"opposing_party": "OPPOSING_PARTY",
		"opposing":       "OPPOSING_PARTY",
		"adverse":        "OPPOSING_PARTY",
		"related_party":  "RELATED_PARTY",
		"related":        "RELATED_PARTY",
		"third_party":    "RELATED_PARTY",
		"guarantor":      "RELATED_PARTY",
		"witness":        "WITNESS",
		"expert":         "EXPERT",
	}
	for input, want := range cases {
		if got := normalizeConflictPartyRole(input); got != want {
			t.Fatalf("normalizeConflictPartyRole(%q) = %q, want %q", input, got, want)
		}
	}
	for _, unknown := range []string{"", "unknown_role", "co_defendant_zh", "  "} {
		if got := normalizeConflictPartyRole(unknown); got != "" {
			t.Fatalf("normalizeConflictPartyRole(%q) = %q, want empty rejection", unknown, got)
		}
	}
}

func TestCasePartyRoleMapping(t *testing.T) {
	if pt, pr := casePartyRoleMapping("WITNESS"); pr != "WITNESS" || pt != "RELATED" {
		t.Fatalf("WITNESS mapping = %q/%q, want RELATED/WITNESS", pt, pr)
	}
	if pt, pr := casePartyRoleMapping("EXPERT"); pr != "EXPERT" || pt != "RELATED" {
		t.Fatalf("EXPERT mapping = %q/%q, want RELATED/EXPERT", pt, pr)
	}
	if pt, pr := casePartyRoleMapping("RELATED_PARTY"); pr != "INTERESTED_PARTY" {
		t.Fatalf("RELATED_PARTY mapping = %q/%q, want INTERESTED_PARTY", pt, pr)
	}
}
