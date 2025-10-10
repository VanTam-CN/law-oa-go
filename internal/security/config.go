package security

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
	"law-oa-go/internal/cache"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// JWT配置
	JWT JWTConfig `json:"jwt" yaml:"jwt"`

	// 加密配置
	Encryption EncryptionConfig `json:"encryption" yaml:"encryption"`

	// API安全配置
	APISecurity APISecurityConfig `json:"api_security" yaml:"api_security"`

	// 认证配置
	Auth AuthConfig `json:"auth" yaml:"auth"`

	// 审计配置
	Audit AuditLogConfig `json:"audit" yaml:"audit"`

	// RBAC配置
	RBAC RBACConfig `json:"rbac" yaml:"rbac"`

	// 验证配置
	Validation ValidationConfig `json:"validation" yaml:"validation"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	AccessTokenTTL  time.Duration `json:"access_token_ttl" yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `json:"refresh_token_ttl" yaml:"refresh_token_ttl"`
	Issuer          string        `json:"issuer" yaml:"issuer"`
	Secret          string        `json:"secret" yaml:"secret"`
	EnableRefresh   bool          `json:"enable_refresh" yaml:"enable_refresh"`
	BlacklistTTL    time.Duration `json:"blacklist_ttl" yaml:"blacklist_ttl"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	AESKey                string   `json:"aes_key" yaml:"aes_key"`
	RSAPrivateKey         string   `json:"rsa_private_key" yaml:"rsa_private_key"`
	RSAPublicKey          string   `json:"rsa_public_key" yaml:"rsa_public_key"`
	DataKeyRotationDays   int      `json:"data_key_rotation_days" yaml:"data_key_rotation_days"`
	EnableFieldEncryption bool     `json:"enable_field_encryption" yaml:"enable_field_encryption"`
	SensitiveFields       []string `json:"sensitive_fields" yaml:"sensitive_fields"`
	HashAlgorithm         string   `json:"hash_algorithm" yaml:"hash_algorithm"`
}

// APISecurityConfig API安全配置
type APISecurityConfig struct {
	EnableRateLimit         bool          `json:"enable_rate_limit" yaml:"enable_rate_limit"`
	EnableIPWhitelist       bool          `json:"enable_ip_whitelist" yaml:"enable_ip_whitelist"`
	EnableIPBlacklist       bool          `json:"enable_ip_blacklist" yaml:"enable_ip_blacklist"`
	EnableRequestSigning    bool          `json:"enable_request_signing" yaml:"enable_request_signing"`
	EnableAPIThrottling     bool          `json:"enable_api_throttling" yaml:"enable_api_throttling"`
	EnableWAFProtection     bool          `json:"enable_waf_protection" yaml:"enable_waf_protection"`
	EnableDDoSProtection    bool          `json:"enable_ddos_protection" yaml:"enable_ddos_protection"`
	EnableRequestValidation bool          `json:"enable_request_validation" yaml:"enable_request_validation"`
	EnableCORS              bool          `json:"enable_cors" yaml:"enable_cors"`
	EnableCSRF              bool          `json:"enable_csrf" yaml:"enable_csrf"`
	RateLimitWindow         time.Duration `json:"rate_limit_window" yaml:"rate_limit_window"`
	RateLimitMaxRequests    int           `json:"rate_limit_max_requests" yaml:"rate_limit_max_requests"`
	WhitelistedIPs          []string      `json:"whitelisted_ips" yaml:"whitelisted_ips"`
	BlacklistedIPs          []string      `json:"blacklisted_ips" yaml:"blacklisted_ips"`
	AllowedOrigins          []string      `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods          []string      `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders          []string      `json:"allowed_headers" yaml:"allowed_headers"`
	MaxRequestSize          int64         `json:"max_request_size" yaml:"max_request_size"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	EnableDeviceCheck     bool           `json:"enable_device_check" yaml:"enable_device_check"`
	EnableIPCheck         bool           `json:"enable_ip_check" yaml:"enable_ip_check"`
	EnableRateLimit       bool           `json:"enable_rate_limit" yaml:"enable_rate_limit"`
	SkipAuthPaths         []string       `json:"skip_auth_paths" yaml:"skip_auth_paths"`
	SkipAuthPrefixes      []string       `json:"skip_auth_prefixes" yaml:"skip_auth_prefixes"`
	RequiredRoles         []string       `json:"required_roles" yaml:"required_roles"`
	SessionTimeout        time.Duration  `json:"session_timeout" yaml:"session_timeout"`
	MaxConcurrentSessions int            `json:"max_concurrent_sessions" yaml:"max_concurrent_sessions"`
	PasswordPolicy        PasswordPolicy `json:"password_policy" yaml:"password_policy"`
}

