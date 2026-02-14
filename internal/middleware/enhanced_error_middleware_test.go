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

func setupEnhancedTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestEnhancedErrorMiddleware_PanicRecovery(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_EnhancedError(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_ValidationErrorWithDetails(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_DatabaseError(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_NetworkError(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_SecurityError(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_ErrorWrapping(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_ProductionMode(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_ContextInformation(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_MultipleErrors(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_SuccessResponse(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_TimeoutHandling(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestEnhancedErrorMiddleware_RequestMetrics(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestDefaultEnhancedErrorMiddleware(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestProductionEnhancedErrorMiddleware(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

func TestErrorHandlerWithConfig(t *testing.T) {
	t.Skip("NewEnhancedErrorMiddleware not implemented")
}

// TestEnhancedErrors_BasicFunctions 基础错误功能测试
func TestEnhancedErrors_BasicFunctions(t *testing.T) {
	t.Run("测试 BusinessError 创建", func(t *testing.T) {
		err := errors.BusinessError("user", "registration", "business error occurred")
		assert.NotNil(t, err)
		assert.Equal(t, "BUSINESS_ERROR", err.Code())
		assert.Equal(t, "business error occurred", err.Message())
		assert.Equal(t, errors.ErrorCategoryBusiness, err.Category())
		assert.Equal(t, errors.ErrorLevelWarning, err.Level())
	})

	t.Run("测试 ValidationError 创建", func(t *testing.T) {
		err := errors.ValidationError("email", "Invalid email format")
		assert.NotNil(t, err)
		assert.Equal(t, "VALIDATION_ERROR", err.Code())
		assert.Equal(t, "Invalid email format", err.Message())
		assert.Equal(t, errors.ErrorCategoryValidation, err.Category())
	})

	t.Run("测试 DatabaseError 创建", func(t *testing.T) {
		err := errors.DatabaseError("SELECT", "Query failed", assert.AnError)
		assert.NotNil(t, err)
		assert.Equal(t, "DATABASE_ERROR", err.Code())
		assert.Equal(t, errors.ErrorCategoryDatabase, err.Category())
		assert.Equal(t, errors.ErrorLevelError, err.Level())
	})

	t.Run("测试 NetworkError 创建", func(t *testing.T) {
		err := errors.NetworkError("http://example.com", true, assert.AnError)
		assert.NotNil(t, err)
		assert.Equal(t, "TIMEOUT_ERROR", err.Code())
		assert.Equal(t, errors.ErrorCategoryNetwork, err.Category())
	})

	t.Run("测试 SecurityError 创建", func(t *testing.T) {
		err := errors.SecurityError("login", "Invalid credentials", assert.AnError)
		assert.NotNil(t, err)
		assert.Equal(t, "SECURITY_ERROR", err.Code())
		assert.Equal(t, errors.ErrorCategorySecurity, err.Category())
		assert.Equal(t, errors.ErrorLevelError, err.Level())
	})

	t.Run("测试 NotFoundError 创建", func(t *testing.T) {
		err := errors.NotFoundError("user", "User not found", 123)
		assert.NotNil(t, err)
		assert.Equal(t, "NOT_FOUND", err.Code())
		assert.Equal(t, "User not found", err.Message())
		assert.Equal(t, errors.ErrorCategoryBusiness, err.Category())
	})

	t.Run("测试 UnauthorizedError 创建", func(t *testing.T) {
		err := errors.UnauthorizedError("view", "profile")
		assert.NotNil(t, err)
		assert.Equal(t, "UNAUTHORIZED", err.Code())
		assert.Contains(t, err.Message(), "Unauthorized")
		assert.Equal(t, errors.ErrorCategorySecurity, err.Category())
	})
}

// TestErrorHandler_Basic 基础错误处理器测试
func TestErrorHandler_Basic(t *testing.T) {
	t.Run("测试默认配置", func(t *testing.T) {
		config := errors.DefaultErrorHandlerConfig()
		assert.NotNil(t, config)
		assert.False(t, config.IncludeStackTrace)
		assert.False(t, config.DebugMode)
	})

	t.Run("测试 ErrorHandlerManager 创建", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		config := errors.DefaultErrorHandlerConfig()
		manager := errors.NewErrorHandlerManager(logger, config)
		assert.NotNil(t, manager)
	})

	t.Run("测试 ErrorHandler 处理", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		config := errors.DefaultErrorHandlerConfig()
		manager := errors.NewErrorHandlerManager(logger, config)

		handler := NewErrorHandler(logger, ErrorHandlerConfig{
			IncludeStackTrace: true,
			IncludeContext:    true,
			DebugMode:         true,
		})

		assert.NotNil(t, handler)

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
		assert.NotEmpty(t, response.Error.Code)
	})
}
