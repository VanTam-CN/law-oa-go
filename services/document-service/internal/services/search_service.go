package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// searchService 搜索服务实现
type searchService struct {
	client      *elasticsearch.Client
	docRepo     repositories.DocumentRepository
	userRepo    repositories.UserRepository
	auditRepo   repositories.DocumentAuditRepository
	logger      *logrus.Logger
	indexName   string
}

// NewSearchService 创建新的搜索服务
func NewSearchService(
	client *elasticsearch.Client,
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	auditRepo repositories.DocumentAuditRepository,
	logger *logrus.Logger,
	indexName string,
) SearchService {
	if indexName == "" {
		indexName = "documents"
	}

	return &searchService{
		client:    client,
		docRepo:   docRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		logger:    logger,
		indexName: indexName,
	}
}

// IndexDocument 索引文档
func (s *searchService) IndexDocument(ctx context.Context, req *IndexDocumentRequest) error {
	// 获取文档信息
	document, err := s.docRepo.GetByID(ctx, req.DocumentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// 构建索引文档
	doc := s.buildSearchDocument(document)

	// 序列化文档
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	// 构建索引请求
	reqES := esapi.IndexRequest{
		Index:      s.indexName,
		DocumentID: fmt.Sprintf("%d", document.ID),
		Body:       bytes.NewReader(docJSON),
		Refresh:    "true",
	}

	// 执行索引请求
	res, err := reqES.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing failed: %s", res.Status())
	}

	s.logger.WithFields(map[string]interface{}{
		"document_id": document.ID,
		"document_name": document.Name,
		"tenant_id":   document.TenantID,
	}).Info("Document indexed successfully")

	return nil
}

// SearchDocuments 搜索文档
func (s *searchService) SearchDocuments(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	// 构建搜索查询
	query := s.buildSearchQuery(req)

	// 序列化查询
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	// 计算分页
	from := (req.Page - 1) * req.PageSize

	// 构建搜索请求
	searchReq := esapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryJSON),
		From:  &from,
		Size:  &req.PageSize,
		Sort:  []string{"_score:desc", "updated_at:desc"},
	}

	// 执行搜索请求
	res, err := searchReq.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.Status())
	}

	// 解析搜索结果
	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// 提取文档和统计信息
	hits := searchResult["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})["value"].(float64)
	hitsList := hits["hits"].([]interface{})

	// 转换搜索结果
	documents := make([]*SearchDocument, len(hitsList))
	for i, hit := range hitsList {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		score := hitMap["_score"].(float64)

		doc := &SearchDocument{
			ID:          uint(source["id"].(float64)),
			UUID:        source["uuid"].(string),
			Name:        source["name"].(string),
			Description: source["description"].(string),
			Content:     source["content"].(string),
			Category:    source["category"].(string),
			TenantID:    source["tenant_id"].(string),
			CreatedBy:   uint(source["created_by"].(float64)),
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
			Score:       score,
		}

		// 解析时间字段
		if createdStr, ok := source["created_at"].(string); ok {
			if created, err := time.Parse(time.RFC3339, createdStr); err == nil {
				doc.CreatedAt = created
			}
		}
		if updatedStr, ok := source["updated_at"].(string); ok {
			if updated, err := time.Parse(time.RFC3339, updatedStr); err == nil {
				doc.UpdatedAt = updated
			}
		}

		documents[i] = doc
	}

	return &SearchResponse{
		Documents: documents,
		Total:     int64(total),
		Page:      req.Page,
		PageSize:  req.PageSize,
		Query:     req.Query,
		Filters:   req.Filters,
	}, nil
}

// DeleteDocument 删除文档索引
func (s *searchService) DeleteDocument(ctx context.Context, documentID uint) error {
	// 构建删除请求
	req := esapi.DeleteRequest{
		Index:      s.indexName,
		DocumentID: fmt.Sprintf("%d", documentID),
		Refresh:    "true",
	}

	// 执行删除请求
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete document from index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete failed: %s", res.Status())
	}

	s.logger.WithField("document_id", documentID).Info("Document deleted from index")

	return nil
}

// UpdateDocument 更新文档索引
func (s *searchService) UpdateDocument(ctx context.Context, req *UpdateDocumentRequest) error {
	// 重新索引文档
	indexReq := &IndexDocumentRequest{
		DocumentID: req.DocumentID,
	}

	return s.IndexDocument(ctx, indexReq)
}

