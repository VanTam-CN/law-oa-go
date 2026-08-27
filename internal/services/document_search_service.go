package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// DocumentSearchService handles document search operations
type DocumentSearchService struct {
	docRepo repositories.DocumentRepository
}

// ErrDocumentSearchUnavailable keeps the legacy search service from returning
// unscoped or synthetic results until it is backed by a viewer-aware index.
var ErrDocumentSearchUnavailable = fmt.Errorf("DOCUMENT_SEARCH_UNAVAILABLE: document search is not connected to a viewer-scoped index")

// NewDocumentSearchService creates a new document search service
func NewDocumentSearchService(docRepo repositories.DocumentRepository) *DocumentSearchService {
	return &DocumentSearchService{
		docRepo: docRepo,
	}
}

// DocumentSearchRequest represents a document search request
type DocumentSearchRequest struct {
	Query       string   `json:"query" form:"query"`
	Category    string   `json:"category" form:"category"`
	EntityType  string   `json:"entity_type" form:"entity_type"`
	EntityID    uint     `json:"entity_id" form:"entity_id"`
	Tags        []string `json:"tags" form:"tags"`
	DateFrom    string   `json:"date_from" form:"date_from"`
	DateTo      string   `json:"date_to" form:"date_to"`
	FileSizeMin int64    `json:"file_size_min" form:"file_size_min"`
	FileSizeMax int64    `json:"file_size_max" form:"file_size_max"`
	SortBy      string   `json:"sort_by" form:"sort_by"`       // name, size, created_at, updated_at
	SortOrder   string   `json:"sort_order" form:"sort_order"` // asc, desc
	Page        int      `json:"page" form:"page"`
	PageSize    int      `json:"page_size" form:"page_size"`
}

// DocumentSearchResult represents a document search result
type DocumentSearchResult struct {
	Documents   []*DocumentSearchItem `json:"documents"`
	TotalCount  int64                 `json:"total_count"`
	Page        int                   `json:"page"`
	PageSize    int                   `json:"page_size"`
	TotalPages  int                   `json:"total_pages"`
	Query       string                `json:"query"`
	SearchTime  time.Duration         `json:"search_time"`
	Filters     SearchFilters         `json:"filters"`
	Suggestions []string              `json:"suggestions"`
}

// DocumentSearchItem represents a document in search results
type DocumentSearchItem struct {
	ID          uint              `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Filename    string            `json:"filename"`
	Filesize    int64             `json:"filesize"`
	MimeType    string            `json:"mime_type"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	EntityID    uint              `json:"entity_id"`
	EntityType  string            `json:"entity_type"`
	Status      string            `json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Score       float64           `json:"score"` // Relevance score
	Highlights  map[string]string `json:"highlights,omitempty"`
}

// SearchFilters represents the filters applied to a search
type SearchFilters struct {
	Category    string    `json:"category"`
	EntityType  string    `json:"entity_type"`
	EntityID    uint      `json:"entity_id"`
	Tags        []string  `json:"tags"`
	DateFrom    time.Time `json:"date_from"`
	DateTo      time.Time `json:"date_to"`
	FileSizeMin int64     `json:"file_size_min"`
	FileSizeMax int64     `json:"file_size_max"`
}

// CategoryFilter represents a category filter option
type CategoryFilter struct {
	Category  string `json:"category"`
	Count     int64  `json:"count"`
	Documents int64  `json:"documents"`
}

// EntityTypeFilter represents an entity type filter option
type EntityTypeFilter struct {
	EntityType string `json:"entity_type"`
	Count      int64  `json:"count"`
	Documents  int64  `json:"documents"`
}

// TagCloud represents a tag cloud for document tags
type TagCloud struct {
	Tag   string `json:"tag"`
	Count int64  `json:"count"`
	Size  int    `json:"size"` // For visualization
}

// AdvancedSearchRequest represents an advanced search request
type AdvancedSearchRequest struct {
	Query     string         `json:"query"`
	Filters   []SearchFilter `json:"filters"`
	Operator  string         `json:"operator"` // AND, OR
	SortBy    string         `json:"sort_by"`
	SortOrder string         `json:"sort_order"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
}

// SearchFilter represents a single search filter
type SearchFilter struct {
	Field    string      `json:"field"`    // category, entity_type, tags, created_at, updated_at, filesize
	Operator string      `json:"operator"` // eq, ne, gt, gte, lt, lte, like, in, not_in
	Value    interface{} `json:"value"`
}

