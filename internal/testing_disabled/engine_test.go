package testing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockLogger 模拟日志器
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.logs = append(m.logs, "INFO: "+msg)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.logs = append(m.logs, "WARN: "+msg)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.logs = append(m.logs, "ERROR: "+msg)
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.logs = append(m.logs, "DEBUG: "+msg)
}

// MockMetrics 模拟指标收集器
type MockMetrics struct {
	executions []ExecutionMetric
	assertions []AssertionMetric
	errors     []ErrorMetric
	custom     []CustomMetric
}

type ExecutionMetric struct {
	TestType TestType
	Duration time.Duration
	Status   TestStatus
}

type AssertionMetric struct {
	TestType TestType
	Passed   bool
}

type ErrorMetric struct {
	TestType  TestType
	ErrorType string
}

type CustomMetric struct {
	Metric string
	Value  float64
	Tags   map[string]string
}

func (m *MockMetrics) RecordExecution(testType TestType, duration time.Duration, status TestStatus) {
	m.executions = append(m.executions, ExecutionMetric{
		TestType: testType,
		Duration: duration,
		Status:   status,
	})
}

func (m *MockMetrics) RecordAssertion(testType TestType, passed bool) {
	m.assertions = append(m.assertions, AssertionMetric{
		TestType: testType,
		Passed:   passed,
	})
}

func (m *MockMetrics) RecordError(testType TestType, errorType string) {
	m.errors = append(m.errors, ErrorMetric{
		TestType:  testType,
		ErrorType: errorType,
	})
}

func (m *MockMetrics) RecordCustom(metric string, value float64, tags map[string]string) {
	m.custom = append(m.custom, CustomMetric{
		Metric: metric,
		Value:  value,
		Tags:   tags,
	})
}

// MockRepository 模拟存储库
type MockRepository struct {
	testSuites    map[string]*TestSuite
	testCases     map[string]*TestCase
	testExecutions map[string]*TestExecution
	testResults   map[string]*TestResult
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		testSuites:    make(map[string]*TestSuite),
		testCases:     make(map[string]*TestCase),
		testExecutions: make(map[string]*TestExecution),
		testResults:   make(map[string]*TestResult),
	}
}

func (m *MockRepository) CreateTestSuite(ctx context.Context, suite *TestSuite) error {
	m.testSuites[suite.ID] = suite
	return nil
}

func (m *MockRepository) GetTestSuite(ctx context.Context, suiteID string) (*TestSuite, error) {
	suite, exists := m.testSuites[suiteID]
	if !exists {
		return nil, ErrTestSuiteNotFound
	}
	return suite, nil
}

func (m *MockRepository) ListTestSuites(ctx context.Context, filter *TestSuiteFilter) ([]*TestSuite, error) {
	suites := make([]*TestSuite, 0, len(m.testSuites))
	for _, suite := range m.testSuites {
		suites = append(suites, suite)
	}
	return suites, nil
}

func (m *MockRepository) UpdateTestSuite(ctx context.Context, suite *TestSuite) error {
	m.testSuites[suite.ID] = suite
	return nil
}

func (m *MockRepository) DeleteTestSuite(ctx context.Context, suiteID string) error {
	delete(m.testSuites, suiteID)
	return nil
}

func (m *MockRepository) CreateTestCase(ctx context.Context, testCase *TestCase) error {
	m.testCases[testCase.ID] = testCase
	return nil
}

func (m *MockRepository) GetTestCase(ctx context.Context, caseID string) (*TestCase, error) {
	testCase, exists := m.testCases[caseID]
	if !exists {
		return nil, ErrTestCaseNotFound
	}
	return testCase, nil
}

func (m *MockRepository) ListTestCases(ctx context.Context, filter *TestCaseFilter) ([]*TestCase, error) {
	testCases := make([]*TestCase, 0, len(m.testCases))
	for _, testCase := range m.testCases {
		testCases = append(testCases, testCase)
	}
	return testCases, nil
}

