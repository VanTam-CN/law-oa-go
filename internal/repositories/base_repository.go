package repositories

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// BaseRepository 基础仓储实现 (Go 1.23+ 泛型支持)
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
		return NewRepositoryError("create", fmt.Sprintf("%T", *entity), err)
	}
	return nil
}

// GetByID 根据ID获取实体
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	err := r.db.WithContext(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("get by id", fmt.Sprintf("%T", entity), id, ErrRecordNotFound)
		}
		return nil, NewRepositoryError("get by id", fmt.Sprintf("%T", entity), err)
	}
	return &entity, nil
}

// Update 更新实体
func (r *BaseRepository[T]) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	var entity T
	result := r.db.WithContext(ctx).Model(&entity).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return NewRepositoryError("update", fmt.Sprintf("%T", entity), result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update", fmt.Sprintf("%T", entity), id, ErrRecordNotFound)
	}
	return nil
}

// Delete 删除实体
func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	result := r.db.WithContext(ctx).Delete(&entity, id)
	if result.Error != nil {
		return NewRepositoryError("delete", fmt.Sprintf("%T", entity), result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("delete", fmt.Sprintf("%T", entity), id, ErrRecordNotFound)
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
		return nil, 0, NewRepositoryError("count", fmt.Sprintf("%T", *new(T)), err)
	}

	// 获取数据
	if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, NewRepositoryError("list", fmt.Sprintf("%T", *new(T)), err)
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
		return 0, NewRepositoryError("count", fmt.Sprintf("%T", *new(T)), err)
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
			return nil, NewRepositoryErrorWithID("find with preload", fmt.Sprintf("%T", entity), id, ErrRecordNotFound)
		}
		return nil, NewRepositoryError("find with preload", fmt.Sprintf("%T", entity), err)
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

// WhereIn 添加IN条件
func (qb *QueryBuilder[T]) WhereIn(column string, values []interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Where(column+" IN ?", values)
	return qb
}

// WhereNot 添加NOT条件
func (qb *QueryBuilder[T]) WhereNot(condition string, args ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Not(condition, args...)
	return qb
}

// WhereLike 添加LIKE条件
func (qb *QueryBuilder[T]) WhereLike(column string, pattern string) *QueryBuilder[T] {
	qb.query = qb.query.Where(column+" LIKE ?", pattern)
	return qb
}

// Order 添加排序
func (qb *QueryBuilder[T]) Order(order string) *QueryBuilder[T] {
	qb.query = qb.query.Order(order)
	return qb
}

// OrderDesc 添加降序排序
func (qb *QueryBuilder[T]) OrderDesc(column string) *QueryBuilder[T] {
	qb.query = qb.query.Order(column + " DESC")
	return qb
}

// OrderAsc 添加升序排序
func (qb *QueryBuilder[T]) OrderAsc(column string) *QueryBuilder[T] {
	qb.query = qb.query.Order(column + " ASC")
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

// Join 添加JOIN
func (qb *QueryBuilder[T]) Join(column string, conditions ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Joins(column, conditions...)
	return qb
}

// LeftJoin 添加LEFT JOIN
func (qb *QueryBuilder[T]) LeftJoin(column string, conditions ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Joins("LEFT JOIN "+column, conditions...)
	return qb
}

// Group 添加GROUP BY
func (qb *QueryBuilder[T]) Group(column string) *QueryBuilder[T] {
	qb.query = qb.query.Group(column)
	return qb
}

// Having 添加HAVING
func (qb *QueryBuilder[T]) Having(condition string, args ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Having(condition, args...)
	return qb
}

// Distinct 添加DISTINCT
func (qb *QueryBuilder[T]) Distinct() *QueryBuilder[T] {
	qb.query = qb.query.Distinct()
	return qb
}

// Select 添加SELECT字段
func (qb *QueryBuilder[T]) Select(query string, args ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Select(query, args...)
	return qb
}

// Find 执行查询
func (qb *QueryBuilder[T]) Find(ctx context.Context) ([]*T, error) {
	var entities []*T
	if err := qb.query.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, NewRepositoryError("find", fmt.Sprintf("%T", *new(T)), err)
	}
	return entities, nil
}

// First 获取第一个
func (qb *QueryBuilder[T]) First(ctx context.Context) (*T, error) {
	var entity T
	err := qb.query.WithContext(ctx).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryError("first", fmt.Sprintf("%T", entity), ErrRecordNotFound)
		}
		return nil, NewRepositoryError("first", fmt.Sprintf("%T", entity), err)
	}
	return &entity, nil
}

// Count 计数
func (qb *QueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := qb.query.WithContext(ctx).Count(&count).Error; err != nil {
		return 0, NewRepositoryError("count", fmt.Sprintf("%T", *new(T)), err)
	}
	return count, nil
}

// Exists 检查是否存在
func (qb *QueryBuilder[T]) Exists(ctx context.Context) (bool, error) {
	count, err := qb.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Scan 扫描到自定义结构体
func (qb *QueryBuilder[T]) Scan(ctx context.Context, dest interface{}) error {
	if err := qb.query.WithContext(ctx).Scan(dest).Error; err != nil {
		return NewRepositoryError("scan", fmt.Sprintf("%T", *new(T)), err)
	}
	return nil
}

// Raw 执行原生SQL
func (qb *QueryBuilder[T]) Raw(sql string, args ...interface{}) *QueryBuilder[T] {
	qb.query = qb.query.Raw(sql, args...)
	return qb
}

// Exec 执行无返回结果的SQL
func (qb *QueryBuilder[T]) Exec(ctx context.Context, sql string, args ...interface{}) error {
	if err := qb.query.WithContext(ctx).Exec(sql, args...).Error; err != nil {
		return NewRepositoryError("exec", fmt.Sprintf("%T", *new(T)), err)
	}
	return nil
}

// AddErrRecordNotFound 添加通用的记录未找到错误
var ErrRecordNotFound = errors.New("record not found")
