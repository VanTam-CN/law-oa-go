package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

// InboxHandler 待办事项处理器
type InboxHandler struct {
	inboxService *services.InboxService
}

// NewInboxHandler 创建待办事项处理器
func NewInboxHandler(inboxService *services.InboxService) *InboxHandler {
	return &InboxHandler{
		inboxService: inboxService,
	}
}

// CreateInboxItem godoc
// @Summary 创建待办事项
// @Description 创建新的待办事项
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateInboxItemRequest true "创建待办事项请求"
// @Success 200 {object} common.APIResponse{data=services.InboxItemResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox [post]
func (h *InboxHandler) CreateInboxItem(c *gin.Context) {
	var req services.CreateInboxItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 从上下文获取用户ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.APIUnauthorized(c, "未授权", "用户未登录")
		return
	}
	req.UserID = uint(userID)

	itemResp, err := h.inboxService.CreateInboxItem(c.Request.Context(), uint(userID), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建待办事项失败", err.Error())
		return
	}

	common.APISuccess(c, itemResp)
}

// GetInboxItem godoc
// @Summary 获取待办事项详情
// @Description 根据ID获取待办事项详细信息
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "待办事项ID"
// @Success 200 {object} common.APIResponse{data=services.InboxItemResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "待办事项不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox/:id [get]
func (h *InboxHandler) GetInboxItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "待办事项ID必须是有效数字")
		return
	}

	itemResp, err := h.inboxService.GetInboxItemByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "待办事项不存在", err.Error())
		return
	}

	common.APISuccess(c, itemResp)
}

// UpdateInboxItem godoc
// @Summary 更新待办事项
// @Description 更新指定待办事项的信息
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "待办事项ID"
// @Param request body services.UpdateInboxItemRequest true "更新待办事项请求"
// @Success 200 {object} common.APIResponse{data=services.InboxItemResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "待办事项不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox/:id [put]
func (h *InboxHandler) UpdateInboxItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "待办事项ID必须是有效数字")
		return
	}

	var req services.UpdateInboxItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	itemResp, err := h.inboxService.UpdateInboxItem(c.Request.Context(), uint(id), &req)
	if err != nil {
		common.APIInternalServerError(c, "更新待办事项失败", err.Error())
		return
	}

	common.APISuccess(c, itemResp)
}

// DeleteInboxItem godoc
// @Summary 删除待办事项
// @Description 删除指定的待办事项
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "待办事项ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "待办事项不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox/:id [delete]
func (h *InboxHandler) DeleteInboxItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "待办事项ID必须是有效数字")
		return
	}

	err = h.inboxService.DeleteInboxItem(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "删除待办事项失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "待办事项删除成功"})
}

// ListInboxItems godoc
// @Summary 获取待办事项列表
// @Description 分页获取当前用户的待办事项列表，支持过滤和搜索
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param is_read query bool false "是否已读"
// @Param is_completed query bool false "是否完成"
// @Param priority query string false "优先级" Enums(critical, high, medium, low)
// @Param source_type query string false "来源类型"
// @Param due_before query string false "到期日期之前" format(date)
// @Param due_after query string false "到期日期之后" format(date)
// @Param search query string false "搜索关键词"
// @Param order_by query string false "排序字段" Enums(due_date, priority, created_at)
// @Success 200 {object} common.APIResponse{data=services.ListInboxItemsResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox [get]
func (h *InboxHandler) ListInboxItems(c *gin.Context) {
	var req services.ListInboxItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "查询参数错误", err.Error())
		return
	}

	// 从上下文获取用户ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.APIUnauthorized(c, "未授权", "用户未登录")
		return
	}

	response, err := h.inboxService.ListInboxItems(c.Request.Context(), uint(userID), &req)
	if err != nil {
		common.APIInternalServerError(c, "获取待办事项列表失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// MarkAsRead godoc
// @Summary 标记为已读
// @Description 将指定待办事项标记为已读
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "待办事项ID"
// @Success 200 {object} common.APIResponse "标记成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "待办事项不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox/:id/read [put]
func (h *InboxHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "待办事项ID必须是有效数字")
		return
	}

	err = h.inboxService.MarkAsRead(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "标记已读失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "已标记为已读"})
}

// MarkAsCompleted godoc
// @Summary 标记为已完成
// @Description 将指定待办事项标记为已完成
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "待办事项ID"
// @Success 200 {object} common.APIResponse "标记成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "待办事项不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox/:id/complete [put]
func (h *InboxHandler) MarkAsCompleted(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "待办事项ID必须是有效数字")
		return
	}

	err = h.inboxService.MarkAsCompleted(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "标记完成失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "已标记为已完成"})
}

// SnoozeInboxItem godoc
// @Summary 延后待办事项
// @Description 将指定待办事项延后到指定时间
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "待办事项ID"
// @Param request body services.SnoozeInboxItemRequest true "延后请求"
// @Success 200 {object} common.APIResponse "延后成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "待办事项不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox/:id/snooze [put]
func (h *InboxHandler) SnoozeInboxItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "待办事项ID必须是有效数字")
		return
	}

	var req services.SnoozeInboxItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	err = h.inboxService.SnoozeInboxItem(c.Request.Context(), uint(id), &req)
	if err != nil {
		common.APIInternalServerError(c, "延后待办失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "待办事项已延后"})
}

// GetInboxStats godoc
// @Summary 获取待办事项统计
// @Description 获取当前用户的待办事项统计信息
// @Tags 待办事项
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.InboxStatsResponse} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /inbox/stats [get]
func (h *InboxHandler) GetInboxStats(c *gin.Context) {
	// 从上下文获取用户ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.APIUnauthorized(c, "未授权", "用户未登录")
		return
	}

	stats, err := h.inboxService.GetInboxStats(c.Request.Context(), uint(userID))
	if err != nil {
		common.APIInternalServerError(c, "获取统计信息失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}
