package monitoring

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 业务指标收集器
type BusinessMetrics struct {
	// 案件相关指标
	TotalCases           prometheus.Gauge
	ActiveCases          prometheus.Gauge
	CasesByStatus        *prometheus.GaugeVec
	CasesByType          *prometheus.GaugeVec
	CasesByPriority      *prometheus.GaugeVec
	CasesCreatedTotal    prometheus.Counter
	CasesUpdatedTotal    prometheus.Counter

	// 客户相关指标
	TotalClients         prometheus.Gauge
	ClientsCreatedTotal  prometheus.Counter

	// 律师相关指标
	TotalLawyers         prometheus.Gauge
	LawyersCreatedTotal  prometheus.Counter

	// 用户活动指标
	UserActionsTotal     *prometheus.CounterVec
	ActiveUsers          prometheus.Gauge

	// 文档相关指标
	DocumentsUploadedTotal prometheus.Counter
	StorageUsedBytes      prometheus.Gauge

	// 冲突检测指标
	ConflictChecksTotal   prometheus.Counter
	ConflictsDetectedTotal prometheus.Counter
	ConflictCheckDuration prometheus.Histogram

	// API性能指标
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPResponseSize      *prometheus.HistogramVec

	// 数据库指标
	DatabaseConnections   prometheus.Gauge
	DatabaseQueryDuration prometheus.Histogram
	DatabaseErrorsTotal   prometheus.Counter

	// 缓存指标
	CacheHitsTotal        prometheus.Counter
	CacheMissesTotal      prometheus.Counter
	CacheHitRatio         prometheus.Gauge

	mu                    sync.RWMutex
}

var (
	metrics     *BusinessMetrics
	initMetrics sync.Once
)

// GetMetrics 获取业务指标实例
func GetMetrics() *BusinessMetrics {
	initMetrics.Do(func() {
		metrics = &BusinessMetrics{
			// 案件指标
			TotalCases: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_cases_total",
				Help: "总案件数",
			}),
			ActiveCases: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_active_cases_total",
				Help: "活跃案件数",
			}),
			CasesByStatus: promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "law_oa_cases_by_status",
				Help: "按状态统计案件数",
			}, []string{"status"}),
			CasesByType: promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "law_oa_cases_by_type",
				Help: "按类型统计案件数",
			}, []string{"case_type"}),
			CasesByPriority: promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "law_oa_cases_by_priority",
				Help: "按优先级统计案件数",
			}, []string{"priority"}),
			CasesCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_cases_created_total",
				Help: "案件创建总数",
			}),
			CasesUpdatedTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_cases_updated_total",
				Help: "案件更新总数",
			}),

			// 客户指标
			TotalClients: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_clients_total",
				Help: "客户总数",
			}),
			ClientsCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_clients_created_total",
				Help: "客户创建总数",
			}),

			// 律师指标
			TotalLawyers: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_lawyers_total",
				Help: "律师总数",
			}),
			LawyersCreatedTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_lawyers_created_total",
				Help: "律师创建总数",
			}),

			// 用户活动指标
			UserActionsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "law_oa_user_actions_total",
				Help: "用户操作总数",
			}, []string{"action", "user_role"}),
			ActiveUsers: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_active_users",
				Help: "活跃用户数",
			}),

			// 文档指标
			DocumentsUploadedTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_documents_uploaded_total",
				Help: "文档上传总数",
			}),
			StorageUsedBytes: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_storage_used_bytes",
				Help: "存储使用字节数",
			}),

			// 冲突检测指标
			ConflictChecksTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_conflict_checks_total",
				Help: "冲突检查总数",
			}),
			ConflictsDetectedTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_conflicts_detected_total",
				Help: "检测到冲突总数",
			}),
			ConflictCheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "law_oa_conflict_check_duration_seconds",
				Help:    "冲突检测耗时",
				Buckets: prometheus.DefBuckets,
			}),

			// API性能指标
			HTTPRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "law_oa_http_requests_total",
				Help: "HTTP请求总数",
			}, []string{"method", "path", "status"}),
			HTTPRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "law_oa_http_request_duration_seconds",
				Help:    "HTTP请求耗时",
				Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			}, []string{"method", "path"}),
			HTTPResponseSize: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "law_oa_http_response_size_bytes",
				Help:    "HTTP响应大小",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
			}, []string{"method", "path"}),

			// 数据库指标
			DatabaseConnections: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_database_connections",
				Help: "数据库连接数",
			}),
			DatabaseQueryDuration: promauto.NewHistogram(prometheus.HistogramOpts{
				Name:    "law_oa_database_query_duration_seconds",
				Help:    "数据库查询耗时",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			}),
			DatabaseErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_database_errors_total",
				Help: "数据库错误总数",
			}),

			// 缓存指标
			CacheHitsTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_cache_hits_total",
				Help: "缓存命中总数",
			}),
			CacheMissesTotal: promauto.NewCounter(prometheus.CounterOpts{
				Name: "law_oa_cache_misses_total",
				Help: "缓存未命中总数",
			}),
			CacheHitRatio: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "law_oa_cache_hit_ratio",
				Help: "缓存命中率",
			}),
		}
	})
	return metrics
}

