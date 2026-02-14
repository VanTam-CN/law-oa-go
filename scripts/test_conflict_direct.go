//go:build ignore

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
	ClientID                  string    `json:"clientId" binding:"required"`
	ClientName                string    `json:"clientName" binding:"required"`
	CaseName                  string    `json:"caseName" binding:"required"`
	CaseType                  string    `json:"caseType" binding:"required"`
	ClientType                string    `json:"clientType" binding:"required"`
	OtherParties              []string  `json:"otherParties"`
	SearchYears               int       `json:"searchYears"`
	IncludeCorporateRelations bool      `json:"includeCorporateRelations"`
	SearchDepth               string    `json:"searchDepth"`
	UserID                    string    `json:"userId"`
	RequestTime               time.Time `json:"requestTime"`
}

func main() {
	fmt.Println("🔍 直接测试后端冲突检测API")
	fmt.Println("测试时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("=", 60))

	// 1. 登录获取令牌
	token, err := loginAndGetToken("zhangwei@jinchenglaw.com", "law123456")
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 登录成功\n\n")

	// 2. 测试张伟律师代理字节跳动vs腾讯的冲突检测
	testConflictScenario(token)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("✅ 测试完成")
}

// loginAndGetToken 登录并获取认证令牌
func loginAndGetToken(email, password string) (string, error) {
	loginURL := "http://localhost:8080/api/auth/login"

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
		return "", fmt.Errorf("解析登录响应失败: %v\n响应内容: %s", err, string(body))
	}

	if !loginResp.Success || loginResp.Data.Token == "" {
		return "", fmt.Errorf("登录失败: %s", loginResp.Error)
	}

	return loginResp.Data.Token, nil
}

// testConflictScenario 测试冲突检测场景
func testConflictScenario(token string) {
	baseURL := "http://localhost:8080/api/v1/conflict/check"

	// 场景：张伟律师为字节跳动代理案件，对手为腾讯
	// 使用符合CheckConflictRequest结构的字段名
	request := TestConflictRequest{
		ClientID:                  "6", // 模拟客户ID
		ClientName:                "字节跳动科技有限公司",
		CaseName:                  "字节跳动诉腾讯商业纠纷案",
		CaseType:                  "commercial", // 小写
		ClientType:                "COMPANY",
		OtherParties:              []string{"深圳市腾讯计算机系统有限公司"},
		SearchYears:               5,
		IncludeCorporateRelations: true,
		SearchDepth:               "DEEP",
		UserID:                    "6", // 张伟律师的用户ID
		RequestTime:               time.Now(),
	}

	fmt.Println("🔍 测试场景: 张伟律师为字节跳动代理案件，对手为腾讯")

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

	// 显示结果
	fmt.Printf("✅ 测试完成 - 状态码: %d\n", resp.StatusCode)
	fmt.Printf("客户: %s\n", request.ClientName)
	if len(request.OtherParties) > 0 {
		fmt.Printf("对方: %s\n", request.OtherParties[0])
	}
	fmt.Printf("请求耗时: %v\n", time.Since(request.RequestTime))

	// 解析并显示响应详情
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		fmt.Printf("响应详情: %s\n", string(body))

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
			}
		}
	} else {
		fmt.Printf("❌ 响应解析失败，原始响应: %s\n", string(body))
	}
}