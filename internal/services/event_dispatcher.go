package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// EventType 事件类型
type EventType string

const (
	// 案件相关事件
	EventCaseCreated          EventType = "case.created"
	EventCaseUpdated          EventType = "case.updated"
	EventCaseDeleted          EventType = "case.deleted"
	EventCaseStatusChanged    EventType = "case.status_changed"
	EventCaseHearingScheduled EventType = "case.hearing_scheduled"
	EventCaseJudgmentReceived EventType = "case.judgment_received"

	// 审批相关事件
	EventApprovalCreated   EventType = "approval.created"
	EventApprovalApproved  EventType = "approval.approved"
	EventApprovalRejected  EventType = "approval.rejected"
	EventApprovalSubmitted EventType = "approval.submitted"

	// 客户相关事件
	EventClientCreated EventType = "client.created"
	EventClientUpdated EventType = "client.updated"

	// 文档相关事件
	EventDocumentUploaded EventType = "document.uploaded"
	EventDocumentUpdated  EventType = "document.updated"

	// 自定义事件
	EventCustomReminder EventType = "custom.reminder"
)

// Event 事件定义
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	SourceID  uint                   `json:"source_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// EventHandler 事件处理器函数
type EventHandler func(ctx context.Context, event *Event) error

// EventDispatcher 事件分发器
type EventDispatcher struct {
	handlers  map[EventType][]EventHandler
	mu        sync.RWMutex
	inboxRepo repositories.InboxRepository
	userRepo  repositories.UserRepository
	caseRepo  repositories.CaseRepository
}

// NewEventDispatcher 创建事件分发器
func NewEventDispatcher(inboxRepo repositories.InboxRepository, userRepo repositories.UserRepository, caseRepo repositories.CaseRepository) *EventDispatcher {
	dispatcher := &EventDispatcher{
		handlers:  make(map[EventType][]EventHandler),
		inboxRepo: inboxRepo,
		userRepo:  userRepo,
		caseRepo:  caseRepo,
	}

	// 注册默认事件处理器
	dispatcher.registerDefaultHandlers()

	return dispatcher
}

// RegisterHandler 注册事件处理器
func (d *EventDispatcher) RegisterHandler(eventType EventType, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Dispatch 分发事件
func (d *EventDispatcher) Dispatch(ctx context.Context, event *Event) error {
	d.mu.RLock()
	handlers, exists := d.handlers[event.Type]
	d.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return nil
	}

	// 复制处理器列表避免并发问题
	handlerCopy := make([]EventHandler, len(handlers))
	copy(handlerCopy, handlers)

	// 异步执行所有处理器
	for _, handler := range handlerCopy {
		go func(h EventHandler) {
			defer func() {
				if recovered := recover(); recovered != nil {
					fmt.Printf("事件处理器panic: type=%v, panic=%v\n", event.Type, recovered)
				}
			}()
			if err := h(ctx, event); err != nil {
				// 记录错误但不中断其他处理器
				fmt.Printf("事件处理器错误: type=%v, error=%v\n", event.Type, err)
			}
		}(handler)
	}

	return nil
}

// DispatchSync 同步分发事件（用于测试）
func (d *EventDispatcher) DispatchSync(ctx context.Context, event *Event) error {
	d.mu.RLock()
	handlers, exists := d.handlers[event.Type]
	d.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return nil
	}

	// 同步执行所有处理器
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("事件处理器错误: %w", err)
		}
	}

	return nil
}

// registerDefaultHandlers 注册默认事件处理器
func (d *EventDispatcher) registerDefaultHandlers() {
	// 案件创建事件
	d.RegisterHandler(EventCaseCreated, d.handleCaseCreated)

	// 案件状态变更事件
	d.RegisterHandler(EventCaseStatusChanged, d.handleCaseStatusChanged)

	// 审批创建事件
	d.RegisterHandler(EventApprovalCreated, d.handleApprovalCreated)

	// 审批通过事件
	d.RegisterHandler(EventApprovalApproved, d.handleApprovalApproved)

	// 庭审安排事件
	d.RegisterHandler(EventCaseHearingScheduled, d.handleHearingScheduled)

	// 判决收到事件
	d.RegisterHandler(EventCaseJudgmentReceived, d.handleJudgmentReceived)
}

// handleCaseCreated 处理案件创建事件
func (d *EventDispatcher) handleCaseCreated(ctx context.Context, event *Event) error {
	caseID := event.SourceID
	case_, err := d.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("获取案件失败: %w", err)
	}
	if case_ == nil {
		return nil
	}

	// 为主办律师创建待办事项
	item := &models.InboxItem{
		UserID:      case_.LawyerID,
		SourceType:  "task",
		SourceID:    caseID,
		Title:       fmt.Sprintf("新案件: %s", case_.Title),
		Content:     fmt.Sprintf("案件类型: %s, 优先级: %s", case_.CaseType, case_.Priority),
		Priority:    case_.Priority,
		DueDateType: "case_intake",
	}

	if err := d.inboxRepo.Create(ctx, item); err != nil {
		return fmt.Errorf("创建待办事项失败: %w", err)
	}

	return nil
}

// handleCaseStatusChanged 处理案件状态变更事件
func (d *EventDispatcher) handleCaseStatusChanged(ctx context.Context, event *Event) error {
	caseID := event.SourceID
	status, _ := event.Metadata["status"].(string)

	case_, err := d.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("获取案件失败: %w", err)
	}
	if case_ == nil {
		return nil
	}

	// 根据状态创建相应的待办事项
	switch status {
	case "active":
		item := &models.InboxItem{
			UserID:      case_.LawyerID,
			SourceType:  "task",
			SourceID:    caseID,
			Title:       fmt.Sprintf("案件已激活: %s", case_.Title),
			Content:     "案件已激活，请尽快开始处理",
			Priority:    "medium",
			DueDateType: "case_active",
		}
		return d.inboxRepo.Create(ctx, item)
	case "closed":
		item := &models.InboxItem{
			UserID:      case_.LawyerID,
			SourceType:  "task",
			SourceID:    caseID,
			Title:       fmt.Sprintf("案件已结案: %s", case_.Title),
			Content:     "案件已结案，请进行归档处理",
			Priority:    "normal",
			DueDateType: "case_closing",
		}
		return d.inboxRepo.Create(ctx, item)
	}

	return nil
}

// handleApprovalCreated 处理审批创建事件
func (d *EventDispatcher) handleApprovalCreated(ctx context.Context, event *Event) error {
	approvalID := event.SourceID
	title, _ := event.Metadata["title"].(string)
	submitterID, _ := event.Metadata["submitter_id"].(uint)

	// 获取审批人列表
	approverIDs, ok := event.Metadata["approver_ids"].([]uint)
	if !ok || len(approverIDs) == 0 {
		return nil
	}

	// 为每个审批人创建待办事项
	for _, approverID := range approverIDs {
		item := &models.InboxItem{
			UserID:      approverID,
			SourceType:  "approval",
			SourceID:    approvalID,
			Title:       fmt.Sprintf("待审批: %s", title),
			Content:     fmt.Sprintf("用户ID: %d 提交的审批申请", submitterID),
			Priority:    "high",
			DueDateType: "approval",
		}
		_ = d.inboxRepo.Create(ctx, item)
	}
	return nil
}

// handleApprovalApproved 处理审批通过事件
func (d *EventDispatcher) handleApprovalApproved(ctx context.Context, event *Event) error {
	submitterID, _ := event.Metadata["submitter_id"].(uint)
	title, _ := event.Metadata["title"].(string)
	approvalID := event.SourceID

	item := &models.InboxItem{
		UserID:      submitterID,
		SourceType:  "approval",
		SourceID:    approvalID,
		Title:       fmt.Sprintf("审批已通过: %s", title),
		Content:     "您的审批申请已通过",
		Priority:    "normal",
		DueDateType: "approval_result",
	}
	_ = d.inboxRepo.Create(ctx, item)
	return nil
}

// handleHearingScheduled 处理庭审安排事件
func (d *EventDispatcher) handleHearingScheduled(ctx context.Context, event *Event) error {
	caseID := event.SourceID
	hearingDate, _ := event.Metadata["hearing_date"].(string)
	court, _ := event.Metadata["court"].(string)

	case_, err := d.caseRepo.FindByID(ctx, caseID)
	if err != nil || case_ == nil {
		return nil
	}

	// 解析庭审日期
	parsedDate, err := time.Parse(time.RFC3339, hearingDate)
	if err != nil {
		return nil
	}

	// 提前7天创建提醒
	reminderDate := parsedDate.AddDate(0, 0, -7)
	now := time.Now()

	item := &models.InboxItem{
		UserID:      case_.LawyerID,
		SourceType:  "deadline",
		SourceID:    caseID,
		Title:       fmt.Sprintf("庭审提醒: %s", case_.Title),
		Content:     fmt.Sprintf("庭审时间: %s, 地点: %s", parsedDate.Format("2006-01-02 15:04"), court),
		Priority:    "important",
		DueDate:     &reminderDate,
		DueDateType: "hearing",
	}

	// 只有当提醒日期在未来时才创建
	if reminderDate.After(now) {
		_ = d.inboxRepo.Create(ctx, item)
	}
	return nil
}

// handleJudgmentReceived 处理判决收到事件 - 自动计算上诉期
func (d *EventDispatcher) handleJudgmentReceived(ctx context.Context, event *Event) error {
	caseID := event.SourceID
	judgmentDate, _ := event.Metadata["judgment_date"].(string)
	caseType, _ := event.Metadata["case_type"].(string)

	case_, err := d.caseRepo.FindByID(ctx, caseID)
	if err != nil || case_ == nil {
		return fmt.Errorf("获取案件失败: %w", err)
	}

	// 解析判决日期
	parsedDate, err := time.Parse(time.RFC3339, judgmentDate)
	if err != nil {
		parsedDate, err = time.Parse("2006-01-02", judgmentDate)
		if err != nil {
			return fmt.Errorf("判决日期格式错误: %w", err)
		}
	}

	// 根据案件类型计算上诉期
	var appealDays int
	switch caseType {
	case "刑事案件":
		appealDays = 10 // 刑事案件上诉期10天
	case "民事案件", "商事案件", "知识产权":
		appealDays = 15 // 民事案件上诉期15天
	case "行政案件":
		appealDays = 15 // 行政案件上诉期15天
	default:
		appealDays = 15 // 默认15天
	}

	// 计算上诉截止日期
	appealDeadline := parsedDate.AddDate(0, 0, appealDays)

	// 创建上诉期提醒待办事项
	item := &models.InboxItem{
		UserID:      case_.LawyerID,
		SourceType:  "deadline",
		SourceID:    caseID,
		Title:       fmt.Sprintf("上诉期截止提醒: %s", case_.Title),
		Content:     fmt.Sprintf("判决日期: %s, 上诉期截止: %s (%d天)", parsedDate.Format("2006-01-02"), appealDeadline.Format("2006-01-02"), appealDays),
		Priority:    "critical",
		DueDate:     &appealDeadline,
		DueDateType: "appeal_deadline",
	}

	if err := d.inboxRepo.Create(ctx, item); err != nil {
		return fmt.Errorf("创建上诉期提醒失败: %w", err)
	}

	// 同时创建提前提醒（上诉期前7天、前3天、前1天）
	reminders := []int{-7, -3, -1}
	for _, days := range reminders {
		reminderDate := appealDeadline.AddDate(0, 0, days)
		if reminderDate.After(time.Now()) {
			reminderItem := &models.InboxItem{
				UserID:      case_.LawyerID,
				SourceType:  "deadline",
				SourceID:    caseID,
				Title:       fmt.Sprintf("上诉期提醒 (%d天): %s", days, case_.Title),
				Content:     fmt.Sprintf("上诉期截止日期: %s, 距离截止还有%d天", appealDeadline.Format("2006-01-02"), -days),
				Priority:    "critical",
				DueDate:     &reminderDate,
				DueDateType: "appeal_deadline",
			}
			_ = d.inboxRepo.Create(ctx, reminderItem)
		}
	}

	return nil
}

// PublishCaseCreated 发布案件创建事件
func (d *EventDispatcher) PublishCaseCreated(ctx context.Context, caseID uint) error {
	return d.Dispatch(ctx, &Event{
		Type:      EventCaseCreated,
		Timestamp: time.Now(),
		SourceID:  caseID,
		Metadata:  make(map[string]interface{}),
	})
}

// PublishCaseStatusChanged 发布案件状态变更事件
func (d *EventDispatcher) PublishCaseStatusChanged(ctx context.Context, caseID uint, status string) error {
	return d.Dispatch(ctx, &Event{
		Type:      EventCaseStatusChanged,
		Timestamp: time.Now(),
		SourceID:  caseID,
		Metadata: map[string]interface{}{
			"status": status,
		},
	})
}

// PublishApprovalCreated 发布审批创建事件
func (d *EventDispatcher) PublishApprovalCreated(ctx context.Context, approvalID uint, title string, submitterID uint, approverIDs []uint) error {
	return d.Dispatch(ctx, &Event{
		Type:      EventApprovalCreated,
		Timestamp: time.Now(),
		SourceID:  approvalID,
		Metadata: map[string]interface{}{
			"title":        title,
			"submitter_id": submitterID,
			"approver_ids": approverIDs,
		},
	})
}

// PublishApprovalApproved 发布审批通过事件
func (d *EventDispatcher) PublishApprovalApproved(ctx context.Context, approvalID uint, title string, submitterID uint) error {
	return d.Dispatch(ctx, &Event{
		Type:      EventApprovalApproved,
		Timestamp: time.Now(),
		SourceID:  approvalID,
		Metadata: map[string]interface{}{
			"title":        title,
			"submitter_id": submitterID,
		},
	})
}

// PublishHearingScheduled 发布庭审安排事件
func (d *EventDispatcher) PublishHearingScheduled(ctx context.Context, caseID uint, hearingDate, court string) error {
	return d.Dispatch(ctx, &Event{
		Type:      EventCaseHearingScheduled,
		Timestamp: time.Now(),
		SourceID:  caseID,
		Metadata: map[string]interface{}{
			"hearing_date": hearingDate,
			"court":        court,
		},
	})
}

// PublishJudgmentReceived 发布判决收到事件
func (d *EventDispatcher) PublishJudgmentReceived(ctx context.Context, caseID uint, judgmentDate, caseType string) error {
	return d.Dispatch(ctx, &Event{
		Type:      EventCaseJudgmentReceived,
		Timestamp: time.Now(),
		SourceID:  caseID,
		Metadata: map[string]interface{}{
			"judgment_date": judgmentDate,
			"case_type":     caseType,
		},
	})
}

// CreateCustomInboxItem 创建自定义待办事项
func (d *EventDispatcher) CreateCustomInboxItem(ctx context.Context, userID uint, title, content string, priority string, dueDate *time.Time, dueDateType string) error {
	item := &models.InboxItem{
		UserID:      userID,
		SourceType:  "task",
		SourceID:    0,
		Title:       title,
		Content:     content,
		Priority:    priority,
		DueDate:     dueDate,
		DueDateType: dueDateType,
	}

	return d.inboxRepo.Create(ctx, item)
}
