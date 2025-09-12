// Package main Law Office Automation System
// @title Law Office Automation API
// @version 1.0
// @description 法律事务所自动化管理系统 API
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
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/router"
	_ "law-oa-go/docs" // Swagger docs
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化优化组件
	if err := database.InitOptimizedComponents(cfg); err != nil {
		log.Fatal("性能优化组件初始化失败:", err)
	}

	// 初始化JWT
	middleware.InitJWT(cfg)

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

	// 应用核心中间件（按照最佳实践顺序）
	app.Use(middleware.RequestID())           // 请求ID追踪
	app.Use(middleware.Logger())             // 日志记录
	app.Use(middleware.Recovery())           // 崩溃恢复
	app.Use(middleware.SecurityHeaders())    // 安全头
	app.Use(middleware.RequestTimeout(30*time.Second)) // 请求超时
	app.Use(middleware.CORS())               // 跨域设置
	app.Use(middleware.RateLimiter())        // 限流控制
	
	// 应用性能监控中间件
	app.Use(middleware.PrometheusMiddleware())
	
	// 应用缓存中间件
	app.Use(middleware.CacheMiddleware(middleware.CacheConfig{
		TTL:          5 * time.Minute,
		KeyGenerator: middleware.DefaultKeyGenerator,
		ShouldCache:  middleware.DefaultShouldCache,
		SkipHeader:   "X-Cache-Skip",
	}))

	// 初始化路由
	router.Init(app, 
		database.GetOptimizedDB().DB, 
		database.GetCacheService().GetClient(), 
		database.GetElasticsearchClient())

	// 添加性能监控端点
	app.GET("/metrics", gin.WrapH(promhttp.Handler()))
	
	// 添加Swagger文档端点
	if !cfg.IsProduction() {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	
	// 添加健康检查端点
	app.GET("/health", func(c *gin.Context) {
		health := database.Health()
		c.JSON(http.StatusOK, health)
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
		
		ctx := c.Request.Context()
		key := "performance:test"
		
		// 测试设置
		if err := cacheService.Set(ctx, key, testData, time.Minute); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cache set failed"})
			return
		}
		
		// 测试获取
		var result map[string]interface{}
		if err := cacheService.Get(ctx, key, &result); err != nil {
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

	// 优雅关闭服务器
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: app,
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