package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// N3 / C09+C11 acceptance: atomic structured resubmit with server-owned snapshots.
// Success persists the snapshot and flips the active pointer inside the revival
// transaction; stale revision, snapshot, or reply inputs are rejected with typed
// 4xx codes and leave no partial state.

func seedResubmitContractWorld(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/resubmit_contract.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE approval_snapshots (id TEXT PRIMARY KEY, approval_request_id TEXT, snapshot_type TEXT, snapshot_data TEXT, source_version INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)").Error)
	require.NoError(t, db.Exec("CREATE UNIQUE INDEX uq_snap_round ON approval_snapshots (approval_request_id, snapshot_type, source_version)").Error)
	// Conflict evidence tables read by the plan 6.4 resubmit gate. The
	// association table is migrated from the production model so the
	// success-path association sync writes the same columns the real
	// migration 000020 schema has.
	require.NoError(t, db.AutoMigrate(
		&models.ApprovalRequest{},
		&models.ApprovalRecord{},
		&models.ApprovalWorkflow{},
		&models.User{},
		&models.InboxItem{},
		&models.ConflictCheckRecord{},
		&models.ConflictReview{},
		&repositories.ConflictAssociation{},
		&repositories.CaseCreationTracking{},
	))
	// Resubmit re-assigns the approver; the default path needs an active
	// admin who is not the applicant.
	require.NoError(t, db.Create(&models.User{
		Username: fmt.Sprintf("admin-%d", time.Now().UnixNano()),
		Name:     "审批管理员",
		Email:    fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()),
		Password: "x",
		Role:     "admin",
		Status:   "active",
	}).Error)
	return db
}

func seedResubmitContractApproval(t *testing.T, db *gorm.DB, intakeID string) (*models.ApprovalRequest, *models.ApprovalRecord) {
	t.Helper()
	checkID := fmt.Sprintf("CCT_RC_%d", time.Now().UnixNano())
	// Preserve the intake's current draft revision: several tests seed a
	// revision first to prove the stale-revision rejection.
	var existing struct {
		Metadata []byte
	}
	require.NoError(t, db.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&existing).Error)
	intakeMeta := map[string]interface{}{}
	if len(existing.Metadata) > 0 {
		_ = json.Unmarshal(existing.Metadata, &intakeMeta)
	}
	revision, _ := intakeMeta["draft_revision"].(string)
	if strings.TrimSpace(revision) == "" {
		revision = "rev-1"
	}
	approval := &models.ApprovalRequest{
		ID:                fmt.Sprintf("APR-RC-%d", time.Now().UnixNano()),
		RequestNumber:     fmt.Sprintf("RN-RC-%d", time.Now().UnixNano()),
		Title:             "structured resubmit contract",
		Type:              "CASE_CREATION",
		Content:           "c09 c11",
		ApplicantID:       "u-lawyer",
		ApplicantName:     "lawyer",
		Status:            models.ApprovalStatusNeedsRevision,
		WorkflowType:      "CASE_APPROVAL",
		CurrentApproverID: "u-approver",
		CreatedBy:         "u-lawyer",
		Metadata:          "{\"intake_id\":\"INTAKE_PLACEHOLDER\"}",
	}
	require.NoError(t, db.Create(approval).Error)
	approval.Metadata = ""
	// Set metadata now that intakeID and the seeded check binding are known.
	approval.Metadata = strings.ReplaceAll(buildIntakeMetadata(approval, intakeID), "CHECK_ID_PLACEHOLDER", checkID)
	require.NoError(t, db.Save(approval).Error)
	seedIntakeConflictBinding(t, db, intakeID, checkID, map[string]string{"draft_revision": revision})
	rec := &models.ApprovalRecord{
		ID:                  fmt.Sprintf("REC-RC-%d", time.Now().UnixNano()),
		ApprovalRequestID:   approval.ID,
		Stage:               "stage-1",
		StageOrder:          1,
		ApproverID:          "u-approver",
		Decision:            models.ApprovalDecisionRequestChanges,
		DecisionReason:      "请修正标题",
		ImposedRequirements: "{\"schema_version\":1,\"items\":[{\"item_id\":\"item-1\",\"category\":\"field_correction\",\"field_key\":\"case_title\",\"description\":\"标题不一致\",\"required\":true}]}",
		FollowUpActions:     "[]",
		ApprovalDate:        time.Now(),
		SupportingDocuments: "[]",
		EvidenceReferences:  "[]",
		Status:              "active",
	}
	require.NoError(t, db.Create(rec).Error)
	return approval, rec
}

