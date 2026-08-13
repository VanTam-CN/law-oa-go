package preview

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/sirupsen/logrus"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/presentation"
	"github.com/unidoc/unioffice/spreadsheet"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"

	"law-oa-go/internal/config"
	"law-oa-go/internal/storage"
)

// RenderEngine 渲染引擎
type RenderEngine struct {
	logger    *logrus.Logger
	config    *config.Config
	storage   storage.Storage
}

// NewRenderEngine 创建渲染引擎
func NewRenderEngine(logger *logrus.Logger, config *config.Config) *RenderEngine {
	// 初始化UniDoc许可证（如果需要）
	// initializeUniDocLicense()

	return &RenderEngine{
		logger:  logger,
		config:  config,
	}
}

// RenderPDF 渲染PDF文档
func (e *RenderEngine) RenderPDF(ctx context.Context, version *DocumentVersion, options PreviewOptions) (*RenderResult, error) {
	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"version_id":  version.ID,
		"file_path":   version.StoragePath,
	}).Info("开始渲染PDF文档")

	// 从存储获取PDF文件
	pdfFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("获取PDF文件失败: %w", err)
	}
	defer pdfFile.Close()

	// 读取PDF数据
	pdfData, err := io.ReadAll(pdfFile)
	if err != nil {
		return nil, fmt.Errorf("读取PDF文件失败: %w", err)
	}

	// 使用pdfcpu渲染PDF
	pages, err := e.renderPDFWithPdfcpu(pdfData, options)
	if err != nil {
		return nil, fmt.Errorf("pdfcpu渲染失败: %w", err)
	}

	// 提取文本内容
	textPages := make(map[int]string)
	for _, page := range pages {
		text, err := e.extractPDFTextWithPdfcpu(pdfData, page.PageNumber)
		if err != nil {
			e.logger.WithError(err).Warnf("提取第%d页文本失败", page.PageNumber)
		}
		textPages[page.PageNumber] = text
		page.TextContent = text
	}

	// 提取链接信息
	for i := range pages {
		links, err := e.extractPDFLinksWithPdfcpu(pdfData, pages[i].PageNumber)
		if err != nil {
			e.logger.WithError(err).Warnf("提取第%d页链接失败", pages[i].PageNumber)
		}
		pages[i].Links = e.convertPDFLinksToRenderLinks(links)
	}

	// 提取注释信息
	if options.Annotations {
		for i := range pages {
			annotations, err := e.extractPDFAnnotationsWithPdfcpu(pdfData, pages[i].PageNumber)
			if err != nil {
				e.logger.WithError(err).Warnf("提取第%d页注释失败", pages[i].PageNumber)
			}
			pages[i].Annotations = e.convertPDFAnnotationsToRenderAnnotations(annotations)
		}
	}

	// 构建渲染结果
	result := &RenderResult{
		Success: true,
		Version: version,
		Pages:   pages,
		Metadata: map[string]interface{}{
			"content_type": "pdf",
			"page_count":   len(pages),
			"render_engine": "pdfcpu",
		},
	}

	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"page_count":  len(pages),
	}).Info("PDF文档渲染完成")

	return result, nil
}

// RenderWord 渲染Word文档
func (e *RenderEngine) RenderWord(ctx context.Context, version *DocumentVersion, options PreviewOptions) (*RenderResult, error) {
	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"version_id":  version.ID,
		"file_path":   version.StoragePath,
	}).Info("开始渲染Word文档")

	// 从存储获取Word文件
	wordFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("获取Word文件失败: %w", err)
	}
	defer wordFile.Close()

	// 读取Word数据
	wordData, err := io.ReadAll(wordFile)
	if err != nil {
		return nil, fmt.Errorf("读取Word文件失败: %w", err)
	}

	// 转换为PDF然后渲染
	pdfData, err := e.convertWordToPDF(wordData)
	if err != nil {
		return nil, fmt.Errorf("Word转PDF失败: %w", err)
	}

	// 渲染PDF
	pages, err := e.renderPDFWithPdfcpu(pdfData, options)
	if err != nil {
		return nil, fmt.Errorf("渲染转换后的PDF失败: %w", err)
	}

	// 提取文本内容
	textPages := make(map[int]string)
	for _, page := range pages {
		text, err := e.extractPDFTextWithPdfcpu(pdfData, page.PageNumber)
		if err != nil {
			e.logger.WithError(err).Warnf("提取第%d页文本失败", page.PageNumber)
		}
		textPages[page.PageNumber] = text
		page.TextContent = text
	}

	// 构建渲染结果
	result := &RenderResult{
		Success: true,
		Version: version,
		Pages:   pages,
		Metadata: map[string]interface{}{
			"content_type":   "docx",
			"page_count":     len(pages),
			"render_engine":  "unioffice",
			"original_format": "word",
		},
	}

	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"page_count":  len(pages),
	}).Info("Word文档渲染完成")

	return result, nil
}

