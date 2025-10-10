package executors

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/testing"
)

// IntegrationExecutor 集成测试执行器
type IntegrationExecutor struct {
	*testing.BaseExecutor
	apiExecutor    *APIExecutor
	uiExecutor     *UIExecutor
	perfExecutor   *PerformanceExecutor
	testPhases     []TestPhase
	currentPhase   int
	phaseResults   map[string]*testing.TestResult
}

// TestPhase 测试阶段
type TestPhase struct {
	Name         string
	Type         testing.TestType
	Tests        []*testing.TestCase
	Setup        []testing.TestStep
	Teardown     []testing.TestStep
	Dependencies []string
	Parallel     bool
	Timeout      time.Duration
	Environment  map[string]interface{}
}

// IntegrationContext 集成测试上下文
type IntegrationContext struct {
	SharedData     map[string]interface{}
	Environment    map[string]string
	Variables      map[string]interface{}
	SessionTokens  map[string]string
	DatabaseState  map[string]interface{}
	MockServices   map[string]interface{}
	Configuration  map[string]interface{}
}

// NewIntegrationExecutor 创建集成测试执行器
func NewIntegrationExecutor(options *testing.TestExecutorOptions, logger testing.TestLogger, metrics testing.TestMetrics) *IntegrationExecutor {
	base := testing.NewBaseExecutor(options, logger, metrics)

	integrationExecutor := &IntegrationExecutor{
		BaseExecutor: base,
		testPhases:   make([]TestPhase, 0),
		phaseResults: make(map[string]*testing.TestResult),
	}

	// 创建子执行器
	integrationExecutor.apiExecutor = NewAPIExecutor(options, logger, metrics)
	integrationExecutor.uiExecutor = NewUIExecutor(options, logger, metrics)
	integrationExecutor.perfExecutor = NewPerformanceExecutor(options, logger, metrics)

	return integrationExecutor
}

// GetExecutorType 获取执行器类型
func (e *IntegrationExecutor) GetExecutorType() testing.TestType {
	return testing.TestTypeIntegration
}

// Setup 设置集成测试执行器
func (e *IntegrationExecutor) Setup(ctx context.Context, executionCtx *testing.ExecutionContext) error {
	// 调用基础设置
	if err := e.BaseExecutor.Setup(ctx, executionCtx); err != nil {
		return err
	}

	// 初始化集成上下文
	integrationCtx := &IntegrationContext{
		SharedData:     make(map[string]interface{}),
		Environment:    make(map[string]string),
		Variables:      make(map[string]interface{}),
		SessionTokens:  make(map[string]string),
		DatabaseState:  make(map[string]interface{}),
		MockServices:   make(map[string]interface{}),
		Configuration:  make(map[string]interface{}),
	}

	// 将集成上下文添加到执行上下文
	if executionCtx.Variables == nil {
		executionCtx.Variables = make(map[string]interface{})
	}
	executionCtx.Variables["integration_context"] = integrationCtx

	// 设置环境变量
	if executionCtx.Environment != "" {
		integrationCtx.Environment["test_environment"] = executionCtx.Environment
	}

	// 设置基础配置
	integrationCtx.Configuration["timeout"] = e.options.Timeout
	integrationCtx.Configuration["retries"] = e.options.Retries

	e.logger.Info("Integration executor setup completed", "phases", len(e.testPhases))
	return nil
}

// Teardown 清理集成测试执行器
func (e *IntegrationExecutor) Teardown(ctx context.Context, executionCtx *testing.ExecutionContext) error {
	defer e.BaseExecutor.Teardown(ctx, executionCtx)

	// 清理所有子执行器
	if e.uiExecutor != nil {
		if err := e.uiExecutor.Teardown(ctx, executionCtx); err != nil {
			e.logger.Warn("Failed to teardown UI executor", "error", err)
		}
	}

	if e.apiExecutor != nil {
		// API执行器不需要特殊的清理
	}

	if e.perfExecutor != nil {
		if err := e.perfExecutor.Teardown(ctx, executionCtx); err != nil {
			e.logger.Warn("Failed to teardown performance executor", "error", err)
		}
	}

	// 生成集成测试报告
	e.generateIntegrationReport()

	e.logger.Info("Integration executor teardown completed")
	return nil
}

