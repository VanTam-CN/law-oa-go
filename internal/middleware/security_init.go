package middleware

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

var (
	defaultSecurityMiddleware *SecurityMiddleware
	securityMiddlewareOnce    sync.Once
)

// InitSecurity 初始化安全中间件
func InitSecurity(cfg *config.Config) error {
	var initErr error

	securityMiddlewareOnce.Do(func() {
		// 创建Redis客户端
		redisClient := redis.NewClient(&redis.Options{
			Addr:            cfg.GetRedisAddr(),
			Password:        cfg.Redis.Password,
			DB:              cfg.Redis.DB,
			PoolSize:        100,
			MinIdleConns:    10,
			ConnMaxLifetime: 30 * time.Minute,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			PoolTimeout:     4 * time.Second,
		})

		// 测试Redis连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := redisClient.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("failed to connect to Redis: %w", err)
			return
		}

		// 创建缓存服务
		cacheService := cache.NewCacheService(redisClient, "lawoa")

		// 创建安全中间件
		defaultSecurityMiddleware = NewSecurityMiddleware(cfg, cacheService, redisClient)

		// 启动定期密钥轮转
		go startKeyRotation(defaultSecurityMiddleware)
	})

	return initErr
}

// GetSecurityMiddleware 获取默认安全中间件
func GetSecurityMiddleware() *SecurityMiddleware {
	if defaultSecurityMiddleware == nil {
		panic("Security middleware not initialized. Call InitSecurity first.")
	}
	return defaultSecurityMiddleware
}

// CombinedMiddleware 获取组合中间件（请求限制 + JWT认证）
func CombinedMiddleware() gin.HandlerFunc {
	return GetSecurityMiddleware().CombinedMiddleware()
}

// NewRateLimitMiddleware 获取请求限制中间件
func NewRateLimitMiddleware() gin.HandlerFunc {
	return GetSecurityMiddleware().RateLimitMiddleware()
}

// NewJWTMiddleware 获取JWT认证中间件
func NewJWTMiddleware() gin.HandlerFunc {
	return GetSecurityMiddleware().JWTMiddleware()
}

// NewGenerateToken 生成令牌对
func NewGenerateToken(
	ctx context.Context,
	user *models.User,
	deviceID, ip, userAgent string,
) (*security.TokenDetails, error) {
	return GetSecurityMiddleware().GenerateToken(ctx, user, deviceID, ip, userAgent)
}

// RefreshToken 刷新令牌
func RefreshToken(ctx context.Context, refreshToken string) (*security.TokenDetails, error) {
	return GetSecurityMiddleware().RefreshToken(ctx, refreshToken)
}

// RevokeToken 撤销令牌
func RevokeToken(ctx context.Context, tokenString string) error {
	return GetSecurityMiddleware().RevokeToken(ctx, tokenString)
}

// RevokeAllUserTokens 撤销用户所有令牌
func RevokeAllUserTokens(ctx context.Context, userID uint) error {
	return GetSecurityMiddleware().RevokeAllUserTokens(ctx, userID)
}

// AddIPToWhitelist 添加IP到白名单
func AddIPToWhitelist(ip string) {
	GetSecurityMiddleware().AddIPToWhitelist(ip)
}

// RemoveIPFromWhitelist 从白名单移除IP
func RemoveIPFromWhitelist(ip string) {
	GetSecurityMiddleware().RemoveIPFromWhitelist(ip)
}

// AddIPToBlacklist 添加IP到黑名单
func AddIPToBlacklist(ip string, duration time.Duration) {
	GetSecurityMiddleware().AddIPToBlacklist(ip, duration)
}

// RemoveIPFromBlacklist 从黑名单移除IP
func RemoveIPFromBlacklist(ip string) {
	GetSecurityMiddleware().RemoveIPFromBlacklist(ip)
}

// GetRateLimitInfo 获取请求限制信息
func GetRateLimitInfo(ctx context.Context, ip, endpoint string, userID string) (*security.RateLimitInfo, error) {
	return GetSecurityMiddleware().GetRateLimitInfo(ctx, ip, endpoint, userID)
}

// GetSecurityStats 获取安全统计信息
func GetSecurityStats() map[string]interface{} {
	return GetSecurityMiddleware().GetSecurityStats()
}

// startKeyRotation 启动定期密钥轮转
func startKeyRotation(sm *SecurityMiddleware) {
	ticker := time.NewTicker(24 * time.Hour) // 每天轮转一次
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if err := sm.RotateKeys(ctx); err != nil {
			log.Printf("Failed to rotate JWT keys: %v", err)
		} else {
			log.Println("JWT keys rotated successfully")
		}
	}
}