// SuggestDocuments 获取搜索建议
func (s *searchService) SuggestDocuments(ctx context.Context, req *SuggestRequest) (*SuggestResponse, error) {
	// 构建建议查询
	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"prefix": map[string]interface{}{
				"prefix": req.Query,
				"completion": map[string]interface{}{
					"field": "name_suggest",
					"size":  req.Size,
				},
			},
		},
	}

	// 序列化查询
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggest query: %w", err)
	}

	// 构建搜索请求
	searchReq := esapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryJSON),
	}

	// 执行搜索请求
	res, err := searchReq.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("suggest failed: %s", res.Status())
	}

	// 解析建议结果
	var suggestResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&suggestResult); err != nil {
		return nil, fmt.Errorf("failed to decode suggest response: %w", err)
	}

	// 提取建议
	suggestions := make([]string, 0)
	if suggestData, ok := suggestResult["suggest"].(map[string]interface{}); ok {
		if prefixData, ok := suggestData["prefix"].([]interface{}); ok && len(prefixData) > 0 {
			if options := prefixData[0].(map[string]interface{})["options"].([]interface{}); ok {
				for _, option := range options {
					if optMap := option.(map[string]interface{}); ok {
						if text, ok := optMap["text"].(string); ok {
							suggestions = append(suggestions, text)
						}
					}
				}
			}
		}
	}

	return &SuggestResponse{
		Query:       req.Query,
		Suggestions: suggestions,
	}, nil
}

// GetSearchStats 获取搜索统计
func (s *searchService) GetSearchStats(ctx context.Context, tenantID string) (*SearchStats, error) {
	// 构建聚合查询
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"tenant_id": tenantID,
			},
		},
		"aggs": map[string]interface{}{
			"categories": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "category",
					"size":  20,
				},
			},
			"created_by": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "created_by",
					"size":  20,
				},
			},
			"dates": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":    "created_at",
					"calendar_interval": "month",
				},
			},
		},
		"size": 0,
	}

	// 序列化查询
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stats query: %w", err)
	}

	// 构建搜索请求
	searchReq := esapi.SearchRequest{
		Index: []string{s.indexName},
		Body:  bytes.NewReader(queryJSON),
	}

	// 执行搜索请求
	res, err := searchReq.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get search stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("stats query failed: %s", res.Status())
	}

	// 解析统计结果
	var statsResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&statsResult); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	// 提取聚合数据
	aggregations := statsResult["aggregations"].(map[string]interface{})

	stats := &SearchStats{
		TotalDocuments: int64(statsResult["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		Categories:     make(map[string]int64),
		Creators:       make(map[string]int64),
		Dates:          make(map[string]int64),
	}

	// 提取分类统计
	if categories, ok := aggregations["categories"].(map[string]interface{})["buckets"].([]interface{}); ok {
		for _, bucket := range categories {
			bucketMap := bucket.(map[string]interface{})
			key := bucketMap["key"].(string)
			count := int64(bucketMap["doc_count"].(float64))
			stats.Categories[key] = count
		}
	}

	// 提取创建者统计
	if creators, ok := aggregations["created_by"].(map[string]interface{})["buckets"].([]interface{}); ok {
		for _, bucket := range creators {
			bucketMap := bucket.(map[string]interface{})
			userID := fmt.Sprintf("%.0f", bucketMap["key"].(float64))
			count := int64(bucketMap["doc_count"].(float64))
			stats.Creators[userID] = count
		}
	}

	// 提取日期统计
	if dates, ok := aggregations["dates"].(map[string]interface{})["buckets"].([]interface{}); ok {
		for _, bucket := range dates {
			bucketMap := bucket.(map[string]interface{})
			dateKey := bucketMap["key_as_string"].(string)
			count := int64(bucketMap["doc_count"].(float64))
			stats.Dates[dateKey] = count
		}
	}

	return stats, nil
}

// CreateIndex 创建索引
func (s *searchService) CreateIndex(ctx context.Context) error {
	// 构建索引映射
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "integer",
				},
				"uuid": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
						"suggest": map[string]interface{}{
							"type": "completion",
						},
					},
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"tenant_id": map[string]interface{}{
					"type": "keyword",
				},
				"created_by": map[string]interface{}{
					"type": "integer",
				},
				"created_at": map[string]interface{}{
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis",
				},
				"updated_at": map[string]interface{}{
					"type": "date",
					"format": "strict_date_optional_time||epoch_millis",
				},
				"mime_type": map[string]interface{}{
					"type": "keyword",
				},
				"size": map[string]interface{}{
					"type": "long",
				},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
	}

	// 序列化映射
	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal index mapping: %w", err)
	}

	// 构建创建索引请求
	req := esapi.IndicesCreateRequest{
		Index: s.indexName,
		Body:  bytes.NewReader(mappingJSON),
	}

	// 执行创建请求
	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 { // 400表示索引已存在
		return fmt.Errorf("create index failed: %s", res.Status())
	}

	s.logger.WithField("index", s.indexName).Info("Index created or already exists")

	return nil
}

