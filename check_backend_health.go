package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 检查后端服务健康状态...")

	client := &http.Client{Timeout: 5 * time.Second}

	// 检查基础健康端点
	endpoints := []string{
		"http://localhost:8080/health",
		"http://localhost:8080/api/v1/health",
		"http://localhost:8080/api/v1/conflict/health",
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\n📡 检查端点: %s\n", endpoint)

		resp, err := client.Get(endpoint)
		if err != nil {
			fmt.Printf("❌ 连接失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 读取响应失败: %v\n", err)
			continue
		}

		fmt.Printf("✅ 状态码: %d\n", resp.StatusCode)
		fmt.Printf("📋 响应: %s\n", string(body))
	}

	fmt.Println("\n💡 如果所有端点都连接失败，请启动后端服务:")
	fmt.Println("   1. 编译: go build -o main main.go")
	fmt.Println("   2. 运行: ./main")
	fmt.Println("   3. 或者直接运行: go run main.go")
}
