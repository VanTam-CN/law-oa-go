package middleware

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// PerformanceCacheConfig 性能缓存配置 - 避免与cache.go中的CacheConfig冲突
type PerformanceCacheConfig struct {
	TTL             time.Duration
	SkipHeader      string
	RedisClient     *redis.Client
	KeyPrefix       string
	MaxBodySize     int64
	SkipRoutes      []string
	CacheableRoutes []string
	DefaultTTL      time.Duration
	Compression     bool
}

// CacheableResponse 可缓存响应
type CacheableResponse struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	ETag      string      `json:"etag"`
}

// PerformanceCache 性能缓存中间件
type PerformanceCache struct {
	config PerformanceCacheConfig
}

// NewPerformanceCache 创建性能缓存中间件
func NewPerformanceCache(config PerformanceCacheConfig) *PerformanceCache {
	// 设置默认值
	if config.TTL == 0 {
		config.TTL = 5 * time.Minute
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 5 * time.Minute
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "lawoa"
	}
	if config.MaxBodySize == 0 {
		config.MaxBodySize = 1024 * 1024 // 1MB
	}

	return &PerformanceCache{config: config}
}

// CacheMiddleware 缓存中间件
func (pc *PerformanceCache) CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只处理GET请求
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// 检查是否跳过缓存
		if pc.shouldSkipCache(c) {
			c.Next()
			return
		}

		// 生成缓存键
		cacheKey := pc.generateCacheKey(c)

		// 尝试从缓存获取
		if cachedData := pc.getFromCache(c, cacheKey); cachedData != nil {
			pc.serveFromCache(c, cachedData)
			return
		}

		// 记录响应以便后续缓存
		cacheWriter := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			buffer:         make([]byte, 0),
			statusCode:     http.StatusOK,
		}
		c.Writer = cacheWriter

		c.Next()

		// 缓存响应
		pc.cacheResponse(c, cacheKey, cacheWriter)
	}
}

// InvalidateCacheMiddleware 缓存失效中间件
func (pc *PerformanceCache) InvalidateCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 对于修改操作，清除相关缓存
		if pc.shouldInvalidateCache(c) {
			pc.invalidateRelatedCache(c)
		}
	}
}

// CacheResponseWriter 缓存响应写入器
type CacheResponseWriter struct {
	gin.ResponseWriter
	buffer     []byte
	statusCode int
	written    bool
}

func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.statusCode = w.Status()
		w.written = true
	}
	w.buffer = append(w.buffer, data...)
	return w.ResponseWriter.Write(data)
}

// shouldSkipCache 检查是否应该跳过缓存
func (pc *PerformanceCache) shouldSkipCache(c *gin.Context) bool {
	// 检查请求头
	if c.GetHeader(pc.config.SkipHeader) != "" {
		return true
	}

	// 检查是否为认证相关请求
	if strings.HasPrefix(c.Request.URL.Path, "/auth") {
		return true
	}

	// 检查跳过路由
	for _, route := range pc.config.SkipRoutes {
		if strings.HasPrefix(c.Request.URL.Path, route) {
			return true
		}
	}

	// 检查请求参数是否包含跳过缓存标志
	if c.Query("no_cache") == "1" || c.Query("nocache") == "1" {
		return true
	}

	return false
}

// shouldInvalidateCache 检查是否应该清除缓存
func (pc *PerformanceCache) shouldInvalidateCache(c *gin.Context) bool {
	method := c.Request.Method
	path := c.Request.URL.Path

	// POST、PUT、DELETE操作需要清除相关缓存
	if method == "POST" || method == "PUT" || method == "DELETE" {
		return true
	}

	// 特定的路径需要清除缓存
	for _, route := range []string{"/cases", "/clients", "/lawyers"} {
		if strings.HasPrefix(path, route) {
			return true
		}
	}

	return false
}

// generateCacheKey 生成缓存键
func (pc *PerformanceCache) generateCacheKey(c *gin.Context) string {
	path := c.Request.URL.Path
	query := c.Request.URL.Query()

	// 创建查询字符串
	queryStr := query.Encode()
	if queryStr != "" {
		path += "?" + queryStr
	}

	// 添加用户信息到缓存键（如果已认证）
	userID := c.GetString("user_id")
	if userID != "" {
		path = fmt.Sprintf("user:%s:%s", userID, path)
	}

	// 生成MD5哈希
	hash := md5.Sum([]byte(path))
	return fmt.Sprintf("%s:cache:%x", pc.config.KeyPrefix, hash)
}

