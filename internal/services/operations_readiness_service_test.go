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

func registerOperationsEvidenceExceptControls(t *testing.T, service *OperationsReadinessService, scope string, excludedControls ...string) {
	t.Helper()
	excluded := make(map[string]bool, len(excludedControls))
	for _, control := range excludedControls {
		excluded[control] = true
	}
	for _, control := range []string{"backup", "restore_drill", "incident_owner", "upgrade", "rollback"} {
		if excluded[control] {
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
		Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: "archive://backup-review",
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

func TestOperationsReadinessEvidenceFormsVerifiedHashChain(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	first, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA,
		EvidenceReference: "qa://backup-2026-08", ReviewedAt: time.Now().Add(-2 * time.Hour),
	})
	require.NoError(t, err)
	require.Empty(t, first.PreviousEvidenceID)
	require.NotEmpty(t, first.IntegrityHash)

	second, err := service.Register(AuthActor{UserID: 8, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "restore_drill", Scope: models.OperationsEvidenceScopeQA,
		EvidenceReference: "qa://restore-2026-08", ReviewedAt: time.Now().Add(-time.Hour),
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.PreviousEvidenceID)
	require.NotEqual(t, first.IntegrityHash, second.IntegrityHash)

	registerOperationsEvidenceExceptControls(t, service, models.OperationsEvidenceScopeQA, "backup", "restore_drill")
	summary, err := service.Summary(models.OperationsEvidenceScopeQA)
	require.NoError(t, err)
	require.Equal(t, "verified", summary.Items[0].Integrity)
	require.True(t, summary.Ready)
	require.Equal(t, 7, summary.Score)

	require.NoError(t, service.db.Exec("UPDATE operations_readiness_evidence SET evidence_reference = ? WHERE id = ?", "qa://tampered-evidence", first.ID).Error)
	summary, err = service.Summary(models.OperationsEvidenceScopeQA)
	require.NoError(t, err)
	require.False(t, summary.Ready)
	require.Equal(t, 0, summary.Score)
	require.Equal(t, 0, summary.VerifiedCount)
	require.Equal(t, "failed", summary.Items[0].Integrity)
}

func TestOperationsReadinessDoesNotTrustLegacyUnchainedRows(t *testing.T) {
	service := NewOperationsReadinessService(newOperationsReadinessTestDB(t))
	legacy := models.OperationsReadinessEvidence{
		ID: "legacy00000000000000000000000001", Control: "backup",
		Scope: models.OperationsEvidenceScopeQA, Result: models.OperationsEvidenceResultPassed,
		EvidenceReference: "qa://legacy-backup", ReviewedBy: 7,
		ReviewedAt: time.Now().Add(-time.Hour),
	}
	require.NoError(t, service.db.Create(&legacy).Error)

	summary, err := service.Summary(models.OperationsEvidenceScopeQA)
	require.NoError(t, err)
	require.Equal(t, "legacy-rows-present", summary.Items[0].Integrity)
	require.False(t, summary.Ready)
	require.Equal(t, 0, summary.Score)
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
	for _, reference := range []string{"password=1", "secret://doc", "token://doc", "credential://doc", "not-a-uri", "qa://", "https://example.test", "mailto:backup@example.test"} {
		_, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
			Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: reference,
			ReviewedAt: time.Now(),
		})
		require.Error(t, err)
	}
	_, err := service.Register(AuthActor{UserID: 7, Role: "admin"}, OperationsReadinessEvidenceInput{
		Control: "backup", Scope: models.OperationsEvidenceScopeQA, EvidenceReference: "controlled-pilot://backup-record",
		ReviewedAt: time.Now().Add(time.Hour),
	})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "future"))
}
