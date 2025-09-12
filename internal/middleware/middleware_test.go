package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa/test"
)

// MiddlewareTest 中间件测试
func TestMiddleware(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("RequestIDMiddleware", func(t *testing.T) {
		// 测试请求ID中间件
		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			requestID := c.GetHeader("X-Request-ID")
			c.JSON(200, gin.H{"request_id": requestID})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response["request_id"])
	})
	
	t.Run("CORSMiddleware", func(t *testing.T) {
		// 测试CORS中间件
		router := gin.New()
		router.Use(CORSMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "http://localhost:3000")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	})
	
	t.Run("LoggingMiddleware", func(t *testing.T) {
		// 测试日志中间件
		router := gin.New()
		router.Use(LoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		// 日志中间件应该正常工作，不会影响响应
	})
	
	t.Run("RecoveryMiddleware", func(t *testing.T) {
		// 测试恢复中间件
		router := gin.New()
		router.Use(RecoveryMiddleware())
		router.GET("/test", func(c *gin.Context) {
			panic("test panic")
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 500, w.Code)
		// 恢复中间件应该捕获panic并返回错误响应
	})
	
	t.Run("RateLimitMiddleware", func(t *testing.T) {
		// 测试限流中间件
		router := gin.New()
		router.Use(RateLimitMiddleware(10, time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		// 测试正常请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		
		// 测试限流
		for i := 0; i < 15; i++ {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/test", nil)
			router.ServeHTTP(w, req)
		}
		assert.Equal(t, 429, w.Code)
	})
	
	t.Run("SecurityMiddleware", func(t *testing.T) {
		// 测试安全中间件
		router := gin.New()
		router.Use(SecurityMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Header().Get("X-Content-Type-Options"), "nosniff")
		assert.Contains(t, w.Header().Get("X-Frame-Options"), "DENY")
		assert.Contains(t, w.Header().Get("X-XSS-Protection"), "1; mode=block")
	})
}

// JWTMiddlewareTest JWT中间件测试
func TestJWTMiddleware(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("ValidJWT", func(t *testing.T) {
		// 测试有效JWT
		router := gin.New()
		router.Use(JWTMiddleware())
		router.GET("/test", func(c *gin.Context) {
			userID := c.GetUint("user_id")
			c.JSON(200, gin.H{"user_id": userID})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+suite.AuthToken)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, suite.TestUser.ID, uint(response["user_id"].(float64)))
	})
	
	t.Run("MissingJWT", func(t *testing.T) {
		// 测试缺失JWT
		router := gin.New()
		router.Use(JWTMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 401, w.Code)
	})
	
	t.Run("InvalidJWT", func(t *testing.T) {
		// 测试无效JWT
		router := gin.New()
		router.Use(JWTMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 401, w.Code)
	})
	
	t.Run("ExpiredJWT", func(t *testing.T) {
		// 测试过期JWT
		router := gin.New()
		router.Use(JWTMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		// 创建过期的JWT
		cfg, err := config.LoadTestConfig()
		require.NoError(t, err)
		
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  suite.TestUser.ID,
			"username": suite.TestUser.Username,
			"role":     suite.TestUser.Role,
			"exp":      time.Now().Add(-time.Hour).Unix(), // 过期时间
		})
		
		tokenString, err := token.SignedString([]byte(cfg.JWT.Secret))
		require.NoError(t, err)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 401, w.Code)
	})
}

// RoleMiddlewareTest 角色中间件测试
func TestRoleMiddleware(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("ValidRole", func(t *testing.T) {
		// 测试有效角色
		router := gin.New()
		router.Use(JWTMiddleware())
		router.Use(RoleMiddleware("admin"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		// 创建管理员用户
		adminUser := &models.User{
			Username: "admin",
			Email:    "admin@example.com",
			Password: "password123",
			FirstName: "Admin",
			LastName: "User",
			Role:     "admin",
			Status:   "active",
		}
		err := suite.DB.Create(adminUser).Error
		require.NoError(t, err)
		
		adminToken := config.GenerateTestJWT(t, adminUser)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
	})
	
	t.Run("InvalidRole", func(t *testing.T) {
		// 测试无效角色
		router := gin.New()
		router.Use(JWTMiddleware())
		router.Use(RoleMiddleware("admin"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+suite.AuthToken) // 普通用户
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 403, w.Code)
	})
	
	t.Run("MultipleRoles", func(t *testing.T) {
		// 测试多个角色
		router := gin.New()
		router.Use(JWTMiddleware())
		router.Use(RoleMiddleware("admin", "manager"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		// 创建经理用户
		managerUser := &models.User{
			Username: "manager",
			Email:    "manager@example.com",
			Password: "password123",
			FirstName: "Manager",
			LastName: "User",
			Role:     "manager",
			Status:   "active",
		}
		err := suite.DB.Create(managerUser).Error
		require.NoError(t, err)
		
		managerToken := config.GenerateTestJWT(t, managerUser)
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+managerToken)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
	})
}

// ValidationMiddlewareTest 验证中间件测试
func TestValidationMiddleware(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("ValidRequest", func(t *testing.T) {
		// 测试有效请求
		type TestRequest struct {
			Name  string `json:"name" binding:"required"`
			Email string `json:"email" binding:"required,email"`
		}
		
		router := gin.New()
		router.Use(ValidationMiddleware[TestRequest]())
		router.POST("/test", func(c *gin.Context) {
			var req TestRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"message": "valid"})
		})
		
		validRequest := map[string]interface{}{
			"name":  "Test User",
			"email": "test@example.com",
		}
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(json.Marshal(validRequest)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
	})
	
	t.Run("InvalidRequest", func(t *testing.T) {
		// 测试无效请求
		type TestRequest struct {
			Name  string `json:"name" binding:"required"`
			Email string `json:"email" binding:"required,email"`
		}
		
		router := gin.New()
		router.Use(ValidationMiddleware[TestRequest]())
		router.POST("/test", func(c *gin.Context) {
			var req TestRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"message": "valid"})
		})
		
		invalidRequest := map[string]interface{}{
			"name":  "", // 空名字
			"email": "invalid-email", // 无效邮箱
		}
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(json.Marshal(invalidRequest)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 400, w.Code)
	})
}

// CacheMiddlewareTest 缓存中间件测试
func TestCacheMiddleware(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("CacheHit", func(t *testing.T) {
		// 测试缓存命中
		router := gin.New()
		router.Use(CacheMiddleware(10*time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"data": "test", "timestamp": time.Now().Unix()})
		})
		
		// 第一次请求
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w1, req1)
		
		assert.Equal(t, 200, w1.Code)
		assert.Contains(t, w1.Header().Get("Cache-Control"), "max-age=600")
		
		// 第二次请求应该从缓存返回
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w2, req2)
		
		assert.Equal(t, 200, w2.Code)
		assert.Equal(t, w1.Body.String(), w2.Body.String())
	})
	
	t.Run("CacheBypass", func(t *testing.T) {
		// 测试缓存绕过
		router := gin.New()
		router.Use(CacheMiddleware(10*time.Minute))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"data": "test", "timestamp": time.Now().Unix()})
		})
		
		// POST请求不应该被缓存
		w1 := httptest.NewRecorder()
		req1, _ := http.NewRequest("POST", "/test", nil)
		router.ServeHTTP(w1, req1)
		
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("POST", "/test", nil)
		router.ServeHTTP(w2, req2)
		
		assert.Equal(t, 200, w1.Code)
		assert.Equal(t, 200, w2.Code)
		// POST请求的响应应该不同
		assert.NotEqual(t, w1.Body.String(), w2.Body.String())
	})
}

