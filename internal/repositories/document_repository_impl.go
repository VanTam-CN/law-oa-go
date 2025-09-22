package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// documentRepository 文档仓库实现
type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建文档仓库
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{
		db: db,
	}
}

// Create 创建文档
func (r *documentRepository) Create(ctx context.Context, document *models.Document) error {
	return r.db.WithContext(ctx).Create(document).Error
}

// FindByID 根据ID查找文档
func (r *documentRepository) FindByID(ctx context.Context, id uint) (*models.Document, error) {
	var document models.Document
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	return &document, nil
}

// List 列出文档
func (r *documentRepository) List(ctx context.Context, params *DocumentListParams) ([]*models.Document, int64, error) {
	var documents []*models.Document
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Document{})

	// 应用过滤条件
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}
	if params.EntityType != "" {
		query = query.Where("entity_type = ?", params.EntityType)
	}
	if params.EntityID > 0 {
		query = query.Where("entity_id = ?", params.EntityID)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR tags LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序
	order := "created_at DESC"
	if params.SortBy != "" {
		if params.SortOrder == "asc" {
			order = params.SortBy + " ASC"
		} else {
			order = params.SortBy + " DESC"
		}
	}
	query = query.Order(order)

	// 应用分页
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	// 执行查询
	if err := query.Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// Update 更新文档
func (r *documentRepository) Update(ctx context.Context, document *models.Document) error {
	document.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(document).Error
}

// Delete 删除文档
func (r *documentRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Document{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return nil
}

// GetStats 获取文档统计
func (r *documentRepository) GetStats(ctx context.Context) (*DocumentStats, error) {
	var stats DocumentStats

	// 总数
	if err := r.db.WithContext(ctx).Model(&models.Document{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 按分类统计
	rows, err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int64
		if err := r.db.ScanRows(rows, &struct {
			Category string
			Count    int64
		}{category, count}); err != nil {
			continue
		}
		stats.ByCategory = append(stats.ByCategory, struct {
			Category string
			Count    int64
		}{category, count})
	}

	// 按实体类型统计
	entityRows, err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Select("entity_type, COUNT(*) as count").
		Group("entity_type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer entityRows.Close()

	for entityRows.Next() {
		var entityType string
		var count int64
		if err := r.db.ScanRows(entityRows, &struct {
			EntityType string
			Count      int64
		}{entityType, count}); err != nil {
			continue
		}
		stats.ByEntityType = append(stats.ByEntityType, struct {
			EntityType string
			Count      int64
		}{entityType, count})
	}

	// 最近上传统计（过去7天）
	var recentCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("created_at >= ?", time.Now().AddDate(0, 0, -7)).
		Count(&recentCount).Error; err != nil {
		return nil, err
	}
	stats.RecentUploads = recentCount

	return &stats, nil
}

// FindByEntity 根据实体查找文档
func (r *documentRepository) FindByEntity(ctx context.Context, entityType string, entityID uint) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Find(&documents).Error
	return documents, err
}
