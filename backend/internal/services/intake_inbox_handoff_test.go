package services

// N4 (C05-C07, C13) acceptance: workflow handoffs open/close with the
// business transaction, cannot be completed manually, and survive retries
// without duplication.

import (
	"context"
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

func seedHandoffWorld(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s", t.TempDir()+"/handoff.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE approval_snapshots (id TEXT PRIMARY KEY, approval_request_id TEXT, snapshot_type TEXT, snapshot_data TEXT, source_version INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)").Error)
	require.NoError(t, db.AutoMigrate(
		&models.ApprovalRequest{},
		&models.ApprovalRecord{},
		&models.ApprovalWorkflow{},
		&models.User{},
		&models.InboxItem{},
		&repositories.CaseCreationTracking{},
	))
	require.NoError(t, db.Create(&models.User{Username: "h-lawyer", Name: "律师甲", Email: "h-lawyer@example.com", Password: "x", Role: "lawyer", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.User{Username: "h-admin", Name: "审批管理员", Email: "h-admin@example.com", Password: "x", Role: "admin", Status: "active"}).Error)
	// Approval actors use string IDs; the handoff resolver needs matching
	// active users for u-h-lawyer / u-h-admin.
	require.NoError(t, db.Exec("UPDATE users SET id = 101 WHERE username = 'h-lawyer'").Error)
	require.NoError(t, db.Exec("UPDATE users SET id = 102 WHERE username = 'h-admin'").Error)
	require.NoError(t, db.Exec("INSERT INTO users (username, name, email, password, role, status) VALUES ('u-h-lawyer', '律师甲', 'u-h-lawyer@example.com', 'x', 'lawyer', 'active')").Error)
	require.NoError(t, db.Exec("INSERT INTO users (username, name, email, password, role, status) VALUES ('u-h-admin', '审批管理员', 'u-h-admin@example.com', 'x', 'admin', 'active')").Error)
	require.NoError(t, db.Exec("CREATE TABLE IF NOT EXISTS case_intakes (id TEXT PRIMARY KEY, client_id TEXT, status TEXT, metadata TEXT, updated_at DATETIME)").Error)
	return db
}

func seedHandoffApproval(t *testing.T, db *gorm.DB) (*models.ApprovalRequest, *models.ApprovalRecord, string) {
	t.Helper()
	now := time.Now().UnixNano()
	intakeID := fmt.Sprintf("INT-H-%d", now)
	approval := &models.ApprovalRequest{
		ID:                fmt.Sprintf("APR-H-%d", now),
		RequestNumber:     fmt.Sprintf("RN-H-%d", now),
		Title:             "handoff world",
		Type:              "CASE_CREATION",
		Content:           "n4",
		ApplicantID:       "u-h-lawyer",
		ApplicantName:     "律师甲",
		Status:            models.ApprovalStatusNeedsRevision,
		WorkflowType:      "CASE_APPROVAL",
		CurrentApproverID: "u-h-admin",
		Metadata:          fmt.Sprintf("{\"intake_id\":\"%s\",\"conflict_check_id\":\"CHK-H-%d\"}", intakeID, now),
	}
	require.NoError(t, db.Create(approval).Error)
	rec := &models.ApprovalRecord{
		ID:                  fmt.Sprintf("REC-H-%d", now),
		ApprovalRequestID:   approval.ID,
		Stage:               "stage-1",
		StageOrder:          1,
		ApproverID:          "u-h-admin",
		Decision:            models.ApprovalDecisionRequestChanges,
		DecisionReason:      "请修正标题",
		ImposedRequirements: "{\"schema_version\":1,\"items\":[{\"item_id\":\"item-1\",\"category\":\"field_correction\",\"field_key\":\"case_title\",\"description\":\"标题\",\"required\":true}]}",
		FollowUpActions:     "[]",
		ApprovalDate:        time.Now(),
		SupportingDocuments: "[]",
		EvidenceReferences:  "[]",
		Status:              "active",
	}
	require.NoError(t, db.Create(rec).Error)
	return approval, rec, intakeID
}

func openTasks(t *testing.T, db *gorm.DB, userID string) []models.InboxItem {
	t.Helper()
	var items []models.InboxItem
	require.NoError(t, db.Raw(
		"SELECT i.* FROM inbox_items i JOIN users u ON u.id = i.user_id "+
			"WHERE u.username = ? AND i.is_completed = ? ORDER BY i.id", userID, false).Scan(&items).Error)
	return items
}

func taskKey(t *testing.T, db *gorm.DB, key string) *models.InboxItem {
	t.Helper()
	var item models.InboxItem
	err := db.Where("workflow_task_key = ?", key).Take(&item).Error
	if err != nil {
		return nil
	}
	return &item
}

// C05: the approval decision handoff opens for the approver when the
// submission commits, carries the approval UUID in source_ref, and is
// workflow-owned.
func TestSubmitApproval_OpensApproverDecisionHandoff(t *testing.T) {
	db := seedHandoffWorld(t)
	svc := NewApprovalService(db)
	now := time.Now().UnixNano()
	approval := &models.ApprovalRequest{
		ID:            fmt.Sprintf("APR-HS-%d", now),
		RequestNumber: fmt.Sprintf("RN-HS-%d", now),
		Title:         "handoff submit",
		Type:          "CASE_CREATION",
		Content:       "n4",
		ApplicantID:   "u-h-lawyer",
		ApplicantName: "律师甲",
		Status:        models.ApprovalStatusDraft,
		WorkflowType:  "CASE_APPROVAL",
		Metadata:      "{}",
	}
	require.NoError(t, db.Create(approval).Error)

	updated, err := svc.SubmitApproval("u-h-lawyer", approval.ID)
	require.NoError(t, err)
	require.Equal(t, models.ApprovalStatusSubmitted, updated.Status)
	require.NotEmpty(t, updated.CurrentApproverID)

	items := openTasks(t, db, "u-h-admin")
	require.Len(t, items, 1)
	require.Equal(t, "workflow", items[0].CompletionMode)
	require.Equal(t, approval.ID, items[0].SourceRef)
}

// C05+C07: a request_changes decision closes the approver handoff and
// opens the applicant revision handoff in the same transaction; the
// workflow revision todo cannot be completed manually by its owner.
func TestRequestChanges_HandoffClosesAndOpensRevisionTask(t *testing.T) {
	db := seedHandoffWorld(t)
	svc := NewApprovalService(db)
	approval, _, intakeID := seedHandoffApproval(t, db)
	// Put the approval into submitted state with the admin as approver and
	// create the c0 decision handoff directly (as SubmitApproval would).
	approval.Status = models.ApprovalStatusSubmitted
	require.NoError(t, db.Save(approval).Error)
	// Intake-bound decisions must pin the active submission snapshot
	// (N3/C11 stale-page guard), so mint the round-1 snapshot that the
	// production initial submission would have committed.
	seedActiveSubmissionRound(t, db, approval, intakeID)
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return svc.openApprovalDecisionTask(tx, approval)
	}))

	_, err := svc.ProcessApprovalDecision("u-h-admin", approval.ID, &models.ApprovalDecisionRequest{
		Decision:                     models.ApprovalDecisionRequestChanges,
		DecisionReason:               "退回补充",
		ExpectedSubmissionSnapshotID: *approval.ActiveSubmissionSnapshotID,
		ImposedRequirements: map[string]interface{}{
			"items": []map[string]interface{}{
				{"category": "field_correction", "field_key": "case_title", "description": "标题不一致", "required": true},
			},
		},
	})
	require.NoError(t, err)

	// The approver handoff is closed...
	// This world already carries one historical request_changes record,
	// so the decision handoff round is c1 and the new revision round is c2.
	approverKey := buildApprovalTaskKey(inboxPurposeApprovalDecision, approval.ID, 1, "u-h-admin")
	closed := taskKey(t, db, approverKey)
	require.NotNil(t, closed)
	require.True(t, closed.IsCompleted)

	// ...and the applicant revision handoff for cycle 1 is open.
	revKey := buildApprovalTaskKey(inboxPurposeApplicantRevision, approval.ID, 2, "u-h-lawyer")
	open := taskKey(t, db, revKey)
	require.NotNil(t, open)
	require.False(t, open.IsCompleted)
	require.Equal(t, "workflow", open.CompletionMode)

	// C07: the owner cannot manually complete the workflow handoff.
	inbox := NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db))
	require.ErrorIs(t, inbox.MarkAsCompleted(context.Background(), open.ID, open.UserID), ErrWorkflowItemManualComplete)
}

