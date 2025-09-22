package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
	test "law-oa-go/tests"
)

// E2ETestSuite E2E测试套件
type E2ETestSuite struct {
	router   *gin.Engine
	db       *gorm.DB
	baseURL  string
	client   *http.Client
	testUser *models.User
	testAuth map[string]string // tokens for different roles
}

// NewE2ETestSuite 创建新的E2E测试套件
func NewE2ETestSuite(t *testing.T) *E2ETestSuite {
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

	// 创建处理器
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	clientHandler := handlers.NewClientHandler(clientService)
	caseHandler := handlers.NewCaseHandler(caseService)

	// 设置路由器
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册路由 - 认证相关
	router.POST("/api/v1/auth/login", authHandler.Login)
	router.POST("/api/v1/auth/register", authHandler.Register)
	router.GET("/api/v1/users/profile", authHandler.GetProfile)
	router.PUT("/api/v1/users/profile", authHandler.UpdateProfile)
	router.POST("/api/v1/users/change-password", authHandler.ChangePassword)
	router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)
	router.POST("/api/v1/auth/logout", authHandler.Logout)

	// 注册路由 - 用户管理
	router.GET("/api/v1/users", userHandler.ListUsers)
	router.POST("/api/v1/users", userHandler.CreateUser)
	router.GET("/api/v1/users/:id", userHandler.GetUser)
	router.PUT("/api/v1/users/:id", userHandler.UpdateUser)
	router.DELETE("/api/v1/users/:id", userHandler.DeleteUser)

	// 注册路由 - 客户管理
	router.GET("/api/v1/clients", clientHandler.ListClients)
	router.POST("/api/v1/clients", clientHandler.CreateClient)
	router.GET("/api/v1/clients/:id", clientHandler.GetClient)
	router.PUT("/api/v1/clients/:id", clientHandler.UpdateClient)
	router.DELETE("/api/v1/clients/:id", clientHandler.DeleteClient)

	// 注册路由 - 案件管理
	router.GET("/api/v1/cases", caseHandler.ListCases)
	router.POST("/api/v1/cases", caseHandler.CreateCase)
	router.GET("/api/v1/cases/:id", caseHandler.GetCase)
	router.PUT("/api/v1/cases/:id", caseHandler.UpdateCase)
	router.DELETE("/api/v1/cases/:id", caseHandler.DeleteCase)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	suite := &E2ETestSuite{
		router:   router,
		db:       db,
		baseURL:  "/api/v1",
		client:   &http.Client{Timeout: 30 * time.Second},
		testAuth: make(map[string]string),
	}

	// 创建测试用户
	suite.setupTestUsers(t)

	return suite
}

// setupTestUsers 设置测试用户
func (suite *E2ETestSuite) setupTestUsers(t *testing.T) {
	testUsers := []struct {
		name     string
		email    string
		password string
		role     string
	}{
		{"Admin User", "admin@test.com", "AdminPassword123!", "admin"},
		{"Lawyer User", "lawyer@test.com", "LawyerPassword123!", "lawyer"},
		{"Regular User", "user@test.com", "UserPassword123!", "user"},
	}

	for _, user := range testUsers {
		// 注册用户
		registerData := map[string]interface{}{
			"name":     user.name,
			"email":    user.email,
			"password": user.password,
			"role":     user.role,
		}

		resp := suite.postRequest(t, "/auth/register", registerData, nil)
		assert.Equal(t, http.StatusOK, resp.Code)

		// 登录获取token
		loginData := map[string]interface{}{
			"email":    user.email,
			"password": user.password,
		}

		resp = suite.postRequest(t, "/auth/login", loginData, nil)
		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		require.NoError(t, err)

		data := result["data"].(map[string]interface{})
		suite.testAuth[user.role] = data["token"].(string)
	}
}

