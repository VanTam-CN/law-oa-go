package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// EscalationService 升级通知服务
type EscalationService struct {
	inboxRepo  repositories.InboxRepository
	userRepo   repositories.UserRepository
	dispatcher *EventDispatcher
}

// EscalationConfig 升级配置
type EscalationConfig struct {
	// Critical 待办超时阈值（天）
	CriticalEscalationDays int
	// High 待办超时阈值（天）
	HighEscalationDays int
	// 最大升级次数
	MaxEscalationLevel int
	// 是否启用自动升级
	AutoEscalationEnabled bool
}

// DefaultEscalationConfig 默认升级配置
var DefaultEscalationConfig = EscalationConfig{
	CriticalEscalationDays: 3,
	HighEscalationDays:     7,
	MaxEscalationLevel:     3,
	AutoEscalationEnabled:  true,
}

// NewEscalationService 创建升级通知服务
func NewEscalationService(
	inboxRepo repositories.InboxRepository,
	userRepo repositories.UserRepository,
	dispatcher *EventDispatcher,
) *EscalationService {
	return &EscalationService{
		inboxRepo:  inboxRepo,
		userRepo:   userRepo,
		dispatcher: dispatcher,
	}
}

// CheckOverdueItems 检查超时待办并执行升级
func (s *EscalationService) CheckOverdueItems(ctx context.Context, config EscalationConfig) error {
	if !config.AutoEscalationEnabled {
		return nil
	}

	// 检查 Critical 待办
	if err := s.checkCriticalItems(ctx, config); err != nil {
		return fmt.Errorf("检查 critical 待办失败: %w", err)
	}

	// 检查 High 待办
	if err := s.checkHighItems(ctx, config); err != nil {
		return fmt.Errorf("检查 high 待办失败: %w", err)
	}

	return nil
}

// checkCriticalItems 检查 Critical 待办
func (s *EscalationService) checkCriticalItems(ctx context.Context, config EscalationConfig) error {
	threshold := time.Now().AddDate(0, 0, -config.CriticalEscalationDays)

	items, err := s.inboxRepo.GetOverdueCriticalItems(ctx, threshold)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Escalated {
			continue
		}

		if err := s.escalateItem(ctx, item, "critical_timeout"); err != nil {
			log.Printf("升级 critical 待办失败 (ID: %d): %v", item.ID, err)
		}
	}

	return nil
}

// checkHighItems 检查 High 待办
func (s *EscalationService) checkHighItems(ctx context.Context, config EscalationConfig) error {
	threshold := time.Now().AddDate(0, 0, -config.HighEscalationDays)

	params := &repositories.InboxListParams{
		Page:        1,
		PageSize:    1000,
		Priority:    "high",
		DueBefore:   &threshold,
		IsCompleted: boolPtr(false),
	}

	items, _, err := s.inboxRepo.List(ctx, params)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.Escalated {
			continue
		}

		if err := s.escalateItem(ctx, item, "high_timeout"); err != nil {
			log.Printf("升级 high 待办失败 (ID: %d): %v", item.ID, err)
		}
	}

	return nil
}

