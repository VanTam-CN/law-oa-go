package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// AppError 应用错误接口
type AppError interface {
	error
	Code() string
	Message() string
	Details() string
	Context() map[string]interface{}
	HTTPStatus() int
	StackTrace() string
	Severity() ErrorSeverity
	AddContext(key string, value interface{})
}

// ErrorSeverity 错误严重程度
type ErrorSeverity int

const (
	SeverityLow ErrorSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// BaseError 基础错误实现
type BaseError struct {
	code     string
	message  string
	details  string
	context  map[string]interface{}
	cause    error
	stack    string
	severity ErrorSeverity
	time     time.Time
}

// Error 实现error接口
func (e *BaseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.message, e.cause.Error())
	}
	return e.message
}

// Code 返回错误代码
func (e *BaseError) Code() string {
	return e.code
}

// Message 返回错误消息
func (e *BaseError) Message() string {
	return e.message
}

// Details 返回错误详情
func (e *BaseError) Details() string {
	return e.details
}

// Context 返回错误上下文
func (e *BaseError) Context() map[string]interface{} {
	return e.context
}

// HTTPStatus 返回HTTP状态码
func (e *BaseError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// StackTrace 返回堆栈跟踪
func (e *BaseError) StackTrace() string {
	return e.stack
}

// Severity 返回错误严重程度
func (e *BaseError) Severity() ErrorSeverity {
	return e.severity
}

// AddContext 添加错误上下文
func (e *BaseError) AddContext(key string, value interface{}) {
	if e.context == nil {
		e.context = make(map[string]interface{})
	}
	e.context[key] = value
}

// Unwrap 实现errors.Unwrap接口
func (e *BaseError) Unwrap() error {
	return e.cause
}

// Is 实现errors.Is接口
func (e *BaseError) Is(target error) bool {
	if other, ok := target.(*BaseError); ok {
		return e.code == other.code
	}
	return false
}

// BusinessError 业务错误
type BusinessError struct {
	BaseError
	EntityType string
	EntityID   interface{}
}

func (e *BusinessError) HTTPStatus() int {
	if e.code == "NOT_FOUND" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// ValidationError 验证错误
type ValidationError struct {
	BaseError
	Field string
	Value interface{}
	Rules []string
}

func (e *ValidationError) HTTPStatus() int {
	return http.StatusUnprocessableEntity
}

// DatabaseError 数据库错误
type DatabaseError struct {
	BaseError
	Operation string
	Table     string
	Query     string
}

func (e *DatabaseError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// AuthorizationError 权限错误
type AuthorizationError struct {
	BaseError
	RequiredPermission string
	CurrentPermission  string
}

func (e *AuthorizationError) HTTPStatus() int {
	return http.StatusForbidden
}

// ConcurrencyError 并发错误
type ConcurrencyError struct {
	BaseError
	ResourceType string
	ResourceID   interface{}
	ConflictType string
}

func (e *ConcurrencyError) HTTPStatus() int {
	return http.StatusConflict
}

// NetworkError 网络错误
type NetworkError struct {
	BaseError
	Endpoint string
	Timeout  bool
}

func (e *NetworkError) HTTPStatus() int {
	if e.Timeout {
		return http.StatusRequestTimeout
	}
	return http.StatusBadGateway
}

// PanicError Panic错误
type PanicError struct {
	BaseError
	PanicValue interface{}
}

func (e *PanicError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// NewBusinessError 创建业务错误
func NewBusinessError(code, message string, cause error) *BusinessError {
	return &BusinessError{
		BaseError: BaseError{
			code:     code,
			message:  message,
			cause:    cause,
			severity: SeverityMedium,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
	}
}

// NewValidationError 创建验证错误
func NewValidationError(field, code, message string, details string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			code:     code,
			message:  message,
			details:  details,
			severity: SeverityLow,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
		Field: field,
	}
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(operation, message string, cause error) *DatabaseError {
	return &DatabaseError{
		BaseError: BaseError{
			code:     operation,
			message:  message,
			cause:    cause,
			severity: SeverityHigh,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
		Operation: operation,
	}
}

// NewAuthorizationError 创建权限错误
func NewAuthorizationError(code, message string, required, current string) *AuthorizationError {
	return &AuthorizationError{
		BaseError: BaseError{
			code:     code,
			message:  message,
			severity: SeverityHigh,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
		RequiredPermission: required,
		CurrentPermission:  current,
	}
}

// NewConcurrencyError 创建并发错误
func NewConcurrencyError(resourceType string, resourceID interface{}, conflictType string, cause error) *ConcurrencyError {
	return &ConcurrencyError{
		BaseError: BaseError{
			code:     "CONCURRENCY_ERROR",
			message:  fmt.Sprintf("Concurrency conflict on %s", resourceType),
			cause:    cause,
			severity: SeverityHigh,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ConflictType: conflictType,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(endpoint string, timeout bool, cause error) *NetworkError {
	code := "NETWORK_ERROR"
	if timeout {
		code = "TIMEOUT_ERROR"
	}

	return &NetworkError{
		BaseError: BaseError{
			code:     code,
			message:  fmt.Sprintf("Network error: %s", endpoint),
			cause:    cause,
			severity: SeverityMedium,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
		Endpoint: endpoint,
		Timeout:  timeout,
	}
}

// NewPanicError 创建Panic错误
func NewPanicError(value interface{}, stack string) *PanicError {
	return &PanicError{
		BaseError: BaseError{
			code:     "PANIC_ERROR",
			message:  fmt.Sprintf("Application panic: %v", value),
			severity: SeverityCritical,
			time:     time.Now(),
			stack:    stack,
		},
		PanicValue: value,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(code, message string, cause error) *BaseError {
	return &BaseError{
		code:     code,
		message:  message,
		cause:    cause,
		severity: SeverityCritical,
		time:     time.Now(),
		stack:    string(debug.Stack()),
	}
}

// NewNotFoundError 创建未找到错误
func NewNotFoundError(entity, message string, id interface{}) *BusinessError {
	return &BusinessError{
		BaseError: BaseError{
			code:     "NOT_FOUND",
			message:  message,
			severity: SeverityLow,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
		EntityType: entity,
		EntityID:   id,
	}
}

// NewConflictError 创建冲突错误
func NewConflictError(message string, cause error) *BusinessError {
	return &BusinessError{
		BaseError: BaseError{
			code:     "CONFLICT",
			message:  message,
			cause:    cause,
			severity: SeverityMedium,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(code, message string) *AuthorizationError {
	return &AuthorizationError{
		BaseError: BaseError{
			code:     code,
			message:  message,
			severity: SeverityHigh,
			time:     time.Now(),
			stack:    string(debug.Stack()),
		},
	}
}

// 错误类型检查函数
func IsBusinessError(err error) bool {
	var businessErr *BusinessError
	return errors.As(err, &businessErr)
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

func IsDatabaseError(err error) bool {
	var databaseErr *DatabaseError
	return errors.As(err, &databaseErr)
}

func IsAuthorizationError(err error) bool {
	var authErr *AuthorizationError
	return errors.As(err, &authErr)
}

func IsConcurrencyError(err error) bool {
	var concurrencyErr *ConcurrencyError
	return errors.As(err, &concurrencyErr)
}

func IsNetworkError(err error) bool {
	var networkErr *NetworkError
	return errors.As(err, &networkErr)
}

func IsPanicError(err error) bool {
	var panicErr *PanicError
	return errors.As(err, &panicErr)
}

// 获取错误代码
func GetErrorCode(err error) string {
	if appErr, ok := err.(AppError); ok {
		return appErr.Code()
	}
	return "INTERNAL_ERROR"
}

// 获取HTTP状态码
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(AppError); ok {
		return appErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// 添加错误上下文
func WithContext(err error, key string, value interface{}) error {
	if appErr, ok := err.(AppError); ok {
		appErr.AddContext(key, value)
		return appErr
	}
	return err
}

// 添加多个上下文
func WithContexts(err error, context map[string]interface{}) error {
	if appErr, ok := err.(AppError); ok {
		for k, v := range context {
			appErr.AddContext(k, v)
		}
		return appErr
	}
	return err
}
