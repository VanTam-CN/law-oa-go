package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

const (
	// 文档锁的 Redis key 前缀
	documentLockKeyPrefix = "doc_lock:"

	// 默认锁过期时间
	defaultLockExpiration = 30 * time.Minute

	// 签出锁过期时间（更长，用于离线编辑）
	checkoutLockExpiration = 24 * time.Hour

	// 锁续期时间（在即将过期时自动续期）
	lockRenewalTime = 5 * time.Minute
)

// DocumentLockService 文档锁服务
type DocumentLockService struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewDocumentLockService 创建文档锁服务
func NewDocumentLockService(db *gorm.DB, redisClient *redis.Client) *DocumentLockService {
	return &DocumentLockService{
		db:    db,
		redis: redisClient,
	}
}

// LockStatus 锁状态
type LockStatus struct {
	DocumentID    uint      `json:"document_id"`
	DocumentName  string    `json:"document_name,omitempty"`
	IsLocked      bool      `json:"is_locked"`
	LockedBy      uint      `json:"locked_by,omitempty"`
	LockedByName  string    `json:"locked_by_name,omitempty"`
	LockedAt      time.Time `json:"locked_at,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	IsCheckedOut  bool      `json:"is_checked_out,omitempty"`
	CanEdit       bool      `json:"can_edit"`
	Reason        string    `json:"reason,omitempty"`
}

// AcquireLockRequest 获取锁请求
type AcquireLockRequest struct {
	DocumentID uint   `json:"document_id" binding:"required"`
	UserID     uint   `json:"user_id" binding:"required"`
	UserName   string `json:"user_name"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	IsCheckout bool   `json:"is_checkout"` // 是否为签出模式（离线编辑）
}

// ReleaseLockRequest 释放锁请求
type ReleaseLockRequest struct {
	DocumentID uint `json:"document_id" binding:"required"`
	UserID     uint `json:"user_id" binding:"required"`
	Force      bool `json:"force"` // 管理员强制解锁
}

// RenewLockRequest 续期锁请求
type RenewLockRequest struct {
	DocumentID uint `json:"document_id" binding:"required"`
	UserID     uint `json:"user_id" binding:"required"`
}

// AcquireLock 获取文档锁
func (s *DocumentLockService) AcquireLock(ctx context.Context, req *AcquireLockRequest) (*LockStatus, error) {
	// 首先检查文档是否存在
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, req.DocumentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to check document: %w", err)
	}

	// 生成 Redis 锁 key
	lockKey := s.getLockKey(req.DocumentID)

	// 尝试获取现有锁
	existingLock, err := s.getLockFromRedis(ctx, lockKey)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check existing lock: %w", err)
	}

	// 检查是否已被锁定
	if existingLock != nil {
		// 检查是否是同一用户
		if existingLock.LockedBy == req.UserID {
			// 同一用户，续期锁
			renewReq := &RenewLockRequest{
				DocumentID: req.DocumentID,
				UserID:     req.UserID,
			}
			return s.RenewLock(ctx, renewReq)
		}

		// 检查锁是否已过期
		if time.Now().Before(existingLock.ExpiresAt) {
			// 锁仍然有效，返回锁状态
			return &LockStatus{
				DocumentID:    req.DocumentID,
				IsLocked:      true,
				LockedBy:      existingLock.LockedBy,
				LockedByName:  s.getUserName(ctx, existingLock.LockedBy),
				LockedAt:      existingLock.LockedAt,
				ExpiresAt:     existingLock.ExpiresAt,
				IsCheckedOut:  existingLock.IsCheckedOut,
				CanEdit:       false,
				Reason:        "Document is locked by another user",
			}, nil
		}

		// 锁已过期，删除旧锁
		s.releaseLockFromRedis(ctx, lockKey)
	}

	// 确定锁过期时间
	expiration := defaultLockExpiration
	if req.IsCheckout {
		expiration = checkoutLockExpiration
	}

	// 创建新锁
	now := time.Now()
	lockData := &models.DocumentLock{
		DocumentID:   req.DocumentID,
		LockedBy:     req.UserID,
		LockedAt:     now,
		ExpiresAt:    now.Add(expiration),
		IsCheckedOut: req.IsCheckout,
	}

	if req.IsCheckout {
		lockData.CheckedOutAt = &now
		lockData.CheckoutIP = req.IPAddress
	}

	lockData.LastActivity = &now

	// 保存到数据库
	if err := s.db.WithContext(ctx).Save(lockData).Error; err != nil {
		return nil, fmt.Errorf("failed to save lock to database: %w", err)
	}

	// 设置 Redis 锁
	if err := s.setLockInRedis(ctx, lockKey, lockData, expiration); err != nil {
		// 回滚数据库
		s.db.WithContext(ctx).Delete(lockData)
		return nil, fmt.Errorf("failed to set lock in redis: %w", err)
	}

	return &LockStatus{
		DocumentID:    req.DocumentID,
		IsLocked:      true,
		LockedBy:      req.UserID,
		LockedByName:  req.UserName,
		LockedAt:      now,
		ExpiresAt:     now.Add(expiration),
		IsCheckedOut:  req.IsCheckout,
		CanEdit:       true,
		Reason:        "Lock acquired successfully",
	}, nil
}

