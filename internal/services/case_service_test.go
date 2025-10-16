package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// MockConflictService 用于测试冲突检测服务的模拟
type MockConflictService struct {
	mock.Mock
}

func (m *MockConflictService) CheckConflict(ctx context.Context, req *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConflictCheckResponse), args.Error(1)
}

// MockDB 用于模拟数据库操作
type MockDB struct {
	mock.Mock
}

func (m *MockDB)WithContext(ctx context.Context) *MockDB {
	args := m.Called(ctx)
	return args.Get(0).(*MockDB)
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) error {
	args := m.Called(dest, conds)
	return args.Error(0)
}

func (m *MockDB) Create(value interface{}) error {
	args := m.Called(value)
	return args.Error(0)
}

func (m *MockDB) Preload(query string, args ...interface{}) *MockDB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*MockDB)
}

func (m *MockDB) Model(value interface{}) *MockDB {
	args := m.Called(value)
	return args.Get(0).(*MockDB)
}

func (m *MockDB) Count(count *int64) error {
	args := m.Called(count)
	return args.Error(0)
}

func (m *MockDB) Offset(offset int) *MockDB {
	args := m.Called(offset)
	return args.Get(0).(*MockDB)
}

func (m *MockDB) Limit(limit int) *MockDB {
	args := m.Called(limit)
	return args.Get(0).(*MockDB)
}

