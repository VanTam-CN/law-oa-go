package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// Inbox Repository Sentinel Errors
var (
	ErrInboxItemNotFound            = errors.New("inbox item not found")
	ErrInboxItemInvalid             = errors.New("invalid inbox item data")
	ErrInboxItemDeletionUnavailable = errors.New("inbox item deletion is disabled for audit retention")
	ErrReminderRuleNotFound         = errors.New("reminder rule not found")
)

// InboxRepository 待办事项数据仓库接口
type InboxRepository interface {
	// Create 创建待办事项
	Create(ctx context.Context, item *models.InboxItem) error
	// FindByID 根据ID查找待办事项
	FindByID(ctx context.Context, id uint) (*models.InboxItem, error)
	// FindByIDAndUserID 根据ID和所属用户查找待办事项
	FindByIDAndUserID(ctx context.Context, id, userID uint) (*models.InboxItem, error)
	// Update 更新待办事项
	Update(ctx context.Context, item *models.InboxItem) error
	// UpdateByUserID 仅更新指定用户的待办事项
	UpdateByUserID(ctx context.Context, item *models.InboxItem, userID uint) error
	// Delete 删除待办事项
	Delete(ctx context.Context, id uint) error
	// DeleteByUserID 仅删除指定用户的待办事项
	DeleteByUserID(ctx context.Context, id, userID uint) error
	// List 查询待办事项列表
	List(ctx context.Context, params *InboxListParams) ([]*models.InboxItem, int64, error)
	// FindByUserID 根据用户ID查询待办事项
	FindByUserID(ctx context.Context, userID uint) ([]*models.InboxItem, error)
	// MarkAsRead 标记为已读
	MarkAsRead(ctx context.Context, id uint) error
	// MarkAsReadByUserID 仅标记指定用户的待办事项为已读
	MarkAsReadByUserID(ctx context.Context, id, userID uint) error
	// MarkAsCompleted 标记为已完成
	MarkAsCompleted(ctx context.Context, id uint) error
	// MarkAsCompletedByUserID 仅标记指定用户的待办事项为已完成
	MarkAsCompletedByUserID(ctx context.Context, id, userID uint) error
	// Snooze 延后待办事项
	Snooze(ctx context.Context, id uint, until time.Time) error
	// SnoozeByUserID 仅延后指定用户的待办事项
	SnoozeByUserID(ctx context.Context, id, userID uint, until time.Time) error
	// Escalate 升级待办事项
	Escalate(ctx context.Context, id uint) error
	// GetUnreadCount 获取未读数量
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	// GetDueItems 获取到期待办事项
	GetDueItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error)
	// GetOverdueCriticalItems 获取超时的critical待办
	GetOverdueCriticalItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error)

	// Reminder Rules
	CreateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error
	GetReminderRules(ctx context.Context, isActive bool) ([]*models.InboxReminderRule, error)
	GetReminderRuleByTypeAndPriority(ctx context.Context, dueDateType, priority string) (*models.InboxReminderRule, error)
	UpdateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error
	DeleteReminderRule(ctx context.Context, id uint) error
}

// InboxReminderRuleRepository 提醒规则数据仓库接口
type InboxReminderRuleRepository interface {
	// Create 创建提醒规则
	Create(ctx context.Context, rule *models.InboxReminderRule) error
	// FindByID 根据ID查找规则
	FindByID(ctx context.Context, id uint) (*models.InboxReminderRule, error)
	// Update 更新规则
	Update(ctx context.Context, rule *models.InboxReminderRule) error
	// Delete 删除规则
	Delete(ctx context.Context, id uint) error
	// FindAll 查询所有规则
	FindAll(ctx context.Context) ([]*models.InboxReminderRule, error)
	// FindActive 查询启用的规则
	FindActive(ctx context.Context) ([]*models.InboxReminderRule, error)
	// FindByDueDateType 根据日期类型查询规则
	FindByDueDateType(ctx context.Context, dueDateType string) ([]*models.InboxReminderRule, error)
}

// InboxListParams 待办事项列表查询参数
type InboxListParams struct {
	Page        int
	PageSize    int
	UserID      uint
	IsRead      *bool
	IsCompleted *bool
	Priority    string
	SourceType  string
	DueBefore   *time.Time
	DueAfter    *time.Time
	Search      string
	OrderBy     string // due_date, priority, created_at
}

