package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	Database     int    `mapstructure:"database"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	PoolTimeout  int    `mapstructure:"pool_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
	IsCluster    bool   `mapstructure:"is_cluster"`
	ClusterNodes []string `mapstructure:"cluster_nodes"`
}

// DefaultRedisConfig 返回默认Redis配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		Database:     0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
		PoolTimeout:  5,
		IdleTimeout:  300,
		IsCluster:    false,
		ClusterNodes: []string{},
	}
}

// GetAddr 获取Redis地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RedisClient Redis客户端管理器
type RedisClient struct {
	Client redis.Cmdable
	config *RedisConfig
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	var client redis.Cmdable
	var err error

	if config.IsCluster && len(config.ClusterNodes) > 0 {
		// 集群模式
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:         config.ClusterNodes,
			Password:      config.Password,
			PoolSize:      config.PoolSize,
			MinIdleConns:  config.MinIdleConns,
			MaxRetries:    config.MaxRetries,
			DialTimeout:   time.Duration(config.DialTimeout) * time.Second,
			ReadTimeout:   time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout:  time.Duration(config.WriteTimeout) * time.Second,
			PoolTimeout:   time.Duration(config.PoolTimeout) * time.Second,
			IdleTimeout:   time.Duration(config.IdleTimeout) * time.Second,
			RouteByLatency: true,
		})
	} else {
		// 单机模式
		client = redis.NewClient(&redis.Options{
			Addr:         config.GetAddr(),
			Password:     config.Password,
			DB:           config.Database,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  time.Duration(config.DialTimeout) * time.Second,
			ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
			PoolTimeout:  time.Duration(config.PoolTimeout) * time.Second,
			IdleTimeout:  time.Duration(config.IdleTimeout) * time.Second,
		})
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		Client: client,
		config: config,
	}, nil
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	switch client := r.Client.(type) {
	case *redis.Client:
		return client.Close()
	case *redis.ClusterClient:
		return client.Close()
	default:
		return fmt.Errorf("unsupported Redis client type")
	}
}

// Health Redis健康检查
func (r *RedisClient) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}
	return nil
}

// Stats 获取Redis统计信息
func (r *RedisClient) Stats() map[string]interface{} {
	switch client := r.Client.(type) {
	case *redis.Client:
		poolStats := client.PoolStats()
		return map[string]interface{}{
			"type":                "single",
			"total_conns":         poolStats.TotalConns,
			"idle_conns":          poolStats.IdleConns,
			"stale_conns":         poolStats.StaleConns,
			"hits":                poolStats.Hits,
			"misses":              poolStats.Misses,
			"timeouts":            poolStats.Timeouts,
			"total_conns_wait_ms": poolStats.TotalConnsWaitMillis,
			"total_wait_ms":       poolStats.TotalWaitMillis,
			"hits_percent":        float64(poolStats.Hits) / float64(poolStats.Hits+poolStats.Misses) * 100,
		}
	case *redis.ClusterClient:
		return map[string]interface{}{
			"type": "cluster",
			"info": "Cluster client stats not available in simple format",
		}
	default:
		return map[string]interface{}{
			"type":  "unknown",
			"error": "Unsupported client type",
		}
	}
}

// Set 设置键值对
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.Client.Set(ctx, key, jsonData, expiration).Err()
}

// Get 获取值
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// SetWithTTL 设置带TTL的键值对
func (r *RedisClient) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.Set(ctx, key, value, ttl)
}

// GetWithTTL 获取值和剩余TTL
func (r *RedisClient) GetWithTTL(ctx context.Context, key string, dest interface{}) (time.Duration, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, fmt.Errorf("key not found: %s", key)
		}
		return 0, fmt.Errorf("failed to get key: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return 0, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	ttl, err := r.Client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

// Increment 原子递增
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Decrement 原子递减
func (r *RedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	return r.Client.Decr(ctx, key).Result()
}

// SetNX 仅在键不存在时设置
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.Client.SetNX(ctx, key, jsonData, expiration).Result()
}

// Keys 获取匹配模式的所有键
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.Client.Keys(ctx, pattern).Result()
}

// FlushDB 清空当前数据库
func (r *RedisClient) FlushDB(ctx context.Context) error {
	return r.Client.FlushDB(ctx).Err()
}

// Pipeline 创建管道
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.Client.Pipeline()
}

// CacheManager 缓存管理器
type CacheManager struct {
	redis *RedisClient
	prefix string
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(redis *RedisClient, prefix string) *CacheManager {
	return &CacheManager{
		redis:  redis,
		prefix: prefix,
	}
}

// buildKey 构建缓存键
func (c *CacheManager) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// Set 设置缓存
func (c *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	cacheKey := c.buildKey(key)
	return c.redis.Set(ctx, cacheKey, value, expiration)
}

// Get 获取缓存
func (c *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	cacheKey := c.buildKey(key)
	return c.redis.Get(ctx, cacheKey, dest)
}

// Delete 删除缓存
func (c *CacheManager) Delete(ctx context.Context, keys ...string) error {
	cacheKeys := make([]string, len(keys))
	for i, key := range keys {
		cacheKeys[i] = c.buildKey(key)
	}
	return c.redis.Delete(ctx, cacheKeys...)
}

// Exists 检查缓存是否存在
func (c *CacheManager) Exists(ctx context.Context, key string) (bool, error) {
	cacheKey := c.buildKey(key)
	return c.redis.Exists(ctx, cacheKey)
}

// ClearPattern 清除匹配模式的所有缓存
func (c *CacheManager) ClearPattern(ctx context.Context, pattern string) error {
	fullPattern := c.buildKey(pattern)
	keys, err := c.redis.Keys(ctx, fullPattern)
	if err != nil {
		return fmt.Errorf("failed to get keys with pattern %s: %w", fullPattern, err)
	}

	if len(keys) > 0 {
		return c.redis.Delete(ctx, keys...)
	}

	return nil
}