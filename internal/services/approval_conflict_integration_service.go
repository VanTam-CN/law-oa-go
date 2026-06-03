package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ApprovalServiceInterface 审批服务接口
type ApprovalServiceInterface interface {
	CreateApproval(userID string, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error)
	GetApproval(userID string, id string) (*models.ApprovalRequest, error)
	GetApprovalByID(id string) (*models.ApprovalRequest, error)
	SubmitApproval(userID string, approvalID string) (*models.ApprovalRequest, error)
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
	approvalService ApprovalServiceInterface
	conflictService ConflictDetectionServiceInterface
	caseService     *CaseService
	integrationRepo repositories.IntegrationRepositoryInterface
}

// NewApprovalConflictIntegrationService 创建审批冲突集成服务
func NewApprovalConflictIntegrationService(
	approvalService ApprovalServiceInterface,
	conflictService ConflictDetectionServiceInterface,
	caseService *CaseService,
	integrationRepo repositories.IntegrationRepositoryInterface,
) ApprovalConflictIntegrationService {
	return &approvalConflictIntegrationService{
		approvalService: approvalService,
		conflictService: conflictService,
		caseService:     caseService,
		integrationRepo: integrationRepo,
	}
}

// IntegrationRequest 集成请求
type IntegrationRequest struct {
	Type                string                       `json:"type" binding:"required"`
	Title               string                       `json:"title" binding:"required"`
	Content             string                       `json:"content" binding:"required"`
	ApplicantName       string                       `json:"applicant_name" binding:"required"`
	DepartmentName      string                       `json:"department_name" binding:"required"`
	Urgency             string                       `json:"urgency"`
	Priority            string                       `json:"priority"`
	WorkflowType        string                       `json:"workflow_type" binding:"required"`
	ExpectedDuration    int                          `json:"expected_duration"`
	Category            string                       `json:"category"`
	Metadata            map[string]interface{}       `json:"metadata"`
	ConflictCheckConfig *models.ConflictCheckRequest `json:"conflict_check_config,omitempty"`
	CaseCreationConfig  map[string]interface{}       `json:"case_creation_config,omitempty"`
}

// IntegrationResult 集成结果
type IntegrationResult struct {
	ApprovalID         string                              `json:"approval_id"`
	Status             string                              `json:"status"`
	Message            string                              `json:"message"`
	CreatedAt          time.Time                           `json:"created_at"`
	ConflictCheck      *IntegrationConflictCheckResult     `json:"conflict_check,omitempty"`
	CaseCreation       *CaseCreationResult                 `json:"case_creation,omitempty"`
	IntegrationDetails *models.ApprovalIntegrationMetadata `json:"integration_details,omitempty"`
	ApprovalSnapshot   map[string]interface{}              `json:"approval_snapshot,omitempty"`
}

// IntegrationConflictCheckResult 集成冲突检测结果
type IntegrationConflictCheckResult struct {
	CheckID         string                 `json:"check_id"`
	Status          string                 `json:"status"`
	HasConflict     bool                   `json:"has_conflict"`
	ConflictCount   int                    `json:"conflict_count"`
	RiskLevel       string                 `json:"risk_level"`
	RiskScore       float64                `json:"risk_score"`
	CheckTime       time.Time              `json:"check_time"`
	Duration        int64                  `json:"duration"`
	ConflictCases   []*models.ConflictCase `json:"conflict_cases,omitempty"`
	Recommendations []string               `json:"recommendations,omitempty"`
}

