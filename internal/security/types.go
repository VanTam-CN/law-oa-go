package security

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	SecretKey          string
	AccessTokenExpiry  int
	RefreshTokenExpiry int
	Issuer             string
	Audience           []string
	Algorithm          string
	RefreshTokenSecret string
	AccessTokenTTL     int64  // 新增字段，改为int64以支持time.Duration
	RefreshTokenTTL    int64  // 新增字段，改为int64以支持time.Duration
	Secret             string // 新增字段
	EnableRefresh      bool   // 新增字段
	BlacklistTTL       int64  // 新增字段，改为int64以支持time.Duration
}

// APISecurityConfig API安全配置
type APISecurityConfig struct {
	EnableHTTPS             bool
	EnableCORS              bool
	AllowedOrigins          []string
	AllowedMethods          []string
	AllowedHeaders          []string
	MaxAge                  int
	RateLimiting            RateLimitConfig
	RequestSizeLimit        int
	ResponseSizeLimit       int
	Timeout                 int
	EnableCompression       bool
	EnableMetrics           bool
	SecurityHeaders         SecurityHeadersConfig
	EnableRateLimit         bool     // 新增字段
	EnableIPWhitelist       bool     // 新增字段
	EnableIPBlacklist       bool     // 新增字段
	EnableRequestSigning    bool     // 新增字段
	EnableAPIThrottling     bool     // 新增字段
	EnableWAFProtection     bool     // 新增字段
	EnableDDoSProtection    bool     // 新增字段
	EnableRequestValidation bool     // 新增字段
	EnableCSRF              bool     // 新增字段
	RateLimitWindow         int64    // 新增字段
	RateLimitMaxRequests    int      // 新增字段
	WhitelistedIPs          []string // 新增字段
	BlacklistedIPs          []string // 新增字段
	MaxRequestSize          int64    // 新增字段
}

// AuthConfig 认证配置
type AuthConfig struct {
	EnableJWT         bool
	EnableOAuth       bool
	EnableLDAP        bool
	Enable2FA         bool
	SessionTimeout    int
	MaxLoginAttempts  int
	LockoutDuration   int
	PasswordPolicy    PasswordPolicy
	TokenBlacklist    bool
	EnableAudit       bool
	FailedLoginWindow int
	EnableDeviceCheck bool // 新增字段
	EnableIPCheck     bool // 新增字段
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSymbols   bool
	PreventReuse     bool
	MaxReuseCount    int
	PreventCommon    bool
	PreventUserInfo  bool
	ExpiryDays       int
	WarningDays      int
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	EnableInputValidation bool
	MaxStringLength       int
	AllowedFileTypes      []string
	MaxFileSize           int
	EnableSQLInjection    bool
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
	TokenID     string
	UserID      string
	Username    string
	ExpiresAt   int64
	IssuedAt    int64
	TokenType   string
	Roles       []string
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
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略 - 限制浏览器功能访问
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		c.Next()
	}
}

// CORS 添加跨域中间件
func CORS() gin.HandlerFunc {
	// 从环境变量读取允许的域名列表，默认允许 localhost 开发端口
	allowedOriginsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins map[string]bool
	if allowedOriginsStr == "" {
		allowedOrigins = map[string]bool{
			"http://localhost:3000": true,
			"http://localhost:8080": true,
		}
	} else {
		allowedOrigins = make(map[string]bool)
		for _, origin := range strings.Split(allowedOriginsStr, ",") {
			allowedOrigins[strings.TrimSpace(origin)] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ipRequestRecord IP请求记录，用于滑动窗口限流
type ipRequestRecord struct {
	count     int        // 当前窗口内的请求计数
	windowEnd int64      // 窗口结束时间（Unix时间戳，秒）
	mu        sync.Mutex // 保护count字段的互斥锁
}

// SlidingWindowRateLimiter 滑动窗口速率限制器
// 使用sync.Map存储每个IP的请求记录，实现基于IP的速率限制
type SlidingWindowRateLimiter struct {
	records sync.Map      // map[string]*ipRequestRecord，使用sync.Map避免全局锁
	maxReqs int           // 时间窗口内最大请求数
	window  time.Duration // 时间窗口大小
}

// NewSlidingWindowRateLimiter 创建滑动窗口速率限制器
// maxReqs: 时间窗口内允许的最大请求数
// window: 时间窗口大小
func NewSlidingWindowRateLimiter(maxReqs int, window time.Duration) *SlidingWindowRateLimiter {
	rl := &SlidingWindowRateLimiter{
		maxReqs: maxReqs,
		window:  window,
	}
	// 启动后台清理定时器
	rl.StartCleanupTimer()
	return rl
}

// Allow 检查是否允许来自指定IP的请求
// 使用滑动窗口算法：每个时间窗口内最多maxReqs个请求
func (rl *SlidingWindowRateLimiter) Allow(ip string) bool {
	now := time.Now().Unix()
	windowEnd := now + int64(rl.window.Seconds())

	// 获取或创建该IP的记录
	value, _ := rl.records.LoadOrStore(ip, &ipRequestRecord{
		count:     0,
		windowEnd: windowEnd,
	})
	record := value.(*ipRequestRecord)

	// 使用细粒度锁保护该IP的记录
	record.mu.Lock()
	defer record.mu.Unlock()

	// 检查窗口是否过期，过期则重置
	if now > record.windowEnd {
		record.count = 1
		record.windowEnd = windowEnd
		rl.records.Store(ip, record)
		return true
	}

	// 当前窗口内，检查是否超限
	if record.count >= rl.maxReqs {
		return false
	}

	// 增加计数
	record.count++
	rl.records.Store(ip, record)
	return true
}

// GetRateLimitInfo 获取指定IP的当前限流状态
// 返回: 当前请求数, 限制, 窗口重置时间
func (rl *SlidingWindowRateLimiter) GetRateLimitInfo(ip string) (current int, limit int, resetTime int64) {
	value, ok := rl.records.Load(ip)
	if !ok {
		return 0, rl.maxReqs, time.Now().Add(rl.window).Unix()
	}
	record := value.(*ipRequestRecord)
	return record.count, rl.maxReqs, record.windowEnd
}

// StartCleanupTimer 启动清理定时器，定期清理过期的IP记录
// 防止内存泄漏，每分钟清理一次过期记录
func (rl *SlidingWindowRateLimiter) StartCleanupTimer() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().Unix()
			rl.records.Range(func(key, value interface{}) bool {
				record := value.(*ipRequestRecord)
				// 清理窗口结束后超过1分钟的记录
				if now > record.windowEnd+60 {
					rl.records.Delete(key)
				}
				return true
			})
		}
	}()
}

// 全局限流器实例：100 req/min per IP
var globalRateLimiter = NewSlidingWindowRateLimiter(100, 1*time.Minute)

// RateLimiterMiddleware 速率限制中间件
// 基于客户端IP实现滑动窗口限流：100 req/min per IP
// 超出限制时返回 429 Too Many Requests
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		ip := c.ClientIP()

		// 检查是否允许请求
		if !globalRateLimiter.Allow(ip) {
			// 获取当前限流状态用于响应头
			_, limit, resetTime := globalRateLimiter.GetRateLimitInfo(ip)
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
