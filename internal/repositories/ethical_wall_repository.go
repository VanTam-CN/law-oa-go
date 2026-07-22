package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

// EthicalWallRepositoryErrors 隔离墙仓库错误
var (
	ErrWhitelistEntryNotFound = errors.New("whitelist entry not found")
	ErrWhitelistEntryExists   = errors.New("user already whitelisted for this case")
)

// EthicalWallRepository 隔离墙数据仓库接口
type EthicalWallRepository interface {
	// IsEthicalWallEnabled 检查案件是否启用隔离墙
	IsEthicalWallEnabled(ctx context.Context, caseID uint) (bool, error)

	// EnableEthicalWall 启用案件隔离墙
	EnableEthicalWall(ctx context.Context, caseID, userID uint, description string) error

	// DisableEthicalWall 禁用案件隔离墙
	DisableEthicalWall(ctx context.Context, caseID uint) error

	// IsUserWhitelisted 检查用户是否在案件白名单中
	IsUserWhitelisted(ctx context.Context, caseID, userID uint) (bool, error)

	// AddToWhitelist 添加用户到案件白名单
	AddToWhitelist(ctx context.Context, caseID, userID, grantedBy uint, reason string) error

	// RemoveFromWhitelist 从案件白名单移除用户
	RemoveFromWhitelist(ctx context.Context, caseID, userID uint) error

	// GetWhitelistByCase 获取案件白名单列表
	GetWhitelistByCase(ctx context.Context, caseID uint) ([]*models.CaseEthicalWallWhitelist, error)

	// GetWhitelistByUser 获取用户可访问的隔离墙案件
	GetWhitelistByUser(ctx context.Context, userID uint) ([]*models.CaseEthicalWallWhitelist, error)

	// LogAccessAttempt 记录访问尝试
	LogAccessAttempt(ctx context.Context, caseID, userID uint, accessType, accessResult, ipAddress, userAgent string) error

	// GetAccessLogs 获取访问日志
	GetAccessLogs(ctx context.Context, caseID uint, limit int) ([]*models.EthicalWallAccessLog, error)

	// ClearWhitelist 清空案件白名单
	ClearWhitelist(ctx context.Context, caseID uint) error
}

// EthicalWallRepositoryImpl 隔离墙数据仓库实现
type EthicalWallRepositoryImpl struct {
	db *gorm.DB
}

// NewEthicalWallRepository 创建隔离墙数据仓库实例
func NewEthicalWallRepository(db *gorm.DB) EthicalWallRepository {
	return &EthicalWallRepositoryImpl{db: db}
}

// IsEthicalWallEnabled 检查案件是否启用隔离墙
func (r *EthicalWallRepositoryImpl) IsEthicalWallEnabled(ctx context.Context, caseID uint) (bool, error) {
	var enabled bool
	err := r.db.WithContext(ctx).
		Model(&models.Case{}).
		Where("id = ?", caseID).
		Pluck("ethical_wall_enabled", &enabled).Error
	return enabled, err
}

