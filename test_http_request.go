package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== 测试HTTP客户API请求 ===")

	// 测试不同的参数组合
	testCases := []struct {
		url  string
		desc string
	}{
		{"http://localhost:8080/api/v1/clients?page=1&page_size=10", "默认10条"},
		{"http://localhost:8080/api/v1/clients?page=1&page_size=20", "20条"},
		{"http://localhost:8080/api/v1/clients?page=1&page_size=100", "100条"},
		{"http://localhost:8080/api/v1/clients?page=1&page_size=9999", "9999条"},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.desc)
		fmt.Printf("请求URL: %s\n", tc.url)

		req, err := http.NewRequest("GET", tc.url, nil)
		if err != nil {
			fmt.Printf("❌ 创建请求失败: %v\n", err)
			continue
		}

		// 添加认证头（如果需要的话）
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 读取响应失败: %v\n", err)
			continue
		}

		fmt.Printf("状态码: %d\n", resp.StatusCode)

		// 尝试解析JSON响应
		var jsonResponse map[string]interface{}
		if err := json.Unmarshal(body, &jsonResponse); err == nil {
			fmt.Printf("响应结构: %+v\n", getStructureSummary(jsonResponse))

			// 检查data字段
			if data, ok := jsonResponse["data"].(map[string]interface{}); ok {
				if clients, ok := data["clients"].([]interface{}); ok {
					fmt.Printf("客户数量: %d\n", len(clients))
				}
				if pagination, ok := data["pagination"].(map[string]interface{}); ok {
					fmt.Printf("分页信息: 总数=%.0f, 页大小=%.0f\n",
						pagination["total"], pagination["page_size"])
				}
			}
		} else {
			fmt.Printf("响应内容: %s\n", string(body))
		}
	}
}

func getStructureSummary(obj map[string]interface{}) map[string]interface{} {
	summary := make(map[string]interface{})
	for k, v := range obj {
		switch val := v.(type) {
		case map[string]interface{}:
			summary[k] = "object{" + getStructureKeys(val) + "}"
		case []interface{}:
			summary[k] = fmt.Sprintf("array[%d]", len(val))
		default:
			summary[k] = fmt.Sprintf("%T", val)
		}
	}
	return summary
}

func getStructureKeys(obj map[string]interface{}) string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	if len(keys) > 3 {
		return fmt.Sprintf("%s...(%d)", keys[:3], len(keys))
	}
	return fmt.Sprintf("%s", keys)
}