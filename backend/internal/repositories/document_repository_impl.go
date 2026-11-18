package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// documentRepository 文档仓库实现
type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建文档仓库
func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{
		db: db,
	}
}

// applyEthicalWallScope 在文档查询上叠加隔离墙过滤：
//
//	排除 entity_type='case' 且对应案件启用隔离墙、且 viewer 不在白名单中的文档。
//
// viewerUserID=0 时不应用过滤（仅供内部/后台任务使用，禁止用于 HTTP 请求路径）。
func applyEthicalWallScope(query *gorm.DB, viewerUserID uint) *gorm.DB {
	if viewerUserID == 0 {
		return query
	}
	return query.Where(
		`NOT EXISTS (
			SELECT 1 FROM cases c
			WHERE documents.entity_type = 'case'
			  AND c.id = documents.entity_id
			  AND c.ethical_wall_enabled = TRUE
			  AND c.deleted_at IS NULL
				AND NOT EXISTS (
			    SELECT 1 FROM case_ethical_wall_whitelist w
			    WHERE w.case_id = c.id AND w.user_id = ?
			  )
		)
		AND NOT EXISTS (
			SELECT 1 FROM cases client_cases
			WHERE documents.entity_type = 'client'
			  AND client_cases.client_id = documents.entity_id
			  AND client_cases.ethical_wall_enabled = TRUE
			  AND client_cases.deleted_at IS NULL
			  AND NOT EXISTS (
			    SELECT 1 FROM case_ethical_wall_whitelist w
			    WHERE w.case_id = client_cases.id AND w.user_id = ?
			  )
		)`,
		viewerUserID, viewerUserID,
	)
}

func applyDocumentOwnerScope(query *gorm.DB, viewerUserID uint) *gorm.DB {
	if viewerUserID == 0 {
		return query
	}
	return query.Where(`
		(
			documents.entity_type = 'case'
			AND EXISTS (
				SELECT 1 FROM cases c
				WHERE c.id = documents.entity_id
				  AND c.deleted_at IS NULL
				  AND (c.lawyer_id = ? OR c.created_by = ?)
				  AND (
					c.ethical_wall_enabled = FALSE
					OR c.lawyer_id = ?
					OR c.created_by = ?
					OR EXISTS (
						SELECT 1 FROM case_ethical_wall_whitelist w
						WHERE w.case_id = c.id AND w.user_id = ?
					)
				  )
			)
		)
		OR (
			documents.entity_type = 'client'
			AND EXISTS (
				SELECT 1 FROM cases c
				WHERE c.client_id = documents.entity_id
				  AND c.deleted_at IS NULL
				  AND (c.lawyer_id = ? OR c.created_by = ?)
				  AND (
					c.ethical_wall_enabled = FALSE
					OR c.lawyer_id = ?
					OR c.created_by = ?
					OR EXISTS (
						SELECT 1 FROM case_ethical_wall_whitelist w
						WHERE w.case_id = c.id AND w.user_id = ?
					)
				  )
			)
		)`, viewerUserID, viewerUserID, viewerUserID, viewerUserID, viewerUserID,
		viewerUserID, viewerUserID, viewerUserID, viewerUserID, viewerUserID)
}

// Create 创建文档
func (r *documentRepository) Create(ctx context.Context, document *models.Document) error {
	return r.db.WithContext(ctx).Create(document).Error
}

// FindByID 根据ID查找文档
func (r *documentRepository) FindByID(ctx context.Context, id uint) (*models.Document, error) {
	var document models.Document
	err := r.db.WithContext(ctx).Where("id = ? AND status <> ?", id, "deleted").First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	return &document, nil
}

