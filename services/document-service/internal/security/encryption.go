package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	"github.com/sirupsen/logrus"
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	// AES配置
	AESKeySize    int    `yaml:"aes_key_size" json:"aes_key_size"`       // AES密钥长度，128/192/256
	EnableAESGCM  bool   `yaml:"enable_aes_gcm" json:"enable_aes_gcm"`   // 是否启用AES-GCM模式

	// PGP配置
	PGPProfile    string `yaml:"pgp_profile" json:"pgp_profile"`         // PGP配置文件 (rfc4880/rfc9580)
	KeyExpiration int    `yaml:"key_expiration" json:"key_expiration"`   // 密钥过期时间（天）

	// 安全配置
	SaltSize      int    `yaml:"salt_size" json:"salt_size"`             // 盐值长度
	IterationCount int   `yaml:"iteration_count" json:"iteration_count"` // 迭代次数

	// 缓存配置
	CacheSize     int    `yaml:"cache_size" json:"cache_size"`           // 缓存大小
	CacheTTL      int    `yaml:"cache_ttl" json:"cache_ttl"`             // 缓存TTL（秒）
}

// DefaultEncryptionConfig 默认加密配置
func DefaultEncryptionConfig() *EncryptionConfig {
	return &EncryptionConfig{
		AESKeySize:     256,
		EnableAESGCM:   true,
		PGPProfile:     "rfc9580",
		KeyExpiration:  365, // 1年
		SaltSize:       32,
		IterationCount: 100000,
		CacheSize:      1000,
		CacheTTL:       3600, // 1小时
	}
}

// EncryptionMethod 加密方法
type EncryptionMethod string

const (
	EncryptionMethodAES     EncryptionMethod = "aes"
	EncryptionMethodPGP     EncryptionMethod = "pgp"
	EncryptionMethodHybrid  EncryptionMethod = "hybrid"
)

// EncryptedDocument 加密文档
type EncryptedDocument struct {
	ID           string            `json:"id" bson:"_id"`
	DocumentID   string            `json:"document_id" bson:"document_id"`
	Method       EncryptionMethod   `json:"method" bson:"method"`
	Content      []byte            `json:"content" bson:"content"`
	Nonce        []byte            `json:"nonce,omitempty" bson:"nonce,omitempty"`
	Salt         []byte            `json:"salt,omitempty" bson:"salt,omitempty"`
	KeyID        string            `json:"key_id,omitempty" bson:"key_id,omitempty"`
	Recipients   []string          `json:"recipients,omitempty" bson:"recipients,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" bson:"updated_at"`
}

// EncryptionResult 加密结果
type EncryptionResult struct {
	Success      bool              `json:"success"`
	EncryptedDoc *EncryptedDocument `json:"encrypted_document,omitempty"`
	Error        string            `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DecryptionResult 解密结果
type DecryptionResult struct {
	Success    bool        `json:"success"`
	Content    []byte      `json:"content,omitempty"`
	Error      string      `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	KeyID      string      `json:"key_id,omitempty"`
}

