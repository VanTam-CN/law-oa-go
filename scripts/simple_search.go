package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	fmt.Println("🔍 客户搜索功能简单测试")
	fmt.Println("=======================")

	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 10 * time.Second}

	// 1. 登录获取认证令牌
	fmt.Println("\n🔑 获取认证令牌...")
	token := getAuthToken(client, baseURL)
	if token == "" {
		fmt.Println("❌ 获取认证令牌失败，无法继续测试")
		return
	}
	fmt.Printf("✅ 获取认证令牌成功\n")

	// 2. 测试客户搜索功能
	testSearch(client, baseURL, token)

	fmt.Println("\n✅ 搜索功能测试完成！")
}

func getAuthToken(client *http.Client, baseURL string) string {
	loginURL := baseURL + "/api/auth/login"

	loginRequest := LoginRequest{
		Email:    "admin@lawoa.com", // 使用正确的管理员邮箱
		Password: "admin123",
	}

	jsonData, err := json.Marshal(loginRequest)
	if err != nil {
		fmt.Printf("❌ 登录请求编码失败: %v\n", err)
		return ""
	}

	resp, err := client.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 登录响应读取失败: %v\n", err)
		return ""
	}

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if data, ok := result["data"].(map[string]interface{}); ok {
				if token, ok := data["token"].(string); ok {
					return token
				}
			}
		}
	}

	fmt.Printf("❌ 登录失败，状态码: %d\n", resp.StatusCode)
	fmt.Printf("   响应内容: %s\n", string(body))
	return ""
}

func testSearch(client *http.Client, baseURL, token string) {
	fmt.Println("\n📋 测试客户搜索功能")

	searchTerms := []string{
		"ABC",
		"科技",
		"测试",
		"有限公司",
	}

	for i, searchTerm := range searchTerms {
		fmt.Printf("\n🔍 测试 %d: 搜索 '%s'\n", i+1, searchTerm)

		// 构建请求URL
		url := baseURL + "/api/v1/clients?search=" + searchTerm

		// 创建请求
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("❌ 创建请求失败: %v\n", err)
			continue
		}

		// 添加认证头
		req.Header.Set("Authorization", "Bearer "+token)

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ 响应读取失败: %v\n", err)
			continue
		}

		fmt.Printf("📥 响应状态: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err == nil {
				if success, ok := result["success"].(bool); ok && success {
					if data, ok := result["data"].([]interface{}); ok {
						fmt.Printf("✅ 搜索成功，找到 %d 条记录\n", len(data))

						// 显示前3条结果
						for j := 0; j < len(data) && j < 3; j++ {
							if client, ok := data[j].(map[string]interface{}); ok {
								if name, ok := client["name"].(string); ok {
									fmt.Printf("   - %s\n", name)
								}
							}
						}
						if len(data) > 3 {
							fmt.Printf("   ... (还有 %d 条记录)\n", len(data)-3)
						}
					}
				} else {
					fmt.Printf("❌ 响应表示失败: %+v\n", result)
				}
			} else {
				fmt.Printf("❌ 响应解析失败: %v\n", err)
			}
		} else {
			fmt.Printf("❌ API调用失败，状态码: %d\n", resp.StatusCode)
			if resp.StatusCode != 401 {
				fmt.Printf("   响应内容: %s\n", string(body))
			}
		}
	}

	// 测试无搜索词（获取所有客户）
	fmt.Println("\n🔍 测试: 获取所有客户")
	url := baseURL + "/api/v1/clients"

	req, err := http.NewRequest("GET", url, nil)
	if err == nil {
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode == 200 {
				var result map[string]interface{}
				if json.Unmarshal(body, &result) == nil {
					if success, ok := result["success"].(bool); ok && success {
						if data, ok := result["data"].([]interface{}); ok {
							fmt.Printf("✅ 获取所有客户成功，共 %d 条记录\n", len(data))
						}
					}
				}
			} else {
				fmt.Printf("❌ 获取所有客户失败，状态码: %d\n", resp.StatusCode)
			}
		}
	}
}