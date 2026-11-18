package repositories

import (
	"context"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// legalTagRepository 法条标签仓储实现
type legalTagRepository struct {
	db *gorm.DB
}

// NewLegalTagRepository 创建法条标签仓储实例
func NewLegalTagRepository(db *gorm.DB) LegalTagRepository {
	return &legalTagRepository{
		db: db,
	}
}

// Create 创建标签
func (r *legalTagRepository) Create(ctx context.Context, tag *models.LegalTag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// GetByID 根据ID获取标签
func (r *legalTagRepository) GetByID(ctx context.Context, id int) (*models.LegalTag, error) {
	var tag models.LegalTag
	err := r.db.WithContext(ctx).First(&tag, id).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetByName 根据名称获取标签
func (r *legalTagRepository) GetByName(ctx context.Context, name string) (*models.LegalTag, error) {
	var tag models.LegalTag
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// Update 更新标签
func (r *legalTagRepository) Update(ctx context.Context, tag *models.LegalTag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// Delete 删除标签
func (r *legalTagRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.LegalTag{}, id).Error
}

// List 获取所有标签
func (r *legalTagRepository) List(ctx context.Context) ([]*models.LegalTag, error) {
	var tags []*models.LegalTag
	err := r.db.WithContext(ctx).
		Order("usage_count DESC, name ASC").
		Find(&tags).Error
	return tags, err
}

// GetPopularTags 获取热门标签
func (r *legalTagRepository) GetPopularTags(ctx context.Context, limit int) ([]*models.LegalTag, error) {
	var tags []*models.LegalTag
	err := r.db.WithContext(ctx).
		Order("usage_count DESC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

// UpdateUsageCount 更新标签使用次数
func (r *legalTagRepository) UpdateUsageCount(ctx context.Context, tagID int) error {
	return r.db.WithContext(ctx).
		Model(&models.LegalTag{}).
		Where("id = ?", tagID).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}