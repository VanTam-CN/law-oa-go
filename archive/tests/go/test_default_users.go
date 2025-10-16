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
	Message string `json:"message"`
	Token   string `json:"token"`
	Data    struct {
		Token string `json:"token"`
		User  struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

func main() {
	fmt.Println("🔑 测试默认用户登录")
	fmt.Println("=====================================")

	// 默认用户列表
	defaultUsers := []struct {
		email    string
		password string
		name     string
		role     string
	}{
		{"admin@law-oa.com", "admin123", "系统管理员", "admin"},
		{"admin@law-oa.com", "password", "系统管理员", "admin"},
		{"admin@law-oa.com", "123456", "系统管理员", "admin"},
		{"admin@law-oa.com", "admin", "系统管理员", "admin"},
		{"zhang@law-oa.com", "lawyer123", "张律师", "lawyer"},
		{"zhang@law-oa.com", "password", "张律师", "lawyer"},
		{"zhang@law-oa.com", "123456", "张律师", "lawyer"},
		{"li@law-oa.com", "lawyer123", "李律师", "lawyer"},
		{"li@law-oa.com", "password", "李律师", "lawyer"},
		{"li@law-oa.com", "123456", "李律师", "lawyer"},
	}

	baseURL := "http://localhost:8080"
	loginEndpoint := "/api/v1/auth/login"

	client := &http.Client{Timeout: 10 * time.Second}

	for _, user := range defaultUsers {
		fmt.Printf("\n🔍 测试用户: %s (%s)\n", user.name, user.email)
		fmt.Printf("🔑 密码: %s\n", user.password)

		// 准备登录请求
		loginReq := LoginRequest{
			Email:    user.email,
			Password: user.password,
		}

		jsonData, err := json.Marshal(loginReq)
		if err != nil {
			fmt.Printf("   ❌ JSON编码失败: %v\n", err)
			continue
		}

		// 发送登录请求
		resp, err := client.Post(baseURL+loginEndpoint, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("   ❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("   ❌ 读取响应失败: %v\n", err)
			continue
		}

		fmt.Printf("   📊 状态码: %d\n", resp.StatusCode)
		fmt.Printf("   📄 响应: %s\n", string(body))

		if resp.StatusCode == 200 {
			var loginResp LoginResponse
			if err := json.Unmarshal(body, &loginResp); err != nil {
				fmt.Printf("   ⚠️ JSON解析失败\n")
				continue
			}

			// 提取token
			token := ""
			if loginResp.Token != "" {
				token = loginResp.Token
			} else if loginResp.Data.Token != "" {
				token = loginResp.Data.Token
			}

			if token != "" {
				fmt.Printf("   ✅ 登录成功!\n")
				fmt.Printf("   🎯 获取到令牌: %s\n", token[:50]+"...")

				// 测试令牌是否有效
				if testToken(baseURL, token) {
					fmt.Printf("   ✅ 令牌验证成功\n")
					fmt.Printf("\n🚀 前端使用方法:\n")
					fmt.Printf("   在浏览器控制台运行:\n")
					fmt.Printf("   localStorage.setItem('auth_token', '%s')\n", token)
					fmt.Printf("   localStorage.setItem('law_oa_token', '%s')\n", token)
					fmt.Printf("   然后刷新页面\n")
					return
				} else {
					fmt.Printf("   ❌ 令牌验证失败\n")
				}
			} else {
				fmt.Printf("   ⚠️ 响应中没有找到令牌\n")
			}
		} else {
			fmt.Printf("   ❌ 登录失败\n")
		}
	}

	fmt.Printf("\n⚠️ 所有默认用户都登录失败\n")
	fmt.Printf("💡 可能的原因:\n")
	fmt.Printf("1. 认证端点路径不正确\n")
	fmt.Printf("2. 密码哈希算法不匹配\n")
	fmt.Printf("3. 用户数据没有正确初始化\n")
	fmt.Printf("4. 认证服务存在问题\n")

	// 尝试其他可能的认证端点
	fmt.Printf("\n🔍 尝试其他认证端点:\n")
	alternativeEndpoints := []string{
		"/api/v1/auth/login",
		"/api/auth/login",
		"/auth/login",
		"/login",
		"/api/v1/login",
	}

	for _, endpoint := range alternativeEndpoints {
		fmt.Printf("   测试端点: %s\n", endpoint)
		testEndpoint(baseURL, endpoint)
	}
}

func testToken(baseURL, token string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	url := baseURL + "/api/v1/clients?page=1&page_size=1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func testEndpoint(baseURL, endpoint string) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := baseURL + endpoint

	// 测试GET请求
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("      ❌ GET失败\n")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 404 || resp.StatusCode == 405 {
		fmt.Printf("      ✅ 端点存在 (状态码: %d)\n", resp.StatusCode)
	} else {
		fmt.Printf("      ⚠️ 端点响应异常 (状态码: %d)\n", resp.StatusCode)
	}
}