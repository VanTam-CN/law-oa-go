package services

import (
	"time"

	"github.com/law-oa-go/document-service/internal/auth"
	"github.com/law-oa-go/document-service/internal/repositories"
)

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string                     `json:"name" validate:"required,min=1,max=100"`
	DisplayName string                     `json:"display_name" validate:"required,min=1,max=100"`
	Description string                     `json:"description" validate:"max=500"`
	Level       int                        `json:"level" validate:"min=0,max=1000"`
	Type        string                     `json:"type" validate:"required,oneof=system tenant custom"`
	Permissions []string                   `json:"permissions"`
	Constraints []auth.RoleConstraint       `json:"constraints"`
	Inherits    []string                   `json:"inherits"`
	TenantID    string                     `json:"tenant_id" validate:"required"`
	Enabled     bool                       `json:"enabled"`
	CreatedBy   uint                       `json:"created_by" validate:"required"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	DisplayName string                     `json:"display_name" validate:"omitempty,min=1,max=100"`
	Description string                     `json:"description" validate:"omitempty,max=500"`
	Level       *int                       `json:"level" validate:"omitempty,min=0,max=1000"`
	Enabled     *bool                      `json:"enabled"`
	Permissions []string                   `json:"permissions"`
	Constraints []auth.RoleConstraint       `json:"constraints"`
	Inherits    []string                   `json:"inherits"`
}

// RoleFilter 角色过滤器
type RoleFilter struct {
	TenantID    string                      `json:"tenant_id" validate:"required"`
	Name        string                      `json:"name"`
	Type        string                      `json:"type"`
	Enabled     *bool                       `json:"enabled"`
	CreatorID   *uint                       `json:"creator_id"`
	CreatedFrom *time.Time                  `json:"created_from"`
	CreatedTo   *time.Time                  `json:"created_to"`
	Pagination  *repositories.Pagination    `json:"pagination"`
	SortBy      string                      `json:"sort_by"`
	SortOrder   string                      `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	DisplayName string               `json:"display_name"`
	Description string               `json:"description"`
	Level       int                  `json:"level"`
	Type        string               `json:"type"`
	Permissions []string             `json:"permissions"`
	Constraints []auth.RoleConstraint `json:"constraints"`
	Inherits    []string             `json:"inherits"`
	TenantID    string               `json:"tenant_id"`
	Enabled     bool                 `json:"enabled"`
	CreatedBy   uint                 `json:"created_by"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Roles    []*RoleResponse `json:"roles"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID     uint                   `json:"user_id" validate:"required"`
	RoleID     uint                   `json:"role_id" validate:"required"`
	TenantID   string                 `json:"tenant_id" validate:"required"`
	Attributes map[string]interface{} `json:"attributes"`
	ExpiresAt  *time.Time             `json:"expires_at"`
}

