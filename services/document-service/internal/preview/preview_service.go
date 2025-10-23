package preview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"law-oa-go/internal/config"
	"law-oa-go/internal/storage"
)

// PreviewService 预览服务
type PreviewService struct {
	db           *gorm.DB
	logger       *logrus.Logger
	config       *config.Config
	storage      storage.Storage
	renderEngine *RenderEngine
}

// NewPreviewService 创建预览服务
func NewPreviewService(db *gorm.DB, logger *logrus.Logger, config *config.Config, storage storage.Storage) *PreviewService {
	renderEngine := NewRenderEngine(logger, config)

	return &PreviewService{
		db:           db,
		logger:       logger,
		config:       config,
		storage:      storage,
		renderEngine: renderEngine,
	}
}

// PreviewOptions 预览选项
type PreviewOptions struct {
	Width         int                    `json:"width"`
	Height        int                    `json:"height"`
	Scale         float64                `json:"scale"`
	Quality       int                    `json:"quality"`        // 图片质量 1-100
	Format        string                 `json:"format"`         // 输出格式: jpg, png, webp
	PageNumbers   []int                  `json:"page_numbers"`   // 指定页码，空表示全部
	Thumbnail     bool                   `json:"thumbnail"`      // 是否生成缩略图
	Watermark     *WatermarkOptions      `json:"watermark"`      // 水印选项
	Annotations   bool                   `json:"annotations"`    // 是否包含注释
	Forms         bool                   `json:"forms"`           // 是否包含表单
	CacheEnabled  bool                   `json:"cache_enabled"`  // 是否启用缓存
	CacheTTL      time.Duration          `json:"cache_ttl"`      // 缓存时间
	CustomOptions map[string]interface{} `json:"custom_options"` // 自定义选项
}

// WatermarkOptions 水印选项
type WatermarkOptions struct {
	Text        string  `json:"text"`
	Font        string  `json:"font"`
	Size        int     `json:"size"`
	Color       string  `json:"color"`
	Opacity     float64 `json:"opacity"`
	Position    string  `json:"position"`    // center, top-left, top-right, bottom-left, bottom-right
	Rotation    float64 `json:"rotation"`    // 旋转角度
	Margin      int     `json:"margin"`      // 边距
	Repeat      bool    `json:"repeat"`      // 是否重复
	Background  bool    `json:"background"`  // 是否作为背景
}

// PreviewResult 预览结果
type PreviewResult struct {
	Success      bool                   `json:"success"`
	Format       string                 `json:"format"`
	PageCount    int                    `json:"page_count"`
	Pages        []PagePreview          `json:"pages"`
	Thumbnails   []Thumbnail            `json:"thumbnails,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	RenderTime   time.Duration          `json:"render_time"`
	CacheKey     string                 `json:"cache_key,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// PagePreview 页面预览
type PagePreview struct {
	PageNumber   int                    `json:"page_number"`
	Width        int                    `json:"width"`
	Height       int                    `json:"height"`
	ImageURL     string                 `json:"image_url"`
	ImageData    []byte                 `json:"image_data,omitempty"`
	TextContent  string                 `json:"text_content,omitempty"`
	Annotations  []AnnotationInfo       `json:"annotations,omitempty"`
	Links        []LinkInfo             `json:"links,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Thumbnail 缩略图
type Thumbnail struct {
	PageNumber int    `json:"page_number"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	ImageURL   string `json:"image_url"`
	ImageData  []byte `json:"image_data,omitempty"`
}

// AnnotationInfo 注释信息
type AnnotationInfo struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Content     string      `json:"content"`
	Position    Rectangle   `json:"position"`
	Color       string      `json:"color"`
	Author      string      `json:"author"`
	CreatedAt   time.Time   `json:"created_at"`
}

// LinkInfo 链接信息
type LinkInfo struct {
	URL      string    `json:"url"`
	Area     Rectangle `json:"area"`
	Text     string    `json:"text"`
	Page     int       `json:"page"`
}

