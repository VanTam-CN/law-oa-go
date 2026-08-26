package health

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthMiddleware 健康检查中间件
type HealthMiddleware struct {
	healthChecker *HealthChecker
	version       string
	environment   string
	startTime     time.Time
}

const productionEnvironmentName = "production"

// NewHealthMiddleware 创建健康检查中间件
func NewHealthMiddleware(healthChecker *HealthChecker, version, environment string) *HealthMiddleware {
	return &HealthMiddleware{
		healthChecker: healthChecker,
		version:       version,
		environment:   environment,
		startTime:     time.Now(),
	}
}

func isProductionEnvironment(environment string) bool {
	return strings.EqualFold(strings.TrimSpace(environment), productionEnvironmentName)
}

func productionEvidenceMissing(result *HealthCheckResult) []string {
	if result == nil {
		return []string{"缺少 conflict_p0_readiness 检查：请确认生产数据库连接和健康检查初始化成功"}
	}
	if result.Status != StatusHealthy {
		return []string{result.Message}
	}
	details, ok := result.Details.(map[string]interface{})
	if !ok {
		return []string{"conflict_p0_readiness 缺少生产证据声明：请检查健康检查实现"}
	}
	ready, ok := details["production_evidence_ready"].(bool)
	if !ok || !ready {
		return []string{result.Message}
	}
	return nil
}

// HealthCheckHandler 健康检查处理器
func (hm *HealthMiddleware) HealthCheckHandler(c *gin.Context) {
	// 快速健康检查
	if hm.healthChecker.IsHealthy() {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   hm.version,
			"uptime":    time.Since(hm.startTime).String(),
		})
		return
	}

	// 详细健康检查
	health := hm.healthChecker.GetOverallHealth(hm.version, hm.environment)
	health.Uptime = time.Since(hm.startTime).String()

	statusCode := http.StatusOK
	switch health.Status {
	case StatusDegraded:
		statusCode = http.StatusPartialContent
	case StatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, health)
}

// DetailedHealthCheckHandler 详细健康检查处理器
func (hm *HealthMiddleware) DetailedHealthCheckHandler(c *gin.Context) {
	results := hm.healthChecker.RunChecks()

	c.JSON(http.StatusOK, gin.H{
		"timestamp":     time.Now(),
		"version":       hm.version,
		"environment":   hm.environment,
		"uptime":        time.Since(hm.startTime).String(),
		"check_results": results,
	})
}

// HealthCheckMetricsHandler 健康检查指标处理器
func (hm *HealthMiddleware) HealthCheckMetricsHandler(c *gin.Context) {
	results := hm.healthChecker.RunChecks()

	metrics := gin.H{
		"timestamp":             time.Now(),
		"total_checks":          len(results),
		"healthy_checks":        0,
		"degraded_checks":       0,
		"unhealthy_checks":      0,
		"average_response_time": 0,
		"slowest_check":         "",
		"fastest_check":         "",
	}

	var totalTime int64
	slowestTime := int64(0)
	fastestTime := int64(999999999999)
	slowestName := ""
	fastestName := ""

	for name, result := range results {
		switch result.Status {
		case StatusHealthy:
			metrics["healthy_checks"] = metrics["healthy_checks"].(int) + 1
		case StatusDegraded:
			metrics["degraded_checks"] = metrics["degraded_checks"].(int) + 1
		case StatusUnhealthy:
			metrics["unhealthy_checks"] = metrics["unhealthy_checks"].(int) + 1
		}

		totalTime += result.Duration

		if result.Duration > slowestTime {
			slowestTime = result.Duration
			slowestName = name
		}

		if result.Duration < fastestTime {
			fastestTime = result.Duration
			fastestName = name
		}
	}

	if len(results) > 0 {
		metrics["average_response_time"] = totalTime / int64(len(results))
		metrics["slowest_check"] = fmt.Sprintf("%s (%dms)", slowestName, slowestTime)
		metrics["fastest_check"] = fmt.Sprintf("%s (%dms)", fastestName, fastestTime)
	}

	c.JSON(http.StatusOK, metrics)
}

// HealthCheckHistoryHandler 健康检查历史处理器
func (hm *HealthMiddleware) HealthCheckHistoryHandler(c *gin.Context) {
	// 这里可以从数据库或缓存中获取历史数据
	// 为了简化，返回当前状态
	history := gin.H{
		"current_status": hm.healthChecker.GetOverallHealth(hm.version, hm.environment),
		"history": []gin.H{
			{
				"timestamp":     time.Now().Add(-5 * time.Minute),
				"status":        "healthy",
				"response_time": 150,
			},
			{
				"timestamp":     time.Now().Add(-10 * time.Minute),
				"status":        "healthy",
				"response_time": 120,
			},
		},
	}

	c.JSON(http.StatusOK, history)
}

