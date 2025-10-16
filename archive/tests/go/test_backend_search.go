package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClientListResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       []Client    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type Client struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   string `json:"status"`
}

type Pagination struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

func testSearchAPI() {
	baseURL := "http://localhost:8080/api/v1" // 根据实际API路径调整

	tests := []struct {
		name string
		url  string
	}{
		{"获取所有客户", baseURL + "/clients?page=1&page_size=10"},
		{"搜索'张三' (name参数)", baseURL + "/clients?page=1&page_size=10&name=张三"},
		{"搜索'张三' (search参数)", baseURL + "/clients?page=1&page_size=10&search=张三"},
		{"筛选个人客户", baseURL + "/clients?page=1&page_size=10&type=个人"},
		{"筛选活跃客户", baseURL + "/clients?page=1&page_size=10&status=active"},
		{"组合搜索: 个人客户+名称包含'张'", baseURL + "/clients?page=1&page_size=10&type=个人&search=张"},
	}

	for _, test := range tests {
		fmt.Printf("\n🔍 测试: %s\n", test.name)
		fmt.Printf("📡 URL: %s\n", test.url)

		resp, err := http.Get(test.url)
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

		fmt.Printf("📊 状态码: %d\n", resp.StatusCode)

		var result ClientListResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("⚠️  JSON解析失败，显示原始响应:\n%s\n", string(body))
			continue
		}

		fmt.Printf("📦 成功: %t\n", result.Success)
		fmt.Printf("📝 消息: %s\n", result.Message)
		fmt.Printf("📋 返回记录数: %d\n", len(result.Data))
		if result.Pagination.Total > 0 {
			fmt.Printf("📈 总记录数: %d\n", result.Pagination.Total)
		}

		// 显示前3条记录
		for i, client := range result.Data {
			if i >= 3 {
				fmt.Printf("   ... 还有 %d 条记录\n", len(result.Data)-3)
				break
			}
			fmt.Printf("   - %s (%s) - %s\n", client.Name, client.Type, client.Status)
		}

		if len(result.Data) == 0 {
			fmt.Printf("   ⚠️  没有找到匹配的记录\n")
		}

		// 短暂延迟避免请求过快
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	fmt.Println("🚀 开始测试客户搜索API")
	fmt.Println("=====================================")

	// 等待服务器启动
	fmt.Println("⏳ 等待2秒确保服务器就绪...")
	time.Sleep(2 * time.Second)

	testSearchAPI()

	fmt.Println("\n✅ 测试完成")
}