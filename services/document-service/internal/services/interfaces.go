package services

import (
	"context"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
)

// DocumentService 文档服务接口
type DocumentService interface {
	// 基础CRUD操作
	CreateDocument(ctx context.Context, req *CreateDocumentRequest) (*DocumentResponse, error)
	GetDocument(ctx context.Context, documentID string) (*DocumentResponse, error)
	GetDocumentByID(ctx context.Context, documentID uint) (*DocumentResponse, error)
	UpdateDocument(ctx context.Context, documentID string, req *UpdateDocumentRequest) (*DocumentResponse, error)
	DeleteDocument(ctx context.Context, documentID string) error

	// 文档查询
	ListDocuments(ctx context.Context, filter *DocumentFilter) (*DocumentListResponse, error)
	SearchDocuments(ctx context.Context, query *SearchRequest) (*SearchResponse, error)
	GetRecentDocuments(ctx context.Context, tenantID string, limit int) ([]*DocumentResponse, error)

	// 文档版本管理
	CreateVersion(ctx context.Context, documentID string, req *CreateVersionRequest) (*DocumentVersionResponse, error)
	GetVersions(ctx context.Context, documentID string) ([]*DocumentVersionResponse, error)
	GetLatestVersion(ctx context.Context, documentID string) (*DocumentVersionResponse, error)
	RestoreVersion(ctx context.Context, documentID string, version int) (*DocumentResponse, error)

	// 文档操作
	UploadDocument(ctx context.Context, req *UploadDocumentRequest) (*DocumentResponse, error)
	DownloadDocument(ctx context.Context, documentID string, version int) ([]byte, *DownloadMetadata, error)
	CopyDocument(ctx context.Context, documentID string, req *CopyDocumentRequest) (*DocumentResponse, error)
	MoveDocument(ctx context.Context, documentID string, req *MoveDocumentRequest) error

	// 批量操作
	BatchCreateDocuments(ctx context.Context, reqs []*CreateDocumentRequest) ([]*DocumentResponse, error)
	BatchUpdateDocuments(ctx context.Context, reqs []*UpdateDocumentRequest) ([]*DocumentResponse, error)
	BatchDeleteDocuments(ctx context.Context, documentIDs []string) error

	// 文档统计
	GetDocumentStats(ctx context.Context, tenantID string) (*DocumentStatsResponse, error)
	GetTenantStats(ctx context.Context, tenantID string) (*TenantStatsResponse, error)
}

// DocumentVersionService 文档版本服务接口
type DocumentVersionService interface {
	// 版本操作
	CreateVersion(ctx context.Context, req *CreateVersionRequest) (*DocumentVersionResponse, error)
	GetVersion(ctx context.Context, versionID string) (*DocumentVersionResponse, error)
	GetVersionByNumber(ctx context.Context, documentID string, version int) (*DocumentVersionResponse, error)
	DeleteVersion(ctx context.Context, versionID string) error
	CompareVersions(ctx context.Context, documentID string, version1, version2 int) (*VersionComparison, error)

	// 版本管理
	GetVersionHistory(ctx context.Context, documentID string, filter *VersionFilter) (*VersionListResponse, error)
	RollbackToVersion(ctx context.Context, documentID string, version int) (*DocumentResponse, error)
	CleanupOldVersions(ctx context.Context, documentID string, keepCount int) error

	// 版本分析
	GetVersionStats(ctx context.Context, documentID string) (*VersionStatsResponse, error)
	GetVersionDiff(ctx context.Context, documentID string, version1, version2 int) (*VersionDiff, error)
}

// PermissionService 权限服务接口
type PermissionService interface {
	// 权限管理
	GrantPermission(ctx context.Context, req *GrantPermissionRequest) error
	RevokePermission(ctx context.Context, req *RevokePermissionRequest) error
	UpdatePermission(ctx context.Context, req *UpdatePermissionRequest) error

	// 权限查询
	GetDocumentPermissions(ctx context.Context, documentID string) (*PermissionListResponse, error)
	GetUserPermissions(ctx context.Context, userID string, documentID string) ([]string, error)
	GetUserAccessibleDocuments(ctx context.Context, userID string, permission string) ([]*DocumentResponse, error)

	// 权限检查
	CheckPermission(ctx context.Context, userID, documentID string, permission string) (bool, error)
	CheckPermissionBatch(ctx context.Context, userID string, documentIDs []string, permission string) (map[string]bool, error)

	// 批量权限操作
	GrantPermissionsBatch(ctx context.Context, reqs []*GrantPermissionRequest) error
	RevokePermissionsBatch(ctx context.Context, reqs []*RevokePermissionRequest) error
}

