package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"law-oa-go/internal/models"
)

// materializeCasePartiesFromCheck re-creates the formal case parties from the
// reviewed conflict-check snapshot. Parties already carry the reviewed role
// and protected identity digest, so the case keeps every role the lawyer
// actually checked instead of only the client archive row. The primary
// client is anchored to the client-archive entity so the case and the
// archive stay one subject.
func (s *approvalConflictIntegrationService) materializeCasePartiesFromCheck(ctx context.Context, checkID, caseID string) error {
	if strings.TrimSpace(checkID) == "" || strings.TrimSpace(caseID) == "" {
		return fmt.Errorf("冲突检测记录或正式案件ID为空，无法写入案件当事人")
	}
	record, err := s.integrationRepo.GetConflictCheckRecord(ctx, checkID)
	if err != nil || record == nil {
		return fmt.Errorf("冲突检测记录不存在，无法写入案件当事人")
	}
	parties := casePartiesFromSearchParameters(record.SearchParameters)
	if client := clientPartyFromRecord(record); client != nil {
		parties = append([]models.ConflictPartyInfo{*client}, parties...)
	}
	return s.integrationRepo.LinkConflictCaseParties(ctx, parseCaseID(caseID), parties)
}

// clientPartyFromRecord rebuilds the primary client row from the check
// record header and its top-level client identifiers. The client archive
// keeps the protected digest, so the case-party row can reuse the archive
// entity by digest match instead of creating a second subject.
func clientPartyFromRecord(record *models.ConflictCheckRecord) *models.ConflictPartyInfo {
	name := strings.TrimSpace(record.ClientName)
	if name == "" {
		return nil
	}
	party := &models.ConflictPartyInfo{Name: name, Role: "CLIENT", EntityType: "LEGAL_PERSON"}
	if strings.EqualFold(strings.TrimSpace(metadataString(record.SearchParameters, "clientType")), "PERSON") {
		party.EntityType = "INDIVIDUAL"
	}
	if raw, ok := record.SearchParameters["clientIdentifiers"]; ok {
		if idents, ok := raw.(map[string]interface{}); ok {
			party.Identifiers = make(map[string]string, len(idents))
			for key, value := range idents {
				if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
					party.Identifiers[strings.ToLower(key)] = text
				}
			}
		}
	}
	return party
}

// casePartiesFromSearchParameters extracts the reviewed party list from the
// stored search parameters; legacy checks without party info return empty
// and simply skip the materialization.
func casePartiesFromSearchParameters(parameters models.JSON) []models.ConflictPartyInfo {
	if len(parameters) == 0 {
		return nil
	}
	raw, ok := parameters["parties"]
	if !ok || raw == nil {
		return nil
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	parties := make([]models.ConflictPartyInfo, 0, 4)
	if err := json.Unmarshal(encoded, &parties); err != nil {
		return nil
	}
	return parties
}

func parseCaseID(value string) uint {
	id, _ := parseUintString(value)
	return id
}
