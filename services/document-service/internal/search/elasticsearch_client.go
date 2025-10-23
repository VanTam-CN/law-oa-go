package search

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
)

// ElasticsearchClient Elasticsearch客户端封装
type ElasticsearchClient struct {
	client       *elasticsearch.Client
	logger       *logrus.Logger
	indexName    string
	indexPattern string
	settings     *ClusterSettings
}

// ClusterSettings 集群设置
type ClusterSettings struct {
	NumberOfShards   int                    `json:"number_of_shards"`
	NumberOfReplicas int                    `json:"number_of_replicas"`
	RefreshInterval  string                 `json:"refresh_interval"`
	MaxResultWindow  int                    `json:"max_result_window"`
	Analysis         map[string]interface{} `json:"analysis,omitempty"`
}

// IndexTemplate 索引模板
type IndexTemplate struct {
	IndexPatterns []string                 `json:"index_patterns"`
	Template      map[string]interface{}   `json:"template"`
	ComposedOf    []string                 `json:"composed_of"`
	Priority      int                      `json:"priority"`
	Version       int                      `json:"version"`
	Settings      map[string]interface{}   `json:"settings"`
	Mappings      map[string]interface{}   `json:"mappings"`
}

// Config Elasticsearch配置
type Config struct {
	Addresses     []string `yaml:"addresses" json:"addresses"`
	Username      string   `yaml:"username" json:"username"`
	Password      string   `yaml:"password" json:"password"`
	APIKey        string   `yaml:"api_key" json:"api_key"`
	CloudID       string   `yaml:"cloud_id" json:"cloud_id"`
	IndexName     string   `yaml:"index_name" json:"index_name"`
	IndexPattern  string   `yaml:"index_pattern" json:"index_pattern"`
	SkipTLSVerify bool     `yaml:"skip_tls_verify" json:"skip_tls_verify"`
	RetryCount    int      `yaml:"retry_count" json:"retry_count"`
	RetryBackoff  int      `yaml:"retry_backoff" json:"retry_backoff"`
	Timeout       int      `yaml:"timeout" json:"timeout"`
}

// NewElasticsearchClient 创建Elasticsearch客户端
func NewElasticsearchClient(config *Config, logger *logrus.Logger) (*ElasticsearchClient, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid elasticsearch config: %w", err)
	}

	// 构建客户端配置
	cfg := elasticsearch.Config{
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
		APIKey:    config.APIKey,
		CloudID:   config.CloudID,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.SkipTLSVerify,
			},
		},
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			return time.Duration(config.RetryBackoff) * time.Duration(i) * time.Millisecond
		},
		MaxRetries: config.RetryCount,
		DiscoverNodesOnStart: true,
		DiscoverNodesInterval: 30 * time.Second,
	}

	// 创建客户端
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 测试连接
	if err := testConnection(client, config.Timeout); err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}

	// 设置默认值
	indexName := config.IndexName
	if indexName == "" {
		indexName = "documents"
	}

	indexPattern := config.IndexPattern
	if indexPattern == "" {
		indexPattern = "documents-*"
	}

	esClient := &ElasticsearchClient{
		client:       client,
		logger:       logger,
		indexName:    indexName,
		indexPattern: indexPattern,
		settings: &ClusterSettings{
			NumberOfShards:   1,
			NumberOfReplicas: 1,
			RefreshInterval:  "1s",
			MaxResultWindow:  50000,
		},
	}

	// 初始化索引
	if err := esClient.initializeIndex(); err != nil {
		logger.WithError(err).Error("Failed to initialize elasticsearch index")
		return nil, fmt.Errorf("failed to initialize index: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"addresses":    config.Addresses,
		"index_name":   indexName,
		"index_pattern": indexPattern,
	}).Info("Elasticsearch client initialized successfully")

	return esClient, nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if len(config.Addresses) == 0 {
		return fmt.Errorf("at least one elasticsearch address is required")
	}

	if config.RetryCount <= 0 {
		config.RetryCount = 3
	}

	if config.RetryBackoff <= 0 {
		config.RetryBackoff = 100
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	return nil
}

