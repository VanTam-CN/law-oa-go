package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// FolderTemplateHandler 卷宗目录模板处理器
type FolderTemplateHandler struct {
	service services.FolderTemplateService
	authz   *services.AuthorizationService
}

// NewFolderTemplateHandler 创建卷宗目录模板处理器
func NewFolderTemplateHandler(service services.FolderTemplateService, authz ...*services.AuthorizationService) *FolderTemplateHandler {
	var authorizationService *services.AuthorizationService
	if len(authz) > 0 {
		authorizationService = authz[0]
	}
	return &FolderTemplateHandler{service: service, authz: authorizationService}
}

// CreateTemplate godoc
// @Summary 创建卷宗目录模板
// @Description 创建新的卷宗目录模板
// @Tags 卷宗目录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateFolderTemplateRequest true "模板请求"
// @Success 200 {object} common.APIResponse "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Router /folder-templates [post]
func (h *FolderTemplateHandler) CreateTemplate(c *gin.Context) {
	actor, ok := h.authorizeTemplateManager(c)
	if !ok {
		return
	}
	var req services.CreateFolderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	// The creator is an audit field, not caller-controlled business data.
	req.CreatedBy = actor.UserID

	template, err := h.service.CreateTemplate(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建模板失败", err.Error())
		return
	}

	common.APISuccess(c, template)
}

// ListTemplates godoc
// @Summary 查询卷宗目录模板列表
// @Tags 卷宗目录
// @Produce json
// @Security BearerAuth
// @Param case_type query string false "案件类型"
// @Param is_active query bool false "是否活跃"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} common.APIResponse
// @Router /folder-templates [get]
func (h *FolderTemplateHandler) ListTemplates(c *gin.Context) {
	params := &repositories.FolderTemplateListParams{
		CaseType: c.Query("case_type"),
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

	templates, total, err := h.service.ListTemplates(c.Request.Context(), params)
	if err != nil {
		common.APIInternalServerError(c, "查询模板失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"items":     templates,
		"total":     total,
		"page":      params.Page,
		"page_size": params.PageSize,
	})
}

// GetTemplate godoc
// @Summary 获取模板详情
// @Tags 卷宗目录
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} common.APIResponse
// @Router /folder-templates/{id} [get]
func (h *FolderTemplateHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的模板ID")
		return
	}

	template, err := h.service.GetTemplate(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取模板失败", err.Error())
		return
	}
	if template == nil {
		common.APINotFound(c, "模板不存在", "")
		return
	}

	common.APISuccess(c, template)
}

