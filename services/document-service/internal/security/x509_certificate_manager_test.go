package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestX509Manager 设置测试X.509证书管理器
func setupTestX509Manager(t *testing.T) *X509CertificateManager {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建测试CA
	caProfile := &CAProfile{
		CommonName:    "测试根CA",
		Organization:  "测试律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  10 * 365,
		IsCA:          true,
		MaxPathLength: 3,
	}

	ca, err := NewCertificateAuthority(caProfile, nil)
	require.NoError(t, err)

	manager := NewX509CertificateManager(ca, logger)
	return manager
}

// TestX509CertificateManagerCreation 测试X.509证书管理器创建
func TestX509CertificateManagerCreation(t *testing.T) {
	manager := setupTestX509Manager(t)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.ca)
	assert.NotNil(t, manager.templateManager)
	assert.NotNil(t, manager.serialGenerator)
	assert.NotNil(t, manager.batchProcessor)

	// 检查默认模板
	templates := manager.ListTemplates()
	assert.GreaterOrEqual(t, len(templates), 4) // tls-server, tls-client, code-signing, email-signing

	templateNames := make(map[string]bool)
	for _, template := range templates {
		templateNames[template.ID] = true
	}

	assert.True(t, templateNames["tls-server"])
	assert.True(t, templateNames["tls-client"])
	assert.True(t, templateNames["code-signing"])
	assert.True(t, templateNames["email-signing"])

	t.Logf("✅ X.509证书管理器创建测试通过")
	t.Logf("   - 默认模板数量: %d", len(templates))
}

