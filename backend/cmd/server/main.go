package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/router"
	"law-oa-go/internal/security"
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

	// 初始化JWT
	middleware.InitJWT(cfg)

	// 初始化监控（待实现）

	// 自动迁移财务模型
	if err := db.AutoMigrate(
		&models.Contract{},
		&models.PaymentMilestone{},
		&models.Invoice{},
		&models.Payment{},
		&models.BadDebtRecord{},
		&models.CommissionRecord{},
		&models.CommissionRule{},
		&models.FeeTemplate{},
	); err != nil {
		log.Printf("⚠️ 财务模型自动迁移失败: %v", err)
	} else {
		log.Println("✅ 财务模型自动迁移成功")
	}
	// 自动迁移Sprint 1模型 (Entity/冲突检测/隔离墙)
	if err := db.AutoMigrate(
		&models.Entity{},
		&models.EntityRelation{},
		&models.EntityNameHistory{},
		&models.CaseParty{},
		&models.ConflictCheck{},
		&models.ConflictDetail{},
		&models.CaseEthicalWallWhitelist{},
		&models.EthicalWallAccessLog{},
	); err != nil {
		log.Printf("⚠️ Sprint1模型自动迁移失败: %v", err)
	} else {
		log.Println("✅ Sprint1模型自动迁移成功")
	}

	app := gin.Default()

	// 添加安全头部中间件
	app.Use(middleware.SecurityHeaders())

	// 添加CORS中间件（白名单模式，从 CORS_ALLOWED_ORIGINS 环境变量读取）
	app.Use(middleware.CORS())

	// 添加速率限制中间件（100 req/min per IP）
	app.Use(security.RateLimiterMiddleware())

	// 初始化路由系统
	router.Init(app, db, redisClient)

	// 启动服务器
	addr := ":" + cfg.GetPort()
	log.Printf("Starting server on %s", addr)

	if err := app.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	return database.InitWithConfig(cfg)
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
