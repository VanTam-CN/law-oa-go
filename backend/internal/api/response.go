package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一API响应结构
type APIResponse struct {
	Success bool          `json:"success"`
	Data    interface{}   `json:"data,omitempty"`
	Error   *ErrorDetail  `json:"error,omitempty"`
	Meta    *ResponseMeta `json:"meta,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Success    bool            `json:"success"`
	Data       interface{}     `json:"data"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
	Error      *ErrorDetail    `json:"error,omitempty"`
	Meta       *ResponseMeta   `json:"meta,omitempty"`
}

// ErrorDetail 错误详情结构
type ErrorDetail struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
}

// ResponseMeta 响应元信息
type ResponseMeta struct {
	Timestamp  time.Time `json:"timestamp"`
	RequestID  string    `json:"request_id"`
	Version    string    `json:"version"`
	ServerInfo string    `json:"server_info,omitempty"`
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

// ListRequest 通用列表查询请求
type ListRequest struct {
	Page      int                    `form:"page" binding:"min=1"`
	PageSize  int                    `form:"page_size" binding:"min=1,max=100"`
	SortBy    string                 `form:"sort_by"`
	SortOrder string                 `form:"sort_order" binding:"oneof=asc,desc"`
	Search    string                 `form:"search"`
	Filter    map[string]interface{} `form:"-"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
		Meta:    NewResponseMeta(),
	}
}

// NewPageResponse 创建分页响应
func NewPageResponse(data interface{}, pagination *PaginationInfo) *PageResponse {
	return &PageResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Meta:       NewResponseMeta(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code, message, details string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: NewResponseMeta(),
	}
}

// NewResponseMeta 创建响应元信息
func NewResponseMeta() *ResponseMeta {
	return &ResponseMeta{
		Timestamp:  time.Now(),
		RequestID:  GenerateRequestID(),
		Version:    "1.0.0",
		ServerInfo: "law-oa-go-server",
	}
}

// GenerateRequestID 生成请求ID
func GenerateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + GenerateRandomString(8)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	// 简化版本，实际应该使用更安全的随机数生成器
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// CalculatePagination 计算分页信息
func CalculatePagination(page, pageSize int, total int64) *PaginationInfo {
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

// ValidateListRequest 验证列表请求参数
func ValidateListRequest(req *ListRequest) error {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}
	return nil
}

// ResponseBuilder 响应构建器
type ResponseBuilder struct {
	version    string
	apiVersion string
}

// NewResponseBuilder 创建响应构建器
func NewResponseBuilder(version, apiVersion string) *ResponseBuilder {
	return &ResponseBuilder{
		version:    version,
		apiVersion: apiVersion,
	}
}

// Success 成功响应
func (rb *ResponseBuilder) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    rb.createMeta(c),
	})
}

// SuccessWithMessage 带消息的成功响应
func (rb *ResponseBuilder) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta:    rb.createMeta(c),
	})
}

// SuccessWithPage 分页成功响应
func (rb *ResponseBuilder) SuccessWithPage(c *gin.Context, data interface{}, total, page, pageSize int) {
	pagination := CalculatePagination(page, pageSize, int64(total))
	c.JSON(http.StatusOK, PageResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Meta:       rb.createMeta(c),
	})
}

// Error 错误响应
func (rb *ResponseBuilder) Error(c *gin.Context, statusCode int, errorCode, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
		Meta: rb.createMeta(c),
	})
}

// ErrorWithDetails 带详细信息的错误响应
func (rb *ResponseBuilder) ErrorWithDetails(c *gin.Context, statusCode int, errorCode, message, details string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Meta: rb.createMeta(c),
	})
}

// ErrorWithSuggestions 带建议的错误响应
func (rb *ResponseBuilder) ErrorWithSuggestions(c *gin.Context, statusCode int, errorCode, message string, suggestions []string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:        errorCode,
			Message:     message,
			Suggestions: suggestions,
		},
		Meta: rb.createMeta(c),
	})
}

// ValidationError 参数验证错误
func (rb *ResponseBuilder) ValidationError(c *gin.Context, message string, fieldErrors map[string]string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Context: map[string]interface{}{
				"field_errors": fieldErrors,
			},
		},
		Meta: rb.createMeta(c),
	})
}

// BadRequest 400错误
func (rb *ResponseBuilder) BadRequest(c *gin.Context, message string) {
	rb.Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized 401错误
func (rb *ResponseBuilder) Unauthorized(c *gin.Context, message string) {
	rb.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden 403错误
func (rb *ResponseBuilder) Forbidden(c *gin.Context, message string) {
	rb.Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound 404错误
func (rb *ResponseBuilder) NotFound(c *gin.Context, message string) {
	rb.Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalServerError 500错误
func (rb *ResponseBuilder) InternalServerError(c *gin.Context, message string) {
	rb.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// createMeta 创建响应元数据
func (rb *ResponseBuilder) createMeta(c *gin.Context) *ResponseMeta {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = GenerateRequestID()
	}

	return &ResponseMeta{
		Timestamp:  time.Now(),
		RequestID:  requestID,
		Version:    rb.version,
		ServerInfo: "law-oa-go-server",
	}
}

// 默认响应构建器实例
var DefaultResponseBuilder = NewResponseBuilder("1.0.0", "v1")

// 便捷函数
func Success(c *gin.Context, data interface{}) {
	DefaultResponseBuilder.Success(c, data)
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	DefaultResponseBuilder.SuccessWithMessage(c, message, data)
}

func SuccessWithPage(c *gin.Context, data interface{}, total, page, pageSize int) {
	DefaultResponseBuilder.SuccessWithPage(c, data, total, page, pageSize)
}

func Error(c *gin.Context, statusCode int, errorCode, message string) {
	DefaultResponseBuilder.Error(c, statusCode, errorCode, message)
}

func ErrorWithDetails(c *gin.Context, statusCode int, errorCode, message, details string) {
	DefaultResponseBuilder.ErrorWithDetails(c, statusCode, errorCode, message, details)
}

func ErrorWithSuggestions(c *gin.Context, statusCode int, errorCode, message string, suggestions []string) {
	DefaultResponseBuilder.ErrorWithSuggestions(c, statusCode, errorCode, message, suggestions)
}

func ValidationError(c *gin.Context, message string, fieldErrors map[string]string) {
	DefaultResponseBuilder.ValidationError(c, message, fieldErrors)
}

func BadRequest(c *gin.Context, message string) {
	DefaultResponseBuilder.BadRequest(c, message)
}

func Unauthorized(c *gin.Context, message string) {
	DefaultResponseBuilder.Unauthorized(c, message)
}

func Forbidden(c *gin.Context, message string) {
	DefaultResponseBuilder.Forbidden(c, message)
}

func NotFound(c *gin.Context, message string) {
	DefaultResponseBuilder.NotFound(c, message)
}

func InternalServerError(c *gin.Context, message string) {
	DefaultResponseBuilder.InternalServerError(c, message)
}

// PageToAPIResponse 将PageResponse转换为APIResponse
func PageToAPIResponse(pageResp *PageResponse) *APIResponse {
	return &APIResponse{
		Success: pageResp.Success,
		Data:    pageResp.Data,
		Error:   pageResp.Error,
		Meta:    pageResp.Meta,
	}
}
