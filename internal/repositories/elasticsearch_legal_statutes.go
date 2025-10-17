package repositories

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/models/elasticsearch"

	esv8 "github.com/elastic/go-elasticsearch/v8"
)

// elasticsearchStatuteRepository Elasticsearch法条搜索仓储实现
type elasticsearchStatuteRepository struct {
	client           *esv8.Client
	indexManager     *elasticsearch.LegalStatuteIndexManager
}

// NewElasticsearchStatuteRepository 创建Elasticsearch法条仓储实例
func NewElasticsearchStatuteRepository(client *esv8.Client) ElasticsearchStatuteRepository {
	return &elasticsearchStatuteRepository{
		client:       client,
		indexManager: elasticsearch.NewLegalStatuteIndexManager(client),
	}
}

// CreateIndex 创建索引
func (r *elasticsearchStatuteRepository) CreateIndex(ctx context.Context) error {
	return r.indexManager.CreateIndex()
}

// DeleteIndex 删除索引
func (r *elasticsearchStatuteRepository) DeleteIndex(ctx context.Context) error {
	return r.indexManager.DeleteIndex()
}

// IndexDocument 索引单个文档
func (r *elasticsearchStatuteRepository) IndexDocument(ctx context.Context, doc *elasticsearch.LegalStatuteDocument) error {
	return r.indexManager.IndexDocument(doc)
}

// BulkIndexDocuments 批量索引文档
func (r *elasticsearchStatuteRepository) BulkIndexDocuments(ctx context.Context, docs []*elasticsearch.LegalStatuteDocument) error {
	return r.indexManager.BulkIndexDocuments(docs)
}

// Search 搜索法条
func (r *elasticsearchStatuteRepository) Search(ctx context.Context, req *elasticsearch.LegalSearchRequest) (*elasticsearch.LegalSearchResponse, error) {
	start := time.Now()

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	response, err := r.indexManager.Search(req)
	if err != nil {
		return nil, err
	}

	// 计算搜索时间
	response.SearchTime = int(time.Since(start).Milliseconds())

	log.Printf("Elasticsearch搜索完成: 查询='%s', 耗时=%dms, 结果数=%d",
		req.Query, response.SearchTime, response.Total)

	return response, nil
}

// GetSuggestion 获取搜索建议
func (r *elasticsearchStatuteRepository) GetSuggestion(ctx context.Context, query string) ([]string, error) {
	if query == "" {
		return []string{}, nil
	}

	// 构建建议查询
	suggestQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"title_suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "title.suggest",
					"size":  10,
				},
			},
			"statute_suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "statute_number.suggest",
					"size":  5,
				},
			},
		},
	}

	// 暂时记录查询，后续实现实际ES查询
	log.Printf("ES建议查询: %+v", suggestQuery)

	// 这里需要实现实际的ES查询逻辑
	// 为了简化，返回一些基础建议
	suggestions := []string{
		query + "相关法条",
		"民法典" + query,
		"公司法" + query,
		"劳动合同法" + query,
	}

	return suggestions, nil
}

// GetRelatedStatutes 获取相关法条
func (r *elasticsearchStatuteRepository) GetRelatedStatutes(ctx context.Context, statuteID int, limit int) ([]*elasticsearch.LegalStatuteDocument, error) {
	// 首先获取当前法条信息
	currentReq := &elasticsearch.LegalSearchRequest{
		Query:    fmt.Sprintf("id:%d", statuteID),
		Page:     1,
		PageSize: 1,
	}

	currentResp, err := r.indexManager.Search(currentReq)
	if err != nil || len(currentResp.Documents) == 0 {
		return nil, fmt.Errorf("未找到指定法条")
	}

	currentDoc := currentResp.Documents[0]

	// 基于当前法条的类别、关键词等查找相关法条
	relatedReq := &elasticsearch.LegalSearchRequest{
		Query:         fmt.Sprintf("category.code:%s OR law_name:%s",
			currentDoc.Category.Code, currentDoc.LawName),
		Page:          1,
		PageSize:      limit,
		SortBy:        "relevance",
		SortOrder:     "desc",
		Highlight:     false,
	}

	// 排除当前法条
	// 这里需要在查询中添加排除条件，简化实现

	relatedResp, err := r.indexManager.Search(relatedReq)
	if err != nil {
		return nil, err
	}

	// 过滤掉当前法条
	var relatedDocs []*elasticsearch.LegalStatuteDocument
	for _, doc := range relatedResp.Documents {
		if doc.ID != statuteID {
			relatedDocs = append(relatedDocs, &doc)
		}
	}

	return relatedDocs, nil
}

// GetCategoryAggregation 获取分类聚合统计
func (r *elasticsearchStatuteRepository) GetCategoryAggregation(ctx context.Context) (map[string]int64, error) {
	// 构建聚合查询
	aggQuery := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"categories": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "category.code.keyword",
					"size":  50,
				},
			},
		},
	}

	// 暂时记录聚合查询，后续实现实际ES查询
	log.Printf("ES分类聚合查询: %+v", aggQuery)

	// 执行聚合查询
	// 这里需要实现实际的ES聚合查询逻辑
	// 为了简化，返回模拟数据
	result := map[string]int64{
		"CIVIL_LAW":      150,
		"CRIMINAL_LAW":   120,
		"COMMERCIAL_LAW": 180,
		"LABOR_LAW":      90,
		"OTHER":          60,
	}

	return result, nil
}

