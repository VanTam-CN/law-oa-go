package config

import (
	"os"
	"path/filepath"
	"testing"

	"law-oa-go/test/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load_DefaultConfig(t *testing.T) {
	t.Run("加载默认配置", func(t *testing.T) {
		// 保存当前环境变量
		originalEnv := saveEnvironmentVariables()
		defer restoreEnvironmentVariables(originalEnv)

		// 清除相关环境变量
		clearConfigEnvironmentVariables()

		// 创建临时配置文件
		configContent := `
environment: test
port: "9090"
database:
  host: "test-host"
  port: "3307"
  username: "test-user"
  password: "test-pass"
  database: "test-db"
redis:
  host: "redis-host"
  port: "6380"
  password: "redis-pass"
  db: 1
jwt:
  secret: "test-jwt-secret-that-is-at-least-32-characters-long"
  expiresIn: 1800
  refreshIn: 3600
`

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 更改工作目录到临时目录
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tempDir)

		// 测试加载配置
		config, err := Load()
		require.NoError(t, err)
		require.NotNil(t, config)

		// 验证配置值
		assert.Equal(t, "test", config.Environment)
		assert.Equal(t, "9090", config.Port)
		assert.Equal(t, "test-host", config.Database.Host)
		assert.Equal(t, "3307", config.Database.Port)
		assert.Equal(t, "test-user", config.Database.Username)
		assert.Equal(t, "test-pass", config.Database.Password)
		assert.Equal(t, "test-db", config.Database.Database)
		assert.Equal(t, "redis-host", config.Redis.Host)
		assert.Equal(t, "6380", config.Redis.Port)
		assert.Equal(t, "redis-pass", config.Redis.Password)
		assert.Equal(t, 1, config.Redis.DB)
		assert.Equal(t, "test-jwt-secret-that-is-at-least-32-characters-long", config.JWT.Secret)
		assert.Equal(t, 1800, config.JWT.ExpiresIn)
		assert.Equal(t, 3600, config.JWT.RefreshIn)
	})
}

func TestConfig_Load_EnvironmentVariables(t *testing.T) {
	t.Run("从环境变量加载配置", func(t *testing.T) {
		// 保存当前环境变量
		originalEnv := saveEnvironmentVariables()
		defer restoreEnvironmentVariables(originalEnv)

		// 设置测试环境变量
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("PORT", "3000")
		os.Setenv("DB_HOST", "env-db-host")
		os.Setenv("DB_PORT", "3308")
		os.Setenv("DB_USERNAME", "env-db-user")
		os.Setenv("DB_PASSWORD", "env-db-pass")
		os.Setenv("DB_DATABASE", "env-db-name")
		os.Setenv("REDIS_HOST", "env-redis-host")
		os.Setenv("REDIS_PORT", "6381")
		os.Setenv("REDIS_PASSWORD", "env-redis-pass")
		os.Setenv("REDIS_DB", "2")
		os.Setenv("JWT_SECRET", "env-jwt-secret-that-is-at-least-32-characters")
		os.Setenv("JWT_EXPIRES_IN", "7200")
		os.Setenv("JWT_REFRESH_IN", "14400")
		os.Setenv("ONLYOFFICE_SECRET", "production-onlyoffice-secret-32-chars-long")
		os.Setenv("ONLYOFFICE_URL", "http://onlyoffice.internal")

		// 更改工作目录到临时目录（确保没有config.yaml文件）
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tempDir)

		// 测试加载配置
		config, err := Load()
		require.NoError(t, err)
		require.NotNil(t, config)

		// 验证环境变量配置
		assert.Equal(t, "production", config.Environment)
		assert.Equal(t, "3000", config.Port)
		assert.Equal(t, "env-db-host", config.Database.Host)
		assert.Equal(t, "3308", config.Database.Port)
		assert.Equal(t, "env-db-user", config.Database.Username)
		assert.Equal(t, "env-db-pass", config.Database.Password)
		assert.Equal(t, "env-db-name", config.Database.Database)
		assert.Equal(t, "env-redis-host", config.Redis.Host)
		assert.Equal(t, "6381", config.Redis.Port)
		assert.Equal(t, "env-redis-pass", config.Redis.Password)
		assert.Equal(t, 2, config.Redis.DB)
		assert.Equal(t, "env-jwt-secret-that-is-at-least-32-characters", config.JWT.Secret)
		assert.Equal(t, 7200, config.JWT.ExpiresIn)
		assert.Equal(t, 14400, config.JWT.RefreshIn)
	})
}

