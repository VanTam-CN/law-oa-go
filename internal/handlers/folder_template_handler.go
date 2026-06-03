package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// FolderTemplateHandler 卷宗目录模板处理器
type FolderTemplateHandler struct {
	service services.FolderTemplateService
}

// NewFolderTemplateHandler 创建卷宗目录模板处理器
func NewFolderTemplateHandler(service services.FolderTemplateService) *FolderTemplateHandler {
	return &FolderTemplateHandler{service: service}
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
	var req services.CreateFolderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

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
// @Param case_id path int true "案件ID"
// @Success 200 {object} common.APIResponse
// @Router /cases/{case_id}/folders [get]
func (h *FolderTemplateHandler) GetCaseFolders(c *gin.Context) {
	caseID, err := strconv.ParseUint(c.Param("case_id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的案件ID")
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
// @Param case_id path int true "案件ID"
// @Param request body services.CreateCaseFolderRequest true "文件夹请求"
// @Success 200 {object} common.APIResponse
// @Router /cases/{case_id}/folders [post]
func (h *FolderTemplateHandler) CreateCaseFolder(c *gin.Context) {
	caseID, err := strconv.ParseUint(c.Param("case_id"), 10, 64)
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
// @Param case_id path int true "案件ID"
// @Param folder_id path int true "文件夹ID"
// @Success 200 {object} common.APIResponse
// @Router /cases/{case_id}/folders/{folder_id} [delete]
func (h *FolderTemplateHandler) DeleteCaseFolder(c *gin.Context) {
	folderID, err := strconv.ParseUint(c.Param("folder_id"), 10, 64)
	if err != nil {
		common.APIBadRequest(c, "参数错误", "无效的文件夹ID")
		return
	}

	if err := h.service.DeleteCaseFolder(c.Request.Context(), uint(folderID)); err != nil {
		common.APIInternalServerError(c, "删除文件夹失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "删除成功"})
}
