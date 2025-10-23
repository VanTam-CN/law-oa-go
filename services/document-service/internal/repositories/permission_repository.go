package repositories

import (
	"context"
	"fmt"

	"github.com/law-oa-go/document-service/internal/models"
	"gorm.io/gorm"
)

// permissionRepository 权限仓库实现
type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建新的权限仓库
func NewPermissionRepository(db *gorm.DB) DocumentPermissionRepository {
	return &permissionRepository{db: db}
}

// Create 创建权限
func (r *permissionRepository) Create(ctx context.Context, permission *models.DocumentPermission) error {
	if err := r.db.WithContext(ctx).Create(permission).Error; err != nil {
		return fmt.Errorf("failed to create document permission: %w", err)
	}
	return nil
}

// GetByID 根据ID获取权限
func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*models.DocumentPermission, error) {
	var permission models.DocumentPermission
	if err := r.db.WithContext(ctx).
		Preload("Document").
		First(&permission, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("document permission not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get document permission by ID: %w", err)
	}
	return &permission, nil
}

// Update 更新权限
func (r *permissionRepository) Update(ctx context.Context, permission *models.DocumentPermission) error {
	if err := r.db.WithContext(ctx).Save(permission).Error; err != nil {
		return fmt.Errorf("failed to update document permission: %w", err)
	}
	return nil
}

// Delete 删除权限
func (r *permissionRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.DocumentPermission{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete document permission: %w", err)
	}
	return nil
}

// FindByDocument 根据文档ID获取权限
func (r *permissionRepository) FindByDocument(ctx context.Context, documentID uint) ([]*models.DocumentPermission, error) {
	var permissions []*models.DocumentPermission
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to find permissions by document ID: %w", err)
	}
	return permissions, nil
}

// FindByUser 根据用户ID获取权限
func (r *permissionRepository) FindByUser(ctx context.Context, userID uint) ([]*models.DocumentPermission, error) {
	var permissions []*models.DocumentPermission
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Document").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to find permissions by user ID: %w", err)
	}
	return permissions, nil
}

// FindByRole 根据角色ID获取权限
func (r *permissionRepository) FindByRole(ctx context.Context, roleID uint) ([]*models.DocumentPermission, error) {
	var permissions []*models.DocumentPermission
	if err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Preload("Document").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to find permissions by role ID: %w", err)
	}
	return permissions, nil
}

// FindByTenant 根据租户ID获取权限
func (r *permissionRepository) FindByTenant(ctx context.Context, tenantID string) ([]*models.DocumentPermission, error) {
	var permissions []*models.DocumentPermission
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Preload("Document").
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to find permissions by tenant ID: %w", err)
	}
	return permissions, nil
}

// CheckUserPermission 检查用户是否有指定权限
func (r *permissionRepository) CheckUserPermission(ctx context.Context, documentID, userID uint, permission string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentPermission{}).
		Where("document_id = ? AND user_id = ? AND permission = ?", documentID, userID, permission).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}
	return count > 0, nil
}

// CheckRolePermission 检查角色是否有指定权限
func (r *permissionRepository) CheckRolePermission(ctx context.Context, documentID, roleID uint, permission string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentPermission{}).
		Where("document_id = ? AND role_id = ? AND permission = ?", documentID, roleID, permission).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check role permission: %w", err)
	}
	return count > 0, nil
}

// GetUserPermissions 获取用户对文档的所有权限
func (r *permissionRepository) GetUserPermissions(ctx context.Context, documentID, userID uint) ([]string, error) {
	var permissions []string
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentPermission{}).
		Where("document_id = ? AND user_id = ?", documentID, userID).
		Pluck("permission", &permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

// GetRolePermissions 获取角色对文档的所有权限
func (r *permissionRepository) GetRolePermissions(ctx context.Context, documentID, roleID uint) ([]string, error) {
	var permissions []string
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentPermission{}).
		Where("document_id = ? AND role_id = ?", documentID, roleID).
		Pluck("permission", &permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return permissions, nil
}

// BatchCreate 批量创建权限
func (r *permissionRepository) BatchCreate(ctx context.Context, permissions []*models.DocumentPermission) error {
	if len(permissions) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(permissions, 100).Error; err != nil {
		return fmt.Errorf("failed to batch create document permissions: %w", err)
	}
	return nil
}

// BatchDeleteByDocument 批量删除文档的所有权限
func (r *permissionRepository) BatchDeleteByDocument(ctx context.Context, documentID uint) error {
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Delete(&models.DocumentPermission{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete document permissions: %w", err)
	}
	return nil
}

// BatchDeleteByUser 批量删除用户的所有权限
func (r *permissionRepository) BatchDeleteByUser(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.DocumentPermission{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete user permissions: %w", err)
	}
	return nil
}

// GetUserAccessibleDocuments 获取用户可访问的文档
func (r *permissionRepository) GetUserAccessibleDocuments(ctx context.Context, userID uint, permission string) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Joins("JOIN document_permissions ON document_permissions.document_id = documents.id").
		Where("document_permissions.user_id = ? AND document_permissions.permission = ?", userID, permission).
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to get user accessible documents: %w", err)
	}
	return documents, nil
}

