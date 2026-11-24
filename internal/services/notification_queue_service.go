package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// NotificationQueueService 通知队列服务
type NotificationQueueService struct {
	queueRepo    repositories.NotificationQueueRepository
	templateRepo repositories.NotificationTemplateRepository
	userRepo     repositories.UserRepository
	clientRepo   repositories.ClientRepository
	caseRepo     repositories.CaseRepository
}

// NewNotificationQueueService 创建通知队列服务实例
func NewNotificationQueueService(
	queueRepo repositories.NotificationQueueRepository,
	templateRepo repositories.NotificationTemplateRepository,
	userRepo repositories.UserRepository,
	clientRepo repositories.ClientRepository,
	caseRepo repositories.CaseRepository,
) *NotificationQueueService {
	return &NotificationQueueService{
		queueRepo:    queueRepo,
		templateRepo: templateRepo,
		userRepo:     userRepo,
		clientRepo:   clientRepo,
		caseRepo:     caseRepo,
	}
}

// CreateNotificationRequest 创建通知请求
type CreateNotificationRequest struct {
	TriggerType   string                 `json:"trigger_type" binding:"required"`
	TriggerID     uint                   `json:"trigger_id" binding:"required"`
	CaseID        *uint                  `json:"case_id,omitempty"`
	RecipientType string                 `json:"recipient_type" binding:"required,oneof=client lawyer admin"`
	RecipientID   uint                   `json:"recipient_id" binding:"required"`
	Channel       string                 `json:"channel" binding:"required,oneof=email sms wechat"`
	Subject       string                 `json:"subject" binding:"max=200"`
	Content       string                 `json:"content" binding:"required"`
	TemplateID    *string                `json:"template_id,omitempty"`
	Priority      string                 `json:"priority" binding:"omitempty,oneof=urgent normal low"`
	AutoSend      bool                   `json:"auto_send"`
	VariableData  map[string]interface{} `json:"variable_data,omitempty"`
}

// NotificationQueueResponse 通知队列响应
type NotificationQueueResponse struct {
	ID                    uint    `json:"id"`
	TriggerType           string  `json:"trigger_type"`
	TriggerID             uint    `json:"trigger_id"`
	CaseID                *uint   `json:"case_id,omitempty"`
	RecipientType         string  `json:"recipient_type"`
	RecipientID           uint    `json:"recipient_id"`
	RecipientName         string  `json:"recipient_name"`
	RecipientContact      string  `json:"recipient_contact,omitempty"`
	Channel               string  `json:"channel"`
	Subject               string  `json:"subject,omitempty"`
	Content               string  `json:"content"`
	TemplateID            *string `json:"template_id,omitempty"`
	Status                string  `json:"status"`
	Priority              string  `json:"priority"`
	ContainsSensitiveInfo bool    `json:"contains_sensitive_info"`
	AutoSend              bool    `json:"auto_send"`
	CreatedBy             uint    `json:"created_by"`
	CreatedAt             string  `json:"created_at"`
	ApprovedBy            *uint   `json:"approved_by,omitempty"`
	ApprovedAt            *string `json:"approved_at,omitempty"`
	SentAt                *string `json:"sent_at,omitempty"`
	SentRetryCount        int     `json:"sent_retry_count"`
	ErrorMessage          string  `json:"error_message,omitempty"`
	// 关联数据
	Case     *NotificationCaseInfo `json:"case,omitempty"`
	Template *TemplateInfo         `json:"template,omitempty"`
}

// NotificationCaseInfo 通知用案件信息
type NotificationCaseInfo struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

// TemplateInfo 模板信息
type TemplateInfo struct {
	ID           uint   `json:"id"`
	TemplateCode string `json:"template_code"`
	TemplateName string `json:"template_name"`
	Channel      string `json:"channel"`
}

// ListNotificationsRequest 通知列表请求
type ListNotificationsRequest struct {
	Page          int    `json:"page" form:"page" binding:"min=1"`
	PageSize      int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status        string `json:"status" form:"status" binding:"omitempty,oneof=pending approved sent cancelled failed"`
	Channel       string `json:"channel" form:"channel" binding:"omitempty,oneof=email sms wechat"`
	RecipientType string `json:"recipient_type" form:"recipient_type" binding:"omitempty,oneof=client lawyer admin"`
	TriggerType   string `json:"trigger_type" form:"trigger_type"`
	CaseID        uint   `json:"case_id" form:"case_id"`
	DateFrom      string `json:"date_from" form:"date_from"`
	DateTo        string `json:"date_to" form:"date_to"`
}

