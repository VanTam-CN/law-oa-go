package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
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

// N3/C12+C14 PostgreSQL isolation evidence (plan §11: C12-C14 must be
// proven on real isolated PG transactions, not only SQLite/mock). All tests
// skip unless INTAKE_MVP_PG_DSN is provided so default CI/dev runs stay
// green. The DSN must point at a DISPOSABLE database: the suite is run
// against a throwaway clone that is dropped after the run. conflict_check_records
// and conflict_reviews are append-only evidence (DB trigger rejects UPDATE/DELETE,
// SQLSTATE P0001); tests must never delete or disable that protection. The
// synthetic evidence rows they accumulate are expected output of each run and
// are discarded together with the clone. Only mutable fixtures (inbox,
// tracking, snapshots, associations, approvals, records, intakes, users) are
// cleaned per exact generated ID, with every cleanup error asserted.
// Schema facts verified against the isolated database:
// case_intakes.id is native uuid, approval_snapshots.id is uuid,
// approval_requests/approval_records ids are varchar(36), jsonb parameters
// need explicit casts, approval_case_creation_tracking.intake_id carries the
// 000081 UNIQUE intake-slot index, and approval_snapshots carries the 000082
// partial unique index on (approval_request_id, snapshot_type,
// source_version) WHERE snapshot_type='intake_submission'.

const pgC12WorkflowType = "INTAKE_PG_NO_MATCH"

type pgResubmitWorld struct {
	db         *gorm.DB
	svc        *ApprovalService
	intakeID   string
	approvalID string
	recID      string
	suffix     string
	checkID    string
	adminName  string
	// snapshotID pins the seeded first-round submission snapshot.
	snapshotID string
}

func newPGResubmitWorld(t *testing.T) *pgResubmitWorld {
	t.Helper()
	dsn := os.Getenv("INTAKE_MVP_PG_DSN")
	if dsn == "" {
		t.Skip("INTAKE_MVP_PG_DSN not set; PG isolation evidence runs separately")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)

	w := &pgResubmitWorld{
		db:  db,
		svc: NewApprovalService(db),
		// 相邻轮次可能在同一时钟读数创建夹具，纳秒后缀有同刻碰撞
		// 风险；UUID 派生后缀保证多轮隔离（与产品 RN 缺陷同类）。
		suffix: strings.ReplaceAll(uuid.NewString(), "-", "")[:16],
	}
	w.intakeID = uuid.NewString()
	w.approvalID = "APR-PG-C12-" + w.suffix
	w.recID = "REC-PG-C12-" + w.suffix
	w.adminName = "admin-pg-" + w.suffix

	// t.Cleanup runs LIFO: registering the full purge first lets any
	// later fixture-level cleanups run before it. This purge is the
	// authoritative sweep for the world's exact IDs; every delete error
	// and every surviving row fails the test, so cleanup can no longer
	// fail silently (C12 leak fix 2026-09-28). Append-only evidence
	// tables (conflict_check_records, conflict_reviews) are intentionally
	// NOT touched: the DB trigger forbids deletes and the rows are
	// synthetic audit data discarded with the disposable clone.
	t.Cleanup(func() { w.purge(t) })

	// Deterministic revival path: no workflow matches the marker type, so
	// AssignApprover falls back to the default approver (first active admin
	// who is not the applicant). Seed that admin explicitly.
	require.NoError(t, db.Exec(
		"INSERT INTO users (username, name, email, password, role, status, department, seniority) VALUES (?, ?, ?, 'x', 'admin', 'active', '综合部', '初级')",
		w.adminName, "审批管理员", fmt.Sprintf("%s@example.com", w.adminName)).Error)

	// case_intakes: uuid PK, intake_code/title NOT NULL, jsonb cast required.
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, intake_code, title, status, metadata, created_at, updated_at) "+
			"VALUES (?, ?, ?, 'needs_revision', ?::jsonb, NOW(), NOW())",
		w.intakeID, "C12-"+w.suffix, "C12 并发重提", `{"draft_revision":"rev-1"}`).Error)

	// approval_requests: varchar(36) id; seed all JSONB columns the revival
	// tx.Save writes back, so empty Go strings do not violate jsonb input
	// syntax.
	require.NoError(t, db.Exec(
		"INSERT INTO approval_requests (id, request_number, title, type, content, applicant_id, applicant_name, status, workflow_type, current_approver_id, created_by, metadata, workflow_config, attachments, conflict_result, created_at, updated_at) "+
			"VALUES (?, ?, 'C12', 'CASE_CREATION', 'c12', 'u-lawyer', 'lawyer', 'needs_revision', ?, 'u-approver', 'u-lawyer', ?::jsonb, '{}', '[]', '{}', NOW(), NOW())",
		w.approvalID, "RN-"+w.approvalID, pgC12WorkflowType, fmt.Sprintf(`{"intake_id":"%s"}`, w.intakeID)).Error)

	// The latest request_changes record: the structured resubmit anchor.
	require.NoError(t, db.Exec(
		"INSERT INTO approval_records (id, approval_request_id, stage, stage_order, approver_id, approver_name, decision, decision_reason, imposed_requirements, follow_up_actions, approval_date, supporting_documents, evidence_references, status) "+
			"VALUES (?, ?, 'stage-1', 1, 'u-approver', 'approver', 'request_changes', '退回', ?::jsonb, '[]'::jsonb, NOW(), '[]'::jsonb, '[]'::jsonb, 'active')",
		w.recID, w.approvalID, `{"schema_version":1,"items":[{"item_id":"item-1","category":"field_correction","field_key":"case_title","description":"标题","required":true}]}`).Error)
	w.checkID = seedPGConflictChain(t, db, w.intakeID, w.approvalID)
	seedPGActiveSubmissionSnapshot(t, db, w)
	t.Logf("c12 world ids: intake=%s approval=%s record=%s admin=%s check=%s snapshot=%s",
		w.intakeID, w.approvalID, w.recID, w.adminName, w.checkID, w.snapshotID)
	return w
}

