package mocks

import (
	"context"
	"errors"

	"github.com/law-oa-go/document-service/internal/models"
)

// RoleRepository 角色仓库模拟
type RoleRepository struct {
	roles   map[uint]*models.Role
	nextID  uint
}

// NewRoleRepository 创建角色仓库模拟
func NewRoleRepository() *RoleRepository {
	return &RoleRepository{
		roles:  make(map[uint]*models.Role),
		nextID: 1,
	}
}

// GetByID 根据ID获取角色
func (r *RoleRepository) GetByID(ctx context.Context, id uint) (*models.Role, error) {
	if role, exists := r.roles[id]; exists {
		return role, nil
	}
	return nil, errors.New("role not found")
}

// GetByName 根据名称获取角色
func (r *RoleRepository) GetByName(ctx context.Context, name, tenantID string) (*models.Role, error) {
	for _, role := range r.roles {
		if role.Name == name && role.TenantID == tenantID {
			return role, nil
		}
	}
	return nil, errors.New("role not found")
}

// Create 创建角色
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	role.ID = r.nextID
	r.roles[role.ID] = role
	r.nextID++
	return nil
}

// Update 更新角色
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	if _, exists := r.roles[role.ID]; exists {
		r.roles[role.ID] = role
		return nil
	}
	return errors.New("role not found")
}

// Delete 删除角色
func (r *RoleRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := r.roles[id]; exists {
		delete(r.roles, id)
		return nil
	}
	return errors.New("role not found")
}

// List 列出角色
func (r *RoleRepository) List(ctx context.Context, options RoleListOptions) ([]*models.Role, int64, error) {
	var result []*models.Role
	var total int64

	for _, role := range r.roles {
		// 应用过滤条件
		if options.TenantID != "" && role.TenantID != options.TenantID {
			continue
		}
		if options.Search != "" && !contains(role.Name, options.Search) && !contains(role.Description, options.Search) {
			continue
		}

		total++
		// 应用分页
		if options.Offset > 0 {
			options.Offset--
			continue
		}
		if options.Limit > 0 && len(result) >= options.Limit {
			continue
		}
		result = append(result, role)
	}

	return result, total, nil
}

// RoleListOptions 角色列表选项
type RoleListOptions struct {
	TenantID  string
	Search    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}