// AuditService 审计服务接口
type AuditService interface {
	// 审计记录
	LogAction(ctx context.Context, req *LogActionRequest) error
	GetAuditLogs(ctx context.Context, filter *AuditFilter) (*AuditListResponse, error)
	GetDocumentHistory(ctx context.Context, documentID string, limit int) ([]*AuditResponse, error)

	// 审计分析
	GetActivityReport(ctx context.Context, tenantID string, filter *ActivityFilter) (*ActivityReport, error)
	GetUserActivity(ctx context.Context, userID string, startDate, endDate time.Time) (*UserActivityReport, error)

	// 审计清理
	CleanupOldLogs(ctx context.Context, beforeDate time.Time) error
	ExportAuditLogs(ctx context.Context, filter *AuditFilter) ([]byte, error)
}

// UserService 用户服务接口
type UserService interface {
	// 用户管理
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetUser(ctx context.Context, userID string) (*UserResponse, error)
	GetUserByUsername(ctx context.Context, username string) (*UserResponse, error)
	UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, userID string) error

	// 用户查询
	ListUsers(ctx context.Context, filter *UserFilter) (*UserListResponse, error)
	SearchUsers(ctx context.Context, query string, tenantID string) ([]*UserResponse, error)

	// 用户角色
	AssignRole(ctx context.Context, userID, roleID string) error
	RemoveRole(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]*RoleResponse, error)
}

// RoleService 角色服务接口
type RoleService interface {
	// 角色管理
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleResponse, error)
	GetRole(ctx context.Context, roleID string) (*RoleResponse, error)
	UpdateRole(ctx context.Context, roleID string, req *UpdateRoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, roleID string) error

	// 角色查询
	ListRoles(ctx context.Context, tenantID string) ([]*RoleResponse, error)
	GetRolePermissions(ctx context.Context, roleID string) ([]string, error)

	// 角色权限
	AssignPermission(ctx context.Context, roleID string, permission string) error
	RemovePermission(ctx context.Context, roleID string, permission string) error
	GetRoleUsers(ctx context.Context, roleID string) ([]*UserResponse, error)
}

// SearchService 搜索服务接口
type SearchService interface {
	// 搜索操作
	Search(ctx context.Context, query *SearchRequest) (*SearchResponse, error)
	Suggest(ctx context.Context, query string, tenantID string) ([]*Suggestion, error)
	GetSimilarDocuments(ctx context.Context, documentID string, limit int) ([]*DocumentResponse, error)

	// 搜索管理
	IndexDocument(ctx context.Context, documentID string) error
	ReindexDocument(ctx context.Context, documentID string) error
	DeleteFromIndex(ctx context.Context, documentID string) error

	// 搜索分析
	GetSearchAnalytics(ctx context.Context, tenantID string, filter *AnalyticsFilter) (*SearchAnalytics, error)
	GetPopularQueries(ctx context.Context, tenantID string, limit int) ([]*PopularQuery, error)
}

// StorageService 存储服务接口
type StorageService interface {
	// 文件操作
	UploadFile(ctx context.Context, req *UploadFileRequest) (*FileResponse, error)
	DownloadFile(ctx context.Context, fileID string) ([]byte, *FileMetadata, error)
	DeleteFile(ctx context.Context, fileID string) error
	CopyFile(ctx context.Context, sourceID, targetID string) error

	// 文件管理
	GetFileInfo(ctx context.Context, fileID string) (*FileResponse, error)
	ListFiles(ctx context.Context, filter *FileFilter) (*FileListResponse, error)
	GetFileURL(ctx context.Context, fileID string, expiry time.Duration) (string, error)

	// 存储分析
	GetStorageStats(ctx context.Context, tenantID string) (*StorageStats, error)
	GetStorageUsage(ctx context.Context, filter *UsageFilter) (*StorageUsage, error)
}

// NotificationService 通知服务接口
type NotificationService interface {
	// 通知发送
	SendNotification(ctx context.Context, req *SendNotificationRequest) error
	SendBatchNotifications(ctx context.Context, reqs []*SendNotificationRequest) error

	// 通知查询
	GetNotifications(ctx context.Context, userID string, filter *NotificationFilter) (*NotificationListResponse, error)
	MarkAsRead(ctx context.Context, notificationID string) error
	MarkAllAsRead(ctx context.Context, userID string) error

	// 通知设置
	GetNotificationSettings(ctx context.Context, userID string) (*NotificationSettings, error)
	UpdateNotificationSettings(ctx context.Context, userID string, settings *NotificationSettings) error
}

