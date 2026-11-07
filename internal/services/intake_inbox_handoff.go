package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

func requireInboxTransaction(tx *gorm.DB) error {
	if tx == nil || tx.Statement == nil {
		return fmt.Errorf("待办交接缺少业务事务")
	}
	if _, ok := tx.Statement.ConnPool.(gorm.TxCommitter); !ok {
		return fmt.Errorf("待办交接必须使用调用方的数据库事务")
	}
	return nil
}

func inboxDigest(parts ...string) string {
	raw, _ := json.Marshal(parts)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func conflictAssignmentGeneration(assignments []models.ConflictReviewerAssignment) string {
	parts := make([]string, 0, len(assignments))
	for _, a := range assignments {
		raw, _ := json.Marshal(struct {
			ID, Status                            string
			UpdatedAt                             time.Time
			EffectiveFrom, EffectiveTo, RevokedAt *time.Time
		}{a.ID, a.Status, a.UpdatedAt, a.EffectiveFrom, a.EffectiveTo, a.RevokedAt})
		parts = append(parts, string(raw))
	}
	sort.Strings(parts)
	return inboxDigest(parts...)
}

func approvalInboxVersion(a *models.ApprovalRequest) string {
	round := ""
	if a.ActiveSubmissionSnapshotID != nil {
		round = strings.TrimSpace(*a.ActiveSubmissionSnapshotID)
	}
	if round == "" {
		round = a.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	// UpdatedAt is the durable action version, also distinguishing A -> B -> A
	// reassignment within one snapshot/node. Reconciliation itself never changes it.
	return inboxDigest(a.ID, round, a.UpdatedAt.UTC().Format(time.RFC3339Nano), a.Status, a.CurrentStage, a.CurrentApproverID)
}

func approvalInboxResultVersion(a *models.ApprovalRequest) string {
	round := ""
	if a.ActiveSubmissionSnapshotID != nil {
		round = strings.TrimSpace(*a.ActiveSubmissionSnapshotID)
	}
	if round == "" {
		if a.SubmissionDate != nil {
			round = a.SubmissionDate.UTC().Format(time.RFC3339Nano)
		} else {
			round = a.CreatedAt.UTC().Format(time.RFC3339Nano)
		}
	}
	// Case linkage/progress is a projection of this outcome, not a second
	// approval outcome. It must not change the result reminder's identity.
	return inboxDigest(a.ID, round, a.Status, "result")
}

func inboxUserID(s string) uint {
	n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0
	}
	return uint(n)
}

func approvalInboxMetadata(a *models.ApprovalRequest) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	if strings.TrimSpace(a.Metadata) != "" {
		if err := json.Unmarshal([]byte(a.Metadata), &m); err != nil {
			return nil, fmt.Errorf("待办交接无法解析审批上下文: %w", err)
		}
	}
	return m, nil
}

func inboxText(m map[string]interface{}, key string) string {
	s, _ := m[key].(string)
	return strings.TrimSpace(s)
}

func newWorkflowInbox(group, version, sourceType, sourceRef, purpose string, userID uint, title string, completed bool) *models.InboxItem {
	item := &models.InboxItem{UserID: userID, SourceType: sourceType, SourceRef: sourceRef,
		Title: title, Priority: "medium", CompletionMode: "workflow", WorkflowGroup: group,
		WorkflowPurpose: purpose, WorkflowVersion: version,
		WorkflowTaskKey: "inbox:" + inboxDigest(group, version, strconv.FormatUint(uint64(userID), 10), purpose), IsCompleted: completed}
	if completed {
		now := time.Now()
		item.CompletedAt = &now
		item.ClosedReason = "completed"
	}
	return item
}

// ReconcileApprovalInboxInTx runs after the caller locks and saves the approval.
// All errors must be returned from that outer transaction. No notifications or
// network calls occur here. The caller owns snapshot validation and lock order.
func ReconcileApprovalInboxInTx(tx *gorm.DB, approval *models.ApprovalRequest) error {
	if err := requireInboxTransaction(tx); err != nil {
		return err
	}
	if approval == nil || strings.TrimSpace(approval.ID) == "" {
		return fmt.Errorf("待办交接缺少审批对象")
	}
	// Read the durable timestamp (PostgreSQL stores microseconds, while the
	// object just passed to Save may still carry nanoseconds). Also prevents a
	// stale caller object from reviving tasks for a superseded state.
	var stored models.ApprovalRequest
	if err := tx.Where("id = ?", approval.ID).First(&stored).Error; err != nil {
		return err
	}
	approval = &stored
	metadata, err := approvalInboxMetadata(approval)
	if err != nil {
		return err
	}
	group, version := "approval:"+approval.ID, approvalInboxVersion(approval)
	applicantID := inboxUserID(approval.ApplicantID)
	if applicantID == 0 {
		return fmt.Errorf("待办交接缺少有效申请人")
	}
	var desired []*models.InboxItem
	closeReason := "completed"
	switch approval.Status {
	case models.ApprovalStatusSubmitted, models.ApprovalStatusUnderReview, models.ApprovalStatusResubmitted:
		approverID := inboxUserID(approval.CurrentApproverID)
		if approverID == 0 || approverID == applicantID {
			return fmt.Errorf("待办交接缺少独立审批人")
		}
		desired = append(desired,
			newWorkflowInbox(group, version, "approval", approval.ID, "decide_approval", approverID, "接案审批待处理", false),
			newWorkflowInbox(group, version, "approval", approval.ID, "waiting", applicantID, "等待审批处理", false))
		desired[0].DueDate = approval.TimeoutAt
		closeReason = "transferred"
	case models.ApprovalStatusNeedsRevision:
		if intakeID := inboxText(metadata, "intake_id"); intakeID != "" {
			var intake struct{ CreatedBy string }
			if err := tx.Table("case_intakes").Select("created_by").Where("id = ?", intakeID).Take(&intake).Error; err != nil {
				return err
			}
			if inboxUserID(intake.CreatedBy) != applicantID {
				return fmt.Errorf("补充待办的接案与审批申请人不一致")
			}
			desired = append(desired, newWorkflowInbox(group, version, "intake", intakeID, "supplement", applicantID, "接案申请待补充", false))
		} else {
			desired = append(desired, newWorkflowInbox(group, version, "approval", approval.ID, "revision_unlinked", applicantID, "申请待补充，接案关联待确认", false))
		}
	case models.ApprovalStatusApproved, models.ApprovalStatusRejected, models.ApprovalStatusCancelled, models.ApprovalStatusExpired:
		desired = append(desired, newWorkflowInbox(group, approvalInboxResultVersion(approval), "approval", approval.ID, "result", applicantID, "审批结果已更新", true))
	case models.ApprovalStatusDraft:
		// A draft has no actionable approval task.
	default:
		return fmt.Errorf("不支持的审批待办状态 %q", approval.Status)
	}
	return reconcileWorkflowInbox(tx, group, desired, closeReason)
}

