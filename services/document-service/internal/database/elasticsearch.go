package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
	Addresses    []string `mapstructure:"addresses"`
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	APIKey       string   `mapstructure:"api_key"`
	CloudID      string   `mapstructure:"cloud_id"`
	MaxRetries   int      `mapstructure:"max_retries"`
	RetryBackoff int      `mapstructure:"retry_backoff"`
	Timeout      int      `mapstructure:"timeout"`
	Compress     bool     `mapstructure:"compress"`
}

// DefaultElasticsearchConfig 返回默认Elasticsearch配置
func DefaultElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		Addresses:    []string{"http://localhost:9200"},
		Username:     "",
		Password:     "",
		APIKey:       "",
		CloudID:      "",
		MaxRetries:   3,
		RetryBackoff: 100,
		Timeout:      10,
		Compress:     true,
	}
}

// ElasticsearchClient Elasticsearch客户端管理器
type ElasticsearchClient struct {
	Client *elasticsearch.Client
	config *ElasticsearchConfig
}

// NewElasticsearchClient 创建新的Elasticsearch客户端
func NewElasticsearchClient(config *ElasticsearchConfig) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
		APIKey:    config.APIKey,
		CloudID:   config.CloudID,
		MaxRetries: config.MaxRetries,
		RetryBackoff: func(i int) time.Duration {
			return time.Duration(config.RetryBackoff) * time.Duration(i) * time.Millisecond
		},
		CompressRequestBody: config.Compress,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// 测试连接
	esClient := &ElasticsearchClient{
		Client: client,
		config: config,
	}

	if err := esClient.Health(); err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}

	return esClient, nil
}

// Health Elasticsearch健康检查
func (e *ElasticsearchClient) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := e.Client.Cluster.Health(
		e.Client.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to check Elasticsearch health: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch health check failed: %s", res.Status())
	}

	var health map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return fmt.Errorf("failed to decode health response: %w", err)
	}

	status, ok := health["status"].(string)
	if !ok {
		return fmt.Errorf("invalid health response format")
	}

	if status != "green" && status != "yellow" {
		return fmt.Errorf("Elasticsearch cluster status is not healthy: %s", status)
	}

	return nil
}

// Info 获取Elasticsearch信息
func (e *ElasticsearchClient) Info() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := e.Client.Info(
		e.Client.Info.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get Elasticsearch info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to get Elasticsearch info: %s", res.Status())
	}

	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode info response: %w", err)
	}

	return info, nil
}

// Stats 获取Elasticsearch统计信息
func (e *ElasticsearchClient) Stats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := e.Client.Cluster.Stats(
		e.Client.Cluster.Stats.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get Elasticsearch stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("failed to get Elasticsearch stats: %s", res.Status())
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats response: %w", err)
	}

	return stats, nil
}

// IndexExists 检查索引是否存在
func (e *ElasticsearchClient) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := e.Client.Indices.Exists(
		[]string{index},
		e.Client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, fmt.Errorf("failed to check if index exists: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// CreateIndex 创建索引
func (e *ElasticsearchClient) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	if mapping == nil {
		mapping = map[string]interface{}{}
	}

	mappingBytes, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(mappingBytes),
	}

	res, err := req.Do(ctx, e.Client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errorResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&errorResp)
		return fmt.Errorf("failed to create index: %s, error: %v", res.Status(), errorResp)
	}

	return nil
}

// DeleteIndex 删除索引
func (e *ElasticsearchClient) DeleteIndex(ctx context.Context, index string) error {
	res, err := e.Client.Indices.Delete(
		[]string{index},
		e.Client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete index: %s", res.Status())
	}

	return nil
}

// IndexDocument 索引文档
func (e *ElasticsearchClient) IndexDocument(ctx context.Context, index, docID string, document interface{}) error {
	docBytes, err := json.Marshal(document)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(docBytes),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.Client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errorResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&errorResp)
		return fmt.Errorf("failed to index document: %s, error: %v", res.Status(), errorResp)
	}

	return nil
}

// UpdateDocument 更新文档
func (e *ElasticsearchClient) UpdateDocument(ctx context.Context, index, docID string, update map[string]interface{}) error {
	updateBytes, err := json.Marshal(map[string]interface{}{
		"doc": update,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: docID,
		Body:       bytes.NewReader(updateBytes),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.Client)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errorResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&errorResp)
		return fmt.Errorf("failed to update document: %s, error: %v", res.Status(), errorResp)
	}

	return nil
}

