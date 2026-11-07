package services

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// InboxService 待办事项服务
type InboxService struct {
	inboxRepo repositories.InboxRepository
	userRepo  repositories.UserRepository
}

// NewInboxService 创建待办事项服务
func NewInboxService(inboxRepo repositories.InboxRepository, userRepo repositories.UserRepository) *InboxService {
	return &InboxService{
		inboxRepo: inboxRepo,
		userRepo:  userRepo,
	}
}

// CreateInboxItemRequest 创建待办事项请求
type CreateInboxItemRequest struct {
	CompletionMode  json.RawMessage `json:"completion_mode"`
	SourceRef       json.RawMessage `json:"source_ref"`
	WorkflowTaskKey json.RawMessage `json:"workflow_task_key"`
	AssignedTo      json.RawMessage `json:"assigned_to"`
	WorkflowGroup   json.RawMessage `json:"workflow_group"`
	WorkflowPurpose json.RawMessage `json:"workflow_purpose"`
	WorkflowVersion json.RawMessage `json:"workflow_version"`

	UserID      uint    `json:"user_id"`
	SourceType  string  `json:"source_type" binding:"omitempty,oneof=deadline approval task handoff waiver conflict"`
	SourceID    uint    `json:"source_id"`
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Content     string  `json:"content" binding:"max=5000"`
	Priority    string  `json:"priority" binding:"required,oneof=critical high medium low"`
	DueDate     *string `json:"due_date"`
	DueDateType string  `json:"due_date_type" binding:"omitempty,max=50"`
}

// UpdateInboxItemRequest 更新待办事项请求
type UpdateInboxItemRequest struct {
	CompletionMode  json.RawMessage `json:"completion_mode"`
	SourceRef       json.RawMessage `json:"source_ref"`
	WorkflowTaskKey json.RawMessage `json:"workflow_task_key"`
	AssignedTo      json.RawMessage `json:"assigned_to"`
	WorkflowGroup   json.RawMessage `json:"workflow_group"`
	WorkflowPurpose json.RawMessage `json:"workflow_purpose"`
	WorkflowVersion json.RawMessage `json:"workflow_version"`

	UserID      json.RawMessage `json:"user_id"`
	SourceType  json.RawMessage `json:"source_type"`
	SourceID    json.RawMessage `json:"source_id"`
	IsCompleted json.RawMessage `json:"is_completed"`
	ArchivedAt  json.RawMessage `json:"archived_at"`

	Title       string  `json:"title" binding:"omitempty,min=1,max=255"`
	Content     string  `json:"content" binding:"omitempty,max=5000"`
	Priority    string  `json:"priority" binding:"omitempty,oneof=critical high medium low"`
	DueDate     *string `json:"due_date"`
	DueDateType string  `json:"due_date_type" binding:"omitempty,max=50"`
}

// SnoozeInboxItemRequest 延后待办事项请求
type SnoozeInboxItemRequest struct {
	Until    string `json:"until" binding:"required"` // ISO 8601 格式时间
	Duration int    `json:"duration"`                 // 延后天数，可选
}

// ListInboxItemsRequest 待办事项列表请求
type ListInboxItemsRequest struct {
	WorkflowBucket string `json:"workflow_bucket" form:"workflow_bucket" binding:"omitempty,oneof=needs_action waiting completed"`
	Archived       *bool  `json:"archived" form:"archived"`

	Page        int    `json:"page" form:"page" binding:"min=1"`
	PageSize    int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	IsRead      *bool  `json:"is_read" form:"is_read"`
	IsCompleted *bool  `json:"is_completed" form:"is_completed"`
	Priority    string `json:"priority" form:"priority" binding:"omitempty,oneof=critical high medium low"`
	SourceType  string `json:"source_type" form:"source_type"`
	DueBefore   string `json:"due_before" form:"due_before"`
	DueAfter    string `json:"due_after" form:"due_after"`
	Search      string `json:"search" form:"search"`
	OrderBy     string `json:"order_by" form:"order_by" binding:"omitempty,oneof=due_date priority created_at"`
}

