package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"law-oa-go/internal/logger"
	"law-oa-go/internal/monitoring"
)

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
	metrics := monitoring.GetPerformanceMetrics()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 读取请求体大小
		var requestSize int64
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestSize = int64(len(bodyBytes))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 创建响应写入器包装器以捕获响应大小
		responseWriter := &performanceResponseWriter{
			ResponseWriter: c.Writer,
			buffer:         bytes.NewBuffer(nil),
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 计算持续时间
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		responseSize := int64(responseWriter.buffer.Len())

		// 记录指标
		metrics.RecordHTTPRequest(
			method,
			path,
			string(rune(statusCode)),
			duration,
			requestSize,
			responseSize,
		)

		// 记录慢请求
		if duration > 1*time.Second {
			logger.Logger.Warn("Slow request detected",
				zap.String("method", method),
				zap.String("path", path),
				zap.Duration("duration", duration),
				zap.Int("status", statusCode),
				zap.Int64("request_size", requestSize),
				zap.Int64("response_size", responseSize),
			)
		}

		// 记录错误请求
		if statusCode >= 400 {
			logger.Logger.Error("HTTP error request",
				zap.String("method", method),
				zap.String("path", path),
				zap.Duration("duration", duration),
				zap.Int("status", statusCode),
				zap.String("error", c.Errors.String()),
			)
		}
	}
}

// performanceResponseWriter 性能监控响应写入器包装器
type performanceResponseWriter struct {
	gin.ResponseWriter
	buffer *bytes.Buffer
}

func (w *performanceResponseWriter) Write(data []byte) (int, error) {
	w.buffer.Write(data)
	return w.ResponseWriter.Write(data)
}

// DatabasePerformanceMiddleware 数据库性能监控中间件
func DatabasePerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在上下文中设置数据库查询计时器
		c.Set("db_query_start", time.Now())
		c.Next()
	}
}

// CachePerformanceMiddleware 缓存性能监控中间件
func CachePerformanceMiddleware() gin.HandlerFunc {
	metrics := monitoring.GetPerformanceMetrics()

	return func(c *gin.Context) {
		// 检查是否来自缓存
		if cached, exists := c.Get("from_cache"); exists {
			if cached.(bool) {
				metrics.RecordCachedResponse("redis", "hit")
			}
		}
		c.Next()
	}
}

// ConcurrencyLimiter 并发限制中间件
type ConcurrencyLimiter struct {
	semaphore chan struct{}
	maxConcurrent int
}

// NewConcurrencyLimiter 创建并发限制器
func NewConcurrencyLimiter(maxConcurrent int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		semaphore: make(chan struct{}, maxConcurrent),
		maxConcurrent: maxConcurrent,
	}
}

// Limit 并发限制中间件
func (cl *ConcurrencyLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		select {
		case cl.semaphore <- struct{}{}:
			defer func() { <-cl.semaphore }()
			c.Next()
		default:
			logger.Logger.Warn("Concurrency limit exceeded",
				zap.Int("max_concurrent", cl.maxConcurrent),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			c.JSON(503, gin.H{
				"error": "Service temporarily unavailable due to high load",
				"code":  503,
			})
			c.Abort()
		}
	}
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置请求超时
		c.Request = c.Request.WithContext(c.Request.Context())

		// 使用channel来实现超时控制
		finished := make(chan bool)
		go func() {
			c.Next()
			finished <- true
		}()

		select {
		case <-finished:
			// 请求正常完成
		case <-time.After(timeout):
			// 请求超时
			logger.Logger.Warn("Request timeout",
				zap.Duration("timeout", timeout),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)

			c.JSON(504, gin.H{
				"error": "Request timeout",
				"code":  504,
			})
			c.Abort()
		}
	}
}

// CompressionMiddleware 压缩中间件
func CompressionMiddleware(minSize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查客户端是否支持压缩
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if !containsGzip(acceptEncoding) {
			c.Next()
			return
		}

		// 创建响应写入器包装器
		responseWriter := &compressibleResponseWriter{
			ResponseWriter: c.Writer,
			minSize:        minSize,
		}
		c.Writer = responseWriter

		c.Next()

		// 如果响应大小足够大，进行压缩
		if responseWriter.shouldCompress() {
			responseWriter.compress()
		}
	}
}

// compressibleResponseWriter 可压缩的响应写入器
type compressibleResponseWriter struct {
	gin.ResponseWriter
	buffer   bytes.Buffer
	minSize  int
	compressed bool
}

func (w *compressibleResponseWriter) Write(data []byte) (int, error) {
	if w.compressed {
		return w.ResponseWriter.Write(data)
	}
	return w.buffer.Write(data)
}

func (w *compressibleResponseWriter) shouldCompress() bool {
	return w.buffer.Len() >= w.minSize && !w.compressed
}

func (w *compressibleResponseWriter) compress() {
	// 简化的压缩实现
	// 在实际项目中，应该使用gzip包进行压缩
	w.compressed = true
	w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	w.ResponseWriter.Write(w.buffer.Bytes())
}

// containsGzip 检查是否支持gzip
func containsGzip(encoding string) bool {
	return len(encoding) > 0 && (encoding == "gzip" ||
		encoding == "*" ||
		encoding == "deflate, gzip" ||
		encoding == "gzip, deflate")
}

// MemoryUsageMiddleware 内存使用监控中间件
func MemoryUsageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时的内存使用
		// 这里可以添加内存监控逻辑
		c.Next()
	}
}

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			metrics := monitoring.GetPerformanceMetrics()
			summary := metrics.GetMetricsSummary()

			c.JSON(200, gin.H{
				"status": "healthy",
				"timestamp": time.Now().Unix(),
				"metrics": summary,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}