// ListNotificationsResponse 通知列表响应
type ListNotificationsResponse struct {
	Notifications []*NotificationQueueResponse `json:"notifications"`
	Pagination    PaginationWithTotalPage      `json:"pagination"`
}

// CreateNotification 创建通知
func (s *NotificationQueueService) CreateNotification(ctx context.Context, req *CreateNotificationRequest, createdBy uint) (*NotificationQueueResponse, error) {
	// 获取接收人信息
	var recipientName, recipientContact string

	switch req.RecipientType {
	case "lawyer", "admin":
		user, err := s.userRepo.FindByID(ctx, req.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("查询用户失败: %w", err)
		}
		if user == nil {
			return nil, errors.New("用户不存在")
		}
		recipientName = user.Name
		recipientContact = s.getUserContact(user, req.Channel)
	case "client":
		client, err := s.clientRepo.FindByID(ctx, req.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("查询客户失败: %w", err)
		}
		if client == nil {
			return nil, errors.New("客户不存在")
		}
		recipientName = client.Name
		recipientContact = s.getClientContact(client, req.Channel)
	default:
		return nil, errors.New("不支持的接收人类型")
	}

	// 如果使用模板，填充内容
	subject := req.Subject
	content := req.Content

	if req.TemplateID != nil {
		template, err := s.templateRepo.FindByCode(ctx, *req.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("查询模板失败: %w", err)
		}
		if template == nil {
			return nil, errors.New("模板不存在")
		}

		// 使用模板内容
		subject, content = s.fillTemplate(template, req.VariableData)
	}

	// 检测敏感信息
	containsSensitiveInfo := s.checkSensitiveInfo(content)

	// 设置默认优先级
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	notification := &models.NotificationQueue{
		TriggerType:           req.TriggerType,
		TriggerID:             req.TriggerID,
		CaseID:                req.CaseID,
		RecipientType:         req.RecipientType,
		RecipientID:           req.RecipientID,
		RecipientName:         recipientName,
		RecipientContact:      recipientContact,
		Channel:               req.Channel,
		Subject:               subject,
		Content:               content,
		TemplateID:            "",
		Status:                "pending",
		Priority:              priority,
		CreatedBy:             createdBy,
		ContainsSensitiveInfo: containsSensitiveInfo,
		AutoSend:              req.AutoSend,
	}

	if err := s.queueRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("创建通知失败: %w", err)
	}

	return s.GetNotificationByID(ctx, notification.ID)
}

// getUserContact 获取用户联系方式
func (s *NotificationQueueService) getUserContact(user *models.User, channel string) string {
	switch channel {
	case "email":
		return user.Email
	case "sms":
		return user.Phone
	case "wechat":
		// User 模型暂无微信OpenID字段，返回手机号作为备用
		return user.Phone
	default:
		return ""
	}
}

// getClientContact 获取客户联系方式
func (s *NotificationQueueService) getClientContact(client *models.Client, channel string) string {
	switch channel {
	case "email":
		return client.Email
	case "sms":
		return client.Phone
	case "wechat":
		// Client 模型暂无微信OpenID字段，返回手机号作为备用
		return client.Phone
	default:
		return ""
	}
}

// fillTemplate 填充模板内容
func (s *NotificationQueueService) fillTemplate(template *models.NotificationTemplate, data map[string]interface{}) (string, string) {
	subject := template.SubjectTemplate
	content := template.ContentTemplate

	// 简单的变量替换
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		content = strings.ReplaceAll(content, placeholder, replacement)
		subject = strings.ReplaceAll(subject, placeholder, replacement)
	}

	return subject, content
}

// checkSensitiveInfo 检查敏感信息
func (s *NotificationQueueService) checkSensitiveInfo(content string) bool {
	// 身份证号
	if regexp.MustCompile(`\d{15}|\d{17}[\dXx]`).MatchString(content) {
		return true
	}
	// 银行卡号
	if regexp.MustCompile(`\d{16,19}`).MatchString(content) {
		return true
	}
	// 手机号
	if regexp.MustCompile(`1[3-9]\d{9}`).MatchString(content) {
		return true
	}

	return false
}

