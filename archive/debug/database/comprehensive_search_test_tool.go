package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type SearchResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       []Client    `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Size  int `json:"size"`
}

type TestCase struct {
	Name         string
	URL          string
	Expected     int
	Description  string
}

type TestResult struct {
	TestCase     TestCase
	Actual       int
	Success      bool
	ResponseTime time.Duration
	Error        error
	response     SearchResponseWithTime
}

func main() {
	fmt.Println("🔍 客户搜索功能端到端测试")
	fmt.Println("=====================================")

	// 从上一个脚本获取的令牌
	authToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNSwidXNlcm5hbWUiOiJzZWFyY2h0ZXN0QGxhdy1vYS5jb20iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NjA0MjI3NTksImlhdCI6MTc2MDMzNjM1OX0.XI56TAE0kporkOUEh2wKk7k0F8HWGQdPQcwT58q8yCQ"

	// 测试用例
	testCases := []TestCase{
		{
			Name:        "搜索'张三'(精确匹配)",
			URL:         "http://localhost:8080/api/v1/clients?search=张三&page=1&page_size=20",
			Expected:    1,
			Description: "应该只返回1条叫'张三'的记录",
		},
		{
			Name:        "搜索'张'(模糊匹配)",
			URL:         "http://localhost:8080/api/v1/clients?search=张&page=1&page_size=20",
			Expected:    1,
			Description: "应该返回名字包含'张'的记录",
		},
		{
			Name:        "搜索'王'(模糊匹配)",
			URL:         "http://localhost:8080/api/v1/clients?search=王&page=1&page_size=20",
			Expected:    1,
			Description: "应该返回名字包含'王'的记录",
		},
		{
			Name:        "搜索'不存在'",
			URL:         "http://localhost:8080/api/v1/clients?search=不存在&page=1&page_size=20",
			Expected:    0,
			Description: "应该返回0条记录",
		},
		{
			Name:        "筛选'个人'客户",
			URL:         "http://localhost:8080/api/v1/clients?type=个人&page=1&page_size=20",
			Expected:    -1, // 不确定具体数量，但应该>0
			Description: "应该只返回个人客户",
		},
		{
			Name:        "筛选'活跃'客户",
			URL:         "http://localhost:8080/api/v1/clients?status=active&page=1&page_size=20",
			Expected:    -1, // 不确定具体数量，但应该>0
			Description: "应该只返回活跃客户",
		},
		{
			Name:        "组合搜索:个人客户+张",
			URL:         "http://localhost:8080/api/v1/clients?type=个人&search=张&page=1&page_size=20",
			Expected:    1,
			Description: "应该返回个人客户中名字包含'张'的记录",
		},
		{
			Name:        "无搜索条件(获取所有)",
			URL:         "http://localhost:8080/api/v1/clients?page=1&page_size=20",
			Expected:    -1, // 不确定具体数量，但应该>0
			Description: "应该返回所有客户",
		},
	}

	fmt.Printf("📋 执行 %d 个测试用例\n\n", len(testCases))

	results := []TestResult{}

	for i, testCase := range testCases {
		fmt.Printf("🧪 测试 %d/%d: %s\n", i+1, len(testCases), testCase.Name)
		fmt.Printf("📡 URL: %s\n", testCase.URL)
		fmt.Printf("📝 期望: %d 条记录 (%s)\n", testCase.Expected, testCase.Description)

		result := executeTest(testCase, authToken)
		results = append(results, result)

		if result.Success {
			fmt.Printf("✅ 通过: %d 条记录 (耗时: %v)\n", result.Actual, result.ResponseTime)
		} else {
			fmt.Printf("❌ 失败: %v\n", result.Error)
		}

		// 检查响应数据
		if result.Actual > 0 && len(result.response.Data) > 0 {
			fmt.Printf("📄 返回的记录:\n")
			maxShow := 3
			if len(result.response.Data) < maxShow {
				maxShow = len(result.response.Data)
			}

			for j := 0; j < maxShow; j++ {
				client := result.response.Data[j]
				fmt.Printf("   %d. %s (%s) - %s\n", j+1, client.Name, client.Type, client.Status)
			}

			if len(result.response.Data) > maxShow {
				fmt.Printf("   ... 还有 %d 条记录\n", len(result.response.Data)-maxShow)
			}
		}

		fmt.Println(strings.Repeat("-", 60))
		time.Sleep(100 * time.Millisecond) // 避免请求过快
	}

	// 生成测试报告
	generateTestReport(results)
}

type SearchResponseWithTime struct {
	SearchResponse
	ResponseTime time.Duration
}

func executeTest(testCase TestCase, token string) TestResult {
	start := time.Now()

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", testCase.URL, nil)
	if err != nil {
		return TestResult{
			TestCase:     testCase,
			Actual:       0,
			Success:      false,
			ResponseTime: 0,
			Error:        fmt.Errorf("创建请求失败: %w", err),
		}
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		return TestResult{
			TestCase:     testCase,
			Actual:       0,
			Success:      false,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("请求失败: %w", err),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{
			TestCase:     testCase,
			Actual:       0,
			Success:      false,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("读取响应失败: %w", err),
		}
	}

	if resp.StatusCode != 200 {
		return TestResult{
			TestCase:     testCase,
			Actual:       0,
			Success:      false,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("API返回状态码: %d, 响应: %s", resp.StatusCode, string(body)),
		}
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return TestResult{
			TestCase:     testCase,
			Actual:       0,
			Success:      false,
			ResponseTime: responseTime,
			Error:        fmt.Errorf("JSON解析失败: %w", err),
		}
	}

	actual := len(searchResp.Data)
	success := false

	// 特殊处理预期数量为-1的情况（不确定具体数量）
	if testCase.Expected == -1 {
		success = actual > 0
	} else {
		success = actual == testCase.Expected
	}

	// 如果是"张三"测试，额外检查返回的记录名称是否正确
	if testCase.Name == "搜索'张三'(精确匹配)" && success {
		foundZhangsan := false
		for _, client := range searchResp.Data {
			if strings.Contains(client.Name, "张三") {
				foundZhangsan = true
				break
			}
		}
		success = foundZhangsan
	}

	responseWithTime := SearchResponseWithTime{
		SearchResponse: searchResp,
		ResponseTime:   responseTime,
	}

	return TestResult{
		TestCase:     testCase,
		Actual:       actual,
		Success:      success,
		ResponseTime: responseTime,
		response:     responseWithTime,
		Error:        nil,
	}
}

func generateTestReport(results []TestResult) {
	fmt.Println("\n📊 测试报告")
	fmt.Println("=====================================")

	total := len(results)
	passed := 0
	failed := 0
	var totalResponseTime time.Duration

	for _, result := range results {
		totalResponseTime += result.ResponseTime
		if result.Success {
			passed++
		} else {
			failed++
		}
	}

	avgResponseTime := totalResponseTime / time.Duration(total)
	successRate := float64(passed) / float64(total) * 100

	fmt.Printf("📈 总体统计:\n")
	fmt.Printf("   总测试数: %d\n", total)
	fmt.Printf("   通过: %d\n", passed)
	fmt.Printf("   失败: %d\n", failed)
	fmt.Printf("   成功率: %.1f%%\n", successRate)
	fmt.Printf("   平均响应时间: %v\n", avgResponseTime)

	fmt.Printf("\n🔍 详细结果:\n")
	for i, result := range results {
		status := "❌"
		if result.Success {
			status = "✅"
		}
		fmt.Printf("   %d. %s %s - %d条记录 (%v)\n",
			i+1, status, result.TestCase.Name, result.Actual, result.ResponseTime)
	}

	// 关键测试结果
	fmt.Printf("\n🎯 关键测试结果:\n")
	for _, result := range results {
		if strings.Contains(result.TestCase.Name, "张三") {
			status := "❌"
			if result.Success {
				status = "✅"
			}
			fmt.Printf("   %s 搜索'张三': %s - 实际返回%d条记录\n",
				status, result.TestCase.Name, result.Actual)
		}
	}

	// 性能分析
	fmt.Printf("\n⚡ 性能分析:\n")
	slowTests := []TestResult{}
	for _, result := range results {
		if result.ResponseTime > 100*time.Millisecond {
			slowTests = append(slowTests, result)
		}
	}

	if len(slowTests) > 0 {
		fmt.Printf("   慢查询 (>100ms): %d个\n", len(slowTests))
		for _, test := range slowTests {
			fmt.Printf("   - %s: %v\n", test.TestCase.Name, test.ResponseTime)
		}
	} else {
		fmt.Printf("   ✅ 所有测试响应时间都 < 100ms\n")
	}

	// 最终结论
	fmt.Printf("\n🎉 最终结论:\n")
	if successRate >= 90 {
		fmt.Printf("   ✅ 搜索功能测试通过！修复成功！\n")
	} else {
		fmt.Printf("   ❌ 搜索功能仍有问题，需要进一步调试\n")
	}
}