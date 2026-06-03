package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// 审批超时相关常量
const (
	defaultApprovalTimeoutHours  = 48           // 审批超时默认阈值（小时）
	defaultEscalationDays         = 3            // 待办升级超时天数
	approvalTimeoutWarningRatio   = 0.8          // 超时预警比例（80%）
	approvalTimeoutBatchSize      = 50           // 每次处理的最大审批数量
	processDueItemsBatchSize      = 1000         // 每次处理的最大待办数量
)

// SchedulerService 定时调度服务
type SchedulerService struct {
	inboxRepo    repositories.InboxRepository
	userRepo     repositories.UserRepository
	approvalRepo *repositories.ApprovalRepository

	// Cron 表达式配置
	reminderCheckInterval   time.Duration
	escalationCheckInterval time.Duration
	timeoutCheckInterval   time.Duration

	// 控制字段
	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex
}

// NewSchedulerService 创建定时调度服务
func NewSchedulerService(inboxRepo repositories.InboxRepository, userRepo repositories.UserRepository, approvalRepo *repositories.ApprovalRepository) *SchedulerService {
	return &SchedulerService{
		inboxRepo:              inboxRepo,
		userRepo:               userRepo,
		approvalRepo:           approvalRepo,
		reminderCheckInterval:  time.Hour,
		escalationCheckInterval: 6 * time.Hour,
		timeoutCheckInterval:   30 * time.Minute,
		stopCh:                 make(chan struct{}),
		running:                false,
	}
}

// Start 启动调度服务
func (s *SchedulerService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true

	// 启动提醒检查任务
	s.wg.Add(1)
	go s.reminderChecker()

	// 启动升级检查任务
	s.wg.Add(1)
	go s.escalationChecker()

	// 启动审批超时检查任务
	if s.approvalRepo != nil {
		s.wg.Add(1)
		go s.approvalTimeoutChecker()
	}

	log.Println("SchedulerService started")
}

// Stop 停止调度服务
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopCh)
	s.wg.Wait()
	s.running = false

	log.Println("SchedulerService stopped")
}

// IsRunning 检查服务是否正在运行
func (s *SchedulerService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// reminderChecker 提醒检查器
func (s *SchedulerService) reminderChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.reminderCheckInterval)
	defer ticker.Stop()

	// 启动时立即执行一次
	s.checkReminders(context.Background())

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := s.checkReminders(ctx); err != nil {
				log.Printf("提醒检查错误: %v", err)
			}
		case <-s.stopCh:
			return
		}
	}
}

// escalationChecker 升级检查器
func (s *SchedulerService) escalationChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.escalationCheckInterval)
	defer ticker.Stop()

	// 启动时立即执行一次
	s.checkEscalations(context.Background())

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := s.checkEscalations(ctx); err != nil {
				log.Printf("升级检查错误: %v", err)
			}
		case <-s.stopCh:
			return
		}
	}
}

// checkReminders 检查并发送提醒
func (s *SchedulerService) checkReminders(ctx context.Context) error {
	now := time.Now()

	// 获取需要提醒的待办事项（到期时间在当前时间之前的未完成待办）
	dueItems, err := s.inboxRepo.GetDueItems(ctx, now)
	if err != nil {
		return fmt.Errorf("获取到期待办事项失败: %w", err)
	}

	// 获取提醒规则
	rules, err := s.inboxRepo.GetReminderRules(ctx, true)
	if err != nil {
		return fmt.Errorf("获取提醒规则失败: %w", err)
	}

	// 构建规则映射
	ruleMap := make(map[string]*models.InboxReminderRule)
	for _, rule := range rules {
		key := rule.DueDateType + ":" + rule.Priority
		ruleMap[key] = rule
	}

	// 为每个到期待办事项检查是否需要发送提醒
	for _, item := range dueItems {
		if item.DueDate == nil {
			continue
		}

		// 获取对应的提醒规则
		var rule *models.InboxReminderRule
		if item.DueDateType != "" {
			if r, ok := ruleMap[item.DueDateType+":"+item.Priority]; ok {
				rule = r
			}
		}

		// 如果没有找到匹配的规则，使用默认规则
		if rule == nil {
			// 默认提醒规则：到期当天提醒
			rule = &models.InboxReminderRule{
				ReminderOffsets: models.FromIntSlice([]int{0}),
			}
		}

		// 解析提醒偏移量
		offsets := s.parseReminderOffsets(rule.ReminderOffsets)
		shouldRemind := false

		// 检查当前时间是否在任何提醒偏移量附近
		for _, offset := range offsets {
			remindTime := item.DueDate.AddDate(0, 0, offset)
			timeDiff := now.Sub(remindTime)

			// 如果当前时间在提醒时间前后1小时内，发送提醒
			if timeDiff >= -time.Hour && timeDiff <= time.Hour {
				// 检查是否已经发送过该偏移量的提醒
				if item.ReminderCount < len(offsets) {
					shouldRemind = true
					break
				}
			}
		}

		if shouldRemind {
			// 发送提醒通知（这里可以集成实际的通知服务）
			if err := s.sendReminder(ctx, item); err != nil {
				log.Printf("发送提醒失败 (ID: %d): %v", item.ID, err)
			}

			// 更新提醒状态
			updatedItem := &models.InboxItem{
				ID:           item.ID,
				ReminderSent: true,
				ReminderCount: item.ReminderCount + 1,
			}
			if err := s.inboxRepo.Update(ctx, updatedItem); err != nil {
				log.Printf("更新提醒状态失败 (ID: %d): %v", item.ID, err)
			}
		}
	}

	return nil
}