// Rectangle 矩形区域
type Rectangle struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// GeneratePreview 生成文档预览
func (s *PreviewService) GeneratePreview(ctx context.Context, documentID uint, versionID *uint, options PreviewOptions) (*PreviewResult, error) {
	startTime := time.Now()

	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"version_id":  versionID,
		"options":     options,
	}).Info("开始生成文档预览")

	// 获取文档版本信息
	version, err := s.getDocumentVersion(documentID, versionID)
	if err != nil {
		return nil, fmt.Errorf("获取文档版本失败: %w", err)
	}

	// 检查缓存
	if options.CacheEnabled {
		if cached := s.checkCache(version, options); cached != nil {
			s.logger.Info("使用缓存的预览结果")
			cached.RenderTime = 0 // 缓存结果不计入渲染时间
			return cached, nil
		}
	}

	// 根据文档类型选择渲染引擎
	renderResult, err := s.renderDocument(ctx, version, options)
	if err != nil {
		s.logger.WithError(err).Error("文档渲染失败")
		return &PreviewResult{
			Success:      false,
			ErrorMessage: err.Error(),
			RenderTime:   time.Since(startTime),
		}, nil
	}

	// 处理渲染结果
	result := s.processRenderResult(renderResult, options)
	result.RenderTime = time.Since(startTime)

	// 缓存结果
	if options.CacheEnabled && result.Success {
		s.saveCache(version, options, result)
	}

	// 更新渲染状态
	err = s.updateRenderStatus(version.ID, "completed", result)
	if err != nil {
		s.logger.WithError(err).Warn("更新渲染状态失败")
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"version_id":  versionID,
		"page_count":  result.PageCount,
		"render_time": result.RenderTime,
	}).Info("文档预览生成完成")

	return result, nil
}

// GenerateThumbnail 生成缩略图
func (s *PreviewService) GenerateThumbnail(ctx context.Context, documentID uint, versionID *uint, pageNumber int, size int) (*Thumbnail, error) {
	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"version_id":  versionID,
		"page_number": pageNumber,
		"size":        size,
	}).Info("开始生成缩略图")

	// 获取文档版本
	version, err := s.getDocumentVersion(documentID, versionID)
	if err != nil {
		return nil, fmt.Errorf("获取文档版本失败: %w", err)
	}

	// 生成预览选项
	options := PreviewOptions{
		Width:        size,
		Height:       size,
		Scale:        1.0,
		Quality:      80,
		Format:       "jpg",
		PageNumbers:  []int{pageNumber},
		Thumbnail:    true,
		CacheEnabled: true,
		CacheTTL:     24 * time.Hour,
	}

	// 渲染指定页面
	renderResult, err := s.renderDocument(ctx, version, options)
	if err != nil {
		return nil, fmt.Errorf("渲染缩略图失败: %w", err)
	}

	// 处理缩略图结果
	if len(renderResult.Pages) == 0 {
		return nil, fmt.Errorf("没有找到指定页面")
	}

	page := renderResult.Pages[0]
	thumbnail := &Thumbnail{
		PageNumber: pageNumber,
		Width:      page.Width,
		Height:     page.Height,
		ImageURL:   page.ImageURL,
		ImageData:  page.ImageData,
	}

	// 保存缩略图到存储
	if thumbnail.ImageData != nil {
		thumbnailPath := s.getThumbnailPath(version, pageNumber, size)
		err = s.storage.Put(ctx, thumbnailPath, bytes.NewReader(thumbnail.ImageData))
		if err != nil {
			s.logger.WithError(err).Warn("保存缩略图失败")
		} else {
			thumbnail.ImageURL = s.storage.GetURL(thumbnailPath)
			thumbnail.ImageData = nil // 清空内存数据
		}
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"page_number": pageNumber,
		"width":       thumbnail.Width,
		"height":      thumbnail.Height,
	}).Info("缩略图生成完成")

	return thumbnail, nil
}

// ExtractText 提取文档文本
func (s *PreviewService) ExtractText(ctx context.Context, documentID uint, versionID *uint) (map[int]string, error) {
	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"version_id":  versionID,
	}).Info("开始提取文档文本")

	// 获取文档版本
	version, err := s.getDocumentVersion(documentID, versionID)
	if err != nil {
		return nil, fmt.Errorf("获取文档版本失败: %w", err)
	}

	// 根据文档类型提取文本
	textPages, err := s.extractDocumentText(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("提取文档文本失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"version_id":  versionID,
		"page_count":  len(textPages),
	}).Info("文档文本提取完成")

	return textPages, nil
}

