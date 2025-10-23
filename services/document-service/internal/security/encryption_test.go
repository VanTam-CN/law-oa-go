package security

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sirupsen/logrus"
)

// MockKeyStore 模拟密钥存储
type MockKeyStore struct {
	keys map[string]*KeyPair
}

func NewMockKeyStore() *MockKeyStore {
	return &MockKeyStore{
		keys: make(map[string]*KeyPair),
	}
}

func (m *MockKeyStore) StoreKeyPair(ctx context.Context, keyPair *KeyPair) error {
	m.keys[keyPair.ID] = keyPair
	return nil
}

func (m *MockKeyStore) GetKeyPair(ctx context.Context, userID string) (*KeyPair, error) {
	for _, key := range m.keys {
		if key.UserID == userID && key.IsActive {
			return key, nil
		}
	}
	return nil, fmt.Errorf("密钥对不存在")
}

func (m *MockKeyStore) GetPublicKey(ctx context.Context, userID string) (string, error) {
	for _, key := range m.keys {
		if key.UserID == userID && key.IsActive {
			return key.PublicKey, nil
		}
	}
	return "", fmt.Errorf("公钥不存在")
}

func (m *MockKeyStore) ListKeys(ctx context.Context, userID string) ([]*KeyPair, error) {
	var keys []*KeyPair
	for _, key := range m.keys {
		if key.UserID == userID {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (m *MockKeyStore) DeleteKey(ctx context.Context, keyID string) error {
	delete(m.keys, keyID)
	return nil
}

func (m *MockKeyStore) UpdateKeyActivity(ctx context.Context, keyID string) error {
	// 注意：KeyPair结构体中没有UpdatedAt字段，这里简化处理
	return nil
}

// MockEncryptionCache 模拟加密缓存
type MockEncryptionCache struct {
	data map[string][]byte
}

func NewMockEncryptionCache() *MockEncryptionCache {
	return &MockEncryptionCache{
		data: make(map[string][]byte),
	}
}

func (m *MockEncryptionCache) Get(key string) ([]byte, bool) {
	value, exists := m.data[key]
	return value, exists
}

func (m *MockEncryptionCache) Set(key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockEncryptionCache) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockEncryptionCache) Clear() error {
	m.data = make(map[string][]byte)
	return nil
}

// 测试辅助函数
func setupTestEncryptor(t *testing.T) DocumentEncryptor {
	config := DefaultEncryptionConfig()
	keyStore := NewMockKeyStore()
	cache := NewMockEncryptionCache()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	encryptor, err := NewDocumentEncryptor(config, keyStore, cache, logger)
	require.NoError(t, err)
	return encryptor
}

func TestDefaultEncryptionConfig(t *testing.T) {
	config := DefaultEncryptionConfig()

	assert.Equal(t, 256, config.AESKeySize)
	assert.True(t, config.EnableAESGCM)
	assert.Equal(t, "rfc9580", config.PGPProfile)
	assert.Equal(t, 365, config.KeyExpiration)
	assert.Equal(t, 32, config.SaltSize)
	assert.Equal(t, 100000, config.IterationCount)
	assert.Equal(t, 1000, config.CacheSize)
	assert.Equal(t, 3600, config.CacheTTL)
}

func TestAESEncryption(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	content := []byte("这是一个需要加密的敏感文档内容")
	options := map[string]interface{}{
		"document_id": "test_doc_001",
		"key":         "dGVzdGtleWZvcmVuY3J5cHRpb24xMjM=", // base64编码的测试密钥
	}

	// 测试加密
	result, err := encryptor.Encrypt(ctx, content, EncryptionMethodAES, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.EncryptedDoc)
	assert.Equal(t, EncryptionMethodAES, result.EncryptedDoc.Method)
	assert.NotEmpty(t, result.EncryptedDoc.Content)
	assert.NotEmpty(t, result.EncryptedDoc.Nonce)
	assert.NotEmpty(t, result.EncryptedDoc.Salt)

	// 测试解密
	decryptResult, err := encryptor.Decrypt(ctx, result.EncryptedDoc, options)
	require.NoError(t, err)
	assert.True(t, decryptResult.Success)
	assert.Equal(t, content, decryptResult.Content)
}

func TestPGPEncryption(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 生成测试密钥对
	keyPair, err := encryptor.GenerateKeyPair(ctx, "test_user", "test@example.com", "ecdsa")
	require.NoError(t, err)

	content := []byte("这是一个需要PGP加密的敏感文档内容")
	options := map[string]interface{}{
		"document_id": "test_doc_002",
		"recipients":  []string{"test_user"},
	}

	// 测试加密
	result, err := encryptor.Encrypt(ctx, content, EncryptionMethodPGP, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.EncryptedDoc)
	assert.Equal(t, EncryptionMethodPGP, result.EncryptedDoc.Method)
	assert.NotEmpty(t, result.EncryptedDoc.Content)
	assert.NotEmpty(t, result.EncryptedDoc.Recipients)

	// 测试解密
	decryptOptions := map[string]interface{}{
		"user_id": "test_user",
	}

	decryptResult, err := encryptor.Decrypt(ctx, result.EncryptedDoc, decryptOptions)
	require.NoError(t, err)
	assert.True(t, decryptResult.Success)
	assert.Equal(t, content, decryptResult.Content)
}

func TestHybridEncryption(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 生成测试密钥对
	keyPair, err := encryptor.GenerateKeyPair(ctx, "test_user", "test@example.com", "ecdsa")
	require.NoError(t, err)

	content := []byte("这是一个需要混合加密的敏感文档内容")
	options := map[string]interface{}{
		"document_id": "test_doc_003",
		"recipients":  []string{"test_user"},
	}

	// 测试加密
	result, err := encryptor.Encrypt(ctx, content, EncryptionMethodHybrid, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.EncryptedDoc)
	assert.Equal(t, EncryptionMethodHybrid, result.EncryptedDoc.Method)
	assert.NotEmpty(t, result.EncryptedDoc.Content)
	assert.NotEmpty(t, result.EncryptedDoc.Recipients)

	// 注意：混合解密目前未完全实现，这里只测试加密
	t.Log("混合加密成功，解密功能待实现")
}

func TestKeyPairGeneration(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 测试ECDSA密钥生成
	keyPair, err := encryptor.GenerateKeyPair(ctx, "test_user", "test@example.com", "ecdsa")
	require.NoError(t, err)
	assert.NotEmpty(t, keyPair.ID)
	assert.Equal(t, "test_user", keyPair.UserID)
	assert.Equal(t, "test@example.com", keyPair.Email)
	assert.Equal(t, "ecdsa", keyPair.KeyType)
	assert.NotEmpty(t, keyPair.PublicKey)
	assert.NotEmpty(t, keyPair.PrivateKey)
	assert.NotEmpty(t, keyPair.Fingerprint)
	assert.True(t, keyPair.IsActive)

	// 测试RSA密钥生成
	rsaKeyPair, err := encryptor.GenerateKeyPair(ctx, "test_user", "test@example.com", "rsa")
	require.NoError(t, err)
	assert.Equal(t, "rsa", rsaKeyPair.KeyType)
	assert.NotEqual(t, keyPair.ID, rsaKeyPair.ID) // 确保生成不同的密钥
}

func TestPublicKeyImport(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 生成一个密钥对
	keyPair, err := encryptor.GenerateKeyPair(ctx, "user1", "user1@example.com", "ecdsa")
	require.NoError(t, err)

	// 获取公钥
	publicKey := keyPair.PublicKey

	// 为另一个用户导入公钥
	err = encryptor.ImportPublicKey(ctx, "user2", publicKey)
	require.NoError(t, err)

	// 验证公钥可以获取
	retrievedKey, err := encryptor.GetPublicKey(ctx, "user2")
	require.NoError(t, err)
	assert.Equal(t, publicKey, retrievedKey)
}

func TestBatchEncryption(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 生成测试密钥对
	keyPair, err := encryptor.GenerateKeyPair(ctx, "test_user", "test@example.com", "ecdsa")
	require.NoError(t, err)

	// 准备多个文档
	documents := [][]byte{
		[]byte("文档1内容"),
		[]byte("文档2内容"),
		[]byte("文档3内容"),
	}

	options := map[string]interface{}{
		"document_id": "batch_test",
		"recipients":  []string{"test_user"},
		"key":         "dGVzdGtleWZvcmVuY3J5cHRpb24xMjM=",
	}

	// 测试批量加密
	results, err := encryptor.BatchEncrypt(ctx, documents, EncryptionMethodAES, options)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	for i, result := range results {
		assert.True(t, result.Success, "文档 %d 加密失败", i)
		assert.NotNil(t, result.EncryptedDoc)
		assert.Equal(t, EncryptionMethodAES, result.EncryptedDoc.Method)
	}

	// 测试批量解密
	encryptedDocs := make([]*EncryptedDocument, len(results))
	for i, result := range results {
		encryptedDocs[i] = result.EncryptedDoc
	}

	decryptOptions := map[string]interface{}{
		"user_id": "test_user",
		"key":     "dGVzdGtleWZvcmVuY3J5cHRpb24xMjM=",
	}

	decryptResults, err := encryptor.BatchDecrypt(ctx, encryptedDocs, decryptOptions)
	require.NoError(t, err)
	assert.Len(t, decryptResults, 3)

	for i, result := range decryptResults {
		assert.True(t, result.Success, "文档 %d 解密失败", i)
		assert.Equal(t, documents[i], result.Content)
	}
}

func TestEncryptionCaching(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	content := []byte("测试缓存功能的文档内容")
	options := map[string]interface{}{
		"document_id": "cache_test_doc",
		"key":         "dGVzdGtleWZvcmVuY3J5cHRpb24xMjM=",
	}

	// 第一次加密
	result1, err := encryptor.Encrypt(ctx, content, EncryptionMethodAES, options)
	require.NoError(t, err)
	assert.True(t, result1.Success)

	// 第一次解密
	decryptResult1, err := encryptor.Decrypt(ctx, result1.EncryptedDoc, options)
	require.NoError(t, err)
	assert.True(t, decryptResult1.Success)
	assert.Equal(t, content, decryptResult1.Content)
	assert.Nil(t, decryptResult1.Metadata["from_cache"])

	// 第二次解密（应该从缓存获取）
	decryptResult2, err := encryptor.Decrypt(ctx, result1.EncryptedDoc, options)
	require.NoError(t, err)
	assert.True(t, decryptResult2.Success)
	assert.Equal(t, content, decryptResult2.Content)
	assert.NotNil(t, decryptResult2.Metadata["from_cache"])
	assert.True(t, decryptResult2.Metadata["from_cache"].(bool))
}

func TestEncryptionErrorHandling(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 测试不支持的加密方法
	content := []byte("测试内容")
	options := map[string]interface{}{
		"document_id": "error_test_doc",
	}

	result, err := encryptor.Encrypt(ctx, content, "unsupported_method", options)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "不支持的加密方法")

	// 测试缺少必要参数的PGP加密
	result, err = encryptor.Encrypt(ctx, content, EncryptionMethodPGP, options)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "缺少PGP收件人")

	// 测试缺少密钥的AES解密
	encryptedDoc := &EncryptedDocument{
		ID:         "test_id",
		DocumentID: "error_test_doc",
		Method:     EncryptionMethodAES,
		Content:    []byte("fake_encrypted_content"),
		Nonce:      []byte("123456789012"),
		Salt:       []byte("12345678901234567890123456789012"),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	decryptResult, err := encryptor.Decrypt(ctx, encryptedDoc, options)
	require.NoError(t, err)
	assert.False(t, decryptResult.Success)
	assert.Contains(t, decryptResult.Error, "缺少解密密钥")
}

func TestMultipleRecipientsPGPEncryption(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 生成多个测试密钥对
	users := []string{"user1", "user2", "user3"}
	for i, user := range users {
		email := fmt.Sprintf("user%d@example.com", i+1)
		_, err := encryptor.GenerateKeyPair(ctx, user, email, "ecdsa")
		require.NoError(t, err)
	}

	content := []byte("需要多收件人加密的文档内容")
	options := map[string]interface{}{
		"document_id": "multi_recipient_doc",
		"recipients":  users,
	}

	// 测试多收件人加密
	result, err := encryptor.Encrypt(ctx, content, EncryptionMethodPGP, options)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.NotNil(t, result.EncryptedDoc)
	assert.Equal(t, EncryptionMethodPGP, result.EncryptedDoc.Method)
	assert.NotEmpty(t, result.EncryptedDoc.Recipients)

	// 验证每个收件人都能解密
	for _, user := range users {
		decryptOptions := map[string]interface{}{
			"user_id": user,
		}

		decryptResult, err := encryptor.Decrypt(ctx, result.EncryptedDoc, decryptOptions)
		require.NoError(t, err)
		assert.True(t, decryptResult.Success)
		assert.Equal(t, content, decryptResult.Content)
	}
}

func TestPGPSigningAndVerification(t *testing.T) {
	encryptor := setupTestEncryptor(t)
	ctx := context.Background()

	// 生成签名者密钥对
	signerKeyPair, err := encryptor.GenerateKeyPair(ctx, "signer", "signer@example.com", "ecdsa")
	require.NoError(t, err)

	// 生成收件人密钥对
	recipientKeyPair, err := encryptor.GenerateKeyPair(ctx, "recipient", "recipient@example.com", "ecdsa")
	require.NoError(t, err)

	content := []byte("需要签名和加密的文档内容")
	options := map[string]interface{}{
		"document_id":      "signed_doc",
		"recipients":       []string{"recipient"},
		"signing_key_id":   "signer",
	}

	// 测试带签名的加密
	result, err := encryptor.Encrypt(ctx, content, EncryptionMethodPGP, options)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// 收件人解密并验证签名
	decryptOptions := map[string]interface{}{
		"user_id":           "recipient",
		"verification_key_id": "signer",
	}

	decryptResult, err := encryptor.Decrypt(ctx, result.EncryptedDoc, decryptOptions)
	require.NoError(t, err)
	assert.True(t, decryptResult.Success)
	assert.Equal(t, content, decryptResult.Content)
}

// 基准测试
func BenchmarkAESEncryption(b *testing.B) {
	encryptor := setupTestEncryptor(&testing.T{})
	ctx := context.Background()

	content := make([]byte, 1024*1024) // 1MB测试数据
	for i := range content {
		content[i] = byte(i % 256)
	}

	options := map[string]interface{}{
		"document_id": "benchmark_doc",
		"key":         "dGVzdGtleWZvcmVuY3J5cHRpb24xMjM=",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Encrypt(ctx, content, EncryptionMethodAES, options)
		if err != nil {
			b.Fatalf("加密失败: %v", err)
		}
	}
}

func BenchmarkPGPEncryption(b *testing.B) {
	encryptor := setupTestEncryptor(&testing.T{})
	ctx := context.Background()

	// 生成测试密钥对
	_, err := encryptor.GenerateKeyPair(ctx, "bench_user", "bench@example.com", "ecdsa")
	if err != nil {
		b.Fatalf("生成密钥对失败: %v", err)
	}

	content := make([]byte, 1024*1024) // 1MB测试数据
	for i := range content {
		content[i] = byte(i % 256)
	}

	options := map[string]interface{}{
		"document_id": "benchmark_doc",
		"recipients":  []string{"bench_user"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Encrypt(ctx, content, EncryptionMethodPGP, options)
		if err != nil {
			b.Fatalf("加密失败: %v", err)
		}
	}
}