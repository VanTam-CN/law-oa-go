package performance

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"law-oa-go/internal/middleware"
	"law-oa-go/internal/monitoring"
)

// TestPerformanceMonitoring 性能监控功能测试
func TestPerformanceMonitoring(t *testing.T) {
	// 初始化性能监控
	monitoring.InitPerformanceMetrics()
	metrics := monitoring.GetPerformanceMetrics()
	require.NotNil(t, metrics)

	// 测试指标记录
	metrics.RecordHTTPRequest("GET", "/test", "200", 100*time.Millisecond, 1024, 2048)
	metrics.RecordDBQuery("SELECT", "users", "success", 50*time.Millisecond)
	metrics.RecordRedisOperation("GET", "success", 5*time.Millisecond)

	// 测试业务指标
	metrics.IncrementCasesCreated()
	metrics.IncrementClientsRegistered()
	metrics.IncrementUsersLoggedIn()

	// 获取指标摘要
	summary := metrics.GetMetricsSummary()
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "system")
	assert.Contains(t, summary, "runtime")
}

// TestCacheEfficiency 缓存效率测试
func TestCacheEfficiency(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { require.NoError(t, rdb.Close()) })
	ctx := context.Background()
	for _, key := range []string{"test:user:one", "test:dashboard:one", "test:other", "another:user:one"} {
		require.NoError(t, rdb.Set(ctx, key, "value", 0).Err())
	}
	middleware.InvalidateCache(rdb, "test", "user:", "dashboard:")
	require.ElementsMatch(t, []string{"test:other", "another:user:one"}, server.Keys())
}

// TestMemoryUsage 内存使用测试
func TestMemoryUsage(t *testing.T) {
	monitoring.InitPerformanceMetrics()
	metrics := monitoring.GetPerformanceMetrics()

	// 记录初始内存使用
	summary1 := metrics.GetMetricsSummary()
	initialGoroutines := summary1["system"].(map[string]interface{})["goroutines"].(int)

	// 创建大量goroutines
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(100 * time.Millisecond)
			done <- true
		}()
	}

	// 等待所有goroutines完成
	for i := 0; i < 100; i++ {
		<-done
	}

	// 等待一段时间让goroutines清理
	time.Sleep(200 * time.Millisecond)

	// 检查goroutines数量是否恢复正常
	summary2 := metrics.GetMetricsSummary()
	finalGoroutines := summary2["system"].(map[string]interface{})["goroutines"].(int)

	// 允许一些goroutines的创建开销
	assert.True(t, finalGoroutines <= initialGoroutines+10,
		"Goroutine count should return to normal level")
}
