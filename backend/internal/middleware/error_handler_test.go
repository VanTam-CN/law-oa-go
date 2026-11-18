//go:build ignore
// +build ignore

package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestErrorHandlingMiddleware_PanicError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{
		IncludeStackTrace: true,
		IncludeContext:    true,
		DebugMode:         true,
	})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

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

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "PANIC_ERROR", response.Error.Code)
	assert.Contains(t, response.Error.Message, "test panic")
	assert.Equal(t, "test-request-id", response.RequestID)
}

func TestErrorHandlingMiddleware_BusinessError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{
		IncludeStackTrace: true,
		IncludeContext:    true,
		DebugMode:         true,
	})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	// 创建一个返回业务错误的路由
	r.GET("/business-error", func(c *gin.Context) {
		bizErr := errors.NewBusinessError("BUSINESS_ERROR", "business error occurred", nil)
		bizErr.AddContext("user_id", 123)
		_ = c.Error(bizErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/business-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("User-Agent", "test-agent")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "BUSINESS_ERROR", response.Error.Code)
	assert.Equal(t, "business error occurred", response.Error.Message)
	assert.NotNil(t, response.Error.Context)
	assert.Equal(t, 123.0, response.Error.Context["user_id"]) // JSON数字会被解析为float64
	assert.Equal(t, "test-request-id", response.RequestID)
}

func TestErrorHandlingMiddleware_ValidationError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{
		IncludeSuggestions: true,
	})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	// 创建一个返回验证错误的路由
	r.GET("/validation-error", func(c *gin.Context) {
		validationErr := errors.NewValidationError("email", "invalid_email_format", "invalid email format", "email")
		_ = c.Error(validationErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/validation-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Contains(t, response.Error.Message, "email")
	assert.NotNil(t, response.Error.Suggestions)
	assert.NotEmpty(t, response.Error.Suggestions)
}

func TestErrorHandlingMiddleware_AuthorizationError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	// 创建一个返回权限错误的路由
	r.GET("/auth-error", func(c *gin.Context) {
		authErr := errors.NewAuthorizationError("authorization_error", "access denied", "admin", "user")
		_ = c.Error(authErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "AUTHORIZATION_ERROR", response.Error.Code)
	assert.Equal(t, "access denied", response.Error.Message)
}

func TestErrorHandlingMiddleware_MultipleErrors(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	// 创建一个返回多个错误的路由
	r.GET("/multiple-errors", func(c *gin.Context) {
		err1 := errors.NewBusinessError("ERROR_1", "first error", nil)
		err2 := errors.NewBusinessError("ERROR_2", "second error", nil)
		_ = c.Error(err1)
		_ = c.Error(err2)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/multiple-errors", nil)

	r.ServeHTTP(w, req)

	// 验证响应 - 应该只处理最后一个错误
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "ERROR_2", response.Error.Code)
}

func TestErrorHandlingMiddleware_NoError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	// 创建一个正常路由
	r.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/success", nil)

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["message"])
}

func TestErrorHandlingMiddleware_Context(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{
		IncludeContext: true,
	})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	// 创建一个带有用户上下文的路由
	r.GET("/context-error", func(c *gin.Context) {
		c.Set("user_id", "123")
		c.Set("username", "testuser")
		c.Set("role", "admin")

		bizErr := errors.NewBusinessError("CONTEXT_ERROR", "context test error", nil)
		_ = c.Error(bizErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/context-error", nil)
	req.Header.Set("X-Request-ID", "test-request-id")
	req.Header.Set("User-Agent", "test-agent")

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Error.Context)
	assert.Equal(t, "test-request-id", response.Error.Context["request_id"])
	assert.Equal(t, "GET", response.Error.Context["method"])
	assert.Equal(t, "/context-error", response.Error.Context["path"])
	assert.Equal(t, "123", response.Error.Context["user_id"])
	assert.Equal(t, "testuser", response.Error.Context["username"])
	assert.Equal(t, "admin", response.Error.Context["role"])
}

func TestErrorHandlingMiddleware_ProductionMode(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{
		IncludeStackTrace: false,
		IncludeContext:    false,
		DebugMode:         false,
	})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	// 创建一个会panic的路由
	r.GET("/production-panic", func(c *gin.Context) {
		panic("production panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/production-panic", nil)

	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "PANIC_ERROR", response.Error.Code)
	assert.Nil(t, response.Error.Context)    // 生产模式下不包含上下文
	assert.Nil(t, response.Error.StackTrace) // 生产模式下不包含堆栈跟踪
}

func TestErrorHandler_GetSuggestions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{
		IncludeSuggestions: true,
	})

	testCases := []struct {
		name         string
		errorCode    string
		expectedSugs []string
	}{
		{
			name:      "validation error",
			errorCode: "VALIDATION_ERROR",
			expectedSugs: []string{
				"请检查输入数据的格式和内容",
				"确保所有必填字段都已填写",
				"检查数据类型是否正确",
			},
		},
		{
			name:      "database error",
			errorCode: "DATABASE_ERROR",
			expectedSugs: []string{
				"请稍后重试",
				"如果问题持续存在，请联系系统管理员",
			},
		},
		{
			name:      "authorization error",
			errorCode: "AUTHORIZATION_ERROR",
			expectedSugs: []string{
				"请确认您有访问此资源的权限",
				"请联系管理员获取相应权限",
			},
		},
		{
			name:      "unknown error",
			errorCode: "UNKNOWN_ERROR",
			expectedSugs: []string{
				"请稍后重试",
				"如果问题持续存在，请联系技术支持",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err errors.AppError
			switch tc.errorCode {
			case "VALIDATION_ERROR":
				err = errors.NewValidationError("field", tc.errorCode, "validation error", "details")
			case "DATABASE_ERROR":
				err = errors.NewDatabaseError(tc.errorCode, "database error", nil)
			case "AUTHORIZATION_ERROR":
				err = errors.NewAuthorizationError(tc.errorCode, "authorization error", "required", "current")
			case "CONCURRENCY_ERROR":
				err = errors.NewConcurrencyError(tc.errorCode, "resource", "concurrency error", nil)
			case "NETWORK_ERROR":
				err = errors.NewNetworkError(tc.errorCode, false, nil)
			case "NOT_FOUND":
				err = errors.NewNotFoundError("resource", tc.errorCode, nil)
			default:
				err = errors.NewInternalError(tc.errorCode, "internal error", nil)
			}
			suggestions := handler.getSuggestions(err)
			assert.Equal(t, tc.expectedSugs, suggestions)
		})
	}
}

func TestErrorHandler_SeverityToString(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{})

	testCases := []struct {
		severity errors.ErrorSeverity
		expected string
	}{
		{errors.SeverityLow, "LOW"},
		{errors.SeverityMedium, "MEDIUM"},
		{errors.SeverityHigh, "HIGH"},
		{errors.SeverityCritical, "CRITICAL"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := handler.severityToString(tc.severity)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// 测试默认错误处理中间件
	defaultMiddleware := DefaultErrorHandlingMiddleware(logger)
	assert.NotNil(t, defaultMiddleware)

	// 测试生产环境错误处理中间件
	productionMiddleware := ProductionErrorHandlingMiddleware(logger)
	assert.NotNil(t, productionMiddleware)
}

func BenchmarkErrorHandlingMiddleware(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	handler := NewErrorHandler(logger, ErrorHandlerConfig{})

	r := setupTestGin()
	r.Use(handler.ErrorHandlingMiddleware())

	r.GET("/error", func(c *gin.Context) {
		err := errors.NewBusinessError("BENCHMARK_ERROR", "benchmark error", nil)
		_ = c.Error(err)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/error", nil)
		r.ServeHTTP(w, req)
	}
}
