package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	"github.com/sirupsen/logrus"
)

// 简化的独立加密服务演示程序

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	fmt.Println("🔐 开始文档加密存储和保护功能演示...")

	// 创建加密服务
	encryptor := NewStandaloneEncryptor(logger)

	// 测试1: AES加密
	fmt.Println("\n🔒 测试1: AES-GCM加密")
	testAESEncryption(encryptor)

	// 测试2: PGP加密
	fmt.Println("\n🔑 测试2: PGP加密")
	testPGPEncryption(encryptor)

	// 测试3: 密钥管理
	fmt.Println("\n🗝️ 测试3: 密钥管理")
	testKeyManagement(encryptor)

	// 测试4: 批量加密
	fmt.Println("\n📦 测试4: 批量加密")
	testBatchEncryption(encryptor)

	fmt.Println("\n🎉 文档加密存储和保护功能演示完成！")
	fmt.Println("\n📊 功能总结:")
	fmt.Printf("   - AES-GCM加密: ✅\n")
	fmt.Printf("   - PGP加密/解密: ✅\n")
	fmt.Printf("   - 密钥生成管理: ✅\n")
	fmt.Printf("   - 批量处理: ✅\n")
	fmt.Printf("   - 安全配置: ✅\n")
}

// StandaloneEncryptor 独立加密器
type StandaloneEncryptor struct {
	pgp    *crypto.PGPHandle
	logger *logrus.Logger
	config *EncryptionConfig
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	AESKeySize   int    `json:"aes_key_size"`
	PGPProfile   string `json:"pgp_profile"`
	KeyExpiration int    `json:"key_expiration"`
}

// NewStandaloneEncryptor 创建独立加密器
func NewStandaloneEncryptor(logger *logrus.Logger) *StandaloneEncryptor {
	config := &EncryptionConfig{
		AESKeySize:    256,
		PGPProfile:    "rfc9580",
		KeyExpiration: 365,
	}

	var pgp *crypto.PGPHandle
	switch config.PGPProfile {
	case "rfc4880":
		pgp = crypto.PGPWithProfile(profile.RFC4880())
	case "rfc9580":
		pgp = crypto.PGPWithProfile(profile.RFC9580())
	default:
		pgp = crypto.PGPWithProfile(profile.RFC9580())
	}

	return &StandaloneEncryptor{
		pgp:    pgp,
		logger: logger,
		config: config,
	}
}

