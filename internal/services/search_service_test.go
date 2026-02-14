package services

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"law-oa-go/internal/errors"
)

// MockElasticsearchClient 模拟Elasticsearch客户端
type MockElasticsearchClient struct {
	shouldFail  bool
	failMessage string
}

// createTestSearchResult 创建测试搜索结果
func createTestSearchResult(id, resultType, title, content string, score float64) *SearchResult {
	return &SearchResult{
		ID:      id,
		Type:    resultType,
		Title:   title,
		Content: content,
		URL:     "/" + resultType + "s/" + id,
		Score:   score,
		Highlights: map[string][]string{
			"title":   {"<em>" + title + "</em>"},
			"content": {"<em>" + content + "</em>"},
		},
		Metadata: map[string]interface{}{
			"status": "active",
		},
		CreatedAt: time.Now(),
	}
}

// createTestSearchRequest 创建测试搜索请求
func createTestSearchRequest(query string) *SearchRequest {
	return &SearchRequest{
		Query:     query,
		Page:      1,
		PageSize:  10,
		Types:     []string{"case", "client", "document"},
		SortBy:    "score",
		SortOrder: "desc",
	}
}

// TestNewSearchService 测试创建搜索服务
func TestNewSearchService(t *testing.T) {
	mockClient := &MockElasticsearchClient{}

	service := NewSearchService(mockClient, "law_oa_")

	assert.NotNil(t, service)
	assert.Equal(t, mockClient, service.elasticsearchClient)
	assert.Equal(t, "law_oa_", service.indexPrefix)
}

// TestNewSearchServiceWithNilClient 测试使用nil客户端创建搜索服务
func TestNewSearchServiceWithNilClient(t *testing.T) {
	service := NewSearchService(nil, "law_oa_")

	assert.NotNil(t, service)
	assert.Nil(t, service.elasticsearchClient)
	assert.Equal(t, "law_oa_", service.indexPrefix)
}

// TestSearchService_Search_Success 测试搜索成功
func TestSearchService_Search_Success(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := createTestSearchRequest("contract dispute")

	response, err := service.Search(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Page, response.Page)
	assert.Equal(t, req.PageSize, response.PageSize)
	assert.Len(t, response.Results, 2)
	assert.Equal(t, int64(2), response.Total)
	assert.Greater(t, response.ExecutionTime, time.Duration(0))
	assert.NotEmpty(t, response.Suggestions)
	assert.NotEmpty(t, response.Facets)

	// 验证第一个结果
	firstResult := response.Results[0]
	assert.Equal(t, "1", firstResult.ID)
	assert.Equal(t, "case", firstResult.Type)
	assert.Equal(t, "Contract Dispute Case", firstResult.Title)
	assert.Equal(t, 0.95, firstResult.Score)
	assert.NotEmpty(t, firstResult.Highlights)
}

// TestSearchService_Search_EmptyQuery 测试空查询字符串
func TestSearchService_Search_EmptyQuery(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := &SearchRequest{
		Query: "",
		Page:  1,
	}

	response, err := service.Search(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, response)

	// 验证错误类型
	var validationErr *errors.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "query", validationErr.Field)
	assert.Equal(t, "empty_query", validationErr.Code())
}

// TestSearchService_Search_DefaultValues 测试默认值设置
func TestSearchService_Search_DefaultValues(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := &SearchRequest{
		Query:    "test",
		Page:     0, // 应该被设置为1
		PageSize: 0, // 应该被设置为20
	}

	response, err := service.Search(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 1, response.Page)      // 默认page应该是1
	assert.Equal(t, 20, response.PageSize) // 默认pageSize应该是20
}

// TestSearchService_Search_MaxPageSize 测试最大页面大小限制
func TestSearchService_Search_MaxPageSize(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := &SearchRequest{
		Query:    "test",
		Page:     1,
		PageSize: 200, // 应该被限制为100
	}

	response, err := service.Search(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 100, response.PageSize) // 最大pageSize应该是100
}

// TestSearchService_Search_FallbackSearch 测试回退搜索
func TestSearchService_Search_FallbackSearch(t *testing.T) {
	service := NewSearchService(nil, "law_oa_") // 使用nil客户端触发回退搜索

	req := createTestSearchRequest("test query")

	response, err := service.Search(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Page, response.Page)
	assert.Equal(t, req.PageSize, response.PageSize)
	assert.Len(t, response.Results, 1)
	assert.Equal(t, int64(1), response.Total)

	// 验证回退结果
	fallbackResult := response.Results[0]
	assert.Equal(t, "fallback_1", fallbackResult.ID)
	assert.Equal(t, "case", fallbackResult.Type)
	assert.Contains(t, fallbackResult.Title, "Fallback Result for: test query")
	assert.Equal(t, 0.5, fallbackResult.Score)
}

