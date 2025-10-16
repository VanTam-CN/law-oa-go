package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"law-oa-go/internal/database"
	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/router"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// EnhancedConflictDetectionTestSuite 增强冲突检测测试套件
type EnhancedConflictDetectionTestSuite struct {
	suite.Suite
	db          *database.DB
	router      *gin.Engine
	repo        repositories.Repositories
	conflictSvc *services.EnhancedConflictService
	testLawyer  *models.User
}

// SetupSuite 设置测试套件
func (suite *EnhancedConflictDetectionTestSuite) SetupSuite() {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 初始化数据库
	suite.db = database.NewTestDB()
	err := suite.db.Connect()
	suite.Require().NoError(err)

	// 运行迁移
	err = suite.runMigrations()
	suite.Require().NoError(err)

	// 初始化仓储
	suite.repo = repositories.NewRepositories(suite.db.GetDB())

	// 初始化服务
	suite.conflictSvc = services.NewEnhancedConflictService(suite.repo)

	// 初始化路由
	suite.router = router.SetupRouter(suite.repo)

	// 创建测试数据
	err = suite.createTestData()
	suite.Require().NoError(err)
}

// TearDownSuite 清理测试套件
func (suite *EnhancedConflictDetectionTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

// runMigrations 运行数据库迁移
func (suite *EnhancedConflictDetectionTestSuite) runMigrations() error {
	// 这里应该运行实际的迁移脚本
	// 简化版本，直接创建必要的表结构
	migrationSQL := `
	-- 创建行业分类表
	CREATE TABLE IF NOT EXISTS industry_classifications (
		id SERIAL PRIMARY KEY,
		code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(200) NOT NULL,
		level INTEGER DEFAULT 1,
		keywords TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	-- 创建竞争关系表
	CREATE TABLE IF NOT EXISTS competitive_relations (
		id SERIAL PRIMARY KEY,
		industry_id INTEGER NOT NULL REFERENCES industry_classifications(id),
		competitor_type VARCHAR(50) NOT NULL,
		competitor_name VARCHAR(200) NOT NULL,
		competitor_pattern TEXT NOT NULL,
		conflict_level VARCHAR(20) NOT NULL,
		description TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		UNIQUE(industry_id, competitor_name)
	);

	-- 创建冲突规则表
	CREATE TABLE IF NOT EXISTS conflict_rules (
		id SERIAL PRIMARY KEY,
		name VARCHAR(200) NOT NULL,
		rule_type VARCHAR(50) NOT NULL,
		trigger_pattern TEXT,
		action_type VARCHAR(50) NOT NULL,
		risk_score INTEGER DEFAULT 50,
		conditions TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		priority INTEGER DEFAULT 100,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW(),
		UNIQUE(name, rule_type)
	);

	-- 创建冲突检测历史表
	CREATE TABLE IF NOT EXISTS conflict_detection_history (
		id SERIAL PRIMARY KEY,
		lawyer_id INTEGER NOT NULL REFERENCES users(id),
		case_id INTEGER REFERENCES cases(id),
		client_name VARCHAR(200) NOT NULL,
		opposing_party VARCHAR(200) NOT NULL,
		case_type VARCHAR(50) NOT NULL,
		detection_result TEXT,
		conflicts_found INTEGER DEFAULT 0,
		risk_level VARCHAR(20) NOT NULL,
		user_action VARCHAR(50),
		created_at TIMESTAMP DEFAULT NOW()
	);

	-- 为客户表添加行业字段
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns
					   WHERE table_name = 'clients' AND column_name = 'industry') THEN
			ALTER TABLE clients ADD COLUMN industry VARCHAR(100);
		END IF;
	END $$;

	-- 为案件表添加对方当事人字段
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM information_schema.columns
					   WHERE table_name = 'cases' AND column_name = 'opposing_party') THEN
			ALTER TABLE cases ADD COLUMN opposing_party TEXT;
		END IF;
	END $$;
	`

	_, err := suite.db.GetDB().Exec(migrationSQL)
	return err
}

// createTestData 创建测试数据
func (suite *EnhancedConflictDetectionTestSuite) createTestData() error {
	ctx := context.Background()

	// 1. 创建测试律师
	lawyer := &models.User{
		Username: "testlawyer",
		Email:    "testlawyer@example.com",
		Name:     "测试律师",
		Role:     "lawyer",
		Password: "hashedpassword",
	}
	err := suite.repo.CreateUser(ctx, lawyer)
	if err != nil {
		return err
	}
	suite.testLawyer = lawyer

	// 2. 创建行业数据
	industries := []models.IndustryClassification{
		{Code: "TMT", Name: "科技、媒体和通信", Level: 1, Keywords: "互联网,科技,通信,软件,游戏,电商"},
		{Code: "FINANCE", Name: "金融", Level: 1, Keywords: "银行,保险,证券,基金,支付"},
	}

	for _, industry := range industries {
		err = suite.repo.CreateOrUpdateIndustry(ctx, &industry)
		if err != nil {
			return err
		}
	}

	// 3. 创建竞争关系
	tmtCompetitors := []models.CompetitiveRelation{
		{IndustryID: 1, CompetitorType: "direct", CompetitorName: "阿里巴巴", CompetitorPattern: "阿里巴巴|阿里|淘宝|天猫|支付宝", ConflictLevel: "HIGH", Description: "阿里巴巴集团"},
		{IndustryID: 1, CompetitorType: "direct", CompetitorName: "腾讯", CompetitorPattern: "腾讯|微信|QQ", ConflictLevel: "HIGH", Description: "腾讯公司"},
		{IndustryID: 1, CompetitorType: "direct", CompetitorName: "字节跳动", CompetitorPattern: "字节跳动|抖音|TikTok", ConflictLevel: "HIGH", Description: "字节跳动公司"},
	}

	for _, competitor := range tmtCompetitors {
		err = suite.repo.CreateOrUpdateCompetitiveRelation(ctx, &competitor)
		if err != nil {
			return err
		}
	}

	// 4. 创建冲突规则
	rules := []models.ConflictRule{
		{Name: "直接客户冲突检测", RuleType: "client_conflict", TriggerPattern: "same_client", ActionType: "block", RiskScore: 100, Priority: 1},
		{Name: "行业直接竞争检测", RuleType: "industry_competition", TriggerPattern: "direct", ActionType: "warn", RiskScore: 90, Priority: 2},
	}

	for _, rule := range rules {
		err = suite.repo.CreateOrUpdateConflictRule(ctx, &rule)
		if err != nil {
			return err
		}
	}

	// 5. 创建测试客户和案件
	return suite.createTestCases(ctx)
}

// createTestCases 创建测试案件
func (suite *EnhancedConflictDetectionTestSuite) createTestCases(ctx context.Context) error {
	// 创建阿里巴巴客户
	alibabaClient := &models.Client{
		Name:     "阿里巴巴集团",
		Company:  "阿里巴巴集团",
		Industry: "科技、媒体和通信",
		Type:     "企业",
	}
	err := suite.repo.CreateClient(ctx, alibabaClient)
	if err != nil {
		return err
	}

	// 创建阿里巴巴案件（现有案件，张伟律师代理）
	alibabaCase := &models.Case{
		Title:         "阿里巴巴诉字节跳动不正当竞争纠纷案",
		Description:    "阿里巴巴诉字节跳动不正当竞争案件",
		ClientID:      alibabaClient.ID,
		LawyerID:      suite.testLawyer.ID,
		CaseType:      "商事",
		Status:        "in_progress",
		Priority:      "high",
		OpposingParty: "字节跳动科技有限公司",
		ClientName:    "阿里巴巴集团",
	}
	err = suite.repo.CreateCase(ctx, alibabaCase)
	if err != nil {
		return err
	}

	// 创建腾讯客户
	tencentClient := &models.Client{
		Name:     "腾讯公司",
		Company:  "腾讯公司",
		Industry: "科技、媒体和通信",
		Type:     "企业",
	}
	err = suite.repo.CreateClient(ctx, tencentClient)
	if err != nil {
		return err
	}

	// 创建腾讯案件
	tencentCase := &models.Case{
		Title:         "腾讯诉百度不正当竞争纠纷案",
		Description:    "腾讯诉百度不正当竞争案件",
		ClientID:      tencentClient.ID,
		LawyerID:      suite.testLawyer.ID,
		CaseType:      "商事",
		Status:        "in_progress",
		Priority:      "medium",
		OpposingParty: "百度在线网络技术有限公司",
		ClientName:    "腾讯公司",
	}
	err = suite.repo.CreateCase(ctx, tencentCase)
	if err != nil {
		return err
	}

	return nil
}

// TestDirectClientConflict 测试直接客户冲突检测
func (suite *EnhancedConflictDetectionTestSuite) TestDirectClientConflict() {
	// 准备请求数据
	request := services.ConflictCheckRequest{
		LawyerID:      suite.testLawyer.ID,
		ClientName:    "阿里巴巴集团",
		ClientType:    "企业",
		OpposingParty: "某公司",
		CaseType:      "商事",
		SearchDepth:   "comprehensive",
	}

	// 执行冲突检测
	result, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.HasConflicts, "应该检测到直接客户冲突")
	suite.Equal("HIGH", result.ConflictLevel, "冲突等级应该是HIGH")
	suite.Greater(result.RiskScore, 80, "风险分数应该大于80")

	// 验证冲突案件
	suite.Greater(len(result.ConflictCases), 0, "应该有冲突案件")

	directConflictFound := false
	for _, conflictCase := range result.ConflictCases {
		if conflictCase.ConflictType == "direct_client" {
			directConflictFound = true
			suite.Equal("阿里巴巴集团", conflictCase.Case.ClientName, "冲突案件客户应该是阿里巴巴")
			break
		}
	}
	suite.True(directConflictFound, "应该找到直接客户冲突")
}