// InboxItemResponse 待办事项响应
type InboxItemResponse struct {
	SourceRef      string          `json:"source_ref,omitempty"`
	CompletionMode string          `json:"completion_mode"`
	WorkflowAction *WorkflowAction `json:"workflow_action"`
	WorkflowBucket string          `json:"workflow_bucket"`
	ClosedReason   string          `json:"closed_reason,omitempty"`
	ArchivedAt     *time.Time      `json:"archived_at,omitempty"`

	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	SourceType    string     `json:"source_type"`
	SourceID      uint       `json:"source_id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Priority      string     `json:"priority"`
	DueDate       *time.Time `json:"due_date"`
	DueDateType   string     `json:"due_date_type"`
	IsRead        bool       `json:"is_read"`
	ReadAt        *time.Time `json:"read_at"`
	IsCompleted   bool       `json:"is_completed"`
	CompletedAt   *time.Time `json:"completed_at"`
	ReminderSent  bool       `json:"reminder_sent"`
	ReminderCount int        `json:"reminder_count"`
	Escalated     bool       `json:"escalated"`
	EscalatedAt   *time.Time `json:"escalated_at"`
	SnoozedUntil  *time.Time `json:"snoozed_until"`
	SnoozedCount  int        `json:"snoozed_count"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ListInboxItemsResponse 待办事项列表响应
type ListInboxItemsResponse struct {
	Items      []*InboxItemResponse    `json:"items"`
	Pagination PaginationWithTotalPage `json:"pagination"`
	Stats      *InboxStatsResponse     `json:"stats,omitempty"`
}

// InboxStatsResponse 待办事项统计响应
type InboxStatsResponse struct {
	NeedsAction      int64 `json:"needs_action"`
	Waiting          int64 `json:"waiting"`
	Archived         int64 `json:"archived"`
	LegacyUnresolved int64 `json:"legacy_unresolved"`

	Total       int64 `json:"total"`
	Unread      int64 `json:"unread"`
	Pending     int64 `json:"pending"`
	Completed   int64 `json:"completed"`
	Critical    int64 `json:"critical"`
	High        int64 `json:"high"`
	Overdue     int64 `json:"overdue"`
	DueToday    int64 `json:"due_today"`
	DueThisWeek int64 `json:"due_this_week"`
}

// CreateInboxItem 创建待办事项
func (s *InboxService) CreateInboxItem(ctx context.Context, userID uint, req *CreateInboxItemRequest) (*InboxItemResponse, error) {
	// The existing trusted client-handoff producer calls this service directly.
	// Its numeric source is kept as an unresolved reminder, never upgraded to a
	// business capability. The public HTTP endpoint only permits personal tasks.
	legacyHandoff := req != nil && req.SourceType == "handoff" && req.SourceID > 0
	if req == nil || req.UserID != 0 && req.UserID != userID || !legacyHandoff && (req.SourceID != 0 || req.SourceType != "" && req.SourceType != "task") || hasInboxProtected(req.CompletionMode, req.SourceRef, req.WorkflowTaskKey, req.AssignedTo, req.WorkflowGroup, req.WorkflowPurpose, req.WorkflowVersion) {
		return nil, ErrInboxProtectedField
	}
	if userID == 0 || s.userRepo == nil {
		return nil, fmt.Errorf("待办接收人验证服务不可用")
	}
	// 验证用户是否存在
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在: %w", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 解析到期日期
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, *req.DueDate)
		if err != nil {
			// 尝试其他格式
			parsedDate, err = time.Parse("2006-01-02T15:04:05", *req.DueDate)
			if err != nil {
				parsedDate, err = time.Parse("2006-01-02", *req.DueDate)
				if err != nil {
					return nil, fmt.Errorf("到期日期格式错误: %w", err)
				}
				// 如果是日期格式，设置为当天结束时间
				parsedDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 0, time.Local)
			}
		}
		dueDate = &parsedDate
	}

	// 创建待办事项
	item := &models.InboxItem{
		UserID:         userID,
		SourceType:     "task",
		CompletionMode: "manual",
		Title:          req.Title,
		Content:        req.Content,
		Priority:       req.Priority,
		DueDate:        dueDate,
		DueDateType:    req.DueDateType,
	}

	if legacyHandoff {
		item.SourceType, item.SourceID, item.CompletionMode = "handoff", req.SourceID, "legacy_unresolved"
	}
	err = s.inboxRepo.Create(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("创建待办事项失败: %w", err)
	}

	return s.projectOneInbox(ctx, item, userID)
}

