package security

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCAManagerBasics 测试CA管理器基本功能
func TestCAManagerBasics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // 测试时只记录错误

	config := &CAConfiguration{
		OrganizationName:       "测试律师事务所",
		RootCAValidity:         1,  // 根CA有效期1年用于测试
		IntermediateCAValidity: 1,  // 中间CA有效期1年用于测试
		DefaultKeySize:         2048, // 使用较小的密钥以加快测试
		DefaultHashAlgorithm:   "SHA256",
		EnableOCSP:            false, // 测试时禁用网络功能
		EnableCRL:             false,
		HSMEnabled:            false,
		BackupEnabled:         false,
	}

	// 测试创建CA管理器
	manager := NewCAManager(config, logger)
	assert.NotNil(t, manager)
	assert.Equal(t, "测试律师事务所", manager.config.OrganizationName)
	assert.Equal(t, 2048, manager.config.DefaultKeySize)

	// 测试初始化
	err := manager.Initialize()
	require.NoError(t, err)
	assert.True(t, manager.IsInitialized())

	// 检查根CA
	rootCA := manager.GetRootCA()
	assert.NotNil(t, rootCA)
	assert.Equal(t, CAStatusActive, rootCA.Status)
	assert.Equal(t, "测试律师事务所 Root CA", rootCA.Profile.CommonName)
	assert.Equal(t, CATypeRoot, rootCA.Profile.Type)

	// 检查信任存储
	trustStore := manager.GetTrustStore()
	assert.NotNil(t, trustStore)
	assert.Len(t, trustStore.RootCertificates, 1)
	assert.Equal(t, "测试律师事务所 Trust Store", trustStore.Name)

	t.Logf("✅ CA管理器基本功能测试通过")
	t.Logf("   - 根CA: %s", rootCA.Profile.CommonName)
	t.Logf("   - 序列号: %s", rootCA.SerialNumber.String())
	t.Logf("   - 有效期至: %s", rootCA.ExpiresAt.Format("2006-01-02"))
}

// TestCAProfiles 测试CA配置文件
func TestCAProfiles(t *testing.T) {
	// 测试默认根CA配置文件
	rootProfile := DefaultRootCAProfile()
	assert.Equal(t, CATypeRoot, rootProfile.Type)
	assert.Equal(t, "Root Certificate Authority", rootProfile.Name)
	assert.Equal(t, "Law Office Root CA", rootProfile.CommonName)
	assert.Equal(t, 4096, rootProfile.KeySize)
	assert.Equal(t, 20*365, rootProfile.ValidityPeriod)
	assert.Equal(t, 3, rootProfile.MaxPathLength)

	// 测试默认中间CA配置文件
	intermediateProfile := DefaultIntermediateCAProfile("测试部门")
	assert.Equal(t, CATypeIntermediate, intermediateProfile.Type)
	assert.Equal(t, "Intermediate Certificate Authority - 测试部门", intermediateProfile.Name)
	assert.Equal(t, "Law Office Intermediate CA - 测试部门", intermediateProfile.CommonName)
	assert.Equal(t, 4096, intermediateProfile.KeySize)
	assert.Equal(t, 10*365, intermediateProfile.ValidityPeriod)
	assert.Equal(t, 2, intermediateProfile.MaxPathLength)

	t.Logf("✅ CA配置文件测试通过")
	t.Logf("   - 根CA有效期: %d天", rootProfile.ValidityPeriod)
	t.Logf("   - 中间CA有效期: %d天", intermediateProfile.ValidityPeriod)
}

// TestTrustStoreBasics 测试信任存储基本功能
func TestTrustStoreBasics(t *testing.T) {
	// 创建信任存储
	trustStore := NewTrustStore("测试信任存储")
	assert.NotNil(t, trustStore)
	assert.Equal(t, "测试信任存储", trustStore.Name)
	assert.Len(t, trustStore.RootCertificates, 0)
	assert.Len(t, trustStore.IntermediateCerts, 0)
	assert.True(t, trustStore.AutoUpdate)
	assert.True(t, trustStore.LastUpdated.Before(time.Now().Add(time.Minute)))

	// 测试证书池
	rootPool := trustStore.GetRootPool()
	assert.NotNil(t, rootPool)

	intermediatePool := trustStore.GetIntermediatePool()
	assert.NotNil(t, intermediatePool)

	t.Logf("✅ 信任存储基本功能测试通过")
	t.Logf("   - 信任存储名称: %s", trustStore.Name)
	t.Logf("   - 创建时间: %s", trustStore.LastUpdated.Format("2006-01-02 15:04:05"))
}

