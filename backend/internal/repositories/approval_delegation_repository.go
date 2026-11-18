package repositories

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// ApprovalDelegationRepository 代理审批仓储接口
type ApprovalDelegationRepository interface {
	Create(ctx context.Context, delegation *models.ApprovalDelegation) error
	GetByID(ctx context.Context, id string) (*models.ApprovalDelegation, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	GetActiveDelegations(ctx context.Context, delegatorID string) ([]*models.ApprovalDelegation, error)
	GetValidDelegate(ctx context.Context, delegatorID string) (*models.ApprovalDelegation, error)
	List(ctx context.Context, params *DelegationListParams) ([]*models.ApprovalDelegation, int64, error)
}

// DelegationListParams 代理配置查询参数
type DelegationListParams struct {
	DelegatorID string
	DelegateID  string
	IsActive   *bool
	Page       int
	PageSize   int
}

// approvalDelegationRepository 代理审批仓储实现
type approvalDelegationRepository struct {
	db *gorm.DB
}

// NewApprovalDelegationRepository 创建代理审批仓储
func NewApprovalDelegationRepository(db *gorm.DB) ApprovalDelegationRepository {
	return &approvalDelegationRepository{db: db}
}

func (r *approvalDelegationRepository) Create(ctx context.Context, delegation *models.ApprovalDelegation) error {
	if err := r.db.WithContext(ctx).Create(delegation).Error; err != nil {
		return fmt.Errorf("创建代理配置失败: %w", err)
	}
	return nil
}

func (r *approvalDelegationRepository) GetByID(ctx context.Context, id string) (*models.ApprovalDelegation, error) {
	var d models.ApprovalDelegation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&d).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取代理配置失败: %w", err)
	}
	return &d, nil
}

func (r *approvalDelegationRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&models.ApprovalDelegation{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新代理配置失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *approvalDelegationRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.ApprovalDelegation{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("删除代理配置失败: %w", result.Error)
	}
	return nil
}

// GetActiveDelegations 获取与用户相关的所有活跃代理配置（作为委托人或作为代理人）
func (r *approvalDelegationRepository) GetActiveDelegations(ctx context.Context, userID string) ([]*models.ApprovalDelegation, error) {
	var delegations []*models.ApprovalDelegation
	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("valid_from <= ? AND (valid_until IS NULL OR valid_until >= ?)", now, now).
		Where("delegator_id = ? OR delegate_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&delegations).Error
	if err != nil {
		return nil, fmt.Errorf("获取代理配置失败: %w", err)
	}
	return delegations, nil
}

// GetValidDelegate 获取委托人当前有效的代理人
// 优先返回有效期内最近创建的配置
func (r *approvalDelegationRepository) GetValidDelegate(ctx context.Context, delegatorID string) (*models.ApprovalDelegation, error) {
	var d models.ApprovalDelegation
	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("delegator_id = ? AND is_active = ?", delegatorID, true).
		Where("valid_from <= ? AND (valid_until IS NULL OR valid_until >= ?)", now, now).
		Order("created_at DESC").
		First(&d).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取有效代理失败: %w", err)
	}
	return &d, nil
}

// List 列表查询代理配置
func (r *approvalDelegationRepository) List(ctx context.Context, params *DelegationListParams) ([]*models.ApprovalDelegation, int64, error) {
	var delegations []*models.ApprovalDelegation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ApprovalDelegation{})

	if params.DelegatorID != "" {
		query = query.Where("delegator_id = ?", params.DelegatorID)
	}
	if params.DelegateID != "" {
		query = query.Where("delegate_id = ?", params.DelegateID)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计代理配置数量失败: %w", err)
	}

	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&delegations).Error; err != nil {
		return nil, 0, fmt.Errorf("查询代理配置失败: %w", err)
	}

	return delegations, total, nil
}