// checkEscalations 检查待办事项升级
func (s *SchedulerService) checkEscalations(ctx context.Context) error {
	// 获取超时的 critical 待办事项
	threshold := time.Now().AddDate(0, 0, -defaultEscalationDays)

	overdueItems, err := s.inboxRepo.GetOverdueCriticalItems(ctx, threshold)
	if err != nil {
		return fmt.Errorf("获取超时待办事项失败: %w", err)
	}

	for _, item := range overdueItems {
		// 检查是否已升级
		if item.Escalated {
			continue
		}

		// 执行升级
		if err := s.escalateItem(ctx, item); err != nil {
			log.Printf("升级待办失败 (ID: %d): %v", item.ID, err)
		}
	}

	return nil
}

// sendReminder 发送提醒通知
func (s *SchedulerService) sendReminder(ctx context.Context, item *models.InboxItem) error {
	// 这里可以集成实际的通知服务
	// 例如：邮件、短信、微信等

	log.Printf("发送提醒通知 (ID: %d, Title: %s, DueDate: %s)",
		item.ID, item.Title, item.DueDate.Format("2006-01-02 15:04"))

	// TODO: 集成通知服务
	// - 邮件通知
	// - 短信通知
	// - 微信通知
	// - 站内通知

	return nil
}

// escalateItem 升级待办事项
func (s *SchedulerService) escalateItem(ctx context.Context, item *models.InboxItem) error {
	// 获取用户信息
	user, err := s.userRepo.FindByID(ctx, item.UserID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查是否有上级律师
	// 注意：这里假设 User 模型有 SupervisorID 字段
	// 如果没有，可以通过其他方式获取上级

	// 标记原待办为已升级
	if err := s.inboxRepo.Escalate(ctx, item.ID); err != nil {
		return fmt.Errorf("标记升级失败: %w", err)
	}

	// 创建升级通知待办事项
	escalationItem := &models.InboxItem{
		UserID:      item.UserID, // 暂时通知本人，实际应该通知上级
		SourceType:  "escalation",
		SourceID:    item.ID,
		Title:       fmt.Sprintf("[升级提醒] %s", item.Title),
		Content:     fmt.Sprintf("原待办事项已超时，请尽快处理。到期时间: %s", item.DueDate.Format("2006-01-02 15:04")),
		Priority:    "critical",
		DueDateType: "escalation",
	}

	if err := s.inboxRepo.Create(ctx, escalationItem); err != nil {
		return fmt.Errorf("创建升级通知失败: %w", err)
	}

	log.Printf("待办事项已升级 (ID: %d, Title: %s)", item.ID, item.Title)

	// TODO: 通知上级律师
	// - 获取用户的上级律师
	// - 为上级律师创建待办事项
	// - 发送升级通知

	return nil
}

// parseReminderOffsets 解析提醒偏移量
func (s *SchedulerService) parseReminderOffsets(offsets models.JSONArray) []int {
	var result []int

	// JSONArray 实际上是 []interface{}
	for _, v := range offsets {
		if f, ok := v.(float64); ok {
			result = append(result, int(f))
		}
	}

	return result
}

// ScheduleReminderForItem 为特定待办事项安排提醒
func (s *SchedulerService) ScheduleReminderForItem(ctx context.Context, item *models.InboxItem) error {
	if item.DueDate == nil {
		return nil
	}

	// 获取提醒规则
	rule, err := s.inboxRepo.GetReminderRuleByTypeAndPriority(ctx, item.DueDateType, item.Priority)
	if err != nil {
		return fmt.Errorf("获取提醒规则失败: %w", err)
	}

	if rule == nil {
		// 没有规则，使用默认提醒
		return nil
	}

	// 解析提醒偏移量
	offsets := s.parseReminderOffsets(rule.ReminderOffsets)

	// 为每个偏移量创建提醒待办事项
	for _, offset := range offsets {
		if offset >= 0 {
			continue // 只处理提前提醒
		}

		remindTime := item.DueDate.AddDate(0, 0, offset)
		now := time.Now()

		// 只有当提醒时间在未来时才创建
		if remindTime.After(now) {
			reminderItem := &models.InboxItem{
				UserID:      item.UserID,
				SourceType:  "reminder",
				SourceID:    item.ID,
				Title:       fmt.Sprintf("提醒: %s", item.Title),
				Content:     fmt.Sprintf("%s 将于 %s 到期", item.Title, item.DueDate.Format("2006-01-02")),
				Priority:    item.Priority,
				DueDate:     &remindTime,
				DueDateType: item.DueDateType + "_reminder",
			}

			if err := s.inboxRepo.Create(ctx, reminderItem); err != nil {
				log.Printf("创建提醒待办失败: %v", err)
			}
		}
	}

	return nil
}

// ProcessDueItemsForUser 处理指定用户的到期待办事项
func (s *SchedulerService) ProcessDueItemsForUser(ctx context.Context, userID uint) error {
	now := time.Now()

	// 获取用户的所有待办事项
	params := &repositories.InboxListParams{
		Page:     1,
		PageSize: processDueItemsBatchSize,
		UserID:   userID,
	}

	items, _, err := s.inboxRepo.List(ctx, params)
	if err != nil {
		return fmt.Errorf("获取待办事项失败: %w", err)
	}

	// 为每个待办事项安排提醒
	for _, item := range items {
		if item.DueDate != nil && item.DueDate.After(now) && !item.ReminderSent {
			if err := s.ScheduleReminderForItem(ctx, item); err != nil {
				log.Printf("安排提醒失败 (ID: %d): %v", item.ID, err)
			}
		}
	}

	return nil
}

// GetSchedulerStatus 获取调度器状态
func (s *SchedulerService) GetSchedulerStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"running":                   s.running,
		"reminder_check_interval":    s.reminderCheckInterval.String(),
		"escalation_check_interval":  s.escalationCheckInterval.String(),
		"approval_timeout_interval": s.timeoutCheckInterval.String(),
	}
}

