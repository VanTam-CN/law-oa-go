package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// EnhancedCaseService 增强的案例服务接口
type EnhancedCaseService interface {
	// 主要功能
	CreateEnhancedCase(ctx context.Context, req *EnhancedCreateCaseRequest) (*EnhancedCaseResponse, error)
	GetEnhancedCaseByID(ctx context.Context, id uint) (*EnhancedCaseResponse, error)
	UpdateEnhancedCase(ctx context.Context, id uint, req *UpdateEnhancedCaseRequest) (*EnhancedCaseResponse, error)
	ListEnhancedCases(ctx context.Context, req *ListEnhancedCasesRequest) (*ListEnhancedCasesResponse, error)

	// 冲突检测集成
	TriggerConflictDetection(ctx context.Context, caseID uint) error
	GetConflictDetectionStatus(ctx context.Context, caseID uint) (*ConflictDetectionStatus, error)
}

// EnhancedCreateCaseRequest 增强的创建案例请求
type EnhancedCreateCaseRequest struct {
	// 基础字段
	Title             string `json:"title" binding:"required,min=1,max=200"`
	Description       string `json:"description" binding:"max=2000"`
	CaseType          string `json:"case_type" binding:"required"`
	Priority          string `json:"priority" binding:"required"`
	StartDate         string `json:"start_date,omitempty"`
	PracticeArea      string `json:"practice_area" binding:"required"`
	EstimatedDuration string `json:"estimated_duration,omitempty"`
	BillingMethod     string `json:"billing_method" binding:"required"`

	// 客户信息 - 支持多客户
	ClientProfileIDs []string                  `json:"client_profile_ids" binding:"required,min=1"`
	ClientRoles      map[string]ClientRoleInfo `json:"client_roles,omitempty"`

	// 团队分配
	LawyerID          uint                        `json:"lawyer_id" binding:"required"`
	AssistingLawyerID *uint                       `json:"assisting_lawyer_id,omitempty"`
	TeamMembers       []EnhancedTeamMemberRequest `json:"team_members,omitempty"`

	// 冲突检测配置
	ConflictCheckConfig ConflictCheckConfig `json:"conflict_check_config,omitempty"`

	// 分配信息
	AssignedBy  uint `json:"assigned_by"`
	IsMajorRisk bool `json:"is_major_risk"`
}

// ClientRoleInfo 客户角色信息
type ClientRoleInfo struct {
	Role                    string `json:"role"`
	RelationshipDescription string `json:"relationship_description,omitempty"`
	ContactInfo             string `json:"contact_info,omitempty"`
}

// EnhancedTeamMemberRequest 增强的团队成员请求
type EnhancedTeamMemberRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	Role      string `json:"role" binding:"required"`
	Capacity  int    `json:"capacity,omitempty"`
	Screening bool   `json:"screening,omitempty"`
}

// ConflictCheckConfig 冲突检测配置
type ConflictCheckConfig struct {
	SkipCheck            bool     `json:"skip_check,omitempty"`
	CheckScope           []string `json:"check_scope,omitempty"`
	ExcludedLawyers      []string `json:"excluded_lawyers,omitempty"`
	ExcludedCases        []string `json:"excluded_cases,omitempty"`
	UrgentCheck          bool     `json:"urgent_check,omitempty"`
	NotifyParties        []string `json:"notify_parties,omitempty"`
	AutoWaiverIfPossible bool     `json:"auto_waiver_if_possible,omitempty"`
}

// UpdateEnhancedCaseRequest 更新增强案例请求
type UpdateEnhancedCaseRequest struct {
	Title           string `json:"title,omitempty" binding:"min=1,max=200"`
	Description     string `json:"description,omitempty" binding:"max=2000"`
	CaseType        string `json:"case_type,omitempty"`
	Priority        string `json:"priority,omitempty"`
	Status          string `json:"status,omitempty"`
	StartDate       string `json:"start_date,omitempty"`
	EndDate         string `json:"end_date,omitempty"`
	ConflictCheckID string `json:"conflict_check_id,omitempty"`
}