// C05+C11: a structured resubmit closes the applicant's revision handoff
// and opens the next round's approver handoff; the closed revision record
// stays completed (no resurrection on replay).
func TestStructuredResubmit_HandoffRoundTrip(t *testing.T) {
	db := seedResubmitContractWorld(t)
	svc := NewApprovalService(db)
	intakeID := "INT-HR-1"
	// The applicant actor is stored as a username in this world; seed the
	// matching active user so the revision handoff can resolve its owner.
	require.NoError(t, db.Create(&models.User{
		Username: "u-lawyer",
		Name:     "lawyer",
		Email:    fmt.Sprintf("u-lawyer-%d@example.com", time.Now().UnixNano()),
		Password: "x",
		Role:     "lawyer",
		Status:   "active",
	}).Error)
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedPostCommitConflictWorld2(t, db, intakeID)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)

	// Seed the open revision handoff from the earlier request_changes.
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return svc.openApplicantRevisionTask(tx, approval)
	}))
	before := taskKey(t, db, buildApprovalTaskKey(inboxPurposeApplicantRevision, approval.ID, 1, "u-lawyer"))
	require.NotNil(t, before)
	require.False(t, before.IsCompleted)

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.NoError(t, err)

	closed := taskKey(t, db, buildApprovalTaskKey(inboxPurposeApplicantRevision, approval.ID, 1, "u-lawyer"))
	require.NotNil(t, closed)
	require.True(t, closed.IsCompleted)

	// The revived approval reassigns to the seeded active admin; the new
	// round's approver handoff must be open for that user (any admin row).
	var adminID uint
	require.NoError(t, db.Raw("SELECT id FROM users WHERE role = 'admin' AND status = 'active' ORDER BY id LIMIT 1").Scan(&adminID).Error)
	newKey := fmt.Sprintf("approval:%s:%s:c1:%d", inboxPurposeApprovalDecision, approval.ID, adminID)
	open := taskKey(t, db, newKey)
	require.NotNil(t, open)
	require.False(t, open.IsCompleted)
}

