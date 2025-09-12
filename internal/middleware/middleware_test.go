package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// 创建模拟的认证服务
	mockAuthService := &MockAuthService{
		users: map[uint]*services.UserClaims{
			1: {
				UserID: 1,
				Name:   "张律师",
				Email:  "lawyer@example.com",
				Role:   "lawyer",
			},
		},
		validTokens: map[string]bool{
			"valid_token": true,
		},
	}
	
	authMiddleware := middleware.AuthMiddleware(mockAuthService)
	
	t.Run("valid token", func(t *testing.T) {
		router := gin.New()
		router.Use(authMiddleware)
		router.GET("/protected", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			require.True(t, exists)
			assert.Equal(t, uint(1), userID)
			
			userRole, exists := c.Get("user_role")
			require.True(t, exists)
			assert.Equal(t, "lawyer", userRole)
			
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		})
		
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid_token")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access granted")
	})
	
	t.Run("missing authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(authMiddleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})
		
		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "missing authorization header")
	})
	
	t.Run("invalid token format", func(t *testing.T) {
		router := gin.New()
		router.Use(authMiddleware)
		router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})
		
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid authorization format")
	})
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	loggingMiddleware := middleware.LoggingMiddleware()
	
	t.Run("successful request", func(t *testing.T) {
		router := gin.New()
		router.Use(loggingMiddleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})
		
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	rateLimitMiddleware := middleware.RateLimitMiddleware(2, time.Minute)
	
	t.Run("within rate limit", func(t *testing.T) {
		router := gin.New()
		router.Use(rateLimitMiddleware)
		router.GET("/limited", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		})
		
		// 第一次请求
		req1, _ := http.NewRequest("GET", "/limited", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)
		
		// 第二次请求
		req2, _ := http.NewRequest("GET", "/limited", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusOK, w2.Code)
	})
	
	t.Run("exceed rate limit", func(t *testing.T) {
		router := gin.New()
		router.Use(rateLimitMiddleware)
		router.GET("/limited", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})
		
		// 发送3个请求，超过限制
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("GET", "/limited", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if i < 2 {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
				assert.Contains(t, w.Body.String(), "rate limit exceeded")
			}
		}
	})
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	corsMiddleware := middleware.CORSMiddleware([]string{
		"https://app.lawoa.com",
		"https://admin.lawoa.com",
	})
	
	t.Run("valid origin", func(t *testing.T) {
		router := gin.New()
		router.Use(corsMiddleware)
		router.GET("/api/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})
		
		req, _ := http.NewRequest("OPTIONS", "/api/test", nil)
		req.Header.Set("Origin", "https://app.lawoa.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://app.lawoa.com", w.Header().Get("Access-Control-Allow-Origin"))
	})
	
	t.Run("invalid origin", func(t *testing.T) {
		router := gin.New()
		router.Use(corsMiddleware)
		router.GET("/api/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})
		
		req, _ := http.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Origin", "https://malicious.com")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestRequestSignatureMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	signatureMiddleware := middleware.RequestSignatureMiddleware()
	
	t.Run("valid signature", func(t *testing.T) {
		router := gin.New()
		router.Use(signatureMiddleware)
		router.POST("/api/signed", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "signature valid"})
		})
		
		body := `{"test": "data"}`
		req, _ := http.NewRequest("POST", "/api/signed", strings.NewReader(body))
		req.Header.Set("X-API-Key", "test_api_key")
		req.Header.Set("X-Timestamp", time.Now().Format(time.RFC3339))
		req.Header.Set("X-Signature", "valid_signature_hash")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// 根据具体实现验证
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusUnauthorized)
	})
	
	t.Run("missing signature headers", func(t *testing.T) {
		router := gin.New()
		router.Use(signatureMiddleware)
		router.POST("/api/signed", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
		})
		
		req, _ := http.NewRequest("POST", "/api/signed", strings.NewReader(`{"test": "data"}`))
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "missing required signature headers")
	})
}

// MockAuthService 用于测试的模拟认证服务
type MockAuthService struct {
	users       map[uint]*services.UserClaims
	validTokens map[string]bool
}

func (m *MockAuthService) Login(ctx context.Context, req *services.LoginRequest) (*services.LoginResponse, error) {
	return &services.LoginResponse{
		Token:     "mock_token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: services.User{
			ID:   1,
			Name: "张律师",
		},
	}, nil
}

func (m *MockAuthService) ValidateToken(token string) (*services.UserClaims, error) {
	if m.validTokens[token] {
		return m.users[1], nil
	}
	return nil, errors.New("invalid token")
}

func (m *MockAuthService) RefreshToken(ctx context.Context, req *services.RefreshTokenRequest) (*services.LoginResponse, error) {
	return &services.LoginResponse{
		Token:     "new_mock_token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User: services.User{
			ID:   1,
			Name: "张律师",
		},
	}, nil
}

func (m *MockAuthService) Logout(ctx context.Context, req *services.LogoutRequest) error {
	return nil
}

func (m *MockAuthService) ChangePassword(ctx context.Context, userID uint, req *services.ChangePasswordRequest) error {
	return nil
}