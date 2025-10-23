package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenExtractor 令牌提取器接口
type TokenExtractor func(*gin.Context) (string, error)

// DefaultTokenExtractor 默认令牌提取器（从Authorization header提取）
func DefaultTokenExtractor(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", ErrTokenMissing
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return "", ErrTokenMalformed
	}

	return parts[1], nil
}

// CookieTokenExtractor Cookie令牌提取器
func CookieTokenExtractor(cookieName string) TokenExtractor {
	return func(c *gin.Context) (string, error) {
		cookie, err := c.Cookie(cookieName)
		if err != nil {
			return "", ErrTokenMissing
		}
		return cookie, nil
	}
}

// QueryTokenExtractor 查询参数令牌提取器
func QueryTokenExtractor(paramName string) TokenExtractor {
	return func(c *gin.Context) (string, error) {
		token := c.Query(paramName)
		if token == "" {
			return "", ErrTokenMissing
		}
		return token, nil
	}
}

// HeaderTokenExtractor Header令牌提取器
func HeaderTokenExtractor(headerName string) TokenExtractor {
	return func(c *gin.Context) (string, error) {
		token := c.GetHeader(headerName)
		if token == "" {
			return "", ErrTokenMissing
		}
		return token, nil
	}
}

// MultiExtractor 多源令牌提取器
func MultiExtractor(extractors ...TokenExtractor) TokenExtractor {
	return func(c *gin.Context) (string, error) {
		for _, extractor := range extractors {
			token, err := extractor(c)
			if err == nil {
				return token, nil
			}
		}
		return "", ErrTokenMissing
	}
}

// ErrorHandler 错误处理函数
type ErrorHandler func(*gin.Context, error)

// DefaultErrorHandler 默认错误处理器
func DefaultErrorHandler(c *gin.Context, err error) {
	switch err {
	case ErrTokenMissing:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authorization token required",
			"code":  "TOKEN_MISSING",
		})
	case ErrTokenExpired:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "token has expired",
			"code":  "TOKEN_EXPIRED",
		})
	case ErrTokenInvalid:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
			"code":  "TOKEN_INVALID",
		})
	case ErrTokenMalformed:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "malformed token",
			"code":  "TOKEN_MALFORMED",
		})
	case ErrTokenTypeMismatch:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "token type mismatch",
			"code":  "TOKEN_TYPE_MISMATCH",
		})
	case ErrTokenBlacklisted:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "token is blacklisted",
			"code":  "TOKEN_BLACKLISTED",
		})
	case ErrClaimsInvalid:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token claims",
			"code":  "CLAIMS_INVALID",
		})
	case ErrIPMismatch:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "IP address mismatch",
			"code":  "IP_MISMATCH",
		})
	case ErrUserAgentMismatch:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user agent mismatch",
			"code":  "USER_AGENT_MISMATCH",
		})
	case ErrPermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{
			"error": "permission denied",
			"code":  "PERMISSION_DENIED",
		})
	case ErrRoleRequired:
		c.JSON(http.StatusForbidden, gin.H{
			"error": "insufficient role",
			"code":  "ROLE_REQUIRED",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
			"code":  "INTERNAL_ERROR",
		})
	}
}

// DetailedErrorHandler 详细错误处理器
func DetailedErrorHandler(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorCode := "INTERNAL_ERROR"
	message := "internal server error"

	switch err {
	case ErrTokenMissing:
		statusCode = http.StatusUnauthorized
		errorCode = "TOKEN_MISSING"
		message = "Authorization token is required. Please provide a valid token in the Authorization header."
	case ErrTokenExpired:
		statusCode = http.StatusUnauthorized
		errorCode = "TOKEN_EXPIRED"
		message = "The provided token has expired. Please refresh your token or login again."
	case ErrTokenInvalid:
		statusCode = http.StatusUnauthorized
		errorCode = "TOKEN_INVALID"
		message = "The provided token is invalid or has been tampered with."
	case ErrTokenMalformed:
		statusCode = http.StatusBadRequest
		errorCode = "TOKEN_MALFORMED"
		message = "The token format is malformed. Expected format: 'Bearer <token>'."
	case ErrTokenTypeMismatch:
		statusCode = http.StatusUnauthorized
		errorCode = "TOKEN_TYPE_MISMATCH"
		message = "The provided token type does not match the expected type for this endpoint."
	case ErrTokenBlacklisted:
		statusCode = http.StatusUnauthorized
		errorCode = "TOKEN_BLACKLISTED"
		message = "The provided token has been revoked and is no longer valid."
	case ErrClaimsInvalid:
		statusCode = http.StatusUnauthorized
		errorCode = "CLAIMS_INVALID"
		message = "The token contains invalid claims."
	case ErrIPMismatch:
		statusCode = http.StatusUnauthorized
		errorCode = "IP_MISMATCH"
		message = "The request IP address does not match the token's authorized IP address."
	case ErrUserAgentMismatch:
		statusCode = http.StatusUnauthorized
		errorCode = "USER_AGENT_MISMATCH"
		message = "The request user agent does not match the token's authorized user agent."
	case ErrPermissionDenied:
		statusCode = http.StatusForbidden
		errorCode = "PERMISSION_DENIED"
		message = "You do not have permission to access this resource."
	case ErrRoleRequired:
		statusCode = http.StatusForbidden
		errorCode = "ROLE_REQUIRED"
		message = "This endpoint requires specific roles that you do not possess."
	}

	// 记录详细错误信息
	c.JSON(statusCode, gin.H{
		"error": message,
		"code":  errorCode,
		"details": map[string]interface{}{
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
			"timestamp": c.GetHeader("X-Request-Time"),
		},
	})
}