// SearchDocuments performs document search
func (s *DocumentSearchService) SearchDocuments(ctx context.Context, req *DocumentSearchRequest) (*DocumentSearchResult, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

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
		if req.SortBy == "" {
			req.SortBy = "created_at"
		}
		if req.SortOrder == "" {
			req.SortOrder = "desc"
		}

		// Parse date filters
		dateFrom, dateTo := s.parseDateFilters(req.DateFrom, req.DateTo)

		// Build search filters
		filters := SearchFilters{
			Category:    req.Category,
			EntityType:  req.EntityType,
			EntityID:    req.EntityID,
			Tags:        req.Tags,
			DateFrom:    dateFrom,
			DateTo:      dateTo,
			FileSizeMin: req.FileSizeMin,
			FileSizeMax: req.FileSizeMax,
		}

		// Perform search
		documents, total, searchTime, err := s.performSearch(ctx, req, filters)
		if err != nil {
			return nil, errors.InternalError("Search failed", err)
		}

		// Calculate pagination
		totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

		// Generate suggestions
		suggestions := s.generateSuggestions(req.Query)

		result := &DocumentSearchResult{
			Documents:   documents,
			TotalCount:  total,
			Page:        req.Page,
			PageSize:    req.PageSize,
			TotalPages:  totalPages,
			Query:       req.Query,
			SearchTime:  searchTime,
			Filters:     filters,
			Suggestions: suggestions,
		}

		return result, nil
	*/
}

// AdvancedSearch performs advanced document search
func (s *DocumentSearchService) AdvancedSearch(ctx context.Context, req *AdvancedSearchRequest) (*DocumentSearchResult, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

		// Convert to basic search request
		basicReq := &DocumentSearchRequest{
			Query:     req.Query,
			SortBy:    req.SortBy,
			SortOrder: req.SortOrder,
			Page:      req.Page,
			PageSize:  req.PageSize,
		}

		// Set defaults
		if basicReq.Page <= 0 {
			basicReq.Page = 1
		}
		if basicReq.PageSize <= 0 {
			basicReq.PageSize = 20
		}
		if basicReq.PageSize > 100 {
			basicReq.PageSize = 100
		}

		// Apply advanced filters
		filters := SearchFilters{}
		for _, filter := range req.Filters {
			switch filter.Field {
			case "category":
				if filter.Operator == "eq" {
					if val, ok := filter.Value.(string); ok {
						filters.Category = val
					}
				}
			case "entity_type":
				if filter.Operator == "eq" {
					if val, ok := filter.Value.(string); ok {
						filters.EntityType = val
					}
				}
			case "entity_id":
				if filter.Operator == "eq" {
					if val, ok := filter.Value.(float64); ok {
						filters.EntityID = uint(val)
					}
				}
			case "tags":
				if filter.Operator == "in" {
					if tags, ok := filter.Value.([]interface{}); ok {
						filterTags := make([]string, len(tags))
						for i, tag := range tags {
							if tagStr, ok := tag.(string); ok {
								filterTags[i] = tagStr
							}
						}
						filters.Tags = filterTags
					}
				}
			case "created_at":
				s.applyDateFilter(&filters, filter)
			case "filesize":
				s.applySizeFilter(&filters, filter)
			}
		}

		// Perform search
		documents, total, searchTime, err := s.performSearch(ctx, basicReq, filters)
		if err != nil {
			return nil, errors.InternalError("Advanced search failed", err)
		}

		// Calculate pagination
		totalPages := int((total + int64(basicReq.PageSize) - 1) / int64(basicReq.PageSize))

		// Generate suggestions
		suggestions := s.generateSuggestions(basicReq.Query)

		result := &DocumentSearchResult{
			Documents:   documents,
			TotalCount:  total,
			Page:        basicReq.Page,
			PageSize:    basicReq.PageSize,
			TotalPages:  totalPages,
			Query:       basicReq.Query,
			SearchTime:  searchTime,
			Filters:     filters,
			Suggestions: suggestions,
		}

		return result, nil
	*/
}

