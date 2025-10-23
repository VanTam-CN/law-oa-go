package services

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
	"time"

	"law-oa-go/internal/models"
)

// TestDocumentService 测试文档上传服务
func TestDocumentService(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "doc_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建模拟仓库和文档服务
	mockRepo := &MockDocumentRepository{}
	docService := NewDocumentService(mockRepo, tempDir)

	t.Run("UploadDocument", func(t *testing.T) {
		// 创建模拟文件
		content := []byte("This is a test document")
		file := &MockFile{content: content, filename: "test.txt"}

		req := &UploadRequest{
			Name:        "Test Document",
			Category:    "Test",
			Description: "Test upload",
			File:        file,
			UploadedBy:  1,
		}

		result, err := docService.UploadDocument(context.Background(), req)
		if err != nil {
			t.Errorf("UploadDocument failed: %v", err)
		}

		if result == nil {
			t.Error("UploadDocument returned nil result")
		}

		// 验证文件是否保存到磁盘
		filePath := filepath.Join(tempDir, result.Filepath)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File was not saved to disk: %s", filePath)
		}
	})

	t.Run("GetDocument", func(t *testing.T) {
		doc, err := docService.GetDocument(context.Background(), 1)
		if err != nil {
			t.Errorf("GetDocument failed: %v", err)
		}

		if doc == nil {
			t.Error("GetDocument returned nil document")
		}
	})

	t.Run("ListDocuments", func(t *testing.T) {
		docs, total, err := docService.ListDocuments(context.Background(), &ListRequest{
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			t.Errorf("ListDocuments failed: %v", err)
		}

		if len(docs) == 0 {
			t.Error("ListDocuments returned empty list")
		}

		if total == 0 {
			t.Error("ListDocuments returned zero total")
		}
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		req := &UpdateRequest{
			DocumentID:   1,
			Name:         "Updated Document",
			Category:     "Updated",
			Description:  "Updated description",
			Tags:         "updated,test",
			UpdatedBy:    1,
		}

		err := docService.UpdateDocument(context.Background(), req)
		if err != nil {
			t.Errorf("UpdateDocument failed: %v", err)
		}
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		err := docService.DeleteDocument(context.Background(), 1, 1)
		if err != nil {
			t.Errorf("DeleteDocument failed: %v", err)
		}
	})
}

// TestDocumentPreviewService 测试文档预览服务
func TestDocumentPreviewService(t *testing.T) {
	mockRepo := &MockDocumentRepository{}
	previewService := NewDocumentPreviewService(mockRepo)

	t.Run("GetDocumentPreview", func(t *testing.T) {
		req := &PreviewRequest{
			DocumentID: 1,
			MaxWidth:   800,
			MaxHeight:  600,
		}

		result, err := previewService.GetDocumentPreview(context.Background(), req)
		if err != nil {
			t.Errorf("GetDocumentPreview failed: %v", err)
		}

		if result == nil {
			t.Error("GetDocumentPreview returned nil result")
		}

		// 验证预览内容不为空
		if result.Content == "" {
			t.Error("Preview content is empty")
		}

		// 验证文档类型
		if result.MimeType == "" {
			t.Error("Preview MIME type is empty")
		}
	})

	t.Run("GetDocumentPreview_NonExistent", func(t *testing.T) {
		req := &PreviewRequest{
			DocumentID: 999, // 不存在的文档ID
		}

		_, err := previewService.GetDocumentPreview(context.Background(), req)
		if err == nil {
			t.Error("Expected error for non-existent document, got nil")
		}
	})
}

