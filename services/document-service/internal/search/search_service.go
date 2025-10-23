package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// SearchService 搜索服务
type SearchService struct {
	indexManager *IndexManager
	queryBuilder  *QueryBuilder
	docRepo       repositories.DocumentRepository
	userRepo      repositories.UserRepository
	logger        *logrus.Logger
	cache         SearchCache
}

// SearchCache 搜索缓存接口
type SearchCache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Delete(key string) error
}

// SearchResult 搜索结果
type SearchResult struct {
	Documents     []*SearchDocument `json:"documents"`
	Total         int64             `json:"total"`
	Page          int               `json:"page"`
	PageSize      int               `json:"page_size"`
	Query         string            `json:"query"`
	Filters       map[string]interface{} `json:"filters"`
	Aggregations  map[string]interface{} `json:"aggregations,omitempty"`
	Suggestions   []string          `json:"suggestions,omitempty"`
	SearchTime    time.Duration     `json:"search_time"`
	CacheHit      bool              `json:"cache_hit"`
}

// SearchDocument 搜索文档
type SearchDocument struct {
	ID            uint                   `json:"id"`
	UUID          string                 `json:"uuid"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Tags          []string               `json:"tags"`
	TenantID      string                 `json:"tenant_id"`
	CreatedBy     uint                   `json:"created_by"`
	CreatorName   string                 `json:"creator_name"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	MIMEType      string                 `json:"mime_type"`
	Size          int64                  `json:"size"`
	Version       int                    `json:"version"`
	Status        string                 `json:"status"`
	AccessLevel   string                 `json:"access_level"`
	Priority      int                    `json:"priority"`
	FileHash      string                 `json:"file_hash"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Score         float64                `json:"score"`
	Highlights    map[string]string      `json:"highlights,omitempty"`
}

// SearchStats 搜索统计
type SearchStats struct {
	TotalDocuments     int64             `json:"total_documents"`
	DocumentsByType    map[string]int64  `json:"documents_by_type"`
	DocumentsByDate    map[string]int64  `json:"documents_by_date"`
	TopSearches        []SearchQuery    `json:"top_searches"`
	AverageSearchTime  time.Duration     `json:"average_search_time"`
	TotalSearches      int64             `json:"total_searches"`
	CacheHitRate       float64           `json:"cache_hit_rate"`
}

// SearchQuery 搜索查询记录
type SearchQuery struct {
	Query     string    `json:"query"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Count     int64     `json:"count"`
	LastSearched time.Time `json:"last_searched"`
}

// NewSearchService 创建搜索服务
func NewSearchService(
	indexManager *IndexManager,
	queryBuilder *QueryBuilder,
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	logger *logrus.Logger,
	cache SearchCache,
) *SearchService {
	return &SearchService{
		indexManager: indexManager,
		queryBuilder:  queryBuilder,
		docRepo:       docRepo,
		userRepo:      userRepo,
		logger:        logger,
		cache:         cache,
	}
}

// Search 执行搜索
func (ss *SearchService) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	startTime := time.Now()

	// 生成缓存键
	cacheKey := ss.generateCacheKey(req)

	// 尝试从缓存获取结果
	if ss.cache != nil {
		if cached, err := ss.cache.Get(cacheKey); err == nil {
			if result, ok := cached.(*SearchResult); ok {
				result.CacheHit = true
				ss.logger.WithField("query", req.Query).Debug("Search result from cache")
				return result, nil
			}
		}
	}

	// 构建搜索查询
	query, err := ss.queryBuilder.BuildSearchQuery(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build search query: %w", err)
	}

	// 序列化查询
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	// 执行搜索
	searchReq := esapi.SearchRequest{
		Index: []string{ss.indexManager.esClient.GetIndexName()},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := searchReq.Do(ctx, ss.indexManager.esClient.GetClient())
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search failed: %s", res.Status())
	}

	// 解析搜索结果
	searchResult, err := ss.parseSearchResponse(res, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// 计算搜索时间
	searchResult.SearchTime = time.Since(startTime)

	// 缓存结果
	if ss.cache != nil {
		ttl := 5 * time.Minute // 缓存5分钟
		if err := ss.cache.Set(cacheKey, searchResult, ttl); err != nil {
			ss.logger.WithError(err).Warn("Failed to cache search result")
		}
	}

	// 记录搜索统计
	ss.recordSearchStats(req, searchResult.SearchTime)

	return searchResult, nil
}

