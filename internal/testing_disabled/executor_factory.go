package testing

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExecutorFactory 执行器工厂接口
type ExecutorFactory interface {
	// CreateExecutor 创建执行器
	CreateExecutor(testType TestType, options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) (TestExecutor, error)

	// GetSupportedTypes 获取支持的测试类型
	GetSupportedTypes() []TestType

	// RegisterExecutor 注册自定义执行器
	RegisterExecutor(testType TestType, creator ExecutorCreator) error
}

// ExecutorCreator 执行器创建函数
type ExecutorCreator func(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) TestExecutor

// DefaultExecutorFactory 默认执行器工厂
type DefaultExecutorFactory struct {
	creators map[TestType]ExecutorCreator
}

// NewDefaultExecutorFactory 创建默认执行器工厂
func NewDefaultExecutorFactory() ExecutorFactory {
	factory := &DefaultExecutorFactory{
		creators: make(map[TestType]ExecutorCreator),
	}

	// 注册内置执行器
	factory.registerBuiltinExecutors()

	return factory
}

// registerBuiltinExecutors 注册内置执行器
func (f *DefaultExecutorFactory) registerBuiltinExecutors() {
	// 注册API执行器
	f.creators[TestTypeAPI] = func(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) TestExecutor {
		return NewAPIExecutor(options, logger, metrics)
	}

	// 注册UI执行器
	f.creators[TestTypeUI] = func(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) TestExecutor {
		return NewUIExecutor(options, logger, metrics)
	}

	// 注册性能执行器
	f.creators[TestTypePerformance] = func(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) TestExecutor {
		return NewPerformanceExecutor(options, logger, metrics)
	}

	// 注册集成执行器
	f.creators[TestTypeIntegration] = func(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) TestExecutor {
		return NewIntegrationExecutor(options, logger, metrics)
	}

	// 注册E2E执行器（使用集成执行器）
	f.creators[TestTypeE2E] = func(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) TestExecutor {
		return NewIntegrationExecutor(options, logger, metrics)
	}
}

// CreateExecutor 创建执行器
func (f *DefaultExecutorFactory) CreateExecutor(testType TestType, options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) (TestExecutor, error) {
	creator, exists := f.creators[testType]
	if !exists {
		return nil, fmt.Errorf("unsupported test type: %s", testType)
	}

	executor := creator(options, logger, metrics)
	return executor, nil
}

// GetSupportedTypes 获取支持的测试类型
func (f *DefaultExecutorFactory) GetSupportedTypes() []TestType {
	types := make([]TestType, 0, len(f.creators))
	for testType := range f.creators {
		types = append(types, testType)
	}
	return types
}

// RegisterExecutor 注册自定义执行器
func (f *DefaultExecutorFactory) RegisterExecutor(testType TestType, creator ExecutorCreator) error {
	if _, exists := f.creators[testType]; exists {
		return fmt.Errorf("test type %s already registered", testType)
	}

	f.creators[testType] = creator
	return nil
}

// ExecutorManager 执行器管理器
type ExecutorManager struct {
	factory   ExecutorFactory
	executors map[TestType]TestExecutor
	options   *TestExecutorOptions
	logger    TestLogger
	metrics   TestMetrics
}

// NewExecutorManager 创建执行器管理器
func NewExecutorManager(factory ExecutorFactory, options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) *ExecutorManager {
	return &ExecutorManager{
		factory:   factory,
		executors: make(map[TestType]TestExecutor),
		options:   options,
		logger:    logger,
		metrics:   metrics,
	}
}

// GetExecutor 获取指定类型的执行器
func (m *ExecutorManager) GetExecutor(testType TestType) (TestExecutor, error) {
	executor, exists := m.executors[testType]
	if exists {
		return executor, nil
	}

	// 创建新的执行器
	newExecutor, err := m.factory.CreateExecutor(testType, m.options, m.logger, m.metrics)
	if err != nil {
		return nil, err
	}

	// 缓存执行器
	m.executors[testType] = newExecutor
	return newExecutor, nil
}

// GetOrCreateExecutor 获取或创建执行器
func (m *ExecutorManager) GetOrCreateExecutor(testType TestType) (TestExecutor, error) {
	return m.GetExecutor(testType)
}

// ReleaseExecutor 释放执行器资源
func (m *ExecutorManager) ReleaseExecutor(testType TestType) error {
	executor, exists := m.executors[testType]
	if !exists {
		return nil
	}

	// 如果执行器实现了清理接口，调用清理
	if cleaner, ok := executor.(interface{ Cleanup() error }); ok {
		if err := cleaner.Cleanup(); err != nil {
			return err
		}
	}

	delete(m.executors, testType)
	return nil
}

// ReleaseAllExecutors 释放所有执行器资源
func (m *ExecutorManager) ReleaseAllExecutors() error {
	for testType := range m.executors {
		if err := m.ReleaseExecutor(testType); err != nil {
			return err
		}
	}
	return nil
}

// GetSupportedTypes 获取支持的测试类型
func (m *ExecutorManager) GetSupportedTypes() []TestType {
	return m.factory.GetSupportedTypes()
}

