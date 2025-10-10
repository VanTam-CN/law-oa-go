package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/monitoring"
	"law-oa-go/internal/router"
)

// setupTestRouter 设置测试路由器
func setupTestRouter() (*gin.Engine, *gorm.DB, *redis.Client) {
	gin.SetMode(gin.TestMode)

	// 内存数据库
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})

	// 内存Redis（使用redis-go的mock或实际Redis实例）
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用测试数据库
	})

	config := &router.RouterConfig{
		DB:             db,
		Redis:          rdb,
		Elasticsearch:  nil,
		AllowedOrigins: []string{"*"},
		RateLimit:      1000,
		Timeout:        30 * time.Second,
	}

	app := router.NewRouter(config)
	return app, db, rdb
}

// BenchmarkAPIRequest API请求性能基准测试
func BenchmarkAPIRequest(b *testing.B) {
	app, _, _ := setupTestRouter()

	// 测试不同的API端点
	testCases := []struct {
		name string
		path string
		method string
		body interface{}
	}{
		{
			name:   "GetDashboardStats",
			path:   "/api/dashboard/statistics",
			method: "GET",
		},
		{
			name:   "GetCases",
			path:   "/api/cases",
			method: "GET",
		},
		{
			name:   "GetClients",
			path:   "/api/clients",
			method: "GET",
		},
		{
			name:   "LoginRequest",
			path:   "/api/auth/login",
			method: "POST",
			body: map[string]interface{}{
				"email":    "test@example.com",
				"password": "testpassword",
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var req *http.Request
				if tc.body != nil {
					body, _ := json.Marshal(tc.body)
					req = httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req = httptest.NewRequest(tc.method, tc.path, nil)
				}

				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)

				// 确保请求成功
				if w.Code >= 400 {
					b.Errorf("Request failed with status %d", w.Code)
				}
			}
		})
	}
}

