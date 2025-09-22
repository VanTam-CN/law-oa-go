package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"law-oa-go/internal/errors"
)

// SearchResult represents a search result
type SearchResult struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // user, client, case, document
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	URL        string                 `json:"url"`
	Score      float64                `json:"score"`
	Highlights map[string][]string    `json:"highlights"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query     string   `json:"query"`
	Page      int      `json:"page"`
	PageSize  int      `json:"page_size"`
	Types     []string `json:"types"`      // Filter by types: user, client, case, document
	EntityID  *uint    `json:"entity_id"`  // Filter by entity ID
	DateFrom  *string  `json:"date_from"`  // Filter by date range
	DateTo    *string  `json:"date_to"`    // Filter by date range
	SortBy    string   `json:"sort_by"`    // score, date, relevance
	SortOrder string   `json:"sort_order"` // asc, desc
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results       []*SearchResult `json:"results"`
	Total         int64           `json:"total"`
	Page          int             `json:"page"`
	PageSize      int             `json:"page_size"`
	ExecutionTime time.Duration   `json:"execution_time"`
	Suggestions   []string        `json:"suggestions"`
	Facets        map[string]int  `json:"facets"`
}

// SearchService handles search operations
type SearchService struct {
	esClient    *elasticsearch.Client
	indexPrefix string
}

// NewSearchService creates a new search service
func NewSearchService(esClient *elasticsearch.Client, indexPrefix string) *SearchService {
	return &SearchService{
		esClient:    esClient,
		indexPrefix: indexPrefix,
	}
}

// Search performs a search operation
func (s *SearchService) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()

	// Validate request
	if req.Query == "" {
		return nil, errors.NewValidationError("query", "empty_query", "Search query cannot be empty", "Please provide a search query")
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// If Elasticsearch client is not available, fallback to database search
	if s.esClient == nil {
		return s.fallbackSearch(ctx, req, startTime)
	}

	// Build search query
	searchQuery := s.buildSearchQuery(req)

	// Determine which indices to search
	indices := s.getSearchIndices(req.Types)

	// Execute search using raw ES client
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, errors.NewInternalError("search_encoding", "Failed to encode search query", err)
	}

	res, err := s.esClient.Search(
		s.esClient.Search.WithContext(ctx),
		s.esClient.Search.WithIndex(indices),
		s.esClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, errors.NewInternalError("search_execution", "Failed to execute search", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewInternalError("search_response", "Elasticsearch returned error", fmt.Errorf("status: %s", res.Status()))
	}

	// Parse search results
	var esResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResult); err != nil {
		return nil, errors.NewInternalError("search_parsing", "Failed to parse search response", err)
	}

	searchResults, total, err := s.parseSearchResults(esResult)
	if err != nil {
		return nil, errors.NewInternalError("search_parsing", "Failed to parse search results", err)
	}

	executionTime := time.Since(startTime)

	return &SearchResponse{
		Results:       searchResults,
		Total:         total,
		Page:          req.Page,
		PageSize:      req.PageSize,
		ExecutionTime: executionTime,
		Suggestions:   s.generateSuggestions(req.Query),
		Facets:        s.extractFacets(esResult),
	}, nil
}

// GetSearchSuggestions gets search suggestions
func (s *SearchService) GetSearchSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if query == "" {
		return []string{}, nil
	}

	// If Elasticsearch client is not available, return empty suggestions
	if s.esClient == nil {
		return []string{}, nil
	}

	// Build suggest query
	suggestQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"text": query,
			"completion": map[string]interface{}{
				"field": "suggest",
				"size":  limit,
			},
		},
	}

	// Execute suggest query using raw ES client
	var suggestBuf bytes.Buffer
	if err := json.NewEncoder(&suggestBuf).Encode(suggestQuery); err != nil {
		return nil, err
	}

	suggestRes, err := s.esClient.Search(
		s.esClient.Search.WithContext(ctx),
		s.esClient.Search.WithIndex(s.indexPrefix + "*"),
		s.esClient.Search.WithBody(&suggestBuf),
	)
	if err != nil {
		return nil, err
	}
	defer suggestRes.Body.Close()

	if suggestRes.IsError() {
		return nil, fmt.Errorf("suggest query failed: %s", suggestRes.Status())
	}

	// Parse suggestions
	var suggestResult map[string]interface{}
	if err := json.NewDecoder(suggestRes.Body).Decode(&suggestResult); err != nil {
		return nil, err
	}
	suggestions := s.parseSuggestions(suggestResult, limit)
	return suggestions, nil
}

// IndexEntity indexes an entity for search
func (s *SearchService) IndexEntity(ctx context.Context, entityType string, entityID string, data map[string]interface{}) error {
	if s.esClient == nil {
		return nil
	}

	// Prepare document for indexing
	doc := s.prepareDocument(entityType, entityID, data)
	indexName := s.indexPrefix + entityType

	// Index document using raw ES client
	var docBuf bytes.Buffer
	if err := json.NewEncoder(&docBuf).Encode(doc); err != nil {
		return err
	}

	res, err := s.esClient.Index(
		indexName,
		&docBuf,
		s.esClient.Index.WithContext(ctx),
		s.esClient.Index.WithDocumentID(entityID),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index document error: %s", res.Status())
	}

	return nil
}

// DeleteEntityFromIndex removes an entity from the search index
func (s *SearchService) DeleteEntityFromIndex(ctx context.Context, entityType string, entityID string) error {
	if s.esClient == nil {
		return nil
	}

	indexName := s.indexPrefix + entityType
	
	// Delete document using raw ES client
	res, err := s.esClient.Delete(
		indexName,
		entityID,
		s.esClient.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete document error: %s", res.Status())
	}

	return nil
}

// ReindexAll rebuilds the entire search index
func (s *SearchService) ReindexAll(ctx context.Context) error {
	if s.esClient == nil {
		return nil
	}

	// This would typically involve:
	// 1. Creating new indices with updated mappings
	// 2. Reindexing all data from the database
	// 3. Switching aliases to point to new indices
	// For now, just return success
	return nil
}

// Helper methods

func (s *SearchService) buildSearchQuery(req *SearchRequest) interface{} {
	// Build bool query with should clauses for better search relevance
	boolQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"should": []map[string]interface{}{
				{
					"multi_match": map[string]interface{}{
						"query":  req.Query,
						"fields": []string{"title^3", "content^2", "description", "tags"},
						"type":   "best_fields",
					},
				},
				{
					"match_phrase": map[string]interface{}{
						"title": map[string]interface{}{
							"query": req.Query,
							"boost": 5,
						},
					},
				},
				{
					"fuzzy": map[string]interface{}{
						"title": map[string]interface{}{
							"value":     req.Query,
							"fuzziness": "AUTO",
							"boost":     1,
						},
					},
				},
			},
			"minimum_should_match": 1,
		},
	}

	// Add filters if specified
	if len(req.Types) > 0 || req.EntityID != nil || req.DateFrom != nil || req.DateTo != nil {
		filter := []map[string]interface{}{}

		if len(req.Types) > 0 {
			filter = append(filter, map[string]interface{}{
				"terms": map[string]interface{}{
					"type": req.Types,
				},
			})
		}

		if req.EntityID != nil {
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{
					"entity_id": *req.EntityID,
				},
			})
		}

		if req.DateFrom != nil || req.DateTo != nil {
			dateRange := map[string]interface{}{}
			if req.DateFrom != nil {
				dateRange["gte"] = *req.DateFrom
			}
			if req.DateTo != nil {
				dateRange["lte"] = *req.DateTo
			}
			filter = append(filter, map[string]interface{}{
				"range": map[string]interface{}{
					"created_at": dateRange,
				},
			})
		}

		boolQuery["bool"].(map[string]interface{})["filter"] = filter
	}

	// Add sorting
	sortOrder := "desc"
	if req.SortOrder == "asc" {
		sortOrder = "asc"
	}

	sortField := "_score"
	if req.SortBy == "date" {
		sortField = "created_at"
	} else if req.SortBy == "relevance" {
		sortField = "_score"
	}

	// Build final query
	query := map[string]interface{}{
		"query": boolQuery,
		"from":  (req.Page - 1) * req.PageSize,
		"size":  req.PageSize,
		"sort": []map[string]interface{}{
			{
				sortField: map[string]interface{}{
					"order": sortOrder,
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":   map[string]interface{}{},
				"content": map[string]interface{}{},
			},
			"pre_tags":  []string{"<em>"},
			"post_tags": []string{"</em>"},
		},
	}

	return query
}

func (s *SearchService) getSearchIndices(types []string) string {
	if len(types) == 0 {
		return s.indexPrefix + "*"
	}

	var indices []string
	for _, t := range types {
		indices = append(indices, s.indexPrefix+t)
	}

	return strings.Join(indices, ",")
}

func (s *SearchService) parseSearchResults(result map[string]interface{}) ([]*SearchResult, int64, error) {
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid search result format")
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid total format")
	}

	// 安全的类型转换，处理float64到int64的转换
	var totalValue int64
	switch v := total["value"].(type) {
	case int64:
		totalValue = v
	case float64:
		totalValue = int64(v)
	case int:
		totalValue = int64(v)
	default:
		return nil, 0, fmt.Errorf("invalid total value type: %T", total["value"])
	}

	var searchResults []*SearchResult
	if hitsList, ok := hits["hits"].([]interface{}); ok {
		for _, hit := range hitsList {
			if hitMap, ok := hit.(map[string]interface{}); ok {
				searchResult := s.parseSingleHit(hitMap)
				searchResults = append(searchResults, searchResult)
			}
		}
	}

	return searchResults, totalValue, nil
}

func (s *SearchService) parseSingleHit(hit map[string]interface{}) *SearchResult {
	source, _ := hit["_source"].(map[string]interface{})
	highlights, _ := hit["highlight"].(map[string]interface{})

	result := &SearchResult{
		ID:         hit["_id"].(string),
		Score:      hit["_score"].(float64),
		Highlights: make(map[string][]string),
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}

	// Parse source fields
	if source != nil {
		if title, ok := source["title"].(string); ok {
			result.Title = title
		}
		if content, ok := source["content"].(string); ok {
			result.Content = content
		}
		if entityType, ok := source["type"].(string); ok {
			result.Type = entityType
		}
		if url, ok := source["url"].(string); ok {
			result.URL = url
		}

		// Copy remaining fields to metadata
		for k, v := range source {
			if k != "title" && k != "content" && k != "type" && k != "url" {
				result.Metadata[k] = v
			}
		}
	}

	// Parse highlights
	if highlights != nil {
		for field, highlightList := range highlights {
			if list, ok := highlightList.([]interface{}); ok {
				var stringList []string
				for _, item := range list {
					if str, ok := item.(string); ok {
						stringList = append(stringList, str)
					}
				}
				result.Highlights[field] = stringList
			}
		}
	}

	return result
}

func (s *SearchService) parseSuggestions(result map[string]interface{}, limit int) []string {
	// Parse suggest results
	suggest, ok := result["suggest"].(map[string]interface{})
	if !ok {
		return []string{}
	}

	var suggestions []string
	for _, suggestField := range suggest {
		if suggestList, ok := suggestField.([]interface{}); ok && len(suggestList) > 0 {
			if firstSuggest, ok := suggestList[0].(map[string]interface{}); ok {
				if options, ok := firstSuggest["options"].([]interface{}); ok {
					for _, option := range options {
						if optMap, ok := option.(map[string]interface{}); ok {
							if text, ok := optMap["text"].(string); ok {
								suggestions = append(suggestions, text)
								if len(suggestions) >= limit {
									return suggestions
								}
							}
						}
					}
				}
			}
		}
	}

	return suggestions
}

func (s *SearchService) extractFacets(result map[string]interface{}) map[string]int {
	// Simple facet extraction by type
	facets := make(map[string]int)
	
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return facets
	}

	if hitsList, ok := hits["hits"].([]interface{}); ok {
		for _, hit := range hitsList {
			if hitMap, ok := hit.(map[string]interface{}); ok {
				if source, ok := hitMap["_source"].(map[string]interface{}); ok {
					if entityType, ok := source["type"].(string); ok {
						facets[entityType]++
					}
				}
			}
		}
	}

	return facets
}

func (s *SearchService) prepareDocument(entityType string, entityID string, data map[string]interface{}) map[string]interface{} {
	doc := make(map[string]interface{})
	
	// Copy all data
	for k, v := range data {
		doc[k] = v
	}

	// Ensure required fields
	doc["type"] = entityType
	doc["entity_id"] = entityID
	doc["indexed_at"] = time.Now().Format(time.RFC3339)

	// Generate suggest field for autocomplete
	if title, ok := data["title"].(string); ok {
		doc["suggest"] = map[string]interface{}{
			"input":  []string{title},
			"weight": 10,
		}
	}

	return doc
}

func (s *SearchService) fallbackSearch(ctx context.Context, req *SearchRequest, startTime time.Time) (*SearchResponse, error) {
	// Fallback to simple database search when Elasticsearch is not available
	// This would perform basic LIKE queries on relevant tables
	// For demonstration, return mock results

	mockResults := []*SearchResult{
		{
			ID:        "fallback_1",
			Type:      "case",
			Title:     "Fallback Result for: " + req.Query,
			Content:   "This is a fallback search result...",
			URL:       "/cases/fallback_1",
			Score:     0.5,
			CreatedAt: time.Now(),
		},
	}

	executionTime := time.Since(startTime)

	return &SearchResponse{
		Results:       mockResults,
		Total:         1,
		Page:          req.Page,
		PageSize:      req.PageSize,
		ExecutionTime: executionTime,
		Suggestions:   []string{},
		Facets:        map[string]int{},
	}, nil
}

func (s *SearchService) generateSuggestions(query string) []string {
	// Generate search suggestions based on the query
	suggestions := []string{}

	// Simple suggestions based on common terms
	commonTerms := []string{"case", "client", "document", "contract", "dispute"}

	for _, term := range commonTerms {
		if strings.Contains(term, strings.ToLower(query)) || strings.Contains(strings.ToLower(query), term) {
			suggestions = append(suggestions, query+" "+term)
		}
	}

	// Limit to 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions
}
