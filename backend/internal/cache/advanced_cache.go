package cache

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// AdvancedCacheService 高级缓存服务，包含性能优化和防护策略
type AdvancedCacheService struct {
	client      *redis.Client
	prefix      string
	defaultTTL  time.Duration
	bloomFilter *BloomFilter // 布隆过滤器，防止缓存穿透
	mutexPool   sync.Pool    // 对象池，减少内存分配
}

// BloomFilter 简单的布隆过滤器实现
type BloomFilter struct {
	bits     []bool
	size     uint
	hashFunc uint
}

// NewBloomFilter 创建布隆过滤器
func NewBloomFilter(size uint, hashFunc uint) *BloomFilter {
	return &BloomFilter{
		bits:     make([]bool, size),
		size:     size,
		hashFunc: hashFunc,
	}
}

// Add 添加元素到布隆过滤器
func (bf *BloomFilter) Add(data []byte) {
	for i := uint(0); i < bf.hashFunc; i++ {
		hash := bf.hash(data, i)
		index := hash % uint64(bf.size)
		bf.bits[index] = true
	}
}

// Test 检查元素是否可能存在
func (bf *BloomFilter) Test(data []byte) bool {
	for i := uint(0); i < bf.hashFunc; i++ {
		hash := bf.hash(data, i)
		index := hash % uint64(bf.size)
		if !bf.bits[index] {
			return false
		}
	}
	return true
}

// hash 简单的哈希函数
func (bf *BloomFilter) hash(data []byte, seed uint) uint64 {
	hash := uint64(seed)
	for _, b := range data {
		hash = hash*31 + uint64(b)
	}
	return hash
}

// CacheStats 缓存统计信息
type CacheStats struct {
	HitCount   int64   `json:"hit_count"`
	MissCount  int64   `json:"miss_count"`
	HitRate    float64 `json:"hit_rate"`
	ErrorCount int64   `json:"error_count"`
	AvgLatency float64 `json:"avg_latency_ms"`
}

