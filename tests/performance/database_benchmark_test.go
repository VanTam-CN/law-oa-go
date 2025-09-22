package performance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

// DatabaseBenchmarkSuite 数据库基准测试套件
type DatabaseBenchmarkSuite struct {
	db            *gorm.DB
	userRepo      repositories.UserRepository
	clientRepo    repositories.ClientRepository
	caseRepo      repositories.CaseRepository
	userService   *services.UserService
	clientService *services.ClientService
	caseService   *services.CaseService
}

// SetupDatabaseBenchmarkSuite 设置数据库基准测试套件
func SetupDatabaseBenchmarkSuite(t *testing.T) *DatabaseBenchmarkSuite {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)

	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)
	caseService := services.NewCaseService(db)

	return &DatabaseBenchmarkSuite{
		db:            db,
		userRepo:      userRepo,
		clientRepo:    clientRepo,
		caseRepo:      caseRepo,
		userService:   userService,
		clientService: clientService,
		caseService:   caseService,
	}
}

// SetupDatabaseBenchmarkSuiteForB 为基准测试设置数据库套件
func SetupDatabaseBenchmarkSuiteForB(b *testing.B) *DatabaseBenchmarkSuite {
	// 创建一个临时的testing.T实例
	var t testing.T
	return SetupDatabaseBenchmarkSuite(&t)
}

// BenchmarkUserCreate 用户创建基准测试
func BenchmarkUserCreate(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &models.User{
			Name:     fmt.Sprintf("Benchmark User %d", i),
			Email:    fmt.Sprintf("bench%d@example.com", i%10000),
			Password: "Password123!",
			Role:     "user",
			Phone:    "1234567890",
			Status:   "active",
		}

		err := suite.userRepo.Create(context.Background(), user)
		if err != nil {
			b.Errorf("Failed to create user: %v", err)
		}
	}
}

// BenchmarkUserFindByEmail 通过邮箱查找用户基准测试
func BenchmarkUserFindByEmail(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	// 创建测试用户
	for i := 0; i < 1000; i++ {
		user := &models.User{
			Name:     fmt.Sprintf("Test User %d", i),
			Email:    fmt.Sprintf("test%d@example.com", i),
			Password: "Password123!",
			Role:     "user",
			Status:   "active",
		}
		suite.userRepo.Create(context.Background(), user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("test%d@example.com", i%1000)
		_, err := suite.userRepo.FindByEmail(context.Background(), email)
		if err != nil {
			b.Errorf("Failed to find user: %v", err)
		}
	}
}

// BenchmarkUserList 用户列表基准测试
func BenchmarkUserList(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	// 创建测试用户
	for i := 0; i < 500; i++ {
		user := &models.User{
			Name:     fmt.Sprintf("List User %d", i),
			Email:    fmt.Sprintf("list%d@example.com", i),
			Password: "Password123!",
			Role:     "user",
			Status:   "active",
		}
		suite.userRepo.Create(context.Background(), user)
	}

	params := &repositories.UserListParams{
		Page:     1,
		PageSize: 50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := suite.userRepo.List(context.Background(), params)
		if err != nil {
			b.Errorf("Failed to list users: %v", err)
		}
	}
}

// BenchmarkUserUpdate 用户更新基准测试
func BenchmarkUserUpdate(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	// 创建测试用户
	var users []*models.User
	for i := 0; i < 100; i++ {
		user := &models.User{
			Name:     fmt.Sprintf("Update User %d", i),
			Email:    fmt.Sprintf("update%d@example.com", i),
			Password: "Password123!",
			Role:     "user",
			Status:   "active",
		}
		suite.userRepo.Create(context.Background(), user)
		users = append(users, user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := users[i%len(users)]
		user.Name = fmt.Sprintf("Updated User %d", i)
		user.Phone = fmt.Sprintf("987654321%d", i%10)

		err := suite.userRepo.Update(context.Background(), user)
		if err != nil {
			b.Errorf("Failed to update user: %v", err)
		}
	}
}

// BenchmarkClientCreate 客户创建基准测试
func BenchmarkClientCreate(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := &models.Client{
			Name:    fmt.Sprintf("Benchmark Client %d", i),
			Email:   fmt.Sprintf("benchclient%d@example.com", i%10000),
			Phone:   "1234567890",
			Address: "123 Benchmark St",
			Company: "Benchmark Company",
			Status:  "active",
		}

		err := suite.clientRepo.Create(context.Background(), client)
		if err != nil {
			b.Errorf("Failed to create client: %v", err)
		}
	}
}

// BenchmarkClientListWithSearch 带搜索的客户列表基准测试
func BenchmarkClientListWithSearch(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	// 创建测试客户
	for i := 0; i < 500; i++ {
		client := &models.Client{
			Name:    fmt.Sprintf("Search Client %d", i),
			Email:   fmt.Sprintf("search%d@example.com", i),
			Phone:   "1234567890",
			Address: "123 Search St",
			Company: fmt.Sprintf("Search Company %d", i%10),
			Status:  "active",
		}
		suite.clientRepo.Create(context.Background(), client)
	}

	params := &repositories.ClientListParams{
		Page:     1,
		PageSize: 20,
		Search:   "Search",
		Company:  "Search Company",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := suite.clientRepo.List(context.Background(), params)
		if err != nil {
			b.Errorf("Failed to list clients with search: %v", err)
		}
	}
}

// BenchmarkCaseCreate 案件创建基准测试
func BenchmarkCaseCreate(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	// 创建测试律师和客户
	lawyer := &models.User{
		Name:     "Test Lawyer",
		Email:    "lawyer@example.com",
		Password: "Password123!",
		Role:     "lawyer",
		Status:   "active",
	}
	suite.userRepo.Create(context.Background(), lawyer)

	client := &models.Client{
		Name:    "Test Client",
		Email:   "client@example.com",
		Phone:   "1234567890",
		Address: "123 Test St",
		Status:  "active",
	}
	suite.clientRepo.Create(context.Background(), client)

	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		caseModel := &models.Case{
			Title:       fmt.Sprintf("Benchmark Case %d", i),
			Description: "This is a benchmark case description",
			ClientID:    client.ID,
			LawyerID:    lawyer.ID,
			CaseType:    "civil",
			Priority:    "medium",
			Status:      "pending",
			StartDate:   &now,
		}

		err := suite.caseRepo.Create(context.Background(), caseModel)
		if err != nil {
			b.Errorf("Failed to create case: %v", err)
		}
	}
}

