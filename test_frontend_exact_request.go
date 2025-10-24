package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 模拟前端发送的确切请求格式
type FrontendConflictRequest struct {
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

func main() {
	fmt.Println("🧪 模拟前端确切请求...")

	// 测试场景：模拟前端用户在新建案件时的冲突检测
	// 这里使用从前端错误信息中看到的确切数据
	testFrontendScenario()
}

func testFrontendScenario() {
	fmt.Println("\n🔍 模拟前端场景: 字节跳动诉腾讯垄断纠纷案")

	// 这是从前端错误日志中看到的确切请求数据
	request := FrontendConflictRequest{
		ClientID:                  "57",
		ClientName:                "字节跳动科技有限公司",
		CaseName:                  "字节跳动诉腾讯垄断纠纷案",
		CaseType:                  "commercial",
		ClientType:                "PERSON", // 注意：前端设置为PERSON
		OtherParties:              []string{"腾讯", "垄断纠纷案"},
		SearchYears:               5,
		IncludeCorporateRelations: true,
		SearchDepth:               "DEEP",
		UserID:                    "45",                       // 张伟律师
		RequestTime:               "2025-10-24T07:21:27.264Z", // 使用前端的时间格式
		CauseOfAction:             "字节跳动诉腾讯垄断纠纷案",
	}

	// 获取认证令牌 - 使用张伟的账号
	token := getZhangweiToken()
	if token == "" {
		fmt.Println("❌ 无法获取张伟的认证令牌，尝试使用陈浩的令牌")
		token = getChenhaoToken()
		if token == "" {
			fmt.Println("❌ 无法获取任何认证令牌")
			return
		}
	}

	callConflictAPI(request, token)
}

func getZhangweiToken() string {
	// 尝试多种可能的张伟账号
	accounts := []struct {
		email    string
		password string
	}{
		{"zhangwei@law.com", "law123456"},
		{"zhangwei", "law123456"},
		{"zhang.wei@law.com", "law123456"},
	}

	for _, account := range accounts {
		token := tryLogin(account.email, account.password, "张伟")
		if token != "" {
			return token
		}
	}

	return ""
}

func getChenhaoToken() string {
	return tryLogin("chenhao@law.com", "law123456", "陈浩")
}

func tryLogin(email, password, name string) string {
	fmt.Printf("🔑 尝试登录: %s (%s)\n", name, email)

	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return ""
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return ""
	}

	if !loginResp.Success {
		fmt.Printf("❌ %s 登录失败: %s\n", name, loginResp.Error.Message)
		return ""
	}

	fmt.Printf("✅ %s 登录成功 (ID: %d)\n", loginResp.Data.User.Name, loginResp.Data.User.ID)
	return loginResp.Data.Token
}

func callConflictAPI(request FrontendConflictRequest, token string) {
	// 序列化请求
	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ 序列化请求失败: %v\n", err)
		return
	}

	fmt.Printf("📤 发送前端模拟请求:\n")
	fmt.Printf("📋 请求数据: %s\n", string(requestJSON))

	// 创建HTTP请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/conflict/check", bytes.NewBuffer(requestJSON))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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

	fmt.Printf("📥 响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("📥 响应内容: %s\n", string(body))

	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	// 分析响应
	analyzeResponse(response)
}

func analyzeResponse(response map[string]interface{}) {
	fmt.Printf("\n📊 前端模拟请求响应分析:\n")

	if success, ok := response["success"].(bool); ok && success {
		if data, ok := response["data"].(map[string]interface{}); ok {
			fmt.Printf("   ✅ 包含data字段\n")

			if hasConflict, ok := data["hasConflict"].(bool); ok {
				fmt.Printf("   🔍 HasConflict: %t\n", hasConflict)

				if hasConflict {
					if conflictCases, ok := data["conflictCases"].([]interface{}); ok {
						fmt.Printf("   📁 发现 %d 个冲突案例\n", len(conflictCases))

						for i, caseData := range conflictCases {
							if caseMap, ok := caseData.(map[string]interface{}); ok {
								fmt.Printf("      案件 %d: %s (风险: %s)\n",
									i+1,
									caseMap["caseName"],
									caseMap["riskLevel"])
							}
						}
					}

					if riskAssessment, ok := data["riskAssessment"].(map[string]interface{}); ok {
						fmt.Printf("   ⚠️ 总体风险: %s\n", riskAssessment["overallRisk"])
					}

					fmt.Println("\n🎯 结论: 后端API正常工作，前端应该能收到冲突结果")
				} else {
					fmt.Println("   ℹ️ 未检测到冲突")
				}
			}
		} else {
			fmt.Printf("   ❌ 缺少data字段 - 这就是前端报错的原因！\n")
		}
	} else {
		fmt.Printf("   ❌ API调用失败\n")
		if errorData, ok := response["error"].(map[string]interface{}); ok {
			fmt.Printf("   错误: %s\n", errorData["message"])
		}
	}
}
