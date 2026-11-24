package handlers

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// ContentFilterHandler 内容过滤处理器
type ContentFilterHandler struct {
	db            *gorm.DB
	filterService *services.ContentFilterService
}

// NewContentFilterHandler 创建内容过滤处理器
func NewContentFilterHandler(db *gorm.DB) *ContentFilterHandler {
	return &ContentFilterHandler{
		db:            db,
		filterService: services.NewContentFilterService(db),
	}
}

// =============================================================================
// 敏感词管理 API
// =============================================================================

// CreateSensitiveWordRequest 创建敏感词请求
type CreateSensitiveWordRequest struct {
	Word        string `json:"word" binding:"required"`
	WordType    string `json:"word_type" binding:"required"`
	Category    string `json:"category"`
	Severity    string `json:"severity" binding:"required,oneof=low medium high critical"`
	Replacement string `json:"replacement"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
}

// CreateSensitiveWord 创建敏感词
func (h *ContentFilterHandler) CreateSensitiveWord(c *gin.Context) {
	var req CreateSensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权")
		return
	}

	word := &models.SensitiveWord{
		Word:        req.Word,
		WordType:    req.WordType,
		Category:    req.Category,
		Severity:    req.Severity,
		Replacement: req.Replacement,
		IsActive:    req.IsActive,
		Description: req.Description,
		CreatedBy:   userID.(uint),
		UpdatedBy:   userID.(uint),
	}

	if err := h.filterService.AddSensitiveWord(context.Background(), word); err != nil {
		common.APIInternalServerError(c, "创建敏感词失败: "+err.Error())
		return
	}

	common.APISuccess(c, word)
}

// GetSensitiveWords 获取敏感词列表
func (h *ContentFilterHandler) GetSensitiveWords(c *gin.Context) {
	wordType := c.Query("word_type")
	category := c.Query("category")
	onlyActive := c.DefaultQuery("only_active", "true") == "true"

	words, err := h.filterService.GetSensitiveWords(
		context.Background(),
		wordType,
		category,
		onlyActive,
	)
	if err != nil {
		common.APIInternalServerError(c, "获取敏感词列表失败: "+err.Error())
		return
	}

	common.APISuccess(c, words)
}

// GetSensitiveWordByID 获取敏感词详情
func (h *ContentFilterHandler) GetSensitiveWordByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的敏感词ID")
		return
	}

	var word models.SensitiveWord
	if err := h.db.First(&word, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "敏感词不存在")
			return
		}
		common.APIInternalServerError(c, "获取敏感词失败: "+err.Error())
		return
	}

	common.APISuccess(c, word)
}

// UpdateSensitiveWordRequest 更新敏感词请求
type UpdateSensitiveWordRequest struct {
	Word        string `json:"word"`
	WordType    string `json:"word_type"`
	Category    string `json:"category"`
	Severity    string `json:"severity" binding:"omitempty,oneof=low medium high critical"`
	Replacement string `json:"replacement"`
	IsActive    *bool  `json:"is_active"`
	Description string `json:"description"`
}

// UpdateSensitiveWord 更新敏感词
func (h *ContentFilterHandler) UpdateSensitiveWord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的敏感词ID")
		return
	}

	var word models.SensitiveWord
	if err := h.db.First(&word, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "敏感词不存在")
			return
		}
		common.APIInternalServerError(c, "获取敏感词失败: "+err.Error())
		return
	}

	var req UpdateSensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if exists {
		word.UpdatedBy = userID.(uint)
	}

	if req.Word != "" {
		word.Word = req.Word
	}
	if req.WordType != "" {
		word.WordType = req.WordType
	}
	if req.Category != "" {
		word.Category = req.Category
	}
	if req.Severity != "" {
		word.Severity = req.Severity
	}
	if req.Replacement != "" {
		word.Replacement = req.Replacement
	}
	if req.IsActive != nil {
		word.IsActive = *req.IsActive
	}
	if req.Description != "" {
		word.Description = req.Description
	}

	if err := h.filterService.UpdateSensitiveWord(context.Background(), &word); err != nil {
		common.APIInternalServerError(c, "更新敏感词失败: "+err.Error())
		return
	}

	common.APISuccess(c, word)
}

// DeleteSensitiveWord 删除敏感词
func (h *ContentFilterHandler) DeleteSensitiveWord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的敏感词ID")
		return
	}

	if err := h.filterService.DeleteSensitiveWord(context.Background(), uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "敏感词不存在")
			return
		}
		common.APIInternalServerError(c, "删除敏感词失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "删除成功"})
}

// =============================================================================
// 内容过滤 API
// =============================================================================

// FilterContentRequest 过滤内容请求
type FilterContentRequest struct {
	Content     string `json:"content" binding:"required"`
	ContentType string `json:"content_type"`
}

// FilterContent 过滤内容
func (h *ContentFilterHandler) FilterContent(c *gin.Context) {
	var req FilterContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	contentType := req.ContentType
	if contentType == "" {
		contentType = "text"
	}

	result, err := h.filterService.FilterContent(context.Background(), req.Content, contentType)
	if err != nil {
		common.APIInternalServerError(c, "过滤内容失败: "+err.Error())
		return
	}

	common.APISuccess(c, result)
}

// CheckContentRequest 检查内容请求
type CheckContentRequest struct {
	Content string `json:"content" binding:"required"`
}

// CheckContent 检查内容（不修改）
func (h *ContentFilterHandler) CheckContent(c *gin.Context) {
	var req CheckContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	hasSensitive, foundWords, err := h.filterService.CheckContent(context.Background(), req.Content)
	if err != nil {
		common.APIInternalServerError(c, "检查内容失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"has_sensitive_words": hasSensitive,
		"found_words":         foundWords,
		"word_count":          len(foundWords),
	})
}

// =============================================================================
// 批量操作 API
// =============================================================================

// BatchImportWordsRequest 批量导入请求
type BatchImportWordsRequest struct {
	Words    []string `json:"words" binding:"required"`
	WordType string   `json:"word_type" binding:"required"`
	Severity string   `json:"severity" binding:"required,oneof=low medium high critical"`
}

// BatchImportWords 批量导入敏感词
func (h *ContentFilterHandler) BatchImportWords(c *gin.Context) {
	var req BatchImportWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权")
		return
	}

	successCount, errors, err := h.filterService.BatchImportWords(
		context.Background(),
		req.Words,
		req.WordType,
		req.Severity,
	)
	if err != nil {
		common.APIInternalServerError(c, "批量导入失败: "+err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message":       "批量导入完成",
		"success_count": successCount,
		"error_count":   len(errors),
		"errors":        errors,
		"user_id":       userID,
	})
}

// BatchToggleWordsRequest 批量切换状态请求
type BatchToggleWordsRequest struct {
	IDs      []uint `json:"ids" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// BatchToggleWords 批量切换敏感词状态
func (h *ContentFilterHandler) BatchToggleWords(c *gin.Context) {
	var req BatchToggleWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权")
		return
	}

	now := time.Now()
	result := h.db.Model(&models.SensitiveWord{}).
		Where("id IN ?", req.IDs).
		Updates(map[string]interface{}{
			"is_active":  req.IsActive,
			"updated_by": userID,
			"updated_at": now,
		})

	if result.Error != nil {
		common.APIInternalServerError(c, "批量更新失败: "+result.Error.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message":       "批量更新成功",
		"affected_rows": result.RowsAffected,
	})
}

