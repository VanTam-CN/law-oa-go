package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// documentService 文档服务实现
type documentService struct {
	docRepo       repositories.DocumentRepository
	versionRepo    repositories.DocumentVersionRepository
	permissionRepo repositories.DocumentPermissionRepository
	auditRepo     repositories.DocumentAuditRepository
	userRepo      repositories.UserRepository
	roleRepo      repositories.RoleRepository
	storageService StorageService
	logger        *logrus.Logger
}

// NewDocumentService 创建新的文档服务
func NewDocumentService(
	docRepo repositories.DocumentRepository,
	versionRepo repositories.DocumentVersionRepository,
	permissionRepo repositories.DocumentPermissionRepository,
	auditRepo repositories.DocumentAuditRepository,
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	storageService StorageService,
	logger *logrus.Logger,
) DocumentService {
	return &documentService{
		docRepo:       docRepo,
		versionRepo:    versionRepo,
		permissionRepo: permissionRepo,
		auditRepo:     auditRepo,
		userRepo:      userRepo,
		roleRepo:      roleRepo,
		storageService: storageService,
		logger:        logger,
	}
}

// CreateDocument 创建文档
func (s *documentService) CreateDocument(ctx context.Context, req *CreateDocumentRequest) (*DocumentResponse, error) {
	// 生成UUID
	documentUUID := uuid.New().String()

	// 创建文档模型
	document := &models.Document{
		UUID:        documentUUID,
		TenantID:    req.TenantID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        strings.Join(req.Tags, ","),
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		Status:      "active",
		CreatedBy:   req.CreatedBy,
	}

	// 使用事务创建文档和权限
	err := s.docRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 创建文档
		if err := s.docRepo.Create(ctx, document); err != nil {
			return fmt.Errorf("failed to create document: %w", err)
		}

		// 给创建者分配所有权限
		permission := &models.DocumentPermission{
			DocumentID: document.ID,
			UserID:     &req.CreatedBy,
			TenantID:   req.TenantID,
			Permission: "admin",
		}

		if err := s.permissionRepo.Create(ctx, permission); err != nil {
			return fmt.Errorf("failed to assign permissions: %w", err)
		}

		// 记录审计日志
		auditReq := &LogActionRequest{
			DocumentID: documentUUID,
			UserID:     fmt.Sprintf("%d", req.CreatedBy),
			Action:     "create",
			Details:    fmt.Sprintf("Created document: %s", req.Name),
			TenantID:   req.TenantID,
		}

		if err := s.logAudit(ctx, auditReq); err != nil {
			s.logger.WithError(err).Warn("Failed to log audit action")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 转换为响应
	return s.convertToDocumentResponse(document, false)
}

// GetDocument 根据UUID获取文档
func (s *documentService) GetDocument(ctx context.Context, documentUUID string) (*DocumentResponse, error) {
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, err
	}

	return s.convertToDocumentResponse(document, true)
}

// GetDocumentByID 根据ID获取文档
func (s *documentService) GetDocumentByID(ctx context.Context, documentID uint) (*DocumentResponse, error) {
	document, err := s.docRepo.GetByID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	return s.convertToDocumentResponse(document, true)
}

// UpdateDocument 更新文档
func (s *documentService) UpdateDocument(ctx context.Context, documentUUID string, req *UpdateDocumentRequest) (*DocumentResponse, error) {
	// 获取现有文档
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		document.Name = req.Name
	}
	if req.Description != "" {
		document.Description = req.Description
	}
	if req.Category != "" {
		document.Category = req.Category
	}
	if req.Tags != nil {
		document.Tags = strings.Join(req.Tags, ",")
	}
	if req.EntityType != "" {
		document.EntityType = req.EntityType
	}
	if req.EntityID > 0 {
		document.EntityID = req.EntityID
	}

	// 保存更新
	if err := s.docRepo.Update(ctx, document); err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: documentUUID,
		Action:     "update",
		Details:    fmt.Sprintf("Updated document: %s", document.Name),
		TenantID:   document.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return s.convertToDocumentResponse(document, false)
}

