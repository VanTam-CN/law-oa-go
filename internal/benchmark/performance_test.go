package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/database"
)

// 性能基准测试相关指标
var (
	benchmarkDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "benchmark_duration_seconds",
		Help:    "Duration of benchmark operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})

	benchmarkOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "benchmark_operations_total",
		Help: "Total number of benchmark operations",
	}, []string{"operation"})

	benchmarkErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "benchmark_errors_total",
		Help: "Total number of benchmark errors",
	}, []string{"operation"})
)

// CacheBenchmark 缓存基准测试
type CacheBenchmark struct {
	cacheService *cache.CacheService
}

func NewCacheBenchmark(cacheService *cache.CacheService) *CacheBenchmark {
	return &CacheBenchmark{cacheService: cacheService}
}

// BenchmarkSet 缓存设置性能测试
func (cb *CacheBenchmark) BenchmarkSet(b *testing.B) {
	ctx := context.Background()
	testData := map[string]interface{}{
		"id":    1,
		"name":  "测试数据",
		"value": "这是一个性能测试的缓存数据",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		key := fmt.Sprintf("benchmark:set:%d", i)
		err := cb.cacheService.Set(ctx, key, testData, time.Minute)
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("cache_set").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("cache_set").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("cache_set").Inc()
			b.Errorf("Failed to set cache: %v", err)
		}
	}
}

// BenchmarkGet 缓存获取性能测试
func (cb *CacheBenchmark) BenchmarkGet(b *testing.B) {
	ctx := context.Background()
	
	// 预先设置一些数据
	testData := map[string]interface{}{
		"id":    1,
		"name":  "测试数据",
		"value": "这是一个性能测试的缓存数据",
	}
	
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("benchmark:get:%d", i)
		cb.cacheService.Set(ctx, key, testData, time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		key := fmt.Sprintf("benchmark:get:%d", i%1000)
		var result map[string]interface{}
		err := cb.cacheService.Get(ctx, key, &result)
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("cache_get").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("cache_get").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("cache_get").Inc()
			b.Errorf("Failed to get cache: %v", err)
		}
	}
}

// BenchmarkGetOrSet 缓存获取或设置性能测试
func (cb *CacheBenchmark) BenchmarkGetOrSet(b *testing.B) {
	ctx := context.Background()
	
	callCount := 0
	fetchFunc := func() (interface{}, error) {
		callCount++
		return map[string]interface{}{
			"id":    callCount,
			"name":  "动态数据",
			"value": "这是动态获取的数据",
		}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		key := fmt.Sprintf("benchmark:get_or_set:%d", i%100)
		var result map[string]interface{}
		err := cb.cacheService.GetOrSet(ctx, key, &result, fetchFunc, time.Minute)
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("cache_get_or_set").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("cache_get_or_set").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("cache_get_or_set").Inc()
			b.Errorf("Failed to get or set cache: %v", err)
		}
	}
}

// DatabaseBenchmark 数据库基准测试
type DatabaseBenchmark struct {
	db *gorm.DB
}

func NewDatabaseBenchmark(db *gorm.DB) *DatabaseBenchmark {
	return &DatabaseBenchmark{db: db}
}

// BenchmarkQuery 数据库查询性能测试
func (dbb *DatabaseBenchmark) BenchmarkQuery(b *testing.B) {
	// 创建测试表
	type TestUser struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100"`
		Email string `gorm:"size:255"`
		Age   int    `gorm:"size:3"`
	}

	// 自动迁移
	if err := dbb.db.AutoMigrate(&TestUser{}); err != nil {
		b.Fatalf("Failed to migrate: %v", err)
	}

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		user := TestUser{
			Name:  fmt.Sprintf("用户%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   i % 100,
		}
		dbb.db.Create(&user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		var users []TestUser
		err := dbb.db.Where("age > ?", 50).Limit(100).Find(&users).Error
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("db_query").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("db_query").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("db_query").Inc()
			b.Errorf("Failed to query database: %v", err)
		}
	}
}

// BenchmarkInsert 数据库插入性能测试
func (dbb *DatabaseBenchmark) BenchmarkInsert(b *testing.B) {
	type TestUser struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100"`
		Email string `gorm:"size:255"`
		Age   int    `gorm:"size:3"`
	}

	// 自动迁移
	if err := dbb.db.AutoMigrate(&TestUser{}); err != nil {
		b.Fatalf("Failed to migrate: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		user := TestUser{
			Name:  fmt.Sprintf("用户%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   i % 100,
		}
		err := dbb.db.Create(&user).Error
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("db_insert").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("db_insert").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("db_insert").Inc()
			b.Errorf("Failed to insert: %v", err)
		}
	}
}

// BenchmarkUpdate 数据库更新性能测试
func (dbb *DatabaseBenchmark) BenchmarkUpdate(b *testing.B) {
	type TestUser struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100"`
		Email string `gorm:"size:255"`
		Age   int    `gorm:"size:3"`
	}

	// 自动迁移
	if err := dbb.db.AutoMigrate(&TestUser{}); err != nil {
		b.Fatalf("Failed to migrate: %v", err)
	}

	// 插入测试数据
	for i := 0; i < 1000; i++ {
		user := TestUser{
			Name:  fmt.Sprintf("用户%d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   i % 100,
		}
		dbb.db.Create(&user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		err := dbb.db.Model(&TestUser{}).Where("id = ?", i%1000+1).Update("name", fmt.Sprintf("更新用户%d", i)).Error
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("db_update").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("db_update").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("db_update").Inc()
			b.Errorf("Failed to update: %v", err)
		}
	}
}