// BenchmarkServiceLayerUserOperations 服务层用户操作基准测试
func BenchmarkServiceLayerUserOperations(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	b.Run("Create User", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := &services.CreateUserRequest{
				Name:     fmt.Sprintf("Service User %d", i),
				Email:    fmt.Sprintf("service%d@example.com", i%10000),
				Password: "Password123!",
				Role:     "user",
				Phone:    "1234567890",
			}

			_, err := suite.userService.CreateUser(context.Background(), req)
			if err != nil {
				b.Errorf("Failed to create user via service: %v", err)
			}
		}
	})

	// 创建一些用户用于其他测试
	for i := 0; i < 100; i++ {
		suite.userService.CreateUser(context.Background(), &services.CreateUserRequest{
			Name:     fmt.Sprintf("Service Test User %d", i),
			Email:    fmt.Sprintf("servicetest%d@example.com", i),
			Password: "Password123!",
			Role:     "user",
			Phone:    "1234567890",
		})
	}

	b.Run("List Users", func(b *testing.B) {
		req := &services.UserListRequest{
			Page:     1,
			PageSize: 50,
		}

		for i := 0; i < b.N; i++ {
			_, _, err := suite.userService.ListUsers(context.Background(), req)
			if err != nil {
				b.Errorf("Failed to list users via service: %v", err)
			}
		}
	})
}

// BenchmarkServiceLayerClientOperations 服务层客户操作基准测试
func BenchmarkServiceLayerClientOperations(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	b.Run("Create Client", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			req := &services.CreateClientRequest{
				Name:    fmt.Sprintf("Service Client %d", i),
				Email:   fmt.Sprintf("serviceclient%d@example.com", i%10000),
				Phone:   "1234567890",
				Address: "123 Service St",
				Company: "Service Company",
			}

			_, err := suite.clientService.CreateClient(context.Background(), req)
			if err != nil {
				b.Errorf("Failed to create client via service: %v", err)
			}
		}
	})

	// 创建一些客户用于其他测试
	for i := 0; i < 100; i++ {
		suite.clientService.CreateClient(context.Background(), &services.CreateClientRequest{
			Name:    fmt.Sprintf("Service Test Client %d", i),
			Email:   fmt.Sprintf("servicetestclient%d@example.com", i),
			Phone:   "1234567890",
			Address: "123 Service Test St",
			Company: "Service Test Company",
		})
	}

	b.Run("List Clients with Search", func(b *testing.B) {
		req := &services.ClientListRequest{
			Page:     1,
			PageSize: 20,
			Search:   "Service",
			Company:  "Service Company",
		}

		for i := 0; i < b.N; i++ {
			_, _, err := suite.clientService.ListClients(context.Background(), req)
			if err != nil {
				b.Errorf("Failed to list clients via service: %v", err)
			}
		}
	})
}

