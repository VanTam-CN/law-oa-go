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

// ConflictCheckRequest 冲突检测请求
type ConflictCheckRequest struct {
	ClientID                 string    `json:"clientId"`
	ClientName               string    `json:"clientName"`
	ClientType               string    `json:"clientType"`
	OtherParties             []string  `json:"otherParties"`
	CaseName                 string    `json:"caseName"`
	CaseType                 string    `json:"caseType"`
	SearchYears             int       `json:"searchYears"`
	IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
	SearchDepth             string    `json:"searchDepth"`
	UserID                   string    `json:"userId"`
	RequestTime              time.Time `json:"requestTime"`
}

// ConflictCheckResponse 冲突检测响应
type ConflictCheckResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
	Timestamp  string `json:"timestamp"`
}

func main() {
	fmt.Println("🔍 测试利益冲突检测功能")
	fmt.Println("=========================")

	// 获取JWT token
	token := getJWTToken()
	if token == "" {
		log.Fatal("无法获取JWT token")
	}

	fmt.Printf("✅ 成功获取JWT token\n\n")

	// 测试场景1：直接利益冲突检测
	testDirectConflict(token)

	// 测试场景2：间接利益冲突检测
	testIndirectConflict(token)

	// 测试场景3：无冲突检测
	testNoConflict(token)

	fmt.Println("\n🎉 利益冲突检测测试完成！")
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

func testDirectConflict(token string) {
	fmt.Println("📋 测试场景1：直接利益冲突检测")
	fmt.Println("-----------------------------------")

	// 测试阿里巴巴客户的新案件，检查与腾讯的冲突
	request := ConflictCheckRequest{
		ClientID:                 "1", // 假设阿里巴巴的客户ID是1
		ClientName:               "阿里巴巴集团",
		ClientType:               "COMPANY",
		OtherParties:             []string{"腾讯控股"},
		CaseName:                 "阿里巴巴新诉腾讯商业诋毁案",
		CaseType:                 "商业纠纷",
		SearchYears:             5,
		IncludeCorporateRelations: true,
		SearchDepth:             "STANDARD",
		UserID:                   "1",
		RequestTime:              time.Now(),
	}

	testConflictCheck(token, request, "阿里巴巴 vs 腾讯")
}

func testIndirectConflict(token string) {
	fmt.Println("\n📋 测试场景2：间接利益冲突检测")
	fmt.Println("-----------------------------------")

	// 测试华为客户的新案件，检查与小米的冲突
	request := ConflictCheckRequest{
		ClientID:                 "2", // 假设华为的客户ID是2
		ClientName:               "华为技术有限公司",
		ClientType:               "COMPANY",
		OtherParties:             []string{"小米科技"},
		CaseName:                 "华为诉小米不正当竞争案",
		CaseType:                 "竞争纠纷",
		SearchYears:             3,
		IncludeCorporateRelations: true,
		SearchDepth:             "STANDARD",
		UserID:                   "1",
		RequestTime:              time.Now(),
	}

	testConflictCheck(token, request, "华为 vs 小米")
}

func testNoConflict(token string) {
	fmt.Println("\n📋 测试场景3：无冲突检测")
	fmt.Println("-----------------------------------")

	// 测试微软客户的新案件，应该无冲突
	request := ConflictCheckRequest{
		ClientID:                 "9", // 假设微软的客户ID是9
		ClientName:               "微软中国",
		ClientType:               "COMPANY",
		OtherParties:             []string{"某软件公司"},
		CaseName:                 "微软中国软件许可合同纠纷案",
		CaseType:                 "合同纠纷",
		SearchYears:             2,
		IncludeCorporateRelations: false,
		SearchDepth:             "BASIC",
		UserID:                   "1",
		RequestTime:              time.Now(),
	}

	testConflictCheck(token, request, "微软中国（无冲突）")
}

func testConflictCheck(token string, request ConflictCheckRequest, scenario string) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("序列化请求失败: %v", err)
		return
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/conflict/check", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
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

	var response ConflictCheckResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("解析响应失败: %v", err)
		return
	}

	// 显示结果
	fmt.Printf("场景: %s\n", scenario)
	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("成功: %t\n", response.Success)
	fmt.Printf("消息: %s\n", response.Message)

	if response.Error != "" {
		fmt.Printf("错误: %s\n", response.Error)
	}

	if response.Data != nil {
		fmt.Printf("数据: %+v\n", response.Data)
	}

	fmt.Printf("时间戳: %s\n\n", response.Timestamp)
}