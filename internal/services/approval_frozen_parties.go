package services

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// New rounds materialize exactly the approved party set. An administrative
// correction to a check record must not silently change approved case facts.
func (s *approvalConflictIntegrationService) approvedCaseParties(ctx context.Context, tx *gorm.DB, approvalID, checkID string) ([]models.ConflictPartyInfo, error) {
	var approval models.ApprovalRequest
	if err := tx.First(&approval, "id = ?", approvalID).Error; err != nil {
		return nil, err
	}
	if approval.ActiveSubmissionSnapshotID == nil {
		record, err := s.integrationRepo.GetConflictCheckRecord(ctx, checkID)
		if err != nil {
			return nil, err
		}
		return buildReviewedParties(record)
	}
	snapshot, err := ResolveApprovalSubmission(tx, &approval)
	if err != nil {
		return nil, err
	}
	metadata := submissionMap(snapshot["metadata"])
	raw, err := json.Marshal(metadata["parties"])
	if err != nil {
		return nil, err
	}
	var parties []models.ConflictPartyInfo
	if err := json.Unmarshal(raw, &parties); err != nil || len(parties) < 2 {
		return nil, snapshotUnresolved("批准快照缺少完整当事人集合，已阻止成案")
	}
	for _, party := range parties {
		if party.Name == "" || party.Role == "" || submissionIdentityDigest(party.Identifiers) == "" {
			return nil, snapshotUnresolved("批准快照的当事人身份或角色不完整")
		}
	}
	return parties, nil
}
