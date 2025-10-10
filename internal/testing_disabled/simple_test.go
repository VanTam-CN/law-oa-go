package testing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

// TestBasicTypes 测试基本类型
func TestBasicTypes(t *testing.T) {
	// 测试TestType
	assert.Equal(t, TestType("api"), TestTypeAPI)
	assert.Equal(t, TestType("ui"), TestTypeUI)
	assert.Equal(t, TestType("performance"), TestTypePerformance)
	assert.Equal(t, TestType("integration"), TestTypeIntegration)
	assert.Equal(t, TestType("e2e"), TestTypeE2E)

	// 测试TestStatus
	assert.Equal(t, TestStatus("passed"), TestStatusPassed)
	assert.Equal(t, TestStatus("failed"), TestStatusFailed)
	assert.Equal(t, TestStatus("skipped"), TestStatusSkipped)
	assert.Equal(t, TestStatus("error"), TestStatusError)
	assert.Equal(t, TestStatus("timeout"), TestStatusTimeout)
	assert.Equal(t, TestStatus("pending"), TestStatusPending)
	assert.Equal(t, TestStatus("running"), TestStatusRunning)
}

// TestTestCase 测试测试用例
func TestTestCase(t *testing.T) {
	testCase := &TestCase{
		ID:          "test1",
		Name:        "Test API Endpoint",
		Description: "Test the GET /api/users endpoint",
		Type:        TestTypeAPI,
		Method:      "GET",
		URL:         "https://api.example.com/users",
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Authorization": "Bearer token123",
		},
		Timeout: 30 * time.Second,
		Expected: &TestExpected{
			Status:       200,
			ContentType: "application/json",
			BodyContains: []string{"users", "data"},
		},
	}

	assert.Equal(t, "test1", testCase.ID)
	assert.Equal(t, "Test API Endpoint", testCase.Name)
	assert.Equal(t, TestTypeAPI, testCase.Type)
	assert.Equal(t, "GET", testCase.Method)
	assert.Equal(t, "https://api.example.com/users", testCase.URL)
	assert.Equal(t, 30*time.Second, testCase.Timeout)
	assert.Equal(t, 200, testCase.Expected.Status)
}

// TestTestResult 测试测试结果
func TestTestResult(t *testing.T) {
	result := &TestResult{
		ID:          "result1",
		TestCaseID:  "test1",
		ExecutionID: "exec1",
		Name:        "Test API Endpoint",
		Type:        TestTypeAPI,
		Status:      TestStatusPassed,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(5 * time.Second),
		Duration:    5 * time.Second,
		Passed:      true,
		Metadata:    make(map[string]interface{}),
		Assertions: []*TestAssertion{
			{
				Name:     "status_code",
				Type:     "status",
				Actual:   200,
				Expected: 200,
				Passed:   true,
			},
			{
				Name:     "response_time",
				Type:     "performance",
				Actual:   1.2,
				Expected: 2.0,
				Passed:   true,
			},
		},
	}

	assert.Equal(t, "result1", result.ID)
	assert.Equal(t, "test1", result.TestCaseID)
	assert.Equal(t, TestStatusPassed, result.Status)
	assert.True(t, result.Passed)
	assert.Len(t, result.Assertions, 2)
	assert.True(t, result.Assertions[0].Passed)
}

// TestExecutionContext 测试执行上下文
func TestExecutionContext(t *testing.T) {
	ctx := &ExecutionContext{
		ExecutionID: "exec1",
		Environment: "test",
		Variables: map[string]interface{}{
			"username": "testuser",
			"password": "testpass",
			"timeout":  30,
		},
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Timeout:  30 * time.Second,
		BaseURL:  "https://api.example.com",
		Parallel: true,
		Retries:  3,
	}

	assert.Equal(t, "exec1", ctx.ExecutionID)
	assert.Equal(t, "test", ctx.Environment)
	assert.Equal(t, "testuser", ctx.Variables["username"])
	assert.Equal(t, "application/json", ctx.Headers["Content-Type"])
	assert.Equal(t, 30*time.Second, ctx.Timeout)
	assert.True(t, ctx.Parallel)
	assert.Equal(t, 3, ctx.Retries)
}

