package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func inboxWorkflowDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "inbox.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.InboxItem{}))
	return db
}

func TestInboxWorkflowCannotBeManuallyCompleted(t *testing.T) {
	db := inboxWorkflowDB(t)
	item := models.InboxItem{UserID: 1, SourceType: "approval", SourceRef: "approval-uuid", Title: "审批任务", CompletionMode: "workflow"}
	require.NoError(t, db.Create(&item).Error)
	err := repositories.NewInboxRepository(db).MarkAsCompletedByUserID(context.Background(), item.ID, 1)
	require.ErrorIs(t, err, repositories.ErrInboxWorkflowActionRequired)
	require.NoError(t, db.First(&item, item.ID).Error)
	require.False(t, item.IsCompleted)
}

func TestInboxLegacyArchiveIsNotBusinessCompletion(t *testing.T) {
	db := inboxWorkflowDB(t)
	item := models.InboxItem{UserID: 1, SourceType: "task", SourceID: 999, Title: "受限历史标题", Content: "受限意见"}
	require.NoError(t, db.Create(&item).Error)
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	response, err := svc.GetInboxItemByID(context.Background(), item.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "legacy_unresolved", response.CompletionMode)
	require.NotContains(t, response.Title, "受限")
	require.Empty(t, response.Content)
	require.Zero(t, response.SourceID)
	require.NoError(t, svc.ArchiveInboxItem(context.Background(), item.ID, 1))
	require.NoError(t, db.First(&item, item.ID).Error)
	require.NotNil(t, item.ArchivedAt)
	require.False(t, item.IsCompleted)
}

func seedInboxWorkflow(t *testing.T) (*gorm.DB, *InboxService, *models.ApprovalRequest) {
	t.Helper()
	db := inboxWorkflowDB(t)
	require.NoError(t, db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{}, &models.ApprovalRequest{}, &models.ApprovalRecord{}, &models.ApprovalSnapshot{}, &models.ConflictCheckRecord{}, &models.ConflictReviewerAssignment{}, &models.ConflictReview{}))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, created_by TEXT, client_id INTEGER, status TEXT, metadata TEXT)").Error)
	require.NoError(t, db.Exec("CREATE TABLE case_ethical_wall_whitelist (case_id INTEGER, user_id INTEGER)").Error)
	for _, u := range []models.User{
		{ID: 1, Username: "lawyer", Email: "lawyer@test.invalid", Name: "申请律师", Password: "x", Role: "lawyer", Status: "active"},
		{ID: 2, Username: "reviewer", Email: "reviewer@test.invalid", Name: "核查人", Password: "x", Role: "partner", Status: "active"},
		{ID: 3, Username: "next", Email: "next@test.invalid", Name: "新审批人", Password: "x", Role: "partner", Status: "active"},
		{ID: 4, Username: "other", Email: "other@test.invalid", Name: "无关律师", Password: "x", Role: "lawyer", Status: "active"},
	} {
		require.NoError(t, db.Create(&u).Error)
	}
	require.NoError(t, db.Create(&models.Client{ID: 100, Name: "待办客户", Type: "个人", CreatedBy: 1}).Error)
	require.NoError(t, db.Create(&models.Case{ID: 10, CaseNumber: "INBOX-10", Title: "无敏感正文", ClientID: 100, LawyerID: 1, CreatedBy: "1"}).Error)
	require.NoError(t, db.Exec(`INSERT INTO case_intakes VALUES ('intake-uuid','1',100,'conflict_ready','{"conflict_check_id":"check-uuid"}')`).Error)
	require.NoError(t, db.Create(&models.ConflictCheckRecord{CheckID: "check-uuid", ClientID: "100", UserID: 1, CheckStatus: "COMPLETED", CheckTime: time.Now(), SearchParameters: models.JSON{"intakeId": "intake-uuid"}, CheckResult: models.JSON{"decision": map[string]interface{}{"coverageStatus": "COMPLETE"}}}).Error)
	snapshot := "snapshot-one"
	a := &models.ApprovalRequest{ID: "approval-uuid", RequestNumber: "APR-INBOX", Type: "case_creation", Title: "测试审批", Content: "测试", ApplicantID: "1", ApplicantName: "申请律师", CurrentApproverID: "2", CurrentApproverName: "核查人", CurrentStage: "初审", Status: models.ApprovalStatusSubmitted, ConflictCheckID: "check-uuid", Metadata: `{"intake_id":"intake-uuid","conflict_check_id":"check-uuid"}`, ActiveSubmissionSnapshotID: &snapshot}
	require.NoError(t, db.Create(a).Error)
	seedInboxSubmission(t, db, a)
	return db, NewInboxService(repositories.NewInboxRepository(db), repositories.NewUserRepository(db)), a
}

