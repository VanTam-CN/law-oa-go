package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
	"law-oa-go/test"
)

// setupTestServer 设置测试服务器
func setupTestServer(t *testing.T) (*gin.Engine, *gorm.DB, func()) {
	// 使用唯一的DSN确保每个测试都有独立的内存数据库
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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

	// 创建处理器
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	clientHandler := handlers.NewClientHandler(clientService)
	caseHandler := handlers.NewCaseHandler(caseService)

	// 设置路由器
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册路由
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/register", authHandler.Register)
	router.GET("/users/profile", authHandler.GetProfile)
	router.PUT("/users/profile", authHandler.UpdateProfile)
	router.POST("/users/change-password", authHandler.ChangePassword)
	router.POST("/auth/refresh", authHandler.RefreshToken)
	router.POST("/auth/logout", authHandler.Logout)

	router.GET("/users/:id", userHandler.GetUser)
	router.GET("/users", userHandler.ListUsers)
	router.POST("/users", userHandler.CreateUser)
	router.PUT("/users/:id", userHandler.UpdateUser)
	router.DELETE("/users/:id", userHandler.DeleteUser)

	router.GET("/clients/:id", clientHandler.GetClient)
	router.GET("/clients", clientHandler.ListClients)
	router.POST("/clients", clientHandler.CreateClient)
	router.PUT("/clients/:id", clientHandler.UpdateClient)
	router.DELETE("/clients/:id", clientHandler.DeleteClient)

	router.GET("/cases/:id", caseHandler.GetCase)
	router.GET("/cases", caseHandler.ListCases)
	router.POST("/cases", caseHandler.CreateCase)
	router.PUT("/cases/:id", caseHandler.UpdateCase)
	router.DELETE("/cases/:id", caseHandler.DeleteCase)

	// 清理函数
	cleanup := func() {
		db.Exec("DELETE FROM cases")
		db.Exec("DELETE FROM clients")
		db.Exec("DELETE FROM users")
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	return router, db, cleanup
}

// TestAuthIntegration 测试认证集成
func TestAuthIntegration(t *testing.T) {
	router, _, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("Complete Auth Flow", func(t *testing.T) {
		// 1. 注册用户
		registerReq := map[string]interface{}{
			"name":     "Test User",
			"email":    "testuser@example.com",
			"password": "Password123!",
			"role":     "user",
			"phone":    "1234567890",
		}

		reqBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 解析响应获取token
		var registerResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &registerResp)
		require.NoError(t, err)

		data := registerResp["data"].(map[string]interface{})
		token := data["token"].(string)

		// 2. 获取用户资料
		profileReq := httptest.NewRequest("GET", "/users/profile", nil)
		profileReq.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, profileReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 3. 更新用户资料
		updateReq := map[string]interface{}{
			"name":  "Updated Name",
			"phone": "9876543210",
		}

		updateBody, _ := json.Marshal(updateReq)
		updateHttpReq := httptest.NewRequest("PUT", "/users/profile", bytes.NewBuffer(updateBody))
		updateHttpReq.Header.Set("Content-Type", "application/json")
		updateHttpReq.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, updateHttpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 4. 修改密码
		changePasswordReq := map[string]interface{}{
			"current_password": "Password123!",
			"new_password":     "NewPassword123!",
		}

		changePasswordBody, _ := json.Marshal(changePasswordReq)
		changePasswordHttpReq := httptest.NewRequest("POST", "/users/change-password", bytes.NewBuffer(changePasswordBody))
		changePasswordHttpReq.Header.Set("Content-Type", "application/json")
		changePasswordHttpReq.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, changePasswordHttpReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 5. 登出
		logoutReq := httptest.NewRequest("POST", "/auth/logout", nil)
		logoutReq.Header.Set("Authorization", "Bearer "+token)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, logoutReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)
	})

	t.Run("Login with Different Roles", func(t *testing.T) {
		// 创建不同角色的用户
		users := []map[string]interface{}{
			{"name": "Admin User", "email": "admin@example.com", "password": "Password123!", "role": "admin"},
			{"name": "Lawyer User", "email": "lawyer@example.com", "password": "Password123!", "role": "lawyer"},
			{"name": "Regular User", "email": "regular@example.com", "password": "Password123!", "role": "user"},
		}

		for _, user := range users {
			// 注册用户
			reqBody, _ := json.Marshal(user)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			test.AssertSuccessResponse(t, w)

			// 登录
			loginReq := map[string]interface{}{
				"email":    user["email"],
				"password": user["password"],
			}

			loginBody, _ := json.Marshal(loginReq)
			loginHttpReq := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginBody))
			loginHttpReq.Header.Set("Content-Type", "application/json")

			w = httptest.NewRecorder()
			router.ServeHTTP(w, loginHttpReq)

			assert.Equal(t, http.StatusOK, w.Code)
			test.AssertSuccessResponse(t, w)
		}
	})
}

