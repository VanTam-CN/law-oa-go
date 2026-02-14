package services

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"testing"
	"time"

	"law-oa-go/internal/errors"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDocumentServiceRepository Mock文档仓储
type MockDocumentServiceRepository struct {
	mock.Mock
}

func (m *MockDocumentServiceRepository) Create(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentServiceRepository) FindByID(ctx context.Context, id uint) (*models.Document, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Document), args.Error(1)
}

func (m *MockDocumentServiceRepository) Update(ctx context.Context, doc *models.Document) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentServiceRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDocumentServiceRepository) List(ctx context.Context, params *repositories.DocumentListParams) ([]*models.Document, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*models.Document), args.Get(1).(int64), args.Error(2)
}

func (m *MockDocumentServiceRepository) GetStats(ctx context.Context) (*repositories.DocumentStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.DocumentStats), args.Error(1)
}

// MockDocumentFileHeader Mock文件头
type MockDocumentFileHeader struct {
	Filename string
	Size     int64
	Header   map[string][]string
}

func (m *MockDocumentFileHeader) Open() (multipart.File, error) {
	return &MockDocumentFile{}, nil
}

// MockDocumentFile Mock文件
type MockDocumentFile struct {
	content string
	pos      int
}

func (m *MockDocumentFile) Read(p []byte) (int, error) {
	if m.pos >= len(m.content) {
		return 0, io.EOF
	}
	n := copy(p, m.content[m.pos:])
	m.pos += n
	return n, nil
}

func (m *MockFile) Close() error {
	return nil
}

// TestNewDocumentService 测试文档服务创建
func TestNewDocumentService(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.docRepo)
	assert.Equal(t, storageDir, service.storageDir)
}

