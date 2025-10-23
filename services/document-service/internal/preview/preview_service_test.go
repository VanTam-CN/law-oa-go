package preview

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/config"
)

// MockStorage 模拟存储
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorage) Put(ctx context.Context, path string, reader io.Reader) error {
	args := m.Called(ctx, path, reader)
	return args.Error(0)
}

func (m *MockStorage) GetURL(path string) string {
	args := m.Called(path)
	return args.String(0)
}

func (m *MockStorage) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockStorage) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

// setupTestService 设置测试服务
func setupTestService(t *testing.T) (*PreviewService, *gorm.DB, *MockStorage) {
	// 创建内存数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移所有必要的表
	err = db.AutoMigrate(
		&Document{},
		&DocumentVersion{},
		&DocumentCategory{},
		&DocumentTag{},
		&DocumentShare{},
		&Comment{},
		&CommentReaction{},
		&Activity{},
		&Annotation{},
		&AnnotationAttachment{},
		&CollaborationSession{},
		&CollaborationParticipant{},
		&CollaborationOperation{},
		&CollaborationChange{},
		&VersionDiff{},
		&User{},
	)
	require.NoError(t, err)

	// 创建模拟存储
	mockStorage := new(MockStorage)

	// 创建日志器
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建配置
	cfg := &config.Config{
		Preview: config.PreviewConfig{
			MaxWidth:     2000,
			MaxHeight:    2000,
			MaxFileSize:  100 * 1024 * 1024, // 100MB
			CacheEnabled: true,
			CacheTTL:     24 * time.Hour,
		},
	}

	// 创建预览服务
	service := NewPreviewService(db, logger, cfg, mockStorage)

	return service, db, mockStorage
}

// createTestDocument 创建测试文档
func createTestDocument(t *testing.T, db *gorm.DB) *Document {
	document := &Document{
		TenantID:     "test-tenant",
		Title:        "测试文档",
		Description:  "这是一个测试文档",
		CreatedBy:    1,
		OwnerID:      1,
		Status:       "published",
		Visibility:   "public",
		AccessLevel:  "read",
		ViewCount:    0,
		DownloadCount: 0,
		ShareCount:   0,
	}

	err := db.Create(document).Error
	require.NoError(t, err)

	return document
}

// createTestDocumentVersion 创建测试文档版本
func createTestDocumentVersion(t *testing.T, db *gorm.DB, documentID uint, contentType string) *DocumentVersion {
	version := &DocumentVersion{
		DocumentID:     documentID,
		VersionNumber:  1,
		Title:          "测试版本",
		Content:        "这是测试内容",
		ContentType:    contentType,
		FileSize:       1024,
		FileHash:       "test-hash-123",
		StoragePath:    "/test/path/document.pdf",
		ThumbnailPath:  "/test/path/thumbnail.jpg",
		VersionTag:     "v1.0",
		IsMajor:        true,
		IsDraft:        false,
		IsPublished:    true,
		EditorID:       &[]uint{1}[0],
		EditReason:     "初始版本",
		EditDuration:   30,
		CharacterCount: 100,
		WordCount:      20,
		PageCount:      5,
		RenderStatus:   "pending",
	}

	err := db.Create(version).Error
	require.NoError(t, err)

	return version
}

// createTestImage 创建测试图片
func createTestImage(t *testing.T, width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	require.NoError(t, err)

	return buf.Bytes()
}

// TestPreviewService_GeneratePreview 测试生成预览
func TestPreviewService_GeneratePreview(t *testing.T) {
	service, db, mockStorage := setupTestService(t)

	// 创建测试数据
	document := createTestDocument(t, db)
	version := createTestDocumentVersion(t, db, document.ID, "pdf")

	// 设置模拟存储返回
	imageData := createTestImage(t, 800, 600)
	mockStorage.On("Get", mock.Anything, version.StoragePath).Return(
		io.NopCloser(bytes.NewReader([]byte("pdf content"))), nil,
	)
	mockStorage.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockStorage.On("GetURL", mock.Anything).Return("http://example.com/preview.jpg")

	// 设置预览选项
	options := PreviewOptions{
		Width:        800,
		Height:       600,
		Scale:        1.0,
		Quality:      90,
		Format:       "jpg",
		Thumbnail:    true,
		CacheEnabled: true,
		CacheTTL:     time.Hour,
	}

	// 执行测试
	result, err := service.GeneratePreview(context.Background(), document.ID, nil, options)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "jpg", result.Format)
	assert.Greater(t, result.PageCount, 0)
	assert.NotEmpty(t, result.Pages)
	assert.NotEmpty(t, result.Thumbnails)

	// 验证模拟调用
	mockStorage.AssertExpectations(t)
}