// GetInboxItemByID 根据ID获取待办事项
func (s *InboxService) GetInboxItemByID(ctx context.Context, id, userID uint) (*InboxItemResponse, error) {
	item, err := s.inboxRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	return s.projectOneInbox(ctx, item, userID)
}

// UpdateInboxItem 更新待办事项
func (s *InboxService) UpdateInboxItem(ctx context.Context, id, userID uint, req *UpdateInboxItemRequest) (*InboxItemResponse, error) {
	if req == nil || hasInboxProtected(req.CompletionMode, req.SourceRef, req.WorkflowTaskKey, req.AssignedTo, req.WorkflowGroup, req.WorkflowPurpose, req.WorkflowVersion, req.UserID, req.SourceType, req.SourceID, req.IsCompleted, req.ArchivedAt) {
		return nil, ErrInboxProtectedField
	}

	// 获取现有待办事项
	item, err := s.inboxRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("获取待办事项失败: %w", err)
	}
	if item == nil {
		return nil, errors.New("待办事项不存在")
	}

	if _, err := s.projectOneInbox(ctx, item, userID); err != nil {
		return nil, err
	}
	if !repositories.IsManualInboxItem(item) {
		return nil, repositories.ErrInboxWorkflowActionRequired
	}

	// 解析到期日期
	if req.DueDate != nil {
		if *req.DueDate == "" {
			item.DueDate = nil
		} else {
			parsedDate, err := time.Parse(time.RFC3339, *req.DueDate)
			if err != nil {
				// 尝试其他格式
				parsedDate, err = time.Parse("2006-01-02T15:04:05", *req.DueDate)
				if err != nil {
					parsedDate, err = time.Parse("2006-01-02", *req.DueDate)
					if err != nil {
						return nil, fmt.Errorf("到期日期格式错误: %w", err)
					}
					parsedDate = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 0, time.Local)
				}
			}
			item.DueDate = &parsedDate
		}
	}

	// 更新字段
	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Content != "" {
		item.Content = req.Content
	}
	if req.Priority != "" {
		item.Priority = req.Priority
	}
	if req.DueDateType != "" {
		item.DueDateType = req.DueDateType
	}

	// 保存更新
	err = s.inboxRepo.UpdateByUserID(ctx, item, userID)
	if err != nil {
		return nil, fmt.Errorf("更新待办事项失败: %w", err)
	}

	return s.projectOneInbox(ctx, item, userID)
}

// DeleteInboxItem 删除待办事项
func (s *InboxService) DeleteInboxItem(ctx context.Context, id, userID uint) error {
	// Preserve the same ownership/not-found response used by every other
	// inbox operation before exposing the retention policy.
	if _, err := s.GetInboxItemByID(ctx, id, userID); err != nil {
		return err
	}
	return s.inboxRepo.DeleteByUserID(ctx, id, userID)
}

