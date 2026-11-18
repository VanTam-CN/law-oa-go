package services

// C4 atomic initial submission (SQLite level): intake-scoped integrated
// submissions go through the real ApprovalService.CreateAndSubmitApproval
// transaction, so an approver-assignment failure must roll back the draft,
// the intake slot binding, and every conflict binding row together. The
// legacy compensating-delete contract is gone by design: there is nothing
// left to compensate when the whole submission is one transaction.

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func newIntakeAtomicSubmitDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/intake_atomic_submit.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE approval_snapshots (id TEXT PRIMARY KEY, approval_request_id TEXT, snapshot_type TEXT, snapshot_data TEXT, source_version INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)").Error)
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, client_id TEXT, title TEXT, case_type TEXT, priority TEXT, description TEXT, created_by TEXT, status TEXT, metadata TEXT, updated_at DATETIME)").Error)
	require.NoError(t, db.Exec("CREATE TABLE case_intake_parties (id INTEGER PRIMARY KEY AUTOINCREMENT, intake_id TEXT, entity_name TEXT, entity_type TEXT, party_role TEXT, relation_depth INTEGER DEFAULT 0, identity_type TEXT, identity_number_digest TEXT, aliases TEXT, metadata TEXT DEFAULT '{}')").Error)
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
	return db
}

type intakeAtomicChainSeed struct {
	IntakeID string
	CheckID  string
}

func seedIntakeAtomicChain(t *testing.T, db *gorm.DB, status, factsVersion, committedFactsVersion, checkStatus, coverage, reviewDecision string) intakeAtomicChainSeed {
	t.Helper()
	now := time.Now().UnixNano()
	seed := intakeAtomicChainSeed{
		IntakeID: fmt.Sprintf("INT-C4-%d", now),
		CheckID:  fmt.Sprintf("CCT-C4-%d", now),
	}
	intakeMeta := fmt.Sprintf(
		`{"draft_revision":"rev-c4","conflict_check_id":"%s","lawyer_facts_version":"%s","conflict_check_facts_version":"%s"}`,
		seed.CheckID, factsVersion, committedFactsVersion)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, client_id, title, case_type, priority, description, created_by, status, metadata, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		seed.IntakeID, "9", "C4 原子提交接案", "CIVIL", "high", "测试事实描述", "100", status, intakeMeta, time.Now()).Error)
	require.NoError(t, db.Exec(
		"INSERT INTO case_intake_parties (intake_id, entity_name, entity_type, party_role, relation_depth, identity_type, identity_number_digest, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, '{}')",
		seed.IntakeID, "对手方测试有限公司", "company", "opposing_party", 1, "uscc", "hmac-sha256:c4-opposing-digest").Error)
	require.NoError(t, db.Create(&models.ConflictCheckRecord{
		CheckID:          seed.CheckID,
		ClientID:         "9",
		ClientName:       "c4-customer",
		CheckStatus:      checkStatus,
		SearchParameters: models.JSON{"intakeId": seed.IntakeID, "clientId": "9"},
		CheckResult:      models.JSON{"decision": map[string]interface{}{"coverageStatus": coverage}},
		UserID:           1,
		CheckTime:        time.Now(),
	}).Error)
	if reviewDecision != "" {
		require.NoError(t, db.Create(&models.ConflictReview{
			CheckID:      seed.CheckID,
			Decision:     reviewDecision,
			Notes:        "ok",
			ReviewerID:   1,
			ReviewerName: "独立复核人",
			EvidenceHash: "sqlite-c4-hash",
		}).Error)
	}
	return seed
}

func intakeAtomicCounts(t *testing.T, db *gorm.DB, intakeID string) (int64, int64, int64) {
	t.Helper()
	var binds, assoc, snaps int64
	require.NoError(t, db.Table("approval_case_creation_tracking").Where("intake_id = ?", intakeID).Count(&binds).Error)
	require.NoError(t, db.Table("approval_conflict_associations").Count(&assoc).Error)
	require.NoError(t, db.Table("approval_snapshots").Where("snapshot_type = ?", models.ApprovalSnapshotTypeIntakeSubmission).Count(&snaps).Error)
	return binds, assoc, snaps
}