// TestDocumentVersionService 测试文档版本管理
func TestDocumentVersionService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "version_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockRepo := &MockDocumentRepository{}
	versionService := NewDocumentVersionService(mockRepo, tempDir)

	t.Run("CreateVersion", func(t *testing.T) {
		content := []byte("This is version 2 of the document")
		file := &MockFile{content: content, filename: "v2.txt"}

		req := &CreateVersionRequest{
			DocumentID:  1,
			Name:        "Version 2",
			Description: "Updated version",
			File:        file,
			CreatedBy:   1,
			Changes:     "Updated content",
		}

		result, err := versionService.CreateVersion(context.Background(), req)
		if err != nil {
			t.Errorf("CreateVersion failed: %v", err)
		}

		if result == nil {
			t.Error("CreateVersion returned nil result")
		}

		// 验证版本号
		if result.Version <= 0 {
			t.Error("Invalid version number")
		}
	})

	t.Run("GetVersions", func(t *testing.T) {
		req := &GetVersionsRequest{
			DocumentID: 1,
			Page:       1,
			PageSize:   10,
		}

		result, err := versionService.GetVersions(context.Background(), req)
		if err != nil {
			t.Errorf("GetVersions failed: %v", err)
		}

		if len(result.Versions) == 0 {
			t.Error("GetVersions returned empty list")
		}
	})

	t.Run("GetVersion", func(t *testing.T) {
		version, err := versionService.GetVersion(context.Background(), 1, 1)
		if err != nil {
			t.Errorf("GetVersion failed: %v", err)
		}

		if version == nil {
			t.Error("GetVersion returned nil version")
		}
	})

	t.Run("CompareVersions", func(t *testing.T) {
		req := &CompareVersionsRequest{
			DocumentID:   1,
			FromVersion: 1,
			ToVersion:   2,
		}

		result, err := versionService.CompareVersions(context.Background(), req)
		if err != nil {
			t.Errorf("CompareVersions failed: %v", err)
		}

		if result == nil {
			t.Error("CompareVersions returned nil result")
		}

		// 验证比较内容
		if result.FromVersion == nil || result.ToVersion == nil {
			t.Error("Missing version information in comparison")
		}
	})

	t.Run("RestoreVersion", func(t *testing.T) {
		req := &RestoreVersionRequest{
			DocumentID: 1,
			Version:    1,
			RestoredBy: 1,
		}

		err := versionService.RestoreVersion(context.Background(), req)
		if err != nil {
			t.Errorf("RestoreVersion failed: %v", err)
		}
	})
}

// TestDocumentPermissionService 测试文档权限控制
func TestDocumentPermissionService(t *testing.T) {
	mockRepo := &MockDocumentRepository{}
	mockUserRepo := &MockUserRepository{}
	permissionService := NewDocumentPermissionService(mockRepo, mockUserRepo)

	t.Run("GrantPermission", func(t *testing.T) {
		req := &PermissionRequest{
			DocumentID: 1,
			UserID:     2,
			Permission: "read",
			GrantedBy:  1,
		}

		result, err := permissionService.GrantPermission(context.Background(), req)
		if err != nil {
			t.Errorf("GrantPermission failed: %v", err)
		}

		if result == nil {
			t.Error("GrantPermission returned nil result")
		}

		// 验证权限信息
		if result.DocumentID != 1 || result.UserID != 2 {
			t.Error("Permission information mismatch")
		}
	})

	t.Run("CheckPermission", func(t *testing.T) {
		req := &CheckPermissionRequest{
			DocumentID: 1,
			UserID:     2,
			Permission: "read",
		}

		hasPermission, err := permissionService.CheckPermission(context.Background(), req)
		if err != nil {
			t.Errorf("CheckPermission failed: %v", err)
		}

		if !hasPermission {
			t.Error("Expected user to have permission")
		}
	})

	t.Run("GetDocumentPermissions", func(t *testing.T) {
		req := &GetPermissionsRequest{
			DocumentID: 1,
		}

		result, err := permissionService.GetDocumentPermissions(context.Background(), req)
		if err != nil {
			t.Errorf("GetDocumentPermissions failed: %v", err)
		}

		if len(result.Permissions) == 0 {
			t.Error("GetDocumentPermissions returned empty list")
		}
	})

	t.Run("ShareDocument", func(t *testing.T) {
		req := &ShareDocumentRequest{
			DocumentID: 1,
			ShareBy:    1,
			Users: []UserPermission{
				{UserID: 2, Permission: "read"},
				{UserID: 3, Permission: "write"},
			},
		}

		result, err := permissionService.ShareDocument(context.Background(), req)
		if err != nil {
			t.Errorf("ShareDocument failed: %v", err)
		}

		if result == nil {
			t.Error("ShareDocument returned nil result")
		}

		// 验证分享用户数量
		if len(result.SharedWith) != 2 {
			t.Error("Shared user count mismatch")
		}
	})

	t.Run("RevokePermission", func(t *testing.T) {
		req := &RevokePermissionRequest{
			DocumentID: 1,
			UserID:     2,
			RevokedBy:  1,
		}

		err := permissionService.RevokePermission(context.Background(), req)
		if err != nil {
			t.Errorf("RevokePermission failed: %v", err)
		}
	})
}