// KeyPair 密钥对
type KeyPair struct {
	ID          string    `json:"id" bson:"_id"`
	UserID      string    `json:"user_id" bson:"user_id"`
	PublicKey   string    `json:"public_key" bson:"public_key"`
	PrivateKey  string    `json:"private_key" bson:"private_key"`
	KeyType     string    `json:"key_type" bson:"key_type"` // rsa/ecdsa
	KeySize     int       `json:"key_size" bson:"key_size"`
	Fingerprint string    `json:"fingerprint" bson:"fingerprint"`
	Email       string    `json:"email" bson:"email"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	ExpiresAt   time.Time `json:"expires_at" bson:"expires_at"`
	IsActive    bool      `json:"is_active" bson:"is_active"`
}

// DocumentEncryptor 文档加密器接口
type DocumentEncryptor interface {
	// 基本加密解密
	Encrypt(ctx context.Context, content []byte, method EncryptionMethod, options map[string]interface{}) (*EncryptionResult, error)
	Decrypt(ctx context.Context, encryptedDoc *EncryptedDocument, options map[string]interface{}) (*DecryptionResult, error)

	// PGP特定方法
	GenerateKeyPair(ctx context.Context, userID, email string, keyType string) (*KeyPair, error)
	ImportPublicKey(ctx context.Context, userID string, publicKey string) error
	GetPublicKey(ctx context.Context, userID string) (string, error)

	// 批量操作
	BatchEncrypt(ctx context.Context, documents [][]byte, method EncryptionMethod, options map[string]interface{}) ([]*EncryptionResult, error)
	BatchDecrypt(ctx context.Context, encryptedDocs []*EncryptedDocument, options map[string]interface{}) ([]*DecryptionResult, error)
}

// documentEncryptor 文档加密器实现
type documentEncryptor struct {
	config    *EncryptionConfig
	pgp       *crypto.PGP
	logger    *logrus.Logger
	keyStore  KeyStore
	cache     EncryptionCache
}

// KeyStore 密钥存储接口
type KeyStore interface {
	StoreKeyPair(ctx context.Context, keyPair *KeyPair) error
	GetKeyPair(ctx context.Context, userID string) (*KeyPair, error)
	GetPublicKey(ctx context.Context, userID string) (string, error)
	ListKeys(ctx context.Context, userID string) ([]*KeyPair, error)
	DeleteKey(ctx context.Context, keyID string) error
	UpdateKeyActivity(ctx context.Context, keyID string) error
}

// EncryptionCache 加密缓存接口
type EncryptionCache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration) error
	Delete(key string) error
	Clear() error
}

// NewDocumentEncryptor 创建文档加密器
func NewDocumentEncryptor(config *EncryptionConfig, keyStore KeyStore, cache EncryptionCache, logger *logrus.Logger) (DocumentEncryptor, error) {
	if config == nil {
		config = DefaultEncryptionConfig()
	}

	var pgpProfile *profile.Profile
	switch config.PGPProfile {
	case "rfc4880":
		pgpProfile = profile.RFC4880()
	case "rfc9580":
		pgpProfile = profile.RFC9580()
	default:
		pgpProfile = profile.RFC9580() // 默认使用最新标准
	}

	pgp := crypto.PGPWithProfile(pgpProfile)

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &documentEncryptor{
		config:   config,
		pgp:      pgp,
		logger:   logger,
		keyStore: keyStore,
		cache:    cache,
	}, nil
}

// Encrypt 加密文档
func (de *documentEncryptor) Encrypt(ctx context.Context, content []byte, method EncryptionMethod, options map[string]interface{}) (*EncryptionResult, error) {
	startTime := time.Now()

	defer func() {
		de.logger.WithFields(logrus.Fields{
			"method":    method,
			"duration":  time.Since(startTime),
			"size":      len(content),
		}).Info("Document encryption completed")
	}()

	var encryptedDoc *EncryptedDocument
	var err error

	switch method {
	case EncryptionMethodAES:
		encryptedDoc, err = de.encryptAES(ctx, content, options)
	case EncryptionMethodPGP:
		encryptedDoc, err = de.encryptPGP(ctx, content, options)
	case EncryptionMethodHybrid:
		encryptedDoc, err = de.encryptHybrid(ctx, content, options)
	default:
		return &EncryptionResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的加密方法: %s", method),
		}, nil
	}

	if err != nil {
		de.logger.WithError(err).Error("文档加密失败")
		return &EncryptionResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	// 缓存加密结果
	if de.cache != nil {
		cacheKey := de.generateCacheKey(encryptedDoc.DocumentID, method)
		de.cache.Set(cacheKey, encryptedDoc.Content, time.Duration(de.config.CacheTTL)*time.Second)
	}

	return &EncryptionResult{
		Success:      true,
		EncryptedDoc: encryptedDoc,
		Metadata: map[string]interface{}{
			"encryption_method": method,
			"encrypted_at":      time.Now(),
			"original_size":     len(content),
			"encrypted_size":    len(encryptedDoc.Content),
		},
	}, nil
}

// Decrypt 解密文档
func (de *documentEncryptor) Decrypt(ctx context.Context, encryptedDoc *EncryptedDocument, options map[string]interface{}) (*DecryptionResult, error) {
	startTime := time.Now()

	defer func() {
		de.logger.WithFields(logrus.Fields{
			"method":    encryptedDoc.Method,
			"duration":  time.Since(startTime),
			"doc_id":    encryptedDoc.DocumentID,
		}).Info("Document decryption completed")
	}()

	// 检查缓存
	if de.cache != nil {
		cacheKey := de.generateCacheKey(encryptedDoc.DocumentID, encryptedDoc.Method)
		if cached, found := de.cache.Get(cacheKey); found {
			return &DecryptionResult{
				Success:  true,
				Content:  cached,
				KeyID:    encryptedDoc.KeyID,
				Metadata: map[string]interface{}{"from_cache": true},
			}, nil
		}
	}

	var content []byte
	var err error

	switch encryptedDoc.Method {
	case EncryptionMethodAES:
		content, err = de.decryptAES(ctx, encryptedDoc, options)
	case EncryptionMethodPGP:
		content, err = de.decryptPGP(ctx, encryptedDoc, options)
	case EncryptionMethodHybrid:
		content, err = de.decryptHybrid(ctx, encryptedDoc, options)
	default:
		return &DecryptionResult{
			Success: false,
			Error:   fmt.Sprintf("不支持的解密方法: %s", encryptedDoc.Method),
		}, nil
	}

	if err != nil {
		de.logger.WithError(err).Error("文档解密失败")
		return &DecryptionResult{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &DecryptionResult{
		Success: true,
		Content: content,
		KeyID:   encryptedDoc.KeyID,
		Metadata: map[string]interface{}{
			"decryption_method": encryptedDoc.Method,
			"decrypted_at":      time.Now(),
			"decrypted_size":    len(content),
		},
	}, nil
}

// encryptAES AES加密
func (de *documentEncryptor) encryptAES(ctx context.Context, content []byte, options map[string]interface{}) (*EncryptedDocument, error) {
	// 生成随机密钥
	key := make([]byte, de.config.AESKeySize/8)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("生成AES密钥失败: %w", err)
	}

	// 生成随机盐值
	salt := make([]byte, de.config.SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("生成盐值失败: %w", err)
	}

	// 生成随机Nonce
	var nonce []byte
	if de.config.EnableAESGCM {
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("创建AES加密器失败: %w", err)
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("创建GCM模式失败: %w", err)
		}

		nonce = make([]byte, gcm.NonceSize())
		if _, err := rand.Read(nonce); err != nil {
			return nil, fmt.Errorf("生成Nonce失败: %w", err)
		}

		// 加密内容
		ciphertext := gcm.Seal(nonce, nonce, content, nil)

		// 创建加密文档
		docID := options["document_id"].(string)
		return &EncryptedDocument{
			ID:         de.generateID(),
			DocumentID: docID,
			Method:     EncryptionMethodAES,
			Content:    ciphertext,
			Nonce:      nonce,
			Salt:       salt,
			Metadata: map[string]string{
				"algorithm":    "AES-GCM",
				"key_size":     fmt.Sprintf("%d", de.config.AESKeySize),
				"nonce_size":   fmt.Sprintf("%d", len(nonce)),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	// 如果不使用GCM，使用AES-CBC模式
	return nil, fmt.Errorf("目前仅支持AES-GCM模式")
}

// decryptAES AES解密
func (de *documentEncryptor) decryptAES(ctx context.Context, encryptedDoc *EncryptedDocument, options map[string]interface{}) ([]byte, error) {
	if !de.config.EnableAESGCM {
		return nil, fmt.Errorf("AES-GCM模式未启用")
	}

	// 从选项中获取密钥或从缓存中获取
	var key []byte
	if keyOption, exists := options["key"]; exists {
		if keyBytes, ok := keyOption.([]byte); ok {
			key = keyBytes
		} else if keyStr, ok := keyOption.(string); ok {
			var err error
			key, err = base64.StdEncoding.DecodeString(keyStr)
			if err != nil {
				return nil, fmt.Errorf("密钥格式错误: %w", err)
			}
		} else {
			return nil, fmt.Errorf("密钥类型错误")
		}
	} else {
		return nil, fmt.Errorf("缺少解密密钥")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES解密器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM模式失败: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedDoc.Content) < nonceSize {
		return nil, fmt.Errorf("加密内容长度不足")
	}

	nonce, ciphertext := encryptedDoc.Content[:nonceSize], encryptedDoc.Content[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("AES解密失败: %w", err)
	}

	return plaintext, nil
}

// encryptPGP PGP加密
func (de *documentEncryptor) encryptPGP(ctx context.Context, content []byte, options map[string]interface{}) (*EncryptedDocument, error) {
	// 获取收件人公钥
	recipients, ok := options["recipients"].([]string)
	if !ok || len(recipients) == 0 {
		return nil, fmt.Errorf("缺少PGP收件人")
	}

	keyRing := crypto.NewKeyRing()

	for _, recipientID := range recipients {
		publicKey, err := de.keyStore.GetPublicKey(ctx, recipientID)
		if err != nil {
			de.logger.WithError(err).WithField("recipient", recipientID).Warn("获取收件人公钥失败")
			continue
		}

		key, err := crypto.NewKeyFromArmored(publicKey)
		if err != nil {
			de.logger.WithError(err).WithField("recipient", recipientID).Warn("解析公钥失败")
			continue
		}

		err = keyRing.AddKey(key)
		if err != nil {
			de.logger.WithError(err).WithField("recipient", recipientID).Warn("添加公钥到密钥环失败")
			continue
		}
	}

	if len(keyRing.Keys) == 0 {
		return nil, fmt.Errorf("没有有效的收件人公钥")
	}

	// 创建加密句柄
	encHandle, err := de.pgp.Encryption().Recipients(keyRing).New()
	if err != nil {
		return nil, fmt.Errorf("创建PGP加密句柄失败: %w", err)
	}

	// 如果需要签名
	if signingKeyID, exists := options["signing_key_id"]; exists {
		if keyPair, err := de.keyStore.GetKeyPair(ctx, signingKeyID.(string)); err == nil {
			privateKey, err := crypto.NewPrivateKeyFromArmored(keyPair.PrivateKey, nil)
			if err == nil {
				encHandle.SigningKey(privateKey)
			}
		}
	}

	// 加密内容
	pgpMessage, err := encHandle.Encrypt(content)
	if err != nil {
		return nil, fmt.Errorf("PGP加密失败: %w", err)
	}

	armored, err := pgpMessage.ArmorBytes()
	if err != nil {
		return nil, fmt.Errorf("PGP编码失败: %w", err)
	}

	// 清理私有参数
	encHandle.ClearPrivateParams()

	docID := options["document_id"].(string)
	return &EncryptedDocument{
		ID:         de.generateID(),
		DocumentID: docID,
		Method:     EncryptionMethodPGP,
		Content:    armored,
		Recipients: recipients,
		Metadata: map[string]string{
			"algorithm":  "PGP",
			"profile":    de.config.PGPProfile,
			"recipients": fmt.Sprintf("%d", len(recipients)),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// decryptPGP PGP解密
func (de *documentEncryptor) decryptPGP(ctx context.Context, encryptedDoc *EncryptedDocument, options map[string]interface{}) ([]byte, error) {
	// 获取私钥
	userID, ok := options["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("缺少用户ID")
	}

	keyPair, err := de.keyStore.GetKeyPair(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户密钥对失败: %w", err)
	}

	privateKey, err := crypto.NewKeyFromArmored(keyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	// 创建解密句柄
	decHandle, err := de.pgp.Decryption().DecryptionKey(privateKey).New()
	if err != nil {
		return nil, fmt.Errorf("创建PGP解密句柄失败: %w", err)
	}

	// 如果需要验证签名
	if verificationKeyID, exists := options["verification_key_id"]; exists {
		if publicKey, err := de.keyStore.GetPublicKey(ctx, verificationKeyID.(string)); err == nil {
			key, err := crypto.NewKeyFromArmored(publicKey)
			if err == nil {
				decHandle.VerificationKey(key)
			}
		}
	}

	// 解密内容
	decrypted, err := decHandle.Decrypt(encryptedDoc.Content, crypto.Armor)
	if err != nil {
		return nil, fmt.Errorf("PGP解密失败: %w", err)
	}

	// 检查签名
	if sigErr := decrypted.SignatureError(); sigErr != nil {
		de.logger.WithError(sigErr).Warn("PGP签名验证失败")
	}

	// 清理私有参数
	decHandle.ClearPrivateParams()

	return decrypted.Bytes(), nil
}

// encryptHybrid 混合加密
func (de *documentEncryptor) encryptHybrid(ctx context.Context, content []byte, options map[string]interface{}) (*EncryptedDocument, error) {
	// 生成随机AES密钥
	aesKey := make([]byte, de.config.AESKeySize/8)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, fmt.Errorf("生成AES密钥失败: %w", err)
	}

	// 使用AES加密内容
	aesOptions := map[string]interface{}{
		"key":         base64.StdEncoding.EncodeToString(aesKey),
		"document_id": options["document_id"],
	}

	aesEncrypted, err := de.encryptAES(ctx, content, aesOptions)
	if err != nil {
		return nil, fmt.Errorf("AES加密失败: %w", err)
	}

	// 使用PGP加密AES密钥
	keyOptions := map[string]interface{}{
		"recipients": options["recipients"],
	}

	keyEncrypted, err := de.encryptPGP(ctx, aesKey, keyOptions)
	if err != nil {
		return nil, fmt.Errorf("PGP加密密钥失败: %w", err)
	}

	// 组合结果
	docID := options["document_id"].(string)
	combinedContent := append(keyEncrypted.Content, aesEncrypted.Content...)

	return &EncryptedDocument{
		ID:         de.generateID(),
		DocumentID: docID,
		Method:     EncryptionMethodHybrid,
		Content:    combinedContent,
		Recipients: aesEncrypted.Recipients,
		Metadata: map[string]string{
			"algorithm":    "Hybrid",
			"aes_size":     fmt.Sprintf("%d", len(aesEncrypted.Content)),
			"pgp_size":     fmt.Sprintf("%d", len(keyEncrypted.Content)),
			"key_id":       keyEncrypted.ID,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// decryptHybrid 混合解密
func (de *documentEncryptor) decryptHybrid(ctx context.Context, encryptedDoc *EncryptedDocument, options map[string]interface{}) ([]byte, error) {
	// 对于混合加密，需要先分离PGP加密的密钥和AES加密的内容
	// 这里简化处理，假设前N字节是PGP加密的密钥，剩余是AES加密的内容

	// 在实际实现中，应该使用更复杂的分割逻辑
	// 这里只是示例实现
	return nil, fmt.Errorf("混合解密功能待实现")
}

// GenerateKeyPair 生成密钥对
func (de *documentEncryptor) GenerateKeyPair(ctx context.Context, userID, email string, keyType string) (*KeyPair, error) {
	var key crypto.Key
	var err error

	// 生成密钥
	keyGenHandle := de.pgp.KeyGeneration().AddUserId(userID, email).New()

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

	privateKey, err := key.ArmorPrivate()
	if err != nil {
		return nil, fmt.Errorf("获取私钥失败: %w", err)
	}

	// 获取指纹
	fingerprint := key.GetFingerprint()

	// 创建密钥对记录
	keyPair := &KeyPair{
		ID:          de.generateID(),
		UserID:      userID,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		KeyType:     keyType,
		KeySize:     de.config.AESKeySize,
		Fingerprint: fingerprint,
		Email:       email,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, de.config.KeyExpiration),
		IsActive:    true,
	}

	// 存储密钥对
	err = de.keyStore.StoreKeyPair(ctx, keyPair)
	if err != nil {
		return nil, fmt.Errorf("存储密钥对失败: %w", err)
	}

	de.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"email":      email,
		"key_type":   keyType,
		"fingerprint": fingerprint,
	}).Info("密钥对生成成功")

	return keyPair, nil
}

// ImportPublicKey 导入公钥
func (de *documentEncryptor) ImportPublicKey(ctx context.Context, userID string, publicKey string) error {
	// 验证公钥格式
	key, err := crypto.NewKeyFromArmored(publicKey)
	if err != nil {
		return fmt.Errorf("公钥格式无效: %w", err)
	}

	// 获取公钥信息
	// 注意：GopenPGP v3 API中没有直接的方法获取主要用户ID，需要从密钥中解析
	// 这里简化处理，使用默认的用户信息
	fingerprint := key.GetFingerprint()

	// 创建密钥对记录（仅包含公钥）
	keyPair := &KeyPair{
		ID:          de.generateID(),
		UserID:      userID,
		PublicKey:   publicKey,
		PrivateKey:  "", // 不存储私钥
		KeyType:     "imported",
		Fingerprint: fingerprint,
		Email:       "imported@unknown.com",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, de.config.KeyExpiration),
		IsActive:    true,
	}

	// 存储公钥
	err = de.keyStore.StoreKeyPair(ctx, keyPair)
	if err != nil {
		return fmt.Errorf("存储公钥失败: %w", err)
	}

	de.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"fingerprint": fingerprint,
	}).Info("公钥导入成功")

	return nil
}

// GetPublicKey 获取公钥
func (de *documentEncryptor) GetPublicKey(ctx context.Context, userID string) (string, error) {
	return de.keyStore.GetPublicKey(ctx, userID)
}

// BatchEncrypt 批量加密
func (de *documentEncryptor) BatchEncrypt(ctx context.Context, documents [][]byte, method EncryptionMethod, options map[string]interface{}) ([]*EncryptionResult, error) {
	results := make([]*EncryptionResult, len(documents))

	for i, doc := range documents {
		docOptions := make(map[string]interface{})
		for k, v := range options {
			docOptions[k] = v
		}
		docOptions["document_id"] = fmt.Sprintf("%s_%d", options["document_id"], i)

		result, err := de.Encrypt(ctx, doc, method, docOptions)
		if err != nil {
			results[i] = &EncryptionResult{
				Success: false,
				Error:   err.Error(),
			}
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// BatchDecrypt 批量解密
func (de *documentEncryptor) BatchDecrypt(ctx context.Context, encryptedDocs []*EncryptedDocument, options map[string]interface{}) ([]*DecryptionResult, error) {
	results := make([]*DecryptionResult, len(encryptedDocs))

	for i, doc := range encryptedDocs {
		result, err := de.Decrypt(ctx, doc, options)
		if err != nil {
			results[i] = &DecryptionResult{
				Success: false,
				Error:   err.Error(),
			}
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// 辅助方法

// generateID 生成唯一ID
func (de *documentEncryptor) generateID() string {
	return fmt.Sprintf("enc_%d_%x", time.Now().UnixNano(), rand.Int31())
}

// generateCacheKey 生成缓存键
func (de *documentEncryptor) generateCacheKey(documentID string, method EncryptionMethod) string {
	return fmt.Sprintf("enc:%s:%s", documentID, method)
}

// generateKeyFromPassword 从密码生成密钥
func (de *documentEncryptor) generateKeyFromPassword(password string, salt []byte) []byte {
	hash := sha256.New()
	for i := 0; i < de.config.IterationCount; i++ {
		hash.Write(salt)
		hash.Write([]byte(password))
	}
	return hash.Sum(nil)
}