package repositories

import (
	"context"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// legalCategoryRepository 法条分类仓储实现
type legalCategoryRepository struct {
	db *gorm.DB
}

// NewLegalCategoryRepository 创建法条分类仓储实例
func NewLegalCategoryRepository(db *gorm.DB) LegalCategoryRepository {
	return &legalCategoryRepository{
		db: db,
	}
}

// Create 创建分类
func (r *legalCategoryRepository) Create(ctx context.Context, category *models.LegalCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID 根据ID获取分类
func (r *legalCategoryRepository) GetByID(ctx context.Context, id int) (*models.LegalCategory, error) {
	var category models.LegalCategory
	err := r.db.WithContext(ctx).
		Preload("Parent").
		First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetByCode 根据代码获取分类
func (r *legalCategoryRepository) GetByCode(ctx context.Context, code string) (*models.LegalCategory, error) {
	var category models.LegalCategory
	err := r.db.WithContext(ctx).
		Preload("Parent").
		Where("code = ?", code).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Update 更新分类
func (r *legalCategoryRepository) Update(ctx context.Context, category *models.LegalCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete 删除分类
func (r *legalCategoryRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.LegalCategory{}, id).Error
}

// List 获取所有分类
func (r *legalCategoryRepository) List(ctx context.Context) ([]*models.LegalCategory, error) {
	var categories []*models.LegalCategory
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("level ASC, name ASC").
		Find(&categories).Error
	return categories, err
}

// GetTree 获取分类树结构
func (r *legalCategoryRepository) GetTree(ctx context.Context) ([]*models.CategoryTreeNode, error) {
	// 获取所有活跃分类
	var categories []*models.LegalCategory
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("level ASC, name ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// 构建树结构
	return r.buildCategoryTree(categories, 0), nil
}

// buildCategoryTree 递归构建分类树
func (r *legalCategoryRepository) buildCategoryTree(categories []*models.LegalCategory, parentID int) []*models.CategoryTreeNode {
	var tree []*models.CategoryTreeNode

	for _, category := range categories {
		if (parentID == 0 && category.ParentID == nil) || (category.ParentID != nil && *category.ParentID == parentID) {
			node := &models.CategoryTreeNode{
				ID:           category.ID,
				Name:         category.Name,
				Code:         category.Code,
				Level:        category.Level,
				StatuteCount: 0, // 需要统计，暂时设为0
				Children:     r.buildCategoryTree(categories, category.ID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// GetChildren 获取子分类
func (r *legalCategoryRepository) GetChildren(ctx context.Context, parentID int) ([]*models.LegalCategory, error) {
	var categories []*models.LegalCategory
	err := r.db.WithContext(ctx).
		Where("parent_id = ? AND is_active = ?", parentID, true).
		Order("name ASC").
		Find(&categories).Error
	return categories, err
}
