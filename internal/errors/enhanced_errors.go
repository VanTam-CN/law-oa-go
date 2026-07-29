package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorCategory 错误分类枚举
type ErrorCategory int

const (
	ErrorCategoryValidation ErrorCategory = iota
	ErrorCategoryDatabase
	ErrorCategoryNetwork
	ErrorCategoryBusiness
	ErrorCategorySystem
	ErrorCategorySecurity
)

func (ec ErrorCategory) String() string {
	switch ec {
	case ErrorCategoryValidation:
		return "validation"
	case ErrorCategoryDatabase:
		return "database"
	case ErrorCategoryNetwork:
		return "network"
	case ErrorCategoryBusiness:
		return "business"
	case ErrorCategorySystem:
		return "system"
	case ErrorCategorySecurity:
		return "security"
	default:
		return "unknown"
	}
}

// ErrorLevel 错误级别枚举
type ErrorLevel int

const (
	ErrorLevelInfo ErrorLevel = iota
	ErrorLevelWarning
	ErrorLevelError
	ErrorLevelCritical
)

func (el ErrorLevel) String() string {
	switch el {
	case ErrorLevelInfo:
		return "INFO"
	case ErrorLevelWarning:
		return "WARNING"
	case ErrorLevelError:
		return "ERROR"
	case ErrorLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// EnhancedError 增强的错误类型
type EnhancedError struct {
	// 基础信息
	id       string
	code     string
	category ErrorCategory
	level    ErrorLevel
	message  string
	details  string

	// 上下文信息
	context   map[string]interface{}
	timestamp time.Time
	traceID   string

	// 根本原因
	cause   error   `json:"-"`
	wrapped []error `json:"-"`

	// 调试信息
	stackTrace string
	file       string
	line       int
	function   string

	// HTTP响应相关
	httpStatus int `json:"-"`

	// 恢复策略
	suggestions []string
}

// Error 实现error接口
func (e *EnhancedError) Error() string {
	message := e.message
	if e.details != "" {
		message = fmt.Sprintf("%s: %s", message, e.details)
	}
	if e.cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.level, message, e.cause)
	}
	return fmt.Sprintf("[%s] %s", e.level, message)
}

// Code 返回错误代码
func (e *EnhancedError) Code() string {
	return e.code
}

// Message 返回错误消息
func (e *EnhancedError) Message() string {
	return e.message
}

// Details 返回错误详情
func (e *EnhancedError) Details() string {
	return e.details
}

// Context 返回错误上下文
func (e *EnhancedError) Context() map[string]interface{} {
	return e.context
}

// Category 返回错误分类
func (e *EnhancedError) Category() ErrorCategory {
	return e.category
}

// Level 返回错误级别
func (e *EnhancedError) Level() ErrorLevel {
	return e.level
}

// HTTPStatus 返回HTTP状态码
func (e *EnhancedError) HTTPStatus() int {
	if e.httpStatus != 0 {
		return e.httpStatus
	}

	// 根据错误级别和分类返回默认HTTP状态码
	switch e.Level() {
	case ErrorLevelCritical:
		return 500
	case ErrorLevelError:
		switch e.Category() {
		case ErrorCategoryValidation:
			return 422
		case ErrorCategorySecurity:
			return 403
		case ErrorCategoryBusiness:
			return 400
		default:
			return 500
		}
	default:
		return 400
	}
}

// StackTrace 返回堆栈跟踪
func (e *EnhancedError) StackTrace() string {
	return e.stackTrace
}

// Unwrap 实现errors.Unwrap接口
func (e *EnhancedError) Unwrap() error {
	return e.cause
}

// Is 实现errors.Is接口
func (e *EnhancedError) Is(target error) bool {
	if other, ok := target.(*EnhancedError); ok {
		return e.code == other.Code() && e.category == other.category
	}
	return errors.Is(e.cause, target)
}

// AddContext 添加错误上下文
func (e *EnhancedError) AddContext(key string, value interface{}) {
	if e.context == nil {
		e.context = make(map[string]interface{})
	}
	e.context[key] = value
}

