package health

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"law-oa-go/internal/concurrency"
)

func TestHealthCheckConfigurator_Basic(t *testing.T) {
	configurator := NewHealthCheckConfigurator()
	require.NotNil(t, configurator)

	registry := configurator.GetRegistry()
	assert.NotNil(t, registry)
}

func TestHealthCheckConfigurator_ConfigureDatabase(t *testing.T) {
	configurator := NewHealthCheckConfigurator()

	// 使用模拟的数据库接口
	mockDB := &MockDatabaseChecker{}

	configurator.ConfigureDatabase(mockDB, 5*time.Second)

	registry := configurator.GetRegistry()
	check, exists := registry.Get("database")

	assert.True(t, exists)
	assert.NotNil(t, check)
	assert.Equal(t, "database", check.GetName())
	assert.Equal(t, 5*time.Second, check.GetTimeout())
}

func TestHealthCheckConfigurator_ConfigureCache(t *testing.T) {
	configurator := NewHealthCheckConfigurator()

	// 使用模拟的缓存接口
	mockCache := &MockCacheChecker{}

	configurator.ConfigureCache(mockCache, 2*time.Second)

	registry := configurator.GetRegistry()
	check, exists := registry.Get("cache")

	assert.True(t, exists)
	assert.NotNil(t, check)
	assert.Equal(t, "cache", check.GetName())
	assert.Equal(t, 2*time.Second, check.GetTimeout())
}

func TestHealthCheckConfigurator_ConfigureConcurrency(t *testing.T) {
	configurator := NewHealthCheckConfigurator()

	// 使用模拟的并发接口
	mockConcurrency := &MockConcurrencyChecker{}

	configurator.ConfigureConcurrency(mockConcurrency, 3*time.Second)

	registry := configurator.GetRegistry()
	check, exists := registry.Get("concurrency")

	assert.True(t, exists)
	assert.NotNil(t, check)
	assert.Equal(t, "concurrency", check.GetName())
	assert.Equal(t, 3*time.Second, check.GetTimeout())
}

func TestHealthCheckConfigurator_ConfigureExternalAPI(t *testing.T) {
	configurator := NewHealthCheckConfigurator()

	url := "https://api.example.com/health"
	timeout := 3 * time.Second

	configurator.ConfigureExternalAPI(url, timeout)

	registry := configurator.GetRegistry()
	check, exists := registry.Get("external_api")

	assert.True(t, exists)
	assert.NotNil(t, check)
	assert.Equal(t, "external_api", check.GetName())
	assert.Equal(t, timeout, check.GetTimeout())
}

func TestHealthCheckConfigurator_ConfigureStorage(t *testing.T) {
	configurator := NewHealthCheckConfigurator()

	path := "/tmp"
	timeout := 2 * time.Second

	configurator.ConfigureStorage(path, timeout)

	registry := configurator.GetRegistry()
	check, exists := registry.Get("storage")

	assert.True(t, exists)
	assert.NotNil(t, check)
	assert.Equal(t, "storage", check.GetName())
	assert.Equal(t, timeout, check.GetTimeout())
}

func TestHealthCheckBuilder_Basic(t *testing.T) {
	builder := NewHealthCheckBuilder()
	require.NotNil(t, builder)

	// 构建健康检查器
	healthChecker := builder.Build()
	require.NotNil(t, healthChecker)

	// 应该没有注册任何检查
	results := healthChecker.RunChecks()
	assert.Empty(t, results)
}

func TestHealthCheckBuilder_WithConfig(t *testing.T) {
	builder := NewHealthCheckBuilder()

	config := &HealthConfig{
		EnableDatabaseCheck: true,
		DatabaseTimeout:     10 * time.Second,
	}

	builder.WithConfig(config)

	// 构建健康检查器
	healthChecker := builder.Build()
	require.NotNil(t, healthChecker)

	// 验证配置被应用
	assert.Equal(t, config, healthChecker.config)
}

