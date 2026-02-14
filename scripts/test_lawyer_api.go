//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	User      interface{} `json:"user"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	baseURL := "http://localhost:8080/api/v1"

	// 1. 登录获取JWT令牌
	loginReq := LoginRequest{
		Email:    "admin@lawoa.com",
		Password: "admin123",
	}

	loginData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		log.Fatal("登录请求失败:", err)
	}
	defer resp.Body.Close()

	var loginResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		log.Fatal("解析登录响应失败:", err)
	}

	if loginResp.Code != 200 {
		log.Fatal("登录失败:", loginResp.Message)
	}

	token := loginResp.Data.(map[string]interface{})["token"].(string)
	fmt.Println("✅ 登录成功，获取到JWT令牌")

	// 2. 测试律师API
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/lawfirm/lawyers?page=1&page_size=5", nil)
	if err != nil {
		log.Fatal("创建律师API请求失败:", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		log.Fatal("律师API请求失败:", err)
	}
	defer resp.Body.Close()

	var lawyerResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&lawyerResp); err != nil {
		log.Fatal("解析律师API响应失败:", err)
	}

	fmt.Printf("律师API响应状态: %d\n", resp.StatusCode)
	fmt.Printf("律师API响应: %+v\n", lawyerResp)

	if lawyerResp.Code == 200 {
		fmt.Println("✅ 律师API测试成功")
		if lawyers, ok := lawyerResp.Data.([]interface{}); ok {
			fmt.Printf("获取到 %d 个律师:\n", len(lawyers))
			for i, lawyer := range lawyers {
				if lawyerMap, ok := lawyer.(map[string]interface{}); ok {
					fmt.Printf("%d. %s (ID: %.0f, 邮箱: %s)\n",
						i+1,
						lawyerMap["name"],
						lawyerMap["id"],
						lawyerMap["email"])
				}
			}
		}
	} else {
		fmt.Printf("❌ 律师API测试失败: %s\n", lawyerResp.Message)
	}
}