package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
)

// TokenManager JWT令牌管理器
type TokenManager struct {
	secret       []byte
	accessTTL    time.Duration
	refreshTTL   time.Duration
	redisClient  *redis.Client
	cacheService *cache.CacheService
	issuer       string
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
}

// NewTokenManager 创建新的令牌管理器
func NewTokenManager(cfg *config.Config, redisClient *redis.Client, cacheService *cache.CacheService) *TokenManager {
	return &TokenManager{
		secret:       []byte(cfg.JWT.Secret),
		accessTTL:    time.Duration(cfg.JWT.ExpiresIn) * time.Second,
		refreshTTL:   time.Duration(cfg.JWT.RefreshIn) * time.Second,
		redisClient:  redisClient,
		cacheService: cacheService,
		issuer:       "law-oa-system",
	}
}

// CreateTokens 创建访问令牌和刷新令牌
func (tm *TokenManager) CreateTokens(ctx context.Context, user *models.User, deviceID, ip, userAgent string) (*TokenDetails, error) {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("create").Observe(time.Since(start).Seconds())
	}()

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(tm.accessTTL).Unix()
	td.RtExpires = time.Now().Add(tm.refreshTTL).Unix()

	// 生成UUID
	td.AccessUUID = generateUUID()
	td.RefreshUUID = generateUUID()

	// 创建访问令牌
	accessToken, err := tm.createToken(user, td.AccessUUID, td.AtExpires, "access", deviceID, ip, userAgent)
	if err != nil {
		jwtAuthErrors.WithLabelValues("create_access").Inc()
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// 创建刷新令牌
	refreshToken, err := tm.createToken(user, td.RefreshUUID, td.RtExpires, "refresh", deviceID, ip, userAgent)
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

	accessKey := fmt.Sprintf("access_token:%s", td.AccessUUID)
	refreshKey := fmt.Sprintf("refresh_token:%s", td.RefreshUUID)

	// 存储访问令牌
	err = tm.redisClient.Set(ctx, accessKey, user.ID, at.Sub(now)).Err()
	if err != nil {
		jwtAuthErrors.WithLabelValues("store_access").Inc()
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	// 存储刷新令牌
	err = tm.redisClient.Set(ctx, refreshKey, user.ID, rt.Sub(now)).Err()
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
		"created_at":   now,
		"last_active":  now,
	}

	err = tm.cacheService.Set(deviceKey, deviceInfo, tm.refreshTTL)
	if err != nil {
		// 不影响主要功能，只记录警告
		fmt.Printf("Warning: failed to store device info: %v\n", err)
	}

	jwtTokensIssued.WithLabelValues("access").Inc()
	jwtTokensIssued.WithLabelValues("refresh").Inc()

	return td, nil
}

// createToken 创建JWT令牌
func (tm *TokenManager) createToken(user *models.User, uuid string, expires int64, tokenType, deviceID, ip, userAgent string) (string, error) {
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
	}

	claims := jwt.MapClaims{
		"payload": payload,
		"uuid":    uuid,
		"issuer":  tm.issuer,
		"exp":     expires,
		"iat":     time.Now().Unix(),
		"type":    tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secret)
}

// VerifyToken 验证JWT令牌
func (tm *TokenManager) VerifyToken(ctx context.Context, tokenString string) (*jwt.MapClaims, error) {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("verify").Observe(time.Since(start).Seconds())
		jwtTokensValidated.Inc()
	}()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil {
		jwtAuthErrors.WithLabelValues("verify").Inc()
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	jwtAuthErrors.WithLabelValues("verify_invalid").Inc()
	return nil, fmt.Errorf("invalid token")
}

// VerifyTokenInterface 验证JWT令牌（接口兼容版本）
func (tm *TokenManager) VerifyTokenInterface(ctx context.Context, tokenString string) (*map[string]interface{}, error) {
	claims, err := tm.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	// 转换 *jwt.MapClaims 到 *map[string]interface{}
	result := map[string]interface{}(*claims)
	return &result, nil
}

// ExtractTokenMetadata 从令牌中提取元数据
func (tm *TokenManager) ExtractTokenMetadata(ctx context.Context, tokenString string) (*TokenPayload, error) {
	claims, err := tm.VerifyToken(ctx, tokenString)
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
	}, nil
}

