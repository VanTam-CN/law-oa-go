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

var ErrDocumentStatisticsUnavailable = errors.New("document statistics are not connected to an audited storage/activity source")

// DocumentStatsService handles document statistics and reporting
type DocumentStatsService struct {
	docRepo      repositories.DocumentRepository
	viewerUserID uint
}

// NewDocumentStatsService creates a new document statistics service
func NewDocumentStatsService(docRepo repositories.DocumentRepository) *DocumentStatsService {
	return &DocumentStatsService{
		docRepo: docRepo,
	}
}

// WithViewer 返回绑定了指定 viewer 的派生实例。
//
// 用于 HTTP 请求路径：handler 在调用前用当前用户 ID 派生 per-request service，
// 内部所有 GetStats/聚合查询自动应用隔离墙过滤，避免条数侧信道。
// 返回的实例与原 service 共享底层仓储，但 viewerUserID 不同。
func (s *DocumentStatsService) WithViewer(viewerUserID uint) *DocumentStatsService {
	clone := *s
	clone.viewerUserID = viewerUserID
	return &clone
}

// DocumentOverviewStats represents overall document statistics
type DocumentOverviewStats struct {
	TotalDocuments      int64               `json:"total_documents"`
	TotalStorage        int64               `json:"total_storage"`
	ActiveDocuments     int64               `json:"active_documents"`
	DeletedDocuments    int64               `json:"deleted_documents"`
	VersionCount        int64               `json:"version_count"`
	RecentUploads       int64               `json:"recent_uploads"`   // Last 7 days
	RecentDownloads     int64               `json:"recent_downloads"` // Last 7 days
	DocumentsByCategory []CategoryStats     `json:"documents_by_category"`
	DocumentsByType     []TypeStats         `json:"documents_by_type"`
	UploadTrends        []TrendData         `json:"upload_trends"`
	StorageGrowth       []GrowthData        `json:"storage_growth"`
	TopUploaders        []UserStats         `json:"top_uploaders"`
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
	Date      string `json:"date"`
	Count     int64  `json:"count"`
	Size      int64  `json:"size"`
	TrendType string `json:"trend_type"` // "upload", "download", "delete"
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
	UsagePercentage float64             `json:"usage_percentage"`
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
	UserID            uint            `json:"user_id"`
	UserName          string          `json:"user_name"`
	UserEmail         string          `json:"user_email"`
	TotalUploads      int64           `json:"total_uploads"`
	TotalDownloads    int64           `json:"total_downloads"`
	TotalDeleted      int64           `json:"total_deleted"`
	LastUpload        time.Time       `json:"last_upload"`
	LastDownload      time.Time       `json:"last_download"`
	ActivityByDate    []DailyActivity `json:"activity_by_date"`
	CategoryBreakdown []CategoryStats `json:"category_breakdown"`
	FileTypeBreakdown []FileTypeStats `json:"file_type_breakdown"`
	GeneratedAt       time.Time       `json:"generated_at"`
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
	TotalDocuments        int64                `json:"total_documents"`
	DocumentsWithTags     int64                `json:"documents_with_tags"`
	DocumentsWithMetadata int64                `json:"documents_with_metadata"`
	UnTaggedDocuments     []DocumentSizeStats  `json:"untagged_documents"`
	CategoryCompliance    []CategoryCompliance `json:"category_compliance"`
	TagCoverage           TagCoverage          `json:"tag_coverage"`
	MetadataCompleteness  MetadataCompleteness `json:"metadata_completeness"`
	GeneratedAt           time.Time            `json:"generated_at"`
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
	TotalTags       int            `json:"total_tags"`
	PopularTags     []TagStats     `json:"popular_tags"`
	UnusedTags      []string       `json:"unused_tags"`
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
	FieldsCompleted    int                 `json:"fields_completed"`
	TotalFields        int                 `json:"total_fields"`
	CompletenessRate   float64             `json:"completeness_rate"`
	MissingFieldsByDoc map[string][]string `json:"missing_fields_by_doc"`
	FieldCoverage      map[string]int64    `json:"field_coverage"`
}

