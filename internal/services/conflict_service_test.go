package services

import (
	"context"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockConflictRepository Mock冲突检测仓储
type MockConflictRepository struct {
	mock.Mock
}

func (m *MockConflictRepository) SaveCheckRecord(ctx context.Context, record *models.ConflictCheckRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockConflictRepository) GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error) {
	args := m.Called(ctx, clientID, limit)
	return args.Get(0).([]*models.ConflictCheckRecord), args.Error(1)
}

func (m *MockConflictRepository) GetCheckDetails(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error) {
	args := m.Called(ctx, checkID)
	return args.Get(0).(*models.ConflictCheckRecord), args.Error(1)
}

func (m *MockConflictRepository) SaveCheckResult(ctx context.Context, result *models.ConflictCheckResponse) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockConflictRepository) GetConflictRules(ctx context.Context) ([]*models.ConflictRule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.ConflictRule), args.Error(1)
}

func (m *MockConflictRepository) UpdateConflictRule(ctx context.Context, rule *models.ConflictRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockConflictRepository) DeleteConflictRule(ctx context.Context, ruleID string) error {
	args := m.Called(ctx, ruleID)
	return args.Error(0)
}

// MockMCPClient Mock MCP客户端
type MockMCPClient struct {
	mock.Mock
}

func (m *MockMCPClient) FetchLatestStandards(ctx context.Context) (*models.MCPStandards, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.MCPStandards), args.Error(1)
}

// MockRuleEngine Mock规则引擎
type MockRuleEngine struct {
	mock.Mock
}

func (m *MockRuleEngine) EvaluateRules(ctx context.Context, request *models.ConflictCheckRequest, rules []*models.ConflictRule) (*models.ConflictCheckResponse, error) {
	args := m.Called(ctx, request, rules)
	return args.Get(0).(*models.ConflictCheckResponse), args.Error(1)
}

// MockRiskAssessor Mock风险评估器
type MockRiskAssessor struct {
	mock.Mock
}

func (m *MockRiskAssessor) AssessRisk(ctx context.Context, response *models.ConflictCheckResponse) (*models.RiskAssessment, error) {
	args := m.Called(ctx, response)
	return args.Get(0).(*models.RiskAssessment), args.Error(1)
}

// TestNewConflictService 测试冲突检测服务创建
func TestNewConflictService(t *testing.T) {
	mockRepo := &MockConflictRepository{}
	mockMCPClient := &MockMCPClient{}
	mockRuleEngine := &MockRuleEngine{}
	mockRiskAssessor := &MockRiskAssessor{}

	service := NewConflictService(mockRepo, mockMCPClient, mockRuleEngine, mockRiskAssessor)

	assert.NotNil(t, service)
}

