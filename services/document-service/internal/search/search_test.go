package search

import (
	"context"
	"testing"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestQueryBuilder 测试查询构建器
func TestQueryBuilder(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建模拟仓库
	docRepo := &mocks.DocumentRepository{}
	userRepo := &mocks.UserRepository{}

	// 创建Elasticsearch客户端模拟
	esClient := &ElasticsearchClient{
		logger:    logger,
		indexName: "test_documents",
	}

	// 创建索引管理器
	indexManager := &IndexManager{
		esClient: esClient,
		docRepo:  docRepo,
		userRepo: userRepo,
		logger:   logger,
	}

	// 创建查询构建器
	queryBuilder := NewQueryBuilder(indexManager)

	t.Run("BuildSearchQuery_Basic", func(t *testing.T) {
		req := &SearchRequest{
			Query:    "test query",
			TenantID: "tenant1",
			Page:     1,
			PageSize: 10,
		}

		query, err := queryBuilder.BuildSearchQuery(req)

		assert.NoError(t, err)
		assert.NotNil(t, query)

		// 验证查询结构
		boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
		mustQueries := boolQuery["must"].([]interface{})
		filterQueries := boolQuery["filter"].([]interface{})

		// 应该有搜索查询
		assert.Greater(t, len(mustQueries), 0)

		// 应该有租户过滤
		assert.Greater(t, len(filterQueries), 0)

		// 应该有分页
		assert.Equal(t, 0, query["from"])
		assert.Equal(t, 10, query["size"])
	})

	t.Run("BuildSearchQuery_WithFilters", func(t *testing.T) {
		req := &SearchRequest{
			Query:    "test query",
			TenantID: "tenant1",
			Filters: map[string]interface{}{
				"category": "document",
				"tags":     []string{"important", "work"},
			},
			Page:     1,
			PageSize: 10,
		}

		query, err := queryBuilder.BuildSearchQuery(req)

		assert.NoError(t, err)
		assert.NotNil(t, query)

		boolQuery := query["query"].(map[string]interface{})["bool"].(map[string]interface{})
		filterQueries := boolQuery["filter"].([]interface{})

		// 应该有多个过滤器
		assert.Greater(t, len(filterQueries), 1)
	})

	t.Run("BuildSearchQuery_WithSort", func(t *testing.T) {
		req := &SearchRequest{
			Query:    "test query",
			TenantID: "tenant1",
			SortBy: []SortField{
				{Field: "created_at", Order: "desc"},
				{Field: "name", Order: "asc"},
			},
			Page:     1,
			PageSize: 10,
		}

		query, err := queryBuilder.BuildSearchQuery(req)

		assert.NoError(t, err)
		assert.NotNil(t, query)

		sorts := query["sort"].([]interface{})
		assert.Len(t, sorts, 2)
	})

	t.Run("BuildSearchQuery_WithHighlight", func(t *testing.T) {
		req := &SearchRequest{
			Query:     "test query",
			TenantID:  "tenant1",
			Highlight: true,
			Page:      1,
			PageSize:  10,
		}

		query, err := queryBuilder.BuildSearchQuery(req)

		assert.NoError(t, err)
		assert.NotNil(t, query)

		highlight := query["highlight"].(map[string]interface{})
		assert.NotNil(t, highlight)
		assert.Contains(t, highlight, "pre_tags")
		assert.Contains(t, highlight, "post_tags")
	})

	t.Run("BuildSearchQuery_WithSuggest", func(t *testing.T) {
		req := &SearchRequest{
			Query:      "test query",
			TenantID:   "tenant1",
			Suggest:    true,
			SuggestSize: 5,
			Page:       1,
			PageSize:   10,
		}

		query, err := queryBuilder.BuildSearchQuery(req)

		assert.NoError(t, err)
		assert.NotNil(t, query)

		suggest := query["suggest"].(map[string]interface{})
		assert.NotNil(t, suggest)
		assert.Contains(t, suggest, "simple_phrase")
	})

	t.Run("ValidateSearchRequest", func(t *testing.T) {
		// 测试无效请求
		req := &SearchRequest{
			Query:    "test",
			Page:     0, // 无效页码
			PageSize: 10,
		}

		queryBuilder := NewQueryBuilder(indexManager)
		err := queryBuilder.BuildSearchQuery(req)

		assert.NoError(t, err) // 应该自动修正
	})

	t.Run("BuildAggregationQuery", func(t *testing.T) {
		// 测试terms聚合
		agg := queryBuilder.BuildAggregationQuery("terms", "category", 10)
		assert.NotNil(t, agg)

		terms := agg["terms"].(map[string]interface{})
		assert.Equal(t, "category", terms["field"])
		assert.Equal(t, 10, terms["size"])

		// 测试date_histogram聚合
		dateAgg := queryBuilder.BuildAggregationQuery("date_histogram", "created_at", 12)
		assert.NotNil(t, dateAgg)

		histogram := dateAgg["date_histogram"].(map[string]interface{})
		assert.Equal(t, "created_at", histogram["field"])
		assert.Equal(t, "month", histogram["calendar_interval"])
	})

	t.Run("BuildAutoCompleteQuery", func(t *testing.T) {
		query, err := queryBuilder.BuildAutoCompleteQuery("test", "name", 5)

		assert.NoError(t, err)
		assert.NotNil(t, query)

		suggest := query["suggest"].(map[string]interface{})
		assert.Contains(t, suggest, "prefix")

		prefix := suggest["prefix"].(map[string]interface{})
		completion := prefix["prefix"].(map[string]interface{})["completion"].(map[string]interface{})
		assert.Equal(t, "name.suggest", completion["field"])
		assert.Equal(t, 5, completion["size"])
	})
}

// TestIndexManager 测试索引管理器
func TestIndexManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建模拟仓库
	docRepo := &mocks.DocumentRepository{}
	userRepo := &mocks.UserRepository{}

	// 创建Elasticsearch客户端模拟
	esClient := &ElasticsearchClient{
		logger:    logger,
		indexName: "test_documents",
	}

	// 创建索引管理器
	indexManager := NewIndexManager(esClient, docRepo, userRepo, logger)

	t.Run("BuildSearchDocument", func(t *testing.T) {
		document := &models.Document{
			ID:          1,
			UUID:        "test-uuid",
			Name:        "Test Document",
			Description: "Test Description",
			Category:    "test",
			TenantID:    "tenant1",
			CreatedBy:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			MIMEType:    "application/pdf",
			Size:        1024,
			Version:     1,
		}

		searchDoc := indexManager.buildSearchDocument(document)

		assert.NotNil(t, searchDoc)
		assert.Equal(t, uint(1), searchDoc["id"])
		assert.Equal(t, "test-uuid", searchDoc["uuid"])
		assert.Equal(t, "Test Document", searchDoc["name"])
		assert.Equal(t, "tenant1", searchDoc["tenant_id"])
		assert.Equal(t, uint(1), searchDoc["created_by"])
		assert.Equal(t, "application/pdf", searchDoc["mime_type"])
		assert.Equal(t, int64(1024), searchDoc["size"])
		assert.Equal(t, 1, searchDoc["version"])
	})

	t.Run("BuildSearchDocument_WithTags", func(t *testing.T) {
		tags := []string{"important", "work"}
		document := &models.Document{
			ID:       1,
			Name:     "Test Document",
			TenantID: "tenant1",
			Tags:     tags,
		}

		searchDoc := indexManager.buildSearchDocument(document)

		assert.NotNil(t, searchDoc)
		assert.Equal(t, tags, searchDoc["tags"])
	})

	t.Run("BuildDocumentMapping", func(t *testing.T) {
		mapping := indexManager.buildDocumentMapping()

		assert.NotNil(t, mapping)
		properties := mapping["properties"].(map[string]interface{})

		// 验证关键字段
		assert.Contains(t, properties, "id")
		assert.Contains(t, properties, "uuid")
		assert.Contains(t, properties, "name")
		assert.Contains(t, properties, "description")
		assert.Contains(t, properties, "tenant_id")
		assert.Contains(t, properties, "created_at")

		// 验证name字段的多字段
		nameField := properties["name"].(map[string]interface{})
		assert.Equal(t, "text", nameField["type"])
		assert.Contains(t, nameField["fields"].(map[string]interface{}), "keyword")
		assert.Contains(t, nameField["fields"].(map[string]interface{}), "suggest")
	})
}

