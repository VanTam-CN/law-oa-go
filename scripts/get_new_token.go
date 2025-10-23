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
	User    struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

func main() {
	fmt.Println("🔑 获取新的认证令牌")
	fmt.Println("=====================================")

	// 尝试不同的登录方式来获取有效令牌
	loginMethods := []struct {
		name     string
		email    string
		password string
		endpoint string
	}{
		{
			name:     "管理员登录",
			email:    "admin@example.com",
			password: "admin123",
			endpoint: "/api/v1/auth/login",
		},
		{
			name:     "用户登录",
			email:    "user@example.com",
			password: "user123",
			endpoint: "/api/v1/auth/login",
		},
		{
			name:     "测试登录",
			email:    "test@example.com",
			password: "test123",
			endpoint: "/api/v1/auth/login",
		},
	}

	baseURL := "http://localhost:8080"

	for _, method := range loginMethods {
		fmt.Printf("\n🔍 尝试: %s\n", method.name)

		token, err := tryLogin(baseURL, method.endpoint, method.email, method.password)
		if err != nil {
			fmt.Printf("   ❌ 登录失败: %v\n", err)
			continue
		}

		if token != "" {
			fmt.Printf("✅ 获取到有效令牌\n")
			fmt.Printf("📋 令牌: %s\n", token)

			// 测试令牌是否有效
			valid := testToken(baseURL, token)
			if valid {
				fmt.Printf("✅ 令牌验证成功\n")
				fmt.Printf("\n🚀 使用方法:\n")
				fmt.Printf("1. 将此令牌设置到前端:\n")
				fmt.Printf("   localStorage.setItem('auth_token', '%s')\n", token)
				fmt.Printf("   localStorage.setItem('law_oa_token', '%s')\n", token)
				fmt.Printf("2. 刷新前端页面\n")
				fmt.Printf("3. 重新测试搜索功能\n")
				return
			} else {
				fmt.Printf("❌ 令牌验证失败\n")
			}
		}
	}

	fmt.Printf("\n⚠️ 所有登录方式都失败了\n")
	fmt.Printf("💡 建议:\n")
	fmt.Printf("1. 检查后端认证服务是否正常运行\n")
	fmt.Printf("2. 确认用户账号和密码是否正确\n")
	fmt.Printf("3. 检查数据库中是否存在用户记录\n")

	// 尝试创建测试用户
	fmt.Printf("\n🔧 尝试创建测试用户...\n")
	if createTestUser(baseURL) {
		fmt.Printf("✅ 测试用户创建成功，请重新运行此脚本获取令牌\n")
	}
}

func tryLogin(baseURL, endpoint, email, password string) (string, error) {
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	url := baseURL + endpoint

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("   📊 登录响应: %d\n", resp.StatusCode)
	fmt.Printf("   📄 响应内容: %s\n", string(body))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", err
	}

	if loginResp.Success && loginResp.Token != "" {
		return loginResp.Token, nil
	}

	return "", fmt.Errorf("登录响应中没有令牌")
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

func createTestUser(baseURL string) bool {
	// 尝试注册测试用户
	userData := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "test123",
		"role":     "admin",
	}

	jsonData, _ := json.Marshal(userData)
	client := &http.Client{Timeout: 10 * time.Second}

	// 尝试注册
	resp, err := client.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ 注册请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   📊 注册响应: %d\n", resp.StatusCode)
	fmt.Printf("   📄 注册内容: %s\n", string(body))

	// 如果注册失败，可能用户已存在，那也是好事
	if resp.StatusCode == 200 || resp.StatusCode == 409 {
		return true
	}

	return false
}