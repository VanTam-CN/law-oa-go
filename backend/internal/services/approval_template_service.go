package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"gorm.io/gorm"
)

// ApprovalTemplateService 审批模板服务
type ApprovalTemplateService struct {
	templateRepo *repositories.ApprovalTemplateRepository
	nodeRepo     *repositories.ApprovalNodeRepository
	approvalRepo *repositories.ApprovalRepository
	userRepo     repositories.UserRepository
	db           *gorm.DB
}

// NewApprovalTemplateService 创建审批模板服务
func NewApprovalTemplateService(db *gorm.DB) *ApprovalTemplateService {
	templateRepo := repositories.NewApprovalTemplateRepository(db)
	nodeRepo := repositories.NewApprovalNodeRepository(db)
	approvalRepo := repositories.NewApprovalRepository(db)
	userRepo := repositories.NewUserRepository(db)

	return &ApprovalTemplateService{
		templateRepo: templateRepo,
		nodeRepo:     nodeRepo,
		approvalRepo: approvalRepo,
		userRepo:     userRepo,
		db:           db,
	}
}

// InitializeDefaultTemplates 初始化默认模板
func (s *ApprovalTemplateService) InitializeDefaultTemplates() error {
	// 检查是否已存在模板
	if templates, _ := s.templateRepo.FindAll(); len(templates) > 0 {
		return nil // 已有模板，无需初始化
	}

	// 创建用印审批模板
	sealApprovalSteps := []models.ApprovalStep{
		{
			Order:        1,
			Name:         "部门主任审批",
			ApproverType: models.ApproverTypeDepartmentHead,
			IsRequired:   true,
			SignType:     models.SignTypeSingle,
			AutoPass:     false,
		},
		{
			Order:        2,
			Name:         "执行合伙人审批",
			ApproverType: models.ApproverTypeRole,
			ApproverRole: "EXECUTIVE_PARTNER",
			IsRequired:   false,
			SignType:     models.SignTypeSingle,
			AutoPass:     false,
		},
	}
	sealConditions := []models.ApprovalCondition{
		{
			Field:    "amount",
			Operator: "gt",
			Value:    1000000, // 100万
			ThenStep: 2,       // 需要执行合伙人审批
			ElseStep: 1,       // 自动跳过
		},
	}
	sealStepsJSON, _ := json.Marshal(sealApprovalSteps)
	sealConditionsJSON, _ := json.Marshal(sealConditions)

	sealTemplate := &models.ApprovalTemplate{
		Name:        models.TemplateSealApproval,
		DisplayName: "用印审批",
		Description: "用于公司印章使用的审批流程，标的额超过100万需要执行合伙人审批",
		Steps:       string(sealStepsJSON),
		Conditions:  string(sealConditionsJSON),
		IsActive:    true,
	}

	// 创建立案审批模板
	caseFilingSteps := []models.ApprovalStep{
		{
			Order:        1,
			Name:         "部门主任审批",
			ApproverType: models.ApproverTypeDepartmentHead,
			IsRequired:   true,
			SignType:     models.SignTypeSingle,
		},
		{
			Order:        2,
			Name:         "合伙人审批",
			ApproverType: models.ApproverTypeRole,
			ApproverRole: "PARTNER",
			IsRequired:   true,
			SignType:     models.SignTypeSingle,
		},
	}
	caseStepsJSON, _ := json.Marshal(caseFilingSteps)

	caseTemplate := &models.ApprovalTemplate{
		Name:        models.TemplateCaseFiling,
		DisplayName: "立案审批",
		Description: "用于新案件立案的审批流程，需要部门主任和合伙人依次审批",
		Steps:       string(caseStepsJSON),
		Conditions:  "[]",
		IsActive:    true,
	}

	// 保存模板
	if err := s.templateRepo.Create(sealTemplate); err != nil {
		return fmt.Errorf("创建用印审批模板失败: %w", err)
	}
	if err := s.templateRepo.Create(caseTemplate); err != nil {
		return fmt.Errorf("创建立案审批模板失败: %w", err)
	}

	return nil
}

