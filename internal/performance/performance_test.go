package performance

import (
	"testing"
	"time"

	"law-oa-go/internal/config"
)

// TestDatabasePerformanceConfig 测试数据库性能配置
func TestDatabasePerformanceConfig(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected config.DatabasePerformanceConfig
	}{
		{
			name: "开发环境默认配置",
			env:  "development",
			expected: config.DatabasePerformanceConfig{
				MaxOpenConns:       25,
				MaxIdleConns:       5,
				ConnMaxLifetime:    30 * time.Minute,
				ConnMaxIdleTime:    5 * time.Minute,
				EnablePerformance:  true,
			},
		},
		{
			name: "生产环境默认配置",
			env:  "production",
			expected: config.DatabasePerformanceConfig{
				MaxOpenConns:       100,
				MaxIdleConns:       10,
				ConnMaxLifetime:    time.Hour,
				ConnMaxIdleTime:    10 * time.Minute,
				EnablePerformance:  true,
			},
		},
		{
			name: "测试环境默认配置",
			env:  "testing",
			expected: config.DatabasePerformanceConfig{
				MaxOpenConns:       5,
				MaxIdleConns:       2,
				ConnMaxLifetime:    5 * time.Minute,
				ConnMaxIdleTime:    time.Minute,
				EnablePerformance:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Environment: tt.env,
				Database: config.DatabaseConfig{
					EnablePerformance: true,
				},
			}

			result := cfg.GetDatabasePerformanceConfig()

			if result.MaxOpenConns != tt.expected.MaxOpenConns {
				t.Errorf("MaxOpenConns = %v, want %v", result.MaxOpenConns, tt.expected.MaxOpenConns)
			}

			if result.MaxIdleConns != tt.expected.MaxIdleConns {
				t.Errorf("MaxIdleConns = %v, want %v", result.MaxIdleConns, tt.expected.MaxIdleConns)
			}

			if result.ConnMaxLifetime != tt.expected.ConnMaxLifetime {
				t.Errorf("ConnMaxLifetime = %v, want %v", result.ConnMaxLifetime, tt.expected.ConnMaxLifetime)
			}

			if result.ConnMaxIdleTime != tt.expected.ConnMaxIdleTime {
				t.Errorf("ConnMaxIdleTime = %v, want %v", result.ConnMaxIdleTime, tt.expected.ConnMaxIdleTime)
			}

			if result.EnablePerformance != tt.expected.EnablePerformance {
				t.Errorf("EnablePerformance = %v, want %v", result.EnablePerformance, tt.expected.EnablePerformance)
			}
		})
	}
}

// TestCustomDatabasePerformanceConfig 测试自定义数据库性能配置
func TestCustomDatabasePerformanceConfig(t *testing.T) {
	cfg := &config.Config{
		Environment: "production",
		Database: config.DatabaseConfig{
			MaxOpenConns:      200,
			MaxIdleConns:      20,
			ConnMaxLifetime:   2 * time.Hour,
			ConnMaxIdleTime:   15 * time.Minute,
			EnablePerformance: true,
		},
	}

	result := cfg.GetDatabasePerformanceConfig()

	// 应该使用自定义配置而不是默认配置
	if result.MaxOpenConns != 200 {
		t.Errorf("MaxOpenConns = %v, want %v", result.MaxOpenConns, 200)
	}

	if result.MaxIdleConns != 20 {
		t.Errorf("MaxIdleConns = %v, want %v", result.MaxIdleConns, 20)
	}

	if result.ConnMaxLifetime != 2*time.Hour {
		t.Errorf("ConnMaxLifetime = %v, want %v", result.ConnMaxLifetime, 2*time.Hour)
	}

	if result.ConnMaxIdleTime != 15*time.Minute {
		t.Errorf("ConnMaxIdleTime = %v, want %v", result.ConnMaxIdleTime, 15*time.Minute)
	}
}

// TestDatabaseConfigCompatibility 测试数据库配置向后兼容性
func TestDatabaseConfigCompatibility(t *testing.T) {
	cfg := config.DatabaseConfig{
		Driver:    "postgres",
		Host:      "localhost",
		Port:      "5432",
		Username:  "test",
		Password:  "test",
		Database:  "test_db",
		Charset:   "utf8",
		ParseTime: true,
		Loc:       "UTC",
		SSLMode:   "disable",
	}

	// 测试兼容性方法
	if cfg.GetDriver() != "postgres" {
		t.Errorf("GetDriver() = %v, want %v", cfg.GetDriver(), "postgres")
	}

	if cfg.GetCharset() != "utf8" {
		t.Errorf("GetCharset() = %v, want %v", cfg.GetCharset(), "utf8")
	}

	if !cfg.GetParseTime() {
		t.Errorf("GetParseTime() = %v, want %v", cfg.GetParseTime(), true)
	}

	if cfg.GetLoc() != "UTC" {
		t.Errorf("GetLoc() = %v, want %v", cfg.GetLoc(), "UTC")
	}

	if cfg.GetHost() != "localhost" {
		t.Errorf("GetHost() = %v, want %v", cfg.GetHost(), "localhost")
	}

	if cfg.GetPort() != "5432" {
		t.Errorf("GetPort() = %v, want %v", cfg.GetPort(), "5432")
	}
}

// BenchmarkDatabasePerformanceConfig 性能基准测试
func BenchmarkDatabasePerformanceConfig(b *testing.B) {
	cfg := &config.Config{
		Environment: "production",
		Database: config.DatabaseConfig{
			MaxOpenConns:      100,
			MaxIdleConns:      10,
			ConnMaxLifetime:   time.Hour,
			ConnMaxIdleTime:   10 * time.Minute,
			EnablePerformance: true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.GetDatabasePerformanceConfig()
	}
}