// GetRoleAccessibleDocuments 获取角色可访问的文档
func (r *permissionRepository) GetRoleAccessibleDocuments(ctx context.Context, roleID uint, permission string) ([]*models.Document, error) {
	var documents []*models.Document
	if err := r.db.WithContext(ctx).
		Joins("JOIN document_permissions ON document_permissions.document_id = documents.id").
		Where("document_permissions.role_id = ? AND document_permissions.permission = ?", roleID, permission).
		Find(&documents).Error; err != nil {
		return nil, fmt.Errorf("failed to get role accessible documents: %w", err)
	}
	return documents, nil
}

// GetDocumentPermissionsWithUsers 获取文档权限及关联用户信息
func (r *permissionRepository) GetDocumentPermissionsWithUsers(ctx context.Context, documentID uint) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 获取用户权限
	type UserPermissionResult struct {
		UserID     uint   `json:"user_id"`
		Username   string `json:"username"`
		Permission string `json:"permission"`
		CreatedAt  string `json:"created_at"`
	}

	var userPermissions []UserPermissionResult
	if err := r.db.WithContext(ctx).
		Table("document_permissions").
		Select("document_permissions.user_id, users.username, document_permissions.permission, document_permissions.created_at").
		Joins("LEFT JOIN users ON users.id = document_permissions.user_id").
		Where("document_permissions.document_id = ? AND document_permissions.user_id IS NOT NULL", documentID).
		Scan(&userPermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get document user permissions: %w", err)
	}

	// 转换为通用格式
	for _, up := range userPermissions {
		result := map[string]interface{}{
			"type":       "user",
			"user_id":    up.UserID,
			"username":   up.Username,
			"permission": up.Permission,
			"created_at": up.CreatedAt,
		}
		results = append(results, result)
	}

	// 获取角色权限
	type RolePermissionResult struct {
		RoleID     uint   `json:"role_id"`
		RoleName   string `json:"role_name"`
		Permission string `json:"permission"`
		CreatedAt  string `json:"created_at"`
	}

	var rolePermissions []RolePermissionResult
	if err := r.db.WithContext(ctx).
		Table("document_permissions").
		Select("document_permissions.role_id, roles.name as role_name, document_permissions.permission, document_permissions.created_at").
		Joins("LEFT JOIN roles ON roles.id = document_permissions.role_id").
		Where("document_permissions.document_id = ? AND document_permissions.role_id IS NOT NULL", documentID).
		Scan(&rolePermissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get document role permissions: %w", err)
	}

	// 转换为通用格式
	for _, rp := range rolePermissions {
		result := map[string]interface{}{
			"type":       "role",
			"role_id":    rp.RoleID,
			"role_name":  rp.RoleName,
			"permission": rp.Permission,
			"created_at": rp.CreatedAt,
		}
		results = append(results, result)
	}

	return results, nil
}

// UpdatePermission 更新用户或角色的权限
func (r *permissionRepository) UpdatePermission(ctx context.Context, documentID, userID, roleID uint, permission string) error {
	// 构建更新条件
	query := r.db.WithContext(ctx).Model(&models.DocumentPermission{}).
		Where("document_id = ?", documentID)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if roleID > 0 {
		query = query.Where("role_id = ?", roleID)
	}

	// 执行更新
	if err := query.Update("permission", permission).Error; err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
	}

	return nil
}

// DeletePermission 删除用户或角色的指定权限
func (r *permissionRepository) DeletePermission(ctx context.Context, documentID, userID, roleID uint, permission string) error {
	// 构建删除条件
	query := r.db.WithContext(ctx).Where("document_id = ?", documentID)

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if roleID > 0 {
		query = query.Where("role_id = ?", roleID)
	}

	if permission != "" {
		query = query.Where("permission = ?", permission)
	}

	// 执行删除
	if err := query.Delete(&models.DocumentPermission{}).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// CountPermissions 统计权限数量
func (r *permissionRepository) CountPermissions(ctx context.Context, tenantID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 统计用户权限
	var userPermissionCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentPermission{}).
		Where("tenant_id = ? AND user_id IS NOT NULL", tenantID).
		Count(&userPermissionCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count user permissions: %w", err)
	}
	stats["user_permissions"] = userPermissionCount

	// 统计角色权限
	var rolePermissionCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.DocumentPermission{}).
		Where("tenant_id = ? AND role_id IS NOT NULL", tenantID).
		Count(&rolePermissionCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count role permissions: %w", err)
	}
	stats["role_permissions"] = rolePermissionCount

	// 按权限类型统计
	var permissionTypeStats []struct {
		Permission string `json:"permission"`
		Count      int64  `json:"count"`
	}

	if err := r.db.WithContext(ctx).
		Model(&models.DocumentPermission{}).
		Select("permission, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("permission").
		Scan(&permissionTypeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to count permissions by type: %w", err)
	}

	for _, stat := range permissionTypeStats {
		stats[fmt.Sprintf("permission_%s", stat.Permission)] = stat.Count
	}

	return stats, nil
}