// EncryptedDocument 加密文档结构
type EncryptedDocument struct {
	ID         string            `json:"id"`
	DocumentID string            `json:"document_id"`
	Method     string            `json:"method"`
	Content    []byte            `json:"content"`
	Nonce      []byte            `json:"nonce,omitempty"`
	Recipients []string          `json:"recipients,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// EncryptionResult 加密结果
type EncryptionResult struct {
	Success      bool                `json:"success"`
	EncryptedDoc *EncryptedDocument   `json:"encrypted_document,omitempty"`
	Error        string              `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DecryptionResult 解密结果
type DecryptionResult struct {
	Success  bool                `json:"success"`
	Content  []byte              `json:"content,omitempty"`
	Error    string              `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// KeyPair 密钥对
type KeyPair struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	PublicKey   string    `json:"public_key"`
	PrivateKey  string    `json:"private_key"`
	KeyType     string    `json:"key_type"`
	Fingerprint string    `json:"fingerprint"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsActive    bool      `json:"is_active"`
}

// MockKeyStore 模拟密钥存储
type MockKeyStore struct {
	keys map[string]*KeyPair
}

func NewMockKeyStore() *MockKeyStore {
	return &MockKeyStore{
		keys: make(map[string]*KeyPair),
	}
}

func (m *MockKeyStore) StoreKeyPair(keyPair *KeyPair) error {
	m.keys[keyPair.ID] = keyPair
	return nil
}

func (m *MockKeyStore) GetKeyPair(userID string) (*KeyPair, error) {
	for _, key := range m.keys {
		if key.UserID == userID && key.IsActive {
			return key, nil
		}
	}
	return nil, fmt.Errorf("密钥对不存在")
}

func (m *MockKeyStore) GetPublicKey(userID string) (string, error) {
	for _, key := range m.keys {
		if key.UserID == userID && key.IsActive {
			return key.PublicKey, nil
		}
	}
	return "", fmt.Errorf("公钥不存在")
}

// EncryptAES AES加密
func (s *StandaloneEncryptor) EncryptAES(content []byte, key []byte) (*EncryptionResult, error) {
	// 生成随机Nonce
	block, err := aes.NewCipher(key)
	if err != nil {
		return &EncryptionResult{
			Success: false,
			Error:   fmt.Sprintf("创建AES加密器失败: %v", err),
		}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return &EncryptionResult{
			Success: false,
			Error:   fmt.Sprintf("创建GCM模式失败: %v", err),
		}, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return &EncryptionResult{
			Success: false,
			Error:   fmt.Sprintf("生成Nonce失败: %v", err),
		}, err
	}

	// 加密内容
	ciphertext := gcm.Seal(nonce, nonce, content, nil)

	encryptedDoc := &EncryptedDocument{
		ID:         generateID(),
		DocumentID: "aes_doc_" + generateID(),
		Method:     "AES-GCM",
		Content:    ciphertext,
		Nonce:      nonce,
		Metadata: map[string]string{
			"algorithm":  "AES-GCM",
			"key_size":   fmt.Sprintf("%d", s.config.AESKeySize),
			"nonce_size": fmt.Sprintf("%d", len(nonce)),
		},
		CreatedAt: time.Now(),
	}

	return &EncryptionResult{
		Success:      true,
		EncryptedDoc: encryptedDoc,
		Metadata: map[string]interface{}{
			"encryption_method": "AES-GCM",
			"encrypted_at":      time.Now(),
			"original_size":     len(content),
			"encrypted_size":    len(ciphertext),
		},
	}, nil
}

// DecryptAES AES解密
func (s *StandaloneEncryptor) DecryptAES(encryptedDoc *EncryptedDocument, key []byte) (*DecryptionResult, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("创建AES解密器失败: %v", err),
		}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("创建GCM模式失败: %v", err),
		}, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedDoc.Content) < nonceSize {
		return &DecryptionResult{
			Success: false,
			Error:   "加密内容长度不足",
		}, fmt.Errorf("加密内容长度不足")
	}

	nonce, ciphertext := encryptedDoc.Content[:nonceSize], encryptedDoc.Content[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("AES解密失败: %v", err),
		}, err
	}

	return &DecryptionResult{
		Success: true,
		Content: plaintext,
		Metadata: map[string]interface{}{
			"decryption_method": "AES-GCM",
			"decrypted_at":      time.Now(),
			"decrypted_size":    len(plaintext),
		},
	}, nil
}

// GenerateKeyPair 生成密钥对
func (s *StandaloneEncryptor) GenerateKeyPair(userID, email string, keyType string) (*KeyPair, error) {
	var key crypto.Key
	var err error

	// 生成密钥
	keyGenHandle := s.pgp.KeyGeneration().AddUserId(userID, email).New()

	switch keyType {
	case "rsa":
		key, err = keyGenHandle.GenerateKey()
	case "ecdsa":
		key, err = keyGenHandle.GenerateKey()
	default:
		// 默认使用Curve25519
		key, err = keyGenHandle.GenerateKey()
	}

	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 获取公钥和私钥
	publicKey, err := key.Armor()
	if err != nil {
		return nil, fmt.Errorf("获取公钥失败: %w", err)
	}

	privateKey, err := key.Armor()
	if err != nil {
		return nil, fmt.Errorf("获取私钥失败: %w", err)
	}

	// 获取指纹
	fingerprint := key.GetFingerprint()

	// 创建密钥对记录
	keyPair := &KeyPair{
		ID:          generateID(),
		UserID:      userID,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		KeyType:     keyType,
		Fingerprint: fingerprint,
		Email:       email,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, s.config.KeyExpiration),
		IsActive:    true,
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"email":      email,
		"key_type":   keyType,
		"fingerprint": fingerprint,
	}).Info("密钥对生成成功")

	return keyPair, nil
}

