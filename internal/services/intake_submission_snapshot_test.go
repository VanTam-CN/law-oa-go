package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func snapshotJSON(t *testing.T, v interface{}) string {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return string(raw)
}
func snapshotSQLite(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "snapshot.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.ApprovalSnapshot{}))
	return db
}
func TestApprovalSubmissionResolverPointerBeatsCreatedTime(t *testing.T) {
	db := snapshotSQLite(t)
	approval := &models.ApprovalRequest{ID: "approval", Metadata: `{"intake_id":"intake","case_creation_config":{"title":"untrusted"}}`}
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("snapshot-%d", i)
		frozen := map[string]interface{}{"intake_id": "intake", "conflict_check_id": fmt.Sprintf("check-%d", i), "case_creation_config": map[string]interface{}{"title": fmt.Sprintf("round-%d", i)}}
		data := map[string]interface{}{"snapshot_id": id, "approval_request_id": approval.ID, "intake_id": "intake", "source_version": i, "conflict_check_id": frozen["conflict_check_id"], "metadata": frozen}
		require.NoError(t, db.Create(&models.ApprovalSnapshot{ID: id, ApprovalRequestID: approval.ID, SnapshotType: models.IntakeSubmissionSnapshotType, SnapshotData: snapshotJSON(t, data), SourceVersion: i, CreatedAt: time.Now().Add(time.Duration(3-i) * time.Hour)}).Error)
	}
	active := "snapshot-2"
	approval.ActiveSubmissionSnapshotID = &active
	snapshot, err := ResolveApprovalSubmission(db, approval)
	require.NoError(t, err)
	require.Equal(t, "snapshot-2", snapshot["snapshot_id"])
	require.NoError(t, ApplyApprovalSubmission(db, approval))
	require.Equal(t, "check-2", approval.ConflictCheckID)
	require.Contains(t, approval.Metadata, "round-2")
	require.NotContains(t, approval.Metadata, "untrusted")
	require.NoError(t, ApplyApprovalSubmission(db, approval))
	approval.ID = "other"
	_, err = ResolveApprovalSubmission(db, approval)
	require.Error(t, err)
}
func TestApprovalSubmissionResolverLegacyAgreement(t *testing.T) {
	db := snapshotSQLite(t)
	facts := map[string]interface{}{"intake_id": "intake", "conflict_check_id": "old-check", "case_creation_config": map[string]interface{}{"title": "old"}}
	snapshot := map[string]interface{}{"snapshot_id": "SNAP_legacy", "metadata": facts}
	approval := &models.ApprovalRequest{ID: "legacy", Metadata: snapshotJSON(t, map[string]interface{}{"intake_id": "intake", "conflict_check_id": "old-check", "case_creation_config": facts["case_creation_config"], "approval_snapshot": snapshot})}
	row := models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: approval.ID, SnapshotType: "conflict_approval", SnapshotData: snapshotJSON(t, snapshot), SourceVersion: 1}
	require.NoError(t, db.Create(&row).Error)
	original := approval.Metadata
	resolved, err := ResolveApprovalSubmission(db, approval)
	require.NoError(t, err)
	require.Equal(t, row.ID, resolved["snapshot_id"])
	require.NoError(t, ApplyApprovalSubmission(db, approval))
	require.NoError(t, ApplyApprovalSubmission(db, approval))
	require.Equal(t, original, approval.Metadata)
	require.NoError(t, db.Table("approval_snapshots").Where("id = ?", row.ID).Update("snapshot_data", `{"metadata":{"intake_id":"different"}}`).Error)
	_, err = ResolveApprovalSubmission(db, approval)
	require.Error(t, err)
	require.NoError(t, db.Table("approval_snapshots").Where("id = ?", row.ID).Update("snapshot_data", row.SnapshotData).Error)
	later := row
	later.ID = uuid.NewString()
	later.SnapshotData = `{"metadata":{"intake_id":"later"}}`
	later.CreatedAt = time.Now().Add(time.Hour)
	require.NoError(t, db.Create(&later).Error)
	_, err = ResolveApprovalSubmission(db, approval)
	require.Error(t, err)
}
func TestApprovalSubmissionResolverNonIntakeLegacy(t *testing.T) {
	approval := &models.ApprovalRequest{ID: "legacy", Metadata: `{"plain":"unchanged"}`}
	require.NoError(t, FreezeIntakeSubmissionInTx(nil, approval, "", nil))
	require.NoError(t, ApplyApprovalSubmission(nil, approval))
	require.Equal(t, `{"plain":"unchanged"}`, approval.Metadata)
	db := snapshotSQLite(t)
	row := models.ApprovalSnapshot{ID: uuid.NewString(), ApprovalRequestID: approval.ID, SnapshotType: "conflict_approval", SnapshotData: `{"conflict_record":{"check_id":"old"}}`, SourceVersion: 1}
	require.NoError(t, db.Create(&row).Error)
	snapshot, err := ResolveApprovalSubmission(db, approval)
	require.NoError(t, err)
	require.Equal(t, "old", submissionText(submissionMap(snapshot["conflict_record"])["check_id"]))
	other := row
	other.ID = uuid.NewString()
	other.SnapshotData = `{"conflict_record":{"check_id":"newer"}}`
	require.NoError(t, db.Create(&other).Error)
	_, err = ResolveApprovalSubmission(db, approval)
	require.Error(t, err)
}

