package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"law-oa-go/internal/cache"
)

// 缓存相关的Prometheus指标
var (
	cacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_middleware_hits_total",
		Help: "Total number of cache hits from middleware",
	}, []string{"endpoint"})

	cacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_middleware_misses_total",
		Help: "Total number of cache misses from middleware",
	}, []string{"endpoint"})

	cacheResponseDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "cache_response_duration_seconds",
		Help:    "Duration of cache operations",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "endpoint"})
)

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL           time.Duration
	SkipHeader    string
	KeyGenerator  func(*gin.Context) string
	ShouldCache   func(*gin.Context) bool
}

// CacheMiddleware 缓存中间件
func CacheMiddleware(config CacheConfig) gin.HandlerFunc {
	if config.TTL == 0 {
		config.TTL = 5 * time.Minute // 默认5分钟
	}
	
	if config.SkipHeader == "" {
		config.SkipHeader = "X-Cache-Skip"
	}

	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator
	}

	if config.ShouldCache == nil {
		config.ShouldCache = defaultShouldCache
	}

	return func(c *gin.Context) {
		start := time.Now()
		endpoint := getEndpointKey(c)

		// 检查是否应该跳过缓存
		if c.GetHeader(config.SkipHeader) != "" || !config.ShouldCache(c) {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := config.KeyGenerator(c)

		// 尝试从缓存获取
		var cachedResponse []byte
		err := cache.DefaultCacheService.Get(c.Request.Context(), cacheKey, &cachedResponse)
		
		if err == nil {
			// 缓存命中
			cacheHits.WithLabelValues(endpoint).Inc()
			
			// 设置缓存头
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Key", cacheKey)
			
			// 解析缓存的响应
			var response map[string]interface{}
			if err := json.Unmarshal(cachedResponse, &response); err == nil {
				// 设置响应头
				if status, ok := response["status"].(float64); ok {
					c.Status(int(status))
				}
				
				// 发送响应
				c.Data(http.StatusOK, "application/json", cachedResponse)
				c.Abort()
				return
			}
		} else {
			// 缓存未命中
			cacheMisses.WithLabelValues(endpoint).Inc()
			c.Header("X-Cache", "MISS")
		}

		// 创建一个响应写入器来捕获响应
		writer := &responseWriter{ResponseWriter: c.Writer, body: []byte{}, statusCode: 200}
		c.Writer = writer

		// 继续处理请求
		c.Next()

		// 只缓存成功的GET请求
		if c.Request.Method == "GET" && writer.statusCode == http.StatusOK {
			// 构建响应数据
			response := map[string]interface{}{
				"status": writer.statusCode,
				"data":    writer.body,
				"headers": c.Writer.Header(),
			}

			// 缓存响应
			responseData, _ := json.Marshal(response)
			if err := cache.DefaultCacheService.Set(c.Request.Context(), cacheKey, responseData, config.TTL); err != nil {
				// 缓存设置失败不影响主要功能
				fmt.Printf("Warning: failed to cache response for %s: %v\n", cacheKey, err)
			}
		}

		// 记录操作耗时
		cacheResponseDuration.WithLabelValues("set", endpoint).Observe(time.Since(start).Seconds())
	}
}

// responseWriter 用于捕获响应的写入器
type responseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body = append(w.body, s...)
	return w.ResponseWriter.WriteString(s)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// defaultKeyGenerator 默认的缓存键生成器
func defaultKeyGenerator(c *gin.Context) string {
	// 构建基础键: method:path
	key := fmt.Sprintf("%s:%s", c.Request.Method, c.FullPath())
	
	// 添加查询参数
	if c.Request.URL.RawQuery != "" {
		key += ":" + c.Request.URL.RawQuery
	}
	
	// 添加用户ID（如果已认证）
	if userID := getUserID(c); userID > 0 {
		key += ":" + strconv.FormatUint(uint64(userID), 10)
	}
	
	return cache.APIKeyGenerator.GenerateKey("response", key)
}

// defaultShouldCache 默认的缓存条件检查
func defaultShouldCache(c *gin.Context) bool {
	// 只缓存GET请求
	if c.Request.Method != "GET" {
		return false
	}
	
	// 不缓存包含特定路径的请求
	skipPaths := []string{"/api/auth", "/api/upload", "/api/export"}
	for _, path := range skipPaths {
		if strings.Contains(c.Request.URL.Path, path) {
			return false
		}
	}
	
	// 只缓存成功的响应
	return c.Writer.Status() == http.StatusOK
}

// getUserID 从上下文获取用户ID
func getUserID(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// getEndpointKey 获取端点键用于监控
func getEndpointKey(c *gin.Context) string {
	// 提取路径用于监控，移除ID参数
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	
	// 简化路径，例如 /api/users/123 -> /api/users/{id}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if _, err := strconv.Atoi(part); err == nil {
			parts[i] = "{id}"
		}
	}
	
	return strings.Join(parts, "/")
}

// CacheByUser 按用户缓存的中间件
func CacheByUser(ttl time.Duration) gin.HandlerFunc {
	return CacheMiddleware(CacheConfig{
		TTL: ttl,
		KeyGenerator: func(c *gin.Context) string {
			userID := getUserID(c)
			if userID == 0 {
				return defaultKeyGenerator(c)
			}
			return cache.APIKeyGenerator.GenerateKey("user", strconv.FormatUint(uint64(userID), 10), c.FullPath())
		},
	})
}

// CacheByRole 按角色缓存的中间件
func CacheByRole(ttl time.Duration) gin.HandlerFunc {
	return CacheMiddleware(CacheConfig{
		TTL: ttl,
		KeyGenerator: func(c *gin.Context) string {
			role := c.GetHeader("X-User-Role")
			if role == "" {
				role = "guest"
			}
			return cache.APIKeyGenerator.GenerateKey("role", role, c.FullPath())
		},
	})
}

// CacheQueryResults 缓存查询结果的工具函数
func CacheQueryResults(ctx context.Context, key string, ttl time.Duration, fetchFunc func() (interface{}, error), dest interface{}) error {
	start := time.Now()
	endpoint := "query"
	
	// 尝试从缓存获取
	if err := cache.DefaultCacheService.Get(ctx, key, dest); err == nil {
		cacheHits.WithLabelValues(endpoint).Inc()
		cacheResponseDuration.WithLabelValues("get", endpoint).Observe(time.Since(start).Seconds())
		return nil
	}
	
	// 缓存未命中
	cacheMisses.WithLabelValues(endpoint).Inc()
	
	// 获取数据
	value, err := fetchFunc()
	if err != nil {
		return err
	}
	
	// 设置缓存
	if err := cache.DefaultCacheService.Set(ctx, key, value, ttl); err != nil {
		fmt.Printf("Warning: failed to cache query results for %s: %v\n", key, err)
	}
	
	// 将值设置到目标变量
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	cacheResponseDuration.WithLabelValues("set", endpoint).Observe(time.Since(start).Seconds())
	return json.Unmarshal(data, dest)
}

// InvalidateCache 使缓存失效的工具函数
func InvalidateCache(ctx context.Context, patterns []string) error {
	for _, pattern := range patterns {
		if err := cache.DefaultCacheService.ClearPattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to clear cache pattern %s: %w", pattern, err)
		}
	}
	return nil
}

// GenerateCacheKey 生成缓存的通用键
func GenerateCacheKey(prefix string, parts ...string) string {
	allParts := append([]string{prefix}, parts...)
	return cache.APIKeyGenerator.GenerateKey(allParts...)
}