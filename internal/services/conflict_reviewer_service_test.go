package services

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func reviewerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "reviewer.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{}, &models.Case{}, &models.ConflictCheckRecord{},
		&models.ConflictReviewerAssignment{}, &models.ComplianceAuditEvent{},
		&models.ApprovalRequest{},
	))
	return db
}

func reviewerErrorCode(t *testing.T, err error) string {
	t.Helper()
	var reviewerErr *ConflictReviewerError
	require.True(t, errors.As(err, &reviewerErr))
	return reviewerErr.Code
}

func seedReviewerCheck(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Create(&models.User{ID: 10, Username: "applicant", Name: "申请律师", Email: "applicant@example.test", Password: "x", Role: "lawyer", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.User{ID: 20, Username: "officer", Name: "独立核查人", Email: "officer@example.test", Password: "x", Role: "conflict_officer", Status: "active"}).Error)
	require.NoError(t, db.Create(&models.ConflictCheckRecord{
		CheckID: "check-reviewer", UserID: 10, CheckStatus: "COMPLETED",
		SearchParameters: models.JSON{"subjectCaseId": "100"},
		CheckResult:      models.JSON{"decision": map[string]interface{}{"coverageStatus": "COMPLETE"}},
	}).Error)
	require.NoError(t, db.Create(&models.Case{ID: 100, CaseNumber: "CASE-REVIEWER", Title: "复核关系测试", LawyerID: 10, CreatedBy: "10", SubjectState: models.SubjectStateEffective}).Error)
}

func TestValidateConflictReviewerRequiresAssignment(t *testing.T) {
	db := reviewerTestDB(t)
	seedReviewerCheck(t, db)

	err := ValidateConflictReviewer(context.Background(), db, "check-reviewer", 100, 20, "conflict_officer")
	if code := reviewerErrorCode(t, err); code != "REVIEWER_ASSIGNMENT_REQUIRED" {
		t.Fatalf("expected assignment gate, got %s", code)
	}

	now := time.Now()
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{
		ID: "assignment-1", CheckID: "check-reviewer", CaseID: func() *uint { id := uint(100); return &id }(),
		ReviewerID: 20, AssignedBy: 20, Status: models.ConflictReviewerAssignmentActive,
		RecusalDeclared: true, EffectiveFrom: &now, CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, ValidateConflictReviewer(context.Background(), db, "check-reviewer", 100, 20, "conflict_officer"))
}

func TestValidateConflictReviewerRejectsTechnicalAdminAndDirectManager(t *testing.T) {
	db := reviewerTestDB(t)
	seedReviewerCheck(t, db)

	require.NoError(t, db.Model(&models.User{}).Where("id = ?", 20).Update("role", "admin").Error)
	err := ValidateConflictReviewer(context.Background(), db, "check-reviewer", 100, 20, "admin")
	if code := reviewerErrorCode(t, err); code != "REVIEWER_ROLE_FORBIDDEN" {
		t.Fatalf("expected technical admin denial, got %s", code)
	}

	// Restore the business role and make the applicant's direct manager the
	// proposed reviewer. Even a valid appointment cannot override recusal.
	require.NoError(t, db.Model(&models.User{}).Where("id = ?", 20).Updates(map[string]interface{}{"role": "conflict_officer", "manager_id": nil}).Error)
	reviewerID := uint(20)
	require.NoError(t, db.Model(&models.User{}).Where("id = ?", 10).Update("manager_id", reviewerID).Error)
	now := time.Now()
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{
		ID: "assignment-manager", CheckID: "check-reviewer", ReviewerID: 20, AssignedBy: 20,
		Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true,
		EffectiveFrom: &now, CreatedAt: now, UpdatedAt: now,
	}).Error)
	err = ValidateConflictReviewer(context.Background(), db, "check-reviewer", 100, 20, "conflict_officer")
	if code := reviewerErrorCode(t, err); code != "REVIEWER_CONFLICTED" {
		t.Fatalf("expected direct-manager recusal, got %s", code)
	}
}

