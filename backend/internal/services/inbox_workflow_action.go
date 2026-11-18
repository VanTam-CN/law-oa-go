package services

// N5 (intake collaboration MVP): workflow action projection (plan §5.1).
//
// The inbox service owns the DTO; this resolver turns the handoff fields
// written by N4 (source_ref / completion_mode / workflow_task_key) into
// server-generated actions. The frontend only maps whitelisted action
// codes to internal routes and never receives a raw navigation URL.
// List rendering is batched (§5.4: no per-row query storms).

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

const (
	// Action codes (plan §5.1 whitelist).
	workflowActionCodeReviewConflict = "review_conflict"
	workflowActionCodeDecideApproval = "decide_approval"
	workflowActionCodeSupplement     = "supplement"
	workflowActionCodeOpenCaseTask   = "open_case_task"
	workflowActionCodeViewOnly       = "view_only"

	// Blocked markers for items whose object cannot be resolved or whose
	// owner is no longer the assigned actor (C05/C06).
	workflowBlockedLegacy = "legacy_unresolved"
	workflowBlockedOwner  = "owner_unauthorized"

	// ownerSourceMismatch keeps read/act ownership aligned: C06 forbids a
	// reassigned owner from acting through the old todo even though the
	// row still sits in their inbox.
	ownerSourceMismatch = "OWNER_SOURCE_MISMATCH"
)

// buildWorkflowActions projects one page of workflow-owned inbox rows.
// The map is keyed by inbox item ID; rows without a business action are
// absent. Every lookup failure degrades that single row to a blocked
// view_only action -- the user keeps their own todo and loses only the
// deep link, never gains access to a foreign object.
func (s *InboxService) buildWorkflowActions(actorID uint, db *gorm.DB, items []*models.InboxItem) map[uint]*WorkflowAction {
	result := make(map[uint]*WorkflowAction, len(items))
	approvalIDs := map[string]bool{}
	checkIDs := map[string]bool{}
	taskIDs := map[string]bool{}
	for _, item := range items {
		if item.CompletionMode != inboxCompletionModeWorkflow {
			continue // manual / legacy rows carry no business action
		}
		action := &WorkflowAction{CompletionMode: item.CompletionMode}
		result[item.ID] = action
		key := item.WorkflowTaskKey
		if key == "" || item.SourceRef == "" {
			setBlockedLegacy(action, "历史事项", "历史事项，关联对象待确认")
			continue
		}
		switch {
		case strings.HasPrefix(key, "conflict_review:"):
			checkIDs[item.SourceRef] = true
		case strings.HasPrefix(key, "approval:"):
			approvalIDs[item.SourceRef] = true
		case strings.HasPrefix(key, "case_task:"):
			// Stage C case-task todos resolve their display row in the same
			// batch; without this branch they would degrade to
			// legacy_unresolved on day one.
			taskIDs[item.SourceRef] = true
		default:
			setBlockedLegacy(action, "待办事项", "关联对象待确认")
		}
	}
	approvals := s.prefetchApprovals(db, approvalIDs)
	cycles := s.prefetchRevisionCycles(db, approvalIDs)
	reviewers := s.prefetchActiveReviewers(db, checkIDs)
	// Conflict-review rows hydrate their matter name through the same
	// server-side authorization gate used by case-name search; a restricted
	// todo keeps its generic label and stays searchable only by its stored
	// title/content.
	titleActorID := actorID
	if titleActorID == 0 && len(items) > 0 {
		// The single-item detail path has already authorized the row to its
		// owner; there is no batched actor there, so the owner is the acting
		// identity. Mixed owners would be a caller bug and stay unhydrated.
		owner := items[0].UserID
		singleOwner := owner != 0
		for _, item := range items {
			if item.UserID != owner {
				singleOwner = false
				break
			}
		}
		if singleOwner {
			titleActorID = owner
		}
	}
	conflictTitles := s.inboxRepo.ConflictReviewAuthorizedItemTitles(titleActorID, "")
	tasks := prefetchCaseTaskRows(db, taskIDs)
	caseAccess := s.prefetchCaseAccess(actorID, db, tasks, items)
	for _, item := range items {
		action, ok := result[item.ID]
		if !ok {
			continue
		}
		key := item.WorkflowTaskKey
		switch {
		case strings.HasPrefix(key, "conflict_review:"):
			fillConflictReviewAction(action, item, reviewers[item.SourceRef], conflictTitles[item.ID])
		case strings.HasPrefix(key, "approval:"):
			fillApprovalAction(action, item, approvals[item.SourceRef], cycles[item.SourceRef], db)
		case strings.HasPrefix(key, "case_task:"):
			fillCaseTaskAction(action, item, tasks[item.SourceRef], caseAccess[item.SourceRef])
		}
	}
	return result
}

