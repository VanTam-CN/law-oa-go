package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/services"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search godoc
// @Summary 全局搜索
// @Description 在所有实体中搜索指定关键词
// @Tags 搜索
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param types query string false "实体类型过滤（逗号分隔）"
// @Param entity_id query int false "实体ID过滤"
// @Param date_from query string false "起始日期 (YYYY-MM-DD)"
// @Param date_to query string false "结束日期 (YYYY-MM-DD)"
// @Param sort_by query string false "排序字段 (score, date, relevance)" default(score)
// @Param sort_order query string false "排序顺序 (asc, desc)" default(desc)
// @Success 200 {object} common.APIResponse{data=services.SearchResponse} "搜索成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		_ = c.Error(errors.ValidationError("query", "Search query is required"))
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Parse types
	typesStr := c.Query("types")
	var types []string
	if typesStr != "" {
		// Split comma-separated types
		types = []string{}
		for _, t := range []string{typesStr} {
			for _, tt := range []string{t} {
				types = append(types, tt)
			}
		}
	}

	// Parse entity ID
	entityIDStr := c.Query("entity_id")
	var entityID *uint
	if entityIDStr != "" {
		id, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err != nil {
			_ = c.Error(errors.ValidationError("entity_id", "Invalid entity ID: must be a valid number"))
			return
		}
		eid := uint(id)
		entityID = &eid
	}

	// Parse dates
	dateFrom := c.Query("date_from")
	var dateFromPtr *string
	if dateFrom != "" {
		dateFromPtr = &dateFrom
	}

	dateTo := c.Query("date_to")
	var dateToPtr *string
	if dateTo != "" {
		dateToPtr = &dateTo
	}

	// Parse sorting
	sortBy := c.DefaultQuery("sort_by", "score")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	req := &services.SearchRequest{
		Query:     query,
		Page:      page,
		PageSize:  pageSize,
		Types:     types,
		EntityID:  entityID,
		DateFrom:  dateFromPtr,
		DateTo:    dateToPtr,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	response, err := h.searchService.Search(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, response)
}

// GetSearchSuggestions godoc
// @Summary 获取搜索建议
// @Description 根据输入获取搜索建议
// @Tags 搜索
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "搜索关键词"
// @Param limit query int false "建议数量限制" default(10)
// @Success 200 {object} common.APIResponse{data=[]string} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /search/suggestions [get]
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		_ = c.Error(errors.ValidationError("query", "Search query is required"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	suggestions, err := h.searchService.GetSearchSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, suggestions)
}

// ReindexAll godoc
// @Summary 重建搜索索引
// @Description 重新构建所有实体的搜索索引
// @Tags 搜索
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse "重建成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /search/reindex [post]
func (h *SearchHandler) ReindexAll(c *gin.Context) {
	err := h.searchService.ReindexAll(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, gin.H{"message": "Search index rebuilt successfully"})
}
