package errors

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseError(t *testing.T) {
	cause := &BaseError{
		code:    "CAUSE_ERROR",
		message: "cause error",
	}

	err := &BaseError{
		code:     "TEST_ERROR",
		message:  "test error",
		details:  "error details",
		cause:    cause,
		severity: SeverityHigh,
		time:     time.Now(),
	}

	// 测试错误信息
	assert.Contains(t, err.Error(), "test error")
	assert.Contains(t, err.Error(), "cause error")

	// 测试错误属性
	assert.Equal(t, "TEST_ERROR", err.Code())
	assert.Equal(t, "test error", err.Message())
	assert.Equal(t, "error details", err.Details())
	assert.Equal(t, SeverityHigh, err.Severity())

	// 测试Unwrap
	assert.Equal(t, cause, err.Unwrap())

	// 测试Is
	assert.True(t, err.Is(&BaseError{code: "TEST_ERROR"}))
	assert.False(t, err.Is(&BaseError{code: "OTHER_ERROR"}))
}

func TestBusinessError(t *testing.T) {
	err := NewBusinessError("BUSINESS_ERROR", "business error", nil)

	// 测试错误类型
	var appErr AppError = err
	assert.IsType(t, &BusinessError{}, appErr)
	businessErr := appErr.(*BusinessError)
	assert.Equal(t, "BUSINESS_ERROR", businessErr.Code())
	assert.Equal(t, 400, businessErr.HTTPStatus())
}

func TestValidationError(t *testing.T) {
	rulesStr := "required, min_length:3"
	err := NewValidationError("username", "username_validation_error", "username too short", rulesStr)

	// 测试错误类型
	var appErr AppError = err
	assert.IsType(t, &ValidationError{}, appErr)
	validationErr := appErr.(*ValidationError)
	assert.Equal(t, "username", validationErr.Field)
	assert.Equal(t, 422, validationErr.HTTPStatus())
}

func TestDatabaseError(t *testing.T) {
	cause := &BaseError{code: "DB_CONNECTION_ERROR"}
	err := NewDatabaseError("SELECT", "Database operation failed", cause)

	// 测试错误类型
	var appErr AppError = err
	assert.IsType(t, &DatabaseError{}, appErr)
	dbErr := appErr.(*DatabaseError)
	assert.Equal(t, "SELECT", dbErr.Operation)
	assert.Equal(t, cause, dbErr.Unwrap())
	assert.Equal(t, 500, dbErr.HTTPStatus())
}

func TestAuthorizationError(t *testing.T) {
	err := NewAuthorizationError("AUTH_ERROR", "access denied", "admin", "user")

	// 测试错误类型
	var appErr AppError = err
	assert.IsType(t, &AuthorizationError{}, appErr)
	authErr := appErr.(*AuthorizationError)
	assert.Equal(t, "admin", authErr.RequiredPermission)
	assert.Equal(t, "user", authErr.CurrentPermission)
	assert.Equal(t, 403, authErr.HTTPStatus())
}

func TestConcurrencyError(t *testing.T) {
	err := NewConcurrencyError("user", 123, "optimistic_lock", nil)

	// 测试错误类型
	var appErr AppError = err
	assert.IsType(t, &ConcurrencyError{}, appErr)
	concErr := appErr.(*ConcurrencyError)
	assert.Equal(t, "user", concErr.ResourceType)
	assert.Equal(t, 123, concErr.ResourceID)
	assert.Equal(t, "optimistic_lock", concErr.ConflictType)
	assert.Equal(t, 409, concErr.HTTPStatus())
}

func TestNetworkError(t *testing.T) {
	err := NewNetworkError("http://api.example.com", true, nil)

	// 测试错误类型
	var appErr AppError = err
	assert.IsType(t, &NetworkError{}, appErr)
	netErr := appErr.(*NetworkError)
	assert.Equal(t, "http://api.example.com", netErr.Endpoint)
	assert.True(t, netErr.Timeout)
	assert.Equal(t, 408, netErr.HTTPStatus())
}

func TestPanicError(t *testing.T) {
	stack := "panic stack trace"
	err := NewPanicError("panic value", stack)

	// 测试错误类型
	var appErr AppError = err
	assert.IsType(t, &PanicError{}, appErr)
	panicErr := appErr.(*PanicError)
	assert.Equal(t, "panic value", panicErr.PanicValue)
	assert.Equal(t, stack, panicErr.StackTrace())
	assert.Equal(t, 500, panicErr.HTTPStatus())
}

func TestContextManagement(t *testing.T) {
	err := NewBusinessError("TEST_ERROR", "test error", nil)

	// 测试添加上下文
	err.AddContext("user_id", 123)
	err.AddContext("action", "create")

	context := err.Context()
	require.NotNil(t, context)
	assert.Equal(t, 123, context["user_id"])
	assert.Equal(t, "create", context["action"])
}

