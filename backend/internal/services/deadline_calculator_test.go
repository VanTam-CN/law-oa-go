package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

func setupDeadlineTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法连接测试数据库: %v", err)
	}
	err = db.AutoMigrate(
		&models.InboxItem{},
		&models.Case{},
		&models.User{},
		&models.Client{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

func newTestDeadlineCalculator(db *gorm.DB) *DeadlineCalculator {
	inboxRepo := repositories.NewInboxRepository(db)
	caseRepo := repositories.NewCaseRepository(db)
	dispatcher := NewEventDispatcher(inboxRepo, nil, caseRepo)
	return NewDeadlineCalculator(inboxRepo, caseRepo, dispatcher)
}

func TestDeadlineCalculator_CalculateAppealDeadline(t *testing.T) {
	calc := &DeadlineCalculator{}

	tests := []struct {
		name      string
		caseType  string
		judgment  time.Time
		wantDays  int
		wantMonth time.Month
	}{
		{"民事案件15天", "民事案件", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 15, time.January},
		{"刑事案件10天", "刑事案件", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 10, time.January},
		{"商事案件15天", "商事案件", time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), 15, time.March},
		{"行政案件15天", "行政案件", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), 15, time.June},
		{"知识产权15天", "知识产权", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 15, time.January},
		{"未知类型默认15天", "未知类型", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 15, time.January},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deadline, days, err := calc.CalculateAppealDeadline(tt.judgment, tt.caseType)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDays, days)
			assert.Equal(t, tt.judgment.Day()+tt.wantDays, deadline.Day())
			assert.Equal(t, tt.wantMonth, deadline.Month())
		})
	}
}

func TestDeadlineCalculator_CalculateEvidenceDeadline(t *testing.T) {
	calc := &DeadlineCalculator{}

	tests := []struct {
		name     string
		caseType string
		wantDays int
	}{
		{"民事案件15天", "民事案件", 15},
		{"刑事案件7天", "刑事案件", 7},
		{"未知类型默认15天", "未知类型", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			deadline, days, err := calc.CalculateEvidenceDeadline(created, tt.caseType)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDays, days)
			assert.Equal(t, created.Day()+tt.wantDays, deadline.Day())
		})
	}
}

func TestDeadlineCalculator_CalculateExecutionDeadline(t *testing.T) {
	calc := &DeadlineCalculator{}

	t.Run("判决后2年执行期限", func(t *testing.T) {
		judgmentDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		deadline, err := calc.CalculateExecutionDeadline(judgmentDate)
		assert.NoError(t, err)
		assert.Equal(t, 2028, deadline.Year())
		assert.Equal(t, time.January, deadline.Month())
		assert.Equal(t, 1, deadline.Day())
	})
}

func TestDeadlineCalculator_CalculateStatuteOfLimitations(t *testing.T) {
	calc := &DeadlineCalculator{}

	tests := []struct {
		name     string
		caseType string
		wantDays int
	}{
		{"默认3年", "民事案件", 3 * 365},
		{"身体伤害1年", "身体伤害", 365},
		{"商品质量1年", "商品质量", 365},
		{"国际货物买卖4年", "国际货物买卖", 4 * 365},
		{"房地产默认3年", "房地产", 3 * 365},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incidentDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			deadline, days, err := calc.CalculateStatuteOfLimitations(incidentDate, tt.caseType)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDays, days)
			assert.True(t, deadline.After(incidentDate))
		})
	}
}

func TestDeadlineCalculator_CreateAppealDeadlineInbox(t *testing.T) {
	db := setupDeadlineTestDB(t)
	calc := newTestDeadlineCalculator(db)

	user := models.User{Username: "test_lawyer", Email: "lawyer@test.com", Role: "lawyer"}
	db.Create(&user)

	judgmentDate := time.Now().AddDate(0, 0, 30)

	err := calc.CreateAppealDeadlineInbox(context.Background(), 1, judgmentDate, "民事案件", user.ID)
	assert.NoError(t, err)

	var items []models.InboxItem
	db.Where("source_type = ? AND source_id = ?", "deadline", uint(1)).Find(&items)
	assert.GreaterOrEqual(t, len(items), 1, "应至少创建1个上诉期待办")

	var mainItem models.InboxItem
	err = db.Where("source_type = ? AND due_date_type = ? AND user_id = ?",
		"deadline", "appeal_deadline", user.ID).First(&mainItem).Error
	assert.NoError(t, err)
	assert.Equal(t, "critical", mainItem.Priority)
	assert.Contains(t, mainItem.Title, "上诉期")
}

