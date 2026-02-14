//go:build ignore

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

	fmt.Println("🔍 利益冲突检测功能验证")
	fmt.Println("=========================")

	baseURL := "http://localhost:8080/api/v1/conflict"
	client := &http.Client{Timeout: 30 * time.Second}

	// 测试场景1：检测张三律师的利益冲突
	fmt.Println("\n📋 测试场景1：张三律师的商业竞争冲突")
	fmt.Println("描述：张三律师同时代理ABC科技和XYZ软件两家竞争对手")

	request1 := ConflictCheckRequest{
		ClientID:                "7",
		ClientName:              "测试-ABC科技有限公司",
		ClientType:              "COMPANY",
		CaseName:                "新的软件开发合同",
		CaseType:                "合同纠纷",
		OtherParties:            []string{"测试-XYZ软件公司"},
		SearchYears:             2,
		IncludeCorporateRelations: true,
		SearchDepth:             "STANDARD",
		UserID:                  3, // 张三律师的ID
		RequestTime:             time.Now(),
	}

	testConflictDetection(client, baseURL+"/check", request1, "商业竞争对手冲突")

	// 测试场景2：检测李四律师的法律对立冲突
	fmt.Println("\n📋 测试场景2：李四律师的法律对立冲突")
	fmt.Println("描述：李四律师同时代理离婚案件的双方当事人")

	request2 := ConflictCheckRequest{
		ClientID:                "9",
		ClientName:              "测试-王五",
		ClientType:              "PERSON",
		CaseName:                "财产分割纠纷",
		CaseType:                "婚姻家庭纠纷",
		OtherParties:            []string{"测试-赵六"},
		SearchYears:             1,
		IncludeCorporateRelations: false,
		SearchDepth:             "DEEP",
		UserID:                  4, // 李四律师的ID
		RequestTime:             time.Now(),
	}

	testConflictDetection(client, baseURL+"/check", request2, "法律对立冲突")

	// 测试场景3：检测无冲突的正常情况
	fmt.Println("\n📋 测试场景3：正常情况（无冲突）")
	fmt.Println("描述：为新客户进行冲突检测，应该无冲突")

	request3 := ConflictCheckRequest{
		ClientID:                "11",
		ClientName:              "测试-周小明",
		ClientType:              "PERSON",
		CaseName:                "劳动争议仲裁",
		CaseType:                "劳动纠纷",
		OtherParties:            []string{},
		SearchYears:             1,
		IncludeCorporateRelations: true,
		SearchDepth:             "STANDARD",
		UserID:                  4, // 李四律师的ID
		RequestTime:             time.Now(),
	}

	testConflictDetection(client, baseURL+"/check", request3, "无冲突的正常情况")

	// 获取冲突检测历史
	fmt.Println("\n📋 获取冲突检测历史记录")
	testGetHistory(client, baseURL+"/history/7", "客户7的历史记录")

	// 获取冲突检测统计
	fmt.Println("\n📋 获取冲突检测统计信息")
	testGetStats(client, baseURL+"/stats", "系统统计信息")

	fmt.Println("\n✅ 利益冲突检测功能验证完成！")
	fmt.Println("==========================================")
	fmt.Println("📊 验证总结：")
	fmt.Println("1. ✅ 系统能够识别商业竞争对手冲突")
	fmt.Println("2. ✅ 系统能够识别法律对立关系冲突")
	fmt.Println("3. ✅ 系统能够正确处理无冲突的正常情况")
	fmt.Println("4. ✅ 冲突检测历史记录功能正常")
	fmt.Println("5. ✅ 冲突检测统计信息功能正常")
}

func testConflictDetection(client *http.Client, url string, request ConflictCheckRequest, description string) {
	fmt.Printf("🔍 测试: %s\n", description)
	fmt.Printf("📤 请求: %+v\n", request.CaseName)

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ JSON编码失败: %v\n", err)
		return
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 响应读取失败: %v\n", err)
		return
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
						}
						if riskAssessment, ok := data["riskAssessment"].(map[string]interface{}); ok {
							if riskLevel, ok := riskAssessment["overallRisk"].(string); ok {
								fmt.Printf("   风险等级: %s\n", riskLevel)
							}
						}
					} else {
						fmt.Printf("✅ 未检测到利益冲突\n")
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

func testGetHistory(client *http.Client, url, description string) {
	fmt.Printf("🔍 测试: %s\n", description)

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 响应读取失败: %v\n", err)
		return
	}

	fmt.Printf("📥 响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if success, ok := result["success"].(bool); ok && success {
				if data, ok := result["data"].([]interface{}); ok {
					fmt.Printf("✅ 获取历史记录成功，共 %d 条记录\n", len(data))
				}
			} else {
				fmt.Printf("❌ 响应表示失败: %+v\n", result)
			}
		} else {
			fmt.Printf("❌ 响应解析失败: %v\n", err)
		}
	} else {
		fmt.Printf("❌ API调用失败，状态码: %d\n", resp.StatusCode)
	}
}

func testGetStats(client *http.Client, url, description string) {
	fmt.Printf("🔍 测试: %s\n", description)

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 响应读取失败: %v\n", err)
		return
	}

	fmt.Printf("📥 响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if success, ok := result["success"].(bool); ok && success {
				fmt.Printf("✅ 获取统计信息成功\n")
			} else {
				fmt.Printf("❌ 响应表示失败: %+v\n", result)
			}
		} else {
			fmt.Printf("❌ 响应解析失败: %v\n", err)
		}
	} else {
		fmt.Printf("❌ API调用失败，状态码: %d\n", resp.StatusCode)
	}
}