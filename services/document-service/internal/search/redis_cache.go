package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client    *redis.Client
	logger    *logrus.Logger
	keyPrefix string
	defaultTTL time.Duration
}

// RedisCacheConfig Redis缓存配置
type RedisCacheConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	Database int    `yaml:"database" json:"database"`
	PoolSize int    `yaml:"pool_size" json:"pool_size"`
	KeyPrefix string `yaml:"key_prefix" json:"key_prefix"`
	DefaultTTL int  `yaml:"default_ttl" json:"default_ttl"` // 秒
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(config *RedisCacheConfig, logger *logrus.Logger) (*RedisCache, error) {
	if config == nil {
		return nil, fmt.Errorf("redis cache config is required")
	}

	// 设置默认值
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 6379
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "doc_search:"
	}
	if config.DefaultTTL == 0 {
		config.DefaultTTL = 300 // 5分钟
	}

	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.Database,
		PoolSize: config.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	cache := &RedisCache{
		client:     rdb,
		logger:     logger,
		keyPrefix:  config.KeyPrefix,
		defaultTTL: time.Duration(config.DefaultTTL) * time.Second,
	}

	logger.WithFields(logrus.Fields{
		"host":     config.Host,
		"port":     config.Port,
		"database": config.Database,
	}).Info("Redis cache initialized")

	return cache, nil
}

// Get 获取缓存
func (rc *RedisCache) Get(key string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	val, err := rc.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("cache key not found")
		}
		return nil, fmt.Errorf("failed to get cache value: %w", err)
	}

	// 尝试解析JSON
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		// 如果不是JSON，直接返回字符串
		return val, nil
	}

	return result, nil
}

// Set 设置缓存
func (rc *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	// 序列化值
	var valStr string
	if str, ok := value.(string); ok {
		valStr = str
	} else {
		valBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal cache value: %w", err)
		}
		valStr = string(valBytes)
	}

	// 设置TTL
	if ttl <= 0 {
		ttl = rc.defaultTTL
	}

	if err := rc.client.Set(ctx, fullKey, valStr, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (rc *RedisCache) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	if err := rc.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to delete cache key: %w", err)
	}

	return nil
}

// DeletePattern 批量删除缓存
func (rc *RedisCache) DeletePattern(pattern string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fullPattern := rc.keyPrefix + pattern

	keys, err := rc.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := rc.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete cache keys: %w", err)
	}

	return nil
}

// Exists 检查缓存是否存在
func (rc *RedisCache) Exists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	count, err := rc.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	return count > 0, nil
}

// SetExpire 设置缓存过期时间
func (rc *RedisCache) SetExpire(key string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	if err := rc.client.Expire(ctx, fullKey, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache expiration: %w", err)
	}

	return nil
}

// GetTTL 获取缓存剩余时间
func (rc *RedisCache) GetTTL(key string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	ttl, err := rc.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get cache ttl: %w", err)
	}

	return ttl, nil
}

// Increment 递增计数器
func (rc *RedisCache) Increment(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	result, err := rc.client.Incr(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache counter: %w", err)
	}

	return result, nil
}

// IncrementBy 按指定值递增计数器
func (rc *RedisCache) IncrementBy(key string, value int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := rc.keyPrefix + key

	result, err := rc.client.IncrBy(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment cache counter by value: %w", err)
	}

	return result, nil
}

// GetStats 获取缓存统计
func (rc *RedisCache) GetStats() (*CacheStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	info, err := rc.client.Info(ctx, "memory", "keyspace", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get redis info: %w", err)
	}

	stats := &CacheStats{
		Info: info,
	}

	// 解析内存使用情况
	lines := parseRedisInfo(info)
	for _, line := range lines {
		if key, value, ok := parseRedisInfoLine(line); ok {
			switch key {
			case "used_memory":
				if mem, err := parseMemoryBytes(value); err == nil {
					stats.UsedMemory = mem
				}
			case "used_memory_human":
				stats.UsedMemoryHuman = value
			case "used_memory_peak":
				if mem, err := parseMemoryBytes(value); err == nil {
					stats.UsedMemoryPeak = mem
				}
			case "used_memory_peak_human":
				stats.UsedMemoryPeakHuman = value
			case "total_commands_processed":
				if commands, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.TotalCommands = commands
				}
			case "keyspace_hits":
				if hits, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.KeyspaceHits = hits
				}
			case "keyspace_misses":
				if misses, err := strconv.ParseInt(value, 10, 64); err == nil {
					stats.KeyspaceMisses = misses
				}
			}
		}
	}

	// 计算命中率
	if stats.KeyspaceHits > 0 || stats.KeyspaceMisses > 0 {
		stats.HitRate = float64(stats.KeyspaceHits) / float64(stats.KeyspaceHits+stats.KeyspaceMisses)
	}

	return stats, nil
}

// CacheStats 缓存统计
type CacheStats struct {
	UsedMemory         int64   `json:"used_memory"`
	UsedMemoryHuman    string  `json:"used_memory_human"`
	UsedMemoryPeak     int64   `json:"used_memory_peak"`
	UsedMemoryPeakHuman string  `json:"used_memory_peak_human"`
	TotalCommands      int64   `json:"total_commands"`
	KeyspaceHits       int64   `json:"keyspace_hits"`
	KeyspaceMisses     int64   `json:"keyspace_misses"`
	HitRate            float64 `json:"hit_rate"`
	Info               string  `json:"info"`
}

// parseRedisInfo 解析Redis INFO响应
func parseRedisInfo(info string) []string {
	lines := make([]string, 0)
	for _, line := range strings.Split(info, "\r\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines
}

// parseRedisInfoLine 解析Redis INFO行
func parseRedisInfoLine(line string) (key, value string, ok bool) {
	if !strings.Contains(line, ":") {
		return "", "", false
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// parseMemoryBytes 解析内存字节数
func parseMemoryBytes(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

// Close 关闭缓存连接
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// WarmUp 预热缓存
func (rc *RedisCache) WarmUp(keys map[string]interface{}, ttl time.Duration) error {
	for key, value := range keys {
		if err := rc.Set(key, value, ttl); err != nil {
			rc.logger.WithError(err).WithField("key", key).Error("Failed to warm up cache key")
		}
	}

	rc.logger.WithField("keys_count", len(keys)).Info("Cache warm up completed")

	return nil
}

// FlushAll 清空所有缓存
func (rc *RedisCache) FlushAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rc.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush redis cache: %w", err)
	}

	rc.logger.Info("Cache flushed")

	return nil
}