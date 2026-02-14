package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// SensitiveWordFilterService 敏感词过滤服务
type SensitiveWordFilterService struct {
	wordRepo repositories.SensitiveWordRepository
}

// NewSensitiveWordFilterService 创建敏感词过滤服务实例
func NewSensitiveWordFilterService(
	wordRepo repositories.SensitiveWordRepository,
) *SensitiveWordFilterService {
	return &SensitiveWordFilterService{
		wordRepo: wordRepo,
	}
}

// CreateWordRequest 创建敏感词请求
type CreateWordRequest struct {
	Word        string `json:"word" binding:"required,min=1,max=200"`
	WordType    string `json:"word_type" binding:"required,oneof=political porn violence other"`
	Category    string `json:"category" binding:"required,max=50"`
	Severity    string `json:"severity" binding:"required,oneof=low medium high critical"`
	Replacement string `json:"replacement" binding:"max=200"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	CreatedBy   uint   `json:"created_by" binding:"required"`
}

// UpdateWordRequest 更新敏感词请求
type UpdateWordRequest struct {
	Word        *string `json:"word,omitempty" binding:"omitempty,min=1,max=200"`
	WordType    *string `json:"word_type,omitempty" binding:"omitempty,oneof=political porn violence other"`
	Category    *string `json:"category,omitempty" binding:"omitempty,max=50"`
	Severity    *string `json:"severity,omitempty" binding:"omitempty,oneof=low medium high critical"`
	Replacement *string `json:"replacement,omitempty" binding:"omitempty,max=200"`
	IsActive    *bool   `json:"is_active,omitempty"`
	Description *string `json:"description,omitempty"`
}

// WordResponse 敏感词响应
type WordResponse struct {
	ID          uint      `json:"id"`
	Word        string    `json:"word"`
	WordType    string    `json:"word_type"`
	Category    string    `json:"category"`
	Severity    string    `json:"severity"`
	Replacement string    `json:"replacement"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	HitCount    int       `json:"hit_count"`
	LastHitAt   *string   `json:"last_hit_at"`
	CreatedBy   uint      `json:"created_by"`
	UpdatedBy   uint      `json:"updated_by"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// ListWordsRequest 敏感词列表请求
type ListWordsRequest struct {
	Page     int    `json:"page" form:"page" binding:"min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	WordType string `json:"word_type" form:"word_type" binding:"omitempty,oneof=political porn violence other"`
	Category string `json:"category" form:"category" binding:"omitempty,max=50"`
	Severity string `json:"severity" form:"severity" binding:"omitempty,oneof=low medium high critical"`
	Search   string `json:"search" form:"search"`
	IsActive *bool  `json:"is_active" form:"is_active"`
}

// WordPagination 分页信息
type WordPagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

// ListWordsResponse 敏感词列表响应
type ListWordsResponse struct {
	Words     []*WordResponse `json:"words"`
	Pagination WordPagination  `json:"pagination"`
}

// CheckTextRequest 文本检测请求
type CheckTextRequest struct {
	Text     string `json:"text" binding:"required"`
	WordType string `json:"word_type" form:"word_type" binding:"omitempty,oneof=political porn violence other"`
}

// CheckTextResponse 文本检测结果
type CheckTextResponse struct {
	OriginalText    string          `json:"original_text"`
	FilteredText    string          `json:"filtered_text"`
	ContainsSensitive bool         `json:"contains_sensitive"`
	SensitiveWords  []*SensitiveWordFound `json:"sensitive_words"`
	SensitiveCount  int             `json:"sensitive_count"`
}

