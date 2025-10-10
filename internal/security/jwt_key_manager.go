package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
)

var (
	jwtAuthDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "jwt_auth_duration_seconds",
		Help:    "Duration of JWT authentication operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})

	jwtAuthErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jwt_auth_errors_total",
		Help: "Total number of JWT authentication errors",
	}, []string{"operation"})

	jwtTokensIssued = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jwt_tokens_issued_total",
		Help: "Total number of JWT tokens issued",
	}, []string{"type"})

	jwtTokensValidated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jwt_tokens_validated_total",
		Help: "Total number of JWT tokens validated",
	})

	jwtKeyRotations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "jwt_key_rotations_total",
		Help: "Total number of JWT key rotations",
	})
)

// JWTKeyManager JWT密钥管理器
type JWTKeyManager struct {
	currentSecret  []byte
	previousSecret []byte
	rotationTime   time.Time
	rotationPeriod time.Duration
	config         *config.Config
	securityConfig *SecurityConfig
	redisClient    *redis.Client
	cacheService   *cache.CacheService
	mu             sync.RWMutex
	keyHistory     map[string][]byte // 存储历史密钥用于验证旧令牌
}

// TokenDetails 令牌详情
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// TokenPayload 令牌负载
type TokenPayload struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	DeviceID   string `json:"device_id"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	TokenType  string `json:"token_type"`
	Authorized bool   `json:"authorized"`
	KeyVersion string `json:"key_version"` // 密钥版本
}

// NewJWTKeyManager 创建新的JWT密钥管理器
func NewJWTKeyManager(cfg *config.Config, securityConfig *SecurityConfig, redisClient *redis.Client, cacheService *cache.CacheService) *JWTKeyManager {
	// 从安全配置获取JWT密钥，如果没有则使用配置文件中的密钥
	secret := securityConfig.JWT.Secret
	if secret == "" {
		secret = cfg.JWT.Secret
	}

	if len(secret) < 32 {
		secret = generateSecureRandomKey(32)
	}

	manager := &JWTKeyManager{
		currentSecret:  []byte(secret),
		rotationTime:   time.Now(),
		rotationPeriod: securityConfig.JWT.BlacklistTTL,
		config:         cfg,
		securityConfig: securityConfig,
		redisClient:    redisClient,
		cacheService:   cacheService,
		keyHistory:     make(map[string][]byte),
	}

	// 初始化密钥历史
	manager.keyHistory["v1"] = manager.currentSecret

	return manager
}

// CreateTokens 创建访问令牌和刷新令牌
func (jkm *JWTKeyManager) CreateTokens(ctx context.Context, user *models.User, deviceID, ip, userAgent string) (*TokenDetails, error) {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("create").Observe(time.Since(start).Seconds())
	}()

	jkm.mu.RLock()
	defer jkm.mu.RUnlock()

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(jkm.securityConfig.JWT.AccessTokenTTL).Unix()
	td.RtExpires = time.Now().Add(jkm.securityConfig.JWT.RefreshTokenTTL).Unix()

	// 生成UUID
	td.AccessUUID = generateUUID()
	td.RefreshUUID = generateUUID()

	// 获取当前密钥版本
	keyVersion := jkm.getCurrentKeyVersion()

	// 创建访问令牌
	accessToken, err := jkm.createToken(user, td.AccessUUID, td.AtExpires, "access", deviceID, ip, userAgent, keyVersion)
	if err != nil {
		jwtAuthErrors.WithLabelValues("create_access").Inc()
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// 创建刷新令牌
	refreshToken, err := jkm.createToken(user, td.RefreshUUID, td.RtExpires, "refresh", deviceID, ip, userAgent, keyVersion)
	if err != nil {
		jwtAuthErrors.WithLabelValues("create_refresh").Inc()
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	td.AccessToken = accessToken
	td.RefreshToken = refreshToken

	// 存储令牌到Redis
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	accessKey := fmt.Sprintf("access_token:%s:%s", td.AccessUUID, keyVersion)
	refreshKey := fmt.Sprintf("refresh_token:%s:%s", td.RefreshUUID, keyVersion)

	// 存储访问令牌
	err = jkm.redisClient.Set(ctx, accessKey, user.ID, at.Sub(now)).Err()
	if err != nil {
		jwtAuthErrors.WithLabelValues("store_access").Inc()
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	// 存储刷新令牌
	err = jkm.redisClient.Set(ctx, refreshKey, user.ID, rt.Sub(now)).Err()
	if err != nil {
		jwtAuthErrors.WithLabelValues("store_refresh").Inc()
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// 存储用户设备信息
	deviceKey := fmt.Sprintf("user_device:%d:%s", user.ID, deviceID)
	deviceInfo := map[string]interface{}{
		"user_id":      user.ID,
		"device_id":    deviceID,
		"ip":           ip,
		"user_agent":   userAgent,
		"access_uuid":  td.AccessUUID,
		"refresh_uuid": td.RefreshUUID,
		"key_version":  keyVersion,
		"created_at":   now,
		"last_active":  now,
	}

	err = jkm.cacheService.Set(deviceKey, deviceInfo, jkm.securityConfig.JWT.RefreshTokenTTL)
	if err != nil {
		// 不影响主要功能，只记录警告
		fmt.Printf("Warning: failed to store device info: %v\n", err)
	}

	jwtTokensIssued.WithLabelValues("access").Inc()
	jwtTokensIssued.WithLabelValues("refresh").Inc()

	return td, nil
}

// createToken 创建JWT令牌
func (jkm *JWTKeyManager) createToken(user *models.User, uuid string, expires int64, tokenType, deviceID, ip, userAgent, keyVersion string) (string, error) {
	payload := &TokenPayload{
		UserID:     user.ID,
		Username:   user.Name, // User模型中使用Name字段，不是Username
		Email:      user.Email,
		Role:       user.Role,
		DeviceID:   deviceID,
		IP:         ip,
		UserAgent:  userAgent,
		TokenType:  tokenType,
		Authorized: true,
		KeyVersion: keyVersion,
	}

	claims := jwt.MapClaims{
		"payload": payload,
		"uuid":    uuid,
		"issuer":  jkm.securityConfig.JWT.Issuer,
		"exp":     expires,
		"iat":     time.Now().Unix(),
		"type":    tokenType,
		"version": keyVersion,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jkm.currentSecret)
}

// VerifyToken 验证JWT令牌
func (jkm *JWTKeyManager) VerifyToken(ctx context.Context, tokenString string) (*jwt.MapClaims, error) {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("verify").Observe(time.Since(start).Seconds())
		jwtTokensValidated.Inc()
	}()

	// 首先尝试用当前密钥验证
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jkm.currentSecret, nil
	})

	if err == nil && token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			return &claims, nil
		}
	}

	// 如果当前密钥验证失败，尝试用历史密钥验证
	jkm.mu.RLock()
	defer jkm.mu.RUnlock()

	for version, secret := range jkm.keyHistory {
		if string(secret) == string(jkm.currentSecret) {
			continue // 跳过当前密钥，已经尝试过了
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				// 验证令牌中的版本是否匹配
				if tokenVersion, ok := claims["version"].(string); ok && tokenVersion == version {
					return &claims, nil
				}
			}
		}
	}

	jwtAuthErrors.WithLabelValues("verify").Inc()
	return nil, fmt.Errorf("invalid token")
}

// ExtractTokenMetadata 从令牌中提取元数据
func (jkm *JWTKeyManager) ExtractTokenMetadata(ctx context.Context, tokenString string) (*TokenPayload, error) {
	claims, err := jkm.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	payloadMap, ok := (*claims)["payload"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid payload format")
	}

	userID, ok := payloadMap["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user_id format")
	}

	username, ok := payloadMap["username"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid username format")
	}

	email, ok := payloadMap["email"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid email format")
	}

	role, ok := payloadMap["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role format")
	}

	deviceID, ok := payloadMap["device_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid device_id format")
	}

	ip, ok := payloadMap["ip"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid ip format")
	}

	userAgent, ok := payloadMap["user_agent"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_agent format")
	}

	tokenType, ok := payloadMap["token_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token_type format")
	}

	authorized, ok := payloadMap["authorized"].(bool)
	if !ok {
		return nil, fmt.Errorf("invalid authorized format")
	}

	keyVersion, ok := payloadMap["key_version"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid key_version format")
	}

	return &TokenPayload{
		UserID:     uint(userID),
		Username:   username,
		Email:      email,
		Role:       role,
		DeviceID:   deviceID,
		IP:         ip,
		UserAgent:  userAgent,
		TokenType:  tokenType,
		Authorized: authorized,
		KeyVersion: keyVersion,
	}, nil
}

// RotateKeys 旋转密钥
func (jkm *JWTKeyManager) RotateKeys(ctx context.Context) error {
	jkm.mu.Lock()
	defer jkm.mu.Unlock()

	// 检查是否需要旋转
	if time.Since(jkm.rotationTime) < jkm.rotationPeriod {
		return nil
	}

	// 保存当前密钥到历史
	oldVersion := jkm.getCurrentKeyVersion()
	jkm.keyHistory[oldVersion] = jkm.currentSecret

	// 生成新密钥
	newSecret := []byte(generateSecureRandomKey(32))
	newVersion := fmt.Sprintf("v%d", len(jkm.keyHistory)+1)

	// 更新当前密钥
	jkm.previousSecret = jkm.currentSecret
	jkm.currentSecret = newSecret
	jkm.rotationTime = time.Now()

	// 添加新版本到历史
	jkm.keyHistory[newVersion] = jkm.currentSecret

	// 更新安全配置
	jkm.securityConfig.JWT.Secret = string(newSecret)

	// 清理过期的历史密钥（保留最近3个）
	if len(jkm.keyHistory) > 3 {
		for version := range jkm.keyHistory {
			if version != newVersion && version != oldVersion && len(jkm.keyHistory) > 3 {
				delete(jkm.keyHistory, version)
			}
		}
	}

	jwtKeyRotations.Inc()

	// 记录密钥旋转事件
	fmt.Printf("JWT key rotated from %s to %s\n", oldVersion, newVersion)

	return nil
}

// getCurrentKeyVersion 获取当前密钥版本
func (jkm *JWTKeyManager) getCurrentKeyVersion() string {
	return fmt.Sprintf("v%d", len(jkm.keyHistory))
}

// generateSecureRandomKey 生成安全的随机密钥
func generateSecureRandomKey(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("failed to generate secure random key: %w", err))
	}
	return hex.EncodeToString(b)
}

// generateUUID 生成UUID
func generateUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("failed to generate UUID: %w", err))
	}
	return hex.EncodeToString(b)
}

// ValidateAccess 验证访问权限
func (jkm *JWTKeyManager) ValidateAccess(ctx context.Context, tokenString string, requiredRoles []string) (*TokenPayload, error) {
	payload, err := jkm.ExtractTokenMetadata(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	// 检查令牌类型
	if payload.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type for access")
	}

	// 检查授权状态
	if !payload.Authorized {
		return nil, fmt.Errorf("token not authorized")
	}

	// 检查角色权限
	if len(requiredRoles) > 0 {
		hasRole := false
		for _, role := range requiredRoles {
			if payload.Role == role {
				hasRole = true
				break
			}
		}
		if !hasRole {
			return nil, fmt.Errorf("insufficient permissions")
		}
	}

	// 验证Redis中的令牌是否存在
	accessKey := fmt.Sprintf("access_token:%s:%s", payload.TokenID(), payload.KeyVersion)
	_, err = jkm.redisClient.Get(ctx, accessKey).Result()
	if err != nil {
		return nil, fmt.Errorf("token not found or expired")
	}

	return payload, nil
}

// TokenID 从payload中提取token ID
func (tp *TokenPayload) TokenID() string {
	// 这个方法需要在创建令牌时存储UUID信息
	// 暂时返回空字符串，实际使用时需要修改
	return ""
}

// RefreshTokens 刷新令牌
func (jkm *JWTKeyManager) RefreshTokens(ctx context.Context, refreshToken string) (*TokenDetails, error) {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("refresh").Observe(time.Since(start).Seconds())
	}()

	// 验证刷新令牌
	claims, err := jkm.VerifyToken(ctx, refreshToken)
	if err != nil {
		jwtAuthErrors.WithLabelValues("refresh_verify").Inc()
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 检查令牌类型
	tokenType, ok := (*claims)["type"].(string)
	if !ok || tokenType != "refresh" {
		jwtAuthErrors.WithLabelValues("refresh_type").Inc()
		return nil, fmt.Errorf("invalid token type")
	}

	// 获取UUID和版本
	uuid, ok := (*claims)["uuid"].(string)
	if !ok {
		jwtAuthErrors.WithLabelValues("refresh_uuid").Inc()
		return nil, fmt.Errorf("invalid token uuid")
	}

	version, ok := (*claims)["version"].(string)
	if !ok {
		jwtAuthErrors.WithLabelValues("refresh_version").Inc()
		return nil, fmt.Errorf("invalid token version")
	}

	// 从Redis获取用户ID
	refreshKey := fmt.Sprintf("refresh_token:%s:%s", uuid, version)
	userIDStr, err := jkm.redisClient.Get(ctx, refreshKey).Result()
	if err != nil {
		jwtAuthErrors.WithLabelValues("refresh_not_found").Inc()
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	// 删除旧的刷新令牌
	jkm.redisClient.Del(ctx, refreshKey)

	// 获取用户信息
	var user models.User
	// 这里需要从数据库获取用户信息，暂时简化处理
	// 实际项目中应该注入数据库连接
	// 解析用户ID字符串为uint
	userIDUint, _ := strconv.ParseUint(userIDStr, 10, 32)
	user.ID = uint(userIDUint)

	// 创建新的令牌对
	return jkm.CreateTokens(ctx, &user, "", "", "")
}

// RevokeToken 撤销令牌
func (jkm *JWTKeyManager) RevokeToken(ctx context.Context, tokenString string) error {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("revoke").Observe(time.Since(start).Seconds())
	}()

	claims, err := jkm.VerifyToken(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	uuid, ok := (*claims)["uuid"].(string)
	if !ok {
		return fmt.Errorf("invalid token uuid")
	}

	version, ok := (*claims)["version"].(string)
	if !ok {
		return fmt.Errorf("invalid token version")
	}

	tokenType, ok := (*claims)["type"].(string)
	if !ok {
		return fmt.Errorf("invalid token type")
	}

	var key string
	switch tokenType {
	case "access":
		key = fmt.Sprintf("access_token:%s:%s", uuid, version)
	case "refresh":
		key = fmt.Sprintf("refresh_token:%s:%s", uuid, version)
	default:
		return fmt.Errorf("unknown token type")
	}

	// 从Redis删除令牌
	err = jkm.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens 撤销用户所有令牌
func (jkm *JWTKeyManager) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("revoke_all").Observe(time.Since(start).Seconds())
	}()

	pattern := fmt.Sprintf("*token:*:%d", userID)
	keys, err := jkm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find user tokens: %w", err)
	}

	if len(keys) > 0 {
		err = jkm.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to revoke user tokens: %w", err)
		}
	}

	// 撤销设备信息
	devicePattern := fmt.Sprintf("user_device:%d:*", userID)
	deviceKeys, err := jkm.redisClient.Keys(ctx, devicePattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find user devices: %w", err)
	}

	if len(deviceKeys) > 0 {
		err = jkm.redisClient.Del(ctx, deviceKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to revoke user devices: %w", err)
		}
	}

	return nil
}

// IsTokenBlacklisted 检查令牌是否在黑名单中
func (jkm *JWTKeyManager) IsTokenBlacklisted(ctx context.Context, tokenString string) bool {
	payload, err := jkm.ExtractTokenMetadata(ctx, tokenString)
	if err != nil {
		return true
	}

	blacklistKey := fmt.Sprintf("blacklist:%s:%s", payload.TokenID(), payload.KeyVersion)
	_, err = jkm.redisClient.Get(ctx, blacklistKey).Result()
	return err == nil
}

// BlacklistToken 将令牌加入黑名单
func (jkm *JWTKeyManager) BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	payload, err := jkm.ExtractTokenMetadata(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("failed to extract token metadata: %w", err)
	}

	blacklistKey := fmt.Sprintf("blacklist:%s:%s", payload.TokenID(), payload.KeyVersion)
	err = jkm.redisClient.Set(ctx, blacklistKey, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// GetKeyStats 获取密钥统计信息
func (jkm *JWTKeyManager) GetKeyStats() map[string]interface{} {
	jkm.mu.RLock()
	defer jkm.mu.RUnlock()

	return map[string]interface{}{
		"current_key_version": jkm.getCurrentKeyVersion(),
		"key_history_size":    len(jkm.keyHistory),
		"last_rotation":       jkm.rotationTime,
		"rotation_period":     jkm.rotationPeriod,
		"next_rotation":       jkm.rotationTime.Add(jkm.rotationPeriod),
	}
}
