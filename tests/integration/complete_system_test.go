package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CompleteSystemTestSuite 完整系统测试套件
type CompleteSystemTestSuite struct {
	suite.Suite
	baseURL    string
	authToken  string
	testClient *http.Client
}

// SetupSuite 测试套件初始化
func (suite *CompleteSystemTestSuite) SetupSuite() {
	suite.baseURL = getBaseURL()
	suite.testClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// 获取认证令牌
	suite.authToken = getAuthToken()

	if suite.baseURL == "" {
		suite.T().Skip("服务器未运行，跳过集成测试")
		return
	}

	fmt.Printf("开始完整系统测试 - 服务器: %s\n", suite.baseURL)
}

// TearDownSuite 测试套件清理
func (suite *CompleteSystemTestSuite) TearDownSuite() {
	fmt.Println("完整系统测试完成")
}

// TestSystemHealth 系统健康检查
func (suite *CompleteSystemTestSuite) TestSystemHealth() {
	fmt.Println("\n=== 系统健康检查 ===")

	// 测试健康检查端点
	resp, err := suite.testClient.Get(suite.baseURL + "/conflict/health")
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)

	var health map[string]interface{}
	err = json.Unmarshal(body, &health)
	suite.NoError(err)

	fmt.Printf("✅ 系统健康状态: %v\n", health["data"])
}

// TestCaseCreationWorkflow 案件创建工作流测试
func (suite *CompleteSystemTestSuite) TestCaseCreationWorkflow() {
	fmt.Println("\n=== 案件创建工作流测试 ===")

	// 1. 获取律师列表（用于案件创建）
	lawyersResp, err := suite.testClient.Get(suite.baseURL + "/lawfirm/lawyers")
	suite.NoError(err)
	suite.Equal(http.StatusOK, lawyersResp.StatusCode)

	// 2. 获取客户列表
	clientsResp, err := suite.testClient.Get(suite.baseURL + "/clients")
	suite.NoError(err)
	suite.Equal(http.StatusOK, clientsResp.StatusCode)

	// 3. 创建测试案件
	caseData := map[string]interface{}{
		"title":       "集成测试案件",
		"description": "这是一个集成测试案件",
		"client_id":   1,
		"lawyer_id":   1,
		"case_type":   "civil",
		"priority":    "medium",
		"status":      "pending",
	}

	jsonData, _ := json.Marshal(caseData)
	resp, err := suite.testClient.Post(
		suite.baseURL + "/cases",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	suite.NoError(err)

	var createResp map[string]interface{}
	err = json.Unmarshal(body, &createResp)
	suite.NoError(err)
	suite.True(createResp["success"].(bool))

	caseID := createResp["data"].(map[string]interface{})["id"]
	fmt.Printf("✅ 案件创建成功，ID: %v\n", caseID)

	// 4. 验证案件详情获取
	resp, err = suite.testClient.Get(fmt.Sprintf("%s/cases/%v", suite.baseURL, caseID))
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	fmt.Printf("✅ 案件详情获取成功\n")
}

// TestCaseFilteringWorkflow 案件筛选工作流测试
func (suite *CompleteSystemTestSuite) TestCaseFilteringWorkflow() {
	fmt.Println("\n=== 案件筛选工作流测试 ===")

	// 测试不同的筛选条件
	testCases := []struct {
		name string
		url  string
		desc string
	}{
		{
			name: "按状态筛选",
			url:  "/cases?status=pending",
			desc: "筛选待处理案件",
		},
		{
			name: "按类型筛选",
			url:  "/cases?case_type=civil",
			desc: "筛选民事案件",
		},
		{
			name: "按优先级筛选",
			url:  "/cases?priority=medium",
			desc: "筛选中等优先级案件",
		},
		{
			name: "组合筛选",
			url:  "/cases?status=pending&case_type=civil&priority=medium",
			desc: "组合筛选条件",
		},
		{
			name: "搜索功能",
			url:  "/cases?search=测试",
			desc: "搜索包含'测试'的案件",
		},
	}

	for _, tc := range testCases {
		resp, err := suite.testClient.Get(suite.baseURL + tc.url)
		suite.NoError(err, "请求失败: %s - %v", tc.name, err)
		suite.Equal(http.StatusOK, resp.StatusCode, "状态码不正确: %s", tc.name)

		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)

		var listResp map[string]interface{}
		err = json.Unmarshal(body, &listResp)
		suite.NoError(err)

		suite.True(listResp["success"].(bool), "响应失败: %s", tc.name)
		fmt.Printf("✅ %s - 成功\n", tc.desc)
	}
}

