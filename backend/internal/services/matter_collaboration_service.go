package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"law-oa-go/internal/models"
)

// materialSnapshotFingerprint is the verifiable reference frozen into the
// intake_submission snapshot: status plus per-file digests. A later
// upload/delete cannot rewrite an old round, and tampering with stored
// bytes is detectable by re-hashing against this fingerprint.
type materialSnapshotFingerprint struct {
	MaterialID    uint                   `json:"material_id"`
	Name          string                 `json:"name,omitempty"`
	MaterialType  string                 `json:"material_type,omitempty"`
	Required      bool                   `json:"required,omitempty"`
	Status        string                 `json:"status"`
	CurrentFileID uint                   `json:"current_file_id,omitempty"`
	Files         []materialSnapshotFile `json:"files,omitempty"`
}

type materialSnapshotFile struct {
	FileID    uint   `json:"file_id"`
	Digest    string `json:"digest"`
	FileName  string `json:"file_name"`
	SizeBytes int64  `json:"size_bytes"`
}

// freezeMaterialFingerprintsForSnapshot reads the material ledger rows
// bound to an intake (including the case created from it) inside the
// caller transaction and returns a stable fingerprint list. The result is
// stored under the "material_fingerprints" key of intake_submission
// snapshots (approval_resubmit_contract.go). Legacy test schemas without
// the tables stay slim: empty list, no error. A snapshot row without this
// key means a legacy round; readers must not invent a pass from absence.
func freezeMaterialFingerprintsForSnapshot(tx *gorm.DB, intakeID string) ([]materialSnapshotFingerprint, error) {
	if !tx.Migrator().HasTable("material_files") {
		return []materialSnapshotFingerprint{}, nil
	}
	rows := []models.CaseMaterial{}
	// Materials are tracked by their real intake_id column; the case row
	// created from the intake keeps the same intake_id on every material
	// (case-creation adoption copies intake_id -> case_id without touching
	// this column), so one predicate covers both stages without inventing
	// any cases column.
	if err := tx.Raw("SELECT * FROM case_materials WHERE intake_id = ? ORDER BY id ASC", intakeID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("读取材料清单失败：%w", err)
	}
	out := make([]materialSnapshotFingerprint, 0, len(rows))
	for _, row := range rows {
		entry := materialSnapshotFingerprint{
			MaterialID:   row.ID,
			Name:         row.Name,
			MaterialType: row.MaterialType,
			Required:     row.Required,
			Status:       row.Status,
		}
		files := []models.MaterialFile{}
		if err := tx.Where("material_id = ? AND deleted_at IS NULL", row.ID).Order("id ASC").Find(&files).Error; err != nil {
			return nil, fmt.Errorf("读取材料附件失败：%w", err)
		}
		// The current pointer is case_materials.storage_url, the same
		// canonical download URL the List enrichment resolves. An explicit
		// pointer must name a live file of THIS row or the snapshot freezes
		// a broken pointer (zero) instead of guessing the newest file;
		// only a completely blank pointer keeps the legacy newest-live-file
		// fallback so old rounds stay comparable without inventing state.
		for _, f := range files {
			entry.Files = append(entry.Files, materialSnapshotFile{FileID: f.ID, Digest: f.ContentSHA256, FileName: f.FileName, SizeBytes: f.SizeBytes})
		}
		if fileID, pointed, _ := resolveMaterialCurrentFilePointer(&row, files); pointed {
			entry.CurrentFileID = fileID
		}
		out = append(out, entry)
	}
	return out, nil
}

