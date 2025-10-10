package testing

import (
	"context"
	"fmt"
	"time"
)

// BaseExecutor 基础测试执行器
type BaseExecutor struct {
	options   *TestExecutorOptions
	logger    TestLogger
	metrics   TestMetrics
}

// NewBaseExecutor 创建基础执行器
func NewBaseExecutor(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) *BaseExecutor {
	return &BaseExecutor{
		options: options,
		logger:  logger,
		metrics: metrics,
	}
}

// Setup 基础设置
func (e *BaseExecutor) Setup(ctx context.Context, executionCtx *ExecutionContext) error {
	e.logger.Info("Setting up test execution", "execution_id", executionCtx.ExecutionID)

	// 设置超时
	if executionCtx.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, executionCtx.Timeout)
		executionCtx.Context = ctx
		executionCtx.Cancel = cancel
	}

	return nil
}

// Teardown 基础清理
func (e *BaseExecutor) Teardown(ctx context.Context, executionCtx *ExecutionContext) error {
	e.logger.Info("Tearing down test execution", "execution_id", executionCtx.ExecutionID)

	// 取消上下文
	if executionCtx.Cancel != nil {
		executionCtx.Cancel()
	}

	return nil
}

// Execute 执行测试的通用逻辑
func (e *BaseExecutor) Execute(ctx context.Context, test *TestCase, executionCtx *ExecutionContext) (*TestResult, error) {
	result := &TestResult{
		ID:          generateTestResultID(),
		TestCaseID:  test.ID,
		ExecutionID: executionCtx.ExecutionID,
		Name:        test.Name,
		Type:        test.Type,
		Status:      TestStatusRunning,
		StartTime:   time.Now(),
		Metadata:    make(map[string]interface{}),
		Logs:        []string{},
		Screenshots: []string{},
	}

	// 记录开始
	e.logger.Info("Starting test execution",
		"test_id", test.ID,
		"test_name", test.Name,
		"test_type", test.Type,
		"execution_id", executionCtx.ExecutionID)

	// 执行前置步骤
	if err := e.executeSteps(ctx, test.Setup, result, executionCtx); err != nil {
		result.Status = TestStatusFailed
		result.Error = &TestError{
			Type:       "setup_error",
			Message:    err.Error(),
			StackTrace: fmt.Sprintf("Setup steps failed: %v", err),
		}
		return e.finalizeResult(result), nil
	}

	// 执行主测试逻辑
	if err := e.executeMainTest(ctx, test, result, executionCtx); err != nil {
		result.Status = TestStatusFailed
		result.Error = &TestError{
			Type:       "execution_error",
			Message:    err.Error(),
			StackTrace: fmt.Sprintf("Test execution failed: %v", err),
		}
	} else {
		result.Status = TestStatusPassed
		result.Passed = true
	}

	// 执行后置步骤
	if err := e.executeSteps(ctx, test.Teardown, result, executionCtx); err != nil {
		e.logger.Warn("Teardown steps failed",
			"test_id", test.ID,
			"error", err.Error())
		// 后置步骤失败不影响整体测试结果
	}

	return e.finalizeResult(result), nil
}

// ExecuteBatch 批量执行测试
func (e *BaseExecutor) ExecuteBatch(ctx context.Context, tests []*TestCase, executionCtx *ExecutionContext) ([]*TestResult, error) {
	results := make([]*TestResult, 0, len(tests))

	if !executionCtx.Parallel || len(tests) == 1 {
		// 串行执行
		for _, test := range tests {
			result, err := e.Execute(ctx, test, executionCtx)
			if err != nil {
				e.logger.Error("Failed to execute test", "test_id", test.ID, "error", err)
			}
			results = append(results, result)
		}
	} else {
		// 并行执行
		resultsChan := make(chan *TestResult, len(tests))
		semaphore := make(chan struct{}, e.options.MaxConcurrent)

		for _, test := range tests {
			go func(t *TestCase) {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				result, err := e.Execute(ctx, t, executionCtx)
				if err != nil {
					e.logger.Error("Failed to execute test in batch", "test_id", t.ID, "error", err)
				}
				resultsChan <- result
			}(test)
		}

		// 收集结果
		for i := 0; i < len(tests); i++ {
			results = append(results, <-resultsChan)
		}
	}

	return results, nil
}