// RenderExcel 渲染Excel文档
func (e *RenderEngine) RenderExcel(ctx context.Context, version *DocumentVersion, options PreviewOptions) (*RenderResult, error) {
	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"version_id":  version.ID,
		"file_path":   version.StoragePath,
	}).Info("开始渲染Excel文档")

	// 从存储获取Excel文件
	excelFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("获取Excel文件失败: %w", err)
	}
	defer excelFile.Close()

	// 读取Excel数据
	excelData, err := io.ReadAll(excelFile)
	if err != nil {
		return nil, fmt.Errorf("读取Excel文件失败: %w", err)
	}

	// 使用unioffice渲染Excel
	pages, err := e.renderExcelWithUniOffice(excelData, options)
	if err != nil {
		return nil, fmt.Errorf("unioffice渲染Excel失败: %w", err)
	}

	// 提取文本内容
	textPages := make(map[int]string)
	for _, page := range pages {
		textPages[page.PageNumber] = page.TextContent
	}

	// 构建渲染结果
	result := &RenderResult{
		Success: true,
		Version: version,
		Pages:   pages,
		Metadata: map[string]interface{}{
			"content_type":  "xlsx",
			"page_count":    len(pages),
			"render_engine": "unioffice",
			"sheets":        e.extractExcelSheets(excelData),
		},
	}

	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"page_count":  len(pages),
	}).Info("Excel文档渲染完成")

	return result, nil
}

// RenderPowerPoint 渲染PowerPoint文档
func (e *RenderEngine) RenderPowerPoint(ctx context.Context, version *DocumentVersion, options PreviewOptions) (*RenderResult, error) {
	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"version_id":  version.ID,
		"file_path":   version.StoragePath,
	}).Info("开始渲染PowerPoint文档")

	// 从存储获取PowerPoint文件
	pptFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("获取PowerPoint文件失败: %w", err)
	}
	defer pptFile.Close()

	// 读取PowerPoint数据
	pptData, err := io.ReadAll(pptFile)
	if err != nil {
		return nil, fmt.Errorf("读取PowerPoint文件失败: %w", err)
	}

	// 使用unioffice渲染PowerPoint
	pages, err := e.renderPowerPointWithUniOffice(pptData, options)
	if err != nil {
		return nil, fmt.Errorf("unioffice渲染PowerPoint失败: %w", err)
	}

	// 提取文本内容
	textPages := make(map[int]string)
	for _, page := range pages {
		textPages[page.PageNumber] = page.TextContent
	}

	// 构建渲染结果
	result := &RenderResult{
		Success: true,
		Version: version,
		Pages:   pages,
		Metadata: map[string]interface{}{
			"content_type":  "pptx",
			"page_count":    len(pages),
			"render_engine": "unioffice",
			"slides":        e.extractPowerPointSlides(pptData),
		},
	}

	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"page_count":  len(pages),
	}).Info("PowerPoint文档渲染完成")

	return result, nil
}

// RenderText 渲染文本文档
func (e *RenderEngine) RenderText(ctx context.Context, version *DocumentVersion, options PreviewOptions) (*RenderResult, error) {
	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"version_id":  version.ID,
		"file_path":   version.StoragePath,
	}).Info("开始渲染文本文档")

	// 从存储获取文本文件
	textFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("获取文本文件失败: %w", err)
	}
	defer textFile.Close()

	// 读取文本数据
	data, err := io.ReadAll(textFile)
	if err != nil {
		return nil, fmt.Errorf("读取文本文件失败: %w", err)
	}

	text := string(data)

	// 将文本分割为页面
	pages := e.splitTextIntoPages(text, options)

	// 生成页面图片
	renderPages := make([]RenderPage, len(pages))
	for i, page := range pages {
		// 生成文本图片
		imgData := e.renderTextToImage(page, options)

		renderPages[i] = RenderPage{
			PageNumber:   i + 1,
			Width:        options.Width,
			Height:       options.Height,
			ImageData:    imgData,
			TextContent:  page,
			Annotations:  []RenderAnnotation{},
			Links:        []RenderLink{},
			Metadata: map[string]interface{}{
				"content_type": "text",
				"char_count":   len(page),
				"line_count":   strings.Count(page, "\n") + 1,
			},
			DocumentID: version.DocumentID,
			VersionID:  version.ID,
		}
	}

	// 构建渲染结果
	result := &RenderResult{
		Success: true,
		Version: version,
		Pages:   renderPages,
		Metadata: map[string]interface{}{
			"content_type":   "text",
			"page_count":     len(renderPages),
			"render_engine":  "internal",
			"total_chars":    len(text),
			"total_lines":    strings.Count(text, "\n") + 1,
		},
	}

	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"page_count":  len(renderPages),
		"char_count":  len(text),
	}).Info("文本文档渲染完成")

	return result, nil
}