// MetricsMiddlewareTest 指标中间件测试
func TestMetricsMiddleware(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("MetricsCollection", func(t *testing.T) {
		// 测试指标收集
		router := gin.New()
		router.Use(MetricsMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})
		
		// 发送请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		// 指标中间件应该正常工作，不会影响响应
	})
	
	t.Run("ErrorMetrics", func(t *testing.T) {
		// 测试错误指标
		router := gin.New()
		router.Use(MetricsMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(500, gin.H{"error": "internal error"})
		})
		
		// 发送请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 500, w.Code)
		// 错误指标应该被记录
	})
}

// DatabaseMiddlewareTest 数据库中间件测试
func TestDatabaseMiddleware(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("DatabaseConnection", func(t *testing.T) {
		// 测试数据库连接
		router := gin.New()
		router.Use(DatabaseMiddleware(suite.DB))
		router.GET("/test", func(c *gin.Context) {
			db := c.MustGet("db").(*gorm.DB)
			
			var userCount int64
			db.Model(&models.User{}).Count(&userCount)
			
			c.JSON(200, gin.H{"user_count": userCount})
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, response["user_count"].(float64), 1.0)
	})
	
	t.Run("Transaction", func(t *testing.T) {
		// 测试事务
		router := gin.New()
		router.Use(DatabaseMiddleware(suite.DB))
		router.Use(TransactionMiddleware())
		router.POST("/test", func(c *gin.Context) {
			db := c.MustGet("db").(*gorm.DB)
			
			// 在事务中创建用户
			user := &models.User{
				Username: "transaction_test",
				Email:    "transaction@example.com",
				Password: "password123",
				FirstName: "Transaction",
				LastName: "Test",
				Role:     "user",
				Status:   "active",
			}
			
			err := db.Create(user).Error
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			
			c.JSON(200, gin.H{"user_id": user.ID})
		})
		
		userData := map[string]interface{}{
			"username": "transaction_test",
			"email":    "transaction@example.com",
			"password": "password123",
			"first_name": "Transaction",
			"last_name": "Test",
			"role":      "user",
			"status":    "active",
		}
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", strings.NewReader(json.Marshal(userData)))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotZero(t, response["user_id"])
	})
}