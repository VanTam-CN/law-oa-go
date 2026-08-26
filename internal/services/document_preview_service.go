package services

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"law-oa-go/internal/errors"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// DocumentPreviewService handles document preview operations
type DocumentPreviewService struct {
	docRepo repositories.DocumentRepository
}

// NewDocumentPreviewService creates a new document preview service
func NewDocumentPreviewService(docRepo repositories.DocumentRepository) *DocumentPreviewService {
	return &DocumentPreviewService{
		docRepo: docRepo,
	}
}

// PreviewRequest represents a document preview request
type PreviewRequest struct {
	DocumentID uint `json:"document_id" binding:"required"`
	Page       int  `json:"page,omitempty"`
	Width      int  `json:"width,omitempty"`
	Height     int  `json:"height,omitempty"`
}

// PreviewResponse represents a document preview response
type PreviewResponse struct {
	DocumentID   uint        `json:"document_id"`
	Name         string      `json:"name"`
	MimeType     string      `json:"mime_type"`
	FileSize     int64       `json:"file_size"`
	PreviewType  string      `json:"preview_type"`
	Content      interface{} `json:"content"`
	PageCount    int         `json:"page_count,omitempty"`
	CurrentPage  int         `json:"current_page,omitempty"`
	ThumbnailURL string      `json:"thumbnail_url,omitempty"`
	PreviewURL   string      `json:"preview_url,omitempty"`
	GeneratedAt  time.Time   `json:"generated_at"`
}

// TextPreview represents text content preview
type TextPreview struct {
	Content string `json:"content"`
	Length  int    `json:"length"`
}

// ImagePreview represents image preview info
type ImagePreview struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Size   int64  `json:"size"`
}

// PDFPreview represents PDF preview info
type PDFPreview struct {
	PageCount    int    `json:"page_count"`
	CurrentPage  int    `json:"current_page"`
	ThumbnailURL string `json:"thumbnail_url"`
	PreviewURL   string `json:"preview_url"`
}

// GetDocumentPreview generates a preview for a document
func (s *DocumentPreviewService) GetDocumentPreview(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error) {
	// Get document from database
	docModel, err := s.docRepo.FindByID(ctx, req.DocumentID)
	if err != nil {
		if err == repositories.ErrDocumentNotFound {
			return nil, errors.NotFoundError("document", "Document not found", req.DocumentID)
		}
		return nil, errors.DatabaseError("get_document", "Failed to get document", err)
	}

	// Generate preview based on document type
	preview, err := s.generatePreview(ctx, docModel, req)
	if err != nil {
		return nil, errors.InternalError("Failed to generate preview", err)
	}

	return preview, nil
}

// generatePreview generates preview based on document mime type
func (s *DocumentPreviewService) generatePreview(ctx context.Context, doc *models.Document, req *PreviewRequest) (*PreviewResponse, error) {
	mimeType := doc.MimeType

	response := &PreviewResponse{
		DocumentID:  doc.ID,
		Name:        doc.Name,
		MimeType:    mimeType,
		FileSize:    doc.Filesize,
		GeneratedAt: time.Now(),
	}

	switch {
	case strings.HasPrefix(mimeType, "text/"):
		return s.generateTextPreview(doc, response)
	case strings.HasPrefix(mimeType, "image/"):
		return s.generateImagePreview(doc, response)
	case mimeType == "application/pdf":
		return s.generatePDFPreview(doc, response, req)
	case strings.Contains(mimeType, "word") || strings.Contains(mimeType, "document"):
		return s.generateDocumentPreview(doc, response)
	case strings.Contains(mimeType, "excel") || strings.Contains(mimeType, "spreadsheet"):
		return s.generateSpreadsheetPreview(doc, response)
	default:
		return s.generateGenericPreview(doc, response)
	}
}

// generateTextPreview generates preview for text files
func (s *DocumentPreviewService) generateTextPreview(doc *models.Document, response *PreviewResponse) (*PreviewResponse, error) {
	// Read first 1000 characters of text file
	file, err := os.Open(doc.Filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open text file: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 1000)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	content := string(buf[:n])

	// Truncate at word boundary
	if len(content) == 1000 {
		if idx := strings.LastIndex(content, " "); idx != -1 {
			content = content[:idx] + "..."
		}
	}

	response.PreviewType = "text"
	response.Content = &TextPreview{
		Content: content,
		Length:  n,
	}

	return response, nil
}

