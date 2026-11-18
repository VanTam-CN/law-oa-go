package services

// C4 initial submission on real isolated PostgreSQL. These tests skip
// unless INTAKE_MVP_PG_DSN points at the isolated intake-mvp database
// (C16 full-chain attempt2 schema). They prove the four DB bindings
// (conflict_check_id column, active association, immutable first-round
// intake_submission snapshot, active_submission_snapshot_id pointer) and
// the atomic rollback of forged pointers, missing reviews, incomplete
// coverage, and a racing fact edit, on the real uuid/jsonb schema that
// SQLite cannot represent.

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
)

const pgC4FactsVersion = "FV-C4-PG-BASE"

type pgC4World struct {
	db           *gorm.DB
	svc          *ApprovalService
	intakeID     string
	checkID      string
	workflowType string
}

func newPGC4World(t *testing.T, intakeStatus string, seedReview bool) *pgC4World {
	t.Helper()
	dsn := os.Getenv("INTAKE_MVP_PG_DSN")
	if dsn == "" {
		t.Skip("INTAKE_MVP_PG_DSN not set; PG isolation evidence runs separately")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	var currentDB string
	require.NoError(t, db.Raw("SELECT current_database()").Scan(&currentDB).Error)
	t.Logf("PG C4 evidence database: %s", currentDB)

	w := &pgC4World{
		db:           db,
		svc:          NewApprovalService(db),
		intakeID:     uuid.NewString(),
		checkID:      "CCT-PG-C4-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		workflowType: "INTAKE_PG_C4_NO_MATCH_" + fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	// Deterministic assignment: a unique workflow type never matches, so
	// AssignApprover falls back to the first active admin who is not the
	// applicant. Seed one explicitly instead of depending on the shared
	// database roster.
	adminName := "admin-c4pg-" + fmt.Sprintf("%d", time.Now().UnixNano())
	require.NoError(t, db.Exec(
		"INSERT INTO users (username, name, email, password, role, status, department, seniority) VALUES (?, ?, ?, 'x', 'admin', 'active', '综合部', '初级')",
		adminName, "C4审批管理员", fmt.Sprintf("%s@example.com", adminName)).Error)
	t.Cleanup(func() { db.Exec("DELETE FROM users WHERE username = ?", adminName) })

	intakeMeta := fmt.Sprintf(
		`{"draft_revision":"rev-c4","conflict_check_id":"%s","lawyer_facts_version":"%s","conflict_check_facts_version":"%s"}`,
		w.checkID, pgC4FactsVersion, pgC4FactsVersion)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, intake_code, title, case_type, priority, description, client_id, created_by, status, metadata, created_at, updated_at) "+
			"VALUES (?, ?, ?, 'CIVIL', 'high', 'PG C4 事实描述', 9, 1, ?, ?::jsonb, NOW(), NOW())",
		w.intakeID, "C4PG-"+w.intakeID[:8], "PG C4 原子提交", intakeStatus, intakeMeta).Error)
	t.Cleanup(func() { db.Exec("DELETE FROM case_intakes WHERE id = ?", w.intakeID) })

	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, relation_depth, identity_type, identity_number_digest, metadata) "+
			"VALUES (?, ?, 'company', 'opposing_party', 1, 'uscc', 'hmac-sha256:c4-pg-digest', '{}'::jsonb)",
		w.intakeID, "对手方PG有限公司").Error)
	t.Cleanup(func() { db.Exec("DELETE FROM case_intake_parties WHERE intake_id = ?", w.intakeID) })

	w.seedConflictChain(t, seedReview)
	w.cleanupApprovals(t)
	return w
}

