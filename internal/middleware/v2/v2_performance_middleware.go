package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"law-oa-go/internal/config"
	"law-oa-go/internal/logger"
)

// V2PerformanceMiddleware 基于最新Gin v2性能最佳实践的中间件
type V2PerformanceMiddleware struct {
	config        *config.Config
	redisClient   *redis.Client
	metrics       *AtomicMetrics
	startTime     time.Time
	requestCount  int64
	slowThreshold time.Duration
	enabled       bool
}

// AtomicMetrics 使用原子操作的性能指标
type AtomicMetrics struct {
	totalRequests   int64
	slowRequests     int64
	errorRequests    int64
	totalLatency     int64 // 纳秒
	activeRequests   int32
	maxConcurrent    int32
	memoryUsage      int64 // 最大内存使用
	mu               sync.RWMutex
	lastGC           time.Time
	gcCount          uint32
}

// NewV2PerformanceMiddleware 创建性能中间件
func NewV2PerformanceMiddleware(config *config.Config, redisClient *redis.Client) *V2PerformanceMiddleware {
	return &V2PerformanceMiddleware{
		config:        config,
		redisClient:   redisClient,
		metrics:       &AtomicMetrics{startTime: time.Now()},
		slowThreshold: 500 * time.Millisecond, // 可配置的慢请求阈值
		enabled:       config.Database.EnablePerformance,
	}
}

// PerformanceMiddleware 主要性能监控中间件
func (pm *V2PerformanceMiddleware) PerformanceMiddleware() gin.HandlerFunc {
	if !pm.enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()

		// 原子递增活跃请求计数
		currentActive := atomic.AddInt32(&pm.metrics.totalRequests, 1)
		defer atomic.AddInt32(&pm.metrics.totalRequests, -1)

		// 更新最大并发数
		for {
			currentMax := atomic.LoadInt32(&pm.metrics.maxConcurrent)
			if currentMax >= currentActive {
				break
			}
			if atomic.CompareAndSwapInt32(&pm.metrics.maxConcurrent, currentMax, currentActive) {
				break
			}
		}

		// 创建自定义ResponseWriter来监控响应
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			size:           0,
			status:         0,
		}
		c.Writer = writer

		// 设置请求ID用于跟踪
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			c.Header("X-Request-ID", requestID)
		}

		// 处理请求
		c.Next()

		// 记录性能指标
		duration := time.Since(start)
		atomic.AddInt64(&pm.metrics.totalRequests, 1)
		atomic.AddInt64(&pm.metrics.totalLatency, duration.Nanoseconds())

		// 检查慢请求
		if duration > pm.slowThreshold {
			atomic.AddInt64(&pm.metrics.slowRequests, 1)
			pm.logSlowRequest(c, duration, writer)
		}

		// 检查错误响应
		if writer.status >= 400 {
			atomic.AddInt64(&pm.metrics.errorRequests, 1)
		}

		// 异步记录到Redis（避免阻塞主流程）
		if pm.redisClient != nil {
			go pm.recordMetricsAsync(c, writer, duration, requestID)
		}
	}
}

// MemoryMonitoringMiddleware 内存监控中间件
func (pm *V2PerformanceMiddleware) MemoryMonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时的内存状态
		var mStats runtime.MemStats
		runtime.ReadMemStats(&mStats)
		initialMemory := mStats.Alloc

		c.Next()

		// 记录请求结束时的内存状态
		runtime.ReadMemStats(&mStats)
		memoryDelta := mStats.Alloc - initialMemory
		finalMemory := mStats.Alloc

		// 更新最大内存使用
		for {
			currentMax := atomic.LoadInt64(&pm.metrics.memoryUsage)
			if currentMax >= int64(finalMemory) {
				break
			}
			if atomic.CompareAndSwapInt64(&pm.metrics.memoryUsage, currentMax, int64(finalMemory)) {
				break
			}
		}

		// 记录GC统计
		gcCount := atomic.LoadUint32(&pm.metrics.gcCount)
		if mStats.NumGC > gcCount {
			atomic.StoreUint32(&pm.metrics.gcCount, mStats.NumGC)
			pm.metrics.mu.Lock()
			pm.metrics.lastGC = time.Now()
			pm.metrics.mu.Unlock()
		}

		// 内存使用警告
		if memoryDelta > 50*1024*1024 { // 50MB
			pm.logHighMemoryUsage(c, memoryDelta, finalMemory)
		}
	}
}

// ConcurrencyControlMiddleware 并发控制中间件 - 基于最新Gin v2
func (pm *V2PerformanceMiddleware) ConcurrencyControlMiddleware(maxConcurrent int) gin.HandlerFunc {
	semaphore := make(chan struct{}, maxConcurrent)

	return func(c *gin.Context) {
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()

			// 检查队列长度
			queueLength := len(semaphore)
			if queueLength > maxConcurrent/2 {
				c.Header("X-Queue-Length", strconv.Itoa(queueLength))
			}

			c.Next()
		default:
			// 达到并发限制
			pm.logConcurrencyLimitExceeded(c, maxConcurrent)

			// 使用最新的Gin格式化JSON响应
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":       "Service temporarily unavailable",
				"code":        http.StatusServiceUnavailable,
				"retry_after": "5",
				"queue_time":  pm.calculateQueueTime(len(semaphore)),
			})
			c.Abort()
		}
	}
}

