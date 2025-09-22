package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"law-oa-go/internal/errors"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// DocumentService handles document management operations
type DocumentService struct {
	docRepo    repositories.DocumentRepository
	storageDir string
}

// Document represents a document entity
type Document struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Filename    string    `json:"filename"`
	Filepath    string    `json:"filepath"`
	Filesize    int64     `json:"filesize"`
	MimeType    string    `json:"mime_type"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	EntityID    uint      `json:"entity_id"`
	EntityType  string    `json:"entity_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocumentStats represents document statistics
type DocumentStats struct {
	TotalDocuments int64            `json:"total_documents"`
	ByCategory     map[string]int64 `json:"by_category"`
	ByEntityType   map[string]int64 `json:"by_entity_type"`
	RecentUploads  int64            `json:"recent_uploads"`
}

// DocumentListRequest represents document listing request parameters
type DocumentListRequest struct {
	Page       int    `form:"page" json:"page"`
	PageSize   int    `form:"page_size" json:"page_size"`
	Category   string `form:"category" json:"category"`
	EntityType string `form:"entity_type" json:"entity_type"`
	EntityID   uint   `form:"entity_id" json:"entity_id"`
	Search     string `form:"search" json:"search"`
	SortBy     string `form:"sort_by" json:"sort_by"`
	SortOrder  string `form:"sort_order" json:"sort_order"`
}

// DocumentUploadRequest represents document upload request
type DocumentUploadRequest struct {
	Name        string                `form:"name" json:"name"`
	Description string                `form:"description" json:"description"`
	Category    string                `form:"category" json:"category"`
	Tags        string                `form:"tags" json:"tags"`
	EntityID    uint                  `form:"entity_id" json:"entity_id"`
	EntityType  string                `form:"entity_type" json:"entity_type"`
	File        *multipart.FileHeader `form:"file" json:"file"`
}

