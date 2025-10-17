package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/config"
)

// Init 初始化数据库连接
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var dsn string
	var db *gorm.DB
	var err error

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn), // 生产环境优化
		PrepareStmt:                              true,                                // 启用预编译语句
		SkipDefaultTransaction:                   false,                               // 保持事务安全
		DisableForeignKeyConstraintWhenMigrating: true,                                // 禁用自动外键
	}

	// 根据数据库类型构建DSN和连接
	if cfg.Driver == "postgres" {
		// PostgreSQL DSN
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)

		// 连接PostgreSQL
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		log.Println("使用PostgreSQL数据库")
	} else {
		// MySQL DSN (向后兼容)
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%v&loc=%s&tls=skip-verify",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset, cfg.ParseTime, cfg.Loc)

		// 连接MySQL
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
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

	// 配置连接池（根据环境和负载优化）
	configureConnectionPool(sqlDB)

	log.Printf("数据库连接成功 - 类型: %s", cfg.Driver)
	return db, nil
}

// InitRedis 初始化Redis连接 - 性能优化版本
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,

		// 连接池优化配置
		PoolSize:        100,          // 连接池大小
		MinIdleConns:    10,           // 最小空闲连接
		MaxRetries:      3,            // 最大重试次数
		DialTimeout:     5 * time.Second,  // 连接超时
		ReadTimeout:     3 * time.Second,  // 读取超时
		WriteTimeout:    3 * time.Second,  // 写入超时
		PoolTimeout:     4 * time.Second,  // 获取连接超时

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

// InitElasticsearch 初始化Elasticsearch连接
func InitElasticsearch(cfg config.ElasticsearchConfig) (*elasticsearch.Client, error) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)},
		Username:  cfg.Username,
		Password:  cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}

	// 测试连接
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch returned error: %s", res.Status())
	}

	log.Println("Elasticsearch连接成功")
	return es, nil
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
			"status":  "error",
			"message": "Redis not initialized",
		}
	}

	// 检查Elasticsearch
	if ElasticsearchClient != nil {
		res, err := ElasticsearchClient.Info()
		if err != nil {
			status["elasticsearch"] = map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			}
		} else {
			defer res.Body.Close()
			if res.IsError() {
				status["elasticsearch"] = map[string]interface{}{
					"status":  "error",
					"message": fmt.Sprintf("Elasticsearch returned error: %s", res.Status()),
				}
			} else {
				status["elasticsearch"] = map[string]interface{}{
					"status":  "healthy",
					"message": "Elasticsearch connection is healthy",
				}
			}
		}
	} else {
		status["elasticsearch"] = map[string]interface{}{
			"status":  "error",
			"message": "Elasticsearch not initialized",
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

// configureConnectionPool 根据环境配置数据库连接池
func configureConnectionPool(sqlDB *sql.DB) {
	// 获取环境变量，默认为development
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	switch env {
	case "production":
		// 生产环境配置：保守的连接数，较短的连接生命周期
		sqlDB.SetMaxOpenConns(25)                  // 最大连接数，避免数据库压力
		sqlDB.SetMaxIdleConns(25)                  // 空闲连接数与最大连接数相等
		sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期，避免长时间占用
		sqlDB.SetConnMaxIdleTime(1 * time.Minute)  // 空闲连接超时，快速释放资源

	case "staging":
		// 预发布环境配置：适中的连接数
		sqlDB.SetMaxOpenConns(20)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
		sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	case "testing":
		// 测试环境配置：较少的连接数
		sqlDB.SetMaxOpenConns(5)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	default: // development
		// 开发环境配置：平衡性能和资源使用
		sqlDB.SetMaxOpenConns(15)                  // 适中的连接数
		sqlDB.SetMaxIdleConns(15)                  // 空闲连接数与最大连接数相等
		sqlDB.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // 空闲连接超时
	}

	log.Printf("数据库连接池配置完成 - 环境: %s, 最大连接数: %d, 最大空闲连接数: %d",
		env, sqlDB.Stats().MaxOpenConnections, sqlDB.Stats().Idle)
}