// AddContexts 批量添加上下文
func (e *EnhancedError) AddContexts(context map[string]interface{}) {
	if e.context == nil {
		e.context = make(map[string]interface{})
	}
	for k, v := range context {
		e.context[k] = v
	}
}

// Wrap 包装另一个错误
func (e *EnhancedError) Wrap(err error) *EnhancedError {
	e.wrapped = append(e.wrapped, err)
	e.cause = err
	return e
}

// SetTraceID 设置跟踪ID
func (e *EnhancedError) SetTraceID(traceID string) {
	e.traceID = traceID
	e.AddContext("trace_id", traceID)
}

// TraceID 返回跟踪ID
func (e *EnhancedError) TraceID() string {
	return e.traceID
}

// SetHTTPStatus 设置HTTP状态码
func (e *EnhancedError) SetHTTPStatus(status int) {
	e.httpStatus = status
}

// SetSuggestions 设置错误建议
func (e *EnhancedError) SetSuggestions(suggestions []string) {
	e.suggestions = suggestions
}

// Timestamp 返回错误时间戳
func (e *EnhancedError) Timestamp() time.Time {
	return e.timestamp
}

// Suggestions 返回错误建议
func (e *EnhancedError) Suggestions() []string {
	return e.suggestions
}

// File 返回源码文件名
func (e *EnhancedError) File() string {
	return e.file
}

// Line 返回源码行号
func (e *EnhancedError) Line() int {
	return e.line
}

// Function 返回源码函数名
func (e *EnhancedError) Function() string {
	return e.function
}

// ID 返回错误ID
func (e *EnhancedError) ID() string {
	return e.id
}