func TestHealthCheckBuilder_WithDatabase(t *testing.T) {
	builder := NewHealthCheckBuilder()

	config := &HealthConfig{
		EnableDatabaseCheck: true,
		DatabaseTimeout:     5 * time.Second,
	}

	builder.WithConfig(config)

	// 使用模拟的数据库接口
	mockDB := &MockDatabaseChecker{}
	builder.WithDatabase(mockDB)

	// 构建健康检查器
	healthChecker := builder.Build()
	require.NotNil(t, healthChecker)

	// 验证数据库检查被注册
	results := healthChecker.RunChecks()
	assert.Contains(t, results, "database")
}

func TestHealthCheckBuilder_WithCache(t *testing.T) {
	builder := NewHealthCheckBuilder()

	config := &HealthConfig{
		EnableCacheCheck: true,
		CacheTimeout:     2 * time.Second,
	}

	builder.WithConfig(config)

	// 使用模拟的缓存接口
	mockCache := &MockCacheChecker{}
	builder.WithCache(mockCache)

	// 构建健康检查器
	healthChecker := builder.Build()
	require.NotNil(t, healthChecker)

	// 验证缓存检查被注册
	results := healthChecker.RunChecks()
	assert.Contains(t, results, "cache")
}

func TestHealthCheckBuilder_WithCustomCheck(t *testing.T) {
	builder := NewHealthCheckBuilder()

	// 添加自定义检查
	customCheck := &MockHealthCheck{
		name:    "custom_check",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "custom_check",
			Status:    StatusHealthy,
			Duration:  10,
			Timestamp: time.Now(),
		},
	}

	builder.WithCustomCheck("custom_check", customCheck)

	// 构建健康检查器
	healthChecker := builder.Build()
	require.NotNil(t, healthChecker)

	// 验证自定义检查被注册
	results := healthChecker.RunChecks()
	assert.Contains(t, results, "custom_check")
}

func TestHealthCheckBuilder_Chaining(t *testing.T) {
	builder := NewHealthCheckBuilder()

	config := &HealthConfig{
		EnableDatabaseCheck:    true,
		EnableCacheCheck:       true,
		EnableConcurrencyCheck: true,
		DatabaseTimeout:        5 * time.Second,
		CacheTimeout:           2 * time.Second,
		ConcurrencyTimeout:     3 * time.Second,
	}

	// 链式调用
	healthChecker := builder.
		WithConfig(config).
		WithDatabase(&MockDatabaseChecker{}).
		WithCache(&MockCacheChecker{}).
		WithConcurrency(&MockConcurrencyChecker{}).
		WithExternalAPI("https://api.example.com/health").
		WithStorage("/tmp").
		Build()

	require.NotNil(t, healthChecker)

	// 验证所有检查都被注册
	results := healthChecker.RunChecks()
	assert.Contains(t, results, "database")
	assert.Contains(t, results, "cache")
	assert.Contains(t, results, "concurrency")
	assert.Contains(t, results, "external_api")
	assert.Contains(t, results, "storage")
}

func TestHealthCheckManager_Basic(t *testing.T) {
	manager := NewHealthCheckManager()
	require.NotNil(t, manager)
	assert.False(t, manager.initialized)
}

func TestHealthCheckManager_Initialize(t *testing.T) {
	manager := NewHealthCheckManager()

	config := &HealthConfig{}
	err := manager.Initialize(config)

	assert.NoError(t, err)
	assert.True(t, manager.initialized)
}