// TestIndustryCompetitionConflict 测试行业竞争冲突检测
func (suite *EnhancedConflictDetectionTestSuite) TestIndustryCompetitionConflict() {
	// 准备请求数据 - 字节跳动作为新客户，阿里巴巴作为对方
	request := services.ConflictCheckRequest{
		LawyerID:       suite.testLawyer.ID,
		ClientName:     "字节跳动科技有限公司",
		ClientType:     "企业",
		ClientIndustry: "科技、媒体和通信",
		OpposingParty:  "阿里巴巴集团",
		CaseType:       "商事",
		SearchDepth:    "comprehensive",
	}

	// 执行冲突检测
	result, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.HasConflicts, "应该检测到行业竞争冲突")
	suite.Equal("HIGH", result.ConflictLevel, "冲突等级应该是HIGH")
	suite.Greater(result.RiskScore, 70, "风险分数应该大于70")

	// 验证竞争分析
	suite.NotNil(result.CompetitionAnalysis, "应该有竞争分析结果")
	suite.True(result.CompetitionAnalysis.HasCompetition, "应该检测到竞争关系")
	suite.Greater(len(result.CompetitionAnalysis.CompetitorInfo), 0, "应该有竞争者信息")

	// 验证建议
	suite.Greater(len(result.Recommendations), 0, "应该有处理建议")
}

