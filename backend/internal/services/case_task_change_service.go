package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"law-oa-go/internal/models"
)

// TaskCapabilities is the server-computed permission projection for one
// task row. The frontend renders its action buttons from these flags
// instead of re-deriving permissions from created_by/assignee_id; the
// server re-validates every write regardless.
type TaskCapabilities struct {
	CanAdjust      bool `json:"can_adjust"`
	CanCancel      bool `json:"can_cancel"`
	CanComplete    bool `json:"can_complete"`
	CanViewChanges bool `json:"can_view_changes"`
}

// TaskListItem is the list DTO: the row plus its capability flags.
type TaskListItem struct {
	models.CaseTask
	Capabilities TaskCapabilities `json:"capabilities"`
}

// caseTaskChangeReasonMaxLen bounds one change reason. 2000 runes keeps
// every free-text legal record in this service under one reviewed cap.
const caseTaskChangeReasonMaxLen = 2000

// creatorStillAuthorized re-checks the task creator against the live
// permission state: active user row, current case readability, and the
// ethical wall. Creation-time access never outlives deactivation,
// team removal, or a wall added later (review 2026-10-02). The service
// passes its transaction handle so the check reads the same snapshot it
// is about to write; without authz (fixture-only unit tests) it fails
// closed like every other live check.
func (s *CaseTaskChangeService) creatorStillAuthorized(tx *gorm.DB, caseID uint, userID uint) bool {
	if s.tasks.authz != nil {
		var roleRow struct{ Role string }
		if err := tx.Table("users").Select("role").Where("id = ? AND deleted_at IS NULL AND status = 'active'", userID).Take(&roleRow).Error; err != nil {
			return false
		}
		actor := AuthActor{UserID: userID, Role: roleRow.Role}
		allowed, err := s.tasks.authz.CanReadCase(context.Background(), actor, caseID)
		return err == nil && allowed
	}
	// Fixture fallback mirrors caseTaskAssigneeAllowed: active user plus
	// ownership/whitelist SQL, always requiring an active user row.
	var count int64
	err := tx.Table("cases").Where("id = ? AND deleted_at IS NULL AND (lawyer_id = ? OR created_by = ? OR EXISTS (SELECT 1 FROM case_ethical_wall_whitelist w WHERE w.case_id = cases.id AND w.user_id = ?))", caseID, userID, fmt.Sprint(userID), userID).Count(&count).Error
	if err != nil || count == 0 {
		return false
	}
	var user struct {
		DeletedAt *time.Time
		Status    *string
	}
	if err := tx.Table("users").Select("deleted_at, status").Where("id = ?", userID).Take(&user).Error; err != nil {
		return false
	}
	return user.DeletedAt == nil && (user.Status == nil || *user.Status == "active")
}

// CanViewChanges is the exported view check used by the HTTP layer.
func (s *CaseTaskChangeService) CanViewChanges(actorID uint, task *models.CaseTask) bool {
	return s.canViewChanges(actorID, task)
}

// canViewChanges reports whether actorID may read this task's audit
// trail: creator (still authorized), current assignee, or genuine case
// manager. The handler calls this before ListChanges.
func (s *CaseTaskChangeService) canViewChanges(actorID uint, task *models.CaseTask) bool {
	if task == nil || actorID == 0 {
		return false
	}
	if actorID == task.AssigneeID {
		return true
	}
	if actorID == task.CreatedBy {
		var ok bool
		err := s.tasks.db.Transaction(func(tx *gorm.DB) error {
			ok = s.creatorStillAuthorized(tx, task.CaseID, actorID)
			return nil
		})
		if err == nil && ok {
			return true
		}
	}
	var managerOK bool
	err := s.tasks.db.Transaction(func(tx *gorm.DB) error {
		managerOK = s.tasks.caseTaskCaseManagerAllowed(tx, task.CaseID, actorID)
		return nil
	})
	return err == nil && managerOK
}