// ReleaseLock 释放文档锁
func (s *DocumentLockService) ReleaseLock(ctx context.Context, req *ReleaseLockRequest) error {
	lockKey := s.getLockKey(req.DocumentID)

	// 获取现有锁
	existingLock, err := s.getLockFromRedis(ctx, lockKey)
	if err != nil {
		if err == redis.Nil {
			// 锁不存在，可能已过期
			return s.releaseLockFromDB(ctx, req.DocumentID, req.UserID)
		}
		return fmt.Errorf("failed to check existing lock: %w", err)
	}

	// 检查权限
	if !req.Force && existingLock.LockedBy != req.UserID {
		return fmt.Errorf("permission denied: lock is held by another user")
	}

	// 从 Redis 释放锁
	if err := s.releaseLockFromRedis(ctx, lockKey); err != nil {
		return fmt.Errorf("failed to release lock from redis: %w", err)
	}

	// 从数据库删除锁
	return s.releaseLockFromDB(ctx, req.DocumentID, req.UserID)
}

// RenewLock 续期文档锁
func (s *DocumentLockService) RenewLock(ctx context.Context, req *RenewLockRequest) (*LockStatus, error) {
	lockKey := s.getLockKey(req.DocumentID)

	// 获取现有锁
	existingLock, err := s.getLockFromRedis(ctx, lockKey)
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("lock not found or expired")
		}
		return nil, fmt.Errorf("failed to check existing lock: %w", err)
	}

	// 检查权限
	if existingLock.LockedBy != req.UserID {
		return nil, fmt.Errorf("permission denied: lock is held by another user")
	}

	// 更新锁时间
	now := time.Now()
	expiration := defaultLockExpiration
	if existingLock.IsCheckedOut {
		expiration = checkoutLockExpiration
	}

	newExpiresAt := now.Add(expiration)
	existingLock.ExpiresAt = newExpiresAt
	existingLock.LastActivity = &now

	// 更新数据库
	if err := s.db.WithContext(ctx).Save(existingLock).Error; err != nil {
		return nil, fmt.Errorf("failed to update lock in database: %w", err)
	}

	// 更新 Redis 锁
	if err := s.setLockInRedis(ctx, lockKey, existingLock, expiration); err != nil {
		return nil, fmt.Errorf("failed to update lock in redis: %w", err)
	}

	// 获取用户名
	userName := s.getUserName(ctx, req.UserID)

	return &LockStatus{
		DocumentID:    req.DocumentID,
		IsLocked:      true,
		LockedBy:      req.UserID,
		LockedByName:  userName,
		LockedAt:      existingLock.LockedAt,
		ExpiresAt:     newExpiresAt,
		IsCheckedOut:  existingLock.IsCheckedOut,
		CanEdit:       true,
		Reason:        "Lock renewed successfully",
	}, nil
}

// GetLockStatus 获取文档锁状态
func (s *DocumentLockService) GetLockStatus(ctx context.Context, documentID, userID uint) (*LockStatus, error) {
	lockKey := s.getLockKey(documentID)

	// 尝试从 Redis 获取
	existingLock, err := s.getLockFromRedis(ctx, lockKey)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check lock status: %w", err)
	}

	// 如果 Redis 中没有锁，检查数据库
	if existingLock == nil {
		var dbLock models.DocumentLock
		if err := s.db.WithContext(ctx).Where("document_id = ?", documentID).First(&dbLock).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 没有锁
				return &LockStatus{
					DocumentID: documentID,
					IsLocked:   false,
					CanEdit:    true,
				}, nil
			}
			return nil, fmt.Errorf("failed to check lock in database: %w", err)
		}

		// 检查数据库中的锁是否过期
		if time.Now().After(dbLock.ExpiresAt) {
			// 锁已过期，删除
			s.db.WithContext(ctx).Delete(&dbLock)
			return &LockStatus{
				DocumentID: documentID,
				IsLocked:   false,
				CanEdit:    true,
			}, nil
		}

		existingLock = &dbLock
	}

	// 获取用户名
	userName := s.getUserName(ctx, existingLock.LockedBy)

	// 返回锁状态
	canEdit := existingLock.LockedBy == userID
	reason := "Document is locked"
	if canEdit {
		reason = "You hold the lock for this document"
	}

	return &LockStatus{
		DocumentID:    documentID,
		IsLocked:      true,
		LockedBy:      existingLock.LockedBy,
		LockedByName:  userName,
		LockedAt:      existingLock.LockedAt,
		ExpiresAt:     existingLock.ExpiresAt,
		IsCheckedOut:  existingLock.IsCheckedOut,
		CanEdit:       canEdit,
		Reason:        reason,
	}, nil
}

