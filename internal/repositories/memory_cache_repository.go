package repositories

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryCacheRepository 内存缓存仓库实现
type MemoryCacheRepository struct {
	cache map[string]cacheEntry
	mu    sync.RWMutex
}

type cacheEntry struct {
	value      string
	expiration time.Time
}

// NewMemoryCacheRepository 创建内存缓存仓库
func NewMemoryCacheRepository() CacheRepository {
	return &MemoryCacheRepository{
		cache: make(map[string]cacheEntry),
	}
}

// Get 获取缓存值
func (r *MemoryCacheRepository) Get(ctx context.Context, key string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.cache[key]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	if time.Now().After(entry.expiration) {
		delete(r.cache, key)
		return "", fmt.Errorf("key expired")
	}

	return entry.value, nil
}

// Set 设置缓存值
func (r *MemoryCacheRepository) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(expiration),
	}

	return nil
}

// Delete 删除缓存值
func (r *MemoryCacheRepository) Delete(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.cache, key)
	return nil
}

// DeletePattern 删除匹配模式的缓存值
func (r *MemoryCacheRepository) DeletePattern(ctx context.Context, pattern string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 简单的模式匹配，支持通配符
	for key := range r.cache {
		if strings.Contains(key, strings.TrimSuffix(pattern, "*")) {
			delete(r.cache, key)
		}
	}

	return nil
}

// CleanupExpired 清理过期的缓存项（可选的后台任务）
func (r *MemoryCacheRepository) CleanupExpired(ctx context.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for key, entry := range r.cache {
		if now.After(entry.expiration) {
			delete(r.cache, key)
		}
	}
}