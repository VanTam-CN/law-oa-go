package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"gorm.io/gorm"
)

type ApprovalService struct {
	approvalRepo *repositories.ApprovalRepository
}

func NewApprovalService(db *gorm.DB) *ApprovalService {
	return &ApprovalService{
		approvalRepo: repositories.NewApprovalRepository(db),
	}
}

// GetApprovalStats 获取审批统计
func (s *ApprovalService) GetApprovalStats(userID string) (*models.ApprovalStats, error) {
	var stats models.ApprovalStats

	// 获取总申请数
	totalRequests, err := s.approvalRepo.CountByUserID(userID, "")
	if err != nil {
		return nil, fmt.Errorf("获取总申请数失败: %v", err)
	}
	stats.TotalRequests = int(totalRequests)

	// 获取待处理审批数（需要当前用户审批的）
	myPendingRequests, err := s.approvalRepo.CountPendingByApproverID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取待处理审批数失败: %v", err)
	}
	stats.MyPendingRequests = int(myPendingRequests)

	// 获取全部待审批数
	pendingRequests, err := s.approvalRepo.CountPending()
	if err != nil {
		return nil, fmt.Errorf("获取全部待审批数失败: %v", err)
	}
	stats.PendingRequests = int(pendingRequests)

	// 获取已通过申请数
	approvedRequests, err := s.approvalRepo.CountByUserID(userID, models.ApprovalStatusApproved)
	if err != nil {
		return nil, fmt.Errorf("获取已通过申请数失败: %v", err)
	}
	stats.ApprovedRequests = int(approvedRequests)

	// 获取已拒绝申请数
	rejectedRequests, err := s.approvalRepo.CountByUserID(userID, models.ApprovalStatusRejected)
	if err != nil {
		return nil, fmt.Errorf("获取已拒绝申请数失败: %v", err)
	}
	stats.RejectedRequests = int(rejectedRequests)

	return &stats, nil
}