// GetDocumentInfo 获取文档信息
func (s *PreviewService) GetDocumentInfo(ctx context.Context, documentID uint, versionID *uint) (map[string]interface{}, error) {
	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"version_id":  versionID,
	}).Info("获取文档信息")

	// 获取文档版本
	version, err := s.getDocumentVersion(documentID, versionID)
	if err != nil {
		return nil, fmt.Errorf("获取文档版本失败: %w", err)
	}

	// 获取文档元信息
	metadata := s.extractDocumentMetadata(ctx, version)

	// 添加基础信息
	info := map[string]interface{}{
		"document_id":   documentID,
		"version_id":    version.ID,
		"version_number": version.VersionNumber,
		"title":         version.Title,
		"content_type":  version.ContentType,
		"file_size":     version.FileSize,
		"page_count":    version.PageCount,
		"word_count":    version.WordCount,
		"character_count": version.CharacterCount,
		"created_at":    version.CreatedAt,
		"updated_at":    version.UpdatedAt,
		"render_status": version.RenderStatus,
	}

	// 合并元数据
	for k, v := range metadata {
		info[k] = v
	}

	return info, nil
}

// SearchInDocument 在文档中搜索
func (s *PreviewService) SearchInDocument(ctx context.Context, documentID uint, versionID *uint, query string, options SearchOptions) (*SearchResult, error) {
	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"version_id":  versionID,
		"query":       query,
	}).Info("在文档中搜索")

	// 获取文档版本
	version, err := s.getDocumentVersion(documentID, versionID)
	if err != nil {
		return nil, fmt.Errorf("获取文档版本失败: %w", err)
	}

	// 提取文档文本
	textPages, err := s.extractDocumentText(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("提取文档文本失败: %w", err)
	}

	// 执行搜索
	result := s.performSearch(textPages, query, options)

	s.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"query":       query,
		"match_count": result.TotalMatches,
		"page_count":  len(result.Matches),
	}).Info("文档搜索完成")

	return result, nil
}

// SearchOptions 搜索选项
type SearchOptions struct {
	CaseSensitive  bool    `json:"case_sensitive"`
	WholeWord      bool    `json:"whole_word"`
	Regex          bool    `json:"regex"`
	ContextLength  int     `json:"context_length"`  // 上下文长度
	MaxResults     int     `json:"max_results"`     // 最大结果数
	HighlightColor string  `json:"highlight_color"` // 高亮颜色
}

// SearchResult 搜索结果
type SearchResult struct {
	Query        string              `json:"query"`
	TotalMatches int                 `json:"total_matches"`
	Matches      []PageSearchMatches `json:"matches"`
	RenderTime   time.Duration       `json:"render_time"`
}

// PageSearchMatches 页面搜索匹配
type PageSearchMatches struct {
	PageNumber int        `json:"page_number"`
	Matches    []TextMatch `json:"matches"`
}

// TextMatch 文本匹配
type TextMatch struct {
	Text        string     `json:"text"`
	StartIndex  int        `json:"start_index"`
	EndIndex    int        `json:"end_index"`
	LineNumber  int        `json:"line_number"`
	ColumnStart int        `json:"column_start"`
	ColumnEnd   int        `json:"column_end"`
	Context     string     `json:"context"`
	Score       float64    `json:"score"`
}

// 内部方法

// getDocumentVersion 获取文档版本
func (s *PreviewService) getDocumentVersion(documentID uint, versionID *uint) (*DocumentVersion, error) {
	var version DocumentVersion
	query := s.db.Where("document_id = ?", documentID)

	if versionID != nil {
		query = query.Where("id = ?", *versionID)
	} else {
		// 获取最新版本
		query = query.Order("version_number desc")
	}

	err := query.First(&version).Error
	if err != nil {
		return nil, err
	}

	return &version, nil
}

// checkCache 检查缓存
func (s *PreviewService) checkCache(version *DocumentVersion, options PreviewOptions) *PreviewResult {
	cacheKey := s.generateCacheKey(version, options)

	// 这里应该从Redis或其他缓存系统获取
	// 简化实现，返回nil表示没有缓存
	return nil
}

