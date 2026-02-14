//go:build ignore

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

	fmt.Println("🔍 客户搜索功能详细测试")
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

	// 2. 获取所有客户数据，看看实际的客户名称
	fmt.Println("\n📋 获取所有客户数据:")
	allClients := getAllClients(client, baseURL, token)
	for i, client := range allClients {
		fmt.Printf("%d. ID: %v, Name: '%s', Type: '%s', Status: '%s'\n",
			i+1, client["id"], client["name"], client["type"], client["status"])
	}

	// 3. 测试详细搜索
	testSearchTerms := []string{
		"ABC",
		"ABC科技",
		"科技有限公司",
		"XYZ",
		"软件公司",
		"测试-ABC",
		"测试-XYZ",
	}

	for _, searchTerm := range testSearchTerms {
		fmt.Printf("\n🔍 测试搜索: '%s'\n", searchTerm)
		searchResults := searchClients(client, baseURL, token, searchTerm)

		if len(searchResults) > 0 {
			fmt.Printf("✅ 找到 %d 条记录:\n", len(searchResults))
			for _, result := range searchResults {
				if name, ok := result["name"].(string); ok {
					fmt.Printf("   - %s\n", name)
				}
			}
		} else {
			fmt.Printf("❌ 没有找到记录\n")
		}
	}

	fmt.Println("\n✅ 详细搜索测试完成！")
}

func getAuthToken(client *http.Client, baseURL string) string {
	loginURL := baseURL + "/api/auth/login"

	loginRequest := LoginRequest{
		Email:    "admin@lawoa.com",
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

func getAllClients(client *http.Client, baseURL, token string) []map[string]interface{} {
	url := baseURL + "/api/v1/clients"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return nil
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 响应读取失败: %v\n", err)
		return nil
	}

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if success, ok := result["success"].(bool); ok && success {
				if data, ok := result["data"].([]interface{}); ok {
					clients := make([]map[string]interface{}, len(data))
					for i, client := range data {
						if clientMap, ok := client.(map[string]interface{}); ok {
							clients[i] = clientMap
						}
					}
					return clients
				}
			}
		}
	}

	fmt.Printf("❌ 获取客户数据失败，状态码: %d\n", resp.StatusCode)
	return nil
}

func searchClients(client *http.Client, baseURL, token, searchTerm string) []map[string]interface{} {
	url := baseURL + "/api/v1/clients?search=" + searchTerm

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ 创建搜索请求失败: %v\n", err)
		return nil
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 搜索请求失败: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 搜索响应读取失败: %v\n", err)
		return nil
	}

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if success, ok := result["success"].(bool); ok && success {
				if data, ok := result["data"].([]interface{}); ok {
					clients := make([]map[string]interface{}, len(data))
					for i, client := range data {
						if clientMap, ok := client.(map[string]interface{}); ok {
							clients[i] = clientMap
						}
					}
					return clients
				}
			}
		}
	}

	fmt.Printf("❌ 搜索失败，状态码: %d\n", resp.StatusCode)
	return nil
}