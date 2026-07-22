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
	// 审批节点处理器
	stageProcessor *ApprovalStageProcessor
}

// NewApprovalStateMachine 创建新的状态机实例
func NewApprovalStateMachine() *ApprovalStateMachine {
	sm := &ApprovalStateMachine{
		validTransitions: make(map[string][]string),
		stageProcessor:   NewApprovalStageProcessor(),
	}
	sm.initializeTransitions()
	return sm
}

// ApprovalStageProcessor 审批节点处理器
type ApprovalStageProcessor struct {
	// 当前节点配置
	CurrentStage *models.ApprovalStage
	// 已完成的节点
	CompletedStages []string
	// 会签/或签状态
	ApprovalState map[string]*StageApprovalState
}

// StageApprovalState 节点审批状态
type StageApprovalState struct {
	StageKey        string
	TotalApprovers  int
	ApprovedCount   int
	RejectedCount   int
	ApproverRecords map[string]bool // approverID -> approved
}

// NewApprovalStageProcessor 创建节点处理器
func NewApprovalStageProcessor() *ApprovalStageProcessor {
	return &ApprovalStageProcessor{
		CompletedStages: make([]string, 0),
		ApprovalState:   make(map[string]*StageApprovalState),
	}
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

// ProcessStageDecision 处理节点审批决定（支持会签/或签）
func (sp *ApprovalStageProcessor) ProcessStageDecision(
	stage *models.ApprovalStage,
	approverID string,
	decision string,
) (shouldComplete bool, shouldReject bool, nextAction string, err error) {
	if stage == nil {
		return false, false, "", errors.New("审批节点不存在")
	}

	// 初始化节点状态
	if _, exists := sp.ApprovalState[stage.StageKey]; !exists {
		sp.ApprovalState[stage.StageKey] = &StageApprovalState{
			StageKey:        stage.StageKey,
			TotalApprovers:  len(stage.Approvers),
			ApprovedCount:   0,
			RejectedCount:   0,
			ApproverRecords: make(map[string]bool),
		}
	}

	state := sp.ApprovalState[stage.StageKey]

	// 检查是否已审批
	if _, approved := state.ApproverRecords[approverID]; approved {
		return false, false, "", errors.New("该审批人已处理过此节点")
	}

	// 记录审批
	state.ApproverRecords[approverID] = true

	switch decision {
	case models.ApprovalDecisionApprove:
		state.ApprovedCount++

		// 根据审批模式判断是否完成
		if stage.ApprovalMode == "or" {
			// 或签：一人通过即可
			return true, false, "next_stage", nil
		}
		// 会签：全部通过
		if state.ApprovedCount >= state.TotalApprovers {
			return true, false, "next_stage", nil
		}
		return false, false, "waiting_others", nil

	case models.ApprovalDecisionReject:
		state.RejectedCount++

		// 任何模式下，有人拒绝即拒绝
		if stage.ApprovalMode == "or" {
			// 或签：一人拒绝即拒绝
			return false, true, "rejected", nil
		}
		// 会签：一人拒绝即拒绝
		return false, true, "rejected", nil

	case models.ApprovalDecisionRequestChanges:
		// 要求修改直接结束流程
		return false, false, "needs_revision", nil

	case models.ApprovalDecisionReassign:
		// 转签不改变计数
		return false, false, "reassigned", nil

	case models.ApprovalDecisionDefer:
		// 延期不改变状态
		return false, false, "deferred", nil

	case models.ApprovalDecisionEscalate:
		// 升级
		return false, false, "escalated", nil

	default:
		return false, false, "", errors.New("未知的审批决定类型")
	}
}

// GetNextStage 获取下一个审批节点
func (sp *ApprovalStageProcessor) GetNextStage(
	currentStageKey string,
	templateConfig *models.ApprovalTemplateConfig,
	metadata map[string]interface{},
) (*models.ApprovalStage, error) {
	// 查找当前节点
	var currentStage *models.ApprovalStage
	var currentIndex = -1

	for i, stage := range templateConfig.Stages {
		if stage.StageKey == currentStageKey {
			currentStage = &stage
			currentIndex = i
			break
		}
	}

	if currentStage == nil {
		return nil, errors.New("当前节点不存在")
	}

	// 如果是条件节点，评估条件
	if currentStage.StageType == "conditional" {
		for _, condition := range templateConfig.Conditions {
			if condition.ThenStageKey == currentStage.StageKey ||
				condition.ElseStageKey == currentStage.StageKey {
				// 评估条件表达式
				if sp.evaluateCondition(condition.Expression, metadata) {
					// 满足条件，跳转到 then_stage
					return sp.findStageByKey(condition.ThenStageKey, templateConfig.Stages)
				} else if condition.ElseStageKey != "" && condition.ElseStageKey != "end" {
					// 不满足条件，跳转到 else_stage
					return sp.findStageByKey(condition.ElseStageKey, templateConfig.Stages)
				} else {
					// 流程结束
					return nil, nil
				}
			}
		}
	}

	// 串行/并行节点：返回下一个节点
	if currentIndex+1 < len(templateConfig.Stages) {
		return &templateConfig.Stages[currentIndex+1], nil
	}

	// 没有更多节点，流程结束
	return nil, nil
}

// findStageByKey 根据 key 查找节点
func (sp *ApprovalStageProcessor) findStageByKey(key string, stages []models.ApprovalStage) (*models.ApprovalStage, error) {
	for i, stage := range stages {
		if stage.StageKey == key {
			return &stages[i], nil
		}
	}
	return nil, errors.New("节点不存在: " + key)
}

// evaluateCondition 评估条件表达式
func (sp *ApprovalStageProcessor) evaluateCondition(expression string, metadata map[string]interface{}) bool {
	// 简化版条件评估
	// 实际项目中应该使用更安全的表达式解析器
	// 这里只做基本演示

	// 检查标的额条件
	if val, ok := metadata["case_value"]; ok {
		caseValue, _ := val.(float64)
		if expression == "case_value >= 1000000" {
			return caseValue >= 1000000
		}
		if expression == "case_value >= 100000" {
			return caseValue >= 100000
		}
	}

	// 检查用印重要性条件
	if val, ok := metadata["seal_importance"]; ok {
		importance, _ := val.(string)
		if expression == "seal_importance == 'high' || seal_count >= 3" {
			return importance == "high"
		}
		if expression == "seal_importance == 'medium' || seal_count >= 1" {
			return importance == "medium" || importance == "high"
		}
	}

	// 默认返回 false
	return false
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

// CanReturnToPrevious 判断是否可以退回上一步
func (sm *ApprovalStateMachine) CanReturnToPrevious(status string, currentStage string, templateConfig *models.ApprovalTemplateConfig) bool {
	if status != models.ApprovalStatusUnderReview {
		return false
	}

	// 查找当前节点
	for i, stage := range templateConfig.Stages {
		if stage.StageKey == currentStage {
			// 第一个节点不能退回
			if i == 0 {
				return false
			}
			// 检查上一个节点是否允许退回
			prevStage := templateConfig.Stages[i-1]
			return prevStage.AllowReturn
		}
	}

	return false
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
