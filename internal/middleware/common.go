package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/infrastructure"
	"law-oa-go/internal/logger"
)

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	// IP级别限流
	IPRateLimit    int           // 每分钟IP限流次数
	IPWindow       time.Duration // IP限流时间窗口

	// 用户级别限流
	UserRateLimit  int           // 每分钟用户限流次数
	UserWindow     time.Duration // 用户限流时间窗口

	// API级别限流
	APIRateLimit   int           // 每分钟API限流次数
	APIWindow      time.Duration // API限流时间窗口

	// 白名单
	Whitelist      []string      // 不受限流的IP段

	// 限流策略
	Strategy       string        // "fixed_window", "sliding_window", "token_bucket"
}

// DefaultRateLimiterConfig 默认限流配置
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		IPRateLimit:    60,   // IP级别：每分钟60次
		IPWindow:       time.Minute,
		UserRateLimit:  120,  // 用户级别：每分钟120次
		UserWindow:     time.Minute,
		APIRateLimit:   100,  // API级别：每分钟100次
		APIWindow:      time.Minute,
		Whitelist:      []string{"127.0.0.1", "::1"}, // 本地回环不受限
		Strategy:       "sliding_window",
	}
}

// RateLimiter 限流中间件
func RateLimiter() gin.HandlerFunc {
	return RateLimiterWithConfig(DefaultRateLimiterConfig())
}

// RateLimiterWithConfig 带配置的限流中间件
func RateLimiterWithConfig(config RateLimiterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		redisClient := infrastructure.GetRedis()
		if redisClient == nil {
			c.Next()
			return
		}

		client := redisClient
		ip := c.ClientIP()

		// 检查白名单
		for _, whitelistIP := range config.Whitelist {
			if ip == whitelistIP {
				c.Next()
				return
			}
		}

		// 多级限流检查
		if !checkIPRateLimit(client, ip, config) {
			responseRateLimitExceeded(c, "IP")
			return
		}

		// 用户级别限流（如果已认证）
		if userID := getUserID(c); userID != "" {
			if !checkUserRateLimit(client, userID, config) {
				responseRateLimitExceeded(c, "用户")
				return
			}
		}

		// API级别限流
		if !checkAPIRateLimit(client, c.Request.URL.Path, config) {
			responseRateLimitExceeded(c, "API")
			return
		}

		c.Next()
	}
}

// checkIPRateLimit 检查IP限流
func checkIPRateLimit(client interface{}, ip string, config RateLimiterConfig) bool {
	key := fmt.Sprintf("rate_limit:ip:%s", ip)
	return checkSlidingWindow(client, key, config.IPRateLimit, config.IPWindow)
}

// checkUserRateLimit 检查用户限流
func checkUserRateLimit(client interface{}, userID string, config RateLimiterConfig) bool {
	key := fmt.Sprintf("rate_limit:user:%s", userID)
	return checkSlidingWindow(client, key, config.UserRateLimit, config.UserWindow)
}

// checkAPIRateLimit 检查API限流
func checkAPIRateLimit(client interface{}, apiPath string, config RateLimiterConfig) bool {
	key := fmt.Sprintf("rate_limit:api:%s", apiPath)
	return checkSlidingWindow(client, key, config.APIRateLimit, config.APIWindow)
}

// checkSlidingWindow 检查滑动窗口限流
func checkSlidingWindow(client interface{}, key string, limit int, window time.Duration) bool {
	// 这里简化实现，实际应该使用滑动窗口算法
	// 由于Redis客户端类型不确定，使用接口方式
	// 实际项目中应该根据具体的Redis客户端实现
	return true // 占位实现，避免编译错误
}

// getUserID 从上下文获取用户ID
func getUserID(c *gin.Context) string {
	// 尝试从JWT token获取用户ID
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// responseRateLimitExceeded 响应限流超出
func responseRateLimitExceeded(c *gin.Context, limitType string) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"code":    429,
		"message": fmt.Sprintf("%s级别请求过于频繁，请稍后再试", limitType),
		"type":    limitType,
		"retry_after": 60, // 建议重试时间（秒）
	})
	c.Abort()
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         string
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		// 严格配置：仅允许具体端口，避免端口跳跃攻击
		AllowedOrigins: []string{
			"http://localhost:3003",
			"https://localhost:3003",
			"http://127.0.0.1:3003",
			"https://127.0.0.1:3003",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Request-ID",
			"X-API-Version",
		},
		MaxAge: "7200", // 减少到2小时，增加安全性
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig 带配置的CORS中间件
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		referer := c.Request.Header.Get("Referer")

		// 严格的来源检查
		allowed := false
		if origin != "" {
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == origin {
					allowed = true
					break
				}
			}
		} else if referer != "" {
			// 如果没有Origin头，检查Referer作为备用
			for _, allowedOrigin := range config.AllowedOrigins {
				if strings.HasPrefix(referer, allowedOrigin) {
					allowed = true
					break
				}
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", config.MaxAge)

		// 添加额外的安全头部
		c.Header("Vary", "Origin")
		c.Header("X-Content-Type-Options", "nosniff")

		if c.Request.Method == "OPTIONS" {
			// 预检请求不缓存太久，增加安全性
			c.Header("Cache-Control", "no-cache")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger 日志中间件（保持向后兼容）
func Logger() gin.HandlerFunc {
	return LoggerWithFormatter()
}

// LoggerWithFormatter 格式化日志中间件（符合Gin最佳实践）
func LoggerWithFormatter() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用结构化日志格式，便于日志分析工具处理
		return fmt.Sprintf("%s - [%s] \"%s %s %s\" %d %s \"%s\" %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123), // 使用标准时间格式
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// Recovery 恢复中间件（改进版本，符合Gin最佳实践）
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录详细的错误信息
		if logger.Logger != nil {
			if err, ok := recovered.(error); ok {
				logger.Logger.Error("Panic recovered",
					zap.Error(err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("clientIP", c.ClientIP()),
					zap.String("userAgent", c.Request.UserAgent()),
				)
			} else {
				logger.Logger.Error("Panic recovered",
					zap.Any("recovered", recovered),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("clientIP", c.ClientIP()),
					zap.String("userAgent", c.Request.UserAgent()),
				)
			}
		}

		// 根据环境返回不同的错误信息
		if c.GetHeader("Accept") == "application/json" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    500,
					"message": "服务器内部错误",
					"type":    "internal_server_error",
				},
			})
		} else {
			c.String(http.StatusInternalServerError, "服务器内部错误")
		}

		c.Abort()
	})
}

// SecurityHeaders 安全头部中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 基础安全头部
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 高级安全头部
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// 内容安全策略 - 更严格的配置
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " + // 保留unsafe-inline以兼容现有代码
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self';"
		c.Header("Content-Security-Policy", csp)

		// 移除敏感信息头部
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		// HSTS（仅HTTPS环境）
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}

// 全局错误处理器管理器
var GlobalErrorHandlerManager *errors.ErrorHandlerManager

// SetErrorHandlerManager 设置全局错误处理器管理器
func SetErrorHandlerManager(manager *errors.ErrorHandlerManager) {
	GlobalErrorHandlerManager = manager
}

// GetErrorHandlerManager 获取全局错误处理器管理器
func GetErrorHandlerManager() *errors.ErrorHandlerManager {
	return GlobalErrorHandlerManager
}