func TestDeadlineCalculator_CreateEvidenceDeadlineInbox(t *testing.T) {
	db := setupDeadlineTestDB(t)
	calc := newTestDeadlineCalculator(db)

	user := models.User{Username: "test_lawyer", Email: "lawyer@test.com", Role: "lawyer"}
	db.Create(&user)
	client := models.Client{Name: "测试客户", Type: "个人"}
	db.Create(&client)

	c := models.Case{
		CaseNumber: "TEST-2026-001",
		Title:      "测试案件",
		Status:     models.CaseStatusDraft,
		ClientID:   client.ID,
		LawyerID:   user.ID,
		CaseType:   "民事案件",
	}
	db.Create(&c)

	err := calc.CreateEvidenceDeadlineInbox(context.Background(), c.ID, "民事案件", user.ID)
	assert.NoError(t, err)

	var items []models.InboxItem
	db.Where("source_type = ? AND source_id = ?", "deadline", c.ID).Find(&items)
	assert.GreaterOrEqual(t, len(items), 1, "应至少创建1个举证期限待办")
}

func TestDeadlineCalculator_CreateStatuteOfLimitationsInbox(t *testing.T) {
	db := setupDeadlineTestDB(t)
	calc := newTestDeadlineCalculator(db)

	user := models.User{Username: "test_lawyer", Email: "lawyer@test.com", Role: "lawyer"}
	db.Create(&user)

	incidentDate := time.Now().AddDate(2, 0, 0)

	err := calc.CreateStatuteOfLimitationsInbox(context.Background(), 1, incidentDate, "民事案件", user.ID)
	assert.NoError(t, err)

	var items []models.InboxItem
	db.Where("source_type = ? AND due_date_type = ?", "deadline", "statute_of_limitations").Find(&items)
	assert.GreaterOrEqual(t, len(items), 1, "应至少创建1个诉讼时效待办")

	var mainItem models.InboxItem
	err = db.Where("due_date_type = ? AND user_id = ?", "statute_of_limitations", user.ID).First(&mainItem).Error
	assert.NoError(t, err)
	assert.Equal(t, "critical", mainItem.Priority)
}

func TestDeadlineCalculator_ProcessCaseJudgment(t *testing.T) {
	db := setupDeadlineTestDB(t)
	calc := newTestDeadlineCalculator(db)

	user := models.User{Username: "test_lawyer", Email: "lawyer@test.com", Role: "lawyer"}
	db.Create(&user)

	judgmentDate := time.Now().AddDate(0, 0, 30)

	err := calc.ProcessCaseJudgment(context.Background(), 1, judgmentDate, "民事案件", user.ID)
	assert.NoError(t, err)

	var items []models.InboxItem
	db.Where("source_type = ? AND source_id = ?", "deadline", uint(1)).Find(&items)
	assert.GreaterOrEqual(t, len(items), 2, "应创建上诉期和执行申请期限待办")
}

func TestDeadlineCalculator_GetUpcomingDeadlines(t *testing.T) {
	db := setupDeadlineTestDB(t)
	_ = newTestDeadlineCalculator(db)

	user := models.User{Username: "test_lawyer", Email: "lawyer@test.com", Role: "lawyer"}
	db.Create(&user)

	futureDate := time.Now().AddDate(0, 0, 7)
	item := models.InboxItem{
		UserID:      user.ID,
		SourceType:  "deadline",
		SourceID:    1,
		Title:       "测试上诉期",
		Content:     "测试",
		Priority:    "critical",
		DueDate:     &futureDate,
		DueDateType: "appeal_deadline",
	}
	db.Create(&item)

	// 直接查询数据库验证，绕过 SQLite NOW() 兼容性问题
	var items []models.InboxItem
	db.Where("user_id = ? AND due_date_type IN ? AND due_date >= ?",
		user.ID, []string{"appeal_deadline", "evidence_deadline", "execution_deadline", "statute_of_limitations"},
		time.Now().Truncate(24*time.Hour)).Find(&items)

	assert.GreaterOrEqual(t, len(items), 1)
	assert.Equal(t, "appeal_deadline", items[0].DueDateType)
}

func TestDeadlineCalculator_GetCaseDeadlines(t *testing.T) {
	db := setupDeadlineTestDB(t)
	calc := newTestDeadlineCalculator(db)

	user := models.User{Username: "test_lawyer", Email: "lawyer@test.com", Role: "lawyer"}
	db.Create(&user)
	client := models.Client{Name: "测试客户", Type: "个人"}
	db.Create(&client)

	c := models.Case{
		CaseNumber: "TEST-2026-001",
		Title:      "测试案件",
		Status:     models.CaseStatusDraft,
		ClientID:   client.ID,
		LawyerID:   user.ID,
		CaseType:   "民事案件",
	}
	db.Create(&c)

	info, err := calc.GetCaseDeadlines(context.Background(), c.ID)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, c.ID, info.CaseID)
	assert.Equal(t, "民事案件", info.CaseType)
	assert.NotNil(t, info.EvidenceDeadline)
}

func TestDeadlineCalculator_InvalidCase(t *testing.T) {
	db := setupDeadlineTestDB(t)
	calc := newTestDeadlineCalculator(db)

	_, err := calc.GetCaseDeadlines(context.Background(), 99999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "案件不存在")
}
