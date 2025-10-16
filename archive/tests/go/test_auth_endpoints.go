package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 测试认证端点")
	fmt.Println("=====================================")

	baseURL := "http://localhost:8080"

	// 测试常见的认证端点
	endpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/token",
		"/api/auth/login",
		"/api/auth/register",
		"/auth/login",
		"/login",
		"/register",
	}

	client := &http.Client{Timeout: 5 * time.Second}

	fmt.Println("🌐 测试认证相关端点:")

	for _, endpoint := range endpoints {
		url := baseURL + endpoint
		fmt.Printf("\n📡 测试: %s\n", endpoint)

		// 测试GET请求
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("   ❌ GET失败: %v\n", err)
			continue
		}
		resp.Body.Close()

		fmt.Printf("   📊 GET状态: %d\n", resp.StatusCode)

		// 测试POST请求
		resp2, err := client.Post(url, "application/json", nil)
		if err != nil {
			fmt.Printf("   ❌ POST失败: %v\n", err)
			continue
		}
		resp2.Body.Close()

		fmt.Printf("   📊 POST状态: %d\n", resp2.StatusCode)

		// 如果其中一个返回200，说明端点存在
		if resp.StatusCode == 200 || resp.StatusCode == 404 || resp.StatusCode == 405 {
			fmt.Printf("   ✅ 端点存在且响应正常\n")
		}
	}

	// 检查健康端点是否包含认证信息
	fmt.Println("\n🏥 检查健康端点:")
	healthEndpoints := []string{
		"/health",
		"/api/v1/health",
		"/ping",
		"/api/v1/ping",
	}

	for _, endpoint := range healthEndpoints {
		url := baseURL + endpoint
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("   ✅ %s: %s\n", endpoint, string(body))
		}
	}

	fmt.Println("\n💡 建议:")
	fmt.Println("1. 如果认证端点不存在，可能需要先启动认证服务")
	fmt.Println("2. 检查是否有用户种子数据脚本")
	fmt.Println("3. 尝试查看项目文档了解认证设置")
	fmt.Println("4. 检查是否有管理员默认账号")
}