// TestIssueCertificate 测试证书签发
func TestIssueCertificate(t *testing.T) {
	manager := setupTestX509Manager(t)

	// 生成密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// 创建证书请求
	req := &CertificateRequest{
		TemplateID:  "tls-server",
		Subject: pkix.Name{
			CommonName:   "www.example.com",
			Organization: []string{"Example Corp"},
			Country:      []string{"CN"},
		},
		PublicKey:  &privateKey.PublicKey,
		Validity:   365 * 24 * time.Hour,
		KeyUsage:   x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:   []string{"www.example.com", "example.com"},
		RequesterID: "test_user",
		Metadata: map[string]interface{}{
			"environment": "test",
		},
		RequestedAt: time.Now(),
		Status:      CertificateStatusPending,
	}

	// 签发证书
	cert, err := manager.IssueCertificate(req)
	require.NoError(t, err)
	require.NotNil(t, cert)

	// 验证证书属性
	assert.Equal(t, "www.example.com", cert.Subject.CommonName)
	assert.Equal(t, "Example Corp", cert.Subject.Organization[0])
	assert.Equal(t, "CN", cert.Subject.Country[0])
	assert.Equal(t, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment, cert.Certificate.KeyUsage)
	assert.Contains(t, cert.Certificate.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	assert.Contains(t, cert.Certificate.DNSNames, "www.example.com")
	assert.Contains(t, cert.Certificate.DNSNames, "example.com")
	assert.Equal(t, CertificateStatusActive, cert.Status)
	assert.Equal(t, "tls-server", cert.TemplateID)
	assert.Equal(t, "test_user", cert.RequesterID)

	// 验证证书链
	err = manager.ValidateCertificateChain(cert.Certificate)
	assert.NoError(t, err)

	t.Logf("✅ 证书签发测试通过")
	t.Logf("   - 证书ID: %s", cert.ID)
	t.Logf("   - 序列号: %s", cert.SerialNumber.String())
	t.Logf("   - 主题: %s", cert.Subject.CommonName)
	t.Logf("   - 颁发者: %s", cert.Issuer.CommonName)
	t.Logf("   - 有效期: %s 至 %s", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
}

// TestBatchIssueCertificate 测试批量证书签发
func TestBatchIssueCertificate(t *testing.T) {
	manager := setupTestX509Manager(t)

	// 生成多个密钥对
	keys := make([]rsa.PrivateKey, 3)
	for i := 0; i < 3; i++ {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		keys[i] = *key
	}

	// 创建批量请求
	singleRequests := []SingleCertificateRequest{
		{
			RequestID: "req1",
			Variables: map[string]interface{}{
				"common_name":    "server1.example.com",
				"organization":  "Example Corp",
				"country":        "CN",
				"dns_names":      []string{"server1.example.com"},
			},
		},
		{
			RequestID: "req2",
			Variables: map[string]interface{}{
				"common_name":    "server2.example.com",
				"organization":  "Example Corp",
				"country":        "CN",
				"dns_names":      []string{"server2.example.com"},
			},
		},
		{
			RequestID: "req3",
			Variables: map[string]interface{}{
				"common_name":    "server3.example.com",
				"organization":  "Example Corp",
				"country":        "CN",
				"dns_names":      []string{"server3.example.com"},
			},
		},
	}

	batchReq := &BatchCertificateRequest{
		BatchID:    "batch_test_001",
		TemplateID: "tls-server",
		Requests:   singleRequests,
		Options: BatchOptions{
			MaxConcurrency: 2,
			RetryAttempts:  1,
			RetryDelay:    1 * time.Second,
			ContinueOnError: true,
		},
		RequestedBy: "test_user",
		RequestedAt: time.Now(),
	}

	// 批量签发
	result, err := manager.BatchIssueCertificate(batchReq)
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证结果
	assert.Equal(t, "batch_test_001", result.BatchID)
	assert.Equal(t, 3, result.Total)
	assert.Equal(t, 3, result.Success)
	assert.Equal(t, 0, result.Failed)
	assert.Len(t, result.Results, 3)
	assert.Greater(t, result.Duration, time.Duration(0))

	// 验证每个结果
	for _, taskResult := range result.Results {
		assert.True(t, taskResult.Success)
		assert.NotEmpty(t, taskResult.RequestID)
		assert.NotEmpty(t, taskResult.TaskID)
		assert.Greater(t, taskResult.Duration, time.Duration(0))
	}

	t.Logf("✅ 批量证书签发测试通过")
	t.Logf("   - 批次ID: %s", result.BatchID)
	t.Logf("   - 总数: %d", result.Total)
	t.Logf("   - 成功: %d", result.Success)
	t.Logf("   - 失败: %d", result.Failed)
	t.Logf("   - 耗时: %v", result.Duration)
}

// TestKeyPairGeneration 测试密钥对生成
func TestKeyPairGeneration(t *testing.T) {
	manager := setupTestX509Manager(t)

	testCases := []struct {
		name      string
		algorithm string
		keySize   int
	}{
		{"RSA-2048", "RSA", 2048},
		{"RSA-4096", "RSA", 4096},
		{"ECDSA-P256", "ECDSA", 256},
		{"ECDSA-P384", "ECDSA", 384},
		{"ECDSA-P521", "ECDSA", 521},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			privateKey, publicKey, err := manager.GenerateKeyPair(tc.algorithm, tc.keySize)
			require.NoError(t, err)
			require.NotNil(t, privateKey)
			require.NotNil(t, publicKey)

			// 验证密钥类型
			switch tc.algorithm {
			case "RSA":
				_, ok := privateKey.(*rsa.PrivateKey)
				assert.True(t, ok, "RSA私钥类型验证失败")
				_, ok = publicKey.(*rsa.PublicKey)
				assert.True(t, ok, "RSA公钥类型验证失败")
			case "ECDSA":
				_, ok := privateKey.(*ecdsa.PrivateKey)
				assert.True(t, ok, "ECDSA私钥类型验证失败")
				_, ok = publicKey.(*ecdsa.PublicKey)
				assert.True(t, ok, "ECDSA公钥类型验证失败")
			}

			t.Logf("✅ %s 密钥对生成测试通过", tc.name)
		})
	}
}

