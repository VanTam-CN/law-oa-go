package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"law-oa-go/internal/models"
)

func TestCaseRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	client := createTestClient(db)
	lawyer := createTestUserWithSuffix(db, "1")

	caseModel := &models.Case{
		Title:       "测试案件",
		CaseType:    "民事",
		Priority:    "normal",
		Status:      "pending",
		ClientID:    client.ID,
		LawyerID:    lawyer.ID,
		Description: "测试案件描述",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(ctx, caseModel)
	assert.NoError(t, err)
	assert.NotZero(t, caseModel.ID)

	// 验证案件确实被创建
	var foundCase models.Case
	err = db.First(&foundCase, caseModel.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, caseModel.Title, foundCase.Title)
	assert.Equal(t, caseModel.ClientID, foundCase.ClientID)
	assert.Equal(t, caseModel.LawyerID, foundCase.LawyerID)
}

func TestCaseRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	// 创建测试案件
	testCase := createTestCase(db, createTestClientWithSuffix(db, "1").ID, createTestUserWithSuffix(db, "1").ID)

	// 测试查找存在的案件
	foundCase, err := repo.FindByID(ctx, testCase.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundCase)
	assert.Equal(t, testCase.Title, foundCase.Title)

	// 测试查找不存在的案件
	notFoundCase, err := repo.FindByID(ctx, 999999)
	assert.NoError(t, err) // CaseRepository.FindByID在找不到时返回nil而不是error
	assert.Nil(t, notFoundCase)
}

func TestCaseRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	testCase := createTestCase(db, createTestClientWithSuffix(db, "1").ID, createTestUserWithSuffix(db, "1").ID)

	// 测试更新案件
	testCase.Title = "更新后的案件标题"
	testCase.Status = "in_progress"
	testCase.CaseType = "刑事"

	err := repo.Update(ctx, testCase)
	assert.NoError(t, err)

	// 验证更新是否成功
	var updatedCase models.Case
	err = db.First(&updatedCase, testCase.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "更新后的案件标题", updatedCase.Title)
	assert.Equal(t, "in_progress", updatedCase.Status)
	assert.Equal(t, "刑事", updatedCase.CaseType)
}

func TestCaseRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	testCase := createTestCase(db, createTestClientWithSuffix(db, "1").ID, createTestUserWithSuffix(db, "1").ID)

	// 测试删除案件
	err := repo.Delete(ctx, testCase.ID)
	assert.NoError(t, err)

	// 验证案件是否被删除
	var deletedCase models.Case
	err = db.First(&deletedCase, testCase.ID).Error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestCaseRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	client := createTestClient(db)
	lawyer := createTestUserWithSuffix(db, "1")

	// 创建多个测试案件
	caseTypes := []string{"民事", "刑事", "行政"}
	statuses := []string{"pending", "in_progress", "completed"}

	for i := 0; i < 5; i++ {
		caseModel := &models.Case{
			Title:       "案件" + string(rune('A'+i)),
			CaseType:    caseTypes[i%len(caseTypes)],
			Priority:    "normal",
			Status:      statuses[i%len(statuses)],
			ClientID:    client.ID,
			LawyerID:    lawyer.ID,
			Description: "测试案件描述",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(caseModel)
	}

	// 测试列表查询
	params := &CaseListParams{
		Page:     1,
		PageSize: 3,
	}
	cases, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, cases, 3)

	// 测试带条件的列表查询
	params.Status = "pending"
	cases, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total) // 应该有2个pending状态的案件
	assert.Len(t, cases, 2)
}