// purge removes every row this world created, in foreign-key-safe
// child-before-parent order, scoped to this world's exact generated IDs
// only (no broad or business-data deletes). It logs per-table RowsAffected
// for the run report, then verifies zero residue per table; any error or
// surviving row fails the test.
func (w *pgResubmitWorld) purge(t *testing.T) {
	t.Helper()
	del := func(label, query string, args ...interface{}) int64 {
		res := w.db.Exec(query, args...)
		require.NoError(t, res.Error, "c12 cleanup %s delete failed", label)
		return res.RowsAffected
	}
	inbox := del("inbox", "DELETE FROM inbox_items WHERE source_ref = ?", w.approvalID)
	assoc := del("association", "DELETE FROM approval_conflict_associations WHERE approval_request_id = ?", w.approvalID)
	tracking := del("tracking", "DELETE FROM approval_case_creation_tracking WHERE approval_request_id = ? OR intake_id = ?", w.approvalID, w.intakeID)
	snaps := del("snapshot", "DELETE FROM approval_snapshots WHERE approval_request_id = ?", w.approvalID)
	records := del("approval-record", "DELETE FROM approval_records WHERE approval_request_id = ?", w.approvalID)
	approvals := del("approval", "DELETE FROM approval_requests WHERE id = ?", w.approvalID)
	intakes := del("intake", "DELETE FROM case_intakes WHERE id = ?", w.intakeID)
	admins := del("admin-user", "DELETE FROM users WHERE username = ?", w.adminName)
	t.Logf("c12 purge rows: inbox=%d assoc=%d tracking=%d snapshots=%d records=%d approvals=%d intakes=%d admins=%d (immutable conflict evidence check=%s stays for the disposable clone)",
		inbox, assoc, tracking, snaps, records, approvals, intakes, admins, w.checkID)

	count := func(label, query string, args ...interface{}) {
		var n int64
		require.NoError(t, w.db.Raw("SELECT COUNT(*) FROM "+query, args...).Scan(&n).Error,
			"c12 cleanup %s verify query failed", label)
		require.Zero(t, n, "c12 cleanup residue in %s", label)
	}
	count("inbox", "inbox_items WHERE source_ref = ?", w.approvalID)
	count("association", "approval_conflict_associations WHERE approval_request_id = ?", w.approvalID)
	count("tracking", "approval_case_creation_tracking WHERE approval_request_id = ? OR intake_id = ?", w.approvalID, w.intakeID)
	count("snapshot", "approval_snapshots WHERE approval_request_id = ?", w.approvalID)
	count("approval-record", "approval_records WHERE approval_request_id = ?", w.approvalID)
	count("approval", "approval_requests WHERE id = ?", w.approvalID)
	count("intake", "case_intakes WHERE id = ?", w.intakeID)
	count("admin-user", "users WHERE username = ?", w.adminName)
}

