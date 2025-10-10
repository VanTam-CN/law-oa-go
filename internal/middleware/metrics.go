package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP请求总数
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "law_oa_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// HTTP请求持续时间
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "law_oa_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 当前活跃连接数
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "law_oa_http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// Redis缓存指标
	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "law_oa_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "law_oa_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	cacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "law_oa_cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
		},
		[]string{"operation", "cache_type"},
	)

	// 业务指标
	usersTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "law_oa_users_total",
			Help: "Total number of users",
		},
		[]string{"role", "status"},
	)

	casesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "law_oa_cases_total",
			Help: "Total number of cases",
		},
		[]string{"status", "priority"},
	)

	clientsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "law_oa_clients_total",
			Help: "Total number of clients",
		},
		[]string{"status"},
	)

	// 错误指标
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)
)

// PrometheusMiddleware Prometheus监控中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// 增加活跃连接数
		activeConnections.Inc()
		defer activeConnections.Dec()

		// 处理请求
		c.Next()

		// 记录指标
		duration := time.Since(start).Seconds()
		method := c.Request.Method
		endpoint := c.FullPath()
		statusCode := strconv.Itoa(c.Writer.Status())

		// 记录HTTP请求指标
		httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)

		// 记录错误指标
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			errorsTotal.WithLabelValues(errorType, "http").Inc()
		}
	})
}

// RecordCacheMetrics 记录缓存指标
func RecordCacheMetrics(operation, cacheType string, hit bool, duration time.Duration, err error) {
	if hit {
		cacheHitsTotal.WithLabelValues(cacheType).Inc()
	} else {
		cacheMissesTotal.WithLabelValues(cacheType).Inc()
	}

	cacheOperationDuration.WithLabelValues(operation, cacheType).Observe(duration.Seconds())

	if err != nil {
		errorsTotal.WithLabelValues("cache_error", "redis").Inc()
	}
}

// UpdateBusinessMetrics 更新业务指标
func UpdateBusinessMetrics(usersByRole map[string]map[string]int64,
	casesByStatus map[string]map[string]int64,
	clientsByStatus map[string]int64) {
	// 更新用户指标
	for role, statusMap := range usersByRole {
		for status, count := range statusMap {
			usersTotal.WithLabelValues(role, status).Set(float64(count))
		}
	}

	// 更新案件指标
	for status, priorityMap := range casesByStatus {
		for priority, count := range priorityMap {
			casesTotal.WithLabelValues(status, priority).Set(float64(count))
		}
	}

	// 更新客户指标
	for status, count := range clientsByStatus {
		clientsTotal.WithLabelValues(status).Set(float64(count))
	}
}

// RecordError 记录错误
func RecordError(errorType, component string) {
	errorsTotal.WithLabelValues(errorType, component).Inc()
}