// HealthCheckMiddleware 健康检查中间件（用于路由）
func (hm *HealthMiddleware) HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只有健康检查端点才需要这个中间件
		if !isHealthCheckEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 快速检查系统健康状态
		if !hm.healthChecker.IsHealthy() {
			health := hm.healthChecker.GetOverallHealth(hm.version, hm.environment)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":         "Service unhealthy",
				"status":        health.Status,
				"failed_checks": health.FailedChecks,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isHealthCheckEndpoint 检查是否为健康检查端点
func isHealthCheckEndpoint(path string) bool {
	healthEndpoints := []string{
		"/health",
		"/api/v1/health",
		"/api/v1/health/detailed",
		"/api/v1/health/metrics",
		"/api/v1/health/history",
	}

	for _, endpoint := range healthEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}

// GracefulShutdownHandler 优雅关闭处理器
func (hm *HealthMiddleware) GracefulShutdownHandler(c *gin.Context) {
	type shutdownRequest struct {
		Reason  string `json:"reason"`
		Timeout int    `json:"timeout"`
		Token   string `json:"token"`
	}

	var req shutdownRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// The endpoint is development-only, but a source-level fallback token would
	// still be usable by any local process or copied deployment. Require an
	// operator-provided secret and prefer a header so it is not echoed in the
	// request body logs by default.
	providedToken := c.GetHeader("X-Shutdown-Token")
	if providedToken == "" {
		providedToken = req.Token
	}
	expectedToken := strings.TrimSpace(os.Getenv("LAW_OA_SHUTDOWN_TOKEN"))
	if expectedToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Shutdown token is not configured"})
		return
	}
	if len(providedToken) != len(expectedToken) || subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid shutdown token"})
		return
	}

	// 记录关闭请求
	c.JSON(http.StatusOK, gin.H{
		"message":   "Graceful shutdown initiated",
		"reason":    req.Reason,
		"timeout":   req.Timeout,
		"timestamp": time.Now(),
	})

	// 在实际应用中，这里会触发优雅关闭流程
	// 包括：停止接受新请求、完成现有请求、关闭数据库连接等

	// 为了演示，我们只是记录日志
	fmt.Printf("优雅关闭请求: 原因=%s, 超时=%d秒\n", req.Reason, req.Timeout)
}

// ReadinessHandler 就绪状态处理器
func (hm *HealthMiddleware) ReadinessHandler(c *gin.Context) {
	// Readiness is a traffic gate, so degraded dependencies are not enough.
	// Liveness remains separate: a process may stay alive while Kubernetes
	// removes it from service until every required dependency is healthy again.
	health := hm.healthChecker.GetOverallHealth(hm.version, hm.environment)
	ready := health.Status == StatusHealthy
	production := isProductionEnvironment(hm.environment)
	productionGate, productionGateExists := hm.healthChecker.GetLastResults()[ConflictP0ReadinessCheckName]
	missingProductionEvidence := productionEvidenceMissing(productionGate)

	// Readiness is environment-scoped. A QA deployment may receive technical
	// traffic while explicitly reporting that production evidence was not
	// evaluated. A production deployment must never interpret that omission
	// as success: the check and its positive evidence declaration are both
	// required before /health/ready can return true.
	if production {
		ready = ready && productionGateExists && len(missingProductionEvidence) == 0
	}

	evidence := gin.H{
		"scope":                         "environment_technical_readiness",
		"production_ready":              production && ready,
		"production_evidence_evaluated": production,
	}
	if production {
		evidence["scope"] = "production_technical_and_governance_readiness"
		evidence["missing"] = missingProductionEvidence
		evidence["next_actions"] = productionEvidenceGateActions
	} else {
		evidence["skipped_gates"] = productionEvidenceGateActions
		evidence["next_actions"] = []string{"保持 QA 流量边界；生产放行前在 production 环境逐项完成 G0-G7 证据门禁"}
	}

	status := gin.H{
		"ready":           ready,
		"timestamp":       time.Now(),
		"version":         hm.version,
		"environment":     hm.environment,
		"readiness_scope": evidence,
		"checks":          hm.healthChecker.GetLastResults(),
	}

	if ready {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

// LivenessHandler 存活状态处理器
func (hm *HealthMiddleware) LivenessHandler(c *gin.Context) {
	// 检查应用是否存活（更简单的检查）
	live := true // 应用正在运行

	status := gin.H{
		"live":      live,
		"timestamp": time.Now(),
		"version":   hm.version,
		"uptime":    time.Since(hm.startTime).String(),
	}

	if live {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusInternalServerError, status)
	}
}

// DependencyHealthHandler 依赖服务健康状态处理器
func (hm *HealthMiddleware) DependencyHealthHandler(c *gin.Context) {
	results := hm.healthChecker.RunChecks()

	dependencies := make(map[string]interface{})

	for name, result := range results {
		dependency := gin.H{
			"name":          name,
			"status":        result.Status,
			"response_time": result.Duration,
			"last_checked":  result.Timestamp,
		}

		if result.Message != "" {
			dependency["message"] = result.Message
		}

		if result.Details != nil {
			dependency["details"] = result.Details
		}

		dependencies[name] = dependency
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp":    time.Now(),
		"version":      hm.version,
		"dependencies": dependencies,
		"all_healthy":  hm.healthChecker.IsHealthy(),
	})
}

