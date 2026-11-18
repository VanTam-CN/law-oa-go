package repositories

import (
	"context"
	"errors"

	"law-oa-go/internal/models"
)

// ErrDocumentNotFound 文档未找到错误
var ErrDocumentNotFound = errors.New("document not found")

// DocumentListParams 文档列表参数
type DocumentListParams struct {
	Page       int
	PageSize   int
	Category   string
	EntityType string
	EntityID   uint
	Search     string
	SortBy     string
	SortOrder  string
	// ViewerUserID 当前请求用户 ID。> 0 时启用隔离墙过滤：
	// 排除"已启用隔离墙且用户未在白名单中"的 case 关联文档。
	// = 0 时不应用过滤（仅供内部/后台任务使用）。
	ViewerUserID uint
	// OwnerScoped limits a non-management HTTP viewer to documents attached to
	// matters they own or are explicitly allowed to access. It also hides
	// unscoped documents from ordinary matter users instead of treating a
	// global list as an implicit read permission.
	OwnerScoped bool
}

// DocumentStats 文档统计
type DocumentStats struct {
	Total      int64
	ByCategory []struct {
		Category string
		Count    int64
	}
	ByEntityType []struct {
		EntityType string
		Count      int64
	}
	RecentUploads int64
}

// DocumentRepository 文档仓库接口
type DocumentRepository interface {
	// Create 创建文档
	Create(ctx context.Context, document *models.Document) error

	// FindByID 根据ID查找文档
	FindByID(ctx context.Context, id uint) (*models.Document, error)

	// List 列出文档
	List(ctx context.Context, params *DocumentListParams) ([]*models.Document, int64, error)

	// Update 更新文档
	Update(ctx context.Context, document *models.Document) error

	// Delete 删除文档
	Delete(ctx context.Context, id uint) error

	// GetStats 获取文档统计
	// viewerUserID > 0 时排除该用户无权访问的隔离墙案件文档，避免条数侧信道。
	GetStats(ctx context.Context, viewerUserID uint) (*DocumentStats, error)

	// FindByEntity 根据实体查找文档
	FindByEntity(ctx context.Context, entityType string, entityID uint) ([]*models.Document, error)
}