// GetSearchFilters retrieves available filter options
func (s *DocumentSearchService) GetSearchFilters(ctx context.Context) (map[string]interface{}, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

		// Get unique categories
		categories, err := s.getUniqueCategories(ctx)
		if err != nil {
			return nil, errors.DatabaseError("get_categories", "Failed to get categories", err)
		}

		// Get unique entity types
		entityTypes, err := s.getUniqueEntityTypes(ctx)
		if err != nil {
			return nil, errors.DatabaseError("get_entity_types", "Failed to get entity types", err)
		}

		// Get tag cloud
		tagCloud, err := s.generateTagCloud(ctx)
		if err != nil {
			return nil, errors.DatabaseError("get_tags", "Failed to get tags", err)
		}

		return map[string]interface{}{
			"categories":   categories,
			"entity_types": entityTypes,
			"tag_cloud":    tagCloud,
			"date_ranges": map[string]interface{}{
				"last_24h":   "Last 24 hours",
				"last_week":  "Last week",
				"last_month": "Last month",
				"last_year":  "Last year",
				"custom":     "Custom range",
			},
			"file_sizes": map[string]interface{}{
				"small":  "< 1MB",
				"medium": "1MB - 10MB",
				"large":  "> 10MB",
			},
		}, nil
	*/
}

// GetDocumentCategories retrieves all document categories
func (s *DocumentSearchService) GetDocumentCategories(ctx context.Context) ([]*CategoryFilter, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

		// In a real implementation, query from database
		// For now, return mock data
		categories := []*CategoryFilter{
			{Category: "legal", Count: 50, Documents: 50},
			{Category: "contract", Count: 30, Documents: 30},
			{Category: "invoice", Count: 20, Documents: 20},
			{Category: "report", Count: 25, Documents: 25},
			{Category: "template", Count: 15, Documents: 15},
			{Category: "other", Count: 10, Documents: 10},
		}

		return categories, nil
	*/
}

// GetEntityTypes retrieves all entity types
func (s *DocumentSearchService) GetEntityTypes(ctx context.Context) ([]*EntityTypeFilter, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

		// In a real implementation, query from database
		// For now, return mock data
		entityTypes := []*EntityTypeFilter{
			{EntityType: "case", Count: 60, Documents: 60},
			{EntityType: "client", Count: 40, Documents: 40},
			{EntityType: "template", Count: 20, Documents: 20},
			{EntityType: "user", Count: 30, Documents: 30},
		}

		return entityTypes, nil
	*/
}

// GetTagCloud generates a tag cloud
func (s *DocumentSearchService) GetTagCloud(ctx context.Context) ([]*TagCloud, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

		// In a real implementation, query from database
		// For now, return mock data
		tagCloud := []*TagCloud{
			{Tag: "urgent", Count: 15, Size: 5},
			{Tag: "legal", Count: 25, Size: 4},
			{Tag: "contract", Count: 20, Size: 3},
			{Tag: "signed", Count: 30, Size: 4},
			{Tag: "draft", Count: 10, Size: 2},
			{Tag: "approved", Count: 18, Size: 3},
			{Tag: "confidential", Count: 12, Size: 3},
		}

		return tagCloud, nil
	*/
}

// GetRecentSearches retrieves recent search queries
func (s *DocumentSearchService) GetRecentSearches(ctx context.Context, userID uint, limit int) ([]string, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

		// In a real implementation, query from search logs table
		// For now, return mock data
		recentSearches := []string{
			"contract agreement",
			"legal document",
			"court filing",
			"client agreement",
		}

		if len(recentSearches) > limit {
			recentSearches = recentSearches[:limit]
		}

		return recentSearches, nil
	*/
}

// GetPopularSearches retrieves popular search queries
func (s *DocumentSearchService) GetPopularSearches(ctx context.Context, limit int) ([]string, error) {
	return nil, ErrDocumentSearchUnavailable
	/*

		// In a real implementation, query from search analytics table
		// For now, return mock data
		popularSearches := []string{
			"contract",
			"legal",
			"agreement",
			"court",
			"client",
			"invoice",
			"report",
		}

		if len(popularSearches) > limit {
			popularSearches = popularSearches[:limit]
		}

		return popularSearches, nil
	*/
}

// Helper methods

