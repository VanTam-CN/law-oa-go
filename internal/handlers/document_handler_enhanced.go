package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/auth"
	"law-oa-go/internal/common"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// DocumentHandlerEnhanced handles enhanced document management operations
type DocumentHandlerEnhanced struct {
	docService        *services.DocumentService
	previewService    *services.DocumentPreviewService
	versionService    *services.DocumentVersionService
	permissionService *services.DocumentPermissionService
	recycleService    *services.DocumentRecycleService
	searchService     *services.DocumentSearchService
	statsService      *services.DocumentStatsService
}

// NewDocumentHandlerEnhanced creates a new enhanced document handler
func NewDocumentHandlerEnhanced(
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	storageDir string,
	recycleDir string,
) *DocumentHandlerEnhanced {
	docService := services.NewDocumentService(docRepo, storageDir)
	previewService := services.NewDocumentPreviewService(docRepo)
	versionService := services.NewDocumentVersionService(docRepo, storageDir)
	permissionService := services.NewDocumentPermissionService(docRepo, userRepo)
	recycleService := services.NewDocumentRecycleService(docRepo, recycleDir)
	searchService := services.NewDocumentSearchService(docRepo)
	statsService := services.NewDocumentStatsService(docRepo)

	return &DocumentHandlerEnhanced{
		docService:        docService,
		previewService:    previewService,
		versionService:    versionService,
		permissionService: permissionService,
		recycleService:    recycleService,
		searchService:     searchService,
		statsService:      statsService,
	}
}

// UploadDocument godoc
// @Summary Upload document
// @Description Upload a new document
// @Tags Document Management
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "Document name"
// @Param description formData string false "Document description"
// @Param category formData string false "Document category"
// @Param tags formData string false "Document tags (comma separated)"
// @Param entity_id formData int false "Entity ID"
// @Param entity_type formData string false "Entity type"
// @Param file formData file true "Document file"
// @Success 200 {object} common.APIResponse{data=services.Document} "Upload successful"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 401 {object} common.APIResponse "Unauthorized"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/upload [post]
func (h *DocumentHandlerEnhanced) UploadDocument(c *gin.Context) {
	// Use the existing document service upload functionality
	var req services.DocumentUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		common.APIBadRequest(c, "Invalid request", err.Error())
		return
	}
	if req.File == nil {
		file, err := c.FormFile("file")
		if err != nil {
			common.APIBadRequest(c, "Invalid request", "file is required")
			return
		}
		req.File = file
	}

	// Get user ID from context
	_, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "Unauthorized", "User not authenticated")
		return
	}

	document, err := h.docService.UploadDocument(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "Failed to upload document", err.Error())
		return
	}

	common.APISuccess(c, document)
}

// GetDocument godoc
// @Summary Get document
// @Description Get document by ID
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Document ID"
// @Success 200 {object} common.APIResponse{data=services.Document} "Document retrieved"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/{id} [get]
func (h *DocumentHandlerEnhanced) GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid document ID", "Document ID must be a valid integer")
		return
	}

	document, err := h.docService.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Document not found", err.Error())
		return
	}

	common.APISuccess(c, document)
}

// ListDocuments godoc
// @Summary List documents
// @Description List documents with pagination and filtering
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param category query string false "Document category"
// @Param entity_type query string false "Entity type"
// @Param entity_id query int false "Entity ID"
// @Param search query string false "Search term"
// @Param sort_by query string false "Sort field"
// @Param sort_order query string false "Sort order"
// @Success 200 {object} common.APIResponse{data=DocumentListResponse} "Documents retrieved"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents [get]
func (h *DocumentHandlerEnhanced) ListDocuments(c *gin.Context) {
	var req services.DocumentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	documents, total, err := h.docService.ListDocuments(c.Request.Context(), &req, auth.GetUserID(c))
	if err != nil {
		common.APIInternalServerError(c, "Failed to list documents", err.Error())
		return
	}

	response := map[string]interface{}{
		"documents": documents,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	}

	common.APISuccess(c, response)
}

// UpdateDocument godoc
// @Summary Update document
// @Description Update document metadata
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Document ID"
// @Param request body services.DocumentUpdateRequest true "Update request"
// @Success 200 {object} common.APIResponse{data=services.Document} "Document updated"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/{id} [put]
func (h *DocumentHandlerEnhanced) UpdateDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid document ID", "Document ID must be a valid integer")
		return
	}

	var req services.DocumentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request body", err.Error())
		return
	}

	document, err := h.docService.UpdateDocument(c.Request.Context(), uint(id), &req)
	if err != nil {
		if err.Error() == "Document not found" {
			common.APINotFound(c, "Document not found", err.Error())
		} else {
			common.APIInternalServerError(c, "Failed to update document", err.Error())
		}
		return
	}

	common.APISuccess(c, document)
}

