package repositories

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/models/elasticsearch"

	"gorm.io/gorm"
)

// LegalStatuteRepository 法条仓储接口
type LegalStatuteRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, statute *models.LegalStatute) error
	GetByID(ctx context.Context, id int) (*models.LegalStatute, error)
	GetByStatuteNumber(ctx context.Context, number string) (*models.LegalStatute, error)
	Update(ctx context.Context, statute *models.LegalStatute) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*models.LegalStatute, error)
	Count(ctx context.Context) (int64, error)

	// 高级查询操作
	FindByCategory(ctx context.Context, categoryID int, offset, limit int) ([]*models.LegalStatute, error)
	FindByLawName(ctx context.Context, lawName string, offset, limit int) ([]*models.LegalStatute, error)
	FindActive(ctx context.Context, offset, limit int) ([]*models.LegalStatute, error)
	SearchByKeyword(ctx context.Context, keyword string, offset, limit int) ([]*models.LegalStatute, error)
	FindByTags(ctx context.Context, tags []string, offset, limit int) ([]*models.LegalStatute, error)

	// 层级关系操作
	GetChildren(ctx context.Context, parentID int) ([]*models.LegalStatute, error)
	GetHierarchy(ctx context.Context, statuteID int) ([]*models.HierarchyItem, error)
	GetRootStatutes(ctx context.Context) ([]*models.LegalStatute, error)

	// 版本管理操作
	CreateVersion(ctx context.Context, version *models.LegalStatuteVersion) error
	GetVersions(ctx context.Context, statuteID int) ([]*models.LegalStatuteVersion, error)
	GetVersion(ctx context.Context, statuteID, versionNumber int) (*models.LegalStatuteVersion, error)

	// 用户收藏操作
	AddToFavorites(ctx context.Context, userID, statuteID int) error
	RemoveFromFavorites(ctx context.Context, userID, statuteID int) error
	GetUserFavorites(ctx context.Context, userID int, offset, limit int) ([]*models.LegalStatute, error)
	IsFavorited(ctx context.Context, userID, statuteID int) (bool, error)

	// 搜索历史操作
	CreateSearchHistory(ctx context.Context, history *models.LegalSearchHistory) error
	GetUserSearchHistory(ctx context.Context, userID int, limit int) ([]*models.LegalSearchHistory, error)
	GetPopularSearches(ctx context.Context, limit int) ([]string, error)

	// 统计操作
	GetCategoryStats(ctx context.Context) ([]*models.CategoryStat, error)
	GetPopularStatutes(ctx context.Context, limit int) ([]*models.LegalStatute, error)
	GetRecentUpdates(ctx context.Context, days int) ([]*models.LegalStatute, error)
}

// LegalCategoryRepository 法条分类仓储接口
type LegalCategoryRepository interface {
	Create(ctx context.Context, category *models.LegalCategory) error
	GetByID(ctx context.Context, id int) (*models.LegalCategory, error)
	GetByCode(ctx context.Context, code string) (*models.LegalCategory, error)
	Update(ctx context.Context, category *models.LegalCategory) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*models.LegalCategory, error)
	GetTree(ctx context.Context) ([]*models.CategoryTreeNode, error)
	GetChildren(ctx context.Context, parentID int) ([]*models.LegalCategory, error)
}

// LegalTagRepository 法条标签仓储接口
type LegalTagRepository interface {
	Create(ctx context.Context, tag *models.LegalTag) error
	GetByID(ctx context.Context, id int) (*models.LegalTag, error)
	GetByName(ctx context.Context, name string) (*models.LegalTag, error)
	Update(ctx context.Context, tag *models.LegalTag) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*models.LegalTag, error)
	GetPopularTags(ctx context.Context, limit int) ([]*models.LegalTag, error)
	UpdateUsageCount(ctx context.Context, tagID int) error
}

// ElasticsearchStatuteRepository Elasticsearch法条搜索仓储接口
type ElasticsearchStatuteRepository interface {
	// 索引管理
	CreateIndex(ctx context.Context) error
	DeleteIndex(ctx context.Context) error
	IndexDocument(ctx context.Context, doc *elasticsearch.LegalStatuteDocument) error
	BulkIndexDocuments(ctx context.Context, docs []*elasticsearch.LegalStatuteDocument) error

	// 搜索操作
	Search(ctx context.Context, req *elasticsearch.LegalSearchRequest) (*elasticsearch.LegalSearchResponse, error)
	GetSuggestion(ctx context.Context, query string) ([]string, error)
	GetRelatedStatutes(ctx context.Context, statuteID int, limit int) ([]*elasticsearch.LegalStatuteDocument, error)

	// 聚合查询
	GetCategoryAggregation(ctx context.Context) (map[string]int64, error)
	GetLawNameAggregation(ctx context.Context) (map[string]int64, error)
	GetTagAggregation(ctx context.Context) (map[string]int64, error)
}

// legalStatuteRepository 法条仓储实现
type legalStatuteRepository struct {
	db *gorm.DB
}

// NewLegalStatuteRepository 创建法条仓储实例
func NewLegalStatuteRepository(db *gorm.DB) LegalStatuteRepository {
	return &legalStatuteRepository{
		db: db,
	}
}

