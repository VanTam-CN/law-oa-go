package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ConflictCheckRequest struct {
	ClientID                 string    `json:"clientId"`
	ClientName               string    `json:"clientName"`
	ClientType               string    `json:"clientType"`
	CaseName                 string    `json:"caseName"`
	CaseType                 string    `json:"caseType"`
	OtherParties             []string  `json:"otherParties"`
	UserID                   string    `json:"userId"`
	SearchYears              int       `json:"searchYears"`
	SearchDepth              string    `json:"searchDepth"`
	IncludeCorporateRelations bool      `json:"includeCorporateRelations"`
	RequestTime              time.Time `json:"requestTime"`
}

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

	// 2. 测试冲突检测API - 使用修复后的数据格式
	fmt.Println("\n🔍 测试冲突检测API...")

	conflictReq := ConflictCheckRequest{
		ClientID:                 "1",
		ClientName:               "阿里巴巴集团",
		ClientType:               "COMPANY",
		CaseName:                 "阿里巴巴诉字节跳动不正当竞争案",
		CaseType:                 "COMMERCIAL",
		OtherParties:             []string{"字节跳动"},
		UserID:                   "1",
		SearchYears:              5,
		SearchDepth:              "DEEP",
		IncludeCorporateRelations: true,
		RequestTime:              time.Now(),
	}

	conflictData, _ := json.Marshal(conflictReq)

	// 创建新的请求
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/api/conflict/check", bytes.NewBuffer(conflictData))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("📤 发送冲突检测请求:\n")
	fmt.Printf("   - 客户: %s (ID: %s, 类型: %s)\n", conflictReq.ClientName, conflictReq.ClientID, conflictReq.ClientType)
	fmt.Printf("   - 案件: %s (类型: %s)\n", conflictReq.CaseName, conflictReq.CaseType)
	fmt.Printf("   - 用户ID: %s\n", conflictReq.UserID)
	fmt.Printf("   - 对方当事人: %v\n", conflictReq.OtherParties)

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("❌ 冲突检测请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Printf("\n📄 冲突检测响应状态: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Printf("✅ 冲突检测API调用成功！\n")
		fmt.Printf("📋 响应内容: %s\n", string(body))

		// 尝试解析响应
		var response map[string]interface{}
		if err := json.Unmarshal(body, &response); err == nil {
			if success, ok := response["success"].(bool); ok && success {
				fmt.Printf("\n🎉 冲突检测功能已修复！\n")
				fmt.Printf("✅ API返回成功状态\n")
				if data, exists := response["data"]; exists {
					fmt.Printf("📊 检测结果数据: %+v\n", data)
				}
			} else {
				fmt.Printf("⚠️ API调用成功但返回失败状态\n")
				if errorMsg, exists := response["error"]; exists {
					fmt.Printf("错误信息: %v\n", errorMsg)
				}
			}
		}
	} else {
		fmt.Printf("❌ 冲突检测API调用失败 (状态码: %d)\n", resp.StatusCode)
		fmt.Printf("📄 响应内容: %s\n", string(body))
	}
}