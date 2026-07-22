package services

import (
	"context"
	"errors"
	"fmt"
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
	UserID      uint    `json:"user_id" binding:"required"`
	SourceType  string  `json:"source_type" binding:"required,oneof=deadline approval task handoff waiver conflict"`
	SourceID    uint    `json:"source_id" binding:"required"`
	Title       string  `json:"title" binding:"required,min=1,max=255"`
	Content     string  `json:"content" binding:"max=5000"`
	Priority    string  `json:"priority" binding:"required,oneof=critical high medium low"`
	DueDate     *string `json:"due_date"`
	DueDateType string  `json:"due_date_type" binding:"omitempty,max=50"`
}

// UpdateInboxItemRequest 更新待办事项请求
type UpdateInboxItemRequest struct {
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
		UserID:      userID,
		SourceType:  req.SourceType,
		SourceID:    req.SourceID,
		Title:       req.Title,
		Content:     req.Content,
		Priority:    req.Priority,
		DueDate:     dueDate,
		DueDateType: req.DueDateType,
	}

	err = s.inboxRepo.Create(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("创建待办事项失败: %w", err)
	}

	return s.convertToResponse(item), nil
}

// GetInboxItemByID 根据ID获取待办事项
func (s *InboxService) GetInboxItemByID(ctx context.Context, id, userID uint) (*InboxItemResponse, error) {
	item, err := s.inboxRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("获取待办事项失败: %w", err)
	}
	if item == nil {
		return nil, errors.New("待办事项不存在")
	}

	return s.convertToResponse(item), nil
}

// UpdateInboxItem 更新待办事项
func (s *InboxService) UpdateInboxItem(ctx context.Context, id, userID uint, req *UpdateInboxItemRequest) (*InboxItemResponse, error) {
	// 获取现有待办事项
	item, err := s.inboxRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("获取待办事项失败: %w", err)
	}
	if item == nil {
		return nil, errors.New("待办事项不存在")
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

	return s.convertToResponse(item), nil
}

// DeleteInboxItem 删除待办事项
func (s *InboxService) DeleteInboxItem(ctx context.Context, id, userID uint) error {
	// Preserve the same ownership/not-found response used by every other
	// inbox operation before exposing the retention policy.
	if _, err := s.inboxRepo.FindByIDAndUserID(ctx, id, userID); err != nil {
		return err
	}
	return s.inboxRepo.DeleteByUserID(ctx, id, userID)
}

// ListInboxItems 获取待办事项列表
func (s *InboxService) ListInboxItems(ctx context.Context, userID uint, req *ListInboxItemsRequest) (*ListInboxItemsResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 解析日期范围
	var dueBefore, dueAfter *time.Time
	if req.DueBefore != "" {
		if t, err := time.Parse("2006-01-02", req.DueBefore); err == nil {
			dueBefore = &t
		}
	}
	if req.DueAfter != "" {
		if t, err := time.Parse("2006-01-02", req.DueAfter); err == nil {
			dueAfter = &t
		}
	}

	// 构建查询参数
	params := repositories.InboxListParams{
		Page:        req.Page,
		PageSize:    req.PageSize,
		UserID:      userID,
		IsRead:      req.IsRead,
		IsCompleted: req.IsCompleted,
		Priority:    req.Priority,
		SourceType:  req.SourceType,
		DueBefore:   dueBefore,
		DueAfter:    dueAfter,
		Search:      req.Search,
		OrderBy:     req.OrderBy,
	}

	// 获取待办事项列表
	items, total, err := s.inboxRepo.List(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("获取待办事项列表失败: %w", err)
	}

	// 转换为响应格式
	itemResponses := make([]*InboxItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = s.convertToResponse(item)
	}

	// 构建分页信息
	pagination := PaginationWithTotalPage{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Total:     total,
		TotalPage: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	return &ListInboxItemsResponse{
		Items:      itemResponses,
		Pagination: pagination,
	}, nil
}

// MarkAsRead 标记为已读
func (s *InboxService) MarkAsRead(ctx context.Context, id, userID uint) error {
	return s.inboxRepo.MarkAsReadByUserID(ctx, id, userID)
}

// MarkAsCompleted 标记为已完成
func (s *InboxService) MarkAsCompleted(ctx context.Context, id, userID uint) error {
	return s.inboxRepo.MarkAsCompletedByUserID(ctx, id, userID)
}

// SnoozeInboxItem 延后待办事项
func (s *InboxService) SnoozeInboxItem(ctx context.Context, id, userID uint, req *SnoozeInboxItemRequest) error {
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
	unread, err := s.inboxRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取未读数量失败: %w", err)
	}

	// 获取所有待办事项
	allItems, total, err := s.inboxRepo.List(ctx, &repositories.InboxListParams{
		Page:     1,
		PageSize: 1000,
		UserID:   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("获取待办事项列表失败: %w", err)
	}

	stats := &InboxStatsResponse{
		Total:    total,
		Unread:   unread,
		Pending:  0,
		Critical: 0,
		High:     0,
		Overdue:  0,
		DueToday: 0,
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekEnd := today.AddDate(0, 0, 7)

	for _, item := range allItems {
		if !item.IsCompleted {
			stats.Pending++
		} else {
			stats.Completed++
		}

		switch item.Priority {
		case "critical":
			stats.Critical++
		case "high":
			stats.High++
		}

		if item.DueDate != nil && !item.IsCompleted {
			if item.DueDate.Before(now) {
				stats.Overdue++
			}
			if item.DueDate.After(today) && item.DueDate.Before(weekEnd) {
				stats.DueThisWeek++
			}
			if item.DueDate.Year() == today.Year() && item.DueDate.YearDay() == today.YearDay() {
				stats.DueToday++
			}
		}
	}

	return stats, nil
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
