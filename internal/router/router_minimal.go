package router

import (
	"log"

	"law-oa-go/internal/auth"
	"law-oa-go/internal/config"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	rdb "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// InitMinimal 初始化最小化路由（仅包含法条搜索基本功能）
func InitMinimal(app *gin.Engine, db *gorm.DB, redisClient *rdb.Client) {
	log.Println("初始化最小化路由系统...")

	// 初始化法条相关Repository
	legalStatuteRepo := repositories.NewLegalStatuteRepository(db)
	legalCategoryRepo := repositories.NewLegalCategoryRepository(db)
	legalTagRepo := repositories.NewLegalTagRepository(db)

	// 初始化法条服务
	legalStatuteService := services.NewLegalStatuteService(db, legalStatuteRepo, legalCategoryRepo, legalTagRepo, nil)

	// 初始化法条处理器
	legalStatuteHandler := handlers.NewLegalStatuteHandler(legalStatuteService)

	// 公开路由 - 认证和法条搜索
	public := app.Group("/api/v1")
	{
		// 认证相关路由
		authRoutes := public.Group("/auth")
		{
			// 创建用户服务用于认证
			userRepo := repositories.NewUserRepository(db)
			userService := services.NewUserService(userRepo)
			cfg, err := config.Load()
			if err != nil {
				log.Printf("加载配置失败: %v", err)
				return
			}
			// Minimal routes still require durable PostgreSQL sessions so a
			// login token can be revoked by UUID.
			tokenManager := auth.NewTokenManager(cfg, redisClient, nil, db)
			authHandler := handlers.NewAuthHandler(userService, nil, tokenManager)

			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", auth.RefreshTokenMiddleware(tokenManager))
			authRoutes.POST("/register", authHandler.Register)
		}

		// 法条搜索（公开，方便测试）
		legal := public.Group("/legal")
		{
			legal.GET("/statutes/search", legalStatuteHandler.SearchStatutes)
			legal.GET("/statutes/:id", legalStatuteHandler.GetStatuteByID)
			legal.GET("/categories", legalStatuteHandler.GetCategories)
			legal.GET("/tags", legalStatuteHandler.GetTags)
			legal.GET("/search/suggestions", legalStatuteHandler.GetSearchSuggestions)
		}
	}

	// 需要认证的法条路由
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		legal := protected.Group("/legal")
		{
			legal.POST("/statutes", legalStatuteHandler.CreateStatute)
			legal.PUT("/statutes/:id", legalStatuteHandler.UpdateStatute)
			legal.DELETE("/statutes/:id", legalStatuteHandler.DeleteStatute)
		}
	}

	// 健康检查（避免冲突，使用不同路径）
	app.GET("/health-minimal", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "法条搜索系统运行正常",
		})
	})

	log.Println("最小化路由系统初始化完成")
}
