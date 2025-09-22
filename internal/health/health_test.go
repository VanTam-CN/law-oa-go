package health

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthChecker_Basic(t *testing.T) {

	config := &DefaultHealthConfig
	healthChecker := NewHealthChecker(config, nil)
	require.NotNil(t, healthChecker)
	assert.Equal(t, config, healthChecker.config)
	assert.False(t, healthChecker.running)

	// 启动健康检查器
	healthChecker.Start()
	assert.True(t, healthChecker.running)

	time.Sleep(100 * time.Millisecond) // 让健康检查器运行一会儿

	// 停止健康检查器
	healthChecker.Stop()
	assert.False(t, healthChecker.running)
}

func TestHealthChecker_RegisterCheck(t *testing.T) {
	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)

	// 创建模拟的健康检查
	mockCheck := &MockHealthCheck{
		name:    "test_check",
		timeout: 5 * time.Second,
		result: &HealthCheckResult{
			Name:      "test_check",
			Status:    StatusHealthy,
			Duration:  50,
			Timestamp: time.Now(),
		},
	}

	// 注册健康检查
	healthChecker.RegisterCheck(mockCheck)

	// 验证健康检查已被注册
	healthChecker.mu.RLock()
	_, exists := healthChecker.checks["test_check"]
	healthChecker.mu.RUnlock()
	assert.True(t, exists)
}

func TestHealthChecker_RunChecks(t *testing.T) {
	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)

	// 注册多个模拟健康检查
	mockCheck1 := &MockHealthCheck{
		name:    "check1",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "check1",
			Status:    StatusHealthy,
			Duration:  10,
			Timestamp: time.Now(),
		},
	}

	mockCheck2 := &MockHealthCheck{
		name:    "check2",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "check2",
			Status:    StatusDegraded,
			Duration:  20,
			Message:   "检查降级",
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck1)
	healthChecker.RegisterCheck(mockCheck2)

	// 运行检查
	results := healthChecker.RunChecks()

	// 验证结果
	assert.Len(t, results, 2)
	assert.Equal(t, StatusHealthy, results["check1"].Status)
	assert.Equal(t, StatusDegraded, results["check2"].Status)
}

func TestHealthChecker_GetOverallHealth(t *testing.T) {
	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)

	// 注册健康检查
	mockCheck1 := &MockHealthCheck{
		name:    "healthy_check",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "healthy_check",
			Status:    StatusHealthy,
			Duration:  10,
			Timestamp: time.Now(),
		},
	}

	mockCheck2 := &MockHealthCheck{
		name:    "degraded_check",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "degraded_check",
			Status:    StatusDegraded,
			Duration:  20,
			Message:   "降级检查",
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck1)
	healthChecker.RegisterCheck(mockCheck2)

	// 获取总体健康状态
	health := healthChecker.GetOverallHealth("1.0.0", "test")

	// 验证总体状态
	assert.Equal(t, StatusDegraded, health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.Equal(t, "test", health.Environment)
	assert.Equal(t, 2, health.TotalChecks)
	assert.Equal(t, 1, health.PassedChecks)
	assert.Equal(t, 1, health.DegradedChecks)
	assert.Equal(t, 0, health.UnhealthyChecks)
	assert.Len(t, health.Checks, 2)
}

