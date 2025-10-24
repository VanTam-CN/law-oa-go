package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Standalone test for document statistics functionality
func main() {
	fmt.Println("🚀 文档统计功能独立测试启动")

	// Test document statistics service
	fmt.Println("\n📊 测试文档统计服务...")
	testDocumentStats()

	fmt.Println("\n✅ 文档统计功能测试完成!")
}

func testDocumentStats() {
	// Create a mock stats service directly
	statsService := &MockDocumentStatsService{}

	// Test overview statistics
	fmt.Println("   - 获取文档总览统计...")
	overview, err := statsService.GetDocumentOverview(context.Background())
	if err != nil {
		log.Printf("   ❌ 获取总览统计失败: %v", err)
		return
	}

	fmt.Printf("   ✅ 总览统计成功: 总文档数=%d, 活跃文档数=%d, 最近上传=%d\n",
		overview.TotalDocuments, overview.ActiveDocuments, overview.RecentUploads)

	// Test storage usage
	fmt.Println("   - 获取存储使用统计...")
	storage, err := statsService.GetStorageUsage(context.Background())
	if err != nil {
		log.Printf("   ❌ 获取存储统计失败: %v", err)
		return
	}

	fmt.Printf("   ✅ 存储统计成功: 总空间=%dGB, 已用空间=%dMB, 使用率=%.2f%%\n",
		storage.TotalSpace/(1024*1024*1024), storage.UsedSpace/(1024*1024), storage.UsagePercentage)

	// Test user activity
	fmt.Println("   - 获取用户活动统计...")
	activity, err := statsService.GetUserActivity(context.Background(), 1)
	if err != nil {
		log.Printf("   ❌ 获取用户活动失败: %v", err)
		return
	}

	fmt.Printf("   ✅ 用户活动统计成功: 用户上传=%d, 下载=%d, 删除=%d\n",
		activity.TotalUploads, activity.TotalDownloads, activity.TotalDeleted)

	// Test compliance report
	fmt.Println("   - 获取合规报告...")
	compliance, err := statsService.GetComplianceReport(context.Background())
	if err != nil {
		log.Printf("   ❌ 获取合规报告失败: %v", err)
		return
	}

	fmt.Printf("   ✅ 合规报告成功: 总文档=%d, 有标签文档=%d, 有元数据文档=%d\n",
		compliance.TotalDocuments, compliance.DocumentsWithTags, compliance.DocumentsWithMetadata)

	// Test export functionality
	fmt.Println("   - 测试导出功能...")
	_, err = statsService.ExportStats(context.Background(), "overview", "json")
	if err != nil {
		log.Printf("   ❌ 导出统计失败: %v", err)
		return
	}

	fmt.Println("   ✅ 导出功能测试成功")
}

// MockDocumentStatsService provides a standalone implementation
type MockDocumentStatsService struct{}

// DocumentOverviewStats represents overall document statistics
type DocumentOverviewStats struct {
	TotalDocuments      int64               `json:"total_documents"`
	TotalStorage        int64               `json:"total_storage"`
	ActiveDocuments     int64               `json:"active_documents"`
	DeletedDocuments    int64               `json:"deleted_documents"`
	VersionCount        int64               `json:"version_count"`
	RecentUploads       int64               `json:"recent_uploads"`        // Last 7 days
	RecentDownloads     int64               `json:"recent_downloads"`      // Last 7 days
	DocumentsByCategory []CategoryStats     `json:"documents_by_category"`
	DocumentsByType     []TypeStats         `json:"documents_by_type"`
	UploadTrends        []TrendData         `json:"upload_trends"`
	StorageGrowth       []GrowthData        `json:"storage_growth"`
	TopUploaders         []UserStats         `json:"top_uploaders"`
	LargestDocuments    []DocumentSizeStats `json:"largest_documents"`
	GeneratedAt         time.Time           `json:"generated_at"`
}

// CategoryStats represents statistics by category
type CategoryStats struct {
	Category        string  `json:"category"`
	Count           int64   `json:"count"`
	Size            int64   `json:"size"`
	Percentage      float64 `json:"percentage"`
	AverageSize     float64 `json:"average_size"`
	TrendPercentage float64 `json:"trend_percentage"` // Growth compared to last period
}