func TestConfig_Load_MissingJWTSecret(t *testing.T) {
	t.Run("缺少JWT密钥", func(t *testing.T) {
		// 保存当前环境变量
		originalEnv := saveEnvironmentVariables()
		defer restoreEnvironmentVariables(originalEnv)

		// 清除JWT密钥环境变量
		os.Unsetenv("JWT_SECRET")

		// 创建没有JWT密钥的配置文件
		configContent := `
environment: test
database:
  host: "test-host"
  username: "test-user"
  database: "test-db"
jwt:
  secret: ""  # 空的JWT密钥
`

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 更改工作目录到临时目录
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tempDir)

		// 测试加载配置应该失败
		_, err = Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret must be configured")
	})
}

func TestConfig_Load_ShortJWTSecret(t *testing.T) {
	t.Run("JWT密钥太短", func(t *testing.T) {
		// 保存当前环境变量
		originalEnv := saveEnvironmentVariables()
		defer restoreEnvironmentVariables(originalEnv)

		// 设置短的JWT密钥
		os.Setenv("JWT_SECRET", "short")

		// 更改工作目录到临时目录
		tempDir := t.TempDir()
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tempDir)

		// 测试加载配置应该失败
		_, err := Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JWT secret must be at least 32 characters long")
	})
}

func TestConfig_Load_IncompleteDatabaseConfig(t *testing.T) {
	t.Run("数据库配置不完整", func(t *testing.T) {
		// 保存当前环境变量
		originalEnv := saveEnvironmentVariables()
		defer restoreEnvironmentVariables(originalEnv)

		// 清除数据库相关环境变量，但保留一些默认值
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_USERNAME")
		os.Unsetenv("DB_DATABASE")

		// 设置有效的JWT密钥
		os.Setenv("JWT_SECRET", "test-jwt-secret-that-is-at-least-32-characters-long")

		// 创建一个配置文件，其中数据库配置不完整
		configContent := `
environment: test
jwt:
  secret: "test-jwt-secret-that-is-at-least-32-characters-long"
database:
  host: ""  # 空的主机
  username: "test-user"  # 有用户名但没有主机和数据库名
`

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 更改工作目录到临时目录
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tempDir)

		// 测试加载配置应该失败
		_, err = Load()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database configuration is incomplete")
	})
}

// TestConfig_Load_ProductionRequiresOnlyOfficeSecret 生产环境必须强制配置 ONLYOFFICE_SECRET
func TestConfig_Load_ProductionRequiresOnlyOfficeSecret(t *testing.T) {
	originalEnv := saveEnvironmentVariables()
	defer restoreEnvironmentVariables(originalEnv)

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("JWT_SECRET", "production-jwt-secret-that-is-at-least-32-chars")
	os.Setenv("DB_HOST", "prod-db")
	os.Setenv("DB_USERNAME", "prod-user")
	os.Setenv("DB_DATABASE", "prod-db")
	os.Unsetenv("ONLYOFFICE_SECRET")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ONLYOFFICE_SECRET must be configured in production")
}

// TestConfig_Load_ProductionRequiresOnlyOfficeSecretMinLength 生产环境 ONLYOFFICE_SECRET 不得少于 32 字符
func TestConfig_Load_ProductionRequiresOnlyOfficeSecretMinLength(t *testing.T) {
	originalEnv := saveEnvironmentVariables()
	defer restoreEnvironmentVariables(originalEnv)

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("JWT_SECRET", "production-jwt-secret-that-is-at-least-32-chars")
	os.Setenv("DB_HOST", "prod-db")
	os.Setenv("DB_USERNAME", "prod-user")
	os.Setenv("DB_DATABASE", "prod-db")
	os.Setenv("ONLYOFFICE_SECRET", "short")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ONLYOFFICE_SECRET must be at least 32 characters")
}

