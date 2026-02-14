//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func main() {
	baseURL := "http://localhost:8080"

	// 1. 登录获取token
	fmt.Println("🔐 登录获取token...")
	loginReq := LoginRequest{
		Email:    "zhangwei@jinchenglaw.com",
		Password: "law123456",
	}

	loginData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	if !loginResp.Success || loginResp.Data.Token == "" {
		fmt.Printf("❌ 登录失败: %s\n", string(body))
		return
	}

	token := loginResp.Data.Token
	fmt.Printf("✅ 登录成功，获取token: %s...\n", token[:20])

	// 2. 测试客户API - 使用不同的参数组合
	fmt.Println("\n🔍 测试客户API...")

	testCases := []struct {
		name     string
		url      string
		expected int
	}{
		{
			name:     "基本客户列表",
			url:      "/api/clients?page=1&page_size=10",
			expected: 200,
		},
		{
			name:     "大页面大小客户列表",
			url:      "/api/clients?page=1&page_size=1000",
			expected: 200,
		},
		{
			name:     "带搜索参数的客户列表",
			url:      "/api/clients?page=1&page_size=1000&search=",
			expected: 200,
		},
		{
			name:     "使用旧式参数名（可能导致422）",
			url:      "/api/clients?pageNum=1&pageSize=1000",
			expected: 422,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\n📋 测试: %s\n", tc.name)
		fmt.Printf("   URL: %s\n", tc.url)

		// 创建新的请求
		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest("GET", baseURL+tc.url, nil)
		if err != nil {
			fmt.Printf("❌ 创建请求失败: %v\n", err)
			continue
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ = io.ReadAll(resp.Body)
		fmt.Printf("   状态码: %d\n", resp.StatusCode)

		if resp.StatusCode == tc.expected {
			fmt.Printf("✅ 测试通过 - 状态码符合预期\n")
		} else {
			fmt.Printf("❌ 测试失败 - 预期状态码 %d，实际 %d\n", tc.expected, resp.StatusCode)
			fmt.Printf("   响应内容: %s\n", string(body))
		}

		// 如果成功，尝试解析响应
		if resp.StatusCode == 200 {
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err == nil {
				if success, ok := response["success"].(bool); ok && success {
					if data, exists := response["data"]; exists {
						if dataMap, ok := data.(map[string]interface{}); ok {
							if list, ok := dataMap["list"]; ok {
								if listArray, ok := list.([]interface{}); ok {
									fmt.Printf("   📊 获取到 %d 个客户\n", len(listArray))
								}
							}
						}
					}
				}
			}
		}
	}
}