// PasswordPolicy 密码策略
type PasswordPolicy struct {
	MinLength             int           `json:"min_length" yaml:"min_length"`
	MaxLength             int           `json:"max_length" yaml:"max_length"`
	RequireUppercase      bool          `json:"require_uppercase" yaml:"require_uppercase"`
	RequireLowercase      bool          `json:"require_lowercase" yaml:"require_lowercase"`
	RequireNumbers        bool          `json:"require_numbers" yaml:"require_numbers"`
	RequireSpecialChars   bool          `json:"require_special_chars" yaml:"require_special_chars"`
	ForbidCommonPasswords bool          `json:"forbid_common_passwords" yaml:"forbid_common_passwords"`
	ForbidPersonalInfo    bool          `json:"forbid_personal_info" yaml:"forbid_personal_info"`
	ExpiryDays            int           `json:"expiry_days" yaml:"expiry_days"`
	HistoryCount          int           `json:"history_count" yaml:"history_count"`
	FailedAttempts        int           `json:"failed_attempts" yaml:"failed_attempts"`
	LockoutDuration       time.Duration `json:"lockout_duration" yaml:"lockout_duration"`
}

// AuditConfig 审计配置
type AuditConfig struct {
	EnableAuditLog       bool             `json:"enable_audit_log" yaml:"enable_audit_log"`
	LogDatabase          bool             `json:"log_database" yaml:"log_database"`
	LogToFile            bool             `json:"log_to_file" yaml:"log_to_file"`
	LogToSyslog          bool             `json:"log_to_syslog" yaml:"log_to_syslog"`
	EnableRealTimeAlert  bool             `json:"enable_real_time_alert" yaml:"enable_real_time_alert"`
	SensitiveEventTypes  []AuditEventType `json:"sensitive_event_types" yaml:"sensitive_event_types"`
	RequiredEventTypes   []AuditEventType `json:"required_event_types" yaml:"required_event_types"`
	RetentionDays        int              `json:"retention_days" yaml:"retention_days"`
	MaxBatchSize         int              `json:"max_batch_size" yaml:"max_batch_size"`
	BatchTimeout         time.Duration    `json:"batch_timeout" yaml:"batch_timeout"`
	EnableCompression    bool             `json:"enable_compression" yaml:"enable_compression"`
	EncryptSensitiveData bool             `json:"encrypt_sensitive_data" yaml:"encrypt_sensitive_data"`
}

// RBACConfig RBAC配置
type RBACConfig struct {
	EnableRBAC              bool          `json:"enable_rbac" yaml:"enable_rbac"`
	EnablePermissionCache   bool          `json:"enable_permission_cache" yaml:"enable_permission_cache"`
	PermissionCacheTTL      time.Duration `json:"permission_cache_ttl" yaml:"permission_cache_ttl"`
	EnableRoleHierarchy     bool          `json:"enable_role_hierarchy" yaml:"enable_role_hierarchy"`
	EnableDynamicRoles      bool          `json:"enable_dynamic_roles" yaml:"enable_dynamic_roles"`
	DefaultRoles            []string      `json:"default_roles" yaml:"default_roles"`
	SuperAdminRoles         []string      `json:"super_admin_roles" yaml:"super_admin_roles"`
	EnablePermissionLogging bool          `json:"enable_permission_logging" yaml:"enable_permission_logging"`
}

// ValidationConfig 验证配置
type ValidationConfig struct {
	EnableInputValidation   bool                   `json:"enable_input_validation" yaml:"enable_input_validation"`
	EnableSQLInjectionCheck bool                   `json:"enable_sql_injection_check" yaml:"enable_sql_injection_check"`
	EnableXSSCheck          bool                   `json:"enable_xss_check" yaml:"enable_xss_check"`
	ValidationRules         map[string]interface{} `json:"validation_rules" yaml:"validation_rules"`
	CustomValidators        []string               `json:"custom_validators" yaml:"custom_validators"`
	MaxStringLength         int                    `json:"max_string_length" yaml:"max_string_length"`
	MaxFileSize             int64                  `json:"max_file_size" yaml:"max_file_size"`
	AllowedFileTypes        []string               `json:"allowed_file_types" yaml:"allowed_file_types"`
}

