package auth

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// JWTManager JWT管理器
type JWTManager struct {
	// 密钥管理
	accessKeyPair  *KeyPair
	refreshKeyPair *KeyPair

	// 配置
	config *JWTConfig

	// 仓库
	tokenRepo      repositories.TokenRepository
	userRepo       repositories.UserRepository
	auditRepo      repositories.AuditRepository

	// 工具
	blacklist      *TokenBlacklist
	validator      *TokenValidator

	// 日志
	logger *logrus.Logger
}

// JWTConfig JWT配置
type JWTConfig struct {
	// 访问令牌配置
	AccessTokenTTL    time.Duration `json:"access_token_ttl" yaml:"access_token_ttl"`
	RefreshTokenTTL   time.Duration `json:"refresh_token_ttl" yaml:"refresh_token_ttl"`

	// 算法配置
	AccessTokenAlg  string `json:"access_token_alg" yaml:"access_token_alg"`
	RefreshTokenAlg string `json:"refresh_token_alg" yaml:"refresh_token_alg"`

	// 签发配置
	AccessTokenSignMethod  jwt.SigningMethod `json:"-" yaml:"-"`
	RefreshTokenSignMethod jwt.SigningMethod `json:"-" yaml:"-"`

	// 密钥配置
	KeyRotationInterval time.Duration `json:"key_rotation_interval" yaml:"key_rotation_interval"`
	KeySizeBits        int             `json:"key_size_bits" yaml:"key_size_bits"`

	// 安全配置
	Issuer               string        `json:"issuer" yaml:"issuer"`
	Audience             string        `json:"audience" yaml:"audience"`
	Leeway               time.Duration `json:"leeway" yaml:"leeway"`
	MaxTokenAge          time.Duration `json:"max_token_age" yaml:"max_token_age"`
	BlacklistEnabled     bool          `json:"blacklist_enabled" yaml:"blacklist_enabled"`

	// 性能配置
	CacheEnabled         bool          `json:"cache_enabled" yaml:"cache_enabled"`
	CacheTTL             time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// 监控配置
	MetricsEnabled       bool          `json:"metrics_enabled" yaml:"metrics_enabled"`
	AuditEnabled         bool          `json:"audit_enabled" yaml:"audit_enabled"`
}

// KeyPair 密钥对
type KeyPair struct {
	PublicKey  interface{} `json:"public_key"`
	PrivateKey interface{} `json:"private_key"`
	KeyID      string     `json:"key_id"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Algorithm  string     `json:"algorithm"`
}

// TokenClaims JWT声明
type TokenClaims struct {
	// 标准声明
	jwt.RegisteredClaims

	// 自定义声明
	UserID      uint                   `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	TenantID    string                 `json:"tenant_id"`
	DeviceID    string                 `json:"device_id"`
	SessionID   string                 `json:"session_id"`

	// 权限相关
	ResourceAccess map[string]interface{} `json:"resource_access,omitempty"`
	Constraints    map[string]interface{} `json:"constraints,omitempty"`

	// 安全相关
	Nonce        string                 `json:"nonce"`
	Fingerprint  string                 `json:"fingerprint"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`

	// 元数据
	Metadata     map[string]interface{} `json:"metadata"`
}

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	Scope        string    `json:"scope"`
}

