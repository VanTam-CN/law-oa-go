package services

import (
	"errors"
	"fmt"
	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// CaseStateMachine 案件状态机
type CaseStateMachine struct {
	db          *gorm.DB
	transitions []models.CaseStatusTransition
}

// NewCaseStateMachine 创建新的状态机实例
func NewCaseStateMachine(db *gorm.DB) *CaseStateMachine {
	sm := &CaseStateMachine{db: db}
	sm.initializeTransitions()
	return sm
}

// initializeTransitions 初始化状态转换规则
// draft → risk_check → signed → in_progress → trial → closed → archived
func (sm *CaseStateMachine) initializeTransitions() {
	sm.transitions = []models.CaseStatusTransition{
		{FromStatus: models.CaseStatusDraft, ToStatus: models.CaseStatusRiskCheck},
		{FromStatus: models.CaseStatusRiskCheck, ToStatus: models.CaseStatusSigned},
		{FromStatus: models.CaseStatusSigned, ToStatus: models.CaseStatusInProgress},
		{FromStatus: models.CaseStatusInProgress, ToStatus: models.CaseStatusTrial},
		{FromStatus: models.CaseStatusTrial, ToStatus: models.CaseStatusInProgress},
		{FromStatus: models.CaseStatusInProgress, ToStatus: models.CaseStatusClosed},
		{FromStatus: models.CaseStatusTrial, ToStatus: models.CaseStatusClosed},
		{FromStatus: models.CaseStatusClosed, ToStatus: models.CaseStatusArchived},
		// 兼容旧状态 pending
		{FromStatus: "pending", ToStatus: models.CaseStatusDraft},
		{FromStatus: "pending", ToStatus: models.CaseStatusRiskCheck},
	}
}

// CanTransition 检查状态转换是否合法
func (sm *CaseStateMachine) CanTransition(from, to string) bool {
	if from == to {
		return true
	}
	for _, t := range sm.transitions {
		if t.FromStatus == from && t.ToStatus == to {
			return true
		}
	}
	return false
}

// GetAvailableTransitions 获取当前状态下可用的转换
func (sm *CaseStateMachine) GetAvailableTransitions(status string) []models.CaseStatusTransition {
	var available []models.CaseStatusTransition
	for _, t := range sm.transitions {
		if t.FromStatus == status {
			available = append(available, t)
		}
	}
	return available
}

// Transition 执行状态转换并记录历史
func (sm *CaseStateMachine) Transition(caseID uint, to string, operatorID uint, reason string) error {
	return sm.db.Transaction(func(tx *gorm.DB) error {
		var c models.Case
		if err := tx.First(&c, caseID).Error; err != nil {
			return fmt.Errorf("未找到案件 ID %d: %w", caseID, err)
		}

		from := c.Status
		if !sm.CanTransition(from, to) {
			return fmt.Errorf("不允许从状态 '%s' 转换到 '%s'", from, to)
		}

		if err := sm.validateBusinessRules(tx, &c, to); err != nil {
			return err
		}

		if err := tx.Model(&c).Update("status", to).Error; err != nil {
			return fmt.Errorf("更新案件状态失败: %w", err)
		}

		history := models.CaseStatusHistory{
			CaseID:     caseID,
			FromStatus: from,
			ToStatus:   to,
			OperatorID: operatorID,
			Reason:     reason,
		}
		if err := tx.Create(&history).Error; err != nil {
			return fmt.Errorf("保存状态变更历史失败: %w", err)
		}

		return nil
	})
}

// validateBusinessRules 验证特定状态转换的业务规则
func (sm *CaseStateMachine) validateBusinessRules(tx *gorm.DB, c *models.Case, to string) error {
	if c.Status == models.CaseStatusClosed && to == models.CaseStatusArchived {
		if err := sm.checkClosingDocuments(tx, c.ID); err != nil {
			return err
		}
	}

	for _, t := range sm.transitions {
		if t.FromStatus == c.Status && t.ToStatus == to && t.NeedsApproval {
			return errors.New("此操作需要通过审批流程进行状态转换")
		}
	}

	return nil
}

// checkClosingDocuments 检查结案文书（预留接口）
func (sm *CaseStateMachine) checkClosingDocuments(tx *gorm.DB, caseID uint) error {
	// TODO: 待文档模块完善后，检查案件是否已上传结案文书
	return nil
}