// TestCertificateRevocation 测试证书吊销
func TestCertificateRevocation(t *testing.T) {
	manager := setupTestX509Manager(t)

	// 先签发一个证书
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	req := &CertificateRequest{
		TemplateID: "tls-client",
		Subject: pkix.Name{
			CommonName:   "client.example.com",
			Organization: []string{"Example Corp"},
			Country:      []string{"CN"},
		},
		PublicKey:  &privateKey.PublicKey,
		Validity:   365 * 24 * time.Hour,
		KeyUsage:   x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		RequesterID: "test_user",
	}

	cert, err := manager.IssueCertificate(req)
	require.NoError(t, err)

	// 验证初始状态
	assert.Equal(t, CertificateStatusActive, cert.Status)
	assert.Nil(t, cert.RevokedAt)
	assert.Equal(t, 0, cert.RevocationReason)

	// 吊销证书
	err = manager.RevokeCertificate(cert.ID, 1) // keyCompromise
	require.NoError(t, err)

	// 验证吊销状态
	revokedCert, err := manager.GetCertificate(cert.ID)
	require.NoError(t, err)
	assert.Equal(t, CertificateStatusRevoked, revokedCert.Status)
	assert.NotNil(t, revokedCert.RevokedAt)
	assert.Equal(t, 1, revokedCert.RevocationReason)

	t.Logf("✅ 证书吊销测试通过")
	t.Logf("   - 证书ID: %s", cert.ID)
	t.Logf("   - 序列号: %s", cert.SerialNumber.String())
	t.Logf("   - 吊销时间: %s", revokedCert.RevokedAt.Format("2006-01-02 15:04:05"))
	t.Logf("   - 吊销原因: %d", revokedCert.RevocationReason)
}

// TestCertificateRenewal 测试证书续期
func TestCertificateRenewal(t *testing.T) {
	manager := setupTestX509Manager(t)

	// 先签发一个证书
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	req := &CertificateRequest{
		TemplateID: "tls-server",
		Subject: pkix.Name{
			CommonName:   "renewal.example.com",
			Organization: []string{"Example Corp"},
			Country:      []string{"CN"},
		},
		PublicKey:  &privateKey.PublicKey,
		Validity:   180 * 24 * time.Hour, // 6个月
		KeyUsage:   x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		RequesterID: "test_user",
	}

	oldCert, err := manager.IssueCertificate(req)
	require.NoError(t, err)

	// 验证初始状态
	assert.Equal(t, CertificateStatusActive, oldCert.Status)
	assert.Equal(t, 0, oldCert.RenewalCount)
	assert.Nil(t, oldCert.PreviousSerialNumber)

	// 续期证书
	newCert, err := manager.RenewCertificate(oldCert.ID, 365*24*time.Hour) // 续期为1年
	require.NoError(t, err)
	require.NotNil(t, newCert)

	// 验证续期结果
	assert.Equal(t, CertificateStatusActive, newCert.Status)
	assert.Equal(t, 1, newCert.RenewalCount)
	assert.Equal(t, oldCert.SerialNumber, newCert.PreviousSerialNumber)
	assert.True(t, newCert.NotAfter.After(oldCert.NotAfter)) // 新证书有效期更长

	// 验证原证书状态
	updatedOldCert, err := manager.GetCertificate(oldCert.ID)
	require.NoError(t, err)
	assert.Equal(t, CertificateStatusRenewed, updatedOldCert.Status)

	t.Logf("✅ 证书续期测试通过")
	t.Logf("   - 原证书ID: %s", oldCert.ID)
	t.Logf("   - 新证书ID: %s", newCert.ID)
	t.Logf("   - 原序列号: %s", oldCert.SerialNumber.String())
	t.Logf("   - 新序列号: %s", newCert.SerialNumber.String())
	t.Logf("   - 续期次数: %d", newCert.RenewalCount)
}

// TestTemplateManagement 测试模板管理
func TestTemplateManagement(t *testing.T) {
	manager := setupTestX509Manager(t)

	// 获取现有模板
	templates := manager.ListTemplates()
	assert.GreaterOrEqual(t, len(templates), 4)

	// 测试获取特定模板
	tlsServerTemplate, err := manager.GetTemplate("tls-server")
	require.NoError(t, err)
	assert.Equal(t, "tls-server", tlsServerTemplate.ID)
	assert.Equal(t, "TLS服务器证书", tlsServerTemplate.Name)
	assert.Equal(t, TemplateStatusActive, tlsServerTemplate.Status)

	// 测试不存在的模板
	_, err = manager.GetTemplate("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板不存在")

	t.Logf("✅ 模板管理测试通过")
	t.Logf("   - 模板数量: %d", len(templates))
	t.Logf("   - TLS服务器模板: %s", tlsServerTemplate.Name)
}

