package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ContentFilterResult 内容过滤结果
type ContentFilterResult struct {
	OriginalContent   string                 `json:"original_content"`
	FilteredContent    string                 `json:"filtered_content"`
	HasSensitiveWords bool                   `json:"has_sensitive_words"`
	Hits              []SensitiveWordHit      `json:"hits"`
	IsBlocked         bool                   `json:"is_blocked"`
	RequiresApproval  bool                   `json:"requires_approval"`
}

// SensitiveWordHit 敏感词命中记录
type SensitiveWordHit struct {
	Word        string    `json:"word"`
	WordType    string    `json:"word_type"`
	Severity    string    `json:"severity"`
	Position    int       `json:"position"`
	Replacement string    `json:"replacement"`
}

// SensitiveWordRepository 敏感词仓库接口
type SensitiveWordRepository interface {
	GetActiveWords(ctx context.Context) ([]*models.SensitiveWord, error)
	Create(ctx context.Context, word *models.SensitiveWord) error
	Update(ctx context.Context, word *models.SensitiveWord) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*models.SensitiveWord, error)
	IncrementHitCount(ctx context.Context, id uint) error
}

// SensitiveWordRepositoryImpl 敏感词仓库实现
type SensitiveWordRepositoryImpl struct {
	db *gorm.DB
}

// NewSensitiveWordRepository 创建敏感词仓库
func NewSensitiveWordRepository(db *gorm.DB) SensitiveWordRepository {
	return &SensitiveWordRepositoryImpl{db: db}
}

func (r *SensitiveWordRepositoryImpl) GetActiveWords(ctx context.Context) ([]*models.SensitiveWord, error) {
	var words []models.SensitiveWord
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&words).Error
	if err != nil {
		return nil, err
	}
	result := make([]*models.SensitiveWord, len(words))
	for i, w := range words {
		result[i] = &w
	}
	return result, nil
}

func (r *SensitiveWordRepositoryImpl) Create(ctx context.Context, word *models.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *SensitiveWordRepositoryImpl) Update(ctx context.Context, word *models.SensitiveWord) error {
	return r.db.WithContext(ctx).Save(word).Error
}

func (r *SensitiveWordRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.SensitiveWord{}, id).Error
}

func (r *SensitiveWordRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.SensitiveWord, error) {
	var word models.SensitiveWord
	err := r.db.WithContext(ctx).First(&word, id).Error
	if err != nil {
		return nil, err
	}
	return &word, nil
}

func (r *SensitiveWordRepositoryImpl) IncrementHitCount(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SensitiveWord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"hit_count":   gorm.Expr("hit_count + 1"),
			"last_hit_at": now,
		}).Error
}

// ContentFilterService 内容过滤服务
type ContentFilterService struct {
	wordRepo     SensitiveWordRepository
	db           *gorm.DB
	initialized  bool
	wordsCache   []*models.SensitiveWord
	regexCache   map[string]*regexp.Regexp
	cacheMutex   sync.RWMutex
	lastCacheUpdate time.Time
}

// NewContentFilterService 创建内容过滤服务
func NewContentFilterService(db *gorm.DB) *ContentFilterService {
	return &ContentFilterService{
		wordRepo:   NewSensitiveWordRepository(db),
		db:         db,
		regexCache: make(map[string]*regexp.Regexp),
	}
}