// TestConflictServiceImpl_CheckConflict_Success 测试冲突检测成功
func TestConflictServiceImpl_CheckConflict_Success(t *testing.T) {
	// 准备测试数据
	mockRepo := &MockConflictRepository{}
	mockMCPClient := &MockMCPClient{}
	mockRuleEngine := &MockRuleEngine{}
	mockRiskAssessor := &MockRiskAssessor{}

	service := NewConflictService(mockRepo, mockMCPClient, mockRuleEngine, mockRiskAssessor)

	ctx := context.Background()
	request := &models.ConflictCheckRequest{
		ClientID:                     "CLIENT001",
		ClientName:                   "测试客户",
		CaseName:                     "测试案件",
		CaseType:                     "civil",
		ClientType:                   "individual",
		OtherParties:                 []string{"对方当事人"},
		SearchYears:                  5,
		IncludeCorporateRelations:    true,
		SearchDepth:                  "STANDARD",
		UserID:                       1,
	}

	// Mock返回数据
	mcpStandards := &models.MCPStandards{
		Version:       "1.0",
		LastUpdated:   time.Now(),
		BestPractices: []string{"规则1", "规则2"},
	}

	rules := []*models.ConflictRule{
		{
			ID:          "RULE001",
			Name:        "基本冲突检测规则",
			Description: "检测基本的利益冲突",
			Type:        "basic",
			Active:      true,
			Priority:    1,
		},
	}

	expectedResponse := &models.ConflictCheckResponse{
		CheckID:      "CC_CLIENT001_1234567890",
		ClientID:     request.ClientID,
		ClientName:   request.ClientName,
		CaseName:     request.CaseName,
		CaseType:     request.CaseType,
		CheckResult:  "NO_CONFLICT",
		ConflictLevel: "LOW",
		RiskScore:    10,
		CheckTime:    time.Now(),
	}

	riskAssessment := &models.RiskAssessment{
		RiskLevel:    "LOW",
		RiskScore:    10,
		Recommendations: []string{"建议继续"},
	}

	// 设置Mock期望
	mockRepo.On("SaveCheckRecord", ctx, mock.AnythingOfType("*models.ConflictCheckRecord")).Return(nil)
	mockMCPClient.On("FetchLatestStandards", ctx).Return(mcpStandards, nil)
	mockRepo.On("GetConflictRules", ctx).Return(rules, nil)
	mockRuleEngine.On("EvaluateRules", ctx, request, rules).Return(expectedResponse, nil)
	mockRiskAssessor.On("AssessRisk", ctx, expectedResponse).Return(riskAssessment, nil)
	mockRepo.On("SaveCheckResult", ctx, mock.AnythingOfType("*models.ConflictCheckResponse")).Return(nil)

	// 执行测试
	result, err := service.CheckConflict(ctx, request)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, request.ClientID, result.ClientID)
	assert.Equal(t, request.ClientName, result.ClientName)
	assert.Equal(t, request.CaseName, result.CaseName)
	assert.Equal(t, request.CaseType, result.CaseType)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
	mockMCPClient.AssertExpectations(t)
	mockRuleEngine.AssertExpectations(t)
	mockRiskAssessor.AssertExpectations(t)
}

// TestConflictServiceImpl_CheckConflict_ValidationError 测试冲突检测验证错误
func TestConflictServiceImpl_CheckConflict_ValidationError(t *testing.T) {
	service := &ConflictServiceImpl{}

	ctx := context.Background()
	request := &models.ConflictCheckRequest{
		ClientID: "", // 空的客户端ID应该导致验证错误
	}

	// 执行测试
	result, err := service.CheckConflict(ctx, request)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "请求验证失败")
}

// TestConflictServiceImpl_GetCheckHistory_Success 测试获取检查历史成功
func TestConflictServiceImpl_GetCheckHistory_Success(t *testing.T) {
	mockRepo := &MockConflictRepository{}
	service := &ConflictServiceImpl{
		conflictRepo: mockRepo,
	}

	ctx := context.Background()
	clientID := "CLIENT001"
	limit := 10

	expectedHistory := []*models.ConflictCheckRecord{
		{
			CheckID:    "CC_CLIENT001_1234567890",
			ClientID:   clientID,
			ClientName: "测试客户",
			CheckStatus: "COMPLETED",
			CheckTime:  time.Now(),
		},
	}

	// 设置Mock期望
	mockRepo.On("GetCheckHistory", ctx, clientID, limit).Return(expectedHistory, nil)

	// 执行测试
	result, err := service.GetCheckHistory(ctx, clientID, limit)

	// 验证结果
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, clientID, result[0].ClientID)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestConflictServiceImpl_GetCheckDetails_Success 测试获取检查详情成功
func TestConflictServiceImpl_GetCheckDetails_Success(t *testing.T) {
	mockRepo := &MockConflictRepository{}
	service := &ConflictServiceImpl{
		conflictRepo: mockRepo,
	}

	ctx := context.Background()
	checkID := "CC_CLIENT001_1234567890"

	expectedDetails := &models.ConflictCheckRecord{
		CheckID:     checkID,
		ClientID:    "CLIENT001",
		ClientName:  "测试客户",
		CaseName:    "测试案件",
		CheckStatus: "COMPLETED",
		CheckTime:   time.Now(),
	}

	// 设置Mock期望
	mockRepo.On("GetCheckDetails", ctx, checkID).Return(expectedDetails, nil)

	// 执行测试
	result, err := service.GetCheckDetails(ctx, checkID)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, checkID, result.CheckID)
	assert.Equal(t, "CLIENT001", result.ClientID)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestConflictServiceImpl_GetConflictRules_Success 测试获取冲突规则成功
