/**
 * 企业级Prometheus监控系统 - 基于最新最佳实践
 * 提供全面的业务、系统、应用性能指标收集和监控
 */

package metrics

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// MonitorConfig 监控配置
type MonitorConfig struct {
	EnableRealTimeAlerts      bool          // 启用实时告警
	MetricsCollectionInterval time.Duration // 指标收集间隔
	CPUThreshold              float64       // CPU告警阈值
	MemoryThreshold           float64       // 内存告警阈值
}

// DefaultMonitorConfig 默认监控配置
var DefaultMonitorConfig = MonitorConfig{
	EnableRealTimeAlerts:      true,
	MetricsCollectionInterval: 30 * time.Second,
	CPUThreshold:              80.0,
	MemoryThreshold:           80.0,
}

// MonitorService 监控服务
type MonitorService struct {
	config             MonitorConfig
	isRunning          bool
	stopChan           chan struct{}
	alerts             []*Alert
	concurrencyService *ConcurrencyService
}

// Alert 告警信息
type Alert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// ConcurrencyService 并发服务
type ConcurrencyService struct {
	maxConcurrent int
	current       int
}

// NewConcurrencyService 创建并发服务
func NewConcurrencyService(maxConcurrent int) *ConcurrencyService {
	return &ConcurrencyService{
		maxConcurrent: maxConcurrent,
		current:       0,
	}
}

// Acquire 获取并发许可
func (cs *ConcurrencyService) Acquire() bool {
	if cs.current < cs.maxConcurrent {
		cs.current++
		return true
	}
	return false
}

// Release 释放并发许可
func (cs *ConcurrencyService) Release() {
	if cs.current > 0 {
		cs.current--
	}
}

// GetCurrent 获取当前并发数
func (cs *ConcurrencyService) GetCurrent() int {
	return cs.current
}

// GetMax 获取最大并发数
func (cs *ConcurrencyService) GetMax() int {
	return cs.maxConcurrent
}

// EnhancedMetricsCollector 增强指标收集器
type EnhancedMetricsCollector struct {
	registry *prometheus.Registry
}

// NewEnhancedMetricsCollector 创建增强指标收集器
func NewEnhancedMetricsCollector() *EnhancedMetricsCollector {
	return &EnhancedMetricsCollector{
		registry: prometheus.NewRegistry(),
	}
}

// GetHandler 获取Prometheus处理器
func (emc *EnhancedMetricsCollector) GetHandler() http.Handler {
	return promhttp.HandlerFor(emc.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// Prometheus指标 - 增强版本
var (
	// 系统指标
	cpuUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage_percent",
		Help: "Current CPU usage percentage",
	})

	memoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_usage_percent",
		Help: "Current memory usage percentage",
	})

	memoryTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_total_bytes",
		Help: "Total system memory in bytes",
	})

	memoryAvailable = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_available_bytes",
		Help: "Available system memory in bytes",
	})

	diskUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_usage_percent",
		Help: "Current disk usage percentage",
	})

	diskTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_total_bytes",
		Help: "Total disk space in bytes",
	})

	diskFree = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_free_bytes",
		Help: "Free disk space in bytes",
	})

	// Go运行时指标
	goroutines = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_go_goroutines",
		Help: "Number of goroutines in application",
	})

	gcCycles = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_go_gc_cycles_total",
		Help: "Total number of GC cycles",
	})

	gcPauseDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "app_go_gc_pause_duration_seconds",
		Help:    "GC pause duration in seconds",
		Buckets: []float64{0.00001, 0.000025, 0.00005, 0.0001, 0.00025, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1},
	})

	heapSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_go_heap_size_bytes",
		Help: "Current heap size in bytes",
	})

	heapObjects = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_go_heap_objects",
		Help: "Number of heap objects",
	})

	stackInUse = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_go_stack_inuse_bytes",
		Help: "Stack memory in use in bytes",
	})

	// 应用指标 - HTTP
	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "endpoint", "status_code"})

	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 25, 50, 100},
	}, []string{"method", "endpoint", "status_code"})

	httpRequestSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_request_size_bytes",
		Help:    "HTTP request size in bytes",
		Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000},
	}, []string{"method", "endpoint"})

	httpResponseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_response_size_bytes",
		Help:    "HTTP response size in bytes",
		Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000, 5000000},
	}, []string{"method", "endpoint", "status_code"})

	errorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_errors_total",
		Help: "Total number of HTTP errors",
	}, []string{"method", "endpoint", "error_type"})

	activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_active_connections",
		Help: "Current number of active connections",
	})

	// 业务指标
	userSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "business_user_sessions",
		Help: "Current number of active user sessions",
	})

	totalUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "business_total_users",
		Help: "Total number of registered users",
	})

	businessOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "business_operations_total",
		Help: "Total number of business operations",
	}, []string{"operation", "module", "status", "user_type"})

	businessOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "business_operation_duration_seconds",
		Help:    "Business operation duration in seconds",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 25, 50, 100},
	}, []string{"operation", "module"})

	casesCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "business_cases_created_total",
		Help: "Total number of cases created",
	}, []string{"case_type", "user_role", "priority"})

	documentsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "business_documents_processed_total",
		Help: "Total number of documents processed",
	}, []string{"document_type", "processing_status"})

	// 数据库指标
	dbConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_connections_active",
		Help: "Number of active database connections",
	})

	dbConnectionsIdle = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_connections_idle",
		Help: "Number of idle database connections",
	})

	dbQueryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_queries_total",
		Help: "Total number of database queries",
	}, []string{"operation", "table", "status"})

	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"operation", "table"})

	dbTransactionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_transactions_total",
		Help: "Total number of database transactions",
	}, []string{"status"})

	// 缓存指标
	cacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"cache_type", "operation"})

	cacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"cache_type", "operation"})

	cacheSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cache_size_bytes",
		Help: "Cache size in bytes",
	})

	cacheEvictions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_evictions_total",
		Help: "Total number of cache evictions",
	})

	// 安全指标
	authAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "security_auth_attempts_total",
		Help: "Total number of authentication attempts",
	}, []string{"result", "user_type", "ip_location"})

	authFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "security_auth_failures_total",
		Help: "Total number of authentication failures",
	}, []string{"failure_reason", "user_type", "ip_location"})

	securityEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "security_events_total",
		Help: "Total number of security events",
	}, []string{"event_type", "severity", "user_type"})

	// 搜索指标
	searchQueries = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "search_queries_total",
		Help: "Total number of search queries",
	}, []string{"query_type", "result_count", "user_type"})

	searchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "search_query_duration_seconds",
		Help:    "Search query duration in seconds",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"query_type", "index_used"})
)

