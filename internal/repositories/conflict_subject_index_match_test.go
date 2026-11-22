package repositories

import (
	"testing"

	"law-oa-go/internal/models"
)

func aliasMatchVersion(name string, sourceType string) models.ConflictSubjectVersion {
	return models.ConflictSubjectVersion{
		ID: "SV-test", SubjectKey: "k", SourceType: sourceType, SourceID: "src",
		SubjectRole: "OPPOSING_PARTY", SubjectType: "COMPANY", OriginalName: name,
		NormalizedName: normalizeConflictIndexName(name, "COMPANY"),
	}
}

func aliasMatchSubject(entityType, name string, aliases ...string) models.ConflictNormalizedSubject {
	return models.ConflictNormalizedSubject{
		OriginalName: name, NormalizedName: normalizeConflictIndexName(name, entityType),
		Role: "OPPOSING_PARTY", EntityType: entityType, Aliases: aliases,
	}
}

// 边界一：别名命中时必须返回实际命中的请求称呼，不能冒充主名。
func TestMatchConflictIndexSubjectReturnsMatchedRequestedAlias(t *testing.T) {
	version := aliasMatchVersion("Historical Star Firm", "CLIENT_ARCHIVE_CASE")
	subject := aliasMatchSubject("COMPANY", "Independent Current Brand", "Historical Star Firm")
	matchType, _, matched := matchConflictIndexSubject(version, nil, subject, map[string]string{}, nil)
	if matchType != "CANDIDATE" || matched != "Historical Star Firm" {
		t.Fatalf("alias hit must report the alias that matched: match=%q matched=%q", matchType, matched)
	}
}

// 边界二：主名命中优先于别名命中，标签保持主名。
func TestMatchConflictIndexSubjectPrimaryNameWinsOverAlias(t *testing.T) {
	version := aliasMatchVersion("Alpha Base Ltd", "CLIENT")
	subject := aliasMatchSubject("COMPANY", "Alpha Base Ltd", "Unrelated Brand")
	matchType, _, matched := matchConflictIndexSubject(version, []string{"Unrelated Brand"}, subject, map[string]string{}, nil)
	if matchType != "CANDIDATE" || matched != "Alpha Base Ltd" {
		t.Fatalf("primary hit must keep the primary label: match=%q matched=%q", matchType, matched)
	}
}

// 边界三：身份证摘要精确优先，别名/主名候选不得提升为 EXACT。
func TestMatchConflictIndexSubjectIDCardExactBeatsNameCandidates(t *testing.T) {
	version := aliasMatchVersion("张三", "CLIENT")
	version.SubjectType = "PERSON"
	subject := aliasMatchSubject("PERSON", "历史用名张三", "张三")
	inputIDs := map[string]string{"ID_CARD": "idcard-digest-1"}
	storedIDs := []models.ConflictSubjectIdentifier{{IdentifierType: "ID_CARD", Digest: "idcard-digest-1"}}
	matchType, ruleCode, matched := matchConflictIndexSubject(version, nil, subject, inputIDs, storedIDs)
	if matchType != "EXACT" || ruleCode != "STRUCTURED_IDENTITY_EXACT" {
		t.Fatalf("id card match must stay EXACT: match=%q rule=%q", matchType, ruleCode)
	}
	if matched != "历史用名张三" {
		t.Fatalf("id card label changed: %q", matched)
	}
	matchType, ruleCode, matched = matchConflictIndexSubject(version, nil, subject, nil, storedIDs)
	if matchType != "CANDIDATE" || ruleCode != "PERSON_NAME_ONLY_INSUFFICIENT" || matched != "张三" {
		t.Fatal("alias alone must remain a person-name candidate", matchType, ruleCode, matched)
	}
}

func TestMatchConflictIndexStackedHistoricalAlias(t *testing.T) {
	label := "甲集团控股有限公司"
	version := aliasMatchVersion("历史主体", "CLIENT_ARCHIVE")
	aliases := normalizeConflictIndexAliases([]string{label}, "COMPANY", version.NormalizedName)
	subject := aliasMatchSubject("COMPANY", "现主体", aliases[0])
	subject.AliasLabels = map[string]string{aliases[0]: label}
	match, _, name := matchConflictIndexSubject(version, aliases, subject, nil, nil)
	if match != "CANDIDATE" || name != label {
		t.Fatalf("match=%q name=%q", match, name)
	}
}
