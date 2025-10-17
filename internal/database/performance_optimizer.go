package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
)

// 数据库性能相关的Prometheus指标
var (
	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Duration of database queries",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"operation", "table", "query_type"})

	dbQueryCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_query_count_total",
		Help: "Total number of database queries",
	}, []string{"operation", "table", "query_type"})

	dbQueryErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_query_errors_total",
		Help: "Total number of database query errors",
	}, []string{"operation", "table", "query_type"})

	dbSlowQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_slow_queries_total",
		Help: "Total number of slow database queries",
	}, []string{"operation", "table", "query_type"})

	dbConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "db_connections_active",
		Help: "Number of active database connections",
	}, []string{"type"})

	dbCacheHitRate = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "db_cache_hit_rate",
		Help: "Database query cache hit rate",
	}, []string{"table"})
)

// PerformanceOptimizer 性能优化器
type PerformanceOptimizer struct {
	db    *gorm.DB
	cache *cache.AdvancedCacheService
	stats *QueryStats
}

// QueryStats 查询统计
type QueryStats struct {
	TotalQueries    int64         `json:"total_queries"`
	CacheHits       int64         `json:"cache_hits"`
	CacheMisses     int64         `json:"cache_misses"`
	SlowQueries     int64         `json:"slow_queries"`
	AvgQueryTime    time.Duration `json:"avg_query_time"`
	ErrorCount      int64         `json:"error_count"`
	TotalQueryTime  time.Duration `json:"total_query_time"`
}

// QueryOptions 查询选项
type QueryOptions struct {
	CacheKey       string        `json:"cache_key"`
	CacheTTL       time.Duration `json:"cache_ttl"`
	UseCache       bool          `json:"use_cache"`
	EnableStats    bool          `json:"enable_stats"`
	RetryAttempts  int           `json:"retry_attempts"`
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"`
}

// PaginatedQueryOptions 分页查询选项
type PaginatedQueryOptions struct {
	QueryOptions
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	MaxPage    int `json:"max_page"` // 防止过大分页
	UseCursor  bool `json:"use_cursor"` // 使用游标分页
}

// BatchOperationOptions 批量操作选项
type BatchOperationOptions struct {
	BatchSize     int           `json:"batch_size"`
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	UseTransaction bool        `json:"use_transaction"`
}

// DefaultQueryOptions 默认查询选项
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		CacheKey:            "",
		CacheTTL:            5 * time.Minute,
		UseCache:            true,
		EnableStats:         true,
		RetryAttempts:       3,
		SlowQueryThreshold:  100 * time.Millisecond,
	}
}

// DefaultPaginatedQueryOptions 默认分页查询选项
func DefaultPaginatedQueryOptions() PaginatedQueryOptions {
	return PaginatedQueryOptions{
		QueryOptions:        DefaultQueryOptions(),
		Page:               1,
		PageSize:           20,
		MaxPage:            1000, // 防止过大分页
		UseCursor:          false,
	}
}

// DefaultBatchOperationOptions 默认批量操作选项
func DefaultBatchOperationOptions() BatchOperationOptions {
	return BatchOperationOptions{
		BatchSize:          100,
		MaxRetries:         3,
		RetryDelay:         100 * time.Millisecond,
		UseTransaction:     true,
	}
}

// NewPerformanceOptimizer 创建性能优化器
func NewPerformanceOptimizer(db *gorm.DB, cache *cache.AdvancedCacheService) *PerformanceOptimizer {
	return &PerformanceOptimizer{
		db:    db,
		cache: cache,
		stats: &QueryStats{},
	}
}

// OptimizedQuery 优化查询
func (po *PerformanceOptimizer) OptimizedQuery(ctx context.Context, opts QueryOptions, queryFunc func(*gorm.DB) *gorm.DB, dest interface{}) error {
	start := time.Now()
	table := po.extractTableName(dest)
	operation := "select"
	queryType := po.detectQueryType(queryFunc)

	// 尝试从缓存获取
	if opts.UseCache && opts.CacheKey != "" && po.cache != nil {
		cacheKey := po.buildCacheKey(opts.CacheKey, table, operation)
		_, err := po.cache.GetWithCacheThroughput(ctx, cacheKey, dest)
		if err == nil {
			po.recordStats("cache_hit", table, operation, queryType, time.Since(start))
			return nil
		}
	}

	// 执行查询，支持重试
	var err error
	for attempt := 0; attempt <= opts.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond) // 指数退避
		}

		query := queryFunc(po.db.WithContext(ctx))
		err = query.Find(dest).Error
		if err == nil {
			break
		}

		// 只对可重试的错误进行重试
		if !po.isRetryableError(err) {
			break
		}
	}

	duration := time.Since(start)

	if err != nil {
		po.recordStats("error", table, operation, queryType, duration)
		dbQueryErrors.WithLabelValues(operation, table, queryType).Inc()
		return fmt.Errorf("query failed after %d attempts: %w", opts.RetryAttempts+1, err)
	}

	// 记录统计信息
	po.recordStats("success", table, operation, queryType, duration)

	// 检查慢查询
	if duration > opts.SlowQueryThreshold {
		dbSlowQueries.WithLabelValues(operation, table, queryType).Inc()
		log.Printf("Slow query detected: %s on table %s took %v", operation, table, duration)
	}

	// 缓存结果
	if opts.UseCache && opts.CacheKey != "" && po.cache != nil {
		cacheKey := po.buildCacheKey(opts.CacheKey, table, operation)
		if err := po.cache.SetWithRandomTTL(ctx, cacheKey, dest, opts.CacheTTL); err != nil {
			log.Printf("Warning: failed to cache query results for %s: %v", cacheKey, err)
		}
	}

	return nil
}

