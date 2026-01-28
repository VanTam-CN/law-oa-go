package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"gorm.io/gorm"
)

// ApprovalAssigner 审批人分配器
// 负责根据工作流配置自动分配审批人
type ApprovalAssigner struct {
	approvalRepo *repositories.ApprovalRepository
	workflowRepo *repositories.ApprovalWorkflowRepository
	userRepo     repositories.UserRepository
}

// NewApprovalAssigner 创建审批人分配器
func NewApprovalAssigner(
	db *gorm.DB,
	approvalRepo *repositories.ApprovalRepository,
	userRepo repositories.UserRepository,
) *ApprovalAssigner {
	return &ApprovalAssigner{
		approvalRepo: approvalRepo,
		workflowRepo: repositories.NewApprovalWorkflowRepository(db),
		userRepo:     userRepo,
	}
}

// AssignApprover 为审批申请分配审批人
func (a *ApprovalAssigner) AssignApprover(approval *models.ApprovalRequest) error {
	if approval.WorkflowType == "" {
		return errors.New("工作流类型不能为空")
	}

	// 获取工作流配置
	workflow, err := a.workflowRepo.FindByType(approval.WorkflowType)
	if err != nil {
		return fmt.Errorf("获取工作流失败: %v", err)
	}

	if workflow == nil {
		// 如果没有找到工作流，使用默认分配逻辑
		return a.assignDefaultApprover(approval)
	}

	// 解析工作流阶段配置
	var stages []WorkflowStage
	if err := json.Unmarshal([]byte(workflow.Stages), &stages); err != nil {
		return fmt.Errorf("解析工作流阶段失败: %v", err)
	}

	if len(stages) == 0 {
		return a.assignDefaultApprover(approval)
	}

	// 获取第一个阶段
	firstStage := stages[0]

	// 根据阶段配置分配审批人
	approver, err := a.findApproverForStage(&firstStage, approval)
	if err != nil {
		return fmt.Errorf("分配审批人失败: %v", err)
	}

	// 更新审批申请
	approval.CurrentStage = firstStage.StageName
	approval.CurrentApproverID = approver.ID
	approval.CurrentApproverName = approver.Name
	approval.Status = models.ApprovalStatusSubmitted

	return nil
}

// assignDefaultApprover 使用默认逻辑分配审批人
func (a *ApprovalAssigner) assignDefaultApprover(approval *models.ApprovalRequest) error {
	approval.CurrentStage = "initial_review"

	// 尝试查找管理员用户（排除申请人自己）
	users, err := a.userRepo.FindByRole("admin", 10)
	if err != nil || len(users) == 0 {
		// 如果没有找到管理员，尝试获取任何用户（除了申请人）
		// 这里简化处理，返回错误而不是使用硬编码ID
		return fmt.Errorf("未找到可用的审批人")
	}

	// 查找第一个不是申请人的管理员
	for _, user := range users {
		userIDStr := strconv.FormatUint(uint64(user.ID), 10)
		// 如果申请人和管理员是同一个人，跳过
		if userIDStr == approval.ApplicantID {
			continue
		}
		approval.CurrentApproverID = userIDStr
		approval.CurrentApproverName = user.Name
		approval.Status = models.ApprovalStatusSubmitted
		return nil
	}

	// 如果所有管理员都是申请人自己，使用第一个管理员
	// 这意味着管理员需要由其他管理员审批（双重审批）
	if len(users) > 0 {
		userIDStr := strconv.FormatUint(uint64(users[0].ID), 10)
		approval.CurrentApproverID = userIDStr
		approval.CurrentApproverName = users[0].Name
		approval.Status = models.ApprovalStatusSubmitted
		return nil
	}

	return fmt.Errorf("未找到可用的审批人")
}

// findApproverForStage 根据阶段配置查找审批人
func (a *ApprovalAssigner) findApproverForStage(stage *WorkflowStage, approval *models.ApprovalRequest) (*ApproverInfo, error) {
	// 根据审批人角色查找
	if stage.ApproverRole != "" {
		return a.findApproverByRole(stage.ApproverRole, approval)
	}

	// 如果指定了具体审批人ID
	if stage.ApproverID != "" {
		return a.findApproverByID(stage.ApproverID)
	}

	// 根据部门查找
	if stage.DepartmentBased {
		return a.findApproverByDepartment(approval.DepartmentID, stage)
	}

	// 默认查找管理员用户
	users, err := a.userRepo.FindByRole("admin", 1)
	if err != nil || len(users) == 0 {
		return nil, fmt.Errorf("未找到可用的审批人")
	}

	return &ApproverInfo{
		ID:   strconv.FormatUint(uint64(users[0].ID), 10),
		Name: users[0].Name,
	}, nil
}