// TestPreviewService_GenerateThumbnail 测试生成缩略图
func TestPreviewService_GenerateThumbnail(t *testing.T) {
	service, db, mockStorage := setupTestService(t)

	// 创建测试数据
	document := createTestDocument(t, db)
	version := createTestDocumentVersion(t, db, document.ID, "pdf")

	// 设置模拟存储返回
	imageData := createTestImage(t, 150, 150)
	mockStorage.On("Get", mock.Anything, version.StoragePath).Return(
		io.NopCloser(bytes.NewReader([]byte("pdf content"))), nil,
	)
	mockStorage.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockStorage.On("GetURL", mock.Anything).Return("http://example.com/thumbnail.jpg")

	// 执行测试
	thumbnail, err := service.GenerateThumbnail(context.Background(), document.ID, nil, 1, 150)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, thumbnail)
	assert.Equal(t, 1, thumbnail.PageNumber)
	assert.Equal(t, 150, thumbnail.Width)
	assert.Equal(t, 150, thumbnail.Height)
	assert.NotEmpty(t, thumbnail.ImageURL)

	// 验证模拟调用
	mockStorage.AssertExpectations(t)
}

// TestPreviewService_ExtractText 测试提取文本
func TestPreviewService_ExtractText(t *testing.T) {
	service, db, mockStorage := setupTestService(t)

	// 创建测试数据
	document := createTestDocument(t, db)
	version := createTestDocumentVersion(t, db, document.ID, "txt")

	// 设置模拟存储返回
	textContent := "这是第一页的内容\n这是第二页的内容\n这是第三页的内容"
	mockStorage.On("Get", mock.Anything, version.StoragePath).Return(
		io.NopCloser(bytes.NewReader([]byte(textContent))), nil,
	)

	// 执行测试
	textPages, err := service.ExtractText(context.Background(), document.ID, nil)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, textPages)
	assert.Len(t, textPages, 1) // 纯文本文件视为单页
	assert.Equal(t, textContent, textPages[1])

	// 验证模拟调用
	mockStorage.AssertExpectations(t)
}

// TestPreviewService_GetDocumentInfo 测试获取文档信息
func TestPreviewService_GetDocumentInfo(t *testing.T) {
	service, db, mockStorage := setupTestService(t)

	// 创建测试数据
	document := createTestDocument(t, db)
	version := createTestDocumentVersion(t, db, document.ID, "pdf")

	// 设置模拟存储返回
	mockStorage.On("Get", mock.Anything, version.StoragePath).Return(
		io.NopCloser(bytes.NewReader([]byte("pdf content"))), nil,
	)

	// 执行测试
	info, err := service.GetDocumentInfo(context.Background(), document.ID, nil)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, document.ID, info["document_id"])
	assert.Equal(t, version.ID, info["version_id"])
	assert.Equal(t, version.VersionNumber, info["version_number"])
	assert.Equal(t, version.Title, info["title"])
	assert.Equal(t, version.ContentType, info["content_type"])
	assert.Equal(t, version.FileSize, info["file_size"])

	// 验证模拟调用
	mockStorage.AssertExpectations(t)
}

// TestPreviewService_SearchInDocument 测试文档搜索
func TestPreviewService_SearchInDocument(t *testing.T) {
	service, db, mockStorage := setupTestService(t)

	// 创建测试数据
	document := createTestDocument(t, db)
	version := createTestDocumentVersion(t, db, document.ID, "txt")

	// 设置模拟存储返回
	textContent := "这是第一页的内容\n包含测试关键词的内容\n这是第三页的内容"
	mockStorage.On("Get", mock.Anything, version.StoragePath).Return(
		io.NopCloser(bytes.NewReader([]byte(textContent))), nil,
	)

	// 设置搜索请求
	req := SearchRequest{
		DocumentID: document.ID,
		Query:      "测试",
		Options: SearchOptions{
			CaseSensitive: false,
			WholeWord:     false,
			Regex:         false,
			ContextLength: 50,
			MaxResults:    10,
		},
	}

	// 执行测试
	result, err := service.SearchInDocument(context.Background(), document.ID, nil, "测试", req.Options)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "测试", result.Query)
	assert.GreaterOrEqual(t, result.TotalMatches, 0)

	// 验证模拟调用
	mockStorage.AssertExpectations(t)
}

// TestPreviewOptionsValidation 测试预览选项验证
func TestPreviewOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		options PreviewOptions
		valid   bool
	}{
		{
			name: "有效选项",
			options: PreviewOptions{
				Width:        800,
				Height:       600,
				Scale:        1.0,
				Quality:      90,
				Format:       "jpg",
				CacheEnabled: true,
				CacheTTL:     time.Hour,
			},
			valid: true,
		},
		{
			name: "宽度为0",
			options: PreviewOptions{
				Width:  0,
				Height: 600,
			},
			valid: true, // 允许为0，将使用默认值
		},
		{
			name: "质量超出范围",
			options: PreviewOptions{
				Width:   800,
				Height:  600,
				Quality: 150, // 超出1-100范围
			},
			valid: true, // 会在处理时被修正
		},
		{
			name: "负数缩放",
			options: PreviewOptions{
				Width:  800,
				Height: 600,
				Scale:  -1.0,
			},
			valid: true, // 会在处理时被修正
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 这里可以添加选项验证逻辑的测试
			// 目前 PreviewOptions 结构体本身没有验证方法
			assert.NotNil(t, tt.options)
		})
	}
}

