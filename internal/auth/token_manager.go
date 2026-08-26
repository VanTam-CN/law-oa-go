package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
)

var ErrRefreshTokenAlreadyRotated = errors.New("refresh token already rotated")

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
	store        *TokenSessionStore
	db           *gorm.DB
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
func NewTokenManager(cfg *config.Config, redisClient *redis.Client, cacheService *cache.CacheService, db *gorm.DB) *TokenManager {
	return &TokenManager{
		secret:       []byte(cfg.JWT.Secret),
		accessTTL:    time.Duration(cfg.JWT.ExpiresIn) * time.Second,
		refreshTTL:   time.Duration(cfg.JWT.RefreshIn) * time.Second,
		redisClient:  redisClient,
		cacheService: cacheService,
		store:        NewTokenSessionStore(db),
		db:           db,
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

	// PostgreSQL is authoritative. Redis is only an optional mirror and must
	// not make token issuance fail when it is unavailable.
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	accessKey := fmt.Sprintf("access_token:%s", td.AccessUUID)
	refreshKey := fmt.Sprintf("refresh_token:%s", td.RefreshUUID)

	session := &models.AuthTokenSession{
		ID:                  generateUUID(),
		UserID:              user.ID,
		DeviceID:            deviceID,
		IP:                  ip,
		UserAgent:           userAgent,
		AccessTokenUUID:     td.AccessUUID,
		RefreshTokenUUID:    td.RefreshUUID,
		AccessTokenExpires:  at,
		RefreshTokenExpires: rt,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := tm.store.Create(ctx, session); err != nil {
		jwtAuthErrors.WithLabelValues("store_access").Inc()
		return nil, fmt.Errorf("failed to store token session: %w", err)
	}

	if tm.redisClient != nil {
		_ = tm.redisClient.Set(ctx, accessKey, user.ID, at.Sub(now)).Err()
		_ = tm.redisClient.Set(ctx, refreshKey, user.ID, rt.Sub(now)).Err()
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

	if tm.cacheService != nil {
		err = tm.cacheService.Set(deviceKey, deviceInfo, tm.refreshTTL)
	}
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
		// Top-level identity claims keep the canonical Gin authentication
		// middleware from falling back to user_id=0 for session-backed tokens.
		"user_id":  user.ID,
		"username": user.Name,
		"role":     user.Role,
		"uuid":     uuid,
		"issuer":   tm.issuer,
		"exp":      expires,
		"iat":      time.Now().Unix(),
		"type":     tokenType,
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

	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(tm.accessTTL).Unix()
	td.RtExpires = time.Now().Add(tm.refreshTTL).Unix()
	td.AccessUUID = generateUUID()
	td.RefreshUUID = generateUUID()
	revokedAt := time.Now()
	var rotatedUserID uint

	err = tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session models.AuthTokenSession
		if err := tx.Where("refresh_token_uuid = ?", uuid).First(&session).Error; err != nil {
			return fmt.Errorf("load refresh session: %w", err)
		}
		if session.RefreshRevokedAt != nil || session.RevokedAt != nil || session.DeviceRevokedAt != nil ||
			!revokedAt.Before(session.RefreshTokenExpires) {
			return ErrRefreshTokenAlreadyRotated
		}

		var user models.User
		if err := tx.First(&user, session.UserID).Error; err != nil {
			return fmt.Errorf("load token user: %w", err)
		}
		if err := lockUserForTokenMutation(tx, user.ID); err != nil {
			return fmt.Errorf("lock token user: %w", err)
		}
		if err := tx.First(&user, session.UserID).Error; err != nil {
			return fmt.Errorf("reload token user: %w", err)
		}
		// The conditional update is the rotation lock. Only one caller can
		// transition an unrevoked, unexpired refresh token to revoked; all
		// concurrent replays see zero rows and roll back.
		result := tx.Model(&models.AuthTokenSession{}).
			Where(
				"refresh_token_uuid = ? AND refresh_revoked_at IS NULL AND revoked_at IS NULL AND device_revoked_at IS NULL AND refresh_token_expires > ?",
				uuid, revokedAt,
			).
			Updates(map[string]interface{}{
				"refresh_revoked_at": revokedAt,
				"updated_at":         revokedAt,
			})
		if result.Error != nil {
			return fmt.Errorf("claim refresh token: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrRefreshTokenAlreadyRotated
		}

		if user.Status != "" && user.Status != "active" {
			return fmt.Errorf("user is not active")
		}
		rotatedUserID = user.ID

		accessToken, err := tm.createToken(&user, td.AccessUUID, td.AtExpires, "access", session.DeviceID, session.IP, session.UserAgent)
		if err != nil {
			return fmt.Errorf("create access token: %w", err)
		}
		refreshToken, err := tm.createToken(&user, td.RefreshUUID, td.RtExpires, "refresh", session.DeviceID, session.IP, session.UserAgent)
		if err != nil {
			return fmt.Errorf("create refresh token: %w", err)
		}
		td.AccessToken = accessToken
		td.RefreshToken = refreshToken

		newSession := &models.AuthTokenSession{
			ID:                  generateUUID(),
			UserID:              user.ID,
			DeviceID:            session.DeviceID,
			IP:                  session.IP,
			UserAgent:           session.UserAgent,
			AccessTokenUUID:     td.AccessUUID,
			RefreshTokenUUID:    td.RefreshUUID,
			AccessTokenExpires:  time.Unix(td.AtExpires, 0),
			RefreshTokenExpires: time.Unix(td.RtExpires, 0),
			CreatedAt:           revokedAt,
			UpdatedAt:           revokedAt,
		}
		if err := tx.Create(newSession).Error; err != nil {
			return fmt.Errorf("create rotated token session: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrRefreshTokenAlreadyRotated) {
			jwtAuthErrors.WithLabelValues("refresh_replay").Inc()
			return nil, fmt.Errorf("refresh token not found or expired")
		}
		jwtAuthErrors.WithLabelValues("refresh_rotate").Inc()
		return nil, fmt.Errorf("rotate refresh token: %w", err)
	}

	// PostgreSQL is authoritative. Redis mirrors are updated only after the
	// atomic rotation commits, and mirror failures do not roll it back.
	if tm.redisClient != nil {
		_ = tm.redisClient.Del(ctx, fmt.Sprintf("refresh_token:%s", uuid)).Err()
		_ = tm.redisClient.Set(ctx, fmt.Sprintf("access_token:%s", td.AccessUUID), rotatedUserID, time.Until(time.Unix(td.AtExpires, 0))).Err()
		_ = tm.redisClient.Set(ctx, fmt.Sprintf("refresh_token:%s", td.RefreshUUID), rotatedUserID, time.Until(time.Unix(td.RtExpires, 0))).Err()
	}

	jwtTokensIssued.WithLabelValues("access").Inc()
	jwtTokensIssued.WithLabelValues("refresh").Inc()
	return td, nil
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

	revokedAt := time.Now()
	rowsAffected, err := tm.store.RevokeTokenUUID(ctx, uuid, tokenType, revokedAt)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("token session not found or already revoked")
	}
	if tm.redisClient != nil {
		_ = tm.redisClient.Del(ctx, key).Err()
	}

	return nil
}

// RevokeAllUserTokens 撤销用户所有令牌
func (tm *TokenManager) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	start := time.Now()
	defer func() {
		jwtAuthDuration.WithLabelValues("revoke_all").Observe(time.Since(start).Seconds())
	}()

	revokedAt := time.Now()
	if err := tm.store.RevokeAllForUser(ctx, userID, revokedAt); err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	if tm.redisClient != nil {
		devicePattern := fmt.Sprintf("user_device:%d:*", userID)
		if deviceKeys, err := tm.redisClient.Keys(ctx, devicePattern).Result(); err == nil && len(deviceKeys) > 0 {
			_ = tm.redisClient.Del(ctx, deviceKeys...).Err()
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

	if !tm.isAccessTokenActive(ctx, accessUUID) {
		return nil, fmt.Errorf("token not found or expired")
	}

	return payload, nil
}

func (tm *TokenManager) isAccessTokenActive(ctx context.Context, uuid string) bool {
	session, err := tm.store.GetByAccessUUID(ctx, uuid)
	if err != nil {
		return false
	}
	return session.AccessRevokedAt == nil && session.RevokedAt == nil && session.DeviceRevokedAt == nil &&
		time.Now().Before(session.AccessTokenExpires)
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

	session, err := tm.store.GetByAccessUUID(ctx, accessUUID)
	if err == nil {
		return session.AccessRevokedAt != nil || session.RevokedAt != nil || session.DeviceRevokedAt != nil ||
			!time.Now().Before(session.AccessTokenExpires)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// PostgreSQL is authoritative. An unavailable session store must not
		// turn an unknown revocation state into an authenticated session.
		return true
	}
	if tm.redisClient == nil {
		return false
	}
	_, err = tm.redisClient.Get(ctx, fmt.Sprintf("blacklist:%s", accessUUID)).Result()
	return err == nil
}

// BlacklistToken 将令牌加入黑名单
func (tm *TokenManager) BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error {
	accessUUID := tm.getTokenUUIDFromToken(tokenString)
	if accessUUID == "" {
		return fmt.Errorf("failed to extract token UUID")
	}

	revokedAt := time.Now()
	if _, err := tm.store.RevokeTokenUUID(ctx, accessUUID, "access", revokedAt); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	if tm.redisClient != nil {
		if err := tm.redisClient.Set(ctx, fmt.Sprintf("blacklist:%s", accessUUID), "1", ttl).Err(); err != nil {
			// PostgreSQL already records revocation; Redis remains an optional mirror.
			return nil
		}
	}

	return nil
}
