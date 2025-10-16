package repositories

import (
	"context"
	"law-oa-go/internal/models"
	"gorm.io/gorm"
)

type LawyerRepository interface {
	List(ctx context.Context, params *LawyerListParams) ([]*models.User, int64, error)
	Delete(ctx context.Context, id uint) error
}

type LawyerRepositoryImpl struct {
	db *gorm.DB
}

func NewLawyerRepository(db *gorm.DB) LawyerRepository {
	return &LawyerRepositoryImpl{db: db}
}

type LawyerListParams struct {
	Page     int
	PageSize int
	Search   string
}

func (r *LawyerRepositoryImpl) List(ctx context.Context, params *LawyerListParams) ([]*models.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", "lawyer")

	if params.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var lawyers []models.User
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("created_at DESC").Find(&lawyers).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.User, len(lawyers))
	for i := range lawyers {
		result[i] = &lawyers[i]
	}

	return result, total, nil
}

func (r *LawyerRepositoryImpl) Delete(ctx context.Context, id uint) error {
	// 直接尝试删除律师（软删除），GORM会自动处理记录不存在的情况
	result := r.db.WithContext(ctx).Where("id = ? AND role = ?", id, "lawyer").Delete(&models.User{})

	// 检查是否影响了任何行
	if result.Error != nil {
		return result.Error
	}

	// 如果没有行被影响，说明律师不存在
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