// ReconcileConflictInboxInTx follows the same locked check row as assignment
// and review. A role alone never becomes a review assignment.
func ReconcileConflictInboxInTx(tx *gorm.DB, checkID string) error {
	if err := requireInboxTransaction(tx); err != nil {
		return err
	}
	var record models.ConflictCheckRecord
	if err := tx.Where("check_id = ?", strings.TrimSpace(checkID)).First(&record).Error; err != nil {
		return err
	}
	assignment, err := GetActiveConflictReviewerAssignment(tx.Statement.Context, tx, record.CheckID)
	if err != nil {
		return err
	}
	var assignments []models.ConflictReviewerAssignment
	if err := tx.Where("check_id = ?", record.CheckID).Find(&assignments).Error; err != nil {
		return err
	}
	var reviewCount int64
	if err := tx.Model(&models.ConflictReview{}).Where("check_id = ?", record.CheckID).Count(&reviewCount).Error; err != nil {
		return err
	}
	group := "conflict:" + record.CheckID
	assignmentID := "unassigned"
	if assignment != nil {
		assignmentID = assignment.ID
	}
	version := inboxDigest(record.CheckID, assignmentID, record.CheckStatus, strconv.FormatInt(reviewCount, 10), conflictAssignmentGeneration(assignments))
	var desired []*models.InboxItem
	closeReason := "transferred"
	if reviewCount > 0 {
		closeReason = "completed"
		resultVersion := inboxDigest(record.CheckID, "reviewed", strconv.FormatInt(reviewCount, 10))
		desired = append(desired, newWorkflowInbox(group, resultVersion, "conflict", record.CheckID, "result", record.UserID, "独立复核已完成", true))
	} else if strings.EqualFold(record.CheckStatus, "COMPLETED") {
		// Reject stale facts before creating another actionable task. Stale
		// evidence closes the old queue; it must never revive a previous check.
		freshErr := EnsureConflictRecordMatchesCurrentIntakeFacts(tx, conflictRecordIntakeID(record.SearchParameters), record.CheckID, record.CheckTime)
		if freshErr != nil {
			if _, ok := freshErr.(*ConflictReviewerError); !ok {
				return freshErr
			}
			return reconcileWorkflowInbox(tx, group, nil, "superseded")
		}
		desired = append(desired, newWorkflowInbox(group, version, "conflict", record.CheckID, "waiting", record.UserID, "等待独立复核", false))
		if assignment != nil {
			if err := ValidateConflictReviewer(tx.Statement.Context, tx, record.CheckID, 0, assignment.ReviewerID, ""); err != nil {
				return err
			}
			item := newWorkflowInbox(group, version, "conflict", record.CheckID, "review_conflict", assignment.ReviewerID, "正式检测待独立复核", false)
			item.DueDate = assignment.SLADueAt
			desired = append(desired, item)
		}
	}
	return reconcileWorkflowInbox(tx, group, desired, closeReason)
}

func reconcileWorkflowInbox(tx *gorm.DB, group string, desired []*models.InboxItem, closeReason string) error {
	keys := make([]string, 0, len(desired))
	for _, item := range desired {
		if item.UserID == 0 || item.SourceRef == "" {
			return fmt.Errorf("待办交接缺少真实来源或接收人")
		}
		keys = append(keys, item.WorkflowTaskKey)
	}
	old := tx.Model(&models.InboxItem{}).Where("workflow_group = ? AND completion_mode = ? AND is_completed = ?", group, "workflow", false)
	if len(keys) > 0 {
		old = old.Where("workflow_task_key NOT IN ?", keys)
	}
	if err := old.Updates(map[string]interface{}{"is_completed": true, "completed_at": time.Now(), "closed_reason": closeReason}).Error; err != nil {
		return fmt.Errorf("关闭旧待办失败: %w", err)
	}
	for _, item := range desired {
		// Never resurrect a previously closed/archived task or reset read/snooze
		// state on retry. A new business generation receives a different key.
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "workflow_task_key"}}, DoNothing: true}).Create(item).Error; err != nil {
			return fmt.Errorf("创建交接待办失败: %w", err)
		}
	}
	return nil
}
