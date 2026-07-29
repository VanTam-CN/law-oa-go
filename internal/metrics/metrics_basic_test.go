//go:build ignore
// +build ignore

package metrics

import (
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsBasicFunctionality(t *testing.T) {
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry

	t.Run("测试应用指标", func(t *testing.T) {
		// 测试记录请求
		RecordRequest("GET", "/api/users", "200", 0.05, 100, 500)

		// 测试记录错误
		RecordError("GET", "/api/users", "not_found")

		// 测试记录数据库查询
		RecordDBQuery("SELECT", "users", "success", 0.01)

		assert.True(t, true) // 如果没有panic就算成功
	})
}

func TestPerformanceMonitorBasic(t *testing.T) {
	// 直接使用 MonitorService 而不是不存在的 PerformanceMonitor
	monitor := NewMonitorService(DefaultMonitorConfig)
	require.NotNil(t, monitor)

	// 测试获取状态
	status := monitor.GetStatus()
	assert.NotNil(t, status)
}

func TestMonitorServiceBasic(t *testing.T) {
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry

	config := DefaultMonitorConfig
	service := NewMonitorService(config)
	require.NotNil(t, service)
	assert.Equal(t, config, service.config)

	// 测试启动服务
	err := service.Start()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond) // 让服务运行一会儿

	// 测试获取状态
	status := service.GetStatus()
	assert.NotNil(t, status)

	// 测试获取告警
	alerts := service.GetAlerts()
	assert.NotNil(t, alerts)

	// 测试强制GC
	service.ForceGC()

	// 停止服务
	service.Stop()
}

func TestGlobalMonitorService(t *testing.T) {
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry

	// 重置全局服务
	defaultMonitorService = nil

	// 测试初始化默认监控服务
	err := InitDefaultMonitorService(DefaultMonitorConfig)
	require.NoError(t, err)

	// 获取默认监控服务
	service := GetDefaultMonitorService()
	require.NotNil(t, service)

	// 停止服务
	service.Stop()
}

func TestConcurrencyService(t *testing.T) {
	cs := NewConcurrencyService(10)

	assert.Equal(t, 0, cs.GetCurrent())
	assert.Equal(t, 10, cs.GetMax())

	// 测试获取许可
	for i := 0; i < 10; i++ {
		assert.True(t, cs.Acquire())
		assert.Equal(t, i+1, cs.GetCurrent())
	}

	// 达到最大值
	assert.False(t, cs.Acquire())
	assert.Equal(t, 10, cs.GetCurrent())

	// 测试释放许可
	for i := 0; i < 5; i++ {
		cs.Release()
		assert.Equal(t, 9-i, cs.GetCurrent())
	}

	// 可以再次获取
	assert.True(t, cs.Acquire())
}

func TestAlertOperations(t *testing.T) {
	service := NewMonitorService(DefaultMonitorConfig)

	// 测试创建告警（通过内部方法）
	service.createAlert("test", "test message", "warning")

	alerts := service.GetAlerts()
	assert.Len(t, alerts, 1)
	assert.Equal(t, "test", alerts[0].Type)
	assert.False(t, alerts[0].Resolved)

	// 测试解决告警
	resolved := service.ResolveAlert(alerts[0].ID)
	assert.True(t, resolved)
	assert.True(t, alerts[0].Resolved)

	// 再次解决应返回false
	resolved = service.ResolveAlert(alerts[0].ID)
	assert.False(t, resolved)
}

func TestGetMetrics(t *testing.T) {
	// 测试获取请求指标
	reqMetrics := GetRequestMetrics()
	assert.NotNil(t, reqMetrics)
	assert.Contains(t, reqMetrics, "total_requests")
	assert.Contains(t, reqMetrics, "error_count")
	assert.Contains(t, reqMetrics, "active_connections")

	// 测试获取系统指标
	sysMetrics := GetSystemMetrics()
	assert.NotNil(t, sysMetrics)
	assert.Contains(t, sysMetrics, "cpu_usage")
	assert.Contains(t, sysMetrics, "memory_usage")
	assert.Contains(t, sysMetrics, "disk_usage")

	// 测试获取业务指标
	bizMetrics := GetBusinessMetrics()
	assert.NotNil(t, bizMetrics)
	assert.Contains(t, bizMetrics, "user_sessions")
	assert.Contains(t, bizMetrics, "total_users")
	assert.Contains(t, bizMetrics, "business_operations")
}

func TestMetricRecording(t *testing.T) {
	t.Run("测试缓存指标记录", func(t *testing.T) {
		RecordCacheHit("redis", "get")
		RecordCacheMiss("redis", "get")
		RecordCacheEviction()

		assert.True(t, true) // 如果没有panic就算成功
	})

	t.Run("测试安全指标记录", func(t *testing.T) {
		RecordAuthAttempt("success", "admin", "local")
		RecordAuthFailure("invalid_password", "admin", "local")
		RecordSecurityEvent("login", "info", "admin")

		assert.True(t, true)
	})

	t.Run("测试业务指标记录", func(t *testing.T) {
		RecordBusinessOperation("create", "case", "success", "lawyer", 0.1)
		RecordCaseCreated("civil", "senior", "high")
		RecordDocumentProcessed("pdf", "success")
		RecordSearchQuery("fulltext", "10", "lawyer", "cases", 0.05)

		assert.True(t, true)
	})

	t.Run("测试数据库指标记录", func(t *testing.T) {
		SetDBConnections(10, 5)
		RecordDBTransaction("success", 0.01)

		assert.True(t, true)
	})

	t.Run("测试系统指标记录", func(t *testing.T) {
		SetActiveConnections(100)
		SetUserSessions(50)
		SetTotalUsers(1000)
		SetCacheSize(1024000)

		assert.True(t, true)
	})
}

func TestEnhancedMetricsCollector(t *testing.T) {
	collector := NewEnhancedMetricsCollector()
	assert.NotNil(t, collector)

	handler := collector.GetHandler()
	assert.NotNil(t, handler)
}

func TestGetPrometheusHandlers(t *testing.T) {
	handler1 := GetPrometheusHandler()
	assert.NotNil(t, handler1)

	handler2 := GetEnhancedPrometheusHandler()
	assert.NotNil(t, handler2)
}
