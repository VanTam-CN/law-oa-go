package repositories

import (
	"context"
	"fmt"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
)

// FolderTemplateRepository 卷宗目录模板仓储接口
type FolderTemplateRepository interface {
	CreateTemplate(ctx context.Context, template *models.CaseFolderTemplate) error
	GetTemplateByID(ctx context.Context, id uint) (*models.CaseFolderTemplate, error)
	UpdateTemplate(ctx context.Context, id uint, updates map[string]interface{}) error
	DeleteTemplate(ctx context.Context, id uint) error
	ListTemplates(ctx context.Context, params *FolderTemplateListParams) ([]*models.CaseFolderTemplate, int64, error)
	GetDefaultTemplate(ctx context.Context, caseType string) (*models.CaseFolderTemplate, error)
	ClearOtherDefaults(ctx context.Context, caseType string, excludeID uint) error

	// 案件文件夹实例
	CreateFolder(ctx context.Context, folder *models.CaseFolder) error
	GetFoldersByCaseID(ctx context.Context, caseID uint) ([]*models.CaseFolder, error)
	GetFolderByID(ctx context.Context, id uint) (*models.CaseFolder, error)
	DeleteFolder(ctx context.Context, id uint) error
	GetFolderPath(ctx context.Context, id uint) (string, error)
}

// FolderTemplateListParams 模板查询参数
type FolderTemplateListParams struct {
	CaseType string
	IsActive *bool
	Page     int
	PageSize int
}

type folderTemplateRepository struct {
	db *gorm.DB
}

// NewFolderTemplateRepository 创建卷宗目录模板仓储
func NewFolderTemplateRepository(db *gorm.DB) FolderTemplateRepository {
	return &folderTemplateRepository{db: db}
}

func (r *folderTemplateRepository) CreateTemplate(ctx context.Context, template *models.CaseFolderTemplate) error {
	if err := r.db.WithContext(ctx).Create(template).Error; err != nil {
		return fmt.Errorf("创建模板失败: %w", err)
	}
	return nil
}

func (r *folderTemplateRepository) GetTemplateByID(ctx context.Context, id uint) (*models.CaseFolderTemplate, error) {
	var t models.CaseFolderTemplate
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取模板失败: %w", err)
	}
	return &t, nil
}

func (r *folderTemplateRepository) UpdateTemplate(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&models.CaseFolderTemplate{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("更新模板失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *folderTemplateRepository) DeleteTemplate(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.CaseFolderTemplate{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除模板失败: %w", result.Error)
	}
	return nil
}

func (r *folderTemplateRepository) ListTemplates(ctx context.Context, params *FolderTemplateListParams) ([]*models.CaseFolderTemplate, int64, error) {
	var templates []*models.CaseFolderTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&models.CaseFolderTemplate{})

	if params.CaseType != "" {
		query = query.Where("case_type = ?", params.CaseType)
	}
	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计模板数量失败: %w", err)
	}

	page := params.Page
	if page <= 0 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	if err := query.Order("is_default DESC, created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, fmt.Errorf("查询模板列表失败: %w", err)
	}

	return templates, total, nil
}

func (r *folderTemplateRepository) GetDefaultTemplate(ctx context.Context, caseType string) (*models.CaseFolderTemplate, error) {
	var t models.CaseFolderTemplate
	query := r.db.WithContext(ctx).Where("is_active = ?", true)
	if caseType != "" {
		query = query.Where("case_type = ?", caseType)
	}
	// 优先返回默认模板
	if err := query.Order("is_default DESC, created_at DESC").First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取默认模板失败: %w", err)
	}
	return &t, nil
}

func (r *folderTemplateRepository) ClearOtherDefaults(ctx context.Context, caseType string, excludeID uint) error {
	result := r.db.WithContext(ctx).Model(&models.CaseFolderTemplate{}).
		Where("case_type = ? AND id != ?", caseType, excludeID).
		Update("is_default", false)
	if result.Error != nil {
		return fmt.Errorf("清除其他默认模板失败: %w", result.Error)
	}
	return nil
}