// GetDocumentOverview returns comprehensive document overview statistics
func (s *DocumentStatsService) GetDocumentOverview(ctx context.Context) (*DocumentOverviewStats, error) {
	return nil, ErrDocumentStatisticsUnavailable
	/*

		// Get basic stats from repository
		stats, err := s.docRepo.GetStats(ctx, s.viewerUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get document stats: %w", err)
		}

		// Get category statistics
		categoryStats, err := s.getCategoryStats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get category stats: %w", err)
		}

		// Get type statistics
		typeStats, err := s.getTypeStats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get type stats: %w", err)
		}

		// Get upload trends (last 30 days)
		uploadTrends, err := s.getUploadTrends(ctx, 30)
		if err != nil {
			return nil, fmt.Errorf("failed to get upload trends: %w", err)
		}

		// Get storage growth (last 30 days)
		storageGrowth, err := s.getStorageGrowth(ctx, 30)
		if err != nil {
			return nil, fmt.Errorf("failed to get storage growth: %w", err)
		}

		// Get top uploaders
		topUploaders, err := s.getTopUploaders(ctx, 10)
		if err != nil {
			return nil, fmt.Errorf("failed to get top uploaders: %w", err)
		}

		// Get largest documents
		largestDocs, err := s.getLargestDocuments(ctx, 10)
		if err != nil {
			return nil, fmt.Errorf("failed to get largest documents: %w", err)
		}

		return &DocumentOverviewStats{
			TotalDocuments:      stats.Total,
			ActiveDocuments:     stats.Total, // Use Total as Active since no Active field
			DeletedDocuments:    0,           // No Deleted field available
			RecentUploads:       stats.RecentUploads,
			DocumentsByCategory: categoryStats,
			DocumentsByType:     typeStats,
			UploadTrends:        uploadTrends,
			StorageGrowth:       storageGrowth,
			TopUploaders:        topUploaders,
			LargestDocuments:    largestDocs,
			GeneratedAt:         time.Now(),
		}, nil
	*/
}

// GetStorageUsage returns detailed storage usage statistics
func (s *DocumentStatsService) GetStorageUsage(ctx context.Context) (*StorageUsageStats, error) {
	return nil, ErrDocumentStatisticsUnavailable
}

// GetUserActivity returns user activity statistics
func (s *DocumentStatsService) GetUserActivity(ctx context.Context, userID uint) (*UserActivityStats, error) {
	return nil, ErrDocumentStatisticsUnavailable
}

// GetComplianceReport returns compliance-related statistics
func (s *DocumentStatsService) GetComplianceReport(ctx context.Context) (*ComplianceReport, error) {
	return nil, ErrDocumentStatisticsUnavailable
	/*

		// Get all documents
		documents, _, err := s.docRepo.List(ctx, &repositories.DocumentListParams{
			Page:     1,
			PageSize: 1000,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list documents: %w", err)
		}

		totalDocs := len(documents)
		docsWithTags := 0
		docsWithMetadata := 0
		var untaggedDocs []DocumentSizeStats

		// Analyze documents for compliance
		for _, doc := range documents {
			hasTags := doc.Tags != ""
			hasMetadata := doc.Description != "" && doc.Category != ""

			if hasTags {
				docsWithTags++
			}
			if hasMetadata {
				docsWithMetadata++
			}

			if !hasTags {
				untaggedDocs = append(untaggedDocs, DocumentSizeStats{
					ID:         doc.ID,
					Name:       doc.Name,
					Category:   doc.Category,
					FileSize:   doc.Filesize,
					UploadedBy: "Unknown", // Would be populated from user data
					UploadedAt: doc.CreatedAt,
				})
			}
		}

		// Generate tag coverage
		tagCoverage := s.generateTagCoverage(documents)

		// Generate metadata completeness
		metadataCompleteness := s.generateMetadataCompleteness(documents)

		categoryCompliance := []CategoryCompliance{}

		return &ComplianceReport{
			TotalDocuments:        int64(totalDocs),
			DocumentsWithTags:     int64(docsWithTags),
			DocumentsWithMetadata: int64(docsWithMetadata),
			UnTaggedDocuments:     untaggedDocs,
			CategoryCompliance:    categoryCompliance,
			TagCoverage:           tagCoverage,
			MetadataCompleteness:  metadataCompleteness,
			GeneratedAt:           time.Now(),
		}, nil
	*/
}

// ExportStats exports statistics in various formats
func (s *DocumentStatsService) ExportStats(ctx context.Context, statsType string, format string) ([]byte, error) {
	return nil, ErrDocumentStatisticsUnavailable
	/*

		switch statsType {
		case "overview":
			stats, err := s.GetDocumentOverview(ctx)
			if err != nil {
				return nil, err
			}
			return s.exportOverview(stats, format)
		case "storage":
			stats, err := s.GetStorageUsage(ctx)
			if err != nil {
				return nil, err
			}
			return s.exportStorage(stats, format)
		case "compliance":
			stats, err := s.GetComplianceReport(ctx)
			if err != nil {
				return nil, err
			}
			return s.exportCompliance(stats, format)
		default:
			return nil, fmt.Errorf("unsupported stats type: %s", statsType)
		}
	*/
}

// Helper methods

