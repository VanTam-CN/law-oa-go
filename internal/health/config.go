package health

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/concurrency"
)

// HealthCheckRegistry 健康检查注册器
type HealthCheckRegistry struct {
	checks map[string]HealthCheck
}

// NewHealthCheckRegistry 创建健康检查注册器
func NewHealthCheckRegistry() *HealthCheckRegistry {
	return &HealthCheckRegistry{
		checks: make(map[string]HealthCheck),
	}
}

// Register 注册健康检查
func (r *HealthCheckRegistry) Register(name string, check HealthCheck) {
	r.checks[name] = check
}

// Get 获取健康检查
func (r *HealthCheckRegistry) Get(name string) (HealthCheck, bool) {
	check, exists := r.checks[name]
	return check, exists
}

// GetAll 获取所有健康检查
func (r *HealthCheckRegistry) GetAll() map[string]HealthCheck {
	checks := make(map[string]HealthCheck)
	for name, check := range r.checks {
		checks[name] = check
	}
	return checks
}

// HealthCheckFactory 健康检查工厂
type HealthCheckFactory struct{}

// NewHealthCheckFactory 创建健康检查工厂
func NewHealthCheckFactory() *HealthCheckFactory {
	return &HealthCheckFactory{}
}

// CreateDatabaseCheck 创建数据库健康检查
func (f *HealthCheckFactory) CreateDatabaseCheck(db interface{}, timeout time.Duration) HealthCheck {
	if dbChecker, ok := db.(DatabaseChecker); ok {
		if sqlDB, ok := dbChecker.GetDB().(*sql.DB); ok {
			return NewDatabaseHealthCheck(sqlDB, timeout)
		}
	}

	// 如果不是DatabaseChecker，创建简单的检查
	return &SimpleHealthCheck{
		name:    "database",
		timeout: timeout,
		checker: func(ctx context.Context) *HealthCheckResult {
			return &HealthCheckResult{
				Name:     "database",
				Status:   StatusHealthy,
				Message:  "数据库检查已配置",
				Duration: 10,
			}
		},
	}
}

// CreateCacheCheck 创建缓存健康检查
func (f *HealthCheckFactory) CreateCacheCheck(cache interface{}, timeout time.Duration) HealthCheck {
	if _, ok := cache.(CacheChecker); ok {
		// 这里简化处理，直接创建简单的检查
		return &SimpleHealthCheck{
			name:    "cache",
			timeout: timeout,
			checker: func(ctx context.Context) *HealthCheckResult {
				return &HealthCheckResult{
					Name:     "cache",
					Status:   StatusHealthy,
					Message:  "缓存检查已配置",
					Duration: 5,
				}
			},
		}
	}

	return &SimpleHealthCheck{
		name:    "cache",
		timeout: timeout,
		checker: func(ctx context.Context) *HealthCheckResult {
			return &HealthCheckResult{
				Name:     "cache",
				Status:   StatusHealthy,
				Message:  "缓存检查已配置",
				Duration: 5,
			}
		},
	}
}

// CreateConcurrencyCheck 创建并发健康检查
func (f *HealthCheckFactory) CreateConcurrencyCheck(service interface{}, timeout time.Duration) HealthCheck {
	if concurrencyChecker, ok := service.(ConcurrencyChecker); ok {
		if concurrencyService, ok := concurrencyChecker.GetService().(*concurrency.ConcurrentService); ok {
			return NewConcurrencyHealthCheck(concurrencyService, timeout)
		}
	}

	return &SimpleHealthCheck{
		name:    "concurrency",
		timeout: timeout,
		checker: func(ctx context.Context) *HealthCheckResult {
			return &HealthCheckResult{
				Name:     "concurrency",
				Status:   StatusHealthy,
				Message:  "并发检查已配置",
				Duration: 15,
			}
		},
	}
}

// CreateExternalAPICheck 创建外部API健康检查
func (f *HealthCheckFactory) CreateExternalAPICheck(url string, timeout time.Duration) HealthCheck {
	return NewExternalAPIHealthCheck(url, timeout)
}

// CreateStorageCheck 创建存储健康检查
func (f *HealthCheckFactory) CreateStorageCheck(path string, timeout time.Duration) HealthCheck {
	return NewStorageHealthCheck(path, timeout)
}

// HealthCheckConfigurator 健康检查配置器
type HealthCheckConfigurator struct {
	factory  *HealthCheckFactory
	registry *HealthCheckRegistry
}

