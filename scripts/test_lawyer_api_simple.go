//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 测试律师API
	baseURL := "http://localhost:8080/api/v1"

	// 1. 先测试公开端点，确认服务器运行
	fmt.Println("🔍 测试服务器连通性...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		log.Fatal("服务器连接失败:", err)
	}
	resp.Body.Close()
	fmt.Println("✅ 服务器连接正常")

	// 2. 测试律师API（无认证）
	fmt.Println("\n🔍 测试律师API（无认证）...")
	resp, err = http.Get(baseURL + "/lawfirm/lawyers?page=1&page_size=5")
	if err != nil {
		log.Fatal("律师API请求失败:", err)
	}
	defer resp.Body.Close()

	fmt.Printf("状态码: %d\n", resp.StatusCode)

	if resp.StatusCode == 401 {
		fmt.Println("✅ 律师API正确需要认证")
		return
	}

	// 解析响应
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Fatal("解析响应失败:", err)
	}

	fmt.Printf("响应结构:\n")
	if code, ok := response["code"]; ok {
		fmt.Printf("  code: %v\n", code)
	}
	if message, ok := response["message"]; ok {
		fmt.Printf("  message: %v\n", message)
	}
	if data, ok := response["data"]; ok {
		fmt.Printf("  data type: %T\n", data)
		if dataArray, ok := data.([]interface{}); ok {
			fmt.Printf("  data length: %d\n", len(dataArray))
			if len(dataArray) > 0 {
				fmt.Printf("  第一个律师: %+v\n", dataArray[0])
			}
		}
	}

	fmt.Printf("\n完整响应:\n")
	prettyJSON, _ := json.MarshalIndent(response, "", "  ")
	fmt.Println(string(prettyJSON))
}