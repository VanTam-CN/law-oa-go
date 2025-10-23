package auth

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/law-oa/services/document-service/internal/repositories"
)

// JWTFactory JWT工厂，用于创建JWT相关组件
type JWTFactory struct {
	db     *gorm.DB
	logger *logrus.Logger
	config *JWTConfig
}

// NewJWTFactory 创建JWT工厂
func NewJWTFactory(db *gorm.DB, logger *logrus.Logger, config *JWTConfig) *JWTFactory {
	return &JWTFactory{
		db:     db,
		logger: logger,
		config: config,
	}
}

// CreateJWTManager 创建JWT管理器
func (f *JWTFactory) CreateJWTManager() (*JWTManager, error) {
	// 创建仓库
	tokenRepo := repositories.NewTokenRepository(f.db)
	userRepo := repositories.NewUserRepository(f.db)
	auditRepo := repositories.NewAuditRepository(f.db)

	// 创建JWT管理器
	manager, err := NewJWTManager(f.config, tokenRepo, userRepo, auditRepo, f.logger)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// CreateTokenValidator 创建令牌验证器
func (f *JWTFactory) CreateTokenValidator() *TokenValidator {
	return NewTokenValidator(f.config, f.logger)
}

// CreateJWTMiddleware 创建JWT中间件
func (f *JWTFactory) CreateJWTMiddleware(options *MiddlewareOptions) (*JWTMiddleware, error) {
	// 创建JWT管理器
	manager, err := f.CreateJWTManager()
	if err != nil {
		return nil, err
	}

	// 创建令牌验证器
	validator := f.CreateTokenValidator()

	// 创建中间件
	middleware := NewJWTMiddleware(manager, validator, f.config, f.logger, options)

	return middleware, nil
}

// CreateJWTService 创建完整的JWT服务
func (f *JWTFactory) CreateJWTService(options *MiddlewareOptions) (*JWTService, error) {
	// 创建JWT管理器
	manager, err := f.CreateJWTManager()
	if err != nil {
		return nil, err
	}

	// 创建令牌验证器
	validator := f.CreateTokenValidator()

	// 创建中间件
	middleware := NewJWTMiddleware(manager, validator, f.config, f.logger, options)

	// 创建JWT服务
	service := &JWTService{
		manager:    manager,
		validator:  validator,
		middleware: middleware,
		config:     f.config,
		logger:     f.logger,
	}

	return service, nil
}

// JWTService JWT服务，提供统一的JWT功能接口
type JWTService struct {
	manager    *JWTManager
	validator  *TokenValidator
	middleware *JWTMiddleware
	config     *JWTConfig
	logger     *logrus.Logger
}

// Manager 获取JWT管理器
func (s *JWTService) Manager() *JWTManager {
	return s.manager
}

// Validator 获取令牌验证器
func (s *JWTService) Validator() *TokenValidator {
	return s.validator
}

// Middleware 获取JWT中间件
func (s *JWTService) Middleware() *JWTMiddleware {
	return s.middleware
}

// GenerateTokenPair 生成令牌对
func (s *JWTService) GenerateTokenPair(userID uint, username, tenantID string, roles, permissions []string, ipAddress, userAgent string) (*TokenPair, error) {
	return s.manager.GenerateTokenPair(userID, username, tenantID, roles, permissions, ipAddress, userAgent)
}

// ValidateToken 验证令牌
func (s *JWTService) ValidateToken(tokenString string) (*TokenClaims, error) {
	return s.manager.ValidateToken(tokenString)
}

// RefreshToken 刷新令牌
func (s *JWTService) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	return s.manager.RefreshToken(refreshTokenString)
}

// RevokeToken 撤销令牌
func (s *JWTService) RevokeToken(jti string) error {
	return s.manager.RevokeToken(jti)
}

// RevokeUserTokens 撤销用户所有令牌
func (s *JWTService) RevokeUserTokens(userID uint) error {
	return s.manager.RevokeUserTokens(userID)
}

// RevokeSessionTokens 撤销会话所有令牌
func (s *JWTService) RevokeSessionTokens(sessionID string) error {
	return s.manager.RevokeSessionTokens(sessionID)
}

