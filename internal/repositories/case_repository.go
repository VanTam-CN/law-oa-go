package repositories

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// CaseRepositoryImpl 案件数据仓库的GORM实现
type CaseRepositoryImpl struct {
	db *gorm.DB
}

// NewCaseRepository 创建案件数据仓库实例
func NewCaseRepository(db *gorm.DB) CaseRepository {
	return &CaseRepositoryImpl{db: db}
}

// Create 创建案件
func (r *CaseRepositoryImpl) Create(ctx context.Context, caseModel *models.Case) error {
	return r.db.WithContext(ctx).Create(caseModel).Error
}

// FindByID 根据ID查找案件（包含关联数据）
func (r *CaseRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Case, error) {
	var caseModel models.Case
	err := r.db.WithContext(ctx).Preload("Client").Preload("Lawyer").First(&caseModel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &caseModel, nil
}

// Update 更新案件信息
func (r *CaseRepositoryImpl) Update(ctx context.Context, caseModel *models.Case) error {
	return r.db.WithContext(ctx).Save(caseModel).Error
}

// Delete 删除案件
func (r *CaseRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Case{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 案件列表查询（包含关联数据）
func (r *CaseRepositoryImpl) List(ctx context.Context, params *CaseListParams) ([]*models.Case, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.Case{}).Preload("Client").Preload("Lawyer")

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.CaseType != "" {
		query = query.Where("case_type = ?", params.CaseType)
	}
	if params.Priority != "" {
		query = query.Where("priority = ?", params.Priority)
	}
	if params.ClientID > 0 {
		query = query.Where("client_id = ?", params.ClientID)
	}
	if params.LawyerID > 0 {
		query = query.Where("lawyer_id = ?", params.LawyerID)
	}
	if params.Search != "" {
		// 改进的搜索策略：支持多词搜索
		searchTerms := strings.Fields(strings.TrimSpace(params.Search))

		if len(searchTerms) == 1 {
			// 单词搜索：使用传统LIKE搜索
			searchLower := strings.ToLower(params.Search)
			fullSearchTerm := "%" + searchLower + "%"
			query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", fullSearchTerm, fullSearchTerm)
		} else {
			// 多词搜索：为每个词创建独立的搜索条件，然后组合
			// 这样可以找到包含所有搜索词的记录
			searchConditions := make([]string, 0, len(searchTerms)*2)
			searchArgs := make([]interface{}, 0, len(searchTerms)*2)

			for _, term := range searchTerms {
				if strings.TrimSpace(term) != "" {
					searchLower := strings.ToLower(term)
					fullSearchTerm := "%" + searchLower + "%"
					searchConditions = append(searchConditions, "LOWER(title) LIKE ?", "LOWER(description) LIKE ?")
					searchArgs = append(searchArgs, fullSearchTerm, fullSearchTerm)
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

	var cases []models.Case
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&cases).Error; err != nil {
		return nil, 0, err
	}

	// 转换为指针切片
	result := make([]*models.Case, len(cases))
	for i, caseModel := range cases {
		result[i] = &caseModel
	}

	return result, total, nil
}

// GetStats 获取案件统计信息
func (r *CaseRepositoryImpl) GetStats(ctx context.Context) (*CaseStats, error) {
	stats := &CaseStats{}

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.Case{}).Count(&stats.TotalCases).Error; err != nil {
		return nil, err
	}

	// 按状态统计（单次查询）
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	err := r.db.WithContext(ctx).Model(&models.Case{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, err
	}

	// 按优先级统计（单次查询）
	type PriorityCount struct {
		Priority string
		Count    int64
	}
	var priorityCounts []PriorityCount
	err = r.db.WithContext(ctx).Model(&models.Case{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Find(&priorityCounts).Error
	if err != nil {
		return nil, err
	}

	// 填充统计数据
	for _, sc := range statusCounts {
		switch sc.Status {
		case "pending":
			stats.PendingCases = sc.Count
		case "active":
			stats.ActiveCases = sc.Count
		case "closed":
			stats.ClosedCases = sc.Count
		case "suspended":
			stats.SuspendedCases = sc.Count
		}
	}

	for _, pc := range priorityCounts {
		switch pc.Priority {
		case "high":
			stats.HighPriority = pc.Count
		case "urgent":
			stats.UrgentCases = pc.Count
		}
	}

	return stats, nil
}

// AssignLawyer 分配律师
func (r *CaseRepositoryImpl) AssignLawyer(ctx context.Context, caseID, lawyerID uint) error {
	result := r.db.WithContext(ctx).Model(&models.Case{}).Where("id = ?", caseID).Update("lawyer_id", lawyerID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateStatus 更新案件状态
func (r *CaseRepositoryImpl) UpdateStatus(ctx context.Context, caseID uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Case{}).Where("id = ?", caseID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
