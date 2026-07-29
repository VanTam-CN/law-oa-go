// Package main 示例律师事务所OA
// @title 示例律师事务所OA API
// @version 1.0
// @description 示例律师事务所OA 后端接口
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	_ "law-oa-go/docs" // Swagger docs
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/health"
	"law-oa-go/internal/metrics"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/router"
	"law-oa-go/internal/security"
)

func main() {
	if len(os.Args) > 1 && strings.EqualFold(strings.TrimSpace(os.Args[1]), "healthcheck") {
		if err := runHealthcheck(); err != nil {
			log.Printf("健康检查失败: %v", err)
			os.Exit(1)
		}
		return
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}
	if err := cfg.ValidateProductionReadiness(); err != nil {
		log.Fatal("生产环境门禁校验失败:", err)
	}

	// 初始化优化组件
	if err := database.InitOptimizedComponents(cfg); err != nil {
		log.Fatal("性能优化组件初始化失败:", err)
	}

	// 初始化全局缓存服务（用于中间件）- 使用数据库模块的缓存服务
	if cacheService := database.GetCacheService(); cacheService != nil {
		cache.DefaultCacheService = cacheService
		log.Println("使用数据库模块的缓存服务")
	} else {
		// 如果数据库模块的缓存服务不可用，则创建新的
		if err := cache.InitCache(); err != nil {
			log.Printf("缓存服务初始化失败，将禁用缓存功能: %v", err)
		} else {
			log.Println("缓存服务初始化成功")
		}
	}

	// 验证缓存服务是否真的初始化了
	if cache.DefaultCacheService == nil {
		log.Fatal("缓存服务初始化失败：DefaultCacheService 仍然为 nil")
	} else {
		log.Println("缓存服务验证成功")
	}

	// 初始化增强错误处理器
	var errorConfig errors.ErrorHandlerConfig
	if cfg.IsProduction() {
		errorConfig = errors.ProductionErrorHandlerConfig()
		log.Println("使用生产环境错误处理配置")
	} else {
		errorConfig = errors.DefaultErrorHandlerConfig()
		log.Println("使用开发环境错误处理配置")
	}

	// 创建错误处理器管理器
	errorManager := errors.NewErrorHandlerManager(slog.Default(), errorConfig)

	// 将错误处理器管理器设置到全局中间件中（供后续使用）
	middleware.SetErrorHandlerManager(errorManager)

	// 初始化JWT
	middleware.InitJWT(cfg)

	// 初始化监控服务
	monitorConfig := metrics.DefaultMonitorConfig
	if cfg.IsProduction() {
		monitorConfig.EnableRealTimeAlerts = true
		monitorConfig.MetricsCollectionInterval = 30 * time.Second
	} else {
		monitorConfig.EnableRealTimeAlerts = false
		monitorConfig.MetricsCollectionInterval = 10 * time.Second
	}

	if err := metrics.InitDefaultMonitorService(monitorConfig); err != nil {
		log.Fatal("监控服务初始化失败:", err)
	}

	// 获取监控服务实例
	monitorService := metrics.GetDefaultMonitorService()

	// 初始化健康检查系统
	healthConfig := &health.DefaultHealthConfig
	if cfg.IsProduction() {
		healthConfig.CheckInterval = 30 * time.Second
		healthConfig.FailureThreshold = 3
		healthConfig.EnableExternalAPICheck = true
		healthConfig.ElasticsearchTimeout = 5 * time.Second
	} else {
		healthConfig.CheckInterval = 15 * time.Second
		healthConfig.FailureThreshold = 2
		healthConfig.EnableExternalAPICheck = false
		healthConfig.ElasticsearchTimeout = 3 * time.Second
	}

	healthChecker := health.NewHealthChecker(healthConfig, slog.Default())

	// 注册健康检查
	if db := database.GetOptimizedDB(); db != nil && db.DB != nil {
		// 从GORM DB获取底层sql.DB
		if sqlDB, err := db.DB.DB(); err == nil {
			healthChecker.RegisterCheck(health.NewDatabaseHealthCheck(sqlDB, healthConfig.DatabaseTimeout))
			healthChecker.RegisterCheck(health.NewConflictP0ReadinessCheck(sqlDB, cfg.IsProduction(), cfg.Database.Driver))
		}
	}

	if cacheService := database.GetCacheService(); cacheService != nil {
		healthChecker.RegisterCheck(health.NewCacheHealthCheck(cacheService, healthConfig.CacheTimeout))
	}

	// 初始化Elasticsearch客户端
	var esClient *elasticsearch.Client

	esClient, err = database.InitElasticsearch(cfg.Elasticsearch)
	if err != nil {
		log.Printf("Elasticsearch初始化失败: %v, 将使用数据库搜索回退", err)
		esClient = nil
	} else {
		log.Println("Elasticsearch客户端初始化成功")
		// 注册ES健康检查
		healthChecker.RegisterCheck(health.NewElasticsearchHealthCheck(esClient, healthConfig.ElasticsearchTimeout))
	}

	// 注册并发服务检查（暂时注释，需要重构MonitorService）
	// if monitorService != nil && monitorService.concurrencyService != nil {
	// 	// 使用监控服务的并发服务（如果有）
	// 	healthChecker.RegisterCheck(health.NewConcurrencyHealthCheck(monitorService.concurrencyService, healthConfig.ConcurrencyTimeout))
	// }

	// 注册存储检查
	healthChecker.RegisterCheck(health.NewStorageHealthCheck(healthConfig.StoragePath, healthConfig.StorageTimeout))

	// 注册外部API检查（生产环境）
	if cfg.IsProduction() && healthConfig.EnableExternalAPICheck {
		healthChecker.RegisterCheck(health.NewExternalAPIHealthCheck(healthConfig.ExternalServiceURL, healthConfig.ExternalAPITimeout))
	}

	// 启动健康检查器
	healthChecker.Start()
	defer healthChecker.Stop()

	// 创建健康检查中间件
	healthMiddleware := health.NewHealthMiddleware(healthChecker, "1.0.0", cfg.Environment)

	// 设置 Gin 模式（按照最佳实践）
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
		// 生产环境禁用调试日志
		gin.DisableConsoleColor()
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建 Gin 引擎
	app := gin.New()

	// 应用核心中间件（按照Gin最佳实践顺序：先处理通用中间件，再处理业务逻辑）
	app.Use(middleware.RequestIDMiddleware()) // 请求ID追踪
	app.Use(middleware.LoggerWithFormatter()) // 使用格式化日志中间件
	app.Use(middleware.Recovery())            // 崩溃恢复
	app.Use(security.SecurityHeaders())       // 安全头
	app.Use(security.CORS())                  // 跨域设置
	app.Use(security.RateLimiterMiddleware()) // 限流控制

	// 注意：不使用 gin.Recovery()，因为自定义 middleware.Recovery() 已在上面注册
	// gin.Recovery() 会先捕获 panic 并返回空 500 body，覆盖自定义 Recovery 的 JSON 响应

	// 应用性能监控中间件
	app.Use(metrics.PrometheusMiddleware())

	// 应用缓存中间件
	app.Use(middleware.CacheMiddleware(middleware.CacheConfig{
		TTL:        5 * time.Minute,
		SkipHeader: "X-Cache-Skip",
	}))

	// 初始化路由系统
	db := database.GetOptimizedDB().DB
	if cfg.IsProduction() {
		if err := validateProductionSchema(db); err != nil {
			log.Fatal("生产数据库结构门禁失败:", err)
		}
	}
	router.Init(app,
		db,
		database.GetCacheService().GetClient(),
		database.GetElasticsearchClient())

	// 注册集成测试路由（仅开发环境）
	if !cfg.IsProduction() {
		handlers.RegisterTestRoutes(app, database.GetOptimizedDB().DB)
		log.Println("🔧 集成测试路由注册完成")
	}

	// Production schema changes must come from the one-shot bootstrap or an
	// reviewed migration. Runtime AutoMigrate is development-only so a new
	// application binary cannot silently mutate a live law-firm database.
	if !cfg.IsProduction() {
		// 自动迁移用户和权限模型。已由 SQL 迁移管理的表跳过运行时 AutoMigrate，避免 GORM 误判历史约束名。
		if migratedTablesExist(db, &models.User{}, &models.Role{}, &models.Permission{}, &models.RolePermission{}, &models.UserRole{}) {
			log.Println("用户/RBAC表已存在，跳过自动迁移")
		} else if err := db.AutoMigrate(
			&models.User{},
			&models.Role{},
			&models.Permission{},
			&models.RolePermission{},
			&models.UserRole{},
		); err != nil {
			log.Printf("用户/RBAC表自动迁移失败: %v", err)
		} else {
			log.Println("用户/RBAC表自动迁移成功")
		}
	} else {
		log.Println("生产环境已完成 schema bootstrap，跳过运行时自动迁移")
	}

	// 初始化RBAC数据
	if err := database.InitRBACData(db); err != nil {
		log.Printf("RBAC数据初始化失败: %v", err)
	} else {
		log.Println("RBAC数据初始化成功")
	}

	if !cfg.IsProduction() {
		// 自动迁移集成相关模型
		if err := db.AutoMigrate(
			&models.ConflictCheckAssociation{},
			&models.CaseCreationAssociation{},
			&models.ApprovalIntegrationMetadata{},
		); err != nil {
			log.Printf("集成模型自动迁移失败: %v", err)
		} else {
			log.Println("集成模型自动迁移成功")
		}

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
			log.Printf("财务模型自动迁移失败: %v", err)
		} else {
			log.Println("财务模型自动迁移成功")
		}

		// 自动迁移试用 MVP 必需模型
		if migratedTablesExist(db, &models.InboxItem{}, &models.ClientTrustAccount{}, &models.ClientTrustTransaction{}) {
			log.Println("MVP模型表已存在，跳过自动迁移")
		} else if err := db.AutoMigrate(
			&models.InboxItem{},
			&models.ClientTrustAccount{},
			&models.ClientTrustTransaction{},
		); err != nil {
			log.Printf("MVP模型自动迁移失败: %v", err)
		} else {
			log.Println("MVP模型自动迁移成功")
		}
	}

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
		log.Printf("财务模型自动迁移失败: %v", err)
	} else {
		log.Println("财务模型自动迁移成功")
	}

	// 自动迁移试用 MVP 必需模型
	if migratedTablesExist(db, &models.InboxItem{}, &models.ClientTrustAccount{}, &models.ClientTrustTransaction{}) {
		log.Println("MVP模型表已存在，跳过自动迁移")
	} else if err := db.AutoMigrate(
		&models.InboxItem{},
		&models.ClientTrustAccount{},
		&models.ClientTrustTransaction{},
	); err != nil {
		log.Printf("MVP模型自动迁移失败: %v", err)
	} else {
		log.Println("MVP模型自动迁移成功")
	}

	// 添加性能监控端点
	app.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 添加Swagger文档端点
	if !cfg.IsProduction() {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 添加基础健康检查端点
	app.GET("/health", healthMiddleware.HealthCheckHandler)
	app.GET("/health/live", healthMiddleware.LivenessHandler)
	app.GET("/health/ready", healthMiddleware.ReadinessHandler)

	// Detailed health responses include dependency names, check details and
	// runtime metadata. Keep them out of the public production surface; the
	// liveness/readiness endpoints above are sufficient for orchestration.
	if !cfg.IsProduction() {
		// 添加详细健康检查端点（/api/v1/health 已在 router.go 的 public 组中注册）
		app.GET("/health/detailed", healthMiddleware.DetailedHealthCheckHandler)
		app.GET("/health/metrics", healthMiddleware.HealthCheckMetricsHandler)
		app.GET("/health/dependencies", healthMiddleware.DependencyHealthHandler)
		app.GET("/health/history", healthMiddleware.HealthCheckHistoryHandler)

		// 添加健康状态页面
		app.GET("/health/status", healthMiddleware.HealthStatusPageHandler)

		// 添加健康指标导出端点
		app.GET("/metrics/health", healthMiddleware.ExportHealthMetricsHandler)
	}

	// Diagnostic and process-control endpoints are development-only. Production
	// exposes the authenticated health/metrics endpoints above instead of an
	// unauthenticated shutdown, GC, cache test, or internal monitor surface.
	if !cfg.IsProduction() {
		// 添加优雅关闭端点（仅开发环境）
		app.POST("/admin/shutdown", healthMiddleware.GracefulShutdownHandler)

		// 添加监控状态端点
		app.GET("/api/v1/monitor/status", func(c *gin.Context) {
			if monitorService != nil {
				status := monitorService.GetStatus()
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    status,
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"success": false,
					"error":   "Monitor service not available",
				})
			}
		})

		// 添加监控仪表板端点
		app.GET("/api/v1/monitor/dashboard", func(c *gin.Context) {
			if monitorService != nil {
				dashboard := monitorService.GetDashboardData()
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    dashboard,
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"success": false,
					"error":   "Monitor service not available",
				})
			}
		})

		// 添加性能统计端点
		app.GET("/api/v1/monitor/performance", func(c *gin.Context) {
			if monitorService != nil {
				stats := monitorService.GetPerformanceStats()
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    stats,
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"success": false,
					"error":   "Monitor service not available",
				})
			}
		})

		// 添加告警端点
		app.GET("/api/v1/monitor/alerts", func(c *gin.Context) {
			if monitorService != nil {
				alerts := monitorService.GetAlerts()
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    alerts,
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"success": false,
					"error":   "Monitor service not available",
				})
			}
		})

		// 添加解决告警端点
		app.POST("/api/v1/monitor/alerts/:id/resolve", func(c *gin.Context) {
			if monitorService == nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"success": false,
					"error":   "Monitor service not available",
				})
				return
			}

			alertID := c.Param("id")
			resolved := monitorService.ResolveAlert(alertID)

			if resolved {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Alert resolved successfully",
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"error":   "Alert not found",
				})
			}
		})

		// 添加强制GC端点
		app.POST("/api/v1/monitor/gc", func(c *gin.Context) {
			if monitorService != nil {
				monitorService.ForceGC()
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Garbage collection triggered successfully",
				})
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"success": false,
					"error":   "Monitor service not available",
				})
			}
		})

		// 添加性能测试端点
		app.GET("/performance/cache", func(c *gin.Context) {
			cacheService := database.GetCacheService()
			if cacheService == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Cache service not available"})
				return
			}

			// 简单的缓存性能测试
			start := time.Now()
			testData := map[string]interface{}{
				"id":    1,
				"name":  "性能测试数据",
				"value": "这是一个缓存性能测试",
			}

			key := "performance:test"

			// 测试设置
			if err := cacheService.Set(key, testData, time.Minute); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Cache set failed"})
				return
			}

			// 测试获取
			var result map[string]interface{}
			if err := cacheService.Get(key, &result); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Cache get failed"})
				return
			}

			duration := time.Since(start)
			c.JSON(http.StatusOK, gin.H{
				"duration_ms": duration.Milliseconds(),
				"cache_hit":   true,
				"data":        result,
			})
		})
	}

	// 优雅关闭服务器
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app,
		ReadHeaderTimeout: 20 * time.Second, // 防止 Slowloris 攻击
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在端口 %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器关闭失败:", err)
	}

	// 关闭数据库连接
	if err := database.Close(); err != nil {
		log.Fatal("数据库关闭失败:", err)
	}

	log.Println("服务器已关闭")
}

