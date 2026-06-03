package services

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// RBACService 角色权限服务
type RBACService struct {
	db                 *gorm.DB
	roleRepo           *repositories.RoleRepository
	permissionRepo     *repositories.PermissionRepository
	userRoleRepo       *repositories.UserRoleRepository
	rolePermissionRepo *repositories.RolePermissionRepository
}

// NewRBACService 创建RBAC服务
func NewRBACService(db *gorm.DB) *RBACService {
	return &RBACService{
		db:                 db,
		roleRepo:           repositories.NewRoleRepository(db),
		permissionRepo:     repositories.NewPermissionRepository(db),
		userRoleRepo:       repositories.NewUserRoleRepository(db),
		rolePermissionRepo: repositories.NewRolePermissionRepository(db),
	}
}

// ========== 角色管理 ==========

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Code        string `json:"code" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
	Status      string `json:"status" binding:"oneof=active inactive"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
	Status      string `json:"status" binding:"oneof=active inactive"`
	SortOrder   int    `json:"sort_order"`
}

// RoleQueryParams 角色查询参数
type RoleQueryParams = repositories.RoleQueryParams

// RolePageResponse 角色分页响应
type RolePageResponse struct {
	List     []models.Role `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page_num"`
	PageSize int           `json:"page_size"`
}

// CreateRole 创建角色
func (s *RBACService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*models.Role, error) {
	if req.Status == "" {
		req.Status = "active"
	}

	// 检查角色名称是否已存在
	existingRole, _ := s.roleRepo.FindByName(ctx, req.Name)
	if existingRole != nil {
		return nil, errors.New("role name already exists")
	}

	// 检查角色代码是否已存在
	existingRole, _ = s.roleRepo.FindByCode(ctx, req.Code)
	if existingRole != nil {
		return nil, errors.New("role code already exists")
	}

	role := &models.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return role, nil
}

// GetRoleById 获取角色详情
func (s *RBACService) GetRoleById(ctx context.Context, id uint) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// GetRoleList 获取角色列表
func (s *RBACService) GetRoleList(ctx context.Context, params *RoleQueryParams) (*RolePageResponse, error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}

	roles, total, err := s.roleRepo.FindByParams(ctx, params)
	if err != nil {
		return nil, err
	}

	return &RolePageResponse{
		List:     roles,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// GetAllRoles 获取所有角色（不分页）
func (s *RBACService) GetAllRoles(ctx context.Context) ([]models.Role, error) {
	roles, err := s.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateRole 更新角色
func (s *RBACService) UpdateRole(ctx context.Context, id uint, req *UpdateRoleRequest) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查角色名称是否已存在（排除自己）
	existingRole, _ := s.roleRepo.FindByName(ctx, req.Name)
	if existingRole != nil && existingRole.ID != id {
		return nil, errors.New("role name already exists")
	}

	role.Name = req.Name
	role.Description = req.Description
	role.Status = req.Status
	role.SortOrder = req.SortOrder

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

// DeleteRole 删除角色
func (s *RBACService) DeleteRole(ctx context.Context, id uint) error {
	// 检查是否有用户使用该角色
	userCount, err := s.userRoleRepo.CountByRoleID(ctx, id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return errors.New("cannot delete role: users are still assigned to this role")
	}

	if err := s.rolePermissionRepo.DeleteByRoleID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// UpdateRoleStatus 更新角色状态
func (s *RBACService) UpdateRoleStatus(ctx context.Context, id uint, status string) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	role.Status = status
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return fmt.Errorf("failed to update role status: %w", err)
	}

	return nil
}

// ========== 权限管理 ==========

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Code      string `json:"code" binding:"required,min=2,max=100"`
	Type      string `json:"type" binding:"oneof=menu button api"`
	ParentID  *uint  `json:"parent_id"`
	Path      string `json:"path" binding:"max=255"`
	Icon      string `json:"icon" binding:"max=100"`
	Component string `json:"component" binding:"max=255"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status" binding:"oneof=active inactive"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Code      string `json:"code" binding:"required,min=2,max=100"`
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Type      string `json:"type" binding:"oneof=menu button api"`
	ParentID  *uint  `json:"parent_id"`
	Path      string `json:"path" binding:"max=255"`
	Icon      string `json:"icon" binding:"max=100"`
	Component string `json:"component" binding:"max=255"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status" binding:"oneof=active inactive"`
}

// PermissionQueryParams 权限查询参数
type PermissionQueryParams = repositories.PermissionQueryParams