// EnableEthicalWall 启用案件隔离墙
func (r *EthicalWallRepositoryImpl) EnableEthicalWall(ctx context.Context, caseID, userID uint, description string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.Case{}).
			Where("id = ?", caseID).
			Updates(map[string]interface{}{
				"ethical_wall_enabled":     true,
				"ethical_wall_description": description,
				"ethical_wall_enabled_by":  userID,
				"ethical_wall_enabled_at":  &now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var owner struct {
			LawyerID  uint
			CreatedBy string
		}
		if err := tx.Model(&models.Case{}).
			Select("lawyer_id, created_by").
			Where("id = ?", caseID).
			Take(&owner).Error; err != nil {
			return err
		}

		allowedUsers := map[uint]string{}
		if owner.LawyerID > 0 {
			allowedUsers[owner.LawyerID] = "案件承办律师"
		}
		if createdBy, err := strconv.ParseUint(owner.CreatedBy, 10, 32); err == nil && createdBy > 0 {
			allowedUsers[uint(createdBy)] = "案件创建人"
		}
		if userID > 0 {
			allowedUsers[userID] = "隔离墙启用人"
		}
		for allowedUserID, reason := range allowedUsers {
			entry := &models.CaseEthicalWallWhitelist{
				CaseID:    caseID,
				UserID:    allowedUserID,
				GrantedBy: userID,
				Reason:    reason,
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DisableEthicalWall 禁用案件隔离墙
func (r *EthicalWallRepositoryImpl) DisableEthicalWall(ctx context.Context, caseID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Case{}).
		Where("id = ?", caseID).
		Updates(map[string]interface{}{
			"ethical_wall_enabled":     false,
			"ethical_wall_description": "",
			"ethical_wall_enabled_by":  nil,
			"ethical_wall_enabled_at":  nil,
		}).Error
}

// IsUserWhitelisted 检查用户是否在案件白名单中
func (r *EthicalWallRepositoryImpl) IsUserWhitelisted(ctx context.Context, caseID, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CaseEthicalWallWhitelist{}).
		Where("case_id = ? AND user_id = ?", caseID, userID).
		Count(&count).Error
	return count > 0, err
}

// AddToWhitelist 添加用户到案件白名单
func (r *EthicalWallRepositoryImpl) AddToWhitelist(ctx context.Context, caseID, userID, grantedBy uint, reason string) error {
	// 检查是否已存在
	var existing models.CaseEthicalWallWhitelist
	err := r.db.WithContext(ctx).
		Where("case_id = ? AND user_id = ?", caseID, userID).
		First(&existing).Error

	if err == nil {
		// 记录已存在
		return ErrWhitelistEntryExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 创建新记录
	entry := &models.CaseEthicalWallWhitelist{
		CaseID:    caseID,
		UserID:    userID,
		GrantedBy: grantedBy,
		Reason:    reason,
	}

	return r.db.WithContext(ctx).Create(entry).Error
}

// RemoveFromWhitelist 从案件白名单移除用户
func (r *EthicalWallRepositoryImpl) RemoveFromWhitelist(ctx context.Context, caseID, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("case_id = ? AND user_id = ?", caseID, userID).
		Delete(&models.CaseEthicalWallWhitelist{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrWhitelistEntryNotFound
	}
	return nil
}

// GetWhitelistByCase 获取案件白名单列表
func (r *EthicalWallRepositoryImpl) GetWhitelistByCase(ctx context.Context, caseID uint) ([]*models.CaseEthicalWallWhitelist, error) {
	var entries []*models.CaseEthicalWallWhitelist
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("GrantedByUser").
		Where("case_id = ?", caseID).
		Order("granted_at DESC").
		Find(&entries).Error
	return entries, err
}

// GetWhitelistByUser 获取用户可访问的隔离墙案件
func (r *EthicalWallRepositoryImpl) GetWhitelistByUser(ctx context.Context, userID uint) ([]*models.CaseEthicalWallWhitelist, error) {
	var entries []*models.CaseEthicalWallWhitelist
	err := r.db.WithContext(ctx).
		Preload("Case").
		Preload("Case.Client").
		Where("user_id = ?", userID).
		Order("granted_at DESC").
		Find(&entries).Error
	return entries, err
}

// LogAccessAttempt 记录访问尝试
func (r *EthicalWallRepositoryImpl) LogAccessAttempt(ctx context.Context, caseID, userID uint, accessType, accessResult, ipAddress, userAgent string) error {
	log := &models.EthicalWallAccessLog{
		CaseID:       caseID,
		UserID:       userID,
		AccessType:   accessType,
		AccessResult: accessResult,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		AttemptedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Create(log).Error
}

// GetAccessLogs 获取访问日志
func (r *EthicalWallRepositoryImpl) GetAccessLogs(ctx context.Context, caseID uint, limit int) ([]*models.EthicalWallAccessLog, error) {
	var logs []*models.EthicalWallAccessLog
	query := r.db.WithContext(ctx).
		Preload("User").
		Order("attempted_at DESC")

	if caseID > 0 {
		query = query.Where("case_id = ?", caseID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// ClearWhitelist 清空案件白名单
func (r *EthicalWallRepositoryImpl) ClearWhitelist(ctx context.Context, caseID uint) error {
	return r.db.WithContext(ctx).
		Where("case_id = ?", caseID).
		Delete(&models.CaseEthicalWallWhitelist{}).Error
}