// GetPendingApprovals 获取待审批列表
func (s *ApprovalService) GetPendingApprovals(userID string, req *models.ApprovalListRequest) (*models.ApprovalListResponse, error) {
	// 使用repository获取待审批列表
	approvals, err := s.approvalRepo.FindPendingByApproverID(userID, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取待审批列表失败: %v", err)
	}

	// 获取总数
	total, err := s.approvalRepo.CountPendingByApproverID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取待审批总数失败: %v", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &models.ApprovalListResponse{
		List:       approvals,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListApprovals 获取审批列表
func (s *ApprovalService) ListApprovals(userID string, req *models.ApprovalListRequest) (*models.ApprovalListResponse, error) {
	// 修复：使用req.ApplicantID参数而不是userID来筛选申请人的审批
	queryUserID := userID // 默认使用当前用户ID
	if req.ApplicantID != "" {
		queryUserID = req.ApplicantID // 如果指定了申请人ID，则使用指定的ID
	}

	// 使用repository获取用户的审批列表
	approvals, err := s.approvalRepo.FindByUserID(queryUserID, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取审批列表失败: %v", err)
	}

	// 获取总数
	total, err := s.approvalRepo.CountByUserID(queryUserID, req.Status)
	if err != nil {
		return nil, fmt.Errorf("获取审批总数失败: %v", err)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	return &models.ApprovalListResponse{
		List:       approvals,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetApproval 获取单个审批详情
func (s *ApprovalService) GetApproval(userID string, id string) (*models.ApprovalRequest, error) {
	// 使用repository获取审批详情
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取审批详情失败: %v", err)
	}

	// 检查权限：确保用户有权限查看（自己申请的或自己需要审批的）
	if approval != nil {
		if approval.ApplicantID != userID && approval.CurrentApproverID != userID {
			return nil, fmt.Errorf("无权查看此审批记录")
		}
	}

	return approval, nil
}

// CreateApproval 创建审批申请
func (s *ApprovalService) CreateApproval(userID string, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error) {
	// 生成申请编号
	requestNumber, err := s.generateRequestNumber()
	if err != nil {
		return nil, fmt.Errorf("生成申请编号失败: %v", err)
	}

	// 获取用户部门信息（这里简化处理，实际应该从用户服务获取）
	departmentID := ""
	departmentName := ""

	// 设置默认值
	if req.Urgency == "" {
		req.Urgency = "normal"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	// 创建审批申请 - 直接设置为已提交状态
	approval := &models.ApprovalRequest{
		ID:                   generateUUID(),
		RequestNumber:        requestNumber,
		Title:                req.Title,
		Type:                 req.Type,
		Category:             req.Category,
		Content:              req.Content,
		ApplicantID:          userID,
		ApplicantName:        userName,
		ApplicantTitle:       "",
		DepartmentID:         departmentID,
		DepartmentName:       departmentName,
		Urgency:              req.Urgency,
		Priority:             req.Priority,
		ExpectedEffectiveDate: nil,
		ExpectedExpiryDate:   nil,
		DurationDays:         req.DurationDays,
		Status:               models.ApprovalStatusSubmitted, // 直接设置为已提交状态
		SubmissionDate:       &time.Time{}, // 设置提交时间
		CurrentStage:         "initial_review",
		CurrentApproverID:    "",
		CurrentApproverName:  "",
		WorkflowType:         req.WorkflowType,
		WorkflowConfig:       "{}",
		Attachments:          "{}",
		Metadata:             "{}",
		CreatedBy:            userID,
		UpdatedBy:            userID,
	}

	// 设置提交时间为当前时间
	now := time.Now()
	approval.SubmissionDate = &now

	// 处理附件和元数据
	if len(req.Attachments) > 0 {
		attachmentsBytes, _ := json.Marshal(req.Attachments)
		approval.Attachments = string(attachmentsBytes)
	}
	if req.Metadata != nil {
		metadataBytes, _ := json.Marshal(req.Metadata)
		approval.Metadata = string(metadataBytes)
	}

	// 处理时间字段
	if req.ExpectedEffectiveDate != "" {
		if effectiveDate, err := time.Parse("2006-01-02", req.ExpectedEffectiveDate); err == nil {
			approval.ExpectedEffectiveDate = &effectiveDate
		}
	}
	if req.ExpectedExpiryDate != "" {
		if expiryDate, err := time.Parse("2006-01-02", req.ExpectedExpiryDate); err == nil {
			approval.ExpectedExpiryDate = &expiryDate
		}
	}

	// 保存到数据库
	if err := s.approvalRepo.Create(approval); err != nil {
		return nil, fmt.Errorf("保存审批申请失败: %v", err)
	}

	// 如果是立即提交，更新状态为已提交
	// 这里可以添加自动提交的逻辑
	// err = s.SubmitApproval(userID, approval.ID)
	// if err != nil {
	//     return nil, fmt.Errorf("提交审批申请失败: %v", err)
	// }

	return approval, nil
}

// UpdateApproval 更新审批申请
func (s *ApprovalService) UpdateApproval(userID string, id string, req *models.UpdateApprovalRequest) (*models.ApprovalRequest, error) {
	// 使用repository查找审批记录
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权更新此审批记录")
	}

	// 只能更新草稿状态的审批
	if approval != nil && approval.Status != models.ApprovalStatusDraft {
		return nil, errors.New("只能更新草稿状态的审批")
	}

	// 更新字段
	if req.Title != "" {
		approval.Title = req.Title
	}
	if req.Content != "" {
		approval.Content = req.Content
	}
	// Urgency和Priority字段不在UpdateApprovalRequest中，暂时忽略

	// 处理附件和元数据
	if len(req.Attachments) > 0 {
		if attachmentsBytes, err := json.Marshal(req.Attachments); err == nil {
			approval.Attachments = string(attachmentsBytes)
		}
	}
	if req.Metadata != nil {
		if metadataBytes, err := json.Marshal(req.Metadata); err == nil {
			approval.Metadata = string(metadataBytes)
		}
	}

	approval.UpdatedBy = userID

	// 执行更新
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, fmt.Errorf("更新审批申请失败: %v", err)
	}

	// 重新加载数据
	updatedApproval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("重新加载审批数据失败: %v", err)
	}

	return updatedApproval, nil
}

// ProcessApproval 处理审批（通过/拒绝）
func (s *ApprovalService) ProcessApproval(userID string, userName string, id string, req *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	// 使用repository查找审批记录
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是当前审批人
	if approval != nil && approval.CurrentApproverID != userID {
		return nil, errors.New("无权审批此申请")
	}

	// 检查审批状态
	if approval != nil && (approval.Status != models.ApprovalStatusSubmitted && approval.Status != models.ApprovalStatusUnderReview) {
		return nil, errors.New("审批状态不允许处理")
	}

	// 创建审批记录
	record := &models.ApprovalRecord{
		ApprovalRequestID:   approval.ID,
		Stage:               approval.CurrentStage,
		StageOrder:          1, // 这里应该从工作流配置中获取
		ApproverID:          userID,
		ApproverName:        userName,
		ApproverTitle:       "",
		ApproverRole:        "",
		Decision:            req.Decision,
		DecisionReason:      req.DecisionReason,
		DecisionComments:    req.DecisionComments,
		ApprovedConditions:  "",
		ImposedRequirements: "",
		FollowUpActions:     "",
		ApprovalDate:        time.Now(),
		EffectiveDate:       nil,
		NextReviewDate:      nil,
		SupportingDocuments: "",
		EvidenceReferences:  "",
		Status:              "active",
	}

	// 处理JSON字段
	if req.ApprovedConditions != nil {
		if data, err := json.Marshal(req.ApprovedConditions); err == nil {
			record.ApprovedConditions = string(data)
		}
	}
	if req.ImposedRequirements != nil {
		if data, err := json.Marshal(req.ImposedRequirements); err == nil {
			record.ImposedRequirements = string(data)
		}
	}
	if len(req.FollowUpActions) > 0 {
		if data, err := json.Marshal(req.FollowUpActions); err == nil {
			record.FollowUpActions = string(data)
		}
	}
	if len(req.SupportingDocuments) > 0 {
		if data, err := json.Marshal(req.SupportingDocuments); err == nil {
			record.SupportingDocuments = string(data)
		}
	}
	if len(req.EvidenceReferences) > 0 {
		if data, err := json.Marshal(req.EvidenceReferences); err == nil {
			record.EvidenceReferences = string(data)
		}
	}

	// 保存审批记录
	if err := s.approvalRepo.CreateRecord(record); err != nil {
		return nil, fmt.Errorf("保存审批记录失败: %v", err)
	}

	// 更新审批申请状态
	switch req.Decision {
	case models.ApprovalDecisionApprove:
		// 审批通过，检查是否需要下一步审批
		// 这里简化处理，直接设置为已通过
		approval.Status = models.ApprovalStatusApproved
		approval.CurrentApproverID = ""
		approval.CurrentApproverName = ""
	case models.ApprovalDecisionReject:
		approval.Status = models.ApprovalStatusRejected
		approval.CurrentApproverID = ""
		approval.CurrentApproverName = ""
	case models.ApprovalDecisionRequestChanges:
		approval.Status = models.ApprovalStatusUnderReview
	case models.ApprovalDecisionReassign:
		// 转派到其他审批人
		if req.NextApproverID != "" {
			approval.CurrentApproverID = req.NextApproverID
			// 这里需要查询新审批人的姓名
		}
	}

	approval.UpdatedBy = userID

	// 执行更新
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, fmt.Errorf("更新审批状态失败: %v", err)
	}

	// 重新加载数据
	updatedApproval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("重新加载审批数据失败: %v", err)
	}

	return updatedApproval, nil
}