// executeMainTest 执行主测试逻辑（需要子类实现）
func (e *BaseExecutor) executeMainTest(ctx context.Context, test *TestCase, result *TestResult, executionCtx *ExecutionContext) error {
	// 这个方法需要在子类中实现
	return fmt.Errorf("executeMainTest must be implemented by subclasses")
}

// executeSteps 执行测试步骤
func (e *BaseExecutor) executeSteps(ctx context.Context, steps []TestStep, result *TestResult, executionCtx *ExecutionContext) error {
	for i, step := range steps {
		e.logger.Debug("Executing step",
			"step", i+1,
			"name", step.Name,
			"type", step.Type,
			"action", step.Action)

		if err := e.executeStep(ctx, step, result, executionCtx); err != nil {
			return fmt.Errorf("step %d failed: %w", i+1, err)
		}

		// 等待指定时间
		if step.Wait > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(step.Wait):
				// 等待完成
			}
		}
	}

	return nil
}

// executeStep 执行单个步骤
func (e *BaseExecutor) executeStep(ctx context.Context, step TestStep, result *TestResult, executionCtx *ExecutionContext) error {
	// 根据步骤类型执行不同的逻辑
	switch step.Type {
	case "navigate":
		return e.executeNavigate(ctx, step, executionCtx)
	case "click":
		return e.executeClick(ctx, step, executionCtx)
	case "fill":
		return e.executeFill(ctx, step, executionCtx)
	case "wait":
		return e.executeWait(ctx, step, executionCtx)
	case "assert":
		return e.executeAssert(ctx, step, result)
	case "screenshot":
		return e.executeScreenshot(ctx, step, result)
	case "javascript":
		return e.executeJavaScript(ctx, step, result, executionCtx)
	default:
		e.logger.Warn("Unknown step type, skipping", "type", step.Type, "action", step.Action)
		return nil
	}
}

// 各种步骤执行方法（需要在子类中具体实现或保持为空实现）
func (e *BaseExecutor) executeNavigate(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	e.logger.Debug("Navigate step not implemented in base executor", "target", step.Target)
	return nil
}

func (e *BaseExecutor) executeClick(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	e.logger.Debug("Click step not implemented in base executor", "target", step.Target)
	return nil
}

func (e *BaseExecutor) executeFill(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	e.logger.Debug("Fill step not implemented in base executor", "target", step.Target)
	return nil
}

func (e *BaseExecutor) executeWait(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	e.logger.Debug("Wait step not implemented in base executor", "target", step.Target)
	return nil
}

func (e *BaseExecutor) executeAssert(ctx context.Context, step TestStep, result *TestResult) error {
	// 执行断言
	assertion := &TestAssertion{
		Name:     step.Name,
		Type:     step.Type,
		Actual:   nil,
		Expected: step.Expected,
		Passed:   false,
	}

	if err := e.evaluateAssertion(step, assertion); err != nil {
		assertion.ErrorMessage = err.Error()
	} else {
		assertion.Passed = true
	}

	result.Assertions = append(result.Assertions, assertion)

	if !assertion.Passed {
		return fmt.Errorf("assertion '%s' failed: %s", assertion.Name, assertion.ErrorMessage)
	}

	return nil
}

func (e *BaseExecutor) evaluateAssertion(step TestStep, assertion *TestAssertion) error {
	// 基础断言评估逻辑
	// 这里需要根据具体的期望类型来实现
	return nil
}

func (e *BaseExecutor) executeScreenshot(ctx context.Context, step TestStep, result *TestResult) error {
	// 截图实现
	e.logger.Debug("Screenshot step requested", "name", step.Name)
	return nil
}

func (e *BaseExecutor) executeJavaScript(ctx context.Context, step TestStep, result *TestResult, executionCtx *ExecutionContext) error {
	// JavaScript执行实现
	e.logger.Debug("JavaScript step not implemented in base executor", "action", step.Action)
	return nil
}

