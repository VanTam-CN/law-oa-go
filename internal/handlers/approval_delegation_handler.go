package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// ApprovalDelegationHandler 代理审批配置处理器
type ApprovalDelegationHandler struct {
	service services.ApprovalDelegationService
}

// NewApprovalDelegationHandler 创建代理审批配置处理器
func NewApprovalDelegationHandler(service services.ApprovalDelegationService) *ApprovalDelegationHandler {
	return &ApprovalDelegationHandler{service: service}
}

// CreateDelegation godoc
// @Summary 创建代理审批配置
// @Description 为审批人配置代理人
// @Tags 审批代理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateDelegationRequest true "代理配置请求"
// @Success 200 {object} common.APIResponse "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Router /approvals/delegations [post]
func (h *ApprovalDelegationHandler) CreateDelegation(c *gin.Context) {
	var req services.CreateDelegationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	currentUserID := strconv.FormatUint(uint64(actor.UserID), 10)
	if services.IsPrivilegedRole(actor.Role) {
		if req.DelegatorID == "" {
			req.DelegatorID = currentUserID
		}
		req.CreatedBy = currentUserID
	} else {
		req.DelegatorID = currentUserID
		req.CreatedBy = currentUserID
	}

	delegation, err := h.service.CreateDelegation(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建代理配置失败", err.Error())
		return
	}

	common.APISuccess(c, delegation)
}

// ListDelegations godoc
// @Summary 查询代理配置列表
// @Description 查询代理审批配置
// @Tags 审批代理
// @Produce json
// @Security BearerAuth
// @Param delegator_id query string false "委托人ID"
// @Param delegate_id query string false "代理人ID"
// @Param is_active query bool false "是否活跃"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} common.APIResponse "查询成功"
// @Router /approvals/delegations [get]
func (h *ApprovalDelegationHandler) ListDelegations(c *gin.Context) {
	// The protected router normally populates the actor. Fail closed here as
	// well, otherwise a missing auth context would fall through to the
	// privileged all-records branch.
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}

	// IDOR 防护：非管理员只能查看自己相关的代理配置
	if !services.IsPrivilegedRole(actor.Role) {
		// 非管理员：强制限定只能查看自己相关的记录
		userIDStr := strconv.FormatUint(uint64(actor.UserID), 10)
		params := &repositories.DelegationListParams{
			Page:     1,
			PageSize: 20,
		}
		if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
			params.Page = page
		}
		if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 && pageSize <= 100 {
			params.PageSize = pageSize
		}
		if active := c.Query("is_active"); active != "" {
			isActive := active == "true"
			params.IsActive = &isActive
		}
		// 非管理员只能看到自己作为委托人或代理人的记录
		if c.Query("filter") == "delegate" {
			params.DelegateID = userIDStr
		} else {
			params.DelegatorID = userIDStr
		}

		delegations, total, err := h.service.ListDelegations(c.Request.Context(), params)
		if err != nil {
			common.APIInternalServerError(c, "查询代理配置失败", err.Error())
			return
		}

		common.APISuccess(c, gin.H{
			"items":     delegations,
			"total":     total,
			"page":      params.Page,
			"page_size": params.PageSize,
		})
		return
	}

	// 管理员：可查看所有代理配置
	params := &repositories.DelegationListParams{
		DelegatorID: c.Query("delegator_id"),
		DelegateID:  c.Query("delegate_id"),
		Page:        1,
		PageSize:    20,
	}

	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		params.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 && pageSize <= 100 {
		params.PageSize = pageSize
	}
	if active := c.Query("is_active"); active != "" {
		isActive := active == "true"
		params.IsActive = &isActive
	}

	delegations, total, err := h.service.ListDelegations(c.Request.Context(), params)
	if err != nil {
		common.APIInternalServerError(c, "查询代理配置失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"items":     delegations,
		"total":     total,
		"page":      params.Page,
		"page_size": params.PageSize,
	})
}

// GetDelegation godoc
// @Summary 获取代理配置详情
// @Description 获取指定代理配置的详细信息
// @Tags 审批代理
// @Produce json
// @Security BearerAuth
// @Param id path string true "代理配置ID"
// @Success 200 {object} common.APIResponse "获取成功"
// @Failure 404 {object} common.APIResponse "不存在"
// @Router /approvals/delegations/{id} [get]
func (h *ApprovalDelegationHandler) GetDelegation(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if id == "" {
		common.APIBadRequest(c, "参数错误", "id不能为空")
		return
	}

	delegation, err := h.service.GetDelegation(c.Request.Context(), id)
	if err != nil {
		common.APIInternalServerError(c, "获取代理配置失败", err.Error())
		return
	}
	if delegation == nil {
		common.APINotFound(c, "代理配置不存在", "ID: "+id)
		return
	}
	userIDStr := strconv.FormatUint(uint64(actor.UserID), 10)
	if !services.IsPrivilegedRole(actor.Role) &&
		delegation.DelegatorID != userIDStr &&
		delegation.DelegateID != userIDStr &&
		delegation.CreatedBy != userIDStr {
		common.APIForbidden(c, "无权访问", "只能查看与自己相关的代理配置")
		return
	}

	common.APISuccess(c, delegation)
}

// RevokeDelegation godoc
// @Summary 撤销代理配置
// @Description 撤销指定的代理审批配置
// @Tags 审批代理
// @Produce json
// @Security BearerAuth
// @Param id path string true "代理配置ID"
// @Success 200 {object} common.APIResponse "撤销成功"
// @Failure 404 {object} common.APIResponse "不存在"
// @Router /approvals/delegations/{id} [delete]
func (h *ApprovalDelegationHandler) RevokeDelegation(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if id == "" {
		common.APIBadRequest(c, "参数错误", "id不能为空")
		return
	}

	// 权限校验：只能撤销自己创建的或自己是委托人的代理配置
	delegation, err := h.service.GetDelegation(c.Request.Context(), id)
	if err != nil {
		common.APIInternalServerError(c, "获取代理配置失败", err.Error())
		return
	}
	if delegation == nil {
		common.APINotFound(c, "代理配置不存在", "ID: "+id)
		return
	}

	// 检查权限：管理员、委托人本人、或创建者可撤销
	canRevoke := false
	userIDStr := strconv.FormatUint(uint64(actor.UserID), 10)
	canRevoke = services.IsPrivilegedRole(actor.Role) || delegation.DelegatorID == userIDStr || delegation.CreatedBy == userIDStr
	if !canRevoke {
		common.APIForbidden(c, "无权操作", "只能撤销自己的代理配置")
		return
	}

	if err := h.service.RevokeDelegation(c.Request.Context(), id); err != nil {
		common.APIInternalServerError(c, "撤销代理配置失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "代理配置已撤销"})
}

// MyDelegations godoc
// @Summary 获取我的代理配置
// @Description 获取当前用户相关的代理配置（作为委托人和作为代理人）
// @Tags 审批代理
// @Produce json
// @Security BearerAuth
// @Param user_id query int true "用户ID"
// @Success 200 {object} common.APIResponse "获取成功"
// @Router /approvals/delegations/my [get]
func (h *ApprovalDelegationHandler) MyDelegations(c *gin.Context) {
	// 从 JWT context 获取当前用户ID，防止 IDOR 越权
	userIDStr, ok := currentUserIDString(c)
	if !ok {
		return
	}

	delegations, err := h.service.GetActiveDelegations(c.Request.Context(), userIDStr)
	if err != nil {
		common.APIInternalServerError(c, "获取代理配置失败", err.Error())
		return
	}

	common.APISuccess(c, delegations)
}