// HealthStatusPageHandler 健康状态页面处理器
func (hm *HealthMiddleware) HealthStatusPageHandler(c *gin.Context) {
	health := hm.healthChecker.GetOverallHealth(hm.version, hm.environment)
	health.Uptime = time.Since(hm.startTime).String()

	// 返回HTML格式的状态页面
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>系统健康状态</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .status { padding: 20px; border-radius: 5px; margin: 10px 0; }
        .healthy { background-color: #d4edda; color: #155724; }
        .degraded { background-color: #fff3cd; color: #856404; }
        .unhealthy { background-color: #f8d7da; color: #721c24; }
        .check-item { margin: 10px 0; padding: 10px; border-left: 4px solid #ddd; }
        .metric { display: inline-block; margin: 5px 10px; }
    </style>
</head>
<body>
    <h1>系统健康状态监控</h1>
    
    <div class="status %s">
        <h2>总体状态: %s</h2>
        <div class="metric">版本: %s</div>
        <div class="metric">环境: %s</div>
        <div class="metric">运行时间: %s</div>
        <div class="metric">检查耗时: %dms</div>
    </div>
    
    <h3>检查详情</h3>
`, health.Status, health.Status, health.Version, health.Environment, health.Uptime, health.CheckDuration)

	for name, status := range health.Checks {
		statusClass := "healthy"
		switch status {
		case StatusDegraded:
			statusClass = "degraded"
		case StatusUnhealthy:
			statusClass = "unhealthy"
		}

		html += fmt.Sprintf(`
    <div class="check-item %s">
        <strong>%s:</strong> %s
    </div>`, statusClass, name, status)
	}

	html += `
    <hr>
    <p><small>最后更新: ` + time.Now().Format("2006-01-02 15:04:05") + `</small></p>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ExportHealthMetricsHandler 导出健康指标处理器
func (hm *HealthMiddleware) ExportHealthMetricsHandler(c *gin.Context) {
	results := hm.healthChecker.RunChecks()

	// Prometheus 格式的指标
	metrics := "# TYPE health_check_status gauge\n"
	metrics += "# TYPE health_check_response_time_ms gauge\n"
	metrics += "# TYPE health_check_up gauge\n"

	for name, result := range results {
		// 健康状态指标 (1=健康, 0.5=降级, 0=不健康)
		statusValue := "1.0"
		if result.Status == StatusDegraded {
			statusValue = "0.5"
		} else if result.Status == StatusUnhealthy {
			statusValue = "0.0"
		}

		metrics += fmt.Sprintf(`health_check_status{name="%s"} %s`, name, statusValue)
		metrics += fmt.Sprintf(" %d\n", time.Now().Unix())

		// 响应时间指标
		metrics += fmt.Sprintf(`health_check_response_time_ms{name="%s"} %d`, name, result.Duration)
		metrics += fmt.Sprintf(" %d\n", time.Now().Unix())

		// 可达性指标
		upValue := "1.0"
		if result.Status == StatusUnhealthy {
			upValue = "0.0"
		}
		metrics += fmt.Sprintf(`health_check_up{name="%s"} %s`, name, upValue)
		metrics += fmt.Sprintf(" %d\n", time.Now().Unix())
	}

	c.Header("Content-Type", "text/plain; version=0.0.4")
	c.String(http.StatusOK, metrics)
}
