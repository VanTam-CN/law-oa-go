package security

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestCAManager 设置测试CA管理器
func setupTestCAManager(t *testing.T) *CAManager {
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

	manager := NewCAManager(config, logger)
	return manager
}

// TestNewCAManager 测试CA管理器创建
func TestNewCAManager(t *testing.T) {
	// 测试使用默认配置
	manager1 := NewCAManager(nil, nil)
	assert.NotNil(t, manager1)
	assert.Equal(t, "律师事务所", manager1.config.OrganizationName)
	assert.Equal(t, 4096, manager1.config.DefaultKeySize)

	// 测试使用自定义配置
	logger := logrus.New()
	config := &CAConfiguration{
		OrganizationName: "测试机构",
		DefaultKeySize:   2048,
	}
	manager2 := NewCAManager(config, logger)
	assert.NotNil(t, manager2)
	assert.Equal(t, "测试机构", manager2.config.OrganizationName)
	assert.Equal(t, 2048, manager2.config.DefaultKeySize)
}

// TestCAManagerInitialize 测试CA管理器初始化
func TestCAManagerInitialize(t *testing.T) {
	manager := setupTestCAManager(t)

	// 测试初始化
	err := manager.Initialize()
	require.NoError(t, err)
	assert.True(t, manager.IsInitialized())

	// 检查根CA是否已创建
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

	t.Logf("✅ CA管理器初始化测试通过")
	t.Logf("   - 根CA: %s", rootCA.Profile.CommonName)
	t.Logf("   - 序列号: %s", rootCA.SerialNumber.String())
	t.Logf("   - 有效期至: %s", rootCA.ExpiresAt.Format("2006-01-02"))
}

// TestCreateIntermediateCA 测试创建中间CA
func TestCreateIntermediateCA(t *testing.T) {
	manager := setupTestCAManager(t)
	err := manager.Initialize()
	require.NoError(t, err)

	rootCA := manager.GetRootCA()
	require.NotNil(t, rootCA)

	// 创建中间CA配置
	profile := DefaultIntermediateCAProfile("测试")
	profile.Organization = "测试律师事务所"
	profile.KeySize = 2048 // 使用较小密钥以加快测试

	// 创建中间CA
	intermediateCA, err := manager.createCertificateAuthority(profile, rootCA)
	require.NoError(t, err)
	require.NotNil(t, intermediateCA)

	// 验证中间CA属性
	assert.Equal(t, CAStatusActive, intermediateCA.Status)
	assert.Equal(t, "测试律师事务所 Intermediate CA - 测试", intermediateCA.Profile.CommonName)
	assert.Equal(t, CATypeIntermediate, intermediateCA.Profile.Type)
	assert.Equal(t, rootCA, intermediateCA.ParentCA)
	assert.Equal(t, rootCA.Certificate.Subject, intermediateCA.Certificate.Issuer)
	assert.Equal(t, 2, intermediateCA.Profile.MaxPathLength)

	// 验证证书链
	trustStore := manager.GetTrustStore()
	trustStore.AddIntermediateCertificate(intermediateCA.Certificate)

	result, err := trustStore.ValidateCertificateChain(intermediateCA.Certificate)
	require.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Len(t, result.Chain, 2) // 中间CA + 根CA

	t.Logf("✅ 中间CA创建测试通过")
	t.Logf("   - 中间CA: %s", intermediateCA.Profile.CommonName)
	t.Logf("   - 序列号: %s", intermediateCA.SerialNumber.String())
	t.Logf("   - 证书链长度: %d", len(result.Chain))
}

// TestCAManagerKeyPairGeneration 测试密钥对生成
func TestCAManagerKeyPairGeneration(t *testing.T) {
	manager := setupTestCAManager(t)

	testCases := []struct {
		name      string
		algorithm string
		keySize   int
	}{
		{
			name:      "RSA-2048",
			algorithm: "RSA",
			keySize:   2048,
		},
		{
			name:      "RSA-4096",
			algorithm: "RSA",
			keySize:   4096,
		},
		{
			name:      "ECDSA-P256",
			algorithm: "ECDSA",
			keySize:   256,
		},
		{
			name:      "ECDSA-P384",
			algorithm: "ECDSA",
			keySize:   384,
		},
		{
			name:      "Ed25519",
			algorithm: "Ed25519",
			keySize:   0, // Ed25519忽略密钥大小
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			privateKey, publicKey, err := manager.generateKeyPair(tc.algorithm, tc.keySize)
			require.NoError(t, err)
			require.NotNil(t, privateKey)
			require.NotNil(t, publicKey)

			// 测试密钥编码
			privateKeyPEM, publicKeyPEM, err := manager.encodeKeyPair(privateKey, publicKey)
			require.NoError(t, err)
			assert.NotEmpty(t, privateKeyPEM)
			assert.NotEmpty(t, publicKeyPEM)

			t.Logf("✅ %s 密钥对生成测试通过", tc.name)
			t.Logf("   - 私钥长度: %d bytes", len(privateKeyPEM))
			t.Logf("   - 公钥长度: %d bytes", len(publicKeyPEM))
		})
	}
}

