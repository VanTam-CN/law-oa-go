package privacy

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"
)

// MaskingCache 脱敏缓存
type MaskingCache struct {
	data    map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
	logger  *slog.Logger
	stats   *CacheStats
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Value      string
	Expiration time.Time
	HitCount   int64
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits        int64 `json:"hits"`
	Misses      int64 `json:"misses"`
	Evictions   int64 `json:"evictions"`
	TotalCount  int64 `json:"total_count"`
	MaxSize     int   `json:"max_size"`
	CurrentSize int   `json:"current_size"`
	HitRate     float64 `json:"hit_rate"`
	mu          sync.RWMutex
}

// NewMaskingCache 创建脱敏缓存
func NewMaskingCache(ttl time.Duration) *MaskingCache {
	cache := &MaskingCache{
		data:   make(map[string]*CacheEntry),
		ttl:    ttl,
		logger: slog.New(slog.NewJSONHandler(nil, &slog.HandlerOptions{Level: slog.LevelInfo})),
		stats: &CacheStats{
			MaxSize: 10000, // 默认最大缓存条目数
		},
	}

	// 启动清理协程
	go cache.startCleanup()

	return cache
}

// Get 获取缓存值
func (mc *MaskingCache) Get(key, strategy, dataType string) (string, bool) {
	cacheKey := mc.generateKey(key, strategy, dataType)

	mc.mu.RLock()
	entry, exists := mc.data[cacheKey]
	mc.mu.RUnlock()

	if !exists {
		mc.stats.incrementMisses()
		return "", false
	}

	// 检查是否过期
	if time.Now().After(entry.Expiration) {
		mc.mu.Lock()
		delete(mc.data, cacheKey)
		mc.mu.Unlock()

		mc.stats.incrementMisses()
		mc.stats.incrementEvictions()
		return "", false
	}

	// 更新命中计数
	entry.HitCount++
	mc.stats.incrementHits()

	return entry.Value, true
}

// Set 设置缓存值
func (mc *MaskingCache) Set(key, value, strategy, dataType string) {
	cacheKey := mc.generateKey(key, strategy, dataType)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// 检查缓存大小限制
	if len(mc.data) >= mc.stats.MaxSize {
		mc.evictOldest()
	}

	mc.data[cacheKey] = &CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(mc.ttl),
		HitCount:   0,
	}

	mc.stats.incrementTotal()
}

// Delete 删除缓存条目
func (mc *MaskingCache) Delete(key, strategy, dataType string) {
	cacheKey := mc.generateKey(key, strategy, dataType)

	mc.mu.Lock()
	delete(mc.data, cacheKey)
	mc.mu.Unlock()
}

// Clear 清空缓存
func (mc *MaskingCache) Clear() {
	mc.mu.Lock()
	mc.data = make(map[string]*CacheEntry)
	mc.mu.Unlock()

	mc.stats.mu.Lock()
	mc.stats.CurrentSize = 0
	mc.stats.mu.Unlock()
}

// generateKey 生成缓存键
func (mc *MaskingCache) generateKey(key, strategy, dataType string) string {
	data := key + "|" + strategy + "|" + dataType
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// evictOldest 淘汰最旧的条目
func (mc *MaskingCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range mc.data {
		if oldestKey == "" || entry.Expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Expiration
		}
	}

	if oldestKey != "" {
		delete(mc.data, oldestKey)
		mc.stats.incrementEvictions()
	}
}

// startCleanup 启动清理协程
func (mc *MaskingCache) startCleanup() {
	ticker := time.NewTicker(mc.ttl / 2) // 每半个TTL清理一次
	defer ticker.Stop()

	for range ticker.C {
		mc.cleanup()
	}
}

// cleanup 清理过期条目
func (mc *MaskingCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	evictedCount := 0

	for key, entry := range mc.data {
		if now.After(entry.Expiration) {
			delete(mc.data, key)
			evictedCount++
		}
	}

	if evictedCount > 0 {
		mc.stats.mu.Lock()
		mc.stats.Evictions += int64(evictedCount)
		mc.stats.CurrentSize = len(mc.data)
		mc.stats.mu.Unlock()

		mc.logger.Info("清理过期缓存条目", "count", evictedCount)
	}
}

// GetStatistics 获取缓存统计信息
func (mc *MaskingCache) GetStatistics() map[string]interface{} {
	mc.stats.mu.Lock()
	defer mc.stats.mu.Unlock()

	mc.stats.CurrentSize = len(mc.data)

	// 计算命中率
	total := mc.stats.Hits + mc.stats.Misses
	if total > 0 {
		mc.stats.HitRate = float64(mc.stats.Hits) / float64(total)
	}

	return map[string]interface{}{
		"hits":         mc.stats.Hits,
		"misses":       mc.stats.Misses,
		"evictions":    mc.stats.Evictions,
		"total_count":  mc.stats.TotalCount,
		"max_size":     mc.stats.MaxSize,
		"current_size": mc.stats.CurrentSize,
		"hit_rate":     mc.stats.HitRate,
	}
}