// ListInboxItems 获取待办事项列表
func (s *InboxService) ListInboxItems(ctx context.Context, userID uint, req *ListInboxItemsRequest) (*ListInboxItemsResponse, error) {
	if req == nil {
		req = &ListInboxItemsRequest{}
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	visible, err := s.visibleInbox(ctx, userID)
	if err != nil {
		return nil, err
	}
	stats := inboxStats(visible)
	var dueBefore, dueAfter *time.Time
	if req.DueBefore != "" {
		at, err := parseInboxBoundary(req.DueBefore)
		if err != nil {
			return nil, err
		}
		dueBefore = &at
	}
	if req.DueAfter != "" {
		at, err := parseInboxBoundary(req.DueAfter)
		if err != nil {
			return nil, err
		}
		dueAfter = &at
	}
	filtered := make([]*InboxItemResponse, 0, len(visible))
	for _, item := range visible {
		archived := req.Archived != nil && *req.Archived
		if (item.ArchivedAt != nil) != archived {
			continue
		}
		if req.WorkflowBucket != "" && item.WorkflowBucket != req.WorkflowBucket {
			continue
		}
		if req.IsRead != nil && item.IsRead != *req.IsRead {
			continue
		}
		if req.IsCompleted != nil && item.IsCompleted != *req.IsCompleted {
			continue
		}
		if req.Priority != "" && item.Priority != req.Priority {
			continue
		}
		if req.SourceType != "" && item.SourceType != req.SourceType {
			continue
		}
		if req.Search != "" && !strings.Contains(strings.ToLower(item.Title+" "+item.Content), strings.ToLower(req.Search)) {
			continue
		}
		if dueBefore != nil && (item.DueDate == nil || item.DueDate.After(*dueBefore)) {
			continue
		}
		if dueAfter != nil && (item.DueDate == nil || item.DueDate.Before(*dueAfter)) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		if req.OrderBy == "priority" && inboxPriority(a.Priority) != inboxPriority(b.Priority) {
			return inboxPriority(a.Priority) > inboxPriority(b.Priority)
		}
		if req.OrderBy != "created_at" {
			if a.DueDate == nil && b.DueDate != nil {
				return false
			}
			if a.DueDate != nil && b.DueDate == nil {
				return true
			}
			if a.DueDate != nil && b.DueDate != nil && !a.DueDate.Equal(*b.DueDate) {
				return a.DueDate.Before(*b.DueDate)
			}
		}
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.After(b.CreatedAt)
		}
		return a.ID > b.ID
	})
	total := int64(len(filtered))
	start := int64(req.Page-1) * int64(req.PageSize)
	page := []*InboxItemResponse{}
	if start >= 0 && start < total {
		end := start + int64(req.PageSize)
		if end > total {
			end = total
		}
		page = filtered[start:end]
	}
	return &ListInboxItemsResponse{Items: page, Stats: stats, Pagination: PaginationWithTotalPage{Page: req.Page, PageSize: req.PageSize, Total: total, TotalPage: (total + int64(req.PageSize) - 1) / int64(req.PageSize)}}, nil
}

// MarkAsRead 标记为已读
func (s *InboxService) MarkAsRead(ctx context.Context, id, userID uint) error {
	if _, err := s.GetInboxItemByID(ctx, id, userID); err != nil {
		return err
	}
	return s.inboxRepo.MarkAsReadByUserID(ctx, id, userID)
}

// MarkAsCompleted 标记为已完成
func (s *InboxService) MarkAsCompleted(ctx context.Context, id, userID uint) error {
	if _, err := s.GetInboxItemByID(ctx, id, userID); err != nil {
		return err
	}
	return s.inboxRepo.MarkAsCompletedByUserID(ctx, id, userID)
}