// DeleteDocument 删除文档
func (s *documentService) DeleteDocument(ctx context.Context, documentUUID string) error {
	// 获取文档信息用于审计
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return err
	}

	// 使用事务删除文档和相关数据
	err = s.docRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 软删除文档
		if err := s.docRepo.SoftDelete(ctx, document.ID); err != nil {
			return fmt.Errorf("failed to delete document: %w", err)
		}

		// 删除所有权限
		if err := s.permissionRepo.BatchDeleteByDocument(ctx, document.ID); err != nil {
			return fmt.Errorf("failed to delete permissions: %w", err)
		}

		// 记录审计日志
		auditReq := &LogActionRequest{
			DocumentID: documentUUID,
			Action:     "delete",
			Details:    fmt.Sprintf("Deleted document: %s", document.Name),
			TenantID:   document.TenantID,
		}

		if err := s.logAudit(ctx, auditReq); err != nil {
			s.logger.WithError(err).Warn("Failed to log audit action")
		}

		return nil
	})

	return err
}

// ListDocuments 列出文档
func (s *documentService) ListDocuments(ctx context.Context, filter *DocumentFilter) (*DocumentListResponse, error) {
	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// 转换过滤器
	repoFilter := s.convertToDocumentFilter(filter)

	// 查询文档
	documents, total, err := s.docRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, err
	}

	// 转换为响应
	docResponses := make([]*DocumentResponse, len(documents))
	for i, doc := range documents {
		response, err := s.convertToDocumentResponse(doc, false)
		if err != nil {
			return nil, err
		}
		docResponses[i] = response
	}

	totalPages := int(total)/filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &DocumentListResponse{
		Documents: docResponses,
		Total:     total,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchDocuments 搜索文档
func (s *documentService) SearchDocuments(ctx context.Context, query *SearchRequest) (*SearchResponse, error) {
	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// 转换搜索请求
	repoFilter := &repositories.DocumentFilter{
		TenantID:   query.TenantID,
		Category:   query.Category,
		Tags:       query.Tags,
		CreatedBy:  query.Creator,
		StartDate:  query.StartDate,
		EndDate:    query.EndDate,
		Page:       query.Page,
		PageSize:   query.PageSize,
		SortBy:     query.SortBy,
		SortOrder:  query.SortOrder,
	}

	// 根据搜索方式处理
	var documents []*models.Document
	var total int64
	var err error

	if query.Query != "" {
		// 使用内容搜索
		documents, err = s.docRepo.SearchByContent(ctx, query.TenantID, query.Query)
		total = int64(len(documents))
	} else {
		// 使用过滤器查询
		documents, total, err = s.docRepo.List(ctx, repoFilter)
	}

	if err != nil {
		return nil, err
	}

	// 转换为响应
	docResponses := make([]*DocumentResponse, len(documents))
	for i, doc := range documents {
		response, err := s.convertToDocumentResponse(doc, false)
		if err != nil {
			return nil, err
		}
		docResponses[i] = response
	}

	totalPages := int(total)/query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	return &SearchResponse{
		Documents: docResponses,
		Total:     total,
		Page:      query.Page,
		PageSize:  query.PageSize,
		TotalPages: totalPages,
		Took:      0, // 这里可以根据实际的搜索时间来设置
	}, nil
}

// GetRecentDocuments 获取最近的文档
func (s *documentService) GetRecentDocuments(ctx context.Context, tenantID string, limit int) ([]*DocumentResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	documents, err := s.docRepo.GetRecentDocuments(ctx, tenantID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*DocumentResponse, len(documents))
	for i, doc := range documents {
		response, err := s.convertToDocumentResponse(doc, false)
		if err != nil {
			return nil, err
		}
		responses[i] = response
	}

	return responses, nil
}

// CreateVersion 创建文档版本
func (s *documentService) CreateVersion(ctx context.Context, documentUUID string, req *CreateVersionRequest) (*DocumentVersionResponse, error) {
	// 获取文档
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, err
	}

	// 计算文件哈希
	fileHash := req.FileHash
	if fileHash == "" {
		hash := sha256.Sum256(req.File)
		fileHash = hex.EncodeToString(hash[:])
	}

	// 获取下一个版本号
	nextVersion, err := s.versionRepo.GetNextVersionNumber(ctx, document.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next version number: %w", err)
	}

	// 创建版本
	version := &models.DocumentVersion{
		DocumentID:  document.ID,
		Version:     nextVersion,
		UUID:        uuid.New().String(),
		StoragePath: "", // 将在存储服务中设置
		FileHash:    fileHash,
		Size:        req.Size,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
	}

	// 上传文件到存储
	uploadReq := &UploadFileRequest{
		FileName:    req.FileName,
		ContentType: req.MimeType,
		Data:        req.File,
		Size:        req.Size,
		TenantID:    document.TenantID,
		Metadata: map[string]interface{}{
			"document_id":   document.ID,
			"document_uuid": documentUUID,
			"version":       nextVersion,
		},
	}

	fileResp, err := s.storageService.UploadFile(ctx, uploadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	version.StoragePath = fileResp.ID

	// 使用事务创建版本
	err = s.docRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.versionRepo.Create(ctx, version); err != nil {
			return fmt.Errorf("failed to create version: %w", err)
		}

		// 更新文档的当前版本
		document.CurrentVersion = nextVersion
		if err := s.docRepo.Update(ctx, document); err != nil {
			return fmt.Errorf("failed to update document current version: %w", err)
		}

		// 记录审计日志
		auditReq := &LogActionRequest{
			DocumentID: documentUUID,
			Action:     "create_version",
			Details:    fmt.Sprintf("Created version %d for document: %s", nextVersion, document.Name),
			TenantID:   document.TenantID,
		}

		if err := s.logAudit(ctx, auditReq); err != nil {
			s.logger.WithError(err).Warn("Failed to log audit action")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &DocumentVersionResponse{
		ID:          version.ID,
		DocumentID:  version.DocumentID,
		Version:     version.Version,
		UUID:        version.UUID,
		StoragePath: version.StoragePath,
		FileHash:    version.FileHash,
		Size:        version.Size,
		Description: version.Description,
		CreatedBy:   version.CreatedBy,
		CreatedAt:   version.CreatedAt,
	}, nil
}

// GetVersions 获取文档的所有版本
func (s *documentService) GetVersions(ctx context.Context, documentUUID string) ([]*DocumentVersionResponse, error) {
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, err
	}

	versions, err := s.versionRepo.GetByDocumentID(ctx, document.ID)
	if err != nil {
		return nil, err
	}

	responses := make([]*DocumentVersionResponse, len(versions))
	for i, version := range versions {
		responses[i] = &DocumentVersionResponse{
			ID:          version.ID,
			DocumentID:  version.DocumentID,
			Version:     version.Version,
			UUID:        version.UUID,
			StoragePath: version.StoragePath,
			FileHash:    version.FileHash,
			Size:        version.Size,
			Description: version.Description,
			CreatedBy:   version.CreatedBy,
			CreatedAt:   version.CreatedAt,
		}
	}

	return responses, nil
}

// GetLatestVersion 获取文档的最新版本
func (s *documentService) GetLatestVersion(ctx context.Context, documentUUID string) (*DocumentVersionResponse, error) {
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, err
	}

	version, err := s.versionRepo.GetLatest(ctx, document.ID)
	if err != nil {
		return nil, err
	}

	return &DocumentVersionResponse{
		ID:          version.ID,
		DocumentID:  version.DocumentID,
		Version:     version.Version,
		UUID:        version.UUID,
		StoragePath: version.StoragePath,
		FileHash:    version.FileHash,
		Size:        version.Size,
		Description: version.Description,
		CreatedBy:   version.CreatedBy,
		CreatedAt:   version.CreatedAt,
	}, nil
}

