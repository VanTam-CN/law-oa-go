package middleware

import (
	"context"
	"fmt"

	"law-oa-go/internal/repositories"
)

// DocumentEntityTypeCase 文档关联实体为案件的字面量，与 DB schema 和 SQL scope 保持一致
const DocumentEntityTypeCase = "case"

// DocumentCaseResolver 基于 DocumentRepository 的 DocumentResolver 实现。
// 通过 FindByID 加载文档实体，避免在中间件中复制 SQL。
type DocumentCaseResolver struct {
	docRepo repositories.DocumentRepository
}

// NewDocumentCaseResolver 构造解析器
func NewDocumentCaseResolver(docRepo repositories.DocumentRepository) *DocumentCaseResolver {
	return &DocumentCaseResolver{docRepo: docRepo}
}

// ResolveDocumentCase 解析文档对应的 caseID。
//   - 文档不存在 -> (0, false, err)，调用方必须 fail-closed 返回 503
//   - 文档存在但 EntityType != "case" -> (0, false, nil)，跳过隔离墙
//   - 文档关联案件 -> (EntityID, true, nil)
func (r *DocumentCaseResolver) ResolveDocumentCase(ctx context.Context, documentID uint) (uint, bool, error) {
	if documentID == 0 {
		return 0, false, nil
	}
	doc, err := r.docRepo.FindByID(ctx, documentID)
	if err != nil {
		return 0, false, fmt.Errorf("load document %d: %w", documentID, err)
	}
	if doc == nil {
		return 0, false, nil
	}
	if doc.EntityType != DocumentEntityTypeCase {
		return 0, false, nil
	}
	return doc.EntityID, true, nil
}
