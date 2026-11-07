package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// WorkflowAction is a server projection, never a capability or an arbitrary
// URL. The destination must reload the object and enforce its write gates.
type WorkflowAction struct {
	ResourceType   string            `json:"resource_type,omitempty"`
	ResourceID     string            `json:"resource_id,omitempty"`
	ApprovalID     string            `json:"approval_id,omitempty"`
	ActionCode     string            `json:"action_code"`
	ActionLabel    string            `json:"action_label"`
	WorkflowStatus string            `json:"workflow_status"`
	AssignedTo     *WorkflowAssignee `json:"assigned_to"`
	BlockedCode    string            `json:"blocked_code,omitempty"`
	BlockedReason  string            `json:"blocked_reason,omitempty"`
	CompletionMode string            `json:"completion_mode"`
	ActionVersion  string            `json:"action_version,omitempty"`
}

type WorkflowAssignee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type inboxIntake struct {
	ID        string
	CreatedBy string
	ClientID  uint
	Status    string
	Metadata  models.JSON `gorm:"type:json"`
}

type inboxProjection struct {
	actor                 models.User
	approvals             map[string]models.ApprovalRequest
	revisions             map[string]inboxRevisionScope
	clients               map[uint]bool
	history               map[string]bool
	checks                map[string]models.ConflictCheckRecord
	intakes               map[string]inboxIntake
	cases                 map[uint]models.Case
	assignments           map[string]models.ConflictReviewerAssignment
	assignmentGenerations map[string]string
	reviews               map[string]models.ConflictReview
	users                 map[uint]models.User
	walls                 map[uint]bool
	now                   time.Time
}

type inboxRevisionScope struct {
	clients []uint
	cases   []uint
	checks  []string
}

// Collect the same frozen object references as AuthorizeApprovalSnapshotContext,
// without performing per-approval queries. These are separate from the current
// editable intake while needs_revision is active.
func inboxFrozenRevisionScope(snapshot map[string]interface{}) (inboxRevisionScope, bool) {
	var scope inboxRevisionScope
	addID := func(ids *[]uint, value interface{}) bool {
		text := submissionText(value)
		if text == "" || text == "0" {
			return true
		}
		id := inboxUserID(text)
		if id == 0 {
			return false
		}
		*ids = append(*ids, id)
		return true
	}
	metadata := submissionMap(snapshot["metadata"])
	for _, object := range []map[string]interface{}{snapshot, metadata, submissionMap(snapshot["conflict_record"]), submissionMap(metadata["conflict_record"]), submissionMap(snapshot["case_creation_config"]), submissionMap(metadata["case_creation_config"])} {
		if !addID(&scope.clients, object["client_id"]) || !addID(&scope.clients, submissionMap(object["client"])["id"]) {
			return scope, false
		}
		for _, source := range []map[string]interface{}{object, submissionMap(object["search_parameters"])} {
			for _, key := range []string{"subjectCaseId", "subject_case_id", "caseId", "case_id"} {
				if !addID(&scope.cases, source[key]) {
					return scope, false
				}
			}
		}
		for _, key := range []string{"conflict_check_id", "conflict_task_id", "check_id"} {
			if id := submissionText(object[key]); id != "" {
				scope.checks = append(scope.checks, id)
			}
		}
	}
	return scope, true
}

// inboxLoad chunks by parameter count, not by individual object. No per-row
// authorization queries are permitted in the projection methods below.
func inboxLoad[T any, K comparable](db *gorm.DB, column string, ids []K) ([]T, error) {
	seen := map[K]bool{}
	unique := make([]K, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			unique = append(unique, id)
			seen[id] = true
		}
	}
	var result []T
	for start := 0; start < len(unique); start += 400 {
		end := start + 400
		if end > len(unique) {
			end = len(unique)
		}
		var rows []T
		if err := db.Session(&gorm.Session{}).Where(column+" IN ?", unique[start:end]).Find(&rows).Error; err != nil {
			return nil, err
		}
		result = append(result, rows...)
	}
	return result, nil
}