// executeMainTest 执行集成测试的主逻辑
func (e *IntegrationExecutor) executeMainTest(ctx context.Context, test *testing.TestCase, result *testing.TestResult, executionCtx *testing.ExecutionContext) error {
	e.logger.Info("Executing integration test", "name", test.Name)

	// 解析测试阶段
	if err := e.parseTestPhases(test); err != nil {
		return fmt.Errorf("failed to parse test phases: %w", err)
	}

	// 按顺序执行测试阶段
	for i, phase := range e.testPhases {
		e.currentPhase = i
		e.logger.Info("Executing integration phase", "phase", phase.Name, "type", phase.Type)

		// 检查依赖
		if err := e.checkPhaseDependencies(phase); err != nil {
			return fmt.Errorf("phase dependencies not satisfied: %w", err)
		}

		// 执行阶段设置
		if err := e.executePhaseSetup(ctx, phase, executionCtx); err != nil {
			return fmt.Errorf("phase setup failed: %w", err)
		}

		// 执行阶段测试
		phaseResult, err := e.executePhaseTests(ctx, phase, executionCtx)
		if err != nil {
			return fmt.Errorf("phase tests failed: %w", err)
		}

		// 记录阶段结果
		e.phaseResults[phase.Name] = phaseResult

		// 执行阶段清理
		if err := e.executePhaseTeardown(ctx, phase, executionCtx); err != nil {
			e.logger.Warn("Phase teardown failed", "phase", phase.Name, "error", err)
		}

		// 如果阶段失败且不是可选的，停止执行
		if phaseResult.Status == testing.TestStatusFailed && !e.isOptionalPhase(phase) {
			result.Status = testing.TestStatusFailed
			result.Error = phaseResult.Error
			return fmt.Errorf("integration phase '%s' failed", phase.Name)
		}
	}

	// 合并所有阶段结果
	e.mergePhaseResults(result)

	// 记录集成测试指标
	e.recordIntegrationMetrics(result)

	return nil
}

// parseTestPhases 解析测试阶段
func (e *IntegrationExecutor) parseTestPhases(test *testing.TestCase) error {
	// 从测试步骤中解析阶段配置
	e.testPhases = make([]TestPhase, 0)

	// 如果有配置数据，解析它
	if test.Metadata != nil {
		if phasesData, exists := test.Metadata["phases"]; exists {
			if phases, ok := phasesData.([]interface{}); ok {
				for _, phaseData := range phases {
					if phaseMap, ok := phaseData.(map[string]interface{}); ok {
						phase := TestPhase{
							Name:         getString(phaseMap, "name", fmt.Sprintf("Phase %d", len(e.testPhases)+1)),
							Type:         testing.TestType(getString(phaseMap, "type", "api")),
							Parallel:     getBool(phaseMap, "parallel", false),
							Dependencies: getStringSlice(phaseMap, "dependencies"),
						}

						// 解析超时
						if timeoutStr, exists := phaseMap["timeout"]; exists {
							if timeout, err := time.ParseDuration(timeoutStr.(string)); err == nil {
								phase.Timeout = timeout
							}
						}

						// 解析环境配置
						if envData, exists := phaseMap["environment"]; exists {
							if env, ok := envData.(map[string]interface{}); ok {
								phase.Environment = env
							}
						}

						e.testPhases = append(e.testPhases, phase)
					}
				}
			}
		}
	}

	// 如果没有明确定义阶段，创建默认阶段
	if len(e.testPhases) == 0 {
		e.testPhases = append(e.testPhases, TestPhase{
			Name:      "Default",
			Type:      testing.TestTypeAPI,
			Tests:     []*testing.TestCase{test},
			Parallel:  false,
			Timeout:   test.Timeout,
		})
	}

	return nil
}

// checkPhaseDependencies 检查阶段依赖
func (e *IntegrationExecutor) checkPhaseDependencies(phase TestPhase) error {
	for _, dependency := range phase.Dependencies {
		if _, exists := e.phaseResults[dependency]; !exists {
			return fmt.Errorf("dependency '%s' not found", dependency)
		}
		if e.phaseResults[dependency].Status == testing.TestStatusFailed {
			return fmt.Errorf("dependency '%s' failed", dependency)
		}
	}
	return nil
}

// executePhaseSetup 执行阶段设置
func (e *IntegrationExecutor) executePhaseSetup(ctx context.Context, phase TestPhase, executionCtx *testing.ExecutionContext) error {
	if len(phase.Setup) == 0 {
		return nil
	}

	e.logger.Debug("Executing phase setup", "phase", phase.Name)

	for i, step := range phase.Setup {
		if err := e.executePhaseStep(ctx, step, executionCtx, phase.Type); err != nil {
			return fmt.Errorf("setup step %d failed: %w", i+1, err)
		}
	}

	return nil
}