// ToJSON 将错误转换为JSON
func (e *EnhancedError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// String 返回简短的字符串表示
func (e *EnhancedError) String() string {
	return fmt.Sprintf("Error[%s:%s]: %s", e.category, e.code, e.message)
}

// 错误构建器
type ErrorBuilder struct {
	error *EnhancedError
}

// NewError 创建新的错误构建器
func NewError(code, message string) *ErrorBuilder {
	return &ErrorBuilder{
		error: &EnhancedError{
			id:        generateErrorID(),
			code:      code,
			message:   message,
			timestamp: time.Now(),
			context:   make(map[string]interface{}),
		},
	}
}

// Category 设置错误分类
func (eb *ErrorBuilder) Category(category ErrorCategory) *ErrorBuilder {
	eb.error.category = category
	return eb
}

// Level 设置错误级别
func (eb *ErrorBuilder) Level(level ErrorLevel) *ErrorBuilder {
	eb.error.level = level
	return eb
}

// Details 设置错误详情
func (eb *ErrorBuilder) Details(details string) *ErrorBuilder {
	eb.error.details = details
	return eb
}

// Context 添加上下文
func (eb *ErrorBuilder) Context(key string, value interface{}) *ErrorBuilder {
	eb.error.AddContext(key, value)
	return eb
}

// Contexts 批量添加上下文
func (eb *ErrorBuilder) Contexts(context map[string]interface{}) *ErrorBuilder {
	eb.error.AddContexts(context)
	return eb
}

// Cause 设置根本原因
func (eb *ErrorBuilder) Cause(err error) *ErrorBuilder {
	eb.error.cause = err
	eb.error.wrapped = append(eb.error.wrapped, err)
	return eb
}

// HTTPStatus 设置HTTP状态码
func (eb *ErrorBuilder) HTTPStatus(status int) *ErrorBuilder {
	eb.error.httpStatus = status
	return eb
}

// Suggestions 设置错误建议
func (eb *ErrorBuilder) Suggestions(suggestions ...string) *ErrorBuilder {
	eb.error.suggestions = suggestions
	return eb
}

// Stack 启用堆栈跟踪
func (eb *ErrorBuilder) Stack() *ErrorBuilder {
	if pc, file, line, ok := runtime.Caller(1); ok {
		eb.error.stackTrace = getStackTrace()
		eb.error.file = file
		eb.error.line = line
		eb.error.function = runtime.FuncForPC(pc).Name()
	}
	return eb
}

// TraceID 设置跟踪ID
func (eb *ErrorBuilder) TraceID(traceID string) *ErrorBuilder {
	eb.error.SetTraceID(traceID)
	return eb
}

// Build 构建最终错误
func (eb *ErrorBuilder) Build() *EnhancedError {
	return eb.error
}

// 预定义错误创建函数

// ValidationError 创建验证错误
func ValidationError(field, message string) *EnhancedError {
	return NewError("VALIDATION_ERROR", message).
		Category(ErrorCategoryValidation).
		Level(ErrorLevelWarning).
		Context("field", field).
		Stack().
		Build()
}

// ValidationErrorWithDetails 创建带详情的验证错误
func ValidationErrorWithDetails(field, message, details string, rules []string) *EnhancedError {
	return NewError("VALIDATION_ERROR", message).
		Category(ErrorCategoryValidation).
		Level(ErrorLevelWarning).
		Details(details).
		Context("field", field).
		Context("rules", rules).
		Stack().
		Build()
}

// DatabaseError 创建数据库错误
func DatabaseError(operation, message string, cause error) *EnhancedError {
	return NewError("DATABASE_ERROR", message).
		Category(ErrorCategoryDatabase).
		Level(ErrorLevelError).
		Cause(cause).
		Context("operation", operation).
		Stack().
		Build()
}

// NetworkError 创建网络错误
func NetworkError(endpoint string, timeout bool, cause error) *EnhancedError {
	code := "NETWORK_ERROR"
	level := ErrorLevelWarning
	if timeout {
		code = "TIMEOUT_ERROR"
		level = ErrorLevelError
	}

	return NewError(code, fmt.Sprintf("Network error: %s", endpoint)).
		Category(ErrorCategoryNetwork).
		Level(level).
		Cause(cause).
		Context("endpoint", endpoint).
		Context("timeout", timeout).
		Stack().
		Build()
}

// BusinessError 创建业务错误
func BusinessError(entity, operation, message string) *EnhancedError {
	return NewError("BUSINESS_ERROR", message).
		Category(ErrorCategoryBusiness).
		Level(ErrorLevelWarning).
		Context("entity", entity).
		Context("operation", operation).
		Stack().
		Build()
}

// SystemError 创建系统错误
func SystemError(component, message string, cause error) *EnhancedError {
	return NewError("SYSTEM_ERROR", message).
		Category(ErrorCategorySystem).
		Level(ErrorLevelCritical).
		Cause(cause).
		Context("component", component).
		Stack().
		Build()
}

// SecurityError 创建安全错误
func SecurityError(action, reason string, cause error) *EnhancedError {
	return NewError("SECURITY_ERROR", reason).
		Category(ErrorCategorySecurity).
		Level(ErrorLevelError).
		Cause(cause).
		Context("action", action).
		Stack().
		Build()
}

// NotFoundError 创建未找到错误
func NotFoundError(entity, message string, id interface{}) *EnhancedError {
	return NewError("NOT_FOUND", message).
		Category(ErrorCategoryBusiness).
		Level(ErrorLevelWarning).
		Context("entity", entity).
		Context("id", id).
		Stack().
		Build()
}

// UnauthorizedError 创建未授权错误
func UnauthorizedError(action, resource string) *EnhancedError {
	return NewError("UNAUTHORIZED", fmt.Sprintf("Unauthorized to %s %s", action, resource)).
		Category(ErrorCategorySecurity).
		Level(ErrorLevelError).
		Context("action", action).
		Context("resource", resource).
		Stack().
		Build()
}

// ConflictError 创建冲突错误
func ConflictError(resource, reason string, cause error) *EnhancedError {
	return NewError("CONFLICT", reason).
		Category(ErrorCategoryBusiness).
		Level(ErrorLevelWarning).
		Cause(cause).
		Context("resource", resource).
		Stack().
		Build()
}

// InternalError 创建内部错误
func InternalError(message string, cause error) *EnhancedError {
	return NewError("INTERNAL_ERROR", message).
		Category(ErrorCategorySystem).
		Level(ErrorLevelCritical).
		Cause(cause).
		Stack().
		Build()
}

// 辅助函数

// generateErrorID 生成唯一错误ID
func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// 错误类型检查函数

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.category == ErrorCategoryValidation
	}
	return false
}