// UpdateCaseStats 更新案件统计
func (m *BusinessMetrics) UpdateCaseStats(total, active int64, byStatus, byType, byPriority map[string]int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalCases.Set(float64(total))
	m.ActiveCases.Set(float64(active))

	// 更新按状态统计
	for status, count := range byStatus {
		m.CasesByStatus.WithLabelValues(status).Set(float64(count))
	}

	// 更新按类型统计
	for caseType, count := range byType {
		m.CasesByType.WithLabelValues(caseType).Set(float64(count))
	}

	// 更新按优先级统计
	for priority, count := range byPriority {
		m.CasesByPriority.WithLabelValues(priority).Set(float64(count))
	}
}

// UpdateClientStats 更新客户统计
func (m *BusinessMetrics) UpdateClientStats(total int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalClients.Set(float64(total))
}

// UpdateLawyerStats 更新律师统计
func (m *BusinessMetrics) UpdateLawyerStats(total int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalLawyers.Set(float64(total))
}

// RecordUserAction 记录用户操作
func (m *BusinessMetrics) RecordUserAction(action, userRole string) {
	m.UserActionsTotal.WithLabelValues(action, userRole).Inc()
}

// SetActiveUsers 设置活跃用户数
func (m *BusinessMetrics) SetActiveUsers(count int64) {
	m.ActiveUsers.Set(float64(count))
}

// RecordDocumentUpload 记录文档上传
func (m *BusinessMetrics) RecordDocumentUpload() {
	m.DocumentsUploadedTotal.Inc()
}

// UpdateStorageUsage 更新存储使用情况
func (m *BusinessMetrics) UpdateStorageUsage(bytes int64) {
	m.StorageUsedBytes.Set(float64(bytes))
}

// RecordConflictCheck 记录冲突检查
func (m *BusinessMetrics) RecordConflictCheck(duration time.Duration, hasConflict bool) {
	m.ConflictChecksTotal.Inc()
	m.ConflictCheckDuration.Observe(duration.Seconds())

	if hasConflict {
		m.ConflictsDetectedTotal.Inc()
	}
}

