package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
		User  struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ConflictCheckRequest 冲突检测请求结构
type ConflictCheckRequest struct {
	ClientID                  string   `json:"clientId"`
	ClientName                string   `json:"clientName"`
	CaseName                  string   `json:"caseName"`
	CaseType                  string   `json:"caseType"`
	ClientType                string   `json:"clientType"`
	OtherParties              []string `json:"otherParties"`
	SearchYears               int      `json:"searchYears"`
	IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
	SearchDepth               string   `json:"searchDepth"`
	UserID                    string   `json:"userId"`
	RequestTime               string   `json:"requestTime"`
	CauseOfAction             string   `json:"causeOfAction"`
}

func main() {
	fmt.Println("🧪 测试后端冲突检测API...")

	// 测试场景1: 张伟律师为字节跳动创建案件，对手方是腾讯
	testScenario1()

	// 测试场景2: 陈浩律师为朱丽倩创建案件
	testScenario2()

	// 测试场景3: 张伟律师为宝能集团创建案件，对手方是万科
	testScenario3()
}

func testScenario1() {
	fmt.Println("\n🔍 场景1: 张伟律师为字节跳动创建案件，对手方是腾讯")

	request := ConflictCheckRequest{
		ClientID:                  "57", // 字节跳动科技有限公司
		ClientName:                "字节跳动科技有限公司",
		CaseName:                  "字节跳动诉腾讯垄断纠纷案",
		CaseType:                  "commercial",
		ClientType:                "COMPANY",
		OtherParties:              []string{"腾讯"},
		SearchYears:               5,
		IncludeCorporateRelations: true,
		SearchDepth:               "DEEP",
		UserID:                    "45", // 张伟律师
		RequestTime:               time.Now().Format(time.RFC3339),
		CauseOfAction:             "垄断纠纷",
	}

	callConflictAPI(request, "场景1")
}

func testScenario2() {
	fmt.Println("\n🔍 场景2: 陈浩律师为朱丽倩创建案件")

	request := ConflictCheckRequest{
		ClientID:                  "64", // 朱丽倩
		ClientName:                "朱丽倩",
		CaseName:                  "朱丽倩离婚财产分割案",
		CaseType:                  "civil",
		ClientType:                "PERSON",
		OtherParties:              []string{},
		SearchYears:               5,
		IncludeCorporateRelations: false,
		SearchDepth:               "STANDARD",
		UserID:                    "48", // 陈浩律师
		RequestTime:               time.Now().Format(time.RFC3339),
		CauseOfAction:             "婚姻家庭纠纷",
	}

	callConflictAPI(request, "场景2")
}

func testScenario3() {
	fmt.Println("\n🔍 场景3: 张伟律师为宝能集团创建案件，对手方是万科")

	// 注意：这里使用一个不存在的客户ID来模拟宝能集团
	request := ConflictCheckRequest{
		ClientID:                  "999", // 模拟宝能集团
		ClientName:                "宝能集团股份有限公司",
		CaseName:                  "宝能集团诉万科股权纠纷案",
		CaseType:                  "commercial",
		ClientType:                "COMPANY",
		OtherParties:              []string{"万科"},
		SearchYears:               5,
		IncludeCorporateRelations: true,
		SearchDepth:               "DEEP",
		UserID:                    "45", // 张伟律师
		RequestTime:               time.Now().Format(time.RFC3339),
		CauseOfAction:             "股权纠纷",
	}

	callConflictAPI(request, "场景3")
}

func callConflictAPI(request ConflictCheckRequest, scenarioName string) {
	// 序列化请求
	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ %s: 序列化请求失败: %v\n", scenarioName, err)
		return
	}

	fmt.Printf("📤 发送请求到: http://localhost:8080/api/v1/conflict/check\n")
	fmt.Printf("📋 请求数据: %s\n", string(requestJSON))

	// 获取认证令牌
	token := getAuthToken()
	if token == "" {
		fmt.Printf("❌ %s: 无法获取认证令牌\n", scenarioName)
		return
	}

	// 创建带认证的HTTP请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/conflict/check", bytes.NewBuffer(requestJSON))
	if err != nil {
		fmt.Printf("❌ %s: 创建请求失败: %v\n", scenarioName, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ %s: 发送请求失败: %v\n", scenarioName, err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ %s: 读取响应失败: %v\n", scenarioName, err)
		return
	}

	fmt.Printf("📥 响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("📥 响应内容: %s\n", string(body))

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ %s: 解析响应失败: %v\n", scenarioName, err)
		return
	}

	// 分析响应
	analyzeResponse(response, scenarioName)
}

func analyzeResponse(response map[string]interface{}, scenarioName string) {
	fmt.Printf("\n📊 %s 响应分析:\n", scenarioName)

	// 检查success字段
	if success, ok := response["success"].(bool); ok {
		fmt.Printf("   ✅ Success: %t\n", success)

		if success {
			// 检查data字段
			if data, ok := response["data"].(map[string]interface{}); ok {
				fmt.Printf("   📋 Data字段存在\n")

				// 检查hasConflict
				if hasConflict, ok := data["hasConflict"].(bool); ok {
					fmt.Printf("   🔍 HasConflict: %t\n", hasConflict)
				}

				// 检查conflictCases
				if conflictCases, ok := data["conflictCases"].([]interface{}); ok {
					fmt.Printf("   📁 ConflictCases: %d 个\n", len(conflictCases))

					for i, caseData := range conflictCases {
						if caseMap, ok := caseData.(map[string]interface{}); ok {
							fmt.Printf("      案件 %d:\n", i+1)
							if caseName, ok := caseMap["caseName"].(string); ok {
								fmt.Printf("         名称: %s\n", caseName)
							}
							if riskLevel, ok := caseMap["riskLevel"].(string); ok {
								fmt.Printf("         风险等级: %s\n", riskLevel)
							}
							if conflictType, ok := caseMap["conflictType"].(string); ok {
								fmt.Printf("         冲突类型: %s\n", conflictType)
							}
						}
					}
				}

				// 检查riskAssessment
				if riskAssessment, ok := data["riskAssessment"].(map[string]interface{}); ok {
					if overallRisk, ok := riskAssessment["overallRisk"].(string); ok {
						fmt.Printf("   ⚠️ 总体风险: %s\n", overallRisk)
					}
				}
			} else {
				fmt.Printf("   ❌ 缺少data字段\n")
			}
		} else {
			// 检查error字段
			if errorData, ok := response["error"].(map[string]interface{}); ok {
				if message, ok := errorData["message"].(string); ok {
					fmt.Printf("   ❌ 错误信息: %s\n", message)
				}
			}
		}
	} else {
		fmt.Printf("   ❌ 响应格式不正确，缺少success字段\n")
	}
}

func getAuthToken() string {
	// 尝试登录陈浩律师账号
	loginReq := LoginRequest{
		Email:    "chenhao@law.com",
		Password: "law123456",
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		fmt.Printf("❌ 序列化登录请求失败: %v\n", err)
		return ""
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("❌ 发送登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取登录响应失败: %v\n", err)
		return ""
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		fmt.Printf("❌ 解析登录响应失败: %v\n", err)
		return ""
	}

	if !loginResp.Success {
		fmt.Printf("❌ 登录失败: %s\n", loginResp.Error.Message)
		return ""
	}

	fmt.Printf("✅ 登录成功: %s (ID: %d)\n", loginResp.Data.User.Name, loginResp.Data.User.ID)
	return loginResp.Data.Token
}
