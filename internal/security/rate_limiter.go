package security

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

var (
	rateLimitRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_requests_total",
		Help: "Total number of rate limit requests",
	}, []string{"action", "status"})

	rateLimitBlocked = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rate_limit_blocked_total",
		Help: "Total number of blocked requests due to rate limiting",
	}, []string{"endpoint", "ip"})

	rateLimitCurrent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rate_limit_current_requests",
		Help: "Current number of requests in rate limit window",
	}, []string{"endpoint", "ip"})
)

// RateLimiter 请求限制器
type RateLimiter struct {
	redisClient    *redis.Client
	config         *APISecurityConfig
	mu             sync.RWMutex
	ipBlacklist    map[string]time.Time
	ipWhitelist    map[string]struct{}
	userLimits     map[string]RateLimitRule
	endpointLimits map[string]RateLimitRule
}

// RateLimitRule 限制规则
type RateLimitRule struct {
	Window      time.Duration `json:"window"`
	MaxRequests int           `json:"max_requests"`
	Burst       int           `json:"burst"`
}

// RateLimitInfo 限制信息
type RateLimitInfo struct {
	CurrentCount int           `json:"current_count"`
	MaxRequests  int           `json:"max_requests"`
	Window       time.Duration `json:"window"`
	ResetTime    time.Time     `json:"reset_time"`
	IsBlocked    bool          `json:"is_blocked"`
}

// NewRateLimiter 创建新的请求限制器
func NewRateLimiter(redisClient *redis.Client, config *APISecurityConfig) *RateLimiter {
	limiter := &RateLimiter{
		redisClient:    redisClient,
		config:         config,
		ipBlacklist:    make(map[string]time.Time),
		ipWhitelist:    make(map[string]struct{}),
		userLimits:     make(map[string]RateLimitRule),
		endpointLimits: make(map[string]RateLimitRule),
	}

	// 初始化IP白名单
	for _, ip := range config.WhitelistedIPs {
		limiter.ipWhitelist[ip] = struct{}{}
	}

	// 初始化默认端点限制
	limiter.endpointLimits["/api/auth/login"] = RateLimitRule{
		Window:      time.Minute,
		MaxRequests: 5, // 登录端点更严格的限制
		Burst:       2,
	}

	limiter.endpointLimits["/api/auth/register"] = RateLimitRule{
		Window:      time.Hour,
		MaxRequests: 3, // 注册端点严格的限制
		Burst:       1,
	}

	return limiter
}

// Middleware 返回Gin中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.config.EnableRateLimit {
			c.Next()
			return
		}

		// 获取客户端IP
		clientIP := rl.getClientIP(c)
		endpoint := c.FullPath()
		userID := rl.getUserID(c)

		// 检查IP白名单
		if rl.isIPWhitelisted(clientIP) {
			c.Next()
			return
		}

		// 检查IP黑名单
		if rl.isIPBlacklisted(clientIP) {
			rl.blockRequest(c, clientIP, endpoint, "IP黑名单")
			return
		}

		// 应用请求限制
		if !rl.allowRequest(c, clientIP, endpoint, userID) {
			return
		}

		c.Next()
	}
}

// getClientIP 获取客户端IP
func (rl *RateLimiter) getClientIP(c *gin.Context) string {
	// 检查X-Forwarded-For头
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		// 取第一个IP（原始客户端IP），X-Forwarded-For可能包含多个IP，用逗号分隔
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 检查X-Real-IP头
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 使用远程地址
	clientIP := c.ClientIP()
	if clientIP == "" && c.Request.RemoteAddr != "" {
		// 如果ClientIP()返回空，直接从RemoteAddr提取IP（去掉端口号）
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err == nil {
			clientIP = host
		} else {
			clientIP = c.Request.RemoteAddr
		}
	}
	return clientIP
}

// getUserID 获取用户ID
func (rl *RateLimiter) getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return strconv.FormatUint(uint64(id), 10)
		}
	}
	return ""
}

// isIPWhitelisted 检查IP是否在白名单中
func (rl *RateLimiter) isIPWhitelisted(ip string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	_, exists := rl.ipWhitelist[ip]
	return exists
}

