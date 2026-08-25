package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ApprovalService 审批服务
type ApprovalService struct {
	approvalRepo *repositories.ApprovalRepository
	userRepo     repositories.UserRepository
	stateMachine *ApprovalStateMachine
	assigner     *ApprovalAssigner
}

// NewApprovalService 创建审批服务
func NewApprovalService(db *gorm.DB) *ApprovalService {
	approvalRepo := repositories.NewApprovalRepository(db)
	userRepo := repositories.NewUserRepository(db)

	return &ApprovalService{
		approvalRepo: approvalRepo,
		userRepo:     userRepo,
		stateMachine: NewApprovalStateMachine(),
		assigner:     NewApprovalAssigner(db, approvalRepo, userRepo),
	}
}

// GetApprovalStats 获取审批统计。
//
// Firm-wide pending totals are only returned when the caller explicitly has a
// business matter-management appointment. The variadic argument preserves the
// old one-argument call shape for internal callers while making the safe,
// user-scoped view the default.
func (s *ApprovalService) GetApprovalStats(userID string, includeFirmTotals ...bool) (*models.ApprovalStats, error) {
	var stats models.ApprovalStats
	showFirmTotals := len(includeFirmTotals) > 0 && includeFirmTotals[0]

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

	// A normal lawyer must not learn the firm's global approval volume. Use the
	// applicant's own pending count for that view; management roles can opt in
	// to the firm-wide count after the handler has checked their role.
	var pendingRequests int64
	if showFirmTotals {
		pendingRequests, err = s.approvalRepo.CountPending()
		if err != nil {
			return nil, fmt.Errorf("获取全部待审批数失败: %v", err)
		}
	} else {
		pendingRequests, err = s.approvalRepo.CountPendingByApplicantID(userID)
		if err != nil {
			return nil, fmt.Errorf("获取用户待审批数失败: %v", err)
		}
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
	approvals, err := s.approvalRepo.FindPendingByApproverID(userID, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取待审批列表失败: %v", err)
	}

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
	queryUserID := userID
	if req.ApplicantID != "" {
		queryUserID = req.ApplicantID
	}

	approvals, err := s.approvalRepo.FindByUserID(queryUserID, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("获取审批列表失败: %v", err)
	}

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
	approval, err := s.GetApprovalForAuthorization(userID, id)
	if err != nil || approval == nil {
		return approval, err
	}
	if err := s.LoadApprovalRecords(approval); err != nil {
		return nil, err
	}
	return approval, nil
}

// GetApprovalForAuthorization loads only the approval subject. Handlers use it
// before ethical-wall authorization so related audit rows are not read until
// the caller's object access has been established.
func (s *ApprovalService) GetApprovalForAuthorization(userID string, id string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取审批详情失败: %v", err)
	}

	// The approval detail is an object-scoped resource. Technical admin is not
	// a business approval role, so it must not bypass the applicant/approver
	// boundary by guessing an approval ID.
	if approval != nil {
		allowed := approval.ApplicantID == userID || approval.CurrentApproverID == userID
		if !allowed {
			wasApprover, historyErr := s.approvalRepo.HasApproverRecord(approval.ID, userID)
			if historyErr != nil {
				return nil, historyErr
			}
			allowed = wasApprover
		}
		if !allowed {
			return nil, fmt.Errorf("无权查看此审批记录")
		}
	}

	return approval, nil
}

// LoadApprovalRecords populates immutable decision history after object-level
// authorization. Legacy development schemas may lack the table; production
// readiness independently requires the complete approval schema.
func (s *ApprovalService) LoadApprovalRecords(approval *models.ApprovalRequest) error {
	if approval == nil || !s.approvalRepo.DB().Migrator().HasTable(&models.ApprovalRecord{}) {
		return nil
	}
	records, err := s.approvalRepo.FindRecordsByApprovalID(approval.ID)
	if err != nil {
		return err
	}
	approval.Records = records
	return nil
}

func (s *ApprovalService) GetApprovalByID(id string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取审批详情失败: %v", err)
	}
	return approval, nil
}

// GetApprovalRecords 获取审批记录
func (s *ApprovalService) GetApprovalRecords(approvalID string) ([]models.ApprovalRecord, error) {
	records, err := s.approvalRepo.FindRecordsByApprovalID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批记录失败: %v", err)
	}
	return records, nil
}

