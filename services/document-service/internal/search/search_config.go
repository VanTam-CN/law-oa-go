package search

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// LoadSearchConfigFromEnv 从环境变量加载搜索配置
func LoadSearchConfigFromEnv() (*SearchConfig, error) {
	config := &SearchConfig{
		Elasticsearch: ElasticsearchConfig{
			Addresses:     getEnvSlice("ELASTICSEARCH_ADDRESSES", []string{"http://localhost:9200"}),
			Username:      getEnv("ELASTICSEARCH_USERNAME", ""),
			Password:      getEnv("ELASTICSEARCH_PASSWORD", ""),
			APIKey:        getEnv("ELASTICSEARCH_API_KEY", ""),
			CloudID:       getEnv("ELASTICSEARCH_CLOUD_ID", ""),
			IndexName:     getEnv("ELASTICSEARCH_INDEX_NAME", "documents"),
			IndexPattern:  getEnv("ELASTICSEARCH_INDEX_PATTERN", "documents-*"),
			SkipTLSVerify: getEnvBool("ELASTICSEARCH_SKIP_TLS_VERIFY", false),
			RetryCount:    getEnvInt("ELASTICSEARCH_RETRY_COUNT", 3),
			RetryBackoff:  getEnvInt("ELASTICSEARCH_RETRY_BACKOFF", 100),
			Timeout:       getEnvInt("ELASTICSEARCH_TIMEOUT", 30),
		},
		Redis: RedisCacheConfig{
			Host:       getEnv("REDIS_HOST", "localhost"),
			Port:       getEnvInt("REDIS_PORT", 6379),
			Password:   getEnv("REDIS_PASSWORD", ""),
			Database:   getEnvInt("REDIS_DB", 0),
			PoolSize:   getEnvInt("REDIS_POOL_SIZE", 10),
			KeyPrefix:  getEnv("REDIS_KEY_PREFIX", "doc_search:"),
			DefaultTTL: getEnvInt("REDIS_DEFAULT_TTL", 300),
		},
		Indexing: IndexingConfig{
			BatchSize:     getEnvInt("SEARCH_BATCH_SIZE", 100),
			Workers:       getEnvInt("SEARCH_WORKERS", 4),
			SyncInterval:  getEnvDuration("SEARCH_SYNC_INTERVAL", time.Hour),
			RetryAttempts: getEnvInt("SEARCH_RETRY_ATTEMPTS", 3),
			RetryDelay:    getEnvDuration("SEARCH_RETRY_DELAY", time.Second),
		},
		Caching: CachingConfig{
			Enabled:    getEnvBool("SEARCH_CACHE_ENABLED", true),
			DefaultTTL: getEnvDuration("SEARCH_CACHE_TTL", 5*time.Minute),
			MaxTTL:     getEnvDuration("SEARCH_CACHE_MAX_TTL", 30*time.Minute),
			KeyPrefix:  getEnv("SEARCH_CACHE_KEY_PREFIX", "search:"),
		},
	}

	return config, nil
}

// ValidateSearchConfig 验证搜索配置
func ValidateSearchConfig(config *SearchConfig) error {
	if config == nil {
		return fmt.Errorf("search config is nil")
	}

	// 验证Elasticsearch配置
	if len(config.Elasticsearch.Addresses) == 0 {
		return fmt.Errorf("elasticsearch addresses are required")
	}

	// 验证Redis配置（如果缓存启用）
	if config.Caching.Enabled {
		if config.Redis.Host == "" {
			return fmt.Errorf("redis host is required when caching is enabled")
		}
		if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
			return fmt.Errorf("redis port must be between 1 and 65535")
		}
	}

	// 验证索引配置
	if config.Indexing.BatchSize <= 0 {
		return fmt.Errorf("indexing batch size must be positive")
	}
	if config.Indexing.Workers <= 0 {
		return fmt.Errorf("indexing workers must be positive")
	}
	if config.Indexing.Workers > 100 {
		return fmt.Errorf("indexing workers should not exceed 100")
	}

	// 验证缓存配置
	if config.Caching.Enabled {
		if config.Caching.DefaultTTL <= 0 {
			return fmt.Errorf("cache default TTL must be positive")
		}
		if config.Caching.MaxTTL <= 0 {
			return fmt.Errorf("cache max TTL must be positive")
		}
		if config.Caching.MaxTTL < config.Caching.DefaultTTL {
			return fmt.Errorf("cache max TTL must be greater than or equal to default TTL")
		}
	}

	return nil
}

// GetProductionConfig 获取生产环境配置
func GetProductionConfig() *SearchConfig {
	return &SearchConfig{
		Elasticsearch: ElasticsearchConfig{
			Addresses:     []string{"http://elasticsearch:9200"},
			Username:      "elastic",
			Password:      os.Getenv("ELASTICSEARCH_PASSWORD"),
			IndexName:     "documents_prod",
			IndexPattern:  "documents_prod-*",
			SkipTLSVerify: false,
			RetryCount:    5,
			RetryBackoff:  200,
			Timeout:       30,
		},
		Redis: RedisCacheConfig{
			Host:       "redis",
			Port:       6379,
			Password:   os.Getenv("REDIS_PASSWORD"),
			Database:   1,
			PoolSize:   20,
			KeyPrefix:  "doc_search_prod:",
			DefaultTTL: 600, // 10分钟
		},
		Indexing: IndexingConfig{
			BatchSize:     200,
			Workers:       8,
			SyncInterval:  30 * time.Minute,
			RetryAttempts: 5,
			RetryDelay:    2 * time.Second,
		},
		Caching: CachingConfig{
			Enabled:    true,
			DefaultTTL: 10 * time.Minute,
			MaxTTL:     1 * time.Hour,
			KeyPrefix:  "search_prod:",
		},
	}
}