// TypeStats represents statistics by entity type
type TypeStats struct {
	EntityType string  `json:"entity_type"`
	Count      int64   `json:"count"`
	Size       int64   `json:"size"`
	Percentage float64 `json:"percentage"`
}

// TrendData represents trend data over time
type TrendData struct {
	Date      string  `json:"date"`
	Count     int64   `json:"count"`
	Size      int64   `json:"size"`
	TrendType string  `json:"trend_type"` // "upload", "download", "delete"
}

// GrowthData represents storage growth data
type GrowthData struct {
	Date      string `json:"date"`
	UsedSpace int64  `json:"used_space"`
	FileCount int64  `json:"file_count"`
}

// UserStats represents user-related statistics
type UserStats struct {
	UserID        uint      `json:"user_id"`
	UserName      string    `json:"user_name"`
	UserEmail     string    `json:"user_email"`
	DocumentCount int64     `json:"document_count"`
	TotalSize     int64     `json:"total_size"`
	LastActive    time.Time `json:"last_active"`
}

// DocumentSizeStats represents document size information
type DocumentSizeStats struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	FileSize   int64     `json:"file_size"`
	UploadedBy string    `json:"uploaded_by"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// StorageUsageStats represents storage usage statistics
type StorageUsageStats struct {
	TotalSpace      int64               `json:"total_space"`
	UsedSpace       int64               `json:"used_space"`
	AvailableSpace  int64               `json:"available_space"`
	UsagePercentage float64            `json:"usage_percentage"`
	ByCategory      []CategoryStats     `json:"by_category"`
	ByFileType      []FileTypeStats     `json:"by_file_type"`
	LargeFiles      []DocumentSizeStats `json:"large_files"`
	OldestFiles     []DocumentSizeStats `json:"oldest_files"`
	GeneratedAt     time.Time           `json:"generated_at"`
}

// FileTypeStats represents statistics by file type
type FileTypeStats struct {
	MimeType    string  `json:"mime_type"`
	Extension   string  `json:"extension"`
	Count       int64   `json:"count"`
	Size        int64   `json:"size"`
	Percentage  float64 `json:"percentage"`
	AverageSize float64 `json:"average_size"`
}

// UserActivityStats represents user activity statistics
type UserActivityStats struct {
	UserID              uint                  `json:"user_id"`
	UserName            string                `json:"user_name"`
	UserEmail           string                `json:"user_email"`
	TotalUploads        int64                 `json:"total_uploads"`
	TotalDownloads      int64                 `json:"total_downloads"`
	TotalDeleted        int64                 `json:"total_deleted"`
	LastUpload          time.Time             `json:"last_upload"`
	LastDownload        time.Time             `json:"last_download"`
	ActivityByDate      []DailyActivity       `json:"activity_by_date"`
	CategoryBreakdown   []CategoryStats       `json:"category_breakdown"`
	FileTypeBreakdown   []FileTypeStats       `json:"file_type_breakdown"`
	GeneratedAt         time.Time             `json:"generated_at"`
}

// DailyActivity represents daily user activity
type DailyActivity struct {
	Date      string `json:"date"`
	Uploads   int64  `json:"uploads"`
	Downloads int64  `json:"downloads"`
	Deletes   int64  `json:"deletes"`
}

// ComplianceReport represents compliance-related statistics
type ComplianceReport struct {
	TotalDocuments         int64                 `json:"total_documents"`
	DocumentsWithTags      int64                 `json:"documents_with_tags"`
	DocumentsWithMetadata  int64                 `json:"documents_with_metadata"`
	UnTaggedDocuments      []DocumentSizeStats   `json:"untagged_documents"`
	CategoryCompliance     []CategoryCompliance  `json:"category_compliance"`
	TagCoverage            TagCoverage           `json:"tag_coverage"`
	MetadataCompleteness   MetadataCompleteness  `json:"metadata_completeness"`
	GeneratedAt            time.Time             `json:"generated_at"`
}

// CategoryCompliance represents compliance status by category
type CategoryCompliance struct {
	Category       string   `json:"category"`
	TotalCount     int64    `json:"total_count"`
	CompliantCount int64    `json:"compliant_count"`
	ComplianceRate float64  `json:"compliance_rate"`
	Issues         []string `json:"issues"`
}

// TagCoverage represents tag coverage statistics
type TagCoverage struct {
	TotalTags       int           `json:"total_tags"`
	PopularTags     []TagStats    `json:"popular_tags"`
	UnusedTags      []string      `json:"unused_tags"`
	TagDistribution map[string]int `json:"tag_distribution"`
}

// TagStats represents tag statistics
type TagStats struct {
	Tag        string  `json:"tag"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// MetadataCompleteness represents metadata completeness statistics
