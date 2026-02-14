package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apperrors "law-oa-go/internal/errors"
)

func TestWithRetry_Success(t *testing.T) {
	config := DefaultRetryConfig
	ctx := context.Background()

	attempts := 0
	result, err := WithRetry(ctx, config, func() (string, error) {
		attempts++
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 1, attempts)
}

func TestWithRetry_EventualSuccess(t *testing.T) {
	config := DefaultRetryConfig
	ctx := context.Background()

	attempts := 0
	result, err := WithRetry(ctx, config, func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attempts)
}

func TestWithRetry_MaxAttemptsReached(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		Multiplier:   1.0,
		RetryableErrors: map[string]bool{
			"TEMPORARY_ERROR": true,
		},
	}
	ctx := context.Background()

	attempts := 0
	result, err := WithRetry(ctx, config, func() (string, error) {
		attempts++
		return "", errors.New("temporary error")
	})

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 2, attempts)
	assert.Contains(t, err.Error(), "max retry attempts")
}

func TestWithRetry_ContextCanceled(t *testing.T) {
	config := DefaultRetryConfig
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := WithRetry(ctx, config, func() (string, error) {
		attempts++
		time.Sleep(100 * time.Millisecond)
		return "", errors.New("slow error")
	})

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Contains(t, err.Error(), "retry canceled")
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	config := DefaultRetryConfig
	ctx := context.Background()

	attempts := 0
	result, err := WithRetry(ctx, config, func() (string, error) {
		attempts++
		return "", errors.New("non-retryable error")
	})

	assert.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 1, attempts)
	assert.NotContains(t, err.Error(), "max retry attempts")
}

func TestWithRetryContext(t *testing.T) {
	config := DefaultRetryConfig
	ctx := context.Background()

	attempts := 0
	result, err := WithRetryContext(ctx, config, func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 3, attempts)
}

func TestDatabaseRetryableOperation(t *testing.T) {
	config := DefaultRetryConfig
	ctx := context.Background()

	operation := NewDatabaseRetryableOperation(func(ctx context.Context) (string, error) {
		dbErr := apperrors.DatabaseError("SELECT", "connection failed", errors.New("connection failed"))
		return "", dbErr
	}, config)

	result, err := ExecuteWithRetry(ctx, operation)

	assert.Error(t, err)
	assert.Equal(t, "", result)
}

func TestAPIRetryableOperation(t *testing.T) {
	config := DefaultRetryConfig
	ctx := context.Background()

	operation := NewAPIRetryableOperation(func(ctx context.Context) (string, error) {
		netErr := apperrors.NetworkError("http://api.example.com", false, errors.New("timeout"))
		return "", netErr
	}, config)

	result, err := ExecuteWithRetry(ctx, operation)

	assert.Error(t, err)
	assert.Equal(t, "", result)
}

func TestCalculateDelay(t *testing.T) {
	testCases := []struct {
		name     string
		config   RetryConfig
		attempt  int
		expected time.Duration
	}{
		{
			name: "first attempt",
			config: RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				Multiplier:   2.0,
				Jitter:       false,
			},
			attempt:  1,
			expected: 100 * time.Millisecond,
		},
		{
			name: "second attempt",
			config: RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				Multiplier:   2.0,
				Jitter:       false,
			},
			attempt:  2,
			expected: 200 * time.Millisecond,
		},
		{
			name: "max delay",
			config: RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     300 * time.Millisecond,
				Multiplier:   4.0,
				Jitter:       false,
			},
			attempt:  3,
			expected: 300 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			delay := calculateDelay(tc.config, tc.attempt)
			assert.Equal(t, tc.expected, delay)
		})
	}
}

func TestWithRetryResult(t *testing.T) {
	config := DefaultRetryConfig
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		result := WithRetryResult(ctx, config, func() (string, error) {
			return "success", nil
		})

		assert.True(t, result.Success)
		assert.Equal(t, "success", result.Value)
		assert.NoError(t, result.Error)
		assert.Equal(t, 1, result.Attempts)
	})

	t.Run("failure", func(t *testing.T) {
		result := WithRetryResult(ctx, config, func() (string, error) {
			return "", errors.New("non-retryable error")
		})

		assert.False(t, result.Success)
		assert.Equal(t, "", result.Value)
		assert.Error(t, result.Error)
		assert.Equal(t, 1, result.Attempts)
	})

	t.Run("retry success", func(t *testing.T) {
		attempts := 0
		result := WithRetryResult(ctx, config, func() (string, error) {
			attempts++
			if attempts < 3 {
				return "", errors.New("retryable error")
			}
			return "success", nil
		})

		assert.True(t, result.Success)
		assert.Equal(t, "success", result.Value)
		assert.NoError(t, result.Error)
		assert.Equal(t, 3, result.Attempts)
	})
}

