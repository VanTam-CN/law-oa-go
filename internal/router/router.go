package router

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
	"time"
)

// RouterConfig 路由器配置
type RouterConfig struct {
	DB             *gorm.DB
	Redis          *redis.Client
	Elasticsearch  *elasticsearch.Client
	AllowedOrigins []string
	RateLimit      int
	Timeout        time.Duration
}

// NewRouter 创建新的路由器
func NewRouter(config *RouterConfig) *gin.Engine {
	app := gin.Default()

	// 基础中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// CORS中间件
	corsConfig := middleware.CORSConfig{
		AllowedOrigins: config.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
		MaxAge:         "86400",
	}
	app.Use(middleware.CORSWithConfig(corsConfig))

	// 限流中间件
	rateLimitConfig := middleware.RateLimitConfig{
		RedisClient: config.Redis,
		KeyPrefix:   "rate_limit:",
		Limit:       int64(config.RateLimit),
		Window:      time.Minute,
		SkipFunc:    nil,
	}
	app.Use(middleware.RateLimitMiddleware(rateLimitConfig))

	// 初始化路由
	Init(app, config.DB, config.Redis, config.Elasticsearch)

	return app
}

func Init(app *gin.Engine, db *gorm.DB, redisClient *redis.Client, esClient *elasticsearch.Client) {
	// 初始化Repository
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	docRepo := repositories.NewDocumentRepository(db) // Assuming this exists

	// 初始化服务
	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)
	caseService := services.NewCaseService(db)
	documentService := services.NewDocumentService(docRepo, "./uploads") // Assuming storage dir
	searchService := services.NewSearchService(esClient, "lawoa_")

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	clientHandler := handlers.NewClientHandler(clientService)
	caseHandler := handlers.NewCaseHandler(caseService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	searchHandler := handlers.NewSearchHandler(searchService)
	dashboardHandler := handlers.NewDashboardHandler(userService, clientService, caseService)

	// API路由组
	api := app.Group("/api/v1")

	// 健康检查 - 移到main.go中统一管理

	// 认证相关路由（无需认证）
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// 用户相关路由
		user := protected.Group("/users")
		{
			user.GET("/profile", authHandler.GetProfile)
			user.PUT("/profile", authHandler.UpdateProfile)
			user.POST("/change-password", authHandler.ChangePassword)
			user.POST("/logout", authHandler.Logout)
		}

		// 管理员用户管理路由
		adminUsers := protected.Group("/admin/users")
		adminUsers.Use(middleware.RoleMiddleware("admin"))
		{
			adminUsers.GET("", userHandler.ListUsers)
			adminUsers.GET("/:id", userHandler.GetUser)
			adminUsers.POST("", userHandler.CreateUser)
			adminUsers.PUT("/:id", userHandler.UpdateUser)
			adminUsers.DELETE("/:id", userHandler.DeleteUser)
		}

		// 客户相关路由
		clients := protected.Group("/clients")
		{
			clients.GET("", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("", clientHandler.CreateClient)
			clients.PUT("/:id", clientHandler.UpdateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
			clients.GET("/stats", clientHandler.GetClientStats)
		}

		// 案件相关路由
		cases := protected.Group("/cases")
		{
			cases.GET("", caseHandler.ListCases)
			cases.GET("/:id", caseHandler.GetCase)
			cases.POST("", caseHandler.CreateCase)
			cases.PUT("/:id", caseHandler.UpdateCase)
			cases.DELETE("/:id", caseHandler.DeleteCase)
			cases.GET("/stats", caseHandler.GetCaseStats)
			cases.POST("/:id/assign", caseHandler.AssignLawyer)
			cases.POST("/:id/status", caseHandler.UpdateCaseStatus)
		}

		// 文档相关路由
		documents := protected.Group("/documents")
		{
			documents.POST("", documentHandler.UploadDocument)
			documents.GET("/:id", documentHandler.GetDocument)
			documents.PUT("/:id", documentHandler.UpdateDocument)
			documents.DELETE("/:id", documentHandler.DeleteDocument)
			documents.GET("", documentHandler.ListDocuments)
			documents.GET("/stats", documentHandler.GetDocumentStats)
			documents.GET("/:id/download", documentHandler.DownloadDocument)
		}

		// 搜索相关路由
		search := protected.Group("/search")
		{
			search.GET("", searchHandler.Search)
			search.GET("/suggestions", searchHandler.GetSearchSuggestions)
			search.POST("/reindex", searchHandler.ReindexAll)
		}

		// 仪表盘相关路由
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/statistics", dashboardHandler.GetStatistics)
			dashboard.GET("/todos", dashboardHandler.GetTodos)
			dashboard.GET("/activities", dashboardHandler.GetActivities)
		}

		// 性能测试路由
		performance := protected.Group("/performance")
		performance.Use(middleware.RoleMiddleware("admin"))
		{
			performance.GET("/cache", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Cache performance test",
					"data":    map[string]interface{}{},
				})
			})

			performance.GET("/database", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Database performance test",
					"data":    map[string]interface{}{},
				})
			})
		}
	}

	// 404处理
	app.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Route not found",
		})
	})
}
