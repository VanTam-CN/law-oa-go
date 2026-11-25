// Package dbmetrics 持有 db_connection_wait_seconds 的唯一收集器实例。
// internal/database 在包初始化时取用（保持立即注册），
// internal/monitoring 在 NewPerformanceMetrics 时取用（保持延迟注册），
// 两者共享同一实例，避免默认注册器重复注册 panic。
package dbmetrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// DBConnectionWaitTime 返回仅注册一次的 db_connection_wait_seconds 收集器。
// 首次注册若 panic，该 panic 会在后续每次调用时重新抛出。
var DBConnectionWaitTime = sync.OnceValue(func() *prometheus.HistogramVec {
	return promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_connection_wait_seconds",
		Help:    "Time spent waiting for database connections",
		Buckets: prometheus.DefBuckets,
	}, []string{"database"})
})
