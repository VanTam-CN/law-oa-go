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
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Token   string `json:"token,omitempty"`
}

func main() {
	fmt.Println("🔧 测试客户API接口...")

	// 尝试登录获取token
	loginData := LoginRequest{
		Email:    "test@client.com",
		Password: "test123",
	}

	loginJSON, _ := json.Marshal(loginData)
	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		fmt.Println("❌ 登录失败:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	token := loginResp.Token
	if token == "" {
		token = loginResp.Data.Token
	}

	if !loginResp.Success || token == "" {
		fmt.Println("❌ 登录失败，无法获取token")
		fmt.Println("响应:", string(body))
		return
	}

	fmt.Println("✅ 登录成功，获取到token")

	// 使用token访问客户API
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "http://localhost:8080/api/clients?page=1&page_size=3", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp2, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ 客户API调用失败:", err)
		return
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)

	fmt.Println("📋 客户API响应:")
	fmt.Println(string(body2))

	// 尝试解析响应
	var apiResponse map[string]interface{}
	json.Unmarshal(body2, &apiResponse)

	fmt.Println("\n🔍 响应结构分析:")
	for key, value := range apiResponse {
		fmt.Printf("- %s: %v\n", key, value)
	}

	// 检查是否有data字段
	if data, ok := apiResponse["data"]; ok {
		fmt.Println("\n✅ 找到data字段")
		if dataArray, ok := data.([]interface{}); ok {
			fmt.Printf("📊 data字段包含 %d 条记录\n", len(dataArray))
			if len(dataArray) > 0 {
				fmt.Println("📋 第一条客户记录结构:")
				if firstClient, ok := dataArray[0].(map[string]interface{}); ok {
					for key, value := range firstClient {
						fmt.Printf("  - %s: %v\n", key, value)
					}
				}
			}
		}
	}
}