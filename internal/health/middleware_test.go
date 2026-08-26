package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthMiddleware_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	assert.NotNil(t, middleware)
	assert.Equal(t, "1.0.0", middleware.version)
	assert.Equal(t, "test", middleware.environment)
	assert.NotNil(t, middleware.startTime)
}

func TestHealthMiddleware_HealthCheckHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 创建路由
	router := gin.New()
	router.GET("/health", middleware.HealthCheckHandler)

	// 测试健康检查
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, "1.0.0", response["version"])
	assert.Contains(t, response, "uptime")
}

func TestHealthMiddleware_DetailedHealthCheckHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 添加一个模拟检查
	mockCheck := &MockHealthCheck{
		name:    "test_check",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "test_check",
			Status:    StatusHealthy,
			Duration:  50,
			Timestamp: time.Now(),
		},
	}
	healthChecker.RegisterCheck(mockCheck)

	// 创建路由
	router := gin.New()
	router.GET("/health/detailed", middleware.DetailedHealthCheckHandler)

	// 测试详细健康检查
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "check_results")
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "test", response["environment"])
	assert.Contains(t, response, "uptime")
}

func TestHealthMiddleware_HealthCheckMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 添加多个模拟检查
	mockCheck1 := &MockHealthCheck{
		name:    "check1",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "check1",
			Status:    StatusHealthy,
			Duration:  50,
			Timestamp: time.Now(),
		},
	}

	mockCheck2 := &MockHealthCheck{
		name:    "check2",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "check2",
			Status:    StatusDegraded,
			Duration:  100,
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck1)
	healthChecker.RegisterCheck(mockCheck2)

	// 创建路由
	router := gin.New()
	router.GET("/health/metrics", middleware.HealthCheckMetricsHandler)

	// 测试健康检查指标
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, float64(2), response["total_checks"])
	assert.Equal(t, float64(1), response["healthy_checks"])
	assert.Equal(t, float64(1), response["degraded_checks"])
	assert.Equal(t, float64(0), response["unhealthy_checks"])
	assert.Contains(t, response, "average_response_time")
}

func TestHealthMiddleware_ReadinessHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 创建路由
	router := gin.New()
	router.GET("/ready", middleware.ReadinessHandler)

	// 测试就绪状态
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["ready"])
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, "1.0.0", response["version"])
}

func TestHealthMiddleware_ReadinessRejectsDegradedChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	healthChecker.RegisterCheck(&MockHealthCheck{
		name:    "degraded_dependency",
		timeout: time.Second,
		result: &HealthCheckResult{
			Name:      "degraded_dependency",
			Status:    StatusDegraded,
			Message:   "dependency is degraded",
			Timestamp: time.Now(),
		},
	})
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "production")
	router := gin.New()
	router.GET("/ready", middleware.ReadinessHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, false, response["ready"])
}

func TestHealthMiddleware_ReadinessQAMarksProductionEvidenceSkipped(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	healthChecker.RegisterCheck(&MockHealthCheck{
		name:    ConflictP0ReadinessCheckName,
		timeout: time.Second,
		result: &HealthCheckResult{
			Name:      ConflictP0ReadinessCheckName,
			Status:    StatusHealthy,
			Message:   "QA技术就绪：生产证据门禁未执行；ready=true不代表生产可用",
			Details:   map[string]interface{}{"production_evidence_ready": false, "next_actions": productionEvidenceGateActions},
			Timestamp: time.Now(),
		},
	})
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "qa")
	router := gin.New()
	router.GET("/ready", middleware.ReadinessHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, true, response["ready"])
	assert.Equal(t, "qa", response["environment"])

	scope, ok := response["readiness_scope"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "environment_technical_readiness", scope["scope"])
	assert.Equal(t, false, scope["production_ready"])
	assert.Equal(t, false, scope["production_evidence_evaluated"])
	assert.Contains(t, scope["skipped_gates"], "G0:完成并签署 PD-01 至 PD-07 决策物，冻结生产适用规则")
	assert.Contains(t, scope["next_actions"], "保持 QA 流量边界；生产放行前在 production 环境逐项完成 G0-G7 证据门禁")

	checks, ok := response["checks"].(map[string]interface{})
	require.True(t, ok)
	conflictGate, ok := checks[ConflictP0ReadinessCheckName].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "healthy", conflictGate["status"])
	assert.Contains(t, conflictGate["message"], "不代表生产可用")
}