// TestProjectWorkflowAction exposes the batched projection for a single
// row in regression tests; production paths always go through
// convertToResponses so the page-level batching stays intact.
func (s *InboxService) TestProjectWorkflowAction(actorID uint, item *models.InboxItem) *WorkflowAction {
	actions := s.buildWorkflowActions(actorID, s.inboxRepo.DB(), []*models.InboxItem{item})
	return actions[item.ID]
}

// ProjectWorkflowActions resolves a page of workflow todos through the same
// batched projection used by the inbox list. Callers outside the service
// (command center) get identical related_title / next_step semantics.
func (s *InboxService) ProjectWorkflowActions(actorID uint, items []*models.InboxItem) map[uint]*WorkflowAction {
	return s.buildWorkflowActions(actorID, s.inboxRepo.DB(), items)
}

func setBlockedLegacy(action *WorkflowAction, label, reason string) {
	action.ResourceType = "approval"
	action.ActionCode = workflowActionCodeViewOnly
	action.ActionLabel = label
	action.BlockedCode = workflowBlockedLegacy
	action.BlockedReason = reason
}

// caseTaskActionRow carries only the display fields the projection needs
// from one case_tasks row.
type caseTaskActionRow struct {
	ID      uint   `gorm:"column:id"`
	CaseID  uint   `gorm:"column:case_id"`
	Title   string `gorm:"column:title"`
	Status  string `gorm:"column:status"`
	OwnerID uint   `gorm:"column:assignee_id"`
}

// prefetchCaseTaskRows batch-reads case_tasks rows for case_task todos.
func prefetchCaseTaskRows(db *gorm.DB, ids map[string]bool) map[string]caseTaskActionRow {
	out := make(map[string]caseTaskActionRow, len(ids))
	if len(ids) == 0 {
		return out
	}
	idList := make([]uint, 0, len(ids))
	for raw := range ids {
		if id, err := strconv.ParseUint(raw, 10, 64); err == nil && id > 0 {
			idList = append(idList, uint(id))
		}
	}
	if len(idList) == 0 {
		return out
	}
	var rows []caseTaskActionRow
	_ = db.Table("case_tasks").Select("id, case_id, title, status, assignee_id").Where("id IN ?", idList).Scan(&rows).Error
	for _, r := range rows {
		out[fmt.Sprint(r.ID)] = r
	}
	return out
}

