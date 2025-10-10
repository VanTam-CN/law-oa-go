package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/models"
)

var (
	encryptionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "encryption_duration_seconds",
		Help:    "Duration of encryption operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})

	encryptionErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "encryption_errors_total",
		Help: "Total number of encryption errors",
	}, []string{"operation", "type"})

	encryptionOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "encryption_operations_total",
		Help: "Total number of encryption operations",
	}, []string{"operation"})
)

// EncryptionConfig 加密配置
type EncryptionServiceConfig struct {
	AESKey                string
	RSAPrivateKey         string
	RSAPublicKey          string
	DataKeyRotationDays   int
	EnableFieldEncryption bool
	SensitiveFields       []string
}

// EncryptionService 加密服务
type EncryptionService struct {
	config        *EncryptionConfig
	aesKey        []byte
	rsaPrivateKey *rsa.PrivateKey
	rsaPublicKey  *rsa.PublicKey
	cacheService  *cache.CacheService
	db            *gorm.DB
}

// NewEncryptionService 创建加密服务
func NewEncryptionService(config *EncryptionConfig, cacheService *cache.CacheService, db *gorm.DB) (*EncryptionService, error) {
	service := &EncryptionService{
		config:       config,
		cacheService: cacheService,
		db:           db,
	}

	// 初始化AES密钥
	if err := service.initAESKey(); err != nil {
		return nil, fmt.Errorf("failed to initialize AES key: %w", err)
	}

	// 初始化RSA密钥对
	if err := service.initRSAKeys(); err != nil {
		return nil, fmt.Errorf("failed to initialize RSA keys: %w", err)
	}

	return service, nil
}

// initAESKey 初始化AES密钥
func (s *EncryptionService) initAESKey() error {
	if s.config.AESKey == "" {
		// 生成新的AES密钥
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("failed to generate AES key: %w", err)
		}
		s.aesKey = key
	} else {
		// 使用配置的AES密钥
		decodedKey, err := base64.StdEncoding.DecodeString(s.config.AESKey)
		if err != nil {
			return fmt.Errorf("failed to decode AES key: %w", err)
		}
		if len(decodedKey) != 32 {
			return errors.New("AES key must be 32 bytes")
		}
		s.aesKey = decodedKey
	}

	return nil
}

// initRSAKeys 初始化RSA密钥对
func (s *EncryptionService) initRSAKeys() error {
	if s.config.RSAPrivateKey == "" || s.config.RSAPublicKey == "" {
		// 生成新的RSA密钥对
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate RSA key pair: %w", err)
		}
		s.rsaPrivateKey = privateKey
		s.rsaPublicKey = &privateKey.PublicKey
	} else {
		// 使用配置的RSA密钥对
		privateKey, err := s.parseRSAPrivateKey(s.config.RSAPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse RSA private key: %w", err)
		}
		s.rsaPrivateKey = privateKey

		publicKey, err := s.parseRSAPublicKey(s.config.RSAPublicKey)
		if err != nil {
			return fmt.Errorf("failed to parse RSA public key: %w", err)
		}
		s.rsaPublicKey = publicKey
	}

	return nil
}

// parseRSAPrivateKey 解析RSA私钥
func (s *EncryptionService) parseRSAPrivateKey(key string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return privateKey, nil
}

// parseRSAPublicKey 解析RSA公钥
func (s *EncryptionService) parseRSAPublicKey(key string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// EncryptAES AES加密
func (s *EncryptionService) EncryptAES(plaintext string) (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("aes_encrypt").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("aes_encrypt").Inc()
	}()

	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		encryptionErrors.WithLabelValues("aes_encrypt", "cipher_creation").Inc()
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		encryptionErrors.WithLabelValues("aes_encrypt", "gcm_creation").Inc()
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		encryptionErrors.WithLabelValues("aes_encrypt", "nonce_generation").Inc()
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES AES解密
func (s *EncryptionService) DecryptAES(ciphertext string) (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("aes_decrypt").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("aes_decrypt").Inc()
	}()

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		encryptionErrors.WithLabelValues("aes_decrypt", "base64_decode").Inc()
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		encryptionErrors.WithLabelValues("aes_decrypt", "cipher_creation").Inc()
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		encryptionErrors.WithLabelValues("aes_decrypt", "gcm_creation").Inc()
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		encryptionErrors.WithLabelValues("aes_decrypt", "invalid_data").Inc()
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		encryptionErrors.WithLabelValues("aes_decrypt", "decryption").Inc()
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptRSA RSA加密
func (s *EncryptionService) EncryptRSA(plaintext string) (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("rsa_encrypt").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("rsa_encrypt").Inc()
	}()

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, s.rsaPublicKey, []byte(plaintext), nil)
	if err != nil {
		encryptionErrors.WithLabelValues("rsa_encrypt", "encryption").Inc()
		return "", fmt.Errorf("failed to encrypt with RSA: %w", err)
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptRSA RSA解密
func (s *EncryptionService) DecryptRSA(ciphertext string) (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("rsa_decrypt").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("rsa_decrypt").Inc()
	}()

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		encryptionErrors.WithLabelValues("rsa_decrypt", "base64_decode").Inc()
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, s.rsaPrivateKey, data, nil)
	if err != nil {
		encryptionErrors.WithLabelValues("rsa_decrypt", "decryption").Inc()
		return "", fmt.Errorf("failed to decrypt with RSA: %w", err)
	}

	return string(plaintext), nil
}