// DeleteDocument godoc
// @Summary Delete document
// @Description Move document to recycle bin
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Document ID"
// @Success 200 {object} common.APIResponse "Document deleted"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/{id} [delete]
func (h *DocumentHandlerEnhanced) DeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid document ID", "Document ID must be a valid integer")
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "Unauthorized", "User not authenticated")
		return
	}

	// Move to recycle bin
	recycledDoc, err := h.recycleService.SoftDelete(c.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		if err.Error() == "Document not found" {
			common.APINotFound(c, "Document not found", err.Error())
		} else {
			common.APIInternalServerError(c, "Failed to delete document", err.Error())
		}
		return
	}

	common.APISuccess(c, recycledDoc)
}

// DownloadDocument godoc
// @Summary Download document
// @Description Download document file
// @Tags Document Management
// @Accept json
// @Produce application/octet-stream
// @Security BearerAuth
// @Param id path int true "Document ID"
// @Success 200 {file} "Document file"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/{id}/download [get]
func (h *DocumentHandlerEnhanced) DownloadDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid document ID", "Document ID must be a valid integer")
		return
	}

	file, document, err := h.docService.DownloadDocument(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "Document not found" {
			common.APINotFound(c, "Document not found", err.Error())
		} else {
			common.APIInternalServerError(c, "Failed to download document", err.Error())
		}
		return
	}

	// Set appropriate headers
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+document.Filename)
	c.Header("Content-Type", document.MimeType)

	// Stream file
	c.DataFromReader(http.StatusOK, document.Filesize, document.MimeType, file, nil)
}

// GetDocumentPreview godoc
// @Summary Get document preview
// @Description Get preview of a document
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Document ID"
// @Param page query int false "Page number (for PDF)"
// @Param width query int false "Preview width"
// @Param height query int false "Preview height"
// @Success 200 {object} common.APIResponse{data=services.PreviewResponse} "Preview generated"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/{id}/preview [get]
func (h *DocumentHandlerEnhanced) GetDocumentPreview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid document ID", "Document ID must be a valid integer")
		return
	}

	var req services.PreviewRequest
	req.DocumentID = uint(id)

	// Bind query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}
	if widthStr := c.Query("width"); widthStr != "" {
		if width, err := strconv.Atoi(widthStr); err == nil {
			req.Width = width
		}
	}
	if heightStr := c.Query("height"); heightStr != "" {
		if height, err := strconv.Atoi(heightStr); err == nil {
			req.Height = height
		}
	}

	preview, err := h.previewService.GetDocumentPreview(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "Document not found" {
			common.APINotFound(c, "Document not found", err.Error())
		} else {
			common.APIInternalServerError(c, "Failed to generate preview", err.Error())
		}
		return
	}

	common.APISuccess(c, preview)
}

// GetDocumentStats godoc
// @Summary Get document statistics
// @Description Get document statistics
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.DocumentStats} "Statistics retrieved"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/stats [get]
func (h *DocumentHandlerEnhanced) GetDocumentStats(c *gin.Context) {
	stats, err := h.docService.GetDocumentStats(c.Request.Context(), auth.GetUserID(c))
	if err != nil {
		common.APIInternalServerError(c, "Failed to get document statistics", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// GetDocumentVersions godoc
// @Summary Get document versions
// @Description Get all versions of a document
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Document ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} common.APIResponse{data=[]services.DocumentVersion} "Versions retrieved"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/{id}/versions [get]
func (h *DocumentHandlerEnhanced) GetDocumentVersions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid document ID", "Document ID must be a valid integer")
		return
	}

	var req services.VersionListRequest
	req.DocumentID = uint(id)

	// Bind query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	}

	versions, total, err := h.versionService.GetVersions(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "Failed to get document versions", err.Error())
		return
	}

	response := map[string]interface{}{
		"versions":  versions,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	}

	common.APISuccess(c, response)
}

// CreateDocumentVersion godoc
// @Summary Create document version
// @Description Create a new version of a document
// @Tags Document Management
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param document_id formData int true "Document ID"
// @Param name formData string true "Version name"
// @Param description formData string false "Version description"
// @Param changes formData string false "Version changes"
// @Param file formData file false "Version file"
// @Param created_by formData int true "Created by user ID"
// @Success 200 {object} common.APIResponse{data=services.DocumentVersion} "Version created"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/versions [post]
func (h *DocumentHandlerEnhanced) CreateDocumentVersion(c *gin.Context) {
	var req services.CreateVersionRequest
	if err := c.ShouldBind(&req); err != nil {
		common.APIBadRequest(c, "Invalid request", err.Error())
		return
	}

	version, err := h.versionService.CreateVersion(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "Failed to create version", err.Error())
		return
	}

	common.APISuccess(c, version)
}