func (m *MockRepository) UpdateTestCase(ctx context.Context, testCase *TestCase) error {
	m.testCases[testCase.ID] = testCase
	return nil
}

func (m *MockRepository) DeleteTestCase(ctx context.Context, caseID string) error {
	delete(m.testCases, caseID)
	return nil
}

func (m *MockRepository) CreateTestExecution(ctx context.Context, execution *TestExecution) error {
	m.testExecutions[execution.ID] = execution
	return nil
}

func (m *MockRepository) GetTestExecution(ctx context.Context, executionID string) (*TestExecution, error) {
	execution, exists := m.testExecutions[executionID]
	if !exists {
		return nil, ErrTestExecutionNotFound
	}
	return execution, nil
}

func (m *MockRepository) ListTestExecutions(ctx context.Context, filter *TestExecutionFilter) ([]*TestExecution, error) {
	executions := make([]*TestExecution, 0, len(m.testExecutions))
	for _, execution := range m.testExecutions {
		executions = append(executions, execution)
	}
	return executions, nil
}

func (m *MockRepository) UpdateTestExecution(ctx context.Context, execution *TestExecution) error {
	m.testExecutions[execution.ID] = execution
	return nil
}

func (m *MockRepository) DeleteTestExecution(ctx context.Context, executionID string) error {
	delete(m.testExecutions, executionID)
	return nil
}

func (m *MockRepository) CreateTestResult(ctx context.Context, result *TestResult) error {
	m.testResults[result.ID] = result
	return nil
}

func (m *MockRepository) GetTestResult(ctx context.Context, resultID string) (*TestResult, error) {
	result, exists := m.testResults[resultID]
	if !exists {
		return nil, ErrTestResultNotFound
	}
	return result, nil
}

func (m *MockRepository) ListTestResults(ctx context.Context, filter *TestResultFilter) ([]*TestResult, error) {
	results := make([]*TestResult, 0, len(m.testResults))
	for _, result := range m.testResults {
		results = append(results, result)
	}
	return results, nil
}

func (m *MockRepository) UpdateTestResult(ctx context.Context, result *TestResult) error {
	m.testResults[result.ID] = result
	return nil
}

func (m *MockRepository) DeleteTestResult(ctx context.Context, resultID string) error {
	delete(m.testResults, resultID)
	return nil
}

func (m *MockRepository) CleanupOldExecutions(ctx context.Context, cutoff time.Time) error {
	// 简单实现：删除所有早于截止时间的执行记录
	for id, execution := range m.testExecutions {
		if execution.StartTime.Before(cutoff) {
			delete(m.testExecutions, id)
		}
	}
	return nil
}

// TestExecutorFactory 测试执行器工厂
func TestExecutorFactory(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	options := DefaultExecutorOptions()

	factory := NewDefaultExecutorFactory()

	// 测试支持的类型
	supportedTypes := factory.GetSupportedTypes()
	assert.Contains(t, supportedTypes, TestTypeAPI)
	assert.Contains(t, supportedTypes, TestTypeUI)
	assert.Contains(t, supportedTypes, TestTypePerformance)
	assert.Contains(t, supportedTypes, TestTypeIntegration)

	// 测试创建API执行器
	apiExecutor, err := factory.CreateExecutor(TestTypeAPI, options, logger, metrics)
	require.NoError(t, err)
	assert.Equal(t, TestTypeAPI, apiExecutor.GetExecutorType())

	// 测试创建不支持的类型
	_, err = factory.CreateExecutor("unsupported", options, logger, metrics)
	assert.Error(t, err)

	// 测试注册自定义执行器
	customCreator := func(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) (TestExecutor, error) {
		return &BaseExecutor{}, nil
	}
	err = factory.RegisterExecutor("custom", customCreator)
	require.NoError(t, err)

	// 验证自定义执行器已注册
	supportedTypes = factory.GetSupportedTypes()
	assert.Contains(t, supportedTypes, "custom")

	// 重复注册应该失败
	err = factory.RegisterExecutor("custom", customCreator)
	assert.Error(t, err)
}

