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
		User  struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func main() {
	fmt.Println("🔐 获取认证令牌...")

	// 尝试登录张伟律师账号
	token, err := login("zhangwei@law.com", "law123456", "张伟")
	if err != nil {
		fmt.Printf("❌ 张伟登录失败: %v\n", err)

		// 尝试陈浩律师账号
		token, err = login("chenhao@law.com", "law123456", "陈浩")
		if err != nil {
			fmt.Printf("❌ 陈浩登录失败: %v\n", err)

			// 尝试王芳律师账号
			token, err = login("wangfang@law.com", "law123456", "王芳")
			if err != nil {
				fmt.Printf("❌ 王芳登录失败: %v\n", err)
				return
			}
		}
	}

	fmt.Printf("✅ 获取到认证令牌: %s\n", token[:50]+"...")

	// 测试令牌是否有效
	testToken(token)
}

func login(email, password, name string) (string, error) {
	fmt.Printf("🔑 尝试登录: %s (%s)\n", name, email)

	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("序列化登录请求失败: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("发送登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取登录响应失败: %w", err)
	}

	fmt.Printf("📥 登录响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("📋 登录响应: %s\n", string(body))

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %w", err)
	}

	if !loginResp.Success {
		return "", fmt.Errorf("登录失败: %s", loginResp.Error.Message)
	}

	if loginResp.Data.Token == "" {
		return "", fmt.Errorf("登录响应中没有令牌")
	}

	fmt.Printf("✅ %s 登录成功 (ID: %d, Role: %s)\n",
		loginResp.Data.User.Name, loginResp.Data.User.ID, loginResp.Data.User.Role)

	return loginResp.Data.Token, nil
}

func testToken(token string) {
	fmt.Println("\n🧪 测试令牌有效性...")

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/conflict/health", nil)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 发送请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("📥 健康检查状态码: %d\n", resp.StatusCode)
	fmt.Printf("📋 健康检查响应: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ 令牌有效，可以进行冲突检测测试")
	} else {
		fmt.Println("❌ 令牌无效或已过期")
	}
}
