package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ClientListResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       []Client    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Client struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type Pagination struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

func main() {
	fmt.Println("🔍 带认证的客户搜索API测试")
	fmt.Println("=====================================")

	// 使用测试token
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYwMjUwODQ2LCJpYXQiOjE3NjAxNjQ0NDZ9.4N-Gj2OCUQQRb_sAh1lxxdGyROfn591sFCQ_kNRSOtc"

	baseURL := "http://localhost:8080/api/v1"

	tests := []struct {
		name string
		url  string
	}{
		{"获取所有客户", baseURL + "/clients?page=1&page_size=20"},
		{"搜索'张三'(name参数)", baseURL + "/clients?page=1&page_size=20&name=张三"},
		{"搜索'张三'(search参数)", baseURL + "/clients?page=1&page_size=20&search=张三"},
		{"筛选个人客户", baseURL + "/clients?page=1&page_size=20&type=个人"},
		{"筛选活跃客户", baseURL + "/clients?page=1&page_size=20&status=active"},
		{"组合搜索: 个人+搜索'张'", baseURL + "/clients?page=1&page_size=20&type=个人&search=张"},
		{"精确搜索'张三'(不分页)", baseURL + "/clients?name=张三&page_size=100"},
	}

	for _, test := range tests {
		fmt.Printf("\n🔍 测试: %s\n", test.name)
		fmt.Printf("📡 URL: %s\n", test.url)

		result, err := makeAuthenticatedRequest(test.url, testToken)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}

		analyzeResult(result, test.name)
		time.Sleep(100 * time.Millisecond)
	}
}

func makeAuthenticatedRequest(url, token string) (*ClientListResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置认证头
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

	fmt.Printf("📊 状态码: %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		fmt.Printf("⚠️ API返回错误状态: %d\n", resp.StatusCode)
		fmt.Printf("📄 响应体: %s\n", string(body))
		return nil, fmt.Errorf("API返回状态码: %d", resp.StatusCode)
	}

	var result ClientListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("⚠️ JSON解析失败，显示原始响应:\n%s\n", string(body))
		return nil, err
	}

	return &result, nil
}

func analyzeResult(result *ClientListResponse, testName string) {
	fmt.Printf("📦 成功: %t\n", result.Success)
	fmt.Printf("📝 消息: %s\n", result.Message)
	fmt.Printf("📋 返回记录数: %d\n", len(result.Data))

	if result.Pagination.Total > 0 {
		fmt.Printf("📈 总记录数: %d\n", result.Pagination.Total)
	}

	// 分析搜索结果
	if strings.Contains(testName, "张三") {
		if len(result.Data) == 1 && result.Data[0].Name == "张三" {
			fmt.Printf("✅ 搜索成功: 精确找到'张三'\n")
		} else if len(result.Data) == 0 {
			fmt.Printf("❌ 搜索失败: 没有找到'张三'\n")
		} else {
			fmt.Printf("⚠️ 搜索异常: 找到%d条记录，期望1条\n", len(result.Data))
		}
	}

	// 显示记录详情
	if len(result.Data) > 0 {
		fmt.Printf("📄 客户记录:\n")
		maxShow := 3
		if len(result.Data) < maxShow {
			maxShow = len(result.Data)
		}

		for i := 0; i < maxShow; i++ {
			client := result.Data[i]
			fmt.Printf("   %d. ID:%d, 姓名:'%s', 类型:%s, 状态:%s\n",
				i+1, client.ID, client.Name, client.Type, client.Status)
		}

		if len(result.Data) > maxShow {
			fmt.Printf("   ... 还有 %d 条记录\n", len(result.Data)-maxShow)
		}
	} else {
		fmt.Printf("   ⚠️ 没有找到匹配的记录\n")
	}

	// 检查响应格式
	if !result.Success {
		fmt.Printf("❌ API返回失败状态\n")
	}
}

// 额外的测试函数
func testSearchAccuracy() {
	fmt.Println("\n🎯 搜索精度测试")
	fmt.Println("=====================================")

	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0LCJ1c2VybmFtZSI6ImFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzYwMjUwODQ2LCJpYXQiOjE3NjAxNjQ0NDZ9.4N-Gj2OCUQQRb_sAh1lxxdGyROfn591sFCQ_kNRSOtc"
	baseURL := "http://localhost:8080/api/v1"

	searchTests := []struct {
		query    string
		expected int
	}{
		{"name=张三", 1},      // 精确匹配
		{"search=张三", 1},    // 搜索匹配
		{"name=张", 1},        // 前缀匹配
		{"search=张", 1},      // 搜索包含
		{"name=不存在", 0},    // 不存在
		{"search=不存在", 0},  // 不存在搜索
	}

	for _, test := range searchTests {
		url := fmt.Sprintf("%s/clients?%s&page_size=100", baseURL, test.query)
		fmt.Printf("\n🔍 测试: %s (期望%d条)\n", test.query, test.expected)

		result, err := makeAuthenticatedRequest(url, testToken)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}

		actual := len(result.Data)
		if actual == test.expected {
			fmt.Printf("✅ 测试通过: 实际%d条\n", actual)
		} else {
			fmt.Printf("❌ 测试失败: 期望%d条，实际%d条\n", test.expected, actual)

			// 显示找到的记录
			if actual > 0 {
				fmt.Printf("   找到的记录:\n")
				for _, client := range result.Data {
					fmt.Printf("   - %s (%s)\n", client.Name, client.Type)
				}
			}
		}
	}
}