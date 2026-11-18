package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// F01 regression: saving an independent conflict review through the real
// ReviewConflict transaction must also close the assigned reviewer's
// workflow inbox handoff in the same transaction. The suite below drives
// the production ReviewConflict service over an isolated SQLite database;
// no business function is mocked.

const (
	f01CheckID   = "CCT_F01_REVIEW_CLOSE"
	f01OwnerID   = uint(101)
	f01Reviewer  = uint(202)
	f01IntakeID  = "INT_F01_REVIEW_CLOSE"
	f01InboxKey  = "conflict_review:" + f01CheckID + ":" + "202"
	f01OwnerRole = "lawyer"
)

func newF01ReviewDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/f01review.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.User{},
		&models.ConflictCheckRecord{},
		&models.ConflictReview{},
		&models.ConflictReviewerAssignment{},
		&models.InboxItem{},
		&models.ComplianceAuditEvent{},
	))
	require.NoError(t, db.Exec("CREATE TABLE case_intakes (id TEXT PRIMARY KEY, created_by INTEGER, metadata TEXT)").Error)
	return db
}

func seedF01Users(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now()
	require.NoError(t, db.Create(&models.User{
		ID: f01OwnerID, Username: "f01-owner", Name: "申请律师", Email: "f01-owner@example.invalid",
		Password: "x", Role: f01OwnerRole, Status: "active", CreatedAt: now, UpdatedAt: now,
	}).Error)
	require.NoError(t, db.Create(&models.User{
		ID: f01Reviewer, Username: "f01-reviewer", Name: "独立复核人", Email: "f01-reviewer@example.invalid",
		Password: "x", Role: "compliance", Status: "active", CreatedAt: now, UpdatedAt: now,
	}).Error)
}

func seedF01Check(t *testing.T, db *gorm.DB, checkID string, inboxOpen bool) {
	seedF01CheckWithIntake(t, db, f01IntakeID, checkID, inboxOpen)
}

