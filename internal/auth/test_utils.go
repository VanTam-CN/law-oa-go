package auth

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa-go/test/mock"
)

// createTestCacheService 创建测试用的缓存服务
func createTestCacheService() *cache.CacheService {
	// 创建一个测试用的缓存服务
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return cache.NewCacheService(redisClient, "test")
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

func createAuthTokenDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL", filepath.Join(t.TempDir(), "auth-token-sessions.db"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.AuthTokenSession{}, &models.TokenRevocationLog{}); err != nil {
		t.Fatalf("migrate auth token models: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		// SQLite cannot upgrade concurrent read transactions to writers without
		// returning a spurious "database is locked" error. Serializing the test
		// connection keeps the assertion focused on refresh replay semantics;
		// production uses PostgreSQL row locks.
		sqlDB.SetMaxOpenConns(1)
	}
	return db
}
