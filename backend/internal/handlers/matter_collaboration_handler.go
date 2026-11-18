package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
	"net/http"
)

// MatterCollaborationHandler serves the material ledger (item 4) and the
// general case tasks (item 5). Every route entry runs an object-level
// permission check before any read or write.
type MatterCollaborationHandler struct {
	guard *materialAccessGuard
	db    *gorm.DB
	// StorageDir is the root directory for uploaded material files. It
	// defaults to ./uploads/materials relative to the working directory.
	StorageDir string
}

func NewMatterCollaborationHandler(authz *services.AuthorizationService, db *gorm.DB) *MatterCollaborationHandler {
	return &MatterCollaborationHandler{
		guard: newMaterialAccessGuard(authz, db),
		db:    db,
	}
}

// NewMatterCollaborationHandlerWithDB builds the handler plus its shared
// authorization service from one gorm handle (router integration shape).
func NewMatterCollaborationHandlerWithDB(db *gorm.DB) *MatterCollaborationHandler {
	return NewMatterCollaborationHandler(services.NewAuthorizationServiceWithDB(db), db)
}

// ---- material ledger ----

// materialUploadLimit caps one attachment at 50 MiB. The reader stops at
// limit+1 bytes so an oversized body is detected by size, not by memory
// pressure; the multipart open is also bounded before any buffering.
const materialUploadLimit int64 = 50 << 20

// materialExtensionWhitelist is the allow-list of safe storage
// extensions. Executable/script formats never reach disk; the download
// handler serves them with attachment disposition only, and the stored
// name is always the server-side digest (never the client filename).
var materialExtensionWhitelist = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".webp": true, ".txt": true, ".csv": true, ".zip": true,
}

// validateMaterialUpload enforces the pre-storage checks: allowed
// extension, sane content type, and a non-empty file.
func validateMaterialUpload(c *gin.Context, originalName, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(originalName))
	if !materialExtensionWhitelist[ext] {
		common.APIBadRequest(c, "文件类型不支持", "仅支持文档/图片/压缩包等安全格式")
		return false
	}
	if strings.HasPrefix(strings.ToLower(contentType), "application/x-httpd-php") {
		common.APIBadRequest(c, "文件类型不支持", "脚本类型文件被拒绝")
		return false
	}
	return true
}

// ---- intake-scope endpoints (UUID id) ----

// createMaterialRequest is the create body shared by the intake and case
// material list endpoints.
type createMaterialRequest struct {
	Name         *string `json:"name"`
	MaterialType *string `json:"material_type"`
	Required     *bool   `json:"required"`
	OwnerID      *uint   `json:"owner_id"`
}