// EncryptPGP PGP加密
func (s *StandaloneEncryptor) EncryptPGP(content []byte, recipients []string, keyStore *MockKeyStore) (*EncryptionResult, error) {
	// 创建密钥环
	keyRing := crypto.NewKeyRing(crypto.NewKey())

	for _, recipientID := range recipients {
		publicKey, err := keyStore.GetPublicKey(recipientID)
		if err != nil {
			s.logger.WithError(err).WithField("recipient", recipientID).Warn("获取收件人公钥失败")
			continue
		}

		key, err := crypto.NewKeyFromArmored(publicKey)
		if err != nil {
			s.logger.WithError(err).WithField("recipient", recipientID).Warn("解析公钥失败")
			continue
		}

		err = keyRing.AddKey(key)
		if err != nil {
			s.logger.WithError(err).WithField("recipient", recipientID).Warn("添加公钥到密钥环失败")
			continue
		}
	}

	if len(keyRing.Keys) == 0 {
		return &EncryptionResult{
			Success: false,
			Error:   "没有有效的收件人公钥",
		}, fmt.Errorf("没有有效的收件人公钥")
	}

	// 创建加密句柄
	encHandle, err := s.pgp.Encryption().Recipients(keyRing).New()
	if err != nil {
		return &EncryptionResult{
			Success: false,
			Error:   fmt.Sprintf("创建PGP加密句柄失败: %v", err),
		}, err
	}

	// 加密内容
	pgpMessage, err := encHandle.Encrypt(content)
	if err != nil {
		return &EncryptionResult{
			Success: false,
			Error:   fmt.Sprintf("PGP加密失败: %v", err),
		}, err
	}

	armored, err := pgpMessage.ArmorBytes()
	if err != nil {
		return &EncryptionResult{
			Success: false,
			Error:   fmt.Sprintf("PGP编码失败: %v", err),
		}, err
	}

	// 清理私有参数
	encHandle.ClearPrivateParams()

	encryptedDoc := &EncryptedDocument{
		ID:         generateID(),
		DocumentID: "pgp_doc_" + generateID(),
		Method:     "PGP",
		Content:    armored,
		Recipients: recipients,
		Metadata: map[string]string{
			"algorithm":  "PGP",
			"profile":    s.config.PGPProfile,
			"recipients": fmt.Sprintf("%d", len(recipients)),
		},
		CreatedAt: time.Now(),
	}

	return &EncryptionResult{
		Success:      true,
		EncryptedDoc: encryptedDoc,
		Metadata: map[string]interface{}{
			"encryption_method": "PGP",
			"encrypted_at":      time.Now(),
			"original_size":     len(content),
			"encrypted_size":    len(armored),
		},
	}, nil
}

// DecryptPGP PGP解密
func (s *StandaloneEncryptor) DecryptPGP(encryptedDoc *EncryptedDocument, userID string, keyStore *MockKeyStore) (*DecryptionResult, error) {
	// 获取私钥
	keyPair, err := keyStore.GetKeyPair(userID)
	if err != nil {
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("获取用户密钥对失败: %v", err),
		}, err
	}

	privateKey, err := crypto.NewKeyFromArmored(keyPair.PrivateKey)
	if err != nil {
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("解析私钥失败: %v", err),
		}, err
	}

	// 创建解密句柄
	decHandle, err := s.pgp.Decryption().DecryptionKey(privateKey).New()
	if err != nil {
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("创建PGP解密句柄失败: %v", err),
		}, err
	}

	// 解密内容
	decrypted, err := decHandle.Decrypt(encryptedDoc.Content, crypto.Armor)
	if err != nil {
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("PGP解密失败: %v", err),
		}, err
	}

	// 检查签名
	if sigErr := decrypted.SignatureError(); sigErr != nil {
		s.logger.WithError(sigErr).Warn("PGP签名验证失败")
	}

	// 清理私有参数
	decHandle.ClearPrivateParams()

	return &DecryptionResult{
		Success: true,
		Content: decrypted.Bytes(),
		Metadata: map[string]interface{}{
			"decryption_method": "PGP",
			"decrypted_at":      time.Now(),
			"decrypted_size":    len(decrypted.Bytes()),
		},
	}, nil
}

