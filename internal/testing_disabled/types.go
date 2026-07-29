package testing

import (
	"context"
	"time"

	"law-oa-go/internal/models"
)

// TestExecutor 测试执行器接口
type TestExecutor interface {
	// Execute 执行单个测试
	Execute(ctx context.Context, test *TestCase, executionCtx *ExecutionContext) (*TestResult, error)

	// ExecuteBatch 批量执行测试
	ExecuteBatch(ctx context.Context, tests []*TestCase, executionCtx *ExecutionContext) ([]*TestResult, error)

	// GetExecutorType 获取执行器类型
	GetExecutorType() TestType

	// Setup 执行前准备
	Setup(ctx context.Context, executionCtx *ExecutionContext) error

	// Teardown 执行后清理
	Teardown(ctx context.Context, executionCtx *ExecutionContext) error
}

// ExecutionContext 执行上下文
type ExecutionContext struct {
	// 基础信息
	ExecutionID string
	Environment string
	Variables   map[string]interface{}
	Headers     map[string]string
	Timeout     time.Duration

	// 路径信息
	WorkingDir string
	BaseURL    string

	// 测试配置
	Parallel   bool
	Retries    int
	RetryDelay time.Duration

	// 日志和监控
	Logger  TestLogger
	Metrics TestMetrics

	// 上下文
	Context context.Context
	Cancel  context.CancelFunc
}