// List 列出文档
func (r *documentRepository) List(ctx context.Context, params *DocumentListParams) ([]*models.Document, int64, error) {
	if params == nil {
		return nil, 0, errors.New("document list params must not be nil")
	}
	var documents []*models.Document
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Document{})

	// 应用过滤条件
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}
	if params.EntityType != "" {
		query = query.Where("entity_type = ?", params.EntityType)
	}
	if params.EntityID > 0 {
		query = query.Where("entity_id = ?", params.EntityID)
	}
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ? OR tags LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// 隔离墙过滤（列表与计数必须共用同一 scope，避免条数侧信道）
	query = applyEthicalWallScope(query, params.ViewerUserID)
	if params.OwnerScoped {
		query = applyDocumentOwnerScope(query, params.ViewerUserID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用排序。字段名必须走白名单，避免把查询参数拼进 ORDER BY。
	order := "created_at DESC"
	if params.SortBy != "" {
		sortColumns := map[string]string{
			"created_at":  "created_at",
			"updated_at":  "updated_at",
			"name":        "name",
			"filename":    "filename",
			"filesize":    "filesize",
			"mime_type":   "mime_type",
			"category":    "category",
			"entity_type": "entity_type",
			"entity_id":   "entity_id",
		}
		if column, ok := sortColumns[strings.ToLower(strings.TrimSpace(params.SortBy))]; ok {
			direction := "DESC"
			if strings.EqualFold(params.SortOrder, "asc") {
				direction = "ASC"
			}
			order = column + " " + direction
		}
	}
	query = query.Order(order)

	// 应用分页
	if params.Page > 0 && params.PageSize > 0 {
		offset := (params.Page - 1) * params.PageSize
		query = query.Offset(offset).Limit(params.PageSize)
	}

	// 执行查询
	if err := query.Find(&documents).Error; err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// Update 更新文档
func (r *documentRepository) Update(ctx context.Context, document *models.Document) error {
	document.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(document).Error
}

// Delete 删除文档
func (r *documentRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Document{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return nil
}

// GetStats 获取文档统计
func (r *documentRepository) GetStats(ctx context.Context, viewerUserID uint) (*DocumentStats, error) {
	var stats DocumentStats

	baseQuery := r.db.WithContext(ctx).Model(&models.Document{})
	scopedQuery := applyEthicalWallScope(baseQuery, viewerUserID)

	// 总数
	if err := scopedQuery.Count(&stats.Total).Error; err != nil {
		return nil, err
	}

	// 按分类统计
	rows, err := scopedQuery.
		Session(&gorm.Session{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int64
		if err := r.db.ScanRows(rows, &struct {
			Category string
			Count    int64
		}{category, count}); err != nil {
			continue
		}
		stats.ByCategory = append(stats.ByCategory, struct {
			Category string
			Count    int64
		}{category, count})
	}

	// 按实体类型统计（独立 session 复用 scope 条件）
	entityRows, err := applyEthicalWallScope(r.db.WithContext(ctx).Model(&models.Document{}), viewerUserID).
		Select("entity_type, COUNT(*) as count").
		Group("entity_type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer entityRows.Close()

	for entityRows.Next() {
		var entityType string
		var count int64
		if err := r.db.ScanRows(entityRows, &struct {
			EntityType string
			Count      int64
		}{entityType, count}); err != nil {
			continue
		}
		stats.ByEntityType = append(stats.ByEntityType, struct {
			EntityType string
			Count      int64
		}{entityType, count})
	}

	// 最近上传统计（过去7天）
	var recentCount int64
	if err := applyEthicalWallScope(r.db.WithContext(ctx).Model(&models.Document{}), viewerUserID).
		Where("created_at >= ?", time.Now().AddDate(0, 0, -7)).
		Count(&recentCount).Error; err != nil {
		return nil, err
	}
	stats.RecentUploads = recentCount

	return &stats, nil
}

// FindByEntity 根据实体查找文档
func (r *documentRepository) FindByEntity(ctx context.Context, entityType string, entityID uint) ([]*models.Document, error) {
	var documents []*models.Document
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Find(&documents).Error
	return documents, err
}
