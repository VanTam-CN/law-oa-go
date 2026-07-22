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

	// AutoCreateCaseFromApprovalForUser 从审批申请自动创建案件，并按当前用户校验审批可见性与状态
	AutoCreateCaseFromApprovalForUser(ctx context.Context, userID, approvalID string, caseData map[string]interface{}) (*CaseCreationResult, error)

	// GetIntegrationStatus 获取集成状态
	GetIntegrationStatus(ctx context.Context, approvalID string) (*IntegrationStatus, error)

	// GetIntegrationStatusForUser 获取集成状态，并按当前用户校验审批可见性
	GetIntegrationStatusForUser(ctx context.Context, userID, approvalID string) (*IntegrationStatus, error)

	// ProcessApprovalWithConflict 处理包含冲突检测的审批申请
	ProcessApprovalWithConflict(ctx context.Context, userID, userName string, approvalID string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error)

	// RetryCaseCreationForUser retries only the failed case-linking step. It
	// never creates a second case for an approval that is already linked.
	RetryCaseCreationForUser(ctx context.Context, userID, approvalID string) (*CaseCreationResult, error)
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
	Accessible    *bool     `json:"accessible,omitempty"`
	LastError     string    `json:"last_error,omitempty"`
	RetryCount    int       `json:"retry_count,omitempty"`
	LastAttempted time.Time `json:"last_attempted,omitempty"`
}

// ApprovalConflictGateError is a user-actionable, fail-closed conflict gate.
// It must be returned as a conflict response instead of a generic server
// error, otherwise a lawyer cannot tell whether the approval is retryable.
type ApprovalConflictGateError struct {
	Code    string
	Message string
}

