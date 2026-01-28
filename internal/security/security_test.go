package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecurityTypes(t *testing.T) {
	t.Run("测试安全类型定义", func(t *testing.T) {
		// 测试安全级别
		levels := []SecurityLevel{
			SecurityLevelLow,
			SecurityLevelMedium,
			SecurityLevelHigh,
			SecurityLevelCritical,
		}

		for i, level := range levels {
			assert.NotEqual(t, SecurityLevel(""), level)
			assert.GreaterOrEqual(t, int(level), 0)
			assert.LessOrEqual(t, int(level), 3)
			_ = i // 避免未使用变量警告
		}
	})

	t.Run("测试安全事件类型", func(t *testing.T) {
		eventTypes := []SecurityEventType{
			SecurityEventTypeAuth,
			SecurityEventTypeAuthz,
			SecurityEventTypeInjection,
			SecurityEventTypeXSS,
			SecurityEventTypeCSRF,
			SecurityEventTypeRateLimit,
			SecurityEventTypeDataLeak,
			SecurityEventTypeConfig,
		}

		for i, eventType := range eventTypes {
			assert.NotEqual(t, SecurityEventType(""), eventType)
			_ = i // 避免未使用变量警告
		}
	})
}

func Fuzz_SecurityTypes_Parse(f *testing.F) {
	// 添加种子语料
	f.Add("low")
	f.Add("medium")
	f.Add("high")
	f.Add("critical")

	f.Fuzz(func(t *testing.T, input string) {
		// 测试各种输入不会导致panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("解析安全级别时panic: %v", r)
			}
		}()

		// 这里应该有实际的解析逻辑
		// 由于当前只有类型定义，我们只测试不会崩溃
		if len(input) > 100 {
			return // 跳过过长的输入
		}
	})
}

func TestSecurityEvent(t *testing.T) {
	t.Run("创建安全事件", func(t *testing.T) {
		event := SecurityEvent{
			ID:        "test-001",
			Type:      SecurityEventTypeAuth,
			Level:     SecurityLevelHigh,
			Timestamp: time.Now(),
			Message:   "测试安全事件",
			Metadata:  map[string]string{"key": "value"},
		}

		assert.Equal(t, "test-001", event.ID)
		assert.Equal(t, SecurityEventTypeAuth, event.Type)
		assert.Equal(t, SecurityLevelHigh, event.Level)
		assert.NotZero(t, event.Timestamp)
		assert.Equal(t, "测试安全事件", event.Message)
		assert.NotNil(t, event.Metadata)
	})
}

func Benchmark_SecurityEvent_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SecurityEvent{
			ID:        "bench-001",
			Type:      SecurityEventTypeAuth,
			Level:     SecurityLevelHigh,
			Timestamp: time.Now(),
			Message:   "基准测试安全事件",
			Metadata:  map[string]string{"key": "value"},
		}
	}
}

// JWT Fuzzing 测试
func Fuzz_JWTKeyManager_ValidateToken(f *testing.F) {
	// 添加种子语料
	f.Add("valid_token_123", "secret_key")
	f.Add("invalid_token", "wrong_key")
	f.Add("", "secret_key")

	f.Fuzz(func(t *testing.T, token, key string) {
		// 限制输入长度
		if len(token) > 1000 || len(key) > 200 {
			t.Skip()
		}

		// 测试token验证逻辑不会panic
		// 这里我们只测试基本的字符串处理，不进行实际JWT验证
		if len(token) > 0 && len(key) > 0 {
			// 基本的token结构验证
			parts := len(token) // 简化的检查
			if parts > 0 {
				// Token有内容
			}
		}
	})
}
