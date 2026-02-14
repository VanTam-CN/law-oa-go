//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 测试客户列表API
	fmt.Println("=== 测试客户列表API ===")

	url := "http://localhost:8080/api/clients?page=1&page_size=5"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	// 添加认证头（如果有token的话）
	// req.Header.Set("Authorization", "Bearer your-token-here")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		fmt.Println("提示：请确保后端服务器正在运行在 http://localhost:8080")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应体:\n%s\n", string(body))

	// 尝试解析JSON响应
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("JSON解析失败: %v\n", err)
		return
	}

	fmt.Println("\n=== 结构化响应分析 ===")

	// 检查常见的响应格式
	if success, ok := apiResponse["success"].(bool); ok && success {
		fmt.Println("✅ API响应格式: {success: true, data: [...]}")

		if data, ok := apiResponse["data"].([]interface{}); ok {
			fmt.Printf("📊 数据条数: %d\n", len(data))

			if len(data) > 0 {
				if firstClient, ok := data[0].(map[string]interface{}); ok {
					fmt.Println("📋 第一个客户字段:")
					for key, value := range firstClient {
						fmt.Printf("  %s: %v\n", key, value)
					}
				}
			}
		}

		if pagination, ok := apiResponse["pagination"].(map[string]interface{}); ok {
			fmt.Println("📄 分页信息:")
			for key, value := range pagination {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	} else if data, ok := apiResponse["data"]; ok {
		fmt.Println("✅ API响应格式: {data: [...]}")
		if dataArray, ok := data.([]interface{}); ok {
			fmt.Printf("📊 数据条数: %d\n", len(dataArray))
		}
	} else if list, ok := apiResponse["list"]; ok {
		fmt.Println("✅ API响应格式: {list: [...]}")
		if dataArray, ok := list.([]interface{}); ok {
			fmt.Printf("📊 数据条数: %d\n", len(dataArray))
		}
	} else {
		fmt.Println("❌ 未知的响应格式")
	}
}