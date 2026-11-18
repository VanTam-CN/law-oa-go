package concurrency

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test_breaker", 5, 30*time.Second)

	assert.NotNil(t, cb)
	assert.Equal(t, "test_breaker", cb.name)
	assert.Equal(t, 5, cb.maxFailures)
	assert.Equal(t, 30*time.Second, cb.resetTimeout)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailures())
}

func TestCircuitBreaker_ExecuteSuccess(t *testing.T) {
	cb := NewCircuitBreaker("test_breaker", 3, 1*time.Second)

	// 执行成功操作
	err := cb.Execute(context.Background(), func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailures())
}

func TestCircuitBreaker_ExecuteFailure(t *testing.T) {
	cb := NewCircuitBreaker("test_breaker", 3, 1*time.Second)

	// 执行失败操作
	testErr := fmt.Errorf("operation failed")
	err := cb.Execute(context.Background(), func() error {
		return testErr
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, testErr)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 1, cb.GetFailures())
}

func TestCircuitBreaker_TripToOpen(t *testing.T) {
	cb := NewCircuitBreaker("test_breaker", 3, 100*time.Millisecond)

	// 连续失败直到熔断器开启
	testErr := fmt.Errorf("operation failed")
	for i := 0; i < 3; i++ {
		err := cb.Execute(context.Background(), func() error {
			return testErr
		})
		assert.Error(t, err)
	}

	// 现在熔断器应该开启
	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 3, cb.GetFailures())

	// 当熔断器开启时，操作应该被拒绝
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "test_breaker is open")
}

func TestCircuitBreaker_ResetTimeout(t *testing.T) {
	cb := NewCircuitBreaker("test_breaker", 3, 100*time.Millisecond)

	// 触发熔断
	testErr := fmt.Errorf("operation failed")
	for i := 0; i < 3; i++ {
		_ = cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	assert.Equal(t, StateOpen, cb.GetState())

	// 等待重置超时
	time.Sleep(150 * time.Millisecond)

	// 再次调用应该转为半开状态
	err := cb.Execute(context.Background(), func() error {
		return nil // 这次成功
	})
	assert.NoError(t, err)

	// 成功后应该重置为关闭状态
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailures())
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker("test_breaker", 3, 100*time.Millisecond)

	// 触发熔断
	testErr := fmt.Errorf("operation failed")
	for i := 0; i < 3; i++ {
		_ = cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	assert.Equal(t, StateOpen, cb.GetState())

	// 等待重置超时
	time.Sleep(150 * time.Millisecond)

	// 在半开状态下失败
	err := cb.Execute(context.Background(), func() error {
		return testErr
	})
	assert.Error(t, err)

	// 应该回到开启状态
	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 4, cb.GetFailures())
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 20) // 每秒10个请求，桶容量20

	assert.NotNil(t, rl)
	assert.Equal(t, 10, rl.rate)
	assert.Equal(t, 20, rl.capacity)
	assert.Equal(t, 20, rl.GetTokens())
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(5, 5) // 每秒5个请求，桶容量5

	// 前5个请求应该被允许
	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow())
	}

	// 第6个请求应该被拒绝
	assert.False(t, rl.Allow())

	// 等待1秒后应该重新填充
	time.Sleep(1 * time.Second)
	assert.True(t, rl.Allow())
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(1, 1) // 每秒1个请求，桶容量1

	// 第一个请求应该立即通过
	assert.True(t, rl.Allow())

	// 第二个请求需要等待
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := rl.Wait(ctx)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, duration, time.Second-10*time.Millisecond) // 允许一些误差
	assert.Less(t, duration, 2*time.Second)
}

func TestRateLimiter_WaitContextTimeout(t *testing.T) {
	rl := NewRateLimiter(1, 1) // 每秒1个请求，桶容量1

	// 使用第一个令牌
	assert.True(t, rl.Allow())

	// 使用短超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 等待令牌应该超时
	err := rl.Wait(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	rl := NewRateLimiter(2, 2) // 每秒2个请求，桶容量2

	// 使用所有令牌
	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())
	assert.False(t, rl.Allow())

	// 等待1秒
	time.Sleep(1 * time.Second)

	// 应该有2个新令牌
	assert.True(t, rl.Allow())
	assert.True(t, rl.Allow())
	assert.False(t, rl.Allow())
}

