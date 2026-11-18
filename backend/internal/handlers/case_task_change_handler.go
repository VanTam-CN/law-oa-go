package handlers

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// CaseTaskChangeHandler serves the adjust / cancel / history / creation
// lookup endpoints (task change refinement, 2026-10-02). Every route
// entry runs an object-level case readability check before touching
// task data; object-level adjust/cancel/view permissions re-validate
// inside the service transaction using the real AuthorizationService.
type CaseTaskChangeHandler struct {
	db    *gorm.DB
	authz *services.AuthorizationService
	guard *materialAccessGuard
	svc   *services.CaseTaskChangeService
}

func NewCaseTaskChangeHandler(authz *services.AuthorizationService, db *gorm.DB) *CaseTaskChangeHandler {
	return &CaseTaskChangeHandler{
		db:    db,
		authz: authz,
		guard: newMaterialAccessGuard(authz, db),
		svc:   services.NewCaseTaskChangeServiceWithAuthz(db, authz),
	}
}

// NewCaseTaskChangeHandlerWithDB is the router integration shape.
func NewCaseTaskChangeHandlerWithDB(db *gorm.DB) *CaseTaskChangeHandler {
	return NewCaseTaskChangeHandler(services.NewAuthorizationServiceWithDB(db), db)
}

func parseTaskChangeID(c *gin.Context, name string) (uint, bool) {
	raw := strings.TrimSpace(c.Param(name))
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || v == 0 {
		common.APIBadRequest(c, "请求参数错误", name+" 必须是正整数")
		return 0, false
	}
	return uint(v), true
}

// GetCreationResult resolves one creation attempt by its idempotency
// key: found (task + version), not_found (the caller may retry the
// original body with the SAME key), or unknown (the server cannot
// prove either way; the client keeps the pending state and never
// rotates the key on its own).
func (h *CaseTaskChangeHandler) GetCreationResult(c *gin.Context) {
	caseID, ok := parseTaskChangeID(c, "id")
	if !ok {
		return
	}
	requestID := strings.TrimSpace(c.Query("request_id"))
	if requestID == "" {
		common.APIBadRequest(c, "请求参数错误", "request_id 不能为空")
		return
	}
	if !h.guard.requireCaseReadable(c, caseID) {
		return
	}
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	// FindTaskByRequestId scopes the lookup to the acting user's own
	// keys, so one lawyer can never probe another's request ids.
	task, version, err := h.svc.FindTaskByRequestId(actor.UserID, requestID)
	if err != nil {
		common.APIInternalServerError(c, "查询创建结果失败", err.Error())
		return
	}
	if task == nil {
		common.APISuccess(c, gin.H{"state": "not_found"})
		return
	}
	if task.CaseID != caseID {
		// A matching key bound to a different case is outside this
		// scope; report unknown instead of leaking the other row.
		common.APISuccess(c, gin.H{"state": "unknown"})
		return
	}
	common.APISuccess(c, gin.H{"state": "found", "task": task, "version": version})
}

// caseTaskChangeRequest is the shared adjust/cancel HTTP body. The
// concurrency guard is mandatory: version or expected_updated_at must
// be present or the write is refused (fail closed).
type caseTaskChangeRequest struct {
	Version           *uint   `json:"version,omitempty"`
	ExpectedUpdatedAt *string `json:"expected_updated_at,omitempty"`
	AssigneeID        *uint   `json:"assignee_id,omitempty"`
	DueDate           *string `json:"due_date,omitempty"`
	ClearDueDate      bool    `json:"clear_due_date,omitempty"`
	Reason            string  `json:"reason,omitempty"`
}

func (r *caseTaskChangeRequest) toService() (*services.CaseTaskChangeRequest, error) {
	out := &services.CaseTaskChangeRequest{
		Version:           r.Version,
		ExpectedUpdatedAt: r.ExpectedUpdatedAt,
		AssigneeID:        r.AssigneeID,
		ClearDueDate:      r.ClearDueDate,
		Reason:            r.Reason,
	}
	if r.DueDate != nil && strings.TrimSpace(*r.DueDate) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*r.DueDate))
		if err != nil {
			return nil, err
		}
		out.DueDate = &t
	}
	return out, nil
}

