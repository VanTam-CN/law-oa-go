package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	apperrors "law-oa-go/internal/errors"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

type ClientHandler struct {
	clientService         *services.ClientService
	contactService        *services.ClientContactService
	authz                 *services.AuthorizationService
	ethicalWallListFilter bool
}

func (h *ClientHandler) SetContactService(service *services.ClientContactService) {
	h.contactService = service
}

// SetEthicalWallListFilterEnabled enables the SQL-side ethical-wall predicate
// for client list/count queries. It is configured by the production router
// after the ethical-wall repository is initialized.
func (h *ClientHandler) SetEthicalWallListFilterEnabled(enabled bool) {
	h.ethicalWallListFilter = enabled
}

func NewClientHandler(clientService *services.ClientService, authz ...*services.AuthorizationService) *ClientHandler {
	var authorizationService *services.AuthorizationService
	if len(authz) > 0 {
		authorizationService = authz[0]
	}
	return &ClientHandler{
		clientService: clientService,
		authz:         authorizationService,
	}
}

func (h *ClientHandler) requireAuthorization(c *gin.Context) bool {
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CLIENT_AUTHZ_UNAVAILABLE", "客户权限服务未初始化，当前不会返回或修改客户数据")
		return false
	}
	return true
}

// CreateClient godoc
// @Summary 创建客户
// @Description 创建新的客户
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateClientRequest true "创建客户请求"
// @Param Idempotency-Key header string false "同次创建重试键，1至120个可见ASCII字符（不含空格）"
// @Success 200 {object} common.APIResponse{data=services.ClientResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 409 {object} common.APIResponse "幂等键对应另一份创建请求"
// @Failure 503 {object} common.APIResponse "身份保护或幂等存储不可用"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients [post]
func (h *ClientHandler) CreateClient(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	var req services.CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	allowed, err := h.authz.CanCreateClient(c.Request.Context(), actor)
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}
	req.CreatedBy = actor.UserID

	keys := c.Request.Header.Values("Idempotency-Key")
	var client *services.ClientResponse
	if len(keys) > 0 {
		if len(keys) != 1 || services.ValidateClientIdempotencyKey(keys[0]) != nil {
			common.NewAPIError(c, http.StatusBadRequest, "CLIENT_CREATE_KEY_INVALID", "Idempotency-Key 必须是1至120个可见ASCII字符，不能包含空格或重复请求头")
			return
		}
		client, err = h.clientService.CreateClientWithIdempotencyKey(c.Request.Context(), &req, actor.UserID, keys[0])
	} else {
		client, err = h.clientService.CreateClient(c.Request.Context(), &req)
	}
	if err != nil {
		if errors.Is(err, security.ErrSubjectDataKeyMissing) || errors.Is(err, security.ErrSubjectDataKeyInvalid) {
			common.NewAPIError(c, http.StatusServiceUnavailable, "SUBJECT_DATA_KEY_REQUIRED", "客户身份保护服务未配置，已阻止保存")
			return
		}
		if len(keys) > 0 {
			var appErr *apperrors.EnhancedError
			switch {
			case errors.Is(err, services.ErrClientIdempotencyConflict):
				common.NewAPIError(c, http.StatusConflict, "CLIENT_CREATE_KEY_CONFLICT", "该创建请求键已用于另一份资料，请核对后重试")
			case errors.Is(err, services.ErrClientIdempotencyIncomplete):
				common.NewAPIError(c, http.StatusConflict, "CLIENT_CREATE_RESULT_UNAVAILABLE", "原创建结果暂不可用，请联系有权人员核实，勿重复建档")
			case errors.As(err, &appErr) && (appErr.Category() == apperrors.ErrorCategoryValidation || appErr.Category() == apperrors.ErrorCategoryBusiness):
				// Do not expose another client's existence or storage error details.
				common.NewAPIError(c, http.StatusBadRequest, "CLIENT_CREATE_REJECTED", "无法创建客户，请核对资料或联系有权人员处理")
			default:
				common.NewAPIError(c, http.StatusServiceUnavailable, "CLIENT_CREATE_UNAVAILABLE", "客户创建服务暂不可用，请保留原请求键重试")
			}
			return
		}
		common.APIInternalServerError(c, "创建客户失败", err.Error())
		return
	}
	if len(keys) > 0 {
		allowed, err := h.authz.CanReadClient(c.Request.Context(), actor, client.ID)
		if err != nil {
			common.NewAPIError(c, http.StatusServiceUnavailable, "CLIENT_AUTHZ_UNAVAILABLE", "客户权限检查暂不可用，请保留原请求键重试")
			return
		}
		if !allowed {
			forbidObjectAccess(c)
			return
		}
	}

	common.APISuccess(c, client)
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
	if !h.requireAuthorization(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "客户ID必须是有效数字")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	allowed, err := h.authz.CanReadClient(c.Request.Context(), actor, uint(id))
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	client, err := h.clientService.GetClientByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "客户不存在", "指定的客户ID不存在")
		return
	}

	common.APISuccess(c, client)
}