// TestTrustStoreOperations 测试信任存储操作
func TestTrustStoreOperations(t *testing.T) {
	manager := setupTestCAManager(t)
	err := manager.Initialize()
	require.NoError(t, err)

	trustStore := manager.GetTrustStore()
	require.NotNil(t, trustStore)

	// 初始状态检查
	assert.Len(t, trustStore.RootCertificates, 1) // 根CA
	assert.Len(t, trustStore.IntermediateCerts, 0)
	assert.Len(t, trustStore.TrustedAnchors, 1)

	// 创建测试证书
	profile := DefaultIntermediateCAProfile("信任存储测试")
	profile.KeySize = 2048

	intermediateCA, err := manager.createCertificateAuthority(profile, manager.GetRootCA())
	require.NoError(t, err)

	// 添加中间证书
	trustStore.AddIntermediateCertificate(intermediateCA.Certificate)
	assert.Len(t, trustStore.IntermediateCerts, 1)
	assert.Len(t, trustStore.IntermediateCertsPEM, 1)

	// 测试证书池获取
	rootPool := trustStore.GetRootPool()
	assert.NotNil(t, rootPool)
	assert.True(t, rootPool.AppendCertsFromPEM([]byte(manager.GetRootCA().CertificatePEM)))

	intermediatePool := trustStore.GetIntermediatePool()
	assert.NotNil(t, intermediatePool)
	assert.True(t, intermediatePool.AppendCertsFromPEM([]byte(intermediateCA.CertificatePEM)))

	t.Logf("✅ 信任存储操作测试通过")
	t.Logf("   - 根证书数: %d", len(trustStore.RootCertificates))
	t.Logf("   - 中间证书数: %d", len(trustStore.IntermediateCerts))
	t.Logf("   - 信任锚数: %d", len(trustStore.TrustedAnchors))
}

// TestCAManagerCertificateChainValidation 测试证书链验证
func TestCAManagerCertificateChainValidation(t *testing.T) {
	manager := setupTestCAManager(t)
	err := manager.Initialize()
	require.NoError(t, err)

	trustStore := manager.GetTrustStore()

	// 测试根CA证书验证
	rootCA := manager.GetRootCA()
	result, err := trustStore.ValidateCertificateChain(rootCA.Certificate)
	require.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Len(t, result.Chain, 1) // 只有根CA
	assert.Equal(t, rootCA.Certificate, result.TrustAnchor)

	// 创建并验证中间CA证书
	intermediateProfile := DefaultIntermediateCAProfile("链验证测试")
	intermediateProfile.KeySize = 2048

	intermediateCA, err := manager.createCertificateAuthority(intermediateProfile, rootCA)
	require.NoError(t, err)

	trustStore.AddIntermediateCertificate(intermediateCA.Certificate)

	result, err = trustStore.ValidateCertificateChain(intermediateCA.Certificate)
	require.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Len(t, result.Chain, 2) // 中间CA + 根CA
	assert.Equal(t, rootCA.Certificate, result.TrustAnchor)

	// 创建并验证终端实体证书
	leafProfile := &CAProfile{
		Name:                "测试终端实体",
		Type:                CATypeLeaf,
		CommonName:          "测试用户",
		Organization:        "测试律师事务所",
		Country:             "CN",
		KeyAlgorithm:        "RSA",
		KeySize:             2048,
		HashAlgorithm:       "SHA256",
		ValidityPeriod:      365, // 1年
		MaxPathLength:       0,
		KeyUsage:            x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:         []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	}

	leafCA, err := manager.createCertificateAuthority(leafProfile, intermediateCA)
	require.NoError(t, err)

	result, err = trustStore.ValidateCertificateChain(leafCA.Certificate)
	require.NoError(t, err)
	assert.True(t, result.IsValid)
	assert.Len(t, result.Chain, 3) // 终端实体 + 中间CA + 根CA
	assert.Equal(t, rootCA.Certificate, result.TrustAnchor)

	t.Logf("✅ 证书链验证测试通过")
	t.Logf("   - 根CA链长度: %d", 1)
	t.Logf("   - 中间CA链长度: %d", len(result.Chain))
	t.Logf("   - 终端实体链长度: %d", len(result.Chain))
	t.Logf("   - 信任锚: %s", result.TrustAnchor.Subject.CommonName)
}

