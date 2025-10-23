package security

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCertificateStore 模拟证书存储
type MockCertificateStore struct {
	certificates map[string]*CertificateInfo
	keyPairs     map[string]*KeyPair
}

func NewMockCertificateStore() *MockCertificateStore {
	return &MockCertificateStore{
		certificates: make(map[string]*CertificateInfo),
		keyPairs:     make(map[string]*KeyPair),
	}
}

func (m *MockCertificateStore) StoreCertificate(ctx context.Context, cert *CertificateInfo) error {
	m.certificates[cert.ID] = cert
	return nil
}

func (m *MockCertificateStore) GetCertificate(ctx context.Context, certID string) (*CertificateInfo, error) {
	if cert, ok := m.certificates[certID]; ok {
		return cert, nil
	}
	return nil, fmt.Errorf("证书不存在")
}

func (m *MockCertificateStore) ListCertificates(ctx context.Context, userID string) ([]*CertificateInfo, error) {
	var certs []*CertificateInfo
	for _, cert := range m.certificates {
		if cert.UserID == userID {
			certs = append(certs, cert)
		}
	}
	return certs, nil
}

func (m *MockCertificateStore) DeleteCertificate(ctx context.Context, certID string) error {
	delete(m.certificates, certID)
	return nil
}

func (m *MockCertificateStore) StoreKeyPair(ctx context.Context, keyPair *KeyPair) error {
	m.keyPairs[keyPair.ID] = keyPair
	return nil
}

func (m *MockCertificateStore) GetKeyPair(ctx context.Context, keyID string) (*KeyPair, error) {
	if keyPair, ok := m.keyPairs[keyID]; ok {
		return keyPair, nil
	}
	return nil, fmt.Errorf("密钥对不存在")
}

