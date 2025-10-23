package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config.MySQL)
	assert.NotNil(t, config.Redis)
	assert.NotNil(t, config.Elasticsearch)

	// 验证默认值
	assert.Equal(t, "localhost", config.MySQL.Host)
	assert.Equal(t, 3306, config.MySQL.Port)
	assert.Equal(t, "document_service", config.MySQL.Database)

	assert.Equal(t, "localhost", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)

	assert.Equal(t, []string{"http://localhost:9200"}, config.Elasticsearch.Addresses)
}

func TestManager_NewManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()

	t.Run("all services available", func(t *testing.T) {
		manager, err := NewManager(config)
		if err != nil {
			t.Skipf("Skipping test due to connection error: %v", err)
		}
		defer manager.Close()

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.MySQL)
		assert.NotNil(t, manager.Redis)
		assert.NotNil(t, manager.Elasticsearch)
		assert.NotNil(t, manager.Cache)
	})

	t.Run("partial config", func(t *testing.T) {
		partialConfig := &Config{
			MySQL: DefaultMySQLConfig(),
			// Redis 和 Elasticsearch 为 nil
		}

		manager, err := NewManager(partialConfig)
		if err != nil {
			t.Skipf("Skipping test due to MySQL connection error: %v", err)
		}
		defer manager.Close()

		assert.NotNil(t, manager.MySQL)
		assert.Nil(t, manager.Redis)
		assert.Nil(t, manager.Elasticsearch)
		assert.Nil(t, manager.Cache)
	})
}

func TestManager_Health(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer manager.Close()

	health := manager.Health()
	assert.NotNil(t, health)

	// 检查各个服务的健康状态
	for service, err := range health {
		if err != nil {
			t.Logf("Service %s is unhealthy: %v", service, err)
		} else {
			t.Logf("Service %s is healthy", service)
		}
	}
}

func TestManager_Stats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer manager.Close()

	stats := manager.Stats()
	assert.NotNil(t, stats)

	// 验证包含各种服务的统计信息
	if manager.MySQL != nil {
		assert.Contains(t, stats, "mysql")
		mysqlStats := stats["mysql"].(map[string]interface{})
		assert.Contains(t, mysqlStats, "open_connections")
	}

	if manager.Redis != nil {
		assert.Contains(t, stats, "redis")
		redisStats := stats["redis"].(map[string]interface{})
		assert.Contains(t, redisStats, "type")
	}

	if manager.Elasticsearch != nil {
		assert.Contains(t, stats, "elasticsearch")
	}
}

func TestManager_Getters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer manager.Close()

	// 测试各种getter方法
	mysql := manager.GetMySQL()
	assert.Equal(t, manager.MySQL, mysql)

	redis := manager.GetRedis()
	assert.Equal(t, manager.Redis, redis)

	es := manager.GetElasticsearch()
	assert.Equal(t, manager.Elasticsearch, es)

	cache := manager.GetCache()
	assert.Equal(t, manager.Cache, cache)

	db := manager.GetDB()
	if manager.MySQL != nil {
		assert.Equal(t, manager.MySQL.GetDB(), db)
	} else {
		assert.Nil(t, db)
	}
}

func TestManager_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	err = manager.Close()
	assert.NoError(t, err)
}

func TestManager_AutoMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer manager.Close()

	if manager.MySQL == nil {
		t.Skip("MySQL not available, skipping migration test")
	}

	// 定义测试模型
	type TestModel struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100;not null"`
		Value string `gorm:"size:255"`
	}

	err = manager.AutoMigrate(&TestModel{})
	assert.NoError(t, err)

	// 验证表是否创建成功
	db := manager.GetDB()
	result := db.Exec("SHOW TABLES LIKE 'test_models'")
	assert.NoError(t, result.Error)
	assert.Greater(t, result.RowsAffected, int64(0))
}