// isIPBlacklisted 检查IP是否在黑名单中
func (rl *RateLimiter) isIPBlacklisted(ip string) bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	// 尝试直接匹配
	if blockTime, exists := rl.ipBlacklist[ip]; exists {
		return time.Now().Before(blockTime)
	}

	// 尝试提取IP地址（去掉端口号）
	if ipWithPort := strings.Split(ip, ":"); len(ipWithPort) > 1 {
		cleanIP := ipWithPort[0]
		if blockTime, exists := rl.ipBlacklist[cleanIP]; exists {
			return time.Now().Before(blockTime)
		}
	}

	return false
}

// allowRequest 检查是否允许请求
func (rl *RateLimiter) allowRequest(c *gin.Context, ip, endpoint, userID string) bool {
	ctx := c.Request.Context()

	// 获取适用的限制规则
	rule := rl.getRateLimitRule(endpoint, userID)

	// 生成限制键
	key := rl.generateRateLimitKey(ip, endpoint, userID)

	// 检查请求限制
	allowed, info, err := rl.checkRateLimit(ctx, key, rule)
	if err != nil {
		// 记录错误但允许请求通过
		fmt.Printf("Rate limit check error: %v\n", err)
		return true
	}

	// 设置响应头
	rl.setRateLimitHeaders(c, info)

	if !allowed {
		rl.blockRequest(c, ip, endpoint, "请求频率超限")
		rateLimitBlocked.WithLabelValues(endpoint, ip).Inc()
		return false
	}

	// 记录当前请求数
	rateLimitCurrent.WithLabelValues(endpoint, ip).Set(float64(info.CurrentCount))
	rateLimitRequests.WithLabelValues("request", "allowed").Inc()

	return true
}

// getRateLimitRule 获取限制规则
func (rl *RateLimiter) getRateLimitRule(endpoint, userID string) RateLimitRule {
	// 检查端点特定规则
	if rule, exists := rl.endpointLimits[endpoint]; exists {
		return rule
	}

	// 检查用户特定规则
	if userID != "" {
		if rule, exists := rl.userLimits[userID]; exists {
			return rule
		}
	}

	// 使用默认规则
	return RateLimitRule{
		Window:      rl.config.RateLimitWindow,
		MaxRequests: rl.config.RateLimitMaxRequests,
		Burst:       rl.config.RateLimitMaxRequests / 2,
	}
}

// generateRateLimitKey 生成限制键
func (rl *RateLimiter) generateRateLimitKey(ip, endpoint, userID string) string {
	if userID != "" {
		return fmt.Sprintf("rate_limit:user:%s:%s", userID, endpoint)
	}
	return fmt.Sprintf("rate_limit:ip:%s:%s", ip, endpoint)
}

// checkRateLimit 检查请求限制
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string, rule RateLimitRule) (bool, *RateLimitInfo, error) {
	// 使用Redis的原子操作来检查和增加计数
	pipe := rl.redisClient.Pipeline()

	// 获取当前计数
	currentCmd := pipe.Get(ctx, key)

	// 设置过期时间（如果不存在）
	pipe.Expire(ctx, key, rule.Window)

	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return false, nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	// 获取当前计数
	currentCount, err := currentCmd.Int()
	if err != nil && err != redis.Nil {
		return false, nil, fmt.Errorf("redis get error: %w", err)
	}

	// 检查是否超过限制
	if currentCount >= rule.MaxRequests {
		// 获取重置时间
		ttlCmd := rl.redisClient.TTL(ctx, key)
		ttl, err := ttlCmd.Result()
		if err != nil {
			return false, nil, fmt.Errorf("redis ttl error: %w", err)
		}

		resetTime := time.Now().Add(ttl)
		info := &RateLimitInfo{
			CurrentCount: currentCount,
			MaxRequests:  rule.MaxRequests,
			Window:       rule.Window,
			ResetTime:    resetTime,
			IsBlocked:    true,
		}
		return false, info, nil
	}

	// 增加计数
	newCount, err := rl.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false, nil, fmt.Errorf("redis incr error: %w", err)
	}

	// 获取重置时间
	ttlCmd := rl.redisClient.TTL(ctx, key)
	ttl, err := ttlCmd.Result()
	if err != nil {
		return false, nil, fmt.Errorf("redis ttl error: %w", err)
	}

	resetTime := time.Now().Add(ttl)
	info := &RateLimitInfo{
		CurrentCount: int(newCount),
		MaxRequests:  rule.MaxRequests,
		Window:       rule.Window,
		ResetTime:    resetTime,
		IsBlocked:    false,
	}

	return true, info, nil
}

