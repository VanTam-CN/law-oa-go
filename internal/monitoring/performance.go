package monitoring

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"law-oa-go/internal/logger"
)

// PerformanceMetrics 性能指标收集器
type PerformanceMetrics struct {
	// HTTP指标
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestSize       *prometheus.HistogramVec
	httpResponseSize      *prometheus.HistogramVec

	// 数据库指标
	dbConnectionsActive   prometheus.Gauge
	dbConnectionWaitTime  *prometheus.HistogramVec
	dbQueryDuration       *prometheus.HistogramVec
	dbQueriesTotal        *prometheus.CounterVec

	// Redis指标
	redisOperationsTotal  *prometheus.CounterVec
	redisOperationDuration *prometheus.HistogramVec
	redisConnectionsActive prometheus.Gauge

	// 系统指标
	cpuUsage              prometheus.Gauge
	memoryUsage           prometheus.Gauge
	goroutineCount        prometheus.Gauge
	gcDuration           *prometheus.HistogramVec

	// 应用指标
	activeConnections     prometheus.Gauge
	cachedResponses       *prometheus.CounterVec
	cacheHitRatio         *prometheus.GaugeVec

	// 自定义业务指标
	casesCreatedTotal     prometheus.Counter
	clientsRegisteredTotal prometheus.Counter
	usersLoggedInTotal    prometheus.Counter

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewPerformanceMetrics 创建性能指标收集器
func NewPerformanceMetrics() *PerformanceMetrics {
	ctx, cancel := context.WithCancel(context.Background())

	pm := &PerformanceMetrics{
		ctx:    ctx,
		cancel: cancel,

		// HTTP指标
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		httpRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method"},
		),
		httpResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "status_code"},
		),

		// 数据库指标
		dbConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
		),
		dbConnectionWaitTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_connection_wait_seconds",
				Help:    "Time spent waiting for database connections",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"database"},
		),
		dbQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2},
			},
			[]string{"operation", "table"},
		),
		dbQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		// Redis指标
		redisOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "redis_operations_total",
				Help: "Total number of Redis operations",
			},
			[]string{"operation", "status"},
		),
		redisOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "redis_operation_duration_seconds",
				Help:    "Redis operation duration in seconds",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05},
			},
			[]string{"operation"},
		),
		redisConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "redis_connections_active",
				Help: "Number of active Redis connections",
			},
		),

		// 系统指标
		cpuUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_cpu_usage_percent",
				Help: "System CPU usage percentage",
			},
		),
		memoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_memory_usage_bytes",
				Help: "System memory usage in bytes",
			},
		),
		goroutineCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "system_goroutines_count",
				Help: "Number of goroutines",
			},
		),
		gcDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "system_gc_duration_seconds",
				Help:    "Garbage collection duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"gc_type"},
		),

		// 应用指标
		activeConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "app_active_connections",
				Help: "Number of active connections",
			},
		),
		cachedResponses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "app_cached_responses_total",
				Help: "Total number of cached responses",
			},
			[]string{"cache_type", "status"},
		),
		cacheHitRatio: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "app_cache_hit_ratio",
				Help: "Cache hit ratio",
			},
			[]string{"cache_type"},
		),

		// 业务指标
		casesCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "business_cases_created_total",
				Help: "Total number of cases created",
			},
		),
		clientsRegisteredTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "business_clients_registered_total",
				Help: "Total number of clients registered",
			},
		),
		usersLoggedInTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "business_users_logged_in_total",
				Help: "Total number of user logins",
			},
		),
	}

	// 启动系统指标收集
	pm.startSystemMetricsCollection()

	return pm
}

// RecordHTTPRequest 记录HTTP请求指标
func (pm *PerformanceMetrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration, requestSize, responseSize int64) {
	status := string(statusCode)
	pm.httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	pm.httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	if requestSize > 0 {
		pm.httpRequestSize.WithLabelValues(method).Observe(float64(requestSize))
	}
	if responseSize > 0 {
		pm.httpResponseSize.WithLabelValues(method, status).Observe(float64(responseSize))
	}
}