// inboxSubmissionProjection mirrors the active-pointer binding checks in
// ResolveApprovalSubmission using an already batch-loaded row. Both object
// authorization and action rendering consume this copy; historical metadata
// must never be rewritten, nor used as a fallback for an invalid active pointer.
func inboxSubmissionProjection(a models.ApprovalRequest, row models.ApprovalSnapshot) (models.ApprovalRequest, bool) {
	if a.ActiveSubmissionSnapshotID == nil {
		return a, true
	}
	if strings.TrimSpace(*a.ActiveSubmissionSnapshotID) == "" || row.ID != *a.ActiveSubmissionSnapshotID || row.ApprovalRequestID != a.ID || row.SnapshotType != models.IntakeSubmissionSnapshotType {
		return a, false
	}
	original, err := submissionMetadata(a.Metadata)
	if err != nil {
		return a, false
	}
	snapshot := submissionMap(row.SnapshotData)
	frozen := submissionMap(snapshot["metadata"])
	intakeID := submissionText(frozen["intake_id"])
	checkID := submissionText(frozen["conflict_check_id"])
	if snapshot == nil || frozen == nil || submissionText(snapshot["snapshot_id"]) != row.ID || submissionText(snapshot["approval_request_id"]) != a.ID || row.SourceVersion < 1 || submissionText(snapshot["source_version"]) != strconv.Itoa(row.SourceVersion) || intakeID == "" || submissionText(snapshot["intake_id"]) != intakeID || checkID == "" || submissionText(snapshot["conflict_check_id"]) != checkID {
		return a, false
	}
	if oldIntakeID := submissionText(original["intake_id"]); oldIntakeID != "" && oldIntakeID != intakeID {
		return a, false
	}
	// Retain the source-ID consistency gate for this round. The old check in
	// original metadata belongs to an earlier round and is not authoritative.
	if taskID := submissionText(frozen["conflict_task_id"]); taskID != "" && taskID != checkID {
		return a, false
	}
	raw, err := json.Marshal(frozen)
	if err != nil {
		return a, false
	}
	a.Metadata = string(raw)
	projectSubmissionConflict(&a, snapshot)
	return a, true
}

