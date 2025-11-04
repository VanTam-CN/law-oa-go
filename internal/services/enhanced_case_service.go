package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// EnhancedCaseService 增强的案例服务
// 集成冲突检测、豁免管理、信息屏障等高级功能
type EnhancedCaseService struct {
	caseRepo                    repositories.CaseRepository
	clientRepo                  repositories.ClientRepository
	userRepo                    repositories.UserRepository
	enhancedConflictRepo        repositories.EnhancedConflictRepository
	conflictDetectionService    ConflictDetectionService
	conflictClassificationService ConflictClassificationService
}

// EnhancedCreateCaseRequest 增强的创建案例请求
type EnhancedCreateCaseRequest struct {
	// 基础字段
	Title              string                       `json:"title" binding:"required,min=1,max=200"`
	Description        string                       `json:"description" binding:"max=2000"`
	CaseType           string                       `json:"case_type" binding:"required"`
	Priority           string                       `json:"priority" binding:"required"`
	StartDate          string                       `json:"start_date,omitempty"`
	PracticeArea       string                       `json:"practice_area" binding:"required"`
	EstimatedDuration  string                       `json:"estimated_duration,omitempty"`
	BillingMethod      string                       `json:"billing_method" binding:"required"`

	// 客户信息 - 支持多客户
	ClientProfileIDs   []string                     `json:"client_profile_ids" binding:"required,min=1"`
	ClientRoles        map[string]ClientRoleInfo    `json:"client_roles,omitempty"` // 客户角色映射

	// 团队分配
	LawyerID           uint                         `json:"lawyer_id" binding:"required"`
	AssistingLawyerID  *uint                        `json:"assisting_lawyer_id,omitempty"`
	TeamMembers        []EnhancedTeamMemberRequest  `json:"team_members,omitempty"`

	// 冲突检测配置
	ConflictCheckConfig ConflictCheckConfig        `json:"conflict_check_config,omitempty"`

	// 分配信息
	AssignedBy         uint                         `json:"assigned_by"`
	IsMajorRisk        bool                         `json:"is_major_risk"`
}

// ClientRoleInfo 客户角色信息
type ClientRoleInfo struct {
	Role                  string   `json:"role"`                   // PRIMARY, SECONDARY, OPPOSING, THIRD_PARTY
	RelationshipDescription string   `json:"relationship_description,omitempty"`
	ContactInfo          string   `json:"contact_info,omitempty"`
}

// EnhancedTeamMemberRequest 增强的团队成员请求
type EnhancedTeamMemberRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	Role      string `json:"role" binding:"required"`
	Capacity  int    `json:"capacity,omitempty"`
	Screening bool   `json:"screening,omitempty"` // 是否需要信息屏障
}

// ConflictCheckConfig 冲突检测配置
type ConflictCheckConfig struct {
	SkipCheck              bool     `json:"skip_check,omitempty"`              // 跳过冲突检测
	CheckScope             []string `json:"check_scope,omitempty"`             // 检测范围
	ExcludedLawyers        []string `json:"excluded_lawyers,omitempty"`        // 排除的律师
	ExcludedCases          []string `json:"excluded_cases,omitempty"`          // 排除的案例
	UrgentCheck            bool     `json:"urgent_check,omitempty"`            // 紧急检测
	NotifyParties          []string `json:"notify_parties,omitempty"`          // 通知相关方
	AutoWaiverIfPossible   bool     `json:"auto_waiver_if_possible,omitempty"` // 自动申请豁免
}

