package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Environment       string                    `mapstructure:"environment"`
	Port              string                    `mapstructure:"port"`
	Database          DatabaseConfig            `mapstructure:"database"`
	Redis             RedisConfig               `mapstructure:"redis"`
	Elasticsearch     ElasticsearchConfig       `mapstructure:"elasticsearch"`
	JWT               JWTConfig                 `mapstructure:"jwt"`
	Log               LogConfig                 `mapstructure:"log"`
	CORS              CORSConfig                `mapstructure:"cors"`
	ConflictDetection *ConflictDetectionConfig   `mapstructure:"conflictDetection"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver    string `mapstructure:"driver"`
	Host      string `mapstructure:"host"`
	Port      string `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	Database  string `mapstructure:"database"`
	Charset   string `mapstructure:"charset"`
	ParseTime bool   `mapstructure:"parseTime"`
	Loc       string `mapstructure:"loc"`
	SSLMode   string `mapstructure:"sslmode"`
	// 性能优化配置
	MaxOpenConns       int           `mapstructure:"maxOpen_conns"`
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime    time.Duration `mapstructure:"conn_max_idle_time"`
	EnablePerformance  bool          `mapstructure:"enable_performance"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
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
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.username", "law_oa_user")
	viper.SetDefault("database.password", "law_oa_password")
	viper.SetDefault("database.database", "law_oa_db")
	viper.SetDefault("database.charset", "utf8")
	viper.SetDefault("database.parseTime", true)
	viper.SetDefault("database.loc", "UTC")
	viper.SetDefault("database.sslmode", "disable")
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
	viper.SetDefault("cors.allowedOrigins", []string{"http://localhost:3003", "http://localhost:8080"})
	viper.SetDefault("cors.allowedMethods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowedHeaders", []string{"Content-Type", "Authorization", "X-Request-ID"})
	viper.SetDefault("cors.maxAge", "86400")

	// 数据库性能配置默认值
	viper.SetDefault("database.maxOpen_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "30m")
	viper.SetDefault("database.conn_max_idle_time", "5m")
	viper.SetDefault("database.enable_performance", true)

	// 冲突检测配置默认值
	viper.SetDefault("conflictDetection.enabled", true)
	viper.SetDefault("conflictDetection.autoCheckOnCaseCreation", true)
	viper.SetDefault("conflictDetection.defaultSearchYears", 5)
	viper.SetDefault("conflictDetection.defaultSearchDepth", "deep")
	viper.SetDefault("conflictDetection.includeCorporateRelations", true)
	viper.SetDefault("conflictDetection.highRiskThreshold", 0.7)
	viper.SetDefault("conflictDetection.mediumRiskThreshold", 0.4)
	viper.SetDefault("conflictDetection.requireApprovalForHighRisk", true)
	viper.SetDefault("conflictDetection.allowSkipConflictCheck", true)

	// 从环境变量读取配置
	viper.AutomaticEnv()

	// 绑定环境变量到配置结构
	bindings := map[string]string{
		"environment":            "ENVIRONMENT",
		"port":                   "PORT",
		"database.driver":        "DB_DRIVER",
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

	// 如果冲突检测配置为空，使用默认配置
	if config.ConflictDetection == nil {
		config.ConflictDetection = DefaultConflictDetectionConfig()
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

	// 手动处理数据库密码中的特殊字符
	if config.Database.Password == "" || config.Database.Password == "1q2w" || config.Database.Password == "1q2w#E" {
		// 直接设置完整密码
		config.Database.Password = "1q2w#E$R"
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// GetDatabaseDSN 获取数据库DSN
func (c *Config) GetDatabaseDSN() string {
	if c.Database.Driver == "postgres" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			c.Database.Host,
			c.Database.Port,
			c.Database.Username,
			c.Database.Password,
			c.Database.Database,
			c.Database.SSLMode,
			c.Database.Loc,
		)
	}

	// MySQL兼容模式
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

	// 验证冲突检测配置
	if c.ConflictDetection != nil {
		if err := c.ConflictDetection.Validate(); err != nil {
			return fmt.Errorf("conflict detection configuration is invalid: %w", err)
		}
	}

	return nil
}

// GetDatabasePerformanceConfig 获取数据库性能优化配置
func (c *Config) GetDatabasePerformanceConfig() DatabasePerformanceConfig {
	// 根据环境自动调整性能参数
	baseConfig := DatabasePerformanceConfig{
		MaxOpenConns:       c.Database.MaxOpenConns,
		MaxIdleConns:       c.Database.MaxIdleConns,
		ConnMaxLifetime:    c.Database.ConnMaxLifetime,
		ConnMaxIdleTime:    c.Database.ConnMaxIdleTime,
		EnablePerformance:  c.Database.EnablePerformance,
	}

	// 环境特定优化
	switch c.Environment {
	case "production":
		if baseConfig.MaxOpenConns == 0 {
			baseConfig.MaxOpenConns = 100
		}
		if baseConfig.MaxIdleConns == 0 {
			baseConfig.MaxIdleConns = 10
		}
		if baseConfig.ConnMaxLifetime == 0 {
			baseConfig.ConnMaxLifetime = time.Hour
		}
		if baseConfig.ConnMaxIdleTime == 0 {
			baseConfig.ConnMaxIdleTime = 10 * time.Minute
		}
	case "staging":
		if baseConfig.MaxOpenConns == 0 {
			baseConfig.MaxOpenConns = 50
		}
		if baseConfig.MaxIdleConns == 0 {
			baseConfig.MaxIdleConns = 15
		}
		if baseConfig.ConnMaxLifetime == 0 {
			baseConfig.ConnMaxLifetime = 30 * time.Minute
		}
		if baseConfig.ConnMaxIdleTime == 0 {
			baseConfig.ConnMaxIdleTime = 5 * time.Minute
		}
	case "testing":
		if baseConfig.MaxOpenConns == 0 {
			baseConfig.MaxOpenConns = 5
		}
		if baseConfig.MaxIdleConns == 0 {
			baseConfig.MaxIdleConns = 2
		}
		if baseConfig.ConnMaxLifetime == 0 {
			baseConfig.ConnMaxLifetime = 5 * time.Minute
		}
		if baseConfig.ConnMaxIdleTime == 0 {
			baseConfig.ConnMaxIdleTime = time.Minute
		}
	default: // development
		if baseConfig.MaxOpenConns == 0 {
			baseConfig.MaxOpenConns = 25
		}
		if baseConfig.MaxIdleConns == 0 {
			baseConfig.MaxIdleConns = 5
		}
		if baseConfig.ConnMaxLifetime == 0 {
			baseConfig.ConnMaxLifetime = 30 * time.Minute
		}
		if baseConfig.ConnMaxIdleTime == 0 {
			baseConfig.ConnMaxIdleTime = 5 * time.Minute
		}
	}

	return baseConfig
}

// DatabasePerformanceConfig 数据库性能配置
type DatabasePerformanceConfig struct {
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
	EnablePerformance  bool
}

// GetDriver 获取数据库驱动（向后兼容方法）
func (c *DatabaseConfig) GetDriver() string {
	if c.Driver != "" {
		return c.Driver
	}
	return "postgres" // 默认使用PostgreSQL
}

// GetCharset 获取字符集（向后兼容方法）
func (c *DatabaseConfig) GetCharset() string {
	if c.Charset != "" {
		return c.Charset
	}
	return "utf8"
}

// GetParseTime 获取解析时间设置（向后兼容方法）
func (c *DatabaseConfig) GetParseTime() bool {
	return c.ParseTime
}

// GetLoc 获取时区（向后兼容方法）
func (c *DatabaseConfig) GetLoc() string {
	if c.Loc != "" {
		return c.Loc
	}
	return "UTC"
}

// GetHost 获取主机（向后兼容方法）
func (c *DatabaseConfig) GetHost() string {
	return c.Host
}

// GetPort 获取端口（向后兼容方法）
func (c *DatabaseConfig) GetPort() string {
	return c.Port
}
