package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ClientRepositoryImpl 客户数据仓库的GORM实现
type ClientRepositoryImpl struct {
	db *gorm.DB
}

// NewClientRepository 创建客户数据仓库实例
func NewClientRepository(db *gorm.DB) ClientRepository {
	return &ClientRepositoryImpl{db: db}
}

// Create 创建客户
func (r *ClientRepositoryImpl) Create(ctx context.Context, client *models.Client) error {
	return r.db.WithContext(ctx).Create(client).Error
}

// FindByID 根据ID查找客户
func (r *ClientRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Client, error) {
	var client models.Client
	err := r.db.WithContext(ctx).First(&client, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &client, nil
}

// FindByEmail 根据邮箱查找客户
func (r *ClientRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.Client, error) {
	var client models.Client
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&client).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &client, nil
}

// Update 更新客户信息
func (r *ClientRepositoryImpl) Update(ctx context.Context, client *models.Client) error {
	return r.db.WithContext(ctx).Save(client).Error
}

// Delete 删除客户
func (r *ClientRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Client{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 客户列表查询
func (r *ClientRepositoryImpl) List(ctx context.Context, params *ClientListParams) ([]*models.Client, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.Client{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Search != "" {
		// 性能优化：智能搜索策略
		searchLower := strings.ToLower(params.Search)

		if len(params.Search) >= 3 {
			// 对于3个字符以上的搜索，使用后缀匹配以利用索引
			suffixTerm := searchLower + "%"
			query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone) LIKE ?", suffixTerm, suffixTerm, suffixTerm)
		} else {
			// 对于短搜索词，使用完整匹配但限制搜索范围
			fullSearchTerm := "%" + searchLower + "%"
			// 优先搜索姓名，因为这是最常用的搜索字段
			query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", fullSearchTerm, fullSearchTerm)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var clients []models.Client
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&clients).Error; err != nil {
		return nil, 0, err
	}

	// 转换为指针切片
	result := make([]*models.Client, len(clients))
	for i, client := range clients {
		result[i] = &client
	}

	return result, total, nil
}

// GetStats 获取客户统计信息
func (r *ClientRepositoryImpl) GetStats(ctx context.Context) (*ClientStats, error) {
	stats := &ClientStats{}

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.Client{}).Count(&stats.TotalClients).Error; err != nil {
		return nil, err
	}

	// 获取活跃客户数
	if err := r.db.WithContext(ctx).Model(&models.Client{}).Where("status = ?", "active").Count(&stats.ActiveClients).Error; err != nil {
		return nil, err
	}

	// 获取非活跃客户数
	if err := r.db.WithContext(ctx).Model(&models.Client{}).Where("status = ?", "inactive").Count(&stats.InactiveClients).Error; err != nil {
		return nil, err
	}

	// 获取本月新增客户数
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1).Truncate(24 * time.Hour)
	if err := r.db.WithContext(ctx).Model(&models.Client{}).Where("created_at >= ?", startOfMonth).Count(&stats.NewClientsThisMonth).Error; err != nil {
		return nil, err
	}

	return stats, nil
}
