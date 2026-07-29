package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Environment       string                   `mapstructure:"environment"`
	Port              string                   `mapstructure:"port"`
	Database          DatabaseConfig           `mapstructure:"database"`
	Redis             RedisConfig              `mapstructure:"redis"`
	Elasticsearch     ElasticsearchConfig      `mapstructure:"elasticsearch"`
	JWT               JWTConfig                `mapstructure:"jwt"`
	Log               LogConfig                `mapstructure:"log"`
	CORS              CORSConfig               `mapstructure:"cors"`
	ConflictDetection *ConflictDetectionConfig `mapstructure:"conflictDetection"`
	OnlyOffice        OnlyOfficeConfig         `mapstructure:"onlyoffice"`
}

// OnlyOfficeConfig OnlyOffice 在线编辑集成配置
type OnlyOfficeConfig struct {
	URL        string `mapstructure:"url"`
	Secret     string `mapstructure:"secret"`
	BackendURL string `mapstructure:"backendUrl"`
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
	MaxOpenConns      int           `mapstructure:"maxOpen_conns"`
	MaxIdleConns      int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime   time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime   time.Duration `mapstructure:"conn_max_idle_time"`
	EnablePerformance bool          `mapstructure:"enable_performance"`
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

// Load 加载应用运行配置，并执行完整的应用校验。
func Load() (*Config, error) {
	return load(true)
}

// LoadForMigration 加载数据库初始化工具所需的配置。
//
// 数据库 bootstrap 不会启动 HTTP 服务、OnlyOffice 回调或 JWT 认证，
// 因此不应要求这些运行时配置已经存在。数据库连接本身仍会执行最小
// 的驱动、地址、账号和库名校验；应用服务启动时继续使用 Load 的完整校验。
func LoadForMigration() (*Config, error) {
	return load(false)
}

func load(validateApplication bool) (*Config, error) {
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
	viper.SetDefault("database.password", "")
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

	// OnlyOffice 默认配置（生产环境必须显式覆盖 Secret）
	viper.SetDefault("onlyoffice.url", "http://localhost:9090")
	viper.SetDefault("onlyoffice.secret", "")
	viper.SetDefault("onlyoffice.backendUrl", "http://localhost:8080")

	// 数据库性能配置默认值
	viper.SetDefault("database.maxOpen_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "30m")
	viper.SetDefault("database.conn_max_idle_time", "5m")
	viper.SetDefault("database.enable_performance", true)

	// 冲突检测配置默认值
	viper.SetDefault("conflictDetection.enabled", true)
	viper.SetDefault("conflictDetection.autoCheckOnCaseCreation", true)
	// 0 means full historical coverage. A five-year default would create a
	// false sense of safety for former clients and is not allowed by the P0 spec.
	viper.SetDefault("conflictDetection.defaultSearchYears", 0)
	viper.SetDefault("conflictDetection.defaultSearchDepth", "deep")
	viper.SetDefault("conflictDetection.includeCorporateRelations", true)
	viper.SetDefault("conflictDetection.highRiskThreshold", 0.7)
	viper.SetDefault("conflictDetection.mediumRiskThreshold", 0.4)
	viper.SetDefault("conflictDetection.requireApprovalForHighRisk", true)
	// A conflict check is a hard gate in the P0 workflow. Any exception must
	// be an explicit, audited policy decision rather than a default setting.
	viper.SetDefault("conflictDetection.allowSkipConflictCheck", false)

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
		"database.sslmode":       "DB_SSLMODE",
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
		"onlyoffice.url":         "ONLYOFFICE_URL",
		"onlyoffice.secret":      "ONLYOFFICE_SECRET",
		"onlyoffice.backendUrl":  "BACKEND_URL",
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
	config.Database.Driver = strings.ToLower(strings.TrimSpace(config.Database.Driver))

	if validateApplication {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	} else if err := validateMigrationConfig(&config); err != nil {
		return nil, fmt.Errorf("migration config validation failed: %w", err)
	}

	return &config, nil
}

func validateMigrationConfig(c *Config) error {
	if c == nil {
		return fmt.Errorf("configuration is nil")
	}

	switch strings.ToLower(strings.TrimSpace(c.Database.Driver)) {
	case "postgres", "postgresql", "mysql":
	default:
		return fmt.Errorf("unsupported database driver %q", c.Database.Driver)
	}
	if strings.TrimSpace(c.Database.Host) == "" ||
		strings.TrimSpace(c.Database.Username) == "" ||
		strings.TrimSpace(c.Database.Database) == "" {
		return fmt.Errorf("database configuration is incomplete")
	}
	if c.IsProduction() {
		if isDefaultDatabasePassword(c.Database.Password) {
			return fmt.Errorf("database password must be configured and cannot use a default value in production")
		}
		driver := strings.ToLower(strings.TrimSpace(c.Database.Driver))
		if (driver == "postgres" || driver == "postgresql") && (strings.TrimSpace(c.Database.SSLMode) == "" || strings.EqualFold(strings.TrimSpace(c.Database.SSLMode), "disable")) {
			return fmt.Errorf("PostgreSQL sslmode must not be disabled in production")
		}
	}
	return nil
}

// GetDatabaseDSN 获取数据库DSN
func (c *Config) GetDatabaseDSN() string {
	if strings.EqualFold(strings.TrimSpace(c.Database.Driver), "postgres") || strings.EqualFold(strings.TrimSpace(c.Database.Driver), "postgresql") {
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
	return c != nil && strings.EqualFold(strings.TrimSpace(c.Environment), "production")
}

// ValidateProductionReadiness checks controls that cannot be represented by
// the legacy config file alone. The server calls this before opening any
// production listener; migration tooling uses LoadForMigration so an operator
// can bootstrap a database before the application starts.
func (c *Config) ValidateProductionReadiness() error {
	if c == nil || !c.IsProduction() {
		return nil
	}

	appSecret := strings.TrimSpace(os.Getenv("APP_SECRET"))
	if len(appSecret) < 32 || isDefaultSecret(appSecret) {
		return fmt.Errorf("APP_SECRET must be configured with a non-default value of at least 32 characters in production")
	}

	if !validSubjectDataKey(os.Getenv("SUBJECT_DATA_KEY")) {
		return fmt.Errorf("SUBJECT_DATA_KEY must decode to exactly 32 bytes in production")
	}

	if strings.EqualFold(strings.TrimSpace(os.Getenv("DEBUG")), "true") {
		return fmt.Errorf("DEBUG must be false or unset in production")
	}

	if c.ConflictDetection != nil && c.ConflictDetection.AllowSkipConflictCheck {
		return fmt.Errorf("conflictDetection.allowSkipConflictCheck must be false in production")
	}

	origins := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if origins == "" {
		origins = strings.Join(c.CORS.AllowedOrigins, ",")
	}
	if origins == "" {
		return fmt.Errorf("CORS_ALLOWED_ORIGINS must contain the production frontend origin")
	}
	for _, origin := range strings.Split(origins, ",") {
		origin = strings.ToLower(strings.TrimSpace(origin))
		if origin == "" || strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") || strings.Contains(origin, "0.0.0.0") {
			return fmt.Errorf("CORS_ALLOWED_ORIGINS cannot contain local or wildcard origins in production")
		}
	}
	return nil
}

func validSubjectDataKey(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return true
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return true
	}
	if decoded, err := hex.DecodeString(raw); err == nil && len(decoded) == 32 {
		return true
	}
	return len(raw) == 32
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
	if isDefaultSecret(c.JWT.Secret) {
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

	// 生产环境强制校验 OnlyOffice 回调密钥，避免回调被伪造
	if c.IsProduction() {
		if isDefaultDatabasePassword(c.Database.Password) {
			return fmt.Errorf("database password must be configured and cannot use a default value in production")
		}
		if strings.EqualFold(c.Database.Driver, "postgres") || strings.EqualFold(c.Database.Driver, "postgresql") {
			sslMode := strings.ToLower(strings.TrimSpace(c.Database.SSLMode))
			if sslMode == "" || sslMode == "disable" {
				return fmt.Errorf("PostgreSQL sslmode must not be disabled in production")
			}
		}
		if c.OnlyOffice.Secret == "" {
			return fmt.Errorf("ONLYOFFICE_SECRET must be configured in production")
		}
		if len(c.OnlyOffice.Secret) < 32 {
			return fmt.Errorf("ONLYOFFICE_SECRET must be at least 32 characters long in production")
		}
		if c.OnlyOffice.URL == "" {
			return fmt.Errorf("ONLYOFFICE_URL must be configured in production")
		}
	}

	return nil
}

func isDefaultSecret(secret string) bool {
	normalized := strings.ToLower(strings.TrimSpace(secret))
	if normalized == "" {
		return true
	}
	defaults := []string{
		"your-secret-key",
		"your-secret-key-change-in-production",
		"your-secret-key-change-in-production-please-use-at-least-32-characters",
		"your-jwt-secret-key-change-in-production",
		"your-very-secure-jwt-secret-key-that-is-at-least-32-characters-long-for-production",
	}
	for _, value := range defaults {
		if normalized == value {
			return true
		}
	}
	return strings.Contains(normalized, "change-in-production") ||
		strings.Contains(normalized, "change-before-production") ||
		strings.Contains(normalized, "local-dev") ||
		strings.Contains(normalized, "development-only") ||
		strings.Contains(normalized, "your-secret") ||
		strings.Contains(normalized, "your-jwt-secret") ||
		strings.Contains(normalized, "default")
}

func isDefaultDatabasePassword(password string) bool {
	switch strings.ToLower(strings.TrimSpace(password)) {
	case "", "password", "law_oa_password", "lawpass", "1q2w", "1q2w#e", "1q2w#e$r":
		return true
	default:
		normalized := strings.ToLower(strings.TrimSpace(password))
		return strings.Contains(normalized, "local-dev") ||
			strings.Contains(normalized, "change-before-production") ||
			strings.Contains(normalized, "change-in-production")
	}
}

// GetDatabasePerformanceConfig 获取数据库性能优化配置
func (c *Config) GetDatabasePerformanceConfig() DatabasePerformanceConfig {
	// 根据环境自动调整性能参数
	baseConfig := DatabasePerformanceConfig{
		MaxOpenConns:      c.Database.MaxOpenConns,
		MaxIdleConns:      c.Database.MaxIdleConns,
		ConnMaxLifetime:   c.Database.ConnMaxLifetime,
		ConnMaxIdleTime:   c.Database.ConnMaxIdleTime,
		EnablePerformance: c.Database.EnablePerformance,
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
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	ConnMaxIdleTime   time.Duration
	EnablePerformance bool
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