// saveCache 保存缓存
func (s *PreviewService) saveCache(version *DocumentVersion, options PreviewOptions, result *PreviewResult) {
	cacheKey := s.generateCacheKey(version, options)

	// 这里应该保存到Redis或其他缓存系统
	s.logger.WithField("cache_key", cacheKey).Debug("保存预览缓存")
}

// generateCacheKey 生成缓存键
func (s *PreviewService) generateCacheKey(version *DocumentVersion, options PreviewOptions) string {
	optionsJSON, _ := json.Marshal(options)
	return fmt.Sprintf("preview:%d:%s:%s", version.ID, version.FileHash, string(optionsJSON))
}

// renderDocument 渲染文档
func (s *PreviewService) renderDocument(ctx context.Context, version *DocumentVersion, options PreviewOptions) (*RenderResult, error) {
	// 根据文档类型选择渲染引擎
	switch strings.ToLower(version.ContentType) {
	case "pdf":
		return s.renderEngine.RenderPDF(ctx, version, options)
	case "docx", "doc":
		return s.renderEngine.RenderWord(ctx, version, options)
	case "xlsx", "xls":
		return s.renderEngine.RenderExcel(ctx, version, options)
	case "pptx", "ppt":
		return s.renderEngine.RenderPowerPoint(ctx, version, options)
	case "txt", "md":
		return s.renderEngine.RenderText(ctx, version, options)
	case "jpg", "jpeg", "png", "gif", "bmp", "webp":
		return s.renderEngine.RenderImage(ctx, version, options)
	default:
		return nil, fmt.Errorf("不支持的文档类型: %s", version.ContentType)
	}
}

// processRenderResult 处理渲染结果
func (s *PreviewService) processRenderResult(renderResult *RenderResult, options PreviewOptions) *PreviewResult {
	result := &PreviewResult{
		Success:    renderResult.Success,
		Format:     options.Format,
		PageCount:  len(renderResult.Pages),
		Pages:      make([]PagePreview, len(renderResult.Pages)),
		Metadata:   renderResult.Metadata,
	}

	// 处理每个页面
	for i, page := range renderResult.Pages {
		// 应用水印
		if options.Watermark != nil {
			page.ImageData = s.applyWatermark(page.ImageData, options.Watermark)
		}

		// 转换格式
		if options.Format != "" && options.Format != "png" {
			page.ImageData = s.convertImageFormat(page.ImageData, options.Format, options.Quality)
		}

		// 保存图片到存储
		if page.ImageData != nil {
			imagePath := s.getPreviewImagePath(renderResult.Version, page.PageNumber, options)
			err := s.storage.Put(context.Background(), imagePath, bytes.NewReader(page.ImageData))
			if err != nil {
				s.logger.WithError(err).Warn("保存预览图片失败")
			} else {
				page.ImageURL = s.storage.GetURL(imagePath)
				page.ImageData = nil // 清空内存数据
			}
		}

		result.Pages[i] = PagePreview{
			PageNumber:   page.PageNumber,
			Width:        page.Width,
			Height:       page.Height,
			ImageURL:     page.ImageURL,
			TextContent:  page.TextContent,
			Annotations:  s.convertAnnotations(page.Annotations),
			Links:        s.convertLinks(page.Links),
			Metadata:     page.Metadata,
		}
	}

	// 生成缩略图
	if options.Thumbnail {
		result.Thumbnails = s.generateThumbnails(renderResult.Pages, options)
	}

	return result
}

// applyWatermark 应用水印
func (s *PreviewService) applyWatermark(imageData []byte, watermark *WatermarkOptions) []byte {
	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		s.logger.WithError(err).Warn("解码图片失败，跳过水印")
		return imageData
	}

	// 这里应该实现水印逻辑
	// 简化实现，返回原图
	buffer := new(bytes.Buffer)
	err = jpeg.Encode(buffer, img, &jpeg.Options{Quality: 90})
	if err != nil {
		s.logger.WithError(err).Warn("编码图片失败，返回原图")
		return imageData
	}

	return buffer.Bytes()
}

// convertImageFormat 转换图片格式
func (s *PreviewService) convertImageFormat(imageData []byte, format string, quality int) []byte {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return imageData
	}

	buffer := new(bytes.Buffer)

	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		err = jpeg.Encode(buffer, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(buffer, img)
	case "webp":
		err = webp.Encode(buffer, img, &webp.Options{Quality: float32(quality)})
	default:
		return imageData
	}

	if err != nil {
		s.logger.WithError(err).Warn("转换图片格式失败")
		return imageData
	}

	return buffer.Bytes()
}

