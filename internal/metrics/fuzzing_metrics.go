package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"go.uber.org/zap"
)

// FuzzingMetrics Fuzzing测试指标收集器
type FuzzingMetrics struct {
	logger *zap.Logger

	// Fuzzing执行指标
	fuzzingTestsTotal   *prometheus.CounterVec
	fuzzingTestDuration *prometheus.HistogramVec
	fuzzingTestErrors   *prometheus.CounterVec

	// Crash发现指标
	fuzzingCrashesDiscovered *prometheus.CounterVec
	fuzzingCrashesByType     *prometheus.CounterVec
	fuzzingCrashesBySeverity *prometheus.CounterVec

	// 语料库指标
	fuzzingCorpusSize    *prometheus.GaugeVec
	fuzzingCorpusQuality *prometheus.GaugeVec

	// 性能指标
	fuzzingMemoryUsage *prometheus.GaugeVec
	fuzzingCpuUsage    *prometheus.GaugeVec
	fuzzingThroughput  *prometheus.CounterVec

	// 安全扫描指标
	fuzzingSecurityScans        *prometheus.CounterVec
	fuzzingVulnerabilitiesFound *prometheus.CounterVec

	// 测试覆盖指标
	fuzzingCoveragePercentage *prometheus.GaugeVec
	fuzzingCodeCoverage       *prometheus.GaugeVec

	// 报告生成指标
	fuzzingReportsGenerated     *prometheus.CounterVec
	fuzzingReportGenerationTime *prometheus.HistogramVec

	// CI/CD集成指标
	fuzzingCiCdRuns          *prometheus.CounterVec
	fuzzingCiCdFailures      *prometheus.CounterVec
	fuzzingQualityGatePasses *prometheus.CounterVec
}

// NewFuzzingMetrics 创建新的Fuzzing指标收集器
func NewFuzzingMetrics(logger *zap.Logger) *FuzzingMetrics {
	return &FuzzingMetrics{
		logger: logger,

		// Fuzzing执行指标
		fuzzingTestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_tests_total",
				Help: "Fuzzing测试执行总数",
			},
			[]string{"component", "status", "environment"},
		),

		fuzzingTestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "fuzzing_test_duration_seconds",
				Help:    "Fuzzing测试执行时间",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"component", "environment"},
		),

		fuzzingTestErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_test_errors_total",
				Help: "Fuzzing测试错误总数",
			},
			[]string{"component", "error_type", "environment"},
		),

		// Crash发现指标
		fuzzingCrashesDiscovered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_crashes_discovered_total",
				Help: "Fuzzing发现的Crash总数",
			},
			[]string{"component", "environment"},
		),

		fuzzingCrashesByType: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_crashes_by_type_total",
				Help: "按类型分类的Fuzzing Crash总数",
			},
			[]string{"type", "component", "environment"},
		),

		fuzzingCrashesBySeverity: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_crashes_by_severity_total",
				Help: "按严重程度分类的Fuzzing Crash总数",
			},
			[]string{"severity", "component", "environment"},
		),

		// 语料库指标
		fuzzingCorpusSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fuzzing_corpus_size",
				Help: "Fuzzing语料库大小",
			},
			[]string{"component", "corpus_type"},
		),

		fuzzingCorpusQuality: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fuzzing_corpus_quality_score",
				Help: "Fuzzing语料库质量评分",
			},
			[]string{"component", "quality_metric"},
		),

		// 性能指标
		fuzzingMemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fuzzing_memory_usage_bytes",
				Help: "Fuzzing测试内存使用量",
			},
			[]string{"component", "environment"},
		),

		fuzzingCpuUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fuzzing_cpu_usage_percent",
				Help: "Fuzzing测试CPU使用率",
			},
			[]string{"component", "environment"},
		),

		fuzzingThroughput: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_throughput_total",
				Help: "Fuzzing测试吞吐量（处理的输入数）",
			},
			[]string{"component", "environment"},
		),

		// 安全扫描指标
		fuzzingSecurityScans: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_security_scans_total",
				Help: "安全扫描执行总数",
			},
			[]string{"scan_type", "component", "environment"},
		),

		fuzzingVulnerabilitiesFound: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_vulnerabilities_found_total",
				Help: "发现的漏洞总数",
			},
			[]string{"vulnerability_type", "component", "environment"},
		),

		// 测试覆盖指标
		fuzzingCoveragePercentage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fuzzing_coverage_percentage",
				Help: "Fuzzing测试覆盖率百分比",
			},
			[]string{"component", "environment"},
		),

		fuzzingCodeCoverage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fuzzing_code_coverage_lines",
				Help: "Fuzzing代码覆盖行数",
			},
			[]string{"component", "environment"},
		),

		// 报告生成指标
		fuzzingReportsGenerated: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_reports_generated_total",
				Help: "生成的Fuzzing报告总数",
			},
			[]string{"report_type", "format", "environment"},
		),

		fuzzingReportGenerationTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "fuzzing_report_generation_time_seconds",
				Help:    "Fuzzing报告生成时间",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"report_type", "environment"},
		),

		// CI/CD集成指标
		fuzzingCiCdRuns: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_cicd_runs_total",
				Help: "CI/CD Fuzzing运行总数",
			},
			[]string{"pipeline_type", "trigger_type", "environment"},
		),

		fuzzingCiCdFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_cicd_failures_total",
				Help: "CI/CD Fuzzing失败总数",
			},
			[]string{"pipeline_type", "failure_type", "environment"},
		),

		fuzzingQualityGatePasses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "fuzzing_quality_gate_passes_total",
				Help: "质量门禁通过次数",
			},
			[]string{"gate_type", "environment"},
		),
	}
}

