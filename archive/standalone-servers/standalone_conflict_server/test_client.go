package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const BASE_URL = "http://localhost:8080"

// ConflictCheckRequest 冲突检测请求
type ConflictCheckRequest struct {
	ClientID                 string    `json:"clientId"`
	ClientName               string    `json:"clientName"`
	ClientType               string    `json:"clientType"`
	OtherParties             []string  `json:"otherParties"`
	CaseName                 string    `json:"caseName"`
	CaseType                 string    `json:"caseType"`
	SearchYears              int       `json:"searchYears"`
	IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
	SearchDepth              string    `json:"searchDepth"`
	UserID                   uint      `json:"userId"`
	RequestTime              time.Time `json:"requestTime"`
}

func main() {
	fmt.Println("🧪 开始独立冲突检测服务API测试")
	fmt.Println("=" + "="*49)

	// 1. 测试服务健康检查
	fmt.Println("1️⃣ 测试服务健康检查...")
	testHealthCheck()

	// 2. 测试基础冲突检测
	fmt.Println("\n2️⃣ 测试基础冲突检测...")
	testBasicConflictDetection()

	// 3. 测试商业竞争冲突检测
	fmt.Println("\n3️⃣ 测试商业竞争冲突检测...")
	testBusinessCompetitionConflict()

	// 4. 测试法律对立冲突检测
	fmt.Println("\n4️⃣ 测试法律对立冲突检测...")
	testLegalOppositionConflict()

	// 5. 测试无冲突场景
	fmt.Println("\n5️⃣ 测试无冲突场景...")
	testNoConflictScenario()

	// 6. 测试检测历史
	fmt.Println("\n6️⃣ 测试检测历史...")
	testCheckHistory()

	fmt.Println("\n" + "="*49)
	fmt.Println("🎉 API测试完成！")
}