// CacheResult 缓存操作结果
type CacheResult struct {
	Data      interface{}   `json:"data"`
	Hit       bool          `json:"hit"`
	TTL       time.Duration `json:"ttl"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewAdvancedCacheService 创建高级缓存服务
func NewAdvancedCacheService(client *redis.Client, prefix string, defaultTTL time.Duration) *AdvancedCacheService {
	return &AdvancedCacheService{
		client:      client,
		prefix:      prefix,
		defaultTTL:  defaultTTL,
		bloomFilter: NewBloomFilter(1000000, 3), // 100万容量，3个哈希函数
		mutexPool: sync.Pool{
			New: func() interface{} {
				return &sync.Mutex{}
			},
		},
	}
}

// GetWithCacheThroughput 通过缓存获取数据，防止缓存穿透
func (c *AdvancedCacheService) GetWithCacheThroughput(ctx context.Context, key string, dest interface{}) (*CacheResult, error) {
	start := time.Now()
	fullKey := c.buildKey(key)

	// 1. 先检查布隆过滤器
	cacheKeyData := []byte(fullKey)
	if !c.bloomFilter.Test(cacheKeyData) {
		// 布隆过滤器确定不存在，直接返回
		return &CacheResult{
			Hit:       false,
			Timestamp: start,
		}, fmt.Errorf("cache key not found (bloom filter)")
	}

	// 2. 尝试从缓存获取
	val, err := c.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		// 缓存未命中，添加到布隆过滤器
		c.bloomFilter.Add(cacheKeyData)
		return &CacheResult{
			Hit:       false,
			Timestamp: start,
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("cache get error: %w", err)
	}

	// 3. 反序列化数据
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return nil, fmt.Errorf("cache unmarshal error: %w", err)
	}

	return &CacheResult{
		Data:      dest,
		Hit:       true,
		Timestamp: start,
	}, nil
}

// SetWithRandomTTL 设置缓存，使用随机TTL防止缓存雪崩
func (c *AdvancedCacheService) SetWithRandomTTL(ctx context.Context, key string, value interface{}, baseTTL time.Duration) error {
	fullKey := c.buildKey(key)

	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	// 计算随机TTL，防止缓存雪崩
	finalTTL := c.calculateRandomTTL(baseTTL)

	// 设置缓存
	if err := c.client.Set(ctx, fullKey, data, finalTTL).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	// 添加到布隆过滤器
	c.bloomFilter.Add([]byte(fullKey))

	return nil
}

// GetWithLock 使用分布式锁防止缓存击穿
func (c *AdvancedCacheService) GetWithLock(ctx context.Context, key string, dest interface{}, lockTimeout time.Duration, fetchFunc func() (interface{}, error)) error {
	fullKey := c.buildKey(key)
	lockKey := fullKey + ":lock"

	// 1. 尝试从缓存获取
	result, err := c.GetWithCacheThroughput(ctx, key, dest)
	if err != nil && err.Error() != "cache key not found (bloom filter)" {
		return err
	}

	if result.Hit {
		return nil // 缓存命中
	}

	// 2. 尝试获取分布式锁
	locked, err := c.client.SetNX(ctx, lockKey, 1, lockTimeout).Result()
	if err != nil {
		return fmt.Errorf("cache lock error: %w", err)
	}

	if !locked {
		// 获取锁失败，等待并重试
		return c.waitForCacheAndRetry(ctx, key, dest, 3, 100*time.Millisecond)
	}

	defer c.client.Del(ctx, lockKey)

	// 3. 获得锁后，再次检查缓存
	result, err = c.GetWithCacheThroughput(ctx, key, dest)
	if err != nil && err.Error() != "cache key not found (bloom filter)" {
		return err
	}

	if result.Hit {
		return nil // 缓存命中
	}

	// 4. 执行数据获取函数
	data, err := fetchFunc()
	if err != nil {
		return fmt.Errorf("fetch function error: %w", err)
	}

	// 5. 更新缓存
	if err := c.SetWithRandomTTL(ctx, key, data, c.defaultTTL); err != nil {
		// 缓存设置失败，但不影响主流程
		fmt.Printf("Warning: failed to set cache for key %s: %v\n", key, err)
	}

	// 6. 更新目标对象
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal fetched data error: %w", err)
	}
	if err := json.Unmarshal(dataBytes, dest); err != nil {
		return fmt.Errorf("unmarshal fetched data error: %w", err)
	}

	return nil
}

// BatchSet 批量设置缓存，使用Pipeline提高性能
func (c *AdvancedCacheService) BatchSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()

	for key, value := range items {
		fullKey := c.buildKey(key)
		data, err := json.Marshal(value)
		if err != nil {
			continue // 跳过序列化失败的项目
		}

		finalTTL := c.calculateRandomTTL(ttl)
		pipe.Set(ctx, fullKey, data, finalTTL)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("batch cache set error: %w", err)
	}

	// 批量添加到布隆过滤器
	for key := range items {
		c.bloomFilter.Add([]byte(c.buildKey(key)))
	}

	return nil
}

// BatchGet 批量获取缓存
func (c *AdvancedCacheService) BatchGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// 构建完整键名
	fullKeys := make([]string, len(keys))
	keyMap := make(map[string]string)
	for i, key := range keys {
		fullKey := c.buildKey(key)
		fullKeys[i] = fullKey
		keyMap[fullKey] = key
	}

	// 批量获取
	values, err := c.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("batch cache get error: %w", err)
	}

	result := make(map[string]interface{})
	for i, value := range values {
		if value != nil {
			var data interface{}
			if err := json.Unmarshal([]byte(value.(string)), &data); err == nil {
				originalKey := keyMap[fullKeys[i]]
				result[originalKey] = data
			}
		}
	}

	return result, nil
}

// DeletePattern 删除匹配模式的缓存
func (c *AdvancedCacheService) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.buildKey(pattern)
	keys, err := c.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return fmt.Errorf("cache pattern keys error: %w", err)
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func (c *AdvancedCacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	// 这里可以添加更详细的统计信息获取逻辑
	stats := &CacheStats{
		HitCount:   0, // 需要通过Redis的stats命令获取
		MissCount:  0,
		HitRate:    0.0,
		ErrorCount: 0,
		AvgLatency: 0.0,
	}

	return stats, nil
}

// WarmupCache 缓存预热
func (c *AdvancedCacheService) WarmupCache(ctx context.Context, dataProvider func() (map[string]interface{}, error)) error {
	// 从数据源获取需要预热的数据
	data, err := dataProvider()
	if err != nil {
		return fmt.Errorf("cache warmup data provider error: %w", err)
	}

	// 批量设置缓存
	return c.BatchSet(ctx, data, c.defaultTTL)
}

// calculateRandomTTL 计算随机TTL防止缓存雪崩
func (c *AdvancedCacheService) calculateRandomTTL(baseTTL time.Duration) time.Duration {
	// 添加10%的随机偏移
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return baseTTL
	}

	randomOffset := time.Duration(randomBytes[0]) * (baseTTL / 10)
	return baseTTL + randomOffset
}

// waitForCacheAndRetry 等待缓存并重试
func (c *AdvancedCacheService) waitForCacheAndRetry(ctx context.Context, key string, dest interface{}, maxRetries int, interval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		time.Sleep(interval)

		result, err := c.GetWithCacheThroughput(ctx, key, dest)
		if err != nil {
			continue
		}

		if result.Hit {
			return nil
		}
	}

	return fmt.Errorf("cache retry timeout for key: %s", key)
}

// buildKey 构建完整的缓存键
func (c *AdvancedCacheService) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// GetMutexFromPool 从对象池获取互斥锁
func (c *AdvancedCacheService) GetMutexFromPool() *sync.Mutex {
	return c.mutexPool.Get().(*sync.Mutex)
}

// PutMutexToPool 将互斥锁放回对象池
func (c *AdvancedCacheService) PutMutexToPool(mutex *sync.Mutex) {
	c.mutexPool.Put(mutex)
}

// 便捷方法，保持与原有CacheService的兼容性
func (c *AdvancedCacheService) Get(key string, dest interface{}) error {
	ctx := context.Background()
	_, err := c.GetWithCacheThroughput(ctx, key, dest)
	return err
}

func (c *AdvancedCacheService) Set(key string, value interface{}, ttl time.Duration) error {
	ctx := context.Background()
	return c.SetWithRandomTTL(ctx, key, value, ttl)
}

func (c *AdvancedCacheService) Delete(key string) error {
	ctx := context.Background()
	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

func (c *AdvancedCacheService) GetClient() *redis.Client {
	return c.client
}

func (c *AdvancedCacheService) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	return err == nil
}