// RestoreVersion 恢复到指定版本
func (s *documentService) RestoreVersion(ctx context.Context, documentUUID string, version int) (*DocumentResponse, error) {
	// 获取文档和版本信息
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, err
	}

	targetVersion, err := s.versionRepo.GetVersionByNumber(ctx, document.ID, version)
	if err != nil {
		return nil, err
	}

	// 创建新版本（复制目标版本的内容）
	newVersionReq := &CreateVersionRequest{
		Description: fmt.Sprintf("Restored from version %d", version),
		FileHash:    targetVersion.FileHash,
		Size:        targetVersion.Size,
		CreatedBy:   1, // TODO: 从上下文获取当前用户ID
	}

	// 从存储服务获取文件内容
	fileData, _, err := s.storageService.DownloadFile(ctx, targetVersion.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download version file: %w", err)
	}

	newVersionReq.File = fileData
	newVersionReq.FileName = fmt.Sprintf("v%d_restored_%s", version, targetVersion.StoragePath)
	newVersionReq.MimeType = "application/octet-stream" // TODO: 根据实际类型设置

	// 创建新版本
	newVersion, err := s.CreateVersion(ctx, documentUUID, newVersionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create restore version: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: documentUUID,
		Action:     "restore_version",
		Details:    fmt.Sprintf("Restored document to version %d: %s", version, document.Name),
		TenantID:   document.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	// 返回更新后的文档
	return s.GetDocument(ctx, documentUUID)
}

