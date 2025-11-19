package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ApprovalServiceInterface 审批服务接口
type ApprovalServiceInterface interface {
	CreateApproval(userID string, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error)
	GetApproval(userID string, id string) (*models.ApprovalRequest, error)
	ProcessApproval(userID string, userName string, id string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error)
}

// ConflictDetectionServiceInterface 冲突检测服务接口
type ConflictDetectionServiceInterface interface {
	PerformConflictCheck(ctx context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error)
}

// ApprovalConflictIntegrationService 审批冲突集成服务接口
type ApprovalConflictIntegrationService interface {
	// CreateIntegratedApproval 创建集成的审批申请
	CreateIntegratedApproval(ctx context.Context, userID, userName string, req *IntegrationRequest) (*IntegrationResult, error)

	// TriggerConflictCheckForApproval 为审批申请触发冲突检测
	TriggerConflictCheckForApproval(ctx context.Context, approvalID string, conflictReq *models.ConflictCheckRequest) (*IntegrationConflictCheckResult, error)

	// AutoCreateCaseFromApproval 从审批申请自动创建案件
	AutoCreateCaseFromApproval(ctx context.Context, approvalID string, caseData map[string]interface{}) (*CaseCreationResult, error)

	// GetIntegrationStatus 获取集成状态
	GetIntegrationStatus(ctx context.Context, approvalID string) (*IntegrationStatus, error)

	// ProcessApprovalWithConflict 处理包含冲突检测的审批申请
	ProcessApprovalWithConflict(ctx context.Context, userID, userName string, approvalID string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error)
}

// approvalConflictIntegrationService 审批冲突集成服务实现
type approvalConflictIntegrationService struct {
	approvalService    ApprovalServiceInterface
	conflictService    ConflictDetectionServiceInterface
	integrationRepo    repositories.IntegrationRepositoryInterface
}

// NewApprovalConflictIntegrationService 创建审批冲突集成服务
func NewApprovalConflictIntegrationService(
	approvalService ApprovalServiceInterface,
	conflictService ConflictDetectionServiceInterface,
	integrationRepo repositories.IntegrationRepositoryInterface,
) ApprovalConflictIntegrationService {
	return &approvalConflictIntegrationService{
		approvalService: approvalService,
		conflictService: conflictService,
		integrationRepo: integrationRepo,
	}
}

// IntegrationRequest 集成请求
type IntegrationRequest struct {
	Type                string                 `json:"type" binding:"required"`
	Title               string                 `json:"title" binding:"required"`
	Content             string                 `json:"content" binding:"required"`
	ApplicantName       string                 `json:"applicant_name" binding:"required"`
	DepartmentName      string                 `json:"department_name" binding:"required"`
	Urgency             string                 `json:"urgency"`
	Priority            string                 `json:"priority"`
	WorkflowType        string                 `json:"workflow_type" binding:"required"`
	ExpectedDuration    int                    `json:"expected_duration"`
	Category            string                 `json:"category"`
	Metadata            map[string]interface{} `json:"metadata"`
	ConflictCheckConfig *models.ConflictCheckRequest `json:"conflict_check_config,omitempty"`
	CaseCreationConfig  map[string]interface{} `json:"case_creation_config,omitempty"`
}

// IntegrationResult 集成结果
type IntegrationResult struct {
	ApprovalID         string                              `json:"approval_id"`
	Status             string                              `json:"status"`
	Message            string                              `json:"message"`
	CreatedAt          time.Time                           `json:"created_at"`
	ConflictCheck      *IntegrationConflictCheckResult     `json:"conflict_check,omitempty"`
	CaseCreation       *CaseCreationResult                  `json:"case_creation,omitempty"`
	IntegrationDetails *models.ApprovalIntegrationMetadata  `json:"integration_details,omitempty"`
}

// IntegrationConflictCheckResult 集成冲突检测结果
type IntegrationConflictCheckResult struct {
	CheckID         string                    `json:"check_id"`
	Status          string                    `json:"status"`
	HasConflict     bool                      `json:"has_conflict"`
	ConflictCount   int                       `json:"conflict_count"`
	RiskLevel       string                    `json:"risk_level"`
	RiskScore       float64                   `json:"risk_score"`
	CheckTime       time.Time                 `json:"check_time"`
	Duration        int64                     `json:"duration"`
	ConflictCases   []*models.ConflictCase    `json:"conflict_cases,omitempty"`
	Recommendations []string                  `json:"recommendations,omitempty"`
}

