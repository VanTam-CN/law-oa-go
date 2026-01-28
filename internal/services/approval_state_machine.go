package services

import (
	"errors"
	"fmt"

	"law-oa-go/internal/models"
)

// ApprovalStateMachine 审批状态机
// 负责管理审批状态转换规则和验证
type ApprovalStateMachine struct {
	// 定义合法的状态转换
	validTransitions map[string][]string
}

// NewApprovalStateMachine 创建新的状态机实例
func NewApprovalStateMachine() *ApprovalStateMachine {
	sm := &ApprovalStateMachine{
		validTransitions: make(map[string][]string),
	}
	sm.initializeTransitions()
	return sm
}

// initializeTransitions 初始化状态转换规则
func (sm *ApprovalStateMachine) initializeTransitions() {
	// 草稿可以转换为：已提交、已撤回
	sm.validTransitions[models.ApprovalStatusDraft] = []string{
		models.ApprovalStatusSubmitted,
		models.ApprovalStatusCancelled,
	}

	// 已提交可以转换为：审核中、已撤回
	sm.validTransitions[models.ApprovalStatusSubmitted] = []string{
		models.ApprovalStatusUnderReview,
		models.ApprovalStatusCancelled,
	}

	// 审核中可以转换为：已通过、已拒绝、需要修改
	sm.validTransitions[models.ApprovalStatusUnderReview] = []string{
		models.ApprovalStatusApproved,
		models.ApprovalStatusRejected,
		models.ApprovalStatusNeedsRevision,
	}

	// 已拒绝可以转换为：重新提交
	sm.validTransitions[models.ApprovalStatusRejected] = []string{
		models.ApprovalStatusResubmitted,
	}

	// 需要修改可以转换为：重新提交
	sm.validTransitions[models.ApprovalStatusNeedsRevision] = []string{
		models.ApprovalStatusResubmitted,
	}

	// 重新提交可以转换为：审核中
	sm.validTransitions[models.ApprovalStatusResubmitted] = []string{
		models.ApprovalStatusUnderReview,
	}

	// 已通过、已撤回是终态，不允许转换
	sm.validTransitions[models.ApprovalStatusApproved] = []string{}
	sm.validTransitions[models.ApprovalStatusCancelled] = []string{}
}

// CanTransition 检查状态转换是否合法
func (sm *ApprovalStateMachine) CanTransition(from, to string) bool {
	// 相同状态不需要转换
	if from == to {
		return true
	}

	// 检查是否存在合法的转换路径
	allowedStates, exists := sm.validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}

	return false
}

// ValidateTransition 验证状态转换，返回错误信息
func (sm *ApprovalStateMachine) ValidateTransition(from, to string) error {
	if !sm.CanTransition(from, to) {
		return fmt.Errorf("不允许从状态 '%s' 转换到 '%s'", from, to)
	}
	return nil
}

// GetNextState 根据审批决定获取下一个状态
func (sm *ApprovalStateMachine) GetNextState(currentStatus string, decision string) (string, error) {
	switch decision {
	case models.ApprovalDecisionApprove:
		// 审批通过，检查是否有下一阶段
		if currentStatus == models.ApprovalStatusUnderReview ||
		   currentStatus == models.ApprovalStatusResubmitted {
			// 这里应该检查是否有多级审批
			// 暂时直接设为已通过
			return models.ApprovalStatusApproved, nil
		}
		return currentStatus, nil

	case models.ApprovalDecisionReject:
		return models.ApprovalStatusRejected, nil

	case models.ApprovalDecisionRequestChanges:
		return models.ApprovalStatusNeedsRevision, nil

	case models.ApprovalDecisionReassign:
		// 转派不改变状态，但需要重新分配审批人
		return currentStatus, nil

	case models.ApprovalDecisionDefer:
		// 延期处理，保持在审核状态
		return currentStatus, nil

	case models.ApprovalDecisionEscalate:
		// 升级，保持在审核状态但通知上级
		return currentStatus, nil

	default:
		return "", errors.New("未知的审批决定类型")
	}
}

// IsFinalState 判断是否为终态（不可再变更）
func (sm *ApprovalStateMachine) IsFinalState(status string) bool {
	return status == models.ApprovalStatusApproved ||
		status == models.ApprovalStatusCancelled ||
		status == models.ApprovalStatusExpired
}

// CanCancel 判断是否可以撤回
func (sm *ApprovalStateMachine) CanCancel(status string, hasRecords bool) bool {
	// 只有已提交且无审批记录时可以撤回
	if status == models.ApprovalStatusSubmitted && !hasRecords {
		return true
	}
	// 审核中但未开始处理也可以撤回
	if status == models.ApprovalStatusUnderReview && !hasRecords {
		return true
	}
	return false
}

// CanResubmit 判断是否可以重新提交
func (sm *ApprovalStateMachine) CanResubmit(status string) bool {
	return status == models.ApprovalStatusRejected ||
		status == models.ApprovalStatusNeedsRevision
}

// CanEdit 判断是否可以编辑
func (sm *ApprovalStateMachine) CanEdit(status string) bool {
	return status == models.ApprovalStatusDraft ||
		status == models.ApprovalStatusNeedsRevision
}

// GetStatusDisplayName 获取状态的显示名称
func (sm *ApprovalStateMachine) GetStatusDisplayName(status string) string {
	displayNames := map[string]string{
		models.ApprovalStatusDraft:         "草稿",
		models.ApprovalStatusSubmitted:     "已提交",
		models.ApprovalStatusUnderReview:   "审核中",
		models.ApprovalStatusApproved:      "已通过",
		models.ApprovalStatusRejected:      "已拒绝",
		models.ApprovalStatusCancelled:     "已撤回",
		models.ApprovalStatusExpired:       "已过期",
		models.ApprovalStatusNeedsRevision: "需要修改",
		models.ApprovalStatusResubmitted:   "重新提交",
	}

	if name, exists := displayNames[status]; exists {
		return name
	}
	return "未知状态"
}

// GetStatusDescription 获取状态的描述信息
func (sm *ApprovalStateMachine) GetStatusDescription(status string) string {
	descriptions := map[string]string{
		models.ApprovalStatusDraft:         "申请尚未提交，可以编辑或删除",
		models.ApprovalStatusSubmitted:     "申请已提交，等待审批人处理",
		models.ApprovalStatusUnderReview:   "审批人正在审核中",
		models.ApprovalStatusApproved:      "申请已通过审批",
		models.ApprovalStatusRejected:      "申请被拒绝，可以修改后重新提交",
		models.ApprovalStatusCancelled:     "申请已被申请人撤回",
		models.ApprovalStatusExpired:       "申请已过期",
		models.ApprovalStatusNeedsRevision: "需要修改后重新提交",
		models.ApprovalStatusResubmitted:   "修改后重新提交，等待审批",
	}

	if desc, exists := descriptions[status]; exists {
		return desc
	}
	return ""
}
