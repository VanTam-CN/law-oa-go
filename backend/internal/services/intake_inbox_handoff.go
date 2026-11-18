package services

// N4 (intake collaboration MVP): transactional approval <-> inbox handoff.
//
// Every workflow todo is opened and closed inside the same database
// transaction as the business write that requires it (plan sections
// C05-C07, C13): a failed handoff rolls the approval write back, and a
// failed approval write removes the handoff. Idempotency rests on the
// partial unique index uq_inbox_workflow_task_key over
// workflow_task_key (migration 000082): retries of the same business
// step reuse the same key and therefore cannot duplicate a task, while a
// new round produces a new key.

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// Inbox completion semantics for workflow-owned items (000082 comment):
// workflow items complete only through the business transition; only
// manual items may be completed by their owner directly.
const inboxCompletionModeWorkflow = "workflow"

// Task purposes embedded in workflow_task_key (position 2). The key
// layout is approval:<purpose>:<approvalID>:c<cycle>:<receiverUserID>;
// reviewer tasks use conflict_review:<checkID>:<reviewerID> instead.
const (
	inboxPurposeApprovalDecision  = "approval_decision"
	inboxPurposeApplicantRevision = "supplement_revision"
)

// buildApprovalTaskKey assembles the deterministic workflow_task_key for
// approval handoffs. cycle counts completed request_changes records so
// the initial submission is c0 and every later revision opens cN+1.
func buildApprovalTaskKey(purpose, approvalID string, cycle int, receiverUserID string) string {
	return fmt.Sprintf("approval:%s:%s:c%d:%s", purpose, approvalID, cycle, receiverUserID)
}

// approvalRevisionCycle counts request_changes decision records for the
// approval. Each return writes exactly one record (decision tx), so the
// count is a stable, monotonic cycle number computed inside the caller's
// transaction.
func approvalRevisionCycle(tx *gorm.DB, approvalID string) (int, error) {
	var n int64
	if err := tx.Table("approval_records").
		Where("approval_request_id = ? AND decision = ?", approvalID, models.ApprovalDecisionRequestChanges).
		Count(&n).Error; err != nil {
		return 0, fmt.Errorf("读取退回轮次失败：%w", err)
	}
	return int(n), nil
}

// resolveWorkflowInboxUser maps a business actor (approval user IDs are
// strings) to an active inbox owner. Unknown or deactivated receivers
// resolve to 0 and the handoff is skipped rather than guessed: the
// approval itself remains the source of truth until someone owns the
// step (C05: unassigned workflows produce no fabricated todo).
func resolveWorkflowInboxUser(tx *gorm.DB, userID string) uint {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return 0
	}
	var row struct {
		ID uint
	}
	// Approval actors store either numeric user IDs ("104") or usernames
	// ("u-h-admin"). Compare numeric input against the integer id column
	// only -- PostgreSQL rejects non-numeric literals in integer
	// comparisons -- then fall back to a username lookup for legacy rows.
	if n, parseErr := strconv.ParseUint(trimmed, 10, 64); parseErr == nil {
		if err := tx.Table("users").Select("id").
			Where("id = ? AND status = 'active' AND deleted_at IS NULL", n).
			Take(&row).Error; err == nil {
			return row.ID
		}
	}
	if err := tx.Table("users").Select("id").
		Where("username = ? AND status = 'active' AND deleted_at IS NULL", trimmed).
		Take(&row).Error; err != nil {
		return 0
	}
	return row.ID
}