// findApproverByRole 根据角色查找审批人
func (a *ApprovalAssigner) findApproverByRole(role string, approval *models.ApprovalRequest) (*ApproverInfo, error) {
	// 根据角色查找用户
	// 这里简化处理，实际应该查询用户角色表
	users, err := a.userRepo.FindByRole(role, 1)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("没有找到角色为 %s 的审批人", role)
	}

	return &ApproverInfo{
		ID:   strconv.FormatUint(uint64(users[0].ID), 10),
		Name: users[0].Name,
	}, nil
}

// findApproverByID 根据ID查找审批人
func (a *ApprovalAssigner) findApproverByID(userID string) (*ApproverInfo, error) {
	user, err := a.userRepo.FindByStringID(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("审批人不存在: %s", userID)
	}

	return &ApproverInfo{
		ID:   strconv.FormatUint(uint64(user.ID), 10),
		Name: user.Name,
	}, nil
}

// findApproverByDepartment 根据部门查找审批人
func (a *ApprovalAssigner) findApproverByDepartment(deptID string, stage *WorkflowStage) (*ApproverInfo, error) {
	if deptID == "" {
		return nil, errors.New("部门ID为空")
	}

	// 查找部门主管
	users, err := a.userRepo.FindDepartmentHead(deptID, 1)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		// 如果没有部门主管，返回管理员
		return &ApproverInfo{
			ID:   "1",
			Name: "系统管理员",
		}, nil
	}

	return &ApproverInfo{
		ID:   strconv.FormatUint(uint64(users[0].ID), 10),
		Name: users[0].Name,
	}, nil
}

// MoveToNextStage 移动到下一个审批阶段
func (a *ApprovalAssigner) MoveToNextStage(approval *models.ApprovalRequest) (bool, error) {
	if approval.WorkflowType == "" {
		// 没有工作流类型，直接通过
		approval.Status = models.ApprovalStatusApproved
		approval.CurrentApproverID = ""
		approval.CurrentApproverName = ""
		return true, nil
	}

	workflow, err := a.workflowRepo.FindByType(approval.WorkflowType)
	if err != nil {
		return false, err
	}

	if workflow == nil {
		// 没有工作流，直接通过
		approval.Status = models.ApprovalStatusApproved
		approval.CurrentApproverID = ""
		approval.CurrentApproverName = ""
		return true, nil
	}

	// 解析工作流阶段
	var stages []WorkflowStage
	if err := json.Unmarshal([]byte(workflow.Stages), &stages); err != nil {
		return false, err
	}

	// 查找当前阶段索引
	currentStageIndex := -1
	for i, stage := range stages {
		if stage.StageName == approval.CurrentStage {
			currentStageIndex = i
			break
		}
	}

	// 如果是最后一个阶段，审批完成
	if currentStageIndex == len(stages)-1 || currentStageIndex == -1 {
		approval.Status = models.ApprovalStatusApproved
		approval.CurrentApproverID = ""
		approval.CurrentApproverName = ""
		approval.CurrentStage = "completed"
		return true, nil
	}

	// 移动到下一个阶段
	nextStage := stages[currentStageIndex+1]
	approver, err := a.findApproverForStage(&nextStage, approval)
	if err != nil {
		return false, err
	}

	approval.CurrentStage = nextStage.StageName
	approval.CurrentApproverID = approver.ID
	approval.CurrentApproverName = approver.Name
	approval.Status = models.ApprovalStatusUnderReview

	return false, nil
}

// WorkflowStage 工作流阶段
type WorkflowStage struct {
	StageName     string `json:"stage_name"`
	StageOrder    int    `json:"stage_order"`
	ApproverRole  string `json:"approver_role"`
	ApproverID    string `json:"approver_id"`
	Required      bool   `json:"required"`
	TimeoutHours  int    `json:"timeout_hours"`
	DepartmentBased bool `json:"department_based"`
}

// ApproverInfo 审批人信息
type ApproverInfo struct {
	ID   string
	Name string
}
