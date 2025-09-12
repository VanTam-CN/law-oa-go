package common

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "操作成功",
		Data:    data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// SuccessWithPage 分页成功响应
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

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError 500错误
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// ValidationError 参数验证错误
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