// fillCaseTaskAction projects a case_task todo. The inbox list already
// authorizes the row to its assignee; the projection adds the display
// fields and keeps the action on the case-detail task panel route family
// (no guessed URL).
// prefetchCaseAccess resolves, per case_task todo, whether the acting user
// may still open the owning case. Fail-closed by design: a missing case row,
// revoked team membership, or an enabled ethical wall without an explicit
// whitelist entry degrades the todo to a blocked action. Business-matter
// management roles keep their aggregate visibility.
func (s *InboxService) prefetchCaseAccess(actorID uint, db *gorm.DB, tasks map[string]caseTaskActionRow, items []*models.InboxItem) map[string]bool {
	out := make(map[string]bool, len(items))
	if actorID == 0 {
		return out
	}
	caseIDs := map[uint]bool{}
	for _, item := range items {
		if !strings.HasPrefix(item.WorkflowTaskKey, "case_task:") {
			continue
		}
		if task, ok := tasks[item.SourceRef]; ok && task.CaseID > 0 {
			caseIDs[task.CaseID] = true
		}
	}
	if len(caseIDs) == 0 {
		return out
	}
	management := s.isManagementActor(actorID)
	idList := make([]uint, 0, len(caseIDs))
	for id := range caseIDs {
		idList = append(idList, id)
	}
	visible := map[uint]bool{}
	// The batched projection is a read model, not a second grant
	// authority: it must apply the same account-eligibility rules as
	// CaseTaskService.caseTaskAssigneeAllowed. That gate requires the
	// actor's account to be live (deleted_at IS NULL, status='active') and
	// its CanReadCase branch denies technical administrators outright, so
	// a whitelisted collaborator who became a technical admin or was
	// deactivated keeps their grant row but loses task eligibility. The
	// open projection must not keep publishing case task ids to that
	// actor. Any lookup failure stays fail-closed.
	var eligibleUser struct {
		ID   uint
		Role string
	}
	_ = db.Table("users").
		Select("id, role").
		Where("id = ? AND deleted_at IS NULL AND status = 'active'", actorID).
		Take(&eligibleUser).Error
	eligible := map[uint]bool{}
	if eligibleUser.ID != 0 && !IsTechnicalAdminRole(eligibleUser.Role) {
		eligible[eligibleUser.ID] = true
	}
	if !management {
		var rows []struct {
			CaseID uint `gorm:"column:id"`
		}
		// The same grant model as AuthorizationService.CanReadCase and the
		// legal assignee check: owner OR creator OR an explicit ethical-wall
		// whitelist row (a deliberate per-matter collaboration grant), with
		// the wall requiring that same row when enabled. Management keeps
		// its separate branch; revoked/missing/deleted cases stay fail-closed.
		_ = db.Table("cases").
			Select("cases.id").
			Where("cases.id IN ? AND cases.deleted_at IS NULL", idList).
			Where("(cases.lawyer_id = ? OR cases.created_by = ? OR EXISTS ("+
				"SELECT 1 FROM case_ethical_wall_whitelist collaborator_grant "+
				"WHERE collaborator_grant.case_id = cases.id AND collaborator_grant.user_id = ?))",
				actorID, fmt.Sprint(actorID), actorID).
			Where(`(cases.ethical_wall_enabled = ? OR EXISTS (
				SELECT 1 FROM case_ethical_wall_whitelist wall_access
				WHERE wall_access.case_id = cases.id AND wall_access.user_id = ?
			))`, false, actorID).
			Scan(&rows).Error
		for _, r := range rows {
			visible[r.CaseID] = true
		}
	} else {
		for _, id := range idList {
			visible[id] = true
		}
	}
	for _, item := range items {
		if !strings.HasPrefix(item.WorkflowTaskKey, "case_task:") {
			continue
		}
		task, ok := tasks[item.SourceRef]
		out[item.SourceRef] = ok && task.CaseID > 0 && eligible[actorID] && visible[task.CaseID]
	}
	return out
}

// isManagementActor resolves the role through the users table because the
// batched projection runs outside a gin context. Any lookup failure is
// treated as non-management (fail-closed).
func (s *InboxService) isManagementActor(actorID uint) bool {
	if actorID == 0 || s.userRepo == nil {
		return false
	}
	user, err := s.userRepo.FindByID(context.Background(), actorID)
	if err != nil || user == nil || user.DeletedAt.Valid {
		return false
	}
	return IsBusinessMatterManagementRole(user.Role)
}