// getPreviewImagePath 获取预览图片路径
func (s *PreviewService) getPreviewImagePath(version *DocumentVersion, pageNumber int, options PreviewOptions) string {
	ext := options.Format
	if ext == "" {
		ext = "png"
	}
	return fmt.Sprintf("previews/%d/%d/page_%d.%s", version.DocumentID, version.ID, pageNumber, ext)
}

// getThumbnailPath 获取缩略图路径
func (s *PreviewService) getThumbnailPath(version *DocumentVersion, pageNumber int, size int) string {
	return fmt.Sprintf("thumbnails/%d/%d/page_%d_%d.jpg", version.DocumentID, version.ID, pageNumber, size)
}

// convertAnnotations 转换注释信息
func (s *PreviewService) convertAnnotations(annotations []RenderAnnotation) []AnnotationInfo {
	result := make([]AnnotationInfo, len(annotations))
	for i, ann := range annotations {
		result[i] = AnnotationInfo{
			ID:        ann.ID,
			Type:      ann.Type,
			Content:   ann.Content,
			Position:  ann.Position,
			Color:     ann.Color,
			Author:    ann.Author,
			CreatedAt: ann.CreatedAt,
		}
	}
	return result
}

// convertLinks 转换链接信息
func (s *PreviewService) convertLinks(links []RenderLink) []LinkInfo {
	result := make([]LinkInfo, len(links))
	for i, link := range links {
		result[i] = LinkInfo{
			URL:  link.URL,
			Area: link.Area,
			Text: link.Text,
			Page: link.Page,
		}
	}
	return result
}

// generateThumbnails 生成缩略图
func (s *PreviewService) generateThumbnails(pages []RenderPage, options PreviewOptions) []Thumbnail {
	thumbnails := make([]Thumbnail, len(pages))
	thumbnailSize := 150 // 默认缩略图大小

	for i, page := range pages {
		// 调整图片大小
		var thumbnailData []byte
		if page.ImageData != nil {
			img, _, err := image.Decode(bytes.NewReader(page.ImageData))
			if err == nil {
				thumbnail := imaging.Fill(img, thumbnailSize, thumbnailSize, imaging.Center, imaging.Lanczos)
				buffer := new(bytes.Buffer)
				err = jpeg.Encode(buffer, thumbnail, &jpeg.Options{Quality: 80})
				if err == nil {
					thumbnailData = buffer.Bytes()
				}
			}
		}

		// 计算缩略图尺寸
		aspectRatio := float64(page.Width) / float64(page.Height)
		var thumbWidth, thumbHeight int
		if aspectRatio > 1 {
			thumbWidth = thumbnailSize
			thumbHeight = int(float64(thumbnailSize) / aspectRatio)
		} else {
			thumbHeight = thumbnailSize
			thumbWidth = int(float64(thumbnailSize) * aspectRatio)
		}

		// 保存缩略图
		var thumbnailURL string
		if thumbnailData != nil {
			thumbnailPath := fmt.Sprintf("thumbnails/%d/%d/thumb_%d.jpg",
				pages[0].DocumentID, pages[0].VersionID, page.PageNumber)
			err := s.storage.Put(context.Background(), thumbnailPath, bytes.NewReader(thumbnailData))
			if err == nil {
				thumbnailURL = s.storage.GetURL(thumbnailPath)
			}
		}

		thumbnail := Thumbnail{
			PageNumber: page.PageNumber,
			Width:      thumbWidth,
			Height:     thumbHeight,
			ImageURL:   thumbnailURL,
		}

		// 如果需要返回图片数据
		if options.Width <= 300 && options.Height <= 300 {
			thumbnail.ImageData = thumbnailData
		}

		thumbnails[i] = thumbnail
	}

	return thumbnails
}