// GetPrimaryContact returns the independently stored primary contact after
// the same object-level authorization used by the client master record.
func (h *ClientHandler) GetPrimaryContact(c *gin.Context) {
	clientID, ok := h.authorizedContactClient(c, false)
	if !ok {
		return
	}
	if h.contactService == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CLIENT_CONTACT_SERVICE_UNAVAILABLE", "客户联系人服务未初始化")
		return
	}
	contact, err := h.contactService.GetPrimaryContact(c.Request.Context(), clientID)
	if err != nil {
		if errors.Is(err, security.ErrSubjectDataKeyMissing) || errors.Is(err, security.ErrSubjectDataKeyInvalid) {
			common.NewAPIError(c, http.StatusServiceUnavailable, "SUBJECT_DATA_KEY_REQUIRED", "客户联系人保护服务未配置")
			return
		}
		common.APIInternalServerError(c, "读取主联系人失败", err.Error())
		return
	}
	common.APISuccess(c, contact)
}

// SavePrimaryContact creates or updates one primary contact without changing
// the client email, notes, or other master-data fields.
func (h *ClientHandler) SavePrimaryContact(c *gin.Context) {
	clientID, ok := h.authorizedContactClient(c, true)
	if !ok {
		return
	}
	if h.contactService == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CLIENT_CONTACT_SERVICE_UNAVAILABLE", "客户联系人服务未初始化")
		return
	}
	var req services.SavePrimaryContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "联系人信息无效", "请检查姓名、电话和邮箱格式")
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	contact, err := h.contactService.SavePrimaryContact(c.Request.Context(), clientID, actor.UserID, req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrClientContactVersionConflict):
			common.NewAPIError(c, http.StatusConflict, "CLIENT_CONTACT_VERSION_CONFLICT", "主联系人已被他人更新，请刷新后重试")
		case errors.Is(err, security.ErrSubjectDataKeyMissing), errors.Is(err, security.ErrSubjectDataKeyInvalid):
			common.NewAPIError(c, http.StatusServiceUnavailable, "SUBJECT_DATA_KEY_REQUIRED", "客户联系人保护服务未配置，已阻止保存")
		default:
			common.APIBadRequest(c, "保存主联系人失败", err.Error())
		}
		return
	}
	common.APISuccess(c, contact)
}

func (h *ClientHandler) authorizedContactClient(c *gin.Context, manage bool) (uint, bool) {
	if !h.requireAuthorization(c) {
		return 0, false
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		common.APIBadRequest(c, "请求参数错误", "客户ID必须是有效数字")
		return 0, false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return 0, false
	}
	var allowed bool
	if manage {
		allowed, err = h.authz.CanManageClient(c.Request.Context(), actor, uint(id))
	} else {
		allowed, err = h.authz.CanReadClient(c.Request.Context(), actor, uint(id))
	}
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return 0, false
	}
	if !allowed {
		forbidObjectAccess(c)
		return 0, false
	}
	return uint(id), true
}

// UpdateClient godoc
// @Summary 更新客户
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
// @Failure 409 {object} common.APIResponse "版本冲突"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients/{id} [put]
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "客户ID必须是有效数字")
		return
	}

	var req services.UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有字段")
		return
	}
	if req.Version == nil || *req.Version == 0 {
		common.APIBadRequest(c, "请求参数错误", "更新客户必须提交有效的version")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	allowed, err := h.authz.CanManageClient(c.Request.Context(), actor, uint(id))
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	client, err := h.clientService.UpdateClient(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, services.ErrClientVersionConflict) {
			common.NewAPIError(c, http.StatusConflict, "CLIENT_VERSION_CONFLICT", "客户信息已被他人更新，请刷新后重试")
			return
		}
		if errors.Is(err, security.ErrSubjectDataKeyMissing) || errors.Is(err, security.ErrSubjectDataKeyInvalid) {
			common.NewAPIError(c, http.StatusServiceUnavailable, "SUBJECT_DATA_KEY_REQUIRED", "客户身份保护服务未配置，已阻止更新")
			return
		}
		common.APIInternalServerError(c, "更新客户失败", err.Error())
		return
	}

	common.APISuccess(c, client)
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
	if !h.requireAuthorization(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "客户ID必须是有效数字")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	allowed, err := h.authz.CanManageClient(c.Request.Context(), actor, uint(id))
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	err = h.clientService.DeleteClient(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "删除客户失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "客户删除成功"})
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
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.APIResponse{data=services.ListClientsResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients [get]
func (h *ClientHandler) ListClients(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	var req services.ClientListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查查询参数")
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	var actor services.AuthActor
	var actorOK bool
	actor, actorOK = currentAuthActor(c)
	if !actorOK {
		return
	}
	if h.ethicalWallListFilter {
		req.EthicalWallUserID = actor.UserID
	}
	if !services.IsBusinessMatterManagementRole(actor.Role) {
		req.AccessibleByUserID = actor.UserID
	}

	clients, total, err := h.clientService.ListClients(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "获取客户列表失败", err.Error())
		return
	}

	response := gin.H{
		"clients": clients,
		"pagination": gin.H{
			"page":       req.Page,
			"page_size":  req.PageSize,
			"total":      total,
			"total_page": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		},
	}

	common.APISuccess(c, response)
}

// GetClientStats godoc
// @Summary 获取客户统计信息
// @Description 获取客户相关的统计数据
// @Tags 客户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=ClientStats} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /clients/stats [get]
func (h *ClientHandler) GetClientStats(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	if !canViewAllMatterData(c) {
		forbidObjectAccess(c)
		return
	}

	stats, err := h.clientService.GetClientStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取客户统计失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}
