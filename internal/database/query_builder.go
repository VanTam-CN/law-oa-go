package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// QueryBuilder 高级查询构建器 - 基于最新GORM最佳实践
type QueryBuilder struct {
	db     *gorm.DB
	ctx    context.Context
	tables []string
}

// NewQueryBuilder 创建新的查询构建器
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:  db,
		ctx: context.Background(),
	}
}

// WithContext 设置上下文
func (qb *QueryBuilder) WithContext(ctx context.Context) *QueryBuilder {
	qb.ctx = ctx
	return qb
}

// BatchCreate 批量创建 - 基于最新GORM性能最佳实践
func (qb *QueryBuilder) BatchCreate(records interface{}, batchSize int) error {
	// 验证输入
	if records == nil {
		return fmt.Errorf("records cannot be nil")
	}

	// 自动检测批次大小
	if batchSize <= 0 {
		batchSize = 1000 // 默认批次大小
	}

	// 使用GORM的CreateBatchSize优化
	return qb.db.WithContext(qb.ctx).Session(&gorm.Session{
		CreateBatchSize: batchSize,
	}).Create(records).Error
}

// SmartPreload 智能预加载 - 避免N+1查询问题
func (qb *QueryBuilder) SmartPreload(preloads map[string]interface{}) *gorm.DB {
	if len(preloads) == 0 {
		return qb.db.WithContext(qb.ctx)
	}

	tx := qb.db.WithContext(qb.ctx)

	// 按优先级处理预加载
	highPriority := []string{"Roles", "Permissions"}
	lowPriority := []string{"AuditLogs", "Sessions"}

	// 处理高优先级预加载
	for _, field := range highPriority {
		if _, exists := preloads[field]; exists {
			tx = tx.Preload(field)
		}
	}

	// 处理自定义预加载
	for field, conditions := range preloads {
		// 跳过已处理的高优先级字段
		for _, hp := range highPriority {
			if field == hp {
				goto nextField
			}
		}

		// 跳过低优先级字段，稍后处理
		for _, lp := range lowPriority {
			if field == lp {
				goto nextField
			}
		}

		// 添加条件预加载
		if conditions != nil {
			tx = tx.Preload(field, conditions)
		} else {
			tx = tx.Preload(field)
		}

	nextField:
	}

	// 处理低优先级预加载（可选）
	for _, field := range lowPriority {
		if _, exists := preloads[field]; exists {
			tx = tx.Preload(field)
		}
	}

	return tx
}

// OptimizeQuery 优化查询 - 基于最新PostgreSQL最佳实践
func (qb *QueryBuilder) OptimizeQuery() *gorm.DB {
	return qb.db.WithContext(qb.ctx).
		// 启用查询优化
		Session(&gorm.Session{
			PrepareStmt: true, // 启用预编译语句
			// SkipDefaultTransaction: true, // 对于查询操作禁用默认事务
		})
}

// PaginateQuery 分页查询优化
func (qb *QueryBuilder) PaginateQuery(page, pageSize int, total *int64) *gorm.DB {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	tx := qb.OptimizeQuery()

	// 如果需要总数统计，单独执行COUNT查询（性能更好）
	if total != nil {
		var countTx *gorm.DB
		if len(qb.tables) > 0 {
			countTx = tx.Table(qb.tables[0])
		} else {
			// 如果没有指定表，使用默认模型
			countTx = tx.Model(&struct{}{})
		}
		countTx.Count(total)
	}

	return tx.Offset(offset).Limit(pageSize)
}

// ScopedQuery 范围化查询 - 基于GORM Scopes
func (qb *QueryBuilder) ScopedQuery(scopes ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	tx := qb.OptimizeQuery()
	for _, scope := range scopes {
		tx = scope(tx)
	}
	return tx
}

// CacheableQuery 可缓存查询
func (qb *QueryBuilder) CacheableQuery(cacheKey string, cacheTTL time.Duration, query func(*gorm.DB) *gorm.DB) *gorm.DB {
	// 尝试从缓存获取
	// if result := qb.cache.Get(cacheKey); result != nil {
	//     return result
	// }

	// 执行查询
	tx := query(qb.OptimizeQuery())

	// 存储到缓存
	// qb.cache.Set(cacheKey, result, cacheTTL)

	return tx
}

// BulkUpdate 批量更新 - 优化的批量操作
func (qb *QueryBuilder) BulkUpdate(model interface{}, updates map[string]interface{}, conditions map[string]interface{}) error {
	return qb.db.WithContext(qb.ctx).Model(model).
		Where(conditions).
		Updates(updates).Error
}

// BulkDelete 批量删除 - 安全的批量删除
func (qb *QueryBuilder) BulkDelete(model interface{}, conditions map[string]interface{}) error {
	return qb.db.WithContext(qb.ctx).Model(model).
		Where(conditions).
		Delete(model).Error
}

