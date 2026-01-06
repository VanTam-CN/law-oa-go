package handlers

import (
	"strconv"
	"strings"

	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

// LegalStatuteHandler 法条处理器
type LegalStatuteHandler struct {
	legalStatuteService services.LegalStatuteService
}

// NewLegalStatuteHandler 创建法条处理器
func NewLegalStatuteHandler(legalStatuteService services.LegalStatuteService) *LegalStatuteHandler {
	return &LegalStatuteHandler{
		legalStatuteService: legalStatuteService,
	}
}

// CreateStatute 创建法条
// @Summary 创建法条
// @Description 创建新的法条
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param request body models.LegalStatuteCreateRequest true "法条创建请求"
// @Success 200 {object} models.Response{data=models.LegalStatuteResponse}
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes [post]
func (h *LegalStatuteHandler) CreateStatute(c *gin.Context) {
	var req models.LegalStatuteCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "请求参数错误", map[string]string{
			"request": err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID := middleware.GetUserID(c)

	// 创建法条
	statute, err := h.legalStatuteService.CreateStatute(c.Request.Context(), &req, userID)
	if err != nil {
		common.APIInternalServerError(c, "创建法条失败", err.Error())
		return
	}

	common.APISuccess(c, statute)
}

// GetStatuteByID 根据ID获取法条
// @Summary 获取法条详情
// @Description 根据ID获取法条详细信息
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param id path int true "法条ID"
// @Success 200 {object} models.Response{data=models.LegalStatuteResponse}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/{id} [get]
func (h *LegalStatuteHandler) GetStatuteByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.APIBadRequest(c, "无效的法条ID，法条ID必须是正整数")
		return
	}

	// 获取当前用户ID（可选）
	userID := middleware.GetUserID(c)

	// 获取法条
	statute, err := h.legalStatuteService.GetStatuteByID(c.Request.Context(), id, userID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			common.APINotFound(c, "法条不存在", err.Error())
		} else {
			common.APIInternalServerError(c, "获取法条失败", err.Error())
		}
		return
	}

	common.APISuccess(c, statute)
}

// GetStatuteByNumber 根据法条编号获取法条
// @Summary 根据编号获取法条
// @Description 根据法条编号获取法条详细信息
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param number path string true "法条编号"
// @Success 200 {object} models.Response{data=models.LegalStatuteResponse}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/number/{number} [get]
func (h *LegalStatuteHandler) GetStatuteByNumber(c *gin.Context) {
	number := c.Param("number")
	if number == "" {
		common.APIBadRequest(c, "法条编号不能为空")
		return
	}

	// 获取当前用户ID（可选）
	userID := middleware.GetUserID(c)

	// 获取法条
	statute, err := h.legalStatuteService.GetStatuteByNumber(c.Request.Context(), number, userID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			common.APINotFound(c, "法条不存在", err.Error())
		} else {
			common.APIInternalServerError(c, "获取法条失败", err.Error())
		}
		return
	}

	common.APISuccess(c, statute)
}

// UpdateStatute 更新法条
// @Summary 更新法条
// @Description 更新现有法条信息
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param id path int true "法条ID"
// @Param request body models.LegalStatuteUpdateRequest true "法条更新请求"
// @Success 200 {object} models.Response{data=models.LegalStatuteResponse}
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/{id} [put]
func (h *LegalStatuteHandler) UpdateStatute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.APIBadRequest(c, "无效的法条ID，法条ID必须是正整数")
		return
	}

	var req models.LegalStatuteUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "请求参数错误", map[string]string{
			"request": err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID := middleware.GetUserID(c)

	// 更新法条
	statute, err := h.legalStatuteService.UpdateStatute(c.Request.Context(), id, &req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			common.APINotFound(c, "法条不存在")
		} else {
			common.APIInternalServerError(c, "更新法条失败", err.Error())
		}
		return
	}

	common.APISuccess(c, statute)
}

// DeleteStatute 删除法条
// @Summary 删除法条
// @Description 删除指定的法条
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param id path int true "法条ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/{id} [delete]
func (h *LegalStatuteHandler) DeleteStatute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.APIBadRequest(c, "无效的法条ID，法条ID必须是正整数")
		return
	}

	// 获取当前用户ID
	userID := middleware.GetUserID(c)

	// 删除法条
	if err := h.legalStatuteService.DeleteStatute(c.Request.Context(), id, userID); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			common.APINotFound(c, "法条不存在")
		} else {
			common.APIInternalServerError(c, "删除法条失败", err.Error())
		}
		return
	}

	common.APISuccess(c, gin.H{"message": "法条删除成功"})
}