// ConfigManager 配置管理器
type ConfigManager struct {
	config     *SecurityConfig
	configFile string
	cache      *cache.CacheService
	mu         sync.RWMutex
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configFile string, cacheService *cache.CacheService) *ConfigManager {
	return &ConfigManager{
		configFile: configFile,
		cache:      cacheService,
		config:     &SecurityConfig{},
	}
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 设置默认值
	cm.setDefaultConfig()

	// 使用viper加载配置
	viper.SetConfigFile(cm.configFile)
	viper.SetConfigType("yaml")

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(cm.configFile); os.IsNotExist(err) {
		log.Printf("Config file not found, creating default config: %s", cm.configFile)
		if err := cm.SaveConfig(); err != nil {
			return fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	if err := viper.Unmarshal(cm.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := cm.validateConfig(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 缓存配置
	cm.cacheConfig()

	log.Println("Security configuration loaded successfully")
	return nil
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 验证配置
	if err := cm.validateConfig(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 序列化配置为YAML
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件（JSON格式，但文件扩展名还是.yaml，这样viper可以读取）
	if err := os.WriteFile(cm.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 更新缓存
	cm.cacheConfig()

	log.Println("Security configuration saved successfully")
	return nil
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig() *SecurityConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 尝试从缓存获取
	var cachedConfig SecurityConfig
	if err := cm.cache.Get("security_config", &cachedConfig); err == nil {
		return &cachedConfig
	}

	// 返回内存中的配置
	return cm.config
}

// UpdateConfig 更新配置
func (cm *ConfigManager) UpdateConfig(updates *SecurityConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 更新配置
	cm.config = updates

	// 保存配置
	if err := cm.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}

	return nil
}

// setDefaultConfig 设置默认配置
func (cm *ConfigManager) setDefaultConfig() {
	cm.config = &SecurityConfig{
		JWT: JWTConfig{
			AccessTokenTTL:  2 * time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
			Issuer:          "law-oa-system",
			Secret:          generateRandomKey(32),
			EnableRefresh:   true,
			BlacklistTTL:    24 * time.Hour,
		},
		Encryption: EncryptionConfig{
			AESKey:                generateRandomKey(32),
			DataKeyRotationDays:   90,
			EnableFieldEncryption: true,
			SensitiveFields:       []string{"email", "phone", "id_card", "address", "bank_account"},
			HashAlgorithm:         "sha256",
		},
		APISecurity: APISecurityConfig{
			EnableRateLimit:         true,
			EnableIPWhitelist:       false,
			EnableIPBlacklist:       true,
			EnableRequestSigning:    false,
			EnableAPIThrottling:     true,
			EnableWAFProtection:     true,
			EnableDDoSProtection:    true,
			EnableRequestValidation: true,
			EnableCORS:              true,
			EnableCSRF:              true,
			RateLimitWindow:         time.Minute,
			RateLimitMaxRequests:    100,
			WhitelistedIPs:          []string{},
			BlacklistedIPs:          []string{},
			AllowedOrigins:          []string{"http://localhost:3003"},
			AllowedMethods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:          []string{"*"},
			MaxRequestSize:          10 * 1024 * 1024, // 10MB
		},
		Auth: AuthConfig{
			EnableDeviceCheck:     true,
			EnableIPCheck:         true,
			EnableRateLimit:       true,
			SkipAuthPaths:         []string{"/health", "/metrics"},
			SkipAuthPrefixes:      []string{"/api/public/"},
			SessionTimeout:        24 * time.Hour,
			MaxConcurrentSessions: 3,
			PasswordPolicy: PasswordPolicy{
				MinLength:             8,
				MaxLength:             128,
				RequireUppercase:      true,
				RequireLowercase:      true,
				RequireNumbers:        true,
				RequireSpecialChars:   true,
				ForbidCommonPasswords: true,
				ForbidPersonalInfo:    true,
				ExpiryDays:            90,
				HistoryCount:          5,
				FailedAttempts:        5,
				LockoutDuration:       30 * time.Minute,
			},
		},
		Audit: AuditLogConfig{
			EnableAuditLog:       true,
			LogDatabase:          true,
			LogToFile:            true,
			LogToSyslog:          false,
			EnableRealTimeAlert:  true,
			SensitiveEventTypes:  []AuditEventType{EventTypeSecurityEvent, EventTypePermissionChange, EventTypeDataDelete},
			RequiredEventTypes:   []AuditEventType{EventTypeLogin, EventTypeLogout, EventTypePermissionChange},
			RetentionDays:        365,
			MaxBatchSize:         100,
			BatchTimeout:         5 * time.Second,
			EnableCompression:    true,
			EncryptSensitiveData: true,
		},
		RBAC: RBACConfig{
			EnableRBAC:              true,
			EnablePermissionCache:   true,
			PermissionCacheTTL:      time.Hour,
			EnableRoleHierarchy:     true,
			EnableDynamicRoles:      false,
			DefaultRoles:            []string{"user"},
			SuperAdminRoles:         []string{"super_admin"},
			EnablePermissionLogging: true,
		},
		Validation: ValidationConfig{
			EnableInputValidation:   true,
			EnableSQLInjectionCheck: true,
			EnableXSSCheck:          true,
			MaxStringLength:         1000,
			MaxFileSize:             50 * 1024 * 1024, // 50MB
			AllowedFileTypes:        []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx", ".xls", ".xlsx"},
		},
	}
}

// validateConfig 验证配置
func (cm *ConfigManager) validateConfig() error {
	config := cm.config

	// 验证JWT配置
	if config.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("JWT access token TTL must be positive")
	}
	if config.JWT.RefreshTokenTTL <= config.JWT.AccessTokenTTL {
		return fmt.Errorf("JWT refresh token TTL must be greater than access token TTL")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	// 验证加密配置
	if config.Encryption.EnableFieldEncryption && len(config.Encryption.SensitiveFields) == 0 {
		return fmt.Errorf("sensitive fields must be specified when field encryption is enabled")
	}
	if config.Encryption.DataKeyRotationDays <= 0 {
		return fmt.Errorf("data key rotation days must be positive")
	}

	// 验证API安全配置
	if config.APISecurity.EnableRateLimit && config.APISecurity.RateLimitMaxRequests <= 0 {
		return fmt.Errorf("rate limit max requests must be positive")
	}
	if config.APISecurity.MaxRequestSize <= 0 {
		return fmt.Errorf("max request size must be positive")
	}

	// 验证密码策略
	if config.Auth.PasswordPolicy.MinLength < 8 {
		return fmt.Errorf("password minimum length must be at least 8")
	}
	if config.Auth.PasswordPolicy.MaxLength < config.Auth.PasswordPolicy.MinLength {
		return fmt.Errorf("password maximum length must be greater than minimum length")
	}
	if config.Auth.PasswordPolicy.ExpiryDays < 0 {
		return fmt.Errorf("password expiry days cannot be negative")
	}

	// 验证审计配置
	if config.Audit.EnableAuditLog && config.Audit.MaxBatchSize <= 0 {
		return fmt.Errorf("audit batch size must be positive")
	}
	if config.Audit.RetentionDays <= 0 {
		return fmt.Errorf("audit retention days must be positive")
	}

	// 验证RBAC配置
	if config.RBAC.EnablePermissionCache && config.RBAC.PermissionCacheTTL <= 0 {
		return fmt.Errorf("permission cache TTL must be positive")
	}

	// 验证验证配置
	if config.Validation.MaxStringLength <= 0 {
		return fmt.Errorf("max string length must be positive")
	}
	if config.Validation.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive")
	}

	return nil
}

// cacheConfig 缓存配置
func (cm *ConfigManager) cacheConfig() {
	if cm.cache != nil {
		cm.cache.Set("security_config", cm.config, time.Hour)
	}
}

// generateRandomKey 生成随机密钥
func generateRandomKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// 如果随机数生成失败，使用简单的伪随机
		for i := range b {
			b[i] = charset[i%len(charset)]
		}
	} else {
		// 将随机字节映射到字符集
		for i := range b {
			b[i] = charset[b[i]%byte(len(charset))]
		}
	}
	return string(b)
}

// ReloadConfig 重新加载配置
func (cm *ConfigManager) ReloadConfig() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.LoadConfig()
}

// GetJWTConfig 获取JWT配置
func (cm *ConfigManager) GetJWTConfig() JWTConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.JWT
}