// CompareDocumentVersions godoc
// @Summary Compare document versions
// @Description Compare two versions of a document
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CompareVersionsRequest true "Comparison request"
// @Success 200 {object} common.APIResponse{data=services.VersionComparison} "Versions compared"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/versions/compare [post]
func (h *DocumentHandlerEnhanced) CompareDocumentVersions(c *gin.Context) {
	var req services.CompareVersionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request", err.Error())
		return
	}

	comparison, err := h.versionService.CompareVersions(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "Failed to compare versions", err.Error())
		return
	}

	common.APISuccess(c, comparison)
}

// RestoreDocumentVersion godoc
// @Summary Restore document version
// @Description Restore document to a specific version
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Document ID"
// @Param version path int true "Version number"
// @Param restored_by query int true "Restored by user ID"
// @Success 200 {object} common.APIResponse "Document restored"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 404 {object} common.APIResponse "Document not found"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/{id}/versions/{version}/restore [post]
func (h *DocumentHandlerEnhanced) RestoreDocumentVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid document ID", "Document ID must be a valid integer")
		return
	}

	versionStr := c.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		common.APIBadRequest(c, "Invalid version number", "Version must be a valid integer")
		return
	}

	restoredByStr := c.Query("restored_by")
	restoredBy, err := strconv.ParseUint(restoredByStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid restored_by parameter", "Restored by must be a valid user ID")
		return
	}

	err = h.versionService.RestoreVersion(c.Request.Context(), int(id), version, uint(restoredBy))
	if err != nil {
		common.APIInternalServerError(c, "Failed to restore version", err.Error())
		return
	}

	common.APISuccess(c, map[string]interface{}{
		"message":     "Document restored successfully",
		"version":     version,
		"restored_by": restoredBy,
	})
}

// GetRecycleBin godoc
// @Summary Get recycle bin
// @Description Get documents in the recycle bin
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search term"
// @Param category query string false "Document category"
// @Param entity_type query string false "Entity type"
// @Param deleted_by query int false "Deleted by user ID"
// @Param expired_only query bool false "Show only expired items"
// @Success 200 {object} common.APIResponse{data=services.RecycleBinResponse} "Recycle bin retrieved"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/recycle-bin [get]
func (h *DocumentHandlerEnhanced) GetRecycleBin(c *gin.Context) {
	var req services.RecycleBinRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters", err.Error())
		return
	}

	recycleBin, err := h.recycleService.GetRecycleBin(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "Failed to get recycle bin", err.Error())
		return
	}

	common.APISuccess(c, recycleBin)
}

// RestoreDocuments godoc
// @Summary Restore documents
// @Description Restore documents from recycle bin
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.RestoreRequest true "Restore request"
// @Success 200 {object} common.APIResponse{data=services.RestoreResponse} "Documents restored"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/restore [post]
func (h *DocumentHandlerEnhanced) RestoreDocuments(c *gin.Context) {
	var req services.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request", err.Error())
		return
	}

	response, err := h.recycleService.RestoreDocuments(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "Failed to restore documents", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// PermanentlyDeleteDocuments godoc
// @Summary Permanently delete documents
// @Description Permanently delete documents from recycle bin
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.PermanentlyDeleteRequest true "Delete request"
// @Success 200 {object} common.APIResponse{data=services.PermanentlyDeleteResponse} "Documents deleted"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/permanent-delete [post]
func (h *DocumentHandlerEnhanced) PermanentlyDeleteDocuments(c *gin.Context) {
	var req services.PermanentlyDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request", err.Error())
		return
	}

	response, err := h.recycleService.PermanentlyDelete(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "Failed to permanently delete documents", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// EmptyRecycleBin godoc
// @Summary Empty recycle bin
// @Description Delete all expired documents from recycle bin
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deleted_by query int true "Deleted by user ID"
// @Success 200 {object} common.APIResponse{data=services.PermanentlyDeleteResponse} "Recycle bin emptied"
// @Failure 400 {object} common.APIResponse "Bad request"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/empty-recycle-bin [post]
func (h *DocumentHandlerEnhanced) EmptyRecycleBin(c *gin.Context) {
	deletedByStr := c.Query("deleted_by")
	deletedBy, err := strconv.ParseUint(deletedByStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid deleted_by parameter", "Deleted by must be a valid user ID")
		return
	}

	response, err := h.recycleService.EmptyRecycleBin(c.Request.Context(), uint(deletedBy))
	if err != nil {
		common.APIInternalServerError(c, "Failed to empty recycle bin", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// GetRecycleStats godoc
// @Summary Get recycle bin statistics
// @Description Get statistics about the recycle bin
// @Tags Document Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.RecycleStats} "Statistics retrieved"
// @Failure 500 {object} common.APIResponse "Internal error"
// @Router /documents/recycle-stats [get]
func (h *DocumentHandlerEnhanced) GetRecycleStats(c *gin.Context) {
	stats, err := h.recycleService.GetRecycleStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "Failed to get recycle bin statistics", err.Error())
		return
	}

	common.APISuccess(c, stats)
}
