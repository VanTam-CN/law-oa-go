package repositories

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// TestPostgreSQL_CaseRepository_Create 测试PostgreSQL下的案件创建
func TestPostgreSQL_CaseRepository_Create(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	// 创建测试用户（律师）
	lawyer := createTestUserPostgreSQL(db)
	// 创建测试客户
	client := createTestClientPostgreSQL(db)

	caseModel := &models.Case{
		Title:       "PostgreSQL测试案件",
		Description: "这是一个PostgreSQL测试案件描述",
		ClientID:    client.ID,
		LawyerID:    lawyer.ID,
		CaseType:    "民事",
		Priority:    "high",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(ctx, caseModel)
	if err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	if caseModel.ID == 0 {
		t.Error("Case ID should not be zero after creation")
	}

	// 验证案件确实被创建
	var foundCase models.Case
	err = db.First(&foundCase, caseModel.ID).Error
	if err != nil {
		t.Fatalf("Failed to find created case: %v", err)
	}

	if foundCase.Title != caseModel.Title {
		t.Errorf("Expected case title %s, got %s", caseModel.Title, foundCase.Title)
	}

	if foundCase.ClientID != client.ID {
		t.Errorf("Expected client ID %d, got %d", client.ID, foundCase.ClientID)
	}

	if foundCase.LawyerID != lawyer.ID {
		t.Errorf("Expected lawyer ID %d, got %d", lawyer.ID, foundCase.LawyerID)
	}
}

// TestPostgreSQL_CaseRepository_FindByID 测试PostgreSQL下的案件查找（包含关联）
func TestPostgreSQL_CaseRepository_FindByID(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	// 创建测试数据
	lawyer := createTestUserPostgreSQL(db)
	client := createTestClientPostgreSQL(db)
	testCase := createTestCasePostgreSQL(db, client.ID, lawyer.ID)

	// 测试查找存在的案件（包含关联数据）
	foundCase, err := repo.FindByID(ctx, testCase.ID)
	if err != nil {
		t.Fatalf("Failed to find case: %v", err)
	}

	if foundCase == nil {
		t.Fatal("Found case should not be nil")
	}

	if foundCase.Title != testCase.Title {
		t.Errorf("Expected case title %s, got %s", testCase.Title, foundCase.Title)
	}

	// 验证关联数据是否正确加载
	if foundCase.Client == nil {
		t.Error("Client should be preloaded")
	}

	if foundCase.Lawyer == nil {
		t.Error("Lawyer should be preloaded")
	}

	if foundCase.Client.Name != client.Name {
		t.Errorf("Expected client name %s, got %s", client.Name, foundCase.Client.Name)
	}

	if foundCase.Lawyer.Name != lawyer.Name {
		t.Errorf("Expected lawyer name %s, got %s", lawyer.Name, foundCase.Lawyer.Name)
	}

	// 测试查找不存在的案件
	notFoundCase, err := repo.FindByID(ctx, 999999)
	if err != nil {
		t.Errorf("FindByID should not return error for non-existent case: %v", err)
	}

	if notFoundCase != nil {
		t.Error("Non-existent case should return nil")
	}
}

// TestPostgreSQL_CaseRepository_List 测试PostgreSQL下的案件列表查询
func TestPostgreSQL_CaseRepository_List(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	// 创建测试数据
	lawyer := createTestUserPostgreSQL(db)
	client := createTestClientPostgreSQL(db)

	// 创建多个测试案件
	cases := []*models.Case{
		{Title: "案件A", CaseType: "民事", Priority: "high", Status: "pending", ClientID: client.ID, LawyerID: lawyer.ID},
		{Title: "案件B", CaseType: "刑事", Priority: "medium", Status: "in_progress", ClientID: client.ID, LawyerID: lawyer.ID},
		{Title: "案件C", CaseType: "民事", Priority: "low", Status: "completed", ClientID: client.ID, LawyerID: lawyer.ID},
	}

	for _, caseModel := range cases {
		caseModel.CreatedAt = time.Now()
		caseModel.UpdatedAt = time.Now()
		if err := db.Create(caseModel).Error; err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}
	}

	// 测试列表查询
	params := &CaseListParams{
		Page:     1,
		PageSize: 10,
	}

	result, total, err := repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list cases: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected total count 3, got %d", total)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 cases, got %d", len(result))
	}

	// 测试状态过滤
	params.Status = "pending"
	result, total, err = repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list pending cases: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 pending case, got %d", total)
	}

	// 测试案件类型过滤
	params.Status = ""
	params.CaseType = "民事"
	result, total, err = repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list civil cases: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 civil cases, got %d", total)
	}

	// 测试优先级过滤
	params.CaseType = ""
	params.Priority = "high"
	result, total, err = repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list high priority cases: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 high priority case, got %d", total)
	}

	// 验证关联数据是否加载
	for _, caseItem := range result {
		if caseItem.Client == nil {
			t.Error("Client should be preloaded in list results")
		}
		if caseItem.Lawyer == nil {
			t.Error("Lawyer should be preloaded in list results")
		}
	}
}

