package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
)

// VersionRepository 文档版本仓库模拟
type VersionRepository struct {
	versions map[uint]*models.DocumentVersion
	nextID   uint
}

// NewVersionRepository 创建版本仓库模拟
func NewVersionRepository() *VersionRepository {
	return &VersionRepository{
		versions: make(map[uint]*models.DocumentVersion),
		nextID:   1,
	}
}

// GetByID 根据ID获取版本
func (r *VersionRepository) GetByID(ctx context.Context, id uint) (*models.DocumentVersion, error) {
	if version, exists := r.versions[id]; exists {
		return version, nil
	}
	return nil, errors.New("version not found")
}

// GetByDocument 根据文档ID获取版本列表
func (r *VersionRepository) GetByDocument(ctx context.Context, documentID uint) ([]*models.DocumentVersion, error) {
	var result []*models.DocumentVersion
	for _, version := range r.versions {
		if version.DocumentID == documentID {
			result = append(result, version)
		}
	}
	return result, nil
}

// GetLatestVersion 获取最新版本
func (r *VersionRepository) GetLatestVersion(ctx context.Context, documentID uint) (*models.DocumentVersion, error) {
	var latest *models.DocumentVersion
	for _, version := range r.versions {
		if version.DocumentID == documentID {
			if latest == nil || version.VersionNumber > latest.VersionNumber {
				latest = version
			}
		}
	}
	if latest == nil {
		return nil, errors.New("no version found")
	}
	return latest, nil
}

// Create 创建版本
func (r *VersionRepository) Create(ctx context.Context, version *models.DocumentVersion) error {
	version.ID = r.nextID
	version.CreatedAt = time.Now()
	r.versions[version.ID] = version
	r.nextID++
	return nil
}

// Update 更新版本
func (r *VersionRepository) Update(ctx context.Context, version *models.DocumentVersion) error {
	if _, exists := r.versions[version.ID]; exists {
		r.versions[version.ID] = version
		return nil
	}
	return errors.New("version not found")
}

// Delete 删除版本
func (r *VersionRepository) Delete(ctx context.Context, id uint) error {
	if _, exists := r.versions[id]; exists {
		delete(r.versions, id)
		return nil
	}
	return errors.New("version not found")
}

// List 列出版本
func (r *VersionRepository) List(ctx context.Context, options VersionListOptions) ([]*models.DocumentVersion, int64, error) {
	var result []*models.DocumentVersion
	var total int64

	for _, version := range r.versions {
		// 应用过滤条件
		if options.DocumentID != 0 && version.DocumentID != options.DocumentID {
			continue
		}
		if options.CreatedBy != 0 && version.CreatedBy != options.CreatedBy {
			continue
		}

		total++
		// 应用分页
		if options.Offset > 0 {
			options.Offset--
			continue
		}
		if options.Limit > 0 && len(result) >= options.Limit {
			continue
		}
		result = append(result, version)
	}

	return result, total, nil
}

// VersionListOptions 版本列表选项
type VersionListOptions struct {
	DocumentID uint
	CreatedBy  uint
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
}