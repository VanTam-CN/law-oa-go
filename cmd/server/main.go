package main

import (
	"context"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/router"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化Redis连接
	redisClient := initRedis(cfg)

	// 初始化Elasticsearch连接
	esClient := initElasticsearch(cfg)

	// 初始化监控（待实现）

	// 创建路由器
	routerConfig := &router.RouterConfig{
		DB:             db,
		Redis:          redisClient,
		Elasticsearch:  esClient,
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		RateLimit:      100,
		Timeout:        30 * time.Second,
	}

	app := router.NewRouter(routerConfig)

	// 启动服务器
	addr := ":" + cfg.GetPort()
	log.Printf("Starting server on %s", addr)

	if err := app.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDatabaseDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connected successfully")
	return db, nil
}

// initRedis 初始化Redis连接
func initRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("Redis connection failed: %v", err)
		return nil
	}

	log.Println("Redis connected successfully")
	return client
}

// initElasticsearch 初始化Elasticsearch连接
func initElasticsearch(cfg *config.Config) *elasticsearch.Client {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.GetElasticsearchURL()},
	}

	if cfg.Elasticsearch.Username != "" && cfg.Elasticsearch.Password != "" {
		esCfg.Username = cfg.Elasticsearch.Username
		esCfg.Password = cfg.Elasticsearch.Password
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Printf("Elasticsearch connection failed: %v", err)
		return nil
	}

	// 测试连接
	if _, err := client.Info(); err != nil {
		log.Printf("Elasticsearch info failed: %v", err)
		return nil
	}

	log.Println("Elasticsearch connected successfully")
	return client
}