// RevokeRoleRequest 撤销角色请求
type RevokeRoleRequest struct {
	UserID   uint   `json:"user_id" validate:"required"`
	RoleID   uint   `json:"role_id" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// UserRoleResponse 用户角色响应
type UserRoleResponse struct {
	RoleID      uint      `json:"role_id"`
	RoleName    string    `json:"role_name"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	Level       int       `json:"level"`
	TenantID    string    `json:"tenant_id"`
	Enabled     bool      `json:"enabled"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserRoleListResponse 用户角色列表响应
type UserRoleListResponse struct {
	UserID uint                `json:"user_id"`
	Roles  []*UserRoleResponse `json:"roles"`
	Total  int64               `json:"total"`
}

// RoleUserResponse 角色用户响应
type RoleUserResponse struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// RoleUserListResponse 角色用户列表响应
type RoleUserListResponse struct {
	RoleID uint               `json:"role_id"`
	Users  []*RoleUserResponse `json:"users"`
	Total  int64              `json:"total"`
}

// CheckPermissionRequest 检查权限请求
type CheckPermissionRequest struct {
	UserID    uint                   `json:"user_id" validate:"required"`
	Username  string                 `json:"username" validate:"required"`
	TenantID  string                 `json:"tenant_id" validate:"required"`
	Resource  string                 `json:"resource" validate:"required"`
	Action    string                 `json:"action" validate:"required"`
	Context   map[string]interface{} `json:"context"`
	RequestID string                 `json:"request_id"`
}

// CheckPermissionResponse 检查权限响应
type CheckPermissionResponse struct {
	Allowed     bool                      `json:"allowed"`
	Reason      string                    `json:"reason"`
	Roles       []string                  `json:"roles"`
	Permissions []string                  `json:"permissions"`
	Duration    time.Duration             `json:"duration"`
	Constraints []auth.AppliedConstraint  `json:"constraints"`
	Attributes  map[string]interface{}    `json:"attributes"`
}

// UserPermissionsResponse 用户权限响应
type UserPermissionsResponse struct {
	UserID      uint                        `json:"user_id"`
	TenantID    string                      `json:"tenant_id"`
	Roles       map[string]*RolePermissionDetail `json:"roles"`
	Permissions []string                    `json:"permissions"`
	TotalRoles  int                         `json:"total_roles"`
	TotalPerms  int                         `json:"total_perms"`
	CheckedAt   time.Time                   `json:"checked_at"`
}

// RolePermissionDetail 角色权限详情
type RolePermissionDetail struct {
	RoleID      uint      `json:"role_id"`
	RoleName    string    `json:"role_name"`
	DisplayName string    `json:"display_name"`
	Type        string    `json:"type"`
	Level       int       `json:"level"`
	Permissions []string  `json:"permissions"`
	Enabled     bool      `json:"enabled"`
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	Resource    string `json:"resource" validate:"required"`
	Action      string `json:"action" validate:"required"`
	Scope       string `json:"scope" validate:"required,oneof=global tenant resource"`
	Category    string `json:"category"`
	System      bool   `json:"system"`
	Enabled     bool   `json:"enabled"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	DisplayName string `json:"display_name" validate:"omitempty,min=1,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Scope       string `json:"scope" validate:"omitempty,oneof=global tenant resource"`
	Category    string `json:"category"`
	Enabled     *bool  `json:"enabled"`
}

// PermissionFilter 权限过滤器
type PermissionFilter struct {
	Name       string                     `json:"name"`
	Resource   string                     `json:"resource"`
	Action     string                     `json:"action"`
	Scope      string                     `json:"scope"`
	Category   string                     `json:"category"`
	System     *bool                      `json:"system"`
	Enabled    *bool                      `json:"enabled"`
	Pagination *repositories.Pagination   `json:"pagination"`
	SortBy     string                     `json:"sort_by"`
	SortOrder  string                     `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// PermissionResponse 权限响应
type PermissionResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Scope       string    `json:"scope"`
	Category    string    `json:"category"`
	System      bool      `json:"system"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Permissions []*PermissionResponse `json:"permissions"`
	Total       int64                 `json:"total"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
}

// RolePermissionRequest 角色权限请求
type RolePermissionRequest struct {
	RoleID         uint   `json:"role_id" validate:"required"`
	PermissionName string `json:"permission_name" validate:"required"`
}

// RoleInheritanceRequest 角色继承请求
type RoleInheritanceRequest struct {
	RoleID       uint   `json:"role_id" validate:"required"`
	ParentRoleID uint   `json:"parent_role_id" validate:"required"`
	TenantID     string `json:"tenant_id" validate:"required"`
}

// RoleInheritanceResponse 角色继承响应
type RoleInheritanceResponse struct {
	RoleID         uint                      `json:"role_id"`
	RoleName       string                    `json:"role_name"`
	ParentRoles    []*RoleInheritanceDetail   `json:"parent_roles"`
	ChildRoles     []*RoleInheritanceDetail   `json:"child_roles"`
	InheritedRoles []*RoleInheritanceDetail   `json:"inherited_roles"`
	TotalInherited int                       `json:"total_inherited"`
}

