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

// Init 初始化完整的路由系统
func Init(app *gin.Engine, db *gorm.DB, redisClient *rdb.Client, esClient interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 恢复路由初始化panic: %v", r)
			log.Printf("🚨 panic堆栈: %+v", r)
		}
	}()
	log.Println("初始化完整路由系统...")

	// 初始化基础仓储
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)

	// 初始化冲突检测仓储
	var conflictRepo repositories.BasicConflictRepository
	if redisClient != nil {
		conflictRepo = repositories.NewConflictRepository(db, redisClient)
	} else {
		// 如果没有Redis，创建一个简化版本
		conflictRepo = repositories.NewConflictRepository(db, nil)
	}

	// 初始化法条相关仓储
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

	// 初始化文档相关仓储
	docRepo := repositories.NewDocumentRepository(db)

	// 初始化集成相关仓储
	integrationRepo := repositories.NewIntegrationRepository(db)

	// 初始化服务
	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)
	caseService := services.NewCaseService(caseRepo, clientRepo, userRepo)

	// 初始化增强案例服务
	enhancedCaseService := services.NewEnhancedCaseService(caseRepo, clientRepo, userRepo)
	log.Println("✅ 增强案例服务初始化完成")

	legalStatuteService := services.NewLegalStatuteService(db, legalStatuteRepo, legalCategoryRepo, legalTagRepo, legalEsRepo)
	// 初始化缓存仓库
	cacheRepo := repositories.NewMemoryCacheRepository()
	teamPermissionService := services.NewTeamPermissionService(userRepo, caseRepo, cacheRepo)

	// 初始化风险评估器
	riskAssessor := services.NewRiskAssessor(nil, nil)

	// 初始化冲突检测服务
	conflictService := services.NewConflictDetectionService(
		conflictRepo,
		riskAssessor,
		userRepo,
		clientRepo,
		caseRepo,
	)

	// 初始化审批服务
	approvalService := services.NewApprovalService(db)

	// 初始化审批冲突集成服务
	integrationService := services.NewApprovalConflictIntegrationService(
		approvalService,
		conflictService,
		integrationRepo,
	)

	// 初始化处理器
	log.Println("🔧 初始化认证处理器...")
	authHandler := handlers.NewAuthHandler(userService)
	log.Println("✅ 认证处理器初始化完成")

	log.Println("🔧 初始化用户处理器...")
	userHandler := handlers.NewUserHandler(userService)
	log.Println("✅ 用户处理器初始化完成")

	log.Println("🔧 初始化客户处理器...")
	clientHandler := handlers.NewClientHandler(clientService)
	log.Println("✅ 客户处理器初始化完成")

	log.Println("🔧 初始化案件处理器...")
	caseHandler := handlers.NewCaseHandler(caseService)
	log.Println("✅ 案件处理器初始化完成")

	log.Println("🔧 初始化增强案例处理器...")
	enhancedCaseHandler := handlers.NewEnhancedCaseHandler(enhancedCaseService)
	log.Println("✅ 增强案例处理器初始化完成")

	log.Println("🔧 初始化法条处理器...")
	legalStatuteHandler := handlers.NewLegalStatuteHandler(legalStatuteService)
	log.Println("✅ 法条处理器初始化完成")

	log.Println("🔧 初始化通知处理器...")
	notificationHandler := handlers.NewNotificationHandler()
	log.Println("✅ 通知处理器初始化完成")

	log.Println("🔧 初始化仪表盘处理器...")
	dashboardHandler := handlers.NewDashboardHandler()
	log.Println("✅ 仪表盘处理器初始化完成")

	// 初始化集成处理器
	log.Println("🔧 初始化集成处理器...")
	integrationHandler := handlers.NewIntegrationHandler(integrationService, conflictService)
	log.Println("✅ 集成处理器初始化完成")

	// 调试：初始化冲突检测处理器
	log.Println("🔧 初始化冲突检测处理器...")
	conflictHandler := handlers.NewConflictHandlerSimple(conflictService)
	log.Println("✅ 冲突检测处理器初始化完成")

	log.Println("🔧 初始化团队处理器...")
	teamHandler := handlers.NewTeamHandler(teamPermissionService, caseService)
	log.Println("✅ 团队处理器初始化完成")

	// 初始化文档管理处理器
	documentHandler := handlers.NewDocumentHandlerEnhanced(docRepo, userRepo, "./uploads", "./recycle")
	documentStatsHandler := handlers.NewDocumentStatsHandler(services.NewDocumentStatsService(docRepo))

	// 公开路由组
	log.Println("🔧 开始注册公开路由组...")
	public := app.Group("/api/v1")
	{
		// 认证路由
		log.Println("🔧 注册认证路由...")
		auth := public.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}
		log.Println("✅ 认证路由注册完成")

		// 法条搜索（公开）
		log.Println("🔧 注册法条搜索路由...")
		legal := public.Group("/legal")
		{
			legal.GET("/statutes/search", legalStatuteHandler.SearchStatutes)
			legal.GET("/statutes/:id", legalStatuteHandler.GetStatuteByID)
			legal.GET("/categories", legalStatuteHandler.GetCategories)
			legal.GET("/tags", legalStatuteHandler.GetTags)
			legal.GET("/search/suggestions", legalStatuteHandler.GetSearchSuggestions)
		}
		log.Println("✅ 法条搜索路由注册完成")

	}
	log.Println("✅ 公开路由组注册完成")

	// 需要认证的路由组
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// 律师接口（多个端点指向同一个处理器）
		protected.GET("/lawfirm/lawyers", caseHandler.GetLawyers)
		protected.GET("/lawyers", caseHandler.GetLawyers)

		// 案件类型接口
		protected.GET("/case-types", caseHandler.GetCaseTypes)

		// 用户管理
		users := protected.Group("/admin/users")
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// 客户管理
		clients := protected.Group("/clients")
		{
			clients.GET("", clientHandler.ListClients)
			clients.GET("/:id", clientHandler.GetClient)
			clients.POST("", clientHandler.CreateClient)
			clients.PUT("/:id", clientHandler.UpdateClient)
			clients.DELETE("/:id", clientHandler.DeleteClient)
			clients.GET("/stats", clientHandler.GetClientStats)
		}

		// 案件管理
		cases := protected.Group("/cases")
		{
			cases.GET("", caseHandler.ListCases)
			cases.GET("/:id", caseHandler.GetCase)
			cases.POST("", caseHandler.CreateCase)
			cases.PUT("/:id", caseHandler.UpdateCase)
			cases.DELETE("/:id", caseHandler.DeleteCase)
		}

		// 增强案例管理
		enhancedCases := protected.Group("/enhanced-cases")
		{
			enhancedCases.GET("", enhancedCaseHandler.ListEnhancedCases)
			enhancedCases.GET("/:id", enhancedCaseHandler.GetEnhancedCase)
			enhancedCases.POST("", enhancedCaseHandler.CreateEnhancedCase)
			enhancedCases.PUT("/:id", enhancedCaseHandler.UpdateEnhancedCase)
			enhancedCases.DELETE("/:id", enhancedCaseHandler.DeleteEnhancedCase)
			enhancedCases.POST("/:id/conflict-check", enhancedCaseHandler.PerformConflictCheck)
			enhancedCases.POST("/:id/clients", enhancedCaseHandler.AddClientToCase)
			enhancedCases.DELETE("/:id/clients/:client_id", enhancedCaseHandler.RemoveClientFromCase)
		}

	
		// 法条管理（需要认证）
		legal := protected.Group("/legal")
		{
			legal.POST("/statutes", legalStatuteHandler.CreateStatute)
			legal.PUT("/statutes/:id", legalStatuteHandler.UpdateStatute)
			legal.DELETE("/statutes/:id", legalStatuteHandler.DeleteStatute)
		}

		// 通知管理
		notifications := protected.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.GET("/stats", notificationHandler.GetNotificationStats)
		}

		// 仪表盘管理
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/statistics", dashboardHandler.GetDashboardStatistics)
			dashboard.GET("/todos", dashboardHandler.GetDashboardTodos)
			dashboard.GET("/activities", dashboardHandler.GetDashboardActivities)
		}

		log.Println("🔧 开始注册冲突检测路由...")
		conflict := protected.Group("/conflict")
		{
			conflict.POST("/check", conflictHandler.CheckConflict)
			conflict.GET("/health", conflictHandler.HealthCheck)
			conflict.GET("/history", conflictHandler.GetCheckHistory)
			conflict.GET("/stats", conflictHandler.GetConflictStats)
		}
		log.Println("✅ 冲突检测路由注册完成")

		// 集成管理
		log.Println("🔧 开始注册集成路由...")
		integration := protected.Group("/integration")
		{
			// 集成工作流相关
			integration.POST("/approvals", integrationHandler.CreateIntegratedApproval)
			integration.POST("/approvals/with-conflict", integrationHandler.CreateApprovalWithConflict)
			integration.GET("/approvals/:id/status", integrationHandler.GetApprovalIntegrationStatus)
			integration.POST("/approvals/:id/case", integrationHandler.PerformCaseCreation)

			// 冲突检测触发
			integration.POST("/conflict-check", integrationHandler.TriggerConflictCheck)

			// 集成统计
			integration.GET("/statistics", integrationHandler.GetIntegrationStatistics)
		}
		log.Println("✅ 集成路由注册完成")

		// 团队管理
		teams := protected.Group("/teams")
		{
			teams.POST("/assign", teamHandler.AssignTeam)                    // 分配团队
			teams.POST("/check-permission", teamHandler.CheckTeamPermission) // 检查权限

			// 案件团队管理
			caseTeams := teams.Group("/case")
			{
				caseTeams.GET("/:id", teamHandler.GetTeamAssignment)                 // 获取案件团队
				caseTeams.GET("/:id/members", teamHandler.GetTeamMembers)           // 获取团队成员列表
				caseTeams.PUT("/:caseId/member/:memberId", teamHandler.UpdateTeamMember) // 更新团队成员
				caseTeams.DELETE("/:caseId/member/:memberId", teamHandler.RemoveTeamMember) // 移除团队成员
			}
		}

		// 文档管理
		documents := protected.Group("/documents")
		{
			// 基础文档操作
			documents.POST("", documentHandler.UploadDocument)                 // 上传文档
			documents.GET("", documentHandler.ListDocuments)                   // 列出文档
			documents.GET("/:id", documentHandler.GetDocument)                 // 获取文档详情
			documents.PUT("/:id", documentHandler.UpdateDocument)              // 更新文档
			documents.DELETE("/:id", documentHandler.DeleteDocument)           // 删除文档
			documents.GET("/:id/download", documentHandler.DownloadDocument)   // 下载文档
			documents.GET("/:id/preview", documentHandler.GetDocumentPreview)   // 预览文档
			// documents.GET("/:id/content", documentHandler.GetDocumentContent)   // 获取文档内容 - 暂时注释

			// 文档版本管理 - 暂时注释掉未实现的方法
			// documents.GET("/:id/versions", documentHandler.GetVersions)                // 获取版本历史
			// documents.POST("/:id/versions", documentHandler.CreateVersion)             // 创建新版本
			// documents.GET("/:id/versions/:versionId", documentHandler.GetVersion)      // 获取特定版本
			// documents.PUT("/:id/versions/:versionId/restore", documentHandler.RestoreVersion) // 恢复版本
			// documents.GET("/:id/versions/:versionId/compare", documentHandler.CompareVersions) // 比较版本

			// 文档权限管理 - 暂时注释掉未实现的方法
			// documents.GET("/:id/permissions", documentHandler.GetDocumentPermissions)   // 获取文档权限
			// documents.POST("/:id/permissions", documentHandler.GrantPermission)        // 授予权限
			// documents.PUT("/:id/permissions/:userId", documentHandler.UpdatePermission) // 更新权限
			// documents.DELETE("/:id/permissions/:userId", documentHandler.RevokePermission) // 撤销权限
			// documents.POST("/:id/share", documentHandler.ShareDocument)                // 分享文档

			// 文档搜索 - 暂时注释掉未实现的方法
			// documents.POST("/search", documentHandler.SearchDocuments)               // 搜索文档
			// documents.POST("/advanced-search", documentHandler.AdvancedSearch)       // 高级搜索
			// documents.GET("/search/filters", documentHandler.GetSearchFilters)       // 获取搜索过滤器
			// documents.GET("/search/suggestions", documentHandler.GetSearchSuggestions) // 获取搜索建议

			// 回收站操作 - 暂时注释掉未实现的方法
			// documents.GET("/recycle", documentHandler.GetRecycleBin)                 // 获取回收站列表
			// documents.POST("/:id/soft-delete", documentHandler.SoftDeleteDocument)   // 软删除文档
			// documents.POST("/restore", documentHandler.RestoreDocuments)             // 恢复文档
			// documents.DELETE("/permanent", documentHandler.PermanentlyDelete)        // 永久删除文档

			// 文档统计
			documentStats := documents.Group("/stats")
			{
				documentStats.GET("/overview", documentStatsHandler.GetOverview)           // 总览统计
				documentStats.GET("/storage", documentStatsHandler.GetStorageUsage)       // 存储使用统计
				documentStats.GET("/users/:user_id", documentStatsHandler.GetUserActivity) // 用户活动统计
				documentStats.GET("/compliance", documentStatsHandler.GetComplianceReport) // 合规报告
				documentStats.GET("/export", documentStatsHandler.ExportStats)            // 导出统计
				documentStats.GET("/dashboard", documentStatsHandler.GetDashboardStats)   // 仪表板统计
			}
		}

		// 审批管理系统
		log.Println("🔧 初始化审批处理器...")
		approvalHandler := handlers.NewApprovalHandler(db)
		log.Println("✅ 审批处理器初始化完成")

		log.Println("🔧 开始注册审批路由...")
		approvals := protected.Group("/approvals")
		{
			approvals.POST("", approvalHandler.CreateApproval)               // 创建审批申请
			approvals.GET("", approvalHandler.ListApprovals)                    // 获取审批列表
			approvals.GET("/stats", approvalHandler.GetApprovalStats)         // 获取审批统计
			approvals.GET("/pending", approvalHandler.GetPendingApprovals)     // 获取待审批列表
			approvals.GET("/:id", approvalHandler.GetApproval)                // 获取审批详情
			approvals.GET("/workflows", approvalHandler.GetApprovalWorkflows) // 获取工作流列表
			approvals.GET("/templates", approvalHandler.GetApprovalTemplates) // 获取模板列表
		}
		log.Println("✅ 审批路由注册完成")

		log.Println("完整路由系统初始化完成")
}
}