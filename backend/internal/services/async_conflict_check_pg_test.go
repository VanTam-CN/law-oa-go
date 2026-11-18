package services

// Async conflict-check regression on real PostgreSQL (e2e-live finding,
// 2026-09-26): CreateTask inserts one QUEUED audit row, then the worker used
// PerformConflictCheck, whose save-inside-service flow re-INSERTed the same
// immutable check ID — SQLSTATE 23505 on PostgreSQL, so every async task
// ended FAILED. The fix routes the worker through PrepareConflictCheck plus
// artifact-only persistence, keeping the single-record invariant.
//
// Runs against the isolated trial database; skipped unless INTAKE_MVP_PG_DSN
// is set (same contract as approval_resubmit_pg_test.go).

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func newAsyncConflictPGWorld(t *testing.T) (*gorm.DB, *asyncConflictCheckService, string, string, string, string) {
	t.Helper()
	t.Setenv("SUBJECT_DATA_KEY", "01234567890123456789012345678901")
	dsn := os.Getenv("INTAKE_MVP_PG_DSN")
	if dsn == "" {
		t.Skip("INTAKE_MVP_PG_DSN not set; PG isolation evidence runs separately")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ConflictCheckRecord{},
		&models.ConflictCase{},
		&models.ConflictSubjectVersion{},
		&models.ConflictSubjectIdentifier{},
		&models.ConflictMatchEvidenceV2{},
	))

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	ownerName := "async-pg-owner-" + suffix
	require.NoError(t, db.Exec(
		"INSERT INTO users (username, name, email, password, role, status, department, seniority) VALUES (?, ?, ?, 'x', 'lawyer', 'active', '综合部', '初级')",
		ownerName, "异步回归律师", fmt.Sprintf("%s@example.com", ownerName)).Error)
	t.Cleanup(func() { db.Exec("DELETE FROM users WHERE username = ?", ownerName) })
	var ownerID uint
	require.NoError(t, db.Raw("SELECT id FROM users WHERE username = ?", ownerName).Scan(&ownerID).Error)

	requesterName := "异步回归客户甲-" + suffix
	require.NoError(t, db.Exec(
		"INSERT INTO clients (name, type, status, version, created_by) VALUES (?, '企业', 'active', 1, ?)",
		requesterName, ownerID).Error)
	t.Cleanup(func() { db.Exec("DELETE FROM clients WHERE name = ?", requesterName) })
	var requesterID uint
	require.NoError(t, db.Raw("SELECT id FROM clients WHERE name = ?", requesterName).Scan(&requesterID).Error)

	adverseName := "异步回归对手乙-" + suffix
	require.NoError(t, db.Exec(
		"INSERT INTO clients (name, type, status, version, created_by) VALUES (?, '企业', 'active', 1, ?)",
		adverseName, ownerID).Error)
	t.Cleanup(func() { db.Exec("DELETE FROM clients WHERE name = ?", adverseName) })
	var adverseID uint
	require.NoError(t, db.Raw("SELECT id FROM clients WHERE name = ?", adverseName).Scan(&adverseID).Error)

	// Historical adverse matter: the adverse party is the *client* of an
	// existing case owned by a different client, which is the P0 hit pattern.
	caseNumber := "ASYNC-PG-" + suffix
	require.NoError(t, db.Exec(
		"INSERT INTO cases (case_number, title, case_type, client_id, lawyer_id, status, priority, created_by) VALUES (?, ?, '商事', ?, ?, 'active', 'medium', ?)",
		caseNumber, "异步回归历史案", adverseID, ownerID, fmt.Sprintf("%d", ownerID)).Error)
	var caseID uint
	require.NoError(t, db.Raw("SELECT id FROM cases WHERE case_number = ?", caseNumber).Scan(&caseID).Error)

	checkIDPrefix := "CCT_ASYNCPG-" + suffix
	cleanup := func() {
		db.Exec("DELETE FROM conflict_match_evidence_v2 WHERE check_id LIKE ?", checkIDPrefix+"%")
		db.Exec("DELETE FROM conflict_subject_identifiers WHERE subject_version_id IN (SELECT id FROM conflict_subject_versions WHERE case_id = ? OR client_id IN (?, ?) OR source_id IN (?, ?))",
			caseID, requesterID, adverseID, fmt.Sprintf("%d", requesterID), fmt.Sprintf("%d", adverseID))
		db.Exec("DELETE FROM conflict_subject_versions WHERE case_id = ? OR client_id IN (?, ?) OR source_id IN (?, ?)",
			caseID, requesterID, adverseID, fmt.Sprintf("%d", requesterID), fmt.Sprintf("%d", adverseID))
		db.Exec("DELETE FROM conflict_cases WHERE check_id LIKE ?", checkIDPrefix+"%")
		db.Exec("DELETE FROM conflict_check_records WHERE check_id LIKE ?", checkIDPrefix+"%")
		db.Exec("DELETE FROM cases WHERE id = ?", caseID)
	}
	cleanup()
	t.Cleanup(cleanup)

	repo := repositories.NewConflictRepository(db, nil)
	detector := NewConflictDetectionService(
		repo,
		NewRiskAssessor(nil, nil),
		repositories.NewUserRepository(db),
		repositories.NewClientRepository(db),
		repositories.NewCaseRepository(db),
	)
	return db, NewAsyncConflictCheckService(repo, detector).(*asyncConflictCheckService),
		checkIDPrefix, requesterName, adverseName, fmt.Sprintf("%d", ownerID)
}

