package auth

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"law-oa-go/internal/cache"
	"net/http"
	"strings"
	"time"
)

var (
	authMiddlewareDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "auth_middleware_duration_seconds",
		Help:    "Duration of auth middleware operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})
	
	authMiddlewareErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_middleware_errors_total",
		Help: "Total number of auth middleware errors",
	}, []string{"operation", "type"})
	
	authAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_attempts_total",
		Help: "Total number of authentication attempts",
	}, []string{"result"})
)

// AuthConfig 认证配置
type AuthConfig struct {
	TokenManager      *TokenManager
	CacheService      *cache.CacheService
	SkipAuthPaths     []string
	SkipAuthPrefixes  []string
	RequiredRoles     []string
	EnableRateLimit   bool
	EnableDeviceCheck bool
	EnableIPCheck     bool
}

// EnhancedAuthMiddleware 增强的认证中间件
func EnhancedAuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			authMiddlewareDuration.WithLabelValues("enhanced_auth").Observe(time.Since(start).Seconds())
		}()
		
		// 检查是否跳过认证
		if shouldSkipAuth(c, config.SkipAuthPaths, config.SkipAuthPrefixes) {
			c.Next()
			return
		}
		
		// 获取认证头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authAttempts.WithLabelValues("missing_token").Inc()
			authMiddlewareErrors.WithLabelValues("auth", "missing_token").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证令牌",
				"error":   "Authorization header is required",
			})
			c.Abort()
			return
		}
		
		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			authAttempts.WithLabelValues("invalid_format").Inc()
			authMiddlewareErrors.WithLabelValues("auth", "invalid_format").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的令牌格式",
				"error":   "Bearer token is required",
			})
			c.Abort()
			return
		}
		
		tokenString := authHeader[7:] // 去掉"Bearer "前缀
		
		// 检查令牌是否在黑名单中
		if config.TokenManager.IsTokenBlacklisted(c.Request.Context(), tokenString) {
			authAttempts.WithLabelValues("blacklisted").Inc()
			authMiddlewareErrors.WithLabelValues("auth", "blacklisted").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "令牌已被撤销",
				"error":   "Token has been revoked",
			})
			c.Abort()
			return
		}
		
		// 验证令牌
		payload, err := config.TokenManager.ValidateAccess(c.Request.Context(), tokenString, config.RequiredRoles)
		if err != nil {
			authAttempts.WithLabelValues("invalid_token").Inc()
			authMiddlewareErrors.WithLabelValues("auth", "invalid_token").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的令牌",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}
		
		// 设备检查
		if config.EnableDeviceCheck {
			if !validateDevice(c, payload, config.CacheService) {
				authAttempts.WithLabelValues("device_mismatch").Inc()
				authMiddlewareErrors.WithLabelValues("auth", "device_mismatch").Inc()
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "设备验证失败",
					"error":   "Device validation failed",
				})
				c.Abort()
				return
			}
		}
		
		// IP地址检查
		if config.EnableIPCheck {
			if !validateIP(c, payload, config.CacheService) {
				authAttempts.WithLabelValues("ip_mismatch").Inc()
				authMiddlewareErrors.WithLabelValues("auth", "ip_mismatch").Inc()
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "IP地址验证失败",
					"error":   "IP validation failed",
				})
				c.Abort()
				return
			}
		}
		
		// 速率限制
		if config.EnableRateLimit {
			if !checkRateLimit(c, payload, config.CacheService) {
				authAttempts.WithLabelValues("rate_limited").Inc()
				authMiddlewareErrors.WithLabelValues("auth", "rate_limited").Inc()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"code":    429,
					"message": "请求过于频繁",
					"error":   "Rate limit exceeded",
				})
				c.Abort()
				return
			}
		}
		
		// 将用户信息存储到上下文中
		c.Set("user_id", payload.UserID)
		c.Set("username", payload.Username)
		c.Set("email", payload.Email)
		c.Set("role", payload.Role)
		c.Set("device_id", payload.DeviceID)
		c.Set("ip", payload.IP)
		c.Set("user_agent", payload.UserAgent)
		c.Set("token_payload", payload)
		
		// 更新用户最后活动时间
		go updateUserActivity(context.Background(), payload, config.CacheService)
		
		authAttempts.WithLabelValues("success").Inc()
		c.Next()
	}
}

// RefreshTokenMiddleware 刷新令牌中间件
func RefreshTokenMiddleware(tokenManager *TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			authMiddlewareDuration.WithLabelValues("refresh_token").Observe(time.Since(start).Seconds())
		}()
		
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			authMiddlewareErrors.WithLabelValues("refresh", "invalid_request").Inc()
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "请求参数错误",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}
		
		tokenDetails, err := tokenManager.RefreshTokens(c.Request.Context(), req.RefreshToken)
		if err != nil {
			authMiddlewareErrors.WithLabelValues("refresh", "failed").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "刷新令牌失败",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "令牌刷新成功",
			"data": gin.H{
				"access_token":  tokenDetails.AccessToken,
				"refresh_token": tokenDetails.RefreshToken,
				"expires_in":    tokenDetails.AtExpires,
			},
		})
	}
}