func TestNewConcurrentSafe(t *testing.T) {
	// 测试只启用熔断器
	cs1 := NewConcurrentSafe(true, false, 0, 0)
	assert.True(t, cs1.IsEnabled())
	assert.NotNil(t, cs1.circuitBreaker)
	assert.Nil(t, cs1.rateLimiter)

	// 测试只启用速率限制
	cs2 := NewConcurrentSafe(false, true, 10, 20)
	assert.True(t, cs2.IsEnabled())
	assert.Nil(t, cs2.circuitBreaker)
	assert.NotNil(t, cs2.rateLimiter)

	// 测试都禁用
	cs3 := NewConcurrentSafe(false, false, 0, 0)
	assert.False(t, cs3.IsEnabled())
	assert.Nil(t, cs3.circuitBreaker)
	assert.Nil(t, cs3.rateLimiter)
}

func TestConcurrentSafe_ExecuteDisabled(t *testing.T) {
	cs := NewConcurrentSafe(false, false, 0, 0)

	// 执行操作应该直接调用原始函数
	result := ""
	err := cs.Execute(context.Background(), func() error {
		result = "executed"
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "executed", result)
}

func TestConcurrentSafe_ExecuteWithCircuitBreaker(t *testing.T) {
	cs := NewConcurrentSafe(true, false, 0, 0)

	// 执行成功操作
	result := ""
	err := cs.Execute(context.Background(), func() error {
		result = "success"
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// 检查熔断器状态
	state, failures := cs.GetCircuitBreakerState()
	assert.Equal(t, StateClosed, state)
	assert.Equal(t, 0, failures)
}

func TestConcurrentSafe_ExecuteWithRateLimiter(t *testing.T) {
	cs := NewConcurrentSafe(false, true, 2, 2) // 每秒2个请求

	// 前2个请求应该成功
	for i := 0; i < 2; i++ {
		err := cs.Execute(context.Background(), func() error {
			return nil
		})
		assert.NoError(t, err)
	}

	// 第3个请求应该被速率限制
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := cs.Execute(ctx, func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestConcurrentSafe_ExecuteWithBoth(t *testing.T) {
	cs := NewConcurrentSafe(true, true, 5, 5)

	// 执行几个成功操作
	for i := 0; i < 3; i++ {
		err := cs.Execute(context.Background(), func() error {
			return nil
		})
		assert.NoError(t, err)
	}

	// 检查状态
	state, failures := cs.GetCircuitBreakerState()
	assert.Equal(t, StateClosed, state)
	assert.Equal(t, 0, failures)

	tokens := cs.GetRateLimiterTokens()
	assert.Less(t, tokens, 5) // 应该用了一些令牌
}

func TestConcurrentSafe_ConcurrentExecution(t *testing.T) {
	cs := NewConcurrentSafe(false, true, 10, 10) // 速率限制器

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]int, 0, 20)

	// 并发执行多个操作
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			err := cs.Execute(context.Background(), func() error {
				mu.Lock()
				results = append(results, i)
				mu.Unlock()
				return nil
			})

			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// 所有操作都应该完成
	assert.Equal(t, 20, len(results))
}

func TestConcurrentSafe_GetStatusMethods(t *testing.T) {
	cs := NewConcurrentSafe(true, true, 5, 5)

	// 测试获取状态方法
	state, failures := cs.GetCircuitBreakerState()
	assert.Equal(t, StateClosed, state)
	assert.Equal(t, 0, failures)

	tokens := cs.GetRateLimiterTokens()
	assert.Equal(t, 5, tokens)

	isEnabled := cs.IsEnabled()
	assert.True(t, isEnabled)

	// 执行一些操作后再次检查
	for i := 0; i < 3; i++ {
		_ = cs.Execute(context.Background(), func() error {
			return nil
		})
	}

	// 令牌应该减少
	tokens = cs.GetRateLimiterTokens()
	assert.Less(t, tokens, 5)
}