// getFromCache 从缓存获取数据
func (pc *PerformanceCache) getFromCache(c *gin.Context, cacheKey string) *CacheableResponse {
	if pc.config.RedisClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	data, err := pc.config.RedisClient.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil
	}

	var cachedResponse CacheableResponse
	err = json.Unmarshal([]byte(data), &cachedResponse)
	if err != nil {
		return nil
	}

	// 检查缓存是否过期
	if time.Since(cachedResponse.Timestamp) > pc.config.TTL {
		pc.config.RedisClient.Del(ctx, cacheKey)
		return nil
	}

	// 检查ETag
	if cachedResponse.ETag != "" {
		clientETag := c.GetHeader("If-None-Match")
		if clientETag == cachedResponse.ETag {
			c.Status(http.StatusNotModified)
			return nil
		}
	}

	return &cachedResponse
}

// serveFromCache 从缓存提供响应
func (pc *PerformanceCache) serveFromCache(c *gin.Context, cachedData *CacheableResponse) {
	// 设置缓存头
	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", int(pc.config.TTL.Seconds())))
	if cachedData.ETag != "" {
		c.Header("ETag", cachedData.ETag)
	}
	c.Header("X-Cache", "HIT")

	// 设置响应头
	c.JSON(http.StatusOK, cachedData.Data)
}

// cacheResponse 缓存响应
func (pc *PerformanceCache) cacheResponse(c *gin.Context, cacheKey string, writer *CacheResponseWriter) {
	// 检查响应是否可缓存
	if !pc.isCacheable(c, writer) {
		return
	}

	if pc.config.RedisClient == nil {
		return
	}

	// 生成ETag
	etag := pc.generateETag(writer.buffer)

	// 创建缓存响应
	cacheableResponse := CacheableResponse{
		Data:      pc.extractResponseData(writer.buffer),
		Timestamp: time.Now(),
		ETag:      etag,
	}

	// 序列化缓存数据
	data, err := json.Marshal(cacheableResponse)
	if err != nil {
		return
	}

	// 异步存储到Redis
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	pc.config.RedisClient.Set(ctx, cacheKey, data, pc.config.TTL)
}

// isCacheable 检查响应是否可缓存
func (pc *PerformanceCache) isCacheable(c *gin.Context, writer *CacheResponseWriter) bool {
	// 只缓存成功响应
	if writer.statusCode != http.StatusOK {
		return false
	}

	// 检查响应大小
	if int64(len(writer.buffer)) > pc.config.MaxBodySize {
		return false
	}

	// 检查内容类型
	contentType := c.Writer.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return false
	}

	// 检查是否为可缓存路由
	for _, route := range pc.config.CacheableRoutes {
		if strings.HasPrefix(c.Request.URL.Path, route) {
			return true
		}
	}

	// 默认情况下，GET请求都可以缓存
	return c.Request.Method == "GET"
}

// generateETag 生成ETag
func (pc *PerformanceCache) generateETag(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf(`"%x"`, hash)
}

// extractResponseData 提取响应数据
func (pc *PerformanceCache) extractResponseData(buffer []byte) interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(buffer, &response)
	if err != nil {
		return nil
	}

	// 返回data字段
	if data, ok := response["data"]; ok {
		return data
	}

	return response
}

// invalidateRelatedCache 清除相关缓存
func (pc *PerformanceCache) invalidateRelatedCache(c *gin.Context) {
	if pc.config.RedisClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	path := c.Request.URL.Path

	// 生成模式匹配的缓存键
	patterns := []string{
		fmt.Sprintf("%s:cache:*", pc.config.KeyPrefix),
	}

	// 根据路径添加特定模式
	if strings.HasPrefix(path, "/cases") {
		patterns = append(patterns, fmt.Sprintf("%s:cache:*cases*", pc.config.KeyPrefix))
	} else if strings.HasPrefix(path, "/clients") {
		patterns = append(patterns, fmt.Sprintf("%s:cache:*clients*", pc.config.KeyPrefix))
	} else if strings.HasPrefix(path, "/lawyers") {
		patterns = append(patterns, fmt.Sprintf("%s:cache:*lawyers*", pc.config.KeyPrefix))
	}

	// 使用SCAN命令找到并删除相关缓存
	for _, pattern := range patterns {
		iter := pc.config.RedisClient.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			key := iter.Val()
			pc.config.RedisClient.Del(ctx, key)
		}
	}
}

