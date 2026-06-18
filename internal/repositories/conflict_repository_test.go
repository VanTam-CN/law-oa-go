package repositories

import "testing"

func TestIsDirectOpposingPartyClientMatchesNormalizedCompanyName(t *testing.T) {
	if !isDirectOpposingPartyClient("上海示例科技有限公司", []string{"示例科技"}) {
		t.Fatal("expected normalized company name to match opposing party")
	}
}

func TestIsDirectOpposingPartyClientIgnoresUnrelatedParty(t *testing.T) {
	if isDirectOpposingPartyClient("上海示例科技有限公司", []string{"北京无关贸易有限公司"}) {
		t.Fatal("expected unrelated company name not to match")
	}
}