func projectWorkflowInbox(ctx context.Context, db *gorm.DB, userID uint, items []*models.InboxItem) (map[uint]*WorkflowAction, error) {
	result := map[uint]*WorkflowAction{}
	var approvalIDs, checkIDs, intakeIDs []string
	var caseIDs, clientIDs []uint
	workflowCount := 0
	for _, item := range items {
		if item.CompletionMode != "workflow" {
			continue
		}
		workflowCount++
		if strings.HasPrefix(item.WorkflowGroup, "approval:") {
			approvalIDs = append(approvalIDs, strings.TrimPrefix(item.WorkflowGroup, "approval:"))
		}
		switch item.SourceType {
		case "approval":
			approvalIDs = append(approvalIDs, item.SourceRef)
		case "conflict":
			checkIDs = append(checkIDs, item.SourceRef)
		case "intake":
			intakeIDs = append(intakeIDs, item.SourceRef)
		case "case":
			caseIDs = append(caseIDs, inboxUserID(item.SourceRef))
		}
	}
	if workflowCount == 0 {
		return result, nil
	}
	if db == nil {
		return nil, fmt.Errorf("待办对象权限服务未初始化")
	}
	db = db.WithContext(ctx)
	p := &inboxProjection{approvals: map[string]models.ApprovalRequest{}, revisions: map[string]inboxRevisionScope{}, clients: map[uint]bool{}, history: map[string]bool{}, checks: map[string]models.ConflictCheckRecord{}, intakes: map[string]inboxIntake{}, cases: map[uint]models.Case{}, assignments: map[string]models.ConflictReviewerAssignment{}, assignmentGenerations: map[string]string{}, reviews: map[string]models.ConflictReview{}, users: map[uint]models.User{}, walls: map[uint]bool{}, now: time.Now()}
	if err := db.Select("id, name, role, status, manager_id").Where("id = ?", userID).First(&p.actor).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return result, nil
		}
		return nil, err
	}
	if !strings.EqualFold(p.actor.Status, "active") || IsTechnicalAdminRole(p.actor.Role) {
		return result, nil
	}
	approvals, err := inboxLoad[models.ApprovalRequest](db, "id", approvalIDs)
	if err != nil {
		return nil, err
	}
	var snapshotIDs []string
	for _, a := range approvals {
		if a.ActiveSubmissionSnapshotID != nil {
			snapshotIDs = append(snapshotIDs, *a.ActiveSubmissionSnapshotID)
		}
	}
	snapshots, err := inboxLoad[models.ApprovalSnapshot](db, "id", snapshotIDs)
	if err != nil {
		return nil, err
	}
	bySnapshotID := make(map[string]models.ApprovalSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		bySnapshotID[snapshot.ID] = snapshot
	}
	userIDs := []uint{userID}
	var historyIDs []string
	for _, a := range approvals {
		var snapshot models.ApprovalSnapshot
		if a.ActiveSubmissionSnapshotID != nil {
			snapshot = bySnapshotID[*a.ActiveSubmissionSnapshotID]
		}
		var valid bool
		a, valid = inboxSubmissionProjection(a, snapshot)
		if !valid {
			continue
		}
		if a.ActiveSubmissionSnapshotID != nil && a.Status == models.ApprovalStatusNeedsRevision {
			scope, valid := inboxFrozenRevisionScope(submissionMap(snapshot.SnapshotData))
			if !valid {
				continue
			}
			p.revisions[a.ID] = scope
			clientIDs = append(clientIDs, scope.clients...)
			caseIDs = append(caseIDs, scope.cases...)
			checkIDs = append(checkIDs, scope.checks...)
		}
		p.approvals[a.ID] = a
		if inboxUserID(a.ApplicantID) != userID && inboxUserID(a.CurrentApproverID) != userID {
			historyIDs = append(historyIDs, a.ID)
		}
		metadata, err := approvalInboxMetadata(&a)
		if err != nil {
			return nil, err
		}
		if id := inboxApprovalCheckID(&a, metadata); id != "" {
			checkIDs = append(checkIDs, id)
		}
		if id := inboxText(metadata, "intake_id"); id != "" {
			intakeIDs = append(intakeIDs, id)
		}
		userIDs = append(userIDs, inboxUserID(a.CurrentApproverID), inboxUserID(a.ApplicantID))
		if id := inboxUserID(a.CreatedCaseID); id > 0 {
			caseIDs = append(caseIDs, id)
		}
	}
	if len(historyIDs) > 0 {
		records, err := inboxLoad[models.ApprovalRecord](db.Select("approval_request_id").Where("approver_id = ?", strconv.FormatUint(uint64(userID), 10)), "approval_request_id", historyIDs)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			p.history[record.ApprovalRequestID] = true
		}
	}
	checks, err := inboxLoad[models.ConflictCheckRecord](db, "check_id", checkIDs)
	if err != nil {
		return nil, err
	}
	for _, record := range checks {
		p.checks[record.CheckID] = record
		if len(p.revisions) > 0 {
			if id := inboxUserID(record.ClientID); id > 0 {
				clientIDs = append(clientIDs, id)
			}
		}
		if id := SubjectIntakeIDFromSearchParameters(record.SearchParameters); id != "" {
			intakeIDs = append(intakeIDs, id)
		}
		if id := subjectCaseIDFromConflictParameters(record.SearchParameters); id > 0 {
			caseIDs = append(caseIDs, id)
		}
		userIDs = append(userIDs, record.UserID)
	}
	intakes, err := inboxLoad[inboxIntake](db.Table("case_intakes"), "id", intakeIDs)
	if err != nil {
		return nil, err
	}
	for _, intake := range intakes {
		p.intakes[intake.ID] = intake
		clientIDs = append(clientIDs, intake.ClientID)
		userIDs = append(userIDs, inboxUserID(intake.CreatedBy))
	}
	if len(p.revisions) > 0 {
		clients, err := inboxLoad[models.Client](db.Select("id"), "id", clientIDs)
		if err != nil {
			return nil, err
		}
		for _, client := range clients {
			p.clients[client.ID] = true
		}
	}
	caseSelect := db.Select("id, client_id, lawyer_id, created_by, ethical_wall_enabled, conflict_check_id")
	byID, err := inboxLoad[models.Case](caseSelect, "id", caseIDs)
	if err != nil {
		return nil, err
	}
	byCheck, err := inboxLoad[models.Case](caseSelect, "conflict_check_id", checkIDs)
	if err != nil {
		return nil, err
	}
	byClient, err := inboxLoad[models.Case](caseSelect, "client_id", clientIDs)
	if err != nil {
		return nil, err
	}
	for _, c := range append(append(byID, byCheck...), byClient...) {
		p.cases[c.ID] = c
		userIDs = append(userIDs, c.LawyerID, inboxUserID(c.CreatedBy))
	}
	var protectedIDs []uint
	for id, c := range p.cases {
		if c.EthicalWallEnabled {
			protectedIDs = append(protectedIDs, id)
		}
	}
	if len(protectedIDs) > 0 {
		type wallRow struct{ CaseID uint }
		rows, err := inboxLoad[wallRow](db.Table("case_ethical_wall_whitelist").Select("case_id").Where("user_id = ?", userID), "case_id", protectedIDs)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			p.walls[row.CaseID] = true
		}
	}
	if len(checkIDs) > 0 {
		assignments, err := inboxLoad[models.ConflictReviewerAssignment](db.Order("created_at DESC, id DESC"), "check_id", checkIDs)
		if err != nil {
			return nil, err
		}
		rosters := map[string][]models.ConflictReviewerAssignment{}
		for _, a := range assignments {
			rosters[a.CheckID] = append(rosters[a.CheckID], a)
			if a.Status != models.ConflictReviewerAssignmentActive || a.EffectiveFrom != nil && a.EffectiveFrom.After(p.now) || a.EffectiveTo != nil && !a.EffectiveTo.After(p.now) {
				continue
			}
			if _, exists := p.assignments[a.CheckID]; !exists {
				p.assignments[a.CheckID] = a
				userIDs = append(userIDs, a.ReviewerID)
			}
		}
		for id, roster := range rosters {
			p.assignmentGenerations[id] = conflictAssignmentGeneration(roster)
		}
		reviews, err := inboxLoad[models.ConflictReview](db.Order("created_at DESC, id DESC"), "check_id", checkIDs)
		if err != nil {
			return nil, err
		}
		for _, r := range reviews {
			if _, ok := p.reviews[r.CheckID]; !ok {
				p.reviews[r.CheckID] = r
			}
		}
	}
	users, err := inboxLoad[models.User](db.Select("id, name, role, status, manager_id"), "id", userIDs)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		p.users[u.ID] = u
	}
	for _, item := range items {
		if item.CompletionMode == "workflow" && item.SourceRef != "" && item.WorkflowTaskKey != "" && item.UserID == userID {
			if action := p.action(item); action != nil {
				result[item.ID] = action
			}
		}
	}
	return result, nil
}