func TestIntakeAtomicSubmissionZeroResidueOnAssignFailure(t *testing.T) {
	db := newIntakeAtomicSubmitDB(t)
	// No admin user is seeded, so approver assignment must fail inside the
	// transaction after the draft insert and the slot bind already ran.
	seed := seedIntakeAtomicChain(t, db, "approval_submitting", "FV-1", "FV-1", "COMPLETED", "COMPLETE", "no_conflict")

	_, err := NewApprovalService(db).CreateAndSubmitApproval("u-lawyer", "律师甲", &models.CreateApprovalRequest{
		Type:         "CASE_CREATION",
		Title:        "C4 原子提交",
		Content:      "atomic submission",
		WorkflowType: "CASE_APPROVAL",
		Metadata:     map[string]interface{}{"intake_id": seed.IntakeID},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "分配审批人失败")

	binds, assoc, snaps := intakeAtomicCounts(t, db, seed.IntakeID)
	require.Zero(t, binds, "draft rollback must remove the intake slot binding")
	require.Zero(t, assoc, "no conflict association may survive a failed submission")
	require.Zero(t, snaps, "no intake_submission snapshot may survive a failed submission")

	var drafts int64
	require.NoError(t, db.Model(&models.ApprovalRequest{}).Count(&drafts).Error)
	require.Zero(t, drafts, "the approval draft itself must roll back")
	var intakeStatus string
	require.NoError(t, db.Raw("SELECT status FROM case_intakes WHERE id = ?", seed.IntakeID).Scan(&intakeStatus).Error)
	require.Equal(t, "approval_submitting", intakeStatus,
		"the handler-level claim is released by the handler, not by the approval service")
}

func TestIntakeAtomicSubmissionHappyPathCommitsAllBindings(t *testing.T) {
	db := newIntakeAtomicSubmitDB(t)
	seed := seedIntakeAtomicChain(t, db, "approval_submitting", "FV-OK", "FV-OK", "COMPLETED", "COMPLETE", "no_conflict")
	require.NoError(t, db.Create(&models.User{
		Username: "admin-c4", Name: "审批管理员", Email: "admin-c4@example.com",
		Password: "x", Role: "admin", Status: "active",
	}).Error)

	created, err := NewApprovalService(db).CreateAndSubmitApproval("u-lawyer", "律师甲", &models.CreateApprovalRequest{
		Type:         "CASE_CREATION",
		Title:        "C4 原子提交正例",
		Content:      "atomic submission happy path",
		WorkflowType: "CASE_APPROVAL",
		Metadata:     map[string]interface{}{"intake_id": seed.IntakeID},
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, models.ApprovalStatusSubmitted, created.Status)
	require.Equal(t, seed.CheckID, created.ConflictCheckID)
	require.NotNil(t, created.ActiveSubmissionSnapshotID)

	binds, assoc, snaps := intakeAtomicCounts(t, db, seed.IntakeID)
	require.Equal(t, int64(1), binds, "the intake slot binding commits with the approval")
	require.Equal(t, int64(1), assoc, "the active conflict association commits with the approval")
	require.Equal(t, int64(1), snaps, "exactly one first-round snapshot commits")

	var snap struct {
		SnapshotData string
	}
	require.NoError(t, db.Table("approval_snapshots").
		Where("approval_request_id = ? AND snapshot_type = ?", created.ID, models.ApprovalSnapshotTypeIntakeSubmission).
		Take(&snap).Error)
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(snap.SnapshotData), &payload))
	structured, ok := payload["structured_facts"].(map[string]interface{})
	require.True(t, ok, "snapshot must carry structured_facts")
	require.Equal(t, "C4 原子提交接案", structured["title"])
	require.Equal(t, "CIVIL", structured["case_type"])
	require.Equal(t, "high", structured["priority"])
	require.Equal(t, "测试事实描述", structured["description"])
	parties, ok := structured["parties"].([]interface{})
	require.True(t, ok, "structured_facts must carry parties")
	require.Len(t, parties, 1, "the seeded intake party must be frozen into the snapshot")
	first, ok2 := parties[0].(map[string]interface{})
	require.True(t, ok2)
	require.Equal(t, "对手方测试有限公司", first["entity_name"])
	require.Equal(t, "opposing_party", first["party_role"])
	require.Equal(t, "uscc", first["identity_type"])
	require.Equal(t, "hmac-sha256:c4-opposing-digest", first["identity_digest"],
		"the snapshot must carry the protected identity digest, never ciphertext")

	var activeAssoc int64
	require.NoError(t, db.Table("approval_conflict_associations").
		Where("approval_request_id = ? AND association_status = ?", created.ID, "active").Count(&activeAssoc).Error)
	require.Equal(t, int64(1), activeAssoc)
	var assocCheckID string
	require.NoError(t, db.Raw("SELECT conflict_check_id FROM approval_conflict_associations WHERE approval_request_id = ?", created.ID).Scan(&assocCheckID).Error)
	require.Equal(t, seed.CheckID, assocCheckID)
}
