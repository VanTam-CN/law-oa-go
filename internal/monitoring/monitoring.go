package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
)

// SystemMetrics 系统指标
type SystemMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	// CPU指标
	CPUUsage float64 `json:"cpu_usage"`
	CPUCount int32   `json:"cpu_count"`
	CPUFreq  float64 `json:"cpu_freq"`

	// 内存指标
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryUsage float64 `json:"memory_usage"`

	// 磁盘指标
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskUsage   float64 `json:"disk_usage"`
	DiskIORead  uint64  `json:"disk_io_read"`
	DiskIOWrite uint64  `json:"disk_io_write"`

	// 网络指标
	NetBytesSent uint64 `json:"net_bytes_sent"`
	NetBytesRecv uint64 `json:"net_bytes_recv"`

	// Go运行时指标
	Goroutines int    `json:"goroutines"`
	GCGC       uint32 `json:"gc_gc"`
	GCPause    uint64 `json:"gc_pause"`

	// 应用指标
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	AvgResponse  float64 `json:"avg_response_time"`

	// 数据库指标
	DBConnections int `json:"db_connections"`
	DBSlowQueries int `json:"db_slow_queries"`

	// Redis指标
	RedisUsedMemory     uint64 `json:"redis_used_memory"`
	RedisConnections    int    `json:"redis_connections"`
	RedisKeyspaceHits   int64  `json:"redis_keyspace_hits"`
	RedisKeyspaceMisses int64  `json:"redis_keyspace_misses"`
}

