package repositories

import (
	"context"

	"law-oa-go/internal/models"
	"gorm.io/gorm"
)

// CommissionRuleRepository 分成规则数据仓库接口
type CommissionRuleRepository interface {
	Create(ctx context.Context, rule *models.CommissionRule) error
	Update(ctx context.Context, rule *models.CommissionRule) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*models.CommissionRule, error)
	FindAll(ctx context.Context) ([]*models.CommissionRule, error)
	FindActiveByRole(ctx context.Context, role string) ([]*models.CommissionRule, error)
	FindByRole(ctx context.Context, role string) ([]*models.CommissionRule, error)
}

// CommissionRuleRepositoryImpl 分成规则数据仓库实现
type CommissionRuleRepositoryImpl struct {
	db *gorm.DB
}

// NewCommissionRuleRepository 创建分成规则数据仓库实例
func NewCommissionRuleRepository(db *gorm.DB) CommissionRuleRepository {
	return &CommissionRuleRepositoryImpl{db: db}
}

func (r *CommissionRuleRepositoryImpl) Create(ctx context.Context, rule *models.CommissionRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *CommissionRuleRepositoryImpl) Update(ctx context.Context, rule *models.CommissionRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *CommissionRuleRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.CommissionRule{}, id).Error
}

func (r *CommissionRuleRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.CommissionRule, error) {
	var rule models.CommissionRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *CommissionRuleRepositoryImpl) FindAll(ctx context.Context) ([]*models.CommissionRule, error) {
	var rules []models.CommissionRule
	if err := r.db.WithContext(ctx).
		Order("role, priority ASC, min_amount ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}
	result := make([]*models.CommissionRule, len(rules))
	for i := range rules {
		result[i] = &rules[i]
	}
	return result, nil
}

func (r *CommissionRuleRepositoryImpl) FindActiveByRole(ctx context.Context, role string) ([]*models.CommissionRule, error) {
	var rules []models.CommissionRule
	if err := r.db.WithContext(ctx).
		Where("role = ? AND active = true", role).
		Order("priority DESC, min_amount ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}
	result := make([]*models.CommissionRule, len(rules))
	for i := range rules {
		result[i] = &rules[i]
	}
	return result, nil
}

func (r *CommissionRuleRepositoryImpl) FindByRole(ctx context.Context, role string) ([]*models.CommissionRule, error) {
	var rules []models.CommissionRule
	if err := r.db.WithContext(ctx).
		Where("role = ?", role).
		Order("priority DESC, min_amount ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}
	result := make([]*models.CommissionRule, len(rules))
	for i := range rules {
		result[i] = &rules[i]
	}
	return result, nil
}
