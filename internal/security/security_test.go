package security

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
)

func TestNewJWTKeyManager(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-32-bytes-long",
		},
	}

	securityConfig := &SecurityConfig{
		JWT: JWTConfig{
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
			Issuer:          "test-issuer",
			Secret:          "test-secret-key-32-bytes-long",
			EnableRefresh:   true,
			BlacklistTTL:    24 * time.Hour,
		},
	}

	// 创建Redis客户端和缓存服务
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cacheService := cache.NewCacheService(redisClient, "test")

	// 创建JWT密钥管理器
	manager := NewJWTKeyManager(cfg, securityConfig, redisClient, cacheService)

	assert.NotNil(t, manager)
	assert.NotEmpty(t, manager.currentSecret)
	assert.NotEmpty(t, manager.keyHistory)
	assert.Equal(t, 1, len(manager.keyHistory))
}

func TestJWTKeyManager_CreateTokens(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-32-bytes-long",
		},
	}

	securityConfig := &SecurityConfig{
		JWT: JWTConfig{
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
			Issuer:          "test-issuer",
			Secret:          "test-secret-key-32-bytes-long",
			EnableRefresh:   true,
			BlacklistTTL:    24 * time.Hour,
		},
	}

	// 创建Redis客户端和缓存服务
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cacheService := cache.NewCacheService(redisClient, "test")

	// 创建JWT密钥管理器
	manager := NewJWTKeyManager(cfg, securityConfig, redisClient, cacheService)

	// 创建测试用户
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}

	// 创建令牌
	tokens, err := manager.CreateTokens(context.Background(), user, "device123", "127.0.0.1", "test-agent")
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.NotEmpty(t, tokens.AccessUUID)
	assert.NotEmpty(t, tokens.RefreshUUID)
}

func TestJWTKeyManager_VerifyToken(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-32-bytes-long",
		},
	}

	securityConfig := &SecurityConfig{
		JWT: JWTConfig{
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
			Issuer:          "test-issuer",
			Secret:          "test-secret-key-32-bytes-long",
			EnableRefresh:   true,
			BlacklistTTL:    24 * time.Hour,
		},
	}

	// 创建Redis客户端和缓存服务
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cacheService := cache.NewCacheService(redisClient, "test")

	// 创建JWT密钥管理器
	manager := NewJWTKeyManager(cfg, securityConfig, redisClient, cacheService)

	// 创建测试用户
	user := &models.User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}

	// 创建令牌
	tokens, err := manager.CreateTokens(context.Background(), user, "device123", "127.0.0.1", "test-agent")
	assert.NoError(t, err)

	// 验证令牌
	claims, err := manager.VerifyToken(context.Background(), tokens.AccessToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
}

func TestJWTKeyManager_RotateKeys(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key-32-bytes-long",
		},
	}

	securityConfig := &SecurityConfig{
		JWT: JWTConfig{
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
			Issuer:          "test-issuer",
			Secret:          "test-secret-key-32-bytes-long",
			EnableRefresh:   true,
			BlacklistTTL:    time.Millisecond, // 设置很短的TTL用于测试
		},
	}

	// 创建Redis客户端和缓存服务
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cacheService := cache.NewCacheService(redisClient, "test")

	// 创建JWT密钥管理器
	manager := NewJWTKeyManager(cfg, securityConfig, redisClient, cacheService)

	// 记录原始密钥
	originalSecret := string(manager.currentSecret)

	// 等待超过旋转周期
	time.Sleep(2 * time.Millisecond)

	// 旋转密钥
	err := manager.RotateKeys(context.Background())
	assert.NoError(t, err)

	// 验证密钥已改变
	assert.NotEqual(t, originalSecret, string(manager.currentSecret))
	assert.Equal(t, 2, len(manager.keyHistory))
}