func TestConflictServiceImpl_GetConflictRules_Success(t *testing.T) {
	mockRepo := &MockConflictRepository{}
	service := &ConflictServiceImpl{
		conflictRepo: mockRepo,
	}

	ctx := context.Background()

	expectedRules := []*models.ConflictRule{
		{
			ID:          "RULE001",
			Name:        "基本冲突检测规则",
			Description: "检测基本的利益冲突",
			RuleType:    "basic",
			Enabled:     true,
			Priority:    1,
		},
		{
			ID:          "RULE002",
			Name:        "高级冲突检测规则",
			Description: "检测复杂的利益冲突",
			RuleType:    "advanced",
			Enabled:     true,
			Priority:    2,
		},
	}

	// 设置Mock期望
	mockRepo.On("GetConflictRules", ctx).Return(expectedRules, nil)

	// 执行测试
	result, err := service.GetConflictRules(ctx)

	// 验证结果
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "RULE001", result[0].ID)
	assert.Equal(t, "RULE002", result[1].ID)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestConflictServiceImpl_UpdateConflictRule_Success 测试更新冲突规则成功
func TestConflictServiceImpl_UpdateConflictRule_Success(t *testing.T) {
	mockRepo := &MockConflictRepository{}
	service := &ConflictServiceImpl{
		conflictRepo: mockRepo,
	}

	ctx := context.Background()

	rule := &models.ConflictRule{
		ID:          "RULE001",
		Name:        "更新后的规则",
		Description: "更新后的描述",
		RuleType:    "basic",
		Enabled:     true,
		Priority:    1,
	}

	// 设置Mock期望
	mockRepo.On("UpdateConflictRule", ctx, rule).Return(nil)

	// 执行测试
	err := service.UpdateConflictRule(ctx, rule)

	// 验证结果
	assert.NoError(t, err)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestConflictServiceImpl_DeleteConflictRule_Success 测试删除冲突规则成功
func TestConflictServiceImpl_DeleteConflictRule_Success(t *testing.T) {
	mockRepo := &MockConflictRepository{}
	service := &ConflictServiceImpl{
		conflictRepo: mockRepo,
	}

	ctx := context.Background()
	ruleID := "RULE001"

	// 设置Mock期望
	mockRepo.On("DeleteConflictRule", ctx, ruleID).Return(nil)

	// 执行测试
	err := service.DeleteConflictRule(ctx, ruleID)

	// 验证结果
	assert.NoError(t, err)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
}

// TestConflictServiceImpl_GetMCPStandards_Success 测试获取MCP标准成功
func TestConflictServiceImpl_GetMCPStandards_Success(t *testing.T) {
	mockMCPClient := &MockMCPClient{}
	service := &ConflictServiceImpl{
		mcpClient: mockMCPClient,
	}

	ctx := context.Background()

	expectedStandards := &models.MCPStandards{
		Version:     "1.0",
		LastUpdated: time.Now(),
		Rules:       []string{"规则1", "规则2"},
	}

	// 设置Mock期望
	mockMCPClient.On("FetchLatestStandards", ctx).Return(expectedStandards, nil)

	// 执行测试
	result, err := service.GetMCPStandards(ctx)

	// 验证结果
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "1.0", result.Version)
	assert.Len(t, result.Rules, 2)

	// 验证Mock调用
	mockMCPClient.AssertExpectations(t)
}

// TestConflictServiceImpl_GetMCPStandards_NoClient 测试没有MCP客户端时的处理
func TestConflictServiceImpl_GetMCPStandards_NoClient(t *testing.T) {
	service := &ConflictServiceImpl{
		mcpClient: nil,
	}

	ctx := context.Background()

	// 执行测试
	result, err := service.GetMCPStandards(ctx)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MCP客户端不可用")
}

