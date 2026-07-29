package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"law-oa-go/internal/concurrency"
)

// ApplicationMetrics 应用级别指标收集器
type ApplicationMetrics struct {
	mu sync.RWMutex

	// 系统指标
	startTime   time.Time
	uptime      prometheus.Gauge
	version     prometheus.Gauge
	buildInfo   *prometheus.GaugeVec
	environment prometheus.Gauge

	// 应用健康指标
	healthStatus    prometheus.GaugeVec
	lastHealthCheck prometheus.Gauge

	// 业务指标聚合器
	userMetrics   *BusinessMetrics
	caseMetrics   *BusinessMetrics
	clientMetrics *BusinessMetrics

	// 并发指标
	concurrencyMetrics *ConcurrencyMetrics

	// 性能指标
	performanceMetrics *PerformanceMetrics
}

// BusinessMetrics 业务指标
type BusinessMetrics struct {
	total   prometheus.GaugeVec
	created prometheus.CounterVec
	updated prometheus.CounterVec
	deleted prometheus.CounterVec
	errors  prometheus.CounterVec
}

// ConcurrencyMetrics 并发指标
type ConcurrencyMetrics struct {
	activeTasks    prometheus.Gauge
	queueSize      prometheus.Gauge
	successRate    prometheus.Gauge
	failureRate    prometheus.Gauge
	avgProcessTime prometheus.Gauge
	retryCount     prometheus.Counter
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	memoryUsage    prometheus.Gauge
	goroutineCount prometheus.Gauge
	responseTime   prometheus.HistogramVec
	throughput     prometheus.Gauge
	errorRate      prometheus.Gauge
}

var (
	// 全局指标收集器实例
	defaultMetrics *ApplicationMetrics
	once           sync.Once
)

// GetDefaultMetrics 获取默认指标收集器
func GetDefaultMetrics() *ApplicationMetrics {
	once.Do(func() {
		defaultMetrics = NewApplicationMetrics()
	})
	return defaultMetrics
}

// NewApplicationMetrics 创建新的应用指标收集器
func NewApplicationMetrics() *ApplicationMetrics {
	metrics := &ApplicationMetrics{
		startTime: time.Now(),

		// 系统指标
		uptime: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Application uptime in seconds",
		}),

		version: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_version",
			Help: "Application version",
		}),

		buildInfo: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "app_build_info",
			Help: "Application build information",
		}, []string{"version", "commit", "build_date"}),

		environment: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_environment",
			Help: "Application environment (dev=0, staging=1, prod=2)",
		}),

		// 健康检查指标
		healthStatus: *promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "app_health_status",
			Help: "Application health status (0=unhealthy, 1=healthy, 2=degraded)",
		}, []string{"component"}),

		lastHealthCheck: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_last_health_check_timestamp",
			Help: "Timestamp of last health check",
		}),

		// 业务指标
		userMetrics: &BusinessMetrics{
			total: *promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "business_users_total",
				Help: "Total number of users",
			}, []string{"role", "status", "department"}),

			created: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_users_created_total",
				Help: "Total number of users created",
			}, []string{"role", "department"}),

			updated: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_users_updated_total",
				Help: "Total number of users updated",
			}, []string{"role"}),

			deleted: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_users_deleted_total",
				Help: "Total number of users deleted",
			}, []string{"role"}),

			errors: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_users_errors_total",
				Help: "Total number of user operation errors",
			}, []string{"operation", "error_type"}),
		},

		caseMetrics: &BusinessMetrics{
			total: *promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "business_cases_total",
				Help: "Total number of cases",
			}, []string{"status", "priority", "type"}),

			created: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_cases_created_total",
				Help: "Total number of cases created",
			}, []string{"priority", "type"}),

			updated: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_cases_updated_total",
				Help: "Total number of cases updated",
			}, []string{"status"}),

			deleted: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_cases_deleted_total",
				Help: "Total number of cases deleted",
			}, []string{"status"}),

			errors: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_cases_errors_total",
				Help: "Total number of case operation errors",
			}, []string{"operation", "error_type"}),
		},

		clientMetrics: &BusinessMetrics{
			total: *promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: "business_clients_total",
				Help: "Total number of clients",
			}, []string{"status", "type", "industry"}),

			created: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_clients_created_total",
				Help: "Total number of clients created",
			}, []string{"type", "industry"}),

			updated: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_clients_updated_total",
				Help: "Total number of clients updated",
			}, []string{"type"}),

			deleted: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_clients_deleted_total",
				Help: "Total number of clients deleted",
			}, []string{"type"}),

			errors: *promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "business_clients_errors_total",
				Help: "Total number of client operation errors",
			}, []string{"operation", "error_type"}),
		},

		// 并发指标
		concurrencyMetrics: &ConcurrencyMetrics{
			activeTasks: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "concurrency_active_tasks",
				Help: "Number of currently active concurrent tasks",
			}),

			queueSize: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "concurrency_queue_size",
				Help: "Current size of the task queue",
			}),

			successRate: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "concurrency_success_rate",
				Help: "Success rate of concurrent tasks (0-1)",
			}),

			failureRate: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "concurrency_failure_rate",
				Help: "Failure rate of concurrent tasks (0-1)",
			}),

			avgProcessTime: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "concurrency_avg_process_time_seconds",
				Help: "Average processing time for concurrent tasks",
			}),

			retryCount: promauto.NewCounter(prometheus.CounterOpts{
				Name: "concurrency_retry_count_total",
				Help: "Total number of task retries",
			}),
		},

		// 性能指标
		performanceMetrics: &PerformanceMetrics{
			memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "performance_memory_usage_bytes",
				Help: "Current memory usage in bytes",
			}),

			goroutineCount: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "performance_goroutine_count",
				Help: "Current number of goroutines",
			}),

			responseTime: *promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "performance_response_time_seconds",
				Help:    "Response time distribution",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
			}, []string{"endpoint", "method", "status"}),

			throughput: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "performance_throughput_rps",
				Help: "Current requests per second",
			}),

			errorRate: promauto.NewGauge(prometheus.GaugeOpts{
				Name: "performance_error_rate",
				Help: "Current error rate (0-1)",
			}),
		},
	}

	// 启动指标收集器
	go metrics.startCollectors()

	return metrics
}

