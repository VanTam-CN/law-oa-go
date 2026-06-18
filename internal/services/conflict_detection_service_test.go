package services

import "testing"

func TestConflictDetectionPartyNameMatchNormalizesCompanySuffix(t *testing.T) {
	service := &conflictDetectionService{}

	if !service.isPartyNameMatch("上海示例科技有限公司", "示例科技") {
		t.Fatal("expected service party matching to normalize common company suffixes")
	}
}

func TestConflictDetectionPartyNameMatchRejectsUnrelatedNames(t *testing.T) {
	service := &conflictDetectionService{}

	if service.isPartyNameMatch("上海示例科技有限公司", "北京无关贸易有限公司") {
		t.Fatal("expected unrelated names not to match")
	}
}