// TokenValidationResult 令牌验证结果
type TokenValidationResult struct {
	Valid      bool                   `json:"valid"`
	Claims     *TokenClaims            `json:"claims,omitempty"`
	Errors     []string               `json:"errors"`
	Warnings   []string               `json:"warnings"`
	Metadata   map[string]interface{} `json:"metadata"`
	ValidationDetails map[string]interface{} `json:"validation_details"`
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(config *JWTConfig, tokenRepo repositories.TokenRepository, userRepo repositories.UserRepository, auditRepo repositories.AuditRepository, logger *logrus.Logger) (*JWTManager, error) {
	manager := &JWTManager{
		config:    config,
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		logger:    logger,
		blacklist: NewTokenBlacklist(config, logger),
		validator: NewTokenValidator(config, logger),
	}

	// 初始化密钥
	if err := manager.initializeKeys(); err != nil {
		return nil, fmt.Errorf("failed to initialize keys: %w", err)
	}

	// 启动密钥轮换
	go manager.startKeyRotation()

	return manager, nil
}

// initializeKeys 初始化密钥
func (m *JWTManager) initializeKeys() error {
	// 初始化访问令牌密钥
	accessKeyPair, err := m.generateKeyPair(config.AccessTokenAlg, config.KeySizeBits)
	if err != nil {
		return fmt.Errorf("failed to generate access key pair: %w", err)
	}
	m.accessKeyPair = accessKeyPair

	// 初始化刷新令牌密钥
	refreshKeyPair, err := m.generateKeyPair(config.RefreshTokenAlg, config.KeySizeBits)
	if err != nil {
		return fmt.Errorf("failed to generate refresh key pair: %w", err)
	}
	m.refreshKeyPair = refreshKeyPair

	m.logger.WithFields(logrus.Fields{
		"access_algorithm": config.AccessTokenAlg,
		"refresh_algorithm": config.RefreshTokenAlg,
		"key_size": config.KeySizeBits,
	}).Info("JWT keys initialized successfully")

	return nil
}

// generateKeyPair 生成密钥对
func (m *JWTManager) generateKeyPair(algorithm string, keySizeBits int) (*KeyPair, error) {
	keyID := fmt.Sprintf("key_%d", time.Now().Unix())

	var publicKey, privateKey interface{}
	var alg string

	switch strings.ToUpper(algorithm) {
	case "RS256":
		if keySizeBits < 2048 {
			keySizeBits = 2048 // RSA最小密钥长度
		}
		privateKey, err := rsa.GenerateKey(rand.Reader, keySizeBits)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}
		publicKey = &privateKey.PublicKey
		alg = "RS256"

	case "RS384":
		if keySizeBits < 3072 {
			keySizeBits = 3072
		}
		privateKey, err := rsa.GenerateKey(rand.Reader, keySizeBits)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}
		publicKey = &privateKey.PublicKey
		alg = "RS384"

	case "RS512":
		if keySizeBits < 4096 {
			keySizeBits = 4096
		}
		privateKey, err := rsa.GenerateKey(rand.Reader, keySizeBits)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
		}
		publicKey = &privateKey.PublicKey
		alg = "RS512"

	case "ES256":
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA key pair: %w", err)
		}
		publicKey = &privateKey.PublicKey
		alg = "ES256"

	case "ES384":
		privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA key pair: %w", err)
		}
		publicKey = &privateKey.PublicKey
		alg = "ES384"

	case "ES512":
		privateKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA key pair: %w", err)
		}
		publicKey = &privateKey.PublicKey
		alg = "ES512"

	case "EdDSA":
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate EdDSA key pair: %w", err)
		}
		alg = "EdDSA"

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	return &KeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		KeyID:      keyID,
		CreatedAt:  time.Now(),
		Algorithm:  alg,
	}, nil
}

// GenerateTokenPair 生成令牌对
func (m *JWTManager) GenerateTokenPair(user *models.User, deviceInfo *DeviceInfo, sessionInfo *SessionInfo) (*TokenPair, error) {
	m.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"username": user.Username,
		"tenant_id": user.TenantID,
	}).Debug("Generating token pair")

	// 生成访问令牌
	accessToken, err := m.generateAccessToken(user, deviceInfo, sessionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := m.generateRefreshToken(user, deviceInfo, sessionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 计算过期时间
	expiresIn := int64(m.config.AccessTokenTTL.Seconds())

	tokenPair := &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		Scope:        m.generateTokenScope(user),
	}

	// 记录令牌到数据库
	if err := m.recordToken(user.ID, accessToken, refreshToken, deviceInfo, sessionInfo); err != nil {
		m.logger.WithError(err).Error("Failed to record tokens to database")
	}

	// 记录审计日志
	if m.config.AuditEnabled {
		m.logTokenOperation("token_generated", user.ID, user.TenantID, map[string]interface{}{
			"device_id":    deviceInfo.ID,
			"session_id":   sessionInfo.ID,
			"expires_in":   expiresIn,
			"scope":        tokenPair.Scope,
		})
	}

	m.logger.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"expires_in": expiresIn,
	}).Info("Token pair generated successfully")

	return tokenPair, nil
}

