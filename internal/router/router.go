package router

import (
	"log"

	"law-oa-go/internal/auth"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
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

	
	// 初始化服务
	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)
	caseService := services.NewCaseService(caseRepo, clientRepo, userRepo)

	// 初始化增强案例服务
	enhancedCaseService := services.NewEnhancedCaseService(caseRepo, clientRepo, userRepo)
	log.Println("✅ 增强案例服务初始化完成")

	// 审批服务将在处理器中初始化
	log.Println("🔧 审批服务将在处理器中初始化")

	
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

	// 初始化认证处理器
	log.Println("🔧 初始化认证处理器...")

	// 初始化令牌管理器
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		return
	}
	cacheService := cache.NewCacheService(redisClient, "law-oa")
	tokenManager := auth.NewTokenManager(cfg, redisClient, cacheService)

	// 使用适配器包装 TokenManager 以实现 TokenManagerInterface
	tokenManagerAdapter := auth.NewTokenManagerAdapter(tokenManager)

	// 初始化令牌撤销服务
	tokenRevocationService := auth.NewTokenRevocationService(tokenManagerAdapter, redisClient, db)

	authHandler := handlers.NewAuthHandler(userService, tokenRevocationService)
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
	notificationHandler := handlers.NewNotificationHandlerWithDB(db)
	log.Println("✅ 通知处理器初始化完成")

	log.Println("🔧 初始化仪表盘处理器...")
	dashboardHandler := handlers.NewDashboardHandler()
	log.Println("✅ 仪表盘处理器初始化完成")

	// 调试：初始化冲突检测处理器
	log.Println("🔧 初始化冲突检测处理器...")
	conflictHandler := handlers.NewConflictHandlerSimple(conflictService)
	log.Println("✅ 冲突检测处理器初始化完成")

	// 初始化隔离墙仓储和服务
	ethicalWallRepo := repositories.NewEthicalWallRepository(db)
	ethicalWallService := services.NewEthicalWallService(ethicalWallRepo, caseRepo, userRepo)
	log.Println("✅ 隔离墙服务初始化完成")

	log.Println("🔧 初始化隔离墙处理器...")
	ethicalWallHandler := handlers.NewEthicalWallHandler(ethicalWallService)
	log.Println("✅ 隔离墙处理器初始化完成")

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

	// 初始化隔离墙中间件配置
	ethicalWallConfig := middleware.EthicalWallConfig{
		EthicalWallRepo: ethicalWallRepo,
		SkipPaths:       middleware.GetEthicalWallSkipPaths(),
		SkipPrefixes:    middleware.GetEthicalWallSkipPrefixes(),
	}

	// 需要认证的路由组
	protected := app.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// 令牌撤销相关路由（需要认证）
		authRevocation := protected.Group("/auth")
		{
			authRevocation.POST("/logout", authHandler.Logout)
			authRevocation.POST("/revoke/user", authHandler.RevokeUserTokens)
			authRevocation.POST("/revoke/device", authHandler.RevokeDeviceTokens)
			authRevocation.POST("/revoke/all", authHandler.RevokeAllTokens)
			authRevocation.GET("/devices/:user_id", authHandler.GetActiveDevices)
			authRevocation.GET("/revocation-history/:user_id", authHandler.GetRevocationHistory)
		}

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

		// 案件管理 - 需要隔离墙保护
		cases := protected.Group("/cases")
		cases.Use(middleware.EthicalWallMiddleware(ethicalWallConfig))
		{
			cases.GET("", caseHandler.ListCases)
			cases.GET("/:id", caseHandler.GetCase)
			cases.POST("", caseHandler.CreateCase)
			cases.PUT("/:id", caseHandler.UpdateCase)
			cases.DELETE("/:id", caseHandler.DeleteCase)
		}

		// 增强案例管理 - 需要隔离墙保护
		enhancedCases := protected.Group("/enhanced-cases")
		enhancedCases.Use(middleware.EthicalWallMiddleware(ethicalWallConfig))
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
			legal.POST("/statutes/import", legalStatuteHandler.BulkImportStatutes)
			legal.PUT("/statutes/:id", legalStatuteHandler.UpdateStatute)
			legal.DELETE("/statutes/:id", legalStatuteHandler.DeleteStatute)
		}

		// 通知管理
		log.Println("🔧 开始注册通知路由...")
		notifications := protected.Group("/notifications")
		{
			// 通知队列 API
			notifications.GET("", notificationHandler.GetNotificationQueue)                        // 获取通知队列列表
			notifications.GET("/stats", notificationHandler.GetNotificationQueueStats)            // 获取通知统计
			notifications.GET("/pending", notificationHandler.GetNotifications)                   // 获取待审批列表（兼容）
			notifications.POST("", notificationHandler.CreateNotification)                        // 创建通知
			notifications.GET("/:id", notificationHandler.GetNotificationByID)                    // 获取通知详情
			notifications.PUT("/:id", notificationHandler.UpdateNotification)                     // 更新通知
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)                  // 删除通知
			notifications.POST("/:id/approve", notificationHandler.ApproveNotification)          // 审批通过
			notifications.POST("/:id/reject", notificationHandler.RejectNotification)             // 审批拒绝
			notifications.POST("/:id/send", notificationHandler.SendNotification)                 // 发送通知
			notifications.POST("/batch/confirm", notificationHandler.BatchConfirmNotification)    // 批量确认
			notifications.POST("/batch/cancel", notificationHandler.BatchCancelNotification)      // 批量取消
		}
		log.Println("✅ 通知路由注册完成")

		// 通知模板管理
		log.Println("🔧 开始注册通知模板路由...")
		templates := protected.Group("/notification-templates")
		{
			templates.GET("", notificationHandler.GetTemplates)                                  // 获取模板列表
			templates.GET("/active", notificationHandler.GetActiveTemplates)                     // 获取启用模板列表
			templates.GET("/code/:code", notificationHandler.GetTemplateByCode)                  // 根据代码获取模板
			templates.POST("", notificationHandler.CreateTemplate)                               // 创建模板
			templates.GET("/:id", notificationHandler.GetNotificationByID)                        // 获取模板详情（复用）
			templates.PUT("/:id", notificationHandler.UpdateTemplate)                            // 更新模板
			templates.DELETE("/:id", notificationHandler.DeleteTemplate)                         // 删除模板
			templates.PUT("/:id/toggle", notificationHandler.ToggleTemplateActive)               // 切换启用状态
		}
		log.Println("✅ 通知模板路由注册完成")

		// 内容过滤管理
		log.Println("🔧 开始注册内容过滤路由...")
		contentFilterHandler := handlers.NewContentFilterHandler(db)
		contentFilter := protected.Group("/content-filter")
		{
			// 敏感词管理
			contentFilter.POST("/words", contentFilterHandler.CreateSensitiveWord)              // 创建敏感词
			contentFilter.GET("/words", contentFilterHandler.GetSensitiveWords)                 // 获取敏感词列表
			contentFilter.GET("/words/:id", contentFilterHandler.GetSensitiveWordByID)          // 获取敏感词详情
			contentFilter.PUT("/words/:id", contentFilterHandler.UpdateSensitiveWord)           // 更新敏感词
			contentFilter.DELETE("/words/:id", contentFilterHandler.DeleteSensitiveWord)        // 删除敏感词
			contentFilter.POST("/words/batch/import", contentFilterHandler.BatchImportWords)    // 批量导入
			contentFilter.POST("/words/batch/toggle", contentFilterHandler.BatchToggleWords)    // 批量切换状态
			contentFilter.POST("/words/batch/delete", contentFilterHandler.BatchDeleteWords)    // 批量删除
			contentFilter.GET("/words/stats", contentFilterHandler.GetSensitiveWordStats)       // 敏感词统计

			// 内容检测和过滤
			contentFilter.POST("/check", contentFilterHandler.CheckContent)                     // 检查内容
			contentFilter.POST("/filter", contentFilterHandler.FilterContent)                   // 过滤内容

			// 过滤日志
			contentFilter.GET("/logs", contentFilterHandler.GetFilterLogs)                      // 获取过滤日志

			// 缓存管理
			contentFilter.POST("/cache/reset", contentFilterHandler.ResetCache)                 // 重置缓存
		}
		log.Println("✅ 内容过滤路由注册完成")

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


		// 团队管理
		teams := protected.Group("/teams")
		{
			teams.POST("/assign", teamHandler.AssignTeam)                    // 分配团队
			teams.POST("/check-permission", teamHandler.CheckTeamPermission) // 检查权限

			// 案件团队管理 - 需要隔离墙保护
			caseTeams := teams.Group("/case")
			caseTeams.Use(middleware.EthicalWallMiddleware(ethicalWallConfig))
			{
				caseTeams.GET("/:id", teamHandler.GetTeamAssignment)                 // 获取案件团队
				caseTeams.GET("/:id/members", teamHandler.GetTeamMembers)           // 获取团队成员列表
				caseTeams.PUT("/:caseId/member/:memberId", teamHandler.UpdateTeamMember) // 更新团队成员
				caseTeams.DELETE("/:caseId/member/:memberId", teamHandler.RemoveTeamMember) // 移除团队成员
			}
		}

		// 文档管理 - 需要隔离墙保护
		documents := protected.Group("/documents")
		documents.Use(middleware.EthicalWallMiddleware(ethicalWallConfig))
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

		// 初始化集成服务
		log.Println("🔧 初始化集成服务...")
		integrationRepo := repositories.NewIntegrationRepository(db)
		approvalConflictIntegrationService := services.NewApprovalConflictIntegrationService(
			approvalHandler.GetApprovalService(),
			conflictService,
			caseService,
			integrationRepo,
		)
		integrationHandler := handlers.NewIntegrationHandler(approvalConflictIntegrationService, conflictService)
		log.Println("✅ 集成服务初始化完成")

		log.Println("🔧 开始注册审批路由...")
		approvals := protected.Group("/approvals")
		{
			approvals.POST("", approvalHandler.CreateApproval)                    // 创建审批申请
			approvals.GET("", approvalHandler.ListApprovals)                      // 获取审批列表
			approvals.GET("/stats", approvalHandler.GetApprovalStats)            // 获取审批统计
			approvals.GET("/pending", approvalHandler.GetPendingApprovals)        // 获取待审批列表
			approvals.GET("/:id", approvalHandler.GetApproval)                   // 获取审批详情
			approvals.GET("/workflows", approvalHandler.GetApprovalWorkflows)   // 获取工作流列表
			approvals.GET("/templates", approvalHandler.GetApprovalTemplates)   // 获取模板列表

			// 新增：审批闭环相关路由
			approvals.POST("/:id/submit", approvalHandler.SubmitApproval)            // 提交审批
			approvals.POST("/:id/decision", approvalHandler.ProcessApprovalDecision) // 处理审批决定
			approvals.POST("/:id/resubmit", approvalHandler.ResubmitApproval)        // 重新提交
			approvals.POST("/:id/cancel", approvalHandler.CancelApproval)            // 取消审批
			approvals.PUT("/:id", approvalHandler.UpdateApproval)                    // 更新审批
		}
		log.Println("✅ 审批路由注册完成")

		// 注册集成审批路由
		log.Println("🔧 开始注册集成审批路由...")
		integration := protected.Group("/integration")
		{
			// 集成审批相关路由
			integration.POST("/approvals/with-conflict", integrationHandler.CreateApprovalWithConflict) // 创建带冲突检测的审批
			integration.GET("/approvals/:id/status", integrationHandler.GetApprovalIntegrationStatus) // 获取集成状态
			integration.POST("/approvals/:id/case", integrationHandler.PerformCaseCreation)          // 执行案件创建
			integration.POST("/approvals/:id/decision", integrationHandler.ProcessApprovalWithConflict) // 处理集成审批

			// 冲突检测相关路由
			integration.POST("/conflict/check", integrationHandler.TriggerConflictCheck) // 触发冲突检测

			// 统计和历史
			integration.GET("/statistics", integrationHandler.GetIntegrationStatistics) // 获取集成统计
			integration.GET("/history", integrationHandler.GetIntegrationHistory)       // 获取集成历史
			integration.GET("/logs", integrationHandler.GetIntegrationLogs)           // 获取集成日志

			// 重试和取消
			integration.POST("/approvals/:id/retry", integrationHandler.RetryFailedIntegration) // 重试失败的集成
			integration.POST("/approvals/:id/cancel", integrationHandler.CancelIntegration)  // 取消集成
		}
		log.Println("✅ 集成审批路由注册完成")

		// 待办事项路由
		log.Println("🔧 开始注册待办事项路由...")
		inboxHandler := handlers.NewInboxHandler(nil)
		{
			inbox := protected.Group("/inbox")
			{
				inbox.GET("", inboxHandler.ListInboxItems)                   // 获取待办列表
				inbox.GET("/stats", inboxHandler.GetInboxStats)             // 获取统计
				inbox.GET("/:id", inboxHandler.GetInboxItem)               // 获取详情
				inbox.POST("", inboxHandler.CreateInboxItem)               // 创建待办
				inbox.PUT("/:id", inboxHandler.UpdateInboxItem)            // 更新待办
				inbox.DELETE("/:id", inboxHandler.DeleteInboxItem)         // 删除待办
				inbox.PUT("/:id/read", inboxHandler.MarkAsRead)            // 标记已读
				inbox.PUT("/:id/complete", inboxHandler.MarkAsCompleted)   // 标记完成
				inbox.PUT("/:id/snooze", inboxHandler.SnoozeInboxItem)      // 延后待办
			}
		}
		log.Println("✅ 待办事项路由注册完成")

		// 隔离墙管理路由
		log.Println("🔧 开始注册隔离墙路由...")
		// 案件隔离墙管理 - 需要隔离墙保护
		casesEthicalWall := protected.Group("/cases/:caseId/ethical-wall")
		casesEthicalWall.Use(middleware.EthicalWallMiddleware(ethicalWallConfig))
		{
			casesEthicalWall.POST("", ethicalWallHandler.EnableEthicalWall)           // 启用隔离墙
			casesEthicalWall.DELETE("", ethicalWallHandler.DisableEthicalWall)        // 禁用隔离墙
			casesEthicalWall.GET("/whitelist", ethicalWallHandler.GetWhitelist)      // 获取白名单
			casesEthicalWall.POST("/whitelist", ethicalWallHandler.AddToWhitelist)   // 添加白名单
			casesEthicalWall.GET("/access-logs", ethicalWallHandler.GetAccessLogs)    // 获取访问日志
		}
		// 白名单管理（指定用户）- 需要隔离墙保护
		casesEthicalWallWhitelist := protected.Group("/cases/:caseId/ethical-wall/whitelist")
		casesEthicalWallWhitelist.Use(middleware.EthicalWallMiddleware(ethicalWallConfig))
		{
			casesEthicalWallWhitelist.DELETE("/:userId", ethicalWallHandler.RemoveFromWhitelist) // 移除白名单
		}
		// 用户可访问的隔离墙案件
		protected.GET("/ethical-wall/accessible-cases", ethicalWallHandler.GetUserAccessibleCases)
		log.Println("✅ 隔离墙路由注册完成")

		// 财务管理路由
		log.Println("🔧 开始注册财务管理路由...")
		// 初始化财务相关仓储
		contractRepo := repositories.NewContractRepository(db)
		milestoneRepo := repositories.NewPaymentMilestoneRepository(db)
		invoiceRepo := repositories.NewInvoiceRepository(db)
		paymentRepo := repositories.NewPaymentRepository(db)
		badDebtRepo := repositories.NewBadDebtRepository(db)
		commissionRepo := repositories.NewCommissionRepository(db)

		// 初始化财务相关服务
		contractService := services.NewContractService(contractRepo, milestoneRepo, clientRepo, caseRepo, userRepo)
		milestoneService := services.NewPaymentMilestoneService(milestoneRepo, contractRepo, invoiceRepo)
		invoiceService := services.NewInvoiceService(invoiceRepo, contractRepo, milestoneRepo, clientRepo, paymentRepo)
		paymentService := services.NewPaymentService(paymentRepo, invoiceRepo, milestoneRepo, userRepo)
		badDebtService := services.NewBadDebtService(badDebtRepo, contractRepo, invoiceRepo, paymentRepo, userRepo)
		commissionService := services.NewCommissionService(commissionRepo, paymentRepo, contractRepo, userRepo, caseRepo)
		commissionService.SetInvoiceRepository(invoiceRepo)

		financeHandler := handlers.NewFinanceHandler(contractService, milestoneService, invoiceService, paymentService, badDebtService, commissionService)
		log.Println("✅ 财务处理器初始化完成")

		// 财务路由组
		finance := protected.Group("/finance")
		{
			// 合同管理
			finance.GET("/contracts", financeHandler.ListContracts)           // 获取合同列表
			finance.GET("/contracts/stats", financeHandler.GetContractStats)  // 获取合同统计
			finance.POST("/contracts", financeHandler.CreateContract)         // 创建合同
			finance.GET("/contracts/:id", financeHandler.GetContract)         // 获取合同详情
			finance.PUT("/contracts/:id", financeHandler.UpdateContract)      // 更新合同
			finance.DELETE("/contracts/:id", financeHandler.DeleteContract)   // 删除合同
			finance.POST("/contracts/:id/activate", financeHandler.ActivateContract) // 激活合同
			finance.POST("/contracts/:id/suspend", financeHandler.SuspendContract)   // 暂停合同
			finance.POST("/contracts/:id/complete", financeHandler.CompleteContract) // 完成合同

			// 付款计划管理
			finance.POST("/milestones", financeHandler.CreateMilestone)        // 创建付款计划
			finance.PUT("/milestones/:id", financeHandler.UpdateMilestone)     // 更新付款计划
			finance.DELETE("/milestones/:id", financeHandler.DeleteMilestone)  // 删除付款计划
			finance.GET("/contracts/:contract_id/milestones", financeHandler.GetMilestonesByContractID) // 获取合同的付款计划

			// 发票管理
			finance.GET("/invoices", financeHandler.ListInvoices)             // 获取发票列表
			finance.GET("/invoices/stats", financeHandler.GetInvoiceStats)   // 获取发票统计
			finance.POST("/invoices", financeHandler.CreateInvoice)           // 创建发票
			finance.GET("/invoices/:id", financeHandler.GetInvoice)           // 获取发票详情
			finance.PUT("/invoices/:id", financeHandler.UpdateInvoice)        // 更新发票
			finance.DELETE("/invoices/:id", financeHandler.DeleteInvoice)     // 删除发票
			finance.POST("/invoices/:id/submit", financeHandler.SubmitInvoice)           // 提交审批
			finance.POST("/invoices/:id/approve", financeHandler.ApproveInvoice)         // 审批通过
			finance.POST("/invoices/:id/reject", financeHandler.RejectInvoice)           // 审批拒绝
			finance.POST("/invoices/:id/issue", financeHandler.IssueInvoice)             // 开票
			finance.POST("/invoices/:id/confirm-receipt", financeHandler.ConfirmInvoiceReceipt) // 客户签收
			finance.POST("/invoices/:id/cancel", financeHandler.CancelInvoice)           // 作废发票
			finance.GET("/invoices/:invoice_id/payments", financeHandler.GetPaymentsByInvoiceID) // 获取发票的回款

			// 回款管理
			finance.GET("/payments", financeHandler.ListPayments)             // 获取回款列表
			finance.GET("/payments/stats", financeHandler.GetPaymentStats)    // 获取回款统计
			finance.POST("/payments", financeHandler.CreatePayment)           // 创建回款记录
			finance.GET("/payments/:id", financeHandler.GetPayment)           // 获取回款详情
			finance.PUT("/payments/:id", financeHandler.UpdatePayment)        // 更新回款记录
			finance.DELETE("/payments/:id", financeHandler.DeletePayment)     // 删除回款记录
			finance.POST("/payments/:id/confirm", financeHandler.ConfirmPayment) // 确认回款
			finance.POST("/payments/:id/reject", financeHandler.RejectPayment)   // 拒绝回款

			// 坏账核销
			finance.GET("/bad-debts", financeHandler.ListBadDebts)            // 获取坏账核销列表
			finance.GET("/bad-debts/pending", financeHandler.GetPendingBadDebts) // 获取待审批坏账核销
			finance.POST("/bad-debts", financeHandler.CreateBadDebt)           // 创建坏账核销申请
			finance.GET("/bad-debts/:id", financeHandler.GetBadDebt)           // 获取坏账核销详情
			finance.POST("/bad-debts/:id/approve", financeHandler.ApproveBadDebt) // 审批通过
			finance.POST("/bad-debts/:id/reject", financeHandler.RejectBadDebt)   // 审批拒绝

			// 提成管理
			finance.GET("/commissions", financeHandler.ListCommissions)        // 获取提成列表
			finance.GET("/commissions/stats", financeHandler.GetCommissionStats) // 获取提成统计
			finance.POST("/commissions/calculate", financeHandler.CalculateCommissions) // 计算提成
			finance.GET("/commissions/:id", financeHandler.GetCommission)       // 获取提成详情
			finance.POST("/commissions/:id/mark-paid", financeHandler.MarkCommissionAsPaid) // 标记已支付
			finance.POST("/commissions/:id/cancel", financeHandler.CancelCommission)    // 取消提成
			finance.GET("/commissions/beneficiary/:beneficiary_id", financeHandler.GetCommissionsByBeneficiary) // 获取受益人的提成

			// 财务概览
			finance.GET("/overview", financeHandler.GetFinanceOverview)       // 获取财务概览
		}
		log.Println("✅ 财务管理路由注册完成")

		// 代管款管理路由
		log.Println("🔧 开始注册代管款管理路由...")
		// 初始化代管款相关仓储
		trustAccountRepo := repositories.NewTrustAccountRepository(db)
		trustTransactionRepo := repositories.NewTrustTransactionRepository(db)

		// 初始化代管款相关服务
		trustAccountService := services.NewTrustAccountService(trustAccountRepo, trustTransactionRepo, clientRepo, caseRepo, userRepo)
		trustTransactionService := services.NewTrustTransactionService(trustTransactionRepo, trustAccountRepo, caseRepo, userRepo)

		trustAccountHandler := handlers.NewTrustAccountHandler(trustAccountService, trustTransactionService)
		log.Println("✅ 代管款处理器初始化完成")

		// 代管款路由组
		trust := protected.Group("/trust")
		{
			// 账户管理
			trust.GET("/accounts", trustAccountHandler.ListAccounts)           // 获取账户列表
			trust.GET("/accounts/stats", trustAccountHandler.GetAccountStats)  // 获取账户统计
			trust.POST("/accounts", trustAccountHandler.CreateAccount)         // 创建账户
			trust.GET("/accounts/:id", trustAccountHandler.GetAccount)         // 获取账户详情
			trust.POST("/accounts/:id/freeze", trustAccountHandler.FreezeAccount)   // 冻结账户
			trust.POST("/accounts/:id/unfreeze", trustAccountHandler.UnfreezeAccount) // 解冻账户
			trust.POST("/accounts/:id/close", trustAccountHandler.CloseAccount)       // 关闭账户
			trust.GET("/accounts/:id/transactions", trustAccountHandler.GetAccountTransactions) // 获取账户交易记录

			// 交易管理
			trust.GET("/transactions", trustAccountHandler.ListTransactions)           // 获取交易列表
			trust.POST("/transactions", trustAccountHandler.CreateTransaction)         // 创建交易
			trust.GET("/transactions/:id", trustAccountHandler.GetTransaction)         // 获取交易详情
			trust.POST("/transactions/:id/approve", trustAccountHandler.ApproveTransaction) // 审批通过
			trust.POST("/transactions/:id/reject", trustAccountHandler.RejectTransaction)   // 审批拒绝

			// 统计
			trust.GET("/stats", trustAccountHandler.GetAccountStats) // 获取代管款统计
		}
		log.Println("✅ 代管款管理路由注册完成")

		log.Println("完整路由系统初始化完成")
}
}