package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultMySQLConfig(t *testing.T) {
	config := DefaultMySQLConfig()

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 3306, config.Port)
	assert.Equal(t, "root", config.Username)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "document_service", config.Database)
	assert.Equal(t, "utf8mb4", config.Charset)
	assert.True(t, config.ParseTime)
	assert.Equal(t, "Local", config.Loc)
	assert.Equal(t, 10, config.MaxIdleConns)
	assert.Equal(t, 100, config.MaxOpenConns)
	assert.Equal(t, 3600, config.ConnMaxLifetime)
	assert.Equal(t, 600, config.ConnMaxIdleTime)
}

func TestMySQLConfig_GetDSN(t *testing.T) {
	config := &MySQLConfig{
		Host:     "test-host",
		Port:     3307,
		Username: "test-user",
		Password: "test-pass",
		Database: "test-db",
		Charset:  "utf8",
		ParseTime: true,
		Loc:      "UTC",
	}

	expected := "test-user:test-pass@tcp(test-host:3307)/test-db?charset=utf8&parseTime=true&loc=UTC"
	assert.Equal(t, expected, config.GetDSN())
}

func TestDatabase_NewDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &MySQLConfig{
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "document_service",
		Charset:  "utf8mb4",
		ParseTime: true,
		Loc:      "Local",
		MaxIdleConns: 5,
		MaxOpenConns: 10,
	}

	// 这个测试需要真实的MySQL连接，在CI环境中可能失败
	t.Run("valid config", func(t *testing.T) {
		db, err := NewDatabase(config)
		if err != nil {
			t.Skipf("Skipping test due to connection error: %v", err)
		}
		defer db.Close()

		assert.NotNil(t, db)
		assert.NotNil(t, db.DB)
	})

	t.Run("invalid config", func(t *testing.T) {
		invalidConfig := &MySQLConfig{
			Host:     "invalid-host",
			Port:     9999,
			Username: "invalid-user",
			Password: "invalid-pass",
			Database: "invalid-db",
		}

		db, err := NewDatabase(invalidConfig)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestDatabase_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultMySQLConfig()
	db, err := NewDatabase(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err)
}

func TestDatabase_Health(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultMySQLConfig()
	db, err := NewDatabase(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer db.Close()

	err = db.Health()
	assert.NoError(t, err)
}

func TestDatabase_Stats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultMySQLConfig()
	db, err := NewDatabase(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer db.Close()

	stats := db.Stats()
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "max_open_connections")
	assert.Contains(t, stats, "open_connections")
	assert.Contains(t, stats, "in_use")
	assert.Contains(t, stats, "idle")
}

func TestDatabase_AutoMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultMySQLConfig()
	db, err := NewDatabase(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer db.Close()

	// 定义一个测试模型
	type TestModel struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100;not null"`
		Value string `gorm:"size:255"`
	}

	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// 验证表是否创建成功
	gormDB := db.GetDB()
	result := gormDB.Exec("SHOW TABLES LIKE 'test_models'")
	assert.NoError(t, result.Error)
	assert.Greater(t, result.RowsAffected, int64(0))
}

func TestDatabase_BeginTx(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultMySQLConfig()
	db, err := NewDatabase(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// 执行一些操作
	result := tx.Exec("SELECT 1")
	assert.NoError(t, result.Error)

	// 回滚事务
	err = tx.Rollback().Error
	assert.NoError(t, err)
}

// Benchmark database operations
func BenchmarkDatabase_Ping(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration test in short mode")
	}

	config := DefaultMySQLConfig()
	db, err := NewDatabase(config)
	if err != nil {
		b.Skipf("Skipping benchmark due to connection error: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := db.Ping(ctx)
		if err != nil {
			b.Fatalf("Ping failed: %v", err)
		}
	}
}

func BenchmarkDatabase_Stats(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration test in short mode")
	}

	config := DefaultMySQLConfig()
	db, err := NewDatabase(config)
	if err != nil {
		b.Skipf("Skipping benchmark due to connection error: %v", err)
	}
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stats := db.Stats()
		if stats == nil {
			b.Fatal("Stats returned nil")
		}
	}
}

// Example usage
func ExampleDatabase() {
	config := DefaultMySQLConfig()

	// 创建数据库实例
	db, err := NewDatabase(config)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 健康检查
	err = db.Health()
	if err != nil {
		panic(err)
	}

	// 获取统计信息
	stats := db.Stats()
	_ = stats

	// 自动迁移
	type User struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100;not null"`
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		panic(err)
	}

	// 获取GORM实例
	gormDB := db.GetDB()
	_ = gormDB
}