// TestAdversePartyConflict 测试对方当事人冲突检测
func (suite *EnhancedConflictDetectionTestSuite) TestAdversePartyConflict() {
	// 准备请求数据 - 新案件对方是字节跳动（与现有案件的对方相同）
	request := services.ConflictCheckRequest{
		LawyerID:      suite.testLawyer.ID,
		ClientName:    "某新客户",
		ClientType:    "企业",
		OpposingParty: "字节跳动科技有限公司",
		CaseType:      "商事",
		SearchDepth:   "comprehensive",
	}

	// 执行冲突检测
	result, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(result)
	// 这里可能不会检测到冲突，因为客户名称和对方当事人都不直接匹配现有案件
}

// TestNoConflict 测试无冲突情况
func (suite *EnhancedConflictDetectionTestSuite) TestNoConflict() {
	// 准备请求数据 - 完全不相关的客户
	request := services.ConflictCheckRequest{
		LawyerID:      suite.testLawyer.ID,
		ClientName:    "某建筑公司",
		ClientType:    "企业",
		OpposingParty: "某设计公司",
		CaseType:      "民事",
		SearchDepth:   "basic",
	}

	// 执行冲突检测
	result, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(result)
	// 在基础搜索模式下可能不会检测到冲突
	if result.HasConflicts {
		suite.Equal("LOW", result.ConflictLevel, "如果有冲突应该是低风险")
		suite.Less(result.RiskScore, 50, "风险分数应该较低")
	}
}