// TimeoutMiddleware 请求超时中间件 - 基于最新Gin最佳实践
func (pm *V2PerformanceMiddleware) TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用Gin的Context创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换原始请求的上下文
		c.Request = c.Request.WithContext(ctx)

		// 使用通道等待请求完成
		finished := make(chan struct{})
		go func() {
			defer func() {
				if r := recover(); r != nil {
					pm.handlePanic(c, r)
				}
				close(finished)
			}()
			c.Next()
		}()

		select {
		case <-finished:
			// 请求正常完成
		case <-ctx.Done():
			// 请求超时
			pm.logTimeout(c, timeout)

			if !c.Writer.Written() {
				c.JSON(http.StatusRequestTimeout, gin.H{
					"error": "Request timeout",
					"code":  http.StatusRequestTimeout,
					"timeout": timeout.String(),
				})
				c.Abort()
			}
		}
	}
}

// RateLimitingMiddleware 基于Redis的分布式限流中间件
func (pm *V2PerformanceMiddleware) RateLimitingMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	if pm.redisClient == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		clientIP := pm.getClientIP(c)
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 使用Redis的滑动窗口算法
		now := time.Now().Unix()
		windowStart := now - int64(window.Seconds())

		// 清理过期记录
		pm.redisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

		// 检查当前窗口内的请求数
		count, err := pm.redisClient.ZCard(ctx, key).Result()
		if err != nil {
			// Redis错误，记录日志但不阻止请求
			logger.Logger.Error("Redis error in rate limiting", zap.Error(err))
			c.Next()
			return
		}

		if count >= int64(maxRequests) {
			pm.logRateLimitExceeded(c, clientIP, count)

			c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(window.Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        http.StatusTooManyRequests,
				"retry_after": window.String(),
			})
			c.Abort()
			return
		}

		// 记录当前请求
		pm.redisClient.ZAdd(ctx, key, &redis.Z{
			Score:  float64(now),
			Member: fmt.Sprintf("%d", now),
		})
		pm.redisClient.Expire(ctx, key, window)

		// 设置响应头
		remaining := maxRequests - int(count) - 1
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now+int64(window.Seconds()), 10))

		c.Next()
	}
}

// CustomLoggerMiddleware 自定义日志中间件 - 基于最新Gin最佳实践
func (pm *V2PerformanceMiddleware) CustomLoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 自定义日志格式，包含性能指标
		return fmt.Sprintf("[%s] \"%s %s %s %d %s \"%s\" %s | Memory: %d | Latency: %v\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
			pm.getCurrentMemoryUsage(),
			param.Latency,
		)
	})
}

// RecoveryMiddleware 自定义恢复中间件
func (pm *V2PerformanceMiddleware) RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Logger.Error("Panic recovered",
				zap.String("error", err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("client_ip", c.ClientIP()),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  http.StatusInternalServerError,
				"request_id": c.GetHeader("X-Request-ID"),
			})
		} else {
			logger.Logger.Error("Unknown panic recovered",
				zap.Any("panic", recovered),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  http.StatusInternalServerError,
			})
		}
		c.Abort()
	})
}

// GetMetrics 获取性能指标
func (pm *V2PerformanceMiddleware) GetMetrics() map[string]interface{} {
	totalReqs := atomic.LoadInt64(&pm.metrics.totalRequests)
	totalLatency := atomic.LoadInt64(&pm.metrics.totalLatency)
	slowReqs := atomic.LoadInt64(&pm.metrics.slowRequests)
	errorReqs := atomic.LoadInt64(&pm.metrics.errorRequests)
	activeReqs := atomic.LoadInt32(&pm.metrics.totalRequests)
	maxConc := atomic.LoadInt32(&pm.metrics.maxConcurrent)
	memUsage := atomic.LoadInt64(&pm.metrics.memoryUsage)
	gcCount := atomic.LoadUint32(&pm.metrics.gcCount)

	var avgLatency time.Duration
	if totalReqs > 0 {
		avgLatency = time.Duration(totalLatency / totalReqs)
	}

	var errorRate float64
	if totalReqs > 0 {
		errorRate = float64(errorReqs) / float64(totalReqs) * 100
	}

	var slowRate float64
	if totalReqs > 0 {
		slowRate = float64(slowReqs) / float64(totalReqs) * 100
	}

	return map[string]interface{}{
		"total_requests":      totalReqs,
		"slow_requests":       slowReqs,
		"error_requests":      errorReqs,
		"active_requests":     activeReqs,
		"max_concurrent":      maxConc,
		"average_latency":     avgLatency.String(),
		"slow_request_rate":   fmt.Sprintf("%.2f%%", slowRate),
		"error_rate":          fmt.Sprintf("%.2f%%", errorRate),
		"memory_usage_mb":     memUsage / (1024 * 1024),
		"gc_count":           gcCount,
		"uptime":             time.Since(pm.metrics.startTime).String(),
		"enabled":             pm.enabled,
		"slow_threshold":      pm.slowThreshold.String(),
	}
}