// ServiceManager 服务管理器接口
type ServiceManager interface {
	// 获取各种服务
	Documents() DocumentService
	Versions() DocumentVersionService
	Permissions() PermissionService
	Audits() AuditService
	Users() UserService
	Roles() RoleService
	Search() SearchService
	Storage() StorageService
	Notifications() NotificationService

	// 健康检查
	Health(ctx context.Context) (map[string]error, error)

	// 关闭服务
	Close() error
}

// Request/Response 结构体
type CreateDocumentRequest struct {
	Name         string            `json:"name" validate:"required,max=255"`
	Description  string            `json:"description" validate:"max=1000"`
	Category     string            `json:"category" validate:"max=50"`
	Tags         []string          `json:"tags"`
	EntityType   string            `json:"entity_type" validate:"max=50"`
	EntityID     uint              `json:"entity_id"`
	TenantID     string            `json:"tenant_id" validate:"required"`
	CreatedBy    uint              `json:"created_by" validate:"required"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type UpdateDocumentRequest struct {
	Name        string            `json:"name" validate:"max=255"`
	Description string            `json:"description" validate:"max=1000"`
	Category    string            `json:"category" validate:"max=50"`
	Tags        []string          `json:"tags"`
	EntityType  string            `json:"entity_type" validate:"max=50"`
	EntityID    uint              `json:"entity_id"`
	Metadata    map[string]interface{} `json:"metadata"`
}

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
	Page       int       `json:"page" validate:"min=1"`
	PageSize   int       `json:"page_size" validate:"min=1,max=100"`
	SortBy     string    `json:"sort_by"`
	SortOrder  string    `json:"sort_order" validate:"oneof=asc desc"`
}

type SearchRequest struct {
	TenantID     string            `json:"tenant_id" validate:"required"`
	Query        string            `json:"query"`
	Category     string            `json:"category"`
	Tags         []string          `json:"tags"`
	Creator      uint              `json:"creator"`
	StartDate    time.Time         `json:"start_date"`
	EndDate      time.Time         `json:"end_date"`
	Page         int               `json:"page" validate:"min=1"`
	PageSize     int               `json:"page_size" validate:"min=1,max=100"`
	SortBy       string            `json:"sort_by"`
	SortOrder    string            `json:"sort_order" validate:"oneof=asc desc"`
	Aggregations map[string]string  `json:"aggregations"`
	Filters      map[string]interface{} `json:"filters"`
}

type UploadDocumentRequest struct {
	Name         string            `json:"name" validate:"required,max=255"`
	Description  string            `json:"description" validate:"max=1000"`
	Category     string            `json:"category" validate:"max=50"`
	Tags         []string          `json:"tags"`
	EntityType   string            `json:"entity_type" validate:"max=50"`
	EntityID     uint              `json:"entity_id"`
	TenantID     string            `json:"tenant_id" validate:"required"`
	CreatedBy    uint              `json:"created_by" validate:"required"`
	File         []byte            `json:"file" validate:"required"`
	FileName     string            `json:"file_name" validate:"required"`
	MimeType     string            `json:"mime_type" validate:"required"`
	Size         int64             `json:"size" validate:"required"`
	FileHash     string            `json:"file_hash"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type CopyDocumentRequest struct {
	Name       string `json:"name" validate:"required,max=255"`
	Category   string `json:"category" validate:"max=50"`
	Tags       []string `json:"tags"`
	EntityType string `json:"entity_type" validate:"max=50"`
	EntityID   uint   `json:"entity_id"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type MoveDocumentRequest struct {
	EntityType string `json:"entity_type" validate:"max=50"`
	EntityID   uint   `json:"entity_id"`
}

type CreateVersionRequest struct {
	Description string `json:"description" validate:"max=1000"`
	File        []byte `json:"file" validate:"required"`
	FileName    string `json:"file_name" validate:"required"`
	MimeType    string `json:"mime_type" validate:"required"`
	Size        int64  `json:"size" validate:"required"`
	FileHash    string `json:"file_hash"`
	CreatedBy   uint   `json:"created_by" validate:"required"`
}

type VersionFilter struct {
	Page     int       `json:"page" validate:"min=1"`
	PageSize int       `json:"page_size" validate:"min=1,max=100"`
	SortBy   string    `json:"sort_by"`
	SortOrder string   `json:"sort_order" validate:"oneof=asc desc"`
}

type GrantPermissionRequest struct {
	DocumentID string `json:"document_id" validate:"required"`
	UserID     string `json:"user_id,omitempty"`
	RoleID     string `json:"role_id,omitempty"`
	Permission string `json:"permission" validate:"required,oneof=read write delete admin"`
	TenantID   string `json:"tenant_id" validate:"required"`
}

type RevokePermissionRequest struct {
	DocumentID string `json:"document_id" validate:"required"`
	UserID     string `json:"user_id,omitempty"`
	RoleID     string `json:"role_id,omitempty"`
	Permission string `json:"permission,omitempty"`
	TenantID   string `json:"tenant_id" validate:"required"`
}

type UpdatePermissionRequest struct {
	DocumentID string `json:"document_id" validate:"required"`
	UserID     string `json:"user_id,omitempty"`
	RoleID     string `json:"role_id,omitempty"`
	OldPermission string `json:"old_permission" validate:"required,oneof=read write delete admin"`
	NewPermission string `json:"new_permission" validate:"required,oneof=read write delete admin"`
	TenantID   string `json:"tenant_id" validate:"required"`
}

type LogActionRequest struct {
	DocumentID string `json:"document_id" validate:"required"`
	UserID     string `json:"user_id" validate:"required"`
	Action     string `json:"action" validate:"required,oneof=create read update delete download share"`
	Details    string `json:"details" validate:"max=1000"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	TenantID   string `json:"tenant_id" validate:"required"`
}

type AuditFilter struct {
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Page      int       `json:"page" validate:"min=1"`
	PageSize  int       `json:"page_size" validate:"min=1,max=100"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Email    string `json:"email" validate:"required,email,max=255"`
	TenantID string `json:"tenant_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"min=3,max=64"`
	Email    string `json:"email" validate:"email,max=255"`
	IsActive *bool  `json:"is_active"`
}

type UserFilter struct {
	TenantID string `json:"tenant_id"`
	IsActive *bool  `json:"is_active"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	SortBy   string `json:"sort_by"`
	SortOrder string `json:"sort_order" validate:"oneof=asc desc"`
}

type CreateRoleRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=64"`
	TenantID  string `json:"tenant_id" validate:"required"`
	IsDefault bool   `json:"is_default"`
}

type UpdateRoleRequest struct {
	Name      string `json:"name" validate:"min=2,max=64"`
	IsDefault *bool  `json:"is_default"`
}

type AnalyticsFilter struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	GroupBy   string    `json:"group_by"`
}