// BenchmarkBatchOperations 批量操作基准测试
func BenchmarkBatchOperations(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	b.Run("Batch Create Users", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			requests := make([]*services.CreateUserRequest, 10)
			for j := 0; j < 10; j++ {
				requests[j] = &services.CreateUserRequest{
					Name:     fmt.Sprintf("Batch User %d-%d", i, j),
					Email:    fmt.Sprintf("batch%d-%d@example.com", i, j),
					Password: "Password123!",
					Role:     "user",
					Phone:    "1234567890",
				}
			}

			_, err := suite.userService.BatchCreateUsers(context.Background(), requests)
			if err != nil {
				b.Errorf("Failed to batch create users: %v", err)
			}
		}
	})

	b.Run("Batch Update Users", func(b *testing.B) {
		// 创建一些用户用于更新
		var users []*models.User
		for i := 0; i < 50; i++ {
			user := &models.User{
				Name:     fmt.Sprintf("Batch Update User %d", i),
				Email:    fmt.Sprintf("batchupdate%d@example.com", i),
				Password: "Password123!",
				Role:     "user",
				Status:   "active",
			}
			suite.userRepo.Create(context.Background(), user)
			users = append(users, user)
		}

		for i := 0; i < b.N; i++ {
			updates := make(map[uint]*services.UpdateUserRequest)
			for j := 0; j < 10; j++ {
				user := users[j%len(users)]
				updates[user.ID] = &services.UpdateUserRequest{
					Name:  stringPtr(fmt.Sprintf("Updated User %d-%d", i, j)),
					Phone: stringPtr(fmt.Sprintf("987654321%d", j)),
				}
			}

			_, err := suite.userService.BatchUpdateUsers(context.Background(), updates)
			if err != nil {
				b.Errorf("Failed to batch update users: %v", err)
			}
		}
	})
}

// BenchmarkDatabaseTransactions 数据库事务基准测试
func BenchmarkDatabaseTransactions(b *testing.B) {
	suite := SetupDatabaseBenchmarkSuiteForB(b)

	b.Run("User Creation Transaction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tx := suite.db.Begin()
			user := &models.User{
				Name:     fmt.Sprintf("Transaction User %d", i),
				Email:    fmt.Sprintf("transaction%d@example.com", i%10000),
				Password: "Password123!",
				Role:     "user",
				Status:   "active",
			}

			err := tx.Create(user).Error
			if err != nil {
				tx.Rollback()
				b.Errorf("Failed to create user in transaction: %v", err)
				continue
			}

			err = tx.Commit().Error
			if err != nil {
				b.Errorf("Failed to commit transaction: %v", err)
			}
		}
	})

	b.Run("Multi-Table Transaction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tx := suite.db.Begin()

			// 创建用户
			user := &models.User{
				Name:     fmt.Sprintf("Multi-User %d", i),
				Email:    fmt.Sprintf("multi%d@example.com", i%10000),
				Password: "Password123!",
				Role:     "lawyer",
				Status:   "active",
			}
			err := tx.Create(user).Error
			if err != nil {
				tx.Rollback()
				b.Errorf("Failed to create user: %v", err)
				continue
			}

			// 创建客户
			client := &models.Client{
				Name:    fmt.Sprintf("Multi-Client %d", i),
				Email:   fmt.Sprintf("multi%d@example.com", i%10000),
				Phone:   "1234567890",
				Address: "123 Multi St",
				Status:  "active",
			}
			err = tx.Create(client).Error
			if err != nil {
				tx.Rollback()
				b.Errorf("Failed to create client: %v", err)
				continue
			}

			// 创建案件
			now := time.Now()
			caseModel := &models.Case{
				Title:       fmt.Sprintf("Multi-Case %d", i),
				Description: "Multi-table transaction case",
				ClientID:    client.ID,
				LawyerID:    user.ID,
				CaseType:    "civil",
				Priority:    "medium",
				Status:      "pending",
				StartDate:   &now,
			}
			err = tx.Create(caseModel).Error
			if err != nil {
				tx.Rollback()
				b.Errorf("Failed to create case: %v", err)
				continue
			}

			err = tx.Commit().Error
			if err != nil {
				b.Errorf("Failed to commit multi-table transaction: %v", err)
			}
		}
	})
}

// 辅助函数：字符串指针
func stringPtr(s string) *string {
	return &s
}
