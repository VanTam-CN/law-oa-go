package router

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/monitoring"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
	"os"
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

	// 初始化性能监控
	monitoring.InitPerformanceMetrics()

	// 性能优化中间件（在最前面）
	app.Use(middleware.Recovery())
	app.Use(middleware.SecurityHeaders())

	// 性能监控中间件
	app.Use(middleware.PerformanceMiddleware())

	// 并发限制中间件（最大100个并发请求）
	concurrencyLimiter := middleware.NewConcurrencyLimiter(100)
	app.Use(concurrencyLimiter.Limit())

	// 健康检查中间件
	app.Use(middleware.HealthCheckMiddleware())

	// 缓存中间件 - 为静态内容和API响应启用缓存
	cacheConfig := middleware.CacheConfig{
		TTL:             5 * time.Minute,
		SkipHeader:      "Cache-Control",
		RedisClient:     config.Redis,
		KeyPrefix:       "lawoa",
		MaxBodySize:     1024 * 1024, // 1MB
		SkipRoutes:      []string{"/api/auth/", "/api/upload", "/api/file/"},
		CacheableRoutes: []string{"/api/dashboard/", "/api/stats", "/api/users/profile"},
	}
	app.Use(middleware.CacheMiddleware(cacheConfig))

	// 日志中间件（优化版本）
	app.Use(middleware.Logger())

	// 统一错误处理中间件
	errorHandler := middleware.NewErrorHandler(nil, middleware.ErrorHandlerConfig{})
	app.Use(errorHandler.ErrorHandlingMiddleware())

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

	// 初始化冲突检测服务
	conflictRepo := repositories.NewConflictRepository(db, redisClient)
	mcpConfig := &services.MCPConfig{
		APIKey:  os.Getenv("MCP_API_KEY"),
		BaseURL: os.Getenv("MCP_API_URL"),
		Timeout: 30 * time.Second,
	}
	mcpClient := services.NewMCPClient(mcpConfig, redisClient)
	ruleEngine := services.NewRuleEngine(conflictRepo, mcpClient)
	riskAssessor := services.NewRiskAssessor(nil, ruleEngine)
	conflictService := services.NewConflictService(conflictRepo, mcpClient, ruleEngine, riskAssessor)

	// 初始化服务
	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)
	caseService := services.NewCaseService(db, conflictService, true) // 启用冲突检测
	documentService := services.NewDocumentService(docRepo, "./uploads")
	searchService := services.NewSearchService(esClient, "lawoa_")
	lawyerService := services.NewLawyerService(lawyerRepo)
	rbacService := services.NewRBACService(db)

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
	rbacHandler := handlers.NewRBACHandler(rbacService)
	conflictHandler := handlers.NewConflictHandler(conflictService)
	notificationHandler := handlers.NewNotificationHandler()

	// API路由组
	apiV1 := app.Group("/api/v1")
	apiV1.Use(middleware.AuthMiddleware())

	// 兼容前端的路由 (不强制认证，用于登录注册)
	api := app.Group("/api")
	// 注意：认证相关路由不需要中间件保护

	// 认证相关路由（无需认证）
	auth := app.Group("/api/auth")
	// 为认证路由添加更严格的速率限制
	authRateLimitConfig := middleware.RateLimitConfig{
		RedisClient: redisClient,
		KeyPrefix:   "auth_rate_limit:",
		Limit:       5,  // 每分钟最多5次尝试
		Window:      time.Minute,
		SkipFunc:    nil,
	}
	auth.Use(middleware.RateLimitMiddleware(authRateLimitConfig))
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
	// 兼容前端的用户管理路由 (需要认证)
	apiAuthenticated := api.Group("")
	apiAuthenticated.Use(middleware.AuthMiddleware())
	{
		// 用户资料路由
		apiAuthenticated.GET("/users/profile", authHandler.GetProfile)
		apiAuthenticated.PUT("/users/profile", authHandler.UpdateProfile)
		apiAuthenticated.POST("/users/change-password", authHandler.ChangePassword)
		apiAuthenticated.POST("/users/logout", authHandler.Logout)

		// 用户管理路由
		apiAuthenticated.GET("/users", userHandler.ListUsers)
		apiAuthenticated.GET("/users/:id", userHandler.GetUser)
		apiAuthenticated.POST("/users", userHandler.CreateUser)
		apiAuthenticated.PUT("/users/:id", userHandler.UpdateUser)
		apiAuthenticated.DELETE("/users/:id", userHandler.DeleteUser)
	}

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
	// 兼容前端的客户管理路由 (需要认证)
	apiAuthenticated.GET("/clients", clientHandler.ListClients)
	apiAuthenticated.POST("/clients", clientHandler.CreateClient)
	apiAuthenticated.GET("/clients/:id", clientHandler.GetClient)
	apiAuthenticated.PUT("/clients/:id", clientHandler.UpdateClient)
	apiAuthenticated.DELETE("/clients/:id", clientHandler.DeleteClient)
	apiAuthenticated.GET("/clients/stats", clientHandler.GetClientStats)

	// 兼容前端的案件管理路由 (需要认证)
	apiAuthenticated.GET("/cases", caseHandler.ListCases)
	apiAuthenticated.POST("/cases", caseHandler.CreateCase)
	apiAuthenticated.GET("/cases/stats", caseHandler.GetCaseStats) // 必须在 /cases/:id 之前
	apiAuthenticated.GET("/cases/:id", caseHandler.GetCase)
	apiAuthenticated.PUT("/cases/:id", caseHandler.UpdateCase)
	apiAuthenticated.DELETE("/cases/:id", caseHandler.DeleteCase)

	// 兼容前端的文件管理路由 (需要认证)
	apiAuthenticated.GET("/files", documentHandler.ListDocuments)
	apiAuthenticated.GET("/files/stats", documentHandler.GetDocumentStats)

	// 兼容前端的律师管理路由 (需要认证)
	apiAuthenticated.GET("/lawyers", lawyerHandler.ListLawyers)
	apiAuthenticated.GET("/lawyers/stats", lawyerHandler.GetLawyerStats)

	// 兼容前端的仪表盘路由 (需要认证)
	apiAuthenticated.GET("/dashboard", dashboardHandler.GetStatistics)
	apiAuthenticated.GET("/todos", dashboardHandler.GetTodos)
	apiAuthenticated.GET("/activities", dashboardHandler.GetActivities)

	// 兼容前端的审批路由 (需要认证)
	apiAuthenticated.GET("/approvals", approvalHandler.ListApprovals)
	apiAuthenticated.GET("/approvals/stats", approvalHandler.GetApprovalStats)
	apiAuthenticated.GET("/approvals/pending", approvalHandler.GetPendingApprovals)
	apiAuthenticated.GET("/approvals/:id", approvalHandler.GetApproval)

	// 兼容前端的用户路由 (需要认证)
	apiAuthenticated.GET("/admin/users", userHandler.ListUsers)
	apiAuthenticated.POST("/admin/users", userHandler.CreateUser)
	apiAuthenticated.PUT("/admin/users/:id", userHandler.UpdateUser)
	apiAuthenticated.DELETE("/admin/users/:id", userHandler.DeleteUser)

	// 兼容前端的冲突检测路由 (需要认证)
	apiAuthenticated.POST("/conflict/check", conflictHandler.CheckConflict)
	apiAuthenticated.GET("/conflict/stats", conflictHandler.GetConflictStats)
	apiAuthenticated.GET("/conflict/rules", conflictHandler.GetConflictRules)
	apiAuthenticated.GET("/conflict/standards", conflictHandler.GetMCPStandards)

	// 律师相关路由
	lawyersV1 := apiV1.Group("/lawfirm/lawyers")
	{
		lawyersV1.GET("", lawyerHandler.ListLawyers)
		lawyersV1.GET("/stats", lawyerHandler.GetLawyerStats)
	}
	// 兼容前端的律师管理路由 (需要认证)
	apiAuthenticated.GET("/lawfirm/lawyers", lawyerHandler.ListLawyers)
	apiAuthenticated.GET("/lawfirm/lawyers/stats", lawyerHandler.GetLawyerStats)

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
	// 兼容前端的仪表盘路由 (需要认证)
	apiAuthenticated.GET("/dashboard/statistics", dashboardHandler.GetStatistics)
	apiAuthenticated.GET("/dashboard/todos", dashboardHandler.GetTodos)
	apiAuthenticated.GET("/dashboard/activities", dashboardHandler.GetActivities)

	// 兼容前端的通知路由 (需要认证)
	apiAuthenticated.GET("/notifications", notificationHandler.GetNotifications)
	apiAuthenticated.GET("/notifications/stats", notificationHandler.GetNotificationStats)
	apiAuthenticated.POST("/notifications/:id/read", notificationHandler.MarkAsRead)
	apiAuthenticated.POST("/notifications/read-all", notificationHandler.MarkAllAsRead)
	apiAuthenticated.DELETE("/notifications/:id", notificationHandler.DeleteNotification)

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
	// 兼容前端的审批路由 (stats和pending路由已在上面注册)
	// api.GET("/approvals/stats", approvalHandler.GetApprovalStats) // 已在上面注册
	// api.GET("/approvals/pending", approvalHandler.GetPendingApprovals) // 已在上面注册

	// RBAC相关路由
	rolesV1 := apiV1.Group("/roles")
	rolesV1.Use(middleware.RoleMiddleware("admin"))
	{
		rolesV1.GET("", rbacHandler.GetRoleList)
		rolesV1.POST("", rbacHandler.CreateRole)
		rolesV1.GET("/:id", rbacHandler.GetRoleById)
		rolesV1.PUT("/:id", rbacHandler.UpdateRole)
		rolesV1.DELETE("/:id", rbacHandler.DeleteRole)
		rolesV1.PUT("/:id/status", rbacHandler.UpdateRoleStatus)
		rolesV1.GET("/:id/permissions", rbacHandler.GetRolePermissions)
		rolesV1.POST("/:id/permissions", rbacHandler.AssignRolePermissions)
	}

	permissionsV1 := apiV1.Group("/permissions")
	permissionsV1.Use(middleware.RoleMiddleware("admin"))
	{
		permissionsV1.GET("", rbacHandler.GetPermissionList)
		permissionsV1.POST("", rbacHandler.CreatePermission)
		permissionsV1.GET("/:id", rbacHandler.GetPermissionById)
		permissionsV1.PUT("/:id", rbacHandler.UpdatePermission)
		permissionsV1.DELETE("/:id", rbacHandler.DeletePermission)
	}

	// 兼容前端的RBAC路由
	api.GET("/roles", rbacHandler.GetRoleList)
	api.GET("/permissions", rbacHandler.GetPermissionList)

	// 当前用户角色和权限路由
	apiAuthenticated.GET("/admin/current-user/roles", rbacHandler.GetCurrentUserRoles)
	apiAuthenticated.GET("/admin/current-user/permissions", rbacHandler.GetCurrentUserPermissions)

	// 冲突检测路由
	conflictV1 := apiV1.Group("/conflict")
	{
		conflictV1.POST("/check", conflictHandler.CheckConflict)
		conflictV1.GET("/history/:clientId", conflictHandler.GetCheckHistory)
		conflictV1.GET("/details/:checkId", conflictHandler.GetCheckDetails)
		conflictV1.GET("/rules", conflictHandler.GetConflictRules)
		conflictV1.POST("/rules", conflictHandler.CreateConflictRule)
		conflictV1.PUT("/rules/:ruleId", conflictHandler.UpdateConflictRule)
		conflictV1.DELETE("/rules/:ruleId", conflictHandler.DeleteConflictRule)
		conflictV1.GET("/standards", conflictHandler.GetMCPStandards)
		conflictV1.GET("/stats", conflictHandler.GetConflictStats)
		conflictV1.GET("/health", conflictHandler.HealthCheck)
	}

	// 兼容前端的冲突检测路由 (已注册)
	// api.POST("/conflict/check", conflictHandler.CheckConflict) // 已在上面注册
	// api.GET("/conflict/history/:clientId", conflictHandler.GetCheckHistory) // 已在上面注册
	// api.GET("/conflict/details/:checkId", conflictHandler.GetCheckDetails) // 已在上面注册
	// api.GET("/conflict/rules", conflictHandler.GetConflictRules) // 已在上面注册
	// api.GET("/conflict/standards", conflictHandler.GetMCPStandards) // 已在上面注册
	// api.GET("/conflict/stats", conflictHandler.GetConflictStats) // 已在上面注册

	// 404处理
	app.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Route not found",
		})
	})
}