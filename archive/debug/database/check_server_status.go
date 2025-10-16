package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🔍 检查服务器状态")
	fmt.Println("=====================================")

	// 检查不同端口
	ports := []int{8080, 3003, 8081, 8000, 9000}

	for _, port := range ports {
		testPort(port)
	}

	// 检查常见的API路径
	testAPIPaths()
}

func testPort(port int) {
	url := fmt.Sprintf("http://localhost:%d", port)
	client := &http.Client{Timeout: 2 * time.Second}

	fmt.Printf("🌐 测试端口 %d:\n", port)

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("   ❌ 连接失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("   ✅ 连接成功 - 状态: %d\n", resp.StatusCode)

	// 读取响应体
	buf := make([]byte, 100)
	n, _ := resp.Body.Read(buf)
	if n > 0 {
		fmt.Printf("   📄 响应预览: %s\n", string(buf[:n]))
	}
}

func testAPIPaths() {
	fmt.Println("\n🛣️ 测试常见API路径:")

	baseURLs := []string{
		"http://localhost:8080",
		"http://localhost:8080/api",
		"http://localhost:8080/api/v1",
	}

	paths := []string{
		"/",
		"/clients",
		"/health",
		"/ping",
		"/api/health",
	}

	for _, baseURL := range baseURLs {
		for _, path := range paths {
			fullURL := baseURL + path
			testPath(fullURL)
		}
	}
}

func testPath(url string) {
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return // 静默失败，避免太多输出
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 404 {
		fmt.Printf("   ✅ %s - %d\n", url, resp.StatusCode)
	}
}