package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// SearchAnalytics 搜索分析器
type SearchAnalytics struct {
	esClient      *ElasticsearchClient
	docRepo       repositories.DocumentRepository
	userRepo      repositories.UserRepository
	logger        *logrus.Logger
	metricsTTL   time.Duration
}

// NewSearchAnalytics 创建搜索分析器
func NewSearchAnalytics(
	esClient *ElasticsearchClient,
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	logger *logrus.Logger,
) *SearchAnalytics {
	return &SearchAnalytics{
		esClient:    esClient,
		docRepo:    docRepo,
		userRepo:   userRepo,
		logger:     logger,
		metricsTTL: 24 * time.Hour,
	}
}

// SearchMetrics 搜索指标
type SearchMetrics struct {
	TotalSearches       int64     `json:"total_searches"`
	UniqueUsers         int64     `json:"unique_users"`
	TopQueries          []QueryMetrics `json:"top_queries"`
	TopCategories       []CategoryMetrics `json:"top_categories"`
	SearchTimeMetrics    TimeMetrics `json:"search_time_metrics"`
	CacheMetrics         CacheMetrics `json:"cache_metrics"`
	DocumentMetrics      DocumentMetrics `json:"document_metrics"`
	UserBehaviorMetrics  UserBehaviorMetrics `json:"user_behavior_metrics"`
	Period               Period     `json:"period"`
	GeneratedAt          time.Time  `json:"generated_at"`
}

// QueryMetrics 查询指标
type QueryMetrics struct {
	Query     string    `json:"query"`
	Count     int64     `json:"count"`
	Users     int64     `json:"users"`
	AvgTime   float64   `json:"avg_time_ms"`
	LastSearched time.Time `json:"last_searched"`
}

// CategoryMetrics 分类指标
type CategoryMetrics struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
	Docs     int64  `json:"documents"`
}

// TimeMetrics 时间指标
type TimeMetrics struct {
	AvgTime  float64 `json:"avg_time_ms"`
	MinTime  float64 `json:"min_time_ms"`
	MaxTime  float64 `json:"max_time_ms"`
	P50Time  float64 `json:"p50_time_ms"`
	P95Time  float64 `json:"p95_time_ms"`
	P99Time  float64 `json:"p99_time_ms"`
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	HitRate      float64 `json:"hit_rate"`
	Hits         int64   `json:"hits"`
	Misses       int64   `json:"misses"`
	AvgHitTime   float64 `json:"avg_hit_time_ms"`
	AvgMissTime  float64 `json:"avg_miss_time_ms"`
}

// DocumentMetrics 文档指标
type DocumentMetrics struct {
	TotalDocs      int64             `json:"total_docs"`
	IndexedDocs   int64             `json:"indexed_docs"`
	DocsByType     map[string]int64  `json:"docs_by_type"`
	DocsBySize     map[string]int64  `json:"docs_by_size"`
	DocsByDate     map[string]int64  `json:"docs_by_date"`
	IndexSize      int64             `json:"index_size"`
}

// UserBehaviorMetrics 用户行为指标
type UserBehaviorMetrics struct {
	ActiveUsers        int64              `json:"active_users"`
	SearchesPerUser    float64            `json:"searches_per_user"`
	UsersWithZeroSearch int64              `json:"users_with_zero_search"`
	UsersWithManySearches int64             `json:"users_with_many_searches"`
	UserRetention      map[string]float64 `json:"user_retention"`
}

// Period 分析周期
type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Label string   `json:"label"`
}

// GetSearchMetrics 获取搜索指标
func (sa *SearchAnalytics) GetSearchMetrics(ctx context.Context, period *Period) (*SearchMetrics, error) {
	metrics := &SearchMetrics{
	Period:      *period,
	GeneratedAt: time.Now(),
	}

	// 获取搜索日志聚合数据
	searchAgg, err := sa.getSearchAggregation(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get search aggregation: %w", err)
	}

	// 获取文档统计
	docStats, err := sa.getDocumentsStats(ctx)
	if err != nil {
		sa.logger.WithError(err).Warn("Failed to get document stats")
	} else {
		metrics.DocumentMetrics = *docStats
	}

	// 获取用户行为统计
	userBehavior, err := sa.getUserBehaviorStats(ctx, period)
	if err != nil {
		sa.logger.WithError(err).Warn("Failed to get user behavior stats")
	} else {
		metrics.UserBehaviorMetrics = *userBehavior
	}

	// 解析搜索聚合数据
	if searchAgg != nil {
		metrics.TotalSearches = sa.getTotalSearches(searchAgg)
		metrics.UniqueUsers = sa.getUniqueUsers(searchAgg)
		metrics.TopQueries = sa.getTopQueries(searchAgg)
		metrics.TopCategories = sa.getTopCategories(searchAgg)
		metrics.SearchTimeMetrics = sa.getSearchTimeMetrics(searchAgg)
		metrics.CacheMetrics = sa.getCacheMetrics(searchAgg)
	}

	return metrics, nil
}