// GetDevelopmentConfig 获取开发环境配置
func GetDevelopmentConfig() *SearchConfig {
	return &SearchConfig{
		Elasticsearch: ElasticsearchConfig{
			Addresses:     []string{"http://localhost:9200"},
			IndexName:     "documents_dev",
			IndexPattern:  "documents_dev-*",
			SkipTLSVerify: true,
			RetryCount:    3,
			RetryBackoff:  100,
			Timeout:       10,
		},
		Redis: RedisCacheConfig{
			Host:       "localhost",
			Port:       6379,
			Database:   0,
			PoolSize:   5,
			KeyPrefix:  "doc_search_dev:",
			DefaultTTL: 300, // 5分钟
		},
		Indexing: IndexingConfig{
			BatchSize:     50,
			Workers:       2,
			SyncInterval:  10 * time.Minute,
			RetryAttempts: 2,
			RetryDelay:    time.Second,
		},
		Caching: CachingConfig{
			Enabled:    true,
			DefaultTTL: 2 * time.Minute,
			MaxTTL:     10 * time.Minute,
			KeyPrefix:  "search_dev:",
		},
	}
}

// GetTestConfig 获取测试环境配置
func GetTestConfig() *SearchConfig {
	return &SearchConfig{
		Elasticsearch: ElasticsearchConfig{
			Addresses:     []string{"http://localhost:9200"},
			IndexName:     "documents_test",
			IndexPattern:  "documents_test-*",
			SkipTLSVerify: true,
			RetryCount:    1,
			RetryBackoff:  50,
			Timeout:       5,
		},
		Redis: RedisCacheConfig{
			Host:       "localhost",
			Port:       6379,
			Database:   15, // 测试数据库
			PoolSize:   2,
			KeyPrefix:  "doc_search_test:",
			DefaultTTL: 30, // 30秒
		},
		Indexing: IndexingConfig{
			BatchSize:     10,
			Workers:       1,
			SyncInterval:  time.Minute,
			RetryAttempts: 1,
			RetryDelay:    500 * time.Millisecond,
		},
		Caching: CachingConfig{
			Enabled:    false, // 测试环境禁用缓存
			DefaultTTL: 30 * time.Second,
			MaxTTL:     5 * time.Minute,
			KeyPrefix:  "search_test:",
		},
	}
}

// LogConfig 记录配置信息
func LogConfig(config *SearchConfig, logger *logrus.Logger) {
	logger.WithFields(logrus.Fields{
		"elasticsearch_addresses": config.Elasticsearch.Addresses,
		"elasticsearch_index":   config.Elasticsearch.IndexName,
		"elasticsearch_timeout": config.Elasticsearch.Timeout,
		"redis_enabled":         config.Caching.Enabled,
		"redis_host":            config.Redis.Host,
		"redis_port":            config.Redis.Port,
		"indexing_batch_size":   config.Indexing.BatchSize,
		"indexing_workers":      config.Indexing.Workers,
		"cache_enabled":         config.Caching.Enabled,
		"cache_default_ttl":      config.Caching.DefaultTTL,
	}).Info("Search configuration loaded")
}

// 辅助函数

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ConfigHealth 配置健康检查
type ConfigHealth struct {
	Elasticsearch ElasticsearchHealth `json:"elasticsearch"`
	Redis         RedisHealth         `json:"redis"`
	Overall       string               `json:"overall"`
	Errors        []string             `json:"errors"`
}

// ElasticsearchHealth ES健康状态
type ElasticsearchHealth struct {
	Connected bool   `json:"connected"`
	Cluster   string `json:"cluster"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// RedisHealth Redis健康状态
type RedisHealth struct {
	Connected bool   `json:"connected"`
	Database  int    `json:"database"`
	Error     string `json:"error,omitempty"`
}

// CheckConfigHealth 检查配置健康状态
func CheckConfigHealth(config *SearchConfig) *ConfigHealth {
	health := &ConfigHealth{
		Errors: make([]string, 0),
	}

	// 检查Elasticsearch配置
	if len(config.Elasticsearch.Addresses) == 0 {
		health.Errors = append(health.Errors, "elasticsearch addresses not configured")
		health.Overall = "unhealthy"
	} else {
		health.Elasticsearch.Connected = true
		health.Elasticsearch.Cluster = "unknown"
		health.Elasticsearch.Status = "unknown"
	}

	// 检查Redis配置
	if config.Caching.Enabled {
		if config.Redis.Host == "" {
			health.Errors = append(health.Errors, "redis host not configured")
			health.Overall = "unhealthy"
		} else {
			health.Redis.Connected = true
			health.Redis.Database = config.Redis.Database
		}
	}

	if len(health.Errors) == 0 {
		health.Overall = "healthy"
	}

	return health
}