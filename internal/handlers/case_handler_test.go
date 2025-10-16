package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// MockCaseService 是案件服务的模拟实现
type MockCaseService struct {
	mock.Mock
}

func (m *MockCaseService) CreateCase(ctx interface{}, req *services.CreateCaseRequest) (*services.CaseResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CaseResponse), args.Error(1)
}

func (m *MockCaseService) GetCaseByID(ctx interface{}, id uint) (*services.CaseResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CaseResponse), args.Error(1)
}

func (m *MockCaseService) UpdateCase(ctx interface{}, id uint, req *services.UpdateCaseRequest) (*services.CaseResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CaseResponse), args.Error(1)
}

func (m *MockCaseService) DeleteCase(ctx interface{}, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCaseService) ListCases(ctx interface{}, req *services.CaseListRequest) ([]*services.CaseResponse, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1), args.Error(2)
	}
	return args.Get(0).([]*services.CaseResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockCaseService) GetCaseStats(ctx interface{}) (*services.CaseStatsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CaseStatsResponse), args.Error(1)
}

func (m *MockCaseService) AssignLawyer(ctx interface{}, caseID, lawyerID uint) error {
	args := m.Called(ctx, caseID, lawyerID)
	return args.Error(0)
}

func (m *MockCaseService) UpdateCaseStatus(ctx interface{}, caseID uint, status string) error {
	args := m.Called(ctx, caseID, status)
	return args.Error(0)
}

func setupTestRouter() (*gin.Engine, *MockCaseService) {
	gin.SetMode(gin.TestMode)

	mockService := &MockCaseService{}
	caseHandler := NewCaseHandler(mockService)

	router := gin.New()

	// 添加错误处理中间件
	router.Use(func(c *gin.Context) {
		c.Next()

		// 处理错误
		for _, err := range c.Errors {
			common.HandleError(c, err.Err)
			return
		}
	})

	v1 := router.Group("/api/v1")
	{
		v1.POST("/cases", caseHandler.CreateCase)
		v1.PUT("/cases/:id", caseHandler.UpdateCase)
		v1.GET("/cases/:id", caseHandler.GetCase)
		v1.DELETE("/cases/:id", caseHandler.DeleteCase)
		v1.GET("/cases", caseHandler.ListCases)
		v1.GET("/cases/stats", caseHandler.GetCaseStats)
		v1.POST("/cases/:id/assign-lawyer", caseHandler.AssignLawyer)
		v1.PUT("/cases/:id/status", caseHandler.UpdateCaseStatus)
	}

	return router, mockService
}

