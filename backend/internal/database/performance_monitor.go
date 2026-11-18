package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/config"
)

// PerformanceMonitor 数据库性能监控器 - 基于最新PostgreSQL最佳实践
type PerformanceMonitor struct {
	db     *gorm.DB
	config *config.DatabasePerformanceConfig
	mu     sync.RWMutex
	stats  *PerformanceStats
}

// PerformanceStats 性能统计
type PerformanceStats struct {
	// 查询统计
	TotalQueries     int64         `json:"total_queries"`
	SlowQueries      int64         `json:"slow_queries"`
	ErrorQueries      int64         `json:"error_queries"`

	// 时间统计
	AverageQueryTime time.Duration `json:"average_query_time"`
	MaxQueryTime     time.Duration `json:"max_query_time"`
	MinQueryTime     time.Duration `json:"min_query_time"`

	// 连接池统计
	OpenConnections  int         `json:"open_connections"`
	InUseConnections   int         `json:"in_use_connections"`
	IdleConnections    int         `json:"idle_connections"`
	WaitCount         int64       `json:"wait_count"`
	WaitDuration      string      `json:"wait_duration"`

	// 缓存统计
	CacheHits         int64       `json:"cache_hits"`
	CacheMisses       int64       `json:"cache_misses"`
	CacheHitRatio     float64     `json:"cache_hit_ratio"`

	// 事务统计
	TotalTransactions int64       `json:"total_transactions"`
	RollbackCount     int64       `json:"rollback_count"`

	// 锁等待统计
	LockWaits         int64       `json:"lock_waits"`
	DeadlockCount     int64       `json:"deadlock_count"`

	// 统计时间范围
	StartTime         time.Time   `json:"start_time"`
	LastUpdateTime    time.Time   `json:"last_update_time"`

	// 详细查询记录
	QueryHistory      []QueryRecord `json:"query_history"`
}

// QueryRecord 查询记录
type QueryRecord struct {
	Timestamp    time.Time     `json:"timestamp"`
	QueryType   string        `json:"query_type"`
	Duration    time.Duration `json:"duration"`
	RowsAffected int64         `json:"rows_affected"`
	Error       string        `json:"error,omitempty"`
	Parameters   []interface{} `json:"parameters,omitempty"`
}

// QueryMetrics 查询指标
type QueryMetrics struct {
	Type       string
	Duration   time.Duration
	RowsAffected int64
	Error      error
}

// NewPerformanceMonitor 创建新的性能监控器
func NewPerformanceMonitor(db *gorm.DB, config *config.DatabasePerformanceConfig) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		db:     db,
		config: config,
		stats: &PerformanceStats{
			StartTime:      time.Now(),
			LastUpdateTime: time.Now(),
			MinQueryTime:    time.Hour, // 初始化为一个大值
			QueryHistory:   make([]QueryRecord, 0),
		},
	}

	// 启动监控协程
	go pm.startMonitoring()
	go pm.startQueryTracking()

	return pm
}

// startMonitoring 启动监控
func (pm *PerformanceMonitor) startMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pm.updateConnectionStats()
		pm.updatePerformanceStats()
	}
}

// startQueryTracking 启动查询跟踪
func (pm *PerformanceMonitor) startQueryTracking() {
	// 使用GORM的回调来跟踪查询
	pm.db.Callback().Query().Before("gorm:query").Register("query_tracker_before", func(db *gorm.DB) {
		// 记录查询开始时间
		db.InstanceSet("query_start", time.Now())
	})

	pm.db.Callback().Query().After("gorm:query").Register("query_tracker_after", func(db *gorm.DB) {
		// 记录查询结束时间和统计
		if start, exists := db.InstanceGet("query_start"); exists {
			if startTime, ok := start.(time.Time); ok {
				duration := time.Since(startTime)
				pm.recordQueryMetrics(db.Statement.SQL.String(), duration, db.RowsAffected, nil)
			}
		}
	})

	pm.db.Callback().Create().Before("gorm:create").Register("create_tracker_before", func(db *gorm.DB) {
		db.InstanceSet("create_start", time.Now())
	})

	pm.db.Callback().Create().After("gorm:create").Register("create_tracker_after", func(db *gorm.DB) {
		if start, exists := db.InstanceGet("create_start"); exists {
			if startTime, ok := start.(time.Time); ok {
				duration := time.Since(startTime)
				pm.recordQueryMetrics("CREATE", duration, db.RowsAffected, nil)
			}
		}
	})

	pm.db.Callback().Update().Before("gorm:update").Register("update_tracker_before", func(db *gorm.DB) {
		db.InstanceSet("update_start", time.Now())
	})

	pm.db.Callback().Update().After("gorm:update").Register("update_tracker_after", func(db *gorm.DB) {
		if start, exists := db.InstanceGet("update_start"); exists {
			if startTime, ok := start.(time.Time); ok {
				duration := time.Since(startTime)
				pm.recordQueryMetrics("UPDATE", duration, db.RowsAffected, nil)
			}
		}
	})

	pm.db.Callback().Delete().Before("gorm:delete").Register("delete_tracker_before", func(db *gorm.DB) {
		db.InstanceSet("delete_start", time.Now())
	})

	pm.db.Callback().Delete().After("gorm:delete").Register("delete_tracker_after", func(db *gorm.DB) {
		if start, exists := db.InstanceGet("delete_start"); exists {
			if startTime, ok := start.(time.Time); ok {
				duration := time.Since(startTime)
				pm.recordQueryMetrics("DELETE", duration, db.RowsAffected, nil)
			}
		}
	})
}