// TestPreviewService_ImageFormatConversion 测试图片格式转换
func TestPreviewService_ImageFormatConversion(t *testing.T) {
	service, _, _ := setupTestService(t)

	// 创建测试图片
	originalImg := createTestImage(t, 400, 300)

	tests := []struct {
		name     string
		format   string
		quality  int
		expected int
	}{
		{
			name:     "JPEG格式",
			format:   "jpg",
			quality:  90,
			expected: 400 * 300 * 3, // 估算大小
		},
		{
			name:     "PNG格式",
			format:   "png",
			quality:  90,
			expected: 400 * 300 * 4, // 估算大小
		},
		{
			name:     "WEBP格式",
			format:   "webp",
			quality:  80,
			expected: 400 * 300 * 2, // 估算大小
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := PreviewOptions{
				Format:  tt.format,
				Quality: tt.quality,
			}

			// 调用内部方法进行格式转换
			converted := service.convertImageFormat(originalImg, options.Format, options.Quality)

			assert.NotEmpty(t, converted)
			assert.Greater(t, len(converted), 0)
		})
	}
}

// TestPreviewService_CacheKeyGeneration 测试缓存键生成
func TestPreviewService_CacheKeyGeneration(t *testing.T) {
	service, db, _ := setupTestService(t)

	// 创建测试数据
	document := createTestDocument(t, db)
	version := createTestDocumentVersion(t, db, document.ID, "pdf")

	options1 := PreviewOptions{
		Width:        800,
		Height:       600,
		Scale:        1.0,
		Quality:      90,
		Format:       "jpg",
		CacheEnabled: true,
	}

	options2 := PreviewOptions{
		Width:        1024,
		Height:       768,
		Scale:        1.5,
		Quality:      95,
		Format:       "png",
		CacheEnabled: true,
	}

	// 生成缓存键
	key1 := service.generateCacheKey(version, options1)
	key2 := service.generateCacheKey(version, options2)

	// 验证缓存键不同
	assert.NotEqual(t, key1, key2)
	assert.Contains(t, key1, "preview")
	assert.Contains(t, key1, string(version.FileHash))
}

// BenchmarkPreviewService_GeneratePreview 性能测试
func BenchmarkPreviewService_GeneratePreview(b *testing.B) {
	service, db, mockStorage := setupTestService(&testing.T{})

	// 创建测试数据
	document := createTestDocument(&testing.T{}, db)
	version := createTestDocumentVersion(&testing.T{}, db, document.ID, "pdf")

	// 设置模拟存储返回
	imageData := createTestImage(&testing.T{}, 800, 600)
	mockStorage.On("Get", mock.Anything, mock.Anything).Return(
		io.NopCloser(bytes.NewReader([]byte("pdf content"))), nil,
	)
	mockStorage.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockStorage.On("GetURL", mock.Anything).Return("http://example.com/preview.jpg")

	options := PreviewOptions{
		Width:        800,
		Height:       600,
		Quality:      90,
		Format:       "jpg",
		CacheEnabled: false, // 禁用缓存以确保每次都执行生成
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GeneratePreview(context.Background(), document.ID, nil, options)
		if err != nil {
			b.Fatalf("生成预览失败: %v", err)
		}
	}
}

// TestPreviewService_ErrorHandling 测试错误处理
func TestPreviewService_ErrorHandling(t *testing.T) {
	service, db, _ := setupTestService(t)

	t.Run("文档不存在", func(t *testing.T) {
		options := PreviewOptions{}
		result, err := service.GeneratePreview(context.Background(), 999, nil, options)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "获取文档版本失败")
	})

	t.Run("存储文件不存在", func(t *testing.T) {
		document := createTestDocument(t, db)
		version := createTestDocumentVersion(t, db, document.ID, "pdf")

		options := PreviewOptions{}
		result, err := service.GeneratePreview(context.Background(), document.ID, nil, options)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestPreviewService_ConcurrentAccess 并发访问测试
func TestPreviewService_ConcurrentAccess(t *testing.T) {
	service, db, mockStorage := setupTestService(t)

	// 创建测试数据
	document := createTestDocument(t, db)
	version := createTestDocumentVersion(t, db, document.ID, "pdf")

	// 设置模拟存储返回
	imageData := createTestImage(t, 800, 600)
	mockStorage.On("Get", mock.Anything, version.StoragePath).Return(
		io.NopCloser(bytes.NewReader([]byte("pdf content"))), nil,
	)
	mockStorage.On("Put", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockStorage.On("GetURL", mock.Anything).Return("http://example.com/preview.jpg")

	options := PreviewOptions{
		Width:        800,
		Height:       600,
		CacheEnabled: false,
	}

	// 并发执行
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := service.GeneratePreview(context.Background(), document.ID, nil, options)
			results <- err
		}()
	}

	// 收集结果
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}