func TestErrorTypeChecking(t *testing.T) {
	businessErr := NewBusinessError("BUSINESS_ERROR", "business error", nil)
	validationErr := NewValidationError("field", "VALIDATION_ERROR", "invalid", "invalid")
	databaseErr := NewDatabaseError("INSERT", "Database operation failed", nil)

	// 测试错误类型检查函数
	assert.True(t, IsBusinessError(businessErr))
	assert.False(t, IsBusinessError(validationErr))

	assert.True(t, IsValidationError(validationErr))
	assert.False(t, IsValidationError(businessErr))

	assert.True(t, IsDatabaseError(databaseErr))
	assert.False(t, IsDatabaseError(businessErr))
}

func TestErrorUtilityFunctions(t *testing.T) {
	err := NewBusinessError("BUSINESS_ERROR", "business error", nil)

	// 测试GetErrorCode
	assert.Equal(t, "BUSINESS_ERROR", GetErrorCode(err))

	// 测试GetHTTPStatus
	assert.Equal(t, 400, GetHTTPStatus(err))

	// 测试WithContext
	contextErr := WithContext(err, "key", "value")
	appErr := contextErr.(AppError)
	require.NotNil(t, appErr.Context())
	assert.Equal(t, "value", appErr.Context()["key"])

	// 测试WithContexts
	multiContext := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	multiContextErr := WithContexts(err, multiContext)
	multiAppErr := multiContextErr.(AppError)
	assert.Equal(t, "value1", multiAppErr.Context()["key1"])
	assert.Equal(t, 123, multiAppErr.Context()["key2"])
}

func TestHelperFunctions(t *testing.T) {
	// 测试NewNotFoundError
	notFoundErr := NewNotFoundError("user", "User not found", 123)
	assert.Equal(t, "NOT_FOUND", notFoundErr.Code())
	assert.Equal(t, "user", notFoundErr.EntityType)
	assert.Equal(t, 123, notFoundErr.EntityID)

	// 测试NewConflictError
	conflictErr := NewConflictError("resource conflict", nil)
	assert.Equal(t, "CONFLICT", conflictErr.Code())
	assert.Equal(t, "resource conflict", conflictErr.Message())

	// 测试NewUnauthorizedError
	unauthorizedErr := NewUnauthorizedError("UNAUTHORIZED", "access denied")
	assert.Equal(t, "UNAUTHORIZED", unauthorizedErr.Code())
	assert.Equal(t, "access denied", unauthorizedErr.Message())
}

func TestErrorWrapping(t *testing.T) {
	originalErr := NewBusinessError("ORIGINAL_ERROR", "original error", nil)
	wrappedErr := NewInternalError("WRAPPER_ERROR", "wrapper error", originalErr)

	// 测试错误链
	assert.Equal(t, originalErr, wrappedErr.Unwrap())
	assert.Contains(t, wrappedErr.Error(), "wrapper error")
	assert.Contains(t, wrappedErr.Error(), "original error")

	// 测试errors.As
	var businessErr *BusinessError
	assert.True(t, errors.As(wrappedErr, &businessErr))
	assert.Equal(t, "ORIGINAL_ERROR", businessErr.Code())

	// 测试errors.Is
	assert.True(t, errors.Is(wrappedErr, originalErr))
}

func TestErrorSeverity(t *testing.T) {
	lowErr := &BaseError{severity: SeverityLow}
	mediumErr := &BaseError{severity: SeverityMedium}
	highErr := &BaseError{severity: SeverityHigh}
	criticalErr := &BaseError{severity: SeverityCritical}

	assert.Equal(t, SeverityLow, lowErr.Severity())
	assert.Equal(t, SeverityMedium, mediumErr.Severity())
	assert.Equal(t, SeverityHigh, highErr.Severity())
	assert.Equal(t, SeverityCritical, criticalErr.Severity())
}

func BenchmarkErrorCreation(b *testing.B) {
	b.Run("BusinessError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewBusinessError("BUSINESS_ERROR", "business error", nil)
		}
	})

	b.Run("ValidationError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewValidationError("field", "VALIDATION_ERROR", "invalid", "invalid")
		}
	})

	b.Run("DatabaseError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewDatabaseError("INSERT", "Database operation failed", nil)
		}
	})
}

func BenchmarkErrorChecking(b *testing.B) {
	businessErr := NewBusinessError("BUSINESS_ERROR", "business error", nil)
	validationErr := NewValidationError("field", "VALIDATION_ERROR", "invalid", "invalid")

	b.Run("IsBusinessError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsBusinessError(businessErr)
		}
	})

	b.Run("IsValidationError", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsValidationError(validationErr)
		}
	})
}
