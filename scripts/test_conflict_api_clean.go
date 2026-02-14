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

    fmt.Printf("登录响应状态码: %d\n", resp.StatusCode)
    fmt.Printf("登录响应内容: %s\n", string(body))

    var loginResponse map[string]interface{}
    json.Unmarshal(body, &loginResponse)

    // 检查响应格式
    if success, ok := loginResponse["success"].(bool); ok && success {
        if data, ok := loginResponse["data"].(map[string]interface{}); ok {
            if token, ok := data["token"].(string); ok {
                fmt.Printf("✅ 获取到token: %s...\n", token[:50])

                // 2. 测试利益冲突检测
                fmt.Println("\n2. 测试利益冲突检测...")

                // 张伟律师为腾讯创建案件，对方当事人是阿里巴巴
                // 这种情况应该检测到商业竞争冲突，因为张伟已经代理了阿里巴巴的案件
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
                fmt.Printf("冲突检测请求数据: %s\n", string(requestJSON))

                req, err := http.NewRequest("POST", "http://localhost:8080/api/conflict/check", bytes.NewBuffer(requestJSON))
                if err != nil {
                    fmt.Printf("创建冲突检测请求失败: %v\n", err)
                    return
                }

                req.Header.Set("Content-Type", "application/json")
                req.Header.Set("Authorization", "Bearer "+token)

                client := &http.Client{Timeout: 10 * time.Second}
                resp, err = client.Do(req)
                if err != nil {
                    fmt.Printf("发送冲突检测请求失败: %v\n", err)
                    return
                }
                defer resp.Body.Close()

                response, err := io.ReadAll(resp.Body)
                if err != nil {
                    fmt.Printf("读取冲突检测响应失败: %v\n", err)
                    return
                }

                fmt.Printf("\n冲突检测响应状态码: %d\n", resp.StatusCode)
                fmt.Printf("冲突检测响应内容: %s\n", string(response))

                if resp.StatusCode == 200 {
                    fmt.Println("\n✅ 利益冲突检测成功！")
                    // 解析响应
                    var conflictResponse map[string]interface{}
                    json.Unmarshal(response, &conflictResponse)

                    if data, ok := conflictResponse["data"].(map[string]interface{}); ok {
                        fmt.Printf("冲突检测结果:\n")

                        // 检查是否检测到冲突
                        if hasConflict, ok := data["hasConflict"].(bool); ok {
                            if hasConflict {
                                fmt.Printf("  ⚠️ 检测到利益冲突！\n")

                                // 显示冲突案例
                                if cases, ok := data["conflictCases"].([]interface{}); ok {
                                    fmt.Printf("  📋 冲突案件数量: %d\n", len(cases))
                                    for i, caseData := range cases {
                                        if caseMap, ok := caseData.(map[string]interface{}); ok {
                                            fmt.Printf("    %d. %s - 客户: %s\n", i+1,
                                                caseMap["caseName"], caseMap["clientName"])
                                            fmt.Printf("       冲突类型: %s, 风险等级: %s\n",
                                                caseMap["conflictType"], caseMap["riskLevel"])
                                        }
                                    }
                                }

                                // 显示风险评估
                                if risk, ok := data["riskAssessment"].(map[string]interface{}); ok {
                                    fmt.Printf("  🎯 综合风险等级: %s (评分: %.1f/100)\n",
                                        risk["overallRisk"], risk["riskScore"])
                                    fmt.Printf("  📝 风险原因: %s\n", risk["riskReason"])
                                }
                            } else {
                                fmt.Printf("  ✅ 未检测到利益冲突\n")
                            }
                        }

                        // 显示检查统计
                        if stats, ok := data["checkStatistics"].(map[string]interface{}); ok {
                            fmt.Printf("  📊 检查统计:\n")
                            fmt.Printf("     总检查案件: %.0f\n", stats["totalCasesChecked"])
                            fmt.Printf("     客户历史案件: %.0f\n", stats["clientHistoryCases"])
                        }
                    }
                } else {
                    fmt.Printf("\n❌ 利益冲突检测失败，状态码: %d\n", resp.StatusCode)

                    // 尝试解析错误信息
                    var errorResponse map[string]interface{}
                    json.Unmarshal(response, &errorResponse)
                    if errMsg, ok := errorResponse["error"].(map[string]interface{}); ok {
                        fmt.Printf("错误详情: %v\n", errMsg)
                    }
                }
            } else {
                fmt.Println("❌ 获取token失败: 响应中没有token字段")
            }
        } else {
            fmt.Println("❌ 获取token失败: 响应中没有data字段")
        }
    } else {
        fmt.Printf("❌ 登录失败: %v\n", loginResponse)
    }
}