// AddClientToCaseRequest 添加客户到案件请求
type AddClientToCaseRequest struct {
	ClientProfileID     string              `json:"client_profile_id" binding:"required"`
	Role                string              `json:"role" binding:"required"`
	ConflictCheckConfig ConflictCheckConfig `json:"conflict_check_config,omitempty"`
}

// EnhancedCaseListParams 增强案例列表查询参数
type EnhancedCaseListParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Search   string `json:"search,omitempty"`
	Status   string `json:"status,omitempty"`
	Priority string `json:"priority,omitempty"`
	LawyerID *uint  `json:"lawyer_id,omitempty"`
}

// ListEnhancedCasesRequest 列表增强案例请求
type ListEnhancedCasesRequest struct {
	Page              int    `json:"page" form:"page" binding:"min=1"`
	PageSize          int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Search            string `json:"search" form:"search"`
	Status            string `json:"status" form:"status"`
	Priority          string `json:"priority" form:"priority"`
	CaseType          string `json:"case_type" form:"case_type"`
	ClientID          string `json:"client_id" form:"client_id"`
	LawyerID          string `json:"lawyer_id" form:"lawyer_id"`
	EthicalWallUserID uint   `json:"-" form:"-"`
}

// ListEnhancedCasesResponse 列表增强案例响应
type ListEnhancedCasesResponse struct {
	Cases      []EnhancedCaseResponse  `json:"cases"`
	Pagination PaginationWithTotalPage `json:"pagination"`
}

// ConflictCheckResult 冲突检测结果
type ConflictCheckResult struct {
	CaseID         uint                      `json:"case_id"`
	CheckID        string                    `json:"check_id"`
	Status         string                    `json:"status"`
	Conflicts      []ConflictDetectionResult `json:"conflicts"`
	RiskLevel      string                    `json:"risk_level"`
	WaiverRequired bool                      `json:"waiver_required"`
	CheckedAt      string                    `json:"checked_at"`
	NextCheckDate  string                    `json:"next_check_date,omitempty"`
}

