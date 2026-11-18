package database

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/config"
)

// CacheManager 智能缓存管理器 - 基于最新缓存策略
type CacheManager struct {
	db     *gorm.DB
	cache  map[string]*CacheEntry
	config *config.DatabasePerformanceConfig
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data      interface{}
	CreatedAt time.Time
	ExpiresAt time.Time
	HitCount  int64
	Size      int64
}

// CacheStrategy 缓存策略
type CacheStrategy struct {
	TTL        time.Duration
	KeyPrefix  string
	Enabled    bool
	MaxSize    int64
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(db *gorm.DB, config *config.DatabasePerformanceConfig) *CacheManager {
	cm := &CacheManager{
		db:     db,
		cache:  make(map[string]*CacheEntry),
		config: config,
	}

	// 启动清理协程
	go cm.cleanupExpiredEntries()

	return cm
}

// GetCacheKey 生成缓存键
func (cm *CacheManager) GetCacheKey(operation string, params ...interface{}) string {
	hash := sha256.New()
	hash.Write([]byte(operation))

	for _, param := range params {
		hash.Write([]byte(fmt.Sprintf("%v", param)))
	}

	return hex.EncodeToString(hash.Sum(nil))[:16]
}

// Set 设置缓存
func (cm *CacheManager) Set(key string, data interface{}, ttl time.Duration) error {
	if !cm.config.EnablePerformance {
		return nil
	}

	entry := &CacheEntry{
		Data:      data,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		HitCount:  0,
		Size:      cm.estimateSize(data),
	}

	// 检查缓存大小限制
	if cm.config.MaxOpenConns > 0 && len(cm.cache) >= cm.config.MaxOpenConns {
		cm.evictLRU()
	}

	cm.cache[key] = entry
	return nil
}

// Get 获取缓存
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	if !cm.config.EnablePerformance {
		return nil, false
	}

	entry, exists := cm.cache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.ExpiresAt) {
		delete(cm.cache, key)
		return nil, false
	}

	// 增加命中计数
	entry.HitCount++
	return entry.Data, true
}

// Delete 删除缓存
func (cm *CacheManager) Delete(key string) {
	delete(cm.cache, key)
}

// Clear 清空缓存
func (cm *CacheManager) Clear() {
	cm.cache = make(map[string]*CacheEntry)
}

// CachedQuery 缓存查询 - 智能缓存策略
func (cm *CacheManager) CachedQuery(query func(*gorm.DB) *gorm.DB, cacheKey string, ttl time.Duration) *gorm.DB {
	if !cm.config.EnablePerformance {
		return query(cm.db)
	}

	// 尝试从缓存获取
	if cachedData, exists := cm.Get(cacheKey); exists {
		// 如果缓存了查询结果，需要特殊处理
		if result, ok := cachedData.(*gorm.DB); ok {
			return result
		}
	}

	// 执行查询
	result := query(cm.db)

	// 对于SELECT查询，可以缓存结果
	if result.Statement.SQL.String() != "" {
		// 注意：这里简化处理，实际应用中需要更复杂的缓存逻辑
		cm.Set(cacheKey, result, ttl)
	}

	return result
}

// InvalidatePattern 按模式使缓存失效
func (cm *CacheManager) InvalidatePattern(pattern string) {
	for key := range cm.cache {
		if strings.Contains(key, pattern) {
			delete(cm.cache, key)
		}
	}
}

// WarmupCache 缓存预热
func (cm *CacheManager) WarmupCache(queries []WarmupQuery) error {
	for _, query := range queries {
		if query.Enabled {
			// 执行预热查询
			result := query.QueryFunc(cm.db)
			if result.Error == nil {
				// 缓存结果
				cacheKey := cm.GetCacheKey(query.KeyPrefix, query.Params...)
				cm.Set(cacheKey, result, query.TTL)
			}
		}
	}
	return nil
}

// WarmupQuery 预热查询
type WarmupQuery struct {
	KeyPrefix string
	Params    []interface{}
	QueryFunc func(*gorm.DB) *gorm.DB
	TTL       time.Duration
	Enabled   bool
}

// estimateSize 估算数据大小
func (cm *CacheManager) estimateSize(data interface{}) int64 {
	if data == nil {
		return 0
	}

	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return int64(v.Len()) * 64 // 估算每个元素64字节
	case reflect.Map:
		return int64(v.Len()) * 128 // 估算每个键值对128字节
	case reflect.Struct:
		return 128 // 估算结构体128字节
	default:
		return 64 // 默认64字节
	}
}

// evictLRU 使用LRU策略驱逐缓存
func (cm *CacheManager) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range cm.cache {
		if first || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(cm.cache, oldestKey)
	}
}

