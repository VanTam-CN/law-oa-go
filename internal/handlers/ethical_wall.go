package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/auth"
	"law-oa-go/internal/common"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// EthicalWallHandler 隔离墙处理器
type EthicalWallHandler struct {
	ethicalWallService *services.EthicalWallService
}

// NewEthicalWallHandler 创建隔离墙处理器实例
func NewEthicalWallHandler(ethicalWallService *services.EthicalWallService) *EthicalWallHandler {
	return &EthicalWallHandler{
		ethicalWallService: ethicalWallService,
	}
}

// EnableEthicalWall 启用案件隔离墙
// @Summary 启用案件隔离墙
// @Description 为指定案件启用隔离墙保护
// @Tags 隔离墙管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Param request body services.EnableEthicalWallRequest true "启用隔离墙请求"
// @Success 200 {object} common.APIResponse "启用成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "隔离墙已启用"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/cases/{caseId}/ethical-wall [post]
func (h *EthicalWallHandler) EnableEthicalWall(c *gin.Context) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	var req services.EnableEthicalWallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}

	// 获取当前用户ID
	userID := auth.GetUserID(c)

	err = h.ethicalWallService.EnableEthicalWall(c.Request.Context(), uint(caseID), userID, req.Description)
	if err != nil {
		switch err {
		case services.ErrCaseNotFound:
			common.APINotFound(c, "案件不存在", "指定的案件ID不存在")
		case services.ErrEthicalWallAlreadyEnabled:
			common.NewAPIError(c, http.StatusForbidden, "ETHICAL_WALL_ALREADY_ENABLED", "该案件已启用隔离墙")
		default:
			common.APIInternalServerError(c, "启用隔离墙失败", err.Error())
		}
		return
	}

	common.APISuccess(c, gin.H{
		"message":   "隔离墙启用成功",
		"case_id":   caseID,
		"enabled":   true,
	})
}

// DisableEthicalWall 禁用案件隔离墙
// @Summary 禁用案件隔离墙
// @Description 禁用指定案件的隔离墙保护
// @Tags 隔离墙管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Success 200 {object} common.APIResponse "禁用成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "隔离墙未启用"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/cases/{caseId}/ethical-wall [delete]
func (h *EthicalWallHandler) DisableEthicalWall(c *gin.Context) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	err = h.ethicalWallService.DisableEthicalWall(c.Request.Context(), uint(caseID))
	if err != nil {
		switch err {
		case services.ErrCaseNotFound:
			common.APINotFound(c, "案件不存在", "指定的案件ID不存在")
		case services.ErrEthicalWallNotEnabled:
			common.NewAPIError(c, http.StatusForbidden, "ETHICAL_WALL_NOT_ENABLED", "该案件未启用隔离墙")
		default:
			common.APIInternalServerError(c, "禁用隔离墙失败", err.Error())
		}
		return
	}

	common.APISuccess(c, gin.H{
		"message":   "隔离墙禁用成功",
		"case_id":   caseID,
		"enabled":   false,
	})
}

// GetWhitelist 获取案件白名单
// @Summary 获取案件白名单
// @Description 获取指定案件的隔离墙白名单
// @Tags 隔离墙管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=[]services.WhitelistEntryResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/cases/{caseId}/ethical-wall/whitelist [get]
func (h *EthicalWallHandler) GetWhitelist(c *gin.Context) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	whitelist, err := h.ethicalWallService.GetWhitelist(c.Request.Context(), uint(caseID))
	if err != nil {
		switch err {
		case services.ErrCaseNotFound:
			common.APINotFound(c, "案件不存在", "指定的案件ID不存在")
		default:
			common.APIInternalServerError(c, "获取白名单失败", err.Error())
		}
		return
	}

	common.APISuccess(c, gin.H{
		"case_id":   caseID,
		"whitelist": whitelist,
		"count":     len(whitelist),
	})
}

