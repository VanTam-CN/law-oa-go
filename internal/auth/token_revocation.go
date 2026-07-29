package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// TokenManagerInterface 令牌管理器接口（用于依赖注入）
type TokenManagerInterface interface {
	VerifyToken(ctx context.Context, tokenString string) (*map[string]interface{}, error)
	ExtractTokenMetadata(ctx context.Context, tokenString string) (*TokenPayload, error)
	RevokeAllUserTokens(ctx context.Context, userID uint) error
	BlacklistToken(ctx context.Context, tokenString string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, tokenString string) bool
}

// RevocationReason 撤销原因
type RevocationReason string

const (
	// RevokeAll 离职 - 撤销用户所有令牌
	RevokeAll RevocationReason = "user_offboarding"
	// RevokeByUser 密码重置 - 撤销用户所有令牌
	RevokeByUser RevocationReason = "password_reset"
	// RevokeByDevice 安全事件 - 撤销设备所有令牌
	RevokeByDevice RevocationReason = "security_event"
	// RevokeSingle 登出 - 撤销单个令牌
	RevokeSingle RevocationReason = "user_logout"
)

// RevocationResult 撤销结果
type RevocationResult struct {
	RevokedCount int              `json:"revoked_count"`
	Reason       RevocationReason `json:"reason"`
	RevokedAt    time.Time        `json:"revoked_at"`
	Message      string           `json:"message"`
}

// TokenRevocationService 令牌撤销服务
type TokenRevocationService struct {
	tokenManager TokenManagerInterface
	redisClient  *redis.Client
	db           *gorm.DB
}

// NewTokenRevocationService 创建令牌撤销服务
func NewTokenRevocationService(
	tokenManager TokenManagerInterface,
	redisClient *redis.Client,
	db *gorm.DB,
) *TokenRevocationService {
	return &TokenRevocationService{
		tokenManager: tokenManager,
		redisClient:  redisClient,
		db:           db,
	}
}

// RevokeSingle 撤销单个令牌（用户登出）
func (s *TokenRevocationService) RevokeSingle(ctx context.Context, tokenString, ipAddress string) (*RevocationResult, error) {
	start := time.Now()

	// 1. 验证并提取令牌信息
	claims, err := s.tokenManager.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	tokenType, _ := (*claims)["type"].(string)
	if tokenType == "" {
		tokenType = "access"
	}
	uuid, _ := (*claims)["uuid"].(string)
	if uuid == "" {
		uuid, _ = (*claims)["jti"].(string)
	}

	ttl := ttlFromClaims(*claims)
	if err := s.blacklistToken(ctx, tokenString, uuid, ttl); err != nil {
		return nil, fmt.Errorf("failed to blacklist token: %w", err)
	}

	if s.redisClient != nil && uuid != "" {
		switch tokenType {
		case "access":
			_ = s.redisClient.Del(ctx, fmt.Sprintf("access_token:%s", uuid)).Err()
		case "refresh":
			_ = s.redisClient.Del(ctx, fmt.Sprintf("refresh_token:%s", uuid)).Err()
		}
	}

	if payload, err := s.tokenManager.ExtractTokenMetadata(ctx, tokenString); err == nil {
		_ = s.recordRevocation(ctx, payload.UserID, tokenType, uuid, RevokeSingle, ipAddress)
	} else if userID, ok := userIDFromClaims(*claims); ok {
		_ = s.recordRevocation(ctx, userID, tokenType, uuid, RevokeSingle, ipAddress)
	}

	return &RevocationResult{
		RevokedCount: 1,
		Reason:       RevokeSingle,
		RevokedAt:    start,
		Message:      "Token revoked successfully",
	}, nil
}