func seedF01CheckWithIntake(t *testing.T, db *gorm.DB, intakeID, checkID string, inboxOpen bool) {
	t.Helper()
	now := time.Now()
	require.NoError(t, db.Exec(
		"INSERT INTO case_intakes (id, created_by, metadata) VALUES (?, ?, ?) "+
			"ON CONFLICT(id) DO UPDATE SET metadata = excluded.metadata",
		intakeID, f01OwnerID, fmt.Sprintf(`{"conflict_check_id":%q}`, checkID)).Error)
	require.NoError(t, db.Create(&models.ConflictCheckRecord{
		CheckID:          checkID,
		ClientName:       "六项复测A2（虚构）客户甲公司",
		CaseName:         "六项复测A2（虚构）测试案",
		CheckStatus:      "COMPLETED",
		SearchParameters: models.JSON{"intakeId": intakeID},
		CheckResult:      models.JSON{"decision": map[string]interface{}{"coverageStatus": "COMPLETE"}},
		UserID:           f01OwnerID,
		CheckTime:        now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Error)
	require.NoError(t, db.Create(&models.ConflictReviewerAssignment{
		ID: uuid.NewString(), CheckID: checkID, ReviewerID: f01Reviewer, AssignedBy: f01Reviewer,
		Status: models.ConflictReviewerAssignmentActive, RecusalDeclared: true,
		IndependenceReason: "本机虚构QA回归：无管理关系", CreatedAt: now, UpdatedAt: now,
	}).Error)
	if inboxOpen {
		require.NoError(t, db.Create(&models.InboxItem{
			UserID: f01Reviewer, SourceType: "approval", SourceID: f01Reviewer,
			CompletionMode: "workflow", WorkflowTaskKey: "conflict_review:" + checkID + ":202",
			Title: "利益冲突独立复核待处理", Content: "有一项利益冲突检测等待你独立复核。",
			SourceRef: checkID, Priority: "high", IsCompleted: false,
		}).Error)
	}
}

func seedF01ReviewWorld(t *testing.T, db *gorm.DB, checkID string, inboxOpen bool) {
	t.Helper()
	seedF01Users(t, db)
	seedF01Check(t, db, checkID, inboxOpen)
}

func newF01ReviewService(db *gorm.DB) *conflictDetectionService {
	svc := NewConflictDetectionService(nil, nil, nil, nil, repositories.NewCaseRepository(db))
	return svc.(*conflictDetectionService)
}

func submitF01Review(t *testing.T, db *gorm.DB, checkID string) (*models.ConflictReview, error) {
	t.Helper()
	return newF01ReviewService(db).ReviewConflict(context.Background(), checkID, "no_conflict",
		"本机虚构QA F01回归：依据当前可见范围提交无冲突复核结论", f01Reviewer, "独立复核人", nil)
}

func loadF01Inbox(t *testing.T, db *gorm.DB, checkID string) *models.InboxItem {
	t.Helper()
	var item models.InboxItem
	require.NoError(t, db.Where("workflow_task_key = ?", "conflict_review:"+checkID+":202").
		Order("id ASC").First(&item).Error)
	return &item
}

// Success path: the real ReviewConflict transaction must close the
// reviewer's own inbox handoff with a completed_at timestamp.
func TestReviewConflictClosesReviewerInboxOnSuccess(t *testing.T) {
	db := newF01ReviewDB(t)
	seedF01ReviewWorld(t, db, f01CheckID, true)
	var frozenCheckResult string
	require.NoError(t, db.Raw("SELECT check_result FROM conflict_check_records WHERE check_id = ?", f01CheckID).Scan(&frozenCheckResult).Error)
	require.NotEmpty(t, frozenCheckResult, "夹具必须真实写入冻结检测结果")

	review, err := submitF01Review(t, db, f01CheckID)
	require.NoError(t, err)
	require.NotNil(t, review)
	require.Equal(t, "no_conflict", review.Decision)

	var count int64
	require.NoError(t, db.Model(&models.ConflictReview{}).Where("check_id = ?", f01CheckID).Count(&count).Error)
	require.Equal(t, int64(1), count)

	item := loadF01Inbox(t, db, f01CheckID)
	require.True(t, item.IsCompleted, "保存复核结论后核查人待办必须同事务关闭")
	require.NotNil(t, item.CompletedAt, "关闭的核查人待办必须写入 completed_at")

	var frozenCheckResultAfter string
	require.NoError(t, db.Raw("SELECT check_result FROM conflict_check_records WHERE check_id = ?", f01CheckID).Scan(&frozenCheckResultAfter).Error)
	require.Equal(t, frozenCheckResult, frozenCheckResultAfter, "ReviewConflict 不得改写冻结的检测结果 JSON")
}

// When the review insert fails, the whole transaction must roll back and
// the inbox handoff must stay open (no orphan close without a review).
func TestReviewConflictKeepsInboxOpenWhenReviewSaveFails(t *testing.T) {
	db := newF01ReviewDB(t)
	seedF01ReviewWorld(t, db, f01CheckID, true)
	removeF01FailHook := db.Callback().Create().Before("gorm:create").Register("f01_fail_review_insert", func(tx *gorm.DB) {
		if tx.Statement != nil && tx.Statement.Table == "conflict_reviews" {
			tx.AddError(errors.New("injected review insert failure"))
		}
	})
	_ = removeF01FailHook

	_, err := submitF01Review(t, db, f01CheckID)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&models.ConflictReview{}).Count(&count).Error)
	require.Equal(t, int64(0), count, "复核保存失败后不得留下复核记录")
	item := loadF01Inbox(t, db, f01CheckID)
	require.False(t, item.IsCompleted, "复核保存失败时核查人待办必须保持打开")
	require.Nil(t, item.CompletedAt)
}

// When the inbox close fails, the review insert must roll back too: the
// reviewer must not end up with a saved review plus a still-open todo.
func TestReviewConflictRollsBackReviewWhenInboxCloseFails(t *testing.T) {
	db := newF01ReviewDB(t)
	seedF01ReviewWorld(t, db, f01CheckID, true)
	removeF01CloseHook := db.Callback().Update().Before("gorm:update").Register("f01_fail_inbox_close", func(tx *gorm.DB) {
		if tx.Statement != nil && tx.Statement.Table == "inbox_items" {
			tx.AddError(errors.New("injected inbox close failure"))
		}
	})
	_ = removeF01CloseHook

	_, err := submitF01Review(t, db, f01CheckID)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&models.ConflictReview{}).Count(&count).Error)
	require.Equal(t, int64(0), count, "关闭待办失败时复核记录必须回滚")
	item := loadF01Inbox(t, db, f01CheckID)
	require.False(t, item.IsCompleted, "关闭失败时待办保持打开")
}