// Sentinel errors for the task change service; the HTTP layer maps each
// to its own status and code.
var (
	ErrCaseTaskChangeStale          = errors.New("CASE_TASK_CHANGE_STALE: 任务已被他人修改，请刷新后基于最新内容重试")
	ErrCaseTaskChangeNoop           = errors.New("CASE_TASK_CHANGE_NOOP: 没有提供需要变更的字段")
	ErrCaseTaskChangeForbidden      = errors.New("CASE_TASK_CHANGE_FORBIDDEN: 只有案件管理者或任务创建人可以调整或取消任务")
	ErrCaseTaskChangeViewForbidden  = errors.New("CASE_TASK_CHANGE_VIEW_FORBIDDEN: 只有任务相关方可查看变更记录")
	ErrCaseTaskChangeNotOpen        = errors.New("CASE_TASK_CHANGE_NOT_OPEN: 任务已结束，不能调整")
	ErrCaseTaskChangeRequestMissing = errors.New("CASE_TASK_CHANGE_REQUEST_MISSING: 缺少并发保护字段或 request_id")
)

// CaseTaskChangeRequest is the shared adjust/cancel contract body. At
// least one of Version or ExpectedUpdatedAt must be present: a client
// that cannot prove what it saw cannot overwrite what others changed.
type CaseTaskChangeRequest struct {
	Version           *uint      `json:"version,omitempty"`
	ExpectedUpdatedAt *string    `json:"expected_updated_at,omitempty"`
	AssigneeID        *uint      `json:"assignee_id,omitempty"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	ClearDueDate      bool       `json:"clear_due_date,omitempty"`
	Reason            string     `json:"reason,omitempty"`
}

// CaseTaskChangeView is the client projection of one audit record.
type CaseTaskChangeView struct {
	ID            uint       `json:"id"`
	CreatedAt     time.Time  `json:"created_at"`
	TaskID        uint       `json:"task_id"`
	CaseID        uint       `json:"case_id"`
	ChangeType    string     `json:"change_type"`
	ActorID       uint       `json:"actor_id"`
	OldAssigneeID *uint      `json:"old_assignee_id,omitempty"`
	NewAssigneeID *uint      `json:"new_assignee_id,omitempty"`
	OldDueDate    *time.Time `json:"old_due_date,omitempty"`
	NewDueDate    *time.Time `json:"new_due_date,omitempty"`
	OldVersion    uint       `json:"old_version"`
	NewVersion    uint       `json:"new_version"`
	Version       *uint      `json:"version,omitempty"`
	Reason        string     `json:"reason,omitempty"`
}

func viewFromChange(m *models.CaseTaskChange) *CaseTaskChangeView {
	return &CaseTaskChangeView{
		ID: m.ID, CreatedAt: m.CreatedAt, TaskID: m.TaskID, CaseID: m.CaseID,
		ChangeType: m.ChangeType, ActorID: m.ActorID,
		OldAssigneeID: m.OldAssigneeID, NewAssigneeID: m.NewAssigneeID,
		OldDueDate: m.OldDueDate, NewDueDate: m.NewDueDate,
		OldVersion: m.OldVersion, NewVersion: m.NewVersion,
		Version: m.Version, Reason: m.Reason,
	}
}

// CaseTaskChangeService implements adjust / cancel / history for case
// tasks. Permission logic reuses CaseTaskService's existing object-level
// checks (caseTaskCaseManagerAllowed plus the creator identity); nothing
// here redefines who may enter a case.
type CaseTaskChangeService struct {
	tasks *CaseTaskService
}

func NewCaseTaskChangeService(db *gorm.DB) *CaseTaskChangeService {
	return &CaseTaskChangeService{tasks: &CaseTaskService{db: db}}
}

// NewCaseTaskChangeServiceWithAuthz is the production constructor: the
// assignee-legality and manager checks run through AuthorizationService.
func NewCaseTaskChangeServiceWithAuthz(db *gorm.DB, authz *AuthorizationService) *CaseTaskChangeService {
	return &CaseTaskChangeService{tasks: &CaseTaskService{db: db, authz: authz}}
}

func normalizeChangeReason(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if len([]rune(trimmed)) > caseTaskChangeReasonMaxLen {
		return "", fmt.Errorf("变更原因超过长度限制（最多 %d 字）", caseTaskChangeReasonMaxLen)
	}
	return trimmed, nil
}

// versionGuard checks the optimistic-lock precondition against the row
// the transaction just locked. A client may present the version number
// or the updated_at it saw; missing both fails closed. The timestamp
// comparison is exact (RFC3339Nano); there is deliberately NO
// second-truncation fallback - two changes can land inside one second
// and truncation would widen the lock. The version column is the
// primary guard; the timestamp leg exists for clients that never saw a
// version (pre-migration rows).
func versionGuard(rowVersion uint, rowUpdatedAt time.Time, req *CaseTaskChangeRequest) error {
	if req == nil {
		return ErrCaseTaskChangeRequestMissing
	}
	hasVersion := req.Version != nil
	rawStamp := ""
	if req.ExpectedUpdatedAt != nil {
		rawStamp = strings.TrimSpace(*req.ExpectedUpdatedAt)
	}
	if !hasVersion && rawStamp == "" {
		return ErrCaseTaskChangeRequestMissing
	}
	if hasVersion && *req.Version != rowVersion {
		return ErrCaseTaskChangeStale
	}
	if rawStamp != "" {
		expected, err := time.Parse(time.RFC3339Nano, rawStamp)
		if err != nil {
			return fmt.Errorf("expected_updated_at 格式无效：%w", err)
		}
		// Exact comparison only: same-second-but-earlier stamps are stale
		// and must be rejected (independent regression
		// TestIndependentTaskTimestampRejectsSameSecondStaleValue).
		if !expected.Equal(rowUpdatedAt) {
			return ErrCaseTaskChangeStale
		}
	}
	return nil
}

// Adjust applies one atomic change (reassign / due-date / combined) with
// optimistic locking, an audit row, and inbox synchronization. Reassign
// keeps the same task ID: the old assignee's open todo closes and the
// new assignee receives a fresh todo in the same commit.
func (s *CaseTaskChangeService) Adjust(actorID uint, taskID uint, req *CaseTaskChangeRequest) (*models.CaseTask, *CaseTaskChangeView, error) {
	if req == nil {
		return nil, nil, ErrCaseTaskChangeRequestMissing
	}
	reason, err := normalizeChangeReason(req.Reason)
	if err != nil {
		return nil, nil, err
	}
	if req.AssigneeID == nil && req.DueDate == nil && !req.ClearDueDate {
		return nil, nil, ErrCaseTaskChangeNoop
	}
	if req.ClearDueDate && req.DueDate != nil {
		return nil, nil, fmt.Errorf("clear_due_date 与 due_date 不能同时提供")
	}
	now := time.Now()
	var outTask *models.CaseTask
	var outChange *CaseTaskChangeView
	txErr := s.tasks.db.Transaction(func(tx *gorm.DB) error {
		var task models.CaseTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", taskID).Take(&task).Error; err != nil {
			return fmt.Errorf("任务不存在：%w", err)
		}
		// Actor must be the task creator (re-checked against the live
		// permission state: active user, current case read access, ethical
		// wall) OR a genuine case manager. Creation-time access alone is
		// never sufficient (review 2026-10-02).
		creatorLive := actorID == task.CreatedBy && s.creatorStillAuthorized(tx, task.CaseID, actorID)
		if !creatorLive && !s.tasks.caseTaskCaseManagerAllowed(tx, task.CaseID, actorID) {
			return ErrCaseTaskChangeForbidden
		}
		if task.Status != models.CaseTaskStatusOpen {
			return ErrCaseTaskChangeNotOpen
		}
		// Read the version column through a projection: the shared
		// models.CaseTask struct has no Version field (this slice must not
		// edit matter_collaboration.go).
		var vrow struct {
			Version   uint
			UpdatedAt time.Time
		}
		if err := tx.Table("case_tasks").Select("version, updated_at").Where("id = ?", taskID).Take(&vrow).Error; err != nil {
			return fmt.Errorf("读取任务版本失败：%w", err)
		}
		if err := versionGuard(vrow.Version, vrow.UpdatedAt, req); err != nil {
			return err
		}
		// Resolve the requested assignee (optional).
		newAssignee := task.AssigneeID
		if req.AssigneeID != nil && *req.AssigneeID != task.AssigneeID {
			if !s.tasks.caseTaskAssigneeAllowed(tx, task.CaseID, *req.AssigneeID) {
				return ErrCaseTaskAssigneeInvalid
			}
			newAssignee = *req.AssigneeID
		}
		// Resolve the requested due date (optional).
		newDue := task.DueDate
		if req.ClearDueDate {
			newDue = nil
		} else if req.DueDate != nil {
			t := req.DueDate.UTC()
			newDue = &t
		}
		if newAssignee == task.AssigneeID && dueDateFingerprint(newDue) == dueDateFingerprint(task.DueDate) {
			return ErrCaseTaskChangeNoop
		}
		newVersion := vrow.Version + 1
		updates := map[string]interface{}{
			"updated_at": now,
			"version":    newVersion,
		}
		if newAssignee != task.AssigneeID {
			updates["assignee_id"] = newAssignee
		}
		if dueDateFingerprint(newDue) != dueDateFingerprint(task.DueDate) {
			updates["due_date"] = newDue
		}
		// CAS on version: the same guard that rejected the stale request in
		//-process also rejects a row that moved between SELECT and UPDATE
		// (another writer committed first). RowsAffected==0 fails closed.
		cas := tx.Model(&models.CaseTask{}).Where("id = ? AND version = ?", taskID, vrow.Version).Updates(updates)
		if cas.Error != nil {
			return fmt.Errorf("更新任务失败：%w", cas.Error)
		}
		if cas.RowsAffected == 0 {
			return ErrCaseTaskChangeStale
		}
		changeType := models.CaseTaskChangeDueDateChange
		if newAssignee != task.AssigneeID {
			changeType = models.CaseTaskChangeReassign
		}
		auditVersion := vrow.Version
		audit := &models.CaseTaskChange{
			TaskID: taskID, CaseID: task.CaseID, ChangeType: changeType,
			ActorID:       actorID,
			OldAssigneeID: &task.AssigneeID, NewAssigneeID: &newAssignee,
			OldDueDate: task.DueDate, NewDueDate: newDue,
			OldVersion: vrow.Version, NewVersion: newVersion,
			Version: &auditVersion, Reason: reason,
		}
		if err := tx.Create(audit).Error; err != nil {
			return fmt.Errorf("写入变更记录失败：%w", err)
		}
		// Inbox synchronization. On reassignment: close the old assignee's
		// open todo and open a new one for the incoming assignee, same
		// task ID, same workflow key prefix.
		if newAssignee != task.AssigneeID {
			nowClose := time.Now()
			if err := tx.Model(&models.InboxItem{}).
				Where("workflow_task_key = ? AND completion_mode = ? AND is_completed = ?", caseTaskWorkflowKey(taskID), "workflow", false).
				Updates(map[string]interface{}{
					"is_completed": true, "completed_at": nowClose, "updated_at": nowClose,
				}).Error; err != nil {
				return fmt.Errorf("关闭旧负责人待办失败：%w", err)
			}
			updated := task
			updated.AssigneeID = newAssignee
			updated.DueDate = newDue
			if err := openCaseTaskInboxTx(tx, &updated); err != nil {
				return err
			}
		} else if _, changed := updates["due_date"]; changed {
			if err := tx.Model(&models.InboxItem{}).
				Where("workflow_task_key = ? AND completion_mode = ? AND is_completed = ?", caseTaskWorkflowKey(taskID), "workflow", false).
				Updates(map[string]interface{}{
					"due_date": newDue, "updated_at": now,
				}).Error; err != nil {
				return fmt.Errorf("同步待办截止时间失败：%w", err)
			}
		}
		outTask = &models.CaseTask{
			ID: task.ID, CreatedAt: task.CreatedAt, UpdatedAt: now,
			CaseID: task.CaseID, Title: task.Title, Detail: task.Detail,
			AssigneeID: newAssignee, CreatedBy: task.CreatedBy,
			DueDate: newDue, Status: task.Status,
			CompletedBy: task.CompletedBy, CompletedAt: task.CompletedAt,
			Version: newVersion,
		}
		outChange = viewFromChange(audit)
		return nil
	})
	if txErr != nil {
		return nil, nil, txErr
	}
	return outTask, outChange, nil
}

// Cancel marks an open task cancelled, records the audit entry, and
// closes every open todo tied to the task in the same transaction. An
// already-cancelled task with a matching version replays idempotently
// without a second audit row.
func (s *CaseTaskChangeService) Cancel(actorID uint, taskID uint, req *CaseTaskChangeRequest) (*models.CaseTask, *CaseTaskChangeView, bool, error) {
	if req == nil {
		return nil, nil, false, ErrCaseTaskChangeRequestMissing
	}
	reason, err := normalizeChangeReason(req.Reason)
	if err != nil {
		return nil, nil, false, err
	}
	replayed := false
	var outTask *models.CaseTask
	var outChange *CaseTaskChangeView
	txErr := s.tasks.db.Transaction(func(tx *gorm.DB) error {
		var task models.CaseTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", taskID).Take(&task).Error; err != nil {
			return fmt.Errorf("任务不存在：%w", err)
		}
		creatorLive := actorID == task.CreatedBy && s.creatorStillAuthorized(tx, task.CaseID, actorID)
		if !creatorLive && !s.tasks.caseTaskCaseManagerAllowed(tx, task.CaseID, actorID) {
			return ErrCaseTaskChangeForbidden
		}
		var vrow struct {
			Version   uint
			UpdatedAt time.Time
		}
		if err := tx.Table("case_tasks").Select("version, updated_at").Where("id = ?", taskID).Take(&vrow).Error; err != nil {
			return fmt.Errorf("读取任务版本失败：%w", err)
		}
		if task.Status == models.CaseTaskStatusCancelled {
			// Replay: still require the caller to know the current version.
			if err := versionGuard(vrow.Version, vrow.UpdatedAt, req); err != nil {
				return err
			}
			replayed = true
			outTask = &task
			outChange = nil
			return nil
		}
		if task.Status != models.CaseTaskStatusOpen {
			return ErrCaseTaskChangeNotOpen
		}
		if err := versionGuard(vrow.Version, vrow.UpdatedAt, req); err != nil {
			return err
		}
		now := time.Now()
		newVersion := vrow.Version + 1
		// CAS on version: a completion that committed first must win over
		// this cancel; RowsAffected==0 reports the race instead of
		// overwriting the completed row.
		cas := tx.Model(&models.CaseTask{}).Where("id = ? AND version = ?", taskID, vrow.Version).Updates(map[string]interface{}{
			"status": models.CaseTaskStatusCancelled, "updated_at": now, "version": newVersion,
		})
		if cas.Error != nil {
			return fmt.Errorf("取消任务失败：%w", cas.Error)
		}
		if cas.RowsAffected == 0 {
			return ErrCaseTaskChangeStale
		}
		auditVersion := vrow.Version
		audit := &models.CaseTaskChange{
			TaskID: taskID, CaseID: task.CaseID, ChangeType: models.CaseTaskChangeCancel,
			ActorID:       actorID,
			OldAssigneeID: &task.AssigneeID, NewAssigneeID: &task.AssigneeID,
			OldDueDate: task.DueDate, NewDueDate: task.DueDate,
			OldVersion: vrow.Version, NewVersion: newVersion,
			Version: &auditVersion, Reason: reason,
		}
		if err := tx.Create(audit).Error; err != nil {
			return fmt.Errorf("写入变更记录失败：%w", err)
		}
		if err := tx.Model(&models.InboxItem{}).
			Where("workflow_task_key = ? AND completion_mode = ? AND is_completed = ?", caseTaskWorkflowKey(taskID), "workflow", false).
			Updates(map[string]interface{}{
				"is_completed": true, "completed_at": now, "updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("关闭任务待办失败：%w", err)
		}
		updated := task
		updated.Status = models.CaseTaskStatusCancelled
		updated.UpdatedAt = now
		updated.Version = newVersion
		outTask = &updated
		outChange = viewFromChange(audit)
		return nil
	})
	if txErr != nil {
		return nil, nil, false, txErr
	}
	return outTask, outChange, replayed, nil
}

// LoadTaskForChange loads one task (with version) for the handler layer.
func (s *CaseTaskChangeService) LoadTaskForChange(taskID uint) (*models.CaseTask, uint, error) {
	var task models.CaseTask
	if err := s.tasks.db.Where("id = ?", taskID).Take(&task).Error; err != nil {
		return nil, 0, fmt.Errorf("任务不存在：%w", err)
	}
	var vrow struct{ Version uint }
	if err := s.tasks.db.Table("case_tasks").Select("version").Where("id = ?", taskID).Take(&vrow).Error; err != nil {
		return nil, 0, fmt.Errorf("读取任务版本失败：%w", err)
	}
	return &task, vrow.Version, nil
}

// FindTaskByRequestId resolves one creation attempt by its idempotency
// key. The key is unique per creator, so the acting user's id scopes the
// lookup: user A cannot probe user B's keys.
func (s *CaseTaskChangeService) FindTaskByRequestId(creatorID uint, requestID string) (*models.CaseTask, uint, error) {
	key := strings.TrimSpace(requestID)
	if key == "" {
		return nil, 0, ErrCaseTaskChangeRequestMissing
	}
	var task models.CaseTask
	if err := s.tasks.db.Where("created_by = ? AND idempotency_key = ?", creatorID, key).Take(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("查询任务失败：%w", err)
	}
	var vrow struct{ Version uint }
	if err := s.tasks.db.Table("case_tasks").Select("version").Where("id = ?", task.ID).Take(&vrow).Error; err != nil {
		return nil, 0, fmt.Errorf("读取任务版本失败：%w", err)
	}
	return &task, vrow.Version, nil
}

// ListChanges returns the change history of one task, newest first. The
// handler enforces view permission before calling this.
func (s *CaseTaskChangeService) ListChanges(taskID uint) ([]*CaseTaskChangeView, error) {
	rows := []models.CaseTaskChange{}
	if err := s.tasks.db.Where("task_id = ?", taskID).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("读取变更记录失败：%w", err)
	}
	views := make([]*CaseTaskChangeView, 0, len(rows))
	for i := range rows {
		views = append(views, viewFromChange(&rows[i]))
	}
	return views, nil
}

// ListTasksWithCapabilities projects one case's task rows into the
// list DTO with server-computed capabilities. It reuses the shared
// CaseTaskService for assignee/manager checks so permission rules
// cannot drift between list rendering and write validation.
func (s *CaseTaskChangeService) ListTasksWithCapabilities(actorID uint, caseID uint) ([]TaskListItem, error) {
	rows := []models.CaseTask{}
	if err := s.tasks.db.Where("case_id = ?", caseID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("读取案件任务失败：%w", err)
	}
	items := make([]TaskListItem, 0, len(rows))
	for i := range rows {
		task := &rows[i]
		var managerOK, assigneeOK bool
		err := s.tasks.db.Transaction(func(tx *gorm.DB) error {
			managerOK = s.tasks.caseTaskCaseManagerAllowed(tx, task.CaseID, actorID)
			if actorID == task.AssigneeID {
				assigneeOK = s.tasks.caseTaskAssigneeAllowed(tx, task.CaseID, actorID)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		creatorLive := actorID == task.CreatedBy && s.canViewChanges(actorID, task)
		items = append(items, TaskListItem{
			CaseTask: *task,
			Capabilities: TaskCapabilities{
				CanAdjust:      creatorLive || managerOK,
				CanCancel:      creatorLive || managerOK,
				CanComplete:    assigneeOK || managerOK,
				CanViewChanges: s.canViewChanges(actorID, task),
			},
		})
	}
	return items, nil
}

// marshalChangeView is a test helper; it exists so the encoding/json
// import stays exercised in non-test builds too.
func marshalChangeView(v interface{}) string {
	payload, _ := json.Marshal(v)
	return string(payload)
}

var _ = marshalChangeView
