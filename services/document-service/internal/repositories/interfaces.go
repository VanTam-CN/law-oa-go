package repositories

import (
	"context"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"gorm.io/gorm"
)

// DocumentRepository 文档仓库接口
type DocumentRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, document *models.Document) error
	GetByID(ctx context.Context, id uint) (*models.Document, error)
	GetByUUID(ctx context.Context, uuid string) (*models.Document, error)
	Update(ctx context.Context, document *models.Document) error
	Delete(ctx context.Context, id uint) error
	SoftDelete(ctx context.Context, id uint) error

	// 查询操作
	List(ctx context.Context, filter *DocumentFilter) ([]*models.Document, int64, error)
	FindByEntity(ctx context.Context, entityType string, entityID uint) ([]*models.Document, error)
	FindByCategory(ctx context.Context, tenantID, category string) ([]*models.Document, error)
	FindByCreator(ctx context.Context, tenantID string, creatorID uint) ([]*models.Document, error)

	// 版本管理
	CreateVersion(ctx context.Context, version *models.DocumentVersion) error
	GetVersions(ctx context.Context, documentID uint) ([]*models.DocumentVersion, error)
	GetLatestVersion(ctx context.Context, documentID uint) (*models.DocumentVersion, error)
	GetVersionByNumber(ctx context.Context, documentID uint, version int) (*models.DocumentVersion, error)

	// 搜索和过滤
	SearchByName(ctx context.Context, tenantID, name string) ([]*models.Document, error)
	SearchByContent(ctx context.Context, tenantID, content string) ([]*models.Document, error)
	FindByTags(ctx context.Context, tenantID string, tags []string) ([]*models.Document, error)
	FindByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]*models.Document, error)

	// 统计操作
	CountByTenant(ctx context.Context, tenantID string) (int64, error)
	CountByCategory(ctx context.Context, tenantID, category string) (int64, error)
	GetSizeByTenant(ctx context.Context, tenantID string) (int64, error)
	GetRecentDocuments(ctx context.Context, tenantID string, limit int) ([]*models.Document, error)

	// 批量操作
	BatchCreate(ctx context.Context, documents []*models.Document) error
	BatchUpdate(ctx context.Context, documents []*models.Document) error
	BatchDelete(ctx context.Context, ids []uint) error
}

// DocumentVersionRepository 文档版本仓库接口
type DocumentVersionRepository interface {
	Create(ctx context.Context, version *models.DocumentVersion) error
	GetByID(ctx context.Context, id uint) (*models.DocumentVersion, error)
	GetByUUID(ctx context.Context, uuid string) (*models.DocumentVersion, error)
	GetByDocumentID(ctx context.Context, documentID uint) ([]*models.DocumentVersion, error)
	GetLatest(ctx context.Context, documentID uint) (*models.DocumentVersion, error)
	Delete(ctx context.Context, id uint) error
	DeleteByDocumentID(ctx context.Context, documentID uint) error
}

// DocumentPermissionRepository 文档权限仓库接口
type DocumentPermissionRepository interface {
	Create(ctx context.Context, permission *models.DocumentPermission) error
	GetByID(ctx context.Context, id uint) (*models.DocumentPermission, error)
	Update(ctx context.Context, permission *models.DocumentPermission) error
	Delete(ctx context.Context, id uint) error

	// 查询操作
	FindByDocument(ctx context.Context, documentID uint) ([]*models.DocumentPermission, error)
	FindByUser(ctx context.Context, userID uint) ([]*models.DocumentPermission, error)
	FindByRole(ctx context.Context, roleID uint) ([]*models.DocumentPermission, error)
	FindByTenant(ctx context.Context, tenantID string) ([]*models.DocumentPermission, error)

	// 权限检查
	CheckUserPermission(ctx context.Context, documentID, userID uint, permission string) (bool, error)
	CheckRolePermission(ctx context.Context, documentID, roleID uint, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, documentID, userID uint) ([]string, error)
	GetRolePermissions(ctx context.Context, documentID, roleID uint) ([]string, error)

	// 批量操作
	BatchCreate(ctx context.Context, permissions []*models.DocumentPermission) error
	BatchDeleteByDocument(ctx context.Context, documentID uint) error
	BatchDeleteByUser(ctx context.Context, userID uint) error
}

// DocumentAuditRepository 文档审计仓库接口
type DocumentAuditRepository interface {
	Create(ctx context.Context, audit *models.DocumentAudit) error
	GetByID(ctx context.Context, id uint) (*models.DocumentAudit, error)
	GetByDocumentID(ctx context.Context, documentID uint, limit int) ([]*models.DocumentAudit, error)
	GetByUserID(ctx context.Context, userID uint, limit int) ([]*models.DocumentAudit, error)
	GetByTenant(ctx context.Context, tenantID string, filter *AuditFilter) ([]*models.DocumentAudit, error)
	GetByAction(ctx context.Context, tenantID string, action string, limit int) ([]*models.DocumentAudit, error)
	GetByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]*models.DocumentAudit, error)

	// 清理操作
	DeleteOldRecords(ctx context.Context, beforeDate time.Time) error
	CountByTenant(ctx context.Context, tenantID string) (int64, error)
}

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uint) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error

	// 查询操作
	List(ctx context.Context, tenantID string, filter *UserFilter) ([]*models.User, int64, error)
	FindByTenant(ctx context.Context, tenantID string) ([]*models.User, error)
	SearchByName(ctx context.Context, tenantID, name string) ([]*models.User, error)

	// 批量操作
	BatchCreate(ctx context.Context, users []*models.User) error
}

