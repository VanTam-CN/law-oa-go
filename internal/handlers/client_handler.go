package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type ClientHandler struct {
	clientService *services.ClientService
}

func NewClientHandler(clientService *services.ClientService) *ClientHandler {
	return &ClientHandler{
		clientService: clientService,
	}
}

// CreateClient godoc
// @Summary 创建客户
// @Description 创建新的客户记录
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateClientRequest true "创建客户请求"
// @Success 200 {object} common.Response{data=services.ClientResponse} "创建成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部错误"
// @Router /clients [post]
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req services.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	client, err := h.clientService.CreateClient(c.Request.Context(), &req)
	if err != nil {
		if common.IsValidationError(err) {
			common.BadRequest(c, "Validation failed: "+err.Error())
			return
		}
		common.InternalServerError(c, "Failed to create client: "+err.Error())
		return
	}

	common.Success(c, client)
}

// GetClient godoc
// @Summary 获取客户详情
// @Description 根据ID获取客户详细信息
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Success 200 {object} common.Response{data=services.ClientResponse} "获取成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 404 {object} common.Response "客户不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /clients/{id} [get]
func (h *ClientHandler) GetClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid client ID: Client ID must be a valid number")
		return
	}

	client, err := h.clientService.GetClientByID(c.Request.Context(), uint(id))
	if err != nil {
		if common.IsNotFoundError(err) {
			common.NotFound(c, "Client not found: The requested client does not exist")
			return
		}
		common.InternalServerError(c, "Failed to get client: "+err.Error())
		return
	}

	common.Success(c, client)
}

// UpdateClient godoc
// @Summary 更新客户信息
// @Description 更新指定客户的信息
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Param request body services.UpdateClientRequest true "更新客户请求"
// @Success 200 {object} common.Response{data=services.ClientResponse} "更新成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 404 {object} common.Response "客户不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /clients/{id} [put]
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid client ID: Client ID must be a valid number")
		return
	}

	var req services.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	client, err := h.clientService.UpdateClient(c.Request.Context(), uint(id), &req)
	if err != nil {
		if common.IsNotFoundError(err) {
			common.NotFound(c, "Client not found: The requested client does not exist")
			return
		}
		if common.IsValidationError(err) {
			common.BadRequest(c, "Validation failed: "+err.Error())
			return
		}
		common.InternalServerError(c, "Failed to update client: "+err.Error())
		return
	}

	common.Success(c, client)
}

// DeleteClient godoc
// @Summary 删除客户
// @Description 删除指定的客户
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Success 200 {object} common.Response "删除成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 404 {object} common.Response "客户不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid client ID: Client ID must be a valid number")
		return
	}

	err = h.clientService.DeleteClient(c.Request.Context(), uint(id))
	if err != nil {
		if common.IsNotFoundError(err) {
			common.NotFound(c, "Client not found: The requested client does not exist")
			return
		}
		common.InternalServerError(c, "Failed to delete client: "+err.Error())
		return
	}

	common.Success(c, gin.H{"message": "Client deleted successfully"})
}

// ListClients godoc
// @Summary 获取客户列表
// @Description 分页获取客户列表，支持过滤和搜索
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "客户状态" Enums(active,inactive)
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.PageResponse{data=[]services.ClientResponse} "获取成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部错误"
// @Router /clients [get]
func (h *ClientHandler) ListClients(c *gin.Context) {
	var req services.ClientListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	clients, total, err := h.clientService.ListClients(c.Request.Context(), &req)
	if err != nil {
		common.InternalServerError(c, "Failed to list clients: "+err.Error())
		return
	}

	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	response := common.PageResponse{
		Data:  clients,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetClientStats godoc
// @Summary 获取客户统计信息
// @Description 获取客户相关的统计数据
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.Response{data=services.ClientStatsResponse} "获取成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部错误"
// @Router /clients/stats [get]
func (h *ClientHandler) GetClientStats(c *gin.Context) {
	stats, err := h.clientService.GetClientStats(c.Request.Context())
	if err != nil {
		common.InternalServerError(c, "Failed to get client statistics: "+err.Error())
		return
	}

	common.Success(c, stats)
}