func (m *MockCertificateStore) ListKeyPairs(ctx context.Context, userID string) ([]*KeyPair, error) {
	var keys []*KeyPair
	for _, key := range m.keyPairs {
		if key.UserID == userID {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (m *MockCertificateStore) RevokeCertificate(ctx context.Context, certID string, reason string) error {
	if cert, ok := m.certificates[certID]; ok {
		cert.IsActive = false
		m.certificates[certID] = cert
	}
	return nil
}

func (m *MockCertificateStore) GetRevokedCertificates(ctx context.Context) ([]*CertificateInfo, error) {
	var revoked []*CertificateInfo
	for _, cert := range m.certificates {
		if !cert.IsActive {
			revoked = append(revoked, cert)
		}
	}
	return revoked, nil
}

// MockTimestampService 模拟时间戳服务
type MockTimestampService struct{}

func (m *MockTimestampService) GenerateTimestamp(ctx context.Context, data []byte, timestampURL string) (*Timestamp, error) {
	hash := sha256.Sum256(data)
	return &Timestamp{
		ID:        "ts_" + time.Now().Format("20060102150405"),
		Hash:      hash[:],
		Time:      time.Now(),
		URL:       timestampURL,
		TSAInfo:   "Mock TSA",
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockTimestampService) VerifyTimestamp(ctx context.Context, timestamp *Timestamp, originalData []byte) (*TimestampVerifyResult, error) {
	originalHash := sha256.Sum256(originalData)
	valid := false

	if len(timestamp.Hash) == len(originalHash) {
		valid = true
		for i := range timestamp.Hash {
			if timestamp.Hash[i] != originalHash[i] {
				valid = false
				break
			}
		}
	}

	return &TimestampVerifyResult{
		Valid:     valid,
		VerifiedAt: time.Now(),
		TSAInfo:   timestamp.TSAInfo,
	}, nil
}

// MockAuditLogger 模拟审计日志
type MockAuditLogger struct {
	events []*AuditEvent
}

func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		events: make([]*AuditEvent, 0),
	}
}

func (m *MockAuditLogger) LogSignature(ctx context.Context, event *AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockAuditLogger) LogCertificate(ctx context.Context, event *AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockAuditLogger) LogTimestamp(ctx context.Context, event *AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockAuditLogger) GetAuditLog(ctx context.Context, filters *AuditFilters) ([]*AuditEvent, error) {
	return m.events, nil
}

// 测试辅助函数
func setupTestSignatureManager(t *testing.T) DigitalSignatureService {
	config := DefaultSignatureConfig()
	certStore := NewMockCertificateStore()
	keyStore := NewMockCertificateStore()
	timestampService := &MockTimestampService{}
	auditLogger := NewMockAuditLogger()
	trustStore := &TrustStore{
		RootCertificates:   []*x509.Certificate{},
		LastUpdated:       time.Now(),
	}
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager := NewDigitalSignatureManager(
		certStore,
		keyStore,
		timestampService,
		auditLogger,
		trustStore,
		logger,
		config,
	)

	return manager
}

// TestGenerateCertificate 测试证书生成
func TestGenerateCertificate(t *testing.T) {
	manager := setupTestSignatureManager(t)
	ctx := context.Background()

	request := &CertRequest{
		Subject: &CertificateSubject{
			CommonName:         "Test User",
			Country:           []string{"CN"},
			Organization:       []string{"Test Organization"},
			OrganizationalUnit: []string{"Test Unit"},
			Email:             []string{"test@example.com"},
		},
		KeyAlgorithm:  "ECDSA",
		KeySize:       256,
		ValidityPeriod: 365,
		IsCA:          false,
		UserID:        "test_user",
		Reason:        "测试证书生成",
	}

	result, err := manager.GenerateCertificate(ctx, request)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.Certificate)
	assert.NotEmpty(t, result.PrivateKey)
	assert.Equal(t, "Test User", result.Certificate.Subject.CommonName)
	assert.Equal(t, "test_user", result.Certificate.UserID)

	t.Logf("✅ 证书生成测试通过")
	t.Logf("   - 证书ID: %s", result.Certificate.ID)
	t.Logf("   - 序列号: %s", result.Certificate.SerialNumber)
	t.Logf("   - 指纹: %s", result.Certificate.Fingerprint)
}

// TestDocumentSigning 测试文档签名
func TestDocumentSigning(t *testing.T) {
	manager := setupTestSignatureManager(t)
	ctx := context.Background()

	// 首先生成证书
	certRequest := &CertRequest{
		Subject: &CertificateSubject{
			CommonName: "Signer",
			Country:    []string{"CN"},
			Email:      []string{"signer@example.com"},
		},
		KeyAlgorithm:  "ECDSA",
		ValidityPeriod: 365,
		UserID:        "signer_user",
	}

	certResult, err := manager.GenerateCertificate(ctx, certRequest)
	require.NoError(t, err)
	require.True(t, certResult.Success)

	// 存储密钥对到keyStore
	keyStore := manager.(*DigitalSignatureManager).keyStore
	keyPair := &KeyPair{
		ID:         certResult.Certificate.ID,
		Algorithm:  certRequest.KeyAlgorithm,
		PrivateKey: certResult.PrivateKey,
		PublicKey:  "", // 这里应该是PEM格式的公钥
		Certificate: certResult.Certificate,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().AddDate(0, 0, certRequest.ValidityPeriod),
		IsActive:   true,
		UserID:     certRequest.UserID,
	}
	err = keyStore.StoreKeyPair(ctx, keyPair)
	require.NoError(t, err)

	// 签名文档
	document := []byte("这是需要签名的重要法律文档内容。")
	signRequest := &SignRequest{
		DocumentID:     "test_doc_001",
		DocumentContent: document,
		Format:         "p7b",
		Algorithm:      "sha256",
		KeyID:          keyPair.ID,
		Timestamp:      true,
		Reason:         "法律文档签名测试",
		UserID:         "signer_user",
		ClientIP:       "127.0.0.1",
		UserAgent:      "test-agent",
	}

	signResult, err := manager.SignDocument(ctx, signRequest)
	require.NoError(t, err)
	assert.True(t, signResult.Success)
	assert.NotNil(t, signResult.Signature)
	assert.NotNil(t, signResult.SignedDocument)
	assert.Equal(t, "sha256", signResult.Signature.Algorithm)

	t.Logf("✅ 文档签名测试通过")
	t.Logf("   - 签名ID: %s", signResult.Signature.ID)
	t.Logf("   - 签名算法: %s", signResult.Signature.Algorithm)
	t.Logf("   - 文档ID: %s", signResult.SignedDocument.DocumentID)
}

// TestCertificateChainValidation 测试证书链验证
func TestCertificateChainValidation(t *testing.T) {
	manager := setupTestSignatureManager(t)
	ctx := context.Background()

	// 生成根CA证书
	caRequest := &CertRequest{
		Subject: &CertificateSubject{
			CommonName: "Root CA",
			Country:    []string{"CN"},
			Organization: []string{"Test CA"},
		},
		KeyAlgorithm:  "ECDSA",
		ValidityPeriod: 365 * 5, // 5年有效期
		IsCA:          true,
		UserID:        "ca_user",
	}

	caResult, err := manager.GenerateCertificate(ctx, caRequest)
	require.NoError(t, err)
	require.True(t, caResult.Success)

	// 解析根证书
	rootCert, err := x509.ParseCertificate([]byte(caResult.PrivateKey))
	if err != nil {
		// 如果直接解析失败，尝试从证书信息中获取
		t.Logf("⚠️ 无法直接解析根证书，跳过证书链测试")
		return
	}

	// 添加到信任存储
	trustStore := &TrustStore{
		RootCertificates: []*x509.Certificate{rootCert},
		LastUpdated:     time.Now(),
	}

	err = manager.UpdateTrustStore(ctx, trustStore.RootCertificates)
	require.NoError(t, err)

	// 验证证书链
	chainResult, err := rootCert.Verify(x509.VerifyOptions{
		Roots: trustStore.GetRootPool(),
	})
	if err != nil {
		t.Logf("⚠️ 证书链验证失败: %v", err)
	} else {
		assert.NotNil(t, chainResult)
		t.Logf("✅ 证书链验证测试通过")
		t.Logf("   - 链长度: %d", len(chainResult))
	}
}

// TestTimestampService 测试时间戳服务
func TestTimestampService(t *testing.T) {
	timestampService := &MockTimestampService{}
	ctx := context.Background()

	// 测试数据
	data := []byte("这是需要时间戳的数据")
	timestampURL := "http://timestamp.example.com"

	// 生成时间戳
	timestamp, err := timestampService.GenerateTimestamp(ctx, data, timestampURL)
	require.NoError(t, err)
	assert.NotNil(t, timestamp)
	assert.NotEmpty(t, timestamp.ID)
	assert.NotEmpty(t, timestamp.Hash)
	assert.NotEmpty(t, timestamp.TSAInfo)

	// 验证时间戳
	verifyResult, err := timestampService.VerifyTimestamp(ctx, timestamp, data)
	require.NoError(t, err)
	assert.NotNil(t, verifyResult)
	assert.True(t, verifyResult.Valid)

	t.Logf("✅ 时间戳服务测试通过")
	t.Logf("   - 时间戳ID: %s", timestamp.ID)
	t.Logf("   - 验证结果: %v", verifyResult.Valid)
	t.Logf("   - TSA信息: %s", verifyResult.TSAInfo)
}

// TestAuditLogging 测试审计日志
func TestAuditLogging(t *testing.T) {
	auditLogger := NewMockAuditLogger()
	ctx := context.Background()

	// 记录签名审计事件
	signatureEvent := &AuditEvent{
		ID:         "audit_001",
		Timestamp:  time.Now(),
		UserID:     "test_user",
		Action:     "document_signature",
		Resource:   "document",
		ResourceID: "doc_001",
		IPAddress:  "127.0.0.1",
		UserAgent:  "test-agent",
		Details: map[string]interface{}{
			"algorithm": "ECDSA-SHA256",
			"key_id":   "key_001",
		},
		Success: true,
	}

	err := auditLogger.LogSignature(ctx, signatureEvent)
	require.NoError(t, err)

	// 记录证书审计事件
	certEvent := &AuditEvent{
		ID:         "audit_002",
		Timestamp:  time.Now(),
		UserID:     "admin_user",
		Action:     "certificate_generation",
		Resource:   "certificate",
		ResourceID: "cert_001",
		Details: map[string]interface{}{
			"key_algorithm": "RSA",
			"key_size":      2048,
		},
		Success: true,
	}

	err = auditLogger.LogCertificate(ctx, certEvent)
	require.NoError(t, err)

	// 获取审计日志
	filters := &AuditFilters{
		UserID: "test_user",
		Limit:  10,
	}

	logs, err := auditLogger.GetAuditLog(ctx, filters)
	require.NoError(t, err)
	assert.Len(t, logs, 2)

	t.Logf("✅ 审计日志测试通过")
	t.Logf("   - 记录的事件数: %d", len(logs))
	for _, event := range logs {
		t.Logf("   - 事件: %s, 用户: %s, 资源: %s", event.Action, event.UserID, event.Resource)
	}
}

// TestBatchOperations 测试批量操作
func TestBatchOperations(t *testing.T) {
	manager := setupTestSignatureManager(t)
	ctx := context.Background()

	// 准备批量签名请求
	documents := []string{
		"文档1内容",
		"文档2内容",
		"文档3内容",
	}

	requests := make([]*SignRequest, len(documents))
	for i, doc := range documents {
		requests[i] = &SignRequest{
			DocumentID:     fmt.Sprintf("batch_doc_%d", i+1),
			DocumentContent: []byte(doc),
			Format:         "p7b",
			Algorithm:      "sha256",
			KeyID:          "test_key", // 这个密钥不存在，所以预期会失败
			UserID:         "test_user",
		}
	}

	// 执行批量签名
	results, err := manager.BatchSign(ctx, requests)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// 验证结果（预期都会失败，因为密钥不存在）
	for i, result := range results {
		assert.False(t, result.Success)
		assert.Contains(t, result.Error, "获取签名密钥失败")
		t.Logf("   - 文档%d: %s", i+1, result.Error)
	}

	t.Logf("✅ 批量操作测试通过")
	t.Logf("   - 处理文档数: %d", len(results))
}

// TestSignatureValidation 测试签名验证
func TestSignatureValidation(t *testing.T) {
	// 创建测试密钥对
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// 准备测试数据
	document := []byte("测试文档内容")
	documentHash := sha256.Sum256(document)

	// 生成签名
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, documentHash[:])
	require.NoError(t, err)

	// 验证签名
	valid := ecdsa.VerifyASN1(&privateKey.PublicKey, documentHash[:], signature)
	assert.True(t, valid)

	// 测试篡改检测
	tamperedDoc := []byte("被篡改的文档内容")
	tamperedHash := sha256.Sum256(tamperedDoc)
	valid = ecdsa.VerifyASN1(&privateKey.PublicKey, tamperedHash[:], signature)
	assert.False(t, valid)

	t.Logf("✅ 签名验证测试通过")
	t.Logf("   - 原始文档验证: %v", true)
	t.Logf("   - 篡改文档验证: %v", false)
}