func seedInboxSubmission(t *testing.T, db *gorm.DB, a *models.ApprovalRequest) {
	t.Helper()
	metadata, err := submissionMetadata(a.Metadata)
	require.NoError(t, err)
	raw, err := json.Marshal(map[string]interface{}{
		"snapshot_id": *a.ActiveSubmissionSnapshotID, "approval_request_id": a.ID,
		"source_version": 1, "intake_id": metadata["intake_id"],
		"conflict_check_id": a.ConflictCheckID, "metadata": metadata,
	})
	require.NoError(t, err)
	require.NoError(t, db.Create(&models.ApprovalSnapshot{ID: *a.ActiveSubmissionSnapshotID, ApprovalRequestID: a.ID,
		SnapshotType: models.IntakeSubmissionSnapshotType, SourceVersion: 1, SnapshotData: string(raw)}).Error)
}

func inboxReconcile(t *testing.T, db *gorm.DB, a *models.ApprovalRequest) {
	t.Helper()
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error { return ReconcileApprovalInboxInTx(tx, a) }))
}

func TestInboxApprovalHandoffIdempotentAndRollback(t *testing.T) {
	db, svc, a := seedInboxWorkflow(t)
	require.Error(t, ReconcileApprovalInboxInTx(db, a))
	inboxReconcile(t, db, a)
	inboxReconcile(t, db, a)
	var count int64
	require.NoError(t, db.Model(&models.InboxItem{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	page, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{Page: 1, PageSize: 10, WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "approval-uuid", page.Items[0].WorkflowAction.ResourceID)
	require.Equal(t, "decide_approval", page.Items[0].WorkflowAction.ActionCode)
	require.Equal(t, a.UpdatedAt.UTC().Format(time.RFC3339Nano), page.Items[0].WorkflowAction.ActionVersion)
	require.EqualValues(t, 1, page.Stats.NeedsAction)
	require.ErrorIs(t, svc.MarkAsCompleted(context.Background(), page.Items[0].ID, 2), repositories.ErrInboxWorkflowActionRequired)
	require.ErrorIs(t, svc.ArchiveInboxItem(context.Background(), page.Items[0].ID, 2), repositories.ErrInboxWorkflowActionRequired)
	// A failure after old-task closure must also roll back the outer business update.
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("inbox:fail", func(tx *gorm.DB) {
		if tx.Statement.Table == "inbox_items" {
			tx.AddError(errors.New("injected inbox write failure"))
		}
	}))
	transition := func(tx *gorm.DB) error {
		if err := tx.Model(&models.ApprovalRequest{}).Where("id = ?", a.ID).Update("status", models.ApprovalStatusNeedsRevision).Error; err != nil {
			return err
		}
		var changed models.ApprovalRequest
		if err := tx.First(&changed, "id = ?", a.ID).Error; err != nil {
			return err
		}
		return ReconcileApprovalInboxInTx(tx, &changed)
	}
	require.ErrorContains(t, db.Transaction(transition), "injected")
	var saved models.ApprovalRequest
	require.NoError(t, db.First(&saved, "id = ?", a.ID).Error)
	require.Equal(t, models.ApprovalStatusSubmitted, saved.Status)
	require.NoError(t, db.Model(&models.InboxItem{}).Where("is_completed = ?", false).Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, db.Callback().Create().Remove("inbox:fail"))
	require.NoError(t, db.Transaction(transition))
	require.NoError(t, db.First(&saved, "id = ?", a.ID).Error)
	inboxReconcile(t, db, &saved)
	page, err = svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "intake-uuid", page.Items[0].WorkflowAction.ResourceID)
	require.Equal(t, "approval-uuid", page.Items[0].WorkflowAction.ApprovalID)
	require.Equal(t, "supplement", page.Items[0].WorkflowAction.ActionCode)
	require.NoError(t, db.Model(&models.InboxItem{}).Count(&count).Error)
	require.EqualValues(t, 3, count)
}

func TestInboxSupplementActionJSONContract(t *testing.T) {
	db, svc, approval := seedInboxWorkflow(t)
	approval.Status = models.ApprovalStatusNeedsRevision
	require.NoError(t, db.Save(approval).Error)
	inboxReconcile(t, db, approval)
	page, err := svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	detail, err := svc.GetInboxItemByID(context.Background(), page.Items[0].ID, 1)
	require.NoError(t, err)
	for _, item := range []*InboxItemResponse{page.Items[0], detail} {
		raw, err := json.Marshal(item)
		require.NoError(t, err)
		var wire map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &wire))
		action, ok := wire["workflow_action"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "intake", action["resource_type"])
		require.Equal(t, "supplement", action["action_code"])
		require.Equal(t, "intake-uuid", action["resource_id"])
		approvalID, ok := action["approval_id"].(string)
		require.True(t, ok, "the client requires an explicit string approval_id")
		require.Equal(t, approval.ID, approvalID)
		require.NotEqual(t, action["resource_id"], approvalID)
	}
	// A stale/mismatched association must not produce a guessed revision route.
	require.NoError(t, db.Model(approval).Update("metadata", `{"intake_id":"different-intake","conflict_check_id":"check-uuid"}`).Error)
	_, err = svc.GetInboxItemByID(context.Background(), page.Items[0].ID, 1)
	require.ErrorIs(t, err, repositories.ErrInboxItemNotFound)
	page, err = svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Zero(t, page.Pagination.Total)
}

