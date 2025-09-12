package services_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
	"law-oa-go/test"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestClientService_CreateClient(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	clientService := services.NewClientService(testDB.DB)
	
	t.Run("successful client creation", func(t *testing.T) {
		req := &services.CreateClientRequest{
			Name:        "张三",
			Email:       "zhangsan@example.com",
			Phone:       "13800138000",
			IDCard:      "110101199001011234",
			Address:     "北京市朝阳区xxx街道",
			CompanyName: test.StringPtr("北京某某公司"),
		}
		
		// 模拟检查邮箱唯一性
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// 模拟检查身份证唯一性
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id_card = ?").
			WithArgs(req.IDCard).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// 模拟创建客户
		testDB.Mock.ExpectBegin()
		testDB.Mock.ExpectExec("INSERT INTO `clients`").
			WithArgs(
				sqlmock.AnyArg(), // ID
				req.Name,
				req.Email,
				req.Phone,
				req.IDCard,
				req.Address,
				req.CompanyName,
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		testDB.Mock.ExpectCommit()
		
		client, err := clientService.CreateClient(context.Background(), req)
		
		require.NoError(t, err)
		assert.Equal(t, req.Name, client.Name)
		assert.Equal(t, req.Email, client.Email)
		assert.Equal(t, req.Phone, client.Phone)
		assert.Equal(t, req.IDCard, client.IDCard)
		assert.Equal(t, req.Address, client.Address)
		assert.Equal(t, *req.CompanyName, client.CompanyName)
	})
	
	t.Run("duplicate email", func(t *testing.T) {
		req := &services.CreateClientRequest{
			Name:  "李四",
			Email: "zhangsan@example.com", // 重复邮箱
			Phone: "13900139000",
			IDCard: "110101199001022345",
		}
		
		// 模拟邮箱已存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		
		_, err := clientService.CreateClient(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})
	
	t.Run("duplicate id card", func(t *testing.T) {
		req := &services.CreateClientRequest{
			Name:  "王五",
			Email: "wangwu@example.com",
			Phone: "13700137000",
			IDCard: "110101199001011234", // 重复身份证
		}
		
		// 模拟邮箱唯一
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// 模拟身份证已存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id_card = ?").
			WithArgs(req.IDCard).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		
		_, err := clientService.CreateClient(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id card already exists")
	})
}

func TestClientService_GetClient(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	clientService := services.NewClientService(testDB.DB)
	
	t.Run("get existing client", func(t *testing.T) {
		clientID := uint(1)
		companyName := "北京某某公司"
		
		// 模拟客户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "id_card", "address", "company_name", "created_at", "updated_at"}).
				AddRow(clientID, "张三", "zhangsan@example.com", "13800138000", "110101199001011234", "北京市朝阳区", &companyName, test.TestTime(), test.TestTime()))
		
		client, err := clientService.GetClient(context.Background(), clientID)
		
		require.NoError(t, err)
		assert.Equal(t, clientID, client.ID)
		assert.Equal(t, "张三", client.Name)
		assert.Equal(t, "zhangsan@example.com", client.Email)
		assert.Equal(t, "北京某某公司", client.CompanyName)
	})
	
	t.Run("client not found", func(t *testing.T) {
		clientID := uint(999)
		
		// 模拟客户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := clientService.GetClient(context.Background(), clientID)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})
}

func TestClientService_UpdateClient(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	clientService := services.NewClientService(testDB.DB)
	
	t.Run("update client successfully", func(t *testing.T) {
		clientID := uint(1)
		req := &services.UpdateClientRequest{
			Name:        test.StringPtr("更新后的姓名"),
			Email:       test.StringPtr("updated@example.com"),
			Phone:       test.StringPtr("13900139000"),
			Address:     test.StringPtr("上海市浦东新区"),
			CompanyName: test.StringPtr("上海某某公司"),
		}
		
		// 模拟查询现有客户
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone"}).
				AddRow(clientID, "原姓名", "original@example.com", "13800138000"))
		
		// 模拟邮箱唯一性检查
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE email = ? AND id != ?").
			WithArgs(*req.Email, clientID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// 模拟更新
		testDB.Mock.ExpectExec("UPDATE `clients`").
			WithArgs(
				*req.Name,
				*req.Email,
				*req.Phone,
				*req.Address,
				*req.CompanyName,
				sqlmock.AnyArg(), // UpdatedAt
				clientID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		client, err := clientService.UpdateClient(context.Background(), clientID, req)
		
		require.NoError(t, err)
		assert.Equal(t, *req.Name, client.Name)
		assert.Equal(t, *req.Email, client.Email)
		assert.Equal(t, *req.Phone, client.Phone)
		assert.Equal(t, *req.Address, client.Address)
		assert.Equal(t, *req.CompanyName, client.CompanyName)
	})
	
	t.Run("client not found", func(t *testing.T) {
		clientID := uint(999)
		req := &services.UpdateClientRequest{
			Name: test.StringPtr("更新的姓名"),
		}
		
		// 模拟客户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := clientService.UpdateClient(context.Background(), clientID, req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})
	
	t.Run("duplicate email", func(t *testing.T) {
		clientID := uint(1)
		req := &services.UpdateClientRequest{
			Email: test.StringPtr("existing@example.com"),
		}
		
		// 模拟查询现有客户
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).
				AddRow(clientID, "original@example.com"))
		
		// 模拟邮箱已被其他客户使用
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE email = ? AND id != ?").
			WithArgs(*req.Email, clientID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		
		_, err := clientService.UpdateClient(context.Background(), clientID, req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})
}

func TestClientService_ListClients(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	clientService := services.NewClientService(testDB.DB)
	
	t.Run("list clients with search", func(t *testing.T) {
		req := &services.ListClientsRequest{
			Page:     1,
			PageSize: 10,
			Search:   "张",
		}
		
		// 模拟查询总数
		testDB.Mock.ExpectQuery("SELECT COUNT(.+) FROM `clients`").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))
		
		// 模拟查询客户列表
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "company_name", "created_at"}).
				AddRow(1, "张三", "zhangsan@example.com", "13800138000", "北京公司", test.TestTime()).
				AddRow(2, "张四", "zhangsi@example.com", "13900139000", "上海公司", test.TestTime()))
		
		resp, err := clientService.ListClients(context.Background(), req)
		
		require.NoError(t, err)
		assert.Equal(t, int64(15), resp.Total)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
		assert.Len(t, resp.Clients, 2)
		assert.Contains(t, resp.Clients[0].Name, "张")
	})
	
	t.Run("list all clients", func(t *testing.T) {
		req := &services.ListClientsRequest{
			Page:     1,
			PageSize: 20,
		}
		
		// 模拟查询总数
		testDB.Mock.ExpectQuery("SELECT COUNT(.+) FROM `clients`").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
		
		// 模拟查询客户列表
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "company_name", "created_at"}).
				AddRow(1, "张三", "zhangsan@example.com", "13800138000", "北京公司", test.TestTime()).
				AddRow(2, "李四", "lisi@example.com", "13900139000", nil, test.TestTime()))
		
		resp, err := clientService.ListClients(context.Background(), req)
		
		require.NoError(t, err)
		assert.Equal(t, int64(100), resp.Total)
		assert.Len(t, resp.Clients, 2)
		assert.Equal(t, "张三", resp.Clients[0].Name)
		assert.Equal(t, "李四", resp.Clients[1].Name)
	})
}