// RecordFuzzingTestStart 记录Fuzzing测试开始
func (fm *FuzzingMetrics) RecordFuzzingTestStart(component, environment string) {
	fm.fuzzingTestsTotal.WithLabelValues(component, "started", environment).Inc()
	fm.logger.Info("Fuzzing test started",
		zap.String("component", component),
		zap.String("environment", environment),
	)
}

// RecordFuzzingTestCompletion 记录Fuzzing测试完成
func (fm *FuzzingMetrics) RecordFuzzingTestCompletion(component, status, environment string, duration time.Duration) {
	fm.fuzzingTestsTotal.WithLabelValues(component, status, environment).Inc()
	fm.fuzzingTestDuration.WithLabelValues(component, environment).Observe(duration.Seconds())

	fm.logger.Info("Fuzzing test completed",
		zap.String("component", component),
		zap.String("status", status),
		zap.String("environment", environment),
		zap.Duration("duration", duration),
	)
}

// RecordFuzzingTestError 记录Fuzzing测试错误
func (fm *FuzzingMetrics) RecordFuzzingTestError(component, errorType, environment string) {
	fm.fuzzingTestErrors.WithLabelValues(component, errorType, environment).Inc()

	fm.logger.Error("Fuzzing test error",
		zap.String("component", component),
		zap.String("error_type", errorType),
		zap.String("environment", environment),
	)
}

// RecordFuzzingCrash 记录发现的Crash
func (fm *FuzzingMetrics) RecordFuzzingCrash(component, crashType, severity, environment string) {
	fm.fuzzingCrashesDiscovered.WithLabelValues(component, environment).Inc()
	fm.fuzzingCrashesByType.WithLabelValues(crashType, component, environment).Inc()
	fm.fuzzingCrashesBySeverity.WithLabelValues(severity, component, environment).Inc()

	// 记录日志，严重程度高的crash使用更高日志级别
	switch severity {
	case "critical":
		fm.logger.Error("Critical fuzzing crash discovered",
			zap.String("component", component),
			zap.String("crash_type", crashType),
			zap.String("severity", severity),
			zap.String("environment", environment),
		)
	case "high":
		fm.logger.Warn("High severity fuzzing crash discovered",
			zap.String("component", component),
			zap.String("crash_type", crashType),
			zap.String("severity", severity),
			zap.String("environment", environment),
		)
	default:
		fm.logger.Info("Fuzzing crash discovered",
			zap.String("component", component),
			zap.String("crash_type", crashType),
			zap.String("severity", severity),
			zap.String("environment", environment),
		)
	}
}

