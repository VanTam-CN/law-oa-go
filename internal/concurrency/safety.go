package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	name         string
	maxFailures  int
	resetTimeout time.Duration
	currentState int32
	failures     int
	lastFailTime time.Time
	mu           sync.RWMutex
}

const (
	StateClosed   int32 = 0 // 关闭状态，正常请求
	StateOpen     int32 = 1 // 开启状态，拒绝请求
	StateHalfOpen int32 = 2 // 半开状态，允许部分请求
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		currentState: StateClosed,
	}
}

// Execute 执行操作
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	if !cb.AllowRequest() {
		return fmt.Errorf("circuit breaker %s is open", cb.name)
	}

	// 检查是否需要从开启状态转换到半开状态
	cb.mu.Lock()
	if cb.currentState == StateOpen && time.Since(cb.lastFailTime) > cb.resetTimeout {
		cb.currentState = StateHalfOpen
		// 保持原有的失败计数，不要重置为0
	}
	cb.mu.Unlock()

	err := operation()
	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// AllowRequest 检查是否允许请求
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.currentState == StateClosed {
		return true
	}

	if cb.currentState == StateOpen {
		// 检查是否超时可以重置
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			return true // 允许请求尝试，实际状态转换在Execute中处理
		}
		return false
	}

	// 半开状态，允许部分请求通过
	return true
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.currentState = StateOpen
	}
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	if cb.currentState == StateHalfOpen {
		cb.currentState = StateClosed
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() int32 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.currentState
}

// GetFailures 获取失败次数
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// RateLimiter 速率限制器
type RateLimiter struct {
	rate       int       // 每秒请求数
	capacity   int       // 桶容量
	tokens     int       // 当前令牌数
	lastRefill time.Time // 上次填充时间
	mu         sync.Mutex
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rate int, capacity int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		capacity:   capacity,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 填充令牌
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	if elapsed >= time.Second {
		tokensToAdd := int(elapsed/time.Second) * rl.rate
		rl.tokens = min(rl.tokens+tokensToAdd, rl.capacity)
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// Wait 等待直到有足够的令牌
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for !rl.Allow() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// 等待一小段时间后重试
		}
	}
	return nil
}

// GetTokens 获取当前令牌数
func (rl *RateLimiter) GetTokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.tokens
}

// ConcurrentSafe 并发安全包装器
type ConcurrentSafe struct {
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
	enabled        bool
}

// NewConcurrentSafe 创建并发安全包装器
func NewConcurrentSafe(enableCircuitBreaker bool, enableRateLimiter bool, rate int, capacity int) *ConcurrentSafe {
	cs := &ConcurrentSafe{
		enabled: enableCircuitBreaker || enableRateLimiter,
	}

	if enableCircuitBreaker {
		cs.circuitBreaker = NewCircuitBreaker("default", 5, 30*time.Second)
	}

	if enableRateLimiter {
		cs.rateLimiter = NewRateLimiter(rate, capacity)
	}

	return cs
}

// Execute 安全执行操作
func (cs *ConcurrentSafe) Execute(ctx context.Context, operation func() error) error {
	if !cs.enabled {
		return operation()
	}

	// 速率限制
	if cs.rateLimiter != nil {
		if err := cs.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit exceeded: %w", err)
		}
	}

	// 熔断器
	if cs.circuitBreaker != nil {
		return cs.circuitBreaker.Execute(ctx, operation)
	}

	return operation()
}

// GetCircuitBreakerState 获取熔断器状态
func (cs *ConcurrentSafe) GetCircuitBreakerState() (int32, int) {
	if cs.circuitBreaker == nil {
		return StateClosed, 0
	}
	return cs.circuitBreaker.GetState(), cs.circuitBreaker.GetFailures()
}

// GetRateLimiterTokens 获取速率限制器令牌数
func (cs *ConcurrentSafe) GetRateLimiterTokens() int {
	if cs.rateLimiter == nil {
		return 0
	}
	return cs.rateLimiter.GetTokens()
}

// IsEnabled 检查是否启用
func (cs *ConcurrentSafe) IsEnabled() bool {
	return cs.enabled
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
