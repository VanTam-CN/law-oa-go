package security

import (
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/test/mock"
)

// createTestCacheService 创建测试用的缓存服务
func createTestCacheService() *cache.CacheService {
	// 对于配置测试，我们不需要真实的缓存服务，返回nil避免Redis连接问题
	// ConfigManager会处理nil缓存的情况
	return nil
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