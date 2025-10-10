package testing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"law-oa-go/internal/models"
)

// TestService 测试服务
type TestService struct {
	repository      TestRepository
	executorManager *ExecutorManager
	scheduler       *TestScheduler
	logger          TestLogger
	metrics         TestMetrics
	config          *TestServiceConfig
	mu              sync.RWMutex
}

// TestServiceConfig 测试服务配置
type TestServiceConfig struct {
	// 执行器配置
	ExecutorConfig *ExecutorConfig `json:"executor_config"`

	// 调度器配置
	SchedulerConfig *SchedulerConfig `json:"scheduler_config"`

	// 服务配置
	MaxStoredResults int          `json:"max_stored_results"`
	ResultTTL       time.Duration `json:"result_ttl"`
	EnableMetrics   bool         `json:"enable_metrics"`
	EnableLogging   bool         `json:"enable_logging"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// NewTestService 创建测试服务
func NewTestService(repository TestRepository, logger TestLogger, metrics TestMetrics, config *TestServiceConfig) *TestService {
	if config == nil {
		config = NewDefaultTestServiceConfig()
	}

	// 创建执行器管理器
	executorOptions := config.ExecutorConfig.ToTestExecutorOptions()
	executorManager := NewExecutorManager(GlobalExecutorFactory, executorOptions, logger, metrics)

	// 创建调度器
	scheduler := NewTestScheduler(executorManager, logger, metrics, config.SchedulerConfig)

	service := &TestService{
		repository:      repository,
		executorManager: executorManager,
		scheduler:       scheduler,
		logger:          logger,
		metrics:         metrics,
		config:          config,
	}

	return service
}

// NewDefaultTestServiceConfig 创建默认测试服务配置
func NewDefaultTestServiceConfig() *TestServiceConfig {
	return &TestServiceConfig{
		ExecutorConfig:    NewDefaultExecutorConfig(),
		SchedulerConfig:   NewDefaultSchedulerConfig(),
		MaxStoredResults:  10000,
		ResultTTL:         7 * 24 * time.Hour, // 7天
		EnableMetrics:     true,
		EnableLogging:     true,
		CleanupInterval:   1 * time.Hour,
	}
}

// Start 启动测试服务
func (s *TestService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 启动调度器
	if err := s.scheduler.Start(); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	// 启动清理任务
	go s.cleanupLoop()

	s.logger.Info("Test service started")
	return nil
}

// Stop 停止测试服务
func (s *TestService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 停止调度器
	if err := s.scheduler.Stop(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	// 释放所有执行器资源
	if err := s.executorManager.ReleaseAllExecutors(); err != nil {
		s.logger.Warn("Failed to release executors", "error", err)
	}

	s.logger.Info("Test service stopped")
	return nil
}

// CreateTestSuite 创建测试套件
func (s *TestService) CreateTestSuite(ctx context.Context, suite *models.TestSuite) error {
	// 验证测试套件
	if err := s.validateTestSuite(suite); err != nil {
		return fmt.Errorf("invalid test suite: %w", err)
	}

	// 保存到数据库
	if err := s.repository.CreateTestSuite(ctx, suite); err != nil {
		return fmt.Errorf("failed to create test suite: %w", err)
	}

	s.logger.Info("Test suite created", "suite_id", suite.ID, "name", suite.Name)
	return nil
}

// GetTestSuite 获取测试套件
func (s *TestService) GetTestSuite(ctx context.Context, suiteID string) (*models.TestSuite, error) {
	suite, err := s.repository.GetTestSuite(ctx, suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test suite: %w", err)
	}

	return suite, nil
}

// ListTestSuites 列出测试套件
func (s *TestService) ListTestSuites(ctx context.Context, filter *TestSuiteFilter) ([]*models.TestSuite, error) {
	suites, err := s.repository.ListTestSuites(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list test suites: %w", err)
	}

	return suites, nil
}

// UpdateTestSuite 更新测试套件
func (s *TestService) UpdateTestSuite(ctx context.Context, suite *models.TestSuite) error {
	// 验证测试套件
	if err := s.validateTestSuite(suite); err != nil {
		return fmt.Errorf("invalid test suite: %w", err)
	}

	// 更新数据库
	if err := s.repository.UpdateTestSuite(ctx, suite); err != nil {
		return fmt.Errorf("failed to update test suite: %w", err)
	}

	s.logger.Info("Test suite updated", "suite_id", suite.ID)
	return nil
}

// DeleteTestSuite 删除测试套件
func (s *TestService) DeleteTestSuite(ctx context.Context, suiteID string) error {
	// 检查是否有正在运行的测试
	if s.hasRunningTestsForSuite(suiteID) {
		return fmt.Errorf("cannot delete test suite with running tests")
	}

	// 删除相关数据
	if err := s.repository.DeleteTestSuite(ctx, suiteID); err != nil {
		return fmt.Errorf("failed to delete test suite: %w", err)
	}

	s.logger.Info("Test suite deleted", "suite_id", suiteID)
	return nil
}

// CreateTestCase 创建测试用例
func (s *TestService) CreateTestCase(ctx context.Context, testCase *TestCase) error {
	// 验证测试用例
	if err := s.validateTestCase(testCase); err != nil {
		return fmt.Errorf("invalid test case: %w", err)
	}

	// 保存到数据库
	if err := s.repository.CreateTestCase(ctx, testCase); err != nil {
		return fmt.Errorf("failed to create test case: %w", err)
	}

	s.logger.Info("Test case created", "case_id", testCase.ID, "name", testCase.Name)
	return nil
}

// GetTestCase 获取测试用例
func (s *TestService) GetTestCase(ctx context.Context, caseID string) (*TestCase, error) {
	testCase, err := s.repository.GetTestCase(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test case: %w", err)
	}

	return testCase, nil
}

// ListTestCases 列出测试用例
func (s *TestService) ListTestCases(ctx context.Context, filter *TestCaseFilter) ([]*TestCase, error) {
	testCases, err := s.repository.ListTestCases(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list test cases: %w", err)
	}

	return testCases, nil
}

// UpdateTestCase 更新测试用例
func (s *TestService) UpdateTestCase(ctx context.Context, testCase *TestCase) error {
	// 验证测试用例
	if err := s.validateTestCase(testCase); err != nil {
		return fmt.Errorf("invalid test case: %w", err)
	}

	// 更新数据库
	if err := s.repository.UpdateTestCase(ctx, testCase); err != nil {
		return fmt.Errorf("failed to update test case: %w", err)
	}

	s.logger.Info("Test case updated", "case_id", testCase.ID)
	return nil
}

// DeleteTestCase 删除测试用例
func (s *TestService) DeleteTestCase(ctx context.Context, caseID string) error {
	// 检查是否有正在运行的测试
	if s.hasRunningTestsForCase(caseID) {
		return fmt.Errorf("cannot delete test case with running tests")
	}

	// 删除相关数据
	if err := s.repository.DeleteTestCase(ctx, caseID); err != nil {
		return fmt.Errorf("failed to delete test case: %w", err)
	}

	s.logger.Info("Test case deleted", "case_id", caseID)
	return nil
}

// RunTestSuite 运行测试套件
func (s *TestService) RunTestSuite(ctx context.Context, suiteID string, options *RunOptions) (*models.TestExecution, error) {
	// 获取测试套件
	suite, err := s.repository.GetTestSuite(ctx, suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test suite: %w", err)
	}

	// 获取测试用例
	testCases, err := s.repository.ListTestCases(ctx, &TestCaseFilter{
		SuiteID: suiteID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}

	if len(testCases) == 0 {
		return nil, fmt.Errorf("no test cases found in suite")
	}

	// 创建执行记录
	execution := &models.TestExecution{
		ID:         generateExecutionID(),
		SuiteID:    suiteID,
		Status:     TestStatusPending,
		StartTime:  time.Now(),
		TotalTests: len(testCases),
		Config:     options,
	}

	if err := s.repository.CreateTestExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create test execution: %w", err)
	}

	// 提交测试到调度器
	for _, testCase := range testCases {
		executionCtx := &ExecutionContext{
			ExecutionID: fmt.Sprintf("%s_%s", execution.ID, testCase.ID),
			Environment: suite.Environment,
			Variables:   suite.Variables,
			Headers:     suite.Headers,
			Timeout:     suite.Timeout,
			BaseURL:     "",
			Parallel:    suite.Parallel,
			Retries:     suite.Retries,
			RetryDelay:  suite.RetryDelay,
		}

		// 合并运行选项
		if options != nil {
			if options.Timeout > 0 {
				executionCtx.Timeout = options.Timeout
			}
			if options.Environment != "" {
				executionCtx.Environment = options.Environment
			}
			for k, v := range options.Variables {
				executionCtx.Variables[k] = v
			}
		}

		if err := s.scheduler.SubmitTest(testCase, executionCtx); err != nil {
			s.logger.Error("Failed to submit test", "case_id", testCase.ID, "error", err)
		}
	}

	s.logger.Info("Test suite execution started", "execution_id", execution.ID, "suite_id", suiteID)
	return execution, nil
}

// RunTestCase 运行单个测试用例
func (s *TestService) RunTestCase(ctx context.Context, caseID string, options *RunOptions) (*models.TestExecution, error) {
	// 获取测试用例
	testCase, err := s.repository.GetTestCase(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test case: %w", err)
	}

	// 创建执行记录
	execution := &models.TestExecution{
		ID:         generateExecutionID(),
		CaseID:     caseID,
		Status:     TestStatusPending,
		StartTime:  time.Now(),
		TotalTests: 1,
		Config:     options,
	}

	if err := s.repository.CreateTestExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create test execution: %w", err)
	}

	// 创建执行上下文
	executionCtx := &ExecutionContext{
		ExecutionID: fmt.Sprintf("%s_%s", execution.ID, testCase.ID),
		Timeout:     testCase.Timeout,
	}

	if options != nil {
		if options.Timeout > 0 {
			executionCtx.Timeout = options.Timeout
		}
		if options.Environment != "" {
			executionCtx.Environment = options.Environment
		}
		for k, v := range options.Variables {
			if executionCtx.Variables == nil {
				executionCtx.Variables = make(map[string]interface{})
			}
			executionCtx.Variables[k] = v
		}
	}

	// 提交测试到调度器
	if err := s.scheduler.SubmitTest(testCase, executionCtx); err != nil {
		return nil, fmt.Errorf("failed to submit test: %w", err)
	}

	s.logger.Info("Test case execution started", "execution_id", execution.ID, "case_id", caseID)
	return execution, nil
}

// GetExecution 获取测试执行信息
func (s *TestService) GetExecution(ctx context.Context, executionID string) (*models.TestExecution, error) {
	execution, err := s.repository.GetTestExecution(ctx, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test execution: %w", err)
	}

	// 获取调度器状态
	if result, exists := s.scheduler.GetTestResult(executionID); exists {
		execution.Status = result.Status
		execution.EndTime = result.EndTime
		execution.Duration = execution.EndTime.Sub(execution.StartTime)
	}

	return execution, nil
}

// ListExecutions 列出测试执行记录
func (s *TestService) ListExecutions(ctx context.Context, filter *TestExecutionFilter) ([]*models.TestExecution, error) {
	executions, err := s.repository.ListTestExecutions(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list test executions: %w", err)
	}

	// 更新执行状态
	for _, execution := range executions {
		if result, exists := s.scheduler.GetTestResult(execution.ID); exists {
			execution.Status = result.Status
			if result.EndTime.IsZero() {
				execution.Duration = time.Since(execution.StartTime)
			} else {
				execution.EndTime = result.EndTime
				execution.Duration = execution.EndTime.Sub(execution.StartTime)
			}
		}
	}

	return executions, nil
}

// CancelExecution 取消测试执行
func (s *TestService) CancelExecution(ctx context.Context, executionID string) error {
	// 调用调度器取消测试
	if err := s.scheduler.CancelTest(executionID); err != nil {
		return fmt.Errorf("failed to cancel test execution: %w", err)
	}

	// 更新数据库状态
	execution := &models.TestExecution{
		ID:        executionID,
		Status:    TestStatusSkipped,
		EndTime:   time.Now(),
		Duration:  0,
	}

	if err := s.repository.UpdateTestExecution(ctx, execution); err != nil {
		s.logger.Warn("Failed to update execution status", "execution_id", executionID, "error", err)
	}

	s.logger.Info("Test execution cancelled", "execution_id", executionID)
	return nil
}

// GetSchedulerStatus 获取调度器状态
func (s *TestService) GetSchedulerStatus() *SchedulerStatus {
	return s.scheduler.GetStatus()
}

// GetTestResults 获取测试结果
func (s *TestService) GetTestResults(ctx context.Context, executionID string) ([]*TestResult, error) {
	results, err := s.repository.ListTestResults(ctx, &TestResultFilter{
		ExecutionID: executionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get test results: %w", err)
	}

	return results, nil
}

// GetTestResult 获取单个测试结果
func (s *TestService) GetTestResult(ctx context.Context, resultID string) (*TestResult, error) {
	result, err := s.repository.GetTestResult(ctx, resultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test result: %w", err)
	}

	return result, nil
}

// validateTestSuite 验证测试套件
func (s *TestService) validateTestSuite(suite *models.TestSuite) error {
	if suite.Name == "" {
		return fmt.Errorf("suite name cannot be empty")
	}

	if suite.Environment == "" {
		suite.Environment = "test"
	}

	if suite.Timeout == 0 {
		suite.Timeout = s.config.ExecutorConfig.DefaultTimeout.ToDuration()
	}

	if suite.MaxWorkers == 0 {
		suite.MaxWorkers = s.config.ExecutorConfig.MaxConcurrent
	}

	return nil
}

// validateTestCase 验证测试用例
func (s *TestService) validateTestCase(testCase *TestCase) error {
	if testCase.Name == "" {
		return fmt.Errorf("test case name cannot be empty")
	}

	if testCase.Type == "" {
		testCase.Type = TestTypeAPI
	}

	if testCase.Timeout == 0 {
		testCase.Timeout = s.config.ExecutorConfig.DefaultTimeout.ToDuration()
	}

	// 验证API测试
	if testCase.Type == TestTypeAPI {
		if testCase.URL == "" {
			return fmt.Errorf("API test case must have URL")
		}
		if testCase.Method == "" {
			testCase.Method = "GET"
		}
	}

	return nil
}

// hasRunningTestsForSuite 检查测试套件是否有正在运行的测试
func (s *TestService) hasRunningTestsForSuite(suiteID string) bool {
	status := s.scheduler.GetStatus()
	for _, runningTest := range status.RunningTests {
		// 这里需要从runningTest中获取suite_id信息
		// 当前实现中可能需要扩展RunningTest来包含suite信息
	}
	return false
}

// hasRunningTestsForCase 检查测试用例是否有正在运行的测试
func (s *TestService) hasRunningTestsForCase(caseID string) bool {
	status := s.scheduler.GetStatus()
	for _, runningTest := range status.RunningTests {
		// 这里需要从runningTest中获取case_id信息
		// 当前实现中可能需要扩展RunningTest来包含case信息
	}
	return false
}

// cleanupLoop 清理循环
func (s *TestService) cleanupLoop() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

// cleanup 清理过期数据
func (s *TestService) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 清理过期的执行记录
	cutoff := time.Now().Add(-s.config.ResultTTL)
	if err := s.repository.CleanupOldExecutions(ctx, cutoff); err != nil {
		s.logger.Warn("Failed to cleanup old executions", "error", err)
	}

	s.logger.Debug("Test service cleanup completed")
}

// RunOptions 运行选项
type RunOptions struct {
	Timeout     time.Duration          `json:"timeout"`
	Environment string                 `json:"environment"`
	Variables   map[string]interface{} `json:"variables"`
	Tags        []string               `json:"tags"`
	Priority    int                    `json:"priority"`
	Parallel    bool                   `json:"parallel"`
}

// TestSuiteFilter 测试套件过滤器
type TestSuiteFilter struct {
	Name        string    `json:"name"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	Tags        []string  `json:"tags"`
	CreatedAfter *time.Time `json:"created_after"`
	CreatedBefore *time.Time `json:"created_before"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}

// TestCaseFilter 测试用例过滤器
type TestCaseFilter struct {
	SuiteID     string    `json:"suite_id"`
	Name        string    `json:"name"`
	Type        TestType  `json:"type"`
	Status      string    `json:"status"`
	Tags        []string  `json:"tags"`
	CreatedAfter *time.Time `json:"created_after"`
	CreatedBefore *time.Time `json:"created_before"`
	Limit       int       `json:"limit"`
	Offset      int       `json:"offset"`
}

// TestExecutionFilter 测试执行过滤器
type TestExecutionFilter struct {
	SuiteID      string    `json:"suite_id"`
	CaseID       string    `json:"case_id"`
	Status       TestStatus `json:"status"`
	StartedAfter *time.Time `json:"started_after"`
	StartedBefore *time.Time `json:"started_before"`
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
}

// TestResultFilter 测试结果过滤器
type TestResultFilter struct {
	ExecutionID  string    `json:"execution_id"`
	TestCaseID   string    `json:"test_case_id"`
	Status       TestStatus `json:"status"`
	StartedAfter *time.Time `json:"started_after"`
	StartedBefore *time.Time `json:"started_before"`
	Limit        int       `json:"limit"`
	Offset       int       `json:"offset"`
}