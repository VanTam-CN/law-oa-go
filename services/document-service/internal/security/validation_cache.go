package security

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ValidationCache 验证结果缓存
type ValidationCache struct {
	cache    map[string]*CacheEntry
	maxSize  int
	ttl      time.Duration
	mutex    sync.RWMutex
	logger   *logrus.Logger
	stats    *CacheStats
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Result    *ValidationResult
	Timestamp time.Time
	Hits      int64
	LastUsed  time.Time
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Evictions   int64 `json:"evictions"`
	TotalSize   int   `json:"total_size"`
	HitRatio    float64 `json:"hit_ratio"`
	mutex       sync.RWMutex
}

// NewValidationCache 创建验证缓存
func NewValidationCache(ttl time.Duration, logger *logrus.Logger) *ValidationCache {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	cache := &ValidationCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: 1000, // 默认最大缓存1000条
		ttl:     ttl,
		logger:  logger,
		stats:   &CacheStats{},
	}

	// 启动清理协程
	go cache.cleanup()

	return cache
}

// Get 获取缓存条目
func (vc *ValidationCache) Get(key string) (*ValidationResult, bool) {
	vc.mutex.RLock()
	defer vc.mutex.RUnlock()

	entry, exists := vc.cache[key]
	if !exists {
		vc.stats.mutex.Lock()
		vc.stats.Misses++
		vc.stats.mutex.Unlock()
		return nil, false
	}

	// 检查是否过期
	if time.Since(entry.Timestamp) > vc.ttl {
		delete(vc.cache, key)
		vc.stats.mutex.Lock()
		vc.stats.Misses++
		vc.stats.TotalSize = len(vc.cache)
		vc.stats.mutex.Unlock()
		return nil, false
	}

	// 更新使用统计
	entry.Hits++
	entry.LastUsed = time.Now()

	vc.stats.mutex.Lock()
	vc.stats.Hits++
	vc.updateHitRatio()
	vc.stats.mutex.Unlock()

	vc.logger.WithFields(logrus.Fields{
		"cache_key": key,
		"hits":      entry.Hits,
	}).Debug("缓存命中")

	return entry.Result, true
}

// Set 设置缓存条目
func (vc *ValidationCache) Set(key string, result *ValidationResult) {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	// 如果缓存已满，执行LRU淘汰
	if len(vc.cache) >= vc.maxSize {
		vc.evictLRU()
	}

	// 添加新条目
	vc.cache[key] = &CacheEntry{
		Result:    result,
		Timestamp: time.Now(),
		Hits:      0,
		LastUsed:  time.Now(),
	}

	vc.stats.mutex.Lock()
	vc.stats.TotalSize = len(vc.cache)
	vc.stats.mutex.Unlock()

	vc.logger.WithField("cache_key", key).Debug("缓存已更新")
}

// Delete 删除缓存条目
func (vc *ValidationCache) Delete(key string) {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	if _, exists := vc.cache[key]; exists {
		delete(vc.cache, key)
		vc.stats.mutex.Lock()
		vc.stats.TotalSize = len(vc.cache)
		vc.stats.mutex.Unlock()

		vc.logger.WithField("cache_key", key).Debug("缓存条目已删除")
	}
}

// Clear 清空缓存
func (vc *ValidationCache) Clear() {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	vc.cache = make(map[string]*CacheEntry)
	vc.stats.mutex.Lock()
	vc.stats.TotalSize = 0
	vc.stats.mutex.Unlock()

	vc.logger.Info("验证缓存已清空")
}

// SetMaxSize 设置最大缓存大小
func (vc *ValidationCache) SetMaxSize(size int) {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	vc.maxSize = size

	// 如果当前缓存大小超过新的限制，执行淘汰
	for len(vc.cache) > size {
		vc.evictLRU()
	}

	vc.logger.WithField("max_size", size).Info("缓存最大大小已更新")
}

// SetTTL 设置缓存TTL
func (vc *ValidationCache) SetTTL(ttl time.Duration) {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()

	vc.ttl = ttl
	vc.logger.WithField("ttl", ttl).Info("缓存TTL已更新")
}

