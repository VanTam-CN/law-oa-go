package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// BaseHandler 基础处理器接口
type BaseHandler interface {
	HandleRequest(c *gin.Context) *APIResponse
	ValidateRequest(c *gin.Context) error
	GetTimeout() time.Duration
}

// BaseHandlerImpl 基础处理器实现
type BaseHandlerImpl struct{}

// HandleRequestWithContext 带上下文的请求处理
func (h *BaseHandlerImpl) HandleRequestWithContext(
	c *gin.Context,
	handler func(ctx context.Context) *APIResponse,
	defaultTimeout time.Duration,
) {
	timeout := defaultTimeout
	if h := GetTimeoutFromContext(c); h > 0 {
		timeout = h
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	// 将context注入到gin.Context中
	c.Set("request_context", ctx)

	// 执行处理函数
	response := handler(ctx)

	// 发送响应
	SendAPIResponse(c, response)
}

// ValidateID 验证ID参数
func (h *BaseHandlerImpl) ValidateID(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	return ParseUintParam(idStr, paramName)
}

// ParseUintParam 解析无符号整数参数
func ParseUintParam(value, paramName string) (uint, error) {
	if value == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}

	var id uint
	_, err := fmt.Sscanf(value, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("invalid %s format: %s", paramName, value)
	}

	if id == 0 {
		return 0, fmt.Errorf("%s cannot be zero", paramName)
	}

	return id, nil
}

// GetTimeoutFromContext 从上下文获取超时时间
func GetTimeoutFromContext(c *gin.Context) time.Duration {
	if timeout, exists := c.Get("handler_timeout"); exists {
		if t, ok := timeout.(time.Duration); ok {
			return t
		}
	}
	return 0
}

// SendAPIResponse 发送API响应
func SendAPIResponse(c *gin.Context, response *APIResponse) {
	statusCode := http.StatusOK
	if !response.Success && response.Error != nil {
		statusCode = GetStatusCodeFromError(response.Error.Code)
	}

	// 确保meta信息存在
	if response.Meta == nil {
		response.Meta = NewResponseMeta()
	}

	// 从context中获取request_id
	if requestID := c.GetString("request_id"); requestID != "" {
		response.Meta.RequestID = requestID
	}

	c.JSON(statusCode, response)
}

// GetStatusCodeFromError 根据错误代码获取HTTP状态码
func GetStatusCodeFromError(errorCode string) int {
	switch errorCode {
	case "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "AUTHENTICATION_ERROR":
		return http.StatusUnauthorized
	case "AUTHORIZATION_ERROR":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// DefaultTimeout 默认超时时间
const DefaultTimeout = 30 * time.Second