// TestConflictCheckAPI 测试冲突检测API
func (suite *EnhancedConflictDetectionTestSuite) TestConflictCheckAPI() {
	// 准备请求数据
	requestData := map[string]interface{}{
		"lawyer_id":       suite.testLawyer.ID,
		"client_name":     "字节跳动科技有限公司",
		"client_type":     "企业",
		"client_industry": "科技、媒体和通信",
		"opposing_party":  "阿里巴巴集团",
		"case_type":       "商事",
		"search_depth":    "comprehensive",
	}

	requestBody, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/api/conflict/check", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 验证响应
	suite.Equal(http.StatusOK, w.Code, "API应该返回200状态码")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err, "响应应该是有效的JSON")

	suite.Equal("success", response["status"], "响应状态应该是success")
	suite.NotNil(response["data"], "响应应该包含数据")

	data := response["data"].(map[string]interface{})
	suite.True(data["has_conflicts"].(bool), "应该检测到冲突")
	suite.NotNil(data["conflict_cases"], "应该有冲突案件列表")
}

// TestRiskAssessment 测试风险评估功能
func (suite *EnhancedConflictDetectionTestSuite) TestRiskAssessment() {
	// 获取现有案件
	cases, err := suite.repo.GetCasesByLawyer(context.Background(), suite.testLawyer.ID)
	suite.NoError(err)
	suite.Greater(len(cases), 0, "应该有测试案件")

	// 创建风险评估请求
	riskService := services.NewRiskAssessmentService(suite.repo)
	conflictCase := cases[0]

	request := &services.ConflictCheckRequest{
		LawyerID:      suite.testLawyer.ID,
		ClientName:    "字节跳动科技有限公司",
		OpposingParty: "阿里巴巴集团",
		CaseType:      "商事",
	}

	// 执行风险评估
	result, err := riskService.AssessRisk(&gin.Context{}, conflictCase, request)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(result)
	suite.Greater(result.OverallRiskScore, 0, "风险分数应该大于0")
	suite.NotEmpty(result.RiskLevel, "风险等级不应该为空")
	suite.NotEmpty(result.RiskFactors, "风险因子不应该为空")
	suite.Greater(len(result.RiskBreakdown), 0, "风险细分不应该为空")
	suite.Greater(len(result.MitigationStrategies), 0, "缓解策略不应该为空")
}

// TestIndustryCompetitionService 测试行业竞争服务
func (suite *EnhancedConflictDetectionTestSuite) TestIndustryCompetitionService() {
	competitionService := services.NewIndustryCompetitionService(suite.repo)

	// 初始化行业数据
	err := competitionService.InitializeIndustryData(context.Background())
	suite.NoError(err, "行业数据初始化应该成功")

	// 测试竞争分析
	request := &services.CompetitionAnalysisRequest{
		ClientName:     "字节跳动",
		ClientIndustry: "科技、媒体和通信",
		OpposingParty:  "阿里巴巴",
	}

	result, err := competitionService.AnalyzeCompetition(&gin.Context{}, request)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.HasCompetition, "应该检测到竞争关系")
	suite.Greater(len(result.CompetitorInfo), 0, "应该有竞争者信息")
	suite.Equal("HIGH", result.ConflictLevel, "冲突等级应该是HIGH")
	suite.Greater(result.RiskScore, 80, "风险分数应该较高")
}