func TestValidateConflictReviewerRejectsSupersededAssignment(t *testing.T) {
	db := reviewerTestDB(t)
	seedReviewerCheck(t, db)
	require.NoError(t, db.Create(&models.User{
		ID: 21, Username: "new-reviewer", Name: "新复核人", Email: "new-reviewer@example.test",
		Password: "x", Role: "compliance", Status: "active",
	}).Error)
	oldTime := time.Now().Add(-time.Hour)
	newTime := time.Now()
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{
		ID: "assignment-old", CheckID: "check-reviewer", ReviewerID: 20, AssignedBy: 20,
		Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true,
		EffectiveFrom: &oldTime, CreatedAt: oldTime, UpdatedAt: oldTime,
	}).Error)
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{
		ID: "assignment-new", CheckID: "check-reviewer", ReviewerID: 21, AssignedBy: 20,
		Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true,
		EffectiveFrom: &newTime, CreatedAt: newTime, UpdatedAt: newTime,
	}).Error)

	err := ValidateConflictReviewer(context.Background(), db, "check-reviewer", 100, 20, "conflict_officer")
	if code := reviewerErrorCode(t, err); code != "REVIEWER_ASSIGNMENT_MISMATCH" {
		t.Fatalf("expected superseded reviewer mismatch, got %s", code)
	}
	require.NoError(t, ValidateConflictReviewer(context.Background(), db, "check-reviewer", 100, 21, "compliance"))
}

func TestAssignConflictReviewerIsIdempotentAndAudited(t *testing.T) {
	db := reviewerTestDB(t)
	seedReviewerCheck(t, db)
	actor := AuthActor{UserID: 20, Role: "conflict_officer"}

	assignment, err := AssignConflictReviewer(context.Background(), db, actor, "check-reviewer", ConflictReviewerAssignmentInput{
		ReviewerID: 20, RecusalDeclared: true, IndependenceReason: "与申请律师不存在直接管理关系",
	})
	require.NoError(t, err)
	require.NotEmpty(t, assignment.ID)

	retried, err := AssignConflictReviewer(context.Background(), db, actor, "check-reviewer", ConflictReviewerAssignmentInput{
		ReviewerID: 20, RecusalDeclared: true,
	})
	require.NoError(t, err)
	if retried.ID != assignment.ID {
		t.Fatalf("retry must return the same active appointment: %s != %s", retried.ID, assignment.ID)
	}
	var auditCount int64
	require.NoError(t, db.Model(&models.ComplianceAuditEvent{}).Where("event_type = ?", "CONFLICT_REVIEWER_ASSIGNED").Count(&auditCount).Error)
	if auditCount != 1 {
		t.Fatalf("expected one assignment audit, got %d", auditCount)
	}
}

func TestAssignConflictReviewerSyncsOpenConflictApprovalApprover(t *testing.T) {
	db := reviewerTestDB(t)
	seedReviewerCheck(t, db)
	require.NoError(t, db.Create(&models.User{
		ID: 21, Username: "compliance", Name: "合规复核人", Email: "compliance@example.test",
		Password: "x", Role: "compliance", Status: "active",
	}).Error)
	approvalID := "approval-reviewer-sync"
	require.NoError(t, db.Create(&models.ApprovalRequest{
		ID: approvalID, RequestNumber: "APR-REVIEWER-SYNC", Title: "冲突审批", Type: "conflict_approval",
		Content: "冲突审批", ApplicantID: "10", ApplicantName: "申请律师", Status: models.ApprovalStatusSubmitted,
		CurrentApproverID: "20", CurrentApproverName: "独立核查人", WorkflowType: "CONFLICT_APPROVAL",
		ConflictCheckID: "check-reviewer", CreatedBy: "10",
		UpdatedAt: time.Now(),
	}).Error)

	assignment, err := AssignConflictReviewer(
		context.Background(),
		db,
		AuthActor{UserID: 20, Role: "conflict_officer"},
		"check-reviewer",
		ConflictReviewerAssignmentInput{ReviewerID: 21, RecusalDeclared: true, IndependenceReason: "与申请律师和指定人不存在管理关系"},
	)
	require.NoError(t, err)
	require.NotEmpty(t, assignment.ID)

	var approval models.ApprovalRequest
	require.NoError(t, db.Where("id = ?", approvalID).Take(&approval).Error)
	require.Equal(t, "21", approval.CurrentApproverID)
	var approverName string
	require.NoError(t, db.Model(&models.ApprovalRequest{}).Where("id = ?", approvalID).
		Pluck("current_approver_name", &approverName).Error)
	require.Equal(t, "合规复核人", approverName)

	require.NoError(t, db.Model(&models.ApprovalRequest{}).Where("id = ?", approvalID).
		Updates(map[string]interface{}{"current_approver_id": "20", "current_approver_name": "独立核查人"}).Error)
	_, err = AssignConflictReviewer(
		context.Background(),
		db,
		AuthActor{UserID: 20, Role: "conflict_officer"},
		"check-reviewer",
		ConflictReviewerAssignmentInput{ReviewerID: 21, RecusalDeclared: true},
	)
	require.NoError(t, err)
	require.NoError(t, db.Where("id = ?", approvalID).Take(&approval).Error)
	require.Equal(t, "21", approval.CurrentApproverID)
}