// generateAccessToken 生成访问令牌
func (m *JWTManager) generateAccessToken(user *models.User, deviceInfo *DeviceInfo, sessionInfo *SessionInfo) (string, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.AccessTokenTTL)

	// 创建声明
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			Audience:  m.config.Audience,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now.Add(-m.config.Leeway)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        m.generateTokenID(),
		},
		UserID:    user.ID,
	Username:  user.Username,
		Email:     user.Email,
		TenantID:  user.TenantID,
		DeviceID:  deviceInfo.ID,
		SessionID: sessionInfo.ID,
		IPAddress: deviceInfo.IPAddress,
		UserAgent: deviceInfo.UserAgent,
		Metadata: map[string]interface{}{
			"token_type": "access",
			"version": "1.0",
		},
	}

	// 添加用户角色和权限
	if roles, err := m.getUserRoles(user.ID); err == nil {
		claims.Roles = roles
	}

	if permissions, err := m.getUserPermissions(user.ID); err == nil {
		claims.Permissions = permissions
	}

	// 添加自定义声明
	if deviceInfo.Fingerprint != "" {
		claims.Fingerprint = deviceInfo.Fingerprint
	}

	if sessionInfo.Nonce != "" {
		claims.Nonce = sessionInfo.Nonce
	}

	// 生成令牌
	token := jwt.NewWithClaims(m.config.AccessTokenSignMethod, claims)

	tokenString, err := token.SignedString(m.accessKeyPair.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// 添加密钥ID到header
	token.Header["kid"] = m.accessKeyPair.KeyID

	return token.SignedString(m.accessKeyPair.PrivateKey)
}

// generateRefreshToken 生成刷新令牌
func (m *JWTManager) generateRefreshToken(user *models.User, deviceInfo *DeviceInfo, sessionInfo *SessionInfo) (string, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.RefreshTokenTTL)

	// 刷新令牌只包含基本声明
	claims := &jwt.RegisteredClaims{
		Issuer:    m.config.Issuer,
		Subject:   strconv.FormatUint(uint64(user.ID), 10),
		Audience:  m.config.Audience + ":refresh",
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(now.Add(-m.config.Leeway)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        m.generateTokenID(),
	}

	// 添加自定义声明
	customClaims := map[string]interface{}{
		"user_id":     user.ID,
	"tenant_id":   user.TenantID,
	"device_id":   deviceInfo.ID,
		"session_id":  sessionInfo.ID,
		"ip_address":  deviceInfo.IPAddress,
	"user_agent":   deviceInfo.UserAgent,
		"token_type":  "refresh",
		"version":     "1.0",
	}

	if deviceInfo.Fingerprint != "" {
		customClaims["fingerprint"] = deviceInfo.Fingerprint
	}

	// 生成令牌
	token := jwt.NewWithClaims(m.config.RefreshTokenSignMethod, claims)

	// 合并自定义声明
	token.Claims = customClaims

	tokenString, err := token.SignedString(m.refreshKeyPair.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// 添加密钥ID到header
	token.Header["kid"] = m.refreshKeyPair.KeyID

	return token.SignedString(m.refreshKeyPair.PrivateKey)
}

// ValidateToken 验证令牌
func (m *JWTManager) ValidateToken(tokenString string) (*TokenValidationResult, error) {
	startTime := time.Now()

	m.logger.WithFields(logrus.Fields{
		"token_length": len(tokenString),
	}).Debug("Validating JWT token")

	result := &TokenValidationResult{
		Valid:            false,
		Errors:           []string{},
		Warnings:         []string{},
		Metadata:         make(map[string]interface{}),
		ValidationDetails: make(map[string]interface{}),
	}

	// 检查令牌黑名单
	if m.config.BlacklistEnabled {
		if m.blacklist.IsBlacklisted(tokenString) {
			result.Errors = append(result.Errors, "token is blacklisted")
			result.Metadata["blacklisted"] = true
			return result, nil
		}
	}

	// 解析令牌
	token, claims, err := m.parseToken(tokenString)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to parse token: %v", err))
		return result, nil
	}

	result.Claims = claims

	// 验证声明
	if validationErrors := m.validator.ValidateClaims(claims); len(validationErrors) > 0 {
		result.Errors = append(result.Errors, validationErrors...)
		return result, nil
	}

	// 验证用户状态
	if err := m.validateUserStatus(claims.UserID); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("user validation failed: %v", err))
		return result, nil
	}

	// 验证设备状态
	if claims.DeviceID != "" {
		if err := m.validateDeviceStatus(claims.DeviceID, claims.IPAddress, claims.Fingerprint); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("device validation warning: %v", err))
		}
	}

	// 验证会话状态
	if claims.SessionID != "" {
		if err := m.validateSessionStatus(claims.SessionID, claims.UserID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("session validation failed: %v", err))
			return result, nil
		}
	}

	result.Valid = true
	result.Metadata["validation_time"] = time.Since(startTime)
	result.Metadata["token_id"] = claims.ID
	result.Metadata["key_id"] = token.Header["kid"]

	// 记录验证结果
	m.logTokenValidation("token_validated", claims.UserID, claims.TenantID, result)

	m.logger.WithFields(logrus.Fields{
		"user_id":    claims.UserID,
		"token_id":   claims.ID,
		"valid":      result.Valid,
		"duration":   time.Since(startTime).Milliseconds(),
		"errors":     len(result.Errors),
		"warnings":   len(result.Warnings),
	}).Debug("Token validation completed")

	return result, nil
}