// CreateFromTemplate 从模板创建审批
func (s *ApprovalTemplateService) CreateFromTemplate(templateName string, userID string, userName string, req *models.CreateApprovalRequest, formData map[string]interface{}) (*models.ApprovalRequest, error) {
	// 查找模板
	template, err := s.templateRepo.FindByName(templateName)
	if err != nil {
		return nil, fmt.Errorf("查找审批模板失败: %w", err)
	}
	if template == nil {
		return nil, errors.New("审批模板不存在或已停用")
	}

	// 获取模板步骤
	steps, err := s.templateRepo.GetSteps(template)
	if err != nil {
		return nil, fmt.Errorf("解析审批步骤失败: %w", err)
	}
	if len(steps) == 0 {
		return nil, errors.New("审批模板没有配置步骤")
	}

	// 设置工作流类型为模板名
	req.WorkflowType = templateName

	// 创建审批申请
	approvalService := NewApprovalService(s.db)
	approval, err := approvalService.CreateApproval(userID, userName, req)
	if err != nil {
		return nil, fmt.Errorf("创建审批申请失败: %w", err)
	}

	// 创建审批节点
	nodes := s.createNodesFromSteps(approval.ID, steps, formData)
	if len(nodes) == 0 {
		return nil, errors.New("创建审批节点失败")
	}

	// 保存节点
	if err := s.nodeRepo.CreateBatch(nodes); err != nil {
		return nil, fmt.Errorf("保存审批节点失败: %w", err)
	}

	// 设置当前审批人
	firstNode := nodes[0]
	approval.CurrentStage = firstNode.NodeName
	if firstNode.ApproverID != nil {
		approval.CurrentApproverID = *firstNode.ApproverID
		// 获取审批人姓名
		if approver, err := s.userRepo.FindByStringID(*firstNode.ApproverID); err == nil && approver != nil {
			approval.CurrentApproverName = approver.Name
		}
	}
	approval.UpdatedBy = userID
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, fmt.Errorf("更新审批信息失败: %w", err)
	}

	return approval, nil
}

