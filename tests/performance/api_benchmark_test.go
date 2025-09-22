package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/handlers"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
	"law-oa-go/test"
)

// APILoadTestSuite API负载测试套件
type APILoadTestSuite struct {
	router      *gin.Engine
	db          *gorm.DB
	userRepo    repositories.UserRepository
	userService *services.UserService
	auth        *handlers.AuthHandler
}

// SetupAPITestSuite 设置API测试套件
func SetupAPITestSuite(t *testing.T) *APILoadTestSuite {
	// 设置测试数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	require.NoError(t, err)

	// 创建仓库
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)

	// 创建服务
	userService := services.NewUserService(userRepo)
	clientService := services.NewClientService(clientRepo)

	// 创建处理器
	authHandler := handlers.NewAuthHandler(userService)
	clientHandler := handlers.NewClientHandler(clientService)

	// 设置路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 认证路由组
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.POST("/logout", middleware.AuthMiddleware(), authHandler.Logout)
	}

	// 用户路由组
	userGroup := router.Group("/users")
	userGroup.Use(middleware.AuthMiddleware())
	{
		userGroup.GET("/profile", authHandler.GetProfile)
		userGroup.PUT("/profile", authHandler.UpdateProfile)
	}

	// 客户路由组
	clientGroup := router.Group("/clients")
	clientGroup.Use(middleware.AuthMiddleware())
	{
		clientGroup.GET("/", clientHandler.ListClients)
		clientGroup.POST("/", clientHandler.CreateClient)
		clientGroup.GET("/:id", clientHandler.GetClient)
		clientGroup.PUT("/:id", clientHandler.UpdateClient)
		clientGroup.DELETE("/:id", clientHandler.DeleteClient)
	}

	return &APILoadTestSuite{
		router:      router,
		db:          db,
		userRepo:    userRepo,
		userService: userService,
		auth:        authHandler,
	}
}

// SetupAPITestSuiteForB 为基准测试设置API测试套件
func SetupAPITestSuiteForB(b *testing.B) *APILoadTestSuite {
	var t testing.T
	return SetupAPITestSuite(&t)
}

// BenchmarkLogin API登录基准测试
func BenchmarkLogin(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	// 创建测试用户
	_ = test.CreateTestUserB(b)

	// 准备请求体
	loginData := map[string]interface{}{
		"email":    "benchmark@example.com",
		"password": "Password123!",
	}
	jsonData, _ := json.Marshal(loginData)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			suite.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Login failed with status: %d", w.Code)
			}
		}
	})
}

// BenchmarkCreateUser API创建用户基准测试
func BenchmarkCreateUser(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			email := fmt.Sprintf("benchmark%d@example.com", i%1000)
			userData := map[string]interface{}{
				"name":     fmt.Sprintf("Benchmark User %d", i),
				"email":    email,
				"password": "Password123!",
				"role":     "user",
				"phone":    "1234567890",
			}
			jsonData, _ := json.Marshal(userData)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			suite.router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				b.Errorf("User creation failed with status: %d", w.Code)
			}
			i++
		}
	})
}

// BenchmarkGetUserProfile API获取用户资料基准测试
func BenchmarkGetUserProfile(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	// 创建测试用户
	_ = test.CreateTestUserB(b)

	// 获取认证令牌
	// 简单的模拟token，实际测试中使用
	token := "test-token-" + test.RandomString(16)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/users/profile", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			suite.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Get profile failed with status: %d", w.Code)
			}
		}
	})
}

// BenchmarkCreateClient API创建客户基准测试
func BenchmarkCreateClient(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	// 创建测试用户和令牌
	_ = test.CreateTestUser(&testing.T{})
	token := "test-token-" + test.RandomString(16)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			clientData := map[string]interface{}{
				"name":    fmt.Sprintf("Benchmark Client %d", i),
				"email":   fmt.Sprintf("client%d@example.com", i%1000),
				"phone":   "1234567890",
				"address": "123 Benchmark St",
				"company": "Benchmark Company",
			}
			jsonData, _ := json.Marshal(clientData)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/clients/", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			suite.router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				b.Errorf("Client creation failed with status: %d", w.Code)
			}
			i++
		}
	})
}

// BenchmarkListClients API列表客户基准测试
func BenchmarkListClients(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	// 创建测试用户和令牌
	_ = test.CreateTestUser(&testing.T{})
	token := "test-token-" + test.RandomString(16)

	// 创建一些测试客户（简化版本）

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/clients/?page=1&pageSize=10", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			suite.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("List clients failed with status: %d", w.Code)
			}
		}
	})
}

// BenchmarkConcurrentRequests 并发请求基准测试
func BenchmarkConcurrentRequests(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	// 创建测试用户和令牌
	_ = test.CreateTestUser(&testing.T{})
	token := "test-token-" + test.RandomString(16)

	// 创建一些测试客户（简化版本）

	requests := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"GET", "/users/profile", nil},
		{"GET", "/clients/?page=1&pageSize=5", nil},
		{"POST", "/clients/", map[string]interface{}{
			"name":    "New Client",
			"email":   "new@example.com",
			"phone":   "1234567890",
			"address": "123 Test St",
		}},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			reqData := requests[i%len(requests)]
			var body *bytes.Buffer
			if reqData.body != nil {
				jsonData, _ := json.Marshal(reqData.body)
				body = bytes.NewBuffer(jsonData)
			} else {
				body = bytes.NewBuffer(nil)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(reqData.method, reqData.path, body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			suite.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK && w.Code != http.StatusCreated {
				b.Errorf("Request failed with status: %d", w.Code)
			}
			i++
		}
	})
}

// BenchmarkAuthenticationMiddleware 认证中间件基准测试
func BenchmarkAuthenticationMiddleware(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	// 创建测试用户和有效令牌
	_ = test.CreateTestUser(&testing.T{})
	validToken := "test-token-" + test.RandomString(16)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/users/profile", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+validToken)

			suite.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Authentication middleware failed with status: %d", w.Code)
			}
		}
	})
}

// BenchmarkDatabaseQueries 数据库查询基准测试
func BenchmarkDatabaseQueries(b *testing.B) {
	suite := SetupAPITestSuiteForB(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 创建用户
			email := fmt.Sprintf("dbtest%d@example.com", i%1000)
			user := &models.User{
				Name:     fmt.Sprintf("DB Test User %d", i),
				Email:    email,
				Password: "Password123!",
				Role:     "user",
				Phone:    "1234567890",
				Status:   "active",
			}

			err := suite.userRepo.Create(context.Background(), user)
			if err != nil {
				b.Errorf("User creation failed: %v", err)
			}

			// 查询用户
			_, err = suite.userRepo.FindByEmail(context.Background(), email)
			if err != nil {
				b.Errorf("User find failed: %v", err)
			}

			// 删除用户
			err = suite.userRepo.Delete(context.Background(), user.ID)
			if err != nil {
				b.Errorf("User deletion failed: %v", err)
			}
			i++
		}
	})
}
