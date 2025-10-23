package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultElasticsearchConfig(t *testing.T) {
	config := DefaultElasticsearchConfig()

	assert.Equal(t, []string{"http://localhost:9200"}, config.Addresses)
	assert.Equal(t, "", config.Username)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, "", config.APIKey)
	assert.Equal(t, "", config.CloudID)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100, config.RetryBackoff)
	assert.Equal(t, 10, config.Timeout)
	assert.True(t, config.Compress)
}

func TestElasticsearchClient_NewElasticsearchClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()

	t.Run("valid config", func(t *testing.T) {
		client, err := NewElasticsearchClient(config)
		if err != nil {
			t.Skipf("Skipping test due to connection error: %v", err)
		}

		assert.NotNil(t, client)
		assert.NotNil(t, client.Client)
	})

	t.Run("invalid config", func(t *testing.T) {
		invalidConfig := &ElasticsearchConfig{
			Addresses: []string{"http://invalid-host:9200"},
		}

		client, err := NewElasticsearchClient(invalidConfig)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestElasticsearchClient_Health(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	err = client.Health()
	assert.NoError(t, err)
}

func TestElasticsearchClient_Info(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	info, err := client.Info()
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Contains(t, info, "tagline")
	assert.Contains(t, info, "version")
}

func TestElasticsearchClient_Stats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	stats, err := client.Stats()
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Contains(t, stats, "cluster_name")
	assert.Contains(t, stats, "nodes")
}

func TestElasticsearchClient_IndexOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	ctx := context.Background()
	indexName := "test-index-" + time.Now().Format("20060102150405")

	// 清理测试数据
	defer func() {
		_ = client.DeleteIndex(ctx, indexName)
	}()

	t.Run("index not exists", func(t *testing.T) {
		exists, err := client.IndexExists(ctx, indexName)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("create index", func(t *testing.T) {
		mapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type": "text",
					},
					"content": map[string]interface{}{
						"type": "text",
					},
					"timestamp": map[string]interface{}{
						"type": "date",
					},
				},
			},
		}

		err := client.CreateIndex(ctx, indexName, mapping)
		assert.NoError(t, err)

		// 验证索引存在
		exists, err := client.IndexExists(ctx, indexName)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("delete index", func(t *testing.T) {
		err := client.DeleteIndex(ctx, indexName)
		assert.NoError(t, err)

		// 验证索引不存在
		exists, err := client.IndexExists(ctx, indexName)
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestElasticsearchClient_DocumentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	ctx := context.Background()
	indexName := "test-docs-" + time.Now().Format("20060102150405")
	docID := "test-doc-1"

	// 创建测试索引
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
				"timestamp": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	err = client.CreateIndex(ctx, indexName, mapping)
	require.NoError(t, err)

	// 清理测试数据
	defer func() {
		_ = client.DeleteIndex(ctx, indexName)
	}()

	t.Run("index document", func(t *testing.T) {
		document := map[string]interface{}{
			"title":     "Test Document",
			"content":   "This is a test document for Elasticsearch operations",
			"timestamp": time.Now(),
		}

		err := client.IndexDocument(ctx, indexName, docID, document)
		assert.NoError(t, err)

		// 等待索引刷新
		time.Sleep(1 * time.Second)
	})

	t.Run("get document", func(t *testing.T) {
		var result map[string]interface{}
		err := client.GetDocument(ctx, indexName, docID, &result)
		assert.NoError(t, err)

		assert.Equal(t, "Test Document", result["title"])
		assert.Equal(t, "This is a test document for Elasticsearch operations", result["content"])
	})

	t.Run("update document", func(t *testing.T) {
		update := map[string]interface{}{
			"title": "Updated Test Document",
		}

		err := client.UpdateDocument(ctx, indexName, docID, update)
		assert.NoError(t, err)

		// 等待索引刷新
		time.Sleep(1 * time.Second)

		// 验证更新
		var result map[string]interface{}
		err = client.GetDocument(ctx, indexName, docID, &result)
		assert.NoError(t, err)

		assert.Equal(t, "Updated Test Document", result["title"])
		assert.Equal(t, "This is a test document for Elasticsearch operations", result["content"])
	})

	t.Run("delete document", func(t *testing.T) {
		err := client.DeleteDocument(ctx, indexName, docID)
		assert.NoError(t, err)

		// 等待索引刷新
		time.Sleep(1 * time.Second)

		// 验证删除
		var result map[string]interface{}
		err = client.GetDocument(ctx, indexName, docID, &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document not found")
	})
}

func TestElasticsearchClient_Search(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	ctx := context.Background()
	indexName := "test-search-" + time.Now().Format("20060102150405")

	// 创建测试索引
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	err = client.CreateIndex(ctx, indexName, mapping)
	require.NoError(t, err)

	// 清理测试数据
	defer func() {
		_ = client.DeleteIndex(ctx, indexName)
	}()

	// 索引测试文档
	documents := []map[string]interface{}{
		{
			"title":    "First Document",
			"content":  "This is the first test document about Elasticsearch",
			"category": "test",
		},
		{
			"title":    "Second Document",
			"content":  "This is the second test document about search",
			"category": "search",
		},
		{
			"title":    "Third Document",
			"content":  "This is the third test document about indexing",
			"category": "index",
		},
	}

	for i, doc := range documents {
		docID := fmt.Sprintf("doc-%d", i+1)
		err := client.IndexDocument(ctx, indexName, docID, doc)
		require.NoError(t, err)
	}

	// 等待索引刷新
	time.Sleep(2 * time.Second)

	t.Run("match all query", func(t *testing.T) {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		result, err := client.Search(ctx, indexName, query)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, result.Hits.Total.Value, 0)
		assert.Equal(t, len(documents), result.Hits.Total.Value)
	})

	t.Run("term query", func(t *testing.T) {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					"category": "test",
				},
			},
		}

		result, err := client.Search(ctx, indexName, query)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.Hits.Total.Value)
	})

	t.Run("match query", func(t *testing.T) {
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"content": "search",
				},
			},
		}

		result, err := client.Search(ctx, indexName, query)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, result.Hits.Total.Value, int64(1))
	})
}