// AddToWhitelist 添加用户到案件白名单
// @Summary 添加用户到白名单
// @Description 将指定用户添加到案件隔离墙白名单
// @Tags 隔离墙管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Param request body services.WhitelistEntryRequest true "添加白名单请求"
// @Success 200 {object} common.APIResponse{data=services.WhitelistEntryResponse} "添加成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "隔离墙未启用"
// @Failure 404 {object} common.APIResponse "案件或用户不存在"
// @Failure 409 {object} common.APIResponse "用户已在白名单中"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/cases/{caseId}/ethical-wall/whitelist [post]
func (h *EthicalWallHandler) AddToWhitelist(c *gin.Context) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	var req services.WhitelistEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查请求体格式")
		return
	}

	// 获取当前用户ID（授权人）
	grantedBy := auth.GetUserID(c)

	err = h.ethicalWallService.AddToWhitelist(c.Request.Context(), uint(caseID), req.UserID, grantedBy, req.Reason)
	if err != nil {
		switch err {
		case services.ErrCaseNotFound:
			common.APINotFound(c, "案件不存在", "指定的案件ID不存在")
		case services.ErrUserNotFound:
			common.APINotFound(c, "用户不存在", "指定的用户ID不存在")
		case services.ErrEthicalWallNotEnabled:
			common.NewAPIError(c, http.StatusForbidden, "ETHICAL_WALL_NOT_ENABLED", "该案件未启用隔离墙，无法添加白名单")
		case repositories.ErrWhitelistEntryExists:
			common.NewAPIError(c, http.StatusConflict, "WHITELIST_ENTRY_EXISTS", "该用户已在案件白名单中")
		default:
			common.APIInternalServerError(c, "添加白名单失败", err.Error())
		}
		return
	}

	// 获取更新后的白名单
	whitelist, err := h.ethicalWallService.GetWhitelist(c.Request.Context(), uint(caseID))
	if err != nil {
		common.APIInternalServerError(c, "获取更新后的白名单失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message":   "添加白名单成功",
		"case_id":   caseID,
		"user_id":   req.UserID,
		"whitelist": whitelist,
	})
}

// RemoveFromWhitelist 从案件白名单移除用户
// @Summary 从白名单移除用户
// @Description 将指定用户从案件隔离墙白名单中移除
// @Tags 隔离墙管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Param userId path int true "用户ID"
// @Success 200 {object} common.APIResponse "移除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件或白名单条目不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/cases/{caseId}/ethical-wall/whitelist/{userId} [delete]
func (h *EthicalWallHandler) RemoveFromWhitelist(c *gin.Context) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "用户ID必须是有效数字")
		return
	}

	err = h.ethicalWallService.RemoveFromWhitelist(c.Request.Context(), uint(caseID), uint(userID))
	if err != nil {
		switch err {
		case services.ErrCaseNotFound:
			common.APINotFound(c, "案件不存在", "指定的案件ID不存在")
		case repositories.ErrWhitelistEntryNotFound:
			common.APINotFound(c, "白名单条目不存在", "指定的用户不在该案件的白名单中")
		default:
			common.APIInternalServerError(c, "移除白名单失败", err.Error())
		}
		return
	}

	// 获取更新后的白名单
	whitelist, err := h.ethicalWallService.GetWhitelist(c.Request.Context(), uint(caseID))
	if err != nil {
		common.APIInternalServerError(c, "获取更新后的白名单失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message":   "移除白名单成功",
		"case_id":   caseID,
		"user_id":   userID,
		"whitelist": whitelist,
	})
}

// GetUserAccessibleCases 获取用户可访问的隔离墙案件
// @Summary 获取用户可访问的隔离墙案件
// @Description 获取当前用户在白名单中的所有隔离墙案件
// @Tags 隔离墙管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]services.WhitelistEntryResponse} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/ethical-wall/accessible-cases [get]
func (h *EthicalWallHandler) GetUserAccessibleCases(c *gin.Context) {
	userID := auth.GetUserID(c)

	cases, err := h.ethicalWallService.GetUserAccessibleCases(c.Request.Context(), userID)
	if err != nil {
		common.APIInternalServerError(c, "获取可访问案件失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"cases": cases,
		"count": len(cases),
	})
}

// GetAccessLogs 获取案件访问日志
// @Summary 获取案件访问日志
// @Description 获取指定案件的隔离墙访问日志
// @Tags 隔离墙管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Param limit query int false "返回数量限制" default(50)
// @Success 200 {object} common.APIResponse{data=[]services.AccessLogResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/cases/{caseId}/ethical-wall/access-logs [get]
func (h *EthicalWallHandler) GetAccessLogs(c *gin.Context) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	logs, err := h.ethicalWallService.GetAccessLogs(c.Request.Context(), uint(caseID), limit)
	if err != nil {
		common.APIInternalServerError(c, "获取访问日志失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"case_id": caseID,
		"logs":    logs,
		"count":   len(logs),
	})
}
