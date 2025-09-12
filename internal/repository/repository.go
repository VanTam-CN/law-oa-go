package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"law-oa-go/internal/common"
)

// Repository 泛型仓储接口 (Go 1.23+ 特性)
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uint) (*T, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int, conditions map[string]interface{}) ([]*T, int64, error)
	Count(ctx context.Context, conditions map[string]interface{}) (int64, error)
}

// BaseRepository 基础仓储实现
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// Create 创建实体
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}
	return nil
}

// GetByID 根据ID获取实体
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: id=%d", common.ErrRecordNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}
	return &entity, nil
}

// Update 更新实体
func (r *BaseRepository[T]) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	var entity T
	result := r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("%w: %v", common.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: id=%d", common.ErrRecordNotFound, id)
	}
	return nil
}

// Delete 删除实体
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	result := r.db.WithContext(ctx).Delete(&entity, id)
	if result.Error != nil {
		return fmt.Errorf("%w: %v", common.ErrDatabaseError, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: id=%d", common.ErrRecordNotFound, id)
	}
	return nil
}

// List 列表查询
func (r *BaseRepository[T]) List(ctx context.Context, offset, limit int, conditions map[string]interface{}) ([]*T, int64, error) {
	var entities []*T
	var total int64

	query := r.db.WithContext(ctx).Model(new(T))
	
	// 应用条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}

	// 获取数据
	if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}

	return entities, total, nil
}

// Count 计数查询
func (r *BaseRepository[T]) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))
	
	for key, value := range conditions {
		query = query.Where(key, value)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}

	return count, nil
}

// BatchCreate 批量创建 (Go 1.23+ 优化)
func (r *BaseRepository[T]) BatchCreate(ctx context.Context, entities []*T, batchSize int) error {
	if len(entities) == 0 {
		return nil
	}

	if batchSize <= 0 {
		batchSize = 100
	}

	return r.db.WithContext(ctx).CreateInBatches(entities, batchSize).Error
}

// FindWithPreload 带预加载的查询
func (r *BaseRepository[T]) FindWithPreload(ctx context.Context, id uint, preloads ...string) (*T, error) {
	var entity T
	query := r.db.WithContext(ctx)
	
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	err := query.First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: id=%d", common.ErrRecordNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}

	return &entity, nil
}

// Transaction 事务操作 (Go 1.23+ 函数式编程)
func (r *BaseRepository[T]) Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// QueryBuilder 查询构建器 (Go 1.23+ 链式调用优化)
type QueryBuilder[T any] struct {
	db    *gorm.DB
	query *gorm.DB
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder[T any](db *gorm.DB) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		db:    db,
		query: db.Model(new(T)),
	}
}

// Where 添加条件
func (qb *QueryBuilder[T]) Where(condition string, args ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Where(condition, args...)
	return qb
}

// Order 添加排序
func (qb *QueryBuilder[T]) Order(order string) *QueryBuilder[T] {
	qb.query = qb.query.Order(order)
	return qb
}

// Limit 限制数量
func (qb *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	qb.query = qb.query.Limit(limit)
	return qb
}

// Offset 偏移量
func (qb *QueryBuilder[T]) Offset(offset int) *QueryBuilder[T] {
	qb.query = qb.query.Offset(offset)
	return qb
}

// Preload 预加载
func (qb *QueryBuilder[T]) Preload(column string, conditions ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Preload(column, conditions...)
	return qb
}

// Find 执行查询
func (qb *QueryBuilder[T]) Find(ctx context.Context) ([]*T, error) {
	var entities []*T
	if err := qb.query.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}
	return entities, nil
}

// First 获取第一个
func (qb *QueryBuilder[T]) First(ctx context.Context) (*T, error) {
	var entity T
	err := qb.query.WithContext(ctx).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w", common.ErrRecordNotFound)
		}
		return nil, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}
	return &entity, nil
}

// Count 计数
func (qb *QueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := qb.query.WithContext(ctx).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%w: %v", common.ErrDatabaseError, err)
	}
	return count, nil
}