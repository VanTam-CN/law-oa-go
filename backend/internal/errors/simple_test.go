package errors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBasicErrorCreation 测试基本错误创建
func TestBasicErrorCreation(t *testing.T) {
	err := NewError("TEST_ERROR", "Test error message").
		Category(ErrorCategoryValidation).
		Level(ErrorLevelWarning).
		Details("Detailed error description").
		Context("field", "username").
		Context("value", "invalid").
		Build()

	// 验证基本属性
	assert.Equal(t, "TEST_ERROR", err.Code())
	assert.Equal(t, "Test error message", err.Message())
	assert.Equal(t, ErrorCategoryValidation, err.Category())
	assert.Equal(t, ErrorLevelWarning, err.Level())
	assert.Equal(t, "Detailed error description", err.Details())

	// 验证上下文
	assert.NotNil(t, err.Context())
	assert.Equal(t, "username", err.Context()["field"])
	assert.Equal(t, "invalid", err.Context()["value"])

	// 测试HTTP状态码
	assert.Equal(t, 400, err.HTTPStatus())

	// 测试错误字符串表示
	errorStr := err.Error()
	assert.Contains(t, errorStr, "[WARNING]")
	assert.Contains(t, errorStr, "Test error message")
}

// TestPredefinedErrors 测试预定义错误函数
func TestPredefinedErrors(t *testing.T) {
	// 测试ValidationError
	validationErr := ValidationError("email", "Invalid email format")
	assert.Equal(t, "VALIDATION_ERROR", validationErr.Code())
	assert.Equal(t, ErrorCategoryValidation, validationErr.Category())
	assert.Equal(t, ErrorLevelWarning, validationErr.Level())
	assert.Equal(t, "email", validationErr.Context()["field"])

	// 测试DatabaseError
	dbErr := DatabaseError("SELECT", "Failed to execute query", nil)
	assert.Equal(t, "DATABASE_ERROR", dbErr.Code())
	assert.Equal(t, ErrorCategoryDatabase, dbErr.Category())
	assert.Equal(t, ErrorLevelError, dbErr.Level())
	assert.Equal(t, "SELECT", dbErr.Context()["operation"])
}

// TestErrorMethods 测试错误方法
func TestErrorMethods(t *testing.T) {
	err := NewError("METHOD_TEST", "Method test").
		Category(ErrorCategorySystem).
		Level(ErrorLevelCritical).
		TraceID("test-trace-123").
		Build()

	// 测试ID方法
	assert.NotEmpty(t, err.ID())

	// 测试时间戳方法
	assert.True(t, time.Since(err.Timestamp()) >= 0)

	// 测试跟踪ID方法
	assert.Equal(t, "test-trace-123", err.TraceID())
}

// TestErrorJSONSerialization 测试JSON序列化
func TestErrorJSONSerialization(t *testing.T) {
	err := NewError("JSON_TEST", "JSON serialization test").
		Category(ErrorCategoryNetwork).
		Level(ErrorLevelError).
		Details("Connection timeout after 30 seconds").
		Context("endpoint", "https://api.example.com").
		Build()

	// 测试JSON序列化
	jsonData, marshalErr := err.ToJSON()
	assert.NoError(t, marshalErr)
	assert.NotEmpty(t, jsonData)
}

// TestErrorBuilder 测试错误构建器
func TestErrorBuilder(t *testing.T) {
	cause := NewError("ROOT_CAUSE", "Root cause error").Build()

	err := NewError("BUILDER_TEST", "Builder pattern test").
		Category(ErrorCategoryBusiness).
		Level(ErrorLevelError).
		Details("Testing builder pattern").
		Cause(cause).
		Context("operation", "test").
		HTTPStatus(400).
		Suggestions("Check input", "Try again").
		Build()

	// 验证所有属性
	assert.Equal(t, "BUILDER_TEST", err.Code())
	assert.Equal(t, "Builder pattern test", err.Message())
	assert.Equal(t, ErrorCategoryBusiness, err.Category())
	assert.Equal(t, ErrorLevelError, err.Level())
	assert.Equal(t, "Testing builder pattern", err.Details())
	assert.Equal(t, cause, err.Unwrap())
	assert.Equal(t, 400, err.HTTPStatus())
	assert.Len(t, err.Suggestions(), 2)
}

// TestErrorCheckingFunctions 测试错误类型检查函数
func TestErrorCheckingFunctions(t *testing.T) {
	// 测试IsValidationError
	validationErr := ValidationError("field", "validation error")
	assert.True(t, IsValidationError(validationErr))
	assert.False(t, IsDatabaseError(validationErr))

	// 测试IsDatabaseError
	dbErr := DatabaseError("operation", "database error", nil)
	assert.True(t, IsDatabaseError(dbErr))
	assert.False(t, IsValidationError(dbErr))

	// 测试IsCritical
	criticalErr := NewError("CRITICAL_ERROR", "Critical error").
		Category(ErrorCategorySystem).
		Level(ErrorLevelCritical).
		Build()
	assert.True(t, IsCritical(criticalErr))
	assert.False(t, IsCritical(validationErr))
}