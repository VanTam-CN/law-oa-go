package services

import (
	"context"
	"fmt"
	"time"

	"github.com/law-oa-go/document-service/internal/auth"
	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// RBACService RBAC服务接口
type RBACService interface {
	// 角色管理
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error)
	GetRole(ctx context.Context, id uint) (*RoleResponse, error)
	UpdateRole(ctx context.Context, id uint, req *UpdateRoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, id uint) error
	ListRoles(ctx context.Context, filter *RoleFilter) (*RoleListResponse, error)

	// 角色分配
	AssignRole(ctx context.Context, req *AssignRoleRequest) error
	RevokeRole(ctx context.Context, req *RevokeRoleRequest) error
	GetUserRoles(ctx context.Context, userID uint, tenantID string) (*UserRoleListResponse, error)
	GetRoleUsers(ctx context.Context, roleID uint, tenantID string) (*RoleUserListResponse, error)

	// 权限检查
	CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error)
	BulkCheckPermissions(ctx context.Context, reqs []*CheckPermissionRequest) ([]*CheckPermissionResponse, error)
	CheckUserPermissions(ctx context.Context, userID uint, tenantID string) (*UserPermissionsResponse, error)

	// 权限管理
	CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*PermissionResponse, error)
	GetPermission(ctx context.Context, id uint) (*PermissionResponse, error)
	UpdatePermission(ctx context.Context, id uint, req *UpdatePermissionRequest) (*PermissionResponse, error)
	DeletePermission(ctx context.Context, id uint) error
	ListPermissions(ctx context.Context, filter *PermissionFilter) (*PermissionListResponse, error)

	// 角色权限管理
	AddPermissionToRole(ctx context.Context, req *RolePermissionRequest) error
	RemovePermissionFromRole(ctx context.Context, req *RolePermissionRequest) error
	GetRolePermissions(ctx context.Context, roleID uint) (*PermissionListResponse, error)

	// 角色继承
SetRoleInheritance(ctx context.Context, req *RoleInheritanceRequest) error
	RemoveRoleInheritance(ctx context.Context, req *RoleInheritanceRequest) error
	GetRoleInheritance(ctx context.Context, roleID uint) (*RoleInheritanceResponse, error)

	// 角色模板
	CreateRoleTemplate(ctx context.Context, req *CreateRoleTemplateRequest) (*RoleTemplateResponse, error)
	GetRoleTemplates(ctx context.Context, filter *RoleTemplateFilter) (*RoleTemplateListResponse, error)
	CreateRoleFromTemplate(ctx context.Context, req *CreateRoleFromTemplateRequest) (*RoleResponse, error)

	// 角色分析和统计
	GetRoleStatistics(ctx context.Context, tenantID string) (*RoleStatistics, error)
	AnalyzeRoleUsage(ctx context.Context, tenantID string) (*RoleUsageAnalysis, error)
	GetPermissionUsage(ctx context.Context, tenantID string) (*PermissionUsageAnalysis, error)

	// 批量操作
	BulkAssignRoles(ctx context.Context, req *BulkAssignRolesRequest) (*BulkOperationResult, error)
	BulkRevokeRoles(ctx context.Context, req *BulkRevokeRolesRequest) (*BulkOperationResult, error)
	BulkCreateRoles(ctx context.Context, req *BulkCreateRolesRequest) (*BulkOperationResult, error)
}

// rbacService RBAC服务实现
type rbacService struct {
	rbacEngine     *auth.RBACEngine
	roleRepo       repositories.RoleRepository
	userRepo       repositories.UserRepository
	permissionRepo repositories.PermissionRepository
	auditRepo      repositories.AuditRepository
	logger         *logrus.Logger
}

// NewRBACService 创建RBAC服务
func NewRBACService(
	rbacEngine *auth.RBACEngine,
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	permissionRepo repositories.PermissionRepository,
	auditRepo repositories.AuditRepository,
	logger *logrus.Logger,
) RBACService {
	return &rbacService{
		rbacEngine:     rbacEngine,
		roleRepo:       roleRepo,
		userRepo:       userRepo,
		permissionRepo: permissionRepo,
		auditRepo:      auditRepo,
		logger:         logger,
	}
}

// CreateRole 创建角色
func (s *rbacService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"name":      req.Name,
		"tenant_id": req.TenantID,
		"type":      req.Type,
	}).Info("Creating new role")

	// 构建角色定义
	roleDef := &auth.RoleDefinition{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Level:       req.Level,
		Type:        req.Type,
		Permissions: req.Permissions,
		Constraints: req.Constraints,
		Inherits:    req.Inherits,
		TenantID:    req.TenantID,
		Enabled:     req.Enabled,
		CreatedBy:   req.CreatedBy,
	}

	if err := s.rbacEngine.CreateRole(ctx, roleDef); err != nil {
		s.logger.WithError(err).Error("Failed to create role")
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return s.buildRoleResponse(roleDef), nil
}