// ResetMetrics 重置性能指标
func (pm *V2PerformanceMiddleware) ResetMetrics() {
	atomic.StoreInt64(&pm.metrics.totalRequests, 0)
	atomic.StoreInt64(&pm.metrics.slowRequests, 0)
	atomic.StoreInt64(&pm.metrics.errorRequests, 0)
	atomic.StoreInt64(&pm.metrics.totalLatency, 0)
	atomic.StoreInt32(&pm.metrics.maxConcurrent, 0)
	atomic.StoreInt64(&pm.metrics.memoryUsage, 0)
	atomic.StoreUint32(&pm.metrics.gcCount, 0)
	pm.metrics.startTime = time.Now()
}

// 内部结构体和方法

type responseWriter struct {
	gin.ResponseWriter
	size   int
	status int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.size += size
	return size, err
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// 辅助方法

func (pm *V2PerformanceMiddleware) getClientIP(c *gin.Context) string {
	// 优先使用X-Forwarded-For头
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 使用X-Real-IP头
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	return c.ClientIP()
}

func (pm *V2PerformanceMiddleware) getCurrentMemoryUsage() int {
	var mStats runtime.MemStats
	runtime.ReadMemStats(&mStats)
	return int(mStats.Alloc / (1024 * 1024)) // MB
}

func (pm *V2PerformanceMiddleware) calculateQueueLength(semaphoreLen int) time.Duration {
	// 估算等待时间
	baseDelay := time.Duration(semaphoreLen) * 100 * time.Millisecond
	if baseDelay > 5*time.Second {
		baseDelay = 5 * time.Second
	}
	return baseDelay
}

func (pm *V2PerformanceMiddleware) logSlowRequest(c *gin.Context, duration time.Duration, writer *responseWriter) {
	logger.Logger.Warn("Slow request detected",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Duration("duration", duration),
		zap.Int("status", writer.status),
		zap.Int64("bytes_in", c.Request.ContentLength),
		zap.Int("bytes_out", int64(writer.size)),
		zap.String("client_ip", pm.getClientIP(c)),
		zap.String("request_id", c.GetHeader("X-Request-ID")),
	)
}

func (pm *V2PerformanceMiddleware) logHighMemoryUsage(c *gin.Context, memoryDelta int64, finalMemory uint64) {
	logger.Logger.Warn("High memory usage detected",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Int64("memory_delta_mb", memoryDelta/(1024*1024)),
		zap.Uint64("final_memory_mb", finalMemory/(1024*1024)),
	)
}

func (pm *V2PerformanceMiddleware) logConcurrencyLimitExceeded(c *gin.Context, maxConcurrent int) {
	logger.Logger.Warn("Concurrency limit exceeded",
		zap.Int("max_concurrent", maxConcurrent),
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", pm.getClientIP(c)),
	)
}

func (pm *V2PerformanceMiddleware) logTimeout(c *gin.Context, timeout time.Duration) {
	logger.Logger.Warn("Request timeout",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Duration("timeout", timeout),
		zap.String("client_ip", pm.getClientIP(c)),
	)
}

func (pm *V2PerformanceMiddleware) logRateLimitExceeded(c *gin.Context, clientIP string, count int64) {
	logger.Logger.Warn("Rate limit exceeded",
		zap.String("client_ip", clientIP),
		zap.Int64("count", count),
		zap.String("path", c.Request.URL.Path),
	)
}

func (pm *V2PerformanceMiddleware) handlePanic(c *gin.Context, recovered interface{}) {
	logger.Logger.Error("Panic recovered in middleware",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Any("panic", recovered),
	)
}

func (pm *V2PerformanceMiddleware) recordMetricsAsync(c *gin.Context, writer *responseWriter, duration time.Duration, requestID string) {
	if pm.redisClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("metrics:%s", time.Now().Format("2006-01-02"))

	metrics := map[string]interface{}{
		"timestamp":    time.Now().Unix(),
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"status":       writer.status,
		"duration_ms":  duration.Milliseconds(),
		"bytes_in":     c.Request.ContentLength,
		"bytes_out":    int64(writer.size),
		"user_agent":   c.GetHeader("User-Agent"),
		"client_ip":    pm.getClientIP(c),
		"request_id":   requestID,
	}

	// 使用Redis管道提高性能
	pipe := pm.redisClient.Pipeline()
	pipe.HMSet(ctx, key, metrics)
	pipe.Expire(ctx, key, 24*time.Hour)
	pipe.Exec(ctx)
}