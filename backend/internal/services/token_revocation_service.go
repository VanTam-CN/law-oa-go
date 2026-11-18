package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// TokenRevocationService 令牌撤销服务接口
type TokenRevocationService interface {
	// 令牌撤销
	RevokeAllTokens(ctx context.Context, userID uint, revokedBy uint, reason string) error
	RevokeTokenType(ctx context.Context, userID uint, tokenType string, revokedBy uint) error
	RevokeSpecificTokens(ctx context.Context, userID uint, tokenIDs []string, revokedBy uint) error

	// 令牌查询
	GetRevocationLogs(ctx context.Context, userID uint) ([]*models.TokenRevocationLog, error)
	IsTokenRevoked(ctx context.Context, tokenID string) (bool, error)

	// 密码重置时撤销
	RevokeOnPasswordReset(ctx context.Context, userID uint) error

	// 清理过期撤销记录
	CleanupOldRecords(ctx context.Context, olderThan time.Duration) (int64, error)
}

// TokenRevocationServiceImpl 令牌撤销服务实现
type TokenRevocationServiceImpl struct {
	db *gorm.DB
}

// NewTokenRevocationService 创建令牌撤销服务
func NewTokenRevocationService(db *gorm.DB) TokenRevocationService {
	return &TokenRevocationServiceImpl{db: db}
}

// RevokeAllTokens 撤销用户所有令牌
func (s *TokenRevocationServiceImpl) RevokeAllTokens(ctx context.Context, userID uint, revokedBy uint, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 记录撤销日志
		log := &models.TokenRevocationLog{
			UserID:         userID,
			RevocationType: "manual",
			RevokedBy:      &revokedBy,
			RevokeAll:      true,
			RevokedTokens:  models.JSON{"tokens": []string{}},
			RevokedAt:      time.Now(),
		}

		if err := tx.Create(log).Error; err != nil {
			return fmt.Errorf("记录撤销日志失败: %w", err)
		}

		// TODO: 实际撤销操作 - 这里需要根据实际的令牌存储机制来实现
		// 例如: 清除 Redis 中的 session、删除 JWT 黑名单等

		return nil
	})
}

// RevokeTokenType 撤销特定类型的令牌
func (s *TokenRevocationServiceImpl) RevokeTokenType(ctx context.Context, userID uint, tokenType string, revokedBy uint) error {
	now := time.Now()
	return s.db.WithContext(ctx).Create(&models.TokenRevocationLog{
		UserID:         userID,
		RevocationType: "manual",
		RevokedBy:      &revokedBy,
		RevokeAll:      false,
		RevokedTokens:  models.JSON{"tokens": []string{tokenType}},
		TokenType:      tokenType,
		RevokedAt:      now,
	}).Error
}

// RevokeSpecificTokens 撤销指定的令牌
func (s *TokenRevocationServiceImpl) RevokeSpecificTokens(ctx context.Context, userID uint, tokenIDs []string, revokedBy uint) error {
	if len(tokenIDs) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Create(&models.TokenRevocationLog{
		UserID:         userID,
		RevocationType: "manual",
		RevokedBy:      &revokedBy,
		RevokeAll:      false,
		RevokedTokens:  models.JSON{"tokens": tokenIDs},
		RevokedAt:      time.Now(),
	}).Error
}

// GetRevocationLogs 获取撤销日志
func (s *TokenRevocationServiceImpl) GetRevocationLogs(ctx context.Context, userID uint) ([]*models.TokenRevocationLog, error) {
	var logs []models.TokenRevocationLog
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("revoked_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	result := make([]*models.TokenRevocationLog, len(logs))
	for i, l := range logs {
		result[i] = &l
	}
	return result, nil
}

// IsTokenRevoked 检查令牌是否已被撤销
func (s *TokenRevocationServiceImpl) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.TokenRevocationLog{}).
		Where("revoked_tokens LIKE ?", "%"+tokenID+"%").
		Count(&count).Error
	return count > 0, err
}

// RevokeOnPasswordReset 密码重置时撤销所有令牌
func (s *TokenRevocationServiceImpl) RevokeOnPasswordReset(ctx context.Context, userID uint) error {
	return s.db.WithContext(ctx).Create(&models.TokenRevocationLog{
		UserID:         userID,
		RevocationType: "password_reset",
		RevokeAll:      true,
		RevokedAt:      time.Now(),
	}).Error
}

// CleanupOldRecords 清理旧的撤销记录
func (s *TokenRevocationServiceImpl) CleanupOldRecords(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := s.db.WithContext(ctx).
		Where("revoked_at < ?", cutoff).
		Delete(&models.TokenRevocationLog{})
	return result.RowsAffected, result.Error
}