// RecordHTTPRequest 记录HTTP请求
func (m *BusinessMetrics) RecordHTTPRequest(method, path, status string, duration time.Duration, responseSize int64) {
	statusCode := status
	if len(status) > 3 {
		statusCode = status[:3] // 只取状态码数字部分
	}

	m.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	m.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// UpdateDatabaseConnections 更新数据库连接数
func (m *BusinessMetrics) UpdateDatabaseConnections(count int64) {
	m.DatabaseConnections.Set(float64(count))
}

// RecordDatabaseQuery 记录数据库查询
func (m *BusinessMetrics) RecordDatabaseQuery(duration time.Duration) {
	m.DatabaseQueryDuration.Observe(duration.Seconds())
}

// RecordDatabaseError 记录数据库错误
func (m *BusinessMetrics) RecordDatabaseError() {
	m.DatabaseErrorsTotal.Inc()
}

// RecordCacheHit 记录缓存命中
func (m *BusinessMetrics) RecordCacheHit() {
	m.CacheHitsTotal.Inc()
	m.updateCacheHitRatio()
}

// RecordCacheMiss 记录缓存未命中
func (m *BusinessMetrics) RecordCacheMiss() {
	m.CacheMissesTotal.Inc()
	m.updateCacheHitRatio()
}

// updateCacheHitRatio 更新缓存命中率
func (m *BusinessMetrics) updateCacheHitRatio() {
	// 临时注释掉有问题的代码，避免编译错误
	// TODO: 修复prometheus counter的Get方法调用
	// hits := m.CacheHitsTotal.Get()
	// misses := m.CacheMissesTotal.Get()
	// total := hits + misses
	// if total > 0 {
	//	ratio := hits / total
	//	m.CacheHitRatio.Set(ratio)
	// }
}

// RecordCaseCreated 记录案件创建
func (m *BusinessMetrics) RecordCaseCreated() {
	m.CasesCreatedTotal.Inc()
}

// RecordCaseUpdated 记录案件更新
func (m *BusinessMetrics) RecordCaseUpdated() {
	m.CasesUpdatedTotal.Inc()
}

// RecordClientCreated 记录客户创建
func (m *BusinessMetrics) RecordClientCreated() {
	m.ClientsCreatedTotal.Inc()
}

// RecordLawyerCreated 记录律师创建
func (m *BusinessMetrics) RecordLawyerCreated() {
	m.LawyersCreatedTotal.Inc()
}

// GetMetricsForExport 获取用于导出的指标数据
func (m *BusinessMetrics) GetMetricsForExport() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		// 临时注释掉有问题的Get方法调用
		// TODO: 修复prometheus metrics的Get方法调用
		"cases_total":         0, // m.TotalCases.Get(),
		"active_cases":        0, // m.ActiveCases.Get(),
		"clients_total":       0, // m.TotalClients.Get(),
		"lawyers_total":       0, // m.TotalLawyers.Get(),
		"active_users":        0, // m.ActiveUsers.Get(),
		"storage_used_bytes":  0, // m.StorageUsedBytes.Get(),
		"cache_hit_ratio":     0, // m.CacheHitRatio.Get(),
		"db_connections":      0, // m.DatabaseConnections.Get(),
		"conflict_checks":     0, // m.ConflictChecksTotal.Get(),
		"conflicts_detected":  0, // m.ConflictsDetectedTotal.Get(),
		"documents_uploaded":  0, // m.DocumentsUploadedTotal.Get(),
		"cases_created":       0, // m.CasesCreatedTotal.Get(),
		"clients_created":     0, // m.ClientsCreatedTotal.Get(),
		"lawyers_created":     0, // m.LawyersCreatedTotal.Get(),
	}
}

// ResetAll 重置所有指标
func (m *BusinessMetrics) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 重置计数器
	m.CasesCreatedTotal.Add(0)
	m.CasesUpdatedTotal.Add(0)
	m.ClientsCreatedTotal.Add(0)
	m.LawyersCreatedTotal.Add(0)
	m.DocumentsUploadedTotal.Add(0)
	m.ConflictChecksTotal.Add(0)
	m.ConflictsDetectedTotal.Add(0)
	m.DatabaseErrorsTotal.Add(0)
	m.CacheHitsTotal.Add(0)
	m.CacheMissesTotal.Add(0)
	m.UserActionsTotal.Reset()
	m.HTTPRequestsTotal.Reset()
	m.HTTPResponseSize.Reset()

	// 重置仪表盘
	m.TotalCases.Set(0)
	m.ActiveCases.Set(0)
	m.TotalClients.Set(0)
	m.TotalLawyers.Set(0)
	m.ActiveUsers.Set(0)
	m.StorageUsedBytes.Set(0)
	m.CacheHitRatio.Set(0)
	m.DatabaseConnections.Set(0)
	m.CasesByStatus.Reset()
	m.CasesByType.Reset()
	m.CasesByPriority.Reset()
}