// extractDocumentText 提取文档文本
func (s *PreviewService) extractDocumentText(ctx context.Context, version *DocumentVersion) (map[int]string, error) {
	switch strings.ToLower(version.ContentType) {
	case "pdf":
		return s.renderEngine.ExtractPDFText(ctx, version)
	case "docx", "doc":
		return s.renderEngine.ExtractWordText(ctx, version)
	case "xlsx", "xls":
		return s.renderEngine.ExtractExcelText(ctx, version)
	case "pptx", "ppt":
		return s.renderEngine.ExtractPowerPointText(ctx, version)
	case "txt", "md":
		return s.extractPlainText(ctx, version)
	default:
		return nil, fmt.Errorf("不支持的文档类型: %s", version.ContentType)
	}
}

// extractPlainText 提取纯文本
func (s *PreviewService) extractPlainText(ctx context.Context, version *DocumentVersion) (map[int]string, error) {
	// 从存储读取文件
	reader, err := s.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	text := string(data)
	result := make(map[int]string)
	result[1] = text // 纯文本文件视为单页

	return result, nil
}

// extractDocumentMetadata 提取文档元数据
func (s *PreviewService) extractDocumentMetadata(ctx context.Context, version *DocumentVersion) map[string]interface{} {
	metadata := make(map[string]interface{})

	switch strings.ToLower(version.ContentType) {
	case "pdf":
		return s.renderEngine.ExtractPDFMetadata(ctx, version)
	case "docx", "doc":
		return s.renderEngine.ExtractWordMetadata(ctx, version)
	case "xlsx", "xls":
		return s.renderEngine.ExtractExcelMetadata(ctx, version)
	case "pptx", "ppt":
		return s.renderEngine.ExtractPowerPointMetadata(ctx, version)
	default:
		return metadata
	}
}

// performSearch 执行搜索
func (s *PreviewService) performSearch(textPages map[int]string, query string, options SearchOptions) *SearchResult {
	startTime := time.Now()
	result := &SearchResult{
		Query:  query,
		Matches: make([]PageSearchMatches, 0),
	}

	for pageNum, text := range textPages {
		matches := s.searchInText(text, query, options)
		if len(matches) > 0 {
			result.Matches = append(result.Matches, PageSearchMatches{
				PageNumber: pageNum,
				Matches:    matches,
			})
			result.TotalMatches += len(matches)
		}
	}

	// 限制结果数量
	if options.MaxResults > 0 && result.TotalMatches > options.MaxResults {
		// 这里应该实现结果截断逻辑
	}

	result.RenderTime = time.Since(startTime)
	return result
}

// searchInText 在文本中搜索
func (s *PreviewService) searchInText(text string, query string, options SearchOptions) []TextMatch {
	// 这里应该实现具体的搜索逻辑
	// 支持正则表达式、全词匹配、大小写敏感等选项
	// 简化实现，返回空结果
	return []TextMatch{}
}

// updateRenderStatus 更新渲染状态
func (s *PreviewService) updateRenderStatus(versionID uint, status string, result *PreviewResult) error {
	updates := map[string]interface{}{
		"render_status": status,
		"preview_json":  nil, // 清空旧的预览数据
	}

	if result.Success {
		previewJSON, _ := json.Marshal(result)
		updates["preview_json"] = string(previewJSON)
	}

	return s.db.Model(&DocumentVersion{}).Where("id = ?", versionID).Updates(updates).Error
}

// RenderResult 渲染结果
type RenderResult struct {
	Success bool         `json:"success"`
	Version *DocumentVersion `json:"version"`
	Pages   []RenderPage `json:"pages"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RenderPage 渲染页面
type RenderPage struct {
	PageNumber   int                `json:"page_number"`
	Width        int                `json:"width"`
	Height       int                `json:"height"`
	ImageData    []byte             `json:"image_data"`
	ImageURL     string             `json:"image_url"`
	TextContent  string             `json:"text_content"`
	Annotations  []RenderAnnotation `json:"annotations"`
	Links        []RenderLink       `json:"links"`
	Metadata     map[string]interface{} `json:"metadata"`
	DocumentID   uint               `json:"document_id"`
	VersionID    uint               `json:"version_id"`
}

// RenderAnnotation 渲染注释
type RenderAnnotation struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Position  Rectangle `json:"position"`
	Color     string    `json:"color"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// RenderLink 渲染链接
type RenderLink struct {
	URL  string    `json:"url"`
	Area Rectangle `json:"area"`
	Text string    `json:"text"`
	Page int       `json:"page"`
}