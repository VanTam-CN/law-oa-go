package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/logger"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL             time.Duration
	SkipHeader      string
	RedisClient     *redis.Client
	KeyPrefix       string
	MaxBodySize     int64
	SkipRoutes      []string
	CacheableRoutes []string
}

// CacheResponse 缓存响应结构
type CacheResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

// CacheMiddleware 优化的缓存中间件
func CacheMiddleware(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否应该跳过缓存
		if shouldSkipCache(c, config) {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := generateCacheKey(c, config.KeyPrefix)

		// 尝试从缓存获取
		if cached, found := getCachedResponse(c, cacheKey, config); found {
			writeCachedResponse(c, cached)
			return
		}

		// 记录响应以便缓存
		captureResponse(c, cacheKey, config)
	}
}

// shouldSkipCache 检查是否应该跳过缓存
func shouldSkipCache(c *gin.Context, config CacheConfig) bool {
	// 检查跳过头部
	if config.SkipHeader != "" && c.GetHeader(config.SkipHeader) != "" {
		return true
	}

	// 只缓存GET请求
	if c.Request.Method != http.MethodGet {
		return true
	}

	// 检查跳过的路由
	for _, route := range config.SkipRoutes {
		if strings.HasPrefix(c.Request.URL.Path, route) {
			return true
		}
	}

	// 检查是否为可缓存路由
	if len(config.CacheableRoutes) > 0 {
		cacheable := false
		for _, route := range config.CacheableRoutes {
			if strings.HasPrefix(c.Request.URL.Path, route) {
				cacheable = true
				break
			}
		}
		if !cacheable {
			return true
		}
	}

	// 检查查询参数是否过多（超过5个参数不缓存）
	if len(c.Request.URL.Query()) > 5 {
		return true
	}

	return false
}

// generateCacheKey 生成缓存键
func generateCacheKey(c *gin.Context, keyPrefix string) string {
	// 路径 + 查询参数 + 用户信息（如果有）
	key := keyPrefix + ":" + c.Request.URL.Path

	// 添加查询参数
	if query := c.Request.URL.RawQuery; query != "" {
		key += ":" + query
	}

	// 添加用户ID（如果已认证）
	if userID, exists := c.Get("user_id"); exists {
		key += ":user:" + fmt.Sprintf("%v", userID)
	}

	return key
}

// getCachedResponse 从缓存获取响应
func getCachedResponse(c *gin.Context, key string, config CacheConfig) (*CacheResponse, bool) {
	// 优先使用新的Redis缓存
	if config.RedisClient != nil {
		return getCachedResponseFromRedis(c, key, config)
	}

	// 回退到旧的缓存服务
	if cache.DefaultCacheService != nil {
		var cachedData interface{}
		if err := cache.DefaultCacheService.Get(key, &cachedData); err == nil {
			// 转换为新的格式
			cached := &CacheResponse{
				Status:  200,
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    mustMarshalJSON(cachedData),
			}
			return cached, true
		}
	}

	return nil, false
}

// getCachedResponseFromRedis 从Redis获取缓存的响应
func getCachedResponseFromRedis(c *gin.Context, key string, config CacheConfig) (*CacheResponse, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data, err := config.RedisClient.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			logger.Logger.Error("Cache get error",
				zap.Error(err),
				zap.String("key", key),
			)
		}
		return nil, false
	}

	var cached CacheResponse
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		logger.Logger.Error("Cache unmarshal error",
			zap.Error(err),
			zap.String("key", key),
		)
		return nil, false
	}

	logger.Logger.Debug("Cache hit",
		zap.String("key", key),
		zap.Int("status", cached.Status),
	)

	return &cached, true
}

// writeCachedResponse 写入缓存的响应
func writeCachedResponse(c *gin.Context, cached *CacheResponse) {
	// 写入状态码
	c.Status(cached.Status)

	// 写入头部
	for key, value := range cached.Headers {
		if key != "Content-Length" { // 让HTTP服务器自动设置Content-Length
			c.Header(key, value)
		}
	}

	// 写入主体
	c.Writer.Write(cached.Body)

	logger.Logger.Info("Served from cache",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
	)
}