// RevokeByUser 撤销用户所有令牌（密码重置）
func (s *TokenRevocationService) RevokeByUser(ctx context.Context, userID uint, ipAddress string) (*RevocationResult, error) {
	start := time.Now()

	// 1. 记录用户级撤销时间，认证中间件会拒绝此时间点前签发的令牌
	if err := s.markUserTokensRevoked(ctx, userID, start); err != nil {
		return nil, fmt.Errorf("failed to mark user tokens revoked: %w", err)
	}
	if err := s.tokenManager.RevokeAllUserTokens(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	// 2. 记录撤销到数据库
	_ = s.recordRevocationEvent(ctx, userID, RevokeByUser, ipAddress)

	return &RevocationResult{
		RevokedCount: -1, // 表示所有令牌
		Reason:       RevokeByUser,
		RevokedAt:    start,
		Message:      fmt.Sprintf("All tokens revoked for user %d", userID),
	}, nil
}

// RevokeByDevice 撤销设备所有令牌（安全事件）
func (s *TokenRevocationService) RevokeByDevice(ctx context.Context, userID uint, deviceID, ipAddress string) (*RevocationResult, error) {
	start := time.Now()

	// 1. 查找该设备的所有令牌
	devicePattern := fmt.Sprintf("user_device:%d:%s", userID, deviceID)
	deviceKeys, err := s.redisClient.Keys(ctx, devicePattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to find device tokens: %w", err)
	}

	revokedCount := 0
	var revokedTokens []map[string]interface{}

	for _, deviceKey := range deviceKeys {
		// 从设备信息中提取令牌UUID
		var deviceInfo map[string]interface{}
		err = s.redisClient.Get(ctx, deviceKey).Scan(&deviceInfo)
		if err != nil {
			continue
		}

		// 撤销访问令牌
		if accessUUID, ok := deviceInfo["access_uuid"].(string); ok && accessUUID != "" {
			key := fmt.Sprintf("access_token:%s", accessUUID)
			if s.redisClient.Del(ctx, key).Err() == nil {
				revokedCount++
				revokedTokens = append(revokedTokens, map[string]interface{}{"type": "access", "uuid": accessUUID})
			}
			// 加入黑名单
			s.redisClient.Set(ctx, fmt.Sprintf("blacklist:%s", accessUUID), "1", time.Hour*24)
		}

		// 撤销刷新令牌
		if refreshUUID, ok := deviceInfo["refresh_uuid"].(string); ok && refreshUUID != "" {
			key := fmt.Sprintf("refresh_token:%s", refreshUUID)
			if s.redisClient.Del(ctx, key).Err() == nil {
				revokedCount++
				revokedTokens = append(revokedTokens, map[string]interface{}{"type": "refresh", "uuid": refreshUUID})
			}
		}

		// 删除设备信息
		s.redisClient.Del(ctx, deviceKey)
	}

	// 2. 记录撤销到数据库
	if revokedCount > 0 {
		log := &models.TokenRevocationLog{
			UserID:         userID,
			RevocationType: string(RevokeByDevice),
			RevokeAll:      false,
			RevokedTokens:  models.JSON{"tokens": revokedTokens, "device_id": deviceID},
			TokenType:      "device",
			RevokedAt:      time.Now(),
			IPAddress:      ipAddress,
		}
		_ = s.db.WithContext(ctx).Create(log).Error
	}

	return &RevocationResult{
		RevokedCount: revokedCount,
		Reason:       RevokeByDevice,
		RevokedAt:    start,
		Message:      fmt.Sprintf("Revoked %d token(s) for device %s", revokedCount, deviceID),
	}, nil
}

// RevokeAll 撤销用户所有令牌（离职）
func (s *TokenRevocationService) RevokeAll(ctx context.Context, userID uint, offboardingData *OffboardingData, ipAddress string) (*RevocationResult, error) {
	start := time.Now()

	// 1. 记录用户级撤销时间，认证中间件会拒绝此时间点前签发的令牌
	if err := s.markUserTokensRevoked(ctx, userID, start); err != nil {
		return nil, fmt.Errorf("failed to mark user tokens revoked: %w", err)
	}
	if err := s.tokenManager.RevokeAllUserTokens(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	// 2. 撤销所有设备的令牌
	if s.redisClient != nil {
		devicePattern := fmt.Sprintf("user_device:%d:*", userID)
		deviceKeys, _ := s.redisClient.Keys(ctx, devicePattern).Result()
		for _, key := range deviceKeys {
			s.redisClient.Del(ctx, key)
		}
	}

	// 3. 记录撤销到数据库
	_ = s.recordOffboardingEvent(ctx, userID, offboardingData, ipAddress)

	return &RevocationResult{
		RevokedCount: -1,
		Reason:       RevokeAll,
		RevokedAt:    start,
		Message:      fmt.Sprintf("All tokens and devices revoked for user %d (offboarding)", userID),
	}, nil
}

// IsTokenRevoked 检查令牌是否已被撤销
func (s *TokenRevocationService) IsTokenRevoked(ctx context.Context, tokenString string) bool {
	return s.isTokenStringRevoked(ctx, tokenString)
}

func (s *TokenRevocationService) IsTokenRevokedForClaims(ctx context.Context, tokenString string, userID uint, issuedAt time.Time) bool {
	if s.isTokenStringRevoked(ctx, tokenString) {
		return true
	}
	if s.redisClient == nil || userID == 0 || issuedAt.IsZero() {
		return false
	}
	value, err := s.redisClient.Get(ctx, userRevokedAfterKey(userID)).Int64()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		return true
	}
	return !issuedAt.After(time.Unix(value, 0))
}

// GetUserActiveDevices 获取用户的活动设备
func (s *TokenRevocationService) GetUserActiveDevices(ctx context.Context, userID uint) ([]map[string]interface{}, error) {
	pattern := fmt.Sprintf("user_device:%d:*", userID)
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to find user devices: %w", err)
	}

	devices := make([]map[string]interface{}, 0, len(keys))
	for _, key := range keys {
		var deviceInfo map[string]interface{}
		err = s.redisClient.Get(ctx, key).Scan(&deviceInfo)
		if err == nil {
			devices = append(devices, deviceInfo)
		}
	}

	return devices, nil
}

