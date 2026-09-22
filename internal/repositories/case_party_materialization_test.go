package repositories

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// setupPartyMaterializationDB builds the minimal schema used by
// LinkConflictCaseParties so the sqlite fixture mirrors the formal tables.
func setupPartyMaterializationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("case_parties_%d_%d.db", time.Now().UnixNano(), rand.Int63()))
	db, err := gorm.Open(sqlite.Open("file:"+dbPath), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Entity{}, &models.CaseParty{}, &models.ConflictCheckRecord{}))
	return db
}

func digestIdentifier(digest string) map[string]string {
	return map[string]string{"id_card": "hmac-sha256:" + digest}
}

// Every reviewed role lands on the formal case: the primary client keeps the
// protected identity digest, the co-client shares the client role, the two
// opposing parties stay separate subjects, and the related guarantor keeps
// its own class. Same-name parties with different digests must never merge.
func TestLinkConflictCasePartiesWritesEveryRole(t *testing.T) {
	db := setupPartyMaterializationDB(t)
	repo := &IntegrationRepository{db: db}
	caseID := uint(31)

	parties := []models.ConflictPartyInfo{
		{Name: "委托方科技有限公司", Role: "CLIENT", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-client")},
		{Name: "共同委托方有限公司", Role: "CO_CLIENT", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-co-client")},
		{Name: "对手方甲有限公司", Role: "OPPOSING_PARTY", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-opposing-a")},
		{Name: "对手方乙有限公司", Role: "OPPOSING_PARTY", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-opposing-b")},
		{Name: "担保人有限公司", Role: "RELATED_PARTY", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-related")},
	}
	require.NoError(t, repo.LinkConflictCaseParties(context.Background(), caseID, parties))

	var rows []models.CaseParty
	require.NoError(t, db.Where("case_id = ?", caseID).Order("display_order").Find(&rows).Error)
	require.Len(t, rows, 5)
	assert.Equal(t, models.PartyRolePlaintiff, rows[0].Role)
	assert.Equal(t, models.PartyTypeClient, rows[0].PartyType)
	assert.Equal(t, models.PartyRolePlaintiff, rows[1].Role)
	assert.Equal(t, models.PartyRoleDefendant, rows[2].Role)
	assert.Equal(t, models.PartyRoleDefendant, rows[3].Role)
	assert.Equal(t, models.PartyRoleInterestedParty, rows[4].Role)

	var entities []models.Entity
	require.NoError(t, db.Order("id").Find(&entities).Error)
	require.Len(t, entities, 5, "same-name parties with different digests must stay separate subjects")
	digests := map[string]bool{}
	for _, entity := range entities {
		digests[entity.IdentityNumberDigest] = true
		assert.Empty(t, entity.IdentityNumber, "plaintext identity must stay empty")
		assert.Empty(t, entity.IdentityNumberCiphertext, "ciphertext identity must stay empty")
	}
	assert.True(t, digests["digest-client"])
	assert.True(t, digests["digest-opposing-a"])
	assert.True(t, digests["digest-opposing-b"])
}

// Re-running the materialization must not duplicate parties: a retry after a
// partial failure reconciles to one complete case instead of double rows.
func TestLinkConflictCasePartiesIsIdempotent(t *testing.T) {
	db := setupPartyMaterializationDB(t)
	repo := &IntegrationRepository{db: db}
	caseID := uint(41)
	parties := []models.ConflictPartyInfo{
		{Name: "重试委托方有限公司", Role: "CLIENT", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-retry")},
		{Name: "重试对手方有限公司", Role: "OPPOSING_PARTY", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-retry-opposing")},
	}
	require.NoError(t, repo.LinkConflictCaseParties(context.Background(), caseID, parties))
	require.NoError(t, repo.LinkConflictCaseParties(context.Background(), caseID, parties))

	var count int64
	require.NoError(t, db.Model(&models.CaseParty{}).Where("case_id = ?", caseID).Count(&count).Error)
	assert.EqualValues(t, 2, count)
}

// Unknown roles must be skipped instead of being written as an unreviewable
// party class, and the call must stay readable for an empty case id.
func TestLinkConflictCasePartiesSkipsUnknownRoles(t *testing.T) {
	db := setupPartyMaterializationDB(t)
	repo := &IntegrationRepository{db: db}
	caseID := uint(51)
	parties := []models.ConflictPartyInfo{
		{Name: "未知角色有限公司", Role: "MYSTERY_ROLE", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-unknown")},
		{Name: "正常对手方有限公司", Role: "opposing", EntityType: "LEGAL_PERSON", Identifiers: digestIdentifier("digest-normal")},
	}
	require.NoError(t, repo.LinkConflictCaseParties(context.Background(), caseID, parties))

	var rows []models.CaseParty
	require.NoError(t, db.Where("case_id = ?", caseID).Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Equal(t, models.PartyRoleDefendant, rows[0].Role)

	require.Error(t, repo.LinkConflictCaseParties(context.Background(), 0, parties))
}
