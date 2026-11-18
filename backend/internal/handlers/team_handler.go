package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// TeamHandler 团队管理处理器
type TeamHandler struct {
	teamPermissionService *services.TeamPermissionService
	caseService           *services.CaseService
}

// NewTeamHandler 创建团队处理器
func NewTeamHandler(teamPermissionService *services.TeamPermissionService, caseService *services.CaseService) *TeamHandler {
	return &TeamHandler{
		teamPermissionService: teamPermissionService,
		caseService:           caseService,
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

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsTechnicalAdminRole(actor.Role) {
		forbidObjectAccess(c)
		return
	}
	req.AssignedBy = actor.UserID

	// 分配团队
	result, err := h.teamPermissionService.AssignTeam(c.Request.Context(), &req)
	if err != nil {
		if isSubjectWorkflowError(err) {
			writeSubjectWorkflowError(c, err)
			return
		}
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

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsTechnicalAdminRole(actor.Role) {
		forbidObjectAccess(c)
		return
	}

	// 获取团队分配信息
	result, err := h.teamPermissionService.GetTeamAssignment(c.Request.Context(), actor.UserID, uint(id))
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

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsTechnicalAdminRole(actor.Role) {
		forbidObjectAccess(c)
		return
	}
	// The permission endpoint answers only for the authenticated actor. An
	// arbitrary user_id would turn this endpoint into a cross-user permission
	// oracle and could expose whether another lawyer is assigned to a case.
	if req.UserID == 0 {
		req.UserID = actor.UserID
	}
	if req.UserID != actor.UserID {
		forbidObjectAccess(c)
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
	caseIdStr := c.Param("id")
	memberIdStr := c.Param("memberId")

	caseId, err := strconv.ParseUint(caseIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	_, err = strconv.ParseUint(memberIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "成员ID必须是有效数字")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsTechnicalAdminRole(actor.Role) {
		forbidObjectAccess(c)
		return
	}

	var updateReq map[string]interface{}
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 检查更新权限
	hasPermission, err := h.teamPermissionService.CheckTeamPermission(c.Request.Context(), &services.TeamPermissionCheck{
		UserID: actor.UserID,
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

	common.NewAPIError(c, http.StatusServiceUnavailable, "TEAM_MEMBER_UPDATE_UNAVAILABLE", "团队成员变更尚未接入正式案件主体重检流程，当前不会保存本次操作")
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
	caseIdStr := c.Param("id")
	memberIdStr := c.Param("memberId")

	caseId, err := strconv.ParseUint(caseIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	_, err = strconv.ParseUint(memberIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "成员ID必须是有效数字")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsTechnicalAdminRole(actor.Role) {
		forbidObjectAccess(c)
		return
	}

	// 检查移除权限
	hasPermission, err := h.teamPermissionService.CheckTeamPermission(c.Request.Context(), &services.TeamPermissionCheck{
		UserID: actor.UserID,
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

	common.NewAPIError(c, http.StatusServiceUnavailable, "TEAM_MEMBER_REMOVE_UNAVAILABLE", "团队成员移除尚未接入正式案件主体重检流程，当前不会保存本次操作")
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
	caseIdStr := c.Param("id")
	caseId, err := strconv.ParseUint(caseIdStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsTechnicalAdminRole(actor.Role) {
		forbidObjectAccess(c)
		return
	}

	// 检查查看权限
	hasPermission, err := h.teamPermissionService.CheckTeamPermission(c.Request.Context(), &services.TeamPermissionCheck{
		UserID: actor.UserID,
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

	common.NewAPIError(c, http.StatusServiceUnavailable, "TEAM_MEMBER_LIST_UNAVAILABLE", "案件团队成员清单尚未接入正式数据源，当前不返回演示数据")
}