type NotificationFilter struct {
	Type      string    `json:"type"`
	IsRead    *bool     `json:"is_read"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Page      int       `json:"page" validate:"min=1"`
	PageSize  int       `json:"page_size" validate:"min=1,max=100"`
}

type SendNotificationRequest struct {
	UserID     string            `json:"user_id" validate:"required"`
	Title      string            `json:"title" validate:"required,max=255"`
	Message    string            `json:"message" validate:"required,max=1000"`
	Type       string            `json:"type" validate:"required,oneof=info warning error success"`
	Data       map[string]interface{} `json:"data"`
	TenantID   string            `json:"tenant_id" validate:"required"`
}

type UploadFileRequest struct {
	FileName    string            `json:"file_name" validate:"required"`
	ContentType string            `json:"content_type" validate:"required"`
	Data        []byte            `json:"data" validate:"required"`
	Size        int64             `json:"size" validate:"required"`
	TenantID    string            `json:"tenant_id" validate:"required"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type FileFilter struct {
	TenantID   string    `json:"tenant_id"`
	ContentType string    `json:"content_type"`
	MinSize    int64     `json:"min_size"`
	MaxSize    int64     `json:"max_size"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Page       int       `json:"page" validate:"min=1"`
	PageSize   int       `json:"page_size" validate:"min=1,max=100"`
}

type UsageFilter struct {
	TenantID   string `json:"tenant_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	GroupBy    string `json:"group_by"`
}

// Response 结构体将在实现文件中定义，这里只声明接口
type DocumentResponse struct {
	ID            uint                      `json:"id"`
	UUID          string                    `json:"uuid"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	OriginalName  string                    `json:"original_name"`
	MIMEType      string                    `json:"mime_type"`
	Size          int64                     `json:"size"`
	Category      string                    `json:"category"`
	Tags          []string                  `json:"tags"`
	EntityType    string                    `json:"entity_type"`
	EntityID      uint                      `json:"entity_id"`
	Status        string                    `json:"status"`
	CurrentVersion int                     `json:"current_version"`
	CreatedBy     uint                      `json:"created_by"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
	Versions      []*DocumentVersionResponse `json:"versions,omitempty"`
	Metadata      map[string]interface{}    `json:"metadata,omitempty"`
}

type DocumentListResponse struct {
	Documents []*DocumentResponse `json:"documents"`
	Total     int64               `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

type SearchResponse struct {
	Documents    []*DocumentResponse `json:"documents"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalPages   int                 `json:"total_pages"`
	Took         int                 `json:"took"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Suggestions  []*Suggestion      `json:"suggestions,omitempty"`
}

type DocumentVersionResponse struct {
	ID          uint      `json:"id"`
	DocumentID  uint      `json:"document_id"`
	Version     int       `json:"version"`
	UUID        string    `json:"uuid"`
	StoragePath string    `json:"storage_path"`
	FileHash    string    `json:"file_hash"`
	Size        int64     `json:"size"`
	Description string    `json:"description"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type VersionListResponse struct {
	Versions []*DocumentVersionResponse `json:"versions"`
	Total    int64                         `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
}

type DocumentStatsResponse struct {
	TotalDocuments int64                       `json:"total_documents"`
	TotalSize      int64                       `json:"total_size"`
	Categories     map[string]int64            `json:"categories"`
	Statuses       map[string]int64            `json:"statuses"`
	RecentDocuments int64                      `json:"recent_documents"`
	ByCreator      map[string]int64            `json:"by_creator"`
	ByEntityType  map[string]int64            `json:"by_entity_type"`
}

type TenantStatsResponse struct {
	DocumentStats *DocumentStatsResponse       `json:"document_stats"`
	UserStats     *UserStatsResponse           `json:"user_stats"`
	StorageStats  *StorageStatsResponse        `json:"storage_stats"`
	ActivityStats *ActivityStatsResponse       `json:"activity_stats"`
}

type PermissionListResponse struct {
	Permissions []*PermissionResponse `json:"permissions"`
	Total       int64                 `json:"total"`
}

type PermissionResponse struct {
	ID         uint   `json:"id"`
	DocumentID uint   `json:"document_id"`
	UserID     *uint  `json:"user_id"`
	RoleID     *uint  `json:"role_id"`
	TenantID   string `json:"tenant_id"`
	Permission string `json:"permission"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AuditResponse struct {
	ID         uint      `json:"id"`
	DocumentID uint      `json:"document_id"`
	UserID     uint      `json:"user_id"`
	TenantID   string    `json:"tenant_id"`
	Action     string    `json:"action"`
	Details    string    `json:"details"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditListResponse struct {
	Audits []*AuditResponse `json:"audits"`
	Total  int64             `json:"total"`
	Page   int               `json:"page"`
	PageSize int             `json:"page_size"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
	IsActive bool   `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponse struct {
	Users []*UserResponse `json:"users"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	PageSize int           `json:"page_size"`
}

type RoleResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenant_id"`
	IsDefault bool   `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Suggestion struct {
	Text  string  `json:"text"`
	Score float64 `json:"score"`
	Type  string  `json:"type"`
}

type VersionComparison struct {
	Version1    *DocumentVersionResponse `json:"version1"`
	Version2    *DocumentVersionResponse `json:"version2"`
	Changes     []*VersionChange          `json:"changes"`
	Summary     string                      `json:"summary"`
}

type VersionChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Type     string      `json:"type"` // added, modified, deleted
}

type VersionStatsResponse struct {
	TotalVersions   int64   `json:"total_versions"`
	TotalSize       int64   `json:"total_size"`
	AverageSize     float64 `json:"average_size"`
	LatestVersion   int     `json:"latest_version"`
	LatestSize      int64   `json:"latest_size"`
	LatestCreatedAt time.Time `json:"latest_created_at"`
}

type VersionDiff struct {
	Version1    *DocumentVersionResponse `json:"version1"`
	Version2    *DocumentVersionResponse `json:"version2"`
	Summary     string                      `json:"summary"`
	Differences []*VersionChange            `json:"differences"`
}

type ActivityReport struct {
	TotalActions    int64                       `json:"total_actions"`
	ActionsByType   map[string]int64            `json:"actions_by_type"`
	ActionsByUser   map[string]int64            `json:"actions_by_user"`
	ActionsByHour   map[string]int64            `json:"actions_by_hour"`
	PeakHours       []string                    `json:"peak_hours"`
	TopUsers        []*UserActivity              `json:"top_users"`
	TopDocuments    []*DocumentActivity           `json:"top_documents"`
}

type UserActivityReport struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	TotalActions int64     `json:"total_actions"`
	ActionsByType map[string]int64 `json:"actions_by_type"`
	DailyActivity map[string]int64   `json:"daily_activity"`
	MostActiveTime string   `json:"most_active_time"`
	TopDocuments   []*DocumentActivity `json:"top_documents"`
}

type UserActivity struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	ActionCount int64  `json:"action_count"`
}

type DocumentActivity struct {
	DocumentID string `json:"document_id"`
	DocumentName string `json:"document_name"`
	ActionCount int64  `json:"action_count"`
}

type SearchAnalytics struct {
	TotalSearches   int64                       `json:"total_searches"`
	UniqueQueries    int64                       `json:"unique_queries"`
	AverageResults   float64                     `json:"average_results"`
	TopQueries       []*PopularQuery             `json:"top_queries"`
	NoResultQueries  []string                    `json:"no_result_queries"`
	SearchesByHour   map[string]int64            `json:"searches_by_hour"`
	SearchesByType   map[string]int64            `json:"searches_by_type"`
}

type PopularQuery struct {
	Query     string `json:"query"`
	Count     int64  `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

type ActivityFilter struct {
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	GroupBy   string    `json:"group_by"`
}

type NotificationSettings struct {
	EmailEnabled   bool     `json:"email_enabled"`
	PushEnabled    bool     `json:"push_enabled"`
	InAppEnabled   bool     `json:"in_app_enabled"`
	Types          []string `json:"types"`
	QuietHours     *QuietHours `json:"quiet_hours,omitempty"`
}

type QuietHours struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Timezone  string `json:"timezone"`
}

type NotificationListResponse struct {
	Notifications []*NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
}

type NotificationResponse struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	IsRead    bool                   `json:"is_read"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

type FileResponse struct {
	ID          string                 `json:"id"`
	FileName    string                 `json:"file_name"`
	ContentType string                 `json:"content_type"`
	Size        int64                  `json:"size"`
	TenantID    string                 `json:"tenant_id"`
	StoragePath string                 `json:"storage_path"`
	FileHash    string                 `json:"file_hash"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type FileMetadata struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	FileHash    string `json:"file_hash"`
}

type DownloadMetadata struct {
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	FileHash    string    `json:"file_hash"`
	LastModified time.Time `json:"last_modified"`
}

type FileListResponse struct {
	Files    []*FileResponse `json:"files"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	PageSize int             `json:"page_size"`
}

type StorageStats struct {
	TotalFiles   int64            `json:"total_files"`
	TotalSize    int64            `json:"total_size"`
	UsedQuota    int64            `json:"used_quota"`
	AvailableQuota int64          `json:"available_quota"`
	FilesByType  map[string]int64 `json:"files_by_type"`
	SizeByType   map[string]int64 `json:"size_by_type"`
	DailyGrowth  map[string]int64 `json:"daily_growth"`
}

type StorageUsage struct {
	TenantID     string            `json:"tenant_id"`
	FileCount    int64             `json:"file_count"`
	TotalSize    int64             `json:"total_size"`
	UsageByType  map[string]int64  `json:"usage_by_type"`
	UsageByDate  map[string]int64  `json:"usage_by_date"`
}

type UserStatsResponse struct {
	TotalUsers     int64             `json:"total_users"`
	ActiveUsers    int64             `json:"active_users"`
	InactiveUsers  int64             `json:"inactive_users"`
	UsersByRole    map[string]int64  `json:"users_by_role"`
	NewUsersByDate map[string]int64  `json:"new_users_by_date"`
}

type ActivityStatsResponse struct {
	TotalActivities  int64                       `json:"total_activities"`
	ActivitiesByType map[string]int64            `json:"activities_by_type"`
	ActivitiesByHour map[string]int64            `json:"activities_by_hour"`
	TopUsers         []*UserActivity              `json:"top_users"`
	TopDocuments     []*DocumentActivity           `json:"top_documents"`
	RecentActivity   []*AuditResponse             `json:"recent_activity"`
}