// GetStats 获取缓存统计信息
func (vc *ValidationCache) GetStats() *CacheStats {
	vc.stats.mutex.RLock()
	defer vc.stats.mutex.RUnlock()

	// 返回统计信息的副本
	return &CacheStats{
		Hits:      vc.stats.Hits,
		Misses:    vc.stats.Misses,
		Evictions: vc.stats.Evictions,
		TotalSize: vc.stats.TotalSize,
		HitRatio:  vc.stats.HitRatio,
	}
}

// evictLRU 执行LRU淘汰
func (vc *ValidationCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range vc.cache {
		if oldestKey == "" || entry.LastUsed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastUsed
		}
	}

	if oldestKey != "" {
		delete(vc.cache, oldestKey)
		vc.stats.mutex.Lock()
		vc.stats.Evictions++
		vc.stats.TotalSize = len(vc.cache)
		vc.stats.mutex.Unlock()

		vc.logger.WithField("evicted_key", oldestKey).Debug("LRU淘汰缓存条目")
	}
}

// cleanup 清理过期缓存
func (vc *ValidationCache) cleanup() {
	ticker := time.NewTicker(vc.ttl / 4) // 每个TTL的1/4时间清理一次
	defer ticker.Stop()

	for range ticker.C {
		vc.mutex.Lock()
		now := time.Now()
		for key, entry := range vc.cache {
			if now.Sub(entry.Timestamp) > vc.ttl {
				delete(vc.cache, key)
				vc.stats.mutex.Lock()
				vc.stats.TotalSize = len(vc.cache)
				vc.stats.mutex.Unlock()
				vc.logger.WithField("expired_key", key).Debug("清理过期缓存条目")
			}
		}
		vc.mutex.Unlock()
	}
}

// updateHitRatio 更新命中率
func (vc *ValidationCache) updateHitRatio() {
	total := vc.stats.Hits + vc.stats.Misses
	if total > 0 {
		vc.stats.HitRatio = float64(vc.stats.Hits) / float64(total)
	}
}

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(cert *x509.Certificate, dnsName string) string {
	data := cert.SerialNumber.String() + "|" + dnsName + "|" + cert.Subject.String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CacheEntryInfo 缓存条目信息
type CacheEntryInfo struct {
	Key       string    `json:"key"`
	Serial    string    `json:"serial"`
	Subject   string    `json:"subject"`
	DNSName   string    `json:"dns_name"`
	Hits      int64     `json:"hits"`
	Timestamp time.Time `json:"timestamp"`
	LastUsed  time.Time `json:"last_used"`
	IsValid   bool      `json:"is_valid"`
}

// GetCacheEntries 获取所有缓存条目信息
func (vc *ValidationCache) GetCacheEntries() []CacheEntryInfo {
	vc.mutex.RLock()
	defer vc.mutex.RUnlock()

	entries := make([]CacheEntryInfo, 0, len(vc.cache))
	now := time.Now()

	for key, entry := range vc.cache {
		isValid := now.Sub(entry.Timestamp) <= vc.ttl
		info := CacheEntryInfo{
			Key:       key,
			Serial:    entry.Result.SerialNumber,
			Subject:   entry.Result.Subject,
			DNSName:   entry.Result.DNSName,
			Hits:      entry.Hits,
			Timestamp: entry.Timestamp,
			LastUsed:  entry.LastUsed,
			IsValid:   isValid,
		}
		entries = append(entries, info)
	}

	return entries
}

// GetHotEntries 获取热点缓存条目（按命中次数排序）
func (vc *ValidationCache) GetHotEntries(limit int) []CacheEntryInfo {
	entries := vc.GetCacheEntries()

	// 按命中次数排序
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].Hits < entries[j].Hits {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// 返回前limit个
	if limit > 0 && limit < len(entries) {
		return entries[:limit]
	}
	return entries
}

// GetExpiredEntries 获取过期缓存条目
func (vc *ValidationCache) GetExpiredEntries() []CacheEntryInfo {
	entries := vc.GetCacheEntries()
	now := time.Now()

	var expired []CacheEntryInfo
	for _, entry := range entries {
		if !entry.IsValid && now.Sub(entry.Timestamp) > vc.ttl {
			expired = append(expired, entry)
		}
	}

	return expired
}