// executePhaseTests 执行阶段测试
func (e *IntegrationExecutor) executePhaseTests(ctx context.Context, phase TestPhase, executionCtx *testing.ExecutionContext) (*testing.TestResult, error) {
	// 创建阶段结果
	phaseResult := &testing.TestResult{
		ID:          fmt.Sprintf("phase_%s_%d", phase.Name, time.Now().Unix()),
		TestCaseID:  phase.Name,
		ExecutionID: executionCtx.ExecutionID,
		Name:        phase.Name,
		Type:        phase.Type,
		Status:      testing.TestStatusRunning,
		StartTime:   time.Now(),
		Metadata:    make(map[string]interface{}),
		Logs:        []string{},
		Screenshots: []string{},
	}

	// 更新执行上下文
	if phase.Timeout > 0 {
		executionCtx.Timeout = phase.Timeout
	}

	// 合并环境配置
	if integrationCtx, exists := executionCtx.Variables["integration_context"]; exists {
		if ic, ok := integrationCtx.(*IntegrationContext); ok {
			for key, value := range phase.Environment {
				ic.Variables[key] = value
			}
		}
	}

	// 选择适当的执行器
	var executor testing.TestExecutor
	switch phase.Type {
	case testing.TestTypeAPI:
		executor = e.apiExecutor
	case testing.TestTypeUI:
		executor = e.uiExecutor
	case testing.TestTypePerformance:
		executor = e.perfExecutor
	default:
		return nil, fmt.Errorf("unsupported test type: %s", phase.Type)
	}

	// 设置执行器
	if err := executor.Setup(ctx, executionCtx); err != nil {
		phaseResult.Status = testing.TestStatusFailed
		phaseResult.Error = &testing.TestError{
			Type:    "setup_error",
			Message: err.Error(),
		}
		return phaseResult, err
	}

	defer executor.Teardown(ctx, executionCtx)

	// 执行测试
	if phase.Parallel && len(phase.Tests) > 1 {
		// 并行执行
		results, err := executor.ExecuteBatch(ctx, phase.Tests, executionCtx)
		if err != nil {
			phaseResult.Status = testing.TestStatusFailed
			phaseResult.Error = &testing.TestError{
				Type:    "execution_error",
				Message: err.Error(),
			}
			return phaseResult, err
		}

		// 合并批量结果
		e.mergeBatchResults(phaseResult, results)
	} else {
		// 串行执行
		for _, test := range phase.Tests {
			testResult, err := executor.Execute(ctx, test, executionCtx)
			if err != nil {
				phaseResult.Status = testing.TestStatusFailed
				phaseResult.Error = &testing.TestError{
					Type:    "execution_error",
					Message: err.Error(),
				}
				return phaseResult, err
			}

			// 合并测试结果
			e.mergeTestResult(phaseResult, testResult)
		}
	}

	// 设置阶段结果为成功
	phaseResult.Status = testing.TestStatusPassed
	phaseResult.Passed = true
	phaseResult.EndTime = time.Now()
	phaseResult.Duration = phaseResult.EndTime.Sub(phaseResult.StartTime)

	e.logger.Info("Phase completed successfully", "phase", phase.Name, "duration", phaseResult.Duration)
	return phaseResult, nil
}

// executePhaseTeardown 执行阶段清理
func (e *IntegrationExecutor) executePhaseTeardown(ctx context.Context, phase TestPhase, executionCtx *testing.ExecutionContext) error {
	if len(phase.Teardown) == 0 {
		return nil
	}

	e.logger.Debug("Executing phase teardown", "phase", phase.Name)

	for i, step := range phase.Teardown {
		if err := e.executePhaseStep(ctx, step, executionCtx, phase.Type); err != nil {
			return fmt.Errorf("teardown step %d failed: %w", i+1, err)
		}
	}

	return nil
}

// executePhaseStep 执行阶段步骤
func (e *IntegrationExecutor) executePhaseStep(ctx context.Context, step testing.TestStep, executionCtx *testing.ExecutionContext, testType testing.TestType) error {
	switch testType {
	case testing.TestTypeAPI:
		return e.apiExecutor.ExecuteStep(ctx, step, executionCtx)
	case testing.TestTypeUI:
		return e.uiExecutor.ExecuteStep(ctx, step, executionCtx)
	case testing.TestTypePerformance:
		return e.perfExecutor.ExecuteStep(ctx, step, executionCtx)
	default:
		return fmt.Errorf("unsupported test type for step execution: %s", testType)
	}
}

// mergeBatchResults 合并批量测试结果
func (e *IntegrationExecutor) mergeBatchResults(phaseResult *testing.TestResult, testResults []*testing.TestResult) {
	for _, result := range testResults {
		e.mergeTestResult(phaseResult, result)
	}
}

