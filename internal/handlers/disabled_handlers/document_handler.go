package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/services"
)

type DocumentHandler struct {
	documentService *services.DocumentService
}

func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// UploadDocument godoc
// @Summary 上传文档
// @Description 上传新的文档文件
// @Tags 文档管理
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param name formData string false "文档名称"
// @Param description formData string false "文档描述"
// @Param category formData string false "文档分类"
// @Param tags formData string false "标签（逗号分隔）"
// @Param entity_id formData int false "关联实体ID"
// @Param entity_type formData string false "关联实体类型"
// @Param file formData file true "文档文件"
// @Success 200 {object} common.APIResponse{data=services.Document} "上传成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /documents [post]
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	// Parse form data
	err := c.Request.ParseMultipartForm(50 << 20) // 50 MB max memory
	if err != nil {
		_ = c.Error(errors.ValidationError("form_parse", "Failed to parse form data: "+err.Error()))
		return
	}

	// Get form values
	name := c.PostForm("name")
	description := c.PostForm("description")
	category := c.PostForm("category")
	tags := c.PostForm("tags")
	entityIDStr := c.PostForm("entity_id")
	entityType := c.PostForm("entity_type")

	// Parse entity ID
	var entityID uint
	if entityIDStr != "" {
		id, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.ValidationError("entity_id", "Invalid entity ID: must be a valid number"))
			return
		}
		entityID = uint(id)
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		_ = c.Error(errors.ValidationError("file", "Missing file: "+err.Error()))
		return
	}

	// Create request
	req := &services.DocumentUploadRequest{
		Name:        name,
		Description: description,
		Category:    category,
		Tags:        tags,
		EntityID:    entityID,
		EntityType:  entityType,
		File:        file,
	}

	// Upload document
	document, err := h.documentService.UploadDocument(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, document)
}

// GetDocument godoc
// @Summary 获取文档详情
// @Description 根据ID获取文档详细信息
// @Tags 文档管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文档ID"
// @Success 200 {object} common.APIResponse{data=services.Document} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "文档不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /documents/{id} [get]
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.ValidationError("id_validation", "Invalid ID: must be a valid number"))
		return
	}

	document, err := h.documentService.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, document)
}

// UpdateDocument godoc
// @Summary 更新文档信息
// @Description 更新指定文档的信息
// @Tags 文档管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文档ID"
// @Param request body services.DocumentUpdateRequest true "更新文档请求"
// @Success 200 {object} common.APIResponse{data=services.Document} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "文档不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /documents/{id} [put]
func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.ValidationError("id_validation", "Invalid ID: must be a valid number"))
		return
	}

	var req services.DocumentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.ValidationError("request_binding", "Invalid request format: "+err.Error()))
		return
	}

	document, err := h.documentService.UpdateDocument(c.Request.Context(), uint(id), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, document)
}

// DeleteDocument godoc
// @Summary 删除文档
// @Description 删除指定的文档
// @Tags 文档管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文档ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "文档不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.ValidationError("id_validation", "Invalid ID: must be a valid number"))
		return
	}

	err = h.documentService.DeleteDocument(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, gin.H{"message": "Document deleted successfully"})
}

// ListDocuments godoc
// @Summary 获取文档列表
// @Description 分页获取文档列表，支持过滤和搜索
// @Tags 文档管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param category query string false "文档分类"
// @Param entity_type query string false "实体类型"
// @Param entity_id query int false "实体ID"
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.APIResponse{data=[]services.Document} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /documents [get]
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	var req services.DocumentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(errors.ValidationError("query_binding", "Invalid query parameters: "+err.Error()))
		return
	}

	documents, total, err := h.documentService.ListDocuments(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
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

	common.APISuccessWithPage(c, documents, total, page, pageSize)
}

// GetDocumentStats godoc
// @Summary 获取文档统计信息
// @Description 获取文档相关的统计数据
// @Tags 文档管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.DocumentStats} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /documents/stats [get]
func (h *DocumentHandler) GetDocumentStats(c *gin.Context) {
	stats, err := h.documentService.GetDocumentStats(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, stats)
}

// DownloadDocument godoc
// @Summary 下载文档
// @Description 下载指定的文档文件
// @Tags 文档管理
// @Accept json
// @Produce octet-stream
// @Security BearerAuth
// @Param id path int true "文档ID"
// @Success 200 {file} binary "下载成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "文档不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /documents/{id}/download [get]
func (h *DocumentHandler) DownloadDocument(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.ValidationError("id_validation", "Invalid ID: must be a valid number"))
		return
	}

	// For now, we'll just send a mock file since we don't have the actual implementation
	// In a real implementation, you would use the actual file service
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=document.pdf")
	c.Header("Content-Type", "application/pdf")

	// Send mock content
	content := "Mock document content for document ID " + idStr
	c.Data(http.StatusOK, "application/pdf", []byte(content))
}
