package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// LegalStatuteDocument 法条ES文档结构
type LegalStatuteDocument struct {
	ID                  int      `json:"id"`
	StatuteNumber       string   `json:"statute_number"`
	Title               string   `json:"title"`
	Content             string   `json:"content"`
	LawName             string   `json:"law_name"`
	Chapter             string   `json:"chapter"`
	Section             string   `json:"section"`
	Part                string   `json:"part"`
	Category            Category `json:"category"`
	EffectiveDate       string   `json:"effective_date"`
	ExpiryDate          string   `json:"expiry_date"`
	PublishingAuthority string   `json:"publishing_authority"`
	Status              string   `json:"status"`
	HierarchyLevel      int      `json:"hierarchy_level"`
	ParentStatuteID     int      `json:"parent_statute_id"`
	OrderInHierarchy    int      `json:"order_in_hierarchy"`
	Tags                []string `json:"tags"`
	Keywords            []string `json:"keywords"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
	ContentLength       int      `json:"content_length"`
	ViewCount           int      `json:"view_count"`
	FavoriteCount       int      `json:"favorite_count"`
	SearchWeight        float64  `json:"search_weight"`
}

// Category 法条分类
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// LegalSearchRequest 法条搜索请求
type LegalSearchRequest struct {
	Query             string   `json:"query"`
	CategoryCodes     []string `json:"category_codes,omitempty"`
	LawNames          []string `json:"law_names,omitempty"`
	Status            string   `json:"status,omitempty"`
	EffectiveDateFrom string   `json:"effective_date_from,omitempty"`
	EffectiveDateTo   string   `json:"effective_date_to,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	SortBy            string   `json:"sort_by,omitempty"`
	SortOrder         string   `json:"sort_order,omitempty"`
	Page              int      `json:"page,omitempty"`
	PageSize          int      `json:"page_size,omitempty"`
	Highlight         bool     `json:"highlight,omitempty"`
}

