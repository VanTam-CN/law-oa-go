package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerManager 错误处理器管理器
type ErrorHandlerManager struct {
	logger *slog.Logger
	config ErrorHandlerConfig
}

// ErrorHandlerConfig 错误处理器配置
type ErrorHandlerConfig struct {
	EnableStackTrace  bool          `json:"enable_stack_trace"`
	EnableContext     bool          `json:"enable_context"`
	EnableSuggestions bool          `json:"enable_suggestions"`
	DebugMode         bool          `json:"debug_mode"`
	LogLevel          string        `json:"log_level"`
	AlertEnabled      bool          `json:"alert_enabled"`
	MaxRetries        int           `json:"max_retries"`
	RetryDelay        time.Duration `json:"retry_delay"`
}

// DefaultErrorHandlerConfig 默认错误处理器配置
func DefaultErrorHandlerConfig() ErrorHandlerConfig {
	return ErrorHandlerConfig{
		EnableStackTrace:  true,
		EnableContext:     true,
		EnableSuggestions: true,
		DebugMode:         true,
		LogLevel:          "info",
		AlertEnabled:      true,
		MaxRetries:        3,
		RetryDelay:        time.Second * 2,
	}
}

// ProductionErrorHandlerConfig 生产环境错误处理器配置
func ProductionErrorHandlerConfig() ErrorHandlerConfig {
	return ErrorHandlerConfig{
		EnableStackTrace:  false,
		EnableContext:     false,
		EnableSuggestions: true,
		DebugMode:         false,
		LogLevel:          "error",
		AlertEnabled:      true,
		MaxRetries:        3,
		RetryDelay:        time.Second * 2,
	}
}

// NewErrorHandlerManager 创建错误处理器管理器
func NewErrorHandlerManager(logger *slog.Logger, config ErrorHandlerConfig) *ErrorHandlerManager {
	return &ErrorHandlerManager{
		logger: logger,
		config: config,
	}
}

// HandleError 处理错误
func (ehm *ErrorHandlerManager) HandleError(ctx context.Context, err error) *ErrorHandlingResult {
	result := &ErrorHandlingResult{
		Error:     err,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
		Metrics:   make(map[string]interface{}),
	}

	if err == nil {
		result.Handled = true
		return result
	}

	// 转换为增强错误
	enhancedErr := ehm.toEnhancedError(err)
	result.EnhancedError = enhancedErr

	// 记录错误日志
	ehm.logError(ctx, enhancedErr)

	// 收集指标
	ehm.collectMetrics(enhancedErr, result)

	// 触发告警（如果需要）
	if ehm.shouldAlert(enhancedErr) {
		ehm.triggerAlert(ctx, enhancedErr)
	}

	// 尝试恢复（如果可恢复）
	if ehm.isRecoverable(enhancedErr) {
		if recoveredErr := ehm.tryRecover(ctx, enhancedErr); recoveredErr != nil {
			result.RecoveryError = recoveredErr
			result.Recovered = false
		} else {
			result.Recovered = true
		}
	}

	result.Handled = true
	return result
}

