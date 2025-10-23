package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TestConflictRequest 冲突检测测试请求
type TestConflictRequest struct {
	LawyerID       int       `json:"lawyer_id"`
	ClientName     string    `json:"client_name"`
	ClientType     string    `json:"client_type"`
	ClientIndustry string    `json:"client_industry"`
	OpposingParty  string    `json:"opposing_party"`
	CaseType       string    `json:"case_type"`
	SearchDepth    string    `json:"search_depth"`
	IncludeRelated bool      `json:"include_related"`
	SearchYears    int       `json:"search_years"`
	TeamMembers    []string  `json:"team_members"`
	Description    string    `json:"description"`
	UserID         string    `json:"user_id"`
	RequestTime    time.Time `json:"request_time"`
}

// 测试字节跳动 vs 腾讯冲突检测场景
func testByteDanceVsTencent() {
	baseURL := "http://localhost:8080/api/conflict/check"

	// 场景1: 字节跳动 vs 腾讯 - 商业竞争冲突
	scenario1 := TestConflictRequest{
		LawyerID:       1,
		ClientName:     "字节跳动科技有限公司",
		ClientType:     "COMPANY",
		ClientIndustry: "科技、媒体和通信",
		OpposingParty:  "深圳市腾讯计算机系统有限公司",
		CaseType:       "commercial",
		SearchDepth:    "comprehensive",
		IncludeRelated: true,
		SearchYears:    5,
		TeamMembers:    []string{"张三", "李四"},
		Description:    "商业合同纠纷案件",
		UserID:         "test_user_001",
		RequestTime:    time.Now(),
	}

	fmt.Println("🔍 测试场景1: 字节跳动 vs 腾讯商业竞争冲突")
	testScenario(baseURL, scenario1)

	// 场景2: 字节跳动 vs 腾讯 - 知识产权冲突
	scenario2 := TestConflictRequest{
		LawyerID:       1,
		ClientName:     "北京字节跳动科技有限公司",
		ClientType:     "COMPANY",
		ClientIndustry: "科技、媒体和通信",
		OpposingParty:  "腾讯科技（深圳）有限公司",
		CaseType:       "commercial",
		SearchDepth:    "comprehensive",
		IncludeRelated: true,
		SearchYears:    3,
		TeamMembers:    []string{"王五", "赵六"},
		Description:    "知识产权侵权纠纷",
		UserID:         "test_user_002",
		RequestTime:    time.Now(),
	}

	fmt.Println("\n🔍 测试场景2: 字节跳动 vs 腾讯知识产权冲突")
	testScenario(baseURL, scenario2)

	// 场景3: 字节跳动内部部门冲突检测
	scenario3 := TestConflictRequest{
		LawyerID:       2,
		ClientName:     "字节跳动抖音事业部",
		ClientType:     "COMPANY",
		ClientIndustry: "科技、媒体和通信",
		OpposingParty:  "字节跳动今日头条事业部",
		CaseType:       "commercial",
		SearchDepth:    "comprehensive",
		IncludeRelated: true,
		SearchYears:    2,
		TeamMembers:    []string{"陈七", "刘八"},
		Description:    "内部业务合作纠纷",
		UserID:         "test_user_003",
		RequestTime:    time.Now(),
	}

	fmt.Println("\n🔍 测试场景3: 字节跳动内部部门冲突")
	testScenario(baseURL, scenario3)
}

func testScenario(baseURL string, request TestConflictRequest) {
	// 序列化请求数据
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ 序列化请求失败: %v\n", err)
		return
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 如果需要认证，添加Authorization头
	req.Header.Set("Authorization", "Bearer test_token")

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 发送请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("响应内容: %s\n", string(body))
		return
	}

	// 显示结果
	fmt.Printf("✅ 测试完成 - 状态码: %d\n", resp.StatusCode)
	fmt.Printf("客户: %s vs 对方: %s\n", request.ClientName, request.OpposingParty)

	if data, ok := result["data"].(map[string]interface{}); ok {
		if hasConflicts, ok := data["has_conflicts"].(bool); ok {
			fmt.Printf("冲突检测结果: %v\n", hasConflicts)
		}
		if conflictLevel, ok := data["conflict_level"].(string); ok {
			fmt.Printf("风险等级: %s\n", conflictLevel)
		}
		if riskScore, ok := data["risk_score"].(float64); ok {
			fmt.Printf("风险分数: %.0f\n", riskScore)
		}
		if conflictCases, ok := data["conflict_cases"].([]interface{}); ok {
			fmt.Printf("冲突案件数量: %d\n", len(conflictCases))
		}

		// 显示竞争分析结果
		if competitionAnalysis, ok := data["competition_analysis"].(map[string]interface{}); ok {
			if hasCompetition, ok := competitionAnalysis["has_competition"].(bool); ok && hasCompetition {
				fmt.Printf("🏭 竞争分析: 检测到行业竞争关系\n")
				if competitorInfo, ok := competitionAnalysis["competitor_info"].([]interface{}); ok {
					fmt.Printf("   竞争者数量: %d\n", len(competitorInfo))
				}
				if riskFactors, ok := competitionAnalysis["risk_factors"].([]interface{}); ok {
					fmt.Printf("   风险因素: %v\n", riskFactors)
				}
			} else {
				fmt.Printf("🏭 竞争分析: 未检测到明显竞争关系\n")
			}
		}

		// 显示建议
		if recommendations, ok := data["recommendations"].([]interface{}); ok {
			fmt.Printf("📋 建议数量: %d\n", len(recommendations))
			for i, rec := range recommendations {
				if i < 3 { // 只显示前3个建议
					fmt.Printf("   %d. %v\n", i+1, rec)
				}
			}
		}
	}

	fmt.Printf("请求耗时: %v\n\n", time.Since(request.RequestTime))
}

func main() {
	fmt.Println("🚀 开始字节跳动 vs 腾讯冲突检测测试")
	fmt.Println("测试时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 60))

	testByteDanceVsTencent()

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ 所有测试场景完成")
}