// GetRevocationHistory 获取用户的撤销历史
func (s *TokenRevocationService) GetRevocationHistory(ctx context.Context, userID uint, limit int) ([]*models.TokenRevocationLog, error) {
	var logs []*models.TokenRevocationLog
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("revoked_at DESC").
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query revocation history: %w", err)
	}

	return logs, nil
}

func (s *TokenRevocationService) blacklistToken(ctx context.Context, tokenString, uuid string, ttl time.Duration) error {
	if s.redisClient == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	if err := s.redisClient.Set(ctx, tokenHashKey(tokenString), "1", ttl).Err(); err != nil {
		return err
	}
	if uuid != "" {
		if err := s.redisClient.Set(ctx, fmt.Sprintf("blacklist:%s", uuid), "1", ttl).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *TokenRevocationService) isTokenStringRevoked(ctx context.Context, tokenString string) bool {
	if s.redisClient == nil {
		return false
	}
	if _, err := s.redisClient.Get(ctx, tokenHashKey(tokenString)).Result(); err == nil {
		return true
	} else if err != redis.Nil {
		return true
	}
	return s.tokenManager.IsTokenBlacklisted(ctx, tokenString)
}

func (s *TokenRevocationService) markUserTokensRevoked(ctx context.Context, userID uint, revokedAt time.Time) error {
	if s.redisClient == nil {
		return nil
	}
	return s.redisClient.Set(ctx, userRevokedAfterKey(userID), revokedAt.Unix(), 0).Err()
}

func tokenHashKey(tokenString string) string {
	sum := sha256.Sum256([]byte(tokenString))
	return "jwt_blacklist:" + hex.EncodeToString(sum[:])
}

func userRevokedAfterKey(userID uint) string {
	return fmt.Sprintf("user_revoked_after:%d", userID)
}

func ttlFromClaims(claims map[string]interface{}) time.Duration {
	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Hour
	}
	ttl := time.Until(time.Unix(int64(exp), 0))
	if ttl <= 0 {
		return time.Second
	}
	return ttl
}

