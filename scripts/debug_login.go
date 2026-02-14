//go:build ignore

package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func main() {
    fmt.Println("=== 调试登录API ===")

    loginData := map[string]string{
        "username": "admin",
        "password": "admin123",
    }

    loginJSON, _ := json.Marshal(loginData)
    fmt.Printf("登录请求数据: %s\n", string(loginJSON))

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

    fmt.Printf("响应状态码: %d\n", resp.StatusCode)
    fmt.Printf("响应内容: %s\n", string(body))

    // 解析响应到map
    var response map[string]interface{}
    json.Unmarshal(body, &response)

    fmt.Printf("\n解析后的响应结构:\n")
    for key, value := range response {
        fmt.Printf("  %s: %v (%T)\n", key, value, value)
    }
}