// seedConflictChain mirrors seedPGConflictChain minus the approval-side
// association: the atomic initial submission writes its own association,
// so the seed only lays the intake pointer, the immutable record, and the
// independent review.
func (w *pgC4World) seedConflictChain(t *testing.T, seedReview bool) {
	t.Helper()
	require.NoError(t, w.db.Exec(
		"UPDATE case_intakes SET status = 'conflict_ready', metadata = metadata || ?::jsonb, updated_at = NOW() WHERE id = ?",
		fmt.Sprintf(`{"conflict_check_id":"%s","lawyer_facts_version":"%s","conflict_check_facts_version":"%s"}`, w.checkID, pgC4FactsVersion, pgC4FactsVersion),
		w.intakeID).Error)
	require.NoError(t, w.db.Create(&models.ConflictCheckRecord{
		CheckID:          w.checkID,
		ClientID:         "9",
		ClientName:       "pg-c4-customer",
		CheckStatus:      "COMPLETED",
		SearchParameters: models.JSON{"intakeId": w.intakeID, "clientId": "9"},
		CheckResult:      models.JSON{"decision": map[string]interface{}{"coverageStatus": "COMPLETE"}},
		UserID:           1,
		CheckTime:        time.Now(),
	}).Error)
	if seedReview {
		require.NoError(t, w.db.Exec(
			"INSERT INTO conflict_reviews (check_id, decision, notes, reviewer_id, reviewer_name, evidence_hash, created_at) "+
				"VALUES (?, 'no_conflict', '独立复核通过', 1, '核查人', 'pg-c4-hash', NOW())", w.checkID).Error)
	}
	t.Cleanup(func() {
		if seedReview {
			w.db.Exec("DELETE FROM conflict_reviews WHERE check_id = ?", w.checkID)
		}
		w.db.Exec("DELETE FROM conflict_check_records WHERE check_id = ?", w.checkID)
	})
}

// cleanupApprovals removes every approval the tests created for this
// intake plus its dependent rows; the red-sample data elsewhere in the
// database is never touched.
func (w *pgC4World) cleanupApprovals(t *testing.T) {
	t.Helper()
	cleanup := func() {
		var ids []string
		w.db.Raw("SELECT id FROM approval_requests WHERE metadata->>'intake_id' = ?", w.intakeID).Scan(&ids)
		for _, id := range ids {
			w.db.Exec("DELETE FROM approval_snapshots WHERE approval_request_id = ?", id)
			w.db.Exec("DELETE FROM approval_case_creation_tracking WHERE approval_request_id = ?", id)
			w.db.Exec("DELETE FROM approval_conflict_associations WHERE approval_request_id = ?", id)
			w.db.Exec("DELETE FROM inbox_items WHERE source_ref = ?", id)
			w.db.Exec("DELETE FROM approval_records WHERE approval_request_id = ?", id)
		}
		w.db.Exec("DELETE FROM approval_requests WHERE metadata->>'intake_id' = ?", w.intakeID)
	}
	cleanup()
	t.Cleanup(cleanup)
}

func (w *pgC4World) createReq() *models.CreateApprovalRequest {
	return &models.CreateApprovalRequest{
		Type:         "CASE_CREATION",
		Title:        "PG C4 原子提交",
		Content:      "atomic submission on real PG",
		WorkflowType: w.workflowType,
		Metadata:     map[string]interface{}{"intake_id": w.intakeID},
	}
}

func (w *pgC4World) intakeLinkedCounts() (approvals, binds, assoc, snaps int64) {
	w.db.Raw("SELECT COUNT(*) FROM approval_requests WHERE metadata->>'intake_id' = ?", w.intakeID).Scan(&approvals)
	w.db.Raw("SELECT COUNT(*) FROM approval_case_creation_tracking t JOIN approval_requests a ON a.id = t.approval_request_id WHERE a.metadata->>'intake_id' = ?", w.intakeID).Scan(&binds)
	w.db.Raw("SELECT COUNT(*) FROM approval_conflict_associations s JOIN approval_requests a ON a.id = s.approval_request_id WHERE a.metadata->>'intake_id' = ?", w.intakeID).Scan(&assoc)
	w.db.Raw("SELECT COUNT(*) FROM approval_snapshots p JOIN approval_requests a ON a.id = p.approval_request_id WHERE a.metadata->>'intake_id' = ?", w.intakeID).Scan(&snaps)
	return
}