// JoinOptimizedQuery 优化的连接查询
func (qb *QueryBuilder) JoinOptimizedQuery(joins []string, fields []string) *gorm.DB {
	tx := qb.OptimizeQuery()

	// 添加连接
	for _, join := range joins {
		tx = tx.Joins(join)
	}

	// 选择特定字段以优化性能
	if len(fields) > 0 {
		tx = tx.Select(fields)
	}

	return tx
}

// SubqueryOptimized 子查询优化
func (qb *QueryBuilder) SubqueryOptimized(table string, subquery *gorm.DB, alias string) *gorm.DB {
	return qb.OptimizeQuery().Table(table).
		Where(fmt.Sprintf("%s IN (?)", alias), subquery)
}

// AggregateQuery 聚合查询优化
func (qb *QueryBuilder) AggregateQuery(aggregateFunc string, field string, groupBy []string, having string) *gorm.DB {
	tx := qb.OptimizeQuery()

	// 构建聚合选择
	selectClause := fmt.Sprintf("%s(%s) as aggregate_result", aggregateFunc, field)
	tx = tx.Select(selectClause)

	// 添加分组
	if len(groupBy) > 0 {
		tx = tx.Group(strings.Join(groupBy, ", "))
	}

	// 添加HAVING条件
	if having != "" {
		tx = tx.Having(having)
	}

	return tx
}

// OptimizedSearch 优化搜索查询
func (qb *QueryBuilder) OptimizedSearch(searchTerm string, searchFields []string, additionalFilters map[string]interface{}) *gorm.DB {
	tx := qb.OptimizeQuery()

	if searchTerm == "" {
		return qb.OptimizeQuery()
	}

	// 使用全文搜索或LIKE查询（根据数据库类型）
	conditions := make([]interface{}, 0, len(searchFields))
	whereClause := ""

	for i, field := range searchFields {
		if i > 0 {
			whereClause += " OR "
		}
		whereClause += fmt.Sprintf("%s ILIKE ?", field)
		conditions = append(conditions, "%"+searchTerm+"%")
	}

	tx = tx.Where(whereClause, conditions...)

	// 添加额外过滤条件
	for key, value := range additionalFilters {
		tx = tx.Where(key, value)
	}

	return tx
}

// OptimizedOrderBy 优化排序
func (qb *QueryBuilder) OptimizedOrderBy(orderBy string, direction string) *gorm.DB {
	return qb.OptimizeQuery().Order(fmt.Sprintf("%s %s", orderBy, direction))
}

// QueryWithTimeout 带超时的查询
func (qb *QueryBuilder) QueryWithTimeout(timeout time.Duration, query func(*gorm.DB) *gorm.DB) *gorm.DB {
	ctx, cancel := context.WithTimeout(qb.ctx, timeout)
	defer cancel()

	return query(qb.db.WithContext(ctx))
}

// OptimizedCount 优化计数查询
func (qb *QueryBuilder) OptimizedCount(table string, conditions map[string]interface{}) (int64, error) {
	var count int64
	tx := qb.db.WithContext(qb.ctx).Table(table)

	for key, value := range conditions {
		tx = tx.Where(key, value)
	}

	err := tx.Count(&count).Error
	return count, err
}

// TransactionalOperation 事务操作优化
func (qb *QueryBuilder) TransactionalOperation(fn func(*gorm.DB) error) error {
	return qb.db.WithContext(qb.ctx).Transaction(fn)
}

// ExplainQuery 查询计划分析
// 注意: 此函数已废弃，因为PostgreSQL的EXPLAIN ANALYZE不支持参数占位符
// 建议直接使用: db.Raw("EXPLAIN (ANALYZE, VERBOSE) SELECT * FROM table WHERE id = $1", value)
func (qb *QueryBuilder) ExplainQuery(sql string, values ...interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 使用PostgreSQL兼容的参数占位符格式
	// 注意: 调用者需要确保SQL中使用$1, $2等占位符，而不是?
	err := qb.db.WithContext(qb.ctx).Raw("EXPLAIN (ANALYZE, VERBOSE, FORMAT JSON) "+sql, values...).Scan(&results).Error
	return results, err
}

// HealthCheckQuery 健康检查查询
func (qb *QueryBuilder) HealthCheckQuery() error {
	return qb.db.WithContext(qb.ctx).Exec("SELECT 1").Error
}

// GetQueryStats 获取查询统计信息
func (qb *QueryBuilder) GetQueryStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取连接池状态
	sqlDB, err := qb.db.DB()
	if err != nil {
		return nil, err
	}

	dbStats := sqlDB.Stats()
	stats["open_connections"] = dbStats.OpenConnections
	stats["in_use"] = dbStats.InUse
	stats["idle"] = dbStats.Idle
	stats["wait_count"] = dbStats.WaitCount
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = dbStats.MaxIdleClosed
	stats["max_idle_time_closed"] = dbStats.MaxIdleTimeClosed
	stats["max_lifetime_closed"] = dbStats.MaxLifetimeClosed

	return stats, nil
}

// Close 关闭查询构建器
func (qb *QueryBuilder) Close() error {
	sqlDB, err := qb.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}