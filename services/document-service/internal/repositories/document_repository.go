package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"gorm.io/gorm"
)

// documentRepository 文档仓库实现
type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建新的文档仓库
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

// Create 创建文档
func (r *documentRepository) Create(ctx context.Context, document *models.Document) error {
	if err := r.db.WithContext(ctx).Create(document).Error; err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	return nil
}

// GetByID 根据ID获取文档
func (r *documentRepository) GetByID(ctx context.Context, id uint) (*models.Document, error) {
	var document models.Document
	if err := r.db.WithContext(ctx).
		Preload("Versions").
		First(&document, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get document by ID: %w", err)
	}
	return &document, nil
}

// GetByUUID 根据UUID获取文档
func (r *documentRepository) GetByUUID(ctx context.Context, uuid string) (*models.Document, error) {
	var document models.Document
	if err := r.db.WithContext(ctx).
		Preload("Versions").
		Where("uuid = ?", uuid).
		First(&document).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not found: %s", uuid)
		}
		return nil, fmt.Errorf("failed to get document by UUID: %w", err)
	}
	return &document, nil
}

// Update 更新文档
func (r *documentRepository) Update(ctx context.Context, document *models.Document) error {
	if err := r.db.WithContext(ctx).Save(document).Error; err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	return nil
}

// Delete 删除文档（硬删除）
func (r *documentRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Unscoped().Delete(&models.Document{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// SoftDelete 软删除文档
func (r *documentRepository) SoftDelete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Document{}, id).Error; err != nil {
		return fmt.Errorf("failed to soft delete document: %w", err)
	}
	return nil
}

// List 根据过滤器列出文档
func (r *documentRepository) List(ctx context.Context, filter *DocumentFilter) ([]*models.Document, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Document{})

	// 应用过滤条件
	if filter != nil {
		if filter.TenantID != "" {
			query = query.Where("tenant_id = ?", filter.TenantID)
		}
		if filter.Category != "" {
			query = query.Where("category = ?", filter.Category)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.CreatedBy > 0 {
			query = query.Where("created_by = ?", filter.CreatedBy)
		}
		if filter.EntityType != "" {
			query = query.Where("entity_type = ?", filter.EntityType)
		}
		if filter.EntityID > 0 {
			query = query.Where("entity_id = ?", filter.EntityID)
		}
		if !filter.StartDate.IsZero() {
			query = query.Where("created_at >= ?", filter.StartDate)
		}
		if !filter.EndDate.IsZero() {
			query = query.Where("created_at <= ?", filter.EndDate)
		}
		if len(filter.Tags) > 0 {
			for _, tag := range filter.Tags {
				query = query.Where("FIND_IN_SET(?, tags)", tag)
			}
		}
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// 应用排序
	orderBy := "created_at DESC"
	if filter != nil && filter.SortBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.SortBy, filter.SortOrder)
		if filter.SortOrder == "" {
			orderBy = filter.SortBy
		}
	}
	query = query.Order(orderBy)

	// 应用分页
	if filter != nil && filter.Page > 0 && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	var documents []*models.Document
	if err := query.Find(&documents).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}

	return documents, total, nil
}

// FindByEntity 根据实体类型和ID查找文档
func (r *documentRepository) FindByEntity(ctx context.Context, entityType string, entityID uint) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to find documents by entity: %w", err)
	}
	return documents, nil
}

// FindByCategory 根据分类查找文档
func (r *documentRepository) FindByCategory(ctx context.Context, tenantID, category string) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND category = ?", tenantID, category).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to find documents by category: %w", err)
	}
	return documents, nil
}

// FindByCreator 根据创建者查找文档
func (r *documentRepository) FindByCreator(ctx context.Context, tenantID string, creatorID uint) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_by = ?", tenantID, creatorID).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to find documents by creator: %w", err)
	}
	return documents, nil
}

// CreateVersion 创建文档版本
func (r *documentRepository) CreateVersion(ctx context.Context, version *models.DocumentVersion) error {
	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return fmt.Errorf("failed to create document version: %w", err)
	}

	// 更新文档的当前版本号
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("id = ?", version.DocumentID).
		Update("current_version", version.Version).Error; err != nil {
		return fmt.Errorf("failed to update document current version: %w", err)
	}

	return nil
}

// GetVersions 获取文档的所有版本
func (r *documentRepository) GetVersions(ctx context.Context, documentID uint) ([]*models.DocumentVersion, error) {
	var versions []*models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("version DESC").
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get document versions: %w", err)
	}
	return versions, nil
}

// GetLatestVersion 获取文档的最新版本
func (r *documentRepository) GetLatestVersion(ctx context.Context, documentID uint) (*models.DocumentVersion, error) {
	var version models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("version DESC").
		First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no versions found for document: %d", documentID)
		}
		return nil, fmt.Errorf("failed to get latest document version: %w", err)
	}
	return &version, nil
}

