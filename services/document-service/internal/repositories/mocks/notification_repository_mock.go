package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
)

// NotificationRepository 通知仓库模拟
type NotificationRepository struct {
	notifications map[uint]*models.Notification
	nextID        uint
}

// NewNotificationRepository 创建通知仓库模拟
func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{
		notifications: make(map[uint]*models.Notification),
		nextID:        1,
	}
}

// GetByID 根据ID获取通知
func (r *NotificationRepository) GetByID(ctx context.Context, id uint) (*models.Notification, error) {
	if notification, exists := r.notifications[id]; exists {
		return notification, nil
	}
	return nil, errors.New("notification not found")
}

// Create 创建通知
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	notification.ID = r.nextID
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()
	r.notifications[notification.ID] = notification
	r.nextID++
	return nil
}

// Update 更新通知
func (r *NotificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	if _, exists := r.notifications[notification.ID]; exists {
		notification.UpdatedAt = time.Now()
		r.notifications[notification.ID] = notification
		return nil
	}
	return errors.New("notification not found")
}

// Delete 删除通知
func (r *NotificationRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := r.notifications[id]; exists {
		delete(r.notifications, id)
		return nil
	}
	return errors.New("notification not found")
}

// List 列出通知
func (r *NotificationRepository) List(ctx context.Context, options NotificationListOptions) ([]*models.Notification, int64, error) {
	var result []*models.Notification
	var total int64

	for _, notification := range r.notifications {
		// 应用过滤条件
		if options.RecipientID != 0 && (notification.RecipientID == nil || *notification.RecipientID != options.RecipientID) {
			continue
		}
		if options.TenantID != "" && notification.TenantID != options.TenantID {
			continue
		}
		if options.Type != "" && notification.Type != options.Type {
			continue
		}
		if options.Priority != "" && notification.Priority != options.Priority {
			continue
		}
		if options.IsRead != nil && notification.IsRead != *options.IsRead {
			continue
		}

		total++
		// 应用分页
		if options.Offset > 0 {
			options.Offset--
			continue
		}
		if options.Limit > 0 && len(result) >= options.Limit {
			continue
		}
		result = append(result, notification)
	}

	return result, total, nil
}

// GetUnreadCount 获取未读通知数量
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uint, tenantID string) (int64, error) {
	var count int64
	for _, notification := range r.notifications {
		if notification.RecipientID != nil && *notification.RecipientID == userID &&
			notification.TenantID == tenantID && !notification.IsRead {
			count++
		}
	}
	return count, nil
}

// MarkAllAsRead 标记所有通知为已读
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uint, tenantID string) error {
	for _, notification := range r.notifications {
		if notification.RecipientID != nil && *notification.RecipientID == userID &&
			notification.TenantID == tenantID && !notification.IsRead {
			notification.IsRead = true
			now := time.Now()
			notification.ReadAt = &now
			notification.UpdatedAt = now
		}
	}
	return nil
}

// NotificationListOptions 通知列表选项
type NotificationListOptions struct {
	RecipientID uint
	Type        string
	IsRead      *bool
	Priority    string
	TenantID    string
	Limit       int
	Offset      int
	SortBy      string
	SortOrder   string
}