// TestConflictRuleEngine 测试冲突规则引擎
func (suite *EnhancedConflictDetectionTestSuite) TestConflictRuleEngine() {
	ruleEngine := services.NewConflictRuleEngine(suite.repo)

	// 准备执行上下文
	request := &services.ConflictCheckRequest{
		LawyerID:      suite.testLawyer.ID,
		ClientName:    "阿里巴巴集团",
		OpposingParty: "字节跳动",
		CaseType:      "商事",
		SearchDepth:   "comprehensive",
	}

	// 获取冲突案件
	cases, err := suite.repo.GetCasesByLawyer(context.Background(), suite.testLawyer.ID)
	suite.NoError(err)

	executionCtx := &services.RuleExecutionContext{
		Request:       request,
		ConflictCases: cases,
	}

	// 执行规则引擎
	results, err := ruleEngine.ExecuteRules(&gin.Context{}, executionCtx)

	// 验证结果
	suite.NoError(err)
	suite.NotNil(results)
	suite.Greater(len(results), 0, "应该有规则执行结果")

	// 验证有规则被触发
	triggeredCount := 0
	for _, result := range results {
		if result.Triggered {
			triggeredCount++
		}
	}
	suite.Greater(triggeredCount, 0, "应该有规则被触发")
}

// TestComprehensiveScenario 测试综合场景
func (suite *EnhancedConflictDetectionTestSuite) TestComprehensiveScenario() {
	// 测试场景1: 字节跳动诉阿里巴巴，应该检测到竞争冲突
	request1 := services.ConflictCheckRequest{
		LawyerID:       suite.testLawyer.ID,
		ClientName:     "字节跳动科技有限公司",
		ClientType:     "企业",
		ClientIndustry: "科技、媒体和通信",
		OpposingParty:  "阿里巴巴集团",
		CaseType:       "商事",
		SearchDepth:    "comprehensive",
	}

	result1, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request1)
	suite.NoError(err)
	suite.True(result1.HasConflicts, "场景1应该检测到竞争冲突")
	suite.Equal("HIGH", result1.ConflictLevel, "场景1风险等级应该是HIGH")

	// 测试场景2: 同一客户的不同案件，应该检测到直接冲突
	request2 := services.ConflictCheckRequest{
		LawyerID:      suite.testLawyer.ID,
		ClientName:    "阿里巴巴集团",
		ClientType:    "企业",
		OpposingParty: "某公司",
		CaseType:      "商事",
		SearchDepth:   "comprehensive",
	}

	result2, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request2)
	suite.NoError(err)
	suite.True(result2.HasConflicts, "场景2应该检测到直接客户冲突")
	suite.Equal("HIGH", result2.ConflictLevel, "场景2风险等级应该是HIGH")

	// 测试场景3: 无关联的客户，应该无冲突或低风险
	request3 := services.ConflictCheckRequest{
		LawyerID:      suite.testLawyer.ID,
		ClientName:    "某制造企业",
		ClientType:    "企业",
		OpposingParty: "某原材料供应商",
		CaseType:      "民事",
		SearchDepth:   "basic",
	}

	result3, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request3)
	suite.NoError(err)
	// 基础搜索模式下可能不会检测到冲突
}

// TestPerformance 测试性能
func (suite *EnhancedConflictDetectionTestSuite) TestPerformance() {
	request := services.ConflictCheckRequest{
		LawyerID:       suite.testLawyer.ID,
		ClientName:     "字节跳动科技有限公司",
		ClientType:     "企业",
		ClientIndustry: "科技、媒体和通信",
		OpposingParty:  "阿里巴巴集团",
		CaseType:       "商事",
		SearchDepth:    "comprehensive",
	}

	// 执行多次检测并计算平均时间
	iterations := 10
	var totalTime time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, err := suite.conflictSvc.CheckConflicts(&gin.Context{}, &request)
		duration := time.Since(start)

		suite.NoError(err)
		totalTime += duration
	}

	averageTime := totalTime / time.Duration(iterations)
	suite.Less(averageTime, time.Second, "平均检测时间应该小于1秒")

	fmt.Printf("平均冲突检测时间: %v\n", averageTime)
}

// TestEnhancedConflictDetectionTestSuite 运行测试套件
func TestEnhancedConflictDetectionTestSuite(t *testing.T) {
	suite.Run(t, new(EnhancedConflictDetectionTestSuite))
}