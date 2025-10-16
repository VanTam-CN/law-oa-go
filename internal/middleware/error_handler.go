package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/errors"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success   bool         `json:"success"`
	Error     *ErrorDetail `json:"error,omitempty"`
	RequestID string       `json:"request_id"`
	Timestamp time.Time    `json:"timestamp"`
}

type ErrorDetail struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
}

// ErrorHandler 错误处理器
type ErrorHandler struct {
	logger *slog.Logger
	config ErrorHandlerConfig
}

// ErrorHandlerConfig 错误处理器配置
type ErrorHandlerConfig struct {
	IncludeStackTrace  bool
	IncludeContext     bool
	IncludeSuggestions bool
	DebugMode          bool
}

// NewErrorHandler 创建错误处理器
func NewErrorHandler(logger *slog.Logger, config ErrorHandlerConfig) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
		config: config,
	}
}

// ErrorHandlingMiddleware 统一错误处理中间件
func (eh *ErrorHandler) ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic错误
				stack := debug.Stack()
				appErr := errors.NewPanicError(err, string(stack))

				// 添加请求上下文
				eh.addRequestContext(appErr, c)

				// 记录错误日志
				eh.logError(appErr, c)

				// 返回错误响应
				eh.respondWithError(c, appErr)

				// 触发告警（对于panic错误）
				eh.triggerAlert(appErr, c)
			}
		}()

		c.Next()

		// 处理处理过程中的错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var appErr errors.AppError

			// 检查是否已经是AppError
			if ae, ok := err.(errors.AppError); ok {
				appErr = ae
			} else {
				// 包装为应用错误
				appErr = errors.NewInternalError("unexpected_error", "Unexpected error occurred", err)
			}

			// 添加请求上下文
			eh.addRequestContext(appErr, c)

			// 记录错误日志
			eh.logError(appErr, c)

			// 返回错误响应
			eh.respondWithError(c, appErr)

			// 触发告警（对于严重错误）
			if appErr.Severity() == errors.SeverityCritical ||
				appErr.Severity() == errors.SeverityHigh {
				eh.triggerAlert(appErr, c)
			}
		}
	}
}

// 添加请求上下文
func (eh *ErrorHandler) addRequestContext(appErr errors.AppError, c *gin.Context) {
	appErr.AddContext("request_id", c.GetString("request_id"))
	appErr.AddContext("method", c.Request.Method)
	appErr.AddContext("path", c.Request.URL.Path)
	appErr.AddContext("query", c.Request.URL.RawQuery)
	appErr.AddContext("client_ip", c.ClientIP())
	appErr.AddContext("user_agent", c.Request.UserAgent())

	// 添加用户信息
	if userID := c.GetString("user_id"); userID != "" {
		appErr.AddContext("user_id", userID)
	}
	if username := c.GetString("username"); username != "" {
		appErr.AddContext("username", username)
	}
	if role := c.GetString("role"); role != "" {
		appErr.AddContext("role", role)
	}
}

// 记录错误日志
func (eh *ErrorHandler) logError(appErr errors.AppError, c *gin.Context) {
	attrs := []slog.Attr{
		slog.String("error_code", appErr.Code()),
		slog.String("error_message", appErr.Message()),
		slog.String("error_severity", eh.severityToString(appErr.Severity())),
		slog.Time("timestamp", time.Now()),
	}

	// 添加请求信息
	if requestID := c.GetString("request_id"); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	// 添加上下文信息
	if context := appErr.Context(); len(context) > 0 {
		contextJSON, _ := json.Marshal(context)
		attrs = append(attrs, slog.String("context", string(contextJSON)))
	}

	// 添加堆栈跟踪（在调试模式下）
	if eh.config.DebugMode && appErr.StackTrace() != "" {
		attrs = append(attrs, slog.String("stack_trace", appErr.StackTrace()))
	}

	// 根据严重程度选择日志级别
	if eh.logger != nil {
		switch appErr.Severity() {
		case errors.SeverityCritical, errors.SeverityHigh:
			eh.logger.Error("Application error", convertSlogAttrsToAny(attrs)...)
		case errors.SeverityMedium:
			eh.logger.Warn("Application warning", convertSlogAttrsToAny(attrs)...)
		default:
			eh.logger.Info("Application info", convertSlogAttrsToAny(attrs)...)
		}
	}
}