// GetLawNameAggregation 获取法律名称聚合统计
func (r *elasticsearchStatuteRepository) GetLawNameAggregation(ctx context.Context) (map[string]int64, error) {
	// 构建聚合查询
	aggQuery := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"law_names": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "law_name.keyword",
					"size":  20,
				},
			},
		},
	}

	// 暂时记录聚合查询，后续实现实际ES查询
	log.Printf("ES法律名称聚合查询: %+v", aggQuery)

	// 执行聚合查询
	// 这里需要实现实际的ES聚合查询逻辑
	// 为了简化，返回模拟数据
	result := map[string]int64{
		"中华人民共和国民法典":       89,
		"中华人民共和国公司法":       67,
		"中华人民共和国劳动合同法":   45,
		"中华人民共和国刑法":         78,
		"中华人民共和国民事诉讼法":   34,
	}

	return result, nil
}

// GetTagAggregation 获取标签聚合统计
func (r *elasticsearchStatuteRepository) GetTagAggregation(ctx context.Context) (map[string]int64, error) {
	// 构建聚合查询
	aggQuery := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"tags": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "tags",
					"size":  30,
				},
			},
		},
	}

	// 暂时记录聚合查询，后续实现实际ES查询
	log.Printf("ES标签聚合查询: %+v", aggQuery)

	// 执行聚合查询
	// 这里需要实现实际的ES聚合查询逻辑
	// 为了简化，返回模拟数据
	result := map[string]int64{
		"常用":   120,
		"重要":   89,
		"最新":   45,
		"基础":   67,
		"专业":   34,
	}

	return result, nil
}

// SyncFromPostgreSQL 从PostgreSQL同步数据到Elasticsearch
func (r *elasticsearchStatuteRepository) SyncFromPostgreSQL(ctx context.Context, postgresRepo LegalStatuteRepository) error {
	// 获取所有法条数据
	offset := 0
	batchSize := 100
	totalSynced := 0

	for {
		statutes, err := postgresRepo.List(ctx, offset, batchSize)
		if err != nil {
			return fmt.Errorf("获取法条数据失败: %v", err)
		}

		if len(statutes) == 0 {
			break
		}

		// 转换为ES文档格式
		esDocs := make([]*elasticsearch.LegalStatuteDocument, 0, len(statutes))
		for _, statute := range statutes {
			esDoc := r.convertToESDocument(statute)
			esDocs = append(esDocs, esDoc)
		}

		// 批量索引到ES
		if err := r.BulkIndexDocuments(ctx, esDocs); err != nil {
			return fmt.Errorf("批量索引失败: %v", err)
		}

		totalSynced += len(esDocs)
		log.Printf("已同步 %d 条法条数据到Elasticsearch", totalSynced)

		offset += batchSize
	}

	log.Printf("数据同步完成，总共同步了 %d 条法条数据", totalSynced)
	return nil
}

// convertToESDocument 将数据库模型转换为ES文档
func (r *elasticsearchStatuteRepository) convertToESDocument(statute *models.LegalStatute) *elasticsearch.LegalStatuteDocument {
	doc := &elasticsearch.LegalStatuteDocument{
		ID:                 statute.ID,
		StatuteNumber:      statute.StatuteNumber,
		Title:              statute.Title,
		Content:            statute.Content,
		LawName:            statute.LawName,
		Chapter:            statute.Chapter,
		Section:            statute.Section,
		Part:               statute.Part,
		EffectiveDate:      "",
		ExpiryDate:         "",
		PublishingAuthority: statute.PublishingAuthority,
		Status:             statute.Status,
		HierarchyLevel:     statute.HierarchyLevel,
		Tags:               statute.Tags,
		Keywords:           statute.Keywords,
		CreatedAt:          statute.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:          statute.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		ContentLength:      len(statute.Content),
		ViewCount:          0,
		FavoriteCount:      0,
		SearchWeight:       1.0,
	}

	// 处理可选字段
	if statute.EffectiveDate != nil {
		doc.EffectiveDate = statute.EffectiveDate.Format("2006-01-02")
	}
	if statute.ExpiryDate != nil {
		doc.ExpiryDate = statute.ExpiryDate.Format("2006-01-02")
	}
	if statute.ParentStatuteID != nil {
		doc.ParentStatuteID = *statute.ParentStatuteID
	}

	// 处理分类信息
	if statute.Category != nil {
		doc.Category = elasticsearch.Category{
			ID:   statute.Category.ID,
			Name: statute.Category.Name,
			Code: statute.Category.Code,
		}
	}

	return doc
}

// RebuildIndex 重建索引
func (r *elasticsearchStatuteRepository) RebuildIndex(ctx context.Context, postgresRepo LegalStatuteRepository) error {
	log.Println("开始重建法条搜索索引...")

	// 1. 删除现有索引
	if err := r.DeleteIndex(ctx); err != nil {
		log.Printf("删除索引失败: %v", err)
	}

	// 2. 创建新索引
	if err := r.CreateIndex(ctx); err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	// 3. 同步数据
	if err := r.SyncFromPostgreSQL(ctx, postgresRepo); err != nil {
		return fmt.Errorf("同步数据失败: %v", err)
	}

	log.Println("法条搜索索引重建完成")
	return nil
}