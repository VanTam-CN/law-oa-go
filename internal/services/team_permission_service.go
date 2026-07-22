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

// TeamPermissionService 团队权限服务
type TeamPermissionService struct {
	userRepo       repositories.UserRepository
	caseRepo       repositories.CaseRepository
	cacheRepo      repositories.CacheRepository
	subjectRecheck *SubjectRecheckService
}

// NewTeamPermissionService 创建团队权限服务
func NewTeamPermissionService(userRepo repositories.UserRepository, caseRepo repositories.CaseRepository, cacheRepo repositories.CacheRepository) *TeamPermissionService {
	return &TeamPermissionService{
		userRepo:  userRepo,
		caseRepo:  caseRepo,
		cacheRepo: cacheRepo,
	}
}

// SetSubjectRecheckService prevents team changes from bypassing the case
// subject version workflow once a case is formally in progress.
func (s *TeamPermissionService) SetSubjectRecheckService(service *SubjectRecheckService) {
	s.subjectRecheck = service
}

// TeamAssignmentRequest 团队分配请求
type TeamAssignmentRequest struct {
	CaseID            uint                `json:"case_id" binding:"required"`
	LawyerID          uint                `json:"lawyer_id" binding:"required"`  // 主办律师
	AssistingLawyerID *uint               `json:"assisting_lawyer_id,omitempty"` // 协办律师
	TeamMembers       []TeamMemberRequest `json:"team_members,omitempty"`        // 其他团队成员
	BillingMethod     string              `json:"billing_method" binding:"required"`
	IsMajorRisk       bool                `json:"is_major_risk"`
	AssignedBy        uint                `json:"assigned_by"` // 分配者ID
}

// TeamPermissionMemberRequest 团队成员请求
type TeamPermissionMemberRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Role     string `json:"role" binding:"required"` // paralegal, assistant, intern
	Capacity int    `json:"capacity,omitempty"`      // 工作容量百分比
}

// TeamAssignmentResponse 团队分配响应
type TeamAssignmentResponse struct {
	CaseID          uint                 `json:"case_id"`
	LeadLawyer      *UserResponse        `json:"lead_lawyer"`
	AssistingLawyer *UserResponse        `json:"assisting_lawyer,omitempty"`
	TeamMembers     []TeamMemberResponse `json:"team_members"`
	BillingMethod   string               `json:"billing_method"`
	IsMajorRisk     bool                 `json:"is_major_risk"`
	AssignedAt      time.Time            `json:"assigned_at"`
	AssignedBy      *UserResponse        `json:"assigned_by"`
	Permissions     map[string]bool      `json:"permissions"` // 当前用户权限
}

// TeamMemberResponse 团队成员响应
type TeamMemberResponse struct {
	UserID     uint      `json:"user_id"`
	UserName   string    `json:"user_name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	Department string    `json:"department"`
	JoinedAt   time.Time `json:"joined_at"`
	Capacity   int       `json:"capacity"`
	IsActive   bool      `json:"is_active"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Department string `json:"department"`
	Seniority  string `json:"seniority"`
}

