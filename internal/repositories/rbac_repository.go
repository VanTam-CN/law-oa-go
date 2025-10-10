package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// RoleQueryParams 角色查询参数
type RoleQueryParams struct {
	Name     string `form:"name"`
	Code     string `form:"code"`
	Status   string `form:"status"`
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
}

// PermissionQueryParams 权限查询参数
type PermissionQueryParams struct {
	Name   string `form:"name"`
	Code   string `form:"code"`
	Type   string `form:"type"`
	Status string `form:"status"`
}

// RoleRepository 角色仓库
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色仓库
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create 创建角色
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// FindByID 根据ID查找角色
func (r *RoleRepository) FindByID(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// FindByName 根据名称查找角色
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// FindByCode 根据代码查找角色
func (r *RoleRepository) FindByCode(ctx context.Context, code string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// FindAll 查找所有角色
func (r *RoleRepository) FindAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).Where("status = ?", "active").Order("sort_order ASC, id ASC").Find(&roles).Error
	return roles, err
}

// FindByParams 根据参数查找角色（分页）
func (r *RoleRepository) FindByParams(ctx context.Context, params *RoleQueryParams) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Role{})

	// 过滤条件
	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Code != "" {
		query = query.Where("code LIKE ?", "%"+params.Code+"%")
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Order("sort_order ASC, id ASC").Offset(offset).Limit(params.PageSize).Find(&roles).Error

	return roles, total, err
}

// Update 更新角色
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete 删除角色
func (r *RoleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

// PermissionRepository 权限仓库
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限仓库
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create 创建权限
func (r *PermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// FindByID 根据ID查找权限
func (r *PermissionRepository) FindByID(ctx context.Context, id uint) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).Preload("Parent").Preload("Children").First(&permission, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// FindByCode 根据代码查找权限
func (r *PermissionRepository) FindByCode(ctx context.Context, code string) (*models.Permission, error) {
	var permission models.Permission
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// FindAll 查找所有权限
func (r *PermissionRepository) FindAll(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.WithContext(ctx).Where("status = ?", "active").Order("parent_id ASC, sort_order ASC, id ASC").Find(&permissions).Error
	return permissions, err
}

// FindByParams 根据参数查找权限
func (r *PermissionRepository) FindByParams(ctx context.Context, params *PermissionQueryParams) ([]models.Permission, error) {
	var permissions []models.Permission
	query := r.db.WithContext(ctx).Model(&models.Permission{})

	// 过滤条件
	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Code != "" {
		query = query.Where("code LIKE ?", "%"+params.Code+"%")
	}
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	err := query.Order("parent_id ASC, sort_order ASC, id ASC").Find(&permissions).Error
	return permissions, err
}

// Update 更新权限
func (r *PermissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// Delete 删除权限
func (r *PermissionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Permission{}, id).Error
}

// CountByParentID 根据父权限ID统计子权限数量
func (r *PermissionRepository) CountByParentID(ctx context.Context, parentID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Permission{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// RolePermissionRepository 角色权限关联仓库
type RolePermissionRepository struct {
	db *gorm.DB
}

// NewRolePermissionRepository 创建角色权限关联仓库
func NewRolePermissionRepository(db *gorm.DB) *RolePermissionRepository {
	return &RolePermissionRepository{db: db}
}

// Create 创建角色权限关联
func (r *RolePermissionRepository) Create(ctx context.Context, rolePermission *models.RolePermission) error {
	return r.db.WithContext(ctx).Create(rolePermission).Error
}

// FindPermissionIDsByRoleID 根据角色ID查找权限ID列表
func (r *RolePermissionRepository) FindPermissionIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	var permissionIDs []uint
	err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &permissionIDs).Error
	return permissionIDs, err
}

// ExistsByRoleIDAndPermissionID 检查角色权限关联是否存在
func (r *RolePermissionRepository) ExistsByRoleIDAndPermissionID(ctx context.Context, roleID, permissionID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}

// CountByRoleID 根据角色ID统计权限数量
func (r *RolePermissionRepository) CountByRoleID(ctx context.Context, roleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	return count, err
}

// CountByPermissionID 根据权限ID统计角色数量
func (r *RolePermissionRepository) CountByPermissionID(ctx context.Context, permissionID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RolePermission{}).
		Where("permission_id = ?", permissionID).
		Count(&count).Error
	return count, err
}

// DeleteByRoleID 根据角色ID删除关联
func (r *RolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error
}

// DeleteByPermissionID 根据权限ID删除关联
func (r *RolePermissionRepository) DeleteByPermissionID(ctx context.Context, permissionID uint) error {
	return r.db.WithContext(ctx).Where("permission_id = ?", permissionID).Delete(&models.RolePermission{}).Error
}

// UserRoleRepository 用户角色关联仓库
type UserRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository 创建用户角色关联仓库
func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

// Create 创建用户角色关联
func (r *UserRoleRepository) Create(ctx context.Context, userRole *models.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

// FindRolesByUserID 根据用户ID查找角色列表
func (r *UserRoleRepository) FindRolesByUserID(ctx context.Context, userID uint) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).Model(&models.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, "active").
		Find(&roles).Error
	return roles, err
}

// CountByRoleID 根据角色ID统计用户数量
func (r *UserRoleRepository) CountByRoleID(ctx context.Context, roleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	return count, err
}

// DeleteByUserID 根据用户ID删除关联
func (r *UserRoleRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserRole{}).Error
}

// DeleteByRoleID 根据角色ID删除关联
func (r *UserRoleRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Where("role_id = ?", roleID).Delete(&models.UserRole{}).Error
}