func (e *ApprovalConflictGateError) Error() string {
	if e == nil {
		return "冲突审批门禁失败"
	}
	return e.Message
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
	if s == nil || s.caseService == nil {
		return nil, NewSubjectWorkflowError("CASE_CREATION_GATE_UNAVAILABLE", "正式成案服务未初始化，已阻止成案")
	}
	if approval, err := s.approvalService.GetApprovalByID(approvalID); err != nil {
		return nil, fmt.Errorf("读取审批成案状态失败：%w", err)
	} else if approval != nil && approval.CaseCreated && approval.CreatedCaseID != "" {
		return &CaseCreationResult{CaseID: approval.CreatedCaseID, Status: "completed", Message: "案件已创建", CreatedAt: time.Now(), CaseData: caseData}, nil
	}
	if tracking, err := s.integrationRepo.GetLatestCaseCreationTracking(ctx, approvalID); err != nil {
		return nil, fmt.Errorf("读取案件创建跟踪失败：%w", err)
	} else if tracking != nil && tracking.CreationStatus == "completed" && tracking.CaseID != nil && strings.TrimSpace(*tracking.CaseID) != "" {
		caseNumber := ""
		if tracking.CaseNumber != nil {
			caseNumber = *tracking.CaseNumber
		}
		association := &models.CaseCreationAssociation{
			Created: true, CaseID: *tracking.CaseID, CaseNumber: caseNumber,
			CreationTime: time.Now(), DataMapping: caseData, Status: "completed", StatusMessage: "案件创建成功",
		}
		if err := s.integrationRepo.UpdateApprovalCaseAssociation(ctx, approvalID, association); err != nil {
			_ = s.integrationRepo.MarkCaseCreationFailed(ctx, approvalID, fmt.Sprintf("正式案件已存在但审批关联回填失败：%v", err))
			return nil, fmt.Errorf("正式案件已存在但审批关联回填失败：%w", err)
		}
		return &CaseCreationResult{CaseID: *tracking.CaseID, CaseNumber: caseNumber, Status: "completed", Message: "案件已创建", CreatedAt: time.Now(), CaseData: caseData}, nil
	}
	claimed, err := s.integrationRepo.ClaimCaseCreation(ctx, approvalID, userIDFromContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("锁定案件创建任务失败：%w", err)
	}
	if !claimed {
		if approval, lookupErr := s.approvalService.GetApprovalByID(approvalID); lookupErr == nil && approval != nil && approval.CaseCreated && approval.CreatedCaseID != "" {
			return &CaseCreationResult{CaseID: approval.CreatedCaseID, Status: "completed", Message: "案件已创建", CreatedAt: time.Now(), CaseData: caseData}, nil
		}
		return nil, NewSubjectWorkflowError("CASE_CREATION_IN_PROGRESS", "该审批的正式案件正在创建或等待重试，请稍后查看成案状态")
	}

	createCaseReq, err := caseCreationRequestFromMap(caseData)
	if err != nil {
		_ = s.integrationRepo.MarkCaseCreationFailed(ctx, approvalID, err.Error())
		return nil, err
	}
	createCaseReq.Approved = true

	// 调用CaseService创建案件
	createdCase, err := s.caseService.CreateCase(ctx, createCaseReq)
	if err != nil {
		log.Printf("创建案件失败: %v", err)
		_ = s.integrationRepo.MarkCaseCreationFailed(ctx, approvalID, err.Error())
		return nil, fmt.Errorf("创建案件失败: %w", err)
	}

	caseID := fmt.Sprintf("%d", createdCase.ID)
	caseNumber := createdCase.CaseNumber
	if caseNumber == "" {
		caseNumber = fmt.Sprintf("CASE-%d", createdCase.ID)
	}
	// Persist the completed tracking record before linking the approval. If
	// only the link update fails, a retry can reconcile this case ID instead of
	// creating a duplicate case.
	_ = s.updateCaseTracking(ctx, approvalID, "completed", &caseID, &caseNumber, "")

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
		_ = s.integrationRepo.MarkCaseCreationFailed(ctx, approvalID, fmt.Sprintf("案件已创建但审批关联回填失败：%v", err))
		return nil, fmt.Errorf("案件已创建但正式案件ID回填失败：%w", err)
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

// RetryCaseCreationForUser retries a previously failed post-approval case
// creation. The authoritative case data is rebuilt from the frozen approval
// snapshot and conflict record; callers cannot replace the client, lawyer or
// conflict-check binding during a retry.
func (s *approvalConflictIntegrationService) RetryCaseCreationForUser(ctx context.Context, userID, approvalID string) (*CaseCreationResult, error) {
	approval, err := s.approvalService.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批申请失败：%w", err)
	}
	if !RequiresCaseCreationApproval(approval) {
		return nil, fmt.Errorf("该审批不包含正式成案步骤")
	}
	if approval.Status != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("审批通过后才能重试成案")
	}
	caseData, err := s.buildCaseDataForApprovedApproval(ctx, approval, userID)
	if err != nil {
		return nil, err
	}
	return s.AutoCreateCaseFromApprovalForUser(ctx, userID, approvalID, caseData)
}

func (s *approvalConflictIntegrationService) AutoCreateCaseFromApprovalForUser(ctx context.Context, userID, approvalID string, caseData map[string]interface{}) (*CaseCreationResult, error) {
	approval, err := s.approvalService.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批申请失败：%w", err)
	}
	if approval.Status != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("审批通过后才能创建案件")
	}
	metadata := parseMetadata(approval.Metadata)
	expectedCheckID := conflictCheckIDFromMetadata(metadata)
	providedCheckID := conflictCheckIDFromMetadata(caseData)
	if expectedCheckID == "" || providedCheckID == "" || providedCheckID != expectedCheckID {
		return nil, NewSubjectWorkflowError("CONFLICT_CHECK_MISMATCH", "成案数据必须绑定该审批记录对应的冲突复核记录")
	}
	record, err := s.integrationRepo.GetConflictCheckRecord(ctx, expectedCheckID)
	if err != nil || record == nil {
		return nil, NewSubjectWorkflowError("CONFLICT_CHECK_REQUIRED", "审批关联的冲突复核记录不存在，已阻止成案")
	}
	if recordClientID, ok := parseUintString(record.ClientID); ok {
		if requestedClientID, exists := metadataUint(caseData, "client_id"); exists && requestedClientID != recordClientID {
			return nil, NewSubjectWorkflowError("SUBJECT_CLIENT_MISMATCH", "成案客户必须与冲突复核记录一致")
		}
		caseData["client_id"] = recordClientID
	}
	if record.UserID > 0 {
		if requestedLawyerID, exists := metadataUint(caseData, "lawyer_id"); exists && requestedLawyerID != record.UserID {
			return nil, NewSubjectWorkflowError("SUBJECT_LAWYER_MISMATCH", "承办律师必须与冲突复核记录一致")
		}
		caseData["lawyer_id"] = record.UserID
		caseData["assigned_by"] = record.UserID
	}
	caseData["conflict_check_id"] = expectedCheckID
	return s.AutoCreateCaseFromApproval(ctx, approvalID, caseData)
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

