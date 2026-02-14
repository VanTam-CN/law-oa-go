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

	// 2. 测试CreateCaseWizard组件需要的API
	fmt.Println("\n🔍 测试新建案件向导所需的API...")

	testCases := []struct {
		name        string
		url         string
		description string
		expected    int
	}{
		{
			name:        "客户列表API",
			url:         "/api/clients?page=1&page_size=100",
			description: "新建案件所需的委托人数据",
			expected:    200,
		},
		{
			name:        "律师列表API",
			url:         "/api/lawfirm/lawyers?pageNum=1&pageSize=100",
			description: "新建案件所需的律师数据",
			expected:    200,
		},
		{
			name:        "利益冲突检查API",
			url:         "/api/conflict/check",
			description: "新建案件的利益冲突检测功能",
			expected:    200,
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

		// 对于利益冲突检查，需要POST请求和数据
		if tc.name == "利益冲突检查API" {
			conflictData := map[string]interface{}{
				"clientId":                 "1",
				"clientName":               "测试客户",
				"clientType":               "个人",
				"caseName":                 "测试案件",
				"caseType":                 "民事诉讼",
				"otherParties":             []string{"测试对方"},
				"userId":                   "1",
				"searchYears":              5,
				"searchDepth":              "DEEP",
				"includeCorporateRelations": true,
				"requestTime":              time.Now(),
			}

			conflictJson, _ := json.Marshal(conflictData)
			req, err = http.NewRequest("POST", baseURL+tc.url, bytes.NewBuffer(conflictJson))
			if err != nil {
				fmt.Printf("❌ 创建POST请求失败: %v\n", err)
				continue
			}
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   状态码: %d\n", resp.StatusCode)

		if resp.StatusCode == tc.expected {
			fmt.Printf("✅ 测试通过\n")
			successCount++

			// 如果成功，尝试解析响应以获取数据数量
			if resp.StatusCode == 200 {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err == nil {
					if success, ok := response["success"].(bool); ok && success {
						if data, exists := response["data"]; exists {
							switch v := data.(type) {
							case []interface{}:
								fmt.Printf("   📊 获取到 %d 条数据\n", len(v))
							case map[string]interface{}:
								if list, ok := v["list"]; ok {
									if listArray, ok := list.([]interface{}); ok {
										fmt.Printf("   📊 获取到 %d 条数据\n", len(listArray))
									}
								}
								if tc.name == "利益冲突检查API" {
									fmt.Printf("   🎯 利益冲突检查功能正常工作\n")
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
		fmt.Printf("✅ 所有测试通过！新建案件向导的API修复成功。\n")
		fmt.Printf("✅ 委托人和律师下拉选项现在应该能正常显示。\n")
		fmt.Printf("✅ 利益冲突检查功能正常工作。\n")
	} else {
		fmt.Printf("❌ 部分测试失败，需要进一步调试。\n")
	}
}