// captureResponse 捕获响应并缓存
func captureResponse(c *gin.Context, cacheKey string, config CacheConfig) {
	// 创建响应写入器包装器
	w := &responseWriter{
		ResponseWriter: c.Writer,
		buffer:         bytes.NewBuffer(nil),
		status:         http.StatusOK,
		headers:        make(map[string]string),
	}

	c.Writer = w

	// 继续处理请求
	c.Next()

	// 根据响应状态决定是否缓存
	if shouldCacheResponse(c, w, config) {
		go cacheResponse(cacheKey, w, config)
	}
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	gin.ResponseWriter
	buffer  *bytes.Buffer
	status  int
	headers map[string]string
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.buffer.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Header() http.Header {
	// 捕获需要缓存的头部
	originalHeader := w.ResponseWriter.Header()
	// 注意：headerCapture 结构体已保留但当前未使用，可用于未来的头部捕获功能
	return originalHeader
}

// headerCapture 头部捕获器
type headerCapture struct {
	http.Header
	headers map[string]string
}

func (h *headerCapture) Set(key, value string) {
	h.Header.Set(key, value)
	// 捕获重要的响应头
	if shouldCacheHeader(key) {
		h.headers[key] = value
	}
}

// shouldCacheHeader 检查是否应该缓存头部
func shouldCacheHeader(key string) bool {
	cacheableHeaders := map[string]bool{
		"Content-Type":     true,
		"Content-Encoding": true,
		"Cache-Control":    true,
		"ETag":             true,
		"Last-Modified":    true,
		"Vary":             true,
	}
	return cacheableHeaders[strings.ToLower(key)]
}

// shouldCacheResponse 检查是否应该缓存响应
func shouldCacheResponse(c *gin.Context, w *responseWriter, config CacheConfig) bool {
	// 只缓存成功的响应
	if w.status < 200 || w.status >= 300 {
		return false
	}

	// 检查响应大小
	if w.buffer.Len() > int(config.MaxBodySize) {
		return false
	}

	// 检查响应类型
	contentType := w.headers["Content-Type"]
	if !shouldCacheContentType(contentType) {
		return false
	}

	return true
}

// shouldCacheContentType 检查是否应该缓存内容类型
func shouldCacheContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	cacheableTypes := []string{
		"application/json",
		"text/html",
		"text/plain",
		"application/xml",
		"text/xml",
	}

	for _, ct := range cacheableTypes {
		if strings.HasPrefix(contentType, ct) {
			return true
		}
	}

	return false
}

// cacheResponse 异步缓存响应
func cacheResponse(key string, w *responseWriter, config CacheConfig) {
	// 如果有Redis客户端，使用新的缓存方式
	if config.RedisClient != nil {
		cacheResponseToRedis(key, w, config)
		return
	}

	// 回退到旧的缓存服务
	if cache.DefaultCacheService != nil && w.status == 200 {
		var responseData interface{}
		if err := json.Unmarshal(w.buffer.Bytes(), &responseData); err == nil {
			cache.DefaultCacheService.Set(key, responseData, config.TTL)
		}
	}
}

// cacheResponseToRedis 缓存响应到Redis
func cacheResponseToRedis(key string, w *responseWriter, config CacheConfig) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 创建缓存响应
	cached := CacheResponse{
		Status:  w.status,
		Headers: w.headers,
		Body:    w.buffer.Bytes(),
	}

	// 序列化
	data, err := json.Marshal(cached)
	if err != nil {
		logger.Logger.Error("Cache marshal error",
			zap.Error(err),
			zap.String("key", key),
		)
		return
	}

	// 计算TTL
	ttl := calculateTTL(w.headers, config.TTL)

	// 存储到Redis
	if err := config.RedisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		logger.Logger.Error("Cache set error",
			zap.Error(err),
			zap.String("key", key),
		)
		return
	}

	logger.Logger.Info("Response cached",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
		zap.Int("size", len(data)),
	)
}

// calculateTTL 计算缓存TTL
func calculateTTL(headers map[string]string, defaultTTL time.Duration) time.Duration {
	// 检查Cache-Control头部
	if cacheControl := headers["Cache-Control"]; cacheControl != "" {
		if strings.Contains(cacheControl, "no-cache") || strings.Contains(cacheControl, "no-store") {
			return time.Minute // 最小缓存时间
		}

		// 解析max-age
		parts := strings.Split(cacheControl, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "max-age=") {
				ageStr := strings.TrimPrefix(part, "max-age=")
				if age, err := strconv.Atoi(ageStr); err == nil {
					return time.Duration(age) * time.Second
				}
			}
		}
	}

	return defaultTTL
}

// InvalidateCache 缓存失效函数
func InvalidateCache(redisClient *redis.Client, keyPrefix string, patterns ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for _, pattern := range patterns {
		key := keyPrefix + ":" + pattern + "*"
		keys, err := redisClient.Keys(ctx, key).Result()
		if err != nil {
			logger.Logger.Error("Cache invalidation error",
				zap.Error(err),
				zap.String("pattern", pattern),
			)
			continue
		}

		if len(keys) > 0 {
			if err := redisClient.Del(ctx, keys...).Err(); err != nil {
				logger.Logger.Error("Cache deletion error",
					zap.Error(err),
					zap.Strings("keys", keys),
				)
			} else {
				logger.Logger.Info("Cache invalidated",
					zap.String("pattern", pattern),
					zap.Int("count", len(keys)),
				)
			}
		}
	}
}

