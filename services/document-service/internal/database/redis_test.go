package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRedisConfig(t *testing.T) {
	config := DefaultRedisConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 6379, config.Port)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, 0, config.Database)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 5, config.MinIdleConns)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5, config.DialTimeout)
	assert.Equal(t, 3, config.ReadTimeout)
	assert.Equal(t, 3, config.WriteTimeout)
	assert.Equal(t, 5, config.PoolTimeout)
	assert.Equal(t, 300, config.IdleTimeout)
	assert.False(t, config.IsCluster)
	assert.Empty(t, config.ClusterNodes)
}

func TestRedisConfig_GetAddr(t *testing.T) {
	config := &RedisConfig{
		Host: "test-host",
		Port: 6380,
	}

	expected := "test-host:6380"
	assert.Equal(t, expected, config.GetAddr())
}

func TestRedisClient_NewRedisClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()

	t.Run("valid config", func(t *testing.T) {
		client, err := NewRedisClient(config)
		if err != nil {
			t.Skipf("Skipping test due to connection error: %v", err)
		}
		defer client.Close()

		assert.NotNil(t, client)
		assert.NotNil(t, client.Client)
	})

	t.Run("invalid config", func(t *testing.T) {
		invalidConfig := &RedisConfig{
			Host: "invalid-host",
			Port: 9999,
		}

		client, err := NewRedisClient(invalidConfig)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestRedisClient_Health(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer client.Close()

	err = client.Health()
	assert.NoError(t, err)
}

func TestRedisClient_Stats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer client.Close()

	stats := client.Stats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "type")
	assert.Equal(t, "single", stats["type"])
	assert.Contains(t, stats, "total_conns")
	assert.Contains(t, stats, "idle_conns")
}

func TestRedisClient_SetAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-key-" + time.Now().Format("20060102150405")

	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestStruct{
		Name:  "test",
		Value: 123,
	}

	// 设置值
	err = client.Set(ctx, key, original, time.Minute)
	require.NoError(t, err)

	// 获取值
	var retrieved TestStruct
	err = client.Get(ctx, key, &retrieved)
	require.NoError(t, err)

	assert.Equal(t, original.Name, retrieved.Name)
	assert.Equal(t, original.Value, retrieved.Value)

	// 清理
	_ = client.Delete(ctx, key)
}

func TestRedisClient_SetWithTTLAndGetWithTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-ttl-" + time.Now().Format("20060102150405")
	value := "test-value"
	ttl := 2 * time.Second

	// 设置带TTL的值
	err = client.SetWithTTL(ctx, key, value, ttl)
	require.NoError(t, err)

	// 获取值和TTL
	var retrieved string
	retrievedTTL, err := client.GetWithTTL(ctx, key, &retrieved)
	require.NoError(t, err)

	assert.Equal(t, value, retrieved)
	assert.True(t, retrievedTTL > 0 && retrievedTTL <= ttl)

	// 等待过期
	time.Sleep(ttl + 100*time.Millisecond)

	// 验证已过期
	err = client.Get(ctx, key, &retrieved)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestRedisClient_ExistsAndDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-exists-" + time.Now().Format("20060102150405")
	value := "test-value"

	// 初始状态不存在
	exists, err := client.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 设置值
	err = client.Set(ctx, key, value, time.Minute)
	require.NoError(t, err)

	// 现在应该存在
	exists, err = client.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 删除
	err = client.Delete(ctx, key)
	assert.NoError(t, err)

	// 再次检查应该不存在
	exists, err = client.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisClient_IncrementAndDecrement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-counter-" + time.Now().Format("20060102150405")

	// 删除可能存在的键
	_ = client.Delete(ctx, key)

	// 初始值为0
	val, err := client.Increment(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// 再次递增
	val, err = client.Increment(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), val)

	// 递减
	val, err = client.Decrement(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// 清理
	_ = client.Delete(ctx, key)
}

func TestRedisClient_SetNX(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-setnx-" + time.Now().Format("20060102150405")
	value := "test-value"

	// 第一次设置应该成功
	success, err := client.SetNX(ctx, key, value, time.Minute)
	assert.NoError(t, err)
	assert.True(t, success)

	// 第二次设置应该失败
	success, err = client.SetNX(ctx, key, "new-value", time.Minute)
	assert.NoError(t, err)
	assert.False(t, success)

	// 获取值应该是原始值
	var retrieved string
	err = client.Get(ctx, key, &retrieved)
	assert.NoError(t, err)
	assert.Equal(t, value, retrieved)

	// 清理
	_ = client.Delete(ctx, key)
}

func TestCacheManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	redisClient, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer redisClient.Close()

	cache := NewCacheManager(redisClient, "test-cache")
	ctx := context.Background()

	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	original := TestStruct{
		ID:   123,
		Name: "test-name",
	}

	// 设置缓存
	err = cache.Set(ctx, "test-key", original, time.Minute)
	require.NoError(t, err)

	// 获取缓存
	var retrieved TestStruct
	err = cache.Get(ctx, "test-key", &retrieved)
	require.NoError(t, err)

	assert.Equal(t, original.ID, retrieved.ID)
	assert.Equal(t, original.Name, retrieved.Name)

	// 检查是否存在
	exists, err := cache.Exists(ctx, "test-key")
	assert.NoError(t, err)
	assert.True(t, exists)

	// 删除缓存
	err = cache.Delete(ctx, "test-key")
	assert.NoError(t, err)

	// 再次检查应该不存在
	exists, err = cache.Exists(ctx, "test-key")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestCacheManager_ClearPattern(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	redisClient, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer redisClient.Close()

	cache := NewCacheManager(redisClient, "pattern-test")
	ctx := context.Background()

	// 设置多个键
	keys := []string{"user:1", "user:2", "session:1", "session:2"}
	for _, key := range keys {
		err := cache.Set(ctx, key, "value", time.Minute)
		require.NoError(t, err)
	}

	// 清除匹配模式的键
	err = cache.ClearPattern(ctx, "user:*")
	assert.NoError(t, err)

	// 验证匹配的键被删除
	exists, err := cache.Exists(ctx, "user:1")
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = cache.Exists(ctx, "user:2")
	assert.NoError(t, err)
	assert.False(t, exists)

	// 验证不匹配的键仍然存在
	exists, err = cache.Exists(ctx, "session:1")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = cache.Exists(ctx, "session:2")
	assert.NoError(t, err)
	assert.True(t, exists)

	// 清理
	err = cache.ClearPattern(ctx, "*")
	assert.NoError(t, err)
}

// Benchmarks
func BenchmarkRedisClient_Set(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		b.Skipf("Skipping benchmark due to connection error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := "bench-key"
		value := "bench-value"
		err := client.Set(ctx, key, value, time.Minute)
		if err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkRedisClient_Get(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration test in short mode")
	}

	config := DefaultRedisConfig()
	client, err := NewRedisClient(config)
	if err != nil {
		b.Skipf("Skipping benchmark due to connection error: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	key := "bench-get-key"
	value := "bench-get-value"

	// 预先设置值
	err = client.Set(ctx, key, value, time.Minute)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	var retrieved string
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := client.Get(ctx, key, &retrieved)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}