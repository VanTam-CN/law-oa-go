package services

// Stage A regression: the inbox action projection must expose the four
// list-page columns the five-direction review found missing. related_title
// resolves through the same batched query as the action (no per-row reads,
// no guessed routing); next_step is a fixed server sentence, not a label
// the frontend derives.

import (
	"fmt"
	"testing"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func inboxProjectionFixture(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// cases + the ethical-wall whitelist back the case_task access check in
	// inbox_workflow_action.go; without the whitelist table the EXISTS probe
	// throws "no such table" in SQLite and every row degrades to blocked.
	if err := db.AutoMigrate(
		&models.InboxItem{},
		&models.ApprovalRequest{},
		&models.ConflictReviewerAssignment{},
		&models.Case{},
		&models.CaseEthicalWallWhitelist{},
		&models.CaseTask{},
		&models.User{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seedCaseTaskFixture inserts the owning case and the assigned task for the
// case_task projection regressions.
func seedCaseTaskFixture(t *testing.T, db *gorm.DB, caseID uint, caseLawyerID uint, taskID uint) {
	t.Helper()
	// The case_task projection gate requires a live account (deleted_at IS
	// NULL, status='active', non-technical-admin role) before any grant
	// counts, exactly like caseTaskAssigneeAllowed. Seed the owning lawyer
	// so the fixture models the real eligibility chain; deactivated or
	// deleted-account negatives seed their own rows.
	if err := db.Create(&models.User{ID: caseLawyerID, Username: fmt.Sprintf("proj-lawyer-%d", caseLawyerID),
		Email:    fmt.Sprintf("proj-lawyer-%d@example.test", caseLawyerID),
		Password: "x", Role: "lawyer", Status: "active"}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Exec(
		"INSERT INTO cases (id, case_number, title, client_id, lawyer_id, case_type, created_by, ethical_wall_enabled) VALUES (?, ?, '投影回归案件', 1, ?, 'civil', ?, FALSE)",
		caseID, fmt.Sprintf("TEST-CASE-%d", caseID), caseLawyerID, fmt.Sprint(caseLawyerID),
	).Error; err != nil {
		t.Fatalf("seed case: %v", err)
	}
	task := &models.CaseTask{
		ID:         taskID,
		CaseID:     caseID,
		Title:      "投影回归任务",
		AssigneeID: 7,
		CreatedBy:  caseLawyerID,
		Status:     "open",
	}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("seed task: %v", err)
	}
}

func TestInboxProjection_ApprovalTodoCarriesRelatedTitleAndNextStep(t *testing.T) {
	db := inboxProjectionFixture(t)
	approval := &models.ApprovalRequest{
		ID:                "apm-proj-1",
		Title:             "华信并购专项冲突审批",
		Status:            "submitted",
		CurrentApproverID: "7",
	}
	if err := db.Create(approval).Error; err != nil {
		t.Fatalf("seed approval: %v", err)
	}
	item := &models.InboxItem{
		UserID:          7,
		SourceType:      "approval",
		SourceID:        1,
		SourceRef:       "apm-proj-1",
		Title:           "请审批",
		CompletionMode:  "workflow",
		WorkflowTaskKey: "approval:apm-proj-1:decide",
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed inbox: %v", err)
	}
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	resp := svc.TestProjectWorkflowAction(7, item)
	if resp == nil {
		t.Fatal("expected workflow action projection, got nil")
	}
	if resp.RelatedTitle != approval.Title {
		t.Fatalf("related_title = %q, want %q", resp.RelatedTitle, approval.Title)
	}
	if resp.NextStep == "" {
		t.Fatal("next_step must be present for an actionable approval todo")
	}
	if resp.AssignedTo == "" {
		t.Fatal("assigned_to must be populated for the list page")
	}
}

func TestInboxProjection_BlockedRowStillShowsReasonAndStep(t *testing.T) {
	db := inboxProjectionFixture(t)
	item := &models.InboxItem{
		UserID:          7,
		SourceType:      "approval",
		SourceID:        1,
		SourceRef:       "apm-missing",
		Title:           "孤儿待办",
		CompletionMode:  "workflow",
		WorkflowTaskKey: "approval:apm-missing:decide",
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed inbox: %v", err)
	}
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	resp := svc.TestProjectWorkflowAction(7, item)
	if resp == nil {
		t.Fatal("expected blocked projection, got nil")
	}
	if resp.BlockedCode == "" || resp.BlockedReason == "" {
		t.Fatal("blocked row must carry code and reason for the list column")
	}
	if resp.NextStep == "" {
		t.Fatal("blocked row must still state the next step")
	}
}

func TestInboxProjection_CaseTaskTodoUsesExplicitCaseID(t *testing.T) {
	db := inboxProjectionFixture(t)
	seedCaseTaskFixture(t, db, 31, 7, 41)
	taskID := uint(41)
	item := &models.InboxItem{
		UserID:          7,
		SourceType:      "task",
		SourceID:        taskID,
		SourceRef:       fmt.Sprint(taskID),
		Title:           "任务待办",
		CompletionMode:  "workflow",
		WorkflowTaskKey: fmt.Sprintf("case_task:%d:open", taskID),
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed inbox: %v", err)
	}
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	resp := svc.TestProjectWorkflowAction(7, item)
	if resp == nil {
		t.Fatal("expected workflow action projection, got nil")
	}
	if resp.ResourceType != "case_task" {
		t.Fatalf("resource_type = %q, want case_task", resp.ResourceType)
	}
	if resp.ResourceID != fmt.Sprint(taskID) {
		t.Fatalf("resource_id = %q, want the TASK id %d (never a case id)", resp.ResourceID, taskID)
	}
	if resp.RelatedCaseID != "31" {
		t.Fatalf("related_case_id = %q, want the explicit case id 31", resp.RelatedCaseID)
	}
	if resp.ActionCode != "open_case_task" {
		t.Fatalf("action_code = %q, want open_case_task for an authorized open task", resp.ActionCode)
	}
}

func TestInboxProjection_CaseTaskTodoFailsClosedWithoutCaseAccess(t *testing.T) {
	db := inboxProjectionFixture(t)
	// Case 32 belongs to lawyer 9; actor 7 has neither ownership nor a
	// whitelist entry, so the projection must degrade to a blocked action.
	seedCaseTaskFixture(t, db, 32, 9, 42)
	taskID := uint(42)
	item := &models.InboxItem{
		UserID:          7,
		SourceType:      "task",
		SourceID:        taskID,
		SourceRef:       fmt.Sprint(taskID),
		Title:           "越权任务",
		CompletionMode:  "workflow",
		WorkflowTaskKey: fmt.Sprintf("case_task:%d:open", taskID),
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatalf("seed inbox: %v", err)
	}
	svc := NewInboxService(repositories.NewInboxRepository(db), nil)
	resp := svc.TestProjectWorkflowAction(7, item)
	if resp == nil {
		t.Fatal("expected blocked projection, got nil")
	}
	if resp.ActionCode != "view_only" || resp.BlockedCode == "" {
		t.Fatalf("revoked access must fail closed, got action_code=%q blocked_code=%q", resp.ActionCode, resp.BlockedCode)
	}
	if resp.RelatedCaseID != "" {
		t.Fatalf("blocked projection must not leak the case id, got %q", resp.RelatedCaseID)
	}
}
