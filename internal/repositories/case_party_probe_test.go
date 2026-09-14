package repositories

// Package-two controlled probe tests, committed in the author tree as
// regression pins for the case-party materialization fixes.
// Evidence layer: sqlite fixture via gorm.io/driver/sqlite — controlled
// repository-layer evidence, NOT real API, NOT real PostgreSQL, NOT
// production. These tests intentionally reproduce suspected defects at
// HEAD 13cdddb913dba7807e50da41b892df62ac01d857 and failed there; the
// materialization fixes make them pass, and they now guard the contract.

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

func probeDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Entity{}, &models.CaseParty{}))
	return db
}

func digestIdent(d string) map[string]string {
	return map[string]string{"id_card": "hmac-sha256:" + d}
}

// P-A: same name, different protected digests. ensureConflictPartyEntity
// checks the in-transaction seenName map before any digest comparison, so
// the second subject merges into the first. Strong identity must win.
func TestProbeSameNameDifferentDigestMustNotMerge(t *testing.T) {
	db := probeDB(t)
	repo := &IntegrationRepository{db: db}
	caseID := uint(9001)
	parties := []models.ConflictPartyInfo{
		{Name: "同名不同人测试有限公司", Role: "CLIENT", Identifiers: digestIdent("digest-A")},
		{Name: "同名不同人测试有限公司", Role: "OPPOSING_PARTY", Identifiers: digestIdent("digest-B")},
	}
	require.NoError(t, repo.LinkConflictCaseParties(context.Background(), caseID, parties))

	var entities []models.Entity
	require.NoError(t, db.Order("id").Find(&entities).Error)
	require.Len(t, entities, 2, "same name + different digest must stay two subjects")

	var rows []models.CaseParty
	require.NoError(t, db.Where("case_id = ?", caseID).Order("id").Find(&rows).Error)
	require.Len(t, rows, 2, "both subjects must appear on the case")
	assert.NotEqual(t, rows[0].EntityID, rows[1].EntityID, "case parties must reference two distinct entities")
	assert.Equal(t, models.PartyTypeClient, rows[0].PartyType)
	assert.Equal(t, models.PartyTypeOpposing, rows[1].PartyType)
}

// P-B: one subject holding two reviewed roles (client AND opposing, e.g.
// a conflict-of-interest scenario recorded during review). The existing-row
// lookup matches on case_id+entity_id only, so the second role is dropped.
// The formal CaseParty schema has no per-role uniqueness column, so the
// design question for the fix is how multi-role is expressed; this probe
// pins the current loss as a defect.
func TestProbeSameSubjectTwoRolesMustBothAppear(t *testing.T) {
	db := probeDB(t)
	repo := &IntegrationRepository{db: db}
	caseID := uint(9002)
	parties := []models.ConflictPartyInfo{
		{Name: "双角色主体测试有限公司", Role: "CLIENT", Identifiers: digestIdent("digest-DUAL")},
		{Name: "双角色主体测试有限公司", Role: "OPPOSING_PARTY", Identifiers: digestIdent("digest-DUAL")},
	}
	require.NoError(t, repo.LinkConflictCaseParties(context.Background(), caseID, parties))

	var rows []models.CaseParty
	require.NoError(t, db.Where("case_id = ?", caseID).Order("id").Find(&rows).Error)
	require.Len(t, rows, 2, "one subject with two reviewed roles must keep both roles")
	roles := []models.PartyRole{rows[0].Role, rows[1].Role}
	assert.Contains(t, roles, models.PartyRolePlaintiff)
	assert.Contains(t, roles, models.PartyRoleDefendant)
}

// P-C: an unreviewable role must NOT silently drop the subject when the
// party carries no other usable role. Current code returns "" and continues,
// losing the party entirely while LinkConflictCaseParties still reports
// success. The contract requires an explicit rejection, not a silent skip.
func TestProbeUnknownRoleMustRejectNotSilentlySkip(t *testing.T) {
	db := probeDB(t)
	repo := &IntegrationRepository{db: db}
	caseID := uint(9003)
	parties := []models.ConflictPartyInfo{
		{Name: "未知角色测试有限公司", Role: "MYSTERY_ROLE", Identifiers: digestIdent("digest-UNK")},
	}
	err := repo.LinkConflictCaseParties(context.Background(), caseID, parties)

	// Expected-by-contract behavior (fails at HEAD): unknown role on a
	// required party is an error, and nothing is written.
	require.Error(t, err, "unknown reviewed role must be rejected explicitly")
	var count int64
	require.NoError(t, db.Model(&models.CaseParty{}).Where("case_id = ?", caseID).Count(&count).Error)
	assert.EqualValues(t, 0, count, "a rejected party batch must not leave partial rows")
	var entityCount int64
	require.NoError(t, db.Model(&models.Entity{}).Count(&entityCount).Error)
	assert.EqualValues(t, 0, entityCount, "transaction rollback must also remove the created entity")
	_ = fmt.Sprintf("probe")
}