func inboxApprovalCheckID(a *models.ApprovalRequest, m map[string]interface{}) string {
	if strings.TrimSpace(a.ConflictCheckID) != "" {
		return strings.TrimSpace(a.ConflictCheckID)
	}
	if id := inboxText(m, "conflict_check_id"); id != "" {
		return id
	}
	return inboxText(m, "conflict_task_id")
}

func (p *inboxProjection) caseVisible(c models.Case) bool {
	return !c.EthicalWallEnabled || p.walls[c.ID]
}

func (p *inboxProjection) intakeVisible(intake inboxIntake) bool {
	if intake.ID == "" || intake.ClientID == 0 || inboxUserID(intake.CreatedBy) == 0 {
		return false
	}
	if inboxUserID(intake.CreatedBy) != p.actor.ID && !IsConflictReviewRole(p.actor.Role) && !IsBusinessMatterManagementRole(p.actor.Role) {
		return false
	}
	return p.clientVisible(intake.ClientID)
}

func (p *inboxProjection) clientVisible(clientID uint) bool {
	total, visible := 0, 0
	for _, c := range p.cases {
		if c.ClientID == clientID {
			total++
			if p.caseVisible(c) {
				visible++
			}
		}
	}
	return total == 0 || visible > 0
}

func (p *inboxProjection) caseContextVisible(c models.Case) bool {
	return p.caseVisible(c) && (IsConflictReviewRole(p.actor.Role) || IsBusinessMatterManagementRole(p.actor.Role) || ownsCase(AuthActor{UserID: p.actor.ID, Role: p.actor.Role}, &c))
}