// GetEncryptionConfig 获取加密配置
func (cm *ConfigManager) GetEncryptionConfig() EncryptionConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Encryption
}

// GetAPISecurityConfig 获取API安全配置
func (cm *ConfigManager) GetAPISecurityConfig() APISecurityConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.APISecurity
}

// GetAuthConfig 获取认证配置
func (cm *ConfigManager) GetAuthConfig() AuthConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Auth
}

// GetAuditConfig 获取审计配置
func (cm *ConfigManager) GetAuditConfig() AuditLogConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Audit
}

// GetRBACConfig 获取RBAC配置
func (cm *ConfigManager) GetRBACConfig() RBACConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.RBAC
}

// GetValidationConfig 获取验证配置
func (cm *ConfigManager) GetValidationConfig() ValidationConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.Validation
}

// ExportConfig 导出配置
func (cm *ConfigManager) ExportConfig() ([]byte, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return json.MarshalIndent(cm.config, "", "  ")
}

// ImportConfig 导入配置
func (cm *ConfigManager) ImportConfig(data []byte) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var newConfig SecurityConfig
	if err := json.Unmarshal(data, &newConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cm.config = &newConfig

	return cm.SaveConfig()
}

// GetConfigHash 获取配置哈希
func (cm *ConfigManager) GetConfigHash() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, _ := json.Marshal(cm.config)
	return fmt.Sprintf("%x", data)
}