// TestConfig_Load_DevelopmentAllowsEmptyOnlyOfficeSecret 开发环境允许 OnlyOffice 密钥为空
func TestConfig_Load_DevelopmentAllowsEmptyOnlyOfficeSecret(t *testing.T) {
	originalEnv := saveEnvironmentVariables()
	defer restoreEnvironmentVariables(originalEnv)

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	os.Setenv("ENVIRONMENT", "development")
	os.Setenv("JWT_SECRET", "dev-jwt-secret-that-is-at-least-32-characters")
	os.Setenv("DB_HOST", "dev-db")
	os.Setenv("DB_USERNAME", "dev-user")
	os.Setenv("DB_DATABASE", "dev-db")
	os.Unsetenv("ONLYOFFICE_SECRET")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestConfig_GetDatabaseDSN(t *testing.T) {
	t.Run("获取数据库DSN", func(t *testing.T) {
		config := &Config{
			Database: DatabaseConfig{
				Host:      "localhost",
				Port:      "3306",
				Username:  "root",
				Password:  "password",
				Database:  "test_db",
				Charset:   "utf8mb4",
				ParseTime: true,
				Loc:       "Local",
			},
		}

		expectedDSN := "root:password@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=true&loc=Local"
		actualDSN := config.GetDatabaseDSN()
		assert.Equal(t, expectedDSN, actualDSN)
	})
}

func TestConfig_GetRedisAddr(t *testing.T) {
	t.Run("获取Redis地址", func(t *testing.T) {
		config := &Config{
			Redis: RedisConfig{
				Host: "localhost",
				Port: "6379",
			},
		}

		expectedAddr := "localhost:6379"
		actualAddr := config.GetRedisAddr()
		assert.Equal(t, expectedAddr, actualAddr)
	})
}

func TestConfig_GetElasticsearchURL(t *testing.T) {
	t.Run("获取Elasticsearch URL", func(t *testing.T) {
		config := &Config{
			Elasticsearch: ElasticsearchConfig{
				Host: "localhost",
				Port: "9200",
			},
		}

		expectedURL := "http://localhost:9200"
		actualURL := config.GetElasticsearchURL()
		assert.Equal(t, expectedURL, actualURL)
	})
}

func TestConfig_IsProduction(t *testing.T) {
	t.Run("判断生产环境", func(t *testing.T) {
		tests := []struct {
			name        string
			environment string
			expected    bool
		}{
			{"生产环境", "production", true},
			{"开发环境", "development", false},
			{"测试环境", "test", false},
			{"未知环境", "unknown", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &Config{Environment: tt.environment}
				assert.Equal(t, tt.expected, config.IsProduction())
			})
		}
	})
}

func TestConfig_IsDevelopment(t *testing.T) {
	t.Run("判断开发环境", func(t *testing.T) {
		tests := []struct {
			name        string
			environment string
			expected    bool
		}{
			{"开发环境", "development", true},
			{"生产环境", "production", false},
			{"测试环境", "test", false},
			{"未知环境", "unknown", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &Config{Environment: tt.environment}
				assert.Equal(t, tt.expected, config.IsDevelopment())
			})
		}
	})
}

func TestConfig_GetPort(t *testing.T) {
	t.Run("获取端口", func(t *testing.T) {
		tests := []struct {
			name     string
			port     string
			expected string
		}{
			{"指定端口", "9090", "9090"},
			{"空端口", "", "8080"},
			{"默认端口", "8080", "8080"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &Config{Port: tt.port}
				assert.Equal(t, tt.expected, config.GetPort())
			})
		}
	})
}

func TestConfig_Load_WithMockFactory(t *testing.T) {
	t.Run("使用Mock工厂测试配置加载", func(t *testing.T) {
		_ = mock.NewTestDataFactory()

		// 保存当前环境变量
		originalEnv := saveEnvironmentVariables()
		defer restoreEnvironmentVariables(originalEnv)

		// 使用Mock工厂创建测试配置数据
		testConfig := map[string]interface{}{
			"environment": "test",
			"port":        "9090",
			"database": map[string]interface{}{
				"host":     "mock-host",
				"port":     "3307",
				"username": "mock-user",
				"password": "mock-pass",
				"database": "mock-db",
			},
			"redis": map[string]interface{}{
				"host": "mock-redis",
				"port": "6380",
			},
			"jwt": map[string]interface{}{
				"secret":    "mock-jwt-secret-that-is-at-least-32-characters-long",
				"expiresIn": 1800,
			},
		}

		// 创建YAML配置内容
		configContent := convertMapToYAML(testConfig)

		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 更改工作目录到临时目录
		originalWd, _ := os.Getwd()
		defer os.Chdir(originalWd)
		os.Chdir(tempDir)

		// 测试加载配置
		config, err := Load()
		require.NoError(t, err)
		require.NotNil(t, config)

		// 使用标准断言验证配置
		assert.NotNil(t, config, "配置对象不应该为空")
		assert.Equal(t, "test", config.Environment, "环境应该为test")
		assert.Equal(t, "9090", config.Port, "端口应该为9090")
		assert.Equal(t, "mock-host", config.Database.Host, "数据库主机应该为mock-host")
		assert.Equal(t, "mock-redis", config.Redis.Host, "Redis主机应该为mock-redis")
		assert.Equal(t, "mock-jwt-secret-that-is-at-least-32-characters-long", config.JWT.Secret, "JWT密钥应该正确")
	})
}

