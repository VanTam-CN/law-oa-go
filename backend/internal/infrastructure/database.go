package infrastructure

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DatabaseProvider 数据库提供者接口
type DatabaseProvider interface {
	GetDB() *gorm.DB
	GetRedis() *redis.Client
}

// GlobalProvider 全局数据库提供者
var GlobalProvider DatabaseProvider

// SetGlobalProvider 设置全局数据库提供者
func SetGlobalProvider(provider DatabaseProvider) {
	GlobalProvider = provider
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	if GlobalProvider == nil {
		return nil
	}
	return GlobalProvider.GetDB()
}

// GetRedis 获取Redis连接
func GetRedis() *redis.Client {
	if GlobalProvider == nil {
		return nil
	}
	return GlobalProvider.GetRedis()
}