// buildIntakeMetadata builds the approval metadata binding to the intake.
func buildIntakeMetadata(approval *models.ApprovalRequest, intakeID string) string {
	return "{\"intake_id\":\"" + intakeID + "\",\"conflict_check_id\":\"CHECK_ID_PLACEHOLDER\"}"
}

func seedResubmitContractIntake(t *testing.T, db *gorm.DB, intakeID, revision string) {
	t.Helper()
	require.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS case_intakes (id TEXT PRIMARY KEY, client_id TEXT, status TEXT, metadata TEXT, updated_at DATETIME)").Error)
	meta := buildIntakeRevisionMetadata(revision)
	require.NoError(t, db.Exec("INSERT INTO case_intakes (id, client_id, status, metadata, updated_at) VALUES (?, NULL, ?, ?, CURRENT_TIMESTAMP)", intakeID, "needs_revision", meta).Error)
}

func buildIntakeRevisionMetadata(revision string) string {
	return "{\"draft_revision\":\"" + revision + "\"}"
}

// seedActiveSubmissionRound mints the first-round snapshot and flips the
// active pointer, mirroring the production initial submission these contract
// tests presuppose (the approval arrives already returned-for-revision from
// a submitted round). Returns the minted snapshot id.
func seedActiveSubmissionRound(t *testing.T, db *gorm.DB, approval *models.ApprovalRequest, intakeID string) string {
	t.Helper()
	var metaRow struct {
		Metadata []byte
	}
	require.NoError(t, db.Raw("SELECT metadata FROM approval_requests WHERE id = ?", approval.ID).Scan(&metaRow).Error)
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal(metaRow.Metadata, &metadata))
	checkID, _ := metadata["conflict_check_id"].(string)
	if checkID == "" {
		checkID, _ = metadata["conflict_task_id"].(string)
	}
	require.NotEmpty(t, checkID)
	payload, err := json.Marshal(map[string]interface{}{
		"approval_id":       approval.ID,
		"intake_id":         intakeID,
		"intake_revision":   "rev-1",
		"round_number":      1,
		"conflict_check_id": checkID,
		"client_id":         "9",
	})
	require.NoError(t, err)
	snapID := fmt.Sprintf("SNAP-RC1-%d-%s", time.Now().UnixNano(), approval.ID)
	require.NoError(t, db.Exec(
		"INSERT INTO approval_snapshots (id, approval_request_id, snapshot_type, snapshot_data, source_version) VALUES (?, ?, ?, ?, 1)",
		snapID, approval.ID, models.ApprovalSnapshotTypeIntakeSubmission, string(payload)).Error)
	require.NoError(t, db.Model(&models.ApprovalRequest{}).Where("id = ?", approval.ID).
		Update("active_submission_snapshot_id", snapID).Error)
	approval.ActiveSubmissionSnapshotID = &snapID
	return snapID
}

// conflictSeeds holds the server-owned evidence rows a resubmit gate reads.
// The committed-check pointer lives in the intake metadata, the check record
// is append-only, and the review ledger is the independent conclusion. The
// seed models the post-commit state produced by CommitIntakeConflictResult:
// status back to conflict_ready, the intake's lawyer_facts_version recorded
// as the committed check's facts version, no pending invalidation stamp, and
// the immutable check record bound to this intake through its search
// parameters.
func seedIntakeConflictBinding(t *testing.T, db *gorm.DB, intakeID, checkID string, meta map[string]string) {
	t.Helper()
	factsVersion := "FV-" + checkID
	metadata := map[string]interface{}{
		"conflict_check_id":            checkID,
		"lawyer_facts_version":         factsVersion,
		"conflict_check_facts_version": factsVersion,
	}
	for k, v := range meta {
		metadata[k] = v
	}
	raw, err := json.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, db.Exec("UPDATE case_intakes SET status = 'conflict_ready', metadata = ? WHERE id = ?", string(raw), intakeID).Error)

	searchParameters, _ := json.Marshal(map[string]interface{}{
		"intakeId": intakeID,
		"clientId": "0",
	})
	coverage := map[string]interface{}{
		"decision": map[string]interface{}{"coverageStatus": "COMPLETE"},
	}
	coverageJSON, _ := json.Marshal(coverage)
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_check_records (check_id, client_id, client_name, check_status, search_parameters, check_result, user_id, check_time, created_at, updated_at) "+
			"VALUES (?, '0', 'gate-customer', 'COMPLETED', ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		checkID, string(searchParameters), string(coverageJSON)).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO conflict_reviews (check_id, decision, notes, reviewer_id, reviewer_name, evidence_hash, created_at) "+
			"VALUES (?, ?, '独立复核通过', 42, '核查人', 'hash', CURRENT_TIMESTAMP)",
		checkID, "no_conflict").Error)
}