// CaseCreationResult 案件创建结果
type CaseCreationResult struct {
	CaseID     string                 `json:"case_id"`
	CaseNumber string                 `json:"case_number"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	CreatedAt  time.Time              `json:"created_at"`
	CaseData   map[string]interface{} `json:"case_data,omitempty"`
}

// IntegrationStatus 集成状态
type IntegrationStatus struct {
	ApprovalID    string               `json:"approval_id"`
	OverallStatus string               `json:"overall_status"`
	ConflictCheck *ConflictCheckStatus `json:"conflict_check,omitempty"`
	CaseCreation  *CaseCreationStatus  `json:"case_creation,omitempty"`
	NextActions   []string             `json:"next_actions,omitempty"`
	Timeline      []TimelineEvent      `json:"timeline,omitempty"`
}

// ConflictCheckStatus 冲突检测状态
type ConflictCheckStatus struct {
	Status      string    `json:"status"`
	CheckID     string    `json:"check_id,omitempty"`
	HasConflict bool      `json:"has_conflict,omitempty"`
	RiskLevel   string    `json:"risk_level,omitempty"`
	LastChecked time.Time `json:"last_checked,omitempty"`
	Duration    int64     `json:"duration,omitempty"`
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
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Actor       string                 `json:"actor"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// CreateIntegratedApproval 创建集成的审批申请
func (s *approvalConflictIntegrationService) CreateIntegratedApproval(ctx context.Context, userID, userName string, req *IntegrationRequest) (*IntegrationResult, error) {
	log.Printf("创建集成审批申请，用户：%s，类型：%s", userName, req.Type)

	// 设置元数据
	metadata := make(map[string]interface{})
	if req.Metadata != nil {
		metadata = copyMetadata(req.Metadata)
	}

	// 添加申请人信息到元数据
	metadata["applicant_name"] = req.ApplicantName
	metadata["department_name"] = req.DepartmentName
	approvalSnapshot := buildApprovalSnapshot(userID, userName, req, metadata)
	metadata["approval_snapshot"] = approvalSnapshot
	metadata["approval_snapshot_version"] = 1
	metadata["approval_snapshot_created_at"] = approvalSnapshot["snapshot_at"]

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
	if submittedApproval, err := s.approvalService.SubmitApproval(userID, createdApproval.ID); err == nil && submittedApproval != nil {
		createdApproval = submittedApproval
	} else if err != nil {
		log.Printf("提交集成审批申请失败：%v", err)
		return nil, fmt.Errorf("提交集成审批申请失败：%w", err)
	}

	// 创建集成元数据
	integrationMetadata := &models.ApprovalIntegrationMetadata{
		IntegrationType: "conflict", // 默认类型
		IntegrationID:   fmt.Sprintf("INT_%s_%d", createdApproval.ID, time.Now().Unix()),
		IntegrationTime: time.Now(),
		AutoSubmitted:   false,
		TriggerSource:   "manual",
		CreatedBy:       userID,
		Version:         1,
	}

	result := &IntegrationResult{
		ApprovalID:         createdApproval.ID,
		Status:             createdApproval.Status,
		Message:            "集成审批申请创建并提交成功",
		CreatedAt:          time.Now(),
		IntegrationDetails: integrationMetadata,
		ApprovalSnapshot:   approvalSnapshot,
	}
	if req.CaseCreationConfig != nil {
		tracking := &repositories.CaseCreationTracking{
			ID:                  fmt.Sprintf("CASE_TRACK_%s_%d", createdApproval.ID, time.Now().UnixNano()),
			ApprovalRequestID:   createdApproval.ID,
			CreationStatus:      "pending",
			ProgressPercentage:  0,
			ErrorDetails:        "{}",
			DataMapping:         jsonString(req.CaseCreationConfig),
			MappedFields:        "{}",
			UnmappedFields:      "{}",
			AppliedConditions:   "{}",
			ImposedRequirements: "{}",
			WorkflowActions:     "[]",
			CreatedBy:           userID,
			CreatedAt:           time.Now(),
		}
		if err := s.integrationRepo.CreateCaseCreationTracking(ctx, tracking); err != nil {
			log.Printf("创建案件跟踪记录失败：%v", err)
		}
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

	// 将map转换为CreateCaseRequest
	jsonData, err := json.Marshal(caseData)
	if err != nil {
		return nil, fmt.Errorf("序列化案件数据失败: %w", err)
	}

	var createCaseReq CreateCaseRequest
	if err := json.Unmarshal(jsonData, &createCaseReq); err != nil {
		return nil, fmt.Errorf("解析案件数据失败: %w", err)
	}

	// 调用CaseService创建案件
	createdCase, err := s.caseService.CreateCase(ctx, &createCaseReq)
	if err != nil {
		log.Printf("创建案件失败: %v", err)
		return nil, fmt.Errorf("创建案件失败: %w", err)
	}

	caseID := fmt.Sprintf("%d", createdCase.ID)
	caseNumber := createdCase.CaseNumber
	if caseNumber == "" {
		caseNumber = fmt.Sprintf("CASE-%d", createdCase.ID)
	}

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
		// 不影响案件创建结果，只记录错误
	}

	result := &CaseCreationResult{
		CaseID:     caseID,
		CaseNumber: caseNumber,
		Status:     "completed",
		Message:    "案件创建成功",
		CreatedAt:  time.Now(),
		CaseData:   caseData,
	}

	log.Printf("案件创建完成，审批ID：%s，案件ID：%s", approvalID, caseID)
	return result, nil
}

// GetIntegrationStatus 获取集成状态
func (s *approvalConflictIntegrationService) GetIntegrationStatus(ctx context.Context, approvalID string) (*IntegrationStatus, error) {
	log.Printf("获取集成状态，审批ID：%s", approvalID)

	// 获取审批申请信息
	approval, err := s.approvalService.GetApprovalByID(approvalID)
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

	if approval.Status == models.ApprovalStatusApproved && !approval.CaseCreated {
		if _, err := s.ensureApprovedApprovalHasCase(ctx, approval, userIDFromContext(ctx), userIDFromContext(ctx)); err != nil {
			log.Printf("修复已通过审批的成案关联失败，审批ID=%s，错误=%v", approvalID, err)
		} else if refreshed, err := s.approvalService.GetApprovalByID(approvalID); err == nil && refreshed != nil {
			approval = refreshed
		}
	}

	// 检查冲突检测状态
	if approval.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(approval.Metadata), &metadata); err == nil {
			if integrationType, ok := metadata["integration_type"].(string); ok && (integrationType == "conflict" || integrationType == "both") {
				checkID := approvalID
				hasConflict := false
				riskLevel := ""

				// 从元数据中读取实际的冲突检测结果
				if cc, ok := metadata["conflict_check"].(map[string]interface{}); ok {
					if hc, ok := cc["has_conflict"].(bool); ok {
						hasConflict = hc
					}
					if rl, ok := cc["risk_level"].(string); ok {
						riskLevel = rl
					}
					if cid, ok := cc["conflict_check_id"].(string); ok {
						checkID = cid
					}
				}

				status.ConflictCheck = &ConflictCheckStatus{
					Status:      "completed",
					CheckID:     checkID,
					HasConflict: hasConflict,
					RiskLevel:   riskLevel,
					LastChecked: time.Now(),
				}
			}

			// 检查案件创建状态
			if shouldExposeCaseCreation(metadata, approval) {
				status.CaseCreation = s.buildCaseCreationStatus(ctx, approvalID, approval)
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

	// 审批通过前检查冲突风险等级
	if decisionReq.Decision == models.ApprovalDecisionApprove {
		association, err := s.integrationRepo.GetConflictAssociationByApprovalID(ctx, approvalID)
		if err == nil && association != nil {
			switch association.RiskLevel {
			case "CRITICAL":
				log.Printf("[冲突阻断] 审批ID=%s 检测到CRITICAL级别冲突，阻断审批通过", approvalID)
				return nil, fmt.Errorf("检测到严重利益冲突(CRITICAL)，无法通过审批")
			case "HIGH":
				log.Printf("[冲突升级] 审批ID=%s 检测到HIGH级别冲突，需要升级审批", approvalID)
				// HIGH风险需要额外审批，标记需要升级
				if !association.RequiresApproval {
					_ = s.integrationRepo.UpdateConflictAssociationStatus(ctx, association.ID, "active")
				}
				return nil, fmt.Errorf("检测到高风险冲突(HIGH)，需要升级至更高级别审批人员处理")
			}
		}
	}

	// 处理审批决定
	approvalUserID := userID
	approvalUserName := userName
	if approval.CurrentApproverID != "" {
		approvalUserID = approval.CurrentApproverID
	}
	if approval.CurrentApproverName != "" {
		approvalUserName = approval.CurrentApproverName
	}
	updatedApproval, err := s.approvalService.ProcessApproval(approvalUserID, approvalUserName, approvalID, decisionReq)
	if err != nil {
		return nil, fmt.Errorf("处理审批决定失败：%w", err)
	}

	// “同意并成案”必须在审批通过后立即创建并回填正式案件。
	if decisionReq.Decision == models.ApprovalDecisionApprove {
		if _, err := s.ensureApprovedApprovalHasCase(ctx, approval, userID, userName); err != nil {
			log.Printf("自动创建案件失败：%v", err)
			// 不影响审批结果，只记录错误
			_ = s.updateCaseTracking(ctx, approvalID, "failed", nil, nil, err.Error())
		} else if refreshed, err := s.approvalService.GetApprovalByID(approvalID); err == nil && refreshed != nil {
			updatedApproval = refreshed
		}
	}

	log.Printf("审批申请处理完成，ID：%s，最终状态：%s", updatedApproval.ID, updatedApproval.Status)
	return updatedApproval, nil
}

// Helper functions

func buildApprovalSnapshot(userID, userName string, req *IntegrationRequest, metadata map[string]interface{}) map[string]interface{} {
	snapshotMetadata := copyMetadata(metadata)
	delete(snapshotMetadata, "approval_snapshot")

	snapshot := map[string]interface{}{
		"snapshot_id":   fmt.Sprintf("SNAP_%d", time.Now().UnixNano()),
		"snapshot_at":   time.Now().UTC().Format(time.RFC3339Nano),
		"snapshot_type": "conflict_approval",
		"applicant": map[string]interface{}{
			"id":              userID,
			"name":            userName,
			"submitted_name":  req.ApplicantName,
			"department_name": req.DepartmentName,
		},
		"approval": map[string]interface{}{
			"type":              req.Type,
			"title":             req.Title,
			"content":           req.Content,
			"category":          req.Category,
			"urgency":           req.Urgency,
			"priority":          req.Priority,
			"workflow_type":     req.WorkflowType,
			"expected_duration": req.ExpectedDuration,
		},
		"metadata": snapshotMetadata,
	}

	if req.ConflictCheckConfig != nil {
		snapshot["conflict_check_config"] = req.ConflictCheckConfig
	}
	if conflictResult, ok := metadata["conflict_result"]; ok {
		snapshot["conflict_result"] = conflictResult
	}
	if req.CaseCreationConfig != nil {
		snapshot["case_creation_config"] = copyMetadata(req.CaseCreationConfig)
	}

	return snapshot
}

func copyMetadata(source map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{}, len(source))
	for key, value := range source {
		copied[key] = value
	}
	return copied
}

func shouldExposeCaseCreation(metadata map[string]interface{}, approval *models.ApprovalRequest) bool {
	if approval.CaseCreated || approval.CreatedCaseID != "" || approval.CaseCreationStatus != "" {
		return true
	}
	if approval.Status == models.ApprovalStatusApproved {
		return true
	}
	if _, ok := metadata["case_creation_config"]; ok {
		return true
	}
	return metadataString(metadata, "integration_type") == "both"
}

func (s *approvalConflictIntegrationService) buildCaseCreationStatus(ctx context.Context, approvalID string, approval *models.ApprovalRequest) *CaseCreationStatus {
	caseStatus := &CaseCreationStatus{Status: "pending"}
	if approval.CaseCreationStatus != "" {
		caseStatus.Status = approval.CaseCreationStatus
	}
	if approval.CreatedCaseID != "" {
		caseStatus.CaseID = approval.CreatedCaseID
	}
	if approval.CaseCreationTime != nil {
		caseStatus.LastAttempted = *approval.CaseCreationTime
	}
	if tracking, err := s.integrationRepo.GetLatestCaseCreationTracking(ctx, approvalID); err == nil && tracking != nil {
		if tracking.CreationStatus != "" {
			caseStatus.Status = tracking.CreationStatus
		}
		caseStatus.RetryCount = tracking.RetryCount
		if tracking.CaseID != nil {
			caseStatus.CaseID = *tracking.CaseID
		}
		if tracking.CaseNumber != nil {
			caseStatus.CaseNumber = *tracking.CaseNumber
		}
		if tracking.ErrorMessage != nil {
			caseStatus.LastError = *tracking.ErrorMessage
		}
		if tracking.CompletedAt != nil {
			caseStatus.LastAttempted = *tracking.CompletedAt
		} else if tracking.ProcessedAt != nil {
			caseStatus.LastAttempted = *tracking.ProcessedAt
		}
	}
	if caseStatus.CaseID != "" && caseStatus.Status == "pending" {
		caseStatus.Status = "completed"
	}
	return caseStatus
}

func (s *approvalConflictIntegrationService) ensureApprovedApprovalHasCase(ctx context.Context, approval *models.ApprovalRequest, userID, userName string) (*CaseCreationResult, error) {
	if approval.CaseCreated && approval.CreatedCaseID != "" {
		return &CaseCreationResult{
			CaseID:    approval.CreatedCaseID,
			Status:    coalesce(approval.CaseCreationStatus, "completed"),
			Message:   "案件已创建",
			CreatedAt: time.Now(),
			CaseData:  map[string]interface{}{},
		}, nil
	}
	caseData, err := s.buildCaseDataForApprovedApproval(ctx, approval, userID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCaseTracking(ctx, approval.ID, "processing", caseData, userID); err != nil {
		log.Printf("创建案件跟踪记录失败：%v", err)
	}
	result, err := s.AutoCreateCaseFromApproval(ctx, approval.ID, caseData)
	if err != nil {
		return nil, err
	}
	if result != nil {
		_ = s.updateCaseTracking(ctx, approval.ID, "completed", &result.CaseID, &result.CaseNumber, "")
		log.Printf("审批通过后已自动成案，审批ID=%s，操作人=%s，案件ID=%s", approval.ID, userName, result.CaseID)
	}
	return result, nil
}

func (s *approvalConflictIntegrationService) buildCaseDataForApprovedApproval(ctx context.Context, approval *models.ApprovalRequest, userID string) (map[string]interface{}, error) {
	metadata := parseMetadata(approval.Metadata)
	if config, ok := metadata["case_creation_config"].(map[string]interface{}); ok {
		return normalizeCaseCreationConfig(config, approval, userID), nil
	}

	var record *models.ConflictCheckRecord
	if checkID := conflictCheckIDFromMetadata(metadata); checkID != "" {
		if found, err := s.integrationRepo.GetConflictCheckRecord(ctx, checkID); err != nil {
			return nil, err
		} else {
			record = found
		}
	}

	caseData := map[string]interface{}{
		"title":          cleanedApprovalCaseTitle(approval.Title),
		"description":    approval.Content,
		"case_type":      coalesce(approval.Type, "commercial"),
		"priority":       coalesce(approval.Priority, "medium"),
		"billing_method": "hourly",
	}
	if record != nil {
		if record.CaseName != "" {
			caseData["title"] = record.CaseName
		}
		if record.CaseType != "" {
			caseData["case_type"] = record.CaseType
		}
		if clientID, ok := parseUintString(record.ClientID); ok {
			caseData["client_id"] = clientID
		}
		if record.UserID > 0 {
			caseData["lawyer_id"] = record.UserID
			caseData["assigned_by"] = record.UserID
		}
	}
	if _, ok := caseData["client_id"]; !ok {
		if clientID, ok := metadataUint(metadata, "client_id"); ok {
			caseData["client_id"] = clientID
		}
	}
	if _, ok := caseData["lawyer_id"]; !ok {
		if lawyerID, ok := parseUintString(approval.ApplicantID); ok {
			caseData["lawyer_id"] = lawyerID
			caseData["assigned_by"] = lawyerID
		}
	}
	if _, ok := caseData["assigned_by"]; !ok {
		if assignedBy, ok := parseUintString(userID); ok {
			caseData["assigned_by"] = assignedBy
		}
	}
	if _, ok := caseData["client_id"]; !ok {
		return nil, fmt.Errorf("无法从审批或冲突检测记录中确定客户ID")
	}
	if _, ok := caseData["lawyer_id"]; !ok {
		return nil, fmt.Errorf("无法从审批或冲突检测记录中确定主办律师ID")
	}
	return caseData, nil
}

func normalizeCaseCreationConfig(config map[string]interface{}, approval *models.ApprovalRequest, userID string) map[string]interface{} {
	caseData := copyMetadata(config)
	if _, ok := caseData["title"]; !ok {
		caseData["title"] = cleanedApprovalCaseTitle(approval.Title)
	}
	if _, ok := caseData["description"]; !ok {
		caseData["description"] = approval.Content
	}
	if _, ok := caseData["priority"]; !ok {
		caseData["priority"] = coalesce(approval.Priority, "medium")
	}
	if _, ok := caseData["billing_method"]; !ok {
		caseData["billing_method"] = "hourly"
	}
	if _, ok := caseData["assigned_by"]; !ok {
		if assignedBy, ok := parseUintString(userID); ok {
			caseData["assigned_by"] = assignedBy
		}
	}
	return caseData
}

func (s *approvalConflictIntegrationService) ensureCaseTracking(ctx context.Context, approvalID, status string, caseData map[string]interface{}, userID string) error {
	tracking, err := s.integrationRepo.GetLatestCaseCreationTracking(ctx, approvalID)
	if err != nil {
		return err
	}
	if tracking != nil {
		return s.integrationRepo.UpdateCaseCreationTracking(ctx, tracking.ID, map[string]interface{}{
			"creation_status": status,
			"processed_at":    time.Now(),
			"data_mapping":    jsonString(caseData),
		})
	}
	return s.integrationRepo.CreateCaseCreationTracking(ctx, &repositories.CaseCreationTracking{
		ID:                  fmt.Sprintf("CASE_TRACK_%s_%d", approvalID, time.Now().UnixNano()),
		ApprovalRequestID:   approvalID,
		CreationStatus:      status,
		ProgressPercentage:  10,
		ErrorDetails:        "{}",
		DataMapping:         jsonString(caseData),
		MappedFields:        "{}",
		UnmappedFields:      "{}",
		AppliedConditions:   "{}",
		ImposedRequirements: "{}",
		WorkflowActions:     "[]",
		CreatedBy:           coalesce(userID, "0"),
		CreatedAt:           time.Now(),
	})
}

func parseMetadata(raw string) map[string]interface{} {
	metadata := map[string]interface{}{}
	if raw != "" {
		_ = json.Unmarshal([]byte(raw), &metadata)
	}
	return metadata
}

func conflictCheckIDFromMetadata(metadata map[string]interface{}) string {
	if value := metadataString(metadata, "conflict_task_id"); value != "" {
		return value
	}
	if value := metadataString(metadata, "conflict_check_id"); value != "" {
		return value
	}
	if conflictResult, ok := metadata["conflict_result"].(map[string]interface{}); ok {
		for _, key := range []string{"checkId", "check_id", "id"} {
			if value := metadataString(conflictResult, key); value != "" {
				return value
			}
		}
	}
	if conflictCheck, ok := metadata["conflict_check"].(map[string]interface{}); ok {
		for _, key := range []string{"conflict_check_id", "check_id", "checkId"} {
			if value := metadataString(conflictCheck, key); value != "" {
				return value
			}
		}
	}
	return ""
}

func metadataString(metadata map[string]interface{}, key string) string {
	if value, ok := metadata[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func metadataUint(metadata map[string]interface{}, key string) (uint, bool) {
	switch value := metadata[key].(type) {
	case float64:
		if value > 0 {
			return uint(value), true
		}
	case int:
		if value > 0 {
			return uint(value), true
		}
	case string:
		return parseUintString(value)
	}
	return 0, false
}

func parseUintString(value string) (uint, bool) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed == 0 {
		return 0, false
	}
	return uint(parsed), true
}

func cleanedApprovalCaseTitle(title string) string {
	cleaned := strings.TrimSpace(title)
	for _, prefix := range []string{"冲突审查审批 - ", "新建案件审批 - ", "立案审批 - "} {
		cleaned = strings.TrimPrefix(cleaned, prefix)
	}
	if cleaned == "" {
		return "审批通过自动成案"
	}
	return cleaned
}

func coalesce(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func jsonString(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (s *approvalConflictIntegrationService) updateCaseTracking(ctx context.Context, approvalID, status string, caseID, caseNumber *string, message string) error {
	tracking, err := s.integrationRepo.GetLatestCaseCreationTracking(ctx, approvalID)
	if err != nil || tracking == nil {
		if err != nil {
			return err
		}
		return s.ensureCaseTracking(ctx, approvalID, status, map[string]interface{}{}, userIDFromContext(ctx))
	}
	now := time.Now()
	updates := map[string]interface{}{
		"creation_status": status,
		"processed_at":    now,
	}
	if status == "completed" {
		updates["completed_at"] = now
		updates["progress_percentage"] = 100
	}
	if caseID != nil {
		updates["case_id"] = *caseID
	}
	if caseNumber != nil {
		updates["case_number"] = *caseNumber
	}
	if message != "" {
		updates["error_message"] = message
	}
	return s.integrationRepo.UpdateCaseCreationTracking(ctx, tracking.ID, updates)
}

// userIDFromContext 从上下文获取用户ID
func userIDFromContext(ctx context.Context) string {
	// 从 context 中获取由 JWT 中间件设置的 user_id (uint)
	if userID, ok := ctx.Value("user_id").(uint); ok {
		return fmt.Sprintf("%d", userID)
	}
	return "0"
}
