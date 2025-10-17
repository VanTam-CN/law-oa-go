package router

import (
	"log"

	"law-oa-go/internal/handlers"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	rdb "github.com/redis/go-redis/v9"
	esv8 "github.com/elastic/go-elasticsearch/v8"
	"gorm.io/gorm"
)

// InitMinimal 初始化最小化路由（仅包含法条搜索基本功能）
func InitMinimal(app *gin.Engine, db *gorm.DB, redisClient *rdb.Client, esClient interface{}) {
	log.Println("初始化最小化路由系统...")

	// 初始化法条相关Repository
	legalStatuteRepo := repositories.NewLegalStatuteRepository(db)
	legalCategoryRepo := repositories.NewLegalCategoryRepository(db)
	legalTagRepo := repositories.NewLegalTagRepository(db)
	var legalEsRepo repositories.ElasticsearchStatuteRepository
	// 类型断言检查Elasticsearch客户端
	if esClient != nil {
		if client, ok := esClient.(*esv8.Client); ok {
			legalEsRepo = repositories.NewElasticsearchStatuteRepository(client)
		} else {
			log.Printf("Elasticsearch客户端类型不匹配，跳过ES功能")
		}
	}

	// 初始化法条服务
	legalStatuteService := services.NewLegalStatuteService(db, legalStatuteRepo, legalCategoryRepo, legalTagRepo, legalEsRepo)

	// 初始化法条处理器
	legalStatuteHandler := handlers.NewLegalStatuteHandler(legalStatuteService)

	// 公开路由 - 认证和法条搜索
	public := app.Group("/api/v1")
	{
		// 认证相关路由
		auth := public.Group("/auth")
		{
			// 创建用户服务用于认证
			userRepo := repositories.NewUserRepository(db)
			userService := services.NewUserService(userRepo)
			authHandler := handlers.NewAuthHandler(userService)

			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
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