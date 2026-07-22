package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

type CaseHandler struct {
	caseService           *services.CaseService
	authz                 *services.AuthorizationService
	ethicalWallListFilter bool
}

// SetEthicalWallListFilterEnabled enables SQL-side filtering for the case
// list. Detail middleware alone cannot protect pagination, search, or counts.
func (h *CaseHandler) SetEthicalWallListFilterEnabled(enabled bool) {
	h.ethicalWallListFilter = enabled
}

func NewCaseHandler(caseService *services.CaseService, authz ...*services.AuthorizationService) *CaseHandler {
	var authorizationService *services.AuthorizationService
	if len(authz) > 0 {
		authorizationService = authz[0]
	}
	return &CaseHandler{
		caseService: caseService,
		authz:       authorizationService,
	}
}

func (h *CaseHandler) requireAuthorization(c *gin.Context) bool {
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CASE_AUTHZ_UNAVAILABLE", "案件权限服务未初始化，当前不会返回或修改案件数据")
		return false
	}
	return true
}

// CreateCase godoc
// @Summary 创建案件
// @Description 创建新的法律案件
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateCaseRequest true "创建案件请求"
// @Success 200 {object} common.APIResponse{data=services.CaseResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases [post]
func (h *CaseHandler) CreateCase(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	var req services.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if req.AssignedBy == 0 {
		req.AssignedBy = actor.UserID
	}
	allowed, err := h.authz.CanCreateCase(c.Request.Context(), actor, &req)
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	caseResp, err := h.caseService.CreateCase(c.Request.Context(), &req)
	if err != nil {
		if isSubjectWorkflowError(err) {
			writeSubjectWorkflowError(c, err)
			return
		}
		common.APIInternalServerError(c, "创建案件失败", err.Error())
		return
	}

	common.APISuccess(c, caseResp)
}

// GetCase godoc
// @Summary 获取案件详情
// @Description 根据ID获取案件详细信息
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=services.CaseResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id} [get]
func (h *CaseHandler) GetCase(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
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
	allowed, err := h.authz.CanReadCase(c.Request.Context(), actor, uint(id))
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	caseResp, err := h.caseService.GetCaseByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "案件不存在", "指定的案件ID不存在")
		return
	}
	if role, _ := middleware.GetCurrentRole(c); role == "lawyer" {
		if userID, ok := middleware.GetCurrentUserID(c); ok && caseResp.LawyerID != userID {
			common.APIForbidden(c, "无权查看该案件", "律师账号只能查看本人承办的案件")
			return
		}
	}

	common.APISuccess(c, caseResp)
}

// UpdateCase godoc
// @Summary 更新案件
// @Description 更新指定案件的信息
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param request body services.UpdateCaseRequest true "更新案件请求"
// @Success 200 {object} common.APIResponse{data=services.CaseResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id} [put]
func (h *CaseHandler) UpdateCase(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "案件ID必须是有效数字")
		return
	}

	var req services.UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有字段")
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	allowed, err := h.authz.CanManageCase(c.Request.Context(), actor, uint(id))
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	caseResp, err := h.caseService.UpdateCase(c.Request.Context(), uint(id), &req)
	if err != nil {
		if isSubjectWorkflowError(err) {
			writeSubjectWorkflowError(c, err)
			return
		}
		common.APIInternalServerError(c, "更新案件失败", err.Error())
		return
	}

	common.APISuccess(c, caseResp)
}

// DeleteCase godoc
// @Summary 删除案件
// @Description 删除指定的案件
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id} [delete]
func (h *CaseHandler) DeleteCase(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
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
	allowed, err := h.authz.CanManageCase(c.Request.Context(), actor, uint(id))
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	err = h.caseService.DeleteCase(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "删除案件失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "案件删除成功"})
}

// ListCases godoc
// @Summary 获取案件列表
// @Description 分页获取案件列表，支持过滤和搜索
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.APIResponse{data=services.ListCasesResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases [get]
func (h *CaseHandler) ListCases(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	var req services.ListCasesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "查询参数错误", "请检查查询参数")
		return
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
		req.LawyerID = actor.UserID
	}

	response, err := h.caseService.ListCases(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "获取案件列表失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// GetLawyers godoc
// @Summary 获取律师列表
// @Description 获取律师用户列表，用于案件管理中的律师选择
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(100)
// @Success 200 {object} common.APIResponse{data=[]services.UserResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /lawfirm/lawyers [get]
func (h *CaseHandler) GetLawyers(c *gin.Context) {
	page := 1
	pageSize := 100

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	lawyers, err := h.caseService.GetLawyers(c.Request.Context(), page, pageSize)
	if err != nil {
		common.APIInternalServerError(c, "获取律师列表失败", err.Error())
		return
	}

	common.APISuccess(c, lawyers)
}

// GetLawyerByID godoc
// @Summary 获取律师详情
// @Description 获取单个律师用户详情，用于律师资源页面
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "律师ID"
// @Success 200 {object} common.APIResponse{data=models.User} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "律师不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /lawfirm/lawyers/{id} [get]
func (h *CaseHandler) GetLawyerByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		common.APIBadRequest(c, "请求参数错误", "律师ID必须是有效数字")
		return
	}

	lawyer, err := h.caseService.GetLawyerByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "律师不存在", "指定律师不存在或已删除")
		return
	}

	common.APISuccess(c, lawyer)
}

// GetCaseTypes godoc
// @Summary 获取案件类型列表
// @Description 获取所有可用的案件类型和案由
// @Tags 案件管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]services.CaseTypeResponse} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /case-types [get]
func (h *CaseHandler) GetCaseTypes(c *gin.Context) {
	caseTypes, err := h.caseService.GetCaseTypes(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取案件类型失败", err.Error())
		return
	}

	common.APISuccess(c, caseTypes)
}
