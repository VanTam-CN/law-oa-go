package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/law-oa/services/document-service/internal/repositories"
)

// MockTokenRepository 模拟令牌仓库
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) Create(token *Token) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenRepository) FindByJTI(jti string) (*Token, error) {
	args := m.Called(jti)
	return args.Get(0).(*Token), args.Error(1)
}

func (m *MockTokenRepository) Update(token *Token) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockTokenRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTokenRepository) FindByUserID(userID uint) ([]*Token, error) {
	args := m.Called(userID)
	return args.Get(0).([]*Token), args.Error(1)
}

func (m *MockTokenRepository) FindBySessionID(sessionID string) ([]*Token, error) {
	args := m.Called(sessionID)
	return args.Get(0).([]*Token), args.Error(1)
}

func (m *MockTokenRepository) FindExpired() ([]*Token, error) {
	args := m.Called()
	return args.Get(0).([]*Token), args.Error(1)
}

func (m *MockTokenRepository) CleanupExpired() error {
	args := m.Called()
	return args.Error(0)
}

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(id uint) (*User, error) {
	args := m.Called(id)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(username string) (*User, error) {
	args := m.Called(username)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) Create(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockAuditRepository 模拟审计仓库
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(audit *Audit) error {
	args := m.Called(audit)
	return args.Error(0)
}

func (m *MockAuditRepository) FindByUserID(userID uint, limit int) ([]*Audit, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]*Audit), args.Error(1)
}

func (m *MockAuditRepository) FindBySessionID(sessionID string, limit int) ([]*Audit, error) {
	args := m.Called(sessionID, limit)
	return args.Get(0).([]*Audit), args.Error(1)
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to test database:", err)
	}

	// 自动迁移测试表
	err = db.AutoMigrate(&Token{}, &User{}, &Audit{})
	if err != nil {
		t.Fatal("Failed to migrate test database:", err)
	}

	return db
}

// setupTestJWTConfig 设置测试JWT配置
func setupTestJWTConfig() *JWTConfig {
	return &JWTConfig{
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 24 * time.Hour,
		Issuer:              "test-service",
		Audience:            "test-client",
		Leeway:              10 * time.Second,
		MaxTokenAge:         24 * time.Hour,
		AccessTokenKeyID:     "test-access-key",
		RefreshTokenKeyID:    "test-refresh-key",
	}
}

// setupTestJWTManager 设置测试JWT管理器
func setupTestJWTManager(t *testing.T) *JWTManager {
	db := setupTestDB(t)
	config := setupTestJWTConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tokenRepo := repositories.NewTokenRepository(db)
	userRepo := repositories.NewUserRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	manager, err := NewJWTManager(config, tokenRepo, userRepo, auditRepo, logger)
	if err != nil {
		t.Fatal("Failed to create JWT manager:", err)
	}

	return manager
}

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	manager := setupTestJWTManager(t)

	userID := uint(1)
	username := "testuser"
	tenantID := "tenant-001"
	roles := []string{"user"}
	permissions := []string{"document:read", "document:write"}
	ipAddress := "127.0.0.1"
	userAgent := "test-agent"

	tokenPair, err := manager.GenerateTokenPair(userID, username, tenantID, roles, permissions, ipAddress, userAgent)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Greater(t, tokenPair.ExpiresIn, int64(0))

	// 验证访问令牌
	accessClaims, err := manager.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, username, accessClaims.Username)
	assert.Equal(t, tenantID, accessClaims.TenantID)
	assert.Equal(t, "access", accessClaims.TokenType)

	// 验证刷新令牌
	refreshClaims, err := manager.ValidateToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, username, refreshClaims.Username)
	assert.Equal(t, tenantID, refreshClaims.TenantID)
	assert.Equal(t, "refresh", refreshClaims.TokenType)
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := setupTestJWTManager(t)

	// 生成令牌
	tokenPair, err := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")
	assert.NoError(t, err)

	// 验证访问令牌
	claims, err := manager.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "tenant-001", claims.TenantID)

	// 验证刷新令牌
	refreshClaims, err := manager.ValidateToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, "refresh", refreshClaims.TokenType)
}

func TestJWTManager_RefreshToken(t *testing.T) {
	manager := setupTestJWTManager(t)

	// 生成初始令牌对
	tokenPair, err := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")
	assert.NoError(t, err)

	// 刷新令牌
	newTokenPair, err := manager.RefreshToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newTokenPair.AccessToken)
	assert.NotEmpty(t, newTokenPair.RefreshToken)
	assert.NotEqual(t, tokenPair.AccessToken, newTokenPair.AccessToken)
	assert.NotEqual(t, tokenPair.RefreshToken, newTokenPair.RefreshToken)
}

func TestJWTManager_RevokeToken(t *testing.T) {
	manager := setupTestJWTManager(t)

	// 生成令牌
	tokenPair, err := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")
	assert.NoError(t, err)

	// 验证令牌有效
	_, err = manager.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)

	// 撤销令牌
	accessClaims, _ := manager.ValidateToken(tokenPair.AccessToken)
	err = manager.RevokeToken(accessClaims.JTI)
	assert.NoError(t, err)

	// 验证令牌被撤销
	assert.True(t, manager.IsBlacklisted(accessClaims.JTI))
}