// refreshCache 刷新敏感词缓存
func (s *ContentFilterService) refreshCache(ctx context.Context) error {
	words, err := s.wordRepo.GetActiveWords(ctx)
	if err != nil {
		return err
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.wordsCache = words
	s.regexCache = make(map[string]*regexp.Regexp)
	s.lastCacheUpdate = time.Now()
	s.initialized = true

	return nil
}

// ensureCache 确保缓存已初始化
func (s *ContentFilterService) ensureCache(ctx context.Context) error {
	s.cacheMutex.RLock()
	if s.initialized && time.Since(s.lastCacheUpdate) < 5*time.Minute {
		s.cacheMutex.RUnlock()
		return nil
	}
	s.cacheMutex.RUnlock()

	return s.refreshCache(ctx)
}

// FilterContent 过滤内容
func (s *ContentFilterService) FilterContent(ctx context.Context, content string, contentType string) (*ContentFilterResult, error) {
	if err := s.ensureCache(ctx); err != nil {
		return nil, err
	}

	result := &ContentFilterResult{
		OriginalContent: content,
		FilteredContent:  content,
		Hits:            []SensitiveWordHit{},
	}

	s.cacheMutex.RLock()
	words := s.wordsCache
	s.cacheMutex.RUnlock()

	contentLower := strings.ToLower(content)
	highestSeverity := ""

	for _, word := range words {
		// 构建或获取正则表达式
		var re *regexp.Regexp
		var exists bool
		
		s.regexCacheMu(func() {
			re, exists = s.regexCache[word.Word]
			if !exists {
				// 转义特殊字符，创建正则表达式
				pattern := regexp.QuoteMeta(word.Word)
				re = regexp.MustCompile(pattern)
				s.regexCache[word.Word] = re
			}
		})

		// 查找所有匹配位置
		matches := re.FindAllStringIndex(content, -1)
		if len(matches) == 0 {
			continue
		}

		// 检查小写版本也匹配
		lowerRe := regexp.MustCompile(regexp.QuoteMeta(strings.ToLower(word.Word)))
		lowerMatches := lowerRe.FindAllStringIndex(contentLower, -1)
		if len(lowerMatches) > 0 {
			matches = append(matches, lowerMatches...)
		}

		result.HasSensitiveWords = true

		// 更新严重程度
		if s.compareSeverity(word.Severity, highestSeverity) > 0 {
			highestSeverity = word.Severity
		}

		// 记录命中并更新统计
		for _, match := range matches {
			hit := SensitiveWordHit{
				Word:        word.Word,
				WordType:    word.WordType,
				Severity:    word.Severity,
				Position:    match[0],
				Replacement: word.Replacement,
			}
			result.Hits = append(result.Hits, hit)
		}

		// 异步更新命中次数
		go func(wordID uint) {
			s.wordRepo.IncrementHitCount(context.Background(), wordID)
		}(word.ID)
	}

	// 根据严重程度决定处理方式
	if result.HasSensitiveWords {
		switch highestSeverity {
		case "critical":
			result.IsBlocked = true
			result.RequiresApproval = true
		case "high":
			result.RequiresApproval = true
			// 替换敏感词
			result.FilteredContent = s.replaceSensitiveWords(content, result.Hits)
		case "medium":
			// 替换敏感词
			result.FilteredContent = s.replaceSensitiveWords(content, result.Hits)
		case "low":
			// 仅记录，不替换
		}
	}

	// 记录过滤日志
	go s.logFilter(context.Background(), contentType, 0, result)

	return result, nil
}

// regexCacheMu 线程安全地访问正则缓存
func (s *ContentFilterService) regexCacheMu(fn func()) {
	var mu sync.RWMutex
	if s.cacheMutex != (sync.RWMutex{}) {
		mu = s.cacheMutex
	}
	mu.Lock()
	defer mu.Unlock()
	fn()
}

// replaceSensitiveWords 替换敏感词
func (s *ContentFilterService) replaceSensitiveWords(content string, hits []SensitiveWordHit) string {
	result := content
	// 按位置倒序处理，避免位置偏移
	for i := len(hits) - 1; i >= 0; i-- {
		hit := hits[i]
		if hit.Replacement != "" {
			result = result[:hit.Position] + hit.Replacement + result[hit.Position+len(hit.Word):]
		} else {
			// 默认替换为星号
			result = result[:hit.Position] + strings.Repeat("*", len(hit.Word)) + result[hit.Position+len(hit.Word):]
		}
	}
	return result
}

// compareSeverity 比较严重程度
func (s *ContentFilterService) compareSeverity(a, b string) int {
	severityOrder := map[string]int{
		"":        0,
		"low":     1,
		"medium":  2,
		"high":    3,
		"critical": 4,
	}
	aLevel := severityOrder[a]
	bLevel := severityOrder[b]
	if aLevel > bLevel {
		return 1
	} else if aLevel < bLevel {
		return -1
	}
	return 0
}

// logFilter 记录过滤日志
func (s *ContentFilterService) logFilter(ctx context.Context, contentType string, contentID uint, result *ContentFilterResult) {
	log := &models.ContentFilterLog{
		ContentType:       contentType,
		ContentID:         contentID,
		OriginalContent:   result.OriginalContent,
		FilteredContent:   result.FilteredContent,
		IsBlocked:         result.IsBlocked,
		RequiresApproval:  result.RequiresApproval,
	}

	// 构建命中详情JSON
	hitsData := make([]map[string]interface{}, len(result.Hits))
	for i, hit := range result.Hits {
		hitsData[i] = map[string]interface{}{
			"word":        hit.Word,
			"word_type":   hit.WordType,
			"severity":    hit.Severity,
			"position":    hit.Position,
			"replacement": hit.Replacement,
		}
	}
	// 将 hitsData 包装为 map 并转换为 models.JSON
	hitsMap := map[string]interface{}{"hits": hitsData}
	hitsJSONBytes, _ := json.Marshal(hitsMap)
	var hitsParsed map[string]interface{}
	json.Unmarshal(hitsJSONBytes, &hitsParsed)
	log.Hits = models.JSON(hitsParsed)

	// 确定采取的行动
	if result.IsBlocked {
		log.ActionTaken = "blocked"
	} else if result.RequiresApproval {
		log.ActionTaken = "approved"
	} else if len(result.Hits) > 0 {
		log.ActionTaken = "replaced"
	} else {
		log.ActionTaken = "none"
	}

	s.db.WithContext(ctx).Create(log)
}

// CheckContent 检查内容是否包含敏感词（不修改）
func (s *ContentFilterService) CheckContent(ctx context.Context, content string) (bool, []string, error) {
	if err := s.ensureCache(ctx); err != nil {
		return false, nil, err
	}

	s.cacheMutex.RLock()
	words := s.wordsCache
	s.cacheMutex.RUnlock()

	found := false
	foundWords := []string{}

	for _, word := range words {
		if strings.Contains(content, word.Word) {
			found = true
			foundWords = append(foundWords, word.Word)
		}
	}

	return found, foundWords, nil
}

// AddSensitiveWord 添加敏感词
func (s *ContentFilterService) AddSensitiveWord(ctx context.Context, word *models.SensitiveWord) error {
	if err := s.wordRepo.Create(ctx, word); err != nil {
		return err
	}
	// 刷新缓存
	return s.refreshCache(ctx)
}

// UpdateSensitiveWord 更新敏感词
func (s *ContentFilterService) UpdateSensitiveWord(ctx context.Context, word *models.SensitiveWord) error {
	if err := s.wordRepo.Update(ctx, word); err != nil {
		return err
	}
	// 刷新缓存
	return s.refreshCache(ctx)
}

// DeleteSensitiveWord 删除敏感词
func (s *ContentFilterService) DeleteSensitiveWord(ctx context.Context, id uint) error {
	if err := s.wordRepo.Delete(ctx, id); err != nil {
		return err
	}
	// 刷新缓存
	return s.refreshCache(ctx)
}

// GetSensitiveWords 获取敏感词列表
func (s *ContentFilterService) GetSensitiveWords(ctx context.Context, wordType string, category string, onlyActive bool) ([]*models.SensitiveWord, error) {
	query := s.db.WithContext(ctx).Model(&models.SensitiveWord{})

	if wordType != "" {
		query = query.Where("word_type = ?", wordType)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if onlyActive {
		query = query.Where("is_active = ?", true)
	}

	var words []models.SensitiveWord
	if err := query.Order("severity DESC, word ASC").Find(&words).Error; err != nil {
		return nil, err
	}

	result := make([]*models.SensitiveWord, len(words))
	for i, w := range words {
		result[i] = &w
	}
	return result, nil
}

// GetSensitiveWordStats 获取敏感词统计
func (s *ContentFilterService) GetSensitiveWordStats(ctx context.Context) (map[string]interface{}, error) {
	var totalWords int64
	var activeWords int64

	s.db.WithContext(ctx).Model(&models.SensitiveWord{}).Count(&totalWords)
	s.db.WithContext(ctx).Model(&models.SensitiveWord{}).Where("is_active = ?", true).Count(&activeWords)

	// 按类型统计
	type statsByType struct {
		WordType string `json:"word_type"`
		Count    int64  `json:"count"`
	}
	var typeStats []statsByType
	s.db.WithContext(ctx).Model(&models.SensitiveWord{}).
		Select("word_type, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("word_type").
		Find(&typeStats)

	return map[string]interface{}{
		"total_words":  totalWords,
		"active_words": activeWords,
		"by_type":      typeStats,
	}, nil
}

// BatchImportWords 批量导入敏感词
func (s *ContentFilterService) BatchImportWords(ctx context.Context, words []string, wordType string, severity string) (int, []string, error) {
	successCount := 0
	errors := []string{}

	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		sensitiveWord := &models.SensitiveWord{
			Word:     word,
			WordType: wordType,
			Severity: severity,
			IsActive: true,
		}

		if err := s.wordRepo.Create(ctx, sensitiveWord); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", word, err.Error()))
		} else {
			successCount++
		}
	}

	// 刷新缓存
	s.refreshCache(ctx)

	return successCount, errors, nil
}
