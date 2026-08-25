package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/config"
)

// Init 初始化数据库连接（向后兼容）
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// 为了向后兼容，创建一个默认的应用配置
	appConfig := &config.Config{
		Environment: "development",
		Database:    cfg,
	}
	return InitWithConfig(appConfig)
}

// InitWithConfig 使用完整配置初始化数据库连接（推荐使用）
func InitWithConfig(appConfig *config.Config) (*gorm.DB, error) {
	var dsn string
	var db *gorm.DB
	var err error

	// 配置GORM - 基于最新最佳实践优化
	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn), // 生产环境优化
		PrepareStmt:                              true,                                // 启用预编译语句缓存
		SkipDefaultTransaction:                   true,                                // 禁用默认事务提升性能
		DisableForeignKeyConstraintWhenMigrating: true,                                // 禁用自动外键
		// 新增性能优化配置
		AllowGlobalUpdate:    false, // 禁止全局更新提升安全性
		DisableAutomaticPing: true,  // 禁用自动ping减少开销
	}

	// 根据数据库类型构建DSN和连接
	driver := strings.ToLower(strings.TrimSpace(appConfig.Database.Driver))
	if driver == "postgres" || driver == "postgresql" {
		// PostgreSQL DSN
		sslMode := appConfig.Database.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			appConfig.Database.Host, appConfig.Database.Port, appConfig.Database.Username, appConfig.Database.Password, appConfig.Database.Database, sslMode, appConfig.GetDatabaseTimeZone())

		// 连接PostgreSQL
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		log.Println("使用PostgreSQL数据库")
	} else {
		// MySQL 使用结构化驱动配置，避免凭据经 DSN 字符串再解析
		db, err = openMySQLGORM(appConfig.Database, gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
		}
		log.Println("使用MySQL数据库")
	}

	// 获取底层sql.DB并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池（基于最新GORM性能最佳实践）
	optimizeConnectionPool(sqlDB, appConfig)

	log.Printf("数据库连接成功 - 类型: %s, 环境: %s", driver, appConfig.Environment)
	return db, nil
}

// InitRedis 初始化Redis连接 - 性能优化版本
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, fmt.Errorf("redis is disabled: REDIS_HOST is empty")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,

		// 连接池优化配置
		PoolSize:     100,             // 连接池大小
		MinIdleConns: 10,              // 最小空闲连接
		MaxRetries:   3,               // 最大重试次数
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 获取连接超时

		// 性能优化配置（使用Redis v9兼容的选项）
		ConnMaxIdleTime: 5 * time.Minute, // 空闲超时
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis连接成功 - 连接池大小: %d, 最小空闲: %d", rdb.Options().PoolSize, rdb.Options().MinIdleConns)
	return rdb, nil
}

// Health 健康检查所有组件
func Health() map[string]interface{} {
	status := make(map[string]interface{})

	// 检查数据库
	if OptimizedDB != nil {
		if err := OptimizedDB.Health(); err != nil {
			status["database"] = map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			}
		} else {
			status["database"] = map[string]interface{}{
				"status":  "healthy",
				"message": "Database connection is healthy",
			}
		}
	} else {
		status["database"] = map[string]interface{}{
			"status":  "error",
			"message": "Database not initialized",
		}
	}

	// 检查Redis
	if CacheService != nil {
		if CacheService != nil && CacheService.Ping() {
			status["redis"] = map[string]interface{}{
				"status":  "healthy",
				"message": "Redis connection is healthy",
			}
		} else {
			status["redis"] = map[string]interface{}{
				"status":  "error",
				"message": "Redis connection failed",
			}
		}
	} else {
		status["redis"] = map[string]interface{}{
			"status":  "disabled",
			"message": "Redis is disabled",
		}
	}

	return status
}

// Close 关闭所有连接
func Close() error {
	var errs []error

	// 关闭数据库连接
	if OptimizedDB != nil {
		if err := OptimizedDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	// 关闭Redis连接
	if CacheService != nil {
		if client := CacheService.GetClient(); client != nil {
			if err := client.Close(); err != nil {
				errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close connections: %v", errs)
	}

	return nil
}

// GetRedis 获取Redis客户端
func GetRedis() interface{} {
	if CacheService != nil {
		return CacheService.GetClient()
	}
	return nil
}

// optimizeConnectionPool 基于最新GORM性能最佳实践优化数据库连接池
func optimizeConnectionPool(sqlDB *sql.DB, appConfig *config.Config) {
	// 获取性能配置
	perfConfig := appConfig.GetDatabasePerformanceConfig()

	// 应用性能配置
	if perfConfig.EnablePerformance {
		sqlDB.SetMaxOpenConns(perfConfig.MaxOpenConns)
		sqlDB.SetMaxIdleConns(perfConfig.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(perfConfig.ConnMaxLifetime)
		sqlDB.SetConnMaxIdleTime(perfConfig.ConnMaxIdleTime)
	} else {
		// 默认配置（向后兼容）
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}

	// 启用数据库连接池统计和监控
	go monitorConnectionPool(sqlDB, appConfig.Environment, appConfig.Database.Driver)

	log.Printf("数据库连接池优化完成 - 环境: %s, 数据库: %s, 最大连接数: %d, 空闲连接数: %d, 性能优化: %v",
		appConfig.Environment, appConfig.Database.Driver,
		sqlDB.Stats().MaxOpenConnections, sqlDB.Stats().Idle, perfConfig.EnablePerformance)
}

// monitorConnectionPool 监控连接池状态并记录性能指标
func monitorConnectionPool(sqlDB *sql.DB, env, dbType string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := sqlDB.Stats()

		// 记录连接池指标
		log.Printf("[DB-POOL-STATS] 环境: %s | 数据库: %s | 打开连接: %d/%d | 空闲连接: %d | 等待连接: %d | 总计等待: %dms",
			env, dbType,
			stats.OpenConnections,
			stats.MaxOpenConnections,
			stats.Idle,
			stats.WaitCount,
			stats.WaitDuration.Milliseconds(),
		)

		// 检查连接池健康状态
		if stats.WaitCount > 0 && stats.WaitDuration > 5*time.Second {
			log.Printf("[WARNING] 数据库连接池出现瓶颈，等待时间过长: %v", stats.WaitDuration)
		}

		if stats.OpenConnections >= stats.MaxOpenConnections*90/100 {
			log.Printf("[WARNING] 数据库连接池使用率过高: %d/%d (%.1f%%)",
				stats.OpenConnections, stats.MaxOpenConnections,
				float64(stats.OpenConnections)/float64(stats.MaxOpenConnections)*100)
		}
	}
}
