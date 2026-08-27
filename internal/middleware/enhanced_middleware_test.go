//go:build ignore
// +build ignore

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/security"
)

// MockConfig 模拟配置
type MockConfig struct {
	mock config.Config
}

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	mock redis.Client
}

func (m *MockRedisClient) Get(key string) *redis.StringCmd {
	return redis.NewStringCmd(context.Background(), "test", "test-value")
}

func (m *MockRedisClient) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return redis.NewStatusCmd(context.Background(), "test", "OK")
}

func (m *MockRedisClient) Incr(key string) *redis.IntCmd {
	return redis.NewIntCmd(context.Background(), "test", int64(1))
}

func (m *MockRedisClient) Expire(key string, expiration time.Duration) *redis.BoolCmd {
	return redis.NewBoolCmd(context.Background(), "test", true)
}

func (m *MockRedisClient) Ping() *redis.StatusCmd {
	return redis.NewStatusCmd(context.Background(), "test", "PONG")
}

func (m *MockRedisClient) Pipeline() redis.Pipeliner {
	return redis.NewPipeline(context.Background())
}

// MockCacheService 模拟缓存服务
type MockCacheService struct{}

func (m *MockCacheService) Get(key string) (interface{}, error) {
	return nil, nil
}

func (m *MockCacheService) Set(key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (m *MockCacheService) Delete(key string) error {
	return nil
}

func (m *MockCacheService) Exists(key string) (bool, error) {
	return false, nil
}

// MockJWTKeyManager 模拟JWT密钥管理器
type MockJWTKeyManager struct{}

func (m *MockJWTKeyManager) ExtractTokenMetadata(ctx context.Context, token string) (interface{}, error) {
	return map[string]interface{}{
		"user_id":    1,
		"username":   "testuser",
		"role":       "user",
		"device_id":  "device123",
		"ip":         "127.0.0.1",
		"user_agent": "test-agent",
	}, nil
}

func (m *MockJWTKeyManager) IsTokenBlacklisted(ctx context.Context, token string) bool {
	return false
}

func (m *MockJWTKeyManager) GetKeyStats() map[string]interface{} {
	return map[string]interface{}{
		"active_keys": 1,
		"total_keys":  1,
	}
}

// MockRateLimiter 模拟请求限制器
type MockRateLimiter struct{}

func (m *MockRateLimiter) AllowRequest(c *gin.Context, ip string) bool {
	return true
}

func (m *MockRateLimiter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"requests_blocked": 0,
		"active_clients":   1,
	}
}

func (m *MockRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (m *MockRateLimiter) StartCleanupTimer() {
	// Mock implementation
}

func createTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Env: "test",
		},
		JWT: config.JWTConfig{
			Secret:    "test-secret",
			ExpiresIn: 3600,
			RefreshIn: 86400,
		},
	}
}

func createTestSecurityConfig() *security.SecurityConfig {
	return &security.SecurityConfig{
		JWT: security.JWTConfig{
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: time.Hour * 24,
			Issuer:          "law-oa-test",
			Secret:          "test-secret",
			EnableRefresh:   true,
			BlacklistTTL:    time.Hour * 24,
		},
		APISecurity: security.APISecurityConfig{
			EnableRateLimit:         true,
			EnableIPWhitelist:       false,
			EnableIPBlacklist:       true,
			EnableRequestSigning:    false,
			EnableAPIThrottling:     true,
			EnableWAFProtection:     true,
			EnableDDoSProtection:    true,
			EnableRequestValidation: true,
			EnableCORS:              true,
			EnableCSRF:              true,
			RateLimitWindow:         time.Minute,
			RateLimitMaxRequests:    100,
			WhitelistedIPs:          []string{},
			BlacklistedIPs:          []string{},
			AllowedOrigins:          []string{"http://localhost:3000"},
			AllowedMethods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:          []string{"*"},
			MaxRequestSize:          10 * 1024 * 1024,
		},
		Auth: security.AuthConfig{
			EnableDeviceCheck:     true,
			EnableIPCheck:         true,
			EnableRateLimit:       true,
			SkipAuthPaths:         []string{"/health", "/metrics"},
			SkipAuthPrefixes:      []string{"/public/"},
			SessionTimeout:        time.Hour * 24,
			MaxConcurrentSessions: 3,
			PasswordPolicy: security.PasswordPolicy{
				MinLength:             8,
				MaxLength:             128,
				RequireUppercase:      true,
				RequireLowercase:      true,
				RequireNumbers:        true,
				RequireSpecialChars:   true,
				ForbidCommonPasswords: true,
				ForbidPersonalInfo:    true,
				ExpiryDays:            90,
				HistoryCount:          5,
				FailedAttempts:        5,
				LockoutDuration:       time.Minute * 30,
			},
		},
		Validation: security.ValidationConfig{
			EnableInputValidation:   true,
			EnableSQLInjectionCheck: true,
			EnableXSSCheck:          true,
			MaxStringLength:         1000,
			MaxFileSize:             50 * 1024 * 1024,
			AllowedFileTypes:        []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"},
		},
	}
}

