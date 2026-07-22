package handlers

import (
	"net/http"
	"strconv"
	"strings"

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
	authz             *services.AuthorizationService
	subjectRecheck    *services.SubjectRecheckService
}

// SetSubjectRecheckService installs the server-side gate for case-bound
// document writes and generated outputs. Reads remain available to the
// authorized matter team while a subject revision is under review.
func (h *DocumentHandlerEnhanced) SetSubjectRecheckService(service *services.SubjectRecheckService) {
	h.subjectRecheck = service
}

func (h *DocumentHandlerEnhanced) requireCaseSubjectAction(c *gin.Context, entityType string, entityID uint, action string) bool {
	if !strings.EqualFold(strings.TrimSpace(entityType), "case") || entityID == 0 {
		return true
	}
	if h.subjectRecheck == nil {
		writeSubjectWorkflowError(c, services.NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "案件文档受控动作门禁未初始化，已阻止操作"))
		return false
	}
	if err := h.subjectRecheck.RequireEffectiveSubject(c.Request.Context(), entityID, action); err != nil {
		writeSubjectWorkflowError(c, err)
		return false
	}
	return true
}

// NewDocumentHandlerEnhanced creates a new enhanced document handler
func NewDocumentHandlerEnhanced(
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	storageDir string,
	recycleDir string,
	authz ...*services.AuthorizationService,
) *DocumentHandlerEnhanced {
	docService := services.NewDocumentService(docRepo, storageDir)
	previewService := services.NewDocumentPreviewService(docRepo)
	versionService := services.NewDocumentVersionService(docRepo, storageDir)
	permissionService := services.NewDocumentPermissionService(docRepo, userRepo)
	recycleService := services.NewDocumentRecycleService(docRepo, recycleDir)
	searchService := services.NewDocumentSearchService(docRepo)
	statsService := services.NewDocumentStatsService(docRepo)
	var authorizationService *services.AuthorizationService
	if len(authz) > 0 {
		authorizationService = authz[0]
	}

	return &DocumentHandlerEnhanced{
		docService:        docService,
		previewService:    previewService,
		versionService:    versionService,
		permissionService: permissionService,
		recycleService:    recycleService,
		searchService:     searchService,
		statsService:      statsService,
		authz:             authorizationService,
	}
}

func (h *DocumentHandlerEnhanced) authorizeDocumentAccess(c *gin.Context, documentID uint, write bool) bool {
	actor, ok := h.requireDocumentAuthorization(c)
	if !ok {
		return false
	}
	var (
		allowed bool
		err     error
	)
	if write {
		allowed, err = h.authz.CanManageDocument(c.Request.Context(), actor, documentID)
	} else {
		allowed, err = h.authz.CanReadDocument(c.Request.Context(), actor, documentID)
	}
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

// requireDocumentAuthorization prevents a missing dependency from turning a
// document route into an unscoped read or write path. Every document HTTP
// operation must have both the authorization service and an authenticated
// actor before it reaches the repository.
func (h *DocumentHandlerEnhanced) requireDocumentAuthorization(c *gin.Context) (services.AuthActor, bool) {
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "DOCUMENT_AUTHZ_UNAVAILABLE", "文档权限服务未初始化，当前已阻止文档操作")
		return services.AuthActor{}, false
	}
	return currentAuthActor(c)
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

	actor, ok := h.requireDocumentAuthorization(c)
	if !ok {
		return
	}
	allowed, err := h.authz.CanCreateDocument(c.Request.Context(), actor, req.EntityType, req.EntityID)
	if err != nil {
		common.APIInternalServerError(c, "权限检查失败", err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}
	if !h.requireCaseSubjectAction(c, req.EntityType, req.EntityID, "case_document_upload") {
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

	if !h.authorizeDocumentAccess(c, uint(id), false) {
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

	actor, ok := h.requireDocumentAuthorization(c)
	if !ok {
		return
	}
	if services.IsTechnicalAdminRole(actor.Role) {
		forbidObjectAccess(c)
		return
	}
	// Keep the actor ID on every HTTP query, including management queries, so
	// the repository applies the ethical-wall predicate before counting or
	// paginating rows. Management access is matter-wide only after the wall
	// check; it is not a bypass of the wall.
	viewerUserID := actor.UserID
	req.OwnerScoped = !services.IsBusinessMatterManagementRole(actor.Role)
	if req.EntityType == "case" && req.EntityID > 0 {
		allowed, err := h.authz.CanReadCase(c.Request.Context(), actor, req.EntityID)
		if err != nil {
			common.APIInternalServerError(c, "权限检查失败", err.Error())
			return
		}
		if !allowed {
			forbidObjectAccess(c)
			return
		}
	}

	documents, total, err := h.docService.ListDocuments(c.Request.Context(), &req, viewerUserID)
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

	if !h.authorizeDocumentAccess(c, uint(id), true) {
		return
	}
	document, err := h.docService.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Document not found", err.Error())
		return
	}
	if !h.requireCaseSubjectAction(c, document.EntityType, document.EntityID, "case_document_metadata_update") {
		return
	}

	document, err = h.docService.UpdateDocument(c.Request.Context(), uint(id), &req)
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

	if !h.authorizeDocumentAccess(c, uint(id), true) {
		return
	}
	document, err := h.docService.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Document not found", err.Error())
		return
	}
	if !h.requireCaseSubjectAction(c, document.EntityType, document.EntityID, "case_document_delete") {
		return
	}
	common.NewAPIError(c, http.StatusConflict, "DOCUMENT_RETENTION_CONTROLLED", "案件文档受保留策略控制，当前版本不提供删除或回收站操作")
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

	if !h.authorizeDocumentAccess(c, uint(id), false) {
		return
	}
	document, err := h.docService.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Document not found", err.Error())
		return
	}
	if !h.requireCaseSubjectAction(c, document.EntityType, document.EntityID, "case_document_download") {
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

	if !h.authorizeDocumentAccess(c, uint(id), false) {
		return
	}
	document, err := h.docService.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Document not found", err.Error())
		return
	}
	if !h.requireCaseSubjectAction(c, document.EntityType, document.EntityID, "case_document_preview") {
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
	if !h.requireDocumentAggregateAccess(c) {
		return
	}
	common.NewAPIError(c, http.StatusServiceUnavailable, "DOCUMENT_STATS_UNAVAILABLE", "文档统计尚未接入真实审计与存储指标，当前不可用")
}