// ErrorHandlingResult 错误处理结果
type ErrorHandlingResult struct {
	Error         error                  `json:"error"`
	EnhancedError *EnhancedError         `json:"enhanced_error,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Context       map[string]interface{} `json:"context"`
	Metrics       map[string]interface{} `json:"metrics"`
	Handled       bool                   `json:"handled"`
	Recovered     bool                   `json:"recovered"`
	RecoveryError error                  `json:"recovery_error,omitempty"`
	Suggestions   []string               `json:"suggestions,omitempty"`
}

// toEnhancedError 转换为增强错误
func (ehm *ErrorHandlerManager) toEnhancedError(err error) *EnhancedError {
	// 如果已经是增强错误，直接返回
	var enhancedErr *EnhancedError
	if err != nil && errors.As(err, &enhancedErr) {
		return enhancedErr
	}

	// 包装为内部错误
	return InternalError("Unexpected error occurred", err)
}

// logError 记录错误日志
func (ehm *ErrorHandlerManager) logError(ctx context.Context, err *EnhancedError) {
	if ehm.logger == nil {
		return
	}

	attrs := []any{
		slog.String("error_id", err.ID()),
		slog.String("error_code", err.Code()),
		slog.String("error_category", err.Category().String()),
		slog.String("error_level", err.Level().String()),
		slog.String("error_message", err.Message()),
		slog.Time("timestamp", err.Timestamp()),
	}

	// 添加跟踪ID
	if err.TraceID() != "" {
		attrs = append(attrs, slog.String("trace_id", err.TraceID()))
	}

	// 添加上下文信息
	if ehm.config.EnableContext && len(err.Context()) > 0 {
		contextJSON, _ := json.Marshal(err.Context())
		attrs = append(attrs, slog.String("context", string(contextJSON)))
	}

	// 添加堆栈跟踪
	if ehm.config.EnableStackTrace && err.StackTrace() != "" {
		attrs = append(attrs, slog.String("stack_trace", err.StackTrace()))
	}

	// 添加源码位置
	if err.File() != "" {
		attrs = append(attrs,
			slog.String("source_file", err.File()),
			slog.Int("source_line", err.Line()),
			slog.String("source_function", err.Function()),
		)
	}

	// 根据错误级别选择日志级别
	switch err.Level() {
	case ErrorLevelCritical:
		ehm.logger.Error("Critical error occurred", attrs...)
	case ErrorLevelError:
		ehm.logger.Error("Error occurred", attrs...)
	case ErrorLevelWarning:
		ehm.logger.Warn("Warning occurred", attrs...)
	default:
		ehm.logger.Info("Info error occurred", attrs...)
	}
}

// collectMetrics 收集指标
func (ehm *ErrorHandlerManager) collectMetrics(err *EnhancedError, result *ErrorHandlingResult) {
	result.Metrics["error_category"] = err.Category().String()
	result.Metrics["error_level"] = err.Level().String()
	result.Metrics["error_code"] = err.Code()
	result.Metrics["timestamp"] = err.Timestamp().Unix()

	if err.HTTPStatus() != 0 {
		result.Metrics["http_status"] = err.HTTPStatus()
	}

	// 添加处理时间
	result.Metrics["handling_duration_ms"] = time.Since(result.Timestamp).Milliseconds()

	// 根据错误类型添加特定指标
	switch err.Category() {
	case ErrorCategoryDatabase:
		result.Metrics["database_errors_total"] = 1
	case ErrorCategoryNetwork:
		result.Metrics["network_errors_total"] = 1
	case ErrorCategorySecurity:
		result.Metrics["security_errors_total"] = 1
	case ErrorCategorySystem:
		result.Metrics["system_errors_total"] = 1
	}
}

// shouldAlert 判断是否需要告警
func (ehm *ErrorHandlerManager) shouldAlert(err *EnhancedError) bool {
	if !ehm.config.AlertEnabled {
		return false
	}

	// 关键错误和系统错误需要告警
	return err.Level() == ErrorLevelCritical ||
		err.Category() == ErrorCategorySystem ||
		err.Category() == ErrorCategorySecurity
}

// triggerAlert 触发告警
func (ehm *ErrorHandlerManager) triggerAlert(ctx context.Context, err *EnhancedError) {
	if !ehm.shouldAlert(err) {
		return
	}

	alert := Alert{
		ID:        err.ID(),
		Type:      "error",
		Severity:  err.Level().String(),
		Category:  err.Category().String(),
		Message:   err.Message(),
		Timestamp: err.Timestamp(),
		TraceID:   err.TraceID(),
		Context:   err.Context(),
	}

	// 记录告警日志
	if ehm.logger != nil {
		ehm.logger.Error("ALERT TRIGGERED",
			slog.String("alert_id", alert.ID),
			slog.String("alert_type", alert.Type),
			slog.String("severity", alert.Severity),
			slog.String("category", alert.Category),
			slog.String("message", alert.Message),
			slog.String("trace_id", alert.TraceID),
		)
	}

	// 这里可以集成到外部告警系统
	// 例如：发送到Slack、钉钉、邮件、PagerDuty等
	ehm.sendAlert(ctx, alert)
}

// Alert 告警结构
type Alert struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"`
	Category  string                 `json:"category"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	TraceID   string                 `json:"trace_id"`
	Context   map[string]interface{} `json:"context"`
}

// sendAlert 发送告警（示例实现）
func (ehm *ErrorHandlerManager) sendAlert(ctx context.Context, alert Alert) {
	// 这里可以集成到具体的告警系统
	// 例如：调用外部API、发送邮件、Slack通知等

	// 为演示目的，只记录日志
	if ehm.logger != nil {
		ehm.logger.Info("Alert sent to external system",
			slog.String("alert_id", alert.ID),
			slog.String("message", alert.Message),
		)
	}
}

// isRecoverable 判断错误是否可恢复
func (ehm *ErrorHandlerManager) isRecoverable(err *EnhancedError) bool {
	// 网络错误和一些业务错误通常可以恢复
	return err.Category() == ErrorCategoryNetwork ||
		(err.Category() == ErrorCategoryBusiness && err.Level() != ErrorLevelCritical)
}

// tryRecover 尝试恢复错误
func (ehm *ErrorHandlerManager) tryRecover(ctx context.Context, err *EnhancedError) error {
	// 这里可以实现具体的恢复逻辑
	// 例如：重试数据库连接、使用缓存降级等

	if err.Category() == ErrorCategoryNetwork {
		return ehm.retryNetworkOperation(ctx, err)
	}

	if err.Category() == ErrorCategoryDatabase {
		return ehm.retryDatabaseOperation(ctx, err)
	}

	// 其他类型的错误暂时不实现恢复逻辑
	return fmt.Errorf("recovery not implemented for error category: %s", err.Category())
}

// retryNetworkOperation 重试网络操作
func (ehm *ErrorHandlerManager) retryNetworkOperation(ctx context.Context, err *EnhancedError) error {
	if ehm.config.MaxRetries <= 0 {
		return fmt.Errorf("network retry not implemented")
	}

	// 实现网络操作重试逻辑
	for i := 0; i < ehm.config.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(ehm.config.RetryDelay):
			// 这里可以重新执行网络操作
			// 为演示目的，只记录重试日志
			if ehm.logger != nil {
				ehm.logger.Info("Retrying network operation",
					slog.Int("attempt", i+1),
					slog.String("error_id", err.ID()),
				)
			}
		}
	}

	return fmt.Errorf("network operation failed after %d retries", ehm.config.MaxRetries)
}

// retryDatabaseOperation 重试数据库操作
func (ehm *ErrorHandlerManager) retryDatabaseOperation(ctx context.Context, err *EnhancedError) error {
	if ehm.config.MaxRetries <= 0 {
		return fmt.Errorf("database retry not implemented")
	}

	// 实现数据库操作重试逻辑
	for i := 0; i < ehm.config.MaxRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(ehm.config.RetryDelay):
			// 这里可以重新执行数据库操作
			if ehm.logger != nil {
				ehm.logger.Info("Retrying database operation",
					slog.Int("attempt", i+1),
					slog.String("error_id", err.ID()),
				)
			}
		}
	}

	return fmt.Errorf("database operation failed after %d retries", ehm.config.MaxRetries)
}

// FormatErrorResponse 格式化错误响应
func (ehm *ErrorHandlerManager) FormatErrorResponse(err *EnhancedError, requestID string) gin.H {
	response := gin.H{
		"success":    false,
		"request_id": requestID,
		"timestamp":  time.Now(),
		"error": gin.H{
			"id":       err.ID(),
			"code":     err.Code(),
			"category": err.Category().String(),
			"level":    err.Level().String(),
			"message":  err.Message(),
		},
	}

	// 添加详细信息
	if err.Details() != "" {
		response["error"].(gin.H)["details"] = err.Details()
	}

	// 添加上下文信息
	if ehm.config.EnableContext && len(err.Context()) > 0 {
		response["error"].(gin.H)["context"] = err.Context()
	}

	// 添加堆栈跟踪
	if ehm.config.EnableStackTrace && ehm.config.DebugMode && err.StackTrace() != "" {
		response["error"].(gin.H)["stack_trace"] = err.StackTrace()
	}

	// 添加建议
	if ehm.config.EnableSuggestions && len(err.Suggestions()) > 0 {
		response["error"].(gin.H)["suggestions"] = err.Suggestions()
	} else {
		response["error"].(gin.H)["suggestions"] = ehm.getDefaultSuggestions(err)
	}

	// 添加源码信息
	if ehm.config.DebugMode && err.File() != "" {
		response["error"].(gin.H)["source"] = gin.H{
			"file":     err.File(),
			"line":     err.Line(),
			"function": err.Function(),
		}
	}

	return response
}

// getDefaultSuggestions 获取默认错误建议
func (ehm *ErrorHandlerManager) getDefaultSuggestions(err *EnhancedError) []string {
	switch err.category {
	case ErrorCategoryValidation:
		return []string{
			"请检查输入数据的格式和内容",
			"确保所有必填字段都已填写",
			"检查数据类型是否正确",
		}
	case ErrorCategoryDatabase:
		return []string{
			"请稍后重试",
			"如果问题持续存在，请联系系统管理员",
		}
	case ErrorCategoryNetwork:
		return []string{
			"请检查网络连接",
			"稍后重试",
			"确认服务是否正常运行",
		}
	case ErrorCategorySecurity:
		return []string{
			"请确认您的访问权限",
			"联系管理员获取相应权限",
			"检查您的认证状态",
		}
	case ErrorCategoryBusiness:
		return []string{
			"请确认操作是否符合业务规则",
			"联系业务管理员获取帮助",
			"查看相关文档了解操作要求",
		}
	case ErrorCategorySystem:
		return []string{
			"请稍后重试",
			"如果问题持续存在，请联系技术支持",
			"记录错误ID以便问题追踪",
		}
	default:
		return []string{
			"请稍后重试",
			"如果问题持续存在，请联系技术支持",
		}
	}
}

// WrapErrorForGin 为Gin框架包装错误
func (ehm *ErrorHandlerManager) WrapErrorForGin(err error, c *gin.Context) {
	if err == nil {
		return
	}

	// 转换为增强错误
	enhancedErr := ehm.toEnhancedError(err)

	// 添加请求上下文
	ehm.addGinContext(enhancedErr, c)

	// 记录错误
	ehm.logError(c.Request.Context(), enhancedErr)

	// 添加到Gin错误队列
	c.Error(enhancedErr)
}

// addGinContext 添加Gin请求上下文
func (ehm *ErrorHandlerManager) addGinContext(err *EnhancedError, c *gin.Context) {
	// 添加请求信息
	err.AddContext("method", c.Request.Method)
	err.AddContext("path", c.Request.URL.Path)
	err.AddContext("client_ip", c.ClientIP())
	err.AddContext("user_agent", c.Request.UserAgent())

	// 添加请求参数
	if c.Request.URL.RawQuery != "" {
		err.AddContext("query_params", c.Request.URL.RawQuery)
	}

	// 添加用户信息
	if userID := c.GetString("user_id"); userID != "" {
		err.AddContext("user_id", userID)
	}
	if username := c.GetString("username"); username != "" {
		err.AddContext("username", username)
	}
	if role := c.GetString("role"); role != "" {
		err.AddContext("role", role)
	}

	// 添加请求ID
	if requestID := c.GetString("request_id"); requestID != "" {
		err.SetTraceID(requestID)
	}
}

// GetErrorStats 获取错误统计信息
func (ehm *ErrorHandlerManager) GetErrorStats() map[string]interface{} {
	// 这里可以实现错误统计逻辑
	// 例如：统计各种类型错误的发生次数、趋势等
	return map[string]interface{}{
		"total_errors":      0,
		"critical_errors":   0,
		"system_errors":     0,
		"network_errors":    0,
		"database_errors":   0,
		"validation_errors": 0,
		"business_errors":   0,
		"security_errors":   0,
		"last_error_time":   nil,
	}
}

// GetCallerInfo 获取调用者信息
func GetCallerInfo(skip int) (file string, line int, function string) {
	pc, file, line, ok := runtime.Caller(skip)
	if ok {
		function = runtime.FuncForPC(pc).Name()
		// 处理文件路径，只显示相对路径
		if slash := strings.LastIndex(file, "/"); slash >= 0 {
			file = file[slash+1:]
		}
	}
	return file, line, function
}
