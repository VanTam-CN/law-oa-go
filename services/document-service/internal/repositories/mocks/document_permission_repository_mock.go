package mocks

import (
	"context"
	"errors"

	"github.com/law-oa-go/document-service/internal/models"
)

// DocumentPermissionRepository 文档权限仓库模拟
type DocumentPermissionRepository struct {
	permissions map[uint]*models.DocumentPermission
	nextID      uint
}

// NewDocumentPermissionRepository 创建文档权限仓库模拟
func NewDocumentPermissionRepository() *DocumentPermissionRepository {
	return &DocumentPermissionRepository{
		permissions: make(map[uint]*models.DocumentPermission),
		nextID:      1,
	}
}

// Create 创建权限
func (r *DocumentPermissionRepository) Create(ctx context.Context, permission *models.DocumentPermission) error {
	permission.ID = r.nextID
	r.permissions[permission.ID] = permission
	r.nextID++
	return nil
}

// GetByID 根据ID获取权限
func (r *DocumentPermissionRepository) GetByID(ctx context.Context, id uint) (*models.DocumentPermission, error) {
	if permission, exists := r.permissions[id]; exists {
		return permission, nil
	}
	return nil, errors.New("permission not found")
}

// FindByDocument 根据文档查找权限
func (r *DocumentPermissionRepository) FindByDocument(ctx context.Context, documentID uint) ([]*models.DocumentPermission, error) {
	var result []*models.DocumentPermission
	for _, permission := range r.permissions {
		if permission.DocumentID == documentID {
			result = append(result, permission)
		}
	}
	return result, nil
}

// CheckUserPermission 检查用户权限
func (r *DocumentPermissionRepository) CheckUserPermission(ctx context.Context, documentID, userID uint, permission string) (bool, error) {
	for _, perm := range r.permissions {
		if perm.DocumentID == documentID && perm.UserID != nil && *perm.UserID == userID && perm.Permission == permission {
			return true, nil
		}
	}
	return false, nil
}

// CheckRolePermission 检查角色权限
func (r *DocumentPermissionRepository) CheckRolePermission(ctx context.Context, documentID, roleID uint, permission string) (bool, error) {
	for _, perm := range r.permissions {
		if perm.DocumentID == documentID && perm.RoleID != nil && *perm.RoleID == roleID && perm.Permission == permission {
			return true, nil
		}
	}
	return false, nil
}

// GetUserPermissions 获取用户权限
func (r *DocumentPermissionRepository) GetUserPermissions(ctx context.Context, documentID, userID uint) ([]string, error) {
	var permissions []string
	for _, perm := range r.permissions {
		if perm.DocumentID == documentID && perm.UserID != nil && *perm.UserID == userID {
			permissions = append(permissions, perm.Permission)
		}
	}
	return permissions, nil
}

// GetRolePermissions 获取角色权限
func (r *DocumentPermissionRepository) GetRolePermissions(ctx context.Context, documentID, roleID uint) ([]string, error) {
	var permissions []string
	for _, perm := range r.permissions {
		if perm.DocumentID == documentID && perm.RoleID != nil && *perm.RoleID == roleID {
			permissions = append(permissions, perm.Permission)
		}
	}
	return permissions, nil
}

// GetUserAccessibleDocuments 获取用户可访问的文档
func (r *DocumentPermissionRepository) GetUserAccessibleDocuments(ctx context.Context, userID uint, permission string) ([]*models.Document, error) {
	// 这里简化实现，实际需要查询文档仓库
	return []*models.Document{}, nil
}

// GetRoleAccessibleDocuments 获取角色可访问的文档
func (r *DocumentPermissionRepository) GetRoleAccessibleDocuments(ctx context.Context, roleID uint, permission string) ([]*models.Document, error) {
	// 这里简化实现，实际需要查询文档仓库
	return []*models.Document{}, nil
}

// UpdatePermission 更新权限
func (r *DocumentPermissionRepository) UpdatePermission(ctx context.Context, documentID, userID, roleID uint, permission string) error {
	for _, perm := range r.permissions {
		if perm.DocumentID == documentID {
			if userID != 0 && perm.UserID != nil && *perm.UserID == userID {
				perm.Permission = permission
				return nil
			}
			if roleID != 0 && perm.RoleID != nil && *perm.RoleID == roleID {
				perm.Permission = permission
				return nil
			}
		}
	}
	return errors.New("permission not found")
}

// DeletePermission 删除权限
func (r *DocumentPermissionRepository) DeletePermission(ctx context.Context, documentID, userID, roleID uint, permission string) error {
	for id, perm := range r.permissions {
		if perm.DocumentID == documentID {
			if userID != 0 && perm.UserID != nil && *perm.UserID == userID {
				delete(r.permissions, id)
				return nil
			}
			if roleID != 0 && perm.RoleID != nil && *perm.RoleID == roleID {
				delete(r.permissions, id)
				return nil
			}
			if permission != "" && perm.Permission == permission {
				delete(r.permissions, id)
				return nil
			}
		}
	}
	return errors.New("permission not found")
}