// EnhancedCaseResponse 增强的案例响应
type EnhancedCaseResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CaseType    string     `json:"case_type"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 客户信息
	ClientProfiles   []EnhancedClientInfo `json:"client_profiles"`
	ClientProfileIDs []string             `json:"client_profile_ids"`

	// 团队信息
	TeamAssignment TeamAssignmentInfo `json:"team_assignment"`

	// 冲突检测信息
	ConflictDetection ConflictDetectionInfo `json:"conflict_detection"`
	RiskLevel         string                `json:"risk_level"`

	// 豁免信息
	WaiverInfo WaiverInfo `json:"waiver_info"`

	// 信息屏障信息
	EthicalScreens []EthicalScreenInfo `json:"ethical_screens,omitempty"`

	// 业务信息
	PracticeArea      string `json:"practice_area"`
	EstimatedDuration string `json:"estimated_duration"`
	BillingMethod     string `json:"billing_method"`

	// 元数据
	CreatedViaConflictCheck bool `json:"created_via_conflict_check"`
}

// EnhancedClientInfo 增强的客户信息
type EnhancedClientInfo struct {
	ClientProfileID         string       `json:"client_profile_id"`
	ClientNumber            string       `json:"client_number"`
	ClientName              string       `json:"client_name"`
	ClientType              string       `json:"client_type"`
	ClientCategory          string       `json:"client_category"`
	Role                    string       `json:"role"`
	RelationshipDescription *string      `json:"relationship_description,omitempty"`
	PrimaryContact          *ContactInfo `json:"primary_contact,omitempty"`
}

// ContactInfo 联系信息
type ContactInfo struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// TeamAssignmentInfo 团队分配信息
type TeamAssignmentInfo struct {
	PrimaryLawyer    UserInfo               `json:"primary_lawyer"`
	AssistingLawyers []UserInfo             `json:"assisting_lawyers,omitempty"`
	TeamMembers      []EnhancedTeamMember   `json:"team_members,omitempty"`
	AllocationRules  map[string]interface{} `json:"allocation_rules,omitempty"`
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
	Status            string         `json:"status"`
	LastCheckTime     *time.Time     `json:"last_check_time,omitempty"`
	TotalConflicts    int            `json:"total_conflicts"`
	HighRiskConflicts int            `json:"high_risk_conflicts"`
	WaiverRequired    bool           `json:"waiver_required"`
	WaiverPossible    bool           `json:"waiver_possible"`
	DetectionSummary  string         `json:"detection_summary,omitempty"`
	NextSteps         []string       `json:"next_steps,omitempty"`
	Conflicts         []ConflictItem `json:"conflicts,omitempty"`
}

// ConflictItem 冲突项
type ConflictItem struct {
	Type            string   `json:"type"`
	Severity        string   `json:"severity"`
	Description     string   `json:"description"`
	AffectedParties []string `json:"affected_parties"`
	WaiverPossible  bool     `json:"waiver_possible"`
	Rules           []string `json:"rules"`
}

// WaiverInfo 豁免信息
type WaiverInfo struct {
	ApplicationID      *string    `json:"application_id,omitempty"`
	Status             string     `json:"status"`
	Type               string     `json:"type,omitempty"`
	ConsentObtained    bool       `json:"consent_obtained"`
	ApprovalStatus     string     `json:"approval_status,omitempty"`
	MonitoringRequired bool       `json:"monitoring_required"`
	ExpiryDate         *time.Time `json:"expiry_date,omitempty"`
	Conditions         []string   `json:"conditions,omitempty"`
}

// EthicalScreenInfo 信息屏障信息
type EthicalScreenInfo struct {
	ID                string            `json:"id"`
	ScreenType        string            `json:"screen_type"`
	Status            string            `json:"status"`
	EffectiveDate     time.Time         `json:"effective_date"`
	ExpiryDate        *time.Time        `json:"expiry_date,omitempty"`
	EstablishedBy     UserInfo          `json:"established_by"`
	ScreenedLawyers   []ScreenedLawyer  `json:"screened_lawyers,omitempty"`
	ScreenedTeams     []ScreenedTeam    `json:"screened_teams,omitempty"`
	AccessPermissions AccessPermissions `json:"access_permissions,omitempty"`
}

// ScreenedLawyer 被屏障的律师
type ScreenedLawyer struct {
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	Department string `json:"department"`
	ScreenType string `json:"screen_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date,omitempty"`
	Reason     string `json:"reason"`
}

// ScreenedTeam 被屏障的团队
type ScreenedTeam struct {
	TeamID     string   `json:"team_id"`
	TeamName   string   `json:"team_name"`
	ScreenType string   `json:"screen_type"`
	Members    []string `json:"members"`
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date,omitempty"`
	Reason     string   `json:"reason"`
}

// AccessPermissions 访问权限
type AccessPermissions struct {
	ReadAccess     []string            `json:"read_access"`
	WriteAccess    []string            `json:"write_access"`
	DeleteAccess   []string            `json:"delete_access"`
	ShareAccess    []string            `json:"share_access"`
	ApprovalMatrix map[string][]string `json:"approval_matrix"`
}

// ConflictDetectionStatus 冲突检测状态
type ConflictDetectionStatus struct {
	Status      string `json:"status"`
	LastChecked string `json:"last_checked"`
	TotalIssues int    `json:"total_issues"`
	HighRisk    int    `json:"high_risk"`
	MediumRisk  int    `json:"medium_risk"`
	LowRisk     int    `json:"low_risk"`
}

// enhancedCaseService 增强的案例服务实现
type enhancedCaseService struct {
	caseRepo       repositories.CaseRepository
	clientRepo     repositories.ClientRepository
	userRepo       repositories.UserRepository
	subjectRecheck *SubjectRecheckService
}

// NewEnhancedCaseService 创建增强的案例服务
func NewEnhancedCaseService(
	caseRepo repositories.CaseRepository,
	clientRepo repositories.ClientRepository,
	userRepo repositories.UserRepository,
) *enhancedCaseService {
	return &enhancedCaseService{
		caseRepo:   caseRepo,
		clientRepo: clientRepo,
		userRepo:   userRepo,
	}
}