// RenderImage 渲染图片文档
func (e *RenderEngine) RenderImage(ctx context.Context, version *DocumentVersion, options PreviewOptions) (*RenderResult, error) {
	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"version_id":  version.ID,
		"file_path":   version.StoragePath,
	}).Info("开始渲染图片文档")

	// 从存储获取图片文件
	imageFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("获取图片文件失败: %w", err)
	}
	defer imageFile.Close()

	// 读取图片数据
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		return nil, fmt.Errorf("读取图片文件失败: %w", err)
	}

	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %w", err)
	}

	// 调整图片大小
	resizedImg := e.resizeImage(img, options)

	// 编码为指定格式
	var outputData []byte
	buffer := new(bytes.Buffer)

	format := options.Format
	if format == "" {
		format = "png"
	}

	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		err = jpeg.Encode(buffer, resizedImg, &jpeg.Options{Quality: options.Quality})
	case "png":
		err = png.Encode(buffer, resizedImg)
	default:
		err = png.Encode(buffer, resizedImg) // 默认PNG
	}

	if err != nil {
		return nil, fmt.Errorf("编码图片失败: %w", err)
	}

	outputData = buffer.Bytes()

	// 构建渲染结果
	bounds := resizedImg.Bounds()
	renderPage := RenderPage{
		PageNumber:   1,
		Width:        bounds.Dx(),
		Height:       bounds.Dy(),
		ImageData:    outputData,
		TextContent:  "",
		Annotations:  []RenderAnnotation{},
		Links:        []RenderLink{},
		Metadata: map[string]interface{}{
			"content_type":  strings.ToLower(filepath.Ext(version.StoragePath)[1:]),
			"original_size": fmt.Sprintf("%dx%d", img.Bounds().Dx(), img.Bounds().Dy()),
			"output_size":   fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy()),
		},
		DocumentID: version.DocumentID,
		VersionID:  version.ID,
	}

	result := &RenderResult{
		Success: true,
		Version: version,
		Pages:   []RenderPage{renderPage},
		Metadata: map[string]interface{}{
			"content_type":  "image",
			"page_count":    1,
			"render_engine": "imaging",
		},
	}

	e.logger.WithFields(logrus.Fields{
		"document_id": version.DocumentID,
		"width":       bounds.Dx(),
		"height":      bounds.Dy(),
	}).Info("图片文档渲染完成")

	return result, nil
}

// 文本提取方法

// ExtractPDFText 提取PDF文本
func (e *RenderEngine) ExtractPDFText(ctx context.Context, version *DocumentVersion) (map[int]string, error) {
	// 从存储获取PDF文件
	pdfFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, err
	}
	defer pdfFile.Close()

	pdfData, err := io.ReadAll(pdfFile)
	if err != nil {
		return nil, err
	}

	textPages := make(map[int]string)
 pageCount := e.getPDFPageCount(pdfData)

	for i := 1; i <= pageCount; i++ {
		text, err := e.extractPDFTextWithPdfcpu(pdfData, i)
		if err != nil {
			e.logger.WithError(err).Warnf("提取第%d页文本失败", i)
			continue
		}
		textPages[i] = text
	}

	return textPages, nil
}

// ExtractWordText 提取Word文本
func (e *RenderEngine) ExtractWordText(ctx context.Context, version *DocumentVersion) (map[int]string, error) {
	// 从存储获取Word文件
	wordFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, err
	}
	defer wordFile.Close()

	wordData, err := io.ReadAll(wordFile)
	if err != nil {
		return nil, err
	}

	// 转换为PDF然后提取文本
	pdfData, err := e.convertWordToPDF(wordData)
	if err != nil {
		return nil, err
	}

	return e.ExtractPDFText(ctx, &DocumentVersion{
		DocumentID: version.DocumentID,
		StoragePath: version.StoragePath,
		FileHash:    version.FileHash,
		ContentType: "pdf",
	})
}