// C13 (SQLite part): a failing handoff write rolls back the revival so no
// snapshot survives a broken handoff. PG injects the same fault at the
// database level in approval_resubmit_pg_test.go.
func TestStructuredResubmit_FailedHandoffRollsBackRevival(t *testing.T) {
	db := seedResubmitContractWorld(t)
	svc := NewApprovalService(db)
	intakeID := "INT-HF-1"
	seedResubmitContractIntake(t, db, intakeID, "rev-1")
	approval, rec := seedPostCommitConflictWorld2(t, db, intakeID)
	initialSnap := seedActiveSubmissionRound(t, db, approval, intakeID)

	// Drop inbox_items after seeding the handoff table: the revival
	// transaction will fail when writing the approver handoff.
	require.NoError(t, db.Migrator().DropTable("inbox_items"))

	_, err := svc.StructuredResubmit("u-lawyer", approval.ID, pinnedResubmitRequest(rec, "rev-1", initialSnap))
	require.Error(t, err)

	// The revival rolled back: only the seeded first round survives; the approval stays needs_revision.
	var n int64
	require.NoError(t, db.Table("approval_snapshots").Where("approval_request_id = ?", approval.ID).Count(&n).Error)
	require.EqualValues(t, 1, n)
	var status string
	require.NoError(t, db.Raw("SELECT status FROM approval_requests WHERE id = ?", approval.ID).Scan(&status).Error)
	require.Equal(t, models.ApprovalStatusNeedsRevision, status)
}