// JSONBenchmark JSON序列化基准测试
type JSONBenchmark struct{}

func NewJSONBenchmark() *JSONBenchmark {
	return &JSONBenchmark{}
}

// BenchmarkMarshal JSON序列化性能测试
func (jb *JSONBenchmark) BenchmarkMarshal(b *testing.B) {
	testData := map[string]interface{}{
		"id":          1,
		"name":        "测试数据",
		"description": "这是一个用于JSON序列化性能测试的复杂数据结构",
		"metadata": map[string]interface{}{
			"created_at": time.Now(),
			"updated_at": time.Now(),
			"version":    "1.0.0",
			"tags":       []string{"性能", "测试", "基准"},
		},
		"items": []map[string]interface{}{
			{"id": 1, "name": "项目1", "value": 100},
			{"id": 2, "name": "项目2", "value": 200},
			{"id": 3, "name": "项目3", "value": 300},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		_, err := json.Marshal(testData)
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("json_marshal").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("json_marshal").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("json_marshal").Inc()
			b.Errorf("Failed to marshal JSON: %v", err)
		}
	}
}

// BenchmarkUnmarshal JSON反序列化性能测试
func (jb *JSONBenchmark) BenchmarkUnmarshal(b *testing.B) {
	testData := map[string]interface{}{
		"id":          1,
		"name":        "测试数据",
		"description": "这是一个用于JSON序列化性能测试的复杂数据结构",
		"metadata": map[string]interface{}{
			"created_at": time.Now(),
			"updated_at": time.Now(),
			"version":    "1.0.0",
			"tags":       []string{"性能", "测试", "基准"},
		},
		"items": []map[string]interface{}{
			{"id": 1, "name": "项目1", "value": 100},
			{"id": 2, "name": "项目2", "value": 200},
			{"id": 3, "name": "项目3", "value": 300},
		},
	}

	jsonData, _ := json.Marshal(testData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		
		var result map[string]interface{}
		err := json.Unmarshal(jsonData, &result)
		
		duration := time.Since(start)
		benchmarkDuration.WithLabelValues("json_unmarshal").Observe(duration.Seconds())
		benchmarkOperations.WithLabelValues("json_unmarshal").Inc()
		
		if err != nil {
			benchmarkErrors.WithLabelValues("json_unmarshal").Inc()
			b.Errorf("Failed to unmarshal JSON: %v", err)
		}
	}
}

// RunAllBenchmarks 运行所有基准测试
func RunAllBenchmarks(cacheService *cache.CacheService, db *gorm.DB) {
	fmt.Println("开始运行性能基准测试...")
	
	// 缓存基准测试
	cacheBenchmark := NewCacheBenchmark(cacheService)
	fmt.Println("1. 缓存设置性能测试")
	testing.Benchmark{}.Run("CacheSet", cacheBenchmark.BenchmarkSet)
	
	fmt.Println("2. 缓存获取性能测试")
	testing.Benchmark{}.Run("CacheGet", cacheBenchmark.BenchmarkGet)
	
	fmt.Println("3. 缓存获取或设置性能测试")
	testing.Benchmark{}.Run("CacheGetOrSet", cacheBenchmark.BenchmarkGetOrSet)
	
	// 数据库基准测试
	dbBenchmark := NewDatabaseBenchmark(db)
	fmt.Println("4. 数据库查询性能测试")
	testing.Benchmark{}.Run("DatabaseQuery", dbBenchmark.BenchmarkQuery)
	
	fmt.Println("5. 数据库插入性能测试")
	testing.Benchmark{}.Run("DatabaseInsert", dbBenchmark.BenchmarkInsert)
	
	fmt.Println("6. 数据库更新性能测试")
	testing.Benchmark{}.Run("DatabaseUpdate", dbBenchmark.BenchmarkUpdate)
	
	// JSON基准测试
	jsonBenchmark := NewJSONBenchmark()
	fmt.Println("7. JSON序列化性能测试")
	testing.Benchmark{}.Run("JSONMarshal", jsonBenchmark.BenchmarkMarshal)
	
	fmt.Println("8. JSON反序列化性能测试")
	testing.Benchmark{}.Run("JSONUnmarshal", jsonBenchmark.BenchmarkUnmarshal)
	
	fmt.Println("性能基准测试完成!")
}

// BenchmarkReport 基准测试报告
type BenchmarkReport struct {
	Operation   string        `json:"operation"`
	Duration    time.Duration `json:"duration"`
	Operations  int           `json:"operations"`
	OpsPerSec   float64       `json:"ops_per_sec"`
	AvgDuration time.Duration `json:"avg_duration"`
	Errors      int           `json:"errors"`
}

// GenerateReport 生成基准测试报告
func GenerateReport() []BenchmarkReport {
	// 这里可以从Prometheus指标中提取数据生成报告
	// 简化版本，实际项目中需要更复杂的实现
	return []BenchmarkReport{
		{
			Operation:   "cache_set",
			Duration:    10 * time.Second,
			Operations:  10000,
			OpsPerSec:   1000,
			AvgDuration: time.Millisecond,
			Errors:      0,
		},
		{
			Operation:   "cache_get",
			Duration:    10 * time.Second,
			Operations:  15000,
			OpsPerSec:   1500,
			AvgDuration: time.Microsecond * 667,
			Errors:      0,
		},
	}
}

// SaveReport 保存基准测试报告
func SaveReport(report []BenchmarkReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	// 这里应该写入文件，简化版本
	fmt.Printf("基准测试报告:\n%s\n", string(data))
	return nil
}