// CreatePermission 创建权限
func (s *RBACService) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*models.Permission, error) {
	if req.Type == "" {
		req.Type = "menu"
	}
	if req.Status == "" {
		req.Status = "active"
	}

	// 检查权限代码是否已存在
	existingPermission, _ := s.permissionRepo.FindByCode(ctx, req.Code)
	if existingPermission != nil {
		return nil, errors.New("permission code already exists")
	}

	permission := &models.Permission{
		Name:      req.Name,
		Code:      req.Code,
		Type:      req.Type,
		ParentID:  req.ParentID,
		Path:      req.Path,
		Icon:      req.Icon,
		Component: req.Component,
		SortOrder: req.SortOrder,
		Status:    req.Status,
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

// GetPermissionList 获取权限列表（树形结构）
func (s *RBACService) GetPermissionList(ctx context.Context, params *PermissionQueryParams) ([]models.Permission, error) {
	permissions, err := s.permissionRepo.FindByParams(ctx, params)
	if err != nil {
		return nil, err
	}

	// 构建树形结构
	return s.buildPermissionTree(permissions), nil
}

// GetAllPermissions 获取所有权限（扁平结构）
func (s *RBACService) GetAllPermissions(ctx context.Context) ([]models.Permission, error) {
	permissions, err := s.permissionRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetPermissionById 获取权限详情
func (s *RBACService) GetPermissionById(ctx context.Context, id uint) (*models.Permission, error) {
	permission, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

// UpdatePermission 更新权限
func (s *RBACService) UpdatePermission(ctx context.Context, id uint, req *UpdatePermissionRequest) (*models.Permission, error) {
	permission, err := s.permissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查权限代码是否已存在（排除自己）
	existingPermission, _ := s.permissionRepo.FindByCode(ctx, req.Code)
	if existingPermission != nil && existingPermission.ID != id {
		return nil, errors.New("permission code already exists")
	}

	permission.Name = req.Name
	permission.Type = req.Type
	permission.ParentID = req.ParentID
	permission.Path = req.Path
	permission.Icon = req.Icon
	permission.Component = req.Component
	permission.SortOrder = req.SortOrder
	permission.Status = req.Status

	if err := s.permissionRepo.Update(ctx, permission); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return permission, nil
}

// DeletePermission 删除权限
func (s *RBACService) DeletePermission(ctx context.Context, id uint) error {
	// 检查是否有子权限
	childCount, err := s.permissionRepo.CountByParentID(ctx, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("cannot delete permission: it has child permissions")
	}

	// 检查是否有角色使用该权限
	roleCount, err := s.rolePermissionRepo.CountByPermissionID(ctx, id)
	if err != nil {
		return err
	}
	if roleCount > 0 {
		return errors.New("cannot delete permission: roles are still assigned to this permission")
	}

	if err := s.rolePermissionRepo.DeleteByPermissionID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	if err := s.permissionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// buildPermissionTree 构建权限树形结构
func (s *RBACService) buildPermissionTree(permissions []models.Permission) []models.Permission {
	permissionMap := make(map[uint]*models.Permission)
	var roots []models.Permission

	// 创建映射
	for i := range permissions {
		permissionMap[permissions[i].ID] = &permissions[i]
	}

	// 构建树形结构
	for i := range permissions {
		permission := &permissions[i]
		if permission.ParentID == nil {
			roots = append(roots, *permission)
		} else {
			if parent, exists := permissionMap[*permission.ParentID]; exists {
				parent.Children = append(parent.Children, *permission)
			}
		}
	}

	return roots
}

// ========== 角色权限关联管理 ==========

// GetRolePermissions 获取角色的权限列表
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID uint) ([]uint, error) {
	permissionIDs, err := s.rolePermissionRepo.FindPermissionIDsByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return permissionIDs, nil
}

// AssignRolePermissions 为角色分配权限
func (s *RBACService) AssignRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	// 删除现有权限关联
	if err := s.rolePermissionRepo.DeleteByRoleID(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete existing role permissions: %w", err)
	}

	// 创建新的权限关联
	for _, permissionID := range permissionIDs {
		rolePermission := &models.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}
		if err := s.rolePermissionRepo.Create(ctx, rolePermission); err != nil {
			return fmt.Errorf("failed to create role permission: %w", err)
		}
	}

	return nil
}

// ========== 用户角色关联管理 ==========

// GetUserRoles 获取用户的角色列表
func (s *RBACService) GetUserRoles(ctx context.Context, userID uint) ([]models.Role, error) {
	roles, err := s.userRoleRepo.FindRolesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// AssignUserRoles 为用户分配角色
func (s *RBACService) AssignUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	// 删除现有角色关联
	if err := s.userRoleRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete existing user roles: %w", err)
	}

	// 创建新的角色关联
	for _, roleID := range roleIDs {
		userRole := &models.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
		if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
			return fmt.Errorf("failed to create user role: %w", err)
		}
	}

	return nil
}

// ========== 权限检查 ==========

// CheckUserPermission 检查用户是否有指定权限
func (s *RBACService) CheckUserPermission(ctx context.Context, userID uint, permissionCode string) (bool, error) {
	// 获取用户的角色
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}

	// 获取权限
	permission, err := s.permissionRepo.FindByCode(ctx, permissionCode)
	if err != nil {
		return false, err
	}

	// 检查每个角色是否有该权限
	for _, role := range roles {
		hasPermission, err := s.rolePermissionRepo.ExistsByRoleIDAndPermissionID(ctx, role.ID, permission.ID)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions 获取用户的所有权限
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uint) ([]models.Permission, error) {
	// 获取用户的角色
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 获取所有权限ID
	var allPermissionIDs []uint
	for _, role := range roles {
		permissionIDs, err := s.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		allPermissionIDs = append(allPermissionIDs, permissionIDs...)
	}

	// 去重
	permissionIDMap := make(map[uint]struct{})
	for _, id := range allPermissionIDs {
		permissionIDMap[id] = struct{}{}
	}

	// 获取权限详情
	var permissions []models.Permission
	for id := range permissionIDMap {
		permission, err := s.permissionRepo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, *permission)
	}

	return permissions, nil
}