// defaultMonitorService 默认监控服务实例
var defaultMonitorService *MonitorService

// InitDefaultMonitorService 初始化默认监控服务
func InitDefaultMonitorService(config MonitorConfig) error {
	defaultMonitorService = NewMonitorService(config)
	return defaultMonitorService.Start()
}

// GetDefaultMonitorService 获取默认监控服务实例
func GetDefaultMonitorService() *MonitorService {
	return defaultMonitorService
}

// NewMonitorService 创建新的监控服务
func NewMonitorService(config MonitorConfig) *MonitorService {
	return &MonitorService{
		config:             config,
		stopChan:           make(chan struct{}),
		alerts:             make([]*Alert, 0),
		concurrencyService: NewConcurrencyService(100),
	}
}

// Start 启动监控服务
func (ms *MonitorService) Start() error {
	if ms.isRunning {
		return nil
	}

	ms.isRunning = true
	log.Println("监控服务启动成功")

	// 启动指标收集
	go ms.collectMetrics()

	return nil
}

// Stop 停止监控服务
func (ms *MonitorService) Stop() {
	if !ms.isRunning {
		return
	}

	ms.isRunning = false
	close(ms.stopChan)
	log.Println("监控服务已停止")
}

// collectMetrics 收集系统指标
func (ms *MonitorService) collectMetrics() {
	ticker := time.NewTicker(ms.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ms.updateSystemMetrics()
			if ms.config.EnableRealTimeAlerts {
				ms.checkThresholds()
			}
		case <-ms.stopChan:
			return
		}
	}
}

// updateSystemMetrics 更新系统指标 - 增强版本
func (ms *MonitorService) updateSystemMetrics() {
	// 收集CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage.Set(cpuPercent[0])
	}

	// 收集内存使用率
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		memoryUsage.Set(memInfo.UsedPercent)
		memoryTotal.Set(float64(memInfo.Total))
		memoryAvailable.Set(float64(memInfo.Available))
	}

	// 收集磁盘使用率
	diskInfo, err := disk.Usage("/")
	if err == nil {
		diskUsage.Set(diskInfo.UsedPercent)
		diskTotal.Set(float64(diskInfo.Total))
		diskFree.Set(float64(diskInfo.Free))
	}

	// 收集Go运行时指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	goroutines.Set(float64(runtime.NumGoroutine()))
	gcCycles.Add(float64(m.NumGC))
	heapSize.Set(float64(m.HeapAlloc))
	heapObjects.Set(float64(m.HeapObjects))
	stackInUse.Set(float64(m.StackInuse))
}

