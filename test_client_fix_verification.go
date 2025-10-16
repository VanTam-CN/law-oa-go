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

	// 2. 测试修复后的客户API - 使用pageSize=100（符合后端验证规则）
	fmt.Println("\n🔍 测试修复后的客户API...")

	testCases := []struct {
		name        string
		url         string
		expected    int
		description string
	}{
		{
			name:        "修复后的客户列表请求",
			url:         "/api/clients?page=1&page_size=100",
			expected:    200,
			description: "使用符合后端验证的pageSize=100",
		},
		{
			name:        "空搜索参数测试",
			url:         "/api/clients?page=1&page_size=100&search=",
			expected:    200,
			description: "测试空搜索参数是否能返回所有客户",
		},
		{
			name:        "默认参数测试",
			url:         "/api/clients?page=1&page_size=10",
			expected:    200,
			description: "测试默认参数是否正常工作",
		},
	}

	successCount := 0
	totalCount := len(testCases)

	for _, tc := range testCases {
		fmt.Printf("\n📋 测试: %s\n", tc.name)
		fmt.Printf("   说明: %s\n", tc.description)
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
			fmt.Printf("✅ 测试通过\n")
			successCount++

			// 如果成功，尝试解析响应以获取客户数量
			if resp.StatusCode == 200 {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err == nil {
					if success, ok := response["success"].(bool); ok && success {
						if data, exists := response["data"]; exists {
							switch v := data.(type) {
							case []interface{}:
								fmt.Printf("   📊 获取到 %d 个客户\n", len(v))
							case map[string]interface{}:
								if list, ok := v["list"]; ok {
									if listArray, ok := list.([]interface{}); ok {
										fmt.Printf("   📊 获取到 %d 个客户\n", len(listArray))
									}
								}
							}
						}
					}
				}
			}
		} else {
			fmt.Printf("❌ 测试失败 - 预期状态码 %d，实际 %d\n", tc.expected, resp.StatusCode)
			fmt.Printf("   响应内容: %s\n", string(body))
		}
	}

	// 3. 总结测试结果
	fmt.Printf("\n🎯 测试总结:\n")
	fmt.Printf("   成功: %d/%d\n", successCount, totalCount)
	if successCount == totalCount {
		fmt.Printf("✅ 所有测试通过！客户API修复成功。\n")
		fmt.Printf("✅ 新建案件的委托人选择功能现在应该能正常工作。\n")
	} else {
		fmt.Printf("❌ 部分测试失败，需要进一步调试。\n")
	}
}