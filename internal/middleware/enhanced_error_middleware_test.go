package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"law-oa-go/internal/errors"
)

func setupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestEnhancedErrorMiddleware_PanicRecovery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个会panic的路由
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "PANIC_ERROR", errorObj["code"])
	assert.Contains(t, errorObj["message"].(string), "test panic")
	assert.Equal(t, "test-request-id", response["request_id"])
	assert.NotNil(t, errorObj["stack_trace"])
}

func TestEnhancedErrorMiddleware_EnhancedError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个返回增强业务错误的路由
	r.GET("/business-error", func(c *gin.Context) {
		bizErr := errors.BusinessError("user", "registration", "business error occurred").
			Context("user_id", 123).
			SetTraceID(c.GetString("request_id")).
			Build()
		_ = c.Error(bizErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/business-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("User-Agent", "test-agent")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "BUSINESS_ERROR", errorObj["code"])
	assert.Equal(t, "business error occurred", errorObj["message"])
	assert.Equal(t, "business", errorObj["category"])
	assert.Equal(t, "WARNING", errorObj["level"])

	// 验证上下文信息
	contextObj := errorObj["context"].(map[string]interface{})
	assert.Equal(t, 123.0, contextObj["user_id"]) // JSON数字会被解析为float64
	assert.Equal(t, "test-request-id", errorObj["trace_id"])
	assert.Equal(t, "GET", contextObj["method"])
	assert.Equal(t, "/business-error", contextObj["path"])
	assert.Equal(t, "::1", contextObj["client_ip"])
	assert.Equal(t, "test-agent", contextObj["user_agent"])
}

func TestEnhancedErrorMiddleware_ValidationErrorWithDetails(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个返回详细验证错误的路由
	r.GET("/validation-error", func(c *gin.Context) {
		validationErr := errors.ValidationErrorWithDetails(
			"email",
			"Invalid email format",
			"Email field contains invalid characters",
			[]string{"must be valid email", "cannot contain special chars"},
		).
			Context("field_value", "invalid-email@").
			SetTraceID(c.GetString("request_id")).
			Build()
		_ = c.Error(validationErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validation-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errorObj["code"])
	assert.Equal(t, "Invalid email format", errorObj["message"])
	assert.Equal(t, "Email field contains invalid characters", errorObj["details"])
	assert.NotNil(t, errorObj["suggestions"])
	assert.NotEmpty(t, errorObj["suggestions"])

	// 验证规则信息
	contextObj := errorObj["context"].(map[string]interface{})
	assert.Equal(t, []interface{}{"must be valid email", "cannot contain special chars"}, contextObj["rules"])
	assert.Equal(t, "invalid-email@", contextObj["field_value"])
}

func TestEnhancedErrorMiddleware_DatabaseError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个返回数据库错误的路由
	r.GET("/database-error", func(c *gin.Context) {
		dbErr := errors.DatabaseError("SELECT", "Failed to execute query",
			errors.New("connection failed")).
			Context("query", "SELECT * FROM users").
			Context("table", "users").
			SetTraceID(c.GetString("request_id")).
			Build()
		_ = c.Error(dbErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/database-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "DATABASE_ERROR", errorObj["code"])
	assert.Equal(t, "database", errorObj["category"])
	assert.Equal(t, "ERROR", errorObj["level"])
}

func TestEnhancedErrorMiddleware_NetworkError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个返回网络错误的路由
	r.GET("/network-error", func(c *gin.Context) {
		netErr := errors.NetworkError("https://api.example.com", true,
			errors.New("timeout after 30 seconds")).
			SetTraceID(c.GetString("request_id")).
			Build()
		_ = c.Error(netErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/network-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "TIMEOUT_ERROR", errorObj["code"])
	assert.Equal(t, "network", errorObj["category"])
	assert.Equal(t, "ERROR", errorObj["level"])
	assert.Equal(t, true, errorObj["context"].(map[string]interface{})["timeout"])
}

func TestEnhancedErrorMiddleware_SecurityError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个返回安全错误的路由
	r.GET("/security-error", func(c *gin.Context) {
		secErr := errors.SecurityError("login_attempt", "Too many failed attempts",
			errors.New("rate limit exceeded")).
			Context("ip_address", "192.168.1.100").
			Context("attempts", 5).
			SetTraceID(c.GetString("request_id")).
			Build()
		_ = c.Error(secErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/security-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "SECURITY_ERROR", errorObj["code"])
	assert.Equal(t, "security", errorObj["category"])
	assert.Equal(t, "ERROR", errorObj["level"])
}

func TestEnhancedErrorMiddleware_ErrorWrapping(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个返回包装错误的路由
	r.GET("/wrapped-error", func(c *gin.Context) {
		originalErr := errors.NewError("ORIGINAL_ERROR", "Original error occurred").
			Category(errors.ErrorCategoryValidation).
			Level(errors.ErrorLevelWarning).
			Build()

		wrappedErr := errors.NewError("WRAPPER_ERROR", "Wrapper error message").
			Category(errors.ErrorCategorySystem).
			Level(errors.ErrorLevelError).
			Cause(originalErr).
			Build()
		_ = c.Error(wrappedErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/wrapped-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "WRAPPER_ERROR", errorObj["code"])
	assert.Equal(t, "Wrapper error message", errorObj["message"])
	assert.Equal(t, "system", errorObj["category"])
}

func TestEnhancedErrorMiddleware_ProductionMode(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.ProductionErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个会panic的路由
	r.GET("/production-panic", func(c *gin.Context) {
		panic("production panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/production-panic", nil)

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "PANIC_ERROR", errorObj["code"])
	assert.Nil(t, errorObj["context"])    // 生产模式下不包含上下文
	assert.Nil(t, errorObj["stack_trace"]) // 生产模式下不包含堆栈跟踪
	assert.Nil(t, errorObj["source"])     // 生产模式下不包含源码信息
}

func TestEnhancedErrorMiddleware_ContextInformation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个带有用户上下文的路由
	r.GET("/context-error", func(c *gin.Context) {
		c.Set("user_id", uint(123))
		c.Set("username", "testuser")
		c.Set("role", "admin")

		bizErr := errors.BusinessError("user", "profile_update", "Failed to update profile").
			SetTraceID(c.GetString("request_id")).
			Build()
		_ = c.Error(bizErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/context-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("User-Agent", "test-agent")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	errorObj := response["error"].(map[string]interface{})
	contextObj := errorObj["context"].(map[string]interface{})

	// 验证请求信息
	assert.Equal(t, "test-request-id", contextObj["request_id"])
	assert.Equal(t, "GET", contextObj["method"])
	assert.Equal(t, "/context-error", contextObj["path"])
	assert.Equal(t, "::1", contextObj["client_ip"])
	assert.Equal(t, "test-agent", contextObj["user_agent"])

	// 验证用户信息
	assert.Equal(t, float64(123), contextObj["user_id"])
	assert.Equal(t, "testuser", contextObj["username"])
	assert.Equal(t, "admin", contextObj["role"])
}

func TestEnhancedErrorMiddleware_MultipleErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个返回多个错误的路由
	r.GET("/multiple-errors", func(c *gin.Context) {
		err1 := errors.BusinessError("user", "validation", "First error occurred").Build()
		err2 := errors.ValidationError("email", "Second error occurred").Build()
		_ = c.Error(err1)
		_ = c.Error(err2)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/multiple-errors", nil)

	r.ServeHTTP(w, req)

	// 验证响应 - 应该只处理最后一个错误
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))
	assert.NotNil(t, response["error"])
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "VALIDATION_ERROR", errorObj["code"])
}

func TestEnhancedErrorMiddleware_SuccessResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个正常路由
	r.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/success", nil)

	r.ServeHTTP(w, req)

	// 验证响应 - 中间件不应该干扰正常响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["message"])
}

func TestEnhancedErrorMiddleware_TimeoutHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)

	// 使用带超时的中间件
	timeoutMiddleware := TimeoutMiddleware(100 * time.Millisecond)

	r := setupTestGin()
	r.Use(timeoutMiddleware)
	r.Use(NewEnhancedErrorMiddleware(manager).Middleware())

	// 创建一个长时间运行的路由
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond) // 超过超时时间
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/slow", nil)

	r.ServeHTTP(w, req)

	// 验证响应应该返回超时错误
	assert.Equal(t, http.StatusRequestTimeout, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(err)

	assert.False(t, response["success"].(bool))
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "REQUEST_TIMEOUT", errorObj["code"])
}

func TestEnhancedErrorMiddleware_RequestMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	// 创建一个快速请求路由
	r.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fast", nil)
	req.Header.Set("X-Request-ID", "metrics-test")

	start := time.Now()
	r.ServeHTTP(w, req)
	duration := time.Since(start)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 验证响应包含时间指标（如果添加的话）
	assert.Equal(t, "success", response["message"])
	assert.Less(t, 100*time.Millisecond, duration)
}

func TestDefaultEnhancedErrorMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// 测试默认中间件创建函数
	middleware := DefaultEnhancedErrorMiddleware(logger)
	assert.NotNil(t, middleware)

	// 测试中间件功能
	r := setupTestGin()
	r.Use(middleware)

	r.GET("/test", func(c *gin.Context) {
		err := errors.ValidationError("field", "validation error").Build()
		_ = c.Error(err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestProductionEnhancedErrorMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// 测试生产环境中间件创建函数
	middleware := ProductionEnhancedErrorMiddleware(logger)
	assert.NotNil(t, middleware)

	// 测试中间件功能
	r := setupTestGin()
	r.Use(middleware)

	r.GET("/test", func(c *gin.Context) {
		err := errors.SystemError("component", "production error", nil).Build()
		_ = c.Error(err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestErrorHandlerWithConfig(t *testing.T) {
	logger := slog.NewTextHandler(&bytes.Buffer{}, nil)

	config := errors.ErrorHandlerConfig{
		EnableStackTrace:   true,
		EnableContext:      true,
		EnableSuggestions:  true,
		DebugMode:           true,
		LogLevel:           "debug",
		AlertEnabled:       false,
		MaxRetries:         0,
		RetryDelay:         0,
	}

	middleware := ErrorHandlerWithConfig(logger, config)
	assert.NotNil(tiddleware)

	// 测试自定义配置的中间件
	r := setupTestGin()
	r.Use(middleware)

	r.GET("/test", func(c *gin.Context) {
		err := errors.DatabaseError("operation", "database error", nil).Build()
		_ = c.Error(err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func BenchmarkEnhancedErrorMiddleware(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	r.GET("/error", func(c *gin.Context) {
		err := errors.BusinessError("test", "benchmark error", nil).Build()
		_ = c.Error(err)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/error", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}

func BenchmarkPanicRecovery(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	config := errors.DefaultErrorHandlerConfig()
	manager := errors.NewErrorHandlerManager(logger, config)
	middleware := NewEnhancedErrorMiddleware(manager)

	r := setupTestGin()
	r.Use(middleware.Middleware())

	r.GET("/panic", func(c *gin.Context) {
		panic("benchmark panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
		w.Body.Reset()
	}
}