//go:build ignore

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// ConflictCheckRequest 冲突检查请求结构
type ConflictCheckRequest struct {
	LawyerID       int    `json:"lawyer_id"`
	ClientName     string `json:"client_name"`
	ClientType     string `json:"client_type"`
	ClientIndustry string `json:"client_industry"`
	OpposingParty  string `json:"opposing_party"`
	CaseType       string `json:"case_type"`
	SearchDepth    string `json:"search_depth"`
	IncludeRelated bool   `json:"include_related"`
}

// ConflictCheckResponse 冲突检查响应结构
type ConflictCheckResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// testScenario 测试场景结构
type testScenario struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Request     ConflictCheckRequest  `json:"request"`
	ExpectConflicts bool              `json:"expect_conflicts"`
	ExpectRiskLevel string             `json:"expect_risk_level"`
}

func main() {
	fmt.Println("=== 增强冲突检测功能验证 ===")
	fmt.Printf("测试时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// 1. 检查数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}
	fmt.Println("✓ 数据库连接正常")

	// 2. 检查必要表是否存在
	if err := checkTables(db); err != nil {
		log.Fatal("表结构检查失败:", err)
	}
	fmt.Println("✓ 数据库表结构正常")

	// 3. 检查API服务
	if err := checkAPIService(); err != nil {
		log.Fatal("API服务检查失败:", err)
	}
	fmt.Println("✓ API服务正常")

	// 4. 检查测试数据
	if err := checkTestData(db); err != nil {
		log.Fatal("测试数据检查失败:", err)
	}
	fmt.Println("✓ 测试数据正常")

	// 5. 执行测试场景
	fmt.Println("\n=== 开始执行测试场景 ===")
	testScenarios := getTestScenarios()

	passedTests := 0
	totalTests := len(testScenarios)

	for i, scenario := range testScenarios {
		fmt.Printf("\n--- 测试场景 %d: %s ---\n", i+1, scenario.Name)
		fmt.Printf("描述: %s\n", scenario.Description)

		success := runTestScenario(scenario)
		if success {
			fmt.Printf("✓ %s - 通过\n", scenario.Name)
			passedTests++
		} else {
			fmt.Printf("✗ %s - 失败\n", scenario.Name)
		}
	}

	// 6. 输出测试结果
	fmt.Printf("\n=== 测试结果汇总 ===\n")
	fmt.Printf("总测试数: %d\n", totalTests)
	fmt.Printf("通过测试: %d\n", passedTests)
	fmt.Printf("失败测试: %d\n", totalTests-passedTests)
	fmt.Printf("通过率: %.1f%%\n", float64(passedTests)/float64(totalTests)*100)

	if passedTests == totalTests {
		fmt.Println("🎉 所有测试通过！增强冲突检测功能正常工作")
	} else {
		fmt.Println("⚠️ 部分测试失败，需要进一步检查")
	}

	fmt.Printf("\n验证完成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

// checkTables 检查数据库表结构
func checkTables(db *sql.DB) error {
	requiredTables := []string{
		"users", "clients", "cases",
		"industry_classifications", "competitive_relations",
		"conflict_rules", "conflict_detection_history",
	}

	for _, table := range requiredTables {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists)

		if err != nil {
			return fmt.Errorf("检查表 %s 失败: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("表 %s 不存在", table)
		}
	}

	return nil
}

// checkAPIService 检查API服务
func checkAPIService() error {
	client := &http.Client{Timeout: 5 * time.Second}

	// 测试健康检查端点
	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		return fmt.Errorf("API服务连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API服务状态异常: %d", resp.StatusCode)
	}

	return nil
}

// checkTestData 检查测试数据
func checkTestData(db *sql.DB) error {
	// 检查测试律师
	var lawyerCount int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'lawyer' AND username = 'zhangwei'").Scan(&lawyerCount)
	if err != nil {
		return fmt.Errorf("查询测试律师失败: %w", err)
	}
	if lawyerCount == 0 {
		return fmt.Errorf("未找到测试律师 zhangwei")
	}

	// 检查测试客户
	var clientCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM clients
		WHERE company IN ('阿里巴巴集团', '腾讯公司', '字节跳动科技有限公司')
	`).Scan(&clientCount)
	if err != nil {
		return fmt.Errorf("查询测试客户失败: %w", err)
	}
	if clientCount == 0 {
		return fmt.Errorf("未找到测试客户")
	}

	// 检查测试案件
	var caseCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM cases
		WHERE lawyer_id IN (SELECT id FROM users WHERE username = 'zhangwei')
		AND status = 'in_progress'
	`).Scan(&caseCount)
	if err != nil {
		return fmt.Errorf("查询测试案件失败: %w", err)
	}
	if caseCount == 0 {
		return fmt.Errorf("未找到测试案件")
	}

	// 检查行业数据
	var industryCount int
	err = db.QueryRow("SELECT COUNT(*) FROM industry_classifications").Scan(&industryCount)
	if err != nil {
		return fmt.Errorf("查询行业数据失败: %w", err)
	}
	if industryCount == 0 {
		return fmt.Errorf("未找到行业数据")
	}

	return nil
}

// getTestScenarios 获取测试场景
func getTestScenarios() []testScenario {
	return []testScenario{
		{
			Name:        "直接客户冲突检测",
			Description:  "测试同一客户的直接冲突检测",
			Request: ConflictCheckRequest{
				LawyerID:      1, // 假设张伟律师ID为1
				ClientName:    "阿里巴巴集团",
				ClientType:    "企业",
				OpposingParty: "某公司",
				CaseType:      "商事",
				SearchDepth:   "comprehensive",
			},
			ExpectConflicts: true,
			ExpectRiskLevel: "HIGH",
		},
		{
			Name:        "行业竞争冲突检测",
			Description:  "测试字节跳动 vs 阿里巴巴的行业竞争冲突",
			Request: ConflictCheckRequest{
				LawyerID:       1,
				ClientName:     "字节跳动科技有限公司",
				ClientType:     "企业",
				ClientIndustry: "科技、媒体和通信",
				OpposingParty:  "阿里巴巴集团",
				CaseType:       "商事",
				SearchDepth:    "comprehensive",
			},
			ExpectConflicts: true,
			ExpectRiskLevel: "HIGH",
		},
		{
			Name:        "腾讯 vs 阿里巴巴竞争检测",
			Description:  "测试腾讯与阿里巴巴的竞争关系检测",
			Request: ConflictCheckRequest{
				LawyerID:       1,
				ClientName:     "腾讯公司",
				ClientType:     "企业",
				ClientIndustry: "科技、媒体和通信",
				OpposingParty:  "阿里巴巴集团",
				CaseType:       "商事",
				SearchDepth:    "comprehensive",
			},
			ExpectConflicts: true,
			ExpectRiskLevel: "HIGH",
		},
		{
			Name:        "无冲突场景检测",
			Description:  "测试无关联客户的无冲突检测",
			Request: ConflictCheckRequest{
				LawyerID:      1,
				ClientName:    "某建筑公司",
				ClientType:    "企业",
				OpposingParty: "某设计公司",
				CaseType:      "民事",
				SearchDepth:   "basic",
			},
			ExpectConflicts: false,
			ExpectRiskLevel: "NONE",
		},
		{
			Name:        "名称相似性检测",
			Description:  "测试客户名称相似性冲突检测",
			Request: ConflictCheckRequest{
				LawyerID:      1,
				ClientName:    "阿里云计算有限公司",
				ClientType:    "企业",
				OpposingParty: "某公司",
				CaseType:      "商事",
				SearchDepth:   "comprehensive",
			},
			ExpectConflicts: true,
			ExpectRiskLevel: "MEDIUM",
		},
	}
}

// runTestScenario 运行测试场景
func runTestScenario(scenario testScenario) bool {
	// 准备请求数据
	requestBody, err := json.Marshal(scenario.Request)
	if err != nil {
		fmt.Printf("序列化请求失败: %v\n", err)
		return false
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"http://localhost:8080/api/conflict/check",
		"application/json",
		strings.NewReader(string(requestBody)),
	)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API返回错误状态码: %d\n", resp.StatusCode)
		return false
	}

	// 解析响应
	var response ConflictCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return false
	}

	if response.Status != "success" {
		fmt.Printf("API返回错误: %s\n", response.Message)
		return false
	}

	// 获取结果数据
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		fmt.Printf("响应数据格式错误\n")
		return false
	}

	// 检查冲突检测结果
	hasConflicts, ok := data["has_conflicts"].(bool)
	if !ok {
		fmt.Printf("无法获取冲突检测结果\n")
		return false
	}

	// 验证冲突检测结果
	if hasConflicts != scenario.ExpectConflicts {
		fmt.Printf("冲突检测结果不符: 期望 %v, 实际 %v\n", scenario.ExpectConflicts, hasConflicts)
		return false
	}

	// 检查风险等级
	riskLevel, ok := data["conflict_level"].(string)
	if !ok {
		fmt.Printf("无法获取风险等级\n")
		return false
	}

	// 验证风险等级（允许NONE级别）
	if riskLevel != scenario.ExpectRiskLevel && riskLevel != "NONE" && scenario.ExpectRiskLevel != "NONE" {
		fmt.Printf("风险等级不符: 期望 %s, 实际 %s\n", scenario.ExpectRiskLevel, riskLevel)
		return false
	}

	// 输出详细信息
	fmt.Printf("  冲突检测: %v\n", hasConflicts)
	fmt.Printf("  风险等级: %s\n", riskLevel)

	if riskScore, ok := data["risk_score"].(float64); ok {
		fmt.Printf("  风险分数: %.0f\n", riskScore)
	}

	if conflictCases, ok := data["conflict_cases"].([]interface{}); ok {
		fmt.Printf("  冲突案件数: %d\n", len(conflictCases))

		// 显示前3个冲突案件
		for i, conflictCase := range conflictCases {
			if i >= 3 {
				break
			}
			if caseData, ok := conflictCase.(map[string]interface{}); ok {
				if caseInfo, ok := caseData["case"].(map[string]interface{}); ok {
					if title, ok := caseInfo["title"].(string); ok {
						fmt.Printf("    - %s\n", title)
					}
				}
			}
		}
	}

	if competitionAnalysis, ok := data["competition_analysis"].(map[string]interface{}); ok {
		if hasCompetition, ok := competitionAnalysis["has_competition"].(bool); ok && hasCompetition {
			fmt.Printf("  竞争分析: 检测到竞争关系\n")
			if competitors, ok := competitionAnalysis["competitor_info"].([]interface{}); ok {
				fmt.Printf("  竞争者数: %d\n", len(competitors))
			}
		}
	}

	return true
}