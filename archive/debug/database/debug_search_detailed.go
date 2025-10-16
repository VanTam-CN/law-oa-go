package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// 直接测试数据库和API来定位问题
func main() {
	fmt.Println("🔍 深度调试客户搜索功能")
	fmt.Println("=====================================")

	// 1. 测试不同的API端点
	testDifferentEndpoints()

	// 2. 检查服务器响应头和详细信息
	testWithHeaders()
}

func testDifferentEndpoints() {
	baseURL := "http://localhost:8080"

	endpoints := []string{
		"/api/v1/clients",
		"/clients",
		"/api/clients",
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\n🌐 测试端点: %s\n", endpoint)

		// 测试基本查询
		url := baseURL + endpoint + "?page=1&page_size=10"
		testURL(url, "基本查询")

		// 测试搜索
		searchURL := baseURL + endpoint + "?page=1&page_size=10&name=张三"
		testURL(searchURL, "搜索'张三'(name参数)")

		// 测试search参数
		searchURL2 := baseURL + endpoint + "?page=1&page_size=10&search=张三"
		testURL(searchURL2, "搜索'张三'(search参数)")
	}
}

func testWithHeaders() {
	fmt.Println("\n📡 带详细信息的请求测试")

	url := "http://localhost:8080/api/v1/clients?name=张三&page=1&page_size=10"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	// 添加常见的请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Search-Debug-Tool/1.0")

	fmt.Printf("📡 请求详情:\n")
	fmt.Printf("   URL: %s\n", req.URL.String())
	fmt.Printf("   Method: %s\n", req.Method)
	fmt.Printf("   Headers:\n")
	for key, values := range req.Header {
		fmt.Printf("     %s: %s\n", key, strings.Join(values, ", "))
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📊 响应详情:\n")
	fmt.Printf("   状态码: %d\n", resp.StatusCode)
	fmt.Printf("   状态文本: %s\n", resp.Status)
	fmt.Printf("   响应头:\n")
	for key, values := range resp.Header {
		fmt.Printf("     %s: %s\n", key, strings.Join(values, ", "))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("📦 响应体 (前500字符):\n")
	bodyStr := string(body)
	if len(bodyStr) > 500 {
		fmt.Printf("   %s...\n", bodyStr[:500])
	} else {
		fmt.Printf("   %s\n", bodyStr)
	}

	// 尝试解析JSON
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("⚠️ JSON解析失败: %v\n", err)
	} else {
		fmt.Printf("✅ JSON解析成功\n")
		prettyJSON, _ := json.MarshalIndent(result, "   ", "  ")
		fmt.Printf("📋 结构化响应:\n%s\n", prettyJSON)
	}
}

func testURL(url, description string) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("   ❌ %s: 请求失败 - %v\n", description, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ %s: 读取失败 - %v\n", description, err)
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("   ⚠️ %s: JSON解析失败 (状态: %d, 长度: %d)\n",
			description, resp.StatusCode, len(body))
		// 显示原始响应的一部分
		bodyStr := string(body)
		if len(bodyStr) > 100 {
			fmt.Printf("      原始响应: %s...\n", bodyStr[:100])
		} else {
			fmt.Printf("      原始响应: %s\n", bodyStr)
		}
		return
	}

	fmt.Printf("   ✅ %s: 状态 %d\n", description, resp.StatusCode)

	// 提取关键信息
	if data, ok := result["data"].([]interface{}); ok {
		fmt.Printf("      返回记录数: %d\n", len(data))

		// 显示前几条记录
		for i, item := range data {
			if i >= 3 {
				fmt.Printf("      ... 还有 %d 条\n", len(data)-3)
				break
			}
			if client, ok := item.(map[string]interface{}); ok {
				name := client["name"]
				clientType := client["type"]
				fmt.Printf("      - %v (%v)\n", name, clientType)
			}
		}
	}

	if pagination, ok := result["pagination"].(map[string]interface{}); ok {
		if total, ok := pagination["total"].(float64); ok {
			fmt.Printf("      总记录数: %.0f\n", total)
		}
	}
}