// LegalSearchResponse 法条搜索响应
type LegalSearchResponse struct {
	Total        int64                  `json:"total"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"page_size"`
	Documents    []LegalStatuteDocument `json:"documents"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Suggestions  []string               `json:"suggestions,omitempty"`
	SearchTime   int                    `json:"search_time_ms"`
}

// LegalStatuteIndexManager 法条索引管理器
type LegalStatuteIndexManager struct {
	client    *elasticsearch.Client
	indexName string
}

// NewLegalStatuteIndexManager 创建法条索引管理器
func NewLegalStatuteIndexManager(client *elasticsearch.Client) *LegalStatuteIndexManager {
	return &LegalStatuteIndexManager{
		client:    client,
		indexName: "legal_statutes",
	}
}

// CreateIndex 创建法条索引
func (m *LegalStatuteIndexManager) CreateIndex() error {
	mapping, err := m.loadMapping()
	if err != nil {
		return fmt.Errorf("加载索引映射失败: %v", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: m.indexName,
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(context.Background(), m.client)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引失败: %s", res.Status())
	}

	log.Printf("法条索引 %s 创建成功", m.indexName)
	return nil
}

// DeleteIndex 删除法条索引
func (m *LegalStatuteIndexManager) DeleteIndex() error {
	req := esapi.IndicesDeleteRequest{
		Index: []string{m.indexName},
	}

	res, err := req.Do(context.Background(), m.client)
	if err != nil {
		return fmt.Errorf("删除索引失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("删除索引失败: %s", res.Status())
	}

	log.Printf("法条索引 %s 删除成功", m.indexName)
	return nil
}

// IndexDocument 索引单个法条文档
func (m *LegalStatuteIndexManager) IndexDocument(doc *LegalStatuteDocument) error {
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("序列化文档失败: %v", err)
	}

	req := esapi.IndexRequest{
		Index:      m.indexName,
		DocumentID: fmt.Sprintf("%d", doc.ID),
		Body:       strings.NewReader(string(docJSON)),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), m.client)
	if err != nil {
		return fmt.Errorf("索引文档失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("索引文档失败: %s", res.Status())
	}

	return nil
}

// BulkIndexDocuments 批量索引法条文档
func (m *LegalStatuteIndexManager) BulkIndexDocuments(docs []*LegalStatuteDocument) error {
	if len(docs) == 0 {
		return nil
	}

	var buf strings.Builder
	for _, doc := range docs {
		// 添加索引操作头
		indexMeta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": m.indexName,
				"_id":    fmt.Sprintf("%d", doc.ID),
			},
		}
		metaJSON, _ := json.Marshal(indexMeta)
		buf.WriteString(string(metaJSON))
		buf.WriteString("\n")

		// 添加文档数据
		docJSON, _ := json.Marshal(doc)
		buf.WriteString(string(docJSON))
		buf.WriteString("\n")
	}

	req := esapi.BulkRequest{
		Body:    strings.NewReader(buf.String()),
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), m.client)
	if err != nil {
		return fmt.Errorf("批量索引失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("批量索引失败: %s", res.Status())
	}

	return nil
}

// Search 搜索法条
func (m *LegalStatuteIndexManager) Search(req *LegalSearchRequest) (*LegalSearchResponse, error) {
	query, err := m.buildSearchQuery(req)
	if err != nil {
		return nil, fmt.Errorf("构建搜索查询失败: %v", err)
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("序列化查询失败: %v", err)
	}

	esReq := esapi.SearchRequest{
		Index: []string{m.indexName},
		Body:  strings.NewReader(string(queryJSON)),
	}

	res, err := esReq.Do(context.Background(), m.client)
	if err != nil {
		return nil, fmt.Errorf("执行搜索失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("搜索失败: %s", res.Status())
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %v", err)
	}

	return m.parseSearchResponse(searchResult, req)
}

// buildSearchQuery 构建ES搜索查询
func (m *LegalStatuteIndexManager) buildSearchQuery(req *LegalSearchRequest) (map[string]interface{}, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{},
			},
		},
		"from": (req.Page - 1) * req.PageSize,
		"size": req.PageSize,
	}

	// 添加主查询
	mustQueries := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	if req.Query != "" {
		mainQuery := map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": req.Query,
				"fields": []string{
					"title^3",
					"statute_number^2",
					"content",
					"keywords^2",
					"law_name^1.5",
				},
				"type":      "best_fields",
				"fuzziness": "AUTO",
			},
		}
		mustQueries = append(mustQueries, mainQuery)
	}

	// 添加分类过滤
	if len(req.CategoryCodes) > 0 {
		categoryFilter := map[string]interface{}{
			"term": map[string]interface{}{
				"category.code": req.CategoryCodes,
			},
		}
		mustQueries = append(mustQueries, categoryFilter)
	}

	// 添加法律名称过滤
	if len(req.LawNames) > 0 {
		lawFilter := map[string]interface{}{
			"terms": map[string]interface{}{
				"law_name.keyword": req.LawNames,
			},
		}
		mustQueries = append(mustQueries, lawFilter)
	}

	// 添加状态过滤
	if req.Status != "" {
		statusFilter := map[string]interface{}{
			"term": map[string]interface{}{
				"status": req.Status,
			},
		}
		mustQueries = append(mustQueries, statusFilter)
	}

	// 添加生效日期过滤
	if req.EffectiveDateFrom != "" || req.EffectiveDateTo != "" {
		dateRange := map[string]interface{}{
			"range": map[string]interface{}{
				"effective_date": map[string]interface{}{},
			},
		}
		if req.EffectiveDateFrom != "" {
			dateRange["range"].(map[string]interface{})["effective_date"].(map[string]interface{})["gte"] = req.EffectiveDateFrom
		}
		if req.EffectiveDateTo != "" {
			dateRange["range"].(map[string]interface{})["effective_date"].(map[string]interface{})["lte"] = req.EffectiveDateTo
		}
		mustQueries = append(mustQueries, dateRange)
	}

	// 添加标签过滤
	if len(req.Tags) > 0 {
		tagsFilter := map[string]interface{}{
			"terms": map[string]interface{}{
				"tags": req.Tags,
			},
		}
		mustQueries = append(mustQueries, tagsFilter)
	}

	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = mustQueries

	// 添加排序
	if req.SortBy != "" {
		sortField := req.SortBy
		if sortField == "relevance" {
			sortField = "_score"
		}

		sortOrder := "desc"
		if req.SortOrder == "asc" {
			sortOrder = "asc"
		}

		query["sort"] = []map[string]interface{}{
			{
				sortField: map[string]interface{}{
					"order": sortOrder,
				},
			},
		}
	}

	// 添加高亮
	if req.Highlight {
		query["highlight"] = map[string]interface{}{
			"fields": map[string]interface{}{
				"title": map[string]interface{}{
					"fragment_size":       150,
					"number_of_fragments": 3,
				},
				"content": map[string]interface{}{
					"fragment_size":       200,
					"number_of_fragments": 5,
				},
			},
			"pre_tags":  []string{"<mark>"},
			"post_tags": []string{"</mark>"},
		}
	}

	return query, nil
}

// parseSearchResponse 解析搜索响应
func (m *LegalStatuteIndexManager) parseSearchResponse(result map[string]interface{}, req *LegalSearchRequest) (*LegalSearchResponse, error) {
	hits := result["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})["value"].(float64)
	hitsList := hits["hits"].([]interface{})

	documents := make([]LegalStatuteDocument, 0, len(hitsList))
	for _, hit := range hitsList {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		var doc LegalStatuteDocument
		sourceJSON, _ := json.Marshal(source)
		json.Unmarshal(sourceJSON, &doc)

		documents = append(documents, doc)
	}

	response := &LegalSearchResponse{
		Total:     int64(total),
		Page:      req.Page,
		PageSize:  req.PageSize,
		Documents: documents,
	}

	return response, nil
}

// loadMapping 加载索引映射配置
func (m *LegalStatuteIndexManager) loadMapping() (string, error) {
	// 这里应该从配置文件加载映射，为了简化直接返回
	return `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"legal_analyzer": {
						"type": "custom",
						"tokenizer": "ik_max_word",
						"filter": ["lowercase", "stop"]
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": {"type": "integer"},
				"statute_number": {
					"type": "text",
					"analyzer": "legal_analyzer",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"title": {
					"type": "text",
					"analyzer": "legal_analyzer",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"content": {
					"type": "text",
					"analyzer": "legal_analyzer"
				},
				"law_name": {
					"type": "text",
					"analyzer": "legal_analyzer",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"category": {
					"type": "object",
					"properties": {
						"id": {"type": "integer"},
						"name": {"type": "text", "analyzer": "legal_analyzer"},
						"code": {"type": "keyword"}
					}
				},
				"effective_date": {"type": "date"},
				"status": {"type": "keyword"},
				"tags": {"type": "keyword"},
				"keywords": {"type": "text", "analyzer": "legal_analyzer"}
			}
		}
	}`, nil
}
