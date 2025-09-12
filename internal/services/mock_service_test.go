package services_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockService 模拟服务，用于测试
type MockService struct {
	data sync.Map
}

func NewMockService() *MockService {
	return &MockService{}
}

func (s *MockService) Add(key, value string) {
	s.data.Store(key, value)
}

func (s *MockService) Get(key string) (string, bool) {
	value, exists := s.data.Load(key)
	if !exists {
		return "", false
	}
	return value.(string), true
}

func (s *MockService) Delete(key string) {
	s.data.Delete(key)
}

func (s *MockService) Count() int {
	count := 0
	s.data.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func TestMockService(t *testing.T) {
	service := NewMockService()

	// 测试添加和获取
	service.Add("test", "value")
	value, exists := service.Get("test")
	assert.True(t, exists, "键应该存在")
	assert.Equal(t, "value", value, "值应该正确")

	// 测试计数
	assert.Equal(t, 1, service.Count(), "计数应该为1")

	// 测试删除
	service.Delete("test")
	_, exists = service.Get("test")
	assert.False(t, exists, "键应该被删除")
	assert.Equal(t, 0, service.Count(), "计数应该为0")
}

func TestMockService_ConcurrentAccess(t *testing.T) {
	service := NewMockService()
	var wg sync.WaitGroup

	// 并发添加数据
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			key := "key" + string(rune('0'+index%10)) // 使用模运算避免key冲突
			service.Add(key, "value"+string(rune('0'+index%10)))
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 验证数据（由于key冲突，实际数量可能少于100）
	count := service.Count()
	assert.GreaterOrEqual(t, count, 1, "至少应该有一些数据")
	assert.LessOrEqual(t, count, 10, "最多应该有10个不同的key")
}

func TestMockService_InvalidOperations(t *testing.T) {
	service := NewMockService()

	// 测试获取不存在的键
	_, exists := service.Get("nonexistent")
	assert.False(t, exists, "不存在的键应该返回false")

	// 测试删除不存在的键（不应该panic）
	service.Delete("nonexistent")
	assert.Equal(t, 0, service.Count(), "计数应该为0")
}