// InboxStats 待办事项统计信息
type InboxStats struct {
	Total       int64
	Unread      int64
	Pending     int64
	Completed   int64
	Critical    int64
	High        int64
	Overdue     int64
	DueToday    int64
	DueThisWeek int64
}

// InboxRepositoryImpl 待办事项数据仓库的GORM实现
type InboxRepositoryImpl struct {
	*BaseRepository[models.InboxItem]
	db *gorm.DB
}

// NewInboxRepository 创建待办事项数据仓库实例
func NewInboxRepository(db *gorm.DB) InboxRepository {
	return &InboxRepositoryImpl{
		BaseRepository: NewBaseRepository[models.InboxItem](db),
		db:             db,
	}
}

// Create 创建待办事项
func (r *InboxRepositoryImpl) Create(ctx context.Context, item *models.InboxItem) error {
	return r.BaseRepository.Create(ctx, item)
}

// FindByID 根据ID查找待办事项
func (r *InboxRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.InboxItem, error) {
	item, err := r.BaseRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("find", "inbox_item", id, ErrInboxItemNotFound)
		}
		return nil, NewRepositoryErrorWithID("find", "inbox_item", id, err)
	}
	return item, nil
}

func (r *InboxRepositoryImpl) FindByIDAndUserID(ctx context.Context, id, userID uint) (*models.InboxItem, error) {
	var item models.InboxItem
	if err := r.ownedItemQuery(ctx, id, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NewRepositoryErrorWithID("find", "inbox_item", id, ErrInboxItemNotFound)
		}
		return nil, NewRepositoryErrorWithID("find", "inbox_item", id, err)
	}
	return &item, nil
}

// Update 更新待办事项
func (r *InboxRepositoryImpl) Update(ctx context.Context, item *models.InboxItem) error {
	return r.BaseRepository.Update(ctx, item.ID, inboxItemUpdates(item))
}

func (r *InboxRepositoryImpl) UpdateByUserID(ctx context.Context, item *models.InboxItem, userID uint) error {
	result := r.ownedItemQuery(ctx, item.ID, userID).
		Model(&models.InboxItem{}).
		Updates(inboxItemUpdates(item))
	if result.Error != nil {
		return NewRepositoryErrorWithID("update", "inbox_item", item.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		// MySQL may report zero affected rows when every submitted value is unchanged.
		if _, err := r.FindByIDAndUserID(ctx, item.ID, userID); err != nil {
			return err
		}
	}
	return nil
}

func inboxItemUpdates(item *models.InboxItem) map[string]interface{} {
	return map[string]interface{}{
		"title":          item.Title,
		"content":        item.Content,
		"priority":       item.Priority,
		"due_date":       item.DueDate,
		"due_date_type":  item.DueDateType,
		"is_read":        item.IsRead,
		"read_at":        item.ReadAt,
		"is_completed":   item.IsCompleted,
		"completed_at":   item.CompletedAt,
		"reminder_sent":  item.ReminderSent,
		"reminder_count": item.ReminderCount,
		"escalated":      item.Escalated,
		"escalated_at":   item.EscalatedAt,
		"snoozed_until":  item.SnoozedUntil,
		"snoozed_count":  item.SnoozedCount,
	}
}

// Delete 删除待办事项
func (r *InboxRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return NewRepositoryErrorWithID("delete", "inbox_item", id, ErrInboxItemDeletionUnavailable)
}

func (r *InboxRepositoryImpl) DeleteByUserID(ctx context.Context, id, userID uint) error {
	return NewRepositoryErrorWithID("delete", "inbox_item", id, ErrInboxItemDeletionUnavailable)
}

// List 查询待办事项列表
func (r *InboxRepositoryImpl) List(ctx context.Context, params *InboxListParams) ([]*models.InboxItem, int64, error) {
	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	queryBuilder := NewQueryBuilder[models.InboxItem](r.db)

	// 添加用户ID过滤
	if params.UserID > 0 {
		queryBuilder = queryBuilder.Where("user_id = ?", params.UserID)
	}

	// 添加已读状态过滤
	if params.IsRead != nil {
		queryBuilder = queryBuilder.Where("is_read = ?", *params.IsRead)
	}

	// 添加完成状态过滤
	if params.IsCompleted != nil {
		queryBuilder = queryBuilder.Where("is_completed = ?", *params.IsCompleted)
	}

	// 添加优先级过滤
	if params.Priority != "" {
		queryBuilder = queryBuilder.Where("priority = ?", params.Priority)
	}

	// 添加来源类型过滤
	if params.SourceType != "" {
		queryBuilder = queryBuilder.Where("source_type = ?", params.SourceType)
	}

	// 添加到期时间范围过滤
	if params.DueBefore != nil {
		queryBuilder = queryBuilder.Where("due_date <= ?", *params.DueBefore)
	}
	if params.DueAfter != nil {
		queryBuilder = queryBuilder.Where("due_date >= ?", *params.DueAfter)
	}

	// 添加搜索条件
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		queryBuilder = queryBuilder.Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm)
	}

	// 排除延后的待办（如果延后时间未到）
	queryBuilder = queryBuilder.Where("(snoozed_until IS NULL OR snoozed_until <= NOW())")

	// 获取总数
	total, err := queryBuilder.Count(ctx)
	if err != nil {
		return nil, 0, NewRepositoryError("list", "inbox_item", err)
	}

	// 排序
	orderBy := params.OrderBy
	if orderBy == "" {
		orderBy = "due_date"
	}
	switch orderBy {
	case "due_date":
		queryBuilder = queryBuilder.Order("due_date ASC").OrderDesc("priority")
	case "priority":
		// 优先级排序: critical > high > medium > low
		queryBuilder = queryBuilder.Raw(`
		 FIELD(priority, 'critical', 'high', 'medium', 'low') ASC,
		 due_date ASC
		`)
	case "created_at":
		queryBuilder = queryBuilder.OrderDesc("created_at")
	default:
		queryBuilder = queryBuilder.Order(orderBy)
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	items, err := queryBuilder.Offset(offset).Limit(pageSize).Find(ctx)
	if err != nil {
		return nil, 0, NewRepositoryError("list", "inbox_item", err)
	}

	return items, total, nil
}

