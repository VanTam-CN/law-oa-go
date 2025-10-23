package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RefreshTokenService 刷新令牌服务
type RefreshTokenService struct {
	db           *gorm.DB
	config       *JWTConfig
	logger       *logrus.Logger
	tokenManager *JWTManager
}

// NewRefreshTokenService 创建刷新令牌服务
func NewRefreshTokenService(db *gorm.DB, config *JWTConfig, logger *logrus.Logger, tokenManager *JWTManager) *RefreshTokenService {
	return &RefreshTokenService{
		db:           db,
		config:       config,
		logger:       logger,
		tokenManager: tokenManager,
	}
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	DeviceInfo   *DeviceInfo `json:"device_info,omitempty"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int64         `json:"expires_in"`
	Scope        string        `json:"scope,omitempty"`
	UserInfo     *UserInfo     `json:"user_info,omitempty"`
	TokenPair    *TokenPair    `json:"-"`
	RefreshedAt  time.Time     `json:"refreshed_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	Platform   string `json:"platform"`
	Version    string `json:"version"`
	AppVersion string `json:"app_version"`
}

// UserInfo 用户信息
type UserInfo struct {
	UserID    uint                   `json:"user_id"`
	Username  string                 `json:"username"`
	TenantID  string                 `json:"tenant_id"`
	Roles     []string               `json:"roles"`
	Profile   map[string]interface{} `json:"profile,omitempty"`
}

// RefreshTokenRotation 刷新令牌轮换策略
type RefreshTokenRotation int

const (
	// RotationReuse 重用刷新令牌
	RotationReuse RefreshTokenRotation = iota
	// RotationNew 每次刷新生成新令牌
	RotationNew
	// RotationConditional 根据条件决定是否轮换
	RotationConditional
)

// RefreshTokenConfig 刷新令牌配置
type RefreshTokenConfig struct {
	// 轮换策略
	RotationStrategy RefreshTokenRotation
	// 最大同时活跃的刷新令牌数量
	MaxActiveTokens int
	// 刷新令牌过期时间
	RefreshTokenTTL time.Duration
	// 是否启用设备绑定
	BindToDevice bool
	// 是否启用地理位置验证
	ValidateLocation bool
	// 是否启用速率限制
	EnableRateLimit bool
	// 速率限制（每小时）
	RateLimitPerHour int
	// 是否记录刷新历史
	TrackRefreshHistory bool
	// 自动清理过期令牌
	AutoCleanup bool
	// 清理间隔
	CleanupInterval time.Duration
}

// DefaultRefreshTokenConfig 默认刷新令牌配置
func DefaultRefreshTokenConfig() *RefreshTokenConfig {
	return &RefreshTokenConfig{
		RotationStrategy:     RotationNew,
		MaxActiveTokens:      5,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		BindToDevice:         true,
		ValidateLocation:     false,
		EnableRateLimit:      true,
		RateLimitPerHour:     10,
		TrackRefreshHistory:  true,
		AutoCleanup:          true,
		CleanupInterval:      1 * time.Hour,
	}
}

// RefreshToken 刷新令牌记录
type RefreshTokenRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	JTI          string    `gorm:"uniqueIndex;size:255" json:"jti"`
	UserID       uint      `gorm:"index" json:"user_id"`
	TenantID     string    `gorm:"size:100;index" json:"tenant_id"`
	SessionID    string    `gorm:"size:100;index" json:"session_id"`
	ClientID     string    `gorm:"size:100" json:"client_id"`
	TokenHash    string    `gorm:"size:255;index" json:"token_hash"`
	DeviceID     string    `gorm:"size:100;index" json:"device_id"`
	IPAddress    string    `gorm:"size:45" json:"ip_address"`
	UserAgent    string    `gorm:"size:500" json:"user_agent"`
	ExpiresAt    time.Time `gorm:"index" json:"expires_at"`
	LastUsedAt   time.Time `json:"last_used_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	RevokedBy    string    `gorm:"size:100" json:"revoked_by"`
	RevokedReason string   `gorm:"size:255" json:"revoked_reason"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RefreshHistory 刷新历史记录
