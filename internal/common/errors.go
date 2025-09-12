package common

import (
	"errors"
	"fmt"
)

// 定义具体的业务错误类型
var (
	// 用户相关错误
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrEmailExists     = errors.New("email already exists")
	ErrWeakPassword    = errors.New("password too weak")
	ErrInvalidRole     = errors.New("invalid role")

	// 客户相关错误
	ErrClientNotFound = errors.New("client not found")
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrInvalidPhone   = errors.New("invalid phone format")

	// 案件相关错误
	ErrCaseNotFound      = errors.New("case not found")
	ErrLawyerNotFound    = errors.New("lawyer not found")
	ErrInvalidCaseStatus = errors.New("invalid case status")

	// 通用错误
	ErrRecordNotFound   = errors.New("record not found")
	ErrDuplicateKey     = errors.New("duplicate key violation")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInternalServer   = errors.New("internal server error")
	ErrValidationFailed = errors.New("validation failed")
	ErrDatabaseError    = errors.New("database error")
	ErrCacheError       = errors.New("cache error")
	ErrExternalService  = errors.New("external service error")
)

// BusinessError 业务错误结构
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BusinessError) Unwrap() error {
	return e.Err
}

// 错误构造函数
func NewBusinessError(code, message string, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewValidationError(message, details string) *BusinessError {
	return &BusinessError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Details: details,
		Err:     ErrValidationFailed,
	}
}

func NewNotFoundError(resource string) *BusinessError {
	return &BusinessError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Err:     ErrRecordNotFound,
	}
}

func NewUnauthorizedError(message string) *BusinessError {
	return &BusinessError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Err:     ErrUnauthorized,
	}
}

func NewForbiddenError(message string) *BusinessError {
	return &BusinessError{
		Code:    "FORBIDDEN",
		Message: message,
		Err:     ErrForbidden,
	}
}

func NewDatabaseError(operation string, err error) *BusinessError {
	return &BusinessError{
		Code:    "DATABASE_ERROR",
		Message: fmt.Sprintf("Database operation failed: %s", operation),
		Err:     fmt.Errorf("%w: %w", ErrDatabaseError, err),
	}
}

func NewInternalError(message string, err error) *BusinessError {
	return &BusinessError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Err:     fmt.Errorf("%w: %w", ErrInternalServer, err),
	}
}

// 错误类型检查函数
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrRecordNotFound) ||
		errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrClientNotFound) ||
		errors.Is(err, ErrCaseNotFound) ||
		errors.Is(err, ErrLawyerNotFound)
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidPassword) ||
		errors.Is(err, ErrWeakPassword) ||
		errors.Is(err, ErrInvalidEmail) ||
		errors.Is(err, ErrInvalidPhone) ||
		errors.Is(err, ErrInvalidRole) ||
		errors.Is(err, ErrInvalidCaseStatus)
}

func IsConflictError(err error) bool {
	return errors.Is(err, ErrEmailExists) ||
		errors.Is(err, ErrDuplicateKey)
}

func IsUnauthorizedError(err error) bool {
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInvalidPassword)
}

func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden)
}

func IsDatabaseError(err error) bool {
	return errors.Is(err, ErrDatabaseError)
}

// 从错误中提取业务错误
func ExtractBusinessError(err error) (*BusinessError, bool) {
	var bizErr *BusinessError
	if errors.As(err, &bizErr) {
		return bizErr, true
	}
	return nil, false
}
