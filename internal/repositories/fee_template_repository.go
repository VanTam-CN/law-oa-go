package repositories

import (
	"context"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// FeeTemplateRepository 费率模板仓储接口
type FeeTemplateRepository interface {
	Create(ctx context.Context, template *models.FeeTemplate) error
	GetByID(ctx context.Context, id uint) (*models.FeeTemplate, error)
	Update(ctx context.Context, template *models.FeeTemplate) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, page, pageSize int) ([]*models.FeeTemplate, int64, error)
	GetByCaseType(ctx context.Context, caseType string) ([]*models.FeeTemplate, error)
}

// FeeTemplateRepositoryImpl 费率模板仓储实现
type FeeTemplateRepositoryImpl struct {
	db *gorm.DB
}

// NewFeeTemplateRepository 创建费率模板仓储
func NewFeeTemplateRepository(db *gorm.DB) FeeTemplateRepository {
	return &FeeTemplateRepositoryImpl{db: db}
}

func (r *FeeTemplateRepositoryImpl) Create(ctx context.Context, template *models.FeeTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

func (r *FeeTemplateRepositoryImpl) GetByID(ctx context.Context, id uint) (*models.FeeTemplate, error) {
	var template models.FeeTemplate
	err := r.db.WithContext(ctx).First(&template, id).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *FeeTemplateRepositoryImpl) Update(ctx context.Context, template *models.FeeTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *FeeTemplateRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.FeeTemplate{}, id).Error
}

func (r *FeeTemplateRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*models.FeeTemplate, int64, error) {
	var templates []*models.FeeTemplate
	var total int64
	db := r.db.WithContext(ctx).Model(&models.FeeTemplate{})
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error
	return templates, total, err
}

func (r *FeeTemplateRepositoryImpl) GetByCaseType(ctx context.Context, caseType string) ([]*models.FeeTemplate, error) {
	var templates []*models.FeeTemplate
	err := r.db.WithContext(ctx).Where("case_type = ? AND active = ?", caseType, true).Find(&templates).Error
	return templates, err
}
