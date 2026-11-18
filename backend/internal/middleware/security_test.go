//go:build ignore
// +build ignore

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
)

func TestSecurityMiddleware_Init(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   0,
		},
	}

	// 测试初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, GetSecurityMiddleware())
}

func TestSecurityMiddleware_RateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   1, // 使用不同的数据库避免测试冲突
		},
	}

	// 初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)

	// 获取安全中间件并设置测试限制
	sm := GetSecurityMiddleware()
	sm.rateLimiter.SetEndpointLimit("/test", security.RateLimitRule{
		Window:      time.Minute,
		MaxRequests: 2, // 设置更小的限制用于测试
		Burst:       1,
	})

	// 创建测试路由
	router := gin.New()
	router.Use(NewRateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 测试正常请求
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.300:12345" // 使用不同的IP避免Redis缓存影响
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// 测试第二个请求
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.300:12345"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// 测试第三个请求（应该被限制）
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.300:12345"
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestSecurityMiddleware_JWTAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   2, // 使用不同的数据库避免测试冲突
		},
	}

	// 初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)

	// 创建测试用户
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}

	// 生成令牌
	tokenDetails, err := NewGenerateToken(context.Background(), user, "device123", "127.0.0.1", "test-agent")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenDetails.AccessToken)

	// 创建测试路由
	router := gin.New()
	router.Use(NewJWTMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		c.JSON(200, gin.H{"user_id": userID})
	})

	// 测试无令牌请求
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 401, w1.Code)

	// 测试有效令牌请求
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenDetails.AccessToken)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

func TestSecurityMiddleware_CombinedMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   3, // 使用不同的数据库避免测试冲突
		},
	}

	// 初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)

	// 创建测试用户
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}

	// 生成令牌
	tokenDetails, err := NewGenerateToken(context.Background(), user, "device123", "127.0.0.1", "test-agent")
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(CombinedMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(200, gin.H{"user_id": userID})
		} else {
			c.JSON(200, gin.H{"message": "public"})
		}
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 测试公共路径（应该跳过认证）
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// 测试无令牌请求（应该被拒绝）
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 401, w2.Code)

	// 测试有效令牌请求
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.Header.Set("Authorization", "Bearer "+tokenDetails.AccessToken)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
}

func TestSecurityMiddleware_TokenRefresh(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   4, // 使用不同的数据库避免测试冲突
		},
	}

	// 初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)

	// 创建测试用户
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}

	// 生成令牌
	tokenDetails, err := NewGenerateToken(context.Background(), user, "device123", "127.0.0.1", "test-agent")
	assert.NoError(t, err)

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)

	// 刷新令牌
	newTokenDetails, err := RefreshToken(context.Background(), tokenDetails.RefreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newTokenDetails.AccessToken)
	assert.NotEqual(t, tokenDetails.AccessToken, newTokenDetails.AccessToken)
}

func TestSecurityMiddleware_IPManagement(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   5, // 使用不同的数据库避免测试冲突
		},
	}

	// 初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)

	sm := GetSecurityMiddleware()

	// 添加IP到白名单
	sm.AddIPToWhitelist("192.168.1.200")

	// 添加IP到黑名单
	sm.AddIPToBlacklist("192.168.1.300", time.Hour)

	// 测试获取限制信息
	info, err := sm.GetRateLimitInfo(context.Background(), "192.168.1.100", "/test", "")
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// 获取安全统计信息
	stats := sm.GetSecurityStats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "jwt_key_manager")
	assert.Contains(t, stats, "rate_limiter")

	// 清理
	sm.RemoveIPFromWhitelist("192.168.1.200")
	sm.RemoveIPFromBlacklist("192.168.1.300")
}

func TestSecurityMiddleware_LegacyCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   6, // 使用不同的数据库避免测试冲突
		},
	}

	// 测试兼容性初始化
	LegacyInitJWT(cfg)
	assert.NotNil(t, GetSecurityMiddleware())

	// 测试兼容性令牌生成
	token, expiresAt, err := LegacyGenerateToken(1, "testuser", "user")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, expiresAt.After(time.Now()))

	// 测试兼容性令牌验证
	claims, err := LegacyValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)

	// 创建测试路由
	router := gin.New()
	router.Use(LegacyAuthMiddleware())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		c.JSON(200, gin.H{"user_id": userID})
	})

	// 测试兼容性中间件
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// 从响应中解析用户ID来验证中间件正确设置了用户信息
	var response map[string]interface{}
	jsonErr := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, jsonErr)
	userID, ok := response["user_id"]
	assert.True(t, ok)
	assert.Equal(t, float64(1), userID) // JSON数字解析为float64

	// 测试兼容性获取用户信息（需要通过中间件处理的context）
	// 创建一个新的请求来测试兼容函数
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/compat-test", nil)
	req2.Header.Set("Authorization", "Bearer "+token)

	// 创建一个处理器来测试兼容函数
	router.GET("/compat-test", func(c *gin.Context) {
		userID, exists := LegacyGetCurrentUserID(c)
		assert.True(t, exists)
		assert.Equal(t, uint(1), userID)

		username, exists := LegacyGetCurrentUsername(c)
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)

		role, exists := LegacyGetCurrentRole(c)
		assert.True(t, exists)
		assert.Equal(t, "user", role)

		c.JSON(200, gin.H{"success": true})
	})

	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
}

func TestSecurityMiddleware_LegacyRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   7, // 使用不同的数据库避免测试冲突
		},
	}

	// 初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)

	// 创建测试路由
	router := gin.New()
	router.Use(LegacyIPRateLimitMiddleware(2, time.Minute, nil))
	router.GET("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 测试正常请求
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/test", nil)
	req1.RemoteAddr = "192.168.1.200:12345" // 使用不同的IP避免Redis缓存影响
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// 测试第二个请求
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/test", nil)
	req2.RemoteAddr = "192.168.1.200:12345"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// 测试第三个请求（应该被限制）
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/api/test", nil)
	req3.RemoteAddr = "192.168.1.200:12345"
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestSecurityMiddleware_ConcurrentAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-32-bytes-long-for-testing",
			ExpiresIn: 3600,
			RefreshIn: 7200,
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DB:   8, // 使用不同的数据库避免测试冲突
		},
	}

	// 初始化安全中间件
	err := InitSecurity(cfg)
	assert.NoError(t, err)

	// 创建测试用户
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}

	// 并发生成令牌
	type result struct {
		token string
		err   error
	}
	results := make(chan result, 10)

	for i := 0; i < 10; i++ {
		go func() {
			tokenDetails, err := NewGenerateToken(context.Background(), user, "device123", "127.0.0.1", "test-agent")
			if err != nil {
				results <- result{err: err}
			} else {
				results <- result{token: tokenDetails.AccessToken}
			}
		}()
	}

	// 收集结果
	for i := 0; i < 10; i++ {
		result := <-results
		assert.NoError(t, result.err)
		assert.NotEmpty(t, result.token)
	}
}