func seedConflictWorld(t *testing.T) (*gorm.DB, *models.ApprovalRequest, *models.ApprovalRecord, string, string) {
	t.Helper()
	db := seedResubmitContractWorld(t)
	intakeID := fmt.Sprintf("INT-GATE-%d", time.Now().UnixNano())
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	// The authoritative check binding is the one seeded with the approval;
	// tests must never mint a second, unrelated check ID. A re-check
	// simulation must wire the new check through the active association and
	// the intake pointer instead.
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(approval.Metadata), &metadata))
	checkID, _ := metadata["conflict_check_id"].(string)
	require.NotEmpty(t, checkID)
	return db, approval, rec, intakeID, checkID
}

// seedPostCommitConflictWorld starts from seedConflictWorld and then moves
// the intake into the exact post-commit state CommitIntakeConflictResult
// produces: status conflict_ready, the committed check bound to the current
// lawyer facts version, no pending invalidation, and the immutable record
// naming this intake. Use it for tests that exercise the full gate chain;
// use seedConflictWorld only when the test intentionally breaks one link.
func seedPostCommitConflictWorld(t *testing.T) (*gorm.DB, *models.ApprovalRequest, *models.ApprovalRecord, string, string) {
	t.Helper()
	db, approval, rec, intakeID, checkID := seedConflictWorld(t)
	var row struct {
		Metadata []byte
	}
	require.NoError(t, db.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&row).Error)
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal(row.Metadata, &metadata))
	metadata["lawyer_facts_version"] = "FV-COMMITTED"
	metadata["conflict_check_facts_version"] = "FV-COMMITTED"
	metadata["conflict_invalidated_at"] = nil
	metadata["conflict_invalidated_by"] = nil
	metadata["conflict_invalidation_reason"] = nil
	raw, err := json.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, db.Exec("UPDATE case_intakes SET status = 'conflict_ready', metadata = ? WHERE id = ?", string(raw), intakeID).Error)
	searchParameters, _ := json.Marshal(map[string]interface{}{
		"intakeId": intakeID,
		"clientId": "0",
	})
	require.NoError(t, db.Exec("UPDATE conflict_check_records SET search_parameters = ? WHERE check_id = ?",
		string(searchParameters), checkID).Error)
	return db, approval, rec, intakeID, checkID
}

// seedPostCommitConflictWorld2 seeds a named intake and returns the
// approval/record pair already carrying the full post-commit conflict chain
// (conflict_ready status, committed facts-version binding, intake-bound
// record). Named intake callers (C11 replay, handoff round-trip) use this.
func seedPostCommitConflictWorld2(t *testing.T, db *gorm.DB, intakeID string) (*models.ApprovalRequest, *models.ApprovalRecord) {
	t.Helper()
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	var row struct {
		Metadata []byte
	}
	require.NoError(t, db.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&row).Error)
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal(row.Metadata, &metadata))
	checkID, _ := metadata["conflict_check_id"].(string)
	require.NotEmpty(t, checkID)
	metadata["lawyer_facts_version"] = "FV-COMMITTED"
	metadata["conflict_check_facts_version"] = "FV-COMMITTED"
	delete(metadata, "conflict_invalidated_at")
	delete(metadata, "conflict_invalidated_by")
	delete(metadata, "conflict_invalidation_reason")
	raw, err := json.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, db.Exec("UPDATE case_intakes SET status = 'conflict_ready', metadata = ? WHERE id = ?", string(raw), intakeID).Error)
	searchParameters, _ := json.Marshal(map[string]interface{}{
		"intakeId": intakeID,
		"clientId": "0",
	})
	require.NoError(t, db.Exec("UPDATE conflict_check_records SET search_parameters = ? WHERE check_id = ?",
		string(searchParameters), checkID).Error)
	return approval, rec
}