// TestExecutorManager 测试执行器管理器
func TestExecutorManager(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	options := DefaultExecutorOptions()
	factory := NewDefaultExecutorFactory()

	manager := NewExecutorManager(factory, options, logger, metrics)

	// 测试获取API执行器
	apiExecutor, err := manager.GetExecutor(TestTypeAPI)
	require.NoError(t, err)
	assert.Equal(t, TestTypeAPI, apiExecutor.GetExecutorType())

	// 测试缓存：再次获取应该返回同一个实例
	apiExecutor2, err := manager.GetExecutor(TestTypeAPI)
	require.NoError(t, err)
	assert.Same(t, apiExecutor, apiExecutor2)

	// 测试释放执行器
	err = manager.ReleaseExecutor(TestTypeAPI)
	require.NoError(t, err)

	// 释放后应该创建新的实例
	apiExecutor3, err := manager.GetExecutor(TestTypeAPI)
	require.NoError(t, err)
	assert.NotSame(t, apiExecutor, apiExecutor3)

	// 测试释放所有执行器
	err = manager.ReleaseAllExecutors()
	require.NoError(t, err)
}

// TestTestQueue 测试队列
func TestTestQueue(t *testing.T) {
	// 测试FIFO队列
	fifoQueue := NewTestQueue(false)

	// 入队
	item1 := &QueueItem{
		Test: &TestCase{ID: "test1"},
		Priority: 1,
		EnqueuedAt: time.Now(),
	}
	item2 := &QueueItem{
		Test: &TestCase{ID: "test2"},
		Priority: 2,
		EnqueuedAt: time.Now(),
	}

	err := fifoQueue.Enqueue(item1)
	require.NoError(t, err)
	err = fifoQueue.Enqueue(item2)
	require.NoError(t, err)

	assert.Equal(t, 2, fifoQueue.Size())

	// 出队应该是FIFO顺序
	dequeued1 := fifoQueue.Dequeue()
	require.NotNil(t, dequeued1)
	assert.Equal(t, "test1", dequeued1.Test.ID)

	dequeued2 := fifoQueue.Dequeue()
	require.NotNil(t, dequeued2)
	assert.Equal(t, "test2", dequeued2.Test.ID)

	assert.Equal(t, 0, fifoQueue.Size())

	// 测试优先级队列
	priorityQueue := NewTestQueue(true)

	item3 := &QueueItem{
		Test: &TestCase{ID: "test3"},
		Priority: 1,
		EnqueuedAt: time.Now(),
	}
	item4 := &QueueItem{
		Test: &TestCase{ID: "test4"},
		Priority: 3,
		EnqueuedAt: time.Now(),
	}
	item5 := &QueueItem{
		Test: &TestCase{ID: "test5"},
		Priority: 2,
		EnqueuedAt: time.Now(),
	}

	priorityQueue.Enqueue(item3)
	priorityQueue.Enqueue(item4)
	priorityQueue.Enqueue(item5)

	// 出队应该按优先级顺序
	dequeued3 := priorityQueue.Dequeue()
	require.NotNil(t, dequeued3)
	assert.Equal(t, "test4", dequeued3.Test.ID) // 优先级最高

	dequeued4 := priorityQueue.Dequeue()
	require.NotNil(t, dequeued4)
	assert.Equal(t, "test5", dequeued4.Test.ID) // 优先级中等

	dequeued5 := priorityQueue.Dequeue()
	require.NotNil(t, dequeued5)
	assert.Equal(t, "test3", dequeued5.Test.ID) // 优先级最低
}