// HashPassword 密码哈希
func (s *EncryptionService) HashPassword(password string) (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("password_hash").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("password_hash").Inc()
	}()

	// 使用scrypt进行密钥派生
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		encryptionErrors.WithLabelValues("password_hash", "salt_generation").Inc()
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 64)
	if err != nil {
		encryptionErrors.WithLabelValues("password_hash", "scrypt").Inc()
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// 使用bcrypt进行二次哈希
	bcryptHash, err := bcrypt.GenerateFromPassword(hash, 12)
	if err != nil {
		encryptionErrors.WithLabelValues("password_hash", "bcrypt").Inc()
		return "", fmt.Errorf("failed to bcrypt hash: %w", err)
	}

	// 返回格式: salt:bcrypt_hash
	return fmt.Sprintf("%x:%s", salt, bcryptHash), nil
}

// VerifyPassword 验证密码
func (s *EncryptionService) VerifyPassword(password, hash string) (bool, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("password_verify").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("password_verify").Inc()
	}()

	parts := strings.Split(hash, ":")
	if len(parts) != 2 {
		encryptionErrors.WithLabelValues("password_verify", "invalid_format").Inc()
		return false, errors.New("invalid password hash format")
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		encryptionErrors.WithLabelValues("password_verify", "salt_decode").Inc()
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	bcryptHash := parts[1]

	// 使用scrypt派生密钥
	scryptHash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 64)
	if err != nil {
		encryptionErrors.WithLabelValues("password_verify", "scrypt").Inc()
		return false, fmt.Errorf("failed to scrypt hash: %w", err)
	}

	// 验证bcrypt哈希
	err = bcrypt.CompareHashAndPassword([]byte(bcryptHash), scryptHash)
	if err != nil {
		encryptionErrors.WithLabelValues("password_verify", "bcrypt_compare").Inc()
		return false, nil
	}

	return true, nil
}

// EncryptField 加密敏感字段
func (s *EncryptionService) EncryptField(value string, fieldType string) (string, error) {
	if !s.config.EnableFieldEncryption || !s.isSensitiveField(fieldType) {
		return value, nil
	}

	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("field_encrypt").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("field_encrypt").Inc()
	}()

	return s.EncryptAES(value)
}

// DecryptField 解密敏感字段
func (s *EncryptionService) DecryptField(value string, fieldType string) (string, error) {
	if !s.config.EnableFieldEncryption || !s.isSensitiveField(fieldType) {
		return value, nil
	}

	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("field_decrypt").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("field_decrypt").Inc()
	}()

	return s.DecryptAES(value)
}

// isSensitiveField 检查是否为敏感字段
func (s *EncryptionService) isSensitiveField(fieldType string) bool {
	for _, field := range s.config.SensitiveFields {
		if fieldType == field {
			return true
		}
	}
	return false
}

// GenerateDataKey 生成数据密钥
func (s *EncryptionService) GenerateDataKey() (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("data_key_generate").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("data_key_generate").Inc()
	}()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		encryptionErrors.WithLabelValues("data_key_generate", "key_generation").Inc()
		return "", fmt.Errorf("failed to generate data key: %w", err)
	}

	return s.EncryptRSA(string(key))
}