// TestSearchService 测试搜索服务
func TestSearchService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建模拟仓库
	docRepo := &mocks.DocumentRepository{}
	userRepo := &mocks.UserRepository{}

	// 创建Elasticsearch客户端模拟
	esClient := &ElasticsearchClient{
		logger:    logger,
		indexName: "test_documents",
	}

	// 创建索引管理器和查询构建器
	indexManager := NewIndexManager(esClient, docRepo, userRepo, logger)
	queryBuilder := NewQueryBuilder(indexManager)

	// 创建搜索服务（不使用缓存）
	searchService := NewSearchService(indexManager, queryBuilder, docRepo, userRepo, logger, nil)

	t.Run("GenerateCacheKey", func(t *testing.T) {
		req := &SearchRequest{
			Query:    "test query",
			TenantID: "tenant1",
			Page:     1,
			PageSize: 10,
			Filters: map[string]interface{}{
				"category": "document",
			},
		}

		cacheKey := searchService.generateCacheKey(req)

		assert.NotEmpty(t, cacheKey)
		assert.Contains(t, cacheKey, "search:tenant1:test query:1:10")
		assert.Contains(t, cacheKey, "category")
	})

	t.Run("ParseSearchDocument", func(t *testing.T) {
		hit := map[string]interface{}{
			"_id":    "1",
			"_score": 1.0,
			"_source": map[string]interface{}{
				"id":        1.0,
				"uuid":      "test-uuid",
				"name":      "Test Document",
				"tenant_id": "tenant1",
				"created_at": "2023-01-01T00:00:00Z",
			},
			"highlight": map[string]interface{}{
				"name": []interface{}{"<mark>Test</mark> Document"},
			},
		}

		doc := searchService.parseSearchDocument(hit)

		assert.NotNil(t, doc)
		assert.Equal(t, uint(1), doc.ID)
		assert.Equal(t, "test-uuid", doc.UUID)
		assert.Equal(t, "Test Document", doc.Name)
		assert.Equal(t, "tenant1", doc.TenantID)
		assert.Equal(t, 1.0, doc.Score)
		assert.NotNil(t, doc.Highlights)
		assert.Equal(t, "<mark>Test</mark> Document", doc.Highlights["name"])
	})
}