// TestTestService 测试服务
func TestTestService(t *testing.T) {
	logger := &MockLogger{}
	metrics := &MockMetrics{}
	repository := NewMockRepository()
	config := NewDefaultTestServiceConfig()

	service := NewTestService(repository, logger, metrics, config)

	// 启动服务
	err := service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// 创建测试套件
	suite := &TestSuite{
		ID:          "suite1",
		Name:        "Test Suite 1",
		Description: "Test suite for testing",
		Environment: "test",
		Timeout:     30 * time.Second,
		Parallel:    true,
		MaxWorkers:  5,
		Variables:   map[string]interface{}{"key": "value"},
		Headers:     map[string]string{"Content-Type": "application/json"},
	}

	err = service.CreateTestSuite(context.Background(), suite)
	require.NoError(t, err)

	// 获取测试套件
	retrievedSuite, err := service.GetTestSuite(context.Background(), suite.ID)
	require.NoError(t, err)
	assert.Equal(t, suite.Name, retrievedSuite.Name)

	// 创建测试用例
	testCase := &TestCase{
		ID:          "case1",
		SuiteID:     suite.ID,
		Name:        "Test Case 1",
		Description: "API test case",
		Type:        TestTypeAPI,
		Method:      "GET",
		URL:         "https://httpbin.org/get",
		Headers:     map[string]string{"Accept": "application/json"},
		Timeout:     10 * time.Second,
		Expected: &TestExpected{
			Status: 200,
			Headers: map[string]string{"Content-Type": "application/json"},
		},
	}

	err = service.CreateTestCase(context.Background(), testCase)
	require.NoError(t, err)

	// 获取测试用例
	retrievedCase, err := service.GetTestCase(context.Background(), testCase.ID)
	require.NoError(t, err)
	assert.Equal(t, testCase.Name, retrievedCase.Name)

	// 运行测试用例（注意：这会实际发起HTTP请求）
	// 在真实环境中，你可能需要使用模拟服务器
	t.Skip("Skipping actual test execution to avoid external HTTP calls")

	// options := &RunOptions{
	// 	Timeout: 30 * time.Second,
	// }
	// execution, err := service.RunTestCase(context.Background(), testCase.ID, options)
	// require.NoError(t, err)
	// assert.Equal(t, TestStatusPending, execution.Status)

	// 列出测试套件
	suites, err := service.ListTestSuites(context.Background(), &TestSuiteFilter{})
	require.NoError(t, err)
	assert.Len(t, suites, 1)

	// 列出测试用例
	testCases, err := service.ListTestCases(context.Background(), &TestCaseFilter{})
	require.NoError(t, err)
	assert.Len(t, testCases, 1)

	// 获取调度器状态
	status := service.GetSchedulerStatus()
	assert.True(t, status.Running)

	// 删除测试用例
	err = service.DeleteTestCase(context.Background(), testCase.ID)
	require.NoError(t, err)

	// 删除测试套件
	err = service.DeleteTestSuite(context.Background(), suite.ID)
	require.NoError(t, err)
}

// TestExecutorConfig 测试执行器配置
func TestExecutorConfig(t *testing.T) {
	config := NewDefaultExecutorConfig()

	// 测试默认配置
	assert.Equal(t, Duration(30*time.Second), config.DefaultTimeout)
	assert.Equal(t, 3, config.DefaultRetries)
	assert.Equal(t, 10, config.MaxConcurrent)
	assert.Equal(t, "chromium", config.BrowserType)
	assert.True(t, config.Headless)
	assert.True(t, config.EnableScreenshots)
	assert.False(t, config.EnableVideos)

	// 测试转换为TestExecutorOptions
	options := config.ToTestExecutorOptions()
	assert.Equal(t, config.DefaultTimeout.ToDuration(), options.Timeout)
	assert.Equal(t, config.DefaultRetries, options.Retries)
	assert.Equal(t, config.MaxConcurrent, options.MaxConcurrent)
	assert.Equal(t, config.BrowserType, options.BrowserType)
	assert.Equal(t, config.Headless, options.Headless)
}