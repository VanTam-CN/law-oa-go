package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// notificationService 通知服务实现
type notificationService struct {
	notificationRepo repositories.NotificationRepository
	userRepo         repositories.UserRepository
	auditRepo        repositories.DocumentAuditRepository
	logger           *logrus.Logger
}

// NewNotificationService 创建新的通知服务
func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	userRepo repositories.UserRepository,
	auditRepo repositories.DocumentAuditRepository,
	logger *logrus.Logger,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		auditRepo:        auditRepo,
		logger:           logger,
	}
}

// CreateNotification 创建通知
func (s *notificationService) CreateNotification(ctx context.Context, req *CreateNotificationRequest) (*NotificationResponse, error) {
	// 验证请求
	if err := s.validateCreateNotificationRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 验证接收者存在
	if req.RecipientID != "" {
		recipientID, err := s.parseUserID(req.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient ID: %w", err)
		}

		_, err = s.userRepo.GetByID(ctx, recipientID)
		if err != nil {
			return nil, fmt.Errorf("recipient not found: %w", err)
		}
	}

	// 创建通知
	notification := &models.Notification{
		Type:         req.Type,
		Title:        req.Title,
		Message:      req.Message,
		RecipientID:  s.parseUserIDPtr(req.RecipientID),
		SenderID:     s.parseUserIDPtr(req.SenderID),
		RelatedType:  req.RelatedType,
		RelatedID:    s.parseRelatedIDPtr(req.RelatedID),
		Priority:     req.Priority,
		IsRead:       false,
		Data:         req.Data,
		TenantID:     req.TenantID,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.SenderID,
		Action:     "create_notification",
		Details:    fmt.Sprintf("Created %s notification: %s", req.Type, req.Title),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return s.convertToNotificationResponse(notification), nil
}

// GetNotification 获取通知详情
func (s *notificationService) GetNotification(ctx context.Context, notificationID string) (*NotificationResponse, error) {
	id, err := s.parseNotificationID(notificationID)
	if err != nil {
		return nil, fmt.Errorf("invalid notification ID: %w", err)
	}

	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return s.convertToNotificationResponse(notification), nil
}

// MarkAsRead 标记通知为已读
func (s *notificationService) MarkAsRead(ctx context.Context, req *MarkAsReadRequest) error {
	notificationID, err := s.parseNotificationID(req.NotificationID)
	if err != nil {
		return fmt.Errorf("invalid notification ID: %w", err)
	}

	// 获取通知
	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	// 更新已读状态
	notification.IsRead = true
	notification.ReadAt = &time.Time{}
	*notification.ReadAt = time.Now()

	if err := s.notificationRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "mark_notification_read",
		Details:    fmt.Sprintf("Marked notification as read: %s", notification.Title),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// MarkAllAsRead 标记所有通知为已读
func (s *notificationService) MarkAllAsRead(ctx context.Context, req *MarkAllAsReadRequest) error {
	// 解析用户ID
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// 标记所有通知为已读
	if err := s.notificationRepo.MarkAllAsRead(ctx, userID, req.TenantID); err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "mark_all_notifications_read",
		Details:    "Marked all notifications as read",
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// DeleteNotification 删除通知
func (s *notificationService) DeleteNotification(ctx context.Context, req *DeleteNotificationRequest) error {
	notificationID, err := s.parseNotificationID(req.NotificationID)
	if err != nil {
		return fmt.Errorf("invalid notification ID: %w", err)
	}

	// 获取通知用于审计
	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	// 删除通知
	if err := s.notificationRepo.Delete(ctx, notificationID); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "delete_notification",
		Details:    fmt.Sprintf("Deleted notification: %s", notification.Title),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// ListNotifications 列出通知
func (s *notificationService) ListNotifications(ctx context.Context, filter *NotificationFilter) (*NotificationListResponse, error) {
	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 计算偏移量
	offset := (filter.Page - 1) * filter.PageSize

	// 构建查询选项
	options := repositories.NotificationListOptions{
		RecipientID: 0,
		Type:        filter.Type,
		IsRead:      filter.IsRead,
		Priority:    filter.Priority,
		TenantID:    filter.TenantID,
		Limit:       filter.PageSize,
		Offset:      offset,
		SortBy:      filter.SortBy,
		SortOrder:   filter.SortOrder,
	}

	// 解析接收者ID
	if filter.RecipientID != "" {
		recipientID, err := s.parseUserID(filter.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient ID filter: %w", err)
		}
		options.RecipientID = recipientID
	}

	// 获取通知列表
	notifications, total, err := s.notificationRepo.List(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	// 转换为响应格式
	responses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = s.convertToNotificationResponse(notification)
	}

	return &NotificationListResponse{
		Notifications: responses,
		Total:         total,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
	}, nil
}

// GetUnreadCount 获取未读通知数量
func (s *notificationService) GetUnreadCount(ctx context.Context, userID, tenantID string) (int64, error) {
	userIDInt, err := s.parseUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %w", err)
	}

	count, err := s.notificationRepo.GetUnreadCount(ctx, userIDInt, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// SendNotification 发送通知
func (s *notificationService) SendNotification(ctx context.Context, req *SendNotificationRequest) error {
	// 验证请求
	if err := s.validateSendNotificationRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 验证发送者存在
	senderID, err := s.parseUserID(req.SenderID)
	if err != nil {
		return fmt.Errorf("invalid sender ID: %w", err)
	}

	sender, err := s.userRepo.GetByID(ctx, senderID)
	if err != nil {
		return fmt.Errorf("sender not found: %w", err)
	}

	// 验证接收者存在
	recipientID, err := s.parseUserID(req.RecipientID)
	if err != nil {
		return fmt.Errorf("invalid recipient ID: %w", err)
	}

	recipient, err := s.userRepo.GetByID(ctx, recipientID)
	if err != nil {
		return fmt.Errorf("recipient not found: %w", err)
	}

	// 创建通知
	notification := &models.Notification{
		Type:         req.Type,
		Title:        req.Title,
		Message:      req.Message,
		RecipientID:  &recipientID,
		SenderID:     &senderID,
		RelatedType:  req.RelatedType,
		RelatedID:    s.parseRelatedIDPtr(req.RelatedID),
		Priority:     req.Priority,
		IsRead:       false,
		Data:         req.Data,
		TenantID:     req.TenantID,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"sender_id":     senderID,
		"recipient_id":  recipientID,
		"notification_type": req.Type,
		"title":         req.Title,
	}).Info("Notification sent")

	return nil
}

// BroadcastNotification 广播通知
func (s *notificationService) BroadcastNotification(ctx context.Context, req *BroadcastNotificationRequest) error {
	// 验证请求
	if err := s.validateBroadcastNotificationRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 获取租户下的所有用户
	users, err := s.userRepo.GetActiveUsers(ctx, req.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get active users: %w", err)
	}

	successCount := 0
	failureCount := 0

	// 为每个用户创建通知
	for _, user := range users {
		// 跳过发送者自己（如果指定）
		if req.SenderID != "" && fmt.Sprintf("%d", user.ID) == req.SenderID {
			continue
		}

		notification := &models.Notification{
			Type:         req.Type,
			Title:        req.Title,
			Message:      req.Message,
			RecipientID:  &user.ID,
			SenderID:     s.parseUserIDPtr(req.SenderID),
			RelatedType:  req.RelatedType,
			RelatedID:    s.parseRelatedIDPtr(req.RelatedID),
			Priority:     req.Priority,
			IsRead:       false,
			Data:         req.Data,
			TenantID:     req.TenantID,
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			failureCount++
			s.logger.WithError(err).WithField("user_id", user.ID).Error("Failed to create broadcast notification")
		} else {
			successCount++
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"tenant_id":      req.TenantID,
		"success_count":  successCount,
		"failure_count":  failureCount,
		"total_users":    len(users),
		"notification_type": req.Type,
		"title":          req.Title,
	}).Info("Broadcast notification completed")

	return nil
}

// GetNotificationSettings 获取通知设置
func (s *notificationService) GetNotificationSettings(ctx context.Context, userID, tenantID string) (*NotificationSettings, error) {
	userIDInt, err := s.parseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// 这里应该从用户设置表获取设置，现在返回默认设置
	settings := &NotificationSettings{
		UserID:        userIDInt,
		TenantID:      tenantID,
		EmailEnabled:  true,
		PushEnabled:   true,
		InAppEnabled:  true,
		Types: map[string]bool{
			"document_shared":   true,
			"permission_changed": true,
			"document_updated":  true,
			"system_announcement": true,
		},
		DailyDigest:    false,
		QuietHours:     &QuietHours{
			Enabled: false,
			Start:   "22:00",
			End:     "08:00",
		},
	}

	return settings, nil
}

// UpdateNotificationSettings 更新通知设置
func (s *notificationService) UpdateNotificationSettings(ctx context.Context, req *UpdateNotificationSettingsRequest) error {
	// 验证请求
	if err := s.validateUpdateNotificationSettingsRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// 这里应该更新用户设置表，现在只记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "update_notification_settings",
		Details:    "Updated notification settings",
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"tenant_id": req.TenantID,
	}).Info("Notification settings updated")

	return nil
}

// GetNotificationStats 获取通知统计
func (s *notificationService) GetNotificationStats(ctx context.Context, tenantID string) (*NotificationStats, error) {
	stats := &NotificationStats{
		TotalNotifications: 0,
		UnreadNotifications: 0,
		NotificationsByType: make(map[string]int64),
		NotificationsByDay:  make(map[string]int64),
	}

	// 获取租户下的所有通知用于统计
	notifications, _, err := s.notificationRepo.List(ctx, repositories.NotificationListOptions{
		TenantID: tenantID,
		Limit:    10000,
		Offset:   0,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get notifications for stats: %w", err)
	}

	stats.TotalNotifications = int64(len(notifications))

	for _, notification := range notifications {
		// 统计未读通知
		if !notification.IsRead {
			stats.UnreadNotifications++
		}

		// 按类型统计
		stats.NotificationsByType[notification.Type]++

		// 按日期统计
		dateKey := notification.CreatedAt.Format("2006-01-02")
		stats.NotificationsByDay[dateKey]++
	}

	return stats, nil
}

// 辅助方法

// validateCreateNotificationRequest 验证创建通知请求
func (s *notificationService) validateCreateNotificationRequest(req *CreateNotificationRequest) error {
	if req.Type == "" {
		return fmt.Errorf("notification type is required")
	}
	if req.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if req.Message == "" {
		return fmt.Errorf("notification message is required")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return nil
}

// validateSendNotificationRequest 验证发送通知请求
func (s *notificationService) validateSendNotificationRequest(req *SendNotificationRequest) error {
	if req.SenderID == "" {
		return fmt.Errorf("sender_id is required")
	}
	if req.RecipientID == "" {
		return fmt.Errorf("recipient_id is required")
	}
	if req.Type == "" {
		return fmt.Errorf("notification type is required")
	}
	if req.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if req.Message == "" {
		return fmt.Errorf("notification message is required")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return nil
}

// validateBroadcastNotificationRequest 验证广播通知请求
func (s *notificationService) validateBroadcastNotificationRequest(req *BroadcastNotificationRequest) error {
	if req.Type == "" {
		return fmt.Errorf("notification type is required")
	}
	if req.Title == "" {
		return fmt.Errorf("notification title is required")
	}
	if req.Message == "" {
		return fmt.Errorf("notification message is required")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return nil
}

// validateUpdateNotificationSettingsRequest 验证更新通知设置请求
func (s *notificationService) validateUpdateNotificationSettingsRequest(req *UpdateNotificationSettingsRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return nil
}

// parseUserID 解析用户ID
func (s *notificationService) parseUserID(userID string) (uint, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %s", userID)
	}
	return uint(id), nil
}

// parseUserIDPtr 解析用户ID指针
func (s *notificationService) parseUserIDPtr(userID string) *uint {
	if userID == "" {
		return nil
	}
	id, err := s.parseUserID(userID)
	if err != nil {
		return nil
	}
	return &id
}

// parseNotificationID 解析通知ID
func (s *notificationService) parseNotificationID(notificationID string) (uint, error) {
	id, err := strconv.ParseUint(notificationID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid notification ID format: %s", notificationID)
	}
	return uint(id), nil
}

// parseRelatedIDPtr 解析相关ID指针
func (s *notificationService) parseRelatedIDPtr(relatedID string) *uint {
	if relatedID == "" {
		return nil
	}
	id, err := strconv.ParseUint(relatedID, 10, 32)
	if err != nil {
		return nil
	}
	result := uint(id)
	return &result
}

// convertToNotificationResponse 转换为通知响应格式
func (s *notificationService) convertToNotificationResponse(notification *models.Notification) *NotificationResponse {
	response := &NotificationResponse{
		ID:          notification.ID,
		Type:        notification.Type,
		Title:       notification.Title,
		Message:     notification.Message,
		Priority:    notification.Priority,
		IsRead:      notification.IsRead,
		ReadAt:      notification.ReadAt,
		Data:        notification.Data,
		TenantID:    notification.TenantID,
		CreatedAt:   notification.CreatedAt,
		UpdatedAt:   notification.UpdatedAt,
	}

	// 转换发送者ID为字符串
	if notification.SenderID != nil {
		response.SenderID = fmt.Sprintf("%d", *notification.SenderID)
		// 尝试获取发送者信息
		if sender, err := s.userRepo.GetByID(context.Background(), *notification.SenderID); err == nil {
			response.SenderName = sender.Username
		}
	}

	// 转换接收者ID为字符串
	if notification.RecipientID != nil {
		response.RecipientID = fmt.Sprintf("%d", *notification.RecipientID)
		// 尝试获取接收者信息
		if recipient, err := s.userRepo.GetByID(context.Background(), *notification.RecipientID); err == nil {
			response.RecipientName = recipient.Username
		}
	}

	// 转换相关ID为字符串
	if notification.RelatedID != nil {
		response.RelatedID = fmt.Sprintf("%d", *notification.RelatedID)
	}
	response.RelatedType = notification.RelatedType

	return response
}

// logAudit 记录审计日志
func (s *notificationService) logAudit(ctx context.Context, req *LogActionRequest) error {
	// 解析用户ID
	var userID uint
	if req.UserID != "" {
		id, err := s.parseUserID(req.UserID)
		if err != nil {
			return err
		}
		userID = id
	}

	audit := &models.DocumentAudit{
		UserID:    userID,
		TenantID:  req.TenantID,
		Action:    req.Action,
		Details:   req.Details,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}

	return s.auditRepo.Create(ctx, audit)
}