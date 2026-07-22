package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func conflictP0SubjectIndexAvailable(caseRepo repositories.CaseRepository) bool {
	if caseRepo == nil || caseRepo.GetDB() == nil {
		return false
	}
	db := caseRepo.GetDB()
	for _, table := range []string{
		"conflict_subject_versions", "conflict_subject_identifiers", "conflict_match_evidence_v2",
		"cases", "clients", "entities", "case_parties", "entity_name_history", "entity_relations", "users",
	} {
		if !db.Migrator().HasTable(table) {
			return false
		}
	}
	return true
}

type indexedConflictSourceSnapshot struct {
	CaseNumber      string    `json:"caseNumber,omitempty"`
	CaseTitle       string    `json:"caseTitle,omitempty"`
	CaseType        string    `json:"caseType,omitempty"`
	CaseDescription string    `json:"caseDescription,omitempty"`
	ClientName      string    `json:"clientName,omitempty"`
	LawyerName      string    `json:"lawyerName,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	RelationType    string    `json:"relationType,omitempty"`
}

func (s *conflictDetectionService) buildIndexedConflictCases(hits []repositories.ConflictP0SubjectIndexHit) []*models.ConflictCase {
	conflicts := make([]*models.ConflictCase, 0, len(hits))
	for _, hit := range hits {
		var source indexedConflictSourceSnapshot
		_ = json.Unmarshal([]byte(hit.Version.Snapshot), &source)
		caseID := strings.TrimSpace(hit.Version.CaseID)
		caseName := source.CaseTitle
		caseNo := source.CaseNumber
		caseType := source.CaseType
		caseStatus := "archive"
		if caseID != "" {
			caseStatus = "active"
		}
		if caseName == "" {
			caseName = "历史主体档案"
		}
		matchType := hit.MatchType
		if matchType == "" {
			matchType = "CANDIDATE"
		}
		ruleCode := hit.RuleCode
		if ruleCode == "" {
			ruleCode = "SUBJECT_CANDIDATE_REVIEW"
		}
		conflictType := "结构化主体候选待核实"
		details := "全所主体索引命中，需由独立冲突核查人核实主体身份、历史角色和保密信息接触范围"
		riskLevel := "MEDIUM"
		if ruleCode == "STRUCTURED_IDENTITY_EXACT" {
			conflictType = "主体身份标识完全匹配待复核"
			details = "全所主体索引中的受保护身份标识一致，但仍须由独立核查人形成最终处置结论"
			riskLevel = "HIGH"
		} else if ruleCode == "PERSON_NAME_ONLY_INSUFFICIENT" {
			conflictType = "自然人同名信息不足"
			details = "仅凭自然人姓名不足以认定为同一主体，请补充证件标识后复核"
		} else if ruleCode == "RELATED_ENTITY_ADVERSE_REVIEW" {
			conflictType = "关联实体命中待复核"
			details = "全所主体索引中的关联实体线索命中，需核实关联方向、控制范围和历史保密信息接触范围"
		} else if ruleCode == "FORMER_NAME_CANDIDATE_REVIEW" {
			conflictType = "曾用名命中待复核"
			details = "全所主体索引中的曾用名/别名命中，需核实是否为同一主体及历史委托范围"
		} else if ruleCode == "CLIENT_ARCHIVE_NAME_CANDIDATE" {
			conflictType = "历史客户主体命中待复核"
			details = "全所主体索引中的历史客户主档命中，需独立核查是否为同一主体及历史保密义务范围"
		}
		if hit.RelationType != "" {
			details += "；关系类型：" + hit.RelationType
		}
		conflict := &models.ConflictCase{
			ID: fmt.Sprintf("indexed_%s_%s", hit.Version.ID, matchType), CaseID: caseID,
			CaseName: caseName, CaseNo: caseNo, CaseType: caseType, ConflictType: conflictType,
			RiskLevel: riskLevel, Description: fmt.Sprintf("当前主体 %q 命中全所主体索引中的 %q，需独立核查", hit.RequestedParty, hit.MatchedName),
			CaseStatus: caseStatus, ClientID: hit.Version.ClientID, OpposingParties: []string{hit.RequestedParty},
			ConflictDetails: details, CreatedAt: source.CreatedAt, MatchType: matchType, RuleCode: ruleCode, RequiresManualReview: true,
		}
		if conflict.CreatedAt.IsZero() {
			conflict.CreatedAt = hit.Version.CreatedAt
		}
		if conflict.CaseID == "" {
			conflict.CaseID = "INDEX:" + hit.Version.ID
		}
		conflict.Evidence = []models.ConflictEvidence{{
			EvidenceID: fmt.Sprintf("EV-INDEX-%s", hit.Version.ID), RuleCode: ruleCode, MatchType: matchType,
			SourceType: indexedEvidenceSource(hit.MatchSource), RequestedParty: hit.RequestedParty, MatchedEntity: hit.MatchedName,
			PartyRole: indexedEvidencePartyRole(hit.MatchSource, hit.Version.SubjectRole), HistoricalRole: indexedHistoricalRole(hit.MatchSource),
			SourceCaseID: caseID, SourceCaseNumber: caseNo, SourceCaseName: caseName, LawyerName: source.LawyerName,
			SourceUpdatedAt: conflict.CreatedAt, Restricted: false, Summary: details,
		}}
		conflicts = append(conflicts, conflict)
	}
	return conflicts
}

func indexedEvidenceSource(source string) string {
	switch source {
	case "RELATED_ENTITY":
		return "CLIENT_RELATION"
	case "CLIENT_ARCHIVE", "CLIENT_ARCHIVE_CASE":
		return "CLIENT_ARCHIVE"
	default:
		return "STRUCTURED_CASE_PARTY"
	}
}

func indexedEvidencePartyRole(source, role string) string {
	if source == "RELATED_ENTITY" {
		return "RELATED_PARTY"
	}
	return normalizeConflictSubjectRole(role)
}

func indexedHistoricalRole(source string) string {
	switch source {
	case "RELATED_ENTITY":
		return "RELATED_ENTITY"
	case "CLIENT_ARCHIVE", "CLIENT_ARCHIVE_CASE":
		return "HISTORICAL_CLIENT"
	default:
		return "HISTORICAL_PARTY"
	}
}