// getSearchAggregation 获取搜索聚合数据
func (sa *SearchAnalytics) getSearchAggregation(ctx context.Context, period *Period) (map[string]interface{}, error) {
	// 构建搜索日志查询（假设有搜索日志索引）
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": period.Start.Format("2006-01-02T15:04:05.000Z"),
					"lte": period.End.Format("2006-01-02T15:04:05.000Z"),
				},
			},
		},
		"aggs": map[string]interface{}{
			"top_queries": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "query.keyword",
									"size":  20,
									"order": map[string]interface{}{
						"_count": "desc",
					},
				},
			},
			"search_by_hour": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":           "timestamp",
					"calendar_interval": "hour",
									"format":          "yyyy-MM-dd HH:mm",
				},
			},
			"search_time_stats": map[string]interface{}{
				"stats": map[string]interface{}{
					"field": "search_time_ms",
				},
			},
			"cache_stats": map[string]interface{}{
				"filters": map[string]interface{}{
					"cache_hit": map[string]interface{}{
						"term": map[string]interface{}{
							"value": true,
						},
					},
				},
				"aggs": map[string]interface{}{
					"cache_hit_time": map[string]interface{}{
						"stats": map[string]interface{}{
							"field": "response_time_ms",
						},
					},
					"cache_miss_time": map[string]interface{}{
						"stats": map[string]interface{}{
							"field": "response_time_ms",
						},
					},
				},
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search aggregation query: %w", err)
	}

	// 这里假设有一个搜索日志索引
	// 实际实现中需要先创建搜索日志记录机制
	return sa.executeAnalyticsQuery(ctx, queryJSON)
}

// getDocumentsStats 获取文档统计
func (sa *SearchAnalytics) getDocumentsStats(ctx context.Context) (*DocumentMetrics, error) {
	// 获取索引统计
	indexStats, err := sa.esClient.GetIndexStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get index stats: %w", err)
	}

	docMetrics := &DocumentMetrics{
		TotalDocs:    indexStats.DocumentCount,
		IndexedDocs: indexStats.DocumentCount,
		IndexSize:   indexStats.StoreSize,
		DocsByType:   make(map[string]int64),
		DocsBySize:   make(map[string]int64),
		DocsByDate:   make(map[string]int64),
	}

	// 获取文档类型和大小分布
	query := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
		"docs_by_type": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "mime_type",
				"size":  20,
			},
		},
		"docs_by_size": map[string]interface{}{
			"range": map[string]interface{}{
				"field": "size",
				"ranges": []map[string]interface{}{
					{"key": "small", "to": 1024 * 1024},
					{"key": "medium", "from": 1024 * 1024, "to": 10 * 1024 * 1024},
					{"key": "large", "from": 10 * 1024 * 1024},
				},
			},
		},
		"docs_by_date": map[string]interface{}{
			"date_histogram": map[string]interface{}{
				"field":           "created_at",
				"calendar_interval": "day",
				"format":          "yyyy-MM-dd",
			},
		},
	},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal docs stats query: %w", err)
	}

	result, err := sa.executeAnalyticsQuery(ctx, queryJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get docs stats: %w", err)
	}

	// 解析聚合结果
	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		if byType, ok := aggs["docs_by_type"].(map[string]interface{})["buckets"].([]interface{}); ok {
			for _, bucket := range byType {
				if bucketMap, ok := bucket.(map[string]interface{}); ok {
					if key, ok := bucketMap["key"].(string); ok {
						if count, ok := bucketMap["doc_count"].(float64); ok {
							docMetrics.DocsByType[key] = int64(count)
						}
					}
				}
			}
		}

		if bySize, ok := aggs["docs_by_size"].(map[string]interface{})["buckets"].([]interface{}); ok {
			for _, bucket := range bySize {
				if bucketMap, ok := bucket.(map[string]interface{}); ok {
					if key, ok := bucketMap["key"].(string); ok {
						if count, ok := bucketMap["doc_count"].(float64); ok {
							docMetrics.DocsBySize[key] = int64(count)
						}
					}
				}
			}
		}

		if byDate, ok := aggs["docs_by_date"].(map[string{})["buckets"].([]interface{}); ok {
			for _, bucket := range byDate {
				if bucketMap, ok := bucket.(map[string]interface{}); ok {
					if key, ok := bucketMap["key_as_string"].(string); ok {
						if count, ok := bucketMap["doc_count"].(float64); ok {
							docMetrics.DocsByDate[key] = int64(count)
						}
					}
				}
			}
		}
	}

	return docMetrics, nil
}

