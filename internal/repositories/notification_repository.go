package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// NotificationQueueRepository 通知队列数据仓库接口
type NotificationQueueRepository interface {
	// Create 创建通知记录
	Create(ctx context.Context, notification *models.NotificationQueue) error
	// FindByID 根据ID查找通知
	FindByID(ctx context.Context, id uint) (*models.NotificationQueue, error)
	// Update 更新通知
	Update(ctx context.Context, notification *models.NotificationQueue) error
	// Delete 删除通知
	Delete(ctx context.Context, id uint) error
	// List 通知列表查询
	List(ctx context.Context, params *NotificationListParams) ([]*models.NotificationQueue, int64, error)
	// GetByStatus 根据状态获取通知列表
	GetByStatus(ctx context.Context, status string) ([]*models.NotificationQueue, error)
	// GetByRecipient 获取接收人的通知列表
	GetByRecipient(ctx context.Context, recipientID uint, recipientType string) ([]*models.NotificationQueue, error)
	// UpdateStatus 更新通知状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// UpdateSentInfo 更新发送信息
	UpdateSentInfo(ctx context.Context, id uint, sentAt time.Time, externalMessageID string, errorMsg string) error
	// Approve 批准通知
	Approve(ctx context.Context, id uint, approvedBy uint) error
	// GetPendingApprovals 获取待审批的通知列表
	GetPendingApprovals(ctx context.Context) ([]*models.NotificationQueue, error)
	// GetPendingSend 获取待发送的通知列表
	GetPendingSend(ctx context.Context, limit int) ([]*models.NotificationQueue, error)
	// BatchUpdateStatus 批量更新状态
	BatchUpdateStatus(ctx context.Context, ids []uint, status string) error
	// CountByStatus 按状态统计
	CountByStatus(ctx context.Context, status string) (int64, error)
	// GetStats 获取通知统计
	GetStats(ctx context.Context) (*NotificationStats, error)
}

// NotificationQueueRepositoryImpl 通知队列数据仓库实现
type NotificationQueueRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationQueueRepository 创建通知队列数据仓库实例
func NewNotificationQueueRepository(db *gorm.DB) NotificationQueueRepository {
	return &NotificationQueueRepositoryImpl{db: db}
}

// Create 创建通知记录
func (r *NotificationQueueRepositoryImpl) Create(ctx context.Context, notification *models.NotificationQueue) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// FindByID 根据ID查找通知
func (r *NotificationQueueRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.NotificationQueue, error) {
	var notification models.NotificationQueue
	err := r.db.WithContext(ctx).
		Preload("CreatedByUser").
		Preload("ApprovedByUser").
		First(&notification, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

// Update 更新通知
func (r *NotificationQueueRepositoryImpl) Update(ctx context.Context, notification *models.NotificationQueue) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// Delete 删除通知
func (r *NotificationQueueRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.NotificationQueue{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// NotificationListParams 通知列表查询参数
type NotificationListParams struct {
	Page            int
	PageSize        int
	Status          string
	Priority        string
	Channel         string
	RecipientType   string
	RecipientID     uint
	TriggerType     string
	CaseID          uint
	CreatedBy       uint
	Search          string
	DateFrom        string
	DateTo          string
	OnlyPending     bool // 只获取待审批的
}

// List 通知列表查询
func (r *NotificationQueueRepositoryImpl) List(ctx context.Context, params *NotificationListParams) ([]*models.NotificationQueue, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.NotificationQueue{}).
		Preload("CreatedByUser").
		Preload("ApprovedByUser")

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 优先级筛选
	if params.Priority != "" {
		query = query.Where("priority = ?", params.Priority)
	}

	// 渠道筛选
	if params.Channel != "" {
		query = query.Where("channel = ?", params.Channel)
	}

	// 接收人类型筛选
	if params.RecipientType != "" {
		query = query.Where("recipient_type = ?", params.RecipientType)
	}

	// 接收人筛选
	if params.RecipientID > 0 {
		query = query.Where("recipient_id = ?", params.RecipientID)
	}

	// 触发类型筛选
	if params.TriggerType != "" {
		query = query.Where("trigger_type = ?", params.TriggerType)
	}

	// 案件筛选
	if params.CaseID > 0 {
		query = query.Where("case_id = ?", params.CaseID)
	}

	// 创建人筛选
	if params.CreatedBy > 0 {
		query = query.Where("created_by = ?", params.CreatedBy)
	}

	// 只获取待审批的
	if params.OnlyPending {
		query = query.Where("status = ? AND contains_sensitive_info = ?", "pending", true)
	}

	// 搜索：标题或内容
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("subject LIKE ? OR content LIKE ?", searchTerm, searchTerm)
	}

	// 日期范围筛选
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var notifications []models.NotificationQueue
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.NotificationQueue, len(notifications))
	for i, n := range notifications {
		result[i] = &n
	}

	return result, total, nil
}

// GetByStatus 根据状态获取通知列表
func (r *NotificationQueueRepositoryImpl) GetByStatus(ctx context.Context, status string) ([]*models.NotificationQueue, error) {
	var notifications []models.NotificationQueue
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Preload("CreatedByUser").
		Order("created_at ASC").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.NotificationQueue, len(notifications))
	for i, n := range notifications {
		result[i] = &n
	}
	return result, nil
}