// These tests use a distinct schema for every test. No production DSN fallback.
func snapshotPG(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("INTAKE_COLLAB_PG_DSN")
	if dsn == "" {
		t.Skip("INTAKE_COLLAB_PG_DSN not set; real PostgreSQL evidence required")
	}
	base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	schema := "snapshot_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, base.Exec(`CREATE SCHEMA "`+schema+`"`).Error)
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		require.NoError(t, base.Exec(`DROP SCHEMA "`+schema+`" CASCADE`).Error)
		sqlBase, _ := base.DB()
		_ = sqlBase.Close()
	})
	statements := []string{
		`CREATE TABLE approval_requests(id varchar(36) PRIMARY KEY, applicant_id text, applicant_name text, status text, type text, metadata jsonb, conflict_check_id text, conflict_risk_level text, conflict_check_time timestamptz, conflict_result jsonb, created_case_id text, case_created boolean DEFAULT false, deleted_at timestamptz, updated_at timestamptz DEFAULT now())`,
		`CREATE TABLE approval_records(id varchar(36) PRIMARY KEY, approval_request_id varchar(36), decision text, decision_reason text, imposed_requirements jsonb, created_at timestamptz DEFAULT now())`,
		`CREATE TABLE approval_snapshots(id uuid PRIMARY KEY, approval_request_id varchar(36) NOT NULL, snapshot_type varchar(80) NOT NULL, snapshot_data jsonb NOT NULL, source_version integer NOT NULL, created_at timestamptz DEFAULT now())`,
		`CREATE TABLE users(id bigint PRIMARY KEY, role text, status text, manager_id bigint, deleted_at timestamptz)`,
		`CREATE TABLE clients(id bigint PRIMARY KEY,name text,identity_number_digest text,deleted_at timestamptz)`,
		`CREATE TABLE cases(id bigserial PRIMARY KEY,client_id bigint,conflict_check_id text,ethical_wall_enabled boolean DEFAULT false,deleted_at timestamptz)`,
		`CREATE TABLE case_intakes(id uuid PRIMARY KEY,intake_code text,client_id bigint,title text,case_type text,status text,priority text,description text,metadata jsonb,created_by bigint,updated_at timestamptz DEFAULT now())`,
		`CREATE TABLE case_materials(id bigserial PRIMARY KEY,intake_id uuid,name text,status text,required boolean,metadata jsonb)`,
		`CREATE TABLE case_intake_parties(id bigserial PRIMARY KEY,intake_id uuid,entity_name text,party_role text,identity_number_digest text,identity_number_ciphertext text,metadata jsonb)`,
	}
	for _, s := range statements {
		require.NoError(t, db.Exec(s).Error)
	}
	require.NoError(t, db.AutoMigrate(&models.ConflictCheckRecord{}, &models.ConflictReview{}, &models.ConflictReviewerAssignment{}, &repositories.CaseCreationTracking{}))
	up, err := os.ReadFile("../../migrations/000082_intake_submission_snapshots.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(up)).Error)
	return db
}