func waitForAsyncConflictTask(t *testing.T, svc AsyncConflictCheckService, checkID string) *ConflictCheckTaskResponse {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		task, err := svc.GetTask(context.Background(), checkID)
		require.NoError(t, err)
		if task.Status == ConflictTaskStatusCompleted || task.Status == ConflictTaskStatusFailed {
			return task
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("async task %s did not reach a terminal state within 30s", checkID)
	return nil
}

func TestAsyncConflictCheckPGCompletesWithoutDuplicateInsert(t *testing.T) {
	db, svc, checkIDPrefix, requesterName, adverseName, ownerID := newAsyncConflictPGWorld(t)
	ctx := context.Background()

	buildRequest := func(checkID string) *models.ConflictCheckRequest {
		return &models.ConflictCheckRequest{
			CheckID:      checkID,
			ClientID:     fmt.Sprintf("%d", dbWorldClientID(t, db, requesterName)),
			ClientName:   requesterName,
			ClientType:   "COMPANY",
			CaseName:     "异步回归新案",
			CaseType:     "commercial",
			OtherParties: []string{adverseName},
			UserID:       ownerID,
		}
	}

	firstID := checkIDPrefix + "-1"
	created, err := svc.CreateTask(ctx, buildRequest(firstID))
	require.NoError(t, err)
	require.Equal(t, ConflictTaskStatusQueued, created.Status)

	first := waitForAsyncConflictTask(t, svc, firstID)
	require.Equal(t, ConflictTaskStatusCompleted, first.Status,
		"async task must complete on PostgreSQL; FAILED here reproduces the pre-fix duplicate-INSERT primary-key defect")

	var recordCount int64
	require.NoError(t, db.Table("conflict_check_records").Where("check_id = ?", firstID).Count(&recordCount).Error)
	require.Equal(t, int64(1), recordCount, "immutable audit row must exist exactly once per check ID")
	require.True(t, first.HasConflict, "seeded adverse-party history must be detected")

	var hitCount int64
	require.NoError(t, db.Table("conflict_cases").Where("check_id = ?", firstID).Count(&hitCount).Error)
	require.Greater(t, hitCount, int64(0), "detected hits must persist as conflict_cases artifacts")

	var evidenceCount int64
	require.NoError(t, db.Table("conflict_match_evidence_v2").Where("check_id = ?", firstID).Count(&evidenceCount).Error)
	require.Greater(t, evidenceCount, int64(0), "P0 subject/evidence snapshot must be persisted for async checks")

	// A second task with a fresh ID must also complete — each worker writes
	// only its own artifacts, never a second record for an existing ID.
	secondID := checkIDPrefix + "-2"
	_, err = svc.CreateTask(ctx, buildRequest(secondID))
	require.NoError(t, err)
	second := waitForAsyncConflictTask(t, svc, secondID)
	require.Equal(t, ConflictTaskStatusCompleted, second.Status)
	var secondCount int64
	require.NoError(t, db.Table("conflict_check_records").Where("check_id = ?", secondID).Count(&secondCount).Error)
	require.Equal(t, int64(1), secondCount)

	// Reusing a caller-supplied check ID stays on the supported idempotency
	// path: return the existing task, never a second row.
	reused, err := svc.CreateTask(ctx, buildRequest(firstID))
	require.NoError(t, err)
	require.Equal(t, firstID, reused.CheckID)
	require.Equal(t, ConflictTaskStatusCompleted, reused.Status)
	var finalCount int64
	require.NoError(t, db.Table("conflict_check_records").Where("check_id = ?", firstID).Count(&finalCount).Error)
	require.Equal(t, int64(1), finalCount)
}

// dbWorldClientID is a tiny lookup helper so the request builder can resolve
// seeded client IDs by their unique names.
func dbWorldClientID(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()
	var id uint
	require.NoError(t, db.Raw("SELECT id FROM clients WHERE name = ?", name).Scan(&id).Error)
	return id
}