// RegisterCustomExecutor 注册自定义执行器
func (m *ExecutorManager) RegisterCustomExecutor(testType TestType, creator ExecutorCreator) error {
	return m.factory.RegisterExecutor(testType, creator)
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	// 基础配置
	DefaultTimeout    Duration  `json:"default_timeout"`
	DefaultRetries    int       `json:"default_retries"`
	DefaultRetryDelay Duration  `json:"default_retry_delay"`
	MaxConcurrent     int       `json:"max_concurrent"`

	// 浏览器配置
	BrowserType       string            `json:"browser_type"`
	Headless          bool              `json:"headless"`
	WindowSize        map[string]int    `json:"window_size"`
	UserAgent         string            `json:"user_agent"`

	// 性能测试配置
	RampUpTime        Duration          `json:"ramp_up_time"`
	TestDuration      Duration          `json:"test_duration"`
	ThinkTime         Duration          `json:"think_time"`

	// 录制配置
	EnableScreenshots bool              `json:"enable_screenshots"`
	EnableVideos      bool              `json:"enable_videos"`
	EnableTracing     bool              `json:"enable_tracing"`

	// 网络配置
	FollowRedirects   bool              `json:"follow_redirects"`
	VerifySSL         bool              `json:"verify_ssl"`
	ProxyURL          string            `json:"proxy_url"`

	// 输出目录
	OutputDirectories map[string]string `json:"output_directories"`

	// 自定义配置
	Custom            map[string]interface{} `json:"custom"`
}

// Duration 自定义时间类型，支持JSON序列化
type Duration time.Duration

// UnmarshalJSON 实现JSON反序列化
func (d *Duration) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case string:
		duration, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		*d = Duration(duration)
	case float64:
		*d = Duration(time.Duration(v) * time.Second)
	case int:
		*d = Duration(time.Duration(v) * time.Second)
	default:
		return fmt.Errorf("invalid duration type: %T", value)
	}

	return nil
}

// MarshalJSON 实现JSON序列化
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// ToDuration 转换为标准时间类型
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

// NewDefaultExecutorConfig 创建默认执行器配置
func NewDefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		DefaultTimeout:    Duration(30 * time.Second),
		DefaultRetries:    3,
		DefaultRetryDelay: Duration(1 * time.Second),
		MaxConcurrent:     10,
		BrowserType:       "chromium",
		Headless:          true,
		WindowSize:        map[string]int{"width": 1280, "height": 720},
		UserAgent:         "Law-OA-Go-Testing-Agent/1.0",
		RampUpTime:        Duration(10 * time.Second),
		TestDuration:      Duration(60 * time.Second),
		ThinkTime:         Duration(1 * time.Second),
		EnableScreenshots: true,
		EnableVideos:      false,
		EnableTracing:     true,
		FollowRedirects:   true,
		VerifySSL:         false,
		OutputDirectories: map[string]string{
			"screenshots": "screenshots",
			"videos":      "videos",
			"traces":      "traces",
			"logs":        "logs",
		},
		Custom: make(map[string]interface{}),
	}
}

// ToTestExecutorOptions 转换为测试执行器选项
func (config *ExecutorConfig) ToTestExecutorOptions() *TestExecutorOptions {
	return &TestExecutorOptions{
		Timeout:         config.DefaultTimeout.ToDuration(),
		Retries:         config.DefaultRetries,
		RetryDelay:      config.DefaultRetryDelay.ToDuration(),
		MaxConcurrent:   config.MaxConcurrent,
		QueueSize:       100,
		BrowserType:     config.BrowserType,
		Headless:        config.Headless,
		WindowSize:      config.WindowSize,
		UserAgent:       config.UserAgent,
		FollowRedirects: config.FollowRedirects,
		VerifySSL:       config.VerifySSL,
		RampUpTime:      config.RampUpTime.ToDuration(),
		Duration:        config.TestDuration.ToDuration(),
		ThinkTime:       config.ThinkTime.ToDuration(),
		Screenshots:     config.EnableScreenshots,
		Videos:          config.EnableVideos,
		NetworkLogs:     true,
		ConsoleLogs:     true,
	}
}

// 全局执行器工厂实例
var GlobalExecutorFactory ExecutorFactory

// init 初始化全局执行器工厂
func init() {
	GlobalExecutorFactory = NewDefaultExecutorFactory()
}

// CreateExecutor 创建执行器（使用全局工厂）
func CreateExecutor(testType TestType, options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) (TestExecutor, error) {
	return GlobalExecutorFactory.CreateExecutor(testType, options, logger, metrics)
}

// GetSupportedTestTypes 获取支持的测试类型（使用全局工厂）
func GetSupportedTestTypes() []TestType {
	return GlobalExecutorFactory.GetSupportedTypes()
}

// RegisterExecutor 注册执行器（使用全局工厂）
func RegisterExecutor(testType TestType, creator ExecutorCreator) error {
	return GlobalExecutorFactory.RegisterExecutor(testType, creator)
}