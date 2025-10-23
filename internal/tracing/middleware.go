/**
 * OpenTelemetry追踪中间件 - 基于Gin框架的分布式追踪中间件
 * 自动追踪HTTP请求，记录请求/响应信息，关联上下游服务
 */

package tracing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MiddlewareConfig 追踪中间件配置
type MiddlewareConfig struct {
	// 跳过的路径
	SkipPaths []string `json:"skipPaths" yaml:"skipPaths"`

	// 是否记录请求体
	LogRequestBody bool `json:"logRequestBody" yaml:"logRequestBody"`

	// 是否记录响应体
	LogResponseBody bool `json:"logResponseBody" yaml:"logResponseBody"`

	// 最大请求体大小
	MaxRequestBodySize int64 `json:"maxRequestBodySize" yaml:"maxRequestBodySize"`

	// 最大响应体大小
	MaxResponseBodySize int64 `json:"maxResponseBodySize" yaml:"maxResponseBodySize"`

	// 是否记录用户代理
	LogUserAgent bool `json:"logUserAgent" yaml:"logUserAgent"`

	// 是否记录客户端IP
	LogClientIP bool `json:"logClientIP" yaml:"logClientIP"`

	// 是否添加自定义属性
	CustomAttributes func(*gin.Context) []attribute.KeyValue `json:"-"`

	// 慢请求阈值
	SlowRequestThreshold time.Duration `json:"slowRequestThreshold" yaml:"slowRequestThreshold"`

	// 生成Span名称的函数
	SpanNameFunc func(*gin.Context) string `json:"-"`
}

// DefaultMiddlewareConfig 返回默认中间件配置
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
			"/robots.txt",
		},
		LogRequestBody:      false,
		LogResponseBody:     false,
		MaxRequestBodySize:  1024 * 1024, // 1MB
		MaxResponseBodySize: 1024 * 1024, // 1MB
		LogUserAgent:        true,
		LogClientIP:         true,
		SlowRequestThreshold: time.Second,
		SpanNameFunc:        defaultSpanNameFunc,
	}
}

// defaultSpanNameFunc 默认的span名称生成函数
func defaultSpanNameFunc(c *gin.Context) string {
	// 使用 HTTP方法 + 路径作为span名称
	return c.Request.Method + " " + c.FullPath()
}

// responseBodyWriter 包装响应体写入器以捕获响应内容
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// TracingMiddleware 创建追踪中间件
func TracingMiddleware(serviceName string, config *MiddlewareConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}

	return func(c *gin.Context) {
		// 检查是否跳过
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// 生成span名称
		spanName := config.SpanNameFunc(c)
		if spanName == "" {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		// 开始span
		ctx, span := StartSpan(c.Request.Context(), spanName,
			trace.WithAttributes(
				HTTPMethod.String(c.Request.Method),
				HTTPURL.String(c.Request.URL.String()),
				HTTPTarget.String(c.Request.URL.Path),
				HTTPHost.String(c.Request.Host),
				HTTPScheme.String(c.Request.URL.Scheme),
			),
		)

		// 设置上下文
		c.Request = c.Request.WithContext(ctx)

		// 记录开始时间
		startTime := time.Now()

		// 添加追踪ID和span ID到上下文
		traceID := GetTraceID(span)
		spanID := GetSpanID(span)
		c.Set("trace_id", traceID)
		c.Set("span_id", spanID)

		// 添加到HTTP头以便下游服务获取
		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		// 捕获请求体
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(io.LimitReader(c.Request.Body, config.MaxRequestBodySize))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 捕获响应体
		var responseBody *bytes.Buffer
		if config.LogResponseBody {
			responseBody = &bytes.Buffer{}
			c.Writer = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           responseBody,
			}
		}

		// 添加请求属性
		if config.LogUserAgent && c.Request.UserAgent() != "" {
			AddSpanAttributes(span, HTTPUserAgent.String(c.Request.UserAgent()))
		}

		if config.LogClientIP {
			clientIP := c.ClientIP()
			if clientIP != "" {
				AddSpanAttributes(span, HTTPRemoteIP.String(clientIP))
			}
		}

		// 添加自定义属性
		if config.CustomAttributes != nil {
			customAttrs := config.CustomAttributes(c)
			if len(customAttrs) > 0 {
				AddSpanAttributes(span, customAttrs...)
			}
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)

		// 记录响应状态码
		statusCode := c.Writer.Status()
		AddSpanAttributes(span, HTTPStatusCode.Int(statusCode))

		// 记录响应大小
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			AddSpanAttributes(span, attribute.Int("http.response_content_length", responseSize))
		}

		// 记录请求体大小
		if len(requestBody) > 0 {
			AddSpanAttributes(span, attribute.Int("http.request_content_length", len(requestBody)))
			AddSpanAttributes(span, attribute.String("http.request_body", truncateString(string(requestBody), 1000)))
		}

		// 记录响应体
		if config.LogResponseBody && responseBody != nil && responseBody.Len() > 0 {
			responseBodyStr := responseBody.String()
			AddSpanAttributes(span, attribute.String("http.response_body", truncateString(responseBodyStr, 1000)))
		}

		// 记录处理时间
		AddSpanAttributes(span, attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6))

		// 记录慢请求
		if duration > config.SlowRequestThreshold {
			AddSpanEvents(span, "slow_request_detected",
				attribute.String("threshold", config.SlowRequestThreshold.String()),
				attribute.String("duration", duration.String()),
			)
		}

		// 设置span状态
		if c.Errors != nil && len(c.Errors) > 0 {
			// 有错误
			SetSpanStatus(span, trace.StatusCodeError, c.Errors.String())
			for _, err := c.Errors {
				RecordError(span, err.Err)
			}
		} else if statusCode >= 500 {
			// 服务器错误
			SetSpanStatus(span, trace.StatusCodeError, http.StatusText(statusCode))
		} else if statusCode >= 400 {
			// 客户端错误
			SetSpanStatus(span, trace.StatusCodeError, http.StatusText(statusCode))
		} else {
			// 成功
			SetSpanStatus(span, trace.StatusCodeOK, http.StatusText(statusCode))
		}

		// 添加业务属性
		if userID, exists := c.Get("user_id"); exists {
			AddSpanAttributes(span, attribute.String("user.id", toString(userID)))
		}

		if module, exists := c.Get("module"); exists {
			AddSpanAttributes(span, attribute.String("module", toString(module)))
		}

		if operation, exists := c.Get("operation"); exists {
			AddSpanAttributes(span, attribute.String("operation", toString(operation)))
		}

		// 结束span
		span.End()
	}
}