// GetByRecipient 获取接收人的通知列表
func (r *NotificationQueueRepositoryImpl) GetByRecipient(ctx context.Context, recipientID uint, recipientType string) ([]*models.NotificationQueue, error) {
	var notifications []models.NotificationQueue
	query := r.db.WithContext(ctx).Where("recipient_id = ?", recipientID)

	if recipientType != "" {
		query = query.Where("recipient_type = ?", recipientType)
	}

	err := query.
		Preload("CreatedByUser").
		Order("created_at DESC").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.NotificationQueue, len(notifications))
	for i, n := range notifications {
		result[i] = &n
	}
	return result, nil
}

// UpdateStatus 更新通知状态
func (r *NotificationQueueRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).
		Model(&models.NotificationQueue{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateSentInfo 更新发送信息
func (r *NotificationQueueRepositoryImpl) UpdateSentInfo(ctx context.Context, id uint, sentAt time.Time, externalMessageID string, errorMsg string) error {
	updates := map[string]interface{}{
		"sent_at":           sentAt,
		"external_message_id": externalMessageID,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
		updates["status"] = "failed"
		updates["sent_retry_count"] = gorm.Expr("sent_retry_count + 1")
	} else {
		updates["status"] = "sent"
	}

	result := r.db.WithContext(ctx).
		Model(&models.NotificationQueue{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Approve 批准通知
func (r *NotificationQueueRepositoryImpl) Approve(ctx context.Context, id uint, approvedBy uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.NotificationQueue{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(map[string]interface{}{
			"status":      "approved",
			"approved_by": approvedBy,
			"approved_at": now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetPendingApprovals 获取待审批的通知列表
func (r *NotificationQueueRepositoryImpl) GetPendingApprovals(ctx context.Context) ([]*models.NotificationQueue, error) {
	var notifications []models.NotificationQueue
	err := r.db.WithContext(ctx).
		Where("status = ? AND contains_sensitive_info = ?", "pending", true).
		Preload("CreatedByUser").
		Order("created_at ASC").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.NotificationQueue, len(notifications))
	for i, n := range notifications {
		result[i] = &n
	}
	return result, nil
}

// GetPendingSend 获取待发送的通知列表
func (r *NotificationQueueRepositoryImpl) GetPendingSend(ctx context.Context, limit int) ([]*models.NotificationQueue, error) {
	var notifications []models.NotificationQueue
	query := r.db.WithContext(ctx).
		Where("(status = ? OR (status = ? AND auto_send = ?))", "approved", "pending", true).
		Order("priority DESC, created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.NotificationQueue, len(notifications))
	for i, n := range notifications {
		result[i] = &n
	}
	return result, nil
}

// BatchUpdateStatus 批量更新状态
func (r *NotificationQueueRepositoryImpl) BatchUpdateStatus(ctx context.Context, ids []uint, status string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&models.NotificationQueue{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}

// CountByStatus 按状态统计
func (r *NotificationQueueRepositoryImpl) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.NotificationQueue{}).
		Where("status = ?", status).
		Count(&count).Error
	return count, err
}

// NotificationStats 通知统计
type NotificationStats struct {
	Total          int64 `json:"total"`
	Pending        int64 `json:"pending"`
	Approved       int64 `json:"approved"`
	Sent           int64 `json:"sent"`
	Failed         int64 `json:"failed"`
	Cancelled      int64 `json:"cancelled"`
	PendingApproval int64 `json:"pending_approval"` // 待审批（包含敏感信息的pending）
	AutoSendCount   int64 `json:"auto_send_count"`  // 自动发送数量
}

// GetStats 获取通知统计
func (r *NotificationQueueRepositoryImpl) GetStats(ctx context.Context) (*NotificationStats, error) {
	var stats NotificationStats

	// 总数
	r.db.WithContext(ctx).Model(&models.NotificationQueue{}).Count(&stats.Total)

	// 按状态统计
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	r.db.WithContext(ctx).Model(&models.NotificationQueue{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts)

	for _, sc := range statusCounts {
		switch sc.Status {
		case "pending":
			stats.Pending = sc.Count
		case "approved":
			stats.Approved = sc.Count
		case "sent":
			stats.Sent = sc.Count
		case "failed":
			stats.Failed = sc.Count
		case "cancelled":
			stats.Cancelled = sc.Count
		}
	}

	// 待审批（包含敏感信息的pending）
	r.db.WithContext(ctx).Model(&models.NotificationQueue{}).
		Where("status = ? AND contains_sensitive_info = ?", "pending", true).
		Count(&stats.PendingApproval)

	// 自动发送数量
	r.db.WithContext(ctx).Model(&models.NotificationQueue{}).
		Where("auto_send = ?", true).
		Count(&stats.AutoSendCount)

	return &stats, nil
}

// NotificationTemplateRepository 通知模板数据仓库接口
type NotificationTemplateRepository interface {
	// Create 创建模板
	Create(ctx context.Context, template *models.NotificationTemplate) error
	// FindByID 根据ID查找模板
	FindByID(ctx context.Context, id uint) (*models.NotificationTemplate, error)
	// FindByCode 根据模板代码查找模板
	FindByCode(ctx context.Context, code string) (*models.NotificationTemplate, error)
	// Update 更新模板
	Update(ctx context.Context, template *models.NotificationTemplate) error
	// Delete 删除模板
	Delete(ctx context.Context, id uint) error
	// List 模板列表查询
	List(ctx context.Context, params *TemplateListParams) ([]*models.NotificationTemplate, int64, error)
	// GetActive 获取启用的模板列表
	GetActive(ctx context.Context) ([]*models.NotificationTemplate, error)
	// GetByChannelAndEvent 根据渠道和事件获取模板
	GetByChannelAndEvent(ctx context.Context, channel, triggerEvent string) ([]*models.NotificationTemplate, error)
	// UpdateActiveStatus 更新启用状态
	UpdateActiveStatus(ctx context.Context, id uint, isActive bool) error
}

// NotificationTemplateRepositoryImpl 通知模板数据仓库实现
type NotificationTemplateRepositoryImpl struct {
	db *gorm.DB
}

// NewNotificationTemplateRepository 创建通知模板数据仓库实例
func NewNotificationTemplateRepository(db *gorm.DB) NotificationTemplateRepository {
	return &NotificationTemplateRepositoryImpl{db: db}
}

// Create 创建模板
func (r *NotificationTemplateRepositoryImpl) Create(ctx context.Context, template *models.NotificationTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// FindByID 根据ID查找模板
func (r *NotificationTemplateRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	err := r.db.WithContext(ctx).First(&template, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// FindByCode 根据模板代码查找模板
func (r *NotificationTemplateRepositoryImpl) FindByCode(ctx context.Context, code string) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	err := r.db.WithContext(ctx).
		Where("template_code = ?", code).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

// Update 更新模板
func (r *NotificationTemplateRepositoryImpl) Update(ctx context.Context, template *models.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// Delete 删除模板
func (r *NotificationTemplateRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.NotificationTemplate{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// TemplateListParams 模板列表查询参数
type TemplateListParams struct {
	Page         int
	PageSize     int
	Channel      string
	RecipientType string
	TriggerEvent string
	IsActive     *bool
	Search       string
}

// List 模板列表查询
func (r *NotificationTemplateRepositoryImpl) List(ctx context.Context, params *TemplateListParams) ([]*models.NotificationTemplate, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.NotificationTemplate{})

	// 渠道筛选
	if params.Channel != "" {
		query = query.Where("channel = ?", params.Channel)
	}

	// 接收人类型筛选
	if params.RecipientType != "" {
		query = query.Where("recipient_type = ?", params.RecipientType)
	}

	// 触发事件筛选
	if params.TriggerEvent != "" {
		query = query.Where("trigger_event = ?", params.TriggerEvent)
	}

	// 启用状态筛选
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// 搜索：模板名称或代码
	if params.Search != "" {
		searchTerm := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where("LOWER(template_name) LIKE ? OR LOWER(template_code) LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var templates []models.NotificationTemplate
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.NotificationTemplate, len(templates))
	for i, t := range templates {
		result[i] = &t
	}

	return result, total, nil
}

// GetActive 获取启用的模板列表
func (r *NotificationTemplateRepositoryImpl) GetActive(ctx context.Context) ([]*models.NotificationTemplate, error) {
	var templates []models.NotificationTemplate
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("channel, trigger_event").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.NotificationTemplate, len(templates))
	for i, t := range templates {
		result[i] = &t
	}
	return result, nil
}

// GetByChannelAndEvent 根据渠道和事件获取模板
func (r *NotificationTemplateRepositoryImpl) GetByChannelAndEvent(ctx context.Context, channel, triggerEvent string) ([]*models.NotificationTemplate, error) {
	var templates []models.NotificationTemplate
	err := r.db.WithContext(ctx).
		Where("channel = ? AND trigger_event = ? AND is_active = ?", channel, triggerEvent, true).
		Find(&templates).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.NotificationTemplate, len(templates))
	for i, t := range templates {
		result[i] = &t
	}
	return result, nil
}

// UpdateActiveStatus 更新启用状态
func (r *NotificationTemplateRepositoryImpl) UpdateActiveStatus(ctx context.Context, id uint, isActive bool) error {
	result := r.db.WithContext(ctx).
		Model(&models.NotificationTemplate{}).
		Where("id = ?", id).
		Update("is_active", isActive)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ===================================
// 代管款仓储接口
// ===================================

// TrustAccountRepository 代管款账户仓储接口
type TrustAccountRepository interface {
	// Create 创建账户
	Create(ctx context.Context, account *models.ClientTrustAccount) error
	// FindByID 根据ID查找账户
	FindByID(ctx context.Context, id uint) (*models.ClientTrustAccount, error)
	// FindByCode 根据账户编号查找账户
	FindByCode(ctx context.Context, code string) (*models.ClientTrustAccount, error)
	// Update 更新账户
	Update(ctx context.Context, account *models.ClientTrustAccount) error
	// Delete 删除账户
	Delete(ctx context.Context, id uint) error
	// List 账户列表查询
	List(ctx context.Context, params *TrustAccountListParams) ([]*models.ClientTrustAccount, int64, error)
	// GetByClientID 获取客户的账户列表
	GetByClientID(ctx context.Context, clientID uint) ([]*models.ClientTrustAccount, int64, error)
}

// TrustTransactionRepository 代管款交易仓储接口
type TrustTransactionRepository interface {
	// Create 创建交易
	Create(ctx context.Context, transaction *models.ClientTrustTransaction) error
	// FindByID 根据ID查找交易
	FindByID(ctx context.Context, id uint) (*models.ClientTrustTransaction, error)
	// FindByCode 根据交易编号查找交易
	FindByCode(ctx context.Context, code string) (*models.ClientTrustTransaction, error)
	// Update 更新交易
	Update(ctx context.Context, transaction *models.ClientTrustTransaction) error
	// Delete 删除交易
	Delete(ctx context.Context, id uint) error
	// List 交易列表查询
	List(ctx context.Context, params *TrustTransactionListParams) ([]*models.ClientTrustTransaction, int64, error)
	// GetByAccountID 获取账户的交易列表
	GetByAccountID(ctx context.Context, accountID uint) ([]*models.ClientTrustTransaction, int64, error)
}


// TrustAccountListParams 代管款账户列表查询参数
type TrustAccountListParams struct {
	Page     int
	PageSize int
	ClientID uint
	Status   string
	Currency string
}

// TrustTransactionListParams 代管款交易列表查询参数
type TrustTransactionListParams struct {
	Page      int
	PageSize  int
	AccountID *uint
	Status    string
	Type      string
	DateFrom  string
	DateTo    string
}

// SensitiveWordRepository 敏感词仓储接口
type SensitiveWordRepository interface {
	// Create 创建敏感词
	Create(ctx context.Context, word *models.SensitiveWord) error
	// FindByID 根据ID查找敏感词
	FindByID(ctx context.Context, id uint) (*models.SensitiveWord, error)
	// FindByWord 根据词语查找敏感词
	FindByWord(ctx context.Context, word string) (*models.SensitiveWord, error)
	// Update 更新敏感词
	Update(ctx context.Context, word *models.SensitiveWord) error
	// Delete 删除敏感词
	Delete(ctx context.Context, id uint) error
	// List 敏感词列表查询
	List(ctx context.Context, params *SensitiveWordListParams) ([]*models.SensitiveWord, int64, error)
	// GetByCategory 获取分类的敏感词列表
	GetByCategory(ctx context.Context, category string) ([]*models.SensitiveWord, error)
	// GetActive 获取启用的敏感词列表
	GetActive(ctx context.Context) ([]*models.SensitiveWord, error)
	// UpdateHitCount 更新命中计数
	UpdateHitCount(ctx context.Context, id uint, hitTime time.Time) error
}

// SensitiveWordListParams 敏感词列表查询参数
type SensitiveWordListParams struct {
	Page     int
	PageSize int
	WordType string
	Category string
	Severity string
	Search   string
	IsActive *bool
}

// SensitiveWordRepositoryImpl 敏感词仓储实现
type SensitiveWordRepositoryImpl struct {
	db *gorm.DB
}

// NewSensitiveWordRepository 创建敏感词仓储实例
func NewSensitiveWordRepository(db *gorm.DB) SensitiveWordRepository {
	return &SensitiveWordRepositoryImpl{db: db}
}

// Create 创建敏感词
func (r *SensitiveWordRepositoryImpl) Create(ctx context.Context, word *models.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

// FindByID 根据ID查找敏感词
func (r *SensitiveWordRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.SensitiveWord, error) {
	var word models.SensitiveWord
	err := r.db.WithContext(ctx).First(&word, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &word, nil
}

// FindByWord 根据词语查找敏感词
func (r *SensitiveWordRepositoryImpl) FindByWord(ctx context.Context, word string) (*models.SensitiveWord, error) {
	var w models.SensitiveWord
	err := r.db.WithContext(ctx).
		Where("word = ?", word).
		First(&w).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

// Update 更新敏感词
func (r *SensitiveWordRepositoryImpl) Update(ctx context.Context, word *models.SensitiveWord) error {
	return r.db.WithContext(ctx).Save(word).Error
}

// Delete 删除敏感词
func (r *SensitiveWordRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.SensitiveWord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 敏感词列表查询
func (r *SensitiveWordRepositoryImpl) List(ctx context.Context, params *SensitiveWordListParams) ([]*models.SensitiveWord, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.SensitiveWord{})

	// 词类型筛选
	if params.WordType != "" {
		query = query.Where("word_type = ?", params.WordType)
	}

	// 分类筛选
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}

	// 严重程度筛选
	if params.Severity != "" {
		query = query.Where("severity = ?", params.Severity)
	}

	// 启用状态筛选
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	// 搜索
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("word LIKE ? OR description LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var words []models.SensitiveWord
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&words).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.SensitiveWord, len(words))
	for i, w := range words {
		result[i] = &w
	}

	return result, total, nil
}

// GetByCategory 获取分类的敏感词列表
func (r *SensitiveWordRepositoryImpl) GetByCategory(ctx context.Context, category string) ([]*models.SensitiveWord, error) {
	var words []models.SensitiveWord
	err := r.db.WithContext(ctx).
		Where("category = ? AND is_active = ?", category, true).
		Order("severity DESC, word ASC").
		Find(&words).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.SensitiveWord, len(words))
	for i, w := range words {
		result[i] = &w
	}
	return result, nil
}

// GetActive 获取启用的敏感词列表
func (r *SensitiveWordRepositoryImpl) GetActive(ctx context.Context) ([]*models.SensitiveWord, error) {
	var words []models.SensitiveWord
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("category, severity DESC, word ASC").
		Find(&words).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.SensitiveWord, len(words))
	for i, w := range words {
		result[i] = &w
	}
	return result, nil
}

// UpdateHitCount 更新命中计数
func (r *SensitiveWordRepositoryImpl) UpdateHitCount(ctx context.Context, id uint, hitTime time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.SensitiveWord{}).
		Where("id = ?", id).
		UpdateColumn("hit_count", gorm.Expr("hit_count + 1")).
		Update("last_hit_at", hitTime).
		Error
}

// ContentFilterLogRepository 内容过滤日志仓储接口
type ContentFilterLogRepository interface {
	// Create 创建过滤日志
	Create(ctx context.Context, log *models.ContentFilterLog) error
	// FindByID 根据ID查找日志
	FindByID(ctx context.Context, id uint) (*models.ContentFilterLog, error)
	// List 日志列表查询
	List(ctx context.Context, params *ContentFilterLogListParams) ([]*models.ContentFilterLog, int64, error)
	// Update 更新日志
	Update(ctx context.Context, log *models.ContentFilterLog) error
}

// ContentFilterLogListParams 内容过滤日志查询参数
type ContentFilterLogListParams struct {
	Page         int
	PageSize     int
	ContentType  string
	ContentID    uint
	IsBlocked    *bool
	ProcessedBy  uint
	DateFrom     string
	DateTo       string
}

// ContentFilterLogRepositoryImpl 内容过滤日志仓储实现
type ContentFilterLogRepositoryImpl struct {
	db *gorm.DB
}

// NewContentFilterLogRepository 创建内容过滤日志仓储实例
func NewContentFilterLogRepository(db *gorm.DB) ContentFilterLogRepository {
	return &ContentFilterLogRepositoryImpl{db: db}
}

// Create 创建过滤日志
func (r *ContentFilterLogRepositoryImpl) Create(ctx context.Context, log *models.ContentFilterLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByID 根据ID查找日志
func (r *ContentFilterLogRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.ContentFilterLog, error) {
	var log models.ContentFilterLog
	err := r.db.WithContext(ctx).First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// List 日志列表查询
func (r *ContentFilterLogRepositoryImpl) List(ctx context.Context, params *ContentFilterLogListParams) ([]*models.ContentFilterLog, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.ContentFilterLog{})

	// 内容类型筛选
	if params.ContentType != "" {
		query = query.Where("content_type = ?", params.ContentType)
	}

	// 内容ID筛选
	if params.ContentID > 0 {
		query = query.Where("content_id = ?", params.ContentID)
	}

	// 是否被拦截筛选
	if params.IsBlocked != nil {
		query = query.Where("is_blocked = ?", *params.IsBlocked)
	}

	// 处理人筛选
	if params.ProcessedBy > 0 {
		query = query.Where("processed_by = ?", params.ProcessedBy)
	}

	// 日期范围筛选
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.ContentFilterLog
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.ContentFilterLog, len(logs))
	for i, l := range logs {
		result[i] = &l
	}

	return result, total, nil
}

// Update 更新日志
func (r *ContentFilterLogRepositoryImpl) Update(ctx context.Context, log *models.ContentFilterLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