func requirePGC4ZeroResidue(t *testing.T, w *pgC4World) {
	t.Helper()
	approvals, binds, assoc, snaps := w.intakeLinkedCounts()
	require.Zero(t, approvals, "no approval draft may survive a rejected submission")
	require.Zero(t, binds, "no intake slot binding may survive a rejected submission")
	require.Zero(t, assoc, "no conflict association may survive a rejected submission")
	require.Zero(t, snaps, "no intake_submission snapshot may survive a rejected submission")
}

func TestPGC4InitialSubmissionBindsAllFour(t *testing.T) {
	w := newPGC4World(t, "approval_submitting", true)

	created, err := w.svc.CreateAndSubmitApproval("u-lawyer-c4", "律师甲", w.createReq())
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusSubmitted, created.Status)
	require.Equal(t, w.checkID, created.ConflictCheckID)
	require.NotNil(t, created.ActiveSubmissionSnapshotID)

	var bindCount int64
	require.NoError(t, w.db.Raw("SELECT COUNT(*) FROM approval_case_creation_tracking WHERE approval_request_id = ? AND intake_id = ?", created.ID, w.intakeID).Scan(&bindCount).Error)
	require.Equal(t, int64(1), bindCount)

	var assocCount int64
	var assocCheckID string
	require.NoError(t, w.db.Raw("SELECT COUNT(*) FROM approval_conflict_associations WHERE approval_request_id = ? AND association_status = 'active'", created.ID).Scan(&assocCount).Error)
	require.Equal(t, int64(1), assocCount)
	require.NoError(t, w.db.Raw("SELECT conflict_check_id FROM approval_conflict_associations WHERE approval_request_id = ? AND association_status = 'active'", created.ID).Scan(&assocCheckID).Error)
	require.Equal(t, w.checkID, assocCheckID)

	var snap struct {
		ID          string
		PayloadData string
	}
	require.NoError(t, w.db.Raw("SELECT id, snapshot_data AS payload_data FROM approval_snapshots WHERE approval_request_id = ? AND snapshot_type = 'intake_submission' AND source_version = 1", created.ID).Take(&snap).Error)
	require.Equal(t, *created.ActiveSubmissionSnapshotID, snap.ID,
		"active_submission_snapshot_id must point at the committed first-round snapshot")
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.PayloadData), &payload))
	require.Equal(t, w.checkID, payload["conflict_check_id"])
	structured, ok := payload["structured_facts"].(map[string]interface{})
	require.True(t, ok, "snapshot must carry structured_facts")
	require.Equal(t, "PG C4 原子提交", structured["title"])
	require.Equal(t, "CIVIL", structured["case_type"])
	require.Equal(t, "PG C4 事实描述", structured["description"])
	parties, ok := structured["parties"].([]interface{})
	require.True(t, ok)
	require.Len(t, parties, 1)
	first, _ := parties[0].(map[string]interface{})
	require.Equal(t, "对手方PG有限公司", first["entity_name"])
	require.Equal(t, "hmac-sha256:c4-pg-digest", first["identity_digest"],
		"snapshot carries the protected digest, never ciphertext")
}

func TestPGC4InitialSubmissionForgedPointerFailsClosed(t *testing.T) {
	w := newPGC4World(t, "approval_submitting", true)
	require.NoError(t, w.db.Exec(
		"UPDATE case_intakes SET metadata = jsonb_set(metadata, '{conflict_check_id}', to_jsonb(?::text)) WHERE id = ?",
		"CCT-PG-C4-FORGED", w.intakeID).Error)

	_, err := w.svc.CreateAndSubmitApproval("u-lawyer-c4", "律师甲", w.createReq())
	require.Error(t, err)
	var gateErr *ApprovalConflictGateError
	require.ErrorAs(t, err, &gateErr)
	require.Equal(t, "CONFLICT_RESULT_NOT_COMPLETED", gateErr.Code,
		"a forged pointer cannot resolve to a completed record owned by this intake")
	requirePGC4ZeroResidue(t, w)
}

