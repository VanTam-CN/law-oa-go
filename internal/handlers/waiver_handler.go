package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type WaiverHandler struct {
	waiverService *services.WaiverWorkflowService
	db            *gorm.DB
	authz         *services.AuthorizationService
}

func NewWaiverHandler(waiverService *services.WaiverWorkflowService) *WaiverHandler {
	return &WaiverHandler{waiverService: waiverService}
}

func NewWaiverHandlerWithAuthorization(waiverService *services.WaiverWorkflowService, db *gorm.DB, authz *services.AuthorizationService) *WaiverHandler {
	return &WaiverHandler{waiverService: waiverService, db: db, authz: authz}
}

// authorizeConflictTask applies the same case-context and ethical-wall gate
// as the conflict result handlers. Waiver payloads contain sensitive conflict
// facts, so the waiver endpoint must not become a disclosure or write side
// channel.
func (h *WaiverHandler) authorizeConflictTask(c *gin.Context, taskID string) bool {
	if h.db == nil || h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_CONTEXT_AUTHZ_UNAVAILABLE", "冲突案件权限服务未初始化，已阻止豁免操作")
		return false
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		common.APIBadRequest(c, "冲突检测任务ID不能为空")
		return false
	}
	var record struct {
		// Read the JSON column in its database representation. The business JSON
		// map type is useful after decoding, but is not portable across GORM's
		// SQLite and PostgreSQL scanners when used in an ad-hoc projection.
		SearchParameters []byte `gorm:"column:search_parameters"`
	}
	if err := h.db.WithContext(c.Request.Context()).Table("conflict_check_records").Where("check_id = ?", taskID).Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.APINotFound(c, "冲突检测任务不存在")
			return false
		}
		common.APIInternalServerError(c, "读取冲突检测案件上下文失败", err.Error())
		return false
	}
	caseIDText := subjectCaseIDFromAggregateSearchParameters(record.SearchParameters)
	if caseIDText == "" || caseIDText == "0" {
		var caseIDs []uint
		if err := h.db.WithContext(c.Request.Context()).Table("cases").Where("conflict_check_id = ? AND deleted_at IS NULL", taskID).Limit(2).Pluck("id", &caseIDs).Error; err != nil {
			common.NewAPIError(c, http.StatusConflict, "CASE_CONTEXT_REQUIRED", "冲突检测缺少可验证的案件上下文，已阻止豁免操作")
			return false
		}
		if len(caseIDs) != 1 {
			common.NewAPIError(c, http.StatusConflict, "CASE_CONTEXT_REQUIRED", "冲突检测缺少可验证的案件上下文，已阻止豁免操作")
			return false
		}
		caseIDText = strconv.FormatUint(uint64(caseIDs[0]), 10)
	}
	caseID, err := strconv.ParseUint(caseIDText, 10, 32)
	if err != nil || caseID == 0 {
		common.NewAPIError(c, http.StatusConflict, "CASE_CONTEXT_INVALID", "冲突检测案件上下文无效，已阻止豁免操作")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, authErr := h.authz.CanReadConflictContext(c.Request.Context(), actor, uint(caseID))
	if authErr != nil {
		common.APIInternalServerError(c, "案件权限校验失败", authErr.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

func (h *WaiverHandler) CreateWaiverRequest(c *gin.Context) {
	approvalID := c.Param("id")
	var req services.CreateWaiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	if !h.authorizeConflictTask(c, req.ConflictCheckID) {
		return
	}

	userID := contextUserID(c)
	userName := contextUserName(c)
	application, err := h.waiverService.CreateWaiver(c.Request.Context(), approvalID, userID, userName, contextUserRole(c), &req)
	if err != nil {
		if errors.Is(err, services.ErrWaiverForbidden) {
			common.NewAPIError(c, http.StatusForbidden, "WAIVER_FORBIDDEN", "无权创建或自批该豁免申请")
			return
		}
		common.APIBadRequest(c, "创建豁免申请失败", err.Error())
		return
	}

	common.APISuccess(c, application)
}

func (h *WaiverHandler) CreateConflictWaiverRequest(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if taskID == "" {
		common.APIBadRequest(c, "冲突检测任务ID不能为空")
		return
	}
	var req services.CreateWaiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	req.ConflictCheckID = taskID
	if !h.authorizeConflictTask(c, taskID) {
		return
	}
	application, err := h.waiverService.CreateWaiver(c.Request.Context(), "", contextUserID(c), contextUserName(c), contextUserRole(c), &req)
	if err != nil {
		if errors.Is(err, services.ErrWaiverForbidden) {
			common.NewAPIError(c, http.StatusForbidden, "WAIVER_FORBIDDEN", "无权创建或自批该豁免申请")
			return
		}
		if errors.Is(err, services.ErrConflictTaskNotFound) {
			common.APINotFound(c, "冲突检测任务不存在")
			return
		}
		common.APIBadRequest(c, "创建豁免申请失败", err.Error())
		return
	}
	common.APISuccess(c, application)
}

func (h *WaiverHandler) GetConflictWaiverRequest(c *gin.Context) {
	taskID := strings.TrimSpace(c.Param("task_id"))
	if !h.authorizeConflictTask(c, taskID) {
		return
	}
	application, err := h.waiverService.GetConflictWaiver(c.Request.Context(), taskID, contextUserID(c), contextUserRole(c))
	if err != nil {
		if errors.Is(err, services.ErrWaiverForbidden) {
			common.NewAPIError(c, http.StatusForbidden, "WAIVER_FORBIDDEN", "无权查看该豁免申请")
			return
		}
		if errors.Is(err, services.ErrConflictTaskNotFound) {
			common.APINotFound(c, "冲突检测任务不存在")
			return
		}
		if errors.Is(err, services.ErrWaiverNotFound) {
			common.NewAPIError(c, http.StatusNotFound, "WAIVER_NOT_FOUND", "该冲突检测暂无豁免申请")
			return
		}
		common.APIInternalServerError(c, "获取豁免申请失败", err.Error())
		return
	}
	common.APISuccess(c, application)
}

func (h *WaiverHandler) GetWaiverRequest(c *gin.Context) {
	waiverID := c.Param("id")
	if waiverID == "" {
		common.APIBadRequest(c, "豁免申请ID不能为空")
		return
	}

	application, err := h.waiverService.GetWaiver(c.Request.Context(), waiverID, contextUserID(c), contextUserRole(c))
	if err != nil {
		if errors.Is(err, services.ErrWaiverForbidden) {
			common.NewAPIError(c, http.StatusForbidden, "WAIVER_FORBIDDEN", "无权查看该豁免申请")
			return
		}
		if errors.Is(err, services.ErrWaiverNotFound) {
			common.APINotFound(c, "豁免申请不存在")
			return
		}
		common.APIInternalServerError(c, "获取豁免申请失败", err.Error())
		return
	}
	if !h.authorizeConflictTask(c, application.ConflictCheckID) {
		return
	}

	common.APISuccess(c, application)
}

func (h *WaiverHandler) DecideWaiverRequest(c *gin.Context) {
	waiverID := c.Param("id")
	if waiverID == "" {
		common.APIBadRequest(c, "豁免申请ID不能为空")
		return
	}

	var req services.WaiverDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	current, lookupErr := h.waiverService.GetWaiver(c.Request.Context(), waiverID, contextUserID(c), contextUserRole(c))
	if lookupErr != nil {
		if errors.Is(lookupErr, services.ErrWaiverForbidden) {
			common.NewAPIError(c, http.StatusForbidden, "WAIVER_FORBIDDEN", "申请人不得自批，且只有指定复核人或管理人员可以处理豁免")
			return
		}
		if errors.Is(lookupErr, services.ErrWaiverNotFound) {
			common.APINotFound(c, "豁免申请不存在")
			return
		}
		common.APIInternalServerError(c, "读取豁免申请失败", lookupErr.Error())
		return
	}
	if !h.authorizeConflictTask(c, current.ConflictCheckID) {
		return
	}

	application, err := h.waiverService.DecideWaiver(c.Request.Context(), waiverID, contextUserID(c), contextUserName(c), contextUserRole(c), &req)
	if err != nil {
		if errors.Is(err, services.ErrWaiverForbidden) {
			common.NewAPIError(c, http.StatusForbidden, "WAIVER_FORBIDDEN", "申请人不得自批，且只有指定复核人或管理人员可以处理豁免")
			return
		}
		if errors.Is(err, services.ErrWaiverNotFound) {
			common.APINotFound(c, "豁免申请不存在")
			return
		}
		common.APIBadRequest(c, "处理豁免决定失败", err.Error())
		return
	}

	common.APISuccess(c, application)
}

func contextUserRole(c *gin.Context) string {
	return strings.ToLower(strings.TrimSpace(c.GetString("role")))
}

func contextUserID(c *gin.Context) string {
	value, exists := c.Get("user_id")
	if !exists {
		return "0"
	}
	switch v := value.(type) {
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case string:
		if v != "" {
			return v
		}
	}
	return "0"
}

func contextUserName(c *gin.Context) string {
	value, exists := c.Get("username")
	if !exists {
		return "未知用户"
	}
	if name, ok := value.(string); ok && name != "" {
		return name
	}
	return "未知用户"
}