// performSearch performs the actual search
func (s *DocumentSearchService) performSearch(ctx context.Context, req *DocumentSearchRequest, filters SearchFilters) ([]*DocumentSearchItem, int64, time.Duration, error) {
	// In a real implementation, use Elasticsearch or database search
	// For now, perform database search with filtering
	startTime := time.Now()

	params := &repositories.DocumentListParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Category:   filters.Category,
		EntityType: filters.EntityType,
		EntityID:   filters.EntityID,
		Search:     req.Query,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
	}

	// Apply additional filters
	docModels, total, err := s.docRepo.List(ctx, params)
	if err != nil {
		return nil, 0, 0, err
	}

	// Convert to search items and apply additional filtering
	documents := make([]*DocumentSearchItem, 0, len(docModels))
	for _, doc := range docModels {
		// Apply file size filter
		if filters.FileSizeMin > 0 && doc.Filesize < filters.FileSizeMin {
			continue
		}
		if filters.FileSizeMax > 0 && doc.Filesize > filters.FileSizeMax {
			continue
		}

		// Apply date filter
		if !filters.DateFrom.IsZero() && doc.CreatedAt.Before(filters.DateFrom) {
			continue
		}
		if !filters.DateTo.IsZero() && doc.CreatedAt.After(filters.DateTo) {
			continue
		}

		// Apply tags filter
		if len(filters.Tags) > 0 {
			docTags := strings.Split(doc.Tags, ",")
			matchFound := false
			for _, filterTag := range filters.Tags {
				for _, docTag := range docTags {
					if strings.TrimSpace(docTag) == filterTag {
						matchFound = true
						break
					}
				}
				if matchFound {
					break
				}
			}
			if !matchFound {
				continue
			}
		}

		// Calculate relevance score
		score := s.calculateRelevanceScore(doc, req.Query)

		// Add highlights for text matching
		highlights := s.generateHighlights(doc, req.Query)

		searchItem := &DocumentSearchItem{
			ID:          doc.ID,
			Name:        doc.Name,
			Description: doc.Description,
			Filename:    doc.Filename,
			Filesize:    doc.Filesize,
			MimeType:    doc.MimeType,
			Category:    doc.Category,
			Tags:        s.parseTags(doc.Tags),
			EntityID:    doc.EntityID,
			EntityType:  doc.EntityType,
			Status:      doc.Status,
			CreatedAt:   doc.CreatedAt,
			UpdatedAt:   doc.UpdatedAt,
			Score:       score,
			Highlights:  highlights,
		}

		documents = append(documents, searchItem)
	}

	searchTime := time.Since(startTime)

	return documents, total, searchTime, nil
}

// calculateRelevanceScore calculates relevance score for a document
func (s *DocumentSearchService) calculateRelevanceScore(doc *models.Document, query string) float64 {
	if query == "" {
		return 1.0
	}

	query = strings.ToLower(query)
	score := 0.0

	// Name matching (highest weight)
	if strings.Contains(strings.ToLower(doc.Name), query) {
		score += 10.0
	}

	// Description matching
	if doc.Description != "" && strings.Contains(strings.ToLower(doc.Description), query) {
		score += 5.0
	}

	// Filename matching
	if strings.Contains(strings.ToLower(doc.Filename), query) {
		score += 3.0
	}

	// Category matching
	if doc.Category != "" && strings.Contains(strings.ToLower(doc.Category), query) {
		score += 2.0
	}

	// Tags matching
	if doc.Tags != "" {
		tags := strings.Split(doc.Tags, ",")
		for _, tag := range tags {
			if strings.Contains(strings.TrimSpace(tag), query) {
				score += 1.5
			}
		}
	}

	// Boost recent documents
	daysSinceCreated := time.Since(doc.CreatedAt)
	if daysSinceCreated < 7*24*time.Hour { // Within 7 days
		score += 1.0
	} else if daysSinceCreated < 30*24*time.Hour { // Within 30 days
		score += 0.5
	}

	return score
}

// generateHighlights generates search highlights
func (s *DocumentSearchService) generateHighlights(doc *models.Document, query string) map[string]string {
	if query == "" {
		return nil
	}

	highlights := make(map[string]string)

	// Highlight matching text in name
	if strings.Contains(strings.ToLower(doc.Name), strings.ToLower(query)) {
		highlights["name"] = s.highlightText(doc.Name, query)
	}

	// Highlight matching text in description
	if doc.Description != "" && strings.Contains(strings.ToLower(doc.Description), strings.ToLower(query)) {
		highlights["description"] = s.highlightText(doc.Description, query)
	}

	return highlights
}

// highlightText highlights matching text in a string
func (s *DocumentSearchService) highlightText(text, query string) string {
	if text == "" || query == "" {
		return text
	}

	start := strings.Index(strings.ToLower(text), strings.ToLower(query))
	if start == -1 {
		return text
	}

	end := start + len(query)
	if end > len(text) {
		end = len(text)
	}

	if start > 20 {
		start = start - 20
	}

	if end < len(text)-20 {
		end = end + 20
	}

	if start < 0 {
		start = 0
	}

	if end > len(text) {
		end = len(text)
	}

	return text[start:end]
}

