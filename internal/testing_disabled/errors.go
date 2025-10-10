package testing

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// TestError 测试错误类型
type TestError struct {
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Details    string                 `json:"details,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Error 实现error接口
func (e *TestError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 返回底层错误
func (e *TestError) Unwrap() error {
	if e.Details != "" {
		return fmt.Errorf(e.Details)
	}
	return nil
}

// 定义测试相关的错误
var (
	// 通用错误
	ErrTestNotFound           = errors.New("test not found")
	ErrTestExecutionNotFound  = errors.New("test execution not found")
	ErrTestResultNotFound     = errors.New("test result not found")
	ErrTestSuiteNotFound      = errors.New("test suite not found")
	ErrTestCaseNotFound       = errors.New("test case not found")
	ErrInvalidTestConfig      = errors.New("invalid test configuration")
	ErrTestExecutionFailed    = errors.New("test execution failed")
	ErrTestTimeout            = errors.New("test execution timeout")
	ErrTestCancelled          = errors.New("test execution cancelled")

	// 执行器错误
	ErrExecutorNotFound       = errors.New("executor not found")
	ErrExecutorSetupFailed    = errors.New("executor setup failed")
	ErrExecutorTeardownFailed = errors.New("executor teardown failed")
	ErrUnsupportedTestType    = errors.New("unsupported test type")

	// 调度器错误
	ErrSchedulerNotRunning    = errors.New("scheduler is not running")
	ErrSchedulerAlreadyRunning = errors.New("scheduler is already running")
	ErrQueueFull              = errors.New("test queue is full")
	ErrNoAvailableSlot        = errors.New("no available execution slot")
	ErrTestAlreadyRunning     = errors.New("test is already running")

	// 验证错误
	ErrInvalidTestSuite       = errors.New("invalid test suite")
	ErrInvalidTestCase        = errors.New("invalid test case")
	ErrInvalidExecution       = errors.New("invalid test execution")
	ErrMissingRequiredField   = errors.New("missing required field")
	ErrInvalidURL             = errors.New("invalid URL")
	ErrInvalidMethod          = errors.New("invalid HTTP method")
	ErrInvalidHeaders         = errors.New("invalid headers")

	// 依赖错误
	ErrDependencyNotFound     = errors.New("dependency not found")
	ErrDependencyFailed       = errors.New("dependency failed")
	ErrCircularDependency     = errors.New("circular dependency detected")

	// 资源错误
	ErrResourceNotFound       = errors.New("resource not found")
	ErrResourceInUse          = errors.New("resource is in use")
	ErrInsufficientResources  = errors.New("insufficient resources")

	// 配置错误
	ErrInvalidConfiguration   = errors.New("invalid configuration")
	ErrConfigurationNotFound  = errors.New("configuration not found")

	// 网络错误
	ErrNetworkTimeout         = errors.New("network timeout")
	ErrConnectionFailed       = errors.New("connection failed")
	ErrHTTPError              = errors.New("HTTP error")
	ErrSSLVerificationFailed  = errors.New("SSL verification failed")

	// 文件系统错误
	ErrFileNotFound           = errors.New("file not found")
	ErrPermissionDenied       = errors.New("permission denied")
	ErrDiskFull               = errors.New("disk full")
	ErrInvalidPath            = errors.New("invalid file path")

	// 数据库错误
	ErrDatabaseConnection     = errors.New("database connection failed")
	ErrRecordNotFound         = errors.New("record not found")
	ErrDuplicateRecord        = errors.New("duplicate record")
	ErrDatabaseConstraint     = errors.New("database constraint violation")

	// 断言错误
	ErrAssertionFailed        = errors.New("assertion failed")
	ErrExpectedValueMismatch  = errors.New("expected value mismatch")
	ErrElementNotFound        = errors.New("element not found")
	ErrElementNotVisible      = errors.New("element not visible")
	ErrElementNotClickable    = errors.New("element not clickable")

	// 时间相关错误
	ErrTimeoutExceeded        = errors.New("timeout exceeded")
	ErrInvalidDuration        = errors.New("invalid duration")
	ErrInvalidTimeFormat      = errors.New("invalid time format")

	// JSON/序列化错误
	ErrJSONParseFailed        = errors.New("JSON parse failed")
	ErrJSONMarshalFailed      = errors.New("JSON marshal failed")
	ErrInvalidJSON            = errors.New("invalid JSON")

	// 浏览器相关错误
	ErrBrowserNotStarted      = errors.New("browser not started")
	ErrPageNotCreated         = errors.New("page not created")
	ErrNavigationFailed       = errors.New("navigation failed")
	ErrScriptExecutionFailed  = errors.New("script execution failed")
	ErrScreenshotFailed       = errors.New("screenshot failed")

	// 性能测试错误
	ErrLoadTestFailed         = errors.New("load test failed")
	ErrInsufficientMetrics    = errors.New("insufficient performance metrics")
	ErrThresholdExceeded      = errors.New("performance threshold exceeded")

	// 安全错误
	ErrAuthenticationFailed   = errors.New("authentication failed")
	ErrAuthorizationFailed    = errors.New("authorization failed")
	ErrTokenExpired           = errors.New("token expired")
	ErrInvalidCredentials     = errors.New("invalid credentials")

	// 系统错误
	ErrSystemError            = errors.New("system error")
	ErrInternalError          = errors.New("internal error")
	ErrNotImplemented         = errors.New("feature not implemented")
	ErrUnsupportedOperation   = errors.New("unsupported operation")
)

// NewTestError 创建新的测试错误
func NewTestError(errorType, message string, details ...string) *TestError {
	err := &TestError{
		Type:    errorType,
		Message: message,
	}

	if len(details) > 0 {
		err.Details = details[0]
	}

	return err
}

// IsTestTypeSupported 检查测试类型是否支持
func IsTestTypeSupported(testType TestType) bool {
	supportedTypes := []TestType{
		TestTypeAPI,
		TestTypeUI,
		TestTypePerformance,
		TestTypeIntegration,
		TestTypeE2E,
	}
	for _, supportedType := range supportedTypes {
		if supportedType == testType {
			return true
		}
	}
	return false
}

// WrapError 包装现有错误
func WrapError(err error, errorType, message string) *TestError {
	return &TestError{
		Type:    errorType,
		Message: message,
		Details: err.Error(),
	}
}

// ValidateURL 验证URL格式
func ValidateURL(url string) error {
	if url == "" {
		return WrapError(ErrInvalidURL, "url", "URL cannot be empty")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return WrapError(ErrInvalidURL, "url", "URL must start with http:// or https://")
	}

	return nil
}

// ValidateHTTPMethod 验证HTTP方法
func ValidateHTTPMethod(method string) error {
	if method == "" {
		return WrapError(ErrInvalidMethod, "method", "HTTP method cannot be empty")
	}

	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return nil
		}
	}

	return WrapError(ErrInvalidMethod, "method", fmt.Sprintf("invalid HTTP method: %s", method))
}

// ValidateHeaders 验证HTTP头部
func ValidateHeaders(headers map[string]string) error {
	for key, value := range headers {
		if key == "" {
			return WrapError(ErrInvalidHeaders, "headers", "header key cannot be empty")
		}
		if strings.Contains(key, " ") {
			return WrapError(ErrInvalidHeaders, "headers", fmt.Sprintf("invalid header key: %s", key))
		}
		if value == "" {
			return WrapError(ErrInvalidHeaders, "headers", fmt.Sprintf("header value cannot be empty for key: %s", key))
		}
	}
	return nil
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是包装的TestError
	var testErr *TestError
	if errors.As(err, &testErr) {
		switch testErr.Type {
		case "network_timeout",
			 "connection_failed",
			 "http_error",
			 "temporary_error":
			return true
		default:
			return false
		}
	}

	// 检查特定错误类型
	switch {
	case errors.Is(err, ErrNetworkTimeout),
		 errors.Is(err, ErrConnectionFailed),
		 errors.Is(err, ErrHTTPError),
		 errors.Is(err, context.DeadlineExceeded):
		return true
	default:
		return false
	}
}

// IsCriticalError 判断错误是否为关键错误
func IsCriticalError(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否是包装的TestError
	var testErr *TestError
	if errors.As(err, &testErr) {
		switch testErr.Type {
		case "system_error",
			 "internal_error",
			 "configuration_error",
			 "authentication_failed":
			return true
		default:
			return false
		}
	}

	// 检查特定错误类型
	switch {
	case errors.Is(err, ErrSystemError),
		 errors.Is(err, ErrInternalError),
		 errors.Is(err, ErrInvalidConfiguration),
		 errors.Is(err, ErrAuthenticationFailed):
		return true
	default:
		return false
	}
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) string {
	if err == nil {
		return ""
	}

	// 检查是否是TestError
	var testErr *TestError
	if errors.As(err, &testErr) {
		if testErr.Code != "" {
			return testErr.Code
		}
		return testErr.Type
	}

	return err.Error()
}

// FormatError 格式化错误信息
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	// 检查是否是TestError
	var testErr *TestError
	if errors.As(err, &testErr) {
		if testErr.Details != "" {
			return fmt.Sprintf("[%s] %s: %s - %s", testErr.Code, testErr.Type, testErr.Message, testErr.Details)
		}
		return fmt.Sprintf("[%s] %s: %s", testErr.Code, testErr.Type, testErr.Message)
	}

	return err.Error()
}