package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// TeamHandler 团队管理处理器
type TeamHandler struct {
	teamPermissionService *services.TeamPermissionService
	caseService          *services.CaseService
}

// NewTeamHandler 创建团队处理器
func NewTeamHandler(teamPermissionService *services.TeamPermissionService, caseService *services.CaseService) *TeamHandler {
	return &TeamHandler{
		teamPermissionService: teamPermissionService,
		caseService:          caseService,
	}
}

// AssignTeam godoc
// @Summary 分配案件团队
// @Description 为指定案件分配律师团队
// @Tags 团队管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.TeamAssignmentRequest true "团队分配请求"
// @Success 200 {object} common.APIResponse{data=services.TeamAssignmentResponse} "分配成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /teams/assign [post]
func (h *TeamHandler) AssignTeam(c *gin.Context) {
	var req services.TeamAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户身份验证失败")
		return
	}
	req.AssignedBy = userID.(uint)

	// 分配团队
	result, err := h.teamPermissionService.AssignTeam(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "您没有权限分配此案件的团队" {
			common.APIForbidden(c, "权限不足", err.Error())
			return
		}
		common.APIInternalServerError(c, "团队分配失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// GetTeamAssignment godoc
// @Summary 获取案件团队信息
// @Description 获取指定案件的团队分配信息
// @Tags 团队管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=services.TeamAssignmentResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /teams/case/{id} [get]
func (h *TeamHandler) GetTeamAssignment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户身份验证失败")
		return
	}

	// 获取团队分配信息
	result, err := h.teamPermissionService.GetTeamAssignment(c.Request.Context(), userID.(uint), uint(id))
	if err != nil {
		if err.Error() == "您没有权限查看此案件的团队信息" {
			common.APIForbidden(c, "权限不足", err.Error())
			return
		}
		if err.Error() == "案件不存在" {
			common.APINotFound(c, "案件不存在", "指定的案件ID不存在")
			return
		}
		common.APIInternalServerError(c, "获取团队信息失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// CheckTeamPermission godoc
// @Summary 检查团队权限
// @Description 检查用户对特定案件团队的权限
// @Tags 团队管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.TeamPermissionCheck true "权限检查请求"
// @Success 200 {object} common.APIResponse{data=map[string]interface{}} "检查结果"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /teams/check-permission [post]
func (h *TeamHandler) CheckTeamPermission(c *gin.Context) {
	var req services.TeamPermissionCheck
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 检查权限
	hasPermission, err := h.teamPermissionService.CheckTeamPermission(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}

	result := map[string]interface{}{
		"has_permission": hasPermission,
		"user_id":        req.UserID,
		"case_id":        req.CaseID,
		"action":         req.Action,
	}

	common.APISuccess(c, result)
}

// UpdateTeamMember godoc
// @Summary 更新团队成员
// @Description 更新案件中的团队成员信息
// @Tags 团队管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Param memberId path int true "成员ID"
// @Param request body map[string]interface{} true "更新请求"
// @Success 200 {object} common.APIResponse{data=services.TeamAssignmentResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /teams/case/{caseId}/member/{memberId} [put]
func (h *TeamHandler) UpdateTeamMember(c *gin.Context) {
	caseIdStr := c.Param("caseId")
	memberIdStr := c.Param("memberId")

	caseId, err := strconv.ParseUint(caseIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	memberId, err := strconv.ParseUint(memberIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "成员ID必须是有效数字")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户身份验证失败")
		return
	}

	var updateReq map[string]interface{}
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 检查更新权限
	hasPermission, err := h.teamPermissionService.CheckTeamPermission(c.Request.Context(), &services.TeamPermissionCheck{
		UserID: userID.(uint),
		CaseID: uint(caseId),
		Action: "update",
	})
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !hasPermission {
		common.APIForbidden(c, "权限不足", "您没有权限更新此案件的团队成员")
		return
	}

	// TODO: 实现更新团队成员逻辑
	// 这里应该调用实际的服务来更新团队成员信息

	result := map[string]interface{}{
		"message":    "团队成员更新功能开发中",
		"case_id":    uint(caseId),
		"member_id":  uint(memberId),
		"updated_by": userID.(uint),
		"updates":    updateReq,
	}

	common.APISuccess(c, result)
}

// RemoveTeamMember godoc
// @Summary 移除团队成员
// @Description 从案件中移除团队成员
// @Tags 团队管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Param memberId path int true "成员ID"
// @Success 200 {object} common.APIResponse{data=map[string]interface{}} "移除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /teams/case/{caseId}/member/{memberId} [delete]
func (h *TeamHandler) RemoveTeamMember(c *gin.Context) {
	caseIdStr := c.Param("caseId")
	memberIdStr := c.Param("memberId")

	caseId, err := strconv.ParseUint(caseIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	memberId, err := strconv.ParseUint(memberIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "成员ID必须是有效数字")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户身份验证失败")
		return
	}

	// 检查移除权限
	hasPermission, err := h.teamPermissionService.CheckTeamPermission(c.Request.Context(), &services.TeamPermissionCheck{
		UserID: userID.(uint),
		CaseID: uint(caseId),
		Action: "remove",
	})
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !hasPermission {
		common.APIForbidden(c, "权限不足", "您没有权限移除此案件的团队成员")
		return
	}

	// TODO: 实现移除团队成员逻辑
	// 这里应该调用实际的服务来移除团队成员

	result := map[string]interface{}{
		"message":    "团队成员移除功能开发中",
		"case_id":    uint(caseId),
		"member_id":  uint(memberId),
		"removed_by": userID.(uint),
		"removed_at": "2025-10-21T19:50:00Z",
	}

	common.APISuccess(c, result)
}

// GetTeamMembers godoc
// @Summary 获取案件团队成员列表
// @Description 获取指定案件的所有团队成员
// @Tags 团队管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=[]services.TeamMemberResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /teams/case/{caseId}/members [get]
func (h *TeamHandler) GetTeamMembers(c *gin.Context) {
	caseIdStr := c.Param("caseId")
	caseId, err := strconv.ParseUint(caseIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户身份验证失败")
		return
	}

	// 检查查看权限
	hasPermission, err := h.teamPermissionService.CheckTeamPermission(c.Request.Context(), &services.TeamPermissionCheck{
		UserID: userID.(uint),
		CaseID: uint(caseId),
		Action: "view",
	})
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !hasPermission {
		common.APIForbidden(c, "权限不足", "您没有权限查看此案件的团队成员")
		return
	}

	// TODO: 实现获取团队成员列表逻辑
	// 这里应该调用实际的服务来获取团队成员列表

	// 模拟数据
	members := []services.TeamMemberResponse{
		{
			UserID:     1,
			UserName:   "张律师",
			Email:      "zhang@lawoa.com",
			Role:       "lead_lawyer",
			Department: "民事部",
			JoinedAt:   c.Request.Context().Value("request_time").(time.Time),
			Capacity:   100,
			IsActive:   true,
		},
	}

	result := map[string]interface{}{
		"case_id":     uint(caseId),
		"members":     members,
		"total_count": len(members),
		"permissions": map[string]bool{
			"can_assign": true,
			"can_remove": false,
			"can_update": true,
		},
	}

	common.APISuccess(c, result)
}