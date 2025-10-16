package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestCaseCreationIntegration 集成测试：验证案件创建功能的完整流程
func TestCaseCreationIntegration(t *testing.T) {
	// 这个测试需要服务器运行，所以跳过
	if testing.Short() {
		t.Skip("跳过集成测试")
		return
	}

	baseURL := getBaseURL()
	if baseURL == "" {
		t.Skip("服务器未运行，跳过集成测试")
	}

	// 测试用例
	testCases := []struct {
		name           string
		requestData    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功创建案件",
			requestData: map[string]interface{}{
				"title":       "集成测试案件",
				"description": "这是一个集成测试案件",
				"client_id":   1,
				"lawyer_id":   1,
				"case_type":   "civil",
				"priority":    "medium",
				"status":      "pending",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "验证失败 - 空标题",
			requestData: map[string]interface{}{
				"title":       "",
				"description": "空标题测试",
				"client_id":   1,
				"lawyer_id":   1,
				"case_type":   "civil",
				"priority":    "medium",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "验证失败 - 缺少必填字段",
			requestData: map[string]interface{}{
				"title":      "缺少字段测试",
				"case_type":  "civil",
				"priority":   "medium",
				// 缺少 client_id 和 lawyer_id
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "验证失败 - 无效枚举值",
			requestData: map[string]interface{}{
				"title":       "无效枚举值测试",
				"description": "测试无效枚举值",
				"client_id":   1,
				"lawyer_id":   1,
				"case_type":   "invalid_type",
				"priority":    "invalid_priority",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "前端兼容性测试",
			requestData: map[string]interface{}{
				// 使用前端可能发送的格式
				"title":       "前端兼容性测试",
				"description": "测试前端数据格式兼容性",
				"client_id":   1,
				"lawyer_id":   1,
				"case_type":   "commercial",
				"priority":    "high",
				"status":      "active",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 准备请求
			body, err := json.Marshal(tc.requestData)
			if err != nil {
				t.Fatalf("序列化请求数据失败: %v", err)
			}

			// 发送请求
			resp, err := sendCreateCaseRequest(baseURL, body)
			if err != nil {
				t.Fatalf("发送请求失败: %v", err)
			}
			defer resp.Body.Close()

			// 检查状态码
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("期望状态码 %d，实际得到 %d", tc.expectedStatus, resp.StatusCode)
			}

			// 读取响应
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("读取响应失败: %v", err)
			}

			// 解析响应
			var response map[string]interface{}
			if err := json.Unmarshal(respBody, &response); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}

			// 验证响应结构
			if success, ok := response["success"].(bool); !ok {
				t.Error("响应中缺少success字段")
			} else if tc.expectedStatus == http.StatusOK && !success {
				t.Error("期望成功但响应显示失败")
			} else if tc.expectedStatus != http.StatusOK && success {
				t.Error("期望失败但响应显示成功")
			}

			// 检查错误信息
			if tc.expectedError != "" {
				if errorObj, ok := response["error"].(map[string]interface{}); ok {
					if code, ok := errorObj["code"].(string); ok && code != tc.expectedError {
						t.Errorf("期望错误代码 %s，实际得到 %s", tc.expectedError, code)
					}
				} else {
					t.Errorf("期望错误对象但得到: %v", response["error"])
				}
			}

			// 记录测试结果
			t.Logf("测试用例 '%s' - 状态码: %d, 成功: %v", tc.name, resp.StatusCode, response["success"])

			// 如果成功创建，打印案件信息
			if resp.StatusCode == http.StatusOK && response["data"] != nil {
				if caseData, ok := response["data"].(map[string]interface{}); ok {
					t.Logf("创建的案件ID: %v, 标题: %v", caseData["id"], caseData["title"])
				}
			}
		})
	}
}

// getBaseURL 获取服务器基础URL
func getBaseURL() string {
	if url := os.Getenv("TEST_SERVER_URL"); url != "" {
		return url
	}
	return "http://localhost:8080/api/v1"
}

// sendCreateCaseRequest 发送创建案件请求
func sendCreateCaseRequest(baseURL string, body []byte) (*http.Response, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", baseURL+"/cases", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 如果需要认证，添加token
	if token := os.Getenv("TEST_AUTH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return client.Do(req)
}

// TestAPIServerConnectivity 测试服务器连通性
func TestAPIServerConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过连通性测试")
		return
	}

	baseURL := getBaseURL()
	if baseURL == "" {
		t.Skip("未配置服务器URL")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 测试健康检查端点
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("无法连接到服务器: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("健康检查失败，状态码: %d", resp.StatusCode)
	}

	t.Logf("服务器连通性正常，状态码: %d", resp.StatusCode)
}

// TestMain 测试入口点
func TestMain(m *testing.M) {
	// 如果需要，可以在这里设置测试环境
	// 例如：启动测试数据库、清理测试数据等

	// 运行测试
	code := m.Run()

	// 清理测试环境
	// 例如：关闭测试数据库连接、清理临时文件等

	os.Exit(code)
}