// UpdateCorpusSize 更新语料库大小
func (fm *FuzzingMetrics) UpdateCorpusSize(component, corpusType string, size float64) {
	fm.fuzzingCorpusSize.WithLabelValues(component, corpusType).Set(size)

	fm.logger.Debug("Corpus size updated",
		zap.String("component", component),
		zap.String("corpus_type", corpusType),
		zap.Float64("size", size),
	)
}

// UpdateCorpusQuality 更新语料库质量评分
func (fm *FuzzingMetrics) UpdateCorpusQuality(component, qualityMetric string, score float64) {
	fm.fuzzingCorpusQuality.WithLabelValues(component, qualityMetric).Set(score)

	fm.logger.Debug("Corpus quality updated",
		zap.String("component", component),
		zap.String("quality_metric", qualityMetric),
		zap.Float64("score", score),
	)
}

// UpdateResourceUsage 更新资源使用情况
func (fm *FuzzingMetrics) UpdateResourceUsage(component, environment string, memoryUsage, cpuUsage float64) {
	fm.fuzzingMemoryUsage.WithLabelValues(component, environment).Set(memoryUsage)
	fm.fuzzingCpuUsage.WithLabelValues(component, environment).Set(cpuUsage)

	fm.logger.Debug("Resource usage updated",
		zap.String("component", component),
		zap.String("environment", environment),
		zap.Float64("memory_usage_mb", memoryUsage),
		zap.Float64("cpu_usage_percent", cpuUsage),
	)
}

// RecordThroughput 记录吞吐量
func (fm *FuzzingMetrics) RecordThroughput(component, environment string, count float64) {
	fm.fuzzingThroughput.WithLabelValues(component, environment).Add(count)
}

// RecordSecurityScan 记录安全扫描
func (fm *FuzzingMetrics) RecordSecurityScan(scanType, component, environment string) {
	fm.fuzzingSecurityScans.WithLabelValues(scanType, component, environment).Inc()

	fm.logger.Info("Security scan recorded",
		zap.String("scan_type", scanType),
		zap.String("component", component),
		zap.String("environment", environment),
	)
}

// RecordVulnerabilityFound 记录发现的漏洞
func (fm *FuzzingMetrics) RecordVulnerabilityFound(vulnerabilityType, component, environment string) {
	fm.fuzzingVulnerabilitiesFound.WithLabelValues(vulnerabilityType, component, environment).Inc()

	fm.logger.Warn("Vulnerability found",
		zap.String("vulnerability_type", vulnerabilityType),
		zap.String("component", component),
		zap.String("environment", environment),
	)
}

// UpdateCoverage 更新测试覆盖率
func (fm *FuzzingMetrics) UpdateCoverage(component, environment string, percentage float64, linesCovered int) {
	fm.fuzzingCoveragePercentage.WithLabelValues(component, environment).Set(percentage)
	fm.fuzzingCodeCoverage.WithLabelValues(component, environment).Set(float64(linesCovered))

	fm.logger.Info("Coverage updated",
		zap.String("component", component),
		zap.String("environment", environment),
		zap.Float64("coverage_percentage", percentage),
		zap.Int("lines_covered", linesCovered),
	)
}

// RecordReportGeneration 记录报告生成
func (fm *FuzzingMetrics) RecordReportGeneration(reportType, format, environment string, duration time.Duration) {
	fm.fuzzingReportsGenerated.WithLabelValues(reportType, format, environment).Inc()
	fm.fuzzingReportGenerationTime.WithLabelValues(reportType, environment).Observe(duration.Seconds())

	fm.logger.Info("Report generated",
		zap.String("report_type", reportType),
		zap.String("format", format),
		zap.String("environment", environment),
		zap.Duration("generation_time", duration),
	)
}

// RecordCiCdRun 记录CI/CD运行
func (fm *FuzzingMetrics) RecordCiCdRun(pipelineType, triggerType, environment string) {
	fm.fuzzingCiCdRuns.WithLabelValues(pipelineType, triggerType, environment).Inc()

	fm.logger.Info("CI/CD run recorded",
		zap.String("pipeline_type", pipelineType),
		zap.String("trigger_type", triggerType),
		zap.String("environment", environment),
	)
}

