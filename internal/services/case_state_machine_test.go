package services

import (
	"testing"

	"law-oa-go/internal/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestCaseDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("无法连接测试数据库: %v", err)
	}
	err = db.AutoMigrate(
		&models.Case{},
		&models.CaseStatusHistory{},
		&models.User{},
		&models.Client{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}
	return db
}

func TestCaseStateMachine_CanTransition(t *testing.T) {
	sm := NewCaseStateMachine(nil)

	tests := []struct {
		name     string
		from     string
		to       string
		expected bool
	}{
		{"草拟->风控审查", models.CaseStatusDraft, models.CaseStatusRiskCheck, true},
		{"风控审查->已签约", models.CaseStatusRiskCheck, models.CaseStatusSigned, true},
		{"已签约->办案中", models.CaseStatusSigned, models.CaseStatusInProgress, true},
		{"办案中->庭审", models.CaseStatusInProgress, models.CaseStatusTrial, true},
		{"庭审->办案中(发回重审)", models.CaseStatusTrial, models.CaseStatusInProgress, true},
		{"办案中->已结案", models.CaseStatusInProgress, models.CaseStatusClosed, true},
		{"庭审->已结案", models.CaseStatusTrial, models.CaseStatusClosed, true},
		{"已结案->已归档", models.CaseStatusClosed, models.CaseStatusArchived, true},
		{"非法:草拟->已结案", models.CaseStatusDraft, models.CaseStatusClosed, false},
		{"非法:已归档->办案中", models.CaseStatusArchived, models.CaseStatusInProgress, false},
		{"非法:风控审查->庭审", models.CaseStatusRiskCheck, models.CaseStatusTrial, false},
		{"兼容:pending->草拟", "pending", models.CaseStatusDraft, true},
		{"兼容:pending->风控审查", "pending", models.CaseStatusRiskCheck, true},
		{"相同状态不变", models.CaseStatusDraft, models.CaseStatusDraft, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sm.CanTransition(tt.from, tt.to))
		})
	}
}

func TestCaseStateMachine_GetAvailableTransitions(t *testing.T) {
	sm := NewCaseStateMachine(nil)

	t.Run("办案中可转换到庭审和已结案", func(t *testing.T) {
		transitions := sm.GetAvailableTransitions(models.CaseStatusInProgress)
		assert.Len(t, transitions, 2)
		targets := make(map[string]bool)
		for _, tr := range transitions {
			targets[tr.ToStatus] = true
		}
		assert.True(t, targets[models.CaseStatusTrial])
		assert.True(t, targets[models.CaseStatusClosed])
	})

	t.Run("庭审可回到办案中或直接结案", func(t *testing.T) {
		transitions := sm.GetAvailableTransitions(models.CaseStatusTrial)
		assert.Len(t, transitions, 2)
	})

	t.Run("已归档无可用转换", func(t *testing.T) {
		transitions := sm.GetAvailableTransitions(models.CaseStatusArchived)
		assert.Empty(t, transitions)
	})
}

func TestCaseStateMachine_Transition(t *testing.T) {
	db := setupTestCaseDB(t)
	sm := NewCaseStateMachine(db)

	user := models.User{Username: "test_lawyer", Email: "lawyer@test.com", Role: "lawyer"}
	db.Create(&user)
	client := models.Client{Name: "测试客户", Type: "个人"}
	db.Create(&client)

	c := models.Case{
		CaseNumber: "TEST-2026-001",
		Title:      "测试案件状态机",
		Status:     models.CaseStatusDraft,
		ClientID:   client.ID,
		LawyerID:   user.ID,
	}
	db.Create(&c)

	t.Run("完整生命周期: draft→risk_check→signed→in_progress→trial→closed→archived", func(t *testing.T) {
		path := []struct {
			to     string
			reason string
		}{
			{models.CaseStatusRiskCheck, "提交风控审查"},
			{models.CaseStatusSigned, "风控通过，签署委托协议"},
			{models.CaseStatusInProgress, "开始办案"},
			{models.CaseStatusTrial, "案件开庭"},
			{models.CaseStatusClosed, "法院判决，案件结案"},
			{models.CaseStatusArchived, "归档"},
		}
		for i, step := range path {
			err := sm.Transition(c.ID, step.to, user.ID, step.reason)
			assert.NoError(t, err, "第%d步转换失败: %s→%s", i+1, c.Status, step.to)

			var updated models.Case
			db.First(&updated, c.ID)
			assert.Equal(t, step.to, updated.Status)
		}

		// 验证历史记录数量
		var count int64
		db.Model(&models.CaseStatusHistory{}).Where("case_id = ?", c.ID).Count(&count)
		assert.Equal(t, int64(6), count)
	})

	t.Run("庭审发回重审", func(t *testing.T) {
		// 直接通过DB重置状态到 in_progress（完整生命周期后案件已归档）
		db.Model(&models.Case{}).Where("id = ?", c.ID).Update("status", models.CaseStatusInProgress)
		sm.Transition(c.ID, models.CaseStatusTrial, user.ID, "二次开庭")

		err := sm.Transition(c.ID, models.CaseStatusInProgress, user.ID, "法院发回重审")
		assert.NoError(t, err)

		var updated models.Case
		db.First(&updated, c.ID)
		assert.Equal(t, models.CaseStatusInProgress, updated.Status)
	})

	t.Run("非法转换不改变状态", func(t *testing.T) {
		// 直接通过DB重置状态到 draft
		db.Model(&models.Case{}).Where("id = ?", c.ID).Update("status", models.CaseStatusDraft)

		err := sm.Transition(c.ID, models.CaseStatusClosed, user.ID, "越权")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不允许从状态")

		var updated models.Case
		db.First(&updated, c.ID)
		assert.Equal(t, models.CaseStatusDraft, updated.Status)
	})

	t.Run("案件不存在返回错误", func(t *testing.T) {
		err := sm.Transition(99999, models.CaseStatusSigned, user.ID, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "未找到案件")
	})
}