// runHealthcheck is used by the scratch production image. It must not load
// application configuration or mutate state; it only probes the already
// running local listener and returns a process exit status for the runtime.
// HEALTHCHECK_PATH lets Compose use strict readiness while the image default
// remains the lightweight liveness probe used by generic container runtimes.
func runHealthcheck() error {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	path := strings.TrimSpace(os.Getenv("HEALTHCHECK_PATH"))
	if path == "" {
		path = "/health/live"
	}
	if !strings.HasPrefix(path, "/health/") {
		return fmt.Errorf("HEALTHCHECK_PATH must be a local health path")
	}
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:"+port+path, nil)
	if err != nil {
		return fmt.Errorf("构造健康检查请求失败: %w", err)
	}
	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("访问本地存活探针失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("本地存活探针返回 HTTP %d", response.StatusCode)
	}
	return nil
}

func migratedTablesExist(db *gorm.DB, models ...interface{}) bool {
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			return false
		}
	}
	return true
}

func validateProductionSchema(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	requiredTables := []string{
		"users",
		"roles",
		"permissions",
		"role_permissions",
		"user_roles",
		"clients",
		"cases",
		"entities",
		"entity_relations",
		"entity_name_history",
		"case_parties",
		"conflict_checks",
		"conflict_details",
		"conflict_rules",
		"conflict_check_records",
		"conflict_cases",
		"conflict_reviews",
		"conflict_reviewer_assignments",
		"conflict_search_scopes",
		"case_subject_revisions",
		"compliance_audit_events",
		"approval_requests",
		"approval_workflows",
		"approval_records",
		"approval_nodes",
		"approval_notifications",
		"approval_delegations",
		"approval_snapshots",
		"approval_conflict_associations",
		"approval_case_creation_tracking",
		"case_intakes",
		"case_intake_parties",
		"case_materials",
		"schema_bootstrap_state",
	}

	missing := make([]string, 0)
	for _, table := range requiredTables {
		if !db.Migrator().HasTable(table) {
			missing = append(missing, table)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少关键表: %s；请先完成目标数据库迁移并重新启动", strings.Join(missing, ", "))
	}

	var schemaVersion string
	if err := db.Table("schema_bootstrap_state").Where("id = ?", 1).Pluck("version", &schemaVersion).Error; err != nil {
		return fmt.Errorf("无法读取生产 schema bootstrap 版本: %w", err)
	}
	if schemaVersion != database.ProductionSchemaVersion {
		return fmt.Errorf("生产 schema 版本不受支持: %s；期望 %s", schemaVersion, database.ProductionSchemaVersion)
	}
	if err := database.ValidateProductionSchemaContract(db); err != nil {
		return err
	}
	return nil
}