// approvalTimeoutChecker 审批超时检查器
func (s *SchedulerService) approvalTimeoutChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.timeoutCheckInterval)
	defer ticker.Stop()

	s.checkApprovalTimeouts(context.Background())

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := s.checkApprovalTimeouts(ctx); err != nil {
				log.Printf("审批超时检查错误: %v", err)
			}
		case <-s.stopCh:
			return
		}
	}
}

// checkApprovalTimeouts 检查超时的审批
func (s *SchedulerService) checkApprovalTimeouts(ctx context.Context) error {
	if s.approvalRepo == nil {
		return nil
	}

	now := time.Now()

	// 查询所有待处理且超时的审批
	var approvals []models.ApprovalRequest
	err := s.approvalRepo.DB().WithContext(ctx).
		Where("status IN ?", []string{
			models.ApprovalStatusSubmitted,
			models.ApprovalStatusUnderReview,
		}).
		Where("submission_date IS NOT NULL").
		Where("escalated = ?", false).
		Where("submission_date < ?", now.Add(-time.Duration(defaultApprovalTimeoutHours)*time.Hour)).
		Limit(approvalTimeoutBatchSize).
		Find(&approvals).Error
	if err != nil {
		return fmt.Errorf("查询超时审批失败: %w", err)
	}

	for _, approval := range approvals {
		if approval.SubmissionDate == nil {
			continue
		}

		elapsed := now.Sub(*approval.SubmissionDate)
		timeoutHours := defaultApprovalTimeoutHours

		// 从工作流配置获取超时时间
		if approval.WorkflowConfig != "" {
			parsed := parseWorkflowTimeouts(approval.WorkflowConfig, approval.CurrentStage)
			if parsed > 0 {
				timeoutHours = parsed
			}
		}

		if elapsed >= time.Duration(timeoutHours)*time.Hour {
			s.handleApprovalTimeout(ctx, &approval, timeoutHours)
		} else if elapsed >= time.Duration(float64(timeoutHours)*approvalTimeoutWarningRatio*float64(time.Hour)) {
			// 接近超时（80%），发送提醒
			s.sendApprovalTimeoutWarning(ctx, &approval, timeoutHours, elapsed)
		}
	}

	return nil
}