// WarmupCache 预热缓存
func (pc *PerformanceCache) WarmupCache() error {
	if pc.config.RedisClient == nil {
		return fmt.Errorf("Redis客户端未配置")
	}

	// 这里可以添加预热逻辑
	// 例如：预加载常用数据、热门案件列表等
	ctx := context.Background()

	// 预热案件列表
	casesKey := fmt.Sprintf("%s:cache:%x", pc.config.KeyPrefix, md5.Sum([]byte("/cases")))

	// 这里应该实际调用内部API而不是HTTP请求
	// 为了演示，我们只是记录预热操作
	pc.config.RedisClient.Set(ctx, casesKey+":warmup", "warmed", pc.config.TTL)

	return nil
}

// GetCacheStats 获取缓存统计信息
func (pc *PerformanceCache) GetCacheStats() map[string]interface{} {
	if pc.config.RedisClient == nil {
		return map[string]interface{}{
			"cache_enabled": false,
			"message":       "Redis未配置",
		}
	}

	ctx := context.Background()

	// 获取所有缓存键
	iter := pc.config.RedisClient.Scan(ctx, 0, fmt.Sprintf("%s:cache:*", pc.config.KeyPrefix), 0).Iterator()

	keyCount := 0
	totalSize := int64(0)

	for iter.Next(ctx) {
		keyCount++
		// 获取键大小（近似值）
		size, _ := pc.config.RedisClient.StrLen(ctx, iter.Val()).Result()
		totalSize += size
	}

	return map[string]interface{}{
		"cache_enabled":    true,
		"total_keys":       keyCount,
		"total_size_bytes": totalSize,
		"ttl_seconds":      pc.config.TTL.Seconds(),
		"compression":      pc.config.Compression,
	}
}

// ClearAllCache 清除所有缓存
func (pc *PerformanceCache) ClearAllCache() error {
	if pc.config.RedisClient == nil {
		return fmt.Errorf("Redis客户端未配置")
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("%s:cache:*", pc.config.KeyPrefix)

	iter := pc.config.RedisClient.Scan(ctx, 0, pattern, 0).Iterator()
	deletedCount := 0

	for iter.Next(ctx) {
		err := pc.config.RedisClient.Del(ctx, iter.Val()).Err()
		if err == nil {
			deletedCount++
		}
	}

	// 记录清除操作
	fmt.Printf("缓存清除完成，删除了 %d 个键\n", deletedCount)

	return nil
}

// 便捷函数：创建默认缓存配置
func DefaultCacheConfig(redisClient *redis.Client) PerformanceCacheConfig {
	return PerformanceCacheConfig{
		TTL:             5 * time.Minute,
		SkipHeader:      "X-Cache-Skip",
		RedisClient:     redisClient,
		KeyPrefix:       "lawoa",
		MaxBodySize:     1024 * 1024,
		SkipRoutes:      []string{"/api/auth", "/api/upload", "/api/file"},
		CacheableRoutes: []string{"/api/dashboard", "/api/stats", "/api/users/profile"},
		DefaultTTL:      5 * time.Minute,
		Compression:     false,
	}
}

// 便捷函数：创建开发环境缓存配置
func DevelopmentCacheConfig(redisClient *redis.Client) PerformanceCacheConfig {
	config := DefaultCacheConfig(redisClient)
	config.TTL = 1 * time.Minute // 开发环境短缓存时间
	config.Compression = false
	return config
}

// 便捷函数：创建生产环境缓存配置
func ProductionCacheConfig(redisClient *redis.Client) PerformanceCacheConfig {
	config := DefaultCacheConfig(redisClient)
	config.TTL = 15 * time.Minute  // 生产环境长缓存时间
	config.Compression = true
	return config
}