// ExtractExcelText 提取Excel文本
func (e *RenderEngine) ExtractExcelText(ctx context.Context, version *DocumentVersion) (map[int]string, error) {
	// 从存储获取Excel文件
	excelFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, err
	}
	defer excelFile.Close()

	excelData, err := io.ReadAll(excelFile)
	if err != nil {
		return nil, err
	}

	ss, err := spreadsheet.New(bytes.NewReader(excelData))
	if err != nil {
		return nil, err
	}
	defer ss.Close()

	textPages := make(map[int]string)
	pageNum := 1

	for i, sheet := range ss.Sheets() {
		var pageText strings.Builder
		pageText.WriteString(fmt.Sprintf("工作表 %d: %s\n\n", i+1, sheet.Name()))

		for _, row := range sheet.Rows() {
			var rowData []string
			for _, cell := range row.Cells() {
				rowData = append(rowData, cell.GetText())
			}
			pageText.WriteString(strings.Join(rowData, "\t") + "\n")
		}

		textPages[pageNum] = pageText.String()
		pageNum++
	}

	return textPages, nil
}

// ExtractPowerPointText 提取PowerPoint文本
func (e *RenderEngine) ExtractPowerPointText(ctx context.Context, version *DocumentVersion) (map[int]string, error) {
	// 从存储获取PowerPoint文件
	pptFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		return nil, err
	}
	defer pptFile.Close()

	pptData, err := io.ReadAll(pptFile)
	if err != nil {
		return nil, err
	}

	ppt, err := presentation.New(bytes.NewReader(pptData))
	if err != nil {
		return nil, err
	}
	defer ppt.Close()

	textPages := make(map[int]string)

	for i, slide := range ppt.Slides() {
		var slideText strings.Builder
		for _, para := range slide.Paragraphs() {
			for _, run := range para.Runs() {
				slideText.WriteString(run.GetText() + " ")
			}
			slideText.WriteString("\n")
		}
		textPages[i+1] = slideText.String()
	}

	return textPages, nil
}

// 元数据提取方法

// ExtractPDFMetadata 提取PDF元数据
func (e *RenderEngine) ExtractPDFMetadata(ctx context.Context, version *DocumentVersion) map[string]interface{} {
	metadata := make(map[string]interface{})

	// 从存储获取PDF文件
	pdfFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		e.logger.WithError(err).Warn("获取PDF文件失败")
		return metadata
	}
	defer pdfFile.Close()

	pdfData, err := io.ReadAll(pdfFile)
	if err != nil {
		e.logger.WithError(err).Warn("读取PDF文件失败")
		return metadata
	}

	// 使用pdfcpu提取元数据
	conf := pdfcpu.NewDefaultConfiguration()
	ctxPdf := pdfcpu.NewContext(conf)

	pdfReader, err := pdfcpu.Read(bytes.NewReader(pdfData), ctxPdf)
	if err != nil {
		e.logger.WithError(err).Warn("读取PDF失败")
		return metadata
	}

	// 提取基本信息
	metadata["page_count"] = len(pdfReader.PageTable.Pages)
	metadata["creator"] = pdfReader.XrefTable.Creator
	metadata["producer"] = pdfReader.XrefTable.Producer
	metadata["creation_date"] = pdfReader.XrefTable.Creation
	metadata["modification_date"] = pdfReader.XrefTable.Modification

	// 提取PDF版本
	metadata["pdf_version"] = pdfReader.XrefTable.HeaderVersion

	return metadata
}

// ExtractWordMetadata 提取Word元数据
func (e *RenderEngine) ExtractWordMetadata(ctx context.Context, version *DocumentVersion) map[string]interface{} {
	metadata := make(map[string]interface{})

	// 从存储获取Word文件
	wordFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		e.logger.WithError(err).Warn("获取Word文件失败")
		return metadata
	}
	defer wordFile.Close()

	wordData, err := io.ReadAll(wordFile)
	if err != nil {
		e.logger.WithError(err).Warn("读取Word文件失败")
		return metadata
	}

	doc, err := document.New(bytes.NewReader(wordData))
	if err != nil {
		e.logger.WithError(err).Warn("打开Word文档失败")
		return metadata
	}
	defer doc.Close()

	// 提取文档属性
	metadata["title"] = doc.GetTitle()
	metadata["author"] = doc.GetAuthor()
	metadata["subject"] = doc.GetSubject()
	metadata["keywords"] = doc.GetKeywords()
	metadata["category"] = doc.GetCategory()
	metadata["comments"] = doc.GetComments()

	// 提取统计信息
	metadata["page_count"] = doc.PageCount()
	metadata["paragraph_count"] = doc.ParagraphCount()
	metadata["word_count"] = doc.WordCount()
	metadata["character_count"] = doc.CharacterCount()

	// 提取创建和修改时间
	metadata["created"] = doc.Created()
	metadata["modified"] = doc.Modified()

	return metadata
}