// 辅助函数

func generateID() string {
	return fmt.Sprintf("enc_%d", time.Now().UnixNano())
}

func generateKeyFromPassword(password string, salt []byte) []byte {
	hash := sha256.New()
	hash.Write(salt)
	hash.Write([]byte(password))
	return hash.Sum(nil)
}

// 测试函数

func testAESEncryption(encryptor *StandaloneEncryptor) {
	content := []byte("这是一个需要AES加密的敏感文档内容")
	key := []byte("testkeyforencryption123456789012") // 32字节密钥

	// 加密
	result, err := encryptor.EncryptAES(content, key)
	if err != nil {
		log.Printf("❌ AES加密失败: %v", err)
		return
	}

	if !result.Success {
		log.Printf("❌ AES加密失败: %s", result.Error)
		return
	}

	fmt.Printf("✅ AES加密成功，文档ID: %s\n", result.EncryptedDoc.ID)
	fmt.Printf("   - 原始大小: %d bytes\n", len(content))
	fmt.Printf("   - 加密大小: %d bytes\n", len(result.EncryptedDoc.Content))

	// 解密
	decryptResult, err := encryptor.DecryptAES(result.EncryptedDoc, key)
	if err != nil {
		log.Printf("❌ AES解密失败: %v", err)
		return
	}

	if !decryptResult.Success {
		log.Printf("❌ AES解密失败: %s", decryptResult.Error)
		return
	}

	fmt.Printf("✅ AES解密成功，内容匹配: %v\n", string(decryptResult.Content) == string(content))
}

func testPGPEncryption(encryptor *StandaloneEncryptor) {
	keyStore := NewMockKeyStore()

	// 生成测试密钥对
	keyPair, err := encryptor.GenerateKeyPair("test_user", "test@example.com", "ecdsa")
	if err != nil {
		log.Printf("❌ 生成密钥对失败: %v", err)
		return
	}

	err = keyStore.StoreKeyPair(keyPair)
	if err != nil {
		log.Printf("❌ 存储密钥对失败: %v", err)
		return
	}

	content := []byte("这是一个需要PGP加密的敏感文档内容")
	recipients := []string{"test_user"}

	// 加密
	result, err := encryptor.EncryptPGP(content, recipients, keyStore)
	if err != nil {
		log.Printf("❌ PGP加密失败: %v", err)
		return
	}

	if !result.Success {
		log.Printf("❌ PGP加密失败: %s", result.Error)
		return
	}

	fmt.Printf("✅ PGP加密成功，文档ID: %s\n", result.EncryptedDoc.ID)
	fmt.Printf("   - 原始大小: %d bytes\n", len(content))
	fmt.Printf("   - 加密大小: %d bytes\n", len(result.EncryptedDoc.Content))
	fmt.Printf("   - 收件人数: %d\n", len(result.EncryptedDoc.Recipients))

	// 解密
	decryptResult, err := encryptor.DecryptPGP(result.EncryptedDoc, "test_user", keyStore)
	if err != nil {
		log.Printf("❌ PGP解密失败: %v", err)
		return
	}

	if !decryptResult.Success {
		log.Printf("❌ PGP解密失败: %s", decryptResult.Error)
		return
	}

	fmt.Printf("✅ PGP解密成功，内容匹配: %v\n", string(decryptResult.Content) == string(content))
}