func TestCaseRepository_List_ByClient(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	client1 := createTestClientWithSuffix(db, "1")
	client2 := createTestClientWithSuffix(db, "2")
	lawyer := createTestUserWithSuffix(db, "1")

	// 为client1创建2个案件
	for i := 0; i < 2; i++ {
		caseModel := &models.Case{
			Title:       "Client1案件" + string(rune('A'+i)),
			CaseType:    "民事",
			Priority:    "normal",
			Status:      "pending",
			ClientID:    client1.ID,
			LawyerID:    lawyer.ID,
			Description: "测试案件描述",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(caseModel)
	}

	// 为client2创建1个案件
	caseModel := &models.Case{
		Title:       "Client2案件",
		CaseType:    "民事",
		Priority:    "normal",
		Status:      "pending",
		ClientID:    client2.ID,
		LawyerID:    lawyer.ID,
		Description: "测试案件描述",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(caseModel)

	// 测试按客户查询
	params := &CaseListParams{
		Page:     1,
		PageSize: 10,
		ClientID: client1.ID,
	}
	cases, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, cases, 2)

	// 验证所有案件都属于client1
	for _, caseItem := range cases {
		assert.Equal(t, client1.ID, caseItem.ClientID)
	}
}

func TestCaseRepository_List_ByLawyer(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	client := createTestClient(db)
	lawyer1 := createTestUserWithSuffix(db, "1")
	lawyer2 := createTestUserWithSuffix(db, "2")

	// 为lawyer1创建3个案件
	for i := 0; i < 3; i++ {
		caseModel := &models.Case{
			Title:       "Lawyer1案件" + string(rune('A'+i)),
			CaseType:    "民事",
			Priority:    "normal",
			Status:      "pending",
			ClientID:    client.ID,
			LawyerID:    lawyer1.ID,
			Description: "测试案件描述",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(caseModel)
	}

	// 为lawyer2创建1个案件
	caseModel := &models.Case{
		Title:       "Lawyer2案件",
		CaseType:    "民事",
		Priority:    "normal",
		Status:      "pending",
		ClientID:    client.ID,
		LawyerID:    lawyer2.ID,
		Description: "测试案件描述",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(caseModel)

	// 测试按律师查询
	params := &CaseListParams{
		Page:     1,
		PageSize: 10,
		LawyerID: lawyer1.ID,
	}
	cases, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, cases, 3)

	// 验证所有案件都属于lawyer1
	for _, caseItem := range cases {
		assert.Equal(t, lawyer1.ID, caseItem.LawyerID)
	}
}

func TestCaseRepository_List_Search(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	client := createTestClient(db)
	lawyer := createTestUserWithSuffix(db, "1")

	// 创建测试案件
	testCases := []*models.Case{
		{
			Title:       "合同纠纷案件",
			CaseType:    "民事",
			Priority:    "normal",
			Status:      "pending",
			ClientID:    client.ID,
			LawyerID:    lawyer.ID,
			Description: "涉及合同纠纷的民事案件",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "离婚诉讼案件",
			CaseType:    "民事",
			Priority:    "high",
			Status:      "in_progress",
			ClientID:    client.ID,
			LawyerID:    lawyer.ID,
			Description: "离婚相关诉讼",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			Title:       "刑事案件",
			CaseType:    "刑事",
			Priority:    "normal",
			Status:      "pending",
			ClientID:    client.ID,
			LawyerID:    lawyer.ID,
			Description: "刑事相关案件",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, caseModel := range testCases {
		db.Create(caseModel)
	}

	// 测试搜索标题
	params := &CaseListParams{
		Page:     1,
		PageSize: 10,
		Search:   "合同",
	}
	cases, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, cases, 1)
	assert.Contains(t, cases[0].Title, "合同")

	// 测试搜索描述
	params.Search = "诉讼"
	cases, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, cases, 1)
	assert.Contains(t, cases[0].Title, "离婚")
}

func TestCaseRepository_GetStats(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	client := createTestClient(db)
	lawyer := createTestUserWithSuffix(db, "1")

	// 创建不同状态的测试案件
	caseTypes := []string{"民事", "刑事", "行政"}
	statuses := []string{"pending", "active", "closed"}

	for i := 0; i < 6; i++ {
		caseModel := &models.Case{
			Title:       "统计案件" + string(rune('A'+i)),
			CaseType:    caseTypes[i%len(caseTypes)],
			Priority:    "normal",
			Status:      statuses[i%len(statuses)],
			ClientID:    client.ID,
			LawyerID:    lawyer.ID,
			Description: "统计测试案件",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(caseModel)
	}

	stats, err := repo.GetStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// 验证统计数据
	assert.Equal(t, int64(6), stats.TotalCases)
	assert.Equal(t, int64(2), stats.PendingCases)
	assert.Equal(t, int64(2), stats.ActiveCases)
	assert.Equal(t, int64(2), stats.ClosedCases)
}

func TestCaseRepository_AssignLawyer(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	client := createTestClient(db)
	lawyer1 := createTestUserWithSuffix(db, "1")
	lawyer2 := createTestUserWithSuffix(db, "2")

	testCase := createTestCase(db, client.ID, lawyer1.ID)

	// 测试分配律师
	err := repo.AssignLawyer(ctx, testCase.ID, lawyer2.ID)
	assert.NoError(t, err)

	// 验证律师分配是否成功
	var updatedCase models.Case
	err = db.First(&updatedCase, testCase.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, lawyer2.ID, updatedCase.LawyerID)
}

func TestCaseRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewCaseRepository(db)
	ctx := context.Background()

	testCase := createTestCase(db, createTestClientWithSuffix(db, "1").ID, createTestUserWithSuffix(db, "1").ID)

	// 测试更新状态
	err := repo.UpdateStatus(ctx, testCase.ID, "in_progress")
	assert.NoError(t, err)

	// 验证状态更新是否成功
	var updatedCase models.Case
	err = db.First(&updatedCase, testCase.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "in_progress", updatedCase.Status)
}