// TestDocumentRecycleService 测试文档回收站
func TestDocumentRecycleService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "recycle_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockRepo := &MockDocumentRepository{}
	recycleService := NewDocumentRecycleService(mockRepo, tempDir)

	t.Run("SoftDeleteDocument", func(t *testing.T) {
		req := &SoftDeleteRequest{
			DocumentID: 1,
			DeletedBy:  1,
		}

		result, err := recycleService.SoftDelete(context.Background(), req)
		if err != nil {
			t.Errorf("SoftDeleteDocument failed: %v", err)
		}

		if result == nil {
			t.Error("SoftDeleteDocument returned nil result")
		}

		// 验证删除状态
		if !result.Success {
			t.Error("Soft delete failed")
		}
	})

	t.Run("GetRecycleBin", func(t *testing.T) {
		req := &GetRecycleBinRequest{
			Page:     1,
			PageSize: 10,
		}

		result, err := recycleService.GetRecycleBin(context.Background(), req)
		if err != nil {
			t.Errorf("GetRecycleBin failed: %v", err)
		}

		if len(result.DeletedDocuments) == 0 {
			t.Error("GetRecycleBin returned empty list")
		}
	})

	t.Run("RestoreDocuments", func(t *testing.T) {
		req := &RestoreRequest{
			DocumentIDs: []uint{1},
			RestoredBy:  1,
		}

		result, err := recycleService.RestoreDocuments(context.Background(), req)
		if err != nil {
			t.Errorf("RestoreDocuments failed: %v", err)
		}

		if len(result.Restored) == 0 {
			t.Error("No documents were restored")
		}
	})

	t.Run("PermanentlyDelete", func(t *testing.T) {
		req := &PermanentlyDeleteRequest{
			DocumentIDs: []uint{1},
			DeletedBy:   1,
			Confirm:     true,
		}

		result, err := recycleService.PermanentlyDelete(context.Background(), req)
		if err != nil {
			t.Errorf("PermanentlyDelete failed: %v", err)
		}

		if len(result.Deleted) == 0 {
			t.Error("No documents were permanently deleted")
		}
	})
}

// TestDocumentSearchService 测试文档搜索功能
func TestDocumentSearchService(t *testing.T) {
	mockRepo := &MockDocumentRepository{}
	searchService := NewDocumentSearchService(mockRepo)

	t.Run("SearchDocuments", func(t *testing.T) {
		req := &DocumentSearchRequest{
			Query:    "test",
			Category: "legal",
			Page:     1,
			PageSize: 10,
		}

		result, err := searchService.SearchDocuments(context.Background(), req)
		if err != nil {
			t.Errorf("SearchDocuments failed: %v", err)
		}

		if result == nil {
			t.Error("SearchDocuments returned nil result")
		}

		// 验证搜索结果
		if len(result.Documents) == 0 {
			t.Error("Search returned no documents")
		}

		if result.TotalCount == 0 {
			t.Error("Search returned zero total count")
		}
	})

	t.Run("AdvancedSearch", func(t *testing.T) {
		req := &AdvancedSearchRequest{
			Query:      "legal contract",
			Category:   "legal",
			EntityType: "case",
			Tags:       []string{"important", "urgent"},
			DateFrom:   "2024-01-01",
			DateTo:     "2024-12-31",
			Page:       1,
			PageSize:   10,
		}

		result, err := searchService.AdvancedSearch(context.Background(), req)
		if err != nil {
			t.Errorf("AdvancedSearch failed: %v", err)
		}

		if result == nil {
			t.Error("AdvancedSearch returned nil result")
		}
	})

	t.Run("GetSearchFilters", func(t *testing.T) {
		result, err := searchService.GetSearchFilters(context.Background())
		if err != nil {
			t.Errorf("GetSearchFilters failed: %v", err)
		}

		if result == nil {
			t.Error("GetSearchFilters returned nil result")
		}

		// 验证过滤器
		if len(result.Categories) == 0 {
			t.Error("No category filters available")
		}
	})

	t.Run("GetSearchSuggestions", func(t *testing.T) {
		req := &SearchSuggestionRequest{
			Query: "legal",
		}

		result, err := searchService.GetSearchSuggestions(context.Background(), req)
		if err != nil {
			t.Errorf("GetSearchSuggestions failed: %v", err)
		}

		if result == nil {
			t.Error("GetSearchSuggestions returned nil result")
		}

		// 验证建议
		if len(result.Suggestions) == 0 {
			t.Error("No search suggestions available")
		}
	})
}

