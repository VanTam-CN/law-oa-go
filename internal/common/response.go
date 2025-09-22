package common

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIError API错误结构
type APIError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// ResponseMeta 响应元数据
type ResponseMeta struct {
	Timestamp   time.Time `json:"timestamp"`
	RequestID   string    `json:"request_id,omitempty"`
	Version     string    `json:"version"`
	Server      string    `json:"server"`
	Environment string    `json:"environment"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// APIResponse 统一API响应结构（新格式）
type APIResponse struct {
	Success    bool            `json:"success"`
	Data       interface{}     `json:"data,omitempty"`
	Error      *APIError       `json:"error,omitempty"`
	Meta       ResponseMeta    `json:"meta"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// Response 旧版响应结构（保持向后兼容）
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse 旧版分页响应结构（保持向后兼容）
type PageResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// ResponseBuilder 响应构建器
type ResponseBuilder struct {
	version     string
	environment string
}

// NewResponseBuilder 创建响应构建器
func NewResponseBuilder(version, environment string) *ResponseBuilder {
	return &ResponseBuilder{
		version:     version,
		environment: environment,
	}
}

// createMeta 创建响应元数据
func (rb *ResponseBuilder) createMeta(c *gin.Context) ResponseMeta {
	meta := ResponseMeta{
		Timestamp:   time.Now(),
		Version:     rb.version,
		Server:      "law-oa-go",
		Environment: rb.environment,
	}

	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		meta.RequestID = requestID
	}

	return meta
}

// Success 成功响应
func (rb *ResponseBuilder) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    rb.createMeta(c),
	})
}

// SuccessWithPage 分页成功响应
func (rb *ResponseBuilder) SuccessWithPage(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	pagination := rb.calculatePagination(page, pageSize, total)
	c.JSON(http.StatusOK, APIResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Meta:       rb.createMeta(c),
	})
}

// Error 错误响应
func (rb *ResponseBuilder) Error(c *gin.Context, statusCode int, errCode, message string, details ...string) {
	apiErr := &APIError{
		Code:    errCode,
		Message: message,
	}

	if len(details) > 0 {
		apiErr.Details = details[0]
	}

	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   apiErr,
		Meta:    rb.createMeta(c),
	})
}

// ErrorWithContext 带上下文的错误响应
func (rb *ResponseBuilder) ErrorWithContext(c *gin.Context, statusCode int, errCode, message string, context map[string]interface{}) {
	apiErr := &APIError{
		Code:    errCode,
		Message: message,
		Context: context,
	}

	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   apiErr,
		Meta:    rb.createMeta(c),
	})
}

// ErrorWithSuggestions 带建议的错误响应
func (rb *ResponseBuilder) ErrorWithSuggestions(c *gin.Context, statusCode int, errCode, message string, suggestions []string) {
	apiErr := &APIError{
		Code:        errCode,
		Message:     message,
		Suggestions: suggestions,
	}

	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   apiErr,
		Meta:    rb.createMeta(c),
	})
}

// ValidationError 参数验证错误
func (rb *ResponseBuilder) ValidationError(c *gin.Context, message string, fieldErrors map[string]string) {
	rb.ErrorWithContext(c, http.StatusBadRequest, "VALIDATION_ERROR", message, map[string]interface{}{
		"field_errors": fieldErrors,
	})
}

// BadRequest 400错误
func (rb *ResponseBuilder) BadRequest(c *gin.Context, message string, details ...string) {
	rb.Error(c, http.StatusBadRequest, "BAD_REQUEST", message, details...)
}

// Unauthorized 401错误
func (rb *ResponseBuilder) Unauthorized(c *gin.Context, message string, details ...string) {
	rb.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, details...)
}

// Forbidden 403错误
func (rb *ResponseBuilder) Forbidden(c *gin.Context, message string, details ...string) {
	rb.Error(c, http.StatusForbidden, "FORBIDDEN", message, details...)
}

// NotFound 404错误
func (rb *ResponseBuilder) NotFound(c *gin.Context, message string, details ...string) {
	rb.Error(c, http.StatusNotFound, "NOT_FOUND", message, details...)
}

// InternalServerError 500错误
func (rb *ResponseBuilder) InternalServerError(c *gin.Context, message string, details ...string) {
	rb.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, details...)
}

// calculatePagination 计算分页信息
func (rb *ResponseBuilder) calculatePagination(page, pageSize int, total int64) *PaginationInfo {
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// 默认响应构建器
var DefaultResponseBuilder = NewResponseBuilder("v1", getEnvironment())

// 便捷函数 - 新版统一API
func APISuccess(c *gin.Context, data interface{}) {
	DefaultResponseBuilder.Success(c, data)
}

func APISuccessWithPage(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	DefaultResponseBuilder.SuccessWithPage(c, data, total, page, pageSize)
}

func NewAPIError(c *gin.Context, statusCode int, errCode, message string, details ...string) {
	DefaultResponseBuilder.Error(c, statusCode, errCode, message, details...)
}

func APIErrorWithContext(c *gin.Context, statusCode int, errCode, message string, context map[string]interface{}) {
	DefaultResponseBuilder.ErrorWithContext(c, statusCode, errCode, message, context)
}

func APIErrorWithSuggestions(c *gin.Context, statusCode int, errCode, message string, suggestions []string) {
	DefaultResponseBuilder.ErrorWithSuggestions(c, statusCode, errCode, message, suggestions)
}

func APIValidationError(c *gin.Context, message string, fieldErrors map[string]string) {
	DefaultResponseBuilder.ValidationError(c, message, fieldErrors)
}

func APIBadRequest(c *gin.Context, message string, details ...string) {
	DefaultResponseBuilder.BadRequest(c, message, details...)
}

func APIUnauthorized(c *gin.Context, message string, details ...string) {
	DefaultResponseBuilder.Unauthorized(c, message, details...)
}

func APIForbidden(c *gin.Context, message string, details ...string) {
	DefaultResponseBuilder.Forbidden(c, message, details...)
}

func APINotFound(c *gin.Context, message string, details ...string) {
	DefaultResponseBuilder.NotFound(c, message, details...)
}

func APIInternalServerError(c *gin.Context, message string, details ...string) {
	DefaultResponseBuilder.InternalServerError(c, message, details...)
}

// 便捷函数 - 旧版兼容（保持向后兼容）
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "操作成功",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

func SuccessWithPage(c *gin.Context, data interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, PageResponse{
		Code:    200,
		Message: "查询成功",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

func ValidationError(c *gin.Context, message string) {
	BadRequest(c, "参数验证失败: "+message)
}

// NewRequestBodyBuffer 创建请求体缓冲区
func NewRequestBodyBuffer(body []byte) *RequestBodyBuffer {
	return &RequestBodyBuffer{
		body:   body,
		offset: 0,
	}
}

// RequestBodyBuffer 请求体缓冲区
type RequestBodyBuffer struct {
	body   []byte
	offset int
}

// Read 实现io.Reader接口
func (r *RequestBodyBuffer) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.body) {
		return 0, io.EOF
	}
	n = copy(p, r.body[r.offset:])
	r.offset += n
	return n, nil
}

// Close 实现io.Closer接口
func (r *RequestBodyBuffer) Close() error {
	r.offset = 0
	return nil
}

// getEnvironment 获取当前环境
func getEnvironment() string {
	// 可以从环境变量或配置中获取
	return "development"
}