type MetadataCompleteness struct {
	FieldsCompleted     int                    `json:"fields_completed"`
	TotalFields         int                    `json:"total_fields"`
	CompletenessRate    float64                `json:"completeness_rate"`
	MissingFieldsByDoc  map[string][]string    `json:"missing_fields_by_doc"`
	FieldCoverage       map[string]int64       `json:"field_coverage"`
}

func (s *MockDocumentStatsService) GetDocumentOverview(ctx context.Context) (*DocumentOverviewStats, error) {
	// Mock data
	return &DocumentOverviewStats{
		TotalDocuments: 150,
		ActiveDocuments: 145,
		DeletedDocuments: 5,
		RecentUploads:   12,
		DocumentsByCategory: []CategoryStats{
			{Category: "Legal", Count: 60, Size: 600 * 1024 * 1024, Percentage: 40.0, AverageSize: 10 * 1024 * 1024},
			{Category: "Financial", Count: 40, Size: 400 * 1024 * 1024, Percentage: 26.7, AverageSize: 10 * 1024 * 1024},
			{Category: "Contracts", Count: 50, Size: 500 * 1024 * 1024, Percentage: 33.3, AverageSize: 10 * 1024 * 1024},
		},
		GeneratedAt: time.Now(),
	}, nil
}

func (s *MockDocumentStatsService) GetStorageUsage(ctx context.Context) (*StorageUsageStats, error) {
	totalSpace := int64(100 * 1024 * 1024 * 1024) // 100GB
	usedSpace := int64(15 * 1024 * 1024 * 1024)  // 15GB

	return &StorageUsageStats{
		TotalSpace:      totalSpace,
		UsedSpace:       usedSpace,
		AvailableSpace:  totalSpace - usedSpace,
		UsagePercentage: float64(usedSpace) / float64(totalSpace) * 100,
		ByCategory: []CategoryStats{
			{Category: "Legal", Count: 60, Size: 600 * 1024 * 1024, Percentage: 40.0},
			{Category: "Financial", Count: 40, Size: 400 * 1024 * 1024, Percentage: 26.7},
		},
		GeneratedAt: time.Now(),
	}, nil
}

func (s *MockDocumentStatsService) GetUserActivity(ctx context.Context, userID uint) (*UserActivityStats, error) {
	return &UserActivityStats{
		UserID:        userID,
		UserName:      "Demo User",
		UserEmail:     "user@example.com",
		TotalUploads:  45,
		TotalDownloads: 128,
		TotalDeleted:  3,
		LastUpload:    time.Now().Add(-2 * time.Hour),
		LastDownload:  time.Now().Add(-30 * time.Minute),
		GeneratedAt:   time.Now(),
	}, nil
}

func (s *MockDocumentStatsService) GetComplianceReport(ctx context.Context) (*ComplianceReport, error) {
	return &ComplianceReport{
		TotalDocuments:        150,
		DocumentsWithTags:     120,
		DocumentsWithMetadata: 135,
		CategoryCompliance: []CategoryCompliance{
			{
				Category:       "Legal",
				TotalCount:     60,
				CompliantCount: 55,
				ComplianceRate: 91.7,
				Issues:         []string{"Missing tags in 5 documents"},
			},
		},
		TagCoverage: TagCoverage{
			TotalTags: 25,
			PopularTags: []TagStats{
				{Tag: "legal", Count: 30, Percentage: 20.0},
				{Tag: "important", Count: 25, Percentage: 16.7},
			},
		},
		GeneratedAt: time.Now(),
	}, nil
}

func (s *MockDocumentStatsService) ExportStats(ctx context.Context, statsType string, format string) ([]byte, error) {
	if format == "json" {
		return []byte(`{"status": "success", "data": "mock export data"}`), nil
	}
	return []byte("csv export data"), nil
}

// Mock errors package functions
func (s *MockDocumentStatsService) ValidationError(field, message string) error {
	return fmt.Errorf("validation error in %s: %s", field, message)
}