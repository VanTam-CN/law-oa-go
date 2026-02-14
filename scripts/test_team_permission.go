//go:build ignore

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
	baseURL := "http://localhost:8080/api/v1"

	// 1. 测试服务器连通性
	fmt.Println("🔍 测试服务器连通性...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		log.Fatal("服务器连接失败:", err)
	}
	resp.Body.Close()
	fmt.Println("✅ 服务器连接正常")

	// 2. 登录获取JWT令牌
	fmt.Println("\n🔐 登录获取JWT令牌...")
	loginData := map[string]string{
		"email":    "admin@lawoa.com",
		"password": "admin123",
	}

	loginJSON, _ := json.Marshal(loginData)
	resp, err = http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		log.Fatal("登录请求失败:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("读取登录响应失败:", err)
	}

	var loginResponse map[string]interface{}
	if err := json.Unmarshal(body, &loginResponse); err != nil {
		log.Fatal("解析登录响应失败:", err)
	}

	if resp.StatusCode != 200 {
		fmt.Printf("❌ 登录失败: %s\n", string(body))
		return
	}

	// 提取JWT令牌
	data, ok := loginResponse["data"].(map[string]interface{})
	if !ok {
		log.Fatal("登录响应格式错误：找不到data字段")
	}

	token, ok := data["token"].(string)
	if !ok {
		log.Fatal("登录响应格式错误：找不到token字段")
	}

	fmt.Printf("✅ 获取JWT令牌成功\n")

	// 3. 测试团队权限检查
	fmt.Println("\n🧪 测试团队权限检查...")
	testTeamPermission(baseURL, token)

	// 4. 测试团队分配
	fmt.Println("\n🏗️ 测试团队分配...")
	testTeamAssignment(baseURL, token)

	// 5. 测试获取团队信息
	fmt.Println("\n👥 测试获取团队信息...")
	testGetTeamInfo(baseURL, token)
}

func testTeamPermission(baseURL, token string) {
	// 测试权限检查
	permissionCheckData := map[string]interface{}{
		"user_id": 1,
		"case_id": 1,
		"action":  "view",
		"context": map[string]interface{}{
			"test_mode": true,
		},
	}

	permissionJSON, _ := json.Marshal(permissionCheckData)
	req, err := http.NewRequest("POST", baseURL+"/teams/check-permission", bytes.NewBuffer(permissionJSON))
	if err != nil {
		fmt.Printf("❌ 创建权限检查请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 权限检查请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取权限检查响应失败: %v\n", err)
		return
	}

	fmt.Printf("权限检查状态码: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("❌ 解析权限检查响应失败: %v\n", err)
			return
		}

		if data, ok := result["data"].(map[string]interface{}); ok {
			fmt.Printf("✅ 权限检查结果:\n")
			fmt.Printf("  - 用户ID: %v\n", data["user_id"])
			fmt.Printf("  - 案件ID: %v\n", data["case_id"])
			fmt.Printf("  - 操作: %v\n", data["action"])
			fmt.Printf("  - 有权限: %v\n", data["has_permission"])
		}
	} else {
		fmt.Printf("⚠️ 权限检查返回: %s\n", string(body))
	}
}

func testTeamAssignment(baseURL, token string) {
	// 测试团队分配
	assignmentData := map[string]interface{}{
		"case_id":            1,
		"lawyer_id":          1,
		"assisting_lawyer_id": nil,
		"team_members": []map[string]interface{}{
			{
				"user_id":   2,
				"role":     "paralegal",
				"capacity": 50,
			},
		},
		"billing_method": "hourly",
		"is_major_risk":   false,
	}

	assignmentJSON, _ := json.Marshal(assignmentData)
	req, err := http.NewRequest("POST", baseURL+"/teams/assign", bytes.NewBuffer(assignmentJSON))
	if err != nil {
		fmt.Printf("❌ 创建团队分配请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 团队分配请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取团队分配响应失败: %v\n", err)
		return
	}

	fmt.Printf("团队分配状态码: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("❌ 解析团队分配响应失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 团队分配成功\n")
		if data, ok := result["data"].(map[string]interface{}); ok {
			fmt.Printf("  - 案件ID: %v\n", data["case_id"])
			fmt.Printf("  - 分配时间: %v\n", data["assigned_at"])
			if permissions, ok := data["permissions"].(map[string]interface{}); ok {
				fmt.Printf("  - 用户权限: %v\n", permissions)
			}
		}
	} else {
		fmt.Printf("⚠️ 团队分配返回: %s\n", string(body))
	}
}

func testGetTeamInfo(baseURL, token string) {
	// 测试获取团队信息
	req, err := http.NewRequest("GET", baseURL+"/teams/case/1", nil)
	if err != nil {
		fmt.Printf("❌ 创建获取团队信息请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 获取团队信息请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取团队信息响应失败: %v\n", err)
		return
	}

	fmt.Printf("获取团队信息状态码: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("❌ 解析团队信息响应失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 获取团队信息成功\n")
		if data, ok := result["data"].(map[string]interface{}); ok {
			fmt.Printf("  - 案件ID: %v\n", data["case_id"])
			if leadLawyer, ok := data["lead_lawyer"].(map[string]interface{}); ok {
				fmt.Printf("  - 主办律师: %v\n", leadLawyer["name"])
			}
			if permissions, ok := data["permissions"].(map[string]interface{}); ok {
				fmt.Printf("  - 当前用户权限: %v\n", permissions)
			}
		}
	} else {
		fmt.Printf("⚠️ 获取团队信息返回: %s\n", string(body))
	}
}