// CreateApproval 创建审批申请
func (s *ApprovalService) CreateApproval(userID string, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error) {
	// 生成申请编号
	requestNumber, err := s.generateRequestNumber()
	if err != nil {
		return nil, fmt.Errorf("生成申请编号失败: %v", err)
	}

	// 如果userName为空，从数据库查询用户信息
	if userName == "" {
		user, err := s.userRepo.FindByStringID(userID)
		if err == nil && user != nil {
			userName = user.Name
		} else {
			userName = "未知用户"
		}
	}

	// 设置默认工作流类型
	workflowType := req.WorkflowType
	if workflowType == "" {
		workflowType = "STANDARD_APPROVAL"
	}

	// 创建审批申请 - 初始状态为草稿
	approval := &models.ApprovalRequest{
		ID:                    generateUUID(),
		RequestNumber:         requestNumber,
		Title:                 req.Title,
		Type:                  req.Type,
		Category:              req.Category,
		Content:               req.Content,
		ApplicantID:           userID,
		ApplicantName:         userName,
		ApplicantTitle:        "",
		DepartmentID:          "",
		DepartmentName:        "",
		Urgency:               req.Urgency,
		Priority:              req.Priority,
		ExpectedEffectiveDate: nil,
		ExpectedExpiryDate:    nil,
		DurationDays:          req.DurationDays,
		Status:                models.ApprovalStatusDraft,
		SubmissionDate:        nil,
		CurrentStage:          "",
		CurrentApproverID:     "",
		CurrentApproverName:   "",
		WorkflowType:          workflowType,
		WorkflowConfig:        "{}",
		Attachments:           "{}",
		Metadata:              "{}",
		ConflictResult:        "{}",
		CreatedBy:             userID,
		UpdatedBy:             userID,
	}

	// 设置默认值
	if approval.Urgency == "" {
		approval.Urgency = models.ApprovalUrgencyNormal
	}
	if approval.Priority == "" {
		approval.Priority = models.ApprovalPriorityMedium
	}

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

	// 保存草稿
	if err := s.approvalRepo.Create(approval); err != nil {
		return nil, fmt.Errorf("保存审批申请失败: %v", err)
	}

	return approval, nil
}

// SubmitApproval 提交审批申请
func (s *ApprovalService) SubmitApproval(userID string, approvalID string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权提交此审批记录")
	}

	// 只能提交草稿或被拒绝/需要修改状态的审批
	if approval != nil && !s.stateMachine.CanEdit(approval.Status) {
		return nil, fmt.Errorf("当前状态 %s 不允许提交", approval.Status)
	}

	// 验证状态转换
	if err := s.stateMachine.ValidateTransition(approval.Status, models.ApprovalStatusSubmitted); err != nil {
		return nil, err
	}

	// 分配审批人
	if err := s.assigner.AssignApprover(approval); err != nil {
		return nil, fmt.Errorf("分配审批人失败: %v", err)
	}

	// 设置提交时间
	now := time.Now()
	approval.SubmissionDate = &now
	approval.UpdatedBy = userID

	// 保存更新
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, fmt.Errorf("提交审批失败: %v", err)
	}

	// 重新加载数据
	return s.approvalRepo.FindByID(approvalID)
}

