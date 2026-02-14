//go:build ignore

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
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type CaseDetailResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID          uint    `json:"id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		ClientID    uint    `json:"client_id"`
		ClientName  string  `json:"client_name,omitempty"`
		LawyerID    uint    `json:"lawyer_id"`
		LawyerName  string  `json:"lawyer_name,omitempty"`
		CaseType    string  `json:"case_type"`
		Priority    string  `json:"priority"`
		Status      string  `json:"status"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
		Client      *struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Email   string `json:"email"`
			Phone   string `json:"phone"`
			Address string `json:"address"`
			Company string `json:"company"`
			Status  string `json:"status"`
		} `json:"client,omitempty"`
		Lawyer *struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"lawyer,omitempty"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func main() {
	baseURL := "http://localhost:8080"

	// 1. 登录获取token
	fmt.Println("🔐 登录获取token...")
	loginReq := LoginRequest{
		Email:    "zhangwei@jinchenglaw.com",
		Password: "law123456",
	}

	loginData, _ := json.Marshal(loginReq)
	resp, err := http.Post(baseURL+"/api/auth/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	if !loginResp.Success || loginResp.Data.Token == "" {
		fmt.Printf("❌ 登录失败: %s\n", string(body))
		return
	}

	token := loginResp.Data.Token
	fmt.Printf("✅ 登录成功，获取token: %s...\n", token[:20])

	// 2. 获取案件列表，找一个有效的案件ID
	fmt.Println("\n📋 获取案件列表...")
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", baseURL+"/api/cases?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 获取案件列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("📄 案件列表响应状态: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Printf("❌ 案件列表请求失败: %s\n", string(body))
		return
	}

	// 打印案件列表响应内容用于调试
	fmt.Printf("📄 案件列表响应内容: %s\n", string(body))

	// 解析案件列表获取第一个案件ID
	var casesListResp struct {
		Success bool `json:"success"`
		Data    []struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(body, &casesListResp)

	if !casesListResp.Success || len(casesListResp.Data) == 0 {
		fmt.Println("❌ 未找到任何案件")
		return
	}

	caseID := casesListResp.Data[0].ID
	fmt.Printf("✅ 找到案件ID: %d\n", caseID)

	// 3. 测试案件详情API
	fmt.Printf("\n🔍 测试案件详情API (ID: %d)...\n", caseID)
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/cases/%d", baseURL, caseID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 案件详情请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("📄 案件详情响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		var caseDetailResp CaseDetailResponse
		json.Unmarshal(body, &caseDetailResp)

		if caseDetailResp.Success {
			fmt.Printf("✅ 案件详情API工作正常！\n")
			fmt.Printf("📋 案件信息:\n")
			fmt.Printf("   - ID: %d\n", caseDetailResp.Data.ID)
			fmt.Printf("   - 标题: %s\n", caseDetailResp.Data.Title)
			fmt.Printf("   - 类型: %s\n", caseDetailResp.Data.CaseType)
			fmt.Printf("   - 状态: %s\n", caseDetailResp.Data.Status)
			fmt.Printf("   - 客户: %s\n", caseDetailResp.Data.ClientName)
			fmt.Printf("   - 律师: %s\n", caseDetailResp.Data.LawyerName)
			fmt.Printf("   - 创建时间: %s\n", caseDetailResp.Data.CreatedAt)
		} else {
			fmt.Printf("❌ 案件详情API返回错误: %s\n", string(body))
		}
	} else {
		fmt.Printf("❌ 案件详情API请求失败 (状态码: %d)\n", resp.StatusCode)
		fmt.Printf("📄 响应内容: %s\n", string(body))
	}
}