// IsDatabaseError 检查是否为数据库错误
func IsDatabaseError(err error) bool {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.category == ErrorCategoryDatabase
	}
	return false
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.category == ErrorCategoryNetwork
	}
	return false
}

// IsBusinessError 检查是否为业务错误
func IsBusinessError(err error) bool {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.category == ErrorCategoryBusiness
	}
	return false
}

// IsSystemError 检查是否为系统错误
func IsSystemError(err error) bool {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.category == ErrorCategorySystem
	}
	return false
}

// IsSecurityError 检查是否为安全错误
func IsSecurityError(err error) bool {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.category == ErrorCategorySecurity
	}
	return false
}

// IsCritical 检查是否为关键错误
func IsCritical(err error) bool {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.level == ErrorLevelCritical
	}
	return false
}

// GetErrorCode 获取错误代码
func GetErrorCode(err error) string {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.Code()
	}
	return "UNKNOWN_ERROR"
}

// GetErrorCategory 获取错误分类
func GetErrorCategory(err error) ErrorCategory {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.category
	}
	return ErrorCategorySystem
}

// GetErrorLevel 获取错误级别
func GetErrorLevel(err error) ErrorLevel {
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return enhancedErr.level
	}
	return ErrorLevelError
}

// WrapError 包装错误，添加上下文
func WrapError(err error, message string) *EnhancedError {
	if err == nil {
		return nil
	}

	// 如果已经是EnhancedError，添加上下文
	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		return NewError(enhancedErr.Code(), message).
			Category(enhancedErr.category).
			Level(enhancedErr.level).
			Cause(err).
			Stack().
			Build()
	}

	// 包装为系统错误
	return InternalError(message, err)
}

// JoinErrors 组合多个错误
func JoinErrors(errs ...error) *EnhancedError {
	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return WrapError(errs[0], "single error")
	}

	messages := make([]string, len(errs))
	for i, err := range errs {
		if err != nil {
			messages[i] = err.Error()
		}
	}

	return NewError("MULTIPLE_ERRORS", strings.Join(messages, "; ")).
		Category(ErrorCategorySystem).
		Level(ErrorLevelError).
		Cause(errors.Join(errs...)).
		Stack().
		Build()
}

// 错误恢复函数

// RecoverAndLog 从panic中恢复并记录日志
func RecoverAndLog(logger interface{}) error {
	if err := recover(); err != nil {
		stack := getStackTrace()
		return NewError("PANIC_RECOVERED", fmt.Sprintf("Application panic recovered: %v", err)).
			Category(ErrorCategorySystem).
			Level(ErrorLevelCritical).
			Context("panic_value", err).
			Details(stack).
			Build()
	}
	return nil
}

// WithContext 为错误添加上下文
func WithContext(err error, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		enhancedErr.AddContexts(context)
		return enhancedErr
	}

	wrappedErr := WrapError(err, "context added")
	wrappedErr.AddContexts(context)
	return wrappedErr
}

// WithTraceID 为错误添加跟踪ID
func WithTraceID(err error, traceID string) error {
	if err == nil {
		return nil
	}

	var enhancedErr *EnhancedError
	if errors.As(err, &enhancedErr) {
		enhancedErr.SetTraceID(traceID)
		return enhancedErr
	}

	wrappedErr := WrapError(err, "trace ID added")
	wrappedErr.SetTraceID(traceID)
	return wrappedErr
}