// ExtractExcelMetadata 提取Excel元数据
func (e *RenderEngine) ExtractExcelMetadata(ctx context.Context, version *DocumentVersion) map[string]interface{} {
	metadata := make(map[string]interface{})

	// 从存储获取Excel文件
	excelFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		e.logger.WithError(err).Warn("获取Excel文件失败")
		return metadata
	}
	defer excelFile.Close()

	excelData, err := io.ReadAll(excelFile)
	if err != nil {
		e.logger.WithError(err).Warn("读取Excel文件失败")
		return metadata
	}

	ss, err := spreadsheet.New(bytes.NewReader(excelData))
	if err != nil {
		e.logger.WithError(err).Warn("打开Excel文档失败")
		return metadata
	}
	defer ss.Close()

	// 提取工作表信息
	var sheets []string
	var sheetRowCounts []int
	var sheetColumnCounts []int

	for _, sheet := range ss.Sheets() {
		sheets = append(sheets, sheet.Name())
		sheetRowCounts = append(sheetRowCounts, len(sheet.Rows()))
		if len(sheet.Rows()) > 0 {
			maxCols := 0
			for _, row := range sheet.Rows() {
				if len(row.Cells()) > maxCols {
					maxCols = len(row.Cells())
				}
			}
			sheetColumnCounts = append(sheetColumnCounts, maxCols)
		} else {
			sheetColumnCounts = append(sheetColumnCounts, 0)
		}
	}

	metadata["sheet_names"] = sheets
	metadata["sheet_counts"] = len(sheets)
	metadata["sheet_row_counts"] = sheetRowCounts
	metadata["sheet_column_counts"] = sheetColumnCounts

	return metadata
}

// ExtractPowerPointMetadata 提取PowerPoint元数据
func (e *RenderEngine) ExtractPowerPointMetadata(ctx context.Context, version *DocumentVersion) map[string]interface{} {
	metadata := make(map[string]interface{})

	// 从存储获取PowerPoint文件
	pptFile, err := e.storage.Get(ctx, version.StoragePath)
	if err != nil {
		e.logger.WithError(err).Warn("获取PowerPoint文件失败")
		return metadata
	}
	defer pptFile.Close()

	pptData, err := io.ReadAll(pptFile)
	if err != nil {
		e.logger.WithError(err).Warn("读取PowerPoint文件失败")
		return metadata
	}

	ppt, err := presentation.New(bytes.NewReader(pptData))
	if err != nil {
		e.logger.WithError(err).Warn("打开PowerPoint文档失败")
		return metadata
	}
	defer ppt.Close()

	// 提取演示文稿信息
	metadata["slide_count"] = len(ppt.Slides())
	metadata["title"] = ppt.GetTitle()
	metadata["author"] = ppt.GetAuthor()
	metadata["subject"] = ppt.GetSubject()

	return metadata
}

// 辅助方法

// renderPDFWithPdfcpu 使用pdfcpu渲染PDF
func (e *RenderEngine) renderPDFWithPdfcpu(pdfData []byte, options PreviewOptions) ([]RenderPage, error) {
	conf := pdfcpu.NewDefaultConfiguration()

	// 设置渲染选项
	if options.Width > 0 && options.Height > 0 {
		conf.ValidationMode = pdfcpu.ValidationRelaxed
	}

	ctxPdf := pdfcpu.NewContext(conf)
	pdfReader, err := pdfcpu.Read(bytes.NewReader(pdfData), ctxPdf)
	if err != nil {
		return nil, err
	}

	pageCount := len(pdfReader.PageTable.Pages)
	pages := make([]RenderPage, 0, pageCount)

	// 确定要渲染的页面
	var pagesToRender []int
	if len(options.PageNumbers) > 0 {
		pagesToRender = options.PageNumbers
	} else {
		for i := 1; i <= pageCount; i++ {
			pagesToRender = append(pagesToRender, i)
		}
	}

	// 渲染每页
	for _, pageNum := range pagesToRender {
		if pageNum < 1 || pageNum > pageCount {
			continue
		}

		// 渲染页面为图片
		imgData, err := e.renderPDFPageToImage(pdfData, pageNum, options)
		if err != nil {
			e.logger.WithError(err).Warnf("渲染第%d页失败", pageNum)
			continue
		}

		// 获取页面尺寸
		width, height := e.getPDFPageSize(pdfData, pageNum)
		if options.Width > 0 && options.Height > 0 {
			width = options.Width
			height = options.Height
		}

		page := RenderPage{
			PageNumber: pageNum,
			Width:      width,
			Height:     height,
			ImageData:  imgData,
			Metadata: map[string]interface{}{
				"render_dpi": 150, // 默认DPI
			},
		}

		pages = append(pages, page)
	}

	return pages, nil
}

