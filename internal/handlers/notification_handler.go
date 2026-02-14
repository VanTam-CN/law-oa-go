package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	db                *gorm.DB
	queueRepo         repositories.NotificationQueueRepository
	templateRepo      repositories.NotificationTemplateRepository
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// NewNotificationHandlerWithDB 创建带DB的通知处理器
func NewNotificationHandlerWithDB(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{
		db:           db,
		queueRepo:    repositories.NewNotificationQueueRepository(db),
		templateRepo: repositories.NewNotificationTemplateRepository(db),
	}
}

// Legacy 通知结构体（兼容旧接口）
type Notification struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	ReadAt    *string `json:"read_at,omitempty"`
}

// NotificationStats 通知统计
type NotificationStats struct {
	Total  int `json:"total"`
	Unread int `json:"unread"`
	Read   int `json:"read"`
}

// =============================================================================
// 旧接口（兼容）
// =============================================================================

// GetNotifications 获取通知列表（旧接口，兼容）
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// 如果有DB，使用新的通知队列
	if h.db != nil {
		h.GetNotificationQueue(c)
		return
	}

	// 模拟通知数据（兼容旧版本）
	notifications := []Notification{
		{
			ID:        "1",
			Title:     "新案件分配",
			Content:   "您有一个新的案件需要处理",
			Type:      "case",
			Status:    "unread",
			CreatedAt: "2025-10-20T14:00:00Z",
		},
		{
			ID:        "2",
			Title:     "客户消息",
			Content:   "客户张三发来了一条新消息",
			Type:      "message",
			Status:    "unread",
			CreatedAt: "2025-10-20T13:30:00Z",
		},
		{
			ID:        "3",
			Title:     "系统通知",
			Content:   "系统将于今晚进行维护",
			Type:      "system",
			Status:    "read",
			CreatedAt: "2025-10-20T12:00:00Z",
			ReadAt:    stringPtr("2025-10-20T13:00:00Z"),
		},
	}

	common.APISuccess(c, gin.H{
		"notifications": notifications,
		"total":         len(notifications),
	})
}

// GetNotificationStats 获取通知统计（旧接口，兼容）
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	// 如果有DB，使用新的通知队列统计
	if h.db != nil {
		h.GetNotificationQueueStats(c)
		return
	}

	stats := NotificationStats{
		Total:  3,
		Unread: 2,
		Read:   1,
	}

	common.APISuccess(c, stats)
}

// =============================================================================
// 通知队列 API
// =============================================================================

// CreateNotificationRequest 创建通知请求
type CreateNotificationRequest struct {
	TriggerType     string `json:"trigger_type" binding:"required"`
	TriggerID       uint   `json:"trigger_id" binding:"required"`
	CaseID          *uint  `json:"case_id"`
	RecipientType   string `json:"recipient_type" binding:"required"`
	RecipientID     uint   `json:"recipient_id" binding:"required"`
	RecipientName   string `json:"recipient_name" binding:"required"`
	RecipientContact string `json:"recipient_contact"`
	Channel         string `json:"channel" binding:"required"`
	Subject         string `json:"subject"`
	Content         string `json:"content" binding:"required"`
	TemplateID      string `json:"template_id"`
	Priority        string `json:"priority"`
	ContainsSensitiveInfo bool `json:"contains_sensitive_info"`
	AutoSend        bool   `json:"auto_send"`
}

// CreateNotification 创建通知
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权")
		return
	}

	notification := &models.NotificationQueue{
		TriggerType:          req.TriggerType,
		TriggerID:            req.TriggerID,
		CaseID:               req.CaseID,
		RecipientType:        req.RecipientType,
		RecipientID:          req.RecipientID,
		RecipientName:        req.RecipientName,
		RecipientContact:     req.RecipientContact,
		Channel:              req.Channel,
		Subject:              req.Subject,
		Content:              req.Content,
		TemplateID:           req.TemplateID,
		Status:               "pending",
		Priority:             req.Priority,
		ContainsSensitiveInfo: req.ContainsSensitiveInfo,
		AutoSend:             req.AutoSend,
		CreatedBy:            userID.(uint),
	}

	// 确定状态：如果包含敏感信息且非自动发送，需要审批
	if req.ContainsSensitiveInfo && !req.AutoSend {
		notification.Status = "pending"
	} else if req.AutoSend {
		notification.Status = "approved"
	}

	if err := h.queueRepo.Create(context.Background(), notification); err != nil {
		common.APIInternalServerError(c, "创建通知失败: "+err.Error())
		return
	}

	common.APISuccess(c, notification)
}

