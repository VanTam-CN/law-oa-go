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
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// HTTP请求持续时间
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// 当前活跃连接数
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// 数据库连接池指标
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	dbConnectionsWaiting = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_waiting",
			Help: "Number of waiting database connections",
		},
	)

	// 数据库查询指标
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"operation", "table"},
	)

	// Redis缓存指标
	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	cacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
		},
		[]string{"operation", "cache_type"},
	)

	// 业务指标
	usersTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "users_total",
			Help: "Total number of users",
		},
		[]string{"role", "status"},
	)

	casesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cases_total",
			Help: "Total number of cases",
		},
		[]string{"status", "priority"},
	)

	clientsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "clients_total",
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

// RecordDBMetrics 记录数据库指标
func RecordDBMetrics(operation, table string, duration time.Duration, err error) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())

	if err != nil {
		errorsTotal.WithLabelValues("database_error", "db").Inc()
	}
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

// UpdateDBConnectionMetrics 更新数据库连接池指标
func UpdateDBConnectionMetrics(active, idle, waiting int) {
	dbConnectionsActive.Set(float64(active))
	dbConnectionsIdle.Set(float64(idle))
	dbConnectionsWaiting.Set(float64(waiting))
}

// RecordError 记录错误
func RecordError(errorType, component string) {
	errorsTotal.WithLabelValues(errorType, component).Inc()
}