// TestCase 测试用例定义
type TestCase struct {
	ID          string                 `json:"id"`
	SuiteID     string                 `json:"suite_id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        TestType               `json:"type"`
	Method      string                 `json:"method,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Body        interface{}            `json:"body,omitempty"`
	Steps       []TestStep             `json:"steps,omitempty"`
	Setup       []TestStep             `json:"setup,omitempty"`
	Teardown    []TestStep             `json:"teardown,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Expected    *TestExpected          `json:"expected,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TestStep 测试步骤
type TestStep struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Action     string                 `json:"action"`
	Target     string                 `json:"target,omitempty"`
	Value      interface{}            `json:"value,omitempty"`
	Expected   interface{}            `json:"expected,omitempty"`
	Wait       time.Duration          `json:"wait,omitempty"`
	Screenshot bool                   `json:"screenshot,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TestExpected 测试期望
type TestExpected struct {
	Status       int                    `json:"status,omitempty"`
	Body         interface{}            `json:"body,omitempty"`
	Headers      map[string]string      `json:"headers,omitempty"`
	BodyContains []string               `json:"body_contains,omitempty"`
	ResponseTime time.Duration          `json:"response_time,omitempty"`
	ContentType  string                 `json:"content_type,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TestResult 测试结果
type TestResult struct {
	// 基础信息
	ID          string     `json:"id"`
	TestCaseID  string     `json:"test_case_id"`
	ExecutionID string     `json:"execution_id"`
	Name        string     `json:"name"`
	Type        TestType   `json:"type"`
	Status      TestStatus `json:"status"`

	// 执行信息
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`

	// 结果信息
	Passed     bool             `json:"passed"`
	Error      *TestError       `json:"error,omitempty"`
	Assertions []*TestAssertion `json:"assertions,omitempty"`

	// 附加信息
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Screenshots []string               `json:"screenshots,omitempty"`
	Logs        []string               `json:"logs,omitempty"`
}

// TestAssertion 测试断言
type TestAssertion struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Actual       interface{}            `json:"actual"`
	Expected     interface{}            `json:"expected"`
	Passed       bool                   `json:"passed"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TestLogger 测试日志接口
type TestLogger interface {
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// TestMetrics 测试指标接口
type TestMetrics interface {
	RecordExecution(testType TestType, duration time.Duration, status TestStatus)
	RecordAssertion(testType TestType, passed bool)
	RecordError(testType TestType, errorType string)
	RecordCustom(metric string, value float64, tags map[string]string)
}

// TestSuiteConfig 测试套件配置
type TestSuiteConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Environment string                 `json:"environment"`
	Timeout     time.Duration          `json:"timeout"`
	Parallel    bool                   `json:"parallel"`
	MaxWorkers  int                    `json:"max_workers"`
	Retries     int                    `json:"retries"`
	RetryDelay  time.Duration          `json:"retry_delay"`
	Variables   map[string]interface{} `json:"variables"`
	Headers     map[string]string      `json:"headers"`
	Setup       []TestStep             `json:"setup,omitempty"`
	Teardown    []TestStep             `json:"teardown,omitempty"`
	BeforeAll   []TestStep             `json:"before_all,omitempty"`
	AfterAll    []TestStep             `json:"after_all,omitempty"`
}

// TestType 测试类型
type TestType string

const (
	TestTypeAPI         TestType = "api"
	TestTypeUI          TestType = "ui"
	TestTypePerformance TestType = "performance"
	TestTypeIntegration TestType = "integration"
	TestTypeE2E         TestType = "e2e"
)

// TestStatus 测试状态
type TestStatus string

const (
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
	TestStatusError   TestStatus = "error"
	TestStatusTimeout TestStatus = "timeout"
	TestStatusPending TestStatus = "pending"
	TestStatusRunning TestStatus = "running"
)

// ToModelStatus 转换为 models.TestExecutionStatus
func (s TestStatus) ToModelStatus() models.TestExecutionStatus {
	switch s {
	case TestStatusPassed:
		return models.TestStatusCompleted
	case TestStatusFailed, TestStatusError, TestStatusTimeout:
		return models.TestStatusFailed
	case TestStatusSkipped:
		return models.TestStatusCancelled
	case TestStatusPending:
		return models.TestStatusPending
	case TestStatusRunning:
		return models.TestStatusRunning
	default:
		return models.TestStatusPending
	}
}

// FromModelStatus 从 models.TestExecutionStatus 转换
func FromModelStatus(status models.TestExecutionStatus) TestStatus {
	switch status {
	case models.TestStatusCompleted:
		return TestStatusPassed
	case models.TestStatusFailed:
		return TestStatusFailed
	case models.TestStatusCancelled:
		return TestStatusSkipped
	case models.TestStatusPending:
		return TestStatusPending
	case models.TestStatusRunning:
		return TestStatusRunning
	default:
		return TestStatusPending
	}
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// 网络指标
	ResponseTime    time.Duration `json:"response_time"`
	TimeToFirstByte time.Duration `json:"time_to_first_byte"`
	DomInteractive  time.Duration `json:"dom_interactive"`

	// 资源指标
	ResourceCount     int   `json:"resource_count"`
	TotalTransferSize int64 `json:"total_transfer_size"`
	ContentSize       int64 `json:"content_size"`

	// 浏览器指标
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  int64   `json:"memory_usage"`
	NetworkUsage int64   `json:"network_usage"`

	// 自定义指标
	CustomMetrics map[string]interface{} `json:"custom_metrics"`
}

// UIMetrics UI指标
type UIMetrics struct {
	// 页面信息
	URL   string `json:"url"`
	Title string `json:"title"`

	// 可见性指标
	ViewportSize map[string]int `json:"viewport_size"`
	PageSize     map[string]int `json:"page_size"`

	// 交互指标
	ElementsFound int `json:"elements_found"`
	ClicksCount   int `json:"clicks_count"`

	// 性能指标
	RenderTime time.Duration `json:"render_time"`
	LoadTime   time.Duration `json:"load_time"`

	// 可访问性指标
	AccessibilityIssues []string `json:"accessibility_issues"`
}

// TestExecutorOptions 执行器选项
type TestExecutorOptions struct {
	// 基础配置
	Timeout    time.Duration
	Retries    int
	RetryDelay time.Duration

	// 并发配置
	MaxConcurrent int
	QueueSize     int

	// 浏览器配置（UI测试用）
	BrowserType string
	Headless    bool
	WindowSize  map[string]int

	// HTTP配置（API测试用）
	UserAgent       string
	FollowRedirects bool
	VerifySSL       bool

	// 性能测试配置
	RampUpTime time.Duration
	Duration   time.Duration
	ThinkTime  time.Duration

	// 其他配置
	Screenshots bool
	Videos      bool
	NetworkLogs bool
	ConsoleLogs bool
}

// DefaultExecutorOptions 默认执行器选项
func DefaultExecutorOptions() *TestExecutorOptions {
	return &TestExecutorOptions{
		Timeout:         30 * time.Second,
		Retries:         3,
		RetryDelay:      1 * time.Second,
		MaxConcurrent:   10,
		QueueSize:       100,
		BrowserType:     "chromium",
		Headless:        true,
		WindowSize:      map[string]int{"width": 1280, "height": 720},
		UserAgent:       "Law-OA-Go-Testing-Agent/1.0",
		FollowRedirects: true,
		VerifySSL:       false,
		Screenshots:     true,
		Videos:          false,
		NetworkLogs:     true,
		ConsoleLogs:     true,
	}
}

// TestRepository 测试仓库接口
type TestRepository interface {
	// 测试套件管理
	CreateTestSuite(ctx context.Context, suite *models.TestSuite) error
	GetTestSuite(ctx context.Context, id string) (*models.TestSuite, error)
	UpdateTestSuite(ctx context.Context, suite *models.TestSuite) error
	DeleteTestSuite(ctx context.Context, id string) error
	ListTestSuites(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.TestSuite, int64, error)

	// 测试用例管理
	CreateTestCase(ctx context.Context, testCase *TestCase) error
	GetTestCase(ctx context.Context, id string) (*TestCase, error)
	UpdateTestCase(ctx context.Context, testCase *TestCase) error
	DeleteTestCase(ctx context.Context, id string) error
	ListTestCases(ctx context.Context, filter *TestCaseFilter) ([]*TestCase, error)

	// 测试执行管理
	CreateTestExecution(ctx context.Context, execution *models.TestExecution) error
	GetTestExecution(ctx context.Context, id string) (*models.TestExecution, error)
	UpdateTestExecution(ctx context.Context, execution *models.TestExecution) error
	ListTestExecutions(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.TestExecution, int64, error)

	// 测试结果管理
	CreateTestResult(ctx context.Context, result *models.TestResult) error
	GetTestResult(ctx context.Context, id string) (*models.TestResult, error)
	UpdateTestResult(ctx context.Context, result *models.TestResult) error
	ListTestResults(ctx context.Context, executionID string) ([]*models.TestResult, error)

	// 清理操作
	CleanupOldExecutions(ctx context.Context, cutoff time.Time) error
}

// 扩展原有的TestLogger接口
type TestLoggerExtended interface {
	TestLogger
	// 测试专用日志
	LogTestStart(executionID string, test *TestCase)
	LogTestEnd(executionID string, result *TestResult)
	LogTestError(executionID string, err error)
	// 执行日志
	LogExecutionStart(execution *models.TestExecution)
	LogExecutionEnd(execution *models.TestExecution)
}

// 扩展原有的TestMetrics接口
type TestMetricsExtended interface {
	TestMetrics
	// 计数器
	IncrementCounter(name string, labels map[string]string)
	IncrementCounterBy(name string, value float64, labels map[string]string)
	// 直方图
	RecordHistogram(name string, value float64, labels map[string]string)
	// 仪表盘
	SetGauge(name string, value float64, labels map[string]string)
	// 计时器
	RecordTimer(name string, duration time.Duration, labels map[string]string)
	// 测试专用指标
	RecordTestExecution(testType string, status string, duration time.Duration)
	RecordTestResult(testType string, result string, assertions int)
}

// 为了兼容性，添加模型别名
type TestCaseAlias = models.TestSuite
type TestResultAlias = models.TestResult
type TestExecutionAlias = models.TestExecution
type TestSuiteAlias = models.TestSuite
