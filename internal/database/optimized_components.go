package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
)

// 数据库查询相关的Prometheus指标
var (
	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Duration of database queries",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "table"})

	dbQueryCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_query_count_total",
		Help: "Total number of database queries",
	}, []string{"operation", "table"})

	dbQueryErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_query_errors_total",
		Help: "Total number of database query errors",
	}, []string{"operation", "table"})

	dbSlowQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_slow_queries_total",
		Help: "Total number of slow database queries (>100ms)",
	}, []string{"operation", "table"})
)

// QueryOptimizer 查询优化器
type QueryOptimizer struct {
	db    *gorm.DB
	cache *cache.CacheService
}

// NewQueryOptimizer 创建新的查询优化器
func NewQueryOptimizer(db *gorm.DB, cache *cache.CacheService) *QueryOptimizer {
	return &QueryOptimizer{
		db:    db,
		cache: cache,
	}
}

// CachedQuery 缓存查询
func (qo *QueryOptimizer) CachedQuery(ctx context.Context, cacheKey string, ttl time.Duration, queryFunc func(*gorm.DB) *gorm.DB, dest interface{}) error {
	start := time.Now()
	table := getTableNameFromDest(dest)
	operation := "select"

	// 尝试从缓存获取
	if err := qo.cache.Get(ctx, cacheKey, dest); err == nil {
		dbQueryCount.WithLabelValues(operation+"_cache", table).Inc()
		return nil
	}

	// 执行查询
	query := queryFunc(qo.db)
	if err := query.Find(dest).Error; err != nil {
		dbQueryErrors.WithLabelValues(operation, table).Inc()
		return fmt.Errorf("query failed: %w", err)
	}

	// 记录查询耗时
	duration := time.Since(start)
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	dbQueryCount.WithLabelValues(operation, table).Inc()

	// 检查慢查询
	if duration > 100*time.Millisecond {
		dbSlowQueries.WithLabelValues(operation, table).Inc()
	}

	// 缓存结果
	if err := qo.cache.Set(ctx, cacheKey, dest, ttl); err != nil {
		fmt.Printf("Warning: failed to cache query results for %s: %v\n", cacheKey, err)
	}

	return nil
}

// PaginatedQuery 分页查询优化
func (qo *QueryOptimizer) PaginatedQuery(ctx context.Context, queryFunc func(*gorm.DB) *gorm.DB, dest interface{}, page, pageSize int) (int64, error) {
	start := time.Now()
	table := getTableNameFromDest(dest)
	operation := "paginate"

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取总数
	var total int64
	countQuery := queryFunc(qo.db)
	if err := countQuery.Count(&total).Error; err != nil {
		dbQueryErrors.WithLabelValues(operation+"_count", table).Inc()
		return 0, fmt.Errorf("count query failed: %w", err)
	}

	// 执行分页查询
	paginatedQuery := queryFunc(qo.db).Offset(offset).Limit(pageSize)
	if err := paginatedQuery.Find(dest).Error; err != nil {
		dbQueryErrors.WithLabelValues(operation+"_data", table).Inc()
		return 0, fmt.Errorf("paginated query failed: %w", err)
	}

	// 记录查询耗时
	duration := time.Since(start)
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	dbQueryCount.WithLabelValues(operation, table).Inc()

	// 检查慢查询
	if duration > 100*time.Millisecond {
		dbSlowQueries.WithLabelValues(operation, table).Inc()
	}

	return total, nil
}

// BatchInsert 批量插入优化
func (qo *QueryOptimizer) BatchInsert(ctx context.Context, tableName string, data interface{}, batchSize int) error {
	start := time.Now()
	operation := "batch_insert"

	// 使用事务确保数据一致性
	return qo.db.Transaction(func(tx *gorm.DB) error {
		// 批量插入
		if err := tx.Table(tableName).CreateInBatches(data, batchSize).Error; err != nil {
			dbQueryErrors.WithLabelValues(operation, tableName).Inc()
			return fmt.Errorf("batch insert failed: %w", err)
		}

		// 记录查询耗时
		duration := time.Since(start)
		dbQueryDuration.WithLabelValues(operation, tableName).Observe(duration.Seconds())
		dbQueryCount.WithLabelValues(operation, tableName).Inc()

		// 检查慢查询
		if duration > 100*time.Millisecond {
			dbSlowQueries.WithLabelValues(operation, tableName).Inc()
		}

		return nil
	})
}

