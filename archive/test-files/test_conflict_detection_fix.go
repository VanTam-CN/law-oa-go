package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"law-oa-go/internal/models"
)

// TestRequest 测试请求结构
type TestRequest struct {
	ClientID                string   `json:"clientId"`
	ClientName              string   `json:"clientName"`
	ClientType              string   `json:"clientType"`
	OtherParties            []string `json:"otherParties"`
	CaseName                string   `json:"caseName"`
	CaseType                string   `json:"caseType"`
	SearchYears             int      `json:"searchYears"`
	IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
	SearchDepth             string   `json:"searchDepth"`
	UserID                  uint     `json:"userId"`
	RequestTime             string   `json:"requestTime"`
}

// TestResponse 测试响应结构
type TestResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	log.Println("🧪 开始测试利益冲突检测服务修复效果")

	// 测试用例
	testCases := []TestRequest{
		{
			ClientID:                "1",
			ClientName:              "张三",
			ClientType:              "PERSON",
			OtherParties:            []string{"李四"},
			CaseName:                "张三诉李四合同纠纷案",
			CaseType:                "民事",
			SearchYears:             5,
			IncludeCorporateRelations: true,
			SearchDepth:             "STANDARD",
			UserID:                  1,
			RequestTime:             time.Now().Format(time.RFC3339),
		},
		{
			ClientID:                "2",
			ClientName:              "阿里巴巴集团控股有限公司",
			ClientType:              "COMPANY",
			OtherParties:            []string{"腾讯控股有限公司"},
			CaseName:                "阿里巴巴诉腾讯垄断纠纷案",
			CaseType:                "商业诉讼",
			SearchYears:             3,
			IncludeCorporateRelations: true,
			SearchDepth:             "DEEP",
			UserID:                  1,
			RequestTime:             time.Now().Format(time.RFC3339),
		},
		{
			ClientID:                "3",
			ClientName:              "字节跳动科技有限公司",
			ClientType:              "COMPANY",
			OtherParties:            []string{"百度", "阿里巴巴"},
			CaseName:                "字节跳动诉百度不正当竞争案",
			CaseType:                "知识产权",
			SearchYears:             7,
			IncludeCorporateRelations: true,
			SearchDepth:             "DEEP",
			UserID:                  2,
			RequestTime:             time.Now().Format(time.RFC3339),
		},
	}

	baseURL := getBaseURL()
	if baseURL == "" {
		log.Fatal("❌ 无法确定服务器地址，请确保服务正在运行")
	}

	log.Printf("📡 测试服务器地址: %s", baseURL)

	// 执行测试
	for i, testCase := range testCases {
		log.Printf("\n🔍 执行测试用例 %d: %s", i+1, testCase.CaseName)

		if err := testConflictDetection(baseURL, testCase); err != nil {
			log.Printf("❌ 测试用例 %d 失败: %v", i+1, err)
		} else {
			log.Printf("✅ 测试用例 %d 通过", i+1)
		}
	}

	// 测试健康检查
	log.Println("\n🏥 测试健康检查端点...")
	if err := testHealthCheck(baseURL); err != nil {
		log.Printf("❌ 健康检查失败: %v", err)
	} else {
		log.Printf("✅ 健康检查通过")
	}

	log.Println("\n🎉 所有测试完成!")
}

func testConflictDetection(baseURL string, request TestRequest) error {
	// 序列化请求
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", baseURL+"/api/v1/conflict/check", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 注意：在实际环境中，需要设置认证头
	// req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var response TestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 验证响应格式
	if response.Code != 200 {
		return fmt.Errorf("API返回错误码: %d, 消息: %s", response.Code, response.Message)
	}

	if response.Data == nil {
		return fmt.Errorf("响应数据为空")
	}

	// 打印测试结果
	log.Printf("  📋 响应码: %d", response.Code)
	log.Printf("  💬 消息: %s", response.Message)

	// 尝试解析具体的数据结构
	if dataMap, ok := response.Data.(map[string]interface{}); ok {
		if hasConflict, exists := dataMap["hasConflict"]; exists {
			log.Printf("  ⚡ 检测到冲突: %v", hasConflict)
		}
		if conflictCases, exists := dataMap["conflictCases"]; exists {
			if cases, ok := conflictCases.([]interface{}); ok {
				log.Printf("  📄 冲突案例数量: %d", len(cases))
			}
		}
		if riskAssessment, exists := dataMap["riskAssessment"]; exists {
			if risk, ok := riskAssessment.(map[string]interface{}); ok {
				if overallRisk, exists := risk["overallRisk"]; exists {
					log.Printf("  🎯 整体风险: %s", overallRisk)
				}
			}
		}
	}

	return nil
}

func testHealthCheck(baseURL string) error {
	// 创建健康检查请求
	req, err := http.NewRequest("GET", baseURL+"/api/v1/conflict/health", nil)
	if err != nil {
		return fmt.Errorf("创建健康检查请求失败: %w", err)
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送健康检查请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取健康检查响应失败: %w", err)
	}

	var response TestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析健康检查响应失败: %w", err)
	}

	log.Printf("  🏥 服务状态: %v", response.Data)

	return nil
}

func getBaseURL() string {
	// 尝试从环境变量获取
	if url := os.Getenv("API_BASE_URL"); url != "" {
		return url
	}

	// 默认本地开发地址
	return "http://localhost:8080"
}