// FindByUserID 根据用户ID查询待办事项
func (r *InboxRepositoryImpl) FindByUserID(ctx context.Context, userID uint) ([]*models.InboxItem, error) {
	var items []*models.InboxItem
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("is_completed = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Order("priority DESC, due_date ASC").
		Find(&items).Error

	if err != nil {
		return nil, NewRepositoryError("find_by_user_id", "inbox_item", err)
	}

	return items, nil
}

// MarkAsRead 标记为已读
func (r *InboxRepositoryImpl) MarkAsRead(ctx context.Context, id uint) error {
	return r.markAsRead(ctx, r.db.WithContext(ctx).Where("id = ?", id), id)
}

func (r *InboxRepositoryImpl) MarkAsReadByUserID(ctx context.Context, id, userID uint) error {
	return r.markAsRead(ctx, r.ownedItemQuery(ctx, id, userID), id)
}

func (r *InboxRepositoryImpl) markAsRead(ctx context.Context, query *gorm.DB, id uint) error {
	now := time.Now()
	result := query.
		Model(&models.InboxItem{}).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("mark_as_read", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("mark_as_read", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

// MarkAsCompleted 标记为已完成
func (r *InboxRepositoryImpl) MarkAsCompleted(ctx context.Context, id uint) error {
	return r.markAsCompleted(ctx, r.db.WithContext(ctx).Where("id = ?", id), id)
}

func (r *InboxRepositoryImpl) MarkAsCompletedByUserID(ctx context.Context, id, userID uint) error {
	return r.markAsCompleted(ctx, r.ownedItemQuery(ctx, id, userID), id)
}

func (r *InboxRepositoryImpl) markAsCompleted(ctx context.Context, query *gorm.DB, id uint) error {
	now := time.Now()
	result := query.
		Model(&models.InboxItem{}).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": now,
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("mark_as_completed", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("mark_as_completed", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

// Snooze 延后待办事项
func (r *InboxRepositoryImpl) Snooze(ctx context.Context, id uint, until time.Time) error {
	return r.snooze(ctx, r.db.WithContext(ctx).Where("id = ?", id), id, until)
}

func (r *InboxRepositoryImpl) SnoozeByUserID(ctx context.Context, id, userID uint, until time.Time) error {
	return r.snooze(ctx, r.ownedItemQuery(ctx, id, userID), id, until)
}

func (r *InboxRepositoryImpl) snooze(ctx context.Context, query *gorm.DB, id uint, until time.Time) error {
	result := query.
		Model(&models.InboxItem{}).
		Updates(map[string]interface{}{
			"snoozed_until": until,
			"snoozed_count": gorm.Expr("snoozed_count + 1"),
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("snooze", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("snooze", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

func (r *InboxRepositoryImpl) ownedItemQuery(ctx context.Context, id, userID uint) *gorm.DB {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID)
}

// Escalate 升级待办事项
func (r *InboxRepositoryImpl) Escalate(ctx context.Context, id uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.InboxItem{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"escalated":    true,
			"escalated_at": now,
		})

	if result.Error != nil {
		return NewRepositoryErrorWithID("escalate", "inbox_item", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("escalate", "inbox_item", id, ErrInboxItemNotFound)
	}

	return nil
}

// GetUnreadCount 获取未读数量
func (r *InboxRepositoryImpl) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.InboxItem{}).
		Where("user_id = ?", userID).
		Where("is_read = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Count(&count).Error

	if err != nil {
		return 0, NewRepositoryError("get_unread_count", "inbox_item", err)
	}

	return count, nil
}

// GetDueItems 获取到期待办事项
func (r *InboxRepositoryImpl) GetDueItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error) {
	var items []*models.InboxItem
	err := r.db.WithContext(ctx).
		Where("is_completed = ?", false).
		Where("due_date <= ?", before).
		Where("reminder_sent = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Find(&items).Error

	if err != nil {
		return nil, NewRepositoryError("get_due_items", "inbox_item", err)
	}

	return items, nil
}

// GetOverdueCriticalItems 获取超时的critical待办
func (r *InboxRepositoryImpl) GetOverdueCriticalItems(ctx context.Context, before time.Time) ([]*models.InboxItem, error) {
	var items []*models.InboxItem
	err := r.db.WithContext(ctx).
		Where("is_completed = ?", false).
		Where("priority = ?", "critical").
		Where("due_date < ?", before).
		Where("escalated = ?", false).
		Where("(snoozed_until IS NULL OR snoozed_until <= NOW())").
		Find(&items).Error

	if err != nil {
		return nil, NewRepositoryError("get_overdue_critical_items", "inbox_item", err)
	}

	return items, nil
}

// CreateReminderRule 创建提醒规则
func (r *InboxRepositoryImpl) CreateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetReminderRules 获取提醒规则
func (r *InboxRepositoryImpl) GetReminderRules(ctx context.Context, isActive bool) ([]*models.InboxReminderRule, error) {
	var rules []*models.InboxReminderRule
	query := r.db.WithContext(ctx)
	if isActive {
		query = query.Where("is_active = ?", true)
	}
	err := query.Find(&rules).Error
	if err != nil {
		return nil, NewRepositoryError("get_reminder_rules", "inbox_reminder_rule", err)
	}
	return rules, nil
}

// GetReminderRuleByTypeAndPriority 根据日期类型和优先级获取规则
func (r *InboxRepositoryImpl) GetReminderRuleByTypeAndPriority(ctx context.Context, dueDateType, priority string) (*models.InboxReminderRule, error) {
	var rule models.InboxReminderRule
	err := r.db.WithContext(ctx).
		Where("due_date_type = ?", dueDateType).
		Where("priority = ?", priority).
		Where("is_active = ?", true).
		First(&rule).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, NewRepositoryError("get_reminder_rule", "inbox_reminder_rule", err)
	}

	return &rule, nil
}

// UpdateReminderRule 更新提醒规则
func (r *InboxRepositoryImpl) UpdateReminderRule(ctx context.Context, rule *models.InboxReminderRule) error {
	result := r.db.WithContext(ctx).
		Model(&models.InboxReminderRule{}).
		Where("id = ?", rule.ID).
		Updates(rule)

	if result.Error != nil {
		return NewRepositoryErrorWithID("update_reminder_rule", "inbox_reminder_rule", rule.ID, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("update_reminder_rule", "inbox_reminder_rule", rule.ID, ErrReminderRuleNotFound)
	}

	return nil
}

// DeleteReminderRule 删除提醒规则
func (r *InboxRepositoryImpl) DeleteReminderRule(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Delete(&models.InboxReminderRule{}, id)

	if result.Error != nil {
		return NewRepositoryErrorWithID("delete_reminder_rule", "inbox_reminder_rule", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return NewRepositoryErrorWithID("delete_reminder_rule", "inbox_reminder_rule", id, ErrReminderRuleNotFound)
	}

	return nil
}
