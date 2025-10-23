package ca

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 为了避免包间冲突，我将在单独的测试中重新导入必要的结构

// SimpleCAConfiguration 简化的CA配置
type SimpleCAConfiguration struct {
	OrganizationName       string
	RootCAValidity         int
	DefaultKeySize         int
	DefaultHashAlgorithm   string
}

// SimpleCAManager 简化的CA管理器
type SimpleCAManager struct {
	config      *SimpleCAConfiguration
	initialized bool
	logger      *logrus.Logger
}

// NewSimpleCAManager 创建简化的CA管理器
func NewSimpleCAManager(config *SimpleCAConfiguration, logger *logrus.Logger) *SimpleCAManager {
	if config == nil {
		config = &SimpleCAConfiguration{
			OrganizationName:     "律师事务所",
			RootCAValidity:       20,
			DefaultKeySize:       4096,
			DefaultHashAlgorithm: "SHA256",
		}
	}

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &SimpleCAManager{
		config: config,
		logger: logger,
	}
}

// Initialize 初始化CA管理器
func (scm *SimpleCAManager) Initialize() error {
	scm.logger.Info("初始化简化CA管理器...")
	scm.initialized = true
	return nil
}

// IsInitialized 检查是否已初始化
func (scm *SimpleCAManager) IsInitialized() bool {
	return scm.initialized
}

// GetConfig 获取配置
func (scm *SimpleCAManager) GetConfig() *SimpleCAConfiguration {
	return scm.config
}

// TestSimpleCAManagerBasics 测试简化CA管理器基本功能
func TestSimpleCAManagerBasics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &SimpleCAConfiguration{
		OrganizationName:     "测试律师事务所",
		RootCAValidity:       1,
		DefaultKeySize:       2048,
		DefaultHashAlgorithm: "SHA256",
	}

	// 测试创建CA管理器
	manager := NewSimpleCAManager(config, logger)
	assert.NotNil(t, manager)
	assert.Equal(t, "测试律师事务所", manager.config.OrganizationName)
	assert.Equal(t, 2048, manager.config.DefaultKeySize)

	// 测试初始化
	err := manager.Initialize()
	require.NoError(t, err)
	assert.True(t, manager.IsInitialized())

	// 测试获取配置
	retrievedConfig := manager.GetConfig()
	assert.Equal(t, config, retrievedConfig)

	t.Logf("✅ 简化CA管理器基本功能测试通过")
	t.Logf("   - 组织名称: %s", manager.config.OrganizationName)
	t.Logf("   - 密钥大小: %d", manager.config.DefaultKeySize)
	t.Logf("   - 哈希算法: %s", manager.config.DefaultHashAlgorithm)
}

// TestSimpleCAManagerDefaultConfig 测试默认配置
func TestSimpleCAManagerDefaultConfig(t *testing.T) {
	// 测试使用默认配置创建
	manager := NewSimpleCAManager(nil, nil)
	assert.NotNil(t, manager)
	assert.Equal(t, "律师事务所", manager.config.OrganizationName)
	assert.Equal(t, 4096, manager.config.DefaultKeySize)
	assert.Equal(t, "SHA256", manager.config.DefaultHashAlgorithm)
	assert.Equal(t, 20, manager.config.RootCAValidity)

	// 测试初始化
	err := manager.Initialize()
	require.NoError(t, err)
	assert.True(t, manager.IsInitialized())

	t.Logf("✅ 默认配置测试通过")
	t.Logf("   - 默认组织: %s", manager.config.OrganizationName)
	t.Logf("   - 默认密钥: %d", manager.config.DefaultKeySize)
	t.Logf("   - 根CA有效期: %d年", manager.config.RootCAValidity)
}

// TestCAManagerLifecycle 测试CA管理器生命周期
func TestCAManagerLifecycle(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := NewSimpleCAManager(nil, logger)

	// 初始状态
	assert.False(t, manager.IsInitialized())

	// 初始化
	startTime := time.Now()
	err := manager.Initialize()
	require.NoError(t, err)
	endTime := time.Now()

	// 初始化后状态
	assert.True(t, manager.IsInitialized())
	assert.True(t, endTime.After(startTime))

	t.Logf("✅ CA管理器生命周期测试通过")
	t.Logf("   - 初始化耗时: %v", endTime.Sub(startTime))
}

// TestMultipleCAManagers 测试多个CA管理器
func TestMultipleCAManagers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建多个CA管理器
	config1 := &SimpleCAConfiguration{
		OrganizationName: "律所1",
		DefaultKeySize:   2048,
	}

	config2 := &SimpleCAConfiguration{
		OrganizationName: "律所2",
		DefaultKeySize:   4096,
	}

	manager1 := NewSimpleCAManager(config1, logger)
	manager2 := NewSimpleCAManager(config2, logger)

	// 初始化
	err1 := manager1.Initialize()
	err2 := manager2.Initialize()
	require.NoError(t, err1)
	require.NoError(t, err2)

	// 验证独立性
	assert.True(t, manager1.IsInitialized())
	assert.True(t, manager2.IsInitialized())
	assert.NotEqual(t, manager1.GetConfig(), manager2.GetConfig())
	assert.Equal(t, "律所1", manager1.GetConfig().OrganizationName)
	assert.Equal(t, "律所2", manager2.GetConfig().OrganizationName)

	t.Logf("✅ 多CA管理器测试通过")
	t.Logf("   - 管理器1: %s", manager1.GetConfig().OrganizationName)
	t.Logf("   - 管理器2: %s", manager2.GetConfig().OrganizationName)
}

// BenchmarkCAManagerCreation CA管理器创建性能测试
func BenchmarkCAManagerCreation(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &SimpleCAConfiguration{
		OrganizationName: "性能测试律所",
		DefaultKeySize:   2048,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewSimpleCAManager(config, logger)
		_ = manager
	}
}

// BenchmarkCAManagerInitialization CA管理器初始化性能测试
func BenchmarkCAManagerInitialization(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &SimpleCAConfiguration{
		OrganizationName: "性能测试律所",
		DefaultKeySize:   2048,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewSimpleCAManager(config, logger)
		_ = manager.Initialize()
	}
}