package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/logger"
)

// AuditMiddleware 审计中间件
func AuditMiddleware() gin.HandlerFunc {
	auditLogger := logger.NewAuditLogger()
	performanceLogger := logger.NewPerformanceLogger()

	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// 设置上下文信息
		c.Set("client_ip", c.ClientIP())
		c.Set("user_agent", c.GetHeader("User-Agent"))

		// 记录请求体（仅对POST/PUT/PATCH请求）
		var requestBody []byte
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if c.Request.Body != nil {
				requestBody, _ = io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
		}

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 记录API性能
		performanceLogger.LogAPIPerformance(
			c.Request.Context(),
			c.Request.Method,
			c.Request.URL.Path,
			duration,
			statusCode,
		)

		// 获取用户ID（如果已认证）
		var userID uint
		if uid, exists := c.Get("user_id"); exists {
			userID = uid.(uint)
		}

		// 记录审计日志
		auditDetails := map[string]interface{}{
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"query":        c.Request.URL.RawQuery,
			"status_code":  statusCode,
			"duration_ms":  duration.Milliseconds(),
			"client_ip":    c.ClientIP(),
			"user_agent":   c.GetHeader("User-Agent"),
			"content_type": c.GetHeader("Content-Type"),
		}

		// 添加请求体（敏感信息需要过滤）
		if len(requestBody) > 0 && len(requestBody) < 1024 { // 限制大小
			// 过滤敏感字段
			if !containsSensitiveData(c.Request.URL.Path) {
				auditDetails["request_body"] = string(requestBody)
			}
		}

		// 记录用户操作
		if userID > 0 {
			action := getActionFromRequest(c.Request.Method, c.Request.URL.Path)
			resource := getResourceFromPath(c.Request.URL.Path)

			auditLogger.LogUserAction(
				c.Request.Context(),
				userID,
				action,
				resource,
				auditDetails,
			)
		} else {
			// 未认证用户的操作
			auditLogger.LogSystemEvent(
				c.Request.Context(),
				"anonymous_request",
				"api",
				auditDetails,
			)
		}

		// 记录安全事件
		if shouldLogSecurityEvent(c.Request.URL.Path, statusCode) {
			severity := getSecurityEventSeverity(statusCode)
			auditLogger.LogSecurityEvent(
				c.Request.Context(),
				"security_event",
				severity,
				auditDetails,
			)
		}
	})
}

// containsSensitiveData 检查路径是否包含敏感数据
func containsSensitiveData(path string) bool {
	sensitivePaths := []string{
		"/auth/login",
		"/auth/register",
		"/users/change-password",
	}

	for _, sensitivePath := range sensitivePaths {
		if path == sensitivePath || path == "/api/v1"+sensitivePath {
			return true
		}
	}
	return false
}

// getActionFromRequest 从请求中获取操作类型
func getActionFromRequest(method, path string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

// getResourceFromPath 从路径中获取资源类型
func getResourceFromPath(path string) string {
	if path == "" {
		return "unknown"
	}

	// 简单的路径解析
	if path == "/api/v1/auth/login" {
		return "auth"
	}
	if path == "/api/v1/auth/register" {
		return "auth"
	}
	if path == "/api/v1/users" || path == "/api/v1/admin/users" {
		return "user"
	}
	if path == "/api/v1/clients" {
		return "client"
	}
	if path == "/api/v1/cases" {
		return "case"
	}

	// 默认返回路径的第一部分
	parts := splitPath(path)
	if len(parts) >= 3 {
		return parts[2] // /api/v1/resource
	}
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

// splitPath 分割路径
func splitPath(path string) []string {
	var parts []string
	current := ""

	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// shouldLogSecurityEvent 判断是否应该记录安全事件
func shouldLogSecurityEvent(path string, statusCode int) bool {
	// 认证相关的失败
	if (path == "/api/v1/auth/login" || path == "/api/v1/auth/register") && statusCode >= 400 {
		return true
	}

	// 权限相关的错误
	if statusCode == 401 || statusCode == 403 {
		return true
	}

	// 服务器错误
	if statusCode >= 500 {
		return true
	}

	return false
}

// getSecurityEventSeverity 获取安全事件严重程度
func getSecurityEventSeverity(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "high"
	case statusCode == 401 || statusCode == 403:
		return "medium"
	case statusCode >= 400:
		return "low"
	default:
		return "info"
	}
}

// DataMaskingMiddleware 数据脱敏中间件
func DataMaskingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// 在响应中脱敏敏感数据
		// 这里可以根据需要实现响应数据的脱敏逻辑
	})
}