// TestUserManagementIntegration 测试用户管理集成
func TestUserManagementIntegration(t *testing.T) {
	router, _, cleanup := setupTestServer(t)
	defer cleanup()

	// 首先创建管理员用户
	adminReq := map[string]interface{}{
		"name":     "Admin User",
		"email":    "admin@example.com",
		"password": "Password123!",
		"role":     "admin",
	}

	reqBody, _ := json.Marshal(adminReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	test.AssertSuccessResponse(t, w)

	// 获取管理员token
	var registerResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &registerResp)
	require.NoError(t, err)

	data := registerResp["data"].(map[string]interface{})
	adminToken := data["token"].(string)

	t.Run("Create and Manage Users", func(t *testing.T) {
		// 1. 创建新用户
		newUser := map[string]interface{}{
			"name":     "New User",
			"email":    "newuser@example.com",
			"password": "Password123!",
			"role":     "user",
			"phone":    "1234567890",
		}

		userBody, _ := json.Marshal(newUser)
		createReq := httptest.NewRequest("POST", "/users", bytes.NewBuffer(userBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+adminToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, createReq)

		assert.Equal(t, http.StatusCreated, w.Code)
		test.AssertSuccessResponse(t, w)

		// 解析响应获取用户ID
		var createResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		require.NoError(t, err)

		// 2. 获取用户详情
		getReq := httptest.NewRequest("GET", "/users/1", nil)
		getReq.Header.Set("Authorization", "Bearer "+adminToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, getReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 3. 更新用户
		updateUser := map[string]interface{}{
			"name":  "Updated User",
			"phone": "9876543210",
		}

		updateBody, _ := json.Marshal(updateUser)
		updateReq := httptest.NewRequest("PUT", "/users/1", bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+adminToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, updateReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 4. 列出用户
		listReq := httptest.NewRequest("GET", "/users?page=1&page_size=10", nil)
		listReq.Header.Set("Authorization", "Bearer "+adminToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, listReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 5. 删除用户
		deleteReq := httptest.NewRequest("DELETE", "/users/1", nil)
		deleteReq.Header.Set("Authorization", "Bearer "+adminToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, deleteReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)
	})

}

// TestClientManagementIntegration 测试客户管理集成
func TestClientManagementIntegration(t *testing.T) {
	router, _, cleanup := setupTestServer(t)
	defer cleanup()

	// 创建律师用户
	lawyerReq := map[string]interface{}{
		"name":     "Lawyer User",
		"email":    "lawyer@example.com",
		"password": "Password123!",
		"role":     "lawyer",
	}

	reqBody, _ := json.Marshal(lawyerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	test.AssertSuccessResponse(t, w)

	// 获取律师token
	var registerResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &registerResp)
	require.NoError(t, err)

	data := registerResp["data"].(map[string]interface{})
	lawyerToken := data["token"].(string)

	t.Run("Create and Manage Clients", func(t *testing.T) {
		// 1. 创建客户
		newClient := map[string]interface{}{
			"name":    "Test Client",
			"email":   "client@example.com",
			"phone":   "1234567890",
			"address": "123 Test St",
			"company": "Test Company",
			"notes":   "Test client notes",
		}

		clientBody, _ := json.Marshal(newClient)
		createReq := httptest.NewRequest("POST", "/clients", bytes.NewBuffer(clientBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, createReq)

		assert.Equal(t, http.StatusCreated, w.Code)
		test.AssertSuccessResponse(t, w)

		// 2. 获取客户详情
		getReq := httptest.NewRequest("GET", "/clients/1", nil)
		getReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, getReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 3. 更新客户
		updateClient := map[string]interface{}{
			"name":    "Updated Client",
			"phone":   "9876543210",
			"address": "456 Updated St",
		}

		updateBody, _ := json.Marshal(updateClient)
		updateReq := httptest.NewRequest("PUT", "/clients/1", bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, updateReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 4. 列出客户
		listReq := httptest.NewRequest("GET", "/clients?page=1&page_size=10", nil)
		listReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, listReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 5. 删除客户
		deleteReq := httptest.NewRequest("DELETE", "/clients/1", nil)
		deleteReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, deleteReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)
	})

	t.Run("Client Search and Filter", func(t *testing.T) {
		// 创建多个客户用于测试搜索
		clients := []map[string]interface{}{
			{"name": "ABC Corp", "email": "abc@example.com", "company": "ABC Corp", "phone": "1111111111"},
			{"name": "XYZ Inc", "email": "xyz@example.com", "company": "XYZ Inc", "phone": "2222222222"},
			{"name": "Test Client", "email": "test@example.com", "company": "Test Company", "phone": "3333333333"},
		}

		for _, client := range clients {
			clientBody, _ := json.Marshal(client)
			createReq := httptest.NewRequest("POST", "/clients", bytes.NewBuffer(clientBody))
			createReq.Header.Set("Content-Type", "application/json")
			createReq.Header.Set("Authorization", "Bearer "+lawyerToken)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, createReq)

			assert.Equal(t, http.StatusCreated, w.Code)
		}

		// 测试搜索
		searchReq := httptest.NewRequest("GET", "/clients?search=ABC", nil)
		searchReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, searchReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 测试公司过滤
		companyReq := httptest.NewRequest("GET", "/clients?company=ABC+Corp", nil)
		companyReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, companyReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)
	})
}

// TestCaseManagementIntegration 测试案件管理集成
func TestCaseManagementIntegration(t *testing.T) {
	router, _, cleanup := setupTestServer(t)
	defer cleanup()

	// 创建律师和客户
	lawyerReq := map[string]interface{}{
		"name":     "Lawyer User",
		"email":    "lawyer@example.com",
		"password": "Password123!",
		"role":     "lawyer",
	}

	reqBody, _ := json.Marshal(lawyerReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	test.AssertSuccessResponse(t, w)

	var registerResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &registerResp)
	require.NoError(t, err)

	data := registerResp["data"].(map[string]interface{})
	lawyerToken := data["token"].(string)

	// 创建客户
	clientReq := map[string]interface{}{
		"name":    "Test Client",
		"email":   "client@example.com",
		"phone":   "1234567890",
		"address": "123 Test St",
	}

	clientBody, _ := json.Marshal(clientReq)
	createClientReq := httptest.NewRequest("POST", "/clients", bytes.NewBuffer(clientBody))
	createClientReq.Header.Set("Content-Type", "application/json")
	createClientReq.Header.Set("Authorization", "Bearer "+lawyerToken)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, createClientReq)

	assert.Equal(t, http.StatusCreated, w.Code)
	test.AssertSuccessResponse(t, w)

	t.Run("Create and Manage Cases", func(t *testing.T) {
		// 1. 创建案件
		newCase := map[string]interface{}{
			"title":       "Test Case",
			"description": "This is a test case description",
			"client_id":   1,
			"case_type":   "civil",
			"priority":    "medium",
		}

		caseBody, _ := json.Marshal(newCase)
		createReq := httptest.NewRequest("POST", "/cases", bytes.NewBuffer(caseBody))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, createReq)

		assert.Equal(t, http.StatusCreated, w.Code)
		test.AssertSuccessResponse(t, w)

		// 2. 获取案件详情
		getReq := httptest.NewRequest("GET", "/cases/1", nil)
		getReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, getReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 3. 更新案件
		updateCase := map[string]interface{}{
			"title":    "Updated Case Title",
			"priority": "high",
			"status":   "in_progress",
		}

		updateBody, _ := json.Marshal(updateCase)
		updateReq := httptest.NewRequest("PUT", "/cases/1", bytes.NewBuffer(updateBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, updateReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 4. 列出案件
		listReq := httptest.NewRequest("GET", "/cases?page=1&page_size=10", nil)
		listReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, listReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)

		// 5. 删除案件
		deleteReq := httptest.NewRequest("DELETE", "/cases/1", nil)
		deleteReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, deleteReq)

		assert.Equal(t, http.StatusOK, w.Code)
		test.AssertSuccessResponse(t, w)
	})
}

// TestAuthorizationIntegration 测试授权集成
func TestAuthorizationIntegration(t *testing.T) {
	router, _, cleanup := setupTestServer(t)
	defer cleanup()

	// 创建不同角色的用户
	users := []struct {
		name     string
		email    string
		password string
		role     string
	}{
		{"Admin User", "admin@example.com", "Password123!", "admin"},
		{"Lawyer User", "lawyer@example.com", "Password123!", "lawyer"},
		{"Regular User", "user@example.com", "Password123!", "user"},
	}

	var tokens []string
	for _, user := range users {
		userReq := map[string]interface{}{
			"name":     user.name,
			"email":    user.email,
			"password": user.password,
			"role":     user.role,
		}

		reqBody, _ := json.Marshal(userReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		tokens = append(tokens, data["token"].(string))
	}

	adminToken, lawyerToken, userToken := tokens[0], tokens[1], tokens[2]

	t.Run("Role-Based Access Control", func(t *testing.T) {
		// 测试用户创建权限
		createUserReq := map[string]interface{}{
			"name":     "New User",
			"email":    "newuser@example.com",
			"password": "Password123!",
			"role":     "user",
		}

		userBody, _ := json.Marshal(createUserReq)

		// 管理员应该能够创建用户
		adminReq := httptest.NewRequest("POST", "/users", bytes.NewBuffer(userBody))
		adminReq.Header.Set("Content-Type", "application/json")
		adminReq.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, adminReq)

		assert.Equal(t, http.StatusCreated, w.Code)

		// 律师不应该能够创建用户
		lawyerReq := httptest.NewRequest("POST", "/users", bytes.NewBuffer(userBody))
		lawyerReq.Header.Set("Content-Type", "application/json")
		lawyerReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, lawyerReq)

		assert.Equal(t, http.StatusForbidden, w.Code)

		// 普通用户不应该能够创建用户
		regularReq := httptest.NewRequest("POST", "/users", bytes.NewBuffer(userBody))
		regularReq.Header.Set("Content-Type", "application/json")
		regularReq.Header.Set("Authorization", "Bearer "+userToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, regularReq)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Resource Access Control", func(t *testing.T) {
		// 测试访问其他用户资料的权限

		// 管理员应该能够访问任何用户资料
		adminReq := httptest.NewRequest("GET", "/users/2", nil)
		adminReq.Header.Set("Authorization", "Bearer "+adminToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, adminReq)

		assert.Equal(t, http.StatusOK, w.Code)

		// 律师应该能够访问用户列表
		lawyerListReq := httptest.NewRequest("GET", "/users", nil)
		lawyerListReq.Header.Set("Authorization", "Bearer "+lawyerToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, lawyerListReq)

		assert.Equal(t, http.StatusOK, w.Code)

		// 普通用户不应该能够访问用户列表
		userListReq := httptest.NewRequest("GET", "/users", nil)
		userListReq.Header.Set("Authorization", "Bearer "+userToken)

		w = httptest.NewRecorder()
		router.ServeHTTP(w, userListReq)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