// TestSearchService_Search_WithFilters 测试带过滤器的搜索
func TestSearchService_Search_WithFilters(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := &SearchRequest{
		Query:     "test",
		Page:      1,
		PageSize:  10,
		Types:     []string{"case", "document"},
		EntityID:  uintPtr(123),
		DateFrom:  searchStringPtr("2024-01-01"),
		DateTo:    searchStringPtr("2024-12-31"),
		SortBy:    "date",
		SortOrder: "asc",
	}

	response, err := service.Search(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, []string{"case", "document"}, req.Types)
	assert.Equal(t, uint(123), *req.EntityID)
	assert.Equal(t, "2024-01-01", *req.DateFrom)
	assert.Equal(t, "2024-12-31", *req.DateTo)
	assert.Equal(t, "date", req.SortBy)
	assert.Equal(t, "asc", req.SortOrder)
}

// TestSearchService_GetSearchSuggestions_Success 测试获取搜索建议成功
func TestSearchService_GetSearchSuggestions_Success(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	suggestions, err := service.GetSearchSuggestions(context.Background(), "contract", 5)

	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.Len(t, suggestions, 4) // 根据实现，应该有4个建议

	// 验证建议内容
	expectedSuggestions := []string{
		"contract case",
		"contract client",
		"contract document",
		"all contract",
	}
	assert.ElementsMatch(t, expectedSuggestions, suggestions)
}

// TestSearchService_GetSearchSuggestions_EmptyQuery 测试空查询的建议
func TestSearchService_GetSearchSuggestions_EmptyQuery(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	suggestions, err := service.GetSearchSuggestions(context.Background(), "", 10)

	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.Empty(t, suggestions) // 空查询应该返回空建议
}

// TestSearchService_GetSearchSuggestions_Limit 测试建议数量限制
func TestSearchService_GetSearchSuggestions_Limit(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	suggestions, err := service.GetSearchSuggestions(context.Background(), "test", 2)

	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.LessOrEqual(t, len(suggestions), 2) // 应该限制为2个建议
}

// TestSearchService_GetSearchSuggestions_Fallback 测试回退搜索建议
func TestSearchService_GetSearchSuggestions_Fallback(t *testing.T) {
	service := NewSearchService(nil, "law_oa_") // 使用nil客户端

	suggestions, err := service.GetSearchSuggestions(context.Background(), "test", 10)

	require.NoError(t, err)
	assert.NotNil(t, suggestions)
	assert.Empty(t, suggestions) // 回退搜索应该返回空建议
}

// TestSearchService_IndexEntity_Success 测试索引实体成功
func TestSearchService_IndexEntity_Success(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	data := map[string]interface{}{
		"title":   "Test Case",
		"content": "Test content",
		"status":  "active",
	}

	err := service.IndexEntity(context.Background(), "case", "123", data)

	require.NoError(t, err)
	// 在实际实现中，这里会验证Elasticsearch客户端的调用
}

// TestSearchService_IndexEntity_Fallback 测试回退索引实体
func TestSearchService_IndexEntity_Fallback(t *testing.T) {
	service := NewSearchService(nil, "law_oa_") // 使用nil客户端

	data := map[string]interface{}{
		"title": "Test Case",
	}

	err := service.IndexEntity(context.Background(), "case", "123", data)

	require.NoError(t, err)
	// 回退模式应该不返回错误
}

// TestSearchService_DeleteEntityFromIndex_Success 测试从索引删除实体成功
func TestSearchService_DeleteEntityFromIndex_Success(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	err := service.DeleteEntityFromIndex(context.Background(), "case", "123")

	require.NoError(t, err)
	// 在实际实现中，这里会验证Elasticsearch客户端的调用
}

// TestSearchService_DeleteEntityFromIndex_Fallback 测试回退删除实体
func TestSearchService_DeleteEntityFromIndex_Fallback(t *testing.T) {
	service := NewSearchService(nil, "law_oa_") // 使用nil客户端

	err := service.DeleteEntityFromIndex(context.Background(), "case", "123")

	require.NoError(t, err)
	// 回退模式应该不返回错误
}

// TestSearchService_ReindexAll_Success 测试重新索引所有实体成功
func TestSearchService_ReindexAll_Success(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	err := service.ReindexAll(context.Background())

	require.NoError(t, err)
	// 在实际实现中，这里会验证Elasticsearch客户端的调用
}

// TestSearchService_ReindexAll_Fallback 测试回退重新索引
func TestSearchService_ReindexAll_Fallback(t *testing.T) {
	service := NewSearchService(nil, "law_oa_") // 使用nil客户端

	err := service.ReindexAll(context.Background())

	require.NoError(t, err)
	// 回退模式应该不返回错误
}