// GetNotificationByID 根据ID获取通知详情
func (s *NotificationQueueService) GetNotificationByID(ctx context.Context, id uint) (*NotificationQueueResponse, error) {
	notification, err := s.queueRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	if notification == nil {
		return nil, errors.New("通知不存在")
	}

	return s.convertToResponse(ctx, notification), nil
}

// ListNotifications 获取通知列表
func (s *NotificationQueueService) ListNotifications(ctx context.Context, req *ListNotificationsRequest) (*ListNotificationsResponse, error) {
	params := &repositories.NotificationListParams{
		Page:          req.Page,
		PageSize:      req.PageSize,
		Status:        req.Status,
		Channel:       req.Channel,
		RecipientType: req.RecipientType,
		TriggerType:   req.TriggerType,
		CaseID:        req.CaseID,
		DateFrom:      req.DateFrom,
		DateTo:        req.DateTo,
	}

	notifications, total, err := s.queueRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询通知列表失败: %w", err)
	}

	totalPage := int64(0)
	if req.PageSize > 0 {
		totalPage = (total + int64(req.PageSize) - 1) / int64(req.PageSize)
	}

	response := &ListNotificationsResponse{
		Notifications: make([]*NotificationQueueResponse, len(notifications)),
		Pagination: PaginationWithTotalPage{
			Page:      req.Page,
			PageSize:  req.PageSize,
			Total:     total,
			TotalPage: totalPage,
		},
	}

	for i, n := range notifications {
		response.Notifications[i] = s.convertToResponse(ctx, n)
	}

	return response, nil
}

// ApproveNotification 审批通过通知
func (s *NotificationQueueService) ApproveNotification(ctx context.Context, id uint, approvedBy uint) (*NotificationQueueResponse, error) {
	notification, err := s.queueRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	if notification == nil {
		return nil, errors.New("通知不存在")
	}

	if notification.Status != "pending" {
		return nil, errors.New("只有待审批状态的通知可以审批")
	}

	now := time.Now()
	notification.Status = "approved"
	notification.ApprovedBy = &approvedBy
	notification.ApprovedAt = &now

	if err := s.queueRepo.Update(ctx, notification); err != nil {
		return nil, fmt.Errorf("审批通知失败: %w", err)
	}

	return s.GetNotificationByID(ctx, id)
}

// RejectNotification 审批拒绝通知
func (s *NotificationQueueService) RejectNotification(ctx context.Context, id uint) (*NotificationQueueResponse, error) {
	notification, err := s.queueRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	if notification == nil {
		return nil, errors.New("通知不存在")
	}

	if notification.Status != "pending" {
		return nil, errors.New("只有待审批状态的通知可以拒绝")
	}

	notification.Status = "cancelled"

	if err := s.queueRepo.Update(ctx, notification); err != nil {
		return nil, fmt.Errorf("拒绝通知失败: %w", err)
	}

	return s.GetNotificationByID(ctx, id)
}

// SendNotification 发送通知
func (s *NotificationQueueService) SendNotification(ctx context.Context, id uint) (*NotificationQueueResponse, error) {
	notification, err := s.queueRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	if notification == nil {
		return nil, errors.New("通知不存在")
	}

	// 检查状态：只能发送 approved 或 pending(autoSend=true) 状态的通知
	canSend := false
	if notification.Status == "approved" {
		canSend = true
	} else if notification.Status == "pending" && notification.AutoSend {
		canSend = true
	}

	if !canSend {
		return nil, errors.New("通知状态不正确，无法发送")
	}

	// 执行实际发送
	now := time.Now()
	externalMessageID, sendErr := s.executeSend(ctx, notification)

	// 更新发送结果
	if sendErr != nil {
		err = s.queueRepo.UpdateSentInfo(ctx, id, now, "", sendErr.Error())
	} else {
		err = s.queueRepo.UpdateSentInfo(ctx, id, now, externalMessageID, "")
	}

	if err != nil {
		return nil, fmt.Errorf("更新发送状态失败: %w", err)
	}

	if sendErr != nil {
		return nil, fmt.Errorf("发送失败: %w", sendErr)
	}

	return s.GetNotificationByID(ctx, id)
}