// AdjustCaseTask handles PUT /case-tasks/:taskId (reassign and/or
// due-date change with optimistic locking and audit).
func (h *CaseTaskChangeHandler) AdjustCaseTask(c *gin.Context) {
	taskID, ok := parseTaskChangeID(c, "taskId")
	if !ok {
		return
	}
	var req caseTaskChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}
	svcReq, err := req.toService()
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "due_date 格式应为 RFC3339 时间")
		return
	}
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	task, change, err := h.svc.Adjust(actor.UserID, taskID, svcReq)
	if err != nil {
		mapCaseTaskChangeError(c, err)
		return
	}
	common.APISuccess(c, gin.H{"task": task, "change": change})
}

// CancelCaseTask handles POST /case-tasks/:taskId/cancel. An
// already-cancelled task with a current guard replays idempotently.
func (h *CaseTaskChangeHandler) CancelCaseTask(c *gin.Context) {
	taskID, ok := parseTaskChangeID(c, "taskId")
	if !ok {
		return
	}
	var req caseTaskChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}
	svcReq, err := req.toService()
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "due_date 格式应为 RFC3339 时间")
		return
	}
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	task, change, replayed, err := h.svc.Cancel(actor.UserID, taskID, svcReq)
	if err != nil {
		mapCaseTaskChangeError(c, err)
		return
	}
	common.APISuccess(c, gin.H{"task": task, "change": change, "replayed": replayed})
}

// ListCaseTaskChanges handles GET /case-tasks/:taskId/changes. View
// permission is creator-still-live / current assignee / case manager;
// a case-readable outsider still gets 403 on the task object.
func (h *CaseTaskChangeHandler) ListCaseTaskChanges(c *gin.Context) {
	taskID, ok := parseTaskChangeID(c, "taskId")
	if !ok {
		return
	}
	actor, aok := currentAuthActor(c)
	if !aok {
		return
	}
	task, version, err := h.svc.LoadTaskForChange(taskID)
	if err != nil {
		common.APINotFound(c, "任务不存在")
		return
	}
	if !h.guard.requireCaseReadable(c, task.CaseID) {
		return
	}
	if !h.svc.CanViewChanges(actor.UserID, task) {
		forbidObjectAccess(c)
		return
	}
	changes, err := h.svc.ListChanges(taskID)
	if err != nil {
		common.APIInternalServerError(c, "读取变更记录失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"task": task, "version": version, "changes": changes})
}

// mapCaseTaskChangeError maps change-service sentinels to HTTP codes.
func mapCaseTaskChangeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrCaseTaskChangeStale):
		common.NewAPIError(c, 409, "CASE_TASK_CHANGE_STALE", "任务已被他人修改，请刷新后基于最新内容重试")
	case errors.Is(err, services.ErrCaseTaskChangeNoop):
		common.APIBadRequest(c, "没有需要变更的字段", "请至少修改负责人或截止时间")
	case errors.Is(err, services.ErrCaseTaskChangeForbidden):
		forbidObjectAccess(c)
	case errors.Is(err, services.ErrCaseTaskChangeViewForbidden):
		forbidObjectAccess(c)
	case errors.Is(err, services.ErrCaseTaskChangeNotOpen):
		common.NewAPIError(c, 409, "CASE_TASK_CHANGE_NOT_OPEN", "任务已结束，不能调整")
	case errors.Is(err, services.ErrCaseTaskChangeRequestMissing):
		common.APIBadRequest(c, "缺少并发保护字段", "请携带 version 或 expected_updated_at")
	case errors.Is(err, services.ErrCaseTaskAssigneeInvalid):
		common.NewAPIError(c, 403, "CASE_TASK_ASSIGNEE_INVALID", "任务负责人必须是案件内合法人员")
	default:
		if strings.Contains(err.Error(), "不存在") {
			common.APINotFound(c, "任务不存在")
			return
		}
		common.APIInternalServerError(c, "案件任务变更失败", err.Error())
	}
}