// ComputeHash 计算哈希值
func (s *EncryptionService) ComputeHash(data string, algorithm string) (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("hash_compute").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("hash_compute").Inc()
	}()

	var hasher hash.Hash

	switch algorithm {
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		encryptionErrors.WithLabelValues("hash_compute", "invalid_algorithm").Inc()
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// EncryptUserData 加密用户数据
func (s *EncryptionService) EncryptUserData(userData *models.User) error {
	if !s.config.EnableFieldEncryption {
		return nil
	}

	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("user_data_encrypt").Observe(time.Since(start).Seconds())
	}()

	// 加密敏感字段
	if userData.Email != "" {
		encrypted, err := s.EncryptField(userData.Email, "email")
		if err != nil {
			return fmt.Errorf("failed to encrypt email: %w", err)
		}
		userData.Email = encrypted
	}

	if userData.Phone != "" {
		encrypted, err := s.EncryptField(userData.Phone, "phone")
		if err != nil {
			return fmt.Errorf("failed to encrypt phone: %w", err)
		}
		userData.Phone = encrypted
	}

	// 加密自定义字段
	if userData.Phone != "" {
		encrypted, err := s.EncryptField(userData.Phone, "remark")
		if err != nil {
			return fmt.Errorf("failed to encrypt remark: %w", err)
		}
		userData.Phone = encrypted
	}

	return nil
}

// DecryptUserData 解密用户数据
func (s *EncryptionService) DecryptUserData(userData *models.User) error {
	if !s.config.EnableFieldEncryption {
		return nil
	}

	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("user_data_decrypt").Observe(time.Since(start).Seconds())
	}()

	// 解密敏感字段
	if userData.Email != "" {
		decrypted, err := s.DecryptField(userData.Email, "email")
		if err != nil {
			return fmt.Errorf("failed to decrypt email: %w", err)
		}
		userData.Email = decrypted
	}

	if userData.Phone != "" {
		decrypted, err := s.DecryptField(userData.Phone, "phone")
		if err != nil {
			return fmt.Errorf("failed to decrypt phone: %w", err)
		}
		userData.Phone = decrypted
	}

	// 解密自定义字段
	if userData.Phone != "" {
		decrypted, err := s.DecryptField(userData.Phone, "remark")
		if err != nil {
			return fmt.Errorf("failed to decrypt remark: %w", err)
		}
		userData.Phone = decrypted
	}

	return nil
}

// SanitizeInput 清理输入数据
func (s *EncryptionService) SanitizeInput(input string) string {
	// 移除潜在的恶意字符
	sanitized := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(input, "")

	// 防止SQL注入模式
	sqlPatterns := []string{
		`(?i)(union\s+select|insert\s+into|delete\s+from|update\s+set|drop\s+table)`,
		`(?i)(or\s+1\s*=\s*1|or\s+1\s*=\s*'1'|or\s+1\s*=\s*"1")`,
		`(?i)(;|--|\/\*|\*\/|@@|@)`,
	}

	for _, pattern := range sqlPatterns {
		sanitized = regexp.MustCompile(pattern).ReplaceAllString(sanitized, "")
	}

	return strings.TrimSpace(sanitized)
}

// GenerateSecureToken 生成安全令牌
func (s *EncryptionService) GenerateSecureToken(length int) (string, error) {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("token_generate").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("token_generate").Inc()
	}()

	if length < 16 {
		length = 16
	}

	token := make([]byte, length)
	if _, err := rand.Read(token); err != nil {
		encryptionErrors.WithLabelValues("token_generate", "token_generation").Inc()
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return hex.EncodeToString(token), nil
}

// RotateKey 旋转密钥
func (s *EncryptionService) RotateKey() error {
	start := time.Now()
	defer func() {
		encryptionDuration.WithLabelValues("key_rotate").Observe(time.Since(start).Seconds())
		encryptionOperations.WithLabelValues("key_rotate").Inc()
	}()

	// 生成新的AES密钥
	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		encryptionErrors.WithLabelValues("key_rotate", "key_generation").Inc()
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	s.aesKey = newKey

	// 清除相关缓存
	if s.cacheService != nil {
		s.cacheService.ClearPattern("encryption:*")
	}

	// 记录密钥旋转事件
	// TODO: 记录到安全审计日志

	return nil
}

// GetKeyFingerprint 获取密钥指纹
func (s *EncryptionService) GetKeyFingerprint() string {
	hash := sha256.Sum256(s.aesKey)
	return hex.EncodeToString(hash[:])
}

// GetEncryptionStatus 获取加密状态
func (s *EncryptionService) GetEncryptionStatus() map[string]interface{} {
	return map[string]interface{}{
		"field_encryption_enabled": s.config.EnableFieldEncryption,
		"key_rotation_days":        s.config.DataKeyRotationDays,
		"sensitive_fields_count":   len(s.config.SensitiveFields),
		"aes_key_fingerprint":      s.GetKeyFingerprint(),
		"rsa_key_size":             s.rsaPublicKey.Size(),
	}
}