// GetRole 获取角色
func (s *rbacService) GetRole(ctx context.Context, id uint) (*RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("role_id", id).Error("Failed to get role")
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	permissions, _ := s.roleRepo.GetPermissions(ctx, role.ID)

	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		Level:       role.Level,
		Type:        role.Type,
		Permissions: permissions,
		TenantID:    role.TenantID,
		Enabled:     role.Enabled,
		CreatedBy:   role.CreatedBy,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

// UpdateRole 更新角色
func (s *rbacService) UpdateRole(ctx context.Context, id uint, req *UpdateRoleRequest) (*RoleResponse, error) {
	s.logger.WithField("role_id", id).Info("Updating role")

	// 构建更新数据
	updates := &auth.RoleDefinition{
		DisplayName: req.DisplayName,
		Description: req.Description,
		Level:       req.Level,
		Enabled:     req.Enabled,
		Permissions: req.Permissions,
	}

	if err := s.rbacEngine.UpdateRole(ctx, id, updates); err != nil {
		s.logger.WithError(err).Error("Failed to update role")
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return s.GetRole(ctx, id)
}

// DeleteRole 删除角色
func (s *rbacService) DeleteRole(ctx context.Context, id uint) error {
	s.logger.WithField("role_id", id).Info("Deleting role")

	if err := s.rbacEngine.DeleteRole(ctx, id); err != nil {
		s.logger.WithError(err).Error("Failed to delete role")
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// ListRoles 列出角色
func (s *rbacService) ListRoles(ctx context.Context, filter *RoleFilter) (*RoleListResponse, error) {
	roles, total, err := s.roleRepo.List(ctx, &repositories.RoleFilter{
		TenantID:    filter.TenantID,
		Name:        filter.Name,
		Type:        filter.Type,
		Enabled:     filter.Enabled,
		CreatorID:   filter.CreatorID,
		CreatedFrom: filter.CreatedFrom,
		CreatedTo:   filter.CreatedTo,
		Pagination:  filter.Pagination,
		SortBy:      filter.SortBy,
		SortOrder:   filter.SortOrder,
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to list roles")
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	var roleResponses []*RoleResponse
	for _, role := range roles {
		permissions, _ := s.roleRepo.GetPermissions(ctx, role.ID)

		roleResponse := &RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			Level:       role.Level,
			Type:        role.Type,
			Permissions: permissions,
			TenantID:    role.TenantID,
			Enabled:     role.Enabled,
			CreatedBy:   role.CreatedBy,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}
		roleResponses = append(roleResponses, roleResponse)
	}

	return &RoleListResponse{
		Roles:    roleResponses,
		Total:    total,
		Page:     filter.Pagination.Page,
		PageSize: filter.Pagination.PageSize,
	}, nil
}

// AssignRole 分配角色
func (s *rbacService) AssignRole(ctx context.Context, req *AssignRoleRequest) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":   req.UserID,
		"role_id":   req.RoleID,
		"tenant_id": req.TenantID,
	}).Info("Assigning role to user")

	return s.rbacEngine.AssignRole(ctx, req.UserID, req.RoleID, req.TenantID, req.Attributes, req.ExpiresAt)
}

// RevokeRole 撤销角色
func (s *rbacService) RevokeRole(ctx context.Context, req *RevokeRoleRequest) error {
	s.logger.WithFields(logrus.Fields{
		"user_id":   req.UserID,
		"role_id":   req.RoleID,
		"tenant_id": req.TenantID,
	}).Info("Revoking role from user")

	return s.rbacEngine.RevokeRole(ctx, req.UserID, req.RoleID, req.TenantID)
}

// GetUserRoles 获取用户角色
func (s *rbacService) GetUserRoles(ctx context.Context, userID uint, tenantID string) (*UserRoleListResponse, error) {
	roleDefs, err := s.rbacEngine.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user roles")
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	var userRoles []*UserRoleResponse
	for _, roleDef := range roleDefs {
		userRole := &UserRoleResponse{
			RoleID:      roleDef.ID,
			RoleName:    roleDef.Name,
			DisplayName: roleDef.DisplayName,
			Type:        roleDef.Type,
			Level:       roleDef.Level,
			TenantID:    roleDef.TenantID,
			Enabled:     roleDef.Enabled,
			CreatedAt:   roleDef.CreatedAt,
		}
		userRoles = append(userRoles, userRole)
	}

	return &UserRoleListResponse{
		UserID: userID,
		Roles:  userRoles,
		Total:  int64(len(userRoles)),
	}, nil
}

// GetRoleUsers 获取角色用户
func (s *rbacService) GetRoleUsers(ctx context.Context, roleID uint, tenantID string) (*RoleUserListResponse, error) {
	users, err := s.userRepo.GetUsersByRole(ctx, roleID, tenantID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get role users")
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}

	var roleUsers []*RoleUserResponse
	for _, user := range users {
		roleUser := &RoleUserResponse{
			UserID:    user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Active:    user.Active,
			CreatedAt: user.CreatedAt,
		}
		roleUsers = append(roleUsers, roleUser)
	}

	return &RoleUserListResponse{
		RoleID: roleID,
		Users:  roleUsers,
		Total:  int64(len(roleUsers)),
	}, nil
}

// CheckPermission 检查权限
func (s *rbacService) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	accessCheck := &auth.AccessCheck{
		UserID:    req.UserID,
		Username:  req.Username,
		TenantID:  req.TenantID,
		Resource:  req.Resource,
		Action:    req.Action,
		Context:   req.Context,
		RequestID: req.RequestID,
		Timestamp: time.Now(),
	}

	result, err := s.rbacEngine.CheckPermission(ctx, accessCheck)
	if err != nil {
		s.logger.WithError(err).Error("Failed to check permission")
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	return &CheckPermissionResponse{
		Allowed:     result.Allowed,
		Reason:      result.Reason,
		Roles:       result.Roles,
		Permissions: result.Permissions,
		Duration:    result.Duration,
		Constraints: result.Constraints,
		Attributes:  result.Attributes,
	}, nil
}

// BulkCheckPermissions 批量检查权限
func (s *rbacService) BulkCheckPermissions(ctx context.Context, reqs []*CheckPermissionRequest) ([]*CheckPermissionResponse, error) {
	var responses []*CheckPermissionResponse
	var errors []error

	for _, req := range reqs {
		response, err := s.CheckPermission(ctx, req)
		if err != nil {
			errors = append(errors, err)
			responses = append(responses, &CheckPermissionResponse{
				Allowed: false,
				Reason:  fmt.Sprintf("Check error: %v", err),
			})
		} else {
			responses = append(responses, response)
		}
	}

	if len(errors) > 0 {
		s.logger.WithField("error_count", len(errors)).Warn("Some permission checks failed")
	}

	return responses, nil
}

// CheckUserPermissions 检查用户权限
func (s *rbacService) CheckUserPermissions(ctx context.Context, userID uint, tenantID string) (*UserPermissionsResponse, error) {
	// 获取用户角色
	roleDefs, err := s.rbacEngine.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 收集所有权限
	var allPermissions []string
	roleDetails := make(map[string]*RolePermissionDetail)

	for _, roleDef := range roleDefs {
		roleDetail := &RolePermissionDetail{
			RoleID:      roleDef.ID,
			RoleName:    roleDef.Name,
			DisplayName: roleDef.DisplayName,
			Type:        roleDef.Type,
			Level:       roleDef.Level,
			Permissions: roleDef.Permissions,
			Enabled:     roleDef.Enabled,
		}
		roleDetails[roleDef.Name] = roleDetail

		for _, perm := range roleDef.Permissions {
			if !s.containsString(allPermissions, perm) {
				allPermissions = append(allPermissions, perm)
			}
		}
	}

	return &UserPermissionsResponse{
		UserID:      userID,
		TenantID:    tenantID,
		Roles:       roleDetails,
		Permissions: allPermissions,
		TotalRoles:  len(roleDefs),
		TotalPerms:  len(allPermissions),
		CheckedAt:   time.Now(),
	}, nil
}

// CreatePermission 创建权限
func (s *rbacService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*PermissionResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"name":     req.Name,
		"resource": req.Resource,
		"action":   req.Action,
	}).Info("Creating new permission")

	permission := &models.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
		Scope:       req.Scope,
		Category:    req.Category,
		System:      req.System,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		s.logger.WithError(err).Error("Failed to create permission")
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return s.buildPermissionResponse(permission), nil
}