func TestJWTMiddleware_RequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := setupTestJWTManager(t)
	config := setupTestJWTConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validator := NewTokenValidator(config, logger)
	options := DefaultMiddlewareOptions()
	middleware := NewJWTMiddleware(manager, validator, config, logger, options)

	// 创建路由器
	router := gin.New()
	router.Use(middleware.Middleware())
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	// 测试没有令牌的请求
	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// 测试有效令牌的请求
	tokenPair, _ := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTMiddleware_RequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := setupTestJWTManager(t)
	config := setupTestJWTConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validator := NewTokenValidator(config, logger)
	options := DefaultMiddlewareOptions()
	middleware := NewJWTMiddleware(manager, validator, config, logger, options)

	// 创建路由器
	router := gin.New()
	router.Use(middleware.Middleware())

	protected := router.Group("/protected")
	protected.Use(middleware.RequirePermission("document", "read"))
	{
		protected.GET("/documents", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
		})
	}

	// 生成具有文档读取权限的令牌
	tokenPair, _ := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")
	req, _ := http.NewRequest("GET", "/protected/documents", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 生成没有文档读取权限的令牌
	tokenPairNoPermission, _ := manager.GenerateTokenPair(uint(2), "testuser2", "tenant-001", []string{"user"}, []string{"other:read"}, "127.0.0.1", "test-agent")
	req, _ = http.NewRequest("GET", "/protected/documents", nil)
	req.Header.Set("Authorization", "Bearer "+tokenPairNoPermission.AccessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestJWTMiddleware_RequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := setupTestJWTManager(t)
	config := setupTestJWTConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validator := NewTokenValidator(config, logger)
	options := DefaultMiddlewareOptions()
	middleware := NewJWTMiddleware(manager, validator, config, logger, options)

	// 创建路由器
	router := gin.New()
	router.Use(middleware.Middleware())

	admin := router.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/users", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Admin access granted"})
		})
	}

	// 测试管理员角色
	adminTokenPair, _ := manager.GenerateTokenPair(uint(1), "admin", "tenant-001", []string{"admin"}, []string{"user:read"}, "127.0.0.1", "test-agent")
	req, _ := http.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminTokenPair.AccessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 测试普通用户角色
	userTokenPair, _ := manager.GenerateTokenPair(uint(2), "user", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")
	req, _ = http.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+userTokenPair.AccessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTokenValidator_ValidateClaims(t *testing.T) {
	config := setupTestJWTConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validator := NewTokenValidator(config, logger)

	// 测试有效声明
	validClaims := &TokenClaims{
		UserID:    1,
		Username:  "testuser",
		TenantID:  "tenant-001",
		DeviceID:  "device-001",
		SessionID: "session-001",
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
		TokenType: "access",
		Roles:     []string{"user"},
		Permissions: []string{"document:read"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Issuer,
			Audience:  []string{config.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        "test-jti-001",
		},
	}

	errors := validator.ValidateClaims(validClaims)
	assert.Empty(t, errors)

	// 测试无效发行者
	invalidIssuerClaims := *validClaims
	invalidIssuerClaims.Issuer = "invalid-issuer"
	errors = validator.ValidateClaims(&invalidIssuerClaims)
	assert.NotEmpty(t, errors)

	// 测试过期令牌
	expiredClaims := *validClaims
	expiredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	errors = validator.ValidateClaims(&expiredClaims)
	assert.NotEmpty(t, errors)
}

func TestErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Token Missing",
			error:          ErrTokenMissing,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "TOKEN_MISSING",
		},
		{
			name:           "Token Expired",
			error:          ErrTokenExpired,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "TOKEN_EXPIRED",
		},
		{
			name:           "Permission Denied",
			error:          ErrPermissionDenied,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "PERMISSION_DENIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", func(c *gin.Context) {
				DefaultErrorHandler(c, tt.error)
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			var response map[string]interface{}
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			assert.Equal(t, tt.expectedCode, response["code"])
		})
	}
}

func BenchmarkJWTManager_GenerateTokenPair(b *testing.B) {
	manager := setupTestJWTManager(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTManager_ValidateToken(b *testing.B) {
	manager := setupTestJWTManager(&testing.T{})
	tokenPair, _ := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ValidateToken(tokenPair.AccessToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTMiddleware_Middleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	manager := setupTestJWTManager(&testing.T{})
	config := setupTestJWTConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	validator := NewTokenValidator(config, logger)
	options := DefaultMiddlewareOptions()
	middleware := NewJWTMiddleware(manager, validator, config, logger, options)

	router := gin.New()
	router.Use(middleware.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	tokenPair, _ := manager.GenerateTokenPair(uint(1), "testuser", "tenant-001", []string{"user"}, []string{"document:read"}, "127.0.0.1", "test-agent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}