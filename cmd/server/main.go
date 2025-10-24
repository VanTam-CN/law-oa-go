package main

import (
	"context"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/middleware"
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

	// 初始化JWT
	middleware.InitJWT(cfg)

	// 初始化监控（待实现）

	// 创建路由器
	app := gin.Default()

	// 初始化路由系统
	router.Init(app, db, redisClient, esClient)

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

	// 优化GORM配置
	gormConfig := &gorm.Config{
		PrepareStmt:            true,  // 预编译语句缓存
		SkipDefaultTransaction: true,  // 跳过默认事务
		DisableForeignKeyConstraintWhenMigrating: true, // 禁用外键约束检查（提高性能）
	}

	var db *gorm.DB
	var err error

	// 根据数据库类型选择驱动
	if cfg.Database.Driver == "postgres" {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			return nil, err
		}
		log.Printf("Connected to PostgreSQL database: %s", cfg.Database.Database)
	} else {
		db, err = gorm.Open(mysql.Open(dsn), gormConfig)
		if err != nil {
			return nil, err
		}
		log.Printf("Connected to MySQL database: %s", cfg.Database.Database)
	}

	// 优化连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 根据环境动态调整连接池大小
	maxOpenConns := 50
	maxIdleConns := 10
	connMaxLifetime := 30 * time.Minute
	connMaxIdleTime := 5 * time.Minute

	if cfg.IsProduction() {
		maxOpenConns = 100
		maxIdleConns = 20
		connMaxLifetime = time.Hour
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Printf("Database connected successfully (Pool: %d idle/%d max, Lifetime: %v)",
		maxIdleConns, maxOpenConns, connMaxLifetime)
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