// GetNotificationQueue 获取通知队列列表
func (h *NotificationHandler) GetNotificationQueue(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	priority := c.Query("priority")
	channel := c.Query("channel")
	recipientType := c.Query("recipient_type")
	triggerType := c.Query("trigger_type")
	search := c.Query("search")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	// 获取当前用户ID（作为筛选条件）
	_, exists := c.Get("user_id")
	if exists {
		// 如果是律师，只看自己的通知
		// TODO: 根据角色判断
	}

	params := &repositories.NotificationListParams{
		Page:          page,
		PageSize:      pageSize,
		Status:        status,
		Priority:      priority,
		Channel:       channel,
		RecipientType: recipientType,
		TriggerType:   triggerType,
		Search:        search,
		DateFrom:      dateFrom,
		DateTo:        dateTo,
	}

	notifications, total, err := h.queueRepo.List(context.Background(), params)
	if err != nil {
		common.APIInternalServerError(c, "获取通知列表失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"data": notifications,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetNotificationQueueStats 获取通知队列统计
func (h *NotificationHandler) GetNotificationQueueStats(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	stats, err := h.queueRepo.GetStats(context.Background())
	if err != nil {
		common.APIInternalServerError(c, "获取统计失败: "+err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// GetNotificationByID 获取通知详情
func (h *NotificationHandler) GetNotificationByID(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的通知ID")
		return
	}

	notification, err := h.queueRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取通知失败: "+err.Error())
		return
	}
	if notification == nil {
		common.APINotFound(c, "通知不存在")
		return
	}

	common.APISuccess(c, notification)
}

// UpdateNotificationRequest 更新通知请求
type UpdateNotificationRequest struct {
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	Priority    string `json:"priority"`
}

// UpdateNotification 更新通知
func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的通知ID")
		return
	}

	notification, err := h.queueRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取通知失败: "+err.Error())
		return
	}
	if notification == nil {
		common.APINotFound(c, "通知不存在")
		return
	}

	var req UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	// 只允许在pending状态下更新
	if notification.Status != "pending" {
		common.APIBadRequest(c, "只能修改待处理状态的通知")
		return
	}

	if req.Subject != "" {
		notification.Subject = req.Subject
	}
	if req.Content != "" {
		notification.Content = req.Content
	}
	if req.Priority != "" {
		notification.Priority = req.Priority
	}

	if err := h.queueRepo.Update(context.Background(), notification); err != nil {
		common.APIInternalServerError(c, "更新通知失败: "+err.Error())
		return
	}

	common.APISuccess(c, notification)
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的通知ID")
		return
	}

	if err := h.queueRepo.Delete(context.Background(), uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "通知不存在")
			return
		}
		common.APIInternalServerError(c, "删除通知失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "删除成功"})
}

// ApproveNotification 审批通过通知
func (h *NotificationHandler) ApproveNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的通知ID")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权")
		return
	}

	if err := h.queueRepo.Approve(context.Background(), uint(id), userID.(uint)); err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "通知不存在或状态不允许审批")
			return
		}
		common.APIInternalServerError(c, "审批失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "审批通过"})
}

// RejectNotificationRequest 拒绝通知请求
type RejectNotificationRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RejectNotification 审批拒绝通知
func (h *NotificationHandler) RejectNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的通知ID")
		return
	}

	var req RejectNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	notification, err := h.queueRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取通知失败: "+err.Error())
		return
	}
	if notification == nil {
		common.APINotFound(c, "通知不存在")
		return
	}

	// 更新为取消状态，记录拒绝原因
	notification.Status = "cancelled"
	notification.ErrorMessage = "审批拒绝: " + req.Reason

	if err := h.queueRepo.Update(context.Background(), notification); err != nil {
		common.APIInternalServerError(c, "拒绝失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "已拒绝"})
}

// BatchConfirmNotificationRequest 批量确认请求
type BatchConfirmNotificationRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// BatchConfirmNotification 批量确认通知（标记为已发送）
func (h *NotificationHandler) BatchConfirmNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	var req BatchConfirmNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	if err := h.queueRepo.BatchUpdateStatus(context.Background(), req.IDs, "approved"); err != nil {
		common.APIInternalServerError(c, "批量确认失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message": "批量确认成功",
		"count":   len(req.IDs),
	})
}

// BatchCancelNotification 批量取消通知
func (h *NotificationHandler) BatchCancelNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	var req BatchConfirmNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	if err := h.queueRepo.BatchUpdateStatus(context.Background(), req.IDs, "cancelled"); err != nil {
		common.APIInternalServerError(c, "批量取消失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message": "批量取消成功",
		"count":   len(req.IDs),
	})
}

// SendNotification 发送通知
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的通知ID")
		return
	}

	// 更新发送信息
	now := time.Now()
	err = h.queueRepo.UpdateSentInfo(context.Background(), uint(id), now, "", "")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "通知不存在")
			return
		}
		common.APIInternalServerError(c, "发送失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "发送成功"})
}

// =============================================================================
// 通知模板 API
// =============================================================================

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	TemplateCode     string   `json:"template_code" binding:"required"`
	TemplateName     string   `json:"template_name" binding:"required"`
	Channel          string   `json:"channel" binding:"required"`
	RecipientType    string   `json:"recipient_type" binding:"required"`
	TriggerEvent     string   `json:"trigger_event" binding:"required"`
	SubjectTemplate  string   `json:"subject_template"`
	ContentTemplate  string   `json:"content_template" binding:"required"`
	Variables        []string `json:"variables"`
	AutoSend         bool     `json:"auto_send"`
	RequiresApproval bool     `json:"requires_approval"`
}