type snapshotWorld struct {
	db                   *gorm.DB
	intakeID, approvalID string
	check                models.ConflictCheckRecord
	originalMetadata     string
}

func seedSnapshotWorld(t *testing.T, db *gorm.DB) *snapshotWorld {
	t.Helper()
	w := &snapshotWorld{db: db, intakeID: uuid.NewString(), approvalID: uuid.NewString()}
	require.NoError(t, db.Exec(`INSERT INTO users(id,role,status) VALUES(1,'lawyer','active'),(2,'conflict_officer','active'),(3,'lawyer','active'),(4,'assistant','active')`).Error)
	clientDigest := strings.Repeat("a", 64)
	partyDigest := strings.Repeat("b", 64)
	require.NoError(t, db.Exec(`INSERT INTO clients(id,name,identity_number_digest) VALUES(10,'虚构客户',?)`, clientDigest).Error)
	now := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	metadata := map[string]interface{}{"draft_revision": "rev-1", "lawyer_facts_version": "facts-1", "lawyer_facts_confirmed_at": now.Add(-time.Second).Format(time.RFC3339), "lawyer_facts_confirmed_by": "1", "conflict_check_id": "check-1", "lawyer_id": "1", "billing_method": "hourly", "business_area": "民商事"}
	require.NoError(t, db.Exec(`INSERT INTO case_intakes(id,intake_code,client_id,title,case_type,status,priority,description,metadata,created_by) VALUES(?,'INT-1',10,'服务器案件','民事诉讼','conflict_ready','medium','服务器事项',?,1)`, w.intakeID, snapshotJSON(t, metadata)).Error)
	require.NoError(t, db.Exec(`INSERT INTO case_intake_parties(intake_id,entity_name,party_role,identity_number_digest,identity_number_ciphertext,metadata) VALUES(?,'虚构对方','opposing_party',?,'protected-not-copied','{"local_id":"party-1"}')`, w.intakeID, partyDigest).Error)
	response := models.ConflictCheckResponse{CheckID: "check-1", CheckTime: now, Decision: &models.ConflictDecisionSummary{CoverageStatus: "COMPLETE"}}
	w.check = models.ConflictCheckRecord{CheckID: "check-1", ClientID: "10", ClientName: "虚构客户", CaseName: "服务器案件", CaseType: "民事诉讼", CheckStatus: "COMPLETED", UserID: 1, CheckTime: now, RiskLevel: "LOW", SearchParameters: models.JSON{"intakeId": w.intakeID, "clientType": "PERSON", "clientIdentifiers": map[string]interface{}{"id_card": "hmac-sha256:" + clientDigest}, "parties": []models.ConflictPartyInfo{{Name: "虚构对方", Role: "OPPOSING_PARTY", Identifiers: map[string]string{"id_card": "hmac-sha256:" + partyDigest}}}}, CheckResult: models.JSON(submissionMap(response))}
	require.NoError(t, db.Create(&w.check).Error)
	review := models.ConflictReview{CheckID: "check-1", Decision: "no_conflict", ReviewerID: 2, ReviewerName: "独立核查", Notes: "虚构资料核验", EvidenceHash: conflictEvidenceHash(&response), CreatedAt: now.Add(time.Second)}
	require.NoError(t, db.Create(&review).Error)
	assignment := models.ConflictReviewerAssignment{ID: uuid.NewString(), CheckID: "check-1", ReviewerID: 2, AssignedBy: 3, Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, IndependenceReason: "独立", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&assignment).Error)
	w.originalMetadata = snapshotJSON(t, map[string]interface{}{"intake_id": w.intakeID, "case_creation_config": map[string]interface{}{"title": "伪造旧审批标题", "lawyer_id": 999}, "forged_fact": "not allowed"})
	require.NoError(t, db.Exec(`INSERT INTO approval_requests(id,applicant_id,applicant_name,status,type,metadata) VALUES(?,'1','申请律师','draft','case_creation',?)`, w.approvalID, w.originalMetadata).Error)
	require.NoError(t, db.Table("approval_requests").Select("metadata").Where("id = ?", w.approvalID).Scan(&w.originalMetadata).Error)
	return w
}
func (w *snapshotWorld) freeze(t *testing.T, req *models.ResubmitApprovalRequest, after func(*gorm.DB, *models.ApprovalRequest) error) error {
	t.Helper()
	return w.db.Transaction(func(tx *gorm.DB) error {
		var approval models.ApprovalRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&approval, "id = ?", w.approvalID).Error; err != nil {
			return err
		}
		if err := FreezeIntakeSubmissionInTx(tx, &approval, "1", req); err != nil {
			return err
		}
		updates := map[string]interface{}{"active_submission_snapshot_id": approval.ActiveSubmissionSnapshotID, "conflict_check_id": approval.ConflictCheckID, "conflict_risk_level": approval.ConflictRiskLevel, "conflict_check_time": approval.ConflictCheckTime, "conflict_result": approval.ConflictResult, "status": "submitted"}
		if req != nil {
			updates["status"] = "resubmitted"
		}
		if err := tx.Model(&models.ApprovalRequest{}).Where("id = ?", approval.ID).Updates(updates).Error; err != nil {
			return err
		}
		if after != nil {
			return after(tx, &approval)
		}
		return nil
	})
}
func (w *snapshotWorld) load(t *testing.T) *models.ApprovalRequest {
	t.Helper()
	var a models.ApprovalRequest
	require.NoError(t, w.db.First(&a, "id = ?", w.approvalID).Error)
	return &a
}
func (w *snapshotWorld) returned(t *testing.T) *models.ResubmitApprovalRequest {
	t.Helper()
	a := w.load(t)
	require.NotNil(t, a.ActiveSubmissionSnapshotID)
	raw, err := NormalizeSupplementRequirements(supplementTestInput(), "请补充两项")
	require.NoError(t, err)
	id := uuid.NewString()
	record := &models.ApprovalRecord{ID: id, ImposedRequirements: raw}
	items, err := ReadSupplementRequests(record)
	require.NoError(t, err)
	require.NoError(t, w.db.Exec(`INSERT INTO approval_records(id,approval_request_id,decision,decision_reason,imposed_requirements,submission_snapshot_id,action_version) VALUES(?,?,'request_changes','请补充两项',?,?,'node-1')`, id, w.approvalID, raw, *a.ActiveSubmissionSnapshotID).Error)
	require.NoError(t, w.db.Model(&models.ApprovalRequest{}).Where("id = ?", w.approvalID).Update("status", "needs_revision").Error)
	return &models.ResubmitApprovalRequest{ExpectedIntakeRevision: "rev-1", ExpectedSubmissionSnapshotID: *a.ActiveSubmissionSnapshotID, ReturnRecordID: id, RevisionNote: "仅补充说明", Responses: []models.SupplementResponse{{ItemID: items[0].ItemID, Response: "已核对"}, {ItemID: items[1].ItemID, Response: "已说明"}}}
}
func TestIntakeSubmissionPGFreezeResponsesReplayAndImmutability(t *testing.T) {
	db := snapshotPG(t)
	w := seedSnapshotWorld(t, db)
	require.NoError(t, w.freeze(t, nil, nil))
	a := w.load(t)
	require.Equal(t, w.originalMetadata, a.Metadata)
	snapshot, err := ResolveApprovalSubmission(db, a)
	require.NoError(t, err)
	raw := snapshotJSON(t, snapshot)
	require.Contains(t, raw, "服务器案件")
	require.NotContains(t, raw, "伪造旧审批")
	require.NotContains(t, raw, "forged_fact")
	require.NotContains(t, raw, "protected-not-copied")
	var originalRow models.ApprovalSnapshot
	require.NoError(t, db.First(&originalRow, "id = ?", *a.ActiveSubmissionSnapshotID).Error)
	req := w.returned(t)
	var revision string
	err = db.Transaction(func(tx *gorm.DB) error {
		var err error
		revision, err = SaveIntakeSupplementResponsesInTx(tx, w.intakeID, w.approvalID, "1", "rev-1", req.ReturnRecordID, req.Responses[:1])
		return err
	})
	require.NoError(t, err)
	var intake submissionIntakeRow
	require.NoError(t, db.Table("case_intakes").First(&intake, "id = ?", w.intakeID).Error)
	metadata := submissionMap(intake.Metadata)
	require.Equal(t, "facts-1", metadata["lawyer_facts_version"])
	require.Equal(t, "conflict_ready", intake.Status)
	require.NotEqual(t, "rev-1", revision)
	require.Error(t, w.freeze(t, req, nil)) // Stale draft revision cannot overwrite replies.
	req.ExpectedIntakeRevision = revision
	require.NoError(t, w.freeze(t, req, nil))
	a = w.load(t)
	require.NotEqual(t, originalRow.ID, *a.ActiveSubmissionSnapshotID)
	replay, err := ValidateResubmissionReplayInTx(db, a, req)
	require.NoError(t, err)
	require.True(t, replay)
	changed := *req
	changed.RevisionNote = "不同内容"
	_, err = ValidateResubmissionReplayInTx(db, a, &changed)
	require.Error(t, err)
	var count int64
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	var unchanged models.ApprovalSnapshot
	require.NoError(t, db.First(&unchanged, "id = ?", originalRow.ID).Error)
	require.Equal(t, originalRow.SnapshotData, unchanged.SnapshotData)
	require.Error(t, db.Exec(`UPDATE approval_snapshots SET snapshot_data='{}' WHERE id=?`, originalRow.ID).Error)
	require.Error(t, db.Exec(`DELETE FROM approval_snapshots WHERE id=?`, originalRow.ID).Error)
	require.Error(t, db.Exec(`UPDATE approval_records SET decision_reason='rewrite' WHERE id=?`, req.ReturnRecordID).Error)
	require.Error(t, db.Exec(`DELETE FROM approval_records WHERE id=?`, req.ReturnRecordID).Error)
	// Constraints also apply to raw SQL inserts that bypass application locks/hooks.
	require.Error(t, db.Exec(`INSERT INTO approval_snapshots(id,approval_request_id,snapshot_type,snapshot_data,source_version) SELECT ?,approval_request_id,snapshot_type,snapshot_data,source_version FROM approval_snapshots WHERE id=?`, uuid.NewString(), originalRow.ID).Error)
	require.Error(t, db.Exec(`INSERT INTO approval_snapshots(id,approval_request_id,snapshot_type,snapshot_data,source_version) SELECT ?,approval_request_id,snapshot_type,snapshot_data,99 FROM approval_snapshots WHERE id=?`, uuid.NewString(), *a.ActiveSubmissionSnapshotID).Error)

}