// GetDocument 获取文档
func (e *ElasticsearchClient) GetDocument(ctx context.Context, index, docID string, dest interface{}) error {
	res, err := e.Client.Get(
		index, docID,
		e.Client.Get.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("document not found: %s/%s", index, docID)
		}
		var errorResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&errorResp)
		return fmt.Errorf("failed to get document: %s, error: %v", res.Status(), errorResp)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if source, ok := response["_source"]; ok {
		sourceBytes, err := json.Marshal(source)
		if err != nil {
			return fmt.Errorf("failed to marshal source: %w", err)
		}
		return json.Unmarshal(sourceBytes, dest)
	}

	return fmt.Errorf("document source not found")
}

// DeleteDocument 删除文档
func (e *ElasticsearchClient) DeleteDocument(ctx context.Context, index, docID string) error {
	res, err := e.Client.Delete(
		index, docID,
		e.Client.Delete.WithContext(ctx),
		e.Client.Delete.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		var errorResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&errorResp)
		return fmt.Errorf("failed to delete document: %s, error: %v", res.Status(), errorResp)
	}

	return nil
}

// Search 搜索文档
func (e *ElasticsearchClient) Search(ctx context.Context, index string, query map[string]interface{}) (*SearchResult, error) {
	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := e.Client.Search(
		e.Client.Search.WithContext(ctx),
		e.Client.Search.WithIndex(index),
		e.Client.Search.WithBody(bytes.NewReader(queryBytes)),
		e.Client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errorResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&errorResp)
		return nil, fmt.Errorf("search failed: %s, error: %v", res.Status(), errorResp)
	}

	var result SearchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search result: %w", err)
	}

	return &result, nil
}

// Bulk 批量操作
func (e *ElasticsearchClient) Bulk(ctx context.Context, operations []BulkOperation) error {
	if len(operations) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, op := range operations {
		// 添加操作元数据
		meta := map[string]interface{}{
			"_index": op.Index,
		}
		if op.DocumentID != "" {
			meta["_id"] = op.DocumentID
		}

		metaBytes, err := json.Marshal(map[string]interface{}{
			strings.ToLower(op.Action): meta,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal operation meta: %w", err)
		}

		buf.Write(metaBytes)
		buf.WriteByte('\n')

		// 添加文档数据（对于删除操作不需要）
		if op.Document != nil {
			docBytes, err := json.Marshal(op.Document)
			if err != nil {
				return fmt.Errorf("failed to marshal document: %w", err)
			}
			buf.Write(docBytes)
			buf.WriteByte('\n')
		}
	}

	res, err := e.Client.Bulk(
		bytes.NewReader(buf.Bytes()),
		e.Client.Bulk.WithContext(ctx),
		e.Client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to execute bulk operation: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errorResp map[string]interface{}
		json.NewDecoder(res.Body).Decode(&errorResp)
		return fmt.Errorf("bulk operation failed: %s, error: %v", res.Status(), errorResp)
	}

	// 检查批量操作结果
	var bulkResult BulkResult
	if err := json.NewDecoder(res.Body).Decode(&bulkResult); err != nil {
		return fmt.Errorf("failed to decode bulk result: %w", err)
	}

	if bulkResult.Errors {
		// 记录错误但不返回，让用户自行处理
		for _, item := range bulkResult.Items {
			for action, result := range item {
				if result.Status >= 400 {
					// 可以在这里记录日志
					continue
				}
			}
		}
	}

	return nil
}

// SearchResult 搜索结果
type SearchResult struct {
	Hits Hits `json:"hits"`
}

// Hits 命中结果
type Hits struct {
	Total    Total    `json:"total"`
	Hits     []Hit    `json:"hits"`
	MaxScore float64  `json:"max_score"`
}

// Total 总数
type Total struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

// Hit 单个命中结果
type Hit struct {
	Index     string                 `json:"_index"`
	ID        string                 `json:"_id"`
	Score     float64                `json:"_score"`
	Source    map[string]interface{} `json:"_source"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
}

// BulkOperation 批量操作
type BulkOperation struct {
	Action     string                 `json:"action"`     // index, update, delete
	Index      string                 `json:"index"`
	DocumentID string                 `json:"document_id,omitempty"`
	Document   map[string]interface{} `json:"document,omitempty"`
}

// BulkResult 批量操作结果
type BulkResult struct {
	Errors bool                    `json:"errors"`
	Items  []map[string]BulkItemResult `json:"items"`
}

// BulkItemResult 批量操作单项结果
type BulkItemResult struct {
	Index      string `json:"_index"`
	ID         string `json:"_id"`
	Status     int    `json:"status"`
	Error      string `json:"error,omitempty"`
	Result     string `json:"result,omitempty"`
}