// GetUserLocks 获取用户持有的所有锁
func (s *DocumentLockService) GetUserLocks(ctx context.Context, userID uint) ([]*LockStatus, error) {
	var dbLocks []models.DocumentLock
	if err := s.db.WithContext(ctx).
		Where("locked_by = ? AND expires_at > ?", userID, time.Now()).
		Find(&dbLocks).Error; err != nil {
		return nil, fmt.Errorf("failed to get user locks: %w", err)
	}

	statuses := make([]*LockStatus, 0, len(dbLocks))
	for _, lock := range dbLocks {
		// 获取用户名
		userName := s.getUserName(ctx, lock.LockedBy)

		// 获取文档信息
		var doc models.Document
		docName := fmt.Sprintf("Document %d", lock.DocumentID)
		if err := s.db.WithContext(ctx).Select("name").First(&doc, lock.DocumentID).Error; err == nil {
			docName = doc.Name
		}

		statuses = append(statuses, &LockStatus{
			DocumentID:    lock.DocumentID,
			DocumentName:  docName,
			IsLocked:      true,
			LockedBy:      lock.LockedBy,
			LockedByName:  userName,
			LockedAt:      lock.LockedAt,
			ExpiresAt:     lock.ExpiresAt,
			IsCheckedOut:  lock.IsCheckedOut,
			CanEdit:       true,
		})
	}

	return statuses, nil
}

// CleanupExpiredLocks 清理过期的锁
func (s *DocumentLockService) CleanupExpiredLocks(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.DocumentLock{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup expired locks: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// ValidateEditPermission 验证编辑权限（在编辑前调用）
func (s *DocumentLockService) ValidateEditPermission(ctx context.Context, documentID, userID uint) (bool, string, error) {
	status, err := s.GetLockStatus(ctx, documentID, userID)
	if err != nil {
		return false, "", err
	}

	if !status.IsLocked {
		return true, "Document is not locked", nil
	}

	if status.CanEdit {
		return true, "You hold the lock for this document", nil
	}

	return false, fmt.Sprintf("Document is locked by %s since %s",
		status.LockedByName, status.LockedAt.Format("2006-01-02 15:04")), nil
}

// Helper methods

// getLockKey 生成 Redis 锁 key
func (s *DocumentLockService) getLockKey(documentID uint) string {
	return fmt.Sprintf("%s%d", documentLockKeyPrefix, documentID)
}

// getLockFromRedis 从 Redis 获取锁
func (s *DocumentLockService) getLockFromRedis(ctx context.Context, key string) (*models.DocumentLock, error) {
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 解析数据（这里简化处理，实际应该序列化/反序列化）
	lock := &models.DocumentLock{}
	// TODO: 实现 JSON 反序列化
	_ = data // 暂时忽略数据，避免 unused 警告
	return lock, nil
}

// setLockInRedis 在 Redis 中设置锁
func (s *DocumentLockService) setLockInRedis(ctx context.Context, key string, lock *models.DocumentLock, expiration time.Duration) error {
	// TODO: 实现 JSON 序列化
	return s.redis.Set(ctx, key, "locked", expiration).Err()
}

// releaseLockFromRedis 从 Redis 释放锁
func (s *DocumentLockService) releaseLockFromRedis(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}

// releaseLockFromDB 从数据库释放锁
func (s *DocumentLockService) releaseLockFromDB(ctx context.Context, documentID, userID uint) error {
	return s.db.WithContext(ctx).
		Where("document_id = ? AND locked_by = ?", documentID, userID).
		Delete(&models.DocumentLock{}).Error
}

// getUserName 获取用户名
func (s *DocumentLockService) getUserName(ctx context.Context, userID uint) string {
	var user models.User
	if err := s.db.WithContext(ctx).Select("name").First(&user, userID).Error; err == nil {
		return user.Name
	}
	return fmt.Sprintf("User %d", userID)
}