// UpdateTemplate godoc
// @Summary 更新模板
// @Tags 卷宗目录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Param request body services.UpdateFolderTemplateRequest true "更新请求"
// @Success 200 {object} common.APIResponse
// @Router /folder-templates/{id} [put]
func (h *FolderTemplateHandler) UpdateTemplate(c *gin.Context) {
	if _, ok := h.authorizeTemplateManager(c); !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的模板ID")
		return
	}

	var req services.UpdateFolderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	if err := h.service.UpdateTemplate(c.Request.Context(), uint(id), &req); err != nil {
		common.APIInternalServerError(c, "更新模板失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "更新成功"})
}

// DeleteTemplate godoc
// @Summary 删除模板
// @Tags 卷宗目录
// @Produce json
// @Security BearerAuth
// @Param id path int true "模板ID"
// @Success 200 {object} common.APIResponse
// @Router /folder-templates/{id} [delete]
func (h *FolderTemplateHandler) DeleteTemplate(c *gin.Context) {
	if _, ok := h.authorizeTemplateManager(c); !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的模板ID")
		return
	}

	if err := h.service.DeleteTemplate(c.Request.Context(), uint(id)); err != nil {
		common.APIInternalServerError(c, "删除模板失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "删除成功"})
}

// ApplyTemplate godoc
// @Summary 应用模板到案件
// @Description 将模板的文件夹结构应用到指定案件
// @Tags 卷宗目录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]interface{} true "应用请求 {case_id, template_id}"
// @Success 200 {object} common.APIResponse
// @Router /folder-templates/apply [post]
func (h *FolderTemplateHandler) ApplyTemplate(c *gin.Context) {
	var req struct {
		CaseID     uint `json:"case_id" binding:"required"`
		TemplateID uint `json:"template_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	if !h.authorizeCase(c, req.CaseID, true) {
		return
	}

	folders, err := h.service.ApplyTemplate(c.Request.Context(), req.CaseID, req.TemplateID)
	if err != nil {
		common.APIInternalServerError(c, "应用模板失败", err.Error())
		return
	}

	common.APISuccess(c, folders)
}

// GetCaseFolders godoc
// @Summary 获取案件卷宗目录
// @Description 获取指定案件的文件夹树形结构
// @Tags 卷宗目录
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse
// @Router /cases/{id}/folders [get]
func (h *FolderTemplateHandler) GetCaseFolders(c *gin.Context) {
	caseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的案件ID")
		return
	}
	if !h.authorizeCase(c, uint(caseID), false) {
		return
	}

	folders, err := h.service.GetCaseFolders(c.Request.Context(), uint(caseID))
	if err != nil {
		common.APIInternalServerError(c, "获取卷宗目录失败", err.Error())
		return
	}

	common.APISuccess(c, folders)
}

// CreateCaseFolder godoc
// @Summary 创建自定义文件夹
// @Tags 卷宗目录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param request body services.CreateCaseFolderRequest true "文件夹请求"
// @Success 200 {object} common.APIResponse
// @Router /cases/{id}/folders [post]
func (h *FolderTemplateHandler) CreateCaseFolder(c *gin.Context) {
	caseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的案件ID")
		return
	}

	var req services.CreateCaseFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	req.CaseID = uint(caseID)
	if !h.authorizeCase(c, uint(caseID), true) {
		return
	}

	folder, err := h.service.CreateCustomFolder(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建文件夹失败", err.Error())
		return
	}

	common.APISuccess(c, folder)
}

// DeleteCaseFolder godoc
// @Summary 删除案件文件夹
// @Tags 卷宗目录
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param folder_id path int true "文件夹ID"
// @Success 200 {object} common.APIResponse
// @Router /cases/{id}/folders/{folder_id} [delete]
func (h *FolderTemplateHandler) DeleteCaseFolder(c *gin.Context) {
	folderID, err := strconv.ParseUint(c.Param("folder_id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的文件夹ID")
		return
	}
	caseID, parseErr := strconv.ParseUint(c.Param("id"), 10, 64)
	if parseErr != nil || !h.authorizeCase(c, uint(caseID), true) {
		return
	}

	if err := h.service.DeleteCaseFolder(c.Request.Context(), uint(caseID), uint(folderID)); err != nil {
		if errors.Is(err, services.ErrFolderCaseMismatch) {
			forbidObjectAccess(c)
			return
		}
		if errors.Is(err, services.ErrFolderNotFound) {
			common.APINotFound(c, "文件夹不存在", "指定案件下不存在该文件夹")
			return
		}
		common.APIInternalServerError(c, "删除文件夹失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "删除成功"})
}

func (h *FolderTemplateHandler) authorizeCase(c *gin.Context, caseID uint, write bool) bool {
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CASE_AUTHZ_UNAVAILABLE", "案件权限服务未初始化")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	allowed, err := h.authz.CanReadCase(c.Request.Context(), actor, caseID)
	if write {
		allowed, err = h.authz.CanManageCase(c.Request.Context(), actor, caseID)
	}
	if err != nil {
		common.APIInternalServerError(c, "案件权限校验失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

func (h *FolderTemplateHandler) authorizeTemplateManager(c *gin.Context) (services.AuthActor, bool) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return services.AuthActor{}, false
	}
	if !services.IsMatterManagementRole(actor.Role) {
		common.APIForbidden(c, "无权修改卷宗目录模板", "只有律所管理或系统配置角色可以修改全所模板")
		return services.AuthActor{}, false
	}
	return actor, true
}