// Suggest 获取搜索建议
func (ss *SearchService) Suggest(ctx context.Context, query, field string, size int) ([]string, error) {
	suggestQuery, err := ss.queryBuilder.BuildAutoCompleteQuery(query, field, size)
	if err != nil {
		return nil, fmt.Errorf("failed to build suggest query: %w", err)
	}

	queryJSON, err := json.Marshal(suggestQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggest query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{ss.indexManager.esClient.GetIndexName()},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, ss.indexManager.esClient.GetClient())
	if err != nil {
		return nil, fmt.Errorf("suggest request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("suggest failed: %s", res.Status())
	}

	return ss.parseSuggestResponse(res)
}

// GetSimilarDocuments 获取相似文档
func (ss *SearchService) GetSimilarDocuments(ctx context.Context, documentID int64, fields []string, size int) ([]*SearchDocument, error) {
	similarQuery, err := ss.queryBuilder.BuildSimilarDocumentsQuery(documentID, fields, size)
	if err != nil {
		return nil, fmt.Errorf("failed to build similar documents query: %w", err)
	}

	queryJSON, err := json.Marshal(similarQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal similar documents query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{ss.indexManager.esClient.GetIndexName()},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, ss.indexManager.esClient.GetClient())
	if err != nil {
		return nil, fmt.Errorf("similar documents request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("similar documents search failed: %s", res.Status())
	}

	searchResult, err := ss.parseSearchResponse(res, &SearchRequest{
		Page:     1,
		PageSize: size,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse similar documents response: %w", err)
	}

	return searchResult.Documents, nil
}

// Aggregate 执行聚合查询
func (ss *SearchService) Aggregate(ctx context.Context, req *SearchRequest, aggregations map[string]map[string]interface{}) (*SearchResult, error) {
	// 添加聚合到请求
	req.Aggregations = make(map[string]interface{})
	for aggName, aggConfig := range aggregations {
		aggType, _ := aggConfig["type"].(string)
		field, _ := aggConfig["field"].(string)
		size := 10
		if s, ok := aggConfig["size"].(int); ok {
			size = s
		}

		req.Aggregations[aggName] = ss.queryBuilder.BuildAggregationQuery(aggType, field, size)
	}

	return ss.Search(ctx, req)
}

// Reindex 重新索引
func (ss *SearchService) Reindex(ctx context.Context, tenantID string) error {
	ss.logger.WithField("tenant_id", tenantID).Info("Starting reindex process")

	return ss.indexManager.SyncTenantDocuments(ctx, tenantID)
}

// GetSearchStats 获取搜索统计
func (ss *SearchService) GetSearchStats(ctx context.Context, tenantID string) (*SearchStats, error) {
	stats := &SearchStats{
		DocumentsByType: make(map[string]int64),
		DocumentsByDate: make(map[string]int64),
		TopSearches:     []SearchQuery{},
	}

	// 获取索引统计
	indexStats, err := ss.indexManager.GetIndexStats(ctx)
	if err != nil {
		ss.logger.WithError(err).Warn("Failed to get index stats")
	} else {
		stats.TotalDocuments = indexStats.DocumentCount
	}

	// 聚合查询：按类型统计
	typeAgg := map[string]map[string]interface{}{
		"by_type": {
			"type":  "terms",
			"field": "category",
			"size":  20,
		},
	}

	typeResult, err := ss.Aggregate(ctx, &SearchRequest{
		TenantID: tenantID,
		Page:     1,
		PageSize: 0,
	}, typeAgg)

	if err == nil && typeResult.Aggregations != nil {
		if byType, ok := typeResult.Aggregations["by_type"].(map[string]interface{}); ok {
			if buckets, ok := byType["buckets"].([]interface{}); ok {
				for _, bucket := range buckets {
					if bucketMap, ok := bucket.(map[string]interface{}); ok {
						if key, ok := bucketMap["key"].(string); ok {
							if count, ok := bucketMap["doc_count"].(float64); ok {
								stats.DocumentsByType[key] = int64(count)
							}
						}
					}
				}
			}
		}
	}

	// 聚合查询：按日期统计
	dateAgg := map[string]map[string]interface{}{
		"by_date": {
			"type":  "date_histogram",
			"field": "created_at",
		},
	}

	dateResult, err := ss.Aggregate(ctx, &SearchRequest{
		TenantID: tenantID,
		Page:     1,
		PageSize: 0,
	}, dateAgg)

	if err == nil && dateResult.Aggregations != nil {
		if byDate, ok := dateResult.Aggregations["by_date"].(map[string]interface{}); ok {
			if buckets, ok := byDate["buckets"].([]interface{}); ok {
				for _, bucket := range buckets {
					if bucketMap, ok := bucket.(map[string]interface{}); ok {
						if timestamp, ok := bucketMap["key_as_string"].(string); ok {
							if count, ok := bucketMap["doc_count"].(float64); ok {
								// 转换时间戳格式
								if t, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp); err == nil {
									dateKey := t.Format("2006-01-02")
									stats.DocumentsByDate[dateKey] = int64(count)
								}
							}
						}
					}
				}
			}
		}
	}

	return stats, nil
}

// parseSearchResponse 解析搜索响应
func (ss *SearchService) parseSearchResponse(res *esapi.Response, req *SearchRequest) (*SearchResult, error) {
	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	result := &SearchResult{
		Documents:   make([]*SearchDocument, 0),
		Total:       0,
		Page:        req.Page,
		PageSize:    req.PageSize,
		Query:       req.Query,
		Filters:     req.Filters,
		Aggregations: make(map[string]interface{}),
		Suggestions: make([]string, 0),
		CacheHit:    false,
	}

	// 解析命中结果
	if hits, ok := searchResult["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{})["value"].(float64); ok {
			result.Total = int64(total)
		}

		if hitsList, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hitsList {
				if hitMap, ok := hit.(map[string]interface{}); ok {
					doc := ss.parseSearchDocument(hitMap)
					if doc != nil {
						result.Documents = append(result.Documents, doc)
					}
				}
			}
		}
	}

	// 解析聚合结果
	if aggs, ok := searchResult["aggregations"].(map[string]interface{}); ok {
		result.Aggregations = aggs
	}

	// 解析建议结果
	if suggest, ok := searchResult["suggest"].(map[string]interface{}); ok {
		if suggestions := ss.parseSuggestions(suggest); len(suggestions) > 0 {
			result.Suggestions = suggestions
		}
	}

	return result, nil
}

