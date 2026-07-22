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

// SubjectDataKey loads the dedicated key used for sensitive subject data.
// It accepts a raw 32-byte value, base64, or hexadecimal encoding so deployment
// systems can store it without exposing the plaintext identity values.
func SubjectDataKey() ([]byte, error) {
	raw := strings.TrimSpace(os.Getenv(subjectDataKeyEnv))
	if raw == "" {
		return nil, errors.New("SUBJECT_DATA_KEY 未配置")
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
	return nil, errors.New("SUBJECT_DATA_KEY 必须解码为32字节")
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