func (s *DocumentStatsService) getCategoryStats(ctx context.Context) ([]CategoryStats, error) {
	stats, err := s.docRepo.GetStats(ctx, s.viewerUserID)
	if err != nil {
		return nil, err
	}

	var categoryStats []CategoryStats
	totalCount := stats.Total

	for _, stat := range stats.ByCategory {
		percentage := float64(stat.Count) / float64(totalCount) * 100
		estimatedSize := stat.Count * 10 * 1024 * 1024 // Estimate: 10MB per document

		categoryStats = append(categoryStats, CategoryStats{
			Category:    stat.Category,
			Count:       stat.Count,
			Size:        estimatedSize,
			Percentage:  percentage,
			AverageSize: float64(estimatedSize) / float64(stat.Count),
		})
	}

	return categoryStats, nil
}

func (s *DocumentStatsService) getTypeStats(ctx context.Context) ([]TypeStats, error) {
	stats, err := s.docRepo.GetStats(ctx, s.viewerUserID)
	if err != nil {
		return nil, err
	}

	var typeStats []TypeStats
	totalCount := stats.Total

	for _, stat := range stats.ByEntityType {
		percentage := float64(stat.Count) / float64(totalCount) * 100

		estimatedSize := stat.Count * 10 * 1024 * 1024 // Estimate: 10MB per document
		typeStats = append(typeStats, TypeStats{
			EntityType: stat.EntityType,
			Count:      stat.Count,
			Size:       estimatedSize,
			Percentage: percentage,
		})
	}

	return typeStats, nil
}

func (s *DocumentStatsService) getUploadTrends(ctx context.Context, days int) ([]TrendData, error) {
	var trends []TrendData

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		// In a real implementation, this would query the database
		// For now, generate mock data
		count := int64(5 + i%3)     // Mock varying upload counts
		size := count * 1024 * 1024 // Mock size calculation

		trends = append(trends, TrendData{
			Date:      dateStr,
			Count:     count,
			Size:      size,
			TrendType: "upload",
		})
	}

	return trends, nil
}

func (s *DocumentStatsService) getStorageGrowth(ctx context.Context, days int) ([]GrowthData, error) {
	var growth []GrowthData

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		// In a real implementation, this would query the database
		// For now, generate mock cumulative growth data
		cumulativeCount := int64((days - i) * 2)
		usedSpace := cumulativeCount * 5 * 1024 * 1024 // Mock cumulative size

		growth = append(growth, GrowthData{
			Date:      dateStr,
			UsedSpace: usedSpace,
			FileCount: cumulativeCount,
		})
	}

	return growth, nil
}

func (s *DocumentStatsService) getTopUploaders(ctx context.Context, limit int) ([]UserStats, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []UserStats{
		{
			UserID:        1,
			UserName:      "Alice Johnson",
			UserEmail:     "alice@example.com",
			DocumentCount: 45,
			TotalSize:     500 * 1024 * 1024,
			LastActive:    time.Now().Add(-2 * time.Hour),
		},
		{
			UserID:        2,
			UserName:      "Bob Smith",
			UserEmail:     "bob@example.com",
			DocumentCount: 32,
			TotalSize:     300 * 1024 * 1024,
			LastActive:    time.Now().Add(-1 * time.Hour),
		},
	}, nil
}

func (s *DocumentStatsService) getLargestDocuments(ctx context.Context, limit int) ([]DocumentSizeStats, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []DocumentSizeStats{
		{
			ID:         1,
			Name:       "Large Contract.pdf",
			Category:   "Contracts",
			FileSize:   25 * 1024 * 1024,
			UploadedBy: "Alice Johnson",
			UploadedAt: time.Now().AddDate(0, -1, 0),
		},
		{
			ID:         2,
			Name:       "Financial Report.xlsx",
			Category:   "Financial",
			FileSize:   18 * 1024 * 1024,
			UploadedBy: "Bob Smith",
			UploadedAt: time.Now().AddDate(0, -2, 0),
		},
	}, nil
}

func (s *DocumentStatsService) getLargeDocuments(ctx context.Context, minSize int64, limit int) ([]DocumentSizeStats, error) {
	return s.getLargestDocuments(ctx, limit)
}

func (s *DocumentStatsService) getOldestDocuments(ctx context.Context, limit int) ([]DocumentSizeStats, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []DocumentSizeStats{
		{
			ID:         3,
			Name:       "Old Document.txt",
			Category:   "Legal",
			FileSize:   1024 * 1024,
			UploadedBy: "Admin",
			UploadedAt: time.Now().AddDate(-1, 0, 0),
		},
		{
			ID:         4,
			Name:       "Archive Document.pdf",
			Category:   "Archive",
			FileSize:   5 * 1024 * 1024,
			UploadedBy: "Admin",
			UploadedAt: time.Now().AddDate(-1, -1, 0),
		},
	}, nil
}