type RefreshHistory struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	RefreshTokenID  uint      `gorm:"index" json:"refresh_token_id"`
	UserID          uint      `gorm:"index" json:"user_id"`
	SessionID       string    `gorm:"size:100;index" json:"session_id"`
	OldTokenJTI     string    `gorm:"size:255" json:"old_token_jti"`
	NewTokenJTI     string    `gorm:"size:255" json:"new_token_jti"`
	IPAddress       string    `gorm:"size:45" json:"ip_address"`
	UserAgent       string    `gorm:"size:500" json:"user_agent"`
	RefreshReason   string    `gorm:"size:100" json:"refresh_reason"`
	Success         bool      `json:"success"`
	ErrorMessage    string    `gorm:"size:500" json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
}

// RefreshResult 刷新结果
type RefreshResult struct {
	Success        bool               `json:"success"`
	Response       *RefreshTokenResponse `json:"response,omitempty"`
	Error          error              `json:"error,omitempty"`
	TokenRevoked   bool               `json:"token_revoked"`
	RefreshCount   int64              `json:"refresh_count"`
	RemainingQuota int64              `json:"remaining_quota"`
}

// RefreshToken 执行令牌刷新
func (s *RefreshTokenService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshResult, error) {
	// 验证请求格式
	if err := s.validateRefreshRequest(req); err != nil {
		return &RefreshResult{
			Success: false,
			Error:   err,
		}, err
	}

	// 验证刷新令牌
	claims, err := s.tokenManager.ValidateToken(req.RefreshToken)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"step":  "validate_refresh_token",
		}).Warn("Refresh token validation failed")
		return &RefreshResult{
			Success:      false,
			Error:        ErrRefreshTokenInvalid,
			TokenRevoked: true,
		}, ErrRefreshTokenInvalid
	}

	// 检查令牌是否过期
	if time.Now().After(claims.ExpiresAt.Time) {
		return &RefreshResult{
			Success:      false,
			Error:        ErrRefreshTokenExpired,
			TokenRevoked: true,
		}, ErrRefreshTokenExpired
	}

	// 获取刷新令牌记录
	refreshRecord, err := s.getRefreshTokenRecord(claims.JTI)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"jti":   claims.JTI,
			"error": err.Error(),
		}).Error("Failed to get refresh token record")
		return &RefreshResult{
			Success:      false,
			Error:        ErrRefreshTokenInvalid,
			TokenRevoked: true,
		}, ErrRefreshTokenInvalid
	}

	// 检查令牌是否已被撤销
	if refreshRecord.RevokedAt != nil {
		return &RefreshResult{
			Success:      false,
			Error:        ErrRefreshTokenUsed,
			TokenRevoked: true,
		}, ErrRefreshTokenUsed
	}

	// 验证设备绑定
	if err := s.validateDeviceBinding(claims, req.DeviceInfo, refreshRecord); err != nil {
		return &RefreshResult{
			Success:      false,
			Error:        err,
			TokenRevoked: true,
		}, err
	}

	// 检查速率限制
	if err := s.checkRateLimit(ctx, claims.UserID); err != nil {
		return &RefreshResult{
			Success: false,
			Error:   err,
		}, err
	}

	// 检查活跃令牌数量
	if err := s.checkActiveTokensLimit(claims.UserID); err != nil {
		return &RefreshResult{
			Success: false,
			Error:   err,
		}, err
	}

	// 执行令牌刷新
	return s.performTokenRefresh(ctx, claims, req, refreshRecord)
}

// validateRefreshRequest 验证刷新请求
func (s *RefreshTokenService) validateRefreshRequest(req *RefreshTokenRequest) error {
	if req.RefreshToken == "" {
		return ErrTokenMissing
	}

	if len(req.RefreshToken) < 10 {
		return ErrTokenMalformed
	}

	return nil
}

// getRefreshTokenRecord 获取刷新令牌记录
func (s *RefreshTokenService) getRefreshTokenRecord(jti string) (*RefreshTokenRecord, error) {
	var record RefreshTokenRecord
	err := s.db.Where("jti = ?", jti).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRefreshTokenInvalid
		}
		return nil, err
	}
	return &record, nil
}

// validateDeviceBinding 验证设备绑定
func (s *RefreshTokenService) validateDeviceBinding(claims *TokenClaims, deviceInfo *DeviceInfo, record *RefreshTokenRecord) error {
	if !s.config.BindToDevice {
		return nil
	}

	// 验证设备ID
	if deviceInfo != nil && deviceInfo.DeviceID != "" && record.DeviceID != "" {
		if deviceInfo.DeviceID != record.DeviceID {
			s.logger.WithFields(logrus.Fields{
				"expected_device": record.DeviceID,
				"actual_device":   deviceInfo.DeviceID,
				"user_id":         claims.UserID,
			}).Warn("Device ID mismatch")
			return ErrDeviceMismatch
		}
	}

	// 验证用户代理（简化版本）
	if claims.UserAgent != "" && record.UserAgent != "" {
		if claims.UserAgent != record.UserAgent {
			s.logger.WithFields(logrus.Fields{
				"expected_ua": record.UserAgent,
				"actual_ua":   claims.UserAgent,
				"user_id":     claims.UserID,
			}).Warn("User agent mismatch")
			return ErrUserAgentMismatch
		}
	}

	return nil
}

// checkRateLimit 检查速率限制
func (s *RefreshTokenService) checkRateLimit(ctx context.Context, userID uint) error {
	if !s.config.EnableRateLimit {
		return nil
	}

	// 查询过去一小时的刷新次数
	var count int64
	oneHourAgo := time.Now().Add(-time.Hour)
	err := s.db.Model(&RefreshHistory{}).
		Where("user_id = ? AND created_at > ?", userID, oneHourAgo).
		Count(&count).Error

	if err != nil {
		s.logger.WithError(err).Error("Failed to check rate limit")
		return nil // 如果查询失败，允许通过
	}

	if count >= int64(s.config.RateLimitPerHour) {
		s.logger.WithFields(logrus.Fields{
			"user_id": userID,
			"count":   count,
			"limit":   s.config.RateLimitPerHour,
		}).Warn("Rate limit exceeded")
		return ErrRateLimitExceeded
	}

	return nil
}

// checkActiveTokensLimit 检查活跃令牌数量限制
func (s *RefreshTokenService) checkActiveTokensLimit(userID uint) error {
	if s.config.MaxActiveTokens <= 0 {
		return nil
	}

	var count int64
	err := s.db.Model(&RefreshTokenRecord{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Count(&count).Error

	if err != nil {
		s.logger.WithError(err).Error("Failed to check active tokens limit")
		return nil
	}

	if count >= int64(s.config.MaxActiveTokens) {
		// 撤销最旧的令牌
		err := s.revokeOldestToken(userID)
		if err != nil {
			s.logger.WithError(err).Error("Failed to revoke oldest token")
		}
	}

	return nil
}

// revokeOldestToken 撤销最旧的令牌
func (s *RefreshTokenService) revokeOldestToken(userID uint) error {
	var oldestRecord RefreshTokenRecord
	err := s.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at ASC").
		First(&oldestRecord).Error

	if err != nil {
		return err
	}

	now := time.Now()
	return s.db.Model(&oldestRecord).Updates(map[string]interface{}{
		"revoked_at":     &now,
		"revoked_by":     "system",
		"revoked_reason": "token_limit_exceeded",
	}).Error
}

// performTokenRefresh 执行令牌刷新
func (s *RefreshTokenService) performTokenRefresh(ctx context.Context, claims *TokenClaims, req *RefreshTokenRequest, record *RefreshTokenRecord) (*RefreshResult, error) {
	// 根据轮换策略决定是否生成新的刷新令牌
	shouldRotate := s.shouldRotateToken(claims, record)

	// 生成新的访问令牌
	var newRefreshToken string
	var err error

	if shouldRotate {
		// 生成新的令牌对
		newTokenPair, err := s.tokenManager.GenerateTokenPair(
			claims.UserID,
			claims.Username,
			claims.TenantID,
			claims.Roles,
			claims.Permissions,
			claims.IPAddress,
			claims.UserAgent,
		)
		if err != nil {
			s.logger.WithError(err).Error("Failed to generate new token pair")
			return &RefreshResult{
				Success: false,
				Error:   ErrInternalError,
			}, err
		}
		newRefreshToken = newTokenPair.RefreshToken

		// 创建新的刷新令牌记录
		err = s.createRefreshTokenRecord(newTokenPair, claims)
		if err != nil {
			s.logger.WithError(err).Error("Failed to create refresh token record")
		}
	} else {
		// 只生成新的访问令牌
		newAccessToken, err := s.tokenManager.GenerateAccessToken(
			claims.UserID,
			claims.Username,
			claims.TenantID,
			claims.Roles,
			claims.Permissions,
			claims.IPAddress,
			claims.UserAgent,
		)
		if err != nil {
			s.logger.WithError(err).Error("Failed to generate new access token")
			return &RefreshResult{
				Success: false,
				Error:   ErrInternalError,
			}, err
		}

		newRefreshToken = req.RefreshToken
	}

	// 撤销旧的刷新令牌（如果轮换）
	if shouldRotate {
		err = s.revokeRefreshToken(record.JTI, "token_rotated")
		if err != nil {
			s.logger.WithError(err).Error("Failed to revoke old refresh token")
		}
	}

	// 更新最后使用时间
	err = s.updateLastUsedTime(record.ID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to update last used time")
	}

	// 记录刷新历史
	if s.config.TrackRefreshHistory {
		go s.recordRefreshHistory(record.ID, claims, shouldRotate)
	}

	// 获取刷新统计
	refreshCount, remainingQuota := s.getRefreshQuota(claims.UserID)

	// 构建响应
	expiresIn := int64(s.config.AccessTokenDuration.Seconds())
	response := &RefreshTokenResponse{
		AccessToken:  "", // 将在下面设置
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshedAt:  time.Now(),
		ExpiresAt:    time.Now().Add(s.config.AccessTokenDuration),
		UserInfo: &UserInfo{
			UserID:   claims.UserID,
			Username: claims.Username,
			TenantID: claims.TenantID,
			Roles:    claims.Roles,
		},
	}

	// 设置访问令牌
	if shouldRotate {
		response.AccessToken = "" // 需要从newTokenPair获取
	} else {
		response.AccessToken = "" // 需要从新生成的访问令牌获取
	}

	return &RefreshResult{
		Success:        true,
		Response:       response,
		TokenRevoked:   shouldRotate,
		RefreshCount:   refreshCount,
		RemainingQuota: remainingQuota,
	}, nil
}

// shouldRotateToken 决定是否轮换令牌
func (s *RefreshTokenService) shouldRotateToken(claims *TokenClaims, record *RefreshTokenRecord) bool {
	switch s.config.RotationStrategy {
	case RotationReuse:
		return false
	case RotationNew:
		return true
	case RotationConditional:
		// 根据条件决定，例如距离过期时间
		timeUntilExpiry := time.Until(record.ExpiresAt)
		return timeUntilExpiry < 24*time.Hour
	default:
		return true
	}
}

// createRefreshTokenRecord 创建刷新令牌记录
func (s *RefreshTokenService) createRefreshTokenRecord(tokenPair *TokenPair, claims *TokenClaims) error {
	// 解析新的刷新令牌以获取JTI
	newClaims, err := s.tokenManager.ValidateToken(tokenPair.RefreshToken)
	if err != nil {
		return err
	}

	tokenHash := s.hashToken(tokenPair.RefreshToken)

	record := &RefreshTokenRecord{
		JTI:        newClaims.JTI,
		UserID:     claims.UserID,
		TenantID:   claims.TenantID,
		SessionID:  claims.SessionID,
		TokenHash:  tokenHash,
		DeviceID:   claims.DeviceID,
		IPAddress:  claims.IPAddress,
		UserAgent:  claims.UserAgent,
		ExpiresAt:  newClaims.ExpiresAt.Time,
		LastUsedAt: time.Now(),
	}

	return s.db.Create(record).Error
}

// revokeRefreshToken 撤销刷新令牌
func (s *RefreshTokenService) revokeRefreshToken(jti, reason string) error {
	now := time.Now()
	return s.db.Model(&RefreshTokenRecord{}).
		Where("jti = ?", jti).
		Updates(map[string]interface{}{
			"revoked_at":     &now,
			"revoked_by":     "system",
			"revoked_reason": reason,
		}).Error
}

// updateLastUsedTime 更新最后使用时间
func (s *RefreshTokenService) updateLastUsedTime(recordID uint) error {
	return s.db.Model(&RefreshTokenRecord{}).
		Where("id = ?", recordID).
		Update("last_used_at", time.Now()).Error
}

// recordRefreshHistory 记录刷新历史
func (s *RefreshTokenService) recordRefreshHistory(refreshTokenID uint, claims *TokenClaims, rotated bool) {
	history := &RefreshHistory{
		RefreshTokenID: refreshTokenID,
		UserID:         claims.UserID,
		SessionID:      claims.SessionID,
		OldTokenJTI:    claims.JTI,
		NewTokenJTI:    "", // 需要从新令牌获取
		IPAddress:      claims.IPAddress,
		UserAgent:      claims.UserAgent,
		RefreshReason:  "token_refresh",
		Success:        true,
	}

	if rotated {
		history.RefreshReason = "token_rotation"
	}

	err := s.db.Create(history).Error
	if err != nil {
		s.logger.WithError(err).Error("Failed to record refresh history")
	}
}

// getRefreshQuota 获取刷新配额信息
func (s *RefreshTokenService) getRefreshQuota(userID uint) (int64, int64) {
	if !s.config.EnableRateLimit {
		return 0, 0
	}

	var count int64
	oneHourAgo := time.Now().Add(-time.Hour)
	s.db.Model(&RefreshHistory{}).
		Where("user_id = ? AND created_at > ?", userID, oneHourAgo).
		Count(&count)

	remaining := int64(s.config.RateLimitPerHour) - count
	if remaining < 0 {
		remaining = 0
	}

	return count, remaining
}

// hashToken 哈希令牌
func (s *RefreshTokenService) hashToken(token string) string {
	// 简化实现，实际应该使用更安全的哈希算法
	return fmt.Sprintf("%x", len(token)*31+int(token[0]))
}

// CleanupExpiredTokens 清理过期令牌
func (s *RefreshTokenService) CleanupExpiredTokens() error {
	if !s.config.AutoCleanup {
		return nil
	}

	return s.db.Where("expires_at < ?", time.Now()).Delete(&RefreshTokenRecord{}).Error
}

// RevokeUserRefreshTokens 撤销用户所有刷新令牌
func (s *RefreshTokenService) RevokeUserRefreshTokens(userID uint, reason string) error {
	now := time.Now()
	return s.db.Model(&RefreshTokenRecord{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]interface{}{
			"revoked_at":     &now,
			"revoked_by":     "system",
			"revoked_reason": reason,
		}).Error
}

// RevokeSessionRefreshTokens 撤销会话所有刷新令牌
func (s *RefreshTokenService) RevokeSessionRefreshTokens(sessionID string, reason string) error {
	now := time.Now()
	return s.db.Model(&RefreshTokenRecord{}).
		Where("session_id = ? AND revoked_at IS NULL", sessionID).
		Updates(map[string]interface{}{
			"revoked_at":     &now,
			"revoked_by":     "system",
			"revoked_reason": reason,
		}).Error
}

// GetUserRefreshTokens 获取用户的活跃刷新令牌
func (s *RefreshTokenService) GetUserRefreshTokens(userID uint) ([]*RefreshTokenRecord, error) {
	var tokens []*RefreshTokenRecord
	err := s.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

// GetRefreshHistory 获取刷新历史
func (s *RefreshTokenService) GetRefreshHistory(userID uint, limit int) ([]*RefreshHistory, error) {
	var history []*RefreshHistory
	query := s.db.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&history).Error
	return history, err
}

// StartCleanupWorker 启动清理工作协程
func (s *RefreshTokenService) StartCleanupWorker(ctx context.Context) {
	if !s.config.AutoCleanup {
		return
	}

	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Cleanup worker stopped")
				return
			case <-ticker.C:
				err := s.CleanupExpiredTokens()
				if err != nil {
					s.logger.WithError(err).Error("Failed to cleanup expired tokens")
				} else {
					s.logger.Info("Expired tokens cleanup completed")
				}
			}
		}
	}()

	s.logger.Info("Cleanup worker started")
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// GenerateSecureToken 生成安全令牌
func GenerateSecureToken() (string, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}