func TestInboxAuthorizationBeforePaginationAndCount(t *testing.T) {
	db, svc, a := seedInboxWorkflow(t)
	inboxReconcile(t, db, a)
	// A forged/stale assignee row is not an authorization grant.
	unrelated := newWorkflowInbox("approval:"+a.ID, approvalInboxVersion(a), "approval", a.ID, "decide_approval", 4, "不能泄漏的标题", false)
	unrelated.Content = "不能泄漏的补充意见"
	require.NoError(t, db.Create(unrelated).Error)
	for _, userID := range []uint{4, 2} {
		if userID == 2 {
			require.NoError(t, db.Model(&models.Case{}).Where("id = 10").Update("ethical_wall_enabled", true).Error)
		}
		page, err := svc.ListInboxItems(context.Background(), userID, &ListInboxItemsRequest{Page: 1, PageSize: 1})
		require.NoError(t, err)
		require.Empty(t, page.Items)
		require.Zero(t, page.Pagination.Total)
		require.Zero(t, page.Stats.Total)
	}
	_, err := svc.GetInboxItemByID(context.Background(), unrelated.ID, 4)
	require.ErrorIs(t, err, repositories.ErrInboxItemNotFound)
	for i := 0; i < 2; i++ {
		require.NoError(t, db.Create(&models.InboxItem{UserID: 4, SourceType: "task", CompletionMode: "manual", Title: "可见个人提醒", CreatedAt: time.Now().Add(-24 * time.Hour)}).Error)
	}
	visiblePage, visibleErr := svc.ListInboxItems(context.Background(), 4, &ListInboxItemsRequest{Page: 1, PageSize: 1, OrderBy: "created_at"})
	require.NoError(t, visibleErr)
	require.EqualValues(t, 2, visiblePage.Pagination.Total)
	require.Len(t, visiblePage.Items, 1)
	require.Equal(t, "可见个人提醒", visiblePage.Items[0].Title)
	require.NoError(t, db.Exec("INSERT INTO case_ethical_wall_whitelist VALUES (10,2)").Error)
	page, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{Page: 1, PageSize: 1})
	require.NoError(t, err)
	require.EqualValues(t, 1, page.Pagination.Total)
	require.Len(t, page.Items, 1)
	for _, role := range []string{"admin", "user", "assistant", "lawyer"} {
		require.NoError(t, db.Model(&models.User{}).Where("id = 2").Update("role", role).Error)
		page, err = svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{})
		require.NoError(t, err)
		require.Zero(t, page.Stats.NeedsAction)
	}
}