func fillCaseTaskAction(action *WorkflowAction, item *models.InboxItem, row caseTaskActionRow, caseAccessible bool) {
	if row.ID == 0 {
		setBlockedLegacy(action, "任务不可用", "关联任务不可用，无法直达办理")
		return
	}
	action.ResourceType = "case_task"
	action.ResourceID = item.SourceRef
	action.RelatedTitle = row.Title
	if !caseAccessible {
		action.ActionCode = workflowActionCodeViewOnly
		action.ActionLabel = "暂无权访问"
		action.BlockedCode = workflowBlockedOwner
		action.BlockedReason = "当前账号已无该案件访问权限，任务入口已停用"
		action.NextStep = "如需继续处理，请联系主办律师恢复案件协作权限"
		return
	}
	// The case id is exposed only AFTER the access check passed: a blocked
	// row must not leak which matter the revoked task belongs to.
	if row.CaseID > 0 {
		action.RelatedCaseID = fmt.Sprint(row.CaseID)
	}
	if row.Status == "completed" {
		action.ActionCode = workflowActionCodeViewOnly
		action.ActionLabel = "查看任务"
		action.NextStep = "任务已完成，可打开案件查看完成记录"
		return
	}
	// Authorized and still open: the action deep-links to the OWNING CASE
	// via related_case_id; resource_id stays the task id and is never
	// reinterpreted as a case number (review note 2026-09-30).
	action.ActionCode = workflowActionCodeOpenCaseTask
	action.ActionLabel = "去办理案件任务"
	action.NextStep = "打开所属案件任务面板核对进度，完成后在任务面板记录完成"
}

// approvalActionRow carries the approval fields the projection needs.
type approvalActionRow struct {
	ID         string `gorm:"column:id"`
	Status     string `gorm:"column:status"`
	ApproverID string `gorm:"column:current_approver_id"`
	Metadata   string `gorm:"column:metadata"`
	Title      string `gorm:"column:title"`
}