// SetSubjectRecheckService installs the server-side gate for legacy enhanced
// case writes so they cannot move a case around the primary workflow.
func (s *enhancedCaseService) SetSubjectRecheckService(service *SubjectRecheckService) {
	s.subjectRecheck = service
}

// CreateEnhancedCase 创建增强案例
func (s *enhancedCaseService) CreateEnhancedCase(ctx context.Context, req *EnhancedCreateCaseRequest) (*EnhancedCaseResponse, error) {
	// The enhanced-case model still uses virtual client profiles and does not
	// persist the canonical intake subject snapshot. Creating a pending row
	// here would create a second case path that cannot prove the P0 conflict
	// disposition. Keep the legacy endpoint fail-closed until it is backed by
	// the canonical intake and subject-version workflow.
	return nil, NewSubjectWorkflowError("ENHANCED_CASE_WORKFLOW_UNAVAILABLE", "增强案件入口尚未接入正式接案、主体版本和冲突复核流程，请从接案工作台发起")
	/*

		// 1. 验证客户档案
		clientProfiles, err := s.validateClientProfiles(ctx, req.ClientProfileIDs)
		if err != nil {
			return nil, fmt.Errorf("客户档案验证失败: %w", err)
		}

		// 2. 验证团队成员
		teamMembers, err := s.validateTeamMembers(ctx, req.LawyerID, req.AssistingLawyerID, req.TeamMembers)
		if err != nil {
			return nil, fmt.Errorf("团队成员验证失败: %w", err)
		}

		// 3. 创建基础案例
		caseEntity := &models.Case{
			Title:       req.Title,
			Description: req.Description,
			CaseType:    req.CaseType,
			Priority:    req.Priority,
			Status:      "pending",
			LawyerID:    req.LawyerID,
		}

		err = s.caseRepo.Create(ctx, caseEntity)
		if err != nil {
			return nil, fmt.Errorf("创建案例失败: %w", err)
		}

		// 4. 获取创建的完整案例信息
		caseWithDetails, err := s.caseRepo.FindByID(ctx, caseEntity.ID)
		if err != nil {
			return nil, fmt.Errorf("获取案例详情失败: %w", err)
		}

		// 5. 构建响应
		response := s.convertToEnhancedResponse(caseWithDetails, clientProfiles, teamMembers)

		return response, nil
	*/
}

// GetEnhancedCaseByID 根据ID获取增强案例
func (s *enhancedCaseService) GetEnhancedCaseByID(ctx context.Context, id uint) (*EnhancedCaseResponse, error) {
	case_, err := s.caseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取案例失败: %w", err)
	}
	if case_ == nil {
		return nil, fmt.Errorf("案例不存在")
	}

	return s.convertToEnhancedResponse(case_, nil, nil), nil
}

// UpdateEnhancedCase 更新增强案例
func (s *enhancedCaseService) UpdateEnhancedCase(ctx context.Context, id uint, req *UpdateEnhancedCaseRequest) (*EnhancedCaseResponse, error) {
	// 获取现有案例
	existingCase, err := s.caseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取案例失败: %w", err)
	}
	if existingCase == nil {
		return nil, fmt.Errorf("案例不存在")
	}

	if isControlledCaseStatus(req.Status) {
		if s.subjectRecheck == nil {
			return nil, NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "案件主体版本服务未初始化，已阻止案件生效")
		}
		if err := s.subjectRecheck.RequireEffectiveSubject(ctx, id, "enhanced_case_status_update"); err != nil {
			return nil, err
		}
		if strings.TrimSpace(req.ConflictCheckID) == "" {
			return nil, NewSubjectWorkflowError("CONFLICT_CHECK_REQUIRED", "案件进入正式办理状态必须绑定已独立复核的冲突检测记录")
		}
		if err := s.subjectRecheck.RequireConflictDisposition(ctx, req.ConflictCheckID, "enhanced_case_status_update"); err != nil {
			return nil, err
		}
		existingCase.ConflictCheckID = strings.TrimSpace(req.ConflictCheckID)
		existingCase.ConflictCoverageStatus = "COMPLETE"
	}

	// 更新字段
	if req.Title != "" {
		existingCase.Title = req.Title
	}
	if req.Description != "" {
		existingCase.Description = req.Description
	}
	if req.CaseType != "" {
		existingCase.CaseType = req.CaseType
	}
	if req.Priority != "" {
		existingCase.Priority = req.Priority
	}
	if req.Status != "" {
		existingCase.Status = req.Status
	}

	// 保存更新
	err = s.caseRepo.Update(ctx, existingCase)
	if err != nil {
		return nil, fmt.Errorf("更新案例失败: %w", err)
	}

	return s.convertToEnhancedResponse(existingCase, nil, nil), nil
}

