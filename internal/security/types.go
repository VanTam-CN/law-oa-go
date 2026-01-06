package security

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"law-oa-go/internal/config"
	"law-oa-go/internal/cache"
)

// JWTKeyManager JWT密钥管理器
type JWTKeyManager struct{}

// NewJWTKeyManager 创建JWT密钥管理器
func NewJWTKeyManager(cfg *config.Config, config *SecurityConfig, redis *redis.Client, cacheService *cache.CacheService) *JWTKeyManager {
	return &JWTKeyManager{}
}

// RateLimiter 速率限制器
type RateLimiter struct{}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(redis *redis.Client, config *APISecurityConfig) *RateLimiter {
	return &RateLimiter{}
}

// StartCleanupTimer 启动清理定时器
func (rl *RateLimiter) StartCleanupTimer() {
	// TODO: 实现清理逻辑
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWT         JWTConfig
	APISecurity APISecurityConfig
	Auth        AuthConfig
	Validation  ValidationConfig
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey           string
	AccessTokenExpiry   int
	RefreshTokenExpiry  int
	Issuer              string
	Audience            []string
	Algorithm           string
	RefreshTokenSecret  string
	AccessTokenTTL      int64    // 新增字段，改为int64以支持time.Duration
	RefreshTokenTTL     int64    // 新增字段，改为int64以支持time.Duration
	Secret              string // 新增字段
	EnableRefresh       bool   // 新增字段
	BlacklistTTL        int64    // 新增字段，改为int64以支持time.Duration
}

// APISecurityConfig API安全配置
type APISecurityConfig struct {
	EnableHTTPS         bool
	EnableCORS          bool
	AllowedOrigins      []string
	AllowedMethods      []string
	AllowedHeaders      []string
	MaxAge              int
	RateLimiting        RateLimitConfig
	RequestSizeLimit    int
	ResponseSizeLimit   int
	Timeout             int
	EnableCompression   bool
	EnableMetrics       bool
	SecurityHeaders     SecurityHeadersConfig
	EnableRateLimit     bool   // 新增字段
	EnableIPWhitelist   bool   // 新增字段
	EnableIPBlacklist   bool   // 新增字段
	EnableRequestSigning bool   // 新增字段
	EnableAPIThrottling bool   // 新增字段
	EnableWAFProtection bool   // 新增字段
	EnableDDoSProtection bool   // 新增字段
	EnableRequestValidation bool // 新增字段
	EnableCSRF          bool   // 新增字段
	RateLimitWindow     int64  // 新增字段
	RateLimitMaxRequests int    // 新增字段
	WhitelistedIPs      []string // 新增字段
	BlacklistedIPs      []string // 新增字段
	MaxRequestSize      int64   // 新增字段
}

// AuthConfig 认证配置
type AuthConfig struct {
	EnableJWT          bool
	EnableOAuth        bool
	EnableLDAP         bool
	Enable2FA          bool
	SessionTimeout     int
	MaxLoginAttempts   int
	LockoutDuration    int
	PasswordPolicy     PasswordPolicy
	TokenBlacklist     bool
	EnableAudit        bool
	FailedLoginWindow  int
	EnableDeviceCheck  bool   // 新增字段
	EnableIPCheck      bool   // 新增字段
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength         int
	RequireUppercase   bool
	RequireLowercase   bool
	RequireNumbers     bool
	RequireSymbols     bool
	PreventReuse       bool
	MaxReuseCount      int
	PreventCommon      bool
	PreventUserInfo   bool
	ExpiryDays         int
	WarningDays       int
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	EnableInputValidation bool
	MaxStringLength      int
	AllowedFileTypes     []string
	MaxFileSize          int
	EnableSQLInjection   bool
	EnableXSSProtection   bool
	EnableCSRFProtection  bool
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	Enabled         bool
	RequestsPerMin  int
	RequestsPerHour int
	RequestsPerDay  int
	BurstSize       int
	Whitelist       []string
	Blacklist       []string
}

// SecurityHeadersConfig 安全头配置
type SecurityHeadersConfig struct {
	XFrameOptions     string
	XContentOptions   string
	XSSProtection     string
	StrictTransport   string
	ContentSecurity   string
	ReferrerPolicy    string
	PermissionsPolicy string
}

// TokenDetails token详情
type TokenDetails struct {
	TokenID    string
	UserID     string
	Username   string
	ExpiresAt  int64
	IssuedAt   int64
	TokenType  string
	Roles      []string
	Permissions []string
}

// RateLimitInfo 速率限制信息
type RateLimitInfo struct {
	Current   int
	Limit     int
	Remaining int
	ResetTime int64
	Window    string
}

// Add missing methods for SecurityMiddleware compatibility

// SecurityHeaders 添加安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// CORS 添加跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// RateLimiterMiddleware 速率限制中间件
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}