// mustMarshalJSON 序列化JSON，失败时返回空字节
func mustMarshalJSON(data interface{}) []byte {
	if bytes, err := json.Marshal(data); err == nil {
		return bytes
	}
	return []byte("{}")
}

// OptimizedRateLimiter 优化的限流中间件
type OptimizedRateLimiter struct {
	redisClient   *redis.Client
	window        time.Duration
	limit         int64
	keyPrefix     string
	localCache    map[string]*rateLimitInfo
	cacheExpiry   time.Duration
	mutex         sync.RWMutex
	cleanupTicker *time.Ticker
}

// rateLimitInfo 本地限流信息
type rateLimitInfo struct {
	Count     int64
	WindowEnd time.Time
}

// NewOptimizedRateLimiter 创建优化的限流器
func NewOptimizedRateLimiter(redisClient *redis.Client, keyPrefix string, limit int64, window time.Duration) *OptimizedRateLimiter {
	limiter := &OptimizedRateLimiter{
		redisClient: redisClient,
		window:      window,
		limit:       limit,
		keyPrefix:   keyPrefix,
		localCache:  make(map[string]*rateLimitInfo),
		cacheExpiry: window / 10, // 缓存窗口的1/10
	}

	// 启动清理协程
	limiter.cleanupTicker = time.NewTicker(limiter.cacheExpiry)
	go limiter.cleanupLocalCache()

	return limiter
}

// CheckRateLimit 检查限流
func (rl *OptimizedRateLimiter) CheckRateLimit(c *gin.Context) bool {
	key := rl.generateKey(c)

	// 首先检查本地缓存
	if rl.checkLocalCache(key) {
		return false // 超出限制
	}

	// 然后检查Redis
	if rl.redisClient != nil {
		return rl.checkRedisLimit(key)
	}

	// 如果没有Redis，使用本地限流
	return rl.checkLocalOnlyLimit(key)
}

// generateKey 生成限流键
func (rl *OptimizedRateLimiter) generateKey(c *gin.Context) string {
	ip := c.ClientIP()
	userID, exists := c.Get("user_id")
	if exists {
		return rl.keyPrefix + ":user:" + fmt.Sprintf("%v", userID)
	}
	return rl.keyPrefix + ":ip:" + ip
}

// checkLocalCache 检查本地缓存
func (rl *OptimizedRateLimiter) checkLocalCache(key string) bool {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	info, exists := rl.localCache[key]
	if !exists {
		return false
	}

	// 检查窗口是否过期
	if time.Now().After(info.WindowEnd) {
		return false
	}

	return info.Count >= rl.limit
}

// checkRedisLimit 检查Redis限流
func (rl *OptimizedRateLimiter) checkRedisLimit(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	pipe := rl.redisClient.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Logger.Error("Rate limit check error", zap.Error(err))
		return false // 出错时不限流
	}

	currentCount := incr.Val()

	// 更新本地缓存
	rl.updateLocalCache(key, currentCount)

	return currentCount > rl.limit
}

// checkLocalOnlyLimit 仅本地限流
func (rl *OptimizedRateLimiter) checkLocalOnlyLimit(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	info, exists := rl.localCache[key]

	if !exists || now.After(info.WindowEnd) {
		// 新窗口
		rl.localCache[key] = &rateLimitInfo{
			Count:     1,
			WindowEnd: now.Add(rl.window),
		}
		return false
	}

	info.Count++
	return info.Count > rl.limit
}

// updateLocalCache 更新本地缓存
func (rl *OptimizedRateLimiter) updateLocalCache(key string, count int64) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.localCache[key] = &rateLimitInfo{
		Count:     count,
		WindowEnd: time.Now().Add(rl.window),
	}
}

// cleanupLocalCache 清理过期的本地缓存
func (rl *OptimizedRateLimiter) cleanupLocalCache() {
	for range rl.cleanupTicker.C {
		rl.mutex.Lock()
		now := time.Now()
		for key, info := range rl.localCache {
			if now.After(info.WindowEnd) {
				delete(rl.localCache, key)
			}
		}
		rl.mutex.Unlock()
	}
}

// Stop 停止限流器
func (rl *OptimizedRateLimiter) Stop() {
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
}