// startCollectors 启动指标收集器
func (m *ApplicationMetrics) startCollectors() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.collectSystemMetrics()
		m.collectBusinessMetrics()
		m.collectConcurrencyMetrics()
		m.collectPerformanceMetrics()
	}
}

// collectSystemMetrics 收集系统指标
func (m *ApplicationMetrics) collectSystemMetrics() {
	m.uptime.Set(time.Since(m.startTime).Seconds())
}

// collectBusinessMetrics 收集业务指标
func (m *ApplicationMetrics) collectBusinessMetrics() {
	// 这里可以从数据库或其他服务获取业务指标
	// 具体实现根据业务需求补充
}

// collectConcurrencyMetrics 收集并发指标
func (m *ApplicationMetrics) collectConcurrencyMetrics() {
	// 从并发服务收集指标
	// 具体实现根据并发服务接口补充
}

// collectPerformanceMetrics 收集性能指标
func (m *ApplicationMetrics) collectPerformanceMetrics() {
	// 收集内存使用、goroutine数量等
	// 具体实现根据 runtime 包补充
}

// UpdateUserMetrics 更新用户指标
func (m *ApplicationMetrics) UpdateUserMetrics(action, role, department string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch action {
	case "create":
		if success {
			m.userMetrics.created.WithLabelValues(role, department).Inc()
		} else {
			m.userMetrics.errors.WithLabelValues("create", "creation_failed").Inc()
		}
	case "update":
		if success {
			m.userMetrics.updated.WithLabelValues(role).Inc()
		} else {
			m.userMetrics.errors.WithLabelValues("update", "update_failed").Inc()
		}
	case "delete":
		if success {
			m.userMetrics.deleted.WithLabelValues(role).Inc()
		} else {
			m.userMetrics.errors.WithLabelValues("delete", "deletion_failed").Inc()
		}
	}
}

// UpdateCaseMetrics 更新案件指标
func (m *ApplicationMetrics) UpdateCaseMetrics(action, status, priority, caseType string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch action {
	case "create":
		if success {
			m.caseMetrics.created.WithLabelValues(priority, caseType).Inc()
		} else {
			m.caseMetrics.errors.WithLabelValues("create", "creation_failed").Inc()
		}
	case "update":
		if success {
			m.caseMetrics.updated.WithLabelValues(status).Inc()
		} else {
			m.caseMetrics.errors.WithLabelValues("update", "update_failed").Inc()
		}
	case "delete":
		if success {
			m.caseMetrics.deleted.WithLabelValues(status).Inc()
		} else {
			m.caseMetrics.errors.WithLabelValues("delete", "deletion_failed").Inc()
		}
	}
}

