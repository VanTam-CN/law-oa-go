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
	docRepo := repositories.NewDocumentRepository(db)
	lawyerRepo := repositories.NewLawyerRepository(db)

	// 初始化服务
	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)
	caseService := services.NewCaseService(db)
	documentService := services.NewDocumentService(docRepo, "./uploads")
	searchService := services.NewSearchService(esClient, "lawoa_")
	lawyerService := services.NewLawyerService(lawyerRepo)

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	clientHandler := handlers.NewClientHandler(clientService)
	caseHandler := handlers.NewCaseHandler(caseService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	searchHandler := handlers.NewSearchHandler(searchService)
	dashboardHandler := handlers.NewDashboardHandler(userService, clientService, caseService)
	avatarHandler := handlers.NewAvatarHandler(userService)
	lawyerHandler := handlers.NewLawyerHandler(lawyerService)
	approvalHandler := handlers.NewApprovalHandler()

	// API路由组
	apiV1 := app.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware())

	// 兼容前端的路由
	api := app.Group("/api")
	api.Use(middleware.AuthMiddleware())

	// 认证相关路由（无需认证）
	auth := app.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// 用户相关路由
	userV1 := apiV1.Group("/users")
	{
		userV1.GET("/profile", authHandler.GetProfile)
		userV1.PUT("/profile", authHandler.UpdateProfile)
		userV1.POST("/change-password", authHandler.ChangePassword)
		userV1.POST("/logout", authHandler.Logout)
		userV1.POST("/avatar", avatarHandler.UploadAvatar)
	}
	api.GET("/users/profile", authHandler.GetProfile)
	api.PUT("/users/profile", authHandler.UpdateProfile)
	api.POST("/users/change-password", authHandler.ChangePassword)
	api.POST("/users/logout", authHandler.Logout)

	// 管理员用户管理路由
	adminUsersV1 := apiV1.Group("/admin/users")
	adminUsersV1.Use(middleware.RoleMiddleware("admin"))
	{
		adminUsersV1.GET("", userHandler.ListUsers)
		adminUsersV1.GET("/:id", userHandler.GetUser)
		adminUsersV1.POST("", userHandler.CreateUser)
		adminUsersV1.PUT("/:id", userHandler.UpdateUser)
		adminUsersV1.DELETE("/:id", userHandler.DeleteUser)
	}
	// 兼容前端的用户管理路由
	api.GET("/users", userHandler.ListUsers)
	api.GET("/users/:id", userHandler.GetUser)
	api.POST("/users", userHandler.CreateUser)
	api.PUT("/users/:id", userHandler.UpdateUser)
	api.DELETE("/users/:id", userHandler.DeleteUser)

	// 客户相关路由
	clientsV1 := apiV1.Group("/clients")
	{
		clientsV1.GET("", clientHandler.ListClients)
		clientsV1.GET("/:id", clientHandler.GetClient)
		clientsV1.POST("", clientHandler.CreateClient)
		clientsV1.PUT("/:id", clientHandler.UpdateClient)
		clientsV1.DELETE("/:id", clientHandler.DeleteClient)
		clientsV1.GET("/stats", clientHandler.GetClientStats)
	}
	api.GET("/clients", clientHandler.ListClients)
	api.GET("/clients/stats", clientHandler.GetClientStats)

	// 律师相关路由
	lawyersV1 := apiV1.Group("/lawfirm/lawyers")
	{
		lawyersV1.GET("", lawyerHandler.ListLawyers)
		lawyersV1.GET("/stats", lawyerHandler.GetLawyerStats)
	}
	api.GET("/lawfirm/lawyers", lawyerHandler.ListLawyers)
	api.GET("/lawfirm/lawyers/stats", lawyerHandler.GetLawyerStats)

	// 案件相关路由
	casesV1 := apiV1.Group("/cases")
	{
		casesV1.GET("", caseHandler.ListCases)
		casesV1.GET("/:id", caseHandler.GetCase)
		casesV1.POST("", caseHandler.CreateCase)
		casesV1.PUT("/:id", caseHandler.UpdateCase)
		casesV1.DELETE("/:id", caseHandler.DeleteCase)
		casesV1.GET("/stats", caseHandler.GetCaseStats)
		casesV1.POST("/:id/assign", caseHandler.AssignLawyer)
		casesV1.POST("/:id/status", caseHandler.UpdateCaseStatus)
	}
	api.GET("/cases", caseHandler.ListCases)
	api.GET("/cases/stats", caseHandler.GetCaseStats)

	// 文档相关路由
	documentsV1 := apiV1.Group("/documents")
	{
		documentsV1.POST("", documentHandler.UploadDocument)
		documentsV1.GET("/:id", documentHandler.GetDocument)
		documentsV1.PUT("/:id", documentHandler.UpdateDocument)
		documentsV1.DELETE("/:id", documentHandler.DeleteDocument)
		documentsV1.GET("", documentHandler.ListDocuments)
		documentsV1.GET("/stats", documentHandler.GetDocumentStats)
		documentsV1.GET("/:id/download", documentHandler.DownloadDocument)
	}

	// 文件相关路由
	filesV1 := apiV1.Group("/file")
	{
		filesV1.GET("/list", documentHandler.ListDocuments)
	}
	api.GET("/file/list", documentHandler.ListDocuments)

	// 搜索相关路由
	searchV1 := apiV1.Group("/search")
	{
		searchV1.GET("", searchHandler.Search)
		searchV1.GET("/suggestions", searchHandler.GetSearchSuggestions)
		searchV1.POST("/reindex", searchHandler.ReindexAll)
	}

	// 仪表盘相关路由
	dashboardV1 := apiV1.Group("/dashboard")
	{
		dashboardV1.GET("/statistics", dashboardHandler.GetStatistics)
		dashboardV1.GET("/todos", dashboardHandler.GetTodos)
		dashboardV1.GET("/activities", dashboardHandler.GetActivities)
	}
	// 兼容前端的路由
	api.GET("/dashboard/statistics", dashboardHandler.GetStatistics)
	api.GET("/dashboard/todos", dashboardHandler.GetTodos)
	api.GET("/dashboard/activities", dashboardHandler.GetActivities)

	// 性能测试路由
	performanceV1 := apiV1.Group("/performance")
	performanceV1.Use(middleware.RoleMiddleware("admin"))
	{
		performanceV1.GET("/cache", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Cache performance test",
				"data":    map[string]interface{}{},
			})
		})

		performanceV1.GET("/database", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Database performance test",
				"data":    map[string]interface{}{},
			})
		})
	}

	// 审批相关路由
	approvalsV1 := apiV1.Group("/approvals")
	{
		approvalsV1.GET("", approvalHandler.ListApprovals)
		approvalsV1.GET("/stats", approvalHandler.GetApprovalStats)
		approvalsV1.GET("/pending", approvalHandler.GetPendingApprovals)
		approvalsV1.GET("/:id", approvalHandler.GetApproval)
		approvalsV1.POST("/:id/approve", approvalHandler.Approve)
		approvalsV1.POST("/:id/reject", approvalHandler.Reject)
	}
	// 兼容前端的审批路由
	api.GET("/approvals/stats", approvalHandler.GetApprovalStats)
	api.GET("/approvals/pending", approvalHandler.GetPendingApprovals)

	// 404处理
	app.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Route not found",
		})
	})
}