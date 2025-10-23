package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Version     string         `mapstructure:"version"`
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Redis       RedisConfig    `mapstructure:"redis"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	MinIO       MinIOConfig    `mapstructure:"minio"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Casbin      CasbinConfig   `mapstructure:"casbin"`
	Logging     LoggingConfig  `mapstructure:"logging"`
	OCR         OCRConfig      `mapstructure:"ocr"`
	Search      SearchConfig   `mapstructure:"search"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Mode         string `mapstructure:"mode"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
	RateLimit    int    `mapstructure:"rate_limit"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Hosts     []string `mapstructure:"hosts"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	IndexName string   `mapstructure:"index_name"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Bucket    string `mapstructure:"bucket"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string `mapstructure:"secret"`
	ExpirationHours int    `mapstructure:"expiration_hours"`
	Issuer          string `mapstructure:"issuer"`
}

// CasbinConfig Casbin配置
type CasbinConfig struct {
	ModelPath  string `mapstructure:"model_path"`
	PolicyPath string `mapstructure:"policy_path"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// OCRConfig OCR配置
type OCRConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Tesseract  string `mapstructure:"tesseract_path"`
	Languages  []string `mapstructure:"languages"`
	Confidence float64 `mapstructure:"min_confidence"`
}

// SearchConfig 搜索配置
type SearchConfig struct {
	IndexBatchSize int `mapstructure:"index_batch_size"`
	SearchTimeout  int `mapstructure:"search_timeout"`
	HighlightSize  int `mapstructure:"highlight_size"`
}

// Load 加载配置
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置环境变量前缀
	viper.SetEnvPrefix("DOC_SERVICE")
	viper.AutomaticEnv()

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Config file not found, using defaults")
		} else {
			log.Fatalf("Error reading config file: %v", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	return &config
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("environment", "development")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 60)
	viper.SetDefault("server.rate_limit", 1000)

	// 数据库默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.database", "document_service")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", 3600)

	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// Elasticsearch默认配置
	viper.SetDefault("elasticsearch.hosts", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("elasticsearch.index_name", "documents")

	// MinIO默认配置
	viper.SetDefault("minio.endpoint", "localhost:9000")
	viper.SetDefault("minio.access_key", "minioadmin")
	viper.SetDefault("minio.secret_key", "minioadmin")
	viper.SetDefault("minio.use_ssl", false)
	viper.SetDefault("minio.bucket", "documents")

	// JWT默认配置
	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration_hours", 24)
	viper.SetDefault("jwt.issuer", "document-service")

	// Casbin默认配置
	viper.SetDefault("casbin.model_path", "./configs/casbin_model.conf")
	viper.SetDefault("casbin.policy_path", "./configs/casbin_policy.csv")

	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)

	// OCR默认配置
	viper.SetDefault("ocr.enabled", true)
	viper.SetDefault("ocr.tesseract_path", "/usr/bin/tesseract")
	viper.SetDefault("ocr.languages", []string{"eng", "chi_sim"})
	viper.SetDefault("ocr.min_confidence", 0.7)

	// 搜索默认配置
	viper.SetDefault("search.index_batch_size", 100)
	viper.SetDefault("search.search_timeout", 30)
	viper.SetDefault("search.highlight_size", 300)
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// GetRedisAddr 获取Redis连接地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}