// ProcessApprovalDecision 处理审批决定（通过/拒绝/要求修改/转派）
func (s *ApprovalService) ProcessApprovalDecision(userID string, approvalID string, req *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是当前审批人
	if approval != nil && approval.CurrentApproverID != userID {
		return nil, errors.New("无权审批此申请")
	}
	if approval != nil && approval.ApplicantID == userID {
		return nil, errors.New("申请人不能审批自己的申请")
	}

	// 检查审批状态
	if approval != nil && approval.Status != models.ApprovalStatusSubmitted &&
		approval.Status != models.ApprovalStatusUnderReview &&
		approval.Status != models.ApprovalStatusResubmitted {
		return nil, errors.New("审批状态不允许处理")
	}

	// 获取下一个状态
	nextStatus, err := s.stateMachine.GetNextState(approval.Status, req.Decision)
	if err != nil {
		return nil, err
	}

	// 更新状态为审核中
	if approval.Status != models.ApprovalStatusUnderReview {
		approval.Status = models.ApprovalStatusUnderReview
	}

	approverName := approval.CurrentApproverName
	approverRole := ""
	if approver, lookupErr := s.userRepo.FindByStringID(userID); lookupErr == nil && approver != nil {
		if approver.Name != "" {
			approverName = approver.Name
		}
		approverRole = approver.Role
	}

	// 创建审批记录
	record := &models.ApprovalRecord{
		ID:                  generateUUID(),
		ApprovalRequestID:   approval.ID,
		Stage:               approval.CurrentStage,
		StageOrder:          1,
		ApproverID:          userID,
		ApproverName:        approverName,
		ApproverTitle:       "",
		ApproverRole:        approverRole,
		Decision:            req.Decision,
		DecisionReason:      req.DecisionReason,
		DecisionComments:    req.DecisionComments,
		ApprovedConditions:  "{}",
		ImposedRequirements: "{}",
		FollowUpActions:     "[]",
		ApprovalDate:        time.Now(),
		EffectiveDate:       nil,
		NextReviewDate:      nil,
		SupportingDocuments: "[]",
		EvidenceReferences:  "[]",
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

	// 根据决定处理状态
	switch req.Decision {
	case models.ApprovalDecisionApprove:
		// 检查是否有多级审批
		isComplete, err := s.assigner.MoveToNextStage(approval)
		if err != nil {
			return nil, fmt.Errorf("移动到下一阶段失败: %v", err)
		}
		if !isComplete {
			// 还有下一级审批，发送通知
			_ = s.sendApprovalNotification(approval, "下一阶段审批")
		}

	case models.ApprovalDecisionReject:
		approval.Status = nextStatus
		approval.CurrentApproverID = ""
		approval.CurrentApproverName = ""
		// 发送拒绝通知
		_ = s.sendApprovalNotification(approval, "审批被拒绝")

	case models.ApprovalDecisionRequestChanges:
		approval.Status = nextStatus
		// 发送要求修改通知
		_ = s.sendApprovalNotification(approval, "需要修改")

	case models.ApprovalDecisionReassign:
		// 转派到其他审批人
		if req.NextApproverID != "" {
			// 验证新审批人存在
			newApprover, err := s.userRepo.FindByStringID(req.NextApproverID)
			if err != nil || newApprover == nil {
				return nil, errors.New("指定的审批人不存在")
			}
			approval.CurrentApproverID = req.NextApproverID
			approval.CurrentApproverName = newApprover.Name
		}
	}

	approval.UpdatedBy = userID

	// 执行更新
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, fmt.Errorf("更新审批状态失败: %v", err)
	}

	// 重新加载数据
	updatedApproval, err := s.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("重新加载审批数据失败: %v", err)
	}

	return updatedApproval, nil
}

// ResubmitApproval 重新提交被拒绝或要求修改的审批
func (s *ApprovalService) ResubmitApproval(userID string, approvalID string, revisionNote string) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权重新提交此审批记录")
	}

	// 检查是否可以重新提交
	if !s.stateMachine.CanResubmit(approval.Status) {
		return nil, fmt.Errorf("当前状态 %s 不允许重新提交", approval.Status)
	}

	// 验证状态转换
	if err := s.stateMachine.ValidateTransition(approval.Status, models.ApprovalStatusResubmitted); err != nil {
		return nil, err
	}

	// 更新元数据，添加修改说明
	var metadata map[string]interface{}
	if approval.Metadata != "{}" && approval.Metadata != "" {
		json.Unmarshal([]byte(approval.Metadata), &metadata)
	} else {
		metadata = make(map[string]interface{})
	}
	metadata["revision_note"] = revisionNote
	metadata["resubmitted_at"] = time.Now().Format(time.RFC3339)
	metadataBytes, _ := json.Marshal(metadata)
	approval.Metadata = string(metadataBytes)

	// 重新分配审批人
	if err := s.assigner.AssignApprover(approval); err != nil {
		return nil, fmt.Errorf("重新分配审批人失败: %v", err)
	}

	approval.Status = models.ApprovalStatusResubmitted
	approval.UpdatedBy = userID

	// 执行更新
	if err := s.approvalRepo.Update(approval); err != nil {
		return nil, fmt.Errorf("重新提交失败: %v", err)
	}

	// 发送重新提交通知
	_ = s.sendApprovalNotification(approval, "重新提交")

	return s.approvalRepo.FindByID(approvalID)
}