// createNodesFromSteps 从步骤创建节点
func (s *ApprovalTemplateService) createNodesFromSteps(approvalID string, steps []models.ApprovalStep, formData map[string]interface{}) []models.ApprovalNode {
	var nodes []models.ApprovalNode

	for _, step := range steps {
		// 检查条件是否满足自动跳过
		if step.AutoPass && s.shouldSkipStep(step, formData) {
			continue
		}

		node := models.ApprovalNode{
			ApprovalID:   approvalID,
			StepOrder:    step.Order,
			NodeName:     step.Name,
			ApproverType: step.ApproverType,
			ApproverRole: step.ApproverRole,
			Action:       models.NodeActionPending,
			SignType:     step.SignType,
			IsFinal:      step.Order == len(steps), // 最后一步是最终节点
		}

		// 根据审批人类型设置审批人
		switch step.ApproverType {
		case models.ApproverTypeSpecificUser:
			if step.ApproverID > 0 {
				approverID := fmt.Sprintf("%d", step.ApproverID)
				node.ApproverID = &approverID
			}
		case models.ApproverTypeDepartmentHead:
			// 部门主任需要根据申请人部门动态获取
			// 这里先设置为空，后续在处理时动态分配
		case models.ApproverTypeRole:
			// 角色审批人，后续动态分配
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// shouldSkipStep 检查是否应该跳过步骤
func (s *ApprovalTemplateService) shouldSkipStep(step models.ApprovalStep, formData map[string]interface{}) bool {
	if !step.AutoPass {
		return false
	}
	// TODO: 实现条件判断逻辑
	return false
}

// ProcessNode 处理审批节点
func (s *ApprovalTemplateService) ProcessNode(nodeID uint, action string, comment string, approverID string) (*models.ApprovalNode, error) {
	// 查找节点
	node, err := s.nodeRepo.FindByID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("查找审批节点失败: %w", err)
	}
	if node == nil {
		return nil, errors.New("审批节点不存在")
	}

	// 检查权限
	if node.ApproverID != nil && *node.ApproverID != approverID {
		return nil, errors.New("无权处理此审批节点")
	}

	// 检查节点状态
	if !node.CanProcess() {
		return nil, errors.New("节点已处理，无法重复操作")
	}

	// 更新节点状态
	switch action {
	case "approve":
		node.MarkAsApproved(comment)
	case "reject":
		node.MarkAsRejected(comment)
	case "transfer":
		node.MarkAsTransferred(comment)
	default:
		return nil, errors.New("无效的审批操作")
	}

	// 计算处理时长
	node.Duration = int(time.Since(node.CreatedAt).Seconds())

	// 保存更新
	if err := s.nodeRepo.Update(node); err != nil {
		return nil, fmt.Errorf("更新审批节点失败: %w", err)
	}

	// 如果审批通过，处理下一步
	if action == "approve" {
		if err := s.processNextStep(node); err != nil {
			return nil, fmt.Errorf("处理下一步失败: %w", err)
		}
	}

	// 如果审批拒绝，更新审批状态
	if action == "reject" {
		if err := s.updateApprovalStatus(node.ApprovalID, models.ApprovalStatusRejected); err != nil {
			return nil, fmt.Errorf("更新审批状态失败: %w", err)
		}
	}

	return node, nil
}

// processNextStep 处理下一步
func (s *ApprovalTemplateService) processNextStep(node *models.ApprovalNode) error {
	// 查找同一审批的所有节点
	nodes, err := s.nodeRepo.FindByApprovalID(node.ApprovalID)
	if err != nil {
		return err
	}

	// 检查当前步骤是否所有节点都已通过（会签场景）
	currentStepNodes := filterNodesByStep(nodes, node.StepOrder)
	if allNodesApproved(currentStepNodes) {
		// 当前步骤完成，移动到下一步
		nextStepNodes := filterNodesByStep(nodes, node.StepOrder+1)
		if len(nextStepNodes) > 0 {
			// 有下一步，更新审批的当前审批人
			return s.updateCurrentApprover(node.ApprovalID, nextStepNodes[0])
		} else {
			// 没有下一步，审批完成
			return s.updateApprovalStatus(node.ApprovalID, models.ApprovalStatusApproved)
		}
	}

	return nil
}

// updateCurrentApprover 更新当前审批人
func (s *ApprovalTemplateService) updateCurrentApprover(approvalID string, nextNode models.ApprovalNode) error {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return err
	}

	approval.CurrentStage = nextNode.NodeName
	if nextNode.ApproverID != nil {
		approval.CurrentApproverID = *nextNode.ApproverID
		// 获取审批人姓名
		if approver, err := s.userRepo.FindByStringID(*nextNode.ApproverID); err == nil && approver != nil {
			approval.CurrentApproverName = approver.Name
		}
	}

	return s.approvalRepo.Update(approval)
}

// updateApprovalStatus 更新审批状态
func (s *ApprovalTemplateService) updateApprovalStatus(approvalID string, status string) error {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return err
	}

	approval.Status = status
	if status == models.ApprovalStatusApproved {
		approval.CurrentApproverID = ""
		approval.CurrentApproverName = ""
	}

	return s.approvalRepo.Update(approval)
}