// RefreshToken 刷新令牌
func (m *JWTManager) RefreshToken(refreshTokenString string, deviceInfo *DeviceInfo, sessionInfo *SessionInfo) (*TokenPair, error) {
	m.logger.Debug("Refreshing JWT token")

	// 验证刷新令牌
	result := m.validateRefreshToken(refreshTokenString)
	if !result.Valid {
		return nil, fmt.Errorf("invalid refresh token: %v", result.Errors)
	}

	claims := result.Claims

	// 获取用户信息
	user, err := m.userRepo.GetByID(context.Background(), claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if !user.Active {
		return nil, fmt.Errorf("user is inactive")
	}

	// 生成新的令牌对
	newTokenPair, err := m.GenerateTokenPair(user, deviceInfo, sessionInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	// 将旧的刷新令牌加入黑名单
	if m.config.BlacklistEnabled {
		m.blacklist.AddToBlacklist(refreshTokenString)
	}

	// 记录审计日志
	if m.config.AuditEnabled {
		m.logTokenOperation("token_refreshed", user.ID, user.TenantID, map[string]interface{}{
			"old_token_id": claims.ID,
			"new_token_id": m.extractTokenID(newTokenPair.AccessToken),
			"device_id":    deviceInfo.ID,
			"session_id":   sessionInfo.ID,
		})
	}

	m.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"old_token_id": claims.ID,
		"new_token_id": m.extractTokenID(newTokenPair.AccessToken),
	}).Info("Token refreshed successfully")

	return newTokenPair, nil
}

// RevokeToken 撤销令牌
func (m *JWTManager) RevokeToken(tokenString string, reason string) error {
	m.logger.WithField("reason", reason).Debug("Revoking JWT token")

	// 验证令牌
	result, err := m.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// 添加到黑名单
	if m.config.BlacklistEnabled {
		m.blacklist.AddToBlacklist(tokenString)
	}

	// 更新数据库状态
	if result.Claims != nil {
		if err := m.tokenRepo.RevokeToken(context.Background(), result.Claims.ID, reason); err != nil {
			m.logger.WithError(err).Error("Failed to revoke token in database")
		}
	}

	// 记录审计日志
	if m.config.AuditEnabled && result.Claims != nil {
		m.logTokenOperation("token_revoked", result.Claims.UserID, result.Claims.TenantID, map[string]interface{}{
			"token_id": result.Claims.ID,
			"reason":   reason,
		})
	}

	m.logger.Info("Token revoked successfully")
	return nil
}

// 辅助方法
func (m *JWTManager) generateTokenID() string {
	return fmt.Sprintf("jwt_%d_%s", time.Now().UnixNano(), generateRandomString(16))
}

func (m *JWTManager) generateTokenScope(user *models.User) string {
	scope := "read write"
	if user.IsAdmin {
		scope += " admin"
	}
	return scope
}

func (m *JWTManager) getUserRoles(userID uint) ([]string, error) {
	roles, err := m.userRepo.GetRoles(context.Background(), userID, "")
	if err != nil {
		return nil, err
	}

	var roleNames []string
	for _, userRole := range roles {
		roleNames = append(roleNames, userRole.Role.Name)
	}

	return roleNames, nil
}

func (m *JWTManager) getUserPermissions(userID uint) ([]string, error) {
	// 简化实现，实际应该从权限服务获取
	return []string{}, nil
}

func (m *JWTManager) recordToken(userID uint, accessToken, refreshToken string, deviceInfo *DeviceInfo, sessionInfo *SessionInfo) error {
	// 实现令牌记录逻辑
	return nil
}