// ListStatutes 获取法条列表
// @Summary 获取法条列表
// @Description 分页获取法条列表
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} models.Response{data=models.LegalSearchResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes [get]
func (h *LegalStatuteHandler) ListStatutes(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 获取当前用户ID（可选）
	userID := middleware.GetUserID(c)

	// 获取法条列表
	response, err := h.legalStatuteService.ListStatutes(c.Request.Context(), page, pageSize, userID)
	if err != nil {
		common.APIInternalServerError(c, "获取法条列表失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// SearchStatutes 搜索法条
// @Summary 搜索法条
// @Description 根据条件搜索法条
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param query query string false "搜索关键词"
// @Param category_id query int false "分类ID"
// @Param law_name query string false "法律名称"
// @Param status query string false "状态"
// @Param tags query []string false "标签"
// @Param sort_by query string false "排序字段" Enums(relevance, date, title)
// @Param sort_order query string false "排序方式" Enums(asc, desc)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} models.Response{data=models.LegalSearchResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/search [get]
func (h *LegalStatuteHandler) SearchStatutes(c *gin.Context) {
	// 构建搜索请求
	req := &models.LegalSearchRequest{
		Query:     c.Query("query"),
		LawName:   c.Query("law_name"),
		Status:    c.Query("status"),
		SortBy:    c.DefaultQuery("sort_by", "relevance"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	// 解析数字参数
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.Atoi(categoryIDStr); err == nil {
			req.CategoryID = categoryID
		}
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	} else {
		req.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = pageSize
		}
	} else {
		req.PageSize = 20
	}

	// 解析标签参数
	if tagsStr := c.Query("tags"); tagsStr != "" {
		req.Tags = strings.Split(tagsStr, ",")
		// 去除空格
		for i, tag := range req.Tags {
			req.Tags[i] = strings.TrimSpace(tag)
		}
	}

	// 获取当前用户ID（可选）
	userID := middleware.GetUserID(c)

	// 执行搜索
	response, err := h.legalStatuteService.SearchStatutes(c.Request.Context(), req, userID)
	if err != nil {
		common.APIInternalServerError(c, "搜索法条失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// GetSearchSuggestions 获取搜索建议
// @Summary 获取搜索建议
// @Description 根据输入获取搜索建议
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param query query string true "搜索关键词"
// @Success 200 {object} models.Response{data=[]string}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/suggestions [get]
func (h *LegalStatuteHandler) GetSearchSuggestions(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		common.APIBadRequest(c, "搜索关键词不能为空")
		return
	}

	// 获取搜索建议
	suggestions, err := h.legalStatuteService.GetSearchSuggestions(c.Request.Context(), query)
	if err != nil {
		common.APIInternalServerError(c, "获取搜索建议失败", err.Error())
		return
	}

	common.APISuccess(c, suggestions)
}

// GetRelatedStatutes 获取相关法条
// @Summary 获取相关法条
// @Description 根据法条ID获取相关法条推荐
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param id path int true "法条ID"
// @Param limit query int false "推荐数量" default(10)
// @Success 200 {object} models.Response{data=[]models.LegalStatuteResponse}
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/{id}/related [get]
func (h *LegalStatuteHandler) GetRelatedStatutes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.APIBadRequest(c, "无效的法条ID，法条ID必须是正整数")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取当前用户ID（可选）
	userID := middleware.GetUserID(c)

	// 获取相关法条
	statutes, err := h.legalStatuteService.GetRelatedStatutes(c.Request.Context(), id, limit, userID)
	if err != nil {
		common.APIInternalServerError(c, "获取相关法条失败", err.Error())
		return
	}

	common.APISuccess(c, statutes)
}

// GetCategories 获取所有分类
// @Summary 获取法条分类
// @Description 获取所有法条分类列表
// @Tags 法条管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.LegalCategoryResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/categories [get]
func (h *LegalStatuteHandler) GetCategories(c *gin.Context) {
	categories, err := h.legalStatuteService.GetCategories(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取分类列表失败", err.Error())
		return
	}

	common.APISuccess(c, categories)
}

// GetCategoryTree 获取分类树
// @Summary 获取分类树结构
// @Description 获取法条分类的树形结构
// @Tags 法条管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.CategoryTreeNode}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/categories/tree [get]
func (h *LegalStatuteHandler) GetCategoryTree(c *gin.Context) {
	tree, err := h.legalStatuteService.GetCategoryTree(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取分类树失败", err.Error())
		return
	}

	common.APISuccess(c, tree)
}

// GetStatutesByCategory 根据分类获取法条
// @Summary 获取分类下的法条
// @Description 根据分类ID获取该分类下的法条列表
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param id path int true "分类ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} models.Response{data=models.LegalSearchResponse}
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/categories/{id}/statutes [get]
func (h *LegalStatuteHandler) GetStatutesByCategory(c *gin.Context) {
	idStr := c.Param("id")
	categoryID, err := strconv.Atoi(idStr)
	if err != nil {
		common.APIBadRequest(c, "无效的分类ID，分类ID必须是正整数")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	response, err := h.legalStatuteService.GetStatutesByCategory(c.Request.Context(), categoryID, page, pageSize)
	if err != nil {
		common.APIInternalServerError(c, "获取分类法条失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// GetTags 获取所有标签
// @Summary 获取法条标签
// @Description 获取所有法条标签列表
// @Tags 法条管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.LegalTagResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/tags [get]
func (h *LegalStatuteHandler) GetTags(c *gin.Context) {
	tags, err := h.legalStatuteService.GetTags(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取标签列表失败", err.Error())
		return
	}

	common.APISuccess(c, tags)
}

// GetPopularTags 获取热门标签
// @Summary 获取热门标签
// @Description 获取使用频率最高的标签
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param limit query int false "标签数量" default(20)
// @Success 200 {object} models.Response{data=[]models.LegalTagResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/tags/popular [get]
func (h *LegalStatuteHandler) GetPopularTags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	tags, err := h.legalStatuteService.GetPopularTags(c.Request.Context(), limit)
	if err != nil {
		common.APIInternalServerError(c, "获取热门标签失败", err.Error())
		return
	}

	common.APISuccess(c, tags)
}

// AddToFavorites 添加到收藏
// @Summary 添加法条收藏
// @Description 将法条添加到用户收藏
// @Tags 用户功能
// @Accept json
// @Produce json
// @Param request body models.FavoriteRequest true "收藏请求"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/favorites [post]
func (h *LegalStatuteHandler) AddToFavorites(c *gin.Context) {
	var req models.FavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "请求参数错误", map[string]string{
			"request": err.Error(),
		})
		return
	}

	// 获取当前用户ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.APIUnauthorized(c, "未授权，请先登录")
		return
	}

	// 添加到收藏
	if err := h.legalStatuteService.AddToFavorites(c.Request.Context(), req.StatuteID, userID); err != nil {
		common.APIInternalServerError(c, "添加收藏失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "添加收藏成功"})
}

// RemoveFromFavorites 移除收藏
// @Summary 移除法条收藏
// @Description 将法条从用户收藏中移除
// @Tags 用户功能
// @Accept json
// @Produce json
// @Param statute_id path int true "法条ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/favorites/{statute_id} [delete]
func (h *LegalStatuteHandler) RemoveFromFavorites(c *gin.Context) {
	idStr := c.Param("statute_id")
	statuteID, err := strconv.Atoi(idStr)
	if err != nil {
		common.APIBadRequest(c, "无效的法条ID，法条ID必须是正整数")
		return
	}

	// 获取当前用户ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.APIUnauthorized(c, "未授权，请先登录")
		return
	}

	// 移除收藏
	if err := h.legalStatuteService.RemoveFromFavorites(c.Request.Context(), statuteID, userID); err != nil {
		common.APIInternalServerError(c, "移除收藏失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "移除收藏成功"})
}

// GetUserFavorites 获取用户收藏
// @Summary 获取用户收藏列表
// @Description 获取当前用户的法条收藏列表
// @Tags 用户功能
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} models.Response{data=models.LegalSearchResponse}
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/favorites [get]
func (h *LegalStatuteHandler) GetUserFavorites(c *gin.Context) {
	// 获取当前用户ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.APIUnauthorized(c, "未授权，请先登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 获取用户收藏
	response, err := h.legalStatuteService.GetUserFavorites(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		common.APIInternalServerError(c, "获取收藏列表失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}

// GetSearchHistory 获取搜索历史
// @Summary 获取搜索历史
// @Description 获取当前用户的搜索历史记录
// @Tags 用户功能
// @Accept json
// @Produce json
// @Param limit query int false "记录数量" default(20)
// @Success 200 {object} models.Response{data=[]models.SearchHistoryResponse}
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/search-history [get]
func (h *LegalStatuteHandler) GetSearchHistory(c *gin.Context) {
	// 获取当前用户ID
	userID := middleware.GetUserID(c)
	if userID == 0 {
		common.APIUnauthorized(c, "未授权，请先登录")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// 获取搜索历史
	histories, err := h.legalStatuteService.GetSearchHistory(c.Request.Context(), userID, limit)
	if err != nil {
		common.APIInternalServerError(c, "获取搜索历史失败", err.Error())
		return
	}

	common.APISuccess(c, histories)
}

// GetPopularSearches 获取热门搜索
// @Summary 获取热门搜索
// @Description 获取热门搜索关键词列表
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param limit query int false "搜索数量" default(10)
// @Success 200 {object} models.Response{data=[]string}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/popular-searches [get]
func (h *LegalStatuteHandler) GetPopularSearches(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取热门搜索
	searches, err := h.legalStatuteService.GetPopularSearches(c.Request.Context(), limit)
	if err != nil {
		common.APIInternalServerError(c, "获取热门搜索失败", err.Error())
		return
	}

	common.APISuccess(c, searches)
}

// GetCategoryStats 获取分类统计
// @Summary 获取分类统计
// @Description 获取各分类下的法条数量统计
// @Tags 统计信息
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=[]models.CategoryStat}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/stats/categories [get]
func (h *LegalStatuteHandler) GetCategoryStats(c *gin.Context) {
	stats, err := h.legalStatuteService.GetCategoryStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取分类统计失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// GetPopularStatutes 获取热门法条
// @Summary 获取热门法条
// @Description 获取最受欢迎的法条列表
// @Tags 统计信息
// @Accept json
// @Produce json
// @Param limit query int false "法条数量" default(10)
// @Success 200 {object} models.Response{data=[]models.LegalStatuteResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/stats/popular [get]
func (h *LegalStatuteHandler) GetPopularStatutes(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取当前用户ID（可选）
	userID := middleware.GetUserID(c)

	// 获取热门法条
	statutes, err := h.legalStatuteService.GetPopularStatutes(c.Request.Context(), limit, userID)
	if err != nil {
		common.APIInternalServerError(c, "获取热门法条失败", err.Error())
		return
	}

	common.APISuccess(c, statutes)
}

// GetRecentUpdates 获取最近更新
// @Summary 获取最近更新
// @Description 获取最近更新的法条列表
// @Tags 统计信息
// @Accept json
// @Produce json
// @Param days query int false "天数" default(7)
// @Success 200 {object} models.Response{data=[]models.LegalStatuteResponse}
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/stats/recent [get]
func (h *LegalStatuteHandler) GetRecentUpdates(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	// 获取最近更新
	statutes, err := h.legalStatuteService.GetRecentUpdates(c.Request.Context(), days)
	if err != nil {
		common.APIInternalServerError(c, "获取最近更新失败", err.Error())
		return
	}

	common.APISuccess(c, statutes)
}

// SyncToElasticsearch 同步到Elasticsearch
// @Summary 同步数据到Elasticsearch
// @Description 将法条数据同步到Elasticsearch搜索引擎
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/admin/sync-elasticsearch [post]
func (h *LegalStatuteHandler) SyncToElasticsearch(c *gin.Context) {
	if err := h.legalStatuteService.SyncToElasticsearch(c.Request.Context()); err != nil {
		common.APIInternalServerError(c, "同步失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "数据同步成功"})
}

// RebuildSearchIndex 重建搜索索引
// @Summary 重建搜索索引
// @Description 重建Elasticsearch搜索索引
// @Tags 系统管理
// @Accept json
// @Produce json
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/admin/rebuild-index [post]
func (h *LegalStatuteHandler) RebuildSearchIndex(c *gin.Context) {
	if err := h.legalStatuteService.RebuildSearchIndex(c.Request.Context()); err != nil {
		common.APIInternalServerError(c, "重建索引失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "索引重建成功"})
}

// BulkImportStatutes 批量导入法条
// @Summary 批量导入法条
// @Description 通过JSON数据批量导入法条到数据库
// @Tags 法条管理
// @Accept json
// @Produce json
// @Param request body models.LegalStatuteImportRequest true "法条导入请求"
// @Success 200 {object} models.Response{data=models.LegalStatuteImportResponse}
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/v1/legal/statutes/import [post]
func (h *LegalStatuteHandler) BulkImportStatutes(c *gin.Context) {
	var req models.LegalStatuteImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "请求参数错误", map[string]string{
			"request": err.Error(),
		})
		return
	}

	// 验证请求
	if len(req.Statutes) == 0 {
		common.APIBadRequest(c, "导入数据不能为空")
		return
	}

	// 限制单次导入数量
	if len(req.Statutes) > 1000 {
		common.APIBadRequest(c, "单次最多导入1000条法条")
		return
	}

	// 获取当前用户ID（可选）
	userID := middleware.GetUserID(c)

	// 执行导入
	response, err := h.legalStatuteService.BulkImportStatutes(c.Request.Context(), &req, userID)
	if err != nil {
		common.APIInternalServerError(c, "导入法条失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}