// OptimizedUpdate 优化更新操作
func (qo *QueryOptimizer) OptimizedUpdate(ctx context.Context, tableName string, id interface{}, updates map[string]interface{}) error {
	start := time.Now()
	operation := "update"

	// 执行更新
	if err := qo.db.Table(tableName).Where("id = ?", id).Updates(updates).Error; err != nil {
		dbQueryErrors.WithLabelValues(operation, tableName).Inc()
		return fmt.Errorf("update failed: %w", err)
	}

	// 记录查询耗时
	duration := time.Since(start)
	dbQueryDuration.WithLabelValues(operation, tableName).Observe(duration.Seconds())
	dbQueryCount.WithLabelValues(operation, tableName).Inc()

	// 检查慢查询
	if duration > 100*time.Millisecond {
		dbSlowQueries.WithLabelValues(operation, tableName).Inc()
	}

	// 清除相关缓存
	if qo.cache != nil {
		cachePattern := fmt.Sprintf("lawoa:%s:*", tableName)
		if err := qo.cache.ClearPattern(ctx, cachePattern); err != nil {
			fmt.Printf("Warning: failed to clear cache for table %s: %v\n", tableName, err)
		}
	}

	return nil
}

// OptimizedDelete 优化删除操作
func (qo *QueryOptimizer) OptimizedDelete(ctx context.Context, tableName string, id interface{}) error {
	start := time.Now()
	operation := "delete"

	// 执行删除
	if err := qo.db.Table(tableName).Where("id = ?", id).Delete(nil).Error; err != nil {
		dbQueryErrors.WithLabelValues(operation, tableName).Inc()
		return fmt.Errorf("delete failed: %w", err)
	}

	// 记录查询耗时
	duration := time.Since(start)
	dbQueryDuration.WithLabelValues(operation, tableName).Observe(duration.Seconds())
	dbQueryCount.WithLabelValues(operation, tableName).Inc()

	// 检查慢查询
	if duration > 100*time.Millisecond {
		dbSlowQueries.WithLabelValues(operation, tableName).Inc()
	}

	// 清除相关缓存
	if qo.cache != nil {
		cachePattern := fmt.Sprintf("lawoa:%s:*", tableName)
		if err := qo.cache.ClearPattern(ctx, cachePattern); err != nil {
			fmt.Printf("Warning: failed to clear cache for table %s: %v\n", tableName, err)
		}
	}

	return nil
}

// 辅助函数
func getTableNameFromDest(dest interface{}) string {
	// 简单的表名提取逻辑
	// 实际项目中可以根据需求实现更复杂的逻辑
	return "unknown"
}

// 全局变量存储优化后的组件
var (
	OptimizedDB         *OptimizedDatabase
	CacheService        *cache.CacheService
	QueryOptimizerInst  *QueryOptimizer
	ElasticsearchClient *elasticsearch.Client
)

// InitOptimizedComponents 初始化所有优化组件
func InitOptimizedComponents(cfg *config.Config) error {
	var err error

	// 初始化优化数据库连接
	OptimizedDB, err = NewOptimizedDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize optimized database: %w", err)
	}

	// 初始化Redis缓存服务
	redisClient, err := InitRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// 初始化缓存服务
	CacheService = cache.NewCacheService(redisClient, "lawoa")

	// 初始化查询优化器
	QueryOptimizerInst = NewQueryOptimizer(OptimizedDB.DB, CacheService)

	// 初始化Elasticsearch（可选）
	ElasticsearchClient, err = InitElasticsearch(cfg.Elasticsearch)
	if err != nil {
		fmt.Printf("Elasticsearch连接失败，跳过初始化: %v\n", err)
		// 不返回错误，Elasticsearch是可选的
	}

	log.Println("所有优化组件初始化完成")
	return nil
}

// GetOptimizedDB 获取优化后的数据库实例
func GetOptimizedDB() *OptimizedDatabase {
	return OptimizedDB
}

// GetCacheService 获取缓存服务实例
func GetCacheService() *cache.CacheService {
	return CacheService
}

// GetQueryOptimizer 获取查询优化器实例
func GetQueryOptimizer() *QueryOptimizer {
	return QueryOptimizerInst
}

// GetElasticsearchClient 获取Elasticsearch客户端实例
func GetElasticsearchClient() *elasticsearch.Client {
	return ElasticsearchClient
}