// RoleBasedAuthMiddleware 基于角色的认证中间件
func RoleBasedAuthMiddleware(tokenManager *TokenManager, requiredRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			authMiddlewareDuration.WithLabelValues("role_based_auth").Observe(time.Since(start).Seconds())
		}()
		
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			authMiddlewareErrors.WithLabelValues("role_auth", "missing_token").Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证令牌",
			})
			c.Abort()
			return
		}
		
		tokenString := authHeader[7:]
		
		payload, err := tokenManager.ValidateAccess(c.Request.Context(), tokenString, requiredRoles)
		if err != nil {
			authMiddlewareErrors.WithLabelValues("role_auth", "permission_denied").Inc()
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}
		
		c.Set("user_id", payload.UserID)
		c.Set("username", payload.Username)
		c.Set("role", payload.Role)
		c.Set("token_payload", payload)
		
		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头部中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 安全相关HTTP头部
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("X-Download-Options", "noopen")
		c.Header("X-Content-Security-Policy", "default-src 'self'")
		
		// 移除敏感头部
		c.Header("Server", "")
		c.Header("X-Powered-By", "")
		
		c.Next()
	}
}

// CSRFProtectionMiddleware CSRF保护中间件
func CSRFProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对于GET、HEAD、OPTIONS请求，跳过CSRF检查
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		
		// 检查CSRF Token
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.PostForm("csrf_token")
		}
		
		if csrfToken == "" {
			authMiddlewareErrors.WithLabelValues("csrf", "missing_token").Inc()
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "CSRF令牌缺失",
				"error":   "CSRF token is required",
			})
			c.Abort()
			return
		}
		
		// 从会话中获取CSRF令牌并验证
		sessionCSRF, exists := c.Get("csrf_token")
		if !exists || sessionCSRF != csrfToken {
			authMiddlewareErrors.WithLabelValues("csrf", "invalid_token").Inc()
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "CSRF令牌无效",
				"error":   "Invalid CSRF token",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// shouldSkipAuth 检查是否跳过认证
func shouldSkipAuth(c *gin.Context, skipPaths, skipPrefixes []string) bool {
	path := c.Request.URL.Path
	
	// 检查精确匹配
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	
	// 检查前缀匹配
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	
	return false
}

// validateDevice 验证设备信息
func validateDevice(c *gin.Context, payload *TokenPayload, cacheService *cache.CacheService) bool {
	if payload.DeviceID == "" {
		return true // 如果没有设备ID，跳过验证
	}
	
	deviceKey := fmt.Sprintf("user_device:%d:%s", payload.UserID, payload.DeviceID)
	var deviceInfo map[string]interface{}
	
	err := cacheService.Get(c.Request.Context(), deviceKey, &deviceInfo)
	if err != nil {
		return false
	}
	
	// 验证设备信息
	if storedDeviceID, ok := deviceInfo["device_id"].(string); ok && storedDeviceID != payload.DeviceID {
		return false
	}
	
	return true
}

// validateIP 验证IP地址
func validateIP(c *gin.Context, payload *TokenPayload, cacheService *cache.CacheService) bool {
	if payload.IP == "" {
		return true // 如果没有IP信息，跳过验证
	}
	
	currentIP := c.ClientIP()
	if payload.IP != currentIP {
		// 检查是否在同一子网（可选的宽松验证）
		// 这里可以根据需要实现更复杂的IP验证逻辑
		return false
	}
	
	return true
}

// checkRateLimit 检查速率限制
func checkRateLimit(c *gin.Context, payload *TokenPayload, cacheService *cache.CacheService) bool {
	rateLimitKey := fmt.Sprintf("rate_limit:%d:%s", payload.UserID, c.ClientIP())
	
	var count int
	err := cacheService.Get(c.Request.Context(), rateLimitKey, &count)
	if err != nil {
		count = 0
	}
	
	// 简单的速率限制：每分钟最多100次请求
	if count >= 100 {
		return false
	}
	
	// 增加计数
	cacheService.Set(c.Request.Context(), rateLimitKey, count+1, time.Minute)
	
	return true
}

// updateUserActivity 更新用户活动信息
func updateUserActivity(ctx context.Context, payload *TokenPayload, cacheService *cache.CacheService) {
	deviceKey := fmt.Sprintf("user_device:%d:%s", payload.UserID, payload.DeviceID)
	
	var deviceInfo map[string]interface{}
	err := cacheService.Get(ctx, deviceKey, &deviceInfo)
	if err == nil {
		deviceInfo["last_active"] = time.Now()
		cacheService.Set(ctx, deviceKey, deviceInfo, time.Hour*24)
	}
}