// escalateItem 升级待办事项
func (s *EscalationService) escalateItem(ctx context.Context, item *models.InboxItem, reason string) error {
	// 获取用户信息
	user, err := s.userRepo.FindByID(ctx, item.UserID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 标记原待办为已升级
	if err := s.inboxRepo.Escalate(ctx, item.ID); err != nil {
		return fmt.Errorf("标记升级失败: %w", err)
	}

	// 获取上级律师
	supervisorID := s.getSupervisorID(user)

	// 为上级律师创建升级通知
	if supervisorID != 0 {
		escalationItem := &models.InboxItem{
			UserID:      supervisorID,
			SourceType:  "escalation",
			SourceID:    item.ID,
			Title:       fmt.Sprintf("[升级提醒] 下属待办超时: %s", item.Title),
			Content:     fmt.Sprintf("下属: %s, 原到期时间: %s, 超时原因: %s", user.Name, item.DueDate.Format("2006-01-02"), reason),
			Priority:    "critical",
			DueDateType: "escalation",
		}

		if err := s.inboxRepo.Create(ctx, escalationItem); err != nil {
			log.Printf("创建上级升级通知失败: %v", err)
		}

		log.Printf("待办已升级到上级 (ID: %d, SupervisorID: %d)", item.ID, supervisorID)
	}

	// 同时通知原用户
	notifyItem := &models.InboxItem{
		UserID:      item.UserID,
		SourceType:  "escalation",
		SourceID:    item.ID,
		Title:       fmt.Sprintf("[升级通知] 您的待办已超时: %s", item.Title),
		Content:     fmt.Sprintf("由于超时，该待办已升级处理。请尽快处理。到期时间: %s", item.DueDate.Format("2006-01-02")),
		Priority:    "critical",
		DueDateType: "escalation_notice",
	}

	if err := s.inboxRepo.Create(ctx, notifyItem); err != nil {
		log.Printf("创建升级通知失败: %v", err)
	}

	// 发布升级事件
	s.dispatcher.Dispatch(ctx, &Event{
		Type:      "inbox.escalated",
		Timestamp: time.Now(),
		SourceID:  item.ID,
		Metadata: map[string]interface{}{
			"user_id":           item.UserID,
			"supervisor_id":     supervisorID,
			"escalation_reason": reason,
			"original_due_date": item.DueDate,
		},
	})

	return nil
}

// ManualEscalate 手动升级待办
func (s *EscalationService) ManualEscalate(ctx context.Context, itemID uint, reason string, notifySupervisor bool) error {
	// 获取待办事项
	item, err := s.inboxRepo.FindByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("获取待办失败: %w", err)
	}
	if item == nil {
		return fmt.Errorf("待办不存在")
	}

	if !notifySupervisor {
		// 仅通知本人
		notifyItem := &models.InboxItem{
			UserID:      item.UserID,
			SourceType:  "escalation",
			SourceID:    itemID,
			Title:       fmt.Sprintf("[升级通知] %s", item.Title),
			Content:     reason,
			Priority:    "high",
			DueDateType: "manual_escalation",
		}

		return s.inboxRepo.Create(ctx, notifyItem)
	}

	// 完整升级流程
	return s.escalateItem(ctx, item, "manual")
}

// GetEscalationHistory 获取升级历史
func (s *EscalationService) GetEscalationHistory(ctx context.Context, userID uint) ([]*models.InboxItem, error) {
	params := &repositories.InboxListParams{
		Page:       1,
		PageSize:   100,
		UserID:     userID,
		SourceType: "escalation",
	}

	items, _, err := s.inboxRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// CancelEscalation 取消升级
func (s *EscalationService) CancelEscalation(ctx context.Context, itemID uint) error {
	item, err := s.inboxRepo.FindByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("获取待办失败: %w", err)
	}
	if item == nil {
		return fmt.Errorf("待办不存在")
	}

	// 取消升级标记
	item.Escalated = false
	item.EscalatedAt = nil

	return s.inboxRepo.Update(ctx, item)
}

// getSupervisorID 获取上级律师ID
func (s *EscalationService) getSupervisorID(user *models.User) uint {
	// 这里需要根据实际的 User 模型获取上级ID
	// 假设 User 模型有 SupervisorID 字段
	// 如果没有，可以通过其他方式获取

	// 暂时返回0，表示没有上级
	// TODO: 实现实际的上级获取逻辑
	return 0
}

// GetEscalationStats 获取升级统计
func (s *EscalationService) GetEscalationStats(ctx context.Context) (*EscalationStats, error) {
	params := &repositories.InboxListParams{
		Page:       1,
		PageSize:   1000,
		SourceType: "escalation",
	}

	items, total, err := s.inboxRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	stats := &EscalationStats{
		TotalEscalations: int(total),
		ByReason:         make(map[string]int),
	}

	for _, item := range items {
		stats.ByReason[item.DueDateType]++
	}

	return stats, nil
}

// EscalationStats 升级统计
type EscalationStats struct {
	TotalEscalations int            `json:"total_escalations"`
	ByReason         map[string]int `json:"by_reason"`
}

// boolPtr 返回 bool 指针
func boolPtr(b bool) *bool {
	return &b
}