func (m *JWTManager) parseToken(tokenString string) (*jwt.Token, *TokenClaims, error) {
	// 首先尝试解析为访问令牌
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if token.Header["alg"] == nil {
			return nil, fmt.Errorf("missing algorithm in header")
		}

		alg := token.Header["alg"].(string)

		if alg == m.accessKeyPair.Algorithm {
			return m.accessKeyPair.PublicKey, nil
		}

		return nil, fmt.Errorf("invalid algorithm: %s", alg)
	})

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(*TokenClaims); ok {
			return token, claims, nil
		}
	}

	// 如果不是访问令牌，尝试解析为刷新令牌
	refreshToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Header["alg"] == nil {
			return nil, fmt.Errorf("missing algorithm in header")
		}

		alg := token.Header["alg"].(string)

		if alg == m.refreshKeyPair.Algorithm {
			return m.refreshKeyPair.PublicKey, nil
		}

		return nil, fmt.Errorf("invalid algorithm: %s", alg)
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 提取自定义声明
	if claimsMap, ok := refreshToken.Claims.(map[string]interface{}); ok {
		claims := &TokenClaims{}
		claimsBytes, _ := json.Marshal(claimsMap)
		json.Unmarshal(claimsBytes, claims)

		return refreshToken, claims, nil
	}

	return nil, nil, fmt.Errorf("invalid token claims")
}

func (m *JWTManager) validateUserStatus(userID uint) error {
	user, err := m.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if !user.Active {
		return fmt.Errorf("user is inactive")
	}

	return nil
}

func (m *JWTManager) validateDeviceStatus(deviceID, ipAddress, fingerprint string) error {
	// 实现设备状态验证
	return nil
}

func (m *JWTManager) validateSessionStatus(sessionID string, userID uint) error {
	// 实现会话状态验证
	return nil
}

func (m *JWTManager) validateRefreshToken(tokenString string) *TokenValidationResult {
	// 实现刷新令牌的特殊验证逻辑
	return &TokenValidationResult{}
}

func (m *JWTManager) extractTokenID(tokenString string) string {
	token, _, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})

	if token != nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if jti, exists := claims["jti"]; exists {
				return fmt.Sprintf("%v", jti)
			}
		}
	}

	return ""
}

func (m *JWTManager) startKeyRotation() {
	if m.config.KeyRotationInterval <= 0 {
		return
	}

	ticker := time.NewTicker(m.config.KeyRotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.logger.Info("Starting key rotation")
			if err := m.rotateKeys(); err != nil {
				m.logger.WithError(err).Error("Failed to rotate keys")
			}
		}
	}
}

func (m *JWTManager) rotateKeys() error {
	// 生成新的访问令牌密钥
	newAccessKeyPair, err := m.generateKeyPair(m.config.AccessTokenAlg, m.config.KeySizeBits)
	if err != nil {
		return fmt.Errorf("failed to generate new access key pair: %w", err)
	}

	// 生成新的刷新令牌密钥
	newRefreshKeyPair, err := m.generateKeyPair(m.config.RefreshTokenAlg, m.config.KeySizeBits)
	if err != nil {
		return fmt.Errorf("failed to generate new refresh key pair: w", err)
	}

	// 更新密钥
	m.accessKeyPair = newAccessKeyPair
	m.refreshKeyPair = newRefreshKeyPair

	m.logger.Info("Keys rotated successfully")
	return nil
}

func (m *JWTManager) logTokenOperation(operation string, userID uint, tenantID string, metadata map[string]interface{}) {
	auditLog := &models.AuditLog{
		UserID:    userID,
		Action:    operation,
		Resource:  "jwt",
		Result:    "success",
		Details:   fmt.Sprintf("%+v", metadata),
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}

	if err := m.auditRepo.Create(context.Background(), auditLog); err != nil {
		m.logger.WithError(err).Error("Failed to log JWT operation")
	}
}

func (m *JWTManager) logTokenValidation(operation string, claims *TokenClaims, tenantID string, result *TokenValidationResult) {
	auditLog := &models.AuditLog{
		UserID:    claims.UserID,
		Action:    operation,
		Resource:  fmt.Sprintf("jwt:%s", claims.ID),
		Result:    "success",
		Details:   fmt.Sprintf("valid: %t, errors: %v", result.Valid, result.Errors),
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}

	if err := m.auditRepo.Create(context.Background(), auditLog); err != nil {
		m.logger.WithError(err).Error("Failed to log JWT validation")
	}
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Platform    string    `json:"platform"`
	OS          string    `json:"os"`
	IPAddress   string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Fingerprint string    `json:"fingerprint"`
	Trusted     bool      `json:"trusted"`
	CreatedAt   time.Time `json:"created_at"`
	LastSeen    time.Time `json:"last_seen"`
}

// SessionInfo 会话信息
type SessionInfo struct {
	ID        string    `json:"id"`
	UserID    uint      `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Nonce     string    `json:"nonce"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}