func TestNewRateLimiter(t *testing.T) {
	// 创建安全配置
	securityConfig := &SecurityConfig{
		APISecurity: APISecurityConfig{
			EnableRateLimit:      true,
			RateLimitWindow:      time.Minute,
			RateLimitMaxRequests: 100,
			WhitelistedIPs:       []string{"192.168.1.1"},
			BlacklistedIPs:       []string{"192.168.1.2"},
		},
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建请求限制器
	limiter := NewRateLimiter(redisClient, &securityConfig.APISecurity)

	assert.NotNil(t, limiter)
	assert.Equal(t, 1, len(limiter.ipWhitelist))
	assert.Contains(t, limiter.ipWhitelist, "192.168.1.1")
	assert.NotEmpty(t, limiter.endpointLimits)
	assert.Contains(t, limiter.endpointLimits, "/api/auth/login")
}

func TestRateLimiter_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建安全配置
	securityConfig := &SecurityConfig{
		APISecurity: APISecurityConfig{
			EnableRateLimit:      true,
			RateLimitWindow:      time.Minute,
			RateLimitMaxRequests: 2, // 设置很低的限制用于测试
		},
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建请求限制器
	limiter := NewRateLimiter(redisClient, &securityConfig.APISecurity)

	// 创建测试路由
	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 测试正常请求
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// 测试第二个请求
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// 测试第三个请求（应该被限制）
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestRateLimiter_IPWhitelist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建安全配置
	securityConfig := &SecurityConfig{
		APISecurity: APISecurityConfig{
			EnableRateLimit:      true,
			RateLimitWindow:      time.Minute,
			RateLimitMaxRequests: 1, // 设置很低的限制
			WhitelistedIPs:       []string{"192.168.1.200"},
		},
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建请求限制器
	limiter := NewRateLimiter(redisClient, &securityConfig.APISecurity)

	// 创建测试路由
	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 测试白名单IP的多个请求
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.200:12345"
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}
}

func TestRateLimiter_IPBlacklist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建安全配置
	securityConfig := &SecurityConfig{
		APISecurity: APISecurityConfig{
			EnableRateLimit:      true,
			EnableIPBlacklist:    true,
			RateLimitWindow:      time.Minute,
			RateLimitMaxRequests: 100,
		},
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建请求限制器
	limiter := NewRateLimiter(redisClient, &securityConfig.APISecurity)

	// 添加IP到黑名单
	limiter.AddToBlacklist("192.168.1.300", time.Hour)

	// 创建测试路由
	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 测试黑名单IP的请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.300:12345"
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestRateLimiter_GetRateLimitInfo(t *testing.T) {
	// 创建安全配置
	securityConfig := &SecurityConfig{
		APISecurity: APISecurityConfig{
			EnableRateLimit:      true,
			RateLimitWindow:      time.Minute,
			RateLimitMaxRequests: 10,
		},
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建请求限制器
	limiter := NewRateLimiter(redisClient, &securityConfig.APISecurity)

	// 获取限制信息
	info, err := limiter.GetRateLimitInfo(context.Background(), "192.168.1.100", "/test", "")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, 10, info.MaxRequests)
	assert.False(t, info.IsBlocked)
}

func TestRateLimiter_GetStats(t *testing.T) {
	// 创建安全配置
	securityConfig := &SecurityConfig{
		APISecurity: APISecurityConfig{
			EnableRateLimit:      true,
			RateLimitWindow:      time.Minute,
			RateLimitMaxRequests: 100,
			WhitelistedIPs:       []string{"192.168.1.1"},
		},
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建请求限制器
	limiter := NewRateLimiter(redisClient, &securityConfig.APISecurity)

	// 获取统计信息
	stats := limiter.GetStats()
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats["ip_whitelist_size"])
	assert.Equal(t, true, stats["enable_rate_limit"])
}

func TestRateLimiter_ClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建安全配置
	securityConfig := &SecurityConfig{
		APISecurity: APISecurityConfig{
			EnableRateLimit:      true,
			RateLimitWindow:      time.Minute,
			RateLimitMaxRequests: 100,
		},
	}

	// 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 创建请求限制器
	limiter := NewRateLimiter(redisClient, &securityConfig.APISecurity)

	// 创建测试请求
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Real-IP", "10.0.0.1")
	req.Header.Set("X-Forwarded-For", "172.16.0.1, 192.168.1.50")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	// 测试IP提取
	ip := limiter.getClientIP(c)
	assert.Equal(t, "172.16.0.1", ip) // 应该获取X-Forwarded-For中的第一个IP
}
