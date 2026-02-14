package services

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// DeadlineCalculator 时效计算服务
type DeadlineCalculator struct {
	inboxRepo  repositories.InboxRepository
	caseRepo   repositories.CaseRepository
	dispatcher *EventDispatcher
}

// AppealPeriod 上诉期配置
type AppealPeriod struct {
	Days   int    `json:"days"`
	Reason string `json:"reason"`
}

// CaseTypeDeadlines 案件类型时效配置
var CaseTypeDeadlines = map[string]AppealPeriod{
	"民事案件":   {Days: 15, Reason: "民事判决上诉期15天"},
	"商事案件":   {Days: 15, Reason: "商事判决上诉期15天"},
	"刑事案件":   {Days: 10, Reason: "刑事判决上诉期10天"},
	"行政案件":   {Days: 15, Reason: "行政判决上诉期15天"},
	"知识产权":  {Days: 15, Reason: "知识产权案件上诉期15天"},
}

// EvidenceDeadline 举证期限配置
var EvidenceDeadlines = map[string]int{
	"民事案件":  15, // 举证期限一般15天
	"商事案件":  15,
	"刑事案件":  7,  // 刑事案件举证期限较短
	"行政案件":  15,
}

// NewDeadlineCalculator 创建时效计算服务
func NewDeadlineCalculator(
	inboxRepo repositories.InboxRepository,
	caseRepo repositories.CaseRepository,
	dispatcher *EventDispatcher,
) *DeadlineCalculator {
	return &DeadlineCalculator{
		inboxRepo:  inboxRepo,
		caseRepo:   caseRepo,
		dispatcher: dispatcher,
	}
}

// CalculateAppealDeadline 计算上诉期限
func (s *DeadlineCalculator) CalculateAppealDeadline(judgmentDate time.Time, caseType string) (time.Time, int, error) {
	period, ok := CaseTypeDeadlines[caseType]
	if !ok {
		// 默认15天
		period = AppealPeriod{Days: 15, Reason: "默认上诉期15天"}
	}

	deadline := judgmentDate.AddDate(0, 0, period.Days)
	return deadline, period.Days, nil
}

// CalculateEvidenceDeadline 计算举证期限
func (s *DeadlineCalculator) CalculateEvidenceDeadline(caseCreatedDate time.Time, caseType string) (time.Time, int, error) {
	days, ok := EvidenceDeadlines[caseType]
	if !ok {
		days = 15 // 默认15天
	}

	deadline := caseCreatedDate.AddDate(0, 0, days)
	return deadline, days, nil
}

// CalculateExecutionDeadline 计算执行申请期限
func (s *DeadlineCalculator) CalculateExecutionDeadline(judgmentDate time.Time) (time.Time, error) {
	// 判决生效后2年内可申请执行
	deadline := judgmentDate.AddDate(2, 0, 0)
	return deadline, nil
}

// CreateAppealDeadlineInbox 创建上诉期限待办事项
func (s *DeadlineCalculator) CreateAppealDeadlineInbox(ctx context.Context, caseID uint, judgmentDate time.Time, caseType string, lawyerID uint) error {
	deadline, days, err := s.CalculateAppealDeadline(judgmentDate, caseType)
	if err != nil {
		return fmt.Errorf("计算上诉期限失败: %w", err)
	}

	// 创建主待办事项
	item := &models.InboxItem{
		UserID:      lawyerID,
		SourceType:  "deadline",
		SourceID:    caseID,
		Title:       fmt.Sprintf("上诉期截止提醒"),
		Content:     fmt.Sprintf("案件类型: %s, 判决日期: %s, 上诉期: %d天, 截止日期: %s",
			caseType, judgmentDate.Format("2006-01-02"), days, deadline.Format("2006-01-02")),
		Priority:    "critical",
		DueDate:     &deadline,
		DueDateType: "appeal_deadline",
	}

	if err := s.inboxRepo.Create(ctx, item); err != nil {
		return fmt.Errorf("创建上诉期待办失败: %w", err)
	}

	// 创建提前提醒待办事项
	reminderDays := []int{-7, -3, -1}
	for _, days := range reminderDays {
		reminderDate := deadline.AddDate(0, 0, days)
		if reminderDate.After(time.Now()) {
			reminderItem := &models.InboxItem{
				UserID:      lawyerID,
				SourceType:  "deadline",
				SourceID:    caseID,
				Title:       fmt.Sprintf("上诉期提醒 (%d天)", -days),
				Content:     fmt.Sprintf("上诉期将于 %s 到期，距离截止还有%d天", deadline.Format("2006-01-02"), -days),
				Priority:    "critical",
				DueDate:     &reminderDate,
				DueDateType: "appeal_deadline_reminder",
			}
			_ = s.inboxRepo.Create(ctx, reminderItem)
		}
	}

	return nil
}

