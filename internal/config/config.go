package config

import (
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Environment   string              `mapstructure:"environment"`
	Port          string              `mapstructure:"port"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Log           LogConfig           `mapstructure:"log"`
	CORS          CORSConfig          `mapstructure:"cors"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host      string `mapstructure:"host"`
	Port      string `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Database  string `mapstructure:"database"`
	Charset   string `mapstructure:"charset"`
	ParseTime bool   `mapstructure:"parseTime"`
	Loc       string `mapstructure:"loc"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expiresIn"`
	RefreshIn int    `mapstructure:"refreshIn"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowedOrigins"`
	AllowedMethods []string `mapstructure:"allowedMethods"`
	AllowedHeaders []string `mapstructure:"allowedHeaders"`
	MaxAge         string   `mapstructure:"maxAge"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		// 如果.env文件不存在，使用环境变量
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	// 设置默认值
	viper.SetDefault("environment", "development")
	viper.SetDefault("port", "8080")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "3306")
	viper.SetDefault("database.username", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.database", "law_oa")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.parseTime", true)
	viper.SetDefault("database.loc", "Local")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("elasticsearch.host", "localhost")
	viper.SetDefault("elasticsearch.port", "9200")
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("jwt.secret", "") // 强制要求配置JWT密钥
	viper.SetDefault("jwt.expiresIn", 3600)
	viper.SetDefault("jwt.refreshIn", 7200)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("cors.allowedOrigins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowedHeaders", []string{"Content-Type", "Authorization", "X-Request-ID"})
	viper.SetDefault("cors.maxAge", "86400")

	// 从环境变量读取配置
	viper.AutomaticEnv()

	// 绑定环境变量到配置结构
	bindings := map[string]string{
		"environment":            "ENVIRONMENT",
		"port":                   "PORT",
		"database.host":          "DB_HOST",
		"database.port":          "DB_PORT",
		"database.username":      "DB_USERNAME",
		"database.password":      "DB_PASSWORD",
		"database.database":      "DB_DATABASE",
		"redis.host":             "REDIS_HOST",
		"redis.port":             "REDIS_PORT",
		"redis.password":         "REDIS_PASSWORD",
		"redis.db":               "REDIS_DB",
		"elasticsearch.host":     "ES_HOST",
		"elasticsearch.port":     "ES_PORT",
		"elasticsearch.username": "ES_USERNAME",
		"elasticsearch.password": "ES_PASSWORD",
		"jwt.secret":             "JWT_SECRET",
		"jwt.expiresIn":          "JWT_EXPIRES_IN",
		"jwt.refreshIn":          "JWT_REFRESH_IN",
	}

	for key, env := range bindings {
		if err := viper.BindEnv(key, env); err != nil {
			return nil, fmt.Errorf("failed to bind env var %s to key %s: %w", env, key, err)
		}
	}

	// 读取配置文件
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/law-oa")

	// 如果配置文件存在，则读取
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 转换数据库端口为字符串
	if config.Database.Port == "" {
		config.Database.Port = strconv.Itoa(viper.GetInt("database.port"))
	}

	// 转换Redis端口为字符串
	if config.Redis.Port == "" {
		config.Redis.Port = strconv.Itoa(viper.GetInt("redis.port"))
	}

	// 转换Elasticsearch端口为字符串
	if config.Elasticsearch.Port == "" {
		config.Elasticsearch.Port = strconv.Itoa(viper.GetInt("elasticsearch.port"))
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// GetDatabaseDSN 获取数据库DSN
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%v&loc=%s",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
		c.Database.Charset,
		c.Database.ParseTime,
		c.Database.Loc,
	)
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// GetElasticsearchURL 获取Elasticsearch URL
func (c *Config) GetElasticsearchURL() string {
	return fmt.Sprintf("http://%s:%s", c.Elasticsearch.Host, c.Elasticsearch.Port)
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetPort 获取端口
func (c *Config) GetPort() string {
	if c.Port == "" {
		return "8080"
	}
	return c.Port
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 验证JWT密钥
	if c.JWT.Secret == "" || c.JWT.Secret == "your-secret-key" {
		return fmt.Errorf("JWT secret must be configured and cannot be the default value")
	}

	// 验证JWT密钥长度
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	// 验证数据库配置
	if c.Database.Host == "" || c.Database.Username == "" || c.Database.Database == "" {
		return fmt.Errorf("database configuration is incomplete")
	}

	return nil
}