// setRateLimitHeaders 设置限制头
func (rl *RateLimiter) setRateLimitHeaders(c *gin.Context, info *RateLimitInfo) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(info.MaxRequests))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(info.MaxRequests-info.CurrentCount))
	c.Header("X-RateLimit-Reset", info.ResetTime.Format(time.RFC3339))
	c.Header("X-RateLimit-Window", info.Window.String())
}

// blockRequest 阻止请求
func (rl *RateLimiter) blockRequest(c *gin.Context, ip, endpoint, reason string) {
	rateLimitRequests.WithLabelValues("request", "blocked").Inc()

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":       "请求频率超限",
		"message":     reason,
		"retry_after": rl.getRetryAfter(c, ip, endpoint),
	})
	c.Abort()
}

// getRetryAfter 获取重试时间
func (rl *RateLimiter) getRetryAfter(c *gin.Context, ip, endpoint string) int {
	ctx := c.Request.Context()
	key := rl.generateRateLimitKey(ip, endpoint, "")

	ttl, err := rl.redisClient.TTL(ctx, key).Result()
	if err != nil {
		return 60 // 默认60秒
	}

	return int(ttl.Seconds())
}

// AddToBlacklist 添加IP到黑名单
func (rl *RateLimiter) AddToBlacklist(ip string, duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.ipBlacklist[ip] = time.Now().Add(duration)
}

// RemoveFromBlacklist 从黑名单移除IP
func (rl *RateLimiter) RemoveFromBlacklist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.ipBlacklist, ip)
}

// AddToWhitelist 添加IP到白名单
func (rl *RateLimiter) AddToWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.ipWhitelist[ip] = struct{}{}
}

// RemoveFromWhitelist 从白名单移除IP
func (rl *RateLimiter) RemoveFromWhitelist(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.ipWhitelist, ip)
}

// SetUserLimit 设置用户特定限制
func (rl *RateLimiter) SetUserLimit(userID string, rule RateLimitRule) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.userLimits[userID] = rule
}

// SetEndpointLimit 设置端点特定限制
func (rl *RateLimiter) SetEndpointLimit(endpoint string, rule RateLimitRule) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.endpointLimits[endpoint] = rule
}

// GetRateLimitInfo 获取限制信息
func (rl *RateLimiter) GetRateLimitInfo(ctx context.Context, ip, endpoint, userID string) (*RateLimitInfo, error) {
	rule := rl.getRateLimitRule(endpoint, userID)
	key := rl.generateRateLimitKey(ip, endpoint, userID)

	_, info, err := rl.checkRateLimit(ctx, key, rule)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// CleanupExpiredBlacklist 清理过期的黑名单条目
func (rl *RateLimiter) CleanupExpiredBlacklist() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, expiry := range rl.ipBlacklist {
		if now.After(expiry) {
			delete(rl.ipBlacklist, ip)
		}
	}
}

// GetStats 获取统计信息
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"ip_blacklist_size":    len(rl.ipBlacklist),
		"ip_whitelist_size":    len(rl.ipWhitelist),
		"user_limits_size":     len(rl.userLimits),
		"endpoint_limits_size": len(rl.endpointLimits),
		"enable_rate_limit":    rl.config.EnableRateLimit,
		"default_window":       rl.config.RateLimitWindow.String(),
		"default_max_requests": rl.config.RateLimitMaxRequests,
	}
}

// StartCleanupTimer 启动清理定时器
func (rl *RateLimiter) StartCleanupTimer() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			rl.CleanupExpiredBlacklist()
		}
	}()
}