// Only a validated active submission in needs_revision may have a different
// editable client. Both sides remain authorized; all other states use the
// original strict check/intake comparison below.
func (p *inboxProjection) revisionVisible(a models.ApprovalRequest, metadata map[string]interface{}) bool {
	scope, ok := p.revisions[a.ID]
	if !ok {
		return false
	}
	intakeID := inboxText(metadata, "intake_id")
	intake, ok := p.intakes[intakeID]
	if !ok || intake.CreatedBy != a.ApplicantID || !p.clients[intake.ClientID] || !p.intakeVisible(intake) {
		return false
	}
	objects := len(scope.clients) + len(scope.cases)
	for _, id := range scope.clients {
		if !p.clients[id] || !p.clientVisible(id) {
			return false
		}
	}
	for _, id := range scope.cases {
		c, exists := p.cases[id]
		if !exists || !p.caseContextVisible(c) {
			return false
		}
	}
	for _, id := range scope.checks {
		record, exists := p.checks[id]
		if !exists {
			return false
		}
		if text := strings.TrimSpace(record.ClientID); text != "" && text != "0" {
			clientID := inboxUserID(text)
			if !p.clients[clientID] || !p.clientVisible(clientID) {
				return false
			}
			objects++
		}
		if c, found := p.checkCase(record); found {
			if !p.caseContextVisible(c) {
				return false
			}
			objects++
		} else if subjectCaseIDFromConflictParameters(record.SearchParameters) > 0 || SubjectIntakeIDFromSearchParameters(record.SearchParameters) != intakeID || record.UserID != inboxUserID(a.ApplicantID) {
			return false
		}
	}
	return objects > 0
}

func (p *inboxProjection) checkCase(record models.ConflictCheckRecord) (models.Case, bool) {
	if id := subjectCaseIDFromConflictParameters(record.SearchParameters); id > 0 {
		c, ok := p.cases[id]
		return c, ok
	}
	var found models.Case
	count := 0
	for _, c := range p.cases {
		if c.ConflictCheckID == record.CheckID {
			found = c
			count++
		}
	}
	return found, count == 1
}

func (p *inboxProjection) checkVisible(record models.ConflictCheckRecord) bool {
	if c, ok := p.checkCase(record); ok {
		return p.caseContextVisible(c)
	}
	if subjectCaseIDFromConflictParameters(record.SearchParameters) > 0 {
		return false
	}
	intake, ok := p.intakes[SubjectIntakeIDFromSearchParameters(record.SearchParameters)]
	return ok && inboxUserID(intake.CreatedBy) == record.UserID && (record.ClientID == "" || record.ClientID == "0" || record.ClientID == strconv.FormatUint(uint64(intake.ClientID), 10)) && p.intakeVisible(intake)
}

func (p *inboxProjection) approvalVisible(a models.ApprovalRequest) bool {
	uid := strconv.FormatUint(uint64(p.actor.ID), 10)
	if a.ApplicantID != uid && a.CurrentApproverID != uid && !p.history[a.ID] {
		return false
	}
	m, err := approvalInboxMetadata(&a)
	if err != nil {
		return false
	}
	checkID := inboxApprovalCheckID(&a, m)
	for _, key := range []string{"conflict_check_id", "conflict_task_id"} {
		if id := inboxText(m, key); id != "" && id != checkID {
			return false
		}
	}
	controlled := a.Type == "case_creation" || a.Type == "conflict" || a.Type == "conflict_approval" || a.Type == "waiver" || checkID != ""
	if !controlled {
		return true
	}
	if a.ActiveSubmissionSnapshotID != nil && a.Status == models.ApprovalStatusNeedsRevision {
		return p.revisionVisible(a, m)
	}
	record, ok := p.checks[checkID]
	return ok && p.checkVisible(record)
}

func blockedInboxAction(kind, id, status, version, code, reason string) *WorkflowAction {
	return &WorkflowAction{ResourceType: kind, ResourceID: id, ActionCode: "view_only", ActionLabel: "查看状态", WorkflowStatus: status, CompletionMode: "workflow", ActionVersion: version, BlockedCode: code, BlockedReason: reason}
}

