package auth

import "errors"

// JWT认证相关错误定义
var (
	// 令牌相关错误
	ErrTokenMissing     = errors.New("authorization token is required")
	ErrTokenExpired     = errors.New("token has expired")
	ErrTokenInvalid     = errors.New("token is invalid")
	ErrTokenMalformed   = errors.New("token is malformed")
	ErrTokenTypeMismatch = errors.New("token type mismatch")
	ErrTokenBlacklisted = errors.New("token is blacklisted")

	// 声明相关错误
	ErrClaimsInvalid    = errors.New("invalid token claims")
	ErrIssuerInvalid    = errors.New("invalid token issuer")
	ErrAudienceInvalid  = errors.New("invalid token audience")
	ErrSubjectInvalid   = errors.New("invalid token subject")
	ErrTokenTooEarly    = errors.New("token is not yet valid")
	ErrTokenTooOld      = errors.New("token is too old")

	// 上下文验证错误
	ErrIPMismatch        = errors.New("IP address mismatch")
	ErrUserAgentMismatch = errors.New("user agent mismatch")
	ErrDeviceMismatch    = errors.New("device mismatch")
	ErrLocationMismatch  = errors.New("location mismatch")
	ErrTimeConstraint    = errors.New("time constraint violation")

	// 权限相关错误
	ErrPermissionDenied = errors.New("permission denied")
	ErrRoleRequired      = errors.New("required role not found")
	ErrAccessDenied      = errors.New("access denied")

	// 安全相关错误
	ErrNonceUsed      = errors.New("nonce has been used")
	ErrReplayAttack   = errors.New("possible replay attack detected")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// 密钥相关错误
	ErrKeyNotFound     = errors.New("key not found")
	ErrKeyExpired      = errors.New("key has expired")
	ErrKeyRevoked      = errors.New("key has been revoked")
	ErrKeyInvalid      = errors.New("key is invalid")
	ErrKeyGenerationFailed = errors.New("key generation failed")

	// 刷新令牌相关错误
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrRefreshTokenUsed   = errors.New("refresh token has already been used")
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")

	// 配置相关错误
	ErrConfigInvalid    = errors.New("configuration is invalid")
	ErrAlgorithmNotSupported = errors.New("signing algorithm is not supported")
	ErrTokenTooLong     = errors.New("token is too long")
	ErrTokenTooShort    = errors.New("token is too short")

	// 存储相关错误
	ErrStorageError     = errors.New("storage operation failed")
	ErrCacheError       = errors.New("cache operation failed")
	ErrDatabaseError    = errors.New("database operation failed")

	// 网络相关错误
	ErrNetworkError     = errors.New("network operation failed")
	ErrTimeoutError     = errors.New("operation timed out")
	ErrUnavailable      = errors.New("service unavailable")

	// 内部错误
	ErrInternalError    = errors.New("internal server error")
	ErrUnexpectedError  = errors.New("unexpected error occurred")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// ErrorCategory 错误类别
type ErrorCategory int

const (
	// ErrorCategoryValidation 验证错误类别
	ErrorCategoryValidation ErrorCategory = iota
	// ErrorCategorySecurity 安全错误类别
	ErrorCategorySecurity
	// ErrorCategoryConfiguration 配置错误类别
	ErrorCategoryConfiguration
	// ErrorCategoryInfrastructure 基础设施错误类别
	ErrorCategoryInfrastructure
	// ErrorCategoryBusiness 业务逻辑错误类别
	ErrorCategoryBusiness
)

// ErrorInfo 错误信息
type ErrorInfo struct {
	Error    error
	Category ErrorCategory
	Code     string
	Message  string
	Severity int // 1=低, 2=中, 3=高, 4=严重
	Retryable bool
}

// ErrorInfoMap 错误信息映射表
var ErrorInfoMap = map[error]ErrorInfo{
	ErrTokenMissing: {
		Error:    ErrTokenMissing,
		Category: ErrorCategoryValidation,
		Code:     "TOKEN_MISSING",
		Message:  "Authorization token is required",
		Severity: 2,
		Retryable: false,
	},
	ErrTokenExpired: {
		Error:    ErrTokenExpired,
		Category: ErrorCategorySecurity,
		Code:     "TOKEN_EXPIRED",
		Message:  "Token has expired",
		Severity: 2,
		Retryable: false,
	},
	ErrTokenInvalid: {
		Error:    ErrTokenInvalid,
		Category: ErrorCategorySecurity,
		Code:     "TOKEN_INVALID",
		Message:  "Token is invalid or has been tampered with",
		Severity: 3,
		Retryable: false,
	},
	ErrTokenMalformed: {
		Error:    ErrTokenMalformed,
		Category: ErrorCategoryValidation,
		Code:     "TOKEN_MALFORMED",
		Message:  "Token format is malformed",
		Severity: 2,
		Retryable: false,
	},
	ErrTokenTypeMismatch: {
		Error:    ErrTokenTypeMismatch,
		Category: ErrorCategoryValidation,
		Code:     "TOKEN_TYPE_MISMATCH",
		Message:  "Token type does not match expected type",
		Severity: 2,
		Retryable: false,
	},
	ErrTokenBlacklisted: {
		Error:    ErrTokenBlacklisted,
		Category: ErrorCategorySecurity,
		Code:     "TOKEN_BLACKLISTED",
		Message:  "Token has been revoked",
		Severity: 3,
		Retryable: false,
	},
	ErrClaimsInvalid: {
		Error:    ErrClaimsInvalid,
		Category: ErrorCategoryValidation,
		Code:     "CLAIMS_INVALID",
		Message:  "Token contains invalid claims",
		Severity: 2,
		Retryable: false,
	},
	ErrIPMismatch: {
		Error:    ErrIPMismatch,
		Category: ErrorCategorySecurity,
		Code:     "IP_MISMATCH",
		Message:  "Request IP address does not match authorized IP",
		Severity: 3,
		Retryable: false,
	},
	ErrUserAgentMismatch: {
		Error:    ErrUserAgentMismatch,
		Category: ErrorCategorySecurity,
		Code:     "USER_AGENT_MISMATCH",
		Message:  "Request user agent does not match authorized user agent",
		Severity: 2,
		Retryable: false,
	},
	ErrPermissionDenied: {
		Error:    ErrPermissionDenied,
		Category: ErrorCategoryBusiness,
		Code:     "PERMISSION_DENIED",
		Message:  "Insufficient permissions to access resource",
		Severity: 2,
		Retryable: false,
	},
	ErrRoleRequired: {
		Error:    ErrRoleRequired,
		Category: ErrorCategoryBusiness,
		Code:     "ROLE_REQUIRED",
		Message:  "Required role not found",
		Severity: 2,
		Retryable: false,
	},
	ErrNonceUsed: {
		Error:    ErrNonceUsed,
		Category: ErrorCategorySecurity,
		Code:     "NONCE_USED",
		Message:  "Nonce has already been used",
		Severity: 3,
		Retryable: false,
	},
	ErrReplayAttack: {
		Error:    ErrReplayAttack,
		Category: ErrorCategorySecurity,
		Code:     "REPLAY_ATTACK",
		Message:  "Possible replay attack detected",
		Severity: 4,
		Retryable: false,
	},
	ErrRateLimitExceeded: {
		Error:    ErrRateLimitExceeded,
		Category: ErrorCategorySecurity,
		Code:     "RATE_LIMIT_EXCEEDED",
		Message:  "Rate limit has been exceeded",
		Severity: 2,
		Retryable: true,
	},
	ErrKeyNotFound: {
		Error:    ErrKeyNotFound,
		Category: ErrorCategoryInfrastructure,
		Code:     "KEY_NOT_FOUND",
		Message:  "Cryptographic key not found",
		Severity: 3,
		Retryable: false,
	},
	ErrKeyExpired: {
		Error:    ErrKeyExpired,
		Category: ErrorCategorySecurity,
		Code:     "KEY_EXPIRED",
		Message:  "Cryptographic key has expired",
		Severity: 3,
		Retryable: false,
	},
	ErrKeyRevoked: {
		Error:    ErrKeyRevoked,
		Category: ErrorCategorySecurity,
		Code:     "KEY_REVOKED",
		Message:  "Cryptographic key has been revoked",
		Severity: 3,
		Retryable: false,
	},
	ErrRefreshTokenExpired: {
		Error:    ErrRefreshTokenExpired,
		Category: ErrorCategorySecurity,
		Code:     "REFRESH_TOKEN_EXPIRED",
		Message:  "Refresh token has expired",
		Severity: 2,
		Retryable: false,
	},
	ErrRefreshTokenUsed: {
		Error:    ErrRefreshTokenUsed,
		Category: ErrorCategorySecurity,
		Code:     "REFRESH_TOKEN_USED",
		Message:  "Refresh token has already been used",
		Severity: 3,
		Retryable: false,
	},
	ErrConfigInvalid: {
		Error:    ErrConfigInvalid,
		Category: ErrorCategoryConfiguration,
		Code:     "CONFIG_INVALID",
		Message:  "JWT configuration is invalid",
		Severity: 3,
		Retryable: false,
	},
	ErrAlgorithmNotSupported: {
		Error:    ErrAlgorithmNotSupported,
		Category: ErrorCategoryConfiguration,
		Code:     "ALGORITHM_NOT_SUPPORTED",
		Message:  "Signing algorithm is not supported",
		Severity: 3,
		Retryable: false,
	},
	ErrInternalError: {
		Error:    ErrInternalError,
		Category: ErrorCategoryInfrastructure,
		Code:     "INTERNAL_ERROR",
		Message:  "Internal server error occurred",
		Severity: 3,
		Retryable: true,
	},
}

// GetErrorInfo 获取错误信息
func GetErrorInfo(err error) ErrorInfo {
	if info, exists := ErrorInfoMap[err]; exists {
		return info
	}

	// 如果没有找到特定错误信息，返回默认信息
	return ErrorInfo{
		Error:    err,
		Category: ErrorCategoryInfrastructure,
		Code:     "UNKNOWN_ERROR",
		Message:  "An unknown error occurred",
		Severity: 2,
		Retryable: false,
	}
}

// IsSecurityError 判断是否为安全相关错误
func IsSecurityError(err error) bool {
	info := GetErrorInfo(err)
	return info.Category == ErrorCategorySecurity
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	info := GetErrorInfo(err)
	return info.Retryable
}

// GetErrorSeverity 获取错误严重程度
func GetErrorSeverity(err error) int {
	info := GetErrorInfo(err)
	return info.Severity
}