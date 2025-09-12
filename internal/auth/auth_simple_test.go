package auth

import (
	"context"
	"testing"
	"time"

	"law-oa-go/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTokenManager_NewTokenManager 测试TokenManager创建
func TestTokenManager_NewTokenManager(t *testing.T) {
	t.Run("创建TokenManager", func(t *testing.T) {
		// 创建测试配置
		cfg := createTestConfig()

		// 创建TokenManager
		tokenManager := NewTokenManager(cfg, createTestRedisClient(), createTestCacheService())

		// 验证TokenManager
		assert.NotNil(t, tokenManager)
		assert.Equal(t, []byte(cfg.JWT.Secret), tokenManager.secret)
		assert.Equal(t, time.Duration(cfg.JWT.ExpiresIn)*time.Second, tokenManager.accessTTL)
		assert.Equal(t, time.Duration(cfg.JWT.RefreshIn)*time.Second, tokenManager.refreshTTL)
		assert.Equal(t, "law-oa-system", tokenManager.issuer)
	})
}

// TestTokenManager_CreateTokens 测试令牌创建
func TestTokenManager_CreateTokens(t *testing.T) {
	t.Run("创建访问令牌和刷新令牌", func(t *testing.T) {
		// 准备测试环境
		tokenManager := NewTokenManager(createTestConfig(), createTestRedisClient(), createTestCacheService())

		// 创建测试用户
		user := &models.User{
			ID:    1,
			Name:  "测试用户",
			Email: "test@example.com",
			Role:  "user",
		}

		// 创建令牌
		tokenDetails, err := tokenManager.CreateTokens(
			context.Background(),
			user,
			"device-123",
			"192.168.1.1",
			"Mozilla/5.0",
		)

		// 验证结果
		require.NoError(t, err)
		assert.NotNil(t, tokenDetails)
		assert.NotEmpty(t, tokenDetails.AccessToken)
		assert.NotEmpty(t, tokenDetails.RefreshToken)
		assert.NotEmpty(t, tokenDetails.AccessUUID)
		assert.NotEmpty(t, tokenDetails.RefreshUUID)
		assert.Greater(t, tokenDetails.AtExpires, time.Now().Unix())
		assert.Greater(t, tokenDetails.RtExpires, time.Now().Unix())
	})
}

// TestTokenManager_VerifyToken 测试令牌验证
func TestTokenManager_VerifyToken(t *testing.T) {
	t.Run("验证有效令牌", func(t *testing.T) {
		// 准备测试环境
		tokenManager := NewTokenManager(createTestConfig(), createTestRedisClient(), createTestCacheService())

		// 创建测试用户和令牌
		user := &models.User{
			ID:    1,
			Name:  "测试用户",
			Email: "test@example.com",
			Role:  "user",
		}

		tokenDetails, err := tokenManager.CreateTokens(
			context.Background(),
			user,
			"device-123",
			"192.168.1.1",
			"Mozilla/5.0",
		)
		require.NoError(t, err)

		// 验证访问令牌
		accessClaims, err := tokenManager.VerifyToken(context.Background(), tokenDetails.AccessToken)
		require.NoError(t, err)
		assert.NotNil(t, accessClaims)

		// 访问payload中的用户信息
		payload, ok := (*accessClaims)["payload"].(map[string]interface{})
		require.True(t, ok, "Payload should be a map")
		assert.Equal(t, float64(user.ID), payload["user_id"])
		assert.Equal(t, user.Name, payload["username"]) // 注意：User模型用的是Name字段
		assert.Equal(t, user.Email, payload["email"])
		assert.Equal(t, user.Role, payload["role"])
	})

	t.Run("验证无效令牌", func(t *testing.T) {
		// 准备测试环境
		tokenManager := NewTokenManager(createTestConfig(), createTestRedisClient(), createTestCacheService())

		// 验证无效令牌
		accessClaims, err := tokenManager.VerifyToken(context.Background(), "invalid.token.here")
		assert.Error(t, err)
		assert.Nil(t, accessClaims)
	})
}