// AlertRule 告警规则
type AlertRule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Metric    string    `json:"metric"`
	Operator  string    `json:"operator"` // >, <, >=, <=, ==, !=
	Threshold float64   `json:"threshold"`
	Duration  int       `json:"duration"` // 持续时间（秒）
	Level     string    `json:"level"`    // info, warning, error, critical
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Alert 告警
type Alert struct {
	ID         string     `json:"id"`
	RuleID     string     `json:"rule_id"`
	RuleName   string     `json:"rule_name"`
	Level      string     `json:"level"`
	Message    string     `json:"message"`
	Value      float64    `json:"value"`
	Threshold  float64    `json:"threshold"`
	Status     string     `json:"status"` // active, resolved
	StartedAt  time.Time  `json:"started_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// MonitoringService 监控服务
type MonitoringService struct {
	config *config.Config
	db     *gorm.DB

	// Prometheus指标
	requestTotal    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	errorTotal      *prometheus.CounterVec

	// 应用指标
	requestCount  int64
	errorCount    int64
	responseTimes []float64

	// 系统指标缓存
	lastMetrics  *SystemMetrics
	metricsCache []*SystemMetrics

	// 告警相关
	alertRules   []*AlertRule
	activeAlerts map[string]*Alert
	alertHistory []*Alert

	// 并发控制
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(cfg *config.Config, db *gorm.DB) *MonitoringService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &MonitoringService{
		config:        cfg,
		db:            db,
		requestCount:  0,
		responseTimes: make([]float64, 0, 1000),
		metricsCache:  make([]*SystemMetrics, 0, 1440), // 保存24小时数据
		activeAlerts:  make(map[string]*Alert),
		alertHistory:  make([]*Alert, 0),
		ctx:           ctx,
		cancel:        cancel,
	}

	// 初始化Prometheus指标
	service.initPrometheusMetrics()

	// 加载告警规则
	service.loadAlertRules()

	// 启动监控
	go service.startMonitoring()

	return service
}

// initPrometheusMetrics 初始化Prometheus指标
func (s *MonitoringService) initPrometheusMetrics() {
	s.requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "law_oa_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	s.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "law_oa_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	s.errorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "law_oa_http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "endpoint", "error_type"},
	)

	// 注册指标
	prometheus.MustRegister(s.requestTotal)
	prometheus.MustRegister(s.requestDuration)
	prometheus.MustRegister(s.errorTotal)
}

// startMonitoring 启动监控
func (s *MonitoringService) startMonitoring() {
	// 立即收集一次指标
	s.collectMetrics()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collectMetrics()
		case <-s.ctx.Done():
			return
		}
	}
}

// collectMetrics 收集系统指标
func (s *MonitoringService) collectMetrics() {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// 收集CPU指标
	cpuPercent, _ := cpu.Percent(time.Second, false)
	if len(cpuPercent) > 0 {
		metrics.CPUUsage = cpuPercent[0]
	}

	cpuInfo, _ := cpu.Info()
	if len(cpuInfo) > 0 {
		metrics.CPUCount = cpuInfo[0].Cores
		metrics.CPUFreq = float64(cpuInfo[0].Mhz)
	}

	// 收集内存指标
	memInfo, _ := mem.VirtualMemory()
	metrics.MemoryTotal = memInfo.Total
	metrics.MemoryUsed = memInfo.Used
	metrics.MemoryUsage = memInfo.UsedPercent

	// 收集磁盘指标
	diskInfo, _ := disk.Usage("/")
	metrics.DiskTotal = diskInfo.Total
	metrics.DiskUsed = diskInfo.Used
	metrics.DiskUsage = diskInfo.UsedPercent

	// 收集网络指标
	netIO, _ := net.IOCounters(false)
	if len(netIO) > 0 {
		metrics.NetBytesSent = netIO[0].BytesSent
		metrics.NetBytesRecv = netIO[0].BytesRecv
	}

	// 收集Go运行时指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	metrics.Goroutines = runtime.NumGoroutine()
	metrics.GCGC = m.NumGC
	metrics.GCPause = m.PauseTotalNs

	// 收集应用指标
	metrics.RequestCount = s.requestCount
	metrics.ErrorCount = s.errorCount

	if len(s.responseTimes) > 0 {
		var sum float64
		for _, t := range s.responseTimes {
			sum += t
		}
		metrics.AvgResponse = sum / float64(len(s.responseTimes))
	}

	// 收集数据库指标
	s.collectDBMetrics(metrics)

	// 收集Redis指标
	s.collectRedisMetrics(metrics)

	// 更新缓存
	s.mu.Lock()
	s.lastMetrics = metrics
	s.metricsCache = append(s.metricsCache, metrics)

	// 保留最近24小时的数据
	cutoff := time.Now().Add(-24 * time.Hour)
	i := 0
	for i < len(s.metricsCache) {
		if s.metricsCache[i].Timestamp.After(cutoff) {
			break
		}
		i++
	}
	s.metricsCache = s.metricsCache[i:]

	// 检查告警
	s.checkAlerts(metrics)
	s.mu.Unlock()

	// 保存到数据库
	go s.saveMetricsToDB(metrics)
}

// collectDBMetrics 收集数据库指标
func (s *MonitoringService) collectDBMetrics(metrics *SystemMetrics) {
	var dbStats sql.DBStats
	if db, err := s.db.DB(); err == nil {
		dbStats = db.Stats()
		metrics.DBConnections = dbStats.OpenConnections
	}

	// 这里可以添加慢查询统计
	// metrics.DBSlowQueries = s.getSlowQueryCount()
}

// collectRedisMetrics 收集Redis指标
func (s *MonitoringService) collectRedisMetrics(metrics *SystemMetrics) {
	// 这里需要连接Redis获取指标
	// 由于Redis配置在cfg中，这里只是示例
	// 实际实现需要根据Redis客户端来获取
}

// checkAlerts 检查告警
func (s *MonitoringService) checkAlerts(metrics *SystemMetrics) {
	for _, rule := range s.alertRules {
		if !rule.Enabled {
			continue
		}

		value := s.getMetricValue(metrics, rule.Metric)
		if value == nil {
			continue
		}

		triggered := false
		switch rule.Operator {
		case ">":
			triggered = *value > rule.Threshold
		case "<":
			triggered = *value < rule.Threshold
		case ">=":
			triggered = *value >= rule.Threshold
		case "<=":
			triggered = *value <= rule.Threshold
		case "==":
			triggered = *value == rule.Threshold
		case "!=":
			triggered = *value != rule.Threshold
		}

		if triggered {
			s.triggerAlert(rule, *value, metrics.Timestamp)
		} else {
			s.resolveAlert(rule.ID, metrics.Timestamp)
		}
	}
}

// getMetricValue 获取指标值
func (s *MonitoringService) getMetricValue(metrics *SystemMetrics, metric string) *float64 {
	switch metric {
	case "cpu_usage":
		return &metrics.CPUUsage
	case "memory_usage":
		return &metrics.MemoryUsage
	case "disk_usage":
		return &metrics.DiskUsage
	case "goroutines":
		return float64Ptr(float64(metrics.Goroutines))
	case "request_count":
		return float64Ptr(float64(metrics.RequestCount))
	case "error_count":
		return float64Ptr(float64(metrics.ErrorCount))
	case "avg_response_time":
		return &metrics.AvgResponse
	default:
		return nil
	}
}

// triggerAlert 触发告警
func (s *MonitoringService) triggerAlert(rule *AlertRule, value float64, timestamp time.Time) {
	alertID := fmt.Sprintf("%s_%d", rule.ID, timestamp.Unix())

	if _, exists := s.activeAlerts[alertID]; !exists {
		alert := &Alert{
			ID:        alertID,
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Level:     rule.Level,
			Message:   fmt.Sprintf("%s: %.2f %s %.2f", rule.Name, value, rule.Operator, rule.Threshold),
			Value:     value,
			Threshold: rule.Threshold,
			Status:    "active",
			StartedAt: timestamp,
			CreatedAt: timestamp,
		}

		s.activeAlerts[alertID] = alert
		s.alertHistory = append(s.alertHistory, alert)

		// 发送告警通知
		go s.sendAlertNotification(alert)
	}
}

// resolveAlert 解决告警
func (s *MonitoringService) resolveAlert(ruleID string, timestamp time.Time) {
	for id, alert := range s.activeAlerts {
		if alert.RuleID == ruleID && alert.Status == "active" {
			alert.Status = "resolved"
			alert.ResolvedAt = &timestamp
			delete(s.activeAlerts, id)
		}
	}
}

// sendAlertNotification 发送告警通知
func (s *MonitoringService) sendAlertNotification(alert *Alert) {
	// 这里可以实现各种通知方式
	// 如邮件、短信、钉钉、企业微信等
	log.Printf("ALERT [%s]: %s", alert.Level, alert.Message)
}

// RecordRequest 记录请求
func (s *MonitoringService) RecordRequest(method, endpoint string, statusCode int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.requestCount++
	s.responseTimes = append(s.responseTimes, duration.Seconds())

	// 保留最近1000个响应时间
	if len(s.responseTimes) > 1000 {
		s.responseTimes = s.responseTimes[1:]
	}

	// 记录Prometheus指标
	s.requestTotal.WithLabelValues(method, endpoint, strconv.Itoa(statusCode)).Inc()
	s.requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())

	if statusCode >= 400 {
		s.errorCount++
		errorType := "client_error"
		if statusCode >= 500 {
			errorType = "server_error"
		}
		s.errorTotal.WithLabelValues(method, endpoint, errorType).Inc()
	}
}

// GetSystemMetrics 获取系统指标
func (s *MonitoringService) GetSystemMetrics() *SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastMetrics
}

// GetMetricsHistory 获取历史指标
func (s *MonitoringService) GetMetricsHistory(start, end time.Time) []*SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*SystemMetrics
	for _, metrics := range s.metricsCache {
		if metrics.Timestamp.After(start) && metrics.Timestamp.Before(end) {
			result = append(result, metrics)
		}
	}
	return result
}

// GetActiveAlerts 获取活跃告警
func (s *MonitoringService) GetActiveAlerts() []*Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*Alert, 0, len(s.activeAlerts))
	for _, alert := range s.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetAlertHistory 获取告警历史
func (s *MonitoringService) GetAlertHistory(limit int) []*Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.alertHistory) {
		limit = len(s.alertHistory)
	}

	return s.alertHistory[len(s.alertHistory)-limit:]
}

// loadAlertRules 加载告警规则
func (s *MonitoringService) loadAlertRules() {
	// 从数据库加载告警规则
	// 这里先使用默认规则
	s.alertRules = []*AlertRule{
		{
			ID:        "cpu_high",
			Name:      "CPU使用率过高",
			Metric:    "cpu_usage",
			Operator:  ">",
			Threshold: 80.0,
			Duration:  300,
			Level:     "warning",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "memory_high",
			Name:      "内存使用率过高",
			Metric:    "memory_usage",
			Operator:  ">",
			Threshold: 85.0,
			Duration:  300,
			Level:     "warning",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "disk_high",
			Name:      "磁盘使用率过高",
			Metric:    "disk_usage",
			Operator:  ">",
			Threshold: 90.0,
			Duration:  300,
			Level:     "error",
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// saveMetricsToDB 保存指标到数据库
func (s *MonitoringService) saveMetricsToDB(metrics *SystemMetrics) {
	// 将指标保存到数据库
	// 这里可以实现持久化存储
}

// Close 关闭监控服务
func (s *MonitoringService) Close() {
	s.cancel()
}

// Helper functions
func float64Ptr(f float64) *float64 {
	return &f
}

// PrometheusMiddleware Prometheus中间件
func (s *MonitoringService) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start)
		s.RecordRequest(c.Request.Method, c.FullPath(), c.Writer.Status(), duration)
	}
}

// MetricsHandler Prometheus指标处理器
func (s *MonitoringService) MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	}
}