// GetPermission 获取权限
func (s *rbacService) GetPermission(ctx context.Context, id uint) (*PermissionResponse, error) {
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return s.buildPermissionResponse(permission), nil
}

// UpdatePermission 更新权限
func (s *rbacService) UpdatePermission(ctx context.Context, id uint, req *UpdatePermissionRequest) (*PermissionResponse, error) {
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	if req.DisplayName != "" {
		permission.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		permission.Description = req.Description
	}
	if req.Scope != "" {
		permission.Scope = req.Scope
	}
	if req.Category != "" {
		permission.Category = req.Category
	}
	if req.Enabled != nil {
		permission.Enabled = *req.Enabled
	}

	permission.UpdatedAt = time.Now()

	if err := s.permissionRepo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return s.buildPermissionResponse(permission), nil
}

// DeletePermission 删除权限
func (s *rbacService) DeletePermission(ctx context.Context, id uint) error {
	if err := s.permissionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// ListPermissions 列出权限
func (s *rbacService) ListPermissions(ctx context.Context, filter *PermissionFilter) (*PermissionListResponse, error) {
	permissions, total, err := s.permissionRepo.List(ctx, &repositories.PermissionFilter{
		Name:        filter.Name,
		Resource:    filter.Resource,
		Action:      filter.Action,
		Scope:       filter.Scope,
		Category:    filter.Category,
		System:      filter.System,
		Enabled:     filter.Enabled,
		Pagination:  filter.Pagination,
		SortBy:      filter.SortBy,
		SortOrder:   filter.SortOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	var permissionResponses []*PermissionResponse
	for _, permission := range permissions {
		permissionResponses = append(permissionResponses, s.buildPermissionResponse(permission))
	}

	return &PermissionListResponse{
		Permissions: permissionResponses,
		Total:       total,
		Page:        filter.Pagination.Page,
		PageSize:    filter.Pagination.PageSize,
	}, nil
}

// AddPermissionToRole 为角色添加权限
func (s *rbacService) AddPermissionToRole(ctx context.Context, req *RolePermissionRequest) error {
	return s.roleRepo.AddPermission(ctx, req.RoleID, req.PermissionName)
}

// RemovePermissionFromRole 从角色移除权限
func (s *rbacService) RemovePermissionFromRole(ctx context.Context, req *RolePermissionRequest) error {
	return s.roleRepo.RemovePermission(ctx, req.RoleID, req.PermissionName)
}

// GetRolePermissions 获取角色权限
func (s *rbacService) GetRolePermissions(ctx context.Context, roleID uint) (*PermissionListResponse, error) {
	permissions, err := s.roleRepo.GetPermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	var permissionResponses []*PermissionResponse
	for _, permName := range permissions {
		permission, err := s.permissionRepo.GetByName(ctx, permName)
		if err == nil {
			permissionResponses = append(permissionResponses, s.buildPermissionResponse(permission))
		}
	}

	return &PermissionListResponse{
		Permissions: permissionResponses,
		Total:       int64(len(permissionResponses)),
	}, nil
}

// 辅助方法
func (s *rbacService) buildRoleResponse(roleDef *auth.RoleDefinition) *RoleResponse {
	return &RoleResponse{
		ID:          roleDef.ID,
		Name:        roleDef.Name,
		DisplayName: roleDef.DisplayName,
		Description: roleDef.Description,
		Level:       roleDef.Level,
		Type:        roleDef.Type,
		Permissions: roleDef.Permissions,
		Constraints: roleDef.Constraints,
		Inherits:    roleDef.Inherits,
		TenantID:    roleDef.TenantID,
		Enabled:     roleDef.Enabled,
		CreatedBy:   roleDef.CreatedBy,
		CreatedAt:   roleDef.CreatedAt,
		UpdatedAt:   roleDef.UpdatedAt,
	}
}

func (s *rbacService) buildPermissionResponse(permission *models.Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		DisplayName: permission.DisplayName,
		Description: permission.Description,
		Resource:    permission.Resource,
		Action:      permission.Action,
		Scope:       permission.Scope,
		Category:    permission.Category,
		System:      permission.System,
		Enabled:     permission.Enabled,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

func (s *rbacService) containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// 占位符方法，需要进一步实现
func (s *rbacService) SetRoleInheritance(ctx context.Context, req *RoleInheritanceRequest) error {
	return fmt.Errorf("not implemented")
}

func (s *rbacService) RemoveRoleInheritance(ctx context.Context, req *RoleInheritanceRequest) error {
	return fmt.Errorf("not implemented")
}

func (s *rbacService) GetRoleInheritance(ctx context.Context, roleID uint) (*RoleInheritanceResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) CreateRoleTemplate(ctx context.Context, req *CreateRoleTemplateRequest) (*RoleTemplateResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) GetRoleTemplates(ctx context.Context, filter *RoleTemplateFilter) (*RoleTemplateListResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) CreateRoleFromTemplate(ctx context.Context, req *CreateRoleFromTemplateRequest) (*RoleResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) GetRoleStatistics(ctx context.Context, tenantID string) (*RoleStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) AnalyzeRoleUsage(ctx context.Context, tenantID string) (*RoleUsageAnalysis, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) GetPermissionUsage(ctx context.Context, tenantID string) (*PermissionUsageAnalysis, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) BulkAssignRoles(ctx context.Context, req *BulkAssignRolesRequest) (*BulkOperationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) BulkRevokeRoles(ctx context.Context, req *BulkRevokeRolesRequest) (*BulkOperationResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *rbacService) BulkCreateRoles(ctx context.Context, req *BulkCreateRolesRequest) (*BulkOperationResult, error) {
	return nil, fmt.Errorf("not implemented")
}