// recordQueryMetrics 记录查询指标
func (pm *PerformanceMonitor) recordQueryMetrics(queryType string, duration time.Duration, rowsAffected int64, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.stats.TotalQueries++
	pm.stats.LastUpdateTime = time.Now()

	// 更新查询时间统计
	if pm.stats.TotalQueries == 1 {
		pm.stats.AverageQueryTime = duration
		pm.stats.MaxQueryTime = duration
		pm.stats.MinQueryTime = duration
	} else {
		// 计算新的平均时间
		total := pm.stats.AverageQueryTime*time.Duration(pm.stats.TotalQueries-1) + duration
		pm.stats.AverageQueryTime = total / time.Duration(pm.stats.TotalQueries)

		// 更新最大最小时间
		if duration > pm.stats.MaxQueryTime {
			pm.stats.MaxQueryTime = duration
		}
		if duration < pm.stats.MinQueryTime {
			pm.stats.MinQueryTime = duration
		}
	}

	// 检查慢查询
	if duration > 1*time.Second {
		pm.stats.SlowQueries++
	}

	// 记录错误
	if err != nil {
		pm.stats.ErrorQueries++
	}

	// 添加到查询历史记录
	queryRecord := QueryRecord{
		Timestamp:    time.Now(),
		QueryType:   queryType,
		Duration:    duration,
		RowsAffected: rowsAffected,
		Error:       "",
	}

	if err != nil {
		queryRecord.Error = err.Error()
	}

	pm.stats.QueryHistory = append(pm.stats.QueryHistory, queryRecord)

	// 限制历史记录数量
	if len(pm.stats.QueryHistory) > 1000 {
		pm.stats.QueryHistory = pm.stats.QueryHistory[1:]
	}
}

// updateConnectionStats 更新连接池统计
func (pm *PerformanceMonitor) updateConnectionStats() {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return
	}

	stats := sqlDB.Stats()

	pm.mu.Lock()
	pm.stats.OpenConnections = stats.OpenConnections
	pm.stats.InUseConnections = stats.InUse
	pm.stats.IdleConnections = stats.Idle
	pm.stats.WaitCount = int64(stats.WaitCount)
	pm.stats.WaitDuration = stats.WaitDuration.String()
	pm.mu.Unlock()
}

// updatePerformanceStats 更新性能统计
func (pm *PerformanceMonitor) updatePerformanceStats() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 计算缓存命中率
	if pm.stats.CacheHits+pm.stats.CacheMisses > 0 {
		pm.stats.CacheHitRatio = float64(pm.stats.CacheHits) / float64(pm.stats.CacheHits+pm.stats.CacheMisses)
	}

	// 这里可以添加更多的性能统计逻辑
}

// GetStats 获取性能统计
func (pm *PerformanceMonitor) GetStats() *PerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// 返回统计信息的副本
	statsCopy := *pm.stats
	return &statsCopy
}

// GetQueryHistory 获取查询历史
func (pm *PerformanceMonitor) GetQueryHistory(limit int) []QueryRecord {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	history := pm.stats.QueryHistory
	if limit > 0 && len(history) > limit {
		return history[len(history)-limit:]
	}
	return history
}

// GetSlowQueries 获取慢查询记录
func (pm *PerformanceMonitor) GetSlowQueries(limit int) []QueryRecord {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var slowQueries []QueryRecord
	for _, query := range pm.stats.QueryHistory {
		if query.Duration > 1*time.Second {
			slowQueries = append(slowQueries, query)
		}
	}

	if limit > 0 && len(slowQueries) > limit {
		return slowQueries[len(slowQueries)-limit:]
	}
	return slowQueries
}