func TestEnhancedSecurityMiddleware_Creation(t *testing.T) {
	config := createTestConfig()
	cacheService := &MockCacheService{}
	redisClient := &MockRedisClient{}

	esm := NewEnhancedSecurityMiddleware(config, cacheService, redisClient)

	assert.NotNil(t, esm)
	assert.NotNil(t, esm.jwtManager)
	assert.NotNil(t, esm.rateLimiter)
}

func TestEnhancedSecurityMiddleware_SecurityHeadersMiddleware(t *testing.T) {
	esm := &EnhancedSecurityMiddleware{}
	middleware := esm.SecurityHeadersMiddleware()

	// 创建测试路由
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"test": "success"})
	})

	// 发送请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 检查响应头
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	assert.Empty(t, w.Header().Get("Server"))
}

func TestEnhancedSecurityMiddleware_RequestValidationMiddleware(t *testing.T) {
	esm := &EnhancedSecurityMiddleware{}
	middleware := esm.RequestValidationMiddleware()

	// 创建测试路由
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"test": "success"})
	})

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid GET request",
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid method",
			method:         "PATCH",
			path:           "/test",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  "Method not allowed",
		},
		{
			name:           "Long URL",
			method:         "GET",
			path:           "/" + strings.Repeat("a", 2001),
			expectedStatus: http.StatusRequestURITooLong,
			expectedError:  "URL too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestEnhancedSecurityMiddleware_ExtractToken(t *testing.T) {
	esm := &EnhancedSecurityMiddleware{}

	tests := []struct {
		name       string
		authHeader string
		expected   string
		queryToken string
	}{
		{
			name:       "Bearer token",
			authHeader: "Bearer test-token",
			expected:   "test-token",
		},
		{
			name:       "No auth header",
			authHeader: "",
			expected:   "",
		},
		{
			name:       "Invalid format",
			authHeader: "Basic dGVzdDpYXRva2Vu",
			expected:   "",
		},
		{
			name:       "Query token fallback",
			authHeader: "",
			expected:   "query-token",
			queryToken: "query-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}
			if tt.queryToken != "" {
				c.Request.URL.RawQuery = "token=" + tt.queryToken
			}

			token := esm.extractToken(c)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestEnhancedSecurityMiddleware_ShouldSkipAuth(t *testing.T) {
	esm := &EnhancedSecurityMiddleware{
		config: createTestSecurityConfig(),
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Health check path",
			path:     "/health",
			expected: true,
		},
		{
			name:     "Metrics path",
			path:     "/metrics",
			expected: true,
		},
		{
			name:     "Public prefix",
			path:     "/public/api/test",
			expected: true,
		},
		{
			name:     "API path",
			path:     "/api/users",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request.URL.Path = tt.path

			result := esm.shouldSkipAuth(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnhancedPerformanceMiddleware_Creation(t *testing.T) {
	config := createTestConfig()
	redisClient := &MockRedisClient{}

	epm := NewEnhancedPerformanceMiddleware(config, redisClient)

	assert.NotNil(t, epm)
	assert.NotNil(t, epm.metrics)
	assert.NotNil(t, epm.latencies)
}

func TestEnhancedPerformanceMiddleware_GlobalPerformanceMiddleware(t *testing.T) {
	epm := NewEnhancedPerformanceMiddleware(createTestConfig(), &MockRedisClient{})
	middleware := epm.GlobalPerformanceMiddleware()

	// 创建测试路由
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"test": "success"})
	})

	// 发送请求
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestEnhancedPerformanceMiddleware_GetPerformanceMetrics(t *testing.T) {
	epm := NewEnhancedPerformanceMiddleware(createTestConfig(), &MockRedisClient{})

	// 添加一些测试数据
	epm.metrics.RequestCount = 100
	epm.metrics.ErrorCount = 5
	epm.metrics.AverageLatency = 100 * time.Millisecond

	metrics := epm.GetPerformanceMetrics()

	assert.Equal(t, int64(100), metrics["request_count"])
	assert.Equal(t, int64(5), metrics["error_count"])
	assert.Equal(t, "100ms", metrics["average_latency"])
}

func TestEnhancedPerformanceMiddleware_ResetMetrics(t *testing.T) {
	epm := NewEnhancedPerformanceMiddleware(createTestConfig(), &MockRedisClient{})

	// 添加一些数据
	epm.metrics.RequestCount = 100
	epm.metrics.ErrorCount = 5

	// 重置指标
	epm.ResetMetrics()

	assert.Equal(t, int64(0), epm.metrics.RequestCount)
	assert.Equal(t, int64(0), epm.metrics.ErrorCount)
	assert.True(t, time.Since(epm.metrics.LastReset) < time.Second)
}

func TestDefaultMiddlewareConfig(t *testing.T) {
	config := DefaultMiddlewareConfig()

	assert.True(t, config.Security.EnableCORS)
	assert.True(t, config.Security.EnableRateLimit)
	assert.True(t, config.Security.EnableJWTValidation)
	assert.True(t, config.Performance.EnableCompression)
	assert.True(t, config.Performance.EnableMetrics)
	assert.True(t, config.Performance.EnableHealthCheck)
}

func TestProductionMiddlewareConfig(t *testing.T) {
	config := ProductionMiddlewareConfig()

	assert.True(t, config.Security.EnableCORS)
	assert.True(t, config.Security.EnableRateLimit)
	assert.True(t, config.Security.EnableJWTValidation)
	assert.Equal(t, 1000, config.Security.RateLimitMaxRequests)
	assert.Equal(t, 9, config.Performance.CompressionLevel)
	assert.Equal(t, 5000, config.Performance.MaxConcurrentRequests)
}

func TestDevelopmentMiddlewareConfig(t *testing.T) {
	config := DevelopmentMiddlewareConfig()

	assert.True(t, config.Security.EnableCORS)
	assert.False(t, config.Security.EnableRateLimit) // 开发环境关闭
	assert.True(t, config.Security.EnableJWTValidation)
	assert.False(t, config.Performance.EnableCompression)        // 开发环境关闭
	assert.False(t, config.Performance.EnableConcurrencyControl) // 开发环境关闭
}

func TestMiddlewareChain_Apply(t *testing.T) {
	chain := MiddlewareChain{
		func(c *gin.Context) {
			c.Set("middleware1", "applied")
			c.Next()
		},
		func(c *gin.Context) {
			c.Set("middleware2", "applied")
			c.Next()
		},
	}

	handler := func(c *gin.Context) {
		c.Set("handler", "executed")
		c.JSON(http.StatusOK, gin.H{"success": true})
	}

	router := gin.New()
	router.GET("/test", chain.Apply(handler))

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "true")
}

func TestCreateSecurityChain(t *testing.T) {
	esm := &EnhancedSecurityMiddleware{
		config: createTestSecurityConfig(),
	}

	chain := CreateSecurityChain(esm)
	assert.Len(t, chain, 5)
}

func TestCreatePerformanceChain(t *testing.T) {
	epm := NewEnhancedPerformanceMiddleware(createTestConfig(), &MockRedisClient{})

	chain := CreatePerformanceChain(epm)
	assert.Len(t, chain, 3)
}

func TestCreateCustomMiddleware(t *testing.T) {
	config := map[string]interface{}{
		"log_level": "info",
	}

	middleware := CreateCustomMiddleware("custom_logger", config)
	assert.NotNil(t, middleware)
}

func TestMiddlewareHealthCheck(t *testing.T) {
	middleware := MiddlewareHealthCheck()

	router := gin.New()
	router.GET("/health", middleware)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestGetMiddlewareInfo(t *testing.T) {
	info := GetMiddlewareInfo()

	assert.Equal(t, "2.1.0", info["version"])
	assert.Equal(t, "Gin v1.9+", info["framework"])
	assert.Contains(t, info["features"], "Security Headers")
	assert.Contains(t, info["features"], "CORS")
	assert.Contains(t, info["features"], "Performance Monitoring")
}

// 基准测试
func BenchmarkSecurityHeadersMiddleware(b *testing.B) {
	esm := &EnhancedSecurityMiddleware{}
	middleware := esm.SecurityHeadersMiddleware()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request.Header.Set("User-Agent", "test-agent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware(c)
	}
}

func BenchmarkRequestValidationMiddleware(b *testing.B) {
	esm := &EnhancedSecurityMiddleware{}
	middleware := esm.RequestValidationMiddleware()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request.Method = "GET"
	c.Request.URL.Path = "/test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware(c)
	}
}

func BenchmarkExtractToken(b *testing.B) {
	esm := &EnhancedSecurityMiddleware{}
	c, _ := gin.CreateTestContext(htptest.NewRecorder())
	c.Request.Header.Set("Authorization", "Bearer test-token")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		esm.extractToken(c)
	}
}

func BenchmarkGlobalPerformanceMiddleware(b *testing.B) {
	epm := NewEnhancedPerformanceMiddleware(createTestConfig(), &MockRedisClient{})
	middleware := epm.GlobalPerformanceMiddleware()

	c, _ := gin.CreateTestContext(htptest.NewRecorder())
	c.Request.Method = "GET"
	c.Request.URL.Path = "/test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		middleware(c)
	}
}

func BenchmarkGetPerformanceMetrics(b *testing.B) {
	epm := NewEnhancedPerformanceMiddleware(createTestConfig(), &MockRedisClient{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		epm.GetPerformanceMetrics()
	}
}