// TestDocumentStatsService 测试文档统计功能
func TestDocumentStatsService(t *testing.T) {
	mockRepo := &MockDocumentRepository{}
	statsService := NewDocumentStatsService(mockRepo)

	t.Run("GetDocumentOverview", func(t *testing.T) {
		result, err := statsService.GetDocumentOverview(context.Background())
		if err != nil {
			t.Errorf("GetDocumentOverview failed: %v", err)
		}

		if result == nil {
			t.Error("GetDocumentOverview returned nil result")
		}

		// 验证统计数据
		if result.TotalDocuments == 0 {
			t.Error("Total documents count is zero")
		}

		if len(result.DocumentsByCategory) == 0 {
			t.Error("No category statistics available")
		}
	})

	t.Run("GetStorageUsage", func(t *testing.T) {
		result, err := statsService.GetStorageUsage(context.Background())
		if err != nil {
			t.Errorf("GetStorageUsage failed: %v", err)
		}

		if result == nil {
			t.Error("GetStorageUsage returned nil result")
		}

		// 验证存储统计
		if result.TotalSpace == 0 {
			t.Error("Total space is zero")
		}

		if result.UsagePercentage < 0 || result.UsagePercentage > 100 {
			t.Error("Invalid usage percentage")
		}
	})

	t.Run("GetUserActivity", func(t *testing.T) {
		result, err := statsService.GetUserActivity(context.Background(), 1)
		if err != nil {
			t.Errorf("GetUserActivity failed: %v", err)
		}

		if result == nil {
			t.Error("GetUserActivity returned nil result")
		}

		// 验证用户活动数据
		if result.UserID == 0 {
			t.Error("Invalid user ID")
		}
	})

	t.Run("GetComplianceReport", func(t *testing.T) {
		result, err := statsService.GetComplianceReport(context.Background())
		if err != nil {
			t.Errorf("GetComplianceReport failed: %v", err)
		}

		if result == nil {
			t.Error("GetComplianceReport returned nil result")
		}

		// 验证合规报告
		if result.TotalDocuments == 0 {
			t.Error("Total documents count is zero")
		}

		if len(result.CategoryCompliance) == 0 {
			t.Error("No category compliance data available")
		}
	})

	t.Run("ExportStats", func(t *testing.T) {
		// 测试JSON导出
		data, err := statsService.ExportStats(context.Background(), "overview", "json")
		if err != nil {
			t.Errorf("ExportStats (JSON) failed: %v", err)
		}

		if len(data) == 0 {
			t.Error("Exported data is empty")
		}

		// 测试CSV导出
		data, err = statsService.ExportStats(context.Background(), "overview", "csv")
		if err != nil {
			t.Errorf("ExportStats (CSV) failed: %v", err)
		}

		if len(data) == 0 {
			t.Error("Exported CSV data is empty")
		}
	})
}

// Mock implementations for testing

// MockFile 模拟上传文件
type MockFile struct {
	content  []byte
	filename string
}

func (m *MockFile) Open() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.content)), nil
}

func (m *MockFile) Filename() string {
	return m.filename
}

func (m *MockFile) Size() int64 {
	return int64(len(m.content))
}

func (m *MockFile) ContentType() string {
	return "application/octet-stream"
}

// MockDocumentRepository 模拟文档仓库
type MockDocumentRepository struct{}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *models.Document) error {
	return nil
}

func (m *MockDocumentRepository) FindByID(ctx context.Context, id uint) (*models.Document, error) {
	return &models.Document{
		ID:        id,
		Name:      "Test Document",
		Filename:  "test.txt",
		Filepath:  "test.txt",
		Filesize:  1024,
		MimeType:  "text/plain",
		Category:  "Test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *MockDocumentRepository) Update(ctx context.Context, doc *models.Document) error {
	return nil
}

func (m *MockDocumentRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *MockDocumentRepository) List(ctx context.Context, params interface{}) ([]*models.Document, int64, error) {
	docs := []*models.Document{
		{
			ID:        1,
			Name:      "Test Document 1",
			Category:  "Legal",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "Test Document 2",
			Category:  "Financial",
			CreatedAt: time.Now(),
		},
	}
	return docs, int64(len(docs)), nil
}

func (m *MockDocumentRepository) GetStats(ctx context.Context) (*struct {
	Total      int64
	ByCategory []struct {
		Category string
		Count    int64
	}
	ByEntityType []struct {
		EntityType string
		Count      int64
	}
	RecentUploads int64
}, error) {
	return &struct {
		Total      int64
		ByCategory []struct {
			Category string
			Count    int64
		}
		ByEntityType []struct {
			EntityType string
			Count      int64
		}
		RecentUploads int64
	}{
		Total: 150,
		ByCategory: []struct {
			Category string
			Count    int64
		}{
			{Category: "Legal", Count: 60},
			{Category: "Financial", Count: 40},
			{Category: "Contracts", Count: 50},
		},
		ByEntityType: []struct {
			EntityType string
			Count      int64
		}{
			{EntityType: "case", Count: 80},
			{EntityType: "client", Count: 70},
		},
		RecentUploads: 12,
	}, nil
}

func (m *MockDocumentRepository) FindByEntity(ctx context.Context, entityType string, entityID uint) ([]*models.Document, error) {
	return []*models.Document{
		{
			ID:        1,
			Name:      "Entity Document",
			Category:  entityType,
			CreatedAt: time.Now(),
		},
	}, nil
}

// MockUserRepository 模拟用户仓库
type MockUserRepository struct{}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	return &models.User{
		ID:    id,
		Name:  "Test User",
		Email: "test@example.com",
	}, nil
}