// BenchmarkCacheMiddleware 缓存中间件性能测试
func BenchmarkCacheMiddleware(b *testing.B) {
	app, _, rdb := setupTestRouter()

	// 设置缓存测试数据
	ctx := context.Background()
	cacheKey := "lawoa:/api/test"
	testResponse := map[string]interface{}{
		"message": "Hello, World!",
		"timestamp": time.Now().Unix(),
	}
	responseData, _ := json.Marshal(testResponse)
	rdb.Set(ctx, cacheKey, responseData, 5*time.Minute)

	// 测试缓存命中
	b.Run("CacheHit", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/test", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// 测试缓存未命中
	b.Run("CacheMiss", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/test-miss", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})
}

// BenchmarkDatabaseQueries 数据库查询性能测试
func BenchmarkDatabaseQueries(b *testing.B) {
	_, db, _ := setupTestRouter()

	// 创建测试数据
	testUsers := make([]models.User, 1000)
	for i := 0; i < 1000; i++ {
		testUsers[i] = models.User{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Role:  "lawyer",
		}
	}
	db.CreateInBatches(testUsers, 100)

	// 测试简单查询
	b.Run("SimpleQuery", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var user models.User
			err := db.First(&user, i%1000+1).Error
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 测试分页查询
	b.Run("PaginatedQuery", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var users []models.User
			offset := (i % 10) * 10
			err := db.Offset(offset).Limit(10).Find(&users).Error
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 测试复杂查询
	b.Run("ComplexQuery", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var count int64
			err := db.Model(&models.User{}).Where("role = ? AND name LIKE ?", "lawyer", "User%").Count(&count).Error
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkRedisOperations Redis操作性能测试
func BenchmarkRedisOperations(b *testing.B) {
	_, _, rdb := setupTestRouter()
	ctx := context.Background()

	// 测试SET操作
	b.Run("SET", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("test:key:%d", i)
			value := fmt.Sprintf("value:%d", i)
			err := rdb.Set(ctx, key, value, time.Minute).Err()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// 测试GET操作
	b.Run("GET", func(b *testing.B) {
		// 预设数据
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("test:key:%d", i)
			value := fmt.Sprintf("value:%d", i)
			rdb.Set(ctx, key, value, time.Minute)
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("test:key:%d", i%1000)
			_, err := rdb.Get(ctx, key).Result()
			if err != nil && err != redis.Nil {
				b.Fatal(err)
			}
		}
	})

	// 测试Pipeline操作
	b.Run("Pipeline", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			pipe := rdb.Pipeline()
			for j := 0; j < 10; j++ {
				key := fmt.Sprintf("pipe:key:%d:%d", i, j)
				value := fmt.Sprintf("value:%d:%d", i, j)
				pipe.Set(ctx, key, value, time.Minute)
			}
			_, err := pipe.Exec(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMiddlewarePerformance 中间件性能测试
func BenchmarkMiddlewarePerformance(b *testing.B) {
	app, _, _ := setupTestRouter()

	// 测试性能监控中间件
	b.Run("PerformanceMiddleware", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/test", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// 测试限流中间件
	b.Run("RateLimitMiddleware", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("GET", "/api/test", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// 测试CORS中间件
	b.Run("CORSMiddleware", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req := httptest.NewRequest("OPTIONS", "/api/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})
}

// BenchmarkConcurrentRequests 并发请求性能测试
func BenchmarkConcurrentRequests(b *testing.B) {
	app, _, _ := setupTestRouter()

	concurrencyLevels := []int{1, 10, 50, 100}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.ResetTimer()
			b.SetParallelism(concurrency)

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := httptest.NewRequest("GET", "/api/dashboard/statistics", nil)
					w := httptest.NewRecorder()
					app.ServeHTTP(w, req)
				}
			})
		})
	}
}

// TestPerformanceMonitoring 性能监控功能测试
func TestPerformanceMonitoring(t *testing.T) {
	// 初始化性能监控
	monitoring.InitPerformanceMetrics()
	metrics := monitoring.GetPerformanceMetrics()
	require.NotNil(t, metrics)

	// 测试指标记录
	metrics.RecordHTTPRequest("GET", "/test", "200", 100*time.Millisecond, 1024, 2048)
	metrics.RecordDBQuery("SELECT", "users", "success", 50*time.Millisecond)
	metrics.RecordRedisOperation("GET", "success", 5*time.Millisecond)

	// 测试业务指标
	metrics.IncrementCasesCreated()
	metrics.IncrementClientsRegistered()
	metrics.IncrementUsersLoggedIn()

	// 获取指标摘要
	summary := metrics.GetMetricsSummary()
	assert.NotNil(t, summary)
	assert.Contains(t, summary, "system")
	assert.Contains(t, summary, "runtime")
}

// TestCacheEfficiency 缓存效率测试
func TestCacheEfficiency(t *testing.T) {
	_, _, rdb := setupTestRouter()
	ctx := context.Background()

	// 创建缓存配置
	config := middleware.CacheConfig{
		TTL:          5 * time.Minute,
		RedisClient:  rdb,
		KeyPrefix:    "test",
		MaxBodySize:  1024 * 1024,
	}

	// 测试缓存失效
	middleware.InvalidateCache(rdb, "test", "user:*", "dashboard:*")

	// 验证缓存已清空
	keys, _ := rdb.Keys(ctx, "test:*").Result()
	assert.Equal(t, 0, len(keys))
}

// TestMemoryUsage 内存使用测试
func TestMemoryUsage(t *testing.T) {
	monitoring.InitPerformanceMetrics()
	metrics := monitoring.GetPerformanceMetrics()

	// 记录初始内存使用
	summary1 := metrics.GetMetricsSummary()
	initialGoroutines := summary1["system"].(map[string]interface{})["goroutines"].(int)

	// 创建大量goroutines
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(100 * time.Millisecond)
			done <- true
		}()
	}

	// 等待所有goroutines完成
	for i := 0; i < 100; i++ {
		<-done
	}

	// 等待一段时间让goroutines清理
	time.Sleep(200 * time.Millisecond)

	// 检查goroutines数量是否恢复正常
	summary2 := metrics.GetMetricsSummary()
	finalGoroutines := summary2["system"].(map[string]interface{})["goroutines"].(int)

	// 允许一些goroutines的创建开销
	assert.True(t, finalGoroutines <= initialGoroutines+10,
		"Goroutine count should return to normal level")
}