// executeSend 执行实际的发送操作
func (s *NotificationQueueService) executeSend(ctx context.Context, notification *models.NotificationQueue) (string, error) {
	// 根据渠道执行不同的发送逻辑
	switch notification.Channel {
	case "email":
		return s.sendEmail(ctx, notification)
	case "sms":
		return s.sendSMS(ctx, notification)
	case "wechat":
		return s.sendWechat(ctx, notification)
	default:
		return "", fmt.Errorf("不支持的通知渠道: %s", notification.Channel)
	}
}

// sendEmail 发送邮件
func (s *NotificationQueueService) sendEmail(ctx context.Context, notification *models.NotificationQueue) (string, error) {
	// TODO: 实际对接邮件服务（如阿里云邮件、SendGrid等）
	// 这里返回模拟的消息ID
	messageID := fmt.Sprintf("EMAIL_%d_%d", notification.ID, time.Now().Unix())
	return messageID, nil
}

// sendSMS 发送短信
func (s *NotificationQueueService) sendSMS(ctx context.Context, notification *models.NotificationQueue) (string, error) {
	// TODO: 实际对接短信服务（如阿里云短信、腾讯云短信等）
	messageID := fmt.Sprintf("SMS_%d_%d", notification.ID, time.Now().Unix())
	return messageID, nil
}

// sendWechat 发送微信消息
func (s *NotificationQueueService) sendWechat(ctx context.Context, notification *models.NotificationQueue) (string, error) {
	// TODO: 实际对接微信公众号/企业微信API
	messageID := fmt.Sprintf("WX_%d_%d", notification.ID, time.Now().Unix())
	return messageID, nil
}

// GetPendingSend 获取待发送的通知列表（用于定时任务）
func (s *NotificationQueueService) GetPendingSend(ctx context.Context, limit int) ([]*NotificationQueueResponse, error) {
	notifications, err := s.queueRepo.GetPendingSend(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("获取待发送通知失败: %w", err)
	}

	result := make([]*NotificationQueueResponse, len(notifications))
	for i, n := range notifications {
		result[i] = s.convertToResponse(ctx, n)
	}

	return result, nil
}

// ProcessPendingSend 处理待发送的通知（定时任务调用）
func (s *NotificationQueueService) ProcessPendingSend(ctx context.Context, limit int) ProcessResult {
	notifications, err := s.queueRepo.GetPendingSend(ctx, limit)
	if err != nil {
		return ProcessResult{Error: err.Error()}
	}

	result := ProcessResult{
		Total:     len(notifications),
		Success:   0,
		Failed:    0,
		SentIDs:   make([]uint, 0),
		FailedIDs: make([]uint, 0),
		Errors:    make(map[uint]string),
	}

	for _, n := range notifications {
		_, err := s.SendNotification(ctx, n.ID)
		if err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, n.ID)
			result.Errors[n.ID] = err.Error()
		} else {
			result.Success++
			result.SentIDs = append(result.SentIDs, n.ID)
		}
	}

	return result
}

// ProcessResult 处理结果
type ProcessResult struct {
	Total     int             `json:"total"`
	Success   int             `json:"success"`
	Failed    int             `json:"failed"`
	SentIDs   []uint          `json:"sent_ids"`
	FailedIDs []uint          `json:"failed_ids"`
	Errors    map[uint]string `json:"errors"`
	Error     string          `json:"error,omitempty"`
}

// MarkAsFailed 标记发送失败
func (s *NotificationQueueService) MarkAsFailed(ctx context.Context, id uint, errorMsg string) (*NotificationQueueResponse, error) {
	notification, err := s.queueRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询通知失败: %w", err)
	}
	if notification == nil {
		return nil, errors.New("通知不存在")
	}

	notification.Status = "failed"
	notification.ErrorMessage = errorMsg
	notification.SentRetryCount++

	if err := s.queueRepo.Update(ctx, notification); err != nil {
		return nil, fmt.Errorf("更新通知状态失败: %w", err)
	}

	return s.GetNotificationByID(ctx, id)
}

// DeleteNotification 删除通知
func (s *NotificationQueueService) DeleteNotification(ctx context.Context, id uint) error {
	notification, err := s.queueRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询通知失败: %w", err)
	}
	if notification == nil {
		return errors.New("通知不存在")
	}

	// 只能删除草稿或已取消的通知
	if notification.Status != "pending" && notification.Status != "cancelled" {
		return errors.New("只能删除待审批或已取消的通知")
	}

	if err := s.queueRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除通知失败: %w", err)
	}

	return nil
}

