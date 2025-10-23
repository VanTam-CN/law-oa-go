package mocks

import (
	"context"
	"errors"

	"github.com/law-oa-go/document-service/internal/models"
)

// UserRepository 用户仓库模拟
type UserRepository struct {
	users    map[uint]*models.User
	nextID   uint
	userRoles map[uint][]*models.UserRole
}

// NewUserRepository 创建用户仓库模拟
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:      make(map[uint]*models.User),
		nextID:     1,
		userRoles:  make(map[uint][]*models.UserRole),
	}
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*models.User, error) {
	if user, exists := r.users[id]; exists {
		return user, nil
	}
	return nil, errors.New("user not found")
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = r.nextID
	r.users[user.ID] = user
	r.nextID++
	return nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	if _, exists := r.users[user.ID]; exists {
		r.users[user.ID] = user
		return nil
	}
	return errors.New("user not found")
}

// Delete 删除用户
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := r.users[id]; exists {
		delete(r.users, id)
		// 删除用户角色关联
		delete(r.userRoles, id)
		return nil
	}
	return errors.New("user not found")
}

// List 列出用户
func (r *UserRepository) List(ctx context.Context, options UserListOptions) ([]*models.User, int64, error) {
	var result []*models.User
	var total int64

	for _, user := range r.users {
		// 应用过滤条件
		if options.TenantID != "" && user.TenantID != options.TenantID {
			continue
		}
		if options.Status != "" && user.Status != options.Status {
			continue
		}
		if options.Search != "" {
			if !contains(user.Username, options.Search) &&
				!contains(user.Email, options.Search) &&
				!contains(user.FirstName, options.Search) &&
				!contains(user.LastName, options.Search) {
				continue
			}
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
		result = append(result, user)
	}

	return result, total, nil
}

// GetUserRoles 获取用户角色
func (r *UserRepository) GetUserRoles(ctx context.Context, userID uint) ([]*models.UserRole, error) {
	if roles, exists := r.userRoles[userID]; exists {
		return roles, nil
	}
	return []*models.UserRole{}, nil
}

// AssignRole 分配角色
func (r *UserRepository) AssignRole(ctx context.Context, userRole *models.UserRole) error {
	if _, exists := r.users[userRole.UserID]; !exists {
		return errors.New("user not found")
	}

	// 检查角色是否已存在
	if roles, exists := r.userRoles[userRole.UserID]; exists {
		for _, role := range roles {
			if role.RoleID == userRole.RoleID {
				return errors.New("role already assigned")
			}
		}
	}

	r.userRoles[userRole.UserID] = append(r.userRoles[userRole.UserID], userRole)
	return nil
}

// RemoveRole 移除角色
func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID uint) error {
	if roles, exists := r.userRoles[userID]; exists {
		for i, role := range roles {
			if role.RoleID == roleID {
				r.userRoles[userID] = append(roles[:i], roles[i+1:]...)
				return nil
			}
		}
	}
	return errors.New("role assignment not found")
}

// GetActiveUsers 获取活跃用户
func (r *UserRepository) GetActiveUsers(ctx context.Context, tenantID string) ([]*models.User, error) {
	var result []*models.User
	for _, user := range r.users {
		if user.TenantID == tenantID && user.Status == "active" {
			result = append(result, user)
		}
	}
	return result, nil
}

// GetUsersByRole 根据角色获取用户
func (r *UserRepository) GetUsersByRole(ctx context.Context, roleID uint) ([]*models.User, error) {
	var result []*models.User
	for userID, roles := range r.userRoles {
		for _, role := range roles {
			if role.RoleID == roleID {
				if user, exists := r.users[userID]; exists {
					result = append(result, user)
				}
				break
			}
		}
	}
	return result, nil
}

// 辅助方法
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 (len(s) > len(substr) &&
		  (s[:len(substr)] == substr ||
		   s[len(s)-len(substr):] == substr ||
		   containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// UserListOptions 用户列表选项
type UserListOptions struct {
	TenantID  string
	Status    string
	RoleID    uint
	Search    string
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}