// finalizeResult 完成结果
func (e *BaseExecutor) finalizeResult(result *TestResult) *TestResult {
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// 记录指标
	e.metrics.RecordExecution(result.Type, result.Duration, result.Status)

	// 记录断言
	for _, assertion := range result.Assertions {
		e.metrics.RecordAssertion(result.Type, assertion.Passed)
	}

	// 记录错误
	if result.Error != nil {
		e.metrics.RecordError(result.Type, result.Error.Type)
	}

	e.logger.Info("Test execution completed",
		"test_id", result.TestCaseID,
		"status", result.Status,
		"duration", result.Duration,
		"passed", result.Passed)

	return result
}

// generateTestResultID 生成测试结果ID
func generateTestResultID() string {
	return fmt.Sprintf("result_%d", time.Now().UnixNano())
}

// validateTestConfig 验证测试配置
func (e *BaseExecutor) validateTestConfig(config *TestSuiteConfig) error {
	if config.Name == "" {
		return fmt.Errorf("test suite name cannot be empty")
	}

	if config.MaxWorkers <= 0 {
		config.MaxWorkers = e.options.MaxConcurrent
	}

	if config.Timeout <= 0 {
		config.Timeout = e.options.Timeout
	}

	if config.Retries < 0 {
		config.Retries = e.options.Retries
	}

	if config.RetryDelay <= 0 {
		config.RetryDelay = e.options.RetryDelay
	}

	return nil
}

// mergeExecutionContext 合并执行上下文
func (e *BaseExecutor) mergeExecutionContext(suiteConfig *TestSuiteConfig, executionCtx *ExecutionContext) *ExecutionContext {
	merged := &ExecutionContext{
		ExecutionID: executionCtx.ExecutionID,
		Environment: executionCtx.Environment,
		Variables:   make(map[string]interface{}),
		Headers:     make(map[string]string),
		Timeout:     executionCtx.Timeout,
		WorkingDir:  executionCtx.WorkingDir,
		BaseURL:     executionCtx.BaseURL,
		Parallel:    executionCtx.Parallel,
		Retries:     executionCtx.Retries,
		RetryDelay:  executionCtx.RetryDelay,
		Logger:      executionCtx.Logger,
		Metrics:     executionCtx.Metrics,
		Context:     executionCtx.Context,
	}

	// 合并套件配置
	if suiteConfig != nil {
		// 合并变量
		for k, v := range suiteConfig.Variables {
			merged.Variables[k] = v
		}

		// 合并请求头
		for k, v := range suiteConfig.Headers {
			merged.Headers[k] = v
		}

		// 设置环境变量
		if merged.Environment == "" && suiteConfig.Environment != "" {
			merged.Environment = suiteConfig.Environment
		}

		// 设置超时
		if merged.Timeout == 0 && suiteConfig.Timeout > 0 {
			merged.Timeout = suiteConfig.Timeout
		}

		// 设置并发配置
		if !merged.Parallel && suiteConfig.Parallel {
			merged.Parallel = true
		}

		// 设置重试配置
		if merged.Retries == 0 && suiteConfig.Retries > 0 {
			merged.Retries = suiteConfig.Retries
		}

		if merged.RetryDelay == 0 && suiteConfig.RetryDelay > 0 {
			merged.RetryDelay = suiteConfig.RetryDelay
		}
	}

	// 应用执行器选项
	if merged.Timeout == 0 && e.options.Timeout > 0 {
		merged.Timeout = e.options.Timeout
	}

	if merged.Retries == 0 && e.options.Retries > 0 {
		merged.Retries = e.options.Retries
	}

	if merged.RetryDelay == 0 && e.options.RetryDelay > 0 {
		merged.RetryDelay = e.options.RetryDelay
	}

	return merged
}

// NewTestResult 创建新的测试结果
func NewTestResult(testCaseID, executionID string) *TestResult {
	return &TestResult{
		ID:          generateTestResultID(),
		TestCaseID:  testCaseID,
		ExecutionID: executionID,
		Status:      TestStatusPending,
		Metadata:    make(map[string]interface{}),
		Logs:        []string{},
		Screenshots: []string{},
	}
}


// GetExecutorType 获取执行器类型
func (e *BaseExecutor) GetExecutorType() TestType {
	return TestTypeAPI // 默认类型，子类应该重写
}