package repositories

import (
	"context"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
)

// PolicyRepository 策略仓库接口
type PolicyRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, policy *models.Policy) error
	GetByID(ctx context.Context, id uint) (*models.Policy, error)
	Update(ctx context.Context, policy *models.Policy) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter *PolicyFilter) ([]*models.Policy, int64, error)

	// 租户相关操作
	GetByTenantID(ctx context.Context, tenantID string) ([]*models.Policy, error)
	GetEnabledByTenantID(ctx context.Context, tenantID string) ([]*models.Policy, error)

	// 策略查询
	GetByName(ctx context.Context, tenantID, name string) (*models.Policy, error)
	GetByResourceType(ctx context.Context, tenantID, resourceType string) ([]*models.Policy, error)
	GetBySubject(ctx context.Context, tenantID, subjectID string) ([]*models.Policy, error)

	// 批量操作
	BulkCreate(ctx context.Context, policies []*models.Policy) error
	BulkUpdate(ctx context.Context, policies []*models.Policy) error
	BulkDelete(ctx context.Context, ids []uint) error

	// 策略版本管理
	GetVersion(ctx context.Context, tenantID, name string, version int) (*models.Policy, error)
	GetLatestVersion(ctx context.Context, tenantID, name string) (*models.Policy, error)
	GetAllVersions(ctx context.Context, tenantID, name string) ([]*models.Policy, error)

	// 策略生效管理
	EnablePolicy(ctx context.Context, id uint) error
	DisablePolicy(ctx context.Context, id uint) error
	GetActivePolicies(ctx context.Context, tenantID string) ([]*models.Policy, error)

	// 策略统计
	CountByTenant(ctx context.Context, tenantID string) (int64, error)
	CountByStatus(ctx context.Context, tenantID string, enabled bool) (int64, error)

	// 策略搜索
	Search(ctx context.Context, query *PolicySearchQuery) ([]*models.Policy, int64, error)
}

// PolicyFilter 策略过滤器
type PolicyFilter struct {
	TenantID     string                 `json:"tenant_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Enabled      *bool                  `json:"enabled"`
	ResourceType string                 `json:"resource_type"`
	ActionType   string                 `json:"action_type"`
	SubjectType  string                 `json:"subject_type"`
	CreatorID    *uint                  `json:"creator_id"`
	Tags         []string               `json:"tags"`
	CreatedFrom  *time.Time             `json:"created_from"`
	CreatedTo    *time.Time             `json:"created_to"`
	UpdatedFrom  *time.Time             `json:"updated_from"`
	UpdatedTo    *time.Time             `json:"updated_to"`
	Pagination   *Pagination            `json:"pagination"`
	SortBy       string                 `json:"sort_by"`
	SortOrder    string                 `json:"sort_order"`
}

// PolicySearchQuery 策略搜索查询
type PolicySearchQuery struct {
	TenantID     string   `json:"tenant_id"`
	Query        string   `json:"query"`
	Fields       []string `json:"fields"`
	ResourceType string   `json:"resource_type"`
	ActionType   string   `json:"action_type"`
	SubjectType  string   `json:"subject_type"`
	Tags         []string `json:"tags"`
	Enabled      *bool    `json:"enabled"`
	CreatorID    *uint    `json:"creator_id"`
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
}

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// GetOffset 获取偏移量
func (p *Pagination) GetOffset() int {
	if p.Page <= 0 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *Pagination) GetLimit() int {
	if p.PageSize <= 0 {
		return 20 // 默认每页20条
	}
	if p.PageSize > 100 {
		return 100 // 最大每页100条
	}
	return p.PageSize
}