// convertToResponse 转换为响应格式
func (s *NotificationQueueService) convertToResponse(ctx context.Context, n *models.NotificationQueue) *NotificationQueueResponse {
	resp := &NotificationQueueResponse{
		ID:                    n.ID,
		TriggerType:           n.TriggerType,
		TriggerID:             n.TriggerID,
		CaseID:                n.CaseID,
		RecipientType:         n.RecipientType,
		RecipientID:           n.RecipientID,
		RecipientName:         n.RecipientName,
		RecipientContact:      n.RecipientContact,
		Channel:               n.Channel,
		Subject:               n.Subject,
		Content:               n.Content,
		TemplateID:            &n.TemplateID,
		Status:                n.Status,
		Priority:              n.Priority,
		ContainsSensitiveInfo: n.ContainsSensitiveInfo,
		AutoSend:              n.AutoSend,
		CreatedBy:             n.CreatedBy,
		CreatedAt:             n.CreatedAt.Format("2006-01-02 15:04:05"),
		SentRetryCount:        n.SentRetryCount,
		ErrorMessage:          n.ErrorMessage,
	}

	if n.ApprovedBy != nil {
		resp.ApprovedBy = n.ApprovedBy
	}
	if n.ApprovedAt != nil {
		formatted := n.ApprovedAt.Format("2006-01-02 15:04:05")
		resp.ApprovedAt = &formatted
	}
	if n.SentAt != nil {
		formatted := n.SentAt.Format("2006-01-02 15:04:05")
		resp.SentAt = &formatted
	}

	return resp
}

// GetNotificationsByRecipient 获取接收人的通知列表
func (s *NotificationQueueService) GetNotificationsByRecipient(ctx context.Context, recipientID uint, recipientType string) ([]*NotificationQueueResponse, error) {
	notifications, err := s.queueRepo.GetByRecipient(ctx, recipientID, recipientType)
	if err != nil {
		return nil, fmt.Errorf("查询通知列表失败: %w", err)
	}

	result := make([]*NotificationQueueResponse, len(notifications))
	for i, n := range notifications {
		result[i] = s.convertToResponse(ctx, n)
	}

	return result, nil
}

// GetPendingApprovals 获取待审批的通知列表
func (s *NotificationQueueService) GetPendingApprovals(ctx context.Context) ([]*NotificationQueueResponse, error) {
	notifications, err := s.queueRepo.GetPendingApprovals(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询待审批通知失败: %w", err)
	}

	result := make([]*NotificationQueueResponse, len(notifications))
	for i, n := range notifications {
		result[i] = s.convertToResponse(ctx, n)
	}

	return result, nil
}

// GetNotificationStats 获取通知统计
func (s *NotificationQueueService) GetNotificationStats(ctx context.Context) (*NotificationStats, error) {
	stats := &NotificationStats{}

	// 获取所有通知进行统计
	allNotifications, total, err := s.queueRepo.List(ctx, &repositories.NotificationListParams{
		Page:     1,
		PageSize: 10000,
	})
	if err != nil {
		return nil, fmt.Errorf("查询通知列表失败: %w", err)
	}

	stats.TotalNotifications = total

	for _, n := range allNotifications {
		switch n.Status {
		case "pending":
			stats.PendingNotifications++
		case "approved":
			stats.ApprovedNotifications++
		case "sent":
			stats.SentNotifications++
		case "failed":
			stats.FailedNotifications++
		case "cancelled":
			stats.CancelledNotifications++
		}

		if n.ContainsSensitiveInfo {
			stats.SensitiveNotifications++
		}
	}

	return stats, nil
}

// NotificationStats 通知统计
type NotificationStats struct {
	TotalNotifications     int64 `json:"total_notifications"`
	PendingNotifications   int64 `json:"pending_notifications"`
	ApprovedNotifications  int64 `json:"approved_notifications"`
	SentNotifications      int64 `json:"sent_notifications"`
	FailedNotifications    int64 `json:"failed_notifications"`
	CancelledNotifications int64 `json:"cancelled_notifications"`
	SensitiveNotifications int64 `json:"sensitive_notifications"`
}
