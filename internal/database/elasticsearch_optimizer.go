package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Elasticsearch性能相关的Prometheus指标
var (
	esIndexLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "es_index_duration_seconds",
		Help:    "Duration of Elasticsearch indexing operations",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"index", "operation"})

	esSearchLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "es_search_duration_seconds",
		Help:    "Duration of Elasticsearch search operations",
		Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"index", "query_type"})

	esIndexCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "es_index_count_total",
		Help: "Total number of Elasticsearch indexing operations",
	}, []string{"index", "operation", "status"})

	esSearchCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "es_search_count_total",
		Help: "Total number of Elasticsearch search operations",
	}, []string{"index", "query_type", "status"})

	esBulkErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "es_bulk_errors_total",
		Help: "Total number of Elasticsearch bulk operation errors",
	}, []string{"operation", "error_type"})

	esConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "es_connections_active",
		Help: "Number of active Elasticsearch connections",
	})
)

// ElasticsearchOptimizer Elasticsearch优化器
type ElasticsearchOptimizer struct {
	client      *elasticsearch.Client
	indexer     esutil.BulkIndexer
	config      *ESConfig
	stats       *ESStats
	httpClient  *http.Client
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	Addresses           []string        `json:"addresses"`
	Username            string          `json:"username"`
	Password            string          `json:"password"`
	MaxRetries          int             `json:"max_retries"`
	RetryOnStatus       []int           `json:"retry_on_status"`
	RetryBackoff        time.Duration   `json:"retry_backoff"`
	CompressRequestBody bool            `json:"compress_request_body"`
	DiscoverNodes       bool            `json:"discover_nodes"`
	DiscoverInterval    time.Duration   `json:"discover_interval"`
	Transport           *http.Transport  `json:"transport"`
	BulkWorkers         int             `json:"bulk_workers"`
	BulkFlushBytes      int             `json:"bulk_flush_bytes"`
	BulkFlushInterval   time.Duration   `json:"bulk_flush_interval"`
	BulkTimeout         time.Duration   `json:"bulk_timeout"`
	EnableMetrics       bool            `json:"enable_metrics"`
}

// ESStats Elasticsearch统计
type ESStats struct {
	TotalIndexOps     int64         `json:"total_index_ops"`
	TotalSearchOps    int64         `json:"total_search_ops"`
	IndexErrors       int64         `json:"index_errors"`
	SearchErrors      int64         `json:"search_errors"`
	BulkErrors        int64         `json:"bulk_errors"`
	AvgIndexLatency   time.Duration `json:"avg_index_latency"`
	AvgSearchLatency  time.Duration `json:"avg_search_latency"`
	TotalIndexTime    time.Duration `json:"total_index_time"`
	TotalSearchTime   time.Duration `json:"total_search_time"`
}

// BulkIndexOptions 批量索引选项
type BulkIndexOptions struct {
	Index          string        `json:"index"`
	Workers        int           `json:"workers"`
	FlushBytes     int           `json:"flush_bytes"`
	FlushInterval  time.Duration `json:"flush_interval"`
	Timeout        time.Duration `json:"timeout"`
	Routing        string        `json:"routing"`
	Pipeline       string        `json:"pipeline"`
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Index          []string      `json:"index"`
	QueryType       string        `json:"query_type"`
	Timeout         time.Duration `json:"timeout"`
	TrackTotalHits  bool          `json:"track_total_hits"`
	PrefetchSearch  bool          `json:"prefetch_search"`
	MinScore        float64       `json:"min_score"`
	MaxResults      int           `json:"max_results"`
	RequestCache    bool          `json:"request_cache"`
}

// DefaultESConfig 默认ES配置
func DefaultESConfig() *ESConfig {
	return &ESConfig{
		MaxRetries:          3,
		RetryOnStatus:       []int{502, 503, 504, 429},
		RetryBackoff:        100 * time.Millisecond,
		CompressRequestBody: true,
		DiscoverNodes:       true,
		DiscoverInterval:    30 * time.Second,
		BulkWorkers:         8,
		BulkFlushBytes:      5 * 1024 * 1024, // 5MB
		BulkFlushInterval:   30 * time.Second,
		BulkTimeout:         60 * time.Second,
		EnableMetrics:       true,
	}
}