// DocumentUpdateRequest represents document update request
type DocumentUpdateRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Category    *string  `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// NewDocumentService creates a new document service
func NewDocumentService(docRepo repositories.DocumentRepository, storageDir string) *DocumentService {
	return &DocumentService{
		docRepo:    docRepo,
		storageDir: storageDir,
	}
}

// UploadDocument uploads a new document
func (s *DocumentService) UploadDocument(ctx context.Context, req *DocumentUploadRequest) (*Document, error) {
	// Validate request
	if req.File == nil {
		return nil, errors.NewValidationError("file", "missing_file", "File is required", "Please provide a file to upload")
	}

	// Validate file size (max 50MB)
	if req.File.Size > 50*1024*1024 {
		return nil, errors.NewValidationError("file", "file_too_large", "File too large", "File size must be less than 50MB")
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"text/plain":         true,
		"text/csv":           true,
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	mimeType := req.File.Header.Get("Content-Type")
	if !allowedTypes[mimeType] {
		return nil, errors.NewValidationError("file", "unsupported_file_type", "Unsupported file type", "Only PDF, Word, Excel, text, and image files are allowed")
	}

	// Create document model
	docModel := &models.Document{
		Name:        req.Name,
		Description: req.Description,
		Filename:    req.File.Filename,
		Filesize:    req.File.Size,
		MimeType:    mimeType,
		Category:    req.Category,
		EntityID:    req.EntityID,
		EntityType:  req.EntityType,
		Status:      "active",
	}

	// Set default values
	if docModel.Name == "" {
		docModel.Name = strings.TrimSuffix(req.File.Filename, filepath.Ext(req.File.Filename))
	}

	// Parse tags
	if req.Tags != "" {
		tagList := strings.Split(req.Tags, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
		docModel.Tags = strings.Join(tagList, ",")
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), req.File.Filename)
	docModel.Filepath = filepath.Join(s.storageDir, filename)

	// Save document to database
	if err := s.docRepo.Create(ctx, docModel); err != nil {
		return nil, errors.NewDatabaseError("create_document", "Failed to create document record", err)
	}

	// Convert to response
	return s.toDocument(docModel), nil
}

// GetDocumentByID retrieves a document by ID
func (s *DocumentService) GetDocumentByID(ctx context.Context, id uint) (*Document, error) {
	docModel, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		if err == repositories.ErrDocumentNotFound {
			return nil, errors.NewNotFoundError("document", "Document not found", id)
		}
		return nil, errors.NewDatabaseError("get_document", "Failed to get document", err)
	}

	return s.toDocument(docModel), nil
}

// ListDocuments lists documents with pagination and filtering
func (s *DocumentService) ListDocuments(ctx context.Context, req *DocumentListRequest) ([]*Document, int64, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	params := &repositories.DocumentListParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Category:   req.Category,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Search:     req.Search,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}

	docModels, total, err := s.docRepo.List(ctx, params)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("list_documents", "Failed to list documents", err)
	}

	documents := make([]*Document, len(docModels))
	for i, docModel := range docModels {
		documents[i] = s.toDocument(docModel)
	}

	return documents, total, nil
}

// UpdateDocument updates a document
func (s *DocumentService) UpdateDocument(ctx context.Context, id uint, req *DocumentUpdateRequest) (*Document, error) {
	docModel, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		if err == repositories.ErrDocumentNotFound {
			return nil, errors.NewNotFoundError("document", "Document not found", id)
		}
		return nil, errors.NewDatabaseError("get_document", "Failed to get document", err)
	}

	// Update fields
	if req.Name != nil {
		docModel.Name = *req.Name
	}
	if req.Description != nil {
		docModel.Description = *req.Description
	}
	if req.Category != nil {
		docModel.Category = *req.Category
	}
	if req.Tags != nil {
		docModel.Tags = strings.Join(req.Tags, ",")
	}

	docModel.UpdatedAt = time.Now()

	if err := s.docRepo.Update(ctx, docModel); err != nil {
		return nil, errors.NewDatabaseError("update_document", "Failed to update document", err)
	}

	return s.toDocument(docModel), nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(ctx context.Context, id uint) error {
	_, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		if err == repositories.ErrDocumentNotFound {
			return errors.NewNotFoundError("document", "Document not found", id)
		}
		return errors.NewDatabaseError("get_document", "Failed to get document", err)
	}

	// Delete document from database
	if err := s.docRepo.Delete(ctx, id); err != nil {
		return errors.NewDatabaseError("delete_document", "Failed to delete document", err)
	}

	return nil
}

// GetDocumentStats retrieves document statistics
func (s *DocumentService) GetDocumentStats(ctx context.Context) (*DocumentStats, error) {
	stats, err := s.docRepo.GetStats(ctx)
	if err != nil {
		return nil, errors.NewDatabaseError("get_document_stats", "Failed to get document statistics", err)
	}

	// Convert category map
	categoryMap := make(map[string]int64)
	for _, stat := range stats.ByCategory {
		categoryMap[stat.Category] = stat.Count
	}

	// Convert entity type map
	entityMap := make(map[string]int64)
	for _, stat := range stats.ByEntityType {
		entityMap[stat.EntityType] = stat.Count
	}

	return &DocumentStats{
		TotalDocuments: stats.Total,
		ByCategory:     categoryMap,
		ByEntityType:   entityMap,
		RecentUploads:  stats.RecentUploads,
	}, nil
}

// DownloadDocument downloads a document file
func (s *DocumentService) DownloadDocument(ctx context.Context, id uint) (io.ReadCloser, *Document, error) {
	docModel, err := s.docRepo.FindByID(ctx, id)
	if err != nil {
		if err == repositories.ErrDocumentNotFound {
			return nil, nil, errors.NewNotFoundError("document", "Document not found", id)
		}
		return nil, nil, errors.NewDatabaseError("get_document", "Failed to get document", err)
	}

	// Open file for reading
	file, err := s.openFile(docModel.Filepath)
	if err != nil {
		return nil, nil, errors.NewInternalError("file_open", "Failed to open document file", err)
	}

	return file, s.toDocument(docModel), nil
}

// Helper methods

func (s *DocumentService) toDocument(model *models.Document) *Document {
	tags := []string{}
	if model.Tags != "" {
		tags = strings.Split(model.Tags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	return &Document{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		Filename:    model.Filename,
		Filepath:    model.Filepath,
		Filesize:    model.Filesize,
		MimeType:    model.MimeType,
		Category:    model.Category,
		Tags:        tags,
		EntityID:    model.EntityID,
		EntityType:  model.EntityType,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// saveToFile saves uploaded file to filesystem
func (s *DocumentService) saveToFile(src multipart.File, dst string) (string, error) {
	// Implementation would save the file to the destination path
	// For now, returning the destination path as a placeholder
	return dst, nil
}

// deleteFile deletes a file from filesystem
func (s *DocumentService) deleteFile(filepath string) error {
	// Implementation would delete the file from the filesystem
	// For now, returning nil as a placeholder
	return nil
}

// openFile opens a file for reading
func (s *DocumentService) openFile(filepath string) (io.ReadCloser, error) {
	// Implementation would open the file for reading
	// For now, returning nil as a placeholder
	return nil, nil
}