// TestCAManagerProfiles 测试CA配置文件
func TestCAManagerProfiles(t *testing.T) {
	// 测试默认根CA配置文件
	rootProfile := DefaultRootCAProfile()
	assert.Equal(t, CATypeRoot, rootProfile.Type)
	assert.Equal(t, "Root Certificate Authority", rootProfile.Name)
	assert.Equal(t, "Law Office Root CA", rootProfile.CommonName)
	assert.Equal(t, 4096, rootProfile.KeySize)
	assert.Equal(t, 20*365, rootProfile.ValidityPeriod)
	assert.Equal(t, 3, rootProfile.MaxPathLength)
	assert.Contains(t, rootProfile.KeyUsage, x509.KeyUsageCertSign)
	assert.Contains(t, rootProfile.KeyUsage, x509.KeyUsageCRLSign)

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

// TestCertificateAuthorityLifecycle 测试证书颁发机构生命周期
func TestCertificateAuthorityLifecycle(t *testing.T) {
	manager := setupTestCAManager(t)
	err := manager.Initialize()
	require.NoError(t, err)

	rootCA := manager.GetRootCA()
	require.NotNil(t, rootCA)

	// 创建中间CA
	intermediateProfile := DefaultIntermediateCAProfile("生命周期测试")
	intermediateProfile.KeySize = 2048

	intermediateCA, err := manager.createCertificateAuthority(intermediateProfile, rootCA)
	require.NoError(t, err)

	// 验证初始状态
	assert.Equal(t, CAStatusActive, intermediateCA.Status)
	assert.Equal(t, 0, intermediateCA.UsageCount)
	assert.True(t, intermediateCA.CreatedAt.Before(time.Now().Add(time.Minute)))
	assert.True(t, intermediateCA.UpdatedAt.Before(time.Now().Add(time.Minute)))

	// 更新使用统计
	intermediateCA.UsageCount++
	intermediateCA.LastUsedAt = time.Now()
	intermediateCA.UpdatedAt = time.Now()

	assert.Equal(t, int64(1), intermediateCA.UsageCount)
	assert.True(t, intermediateCA.LastUsedAt.After(intermediateCA.CreatedAt))

	// 测试状态变更
	intermediateCA.Status = CAStatusSuspended
	assert.Equal(t, CAStatusSuspended, intermediateCA.Status)

	intermediateCA.Status = CAStatusRevoked
	assert.Equal(t, CAStatusRevoked, intermediateCA.Status)

	t.Logf("✅ 证书颁发机构生命周期测试通过")
	t.Logf("   - 创建时间: %s", intermediateCA.CreatedAt.Format("2006-01-02 15:04:05"))
	t.Logf("   - 更新时间: %s", intermediateCA.UpdatedAt.Format("2006-01-02 15:04:05"))
	t.Logf("   - 使用次数: %d", intermediateCA.UsageCount)
	t.Logf("   - 当前状态: %s", intermediateCA.Status)
}

// BenchmarkCreateRootCA 根CA创建性能测试
func BenchmarkCreateRootCA(b *testing.B) {
	manager := setupTestCAManager(&testing.T{})
	profile := DefaultRootCAProfile()
	profile.KeySize = 2048 // 使用较小密钥以提高性能

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.createCertificateAuthority(profile, nil)
		if err != nil {
			b.Fatalf("创建根CA失败: %v", err)
		}
	}
}

// BenchmarkCreateIntermediateCA 中间CA创建性能测试
func BenchmarkCreateIntermediateCA(b *testing.B) {
	manager := setupTestCAManager(&testing.T{})
	err := manager.Initialize()
	if err != nil {
		b.Fatalf("初始化CA管理器失败: %v", err)
	}

	rootCA := manager.GetRootCA()
	profile := DefaultIntermediateCAProfile("性能测试")
	profile.KeySize = 2048

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.createCertificateAuthority(profile, rootCA)
		if err != nil {
			b.Fatalf("创建中间CA失败: %v", err)
		}
	}
}

// BenchmarkCertificateChainValidation 证书链验证性能测试
func BenchmarkCertificateChainValidation(b *testing.B) {
	manager := setupTestCAManager(&testing.T{})
	err := manager.Initialize()
	if err != nil {
		b.Fatalf("初始化CA管理器失败: %v", err)
	}

	// 创建测试证书链
	rootCA := manager.GetRootCA()
	intermediateProfile := DefaultIntermediateCAProfile("性能测试")
	intermediateProfile.KeySize = 2048

	intermediateCA, err := manager.createCertificateAuthority(intermediateProfile, rootCA)
	if err != nil {
		b.Fatalf("创建中间CA失败: %v", err)
	}

	trustStore := manager.GetTrustStore()
	trustStore.AddIntermediateCertificate(intermediateCA.Certificate)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := trustStore.ValidateCertificateChain(intermediateCA.Certificate)
		if err != nil {
			b.Fatalf("验证证书链失败: %v", err)
		}
	}
}