package repositories

import (
	"context"
	"fmt"

	"github.com/law-oa-go/document-service/internal/models"
	"gorm.io/gorm"
)

// documentVersionRepository 文档版本仓库实现
type documentVersionRepository struct {
	db *gorm.DB
}

// NewDocumentVersionRepository 创建新的文档版本仓库
func NewDocumentVersionRepository(db *gorm.DB) DocumentVersionRepository {
	return &documentVersionRepository{db: db}
}

// Create 创建文档版本
func (r *documentVersionRepository) Create(ctx context.Context, version *models.DocumentVersion) error {
	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return fmt.Errorf("failed to create document version: %w", err)
	}
	return nil
}

// GetByID 根据ID获取文档版本
func (r *documentVersionRepository) GetByID(ctx context.Context, id uint) (*models.DocumentVersion, error) {
	var version models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Preload("Document").
		First(&version, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document version not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get document version by ID: %w", err)
	}
	return &version, nil
}

// GetByUUID 根据UUID获取文档版本
func (r *documentVersionRepository) GetByUUID(ctx context.Context, uuid string) (*models.DocumentVersion, error) {
	var version models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Preload("Document").
		Where("uuid = ?", uuid).
		First(&version).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document version not found: %s", uuid)
		}
		return nil, fmt.Errorf("failed to get document version by UUID: %w", err)
	}
	return &version, nil
}

// GetByDocumentID 获取文档的所有版本
func (r *documentVersionRepository) GetByDocumentID(ctx context.Context, documentID uint) ([]*models.DocumentVersion, error) {
	var versions []*models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("version DESC").
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get document versions by document ID: %w", err)
	}
	return versions, nil
}

// GetLatest 获取文档的最新版本
func (r *documentVersionRepository) GetLatest(ctx context.Context, documentID uint) (*models.DocumentVersion, error) {
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

// Delete 删除文档版本
func (r *documentVersionRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.DocumentVersion{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete document version: %w", err)
	}
	return nil
}

// DeleteByDocumentID 删除文档的所有版本
func (r *documentVersionRepository) DeleteByDocumentID(ctx context.Context, documentID uint) error {
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Delete(&models.DocumentVersion{}).Error; err != nil {
		return fmt.Errorf("failed to delete document versions by document ID: %w", err)
	}
	return nil
}

// GetVersionByNumber 根据版本号获取文档版本
func (r *documentVersionRepository) GetVersionByNumber(ctx context.Context, documentID uint, version int) (*models.DocumentVersion, error) {
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

// GetVersionRange 获取指定范围的版本
func (r *documentVersionRepository) GetVersionRange(ctx context.Context, documentID uint, startVersion, endVersion int) ([]*models.DocumentVersion, error) {
	var versions []*models.DocumentVersion
	query := r.db.WithContext(ctx).
		Where("document_id = ?", documentID)

	if startVersion > 0 {
		query = query.Where("version >= ?", startVersion)
	}
	if endVersion > 0 {
		query = query.Where("version <= ?", endVersion)
	}

	if err := query.Order("version DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get document version range: %w", err)
	}
	return versions, nil
}

// GetNextVersionNumber 获取下一个版本号
func (r *documentVersionRepository) GetNextVersionNumber(ctx context.Context, documentID uint) (int, error) {
	var latestVersion int
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Where("document_id = ?", documentID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&latestVersion).Error; err != nil {
		return 0, fmt.Errorf("failed to get next version number: %w", err)
	}
	return latestVersion + 1, nil
}

// CountVersions 统计文档版本数量
func (r *documentVersionRepository) CountVersions(ctx context.Context, documentID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Where("document_id = ?", documentID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count document versions: %w", err)
	}
	return count, nil
}

// GetVersionsBySize 根据文件大小获取版本
func (r *documentVersionRepository) GetVersionsBySize(ctx context.Context, documentID uint, minSize, maxSize int64) ([]*models.DocumentVersion, error) {
	var versions []*models.DocumentVersion
	query := r.db.WithContext(ctx).
		Where("document_id = ?", documentID)

	if minSize > 0 {
		query = query.Where("size >= ?", minSize)
	}
	if maxSize > 0 {
		query = query.Where("size <= ?", maxSize)
	}

	if err := query.Order("version DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get document versions by size: %w", err)
	}
	return versions, nil
}

// GetVersionsByHash 根据文件哈希获取版本
func (r *documentVersionRepository) GetVersionsByHash(ctx context.Context, fileHash string) ([]*models.DocumentVersion, error) {
	var versions []*models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Where("file_hash = ?", fileHash).
		Preload("Document").
		Order("created_at DESC").
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to get document versions by hash: %w", err)
	}
	return versions, nil
}

// GetVersionStats 获取版本统计信息
func (r *documentVersionRepository) GetVersionStats(ctx context.Context, documentID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 版本总数
	totalCount, err := r.CountVersions(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_versions"] = totalCount

	// 总大小
	var totalSize int64
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Where("document_id = ?", documentID).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error; err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}
	stats["total_size"] = totalSize

	// 平均大小
	if totalCount > 0 {
		stats["average_size"] = float64(totalSize) / float64(totalCount)
	} else {
		stats["average_size"] = 0
	}

	// 最新版本信息
	latest, err := r.GetLatest(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	stats["latest_version"] = latest.Version
	stats["latest_size"] = latest.Size
	stats["latest_created_at"] = latest.CreatedAt

	return stats, nil
}

// CleanupOldVersions 清理旧版本（保留指定数量的最新版本）
func (r *documentVersionRepository) CleanupOldVersions(ctx context.Context, documentID uint, keepCount int) error {
	if keepCount <= 0 {
		keepCount = 1 // 至少保留一个版本
	}

	// 获取需要删除的版本ID
	var versionIDs []uint
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Where("document_id = ?", documentID).
		Order("version DESC").
		Offset(keepCount).
		Pluck("id", &versionIDs).Error; err != nil {
		return fmt.Errorf("failed to get old version IDs: %w", err)
	}

	if len(versionIDs) == 0 {
		return nil // 没有需要删除的版本
	}

	// 删除旧版本
	if err := r.db.WithContext(ctx).
		Delete(&models.DocumentVersion{}, versionIDs).Error; err != nil {
		return fmt.Errorf("failed to delete old versions: %w", err)
	}

	return nil
}

// BatchCreate 批量创建版本
func (r *documentVersionRepository) BatchCreate(ctx context.Context, versions []*models.DocumentVersion) error {
	if len(versions) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(versions, 100).Error; err != nil {
		return fmt.Errorf("failed to batch create document versions: %w", err)
	}
	return nil
}

// GetVersionsWithPagination 分页获取版本
func (r *documentVersionRepository) GetVersionsWithPagination(ctx context.Context, documentID uint, page, pageSize int) ([]*models.DocumentVersion, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 获取总数
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentVersion{}).
		Where("document_id = ?", documentID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count document versions: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	var versions []*models.DocumentVersion
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("version DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get document versions with pagination: %w", err)
	}

	return versions, total, nil
}