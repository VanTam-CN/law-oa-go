package database

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository 仓储模式基类 - 基于最新GORM最佳实践
type Repository[T any] struct {
	db        *gorm.DB
	cache     *SmartCacheManager
	batchSize int
}

// NewRepository 创建新的仓储
func NewRepository[T any](db *gorm.DB, cache *SmartCacheManager, batchSize int) *Repository[T] {
	return &Repository[T]{
		db:        db,
		cache:     cache,
		batchSize: batchSize,
	}
}

// FindByID 根据ID查找
func (r *Repository[T]) FindByID(ctx context.Context, id interface{}) (*T, error) {
	var result T
	cacheKey := r.cache.GetCacheKey("find_by_id", id)

	// 尝试从缓存获取
	if cached, exists := r.cache.SmartGet("find_by_id", cacheKey); exists {
		if model, ok := cached.(T); ok {
			return &model, nil
		}
	}

	// 从数据库查询
	tx := r.db.WithContext(ctx).First(&result, id)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// 缓存结果
	r.cache.SmartSet("find_by_id", cacheKey, result)

	return &result, nil
}

// FindMany 批量查找
func (r *Repository[T]) FindMany(ctx context.Context, ids []interface{}) ([]T, error) {
	var results []T
	if len(ids) == 0 {
		return results, nil
	}

	// 使用GORM的Find方法批量查询
	tx := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&results)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return results, nil
}

// Create 创建记录
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
	tx := r.db.WithContext(ctx).Create(entity)
	if tx.Error != nil {
		return tx.Error
	}

	// 清除相关缓存
	r.cache.InvalidatePattern("find_by_id")
	return nil
}

// CreateBatch 批量创建 - 基于最新GORM性能最佳实践
func (r *Repository[T]) CreateBatch(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return nil
	}

	// 使用优化的批量创建
	tx := r.db.WithContext(ctx).CreateInBatches(entities, r.batchSize)
	if tx.Error != nil {
		return tx.Error
	}

	// 清除相关缓存
	r.cache.InvalidatePattern("find_by_id")
	return nil
}

// Update 更新记录
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
	tx := r.db.WithContext(ctx).Save(entity)
	if tx.Error != nil {
		return tx.Error
	}

	// 清除相关缓存
	r.cache.InvalidatePattern("find_by_id")
	return nil
}

// UpdateFields 更新特定字段
func (r *Repository[T]) UpdateFields(ctx context.Context, id interface{}, updates map[string]interface{}) error {
	var model T
	tx := r.db.WithContext(ctx).Model(&model).Where("id = ?", id).Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}

	// 清除相关缓存
	r.cache.InvalidatePattern("find_by_id")
	return nil
}

// Delete 删除记录
func (r *Repository[T]) Delete(ctx context.Context, id interface{}) error {
	var model T
	tx := r.db.WithContext(ctx).Delete(&model, id)
	if tx.Error != nil {
		return tx.Error
	}

	// 清除相关缓存
	r.cache.InvalidatePattern("find_by_id")
	return nil
}

// DeleteBatch 批量删除
func (r *Repository[T]) DeleteBatch(ctx context.Context, ids []interface{}) error {
	if len(ids) == 0 {
		return nil
	}

	var model T
	tx := r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model)
	if tx.Error != nil {
		return tx.Error
	}

	// 清除相关缓存
	r.cache.InvalidatePattern("find_by_id")
	return nil
}

// Exists 检查记录是否存在
func (r *Repository[T]) Exists(ctx context.Context, conditions map[string]interface{}) (bool, error) {
	var count int64
	var model T
	tx := r.db.WithContext(ctx).Model(&model)

	for key, value := range conditions {
		tx = tx.Where(key, value)
	}

	err := tx.Count(&count).Error
	return count > 0, err
}

// Count 统计记录数量
func (r *Repository[T]) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	var model T
	tx := r.db.WithContext(ctx).Model(&model)

	for key, value := range conditions {
		tx = tx.Where(key, value)
	}

	err := tx.Count(&count).Error
	return count, err
}

// Upsert 插入或更新
func (r *Repository[T]) Upsert(ctx context.Context, entity *T, conflictColumns []string) error {
	// 使用GORM的OnConflict功能
	columns := make([]clause.Column, len(conflictColumns))
	for i, col := range conflictColumns {
		columns[i] = clause.Column{Name: col}
	}

	tx := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   columns,
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).Create(entity)

	if tx.Error != nil {
		return tx.Error
	}

	// 清除相关缓存
	r.cache.InvalidatePattern("find_by_id")
	return nil
}

// HealthCheck 健康检查
func (r *Repository[T]) HealthCheck(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("SELECT 1").Error
}