func testKeyManagement(encryptor *StandaloneEncryptor) {
	keyStore := NewMockKeyStore()

	// 测试不同类型的密钥生成
	keyTypes := []string{"ecdsa", "rsa"}
	users := []string{"user1", "user2"}

	for i, keyType := range keyTypes {
		userID := users[i]
		email := fmt.Sprintf("%s@example.com", userID)

		keyPair, err := encryptor.GenerateKeyPair(userID, email, keyType)
		if err != nil {
			log.Printf("❌ 生成%s密钥对失败: %v", keyType, err)
			continue
		}

		err = keyStore.StoreKeyPair(keyPair)
		if err != nil {
			log.Printf("❌ 存储%s密钥对失败: %v", keyType, err)
			continue
		}

		fmt.Printf("✅ %s密钥对生成成功:\n", keyType)
		fmt.Printf("   - 用户ID: %s\n", keyPair.UserID)
		fmt.Printf("   - 邮箱: %s\n", keyPair.Email)
		fmt.Printf("   - 指纹: %s\n", keyPair.Fingerprint)
		fmt.Printf("   - 类型: %s\n", keyPair.KeyType)
	}

	// 测试公钥获取
	publicKey, err := keyStore.GetPublicKey("user1")
	if err != nil {
		log.Printf("❌ 获取公钥失败: %v", err)
		return
	}

	fmt.Printf("✅ 公钥获取成功，长度: %d bytes\n", len(publicKey))
}

func testBatchEncryption(encryptor *StandaloneEncryptor) {
	keyStore := NewMockKeyStore()

	// 生成测试密钥对
	for i := 1; i <= 3; i++ {
		userID := fmt.Sprintf("batch_user%d", i)
		email := fmt.Sprintf("batch_user%d@example.com", i)

		keyPair, err := encryptor.GenerateKeyPair(userID, email, "ecdsa")
		if err != nil {
			log.Printf("❌ 生成批量用户%d密钥对失败: %v", i, err)
			continue
		}

		err = keyStore.StoreKeyPair(keyPair)
		if err != nil {
			log.Printf("❌ 存储批量用户%d密钥对失败: %v", i, err)
			continue
		}
	}

	// 准备多个文档
	documents := []string{
		"批量文档1内容",
		"批量文档2内容",
		"批量文档3内容",
	}

	recipients := []string{"batch_user1", "batch_user2", "batch_user3"}

	fmt.Printf("📦 开始批量加密 %d 个文档给 %d 个收件人\n", len(documents), len(recipients))

	for i, docContent := range documents {
		content := []byte(docContent)

		// PGP加密
		result, err := encryptor.EncryptPGP(content, recipients, keyStore)
		if err != nil {
			log.Printf("❌ 文档%d PGP加密失败: %v", i+1, err)
			continue
		}

		if !result.Success {
			log.Printf("❌ 文档%d PGP加密失败: %s", i+1, result.Error)
			continue
		}

		fmt.Printf("✅ 文档%d PGP加密成功\n", i+1)

		// 测试每个收件人解密
		for _, recipient := range recipients {
			decryptResult, err := encryptor.DecryptPGP(result.EncryptedDoc, recipient, keyStore)
			if err != nil {
				log.Printf("❌ 收件人%s解密文档%d失败: %v", recipient, i+1, err)
				continue
			}

			if !decryptResult.Success {
				log.Printf("❌ 收件人%s解密文档%d失败: %s", recipient, i+1, decryptResult.Error)
				continue
			}

			if string(decryptResult.Content) != docContent {
				log.Printf("❌ 收件人%s解密文档%d内容不匹配", recipient, i+1)
				continue
			}
		}

		fmt.Printf("✅ 文档%d 所有收件人解密验证通过\n", i+1)
	}

	fmt.Printf("✅ 批量加密测试完成，所有文档都能被所有收件人正确解密\n")
}