func (s *approvalConflictIntegrationService) GetIntegrationStatusForUser(ctx context.Context, userID, approvalID string) (*IntegrationStatus, error) {
	approval, err := s.approvalService.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批申请失败：%w", err)
	}
	status, err := s.GetIntegrationStatus(ctx, approvalID)
	if err != nil {
		return nil, err
	}
	if status.CaseCreation != nil && status.CaseCreation.CaseID != "" && approval.ApplicantID == userID {
		accessible := false
		if caseID, parseErr := strconv.ParseUint(status.CaseCreation.CaseID, 10, 32); parseErr == nil {
			if caseDetail, caseErr := s.caseService.GetCaseByID(ctx, uint(caseID)); caseErr == nil && caseDetail != nil {
				if applicantID, applicantErr := strconv.ParseUint(userID, 10, 32); applicantErr == nil {
					accessible = caseDetail.LawyerID == uint(applicantID)
				}
			}
		}
		status.CaseCreation.Accessible = &accessible
	}
	return status, nil
}

// ProcessApprovalWithConflict 处理包含冲突检测的审批申请
func (s *approvalConflictIntegrationService) ProcessApprovalWithConflict(ctx context.Context, userID, userName string, approvalID string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	if decisionReq == nil {
		return nil, fmt.Errorf("审批决定不能为空")
	}
	log.Printf("处理包含冲突检测的审批申请，审批ID：%s，决定：%s", approvalID, decisionReq.Decision)

	// 获取当前的审批申请
	approval, err := s.approvalService.GetApproval(userID, approvalID)
	if err != nil {
		return nil, fmt.Errorf("获取审批申请失败：%w", err)
	}

	requiresCaseCreation := RequiresCaseCreationApproval(approval)
	var latestConflictReview *models.ConflictReview
	if decisionReq.Decision == models.ApprovalDecisionApprove && requiresCaseCreation {
		latestConflictReview, err = s.validateConflictApprovalGate(ctx, approval)
		if err != nil {
			return nil, err
		}
	}
	var preflightCaseData map[string]interface{}
	// Formal case creation is preflighted before the approval state changes. A
	// missing review, incomplete archive coverage or invalid owner therefore
	// leaves the approval awaiting action instead of producing a false success.
	if decisionReq.Decision == models.ApprovalDecisionApprove && requiresCaseCreation {
		if s.caseService == nil {
			return nil, NewSubjectWorkflowError("CASE_CREATION_GATE_UNAVAILABLE", "正式成案服务未初始化，已阻止审批通过")
		}
		preflightCaseData, err = s.buildCaseDataForApprovedApproval(ctx, approval, userID)
		if err != nil {
			return nil, fmt.Errorf("正式成案前置校验失败：%w", err)
		}
		preflightRequest, requestErr := caseCreationRequestFromMap(preflightCaseData)
		if requestErr != nil {
			return nil, requestErr
		}
		preflightRequest.Approved = true
		if err := s.caseService.ValidateApprovedCase(ctx, preflightRequest); err != nil {
			return nil, fmt.Errorf("正式成案前置校验失败：%w", err)
		}
	}

	// 审批通过前检查冲突风险等级
	if decisionReq.Decision == models.ApprovalDecisionApprove {
		association, err := s.integrationRepo.GetConflictAssociationByApprovalID(ctx, approvalID)
		if err == nil && association != nil {
			switch association.RiskLevel {
			case "CRITICAL":
				log.Printf("[冲突阻断] 审批ID=%s 检测到CRITICAL级别冲突，阻断审批通过", approvalID)
				return nil, &ApprovalConflictGateError{Code: "CONFLICT_CRITICAL", Message: "检测到严重利益冲突，当前审批不能同意并成案"}
			case "HIGH":
				if latestConflictReview != nil && isProceedingConflictReview(latestConflictReview.Decision) {
					break
				}
				log.Printf("[冲突升级] 审批ID=%s 检测到HIGH级别冲突，需要升级审批", approvalID)
				// HIGH风险需要额外审批，标记需要升级
				if !association.RequiresApproval {
					_ = s.integrationRepo.UpdateConflictAssociationStatus(ctx, association.ID, "active")
				}
				return nil, &ApprovalConflictGateError{Code: "CONFLICT_HIGH_ESCALATION_REQUIRED", Message: "检测到高风险冲突，当前审批不能直接同意并成案，请升级至更高级别审批人员处理"}
			}
		}
	}

	// 处理审批决定：审批服务必须使用当前登录用户校验是否为实际审批人。
	updatedApproval, err := s.approvalService.ProcessApproval(userID, userName, approvalID, decisionReq)
	if err != nil {
		return nil, fmt.Errorf("处理审批决定失败：%w", err)
	}

	// “同意并成案”必须在审批通过后立即创建并回填正式案件。只有明确
	// 绑定成案的审批类型才进入此分支，普通合同/财务等审批不会误成案。
	if decisionReq.Decision == models.ApprovalDecisionApprove && requiresCaseCreation {
		if _, err := s.ensureApprovedApprovalHasCase(ctx, approval, userID, userName); err != nil {
			log.Printf("自动创建案件失败：%v", err)
			_ = s.updateCaseTracking(ctx, approvalID, "failed", nil, nil, err.Error())
			return nil, fmt.Errorf("审批已记录，但正式成案失败，已进入失败重试状态：%w", err)
		} else if refreshed, err := s.approvalService.GetApprovalByID(approvalID); err == nil && refreshed != nil {
			updatedApproval = refreshed
		}
	}

	log.Printf("审批申请处理完成，ID：%s，最终状态：%s", updatedApproval.ID, updatedApproval.Status)
	return updatedApproval, nil
}