// TestDocumentService_UploadDocument_Success 测试上传文档成功
func TestDocumentService_UploadDocument_Success(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	// 创建Mock文件
	mockFile := &MockDocumentFileHeader{
		Filename: "test.pdf",
		Size:     1024,
		Header:   map[string][]string{"Content-Type": {"application/pdf"}},
	}

	req := &DocumentUploadRequest{
		Name:        "测试文档",
		Description: "测试文档描述",
		Category:    "legal",
		Tags:        "标签1,标签2",
		EntityID:    1,
		EntityType:  "case",
		File:        mockFile,
	}

	// 设置Mock期望
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	// 执行测试
	result, err := service.UploadDocument(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "测试文档", result.Name)
	assert.Equal(t, "测试文档描述", result.Description)
	assert.Equal(t, "legal", result.Category)
	assert.Equal(t, []string{"标签1", "标签2"}, result.Tags)
	assert.Equal(t, uint(1), result.EntityID)
	assert.Equal(t, "case", result.EntityType)
	assert.Equal(t, "test.pdf", result.Filename)
	assert.Equal(t, int64(1024), result.Filesize)
	assert.Equal(t, "application/pdf", result.MimeType)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_UploadDocument_MissingFile 测试缺少文件
func TestDocumentService_UploadDocument_MissingFile(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	req := &DocumentUploadRequest{
		Name:        "测试文档",
		Description: "测试文档描述",
		File:        nil, // 缺少文件
	}

	// 执行测试
	result, err := service.UploadDocument(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "File is required")

	// 验证Mock调用
	mockRepo.AssertNotCalled(t, "Create")
}

// TestDocumentService_UploadDocument_FileTooLarge 测试文件过大
func TestDocumentService_UploadDocument_FileTooLarge(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	// 创建过大的Mock文件 (60MB)
	mockFile := &MockDocumentFileHeader{
		Filename: "large.pdf",
		Size:     60 * 1024 * 1024, // 60MB
		Header:   map[string][]string{"Content-Type": {"application/pdf"}},
	}

	req := &DocumentUploadRequest{
		Name: "大文件",
		File: mockFile,
	}

	// 执行测试
	result, err := service.UploadDocument(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "File too large")

	// 验证Mock调用
	mockRepo.AssertNotCalled(t, "Create")
}

// TestDocumentService_UploadDocument_UnsupportedFileType 测试不支持的文件类型
func TestDocumentService_UploadDocument_UnsupportedFileType(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	// 创建不支持的Mock文件
	mockFile := &MockDocumentFileHeader{
		Filename: "test.exe",
		Size:     1024,
		Header:   map[string][]string{"Content-Type": {"application/octet-stream"}},
	}

	req := &DocumentUploadRequest{
		Name: "可执行文件",
		File: mockFile,
	}

	// 执行测试
	result, err := service.UploadDocument(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Unsupported file type")

	// 验证Mock调用
	mockRepo.AssertNotCalled(t, "Create")
}

// TestDocumentService_GetDocumentByID_Success 测试根据ID获取文档成功
func TestDocumentService_GetDocumentByID_Success(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()
	docID := uint(1)

	expectedDoc := &models.Document{
		ID:          docID,
		Name:        "测试文档",
		Description: "测试描述",
		Filename:    "test.pdf",
		Filesize:    1024,
		MimeType:    "application/pdf",
		Category:    "legal",
		Tags:        "标签1,标签2",
		EntityID:    1,
		EntityType:  "case",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, docID).Return(expectedDoc, nil)

	// 执行测试
	result, err := service.GetDocumentByID(ctx, docID)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, docID, result.ID)
	assert.Equal(t, "测试文档", result.Name)
	assert.Equal(t, "测试描述", result.Description)
	assert.Equal(t, []string{"标签1", "标签2"}, result.Tags)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_GetDocumentByID_NotFound 测试文档不存在
func TestDocumentService_GetDocumentByID_NotFound(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()
	docID := uint(999)

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, docID).Return(nil, repositories.ErrDocumentNotFound)

	// 执行测试
	result, err := service.GetDocumentByID(ctx, docID)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Document not found")

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_ListDocuments_Success 测试获取文档列表成功
func TestDocumentService_ListDocuments_Success(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	req := &DocumentListRequest{
		Page:       1,
		PageSize:   10,
		Category:   "legal",
		EntityType: "case",
		Search:     "测试",
		SortBy:     "created_at",
		SortOrder:  "desc",
	}

	expectedDocs := []*models.Document{
		{
			ID:       1,
			Name:     "测试文档1",
			Category: "legal",
		},
		{
			ID:       2,
			Name:     "测试文档2",
			Category: "legal",
		},
	}

	// 设置Mock期望
	mockRepo.On("List", ctx, &repositories.DocumentListParams{
		Page:       1,
		PageSize:   10,
		Category:   "legal",
		EntityType: "case",
		Search:     "测试",
		SortBy:     "created_at",
		SortOrder:  "desc",
	}).Return(expectedDocs, int64(2), nil)

	// 执行测试
	result, total, err := service.ListDocuments(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, "测试文档1", result[0].Name)
	assert.Equal(t, "测试文档2", result[1].Name)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_UpdateDocument_Success 测试更新文档成功
func TestDocumentService_UpdateDocument_Success(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()
	docID := uint(1)

	existingDoc := &models.Document{
		ID:          docID,
		Name:        "原始名称",
		Description: "原始描述",
		Category:    "legal",
		Tags:        "标签1,标签2",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updateReq := &DocumentUpdateRequest{
		Name:        stringPtr("更新后的名称"),
		Description: stringPtr("更新后的描述"),
		Category:    stringPtr("contract"),
		Tags:        []string{"新标签1", "新标签2", "新标签3"},
	}

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, docID).Return(existingDoc, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	// 执行测试
	result, err := service.UpdateDocument(ctx, docID, updateReq)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "更新后的名称", result.Name)
	assert.Equal(t, "更新后的描述", result.Description)
	assert.Equal(t, "contract", result.Category)
	assert.Equal(t, []string{"新标签1", "新标签2", "新标签3"}, result.Tags)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_DeleteDocument_Success 测试删除文档成功
func TestDocumentService_DeleteDocument_Success(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()
	docID := uint(1)

	existingDoc := &models.Document{
		ID:     docID,
		Name:   "测试文档",
		Status: "active",
	}

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, docID).Return(existingDoc, nil)
	mockRepo.On("Delete", ctx, docID).Return(nil)

	// 执行测试
	err := service.DeleteDocument(ctx, docID)

	// 验证结果
	assert.NoError(t, err)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_GetDocumentStats_Success 测试获取文档统计成功
func TestDocumentService_GetDocumentStats_Success(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	expectedStats := &repositories.DocumentStats{
		Total: 100,
		ByCategory: []repositories.CategoryStat{
			{Category: "legal", Count: 50},
			{Category: "contract", Count: 30},
			{Category: "evidence", Count: 20},
		},
		ByEntityType: []repositories.EntityTypeStat{
			{EntityType: "case", Count: 60},
			{EntityType: "client", Count: 40},
		},
		RecentUploads: 5,
	}

	// 设置Mock期望
	mockRepo.On("GetStats", ctx).Return(expectedStats, nil)

	// 执行测试
	result, err := service.GetDocumentStats(ctx)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(100), result.TotalDocuments)
	assert.Equal(t, int64(50), result.ByCategory["legal"])
	assert.Equal(t, int64(30), result.ByCategory["contract"])
	assert.Equal(t, int64(20), result.ByCategory["evidence"])
	assert.Equal(t, int64(60), result.ByEntityType["case"])
	assert.Equal(t, int64(40), result.ByEntityType["client"])
	assert.Equal(t, int64(5), result.RecentUploads)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_DownloadDocument_Success 测试下载文档成功
func TestDocumentService_DownloadDocument_Success(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()
	docID := uint(1)

	existingDoc := &models.Document{
		ID:       docID,
		Name:     "测试文档",
		Filepath: "/tmp/documents/12345_test.pdf",
		Status:   "active",
	}

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, docID).Return(existingDoc, nil)

	// 执行测试
	reader, document, err := service.DownloadDocument(ctx, docID)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, document)
	assert.NotNil(t, reader) // 即使是Mock，也应该返回一个reader
	assert.Equal(t, docID, document.ID)
	assert.Equal(t, "测试文档", document.Name)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)

	// 清理
	if reader != nil {
		reader.Close()
	}
}

// TestDocumentService_toDocument 测试文档模型转换
func TestDocumentService_toDocument(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	model := &models.Document{
		ID:          1,
		Name:        "测试文档",
		Description: "测试描述",
		Filename:    "test.pdf",
		Filepath:    "/tmp/test.pdf",
		Filesize:    1024,
		MimeType:    "application/pdf",
		Category:    "legal",
		Tags:        "标签1,标签2, 标签3",
		EntityID:    1,
		EntityType:  "case",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 执行测试
	result := service.toDocument(model)

	// 验证结果
	assert.NotNil(t, result)
	assert.Equal(t, model.ID, result.ID)
	assert.Equal(t, model.Name, result.Name)
	assert.Equal(t, model.Description, result.Description)
	assert.Equal(t, model.Filename, result.Filename)
	assert.Equal(t, model.Filepath, result.Filepath)
	assert.Equal(t, model.Filesize, result.Filesize)
	assert.Equal(t, model.MimeType, result.MimeType)
	assert.Equal(t, model.Category, result.Category)
	assert.Equal(t, []string{"标签1", "标签2", "标签3"}, result.Tags)
	assert.Equal(t, model.EntityID, result.EntityID)
	assert.Equal(t, model.EntityType, result.EntityType)
	assert.Equal(t, model.CreatedAt, result.CreatedAt)
	assert.Equal(t, model.UpdatedAt, result.UpdatedAt)
}

// TestDocumentService_UploadDocument_DefaultName 测试使用默认名称
func TestDocumentService_UploadDocument_DefaultName(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	// 创建Mock文件，不提供名称
	mockFile := &MockDocumentFileHeader{
		Filename: "test-document.pdf",
		Size:     1024,
		Header:   map[string][]string{"Content-Type": {"application/pdf"}},
	}

	req := &DocumentUploadRequest{
		Name: "", // 空名称，应该使用文件名
		File: mockFile,
	}

	// 设置Mock期望
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	// 执行测试
	result, err := service.UploadDocument(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-document", result.Name) // 应该去掉扩展名

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_ListDocuments_DefaultParameters 测试默认参数
func TestDocumentService_ListDocuments_DefaultParameters(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	req := &DocumentListRequest{
		Page:     0,      // 应该设置为1
		PageSize: 0,      // 应该设置为20
	}

	expectedDocs := []*models.Document{}

	// 设置Mock期望
	mockRepo.On("List", ctx, &repositories.DocumentListParams{
		Page:     1,  // 默认值
		PageSize: 20, // 默认值
	}).Return(expectedDocs, int64(0), nil)

	// 执行测试
	result, total, err := service.ListDocuments(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.Len(t, result, 0)
	assert.Equal(t, int64(0), total)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestDocumentService_ListDocuments_MaxPageSize 测试最大页面大小限制
func TestDocumentService_ListDocuments_MaxPageSize(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	req := &DocumentListRequest{
		Page:     1,
		PageSize: 200, // 超过最大值100
	}

	expectedDocs := []*models.Document{}

	// 设置Mock期望
	mockRepo.On("List", ctx, &repositories.DocumentListParams{
		Page:     1,
		PageSize: 100, // 应该被限制为100
	}).Return(expectedDocs, int64(0), nil)

	// 执行测试
	result, total, err := service.ListDocuments(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.Len(t, result, 0)
	assert.Equal(t, int64(0), total)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// BenchmarkDocumentService_UploadDocument 基准测试上传文档性能
func BenchmarkDocumentService_UploadDocument(b *testing.B) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	mockFile := &MockDocumentFileHeader{
		Filename: "benchmark.pdf",
		Size:     1024,
		Header:   map[string][]string{"Content-Type": {"application/pdf"}},
	}

	req := &DocumentUploadRequest{
		Name: "基准测试文档",
		File: mockFile,
	}

	// 设置Mock期望
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.UploadDocument(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestDocumentService_Integration_CompleteWorkflow 测试文档服务完整工作流
func TestDocumentService_Integration_CompleteWorkflow(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	// 1. 上传文档
	mockFile := &MockDocumentFileHeader{
		Filename: "workflow_test.pdf",
		Size:     1024,
		Header:   map[string][]string{"Content-Type": {"application/pdf"}},
	}

	uploadReq := &DocumentUploadRequest{
		Name:        "工作流测试文档",
		Description: "完整工作流测试",
		Category:    "legal",
		Tags:        "测试,工作流",
		EntityID:    1,
		EntityType:  "case",
		File:        mockFile,
	}

	createdDoc := &models.Document{
		ID:          1,
		Name:        uploadReq.Name,
		Description: uploadReq.Description,
		Category:    uploadReq.Category,
		Tags:        "测试,工作流",
		EntityID:    uploadReq.EntityID,
		EntityType:  uploadReq.EntityType,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 2. 获取文档
	getDoc := &models.Document{
		ID:          1,
		Name:        uploadReq.Name,
		Description: uploadReq.Description,
		Category:    uploadReq.Category,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 3. 更新文档
	updatedDoc := &models.Document{
		ID:          1,
		Name:        "更新后的工作流测试文档",
		Description: "更新后的描述",
		Category:    "contract",
		Tags:        "测试,工作流,更新",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 设置Mock期望 - 上传
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	// 设置Mock期望 - 获取
	mockRepo.On("FindByID", ctx, uint(1)).Return(getDoc, nil)

	// 设置Mock期望 - 更新
	mockRepo.On("FindByID", ctx, uint(1)).Return(getDoc, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Document")).Return(nil)

	// 设置Mock期望 - 删除
	mockRepo.On("FindByID", ctx, uint(1)).Return(updatedDoc, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	// 执行工作流测试

	// 1. 上传文档
	uploaded, err := service.UploadDocument(ctx, uploadReq)
	require.NoError(t, err)
	assert.NotNil(t, uploaded)

	// 2. 获取文档
	fetched, err := service.GetDocumentByID(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, fetched)

	// 3. 更新文档
	updateReq := &DocumentUpdateRequest{
		Name:        stringPtr("更新后的工作流测试文档"),
		Description: stringPtr("更新后的描述"),
		Category:    stringPtr("contract"),
		Tags:        []string{"测试", "工作流", "更新"},
	}
	updated, err := service.UpdateDocument(ctx, 1, updateReq)
	require.NoError(t, err)
	assert.NotNil(t, updated)

	// 4. 删除文档
	err = service.DeleteDocument(ctx, 1)
	assert.NoError(t, err)

	// 验证所有Mock调用
	mockRepo.AssertExpectations(t)
}


// TestDocumentService_UploadDocument_AllowedFileTypes 测试允许的文件类型
func TestDocumentService_UploadDocument_AllowedFileTypes(t *testing.T) {
	mockRepo := &MockDocumentServiceRepository{}
	storageDir := "/tmp/documents"
	service := NewDocumentService(mockRepo, storageDir)

	ctx := context.Background()

	allowedTypes := []struct {
		filename string
		mimeType string
	}{
		{"test.txt", "text/plain"},
		{"test.csv", "text/csv"},
		{"test.pdf", "application/pdf"},
		{"test.doc", "application/msword"},
		{"test.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"test.xls", "application/vnd.ms-excel"},
		{"test.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"test.jpg", "image/jpeg"},
		{"test.png", "image/png"},
		{"test.gif", "image/gif"},
	}

	for _, fileType := range allowedTypes {
		t.Run(fileType.mimeType, func(t *testing.T) {
			mockFile := &MockDocumentFileHeader{
				Filename: fileType.filename,
				Size:     1024,
				Header:   map[string][]string{"Content-Type": {fileType.mimeType}},
			}

			req := &DocumentUploadRequest{
				Name: "测试文档",
				File: mockFile,
			}

			// 设置Mock期望
			mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Document")).Return(nil).Once()

			// 执行测试
			result, err := service.UploadDocument(ctx, req)

			// 验证结果
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, fileType.mimeType, result.MimeType)
		})
	}

	// 验证所有Mock调用
	mockRepo.AssertExpectations(t)
}