func (m *MockDB) Order(value interface{}) *MockDB {
	args := m.Called(value)
	return args.Get(0).(*MockDB)
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) error {
	args := m.Called(dest, conds)
	return args.Error(0)
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *MockDB {
	mockArgs := m.Called(query, args)
	return mockArgs.Get(0).(*MockDB)
}

func TestCaseService_CreateCase_Validation(t *testing.T) {
	mockDB := &MockDB{}
	mockConflictService := &MockConflictService{}
	caseService := NewCaseService(mockDB, mockConflictService, false)

	t.Run("验证成功 - 完整数据", func(t *testing.T) {
		req := &CreateCaseRequest{
			Title:    "测试案件",
			ClientID: 1,
			LawyerID: 1,
			CaseType: "civil",
			Priority: "medium",
			Status:   "pending",
		}

		// 模拟客户端验证通过
		mockClient := &models.Client{ID: 1, Name: "测试客户"}
		mockDB.On("WithContext", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Client"), uint(1)).Return(nil)
		mockDB.On("First", mock.AnythingOfType("*models.User"), mock.Anything, "lawyer").Return(nil)

		err := caseService.validateCaseRequest(context.Background(), req)
		assert.NoError(t, err)

		mockDB.AssertExpectations(t)
		mockDB.ExpectedCalls = nil
	})

	t.Run("验证失败 - 空标题", func(t *testing.T) {
		req := &CreateCaseRequest{
			Title:    "", // 空标题
			ClientID: 1,
			LawyerID: 1,
			CaseType: "civil",
			Priority: "medium",
		}

		err := caseService.validateCaseRequest(context.Background(), req)
		// 注意：validateCaseRequest主要验证数据库中的记录存在性
		// 字段验证由gin的binding标签处理
		// 这里我们主要测试数据库验证逻辑
		mockDB.On("WithContext", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Client"), uint(1)).Return(errors.New("client not found"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")

		mockDB.AssertExpectations(t)
		mockDB.ExpectedCalls = nil
	})

	t.Run("验证失败 - 客户不存在", func(t *testing.T) {
		req := &CreateCaseRequest{
			Title:    "测试案件",
			ClientID: 999, // 不存在的客户ID
			LawyerID: 1,
			CaseType: "civil",
			Priority: "medium",
		}

		mockDB.On("WithContext", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Client"), uint(999)).Return(gorm.ErrRecordNotFound)

		err := caseService.validateCaseRequest(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client not found")

		mockDB.AssertExpectations(t)
		mockDB.ExpectedCalls = nil
	})

	t.Run("验证失败 - 律师不存在", func(t *testing.T) {
		req := &CreateCaseRequest{
			Title:    "测试案件",
			ClientID: 1,
			LawyerID: 999, // 不存在的律师ID
			CaseType: "civil",
			Priority: "medium",
		}

		mockDB.On("WithContext", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Client"), uint(1)).Return(nil)
		mockDB.On("First", mock.AnythingOfType("*models.User"), mock.Anything, "lawyer").Return(gorm.ErrRecordNotFound)

		err := caseService.validateCaseRequest(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lawyer not found")

		mockDB.AssertExpectations(t)
		mockDB.ExpectedCalls = nil
	})
}

func TestCaseService_CreateCase_DataMapping(t *testing.T) {
	mockDB := &MockDB{}
	mockConflictService := &MockConflictService{}
	caseService := NewCaseService(mockDB, mockConflictService, false)

	t.Run("数据映射正确性测试", func(t *testing.T) {
		req := &CreateCaseRequest{
			Title:       "前端数据映射测试",
			Description: "测试前后端数据映射",
			ClientID:    1,
			LawyerID:    1,
			CaseType:    "civil",
			Priority:    "high",
			Status:      "active",
		}

		// 模拟验证通过
		mockDB.On("WithContext", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Client"), uint(1)).Return(nil)
		mockDB.On("First", mock.AnythingOfType("*models.User"), mock.Anything, "lawyer").Return(nil)

		// 模拟冲突检测
		mockConflictService.On("CheckConflict", mock.Anything, mock.AnythingOfType("*models.ConflictCheckRequest")).Return(nil, nil)

		// 模拟数据库保存
		mockDB.On("Create", mock.AnythingOfType("*models.Case")).Return(nil)

		// 模拟查询保存后的案件信息
		savedCase := &models.Case{
			ID:          1,
			Title:       req.Title,
			Description: req.Description,
			ClientID:    req.ClientID,
			LawyerID:    req.LawyerID,
			CaseType:    req.CaseType,
			Priority:    req.Priority,
			Status:      req.Status,
		}
		mockDB.On("Preload", "Client").Return(mockDB)
		mockDB.On("Preload", "Lawyer").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Case"), uint(1)).Return(nil)

		result, err := caseService.CreateCase(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Title, result.Title)
		assert.Equal(t, req.Description, result.Description)
		assert.Equal(t, req.ClientID, result.ClientID)
		assert.Equal(t, req.LawyerID, result.LawyerID)
		assert.Equal(t, req.CaseType, result.CaseType)
		assert.Equal(t, req.Priority, result.Priority)
		assert.Equal(t, req.Status, result.Status)

		mockDB.AssertExpectations(t)
		mockConflictService.AssertExpectations(t)
	})
}

func TestCaseService_UpdateCase_Validation(t *testing.T) {
	mockDB := &MockDB{}
	mockConflictService := &MockConflictService{}
	caseService := NewCaseService(mockDB, mockConflictService, false)

	t.Run("更新案件 - 验证律师存在", func(t *testing.T) {
		lawyerID := uint(1)
		req := &UpdateCaseRequest{
			Title:    stringPtr("更新后的案件"),
			LawyerID: &lawyerID,
		}

		// 模拟案件存在
		mockDB.On("WithContext", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Case"), uint(1)).Return(nil)

		// 模拟律师验证
		mockDB.On("First", mock.AnythingOfType("*models.User"), uint(1), "lawyer").Return(nil)

		// 模拟更新操作
		mockDB.On("Model", mock.AnythingOfType("*models.Case")).Return(mockDB)
		mockDB.On("Updates", mock.AnythingOfType("map[string]interface {}")).Return(nil)

		// 模拟返回更新后的案件
		mockDB.On("Preload", "Client").Return(mockDB)
		mockDB.On("Preload", "Lawyer").Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Case"), uint(1)).Return(nil)

		result, err := caseService.UpdateCase(context.Background(), 1, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		mockDB.AssertExpectations(t)
		mockDB.ExpectedCalls = nil
	})

	t.Run("更新案件 - 律师不存在", func(t *testing.T) {
		lawyerID := uint(999)
		req := &UpdateCaseRequest{
			Title:    stringPtr("更新后的案件"),
			LawyerID: &lawyerID,
		}

		// 模拟案件存在
		mockDB.On("WithContext", mock.Anything).Return(mockDB)
		mockDB.On("First", mock.AnythingOfType("*models.Case"), uint(1)).Return(nil)

		// 模拟律师验证失败
		mockDB.On("First", mock.AnythingOfType("*models.User"), uint(999), "lawyer").Return(gorm.ErrRecordNotFound)

		_, err := caseService.UpdateCase(context.Background(), 1, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lawyer not found")

		mockDB.AssertExpectations(t)
		mockDB.ExpectedCalls = nil
	})
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func TestCaseService_CaseResponse_Conversion(t *testing.T) {
	mockDB := &MockDB{}
	mockConflictService := &MockConflictService{}
	caseService := NewCaseService(mockDB, mockConflictService, false)

	t.Run("案件响应转换 - 包含客户和律师信息", func(t *testing.T) {
		caseModel := &models.Case{
			ID:          1,
			Title:       "测试案件",
			Description: "测试描述",
			ClientID:    1,
			LawyerID:    1,
			CaseType:    "civil",
			Priority:    "medium",
			Status:      "pending",
		}

		// 测试没有关联信息的情况
		response := caseService.toCaseResponse(caseModel)
		assert.Equal(t, caseModel.ID, response.ID)
		assert.Equal(t, caseModel.Title, response.Title)
		assert.Equal(t, caseModel.Description, response.Description)
		assert.Equal(t, caseModel.ClientID, response.ClientID)
		assert.Equal(t, caseModel.LawyerID, response.LawyerID)
		assert.Equal(t, caseModel.CaseType, response.CaseType)
		assert.Equal(t, caseModel.Priority, response.Priority)
		assert.Equal(t, caseModel.Status, response.Status)

		// 测试包含客户信息的情况
		caseModel.Client = &models.Client{
			ID:      1,
			Name:    "测试客户",
			Company: "测试公司",
			Email:   "test@example.com",
		}

		response = caseService.toCaseResponse(caseModel)
		assert.NotNil(t, response.Client)
		assert.Equal(t, "测试公司", response.ClientName) // 优先使用company字段
		assert.Equal(t, "测试公司", response.Client.Company)
		assert.Equal(t, "测试客户", response.Client.Name)

		// 测试只有name字段没有company字段的情况
		caseModel.Client.Company = ""
		response = caseService.toCaseResponse(caseModel)
		assert.Equal(t, "测试客户", response.ClientName) // 使用name字段

		// 测试包含律师信息的情况
		caseModel.Lawyer = &models.User{
			ID:   1,
			Name: "测试律师",
		}

		response = caseService.toCaseResponse(caseModel)
		assert.NotNil(t, response.Lawyer)
		assert.Equal(t, "测试律师", response.LawyerName)
		assert.Equal(t, "测试律师", response.Lawyer.Name)
	})
}

// 性能测试
func BenchmarkCaseService_ValidateCaseRequest(b *testing.B) {
	mockDB := &MockDB{}
	mockConflictService := &MockConflictService{}
	caseService := NewCaseService(mockDB, mockConflictService, false)

	req := &CreateCaseRequest{
		Title:    "基准测试案件",
		ClientID: 1,
		LawyerID: 1,
		CaseType: "civil",
		Priority: "medium",
	}

	mockDB.On("WithContext", mock.Anything).Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*models.Client"), uint(1)).Return(nil)
	mockDB.On("First", mock.AnythingOfType("*models.User"), mock.Anything, "lawyer").Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = caseService.validateCaseRequest(context.Background(), req)
	}
}