// GetVersionByNumber 根据版本号获取文档版本
func (r *documentRepository) GetVersionByNumber(ctx context.Context, documentID uint, version int) (*models.DocumentVersion, error) {
	var docVersion models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Where("document_id = ? AND version = ?", documentID, version).
		First(&docVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document version not found: document_id=%d, version=%d", documentID, version)
		}
		return nil, fmt.Errorf("failed to get document version by number: %w", err)
	}
	return &docVersion, nil
}

// SearchByName 根据名称搜索文档
func (r *documentRepository) SearchByName(ctx context.Context, tenantID, name string) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND name LIKE ?", tenantID, "%"+name+"%").
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to search documents by name: %w", err)
	}
	return documents, nil
}

// SearchByContent 根据内容搜索文档（使用数据库的LIKE查询，实际项目中应使用搜索引擎）
func (r *documentRepository) SearchByContent(ctx context.Context, tenantID, content string) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND (name LIKE ? OR description LIKE ? OR tags LIKE ?)",
			tenantID, "%"+content+"%", "%"+content+"%", "%"+content+"%").
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to search documents by content: %w", err)
	}
	return documents, nil
}

// FindByTags 根据标签查找文档
func (r *documentRepository) FindByTags(ctx context.Context, tenantID string, tags []string) ([]*models.Document, error) {
	if len(tags) == 0 {
		return []*models.Document{}, nil
	}

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	for _, tag := range tags {
		query = query.Where("FIND_IN_SET(?, tags)", tag)
	}

	var documents []*models.Document
	if err := query.Order("created_at DESC").Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to find documents by tags: %w", err)
	}
	return documents, nil
}

// FindByDateRange 根据日期范围查找文档
func (r *documentRepository) FindByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND created_at BETWEEN ? AND ?", tenantID, startDate, endDate).
		Order("created_at DESC").
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to find documents by date range: %w", err)
	}
	return documents, nil
}

// CountByTenant 统计租户文档数量
func (r *documentRepository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count documents by tenant: %w", err)
	}
	return count, nil
}

// CountByCategory 统计分类文档数量
func (r *documentRepository) CountByCategory(ctx context.Context, tenantID, category string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("tenant_id = ? AND category = ?", tenantID, category).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count documents by category: %w", err)
	}
	return count, nil
}

// GetSizeByTenant 获取租户文档总大小
func (r *documentRepository) GetSizeByTenant(ctx context.Context, tenantID string) (int64, error) {
	var totalSize int64
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error; err != nil {
		return 0, fmt.Errorf("failed to get document size by tenant: %w", err)
	}
	return totalSize, nil
}

// GetRecentDocuments 获取最近的文档
func (r *documentRepository) GetRecentDocuments(ctx context.Context, tenantID string, limit int) ([]*models.Document, error) {
	if limit <= 0 {
		limit = 10
	}

	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent documents: %w", err)
	}
	return documents, nil
}

// BatchCreate 批量创建文档
func (r *documentRepository) BatchCreate(ctx context.Context, documents []*models.Document) error {
	if len(documents) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(documents, 100).Error; err != nil {
		return fmt.Errorf("failed to batch create documents: %w", err)
	}
	return nil
}

// BatchUpdate 批量更新文档
func (r *documentRepository) BatchUpdate(ctx context.Context, documents []*models.Document) error {
	if len(documents) == 0 {
		return nil
	}

	// 使用事务进行批量更新
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, doc := range documents {
			if err := tx.Save(doc).Error; err != nil {
				return fmt.Errorf("failed to update document %d: %w", doc.ID, err)
			}
		}
		return nil
	})
}

// BatchDelete 批量删除文档
func (r *documentRepository) BatchDelete(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Delete(&models.Document{}, ids).Error; err != nil {
		return fmt.Errorf("failed to batch delete documents: %w", err)
	}
	return nil
}

// GetDocumentStats 获取文档统计信息
func (r *documentRepository) GetDocumentStats(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总文档数
	totalCount, err := r.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_documents"] = totalCount

	// 总大小
	totalSize, err := r.GetSizeByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}
	stats["total_size"] = totalSize

	// 按分类统计
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("category").
		Scan(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category stats: %w", err)
	}
	stats["categories"] = categoryStats

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Select("status, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}
	stats["statuses"] = statusStats

	// 最近30天的文档数量
	var recentCount int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := r.db.WithContext(ctx).
		Model(&models.Document{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, thirtyDaysAgo).
		Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent count: %w", err)
	}
	stats["recent_documents"] = recentCount

	return stats, nil
}