// IsBlacklisted 检查令牌是否在黑名单中
func (s *JWTService) IsBlacklisted(jti string) bool {
	return s.manager.IsBlacklisted(jti)
}

// GetMiddlewareOptions 获取中间件选项
func (s *JWTService) GetMiddlewareOptions() *MiddlewareOptions {
	return s.middleware.options
}

// GetConfig 获取配置
func (s *JWTService) GetConfig() *JWTConfig {
	return s.config
}

// CleanupExpiredTokens 清理过期令牌
func (s *JWTService) CleanupExpiredTokens() error {
	return s.manager.CleanupExpiredTokens()
}

// RotateKeys 轮换密钥
func (s *JWTService) RotateKeys() error {
	return s.manager.RotateKeys()
}

// GetTokenStats 获取令牌统计信息
func (s *JWTService) GetTokenStats() (*TokenStats, error) {
	return s.manager.GetTokenStats()
}

// TokenStats 令牌统计信息
type TokenStats struct {
	TotalTokens    int64     `json:"total_tokens"`
	ActiveTokens   int64     `json:"active_tokens"`
	ExpiredTokens  int64     `json:"expired_tokens"`
	RevokedTokens  int64     `json:"revoked_tokens"`
	BlacklistedTokens int64   `json:"blacklisted_tokens"`
	LastCleanup    time.Time `json:"last_cleanup"`
}

// CreateDevelopmentJWTFactory 创建开发环境JWT工厂
func CreateDevelopmentJWTFactory(db *gorm.DB, logger *logrus.Logger) *JWTFactory {
	config := &JWTConfig{
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:              "document-service-dev",
		Audience:            "document-service",
		Leeway:              10 * time.Second,
		MaxTokenAge:         24 * time.Hour,
		AccessTokenKeyID:     "access-key-dev",
		RefreshTokenKeyID:    "refresh-key-dev",
	}

	return NewJWTFactory(db, logger, config)
}

// CreateProductionJWTFactory 创建生产环境JWT工厂
func CreateProductionJWTFactory(db *gorm.DB, logger *logrus.Logger, issuer, audience string) *JWTFactory {
	config := &JWTConfig{
		AccessTokenDuration:  30 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:              issuer,
		Audience:            audience,
		Leeway:              30 * time.Second,
		MaxTokenAge:         24 * time.Hour,
		AccessTokenKeyID:     "access-key-prod",
		RefreshTokenKeyID:    "refresh-key-prod",
	}

	return NewJWTFactory(db, logger, config)
}

// DefaultJWTOptions 默认JWT选项
func DefaultJWTOptions() *MiddlewareOptions {
	return DefaultMiddlewareOptions()
}

// DevelopmentJWTOptions 开发环境JWT选项
func DevelopmentJWTOptions() *MiddlewareOptions {
	return &MiddlewareOptions{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/api/v1/auth/login",
			"/api/v1/auth/refresh",
			"/api/v1/auth/register",
			"/debug/*",
		},
		SkipMethods:    []string{"GET", "OPTIONS"},
		ValidateRefresh: false,
		CheckBlacklist: true,
		EnableCache:    false, // 开发环境禁用缓存以便调试
		CacheTTL:       1 * time.Minute,
		Extractor:      MultiExtractor(
			DefaultTokenExtractor,
			CookieTokenExtractor("token"),
			QueryTokenExtractor("token"),
		),
		ErrorHandler: JSONErrorHandler,
	}
}

// ProductionJWTOptions 生产环境JWT选项
func ProductionJWTOptions() *MiddlewareOptions {
	return &MiddlewareOptions{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/api/v1/auth/login",
			"/api/v1/auth/refresh",
		},
		SkipMethods:    []string{"OPTIONS"},
		ValidateRefresh: false,
		CheckBlacklist: true,
		EnableCache:    true,
		CacheTTL:       5 * time.Minute,
		Extractor:      DefaultTokenExtractor,
		ErrorHandler:   DetailedErrorHandler,
	}
}