// parseSearchDocument 解析搜索文档
func (ss *SearchService) parseSearchDocument(hit map[string]interface{}) *SearchDocument {
	source, ok := hit["_source"].(map[string]interface{})
	if !ok {
		return nil
	}

	doc := &SearchDocument{
		Metadata: make(map[string]interface{}),
	}

	// 解析基础字段
	if id, ok := source["id"].(float64); ok {
		doc.ID = uint(id)
	}

	if uuid, ok := source["uuid"].(string); ok {
		doc.UUID = uuid
	}

	if name, ok := source["name"].(string); ok {
		doc.Name = name
	}

	if description, ok := source["description"].(string); ok {
		doc.Description = description
	}

	if category, ok := source["category"].(string); ok {
		doc.Category = category
	}

	if tenantID, ok := source["tenant_id"].(string); ok {
		doc.TenantID = tenantID
	}

	if createdBy, ok := source["created_by"].(float64); ok {
		doc.CreatedBy = uint(createdBy)
	}

	if creatorName, ok := source["creator_name"].(string); ok {
		doc.CreatorName = creatorName
	}

	if mimeType, ok := source["mime_type"].(string); ok {
		doc.MIMEType = mimeType
	}

	if size, ok := source["size"].(float64); ok {
		doc.Size = int64(size)
	}

	if version, ok := source["version"].(float64); ok {
		doc.Version = int(version)
	}

	if status, ok := source["status"].(string); ok {
		doc.Status = status
	}

	if accessLevel, ok := source["access_level"].(string); ok {
		doc.AccessLevel = accessLevel
	}

	if priority, ok := source["priority"].(float64); ok {
		doc.Priority = int(priority)
	}

	if fileHash, ok := source["file_hash"].(string); ok {
		doc.FileHash = fileHash
	}

	// 解析时间字段
	if createdAtStr, ok := source["created_at"].(string); ok {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			doc.CreatedAt = createdAt
		}
	}

	if updatedAtStr, ok := source["updated_at"].(string); ok {
		if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			doc.UpdatedAt = updatedAt
		}
	}

	// 解析标签
	if tags, ok := source["tags"].([]interface{}); ok {
		doc.Tags = make([]string, len(tags))
		for i, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				doc.Tags[i] = tagStr
			}
		}
	}

	// 解析元数据
	if metadata, ok := source["metadata"].(map[string]interface{}); ok {
		doc.Metadata = metadata
	}

	// 解析分数
	if score, ok := hit["_score"].(float64); ok {
		doc.Score = score
	}

	// 解析高亮
	if highlight, ok := hit["highlight"].(map[string]interface{}); ok {
		doc.Highlights = make(map[string]string)
		for field, highlights := range highlight {
			if highlightList, ok := highlights.([]interface{}); ok && len(highlightList) > 0 {
				if highlightText, ok := highlightList[0].(string); ok {
					doc.Highlights[field] = highlightText
				}
			}
		}
	}

	return doc
}