// shouldSkipPath 检查是否应该跳过路径
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// toString 将任意类型转换为字符串
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// ==============================
// 简化的中间件函数
// ==============================

// GinTracing 使用默认配置的Gin追踪中间件
func GinTracing(serviceName string) gin.HandlerFunc {
	return TracingMiddleware(serviceName, DefaultMiddlewareConfig())
}

// GinTracingWithConfig 使用自定义配置的Gin追踪中间件
func GinTracingWithConfig(serviceName string, config *MiddlewareConfig) gin.HandlerFunc {
	return TracingMiddleware(serviceName, config)
}

// ==============================
// OpenTelemetry Gin中间件封装
// ==============================

// OtelGinMiddleware 使用OpenTelemetry官方Gin中间件
func OtelGinMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// OtelGinMiddlewareWithSkip 使用带跳过路径的OpenTelemetry Gin中间件
func OtelGinMiddlewareWithSkip(serviceName string, skipPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过
		if shouldSkipPath(c.Request.URL.Path, skipPaths) {
			c.Next()
			return
		}

		// 使用官方中间件
		otelgin.Middleware(serviceName)(c)
	}
}

// ==============================
// 辅助函数
// ==============================

// GetTraceIDFromGinContext 从Gin上下文获取追踪ID
func GetTraceIDFromGinContext(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		return toString(traceID)
	}
	return GetTraceIDFromContext(c.Request.Context())
}

// GetSpanIDFromGinContext 从Gin上下文获取span ID
func GetSpanIDFromGinContext(c *gin.Context) string {
	if spanID, exists := c.Get("span_id"); exists {
		return toString(spanID)
	}
	return GetSpanIDFromContext(c.Request.Context())
}

// SetBusinessAttributes 设置业务属性到当前span
func SetBusinessAttributes(c *gin.Context, attrs map[string]interface{}) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		for k, v := range attrs {
			AddSpanAttributes(span, attribute.String(k, toString(v)))
		}
	}
}

// SetUserInfo 设置用户信息到当前span
func SetUserInfo(c *gin.Context, userID interface{}, userName string) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		AddSpanAttributes(span,
			attribute.String("user.id", toString(userID)),
			attribute.String("user.name", userName),
		)
	}
}

// SetOperationInfo 设置操作信息到当前span
func SetOperationInfo(c *gin.Context, module, operation string) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		AddSpanAttributes(span,
			attribute.String("module", module),
			attribute.String("operation", operation),
		)
	}
}

// AddErrorToSpan 添加错误到当前span
func AddErrorToSpan(c *gin.Context, err error) {
	span := trace.SpanFromContext(c.Request.Context())
	if span != nil {
		RecordError(span, err)
		SetSpanStatus(span, trace.StatusCodeError, err.Error())
	}
}