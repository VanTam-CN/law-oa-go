# 统一错误处理模式设计

## 1. 错误处理架构设计

### 1.1 分层错误处理策略

基于Go 1.23最新最佳实践，设计统一的错误处理模式：

```
Repository Layer → Service Layer → Handler Layer → HTTP Response
```

### 1.2 错误类型定义

```go
// 错误层次结构
type AppError interface {
    error
    Code() string
    Message() string
    Details() string
    Context() map[string]interface{}
    HTTPStatus() int
    StackTrace() string
}

// 基础错误实现
type BaseError struct {
    code     string
    message  string
    details  string
    context  map[string]interface{}
    cause    error
    stack    string
    severity ErrorSeverity
}

// 错误严重程度
type ErrorSeverity int

const (
    SeverityLow ErrorSeverity = iota
    SeverityMedium
    SeverityHigh
    SeverityCritical
)
```

### 1.3 预定义错误类型

```go
// 业务错误
type BusinessError struct {
    BaseError
    EntityType string
    EntityID   interface{}
}

// 验证错误
type ValidationError struct {
    BaseError
    Field   string
    Value   interface{}
    Rules   []string
}

// 数据库错误
type DatabaseError struct {
    BaseError
    Operation string
    Table     string
    Query     string
}

// 权限错误
type AuthorizationError struct {
    BaseError
    RequiredPermission string
    CurrentPermission string
}

// 并发错误
type ConcurrencyError struct {
    BaseError
    ResourceType string
    ResourceID   interface{}
    ConflictType string
}
```

## 2. 错误处理中间件

### 2.1 错误恢复中间件

```go
// 统一错误处理中间件
func ErrorHandlingMiddleware(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 记录panic错误
                stack := debug.Stack()
                appErr := NewPanicError(err, string(stack))
                
                // 记录错误日志
                logError(logger, appErr, c)
                
                // 返回错误响应
                respondWithError(c, appErr)
                
                // 触发告警
                triggerAlert(appErr, c)
            }
        }()
        
        c.Next()
        
        // 处理处理过程中的错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            if appErr, ok := err.(AppError); ok {
                respondWithError(c, appErr)
            } else {
                // 包装为应用错误
                appErr = NewInternalError("unexpected error", err)
                respondWithError(c, appErr)
            }
        }
    }
}
```

### 2.2 错误响应格式

```go
// 统一错误响应
type ErrorResponse struct {
    Success   bool                   `json:"success"`
    Error     *ErrorDetail          `json:"error,omitempty"`
    RequestID string                `json:"request_id"`
    Timestamp time.Time             `json:"timestamp"`
}

type ErrorDetail struct {
    Code        string                 `json:"code"`
    Message     string                 `json:"message"`
    Details     string                 `json:"details,omitempty"`
    Context     map[string]interface{} `json:"context,omitempty"`
    Suggestions []string               `json:"suggestions,omitempty"`
    StackTrace  string                 `json:"stack_trace,omitempty"`
}
```

## 3. 错误创建和包装工具

### 3.1 错误创建函数

```go
// 业务错误
func NewBusinessError(code, message string, cause error) *BusinessError {
    return &BusinessError{
        BaseError: BaseError{
            code:     code,
            message:  message,
            cause:    cause,
            severity: SeverityMedium,
        },
    }
}

// 验证错误
func NewValidationError(field string, value interface{}, message string, rules []string) *ValidationError {
    return &ValidationError{
        BaseError: BaseError{
            code:     "VALIDATION_ERROR",
            message:  fmt.Sprintf("%s: %s", field, message),
            severity: SeverityLow,
        },
        Field: field,
        Value: value,
        Rules: rules,
    }
}

// 数据库错误
func NewDatabaseError(operation string, cause error) *DatabaseError {
    return &DatabaseError{
        BaseError: BaseError{
            code:     "DATABASE_ERROR",
            message:  fmt.Sprintf("Database operation failed: %s", operation),
            cause:    cause,
            severity: SeverityHigh,
        },
        Operation: operation,
    }
}

// 权限错误
func NewAuthorizationError(message string, required, current string) *AuthorizationError {
    return &AuthorizationError{
        BaseError: BaseError{
            code:     "AUTHORIZATION_ERROR",
            message:  message,
            severity: SeverityHigh,
        },
        RequiredPermission: required,
        CurrentPermission:  current,
    }
}
```

### 3.2 错误检查和转换工具

```go
// 检查错误类型
func IsBusinessError(err error) bool {
    var businessErr *BusinessError
    return errors.As(err, &businessErr)
}

func IsValidationError(err error) bool {
    var validationErr *ValidationError
    return errors.As(err, &validationErr)
}

func IsDatabaseError(err error) bool {
    var databaseErr *DatabaseError
    return errors.As(err, &databaseErr)
}

func IsAuthorizationError(err error) bool {
    var authErr *AuthorizationError
    return errors.As(err, &authErr)
}

// 获取错误代码
func GetErrorCode(err error) string {
    if appErr, ok := err.(AppError); ok {
        return appErr.Code()
    }
    return "INTERNAL_ERROR"
}

// 获取HTTP状态码
func GetHTTPStatus(err error) int {
    if appErr, ok := err.(AppError); ok {
        return appErr.HTTPStatus()
    }
    return http.StatusInternalServerError
}
```

## 4. 结构化错误日志

### 4.1 错误日志记录

