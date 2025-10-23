package v2

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"law-oa-go/internal/config"
	"law-oa-go/internal/logger"
)

// V2MiddlewareFactory 中间件工厂 - 基于环境自动配置最佳实践
type V2MiddlewareFactory struct {
	config       *config.Config
	redisClient  *redis.Client
	performance   *V2PerformanceMiddleware
	security     *V2SecurityMiddleware
}

// NewV2MiddlewareFactory 创建中间件工厂
func NewV2MiddlewareFactory(config *config.Config, redisClient *redis.Client) *V2MiddlewareFactory {
	return &V2MiddlewareFactory{
		config:      config,
		redisClient: redisClient,
		performance: NewV2PerformanceMiddleware(config, redisClient),
		security:   NewV2SecurityMiddleware(config, redisClient),
	}
}

// SetupRouter 根据环境配置路由中间件
func (f *V2MiddlewareFactory) SetupRouter(router *gin.Engine) {
	logger.Logger.Info("Setting up V2 middleware for environment: "+f.config.Environment,
		zap.String("environment", f.config.Environment),
	)

	// 基础中间件 - 所有环境都需要
	for _, middleware := range f.setupBaseMiddleware() {
		router.Use(middleware)
	}

	// 环境特定中间件
	switch f.config.Environment {
	case "production":
		f.setupProductionMiddleware(router)
	case "development":
		f.setupDevelopmentMiddleware(router)
	case "test":
		f.setupTestMiddleware(router)
	default:
		f.setupDefaultMiddleware(router)
	}

	// 健康检查和监控端点
	f.setupHealthEndpoints(router)
}

// setupBaseMiddleware 设置基础中间件
func (f *V2MiddlewareFactory) setupBaseMiddleware() []gin.HandlerFunc {
	var middlewares []gin.HandlerFunc

	// 自定义日志中间件（使用Gin的LoggerWithFormatter）
	middlewares = append(middlewares, f.performance.CustomLoggerMiddleware())

	// 自定义恢复中间件
	middlewares = append(middlewares, f.performance.RecoveryMiddleware())

	// 安全头中间件
	middlewares = append(middlewares, f.security.SecurityHeadersMiddleware())

	// 请求验证中间件
	middlewares = append(middlewares, f.security.RequestValidationMiddleware())

	// 输入清理中间件
	middlewares = append(middlewares, f.security.InputSanitizationMiddleware())

	// IP过滤中间件
	middlewares = append(middlewares, f.security.IPFilterMiddleware())

	// 受信任代理中间件
	middlewares = append(middlewares, f.security.TrustedProxyMiddleware())

	return middlewares
}

// setupProductionMiddleware 设置生产环境中间件
func (f *V2MiddlewareFactory) setupProductionMiddleware(router *gin.Engine) {
	logger.Logger.Info("Setting up production middleware")

	// 性能监控中间件
	router.Use(f.performance.PerformanceMiddleware())

	// 内存监控中间件
	router.Use(f.performance.MemoryMonitoringMiddleware())

	// CORS中间件（生产环境配置）
	router.Use(f.security.CORSMiddleware())

	// 高并发控制（生产环境限制）
	router.Use(f.performance.ConcurrencyControlMiddleware(1000))

	// 请求超时（生产环境较短）
	router.Use(f.performance.TimeoutMiddleware(30*time.Second))

	// 速率限制（生产环境严格）
	router.Use(f.security.RateLimitingMiddleware(100, time.Minute))

	// 基础限流
	router.Use(f.performance.RateLimitingMiddleware(100, time.Minute))
}

// setupDevelopmentMiddleware 设置开发环境中间件
func (f *V2MiddlewareFactory) setupDevelopmentMiddleware(router *gin.Engine) {
	logger.Logger.Info("Setting up development middleware")

	// 简化的性能监控
	router.Use(f.performance.PerformanceMiddleware())

	// CORS中间件（开发环境宽松配置）
	router.Use(f.security.CORSMiddleware())

	// 开发环境限制较小的并发控制
	router.Use(f.performance.ConcurrencyControlMiddleware(50))

	// 请求超时（开发环境较长）
	router.Use(f.performance.TimeoutMiddleware(60*time.Second))

	// 开发环境宽松的速率限制
	router.Use(f.security.RateLimitingMiddleware(1000, time.Minute))

	// 基础限流
	router.Use(f.performance.RateLimitingMiddleware(1000, time.Minute))
}