// DefaultBulkIndexOptions 默认批量索引选项
func DefaultBulkIndexOptions() BulkIndexOptions {
	return BulkIndexOptions{
		Workers:       8,
		FlushBytes:    5 * 1024 * 1024, // 5MB
		FlushInterval: 30 * time.Second,
		Timeout:       60 * time.Second,
	}
}

// DefaultSearchOptions 默认搜索选项
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Timeout:        10 * time.Second,
		TrackTotalHits: true,
		MaxResults:    10000,
		RequestCache:   true,
	}
}

// NewElasticsearchOptimizer 创建Elasticsearch优化器
func NewElasticsearchOptimizer(cfg *ESConfig) (*ElasticsearchOptimizer, error) {
	if cfg == nil {
		cfg = DefaultESConfig()
	}

	// 创建HTTP传输配置
	transport := &http.Transport{
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: cfg.RetryBackoff,
	}

	if cfg.Transport != nil {
		transport = cfg.Transport
	}

	// 创建Elasticsearch客户端
	esConfig := elasticsearch.Config{
		Addresses:         cfg.Addresses,
		Username:          cfg.Username,
		Password:          cfg.Password,
		MaxRetries:        cfg.MaxRetries,
		RetryOnStatus:     cfg.RetryOnStatus,
		RetryBackoff: func(i int) time.Duration {
			return time.Duration(i) * cfg.RetryBackoff
		},
		CompressRequestBody: cfg.CompressRequestBody,
		DiscoverNodesOnStart: cfg.DiscoverNodes,
		DiscoverNodesInterval: cfg.DiscoverInterval,
		Transport:          transport,
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// 测试连接
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch returned error: %s", res.Status())
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: cfg.BulkTimeout,
		Transport: transport,
	}

	optimizer := &ElasticsearchOptimizer{
		client:     client,
		config:     cfg,
		stats:      &ESStats{},
		httpClient: httpClient,
	}

	// 更新连接指标
	if cfg.EnableMetrics {
		esConnections.Set(1)
	}

	log.Printf("Elasticsearch优化器初始化成功 - 地址: %v, 启用指标: %v", cfg.Addresses, cfg.EnableMetrics)
	return optimizer, nil
}

// InitializeBulkIndexer 初始化批量索引器
func (eso *ElasticsearchOptimizer) InitializeBulkIndexer(index string, options BulkIndexOptions) error {
	if options.Workers <= 0 {
		options.Workers = eso.config.BulkWorkers
	}
	if options.FlushBytes <= 0 {
		options.FlushBytes = eso.config.BulkFlushBytes
	}
	if options.FlushInterval == 0 {
		options.FlushInterval = eso.config.BulkFlushInterval
	}
	if options.Timeout == 0 {
		options.Timeout = eso.config.BulkTimeout
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         index,
		Client:        eso.client,
		NumWorkers:    options.Workers,
		FlushBytes:    options.FlushBytes,
		FlushInterval: options.FlushInterval,
		Timeout:       options.Timeout,

		// 错误处理
		OnError: func(ctx context.Context, err error) {
			eso.recordBulkError("index", err)
			if eso.config.EnableMetrics {
				esBulkErrors.WithLabelValues("index", "general").Inc()
			}
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create bulk indexer: %w", err)
	}

	eso.indexer = indexer
	return nil
}

// BulkIndexDocument 批量索引文档
func (eso *ElasticsearchOptimizer) BulkIndexDocument(ctx context.Context, index string, docs []interface{}, options BulkIndexOptions) error {
	start := time.Now()

	// 确保索引器已初始化
	if eso.indexer == nil {
		if err := eso.InitializeBulkIndexer(index, options); err != nil {
			return err
		}
	}

	// 添加文档到索引器
	for i, doc := range docs {
		data, err := json.Marshal(doc)
		if err != nil {
			eso.recordBulkError("marshal", err)
			continue
		}

		// 提取文档ID（假设文档有ID字段）
		docMap, ok := doc.(map[string]interface{})
		if !ok {
			eso.recordBulkError("type_conversion", fmt.Errorf("document is not a map"))
			continue
		}

		var docID string
		if id, exists := docMap["id"]; exists {
			docID = fmt.Sprintf("%v", id)
		} else {
			docID = fmt.Sprintf("doc_%d", i) // 生成默认ID
		}

		err = eso.indexer.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: docID,
			Body:       bytes.NewReader(data),
			Routing:    options.Routing,
		})

		if err != nil {
			eso.recordBulkError("add_to_indexer", err)
			continue
		}
	}

	// 关闭索引器（触发刷新）
	if err := eso.indexer.Close(ctx); err != nil {
		eso.recordBulkError("close_indexer", err)
		return fmt.Errorf("failed to close bulk indexer: %w", err)
	}

	duration := time.Since(start)
	eso.recordIndexLatency(duration)

	return nil
}

