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

// UpdateWithVersion 使用乐观锁版本号更新客户信息
func (r *ClientRepositoryImpl) UpdateWithVersion(ctx context.Context, client *models.Client, expectedVersion uint) error {
	updates := map[string]interface{}{
		"name":           client.Name,
		"type":           client.Type,
		"email":          client.Email,
		"phone":          client.Phone,
		"address":        client.Address,
		"company":        client.Company,
		"id_card":        client.IDCard,
		"industry":       client.Industry,
		"contact_person": client.ContactPerson,
		"contact_phone":  client.ContactPhone,
		"source":         client.Source,
		"notes":          client.Notes,
		"status":         client.Status,
		"version":        expectedVersion + 1,
	}

	result := r.db.WithContext(ctx).
		Model(&models.Client{}).
		Where("id = ? AND version = ?", client.ID, expectedVersion).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrClientVersionConflict
	}

	client.Version = expectedVersion + 1
	return nil
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
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Search != "" {
		// 改进的搜索策略：支持多词搜索
		searchTerms := strings.Fields(strings.TrimSpace(params.Search))

		if len(searchTerms) == 1 {
			// 单词搜索：使用传统LIKE搜索
			searchLower := strings.ToLower(params.Search)
			// 统一使用完整匹配，支持真正的模糊搜索
			fullSearchTerm := "%" + searchLower + "%"
			// 搜索姓名、邮箱、电话和公司字段
			query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone) LIKE ? OR LOWER(company) LIKE ?",
				fullSearchTerm, fullSearchTerm, fullSearchTerm, fullSearchTerm)
		} else {
			// 多词搜索：为每个词创建独立的搜索条件，然后组合
			searchConditions := make([]string, 0, len(searchTerms)*3)
			searchArgs := make([]interface{}, 0, len(searchTerms)*3)

			for _, term := range searchTerms {
				if strings.TrimSpace(term) != "" {
					searchLower := strings.ToLower(term)
					fullSearchTerm := "%" + searchLower + "%"
					searchConditions = append(searchConditions, "LOWER(name) LIKE ?", "LOWER(email) LIKE ?", "LOWER(phone) LIKE ?", "LOWER(company) LIKE ?")
					searchArgs = append(searchArgs, fullSearchTerm, fullSearchTerm, fullSearchTerm, fullSearchTerm)
				}
			}

			if len(searchConditions) > 0 {
				// 使用OR连接所有搜索条件，这样只要包含任何一个词就能匹配
				combinedCondition := strings.Join(searchConditions, " OR ")
				query = query.Where("("+combinedCondition+")", searchArgs...)
			}
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