// renderPDFPageToImage 渲染PDF页面为图片
func (e *RenderEngine) renderPDFPageToImage(pdfData []byte, pageNum int, options PreviewOptions) ([]byte, error) {
	// 这里应该实现PDF页面到图片的转换
	// 可以使用pdfcpu的image命令或者其他PDF渲染库

	// 简化实现，返回一个占位符图片
	width := options.Width
	height := options.Height
	if width <= 0 {
		width = 800
	}
	if height <= 0 {
		height = 600
	}

	// 创建一个简单的占位符图片
	img := imaging.New(width, height, color.NRGBA{R: 240, G: 240, B: 240, A: 255})

	// 添加页码文本
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: options.Quality})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// extractPDFTextWithPdfcpu 使用pdfcpu提取PDF文本
func (e *RenderEngine) extractPDFTextWithPdfcpu(pdfData []byte, pageNum int) (string, error) {
	// 使用pdfcpu提取文本
	conf := pdfcpu.NewDefaultConfiguration()
	ctxPdf := pdfcpu.NewContext(conf)

	pdfReader, err := pdfcpu.Read(bytes.NewReader(pdfData), ctxPdf)
	if err != nil {
		return "", err
	}

	if pageNum < 1 || pageNum > len(pdfReader.PageTable.Pages) {
		return "", fmt.Errorf("页码超出范围: %d", pageNum)
	}

	// 提取页面文本
	// 这里应该调用pdfcpu的文本提取功能
	// 简化实现，返回模拟文本
	return fmt.Sprintf("第%d页的文本内容...", pageNum), nil
}

// extractPDFLinksWithPdfcpu 使用pdfcpu提取PDF链接
func (e *RenderEngine) extractPDFLinksWithPdfcpu(pdfData []byte, pageNum int) ([]map[string]interface{}, error) {
	// 这里应该实现PDF链接提取
	// 简化实现，返回空结果
	return []map[string]interface{}{}, nil
}

// extractPDFAnnotationsWithPdfcpu 使用pdfcpu提取PDF注释
func (e *RenderEngine) extractPDFAnnotationsWithPdfcpu(pdfData []byte, pageNum int) ([]map[string]interface{}, error) {
	// 这里应该实现PDF注释提取
	// 简化实现，返回空结果
	return []map[string]interface{}{}, nil
}

// convertPDFLinksToRenderLinks 转换PDF链接为渲染链接
func (e *RenderEngine) convertPDFLinksToRenderLinks(links []map[string]interface{}) []RenderLink {
	renderLinks := make([]RenderLink, len(links))
	for i, link := range links {
		renderLinks[i] = RenderLink{
			URL:  getStringValue(link, "url"),
			Text: getStringValue(link, "text"),
			Page: getIntValue(link, "page"),
			Area: Rectangle{
				X:      getFloatValue(link, "x"),
				Y:      getFloatValue(link, "y"),
				Width:  getFloatValue(link, "width"),
				Height: getFloatValue(link, "height"),
			},
		}
	}
	return renderLinks
}

// convertPDFAnnotationsToRenderAnnotations 转换PDF注释为渲染注释
func (e *RenderEngine) convertPDFAnnotationsToRenderAnnotations(annotations []map[string]interface{}) []RenderAnnotation {
	renderAnnotations := make([]RenderAnnotation, len(annotations))
	for i, ann := range annotations {
		renderAnnotations[i] = RenderAnnotation{
			ID:      getStringValue(ann, "id"),
			Type:    getStringValue(ann, "type"),
			Content: getStringValue(ann, "content"),
			Color:   getStringValue(ann, "color"),
			Author:  getStringValue(ann, "author"),
			Position: Rectangle{
				X:      getFloatValue(ann, "x"),
				Y:      getFloatValue(ann, "y"),
				Width:  getFloatValue(ann, "width"),
				Height: getFloatValue(ann, "height"),
			},
			CreatedAt: time.Now(), // 应该从注释中提取实际时间
		}
	}
	return renderAnnotations
}

// convertWordToPDF 转换Word为PDF
func (e *RenderEngine) convertWordToPDF(wordData []byte) ([]byte, error) {
	// 这里应该实现Word到PDF的转换
	// 可以使用LibreOffice或其他转换工具
	// 简化实现，返回空字节数组
	return []byte{}, nil
}

