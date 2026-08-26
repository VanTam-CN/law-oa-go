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
	for _, role := range []string{"lawyer", "compliance"} {
		_, err := service.Register(AuthActor{UserID: 7, Role: role}, OperationsReadinessEvidenceInput{
			Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: "qa://backup-review",
			ReviewedAt: time.Now().Add(-time.Minute),
		})
		require.ErrorContains(t, err, "only directors or system administrators")
	}

	_, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: "production", EvidenceReference: "ops://evidence",
		ReviewedAt: time.Now(),
	})
	require.ErrorContains(t, err, "scope must be qa or controlled_pilot")

	directorEvidence, err := service.Register(AuthActor{UserID: 7, Role: "director"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: "director://backup-review",
		ReviewedAt: time.Now(),
	})
	require.NoError(t, err)
	require.Equal(t, uint(7), directorEvidence.ReviewedBy)

	evidence, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: "qa://backup-review",
		ReviewedAt: time.Now().Add(-time.Minute),
	})
	require.NoError(t, err)
	require.Equal(t, "backup", evidence.Control)
	require.Equal(t, models.OperationsEvidenceScopeQA, evidence.Scope)
	require.NotEmpty(t, evidence.EvidenceReference)
}

func TestOperationsReadinessUsesLatestEvidenceAndPreservesHistory(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	older, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA,
		EvidenceReference: "qa://backup-2026-07", Notes: "older drill",
		ReviewedAt: time.Now().Add(-48 * time.Hour),
	})
	require.NoError(t, err)
	latest, err := service.Register(AuthActor{UserID: 8, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA,
		EvidenceReference: "qa://backup-2026-08", Notes: "latest drill",
		ReviewedAt: time.Now().Add(-time.Hour),
	})
	require.NoError(t, err)

	var count int64
	require.NoError(t, service.db.Model(&models.OperationsReadinessEvidence{}).Where("control = ? AND scope = ?", "backup", models.OperationsEvidenceScopeQA).Count(&count).Error)
	require.Equal(t, int64(2), count)

	summary, err := service.Summary(models.OperationsEvidenceScopeQA)
	require.NoError(t, err)
	var item *OperationsReadinessControlSummary
	for index := range summary.Items {
		if summary.Items[index].Control == "backup" {
			item = &summary.Items[index]
			break
		}
	}
	require.NotNil(t, item)
	require.Equal(t, "backup", item.Control)
	require.Equal(t, "verified", item.Status)
	require.Equal(t, latest.ID, item.Evidence.ID)
	require.Equal(t, latest.EvidenceReference, item.Evidence.EvidenceReference)
	require.Equal(t, latest.Notes, item.Evidence.Notes)
	require.Equal(t, uint(8), item.Evidence.ReviewedBy)
	require.NotEqual(t, older.EvidenceReference, item.Evidence.EvidenceReference)
}

func TestOperationsReadinessEvidenceRemainsAppendOnly(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	evidence, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA,
		EvidenceReference: "qa://backup-review", ReviewedAt: time.Now().Add(-time.Hour),
	})
	require.NoError(t, err)

	err = service.db.Model(&models.OperationsReadinessEvidence{}).Where("id = ?", evidence.ID).Update("notes", "must not update").Error
	require.ErrorContains(t, err, "append-only")
	err = service.db.Where("id = ?", evidence.ID).Delete(&models.OperationsReadinessEvidence{}).Error
	require.ErrorContains(t, err, "append-only")

	var count int64
	require.NoError(t, service.db.Model(&models.OperationsReadinessEvidence{}).Where("id = ?", evidence.ID).Count(&count).Error)
	require.Equal(t, int64(1), count)
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