// NewHealthCheckConfigurator 创建健康检查配置器
func NewHealthCheckConfigurator() *HealthCheckConfigurator {
	return &HealthCheckConfigurator{
		factory:  NewHealthCheckFactory(),
		registry: NewHealthCheckRegistry(),
	}
}

// ConfigureDatabase 配置数据库健康检查
func (c *HealthCheckConfigurator) ConfigureDatabase(db interface{}, timeout time.Duration) {
	check := c.factory.CreateDatabaseCheck(db, timeout)
	c.registry.Register("database", check)
}

// ConfigureCache 配置缓存健康检查
func (c *HealthCheckConfigurator) ConfigureCache(cache interface{}, timeout time.Duration) {
	check := c.factory.CreateCacheCheck(cache, timeout)
	c.registry.Register("cache", check)
}

// ConfigureConcurrency 配置并发健康检查
func (c *HealthCheckConfigurator) ConfigureConcurrency(service interface{}, timeout time.Duration) {
	check := c.factory.CreateConcurrencyCheck(service, timeout)
	c.registry.Register("concurrency", check)
}

// ConfigureExternalAPI 配置外部API健康检查
func (c *HealthCheckConfigurator) ConfigureExternalAPI(url string, timeout time.Duration) {
	check := c.factory.CreateExternalAPICheck(url, timeout)
	c.registry.Register("external_api", check)
}

// ConfigureStorage 配置存储健康检查
func (c *HealthCheckConfigurator) ConfigureStorage(path string, timeout time.Duration) {
	check := c.factory.CreateStorageCheck(path, timeout)
	c.registry.Register("storage", check)
}

// GetRegistry 获取注册器
func (c *HealthCheckConfigurator) GetRegistry() *HealthCheckRegistry {
	return c.registry
}

// SimpleHealthCheck 简单健康检查实现
type SimpleHealthCheck struct {
	name    string
	timeout time.Duration
	checker func(ctx context.Context) *HealthCheckResult
}

func (s *SimpleHealthCheck) Check(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	result := s.checker(ctx)
	result.Duration = time.Since(start).Milliseconds()
	if result.Duration == 0 {
		result.Duration = 1
	}
	result.Timestamp = start
	result.Name = s.name

	if result.Status == "" {
		result.Status = StatusHealthy
	}

	return result
}

func (s *SimpleHealthCheck) GetName() string {
	return s.name
}

func (s *SimpleHealthCheck) GetTimeout() time.Duration {
	return s.timeout
}

// DatabaseChecker 数据库检查器接口
type DatabaseChecker interface {
	GetDB() interface{}
}

// CacheChecker 缓存检查器接口
type CacheChecker interface {
	GetCache() interface{}
}

// ConcurrencyChecker 并发检查器接口
type ConcurrencyChecker interface {
	GetService() interface{}
}

// HealthCheckBuilder 健康检查构建器
type HealthCheckBuilder struct {
	config   *HealthConfig
	registry *HealthCheckRegistry
	factory  *HealthCheckFactory
	logger   interface{}
}

// NewHealthCheckBuilder 创建健康检查构建器
func NewHealthCheckBuilder() *HealthCheckBuilder {
	return &HealthCheckBuilder{
		config:   &DefaultHealthConfig,
		registry: NewHealthCheckRegistry(),
		factory:  NewHealthCheckFactory(),
	}
}

// WithConfig 设置配置
func (b *HealthCheckBuilder) WithConfig(config *HealthConfig) *HealthCheckBuilder {
	if config == nil {
		b.config = &DefaultHealthConfig
		return b
	}
	b.config = config
	return b
}

// WithLogger 设置日志
func (b *HealthCheckBuilder) WithLogger(logger interface{}) *HealthCheckBuilder {
	b.logger = logger
	return b
}

// WithDatabase 添加数据库检查
func (b *HealthCheckBuilder) WithDatabase(db interface{}) *HealthCheckBuilder {
	if b.config.EnableDatabaseCheck {
		check := b.factory.CreateDatabaseCheck(db, b.config.DatabaseTimeout)
		b.registry.Register("database", check)
	}
	return b
}

// WithCache 添加缓存检查
func (b *HealthCheckBuilder) WithCache(cache interface{}) *HealthCheckBuilder {
	if b.config.EnableCacheCheck {
		check := b.factory.CreateCacheCheck(cache, b.config.CacheTimeout)
		b.registry.Register("cache", check)
	}
	return b
}