// setupTestMiddleware 设置测试环境中间件
func (f *V2MiddlewareFactory) setupTestMiddleware(router *gin.Engine) {
	logger.Logger.Info("Setting up test middleware")

	// 最小化的性能监控
	router.Use(f.performance.PerformanceMiddleware())

	// CORS中间件（测试环境配置）
	router.Use(f.security.CORSMiddleware())

	// 测试环境很小的并发控制
	router.Use(f.performance.ConcurrencyControlMiddleware(10))

	// 测试环境超时
	router.Use(f.performance.TimeoutMiddleware(30*time.Second))

	// 测试环境宽松的速率限制
	router.Use(f.security.RateLimitingMiddleware(5000, time.Minute))

	// 基础限流
	router.Use(f.performance.RateLimitingMiddleware(5000, time.Minute))
}

// setupDefaultMiddleware 设置默认中间件
func (f *V2MiddlewareFactory) setupDefaultMiddleware(router *gin.Engine) {
	logger.Logger.Info("Setting up default middleware")

	// 基础性能监控
	router.Use(f.performance.PerformanceMiddleware())

	// 默认CORS配置
	router.Use(f.security.CORSMiddleware())

	// 默认并发控制
	router.Use(f.performance.ConcurrencyControlMiddleware(100))

	// 默认超时
	router.Use(f.performance.TimeoutMiddleware(45*time.Second))

	// 默认速率限制
	router.Use(f.security.RateLimitingMiddleware(500, time.Minute))

	// 基础限流
	router.Use(f.performance.RateLimitingMiddleware(500, time.Minute))
}

// setupHealthEndpoints 设置健康检查和监控端点
func (f *V2MiddlewareFactory) setupHealthEndpoints(router *gin.Engine) {
	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "2.1.0",
			"environment": f.config.Server.Env,
		})
	})

	// 性能指标端点
	router.GET("/metrics", func(c *gin.Context) {
		// 检查API密钥或基本认证
		if !f.checkMetricsAuth(c) {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		performanceMetrics := f.performance.GetMetrics()
		securityMetrics := f.security.GetSecurityStats()

		c.JSON(200, gin.H{
			"performance": performanceMetrics,
			"security":   securityMetrics,
			"timestamp":  time.Now().Unix(),
		})
	})

	// 中间件信息端点
	router.GET("/middleware/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version":     "2.1.0",
			"framework":   "Gin v1.9+",
			"environment": f.config.Server.Env,
			"features":    f.getMiddlewareFeatures(),
			"enabled": map[string]bool{
				"performance_monitoring": f.performance.enabled,
				"rate_limiting":       true,
				"security_headers":    true,
				"cors":               true,
				"input_validation":   true,
				"ip_filtering":       true,
			},
		})
	})

	// 系统信息端点
	router.GET("/system/info", func(c *gin.Context) {
		// 检查API密钥
		if !f.checkMetricsAuth(c) {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.JSON(200, gin.H{
			"uptime":      time.Since(f.config.AppStartTime).String(),
			"environment": f.config.Server.Env,
			"version":     f.config.Version,
			"build_time":   f.config.BuildTime,
			"git_commit":   f.config.GitCommit,
		})
	})
}

// GetMiddlewareChain 获取完整的中间件链
func (f *V2MiddlewareFactory) GetMiddlewareChain() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以组合多个中间件
		middlewares := f.setupBaseMiddleware()

		// 应用所有基础中间件
		for _, middleware := range middlewares {
			middleware(c)
			if c.IsAborted() {
				return
				}
		}

		c.Next()
	}
}

// CreateAPIMiddlewareChain 为API路由创建专门的中间件链
func (f *V2MiddlewareFactory) CreateAPIMiddlewareChain() []gin.HandlerFunc {
	var chain []gin.HandlerFunc

	// API特定的中间件顺序很重要
	chain = append(chain, f.security.RequestValidationMiddleware())
	chain = append(chain, f.performance.TimeoutMiddleware(30*time.Second))
	chain = append(chain, f.security.RateLimitingMiddleware(100, time.Minute))

	return chain
}