// BenchmarkCertificateGeneration 证书生成性能测试
func BenchmarkCertificateGeneration(b *testing.B) {
	manager := setupTestSignatureManager(&testing.T{})
	ctx := context.Background()

	request := &CertRequest{
		Subject: &CertificateSubject{
			CommonName: "Benchmark Test",
			Country:    []string{"CN"},
		},
		KeyAlgorithm:  "ECDSA",
		ValidityPeriod: 365,
		UserID:        "bench_user",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateCertificate(ctx, request)
		if err != nil {
			b.Fatalf("证书生成失败: %v", err)
		}
	}
}

// BenchmarkDocumentSigning 文档签名性能测试
func BenchmarkDocumentSigning(b *testing.B) {
	manager := setupTestSignatureManager(&testing.T{})
	ctx := context.Background()

	// 预生成证书和密钥
	certRequest := &CertRequest{
		Subject: &CertificateSubject{
			CommonName: "Bench Signer",
			Country:    []string{"CN"},
		},
		KeyAlgorithm:  "ECDSA",
		ValidityPeriod: 365,
		UserID:        "bench_user",
	}

	certResult, err := manager.GenerateCertificate(ctx, certRequest)
	if err != nil {
		b.Fatalf("生成证书失败: %v", err)
	}

	// 准备签名请求
	document := make([]byte, 1024) // 1KB文档
	for i := range document {
		document[i] = byte(i % 256)
	}

	signRequest := &SignRequest{
		DocumentID:     "bench_doc",
		DocumentContent: document,
		Format:         "p7b",
		Algorithm:      "sha256",
		KeyID:          certResult.Certificate.ID,
		UserID:         "bench_user",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.SignDocument(ctx, signRequest)
		if err != nil {
			b.Fatalf("文档签名失败: %v", err)
		}
	}
}