package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🧪 测试律师删除API响应详情")
	fmt.Println("============================")

	// 获取JWT token
	token := getJWTToken()
	if token == "" {
		log.Fatal("无法获取JWT token")
	}

	fmt.Printf("✅ 成功获取JWT token\n\n")

	// 测试删除不存在的律师
	testDeleteLawyerDetail(token, "99999")
}

func getJWTToken() string {
	// 登录获取token
	loginData := map[string]string{
		"email":    "admin@example.com",
		"password": "admin123",
	}

	jsonData, _ := json.Marshal(loginData)

	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("登录失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
		return ""
	}

	var loginResp map[string]interface{}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		log.Printf("解析响应失败: %v", err)
		return ""
	}

	if data, ok := loginResp["data"].(map[string]interface{}); ok {
		if token, ok := data["token"].(string); ok {
			return token
		}
	}

	return ""
}

func testDeleteLawyerDetail(token string, lawyerID string) {
	fmt.Printf("📋 测试删除律师详情 (ID: %s)\n", lawyerID)
	fmt.Println("-----------------------------------")

	// 创建HTTP请求
	url := fmt.Sprintf("http://localhost:8080/api/lawfirm/lawyers/%s", lawyerID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
		return
	}

	// 格式化输出JSON响应
	var prettyJSON interface{}
	if err := json.Unmarshal(body, &prettyJSON); err == nil {
		formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
		fmt.Printf("响应内容:\n%s\n", string(formatted))
	} else {
		fmt.Printf("响应内容 (原始):\n%s\n", string(body))
	}

	fmt.Printf("\n状态码: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
}