// getUserBehaviorStats 获取用户行为统计
func (sa *SearchAnalytics) getUserBehaviorStats(ctx context.Context, period *Period) (*UserBehaviorMetrics, error) {
	userBehavior := &UserBehaviorMetrics{
		UserRetention: make(map[string]float64),
	}

	// 这里简化处理，实际应该从搜索日志中分析用户行为
	// 获取活跃用户数
	activeUsers := int64(0) // 实际实现需要从数据库查询
	usersWithZeroSearch := int64(0)
	usersWithManySearches := int64(0)

	userBehavior.ActiveUsers = activeUsers
	userBehavior.UsersWithZeroSearch = usersWithZeroSearch
	userBehavior.UsersWithManySearches = usersWithManySearches

	// 计算平均搜索次数
	if activeUsers > 0 {
		totalSearches := sa.getTotalSearches(map[string]interface{}{}) // 需要从搜索聚合获取
		userBehavior.SearchesPerUser = float64(totalSearches) / float64(activeUsers)
	}

	return userBehavior, nil
}

// executeAnalyticsQuery 执行分析查询
func (sa *SearchAnalytics) executeAnalyticsQuery(ctx context.Context, queryJSON []byte) (map[string]interface{}, error) {
	// 这里需要根据实际的索引名称和查询逻辑调整
	// 暂时返回空结果，实际实现需要完整的搜索日志索引

	sa.logger.Debug("Executing analytics query")
	return make(map[string]interface{}), nil
}

// getTotalSearches 获取总搜索次数
func (sa *SearchAnalytics) getTotalSearches(aggregation map[string]interface{}) int64 {
	if total, ok := aggregation["total"].(float64); ok {
		return int64(total)
	}
	return 0
}

// getUniqueUsers 获取独立用户数
func (sa *SearchAnalytics) getUniqueUsers(aggregation map[string]interface{}) int64 {
	if users, ok := aggregation["unique_users"].(float64); ok {
		return int64(users)
	}
	return 0
}

// getTopQueries 获取热门查询
func (sa *SearchAnalytics) getTopQueries(aggregation map[string]interface{}) []QueryMetrics {
	var queries []QueryMetrics

	if topQueries, ok := aggregation["top_queries"].(map[string]interface{})["buckets"].([]interface{}); ok {
		for _, bucket := range topQueries {
			if bucketMap, ok := bucket.(map[string]interface{}); ok {
				query := QueryMetrics{
					Query: sa.getString(bucketMap, "key"),
					Count: sa.getInt64(bucketMap, "doc_count"),
				}
				queries = append(queries, query)
			}
		}
	}

	return queries
}

// getTopCategories 获取热门分类
func (sa *SearchAnalytics) getTopCategories(aggregation map[string]interface{}) []CategoryMetrics {
	var categories []CategoryMetrics

	if topCategories, ok := aggregation["top_categories"].(map[string]interface{})["buckets"].([]interface{}); ok {
		for _, bucket := range topCategories {
			if bucketMap, ok := bucket.(map[string]interface{}); ok {
				category := CategoryMetrics{
					Category: sa.getString(bucketMap, "key"),
					Count:    sa.getInt64(bucketMap, "doc_count"),
				}
				categories = append(categories, category)
			}
		}
	}

	return categories
}

