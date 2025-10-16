package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"
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
			ID       int    `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Role     string `json:"role"`
		} `json:"user"`
	} `json:"data"`
}

func main() {
	fmt.Println("🔧 创建测试用户并获取认证令牌")
	fmt.Println("=====================================")

	// 第一步：创建测试用户
	fmt.Println("\n📋 第一步：创建测试用户")
	userInfo, err := createTestUser()
	if err != nil {
		log.Fatalf("❌ 创建测试用户失败: %v", err)
	}
	fmt.Printf("✅ 测试用户创建成功\n")
	fmt.Printf("   📧 邮箱: %s\n", userInfo.Email)
	fmt.Printf("   🔑 密码: %s\n", userInfo.Password)
	fmt.Printf("   🎭 角色: %s\n", userInfo.Role)

	// 第二步：获取认证令牌
	fmt.Println("\n📋 第二步：获取认证令牌")
	token, err := getAuthToken(userInfo.Email, userInfo.Password)
	if err != nil {
		log.Fatalf("❌ 获取认证令牌失败: %v", err)
	}
	fmt.Printf("✅ 认证令牌获取成功\n")
	fmt.Printf("   🎫 令牌: %s...\n", token[:50])

	// 第三步：验证令牌有效性
	fmt.Println("\n📋 第三步：验证令牌有效性")
	if validateToken(token) {
		fmt.Printf("✅ 令牌验证成功\n")
	} else {
		log.Fatalf("❌ 令牌验证失败")
	}

	// 第四步：生成前端配置
	fmt.Println("\n📋 第四步：生成前端配置")
	generateFrontendConfig(token)

	fmt.Println("\n🎉 测试环境准备完成！")
	fmt.Println("=====================================")
	fmt.Println("现在可以进行搜索功能测试了")
}

type TestUserInfo struct {
	Email    string
	Password string
	Role     string
}

func createTestUser() (*TestUserInfo, error) {
	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	defer db.Close()

	// 用户信息
	userInfo := &TestUserInfo{
		Email:    "searchtest@law-oa.com",
		Password: "searchtest123",
		Role:     "admin",
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInfo.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码哈希生成失败: %w", err)
	}

	// 插入或更新用户
	query := `
		INSERT INTO users (name, username, email, password, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'active', NOW(), NOW())
		ON DUPLICATE KEY UPDATE
		password = ?, name = ?, username = ?, role = ?, status = 'active', updated_at = NOW()
	`

	_, err = db.Exec(query, "搜索测试用户", "searchtest", userInfo.Email, string(hashedPassword), userInfo.Role,
		string(hashedPassword), "搜索测试用户", "searchtest", userInfo.Role)
	if err != nil {
		return nil, fmt.Errorf("插入用户失败: %w", err)
	}

	// 验证用户创建成功
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", userInfo.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("验证用户失败: %w", err)
	}

	if count == 0 {
		return nil, fmt.Errorf("用户创建验证失败")
	}

	return userInfo, nil
}

func getAuthToken(email, password string) (string, error) {
	// 准备登录请求
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", fmt.Errorf("JSON编码失败: %w", err)
	}

	// 尝试不同的登录端点
	endpoints := []string{
		"http://localhost:8080/api/v1/auth/login",
		"http://localhost:8080/api/auth/login",
		"http://localhost:8080/auth/login",
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, endpoint := range endpoints {
		fmt.Printf("   🔍 尝试端点: %s\n", endpoint)

		resp, err := client.Post(endpoint, "application/json",
			bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("     ❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("     ❌ 读取响应失败: %v\n", err)
			continue
		}

		fmt.Printf("     📊 状态码: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			var loginResp LoginResponse
			if err := json.Unmarshal(body, &loginResp); err != nil {
				fmt.Printf("     ⚠️ JSON解析失败: %v\n", err)
				continue
			}

			if loginResp.Success && loginResp.Data.Token != "" {
				fmt.Printf("     ✅ 登录成功\n")
				return loginResp.Data.Token, nil
			}
		}
	}

	return "", fmt.Errorf("所有登录端点都失败了")
}

func validateToken(token string) bool {
	client := &http.Client{Timeout: 5 * time.Second}

	// 测试客户列表API
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/clients?page=1&page_size=1", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func generateFrontendConfig(token string) {
	fmt.Println("📱 前端配置代码:")
	fmt.Println("=====================================")
	fmt.Printf("// 在浏览器控制台中运行以下代码\n\n")
	fmt.Printf("localStorage.setItem('auth_token', '%s');\n", token)
	fmt.Printf("localStorage.setItem('law_oa_token', '%s');\n\n", token)
	fmt.Printf("// 然后刷新页面\n")
	fmt.Printf("location.reload();\n\n")

	// 生成前端配置信息
	fmt.Println("📄 生成的配置信息:")
	fmt.Println("=====================================")
	fmt.Printf("// 令牌: %s\n", token)
	fmt.Printf("localStorage.setItem('auth_token', '%s');\n", token)
	fmt.Printf("localStorage.setItem('law_oa_token', '%s');\n", token)
	fmt.Printf("location.reload();\n")
}