// incrementHits 增加命中计数
func (cs *CacheStats) incrementHits() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Hits++
}

// incrementMisses 增加未命中计数
func (cs *CacheStats) incrementMisses() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Misses++
}

// incrementTotal 增加总数计数
func (cs *CacheStats) incrementTotal() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.TotalCount++
}

// incrementEvictions 增加淘汰计数
func (cs *CacheStats) incrementEvictions() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Evictions++
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	ProcessingTime map[string]time.Duration `json:"processing_time"`
	Throughput     map[string]int64         `json:"throughput"`
	ErrorRate       map[string]float64       `json:"error_rate"`
	mu              sync.RWMutex
}

// NewPerformanceMetrics 创建性能指标
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		ProcessingTime: make(map[string]time.Duration),
		Throughput:     make(map[string]int64),
		ErrorRate:      make(map[string]float64),
	}
}

// RecordProcessingTime 记录处理时间
func (pm *PerformanceMetrics) RecordProcessingTime(dataType string, duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.ProcessingTime[dataType]; !exists {
		pm.ProcessingTime[dataType] = 0
	}

	// 使用滑动平均
	current := pm.ProcessingTime[dataType]
	pm.ProcessingTime[dataType] = (current + duration) / 2
}

// RecordThroughput 记录吞吐量
func (pm *PerformanceMetrics) RecordThroughput(dataType string, count int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.Throughput[dataType]; !exists {
		pm.Throughput[dataType] = 0
	}

	pm.Throughput[dataType] += count
}

// RecordErrorRate 记录错误率
func (pm *PerformanceMetrics) RecordErrorRate(dataType string, rate float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.ErrorRate[dataType] = rate
}

// GetMetrics 获取性能指标
func (pm *PerformanceMetrics) GetMetrics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return map[string]interface{}{
		"processing_time": pm.ProcessingTime,
		"throughput":      pm.Throughput,
		"error_rate":      pm.ErrorRate,
	}
}

// ConcurrentProcessor 并发处理器
type ConcurrentProcessor struct {
	workers    int
	jobQueue    chan Job
	resultQueue chan Result
	workerPool  chan struct{}
	wg          sync.WaitGroup
}

// Job 作业
type Job struct {
	ID      string
	Type    string
	Data    interface{}
	Handler func(interface{}) (interface{}, error)
}

// Result 结果
type Result struct {
	ID      string
	Type    string
	Data    interface{}
	Error   error
	Duration time.Duration
}

// NewConcurrentProcessor 创建并发处理器
func NewConcurrentProcessor(workers int) *ConcurrentProcessor {
	return &ConcurrentProcessor{
		workers:     workers,
		jobQueue:    make(chan Job, workers*2),
		resultQueue: make(chan Result, workers*2),
		workerPool:  make(chan struct{}, workers),
	}
}

// Start 启动处理器
func (cp *ConcurrentProcessor) Start() {
	for i := 0; i < cp.workers; i++ {
		cp.wg.Add(1)
		go cp.worker()
	}
}

// Stop 停止处理器
func (cp *ConcurrentProcessor) Stop() {
	close(cp.jobQueue)
	cp.wg.Wait()
	close(cp.resultQueue)
}

// Submit 提交作业
func (cp *ConcurrentProcessor) Submit(job Job) {
	cp.jobQueue <- job
}

// GetResult 获取结果
func (cp *ConcurrentProcessor) GetResult() Result {
	return <-cp.resultQueue
}

// worker 工作协程
func (cp *ConcurrentProcessor) worker() {
	defer cp.wg.Done()

	for job := range cp.jobQueue {
		cp.workerPool <- struct{}{}

		start := time.Now()
		result, err := job.Handler(job.Data)
		duration := time.Since(start)

		cp.resultQueue <- Result{
			ID:       job.ID,
			Type:     job.Type,
			Data:     result,
			Error:    err,
			Duration: duration,
		}

		<-cp.workerPool
	}
}

// BatchProcessor 批处理器
type BatchProcessor struct {
	batchSize   int
	flushInterval time.Duration
	processor   func([]interface{}) ([]interface{}, error)
	currentBatch []interface{}
	lastFlush    time.Time
	mu          sync.Mutex
}