func TestRetryMetrics(t *testing.T) {
	metrics := NewRetryMetrics()

	t.Run("record success", func(t *testing.T) {
		result := &RetryResult[any]{
			Success:  true,
			Attempts: 1,
		}
		metrics.RecordResult(result)

		assert.Equal(t, int64(1), metrics.TotalAttempts)
		assert.Equal(t, int64(1), metrics.SuccessfulAttempts)
		assert.Equal(t, int64(0), metrics.FailedAttempts)
	})

	t.Run("record failure", func(t *testing.T) {
		result := &RetryResult[any]{
			Success:  false,
			Attempts: 1,
		}
		metrics.RecordResult(result)

		assert.Equal(t, int64(2), metrics.TotalAttempts)
		assert.Equal(t, int64(1), metrics.SuccessfulAttempts)
		assert.Equal(t, int64(1), metrics.FailedAttempts)
	})

	t.Run("record retry success", func(t *testing.T) {
		result := &RetryResult[any]{
			Success:  true,
			Attempts: 3,
		}
		metrics.RecordResult(result)

		assert.Equal(t, int64(3), metrics.TotalAttempts)
		assert.Equal(t, int64(2), metrics.SuccessfulAttempts)
		assert.Equal(t, int64(1), metrics.FailedAttempts)
		assert.Equal(t, int64(2), metrics.RetryCount)
	})

	t.Run("get statistics", func(t *testing.T) {
		successRate := metrics.GetSuccessRate()
		assert.Equal(t, 2.0/3.0, successRate)

		avgRetryCount := metrics.GetAverageRetryCount()
		assert.Equal(t, 1.0, avgRetryCount)
	})
}

func TestIsRetryableError(t *testing.T) {
	config := RetryConfig{
		RetryableErrors: map[string]bool{
			"CUSTOM_ERROR": true,
		},
	}

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "custom retryable error",
			err:      apperrors.BusinessError("entity", "CUSTOM_ERROR", "custom error"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      apperrors.BusinessError("entity", "NON_RETRYABLE_ERROR", "non-retryable"),
			expected: false,
		},
		{
			name:     "database error",
			err:      apperrors.DatabaseError("SELECT", "some error", nil),
			expected: true,
		},
		{
			name:     "network error",
			err:      apperrors.NetworkError("http://example.com", false, nil),
			expected: true,
		},
		{
			name:     "validation error",
			err:      apperrors.ValidationError("field", "invalid value"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isRetryableError(tc.err, config)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRetryWithJitter(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
		MaxDelay:     500 * time.Millisecond,
	}

	// 多次计算延迟，验证抖动效果
	delays := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		delays[i] = calculateDelay(config, 2)
	}

	// 验证延迟在合理范围内且存在变化
	for _, delay := range delays {
		assert.Greater(t, delay, time.Duration(0))
		assert.Less(t, delay, 500*time.Millisecond)
	}

	// 验证存在不同的延迟值（抖动效果）
	hasVariation := false
	for i := 1; i < len(delays); i++ {
		if delays[i] != delays[i-1] {
			hasVariation = true
			break
		}
	}
	assert.True(t, hasVariation, "Expected jitter to create variation in delays")
}

func TestRetryWithZeroInitialDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 0,
		Multiplier:   1.0,
		Jitter:       false,
		MaxAttempts:  2,
	}

	delay := calculateDelay(config, 1)
	assert.Equal(t, time.Duration(0), delay)

	delay = calculateDelay(config, 2)
	assert.Equal(t, time.Duration(0), delay)
}

func TestRetryWithMaxDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     300 * time.Millisecond,
		Multiplier:   10.0,
		Jitter:       false,
	}

	// 测试超过最大延迟的情况
	delay := calculateDelay(config, 5)
	assert.Equal(t, 300*time.Millisecond, delay)

	delay = calculateDelay(config, 10)
	assert.Equal(t, 300*time.Millisecond, delay)
}

func BenchmarkWithRetry(b *testing.B) {
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		Multiplier:   1.0,
		Jitter:       false,
		RetryableErrors: map[string]bool{
			"BENCHMARK_ERROR": true,
		},
	}
	ctx := context.Background()

	b.Run("success", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := WithRetry(ctx, config, func() (string, error) {
				return "success", nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("with_retry", func(b *testing.B) {
		attempts := 0
		for i := 0; i < b.N; i++ {
			attempts = 0
			_, err := WithRetry(ctx, config, func() (string, error) {
				attempts++
				if attempts < 3 {
					return "", errors.New("BENCHMARK_ERROR")
				}
				return "success", nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestCustomRetryableOperation(t *testing.T) {
	// 自定义可重试操作
	customOp := &CustomRetryableOperation[string]{
		operation: func() (string, error) {
			return "custom", nil
		},
		isRetryable: func(err error) bool {
			return err.Error() == "custom_retry"
		},
		config: DefaultRetryConfig,
	}

	ctx := context.Background()
	result, err := ExecuteWithRetry[string](ctx, customOp)

	require.NoError(t, err)
	assert.Equal(t, "custom", result)
}

// CustomRetryableOperation 自定义可重试操作实现
type CustomRetryableOperation[T any] struct {
	operation   func() (T, error)
	isRetryable func(error) bool
	config      RetryConfig
}

func (op *CustomRetryableOperation[T]) Execute(ctx context.Context) (T, error) {
	return op.operation()
}

func (op *CustomRetryableOperation[T]) IsRetryable(err error) bool {
	return op.isRetryable(err)
}

func (op *CustomRetryableOperation[T]) GetRetryConfig() RetryConfig {
	return op.config
}
