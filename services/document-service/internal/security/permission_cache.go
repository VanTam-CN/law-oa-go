package security

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// NewPermissionCache 创建权限缓存
func NewPermissionCache(ttl time.Duration) *PermissionCache {
	cache := &PermissionCache{
		data:    make(map[string]*CacheEntry),
		ttl:     ttl,
		enabled: true,
	}

	// 启动清理协程
	go cache.cleanup()

	return cache
}

// Get 获取缓存值
func (pc *PermissionCache) Get(key string) *AccessDecision {
	if !pc.enabled {
		return nil
	}

	pc.mu.RLock()
	defer pc.mu.RUnlock()

	entry, exists := pc.data[key]
	if !exists {
		return nil
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry.Decision
}

// Set 设置缓存值
func (pc *PermissionCache) Set(key string, decision *AccessDecision) {
	if !pc.enabled {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	entry := &CacheEntry{
		Value:     decision.Allowed,
		ExpiresAt: time.Now().Add(pc.ttl),
		Decision:  decision,
	}

	pc.data[key] = entry
}

// Invalidate 使缓存失效
func (pc *PermissionCache) Invalidate(key string) {
	if !pc.enabled {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	delete(pc.data, key)
}

// InvalidateUser 使用户相关缓存失效
func (pc *PermissionCache) InvalidateUser(userID string) {
	if !pc.enabled {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	prefix := fmt.Sprintf("perm:%s:", userID)
	for key := range pc.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(pc.data, key)
		}
	}
}

// InvalidateResource 使资源相关缓存失效
func (pc *PermissionCache) InvalidateResource(resourceID string) {
	if !pc.enabled {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	for key := range pc.data {
		if len(key) > 20 { // perm:userid:resourceid:action:hash 格式
			parts := parseCacheKey(key)
			if len(parts) >= 3 && parts[2] == resourceID {
				delete(pc.data, key)
			}
		}
	}
}

// Clear 清空所有缓存
func (pc *PermissionCache) Clear() {
	if !pc.enabled {
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.data = make(map[string]*CacheEntry)
}

// GetStats 获取缓存统计信息
func (pc *PermissionCache) GetStats() map[string]interface{} {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	total := len(pc.data)
	expired := 0
	now := time.Now()

	for _, entry := range pc.data {
		if now.After(entry.ExpiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_entries": total,
		"expired_entries": expired,
		"valid_entries": total - expired,
		"enabled": pc.enabled,
		"ttl_seconds": pc.ttl.Seconds(),
	}
}

// SetEnabled 启用或禁用缓存
func (pc *PermissionCache) SetEnabled(enabled bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.enabled = enabled
	if !enabled {
		pc.data = make(map[string]*CacheEntry)
	}
}

// cleanup 定期清理过期缓存
func (pc *PermissionCache) cleanup() {
	ticker := time.NewTicker(pc.ttl / 4) // 每个TTL的1/4时间清理一次
	defer ticker.Stop()

	for range ticker.C {
		if !pc.enabled {
			continue
		}

		pc.mu.Lock()
		now := time.Now()
		for key, entry := range pc.data {
			if now.After(entry.ExpiresAt) {
				delete(pc.data, key)
			}
		}
		pc.mu.Unlock()
	}
}

// parseCacheKey 解析缓存键
func parseCacheKey(key string) []string {
	// 简化实现，实际应该根据具体格式解析
	parts := make([]string, 0, 5)
	start := 0
	for i, char := range key {
		if char == ':' {
			if i > start {
				parts = append(parts, key[start:i])
			}
			start = i + 1
		}
	}
	if start < len(key) {
		parts = append(parts, key[start:])
	}
	return parts
}

// HashKey 生成缓存键的哈希
func HashKey(key string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}