```go
// 结构化错误日志
func logError(logger *slog.Logger, err AppError, c *gin.Context) {
    attrs := []slog.Attr{
        slog.String("error_code", err.Code()),
        slog.String("error_message", err.Message()),
        slog.String("request_id", c.GetString("request_id")),
        slog.String("method", c.Request.Method),
        slog.String("path", c.Request.URL.Path),
        slog.String("ip", c.ClientIP()),
        slog.String("user_agent", c.Request.UserAgent()),
    }
    
    // 添加上下文信息
    if context := err.Context(); len(context) > 0 {
        attrs = append(attrs, slog.Any("context", context))
    }
    
    // 添加用户信息
    if userID := c.GetString("user_id"); userID != "" {
        attrs = append(attrs, slog.String("user_id", userID))
    }
    
    // 根据严重程度选择日志级别
    switch err.Severity() {
    case SeverityCritical, SeverityHigh:
        logger.Error("Application error", attrs...)
    case SeverityMedium:
        logger.Warn("Application warning", attrs...)
    default:
        logger.Info("Application info", attrs...)
    }
}
```

### 4.2 错误上下文管理

```go
// 添加错误上下文
func WithContext(err error, key string, value interface{}) error {
    if appErr, ok := err.(AppError); ok {
        appErr.AddContext(key, value)
        return appErr
    }
    return err
}

// 添加多个上下文
func WithContexts(err error, context map[string]interface{}) error {
    if appErr, ok := err.(AppError); ok {
        for k, v := range context {
            appErr.AddContext(k, v)
        }
        return appErr
    }
    return err
}
```

## 5. 错误恢复和重试机制

### 5.1 重试配置

```go
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
    Jitter       bool
    RetryableErrors map[string]bool
}

// 默认重试配置
var DefaultRetryConfig = RetryConfig{
    MaxAttempts: 3,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     5 * time.Second,
    Multiplier:   2.0,
    Jitter:       true,
    RetryableErrors: map[string]bool{
        "DATABASE_CONNECTION_ERROR": true,
        "NETWORK_ERROR":           true,
        "TIMEOUT_ERROR":           true,
        "CONCURRENCY_ERROR":       true,
    },
}
```

### 5.2 重试函数

```go
// 带重试的函数执行
func WithRetry[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
    var lastErr error
    var zero T
    
    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        result, err := fn()
        if err == nil {
            return result, nil
        }
        
        // 检查是否可重试
        errorCode := GetErrorCode(err)
        if !config.RetryableErrors[errorCode] {
            return zero, err
        }
        
        lastErr = err
        
        // 计算延迟时间
        delay := calculateDelay(config, attempt)
        
        // 等待延迟
        select {
        case <-time.After(delay):
            continue
        case <-ctx.Done():
            return zero, ctx.Err()
        }
    }
    
    return zero, fmt.Errorf("max retry attempts reached, last error: %w", lastErr)
}
```

## 6. 错误监控和告警

### 6.1 错误指标收集

```go
// 错误指标
type ErrorMetrics struct {
    errorsTotal      *prometheus.CounterVec
    errorsByType     *prometheus.CounterVec
    errorsByEndpoint *prometheus.CounterVec
    errorDuration    *prometheus.HistogramVec
}

func (em *ErrorMetrics) RecordError(err error, endpoint string) {
    errorCode := GetErrorCode(err)
    
    em.errorsTotal.WithLabelValues(errorCode).Inc()
    em.errorsByType.WithLabelValues(errorCode, endpoint).Inc()
    em.errorsByEndpoint.WithLabelValues(endpoint).Inc()
    
    // 记录错误处理时间
    em.errorDuration.Observe(time.Since(startTime).Seconds())
}
```

### 6.2 告警触发

```go
// 告警触发器
type AlertTrigger struct {
    errorCounts      map[string]int64
    alertThresholds  map[string]int64
    lastAlertTime    map[string]time.Time
    cooldownPeriod   time.Duration
}

func (at *AlertTrigger) CheckAndTrigger(err AppError, endpoint string) {
    errorCode := err.Code()
    
    // 更新错误计数
    at.errorCounts[errorCode]++
    
    // 检查是否达到告警阈值
    if threshold, exists := at.alertThresholds[errorCode]; exists {
        if at.errorCounts[errorCode] >= threshold {
            lastAlert, exists := at.lastAlertTime[errorCode]
            if !exists || time.Since(lastAlert) > at.cooldownPeriod {
                at.triggerAlert(err, endpoint)
                at.lastAlertTime[errorCode] = time.Now()
                at.errorCounts[errorCode] = 0 // 重置计数
            }
        }
    }
}
```

## 7. 迁移策略

### 7.1 逐步迁移计划

1. **第一阶段**: 实现基础错误处理框架
2. **第二阶段**: 更新Handler层错误处理
3. **第三阶段**: 更新Service层错误处理
4. **第四阶段**: 更新Repository层错误处理
5. **第五阶段**: 添加监控和告警

### 7.2 兼容性考虑

- 保持现有API兼容性
- 逐步替换字符串比较错误检查
- 保持错误消息格式一致
- 提供迁移工具和脚本

## 8. 测试策略

### 8.1 单元测试

- 测试错误创建和包装
- 测试错误类型检查
- 测试错误上下文管理
- 测试重试机制

### 8.2 集成测试

- 测试错误处理中间件
- 测试错误响应格式
- 测试错误日志记录
- 测试告警触发

### 8.3 性能测试

- 测试错误处理性能影响
- 测试高并发下的错误处理
- 测试内存使用情况

这个统一错误处理模式设计解决了当前代码中的所有问题，提供了完整的错误处理解决方案，符合Go 1.23的最佳实践。