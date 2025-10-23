package search

import (
	"context"
	"fmt"
	"time"

	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// SearchManager 搜索管理器
type SearchManager struct {
	searchService  *SearchService
	indexManager   *IndexManager
	esClient       *ElasticsearchClient
	queryBuilder    *QueryBuilder
	logger          *logrus.Logger
	isHealthy       bool
	lastHealthCheck time.Time
}

// SearchConfig 搜索配置
type SearchConfig struct {
	Elasticsearch ElasticsearchConfig   `yaml:"elasticsearch" json:"elasticsearch"`
	Redis         RedisCacheConfig     `yaml:"redis" json:"redis"`
	Indexing      IndexingConfig       `yaml:"indexing" json:"indexing"`
	Caching       CachingConfig        `yaml:"caching" json:"caching"`
}

// ElasticsearchConfig Elasticsearch配置
type ElasticsearchConfig struct {
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

// IndexingConfig 索引配置
type IndexingConfig struct {
	BatchSize      int           `yaml:"batch_size" json:"batch_size"`
	Workers        int           `yaml:"workers" json:"workers"`
	SyncInterval   time.Duration `yaml:"sync_interval" json:"sync_interval"`
	RetryAttempts  int           `yaml:"retry_attempts" json:"retry_attempts"`
	RetryDelay     time.Duration `yaml:"retry_delay" json:"retry_delay"`
}

// CachingConfig 缓存配置
type CachingConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled"`
	DefaultTTL time.Duration `yaml:"default_ttl" json:"default_ttl"`
	MaxTTL     time.Duration `yaml:"max_ttl" json:"max_ttl"`
	KeyPrefix  string        `yaml:"key_prefix" json:"key_prefix"`
}

// NewSearchManager 创建搜索管理器
func NewSearchManager(
	config *SearchConfig,
	docRepo repositories.DocumentRepository,
	userRepo repositories.UserRepository,
	logger *logrus.Logger,
) (*SearchManager, error) {
	if err := validateSearchConfig(config); err != nil {
		return nil, fmt.Errorf("invalid search config: %w", err)
	}

	// 创建Elasticsearch客户端
	esConfig := &Config{
		Addresses:     config.Elasticsearch.Addresses,
		Username:      config.Elasticsearch.Username,
		Password:      config.Elasticsearch.Password,
		APIKey:        config.Elasticsearch.APIKey,
		CloudID:       config.Elasticsearch.CloudID,
		IndexName:     config.Elasticsearch.IndexName,
		IndexPattern:  config.Elasticsearch.IndexPattern,
		SkipTLSVerify: config.Elasticsearch.SkipTLSVerify,
		RetryCount:    config.Elasticsearch.RetryCount,
		RetryBackoff:  config.Elasticsearch.RetryBackoff,
		Timeout:       config.Elasticsearch.Timeout,
	}

	esClient, err := NewElasticsearchClient(esConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 创建索引管理器
	indexManager := NewIndexManager(esClient, docRepo, userRepo, logger)

	// 创建查询构建器
	queryBuilder := NewQueryBuilder(indexManager)

	// 创建缓存
	var cache SearchCache
	if config.Caching.Enabled {
		redisConfig := &RedisCacheConfig{
			Host:       config.Redis.Host,
			Port:       config.Redis.Port,
			Password:   config.Redis.Password,
			Database:   config.Redis.Database,
			PoolSize:   config.Redis.PoolSize,
			KeyPrefix:  config.Caching.KeyPrefix,
			DefaultTTL: int(config.Caching.DefaultTTL.Seconds()),
		}

		cache, err = NewRedisCache(redisConfig, logger)
		if err != nil {
			logger.WithError(err).Warn("Failed to create redis cache, running without cache")
		}
	}

	// 创建搜索服务
	searchService := NewSearchService(indexManager, queryBuilder, docRepo, userRepo, logger, cache)

	manager := &SearchManager{
		searchService:   searchService,
		indexManager:    indexManager,
		esClient:        esClient,
		queryBuilder:     queryBuilder,
		logger:          logger,
		isHealthy:       true,
		lastHealthCheck: time.Now(),
	}

	// 启动健康检查
	go manager.startHealthCheck()

	// 启动定期索引同步
	if config.Indexing.SyncInterval > 0 {
		go manager.startPeriodicSync(config.Indexing.SyncInterval)
	}

	logger.Info("Search manager initialized successfully")

	return manager, nil
}

// validateSearchConfig 验证搜索配置
func validateSearchConfig(config *SearchConfig) error {
	if config == nil {
		return fmt.Errorf("search config is required")
	}

	if len(config.Elasticsearch.Addresses) == 0 {
		return fmt.Errorf("elasticsearch addresses are required")
	}

	// 设置默认值
	if config.Elasticsearch.IndexName == "" {
		config.Elasticsearch.IndexName = "documents"
	}
	if config.Elasticsearch.Timeout == 0 {
		config.Elasticsearch.Timeout = 30
	}
	if config.Elasticsearch.RetryCount == 0 {
		config.Elasticsearch.RetryCount = 3
	}
	if config.Elasticsearch.RetryBackoff == 0 {
		config.Elasticsearch.RetryBackoff = 100
	}

	if config.Indexing.BatchSize == 0 {
		config.Indexing.BatchSize = 100
	}
	if config.Indexing.Workers == 0 {
		config.Indexing.Workers = 4
	}
	if config.Indexing.RetryAttempts == 0 {
		config.Indexing.RetryAttempts = 3
	}
	if config.Indexing.RetryDelay == 0 {
		config.Indexing.RetryDelay = time.Second
	}

	if config.Caching.DefaultTTL == 0 {
		config.Caching.DefaultTTL = 5 * time.Minute
	}
	if config.Caching.MaxTTL == 0 {
		config.Caching.MaxTTL = 30 * time.Minute
	}

	return nil
}

// Search 执行搜索
func (sm *SearchManager) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	if !sm.isHealthy {
		return nil, fmt.Errorf("search service is unhealthy")
	}

	return sm.searchService.Search(ctx, req)
}

// Suggest 获取搜索建议
func (sm *SearchManager) Suggest(ctx context.Context, query, field string, size int) ([]string, error) {
	if !sm.isHealthy {
		return nil, fmt.Errorf("search service is unhealthy")
	}

	return sm.searchService.Suggest(ctx, query, field, size)
}

// GetSimilarDocuments 获取相似文档
func (sm *SearchManager) GetSimilarDocuments(ctx context.Context, documentID int64, fields []string, size int) ([]*SearchDocument, error) {
	if !sm.isHealthy {
		return nil, fmt.Errorf("search service is unhealthy")
	}

	return sm.searchService.GetSimilarDocuments(ctx, documentID, fields, size)
}

// Aggregate 执行聚合查询
func (sm *SearchManager) Aggregate(ctx context.Context, req *SearchRequest, aggregations map[string]map[string]interface{}) (*SearchResult, error) {
	if !sm.isHealthy {
		return nil, fmt.Errorf("search service is unhealthy")
	}

	return sm.searchService.Aggregate(ctx, req, aggregations)
}

// IndexDocument 索引文档
func (sm *SearchManager) IndexDocument(ctx context.Context, document *models.Document) error {
	if !sm.isHealthy {
		return fmt.Errorf("search service is unhealthy")
	}

	return sm.indexManager.IndexDocument(ctx, document)
}

// UpdateDocument 更新文档索引
func (sm *SearchManager) UpdateDocument(ctx context.Context, document *models.Document) error {
	if !sm.isHealthy {
		return fmt.Errorf("search service is unhealthy")
	}

	return sm.indexManager.UpdateDocument(ctx, document)
}

// DeleteDocument 删除文档索引
func (sm *SearchManager) DeleteDocument(ctx context.Context, documentID uint) error {
	if !sm.isHealthy {
		return fmt.Errorf("search service is unhealthy")
	}

	return sm.indexManager.DeleteDocument(ctx, documentID)
}

// BulkIndexDocuments 批量索引文档
func (sm *SearchManager) BulkIndexDocuments(ctx context.Context, documents []*models.Document) error {
	if !sm.isHealthy {
		return fmt.Errorf("search service is unhealthy")
	}

	return sm.indexManager.BulkIndexDocuments(ctx, documents)
}

// Reindex 重新索引
func (sm *SearchManager) Reindex(ctx context.Context, tenantID string) error {
	if !sm.isHealthy {
		return fmt.Errorf("search service is unhealthy")
	}

	return sm.searchService.Reindex(ctx, tenantID)
}

// GetSearchStats 获取搜索统计
func (sm *SearchManager) GetSearchStats(ctx context.Context, tenantID string) (*SearchStats, error) {
	return sm.searchService.GetSearchStats(ctx, tenantID)
}

// GetIndexStats 获取索引统计
func (sm *SearchManager) GetIndexStats(ctx context.Context) (*IndexStats, error) {
	return sm.indexManager.GetIndexStats(ctx)
}

// HealthCheck 健康检查
func (sm *SearchManager) HealthCheck(ctx context.Context) error {
	if err := sm.esClient.Ping(ctx); err != nil {
		sm.isHealthy = false
		return fmt.Errorf("elasticsearch health check failed: %w", err)
	}

	sm.isHealthy = true
	sm.lastHealthCheck = time.Now()
	return nil
}

// IsHealthy 检查服务是否健康
func (sm *SearchManager) IsHealthy() bool {
	return sm.isHealthy
}

// GetSearchService 获取搜索服务
func (sm *SearchManager) GetSearchService() *SearchService {
	return sm.searchService
}

// GetIndexManager 获取索引管理器
func (sm *SearchManager) GetIndexManager() *IndexManager {
	return sm.indexManager
}

// GetElasticsearchClient 获取Elasticsearch客户端
func (sm *SearchManager) GetElasticsearchClient() *ElasticsearchClient {
	return sm.esClient
}

// startHealthCheck 启动健康检查
func (sm *SearchManager) startHealthCheck() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := sm.HealthCheck(ctx); err != nil {
				sm.logger.WithError(err).Error("Health check failed")
			}
			cancel()
		}
	}
}