func (p *inboxProjection) assignee(userID uint) *WorkflowAssignee {
	u, ok := p.users[userID]
	if !ok {
		return nil
	}
	return &WorkflowAssignee{ID: strconv.FormatUint(uint64(u.ID), 10), Name: u.Name}
}

func (p *inboxProjection) action(item *models.InboxItem) *WorkflowAction {
	switch item.SourceType {
	case "approval", "intake":
		approvalID := strings.TrimPrefix(item.WorkflowGroup, "approval:")
		if !strings.HasPrefix(item.WorkflowGroup, "approval:") {
			return nil
		}
		a, ok := p.approvals[approvalID]
		if !ok || !p.approvalVisible(a) {
			return nil
		}
		if item.SourceType == "approval" && item.SourceRef != a.ID {
			return nil
		}
		version := a.UpdatedAt.UTC().Format(time.RFC3339Nano)
		action := blockedInboxAction(item.SourceType, item.SourceRef, a.Status, version, "WAITING_FOR_OTHER", "等待当前处理人办理")
		action.AssignedTo = p.assignee(inboxUserID(a.CurrentApproverID))
		current := !item.IsCompleted && item.WorkflowVersion == approvalInboxVersion(&a)
		if item.SourceType == "intake" {
			m, _ := approvalInboxMetadata(&a)
			intake, exists := p.intakes[item.SourceRef]
			if !exists || inboxText(m, "intake_id") != item.SourceRef || !p.intakeVisible(intake) {
				return nil
			}
			action.ApprovalID = a.ID
			action.AssignedTo = p.assignee(inboxUserID(a.ApplicantID))
			if current && a.Status == models.ApprovalStatusNeedsRevision && item.WorkflowPurpose == "supplement" && inboxUserID(a.ApplicantID) == p.actor.ID && inboxUserID(intake.CreatedBy) == p.actor.ID && (normalizeRole(p.actor.Role) == "lawyer" || IsBusinessMatterManagementRole(p.actor.Role)) {
				action.ActionCode, action.ActionLabel, action.BlockedCode, action.BlockedReason = "supplement", "补充申请", "", ""
			}
		} else if current && item.WorkflowPurpose == "decide_approval" && inboxUserID(a.CurrentApproverID) == p.actor.ID && a.ApplicantID != a.CurrentApproverID && (a.Status == models.ApprovalStatusSubmitted || a.Status == models.ApprovalStatusUnderReview || a.Status == models.ApprovalStatusResubmitted) && (normalizeRole(p.actor.Role) == "lawyer" || IsConflictReviewRole(p.actor.Role) || IsBusinessMatterManagementRole(p.actor.Role)) {
			action.ActionCode, action.ActionLabel, action.BlockedCode, action.BlockedReason = "decide_approval", "办理审批", "", ""
		}
		if !current && item.WorkflowPurpose != "result" {
			action.BlockedCode, action.BlockedReason = "TASK_SUPERSEDED", "该任务已完成或转交，请查看当前业务状态"
		}
		if a.Status == models.ApprovalStatusApproved && RequiresCaseCreationApproval(&a) && !a.CaseCreated {
			action.WorkflowStatus = "approved_pending_case"
			action.BlockedCode, action.BlockedReason = "CASE_CREATION_PENDING", "审批已通过，正式成案尚未完成"
			if a.CaseCreationStatus == "failed" {
				action.WorkflowStatus, action.BlockedCode, action.BlockedReason = "case_creation_failed", "CASE_CREATION_FAILED", "审批已通过，成案失败，可从审批详情重试"
			}
		}
		if a.Status == models.ApprovalStatusApproved && RequiresCaseCreationApproval(&a) && a.CaseCreated {
			metadata, _ := approvalInboxMetadata(&a)
			linkedCase, exists := p.cases[inboxUserID(a.CreatedCaseID)]
			if exists && linkedCase.ConflictCheckID != "" && linkedCase.ConflictCheckID == inboxApprovalCheckID(&a, metadata) {
				action.WorkflowStatus, action.BlockedCode, action.BlockedReason = "case_created", "", ""
			} else {
				action.WorkflowStatus, action.BlockedCode, action.BlockedReason = "case_creation_incomplete", "CASE_CREATION_INCOMPLETE", "审批已通过，正式案件关联尚未验证"
			}
		}
		return action
	case "conflict":
		record, ok := p.checks[item.SourceRef]
		if !ok || item.WorkflowGroup != "conflict:"+record.CheckID || !p.checkVisible(record) {
			return nil
		}
		assignment, assigned := p.assignments[record.CheckID]
		_, reviewed := p.reviews[record.CheckID]
		assignmentID := "unassigned"
		if assigned {
			assignmentID = assignment.ID
		}
		count := "0"
		if reviewed {
			count = "1"
		}
		generation := p.assignmentGenerations[record.CheckID]
		if generation == "" {
			generation = conflictAssignmentGeneration(nil)
		}
		version := inboxDigest(record.CheckID, assignmentID, record.CheckStatus, count, generation)
		action := blockedInboxAction("conflict", record.CheckID, record.CheckStatus, version, "WAITING_FOR_REVIEWER", "等待指定的独立复核人办理")
		if assigned {
			action.AssignedTo = p.assignee(assignment.ReviewerID)
		} else {
			action.BlockedCode, action.BlockedReason = "REVIEWER_ASSIGNMENT_REQUIRED", "尚未指定独立复核人"
		}
		if reviewed {
			action.WorkflowStatus, action.BlockedCode, action.BlockedReason = "reviewed", "REVIEW_COMPLETED", "独立复核已完成"
			return action
		}
		if item.IsCompleted || item.WorkflowVersion != version {
			action.BlockedCode, action.BlockedReason = "TASK_SUPERSEDED", "复核任务已转交或失效"
			return action
		}
		if item.WorkflowPurpose == "review_conflict" && p.canReview(record, assignment) {
			action.ActionCode, action.ActionLabel, action.BlockedCode, action.BlockedReason = "review_conflict", "独立复核", "", ""
		} else if item.WorkflowPurpose == "review_conflict" && assignment.ReviewerID == p.actor.ID {
			action.BlockedCode, action.BlockedReason = "REVIEW_NOT_READY", "当前证据、事实版本或独立性条件尚不满足复核要求"
		}
		return action
	case "case":
		c, ok := p.cases[inboxUserID(item.SourceRef)]
		if !ok || !p.caseVisible(c) || (!IsBusinessMatterManagementRole(p.actor.Role) && !ownsCase(AuthActor{UserID: p.actor.ID, Role: p.actor.Role}, &c)) {
			return nil
		}
		return blockedInboxAction("case", item.SourceRef, "created", item.WorkflowVersion, "", "")
	}
	return nil
}

