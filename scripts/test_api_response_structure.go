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

	// 2. 测试完整API响应格式
	fmt.Println("\n🔍 测试完整API响应格式...")

	testCases := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "客户API完整响应",
			url:         "/api/clients?page=1&page_size=5",
			description: "检查客户API的完整响应结构",
		},
		{
			name:        "律师API完整响应",
			url:         "/api/lawfirm/lawyers?pageNum=1&pageSize=5",
			description: "检查律师API的完整响应结构",
		},
	}

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
		fmt.Printf("   完整响应: %s\n", string(body))
	}
}