// TestConflictServiceImpl_CheckConflict_MCPFailure 测试MCP服务失败时的处理
func TestConflictServiceImpl_CheckConflict_MCPFailure(t *testing.T) {
	mockRepo := &MockConflictRepository{}
	mockMCPClient := &MockMCPClient{}
	mockRuleEngine := &MockRuleEngine{}
	mockRiskAssessor := &MockRiskAssessor{}

	service := NewConflictService(mockRepo, mockMCPClient, mockRuleEngine, mockRiskAssessor)

	ctx := context.Background()
	request := &models.ConflictCheckRequest{
		ClientID:     "CLIENT001",
		ClientName:   "测试客户",
		CaseName:     "测试案件",
		CaseType:     "civil",
		ClientType:   "individual",
		SearchYears:  5,
		UserID:       1,
	}

	rules := []*models.ConflictRule{
		{
			ID:       "RULE001",
			Name:     "基本冲突检测规则",
			Enabled:  true,
			Priority: 1,
		},
	}

	expectedResponse := &models.ConflictCheckResponse{
		CheckID:      "CC_CLIENT001_1234567890",
		ClientID:     request.ClientID,
		CheckResult:  "NO_CONFLICT",
		ConflictLevel: "LOW",
		RiskScore:    10,
		CheckTime:    time.Now(),
	}

	riskAssessment := &models.RiskAssessment{
		RiskLevel: "LOW",
		RiskScore: 10,
	}

	// 设置Mock期望 - MCP失败
	mockRepo.On("SaveCheckRecord", ctx, mock.AnythingOfType("*models.ConflictCheckRecord")).Return(nil)
	mockMCPClient.On("FetchLatestStandards", ctx).Return(nil, assert.AnError)
	mockRepo.On("GetConflictRules", ctx).Return(rules, nil)
	mockRuleEngine.On("EvaluateRules", ctx, request, rules).Return(expectedResponse, nil)
	mockRiskAssessor.On("AssessRisk", ctx, expectedResponse).Return(riskAssessment, nil)
	mockRepo.On("SaveCheckResult", ctx, mock.AnythingOfType("*models.ConflictCheckResponse")).Return(nil)

	// 执行测试
	result, err := service.CheckConflict(ctx, request)

	// 验证结果 - MCP失败应该不影响主要功能
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, request.ClientID, result.ClientID)

	// 验证Mock调用
	mockRepo.AssertExpectations(t)
	mockMCPClient.AssertExpectations(t)
	mockRuleEngine.AssertExpectations(t)
	mockRiskAssessor.AssertExpectations(t)
}

// BenchmarkConflictServiceImpl_CheckConflict 基准测试冲突检测性能
func BenchmarkConflictServiceImpl_CheckConflict(b *testing.B) {
	mockRepo := &MockConflictRepository{}
	mockMCPClient := &MockMCPClient{}
	mockRuleEngine := &MockRuleEngine{}
	mockRiskAssessor := &MockRiskAssessor{}

	service := NewConflictService(mockRepo, mockMCPClient, mockRuleEngine, mockRiskAssessor)

	ctx := context.Background()
	request := &models.ConflictCheckRequest{
		ClientID:     "CLIENT001",
		ClientName:   "基准测试客户",
		CaseName:     "基准测试案件",
		CaseType:     "civil",
		ClientType:   "individual",
		SearchYears:  5,
		UserID:       1,
	}

	// 设置Mock期望
	mockRepo.On("SaveCheckRecord", ctx, mock.AnythingOfType("*models.ConflictCheckRecord")).Return(nil)
	mockMCPClient.On("FetchLatestStandards", ctx).Return(&models.MCPStandards{}, nil)
	mockRepo.On("GetConflictRules", ctx).Return([]*models.ConflictRule{}, nil)
	mockRuleEngine.On("EvaluateRules", ctx, request, mock.AnythingOfType("[]*models.ConflictRule")).Return(&models.ConflictCheckResponse{}, nil)
	mockRiskAssessor.On("AssessRisk", ctx, mock.AnythingOfType("*models.ConflictCheckResponse")).Return(&models.RiskAssessment{}, nil)
	mockRepo.On("SaveCheckResult", ctx, mock.AnythingOfType("*models.ConflictCheckResponse")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CheckConflict(ctx, request)
		if err != nil {
			b.Fatal(err)
		}
	}
}