func TestInboxMoreThan1000AndSnoozeKeepsActionableCount(t *testing.T) {
	db := inboxWorkflowDB(t)
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	until := time.Now().Add(24 * time.Hour)
	items := make([]models.InboxItem, 1107)
	for i := range items {
		items[i] = models.InboxItem{UserID: 1, SourceType: "task", CompletionMode: "manual", Title: fmt.Sprintf("提醒%04d", i), Priority: "medium", SnoozedUntil: &until}
	}
	require.NoError(t, db.CreateInBatches(&items, 100).Error)
	page, err := svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{Page: 12, PageSize: 100, WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 7)
	require.EqualValues(t, 1107, page.Pagination.Total)
	require.EqualValues(t, 1107, page.Stats.NeedsAction)
	stats, err := svc.GetInboxStats(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, page.Stats, stats)
	require.NoError(t, svc.MarkAsCompleted(context.Background(), items[0].ID, 1))
	stats, err = svc.GetInboxStats(context.Background(), 1)
	require.NoError(t, err)
	require.EqualValues(t, 1106, stats.Pending)
	require.EqualValues(t, 1, stats.Completed)
}

func TestInboxRejectsForgedServiceFields(t *testing.T) {
	svc := NewInboxService(nil, nil)
	for _, req := range []*CreateInboxItemRequest{
		{SourceType: "approval", SourceID: 1}, {SourceType: "task", SourceID: 2}, {CompletionMode: json.RawMessage(`"manual"`)}, {SourceRef: json.RawMessage(`null`)}, {WorkflowTaskKey: json.RawMessage(`"forged"`)}, {AssignedTo: json.RawMessage(`2`)}, {UserID: 2},
	} {
		_, err := svc.CreateInboxItem(context.Background(), 1, req)
		require.ErrorIs(t, err, ErrInboxProtectedField)
	}
}