// ValidateSupplementMaterialItems checks a request_changes payload that
// contains material_request items: every material_id must exist and must
// belong to the approval bound intake (or the case created from it).
// Cross-application references are rejected at decision time, before the
// item ever lands in ImposedRequirements.
func ValidateSupplementMaterialItems(db *gorm.DB, approvalID string, raw map[string]interface{}) error {
	if db == nil {
		return fmt.Errorf("材料归属校验不可用")
	}
	items, _ := raw["items"].([]interface{})
	if len(items) == 0 {
		return nil
	}
	var metadataRow struct {
		Metadata string
	}
	if err := db.Table("approval_requests").Select("metadata").Where("id = ?", approvalID).Take(&metadataRow).Error; err != nil {
		return fmt.Errorf("读取审批失败：%w", err)
	}
	metadata := map[string]interface{}{}
	if strings.TrimSpace(metadataRow.Metadata) != "" {
		if err := json.Unmarshal([]byte(metadataRow.Metadata), &metadata); err != nil {
			return fmt.Errorf("解析审批元数据失败：%w", err)
		}
	}
	intakeID, _ := metadata["intake_id"].(string)
	intakeID = strings.TrimSpace(intakeID)
	if intakeID == "" {
		for i, rawItem := range items {
			obj, ok := rawItem.(map[string]interface{})
			if !ok {
				continue
			}
			if cat, _ := obj["category"].(string); cat == "material_request" {
				return fmt.Errorf("第 %d 项 material_request 只能用于接案材料", i+1)
			}
		}
		return nil
	}
	referenced := map[string]bool{}
	for i, rawItem := range items {
		obj, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		cat, _ := obj["category"].(string)
		if cat != "material_request" {
			continue
		}
		mid := ""
		switch v := obj["material_id"].(type) {
		case string:
			mid = strings.TrimSpace(v)
		case float64:
			mid = fmt.Sprint(int64(v))
		case int:
			mid = fmt.Sprint(v)
		case uint:
			mid = fmt.Sprint(v)
		}
		if mid == "" {
			return fmt.Errorf("第 %d 项 material_request 缺少 material_id", i+1)
		}
		referenced[mid] = true
	}
	if len(referenced) == 0 {
		return nil
	}
	rows := []struct {
		ID uint
	}{}
	// Same real-column contract as the snapshot freeze: the approval
	// metadata maps approval -> intake_id, and the material rows carry
	// that intake_id for their whole life (intake stage and adopted case
	// stage alike). No cases.created_from_intake_id is assumed.
	if err := db.Table("case_materials").Select("id").
		Where("intake_id = ?", intakeID).
		Find(&rows).Error; err != nil {
		return fmt.Errorf("读取接案材料失败：%w", err)
	}
	valid := map[string]bool{}
	for _, r := range rows {
		valid[fmt.Sprint(r.ID)] = true
	}
	for mid := range referenced {
		if !valid[mid] {
			return fmt.Errorf("material_id %s 不属于该审批的接案材料", mid)
		}
	}
	return nil
}