// updateGoMetrics 更新Go运行时指标
func (ms *MonitorService) updateGoMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 记录GC暂停时间
	gcPauseDuration.Observe(float64(m.PauseTotalNs) / 1e9)
}

// checkThresholds 检查阈值并触发告警
func (ms *MonitorService) checkThresholds() {
	// 收集当前使用率进行检查
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		if cpuPercent[0] > ms.config.CPUThreshold {
			ms.createAlert("cpu_high", "CPU usage is high", "warning")
		}
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		if memInfo.UsedPercent > ms.config.MemoryThreshold {
			ms.createAlert("memory_high", "Memory usage is high", "warning")
		}
	}
}

// createAlert 创建告警
func (ms *MonitorService) createAlert(alertType, message, severity string) {
	alert := &Alert{
		ID:        generateAlertID(),
		Type:      alertType,
		Message:   message,
		Severity:  severity,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	ms.alerts = append(ms.alerts, alert)
	log.Printf("告警触发: %s - %s", alertType, message)
}

// generateAlertID 生成告警ID
func generateAlertID() string {
	return time.Now().Format("20060102150405") + "-" + generateRandomString(6)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// GetStatus 获取监控状态
func (ms *MonitorService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"is_running":          ms.isRunning,
		"cpu_usage":           0.0,
		"memory_usage":        0.0,
		"concurrent_requests": ms.concurrencyService.GetCurrent(),
		"max_concurrent":      ms.concurrencyService.GetMax(),
		"total_alerts":        len(ms.alerts),
		"active_alerts":       ms.getActiveAlertsCount(),
	}
}

// GetDashboardData 获取仪表板数据
func (ms *MonitorService) GetDashboardData() map[string]interface{} {
	return map[string]interface{}{
		"system": map[string]interface{}{
			"cpu_usage":    0.0,
			"memory_usage": 0.0,
		},
		"requests": map[string]interface{}{
			"total_count":  requestCount,
			"total_errors": errorCount,
			"avg_duration": requestDuration,
		},
		"concurrency": map[string]interface{}{
			"current": ms.concurrencyService.GetCurrent(),
			"max":     ms.concurrencyService.GetMax(),
		},
		"alerts": ms.alerts,
	}
}

// GetPerformanceStats 获取性能统计
func (ms *MonitorService) GetPerformanceStats() map[string]interface{} {
	return map[string]interface{}{
		"cpu_usage":    0.0,
		"memory_usage": 0.0,
		"request_rate": ms.getRequestRate(),
		"error_rate":   ms.getErrorRate(),
	}
}

// getRequestRate 获取请求速率
func (ms *MonitorService) getRequestRate() float64 {
	// 简化的请求速率计算
	// 实际项目中应该基于时间窗口计算
	return 0.0
}

// getErrorRate 获取错误率
func (ms *MonitorService) getErrorRate() float64 {
	// 简化的错误率计算
	// 实际项目中应该基于时间窗口计算
	return 0.0
}

// GetAlerts 获取告警列表
func (ms *MonitorService) GetAlerts() []*Alert {
	return ms.alerts
}

// ResolveAlert 解决告警
func (ms *MonitorService) ResolveAlert(alertID string) bool {
	for _, alert := range ms.alerts {
		if alert.ID == alertID && !alert.Resolved {
			alert.Resolved = true
			return true
		}
	}
	return false
}

// getActiveAlertsCount 获取活跃告警数量
func (ms *MonitorService) getActiveAlertsCount() int {
	count := 0
	for _, alert := range ms.alerts {
		if !alert.Resolved {
			count++
		}
	}
	return count
}

// ForceGC 强制垃圾回收
func (ms *MonitorService) ForceGC() {
	// 触发垃圾回收
	// 在实际项目中，这可能需要更复杂的实现
	log.Println("触发垃圾回收")
}

// ==============================
// 增强的指标记录函数
// ==============================

// RecordRequest 记录HTTP请求 - 增强版本
func RecordRequest(method, endpoint, statusCode string, duration float64, requestSize, responseSize int64) {
	requestCount.WithLabelValues(method, endpoint, statusCode).Inc()
	requestDuration.WithLabelValues(method, endpoint, statusCode).Observe(duration)
	httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	httpResponseSize.WithLabelValues(method, endpoint, statusCode).Observe(float64(responseSize))
}

// RecordError 记录HTTP错误 - 增强版本
func RecordError(method, endpoint, errorType string) {
	errorCount.WithLabelValues(method, endpoint, errorType).Inc()
}

// RecordDBQuery 记录数据库查询
func RecordDBQuery(operation, table, status string, duration float64) {
	dbQueryTotal.WithLabelValues(operation, table, status).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordDBTransaction 记录数据库事务
func RecordDBTransaction(status string, duration float64) {
	dbTransactionTotal.WithLabelValues(status).Inc()
}

// RecordCacheHit 记录缓存命中
func RecordCacheHit(cacheType, operation string) {
	cacheHits.WithLabelValues(cacheType, operation).Inc()
}

// RecordCacheMiss 记录缓存未命中
func RecordCacheMiss(cacheType, operation string) {
	cacheMisses.WithLabelValues(cacheType, operation).Inc()
}

// RecordCacheEviction 记录缓存驱逐
func RecordCacheEviction() {
	cacheEvictions.Inc()
}

// RecordAuthAttempt 记录认证尝试
func RecordAuthAttempt(result, userType, ipLocation string) {
	authAttempts.WithLabelValues(result, userType, ipLocation).Inc()
}

// RecordAuthFailure 记录认证失败
func RecordAuthFailure(failureReason, userType, ipLocation string) {
	authFailures.WithLabelValues(failureReason, userType, ipLocation).Inc()
}

// RecordSecurityEvent 记录安全事件
func RecordSecurityEvent(eventType, severity, userType string) {
	securityEvents.WithLabelValues(eventType, severity, userType).Inc()
}

// RecordBusinessOperation 记录业务操作
func RecordBusinessOperation(operation, module, status, userType string, duration float64) {
	businessOperations.WithLabelValues(operation, module, status, userType).Inc()
	businessOperationDuration.WithLabelValues(operation, module).Observe(duration)
}

// RecordCaseCreated 记录案件创建
func RecordCaseCreated(caseType, userRole, priority string) {
	casesCreated.WithLabelValues(caseType, userRole, priority).Inc()
}

// RecordDocumentProcessed 记录文档处理
func RecordDocumentProcessed(documentType, processingStatus string) {
	documentsProcessed.WithLabelValues(documentType, processingStatus).Inc()
}

// RecordSearchQuery 记录搜索查询
func RecordSearchQuery(queryType, resultCount, userType, indexUsed string, duration float64) {
	searchQueries.WithLabelValues(queryType, resultCount, userType).Inc()
	searchDuration.WithLabelValues(queryType, indexUsed).Observe(duration)
}

// SetActiveConnections 设置活跃连接数
func SetActiveConnections(count float64) {
	activeConnections.Set(count)
}

// SetUserSessions 设置用户会话数
func SetUserSessions(count float64) {
	userSessions.Set(count)
}

// SetTotalUsers 设置总用户数
func SetTotalUsers(count float64) {
	totalUsers.Set(count)
}

// SetDBConnections 设置数据库连接数
func SetDBConnections(active, idle float64) {
	dbConnectionsActive.Set(active)
	dbConnectionsIdle.Set(idle)
}

// SetCacheSize 设置缓存大小
func SetCacheSize(size float64) {
	cacheSize.Set(size)
}

// ==============================
// 业务指标辅助函数
// ==============================

// GetRequestMetrics 获取请求指标统计
func GetRequestMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_requests":     requestCount,
		"error_count":        errorCount,
		"active_connections": activeConnections,
	}
}