// EnhancedCaseResponse 增强的案例响应
type EnhancedCaseResponse struct {
	// 基础字段
	ID          uint                   `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	CaseType    string                 `json:"case_type"`
	Priority    string                 `json:"priority"`
	Status      string                 `json:"status"`
	StartDate   *time.Time             `json:"start_date"`
	EndDate     *time.Time             `json:"end_date"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`

	// 客户信息
	ClientProfiles    []EnhancedClientInfo   `json:"client_profiles"`
	ClientProfileIDs  []string               `json:"client_profile_ids"`

	// 团队信息
	TeamAssignment    TeamAssignmentInfo     `json:"team_assignment"`

	// 冲突检测信息
	ConflictDetection ConflictDetectionInfo  `json:"conflict_detection"`
	RiskLevel         string                 `json:"risk_level"`

	// 豁免信息
	WaiverInfo        WaiverInfo             `json:"waiver_info"`

	// 信息屏障信息
	EthicalScreens    []EthicalScreenInfo    `json:"ethical_screens,omitempty"`

	// 业务信息
	PracticeArea      string                 `json:"practice_area"`
	EstimatedDuration string                 `json:"estimated_duration"`
	BillingMethod     string                 `json:"billing_method"`

	// 元数据
	CreatedViaConflictCheck bool               `json:"created_via_conflict_check"`
	ConflictMetadata        ConflictMetadata    `json:"conflict_metadata"`
}

// EnhancedClientInfo 增强的客户信息
type EnhancedClientInfo struct {
	ClientProfileID   string    `json:"client_profile_id"`
	ClientNumber      string    `json:"client_number"`
	ClientName        string    `json:"client_name"`
	ClientType        string    `json:"client_type"`
	ClientCategory    string    `json:"client_category"`
	Role              string    `json:"role"`
	RelationshipDescription *string `json:"relationship_description,omitempty"`
	PrimaryContact    *ContactInfo `json:"primary_contact,omitempty"`
}

// ContactInfo 联系信息
type ContactInfo struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// TeamAssignmentInfo 团队分配信息
type TeamAssignmentInfo struct {
	PrimaryLawyer      UserInfo             `json:"primary_lawyer"`
	AssistingLawyers   []UserInfo           `json:"assisting_lawyers,omitempty"`
	TeamMembers        []EnhancedTeamMember `json:"team_members,omitempty"`
	AllocationRules    map[string]interface{} `json:"allocation_rules,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Seniority  string `json:"seniority"`
}

// EnhancedTeamMember 增强的团队成员
type EnhancedTeamMember struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Capacity  int    `json:"capacity"`
	Screening bool   `json:"screening"`
}

// ConflictDetectionInfo 冲突检测信息
type ConflictDetectionInfo struct {
	Status              string                      `json:"status"`
	LastCheckTime       *time.Time                  `json:"last_check_time,omitempty"`
	TotalConflicts      int                         `json:"total_conflicts"`
	HighRiskConflicts   int                         `json:"high_risk_conflicts"`
	WaiverRequired      bool                        `json:"waiver_required"`
	WaiverPossible      bool                        `json:"waiver_possible"`
	DetectionSummary    string                      `json:"detection_summary,omitempty"`
	NextSteps           []string                    `json:"next_steps,omitempty"`
	Conflicts           []models.ConflictItem       `json:"conflicts,omitempty"`
}

// WaiverInfo 豁免信息
type WaiverInfo struct {
	ApplicationID      *string                    `json:"application_id,omitempty"`
	Status             string                     `json:"status"`
	Type               string                     `json:"type,omitempty"`
	ConsentObtained    bool                       `json:"consent_obtained"`
	ApprovalStatus     string                     `json:"approval_status,omitempty"`
	MonitoringRequired bool                       `json:"monitoring_required"`
	ExpiryDate         *time.Time                 `json:"expiry_date,omitempty"`
	Conditions         []string                   `json:"conditions,omitempty"`
}

// EthicalScreenInfo 信息屏障信息
type EthicalScreenInfo struct {
	ID                   string                 `json:"id"`
	ScreenType           string                 `json:"screen_type"`
	Status               string                 `json:"status"`
	EffectiveDate        time.Time              `json:"effective_date"`
	ExpiryDate           *time.Time             `json:"expiry_date,omitempty"`
	EstablishedBy        UserInfo               `json:"established_by"`
	ScreenedLawyers      []models.ScreenedLawyer `json:"screened_lawyers,omitempty"`
	ScreenedTeams        []models.ScreenedTeam   `json:"screened_teams,omitempty"`
	AccessPermissions    models.AccessPermissions `json:"access_permissions,omitempty"`
}

// NewEnhancedCaseService 创建增强的案例服务
func NewEnhancedCaseService(
	caseRepo repositories.CaseRepository,
	clientRepo repositories.ClientRepository,
	userRepo repositories.UserRepository,
	enhancedConflictRepo repositories.EnhancedConflictRepository,
	conflictDetectionService ConflictDetectionService,
	conflictClassificationService ConflictClassificationService,
) *EnhancedCaseService {
	return &EnhancedCaseService{
		caseRepo:                    caseRepo,
		clientRepo:                  clientRepo,
		userRepo:                    userRepo,
		enhancedConflictRepo:        enhancedConflictRepo,
		conflictDetectionService:    conflictDetectionService,
		conflictClassificationService: conflictClassificationService,
	}
}

