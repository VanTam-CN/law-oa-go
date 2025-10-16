package repositories

import (
	"database/sql"
	"log"

	"gorm.io/gorm"
	redis "github.com/redis/go-redis/v9"
)

// DBAdapter 数据库适配器，用于处理不同类型的数据库连接
type DBAdapter struct {
	sqlDB *sql.DB
	gormDB *gorm.DB
}

// NewDBAdapter 创建数据库适配器
func NewDBAdapter(db interface{}) *DBAdapter {
	adapter := &DBAdapter{}

	switch v := db.(type) {
	case *sql.DB:
		adapter.sqlDB = v
	case *gorm.DB:
		adapter.gormDB = v
		// 尝试获取底层的sql.DB
		if sqlDB, err := v.DB(); err == nil {
			adapter.sqlDB = sqlDB
		} else {
			log.Printf("无法转换gorm.DB到sql.DB: %v", err)
		}
	default:
		log.Printf("不支持的数据库类型: %T", db)
	}

	return adapter
}

// GetSQLDB 获取SQL数据库连接
func (a *DBAdapter) GetSQLDB() *sql.DB {
	return a.sqlDB
}

// GetGormDB 获取GORM数据库连接
func (a *DBAdapter) GetGormDB() *gorm.DB {
	return a.gormDB
}

// ToGormDB 转换为GORM数据库连接
func ToGormDB(db interface{}) *gorm.DB {
	if gormDB, ok := db.(*gorm.DB); ok {
		return gormDB
	}
	log.Printf("无法将 %T 转换为 *gorm.DB", db)
	return nil
}

// ToSQLDB 转换为SQL数据库连接
func ToSQLDB(db interface{}) *sql.DB {
	if sqlDB, ok := db.(*sql.DB); ok {
		return sqlDB
	}
	if gormDB, ok := db.(*gorm.DB); ok {
		if sqlDB, err := gormDB.DB(); err == nil {
			return sqlDB
		}
	}
	if _, ok := db.(*redis.Client); ok {
		// 对于Redis客户端，我们不能转换为*sql.DB，返回nil
		log.Printf("Redis客户端无法转换为 *sql.DB")
		return nil
	}
	log.Printf("无法将 %T 转换为 *sql.DB", db)
	return nil
}

// ToRedisClient 转换为Redis客户端连接
func ToRedisClient(r interface{}) *redis.Client {
	if redisClient, ok := r.(*redis.Client); ok {
		return redisClient
	}
	log.Printf("无法将 %T 转换为 *redis.Client", r)
	return nil
}