// mergeTestResult 合并单个测试结果
func (e *IntegrationExecutor) mergeTestResult(phaseResult *testing.TestResult, testResult *testing.TestResult) {
	// 合并断言
	phaseResult.Assertions = append(phaseResult.Assertions, testResult.Assertions...)

	// 合并截图
	phaseResult.Screenshots = append(phaseResult.Screenshots, testResult.Screenshots...)

	// 合并日志
	phaseResult.Logs = append(phaseResult.Logs, testResult.Logs...)

	// 合并元数据
	for key, value := range testResult.Metadata {
		phaseResult.Metadata[key] = value
	}

	// 如果任何测试失败，阶段失败
	if testResult.Status == testing.TestStatusFailed {
		phaseResult.Status = testing.TestStatusFailed
		phaseResult.Passed = false
		if testResult.Error != nil {
			phaseResult.Error = testResult.Error
		}
	}
}

// mergePhaseResults 合并所有阶段结果
func (e *IntegrationExecutor) mergePhaseResults(result *testing.TestResult) {
	result.Metadata["phase_results"] = e.phaseResults
	result.Metadata["total_phases"] = len(e.testPhases)

	// 统计成功和失败的阶段
	successCount := 0
	failureCount := 0
	for _, phaseResult := range e.phaseResults {
		if phaseResult.Status == testing.TestStatusPassed {
			successCount++
		} else {
			failureCount++
		}
	}

	result.Metadata["successful_phases"] = successCount
	result.Metadata["failed_phases"] = failureCount

	// 合并所有阶段的截图和日志
	for _, phaseResult := range e.phaseResults {
		result.Screenshots = append(result.Screenshots, phaseResult.Screenshots...)
		result.Logs = append(result.Logs, phaseResult.Logs...)
	}
}

// recordIntegrationMetrics 记录集成测试指标
func (e *IntegrationExecutor) recordIntegrationMetrics(result *testing.TestResult) {
	// 记录总体执行时间
	result.Metadata["total_execution_time"] = result.Duration
	result.Metadata["average_phase_time"] = result.Duration / time.Duration(len(e.testPhases))

	// 记录阶段分布
	phaseDistribution := make(map[testing.TestType]int)
	for _, phase := range e.testPhases {
		phaseDistribution[phase.Type]++
	}
	result.Metadata["phase_distribution"] = phaseDistribution

	// 记录依赖解析结果
	dependencyGraph := make(map[string][]string)
	for _, phase := range e.testPhases {
		dependencyGraph[phase.Name] = phase.Dependencies
	}
	result.Metadata["dependency_graph"] = dependencyGraph
}

// generateIntegrationReport 生成集成测试报告
func (e *IntegrationExecutor) generateIntegrationReport() {
	e.logger.Info("=== Integration Test Report ===")
	e.logger.Info("Total Phases", "count", len(e.testPhases))

	successCount := 0
	failureCount := 0
	for _, phase := range e.testPhases {
		if result, exists := e.phaseResults[phase.Name]; exists {
			if result.Status == testing.TestStatusPassed {
				successCount++
				e.logger.Info("Phase Success", "name", phase.Name, "duration", result.Duration)
			} else {
				failureCount++
				e.logger.Info("Phase Failed", "name", phase.Name, "error", result.Error.Message)
			}
		}
	}

	e.logger.Info("Phase Summary", "success", successCount, "failure", failureCount)
	e.logger.Info("=== End Integration Report ===")
}

// isOptionalPhase 检查阶段是否为可选
func (e *IntegrationExecutor) isOptionalPhase(phase TestPhase) bool {
	// 可以根据业务逻辑判断阶段是否为可选
	// 例如：某些性能测试可能失败但不应阻止其他测试
	return false // 默认所有阶段都是必需的
}

// GetPhaseResults 获取阶段结果
func (e *IntegrationExecutor) GetPhaseResults() map[string]*testing.TestResult {
	return e.phaseResults
}

// GetCurrentPhase 获取当前阶段
func (e *IntegrationExecutor) GetCurrentPhase() int {
	return e.currentPhase
}

// AddTestPhase 添加测试阶段
func (e *IntegrationExecutor) AddTestPhase(phase TestPhase) {
	e.testPhases = append(e.testPhases, phase)
}

// 清理阶段
func (e *IntegrationExecutor) ClearTestPhases() {
	e.testPhases = make([]TestPhase, 0)
	e.phaseResults = make(map[string]*testing.TestResult)
}

// 辅助函数
func getString(m map[string]interface{}, key, defaultValue string) string {
	if value, exists := m[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getBool(m map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := m[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if value, exists := m[key]; exists {
		if slice, ok := value.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, item := range slice {
				if str, ok := item.(string); ok {
					result[i] = str
				}
			}
			return result
		}
	}
	return []string{}
}