// CreateEnhancedCase 创建增强案例
func (s *EnhancedCaseService) CreateEnhancedCase(ctx context.Context, req *EnhancedCreateCaseRequest) (*EnhancedCaseResponse, error) {
	// 1. 验证客户档案是否存在
	clientProfiles, err := s.validateClientProfiles(ctx, req.ClientProfileIDs)
	if err != nil {
		return nil, fmt.Errorf("客户档案验证失败: %w", err)
	}

	// 2. 验证团队成员
	teamMembers, err := s.validateTeamMembers(ctx, req.LawyerID, req.AssistingLawyerID, req.TeamMembers)
	if err != nil {
		return nil, fmt.Errorf("团队成员验证失败: %w", err)
	}

	// 3. 解析开始日期
	var startDate *time.Time
	if req.StartDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		startDate = &parsedDate
	}

	// 4. 创建增强案例实体
	enhancedCase := &models.EnhancedCase{
		Title:                   req.Title,
		Description:             req.Description,
		CaseType:                req.CaseType,
		Priority:                req.Priority,
		Status:                  "pending",
		StartDate:               startDate,
		ClientProfileIDs:        models.ClientProfileIDs(req.ClientProfileIDs),
		PracticeArea:            req.PracticeArea,
		EstimatedDuration:       req.EstimatedDuration,
		BillingMethod:           req.BillingMethod,
		AssignedBy:              func() *string { s := fmt.Sprintf("%d", req.AssignedBy); return &s }(),
		CreatedViaConflictCheck: false,
		ConflictDetectionStatus: models.ConflictDetectionStatusPending,
		RiskLevel:               models.RiskLevelNotAssessed,
		WaiverStatus:            models.WaiverStatusNone,
	}

	// 5. 如果不跳过冲突检测，执行冲突检测流程
	if !req.ConflictCheckConfig.SkipCheck {
		conflictResult, err := s.executeConflictDetection(ctx, enhancedCase, clientProfiles, req.ConflictCheckConfig)
		if err != nil {
			return nil, fmt.Errorf("冲突检测失败: %w", err)
		}

		// 更新冲突检测结果
		enhancedCase.ConflictDetectionStatus = models.ConflictDetectionStatusCompleted
		enhancedCase.RiskLevel = models.RiskLevel(conflictResult.RiskLevel)
		enhancedCase.ConflictDetectionResult = &conflictResult

		// 如果需要豁免且配置了自动申请，创建豁免申请
		if conflictResult.WaiverRequired && req.ConflictCheckConfig.AutoWaiverIfPossible {
			waiverApp, err := s.createAutoWaiverApplication(ctx, enhancedCase, conflictResult)
			if err != nil {
				return nil, fmt.Errorf("自动豁免申请失败: %w", err)
			}
			waiverID := waiverApp.ID
			enhancedCase.WaiverApplicationID = &waiverID
			enhancedCase.WaiverStatus = models.WaiverStatusPending
		}
	} else {
		// 跳过冲突检测
		enhancedCase.ConflictDetectionStatus = models.ConflictDetectionStatusNotRequired
		enhancedCase.RiskLevel = models.RiskLevelLow
	}

	// 6. 构建团队分配信息
	teamAssignment := s.buildTeamAssignment(teamMembers)

	// 7. 设置团队分配信息
	teamAssignmentJSON, err := json.Marshal(teamAssignment)
	if err != nil {
		return nil, fmt.Errorf("团队分配信息序列化失败: %w", err)
	}
	enhancedCase.TeamAssignment = models.TeamAssignment{}
	err = json.Unmarshal(teamAssignmentJSON, &enhancedCase.TeamAssignment)
	if err != nil {
		return nil, fmt.Errorf("团队分配信息设置失败: %w", err)
	}

	// 8. 保存案例到数据库
	err = s.caseRepo.Create(ctx, (*models.Case)(enhancedCase))
	if err != nil {
		return nil, fmt.Errorf("创建案例失败: %w", err)
	}

	// 9. 如果需要建立信息屏障，创建屏障记录
	if s.needsEthicalScreen(enhancedCase, teamMembers) {
		err = s.createEthicalScreen(ctx, enhancedCase, teamMembers, req.AssignedBy)
		if err != nil {
			return nil, fmt.Errorf("创建信息屏障失败: %w", err)
		}
		enhancedCase.EthicalScreenEstablished = true
	}

	// 10. 构建响应
	response := s.convertToEnhancedResponse(enhancedCase, clientProfiles, teamMembers)

	return response, nil
}