// CreateEvidenceDeadlineInbox 创建举证期限待办事项
func (s *DeadlineCalculator) CreateEvidenceDeadlineInbox(ctx context.Context, caseID uint, caseType string, lawyerID uint) error {
	case_, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return fmt.Errorf("获取案件失败: %w", err)
	}
	if case_ == nil || case_.CreatedAt.IsZero() {
		return fmt.Errorf("案件创建时间无效")
	}

	deadline, days, err := s.CalculateEvidenceDeadline(case_.CreatedAt, caseType)
	if err != nil {
		return fmt.Errorf("计算举证期限失败: %w", err)
	}

	// 只有当截止日期在未来时才创建
	if deadline.Before(time.Now()) {
		return nil
	}

	item := &models.InboxItem{
		UserID:      lawyerID,
		SourceType:  "deadline",
		SourceID:    caseID,
		Title:       "举证期限提醒",
		Content:     fmt.Sprintf("案件类型: %s, 举证期限: %d天, 截止日期: %s", caseType, days, deadline.Format("2006-01-02")),
		Priority:    "critical",
		DueDate:     &deadline,
		DueDateType: "evidence_deadline",
	}

	if err := s.inboxRepo.Create(ctx, item); err != nil {
		return fmt.Errorf("创建举证期限待办失败: %w", err)
	}

	// 创建提前提醒
	reminderDays := []int{-7, -3, -1}
	for _, days := range reminderDays {
		reminderDate := deadline.AddDate(0, 0, days)
		if reminderDate.After(time.Now()) {
			reminderItem := &models.InboxItem{
				UserID:      lawyerID,
				SourceType:  "deadline",
				SourceID:    caseID,
				Title:       fmt.Sprintf("举证期限提醒 (%d天)", -days),
				Content:     fmt.Sprintf("举证期限将于 %s 到期", deadline.Format("2006-01-02")),
				Priority:    "critical",
				DueDate:     &reminderDate,
				DueDateType: "evidence_deadline_reminder",
			}
			_ = s.inboxRepo.Create(ctx, reminderItem)
		}
	}

	return nil
}

// CreateExecutionDeadlineInbox 创建执行申请期限待办事项
func (s *DeadlineCalculator) CreateExecutionDeadlineInbox(ctx context.Context, caseID uint, judgmentDate time.Time, lawyerID uint) error {
	deadline, err := s.CalculateExecutionDeadline(judgmentDate)
	if err != nil {
		return fmt.Errorf("计算执行申请期限失败: %w", err)
	}

	// 判决生效后6个月提醒（执行申请期限2年，提前提醒）
	reminderDate := judgmentDate.AddDate(0, 6, 0)

	item := &models.InboxItem{
		UserID:      lawyerID,
		SourceType:  "deadline",
		SourceID:    caseID,
		Title:       "执行申请期限提醒",
		Content:     fmt.Sprintf("判决日期: %s, 执行申请期限: 2年, 截止日期: %s", judgmentDate.Format("2006-01-02"), deadline.Format("2006-01-02")),
		Priority:    "important",
		DueDate:     &reminderDate,
		DueDateType: "execution_deadline",
	}

	if err := s.inboxRepo.Create(ctx, item); err != nil {
		return fmt.Errorf("创建执行申请期限待办失败: %w", err)
	}

	return nil
}

// ProcessCaseJudgment 处理案件判决，自动创建相关待办事项
func (s *DeadlineCalculator) ProcessCaseJudgment(ctx context.Context, caseID uint, judgmentDate time.Time, caseType string, lawyerID uint) error {
	// 1. 创建上诉期限待办事项
	if err := s.CreateAppealDeadlineInbox(ctx, caseID, judgmentDate, caseType, lawyerID); err != nil {
		return fmt.Errorf("创建上诉期限待办失败: %w", err)
	}

	// 2. 创建执行申请期限待办事项
	if err := s.CreateExecutionDeadlineInbox(ctx, caseID, judgmentDate, lawyerID); err != nil {
		return fmt.Errorf("创建执行申请期限待办失败: %w", err)
	}

	// 3. 发布判决事件
	_ = s.dispatcher.Dispatch(ctx, &Event{
		Type:      EventCaseJudgmentReceived,
		Timestamp: time.Now(),
		SourceID:  caseID,
		Metadata: map[string]interface{}{
			"judgment_date": judgmentDate.Format(time.RFC3339),
			"case_type":     caseType,
		},
	})

	return nil
}

// ProcessNewCase 处理新案件，自动创建相关待办事项
func (s *DeadlineCalculator) ProcessNewCase(ctx context.Context, caseID uint, caseType string, lawyerID uint) error {
	// 创建举证期限待办事项
	if err := s.CreateEvidenceDeadlineInbox(ctx, caseID, caseType, lawyerID); err != nil {
		return fmt.Errorf("创建举证期限待办失败: %w", err)
	}

	return nil
}

