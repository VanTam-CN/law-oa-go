package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockClientRepository Mock客户端仓储
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Create(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) FindByID(ctx context.Context, id uint) (*models.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientRepository) FindByEmail(ctx context.Context, email string) (*models.Client, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepository) List(ctx context.Context, params *repositories.ClientListParams) ([]*models.Client, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*models.Client), args.Get(1).(int64), args.Error(2)
}

func (m *MockClientRepository) GetStats(ctx context.Context) (*repositories.ClientStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ClientStats), args.Error(1)
}

// TestNewClientService 测试客户端服务创建
func TestNewClientService(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.clientRepo)
}

// TestClientService_CreateClient_Success 测试创建客户端成功
func TestClientService_CreateClient_Success(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	req := &CreateClientRequest{
		Name:    "测试客户",
		Email:   "test@example.com",
		Phone:   "13800138000",
		Address: "测试地址",
		Company: "测试公司",
		Notes:   "测试备注",
	}

	createdClient := &models.Client{
		ID:        1,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Address:   req.Address,
		Company:   req.Company,
		Notes:     req.Notes,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置Mock期望
	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Client")).Return(nil)

	// 执行测试
	result, err := service.CreateClient(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Email, result.Email)
	assert.Equal(t, req.Phone, result.Phone)
	assert.Equal(t, req.Address, result.Address)
	assert.Equal(t, req.Company, result.Company)
	assert.Equal(t, req.Notes, result.Notes)
	assert.Equal(t, "active", result.Status)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_CreateClient_EmailAlreadyExists 测试邮箱已存在
func TestClientService_CreateClient_EmailAlreadyExists(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	req := &CreateClientRequest{
		Name:  "测试客户",
		Email: "existing@example.com",
	}

	existingClient := &models.Client{
		ID:    1,
		Name:  "已存在客户",
		Email: "existing@example.com",
	}

	// 设置Mock期望
	mockRepo.On("FindByEmail", ctx, req.Email).Return(existingClient, nil)

	// 执行测试
	result, err := service.CreateClient(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Email already exists")

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_CreateClient_InvalidEmail 测试无效邮箱
func TestClientService_CreateClient_InvalidEmail(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	req := &CreateClientRequest{
		Name:  "测试客户",
		Email: "invalid-email",
	}

	// 执行测试
	result, err := service.CreateClient(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Invalid email format")

	// 验证Mock调用 - 不应该有任何数据库调用
	mockRepo.AssertNotCalled(t, "FindByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

// TestClientService_CreateClient_InvalidPhone 测试无效电话
func TestClientService_CreateClient_InvalidPhone(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	req := &CreateClientRequest{
		Name:  "测试客户",
		Phone: "invalid-phone-with-chinese-中文",
	}

	// 执行测试
	result, err := service.CreateClient(ctx, req)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Invalid phone format")
}

// TestClientService_GetClientByID_Success 测试根据ID获取客户端成功
func TestClientService_GetClientByID_Success(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	clientID := uint(1)

	expectedClient := &models.Client{
		ID:        clientID,
		Name:      "测试客户",
		Email:     "test@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, clientID).Return(expectedClient, nil)

	// 执行测试
	result, err := service.GetClientByID(ctx, clientID)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, clientID, result.ID)
	assert.Equal(t, expectedClient.Name, result.Name)
	assert.Equal(t, expectedClient.Email, result.Email)
	assert.Equal(t, expectedClient.Status, result.Status)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_GetClientByID_NotFound 测试客户端不存在
func TestClientService_GetClientByID_NotFound(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	clientID := uint(999)

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, clientID).Return(nil, nil)

	// 执行测试
	result, err := service.GetClientByID(ctx, clientID)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Client not found")

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_UpdateClient_Success 测试更新客户端成功
func TestClientService_UpdateClient_Success(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	clientID := uint(1)

	existingClient := &models.Client{
		ID:        clientID,
		Name:      "原始名称",
		Email:     "original@example.com",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updateReq := &UpdateClientRequest{
		Name:    stringPtr("更新名称"),
		Email:   stringPtr("updated@example.com"),
		Phone:   stringPtr("13900139000"),
		Address: stringPtr("更新地址"),
		Status:  stringPtr("inactive"),
	}

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("FindByEmail", ctx, "updated@example.com").Return(nil, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Client")).Return(nil)
	mockRepo.On("FindByID", ctx, clientID).Return(existingClient, nil) // 更新后再次查询

	// 执行测试
	result, err := service.UpdateClient(ctx, clientID, updateReq)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_UpdateClient_EmailConflict 测试更新时邮箱冲突
func TestClientService_UpdateClient_EmailConflict(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	clientID := uint(1)

	existingClient := &models.Client{
		ID:    clientID,
		Name:  "原始客户",
		Email: "original@example.com",
	}

	anotherClient := &models.Client{
		ID:    2,
		Name:  "另一个客户",
		Email: "another@example.com",
	}

	updateReq := &UpdateClientRequest{
		Name:  stringPtr("更新名称"),
		Email: stringPtr("another@example.com"), // 使用已存在的邮箱
	}

	// 设置Mock期望
	mockRepo.On("FindByID", ctx, clientID).Return(existingClient, nil)
	mockRepo.On("FindByEmail", ctx, "another@example.com").Return(anotherClient, nil)

	// 执行测试
	result, err := service.UpdateClient(ctx, clientID, updateReq)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Email already exists")

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_DeleteClient_Success 测试删除客户端成功
func TestClientService_DeleteClient_Success(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	clientID := uint(1)

	// 设置Mock期望
	mockRepo.On("Delete", ctx, clientID).Return(nil)

	// 执行测试
	err := service.DeleteClient(ctx, clientID)

	// 验证结果
	assert.NoError(t, err)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_DeleteClient_NotFound 测试删除不存在的客户端
func TestClientService_DeleteClient_NotFound(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	clientID := uint(999)

	// 设置Mock期望
	mockRepo.On("Delete", ctx, clientID).Return(gorm.ErrRecordNotFound)

	// 执行测试
	err := service.DeleteClient(ctx, clientID)

	// 验证结果
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Client not found")

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_ListClients_Success 测试获取客户端列表成功
func TestClientService_ListClients_Success(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	req := &ClientListRequest{
		Page:     1,
		PageSize: 10,
		Status:   "active",
		Search:   "测试",
	}

	expectedClients := []*models.Client{
		{
			ID:     1,
			Name:   "测试客户1",
			Status: "active",
		},
		{
			ID:     2,
			Name:   "测试客户2",
			Status: "active",
		},
	}

	// 设置Mock期望
	mockRepo.On("List", ctx, &repositories.ClientListParams{
		Page:     1,
		PageSize: 10,
		Status:   "active",
		Search:   "测试",
	}).Return(expectedClients, int64(2), nil)

	// 执行测试
	result, total, err := service.ListClients(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, "测试客户1", result[0].Name)
	assert.Equal(t, "测试客户2", result[1].Name)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_GetClientStats_Success 测试获取客户端统计成功
func TestClientService_GetClientStats_Success(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()

	expectedStats := &repositories.ClientStats{
		TotalClients:        100,
		ActiveClients:       80,
		InactiveClients:     20,
		NewClientsThisMonth: 5,
	}

	// 设置Mock期望
	mockRepo.On("GetStats", ctx).Return(expectedStats, nil)

	// 执行测试
	result, err := service.GetClientStats(ctx)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(100), result.TotalClients)
	assert.Equal(t, int64(80), result.ActiveClients)
	assert.Equal(t, int64(20), result.InactiveClients)
	assert.Equal(t, int64(5), result.NewClientsThisMonth)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_CreateClient_WithoutEmail 测试不提供邮箱创建客户端
func TestClientService_CreateClient_WithoutEmail(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	req := &CreateClientRequest{
		Name:    "测试客户",
		Phone:   "13800138000",
		Address: "测试地址",
	}

	// 设置Mock期望
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Client")).Return(nil)

	// 执行测试
	result, err := service.CreateClient(ctx, req)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Phone, result.Phone)
	assert.Equal(t, "", result.Email) // 邮箱应该为空

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestClientService_CreateClient_ValidationError 测试创建客户端验证错误
func TestClientService_CreateClient_ValidationError(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()

	testCases := []struct {
		name    string
		req     *CreateClientRequest
		errMsg  string
	}{
		{
			name: "空名称",
			req: &CreateClientRequest{
				Name: "",
			},
			errMsg: "name is required",
		},
		{
			name: "名称过长",
			req: &CreateClientRequest{
				Name: string(make([]byte, 101)), // 101个字符
			},
			errMsg: "name is too long",
		},
		{
			name: "电话过长",
			req: &CreateClientRequest{
				Name:  "正常名称",
				Phone: string(make([]byte, 21)), // 21个字符
			},
			errMsg: "phone is too long",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行测试
			result, err := service.CreateClient(ctx, tc.req)

			// 验证结果
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

// TestClientService_toClientResponse 测试客户端响应转换
func TestClientService_toClientResponse(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	client := &models.Client{
		ID:        1,
		Name:      "测试客户",
		Email:     "test@example.com",
		Phone:     "13800138000",
		Address:   "测试地址",
		Company:   "测试公司",
		Notes:     "测试备注",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 执行测试
	result := service.toClientResponse(client)

	// 验证结果
	assert.NotNil(t, result)
	assert.Equal(t, client.ID, result.ID)
	assert.Equal(t, client.Name, result.Name)
	assert.Equal(t, client.Email, result.Email)
	assert.Equal(t, client.Phone, result.Phone)
	assert.Equal(t, client.Address, result.Address)
	assert.Equal(t, client.Company, result.Company)
	assert.Equal(t, client.Notes, result.Notes)
	assert.Equal(t, client.Status, result.Status)
	assert.Equal(t, client.CreatedAt, result.CreatedAt)
	assert.Equal(t, client.UpdatedAt, result.UpdatedAt)
}

// BenchmarkClientService_CreateClient 基准测试创建客户端性能
func BenchmarkClientService_CreateClient(b *testing.B) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()
	req := &CreateClientRequest{
		Name:    "基准测试客户",
		Email:   "benchmark@example.com",
		Phone:   "13800138000",
		Address: "基准测试地址",
	}

	// 设置Mock期望
	mockRepo.On("FindByEmail", ctx, req.Email).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Client")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateClient(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

// TestClientService_Integration_CompleteWorkflow 测试客户端服务完整工作流
func TestClientService_Integration_CompleteWorkflow(t *testing.T) {
	mockRepo := &MockClientRepository{}
	service := NewClientService(mockRepo)

	ctx := context.Background()

	// 1. 创建客户端
	createReq := &CreateClientRequest{
		Name:    "集成测试客户",
		Email:   "integration@example.com",
		Phone:   "13800138000",
		Address: "集成测试地址",
	}

	createdClient := &models.Client{
		ID:        1,
		Name:      createReq.Name,
		Email:     createReq.Email,
		Phone:     createReq.Phone,
		Address:   createReq.Address,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 2. 获取客户端
	getClient := &models.Client{
		ID:        1,
		Name:      createReq.Name,
		Email:     createReq.Email,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 3. 更新客户端
	updatedClient := &models.Client{
		ID:        1,
		Name:      "更新后的集成测试客户",
		Email:     createReq.Email,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置Mock期望 - 创建
	mockRepo.On("FindByEmail", ctx, createReq.Email).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Client")).Return(nil)

	// 设置Mock期望 - 获取
	mockRepo.On("FindByID", ctx, uint(1)).Return(getClient, nil)

	// 设置Mock期望 - 更新
	mockRepo.On("FindByID", ctx, uint(1)).Return(getClient, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.Client")).Return(nil)
	mockRepo.On("FindByID", ctx, uint(1)).Return(updatedClient, nil)

	// 设置Mock期望 - 删除
	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	// 执行工作流测试

	// 1. 创建客户端
	created, err := service.CreateClient(ctx, createReq)
	require.NoError(t, err)
	assert.NotNil(t, created)

	// 2. 获取客户端
	fetched, err := service.GetClientByID(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, fetched)

	// 3. 更新客户端
	updateReq := &UpdateClientRequest{
		Name: stringPtr("更新后的集成测试客户"),
	}
	updated, err := service.UpdateClient(ctx, 1, updateReq)
	require.NoError(t, err)
	assert.NotNil(t, updated)

	// 4. 删除客户端
	err = service.DeleteClient(ctx, 1)
	assert.NoError(t, err)

	// 验证所有Mock调用
	mockRepo.AssertExpectations(t)
}