// renderExcelWithUniOffice 使用unioffice渲染Excel
func (e *RenderEngine) renderExcelWithUniOffice(excelData []byte, options PreviewOptions) ([]RenderPage, error) {
	ss, err := spreadsheet.New(bytes.NewReader(excelData))
	if err != nil {
		return nil, err
	}
	defer ss.Close()

	var pages []RenderPage
	pageNum := 1

	for _, sheet := range ss.Sheets() {
		// 将工作表渲染为图片
		imgData := e.renderExcelSheetToImage(sheet, options)

		width := options.Width
		height := options.Height
		if width <= 0 {
			width = 1200
		}
		if height <= 0 {
			height = 800
		}

		// 提取工作表文本
		var sheetText strings.Builder
		for _, row := range sheet.Rows() {
			var rowData []string
			for _, cell := range row.Cells() {
				rowData = append(rowData, cell.GetText())
			}
			sheetText.WriteString(strings.Join(rowData, "\t") + "\n")
		}

		page := RenderPage{
			PageNumber:   pageNum,
			Width:        width,
			Height:       height,
			ImageData:    imgData,
			TextContent:  sheetText.String(),
			Annotations:  []RenderAnnotation{},
			Links:        []RenderLink{},
			Metadata: map[string]interface{}{
				"sheet_name":   sheet.Name(),
				"row_count":    len(sheet.Rows()),
				"column_count": e.getExcelColumnCount(sheet),
			},
		}

		pages = append(pages, page)
		pageNum++
	}

	return pages, nil
}

// renderPowerPointWithUniOffice 使用unioffice渲染PowerPoint
func (e *RenderEngine) renderPowerPointWithUniOffice(pptData []byte, options PreviewOptions) ([]RenderPage, error) {
	ppt, err := presentation.New(bytes.NewReader(pptData))
	if err != nil {
		return nil, err
	}
	defer ppt.Close()

	var pages []RenderPage

	for i, slide := range ppt.Slides() {
		// 将幻灯片渲染为图片
		imgData := e.renderPowerPointSlideToImage(slide, options)

		width := options.Width
		height := options.Height
		if width <= 0 {
			width = 1920
		}
		if height <= 0 {
			height = 1080
		}

		// 提取幻灯片文本
		var slideText strings.Builder
		for _, para := range slide.Paragraphs() {
			for _, run := range para.Runs() {
				slideText.WriteString(run.GetText() + " ")
			}
			slideText.WriteString("\n")
		}

		page := RenderPage{
			PageNumber:   i + 1,
			Width:        width,
			Height:       height,
			ImageData:    imgData,
			TextContent:  slideText.String(),
			Annotations:  []RenderAnnotation{},
			Links:        []RenderLink{},
			Metadata: map[string]interface{}{
				"slide_number": i + 1,
				"slide_title":  slide.Title(),
			},
		}

		pages = append(pages, page)
	}

	return pages, nil
}

// splitTextIntoPages 将文本分割为页面
func (e *RenderEngine) splitTextIntoPages(text string, options PreviewOptions) []string {
	// 简单的文本分页逻辑
	maxCharsPerPage := 2000 // 每页最大字符数
	lines := strings.Split(text, "\n")

	var pages []string
	var currentPage strings.Builder
	currentCharCount := 0

	for _, line := range lines {
		lineLength := len(line) + 1 // +1 for newline
		if currentCharCount+lineLength > maxCharsPerPage && currentPage.Len() > 0 {
			pages = append(pages, currentPage.String())
			currentPage.Reset()
			currentCharCount = 0
		}
		currentPage.WriteString(line + "\n")
		currentCharCount += lineLength
	}

	if currentPage.Len() > 0 {
		pages = append(pages, currentPage.String())
	}

	if len(pages) == 0 {
		pages = append(pages, text)
	}

	return pages
}

// renderTextToImage 将文本渲染为图片
func (e *RenderEngine) renderTextToImage(text string, options PreviewOptions) []byte {
	width := options.Width
	height := options.Height
	if width <= 0 {
		width = 800
	}
	if height <= 0 {
		height = 600
	}

	// 创建一个简单的占位符图片（白色背景）
	img := imaging.New(width, height, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	// 这里应该实现文本渲染逻辑
	// 可以使用freetype或其他文本渲染库

	// 编码为JPEG
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: options.Quality})
	if err != nil {
		return nil
	}

	return buf.Bytes()
}

