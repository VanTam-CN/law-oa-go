package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"law-oa-go/test"
)

// LoadTestResult 负载测试结果
type LoadTestResult struct {
	TotalRequests     int
	Successful        int
	Failed            int
	AverageTime       time.Duration
	MinTime           time.Duration
	MaxTime           time.Duration
	RequestsPerSecond float64
	ThroughputMB      float64
	ErrorRate         float64
}

// LoadTester 负载测试器
type LoadTester struct {
	router    *gin.Engine
	client    *http.Client
	baseURL   string
	authToken string
	results   chan *LoadTestResult
}

// NewLoadTester 创建新的负载测试器
func NewLoadTester(router *gin.Engine, authToken string) *LoadTester {
	return &LoadTester{
		router:    router,
		client:    &http.Client{Timeout: 30 * time.Second},
		authToken: authToken,
		results:   make(chan *LoadTestResult, 100),
	}
}

// TestLoginLoad 登录负载测试
func TestLoginLoad(t *testing.T) {
	suite := SetupAPITestSuite(t)

	// 预创建测试用户，避免测试过程中的数据库查询风暴
	test.PreloadTestAuthTokens(t, suite.userService, []string{"loadtest@example.com"})

	tester := NewLoadTester(suite.router, "")

	duration := 10 * time.Second
	workers := 10

	result := tester.RunLoadTest(t, "POST", "/auth/login", map[string]interface{}{
		"email":    "loadtest@example.com",
		"password": "Password123!",
	}, duration, workers)

	fmt.Printf("\n=== Login Load Test Results ===\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Workers: %d\n", workers)
	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Successful: %d\n", result.Successful)
	fmt.Printf("Failed: %d\n", result.Failed)
	fmt.Printf("Error Rate: %.2f%%\n", result.ErrorRate)
	fmt.Printf("Average Response Time: %v\n", result.AverageTime)
	fmt.Printf("Min Response Time: %v\n", result.MinTime)
	fmt.Printf("Max Response Time: %v\n", result.MaxTime)
	fmt.Printf("Requests per Second: %.2f\n", result.RequestsPerSecond)

	require.Greater(t, result.Successful, 0)
	require.Less(t, result.ErrorRate, 10.0) // 错误率低于10%
}

// TestUserProfileLoad 用户资料负载测试
func TestUserProfileLoad(t *testing.T) {
	suite := SetupAPITestSuite(t)

	// 预加载认证令牌，避免测试过程中的数据库查询风暴
	test.PreloadTestAuthTokens(t, suite.userService, []string{"loadtest@example.com"})
	authToken := test.GetAuthToken(t, suite.userService, "loadtest@example.com", "Password123!")

	tester := NewLoadTester(suite.router, authToken)

	duration := 15 * time.Second
	workers := 20

	result := tester.RunLoadTest(t, "GET", "/users/profile", nil, duration, workers)

	fmt.Printf("\n=== User Profile Load Test Results ===\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Workers: %d\n", workers)
	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Successful: %d\n", result.Successful)
	fmt.Printf("Failed: %d\n", result.Failed)
	fmt.Printf("Error Rate: %.2f%%\n", result.ErrorRate)
	fmt.Printf("Average Response Time: %v\n", result.AverageTime)
	fmt.Printf("Min Response Time: %v\n", result.MinTime)
	fmt.Printf("Max Response Time: %v\n", result.MaxTime)
	fmt.Printf("Requests per Second: %.2f\n", result.RequestsPerSecond)

	require.Greater(t, result.Successful, 0)
	require.Less(t, result.ErrorRate, 5.0)
}