// CaseTaskCreateRequest is the server contract for creating a case task.
type CaseTaskCreateRequest struct {
	CaseID     uint       `json:"case_id"`
	Title      string     `json:"title"`
	Detail     string     `json:"detail,omitempty"`
	AssigneeID uint       `json:"assignee_id"`
	DueDate    *time.Time `json:"due_date,omitempty"`
	// IdempotencyKey is optional. When the client omits it, the server
	// mints a unique per-request key, so keyless retries are NOT deduped
	// (deliberate) and no empty-string row can collide under the unique
	// index. Uniqueness is scoped per creator: (created_by, key).
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// caseTaskCompletionNoteMaxLen bounds one completion record. 2000 runes
// mirrors the supplement-reply plain-text cap, keeping every free-text
// legal record in the system under one reviewed size.
// CaseTaskCompletionNoteMaxLen is exported for handler tests asserting
// the over-length rejection path.
const CaseTaskCompletionNoteMaxLen = 2000

// dueDateFingerprint renders the due date for byte-stable comparison.
func dueDateFingerprint(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// idempotencyFingerprint covers every mutable input of one create request:
// case, title, detail, assignee, and due date. The same key with any
// differing field is a conflict, never a silent reuse (review 2026-09-30:
// the comparison must include creator/detail/due_date, not only
// case/title/assignee).
func (r *CaseTaskCreateRequest) idempotencyFingerprint() string {
	payload, _ := json.Marshal(map[string]interface{}{
		"case_id":     r.CaseID,
		"title":       strings.TrimSpace(r.Title),
		"detail":      r.Detail,
		"assignee_id": r.AssigneeID,
		"due_date":    dueDateFingerprint(r.DueDate),
	})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// CaseTaskService owns creation, completion, and todo mirroring for the
// general case tasks (item 5). Completion must flip the task, its todo,
// and the workflow completion-mode inside ONE transaction; the todo can
// never be marked done manually from the inbox.
type CaseTaskService struct {
	db *gorm.DB
	// authz reuses the shared object-level authorization for assignee
	// validation: a candidate must be someone who can actually read the
	// case, so deleted/deactivated users and ethical-wall outsiders are
	// rejected by the same code path as every other object check.
	authz *AuthorizationService
}

func NewCaseTaskService(db *gorm.DB) *CaseTaskService { return &CaseTaskService{db: db} }

// NewCaseTaskServiceWithAuthz is the production constructor: assignee
// validation runs through AuthorizationService.CanReadCase.
func NewCaseTaskServiceWithAuthz(db *gorm.DB, authz *AuthorizationService) *CaseTaskService {
	return &CaseTaskService{db: db, authz: authz}
}

// caseTaskWorkflowKey builds the deterministic key the inbox projection
// already understands (case_task:<id> prefix).
func caseTaskWorkflowKey(taskID uint) string {
	return "case_task:" + fmt.Sprint(taskID)
}

// CreateTask validates assignee legality and writes the task plus its
// mirrored todo atomically. IdempotencyKey dedups retries; a duplicate
// key with a different payload returns ErrCaseTaskIdempotencyConflict.
func (s *CaseTaskService) CreateTask(creatorID uint, req *CaseTaskCreateRequest) (*models.CaseTask, error) {
	if req == nil || req.CaseID == 0 || strings.TrimSpace(req.Title) == "" || req.AssigneeID == 0 {
		return nil, fmt.Errorf("案件任务缺少必要字段")
	}
	fingerprint := req.idempotencyFingerprint()
	trimmedKey := strings.TrimSpace(req.IdempotencyKey)
	if trimmedKey == "" {
		// Keyless requests get a unique server key: no two keyless creates
		// collide on the unique (created_by, idempotency_key) index, and a
		// keyless retry intentionally creates a second task instead of
		// silently replaying the first.
		trimmedKey = "auto-" + generateUUID()
	}
	var out *models.CaseTask
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var caseRow struct {
			ID uint
		}
		if err := tx.Table("cases").Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id = ? AND deleted_at IS NULL", req.CaseID).Take(&caseRow).Error; err != nil {
			return fmt.Errorf("案件不存在或已删除：%w", err)
		}
		if !s.caseTaskAssigneeAllowed(tx, req.CaseID, req.AssigneeID) {
			return ErrCaseTaskAssigneeInvalid
		}
		var existing models.CaseTask
		err := tx.Where("created_by = ? AND idempotency_key = ?", creatorID, trimmedKey).Take(&existing).Error
		if err == nil {
			if existing.IdempotencyFingerprint != fingerprint {
				return ErrCaseTaskIdempotencyConflict
			}
			out = &existing
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		task := &models.CaseTask{
			CaseID: req.CaseID, Title: strings.TrimSpace(req.Title), Detail: req.Detail,
			AssigneeID: req.AssigneeID, CreatedBy: creatorID, DueDate: req.DueDate,
			Status: models.CaseTaskStatusOpen, IdempotencyKey: trimmedKey,
			IdempotencyFingerprint: fingerprint, Version: 1,
		}
		if err := tx.Create(task).Error; err != nil {
			return fmt.Errorf("创建案件任务失败：%w", err)
		}
		if err := openCaseTaskInboxTx(tx, task); err != nil {
			return err
		}
		out = task
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CompleteTask closes the task and its workflow todo in one transaction.
// Only the assignee (or a case manager) may complete. Even an already
// completed task re-runs the actor authorization before the idempotent
// nil return, so a revoked outsider cannot harvest success from a replay
// (review 2026-09-30). The mirrored todo carries completion_mode=workflow
// so the inbox 409 manual-complete guard keeps protecting it.
func (s *CaseTaskService) CompleteTask(actorID uint, taskID uint, completionNote string) (*models.CaseTask, bool, error) {
	// Bounds-check the note BEFORE the transaction opens: the response
	// claims completion_note_recorded only when the text actually landed
	// in the same commit as the status flip (review 2026-09-30, defect 4).
	note := strings.TrimSpace(completionNote)
	if len([]rune(note)) > CaseTaskCompletionNoteMaxLen {
		return nil, false, ErrCaseTaskNoteTooLong
	}
	replayed := false
	var completed *models.CaseTask
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		var task models.CaseTask
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", taskID).Take(&task).Error; err != nil {
			return fmt.Errorf("任务不存在：%w", err)
		}
		// Every path - first completion AND completed replay - re-validates
		// the CURRENT actor: active user row + live case authorization.
		// A later wall/team removal or deactivation therefore denies even
		// the original assignee's replay, and no assignee shortcut can
		// bypass the manager-level case-scope check (review 2026-09-30).
		if !s.caseTaskManagerAllowed(tx, task.CaseID, actorID, task.AssigneeID) {
			return ErrCaseTaskForbidden
		}
		if task.Status == models.CaseTaskStatusCompleted {
			replayed = true
			completed = &task
			return nil
		}
		if task.Status == models.CaseTaskStatusCancelled {
			return ErrCaseTaskCancelled
		}
		now := time.Now()
		// CAS on version: the UPDATE only lands when the row the actor read
		// is still current, so a concurrent cancel/reassign cannot be
		// overwritten by a late completion (and vice versa). RowsAffected==0
		// means the task moved underneath us; fail closed with the stale
		// sentinel instead of silently second-writing.
		res := tx.Model(&models.CaseTask{}).
			Where("id = ? AND version = ?", taskID, task.Version).
			Updates(map[string]interface{}{
				"status": models.CaseTaskStatusCompleted, "completed_by": actorID, "completed_at": now,
				"completion_note": note, "updated_at": now, "version": task.Version + 1,
			})
		if res.Error != nil {
			return fmt.Errorf("完成任务失败：%w", res.Error)
		}
		if res.RowsAffected == 0 {
			return ErrCaseTaskChangeStale
		}
		if err := tx.Model(&models.InboxItem{}).
			Where("workflow_task_key = ? AND completion_mode = ? AND is_completed = ?", caseTaskWorkflowKey(taskID), "workflow", false).
			Updates(map[string]interface{}{
				"is_completed": true, "completed_at": now, "updated_at": now,
			}).Error; err != nil {
			return fmt.Errorf("关闭工作流待办失败：%w", err)
		}
		// Sync the in-memory struct with what actually landed in the DB
		// inside this same transaction: tx.Updates(map) does not mutate
		// the Go struct, so handing the caller the stale copy made the
		// first completion response report completion_note_recorded=false
		// while the note was in fact persisted (review 2026-10-01).
		completed = &task
		completed.Status = models.CaseTaskStatusCompleted
		completedBy := actorID
		completed.CompletedBy = &completedBy
		completed.CompletedAt = &now
		completed.CompletionNote = note
		completed.UpdatedAt = now
		return nil
	})
	if txErr != nil {
		return nil, false, txErr
	}
	return completed, replayed, nil
}

// caseTaskAssigneeAllowed reports whether userID may hold or complete a
// task inside caseID. When the service carries an AuthorizationService it
// delegates to CanReadCase: that single call covers deleted/inactive
// users, ownership, management roles, and the ethical-wall whitelist, so
// there is no second rule set to drift (review 2026-09-30). Without authz
// (unit tests only) it falls back to the same ownership/whitelist SQL the
// inbox projection uses, always requiring an active user row.
func (s *CaseTaskService) caseTaskAssigneeAllowed(tx *gorm.DB, caseID uint, userID uint) bool {
	if userID == 0 || caseID == 0 {
		return false
	}
	if s.authz != nil {
		actor := AuthActor{UserID: userID}
		var roleRow struct {
			Role string
		}
		if err := tx.Table("users").Select("role").Where("id = ? AND deleted_at IS NULL AND status = 'active'", userID).Take(&roleRow).Error; err != nil {
			return false
		}
		actor.Role = roleRow.Role
		allowed, err := s.authz.CanReadCase(context.Background(), actor, caseID)
		return err == nil && allowed
	}
	// Fallback: active user + ownership/whitelist SQL (same shape as
	// inbox prefetchCaseAccess).
	var count int64
	err := tx.Table("cases").Where("id = ? AND deleted_at IS NULL AND (lawyer_id = ? OR created_by = ? OR EXISTS (SELECT 1 FROM case_ethical_wall_whitelist w WHERE w.case_id = cases.id AND w.user_id = ?))", caseID, userID, fmt.Sprint(userID), userID).Count(&count).Error
	if err != nil {
		return false
	}
	if count > 0 {
		return true
	}
	var user models.User
	if err := tx.Where("id = ? AND deleted_at IS NULL AND status = 'active'", userID).Take(&user).Error; err != nil {
		return false
	}
	return IsBusinessMatterManagementRole(user.Role)
}

// caseTaskManagerAllowed reports whether userID may complete someone
// else's task inside caseID: either the assignee themself or a genuine
// matter manager (owner / business-management role behind the ethical
// wall). Plain read-team membership is NOT management (review
// 2026-10-01: the previous delegation to caseTaskAssigneeAllowed let any
// reader finish other people's tasks).
func (s *CaseTaskService) caseTaskManagerAllowed(tx *gorm.DB, caseID uint, userID uint, assigneeID uint) bool {
	if userID == 0 || caseID == 0 {
		return false
	}
	if assigneeID != 0 && userID == assigneeID {
		return s.caseTaskAssigneeAllowed(tx, caseID, userID)
	}
	return s.caseTaskCaseManagerAllowed(tx, caseID, userID)
}

// caseTaskCaseManagerAllowed reports true management capability without
// the assignee shortcut, used for completing other people's tasks.
func (s *CaseTaskService) caseTaskCaseManagerAllowed(tx *gorm.DB, caseID uint, userID uint) bool {
	if s.authz != nil {
		actor := AuthActor{UserID: userID}
		var roleRow struct {
			Role string
		}
		if err := tx.Table("users").Select("role").Where("id = ? AND deleted_at IS NULL AND status = 'active'", userID).Take(&roleRow).Error; err != nil {
			return false
		}
		actor.Role = roleRow.Role
		allowed, err := s.authz.CanManageCase(context.Background(), actor, caseID)
		return err == nil && allowed
	}
	// Fallback mirrors the read fallback's ownership SQL but demands
	// ownership/management rather than mere whitelist visibility.
	var count int64
	err := tx.Table("cases").Where("id = ? AND deleted_at IS NULL AND (lawyer_id = ? OR created_by = ?)", caseID, userID, fmt.Sprint(userID)).Count(&count).Error
	if err != nil {
		return false
	}
	if count > 0 {
		return true
	}
	var user models.User
	if err := tx.Where("id = ? AND deleted_at IS NULL AND status = 'active'", userID).Take(&user).Error; err != nil {
		return false
	}
	return IsBusinessMatterManagementRole(user.Role)
}

// openCaseTaskInboxTx writes the assignee todo inside the creator
// transaction, with the workflow completion mode and deterministic key.
func openCaseTaskInboxTx(tx *gorm.DB, task *models.CaseTask) error {
	now := time.Now()
	item := &models.InboxItem{
		UserID: task.AssigneeID, SourceType: "task", SourceID: task.ID,
		SourceRef: fmt.Sprint(task.ID), CompletionMode: "workflow",
		WorkflowTaskKey: caseTaskWorkflowKey(task.ID),
		Title:           "案件任务：" + task.Title,
		Content:         task.Detail,
		Priority:        "medium",
		DueDate:         task.DueDate,
	}
	item.CreatedAt = now
	item.UpdatedAt = now
	return tx.Create(item).Error
}

// Sentinel errors, exported for handler HTTP mapping.
var (
	ErrCaseTaskAssigneeInvalid     = errors.New("CASE_TASK_ASSIGNEE_INVALID: 任务负责人不在案件内合法人员范围")
	ErrCaseTaskIdempotencyConflict = errors.New("CASE_TASK_IDEMPOTENCY_CONFLICT: 同一幂等键对应不同任务内容")
	ErrCaseTaskCancelled           = errors.New("CASE_TASK_CANCELLED: 任务已取消")
	ErrCaseTaskForbidden           = errors.New("CASE_TASK_FORBIDDEN: 只有任务负责人或案件管理者可以完成任务")
	// ErrCaseTaskNoteTooLong guards the completion record cap; the HTTP
	// layer maps it to 400 because the whole transaction refused to open.
	ErrCaseTaskNoteTooLong = errors.New("CASE_TASK_NOTE_TOO_LONG: 完成记录超过长度限制")
)

// ValidateSupplementMaterialUploadsTx is the resubmit-time gate for
// material_request items (review 2026-09-30, defect 7): ownership alone
// is not enough. Inside the caller's transaction every referenced
// material must (a) belong to the approval's bound intake via the real
// intake_id column, and (b) actually carry an uploaded status plus at
// least one live attachment row, so a text-only reply can never satisfy
// a material return. Fail-closed: once a material_request is present the
// backing tables MUST exist; a slim fixture without them is an unfinished
// schema, not a license to wave the request through (review 2026-10-01).
// Only the len(referenced)==0 text-only legacy contract returns nil early.
func ValidateSupplementMaterialUploadsTx(tx *gorm.DB, intakeID string, items []SupplementItem) error {
	if tx == nil || strings.TrimSpace(intakeID) == "" {
		return nil
	}
	referenced := map[uint]bool{}
	for _, it := range items {
		if it.Category == "material_request" && it.MaterialID > 0 {
			referenced[it.MaterialID] = true
		}
	}
	if len(referenced) == 0 {
		return nil
	}
	if !tx.Migrator().HasTable("case_materials") || !tx.Migrator().HasTable("material_files") {
		return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: "审批退回要求补充材料，但接案材料表尚未就绪，已阻止重提（fail-closed）"}
	}
	ids := make([]uint, 0, len(referenced))
	for id := range referenced {
		ids = append(ids, id)
	}
	rows := []struct {
		ID        uint
		Status    string
		FileCount int64
	}{}
	if err := tx.Table("case_materials").
		Select("case_materials.id, case_materials.status, (SELECT COUNT(*) FROM material_files f WHERE f.material_id = case_materials.id AND f.deleted_at IS NULL) AS file_count").
		Where("case_materials.intake_id = ? AND case_materials.id IN ?", intakeID, ids).
		Scan(&rows).Error; err != nil {
		return fmt.Errorf("读取接案材料状态失败：%w", err)
	}
	uploaded := map[uint]bool{}
	for _, r := range rows {
		if r.Status == models.MaterialStatusUploaded && r.FileCount > 0 {
			uploaded[r.ID] = true
		}
	}
	for id := range referenced {
		if !uploaded[id] {
			return &ResubmitContractError{Code: models.ErrCodeSupplementItemInvalid, Message: fmt.Sprintf("材料 %d 尚未实际上传完成，不能以纯文本回复完成重提", id)}
		}
	}
	return nil
}

// sha256Digest is a small helper reused by tests.
func sha256Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
