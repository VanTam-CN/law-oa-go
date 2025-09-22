package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	jwtManager     *security.JWTKeyManager
	rateLimiter    *security.RateLimiter
	config         *security.SecurityConfig
	cacheService   *cache.CacheService
	redisClient    *redis.Client
	securityConfig *config.Config
}

// NewSecurityMiddleware 创建新的安全中间件
func NewSecurityMiddleware(
	securityConfig *config.Config,
	cacheService *cache.CacheService,
	redisClient *redis.Client,
) *SecurityMiddleware {
	// 创建安全配置
	secConfig := createDefaultSecurityConfig(securityConfig)

	// 创建JWT密钥管理器
	jwtManager := security.NewJWTKeyManager(securityConfig, secConfig, redisClient, cacheService)

	// 创建请求限制器
	rateLimiter := security.NewRateLimiter(redisClient, &secConfig.APISecurity)

	// 启动清理定时器
	rateLimiter.StartCleanupTimer()

	return &SecurityMiddleware{
		jwtManager:     jwtManager,
		rateLimiter:    rateLimiter,
		config:         secConfig,
		cacheService:   cacheService,
		redisClient:    redisClient,
		securityConfig: securityConfig,
	}
}

// createDefaultSecurityConfig 创建默认安全配置
func createDefaultSecurityConfig(cfg *config.Config) *security.SecurityConfig {
	return &security.SecurityConfig{
		JWT: security.JWTConfig{
			AccessTokenTTL:  time.Duration(cfg.JWT.ExpiresIn) * time.Second,
			RefreshTokenTTL: time.Duration(cfg.JWT.RefreshIn) * time.Second,
			Issuer:          "law-oa-system",
			Secret:          cfg.JWT.Secret,
			EnableRefresh:   true,
			BlacklistTTL:    24 * time.Hour,
		},
		APISecurity: security.APISecurityConfig{
			EnableRateLimit:         true,
			EnableIPWhitelist:       false,
			EnableIPBlacklist:       true,
			EnableRequestSigning:    false,
			EnableAPIThrottling:     true,
			EnableWAFProtection:     true,
			EnableDDoSProtection:    true,
			EnableRequestValidation: true,
			EnableCORS:              true,
			EnableCSRF:              true,
			RateLimitWindow:         time.Minute,
			RateLimitMaxRequests:    100,
			WhitelistedIPs:          []string{},
			BlacklistedIPs:          []string{},
			AllowedOrigins:          []string{"http://localhost:3000"},
			AllowedMethods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:          []string{"*"},
			MaxRequestSize:          10 * 1024 * 1024, // 10MB
		},
		Auth: security.AuthConfig{
			EnableDeviceCheck:     true,
			EnableIPCheck:         true,
			EnableRateLimit:       true,
			SkipAuthPaths:         []string{"/health", "/metrics"},
			SkipAuthPrefixes:      []string{"/api/public/"},
			SessionTimeout:        24 * time.Hour,
			MaxConcurrentSessions: 3,
			PasswordPolicy: security.PasswordPolicy{
				MinLength:             8,
				MaxLength:             128,
				RequireUppercase:      true,
				RequireLowercase:      true,
				RequireNumbers:        true,
				RequireSpecialChars:   true,
				ForbidCommonPasswords: true,
				ForbidPersonalInfo:    true,
				ExpiryDays:            90,
				HistoryCount:          5,
				FailedAttempts:        5,
				LockoutDuration:       30 * time.Minute,
			},
		},
		Validation: security.ValidationConfig{
			EnableInputValidation:   true,
			EnableSQLInjectionCheck: true,
			EnableXSSCheck:          true,
			MaxStringLength:         1000,
			MaxFileSize:             50 * 1024 * 1024, // 50MB
			AllowedFileTypes:        []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".xls", ".xlsx"},
		},
	}
}

// CombinedMiddleware 组合中间件（请求限制 + JWT认证）
func (sm *SecurityMiddleware) CombinedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 首先检查是否跳过认证
		if sm.shouldSkipAuth(c) {
			c.Next()
			return
		}

		// 应用JWT认证
		sm.authenticateJWT(c)
	}
}

// RateLimitMiddleware 仅请求限制中间件
func (sm *SecurityMiddleware) RateLimitMiddleware() gin.HandlerFunc {
	return sm.rateLimiter.Middleware()
}

// JWTMiddleware 仅JWT认证中间件
func (sm *SecurityMiddleware) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sm.shouldSkipAuth(c) {
			c.Next()
			return
		}
		sm.authenticateJWT(c)
	}
}

