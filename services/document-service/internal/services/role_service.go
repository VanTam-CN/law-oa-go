package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// roleService 角色服务实现
type roleService struct {
	roleRepo  repositories.RoleRepository
	userRepo  repositories.UserRepository
	auditRepo repositories.DocumentAuditRepository
	logger    *logrus.Logger
}

// NewRoleService 创建新的角色服务
func NewRoleService(
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	auditRepo repositories.DocumentAuditRepository,
	logger *logrus.Logger,
) RoleService {
	return &roleService{
		roleRepo:  roleRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// CreateRole 创建角色
func (s *roleService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error) {
	// 验证请求
	if err := s.validateCreateRoleRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 检查角色名称是否已存在
	existingRole, err := s.roleRepo.GetByName(ctx, req.Name, req.TenantID)
	if err == nil && existingRole != nil {
		return nil, fmt.Errorf("role with name %s already exists", req.Name)
	}

	// 创建角色
	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		TenantID:    req.TenantID,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		Action:     "create_role",
		Details:    fmt.Sprintf("Created role: %s", role.Name),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return s.convertToRoleResponse(role), nil
}

// GetRole 获取角色信息
func (s *roleService) GetRole(ctx context.Context, roleID string) (*RoleResponse, error) {
	id, err := s.parseRoleID(roleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return s.convertToRoleResponse(role), nil
}

// UpdateRole 更新角色信息
func (s *roleService) UpdateRole(ctx context.Context, req *UpdateRoleRequest) (*RoleResponse, error) {
	// 解析角色ID
	roleID, err := s.parseRoleID(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	// 获取现有角色
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		// 检查新名称是否与其他角色冲突
		existingRole, err := s.roleRepo.GetByName(ctx, req.Name, role.TenantID)
		if err == nil && existingRole != nil && existingRole.ID != roleID {
			return nil, fmt.Errorf("role with name %s already exists", req.Name)
		}
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	role.UpdatedAt = time.Now()

	// 保存更新
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		Action:     "update_role",
		Details:    fmt.Sprintf("Updated role: %s", role.Name),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return s.convertToRoleResponse(role), nil
}

// DeleteRole 删除角色
func (s *roleService) DeleteRole(ctx context.Context, req *DeleteRoleRequest) error {
	roleID, err := s.parseRoleID(req.RoleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	// 获取角色信息用于审计
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查角色是否被用户使用
	users, err := s.userRepo.GetUsersByRole(ctx, roleID)
	if err == nil && len(users) > 0 {
		return fmt.Errorf("cannot delete role: it is assigned to %d users", len(users))
	}

	// 删除角色
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		Action:     "delete_role",
		Details:    fmt.Sprintf("Deleted role: %s", role.Name),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// ListRoles 列出角色
func (s *roleService) ListRoles(ctx context.Context, filter *RoleFilter) (*RoleListResponse, error) {
	// 设置默认值
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	// 计算偏移量
	offset := (filter.Page - 1) * filter.PageSize

	// 获取角色列表
	roles, total, err := s.roleRepo.List(ctx, repositories.RoleListOptions{
		TenantID:  filter.TenantID,
		Search:    filter.Search,
		Limit:     filter.PageSize,
		Offset:    offset,
		SortBy:    filter.SortBy,
		SortOrder: filter.SortOrder,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	// 转换为响应格式
	responses := make([]*RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = s.convertToRoleResponse(role)
	}

	return &RoleListResponse{
		Roles:    responses,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

// AssignRoleToUser 为用户分配角色
func (s *roleService) AssignRoleToUser(ctx context.Context, req *AssignRoleToUserRequest) error {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	roleID, err := s.parseRoleID(req.RoleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	// 验证用户存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证角色存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查用户是否已拥有该角色
	userRoles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err == nil {
		for _, userRole := range userRoles {
			if userRole.RoleID == roleID {
				return fmt.Errorf("user already has this role")
			}
		}
	}

	// 分配角色
	userRole := &models.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.AssignRole(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "assign_role_to_user",
		Details:    fmt.Sprintf("Assigned role %s to user %s", role.Name, user.Username),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// RemoveRoleFromUser 从用户移除角色
func (s *roleService) RemoveRoleFromUser(ctx context.Context, req *RemoveRoleFromUserRequest) error {
	userID, err := s.parseUserID(req.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	roleID, err := s.parseRoleID(req.RoleID)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	// 验证用户存在
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 验证角色存在
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 移除角色
	if err := s.userRepo.RemoveRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		UserID:     req.UserID,
		Action:     "remove_role_from_user",
		Details:    fmt.Sprintf("Removed role %s from user %s", role.Name, user.Username),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// GetRoleUsers 获取角色用户
func (s *roleService) GetRoleUsers(ctx context.Context, roleID string) ([]*UserResponse, error) {
	id, err := s.parseRoleID(roleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	// 获取角色用户关联
	users, err := s.userRepo.GetUsersByRole(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}

	// 转换为响应格式
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		response := &UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			Status:      user.Status,
			TenantID:    user.TenantID,
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}

		// 拼接全名
		if user.FirstName != "" && user.LastName != "" {
			response.FullName = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
		} else if user.FirstName != "" {
			response.FullName = user.FirstName
		} else if user.LastName != "" {
			response.FullName = user.LastName
		} else {
			response.FullName = user.Username
		}

		responses[i] = response
	}

	return responses, nil
}

// GetRolePermissions 获取角色权限
func (s *roleService) GetRolePermissions(ctx context.Context, roleID string) ([]*PermissionResponse, error) {
	// 这个功能需要扩展权限仓库来支持角色权限查询
	// 这里先返回空数组，实际实现需要根据具体的权限模型来实现
	return []*PermissionResponse{}, nil
}

// BatchAssignRoles 批量分配角色
func (s *roleService) BatchAssignRoles(ctx context.Context, req *BatchAssignRolesRequest) error {
	if len(req.UserIDs) == 0 || len(req.RoleIDs) == 0 {
		return fmt.Errorf("user IDs and role IDs are required")
	}

	// 解析ID
	userIDs := make([]uint, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		id, err := s.parseUserID(userID)
		if err != nil {
			return fmt.Errorf("invalid user ID %s: %w", userID, err)
		}
		userIDs[i] = id
	}

	roleIDs := make([]uint, len(req.RoleIDs))
	for i, roleID := range req.RoleIDs {
		id, err := s.parseRoleID(roleID)
		if err != nil {
			return fmt.Errorf("invalid role ID %s: %w", roleID, err)
		}
		roleIDs[i] = id
	}

	// 批量分配
	for _, userID := range userIDs {
		for _, roleID := range roleIDs {
			userRole := &models.UserRole{
				UserID:    userID,
				RoleID:    roleID,
				CreatedAt: time.Now(),
			}

			if err := s.userRepo.AssignRole(ctx, userRole); err != nil {
				s.logger.WithError(err).WithFields(map[string]interface{}{
					"user_id": userID,
					"role_id": roleID,
				}).Error("Failed to assign role in batch operation")
				continue
			}
		}
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		Action:     "batch_assign_roles",
		Details:    fmt.Sprintf("Batch assigned %d roles to %d users", len(req.RoleIDs), len(req.UserIDs)),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// BatchRemoveRoles 批量移除角色
func (s *roleService) BatchRemoveRoles(ctx context.Context, req *BatchRemoveRolesRequest) error {
	if len(req.UserIDs) == 0 || len(req.RoleIDs) == 0 {
		return fmt.Errorf("user IDs and role IDs are required")
	}

	// 解析ID
	userIDs := make([]uint, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		id, err := s.parseUserID(userID)
		if err != nil {
			return fmt.Errorf("invalid user ID %s: %w", userID, err)
		}
		userIDs[i] = id
	}

	roleIDs := make([]uint, len(req.RoleIDs))
	for i, roleID := range req.RoleIDs {
		id, err := s.parseRoleID(roleID)
		if err != nil {
			return fmt.Errorf("invalid role ID %s: %w", roleID, err)
		}
		roleIDs[i] = id
	}

	// 批量移除
	for _, userID := range userIDs {
		for _, roleID := range roleIDs {
			if err := s.userRepo.RemoveRole(ctx, userID, roleID); err != nil {
				s.logger.WithError(err).WithFields(map[string]interface{}{
					"user_id": userID,
					"role_id": roleID,
				}).Error("Failed to remove role in batch operation")
				continue
			}
		}
	}

	// 记录审计日志
	auditReq := &LogActionRequest{
		Action:     "batch_remove_roles",
		Details:    fmt.Sprintf("Batch removed %d roles from %d users", len(req.RoleIDs), len(req.UserIDs)),
		TenantID:   req.TenantID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
	}

	if err := s.logAudit(ctx, auditReq); err != nil {
		s.logger.WithError(err).Warn("Failed to log audit action")
	}

	return nil
}

// GetRoleStats 获取角色统计
func (s *roleService) GetRoleStats(ctx context.Context, tenantID string) (*RoleStats, error) {
	roles, err := s.roleRepo.List(ctx, repositories.RoleListOptions{
		TenantID: tenantID,
		Limit:    1000, // 获取所有角色
		Offset:   0,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get roles for stats: %w", err)
	}

	stats := &RoleStats{
		TotalRoles:    int64(len(roles)),
		RolesByStatus: make(map[string]int64),
	}

	for _, role := range roles {
		// 获取角色用户数
		users, err := s.userRepo.GetUsersByRole(ctx, role.ID)
		if err == nil {
			stats.UsersByRole[role.Name] = int64(len(users))
		}

		// 按状态统计（如果角色有状态字段）
		status := "active" // 默认状态
		stats.RolesByStatus[status]++
	}

	return stats, nil
}

// 辅助方法

// validateCreateRoleRequest 验证创建角色请求
func (s *roleService) validateCreateRoleRequest(req *CreateRoleRequest) error {
	if req.Name == "" {
		return fmt.Errorf("role name is required")
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		return fmt.Errorf("role name must be between 2 and 100 characters")
	}
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	return nil
}

// parseUserID 解析用户ID
func (s *roleService) parseUserID(userID string) (uint, error) {
	id, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %s", userID)
	}
	return uint(id), nil
}

// parseRoleID 解析角色ID
func (s *roleService) parseRoleID(roleID string) (uint, error) {
	id, err := strconv.ParseUint(roleID, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid role ID format: %s", roleID)
	}
	return uint(id), nil
}

// convertToRoleResponse 转换为角色响应格式
func (s *roleService) convertToRoleResponse(role *models.Role) *RoleResponse {
	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		TenantID:    role.TenantID,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

// logAudit 记录审计日志
func (s *roleService) logAudit(ctx context.Context, req *LogActionRequest) error {
	// 解析用户ID
	var userID uint
	if req.UserID != "" {
		id, err := s.parseUserID(req.UserID)
		if err != nil {
			return err
		}
		userID = id
	}

	audit := &models.DocumentAudit{
		UserID:    userID,
		TenantID:  req.TenantID,
		Action:    req.Action,
		Details:   req.Details,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
	}

	return s.auditRepo.Create(ctx, audit)
}