// testHealthCheck 测试健康检查
func testHealthCheck() {
	resp, err := http.Get(BASE_URL + "/api/v1/conflict/health")
	if err != nil {
		fmt.Printf("❌ 健康检查请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ 健康检查失败，状态码: %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	if response["code"] != float64(200) {
		fmt.Printf("❌ 健康检查响应码错误: %v\n", response["code"])
		return
	}

	data := response["data"].(map[string]interface{})
	service := data["service"].(string)
	status := data["status"].(string)

	fmt.Printf("✅ 健康检查通过: 服务=%s, 状态=%s\n", service, status)
}

// testBasicConflictDetection 测试基础冲突检测
func testBasicConflictDetection() {
	request := ConflictCheckRequest{
		ClientID:                 "1",
		ClientName:               "字节跳动",
		ClientType:               "COMPANY",
		OtherParties:             []string{"腾讯科技"},
		CaseName:                 "短视频平台版权纠纷",
		CaseType:                 "知识产权",
		SearchYears:              5,
		IncludeCorporateRelations: true,
		SearchDepth:              "STANDARD",
		UserID:                   1,
		RequestTime:              time.Now(),
	}

	testConflictDetection(request, "基础冲突检测")
}

// testBusinessCompetitionConflict 测试商业竞争冲突
func testBusinessCompetitionConflict() {
	request := ConflictCheckRequest{
		ClientID:                 "2",
		ClientName:               "新客户科技公司",
		ClientType:               "COMPANY",
		OtherParties:             []string{"美团", "阿里巴巴"},
		CaseName:                 "电商平台合作协议",
		CaseType:                 "商业纠纷",
		SearchYears:              3,
		IncludeCorporateRelations: true,
		SearchDepth:              "DEEP",
		UserID:                   2,
		RequestTime:              time.Now(),
	}

	testConflictDetection(request, "商业竞争冲突检测")
}

// testLegalOppositionConflict 测试法律对立冲突
func testLegalOppositionConflict() {
	request := ConflictCheckRequest{
		ClientID:                 "3",
		ClientName:               "原告公司",
		ClientType:               "COMPANY",
		OtherParties:             []string{"被告公司", "第三方公司"},
		CaseName:                 "合同违约诉讼",
		CaseType:                 "合同纠纷",
		SearchYears:              2,
		IncludeCorporateRelations: false,
		SearchDepth:              "STANDARD",
		UserID:                   1,
		RequestTime:              time.Now(),
	}

	testConflictDetection(request, "法律对立冲突检测")
}

// testNoConflictScenario 测试无冲突场景
func testNoConflictScenario() {
	request := ConflictCheckRequest{
		ClientID:                 "1",
		ClientName:               "全新客户公司",
		ClientType:               "COMPANY",
		OtherParties:             []string{"不相关公司"},
		CaseName:                 "内部法律咨询",
		CaseType:                 "法律咨询",
		SearchYears:              1,
		IncludeCorporateRelations: false,
		SearchDepth:              "BASIC",
		UserID:                   3,
		RequestTime:              time.Now(),
	}

	testConflictDetection(request, "无冲突场景检测")
}

// testConflictDetection 通用冲突检测测试
func testConflictDetection(request ConflictCheckRequest, testName string) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ %s - 请求序列化失败: %v\n", testName, err)
		return
	}

	resp, err := http.Post(
		BASE_URL+"/api/v1/conflict/check",
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		fmt.Printf("❌ %s - 请求失败: %v\n", testName, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ %s - HTTP错误: %d\n", testName, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("响应内容: %s\n", string(body))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ %s - 读取响应失败: %v\n", testName, err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ %s - 解析响应失败: %v\n", testName, err)
		return
	}

	if response["code"] != float64(200) {
		fmt.Printf("❌ %s - API错误: %v\n", testName, response["message"])
		return
	}

	data := response["data"].(map[string]interface{})

	// 解析响应数据
	checkId := data["checkId"].(string)
	hasConflict := data["hasConflict"].(bool)
	conflictCases := data["conflictCases"].([]interface{})
	riskAssessment := data["riskAssessment"].(map[string]interface{})
	checkStatistics := data["checkStatistics"].(map[string]interface{})

	// 输出结果
	fmt.Printf("✅ %s\n", testName)
	fmt.Printf("   检测ID: %s\n", checkId)
	fmt.Printf("   发现冲突: %t\n", hasConflict)
	fmt.Printf("   冲突案例数: %d\n", len(conflictCases))

	if riskLevel, ok := riskAssessment["overallRisk"]; ok {
		fmt.Printf("   风险等级: %v\n", riskLevel)
	}

	if riskScore, ok := riskAssessment["riskScore"]; ok {
		fmt.Printf("   风险分数: %.1f\n", riskScore)
	}

	if totalCases, ok := checkStatistics["totalCasesChecked"]; ok {
		fmt.Printf("   检查案件数: %v\n", totalCases)
	}

	// 显示冲突案例详情
	if len(conflictCases) > 0 {
		fmt.Printf("   冲突案例:\n")
		for i, caseInterface := range conflictCases {
			if i >= 3 { // 最多显示3个案例
				fmt.Printf("     ... 还有%d个案例\n", len(conflictCases)-3)
				break
			}
			conflictCase := caseInterface.(map[string]interface{})
			caseName := conflictCase["caseName"].(string)
			conflictType := conflictCase["conflictType"].(string)
			riskLevel := conflictCase["riskLevel"].(string)
			fmt.Printf("     %d. %s (%s) - %s\n", i+1, caseName, conflictType, riskLevel)
		}
	}

	// 显示建议
	if recommendations, ok := data["recommendations"].([]interface{}); ok && len(recommendations) > 0 {
		fmt.Printf("   建议:\n")
		for i, recInterface := range recommendations {
			if i >= 3 { // 最多显示3个建议
				fmt.Printf("     ... 还有%d个建议\n", len(recommendations)-3)
				break
			}
			rec := recInterface.(string)
			fmt.Printf("     %d. %s\n", i+1, rec)
		}
	}
}

// testCheckHistory 测试检测历史
func testCheckHistory() {
	resp, err := http.Get(BASE_URL + "/api/v1/conflict/history?clientId=1&limit=5")
	if err != nil {
		fmt.Printf("❌ 检测历史请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("❌ 检测历史失败，状态码: %d\n", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取检测历史响应失败: %v\n", err)
		return
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("❌ 解析检测历史响应失败: %v\n", err)
		return
	}

	if response["code"] != float64(200) {
		fmt.Printf("❌ 检测历史API错误: %v\n", response["message"])
		return
	}

	data := response["data"]
	if data == nil {
		fmt.Printf("✅ 检测历史: 无历史记录\n")
		return
	}

	history := data.([]interface{})
	fmt.Printf("✅ 检测历史: 找到%d条记录\n", len(history))
}