package monitoring

import (
	"context"
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
	"gorm.io/gorm"
	"law-oa-go/internal/config"
)

// SystemMetrics 系统指标
type SystemMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	// CPU指标
	CPUUsage float64 `json:"cpu_usage"`

	// 内存指标
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryUsage float64 `json:"memory_usage"`

	// 磁盘指标
	DiskTotal uint64  `json:"disk_total"`
	DiskUsed  uint64  `json:"disk_used"`
	DiskUsage float64 `json:"disk_usage"`

	// Go运行时指标
	Goroutines int `json:"goroutines"`

	// 应用指标
	RequestCount int64   `json:"request_count"`
	ErrorCount   int64   `json:"error_count"`
	AvgResponse  float64 `json:"avg_response_time"`
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
		ctx:           ctx,
		cancel:        cancel,
	}

	// 初始化Prometheus指标
	service.initPrometheusMetrics()

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

	// 收集Go运行时指标
	metrics.Goroutines = runtime.NumGoroutine()

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
	s.mu.Unlock()
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

// Close 关闭监控服务
func (s *MonitoringService) Close() {
	s.cancel()
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
