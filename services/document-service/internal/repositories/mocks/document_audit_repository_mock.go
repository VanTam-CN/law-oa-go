package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
)

// DocumentAuditRepository 文档审计仓库模拟
type DocumentAuditRepository struct {
	audits map[uint]*models.DocumentAudit
	nextID uint
}

// NewDocumentAuditRepository 创建文档审计仓库模拟
func NewDocumentAuditRepository() *DocumentAuditRepository {
	return &DocumentAuditRepository{
		audits: make(map[uint]*models.DocumentAudit),
		nextID: 1,
	}
}

// Create 创建审计记录
func (r *DocumentAuditRepository) Create(ctx context.Context, audit *models.DocumentAudit) error {
	audit.ID = r.nextID
	audit.CreatedAt = time.Now()
	r.audits[audit.ID] = audit
	r.nextID++
	return nil
}

// GetByID 根据ID获取审计记录
func (r *DocumentAuditRepository) GetByID(ctx context.Context, id uint) (*models.DocumentAudit, error) {
	if audit, exists := r.audits[id]; exists {
		return audit, nil
	}
	return nil, errors.New("audit not found")
}

// List 列出审计记录
func (r *DocumentAuditRepository) List(ctx context.Context, options AuditListOptions) ([]*models.DocumentAudit, int64, error) {
	var result []*models.DocumentAudit
	var total int64

	for _, audit := range r.audits {
		// 应用过滤条件
		if options.TenantID != "" && audit.TenantID != options.TenantID {
			continue
		}
		if options.UserID != 0 && (audit.UserID == nil || *audit.UserID != options.UserID) {
			continue
		}
		if options.DocumentID != 0 && (audit.DocumentID == nil || *audit.DocumentID != options.DocumentID) {
			continue
		}
		if options.Action != "" && audit.Action != options.Action {
			continue
		}
		if !options.StartDate.IsZero() && audit.CreatedAt.Before(options.StartDate) {
			continue
		}
		if !options.EndDate.IsZero() && audit.CreatedAt.After(options.EndDate) {
			continue
		}
		if options.Search != "" && !contains(audit.Action, options.Search) && !contains(audit.Details, options.Search) {
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
		result = append(result, audit)
	}

	return result, total, nil
}

// DeleteBeforeDate 删除指定日期之前的审计记录
func (r *DocumentAuditRepository) DeleteBeforeDate(ctx context.Context, tenantID string, beforeDate time.Time) (int64, error) {
	var deletedCount int64
	for id, audit := range r.audits {
		if audit.TenantID == tenantID && audit.CreatedAt.Before(beforeDate) {
			delete(r.audits, id)
			deletedCount++
		}
	}
	return deletedCount, nil
}

// AuditListOptions 审计列表选项
type AuditListOptions struct {
	TenantID   string
	UserID     uint
	DocumentID uint
	Action     string
	StartDate  time.Time
	EndDate    time.Time
	Limit      int
	Offset     int
	SortBy     string
	SortOrder  string
	Search     string
}