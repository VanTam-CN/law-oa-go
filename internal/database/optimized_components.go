package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
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
	if err := qo.cache.Get(cacheKey, dest); err == nil {
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
	if err := qo.cache.Set(cacheKey, dest, ttl); err != nil {
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
		if err := qo.cache.ClearPattern(cachePattern); err != nil {
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
		if err := qo.cache.ClearPattern(cachePattern); err != nil {
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
	OptimizedDB              *OptimizedDatabase
	CacheService             *cache.CacheService
	AdvancedCacheService     *cache.AdvancedCacheService
	QueryOptimizerInst       *QueryOptimizer
	PerformanceOptimizerInst *PerformanceOptimizer
	ElasticsearchClient      *elasticsearch.Client
)

// InitOptimizedComponents 初始化所有优化组件
func InitOptimizedComponents(cfg *config.Config) error {
	var err error

	// 初始化优化数据库连接
	OptimizedDB, err = NewOptimizedDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize optimized database: %w", err)
	}

	// 跳过自动迁移，使用现有数据库结构
	// err = OptimizedDB.DB.AutoMigrate(
	// 	&models.User{},
	// 	&models.Client{},
	// 	&models.Case{},
	// 	&models.Document{},
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to auto migrate database: %w", err)
	// }

	// Production must never create demo users or demo matters implicitly. Seed
	// data is reserved for local development and controlled trial environments.
	if !cfg.IsProduction() {
		err = initSeedData(OptimizedDB.DB)
		if err != nil {
			return fmt.Errorf("failed to initialize seed data: %w", err)
		}
	} else {
		log.Println("生产环境跳过演示种子数据初始化")
	}

	// 初始化Redis缓存服务
	redisClient, err := InitRedis(cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// 初始化缓存服务
	CacheService = cache.NewCacheService(redisClient, "lawoa")

	// 初始化高级缓存服务
	AdvancedCacheService = cache.NewAdvancedCacheService(redisClient, "lawoa_advanced", 30*time.Minute)

	// 初始化查询优化器
	QueryOptimizerInst = NewQueryOptimizer(OptimizedDB.DB, CacheService)

	// 初始化性能优化器
	PerformanceOptimizerInst = NewPerformanceOptimizer(OptimizedDB.DB, AdvancedCacheService)

	// 初始化Elasticsearch（可选）
	ElasticsearchClient, err = InitElasticsearch(cfg.Elasticsearch)
	if err != nil {
		fmt.Printf("Elasticsearch连接失败，跳过初始化: %v\n", err)
		// 不返回错误，Elasticsearch是可选的
	} else {
		// 初始化Elasticsearch优化器
		esConfig := DefaultESConfig()
		esConfig.Addresses = []string{elasticsearchAddress(cfg.Elasticsearch.Host, cfg.Elasticsearch.Port)}
		esConfig.Username = cfg.Elasticsearch.Username
		esConfig.Password = cfg.Elasticsearch.Password

		_, err = NewElasticsearchOptimizer(esConfig)
		if err != nil {
			fmt.Printf("Elasticsearch优化器初始化失败: %v\n", err)
			// 不返回错误，优化器是可选的
		}
	}

	log.Println("所有优化组件初始化完成 - 包含高级缓存和性能优化器")
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

// GetAdvancedCacheService 获取高级缓存服务实例
func GetAdvancedCacheService() *cache.AdvancedCacheService {
	return AdvancedCacheService
}

// GetPerformanceOptimizer 获取性能优化器实例
func GetPerformanceOptimizer() *PerformanceOptimizer {
	return PerformanceOptimizerInst
}

// initSeedData 初始化种子数据
func initSeedData(db *gorm.DB) error {
	// 检查是否已有管理员用户
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	// 如果已有数据，则跳过初始化
	if userCount > 0 {
		return nil
	}

	// 生成密码哈希
	passwordHash := "$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG" // password

	// 创建管理员用户
	adminUser := &models.User{
		Name:     "admin",
		Password: passwordHash,
		Email:    "admin@example.com",
		Role:     "admin",
		Status:   "active",
	}
	if err := db.Create(adminUser).Error; err != nil {
		return err
	}

	// 创建测试客户
	clients := []*models.Client{
		{Name: "张三", Phone: "13800138001", Email: "zhangsan@example.com", Address: "北京市朝阳区", Status: "active"},
		{Name: "李四", Phone: "13800138002", Email: "lisi@example.com", Address: "北京市海淀区", Status: "active"},
		{Name: "王五公司", Phone: "13800138003", Email: "wangwu@example.com", Company: "王五科技有限公司", Address: "北京市西城区", Status: "active"},
	}
	for _, client := range clients {
		if err := db.Create(client).Error; err != nil {
			return err
		}
	}

	// 创建测试案件 (需要先创建一个律师用户)
	lawyerUser := &models.User{
		Name:     "lawyer1",
		Password: passwordHash,
		Email:    "lawyer1@example.com",
		Role:     "lawyer",
		Status:   "active",
	}

	// 检查用户是否已存在，如果不存在则创建
	var existingUser models.User
	if err := db.Where("email = ?", lawyerUser.Email).First(&existingUser).Error; err != nil {
		// 用户不存在，创建新用户
		if err := db.Create(lawyerUser).Error; err != nil {
			log.Printf("创建律师用户失败: %v", err)
			return err
		}
	} else {
		// 用户已存在，使用现有用户
		lawyerUser = &existingUser
		log.Println("律师用户已存在，跳过创建")
	}

	cases := []*models.Case{
		{Title: "张三借款合同纠纷", CaseType: "借款合同", ClientID: clients[0].ID, LawyerID: lawyerUser.ID, Status: "active", Description: "张三与李四之间的借款合同纠纷案件，涉及金额50万元。"},
		{Title: "王五公司劳动合同纠纷", CaseType: "劳动合同", ClientID: clients[1].ID, LawyerID: lawyerUser.ID, Status: "pending", Description: "王五公司与员工之间的劳动合同纠纷，涉及经济补偿金。"},
		{Title: "赵六房屋买卖合同纠纷", CaseType: "房屋买卖", ClientID: clients[2].ID, LawyerID: lawyerUser.ID, Status: "active", Description: "赵六与开发商之间的房屋买卖合同纠纷，涉及房屋质量问题。"},
	}
	for _, caseItem := range cases {
		if err := db.Create(caseItem).Error; err != nil {
			return err
		}
	}

	log.Println("种子数据初始化完成")
	return nil
}