// DeleteApproval 删除审批
func (s *ApprovalService) DeleteApproval(userID string, id string) error {
	// 使用repository查找审批记录
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return fmt.Errorf("无权删除此审批记录")
	}

	// 只能删除草稿状态的审批
	if approval != nil && approval.Status != models.ApprovalStatusDraft {
		return errors.New("只能删除草稿状态的审批")
	}

	// 删除审批记录
	if err := s.approvalRepo.Delete(id); err != nil {
		return fmt.Errorf("删除审批记录失败: %v", err)
	}

	return nil
}

// CancelApproval 撤回审批
func (s *ApprovalService) CancelApproval(userID string, id string) error {
	// 使用repository查找审批记录
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return fmt.Errorf("无权撤回此审批记录")
	}

	// 只能撤回已提交但未处理的审批
	if approval != nil && (approval.Status != models.ApprovalStatusSubmitted && approval.Status != models.ApprovalStatusUnderReview) {
		return errors.New("只能撤回已提交但未处理的审批")
	}

	// 检查是否有审批记录（如果有则说明已经开始处理）
	records, err := s.approvalRepo.FindRecordsByApprovalID(id)
	if err != nil {
		return fmt.Errorf("检查审批记录失败: %v", err)
	}
	if len(records) > 0 {
		return errors.New("审批已被处理，无法撤回")
	}

	// 更新状态为已撤回
	if approval != nil {
		approval.Status = models.ApprovalStatusCancelled
		approval.UpdatedBy = userID

		if err := s.approvalRepo.Update(approval); err != nil {
			return fmt.Errorf("撤回审批失败: %v", err)
		}
	}

	return nil
}

// GetApprovalWorkflows 获取审批工作流列表
func (s *ApprovalService) GetApprovalWorkflows() ([]models.ApprovalWorkflow, error) {
	// 使用repository获取工作流列表
	workflows, err := s.approvalRepo.FindWorkflows()
	if err != nil {
		return nil, fmt.Errorf("获取审批工作流失败: %v", err)
	}

	return workflows, nil
}

// GetApprovalTemplates 获取审批模板列表
func (s *ApprovalService) GetApprovalTemplates(templateType string, category string) ([]models.ApprovalTemplate, error) {
	// 使用repository获取模板列表
	templates, err := s.approvalRepo.FindTemplates(templateType, category)
	if err != nil {
		return nil, fmt.Errorf("获取审批模板失败: %v", err)
	}

	return templates, nil
}

// generateRequestNumber 生成申请编号
func (s *ApprovalService) generateRequestNumber() (string, error) {
	// 使用简单的时间戳方式生成申请编号
	timestamp := time.Now().Format("20060102")
	random := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
	return fmt.Sprintf("AP-%s-%s", timestamp, random), nil
}

// generateUUID 生成UUID
func generateUUID() string {
	// 这里使用简单的UUID生成，实际项目中可以使用更好的UUID库
	return fmt.Sprintf("%x", time.Now().UnixNano())
}