// GetConnectionPoolInfo 获取连接池详细信息
func (pm *PerformanceMonitor) GetConnectionPoolInfo() (map[string]interface{}, error) {
	sqlDB, err := pm.db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections":    stats.MaxOpenConnections,
		"open_connections":       stats.OpenConnections,
		"in_use":                 stats.InUse,
		"idle":                   stats.Idle,
		"wait_count":             stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}

// RecordTransaction 记录事务统计
func (pm *PerformanceMonitor) RecordTransaction(duration time.Duration, isRollback bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.stats.TotalTransactions++
	if isRollback {
		pm.stats.RollbackCount++
	}
}

// RecordCacheHit 记录缓存命中
func (pm *PerformanceMonitor) RecordCacheHit() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.stats.CacheHits++
}

// RecordCacheMiss 记录缓存未命中
func (pm *PerformanceMonitor) RecordCacheMiss() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.stats.CacheMisses++
}

// AnalyzePerformance 分析性能并返回建议
func (pm *PerformanceMonitor) AnalyzePerformance() []string {
	stats := pm.GetStats()
	suggestions := make([]string, 0)

	// 分析慢查询
	if stats.SlowQueries > 0 {
		slowQueryRatio := float64(stats.SlowQueries) / float64(stats.TotalQueries) * 100
		if slowQueryRatio > 10 {
			suggestions = append(suggestions, fmt.Sprintf("慢查询比例过高: %.2f%%，建议优化查询或添加索引", slowQueryRatio))
		}
	}

	// 分析连接池使用情况
	if stats.OpenConnections > 80 {
		suggestions = append(suggestions, "数据库连接使用率过高，建议增加连接池大小")
	}

	// 分析平均查询时间
	if stats.AverageQueryTime > 500*time.Millisecond {
		suggestions = append(suggestions, "平均查询时间过长，建议优化查询或使用缓存")
	}

	// 分析缓存命中率
	if stats.CacheHitRatio < 0.5 && (stats.CacheHits+stats.CacheMisses) > 100 {
		suggestions = append(suggestions, "缓存命中率较低，建议增加缓存策略")
	}

	// 分析锁等待
	if stats.WaitCount > 0 && stats.WaitDuration != "" {
		suggestions = append(suggestions, "检测到连接等待，建议优化连接池配置")
	}

	return suggestions
}

// ResetStats 重置统计信息
func (pm *PerformanceMonitor) ResetStats() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.stats = &PerformanceStats{
		StartTime:      time.Now(),
		LastUpdateTime: time.Now(),
		MinQueryTime:    time.Hour,
		QueryHistory:   make([]QueryRecord, 0),
	}
}

// ExportStats 导出统计信息为JSON
func (pm *PerformanceMonitor) ExportStats() (string, error) {
	stats := pm.GetStats()
	return pm.statsToJSON(stats)
}

// statsToJSON 将统计信息转换为JSON
func (pm *PerformanceMonitor) statsToJSON(stats *PerformanceStats) (string, error) {
	// 这里可以使用encoding/json包来序列化
	// 为了简化，直接返回格式化的字符串
	return fmt.Sprintf(`{
		"total_queries": %d,
		"slow_queries": %d,
	"error_queries": %d,
		"average_query_time": "%s",
		"max_query_time": "%s",
	"min_query_time": "%s",
		"open_connections": %d,
	"in_use_connections": %d,
	"idle_connections": %d,
	"cache_hit_ratio": %.2f,
	"start_time": "%s"
	}`,
		stats.TotalQueries,
		stats.SlowQueries,
		stats.ErrorQueries,
		stats.AverageQueryTime,
		stats.MaxQueryTime,
		stats.MinQueryTime,
		stats.OpenConnections,
		stats.InUseConnections,
		stats.IdleConnections,
		stats.CacheHitRatio,
		stats.StartTime.Format(time.RFC3339),
	), nil
}

// Close 关闭监控器
func (pm *PerformanceMonitor) Close() error {
	// 清理资源
	pm.mu.Lock()
	pm.stats = nil
	pm.mu.Unlock()

	return nil
}

// PerformanceMonitorMiddleware 性能监控中间件
func PerformanceMonitorMiddleware(pm *PerformanceMonitor) func(next func()) func() {
	return func(next func()) func() {
		return func() {
			start := time.Now()
			next()
			duration := time.Since(start)

			// 记录请求级别的性能指标
			if duration > 5*time.Second {
				log.Printf("Slow request detected: %v", duration)
			}

			// 这里可以添加更多的性能记录逻辑
		}
	}
}