// SensitiveWordFound 敏感词发现
type SensitiveWordFound struct {
	Word       string `json:"word"`
	WordType   string `json:"word_type"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	Position   int    `json:"position"`
	Length     int    `json:"length"`
	Replaced   bool   `json:"replaced"`
	Replacement string `json:"replacement,omitempty"`
}

// CreateWord 创建敏感词
func (s *SensitiveWordFilterService) CreateWord(ctx context.Context, req *CreateWordRequest) (*WordResponse, error) {
	word := &models.SensitiveWord{
		Word:        req.Word,
		WordType:    req.WordType,
		Category:    req.Category,
		Severity:    req.Severity,
		Replacement: req.Replacement,
		IsActive:    req.IsActive,
		Description: req.Description,
		CreatedBy:   req.CreatedBy,
		UpdatedBy:   req.CreatedBy,
		HitCount:    0,
	}

	if err := s.wordRepo.Create(ctx, word); err != nil {
		return nil, fmt.Errorf("创建敏感词失败: %w", err)
	}

	return s.GetWordByID(ctx, word.ID)
}

// GetWordByID 根据ID获取敏感词详情
func (s *SensitiveWordFilterService) GetWordByID(ctx context.Context, id uint) (*WordResponse, error) {
	word, err := s.wordRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询敏感词失败: %w", err)
	}
	if word == nil {
		return nil, errors.New("敏感词不存在")
	}

	return s.convertToResponse(word), nil
}

// ListWords 获取敏感词列表
func (s *SensitiveWordFilterService) ListWords(ctx context.Context, req *ListWordsRequest) (*ListWordsResponse, error) {
	params := &repositories.SensitiveWordListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		WordType: req.WordType,
		Category: req.Category,
		Severity: req.Severity,
		Search:   req.Search,
		IsActive: req.IsActive,
	}

	words, total, err := s.wordRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询敏感词列表失败: %w", err)
	}

	response := &ListWordsResponse{
		Words: make([]*WordResponse, len(words)),
		Pagination: WordPagination{
			Page:    req.Page,
			PageSize: req.PageSize,
			Total:   total,
		},
	}

	for i, w := range words {
		response.Words[i] = s.convertToResponse(w)
	}

	return response, nil
}

// UpdateWord 更新敏感词
func (s *SensitiveWordFilterService) UpdateWord(ctx context.Context, id uint, req *UpdateWordRequest) (*WordResponse, error) {
	word, err := s.wordRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询敏感词失败: %w", err)
	}
	if word == nil {
		return nil, errors.New("敏感词不存在")
	}

	// 更新字段
	if req.Word != nil {
		word.Word = *req.Word
	}
	if req.WordType != nil {
		word.WordType = *req.WordType
	}
	if req.Category != nil {
		word.Category = *req.Category
	}
	if req.Severity != nil {
		word.Severity = *req.Severity
	}
	if req.Replacement != nil {
		word.Replacement = *req.Replacement
	}
	if req.IsActive != nil {
		word.IsActive = *req.IsActive
	}
	if req.Description != nil {
		word.Description = *req.Description
	}

	if err := s.wordRepo.Update(ctx, word); err != nil {
		return nil, fmt.Errorf("更新敏感词失败: %w", err)
	}

	return s.GetWordByID(ctx, id)
}

// DeleteWord 删除敏感词
func (s *SensitiveWordFilterService) DeleteWord(ctx context.Context, id uint) error {
	word, err := s.wordRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询敏感词失败: %w", err)
	}
	if word == nil {
		return errors.New("敏感词不存在")
	}

	if err := s.wordRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除敏感词失败: %w", err)
	}

	return nil
}

// CheckText 检测文本中的敏感词
func (s *SensitiveWordFilterService) CheckText(ctx context.Context, req *CheckTextRequest) (*CheckTextResponse, error) {
	// 获取所有启用的敏感词
	params := &repositories.SensitiveWordListParams{
		Page:     1,
		PageSize: 10000,
		IsActive: boolPtrVal(true),
	}

	if req.WordType != "" {
		params.WordType = req.WordType
	}

	words, _, err := s.wordRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询敏感词列表失败: %w", err)
	}

	foundWords := make([]*SensitiveWordFound, 0)
	filteredText := req.Text
	offset := 0 // 用于跟踪替换操作后文本位置的变化

	// 记录已匹配的位置，避免重复
	matchedRanges := make(map[[2]int]bool)

	for _, word := range words {
		if !word.IsActive {
			continue
		}

		// 普通字符串匹配
		index := 0
		for {
			pos := strings.Index(req.Text[index:], word.Word)
			if pos == -1 {
				break
			}
			actualPos := index + pos

			// 检查是否已经匹配过这个范围
			rangeKey := [2]int{actualPos, actualPos + len(word.Word)}
			if matchedRanges[rangeKey] {
				index = actualPos + len(word.Word)
				continue
			}
			matchedRanges[rangeKey] = true

			// 获取替换词
			replacement := word.Replacement
			if replacement == "" {
				replacement = s.getReplacement(word.Word)
			}

			foundWords = append(foundWords, &SensitiveWordFound{
				Word:       word.Word,
				WordType:   word.WordType,
				Category:   word.Category,
				Severity:   word.Severity,
				Position:   actualPos,
				Length:     len(word.Word),
				Replaced:   true,
				Replacement: replacement,
			})

			// 替换敏感词 (使用原始文本位置)
			filteredText = filteredText[:actualPos+offset] + replacement + filteredText[actualPos+offset+len(word.Word):]
			offset += len(replacement) - len(word.Word)

			index = actualPos + len(word.Word)
		}

		// 更新命中统计
		if len(foundWords) > 0 {
			s.updateHitCount(ctx, word.ID)
		}
	}

	return &CheckTextResponse{
		OriginalText:      req.Text,
		FilteredText:      filteredText,
		ContainsSensitive: len(foundWords) > 0,
		SensitiveWords:    foundWords,
		SensitiveCount:    len(foundWords),
	}, nil
}

// getReplacement 获取替换文本
func (s *SensitiveWordFilterService) getReplacement(word string) string {
	// 默认替换为星号
	runes := []rune(word)
	for i := range runes {
		runes[i] = '*'
	}
	return string(runes)
}

// MaskText 脱敏文本（直接替换敏感词为星号）
func (s *SensitiveWordFilterService) MaskText(ctx context.Context, text string) (string, error) {
	req := &CheckTextRequest{Text: text}
	result, err := s.CheckText(ctx, req)
	if err != nil {
		return "", err
	}
	return result.FilteredText, nil
}

// GetCategories 获取敏感词分类列表
func (s *SensitiveWordFilterService) GetCategories(ctx context.Context) ([]*CategoryInfo, error) {
	categories := []*CategoryInfo{
		{Code: "political", Name: "政治敏感", Description: "涉及政治的敏感词汇"},
		{Code: "porn", Name: "色情词汇", Description: "涉及色情的敏感词汇"},
		{Code: "violence", Name: "暴力词汇", Description: "涉及暴力的敏感词汇"},
		{Code: "other", Name: "其他", Description: "其他敏感词汇"},
	}
	return categories, nil
}

// updateHitCount 更新命中统计
func (s *SensitiveWordFilterService) updateHitCount(ctx context.Context, wordID uint) {
	now := time.Now()
	s.wordRepo.UpdateHitCount(ctx, wordID, now)
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// BatchImportWords 批量导入敏感词
func (s *SensitiveWordFilterService) BatchImportWords(ctx context.Context, words []*CreateWordRequest) (*ImportResult, error) {
	result := &ImportResult{
		Total:     len(words),
		Success:   0,
		Failed:    0,
		Errors:    make([]string, 0),
	}

	for _, wordReq := range words {
		_, err := s.CreateWord(ctx, wordReq)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", wordReq.Word, err.Error()))
		} else {
			result.Success++
		}
	}

	return result, nil
}

// ImportResult 导入结果
type ImportResult struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}

// convertToResponse 转换为响应格式
func (s *SensitiveWordFilterService) convertToResponse(word *models.SensitiveWord) *WordResponse {
	var lastHitAtStr *string
	if word.LastHitAt != nil {
		str := word.LastHitAt.Format("2006-01-02 15:04:05")
		lastHitAtStr = &str
	}

	return &WordResponse{
		ID:          word.ID,
		Word:        word.Word,
		WordType:    word.WordType,
		Category:    word.Category,
		Severity:    word.Severity,
		Replacement: word.Replacement,
		IsActive:    word.IsActive,
		Description: word.Description,
		HitCount:    word.HitCount,
		LastHitAt:   lastHitAtStr,
		CreatedBy:   word.CreatedBy,
		UpdatedBy:   word.UpdatedBy,
		CreatedAt:   word.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   word.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// boolPtrVal 返回布尔值指针
func boolPtrVal(b bool) *bool {
	return &b
}