func (h *DocumentHandlerEnhanced) requireDocumentAggregateAccess(c *gin.Context) bool {
	actor, ok := h.requireDocumentAuthorization(c)
	if !ok {
		return false
	}
	role := actor.Role
	if !services.IsBusinessMatterManagementRole(role) {
		common.NewAPIError(c, http.StatusForbidden, "DOCUMENT_AGGREGATE_FORBIDDEN", "文档聚合统计仅限业务管理角色查看")
		return false
	}
	return true
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
	if !h.authorizeDocumentAccess(c, uint(id), false) {
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
	viewerID := auth.GetUserID(c)
	if viewerID == 0 {
		common.APIUnauthorized(c, "Unauthorized", "User not authenticated")
		return
	}
	if req.CreatedBy != 0 && req.CreatedBy != viewerID {
		forbidObjectAccess(c)
		return
	}
	req.CreatedBy = viewerID
	if !h.authorizeDocumentAccess(c, req.DocumentID, true) {
		return
	}
	document, err := h.docService.GetDocumentByID(c.Request.Context(), req.DocumentID)
	if err != nil {
		common.APINotFound(c, "Document not found", err.Error())
		return
	}
	if !h.requireCaseSubjectAction(c, document.EntityType, document.EntityID, "case_document_version_create") {
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
	if !h.authorizeDocumentAccess(c, req.DocumentID, false) {
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
	if !h.authorizeDocumentAccess(c, uint(id), true) {
		return
	}
	document, err := h.docService.GetDocumentByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Document not found", err.Error())
		return
	}
	if !h.requireCaseSubjectAction(c, document.EntityType, document.EntityID, "case_document_version_restore") {
		return
	}

	versionStr := c.Param("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		common.APIBadRequest(c, "Invalid version number", "Version must be a valid integer")
		return
	}

	viewerID := auth.GetUserID(c)
	if viewerID == 0 {
		common.APIUnauthorized(c, "Unauthorized", "User not authenticated")
		return
	}
	restoredBy := uint64(viewerID)
	if restoredByStr := c.Query("restored_by"); restoredByStr != "" {
		requestedBy, parseErr := strconv.ParseUint(restoredByStr, 10, 32)
		if parseErr != nil {
			common.APIBadRequest(c, "Invalid restored_by parameter", "Restored by must be a valid user ID")
			return
		}
		if requestedBy != restoredBy {
			forbidObjectAccess(c)
			return
		}
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
	common.NewAPIError(c, http.StatusServiceUnavailable, "DOCUMENT_RECYCLE_UNAVAILABLE", "文档回收站尚未接入真实保留台账，当前不可用")
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
	common.NewAPIError(c, http.StatusServiceUnavailable, "DOCUMENT_RECYCLE_UNAVAILABLE", "文档回收站尚未接入真实保留台账，当前不可用")
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
	common.NewAPIError(c, http.StatusConflict, "DOCUMENT_RETENTION_CONTROLLED", "案件文档受保留策略控制，当前版本禁止物理删除")
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
	common.NewAPIError(c, http.StatusConflict, "DOCUMENT_RETENTION_CONTROLLED", "案件文档受保留策略控制，当前版本禁止清空回收站")
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
	common.NewAPIError(c, http.StatusServiceUnavailable, "DOCUMENT_RECYCLE_UNAVAILABLE", "文档回收站尚未接入真实保留台账，当前不可用")
}