// handleApprovalTimeout 处理超时的审批
func (s *SchedulerService) handleApprovalTimeout(ctx context.Context, approval *models.ApprovalRequest, timeoutHours int) {
	now := time.Now()

	// 标记为已升级
	updates := map[string]interface{}{
		"escalated":    true,
		"escalated_at": now,
	}

	// 设置超时时间
	timeoutAt := approval.SubmissionDate.Add(time.Duration(timeoutHours) * time.Hour)
	updates["timeout_at"] = timeoutAt

	if err := s.approvalRepo.DB().WithContext(ctx).Model(&models.ApprovalRequest{}).Where("id = ?", approval.ID).Updates(updates).Error; err != nil {
		log.Printf("标记审批超时失败 (ID: %s): %v", approval.ID, err)
		return
	}

	// 创建升级待办通知
	item := &models.InboxItem{
		UserID:      0, // 系统级通知，发送给管理员
		SourceType:  "approval",
		SourceID:    0,
		Title:       fmt.Sprintf("审批超时: %s", approval.Title),
		Content:     fmt.Sprintf("审批 %s 已超时 %d 小时未处理，请及时关注。申请人: %s", approval.RequestNumber, timeoutHours, approval.ApplicantName),
		Priority:    "high",
		DueDateType: "approval_timeout",
	}

	if s.inboxRepo != nil {
		_ = s.inboxRepo.Create(ctx, item)
	}

	log.Printf("审批超时已升级 (ID: %s, Number: %s, Hours: %d)", approval.ID, approval.RequestNumber, timeoutHours)
}

// sendApprovalTimeoutWarning 发送超时预警
func (s *SchedulerService) sendApprovalTimeoutWarning(ctx context.Context, approval *models.ApprovalRequest, timeoutHours int, elapsed time.Duration) {
	if approval.CurrentApproverID == "" {
		return
	}

	approverIDParsed, err := strconv.ParseUint(approval.CurrentApproverID, 10, 32)
	if err != nil {
		log.Printf("[Scheduler] 解析 CurrentApproverID 失败 (ApprovalID: %s, ApproverID: %s): %v", approval.ID, approval.CurrentApproverID, err)
		return
	}
	approverID := uint(approverIDParsed)
	if approverID == 0 {
		return
	}

	remaining := time.Duration(timeoutHours)*time.Hour - elapsed

	item := &models.InboxItem{
		UserID:      approverID,
		SourceType:  "approval",
		SourceID:    0,
		Title:       fmt.Sprintf("审批即将超时: %s", approval.Title),
		Content:     fmt.Sprintf("审批 %s 将在 %s 后超时（已等待 %d 小时），请尽快处理", approval.RequestNumber, remaining.Round(time.Hour), int(elapsed.Hours())),
		Priority:    "important",
		DueDateType: "approval_timeout_warning",
	}

	if s.inboxRepo != nil {
		_ = s.inboxRepo.Create(ctx, item)
	}
}

// parseWorkflowTimeouts 从工作流配置JSON中解析指定阶段的超时时间
// 格式示例: {"timeouts": {"department_review": 24, "partner_review": 48}}
func parseWorkflowTimeouts(workflowConfigJSON, currentStage string) int {
	if len(workflowConfigJSON) == 0 || len(currentStage) == 0 {
		return 0
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(workflowConfigJSON), &config); err != nil {
		log.Printf("[Scheduler] 解析工作流配置失败: %v", err)
		return 0
	}

	timeouts, ok := config["timeouts"].(map[string]interface{})
	if !ok {
		return 0
	}

	stageTimeout, ok := timeouts[currentStage]
	if !ok {
		return 0
	}

	switch v := stageTimeout.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}
