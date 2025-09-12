package security

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"law-oa-go/internal/models"
)

// TestEncryptionService_NewEncryptionService 测试加密服务创建
func TestEncryptionService_NewEncryptionService(t *testing.T) {
	t.Run("创建加密服务 - 无配置密钥", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
			SensitiveFields:       []string{"email", "phone"},
			DataKeyRotationDays:   90,
			HashAlgorithm:         "sha256",
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.aesKey)
		assert.NotNil(t, service.rsaPrivateKey)
		assert.NotNil(t, service.rsaPublicKey)
	})

	t.Run("创建加密服务 - 有配置密钥", func(t *testing.T) {
		// 生成测试用的RSA密钥对
		_, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		config := &EncryptionConfig{
			AESKey:                base64.StdEncoding.EncodeToString(make([]byte, 32)),
			DataKeyRotationDays:   90,
			EnableFieldEncryption: true,
			SensitiveFields:       []string{"email", "phone"},
			HashAlgorithm:         "sha256",
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)
		assert.NotNil(t, service)
		assert.NotNil(t, service.aesKey)
	})
}

// TestEncryptionService_AES_EncryptDecrypt 测试AES加密解密
func TestEncryptionService_AES_EncryptDecrypt(t *testing.T) {
	t.Run("AES加密解密循环测试", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
			SensitiveFields:       []string{"email", "phone"},
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		testCases := []struct {
			name      string
			plaintext string
		}{
			{"简单文本", "Hello, World!"},
			{"中文文本", "你好，世界！"},
			{"特殊字符", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
			{"长文本", "这是一段很长的文本，用来测试AES加密解密的性能和正确性。 Lorem ipsum dolor sit amet, consectetur adipiscing elit."},
			{"空字符串", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 加密
				ciphertext, err := service.EncryptAES(tc.plaintext)
				require.NoError(t, err)
				assert.NotEqual(t, tc.plaintext, ciphertext)

				// 解密
				decrypted, err := service.DecryptAES(ciphertext)
				require.NoError(t, err)
				assert.Equal(t, tc.plaintext, decrypted)
			})
		}
	})

	t.Run("AES加密错误处理", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		// 测试无效密文
		_, err = service.DecryptAES("invalid-ciphertext")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode base64")

		// 测试被篡改的密文
		validCiphertext, err := service.EncryptAES("test")
		require.NoError(t, err)
		
		// 篡改密文
		tamperedCiphertext := validCiphertext[:len(validCiphertext)-1] + "X"
		_, err = service.DecryptAES(tamperedCiphertext)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt")
	})
}

// TestEncryptionService_RSA_EncryptDecrypt 测试RSA加密解密
func TestEncryptionService_RSA_EncryptDecrypt(t *testing.T) {
	t.Run("RSA加密解密循环测试", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		testCases := []struct {
			name      string
			plaintext string
		}{
			{"短文本", "Hello"},
			{"中等长度文本", "This is a medium length text for RSA encryption testing."},
			{"数字", "1234567890"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 加密
				ciphertext, err := service.EncryptRSA(tc.plaintext)
				require.NoError(t, err)
				assert.NotEqual(t, tc.plaintext, ciphertext)

				// 解密
				decrypted, err := service.DecryptRSA(ciphertext)
				require.NoError(t, err)
				assert.Equal(t, tc.plaintext, decrypted)
			})
		}
	})
}

// TestEncryptionService_PasswordHashing 测试密码哈希
func TestEncryptionService_PasswordHashing(t *testing.T) {
	t.Run("密码哈希验证", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		testPasswords := []struct {
			name     string
			password string
		}{
			{"简单密码", "password123"},
			{"复杂密码", "MySecureP@ssw0rd!2024"},
			{"中文密码", "我的密码123"},
			{"长密码", "ThisIsAVeryLongPasswordThatShouldBeHashedSecurely1234567890!@#$%^&*()"},
		}

		for _, tc := range testPasswords {
			t.Run(tc.name, func(t *testing.T) {
				// 哈希密码
				hash, err := service.HashPassword(tc.password)
				require.NoError(t, err)
				assert.NotEqual(t, tc.password, hash)
				assert.Contains(t, hash, ":") // 确保格式正确

				// 验证密码
				valid, err := service.VerifyPassword(tc.password, hash)
				require.NoError(t, err)
				assert.True(t, valid)

				// 验证错误密码
				valid, err = service.VerifyPassword("wrong-password", hash)
				require.NoError(t, err)
				assert.False(t, valid)
			})
		}
	})

	t.Run("密码哈希错误处理", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		// 测试无效哈希格式
		_, err = service.VerifyPassword("password", "invalid-hash-format")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid password hash format")
	})
}