// resizeImage 调整图片大小
func (e *RenderEngine) resizeImage(img image.Image, options PreviewOptions) image.Image {
	if options.Width <= 0 && options.Height <= 0 {
		return img // 不调整大小
	}

	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	var newWidth, newHeight int

	if options.Width > 0 && options.Height > 0 {
		// 指定了宽度和高度
		newWidth = options.Width
		newHeight = options.Height
	} else if options.Width > 0 {
		// 只指定宽度，按比例计算高度
		newWidth = options.Width
		newHeight = int(float64(originalHeight) * float64(options.Width) / float64(originalWidth))
	} else {
		// 只指定高度，按比例计算宽度
		newHeight = options.Height
		newWidth = int(float64(originalWidth) * float64(options.Height) / float64(originalHeight))
	}

	// 如果设置了缩放比例
	if options.Scale > 0 {
		newWidth = int(float64(newWidth) * options.Scale)
		newHeight = int(float64(newHeight) * options.Scale)
	}

	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

// 辅助函数

// getPDFPageCount 获取PDF页数
func (e *RenderEngine) getPDFPageCount(pdfData []byte) int {
	conf := pdfcpu.NewDefaultConfiguration()
	ctxPdf := pdfcpu.NewContext(conf)

	pdfReader, err := pdfcpu.Read(bytes.NewReader(pdfData), ctxPdf)
	if err != nil {
		return 0
	}

	return len(pdfReader.PageTable.Pages)
}

// getPDFPageSize 获取PDF页面尺寸
func (e *RenderEngine) getPDFPageSize(pdfData []byte, pageNum int) (int, int) {
	// 这里应该实现获取PDF页面尺寸的逻辑
	// 简化实现，返回默认尺寸
	return 800, 600
}

// renderExcelSheetToImage 渲染Excel工作表为图片
func (e *RenderEngine) renderExcelSheetToImage(sheet spreadsheet.Sheet, options PreviewOptions) []byte {
	width := options.Width
	height := options.Height
	if width <= 0 {
		width = 1200
	}
	if height <= 0 {
		height = 800
	}

	// 创建一个简单的占位符图片
	img := imaging.New(width, height, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: options.Quality})
	if err != nil {
		return nil
	}

	return buf.Bytes()
}

// renderPowerPointSlideToImage 渲染PowerPoint幻灯片为图片
func (e *RenderEngine) renderPowerPointSlideToImage(slide presentation.Slide, options PreviewOptions) []byte {
	width := options.Width
	height := options.Height
	if width <= 0 {
		width = 1920
	}
	if height <= 0 {
		height = 1080
	}

	// 创建一个简单的占位符图片
	img := imaging.New(width, height, color.NRGBA{R: 255, G: 255, B: 255, A: 255})

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: options.Quality})
	if err != nil {
		return nil
	}

	return buf.Bytes()
}

// getExcelColumnCount 获取Excel列数
func (e *RenderEngine) getExcelColumnCount(sheet spreadsheet.Sheet) int {
	maxCols := 0
	for _, row := range sheet.Rows() {
		if len(row.Cells()) > maxCols {
			maxCols = len(row.Cells())
		}
	}
	return maxCols
}

// extractExcelSheets 提取Excel工作表信息
func (e *RenderEngine) extractExcelSheets(excelData []byte) []map[string]interface{} {
	ss, err := spreadsheet.New(bytes.NewReader(excelData))
	if err != nil {
		return []map[string]interface{}{}
	}
	defer ss.Close()

	var sheets []map[string]interface{}
	for _, sheet := range ss.Sheets() {
		sheetInfo := map[string]interface{}{
			"name":        sheet.Name(),
			"row_count":   len(sheet.Rows()),
			"column_count": e.getExcelColumnCount(sheet),
		}
		sheets = append(sheets, sheetInfo)
	}

	return sheets
}

// extractPowerPointSlides 提取PowerPoint幻灯片信息
func (e *RenderEngine) extractPowerPointSlides(pptData []byte) []map[string]interface{} {
	ppt, err := presentation.New(bytes.NewReader(pptData))
	if err != nil {
		return []map[string]interface{}{}
	}
	defer ppt.Close()

	var slides []map[string]interface{}
	for i, slide := range ppt.Slides() {
		slideInfo := map[string]interface{}{
			"slide_number": i + 1,
			"title":        slide.Title(),
		}
		slides = append(slides, slideInfo)
	}

	return slides
}

// 工具函数

// getStringValue 从map中获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getIntValue 从map中获取整数值
func getIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(int); ok {
			return num
		}
		if str, ok := val.(string); ok {
			if num, err := strconv.Atoi(str); err == nil {
				return num
			}
		}
	}
	return 0
}

// getFloatValue 从map中获取浮点数值
func getFloatValue(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return num
		}
		if num, ok := val.(int); ok {
			return float64(num)
		}
		if str, ok := val.(string); ok {
			if num, err := strconv.ParseFloat(str, 64); err == nil {
				return num
			}
		}
	}
	return 0.0
}

// color 颜色类型（用于创建占位符图片）
type color struct {
	R, G, B, A uint8
}

func (c color) RGBA() (r, g, b, a uint32) {
	return uint32(c.R), uint32(c.G), uint32(c.B), uint32(c.A)
}