// seedPGConflictChain builds the post-commit conflict evidence the plan 6.4
// resubmit gate reads on the real isolated PG schema: committed check pointer
// plus facts-version binding in intake metadata, a completed immutable record
// naming this intake, an independent review, and the active association the
// final approval gate resolves first.
func seedPGConflictChain(t *testing.T, db *gorm.DB, intakeID, approvalID string) string {
	t.Helper()
	checkID := "CCT-PG-" + intakeID[:8] + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
	factsVersion := "FV-PG-" + checkID
	intakeMeta := fmt.Sprintf(
		`{"draft_revision":"rev-1","conflict_check_id":"%s","lawyer_facts_version":"%s","conflict_check_facts_version":"%s"}`,
		checkID, factsVersion, factsVersion)
	require.NoError(t, db.Exec(
		"UPDATE case_intakes SET status = 'conflict_ready', metadata = ?::jsonb, updated_at = NOW() WHERE id = ?",
		intakeMeta, intakeID).Error)
	require.NoError(t, db.Create(&models.ConflictCheckRecord{
		CheckID:          checkID,
		ClientID:         "0",
		ClientName:       "pg-gate-customer",
		CaseName:         "PG 重提门禁",
		CaseType:         "民事",
		CheckStatus:      "COMPLETED",
		SearchParameters: models.JSON{"intakeId": intakeID, "clientId": "0"},
		CheckResult:      models.JSON{"decision": map[string]interface{}{"coverageStatus": "COMPLETE"}},
		UserID:           1,
		CheckTime:        time.Now(),
	}).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_reviews (check_id, decision, notes, reviewer_id, reviewer_name, evidence_hash, created_at) "+
			"VALUES (?, 'no_conflict', '独立复核通过', 1, '核查人', 'pg-hash', NOW())", checkID).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO approval_conflict_associations (id, approval_request_id, conflict_check_id, association_status, created_by, created_at, updated_at) "+
			"VALUES (?, ?, ?, 'active', 'u-lawyer', NOW(), NOW())",
		"ASSOC-PG-"+checkID, approvalID, checkID).Error)
	// Only the mutable association row is cleaned, by its exact seeded
	// primary key; the world purge verifies zero residue afterwards.
	// conflict_reviews / conflict_check_records are append-only evidence
	// (SQLSTATE P0001 on delete) and are never deleted here — they are
	// synthetic audit data discarded with the disposable clone. Errors
	// must fail the test instead of being ignored.
	t.Cleanup(func() {
		res := db.Exec("DELETE FROM approval_conflict_associations WHERE id = ?", "ASSOC-PG-"+checkID)
		require.NoError(t, res.Error, "seed conflict-chain association cleanup failed")
	})
	return checkID
}

// seedPGActiveSubmissionSnapshot mints the first-round intake-submission
// snapshot on the real PG schema and flips the active pointer, mirroring the
// production initial submission the C12 resubmit scenarios presuppose.
func seedPGActiveSubmissionSnapshot(t *testing.T, db *gorm.DB, w *pgResubmitWorld) {
	t.Helper()
	var metaRow struct {
		Metadata string
	}
	require.NoError(t, db.Raw("SELECT metadata FROM case_intakes WHERE id = ?", w.intakeID).Scan(&metaRow).Error)
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(metaRow.Metadata), &metadata))
	checkID, _ := metadata["conflict_check_id"].(string)
	require.NotEmpty(t, checkID)
	payload, err := json.Marshal(map[string]interface{}{
		"approval_id":       w.approvalID,
		"intake_id":         w.intakeID,
		"intake_revision":   "rev-1",
		"round_number":      1,
		"conflict_check_id": checkID,
		"client_id":         "0",
	})
	require.NoError(t, err)
	w.snapshotID = uuid.NewString()
	require.NoError(t, db.Exec(
		"INSERT INTO approval_snapshots (id, approval_request_id, snapshot_type, snapshot_data, source_version) VALUES (?, ?, ?, ?::jsonb, 1)",
		w.snapshotID, w.approvalID, models.ApprovalSnapshotTypeIntakeSubmission, string(payload)).Error)
	require.NoError(t, db.Exec(
		"UPDATE approval_requests SET active_submission_snapshot_id = ?, updated_at = NOW() WHERE id = ?",
		w.snapshotID, w.approvalID).Error)
}