func isProceedingConflictReview(decision string) bool {
	switch strings.ToLower(strings.TrimSpace(decision)) {
	case "no_conflict", "false_positive":
		return true
	default:
		return false
	}
}

// validateConflictApprovalGate checks the immutable conflict evidence and the
// latest independent conclusion immediately before a formal case can be
// created. A stale approval snapshot or a browser-side button state cannot
// bypass this check.
func (s *approvalConflictIntegrationService) validateConflictApprovalGate(ctx context.Context, approval *models.ApprovalRequest) (*models.ConflictReview, error) {
	metadata := parseMetadata(approval.Metadata)
	checkID := conflictCheckIDFromMetadata(metadata)
	association, err := s.integrationRepo.GetConflictAssociationByApprovalID(ctx, approval.ID)
	if err != nil {
		return nil, fmt.Errorf("读取审批冲突关联失败：%w", err)
	}
	if association != nil {
		checkID = association.ConflictCheckID
	}
	if checkID == "" && !strings.EqualFold(strings.TrimSpace(approval.Type), "conflict_approval") {
		return nil, nil
	}
	if checkID == "" {
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_CHECK_LINK_REQUIRED", Message: "审批缺少可验证的冲突检测记录，已阻止同意并成案"}
	}

	record, err := s.integrationRepo.GetConflictCheckRecord(ctx, checkID)
	if err != nil {
		return nil, fmt.Errorf("读取冲突检测记录失败：%w", err)
	}
	if record == nil {
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_CHECK_NOT_FOUND", Message: "关联的冲突检测记录不存在，已阻止同意并成案"}
	}
	var response models.ConflictCheckResponse
	raw, marshalErr := json.Marshal(record.CheckResult)
	if marshalErr != nil || json.Unmarshal(raw, &response) != nil || response.Decision == nil || !conflictRecordCoverageComplete(record, &response) {
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_COVERAGE_INCOMPLETE", Message: "冲突检测覆盖范围未确认完整，已阻止同意并成案"}
	}

	review, err := s.integrationRepo.GetLatestConflictReview(ctx, checkID)
	if err != nil {
		return nil, fmt.Errorf("读取冲突复核结论失败：%w", err)
	}
	if review == nil {
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_REVIEW_REQUIRED", Message: "冲突检测尚未完成独立人工复核，不能同意并成案"}
	}
	switch strings.ToLower(strings.TrimSpace(review.Decision)) {
	case "no_conflict", "false_positive":
		return review, nil
	case "insufficient_information":
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_REVIEW_INSUFFICIENT", Message: "冲突复核结论为信息不足，请补充主体身份材料并重新复核后再处理"}
	case "confirmed_conflict":
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_CONFIRMED", Message: "独立复核已确认存在利益冲突，不能同意并成案"}
	case "waiver_requested":
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_WAIVER_REQUIRED", Message: "冲突复核已申请豁免，取得有效豁免批准前不能同意并成案"}
	default:
		return nil, &ApprovalConflictGateError{Code: "CONFLICT_REVIEW_REQUIRED", Message: "冲突复核结论尚未达到可成案状态，不能同意并成案"}
	}
}

