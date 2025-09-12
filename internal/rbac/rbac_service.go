package rbac

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
	"law-oa-go/internal/cache"
	"law-oa-go/internal/models"
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

// RBACService RBAC服务
type RBACService struct {
	db     *gorm.DB
	cache  *cache.LayeredCache
	logger interface{}
}

// NewRBACService 创建RBAC服务
func NewRBACService(db *gorm.DB, cache *cache.LayeredCache) *RBACService {
	return &RBACService{
		db:    db,
		cache: cache,
	}
}

// CheckUserPermission 检查用户权限
func (s *RBACService) CheckUserPermission(ctx context.Context, userID uint, permissionKey string) (bool, error) {
	timer := prometheus.NewTimer(rbacDuration.WithLabelValues("check_user_permission"))
	defer timer.ObserveDuration()

	// 查询用户权限（简化版，基于用户角色）
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		rbacErrors.WithLabelValues("check_user_permission").Inc()
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// 简化版：基于用户角色检查权限
	hasPermission := s.checkRolePermission(user.Role, permissionKey)
	if hasPermission {
		// 缓存权限
		permissions := s.getRolePermissions(user.Role)
		cacheKey := fmt.Sprintf("user_permissions:%d", userID)
		s.cache.Set(ctx, cacheKey, permissions)
		return true, nil
	}
	return false, nil
}

// checkRolePermission 检查角色权限
func (s *RBACService) checkRolePermission(role, permissionKey string) bool {
	rolePermissions := map[string][]string{
		"admin": {"user:view", "user:create", "user:update", "user:delete", "user:assign_role",
			"role:view", "role:create", "role:update", "role:delete", "role:assign_permission",
			"permission:view", "permission:create", "permission:update", "permission:delete",
			"case:view", "case:create", "case:update", "case:delete", "case:assign",
			"client:view", "client:create", "client:update", "client:delete",
			"document:view", "document:upload", "document:update", "document:delete", "document:download"},
		"lawyer": {"case:view", "case:create", "case:update", "case:assign",
			"client:view", "client:create", "client:update",
			"document:view", "document:upload", "document:update", "document:download"},
		"user": {"case:view", "client:view", "document:view", "document:download"},
	}

	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, perm := range permissions {
		if perm == permissionKey {
			return true
		}
	}
	return false
}

// getRolePermissions 获取角色权限列表
func (s *RBACService) getRolePermissions(role string) []string {
	rolePermissions := map[string][]string{
		"admin": {"user:view", "user:create", "user:update", "user:delete", "user:assign_role",
			"role:view", "role:create", "role:update", "role:delete", "role:assign_permission",
			"permission:view", "permission:create", "permission:update", "permission:delete",
			"case:view", "case:create", "case:update", "case:delete", "case:assign",
			"client:view", "client:create", "client:update", "client:delete",
			"document:view", "document:upload", "document:update", "document:delete", "document:download"},
		"lawyer": {"case:view", "case:create", "case:update", "case:assign",
			"client:view", "client:create", "client:update",
			"document:view", "document:upload", "document:update", "document:download"},
		"user": {"case:view", "client:view", "document:view", "document:download"},
	}

	return rolePermissions[role]
}

// RequirePermission 权限验证中间件
func (s *RBACService) RequirePermission(permissionKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未授权访问",
			})
			c.Abort()
			return
		}

		hasPermission, err := s.CheckUserPermission(c.Request.Context(), userID.(uint), permissionKey)
		if err != nil || !hasPermission {
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