func (w *pgResubmitWorld) validRequest() *ResubmitApprovalRequest {
	return &ResubmitApprovalRequest{
		ExpectedIntakeRevision:       "rev-1",
		ReturnRecordID:               w.recID,
		RevisionNote:                 "PG 取证重提",
		Replies:                      []SupplementReply{{ItemID: "item-1", Response: "已补充"}},
		ExpectedSubmissionSnapshotID: w.snapshotID,
	}
}

func (w *pgResubmitWorld) snapshotCount() int64 {
	var n int64
	cnt := w.db.Table("approval_snapshots").
		Where("approval_request_id = ? AND snapshot_type = ?", w.approvalID, models.ApprovalSnapshotTypeIntakeSubmission).
		Count(&n).Error
	if cnt != nil {
		panic("snapshot count query failed: " + cnt.Error())
	}
	return n
}

func (w *pgResubmitWorld) approvalStatus() string {
	var status string
	err := w.db.Raw("SELECT status FROM approval_requests WHERE id = ?", w.approvalID).Scan(&status).Error
	if err != nil {
		panic("approval status query failed: " + err.Error())
	}
	return status
}

func (w *pgResubmitWorld) bindCount(approvalID string) int64 {
	var n int64
	err := w.db.Table("approval_case_creation_tracking").
		Where("approval_request_id = ?", approvalID).Count(&n).Error
	if err != nil {
		panic("bind count query failed: " + err.Error())
	}
	return n
}

// C12 scenario 1 (identical content): two racing structured resubmits
// serialize on the intake row lock. The winner creates the round; the
// identical second request resolves as an idempotent replay of the same
// round (C11) with zero additional side effects — one snapshot, one bind
// row, resubmitted status.
func TestIntakeMVP_PG_ConcurrentResubmitSingleWinner(t *testing.T) {
	// Plan SS11 C12: same-start concurrency must be proven over at least
	// 3 independent rounds. Each subtest seeds its own world (unique
	// fixtures, independent t.Cleanup) and releases both racers through
	// a channel barrier so they truly start together.
	for round := 1; round <= 3; round++ {
		t.Run(fmt.Sprintf("round-%d", round), func(t *testing.T) {
			w := newPGResubmitWorld(t)

			type outcome struct {
				err      error
				snapshot string
			}
			outcomes := make([]outcome, 2)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					<-start
					updated, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, w.validRequest())
					if err == nil && updated != nil && updated.ActiveSubmissionSnapshotID != nil {
						outcomes[idx].snapshot = *updated.ActiveSubmissionSnapshotID
					}
					outcomes[idx].err = err
				}(i)
			}
			close(start)
			wg.Wait()

			successes := 0
			for _, oc := range outcomes {
				if oc.err == nil {
					successes++
					require.NotEmpty(t, oc.snapshot)
				} else {
					require.True(t, typedResubmitRejection(oc.err), "unexpected error: %v", oc.err)
				}
			}
			require.Equal(t, 2, successes, "winner plus idempotent replay both succeed with zero errors")
			require.Equal(t, int64(2), w.snapshotCount())
			require.Equal(t, int64(1), w.bindCount(w.approvalID))
			require.Equal(t, models.ApprovalStatusResubmitted, w.approvalStatus())
		})
	}
}

// C12 scenario 1b (different content): racing resubmits with different
// payloads serialize on the intake row lock; the loser is a true
// competitor and must be rejected with the typed stale 409, leaving the
// winner's round untouched.
func TestIntakeMVP_PG_ConcurrentDifferentContentOneWinner(t *testing.T) {
	// Plan SS11 C12: same-start concurrency over 3 independent rounds, with
	// a channel barrier releasing both rivals at the same instant.
	for round := 1; round <= 3; round++ {
		t.Run(fmt.Sprintf("round-%d", round), func(t *testing.T) {
			w := newPGResubmitWorld(t)

			rival := w.validRequest()
			rival.RevisionNote = "竞争版本：不同的重提内容"
			rival.Replies[0].Response = "竞争者提交的不同回复"

			errs := make([]error, 2)
			requests := []*ResubmitApprovalRequest{w.validRequest(), rival}
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					<-start
					_, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, requests[idx])
					errs[idx] = err
				}(i)
			}
			close(start)
			wg.Wait()

			wins := 0
			for _, err := range errs {
				if err == nil {
					wins++
				} else {
					var contractErr *ResubmitContractError
					require.ErrorAs(t, err, &contractErr)
					require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)
				}
			}
			require.Equal(t, 1, wins, "exactly one rival resubmit may win")
			require.Equal(t, int64(2), w.snapshotCount())
			require.Equal(t, models.ApprovalStatusResubmitted, w.approvalStatus())
		})
	}
}

