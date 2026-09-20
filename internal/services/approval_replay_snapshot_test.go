package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// TestReplaySnapshotMatchesSameDigestDifferentName pins the replay-idempotency
// contract: after LinkConflictCaseParties merges two reviewed roles of one
// digest-bearing subject into the canonical entity, a byte-identical approve
// replay must still match the snapshot even though the canonical entity name
// differs from the reviewed snapshot name. Same-name parties with different
// digests must stay separated (digest-first identity, never name merging).
func TestReplaySnapshotMatchesSameDigestDifferentName(t *testing.T) {
	db := setupApprovalCaseTxDB(t)
	client := &models.Client{Name: "回放客户有限公司", Email: "replay-client@example.com"}
	require.NoError(t, db.Create(client).Error)
	case_ := &models.Case{CaseNumber: "CASE-REPLAY-1", Title: "t", ClientID: client.ID, CaseType: "CIVIL", Status: "active"}
	require.NoError(t, db.Create(case_).Error)

	snapshot := []models.ConflictPartyInfo{
		{Name: "保证人申报名有限公司", Role: "GUARANTOR", EntityType: "LEGAL_PERSON", Identifiers: map[string]string{"social_credit_code": "hmac-sha256:replay-shared-digest"}},
		{Name: "见证人申报名有限公司", Role: "WITNESS", EntityType: "LEGAL_PERSON", Identifiers: map[string]string{"social_credit_code": "hmac-sha256:replay-shared-digest"}},
	}
	repo := repositories.NewIntegrationRepository(db)
	require.NoError(t, repo.LinkConflictCaseParties(t.Context(), case_.ID, snapshot))

	var rows []models.CaseParty
	require.NoError(t, db.Where("case_id = ?", case_.ID).Find(&rows).Error)
	require.Len(t, rows, 2)

	// Byte-identical replay: same digest, snapshot names kept as reviewed.
	assert.True(t, casePartiesMatchSnapshot(db, case_.ID, snapshot))

	// Same reviewed name, different digest: must not match the merged entity.
	otherDigest := []models.ConflictPartyInfo{
		{Name: "保证人申报名有限公司", Role: "GUARANTOR", EntityType: "LEGAL_PERSON", Identifiers: map[string]string{"social_credit_code": "hmac-sha256:replay-other-digest"}},
	}
	assert.False(t, casePartiesMatchSnapshot(db, case_.ID, otherDigest))
}