func TestInboxReassignmentAndCaseFailureProjection(t *testing.T) {
	db, svc, a := seedInboxWorkflow(t)
	inboxReconcile(t, db, a)
	var old models.InboxItem
	require.NoError(t, db.Where("user_id = ? AND workflow_purpose = ?", 2, "decide_approval").First(&old).Error)
	require.NoError(t, db.Create(&models.ApprovalRecord{ID: "prior-review", ApprovalRequestID: a.ID, ApproverID: "2", Decision: "reassign"}).Error)
	a.CurrentApproverID = "3"
	require.NoError(t, db.Save(a).Error)
	inboxReconcile(t, db, a)
	oldResponse, err := svc.GetInboxItemByID(context.Background(), old.ID, 2)
	require.NoError(t, err)
	require.Equal(t, "view_only", oldResponse.WorkflowAction.ActionCode)
	require.Equal(t, "transferred", oldResponse.ClosedReason)
	page, err := svc.ListInboxItems(context.Background(), 3, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	// Reassigning back in the same snapshot must create a new generation,
	// rather than collide with the original (now closed) task's key.
	a.CurrentApproverID = "2"
	require.NoError(t, db.Save(a).Error)
	inboxReconcile(t, db, a)
	back, backErr := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, backErr)
	require.Len(t, back.Items, 1)
	require.NotEqual(t, old.ID, back.Items[0].ID)
	a.Status = models.ApprovalStatusApproved
	a.CurrentApproverID = ""
	a.CaseCreationStatus = "pending"
	require.NoError(t, db.Save(a).Error)
	inboxReconcile(t, db, a)
	var resultBefore models.InboxItem
	require.NoError(t, db.Where("workflow_group = ? AND workflow_purpose = ?", "approval:"+a.ID, "result").First(&resultBefore).Error)
	pending, err := svc.GetInboxItemByID(context.Background(), resultBefore.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "approved_pending_case", pending.WorkflowAction.WorkflowStatus)
	require.Equal(t, "CASE_CREATION_PENDING", pending.WorkflowAction.BlockedCode)
	require.True(t, pending.IsCompleted) // Approval outcome, not a completed case.
	require.Equal(t, "completed", pending.WorkflowBucket)
	a.CaseCreationStatus = "failed"
	require.NoError(t, db.Save(a).Error)
	inboxReconcile(t, db, a)
	page, err = svc.ListInboxItems(context.Background(), 1, &ListInboxItemsRequest{})
	require.NoError(t, err)
	found := false
	for _, item := range page.Items {
		if item.WorkflowAction.WorkflowStatus == "case_creation_failed" {
			found = true
			require.Equal(t, "CASE_CREATION_FAILED", item.WorkflowAction.BlockedCode)
		}
	}
	require.True(t, found)
	// Case progress updates the same result projection, never creates a second
	// completed outcome notification for the same approved submission.
	failed, err := svc.GetInboxItemByID(context.Background(), resultBefore.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "case_creation_failed", failed.WorkflowAction.WorkflowStatus)
	require.Equal(t, "view_only", failed.WorkflowAction.ActionCode)
	require.Equal(t, "completed", failed.WorkflowBucket)
	a.CaseCreated = true
	a.CreatedCaseID = "10"
	a.CaseCreationStatus = "completed"
	require.NoError(t, db.Save(a).Error)
	inboxReconcile(t, db, a)
	incomplete, err := svc.GetInboxItemByID(context.Background(), resultBefore.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "case_creation_incomplete", incomplete.WorkflowAction.WorkflowStatus)
	require.NoError(t, db.Model(&models.Case{}).Where("id = 10").Update("conflict_check_id", "check-uuid").Error)
	inboxReconcile(t, db, a)
	var resultCount int64
	require.NoError(t, db.Model(&models.InboxItem{}).Where("workflow_group = ? AND workflow_purpose = ?", "approval:"+a.ID, "result").Count(&resultCount).Error)
	require.EqualValues(t, 1, resultCount)
	resultAfter, err := svc.GetInboxItemByID(context.Background(), resultBefore.ID, 1)
	require.NoError(t, err)
	require.Equal(t, "case_created", resultAfter.WorkflowAction.WorkflowStatus)
}