// clearConflictBinding removes every server-owned conflict anchor so a test
// can prove the fail-closed path when no binding is verifiable.
func clearConflictBinding(t *testing.T, db *gorm.DB, approval *models.ApprovalRequest, intakeID string) {
	t.Helper()
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(approval.Metadata), &metadata))
	checkID, _ := metadata["conflict_check_id"].(string)
	require.NoError(t, db.Exec("DELETE FROM approval_conflict_associations WHERE approval_request_id = ?", approval.ID).Error)
	require.NoError(t, db.Exec("DELETE FROM conflict_reviews WHERE check_id = ?", checkID).Error)
	require.NoError(t, db.Exec("DELETE FROM conflict_check_records WHERE check_id = ?", checkID).Error)
	delete(metadata, "conflict_check_id")
	raw, err := json.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, db.Model(&models.ApprovalRequest{}).Where("id = ?", approval.ID).Update("metadata", string(raw)).Error)
	require.NoError(t, db.Exec("UPDATE case_intakes SET metadata = ? WHERE id = ?", buildIntakeRevisionMetadata("rev-1"), intakeID).Error)
}

func removeCommittedCheckPointer(t *testing.T, db *gorm.DB, intakeID string) {
	t.Helper()
	require.NoError(t, db.Exec("UPDATE case_intakes SET metadata = ? WHERE id = ?", buildIntakeRevisionMetadata("rev-1"), intakeID).Error)
}

// reconfirmFactsAfterEdit models the UpdateCaseIntake -> ConfirmIntakeFacts
// sequence for a lawyer facts edit: the server rolls the intake back to
// draft, stamps the invalidation, drops the facts version, and keeps the
// stale check pointer; confirmation then rotates the facts version and moves
// to lawyer_facts_confirmed. No new formal check has committed yet.
func reconfirmFactsAfterEdit(t *testing.T, db *gorm.DB, intakeID string) {
	t.Helper()
	var row struct {
		Metadata []byte
	}
	require.NoError(t, db.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&row).Error)
	metadata := map[string]interface{}{}
	if len(row.Metadata) > 0 {
		require.NoError(t, json.Unmarshal(row.Metadata, &metadata))
	}
	revision, _ := metadata["draft_revision"].(string)
	if strings.TrimSpace(revision) == "" {
		revision = "rev-1"
	}
	metadata = map[string]interface{}{
		"draft_revision":               revision,
		"conflict_check_id":            metadata["conflict_check_id"],
		"conflict_check_facts_version": "FV-EDITED-SNAPSHOT",
		"conflict_invalidated_at":      "2026-09-26T12:00:00Z",
		"lawyer_facts_version":         "FV-EDITED-SNAPSHOT",
	}
	raw, err := json.Marshal(metadata)
	require.NoError(t, err)
	require.NoError(t, db.Exec("UPDATE case_intakes SET status = 'lawyer_facts_confirmed', metadata = ? WHERE id = ?", string(raw), intakeID).Error)
}

func recommitNewerCheck(t *testing.T, db *gorm.DB, intakeID string) {
	t.Helper()
	metadata := "{\"draft_revision\":\"rev-1\",\"conflict_check_id\":\"CCT_NEWER\"}"
	require.NoError(t, db.Exec("UPDATE case_intakes SET metadata = ? WHERE id = ?", metadata, intakeID).Error)
}

func validResubmitRequest(rec *models.ApprovalRecord, revision string) *ResubmitApprovalRequest {
	return &ResubmitApprovalRequest{
		ExpectedIntakeRevision: revision,
		ReturnRecordID:         rec.ID,
		RevisionNote:           "已按清单补充",
		Replies: []SupplementReply{
			{ItemID: "item-1", Response: "已修改标题并保存"},
		},
	}
}
func pinnedResubmitRequest(rec *models.ApprovalRecord, revision, snapshotID string) *ResubmitApprovalRequest {
	req := validResubmitRequest(rec, revision)
	req.ExpectedSubmissionSnapshotID = snapshotID
	return req
}

