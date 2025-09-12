package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa/test"
)

// ClientServiceTest 客户服务测试
func TestClientService(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	t.Run("CreateClient", func(t *testing.T) {
		// 测试创建客户
		client := &models.Client{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "13800138000",
			Address:   "123 Main St",
			City:      "Shanghai",
			Province:  "Shanghai",
			Country:   "China",
			Status:    "active",
		}
		
		createdClient, err := clientService.CreateClient(client)
		require.NoError(t, err)
		assert.NotZero(t, createdClient.ID)
		assert.Equal(t, client.FirstName, createdClient.FirstName)
		assert.Equal(t, client.LastName, createdClient.LastName)
		assert.Equal(t, client.Email, createdClient.Email)
	})
	
	t.Run("GetClient", func(t *testing.T) {
		// 测试获取客户
		client := config.CreateTestClient(t, suite.DB)
		
		retrievedClient, err := clientService.GetClient(client.ID)
		require.NoError(t, err)
		assert.Equal(t, client.ID, retrievedClient.ID)
		assert.Equal(t, client.FirstName, retrievedClient.FirstName)
		assert.Equal(t, client.LastName, retrievedClient.LastName)
	})
	
	t.Run("GetClientNotFound", func(t *testing.T) {
		// 测试获取不存在的客户
		_, err := clientService.GetClient(99999)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
	
	t.Run("UpdateClient", func(t *testing.T) {
		// 测试更新客户
		client := config.CreateTestClient(t, suite.DB)
		
		updatedClient := *client
		updatedClient.FirstName = "Updated"
		updatedClient.Phone = "13900139000"
		
		result, err := clientService.UpdateClient(client.ID, &updatedClient)
		require.NoError(t, err)
		assert.Equal(t, "Updated", result.FirstName)
		assert.Equal(t, "13900139000", result.Phone)
	})
	
	t.Run("DeleteClient", func(t *testing.T) {
		// 测试删除客户
		client := config.CreateTestClient(t, suite.DB)
		
		err := clientService.DeleteClient(client.ID)
		require.NoError(t, err)
		
		// 验证客户已被删除
		_, err = clientService.GetClient(client.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
	
	t.Run("ListClients", func(t *testing.T) {
		// 测试客户列表
		// 创建多个客户
		for i := 0; i < 5; i++ {
			config.CreateTestClient(t, suite.DB)
		}
		
		clients, total, err := clientService.ListClients(1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(clients), 5)
		assert.GreaterOrEqual(t, total, 5)
	})
	
	t.Run("SearchClients", func(t *testing.T) {
		// 测试搜索客户
		client := config.CreateTestClient(t, suite.DB)
		
		searchTerm := client.FirstName
		clients, err := clientService.SearchClients(searchTerm)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(clients), 1)
	})
	
	t.Run("GetClientCases", func(t *testing.T) {
		// 测试获取客户案件
		client := config.CreateTestClient(t, suite.DB)
		lawyer := config.CreateTestLawyer(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		cases, err := clientService.GetClientCases(client.ID)
		require.NoError(t, err)
		assert.Len(t, cases, 1)
		assert.Equal(t, case.ID, cases[0].ID)
	})
}

// ClientHandlerTest 客户处理器测试
func TestClientHandler(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	handler := NewClientHandler(suite.DB)
	
	// 设置路由
	router := gin.New()
	router.POST("/clients", handler.CreateClient)
	router.GET("/clients/:id", handler.GetClient)
	router.PUT("/clients/:id", handler.UpdateClient)
	router.DELETE("/clients/:id", handler.DeleteClient)
	router.GET("/clients", handler.ListClients)
	router.GET("/clients/search", handler.SearchClients)
	router.GET("/clients/:id/cases", handler.GetClientCases)
	
	suite.Router = router
	
	t.Run("CreateClientAPI", func(t *testing.T) {
		// 测试创建客户API
		client := map[string]interface{}{
			"first_name": "Jane",
			"last_name":  "Smith",
			"email":      "jane.smith@example.com",
			"phone":      "13700137000",
			"address":    "456 Oak Ave",
			"city":       "Beijing",
			"province":   "Beijing",
			"country":    "China",
			"status":     "active",
		}
		
		w := suite.PerformAuthRequest("POST", "/clients", client)
		data := suite.AssertSuccess(t, w)
		
		assert.NotZero(t, data["id"])
		assert.Equal(t, "Jane", data["first_name"])
		assert.Equal(t, "Smith", data["last_name"])
	})
	
	t.Run("GetClientAPI", func(t *testing.T) {
		// 测试获取客户API
		client := config.CreateTestClient(t, suite.DB)
		
		w := suite.PerformAuthRequest("GET", "/clients/"+string(client.ID), nil)
		data := suite.AssertSuccess(t, w)
		
		assert.Equal(t, client.ID, uint(data["id"].(float64)))
		assert.Equal(t, client.FirstName, data["first_name"])
		assert.Equal(t, client.LastName, data["last_name"])
	})
	
	t.Run("UpdateClientAPI", func(t *testing.T) {
		// 测试更新客户API
		client := config.CreateTestClient(t, suite.DB)
		
		updateData := map[string]interface{}{
			"first_name": "Updated",
			"phone":      "13900139000",
		}
		
		w := suite.PerformAuthRequest("PUT", "/clients/"+string(client.ID), updateData)
		data := suite.AssertSuccess(t, w)
		
		assert.Equal(t, "Updated", data["first_name"])
		assert.Equal(t, "13900139000", data["phone"])
	})
	
	t.Run("DeleteClientAPI", func(t *testing.T) {
		// 测试删除客户API
		client := config.CreateTestClient(t, suite.DB)
		
		w := suite.PerformAuthRequest("DELETE", "/clients/"+string(client.ID), nil)
		suite.AssertSuccess(t, w)
		
		// 验证客户已被删除
		w = suite.PerformAuthRequest("GET", "/clients/"+string(client.ID), nil)
		suite.AssertError(t, w, 404)
	})
	
	t.Run("ListClientsAPI", func(t *testing.T) {
		// 测试客户列表API
		// 创建多个客户
		for i := 0; i < 3; i++ {
			config.CreateTestClient(t, suite.DB)
		}
		
		w := suite.PerformAuthRequest("GET", "/clients?page=1&size=10", nil)
		data := suite.AssertSuccess(t, w)
		
		clients := data["clients"].([]interface{})
		total := data["total"].(float64)
		
		assert.GreaterOrEqual(t, len(clients), 3)
		assert.GreaterOrEqual(t, total, 3)
	})
	
	t.Run("SearchClientsAPI", func(t *testing.T) {
		// 测试搜索客户API
		client := config.CreateTestClient(t, suite.DB)
		
		w := suite.PerformAuthRequest("GET", "/clients/search?q="+client.FirstName, nil)
		data := suite.AssertSuccess(t, w)
		
		clients := data["clients"].([]interface{})
		assert.GreaterOrEqual(t, len(clients), 1)
	})
	
	t.Run("GetClientCasesAPI", func(t *testing.T) {
		// 测试获取客户案件API
		client := config.CreateTestClient(t, suite.DB)
		lawyer := config.CreateTestLawyer(t, suite.DB)
		case := config.CreateTestCase(t, suite.DB, lawyer.ID, client.ID)
		
		w := suite.PerformAuthRequest("GET", "/clients/"+string(client.ID)+"/cases", nil)
		data := suite.AssertSuccess(t, w)
		
		cases := data["cases"].([]interface{})
		assert.Len(t, cases, 1)
		assert.Equal(t, case.ID, uint(cases[0].(map[string]interface{})["id"].(float64)))
	})
	
	t.Run("CreateClientValidation", func(t *testing.T) {
		// 测试创建客户验证
		invalidClient := map[string]interface{}{
			"first_name": "", // 空名字
			"email":      "invalid-email", // 无效邮箱
		}
		
		w := suite.PerformAuthRequest("POST", "/clients", invalidClient)
		suite.AssertError(t, w, 400)
	})
	
	t.Run("UnauthorizedAccess", func(t *testing.T) {
		// 测试未授权访问
		w := suite.PerformRequest("GET", "/clients", nil, nil)
		suite.AssertError(t, w, 401)
	})
}

// ClientValidatorTest 客户验证器测试
func TestClientValidator(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	validator := NewClientValidator()
	
	t.Run("ValidClient", func(t *testing.T) {
		// 测试有效客户验证
		client := &models.Client{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "13800138000",
		}
		
		err := validator.ValidateClient(client)
		assert.NoError(t, err)
	})
	
	t.Run("InvalidEmail", func(t *testing.T) {
		// 测试无效邮箱
		client := &models.Client{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "invalid-email",
			Phone:     "13800138000",
		}
		
		err := validator.ValidateClient(client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email")
	})
	
	t.Run("MissingRequiredFields", func(t *testing.T) {
		// 测试缺失必填字段
		client := &models.Client{
			FirstName: "",
			LastName:  "",
			Email:     "",
		}
		
		err := validator.ValidateClient(client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first_name")
		assert.Contains(t, err.Error(), "last_name")
		assert.Contains(t, err.Error(), "email")
	})
	
	t.Run("InvalidPhone", func(t *testing.T) {
		// 测试无效手机号
		client := &models.Client{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "123", // 太短
		}
		
		err := validator.ValidateClient(client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "phone")
	})
}

// ClientRepositoryTest 客户仓库测试
func TestClientRepository(t *testing.T) {
	suite := test.NewTestSuite(t)
	defer suite.Tearardown()
	
	repo := NewClientRepository(suite.DB)
	
	t.Run("CreateClient", func(t *testing.T) {
		// 测试创建客户
		client := &models.Client{
			FirstName: "Test",
			LastName:  "Client",
			Email:     "test@example.com",
			Phone:     "13800138000",
		}
		
		err := repo.Create(client)
		require.NoError(t, err)
		assert.NotZero(t, client.ID)
	})
	
	t.Run("FindByID", func(t *testing.T) {
		// 测试根据ID查找
		client := config.CreateTestClient(t, suite.DB)
		
		foundClient, err := repo.FindByID(client.ID)
		require.NoError(t, err)
		assert.Equal(t, client.ID, foundClient.ID)
		assert.Equal(t, client.FirstName, foundClient.FirstName)
	})
	
	t.Run("FindByEmail", func(t *testing.T) {
		// 测试根据邮箱查找
		client := config.CreateTestClient(t, suite.DB)
		
		foundClient, err := repo.FindByEmail(client.Email)
		require.NoError(t, err)
		assert.Equal(t, client.ID, foundClient.ID)
		assert.Equal(t, client.Email, foundClient.Email)
	})
	
	t.Run("Update", func(t *testing.T) {
		// 测试更新
		client := config.CreateTestClient(t, suite.DB)
		
		client.FirstName = "Updated"
		client.Phone = "13900139000"
		
		err := repo.Update(client)
		require.NoError(t, err)
		
		// 验证更新
		updatedClient, err := repo.FindByID(client.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updatedClient.FirstName)
		assert.Equal(t, "13900139000", updatedClient.Phone)
	})
	
	t.Run("Delete", func(t *testing.T) {
		// 测试删除
		client := config.CreateTestClient(t, suite.DB)
		
		err := repo.Delete(client.ID)
		require.NoError(t, err)
		
		// 验证删除
		_, err = repo.FindByID(client.ID)
		assert.Error(t, err)
	})
	
	t.Run("List", func(t *testing.T) {
		// 测试列表
		// 创建多个客户
		for i := 0; i < 5; i++ {
			config.CreateTestClient(t, suite.DB)
		}
		
		clients, total, err := repo.List(1, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(clients), 5)
		assert.GreaterOrEqual(t, total, 5)
	})
	
	t.Run("Search", func(t *testing.T) {
		// 测试搜索
		client := config.CreateTestClient(t, suite.DB)
		
		clients, err := repo.Search(client.FirstName)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(clients), 1)
	})
}