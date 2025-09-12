package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 业务监控指标
var (
	// 案件相关指标
	caseCreationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "case_creation_duration_seconds",
			Help:    "Case creation duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
		},
		[]string{"case_type", "priority"},
	)

	caseUpdatesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "case_updates_total",
			Help: "Total number of case updates",
		},
		[]string{"case_type", "update_type"},
	)

	// 用户相关指标
	userLoginDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_login_duration_seconds",
			Help:    "User login duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"role", "result"},
	)

	userRegistrationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
		[]string{"role"},
	)

	// 客户相关指标
	clientCreationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "client_creation_duration_seconds",
			Help:    "Client creation duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"client_type"},
	)

	// 数据库性能指标
	dbSlowQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_slow_queries_total",
			Help: "Total number of slow database queries",
		},
		[]string{"operation", "table", "threshold"},
	)

	// 缓存性能指标
	cacheHitRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "Cache hit rate percentage",
		},
		[]string{"cache_type"},
	)

	// 业务逻辑指标
	activeUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of active users",
		},
		[]string{"role"},
	)

	revenueTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "revenue_total",
			Help: "Total revenue",
		},
		[]string{"case_type", "billing_method"},
	)
)

// BusinessMetricsMiddleware 业务监控中间件
func BusinessMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		
		// 根据路由路径记录业务指标
		switch c.FullPath() {
		case "/api/v1/cases":
			if c.Request.Method == "POST" {
				caseType := c.PostForm("case_type")
				priority := c.PostForm("priority")
				caseCreationDuration.WithLabelValues(caseType, priority).Observe(duration.Seconds())
			}
		case "/api/v1/auth/login":
			role := c.GetString("user_role")
			result := "success"
			if c.Writer.Status() >= 400 {
				result = "failure"
			}
			userLoginDuration.WithLabelValues(role, result).Observe(duration.Seconds())
		case "/api/v1/auth/register":
			role := c.PostForm("role")
			userRegistrationsTotal.WithLabelValues(role).Inc()
		case "/api/v1/clients":
			if c.Request.Method == "POST" {
				clientType := c.PostForm("client_type")
				clientCreationDuration.WithLabelValues(clientType).Observe(duration.Seconds())
			}
		}
	}
}

// RecordCaseUpdate 记录案件更新
func RecordCaseUpdate(caseType, updateType string) {
	caseUpdatesTotal.WithLabelValues(caseType, updateType).Inc()
}

// RecordSlowQuery 记录慢查询
func RecordSlowQuery(operation, table string, duration time.Duration) {
	threshold := "1s"
	if duration > 5*time.Second {
		threshold = "5s"
	} else if duration > 2*time.Second {
		threshold = "2s"
	}
	
	dbSlowQueriesTotal.WithLabelValues(operation, table, threshold).Inc()
}

// UpdateActiveUsers 更新活跃用户数
func UpdateActiveUsers(role string, count int) {
	activeUsers.WithLabelValues(role).Set(float64(count))
}

// RecordRevenue 记录收入
func RecordRevenue(caseType, billingMethod string, amount float64) {
	revenueTotal.WithLabelValues(caseType, billingMethod).Add(amount)
}

// UpdateCacheHitRate 更新缓存命中率
func UpdateCacheHitRate(cacheType string, hitRate float64) {
	cacheHitRate.WithLabelValues(cacheType).Set(hitRate)
}