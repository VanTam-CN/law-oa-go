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
	fmt.Println("🧪 测试律师删除功能")
	fmt.Println("===================")

	// 获取JWT token
	token := getJWTToken()
	if token == "" {
		log.Fatal("无法获取JWT token")
	}

	fmt.Printf("✅ 成功获取JWT token\n\n")

	// 测试删除不存在的律师
	testDeleteNonExistentLawyer(token)

	// 测试删除存在的律师（先获取律师列表，然后删除第一个）
	lawyerID := getFirstLawyerID(token)
	if lawyerID != "" {
		testDeleteExistingLawyer(token, lawyerID)
	} else {
		fmt.Println("⚠️  没有找到可删除的律师")
	}

	fmt.Println("\n🎉 律师删除功能测试完成！")
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

func getFirstLawyerID(token string) string {
	// 创建HTTP请求获取律师列表
	req, err := http.NewRequest("GET", "http://localhost:8080/api/lawfirm/lawyers?page=1&page_size=1", nil)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return ""
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
		return ""
	}

	var listResp map[string]interface{}
	if err := json.Unmarshal(body, &listResp); err != nil {
		log.Printf("解析响应失败: %v", err)
		return ""
	}

	if data, ok := listResp["data"].(map[string]interface{}); ok {
		if items, ok := data["items"].([]interface{}); ok && len(items) > 0 {
			if lawyer, ok := items[0].(map[string]interface{}); ok {
				if id, ok := lawyer["id"].(float64); ok {
					return fmt.Sprintf("%.0f", id)
				}
			}
		}
	}

	return ""
}

func testDeleteNonExistentLawyer(token string) {
	fmt.Println("📋 测试场景1：删除不存在的律师")
	fmt.Println("-----------------------------------")

	// 尝试删除一个不存在的律师ID
	lawyerID := "99999"
	testDeleteLawyer(token, lawyerID, "不存在的律师")
}

func testDeleteExistingLawyer(token string, lawyerID string) {
	fmt.Println("\n📋 测试场景2：删除存在的律师")
	fmt.Println("-----------------------------------")

	testDeleteLawyer(token, lawyerID, "存在的律师")
}

func testDeleteLawyer(token string, lawyerID string, description string) {
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

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("解析响应失败: %v", err)
		return
	}

	// 显示结果
	fmt.Printf("测试: %s (ID: %s)\n", description, lawyerID)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("成功: %t\n", response["success"])

	if message, ok := response["message"].(string); ok {
		fmt.Printf("消息: %s\n", message)
	}

	if errorMsg, ok := response["error"].(string); ok && errorMsg != "" {
		fmt.Printf("错误: %s\n", errorMsg)
	}

	fmt.Printf("时间戳: %s\n\n", response["timestamp"])
}