// TestSANExtension 测试主题备用名称扩展
func TestSANExtension(t *testing.T) {
	manager := setupTestX509Manager(t)

	// 生成密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// 创建包含多种SAN类型的证书请求
	req := &CertificateRequest{
		TemplateID: "tls-server",
		Subject: pkix.Name{
			CommonName:   "multisan.example.com",
			Organization: []string{"Example Corp"},
			Country:      []string{"CN"},
		},
		PublicKey:      &privateKey.PublicKey,
		Validity:       365 * 24 * time.Hour,
		KeyUsage:       x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:       []string{"multisan.example.com", "www.multisan.example.com"},
		IPAddresses:    []net.IP{net.ParseIP("192.168.1.1"), net.ParseIP("10.0.0.1")},
		EmailAddresses: []string{"admin@multisan.example.com", "support@multisan.example.com"},
		RequesterID:    "test_user",
	}

	// 签发证书
	cert, err := manager.IssueCertificate(req)
	require.NoError(t, err)

	// 验证SAN扩展
	assert.Contains(t, cert.Certificate.DNSNames, "multisan.example.com")
	assert.Contains(t, cert.Certificate.DNSNames, "www.multisan.example.com")
	assert.Contains(t, cert.Certificate.IPAddresses, net.ParseIP("192.168.1.1"))
	assert.Contains(t, cert.Certificate.IPAddresses, net.ParseIP("10.0.0.1"))
	assert.Contains(t, cert.Certificate.EmailAddresses, "admin@multisan.example.com")
	assert.Contains(t, cert.Certificate.EmailAddresses, "support@multisan.example.com")

	t.Logf("✅ SAN扩展测试通过")
	t.Logf("   - DNS名称数量: %d", len(cert.Certificate.DNSNames))
	t.Logf("   - IP地址数量: %d", len(cert.Certificate.IPAddresses))
	t.Logf("   - 邮箱地址数量: %d", len(cert.Certificate.EmailAddresses))
}

// TestCertificateValidation 测试证书验证
func TestCertificateValidation(t *testing.T) {
	manager := setupTestX509Manager(t)

	// 生成密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// 测试有效请求
	validReq := &CertificateRequest{
		TemplateID:  "tls-server",
		Subject: pkix.Name{
			CommonName:   "valid.example.com",
			Organization: []string{"Example Corp"},
			Country:      []string{"CN"},
		},
		PublicKey:  &privateKey.PublicKey,
		Validity:   365 * 24 * time.Hour,
		KeyUsage:   x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:   []string{"valid.example.com"},
		RequesterID: "test_user",
	}

	// 验证有效请求
	err = manager.validateCertificateRequest(validReq)
	assert.NoError(t, err)

	// 测试无效请求
	invalidReq := &CertificateRequest{
		TemplateID:  "tls-server",
		Subject: pkix.Name{
			CommonName:   "",
			Organization: []string{},
			Country:      []string{},
		},
		PublicKey:  nil, // 缺少公钥
		Validity:   0,   // 无效有效期
		KeyUsage:   0,   // 无密钥用法
		RequesterID: "test_user",
	}

	err = manager.validateCertificateRequest(invalidReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "公钥不能为空")

	t.Logf("✅ 证书验证测试通过")
}

// BenchmarkCertificateIssuance 证书签发性能测试
func BenchmarkCertificateIssuance(b *testing.B) {
	manager := setupTestX509Manager(&testing.T{})

	// 预生成密钥对
	privateKeys := make([]rsa.PrivateKey, b.N)
	for i := 0; i < b.N; i++ {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			b.Fatalf("生成密钥失败: %v", err)
		}
		privateKeys[i] = *key
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &CertificateRequest{
			TemplateID: "tls-server",
			Subject: pkix.Name{
				CommonName:   fmt.Sprintf("bench%d.example.com", i),
				Organization: []string{"Bench Corp"},
				Country:      []string{"CN"},
			},
			PublicKey:  &privateKeys[i].PublicKey,
			Validity:   365 * 24 * time.Hour,
			KeyUsage:   x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			RequesterID: "bench_user",
		}

		_, err := manager.IssueCertificate(req)
		if err != nil {
			b.Fatalf("签发证书失败: %v", err)
		}
	}
}

// BenchmarkKeyPairGeneration 密钥对生成性能测试
func BenchmarkKeyPairGeneration(b *testing.B) {
	manager := setupTestX509Manager(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := manager.GenerateKeyPair("RSA", 2048)
		if err != nil {
			b.Fatalf("生成密钥对失败: %v", err)
		}
	}
}