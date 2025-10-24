package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/services"
)

// Simplified test program for document management system
func main() {
	fmt.Println("🚀 文档管理系统测试启动")

	// Test document statistics service
	fmt.Println("\n📊 测试文档统计服务...")
	testDocumentStats()

	fmt.Println("\n✅ 文档管理系统测试完成!")
}

func testDocumentStats() {
	// Create a mock document repository
	mockRepo := &MockDocumentRepository{}

	// Create stats service
	statsService := services.NewDocumentStatsService(mockRepo)

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

// MockDocumentRepository implements a minimal repository for testing
type MockDocumentRepository struct{}

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

func (m *MockDocumentRepository) List(ctx context.Context, params interface{}) ([]interface{}, int64, error) {
	// Mock document list for compliance testing
	mockDocs := []interface{}{
		struct {
			ID          uint
			Name        string
			Tags        string
			Description string
			Category    string
		}{ID: 1, Name: "Contract.pdf", Tags: "legal,important", Description: "Legal contract", Category: "Legal"},
		{ID: 2, Name: "Report.docx", Tags: "", Description: "", Category: "Reports"},
		{ID: 3, Name: "Invoice.pdf", Tags: "financial", Description: "Financial invoice", Category: "Financial"},
	}

	return mockDocs, int64(len(mockDocs)), nil
}