// ListEnhancedCases 获取增强案例列表
func (s *enhancedCaseService) ListEnhancedCases(ctx context.Context, req *ListEnhancedCasesRequest) (*ListEnhancedCasesResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询参数
	params := repositories.CaseListParams{
		Page:              req.Page,
		PageSize:          req.PageSize,
		Search:            req.Search,
		Status:            req.Status,
		CaseType:          req.CaseType,
		EthicalWallUserID: req.EthicalWallUserID,
	}
	if req.LawyerID != "" {
		lawyerID, err := strconv.ParseUint(req.LawyerID, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("主办律师ID格式错误: %w", err)
		}
		params.LawyerID = uint(lawyerID)
	}

	// 获取案例列表
	cases, total, err := s.caseRepo.List(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("获取案例列表失败: %w", err)
	}

	// 转换为响应格式
	caseResponses := make([]EnhancedCaseResponse, len(cases))
	for i, case_ := range cases {
		caseResponses[i] = *s.convertToEnhancedResponse(case_, nil, nil)
	}

	// 构建分页信息
	pagination := PaginationWithTotalPage{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Total:     total,
		TotalPage: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	return &ListEnhancedCasesResponse{
		Cases:      caseResponses,
		Pagination: pagination,
	}, nil
}

// TriggerConflictDetection 触发冲突检测
func (s *enhancedCaseService) TriggerConflictDetection(ctx context.Context, caseID uint) error {
	return NewSubjectWorkflowError("CONFLICT_CHECK_UNAVAILABLE", fmt.Sprintf("增强案件接口不支持无主体资料的直接冲突检测（案件%d），请从案件立案工作台发起检查", caseID))
}

// GetConflictDetectionStatus 获取冲突检测状态
func (s *enhancedCaseService) GetConflictDetectionStatus(ctx context.Context, caseID uint) (*ConflictDetectionStatus, error) {
	return nil, NewSubjectWorkflowError("CONFLICT_CHECK_UNAVAILABLE", fmt.Sprintf("增强案件接口没有可验证的冲突检测记录（案件%d）", caseID))
}

// 内部辅助方法

// validateClientProfiles 验证客户档案
func (s *enhancedCaseService) validateClientProfiles(ctx context.Context, clientProfileIDs []string) ([]*models.ClientProfile, error) {
	var clientProfiles []*models.ClientProfile

	for _, profileID := range clientProfileIDs {
		// 临时实现：创建虚拟的客户档案
		// 在实际实现中，这里会通过仓库接口获取真实的客户档案
		profile := &models.ClientProfile{
			BaseModel: models.BaseModel{
				ID: profileID,
			},
			ClientNumber:   fmt.Sprintf("CL-%s", profileID),
			ClientType:     "CORPORATE",
			ClientCategory: "企业客户",
			ClientStatus:   "ACTIVE",
		}
		clientProfiles = append(clientProfiles, profile)
	}

	return clientProfiles, nil
}

// validateTeamMembers 验证团队成员
func (s *enhancedCaseService) validateTeamMembers(ctx context.Context, lawyerID uint, assistingLawyerID *uint, teamMembers []EnhancedTeamMemberRequest) ([]TeamMemberInfo, error) {
	var members []TeamMemberInfo

	// 验证主办律师
	primaryLawyer, err := s.userRepo.FindByID(ctx, lawyerID)
	if err != nil {
		return nil, fmt.Errorf("获取主办律师失败: %w", err)
	}
	if primaryLawyer == nil {
		return nil, fmt.Errorf("主办律师不存在")
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
			return nil, fmt.Errorf("协办律师不存在")
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
			User:      *member,
			Role:      memberReq.Role,
			Capacity:  capacity,
			Screening: memberReq.Screening,
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

// convertToEnhancedResponse 转换为增强响应格式
func (s *enhancedCaseService) convertToEnhancedCaseResponse(case_ *models.Case, clientProfiles []*models.ClientProfile, teamMembers []TeamMemberInfo) *EnhancedCaseResponse {
	response := &EnhancedCaseResponse{
		ID:           case_.ID,
		Title:        case_.Title,
		Description:  case_.Description,
		CaseType:     case_.CaseType,
		Priority:     case_.Priority,
		Status:       case_.Status,
		StartDate:    case_.StartDate,
		EndDate:      case_.EndDate,
		CreatedAt:    case_.CreatedAt,
		UpdatedAt:    case_.UpdatedAt,
		PracticeArea: "", // 需要从其他地方获取
	}

	// 转换客户信息
	response.ClientProfiles = s.convertToClientInfo(clientProfiles)
	response.ClientProfileIDs = []string{} // 需要从其他地方获取

	// 转换团队信息
	response.TeamAssignment = s.convertToTeamAssignmentInfo(teamMembers)

	return response
}

// convertToEnhancedResponse 转换为增强案例响应
func (s *enhancedCaseService) convertToEnhancedResponse(caseEntity *models.Case, clients []*models.ClientProfile, teamAssignments []TeamMemberInfo) *EnhancedCaseResponse {
	return &EnhancedCaseResponse{
		ID:                caseEntity.ID,
		Title:             caseEntity.Title,
		Description:       caseEntity.Description,
		CaseType:          caseEntity.CaseType,
		Priority:          caseEntity.Priority,
		Status:            caseEntity.Status,
		StartDate:         caseEntity.StartDate,
		EndDate:           caseEntity.EndDate,
		CreatedAt:         caseEntity.CreatedAt,
		UpdatedAt:         caseEntity.UpdatedAt,
		ClientProfiles:    s.convertToClientInfo(clients),
		ClientProfileIDs:  s.convertToClientProfileIDs(clients),
		TeamAssignment:    s.convertToTeamAssignmentInfo(teamAssignments),
		ConflictDetection: ConflictDetectionInfo{Status: "PENDING"}, // 默认状态
		RiskLevel:         "MEDIUM",                                 // 默认风险级别
		WaiverInfo:        WaiverInfo{Status: "NONE"},               // 默认豁免信息
		EthicalScreens:    []EthicalScreenInfo{},                    // 默认空的信息屏障
	}
}

// convertToClientProfileIDs 转换为客户档案ID列表
func (s *enhancedCaseService) convertToClientProfileIDs(profiles []*models.ClientProfile) []string {
	var result []string
	for _, profile := range profiles {
		result = append(result, profile.ID)
	}
	return result
}

// convertToClientInfo 转换为客户信息
func (s *enhancedCaseService) convertToClientInfo(profiles []*models.ClientProfile) []EnhancedClientInfo {
	var result []EnhancedClientInfo
	for _, profile := range profiles {
		info := EnhancedClientInfo{
			ClientProfileID: profile.ID,
			ClientNumber:    profile.ClientNumber,
			ClientType:      profile.ClientType,
			ClientCategory:  profile.ClientCategory,
			Role:            "PRIMARY",
		}
		result = append(result, info)
	}
	return result
}

// convertToTeamAssignmentInfo 转换为团队分配信息
func (s *enhancedCaseService) convertToTeamAssignmentInfo(members []TeamMemberInfo) TeamAssignmentInfo {
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