// TestRedisCache 测试Redis缓存
func TestRedisCache(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("RedisCacheConfig_Validation", func(t *testing.T) {
		// 测试空配置
		_, err := NewRedisCache(nil, logger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis cache config is required")

		// 测试默认值设置
		config := &RedisCacheConfig{
			Host: "localhost",
			Port: 6379,
		}

		// 这里不实际连接Redis，只测试配置验证
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 6379, config.Port)
	})

	t.Run("ParseMemoryBytes", func(t *testing.T) {
		// 测试内存字节数解析
		mem, err := parseMemoryBytes("1048576")
		assert.NoError(t, err)
		assert.Equal(t, int64(1048576), mem)

		// 测试无效值
		_, err = parseMemoryBytes("invalid")
		assert.Error(t, err)
	})

	t.Run("ParseRedisInfo", func(t *testing.T) {
		info := `# Server
redis_version:7.0.0
# Memory
used_memory:1048576
used_memory_human:1M
# Stats
total_commands_processed:1000
keyspace_hits:800
keyspace_misses:200`

		lines := parseRedisInfo(info)
		assert.Len(t, lines, 5)

		for _, line := range lines {
			key, value, ok := parseRedisInfoLine(line)
			assert.True(t, ok, "Failed to parse line: "+line)
			assert.NotEmpty(t, key)
			assert.NotEmpty(t, value)
		}
	})
}

// BenchmarkQueryBuilder 性能测试
func BenchmarkQueryBuilder_BuildSearchQuery(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建模拟仓库
	docRepo := &mocks.DocumentRepository{}
	userRepo := &mocks.UserRepository{}

	// 创建Elasticsearch客户端模拟
	esClient := &ElasticsearchClient{
		logger:    logger,
		indexName: "test_documents",
	}

	// 创建索引管理器和查询构建器
	indexManager := NewIndexManager(esClient, docRepo, userRepo, logger)
	queryBuilder := NewQueryBuilder(indexManager)

	// 创建复杂的搜索请求
	req := &SearchRequest{
		Query: "test query with multiple words",
		TenantID: "tenant1",
		Filters: map[string]interface{}{
			"category": "document",
			"tags":     []string{"important", "work", "project"},
			"date_range": map[string]interface{}{
				"start": "2023-01-01",
				"end":   "2023-12-31",
			},
		},
		SortBy: []SortField{
			{Field: "created_at", Order: "desc"},
			{Field: "name", Order: "asc"},
		},
		Highlight:    true,
		Suggest:      true,
		SuggestSize:  5,
		Page:         2,
		PageSize:     20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := queryBuilder.BuildSearchQuery(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestSearchManager 测试搜索管理器
func TestSearchManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("SearchConfig_Validation", func(t *testing.T) {
		// 测试空配置
		_, err := NewSearchManager(nil, nil, nil, logger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "search config is required")

		// 测试无效配置（没有ES地址）
		config := &SearchConfig{}
		_, err = NewSearchManager(config, nil, nil, logger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "elasticsearch addresses are required")

		// 测试默认值设置
		config = &SearchConfig{
			Elasticsearch: ElasticsearchConfig{
				Addresses: []string{"http://localhost:9200"},
			},
		}

		// 设置默认值
		validateSearchConfig(config)
		assert.Equal(t, "documents", config.Elasticsearch.IndexName)
		assert.Equal(t, 30, config.Elasticsearch.Timeout)
		assert.Equal(t, 3, config.Elasticsearch.RetryCount)
		assert.Equal(t, 100, config.Indexing.BatchSize)
		assert.Equal(t, 5*time.Minute, config.Caching.DefaultTTL)
	})
}