// validateClientProfiles 验证客户档案
func (s *EnhancedCaseService) validateClientProfiles(ctx context.Context, clientProfileIDs []string) ([]*models.ClientProfile, error) {
	var clientProfiles []*models.ClientProfile

	for _, profileID := range clientProfileIDs {
		profile, err := s.enhancedConflictRepo.GetClientProfileByID(ctx, profileID)
		if err != nil {
			return nil, fmt.Errorf("获取客户档案失败 %s: %w", profileID, err)
		}
		if profile == nil {
			return nil, fmt.Errorf("客户档案不存在: %s", profileID)
		}
		if profile.ClientStatus != "ACTIVE" {
			return nil, fmt.Errorf("客户档案状态无效: %s (状态: %s)", profileID, profile.ClientStatus)
		}
		clientProfiles = append(clientProfiles, profile)
	}

	return clientProfiles, nil
}

// validateTeamMembers 验证团队成员
func (s *EnhancedCaseService) validateTeamMembers(ctx context.Context, lawyerID uint, assistingLawyerID *uint, teamMembers []EnhancedTeamMemberRequest) ([]TeamMemberInfo, error) {
	var members []TeamMemberInfo

	// 验证主办律师
	primaryLawyer, err := s.userRepo.FindByID(ctx, lawyerID)
	if err != nil {
		return nil, fmt.Errorf("获取主办律师失败: %w", err)
	}
	if primaryLawyer == nil {
		return nil, errors.New("主办律师不存在")
	}

	members = append(members, TeamMemberInfo{
		User:     *primaryLawyer,
		Role:     "primary_lawyer",
		Capacity: 100,
	})

	// 验证协办律师
	if assistingLawyerID != nil {
		assistingLawyer, err := s.userRepo.FindByID(ctx, *assistingLawyerID)
		if err != nil {
			return nil, fmt.Errorf("获取协办律师失败: %w", err)
		}
		if assistingLawyer == nil {
			return nil, errors.New("协办律师不存在")
		}

		members = append(members, TeamMemberInfo{
			User:     *assistingLawyer,
			Role:     "assisting_lawyer",
			Capacity: 80,
		})
	}

	// 验证其他团队成员
	for _, memberReq := range teamMembers {
		member, err := s.userRepo.FindByID(ctx, memberReq.UserID)
		if err != nil {
			return nil, fmt.Errorf("获取团队成员失败: %w", err)
		}
		if member == nil {
			return nil, fmt.Errorf("团队成员不存在: %d", memberReq.UserID)
		}

		capacity := memberReq.Capacity
		if capacity == 0 {
			capacity = 50
		}

		members = append(members, TeamMemberInfo{
			User:       *member,
			Role:       memberReq.Role,
			Capacity:   capacity,
			Screening:  memberReq.Screening,
		})
	}

	return members, nil
}

// TeamMemberInfo 团队成员信息（内部使用）
type TeamMemberInfo struct {
	User      models.User
	Role      string
	Capacity  int
	Screening bool
}