// Create 创建法条
func (r *legalStatuteRepository) Create(ctx context.Context, statute *models.LegalStatute) error {
	return r.db.WithContext(ctx).Create(statute).Error
}

// GetByID 根据ID获取法条
func (r *legalStatuteRepository) GetByID(ctx context.Context, id int) (*models.LegalStatute, error) {
	var statute models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("ParentStatute").
		Preload("TagsRelation.Tag").
		First(&statute, id).Error
	if err != nil {
		return nil, err
	}
	return &statute, nil
}

// GetByStatuteNumber 根据法条编号获取法条
func (r *legalStatuteRepository) GetByStatuteNumber(ctx context.Context, number string) (*models.LegalStatute, error) {
	var statute models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("ParentStatute").
		Preload("TagsRelation.Tag").
		Where("statute_number = ?", number).
		First(&statute).Error
	if err != nil {
		return nil, err
	}
	return &statute, nil
}

// Update 更新法条
func (r *legalStatuteRepository) Update(ctx context.Context, statute *models.LegalStatute) error {
	return r.db.WithContext(ctx).Save(statute).Error
}

// Delete 删除法条
func (r *legalStatuteRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.LegalStatute{}, id).Error
}

// List 获取法条列表
func (r *legalStatuteRepository) List(ctx context.Context, offset, limit int) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&statutes).Error
	return statutes, err
}

// Count 获取法条总数
func (r *legalStatuteRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.LegalStatute{}).Count(&count).Error
	return count, err
}

// FindByCategory 根据分类查找法条
func (r *legalStatuteRepository) FindByCategory(ctx context.Context, categoryID int, offset, limit int) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("category_id = ?", categoryID).
		Offset(offset).
		Limit(limit).
		Order("order_in_hierarchy ASC, statute_number ASC").
		Find(&statutes).Error
	return statutes, err
}

// FindByLawName 根据法律名称查找法条
func (r *legalStatuteRepository) FindByLawName(ctx context.Context, lawName string, offset, limit int) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("law_name = ?", lawName).
		Offset(offset).
		Limit(limit).
		Order("order_in_hierarchy ASC, statute_number ASC").
		Find(&statutes).Error
	return statutes, err
}

// FindActive 查找生效中的法条
func (r *legalStatuteRepository) FindActive(ctx context.Context, offset, limit int) ([]*models.LegalStatute, error) {
	now := time.Now()
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("status = ? AND (effective_date IS NULL OR effective_date <= ?) AND (expiry_date IS NULL OR expiry_date > ?)",
			"active", now, now).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&statutes).Error
	return statutes, err
}

// SearchByKeyword 关键词搜索
func (r *legalStatuteRepository) SearchByKeyword(ctx context.Context, keyword string, offset, limit int) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	searchPattern := "%" + keyword + "%"

	// 添加调试日志
	fmt.Printf("DEBUG: 搜索关键词: %s, 模式: %s\n", keyword, searchPattern)

	// 先测试简单查询
	var testCount int64
	r.db.WithContext(ctx).Model(&models.LegalStatute{}).Count(&testCount)
	fmt.Printf("DEBUG: 法条表总数: %d\n", testCount)

	// 测试是否有隐私权相关的记录
	var privacyCount int64
	r.db.WithContext(ctx).Model(&models.LegalStatute{}).Where("title ILIKE ?", "%隐私权%").Count(&privacyCount)
	fmt.Printf("DEBUG: 包含隐私权的记录数: %d\n", privacyCount)

	query := r.db.WithContext(ctx).
		Preload("Category").
		Where("title ILIKE ? OR content ILIKE ? OR statute_number ILIKE ? OR ? = ANY(keywords)",
			searchPattern, searchPattern, searchPattern, keyword).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC")

	err := query.Find(&statutes).Error
	fmt.Printf("DEBUG: 搜索结果: 找到 %d 条记录, 错误: %v\n", len(statutes), err)

	return statutes, err
}

// FindByTags 根据标签查找法条
func (r *legalStatuteRepository) FindByTags(ctx context.Context, tags []string, offset, limit int) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("TagsRelation.Tag").
		Joins("JOIN legal_statute_tags ON legal_statutes.id = legal_statute_tags.statute_id").
		Joins("JOIN legal_tags ON legal_statute_tags.tag_id = legal_tags.id").
		Where("legal_tags.name = ANY(?)", tags).
		Distinct().
		Offset(offset).
		Limit(limit).
		Order("legal_statutes.created_at DESC").
		Find(&statutes).Error
	return statutes, err
}

// GetChildren 获取子法条
func (r *legalStatuteRepository) GetChildren(ctx context.Context, parentID int) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("parent_statute_id = ?", parentID).
		Order("order_in_hierarchy ASC, statute_number ASC").
		Find(&statutes).Error
	return statutes, err
}

