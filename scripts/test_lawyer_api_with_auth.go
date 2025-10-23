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

	fmt.Printf("登录响应状态码: %d\n", resp.StatusCode)
	if resp.StatusCode != 200 {
		fmt.Printf("登录失败: %s\n", string(body))
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

	fmt.Printf("✅ 获取JWT令牌成功: %s...\n", token[:min(len(token), 30)])

	// 3. 测试律师API（带认证）
	fmt.Println("\n🔍 测试律师API（带认证）...")
	req, err := http.NewRequest("GET", baseURL+"/lawfirm/lawyers?page=1&page_size=10", nil)
	if err != nil {
		log.Fatal("创建律师API请求失败:", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal("律师API请求失败:", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("读取律师API响应失败:", err)
	}

	fmt.Printf("律师API状态码: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		var lawyersResponse map[string]interface{}
		if err := json.Unmarshal(body, &lawyersResponse); err != nil {
			log.Fatal("解析律师API响应失败:", err)
		}

		fmt.Printf("\n✅ 律师API响应结构分析:\n")

		// 检查响应格式
		if code, ok := lawyersResponse["code"]; ok {
			fmt.Printf("  code: %v\n", code)
		}
		if message, ok := lawyersResponse["message"]; ok {
			fmt.Printf("  message: %v\n", message)
		}

		if data, ok := lawyersResponse["data"]; ok {
			fmt.Printf("  data type: %T\n", data)

			switch v := data.(type) {
			case []interface{}:
				fmt.Printf("  lawyers count: %d\n", len(v))
				if len(v) > 0 {
					fmt.Printf("  第一个律师结构:\n")
					if lawyer, ok := v[0].(map[string]interface{}); ok {
						for key, value := range lawyer {
							fmt.Printf("    %s: %v (%T)\n", key, value, value)
						}
					}
				}
			case map[string]interface{}:
				fmt.Printf("  data是对象类型，包含字段:\n")
				for key, value := range v {
					fmt.Printf("    %s: %v (%T)\n", key, value, value)

					// 如果是lawyers字段，进一步分析
					if key == "lawyers" && value != nil {
						if lawyers, ok := value.([]interface{}); ok {
							fmt.Printf("    lawyers数组长度: %d\n", len(lawyers))
							if len(lawyers) > 0 {
								fmt.Printf("    第一个律师详情:\n")
								if lawyer, ok := lawyers[0].(map[string]interface{}); ok {
									for k, val := range lawyer {
										fmt.Printf("      %s: %v (%T)\n", k, val, val)
									}
								}
							}
						}
					}
				}
			default:
				fmt.Printf("  data内容: %v\n", data)
			}
		}

		if meta, ok := lawyersResponse["meta"]; ok {
			fmt.Printf("  meta: %v\n", meta)
		}

		fmt.Printf("\n📄 完整响应JSON:\n")
		prettyJSON, _ := json.MarshalIndent(lawyersResponse, "", "  ")
		fmt.Println(string(prettyJSON))

	} else {
		fmt.Printf("❌ 律师API调用失败: %s\n", string(body))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}