// TestPostgreSQL_CaseRepository_Update 测试PostgreSQL下的案件更新
func TestPostgreSQL_CaseRepository_Update(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	// 创建测试数据
	lawyer := createTestUserPostgreSQL(db)
	client := createTestClientPostgreSQL(db)
	testCase := createTestCasePostgreSQL(db, client.ID, lawyer.ID)

	// 更新案件信息
	testCase.Title = "更新后的案件标题"
	testCase.Description = "更新后的案件描述"
	testCase.Status = "in_progress"
	testCase.UpdatedAt = time.Now()

	err := repo.Update(ctx, testCase)
	if err != nil {
		t.Fatalf("Failed to update case: %v", err)
	}

	// 验证更新是否成功
	var updatedCase models.Case
	err = db.First(&updatedCase, testCase.ID).Error
	if err != nil {
		t.Fatalf("Failed to find updated case: %v", err)
	}

	if updatedCase.Title != "更新后的案件标题" {
		t.Errorf("Expected updated title '更新后的案件标题', got '%s'", updatedCase.Title)
	}

	if updatedCase.Status != "in_progress" {
		t.Errorf("Expected updated status 'in_progress', got '%s'", updatedCase.Status)
	}
}

// TestPostgreSQL_CaseRepository_Delete 测试PostgreSQL下的案件删除
func TestPostgreSQL_CaseRepository_Delete(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	// 创建测试数据
	lawyer := createTestUserPostgreSQL(db)
	client := createTestClientPostgreSQL(db)
	testCase := createTestCasePostgreSQL(db, client.ID, lawyer.ID)

	// 删除案件
	err := repo.Delete(ctx, testCase.ID)
	if err != nil {
		t.Fatalf("Failed to delete case: %v", err)
	}

	// 验证案件是否被删除（软删除）
	var deletedCase models.Case
	err = db.First(&deletedCase, testCase.ID).Error
	if err == nil {
		t.Error("Soft deleted case should not be found with First()")
	}

	// 检查软删除标记
	err = db.Unscoped().First(&deletedCase, testCase.ID).Error
	if err != nil {
		t.Errorf("Soft deleted case should be found with Unscoped(): %v", err)
	}

	if deletedCase.DeletedAt.Time.IsZero() {
		t.Error("DeletedAt should not be zero for soft deleted case")
	}
}

// TestPostgreSQL_CaseRepository_Search 测试PostgreSQL下的案件搜索功能
func TestPostgreSQL_CaseRepository_Search(t *testing.T) {
	db := setupPostgreSQLTestDB(t)
	defer teardownPostgreSQLTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	// 创建测试数据
	lawyer := createTestUserPostgreSQL(db)
	client := createTestClientPostgreSQL(db)

	cases := []*models.Case{
		{Title: "合同纠纷案件", Description: "关于商业合同的纠纷处理", CaseType: "民事", ClientID: client.ID, LawyerID: lawyer.ID},
		{Title: "劳动争议案件", Description: "员工与公司的劳动争议", CaseType: "民事", ClientID: client.ID, LawyerID: lawyer.ID},
		{Title: "刑事案件", Description: "涉及刑事责任的案件", CaseType: "刑事", ClientID: client.ID, LawyerID: lawyer.ID},
	}

	for _, caseModel := range cases {
		caseModel.CreatedAt = time.Now()
		caseModel.UpdatedAt = time.Now()
		if err := db.Create(caseModel).Error; err != nil {
			t.Fatalf("Failed to create test case: %v", err)
		}
	}

	// 测试按标题搜索
	params := &CaseListParams{
		Page:     1,
		PageSize: 10,
		Search:   "合同",
	}

	result, total, err := repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to search cases: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 case for '合同' search, got %d", total)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 result for '合同' search, got %d", len(result))
	}

	if result[0].Title != "合同纠纷案件" {
		t.Errorf("Expected '合同纠纷案件', got '%s'", result[0].Title)
	}

	// 测试按描述搜索
	params.Search = "劳动"
	result, total, err = repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to search cases by description: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 case for '劳动' search, got %d", total)
	}

	// 测试不匹配的搜索
	params.Search = "不存在的关键词"
	result, total, err = repo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to search cases with non-existent term: %v", err)
	}

	if total != 0 {
		t.Errorf("Expected 0 results for non-existent search term, got %d", total)
	}
}

// createTestUserPostgreSQL 创建PostgreSQL测试用户
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

// createTestCasePostgreSQL 创建PostgreSQL测试案件
func createTestCasePostgreSQL(db *gorm.DB, clientID, lawyerID uint) *models.Case {
	caseModel := &models.Case{
		Title:       "PostgreSQL测试案件",
		Description: "这是一个PostgreSQL测试案件",
		ClientID:    clientID,
		LawyerID:    lawyerID,
		CaseType:    "民事",
		Priority:    "medium",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(caseModel).Error; err != nil {
		panic(err)
	}

	return caseModel
}