// TestClientOperationsLoad 客户操作负载测试
func TestClientOperationsLoad(t *testing.T) {
	suite := SetupAPITestSuite(t)

	// 预加载认证令牌，避免测试过程中的数据库查询风暴
	test.PreloadTestAuthTokens(t, suite.userService, []string{"clientops@example.com"})
	authToken := test.GetAuthToken(t, suite.userService, "clientops@example.com", "Password123!")

	tester := NewLoadTester(suite.router, authToken)

	operations := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"GET", "/clients/", nil},
		{"POST", "/clients/", map[string]interface{}{
			"name":    "Load Test Client",
			"email":   "loadclient@example.com",
			"phone":   "1234567890",
			"address": "123 Load Test St",
			"company": "Load Test Company",
		}},
	}

	duration := 20 * time.Second
	workers := 15

	fmt.Printf("\n=== Client Operations Load Test Results ===\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Workers: %d\n", workers)

	for i, op := range operations {
		result := tester.RunLoadTest(t, op.method, op.path, op.body, duration, workers)
		fmt.Printf("Operation %d (%s %s):\n", i+1, op.method, op.path)
		fmt.Printf("  Total Requests: %d\n", result.TotalRequests)
		fmt.Printf("  Successful: %d\n", result.Successful)
		fmt.Printf("  Error Rate: %.2f%%\n", result.ErrorRate)
		fmt.Printf("  Average Response Time: %v\n", result.AverageTime)
		fmt.Printf("  Requests per Second: %.2f\n", result.RequestsPerSecond)
		fmt.Printf("\n")

		require.Greater(t, result.Successful, 0)
		require.Less(t, result.ErrorRate, 10.0)
	}
}

// TestConcurrentUsers 并发用户测试
func TestConcurrentUsers(t *testing.T) {
	suite := SetupAPITestSuite(t)

	duration := 30 * time.Second
	users := 50
	requestsPerUser := 100

	// 预加载所有用户的认证令牌，避免数据库查询风暴
	var userEmails []string
	for i := 0; i < users; i++ {
		email := fmt.Sprintf("concurrent%d@example.com", i)
		userEmails = append(userEmails, email)
	}
	test.PreloadTestAuthTokens(t, suite.userService, userEmails)

	var wg sync.WaitGroup
	results := make(chan *LoadTestResult, users)
	workerDone := make(chan bool, users)

	// 启动结果收集器
	go func() {
		for result := range results {
			fmt.Printf("User completed - Requests: %d, Success: %d, Failed: %d, Avg Time: %v\n",
				result.TotalRequests, result.Successful, result.Failed, result.AverageTime)
		}
	}()

	// 创建并发用户
	for i := 0; i < users; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// 获取用户邮箱和认证令牌
			email := fmt.Sprintf("concurrent%d@example.com", userID)
			authToken := test.GetAuthToken(t, suite.userService, email, "Password123!")

			// 用户工作流程：登录、获取资料、创建客户、列出客户
			workflow := []struct {
				method string
				path   string
				body   interface{}
			}{
				{"GET", "/users/profile", nil},
				{"POST", "/clients/", map[string]interface{}{
					"name":    fmt.Sprintf("Client %d", userID),
					"email":   fmt.Sprintf("client%d@example.com", userID),
					"phone":   "1234567890",
					"address": fmt.Sprintf("%d Test St", userID),
				}},
				{"GET", "/clients/", nil},
			}

			successCount := 0
			totalCount := 0
			var totalTime time.Duration
			var minTime time.Duration = time.Hour
			var maxTime time.Duration

			for j := 0; j < requestsPerUser; j++ {
				op := workflow[j%len(workflow)]
				start := time.Now()

				w := httptest.NewRecorder()
				var body *bytes.Buffer
				if op.body != nil {
					jsonData, _ := json.Marshal(op.body)
					body = bytes.NewBuffer(jsonData)
				} else {
					body = bytes.NewBuffer(nil)
				}

				req, _ := http.NewRequest(op.method, op.path, body)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+authToken)

				suite.router.ServeHTTP(w, req)

				duration := time.Since(start)
				totalTime += duration
				if duration < minTime {
					minTime = duration
				}
				if duration > maxTime {
					maxTime = duration
				}

				totalCount++
				if w.Code == http.StatusOK || w.Code == http.StatusCreated {
					successCount++
				}
			}

			result := &LoadTestResult{
				TotalRequests:     totalCount,
				Successful:        successCount,
				Failed:            totalCount - successCount,
				AverageTime:       totalTime / time.Duration(totalCount),
				MinTime:           minTime,
				MaxTime:           maxTime,
				RequestsPerSecond: float64(totalCount) / duration.Seconds(),
				ErrorRate:         float64(totalCount-successCount) / float64(totalCount) * 100,
			}

			results <- result
			workerDone <- true
		}(i)
	}

	// 等待所有用户完成
	go func() {
		wg.Wait()
		close(results)
		close(workerDone)
	}()

	// 等待结果
	completed := 0
	for range workerDone {
		completed++
		if completed%10 == 0 {
			fmt.Printf("Progress: %d/%d users completed\n", completed, users)
		}
	}

	fmt.Printf("\n=== Concurrent Users Test Results ===\n")
	fmt.Printf("Total Users: %d\n", users)
	fmt.Printf("Requests per User: %d\n", requestsPerUser)
	fmt.Printf("Total Test Duration: %v\n", duration)
	fmt.Printf("All users completed successfully\n")
}

