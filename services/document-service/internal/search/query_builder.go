package search

import (
	"fmt"
	"strings"
	"time"
)

// SearchRequest 搜索请求
type SearchRequest struct {
	Query          string                 `json:"query"`
	Filters        map[string]interface{} `json:"filters"`
	TenantID       string                 `json:"tenant_id"`
	UserID         string                 `json:"user_id,omitempty"`
	SortBy         []SortField            `json:"sort_by,omitempty"`
	Page           int                    `json:"page"`
	PageSize       int                    `json:"page_size"`
	Highlight      bool                   `json:"highlight,omitempty"`
	Aggregations   map[string]interface{} `json:"aggregations,omitempty"`
	Suggest        bool                   `json:"suggest,omitempty"`
	SuggestSize    int                    `json:"suggest_size,omitempty"`
	SearchFields   []string               `json:"search_fields,omitempty"`
	IncludeContent bool                  `json:"include_content,omitempty"`
}

// SortField 排序字段
type SortField struct {
	Field    string `json:"field"`
	Order    string `json:"order"`    // asc, desc
	Mode     string `json:"mode,omitempty"`    // min, max, sum, avg
	NestedPath string `json:"nested_path,omitempty"` // 嵌套路径
}

// Filter 过滤器
type Filter struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`  // eq, ne, gt, gte, lt, lte, in, nin, exists, range
	Value     interface{} `json:"value"`
	Values    []interface{} `json:"values,omitempty"`
	Boost     float64     `json:"boost,omitempty"`
}

// QueryBuilder 查询构建器
type QueryBuilder struct {
	indexManager *IndexManager
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(indexManager *IndexManager) *QueryBuilder {
	return &QueryBuilder{
		indexManager: indexManager,
	}
}

// BuildSearchQuery 构建搜索查询
func (qb *QueryBuilder) BuildSearchQuery(req *SearchRequest) (map[string]interface{}, error) {
	if err := qb.validateSearchRequest(req); err != nil {
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	// 构建基础查询
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   []interface{}{},
				"filter": []interface{}{},
			},
		},
	}

	boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
	mustQueries := boolQuery["must"].([]interface{})
	filterQueries := boolQuery["filter"].([]interface{})

	// 1. 添加租户过滤（强制）
	tenantFilter := map[string]interface{}{
		"term": map[string]interface{}{
			"tenant_id": req.TenantID,
		},
	}
	filterQueries = append(filterQueries, tenantFilter)

	// 2. 添加搜索条件
	if req.Query != "" {
		searchQuery := qb.buildSearchQuery(req.Query, req.SearchFields)
		mustQueries = append(mustQueries, searchQuery)
	}

	// 3. 添加过滤器
	for field, value := range req.Filters {
		filter := qb.buildFilter(field, value)
		if filter != nil {
			filterQueries = append(filterQueries, filter)
		}
	}

	// 4. 添加权限过滤（如果有用户ID）
	if req.UserID != "" {
		permissionFilter := qb.buildPermissionFilter(req.UserID)
		if permissionFilter != nil {
			filterQueries = append(filterQueries, permissionFilter)
		}
	}

	// 更新查询
	boolQuery["must"] = mustQueries
	boolQuery["filter"] = filterQueries

	// 5. 添加排序
	if len(req.SortBy) > 0 {
		query["sort"] = qb.buildSortFields(req.SortBy)
	} else {
		// 默认排序：按相关性分数，然后按更新时间
		query["sort"] = []interface{}{
			map[string]interface{}{
				"_score": map[string]interface{}{
					"order": "desc",
				},
			},
			map[string]interface{}{
				"updated_at": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	}

	// 6. 添加分页
	from := (req.Page - 1) * req.PageSize
	query["from"] = from
	query["size"] = req.PageSize

	// 7. 添加高亮
	if req.Highlight {
		query["highlight"] = qb.buildHighlight()
	}

	// 8. 添加聚合
	if len(req.Aggregations) > 0 {
		query["aggs"] = req.Aggregations
	}

	// 9. 添加建议
	if req.Suggest && req.Query != "" {
		query["suggest"] = qb.buildSuggest(req.Query, req.SuggestSize)
	}

	return query, nil
}

// buildSearchQuery 构建搜索查询
func (qb *QueryBuilder) buildSearchQuery(query string, fields []string) map[string]interface{} {
	if len(fields) == 0 {
		// 默认搜索字段
		fields = []string{
			"name^3",
			"description^2",
			"creator_name",
			"tags^2",
			"category",
		}
	}

	// 构建多字段查询
	searchQuery := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  query,
			"fields": fields,
			"type":   "best_fields",
			"fuzziness": "AUTO",
			"prefix_length": 1,
			"max_expansions": 50,
			"operator": "and",
		},
	}

	// 如果查询较短，添加模糊匹配和自动建议
	if len(strings.TrimSpace(query)) <= 3 {
		searchQuery["multi_match"].(map[string]interface{})["type"] = "bool_prefix"
		searchQuery["multi_match"].(map[string]interface{})["operator"] = "or"
	}

	return searchQuery
}

// buildFilter 构建过滤器
func (qb *QueryBuilder) buildFilter(field string, value interface{}) map[string]interface{} {
	switch field {
	case "category":
		return map[string]interface{}{
			"term": map[string]interface{}{
				"category": value,
			},
		}
	case "tags":
		if values, ok := value.([]interface{}); ok {
			return map[string]interface{}{
				"terms": map[string]interface{}{
					"tags": values,
				},
			}
		} else if tag, ok := value.(string); ok {
			return map[string]interface{}{
				"term": map[string]interface{}{
					"tags": tag,
				},
			}
		}
	case "created_by":
		return map[string]interface{}{
			"term": map[string]interface{}{
				"created_by": value,
			},
		}
	case "mime_type":
		if values, ok := value.([]interface{}); ok {
			return map[string]interface{}{
				"terms": map[string]interface{}{
					"mime_type": values,
				},
			}
		} else {
			return map[string]interface{}{
				"term": map[string]interface{}{
					"mime_type": value,
				},
			}
		}
	case "date_range":
		if dateRange, ok := value.(map[string]interface{}); ok {
			rangeQuery := map[string]interface{}{
				"range": map[string]interface{}{
					"created_at": map[string]interface{}{},
				},
			}

			if start, ok := dateRange["start"]; ok {
				rangeQuery["range"].(map[string]interface{})["created_at"].(map[string]interface{})["gte"] = start
			}
			if end, ok := dateRange["end"]; ok {
				rangeQuery["range"].(map[string]interface{})["created_at"].(map[string]interface{})["lte"] = end
			}

			return rangeQuery
		}
	case "size_range":
		if sizeRange, ok := value.(map[string]interface{}); ok {
			rangeQuery := map[string]interface{}{
				"range": map[string]interface{}{
					"size": map[string]interface{}{},
				},
			}

			if min, ok := sizeRange["min"]; ok {
				rangeQuery["range"].(map[string]interface{})["size"].(map[string]interface{})["gte"] = min
			}
			if max, ok := sizeRange["max"]; ok {
				rangeQuery["range"].(map[string]interface{})["size"].(map[string]interface{})["lte"] = max
			}

			return rangeQuery
		}
	}

	return nil
}

// buildPermissionFilter 构建权限过滤器
func (qb *QueryBuilder) buildPermissionFilter(userID string) map[string]interface{} {
	// 这里简化处理，实际应该根据用户权限构建复杂的权限查询
	// 目前假设用户可以访问自己创建的文档和有权限的文档

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"should": []interface{}{
				map[string]interface{}{
					"term": map[string]interface{}{
						"created_by": userID,
					},
				},
				// 这里可以添加更复杂的权限查询
				// 例如：查询用户有访问权限的文档
			},
		},
	}
}

// buildSortFields 构建排序字段
func (qb *QueryBuilder) buildSortFields(sortFields []SortField) []interface{} {
	var sorts []interface{}

	for _, sortField := range sortFields {
		sort := map[string]interface{}{
			sortField.Field: map[string]interface{}{
				"order": sortField.Order,
			},
		}

		if sortField.Mode != "" {
			sort[sortField.Field].(map[string]interface{})["mode"] = sortField.Mode
		}

		if sortField.NestedPath != "" {
			sort[sortField.Field].(map[string]interface{})["nested_path"] = sortField.NestedPath
		}

		sorts = append(sorts, sort)
	}

	return sorts
}

// buildHighlight 构建高亮配置
func (qb *QueryBuilder) buildHighlight() map[string]interface{} {
	return map[string]interface{}{
		"pre_tags":  []string{"<mark>"},
		"post_tags": []string{"</mark>"},
		"fields": map[string]interface{}{
			"name": map[string]interface{}{
				"fragment_size": 150,
				"number_of_fragments": 3,
			},
			"description": map[string]interface{}{
				"fragment_size": 200,
				"number_of_fragments": 2,
			},
			"creator_name": map[string]interface{}{
				"fragment_size": 100,
				"number_of_fragments": 1,
			},
		},
		"require_field_match": false,
		"encoder": "html",
	}
}

// buildSuggest 构建建议配置
func (qb *QueryBuilder) buildSuggest(query string, size int) map[string]interface{} {
	if size <= 0 {
		size = 5
	}

	return map[string]interface{}{
		"text": query,
		"simple_phrase": map[string]interface{}{
			"phrase": map[string]interface{}{
				"field": "name.edge_ngram",
				"size":  size,
				"gram_size": 1,
				"direct_generator": []map[string]interface{}{
					{
						"field": "name.edge_ngram",
						"suggest_mode": "missing",
					},
				},
			},
			"collate": map[string]interface{}{
				"query": map[string]interface{}{
					"source": map[string]interface{}{
						"match": map[string]interface{}{
							"{{field_name}}": "{{suggestion}}",
						},
					},
				},
				"prune": true,
			},
		},
	}
}

// validateSearchRequest 验证搜索请求
func (qb *QueryBuilder) validateSearchRequest(req *SearchRequest) error {
	if req.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return nil
}

// BuildAggregationQuery 构建聚合查询
func (qb *QueryBuilder) BuildAggregationQuery(aggType string, field string, size int) map[string]interface{} {
	if size <= 0 {
		size = 10
	}

	switch aggType {
	case "terms":
		return map[string]interface{}{
			"terms": map[string]interface{}{
				"field": field,
				"size":  size,
				"order": map[string]interface{}{
					"_count": "desc",
				},
			},
		}
	case "date_histogram":
		return map[string]interface{}{
			"date_histogram": map[string]interface{}{
				"field":          field,
				"calendar_interval": "month",
				"format":         "yyyy-MM",
				"min_doc_count":  1,
			},
		}
	case "range":
		return map[string]interface{}{
			"range": map[string]interface{}{
				"field": field,
				"ranges": []map[string]interface{}{
					{
						"key":  "small",
						"from": 0,
						"to":   1024 * 1024, // 1MB
					},
					{
						"key":  "medium",
						"from": 1024 * 1024,
						"to":   10 * 1024 * 1024, // 10MB
					},
					{
						"key":  "large",
						"from": 10 * 1024 * 1024,
					},
				},
			},
		}
	case "stats":
		return map[string]interface{}{
			"stats": map[string]interface{}{
				"field": field,
			},
		}
	default:
		return nil
	}
}

// BuildAutoCompleteQuery 构建自动完成查询
func (qb *QueryBuilder) BuildAutoCompleteQuery(query, field string, size int) (map[string]interface{}, error) {
	if size <= 0 {
		size = 5
	}

	return map[string]interface{}{
		"suggest": map[string]interface{}{
			"prefix": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": field + ".suggest",
					"size":  size,
					"skip_duplicates": true,
				},
			},
		},
	}, nil
}

// BuildSimilarDocumentsQuery 构建相似文档查询
func (qb *QueryBuilder) BuildSimilarDocumentsQuery(documentID int64, fields []string, size int) (map[string]interface{}, error) {
	if size <= 0 {
		size = 10
	}

	if len(fields) == 0 {
		fields = []string{"name", "description", "category"}
	}

	return map[string]interface{}{
		"size": size,
		"query": map[string]interface{}{
			"more_like_this": map[string]interface{}{
				"fields":    fields,
				"like":      []map[string]interface{}{{"_index": qb.indexManager.esClient.GetIndexName(), "_id": fmt.Sprintf("%d", documentID)}},
				"min_term_freq": 1,
				"max_query_terms": 12,
				"min_doc_freq": 1,
			},
		},
	}, nil
}