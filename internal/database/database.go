package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"law-oa-go/internal/config"
)

// Init 初始化数据库连接
func Init(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%v&loc=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
		cfg.ParseTime,
		cfg.Loc,
	)

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn), // 生产环境优化
		PrepareStmt:                              true,                                // 启用预编译语句
		SkipDefaultTransaction:                   false,                               // 保持事务安全
		DisableForeignKeyConstraintWhenMigrating: true,                                // 禁用自动外键
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层sql.DB并配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池（按照GORM最佳实践）
	sqlDB.SetMaxOpenConns(100)                 // 最大连接数
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(1 * time.Hour)    // 连接最大生命周期
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 连接最大空闲时间

	log.Println("数据库连接成功")
	return db, nil
}

// InitRedis 初始化Redis连接
func InitRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis连接成功")
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