func userIDFromClaims(claims map[string]interface{}) (uint, bool) {
	switch value := claims["user_id"].(type) {
	case float64:
		return uint(value), value > 0
	case uint:
		return value, value > 0
	case int:
		return uint(value), value > 0
	default:
		return 0, false
	}
}

// CleanupOldRevocationLogs 清理旧的撤销记录（超过90天）
func (s *TokenRevocationService) CleanupOldRevocationLogs(ctx context.Context) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -90)
	result := s.db.WithContext(ctx).
		Where("revoked_at < ?", cutoffDate).
		Delete(&models.TokenRevocationLog{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old revocation logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// recordRevocation 记录单个令牌撤销信息到数据库
func (s *TokenRevocationService) recordRevocation(ctx context.Context, userID uint, tokenType, uuid string, reason RevocationReason, ipAddress string) error {
	revokedTokens := []map[string]interface{}{
		{"type": tokenType, "uuid": uuid},
	}

	log := &models.TokenRevocationLog{
		UserID:         userID,
		RevocationType: string(reason),
		RevokeAll:      false,
		RevokedTokens:  models.JSON{"tokens": revokedTokens},
		TokenType:      tokenType,
		RevokedAt:      time.Now(),
		IPAddress:      ipAddress,
	}

	return s.db.WithContext(ctx).Create(log).Error
}

// recordRevocationEvent 记录批量撤销事件
func (s *TokenRevocationService) recordRevocationEvent(ctx context.Context, userID uint, revocationType RevocationReason, ipAddress string) error {
	log := &models.TokenRevocationLog{
		UserID:         userID,
		RevocationType: string(revocationType),
		RevokeAll:      true,
		RevokedTokens:  models.JSON{"tokens": []string{}}, // 批量撤销时不需要详细记录
		TokenType:      "all",
		RevokedAt:      time.Now(),
		IPAddress:      ipAddress,
	}

	return s.db.WithContext(ctx).Create(log).Error
}

// recordOffboardingEvent 记录离职事件
func (s *TokenRevocationService) recordOffboardingEvent(ctx context.Context, userID uint, data *OffboardingData, ipAddress string) error {
	notes := fmt.Sprintf("Offboarding: successor=%d, transferred_cases=%d", data.SuccessorID, data.TransferredCaseCount)
	if data.HandoverNote != "" {
		notes += ", note: " + data.HandoverNote
	}

	log := &models.TokenRevocationLog{
		UserID:         userID,
		RevocationType: string(RevokeAll),
		RevokeAll:      true,
		RevokedTokens: models.JSON{
			"tokens":            []string{},
			"successor_id":      data.SuccessorID,
			"transferred_cases": data.TransferredCaseIDs,
			"notes":             data.HandoverNote,
		},
		TokenType: "offboarding",
		RevokedAt: time.Now(),
		IPAddress: ipAddress,
	}

	return s.db.WithContext(ctx).Create(log).Error
}

// OffboardingData 离职交接数据
type OffboardingData struct {
	SuccessorID          uint   `json:"successor_id"`
	TransferredCaseCount int    `json:"transferred_case_count"`
	TransferredCaseIDs   []uint `json:"transferred_case_ids"`
	HandoverNote         string `json:"handover_note"`
}

// TokenManagerAdapter 适配器，让 TokenManager 实现 TokenManagerInterface
type TokenManagerAdapter struct {
	*TokenManager
}

// VerifyToken 实现接口方法，将返回类型转换为 *map[string]interface{}
func (a *TokenManagerAdapter) VerifyToken(ctx context.Context, tokenString string) (*map[string]interface{}, error) {
	claims, err := a.TokenManager.VerifyToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	// 转换 *jwt.MapClaims 到 *map[string]interface{}
	result := map[string]interface{}(*claims)
	return &result, nil
}

// NewTokenManagerAdapter 创建适配器
func NewTokenManagerAdapter(tm *TokenManager) TokenManagerInterface {
	return &TokenManagerAdapter{TokenManager: tm}
}
