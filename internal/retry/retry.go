package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"law-oa-go/internal/errors"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	Multiplier      float64
	Jitter          bool
	RetryableErrors map[string]bool
}

// 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:  3,
	InitialDelay: 100 * time.Millisecond,
	MaxDelay:     5 * time.Second,
	Multiplier:   2.0,
	Jitter:       true,
	RetryableErrors: map[string]bool{
		"DATABASE_CONNECTION_ERROR": true,
		"NETWORK_ERROR":             true,
		"TIMEOUT_ERROR":             true,
		"CONCURRENCY_ERROR":         true,
		"INTERNAL_ERROR":            true,
	},
}

// RetryableErrorChecker 可重试错误检查器
type RetryableErrorChecker func(error) bool

// WithRetry 带重试的函数执行
func WithRetry[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) (T, error) {
	var lastErr error
	var zero T

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		// 检查是否可重试
		if !isRetryableError(err, config) {
			return zero, err
		}

		lastErr = err

		// 计算延迟时间
		delay := calculateDelay(config, attempt)

		// 等待延迟或上下文取消
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return zero, fmt.Errorf("retry canceled: %w", ctx.Err())
		}
	}

	return zero, fmt.Errorf("max retry attempts (%d) reached, last error: %w", config.MaxAttempts, lastErr)
}

// WithRetryContext 带重试和错误上下文的函数执行
func WithRetryContext[T any](ctx context.Context, config RetryConfig, fn func(context.Context) (T, error)) (T, error) {
	var lastErr error
	var zero T

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		// 检查是否可重试
		if !isRetryableError(err, config) {
			return zero, err
		}

		lastErr = err

		// 计算延迟时间
		delay := calculateDelay(config, attempt)

		// 等待延迟或上下文取消
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return zero, fmt.Errorf("retry canceled: %w", ctx.Err())
		}
	}

	return zero, fmt.Errorf("max retry attempts (%d) reached, last error: %w", config.MaxAttempts, lastErr)
}

// RetryableOperation 可重试操作接口
type RetryableOperation[T any] interface {
	Execute(ctx context.Context) (T, error)
	IsRetryable(err error) bool
	GetRetryConfig() RetryConfig
}

// ExecuteWithRetry 执行可重试操作
func ExecuteWithRetry[T any](ctx context.Context, op RetryableOperation[T]) (T, error) {
	config := op.GetRetryConfig()
	return WithRetryContext(ctx, config, op.Execute)
}

// DatabaseRetryableOperation 数据库可重试操作
type DatabaseRetryableOperation[T any] struct {
	operation func(ctx context.Context) (T, error)
	config    RetryConfig
}

func (op *DatabaseRetryableOperation[T]) Execute(ctx context.Context) (T, error) {
	return op.operation(ctx)
}

func (op *DatabaseRetryableOperation[T]) IsRetryable(err error) bool {
	// 数据库连接错误通常可以重试
	if errors.IsDatabaseError(err) {
		return true
	}

	// 检查特定错误代码
	errorCode := errors.GetErrorCode(err)
	return op.config.RetryableErrors[errorCode]
}

func (op *DatabaseRetryableOperation[T]) GetRetryConfig() RetryConfig {
	return op.config
}

// NewDatabaseRetryableOperation 创建数据库可重试操作
func NewDatabaseRetryableOperation[T any](
	operation func(ctx context.Context) (T, error),
	config RetryConfig,
) *DatabaseRetryableOperation[T] {
	return &DatabaseRetryableOperation[T]{
		operation: operation,
		config:    config,
	}
}

// APIRetryableOperation API可重试操作
type APIRetryableOperation[T any] struct {
	operation func(ctx context.Context) (T, error)
	config    RetryConfig
}

func (op *APIRetryableOperation[T]) Execute(ctx context.Context) (T, error) {
	return op.operation(ctx)
}

func (op *APIRetryableOperation[T]) IsRetryable(err error) bool {
	// 网络错误通常可以重试
	if errors.IsNetworkError(err) {
		return true
	}

	// 检查特定错误代码
	errorCode := errors.GetErrorCode(err)
	return op.config.RetryableErrors[errorCode]
}

func (op *APIRetryableOperation[T]) GetRetryConfig() RetryConfig {
	return op.config
}

// NewAPIRetryableOperation 创建API可重试操作
func NewAPIRetryableOperation[T any](
	operation func(ctx context.Context) (T, error),
	config RetryConfig,
) *APIRetryableOperation[T] {
	return &APIRetryableOperation[T]{
		operation: operation,
		config:    config,
	}
}

// isRetryableError 检查错误是否可重试
func isRetryableError(err error, config RetryConfig) bool {
	// 检查错误代码
	errorCode := errors.GetErrorCode(err)
	if retryable, exists := config.RetryableErrors[errorCode]; exists {
		return retryable
	}

	// 检查错误类型
	if errors.IsNetworkError(err) || errors.IsDatabaseError(err) {
		return true
	}

	return false
}

// calculateDelay 计算重试延迟时间
func calculateDelay(config RetryConfig, attempt int) time.Duration {
	// 指数退避
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt-1))

	// 限制最大延迟
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// 添加抖动
	if config.Jitter {
		delay = delay * (0.5 + 0.5*rand.Float64())
	}

	return time.Duration(delay)
}

// RetryResult 重试结果
type RetryResult[T any] struct {
	Value    T
	Error    error
	Attempts int
	Duration time.Duration
	Success  bool
}

// WithRetryResult 带重试结果统计的函数执行
func WithRetryResult[T any](ctx context.Context, config RetryConfig, fn func() (T, error)) *RetryResult[T] {
	start := time.Now()
	var lastErr error

	result := &RetryResult[T]{
		Attempts: 0,
		Success:  false,
	}

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		result.Attempts = attempt

		value, err := fn()
		if err == nil {
			result.Value = value
			result.Success = true
			result.Duration = time.Since(start)
			return result
		}

		lastErr = err

		// 检查是否可重试
		if !isRetryableError(err, config) {
			break
		}

		// 计算延迟时间
		delay := calculateDelay(config, attempt)

		// 等待延迟或上下文取消
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			lastErr = ctx.Err()
			break
		}
	}

	result.Error = lastErr
	result.Duration = time.Since(start)
	return result
}

// RetryMetrics 重试指标
type RetryMetrics struct {
	TotalAttempts      int64
	SuccessfulAttempts int64
	FailedAttempts     int64
	RetryCount         int64
	AverageRetryDelay  time.Duration
	MaxRetryDelay      time.Duration
}

// NewRetryMetrics 创建重试指标
func NewRetryMetrics() *RetryMetrics {
	return &RetryMetrics{}
}

// RecordResult 记录重试结果
func (rm *RetryMetrics) RecordResult(result *RetryResult[any]) {
	rm.TotalAttempts++

	if result.Success {
		rm.SuccessfulAttempts++
	} else {
		rm.FailedAttempts++
	}

	if result.Attempts > 1 {
		rm.RetryCount += int64(result.Attempts - 1)
	}
}

// GetSuccessRate 获取成功率
func (rm *RetryMetrics) GetSuccessRate() float64 {
	if rm.TotalAttempts == 0 {
		return 0.0
	}
	return float64(rm.SuccessfulAttempts) / float64(rm.TotalAttempts)
}

// GetAverageRetryCount 获取平均重试次数
func (rm *RetryMetrics) GetAverageRetryCount() float64 {
	if rm.SuccessfulAttempts == 0 {
		return 0.0
	}
	return float64(rm.RetryCount) / float64(rm.SuccessfulAttempts)
}