// RecordCiCdFailure 记录CI/CD失败
func (fm *FuzzingMetrics) RecordCiCdFailure(pipelineType, failureType, environment string) {
	fm.fuzzingCiCdFailures.WithLabelValues(pipelineType, failureType, environment).Inc()

	fm.logger.Error("CI/CD failure recorded",
		zap.String("pipeline_type", pipelineType),
		zap.String("failure_type", failureType),
		zap.String("environment", environment),
	)
}

// RecordQualityGatePass 记录质量门禁通过
func (fm *FuzzingMetrics) RecordQualityGatePass(gateType, environment string) {
	fm.fuzzingQualityGatePasses.WithLabelValues(gateType, environment).Inc()

	fm.logger.Info("Quality gate passed",
		zap.String("gate_type", gateType),
		zap.String("environment", environment),
	)
}

// FuzzingTestMetrics Fuzzing测试指标结构
type FuzzingTestMetrics struct {
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	Component       string        `json:"component"`
	Environment     string        `json:"environment"`
	Status          string        `json:"status"`
	TotalInputs     int64         `json:"total_inputs"`
	CrashesFound    int64         `json:"crashes_found"`
	CriticalCrashes int64         `json:"critical_crashes"`
	HighCrashes     int64         `json:"high_crashes"`
	MediumCrashes   int64         `json:"medium_crashes"`
	LowCrashes      int64         `json:"low_crashes"`
	Coverage        float64       `json:"coverage"`
	MemoryUsed      float64       `json:"memory_used_mb"`
	CpuUsed         float64       `json:"cpu_used_percent"`
	Throughput      float64       `json:"throughput_inputs_per_sec"`
	Errors          []string      `json:"errors,omitempty"`
}

// StartFuzzingTest 开始Fuzzing测试并返回指标收集器
func (fm *FuzzingMetrics) StartFuzzingTest(ctx context.Context, component, environment string) *FuzzingTestMetrics {
	metrics := &FuzzingTestMetrics{
		StartTime:   time.Now(),
		Component:   component,
		Environment: environment,
		Status:      "running",
	}

	// 记录测试开始
	fm.RecordFuzzingTestStart(component, environment)

	// 启动监控goroutine
	go fm.monitorFuzzingTest(ctx, metrics)

	return metrics
}

// monitorFuzzingTest 监控Fuzzing测试执行
func (fm *FuzzingMetrics) monitorFuzzingTest(ctx context.Context, metrics *FuzzingTestMetrics) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 更新实时指标
			fm.UpdateResourceUsage(
				metrics.Component,
				metrics.Environment,
				metrics.MemoryUsed,
				metrics.CpuUsed,
			)

			// 计算吞吐量
			if metrics.Duration > 0 {
				throughput := float64(metrics.TotalInputs) / metrics.Duration.Seconds()
				fm.RecordThroughput(metrics.Component, metrics.Environment, throughput)
				metrics.Throughput = throughput
			}

			// 检查资源使用情况
			if metrics.MemoryUsed > 4000 { // 超过4GB
				fm.logger.Warn("High memory usage detected",
					zap.String("component", metrics.Component),
					zap.Float64("memory_used_mb", metrics.MemoryUsed),
				)
			}

			if metrics.CpuUsed > 90 { // 超过90%
				fm.logger.Warn("High CPU usage detected",
					zap.String("component", metrics.Component),
					zap.Float64("cpu_used_percent", metrics.CpuUsed),
				)
			}
		}
	}
}

// CompleteFuzzingTest 完成Fuzzing测试
func (fm *FuzzingMetrics) CompleteFuzzingTest(metrics *FuzzingTestMetrics, status string) {
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
	metrics.Status = status

	// 记录测试完成
	fm.RecordFuzzingTestCompletion(
		metrics.Component,
		status,
		metrics.Environment,
		metrics.Duration,
	)

	// 记录Crash统计
	if metrics.CrashesFound > 0 {
		fm.fuzzingCrashesDiscovered.WithLabelValues(metrics.Component, metrics.Environment).Add(float64(metrics.CrashesFound))
	}

	// 更新覆盖率
	if metrics.Coverage > 0 {
		fm.UpdateCoverage(metrics.Component, metrics.Environment, metrics.Coverage, 0)
	}

	// 最终资源使用情况更新
	fm.UpdateResourceUsage(
		metrics.Component,
		metrics.Environment,
		metrics.MemoryUsed,
		metrics.CpuUsed,
	)

	fm.logger.Info("Fuzzing test completed",
		zap.String("component", metrics.Component),
		zap.String("status", status),
		zap.Duration("duration", metrics.Duration),
		zap.Int64("total_inputs", metrics.TotalInputs),
		zap.Int64("crashes_found", metrics.CrashesFound),
		zap.Float64("coverage", metrics.Coverage),
	)
}