// --- 案件文件夹实例 ---

func (r *folderTemplateRepository) CreateFolder(ctx context.Context, folder *models.CaseFolder) error {
	if err := r.db.WithContext(ctx).Create(folder).Error; err != nil {
		return fmt.Errorf("创建案件文件夹失败: %w", err)
	}
	return nil
}

func (r *folderTemplateRepository) GetFoldersByCaseID(ctx context.Context, caseID uint) ([]*models.CaseFolder, error) {
	var folders []*models.CaseFolder
	// LEFT JOIN documents 统计每个文件夹内的文档数量
	query := `
		SELECT cf.*, COALESCE(d.doc_count, 0) AS document_count
		FROM case_folders cf
		LEFT JOIN (
			SELECT folder_id, COUNT(*) AS doc_count FROM documents WHERE folder_id IS NOT NULL GROUP BY folder_id
		) d ON d.folder_id = cf.id
		WHERE cf.case_id = ?
		ORDER BY cf.display_order ASC, cf.id ASC
	`
	if err := r.db.WithContext(ctx).Raw(query, caseID).Scan(&folders).Error; err != nil {
		return nil, fmt.Errorf("获取案件文件夹失败: %w", err)
	}
	return folders, nil
}

func (r *folderTemplateRepository) GetFolderByID(ctx context.Context, id uint) (*models.CaseFolder, error) {
	var f models.CaseFolder
	if err := r.db.WithContext(ctx).First(&f, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("获取文件夹失败: %w", err)
	}
	return &f, nil
}

func (r *folderTemplateRepository) DeleteFolder(ctx context.Context, id uint) error {
	// 使用递归CTE查询所有子节点ID，然后批量删除
	// PostgreSQL递归CTE语法：WITH RECURSIVE
	query := `
		WITH RECURSIVE folder_tree AS (
			-- 根节点
			SELECT id FROM case_folders WHERE id = ?
			UNION ALL
			-- 递归查找所有子节点
			SELECT cf.id FROM case_folders cf
			INNER JOIN folder_tree ft ON cf.parent_id = ft.id
		)
		DELETE FROM case_folders WHERE id IN (SELECT id FROM folder_tree)
	`
	if err := r.db.WithContext(ctx).Exec(query, id).Error; err != nil {
		return fmt.Errorf("删除文件夹失败: %w", err)
	}
	return nil
}

// GetFolderPath 获取文件夹路径（使用递归 CTE 单次查询，避免 N+1）
func (r *folderTemplateRepository) GetFolderPath(ctx context.Context, id uint) (string, error) {
	type pathRow struct {
		Name  string
		Depth int
	}
	var rows []pathRow

	query := `
		WITH RECURSIVE folder_path AS (
			SELECT name, 0 AS depth FROM case_folders WHERE id = ?
			UNION ALL
			SELECT cf.name, fp.depth + 1
			FROM case_folders cf
			INNER JOIN folder_path fp ON cf.id = fp.name  -- placeholder, fixed below
		)
		SELECT name, depth FROM folder_path ORDER BY depth DESC
	`
	// 正确的递归 CTE：从子节点向上遍历 parent_id
	query = `
		WITH RECURSIVE folder_path AS (
			SELECT id, name, parent_id, 0 AS depth FROM case_folders WHERE id = ?
			UNION ALL
			SELECT cf.id, cf.name, cf.parent_id, fp.depth + 1
			FROM case_folders cf
			INNER JOIN folder_path fp ON cf.id = fp.parent_id
		)
		SELECT name FROM folder_path ORDER BY depth DESC
	`

	if err := r.db.WithContext(ctx).Raw(query, id).Scan(&rows).Error; err != nil {
		return "", fmt.Errorf("获取文件夹路径失败: %w", err)
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("文件夹不存在")
	}

	path := rows[0].Name
	for i := 1; i < len(rows); i++ {
		path += "/" + rows[i].Name
	}
	return path, nil
}