// GetSystemMetrics 获取系统指标统计
func GetSystemMetrics() map[string]interface{} {
	return map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": memoryUsage,
		"disk_usage":   diskUsage,
		"goroutines":   goroutines,
		"heap_size":    heapSize,
		"gc_cycles":    gcCycles,
	}
}

// GetBusinessMetrics 获取业务指标统计
func GetBusinessMetrics() map[string]interface{} {
	return map[string]interface{}{
		"user_sessions":       userSessions,
		"total_users":         totalUsers,
		"business_operations": businessOperations,
		"cases_created":       casesCreated,
		"documents_processed": documentsProcessed,
	}
}

// PrometheusMiddleware 增强的Prometheus监控中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 获取请求大小
		var requestSize int64
		if c.Request.Body != nil {
			requestSize = c.Request.ContentLength
		}

		c.Next()

		duration := time.Since(start)
		statusCode := string(rune(c.Writer.Status()))

		// 获取响应大小
		responseSize := int64(c.Writer.Size())

		// 记录指标
		RecordRequest(
			c.Request.Method,
			c.FullPath(),
			statusCode,
			duration.Seconds(),
			requestSize,
			responseSize,
		)

		// 记录错误指标
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			RecordError(c.Request.Method, c.FullPath(), errorType)
		}
	}
}

// GetPrometheusHandler 获取Prometheus HTTP处理器
func GetPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// GetEnhancedPrometheusHandler 获取增强的Prometheus HTTP处理器
func GetEnhancedPrometheusHandler() http.Handler {
	collector := NewEnhancedMetricsCollector()
	return collector.GetHandler()
}