func TestElasticsearchClient_Bulk(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		t.Skipf("Skipping test due to connection error: %v", err)
	}

	ctx := context.Background()
	indexName := "test-bulk-" + time.Now().Format("20060102150405")

	// 创建测试索引
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
			},
		},
	}

	err = client.CreateIndex(ctx, indexName, mapping)
	require.NoError(t, err)

	// 清理测试数据
	defer func() {
		_ = client.DeleteIndex(ctx, indexName)
	}()

	t.Run("bulk index operations", func(t *testing.T) {
		operations := []BulkOperation{
			{
				Action:   "index",
				Index:    indexName,
				DocumentID: "bulk-1",
				Document: map[string]interface{}{
					"title":   "Bulk Document 1",
					"content": "First bulk document",
				},
			},
			{
				Action:   "index",
				Index:    indexName,
				DocumentID: "bulk-2",
				Document: map[string]interface{}{
					"title":   "Bulk Document 2",
					"content": "Second bulk document",
				},
			},
		}

		err := client.Bulk(ctx, operations)
		assert.NoError(t, err)

		// 等待索引刷新
		time.Sleep(2 * time.Second)

		// 验证文档被索引
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		result, err := client.Search(ctx, indexName, query)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Hits.Total.Value)
	})

	t.Run("bulk delete operations", func(t *testing.T) {
		deleteOperations := []BulkOperation{
			{
				Action:   "delete",
				Index:    indexName,
				DocumentID: "bulk-1",
			},
			{
				Action:   "delete",
				Index:    indexName,
				DocumentID: "bulk-2",
			},
		}

		err := client.Bulk(ctx, deleteOperations)
		assert.NoError(t, err)

		// 等待索引刷新
		time.Sleep(2 * time.Second)

		// 验证文档被删除
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		result, err := client.Search(ctx, indexName, query)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), result.Hits.Total.Value)
	})
}

// Benchmarks
func BenchmarkElasticsearchClient_IndexDocument(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration test in short mode")
	}

	config := DefaultElasticsearchConfig()
	client, err := NewElasticsearchClient(config)
	if err != nil {
		b.Skipf("Skipping benchmark due to connection error: %v", err)
	}

	ctx := context.Background()
	indexName := "bench-index-" + time.Now().Format("20060102150405")

	// 创建测试索引
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
			},
		},
	}

	err = client.CreateIndex(ctx, indexName, mapping)
	if err != nil {
		b.Fatalf("Failed to create index: %v", err)
	}

	// 清理测试数据
	defer func() {
		_ = client.DeleteIndex(ctx, indexName)
	}()

	document := map[string]interface{}{
		"title": "Benchmark Document",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		docID := fmt.Sprintf("doc-%d", i)
		err := client.IndexDocument(ctx, indexName, docID, document)
		if err != nil {
			b.Fatalf("IndexDocument failed: %v", err)
		}
	}
}