// shouldSkipAuth 检查是否跳过认证
func (sm *SecurityMiddleware) shouldSkipAuth(c *gin.Context) bool {
	path := c.Request.URL.Path

	// 检查跳过路径
	for _, skipPath := range sm.config.Auth.SkipAuthPaths {
		if path == skipPath {
			return true
		}
	}

	// 检查跳过前缀
	for _, prefix := range sm.config.Auth.SkipAuthPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// authenticateJWT JWT认证
func (sm *SecurityMiddleware) authenticateJWT(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请提供认证令牌",
		})
		c.Abort()
		return
	}

	// 检查 Bearer 前缀
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "无效的令牌格式",
		})
		c.Abort()
		return
	}

	tokenString := authHeader[7:] // 去掉 "Bearer " 前缀

	// 验证令牌
	payload, err := sm.jwtManager.ExtractTokenMetadata(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "无效的令牌",
		})
		c.Abort()
		return
	}

	// 检查令牌是否在黑名单中
	if sm.jwtManager.IsTokenBlacklisted(c.Request.Context(), tokenString) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "令牌已被撤销",
		})
		c.Abort()
		return
	}

	// 将用户信息存储到上下文中
	c.Set("user_id", payload.UserID)
	c.Set("username", payload.Username)
	c.Set("role", payload.Role)
	c.Set("device_id", payload.DeviceID)
	c.Set("ip", payload.IP)
	c.Set("user_agent", payload.UserAgent)

	c.Next()
}

// GenerateToken 生成令牌对
func (sm *SecurityMiddleware) GenerateToken(
	ctx context.Context,
	user *models.User,
	deviceID, ip, userAgent string,
) (*security.TokenDetails, error) {
	return sm.jwtManager.CreateTokens(ctx, user, deviceID, ip, userAgent)
}

// RefreshToken 刷新令牌
func (sm *SecurityMiddleware) RefreshToken(ctx context.Context, refreshToken string) (*security.TokenDetails, error) {
	return sm.jwtManager.RefreshTokens(ctx, refreshToken)
}

// RevokeToken 撤销令牌
func (sm *SecurityMiddleware) RevokeToken(ctx context.Context, tokenString string) error {
	return sm.jwtManager.RevokeToken(ctx, tokenString)
}

// RevokeAllUserTokens 撤销用户所有令牌
func (sm *SecurityMiddleware) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	return sm.jwtManager.RevokeAllUserTokens(ctx, userID)
}

// RotateKeys 旋转密钥
func (sm *SecurityMiddleware) RotateKeys(ctx context.Context) error {
	return sm.jwtManager.RotateKeys(ctx)
}

// AddIPToWhitelist 添加IP到白名单
func (sm *SecurityMiddleware) AddIPToWhitelist(ip string) {
	sm.rateLimiter.AddToWhitelist(ip)
}

// RemoveIPFromWhitelist 从白名单移除IP
func (sm *SecurityMiddleware) RemoveIPFromWhitelist(ip string) {
	sm.rateLimiter.RemoveFromWhitelist(ip)
}

// AddIPToBlacklist 添加IP到黑名单
func (sm *SecurityMiddleware) AddIPToBlacklist(ip string, duration time.Duration) {
	sm.rateLimiter.AddToBlacklist(ip, duration)
}

// RemoveIPFromBlacklist 从黑名单移除IP
func (sm *SecurityMiddleware) RemoveIPFromBlacklist(ip string) {
	sm.rateLimiter.RemoveFromBlacklist(ip)
}

// GetRateLimitInfo 获取请求限制信息
func (sm *SecurityMiddleware) GetRateLimitInfo(ctx context.Context, ip, endpoint string, userID string) (*security.RateLimitInfo, error) {
	return sm.rateLimiter.GetRateLimitInfo(ctx, ip, endpoint, userID)
}

// GetSecurityStats 获取安全统计信息
func (sm *SecurityMiddleware) GetSecurityStats() map[string]interface{} {
	return map[string]interface{}{
		"jwt_key_manager": sm.jwtManager.GetKeyStats(),
		"rate_limiter":    sm.rateLimiter.GetStats(),
	}
}

// getUserID 获取用户ID
func (sm *SecurityMiddleware) getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return fmt.Sprintf("%d", id)
		}
	}
	return ""
}

// LegacyAuthMiddleware 兼容旧版认证中间件
func (sm *SecurityMiddleware) LegacyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sm.shouldSkipAuth(c) {
			c.Next()
			return
		}
		sm.authenticateJWT(c)
	}
}

// LegacyRoleMiddleware 兼容旧版角色中间件
func (sm *SecurityMiddleware) LegacyRoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无访问权限",
			})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限验证失败",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		for _, role := range roles {
			if role == roleStr {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无访问权限",
		})
		c.Abort()
	}
}

// GetCurrentUserID 获取当前用户ID（兼容函数）
func (sm *SecurityMiddleware) GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	if id, ok := userID.(uint); ok {
		return id, true
	}

	return 0, false
}

// GetCurrentUsername 获取当前用户名（兼容函数）
func (sm *SecurityMiddleware) GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	if name, ok := username.(string); ok {
		return name, true
	}

	return "", false
}

// GetCurrentRole 获取当前用户角色（兼容函数）
func (sm *SecurityMiddleware) GetCurrentRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}

	if r, ok := role.(string); ok {
		return r, true
	}

	return "", false
}
