package auth

import (
	"sync"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/test/mock"
)

var (
	testRedisOnce   sync.Once
	testRedisServer *miniredis.Miniredis
	testRedisAddr   string
)

func sharedTestRedisAddr() string {
	testRedisOnce.Do(func() {
		server, err := miniredis.Run()
		if err != nil {
			panic("failed to start isolated test Redis: " + err.Error())
		}
		testRedisServer = server
		testRedisAddr = server.Addr()
	})
	return testRedisAddr
}

// createTestCacheService 创建测试用的缓存服务
func createTestCacheService() *cache.CacheService {
	// 创建一个测试用的缓存服务
	redisClient := redis.NewClient(&redis.Options{Addr: sharedTestRedisAddr()})
	return cache.NewCacheService(redisClient, "test")
}

// createTestRedisClient 创建测试用的Redis客户端
func createTestRedisClient() *redis.Client {
	// 创建一个测试用的Redis客户端
	return redis.NewClient(&redis.Options{
		Addr:     sharedTestRedisAddr(),
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