// TestCAConfiguration 测试CA配置
func TestCAConfiguration(t *testing.T) {
	// 测试默认配置
	defaultConfig := DefaultCAConfiguration()
	assert.Equal(t, "律师事务所", defaultConfig.OrganizationName)
	assert.Equal(t, 20, defaultConfig.RootCAValidity)
	assert.Equal(t, 10, defaultConfig.IntermediateCAValidity)
	assert.Equal(t, 4096, defaultConfig.DefaultKeySize)
	assert.Equal(t, "SHA256", defaultConfig.DefaultHashAlgorithm)
	assert.True(t, defaultConfig.EnableOCSP)
	assert.True(t, defaultConfig.EnableCRL)
	assert.False(t, defaultConfig.HSMEnabled)
	assert.True(t, defaultConfig.BackupEnabled)

	// 测试自定义配置
	customConfig := &CAConfiguration{
		OrganizationName: "自定义律师事务所",
		RootCAValidity:   15,
		DefaultKeySize:   2048,
		EnableOCSP:       false,
		HSMEnabled:       true,
	}

	manager := NewCAManager(customConfig, nil)
	assert.Equal(t, "自定义律师事务所", manager.config.OrganizationName)
	assert.Equal(t, 15, manager.config.RootCAValidity)
	assert.Equal(t, 2048, manager.config.DefaultKeySize)
	assert.False(t, manager.config.EnableOCSP)
	assert.True(t, manager.config.HSMEnabled)

	t.Logf("✅ CA配置测试通过")
	t.Logf("   - 组织名称: %s", manager.config.OrganizationName)
	t.Logf("   - 根CA有效期: %d年", manager.config.RootCAValidity)
	t.Logf("   - 默认密钥大小: %d位", manager.config.DefaultKeySize)
}

// TestCATypes 测试CA类型
func TestCATypes(t *testing.T) {
	// 测试CA类型常量
	assert.Equal(t, CAType("root"), CATypeRoot)
	assert.Equal(t, CAType("intermediate"), CATypeIntermediate)
	assert.Equal(t, CAType("leaf"), CATypeLeaf)

	// 测试CA状态常量
	assert.Equal(t, CAStatus("active"), CAStatusActive)
	assert.Equal(t, CAStatus("suspended"), CAStatusSuspended)
	assert.Equal(t, CAStatus("revoked"), CAStatusRevoked)
	assert.Equal(t, CAStatus("expired"), CAStatusExpired)

	t.Logf("✅ CA类型和状态测试通过")
}

// TestCertificateAuthorityStructure 测试证书颁发机构结构
func TestCertificateAuthorityStructure(t *testing.T) {
	// 创建CA配置文件
	profile := DefaultRootCAProfile()
	profile.KeySize = 2048 // 使用较小密钥以加快测试

	// 创建CA实例（不实际生成证书）
	ca := &CertificateAuthority{
		ID:           "test_ca_001",
		Profile:      profile,
		Status:       CAStatusActive,
		SerialNumber: nil, // 这里不实际生成
		NextSerial:   nil,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    time.Now().AddDate(1, 0, 0),
		UsageCount:   0,
	}

	// 验证CA结构
	assert.Equal(t, "test_ca_001", ca.ID)
	assert.Equal(t, profile, ca.Profile)
	assert.Equal(t, CAStatusActive, ca.Status)
	assert.Equal(t, int64(0), ca.UsageCount)
	assert.True(t, ca.CreatedAt.Before(time.Now().Add(time.Minute)))
	assert.True(t, ca.UpdatedAt.Before(time.Now().Add(time.Minute)))

	t.Logf("✅ 证书颁发机构结构测试通过")
	t.Logf("   - CA ID: %s", ca.ID)
	t.Logf("   - 状态: %s", ca.Status)
	t.Logf("   - 类型: %s", ca.Profile.Type)
}