// C12 scenario 2a: the resubmit runs first, then a competing NEW approval
// creation tries to bind the same intake. The 000081 UNIQUE index must
// reject the newcomer with the 409 contract, and the resubmitted approval
// keeps the slot.
func TestIntakeMVP_PG_ResubmitWinsThenNewApprovalRejected(t *testing.T) {
	w := newPGResubmitWorld(t)

	_, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, w.validRequest())
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, w.approvalStatus())
	require.Equal(t, int64(1), w.bindCount(w.approvalID), "revival owns the intake slot")

	// Competing new approval creation binds the same intake: the
	// approval_case_creation_tracking UNIQUE index on intake_id is the same
	// database invariant the live conflict-integration creation path enforces
	// when a newcomer tries to claim an already-bound intake.
	competitorIntake := w.intakeID
	bindErr := w.db.Exec(
		"INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, creation_status, error_details, data_mapping, mapped_fields, unmapped_fields, applied_conditions, imposed_requirements, workflow_actions, created_by) "+
			"VALUES (?, ?, ?, 'pending', '{}', '{}', '{}', '{}', '{}', '{}', '[]', 'u-lawyer-2')",
		"INTAKE_BIND_APR-PG-COMP2-"+w.suffix, "APR-PG-COMP2-"+w.suffix, competitorIntake).Error
	require.Error(t, bindErr)
	require.Contains(t, bindErr.Error(), "duplicate key value violates unique constraint")

	// The winner still owns the slot; no competing bind row survived.
	require.Equal(t, int64(1), w.bindCount(w.approvalID))
	require.Equal(t, models.ApprovalStatusResubmitted, w.approvalStatus())
}

// C12 scenario 2b: a competing NEW approval creation binds the intake first;
// the structured resubmit must then be refused with the 409 contract and
// leave zero snapshot rows plus the original needs_revision state.
func TestIntakeMVP_PG_NewApprovalWinsThenResubmitRejected(t *testing.T) {
	w := newPGResubmitWorld(t)

	// Competitor binds the intake slot first; the same UNIQUE guard backs
	// the live creation path.
	competitorBindID := "INTAKE_BIND_APR-PG-COMP-" + w.suffix
	require.NoError(t, w.db.Exec(
		"INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, creation_status, error_details, data_mapping, mapped_fields, unmapped_fields, applied_conditions, imposed_requirements, workflow_actions, created_by) "+
			"VALUES (?, ?, ?, 'pending', '{}', '{}', '{}', '{}', '{}', '{}', '[]', 'u-lawyer-2')",
		competitorBindID, "APR-PG-COMP-"+w.suffix, w.intakeID).Error)
	t.Cleanup(func() {
		res := w.db.Exec("DELETE FROM approval_case_creation_tracking WHERE id = ?", competitorBindID)
		require.NoError(t, res.Error, "competitor bind fixture cleanup failed")
	})

	_, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, w.validRequest())
	require.Error(t, err)
	var gateErr *ApprovalConflictGateError
	require.ErrorAs(t, err, &gateErr)
	require.Equal(t, "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", gateErr.Code)

	// The failed revival must roll back completely: the seeded first-round
	// snapshot survives, no status change, no self bind row.
	require.Equal(t, int64(1), w.snapshotCount())
	require.Equal(t, models.ApprovalStatusNeedsRevision, w.approvalStatus())
	require.Equal(t, int64(0), w.bindCount(w.approvalID))
}