func (s *InboxService) prefetchApprovals(db *gorm.DB, ids map[string]bool) map[string]approvalActionRow {
	out := make(map[string]approvalActionRow, len(ids))
	if len(ids) == 0 {
		return out
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	var rows []approvalActionRow
	_ = db.Model(&models.ApprovalRequest{}).
		Select("id, status, current_approver_id, applicant_id, metadata, title").
		Where("id IN ?", idList).
		Scan(&rows).Error
	for _, r := range rows {
		out[r.ID] = r
	}
	return out
}

func (s *InboxService) prefetchRevisionCycles(db *gorm.DB, ids map[string]bool) map[string]int {
	out := make(map[string]int, len(ids))
	if len(ids) == 0 {
		return out
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	var rows []struct {
		ApprovalRequestID string
		Cnt               int
	}
	_ = db.Table("approval_records").
		Select("approval_request_id, COUNT(*) AS cnt").
		Where("approval_request_id IN ? AND decision = ?", idList, models.ApprovalDecisionRequestChanges).
		Group("approval_request_id").
		Scan(&rows).Error
	for _, r := range rows {
		out[r.ApprovalRequestID] = r.Cnt
	}
	return out
}

func (s *InboxService) prefetchActiveReviewers(db *gorm.DB, ids map[string]bool) map[string]uint {
	out := make(map[string]uint, len(ids))
	if len(ids) == 0 {
		return out
	}
	idList := make([]string, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	var rows []struct {
		CheckID    string
		ReviewerID uint
	}
	// Latest active assignment per check wins; the two-key window keeps
	// it to one query for the whole page.
	_ = db.Table("conflict_reviewer_assignments").
		Select("check_id, reviewer_id").
		Where("check_id IN ? AND status = ?", idList, models.ConflictReviewerAssignmentActive).
		Order("created_at DESC").
		Scan(&rows).Error
	for _, r := range rows {
		if _, seen := out[r.CheckID]; !seen {
			out[r.CheckID] = r.ReviewerID
		}
	}
	return out
}

func fillConflictReviewAction(action *WorkflowAction, item *models.InboxItem, activeReviewer uint, caseTitle string) {
	action.ResourceType = "conflict"
	action.ResourceID = item.SourceRef
	action.WorkflowStatus = "review_pending"
	if activeReviewer == 0 {
		// Appointment gone (revoked/reassigned without a replacement):
		// the todo keeps informing the old owner but grants no action.
		action.ActionCode = workflowActionCodeViewOnly
		action.ActionLabel = "复核安排已变更"
		action.BlockedCode = workflowBlockedOwner
		action.BlockedReason = "复核安排已变更，等待重新指派"
		action.NextStep = "等待合规负责人重新指派复核人后再办理"
		return
	}
	if activeReviewer != item.UserID {
		action.AssignedTo = "其他复核人"
		action.BlockedCode = ownerSourceMismatch
		action.BlockedReason = "该复核已转交其他复核人"
	}
	// The title is the authorized matter name resolved by the repository
	// gate; an empty value keeps the todo generic instead of leaking a
	// restricted context.
	action.RelatedTitle = caseTitle
	if caseTitle != "" {
		action.AssignedTo = "独立核查人"
	}
	action.ActionCode = workflowActionCodeReviewConflict
	action.ActionLabel = "去复核"
	action.NextStep = "打开冲突复核页，给出通过或退回结论"
}

func fillApprovalAction(action *WorkflowAction, item *models.InboxItem, row approvalActionRow, cycle int, db *gorm.DB) {
	action.ResourceType = "approval"
	action.ResourceID = item.SourceRef
	if row.ApproverID == "" && row.Status == "" {
		// Approval missing: keep the todo, block the action.
		setBlockedLegacy(action, "审批不可用", "审批记录不可用，无法直达办理")
		action.NextStep = "联系管理员确认该审批记录状态"
		return
	}
	action.WorkflowStatus = row.Status
	// Applicant revision todos belong to the applicant, not the approver:
	// decide ownership before the approver comparison so a transfer cannot
	// downgrade the applicant's supplement todo to view_only (defect 7).
	if strings.Contains(item.WorkflowTaskKey, ":"+inboxPurposeApplicantRevision+":") {
		action.ActionCode = workflowActionCodeSupplement
		action.ActionLabel = "去补充"
		action.ResourceType = "intake"
		action.ResourceID = intakeIDFromApproval(row.Metadata)
		action.RelatedTitle = row.Title
		// The revision gate also needs the approval UUID: without it the
		// workbench would open a blank new-intake form instead of the
		// round (the live C05 defect). item.SourceRef is the approval ID.
		action.RelatedApprovalID = item.SourceRef
		action.AssignedTo = "申请人"
		action.NextStep = "逐项回复退回意见后重新提交，等待审批复核"
		if action.ResourceID == "" {
			// §6.2: the navigation contract requires a validated intake;
			// without the binding the applicant gets a blocked view action
			// instead of a guessed /case/create URL.
			setBlockedLegacy(action, "补充对象待确认", "无法确认关联接案，请从审批详情处理")
			action.ResourceID = item.SourceRef
			// Fail closed: a blocked supplement exposes no approval link
			// either, so the projection never hands out a navigation
			// anchor for an action that cannot route.
			action.RelatedApprovalID = ""
			action.NextStep = "从审批详情打开本轮退回，确认关联接案后补充"
		}
		// Action version anchors to the revision cycle (request_changes count),
		// the same monotonic marker the handoff keys use.
		action.ActionVersion = fmt.Sprintf("approval:%s:c%d", item.SourceRef, cycle)
		return
	}
	if approverUserID := resolveWorkflowInboxUser(db, row.ApproverID); approverUserID != 0 && approverUserID == item.UserID {
		action.ActionCode = workflowActionCodeDecideApproval
		action.ActionLabel = "去审批"
		action.RelatedTitle = row.Title
		action.NextStep = "打开审批详情作出通过或退回决定"
	} else {
		action.AssignedTo = "其他审批人"
		action.BlockedCode = ownerSourceMismatch
		action.BlockedReason = "该审批已转交其他审批人"
		action.ActionCode = workflowActionCodeViewOnly
		action.ActionLabel = "查看审批"
		action.RelatedTitle = row.Title
		action.NextStep = "审批已转交，等待新审批人处理"
	}
	// Action version anchors to the revision cycle (request_changes count),
	// the same monotonic marker the handoff keys use.
	action.ActionVersion = fmt.Sprintf("approval:%s:c%d", item.SourceRef, cycle)
}

// intakeIDFromApproval reads the intake_id written by the submission
// contract into approval metadata (plan §6.2 navigation contract).
func intakeIDFromApproval(metadata string) string {
	var m map[string]interface{}
	if json.Unmarshal([]byte(metadata), &m) != nil {
		return ""
	}
	id, _ := m["intake_id"].(string)
	return strings.TrimSpace(id)
}