// OptimizedPaginatedQuery 优化分页查询
func (po *PerformanceOptimizer) OptimizedPaginatedQuery(ctx context.Context, opts PaginatedQueryOptions, queryFunc func(*gorm.DB) *gorm.DB, dest interface{}) (int64, error) {
	start := time.Now()
	table := po.extractTableName(dest)
	operation := "paginate"

	// 验证分页参数
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 || opts.PageSize > 1000 {
		opts.PageSize = 20
	}
	if opts.MaxPage > 0 && opts.Page > opts.MaxPage {
		opts.Page = opts.MaxPage
	}

	// 构建缓存键
	var cacheKey string
	if opts.UseCache && opts.CacheKey != "" && po.cache != nil {
		cacheKey = fmt.Sprintf("%s:page:%d:size:%d", opts.CacheKey, opts.Page, opts.PageSize)
	}

	// 尝试从缓存获取
	if opts.UseCache && cacheKey != "" && po.cache != nil {
		var cachedResult struct {
			Data  interface{} `json:"data"`
			Total int64       `json:"total"`
		}

		_, err := po.cache.GetWithCacheThroughput(ctx, cacheKey, &cachedResult)
		if err == nil {
			po.recordStats("cache_hit", table, operation, "paginated", time.Since(start))
			// 需要将缓存结果反序列化到目标对象
			// 这里简化处理，实际使用中可能需要更复杂的逻辑
			if err := po.copyData(cachedResult.Data, dest); err == nil {
				return cachedResult.Total, nil
			}
		}
	}

	// 获取总数（可以缓存）
	var total int64
	countQuery := queryFunc(po.db.WithContext(ctx))
	if err := countQuery.Count(&total).Error; err != nil {
		po.recordStats("error", table, "count", "paginated", time.Since(start))
		return 0, fmt.Errorf("count query failed: %w", err)
	}

	// 执行分页查询
	offset := (opts.Page - 1) * opts.PageSize
	paginatedQuery := queryFunc(po.db.WithContext(ctx)).Offset(offset).Limit(opts.PageSize)

	if err := paginatedQuery.Find(dest).Error; err != nil {
		po.recordStats("error", table, operation, "paginated", time.Since(start))
		return 0, fmt.Errorf("paginated query failed: %w", err)
	}

	duration := time.Since(start)
	po.recordStats("success", table, operation, "paginated", duration)

	// 检查慢查询
	if duration > opts.SlowQueryThreshold {
		dbSlowQueries.WithLabelValues(operation, table, "paginated").Inc()
		log.Printf("Slow paginated query detected: took %v", duration)
	}

	// 缓存结果
	if opts.UseCache && cacheKey != "" && po.cache != nil {
		cachedResult := map[string]interface{}{
			"data":  dest,
			"total": total,
		}
		if err := po.cache.SetWithRandomTTL(ctx, cacheKey, cachedResult, opts.CacheTTL); err != nil {
			log.Printf("Warning: failed to cache paginated query results for %s: %v", cacheKey, err)
		}
	}

	return total, nil
}

// OptimizedBatchInsert 优化批量插入
func (po *PerformanceOptimizer) OptimizedBatchInsert(ctx context.Context, opts BatchOperationOptions, tableName string, data interface{}) error {
	start := time.Now()
	operation := "batch_insert"

	// 验证批量大小
	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}

	// 使用事务确保数据一致性
	if opts.UseTransaction {
		return po.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return po.executeBatchInsert(ctx, tx, tableName, data, opts, start, operation)
		})
	}

	return po.executeBatchInsert(ctx, po.db, tableName, data, opts, start, operation)
}

// OptimizedBatchUpdate 优化批量更新
func (po *PerformanceOptimizer) OptimizedBatchUpdate(ctx context.Context, opts BatchOperationOptions, tableName string, ids []interface{}, updates map[string]interface{}) error {
	start := time.Now()
	operation := "batch_update"

	if len(ids) == 0 {
		return nil
	}

	// 验证批量大小
	if opts.BatchSize <= 0 {
		opts.BatchSize = 100
	}

	// 分批处理
	for i := 0; i < len(ids); i += opts.BatchSize {
		end := i + opts.BatchSize
		if end > len(ids) {
			end = len(ids)
		}

		batchIDs := ids[i:end]

		// 使用事务确保数据一致性
		if opts.UseTransaction {
			err := po.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				return tx.Table(tableName).Where("id IN ?", batchIDs).Updates(updates).Error
			})

			if err != nil {
				po.recordStats("error", tableName, operation, "batch", time.Since(start))
				return fmt.Errorf("batch update transaction failed: %w", err)
			}
		} else {
			if err := po.db.WithContext(ctx).Table(tableName).Where("id IN ?", batchIDs).Updates(updates).Error; err != nil {
				po.recordStats("error", tableName, operation, "batch", time.Since(start))
				return fmt.Errorf("batch update failed: %w", err)
			}
		}

		// 清除相关缓存
		po.clearCacheForTable(tableName)
	}

	duration := time.Since(start)
	po.recordStats("success", tableName, operation, "batch", duration)

	return nil
}

