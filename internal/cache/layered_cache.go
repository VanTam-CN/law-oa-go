package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheItem 缓存项
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// LayeredCache 分层缓存
type LayeredCache struct {
	l1Cache *sync.Map     // L1: 内存缓存
	l2Cache *redis.Client // L2: Redis缓存
	ttl     time.Duration
	logger  *slog.Logger
}

func NewLayeredCache(l2Cache *redis.Client, ttl time.Duration, logger *slog.Logger) *LayeredCache {
	return &LayeredCache{
		l1Cache: &sync.Map{},
		l2Cache: l2Cache,
		ttl:     ttl,
		logger:  logger,
	}
}

// Get 从缓存获取数据
func (lc *LayeredCache) Get(ctx context.Context, key string) (interface{}, bool) {
	start := time.Now()
	defer func() {
		// TODO: 可以通过回调函数或者事件总线来记录指标，避免循环依赖
		_ = time.Since(start)
	}()

	// 先查L1缓存
	if item, ok := lc.l1Cache.Load(key); ok {
		cacheItem := item.(*CacheItem)
		if time.Now().Before(cacheItem.Expiration) {
			lc.logger.Debug("L1 cache hit", "key", key)
			return cacheItem.Value, true
		}
		// L1缓存过期，删除
		lc.l1Cache.Delete(key)
	}

	// 再查L2缓存
	val, err := lc.l2Cache.Get(ctx, key).Result()
	if err == nil {
		lc.logger.Debug("L2 cache hit", "key", key)

		// 反序列化
		var result interface{}
		if err := json.Unmarshal([]byte(val), &result); err == nil {
			// 回填L1缓存
			lc.l1Cache.Store(key, &CacheItem{
				Value:      result,
				Expiration: time.Now().Add(lc.ttl / 2), // L1缓存时间减半
			})
			return result, true
		}
	}

	lc.logger.Debug("Cache miss", "key", key)
	return nil, false
}

// Set 设置缓存
func (lc *LayeredCache) Set(ctx context.Context, key string, value interface{}) error {
	start := time.Now()
	defer func() {
		// TODO: 可以通过回调函数或者事件总线来记录指标，避免循环依赖
		_ = time.Since(start)
	}()

	// 序列化
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 设置L1缓存
	lc.l1Cache.Store(key, &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(lc.ttl / 2),
	})

	// 设置L2缓存
	err = lc.l2Cache.Set(ctx, key, data, lc.ttl).Err()
	if err != nil {
		lc.logger.Error("Failed to set L2 cache", "key", key, "error", err)
		return err
	}

	lc.logger.Debug("Cache set", "key", key, "ttl", lc.ttl)
	return nil
}

// Delete 删除缓存
func (lc *LayeredCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		// TODO: 可以通过回调函数或者事件总线来记录指标，避免循环依赖
		_ = time.Since(start)
	}()

	// 删除L1缓存
	lc.l1Cache.Delete(key)

	// 删除L2缓存
	err := lc.l2Cache.Del(ctx, key).Err()
	if err != nil {
		lc.logger.Error("Failed to delete L2 cache", "key", key, "error", err)
		return err
	}

	lc.logger.Debug("Cache deleted", "key", key)
	return nil
}

// ClearExpired 清理过期缓存
func (lc *LayeredCache) ClearExpired() {
	lc.l1Cache.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok {
			if time.Now().After(item.Expiration) {
				lc.l1Cache.Delete(key)
			}
		}
		return true
	})
}

// GetOrSet 获取或设置缓存
func (lc *LayeredCache) GetOrSet(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	if val, ok := lc.Get(ctx, key); ok {
		return val, nil
	}

	val, err := fn()
	if err != nil {
		return nil, err
	}

	if err := lc.Set(ctx, key, val); err != nil {
		lc.logger.Error("Failed to cache value", "key", key, "error", err)
	}

	return val, nil
}