// RefreshTokens 刷新令牌
func (tm *TokenManager) RefreshTokens(ctx context.Context, refreshToken string) (*TokenDetails, error) {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("refresh").Observe(time.Since(start).Seconds())
	}()

	// 验证刷新令牌
	claims, err := tm.VerifyToken(ctx, refreshToken)
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

	// 获取UUID
	uuid, ok := (*claims)["uuid"].(string)
	if !ok {
		jwtAuthErrors.WithLabelValues("refresh_uuid").Inc()
		return nil, fmt.Errorf("invalid token uuid")
	}

	// 从Redis获取用户ID
	refreshKey := fmt.Sprintf("refresh_token:%s", uuid)
	_, err = tm.redisClient.Get(ctx, refreshKey).Result()
	if err != nil {
		jwtAuthErrors.WithLabelValues("refresh_not_found").Inc()
		return nil, fmt.Errorf("refresh token not found or expired")
	}

	// 删除旧的刷新令牌
	tm.redisClient.Del(ctx, refreshKey)

	// 获取用户信息
	var user models.User
	// 这里需要从数据库获取用户信息，暂时简化处理
	// 实际项目中应该注入数据库连接

	// 创建新的令牌对
	return tm.CreateTokens(ctx, &user, "", "", "")
}

// RevokeToken 撤销令牌
func (tm *TokenManager) RevokeToken(ctx context.Context, tokenString string) error {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("revoke").Observe(time.Since(start).Seconds())
	}()

	claims, err := tm.VerifyToken(ctx, tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	uuid, ok := (*claims)["uuid"].(string)
	if !ok {
		return fmt.Errorf("invalid token uuid")
	}

	tokenType, ok := (*claims)["type"].(string)
	if !ok {
		return fmt.Errorf("invalid token type")
	}

	var key string
	switch tokenType {
	case "access":
		key = fmt.Sprintf("access_token:%s", uuid)
	case "refresh":
		key = fmt.Sprintf("refresh_token:%s", uuid)
	default:
		return fmt.Errorf("unknown token type")
	}

	// 从Redis删除令牌
	err = tm.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens 撤销用户所有令牌
func (tm *TokenManager) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("revoke_all").Observe(time.Since(start).Seconds())
	}()

	pattern := fmt.Sprintf("*token:*:%d", userID)
	keys, err := tm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find user tokens: %w", err)
	}

	if len(keys) > 0 {
		err = tm.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to revoke user tokens: %w", err)
		}
	}

	// 撤销设备信息
	devicePattern := fmt.Sprintf("user_device:%d:*", userID)
	deviceKeys, err := tm.redisClient.Keys(ctx, devicePattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find user devices: %w", err)
	}

	if len(deviceKeys) > 0 {
		err = tm.redisClient.Del(ctx, deviceKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to revoke user devices: %w", err)
		}
	}

	return nil
}

// ValidateAccess 验证访问权限
func (tm *TokenManager) ValidateAccess(ctx context.Context, tokenString string, requiredRoles []string) (*TokenPayload, error) {
	payload, err := tm.ExtractTokenMetadata(ctx, tokenString)
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
	accessUUID := tm.getTokenUUIDFromToken(tokenString)
	if accessUUID == "" {
		return nil, fmt.Errorf("failed to extract token UUID")
	}

	accessKey := fmt.Sprintf("access_token:%s", accessUUID)
	_, err = tm.redisClient.Get(ctx, accessKey).Result()
	if err != nil {
		return nil, fmt.Errorf("token not found or expired")
	}

	return payload, nil
}

// getTokenUUIDFromToken 从令牌字符串中提取UUID
func (tm *TokenManager) getTokenUUIDFromToken(tokenString string) string {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return tm.secret, nil
	})

	if err != nil {
		return ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uuid, ok := claims["uuid"].(string); ok {
			return uuid
		}
	}

	return ""
}

// generateUUID 生成UUID
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// IsTokenBlacklisted 检查令牌是否在黑名单中
func (tm *TokenManager) IsTokenBlacklisted(ctx context.Context, tokenString string) bool {
	accessUUID := tm.getTokenUUIDFromToken(tokenString)
	if accessUUID == "" {
		return true
	}

	blacklistKey := fmt.Sprintf("blacklist:%s", accessUUID)
	_, err := tm.redisClient.Get(ctx, blacklistKey).Result()
	return err == nil
}

// BlacklistToken 将令牌加入黑名单
func (tm *TokenManager) BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	accessUUID := tm.getTokenUUIDFromToken(tokenString)
	if accessUUID == "" {
		return fmt.Errorf("failed to extract token UUID")
	}

	blacklistKey := fmt.Sprintf("blacklist:%s", accessUUID)
	err := tm.redisClient.Set(ctx, blacklistKey, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}