func TestHealthCheckManager_DoubleInitialize(t *testing.T) {
	manager := NewHealthCheckManager()

	config := &HealthConfig{}
	err := manager.Initialize(config)
	require.NoError(t, err)

	// 第二次初始化应该失败
	err = manager.Initialize(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

func TestHealthCheckManager_AddChecks(t *testing.T) {
	manager := NewHealthCheckManager()

	config := &HealthConfig{
		EnableDatabaseCheck: true,
		EnableCacheCheck:    true,
	}

	err := manager.Initialize(config)
	require.NoError(t, err)

	// 添加检查
	err = manager.AddDatabaseCheck(&MockDatabaseChecker{})
	assert.NoError(t, err)

	err = manager.AddCacheCheck(&MockCacheChecker{})
	assert.NoError(t, err)

	// 添加自定义检查
	customCheck := &MockHealthCheck{
		name:    "custom",
		timeout: 1 * time.Second,
		result: &HealthCheckResult{
			Name:      "custom",
			Status:    StatusHealthy,
			Duration:  10,
			Timestamp: time.Now(),
		},
	}

	err = manager.AddCustomCheck("custom", customCheck)
	assert.NoError(t, err)
}

func TestHealthCheckManager_AddChecksWithoutInitialize(t *testing.T) {
	manager := NewHealthCheckManager()

	// 未初始化时添加检查应该失败
	err := manager.AddDatabaseCheck(&MockDatabaseChecker{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestHealthCheckManager_BuildAndStart(t *testing.T) {
	manager := NewHealthCheckManager()

	config := &HealthConfig{
		EnableDatabaseCheck: true,
		DatabaseTimeout:     5 * time.Second,
	}

	err := manager.Initialize(config)
	require.NoError(t, err)

	err = manager.AddDatabaseCheck(&MockDatabaseChecker{})
	require.NoError(t, err)

	// 构建并启动
	healthChecker, err := manager.BuildAndStart()
	assert.NoError(t, err)
	assert.NotNil(t, healthChecker)
	assert.True(t, healthChecker.running)

	// 停止
	manager.Stop()
	assert.False(t, healthChecker.running)
}

func TestHealthCheckManager_GetHealthChecker(t *testing.T) {
	manager := NewHealthCheckManager()

	// 构建前应该返回nil
	healthChecker := manager.GetHealthChecker()
	assert.Nil(t, healthChecker)

	config := &HealthConfig{}
	err := manager.Initialize(config)
	require.NoError(t, err)

	// 构建并启动
	healthChecker, err = manager.BuildAndStart()
	require.NoError(t, err)

	// 现在应该返回健康检查器
	healthChecker = manager.GetHealthChecker()
	assert.NotNil(t, healthChecker)
}

func TestSimpleHealthCheck(t *testing.T) {
	check := &SimpleHealthCheck{
		name:    "simple_check",
		timeout: 1 * time.Second,
		checker: func(ctx context.Context) *HealthCheckResult {
			return &HealthCheckResult{
				Message: "Simple check passed",
			}
		},
	}

	result := check.Check(context.Background())

	assert.Equal(t, "simple_check", result.Name)
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "Simple check passed", result.Message)
	assert.Greater(t, result.Duration, int64(0))
	assert.NotZero(t, result.Timestamp)
}

func TestSimpleHealthCheck_EmptyStatus(t *testing.T) {
	check := &SimpleHealthCheck{
		name:    "test",
		timeout: 1 * time.Second,
		checker: func(ctx context.Context) *HealthCheckResult {
			return &HealthCheckResult{
				Message: "Test",
			}
		},
	}

	result := check.Check(context.Background())
	assert.Equal(t, StatusHealthy, result.Status)
}

// MockDatabaseChecker 模拟数据库检查器
type MockDatabaseChecker struct{}

func (m *MockDatabaseChecker) GetDB() interface{} {
	// 返回一个模拟的 sql.DB
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

// MockCacheChecker 模拟缓存检查器
type MockCacheChecker struct{}

func (m *MockCacheChecker) GetCache() interface{} {
	// 返回一个模拟的缓存服务
	return &MockCacheService{}
}

// MockConcurrencyChecker 模拟并发检查器
type MockConcurrencyChecker struct{}

func (m *MockConcurrencyChecker) GetService() interface{} {
	// 返回一个模拟的并发服务
	return &concurrency.ConcurrentService{}
}

// MockCacheService 模拟缓存服务
type MockCacheService struct{}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (m *MockCacheService) Get(ctx context.Context, key string, value interface{}) error {
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *MockCacheService) Clear(ctx context.Context) error {
	return nil
}