// SearchDocument 搜索文档
func (eso *ElasticsearchOptimizer) SearchDocument(ctx context.Context, options SearchOptions, query interface{}, result interface{}) error {
	start := time.Now()

	if len(options.Index) == 0 {
		return fmt.Errorf("index names are required")
	}

	// 构建搜索请求
	searchBody, err := json.Marshal(map[string]interface{}{
		"query":           query,
		"from":            0,
		"size":            options.MaxResults,
		"min_score":       options.MinScore,
		"request_cache":   options.RequestCache,
		"track_total_hits": options.TrackTotalHits,
	})

	if err != nil {
		eso.recordSearchError("marshal_query", err)
		return fmt.Errorf("failed to marshal search query: %w", err)
	}

	// 执行搜索
	res, err := eso.client.Search(
		eso.client.Search.WithContext(ctx),
		eso.client.Search.WithIndex(options.Index...),
		eso.client.Search.WithBody(bytes.NewReader(searchBody)),
		eso.client.Search.WithTimeout(options.Timeout),
		eso.client.Search.WithTrackTotalHits(options.TrackTotalHits),
		eso.client.Search.WithRequestCache(options.RequestCache),
	)

	if err != nil {
		eso.recordSearchError("search", err)
		return fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		eso.recordSearchError("search_response", fmt.Errorf("Elasticsearch error: %s", res.Status()))
		return fmt.Errorf("Elasticsearch returned error: %s", res.Status())
	}

	// 解析结果
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		eso.recordSearchError("decode_result", err)
		return fmt.Errorf("failed to decode search result: %w", err)
	}

	duration := time.Since(start)
	eso.recordSearchLatency(duration, options.QueryType)

	return nil
}

// SearchWithAggregation 带聚合的搜索
func (eso *ElasticsearchOptimizer) SearchWithAggregation(ctx context.Context, options SearchOptions, query map[string]interface{}, aggregations map[string]interface{}, result interface{}) error {
	start := time.Now()

	if len(options.Index) == 0 {
		return fmt.Errorf("index names are required")
	}

	// 构建搜索请求
	searchQuery := map[string]interface{}{
		"query":           query["query"],
		"from":            0,
		"size":            0, // 不返回文档，只返回聚合结果
		"min_score":       options.MinScore,
		"request_cache":   options.RequestCache,
		"track_total_hits": options.TrackTotalHits,
		"aggs":            aggregations,
	}

	// 如果有分页参数，包含在查询中
	if _, hasSize := query["size"]; hasSize {
		searchQuery["size"] = query["size"]
	}
	if _, hasFrom := query["from"]; hasFrom {
		searchQuery["from"] = query["from"]
	}

	searchBody, err := json.Marshal(searchQuery)
	if err != nil {
		eso.recordSearchError("marshal_aggregation", err)
		return fmt.Errorf("failed to marshal aggregation query: %w", err)
	}

	// 执行搜索
	res, err := eso.client.Search(
		eso.client.Search.WithContext(ctx),
		eso.client.Search.WithIndex(options.Index...),
		eso.client.Search.WithBody(bytes.NewReader(searchBody)),
		eso.client.Search.WithTimeout(options.Timeout),
	)

	if err != nil {
		eso.recordSearchError("search_aggregation", err)
		return fmt.Errorf("aggregation search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		eso.recordSearchError("aggregation_response", fmt.Errorf("Elasticsearch error: %s", res.Status()))
		return fmt.Errorf("Elasticsearch returned error: %s", res.Status())
	}

	// 解析结果
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		eso.recordSearchError("decode_aggregation", err)
		return fmt.Errorf("failed to decode aggregation result: %w", err)
	}

	duration := time.Since(start)
	eso.recordSearchLatency(duration, "aggregation")

	return nil
}