// TestIntegration 测试集成功能
func TestIntegration(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "integration_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 初始化所有服务
	mockRepo := &MockDocumentRepository{}
	mockUserRepo := &MockUserRepository{}

	docService := NewDocumentService(mockRepo, tempDir)
	previewService := NewDocumentPreviewService(mockRepo)
	versionService := NewDocumentVersionService(mockRepo, tempDir)
	permissionService := NewDocumentPermissionService(mockRepo, mockUserRepo)
	recycleService := NewDocumentRecycleService(mockRepo, tempDir)
	searchService := NewDocumentSearchService(mockRepo)
	statsService := NewDocumentStatsService(mockRepo)

	ctx := context.Background()

	t.Run("DocumentLifecycle", func(t *testing.T) {
		// 1. 上传文档
		content := []byte("Integration test document")
		file := &MockFile{content: content, filename: "integration.txt"}

		uploadReq := &UploadRequest{
			Name:        "Integration Test Document",
			Category:    "Test",
			Description: "Integration test upload",
			File:        file,
			UploadedBy:  1,
		}

		doc, err := docService.UploadDocument(ctx, uploadReq)
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}

		// 2. 预览文档
		previewReq := &PreviewRequest{
			DocumentID: doc.ID,
		}

		_, err = previewService.GetDocumentPreview(ctx, previewReq)
		if err != nil {
			t.Errorf("Preview failed: %v", err)
		}

		// 3. 创建版本
		versionReq := &CreateVersionRequest{
			DocumentID:  doc.ID,
			Name:        "Version 2",
			Description: "Integration test version",
			CreatedBy:   1,
		}

		_, err = versionService.CreateVersion(ctx, versionReq)
		if err != nil {
			t.Errorf("Create version failed: %v", err)
		}

		// 4. 授予权限
		permReq := &PermissionRequest{
			DocumentID: doc.ID,
			UserID:     2,
			Permission: "read",
			GrantedBy:  1,
		}

		_, err = permissionService.GrantPermission(ctx, permReq)
		if err != nil {
			t.Errorf("Grant permission failed: %v", err)
		}

		// 5. 搜索文档
		searchReq := &DocumentSearchRequest{
			Query: "Integration",
			Page:  1,
		}

		searchResult, err := searchService.SearchDocuments(ctx, searchReq)
		if err != nil {
			t.Errorf("Search failed: %v", err)
		}

		if len(searchResult.Documents) == 0 {
			t.Error("Search found no documents")
		}

		// 6. 软删除
		deleteReq := &SoftDeleteRequest{
			DocumentID: doc.ID,
			DeletedBy:  1,
		}

		_, err = recycleService.SoftDelete(ctx, deleteReq)
		if err != nil {
			t.Errorf("Soft delete failed: %v", err)
		}

		// 7. 恢复文档
		restoreReq := &RestoreRequest{
			DocumentIDs: []uint{doc.ID},
			RestoredBy:  1,
		}

		_, err = recycleService.RestoreDocuments(ctx, restoreReq)
		if err != nil {
			t.Errorf("Restore failed: %v", err)
		}

		// 8. 获取统计信息
		_, err = statsService.GetDocumentOverview(ctx)
		if err != nil {
			t.Errorf("Get stats failed: %v", err)
		}
	})
}

// BenchmarkSearchDocument 性能测试
func BenchmarkSearchDocument(b *testing.B) {
	mockRepo := &MockDocumentRepository{}
	searchService := NewDocumentSearchService(mockRepo)

	req := &DocumentSearchRequest{
		Query:    "test",
		Category: "legal",
		Page:     1,
		PageSize: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := searchService.SearchDocuments(context.Background(), req)
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
	}
}

// BenchmarkDocumentStats 性能测试
func BenchmarkDocumentStats(b *testing.B) {
	mockRepo := &MockDocumentRepository{}
	statsService := NewDocumentStatsService(mockRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := statsService.GetDocumentOverview(context.Background())
		if err != nil {
			b.Fatalf("Get stats failed: %v", err)
		}
	}
}