//go:build ignore
// +build ignore

package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"law-oa-go/internal/models"
)

func TestClientRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	client := &models.Client{
		Name:      "测试客户",
		Email:     "test@example.com",
		Phone:     "13900139000",
		Address:   "测试地址",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, client)
	assert.NoError(t, err)
	assert.NotZero(t, client.ID)

	// 验证客户确实被创建
	var foundClient models.Client
	err = db.First(&foundClient, client.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, client.Name, foundClient.Name)
	assert.Equal(t, client.Email, foundClient.Email)
}

func TestClientRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建测试客户
	testClient := createTestClient(db)

	// 测试查找存在的客户
	foundClient, err := repo.FindByID(ctx, testClient.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundClient)
	assert.Equal(t, testClient.Name, foundClient.Name)

	// 测试查找不存在的客户
	notFoundClient, err := repo.FindByID(ctx, 999999)
	assert.NoError(t, err) // ClientRepository.FindByID在找不到时返回nil而不是error
	assert.Nil(t, notFoundClient)
}

func TestClientRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建测试客户
	testClient := createTestClient(db)

	// 测试查找存在的客户
	foundClient, err := repo.FindByEmail(ctx, testClient.Email)
	assert.NoError(t, err)
	assert.NotNil(t, foundClient)
	assert.Equal(t, testClient.Name, foundClient.Name)

	// 测试查找不存在的客户
	notFoundClient, err := repo.FindByEmail(ctx, "nonexistent@example.com")
	assert.NoError(t, err)
	assert.Nil(t, notFoundClient)
}

func TestClientRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	testClient := createTestClient(db)

	// 测试更新客户
	testClient.Name = "更新后的客户名"
	testClient.Status = "inactive"
	testClient.Address = "更新的地址"

	err := repo.Update(ctx, testClient)
	assert.NoError(t, err)

	// 验证更新是否成功
	var updatedClient models.Client
	err = db.First(&updatedClient, testClient.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "更新后的客户名", updatedClient.Name)
	assert.Equal(t, "inactive", updatedClient.Status)
	assert.Equal(t, "更新的地址", updatedClient.Address)
}

func TestClientRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	testClient := createTestClient(db)

	// 测试删除客户
	err := repo.Delete(ctx, testClient.ID)
	assert.NoError(t, err)

	// 验证客户是否被删除
	var deletedClient models.Client
	err = db.First(&deletedClient, testClient.ID).Error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestClientRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建多个测试客户
	statuses := []string{"active", "inactive"}

	for i := 0; i < 5; i++ {
		client := &models.Client{
			Name:      "客户" + string(rune('A'+i)),
			Email:     "client" + string(rune('A'+i)) + "@example.com",
			Phone:     "13900139000",
			Address:   "测试地址",
			Status:    statuses[i%len(statuses)],
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(client)
	}

	// 测试列表查询
	params := &ClientListParams{
		Page:     1,
		PageSize: 3,
	}
	clients, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, clients, 3)

	// 测试带条件的列表查询
	params.Status = "active"
	clients, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total) // 应该有3个active状态的客户
	assert.Len(t, clients, 3)
}

func TestClientRepository_List_Search(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建测试客户
	testClients := []*models.Client{
		{
			Name:      "张三客户",
			Email:     "zhangsan@example.com",
			Phone:     "13900139000",
			Address:   "北京地址",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:      "李四客户",
			Email:     "lisi@example.com",
			Phone:     "13800138000",
			Address:   "上海地址",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Name:      "王五客户",
			Email:     "wangwu@example.com",
			Phone:     "13700137000",
			Address:   "广州地址",
			Status:    "inactive",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, client := range testClients {
		db.Create(client)
	}

	// 测试搜索名称
	params := &ClientListParams{
		Page:     1,
		PageSize: 10,
		Search:   "张三",
	}
	clients, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, clients, 1)
	assert.Contains(t, clients[0].Name, "张三")

	// 测试搜索邮箱
	params.Search = "lisi"
	clients, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, clients, 1)
	assert.Contains(t, clients[0].Email, "lisi")

	// 测试搜索电话
	params.Search = "138"
	clients, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, clients, 1)
	assert.Contains(t, clients[0].Phone, "138")
}

func TestClientRepository_GetStats(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建不同状态的测试客户
	statuses := []string{"active", "inactive", "active", "inactive", "active"}

	for i, status := range statuses {
		client := &models.Client{
			Name:      "统计客户" + string(rune('A'+i)),
			Email:     "stats" + string(rune('A'+i)) + "@example.com",
			Phone:     "13900139000",
			Address:   "测试地址",
			Status:    status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(client)
	}

	stats, err := repo.GetStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// 验证统计数据
	assert.Equal(t, int64(5), stats.TotalClients)
	assert.Equal(t, int64(3), stats.ActiveClients)
	assert.Equal(t, int64(2), stats.InactiveClients)
	assert.GreaterOrEqual(t, stats.NewClientsThisMonth, int64(5)) // 至少包含刚创建的5个客户
}

func TestClientRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建第一个客户
	client1 := &models.Client{
		Name:      "客户1",
		Email:     "same@example.com",
		Phone:     "13900139000",
		Address:   "地址1",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := repo.Create(ctx, client1)
	assert.NoError(t, err)

	// 尝试创建相同邮箱的客户
	client2 := &models.Client{
		Name:      "客户2",
		Email:     "same@example.com", // 相同邮箱
		Phone:     "13900139001",
		Address:   "地址2",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, client2)
	assert.Error(t, err) // 应该失败，邮箱唯一约束
}

func TestClientRepository_List_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	params := &ClientListParams{
		Page:     1,
		PageSize: 10,
	}
	clients, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, clients)
}

func TestClientRepository_List_InvalidPagination(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建测试客户
	_ = createTestClient(db) // 创建一个客户用于测试

	// 测试无效的分页参数（应该使用默认值）
	params := &ClientListParams{
		Page:     -1, // 无效页码
		PageSize: 0,  // 无效页面大小
	}
	clients, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.NotEmpty(t, clients)
}

func TestClientRepository_List_SearchCaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewClientRepository(db)
	ctx := context.Background()

	// 创建大小写混合的客户名称
	client := &models.Client{
		Name:      "Test Client Case",
		Email:     "test@example.com",
		Phone:     "13900139000",
		Address:   "Test Address",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	db.Create(client)

	// 测试小写搜索
	params := &ClientListParams{
		Page:     1,
		PageSize: 10,
		Search:   "test client",
	}
	clients, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, clients, 1)
	assert.Equal(t, "Test Client Case", clients[0].Name)

	// 测试大写搜索
	params.Search = "TEST CLIENT"
	clients, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, clients, 1)
	assert.Equal(t, "Test Client Case", clients[0].Name)
}