// getSearchTimeMetrics 获取搜索时间指标
func (sa *SearchAnalytics) getSearchTimeMetrics(aggregation map[string]interface{}) TimeMetrics {
	timeMetrics := TimeMetrics{}

	if stats, ok := aggregation["search_time_stats"].(map[string{})["stats"].(map[string]interface{}); ok {
		timeMetrics.AvgTime = sa.getFloat64(stats, "avg")
		timeMetrics.MinTime = sa.getFloat64(stats, "min")
		timeMetrics.MaxTime = sa.getFloat64(stats, "max")
		timeMetrics.P50Time = sa.getFloat64(stats, "p50")
		timeMetrics.P95Time = sa.getFloat64(stats, "p95")
		timeMetrics.P99Time = sa.getFloat64(stats, "p99")
	}

	return timeMetrics
}

// getCacheMetrics 获取缓存指标
func (sa *SearchAnalytics) getCacheMetrics(aggregation map[string]interface{}) CacheMetrics {
	cacheMetrics := CacheMetrics{
		Hits:   0,
		Misses: 0,
	}

	// 获取命中和未命中次数
	if filters, ok := aggregation["cache_stats"].(map[string{})["filters"].(map[string{})["cache_hit"].(map[string{})["doc_count"].(float64)); ok {
		cacheMetrics.Hits = int64(filters)
	}

	if filters, ok := aggregation["cache_stats"].(map[string{})["filters"].(map[string{})["cache_miss"].(map[string{})["doc_count"].(float64)); ok {
		cacheMetrics.Misses = int64(filters)
	}

	// 计算命中率
	if cacheMetrics.Hits > 0 || cacheMetrics.Misses > 0 {
		cacheMetrics.HitRate = float64(cacheMetrics.Hits) / float64(cacheMetrics.Hits+cacheMetrics.Misses)
	}

	// 获取缓存时间指标
	if aggs, ok := aggregation["cache_stats"].(map[string{})["aggs"]; ok {
		if hitStats, ok := aggs["cache_hit_time"].(map[string{})["stats"].(map[string]interface{}); ok {
			cacheMetrics.AvgHitTime = sa.getFloat64(hitStats, "avg")
		}

		if missStats, ok := aggs["cache_miss_time"].(map[string{})["stats"].(map[string]interface{}); ok {
			cacheMetrics.AvgMissTime = sa.getFloat64(missStats, "avg")
		}
	}

	return cacheMetrics
}

// 辅助方法
func (sa *SearchAnalytics) getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func (sa *SearchAnalytics) getInt64(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(float64); ok {
		return int64(val)
	}
	return 0
}

func (sa *SearchAnalytics) getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

// RecordSearchEvent 记录搜索事件
func (sa *SearchAnalytics) RecordSearchEvent(ctx context.Context, event *SearchEvent) error {
	// 实现搜索事件记录逻辑
	sa.logger.WithFields(logrus.Fields{
		"query":    event.Query,
		"user_id":   event.UserID,
		"tenant_id": event.TenantID,
		"cache_hit": event.CacheHit,
		"response_time": event.ResponseTime,
	}).Debug("Recording search event")

	// 这里应该将事件发送到搜索日志索引或消息队列
	// 实际实现需要集成事件处理系统

	return nil
}

// SearchEvent 搜索事件
type SearchEvent struct {
	Query       string        `json:"query"`
	UserID      string        `json:"user_id"`
	TenantID    string        `json:"tenant_id"`
	CacheHit    bool          `json:"cache_hit"`
	ResponseTime time.Duration `json:"response_time"`
	Timestamp   time.Time     `json:"timestamp"`
	Filters     map[string]interface{} `json:"filters"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
}

// GetRealTimeMetrics 获取实时指标
func (sa *SearchAnalytics) GetRealTimeMetrics(ctx context.Context) (*RealTimeMetrics, error) {
	metrics := &RealTimeMetrics{
		CurrentSearchRate: 0,
		ActiveSearches:   0,
		AvgResponseTime:  0,
		CacheHitRate:     0,
		GeneratedAt:      time.Now(),
	}

	// 实现实时指标收集逻辑
	// 这里可以通过查询Redis缓存中的实时数据来获取指标

	return metrics, nil
}

// RealTimeMetrics 实时指标
type RealTimeMetrics struct {
	CurrentSearchRate float64   `json:"current_search_rate"`
	ActiveSearches   int64       `json:"active_searches"`
	AvgResponseTime  float64   `json:"avg_response_time"`
	CacheHitRate     float64   `json:"cache_hit_rate"`
	GeneratedAt      time.Time  `json:"generated_at"`
}