// GetUpcomingDeadlines 获取即将到期的期限
func (s *DeadlineCalculator) GetUpcomingDeadlines(ctx context.Context, userID uint, days int) ([]*models.InboxItem, error) {
	threshold := time.Now().AddDate(0, 0, days)

	params := &repositories.InboxListParams{
		Page:      1,
		PageSize:  100,
		UserID:    userID,
		DueAfter:  timePtr(time.Now()),
		DueBefore: &threshold,
		OrderBy:   "due_date",
	}

	items, _, err := s.inboxRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// 过滤出期限相关的待办
	var deadlineItems []*models.InboxItem
	for _, item := range items {
		if item.DueDateType == "appeal_deadline" ||
			item.DueDateType == "evidence_deadline" ||
			item.DueDateType == "execution_deadline" ||
			item.DueDateType == "statute_of_limitations" {
			deadlineItems = append(deadlineItems, item)
		}
	}

	return deadlineItems, nil
}

// DeadlineInfo 期限信息
type DeadlineInfo struct {
	CaseID       uint       `json:"case_id"`
	CaseType     string     `json:"case_type"`
	JudgmentDate *time.Time `json:"judgment_date"`
	AppealDeadline *time.Time `json:"appeal_deadline"`
	AppealDays   int        `json:"appeal_days"`
	EvidenceDeadline *time.Time `json:"evidence_deadline"`
	ExecutionDeadline *time.Time `json:"execution_deadline"`
}

// GetCaseDeadlines 获取案件所有期限信息
func (s *DeadlineCalculator) GetCaseDeadlines(ctx context.Context, caseID uint) (*DeadlineInfo, error) {
	case_, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("获取案件失败: %w", err)
	}
	if case_ == nil {
		return nil, fmt.Errorf("案件不存在")
	}

	info := &DeadlineInfo{
		CaseID:   caseID,
		CaseType: case_.CaseType,
	}

	// 计算举证期限
	evidenceDeadline, _, _ := s.CalculateEvidenceDeadline(case_.CreatedAt, case_.CaseType)
	info.EvidenceDeadline = &evidenceDeadline

	// 如果有判决日期，计算上诉期限和执行申请期限
	// 这里可以从案件扩展字段获取判决日期
	// judgmentDate := case_.JudgmentDate
	// if judgmentDate != nil {
	//     appealDeadline, appealDays, _ := s.CalculateAppealDeadline(*judgmentDate, case_.CaseType)
	//     info.JudgmentDate = judgmentDate
	//     info.AppealDeadline = &appealDeadline
	//     info.AppealDays = appealDays
	//
	//     executionDeadline, _ := s.CalculateExecutionDeadline(*judgmentDate)
	//     info.ExecutionDeadline = &executionDeadline
	// }

	return info, nil
}

// timePtr 返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}

// CalculateStatuteOfLimitations 计算诉讼时效
func (s *DeadlineCalculator) CalculateStatuteOfLimitations(incidentDate time.Time, caseType string) (time.Time, int, error) {
	// 一般民事诉讼时效为3年
	// 特殊情况：
	// - 身体伤害: 1年
	// - 商品质量: 1年
	// - 房地产: 3年（从知道权利被侵害时起）
	// - 国际货物买卖: 4年

	var years int
	switch caseType {
	case "身体伤害":
		years = 1
	case "商品质量":
		years = 1
	case "国际货物买卖":
		years = 4
	default:
		years = 3
	}

	deadline := incidentDate.AddDate(years, 0, 0)
	return deadline, years * 365, nil
}

// CreateStatuteOfLimitationsInbox 创建诉讼时效待办事项
func (s *DeadlineCalculator) CreateStatuteOfLimitationsInbox(ctx context.Context, caseID uint, incidentDate time.Time, caseType string, lawyerID uint) error {
	deadline, days, err := s.CalculateStatuteOfLimitations(incidentDate, caseType)
	if err != nil {
		return fmt.Errorf("计算诉讼时效失败: %w", err)
	}

	// 创建诉讼时效待办事项
	item := &models.InboxItem{
		UserID:      lawyerID,
		SourceType:  "deadline",
		SourceID:    caseID,
		Title:       "诉讼时效提醒",
		Content:     fmt.Sprintf("案件类型: %s, 事发起: %s, 诉讼时效: %d天, 截止日期: %s", caseType, incidentDate.Format("2006-01-02"), days, deadline.Format("2006-01-02")),
		Priority:    "critical",
		DueDate:     &deadline,
		DueDateType: "statute_of_limitations",
	}

	if err := s.inboxRepo.Create(ctx, item); err != nil {
		return fmt.Errorf("创建诉讼时效待办失败: %w", err)
	}

	// 创建提前提醒（90天、30天、7天、1天前）
	reminderDays := []int{-90, -30, -7, -1}
	for _, days := range reminderDays {
		reminderDate := deadline.AddDate(0, 0, days)
		if reminderDate.After(time.Now()) {
			reminderItem := &models.InboxItem{
				UserID:      lawyerID,
				SourceType:  "deadline",
				SourceID:    caseID,
				Title:       fmt.Sprintf("诉讼时效提醒 (%d天)", -days),
				Content:     fmt.Sprintf("诉讼时效将于 %s 到期", deadline.Format("2006-01-02")),
				Priority:    "critical",
				DueDate:     &reminderDate,
				DueDateType: "statute_of_limitations_reminder",
			}
			_ = s.inboxRepo.Create(ctx, reminderItem)
		}
	}

	return nil
}