func TestStructuredResubmit_HappyPathPersistsSnapshotAndPointer(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-RC-1"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)
	require.NotEmpty(t, updated.ActiveSubmissionSnapshotID)

	var snap models.ApprovalSnapshot
	require.NoError(t, db.Where("id = ?", updated.ActiveSubmissionSnapshotID).Take(&snap).Error)
	require.Equal(t, models.ApprovalSnapshotTypeIntakeSubmission, snap.SnapshotType)
	require.Equal(t, 2, snap.SourceVersion)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.SnapshotData), &payload))
	require.Equal(t, intakeID, payload["intake_id"])
	require.Equal(t, "rev-1", payload["intake_revision"])
	require.Equal(t, rec.ID, payload["return_record_id"])
}

// Round diff acceptance: the new round snapshot carries a before/after
// summary of the mixed checklist (field corrections plus material returns),
// anchored on material ids, without rewriting the returned-round snapshot.
func TestStructuredResubmitPersistsRoundDiffForApproverAudit(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-RC-DIFF"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.NotEmpty(t, updated.ActiveSubmissionSnapshotID)

	var snap models.ApprovalSnapshot
	require.NoError(t, db.Where("id = ?", updated.ActiveSubmissionSnapshotID).Take(&snap).Error)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.SnapshotData), &payload))
	roundDiff, ok := payload["round_diff"].(map[string]interface{})
	require.True(t, ok, "structured resubmit must persist a round_diff for approver audit")
	require.Contains(t, roundDiff, "field_changes")
	require.Contains(t, roundDiff, "material_changes")

	// The returned round stays untouched: the previous snapshot payload has
	// no round_diff key after the new round lands.
	var prevSnap models.ApprovalSnapshot
	require.NoError(t, db.Where("id = ?", initialSnap).Take(&prevSnap).Error)
	var prevPayload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(prevSnap.SnapshotData), &prevPayload))
	require.NotContains(t, prevPayload, "round_diff")

	// Material changes are anchored on material ids with the new/carry
	// distinction, so a swapped row cannot masquerade as a re-upload.
	if changes, ok := roundDiff["material_changes"].([]interface{}); ok {
		for _, item := range changes {
			entry, _ := item.(map[string]interface{})
			require.Contains(t, entry, "material_id")
			require.Contains(t, entry, "new_files")
			require.Contains(t, entry, "carried_over")
		}
	}
}

func TestStructuredResubmit_StaleRevisionRejectedWithoutStateChange(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-RC-2"
	seedResubmitContractIntake(t, db, intakeID, "rev-2")
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)

	var fresh models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&fresh).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, fresh.Status)
	require.Empty(t, fresh.ActiveSubmissionSnapshotID)
	var snapCount int64
	require.NoError(t, db.Table("approval_snapshots").Where("approval_request_id = ?", approval.ID).Count(&snapCount).Error)
	require.Zero(t, snapCount)
}

func TestStructuredResubmit_RejectedWhenReturnRecordNotLatest(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-RC-3"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	newer := &models.ApprovalRecord{
		ID:                  rec.ID + "-newer",
		ApprovalRequestID:   approval.ID,
		Stage:               "stage-1",
		StageOrder:          1,
		ApproverID:          "u-approver",
		Decision:            models.ApprovalDecisionRequestChanges,
		DecisionReason:      "第二轮退回",
		ImposedRequirements: "{\"schema_version\":1,\"items\":[{\"item_id\":\"item-2\",\"category\":\"matter_explanation\",\"field_key\":\"case_summary\",\"description\":\"补充说明\",\"required\":true}]}",
		FollowUpActions:     "[]",
		ApprovalDate:        time.Now().Add(time.Minute),
		SupportingDocuments: "[]",
		EvidenceReferences:  "[]",
		Status:              "active",
	}
	require.NoError(t, db.Create(newer).Error)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)
}

func TestStructuredResubmit_UnknownReplyItemRejected(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-RC-4"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	svc := NewApprovalService(db)

	req := validResubmitRequest(rec, "rev-1")
	req.Replies = []SupplementReply{{ItemID: "ghost-item", Response: "回复不存在项"}}
	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, req)
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeSupplementItemInvalid, contractErr.Code)
}