// GetApprovalFlow 获取审批流程可视化数据
func (s *ApprovalTemplateService) GetApprovalFlow(approvalID string) (map[string]interface{}, error) {
	// 查找审批
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, err
	}

	// 查找所有节点
	nodes, err := s.nodeRepo.FindByApprovalID(approvalID)
	if err != nil {
		return nil, err
	}

	// 构建流程数据
	flow := map[string]interface{}{
		"approval_id":      approvalID,
		"request_number":   approval.RequestNumber,
		"title":            approval.Title,
		"status":           approval.Status,
		"current_stage":    approval.CurrentStage,
		"current_approver": approval.CurrentApproverName,
		"nodes":            nodes,
		"total_nodes":      len(nodes),
		"approved_nodes":   countNodesByAction(nodes, models.NodeActionApproved),
		"pending_nodes":    countNodesByAction(nodes, models.NodeActionPending),
		"rejected_nodes":   countNodesByAction(nodes, models.NodeActionRejected),
	}

	return flow, nil
}

// GetAllTemplates 获取所有模板
func (s *ApprovalTemplateService) GetAllTemplates() ([]models.ApprovalTemplate, error) {
	return s.templateRepo.FindAll()
}

// SupportCountersign 支持会签
func (s *ApprovalTemplateService) SupportCountersign(approvalID string, stepOrder int) error {
	nodes, err := s.nodeRepo.GetStepNodes(approvalID, stepOrder)
	if err != nil {
		return err
	}

	// 设置为会签类型
	for i := range nodes {
		nodes[i].SignType = models.SignTypeCountersign
		if err := s.nodeRepo.Update(&nodes[i]); err != nil {
			return err
		}
	}

	return nil
}

// SupportOrSign 支持或签
func (s *ApprovalTemplateService) SupportOrSign(approvalID string, stepOrder int) error {
	nodes, err := s.nodeRepo.GetStepNodes(approvalID, stepOrder)
	if err != nil {
		return err
	}

	// 设置为或签类型
	for i := range nodes {
		nodes[i].SignType = models.SignTypeOrSign
		if err := s.nodeRepo.Update(&nodes[i]); err != nil {
			return err
		}
	}

	return nil
}

// ReturnToPrevious 退回到上一步
func (s *ApprovalTemplateService) ReturnToPrevious(approvalID string, stepOrder int) error {
	nodes, err := s.nodeRepo.FindByApprovalID(approvalID)
	if err != nil {
		return err
	}

	// 重置当前步骤的节点状态
	currentNodes := filterNodesByStep(nodes, stepOrder)
	for i := range currentNodes {
		currentNodes[i].Action = models.NodeActionPending
		currentNodes[i].Comment = ""
		if err := s.nodeRepo.Update(&currentNodes[i]); err != nil {
			return err
		}
	}

	// 恢复上一步节点状态
	previousNodes := filterNodesByStep(nodes, stepOrder-1)
	if len(previousNodes) > 0 {
		for i := range previousNodes {
			previousNodes[i].Action = models.NodeActionPending
			previousNodes[i].Comment = ""
			if err := s.nodeRepo.Update(&previousNodes[i]); err != nil {
				return err
			}
		}

		// 更新审批的当前审批人
		return s.updateCurrentApprover(approvalID, previousNodes[0])
	}

	return nil
}

// 辅助函数

func filterNodesByStep(nodes []models.ApprovalNode, stepOrder int) []models.ApprovalNode {
	var result []models.ApprovalNode
	for _, n := range nodes {
		if n.StepOrder == stepOrder {
			result = append(result, n)
		}
	}
	return result
}

func allNodesApproved(nodes []models.ApprovalNode) bool {
	for _, n := range nodes {
		if n.SignType == models.SignTypeOrSign {
			// 或签：任意一个通过即可
			if n.IsApproved() {
				return true
			}
		} else {
			// 会签/单签：所有都要通过
			if !n.IsApproved() {
				return false
			}
		}
	}
	return len(nodes) > 0
}

func countNodesByAction(nodes []models.ApprovalNode, action string) int {
	count := 0
	for _, n := range nodes {
		if n.Action == action {
			count++
		}
	}
	return count
}