// TestConflictDetectionWorkflow 冲突检测工作流测试
func (suite *CompleteSystemTestSuite) TestConflictDetectionWorkflow() {
	fmt.Println("\n=== 冲突检测工作流测试 ===")

	// 1. 健康检查
	resp, err := suite.testClient.Get(suite.baseURL + "/conflict/health")
	suite.NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	// 2. 获取MCP标准
	resp, err = suite.testClient.Get(suite.baseURL + "/conflict/standards")
	suite.NoError(err)
	fmt.Printf("✅ MCP标准获取: %d\n", resp.StatusCode)

	// 3. 获取冲突规则
	resp, err = suite.testClient.Get(suite.baseURL + "/conflict/rules")
	suite.NoError(err)
	fmt.Printf("✅ 冲突规则获取: %d\n", resp.StatusCode)

	// 4. 执行冲突检测
	conflictData := map[string]interface{}{
		"clientId":                 "1",
		"clientName":               "测试客户公司",
		"caseName":                 "商业合同纠纷案",
		"caseType":                 "commercial",
		"clientType":               "COMPANY",
		"otherParties":             []string{"对手公司"},
		"searchYears":              5,
		"includeCorporateRelations": true,
		"searchDepth":              "deep",
		"userId":                   "1",
		"requestTime":              time.Now(),
	}

	jsonData, _ := json.Marshal(conflictData)
	resp, err = suite.testClient.Post(
		suite.baseURL + "/conflict/check",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	suite.NoError(err)

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)

		var conflictResp map[string]interface{}
		err = json.Unmarshal(body, &conflictResp)
		suite.NoError(err)

		if conflictResp["success"].(bool) {
			fmt.Printf("✅ 冲突检测执行成功\n")
			if data, ok := conflictResp["data"].(map[string]interface{}); ok {
				hasConflict := data["has_conflict"].(bool)
				if hasConflict {
					fmt.Printf("⚠️  发现冲突风险\n")
				} else {
					fmt.Printf("✅ 未发现冲突\n")
				}
			}
		} else {
			fmt.Printf("⚠️  冲突检测返回成功状态但执行失败\n")
		}
	} else {
		fmt.Printf("⚠️  冲突检测失败，状态码: %d\n", resp.StatusCode)
	}
}

// TestErrorHandling 错误处理测试
func (suite *CompleteSystemTestSuite) TestErrorHandling() {
	fmt.Println("\n=== 错误处理测试 ===")

	testCases := []struct {
		name     string
		method   string
		url      string
		body     interface{}
		expected int
		desc     string
	}{
		{
			name:     "无效的案件创建参数",
			method:   "POST",
			url:      "/cases",
			body:     map[string]interface{}{},
			expected: http.StatusBadRequest,
			desc:     "测试参数验证错误处理",
		},
		{
			name:     "不存在的资源",
			method:   "GET",
			url:      "/cases/99999",
			expected: http.StatusNotFound,
			desc:     "测试资源不存在错误处理",
		},
		{
			name:     "无效的筛选参数",
			method:   "GET",
			url:      "/cases?status=invalid_status",
			expected: http.StatusBadRequest,
			desc:     "测试筛选参数验证错误处理",
		},
	}

	for _, tc := range testCases {
		var resp *http.Response
		var err error

		if tc.method == "GET" {
			resp, err = suite.testClient.Get(suite.baseURL + tc.url)
		} else if tc.method == "POST" {
			jsonData, _ := json.Marshal(tc.body)
			resp, err = suite.testClient.Post(
				suite.baseURL + tc.url,
				"application/json",
				bytes.NewBuffer(jsonData),
			)
		}

		suite.NoError(err, "请求失败: %s - %v", tc.name, err)
		suite.Equal(tc.expected, resp.StatusCode, "状态码不正确: %s - 期望 %d, 实际 %d", tc.name, tc.expected, resp.StatusCode)

		// 验证错误响应格式
		body, err := io.ReadAll(resp.Body)
		suite.NoError(err)

		var errorResp map[string]interface{}
		err = json.Unmarshal(body, &errorResp)
		suite.NoError(err)

		// 检查响应格式是否符合统一标准
		suite.Contains(errorResp, "success")
		suite.Contains(errorResp, "timestamp")
		suite.Contains(errorResp, "meta")

		fmt.Printf("✅ %s - 错误处理正确\n", tc.desc)
	}
}