// JSONErrorHandler JSON格式错误处理器（API风格）
func JSONErrorHandler(c *gin.Context, err error) {
	response := gin.H{
		"success": false,
		"data":    nil,
		"error": gin.H{
			"code":    "UNKNOWN_ERROR",
			"message": "An unknown error occurred",
		},
		"meta": gin.H{
			"timestamp": c.GetHeader("X-Request-Time"),
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
		},
	}

	switch err {
	case ErrTokenMissing:
		response["error"] = gin.H{
			"code":    "TOKEN_MISSING",
			"message": "Authorization token is required",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrTokenExpired:
		response["error"] = gin.H{
			"code":    "TOKEN_EXPIRED",
			"message": "Token has expired",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrTokenInvalid:
		response["error"] = gin.H{
			"code":    "TOKEN_INVALID",
			"message": "Invalid token",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrTokenMalformed:
		response["error"] = gin.H{
			"code":    "TOKEN_MALFORMED",
			"message": "Malformed token format",
		}
		c.JSON(http.StatusBadRequest, response)
	case ErrTokenTypeMismatch:
		response["error"] = gin.H{
			"code":    "TOKEN_TYPE_MISMATCH",
			"message": "Token type mismatch",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrTokenBlacklisted:
		response["error"] = gin.H{
			"code":    "TOKEN_BLACKLISTED",
			"message": "Token has been revoked",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrClaimsInvalid:
		response["error"] = gin.H{
			"code":    "CLAIMS_INVALID",
			"message": "Invalid token claims",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrIPMismatch:
		response["error"] = gin.H{
			"code":    "IP_MISMATCH",
			"message": "IP address mismatch",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrUserAgentMismatch:
		response["error"] = gin.H{
			"code":    "USER_AGENT_MISMATCH",
			"message": "User agent mismatch",
		}
		c.JSON(http.StatusUnauthorized, response)
	case ErrPermissionDenied:
		response["error"] = gin.H{
			"code":    "PERMISSION_DENIED",
			"message": "Permission denied",
		}
		c.JSON(http.StatusForbidden, response)
	case ErrRoleRequired:
		response["error"] = gin.H{
			"code":    "ROLE_REQUIRED",
			"message": "Insufficient role privileges",
		}
		c.JSON(http.StatusForbidden, response)
	default:
		response["error"] = gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		}
		c.JSON(http.StatusInternalServerError, response)
	}
}

// CustomErrorHandler 自定义错误处理器
func CustomErrorHandler(errorHandler func(*gin.Context, error, map[string]interface{})) ErrorHandler {
	return func(c *gin.Context, err error) {
		context := map[string]interface{}{
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"ip":        c.ClientIP(),
			"userAgent": c.GetHeader("User-Agent"),
			"timestamp": c.GetHeader("X-Request-Time"),
		}
		errorHandler(c, err, context)
	}
}

// LoggingErrorHandler 记录日志的错误处理器
func LoggingErrorHandler(baseHandler ErrorHandler, logger interface{ Info(...interface{}) }) ErrorHandler {
	return func(c *gin.Context, err error) {
		logger.Info(fmt.Sprintf("JWT Auth Error: %v, Path: %s, IP: %s", err, c.Request.URL.Path, c.ClientIP()))
		baseHandler(c, err)
	}
}