package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

// APIResponse 统一API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func main() {
	fmt.Println("🧪 测试冲突检测API修复...")

	// 创建测试请求
	request := ConflictCheckRequest{
		ClientID:                  "57",
		ClientName:                "字节跳动科技有限公司",
		CaseName:                  "字节跳动诉腾讯垄断纠纷案",
		CaseType:                  "commercial",
		ClientType:                "PERSON",
		OtherParties:              []string{"腾讯", "垄断纠纷案"},
		SearchYears:               5,
		IncludeCorporateRelations: true,
		SearchDepth:               "DEEP",
		UserID:                    "45",
		RequestTime:               time.Now().Format(time.RFC3339),
		CauseOfAction:             "字节跳动诉腾讯垄断纠纷案",
	}

	// 序列化请求
	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ 序列化请求失败: %v\n", err)
		return
	}

	fmt.Printf("📤 发送请求: %s\n", string(requestJSON))

	// 发送HTTP请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post("http://localhost:8080/api/v1/conflict/check", "application/json", bytes.NewBuffer(requestJSON))
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
	var apiResponse APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	// 验证响应格式
	if resp.StatusCode == 200 {
		if apiResponse.Success {
			if apiResponse.Data != nil {
				fmt.Println("✅ API修复成功！")
				fmt.Println("✅ 响应包含正确的data字段")
				fmt.Println("✅ 冲突检测功能正常工作")

				// 打印数据详情
				if dataMap, ok := apiResponse.Data.(map[string]interface{}); ok {
					fmt.Printf("📊 检测结果:\n")
					fmt.Printf("   - 检测ID: %v\n", dataMap["checkId"])
					fmt.Printf("   - 是否有冲突: %v\n", dataMap["hasConflict"])
					fmt.Printf("   - 风险评估: %v\n", dataMap["riskAssessment"])
				}
			} else {
				fmt.Println("❌ 响应缺少data字段")
			}
		} else {
			fmt.Printf("❌ API返回失败: %v\n", apiResponse.Error)
		}
	} else {
		fmt.Printf("❌ HTTP状态码错误: %d\n", resp.StatusCode)
	}
}