// TestTestError 测试错误类型
func TestTestError(t *testing.T) {
	// 测试简单错误
	err := NewTestError("validation_error", "Invalid URL format")
	assert.Equal(t, "validation_error: Invalid URL format", err.Error())
	assert.Equal(t, "validation_error", err.Type)
	assert.Equal(t, "Invalid URL format", err.Message)

	// 测试带详细信息的错误
	err2 := NewTestError("network_error", "Connection failed", "Timeout after 30 seconds")
	assert.Equal(t, "network_error: Connection failed - Timeout after 30 seconds", err2.Error())
	assert.Equal(t, "Timeout after 30 seconds", err2.Details)

	// 测试错误包装
	baseErr := NewTestError("base_error", "Base error")
	wrappedErr := WrapError(baseErr, "wrapper_error", "Wrapped error")
	assert.Equal(t, "wrapper_error: Wrapped error - base_error: Base error", wrappedErr.Error())
	assert.Equal(t, "wrapper_error", wrappedErr.Type)
}

// TestErrorValidation 测试错误验证函数
func TestErrorValidation(t *testing.T) {
	// 测试URL验证
	err := ValidateURL("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "URL cannot be empty")

	err = ValidateURL("ftp://example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must start with http:// or https://")

	err = ValidateURL("https://example.com")
	assert.NoError(t, err)

	// 测试HTTP方法验证
	err = ValidateHTTPMethod("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP method cannot be empty")

	err = ValidateHTTPMethod("INVALID")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid HTTP method")

	err = ValidateHTTPMethod("GET")
	assert.NoError(t, err)

	// 测试头部验证
	err = ValidateHeaders(map[string]string{
		"": "value",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "header key cannot be empty")

	err = ValidateHeaders(map[string]string{
		"key with spaces": "value",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid header key")

	err = ValidateHeaders(map[string]string{
		"valid": "value",
	})
	assert.NoError(t, err)
}

// TestTestTypeSupport 测试测试类型支持检查
func TestTestTypeSupport(t *testing.T) {
	assert.True(t, IsTestTypeSupported(TestTypeAPI))
	assert.True(t, IsTestTypeSupported(TestTypeUI))
	assert.True(t, IsTestTypeSupported(TestTypePerformance))
	assert.True(t, IsTestTypeSupported(TestTypeIntegration))
	assert.True(t, IsTestTypeSupported(TestTypeE2E))
	assert.False(t, IsTestTypeSupported("unsupported"))
}

// TestDefaultExecutorOptions 测试默认执行器选项
func TestDefaultExecutorOptions(t *testing.T) {
	options := DefaultExecutorOptions()

	assert.Equal(t, 30*time.Second, options.Timeout)
	assert.Equal(t, 3, options.Retries)
	assert.Equal(t, 1*time.Second, options.RetryDelay)
	assert.Equal(t, 10, options.MaxConcurrent)
	assert.Equal(t, 100, options.QueueSize)
	assert.Equal(t, "chromium", options.BrowserType)
	assert.True(t, options.Headless)
	assert.Equal(t, 1280, options.WindowSize["width"])
	assert.Equal(t, 720, options.WindowSize["height"])
	assert.Contains(t, options.UserAgent, "Law-OA-Go")
	assert.True(t, options.FollowRedirects)
	assert.False(t, options.VerifySSL)
	assert.True(t, options.Screenshots)
	assert.False(t, options.Videos)
	assert.True(t, options.NetworkLogs)
	assert.True(t, options.ConsoleLogs)
}

// TestBasicExecutor 测试基础执行器功能
func TestBasicExecutor(t *testing.T) {
	// 测试创建测试结果
	result := &TestResult{
		ID:          "result1",
		TestCaseID:  "test1",
		ExecutionID: "exec1",
		Name:        "Test API Endpoint",
		Type:        TestTypeAPI,
		Status:      TestStatusPending,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(5 * time.Second),
		Duration:    5 * time.Second,
		Passed:      false,
	}

	assert.Equal(t, "result1", result.ID)
	assert.Equal(t, "test1", result.TestCaseID)
	assert.Equal(t, "exec1", result.ExecutionID)
	assert.Equal(t, TestStatusPending, result.Status)
	assert.False(t, result.Passed)
}