// ensureWorkflowInboxItem opens (or refreshes) the workflow handoff item
// for receiverID. Behavior:
//   - receiverID == 0: skip silently (unassigned workflow stays
//     projection-only; C05 boundary).
//   - open item with the same key: refresh title/content/source and keep
//     it open (idempotent retry of the same business step).
//   - completed item with the same key: keep the terminal handoff record
//     (the business step already produced a visible handoff once; a
//     replayed transaction must not resurrect it -- replayed requests
//     return before reaching this helper anyway).
//   - no row: create it; a concurrent writer winning the partial unique
//     index is treated as already-open (C12-style race backstop).
func ensureWorkflowInboxItem(tx *gorm.DB, receiverID uint, key, title, content, sourceRef string) error {
	if receiverID == 0 {
		return nil
	}
	var existing struct {
		ID          uint
		IsCompleted bool
	}
	err := tx.Table("inbox_items").Select("id, is_completed").
		Where("workflow_task_key = ?", key).
		Take(&existing).Error
	if err == nil {
		if existing.IsCompleted {
			return nil
		}
		return tx.Model(&models.InboxItem{}).Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"title":      title,
				"content":    content,
				"source_ref": sourceRef,
			}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("读取协作待办失败：%w", err)
	}
	item := &models.InboxItem{
		UserID:          receiverID,
		SourceType:      "approval",
		SourceID:        0,
		SourceRef:       sourceRef,
		CompletionMode:  inboxCompletionModeWorkflow,
		WorkflowTaskKey: key,
		Title:           title,
		Content:         content,
		Priority:        "high",
	}
	if err := tx.Create(item).Error; err != nil {
		if isUniqueViolationErr(err) {
			// A concurrent transaction opened the same task first;
			// treat it as the authoritative handoff.
			return nil
		}
		return fmt.Errorf("写入协作待办失败：%w", err)
	}
	return nil
}

// closeApprovalWorkflowItems completes every open workflow item of the
// given purposes for one approval (prefix match on the key). Pattern
// closure -- instead of remembering row IDs across transactions -- keeps
// close idempotent, covers reassignment leftovers, and never resurrects
// already-completed records.
func closeApprovalWorkflowItems(tx *gorm.DB, approvalID string, purposes ...string) error {
	for _, purpose := range purposes {
		prefix := fmt.Sprintf("approval:%s:%s:", purpose, approvalID)
		if err := tx.Model(&models.InboxItem{}).
			Where("workflow_task_key LIKE ? AND is_completed = ?", prefix+"%", false).
			Updates(map[string]interface{}{
				"is_completed": true,
				"completed_at": time.Now(),
			}).Error; err != nil {
			return fmt.Errorf("关闭协作待办失败：%w", err)
		}
	}
	return nil
}

// openApprovalDecisionTask opens the current approver's decision handoff
// inside the caller's transaction. cycle anchors the task to the return
// cycle that put the approval back in review.
func (s *ApprovalService) openApprovalDecisionTask(tx *gorm.DB, approval *models.ApprovalRequest) error {
	cycle, err := approvalRevisionCycle(tx, approval.ID)
	if err != nil {
		return err
	}
	receiverID := resolveWorkflowInboxUser(tx, approval.CurrentApproverID)
	key := buildApprovalTaskKey(inboxPurposeApprovalDecision, approval.ID, cycle, approval.CurrentApproverID)
	title := "接案审批待处理"
	content := fmt.Sprintf("「%s」正在等待你的审批决定。", approval.Title)
	return ensureWorkflowInboxItem(tx, receiverID, key, title, content, approval.ID)
}

// openApplicantRevisionTask opens the applicant's revision handoff after
// a request_changes decision, anchored to the return cycle.
func (s *ApprovalService) openApplicantRevisionTask(tx *gorm.DB, approval *models.ApprovalRequest) error {
	cycle, err := approvalRevisionCycle(tx, approval.ID)
	if err != nil {
		return err
	}
	receiverID := resolveWorkflowInboxUser(tx, approval.ApplicantID)
	key := buildApprovalTaskKey(inboxPurposeApplicantRevision, approval.ID, cycle, approval.ApplicantID)
	title := "退回补充待处理"
	content := fmt.Sprintf("「%s」被退回补充，请逐项回复后重新提交。", approval.Title)
	return ensureWorkflowInboxItem(tx, receiverID, key, title, content, approval.ID)
}

// closeConflictReviewTasks closes one reviewer's open handoff for a
// conflict check (used when the appointment moves to another reviewer so
// the old reviewer cannot act on a revoked assignment; C06).
func closeConflictReviewTasks(tx *gorm.DB, checkID string, reviewerID uint) error {
	if err := tx.Model(&models.InboxItem{}).
		Where("workflow_task_key LIKE ? AND is_completed = ?",
			fmt.Sprintf("conflict_review:%s:%%", checkID), false).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("关闭复核协作待办失败：%w", err)
	}
	return nil
}