// parseTags parses tags string into array
func (s *DocumentSearchService) parseTags(tagsString string) []string {
	if tagsString == "" {
		return []string{}
	}

	tags := strings.Split(tagsString, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	return tags
}

// parseDateFilters parses date range filters
func (s *DocumentSearchService) parseDateFilters(dateFromStr, dateToStr string) (time.Time, time.Time) {
	var dateFrom, dateTo time.Time

	if dateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = parsed
		}
	}

	if dateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = parsed.Add(23 * time.Hour) // End of day
		}
	}

	return dateFrom, dateTo
}

// applyDateFilter applies date filter to search filters
func (s *DocumentSearchService) applyDateFilter(filters *SearchFilters, filter SearchFilter) {
	if filter.Operator == "gte" {
		if val, ok := filter.Value.(string); ok {
			if parsed, err := time.Parse("2006-01-02", val); err == nil {
				filters.DateFrom = parsed
			}
		}
	} else if filter.Operator == "lte" {
		if val, ok := filter.Value.(string); ok {
			if parsed, err := time.Parse("2006-01-02", val); err == nil {
				filters.DateTo = parsed.Add(23 * time.Hour)
			}
		}
	}
}

// applySizeFilter applies size filter to search filters
func (s *DocumentSearchService) applySizeFilter(filters *SearchFilters, filter SearchFilter) {
	if filter.Operator == "gte" {
		if val, ok := filter.Value.(float64); ok {
			filters.FileSizeMin = int64(val)
		}
	} else if filter.Operator == "lte" {
		if val, ok := filter.Value.(float64); ok {
			filters.FileSizeMax = int64(val)
		}
	}
}

// generateSuggestions generates search suggestions
func (s *DocumentSearchService) generateSuggestions(query string) []string {
	if query == "" {
		return []string{}
	}

	suggestions := make([]string, 0, 5)

	// Simple suggestion logic
	if len(query) < 3 {
		return suggestions
	}

	// Add common prefixes
	suggestions = append(suggestions, query+"ment")
	suggestions = append(suggestions, query+"s")
	suggestions = append(suggestions, query+"ed")

	// Add related terms
	if query == "contract" {
		suggestions = append(suggestions, "agreement")
		suggestions = append(suggestions, "legal")
		suggestions = append(suggestions, "binding")
	}

	return suggestions
}

// getUniqueCategories retrieves unique categories from database
func (s *DocumentSearchService) getUniqueCategories(ctx context.Context) ([]*CategoryFilter, error) {
	// In a real implementation, use SQL GROUP BY
	// For now, return mock data
	return []*CategoryFilter{
		{Category: "legal", Count: 50, Documents: 50},
		{Category: "contract", Count: 30, Documents: 30},
		{Category: "invoice", Count: 20, Documents: 20},
		{Category: "report", Count: 25, Documents: 25},
		{Category: "template", Count: 15, Documents: 15},
		{Category: "other", Count: 10, Documents: 10},
	}, nil
}

// getUniqueEntityTypes retrieves unique entity types from database
func (s *DocumentSearchService) getUniqueEntityTypes(ctx context.Context) ([]*EntityTypeFilter, error) {
	// In a real implementation, use SQL GROUP BY
	// For now, return mock data
	return []*EntityTypeFilter{
		{EntityType: "case", Count: 60, Documents: 60},
		{EntityType: "client", Count: 40, Documents: 40},
		{EntityType: "template", Count: 20, Documents: 20},
		{EntityType: "user", Count: 30, Documents: 30},
	}, nil
}

// generateTagCloud generates tag cloud from database
func (s *DocumentSearchService) generateTagCloud(ctx context.Context) ([]*TagCloud, error) {
	// In a real implementation, analyze tags from documents
	// For now, return mock data
	return []*TagCloud{
		{Tag: "urgent", Count: 15, Size: 5},
		{Tag: "legal", Count: 25, Size: 4},
		{Tag: "contract", Count: 20, Size: 3},
		{Tag: "signed", Count: 30, Size: 4},
		{Tag: "draft", Count: 10, Size: 2},
		{Tag: "approved", Count: 18, Size: 3},
		{Tag: "confidential", Count: 12, Size: 3},
	}, nil
}