// CancelApproval 撤回审批
func (s *ApprovalService) CancelApproval(userID string, approvalID string) error {
	approval, err := s.approvalRepo.FindByID(approvalID)
	if err != nil {
		return fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return fmt.Errorf("无权撤回此审批记录")
	}

	// 检查是否有审批记录（如果有则说明已经开始处理）
	records, err := s.approvalRepo.FindRecordsByApprovalID(approvalID)
	if err != nil {
		return fmt.Errorf("检查审批记录失败: %v", err)
	}

	// 检查是否可以撤回
	if !s.stateMachine.CanCancel(approval.Status, len(records) > 0) {
		return errors.New("当前状态不允许撤回或已有审批记录")
	}

	// 验证状态转换
	if err := s.stateMachine.ValidateTransition(approval.Status, models.ApprovalStatusCancelled); err != nil {
		return err
	}

	// 更新状态为已撤回
	approval.Status = models.ApprovalStatusCancelled
	approval.CurrentApproverID = ""
	approval.CurrentApproverName = ""
	approval.UpdatedBy = userID

	if err := s.approvalRepo.Update(approval); err != nil {
		return fmt.Errorf("撤回审批失败: %v", err)
	}

	return nil
}

// UpdateApproval 更新审批申请（草稿或需要修改状态）
func (s *ApprovalService) UpdateApproval(userID string, id string, req *models.UpdateApprovalRequest) (*models.ApprovalRequest, error) {
	approval, err := s.approvalRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("查找审批记录失败: %v", err)
	}

	// 确保是申请人本人
	if approval != nil && approval.ApplicantID != userID {
		return nil, fmt.Errorf("无权更新此审批记录")
	}

	// 只能更新可编辑状态的审批
	if !s.stateMachine.CanEdit(approval.Status) {
		return nil, fmt.Errorf("当前状态 %s 不允许编辑", approval.Status)
	}

	// 更新字段
	if req.Title != "" {
		approval.Title = req.Title
	}
	if req.Content != "" {
		approval.Content = req.Content
	}

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
	return s.approvalRepo.FindByID(id)
}

// DeleteApproval 删除审批（仅草稿）
func (s *ApprovalService) DeleteApproval(userID string, id string) error {
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

	if err := s.approvalRepo.Delete(id); err != nil {
		return fmt.Errorf("删除审批记录失败: %v", err)
	}

	return nil
}

// GetApprovalWorkflows 获取审批工作流列表
func (s *ApprovalService) GetApprovalWorkflows() ([]models.ApprovalWorkflow, error) {
	workflows, err := s.approvalRepo.FindWorkflows()
	if err != nil {
		return nil, fmt.Errorf("获取审批工作流失败: %v", err)
	}
	return workflows, nil
}

// GetApprovalTemplates 获取审批模板列表
func (s *ApprovalService) GetApprovalTemplates(templateType string, category string) ([]models.ApprovalTemplateV2, error) {
	templates, err := s.approvalRepo.FindTemplates(templateType, category)
	if err != nil {
		return nil, fmt.Errorf("获取审批模板失败: %v", err)
	}
	return templates, nil
}

// sendApprovalNotification 发送审批通知（简化版）
func (s *ApprovalService) sendApprovalNotification(approval *models.ApprovalRequest, message string) error {
	// TODO: 实现实际的通知发送逻辑
	// 可以通过WebSocket、邮件等方式发送通知
	fmt.Printf("通知发送: 申请 %s - %s\n", approval.RequestNumber, message)
	return nil
}

// generateRequestNumber 生成申请编号
func (s *ApprovalService) generateRequestNumber() (string, error) {
	timestamp := time.Now().Format("20060102")
	random := strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
	return fmt.Sprintf("AP-%s-%s", timestamp, random), nil
}

// generateUUID 生成UUID
func generateUUID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