// startPeriodicSync 启动定期索引同步
func (sm *SearchManager) startPeriodicSync(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.logger.Info("Starting periodic index sync")
			// 这里可以添加定期同步逻辑
			// 例如：同步所有租户的文档变更
		}
	}
}

// Close 关闭搜索管理器
func (sm *SearchManager) Close() error {
	sm.logger.Info("Closing search manager")

	// 关闭Elasticsearch客户端
	if sm.esClient != nil {
		// Elasticsearch客户端通常不需要显式关闭
	}

	// 关闭缓存
	if sm.searchService != nil && sm.searchService.cache != nil {
		if redisCache, ok := sm.searchService.cache.(*RedisCache); ok {
			if err := redisCache.Close(); err != nil {
				sm.logger.WithError(err).Error("Failed to close redis cache")
			}
		}
	}

	return nil
}

// WarmUpCache 预热缓存
func (sm *SearchManager) WarmUpCache(ctx context.Context, queries []string) error {
	if sm.searchService.cache == nil {
		return fmt.Errorf("cache is not enabled")
	}

	redisCache, ok := sm.searchService.cache.(*RedisCache)
	if !ok {
		return fmt.Errorf("cache is not redis cache")
	}

	// 预热常见查询结果
	cacheData := make(map[string]interface{})
	for _, query := range queries {
		req := &SearchRequest{
			Query:    query,
			TenantID: "system", // 系统级预热
			Page:     1,
			PageSize: 10,
		}

		result, err := sm.searchService.Search(ctx, req)
		if err != nil {
			sm.logger.WithError(err).WithField("query", query).Warn("Failed to warm up cache for query")
			continue
		}

		cacheKey := sm.searchService.generateCacheKey(req)
		cacheData[cacheKey] = result
	}

	// 设置缓存
	return redisCache.WarmUp(cacheData, 5*time.Minute)
}