// testConnection 测试连接
func testConnection(client *elasticsearch.Client, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	// 执行简单的集群健康检查
	res, err := client.Info(
		client.Info.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("cluster info request failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("cluster info returned error: %s", res.Status())
	}

	return nil
}

// initializeIndex 初始化索引
func (es *ElasticsearchClient) initializeIndex() error {
	// 检查索引是否存在
	exists, err := es.IndexExists()
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if !exists {
		// 创建索引
		if err := es.CreateIndex(); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
		es.logger.Info("Created new elasticsearch index")
	} else {
		// 检查索引映射是否需要更新
		if err := es.updateIndexMapping(); err != nil {
			es.logger.WithError(err).Warn("Failed to update index mapping")
		}
	}

	// 创建索引模板
	if err := es.createIndexTemplate(); err != nil {
		es.logger.WithError(err).Warn("Failed to create index template")
	}

	return nil
}

// createIndex 创建索引
func (es *ElasticsearchClient) createIndex() error {
	// 构建索引设置
	settings := map[string]interface{}{
		"number_of_shards":   es.settings.NumberOfShards,
		"number_of_replicas": es.settings.NumberOfReplicas,
		"refresh_interval":   es.settings.RefreshInterval,
		"max_result_window":  es.settings.MaxResultWindow,
		"analysis": map[string]interface{}{
			"analyzer": map[string]interface{}{
				"document_analyzer": map[string]interface{}{
					"type":      "custom",
					"tokenizer": "standard",
					"filter": []string{
						"lowercase",
						"stop",
						"snowball",
					},
				},
				"search_analyzer": map[string]interface{}{
					"type":      "custom",
					"tokenizer": "standard",
					"filter": []string{
						"lowercase",
						"stop",
					},
				},
			},
			"filter": map[string]interface{}{
				"document_filter": map[string]interface{}{
					"type":     "edge_ngram",
					"min_gram": 2,
					"max_gram": 20,
				},
			},
		},
	}

	// 构建索引映射
	mappings := es.buildDocumentMapping()

	// 构建完整的索引配置
	indexConfig := map[string]interface{}{
		"settings": settings,
		"mappings": mappings,
	}

	// 序列化配置
	configJSON, err := json.Marshal(indexConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal index config: %w", err)
	}

	// 创建索引请求
	req := esapi.IndicesCreateRequest{
		Index: es.indexName,
		Body:  bytes.NewReader(configJSON),
	}

	// 执行请求
	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index creation failed: %s", res.Status())
	}

	return nil
}

// buildDocumentMapping 构建文档映射
func (es *ElasticsearchClient) buildDocumentMapping() map[string]interface{} {
	return map[string]interface{}{
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
						"type":         "keyword",
						"ignore_above": 256,
					},
					"suggest": map[string]interface{}{
						"type":         "completion",
						"preserve_separators": true,
						"preserve_position_increments": true,
						"max_input_length": 50,
					},
					"edge_ngram": map[string]interface{}{
						"type":     "text",
						"analyzer": "document_analyzer",
						"search_analyzer": "search_analyzer",
					},
				},
				"analyzer": "document_analyzer",
				"search_analyzer": "search_analyzer",
			},
			"description": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 512,
					},
				},
				"analyzer": "document_analyzer",
				"search_analyzer": "search_analyzer",
			},
			"content": map[string]interface{}{
				"type": "text",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":         "keyword",
						"ignore_above": 1024,
					},
				},
				"analyzer": "document_analyzer",
				"search_analyzer": "search_analyzer",
			},
			"category": map[string]interface{}{
				"type": "keyword",
				"fields": map[string]interface{}{
					"text": map[string]interface{}{
						"type": "text",
						"analyzer": "keyword",
					},
				},
			},
			"tags": map[string]interface{}{
				"type": "keyword",
			},
			"tenant_id": map[string]interface{}{
				"type": "keyword",
			},
			"created_by": map[string]interface{}{
				"type": "integer",
			},
			"updated_by": map[string]interface{}{
				"type": "integer",
			},
			"created_at": map[string]interface{}{
				"type":   "date",
				"format": "strict_date_optional_time||epoch_millis",
			},
			"updated_at": map[string]interface{}{
				"type":   "date",
				"format": "strict_date_optional_time||epoch_millis",
			},
			"mime_type": map[string]interface{}{
				"type": "keyword",
			},
			"size": map[string]interface{}{
				"type": "long",
			},
			"file_hash": map[string]interface{}{
				"type": "keyword",
			},
			"version": map[string]interface{}{
				"type": "integer",
			},
			"status": map[string]interface{}{
				"type": "keyword",
			},
			"access_level": map[string]interface{}{
				"type": "keyword",
			},
			"priority": map[string]interface{}{
				"type": "integer",
			},
			"metadata": map[string]interface{}{
				"type": "object",
				"dynamic": true,
			},
			"permissions": map[string]interface{}{
				"type": "nested",
				"properties": map[string]interface{}{
					"user_id": map[string]interface{}{
						"type": "integer",
					},
					"role_id": map[string]interface{}{
						"type": "integer",
					},
					"permission": map[string]interface{}{
						"type": "keyword",
					},
					"granted_at": map[string]interface{}{
						"type": "date",
					},
				},
			},
		},
	}
}