// RoleRepository 角色仓库接口
type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	GetByID(ctx context.Context, id uint) (*models.Role, error)
	GetByName(ctx context.Context, name, tenantID string) (*models.Role, error)
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id uint) error

	// 查询操作
	List(ctx context.Context, tenantID string) ([]*models.Role, error)
	FindByTenant(ctx context.Context, tenantID string) ([]*models.Role, error)
	GetDefaultRoles(ctx context.Context, tenantID string) ([]*models.Role, error)

	// 批量操作
	BatchCreate(ctx context.Context, roles []*models.Role) error
}

// UserRoleRepository 用户角色关联仓库接口
type UserRoleRepository interface {
	Create(ctx context.Context, userRole *models.UserRole) error
	Delete(ctx context.Context, userID, roleID uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
	DeleteByRoleID(ctx context.Context, roleID uint) error

	// 查询操作
	GetByUserID(ctx context.Context, userID uint) ([]*models.UserRole, error)
	GetByRoleID(ctx context.Context, roleID uint) ([]*models.UserRole, error)
	GetUserRoles(ctx context.Context, userID uint) ([]*models.Role, error)
	GetRoleUsers(ctx context.Context, roleID uint) ([]*models.User, error)
	CheckUserRole(ctx context.Context, userID, roleID uint) (bool, error)

	// 批量操作
	BatchCreate(ctx context.Context, userRoles []*models.UserRole) error
	AssignRolesToUser(ctx context.Context, userID uint, roleIDs []uint) error
	RemoveRolesFromUser(ctx context.Context, userID uint, roleIDs []uint) error
}

// SearchRepository 搜索仓库接口
type SearchRepository interface {
	// 索引操作
	IndexDocument(ctx context.Context, document *models.Document) error
	UpdateDocument(ctx context.Context, document *models.Document) error
	DeleteDocument(ctx context.Context, documentUUID string) error
	BulkIndexDocuments(ctx context.Context, documents []*models.Document) error

	// 搜索操作
	SearchDocuments(ctx context.Context, query *SearchQuery) (*SearchResult, error)
	SuggestDocuments(ctx context.Context, query string, limit int) ([]*Suggestion, error)
	GetSimilarDocuments(ctx context.Context, documentUUID string, limit int) ([]*models.Document, error)

	// 聚合搜索
	SearchByCategory(ctx context.Context, tenantID string) (map[string]int64, error)
	SearchByTags(ctx context.Context, tenantID string) (map[string]int64, error)
	SearchByCreator(ctx context.Context, tenantID string) (map[string]int64, error)

	// 索引管理
	CreateIndex(ctx context.Context) error
	DeleteIndex(ctx context.Context) error
	ReindexAllDocuments(ctx context.Context) error
	GetIndexStats(ctx context.Context) (map[string]interface{}, error)
}

// Filter 结构体
type DocumentFilter struct {
	TenantID   string    `json:"tenant_id"`
	Category   string    `json:"category"`
	Status     string    `json:"status"`
	CreatedBy  uint      `json:"created_by"`
	Tags       []string  `json:"tags"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	EntityType string    `json:"entity_type"`
	EntityID   uint      `json:"entity_id"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	SortBy     string    `json:"sort_by"`
	SortOrder  string    `json:"sort_order"`
}

type UserFilter struct {
	TenantID string `json:"tenant_id"`
	IsActive *bool  `json:"is_active"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	SortBy   string `json:"sort_by"`
	SortOrder string `json:"sort_order"`
}

type AuditFilter struct {
	UserID    uint      `json:"user_id"`
	Action    string    `json:"action"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
}

type SearchQuery struct {
	TenantID     string            `json:"tenant_id"`
	Query        string            `json:"query"`
	Category     string            `json:"category"`
	Tags         []string          `json:"tags"`
	Creator      uint              `json:"creator"`
	StartDate    time.Time         `json:"start_date"`
	EndDate      time.Time         `json:"end_date"`
	Page         int               `json:"page"`
	PageSize     int               `json:"page_size"`
	SortBy       string            `json:"sort_by"`
	SortOrder    string            `json:"sort_order"`
	Aggregations map[string]string  `json:"aggregations"`
	Filters      map[string]interface{} `json:"filters"`
}

type SearchResult struct {
	Documents    []*models.Document `json:"documents"`
	Total        int64              `json:"total"`
	Page         int                `json:"page"`
	PageSize     int                `json:"page_size"`
	TotalPages   int                `json:"total_pages"`
	Aggregations map[string]interface{} `json:"aggregations"`
	Suggestions  []*Suggestion      `json:"suggestions"`
	Took         int                `json:"took"`
}

type Suggestion struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
	Type  string  `json:"type"`
}

// TransactionManager 事务管理器接口
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error
	BeginTx(ctx context.Context) (*gorm.DB, error)
}

// RepositoryManager 仓库管理器接口
type RepositoryManager interface {
	// 获取各种仓库
	Documents() DocumentRepository
	Versions() DocumentVersionRepository
	Permissions() DocumentPermissionRepository
	Audits() DocumentAuditRepository
	Users() UserRepository
	Roles() RoleRepository
	UserRoles() UserRoleRepository
	Search() SearchRepository

	// 事务管理
	TransactionManager() TransactionManager

	// 健康检查
	Health(ctx context.Context) error

	// 关闭连接
	Close() error
}