// C14: a forced revival-write failure must roll the whole resubmit
// transaction back (no snapshot, no pointer, no status change). The fault
// is a residual INTAKE_BIND row for this approval (crashed-attempt
// leftover), so the UNIQUE intake guard fires inside the revival step
// AFTER the snapshot write — proving the transaction boundary covers every
// side effect. After removing the fault, the same request succeeds once.
func TestIntakeMVP_PG_SnapshotFailureRollsBackThenRetrySucceeds(t *testing.T) {
	w := newPGResubmitWorld(t)

	// Fault injection: a residual bind row for THIS approval already holds
	// the intake slot, so the revival insert collides on 000081.
	residualBindID := "INTAKE_BIND_" + w.approvalID
	require.NoError(t, w.db.Exec(
		"INSERT INTO approval_case_creation_tracking (id, approval_request_id, intake_id, creation_status, error_details, data_mapping, mapped_fields, unmapped_fields, applied_conditions, imposed_requirements, workflow_actions, created_by) "+
			"VALUES (?, ?, ?, 'pending', '{}', '{}', '{}', '{}', '{}', '{}', '[]', 'u-legacy')",
		residualBindID, w.approvalID, w.intakeID).Error)
	t.Cleanup(func() {
		res := w.db.Exec("DELETE FROM approval_case_creation_tracking WHERE id = ?", residualBindID)
		require.NoError(t, res.Error, "residual bind fixture cleanup failed")
	})

	_, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, w.validRequest())
	var gateErr *ApprovalConflictGateError
	require.ErrorAs(t, err, &gateErr)
	require.Equal(t, "INTAKE_APPROVAL_ALREADY_IN_PROGRESS", gateErr.Code)

	// Whole-transaction rollback: the revival snapshot is discarded and only
	// the seeded first round survives.
	require.Equal(t, int64(1), w.snapshotCount())
	require.Equal(t, models.ApprovalStatusNeedsRevision, w.approvalStatus())
	require.Equal(t, int64(1), w.bindCount(w.approvalID),
		"rollback restores prior state; the residual fault row is untouched")

	// Clear the fault and retry the identical request.
	require.NoError(t, w.db.Exec("DELETE FROM approval_case_creation_tracking WHERE id = ?", residualBindID).Error)

	updated, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, w.validRequest())
	require.NoError(t, err)
	require.NotEmpty(t, updated.ActiveSubmissionSnapshotID)
	require.Equal(t, int64(2), w.snapshotCount())
	require.Equal(t, models.ApprovalStatusResubmitted, w.approvalStatus())
	require.Equal(t, int64(1), w.bindCount(w.approvalID))
}

// N4/C13 (PG isolation): a database-level fault in the inbox handoff write
// rolls the whole resubmit back -- no snapshot, no revival, no handoff row.
// After the fault is removed the identical request retries successfully and
// opens exactly one approver handoff.
func TestIntakeMVP_PG_InboxHandoffFaultRollsBackResubmit(t *testing.T) {
	w := newPGResubmitWorld(t)

	// The applicant actor is a username; map it to a real active user so
	// the revision handoff (closed inside the tx) can resolve its owner.
	lawyerName := "u-lawyer-" + w.suffix
	require.NoError(t, w.db.Exec(
		"INSERT INTO users (username, name, email, password, role, status, department, seniority) VALUES (?, ?, ?, 'x', 'lawyer', 'active', '综合部', '初级')",
		lawyerName, "律师", fmt.Sprintf("%s@example.com", lawyerName)).Error)
	t.Cleanup(func() {
		res := w.db.Exec("DELETE FROM users WHERE username = ?", lawyerName)
		require.NoError(t, res.Error, "c13 lawyer fixture cleanup failed")
	})

	// The default approver fallback picks the first active non-applicant
	// admin; resolve the same user the revival tx will assign.
	var adminID int
	require.NoError(t, w.db.Raw(
		"SELECT id FROM users WHERE role = 'admin' AND status = 'active' ORDER BY id LIMIT 1").Scan(&adminID).Error)

	// Database-level fault: any inbox handoff write aborts the tx.
	require.NoError(t, w.db.Exec(
		"CREATE FUNCTION intake_mvp_c13_fault() RETURNS trigger AS $$ BEGIN RAISE EXCEPTION 'intake-mvp-c13 fault'; END; $$ LANGUAGE plpgsql").Error)
	require.NoError(t, w.db.Exec(
		"CREATE TRIGGER trg_intake_mvp_c13_fault BEFORE INSERT ON inbox_items FOR EACH ROW EXECUTE FUNCTION intake_mvp_c13_fault()").Error)
	t.Cleanup(func() {
		require.NoError(t, w.db.Exec("DROP TRIGGER IF EXISTS trg_intake_mvp_c13_fault ON inbox_items").Error, "c13 fault trigger drop failed")
		require.NoError(t, w.db.Exec("DROP FUNCTION IF EXISTS intake_mvp_c13_fault()").Error, "c13 fault function drop failed")
	})

	_, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, w.validRequest())
	require.Error(t, err)

	// Whole-transaction rollback: only the seeded first-round snapshot
	// remains, approval stays needs_revision,
	// and no handoff row leaked for this approval.
	require.Equal(t, int64(1), w.snapshotCount())
	require.Equal(t, models.ApprovalStatusNeedsRevision, w.approvalStatus())
	var handoffRows int64
	require.NoError(t, w.db.Table("inbox_items").
		Where("source_ref = ?", w.approvalID).Count(&handoffRows).Error)
	require.Zero(t, handoffRows)

	// Remove the fault and retry the identical request.
	require.NoError(t, w.db.Exec("DROP TRIGGER trg_intake_mvp_c13_fault ON inbox_items").Error)
	require.NoError(t, w.db.Exec("DROP FUNCTION intake_mvp_c13_fault()").Error)

	updated, err := w.svc.StructuredResubmit("u-lawyer", w.approvalID, w.validRequest())
	require.NoError(t, err)
	require.NotEmpty(t, updated.ActiveSubmissionSnapshotID)
	require.Equal(t, int64(2), w.snapshotCount())
	require.Equal(t, models.ApprovalStatusResubmitted, w.approvalStatus())

	// The new round's approver handoff is open exactly once for the
	// assigned admin (cycle 1: one historical request_changes record).
	require.NoError(t, w.db.Table("inbox_items").
		Where("workflow_task_key = ? AND is_completed = ?",
			fmt.Sprintf("approval:approval_decision:%s:c1:%d", w.approvalID, adminID), false).
		Count(&handoffRows).Error)
	require.Equal(t, int64(1), handoffRows)
	t.Cleanup(func() {
		res := w.db.Exec("DELETE FROM inbox_items WHERE source_ref = ?", w.approvalID)
		require.NoError(t, res.Error, "c13 handoff fixture cleanup failed")
	})
}

