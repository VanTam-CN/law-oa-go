package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"law-oa-go/internal/models"
)

const (
	// 默认锁过期时间
	defaultLockExpiration = 30 * time.Minute

	// 签出锁过期时间（更长，用于离线编辑）
	checkoutLockExpiration = 24 * time.Hour

	// 锁续期时间（在即将过期时自动续期）
	lockRenewalTime = 5 * time.Minute
)

// DocumentLockService 文档锁服务
type DocumentLockService struct {
	db *gorm.DB
}

// NewDocumentLockService 创建文档锁服务。Redis 参数仅为兼容保留，不再参与锁语义。
func NewDocumentLockService(db *gorm.DB, _ interface{}) *DocumentLockService {
	return &DocumentLockService{
		db: db,
	}
}

// LockStatus 锁状态
type LockStatus struct {
	DocumentID   uint      `json:"document_id"`
	DocumentName string    `json:"document_name,omitempty"`
	IsLocked     bool      `json:"is_locked"`
	LockedBy     uint      `json:"locked_by,omitempty"`
	LockedByName string    `json:"locked_by_name,omitempty"`
	LockedAt     time.Time `json:"locked_at,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	IsCheckedOut bool      `json:"is_checked_out,omitempty"`
	CanEdit      bool      `json:"can_edit"`
	Reason       string    `json:"reason,omitempty"`
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
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "document_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"locked_by", "locked_at", "expires_at", "is_checked_out",
			"checked_out_at", "checkout_ip", "last_activity",
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			gorm.Expr("document_locks.expires_at < ? OR document_locks.locked_by = ?", now, req.UserID),
		}},
	}).Create(lockData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to save lock to database: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		existingLock, err := s.getActiveLock(ctx, req.DocumentID, now)
		if err != nil {
			return nil, err
		}
		if existingLock == nil {
			return nil, fmt.Errorf("failed to acquire document lock")
		}
		return &LockStatus{
			DocumentID:   req.DocumentID,
			IsLocked:     true,
			LockedBy:     existingLock.LockedBy,
			LockedByName: s.getUserName(ctx, existingLock.LockedBy),
			LockedAt:     existingLock.LockedAt,
			ExpiresAt:    existingLock.ExpiresAt,
			IsCheckedOut: existingLock.IsCheckedOut,
			CanEdit:      false,
			Reason:       "Document is locked by another user",
		}, nil
	}

	return &LockStatus{
		DocumentID:   req.DocumentID,
		IsLocked:     true,
		LockedBy:     req.UserID,
		LockedByName: req.UserName,
		LockedAt:     now,
		ExpiresAt:    now.Add(expiration),
		IsCheckedOut: req.IsCheckout,
		CanEdit:      true,
		Reason:       "Lock acquired successfully",
	}, nil
}

// ReleaseLock 释放文档锁
func (s *DocumentLockService) ReleaseLock(ctx context.Context, req *ReleaseLockRequest) error {
	query := s.db.WithContext(ctx).Model(&models.DocumentLock{}).
		Where("document_id = ?", req.DocumentID)
	if !req.Force {
		query = query.Where("locked_by = ?", req.UserID)
	}
	result := query.Delete(&models.DocumentLock{})
	if result.Error != nil {
		return fmt.Errorf("failed to release document lock: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission denied: lock is held by another user")
	}
	return nil
}

// RenewLock 续期文档锁
func (s *DocumentLockService) RenewLock(ctx context.Context, req *RenewLockRequest) (*LockStatus, error) {
	existingLock, err := s.getActiveLock(ctx, req.DocumentID, time.Now())
	if err != nil {
		return nil, err
	}
	if existingLock == nil {
		return nil, fmt.Errorf("lock not found or expired")
	}
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

	result := s.db.WithContext(ctx).Model(&models.DocumentLock{}).
		Where("document_id = ? AND locked_by = ?", req.DocumentID, req.UserID).
		Updates(map[string]interface{}{
			"expires_at":    newExpiresAt,
			"last_activity": now,
		})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update lock in database: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("lock not found or expired")
	}

	// 获取用户名
	userName := s.getUserName(ctx, req.UserID)

	return &LockStatus{
		DocumentID:   req.DocumentID,
		IsLocked:     true,
		LockedBy:     req.UserID,
		LockedByName: userName,
		LockedAt:     existingLock.LockedAt,
		ExpiresAt:    newExpiresAt,
		IsCheckedOut: existingLock.IsCheckedOut,
		CanEdit:      true,
		Reason:       "Lock renewed successfully",
	}, nil
}

// GetLockStatus 获取文档锁状态
func (s *DocumentLockService) GetLockStatus(ctx context.Context, documentID, userID uint) (*LockStatus, error) {
	existingLock, err := s.getActiveLock(ctx, documentID, time.Now())
	if err != nil {
		return nil, err
	}
	if existingLock == nil {
		return &LockStatus{
			DocumentID: documentID,
			IsLocked:   false,
			CanEdit:    true,
		}, nil
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
		DocumentID:   documentID,
		IsLocked:     true,
		LockedBy:     existingLock.LockedBy,
		LockedByName: userName,
		LockedAt:     existingLock.LockedAt,
		ExpiresAt:    existingLock.ExpiresAt,
		IsCheckedOut: existingLock.IsCheckedOut,
		CanEdit:      canEdit,
		Reason:       reason,
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
			DocumentID:   lock.DocumentID,
			DocumentName: docName,
			IsLocked:     true,
			LockedBy:     lock.LockedBy,
			LockedByName: userName,
			LockedAt:     lock.LockedAt,
			ExpiresAt:    lock.ExpiresAt,
			IsCheckedOut: lock.IsCheckedOut,
			CanEdit:      true,
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

// getActiveLock reads a non-expired lock and opportunistically removes expired rows.
func (s *DocumentLockService) getActiveLock(ctx context.Context, documentID uint, now time.Time) (*models.DocumentLock, error) {
	var lock models.DocumentLock
	err := s.db.WithContext(ctx).Where("document_id = ?", documentID).First(&lock).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check document lock: %w", err)
	}
	if !now.Before(lock.ExpiresAt) {
		if err := s.db.WithContext(ctx).Delete(&lock).Error; err != nil {
			return nil, fmt.Errorf("failed to cleanup expired document lock: %w", err)
		}
		return nil, nil
	}
	return &lock, nil
}

// getUserName 获取用户名
func (s *DocumentLockService) getUserName(ctx context.Context, userID uint) string {
	var user models.User
	if err := s.db.WithContext(ctx).Select("name").First(&user, userID).Error; err == nil {
		return user.Name
	}
	return fmt.Sprintf("User %d", userID)
}
