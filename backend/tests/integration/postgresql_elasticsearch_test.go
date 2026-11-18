package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	esclient "law-oa-go/internal/elasticsearch"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// TestPostgreSQL_ElasticsearchIntegration 测试PostgreSQL与Elasticsearch的集成
func TestPostgreSQL_ElasticsearchIntegration(t *testing.T) {
	if os.Getenv("LAW_OA_RUN_POSTGRES_ES_INTEGRATION") != "1" || os.Getenv("LAW_OA_POSTGRES_TEST_DSN") == "" {
		t.Skip("需要显式提供专用 LAW_OA_POSTGRES_TEST_DSN，并设置 LAW_OA_RUN_POSTGRES_ES_INTEGRATION=1")
	}

	// 跳过如果Elasticsearch不可用
	if testing.Short() {
		t.Skip("Skipping Elasticsearch integration test in short mode")
	}

	// 创建PostgreSQL测试数据库连接
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	// 尝试创建Elasticsearch客户端
	esClientWrapper, err := esclient.NewClient("localhost", "9200", "", "")
	if err != nil {
		t.Skipf("Elasticsearch not available, skipping integration test: %v", err)
		return
	}

	// 创建搜索服务
	indexPrefix := "law_oa_test_"
	searchService := services.NewSearchService(esClientWrapper.GetClient(), indexPrefix)

	ctx := context.Background()

	// 测试索引创建
	indexName := "law_oa_test_cases"
	mapping := map[string]interface{}{
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":     "text",
				"analyzer": "standard",
				"fields": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type": "keyword",
					},
				},
			},
			"description": map[string]interface{}{
				"type":     "text",
				"analyzer": "standard",
			},
			"case_type": map[string]interface{}{
				"type": "keyword",
			},
			"status": map[string]interface{}{
				"type": "keyword",
			},
			"priority": map[string]interface{}{
				"type": "keyword",
			},
			"created_at": map[string]interface{}{
				"type": "date",
			},
			"type": map[string]interface{}{
				"type": "keyword",
			},
			"suggest": map[string]interface{}{
				"type":     "completion",
				"analyzer": "simple",
			},
		},
	}

	// 删除已存在的索引（如果存在）
	if exists, _ := esClientWrapper.IndexExists(ctx, indexName); exists {
		esClientWrapper.DeleteIndex(ctx, indexName)
	}

	// 创建索引
	err = esClientWrapper.CreateIndex(ctx, indexName, mapping)
	if err != nil {
		t.Fatalf("Failed to create Elasticsearch index: %v", err)
	}
	t.Logf("Created Elasticsearch index: %s", indexName)

	// 创建测试数据
	user := createTestUserPostgreSQL(db)
	client := createTestClientPostgreSQL(db)

	testCase := &models.Case{
		Title:       "Elasticsearch测试案件",
		Description: "这是一个用于测试Elasticsearch集成的案件",
		CaseType:    "民事",
		Priority:    "high",
		Status:      "pending",
		ClientID:    client.ID,
		LawyerID:    user.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(testCase).Error; err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}

	// 索引案件到Elasticsearch
	caseData := map[string]interface{}{
		"title":       testCase.Title,
		"description": testCase.Description,
		"case_type":   testCase.CaseType,
		"status":      testCase.Status,
		"priority":    testCase.Priority,
		"created_at":  testCase.CreatedAt.Format(time.RFC3339),
		"url":         fmt.Sprintf("/cases/%d", testCase.ID),
	}

	err = searchService.IndexEntity(ctx, "cases", fmt.Sprintf("%d", testCase.ID), caseData)
	if err != nil {
		t.Fatalf("Failed to index case: %v", err)
	}
	t.Logf("Indexed case to Elasticsearch: %s", testCase.Title)

	// 等待索引刷新
	time.Sleep(1 * time.Second)

	// 测试搜索功能
	searchReq := &services.SearchRequest{
		Query:    "Elasticsearch",
		Page:     1,
		PageSize: 10,
		Types:    []string{"cases"},
	}

	searchRes, err := searchService.Search(ctx, searchReq)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if searchRes.Total < 1 {
		t.Errorf("Expected at least 1 search result, got %d", searchRes.Total)
	}

	if len(searchRes.Results) < 1 {
		t.Errorf("Expected at least 1 search result item, got %d", len(searchRes.Results))
	} else {
		// 验证搜索结果
		result := searchRes.Results[0]
		if result.Type != "cases" {
			t.Errorf("Expected result type 'cases', got '%s'", result.Type)
		}
		if !contains(result.Title, "Elasticsearch") {
			t.Errorf("Expected title to contain 'Elasticsearch', got '%s'", result.Title)
		}
		t.Logf("Search result: %s (score: %.2f)", result.Title, result.Score)
	}

	// 测试搜索建议（可选功能）
	suggestions, err := searchService.GetSearchSuggestions(ctx, "Elastic", 5)
	if err != nil {
		t.Logf("Search suggestions not available: %v", err)
	} else {
		t.Logf("Search suggestions for 'Elastic': %v", suggestions)
	}

	// 清理：删除索引
	err = esClientWrapper.DeleteIndex(ctx, indexName)
	if err != nil {
		t.Logf("Warning: Failed to delete index: %v", err)
	}

	t.Log("Elasticsearch integration test completed successfully")
}

// 辅助函数
func setupPostgreSQLTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("LAW_OA_POSTGRES_TEST_DSN")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL test database: %v", err)
	}

	// 自动迁移模型（忽略已存在的约束错误）
	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	if err != nil {
		// 尝试先删除表再重新创建
		db.Migrator().DropTable(&models.Case{}, &models.Client{}, &models.User{})
		err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
		if err != nil {
			t.Fatalf("Failed to migrate models: %v", err)
		}
	}

	// 仅清理显式指定的专用测试数据库，避免测试默认触碰业务库。
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	if _, err := sqlDB.Exec("TRUNCATE TABLE cases, clients, users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("Failed to reset dedicated PostgreSQL test tables: %v", err)
	}

	return db
}

func teardownPostgreSQLTestDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func createTestUserPostgreSQL(db *gorm.DB) *models.User {
	user := &models.User{
		Username:  "pgsql_test_user",
		Name:      "PostgreSQL测试用户",
		Email:     "pgsql_user@example.com",
		Password:  "hashed_password",
		Role:      "lawyer",
		Phone:     "13800138001",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(user).Error; err != nil {
		panic(err)
	}

	return user
}

func createTestClientPostgreSQL(db *gorm.DB) *models.Client {
	client := &models.Client{
		Name:      "PostgreSQL测试客户",
		Email:     "pgsql_client@example.com",
		Phone:     "13900139002",
		Address:   "PostgreSQL测试地址",
		Status:    "active",
		Type:      "个人",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(client).Error; err != nil {
		panic(err)
	}

	return client
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