// typedResubmitRejection reports whether the caller surfaced a typed 4xx.
func typedResubmitRejection(err error) bool {
	var contractErr *ResubmitContractError
	var gateErr *ApprovalConflictGateError
	return errors.As(err, &contractErr) || errors.As(err, &gateErr)
}

// TestPGSupplementReplyDraft_ConcurrentResubmitGuard proves the draft-save
// transactional guard on the real isolated PG schema: the conditional row
// lock rejects the save once a concurrent resubmit flips the approval out of
// needs_revision, and RowsAffected stays the authoritative signal.
func TestPGSupplementReplyDraft_ConcurrentResubmitGuard(t *testing.T) {
	w := newPGResubmitWorld(t)
	req := w.validRequest()

	// Manual sim of the racing resubmit: flip the status first, then attempt
	// the draft save. The conditional UPDATE ... WHERE status= needs_revision
	// must match zero rows and the save must fail with the typed stale error.
	require.NoError(t, w.db.Exec(
		"UPDATE approval_requests SET status = 'resubmitted', updated_at = NOW() WHERE id = ?",
		w.approvalID).Error)

	_, err := w.svc.SaveSupplementReplies("u-lawyer", w.approvalID, req)
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)

	// Nothing may be stored on the closed round.
	var stored string
	require.NoError(t, w.db.Raw("SELECT supplement_replies FROM approval_records WHERE id = ?", w.recID).Scan(&stored).Error)
	require.NotContains(t, stored, "item-1")

	// The happy path also proves RowsAffected>0 writes once: restore the
	// needs_revision status and save; the stored JSONB must carry the replies.
	require.NoError(t, w.db.Exec(
		"UPDATE approval_requests SET status = 'needs_revision', updated_at = NOW() WHERE id = ?",
		w.approvalID).Error)
	saved, err := w.svc.SaveSupplementReplies("u-lawyer", w.approvalID, req)
	require.NoError(t, err)
	require.NotNil(t, saved.SupplementReplies)
	require.NotEmpty(t, *saved.SupplementReplies)
	require.NoError(t, w.db.Raw("SELECT supplement_replies FROM approval_records WHERE id = ?", w.recID).Scan(&stored).Error)
	require.Contains(t, stored, "已补充")
}