// RecordDBQuery 记录数据库查询指标
func (pm *PerformanceMetrics) RecordDBQuery(operation, table, status string, duration time.Duration) {
	pm.dbQueriesTotal.WithLabelValues(operation, table, status).Inc()
	pm.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordRedisOperation 记录Redis操作指标
func (pm *PerformanceMetrics) RecordRedisOperation(operation, status string, duration time.Duration) {
	pm.redisOperationsTotal.WithLabelValues(operation, status).Inc()
	pm.redisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordCachedResponse 记录缓存响应指标
func (pm *PerformanceMetrics) RecordCachedResponse(cacheType, status string) {
	pm.cachedResponses.WithLabelValues(cacheType, status).Inc()
}

// UpdateDBConnections 更新数据库连接数
func (pm *PerformanceMetrics) UpdateDBConnections(count float64) {
	pm.dbConnectionsActive.Set(count)
}

// UpdateRedisConnections 更新Redis连接数
func (pm *PerformanceMetrics) UpdateRedisConnections(count float64) {
	pm.redisConnectionsActive.Set(count)
}

// UpdateActiveConnections 更新活跃连接数
func (pm *PerformanceMetrics) UpdateActiveConnections(count float64) {
	pm.activeConnections.Set(count)
}

// UpdateCacheHitRatio 更新缓存命中率
func (pm *PerformanceMetrics) UpdateCacheHitRatio(cacheType string, ratio float64) {
	pm.cacheHitRatio.WithLabelValues(cacheType).Set(ratio)
}

// IncrementCasesCreated 增加创建案件计数
func (pm *PerformanceMetrics) IncrementCasesCreated() {
	pm.casesCreatedTotal.Inc()
}

// IncrementClientsRegistered 增加注册客户计数
func (pm *PerformanceMetrics) IncrementClientsRegistered() {
	pm.clientsRegisteredTotal.Inc()
}

// IncrementUsersLoggedIn 增加用户登录计数
func (pm *PerformanceMetrics) IncrementUsersLoggedIn() {
	pm.usersLoggedInTotal.Inc()
}

// startSystemMetricsCollection 启动系统指标收集
func (pm *PerformanceMetrics) startSystemMetricsCollection() {
	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-pm.ctx.Done():
				return
			case <-ticker.C:
				pm.collectSystemMetrics()
			}
		}
	}()
}

// collectSystemMetrics 收集系统指标
func (pm *PerformanceMetrics) collectSystemMetrics() {
	// 获取内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pm.memoryUsage.Set(float64(m.Alloc))
	pm.goroutineCount.Set(float64(runtime.NumGoroutine()))

	// 记录GC指标
	pm.gcDuration.WithLabelValues("gc_cpu").Observe(float64(m.PauseTotalNs) / 1e9)

	// 获取CPU使用率（简化版本）
	// 在实际生产环境中，可能需要使用更复杂的CPU监控
	if m.Sys > 0 {
		cpuPercent := float64(m.Alloc) / float64(m.Sys) * 100
		pm.cpuUsage.Set(cpuPercent)
	}
}

// GetMetricsSummary 获取指标摘要
func (pm *PerformanceMetrics) GetMetricsSummary() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"system": map[string]interface{}{
			"goroutines":        runtime.NumGoroutine(),
			"memory_alloc":      m.Alloc,
			"memory_total_alloc": m.TotalAlloc,
			"memory_sys":        m.Sys,
			"gc_cpu_fraction":   m.GCCPUFraction,
			"num_gc":           m.NumGC,
		},
		"runtime": map[string]interface{}{
			"go_version":   runtime.Version(),
			"go_os":        runtime.GOOS,
			"go_arch":      runtime.GOARCH,
			"num_cpu":      runtime.NumCPU(),
		},
	}
}

// Stop 停止性能指标收集
func (pm *PerformanceMetrics) Stop() {
	if pm.cancel != nil {
		pm.cancel()
	}
	pm.wg.Wait()
}

// 全局性能指标实例
var DefaultPerformanceMetrics *PerformanceMetrics

// InitPerformanceMetrics 初始化性能指标
func InitPerformanceMetrics() {
	DefaultPerformanceMetrics = NewPerformanceMetrics()
	if logger.Logger != nil {
		logger.Logger.Info("Performance metrics initialized")
	}
}

// GetPerformanceMetrics 获取性能指标实例
func GetPerformanceMetrics() *PerformanceMetrics {
	if DefaultPerformanceMetrics == nil {
		InitPerformanceMetrics()
	}
	return DefaultPerformanceMetrics
}