func TestClientService_DeleteClient(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	clientService := services.NewClientService(testDB.DB)
	
	t.Run("delete client successfully", func(t *testing.T) {
		clientID := uint(1)
		
		// 模拟查询客户是否存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(clientID, "张三"))
		
		// 模拟检查是否有关联案件
		testDB.Mock.ExpectQuery("SELECT COUNT(.+) FROM `cases` WHERE client_id = ?").
			WithArgs(clientID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		
		// 模拟删除客户
		testDB.Mock.ExpectExec("DELETE FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		err := clientService.DeleteClient(context.Background(), clientID)
		
		require.NoError(t, err)
	})
	
	t.Run("client has related cases", func(t *testing.T) {
		clientID := uint(1)
		
		// 模拟查询客户存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(clientID, "张三"))
		
		// 模拟有关联案件
		testDB.Mock.ExpectQuery("SELECT COUNT(.+) FROM `cases` WHERE client_id = ?").
			WithArgs(clientID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
		
		err := clientService.DeleteClient(context.Background(), clientID)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client has related cases")
	})
	
	t.Run("client not found", func(t *testing.T) {
		clientID := uint(999)
		
		// 模拟客户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients` WHERE id = ?").
			WithArgs(clientID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		err := clientService.DeleteClient(context.Background(), clientID)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")
	})
}

func TestClientService_SearchClients(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	clientService := services.NewClientService(testDB.DB)
	
	t.Run("search by name", func(t *testing.T) {
		keyword := "张"
		
		// 模拟搜索客户
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "company_name"}).
				AddRow(1, "张三", "zhangsan@example.com", "13800138000", "北京公司").
				AddRow(2, "张四", "zhangsi@example.com", "13900139000", "上海公司"))
		
		clients, err := clientService.SearchClients(context.Background(), keyword)
		
		require.NoError(t, err)
		assert.Len(t, clients, 2)
		assert.Contains(t, clients[0].Name, keyword)
		assert.Contains(t, clients[1].Name, keyword)
	})
	
	t.Run("search by company", func(t *testing.T) {
		keyword := "公司"
		
		// 模拟搜索客户
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "company_name"}).
				AddRow(1, "张三", "zhangsan@example.com", "13800138000", "北京某某公司").
				AddRow(2, "李四", "lisi@example.com", "13900139000", "上海某某公司"))
		
		clients, err := clientService.SearchClients(context.Background(), keyword)
		
		require.NoError(t, err)
		assert.Len(t, clients, 2)
		assert.Contains(t, clients[0].CompanyName, keyword)
		assert.Contains(t, clients[1].CompanyName, keyword)
	})
	
	t.Run("empty search result", func(t *testing.T) {
		keyword := "不存在的关键词"
		
		// 模拟搜索无结果
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `clients`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "phone", "company_name"}))
		
		clients, err := clientService.SearchClients(context.Background(), keyword)
		
		require.NoError(t, err)
		assert.Empty(t, clients)
	})
}