func TestManager_InitializeElasticsearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer manager.Close()

	if manager.Elasticsearch == nil {
		t.Skip("Elasticsearch not available, skipping ES initialization test")
	}

	ctx := context.Background()
	err = manager.InitializeElasticsearch(ctx)
	assert.NoError(t, err)

	// 验证索引是否创建成功
	exists, err := manager.Elasticsearch.IndexExists(ctx, "documents")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestHealthChecker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer manager.Close()

	checker := NewHealthChecker(manager, 2*time.Second)

	// 启动健康检查
	checker.Start()

	// 等待一段时间让健康检查运行
	time.Sleep(3 * time.Second)

	// 停止健康检查
	checker.Stop()
}

func TestTransactionManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}
	defer manager.Close()

	if manager.MySQL == nil {
		t.Skip("MySQL not available, skipping transaction test")
	}

	txManager := NewTransactionManager(manager.GetDB())
	ctx := context.Background()

	// 定义测试模型
	type TestTransaction struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100;not null"`
		Value string `gorm:"size:255"`
	}

	// 自动迁移
	err = manager.AutoMigrate(&TestTransaction{})
	require.NoError(t, err)

	t.Run("successful transaction", func(t *testing.T) {
		err := txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
			// 创建记录
			record := TestTransaction{
				Name:  "test-name",
				Value: "test-value",
			}
			return tx.Create(&record).Error
		})
		assert.NoError(t, err)

		// 验证记录存在
		var count int64
		manager.GetDB().Model(&TestTransaction{}).Count(&count)
		assert.Greater(t, count, int64(0))
	})

	t.Run("failed transaction", func(t *testing.T) {
		err := txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
			// 创建记录
			record := TestTransaction{
				Name:  "test-name-2",
				Value: "test-value-2",
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}

			// 故意返回错误
			return assert.AnError
		})
		assert.Error(t, err)

		// 验证记录不存在
		var count int64
		manager.GetDB().Model(&TestTransaction{}).Where("name = ?", "test-name-2").Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("manual transaction control", func(t *testing.T) {
		tx, err := txManager.BeginTx(ctx)
		require.NoError(t, err)

		// 创建记录
		record := TestTransaction{
			Name:  "test-name-3",
			Value: "test-value-3",
		}
		err = tx.Create(&record).Error
		assert.NoError(t, err)

		// 手动回滚
		err = tx.Rollback().Error
		assert.NoError(t, err)

		// 验证记录不存在
		var count int64
		manager.GetDB().Model(&TestTransaction{}).Where("name = ?", "test-name-3").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

// Benchmark tests
func BenchmarkManager_Health(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		b.Skipf("Skipping benchmark due to connection error: %v", err)
	}
	defer manager.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		health := manager.Health()
		if health == nil {
			b.Fatal("Health returned nil")
		}
	}
}

func BenchmarkManager_Stats(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		b.Skipf("Skipping benchmark due to connection error: %v", err)
	}
	defer manager.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stats := manager.Stats()
		if stats == nil {
			b.Fatal("Stats returned nil")
		}
	}
}

// Example usage
func ExampleManager() {
	config := DefaultConfig()

	// 创建数据库管理器
	manager, err := NewManager(config)
	if err != nil {
		panic(err)
	}
	defer manager.Close()

	// 健康检查
	health := manager.Health()
	for service, err := range health {
		if err != nil {
			fmt.Printf("Service %s is unhealthy: %v\n", service, err)
		} else {
			fmt.Printf("Service %s is healthy\n", service)
		}
	}

	// 获取统计信息
	stats := manager.Stats()
	fmt.Printf("Database stats: %+v\n", stats)

	// 使用缓存
	cache := manager.GetCache()
	ctx := context.Background()
	err = cache.Set(ctx, "test-key", "test-value", time.Minute)
	if err != nil {
		panic(err)
	}

	var value string
	err = cache.Get(ctx, "test-key", &value)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cached value: %s\n", value)

	// 事务管理
	txManager := NewTransactionManager(manager.GetDB())
	err = txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		// 在事务中执行操作
		return nil
	})
	if err != nil {
		panic(err)
	}
}