func TestStructuredResubmit_EmptyRepliesRejected(t *testing.T) {
	db := seedResubmitContractWorld(t)
	intakeID := "INT-RC-5"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedResubmitContractApproval(t, db, intakeID)
	svc := NewApprovalService(db)

	req := validResubmitRequest(rec, "rev-1")
	req.Replies = nil
	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, req)
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeSupplementItemInvalid, contractErr.Code)
}

// C09 / plan 6.4: once the intake facts are re-checked, the approval's old
// binding no longer matches the committed pointer. Resubmit must be a typed
// 409 and leave no snapshot, no pointer, and no status change.
func TestStructuredResubmit_NewerCommittedCheckRejectedWithoutStateChange(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	svc := NewApprovalService(db)

	recommitNewerCheck(t, db, intakeID)
	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)

	var fresh models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&fresh).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, fresh.Status)
	require.Empty(t, fresh.ActiveSubmissionSnapshotID)
	var snapCount int64
	require.NoError(t, db.Table("approval_snapshots").Where("approval_request_id = ?", approval.ID).Count(&snapCount).Error)
	require.Zero(t, snapCount)
}

// A committed pointer that never existed as evidence is a broken binding:
// fail closed with the unresolved-snapshot code instead of reusing anything.
func TestStructuredResubmit_CommittedPointerWithoutRecordRejected(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	recommitNewerCheck(t, db, intakeID)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)
}

// Pure-reply reuse is only allowed when the server can prove the binding:
// same committed check, completed record, full coverage, independent review.
func TestStructuredResubmit_ProvenBindingAllowsReplyOnlyReuse(t *testing.T) {
	db, approval, rec, intakeID, checkID := seedConflictWorld(t)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)
	require.NotEmpty(t, updated.ActiveSubmissionSnapshotID)

	var snap models.ApprovalSnapshot
	require.NoError(t, db.Where("id = ?", *updated.ActiveSubmissionSnapshotID).Take(&snap).Error)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.SnapshotData), &payload))
	require.Equal(t, checkID, payload["conflict_check_id"])
	require.Equal(t, "independent_review_current", payload["conflict_review_state"])
}

// No server-verifiable review (or an unacceptable one) blocks the reply-only
// reuse path: replies alone never satisfy plan 6.4.
func TestStructuredResubmit_MissingReviewBlocksReuse(t *testing.T) {
	db, approval, rec, _, _ := seedConflictWorld(t)
	require.NoError(t, db.Exec("DELETE FROM conflict_reviews").Error)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalSnapshotUnresolved, contractErr.Code)

	var fresh models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&fresh).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, fresh.Status)
}

// No server-verifiable check binding at all blocks resubmit; the approval
// must go back through facts confirmation plus a new check and review.
func TestStructuredResubmit_MissingCheckBindingBlocksResubmit(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	clearConflictBinding(t, db, approval, intakeID)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalSnapshotUnresolved, contractErr.Code)

	var fresh models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&fresh).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, fresh.Status)
}

// Coverage that was never confirmed COMPLETE blocks the reuse path, matching
// validateConflictApprovalGate's evidence rule.
func TestStructuredResubmit_IncompleteCoverageBlocksResubmit(t *testing.T) {
	db, approval, rec, _, _ := seedConflictWorld(t)
	require.NoError(t, db.Exec("UPDATE conflict_check_records SET check_result = ?",
		`{"decision":{"coverageStatus":"COVERAGE_LIMITED"}}`).Error)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalSnapshotUnresolved, contractErr.Code)

	var snapCount int64
	require.NoError(t, db.Table("approval_snapshots").Where("approval_request_id = ?", approval.ID).Count(&snapCount).Error)
	require.Zero(t, snapCount)
}

