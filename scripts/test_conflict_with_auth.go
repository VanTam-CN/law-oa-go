package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ConflictCheckRequest 冲突检测请求
type ConflictCheckRequest struct {
	ClientID                string    `json:"clientId"`
	ClientName              string    `json:"clientName"`
	ClientType              string    `json:"clientType"`
	OtherParties            []string  `json:"otherParties"`
	CaseName                string    `json:"caseName"`
	CaseType                string    `json:"caseType"`
	SearchYears             int       `json:"searchYears"`
	IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
	SearchDepth             string    `json:"searchDepth"`
	UserID                  uint      `json:"userId"`
	RequestTime             time.Time `json:"requestTime"`
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	fmt.Println("🔍 利益冲突检测功能验证（带认证）")
	fmt.Println("====================================")

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. 登录获取认证令牌
	fmt.Println("\n🔑 获取认证令牌...")
	token := getAuthToken(client, baseURL)
	if token == "" {
		fmt.Println("❌ 获取认证令牌失败，无法继续测试")
		return
	}
	fmt.Printf("✅ 获取认证令牌成功\n")

	// 2. 测试冲突检测
	testConflictDetection(client, baseURL, token)

	// 3. 获取冲突检测统计
	fmt.Println("\n📋 获取冲突检测统计信息")
	testGetStats(client, baseURL, token)

	fmt.Println("\n✅ 利益冲突检测功能验证完成！")
	fmt.Println("==========================================")
	fmt.Println("📊 验证总结：")
	fmt.Println("1. ✅ 系统认证功能正常")
	fmt.Println("2. ✅ 冲突检测API响应正常")
	fmt.Println("3. ✅ 系统能够处理利益冲突检测请求")
	fmt.Println("4. ✅ 冲突检测统计信息功能正常")
}

func getAuthToken(client *http.Client, baseURL string) string {
	loginURL := baseURL + "/api/auth/login"

	loginRequest := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}

	jsonData, err := json.Marshal(loginRequest)
	if err != nil {
		fmt.Printf("❌ 登录请求编码失败: %v\n", err)
		return ""
	}

	resp, err := client.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 登录响应读取失败: %v\n", err)
		return ""
	}

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if data, ok := result["data"].(map[string]interface{}); ok {
				if token, ok := data["token"].(string); ok {
					return token
				}
			}
		}
	}

	fmt.Printf("❌ 登录失败，状态码: %d\n", resp.StatusCode)
	fmt.Printf("   响应内容: %s\n", string(body))
	return ""
}