// GetIndexHealth 获取索引健康状态
func (eso *ElasticsearchOptimizer) GetIndexHealth(ctx context.Context, indices []string) (map[string]interface{}, error) {
	res, err := eso.client.Cluster.Health(
		eso.client.Cluster.Health.WithContext(ctx),
		eso.client.Cluster.Health.WithIndex(strings.Join(indices, ",")),
		eso.client.Cluster.Health.WithWaitForStatus("yellow"),
		eso.client.Cluster.Health.WithTimeout(10*time.Second),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get cluster health: %w", err)
	}
	defer res.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return health, nil
}

// GetStats 获取统计信息
func (eso *ElasticsearchOptimizer) GetStats() *ESStats {
	return eso.stats
}

// GetClient 获取Elasticsearch客户端
func (eso *ElasticsearchOptimizer) GetClient() *elasticsearch.Client {
	return eso.client
}

// Close 关闭优化器
func (eso *ElasticsearchOptimizer) Close() error {
	if eso.indexer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := eso.indexer.Close(ctx); err != nil {
			return fmt.Errorf("failed to close bulk indexer: %w", err)
		}
	}

	if eso.config.EnableMetrics {
		esConnections.Set(0)
	}

	return nil
}

// recordIndexSuccess 记录索引成功
func (eso *ElasticsearchOptimizer) recordIndexSuccess() {
	eso.stats.TotalIndexOps++
}

// recordIndexLatency 记录索引延迟
func (eso *ElasticsearchOptimizer) recordIndexLatency(duration time.Duration) {
	eso.stats.TotalIndexTime += duration
	if eso.stats.TotalIndexOps > 0 {
		eso.stats.AvgIndexLatency = eso.stats.TotalIndexTime / time.Duration(eso.stats.TotalIndexOps)
	}

	if eso.config.EnableMetrics {
		esIndexLatency.WithLabelValues("default", "index").Observe(duration.Seconds())
	}
}

// recordSearchLatency 记录搜索延迟
func (eso *ElasticsearchOptimizer) recordSearchLatency(duration time.Duration, queryType string) {
	eso.stats.TotalSearchOps++
	eso.stats.TotalSearchTime += duration
	if eso.stats.TotalSearchOps > 0 {
		eso.stats.AvgSearchLatency = eso.stats.TotalSearchTime / time.Duration(eso.stats.TotalSearchOps)
	}

	if eso.config.EnableMetrics {
		esSearchLatency.WithLabelValues("default", queryType).Observe(duration.Seconds())
		esSearchCount.WithLabelValues("default", queryType, "success").Inc()
	}
}

// recordBulkError 记录批量操作错误
func (eso *ElasticsearchOptimizer) recordBulkError(operation string, err error) {
	eso.stats.BulkErrors++
	log.Printf("Bulk operation error [%s]: %v", operation, err)
}

// recordSearchError 记录搜索错误
func (eso *ElasticsearchOptimizer) recordSearchError(operation string, err error) {
	eso.stats.SearchErrors++
	log.Printf("Search operation error [%s]: %v", operation, err)

	if eso.config.EnableMetrics {
		esSearchCount.WithLabelValues("default", operation, "error").Inc()
	}
}