// RoleInheritanceDetail 角色继承详情
type RoleInheritanceDetail struct {
	RoleID      uint   `json:"role_id"`
	RoleName    string `json:"role_name"`
	DisplayName string `json:"display_name"`
	Level       int    `json:"level"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
}

// CreateRoleTemplateRequest 创建角色模板请求
type CreateRoleTemplateRequest struct {
	Name        string                     `json:"name" validate:"required,min=1,max=100"`
	Description string                     `json:"description" validate:"max=500"`
	Category    string                     `json:"category" validate:"required"`
	Type        string                     `json:"type" validate:"required,oneof=system tenant custom"`
	Level       int                        `json:"level" validate:"min=0,max=1000"`
	Permissions []string                   `json:"permissions"`
	Constraints []auth.RoleConstraint       `json:"constraints"`
	Inherits    []string                   `json:"inherits"`
	Public      bool                       `json:"public"`
	Enabled     bool                       `json:"enabled"`
	CreatedBy   uint                       `json:"created_by" validate:"required"`
}

// RoleTemplateFilter 角色模板过滤器
type RoleTemplateFilter struct {
	Category   string                     `json:"category"`
	Name       string                     `json:"name"`
	Type       string                     `json:"type"`
	Public     *bool                      `json:"public"`
	Enabled    *bool                      `json:"enabled"`
	Tags       []string                   `json:"tags"`
	Pagination *repositories.Pagination   `json:"pagination"`
	SortBy     string                     `json:"sort_by"`
	SortOrder  string                     `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// RoleTemplateResponse 角色模板响应
type RoleTemplateResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Type        string               `json:"type"`
	Level       int                  `json:"level"`
	Permissions []string             `json:"permissions"`
	Constraints []auth.RoleConstraint `json:"constraints"`
	Inherits    []string             `json:"inherits"`
	Tags        []string             `json:"tags"`
	Public      bool                 `json:"public"`
	Enabled     bool                 `json:"enabled"`
	CreatedBy   uint                 `json:"created_by"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// RoleTemplateListResponse 角色模板列表响应
type RoleTemplateListResponse struct {
	Templates []*RoleTemplateResponse `json:"templates"`
	Total     int64                   `json:"total"`
	Page      int                     `json:"page"`
	PageSize  int                     `json:"page_size"`
}

// CreateRoleFromTemplateRequest 从模板创建角色请求
type CreateRoleFromTemplateRequest struct {
	TemplateID   uint                        `json:"template_id" validate:"required"`
	Name         string                      `json:"name" validate:"required,min=1,max=100"`
	DisplayName  string                      `json:"display_name" validate:"required,min=1,max=100"`
	Description  string                      `json:"description"`
	TenantID     string                      `json:"tenant_id" validate:"required"`
	Parameters   map[string]interface{}      `json:"parameters"`
	Adjustments  *RoleTemplateAdjustments    `json:"adjustments"`
	CreatedBy    uint                        `json:"created_by" validate:"required"`
}

// RoleTemplateAdjustments 角色模板调整
type RoleTemplateAdjustments struct {
	AddPermissions    []string             `json:"add_permissions"`
	RemovePermissions []string             `json:"remove_permissions"`
	AddConstraints     []auth.RoleConstraint `json:"add_constraints"`
	RemoveConstraints  []string             `json:"remove_constraints"`
	SetLevel          *int                 `json:"set_level"`
}

// RoleStatistics 角色统计
type RoleStatistics struct {
	TotalRoles       int64                      `json:"total_roles"`
	EnabledRoles     int64                      `json:"enabled_roles"`
	DisabledRoles    int64                      `json:"disabled_roles"`
	RolesByType      map[string]int64           `json:"roles_by_type"`
	RolesByLevel     map[string]int64           `json:"roles_by_level"`
	RolesByCategory  map[string]int64           `json:"roles_by_category"`
	RoleAssignments  int64                      `json:"role_assignments"`
	UsersWithRoles   int64                      `json:"users_with_roles"`
	AverageRolesPerUser float64                  `json:"average_roles_per_user"`
	TopRoles         []RoleUsageStat            `json:"top_roles"`
	GeneratedAt      time.Time                  `json:"generated_at"`
}

// RoleUsageStat 角色使用统计
type RoleUsageStat struct {
	RoleID      uint   `json:"role_id"`
	RoleName    string `json:"role_name"`
	DisplayName string `json:"display_name"`
	UserCount   int64  `json:"user_count"`
	Type        string `json:"type"`
	Level       int    `json:"level"`
}

// RoleUsageAnalysis 角色使用分析
type RoleUsageAnalysis struct {
	Summary           RoleUsageSummary            `json:"summary"`
	RoleUsageDetails  []RoleUsageDetail           `json:"role_usage_details"`
	UserDistribution  []UserRoleDistribution       `json:"user_distribution"`
	TemporalAnalysis  []RoleUsageTemporalStat     `json:"temporal_analysis"`
	Recommendations   []string                    `json:"recommendations"`
	AnalyzedAt        time.Time                   `json:"analyzed_at"`
}

// RoleUsageSummary 角色使用摘要
type RoleUsageSummary struct {
	TotalRoles            int64   `json:"total_roles"`
	ActiveRoles           int64   `json:"active_roles"`
	TotalAssignments      int64   `json:"total_assignments"`
	UniqueUsers           int64   `json:"unique_users"`
	AverageRolesPerUser   float64 `json:"average_roles_per_user"`
	MostUsedRole         string  `json:"most_used_role"`
	LeastUsedRole        string  `json:"least_used_role"`
	OrphanedRoles        int64   `json:"orphaned_roles"`
	OverprivilegedUsers   int64   `json:"overprivileged_users"`
	UnderprivilegedUsers int64   `json:"underprivileged_users"`
}

// RoleUsageDetail 角色使用详情
type RoleUsageDetail struct {
	RoleID          uint      `json:"role_id"`
	RoleName        string    `json:"role_name"`
	DisplayName     string    `json:"display_name"`
	Type            string    `json:"type"`
	Level           int       `json:"level"`
	UserCount       int64     `json:"user_count"`
	ActiveUsers     int64     `json:"active_users"`
	InactiveUsers   int64     `json:"inactive_users"`
	ExpiringSoon    int64     `json:"expiring_soon"`
	Expired         int64     `json:"expired"`
	LastAssigned    time.Time `json:"last_assigned"`
	AssignmentRate  float64   `json:"assignment_rate"`
	UsageFrequency  float64   `json:"usage_frequency"`
}

// UserRoleDistribution 用户角色分布
type UserRoleDistribution struct {
	UserID        uint      `json:"user_id"`
	Username      string    `json:"username"`
	RoleCount     int       `json:"role_count"`
	Roles         []string  `json:"roles"`
	LastLogin     time.Time `json:"last_login"`
	ActiveAssignments int   `json:"active_assignments"`
	ExpiredAssignments int   `json:"expired_assignments"`
}

// RoleUsageTemporalStat 角色使用时间统计
type RoleUsageTemporalStat struct {
	Date           time.Time `json:"date"`
	NewAssignments int64     `json:"new_assignments"`
	ExpiredAssignments int64  `json:"expired_assignments"`
	RevokedAssignments int64  `json:"revoked_assignments"`
	ActiveAssignments int64   `json:"active_assignments"`
	TotalAssignments int64    `json:"total_assignments"`
}

// PermissionUsageAnalysis 权限使用分析
type PermissionUsageAnalysis struct {
	Summary                PermissionUsageSummary       `json:"summary"`
	PermissionUsageDetails  []PermissionUsageDetail      `json:"permission_usage_details"`
	RolePermissionMatrix    []RolePermissionMatrixEntry  `json:"role_permission_matrix"`
	ResourceUsage          []ResourceUsageStat          `json:"resource_usage"`
	ActionUsage            []ActionUsageStat            `json:"action_usage"`
	Recommendations        []string                     `json:"recommendations"`
	AnalyzedAt             time.Time                    `json:"analyzed_at"`
}

// PermissionUsageSummary 权限使用摘要
type PermissionUsageSummary struct {
	TotalPermissions       int64   `json:"total_permissions"`
	ActivePermissions      int64   `json:"active_permissions"`
	TotalRolePermissions   int64   `json:"total_role_permissions"`
	UniqueResourceActions  int64   `json:"unique_resource_actions"`
	MostUsedPermission    string  `json:"most_used_permission"`
	LeastUsedPermission   string  `json:"least_used_permission"`
	UnusedPermissions     int64   `json:"unused_permissions"`
	OverlappingRoles      int64   `json:"overlapping_roles"`
	CriticalPermissions   int64   `json:"critical_permissions"`
}

// PermissionUsageDetail 权限使用详情
type PermissionUsageDetail struct {
	PermissionID    uint    `json:"permission_id"`
	PermissionName  string  `json:"permission_name"`
	DisplayName     string  `json:"display_name"`
	Resource        string  `json:"resource"`
	Action          string  `json:"action"`
	Scope           string  `json:"scope"`
	Category        string  `json:"category"`
	RoleCount       int64   `json:"role_count"`
	System          bool    `json:"system"`
	Enabled         bool    `json:"enabled"`
	UsageCount      int64   `json:"usage_count"`
	UsageFrequency  float64 `json:"usage_frequency"`
	LastUsed        time.Time `json:"last_used"`
	RiskLevel       string  `json:"risk_level"`
}

// RolePermissionMatrixEntry 角色权限矩阵条目
type RolePermissionMatrixEntry struct {
	RoleID         uint   `json:"role_id"`
	RoleName       string `json:"role_name"`
	PermissionID   uint   `json:"permission_id"`
	PermissionName string `json:"permission_name"`
	Resource       string `json:"resource"`
	Action         string `json:"action"`
	Direct         bool   `json:"direct"`
	Inherited      bool   `json:"inherited"`
}

// ResourceUsageStat 资源使用统计
type ResourceUsageStat struct {
	Resource     string  `json:"resource"`
	PermissionCount int64 `json:"permission_count"`
	RoleCount       int64 `json:"role_count"`
	UsageCount      int64 `json:"usage_count"`
	AverageRisk     float64 `json:"average_risk"`
}

// ActionUsageStat 动作使用统计
type ActionUsageStat struct {
	Action          string  `json:"action"`
	PermissionCount int64   `json:"permission_count"`
	RoleCount       int64   `json:"role_count"`
	UsageCount      int64   `json:"usage_count"`
	AverageRisk     float64 `json:"average_risk"`
}

// BulkAssignRolesRequest 批量分配角色请求
type BulkAssignRolesRequest struct {
	UserIDs   []uint                     `json:"user_ids" validate:"required,min=1"`
	RoleID    uint                       `json:"role_id" validate:"required"`
	TenantID  string                     `json:"tenant_id" validate:"required"`
	Attributes map[string]interface{}     `json:"attributes"`
	ExpiresAt *time.Time                 `json:"expires_at"`
}

// BulkRevokeRolesRequest 批量撤销角色请求
type BulkRevokeRolesRequest struct {
	UserIDs  []uint `json:"user_ids" validate:"required,min=1"`
	RoleID   uint   `json:"role_id" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required"`
}

// BulkCreateRolesRequest 批量创建角色请求
type BulkCreateRolesRequest struct {
	Roles []CreateRoleRequest `json:"roles" validate:"required,min=1,dive"`
}

// BulkOperationResult 批量操作结果
type BulkOperationResult struct {
	Total      int                    `json:"total"`
	Success    int                    `json:"success"`
	Failed     int                    `json:"failed"`
	Errors     []BulkOperationError   `json:"errors"`
	Succeeded  []BulkOperationSuccess `json:"succeeded"`
	ProcessedAt time.Time             `json:"processed_at"`
	Duration   time.Duration          `json:"duration"`
}

// BulkOperationError 批量操作错误
type BulkOperationError struct {
	Index   int    `json:"index"`
	Item    interface{} `json:"item"`
	Error   string `json:"error"`
	Code    string `json:"code"`
}

// BulkOperationSuccess 批量操作成功
type BulkOperationSuccess struct {
	Index int         `json:"index"`
	Item  interface{} `json:"item"`
	ID    uint        `json:"id,omitempty"`
}