// TestPerformance 性能测试
func (suite *CompleteSystemTestSuite) TestPerformance() {
	fmt.Println("\n=== 性能测试 ===")

	// 测试案件列表加载性能
	start := time.Now()
	resp, err := suite.testClient.Get(suite.baseURL + "/cases?page=1&page_size=20")
	suite.NoError(err)
	duration := time.Since(start)

	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Less(duration, 2*time.Second, "案件列表加载时间过长")
	fmt.Printf("✅ 案件列表加载时间: %v\n", duration)

	// 测试搜索性能
	start = time.Now()
	resp, err = suite.testClient.Get(suite.baseURL + "/cases?search=test")
	suite.NoError(err)
	duration = time.Since(start)

	suite.Equal(http.StatusOK, resp.StatusCode)
	suite.Less(duration, 3*time.Second, "搜索响应时间过长")
	fmt.Printf("✅ 搜索响应时间: %v\n", duration)
}

// TestConcurrentRequests 并发请求测试
func (suite *CompleteSystemTestSuite) TestConcurrentRequests() {
	fmt.Println("\n=== 并发请求测试 ===")

	const numRequests = 10
	results := make(chan bool, numRequests)

	start := time.Now()

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := suite.testClient.Get(suite.baseURL + "/cases")
			if err != nil {
				results <- false
				return
			}
			results <- resp.StatusCode == http.StatusOK
		}()
	}

	successCount := 0
	for i := 0; i < numRequests; i++ {
		if <-results {
			successCount++
		}
	}

	duration := time.Since(start)

	suite.GreaterOrEqual(successCount, numRequests*80/100, "成功率过低")
	suite.Less(duration, 10*time.Second, "并发请求响应时间过长")

	fmt.Printf("✅ 并发测试完成 - 成功率: %d/%d, 总耗时: %v\n", successCount, numRequests, duration)
}

// 辅助函数
func getBaseURL() string {
	if url := os.Getenv("TEST_SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func getAuthToken() string {
	if token := os.Getenv("TEST_AUTH_TOKEN"); token != "" {
		return token
	}
	// 在实际测试中，这里应该实现登录逻辑获取token
	return "test-token"
}

// TestCompleteSystemSuite 运行完整测试套件
func TestCompleteSystemSuite(t *testing.T) {
	suite.Run(t, new(CompleteSystemTestSuite))
}

// 单独运行各个测试
func TestSystemHealth(t *testing.T) {
	suite := new(CompleteSystemTestSuite)
	suite.SetupSuite()
	defer suite.TearDownSuite()
	suite.TestSystemHealth()
}

func TestCaseCreationWorkflow(t *testing.T) {
	suite := new(CompleteSystemTestSuite)
	suite.SetupSuite()
	defer suite.TearDownSuite()
	suite.TestCaseCreationWorkflow()
}

func TestCaseFilteringWorkflow(t *testing.T) {
	suite := new(CompleteSystemTestSuite)
	suite.SetupSuite()
	defer suite.TearDownSuite()
	suite.TestCaseFilteringWorkflow()
}

func TestConflictDetectionWorkflow(t *testing.T) {
	suite := new(CompleteSystemTestSuite)
	suite.SetupSuite()
	defer suite.TearDownSuite()
	suite.TestConflictDetectionWorkflow()
}

func TestErrorHandling(t *testing.T) {
	suite := new(CompleteSystemTestSuite)
	suite.SetupSuite()
	defer suite.TearDownSuite()
	suite.TestErrorHandling()
}

func TestPerformance(t *testing.T) {
	suite := new(CompleteSystemTestSuite)
	suite.SetupSuite()
	defer suite.TearDownSuite()
	suite.TestPerformance()
}

func TestConcurrentRequests(t *testing.T) {
	suite := new(CompleteSystemTestSuite)
	suite.SetupSuite()
	defer suite.TearDownSuite()
	suite.TestConcurrentRequests()
}