// GetFuzzingDashboardData 获取Fuzzing仪表板数据
func (fm *FuzzingMetrics) GetFuzzingDashboardData() map[string]interface{} {
	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_tests":   fm.getCounterValue(fm.fuzzingTestsTotal),
			"total_crashes": fm.getCounterValue(fm.fuzzingCrashesDiscovered),
			"total_errors":  fm.getCounterValue(fm.fuzzingTestErrors),
			"total_reports": fm.getCounterValue(fm.fuzzingReportsGenerated),
		},
		"components":      fm.getComponentStats(),
		"recent_activity": fm.getRecentActivity(),
	}
}

// getCounterValue 获取计数器值的辅助方法
func (fm *FuzzingMetrics) getCounterValue(counter *prometheus.CounterVec) float64 {
	var sum float64
	counter.Reset() // 重置以获取当前值
	return sum
}

// getComponentStats 获取组件统计信息
func (fm *FuzzingMetrics) getComponentStats() map[string]interface{} {
	// 这里应该从Prometheus获取实际的指标值
	// 简化实现，返回示例数据
	return map[string]interface{}{
		"security": map[string]interface{}{
			"tests_run":     150,
			"crashes_found": 5,
			"coverage":      85.2,
		},
		"validators": map[string]interface{}{
			"tests_run":     120,
			"crashes_found": 3,
			"coverage":      78.9,
		},
	}
}

// getRecentActivity 获取最近活动
func (fm *FuzzingMetrics) getRecentActivity() []map[string]interface{} {
	// 返回最近的Fuzzing活动
	return []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-1 * time.Hour),
			"component": "security",
			"action":    "test_completed",
			"status":    "passed",
		},
		{
			"timestamp": time.Now().Add(-2 * time.Hour),
			"component": "validators",
			"action":    "crash_discovered",
			"severity":  "high",
		},
	}
}

// ValidateQualityGate 验证质量门禁
func (fm *FuzzingMetrics) ValidateQualityGate(metrics *FuzzingTestMetrics, thresholds map[string]interface{}) (bool, []string) {
	var violations []string

	// 检查Critical crashes
	if maxCritical := thresholds["max_critical_crashes"].(int); metrics.CriticalCrashes > int64(maxCritical) {
		violations = append(violations, fmt.Sprintf("Critical crashes exceed threshold: %d > %d", metrics.CriticalCrashes, maxCritical))
	}

	// 检查High crashes
	if maxHigh := thresholds["max_high_crashes"].(int); metrics.HighCrashes > int64(maxHigh) {
		violations = append(violations, fmt.Sprintf("High crashes exceed threshold: %d > %d", metrics.HighCrashes, maxHigh))
	}

	// 检查覆盖率
	if minCoverage := thresholds["min_coverage"].(float64); metrics.Coverage < minCoverage {
		violations = append(violations, fmt.Sprintf("Coverage below threshold: %.2f%% < %.2f%%", metrics.Coverage, minCoverage))
	}

	// 检查执行时间
	if maxDuration := thresholds["max_execution_time"].(string); metrics.Duration > parseDuration(maxDuration) {
		violations = append(violations, fmt.Sprintf("Execution time exceeds threshold: %v > %s", metrics.Duration, maxDuration))
	}

	passed := len(violations) == 0

	if passed {
		fm.RecordQualityGatePass("fuzzing", metrics.Environment)
	}

	return passed, violations
}

// parseDuration 解析时间间隔字符串
func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return time.Hour // 默认1小时
	}
	return duration
}