// ListIntakeMaterials returns the ledger of one intake (UUID id). Read is
// the intake boundary: the creator plus actors directly bound to the
// intake's approval/review workflow; an unrelated lawyer (even one with a
// global business role) gets 403.
func (h *MatterCollaborationHandler) ListIntakeMaterials(c *gin.Context) {
	intakeID := strings.TrimSpace(c.Param("id"))
	if intakeID == "" {
		common.APIBadRequest(c, "请求参数错误", "id 不能为空")
		return
	}
	if !h.guard.authorizeIntakeMaterial(c, intakeID, false) {
		return
	}
	rows := []models.CaseMaterial{}
	if err := h.db.Where("intake_id = ?", intakeID).Order("id ASC").Find(&rows).Error; err != nil {
		common.APIInternalServerError(c, "读取接案材料失败", err.Error())
		return
	}
	items, err := h.enrichMaterialList(rows)
	if err != nil {
		common.APIInternalServerError(c, "读取接案材料附件失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"materials": items})
}

// CreateIntakeMaterial adds one ledger row to an intake. Only the intake
// creator may write; status always starts at missing (uploaded is only
// reachable through the upload transaction).
func (h *MatterCollaborationHandler) CreateIntakeMaterial(c *gin.Context) {
	intakeID := strings.TrimSpace(c.Param("id"))
	if intakeID == "" {
		common.APIBadRequest(c, "请求参数错误", "id 不能为空")
		return
	}
	var req createMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		common.APIBadRequest(c, "材料名称必填", "name 不能为空")
		return
	}
	if !h.guard.authorizeIntakeMaterial(c, intakeID, true) {
		return
	}
	row := &models.CaseMaterial{
		IntakeID: &intakeID,
		Name:     strings.TrimSpace(*req.Name),
		Status:   models.MaterialStatusMissing,
		Required: req.Required == nil || *req.Required,
	}
	if req.MaterialType != nil {
		row.MaterialType = strings.TrimSpace(*req.MaterialType)
	}
	if req.OwnerID != nil && *req.OwnerID > 0 {
		// Intake rows authorize the responsible person inside the intake
		// collaboration boundary, never by arbitrary id (review
		// 2026-09-30, defect 2). The writer is already the intake creator
		// (write=creator-only), so this check binds the assignee itself.
		if !h.intakeOwnerAllowed(intakeID, *req.OwnerID) {
			common.APIForbidden(c, "责任人不在接案协作范围", "材料责任人必须是接案创建人或已在本接案承担材料责任的人员")
			return
		}
		row.OwnerID = req.OwnerID
	}
	if err := h.db.Create(row).Error; err != nil {
		common.APIInternalServerError(c, "创建接案材料失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"material": row})
}

// CreateCaseMaterial adds one ledger row directly inside an adopted case.
func (h *MatterCollaborationHandler) CreateCaseMaterial(c *gin.Context) {
	caseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req createMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		common.APIBadRequest(c, "材料名称必填", "name 不能为空")
		return
	}
	if !h.guard.requireCaseManageable(c, caseID) {
		return
	}
	row := &models.CaseMaterial{
		CaseID:   &caseID,
		Name:     strings.TrimSpace(*req.Name),
		Status:   models.MaterialStatusMissing,
		Required: req.Required == nil || *req.Required,
	}
	if req.MaterialType != nil {
		row.MaterialType = strings.TrimSpace(*req.MaterialType)
	}
	if req.OwnerID != nil && *req.OwnerID > 0 {
		if !h.assigneeAllowed(caseID, *req.OwnerID) {
			common.APIForbidden(c, "责任人不在案件团队", "材料责任人必须是案件内合法人员")
			return
		}
		row.OwnerID = req.OwnerID
	}
	if err := h.db.Create(row).Error; err != nil {
		common.APIInternalServerError(c, "创建案件材料失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"material": row})
}

// materialFileView is the response-only view of one stored attachment.
// It deliberately omits storage_path: the physical location is internal,
// and every byte transfer goes through the permission-checked download
// endpoint.
type materialFileView struct {
	ID          uint   `json:"id"`
	FileName    string `json:"file_name"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type,omitempty"`
	// ContentSHA256 lets the client cross-check a version against the
	// frozen intake snapshot digests without exposing any storage path.
	ContentSHA256 string    `json:"content_sha256,omitempty"`
	UploadedBy    uint      `json:"uploaded_by"`
	UploadedAt    time.Time `json:"uploaded_at"`
	// Version is the 1-based position in upload order (ascending ID),
	// so the newest upload is always the highest version of its row.
	Version     int    `json:"version"`
	DownloadURL string `json:"download_url"`
}

// materialListItem is one ledger row enriched with the current attachment
// metadata, so the client can show 文件名/上传者/时间/当前版本 without an
// extra request per row. Embedded fields stay flat in the JSON contract.
type materialListItem struct {
	models.CaseMaterial
	FileCount   int               `json:"file_count"`
	FileVersion int               `json:"file_version"`
	CurrentFile *materialFileView `json:"current_file,omitempty"`
}

// materialFileDownloadURL builds the canonical download path for one
// attachment; the upload transaction stores exactly this string in
// storage_url, so list enrichment matches on it instead of guessing.
func materialFileDownloadURL(materialID, fileID uint) string {
	return "/api/v1/case-materials/" + strconv.FormatUint(uint64(materialID), 10) +
		"/files/" + strconv.FormatUint(uint64(fileID), 10) + "/download"
}

// enrichMaterialList attaches per-row attachment metadata using ONE query
// for all listed materials. The current version is the file referenced by
// storage_url; only a missing pointer falls back to the highest file ID
// (upload order). An explicit invalid pointer leaves current_file absent
// and file_version zero. Rows are copied, never mutated in place.
func (h *MatterCollaborationHandler) enrichMaterialList(rows []models.CaseMaterial) ([]materialListItem, error) {
	items := make([]materialListItem, len(rows))
	for i := range rows {
		items[i] = materialListItem{CaseMaterial: rows[i]}
	}
	if len(rows) == 0 {
		return items, nil
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	files := []models.MaterialFile{}
	if err := h.db.Where("material_id IN ?", ids).Order("material_id ASC, id ASC").Find(&files).Error; err != nil {
		return nil, err
	}
	byMaterial := map[uint][]models.MaterialFile{}
	for _, f := range files {
		byMaterial[f.MaterialID] = append(byMaterial[f.MaterialID], f)
	}
	for i := range items {
		row := &items[i]
		rowFiles := byMaterial[row.ID]
		row.FileCount = len(rowFiles)
		if len(rowFiles) == 0 {
			continue
		}
		currentIdx := -1
		if strings.TrimSpace(row.StorageURL) == "" {
			currentIdx = len(rowFiles) - 1
		}
		views := make([]materialFileView, 0, len(rowFiles))
		for idx, f := range rowFiles {
			url := materialFileDownloadURL(row.ID, f.ID)
			if url == row.StorageURL {
				currentIdx = idx
			}
			views = append(views, materialFileView{
				ID: f.ID, FileName: f.FileName, SizeBytes: f.SizeBytes,
				ContentType: f.ContentType, ContentSHA256: f.ContentSHA256,
				UploadedBy: f.UploadedBy, UploadedAt: f.CreatedAt,
				Version: idx + 1, DownloadURL: url,
			})
		}
		if currentIdx < 0 {
			continue
		}
		current := views[currentIdx]
		row.CurrentFile = &current
		row.FileVersion = current.Version
	}
	return items, nil
}

// ---- case task endpoints ----

// completeTaskRequest is the body of POST /case-tasks/:taskId/complete.
type completeTaskRequest struct {
	// CompletionNote is a free-text completion record; it never feeds a
	// statutory deadline computation.
	CompletionNote string `json:"completion_note"`
}

// ListCaseTasks returns the tasks of one case (numeric id), readable by
// everyone who can read the case. Each row carries the server-computed
// capability flags (can_adjust / can_cancel / can_complete /
// can_view_changes) plus the optimistic-lock version, so the frontend
// renders buttons from server truth instead of re-deriving permissions
// from created_by/assignee_id (review 2026-10-02). The server still
// re-validates every write regardless of what the client renders.
func (h *MatterCollaborationHandler) ListCaseTasks(c *gin.Context) {
	caseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if !h.guard.requireCaseReadable(c, caseID) {
		return
	}
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	items, err := h.caseTaskChangeService().ListTasksWithCapabilities(actor.UserID, caseID)
	if err != nil {
		common.APIInternalServerError(c, "读取案件任务失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"tasks": items})
}

// CreateCaseTask creates one general matter task and mirrors its todo
// atomically. The creator must be able to read the case; the assignee must
// pass the assignee legality check (which itself re-runs object-level
// authorization, so deleted/inactive/walled users fail closed).
func (h *MatterCollaborationHandler) CreateCaseTask(c *gin.Context) {
	caseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if !h.guard.requireCaseReadable(c, caseID) {
		return
	}
	var req services.CaseTaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}
	// The path is authoritative: a body case_id cannot point the task at
	// another case that the actor happens to be allowed to read.
	req.CaseID = caseID
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	svc := h.caseTaskService()
	task, err := svc.CreateTask(actor.UserID, &req)
	if err != nil {
		mapCaseTaskError(c, err)
		return
	}
	common.APISuccess(c, gin.H{"task": task})
}

// CompleteCaseTask closes one task plus its workflow todo in one
// transaction; the assignee re-authorization runs before any idempotent
// short-circuit.
func (h *MatterCollaborationHandler) CompleteCaseTask(c *gin.Context) {
	taskID, ok := parseUintParam(c, "taskId")
	if !ok {
		return
	}
	var req completeTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	svc := h.caseTaskService()
	task, replayed, err := svc.CompleteTask(actor.UserID, taskID, req.CompletionNote)
	if err != nil {
		mapCaseTaskError(c, err)
		return
	}
	// Report what actually landed: a replay keeps the task's original
	// note, so completion_note_recorded must reflect the stored column,
	// never the request payload alone.
	common.APISuccess(c, gin.H{
		"ok":                       true,
		"replayed":                 replayed,
		"completion_note_recorded": strings.TrimSpace(task.CompletionNote) != "",
		"completion_note":          task.CompletionNote,
	})
}

func (h *MatterCollaborationHandler) caseTaskService() *services.CaseTaskService {
	return services.NewCaseTaskServiceWithAuthz(h.db, h.guard.authz)
}

// caseTaskChangeService builds the change service on the shared authz
// handle; both services must see the same permission source.
func (h *MatterCollaborationHandler) caseTaskChangeService() *services.CaseTaskChangeService {
	return services.NewCaseTaskChangeServiceWithAuthz(h.db, h.guard.authz)
}

// mapCaseTaskError maps service sentinel errors to the HTTP contract.
func mapCaseTaskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrCaseTaskAssigneeInvalid):
		common.NewAPIError(c, 403, "CASE_TASK_ASSIGNEE_INVALID", "任务负责人必须是案件内合法人员")
	case errors.Is(err, services.ErrCaseTaskIdempotencyConflict):
		common.NewAPIError(c, 409, "CASE_TASK_IDEMPOTENCY_CONFLICT", "同一幂等键对应不同任务内容")
	case errors.Is(err, services.ErrCaseTaskCancelled):
		common.NewAPIError(c, 409, "CASE_TASK_CANCELLED", "任务已取消，不能完成")
	case errors.Is(err, services.ErrCaseTaskForbidden):
		common.NewAPIError(c, 403, "CASE_TASK_FORBIDDEN", "只有任务负责人或案件管理者可以完成任务")
	case errors.Is(err, services.ErrCaseTaskNoteTooLong):
		common.APIBadRequest(c, "完成记录过长", "completion_note 不能超过 2000 字")
	case errors.Is(err, services.ErrCaseTaskChangeStale):
		common.NewAPIError(c, 409, "CASE_TASK_CHANGE_STALE", "任务已被他人修改，请刷新后基于最新内容重试")
	default:
		if strings.Contains(err.Error(), "不存在") {
			common.APINotFound(c, "任务不存在")
			return
		}
		common.APIInternalServerError(c, "案件任务操作失败", err.Error())
	}
}

type updateMaterialRequest struct {
	OwnerID      *uint   `json:"owner_id"`
	Status       *string `json:"status"`
	Required     *bool   `json:"required"`
	Name         *string `json:"name"`
	MaterialType *string `json:"material_type"`
}

// ListCaseMaterials returns the material ledger rows of one case. The
// response hides storage internals; downloads go through the dedicated
// download endpoint so every byte transfer is permission-checked.
func (h *MatterCollaborationHandler) ListCaseMaterials(c *gin.Context) {
	caseID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if !h.guard.requireCaseReadable(c, caseID) {
		return
	}
	rows := []models.CaseMaterial{}
	if err := h.db.Where("case_id = ?", caseID).Order("id ASC").Find(&rows).Error; err != nil {
		common.APIInternalServerError(c, "读取材料清单失败", err.Error())
		return
	}
	items, err := h.enrichMaterialList(rows)
	if err != nil {
		common.APIInternalServerError(c, "读取材料附件失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"materials": items})
}

// ListCaseMaterialFiles returns every stored attachment of ONE material
// row, oldest first, each labeled with its 1-based version. The list
// endpoint's object-level permission rules apply unchanged: case rows
// need case read access, intake rows need the intake material boundary.
func (h *MatterCollaborationHandler) ListCaseMaterialFiles(c *gin.Context) {
	materialID, ok := parseUintParam(c, "materialId")
	if !ok {
		return
	}
	scope, ok := h.guard.resolveCaseScope(c, materialID)
	if !ok {
		return
	}
	if !h.authorizeMaterialScope(c, scope, false) {
		return
	}
	files := []models.MaterialFile{}
	if err := h.db.Where("material_id = ?", materialID).Order("id ASC").Find(&files).Error; err != nil {
		common.APIInternalServerError(c, "读取附件历史失败", err.Error())
		return
	}
	views := make([]materialFileView, 0, len(files))
	for i, f := range files {
		views = append(views, materialFileView{
			ID: f.ID, FileName: f.FileName, SizeBytes: f.SizeBytes,
			ContentType: f.ContentType, ContentSHA256: f.ContentSHA256,
			UploadedBy: f.UploadedBy, UploadedAt: f.CreatedAt,
			Version:     i + 1,
			DownloadURL: materialFileDownloadURL(materialID, f.ID),
		})
	}
	common.APISuccess(c, gin.H{"files": views})
}

// UpdateCaseMaterial edits ownership / required flag / non-upload status.
// status=uploaded is rejected here: it is only reachable through the
// upload transaction, so a client can never fake a complete material
// without a real file.
func (h *MatterCollaborationHandler) UpdateCaseMaterial(c *gin.Context) {
	materialID, ok := parseUintParam(c, "materialId")
	if !ok {
		return
	}
	var req updateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}
	scope, ok := h.guard.resolveCaseScope(c, materialID)
	if !ok {
		return
	}
	if !h.authorizeMaterialScope(c, scope, true) {
		return
	}
	updates := map[string]interface{}{}
	if req.OwnerID != nil {
		if *req.OwnerID == 0 {
			updates["owner_id"] = nil
		} else {
			// Intake rows (no case yet) authorize the responsible person
			// against the intake boundary: the intake creator plus anyone
			// already assigned as an owner on the same intake (review
			// 2026-09-30, defect 2). Case rows use the case-team check.
			if scope.CaseID > 0 {
				if !h.assigneeAllowed(scope.CaseID, *req.OwnerID) {
					common.APIForbidden(c, "责任人不在案件团队", "材料责任人必须是案件内合法人员")
					return
				}
			} else if !h.intakeOwnerAllowed(scope.IntakeID, *req.OwnerID) {
				common.APIForbidden(c, "责任人不在接案协作范围", "材料责任人必须是接案创建人或已在本接案承担材料责任的人员")
				return
			}
			updates["owner_id"] = *req.OwnerID
		}
	}
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			common.APIBadRequest(c, "材料名称无效", "name 不能为空")
			return
		}
		updates["name"] = trimmed
	}
	if req.MaterialType != nil {
		updates["material_type"] = strings.TrimSpace(*req.MaterialType)
	}
	if req.Required != nil {
		updates["required"] = *req.Required
	}
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st == models.MaterialStatusUploaded {
			common.APIBadRequest(c, "状态不允许直接设置", "uploaded 状态只能通过实际上传获得")
			return
		}
		if st != models.MaterialStatusMissing && st != models.MaterialStatusNotRequired {
			common.APIBadRequest(c, "状态值无效", "status 只支持 missing 或 not_required")
			return
		}
		updates["status"] = st
	}
	if len(updates) == 0 {
		common.APIBadRequest(c, "无可更新字段", "至少提供一个更新字段")
		return
	}
	updates["updated_at"] = time.Now()
	if err := h.db.Model(&models.CaseMaterial{}).Where("id = ?", materialID).Updates(updates).Error; err != nil {
		common.APIInternalServerError(c, "更新材料失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"ok": true})
}

// UploadCaseMaterial stores one physical file and flips the row to
// uploaded inside the same transaction that records the digest. The
// client never supplies the status.
func (h *MatterCollaborationHandler) UploadCaseMaterial(c *gin.Context) {
	// Bind the WHOLE request body to the upload cap before any multipart
	// parse: http.MaxBytesReader makes the transport itself reject a body
	// over the limit, not just the parsed header size (review 2026-09-30,
	// defect 6). Error pages stay small regardless.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, materialUploadLimit+(1<<20))
	materialID, ok := parseUintParam(c, "materialId")
	if !ok {
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		common.APIBadRequest(c, "缺少文件", "必须通过 multipart 字段 file 上传")
		return
	}
	scope, ok := h.guard.resolveCaseScope(c, materialID)
	if !ok {
		return
	}
	if !h.authorizeMaterialScope(c, scope, true) {
		return
	}
	if fileHeader.Size > materialUploadLimit {
		common.NewAPIError(c, 413, "MATERIAL_FILE_TOO_LARGE", "单个文件不能超过 50MB")
		return
	}
	if !validateMaterialUpload(c, fileHeader.Filename, fileHeader.Header.Get("Content-Type")) {
		return
	}
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	src, err := fileHeader.Open()
	if err != nil {
		common.APIBadRequest(c, "文件不可读", err.Error())
		return
	}
	defer src.Close()
	hasher := sha256.New()
	var buf []byte
	// Read at most limit+1 bytes: a body that still has content after the
	// cap fails closed instead of silently truncating a huge file.
	limited := io.LimitReader(src, materialUploadLimit+1)
	if _, err := io.Copy(hasher, io.TeeReader(limited, &byteWriter{&buf})); err != nil {
		common.APIInternalServerError(c, "读取文件失败", err.Error())
		return
	}
	if int64(len(buf)) > materialUploadLimit {
		common.NewAPIError(c, 413, "MATERIAL_FILE_TOO_LARGE", "单个文件不能超过 50MB")
		return
	}
	if len(buf) == 0 {
		common.APIBadRequest(c, "文件为空", "不能上传空文件")
		return
	}
	digest := hex.EncodeToString(hasher.Sum(nil))
	// One file per attempt: the digest prefix keeps dedup grep-able, but
	// the random suffix makes every upload a fresh inode, so a failed
	// commit below can safely delete ITS OWN file even when identical
	// bytes are already in use by another material row (review
	// 2026-09-30, defect 5).
	storagePath, createdFile := materialStoragePath(h.storageRoot(), digest, fileHeader.Filename)
	if err := os.WriteFile(storagePath, buf, 0o640); err != nil {
		common.APIInternalServerError(c, "保存文件失败", err.Error())
		return
	}
	now := time.Now()
	err = h.db.Transaction(func(tx *gorm.DB) error {
		row := &models.MaterialFile{
			MaterialID: materialID, ContentSHA256: digest, FileName: fileHeader.Filename,
			SizeBytes: int64(len(buf)), ContentType: fileHeader.Header.Get("Content-Type"),
			StoragePath: storagePath, UploadedBy: actor.UserID,
		}
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		return tx.Model(&models.CaseMaterial{}).Where("id = ?", materialID).Updates(map[string]interface{}{
			"status": models.MaterialStatusUploaded, "uploaded_by": actor.UserID, "uploaded_at": now,
			"storage_url": "/api/v1/case-materials/" + strconv.FormatUint(uint64(materialID), 10) + "/files/" + strconv.FormatUint(uint64(row.ID), 10) + "/download",
			"updated_at":  now,
		}).Error
	})
	if err != nil {
		if createdFile {
			_ = os.Remove(storagePath)
		}
		common.APIInternalServerError(c, "上传失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"ok": true, "digest": digest})
}

// DownloadCaseMaterialFile streams one stored attachment after the
// same object-level check the list endpoint uses.
func (h *MatterCollaborationHandler) DownloadCaseMaterialFile(c *gin.Context) {
	materialID, ok1 := parseUintParam(c, "materialId")
	fileID, ok2 := parseUintParam(c, "fileId")
	if !ok1 || !ok2 {
		return
	}
	var row models.MaterialFile
	if err := h.db.Where("id = ? AND material_id = ?", fileID, materialID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.APINotFound(c, "附件不存在")
		} else {
			common.APIInternalServerError(c, "读取附件失败", err.Error())
		}
		return
	}
	scope, ok := h.guard.resolveCaseScope(c, materialID)
	if !ok {
		return
	}
	// Intake-stage materials authorize through the intake boundary (not
	// CanReadCase), so an un-adopted UUID material never hits a numeric
	// case check that would always 403 it (review 2026-09-30).
	if !h.authorizeMaterialScope(c, scope, false) {
		return
	}
	if _, err := os.Stat(row.StoragePath); err != nil {
		common.APIInternalServerError(c, "附件文件缺失", err.Error())
		return
	}
	c.FileAttachment(row.StoragePath, row.FileName)
}

func (h *MatterCollaborationHandler) storageRoot() string {
	if strings.TrimSpace(h.StorageDir) != "" {
		return h.StorageDir
	}
	return "."
}

// authorizeMaterialScope dispatches one material row to its real boundary:
// adopted case rows go through the case check; intake rows through the
// intake boundary. Read paths let the current approval/review actors
// through (write stays owner-only); write paths stay strictly
// creator-or-management so a reviewer can never mutate materials.
func (h *MatterCollaborationHandler) authorizeMaterialScope(c *gin.Context, scope materialScope, write bool) bool {
	if scope.CaseID > 0 {
		if write {
			return h.guard.requireCaseManageable(c, scope.CaseID)
		}
		return h.guard.requireCaseReadable(c, scope.CaseID)
	}
	if scope.IntakeID != "" {
		return h.guard.authorizeIntakeMaterial(c, scope.IntakeID, write)
	}
	forbidObjectAccess(c)
	return false
}

// assigneeAllowed reports whether userID is a legal assignee inside the
// case: the case lawyer, the creator, a matter-management role holder,
// or an explicit ethical-wall whitelist member. New rows are re-validated
// at creation time with the same query.
func (h *MatterCollaborationHandler) assigneeAllowed(caseID uint, userID uint) bool {
	if userID == 0 || caseID == 0 {
		return false
	}
	var count int64
	err := h.db.Table("cases").Where("id = ? AND deleted_at IS NULL AND (lawyer_id = ? OR created_by = ? OR EXISTS (SELECT 1 FROM case_ethical_wall_whitelist w WHERE w.case_id = cases.id AND w.user_id = ?))", caseID, userID, fmt.Sprint(userID), userID).Count(&count).Error
	if err != nil || count > 0 {
		return err == nil && count > 0
	}
	// Management roles may hold cases in aggregate; check the users table.
	var user models.User
	if err := h.db.Where("id = ? AND deleted_at IS NULL", userID).Take(&user).Error; err != nil {
		return false
	}
	return services.IsBusinessMatterManagementRole(user.Role)
}

// intakeOwnerAllowed reports whether userID may be named the responsible
// person of a material row that has no case yet: strictly the intake
// creator. No team and no global role shortcuts (review 2026-09-30,
// defect 2; trial attempt2 owner-propagation finding).
func (h *MatterCollaborationHandler) intakeOwnerAllowed(intakeID string, userID uint) bool {
	if userID == 0 || strings.TrimSpace(intakeID) == "" {
		return false
	}
	var count int64
	// The responsibility boundary is the intake creator, full stop: the
	// previous "anyone already carrying an owner assignment" branch let
	// one legal assignment admit that user for every future row on this
	// intake. Owner candidates must be provable collaborators; without
	// a collaboration table the creator is the only member this query
	// can prove, so anyone else needs a real collaboration grant, not
	// row history.
	err := h.db.Table("case_intakes").
		Where("id = ? AND created_by = ?", intakeID, fmt.Sprint(userID)).
		Count(&count).Error
	return err == nil && count > 0
}

type byteWriter struct{ buf *[]byte }

func (w *byteWriter) Write(b []byte) (int, error) { *w.buf = append(*w.buf, b...); return len(b), nil }

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	raw := strings.TrimSpace(c.Param(name))
	v, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || v == 0 {
		common.APIBadRequest(c, "请求参数错误", name+" 必须是有效数字")
		return 0, false
	}
	return uint(v), true
}
