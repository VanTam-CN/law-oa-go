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

type ConflictCheckRequest struct {
    ClientID                  string   `json:"clientId"`
    ClientName                string   `json:"clientName"`
    CaseName                  string   `json:"caseName"`
    CaseType                  string   `json:"caseType"`
    ClientType                string   `json:"clientType"`
    OtherParties              []string `json:"otherParties"`
    SearchYears               int      `json:"searchYears"`
    IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
    SearchDepth               string   `json:"searchDepth"`
    UserID                    string   `json:"userId"`
    RequestTime               time.Time `json:"requestTime"`
}

func main() {
    fmt.Println("=== 测试利益冲突检测API ===")

    // 1. 先获取登录token
    fmt.Println("\n1. 获取登录token...")
    loginData := map[string]string{
        "email":    "admin@lawoa.com",
        "password": "admin123",
    }

    loginJSON, _ := json.Marshal(loginData)
    resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(loginJSON))
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

    var loginResponse map[string]interface{}
    json.Unmarshal(body, &loginResponse)

    // 检查响应格式
    if success, ok := loginResponse["success"].(bool); ok && success {
        if data, ok := loginResponse["data"].(map[string]interface{}); ok {
            if token, ok := data["token"].(string); ok {
                fmt.Printf("获取到token: %s...\n", token[:50])

                // 2. 测试利益冲突检测
                fmt.Println("\n2. 测试利益冲突检测...")

                // 张伟律师为腾讯创建案件，对方当事人是阿里巴巴
                conflictRequest := ConflictCheckRequest{
                    ClientID:                  "13",           // 腾讯控股有限公司
                    ClientName:                "腾讯控股有限公司",
                    CaseName:                  "腾讯诉阿里巴巴不正当竞争案",
                    CaseType:                  "commercial",
                    ClientType:                "COMPANY",
                    OtherParties:              []string{"阿里巴巴集团控股有限公司"},
                    SearchYears:               5,
                    IncludeCorporateRelations: true,
                    SearchDepth:               "STANDARD",
                    UserID:                    "6",            // 张伟律师ID
                    RequestTime:               time.Now(),
                }

                requestJSON, _ := json.Marshal(conflictRequest)
                fmt.Printf("请求数据: %s\n", string(requestJSON))

                req, err := http.NewRequest("POST", "http://localhost:8080/api/conflict/check", bytes.NewBuffer(requestJSON))
                if err != nil {
                    fmt.Printf("创建请求失败: %v\n", err)
                    return
                }

                req.Header.Set("Content-Type", "application/json")
                req.Header.Set("Authorization", "Bearer "+token)

                client := &http.Client{Timeout: 10 * time.Second}
                resp, err = client.Do(req)
                if err != nil {
                    fmt.Printf("发送请求失败: %v\n", err)
                    return
                }
                defer resp.Body.Close()

                response, err := io.ReadAll(resp.Body)
                if err != nil {
                    fmt.Printf("读取响应失败: %v\n", err)
                    return
                }

                fmt.Printf("\n响应状态码: %d\n", resp.StatusCode)
                fmt.Printf("响应内容: %s\n", string(response))

                if resp.StatusCode == 200 {
                    fmt.Println("\n✅ 利益冲突检测成功！")
                    // 解析响应
                    var conflictResponse map[string]interface{}
                    json.Unmarshal(response, &conflictResponse)

                    if data, ok := conflictResponse["data"].(map[string]interface{}); ok {
                        fmt.Printf("冲突检测结果: %v\n", data)
                    }
                } else {
                    fmt.Printf("\n❌ 利益冲突检测失败，状态码: %d\n", resp.StatusCode)
                }
            } else {
                fmt.Println("获取token失败: 响应中没有token字段")
            }
        } else {
            fmt.Println("获取token失败: 响应中没有data字段")
        }
    } else {
        fmt.Printf("登录失败: %v\n", loginResponse)
    }

    // 2. 测试利益冲突检测
    fmt.Println("\n2. 测试利益冲突检测...")

    // 张伟律师为腾讯创建案件，对方当事人是阿里巴巴
    conflictRequest := ConflictCheckRequest{
        ClientID:                  "13",           // 腾讯控股有限公司
        ClientName:                "腾讯控股有限公司",
        CaseName:                  "腾讯诉阿里巴巴不正当竞争案",
        CaseType:                  "commercial",
        ClientType:                "COMPANY",
        OtherParties:              []string{"阿里巴巴集团控股有限公司"},
        SearchYears:               5,
        IncludeCorporateRelations: true,
        SearchDepth:               "STANDARD",
        UserID:                    "6",            // 张伟律师ID
        RequestTime:               time.Now(),
    }

    requestJSON, _ := json.Marshal(conflictRequest)
    fmt.Printf("请求数据: %s\n", string(requestJSON))

    req, err := http.NewRequest("POST", "http://localhost:8080/api/conflict/check", bytes.NewBuffer(requestJSON))
    if err != nil {
        fmt.Printf("创建请求失败: %v\n", err)
        return
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err = client.Do(req)
    if err != nil {
        fmt.Printf("发送请求失败: %v\n", err)
        return
    }
    defer resp.Body.Close()

    response, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("读取响应失败: %v\n", err)
        return
    }

    fmt.Printf("\n响应状态码: %d\n", resp.StatusCode)
    fmt.Printf("响应内容: %s\n", string(response))

    if resp.StatusCode == 200 {
        fmt.Println("\n✅ 利益冲突检测成功！")
        // 解析响应
        var conflictResponse map[string]interface{}
        json.Unmarshal(response, &conflictResponse)

        if data, ok := conflictResponse["data"].(map[string]interface{}); ok {
            fmt.Printf("冲突检测结果: %v\n", data)
        }
    } else {
        fmt.Printf("\n❌ 利益冲突检测失败，状态码: %d\n", resp.StatusCode)
    }
}