// cleanupExpiredEntries 清理过期条目
func (cm *CacheManager) cleanupExpiredEntries() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for key, entry := range cm.cache {
			if now.After(entry.ExpiresAt) {
				delete(cm.cache, key)
			}
		}
	}
}

// GetStats 获取缓存统计信息
func (cm *CacheManager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_entries"] = len(cm.cache)
	stats["enabled"] = cm.config.EnablePerformance

	var totalHits int64
	var totalSize int64
	var expiredCount int
	now := time.Now()

	for _, entry := range cm.cache {
		totalHits += entry.HitCount
		totalSize += entry.Size
		if now.After(entry.ExpiresAt) {
			expiredCount++
		}
	}

	stats["total_hits"] = totalHits
	stats["total_size"] = totalSize
	stats["expired_count"] = expiredCount

	return stats
}

// SmartCacheStrategy 智能缓存策略
type SmartCacheStrategy struct {
	ReadHeavy  CacheStrategy // 读密集型数据
	WriteHeavy CacheStrategy // 写密集型数据
	StaticData CacheStrategy // 静态数据
}

// GetStrategy 根据数据类型获取缓存策略
func (scm *SmartCacheManager) GetStrategy(dataType string) CacheStrategy {
	switch dataType {
	case "user_profile", "settings", "configuration":
		return scm.strategies.StaticData
	case "audit_logs", "user_sessions":
		return scm.strategies.WriteHeavy
	default:
		return scm.strategies.ReadHeavy
	}
}

// SmartCacheManager 智能缓存管理器
type SmartCacheManager struct {
	*CacheManager
	strategies SmartCacheStrategy
}

// NewSmartCacheManager 创建智能缓存管理器
func NewSmartCacheManager(db *gorm.DB, config *config.DatabasePerformanceConfig) *SmartCacheManager {
	scm := &SmartCacheManager{
		CacheManager: NewCacheManager(db, config),
		strategies: SmartCacheStrategy{
			ReadHeavy: CacheStrategy{
				TTL:       5 * time.Minute,
				KeyPrefix: "read:",
				Enabled:   true,
				MaxSize:   10000,
			},
			WriteHeavy: CacheStrategy{
				TTL:       1 * time.Minute,
				KeyPrefix: "write:",
				Enabled:   true,
				MaxSize:   1000,
			},
			StaticData: CacheStrategy{
				TTL:       30 * time.Minute,
				KeyPrefix: "static:",
				Enabled:   true,
				MaxSize:   5000,
			},
		},
	}

	return scm
}

// SmartGet 智能获取缓存
func (scm *SmartCacheManager) SmartGet(dataType string, key string) (interface{}, bool) {
	strategy := scm.GetStrategy(dataType)
	if !strategy.Enabled {
		return nil, false
	}

	fullKey := strategy.KeyPrefix + key
	return scm.Get(fullKey)
}

// SmartSet 智能设置缓存
func (scm *SmartCacheManager) SmartSet(dataType string, key string, data interface{}) error {
	strategy := scm.GetStrategy(dataType)
	if !strategy.Enabled {
		return nil
	}

	fullKey := strategy.KeyPrefix + key
	return scm.Set(fullKey, data, strategy.TTL)
}

// CacheTag 缓存标签管理
type CacheTag struct {
	Name    string
	Keys    []string
	Version int64
}

// TaggedCache 带标签的缓存管理
type TaggedCache struct {
	*CacheManager
	tags map[string]*CacheTag
}

// NewTaggedCache 创建带标签的缓存管理器
func NewTaggedCache(db *gorm.DB, config *config.DatabasePerformanceConfig) *TaggedCache {
	return &TaggedCache{
		CacheManager: NewCacheManager(db, config),
		tags:         make(map[string]*CacheTag),
	}
}

// AddTag 添加缓存标签
func (tc *TaggedCache) AddTag(tagName string, keys ...string) {
	tag, exists := tc.tags[tagName]
	if !exists {
		tag = &CacheTag{
			Name:    tagName,
			Keys:    make([]string, 0),
			Version: 1,
		}
		tc.tags[tagName] = tag
	}

	tag.Keys = append(tag.Keys, keys...)
}

// InvalidateTag 使标签缓存失效
func (tc *TaggedCache) InvalidateTag(tagName string) {
	tag, exists := tc.tags[tagName]
	if !exists {
		return
	}

	// 删除所有关联的缓存键
	for _, key := range tag.Keys {
		tc.Delete(key)
	}

	// 增加版本号
	tag.Version++
	tag.Keys = make([]string, 0) // 清空键列表
}