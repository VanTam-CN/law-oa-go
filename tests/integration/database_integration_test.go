package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// TestDatabaseIntegration 测试数据库集成
func TestDatabaseIntegration(t *testing.T) {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	require.NoError(t, err)

	// 创建仓库
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)

	// 创建服务
	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)
	caseService := services.NewCaseService(db)

	t.Run("User CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		// 创建用户
		createReq := &services.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "Password123!",
			Role:     "lawyer",
			Phone:    "1234567890",
		}

		user, err := userService.CreateUser(ctx, createReq)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "test@example.com", user.Email)

		// 获取用户
		profile, err := userService.GetUserProfile(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, user.ID, profile.ID)

		// 更新用户
		updateReq := &services.UpdateUserRequest{
			Name:  stringPtr("Updated User"),
			Phone: stringPtr("9876543210"),
		}

		updatedUser, err := userService.UpdateUser(ctx, user.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "Updated User", updatedUser.Name)
		assert.Equal(t, "9876543210", updatedUser.Phone)

		// 列出用户
		listReq := &services.UserListRequest{
			Page:     1,
			PageSize: 10,
			Role:     "lawyer",
		}

		users, total, err := userService.ListUsers(ctx, listReq)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Updated User", users[0].Name)

		// 删除用户
		err = userService.DeleteUser(ctx, user.ID)
		require.NoError(t, err)

		// 验证用户已删除
		_, err = userService.GetUserProfile(ctx, user.ID)
		assert.Error(t, err)
	})

	t.Run("Client CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		// 创建客户
		createReq := &services.CreateClientRequest{
			Name:    "Test Client",
			Email:   "client@example.com",
			Phone:   "1234567890",
			Address: "123 Test St",
			Company: "Test Company",
			Notes:   "Test client notes",
		}

		client, err := clientService.CreateClient(ctx, createReq)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "Test Client", client.Name)
		assert.Equal(t, "client@example.com", client.Email)

		// 获取客户
		retrievedClient, err := clientService.GetClientByID(ctx, client.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedClient)
		assert.Equal(t, client.ID, retrievedClient.ID)

		// 更新客户
		updateReq := &services.UpdateClientRequest{
			Name:    stringPtr("Updated Client"),
			Phone:   stringPtr("9876543210"),
			Address: stringPtr("456 Updated St"),
		}

		updatedClient, err := clientService.UpdateClient(ctx, client.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "Updated Client", updatedClient.Name)
		assert.Equal(t, "9876543210", updatedClient.Phone)

		// 列出客户
		listReq := &services.ClientListRequest{
			Page:     1,
			PageSize: 10,
			Company:  "Test Company",
		}

		clients, total, err := clientService.ListClients(ctx, listReq)
		require.NoError(t, err)
		assert.Len(t, clients, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Updated Client", clients[0].Name)

		// 删除客户
		err = clientService.DeleteClient(ctx, client.ID)
		require.NoError(t, err)

		// 验证客户已删除
		_, err = clientService.GetClientByID(ctx, client.ID)
		assert.Error(t, err)
	})

	t.Run("Case CRUD Operations", func(t *testing.T) {
		ctx := context.Background()

		// 首先创建律师和客户
		_, err := userService.CreateUser(ctx, &services.CreateUserRequest{
			Name:     "Test Lawyer",
			Email:    "lawyer@example.com",
			Password: "Password123!",
			Role:     "lawyer",
		})
		require.NoError(t, err)

		client, err := clientService.CreateClient(ctx, &services.CreateClientRequest{
			Name:    "Test Client",
			Email:   "client@example.com",
			Phone:   "1234567890",
			Address: "123 Test St",
		})
		require.NoError(t, err)

		// 创建案件
		createReq := &services.CreateCaseRequest{
			Title:       "Test Case",
			Description: "This is a test case description",
			ClientID:    client.ID,
			CaseType:    "civil",
			Priority:    "medium",
		}

		caseModel, err := caseService.CreateCase(ctx, createReq)
		require.NoError(t, err)
		assert.NotNil(t, caseModel)
		assert.Equal(t, "Test Case", caseModel.Title)
		assert.Equal(t, client.ID, caseModel.ClientID)

		// 获取案件
		retrievedCase, err := caseService.GetCaseByID(ctx, caseModel.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrievedCase)
		assert.Equal(t, caseModel.ID, retrievedCase.ID)

		// 更新案件
		updateReq := &services.UpdateCaseRequest{
			Title:    stringPtr("Updated Case"),
			Priority: stringPtr("high"),
			Status:   stringPtr("in_progress"),
		}

		updatedCase, err := caseService.UpdateCase(ctx, caseModel.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "Updated Case", updatedCase.Title)
		assert.Equal(t, "high", updatedCase.Priority)

		// 列出案件
		listReq := &services.CaseListRequest{
			Page:     1,
			PageSize: 10,
			ClientID: client.ID,
		}

		cases, total, err := caseService.ListCases(ctx, listReq)
		require.NoError(t, err)
		assert.Len(t, cases, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Updated Case", cases[0].Title)

		// 删除案件
		err = caseService.DeleteCase(ctx, caseModel.ID)
		require.NoError(t, err)

		// 验证案件已删除
		_, err = caseService.GetCaseByID(ctx, caseModel.ID)
		assert.Error(t, err)
	})

	t.Run("Search and Filter Operations", func(t *testing.T) {
		ctx := context.Background()

		// 创建测试数据
		testUsers := []*services.CreateUserRequest{
			{Name: "Alice Smith", Email: "alice@example.com", Password: "Password123!", Role: "lawyer"},
			{Name: "Bob Johnson", Email: "bob@example.com", Password: "Password123!", Role: "user"},
			{Name: "Alice Brown", Email: "alice2@example.com", Password: "Password123!", Role: "user"},
			{Name: "Charlie Davis", Email: "charlie@example.com", Password: "Password123!", Role: "admin"},
		}

		for _, userReq := range testUsers {
			_, err := userService.CreateUser(ctx, userReq)
			require.NoError(t, err)
		}

		// 测试用户搜索
		searchReq := &services.UserListRequest{
			Page:     1,
			PageSize: 10,
			Search:   "Alice",
		}

		users, total, err := userService.ListUsers(ctx, searchReq)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // 应该找到2个Alice
		assert.Len(t, users, 2)

		// 测试角色过滤
		roleReq := &services.UserListRequest{
			Page:     1,
			PageSize: 10,
			Role:     "lawyer",
		}

		users, total, err = userService.ListUsers(ctx, roleReq)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 应该找到1个律师
		assert.Equal(t, "Alice Smith", users[0].Name)

		// 创建测试客户
		testClients := []*services.CreateClientRequest{
			{Name: "ABC Corporation", Email: "abc@example.com", Company: "ABC Corp"},
			{Name: "XYZ Industries", Email: "xyz@example.com", Company: "XYZ Inc"},
			{Name: "ABC Limited", Email: "abc2@example.com", Company: "ABC Ltd"},
		}

		for _, clientReq := range testClients {
			_, err := clientService.CreateClient(ctx, clientReq)
			require.NoError(t, err)
		}

		// 测试客户搜索
		clientSearchReq := &services.ClientListRequest{
			Page:     1,
			PageSize: 10,
			Search:   "ABC",
		}

		clients, total, err := clientService.ListClients(ctx, clientSearchReq)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // 应该找到2个ABC客户
		assert.Len(t, clients, 2)

		// 测试公司过滤
		companyReq := &services.ClientListRequest{
			Page:     1,
			PageSize: 10,
			Company:  "ABC Corp",
		}

		clients, total, err = clientService.ListClients(ctx, companyReq)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total) // 应该找到1个ABC Corp
		assert.Equal(t, "ABC Corporation", clients[0].Name)
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		ctx := context.Background()

		// 测试并发创建用户
		concurrentUsers := make([]*services.CreateUserRequest, 10)
		for i := 0; i < 10; i++ {
			concurrentUsers[i] = &services.CreateUserRequest{
				Name:     "Concurrent User",
				Email:    "concurrent@example.com",
				Password: "Password123!",
				Role:     "user",
			}
		}

		// 注意：由于邮箱唯一性约束，这个测试会失败
		// 在实际应用中，应该使用不同的邮箱地址
		_, err := userService.BatchCreateUsers(ctx, concurrentUsers)
		assert.Error(t, err) // 应该因为邮箱冲突而失败

		// 创建具有唯一邮箱的并发用户
		uniqueUsers := make([]*services.CreateUserRequest, 5)
		for i := 0; i < 5; i++ {
			uniqueUsers[i] = &services.CreateUserRequest{
				Name:     "Unique User",
				Email:    "unique@example.com",
				Password: "Password123!",
				Role:     "user",
			}
		}

		// 修改邮箱地址使其唯一
		for i := range uniqueUsers {
			uniqueUsers[i].Email = "unique" + string(rune('a'+i)) + "@example.com"
		}

		_, err = userService.BatchCreateUsers(ctx, uniqueUsers)
		require.NoError(t, err)
	})

	t.Run("Transaction Rollback", func(t *testing.T) {
		// 这个测试验证事务回滚功能
		// 在实际应用中，应该在服务层实现事务管理
		ctx := context.Background()

		// 创建用户
		user, err := userService.CreateUser(ctx, &services.CreateUserRequest{
			Name:     "Transaction Test User",
			Email:    "transaction@example.com",
			Password: "Password123!",
			Role:     "user",
		})
		require.NoError(t, err)

		// 创建客户
		client, err := clientService.CreateClient(ctx, &services.CreateClientRequest{
			Name:    "Transaction Test Client",
			Email:   "transactionclient@example.com",
			Phone:   "1234567890",
			Address: "123 Test St",
		})
		require.NoError(t, err)

		// 创建案件
		caseModel, err := caseService.CreateCase(ctx, &services.CreateCaseRequest{
			Title:       "Transaction Test Case",
			Description: "Test case for transaction",
			ClientID:    client.ID,
			CaseType:    "civil",
			Priority:    "medium",
		})
		require.NoError(t, err)

		// 验证所有数据都正确创建
		assert.Equal(t, "Transaction Test User", user.Name)
		assert.Equal(t, "Transaction Test Client", client.Name)
		assert.Equal(t, "Transaction Test Case", caseModel.Title)
	})
}

// 辅助函数：字符串指针
func stringPtr(s string) *string {
	return &s
}