// Duplicate submission keeps the terminal state: REVIEW_ALREADY_EXISTS and
// the already-closed inbox must not be reopened.
func TestReviewConflictDuplicateKeepsInboxClosed(t *testing.T) {
	db := newF01ReviewDB(t)
	seedF01ReviewWorld(t, db, f01CheckID, true)

	_, err := submitF01Review(t, db, f01CheckID)
	require.NoError(t, err)
	first := loadF01Inbox(t, db, f01CheckID)
	require.True(t, first.IsCompleted)
	completedAt := *first.CompletedAt

	_, err = submitF01Review(t, db, f01CheckID)
	require.Error(t, err)
	var reviewerErr *ConflictReviewerError
	require.True(t, errors.As(err, &reviewerErr))
	require.Equal(t, "REVIEW_ALREADY_EXISTS", reviewerErr.Code)

	item := loadF01Inbox(t, db, f01CheckID)
	require.True(t, item.IsCompleted)
	require.NotNil(t, item.CompletedAt)
	require.True(t, item.CompletedAt.Equal(completedAt), "重复提交不得重新打开或改写已关闭待办")
}

// A re-run conflict check for the same intake produces a fresh todo; the
// old completed handoff stays terminal and is never reopened by the new
// check's review.
func TestReviewConflictNewCheckGetsFreshInboxOldStaysClosed(t *testing.T) {
	db := newF01ReviewDB(t)
	seedF01ReviewWorld(t, db, f01CheckID, true)
	_, err := submitF01Review(t, db, f01CheckID)
	require.NoError(t, err)
	oldItem := loadF01Inbox(t, db, f01CheckID)
	require.True(t, oldItem.IsCompleted)
	oldCompletedAt := *oldItem.CompletedAt

	nextCheckID := "CCT_F01_REVIEW_CLOSE_RERUN"
	seedF01Check(t, db, nextCheckID, true)
	nextKey := "conflict_review:" + nextCheckID + ":202"
	_, err = submitF01Review(t, db, nextCheckID)
	require.NoError(t, err)

	var openCount int64
	require.NoError(t, db.Model(&models.InboxItem{}).
		Where("workflow_task_key LIKE ? AND is_completed = ?", "conflict_review:%", false).
		Count(&openCount).Error)
	require.Equal(t, int64(0), openCount, "两轮复核完成后不得留有未关闭的核查人待办")

	reloaded := loadF01Inbox(t, db, f01CheckID)
	require.True(t, reloaded.IsCompleted, "旧已完成待办不得被重新打开")
	require.True(t, reloaded.CompletedAt.Equal(oldCompletedAt))

	var nextRows int64
	require.NoError(t, db.Model(&models.InboxItem{}).Where("workflow_task_key = ?", nextKey).Count(&nextRows).Error)
	require.Equal(t, int64(1), nextRows, "重检产生的新待办保持独立记录")
	newItem := loadF01Inbox(t, db, nextCheckID)
	require.True(t, newItem.IsCompleted)
	require.True(t, strings.HasPrefix(newItem.WorkflowTaskKey, "conflict_review:"+nextCheckID+":"))
}

// Closing the target check's todo must not leak into other checks: an
// unrelated open handoff for a different (never-reviewed) check keeps its
// original is_completed/completed_at state while the target todo closes.
func TestReviewConflictCloseScopeLimitedToTargetCheck(t *testing.T) {
	db := newF01ReviewDB(t)
	seedF01ReviewWorld(t, db, f01CheckID, true)
	otherCheckID := "CCT_F01_OTHER_CHECK"
	seedF01CheckWithIntake(t, db, "INT_F01_"+otherCheckID, otherCheckID, true)
	otherBefore := loadF01Inbox(t, db, otherCheckID)
	require.False(t, otherBefore.IsCompleted)
	require.Nil(t, otherBefore.CompletedAt)

	_, err := submitF01Review(t, db, f01CheckID)
	require.NoError(t, err)

	target := loadF01Inbox(t, db, f01CheckID)
	require.True(t, target.IsCompleted, "目标检测的核查人待办必须关闭")
	require.NotNil(t, target.CompletedAt)

	other := loadF01Inbox(t, db, otherCheckID)
	require.False(t, other.IsCompleted, "未复核的其他检测待办不得被顺带关闭")
	require.Nil(t, other.CompletedAt, "未复核的其他检测待办不得被写入完成时间")
	require.Equal(t, otherBefore.CreatedAt, other.CreatedAt, "其他检测待办行保持原样")
	var otherReviewCount int64
	require.NoError(t, db.Model(&models.ConflictReview{}).Where("check_id = ?", otherCheckID).Count(&otherReviewCount).Error)
	require.Equal(t, int64(0), otherReviewCount, "目标检测复核不得为其他检测生成复核记录")
}