func TestHealthMiddleware_ReadinessProductionMissingEvidenceNotReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	healthChecker.RegisterCheck(&MockHealthCheck{
		name:    ConflictP0ReadinessCheckName,
		timeout: time.Second,
		result: &HealthCheckResult{
			Name:      ConflictP0ReadinessCheckName,
			Status:    StatusHealthy,
			Message:   "QA技术就绪：生产证据门禁未执行；ready=true不代表生产可用",
			Details:   map[string]interface{}{"production_evidence_ready": false, "next_actions": productionEvidenceGateActions},
			Timestamp: time.Now(),
		},
	})
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "production")
	router := gin.New()
	router.GET("/ready", middleware.ReadinessHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, false, response["ready"])

	scope, ok := response["readiness_scope"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "production_technical_and_governance_readiness", scope["scope"])
	assert.Equal(t, false, scope["production_ready"])
	assert.Equal(t, true, scope["production_evidence_evaluated"])
	assert.ElementsMatch(t, scope["missing"], []interface{}{"QA技术就绪：生产证据门禁未执行；ready=true不代表生产可用"})
	assert.Contains(t, scope["next_actions"], "G7:重新采集 /health/ready 与四类索引运行、政策版本、签署记录")
}

func TestHealthMiddleware_ReadinessProductionRequiresEvidenceCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "production")
	router := gin.New()
	router.GET("/ready", middleware.ReadinessHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, false, response["ready"])

	scope, ok := response["readiness_scope"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, scope["production_ready"])
	assert.ElementsMatch(t, scope["missing"], []interface{}{"缺少 conflict_p0_readiness 检查：请确认生产数据库连接和健康检查初始化成功"})
}

func TestHealthMiddleware_ReadinessProductionCompleteEvidenceReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	healthChecker.RegisterCheck(&MockHealthCheck{
		name:    ConflictP0ReadinessCheckName,
		timeout: time.Second,
		result: &HealthCheckResult{
			Name:      ConflictP0ReadinessCheckName,
			Status:    StatusHealthy,
			Message:   "生产证据门禁通过：冲突档案覆盖完整",
			Details:   map[string]interface{}{"production_evidence_ready": true, "next_actions": productionEvidenceGateActions},
			Timestamp: time.Now(),
		},
	})
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "production")
	router := gin.New()
	router.GET("/ready", middleware.ReadinessHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ready", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, true, response["ready"])

	scope, ok := response["readiness_scope"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "production_technical_and_governance_readiness", scope["scope"])
	assert.Equal(t, true, scope["production_ready"])
	assert.Equal(t, true, scope["production_evidence_evaluated"])
	assert.Empty(t, scope["missing"])
}

func TestHealthMiddleware_LivenessHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 创建路由
	router := gin.New()
	router.GET("/live", middleware.LivenessHandler)

	// 测试存活状态
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/live", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, true, response["live"])
	assert.Contains(t, response, "timestamp")
	assert.Equal(t, "1.0.0", response["version"])
	assert.Contains(t, response, "uptime")
}

func TestHealthMiddleware_DependencyHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 添加模拟检查
	mockCheck := &MockHealthCheck{
		name:    "database",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "database",
			Status:    StatusHealthy,
			Duration:  25,
			Message:   "连接正常",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"version": "3.37.0",
			},
		},
	}

	healthChecker.RegisterCheck(mockCheck)

	// 创建路由
	router := gin.New()
	router.GET("/health/dependencies", middleware.DependencyHealthHandler)

	// 测试依赖健康状态
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/dependencies", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "dependencies")
	assert.Equal(t, true, response["all_healthy"])

	dependencies := response["dependencies"].(map[string]interface{})
	database := dependencies["database"].(map[string]interface{})
	assert.Equal(t, "database", database["name"])
	assert.Equal(t, "healthy", database["status"])
	assert.Equal(t, float64(25), database["response_time"])
}