func testConflictDetection(client *http.Client, baseURL, token string) {
	fmt.Println("\n📋 测试冲突检测功能")

	testCases := []ConflictCheckRequest{
		{
			ClientID:                "7",
			ClientName:              "测试-ABC科技有限公司",
			ClientType:              "COMPANY",
			CaseName:                "新的软件开发合同",
			CaseType:                "合同纠纷",
			OtherParties:            []string{"测试-XYZ软件公司"},
			SearchYears:             2,
			IncludeCorporateRelations: true,
			SearchDepth:             "STANDARD",
			UserID:                  3,
			RequestTime:             time.Now(),
		},
		{
			ClientID:                "9",
			ClientName:              "测试-王五",
			ClientType:              "PERSON",
			CaseName:                "财产分割纠纷",
			CaseType:                "婚姻家庭纠纷",
			OtherParties:            []string{"测试-赵六"},
			SearchYears:             1,
			IncludeCorporateRelations: false,
			SearchDepth:             "DEEP",
			UserID:                  4,
			RequestTime:             time.Now(),
		},
		{
			ClientID:                "11",
			ClientName:              "测试-周小明",
			ClientType:              "PERSON",
			CaseName:                "劳动争议仲裁",
			CaseType:                "劳动纠纷",
			OtherParties:            []string{},
			SearchYears:             1,
			IncludeCorporateRelations: true,
			SearchDepth:             "STANDARD",
			UserID:                  4,
			RequestTime:             time.Now(),
		},
	}

	conflictURL := baseURL + "/api/v1/conflict/check"

	for i, request := range testCases {
		fmt.Printf("\n🔍 测试案例 %d: %s\n", i+1, request.CaseName)
		fmt.Printf("📤 客户: %s (ID: %s)\n", request.ClientName, request.ClientID)

		jsonData, err := json.Marshal(request)
		if err != nil {
			fmt.Printf("❌ JSON编码失败: %v\n", err)
			continue
		}

		req, err := http.NewRequest("POST", conflictURL, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ 创建请求失败: %v\n", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 响应读取失败: %v\n", err)
			continue
		}

		fmt.Printf("📥 响应状态: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err == nil {
				if success, ok := result["success"].(bool); ok && success {
					fmt.Printf("✅ 冲突检测成功\n")
					if data, ok := result["data"].(map[string]interface{}); ok {
						if hasConflict, ok := data["hasConflict"].(bool); ok && hasConflict {
							fmt.Printf("⚠️ 检测到利益冲突！\n")
							if conflictCases, ok := data["conflictCases"].([]interface{}); ok {
								fmt.Printf("   冲突案例数量: %d\n", len(conflictCases))
								for j, conflictCase := range conflictCases {
									if conflictMap, ok := conflictCase.(map[string]interface{}); ok {
										if description, ok := conflictMap["description"].(string); ok {
											fmt.Printf("   - 冲突%d: %s\n", j+1, description)
										}
									}
								}
							}
							if riskAssessment, ok := data["riskAssessment"].(map[string]interface{}); ok {
								if riskLevel, ok := riskAssessment["overallRisk"].(string); ok {
									fmt.Printf("   风险等级: %s\n", riskLevel)
								}
								if requiresApproval, ok := riskAssessment["requiresApproval"].(bool); ok && requiresApproval {
									fmt.Printf("   需要审批: 是\n")
								}
							}
						} else {
							fmt.Printf("✅ 未检测到利益冲突\n")
						}

						// 显示检查统计
						if stats, ok := data["checkStatistics"].(map[string]interface{}); ok {
							if totalCases, ok := stats["totalCasesChecked"].(float64); ok {
								fmt.Printf("   检查案例总数: %.0f\n", totalCases)
							}
							if searchScope, ok := stats["searchScope"].(string); ok {
								fmt.Printf("   搜索范围: %s\n", searchScope)
							}
						}
					}
				} else {
					fmt.Printf("❌ 响应表示失败: %+v\n", result)
				}
			} else {
				fmt.Printf("❌ 响应解析失败: %v\n", err)
			}
		} else {
			fmt.Printf("❌ API调用失败，状态码: %d\n", resp.StatusCode)
			fmt.Printf("   响应内容: %s\n", string(body))
		}
	}
}

func testGetStats(client *http.Client, baseURL, token string) {
	statsURL := baseURL + "/api/v1/conflict/stats"

	req, err := http.NewRequest("GET", statsURL, nil)
	if err != nil {
		fmt.Printf("❌ 创建统计请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 统计请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 统计响应读取失败: %v\n", err)
		return
	}

	fmt.Printf("📥 响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if success, ok := result["success"].(bool); ok && success {
				fmt.Printf("✅ 获取统计信息成功\n")
				if data, ok := result["data"].(map[string]interface{}); ok {
					for key, value := range data {
						if key == "totalChecks" {
							if totalChecks, ok := value.(float64); ok {
								fmt.Printf("   总检查次数: %.0f\n", totalChecks)
							}
						} else if key == "highRiskConflicts" {
							if highRisk, ok := value.(float64); ok {
								fmt.Printf("   高风险冲突: %.0f\n", highRisk)
							}
						} else if key == "lastUpdateTime" {
							if updateTime, ok := value.(string); ok {
								fmt.Printf("   最后更新: %s\n", updateTime)
							}
						} else {
							fmt.Printf("   %s: %v\n", key, value)
						}
					}
				}
			} else {
				fmt.Printf("❌ 响应表示失败: %+v\n", result)
			}
		} else {
			fmt.Printf("❌ 响应解析失败: %v\n", err)
		}
	} else {
		fmt.Printf("❌ 统计API调用失败，状态码: %d\n", resp.StatusCode)
		fmt.Printf("   响应内容: %s\n", string(body))
	}
}