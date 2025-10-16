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
	ClientID                 string    `json:"clientId"`
	ClientName               string    `json:"clientName"`
	CaseName                 string    `json:"caseName"`
	CaseType                 string    `json:"caseType"`
	ClientType               string    `json:"clientType"`
	OtherParties             []string  `json:"otherParties"`
	SearchYears              int       `json:"searchYears"`
	IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
	SearchDepth              string    `json:"searchDepth"`
	UserID                   string    `json:"userId"`
	RequestTime              time.Time `json:"requestTime"`
}

func main() {
	baseURL := "http://localhost:8080"

	// 测试冲突检测API
	fmt.Println("=== 测试利益冲突检测API ===")

	// 1. 测试健康检查
	fmt.Println("\n1. 测试健康检查端点...")
	testHealthCheck(baseURL)

	// 2. 测试冲突检测
	fmt.Println("\n2. 测试冲突检测端点...")
	testConflictCheck(baseURL)

	// 3. 测试MCP标准获取
	fmt.Println("\n3. 测试MCP标准获取...")
	testMCPStandards(baseURL)

	// 4. 测试冲突规则获取
	fmt.Println("\n4. 测试冲突规则获取...")
	testConflictRules(baseURL)

	fmt.Println("\n=== API测试完成 ===")
}

func testHealthCheck(baseURL string) {
	url := baseURL + "/conflict/health"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 健康检查响应 (状态码: %d):\n%s\n", resp.StatusCode, string(body))
}

func testConflictCheck(baseURL string) {
	url := baseURL + "/conflict/check"

	// 构造测试请求
	request := ConflictCheckRequest{
		ClientID:                 "1",
		ClientName:               "测试客户",
		CaseName:                 "测试案件",
		CaseType:                 "civil",
		ClientType:               "PERSON",
		OtherParties:             []string{"对方当事人"},
		SearchYears:              5,
		IncludeCorporateRelations: true,
		SearchDepth:              "deep",
		UserID:                   "1",
		RequestTime:              time.Now(),
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("❌ 序列化请求失败: %v\n", err)
		return
	}

	fmt.Printf("📤 发送请求: %s\n", string(jsonData))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 冲突检测失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 冲突检测响应 (状态码: %d):\n%s\n", resp.StatusCode, string(body))
}

func testMCPStandards(baseURL string) {
	url := baseURL + "/conflict/standards"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ 获取MCP标准失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("✅ MCP标准响应 (状态码: %d):\n%s\n", resp.StatusCode, string(body))
}

func testConflictRules(baseURL string) {
	url := baseURL + "/conflict/rules"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ 获取冲突规则失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 冲突规则响应 (状态码: %d):\n%s\n", resp.StatusCode, string(body))
}