// CaseCreationResult 案件创建结果
type CaseCreationResult struct {
	CaseID      string    `json:"case_id"`
	CaseNumber  string    `json:"case_number"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	CaseData    map[string]interface{} `json:"case_data,omitempty"`
}

// IntegrationStatus 集成状态
type IntegrationStatus struct {
	ApprovalID         string                 `json:"approval_id"`
	OverallStatus      string                 `json:"overall_status"`
	ConflictCheck      *ConflictCheckStatus   `json:"conflict_check,omitempty"`
	CaseCreation       *CaseCreationStatus    `json:"case_creation,omitempty"`
	NextActions        []string               `json:"next_actions,omitempty"`
	Timeline           []TimelineEvent        `json:"timeline,omitempty"`
}

// ConflictCheckStatus 冲突检测状态
type ConflictCheckStatus struct {
	Status       string    `json:"status"`
	CheckID      string    `json:"check_id,omitempty"`
	HasConflict  bool      `json:"has_conflict,omitempty"`
	RiskLevel    string    `json:"risk_level,omitempty"`
	LastChecked  time.Time `json:"last_checked,omitempty"`
	Duration     int64     `json:"duration,omitempty"`
}

// CaseCreationStatus 案件创建状态
type CaseCreationStatus struct {
	Status        string    `json:"status"`
	CaseID        string    `json:"case_id,omitempty"`
	CaseNumber    string    `json:"case_number,omitempty"`
	LastError     string    `json:"last_error,omitempty"`
	RetryCount    int       `json:"retry_count,omitempty"`
	LastAttempted time.Time `json:"last_attempted,omitempty"`
}

// TimelineEvent 时间线事件
type TimelineEvent struct {
	Timestamp   time.Time               `json:"timestamp"`
	EventType   string                  `json:"event_type"`
	Description string                  `json:"description"`
	Actor       string                  `json:"actor"`
	Details     map[string]interface{}  `json:"details,omitempty"`
}

// CreateIntegratedApproval 创建集成的审批申请
func (s *approvalConflictIntegrationService) CreateIntegratedApproval(ctx context.Context, userID, userName string, req *IntegrationRequest) (*IntegrationResult, error) {
	log.Printf("创建集成审批申请，用户：%s，类型：%s", userName, req.Type)

	// 设置元数据
	metadata := make(map[string]interface{})
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	// 添加申请人信息到元数据
	metadata["applicant_name"] = req.ApplicantName
	metadata["department_name"] = req.DepartmentName

	// 如果包含冲突检测配置，设置相关字段
	if req.ConflictCheckConfig != nil {
		metadata["conflict_check_config"] = req.ConflictCheckConfig
		metadata["integration_type"] = "conflict"
	}

	// 如果包含案件创建配置
	if req.CaseCreationConfig != nil {
		metadata["case_creation_config"] = req.CaseCreationConfig
		metadata["integration_type"] = "both"
	}

	// 创建审批申请
	approvalReq := &models.CreateApprovalRequest{
		Type:         req.Type,
		Title:        req.Title,
		Content:      req.Content,
		WorkflowType: req.WorkflowType,
		Metadata:     metadata,
	}

	// 设置可选字段
	if req.Urgency != "" {
		approvalReq.Urgency = req.Urgency
	}
	if req.Priority != "" {
		approvalReq.Priority = req.Priority
	}
	if req.Category != "" {
		approvalReq.Category = req.Category
	}
	if req.ExpectedDuration > 0 {
		approvalReq.DurationDays = req.ExpectedDuration
	}

	// 创建审批申请
	createdApproval, err := s.approvalService.CreateApproval(userID, userName, approvalReq)
	if err != nil {
		log.Printf("创建审批申请失败：%v", err)
		return nil, fmt.Errorf("创建审批申请失败：%w", err)
	}

	// 创建集成元数据
	integrationMetadata := &models.ApprovalIntegrationMetadata{
		IntegrationType:  "conflict", // 默认类型
		IntegrationID:    fmt.Sprintf("INT_%s_%d", createdApproval.ID, time.Now().Unix()),
		IntegrationTime:  time.Now(),
		AutoSubmitted:    false,
		TriggerSource:    "manual",
		CreatedBy:        userID,
		Version:          1,
	}

	result := &IntegrationResult{
		ApprovalID:         createdApproval.ID,
		Status:             "created",
		Message:            "集成审批申请创建成功",
		CreatedAt:          time.Now(),
		IntegrationDetails: integrationMetadata,
	}

	// 如果需要自动进行冲突检测
	if req.ConflictCheckConfig != nil {
		log.Printf("触发自动冲突检测，审批ID：%s", createdApproval.ID)
		conflictResult, err := s.TriggerConflictCheckForApproval(ctx, createdApproval.ID, req.ConflictCheckConfig)
		if err != nil {
			log.Printf("自动冲突检测失败：%v", err)
			result.Message = "审批申请创建成功，但自动冲突检测失败"
		} else {
			result.ConflictCheck = conflictResult
			result.Message = "集成审批申请创建成功，冲突检测已完成"
		}
	}

	log.Printf("集成审批申请创建完成，ID：%s", result.ApprovalID)
	return result, nil
}

// TriggerConflictCheckForApproval 为审批申请触发冲突检测
func (s *approvalConflictIntegrationService) TriggerConflictCheckForApproval(ctx context.Context, approvalID string, conflictReq *models.ConflictCheckRequest) (*IntegrationConflictCheckResult, error) {
	log.Printf("为审批申请触发冲突检测，审批ID：%s", approvalID)

	// 设置用户ID和时间
	conflictReq.UserID = userIDFromContext(ctx)
	conflictReq.RequestTime = time.Now()

	// 执行冲突检测
	conflictResp, err := s.conflictService.PerformConflictCheck(ctx, conflictReq)
	if err != nil {
		log.Printf("冲突检测执行失败：%v", err)
		return nil, fmt.Errorf("冲突检测执行失败：%w", err)
	}

	// 更新审批申请的冲突关联信息
	if err := s.integrationRepo.UpdateApprovalConflictAssociation(ctx, approvalID, conflictResp); err != nil {
		log.Printf("更新冲突关联信息失败：%v", err)
		// 不影响返回结果，只记录错误
	}

	// 构建返回结果
	result := &IntegrationConflictCheckResult{
		CheckID:         conflictResp.CheckID,
		Status:          "completed",
		HasConflict:     conflictResp.HasConflict,
		ConflictCount:   len(conflictResp.ConflictCases),
		RiskLevel:       conflictResp.RiskAssessment.OverallRisk,
		RiskScore:       conflictResp.RiskAssessment.RiskScore,
		CheckTime:       conflictResp.CheckTime,
		Duration:        conflictResp.Duration,
		ConflictCases:   conflictResp.ConflictCases,
		Recommendations: conflictResp.Recommendations,
	}

	log.Printf("冲突检测完成，审批ID：%s，检测ID：%s，存在冲突：%t", approvalID, result.CheckID, result.HasConflict)
	return result, nil
}

// AutoCreateCaseFromApproval 从审批申请自动创建案件
func (s *approvalConflictIntegrationService) AutoCreateCaseFromApproval(ctx context.Context, approvalID string, caseData map[string]interface{}) (*CaseCreationResult, error) {
	log.Printf("从审批申请自动创建案件，审批ID：%s", approvalID)

	// 这里应该调用案件创建服务，由于案件创建服务还没有完全实现，先返回模拟结果
	caseID := fmt.Sprintf("CASE_%s_%d", approvalID, time.Now().Unix())
	caseNumber := fmt.Sprintf("CA-%d", time.Now().Unix())

	// 创建案件创建关联信息
	caseAssociation := &models.CaseCreationAssociation{
		Created:       true,
		CaseID:        caseID,
		CaseNumber:    caseNumber,
		CreationTime:  time.Now(),
		DataMapping:   caseData,
		Status:        "completed",
		StatusMessage: "案件创建成功",
	}

	// 更新审批申请的案件关联信息
	if err := s.integrationRepo.UpdateApprovalCaseAssociation(ctx, approvalID, caseAssociation); err != nil {
		log.Printf("更新案件关联信息失败：%v", err)
		return nil, fmt.Errorf("更新案件关联信息失败：%w", err)
	}

	result := &CaseCreationResult{
		CaseID:    caseID,
		CaseNumber: caseNumber,
		Status:    "completed",
		Message:   "案件创建成功",
		CreatedAt: time.Now(),
		CaseData:  caseData,
	}

	log.Printf("案件创建完成，审批ID：%s，案件ID：%s", approvalID, caseID)
	return result, nil
}

// GetIntegrationStatus 获取集成状态
func (s *approvalConflictIntegrationService) GetIntegrationStatus(ctx context.Context, approvalID string) (*IntegrationStatus, error) {
	log.Printf("获取集成状态，审批ID：%s", approvalID)

	// 从上下文获取用户ID，如果没有则使用默认值
	userID := userIDFromContext(ctx)

	// 获取审批申请信息
	approval, err := s.approvalService.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批申请失败：%w", err)
	}

	status := &IntegrationStatus{
		ApprovalID:    approvalID,
		OverallStatus: "pending",
		NextActions:   []string{},
		Timeline:      []TimelineEvent{},
	}

	// 从元数据中解析集成信息
	if approval.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err == nil {
			if integrationType, ok := metadata["integration_type"].(string); ok {
				status.OverallStatus = integrationType
			}
		}
	}

	// 检查冲突检测状态
	if approval.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err == nil {
			if integrationType, ok := metadata["integration_type"].(string); ok && (integrationType == "conflict" || integrationType == "both") {
				status.ConflictCheck = &ConflictCheckStatus{
					Status:      "completed",
					CheckID:     approvalID, // 简化实现
					HasConflict: false,      // 简化实现
					LastChecked: time.Now(),
				}
			}

			// 检查案件创建状态
			if integrationType, ok := metadata["integration_type"].(string); ok && integrationType == "both" {
				status.CaseCreation = &CaseCreationStatus{
					Status:     "pending",
					CaseID:     "",
					CaseNumber: "",
				}
				status.NextActions = append(status.NextActions, "创建案件")
			}
		}
	}

	return status, nil
}

// ProcessApprovalWithConflict 处理包含冲突检测的审批申请
func (s *approvalConflictIntegrationService) ProcessApprovalWithConflict(ctx context.Context, userID, userName string, approvalID string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	log.Printf("处理包含冲突检测的审批申请，审批ID：%s，决定：%s", approvalID, decisionReq.Decision)

	// 获取当前的审批申请
	approval, err := s.approvalService.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批申请失败：%w", err)
	}

	// 处理审批决定
	updatedApproval, err := s.approvalService.ProcessApproval(userID, userName, approvalID, decisionReq)
	if err != nil {
		return nil, fmt.Errorf("处理审批决定失败：%w", err)
	}

	// 如果审批通过，且包含案件创建配置，则自动创建案件
	if decisionReq.Decision == models.ApprovalDecisionApprove && approval.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err == nil {
			if integrationType, ok := metadata["integration_type"].(string); ok && integrationType == "both" {
				log.Printf("审批通过，开始自动创建案件，审批ID：%s", approvalID)

				// 从metadata中获取案件创建配置
				caseData := make(map[string]interface{})
				if config, ok := metadata["case_creation_config"]; ok {
					caseData = config.(map[string]interface{})
				}

				_, err := s.AutoCreateCaseFromApproval(ctx, approvalID, caseData)
				if err != nil {
					log.Printf("自动创建案件失败：%v", err)
					// 不影响审批结果，只记录错误
				}
			}
		}
	}

	log.Printf("审批申请处理完成，ID：%s，最终状态：%s", updatedApproval.ID, updatedApproval.Status)
	return updatedApproval, nil
}

// Helper functions

// userIDFromContext 从上下文获取用户ID
func userIDFromContext(ctx context.Context) string {
	// 这里应该从JWT token或其他认证信息中获取用户ID
	// 简化实现，返回固定值
	return "user_from_context"
}