// SnoozeInboxItem 延后待办事项
func (s *InboxService) SnoozeInboxItem(ctx context.Context, id, userID uint, req *SnoozeInboxItemRequest) error {
	if _, err := s.GetInboxItemByID(ctx, id, userID); err != nil {
		return err
	}
	var until time.Time
	var err error

	// 解析时间
	if req.Until != "" {
		until, err = time.Parse(time.RFC3339, req.Until)
		if err != nil {
			until, err = time.Parse("2006-01-02T15:04:05", req.Until)
			if err != nil {
				until, err = time.Parse("2006-01-02", req.Until)
				if err != nil {
					return fmt.Errorf("时间格式错误: %w", err)
				}
				until = time.Date(until.Year(), until.Month(), until.Day(), 9, 0, 0, 0, time.Local)
			}
		}
	} else if req.Duration > 0 {
		until = time.Now().AddDate(0, 0, req.Duration)
	} else {
		return errors.New("必须指定时间或延后天数")
	}

	return s.inboxRepo.SnoozeByUserID(ctx, id, userID, until)
}

// GetInboxStats 获取待办事项统计
func (s *InboxService) GetInboxStats(ctx context.Context, userID uint) (*InboxStatsResponse, error) {
	items, err := s.visibleInbox(ctx, userID)
	if err != nil {
		return nil, err
	}
	return inboxStats(items), nil
}

func inboxStats(items []*InboxItemResponse) *InboxStatsResponse {
	stats := &InboxStatsResponse{}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekEnd := today.AddDate(0, 0, 7)
	for _, item := range items {
		if item.ArchivedAt != nil {
			stats.Archived++
			continue
		}
		stats.Total++
		if !item.IsRead {
			stats.Unread++
		}
		if item.CompletionMode == "legacy_unresolved" {
			stats.LegacyUnresolved++
			continue
		}
		if item.IsCompleted {
			stats.Completed++
		} else {
			stats.Pending++
		}
		if item.WorkflowBucket == "needs_action" {
			stats.NeedsAction++
		}
		if item.WorkflowBucket == "waiting" {
			stats.Waiting++
		}
		if item.Priority == "critical" {
			stats.Critical++
		}
		if item.Priority == "high" {
			stats.High++
		}
		if item.DueDate != nil && !item.IsCompleted {
			if item.DueDate.Before(now) {
				stats.Overdue++
			}
			if !item.DueDate.Before(today) && item.DueDate.Before(weekEnd) {
				stats.DueThisWeek++
			}
			if !item.DueDate.Before(today) && item.DueDate.Before(today.AddDate(0, 0, 1)) {
				stats.DueToday++
			}
		}
	}
	return stats
}

var ErrInboxProtectedField = errors.New("待办来源、模式、指派及业务状态由服务器管理")

func hasInboxProtected(fields ...json.RawMessage) bool {
	for _, field := range fields {
		if len(field) > 0 {
			return true
		}
	}
	return false
}

func parseInboxBoundary(raw string) (time.Time, error) {
	if at, err := time.Parse(time.RFC3339, raw); err == nil {
		return at, nil
	}
	return time.Parse("2006-01-02", raw)
}

