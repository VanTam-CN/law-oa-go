package rbac

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/models"
	"strings"
	"time"
)

var (
	rbacDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rbac_duration_seconds",
		Help:    "Duration of RBAC operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})
	
	rbacErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rbac_errors_total",
		Help: "Total number of RBAC errors",
	}, []string{"operation"})
	
	permissionChecks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rbac_permission_checks_total",
		Help: "Total number of permission checks",
	})
)

// Permission 权限定义
type Permission struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// Role 角色定义
type Role struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Key         string       `json:"key"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// RBACService RBAC服务
type RBACService struct {
	db           *gorm.DB
	cacheService *cache.CacheService
}

// NewRBACService 创建RBAC服务
func NewRBACService(db *gorm.DB, cacheService *cache.CacheService) *RBACService {
	return &RBACService{
		db:           db,
		cacheService: cacheService,
	}
}

// CreatePermission 创建权限
func (s *RBACService) CreatePermission(ctx context.Context, permission *Permission) error {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("create_permission").Observe(time.Since(start).Seconds())
	}()
	
	modelPermission := &models.Permission{
		PermissionName: permission.Name,
		PermissionKey:  permission.Key,
		Path:          permission.Resource,
		Component:     permission.Action,
		Remark:        permission.Description,
		Status:        "active",
	}
	
	if err := s.db.Create(modelPermission).Error; err != nil {
		rbacErrors.WithLabelValues("create_permission").Inc()
		return fmt.Errorf("failed to create permission: %w", err)
	}
	
	permission.ID = modelPermission.ID
	
	// 清除缓存
	s.clearPermissionCache(ctx)
	
	return nil
}

// CreateRole 创建角色
func (s *RBACService) CreateRole(ctx context.Context, role *Role, permissionIDs []uint) error {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("create_role").Observe(time.Since(start).Seconds())
	}()
	
	modelRole := &models.Role{
		RoleName: role.Name,
		RoleKey:  role.Key,
		Remark:   role.Description,
		Status:   "active",
	}
	
	if err := s.db.Create(modelRole).Error; err != nil {
		rbacErrors.WithLabelValues("create_role").Inc()
		return fmt.Errorf("failed to create role: %w", err)
	}
	
	// 分配权限
	if len(permissionIDs) > 0 {
		if err := s.AssignPermissionsToRole(ctx, modelRole.ID, permissionIDs); err != nil {
			rbacErrors.WithLabelValues("assign_permissions").Inc()
			return fmt.Errorf("failed to assign permissions: %w", err)
		}
	}
	
	role.ID = modelRole.ID
	
	// 清除缓存
	s.clearRoleCache(ctx)
	
	return nil
}

// AssignPermissionsToRole 为角色分配权限
func (s *RBACService) AssignPermissionsToRole(ctx context.Context, roleID uint, permissionIDs []uint) error {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("assign_permissions").Observe(time.Since(start).Seconds())
	}()
	
	// 创建角色-权限关联表（如果不存在）
	if !s.db.Migrator().HasTable("role_permissions") {
		if err := s.db.Migrator().CreateTable(&RolePermission{}); err != nil {
			return fmt.Errorf("failed to create role_permissions table: %w", err)
		}
	}
	
	// 清除现有权限
	if err := s.db.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
		return fmt.Errorf("failed to clear existing permissions: %w", err)
	}
	
	// 分配新权限
	for _, permID := range permissionIDs {
		rolePerm := &RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		}
		if err := s.db.Create(rolePerm).Error; err != nil {
			return fmt.Errorf("failed to assign permission %d: %w", permID, err)
		}
	}
	
	// 清除缓存
	s.clearRoleCache(ctx)
	s.clearUserPermissionsCache(ctx, roleID)
	
	return nil
}

// AssignRoleToUser 为用户分配角色
func (s *RBACService) AssignRoleToUser(ctx context.Context, userID, roleID uint) error {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("assign_role").Observe(time.Since(start).Seconds())
	}()
	
	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		rbacErrors.WithLabelValues("user_not_found").Inc()
		return fmt.Errorf("user not found: %w", err)
	}
	
	// 检查角色是否存在
	var role models.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		rbacErrors.WithLabelValues("role_not_found").Inc()
		return fmt.Errorf("role not found: %w", err)
	}
	
	// 分配角色（替换现有角色）
	if err := s.db.Model(&user).Association("Roles").Replace(&role); err != nil {
		rbacErrors.WithLabelValues("assign_role_failed").Inc()
		return fmt.Errorf("failed to assign role: %w", err)
	}
	
	// 更新用户的角色字段
	if err := s.db.Model(&user).Update("role", role.RoleKey).Error; err != nil {
		rbacErrors.WithLabelValues("update_user_role").Inc()
		return fmt.Errorf("failed to update user role: %w", err)
	}
	
	// 清除缓存
	s.clearUserCache(ctx, userID)
	
	return nil
}

// GetUserPermissions 获取用户权限
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uint) ([]Permission, error) {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("get_user_permissions").Observe(time.Since(start).Seconds())
	}()
	
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("user_permissions:%d", userID)
	var cachedPerms []Permission
	if err := s.cacheService.Get(ctx, cacheKey, &cachedPerms); err == nil {
		return cachedPerms, nil
	}
	
	// 从数据库获取
	var user models.User
	if err := s.db.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
		rbacErrors.WithLabelValues("user_not_found").Inc()
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	var permissions []Permission
	seen := make(map[string]bool)
	
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if !seen[perm.PermissionKey] {
				permissions = append(permissions, Permission{
					ID:          perm.ID,
					Name:        perm.PermissionName,
					Key:         perm.PermissionKey,
					Description: perm.Remark,
					Resource:    perm.Path,
					Action:      perm.Component,
				})
				seen[perm.PermissionKey] = true
			}
		}
	}
	
	// 缓存结果
	s.cacheService.Set(ctx, cacheKey, permissions, time.Hour)
	
	return permissions, nil
}

// CheckPermission 检查用户是否有特定权限
func (s *RBACService) CheckPermission(ctx context.Context, userID uint, permissionKey string) (bool, error) {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("check_permission").Observe(time.Since(start).Seconds())
		permissionChecks.Inc()
	}()
	
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	
	for _, perm := range permissions {
		if perm.Key == permissionKey {
			return true, nil
		}
	}
	
	return false, nil
}

// HasPermission 中间件：检查用户是否有特定权限
func (s *RBACService) HasPermission(permissionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}
		
		hasPermission, err := s.CheckPermission(c.Request.Context(), userID.(uint), permissionKey)
		if err != nil {
			rbacErrors.WithLabelValues("permission_check_failed").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限检查失败",
			})
			c.Abort()
			return
		}
		
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// HasAnyPermission 中间件：检查用户是否有任意一个权限
func (s *RBACService) HasAnyPermission(permissionKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}
		
		permissions, err := s.GetUserPermissions(c.Request.Context(), userID.(uint))
		if err != nil {
			rbacErrors.WithLabelValues("permission_check_failed").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限检查失败",
			})
			c.Abort()
			return
		}
		
		permissionMap := make(map[string]bool)
		for _, perm := range permissions {
			permissionMap[perm.Key] = true
		}
		
		for _, requiredKey := range permissionKeys {
			if permissionMap[requiredKey] {
				c.Next()
				return
			}
		}
		
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足",
		})
		c.Abort()
	}
}

// HasAllPermissions 中间件：检查用户是否有所有权限
func (s *RBACService) HasAllPermissions(permissionKeys []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}
		
		permissions, err := s.GetUserPermissions(c.Request.Context(), userID.(uint))
		if err != nil {
			rbacErrors.WithLabelValues("permission_check_failed").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限检查失败",
			})
			c.Abort()
			return
		}
		
		permissionMap := make(map[string]bool)
		for _, perm := range permissions {
			permissionMap[perm.Key] = true
		}
		
		for _, requiredKey := range permissionKeys {
			if !permissionMap[requiredKey] {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "权限不足",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// HasRole 中间件：检查用户是否有特定角色
func (s *RBACService) HasRole(roleKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}
		
		if userRole.(string) != roleKey {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// GetRolePermissions 获取角色的所有权限
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID uint) ([]Permission, error) {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("get_role_permissions").Observe(time.Since(start).Seconds())
	}()
	
	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("role_permissions:%d", roleID)
	var cachedPerms []Permission
	if err := s.cacheService.Get(ctx, cacheKey, &cachedPerms); err == nil {
		return cachedPerms, nil
	}
	
	// 从数据库获取
	var role models.Role
	if err := s.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		rbacErrors.WithLabelValues("role_not_found").Inc()
		return nil, fmt.Errorf("role not found: %w", err)
	}
	
	var permissions []Permission
	for _, perm := range role.Permissions {
		permissions = append(permissions, Permission{
			ID:          perm.ID,
			Name:        perm.PermissionName,
			Key:         perm.PermissionKey,
			Description: perm.Remark,
			Resource:    perm.Path,
			Action:      perm.Component,
		})
	}
	
	// 缓存结果
	s.cacheService.Set(ctx, cacheKey, permissions, time.Hour)
	
	return permissions, nil
}

// RemoveRoleFromUser 移除用户的角色
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID uint) error {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("remove_role").Observe(time.Since(start).Seconds())
	}()
	
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		rbacErrors.WithLabelValues("user_not_found").Inc()
		return fmt.Errorf("user not found: %w", err)
	}
	
	// 清除角色关联
	if err := s.db.Model(&user).Association("Roles").Clear(); err != nil {
		rbacErrors.WithLabelValues("remove_role_failed").Inc()
		return fmt.Errorf("failed to remove role: %w", err)
	}
	
	// 更新用户的角色字段
	if err := s.db.Model(&user).Update("role", "").Error; err != nil {
		rbacErrors.WithLabelValues("update_user_role").Inc()
		return fmt.Errorf("failed to update user role: %w", err)
	}
	
	// 清除缓存
	s.clearUserCache(ctx, userID)
	
	return nil
}

// DeleteRole 删除角色
func (s *RBACService) DeleteRole(ctx context.Context, roleID uint) error {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("delete_role").Observe(time.Since(start).Seconds())
	}()
	
	if err := s.db.Delete(&models.Role{}, roleID).Error; err != nil {
		rbacErrors.WithLabelValues("delete_role_failed").Inc()
		return fmt.Errorf("failed to delete role: %w", err)
	}
	
	// 清除相关缓存
	s.clearRoleCache(ctx)
	s.clearUserPermissionsCache(ctx, roleID)
	
	return nil
}

// GetResourcePermissions 获取资源的所有权限
func (s *RBACService) GetResourcePermissions(ctx context.Context, resource string) ([]Permission, error) {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("get_resource_permissions").Observe(time.Since(start).Seconds())
	}()
	
	var modelsPermissions []models.Permission
	if err := s.db.Where("path = ?", resource).Find(&modelsPermissions).Error; err != nil {
		rbacErrors.WithLabelValues("resource_not_found").Inc()
		return nil, fmt.Errorf("failed to get resource permissions: %w", err)
	}
	
	var permissions []Permission
	for _, perm := range modelsPermissions {
		permissions = append(permissions, Permission{
			ID:          perm.ID,
			Name:        perm.PermissionName,
			Key:         perm.PermissionKey,
			Description: perm.Remark,
			Resource:    perm.Path,
			Action:      perm.Component,
		})
	}
	
	return permissions, nil
}

// 缓存清理方法
func (s *RBACService) clearUserCache(ctx context.Context, userID uint) {
	cacheKey := fmt.Sprintf("user_permissions:%d", userID)
	s.cacheService.Delete(ctx, cacheKey)
}

func (s *RBACService) clearRoleCache(ctx context.Context) {
	// 清除所有角色相关的缓存
	pattern := "role_permissions:*"
	s.cacheService.ClearPattern(ctx, pattern)
}

func (s *RBACService) clearUserPermissionsCache(ctx context.Context, roleID uint) {
	// 清除使用该角色的用户权限缓存
	var users []models.User
	s.db.Where("role_id = ?", roleID).Find(&users)
	
	for _, user := range users {
		s.clearUserCache(ctx, user.ID)
	}
}

func (s *RBACService) clearPermissionCache(ctx context.Context) {
	// 清除所有权限相关的缓存
	s.cacheService.ClearPattern(ctx, "user_permissions:*")
	s.cacheService.ClearPattern(ctx, "role_permissions:*")
}

// RolePermission 角色-权限关联表
type RolePermission struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	RoleID       uint   `json:"role_id" gorm:"not null;index"`
	PermissionID uint   `json:"permission_id" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

// IsResourceAccessible 检查用户是否可以访问特定资源
func (s *RBACService) IsResourceAccessible(ctx context.Context, userID uint, resource, action string) (bool, error) {
	start := time.Now()
	defer func() {
		rbacDuration.WithLabelValues("check_resource_access").Observe(time.Since(start).Seconds())
	}()
	
	permissions, err := s.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	
	for _, perm := range permissions {
		if perm.Resource == resource && (perm.Action == action || perm.Action == "*") {
			return true, nil
		}
	}
	
	return false, nil
}

// ResourceAccessMiddleware 资源访问权限中间件
func (s *RBACService) ResourceAccessMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户未认证",
			})
			c.Abort()
			return
		}
		
		accessible, err := s.IsResourceAccessible(c.Request.Context(), userID.(uint), resource, action)
		if err != nil {
			rbacErrors.WithLabelValues("resource_access_failed").Inc()
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "权限检查失败",
			})
			c.Abort()
			return
		}
		
		if !accessible {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}