func conflictRecordCoverageComplete(record *models.ConflictCheckRecord, response *models.ConflictCheckResponse) bool {
	if response != nil && response.Decision != nil && strings.EqualFold(strings.TrimSpace(response.Decision.CoverageStatus), "COMPLETE") {
		return true
	}
	if record == nil {
		return false
	}
	var searchParameters map[string]interface{}
	raw, err := json.Marshal(record.SearchParameters)
	if err != nil || json.Unmarshal(raw, &searchParameters) != nil {
		return false
	}
	for key, value := range searchParameters {
		if strings.EqualFold(strings.TrimSpace(key), "coverageStatus") || strings.EqualFold(strings.TrimSpace(key), "coverage_status") {
			return strings.EqualFold(strings.TrimSpace(fmt.Sprint(value)), "COMPLETE")
		}
	}
	return false
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

func caseCreationRequestFromMap(caseData map[string]interface{}) (*CreateCaseRequest, error) {
	jsonData, err := json.Marshal(caseData)
	if err != nil {
		return nil, fmt.Errorf("序列化案件前置校验数据失败：%w", err)
	}
	var request CreateCaseRequest
	if err := json.Unmarshal(jsonData, &request); err != nil {
		return nil, fmt.Errorf("解析案件前置校验数据失败：%w", err)
	}
	return &request, nil
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

// RequiresCaseCreationApproval identifies only approval types that are
// allowed to trigger formal case creation. Generic approvals must never be
// able to create a case merely because they reached the approved state.
func RequiresCaseCreationApproval(approval *models.ApprovalRequest) bool {
	if approval == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(approval.Type)) {
	case "case_creation", "conflict_approval":
		return true
	}
	metadata := parseMetadata(approval.Metadata)
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
	var record *models.ConflictCheckRecord
	if checkID := conflictCheckIDFromMetadata(metadata); checkID != "" {
		found, err := s.integrationRepo.GetConflictCheckRecord(ctx, checkID)
		if err != nil {
			return nil, err
		}
		record = found
	}
	if config, ok := metadata["case_creation_config"].(map[string]interface{}); ok {
		caseData := normalizeCaseCreationConfig(config, approval, userID)
		if approval.Type == "conflict_approval" {
			applyConflictRecordOwnership(caseData, record)
		}
		return validateCaseCreationOwnership(caseData)
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
		applyConflictRecordOwnership(caseData, record)
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
	return validateCaseCreationOwnership(caseData)
}

func applyConflictRecordOwnership(caseData map[string]interface{}, record *models.ConflictCheckRecord) {
	if record == nil {
		return
	}
	if record.CheckID != "" {
		caseData["conflict_check_id"] = record.CheckID
	}
	if clientID, ok := parseUintString(record.ClientID); ok {
		caseData["client_id"] = clientID
	}
	if record.UserID > 0 {
		caseData["lawyer_id"] = record.UserID
		caseData["assigned_by"] = record.UserID
	}
}

func validateCaseCreationOwnership(caseData map[string]interface{}) (map[string]interface{}, error) {
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
