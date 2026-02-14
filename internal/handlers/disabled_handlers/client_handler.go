package handlers

import (
	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
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
// @Success 200 {object} common.APIResponse{data=services.ClientResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients [post]
func (h *ClientHandler) CreateClient(c *gin.Context) {
	handler := APICreateHandler(func(c *gin.Context, req *services.CreateClientRequest) (*services.ClientResponse, error) {
		return h.clientService.CreateClient(c.Request.Context(), req)
	}, "client")
	handler(c)
}

// GetClient godoc
// @Summary 获取客户详情
// @Description 根据ID获取客户详细信息
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Success 200 {object} common.APIResponse{data=services.ClientResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "客户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients/{id} [get]
func (h *ClientHandler) GetClient(c *gin.Context) {
	handler := APIGetHandler(func(c *gin.Context, id uint) (*services.ClientResponse, error) {
		return h.clientService.GetClientByID(c.Request.Context(), id)
	}, "client")
	handler(c)
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
// @Success 200 {object} common.APIResponse{data=services.ClientResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "客户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients/{id} [put]
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	handler := APIUpdateHandler(func(c *gin.Context, id uint, req *services.UpdateClientRequest) (*services.ClientResponse, error) {
		return h.clientService.UpdateClient(c.Request.Context(), id, req)
	}, "client")
	handler(c)
}

// DeleteClient godoc
// @Summary 删除客户
// @Description 删除指定的客户
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "客户ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "客户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c *gin.Context) {
	handler := APIDeleteHandler(func(c *gin.Context, id uint) error {
		return h.clientService.DeleteClient(c.Request.Context(), id)
	}, "client")
	handler(c)
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
// @Success 200 {object} common.APIResponse{data=[]services.ClientResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients [get]
func (h *ClientHandler) ListClients(c *gin.Context) {
	handler := APIListHandler(func(c *gin.Context, req *services.ClientListRequest) ([]*services.ClientResponse, int64, error) {
		return h.clientService.ListClients(c.Request.Context(), req)
	}, "clients")
	handler(c)
}

// GetClientStats godoc
// @Summary 获取客户统计信息
// @Description 获取客户相关的统计数据
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.ClientStatsResponse} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients/stats [get]
func (h *ClientHandler) GetClientStats(c *gin.Context) {
	stats, err := h.clientService.GetClientStats(c.Request.Context())
	if err != nil {
		_ = c.Error(errors.InternalError("Failed to get client statistics", err))
		return
	}

	common.APISuccess(c, stats)
}