// ReindexAllDocuments 重新索引所有文档
func (s *searchService) ReindexAllDocuments(ctx context.Context, tenantID string) error {
	// 获取所有文档
	documents, err := s.docRepo.List(ctx, repositories.DocumentListOptions{
		TenantID: tenantID,
		Limit:    10000, // 批量处理
		Offset:   0,
	})

	if err != nil {
		return fmt.Errorf("failed to get documents for reindexing: %w", err)
	}

	successCount := 0
	failureCount := 0

	for _, document := range documents {
		req := &IndexDocumentRequest{
			DocumentID: document.ID,
		}

		if err := s.IndexDocument(ctx, req); err != nil {
			failureCount++
			s.logger.WithError(err).WithField("document_id", document.ID).Error("Failed to index document")
		} else {
			successCount++
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"tenant_id":    tenantID,
		"success_count": successCount,
		"failure_count": failureCount,
		"total_count":   len(documents),
	}).Info("Reindexing completed")

	return nil
}

// 辅助方法

// buildSearchDocument 构建搜索文档
func (s *searchService) buildSearchDocument(doc *models.Document) map[string]interface{} {
	return map[string]interface{}{
		"id":          doc.ID,
		"uuid":        doc.UUID,
		"name":        doc.Name,
		"description": doc.Description,
		"content":     "", // 实际实现中需要提取文档内容
		"category":    doc.Category,
		"tenant_id":   doc.TenantID,
		"created_by":  doc.CreatedBy,
		"created_at":  doc.CreatedAt.Format(time.RFC3339),
		"updated_at":  doc.UpdatedAt.Format(time.RFC3339),
		"mime_type":   doc.MIMEType,
		"size":        doc.Size,
	}
}

// buildSearchQuery 构建搜索查询
func (s *searchService) buildSearchQuery(req *SearchRequest) map[string]interface{} {
	// 构建基础查询
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{},
			},
		},
	}

	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
	mustQueries := boolQuery["must"].([]interface{})

	// 添加租户过滤
	if req.TenantID != "" {
		tenantQuery := map[string]interface{}{
			"term": map[string]interface{}{
				"tenant_id": req.TenantID,
			},
		}
		mustQueries = append(mustQueries, tenantQuery)
	}

	// 添加搜索条件
	if req.Query != "" {
		searchQuery := map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"name^3", "description^2", "content"},
				"type":   "best_fields",
			},
		}
		mustQueries = append(mustQueries, searchQuery)
	}

	// 添加过滤器
	if len(req.Filters) > 0 {
		filterQueries := make([]interface{}, 0)
		for key, value := range req.Filters {
			if key == "category" {
				filterQueries = append(filterQueries, map[string]interface{}{
					"term": map[string]interface{}{
						"category": value,
					},
				})
			} else if key == "created_by" {
				filterQueries = append(filterQueries, map[string]interface{}{
					"term": map[string]interface{}{
						"created_by": value,
					},
				})
			} else if key == "date_range" {
				if dateRange, ok := value.(map[string]interface{}); ok {
					if start, ok := dateRange["start"]; ok {
						if end, ok := dateRange["end"]; ok {
							filterQueries = append(filterQueries, map[string]interface{}{
								"range": map[string]interface{}{
									"created_at": map[string]interface{}{
										"gte": start,
										"lte": end,
									},
								},
							})
						}
					}
				}
			}
		}

		if len(filterQueries) > 0 {
			boolQuery["filter"] = filterQueries
		}
	}

	// 更新must查询
	boolQuery["must"] = mustQueries

	return query
}