func TestCreateCase_Success(t *testing.T) {
	router, mockService := setupTestRouter()

	expectedResponse := &services.CaseResponse{
		ID:       1,
		Title:    "测试案件",
		CaseType: "civil",
		Priority: "medium",
		Status:   "pending",
		ClientID: 1,
		LawyerID: 1,
	}

	mockService.On("CreateCase", mock.Anything, mock.AnythingOfType("*services.CreateCaseRequest")).Return(expectedResponse, nil)

	requestBody := map[string]interface{}{
		"title":       "测试案件",
		"description": "测试描述",
		"client_id":   1,
		"lawyer_id":   1,
		"case_type":   "civil",
		"priority":    "medium",
		"status":      "pending",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/cases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	mockService.AssertExpectations(t)
}

func TestCreateCase_ValidationError_EmptyTitle(t *testing.T) {
	router, mockService := setupTestRouter()

	requestBody := map[string]interface{}{
		"title":       "", // 空标题应该导致验证失败
		"description": "测试描述",
		"client_id":   1,
		"lawyer_id":   1,
		"case_type":   "civil",
		"priority":    "medium",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/cases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Contains(t, response.Error.Message, "请求参数验证失败")

	mockService.AssertNotCalled(t, "CreateCase")
}

func TestCreateCase_ValidationError_MissingRequiredFields(t *testing.T) {
	router, mockService := setupTestRouter()

	// 测试缺少必填字段的情况
	testCases := []struct {
		name     string
		body     map[string]interface{}
		expected string
	}{
		{
			name: "缺少client_id",
			body: map[string]interface{}{
				"title":     "测试案件",
				"lawyer_id": 1,
				"case_type": "civil",
				"priority":  "medium",
			},
			expected: "client_id",
		},
		{
			name: "缺少lawyer_id",
			body: map[string]interface{}{
				"title":     "测试案件",
				"client_id": 1,
				"case_type": "civil",
				"priority":  "medium",
			},
			expected: "lawyer_id",
		},
		{
			name: "缺少case_type",
			body: map[string]interface{}{
				"title":     "测试案件",
				"client_id": 1,
				"lawyer_id": 1,
				"priority":  "medium",
			},
			expected: "case_type",
		},
		{
			name: "缺少priority",
			body: map[string]interface{}{
				"title":     "测试案件",
				"client_id": 1,
				"lawyer_id": 1,
				"case_type": "civil",
			},
			expected: "priority",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/api/v1/cases", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response common.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

			mockService.AssertNotCalled(t, "CreateCase")
			mockService.ExpectedCalls = nil // 重置mock调用期望
		})
	}
}

func TestCreateCase_ValidationError_InvalidEnumValues(t *testing.T) {
	router, mockService := setupTestRouter()

	// 测试无效的枚举值
	testCases := []struct {
		name     string
		body     map[string]interface{}
		expected string
	}{
		{
			name: "无效的case_type",
			body: map[string]interface{}{
				"title":     "测试案件",
				"client_id": 1,
				"lawyer_id": 1,
				"case_type": "invalid_type",
				"priority":  "medium",
			},
			expected: "case_type",
		},
		{
			name: "无效的priority",
			body: map[string]interface{}{
				"title":     "测试案件",
				"client_id": 1,
				"lawyer_id": 1,
				"case_type": "civil",
				"priority":  "invalid_priority",
			},
			expected: "priority",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			req := httptest.NewRequest("POST", "/api/v1/cases", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response common.APIResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

			mockService.AssertNotCalled(t, "CreateCase")
			mockService.ExpectedCalls = nil // 重置mock调用期望
		})
	}
}

func TestCreateCase_ValidationError_EmptyRequestBody(t *testing.T) {
	router, mockService := setupTestRouter()

	req := httptest.NewRequest("POST", "/api/v1/cases", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

	mockService.AssertNotCalled(t, "CreateCase")
}

func TestUpdateCase_Success(t *testing.T) {
	router, mockService := setupTestRouter()

	expectedResponse := &services.CaseResponse{
		ID:       1,
		Title:    "更新后的测试案件",
		CaseType: "commercial",
		Priority: "high",
		Status:   "active",
		ClientID: 1,
		LawyerID: 1,
	}

	mockService.On("UpdateCase", mock.Anything, uint(1), mock.AnythingOfType("*services.UpdateCaseRequest")).Return(expectedResponse, nil)

	requestBody := map[string]interface{}{
		"title":     "更新后的测试案件",
		"case_type": "commercial",
		"priority":  "high",
		"status":    "active",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/api/v1/cases/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	mockService.AssertExpectations(t)
}

func TestUpdateCase_ValidationError_InvalidEnumValues(t *testing.T) {
	router, mockService := setupTestRouter()

	requestBody := map[string]interface{}{
		"case_type": "invalid_type",
		"priority":  "invalid_priority",
		"status":    "invalid_status",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PUT", "/api/v1/cases/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response common.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Contains(t, response.Error.Message, "更新参数验证失败")

	mockService.AssertNotCalled(t, "UpdateCase")
}

// 测试前端兼容性 - 验证前端数据格式能够正确映射
func TestCreateCase_FrontendCompatibility(t *testing.T) {
	router, mockService := setupTestRouter()

	expectedResponse := &services.CaseResponse{
		ID:       1,
		Title:    "前端兼容性测试",
		CaseType: "civil",
		Priority: "medium",
		Status:   "pending",
		ClientID: 1,
		LawyerID: 1,
	}

	mockService.On("CreateCase", mock.Anything, mock.AnythingOfType("*services.CreateCaseRequest")).Return(expectedResponse, nil)

	// 模拟前端发送的数据格式（经过前端服务转换后）
	requestBody := map[string]interface{}{
		"title":       "前端兼容性测试",  // 前端会将caseName转换为title
		"description": "测试描述",
		"client_id":   1,             // 前端会将clientId转换为client_id
		"lawyer_id":   1,             // 前端会将lawyerId转换为lawyer_id
		"case_type":   "civil",       // 前端会将caseType转换为case_type
		"priority":    "medium",
		"status":      "pending",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/cases", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	mockService.AssertExpectations(t)
}

// 基准测试 - 验证修复后的性能
func BenchmarkCreateCase_Success(b *testing.B) {
	router, mockService := setupTestRouter()

	expectedResponse := &services.CaseResponse{
		ID:       1,
		Title:    "基准测试案件",
		CaseType: "civil",
		Priority: "medium",
		Status:   "pending",
		ClientID: 1,
		LawyerID: 1,
	}

	mockService.On("CreateCase", mock.Anything, mock.AnythingOfType("*services.CreateCaseRequest")).Return(expectedResponse, nil)

	requestBody := map[string]interface{}{
		"title":       "基准测试案件",
		"description": "基准测试描述",
		"client_id":   1,
		"lawyer_id":   1,
		"case_type":   "civil",
		"priority":    "medium",
		"status":      "pending",
	}

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/v1/cases", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}