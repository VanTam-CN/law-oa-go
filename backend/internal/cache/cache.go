package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService 缓存服务
type CacheService struct {
	client *redis.Client
	prefix string
}

// NewCacheService 创建新的缓存服务
func NewCacheService(client *redis.Client, prefix string) *CacheService {
	return &CacheService{
		client: client,
		prefix: prefix,
	}
}

// Get 获取缓存值
func (c *CacheService) Get(key string, dest interface{}) error {
	fullKey := c.buildKey(key)
	ctx := context.Background()
	val, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		return err
	}

	return decodeInto(val, dest)
}

// decodeInto parses one cached value into dest. The destination is
// validated as a non-nil pointer before any dereference, so nil
// *string and *[]byte destinations fail closed instead of panicking
// in the raw passthrough cases. Raw strings and bytes keep the
// pre-change behavior; every other destination is decoded as JSON
// written by Set. The stored value must be valid JSON, otherwise the
// read fails closed.
func decodeInto(val string, dest interface{}) error {
	if dest == nil || reflect.ValueOf(dest).Kind() != reflect.Ptr || reflect.ValueOf(dest).IsNil() {
		return fmt.Errorf("cache: destination must be a non-nil pointer")
	}
	switch v := dest.(type) {
	case *string:
		*v = val
		return nil
	case *[]byte:
		*v = []byte(val)
		return nil
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("cache: decode cached value: %w", err)
	}
	return nil
}

// isNullValue reports whether value is untyped nil or a typed nil
// (pointer, map, or slice) that encoding/json would serialize as the
// literal "null". Such values cannot round-trip into a usable read,
// so Set rejects them before writing.
func isNullValue(value interface{}) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// Set 设置缓存值
func (c *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	fullKey := c.buildKey(key)
	ctx := context.Background()
	// Strings and bytes stay raw for backward compatibility with every key
	// written before this change. Structured values are JSON so the auth
	// device map and rate-limit counters round-trip. Untyped and typed nil
	// values would serialize as the literal "null" and cannot round-trip,
	// so they fail closed with a clear error.
	switch v := value.(type) {
	case string:
		return c.client.Set(ctx, fullKey, v, ttl).Err()
	case []byte:
		return c.client.Set(ctx, fullKey, v, ttl).Err()
	}
	if isNullValue(value) {
		return fmt.Errorf("cache: unsupported value type %T", value)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: unsupported value type %T: %w", value, err)
	}
	return c.client.Set(ctx, fullKey, encoded, ttl).Err()
}

// Delete 删除缓存
func (c *CacheService) Delete(key string) error {
	fullKey := c.buildKey(key)
	ctx := context.Background()
	return c.client.Del(ctx, fullKey).Err()
}

// ClearPattern 清除匹配模式的缓存
func (c *CacheService) ClearPattern(pattern string) error {
	fullPattern := c.buildKey(pattern)
	ctx := context.Background()
	keys, err := c.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}

// GetClient 获取Redis客户端
func (c *CacheService) GetClient() *redis.Client {
	return c.client
}

// Ping 测试Redis连接
func (c *CacheService) Ping() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	return err == nil
}

// buildKey 构建完整的缓存键
func (c *CacheService) buildKey(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}

// DefaultCacheService 默认缓存服务实例
var DefaultCacheService *CacheService

// InitCache is retained for source compatibility. Redis is optional, so a nil
// default cache service is a valid degraded state.
func InitCache() error {
	return nil
}