func TestIntakeSubmissionPGNegativeGates(t *testing.T) {
	for _, tc := range []struct {
		name, sql string
		args      []interface{}
	}{
		{"changed party identity", `UPDATE case_intake_parties SET identity_number_digest='different'`, nil},
		{"changed party role", `UPDATE case_intake_parties SET party_role='witness'`, nil},
		{"changed client", `UPDATE clients SET identity_number_digest='different'`, nil},
		{"missing current check", `UPDATE case_intakes SET metadata=metadata-'conflict_check_id'`, nil},
		{"unconfirmed facts", `UPDATE case_intakes SET metadata=metadata-'lawyer_facts_version'`, nil},
		{"review mismatch", `UPDATE conflict_reviews SET evidence_hash='wrong'`, nil},
		{"expired review", `UPDATE conflict_reviews SET next_review_at=now()-interval '1 day'`, nil},
		{"self review", `UPDATE conflict_reviews SET reviewer_id=1`, nil},
		{"manager review", `UPDATE users SET manager_id=2 WHERE id=1`, nil},
		{"assignment replaced after review", `UPDATE conflict_reviewer_assignments SET effective_from=now()-interval '10 seconds'`, nil},
		{"assignment bound to other case", `UPDATE conflict_reviewer_assignments SET case_id=77`, nil},
		{"assignment revoked", `UPDATE conflict_reviewer_assignments SET status='revoked'`, nil},
		{"missing dependency", `DROP TABLE conflict_reviews`, nil},
		{"insufficient decision", `UPDATE conflict_reviews SET decision='insufficient_information'`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := snapshotPG(t)
			w := seedSnapshotWorld(t, db)
			require.NoError(t, db.Exec(tc.sql, tc.args...).Error)
			require.Error(t, w.freeze(t, nil, nil))
			var count int64
			require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
			require.Zero(t, count)
			require.Nil(t, w.load(t).ActiveSubmissionSnapshotID)
		})
	}
}
func TestIntakeSubmissionPGTransactionRollbackAndConcurrentReplay(t *testing.T) {
	db := snapshotPG(t)
	w := seedSnapshotWorld(t, db)
	require.Error(t, w.freeze(t, nil, func(tx *gorm.DB, a *models.ApprovalRequest) error {
		return errors.New("injected downstream inbox failure")
	}))
	require.Nil(t, w.load(t).ActiveSubmissionSnapshotID)
	var count int64
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, w.freeze(t, nil, nil))
	req := w.returned(t)
	// Three concurrent rounds use the same caller lock order and idempotency gate.
	for round := 0; round < 3; round++ {
		var wg sync.WaitGroup
		start := make(chan struct{})
		errs := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				errs <- db.Transaction(func(tx *gorm.DB) error {
					var a models.ApprovalRequest
					if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&a, "id = ?", w.approvalID).Error; err != nil {
						return err
					}
					replay, err := ValidateResubmissionReplayInTx(tx, &a, req)
					if err != nil || replay {
						return err
					}
					if err := FreezeIntakeSubmissionInTx(tx, &a, "1", req); err != nil {
						return err
					}
					return tx.Model(&models.ApprovalRequest{}).Where("id = ?", a.ID).Updates(map[string]interface{}{"active_submission_snapshot_id": a.ActiveSubmissionSnapshotID, "status": "resubmitted"}).Error
				})
			}()
		}
		close(start)
		wg.Wait()
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}
		require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
		require.EqualValues(t, round+2, count)
		if round < 2 {
			req = w.returned(t)
		}
	}
}
func TestIntakeSubmissionPGMigrationUpDownAndRefusal(t *testing.T) {
	db := snapshotPG(t)
	down, err := os.ReadFile("../../migrations/000082_intake_submission_snapshots.down.sql")
	require.NoError(t, err)
	up, err := os.ReadFile("../../migrations/000082_intake_submission_snapshots.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(`INSERT INTO approval_snapshots(id,approval_request_id,snapshot_type,snapshot_data,source_version) VALUES(?,'old','legacy','{}',1),(?,'old','legacy','{}',1)`, uuid.NewString(), uuid.NewString()).Error)
	require.NoError(t, db.Exec(string(down)).Error)
	require.NoError(t, db.Exec(string(up)).Error)
	require.NoError(t, db.Exec(string(up)).Error)
	w := seedSnapshotWorld(t, db)
	require.NoError(t, w.freeze(t, nil, nil))
	require.Error(t, db.Exec(string(down)).Error)
	require.True(t, db.Migrator().HasColumn(&models.ApprovalRequest{}, "active_submission_snapshot_id"))
	var count int64
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
	require.EqualValues(t, 3, count)
}

func TestIntakeSubmissionPGControlledResponsePermissionsAndRollback(t *testing.T) {
	db := snapshotPG(t)
	w := seedSnapshotWorld(t, db)
	require.NoError(t, w.freeze(t, nil, nil))
	req := w.returned(t)
	var original string
	require.NoError(t, db.Table("case_intakes").Select("metadata").Where("id = ?", w.intakeID).Scan(&original).Error)
	for _, tc := range []struct {
		name, user, revision, returnID string
		responses                      []models.SupplementResponse
	}{
		{"other lawyer", "3", "rev-1", req.ReturnRecordID, req.Responses},
		{"assistant", "4", "rev-1", req.ReturnRecordID, req.Responses},
		{"stale revision", "1", "old", req.ReturnRecordID, req.Responses},
		{"wrong return", "1", "rev-1", uuid.NewString(), req.Responses},
		{"forged item", "1", "rev-1", req.ReturnRecordID, []models.SupplementResponse{{ItemID: "forged", Response: "完成"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := db.Transaction(func(tx *gorm.DB) error {
				_, err := SaveIntakeSupplementResponsesInTx(tx, w.intakeID, w.approvalID, tc.user, tc.revision, tc.returnID, tc.responses)
				return err
			})
			require.Error(t, err)
		})
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		_, err := SaveIntakeSupplementResponsesInTx(tx, w.intakeID, w.approvalID, "1", "rev-1", req.ReturnRecordID, req.Responses)
		if err != nil {
			return err
		}
		return errors.New("injected failure after response write")
	})
	require.Error(t, err)
	var unchanged string
	require.NoError(t, db.Table("case_intakes").Select("metadata").Where("id = ?", w.intakeID).Scan(&unchanged).Error)
	require.Equal(t, original, unchanged)
	missing := *req
	missing.Responses = missing.Responses[:1]
	require.Error(t, w.freeze(t, &missing, nil))
	wrong := *req
	wrong.ExpectedSubmissionSnapshotID = uuid.NewString()
	require.Error(t, w.freeze(t, &wrong, nil))
	var count int64
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}
func TestIntakeSubmissionPGSnapshotWriteFailureRollsBack(t *testing.T) {
	db := snapshotPG(t)
	w := seedSnapshotWorld(t, db)
	require.NoError(t, db.Exec(`CREATE FUNCTION fail_submission_insert() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected snapshot failure'; END $$; CREATE TRIGGER snapshot_write_fault BEFORE INSERT ON approval_snapshots FOR EACH ROW EXECUTE FUNCTION fail_submission_insert();`).Error)
	require.Error(t, w.freeze(t, nil, nil))
	require.Nil(t, w.load(t).ActiveSubmissionSnapshotID)
	var count int64
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, db.Exec(`DROP TRIGGER snapshot_write_fault ON approval_snapshots`).Error)
	require.NoError(t, w.freeze(t, nil, nil))
	require.NoError(t, db.Model(&models.ApprovalSnapshot{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestApprovalSubmissionPreservesBusinessRole(t *testing.T) {
	digest := strings.Repeat("c", 64)
	check := &models.ConflictCheckRecord{SearchParameters: models.JSON{"parties": []models.ConflictPartyInfo{{Name: "虚构担保方", Role: "GUARANTOR", Identifiers: map[string]string{"id_card": "hmac-sha256:" + digest}}}}}
	rows := []map[string]interface{}{{"entity_name": "虚构担保方", "party_role": "related_party", "identity_number_digest": digest, "metadata": map[string]interface{}{"party_type": "GUARANTOR"}}}
	require.NoError(t, validateSubmissionPartyRows(rows, check))
	rows[0]["metadata"] = map[string]interface{}{"party_type": "RELATED_PARTY"}
	require.Error(t, validateSubmissionPartyRows(rows, check), "archive role aliases cannot hide a changed professional role")
}