func TestInboxReconcileReadsDurableApprovalInsteadOfCallerTimestamp(t *testing.T) {
	db, _, a := seedInboxWorkflow(t)
	inboxReconcile(t, db, a)
	stale := *a
	stale.UpdatedAt = stale.UpdatedAt.Add(523 * time.Nanosecond)
	stale.Status = models.ApprovalStatusNeedsRevision
	inboxReconcile(t, db, &stale)
	var count int64
	require.NoError(t, db.Model(&models.InboxItem{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
	require.NoError(t, db.Model(&models.InboxItem{}).Where("workflow_purpose = ?", "supplement").Count(&count).Error)
	require.Zero(t, count)
}

func TestInboxTrustedLegacyHandoffDoesNotGainBusinessCapability(t *testing.T) {
	db, svc, _ := seedInboxWorkflow(t)
	response, err := svc.CreateInboxItem(context.Background(), 1, &CreateInboxItemRequest{UserID: 1, SourceType: "handoff", SourceID: 100, Title: "旧来源移交", Priority: "medium"})
	require.NoError(t, err)
	require.Equal(t, "legacy_unresolved", response.CompletionMode)
	require.Empty(t, response.WorkflowAction.ResourceID)
	var row models.InboxItem
	require.NoError(t, db.First(&row, response.ID).Error)
	require.EqualValues(t, 100, row.SourceID)
}

func TestInboxConflictAssignmentCompletionAndRollback(t *testing.T) {
	db, svc, _ := seedInboxWorkflow(t)
	now := time.Now().Add(-time.Minute)
	assignment := models.ConflictReviewerAssignment{ID: "assignment-one", CheckID: "check-uuid", ReviewerID: 2, AssignedBy: 3, Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, EffectiveFrom: &now, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&assignment).Error)
	call := func(tx *gorm.DB) error { return ReconcileConflictInboxInTx(tx, "check-uuid") }
	require.NoError(t, db.Transaction(call))
	require.NoError(t, db.Transaction(call))
	page, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "review_conflict", page.Items[0].WorkflowAction.ActionCode)
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("inbox:conflict-fail", func(tx *gorm.DB) {
		if tx.Statement.Table == "inbox_items" {
			tx.AddError(errors.New("injected conflict handoff failure"))
		}
	}))
	review := func(tx *gorm.DB) error {
		if err := tx.Create(&models.ConflictReview{CheckID: "check-uuid", ReviewerID: 2, Decision: "no_conflict", Notes: "独立测试"}).Error; err != nil {
			return err
		}
		return call(tx)
	}
	require.Error(t, db.Transaction(review))
	var count int64
	require.NoError(t, db.Model(&models.ConflictReview{}).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, db.Callback().Create().Remove("inbox:conflict-fail"))
	require.NoError(t, db.Transaction(review))
	page, err = svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Zero(t, page.Pagination.Total)
	// Subsequent assignment administration cannot duplicate an immutable review outcome.
	require.NoError(t, db.Model(&models.ConflictReviewerAssignment{}).Where("id = ?", assignment.ID).Updates(map[string]interface{}{"status": models.ConflictReviewerAssignmentRevoked, "updated_at": time.Now()}).Error)
	require.NoError(t, db.Transaction(call))
	require.NoError(t, db.Model(&models.InboxItem{}).Where("workflow_group = ? AND workflow_purpose = ?", "conflict:check-uuid", "result").Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestInboxProjectionUsesBatchQueries(t *testing.T) {
	db, svc, a := seedInboxWorkflow(t)
	inboxReconcile(t, db, a)
	queries := 0
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("inbox:count", func(*gorm.DB) { queries++ }))
	_, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{})
	require.NoError(t, err)
	baseline := queries
	for i := 0; i < 75; i++ {
		item := newWorkflowInbox("approval:"+a.ID, approvalInboxVersion(a), "approval", a.ID, fmt.Sprintf("notification-%d", i), 2, "已授权通知", false)
		require.NoError(t, db.Create(item).Error)
	}
	queries = 0
	_, err = svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{})
	require.NoError(t, err)
	require.LessOrEqual(t, queries, baseline+1, "queries must not grow per inbox row")
}

func TestInboxConflictReassignmentAndReturnToEarlierActiveAssignment(t *testing.T) {
	db, svc, _ := seedInboxWorkflow(t)
	now := time.Now().Add(-time.Minute)
	a := models.ConflictReviewerAssignment{ID: "first", CheckID: "check-uuid", ReviewerID: 2, Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, db.Create(&a).Error)
	call := func(tx *gorm.DB) error { return ReconcileConflictInboxInTx(tx, "check-uuid") }
	require.NoError(t, db.Transaction(call))
	page, err := svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	oldID := page.Items[0].ID
	b := a
	b.ID = "second"
	b.ReviewerID = 3
	b.CreatedAt = now.Add(time.Second)
	b.UpdatedAt = b.CreatedAt
	require.NoError(t, db.Create(&b).Error)
	require.NoError(t, db.Transaction(call))
	page, err = svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Empty(t, page.Items)
	page, err = svc.ListInboxItems(context.Background(), 3, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	// Mirrors the existing latest-active-assignment authority: revoking the
	// newest assignment makes the earlier active one current again.
	require.NoError(t, db.Model(&models.ConflictReviewerAssignment{}).Where("id = ?", b.ID).Updates(map[string]interface{}{"status": models.ConflictReviewerAssignmentRevoked, "updated_at": time.Now()}).Error)
	require.NoError(t, db.Transaction(call))
	require.NoError(t, db.Transaction(call))
	page, err = svc.ListInboxItems(context.Background(), 2, &ListInboxItemsRequest{WorkflowBucket: "needs_action"})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.NotEqual(t, oldID, page.Items[0].ID)
}