// LegacyInitJWT 兼容旧版JWT初始化
func LegacyInitJWT(cfg *config.Config) {
	if err := InitSecurity(cfg); err != nil {
		log.Printf("Failed to initialize security middleware: %v", err)
	}
}

// LegacyGenerateToken 兼容旧版令牌生成
func LegacyGenerateToken(userID uint, username, role string) (string, time.Time, error) {
	sm := GetSecurityMiddleware()

	user := &models.User{
		ID:    userID,
		Name:  username,
		Email: "", // 需要用户提供
		Role:  role,
	}

	tokenDetails, err := sm.GenerateToken(context.Background(), user, "", "", "")
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenDetails.AccessToken, time.Unix(tokenDetails.AtExpires, 0), nil
}

// LegacyValidateToken 兼容旧版令牌验证
func LegacyValidateToken(tokenString string) (*JWTClaims, error) {
	sm := GetSecurityMiddleware()

	payload, err := sm.jwtManager.ExtractTokenMetadata(context.Background(), tokenString)
	if err != nil {
		return nil, err
	}

	return &JWTClaims{
		UserID:   payload.UserID,
		Username: payload.Username,
		Role:     payload.Role,
	}, nil
}

// LegacyAuthMiddleware 兼容旧版认证中间件
func LegacyAuthMiddleware() gin.HandlerFunc {
	return GetSecurityMiddleware().LegacyAuthMiddleware()
}

// LegacyRoleMiddleware 兼容旧版角色中间件
func LegacyRoleMiddleware(roles ...string) gin.HandlerFunc {
	return GetSecurityMiddleware().LegacyRoleMiddleware(roles...)
}

// LegacyGetCurrentUserID 兼容旧版获取用户ID
func LegacyGetCurrentUserID(c *gin.Context) (uint, bool) {
	return GetSecurityMiddleware().GetCurrentUserID(c)
}

// LegacyGetCurrentUsername 兼容旧版获取用户名
func LegacyGetCurrentUsername(c *gin.Context) (string, bool) {
	return GetSecurityMiddleware().GetCurrentUsername(c)
}

// LegacyGetCurrentRole 兼容旧版获取角色
func LegacyGetCurrentRole(c *gin.Context) (string, bool) {
	return GetSecurityMiddleware().GetCurrentRole(c)
}

// LegacyIPRateLimitMiddleware 兼容旧版IP限流中间件
func LegacyIPRateLimitMiddleware(limit int64, window time.Duration, redisClient *redis.Client) gin.HandlerFunc {
	sm := GetSecurityMiddleware()

	// 配置IP级别限流 - 设置一个通用的默认规则
	sm.rateLimiter.SetEndpointLimit("*", security.RateLimitRule{
		Window:      window,
		MaxRequests: int(limit),
		Burst:       int(limit) / 2,
	})

	return sm.RateLimitMiddleware()
}

// LegacyUserRateLimitMiddleware 兼容旧版用户限流中间件
func LegacyUserRateLimitMiddleware(limit int64, window time.Duration, redisClient *redis.Client) gin.HandlerFunc {
	sm := GetSecurityMiddleware()

	// 配置用户级别限流
	sm.rateLimiter.SetUserLimit("*", security.RateLimitRule{
		Window:      window,
		MaxRequests: int(limit),
		Burst:       int(limit) / 2,
	})

	return sm.RateLimitMiddleware()
}

// LegacyAPIRateLimitMiddleware 兼容旧版API限流中间件
func LegacyAPIRateLimitMiddleware(limit int64, window time.Duration, redisClient *redis.Client) gin.HandlerFunc {
	sm := GetSecurityMiddleware()

	// 配置API级别限流
	sm.rateLimiter.SetEndpointLimit("*", security.RateLimitRule{
		Window:      window,
		MaxRequests: int(limit),
		Burst:       int(limit) / 2,
	})

	return sm.RateLimitMiddleware()
}

// LegacyBruteForceProtectionMiddleware 兼容旧版暴力破解保护中间件
func LegacyBruteForceProtectionMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	sm := GetSecurityMiddleware()

	// 配置登录端点严格限制
	sm.rateLimiter.SetEndpointLimit("/api/auth/login", security.RateLimitRule{
		Window:      time.Minute,
		MaxRequests: 5,
		Burst:       2,
	})

	return sm.RateLimitMiddleware()
}
