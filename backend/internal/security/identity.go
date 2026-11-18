package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const subjectDataKeyEnv = "SUBJECT_DATA_KEY"

var (
	ErrSubjectDataKeyMissing = errors.New("SUBJECT_DATA_KEY 未配置")
	ErrSubjectDataKeyInvalid = errors.New("SUBJECT_DATA_KEY 必须解码为32字节")
)

// SubjectDataKey loads the dedicated key used for sensitive subject data.
// It accepts a raw 32-byte value, base64, or hexadecimal encoding so deployment
// systems can store it without exposing the plaintext identity values.
func SubjectDataKey() ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv(subjectDataKeyEnv))
	if raw == "" {
		return nil, ErrSubjectDataKeyMissing
	}
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := hex.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	return nil, ErrSubjectDataKeyInvalid
}

// IdentityDigest returns a keyed digest for exact lookup. Plain SHA-256 is
// intentionally avoided because common identity numbers are guessable.
func IdentityDigest(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	key, err := SubjectDataKey()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	_, _ = io.WriteString(mac, value)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// NormalizeIdentityNumber produces the canonical lookup value shared by
// intake, entity registration and duplicate detection. It does not validate
// document authenticity; that remains a professional review responsibility.
func NormalizeIdentityNumber(identityType, value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToUpper(strings.TrimSpace(identityType)) {
	case "ID_CARD", "BUSINESS_LICENSE", "ORGANIZATION_CODE", "SOCIAL_CREDIT_CODE":
		value = strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", "-", "").Replace(value)
		return strings.ToUpper(value)
	case "PASSPORT":
		return strings.ToUpper(strings.ReplaceAll(value, " ", ""))
	default:
		return value
	}
}

// ProtectIdentityNumber encrypts a subject identity and returns ciphertext and
// its keyed lookup digest. The plaintext is never returned for persistence.
func ProtectIdentityNumber(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", nil
	}
	key, err := SubjectDataKey()
	if err != nil {
		return "", "", err
	}
	digest, err := IdentityDigest(value)
	if err != nil {
		return "", "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("初始化身份信息加密失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("初始化身份信息密封失败: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", fmt.Errorf("生成身份信息随机数失败: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(value), nil)
	return base64.RawStdEncoding.EncodeToString(sealed), digest, nil
}

// ProtectSensitiveValue encrypts non-searchable personal data with a
// purpose-derived key. Purpose separation prevents a ciphertext copied from
// one field family from being accepted as another field family.
func ProtectSensitiveValue(purpose, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	key, err := purposeKey(purpose)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("初始化敏感信息加密失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("初始化敏感信息密封失败: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成敏感信息随机数失败: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(value), []byte(strings.TrimSpace(purpose)))
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

// DecryptSensitiveValue decrypts a purpose-bound value for an already
// authorized server-side response. Callers must never log the plaintext.
func DecryptSensitiveValue(purpose, encoded string) (string, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return "", nil
	}
	key, err := purposeKey(purpose)
	if err != nil {
		return "", err
	}
	sealed, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("敏感信息密文格式无效")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("初始化敏感信息解密失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("初始化敏感信息解封失败: %w", err)
	}
	if len(sealed) < gcm.NonceSize() {
		return "", errors.New("敏感信息密文长度无效")
	}
	nonce, ciphertext := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(strings.TrimSpace(purpose)))
	if err != nil {
		return "", errors.New("敏感信息解密失败")
	}
	return string(plaintext), nil
}

func purposeKey(purpose string) ([]byte, error) {
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return nil, errors.New("敏感信息加密用途不能为空")
	}
	masterKey, err := SubjectDataKey()
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, masterKey)
	_, _ = io.WriteString(mac, "law-oa-go:sensitive-value:"+purpose)
	return mac.Sum(nil), nil
}

// DecryptIdentityNumber is only for trusted server-side workflows that need to
// construct an in-memory conflict query. Callers must not serialize its result.
func DecryptIdentityNumber(encoded string) (string, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return "", nil
	}
	key, err := SubjectDataKey()
	if err != nil {
		return "", err
	}
	sealed, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("身份信息密文格式无效")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("初始化身份信息解密失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("初始化身份信息解封失败: %w", err)
	}
	if len(sealed) < gcm.NonceSize() {
		return "", errors.New("身份信息密文长度无效")
	}
	nonce, ciphertext := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("身份信息解密失败")
	}
	return string(plaintext), nil
}

func IdentityPresent(raw, ciphertext, digest string) bool {
	return strings.TrimSpace(raw) != "" || strings.TrimSpace(ciphertext) != "" || strings.TrimSpace(digest) != ""
}

func MaskIdentityNumber(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return ""
	}
	if len(runes) <= 4 {
		return "****"
	}
	keepPrefix := 3
	if len(runes) < 8 {
		keepPrefix = 1
	}
	return string(runes[:keepPrefix]) + strings.Repeat("*", len(runes)-keepPrefix-4) + string(runes[len(runes)-4:])
}