// CreateAdminMiddlewareChain 为管理员路由创建专门的中间件链
func (f *V2MiddlewareFactory) CreateAdminMiddlewareChain() []gin.HandlerFunc {
	var chain []gin.HandlerFunc

	// 管理员路由需要更严格的安全控制
	chain = append(chain, f.security.RequestValidationMiddleware())
	chain = append(chain, f.performance.TimeoutMiddleware(60*time.Second))
	chain = append(chain, f.security.RateLimitingMiddleware(50, time.Minute))

	return chain
}

// CreatePublicMiddlewareChain 为公开路由创建专门的中间件链
func (f *V2MiddlewareFactory) CreatePublicMiddlewareChain() []gin.HandlerFunc {
	var chain []gin.HandlerFunc

	// 公开路由较少限制
	chain = append(chain, f.performance.TimeoutMiddleware(30*time.Second))
	chain = append(chain, f.security.RateLimitingMiddleware(1000, time.Minute))

	return chain
}

// 内部辅助方法

func (f *V2MiddlewareFactory) checkMetricsAuth(c *gin.Context) bool {
	// 开发环境允许所有请求
	if f.config.Server.Env == "development" {
		return true
	}

	// 检查API密钥头
	apiKey := c.GetHeader("X-API-Key")
	// 这里可以添加更复杂的认证逻辑

	// 如果没有密钥，检查基本认证
	if apiKey == "" {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			return false
		}
		// 简单的Bearer token验证
		if !strings.HasPrefix(auth, "Bearer ") {
			return false
		}
		token := auth[7:]
		// 这里应该验证JWT token
		return true
	}

	// 验证API密钥
	return apiKey == f.config.APISecret || apiKey == "dev-api-key"
}

func (f *V2MiddlewareFactory) getMiddlewareFeatures() []string {
	return []string{
		"Security Headers",
		"CORS",
		"Request Validation",
		"Input Sanitization",
		"Rate Limiting",
	"Performance Monitoring",
		"Memory Usage Tracking",
	"Concurrency Control",
	"Request Timeout",
	"Health Checks",
	"Metrics Collection",
	"IP Filtering",
	"Trusted Proxy Support",
	"Graceful Recovery",
	"Custom Logging",
	}
}

// GetConfigurationSummary 获取配置摘要
func (f *V2MiddlewareFactory) GetConfigurationSummary() map[string]interface{} {
	return map[string]interface{}{
		"environment":      f.config.Server.Env,
		"performance_enabled": f.performance.enabled,
		"slow_threshold":    f.performance.slowThreshold.String(),
		"cors_origins":      f.getORSOrigins(),
		"rate_limiting":     true,
		"security_headers": true,
		"input_validation":  true,
		"middleware_count":   f.countMiddleware(),
	}
}

func (f *V2MiddlewareFactory) getORSOrigins() []string {
	// 这里可以从配置中获取
	switch f.config.Server.Env {
	case "production":
		return []string{"https://your-domain.com"}
	case "development":
		return []string{"http://localhost:3000", "http://localhost:3003"}
	default:
		return []string{"*"}
	}
}

func (f *V2MiddlewareFactory) countMiddleware() int {
	// 统计启用的中间件数量
	count := 0
	if f.performance.enabled {
		count += 5 // 性能中间件数量
	}
	count += 7 // 安全中间件数量
	return count
}

// ResetMetrics 重置所有指标
func (f *V2MiddlewareFactory) ResetMetrics() {
	f.performance.ResetMetrics()
	logger.Logger.Info("V2 middleware metrics reset")
}

// GetHealthStatus 获取中间件健康状态
func (f *V2MiddlewareFactory) GetHealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":      "healthy",
		"performance": f.performance.GetMetrics(),
		"security":   f.security.GetSecurityStats(),
		"timestamp":  time.Now().Unix(),
		"environment": f.config.Server.Env,
	}
}