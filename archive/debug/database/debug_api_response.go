package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 调试API响应结构")
	fmt.Println("=====================================")

	authToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNSwidXNlcm5hbWUiOiJzZWFyY2h0ZXN0QGxhdy1vYS5jb20iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NjA0MjI3NTksImlhdCI6MTc2MDMzNjM1OX0.XI56TAE0kporkOUEh2wKk7k0F8HWGQdPQcwT58q8yCQ"

	// 测试API响应结构
	testURLs := []string{
		"http://localhost:8080/api/v1/clients?search=张三&page=1&page_size=1",
		"http://localhost:8080/api/v1/clients?page=1&page_size=1",
		"http://localhost:8080/api/v1/clients?type=个人&page=1&page_size=1",
	}

	for i, url := range testURLs {
		fmt.Printf("\n🧪 测试 %d: %s\n", i+1, url)

		response, err := makeDetailedRequest(url, authToken)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}

		fmt.Printf("✅ 请求成功\n")
		fmt.Printf("📊 状态码: %d\n", response.StatusCode)
		fmt.Printf("📄 响应内容:\n")

		// 格式化JSON输出
		if response.FormattedJSON != "" {
			fmt.Println(response.FormattedJSON)
		} else {
			fmt.Printf("   无法格式化JSON\n")
		}

		// 分析响应结构
		analyzeResponseStructure(response.RawJSON)
	}

	fmt.Println("\n🔍 结构分析完成")
}

type APIResponse struct {
	StatusCode int
	RawJSON    string
	Headers    map[string][]string
	FormattedJSON string
}

func makeDetailedRequest(url, token string) (*APIResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 提取响应头
	headers := make(map[string][]string)
	for key, values := range resp.Header {
		headers[key] = values
	}

	// 尝试格式化JSON
	var formattedJSON string
	if json.Valid(body) {
		var prettyJSON interface{}
		if err := json.Unmarshal(body, &prettyJSON); err == nil {
			prettyBytes, err := json.MarshalIndent(prettyJSON, "   ", "  ")
			if err == nil {
				formattedJSON = string(prettyBytes)
			}
		}
	}

	return &APIResponse{
		StatusCode:    resp.StatusCode,
		RawJSON:       string(body),
		Headers:       headers,
		FormattedJSON: formattedJSON,
	}, nil
}

func analyzeResponseStructure(rawJSON string) {
	fmt.Println("   📋 结构分析:")

	// 解析JSON到interface{}
	var data interface{}
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		fmt.Printf("   ❌ JSON解析失败: %v\n", err)
		return
	}

	// 使用类型断言分析结构
	switch v := data.(type) {
	case map[string]interface{}:
		fmt.Printf("   📝 根对象字段: %d个\n", len(v))

		// 检查常见字段
		if success, ok := v["success"].(bool); ok {
			fmt.Printf("   ✅ success字段: %v\n", success)
		}

		if dataField, ok := v["data"]; ok {
			if dataArray, ok := dataField.([]interface{}); ok {
				fmt.Printf("   ✅ data字段是数组，长度: %d\n", len(dataArray))

				if len(dataArray) > 0 {
					if firstItem, ok := dataArray[0].(map[string]interface{}); ok {
						fmt.Printf("   📋 数组第一项字段: %d个\n", len(firstItem))

						// 检查客户相关字段
						if id, ok := firstItem["id"]; ok {
							fmt.Printf("   ✅ id字段: %v\n", id)
						}
						if name, ok := firstItem["name"]; ok {
							fmt.Printf("   ✅ name字段: %v\n", name)
						} else {
							fmt.Printf("   ❌ 缺少name字段\n")
						}
						if typeField, ok := firstItem["type"]; ok {
							fmt.Printf("   ✅ type字段: %v\n", typeField)
						} else {
							fmt.Printf("   ❌ 缺少type字段\n")
						}
						if email, ok := firstItem["email"]; ok {
							fmt.Printf("   ✅ email字段: %v\n", email)
						}

						// 显示所有字段
						fmt.Printf("   📋 所有字段: ")
						keys := make([]string, 0, len(firstItem))
						for key := range firstItem {
							keys = append(keys, key)
						}
						fmt.Printf("%v\n", keys)
					}
				}
			} else {
				fmt.Printf("   ⚠️ data字段不是数组: %T\n", dataField)
			}
		}

		if pagination, ok := v["pagination"].(map[string]interface{}); ok {
			fmt.Printf("   ✅ pagination字段存在\n")
			if total, ok := pagination["total"]; ok {
				fmt.Printf("   📊 pagination.total: %v\n", total)
			}
		}

	default:
		fmt.Printf("   ⚠️ 根对象不是map: %T\n", v)
	}
}