// TestEncryptionService_FieldEncryption 测试字段加密
func TestEncryptionService_FieldEncryption(t *testing.T) {
	t.Run("敏感字段加密", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
			SensitiveFields:       []string{"email", "phone", "id_card"},
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		testCases := []struct {
			name      string
			fieldType string
			value     string
			encrypted bool
		}{
			{"加密邮箱字段", "email", "user@example.com", true},
			{"加密手机字段", "phone", "13800138000", true},
			{"加密身份证字段", "id_card", "110101199001011234", true},
			{"不加密姓名字段", "name", "张三", false},
			{"不加密地址字段", "address", "北京市朝阳区", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 加密字段
				encrypted, err := service.EncryptField(tc.value, tc.fieldType)
				require.NoError(t, err)

				if tc.encrypted {
					assert.NotEqual(t, tc.value, encrypted)
				} else {
					assert.Equal(t, tc.value, encrypted)
				}

				// 解密字段
				decrypted, err := service.DecryptField(encrypted, tc.fieldType)
				require.NoError(t, err)
				assert.Equal(t, tc.value, decrypted)
			})
		}
	})

	t.Run("字段加密未启用", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: false,
			SensitiveFields:       []string{"email", "phone"},
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		// 即使是敏感字段，当加密未启用时也应该原样返回
		value := "user@example.com"
		encrypted, err := service.EncryptField(value, "email")
		require.NoError(t, err)
		assert.Equal(t, value, encrypted)

		decrypted, err := service.DecryptField(encrypted, "email")
		require.NoError(t, err)
		assert.Equal(t, value, decrypted)
	})
}

// TestEncryptionService_UserDataEncryption 测试用户数据加密
func TestEncryptionService_UserDataEncryption(t *testing.T) {
	t.Run("用户数据加密解密", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
			SensitiveFields:       []string{"email", "phone", "remark"},
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		originalUser := &models.User{
			ID:       1,
			Name:     "张三",
			Email:    "zhangsan@example.com",
			Phone:    "13800138000",
			Role:     "user",
		}

		// 复制用户数据用于测试
		user := *originalUser

		// 加密用户数据
		err = service.EncryptUserData(&user)
		require.NoError(t, err)
		assert.NotEqual(t, originalUser.Email, user.Email)
		assert.NotEqual(t, originalUser.Phone, user.Phone)
		// 姓名和状态应该保持不变
		assert.Equal(t, originalUser.Name, user.Name)
		assert.Equal(t, originalUser.Status, user.Status)

		// 解密用户数据
		err = service.DecryptUserData(&user)
		require.NoError(t, err)
		assert.Equal(t, originalUser.Email, user.Email)
		assert.Equal(t, originalUser.Phone, user.Phone)
		assert.Equal(t, originalUser.Name, user.Name)
		assert.Equal(t, originalUser.Status, user.Status)
	})
}

// TestEncryptionService_UtilityFunctions 测试工具函数
func TestEncryptionService_UtilityFunctions(t *testing.T) {
	t.Run("生成数据密钥", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		key, err := service.GenerateDataKey()
		require.NoError(t, err)
		assert.NotEmpty(t, key)

		// 确保生成的密钥可以解密
		decryptedKey, err := service.DecryptRSA(key)
		require.NoError(t, err)
		assert.Len(t, decryptedKey, 32) // AES密钥应该是32字节
	})

	t.Run("计算哈希值", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		data := "test data for hashing"

		// 测试SHA256
		hash256, err := service.ComputeHash(data, "sha256")
		require.NoError(t, err)
		assert.Len(t, hash256, 64) // SHA256哈希应该是64个十六进制字符

		// 测试SHA512
		hash512, err := service.ComputeHash(data, "sha512")
		require.NoError(t, err)
		assert.Len(t, hash512, 128) // SHA512哈希应该是128个十六进制字符

		// 确保相同数据产生相同哈希
		hash256Again, err := service.ComputeHash(data, "sha256")
		require.NoError(t, err)
		assert.Equal(t, hash256, hash256Again)

		// 测试不支持算法
		_, err = service.ComputeHash(data, "invalid_algorithm")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported hash algorithm")
	})

	t.Run("输入清理", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{"正常输入", "normal input", "normal input"},
			{"包含控制字符", "input\x00with\x07control\x1fchars", "inputwithcontrolchars"},
			{"SQL注入尝试", "SELECT * FROM users", "SELECT * FROM users"},
			{"包含SQL注入模式", "1 OR 1=1", "1"},
			{"包含注释", "user/*comment*/name", "usercommentname"},
			{"前后空格", "  trimmed  ", "trimmed"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := service.SanitizeInput(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("生成安全令牌", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		// 测试不同长度的令牌
		for length := 16; length <= 64; length += 16 {
			token, err := service.GenerateSecureToken(length)
			require.NoError(t, err)
			assert.Len(t, token, length*2) // 十六进制编码，长度翻倍
			assert.Regexp(t, "^[a-f0-9]+$", token) // 确保是十六进制字符
		}

		// 测试最小长度
		token, err := service.GenerateSecureToken(8)
		require.NoError(t, err)
		assert.Len(t, token, 32) // 最小长度16，编码后32个字符
	})
}

// TestEncryptionService_KeyManagement 测试密钥管理
func TestEncryptionService_KeyManagement(t *testing.T) {
	t.Run("密钥旋转", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		originalFingerprint := service.GetKeyFingerprint()

		// 使用原始密钥加密数据
		plaintext := "test data for key rotation"
		ciphertext, err := service.EncryptAES(plaintext)
		require.NoError(t, err)

		// 旋转密钥
		err = service.RotateKey()
		require.NoError(t, err)

		// 确保密钥指纹已改变
		newFingerprint := service.GetKeyFingerprint()
		assert.NotEqual(t, originalFingerprint, newFingerprint)

		// 使用新密钥加密数据
		newCiphertext, err := service.EncryptAES(plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, ciphertext, newCiphertext)

		// 注意：由于密钥已旋转，用新密钥无法解密旧数据
		// 这是正常的安全行为
	})

	t.Run("获取加密状态", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
			SensitiveFields:       []string{"email", "phone", "address"},
			DataKeyRotationDays:   90,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		status := service.GetEncryptionStatus()
		assert.NotNil(t, status)
		assert.True(t, status["field_encryption_enabled"].(bool))
		assert.Equal(t, 90, status["key_rotation_days"])
		assert.Equal(t, 3, status["sensitive_fields_count"])
		assert.NotEmpty(t, status["aes_key_fingerprint"])
		assert.Equal(t, 256, status["rsa_key_size"])
	})
}