// UploadDocument 上传文档
func (s *documentService) UploadDocument(ctx context.Context, req *UploadDocumentRequest) (*DocumentResponse, error) {
	// 验证文件哈希
	fileHash := req.FileHash
	if fileHash == "" {
		hash := sha256.Sum256(req.File)
		fileHash = hex.EncodeToString(hash[:])
	}

	// 创建文档
	createReq := &CreateDocumentRequest{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		TenantID:    req.TenantID,
		CreatedBy:   req.CreatedBy,
		Metadata:    req.Metadata,
	}

	document, err := s.CreateDocument(ctx, createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// 创建初始版本
	versionReq := &CreateVersionRequest{
		Description: "Initial upload",
		File:        req.File,
		FileName:    req.FileName,
		MimeType:    req.MimeType,
		Size:        req.Size,
		FileHash:    fileHash,
		CreatedBy:   req.CreatedBy,
	}

	_, err = s.CreateVersion(ctx, document.UUID, versionReq)
	if err != nil {
		// 如果版本创建失败，删除文档
		_ = s.DeleteDocument(ctx, document.UUID)
		return nil, fmt.Errorf("failed to create initial version: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: document.UUID,
		Action:     "upload",
		Details:    fmt.Sprintf("Uploaded document: %s (%s)", req.Name, req.FileName),
		TenantID:   req.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return document, nil
}

// DownloadDocument 下载文档
func (s *documentService) DownloadDocument(ctx context.Context, documentUUID string, version int) ([]byte, *DownloadMetadata, error) {
	// 获取文档
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, nil, err
	}

	// 检查权限（这里简化处理，实际应该检查用户权限）
	// TODO: 实现权限检查

	var targetVersion *models.DocumentVersion
	if version <= 0 || version == document.CurrentVersion {
		// 获取最新版本
		targetVersion, err = s.versionRepo.GetLatest(ctx, document.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get latest version: %w", err)
		}
	} else {
		// 获取指定版本
		targetVersion, err = s.versionRepo.GetVersionByNumber(ctx, document.ID, version)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get version %d: %w", version, err)
		}
	}

	// 从存储服务下载文件
	fileData, fileMetadata, err := s.storageService.DownloadFile(ctx, targetVersion.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: documentUUID,
		Action:     "download",
		Details:    fmt.Sprintf("Downloaded document: %s (version %d)", document.Name, targetVersion.Version),
		TenantID:   document.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	downloadMetadata := &DownloadMetadata{
		FileName:    fileMetadata.FileName,
		ContentType: fileMetadata.ContentType,
		Size:        fileMetadata.Size,
		FileHash:    targetVersion.FileHash,
		LastModified: targetVersion.CreatedAt,
	}

	return fileData, downloadMetadata, nil
}

// CopyDocument 复制文档
func (s *documentService) CopyDocument(ctx context.Context, documentUUID string, req *CopyDocumentRequest) (*DocumentResponse, error) {
	// 获取源文档
	sourceDoc, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	// TODO: 实现权限检查

	// 创建新文档
	createReq := &CreateDocumentRequest{
		Name:        req.Name,
		Description: sourceDoc.Description,
		Category:    req.Category,
		Tags:        req.Tags,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		TenantID:    sourceDoc.TenantID,
		CreatedBy:   1, // TODO: 从上下文获取当前用户ID
		Metadata:    req.Metadata,
	}

	newDoc, err := s.CreateDocument(ctx, createReq)
	if err != nil {
		return nil, err
	}

	// 复制最新版本
	latestVersion, err := s.versionRepo.GetLatest(ctx, sourceDoc.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	// 从存储服务复制文件
	fileData, _, err := s.storageService.DownloadFile(ctx, latestVersion.StoragePath)
	if err != nil {
		// 删除新创建的文档
		_ = s.DeleteDocument(ctx, newDoc.UUID)
		return nil, fmt.Errorf("failed to download source file: %w", err)
	}

	// 创建新文档的版本
	versionReq := &CreateVersionRequest{
		Description: fmt.Sprintf("Copied from document %s", documentUUID),
		File:        fileData,
		FileName:    latestVersion.StoragePath,
		MimeType:    "application/octet-stream", // TODO: 根据实际类型设置
		Size:        latestVersion.Size,
		FileHash:    latestVersion.FileHash,
		CreatedBy:   1, // TODO: 从上下文获取当前用户ID
	}

	_, err = s.CreateVersion(ctx, newDoc.UUID, versionReq)
	if err != nil {
		// 删除新创建的文档
		_ = s.DeleteDocument(ctx, newDoc.UUID)
		return nil, fmt.Errorf("failed to create copied version: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: newDoc.UUID,
		Action:     "copy",
		Details:    fmt.Sprintf("Copied document from %s to %s", documentUUID, newDoc.UUID),
		TenantID:   newDoc.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return newDoc, nil
}

// MoveDocument 移动文档
func (s *documentService) MoveDocument(ctx context.Context, documentUUID string, req *MoveDocumentRequest) error {
	// 获取文档
	document, err := s.docRepo.GetByUUID(ctx, documentUUID)
	if err != nil {
		return err
	}

	// 检查权限
	// TODO: 实现权限检查

	// 更新文档实体关联
	document.EntityType = req.EntityType
	document.EntityID = req.EntityID

	if err := s.docRepo.Update(ctx, document); err != nil {
		return fmt.Errorf("failed to move document: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		DocumentID: documentUUID,
		Action:     "move",
		Details:    fmt.Sprintf("Moved document to %s:%d", req.EntityType, req.EntityID),
		TenantID:   document.TenantID,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// BatchCreateDocuments 批量创建文档
func (s *documentService) BatchCreateDocuments(ctx context.Context, reqs []*CreateDocumentRequest) ([]*DocumentResponse, error) {
	if len(reqs) == 0 {
		return []*DocumentResponse{}, nil
	}

	responses := make([]*DocumentResponse, len(reqs))

	// 使用事务批量创建
	err := s.docRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		for i, req := range reqs {
			doc, err := s.CreateDocument(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create document %d: %w", i, err)
			}
			responses[i] = doc
		}
		return nil
	})

	return responses, err
}

// BatchUpdateDocuments 批量更新文档
func (s *documentService) BatchUpdateDocuments(ctx context.Context, reqs []*UpdateDocumentRequest) ([]*DocumentResponse, error) {
	if len(reqs) == 0 {
		return []*DocumentResponse{}, nil
	}

	responses := make([]*DocumentResponse, len(reqs))

	// 这里简化处理，实际应该根据文档UUID来更新
	// TODO: 实现基于UUID的批量更新

	return responses, nil
}

// BatchDeleteDocuments 批量删除文档
func (s *documentService) BatchDeleteDocuments(ctx context.Context, documentUUIDs []string) error {
	if len(documentUUIDs) == 0 {
		return nil
	}

	// 使用事务批量删除
	err := s.docRepo.WithTransaction(ctx, func(tx *gorm.DB) error {
		for _, uuid := range documentUUIDs {
			if err := s.DeleteDocument(ctx, uuid); err != nil {
				return fmt.Errorf("failed to delete document %s: %w", uuid, err)
			}
		}
		return nil
	})

	return err
}

// GetDocumentStats 获取文档统计
func (s *documentService) GetDocumentStats(ctx context.Context, tenantID string) (*DocumentStatsResponse, error) {
	stats, err := s.docRepo.GetDocumentStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &DocumentStatsResponse{
		TotalDocuments: stats["total_documents"].(int64),
		TotalSize:      stats["total_size"].(int64),
		Categories:     stats["categories"].(map[string]int64),
		Statuses:       stats["statuses"].(map[string]int64),
		RecentDocuments: stats["recent_documents"].(int64),
		ByCreator:      stats["by_creator"].(map[string]int64),
		ByEntityType:  stats["by_entity_type"].(map[string]int64),
	}, nil
}

// GetTenantStats 获取租户统计
func (s *documentService) GetTenantStats(ctx context.Context, tenantID string) (*TenantStatsResponse, error) {
	// 获取文档统计
	docStats, err := s.GetDocumentStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// TODO: 获取其他统计数据
	// - 用户统计
	// - 存储统计
	// - 活动统计

	return &TenantStatsResponse{
		DocumentStats: docStats,
		// TODO: 添加其他统计
	}, nil
}

// 辅助方法

// convertToDocumentResponse 转换文档模型为响应
func (s *documentService) convertToDocumentResponse(document *models.Document, includeVersions bool) (*DocumentResponse, error) {
	response := &DocumentResponse{
		ID:             document.ID,
		UUID:           document.UUID,
		Name:           document.Name,
		Description:    document.Description,
		OriginalName:   document.OriginalName,
		MIMEType:       document.MIMEType,
		Size:           document.Size,
		Category:       document.Category,
		CurrentVersion: document.CurrentVersion,
		CreatedBy:      document.CreatedBy,
		CreatedAt:      document.CreatedAt,
		UpdatedAt:      document.UpdatedAt,
	}

	// 处理标签
	if document.Tags != "" {
		response.Tags = strings.Split(document.Tags, ",")
	}

	// 包含版本信息
	if includeVersions {
		versions, err := s.versionRepo.GetByDocumentID(context.Background(), document.ID)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to get document versions")
		} else {
			response.Versions = make([]*DocumentVersionResponse, len(versions))
			for i, version := range versions {
				response.Versions[i] = &DocumentVersionResponse{
					ID:          version.ID,
					DocumentID:  version.DocumentID,
					Version:     version.Version,
					UUID:        version.UUID,
					StoragePath: version.StoragePath,
					FileHash:    version.FileHash,
					Size:        version.Size,
					Description: version.Description,
					CreatedBy:   version.CreatedBy,
					CreatedAt:   version.CreatedAt,
				}
			}
		}
	}

	return response, nil
}

// convertToDocumentFilter 转换过滤器
func (s *documentService) convertToDocumentFilter(filter *DocumentFilter) *repositories.DocumentFilter {
	if filter == nil {
		return &repositories.DocumentFilter{}
	}

	repoFilter := &repositories.DocumentFilter{
		TenantID:   filter.TenantID,
		Category:   filter.Category,
		Status:     filter.Status,
		CreatedBy:  filter.CreatedBy,
		Tags:       filter.Tags,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		EntityType: filter.EntityType,
		EntityID:   filter.EntityID,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
	}

	return repoFilter
}

// logAudit 记录审计日志
func (s *documentService) logAudit(ctx context.Context, req *LogActionRequest) error {
	// 解析用户ID
	var userID uint
	if req.UserID != "" {
		// 尝试解析为整数
		_, err := fmt.Sscanf(req.UserID, "%d", &userID)
		if err != nil {
			// 如果解析失败，尝试通过用户名查找
			user, err := s.userRepo.GetByUsername(ctx, req.UserID)
			if err != nil {
				return err
			}
			userID = user.ID
		}
	}

	// 解析文档ID
	var documentID uint
	if req.DocumentID != "" {
		doc, err := s.docRepo.GetByUUID(ctx, req.DocumentID)
		if err != nil {
			return err
		}
		documentID = doc.ID
	}

	audit := &models.DocumentAudit{
		DocumentID: documentID,
		UserID:     userID,
		TenantID:   req.TenantID,
		Action:     req.Action,
		Details:    req.Details,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	return s.auditRepo.Create(ctx, audit)
}