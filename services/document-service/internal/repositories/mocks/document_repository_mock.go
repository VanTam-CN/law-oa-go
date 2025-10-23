package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
)

// DocumentRepository 文档仓库模拟
type DocumentRepository struct {
	documents map[uint]*models.Document
	nextID    uint
}

// NewDocumentRepository 创建文档仓库模拟
func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{
		documents: make(map[uint]*models.Document),
		nextID:    1,
	}
}

// GetByID 根据ID获取文档
func (r *DocumentRepository) GetByID(ctx context.Context, id uint) (*models.Document, error) {
	if doc, exists := r.documents[id]; exists {
		return doc, nil
	}
	return nil, errors.New("document not found")
}

// GetByUUID 根据UUID获取文档
func (r *DocumentRepository) GetByUUID(ctx context.Context, uuid string) (*models.Document, error) {
	for _, doc := range r.documents {
		if doc.UUID == uuid {
			return doc, nil
		}
	}
	return nil, errors.New("document not found")
}

// Create 创建文档
func (r *DocumentRepository) Create(ctx context.Context, document *models.Document) error {
	document.ID = r.nextID
	document.CreatedAt = time.Now()
	document.UpdatedAt = time.Now()
	r.documents[document.ID] = document
	r.nextID++
	return nil
}

// Update 更新文档
func (r *DocumentRepository) Update(ctx context.Context, document *models.Document) error {
	if _, exists := r.documents[document.ID]; exists {
		document.UpdatedAt = time.Now()
		r.documents[document.ID] = document
		return nil
	}
	return errors.New("document not found")
}

// Delete 删除文档
func (r *DocumentRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := r.documents[id]; exists {
		delete(r.documents, id)
		return nil
	}
	return errors.New("document not found")
}

// List 列出文档
func (r *DocumentRepository) List(ctx context.Context, options DocumentListOptions) ([]*models.Document, int64, error) {
	var result []*models.Document
	var total int64

	for _, doc := range r.documents {
		// 应用过滤条件
		if options.TenantID != "" && doc.TenantID != options.TenantID {
			continue
		}
		if options.Category != "" && doc.Category != options.Category {
			continue
		}
		if options.CreatedBy != 0 && doc.CreatedBy != options.CreatedBy {
			continue
		}
		if options.Search != "" {
			if !contains(doc.Name, options.Search) && !contains(doc.Description, options.Search) {
				continue
			}
		}

		total++
		// 应用分页
		if options.Offset > 0 {
			options.Offset--
			continue
		}
		if options.Limit > 0 && len(result) >= options.Limit {
			continue
		}
		result = append(result, doc)
	}

	return result, total, nil
}

// DocumentListOptions 文档列表选项
type DocumentListOptions struct {
	TenantID  string
	Category  string
	CreatedBy uint
	Search    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}