// GetHierarchy 获取法条层级结构
func (r *legalStatuteRepository) GetHierarchy(ctx context.Context, statuteID int) ([]*models.HierarchyItem, error) {
	var items []*models.HierarchyItem
	err := r.db.WithContext(ctx).
		Table("legal_hierarchy").
		Select("legal_hierarchy.depth, legal_statutes.*").
		Joins("JOIN legal_statutes ON legal_statutes.id = legal_hierarchy.descendant_id").
		Where("legal_hierarchy.ancestor_id = ?", statuteID).
		Order("legal_hierarchy.depth ASC, legal_statutes.order_in_hierarchy ASC").
		Find(&items).Error
	return items, err
}

// GetRootStatutes 获取根级法条
func (r *legalStatuteRepository) GetRootStatutes(ctx context.Context) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("parent_statute_id IS NULL").
		Order("order_in_hierarchy ASC, statute_number ASC").
		Find(&statutes).Error
	return statutes, err
}

// CreateVersion 创建版本记录
func (r *legalStatuteRepository) CreateVersion(ctx context.Context, version *models.LegalStatuteVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersions 获取法条版本历史
func (r *legalStatuteRepository) GetVersions(ctx context.Context, statuteID int) ([]*models.LegalStatuteVersion, error) {
	var versions []*models.LegalStatuteVersion
	err := r.db.WithContext(ctx).
		Preload("Creator").
		Where("statute_id = ?", statuteID).
		Order("version_number DESC").
		Find(&versions).Error
	return versions, err
}

// GetVersion 获取指定版本
func (r *legalStatuteRepository) GetVersion(ctx context.Context, statuteID, versionNumber int) (*models.LegalStatuteVersion, error) {
	var version models.LegalStatuteVersion
	err := r.db.WithContext(ctx).
		Preload("Creator").
		Where("statute_id = ? AND version_number = ?", statuteID, versionNumber).
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// AddToFavorites 添加到收藏
func (r *legalStatuteRepository) AddToFavorites(ctx context.Context, userID, statuteID int) error {
	favorite := &models.UserLegalFavorite{
		UserID:    userID,
		StatuteID: statuteID,
	}
	return r.db.WithContext(ctx).Create(favorite).Error
}

// RemoveFromFavorites 移除收藏
func (r *legalStatuteRepository) RemoveFromFavorites(ctx context.Context, userID, statuteID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND statute_id = ?", userID, statuteID).
		Delete(&models.UserLegalFavorite{}).Error
}

// GetUserFavorites 获取用户收藏
func (r *legalStatuteRepository) GetUserFavorites(ctx context.Context, userID int, offset, limit int) ([]*models.LegalStatute, error) {
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Joins("JOIN user_legal_favorites ON legal_statutes.id = user_legal_favorites.statute_id").
		Where("user_legal_favorites.user_id = ?", userID).
		Order("user_legal_favorites.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&statutes).Error
	return statutes, err
}

// IsFavorited 检查是否已收藏
func (r *legalStatuteRepository) IsFavorited(ctx context.Context, userID, statuteID int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserLegalFavorite{}).
		Where("user_id = ? AND statute_id = ?", userID, statuteID).
		Count(&count).Error
	return count > 0, err
}

// CreateSearchHistory 创建搜索历史
func (r *legalStatuteRepository) CreateSearchHistory(ctx context.Context, history *models.LegalSearchHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetUserSearchHistory 获取用户搜索历史
func (r *legalStatuteRepository) GetUserSearchHistory(ctx context.Context, userID int, limit int) ([]*models.LegalSearchHistory, error) {
	var histories []*models.LegalSearchHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&histories).Error
	return histories, err
}

// GetPopularSearches 获取热门搜索
func (r *legalStatuteRepository) GetPopularSearches(ctx context.Context, limit int) ([]string, error) {
	var queries []string
	err := r.db.WithContext(ctx).
		Model(&models.LegalSearchHistory{}).
		Select("search_query").
		Group("search_query").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("search_query", &queries).Error
	return queries, err
}

// GetCategoryStats 获取分类统计
func (r *legalStatuteRepository) GetCategoryStats(ctx context.Context) ([]*models.CategoryStat, error) {
	var stats []*models.CategoryStat
	err := r.db.WithContext(ctx).
		Table("legal_categories").
		Select("legal_categories.*, COUNT(legal_statutes.id) as statute_count").
		Joins("LEFT JOIN legal_statutes ON legal_categories.id = legal_statutes.category_id").
		Group("legal_categories.id").
		Find(&stats).Error
	return stats, err
}

// GetPopularStatutes 获取热门法条
func (r *legalStatuteRepository) GetPopularStatutes(ctx context.Context, limit int) ([]*models.LegalStatute, error) {
	// 这里可以根据浏览量、收藏量等排序
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Order("created_at DESC"). // 临时按创建时间排序，后续可以添加浏览量字段
		Limit(limit).
		Find(&statutes).Error
	return statutes, err
}

// GetRecentUpdates 获取最近更新的法条
func (r *legalStatuteRepository) GetRecentUpdates(ctx context.Context, days int) ([]*models.LegalStatute, error) {
	since := time.Now().AddDate(0, 0, -days)
	var statutes []*models.LegalStatute
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("updated_at >= ?", since).
		Order("updated_at DESC").
		Find(&statutes).Error
	return statutes, err
}