// TeamPermissionCheck 团队权限检查请求
type TeamPermissionCheck struct {
	UserID     uint                   `json:"user_id"`
	CaseID     uint                   `json:"case_id"`
	Action     string                 `json:"action"` // assign, remove, update, view
	TargetRole string                 `json:"target_role,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// 法律行业权限矩阵
var legalPermissionMatrix = map[string]map[string][]string{
	"admin": {
		"case_team": {"assign", "remove", "update", "view", "manage_billing", "approve_major_risk"},
	},
	"partner": {
		"case_team": {"assign", "remove", "update", "view", "manage_billing", "approve_major_risk"},
	},
	"senior_lawyer": {
		"case_team": {"assign", "update", "view", "manage_billing"},
	},
	"associate": {
		"case_team": {"view", "update_assigned"},
	},
	"lawyer": {
		"case_team": {"view", "update_assigned"},
	},
	"paralegal": {
		"case_team": {"view_assigned", "update_basic"},
	},
	"assistant": {
		"case_team": {"view_assigned"},
	},
}

// CheckTeamPermission 检查团队权限
func (s *TeamPermissionService) CheckTeamPermission(ctx context.Context, check *TeamPermissionCheck) (bool, error) {
	// 1. 获取用户信息
	user, err := s.userRepo.FindByID(ctx, check.UserID)
	if err != nil {
		return false, fmt.Errorf("获取用户信息失败: %w", err)
	}
	if user == nil {
		return false, errors.New("用户不存在")
	}
	if IsTechnicalAdminRole(user.Role) {
		// Account administration is not a professional appointment. Do not let
		// a platform admin use this service to infer or mutate matter teams.
		return false, nil
	}

	// 2. 获取案件信息
	caseInfo, err := s.caseRepo.FindByID(ctx, check.CaseID)
	if err != nil {
		return false, fmt.Errorf("获取案件信息失败: %w", err)
	}
	if caseInfo == nil {
		return false, errors.New("案件不存在")
	}

	// 3. 检查基础角色权限
	if !s.hasBasicRolePermission(user.Role, check.Action) {
		return false, nil
	}

	// 4. 检查案件特定权限
	return s.checkCaseSpecificPermission(ctx, user, caseInfo, check)
}

// hasBasicRolePermission 检查基础角色权限
func (s *TeamPermissionService) hasBasicRolePermission(userRole, action string) bool {
	permissions, exists := legalPermissionMatrix[userRole]
	if !exists {
		return false
	}

	for _, perm := range permissions["case_team"] {
		if perm == action {
			return true
		}
	}
	return false
}

// checkCaseSpecificPermission 检查案件特定权限
func (s *TeamPermissionService) checkCaseSpecificPermission(ctx context.Context, user *models.User, caseInfo *models.Case, check *TeamPermissionCheck) (bool, error) {
	// 1. 检查是否是案件主办律师
	if caseInfo.LawyerID == check.UserID {
		return true, nil
	}

	// 2. 检查是否是案件团队成员
	isTeamMember, memberRole, err := s.isCaseTeamMember(ctx, check.UserID, check.CaseID)
	if err != nil {
		return false, err
	}
	if isTeamMember {
		hasPermission := s.checkTeamMemberPermission(memberRole, check.Action)
		return hasPermission, nil
	}

	// 3. 检查特殊权限规则
	return s.checkSpecialPermissionRules(user, caseInfo, check)
}

// isCaseTeamMember 检查是否是案件团队成员
func (s *TeamPermissionService) isCaseTeamMember(ctx context.Context, userID, caseID uint) (bool, string, error) {
	// 这里应该查询案件团队表，暂时使用简化逻辑
	// TODO: 实现完整的案件团队查询逻辑

	cacheKey := fmt.Sprintf("case_team:%d:%d", caseID, userID)
	cached, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var result struct {
			IsMember bool   `json:"is_member"`
			Role     string `json:"role"`
		}
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result.IsMember, result.Role, nil
		}
	}

	// 模拟查询结果，实际应该从数据库查询
	isMember := false
	role := ""

	// 缓存结果
	result := map[string]interface{}{
		"is_member": isMember,
		"role":      role,
	}
	resultJSON, _ := json.Marshal(result)
	s.cacheRepo.Set(ctx, cacheKey, string(resultJSON), 5*time.Minute)

	return isMember, role, nil
}

// checkTeamMemberPermission 检查团队成员权限
func (s *TeamPermissionService) checkTeamMemberPermission(memberRole, action string) bool {
	switch memberRole {
	case "assisting_lawyer":
		return action == "view" || action == "update_assigned"
	case "paralegal":
		return action == "view_assigned" || action == "update_basic"
	case "assistant":
		return action == "view_assigned"
	default:
		return false
	}
}

// checkSpecialPermissionRules 检查特殊权限规则
func (s *TeamPermissionService) checkSpecialPermissionRules(user *models.User, caseInfo *models.Case, check *TeamPermissionCheck) (bool, error) {
	// 规则1: 高级用户可以查看所有案件团队
	if user.Role == "admin" || user.Role == "partner" {
		return check.Action == "view", nil
	}

	// 规则2: 同部门律师可以查看团队信息
	if user.Role == "senior_lawyer" && check.Action == "view" {
		return true, nil
	}

	// 规则3: 重大风险案件需要高级别权限
	if caseInfo.Priority == "high" && check.Action != "view" {
		return user.Role == "admin" || user.Role == "partner", nil
	}

	return false, nil
}

// AssignTeam 分配团队
func (s *TeamPermissionService) AssignTeam(ctx context.Context, req *TeamAssignmentRequest) (*TeamAssignmentResponse, error) {
	// 1. 权限检查
	hasPermission, err := s.CheckTeamPermission(ctx, &TeamPermissionCheck{
		UserID: req.AssignedBy,
		CaseID: req.CaseID,
		Action: "assign",
	})
	if err != nil {
		return nil, fmt.Errorf("权限检查失败: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("您没有权限分配此案件的团队")
	}
	caseInfo, err := s.caseRepo.FindByID(ctx, req.CaseID)
	if err != nil {
		return nil, fmt.Errorf("获取案件信息失败: %w", err)
	}
	if caseInfo == nil {
		return nil, errors.New("案件不存在")
	}
	if isControlledCaseStatus(caseInfo.Status) {
		if s.subjectRecheck == nil {
			return nil, NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "案件主体版本服务未初始化，已阻止正式案件团队变更")
		}
		return nil, NewSubjectWorkflowError("RECHECK_REQUIRED", "正式案件新增或变更承办团队必须先提交主体重检，复核通过后才可生效")
	}

	// 2. 验证主办律师
	leadLawyer, err := s.userRepo.FindByID(ctx, req.LawyerID)
	if err != nil {
		return nil, fmt.Errorf("获取主办律师信息失败: %w", err)
	}
	if leadLawyer == nil || leadLawyer.Role != "lawyer" {
		return nil, errors.New("主办律师必须是律师角色")
	}

	// 3. 验证协办律师（如果存在）
	var assistingLawyer *models.User
	if req.AssistingLawyerID != nil {
		assistingLawyer, err = s.userRepo.FindByID(ctx, *req.AssistingLawyerID)
		if err != nil {
			return nil, fmt.Errorf("获取协办律师信息失败: %w", err)
		}
		if assistingLawyer == nil || assistingLawyer.Role != "lawyer" {
			return nil, errors.New("协办律师必须是律师角色")
		}
	}

	// 4. 验证其他团队成员
	teamMembers := make([]TeamMemberResponse, 0)
	for _, memberReq := range req.TeamMembers {
		member, err := s.userRepo.FindByID(ctx, memberReq.UserID)
		if err != nil {
			return nil, fmt.Errorf("获取团队成员信息失败: %w", err)
		}
		if member == nil {
			return nil, fmt.Errorf("团队成员 %d 不存在", memberReq.UserID)
		}

		// 验证角色
		if !isValidTeamRole(memberReq.Role) {
			return nil, fmt.Errorf("无效的团队角色: %s", memberReq.Role)
		}

		teamMembers = append(teamMembers, TeamMemberResponse{
			UserID:     member.ID,
			UserName:   member.Name,
			Email:      member.Email,
			Role:       memberReq.Role,
			Department: member.Department,
			JoinedAt:   time.Now(),
			Capacity:   memberReq.Capacity,
			IsActive:   true,
		})
	}

	// 5. 更新案件主办律师
	err = s.caseRepo.UpdateLawyer(ctx, req.CaseID, req.LawyerID)
	if err != nil {
		return nil, fmt.Errorf("更新案件主办律师失败: %w", err)
	}

	// 6. 记录团队分配日志
	err = s.logTeamAssignment(ctx, req)
	if err != nil {
		// 日志记录失败不应该影响主要功能
		fmt.Printf("记录团队分配日志失败: %v\n", err)
	}

	// 7. 清除相关缓存
	s.clearTeamCache(ctx, req.CaseID)

	// 8. 构建响应
	assignedBy, _ := s.userRepo.FindByID(ctx, req.AssignedBy)
	response := &TeamAssignmentResponse{
		CaseID:        req.CaseID,
		LeadLawyer:    s.convertUserToResponse(leadLawyer),
		BillingMethod: req.BillingMethod,
		IsMajorRisk:   req.IsMajorRisk,
		AssignedAt:    time.Now(),
		AssignedBy:    s.convertUserToResponse(assignedBy),
		TeamMembers:   teamMembers,
		Permissions:   s.getCurrentUserPermissions(ctx, req.AssignedBy, req.CaseID),
	}

	if assistingLawyer != nil {
		response.AssistingLawyer = s.convertUserToResponse(assistingLawyer)
	}

	return response, nil
}

// isValidTeamRole 验证团队角色是否有效
func isValidTeamRole(role string) bool {
	validRoles := []string{"paralegal", "assistant", "intern"}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// convertUserToResponse 转换用户信息为响应格式
func (s *TeamPermissionService) convertUserToResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
		Department: user.Department,
		Seniority:  user.Seniority,
	}
}

// getCurrentUserPermissions 获取当前用户权限
func (s *TeamPermissionService) getCurrentUserPermissions(ctx context.Context, userID, caseID uint) map[string]bool {
	permissions := make(map[string]bool)

	// 检查各种权限
	actions := []string{"view", "assign", "remove", "update", "manage_billing", "approve_major_risk"}
	for _, action := range actions {
		hasPerm, _ := s.CheckTeamPermission(ctx, &TeamPermissionCheck{
			UserID: userID,
			CaseID: caseID,
			Action: action,
		})
		permissions[action] = hasPerm
	}

	return permissions
}

// logTeamAssignment 记录团队分配日志
func (s *TeamPermissionService) logTeamAssignment(ctx context.Context, req *TeamAssignmentRequest) error {
	logData := map[string]interface{}{
		"action":         "team_assignment",
		"case_id":        req.CaseID,
		"lawyer_id":      req.LawyerID,
		"assigned_by":    req.AssignedBy,
		"billing_method": req.BillingMethod,
		"is_major_risk":  req.IsMajorRisk,
		"timestamp":      time.Now(),
	}

	logJSON, err := json.Marshal(logData)
	if err != nil {
		return err
	}

	// 记录到审计日志
	cacheKey := fmt.Sprintf("audit:team_assignment:%d:%d", req.CaseID, req.AssignedBy)
	return s.cacheRepo.Set(ctx, cacheKey, string(logJSON), 30*24*time.Hour) // 保存30天
}

// clearTeamCache 清除团队相关缓存
func (s *TeamPermissionService) clearTeamCache(ctx context.Context, caseID uint) {
	patterns := []string{
		fmt.Sprintf("case_team:%d:*", caseID),
		fmt.Sprintf("case_permissions:%d:*", caseID),
	}

	for _, pattern := range patterns {
		s.cacheRepo.DeletePattern(ctx, pattern)
	}
}

// GetTeamAssignment 获取团队分配信息
func (s *TeamPermissionService) GetTeamAssignment(ctx context.Context, userID, caseID uint) (*TeamAssignmentResponse, error) {
	// 权限检查
	hasPermission, err := s.CheckTeamPermission(ctx, &TeamPermissionCheck{
		UserID: userID,
		CaseID: caseID,
		Action: "view",
	})
	if err != nil {
		return nil, fmt.Errorf("权限检查失败: %w", err)
	}
	if !hasPermission {
		return nil, errors.New("您没有权限查看此案件的团队信息")
	}

	// 获取案件信息
	caseInfo, err := s.caseRepo.FindByID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("获取案件信息失败: %w", err)
	}
	if caseInfo == nil {
		return nil, errors.New("案件不存在")
	}

	// 获取主办律师信息
	var leadLawyer *UserResponse
	if caseInfo.LawyerID != 0 {
		lawyer, err := s.userRepo.FindByID(ctx, caseInfo.LawyerID)
		if err == nil && lawyer != nil {
			leadLawyer = s.convertUserToResponse(lawyer)
		}
	}

	// 构建响应
	response := &TeamAssignmentResponse{
		CaseID:      caseID,
		LeadLawyer:  leadLawyer,
		TeamMembers: make([]TeamMemberResponse, 0), // TODO: 从数据库查询实际团队成员
		AssignedAt:  caseInfo.CreatedAt,
		Permissions: s.getCurrentUserPermissions(ctx, userID, caseID),
	}

	return response, nil
}
