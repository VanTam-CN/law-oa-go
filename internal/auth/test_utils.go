package auth

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/test/mock"
)

// createTestCacheService 创建测试用的缓存服务
func createTestCacheService() *cache.CacheService {
	// 创建一个测试用的缓存服务
	cacheService, err := cache.NewCacheServiceWithConfig()
	if err != nil {
		// 如果Redis不可用，返回nil，测试应该跳过依赖Redis的功能
		return nil
	}
	return cacheService
}

// createTestRedisClient 创建测试用的Redis客户端
func createTestRedisClient() *redis.Client {
	// 创建一个测试用的Redis客户端
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

// createTestDB 创建测试用的数据库连接
func createTestDB() *gorm.DB {
	// 创建一个测试用的数据库连接
	mockDB, err := mock.NewMockDB()
	if err != nil {
		// 如果数据库不可用，返回nil，测试应该跳过依赖数据库的功能
		return nil
	}
	return mockDB.DB
}

// createTestConfig 创建测试配置
func createTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:    "test-secret-key-12345678901234567890",
			ExpiresIn: 3600,
			RefreshIn: 86400,
		},
	}
}