// BatchDeleteWordsRequest 批量删除请求
type BatchDeleteWordsRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// BatchDeleteWords 批量删除敏感词
func (h *ContentFilterHandler) BatchDeleteWords(c *gin.Context) {
	var req BatchDeleteWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, err.Error())
		return
	}

	result := h.db.Delete(&models.SensitiveWord{}, req.IDs)
	if result.Error != nil {
		common.APIInternalServerError(c, "批量删除失败: "+result.Error.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"message":       "批量删除成功",
		"affected_rows": result.RowsAffected,
	})
}

// =============================================================================
// 统计 API
// =============================================================================

// GetSensitiveWordStats 获取敏感词统计
func (h *ContentFilterHandler) GetSensitiveWordStats(c *gin.Context) {
	stats, err := h.filterService.GetSensitiveWordStats(context.Background())
	if err != nil {
		common.APIInternalServerError(c, "获取统计失败: "+err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// GetFilterLogs 获取过滤日志
func (h *ContentFilterHandler) GetFilterLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	contentType := c.Query("content_type")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	query := h.db.Model(&models.ContentFilterLog{})

	if contentType != "" {
		query = query.Where("content_type = ?", contentType)
	}
	if dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
	}

	var total int64
	query.Count(&total)

	var logs []models.ContentFilterLog
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs)

	common.APISuccess(c, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// ResetCache 重置缓存
func (h *ContentFilterHandler) ResetCache(c *gin.Context) {
	// 通过重新初始化服务来重置缓存
	h.filterService = services.NewContentFilterService(h.db)

	common.APISuccess(c, gin.H{"message": "缓存已重置"})
}

// =============================================================================
// 辅助函数
// =============================================================================

// normalizeWord 标准化敏感词（去除首尾空格，统一大小写）
func normalizeWord(word string) string {
	return strings.TrimSpace(strings.ToLower(word))
}