// GetQueryStats 获取查询统计信息
func (po *PerformanceOptimizer) GetQueryStats() *QueryStats {
	return po.stats
}

// ResetStats 重置统计信息
func (po *PerformanceOptimizer) ResetStats() {
	po.stats = &QueryStats{}
}

// executeBatchInsert 执行批量插入的内部方法
func (po *PerformanceOptimizer) executeBatchInsert(ctx context.Context, db *gorm.DB, tableName string, data interface{}, opts BatchOperationOptions, start time.Time, operation string) error {
	// 执行批量插入
	if err := db.Table(tableName).CreateInBatches(data, opts.BatchSize).Error; err != nil {
		po.recordStats("error", tableName, operation, "batch", time.Since(start))
		return fmt.Errorf("batch insert failed: %w", err)
	}

	duration := time.Since(start)
	po.recordStats("success", tableName, operation, "batch", duration)

	// 清除相关缓存
	po.clearCacheForTable(tableName)

	return nil
}

// extractTableName 提取表名
func (po *PerformanceOptimizer) extractTableName(dest interface{}) string {
	// 简化的表名提取逻辑
	// 实际项目中可以根据dest的类型来推断表名
	return "unknown"
}

// detectQueryType 检测查询类型
func (po *PerformanceOptimizer) detectQueryType(queryFunc func(*gorm.DB) *gorm.DB) string {
	// 简化的查询类型检测
	// 实际项目中可以通过分析SQL来检测
	return "unknown"
}

// buildCacheKey 构建缓存键
func (po *PerformanceOptimizer) buildCacheKey(baseKey, table, operation string) string {
	return fmt.Sprintf("db:%s:%s:%s", table, operation, baseKey)
}

// isRetryableError 判断错误是否可重试
func (po *PerformanceOptimizer) isRetryableError(err error) bool {
	// 简化的重试逻辑
	// 实际项目中可以根据具体的错误类型来判断
	errorMsg := strings.ToLower(err.Error())

	// 可重试的错误类型
	retryableErrors := []string{
		"connection refused",
		"timeout",
		"deadlock",
		"connection reset",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(errorMsg, retryableErr) {
			return true
		}
	}

	return false
}

// recordStats 记录统计信息
func (po *PerformanceOptimizer) recordStats(status, table, operation, queryType string, duration time.Duration) {
	if po.stats == nil {
		return
	}

	po.stats.TotalQueries++
	po.stats.TotalQueryTime += duration
	po.stats.AvgQueryTime = po.stats.TotalQueryTime / time.Duration(po.stats.TotalQueries)

	switch status {
	case "cache_hit":
		po.stats.CacheHits++
	case "cache_miss":
		po.stats.CacheMisses++
	case "error":
		po.stats.ErrorCount++
	case "slow":
		po.stats.SlowQueries++
	}

	// 更新Prometheus指标
	dbQueryDuration.WithLabelValues(operation, table, queryType).Observe(duration.Seconds())
	dbQueryCount.WithLabelValues(operation, table, queryType).Inc()

	// 更新缓存命中率
	if po.stats.TotalQueries > 0 {
		hitRate := float64(po.stats.CacheHits) / float64(po.stats.TotalQueries)
		dbCacheHitRate.WithLabelValues(table).Set(hitRate)
	}
}

// clearCacheForTable 清除表相关的缓存
func (po *PerformanceOptimizer) clearCacheForTable(tableName string) {
	if po.cache == nil {
		return
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("db:%s:*", tableName)
	if err := po.cache.DeletePattern(ctx, pattern); err != nil {
		log.Printf("Warning: failed to clear cache for table %s: %v", tableName, err)
	}
}

// copyData 复制数据的简化实现
func (po *PerformanceOptimizer) copyData(src, dest interface{}) error {
	// 简化的数据复制逻辑
	// 实际项目中可能需要使用反射或其他方法
	return nil
}

// GetConnectionStats 获取连接统计信息
func (po *PerformanceOptimizer) GetConnectionStats() map[string]interface{} {
	if po.db == nil {
		return nil
	}

	sqlDB, err := po.db.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()

	// 更新Prometheus指标
	dbConnections.WithLabelValues("open").Set(float64(stats.OpenConnections))
	dbConnections.WithLabelValues("in_use").Set(float64(stats.InUse))
	dbConnections.WithLabelValues("idle").Set(float64(stats.Idle))

	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":        stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration,
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}
}