// The intake metadata pointer is the authoritative anchor: when
// CommitIntakeConflictResult rotates it to a new check, a stale active
// association naming the old check must NOT block the legal resubmit. The
// new check is complete and independently reviewed, so resubmit passes,
// the snapshot records the new check, and the success sync supersedes the
// stale association.
func TestStructuredResubmit_StaleAssociationDoesNotBlockCurrentChain(t *testing.T) {
	db, approval, rec, intakeID, checkID := seedConflictWorld(t)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	newCheckID := "CCT_R03_RECHECK_ASSOC"
	seedIntakeConflictBinding(t, db, intakeID, newCheckID, map[string]string{"draft_revision": "rev-1"})
	require.NoError(t, db.Exec(
		"INSERT INTO approval_conflict_associations (id, approval_request_id, conflict_check_id, association_status, created_by, updated_at) "+
			"VALUES ('assoc-stale-old', ?, ?, 'active', 'u-lawyer', CURRENT_TIMESTAMP)", approval.ID, checkID).Error)
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)

	var snap models.ApprovalSnapshot
	require.NoError(t, db.Where("id = ?", *updated.ActiveSubmissionSnapshotID).Take(&snap).Error)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.SnapshotData), &payload))
	require.Equal(t, newCheckID, payload["conflict_check_id"])

	// Success-path sync: the old association is superseded and the only
	// active row now names the verified check.
	var superseded int64
	require.NoError(t, db.Table("approval_conflict_associations").
		Where("approval_request_id = ? AND association_status = 'superseded'", approval.ID).
		Count(&superseded).Error)
	require.EqualValues(t, 1, superseded)
	var activeAssoc struct {
		CheckID string
	}
	require.NoError(t, db.Raw("SELECT conflict_check_id AS check_id FROM approval_conflict_associations WHERE approval_request_id = ? AND association_status = 'active'",
		approval.ID).Scan(&activeAssoc).Error)
	require.Equal(t, newCheckID, activeAssoc.CheckID)
}

// Without any active association the resubmit still resolves the binding
// from the committed intake pointer, and the success path creates a fresh
// active association so the later approval gate reads the verified check.
func TestStructuredResubmit_CreatesAssociationWhenNoneExists(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	newCheckID := "CCT_R03_NO_ASSOC"
	seedIntakeConflictBinding(t, db, intakeID, newCheckID, map[string]string{"draft_revision": "rev-1"})
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)

	var activeAssoc struct {
		CheckID string
	}
	require.NoError(t, db.Raw("SELECT conflict_check_id AS check_id FROM approval_conflict_associations WHERE approval_request_id = ? AND association_status = 'active'",
		approval.ID).Take(&activeAssoc).Error)
	require.Equal(t, newCheckID, activeAssoc.CheckID)
}

// Final-gate authority after resubmit: validateConflictApprovalGate resolves
// the check from the active association first and approval metadata second.
// After a re-check resubmit (with the old association superseded by the
// success sync), the gate must read exactly the newly committed check and
// its independent review, never the approval's stale metadata anchor.
func TestStructuredResubmit_FinalGateReadsNewCheckAfterResubmit(t *testing.T) {
	db, approval, rec, intakeID, oldCheckID := seedConflictWorld(t)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	newCheckID := "CCT_R03_FINAL_GATE"
	seedIntakeConflictBinding(t, db, intakeID, newCheckID, map[string]string{"draft_revision": "rev-1"})
	// A stale active association left behind by a historical integration
	// path: the success sync must supersede it, not merely ignore it.
	require.NoError(t, db.Exec(
		"INSERT INTO approval_conflict_associations (id, approval_request_id, conflict_check_id, association_status, created_by, updated_at) "+
			"VALUES ('assoc-final-stale', ?, ?, 'active', 'u-lawyer', CURRENT_TIMESTAMP)", approval.ID, oldCheckID).Error)
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)

	// Mirror validateConflictApprovalGate's exact resolution order through
	// the same repository the production gate uses.
	ctx := context.Background()
	repo := repositories.NewIntegrationRepository(db)
	var refreshed models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&refreshed).Error)
	gateCheckID := conflictCheckIDFromMetadata(parseMetadata(refreshed.Metadata))
	association, err := repo.GetConflictAssociationByApprovalID(ctx, approval.ID)
	require.NoError(t, err)
	require.NotNil(t, association, "success sync must leave exactly one active association")
	gateCheckID = association.ConflictCheckID
	require.Equal(t, newCheckID, gateCheckID, "final gate must read the new check, not the stale metadata anchor")

	record, err := repo.GetConflictCheckRecord(ctx, gateCheckID)
	require.NoError(t, err)
	require.NotNil(t, record)
	require.Equal(t, "COMPLETED", record.CheckStatus)
	review, err := repo.GetLatestConflictReview(ctx, gateCheckID)
	require.NoError(t, err)
	require.NotNil(t, review)
	require.Equal(t, "no_conflict", review.Decision)
}

