package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	// 先获取一个有效的token
	loginData := `{
		"username": "admin",
		"password": "admin123"
	}`

	// 发送登录请求
	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", strings.NewReader(loginData))
	if err != nil {
		fmt.Printf("登录请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取登录响应失败: %v\n", err)
		return
	}

	fmt.Printf("登录响应: %s\n", string(body))

	// 提取token（简单处理）
	responseStr := string(body)
	tokenStart := strings.Index(responseStr, `"token":"`)
	if tokenStart == -1 {
		fmt.Println("未找到token")
		return
	}
	tokenStart += 9
	tokenEnd := strings.Index(responseStr[tokenStart:], `"`)
	if tokenEnd == -1 {
		fmt.Println("token格式错误")
		return
	}
	token := responseStr[tokenStart : tokenStart+tokenEnd]

	fmt.Printf("获取到token: %s...\n", token[:min(50, len(token))])

	// 使用token获取客户数据
	clientReq, err := http.NewRequest("GET", "http://localhost:8080/api/clients?pageNum=1&pageSize=10", nil)
	if err != nil {
		fmt.Printf("创建客户请求失败: %v\n", err)
		return
	}

	clientReq.Header.Set("Authorization", "Bearer "+token)
	clientReq.Header.Set("Content-Type", "application/json")

	clientResp, err := http.DefaultClient.Do(clientReq)
	if err != nil {
		fmt.Printf("获取客户数据失败: %v\n", err)
		return
	}
	defer clientResp.Body.Close()

	clientBody, err := io.ReadAll(clientResp.Body)
	if err != nil {
		fmt.Printf("读取客户数据失败: %v\n", err)
		return
	}

	fmt.Printf("\n=== 客户数据 ===\n%s\n", string(clientBody))

	// 获取案件数据
	caseReq, err := http.NewRequest("GET", "http://localhost:8080/api/cases?page=1&page_size=5", nil)
	if err != nil {
		fmt.Printf("创建案件请求失败: %v\n", err)
		return
	}

	caseReq.Header.Set("Authorization", "Bearer "+token)
	caseReq.Header.Set("Content-Type", "application/json")

	caseResp, err := http.DefaultClient.Do(caseReq)
	if err != nil {
		fmt.Printf("获取案件数据失败: %v\n", err)
		return
	}
	defer caseResp.Body.Close()

	caseBody, err := io.ReadAll(caseResp.Body)
	if err != nil {
		fmt.Printf("读取案件数据失败: %v\n", err)
		return
	}

	fmt.Printf("\n=== 案件数据 ===\n%s\n", string(caseBody))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}