func (p *inboxProjection) canReview(record models.ConflictCheckRecord, assignment models.ConflictReviewerAssignment) bool {
	if assignment.ReviewerID != p.actor.ID || !assignment.RecusalDeclared || !IsConflictReviewRole(p.actor.Role) || record.UserID == p.actor.ID || !strings.EqualFold(record.CheckStatus, "COMPLETED") {
		return false
	}
	decision, _ := record.CheckResult["decision"].(map[string]interface{})
	if !strings.EqualFold(inboxText(decision, "coverageStatus"), "COMPLETE") {
		return false
	}
	involved := []uint{record.UserID}
	if c, ok := p.checkCase(record); ok {
		involved = append(involved, c.LawyerID, inboxUserID(c.CreatedBy))
	}
	if intakeID := SubjectIntakeIDFromSearchParameters(record.SearchParameters); intakeID != "" {
		intake, ok := p.intakes[intakeID]
		if !ok || inboxText(intake.Metadata, "conflict_check_id") != record.CheckID {
			return false
		}
		if raw := inboxText(intake.Metadata, "lawyer_facts_confirmed_at"); raw != "" {
			at, err := time.Parse(time.RFC3339, raw)
			if err != nil || record.CheckTime.Before(at) {
				return false
			}
		}
		involved = append(involved, inboxUserID(intake.CreatedBy))
	}
	for _, id := range involved {
		if id == p.actor.ID || (p.actor.ManagerID != nil && *p.actor.ManagerID == id) {
			return false
		}
		if u, ok := p.users[id]; ok && u.ManagerID != nil && *u.ManagerID == p.actor.ID {
			return false
		}
	}
	return true
}