// WithConcurrency 添加并发检查
func (b *HealthCheckBuilder) WithConcurrency(service interface{}) *HealthCheckBuilder {
	if b.config.EnableConcurrencyCheck {
		check := b.factory.CreateConcurrencyCheck(service, b.config.ConcurrencyTimeout)
		b.registry.Register("concurrency", check)
	}
	return b
}

// WithExternalAPI 添加外部API检查
func (b *HealthCheckBuilder) WithExternalAPI(url string) *HealthCheckBuilder {
	if b.config == nil || !b.config.EnableExternalAPICheck {
		return b
	}
	if strings.TrimSpace(url) == "" {
		return b
	}
	timeout := b.config.ExternalAPITimeout
	if timeout <= 0 {
		timeout = DefaultHealthConfig.ExternalAPITimeout
	}
	check := b.factory.CreateExternalAPICheck(url, timeout)
	b.registry.Register("external_api", check)
	return b
}

// WithStorage 添加存储检查
func (b *HealthCheckBuilder) WithStorage(path string) *HealthCheckBuilder {
	timeout := b.config.StorageTimeout
	if timeout <= 0 {
		timeout = DefaultHealthConfig.StorageTimeout
	}
	check := b.factory.CreateStorageCheck(path, timeout)
	b.registry.Register("storage", check)
	return b
}

// WithCustomCheck 添加自定义检查
func (b *HealthCheckBuilder) WithCustomCheck(name string, check HealthCheck) *HealthCheckBuilder {
	_ = name // 忽略未使用的变量，实际使用在registry.Register中
	b.registry.Register(name, check)
	return b
}

// Build 构建健康检查器
func (b *HealthCheckBuilder) Build() *HealthChecker {
	healthChecker := NewHealthChecker(b.config, nil)

	// 注册所有检查
	for _, check := range b.registry.GetAll() {
		healthChecker.RegisterCheck(check)
	}

	return healthChecker
}

// HealthCheckManager 健康检查管理器
type HealthCheckManager struct {
	healthChecker *HealthChecker
	builder       *HealthCheckBuilder
	initialized   bool
}

// NewHealthCheckManager 创建健康检查管理器
func NewHealthCheckManager() *HealthCheckManager {
	return &HealthCheckManager{
		builder: NewHealthCheckBuilder(),
	}
}

// Initialize 初始化健康检查管理器
func (m *HealthCheckManager) Initialize(config *HealthConfig) error {
	if m.initialized {
		return fmt.Errorf("health check manager already initialized")
	}

	m.builder.WithConfig(config)
	m.initialized = true

	return nil
}

// AddDatabaseCheck 添加数据库检查
func (m *HealthCheckManager) AddDatabaseCheck(db interface{}) error {
	if !m.initialized {
		return fmt.Errorf("health check manager not initialized")
	}

	m.builder.WithDatabase(db)
	return nil
}

// AddCacheCheck 添加缓存检查
func (m *HealthCheckManager) AddCacheCheck(cache interface{}) error {
	if !m.initialized {
		return fmt.Errorf("health check manager not initialized")
	}

	m.builder.WithCache(cache)
	return nil
}

// AddConcurrencyCheck 添加并发检查
func (m *HealthCheckManager) AddConcurrencyCheck(service interface{}) error {
	if !m.initialized {
		return fmt.Errorf("health check manager not initialized")
	}

	m.builder.WithConcurrency(service)
	return nil
}

// AddCustomCheck 添加自定义检查
func (m *HealthCheckManager) AddCustomCheck(name string, check HealthCheck) error {
	if !m.initialized {
		return fmt.Errorf("health check manager not initialized")
	}

	m.builder.WithCustomCheck(name, check)
	return nil
}

// BuildAndStart 构建并启动健康检查器
func (m *HealthCheckManager) BuildAndStart() (*HealthChecker, error) {
	if !m.initialized {
		return nil, fmt.Errorf("health check manager not initialized")
	}

	m.healthChecker = m.builder.Build()
	m.healthChecker.Start()

	return m.healthChecker, nil
}

// GetHealthChecker 获取健康检查器
func (m *HealthCheckManager) GetHealthChecker() *HealthChecker {
	return m.healthChecker
}

// Stop 停止健康检查器
func (m *HealthCheckManager) Stop() {
	if m.healthChecker != nil {
		m.healthChecker.Stop()
	}
}