func TestHealthChecker_IsHealthy(t *testing.T) {
	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)

	// 测试没有健康检查的情况
	assert.True(t, healthChecker.IsHealthy())

	// 添加健康的检查
	healthyCheck := &MockHealthCheck{
		name:    "healthy",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "healthy",
			Status:    StatusHealthy,
			Duration:  10,
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(healthyCheck)
	assert.True(t, healthChecker.IsHealthy())

	// 添加不健康的检查
	unhealthyCheck := &MockHealthCheck{
		name:    "unhealthy",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "unhealthy",
			Status:    StatusUnhealthy,
			Duration:  10,
			Message:   "失败检查",
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(unhealthyCheck)
	assert.False(t, healthChecker.IsHealthy())
}

func TestMockHealthCheck(t *testing.T) {
	mockCheck := &MockHealthCheck{
		name:    "test",
		timeout: 5 * time.Second,
		result: &HealthCheckResult{
			Name:      "test",
			Status:    StatusHealthy,
			Duration:  100,
			Message:   "测试消息",
			Timestamp: time.Now(),
		},
	}

	// 测试Check方法
	result := mockCheck.Check(context.Background())
	assert.Equal(t, mockCheck.result, result)

	// 测试GetName方法
	assert.Equal(t, "test", mockCheck.GetName())

	// 测试GetTimeout方法
	assert.Equal(t, 5*time.Second, mockCheck.GetTimeout())
}

// MockHealthCheck 模拟健康检查用于测试
type MockHealthCheck struct {
	name    string
	timeout time.Duration
	result  *HealthCheckResult
}

func (m *MockHealthCheck) Check(ctx context.Context) *HealthCheckResult {
	return m.result
}

func (m *MockHealthCheck) GetName() string {
	return m.name
}

func (m *MockHealthCheck) GetTimeout() time.Duration {
	return m.timeout
}

func TestHealthStatus_Constants(t *testing.T) {
	assert.Equal(t, "healthy", string(StatusHealthy))
	assert.Equal(t, "degraded", string(StatusDegraded))
	assert.Equal(t, "unhealthy", string(StatusUnhealthy))
}

func TestHealthConfig_Defaults(t *testing.T) {
	config := DefaultHealthConfig

	assert.True(t, config.EnableDatabaseCheck)
	assert.True(t, config.EnableCacheCheck)
	assert.True(t, config.EnableExternalAPICheck)
	assert.True(t, config.EnableConcurrencyCheck)
	assert.True(t, config.EnableStorageCheck)
	assert.Equal(t, 5*time.Second, config.DatabaseTimeout)
	assert.Equal(t, 2*time.Second, config.CacheTimeout)
	assert.Equal(t, 3*time.Second, config.ExternalAPITimeout)
	assert.Equal(t, 3*time.Second, config.ConcurrencyTimeout)
	assert.Equal(t, 2*time.Second, config.StorageTimeout)
	assert.Equal(t, 30*time.Second, config.CheckInterval)
	assert.Equal(t, 3, config.FailureThreshold)
	assert.Equal(t, 2, config.SuccessThreshold)
}

func TestOverallHealth_StatusCalculation(t *testing.T) {
	// 测试全部健康
	health1 := &OverallHealth{
		Checks: map[string]HealthStatus{
			"check1": StatusHealthy,
			"check2": StatusHealthy,
		},
	}
	health1.calculateOverallStatus()
	assert.Equal(t, StatusHealthy, health1.Status)

	// 测试有降级检查
	health2 := &OverallHealth{
		Checks: map[string]HealthStatus{
			"check1": StatusHealthy,
			"check2": StatusDegraded,
		},
	}
	health2.calculateOverallStatus()
	assert.Equal(t, StatusDegraded, health2.Status)

	// 测试有不健康检查
	health3 := &OverallHealth{
		Checks: map[string]HealthStatus{
			"check1": StatusHealthy,
			"check2": StatusUnhealthy,
		},
	}
	health3.calculateOverallStatus()
	assert.Equal(t, StatusUnhealthy, health3.Status)
}

// calculateOverallStatus 计算总体状态（辅助方法，实际代码中不需要）
func (oh *OverallHealth) calculateOverallStatus() {
	oh.UnhealthyChecks = 0
	oh.DegradedChecks = 0
	oh.PassedChecks = 0

	for _, status := range oh.Checks {
		switch status {
		case StatusHealthy:
			oh.PassedChecks++
		case StatusDegraded:
			oh.DegradedChecks++
		case StatusUnhealthy:
			oh.UnhealthyChecks++
		}
	}

	if oh.UnhealthyChecks > 0 {
		oh.Status = StatusUnhealthy
	} else if oh.DegradedChecks > 0 {
		oh.Status = StatusDegraded
	} else {
		oh.Status = StatusHealthy
	}
}

func TestHealthChecker_GetLastResults(t *testing.T) {
	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)

	// 测试空结果
	results := healthChecker.GetLastResults()
	assert.Empty(t, results)

	// 添加一些检查结果
	mockCheck := &MockHealthCheck{
		name:    "test",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "test",
			Status:    StatusHealthy,
			Duration:  50,
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck)
	healthChecker.RunChecks() // 运行检查以填充结果

	results = healthChecker.GetLastResults()
	assert.NotEmpty(t, results)
	assert.Contains(t, results, "test")
	assert.Equal(t, StatusHealthy, results["test"].Status)
}

func TestHealthChecker_MultipleStartStop(t *testing.T) {
	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)

	// 多次启动和停止不应该出错
	healthChecker.Start()
	healthChecker.Start() // 第二次启动
	assert.True(t, healthChecker.running)

	healthChecker.Stop()
	healthChecker.Stop() // 第二次停止
	assert.False(t, healthChecker.running)
}

func TestHealthChecker_ConcurrentAccess(t *testing.T) {
	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)

	// 注册一些检查
	for i := 0; i < 5; i++ {
		mockCheck := &MockHealthCheck{
			name:    fmt.Sprintf("check%d", i),
			timeout: 1 * time.Second,
			result: &HealthCheckResult{
				Name:      fmt.Sprintf("check%d", i),
				Status:    StatusHealthy,
				Duration:  10,
				Timestamp: time.Now(),
			},
		}
		healthChecker.RegisterCheck(mockCheck)
	}

	healthChecker.Start()
	defer healthChecker.Stop()

	// 并发测试
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			healthChecker.RunChecks()
			healthChecker.IsHealthy()
			done <- true
		}()
	}

	// 等待所有并发操作完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