// executeConflictDetection 执行冲突检测
func (s *EnhancedCaseService) executeConflictDetection(ctx context.Context, enhancedCase *models.EnhancedCase, clientProfiles []*models.ClientProfile, config ConflictCheckConfig) (*models.ConflictDetectionResult, error) {
	// 构建冲突检测请求
	detectionRequest := &models.ProfessionalConflictCheckRequest{
		CheckNumber:  s.generateCheckNumber(),
		CheckType:    "NEW_MATTER_ENGAGEMENT",
		Priority:     "HIGH",
		RequestedBy:  *enhancedCase.AssignedBy,
		Department:   s.getDepartmentByLawyerID(ctx, enhancedCase.AssignedBy),
		Status:       "IN_PROGRESS",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 创建冲突检测请求
	err := s.enhancedConflictRepo.CreateConflictCheckRequest(ctx, detectionRequest)
	if err != nil {
		return nil, fmt.Errorf("创建冲突检测请求失败: %w", err)
	}

	// 设置关联ID
	enhancedCase.ConflictCheckRequestID = &detectionRequest.ID

	// 执行多维度冲突检测
	conflictResults, err := s.conflictDetectionService.PerformMultiDimensionalDetection(ctx, detectionRequest.ID, clientProfiles)
	if err != nil {
		return nil, fmt.Errorf("执行冲突检测失败: %w", err)
	}

	// 分类冲突并生成报告
	classificationResult, err := s.conflictClassificationService.ClassifyConflicts(ctx, conflictResults)
	if err != nil {
		return nil, fmt.Errorf("冲突分类失败: %w", err)
	}

	// 构建冲突检测结果
	result := &models.ConflictDetectionResult{
		TotalConflicts:      len(conflictResults),
		HighRiskConflicts:   classificationResult.HighRiskCount,
		MediumRiskConflicts: classificationResult.MediumRiskCount,
		LowRiskConflicts:    classificationResult.LowRiskCount,
		DetectionSummary:    classificationResult.Summary,
		WaiverRequired:      classificationResult.WaiverRequired,
		WaiverPossible:      classificationResult.WaiverPossible,
		DetectedConflicts:   s.convertToConflictItems(conflictResults),
		Recommendations:     classificationResult.Recommendations,
		NextSteps:          classificationResult.NextSteps,
	}

	// 更新冲突检测请求状态
	detectionRequest.Status = "COMPLETED"
	detectionRequest.UpdatedAt = time.Now()
	err = s.enhancedConflictRepo.UpdateConflictCheckRequest(ctx, detectionRequest)
	if err != nil {
		return nil, fmt.Errorf("更新冲突检测状态失败: %w", err)
	}

	return result, nil
}

// createAutoWaiverApplication 创建自动豁免申请
func (s *EnhancedCaseService) createAutoWaiverApplication(ctx context.Context, enhancedCase *models.EnhancedCase, conflictResult *models.ConflictDetectionResult) (*models.WaiverApplication, error) {
	// 构建豁免申请
	waiverApp := &models.WaiverApplication{
		ApplicationNumber:        s.generateWaiverNumber(),
		WaiverType:              "INFORMED_CONSENT",
		WaiverCategory:          "CLIENT_CONSENT",
		ConflictSummary:         conflictResult.DetectionSummary,
		RequestedEffectiveDate:   time.Now().AddDate(0, 0, 1),
		Rationale:                "系统自动生成的豁免申请，基于冲突检测结果",
		ClientRepresentativeName: s.getClientRepresentativeName(ctx, enhancedCase.ClientProfileIDs),
		RequestingLawyerName:    s.getLawyerName(ctx, enhancedCase.AssignedBy),
		Status:                   "DRAFT",
		CreatedBy:                *enhancedCase.AssignedBy,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	// 创建豁免申请
	err := s.enhancedConflictRepo.CreateWaiverApplication(ctx, waiverApp)
	if err != nil {
		return nil, fmt.Errorf("创建豁免申请失败: %w", err)
	}

	return waiverApp, nil
}

// buildTeamAssignment 构建团队分配信息
func (s *EnhancedCaseService) buildTeamAssignment(members []TeamMemberInfo) models.TeamAssignment {
	assignment := models.TeamAssignment{
		AllocationRules:     make(map[string]interface{}),
		WorkloadDistribution: make(map[string]int),
	}

	for _, member := range members {
		switch member.Role {
		case "primary_lawyer":
			assignment.PrimaryLawyer = member.User.Username
		case "assisting_lawyer":
			assignment.AssistingLawyers = append(assignment.AssistingLawyers, member.User.Username)
		default:
			assignment.TeamMembers = append(assignment.TeamMembers, models.TeamMember{
				UserID:   fmt.Sprintf("%d", member.User.ID),
				Name:     member.User.Name,
				Role:     member.Role,
				Capacity: member.Capacity,
				Assigned: true,
			})
		}

		// 记录工作负载分配
		assignment.WorkloadDistribution[member.User.Username] = member.Capacity
	}

	return assignment
}

// needsEthicalScreen 判断是否需要建立信息屏障
func (s *EnhancedCaseService) needsEthicalScreen(enhancedCase *models.EnhancedCase, members []TeamMemberInfo) bool {
	// 如果有高风险冲突，需要建立信息屏障
	if enhancedCase.RiskLevel == models.RiskLevelHigh || enhancedCase.RiskLevel == models.RiskLevelCritical {
		return true
	}

	// 如果有团队成员需要信息屏障
	for _, member := range members {
		if member.Screening {
			return true
		}
	}

	// 如果有多个客户，需要考虑信息屏障
	if len(enhancedCase.ClientProfileIDs) > 1 {
		return true
	}

	return false
}

// createEthicalScreen 创建信息屏障
func (s *EnhancedCaseService) createEthicalScreen(ctx context.Context, enhancedCase *models.EnhancedCase, members []TeamMemberInfo, assignedBy uint) error {
	// 构建被屏障的律师列表
	var screenedLawyers []models.ScreenedLawyer
	for _, member := range members {
		if member.Screening || member.Role == "assisting_lawyer" {
			screenedLawyer := models.ScreenedLawyer{
				UserID:     fmt.Sprintf("%d", member.User.ID),
				Name:       member.User.Name,
				Department: member.User.Department,
				ScreenType: "INFORMATION_BARRIER",
				StartDate:  time.Now(),
				Reason:     "基于冲突检测结果自动建立信息屏障",
			}
			screenedLawyers = append(screenedLawyers, screenedLawyer)
		}
	}

	// 创建信息屏障记录
	ethicalScreen := &models.CaseEthicalScreen{
		CaseID:        enhancedCase.ID,
		ScreenType:    models.EthicalScreenTypeInformationBarrier,
		ScreenedLawyers: screenedLawyers,
		RestrictedInformation: models.RestrictedInformation{
			Categories:        []string{"案件详情", "客户信息", "策略文档"},
			DocumentTypes:     []string{"法律意见", "证据材料", "通信记录"},
			Communication:     []string{"邮件", "会议记录", "电话记录"},
			AccessLevel:       "RESTRICTED",
			ApprovalRequired:  true,
		},
		AccessPermissions: models.AccessPermissions{
			ReadAccess:   []string{},
			WriteAccess:  []string{},
			DeleteAccess: []string{},
			ShareAccess:  []string{},
			ApprovalMatrix: map[string][]string{
				"read":  {},
				"write": {},
				"share": {},
			},
		},
		MonitoringPlan: models.MonitoringPlan{
			Frequency:      "WEEKLY",
			Methods:        []string{"系统监控", "人工审查"},
			Triggers:       []string{"访问尝试", "信息共享", "异常行为"},
			ReportTemplate: "ethical_screen_report",
			NotifyPeople:   []string{"合规部门", "管理合伙人"},
		},
		EffectiveDate:  time.Now(),
		Status:         "ACTIVE",
		EstablishedBy:  fmt.Sprintf("%d", assignedBy),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存信息屏障
	return s.enhancedConflictRepo.CreateEthicalScreen(ctx, ethicalScreen)
}

// convertToEnhancedResponse 转换为增强响应格式
func (s *EnhancedCaseService) convertToEnhancedResponse(enhancedCase *models.EnhancedCase, clientProfiles []*models.ClientProfile, teamMembers []TeamMemberInfo) *EnhancedCaseResponse {
	response := &EnhancedCaseResponse{
		ID:                      enhancedCase.ID,
		Title:                   enhancedCase.Title,
		Description:             enhancedCase.Description,
		CaseType:                enhancedCase.CaseType,
		Priority:                enhancedCase.Priority,
		Status:                  enhancedCase.Status,
		StartDate:               enhancedCase.StartDate,
		EndDate:                 enhancedCase.EndDate,
		CreatedAt:               enhancedCase.CreatedAt,
		UpdatedAt:               enhancedCase.UpdatedAt,
		ClientProfileIDs:        []string(enhancedCase.ClientProfileIDs),
		PracticeArea:            enhancedCase.PracticeArea,
		EstimatedDuration:       enhancedCase.EstimatedDuration,
		BillingMethod:           enhancedCase.BillingMethod,
		CreatedViaConflictCheck: enhancedCase.CreatedViaConflictCheck,
		RiskLevel:               string(enhancedCase.RiskLevel),
	}

	// 转换客户信息
	response.ClientProfiles = s.convertToClientInfo(clientProfiles)

	// 转换团队信息
	response.TeamAssignment = s.convertToTeamAssignmentInfo(teamMembers)

	// 转换冲突检测信息
	if enhancedCase.ConflictDetectionResult != nil {
		response.ConflictDetection = s.convertToConflictDetectionInfo(enhancedCase.ConflictDetectionStatus, enhancedCase.ConflictDetectionResult)
	} else {
		response.ConflictDetection = ConflictDetectionInfo{
			Status:   string(enhancedCase.ConflictDetectionStatus),
			NextSteps: []string{"等待冲突检测"},
		}
	}

	// 转换豁免信息
	if enhancedCase.WaiverApplicationID != nil {
		response.WaiverInfo = WaiverInfo{
			ApplicationID: enhancedCase.WaiverApplicationID,
			Status:        string(enhancedCase.WaiverStatus),
		}
	} else {
		response.WaiverInfo = WaiverInfo{
			Status: "NONE",
		}
	}

	return response
}

// 辅助方法
func (s *EnhancedCaseService) generateCheckNumber() string {
	return fmt.Sprintf("CC-%s-%06d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
}

func (s *EnhancedCaseService) generateWaiverNumber() string {
	return fmt.Sprintf("WA-%s-%06d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
}

func (s *EnhancedCaseService) getDepartmentByLawyerID(ctx context.Context, lawyerID *string) string {
	// 实现获取律师部门的逻辑
	return "公司法务部" // 占位符
}

func (s *EnhancedCaseService) getClientRepresentativeName(ctx context.Context, clientIDs models.ClientProfileIDs) string {
	// 实现获取客户代表姓名的逻辑
	return "客户代表" // 占位符
}

func (s *EnhancedCaseService) getLawyerName(ctx context.Context, lawyerID *string) string {
	// 实现获取律师姓名的逻辑
	return "律师姓名" // 占位符
}

// 其他转换方法...
func (s *EnhancedCaseService) convertToClientInfo(profiles []*models.ClientProfile) []EnhancedClientInfo {
	var result []EnhancedClientInfo
	for _, profile := range profiles {
		info := EnhancedClientInfo{
			ClientProfileID: profile.ID,
			ClientNumber:    profile.ClientNumber,
			ClientName:      profile.ClientName,
			ClientType:      profile.ClientType,
			ClientCategory:  profile.ClientCategory,
			Role:            "PRIMARY",
		}
		if profile.PrimaryContactName != "" {
			info.PrimaryContact = &ContactInfo{
				Name:  profile.PrimaryContactName,
				Phone: profile.PrimaryContactPhone,
				Email: profile.PrimaryContactEmail,
			}
		}
		result = append(result, info)
	}
	return result
}

func (s *EnhancedCaseService) convertToTeamAssignmentInfo(members []TeamMemberInfo) TeamAssignmentInfo {
	assignment := TeamAssignmentInfo{
		TeamMembers: make([]EnhancedTeamMember, 0),
	}

	for _, member := range members {
		userInfo := UserInfo{
			ID:         member.User.ID,
			Username:   member.User.Username,
			Name:       member.User.Name,
			Email:      member.User.Email,
			Department: member.User.Department,
			Seniority:  member.User.Seniority,
		}

		switch member.Role {
		case "primary_lawyer":
			assignment.PrimaryLawyer = userInfo
		case "assisting_lawyer":
			assignment.AssistingLawyers = append(assignment.AssistingLawyers, userInfo)
		default:
			enhancedMember := EnhancedTeamMember{
				UserID:    member.User.ID,
				Username:  member.User.Username,
				Name:      member.User.Name,
				Role:      member.Role,
				Capacity:  member.Capacity,
				Screening: member.Screening,
			}
			assignment.TeamMembers = append(assignment.TeamMembers, enhancedMember)
		}
	}

	return assignment
}

func (s *EnhancedCaseService) convertToConflictDetectionInfo(status models.ConflictDetectionStatus, result *models.ConflictDetectionResult) ConflictDetectionInfo {
	return ConflictDetectionInfo{
		Status:              string(status),
		TotalConflicts:      result.TotalConflicts,
		HighRiskConflicts:   result.HighRiskConflicts,
		MediumRiskConflicts: result.MediumRiskConflicts,
		LowRiskConflicts:    result.LowRiskConflicts,
		WaiverRequired:      result.WaiverRequired,
		WaiverPossible:      result.WaiverPossible,
		DetectionSummary:    result.DetectionSummary,
		NextSteps:           result.NextSteps,
		Conflicts:           result.DetectedConflicts,
	}
}

func (s *EnhancedCaseService) convertToConflictItems(conflicts []interface{}) []models.ConflictItem {
	// 实现冲突项转换逻辑
	return []models.ConflictItem{} // 占位符
}