// generateImagePreview generates preview for image files
func (s *DocumentPreviewService) generateImagePreview(doc *models.Document, response *PreviewResponse) (*PreviewResponse, error) {
	response.PreviewType = "image"
	response.Content = &ImagePreview{
		URL:    fmt.Sprintf("/api/v1/documents/%d/content", doc.ID),
		Width:  0, // Would be populated by image processing
		Height: 0, // Would be populated by image processing
		Size:   doc.Filesize,
	}

	return response, nil
}

// generatePDFPreview generates preview for PDF files
func (s *DocumentPreviewService) generatePDFPreview(doc *models.Document, response *PreviewResponse, req *PreviewRequest) (*PreviewResponse, error) {
	response.PreviewType = "pdf"
	response.PageCount = 1 // Would be populated by PDF processing

	if req.Page == 0 {
		req.Page = 1
	}

	response.Content = &PDFPreview{
		PageCount:    1, // Would be determined by PDF processing
		CurrentPage:  req.Page,
		ThumbnailURL: fmt.Sprintf("/api/v1/documents/%d/thumbnail?page=%d", doc.ID, req.Page),
		PreviewURL:   fmt.Sprintf("/api/v1/documents/%d/preview?page=%d", doc.ID, req.Page),
	}

	return response, nil
}

// generateDocumentPreview generates preview for Word documents
func (s *DocumentPreviewService) generateDocumentPreview(doc *models.Document, response *PreviewResponse) (*PreviewResponse, error) {
	response.PreviewType = "document"
	response.Content = map[string]interface{}{
		"type":        "word",
		"description": "Word document preview would require specialized libraries",
		"thumbnail":   fmt.Sprintf("/api/v1/documents/%d/thumbnail", doc.ID),
	}

	return response, nil
}

// generateSpreadsheetPreview generates preview for Excel files
func (s *DocumentPreviewService) generateSpreadsheetPreview(doc *models.Document, response *PreviewResponse) (*PreviewResponse, error) {
	response.PreviewType = "spreadsheet"
	response.Content = map[string]interface{}{
		"type":        "excel",
		"description": "Excel spreadsheet preview would require specialized libraries",
		"thumbnail":   fmt.Sprintf("/api/v1/documents/%d/thumbnail", doc.ID),
	}

	return response, nil
}

// generateGenericPreview generates preview for unsupported file types
func (s *DocumentPreviewService) generateGenericPreview(doc *models.Document, response *PreviewResponse) (*PreviewResponse, error) {
	response.PreviewType = "generic"
	response.Content = map[string]interface{}{
		"type":         "file",
		"description":  "Preview not available for this file type",
		"download_url": fmt.Sprintf("/api/v1/documents/%d/download", doc.ID),
	}

	return response, nil
}

// GetDocumentContent retrieves the actual content of a document for preview/download
func (s *DocumentPreviewService) GetDocumentContent(ctx context.Context, documentID uint) (io.ReadCloser, *models.Document, error) {
	docModel, err := s.docRepo.FindByID(ctx, documentID)
	if err != nil {
		if err == repositories.ErrDocumentNotFound {
			return nil, nil, errors.NotFoundError("document", "Document not found", documentID)
		}
		return nil, nil, errors.DatabaseError("get_document", "Failed to get document", err)
	}

	// Open file for reading
	file, err := os.Open(docModel.Filepath)
	if err != nil {
		return nil, nil, errors.InternalError("Failed to open document file", err)
	}

	return file, docModel, nil
}

// DetectMimeType detects the mime type of a file
func (s *DocumentPreviewService) DetectMimeType(filename string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

// GetPreviewSettings returns preview settings for different document types
func (s *DocumentPreviewService) GetPreviewSettings() map[string]interface{} {
	return map[string]interface{}{
		"max_text_preview_length": 1000,
		"supported_preview_types": []string{
			"text/plain",
			"text/csv",
			"image/jpeg",
			"image/png",
			"image/gif",
			"application/pdf",
		},
		"image_max_width":   800,
		"image_max_height":  600,
		"pdf_max_pages":     50,
		"cache_ttl_minutes": 30,
	}
}