func inboxPriority(priority string) int {
	switch priority {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

type inboxProjectionRepository interface {
	InboxDB() *gorm.DB
	ListOwnedForProjection(context.Context, uint) ([]*models.InboxItem, error)
	ArchiveByUserID(context.Context, uint, uint) error
}

func (s *InboxService) visibleInbox(ctx context.Context, userID uint) ([]*InboxItemResponse, error) {
	repo, ok := s.inboxRepo.(inboxProjectionRepository)
	if !ok {
		return nil, fmt.Errorf("待办授权查询不可用")
	}
	items, err := repo.ListOwnedForProjection(ctx, userID)
	if err != nil {
		return nil, err
	}
	actions, err := projectWorkflowInbox(ctx, repo.InboxDB(), userID, items)
	if err != nil {
		return nil, err
	}
	responses := make([]*InboxItemResponse, 0, len(items))
	for _, item := range items {
		if item.CompletionMode == "workflow" && actions[item.ID] == nil {
			continue
		}
		responses = append(responses, s.inboxResponse(item, actions[item.ID]))
	}
	return responses, nil
}

func (s *InboxService) projectOneInbox(ctx context.Context, item *models.InboxItem, userID uint) (*InboxItemResponse, error) {
	if item == nil || userID == 0 || item.UserID != userID {
		return nil, repositories.ErrInboxItemNotFound
	}
	var db *gorm.DB
	if repo, ok := s.inboxRepo.(inboxProjectionRepository); ok {
		db = repo.InboxDB()
	}
	actions, err := projectWorkflowInbox(ctx, db, userID, []*models.InboxItem{item})
	if err != nil {
		return nil, err
	}
	if item.CompletionMode == "workflow" && actions[item.ID] == nil {
		return nil, repositories.ErrInboxItemNotFound
	}
	return s.inboxResponse(item, actions[item.ID]), nil
}

func (s *InboxService) inboxResponse(item *models.InboxItem, action *WorkflowAction) *InboxItemResponse {
	r := s.convertToResponse(item)
	r.ArchivedAt = item.ArchivedAt
	r.ClosedReason = item.ClosedReason
	if item.CompletionMode == "workflow" {
		r.CompletionMode, r.SourceRef, r.WorkflowAction = "workflow", item.SourceRef, action
		r.WorkflowBucket = "waiting"
		if action != nil && action.ActionCode != "view_only" {
			r.WorkflowBucket = "needs_action"
		}
		if item.IsCompleted {
			r.WorkflowBucket = "completed"
		}
	} else if repositories.IsManualInboxItem(item) {
		r.CompletionMode = "manual"
		r.WorkflowBucket = "needs_action"
		status := "pending"
		if item.IsCompleted {
			r.WorkflowBucket, status = "completed", "completed"
		}
		r.WorkflowAction = &WorkflowAction{ActionCode: "view_only", ActionLabel: "个人提醒", WorkflowStatus: status, CompletionMode: "manual"}
	} else {
		r.CompletionMode = "legacy_unresolved"
		r.Title = "历史事项，关联对象待确认"
		r.Content = ""
		r.DueDateType = ""
		r.SourceID = 0
		r.SourceRef = ""
		r.IsCompleted = false
		r.CompletedAt = nil
		r.ClosedReason = ""
		r.WorkflowBucket = "legacy_unresolved"
		r.WorkflowAction = &WorkflowAction{ActionCode: "view_only", ActionLabel: "归档提醒", WorkflowStatus: "legacy_unresolved", CompletionMode: "legacy_unresolved", BlockedCode: "LEGACY_SOURCE_UNRESOLVED", BlockedReason: "历史关联无法确认，归档只收起提醒，不改变业务状态"}
	}
	return r
}

func (s *InboxService) ArchiveInboxItem(ctx context.Context, id, userID uint) error {
	item, err := s.GetInboxItemByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if item.CompletionMode == "workflow" && !item.IsCompleted {
		return repositories.ErrInboxWorkflowActionRequired
	}
	repo, ok := s.inboxRepo.(inboxProjectionRepository)
	if !ok {
		return fmt.Errorf("待办归档服务不可用")
	}
	return repo.ArchiveByUserID(ctx, id, userID)
}

// convertToResponse 转换为响应格式
func (s *InboxService) convertToResponse(item *models.InboxItem) *InboxItemResponse {
	return &InboxItemResponse{
		ID:            item.ID,
		UserID:        item.UserID,
		SourceType:    item.SourceType,
		SourceID:      item.SourceID,
		Title:         item.Title,
		Content:       item.Content,
		Priority:      item.Priority,
		DueDate:       item.DueDate,
		DueDateType:   item.DueDateType,
		IsRead:        item.IsRead,
		ReadAt:        item.ReadAt,
		IsCompleted:   item.IsCompleted,
		CompletedAt:   item.CompletedAt,
		ReminderSent:  item.ReminderSent,
		ReminderCount: item.ReminderCount,
		Escalated:     item.Escalated,
		EscalatedAt:   item.EscalatedAt,
		SnoozedUntil:  item.SnoozedUntil,
		SnoozedCount:  item.SnoozedCount,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}