func TestPGC4InitialSubmissionMissingReviewFailsClosed(t *testing.T) {
	w := newPGC4World(t, "approval_submitting", false)
	var reviewCount int64
	require.NoError(t, w.db.Raw("SELECT COUNT(*) FROM conflict_reviews WHERE check_id = ?", w.checkID).Scan(&reviewCount).Error)
	require.Zero(t, reviewCount, "fixture starts without any independent review")

	_, err := w.svc.CreateAndSubmitApproval("u-lawyer-c4", "律师甲", w.createReq())
	require.Error(t, err)
	var gateErr *ApprovalConflictGateError
	require.ErrorAs(t, err, &gateErr)
	require.Equal(t, "CONFLICT_REVIEW_REQUIRED", gateErr.Code)
	requirePGC4ZeroResidue(t, w)
}

func TestPGC4InitialSubmissionIncompleteCoverageFailsClosed(t *testing.T) {
	w := newPGC4World(t, "approval_submitting", true)
	require.NoError(t, w.db.Exec(
		"UPDATE conflict_check_records SET check_result = jsonb_build_object('decision', jsonb_build_object('coverageStatus', 'PARTIAL')) WHERE check_id = ?",
		w.checkID).Error)

	_, err := w.svc.CreateAndSubmitApproval("u-lawyer-c4", "律师甲", w.createReq())
	require.Error(t, err)
	var gateErr *ApprovalConflictGateError
	require.ErrorAs(t, err, &gateErr)
	require.Equal(t, "CONFLICT_COVERAGE_INCOMPLETE", gateErr.Code)
	requirePGC4ZeroResidue(t, w)
}

// TestPGC4InitialSubmissionRaceOnlyCurrentFactsWin proves the row lock
// contract: a concurrent fact edit either commits after the whole
// submission (snapshot carries the exact pre-edit facts and version) or
// invalidates the binding before submission (fail closed). A torn
// snapshot mixing versions must never exist.
func TestPGC4InitialSubmissionRaceOnlyCurrentFactsWin(t *testing.T) {
	w := newPGC4World(t, "conflict_ready", true)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(30 * time.Millisecond)
		w.db.Exec(
			"UPDATE case_intakes SET title = 'AFTER_RACE_TITLE', description = 'AFTER_RACE_DESC', "+
				"metadata = jsonb_set(metadata, '{lawyer_facts_version}', to_jsonb('FV-C4-PG-RACE'::text)), updated_at = NOW() WHERE id = ?",
			w.intakeID)
	}()

	created, submitErr := w.svc.CreateAndSubmitApproval("u-lawyer-c4", "律师甲", w.createReq())
	wg.Wait()
	if submitErr != nil {
		// The edit won the lock first and invalidated the binding: fail
		// closed with zero residue is the only acceptable outcome.
		var gateErr *ApprovalConflictGateError
		require.ErrorAs(t, submitErr, &gateErr)
		requirePGC4ZeroResidue(t, w)
		return
	}

	var snap struct {
		PayloadData string
	}
	require.NoError(t, w.db.Raw("SELECT snapshot_data AS payload_data FROM approval_snapshots WHERE approval_request_id = ? AND snapshot_type = 'intake_submission'", created.ID).Take(&snap).Error)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.PayloadData), &payload))
	structured := payload["structured_facts"].(map[string]interface{})
	title, _ := structured["title"].(string)
	factsVersion, _ := payload["lawyer_facts_version"].(string)
	switch title {
	case "PG C4 原子提交":
		require.Equal(t, pgC4FactsVersion, factsVersion,
			"a pre-edit snapshot must carry the pre-edit facts version, never a mix")
	case "AFTER_RACE_TITLE":
		require.Equal(t, "FV-C4-PG-RACE", factsVersion,
			"a post-edit snapshot must carry the post-edit facts version, never a mix")
	default:
		t.Fatalf("snapshot title %q is neither committed fact version: torn read", title)
	}
}