// 返回错误响应
func (eh *ErrorHandler) respondWithError(c *gin.Context, appErr errors.AppError) {
	response := ErrorResponse{
		Success:   false,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now(),
	}

	// 构建错误详情
	errorDetail := ErrorDetail{
		Code:    appErr.Code(),
		Message: appErr.Message(),
		Details: appErr.Details(),
	}

	// 添加上下文信息
	if eh.config.IncludeContext && len(appErr.Context()) > 0 {
		errorDetail.Context = appErr.Context()
	}

	// 添加堆栈跟踪
	if eh.config.IncludeStackTrace && eh.config.DebugMode && appErr.StackTrace() != "" {
		errorDetail.StackTrace = appErr.StackTrace()
	}

	// 添加建议
	if eh.config.IncludeSuggestions {
		errorDetail.Suggestions = eh.getSuggestions(appErr)
	}

	response.Error = &errorDetail

	// 设置HTTP状态码
	statusCode := appErr.HTTPStatus()
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, response)
}

// 获取错误建议
func (eh *ErrorHandler) getSuggestions(appErr errors.AppError) []string {
	var suggestions []string

	switch appErr.Code() {
	case "VALIDATION_ERROR":
		suggestions = []string{
			"请检查输入数据的格式和内容",
			"确保所有必填字段都已填写",
			"检查数据类型是否正确",
		}
	case "DATABASE_ERROR":
		suggestions = []string{
			"请稍后重试",
			"如果问题持续存在，请联系系统管理员",
		}
	case "AUTHORIZATION_ERROR":
		suggestions = []string{
			"请确认您有访问此资源的权限",
			"请联系管理员获取相应权限",
		}
	case "CONCURRENCY_ERROR":
		suggestions = []string{
			"请稍后重试",
			"可能是其他用户正在修改相同资源",
		}
	case "NETWORK_ERROR":
		suggestions = []string{
			"请检查网络连接",
			"稍后重试",
		}
	case "NOT_FOUND":
		suggestions = []string{
			"请确认资源ID是否正确",
			"可能该资源已被删除",
		}
	default:
		suggestions = []string{
			"请稍后重试",
			"如果问题持续存在，请联系技术支持",
		}
	}

	return suggestions
}

// 触发告警
func (eh *ErrorHandler) triggerAlert(appErr errors.AppError, c *gin.Context) {
	// 这里可以集成到外部告警系统
	// 例如：发送到Slack、邮件、PagerDuty等

	// 对于演示，我们只记录告警日志
	if eh.logger != nil {
		eh.logger.Error("ALERT TRIGGERED",
			slog.String("error_code", appErr.Code()),
			slog.String("error_message", appErr.Message()),
			slog.String("severity", eh.severityToString(appErr.Severity())),
			slog.String("request_id", c.GetString("request_id")),
			slog.String("endpoint", c.Request.URL.Path),
		)
	}
}

// 将严重程度转换为字符串
func (eh *ErrorHandler) severityToString(severity errors.ErrorSeverity) string {
	switch severity {
	case errors.SeverityLow:
		return "LOW"
	case errors.SeverityMedium:
		return "MEDIUM"
	case errors.SeverityHigh:
		return "HIGH"
	case errors.SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// 便捷函数：创建默认错误处理中间件
func DefaultErrorHandlingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	config := ErrorHandlerConfig{
		IncludeStackTrace:  true,
		IncludeContext:     true,
		IncludeSuggestions: true,
		DebugMode:          true,
	}

	handler := NewErrorHandler(logger, config)
	return handler.ErrorHandlingMiddleware()
}

// 便捷函数：创建生产环境错误处理中间件
func ProductionErrorHandlingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	config := ErrorHandlerConfig{
		IncludeStackTrace:  false,
		IncludeContext:     false,
		IncludeSuggestions: true,
		DebugMode:          false,
	}

	handler := NewErrorHandler(logger, config)
	return handler.ErrorHandlingMiddleware()
}

// convertSlogAttrsToAny 将slog.Attr转换为[]any
func convertSlogAttrsToAny(attrs []slog.Attr) []any {
	result := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		result = append(result, attr.Key, attr.Value.Any())
	}
	return result
}