// 辅助函数

// saveEnvironmentVariables 保存当前环境变量
func saveEnvironmentVariables() map[string]string {
	envVars := []string{
		"ENVIRONMENT", "PORT",
		"DB_HOST", "DB_PORT", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"ES_HOST", "ES_PORT", "ES_USERNAME", "ES_PASSWORD",
		"JWT_SECRET", "JWT_EXPIRES_IN", "JWT_REFRESH_IN",
		"ONLYOFFICE_URL", "ONLYOFFICE_SECRET", "BACKEND_URL",
	}

	saved := make(map[string]string)
	for _, env := range envVars {
		if value := os.Getenv(env); value != "" {
			saved[env] = value
		}
	}
	return saved
}

// restoreEnvironmentVariables 恢复环境变量
func restoreEnvironmentVariables(saved map[string]string) {
	// 清除相关环境变量
	envVars := []string{
		"ENVIRONMENT", "PORT",
		"DB_HOST", "DB_PORT", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"ES_HOST", "ES_PORT", "ES_USERNAME", "ES_PASSWORD",
		"JWT_SECRET", "JWT_EXPIRES_IN", "JWT_REFRESH_IN",
		"ONLYOFFICE_URL", "ONLYOFFICE_SECRET", "BACKEND_URL",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}

	// 恢复保存的环境变量
	for key, value := range saved {
		os.Setenv(key, value)
	}
}

// clearConfigEnvironmentVariables 清除配置相关的环境变量
func clearConfigEnvironmentVariables() {
	envVars := []string{
		"ENVIRONMENT", "PORT",
		"DB_HOST", "DB_PORT", "DB_USERNAME", "DB_PASSWORD", "DB_DATABASE",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"ES_HOST", "ES_PORT", "ES_USERNAME", "ES_PASSWORD",
		"JWT_SECRET", "JWT_EXPIRES_IN", "JWT_REFRESH_IN",
		"ONLYOFFICE_URL", "ONLYOFFICE_SECRET", "BACKEND_URL",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

// convertMapToYAML 将map转换为YAML字符串（简化版）
func convertMapToYAML(data map[string]interface{}) string {
	result := ""

	for key, value := range data {
		switch v := value.(type) {
		case string:
			result += key + ": " + v + "\n"
		case map[string]interface{}:
			result += key + ":\n"
			for subKey, subValue := range v {
				if subStr, ok := subValue.(string); ok {
					result += "  " + subKey + ": " + subStr + "\n"
				}
			}
		}
	}

	return result
}

// Benchmark tests
func BenchmarkConfig_Load(b *testing.B) {
	// 保存当前环境变量
	originalEnv := saveEnvironmentVariables()
	defer restoreEnvironmentVariables(originalEnv)

	// 设置测试环境变量
	os.Setenv("JWT_SECRET", "benchmark-jwt-secret-that-is-at-least-32-characters-long")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Load()
		if err != nil {
			b.Fatalf("加载配置失败: %v", err)
		}
	}
}

func BenchmarkConfig_GetDatabaseDSN(b *testing.B) {
	config := &Config{
		Database: DatabaseConfig{
			Host:      "localhost",
			Port:      "3306",
			Username:  "root",
			Password:  "password",
			Database:  "test_db",
			Charset:   "utf8mb4",
			ParseTime: true,
			Loc:       "Local",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = config.GetDatabaseDSN()
	}
}

func BenchmarkConfig_GetRedisAddr(b *testing.B) {
	config := &Config{
		Redis: RedisConfig{
			Host: "localhost",
			Port: "6379",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = config.GetRedisAddr()
	}
}