// R03 red case: after a lawyer edits the confirmed facts, the server rolls
// the intake back to draft, stamps the invalidation, and drops the facts
// version; confirmation rotates the facts version while the superseded
// check pointer and its recorded facts binding survive. The old check must
// no longer back a resubmit even though the pointer and the record agree.
func TestStructuredResubmit_EditedFactsBlockSupersededCheckReuse(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	reconfirmFactsAfterEdit(t, db, intakeID)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)

	var fresh models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&fresh).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, fresh.Status)
	var snapCount int64
	require.NoError(t, db.Table("approval_snapshots").Where("approval_request_id = ?", approval.ID).Count(&snapCount).Error)
	require.Zero(t, snapCount)
}

// The post-commit state produced by CommitIntakeConflictResult (status back
// to conflict_ready, invalidation cleared, intake facts version equal to the
// committed binding, record bound to this intake) is the only path back to
// a provable binding.
func TestStructuredResubmit_NewCheckAfterFactsEditRestoresBinding(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	reconfirmFactsAfterEdit(t, db, intakeID)
	newCheckID := "CCT_R03_RECHECK"
	seedIntakeConflictBinding(t, db, intakeID, newCheckID, map[string]string{"draft_revision": "rev-1"})
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)

	var snap models.ApprovalSnapshot
	require.NoError(t, db.Where("id = ?", *updated.ActiveSubmissionSnapshotID).Take(&snap).Error)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.SnapshotData), &payload))
	require.Equal(t, newCheckID, payload["conflict_check_id"])
}

// A resubmit cannot run while the facts edit is still unconfirmed (the
// intake is in draft with a pending invalidation stamp), no matter which
// anchors still agree.
func TestStructuredResubmit_DraftIntakeWithPendingInvalidationBlocks(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	reconfirmFactsAfterEdit(t, db, intakeID)
	require.NoError(t, db.Exec("UPDATE case_intakes SET status = 'draft' WHERE id = ?", intakeID).Error)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)

	var fresh models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&fresh).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, fresh.Status)
}

// The committed facts version must equal the intake's current lawyer facts
// version: a check committed against an older snapshot cannot back the
// current facts even when every other anchor still matches.
func TestStructuredResubmit_FactsVersionMismatchBlocksReuse(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	var row struct {
		Metadata []byte
	}
	require.NoError(t, db.Table("case_intakes").Select("metadata").Where("id = ?", intakeID).Take(&row).Error)
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal(row.Metadata, &metadata))
	metadata["lawyer_facts_version"] = "FV-ROTATED-AFTER-COMMIT"
	raw, _ := json.Marshal(metadata)
	require.NoError(t, db.Exec("UPDATE case_intakes SET metadata = ? WHERE id = ?", string(raw), intakeID).Error)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalRoundStale, contractErr.Code)
}

// A binding that resolves through the task-id fallback is still verifiable
// when the intake pointer, record, and review all agree.
func TestStructuredResubmit_TaskIDFallbackStillProvable(t *testing.T) {
	db, approval, rec, intakeID, _ := seedConflictWorld(t)
	var metadata map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(approval.Metadata), &metadata))
	checkID, _ := metadata["conflict_check_id"].(string)
	require.NotEmpty(t, checkID)
	delete(metadata, "conflict_check_id")
	metadata["conflict_task_id"] = checkID
	raw, _ := json.Marshal(metadata)
	require.NoError(t, db.Model(&models.ApprovalRequest{}).Where("id = ?", approval.ID).Update("metadata", string(raw)).Error)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)
	svc := NewApprovalService(db)

	updated, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusResubmitted, updated.Status)
}

// Unacceptable review conclusions block the reuse path with the
// unresolved-snapshot code and no state change.
func TestStructuredResubmit_ConfirmedConflictReviewBlocksResubmit(t *testing.T) {
	db, approval, rec, _, _ := seedConflictWorld(t)
	require.NoError(t, db.Exec("UPDATE conflict_reviews SET decision = 'confirmed_conflict'").Error)
	svc := NewApprovalService(db)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, validResubmitRequest(rec, "rev-1"))
	require.Error(t, err)
	var contractErr *ResubmitContractError
	require.ErrorAs(t, err, &contractErr)
	require.Equal(t, models.ErrCodeApprovalSnapshotUnresolved, contractErr.Code)

	var fresh models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approval.ID).Take(&fresh).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, fresh.Status)
}