// UpdateClientMetrics 更新客户指标
func (m *ApplicationMetrics) UpdateClientMetrics(action, clientType, industry string, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch action {
	case "create":
		if success {
			m.clientMetrics.created.WithLabelValues(clientType, industry).Inc()
		} else {
			m.clientMetrics.errors.WithLabelValues("create", "creation_failed").Inc()
		}
	case "update":
		if success {
			m.clientMetrics.updated.WithLabelValues(clientType).Inc()
		} else {
			m.clientMetrics.errors.WithLabelValues("update", "update_failed").Inc()
		}
	case "delete":
		if success {
			m.clientMetrics.deleted.WithLabelValues(clientType).Inc()
		} else {
			m.clientMetrics.errors.WithLabelValues("delete", "deletion_failed").Inc()
		}
	}
}

// UpdateConcurrencyMetrics 更新并发指标
func (m *ApplicationMetrics) UpdateConcurrencyMetrics(poolMetrics *concurrency.PoolMetricsSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.concurrencyMetrics.activeTasks.Set(float64(poolMetrics.ActiveWorkers))
	m.concurrencyMetrics.queueSize.Set(float64(poolMetrics.QueueSize))

	totalTasks := poolMetrics.SuccessTasks + poolMetrics.FailedTasks
	if totalTasks > 0 {
		successRate := float64(poolMetrics.SuccessTasks) / float64(totalTasks)
		failureRate := float64(poolMetrics.FailedTasks) / float64(totalTasks)

		m.concurrencyMetrics.successRate.Set(successRate)
		m.concurrencyMetrics.failureRate.Set(failureRate)
	}

	m.concurrencyMetrics.avgProcessTime.Set(poolMetrics.AverageProcessTime.Seconds())
	m.concurrencyMetrics.retryCount.Add(float64(poolMetrics.RetriedTasks))
}

// UpdateHealthStatus 更新健康状态
func (m *ApplicationMetrics) UpdateHealthStatus(component string, status float64) {
	m.healthStatus.WithLabelValues(component).Set(status)
	m.lastHealthCheck.Set(float64(time.Now().Unix()))
}

// RecordResponseTime 记录响应时间
func (m *ApplicationMetrics) RecordResponseTime(endpoint, method, status string, duration time.Duration) {
	m.performanceMetrics.responseTime.WithLabelValues(endpoint, method, status).Observe(duration.Seconds())
}

// UpdateThroughput 更新吞吐量
func (m *ApplicationMetrics) UpdateThroughput(rps float64) {
	m.performanceMetrics.throughput.Set(rps)
}

// UpdateErrorRate 更新错误率
func (m *ApplicationMetrics) UpdateErrorRate(rate float64) {
	m.performanceMetrics.errorRate.Set(rate)
}

// SetVersion 设置版本信息
func (m *ApplicationMetrics) SetVersion(version string) {
	// 将版本字符串转换为数字用于显示
	// 这里简化处理，实际可以更复杂
	m.version.Set(1.0)
}

// SetEnvironment 设置环境信息
func (m *ApplicationMetrics) SetEnvironment(env string) {
	envValue := 0.0
	switch env {
	case "development":
		envValue = 0
	case "staging":
		envValue = 1
	case "production":
		envValue = 2
	}
	m.environment.Set(envValue)
}

// GetMetricsSummary 获取指标摘要
func (m *ApplicationMetrics) GetMetricsSummary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Since(m.startTime)

	return map[string]interface{}{
		"uptime":            uptime.String(),
		"version":           "1.0.0",
		"environment":       "development",
		"last_health_check": "active",
		"business": map[string]interface{}{
			"users":   "active",
			"cases":   "active",
			"clients": "active",
		},
		"concurrency": map[string]interface{}{
			"active_tasks":     "monitoring",
			"queue_size":       "monitoring",
			"success_rate":     "monitoring",
			"failure_rate":     "monitoring",
			"avg_process_time": "monitoring",
		},
		"performance": map[string]interface{}{
			"memory_usage":    "monitoring",
			"goroutine_count": "monitoring",
			"throughput":      "monitoring",
			"error_rate":      "monitoring",
		},
	}
}
