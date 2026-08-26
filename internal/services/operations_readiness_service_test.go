package services

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func newOperationsReadinessTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/operations.db"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.OperationsReadinessEvidence{}))
	return db
}

func registerAllOperationsEvidence(t *testing.T, service *OperationsReadinessService, scope string) {
	t.Helper()
	for _, control := range []string{"backup", "restore_drill", "incident_owner", "upgrade", "rollback"} {
		_, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
			Control: control, Scope: scope, EvidenceReference: "ticket://" + control,
			ReviewedAt: time.Now().Add(-time.Hour),
		})
		require.NoError(t, err)
	}
}

func registerOperationsEvidenceExcept(t *testing.T, service *OperationsReadinessService, scope, excludedControl string) {
	t.Helper()
	for _, control := range []string{"backup", "restore_drill", "incident_owner", "upgrade", "rollback"} {
		if control == excludedControl {
			continue
		}
		_, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
			Control: control, Scope: scope, EvidenceReference: "ticket://" + control,
			ReviewedAt: time.Now().Add(-time.Hour),
		})
		require.NoError(t, err)
	}
}

func TestOperationsReadinessSummaryStartsPendingAndCannotUseHealthChecks(t *testing.T) {
	summary, err := NewOperationsReadinessService(newOperationsReadinessTestDB(t)).Summary(models.OperationsEvidenceScopeControlledPilot)
	require.NoError(t, err)
	require.False(t, summary.Ready)
	require.Equal(t, 0, summary.VerifiedCount)
	require.Equal(t, 5, summary.Total)
	require.Equal(t, 0, summary.Score)
	require.Equal(t, 7, summary.MaximumScore)
	require.False(t, summary.ProductionReady)
	require.Equal(t, "production_external_evidence", summary.ProductionGate)
}

func TestOperationsReadinessRegistrationIsControlledAndAuditable(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	_, err := service.Register(AuthActor{UserID: 7, Role: "lawyer"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: "production", EvidenceReference: "ops://evidence",
		ReviewedAt: time.Now(),
	})
	require.Error(t, err)

	_, err = service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: "production", EvidenceReference: "ops://evidence",
		ReviewedAt: time.Now(),
	})
	require.Error(t, err)

	evidence, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: "qa://backup-review",
		ReviewedAt: time.Now().Add(-time.Minute),
	})
	require.NoError(t, err)
	require.Equal(t, "backup", evidence.Control)
	require.Equal(t, models.OperationsEvidenceScopeQA, evidence.Scope)
	require.NotEmpty(t, evidence.EvidenceReference)
}

func TestOperationsReadinessCapsControlledScopeAtSeven(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	registerAllOperationsEvidence(t, service, models.OperationsEvidenceScopeControlledPilot)
	summary, err := service.Summary(models.OperationsEvidenceScopeControlledPilot)
	require.NoError(t, err)
	require.True(t, summary.Ready)
	require.Equal(t, 5, summary.VerifiedCount)
	require.Equal(t, 7, summary.Score)
	require.Equal(t, 7, summary.MaximumScore)
	require.False(t, summary.ProductionReady)
}

func TestOperationsReadinessDoesNotAwardCompletionBonusWhenEvidenceIsMissing(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	registerOperationsEvidenceExcept(t, service, models.OperationsEvidenceScopeQA, "rollback")

	summary, err := service.Summary(models.OperationsEvidenceScopeQA)
	require.NoError(t, err)
	require.False(t, summary.Ready)
	require.Equal(t, 4, summary.VerifiedCount)
	require.Equal(t, 4, summary.Score)
	require.Equal(t, 7, summary.MaximumScore)
	require.Less(t, summary.Score, summary.MaximumScore)
	require.False(t, summary.ProductionReady)
}

func TestOperationsReadinessRejectsCredentialLikeReferencesAndFutureReviews(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	for _, reference := range []string{"password=1", "secret://doc", "token://doc"} {
		_, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
			Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: reference,
			ReviewedAt: time.Now(),
		})
		require.Error(t, err)
	}
	_, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: "qa://backup",
		ReviewedAt: time.Now().Add(time.Hour),
	})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "future"))
}
