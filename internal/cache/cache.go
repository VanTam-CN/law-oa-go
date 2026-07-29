package cache

import (
	"context"
	"fmt"
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

	// 简单的类型处理，实际项目中可能需要更复杂的序列化/反序列化
	switch v := dest.(type) {
	case *string:
		*v = val
	case *[]byte:
		*v = []byte(val)
	default:
		return fmt.Errorf("unsupported destination type")
	}

	return nil
}

// Set 设置缓存值
func (c *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	fullKey := c.buildKey(key)
	ctx := context.Background()
	return c.client.Set(ctx, fullKey, value, ttl).Err()
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

// InitCache 初始化缓存服务
func InitCache() error {
	// 这里会从数据库模块获取Redis客户端
	// 具体实现在database模块中
	return nil
}