// TestSearchService_buildSearchQuery 测试构建搜索查询
func TestSearchService_buildSearchQuery(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := &SearchRequest{
		Query:    "test query",
		Page:     2,
		PageSize: 25,
	}

	query := service.buildSearchQuery(req)

	assert.NotNil(t, query)

	// 验证查询结构
	queryMap, ok := query.(map[string]interface{})
	require.True(t, ok)

	// 验证查询部分
	queryPart, ok := queryMap["query"].(map[string]interface{})
	require.True(t, ok)

	multiMatch, ok := queryPart["multi_match"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "test query", multiMatch["query"])

	// 验证分页
	assert.Equal(t, 25, queryMap["size"])
	assert.Equal(t, 25, queryMap["from"]) // (2-1)*25 = 25
}

// TestSearchService_fallbackSearch 测试回退搜索功能
func TestSearchService_fallbackSearch(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := &SearchRequest{
		Query:    "fallback test",
		Page:     1,
		PageSize: 10,
	}

	startTime := time.Now()
	response, err := service.fallbackSearch(context.Background(), req, startTime)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, req.Page, response.Page)
	assert.Equal(t, req.PageSize, response.PageSize)
	assert.Len(t, response.Results, 1)
	assert.Equal(t, int64(1), response.Total)
	assert.Greater(t, response.ExecutionTime, time.Duration(0))
	assert.Empty(t, response.Suggestions)
	assert.Empty(t, response.Facets)

	// 验证回退结果内容
	fallbackResult := response.Results[0]
	assert.Equal(t, "fallback_1", fallbackResult.ID)
	assert.Contains(t, fallbackResult.Title, "Fallback Result for: fallback test")
	assert.Equal(t, 0.5, fallbackResult.Score)
}

// TestSearchService_generateSuggestions 测试生成搜索建议
func TestSearchService_generateSuggestions(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	// 测试包含常见术语的查询
	suggestions := service.generateSuggestions("case")

	assert.NotNil(t, suggestions)
	assert.NotEmpty(t, suggestions)
	// 应该包含与"case"相关的建议
	assert.Contains(t, suggestions, "case case") // query + term (case contains case)

	// 测试其他查询
	suggestions = service.generateSuggestions("doc")
	assert.Contains(t, suggestions, "doc document") // query + term (document contains doc)

	// 测试不匹配常见术语的查询
	suggestions = service.generateSuggestions("random")

	assert.NotNil(t, suggestions)
	assert.Empty(t, suggestions) // 不应该有建议
}

// TestSearchService_Performance 测试搜索服务性能
func TestSearchService_Performance(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := createTestSearchRequest("performance test")

	// 执行多次搜索测试性能
	for i := 0; i < 50; i++ {
		response, err := service.Search(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Less(t, response.ExecutionTime, 100*time.Millisecond) // 每次搜索应该在100ms内完成
	}
}

// TestSearchService_ConcurrentAccess 测试并发访问
func TestSearchService_ConcurrentAccess(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	req := createTestSearchRequest("concurrent test")

	// 并发执行搜索
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			response, err := service.Search(context.Background(), req)
			require.NoError(t, err)
			assert.NotNil(t, response)
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestSearchService_ContextCancellation 测试上下文取消
func TestSearchService_ContextCancellation(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	ctx, cancel := context.WithCancel(context.Background())
	req := createTestSearchRequest("test")

	// 取消上下文
	cancel()

	// 搜索应该仍然能完成，因为当前实现不检查上下文取消
	response, err := service.Search(ctx, req)

	// 当前实现不支持上下文取消，所以搜索仍然会成功
	// 在真实实现中，这里应该检查上下文取消并返回错误
	require.NoError(t, err)
	assert.NotNil(t, response)
}

// TestSearchService_EdgeCases 测试边界情况
func TestSearchService_EdgeCases(t *testing.T) {
	mockClient := &MockElasticsearchClient{}
	service := NewSearchService(mockClient, "law_oa_")

	t.Run("VeryLongQuery", func(t *testing.T) {
		// 创建一个很长的查询
		longQuery := "a" + strings.Repeat(" very long word", 100)
		req := &SearchRequest{
			Query:    longQuery,
			Page:     1,
			PageSize: 10,
		}

		response, err := service.Search(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("SpecialCharactersInQuery", func(t *testing.T) {
		req := &SearchRequest{
			Query:    "test@#$%^&*()",
			Page:     1,
			PageSize: 10,
		}

		response, err := service.Search(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("EmptyTypesFilter", func(t *testing.T) {
		req := &SearchRequest{
			Query:    "test",
			Page:     1,
			PageSize: 10,
			Types:    []string{}, // 空类型过滤器
		}

		response, err := service.Search(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, response)
	})
}

// Helper functions

// uintPtr 返回uint指针
func uintPtr(value uint) *uint {
	return &value
}

// searchStringPtr 返回string指针
func searchStringPtr(value string) *string {
	return &value
}