// NewBatchProcessor 创建批处理器
func NewBatchProcessor(batchSize int, flushInterval time.Duration, processor func([]interface{}) ([]interface{}, error)) *BatchProcessor {
	return &BatchProcessor{
		batchSize:     batchSize,
		flushInterval: flushInterval,
		processor:     processor,
		lastFlush:     time.Now(),
	}
}

// Add 添加项目到批次
func (bp *BatchProcessor) Add(item interface{}) ([]interface{}, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.currentBatch = append(bp.currentBatch, item)

	// 检查是否需要刷新
	if len(bp.currentBatch) >= bp.batchSize ||
	   time.Since(bp.lastFlush) >= bp.flushInterval {
		return bp.flush()
	}

	return nil, nil
}

// Flush 刷新当前批次
func (bp *BatchProcessor) Flush() ([]interface{}, error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	return bp.flush()
}

// flush 内部刷新方法
func (bp *BatchProcessor) flush() ([]interface{}, error) {
	if len(bp.currentBatch) == 0 {
		return nil, nil
	}

	batch := make([]interface{}, len(bp.currentBatch))
	copy(batch, bp.currentBatch)
	bp.currentBatch = bp.currentBatch[:0]
	bp.lastFlush = time.Now()

	return bp.processor(batch)
}

// MemoryOptimizedCache 内存优化缓存
type MemoryOptimizedCache struct {
	hotData    map[string]*CacheEntry
	coldData    map[string]*CacheEntry
	hotLimit    int
	totalLimit  int
	mu          sync.RWMutex
}

// NewMemoryOptimizedCache 创建内存优化缓存
func NewMemoryOptimizedCache(hotLimit, totalLimit int) *MemoryOptimizedCache {
	return &MemoryOptimizedCache{
		hotData:   make(map[string]*CacheEntry),
		coldData:   make(map[string]*CacheEntry),
		hotLimit:   hotLimit,
		totalLimit: totalLimit,
	}
}

// Get 获取缓存值
func (moc *MemoryOptimizedCache) Get(key string) (string, bool) {
	moc.mu.RLock()
	defer moc.mu.RUnlock()

	// 优先检查热数据
	if entry, exists := moc.hotData[key]; exists {
		if !entry.Expired() {
			return entry.Value, true
		}
		delete(moc.hotData, key)
		return "", false
	}

	// 检查冷数据
	if entry, exists := moc.coldData[key]; exists {
		if !entry.Expired() {
			// 提升到热数据
			delete(moc.coldData, key)
			moc.hotData[key] = entry
			return entry.Value, true
		}
		delete(moc.coldData, key)
	}

	return "", false
}

// Set 设置缓存值
func (moc *MemoryOptimizedCache) Set(key, value string, ttl time.Duration) {
	moc.mu.Lock()
	defer moc.mu.Unlock()

	entry := &CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(ttl),
		HitCount:   0,
	}

	// 检查总数限制
	if len(moc.hotData)+len(moc.coldData) >= moc.totalLimit {
		moc.evictLRU()
	}

	// 添加到热数据
	moc.hotData[key] = entry

	// 如果热数据过多，移动一些到冷数据
	if len(moc.hotData) > moc.hotLimit {
		moc.demoteToCold()
	}
}

// Expired 检查条目是否过期
func (ce *CacheEntry) Expired() bool {
	return time.Now().After(ce.Expiration)
}

// evictLRU 使用LRU策略淘汰
func (moc *MemoryOptimizedCache) evictLRU() {
	var oldestKey string
	var oldestHit int64 = -1

	// 检查热数据
	for key, entry := range moc.hotData {
		if oldestKey == "" || entry.HitCount < oldestHit {
			oldestKey = key
			oldestHit = entry.HitCount
		}
	}

	// 检查冷数据
	for key, entry := range moc.coldData {
		if oldestKey == "" || entry.HitCount < oldestHit {
			oldestKey = key
			oldestHit = entry.HitCount
		}
	}

	if oldestKey != "" {
		delete(moc.hotData, oldestKey)
		delete(moc.coldData, oldestKey)
	}
}

// demoteToCold 将热数据降级到冷数据
func (moc *MemoryOptimizedCache) demoteToCold() {
	var lowestHitKey string
	var lowestHit int64 = -1

	for key, entry := range moc.hotData {
		if lowestHitKey == "" || entry.HitCount < lowestHit {
			lowestHitKey = key
			lowestHit = entry.HitCount
		}
	}

	if lowestHitKey != "" {
		if entry, exists := moc.hotData[lowestHitKey]; exists {
			moc.coldData[lowestHitKey] = entry
			delete(moc.hotData, lowestHitKey)
		}
	}
}