func (s *DocumentStatsService) getFileTypeStats(ctx context.Context) ([]FileTypeStats, error) {
	// In a real implementation, this would query the database
	// For now, return mock data
	return []FileTypeStats{
		{
			MimeType:    "application/pdf",
			Extension:   ".pdf",
			Count:       150,
			Size:        1024 * 1024 * 1024,
			Percentage:  60.0,
			AverageSize: 6.8 * 1024 * 1024,
		},
		{
			MimeType:    "application/msword",
			Extension:   ".doc",
			Count:       80,
			Size:        400 * 1024 * 1024,
			Percentage:  32.0,
			AverageSize: 5.0 * 1024 * 1024,
		},
		{
			MimeType:    "text/plain",
			Extension:   ".txt",
			Count:       20,
			Size:        80 * 1024 * 1024,
			Percentage:  8.0,
			AverageSize: 4.0 * 1024 * 1024,
		},
	}, nil
}

func (s *DocumentStatsService) generateMockDailyActivity(days int) []DailyActivity {
	var activity []DailyActivity

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		activity = append(activity, DailyActivity{
			Date:      dateStr,
			Uploads:   int64(1 + i%3),
			Downloads: int64(3 + i%5),
			Deletes:   int64(i % 2),
		})
	}

	return activity
}

func (s *DocumentStatsService) generateTagCoverage(documents []*models.Document) TagCoverage {
	// Count tag frequencies
	tagCounts := make(map[string]int)
	totalTags := 0

	for _, doc := range documents {
		if doc.Tags != "" {
			tags := parseTags(doc.Tags)
			for _, tag := range tags {
				tagCounts[tag]++
				totalTags++
			}
		}
	}

	// Create popular tags
	var popularTags []TagStats
	for tag, count := range tagCounts {
		percentage := float64(count) / float64(len(documents)) * 100
		popularTags = append(popularTags, TagStats{
			Tag:        tag,
			Count:      int64(count),
			Percentage: percentage,
		})
	}

	return TagCoverage{
		TotalTags:       totalTags,
		PopularTags:     popularTags,
		UnusedTags:      []string{}, // Would be determined from predefined tag list
		TagDistribution: tagCounts,
	}
}

func (s *DocumentStatsService) generateMetadataCompleteness(documents []*models.Document) MetadataCompleteness {
	totalFields := 3 // name, description, category
	completedFields := 0
	missingFieldsByDoc := make(map[string][]string)
	fieldCoverage := map[string]int64{
		"name":        int64(len(documents)), // Name is always present
		"description": 0,
		"category":    0,
	}

	for _, doc := range documents {
		var missingFields []string

		if doc.Description != "" {
			completedFields++
			fieldCoverage["description"]++
		} else {
			missingFields = append(missingFields, "description")
		}

		if doc.Category != "" {
			completedFields++
			fieldCoverage["category"]++
		} else {
			missingFields = append(missingFields, "category")
		}

		if len(missingFields) > 0 {
			missingFieldsByDoc[doc.Name] = missingFields
		}
	}

	completenessRate := float64(completedFields) / float64(len(documents)*totalFields) * 100

	return MetadataCompleteness{
		FieldsCompleted:    completedFields,
		TotalFields:        len(documents) * totalFields,
		CompletenessRate:   completenessRate,
		MissingFieldsByDoc: missingFieldsByDoc,
		FieldCoverage:      fieldCoverage,
	}
}

func (s *DocumentStatsService) exportOverview(stats *DocumentOverviewStats, format string) ([]byte, error) {
	switch format {
	case "json":
		return s.toJSON(stats)
	case "csv":
		return s.toCSV(stats)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *DocumentStatsService) exportStorage(stats *StorageUsageStats, format string) ([]byte, error) {
	switch format {
	case "json":
		return s.toJSON(stats)
	case "csv":
		return s.toCSV(stats)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *DocumentStatsService) exportCompliance(stats *ComplianceReport, format string) ([]byte, error) {
	switch format {
	case "json":
		return s.toJSON(stats)
	case "csv":
		return s.toCSV(stats)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (s *DocumentStatsService) toJSON(data interface{}) ([]byte, error) {
	// In a real implementation, this would use json.Marshal
	return []byte(`{"json": "export"}`), nil
}

func (s *DocumentStatsService) toCSV(data interface{}) ([]byte, error) {
	// In a real implementation, this would generate CSV format
	return []byte(`csv,export`), nil
}

func parseTags(tagsStr string) []string {
	// Simple tag parsing - in a real implementation, this would be more sophisticated
	if tagsStr == "" {
		return []string{}
	}

	tags := make([]string, 0)
	for _, tag := range strings.Split(tagsStr, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags
}