// createIndexTemplate 创建索引模板
func (es *ElasticsearchClient) createIndexTemplate() error {
	template := IndexTemplate{
		IndexPatterns: []string{es.indexPattern},
		Template: map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   es.settings.NumberOfShards,
				"number_of_replicas": es.settings.NumberOfReplicas,
				"refresh_interval":   es.settings.RefreshInterval,
				"analysis": map[string]interface{}{
					"analyzer": map[string]interface{}{
						"document_analyzer": map[string]interface{}{
							"type":      "custom",
							"tokenizer": "standard",
							"filter": []string{
								"lowercase",
								"stop",
								"snowball",
							},
						},
						"search_analyzer": map[string]interface{}{
							"type":      "custom",
							"tokenizer": "standard",
							"filter": []string{
								"lowercase",
								"stop",
							},
						},
					},
				},
			},
			"mappings": es.buildDocumentMapping(),
		},
		Priority: 100,
		Version:  1,
	}

	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal index template: %w", err)
	}

	req := esapi.IndicesPutIndexTemplateRequest{
		Name: "documents-template",
		Body: bytes.NewReader(templateJSON),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to create index template: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("index template creation failed: %s", res.Status())
	}

	return nil
}

// IndexExists 检查索引是否存在
func (es *ElasticsearchClient) IndexExists() (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{es.indexName},
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// updateIndexMapping 更新索引映射
func (es *ElasticsearchClient) updateIndexMapping() error {
	mappings := es.buildDocumentMapping()
	mappingsJSON, err := json.Marshal(map[string]interface{}{
		"properties": mappings["properties"],
	})
	if err != nil {
		return fmt.Errorf("failed to marshal mappings: %w", err)
	}

	req := esapi.IndicesPutMappingRequest{
		Index: []string{es.indexName},
		Body:  bytes.NewReader(mappingsJSON),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to update index mapping: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("index mapping update failed: %s", res.Status())
	}

	return nil
}

// GetClient 获取原始Elasticsearch客户端
func (es *ElasticsearchClient) GetClient() *elasticsearch.Client {
	return es.client
}

// GetIndexName 获取索引名称
func (es *ElasticsearchClient) GetIndexName() string {
	return es.indexName
}

// GetIndexPattern 获取索引模式
func (es *ElasticsearchClient) GetIndexPattern() string {
	return es.indexPattern
}

// Ping 健康检查
func (es *ElasticsearchClient) Ping(ctx context.Context) error {
	res, err := es.client.Ping(
		es.client.Ping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping returned error: %s", res.Status())
	}

	return nil
}

// GetClusterInfo 获取集群信息
func (es *ElasticsearchClient) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	res, err := es.client.Info(
		es.client.Info.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}
	defer res.Body.Close()

	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode cluster info: %w", err)
	}

	clusterInfo := &ClusterInfo{
		Name: getString(info, "cluster_name"),
		UUID: getString(info, "cluster_uuid"),
	}

	if version, ok := info["version"].(map[string]interface{}); ok {
		clusterInfo.Version.Elasticsearch = getString(version, "number")
		clusterInfo.Version.Lucene = getString(version, "lucene_version")
	}

	if tagline, ok := info["tagline"].(string); ok {
		clusterInfo.Tagline = tagline
	}

	return clusterInfo, nil
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Version struct {
		Elasticsearch string `json:"elasticsearch"`
		Lucene        string `json:"lucene"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

// getString 安全获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}