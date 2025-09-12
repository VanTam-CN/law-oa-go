package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"law-oa-go/internal/logger"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RedisClient *redis.Client
	KeyPrefix   string
	Limit       int64                   // 请求限制数量
	Window      time.Duration           // 时间窗口
	SkipFunc    func(*gin.Context) bool // 跳过限流的条件
}

// RateLimitMiddleware 高级限流中间件
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	auditLogger := logger.NewAuditLogger()

	return func(c *gin.Context) {
		// 检查是否跳过限流
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		// 生成限流键
		key := generateRateLimitKey(c, config.KeyPrefix)

		// 检查限流
		allowed, remaining, resetTime, err := checkRateLimit(c, config.RedisClient, key, config.Limit, config.Window)
		if err != nil {
			// 限流检查失败，记录错误但不阻止请求
			logger.WithContext(c.Request.Context()).Error("Rate limit check failed",
				zap.String("error", err.Error()),
				zap.String("key", key),
			)
			c.Next()
			return
		}

		// 设置响应头
		c.Header("X-RateLimit-Limit", strconv.FormatInt(config.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			// 记录限流事件
			auditLogger.LogSecurityEvent(c.Request.Context(), "rate_limit_exceeded", "medium", map[string]interface{}{
				"client_ip":  c.ClientIP(),
				"user_agent": c.GetHeader("User-Agent"),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"limit":      config.Limit,
				"window":     config.Window.String(),
			})

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     fmt.Sprintf("Too many requests. Limit: %d per %v", config.Limit, config.Window),
				"retry_after": resetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateRateLimitKey 生成限流键
func generateRateLimitKey(c *gin.Context, prefix string) string {
	// 优先使用用户ID，其次使用IP
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("%s:user:%v", prefix, userID)
	}
	return fmt.Sprintf("%s:ip:%s", prefix, c.ClientIP())
}

// checkRateLimit 检查限流
func checkRateLimit(c *gin.Context, client *redis.Client, key string, limit int64, window time.Duration) (bool, int64, time.Time, error) {
	ctx := c.Request.Context()

	// 使用Redis的INCR和EXPIRE命令实现滑动窗口限流
	pipe := client.Pipeline()

	// 增加计数
	incrCmd := pipe.Incr(ctx, key)
	// 设置过期时间
	_ = pipe.Expire(ctx, key, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, err
	}

	count := incrCmd.Val()

	// 如果是第一次请求，确保设置了过期时间
	if count == 1 {
		client.Expire(ctx, key, window)
	}

	// 计算剩余次数和重置时间
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := time.Now().Add(window)

	return count <= limit, remaining, resetTime, nil
}

// IPRateLimitMiddleware IP级别限流
func IPRateLimitMiddleware(limit int64, window time.Duration, redisClient *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		RedisClient: redisClient,
		KeyPrefix:   "rate_limit_ip",
		Limit:       limit,
		Window:      window,
	})
}

// UserRateLimitMiddleware 用户级别限流
func UserRateLimitMiddleware(limit int64, window time.Duration, redisClient *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		RedisClient: redisClient,
		KeyPrefix:   "rate_limit_user",
		Limit:       limit,
		Window:      window,
		SkipFunc: func(c *gin.Context) bool {
			// 未登录用户跳过用户级限流
			_, exists := c.Get("user_id")
			return !exists
		},
	})
}

// APIRateLimitMiddleware API端点级别限流
func APIRateLimitMiddleware(limit int64, window time.Duration, redisClient *redis.Client) gin.HandlerFunc {
	return RateLimitMiddleware(RateLimitConfig{
		RedisClient: redisClient,
		KeyPrefix:   "rate_limit_api",
		Limit:       limit,
		Window:      window,
	})
}

// BruteForceProtectionMiddleware 暴力破解保护
func BruteForceProtectionMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	auditLogger := logger.NewAuditLogger()

	return func(c *gin.Context) {
		// 只对登录接口进行暴力破解保护
		if c.Request.URL.Path != "/api/v1/auth/login" {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("brute_force:%s", clientIP)

		// 检查是否已被锁定
		locked, err := redisClient.Exists(c.Request.Context(), key+":locked").Result()
		if err == nil && locked > 0 {
			auditLogger.LogSecurityEvent(c.Request.Context(), "brute_force_blocked", "high", map[string]interface{}{
				"client_ip":  clientIP,
				"user_agent": c.GetHeader("User-Agent"),
				"path":       c.Request.URL.Path,
			})

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Account temporarily locked",
				"message": "Too many failed login attempts. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()

		// 检查登录是否失败
		if c.Writer.Status() == http.StatusUnauthorized {
			// 增加失败计数
			count, err := redisClient.Incr(c.Request.Context(), key).Result()
			if err == nil {
				// 设置过期时间
				redisClient.Expire(c.Request.Context(), key, 15*time.Minute)

				// 如果失败次数超过阈值，锁定账户
				if count >= 5 {
					redisClient.Set(c.Request.Context(), key+":locked", "1", 30*time.Minute)

					auditLogger.LogSecurityEvent(c.Request.Context(), "brute_force_detected", "critical", map[string]interface{}{
						"client_ip":     clientIP,
						"user_agent":    c.GetHeader("User-Agent"),
						"failed_count":  count,
						"lock_duration": "30 minutes",
					})
				}
			}
		} else if c.Writer.Status() == http.StatusOK {
			// 登录成功，清除失败计数
			redisClient.Del(c.Request.Context(), key, key+":locked")
		}
	}
}
