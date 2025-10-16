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

	// 2. 测试API数据格式
	fmt.Println("\n🔍 测试API数据格式...")

	testCases := []struct {
		name        string
		url         string
		expectedKey string
		description string
	}{
		{
			name:        "客户API数据格式",
			url:         "/api/clients?page=1&page_size=5",
			expectedKey: "list",
			description: "检查客户数据返回格式",
		},
		{
			name:        "律师API数据格式",
			url:         "/api/lawfirm/lawyers?pageNum=1&pageSize=5",
			expectedKey: "list",
			description: "检查律师数据返回格式",
		},
	}

	successCount := 0
	totalCount := len(testCases)

	for _, tc := range testCases {
		fmt.Printf("\n📋 测试: %s\n", tc.name)
		fmt.Printf("   说明: %s\n", tc.description)
		fmt.Printf("   URL: %s\n", tc.url)

		// 创建请求
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

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   状态码: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err == nil {
				fmt.Printf("✅ JSON解析成功\n")

				if success, ok := response["success"].(bool); ok && success {
					if data, exists := response["data"]; exists {
						fmt.Printf("📊 data类型: %T\n", data)

						switch v := data.(type) {
						case []interface{}:
							fmt.Printf("   📈 数据数组长度: %d\n", len(v))
							if len(v) > 0 {
								fmt.Printf("   📋 第一个元素: %v\n", v[0])
							}
						case map[string]interface{}:
							fmt.Printf("   🗂️ 数据对象字段: %v\n", getMapKeys(v))
							if list, ok := v["list"]; ok {
								if listArray, ok := list.([]interface{}); ok {
									fmt.Printf("   📈 list数组长度: %d\n", len(listArray))
									if len(listArray) > 0 {
										fmt.Printf("   📋 第一个客户/律师: %v\n", listArray[0])
									}
								}
							}
						}
					}
					successCount++
				} else {
					fmt.Printf("❌ API返回失败: %v\n", response)
				}
			} else {
				fmt.Printf("❌ JSON解析失败: %v\n", err)
			}
		} else {
			fmt.Printf("❌ 状态码错误: %d\n", resp.StatusCode)
			fmt.Printf("   响应内容: %s\n", string(body))
		}
	}

	// 3. 总结
	fmt.Printf("\n🎯 数据格式测试总结:\n")
	fmt.Printf("   成功: %d/%d\n", successCount, totalCount)
	if successCount == totalCount {
		fmt.Printf("✅ 所有API数据格式正常\n")
	} else {
		fmt.Printf("❌ 部分API数据格式有问题\n")
	}
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}