func TestHealthMiddleware_GracefulShutdownHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 创建路由
	router := gin.New()
	router.POST("/shutdown", middleware.GracefulShutdownHandler)

	// 测试有效令牌
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/shutdown", nil)
	req.Body = http.NoBody
	router.ServeHTTP(w, req)

	// 注意：在实际测试中，我们需要设置正确的请求体
	// 这里简化测试，主要验证路由存在
	assert.Equal(t, http.StatusBadRequest, w.Code) // 应该是400，因为没有正确的请求体
}

func TestHealthMiddleware_HealthStatusPageHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 添加模拟检查
	mockCheck1 := &MockHealthCheck{
		name:    "database",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "database",
			Status:    StatusHealthy,
			Duration:  25,
			Timestamp: time.Now(),
		},
	}

	mockCheck2 := &MockHealthCheck{
		name:    "cache",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "cache",
			Status:    StatusDegraded,
			Duration:  40,
			Message:   "响应缓慢",
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck1)
	healthChecker.RegisterCheck(mockCheck2)

	// 创建路由
	router := gin.New()
	router.GET("/health/status", middleware.HealthStatusPageHandler)

	// 测试状态页面
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "系统健康状态监控")
	assert.Contains(t, w.Body.String(), "总体状态")
	assert.Contains(t, w.Body.String(), "检查详情")
}

func TestHealthMiddleware_ExportHealthMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 添加模拟检查
	mockCheck := &MockHealthCheck{
		name:    "database",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "database",
			Status:    StatusHealthy,
			Duration:  25,
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck)

	// 创建路由
	router := gin.New()
	router.GET("/metrics/health", middleware.ExportHealthMetricsHandler)

	// 测试指标导出
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "health_check_status")
	assert.Contains(t, w.Body.String(), "health_check_response_time_ms")
	assert.Contains(t, w.Body.String(), "health_check_up")
	assert.Contains(t, w.Body.String(), "name=\"database\"")
}

func TestHealthMiddleware_HealthCheckMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 创建路由
	router := gin.New()

	// 添加中间件
	router.Use(middleware.HealthCheckMiddleware())

	// 添加端点
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// 测试非健康检查端点
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIsHealthCheckEndpoint(t *testing.T) {
	// 测试健康检查端点识别
	assert.True(t, isHealthCheckEndpoint("/health"))
	assert.True(t, isHealthCheckEndpoint("/api/v1/health"))
	assert.True(t, isHealthCheckEndpoint("/api/v1/health/detailed"))
	assert.True(t, isHealthCheckEndpoint("/api/v1/health/metrics"))
	assert.True(t, isHealthCheckEndpoint("/api/v1/health/history"))

	// 测试非健康检查端点
	assert.False(t, isHealthCheckEndpoint("/api/users"))
	assert.False(t, isHealthCheckEndpoint("/health/custom"))
	assert.False(t, isHealthCheckEndpoint("/api/v2/health"))
}

func TestHealthMiddleware_UnhealthySystem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 添加不健康的检查
	mockCheck := &MockHealthCheck{
		name:    "critical_service",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "critical_service",
			Status:    StatusUnhealthy,
			Duration:  100,
			Message:   "服务不可用",
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck)

	// 创建路由
	router := gin.New()
	router.GET("/health", middleware.HealthCheckHandler)

	// 测试不健康系统的响应
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// 应该返回服务不可用状态
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEqual(t, "healthy", response["status"])
}

func TestHealthMiddleware_WithTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	healthChecker := NewHealthChecker(&DefaultHealthConfig, nil)
	middleware := NewHealthMiddleware(healthChecker, "1.0.0", "test")

	// 添加超时的检查
	mockCheck := &MockHealthCheck{
		name:    "timeout_check",
		timeout: 1 * time.Millisecond,
		result: &HealthCheckResult{
			Name:      "timeout_check",
			Status:    StatusHealthy,
			Duration:  1000, // 模拟长延迟
			Timestamp: time.Now(),
		},
	}

	healthChecker.RegisterCheck(mockCheck)

	// 创建路由
	router := gin.New()
	router.GET("/health/detailed", middleware.DetailedHealthCheckHandler)

	// 测试超时处理
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health/detailed", nil)
	router.ServeHTTP(w, req)

	// 即使有超时，也应该返回响应
	assert.Equal(t, http.StatusOK, w.Code)
}