// parseSuggestResponse 解析建议响应
func (ss *SearchService) parseSuggestResponse(res *esapi.Response) ([]string, error) {
	var suggestResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&suggestResult); err != nil {
		return nil, fmt.Errorf("failed to decode suggest response: %w", err)
	}

	if suggest, ok := suggestResult["suggest"].(map[string]interface{}); ok {
		return ss.parseSuggestions(suggest), nil
	}

	return []string{}, nil
}

// parseSuggestions 解析建议
func (ss *SearchService) parseSuggestions(suggest map[string]interface{}) []string {
	var suggestions []string

	for _, suggestType := range suggest {
		if options, ok := suggestType.(map[string]interface{})["options"].([]interface{}); ok {
			for _, option := range options {
				if optionMap, ok := option.(map[string]interface{}); ok {
					if text, ok := optionMap["text"].(string); ok {
						suggestions = append(suggestions, text)
					}
				}
			}
		}
	}

	return suggestions
}

// generateCacheKey 生成缓存键
func (ss *SearchService) generateCacheKey(req *SearchRequest) string {
	key := fmt.Sprintf("search:%s:%s:%d:%d", req.TenantID, req.Query, req.Page, req.PageSize)

	// 添加过滤器到缓存键
	if len(req.Filters) > 0 {
		filtersJSON, _ := json.Marshal(req.Filters)
		key += ":" + string(filtersJSON)
	}

	return key
}

// recordSearchStats 记录搜索统计
func (ss *SearchService) recordSearchStats(req *SearchRequest, searchTime time.Duration) {
	// 这里可以实现搜索统计记录，比如：
	// 1. 记录热门搜索词
	// 2. 记录平均搜索时间
	// 3. 记录用户搜索行为
	// 目前简化处理，只记录日志

	ss.logger.WithFields(logrus.Fields{
		"query":       req.Query,
		"tenant_id":   req.TenantID,
		"user_id":     req.UserID,
		"search_time": searchTime.Milliseconds(),
		"filters":     len(req.Filters),
	}).Debug("Search performed")
}