// TearDown 清理测试资源
func (suite *E2ETestSuite) TearDown() {
	if suite.db != nil {
		sqlDB, err := suite.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// postRequest 发送POST请求
func (suite *E2ETestSuite) postRequest(t *testing.T, path string, data interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var body bytes.Buffer
	if data != nil {
		json.NewEncoder(&body).Encode(data)
	}

	req := httptest.NewRequest("POST", suite.baseURL+path, &body)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// getRequest 发送GET请求
func (suite *E2ETestSuite) getRequest(t *testing.T, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", suite.baseURL+path, nil)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// putRequest 发送PUT请求
func (suite *E2ETestSuite) putRequest(t *testing.T, path string, data interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var body bytes.Buffer
	if data != nil {
		json.NewEncoder(&body).Encode(data)
	}

	req := httptest.NewRequest("PUT", suite.baseURL+path, &body)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// deleteRequest 发送DELETE请求
func (suite *E2ETestSuite) deleteRequest(t *testing.T, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", suite.baseURL+path, nil)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// getAuthHeaders 获取认证头
func (suite *E2ETestSuite) getAuthHeaders(role string) map[string]string {
	token, exists := suite.testAuth[role]
	if !exists {
		return map[string]string{}
	}
	return map[string]string{"Authorization": "Bearer " + token}
}

// TestCompleteUserWorkflow 测试完整的用户工作流程
func TestCompleteUserWorkflow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("New User Onboarding", func(t *testing.T) {
		// 1. 新用户注册
		newUser := map[string]interface{}{
			"name":     "John Doe",
			"email":    "john.doe@example.com",
			"password": "JohnPassword123!",
			"role":     "user",
			"phone":    "1234567890",
		}

		resp := suite.postRequest(t, "/auth/register", newUser, nil)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 解析注册响应
		var registerResp map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &registerResp)
		require.NoError(t, err)

		data := registerResp["data"].(map[string]interface{})
		token := data["token"].(string)

		authHeaders := map[string]string{"Authorization": "Bearer " + token}

		// 2. 获取用户资料
		resp = suite.getRequest(t, "/users/profile", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 3. 更新用户资料
		updateProfile := map[string]interface{}{
			"name":  "John Smith",
			"phone": "0987654321",
		}

		resp = suite.putRequest(t, "/users/profile", updateProfile, authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 4. 修改密码
		changePassword := map[string]interface{}{
			"current_password": "JohnPassword123!",
			"new_password":     "NewJohnPassword123!",
		}

		resp = suite.postRequest(t, "/users/change-password", changePassword, authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 5. 使用新密码登录
		loginReq := map[string]interface{}{
			"email":    "john.doe@example.com",
			"password": "NewJohnPassword123!",
		}

		resp = suite.postRequest(t, "/auth/login", loginReq, nil)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 6. 登出
		resp = suite.postRequest(t, "/auth/logout", nil, authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)
	})
}

// TestLawyerClientManagementWorkflow 测试律师客户管理工作流程
func TestLawyerClientManagementWorkflow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.TearDown()

	authHeaders := suite.getAuthHeaders("lawyer")

	t.Run("Client Management Lifecycle", func(t *testing.T) {
		// 1. 创建新客户
		newClient := map[string]interface{}{
			"name":    "Acme Corporation",
			"email":   "contact@acme.com",
			"phone":   "5551234567",
			"address": "123 Business Ave, Suite 100",
			"company": "Acme Corporation",
			"notes":   "Important corporate client",
		}

		resp := suite.postRequest(t, "/clients", newClient, authHeaders)
		assert.Equal(t, http.StatusCreated, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusCreated)

		// 解析创建响应获取客户ID
		var createResp map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &createResp)
		require.NoError(t, err)

		clientData := createResp["data"].(map[string]interface{})
		clientID := uint(clientData["id"].(float64))

		// 2. 获取客户详情
		resp = suite.getRequest(t, fmt.Sprintf("/clients/%d", clientID), authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 3. 更新客户信息
		updateClient := map[string]interface{}{
			"name":    "Acme Corporation Ltd",
			"phone":   "5559876543",
			"address": "456 Corporate Blvd, Floor 20",
			"notes":   "VIP corporate client - high priority",
		}

		resp = suite.putRequest(t, fmt.Sprintf("/clients/%d", clientID), updateClient, authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 4. 创建案件
		newCase := map[string]interface{}{
			"title":       "Acme Corporation Contract Dispute",
			"description": "Breach of contract case involving significant damages",
			"client_id":   clientID,
			"case_type":   "commercial",
			"priority":    "high",
		}

		resp = suite.postRequest(t, "/cases", newCase, authHeaders)
		assert.Equal(t, http.StatusCreated, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusCreated)

		// 解析案件响应获取案件ID
		var caseResp map[string]interface{}
		err = json.Unmarshal(resp.Body.Bytes(), &caseResp)
		require.NoError(t, err)

		caseData := caseResp["data"].(map[string]interface{})
		caseID := uint(caseData["id"].(float64))

		// 5. 更新案件状态
		updateCase := map[string]interface{}{
			"title":    "Acme Corporation Contract Dispute - Updated",
			"priority": "high",
			"status":   "in_progress",
		}

		resp = suite.putRequest(t, fmt.Sprintf("/cases/%d", caseID), updateCase, authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 6. 获取案件列表
		resp = suite.getRequest(t, "/cases?page=1&page_size=10&client_id=1", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 7. 获取客户列表
		resp = suite.getRequest(t, "/clients?page=1&page_size=10&search=Acme", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 8. 删除案件
		resp = suite.deleteRequest(t, fmt.Sprintf("/cases/%d", caseID), authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 9. 删除客户
		resp = suite.deleteRequest(t, fmt.Sprintf("/clients/%d", clientID), authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)
	})
}

// TestAdminUserManagementWorkflow 测试管理员用户管理工作流程
func TestAdminUserManagementWorkflow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.TearDown()

	authHeaders := suite.getAuthHeaders("admin")

	t.Run("User Management Operations", func(t *testing.T) {
		// 1. 获取用户列表
		resp := suite.getRequest(t, "/users?page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 2. 按角色过滤用户
		resp = suite.getRequest(t, "/users?role=lawyer&page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 3. 搜索用户
		resp = suite.getRequest(t, "/users?search=Alice&page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 4. 获取特定用户详情
		resp = suite.getRequest(t, "/users/1", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 5. 更新用户角色
		updateUser := map[string]interface{}{
			"role": "admin",
		}

		resp = suite.putRequest(t, "/users/1", updateUser, authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 6. 验证删除后的用户列表
		resp = suite.getRequest(t, "/users?page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)
	})
}

// TestCaseManagementWorkflow 测试案件管理工作流程
func TestCaseManagementWorkflow(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.TearDown()

	authHeaders := suite.getAuthHeaders("lawyer")

	t.Run("Case Management Lifecycle", func(t *testing.T) {
		// 1. 创建多个客户
		clients := []map[string]interface{}{
			{
				"name":    "TechStart Inc",
				"email":   "legal@techstart.com",
				"phone":   "5551112222",
				"company": "TechStart Inc",
			},
			{
				"name":    "Global Corp",
				"email":   "legal@globalcorp.com",
				"phone":   "5553334444",
				"company": "Global Corp",
			},
		}

		var clientIDs []uint
		for _, client := range clients {
			resp := suite.postRequest(t, "/clients", client, authHeaders)
			assert.Equal(t, http.StatusCreated, resp.Code)

			var clientResp map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &clientResp)
			require.NoError(t, err)

			clientData := clientResp["data"].(map[string]interface{})
			clientIDs = append(clientIDs, uint(clientData["id"].(float64)))
		}

		// 2. 为每个客户创建多个案件
		cases := []map[string]interface{}{
			{
				"title":       "TechStart Patent Infringement",
				"description": "Patent infringement case against competitor",
				"client_id":   clientIDs[0],
				"case_type":   "intellectual_property",
				"priority":    "high",
			},
			{
				"title":       "TechStart Employment Dispute",
				"description": "Wrongful termination lawsuit",
				"client_id":   clientIDs[0],
				"case_type":   "employment",
				"priority":    "medium",
			},
			{
				"title":       "Global Corp Merger Review",
				"description": "Legal review of proposed merger",
				"client_id":   clientIDs[1],
				"case_type":   "corporate",
				"priority":    "high",
			},
		}

		var caseIDs []uint
		for _, caseData := range cases {
			resp := suite.postRequest(t, "/cases", caseData, authHeaders)
			assert.Equal(t, http.StatusCreated, resp.Code)

			var caseResp map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &caseResp)
			require.NoError(t, err)

			caseDataResp := caseResp["data"].(map[string]interface{})
			caseIDs = append(caseIDs, uint(caseDataResp["id"].(float64)))
		}

		// 3. 获取案件列表并验证
		resp := suite.getRequest(t, "/cases?page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 4. 按客户过滤案件
		resp = suite.getRequest(t, fmt.Sprintf("/cases?client_id=%d&page=1&page_size=10", clientIDs[0]), authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 5. 按案件类型过滤
		resp = suite.getRequest(t, "/cases?case_type=intellectual_property&page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 6. 按优先级过滤
		resp = suite.getRequest(t, "/cases?priority=high&page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 7. 搜索案件
		resp = suite.getRequest(t, "/cases?search=TechStart&page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 8. 更新案件状态
		caseUpdates := []struct {
			caseID uint
			status string
		}{
			{caseIDs[0], "in_progress"},
			{caseIDs[1], "pending_review"},
			{caseIDs[2], "completed"},
		}

		for _, update := range caseUpdates {
			updateData := map[string]interface{}{
				"status": update.status,
			}

			resp = suite.putRequest(t, fmt.Sprintf("/cases/%d", update.caseID), updateData, authHeaders)
			assert.Equal(t, http.StatusOK, resp.Code)
			test.AssertSuccessResponse(t, resp, http.StatusOK)
		}

		// 9. 验证状态更新
		resp = suite.getRequest(t, "/cases?status=in_progress&page=1&page_size=10", authHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)
		test.AssertSuccessResponse(t, resp, http.StatusOK)

		// 10. 删除剩余案件
		for _, caseID := range caseIDs {
			resp = suite.deleteRequest(t, fmt.Sprintf("/cases/%d", caseID), authHeaders)
			assert.Equal(t, http.StatusOK, resp.Code)
		}

		// 11. 删除客户
		for _, clientID := range clientIDs {
			resp = suite.deleteRequest(t, fmt.Sprintf("/clients/%d", clientID), authHeaders)
			assert.Equal(t, http.StatusOK, resp.Code)
		}
	})
}

// TestAuthorizationAndSecurity 测试授权和安全性
func TestAuthorizationAndSecurity(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.TearDown()

	t.Run("Role-Based Access Control", func(t *testing.T) {
		// 测试不同角色的访问权限

		// 1. 普通用户尝试访问用户管理功能
		userHeaders := suite.getAuthHeaders("user")

		// 尝试获取用户列表 - 应该失败
		resp := suite.getRequest(t, "/users", userHeaders)
		assert.Equal(t, http.StatusForbidden, resp.Code)

		// 尝试创建用户 - 应该失败
		newUser := map[string]interface{}{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "Password123!",
			"role":     "user",
		}

		resp = suite.postRequest(t, "/users", newUser, userHeaders)
		assert.Equal(t, http.StatusForbidden, resp.Code)

		// 2. 律师尝试访问用户管理功能
		lawyerHeaders := suite.getAuthHeaders("lawyer")

		// 尝试获取用户列表 - 应该成功
		resp = suite.getRequest(t, "/users", lawyerHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)

		// 尝试创建用户 - 应该失败
		resp = suite.postRequest(t, "/users", newUser, lawyerHeaders)
		assert.Equal(t, http.StatusForbidden, resp.Code)

		// 3. 管理员访问用户管理功能
		adminHeaders := suite.getAuthHeaders("admin")

		// 尝试获取用户列表 - 应该成功
		resp = suite.getRequest(t, "/users", adminHeaders)
		assert.Equal(t, http.StatusOK, resp.Code)

		// 尝试创建用户 - 应该成功
		resp = suite.postRequest(t, "/users", newUser, adminHeaders)
		assert.Equal(t, http.StatusCreated, resp.Code)
	})

	t.Run("Unauthenticated Access", func(t *testing.T) {
		// 测试未认证访问

		// 尝试访问需要认证的端点
		protectedEndpoints := []string{
			"/users/profile",
			"/users",
			"/clients",
			"/cases",
		}

		for _, endpoint := range protectedEndpoints {
			resp := suite.getRequest(t, endpoint, nil)
			assert.Equal(t, http.StatusUnauthorized, resp.Code)
		}
	})

	t.Run("Invalid Authentication", func(t *testing.T) {
		// 测试无效的认证token

		invalidHeaders := map[string]string{"Authorization": "Bearer invalid_token"}

		// 尝试使用无效token访问
		resp := suite.getRequest(t, "/users/profile", invalidHeaders)
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

// TestAPIErrorHandling 测试API错误处理
func TestAPIErrorHandling(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.TearDown()

	authHeaders := suite.getAuthHeaders("lawyer")

	t.Run("Invalid Input Handling", func(t *testing.T) {
		// 测试无效输入处理

		// 1. 无效的JSON
		invalidJSON := `{"name": "Test", "email": "invalid"}`

		req := httptest.NewRequest("POST", suite.baseURL+"/clients", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeaders["Authorization"])

		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		// 2. 缺少必填字段
		incompleteClient := map[string]interface{}{
			"name": "Test Client",
			// 缺少email字段
		}

		resp := suite.postRequest(t, "/clients", incompleteClient, authHeaders)
		assert.Equal(t, http.StatusBadRequest, resp.Code)

		// 3. 无效的邮箱格式
		invalidEmailClient := map[string]interface{}{
			"name":  "Test Client",
			"email": "invalid-email-format",
		}

		resp = suite.postRequest(t, "/clients", invalidEmailClient, authHeaders)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("Resource Not Found", func(t *testing.T) {
		// 测试资源不存在的情况

		// 尝试访问不存在的用户
		resp := suite.getRequest(t, "/users/99999", authHeaders)
		assert.Equal(t, http.StatusNotFound, resp.Code)

		// 尝试访问不存在的客户
		resp = suite.getRequest(t, "/clients/99999", authHeaders)
		assert.Equal(t, http.StatusNotFound, resp.Code)

		// 尝试访问不存在的案件
		resp = suite.getRequest(t, "/cases/99999", authHeaders)
		assert.Equal(t, http.StatusNotFound, resp.Code)
	})

	t.Run("Duplicate Resource Creation", func(t *testing.T) {
		// 测试重复资源创建

		// 创建第一个客户
		client := map[string]interface{}{
			"name":  "Duplicate Test Client",
			"email": "duplicate@example.com",
		}

		resp := suite.postRequest(t, "/clients", client, authHeaders)
		assert.Equal(t, http.StatusCreated, resp.Code)

		// 尝试创建相同邮箱的客户
		resp = suite.postRequest(t, "/clients", client, authHeaders)
		assert.Equal(t, http.StatusConflict, resp.Code)
	})
}

// TestAPIPerformance 测试API性能
func TestAPIPerformance(t *testing.T) {
	suite := NewE2ETestSuite(t)
	defer suite.TearDown()

	authHeaders := suite.getAuthHeaders("admin")

	t.Run("Response Time Testing", func(t *testing.T) {
		// 测试API响应时间

		// 1. 测试健康检查响应时间
		start := time.Now()
		resp := suite.getRequest(t, "/health", nil)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Less(t, duration, 100*time.Millisecond, "Health check should be fast")

		// 2. 测试用户列表响应时间
		start = time.Now()
		resp = suite.getRequest(t, "/users", authHeaders)
		duration = time.Since(start)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Less(t, duration, 500*time.Millisecond, "User list should respond within 500ms")
	})

	t.Run("Concurrent Request Handling", func(t *testing.T) {
		// 测试并发请求处理

		// 创建多个客户
		clients := make([]map[string]interface{}, 10)
		for i := 0; i < 10; i++ {
			clients[i] = map[string]interface{}{
				"name":    fmt.Sprintf("Concurrent Client %d", i+1),
				"email":   fmt.Sprintf("concurrent%d@example.com", i+1),
				"phone":   fmt.Sprintf("555%07d", i+1),
				"company": fmt.Sprintf("Company %d", i+1),
			}
		}

		// 并发发送创建请求
		results := make(chan *httptest.ResponseRecorder, len(clients))
		for _, client := range clients {
			go func(c map[string]interface{}) {
				resp := suite.postRequest(t, "/clients", c, authHeaders)
				results <- resp
			}(client)
		}

		// 收集结果
		successCount := 0
		for i := 0; i < len(clients); i++ {
			resp := <-results
			if resp.Code == http.StatusCreated {
				successCount++
			}
		}

		assert.Equal(t, len(clients), successCount, "All concurrent requests should succeed")
	})
}