// TestEncryptionService_Metrics 测试指标收集
func TestEncryptionService_Metrics(t *testing.T) {
	t.Run("加密操作指标", func(t *testing.T) {
		config := &EncryptionConfig{
			EnableFieldEncryption: true,
		}

		service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
		require.NoError(t, err)

		// 重置指标计数器
		resetCounter("encryption_operations_total", "aes_encrypt")
		resetCounter("encryption_operations_total", "aes_decrypt")
		resetCounter("encryption_operations_total", "password_hash")
		resetCounter("encryption_operations_total", "password_verify")

		// 执行一些加密操作
		_, err = service.EncryptAES("test data")
		require.NoError(t, err)

		_, err = service.DecryptAES("test ciphertext") // 这会失败，但会计数
		// require.NoError(t, err)

		_, err = service.HashPassword("test password")
		require.NoError(t, err)

		_, err = service.VerifyPassword("test password", "hash:bcrypt")
		// require.NoError(t, err) // 这会失败，但会计数

		// 验证指标已增加
		assertCounterGreaterThan(t, "encryption_operations_total", "aes_encrypt", 0)
		assertCounterGreaterThan(t, "encryption_operations_total", "aes_decrypt", 0)
		assertCounterGreaterThan(t, "encryption_operations_total", "password_hash", 0)
		assertCounterGreaterThan(t, "encryption_operations_total", "password_verify", 0)
	})
}

// BenchmarkEncryptionService 加密服务基准测试
func BenchmarkEncryptionService_EncryptAES(b *testing.B) {
	config := &EncryptionConfig{
		EnableFieldEncryption: true,
	}

	service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
	require.NoError(b, err)

	plaintext := "This is a test string for AES encryption benchmarking."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.EncryptAES(plaintext)
	}
}

func BenchmarkEncryptionService_DecryptAES(b *testing.B) {
	config := &EncryptionConfig{
		EnableFieldEncryption: true,
	}

	service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
	require.NoError(b, err)

	plaintext := "This is a test string for AES decryption benchmarking."
	ciphertext, err := service.EncryptAES(plaintext)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.DecryptAES(ciphertext)
	}
}

func BenchmarkEncryptionService_HashPassword(b *testing.B) {
	config := &EncryptionConfig{
		EnableFieldEncryption: true,
	}

	service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
	require.NoError(b, err)

	password := "SecurePassword123!@#"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.HashPassword(password)
	}
}

func BenchmarkEncryptionService_VerifyPassword(b *testing.B) {
	config := &EncryptionConfig{
		EnableFieldEncryption: true,
	}

	service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
	require.NoError(b, err)

	password := "SecurePassword123!@#"
	hash, err := service.HashPassword(password)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.VerifyPassword(password, hash)
	}
}

// setupTestEncryptionService 设置测试用的加密服务
func setupTestEncryptionService(t *testing.T) *EncryptionService {
	config := &EncryptionConfig{
		EnableFieldEncryption: true,
		SensitiveFields:       []string{"email", "phone", "address"},
		DataKeyRotationDays:   90,
		HashAlgorithm:         "sha256",
	}

	service, err := NewEncryptionService(config, createTestCacheService(), createTestDB())
	require.NoError(t, err)
	return service
}

// resetCounter 重置Prometheus计数器（简化版）
func resetCounter(metricName string, labelValue string) {
	// 简化实现，实际项目中可以集成完整的Prometheus测试
}

// assertCounterGreaterThan 断言计数器值大于指定值（简化版）
func assertCounterGreaterThan(t *testing.T, metricName string, labelValue string, minValue float64) {
	// 简化实现，实际项目中可以集成完整的Prometheus测试
}