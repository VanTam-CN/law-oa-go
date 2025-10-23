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

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool   `json:"success"`
	Data    struct {
		Token string `json:"token"`
		User  struct {
			ID       int    `json:"id"`
			Username string `json:"username"`
			Name     string `json:"name"`
			Role     string `json:"role"`
		} `json:"user"`
	} `json:"data"`
	Error string `json:"error"`
}

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

func main() {
	fmt.Println("🚀 开始字节跳动 vs 腾讯冲突检测测试（带认证）")
	fmt.Println("测试时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 60))

	// 1. 登录获取令牌
	token, err := loginAndGetToken("zhangwei@jinchenglaw.com", "law123456")
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 登录成功，获取到令牌\n\n")

	// 2. 测试字节跳动vs腾讯冲突场景
	testBytedanceVsTencent(token)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ 所有测试场景完成")
}

// loginAndGetToken 登录并获取认证令牌
func loginAndGetToken(email, password string) (string, error) {
	loginURL := "http://localhost:8080/api/v1/auth/login"

	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("序列化登录请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建登录请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送登录请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取登录响应失败: %v", err)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %v", err)
	}

	if !loginResp.Success || loginResp.Data.Token == "" {
		return "", fmt.Errorf("登录失败: %s", loginResp.Error)
	}

	return loginResp.Data.Token, nil
}

// testBytedanceVsTencent 测试字节跳动vs腾讯冲突检测场景
func testBytedanceVsTencent(token string) {
	baseURL := "http://localhost:8080/api/v1/conflict/check"

	// 场景1: 张伟律师为字节跳动代理案件，对手为腾讯 - 商业竞争冲突
	scenario1 := TestConflictRequest{
		LawyerID:       6, // 张伟律师的ID
		ClientName:     "字节跳动科技有限公司",
		ClientType:     "企业",
		ClientIndustry: "互联网科技",
		OpposingParty:  "深圳市腾讯计算机系统有限公司",
		CaseType:       "commercial",
		SearchDepth:    "comprehensive",
		IncludeRelated: true,
		SearchYears:    5,
		TeamMembers:    []string{"助理1", "助理2"},
		Description:    "商业合同纠纷案件",
		UserID:         "zhangwei",
		RequestTime:    time.Now(),
	}

	fmt.Println("🔍 测试场景1: 张伟律师为字节跳动代理案件，对手为腾讯")
	testScenarioWithAuth(baseURL, scenario1, token)

	// 场景2: 字节跳动 vs 腾讯 - 知识产权冲突
	scenario2 := TestConflictRequest{
		LawyerID:       6,
		ClientName:     "北京字节跳动科技有限公司",
		ClientType:     "企业",
		ClientIndustry: "互联网科技",
		OpposingParty:  "腾讯科技（深圳）有限公司",
		CaseType:       "commercial",
		SearchDepth:    "comprehensive",
		IncludeRelated: true,
		SearchYears:    3,
		TeamMembers:    []string{"法务1", "法务2"},
		Description:    "知识产权侵权纠纷",
		UserID:         "zhangwei",
		RequestTime:    time.Now(),
	}

	fmt.Println("\n🔍 测试场景2: 字节跳动 vs 腾讯知识产权冲突")
	testScenarioWithAuth(baseURL, scenario2, token)

	// 场景3: 测试无冲突场景 - 为新客户代理案件
	scenario3 := TestConflictRequest{
		LawyerID:       6,
		ClientName:     "测试新客户科技有限公司",
		ClientType:     "企业",
		ClientIndustry: "制造业",
		OpposingParty:  "某某制造公司",
		CaseType:       "commercial",
		SearchDepth:    "comprehensive",
		IncludeRelated: true,
		SearchYears:    2,
		TeamMembers:    []string{"助理3"},
		Description:    "普通合同纠纷",
		UserID:         "zhangwei",
		RequestTime:    time.Now(),
	}

	fmt.Println("\n🔍 测试场景3: 无冲突场景 - 新客户代理")
	testScenarioWithAuth(baseURL, scenario3, token)
}

func testScenarioWithAuth(baseURL string, request TestConflictRequest, token string) {
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
	req.Header.Set("Authorization", "Bearer "+token)

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

	if resp.StatusCode == 200 {
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
				// 显示冲突案件详情
				for i, conflictCase := range conflictCases {
					if caseData, ok := conflictCase.(map[string]interface{}); ok {
						if title, ok := caseData["title"].(string); ok {
							fmt.Printf("   冲突案件 %d: %s\n", i+1, title)
						}
					}
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
	} else {
		fmt.Printf("❌ 请求失败，错误信息: %v\n", result)
	}

	fmt.Printf("请求耗时: %v\n\n", time.Since(request.RequestTime))
}