// CreateTemplate 创建通知模板
func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	// 获取当前用户ID
	_, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权")
		return
	}

	template := &models.NotificationTemplate{
		TemplateCode:     req.TemplateCode,
		TemplateName:     req.TemplateName,
		Channel:          req.Channel,
		RecipientType:    req.RecipientType,
		TriggerEvent:     req.TriggerEvent,
		SubjectTemplate:  req.SubjectTemplate,
		ContentTemplate:  req.ContentTemplate,
		Variables:        models.JSON{},
		AutoSend:         req.AutoSend,
		RequiresApproval: req.RequiresApproval,
		IsActive:         true,
	}

	// 构建变量JSON
	if len(req.Variables) > 0 {
		varMap := make(map[string]interface{})
		for _, v := range req.Variables {
			varMap[v] = ""
		}
		template.Variables = models.JSON(varMap)
	}

	if err := h.templateRepo.Create(context.Background(), template); err != nil {
		common.APIInternalServerError(c, "创建模板失败: "+err.Error())
		return
	}

	common.APISuccess(c, template)
}

// GetTemplates 获取通知模板列表
func (h *NotificationHandler) GetTemplates(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	channel := c.Query("channel")
	recipientType := c.Query("recipient_type")
	triggerEvent := c.Query("trigger_event")
	search := c.Query("search")

	// 解析is_active
	var isActive *bool
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActiveStr == "true" {
			val := true
			isActive = &val
		} else if isActiveStr == "false" {
			val := false
			isActive = &val
		}
	}

	params := &repositories.TemplateListParams{
		Page:          page,
		PageSize:      pageSize,
		Channel:       channel,
		RecipientType: recipientType,
		TriggerEvent:  triggerEvent,
		IsActive:      isActive,
		Search:        search,
	}

	templates, total, err := h.templateRepo.List(context.Background(), params)
	if err != nil {
		common.APIInternalServerError(c, "获取模板列表失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"data": templates,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetActiveTemplates 获取启用的模板列表
func (h *NotificationHandler) GetActiveTemplates(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	templates, err := h.templateRepo.GetActive(context.Background())
	if err != nil {
		common.APIInternalServerError(c, "获取模板失败: "+err.Error())
		return
	}

	common.APISuccess(c, templates)
}

// GetTemplateByCode 根据代码获取模板
func (h *NotificationHandler) GetTemplateByCode(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	code := c.Param("code")
	template, err := h.templateRepo.FindByCode(context.Background(), code)
	if err != nil {
		common.APIInternalServerError(c, "获取模板失败: "+err.Error())
		return
	}
	if template == nil {
		common.APINotFound(c, "模板不存在")
		return
	}

	common.APISuccess(c, template)
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	TemplateName     string   `json:"template_name"`
	SubjectTemplate  string   `json:"subject_template"`
	ContentTemplate  string   `json:"content_template"`
	Variables        []string `json:"variables"`
	AutoSend         bool     `json:"auto_send"`
	RequiresApproval bool     `json:"requires_approval"`
	IsActive         *bool    `json:"is_active"`
}

// UpdateTemplate 更新通知模板
func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的模板ID")
		return
	}

	template, err := h.templateRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取模板失败: "+err.Error())
		return
	}
	if template == nil {
		common.APINotFound(c, "模板不存在")
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	if req.TemplateName != "" {
		template.TemplateName = req.TemplateName
	}
	if req.SubjectTemplate != "" {
		template.SubjectTemplate = req.SubjectTemplate
	}
	if req.ContentTemplate != "" {
		template.ContentTemplate = req.ContentTemplate
	}
	if req.Variables != nil {
		varMap := make(map[string]interface{})
		for _, v := range req.Variables {
			varMap[v] = ""
		}
		template.Variables = models.JSON(varMap)
	}
	template.AutoSend = req.AutoSend
	template.RequiresApproval = req.RequiresApproval
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}

	if err := h.templateRepo.Update(context.Background(), template); err != nil {
		common.APIInternalServerError(c, "更新模板失败: "+err.Error())
		return
	}

	common.APISuccess(c, template)
}

// DeleteTemplate 删除通知模板
func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的模板ID")
		return
	}

	if err := h.templateRepo.Delete(context.Background(), uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "模板不存在")
			return
		}
		common.APIInternalServerError(c, "删除模板失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "删除成功"})
}

// ToggleTemplateActive 切换模板启用状态
func (h *NotificationHandler) ToggleTemplateActive(c *gin.Context) {
	if h.db == nil {
		common.APIInternalServerError(c, "数据库未初始化")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的模板ID")
		return
	}

	template, err := h.templateRepo.FindByID(context.Background(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取模板失败: "+err.Error())
		return
	}
	if template == nil {
		common.APINotFound(c, "模板不存在")
		return
	}

	// 切换状态
	template.IsActive = !template.IsActive
	if err := h.templateRepo.Update(context.Background(), template); err != nil {
		common.APIInternalServerError(c, "更新模板失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message":   "状态更新成功",
		"is_active": template.IsActive,
	})
}

// =============================================================================
// 辅助函数
// =============================================================================

func stringPtr(s string) *string {
	return &s
}