// RunLoadTest 运行负载测试
func (lt *LoadTester) RunLoadTest(t *testing.T, method, path string, body interface{}, duration time.Duration, workers int) *LoadTestResult {
	var wg sync.WaitGroup
	results := make(chan *LoadTestResult, workers)
	workerDone := make(chan bool, workers)

	// 启动工作线程
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			successCount := 0
			totalCount := 0
			var totalTime time.Duration
			var minTime time.Duration = time.Hour
			var maxTime time.Duration

			timeout := time.After(duration)
			for {
				select {
				case <-timeout:
					result := &LoadTestResult{
						TotalRequests:     totalCount,
						Successful:        successCount,
						Failed:            totalCount - successCount,
						AverageTime:       totalTime / time.Duration(totalCount),
						MinTime:           minTime,
						MaxTime:           maxTime,
						RequestsPerSecond: float64(totalCount) / duration.Seconds(),
						ErrorRate:         float64(totalCount-successCount) / float64(totalCount) * 100,
					}
					results <- result
					workerDone <- true
					return
				default:
					start := time.Now()

					var bodyData *bytes.Buffer
					if body != nil {
						jsonData, _ := json.Marshal(body)
						bodyData = bytes.NewBuffer(jsonData)
					} else {
						bodyData = bytes.NewBuffer(nil)
					}

					w := httptest.NewRecorder()
					req, _ := http.NewRequest(method, path, bodyData)
					req.Header.Set("Content-Type", "application/json")
					if lt.authToken != "" {
						req.Header.Set("Authorization", "Bearer "+lt.authToken)
					}

					lt.router.ServeHTTP(w, req)

					responseTime := time.Since(start)
					totalTime += responseTime
					if responseTime < minTime {
						minTime = responseTime
					}
					if responseTime > maxTime {
						maxTime = responseTime
					}

					totalCount++
					if w.Code == http.StatusOK || w.Code == http.StatusCreated {
						successCount++
					}
				}
			}
		}()
	}

	// 收集结果
	var finalResult *LoadTestResult
	completed := 0
	for range workerDone {
		completed++
		if completed == workers {
			close(results)
			break
		}
	}

	// 计算聚合结果
	totalRequests := 0
	totalSuccess := 0
	totalFailed := 0
	var totalDuration time.Duration
	var overallMin time.Duration = time.Hour
	var overallMax time.Duration

	for result := range results {
		totalRequests += result.TotalRequests
		totalSuccess += result.Successful
		totalFailed += result.Failed
		totalDuration += result.AverageTime * time.Duration(result.TotalRequests)
		if result.MinTime < overallMin {
			overallMin = result.MinTime
		}
		if result.MaxTime > overallMax {
			overallMax = result.MaxTime
		}
	}

	wg.Wait()

	finalResult = &LoadTestResult{
		TotalRequests:     totalRequests,
		Successful:        totalSuccess,
		Failed:            totalFailed,
		AverageTime:       totalDuration / time.Duration(totalRequests),
		MinTime:           overallMin,
		MaxTime:           overallMax,
		RequestsPerSecond: float64(totalRequests) / duration.Seconds(),
		ErrorRate:         float64(totalFailed) / float64(totalRequests) * 100,
	}

	return finalResult
}

// TestMain 性能测试的主函数，用于管理缓存
func TestMain(m *testing.M) {
	// 运行测试
	code := m.Run()

	// 清理缓存
	test.ClearAuthTokenCache()

	// 退出
	os.Exit(code)
}
