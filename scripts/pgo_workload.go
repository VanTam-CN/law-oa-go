//go:build ignore

// PGO性能剖析工作负载脚本
// 用于生成Profile-Guided Optimization所需的性能剖析数据

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func mainPGOWorkload() {
	runPGOWorkload()
}

// 模拟请求负载配置
type WorkloadConfig struct {
	// 并发用户数
	ConcurrentUsers int `json:"concurrent_users"`
	// 每个用户的请求数
	RequestsPerUser int `json:"requests_per_user"`
	// 请求间隔（毫秒）
	RequestInterval int `json:"request_interval"`
	// 工作负载持续时间
	Duration time.Duration `json:"duration"`
	// 测试端点
	Endpoints []EndpointConfig `json:"endpoints"`
}

// 端点配置
type EndpointConfig struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Weight  float64           `json:"weight"` // 请求权重
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// 性能统计
type PerformanceStats struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	TotalDuration       time.Duration `json:"total_duration"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	RequestsPerSecond   float64       `json:"requests_per_second"`
}

// HTTP客户端
type httpClient struct {
	client  *http.Client
	headers map[string]string
}

// 工作负载生成器
type WorkloadGenerator struct {
	config WorkloadConfig
	stats  *PerformanceStats
	mu     sync.RWMutex
	client *httpClient
}

// runPGOWorkload 运行PGO工作负载
func runPGOWorkload() {
	// 配置工作负载
	config := WorkloadConfig{
		ConcurrentUsers: 50,
		RequestsPerUser: 100,
		RequestInterval: 10, // 10毫秒
		Duration:        5 * time.Minute,
		Endpoints: []EndpointConfig{
			{
				Path:   "/health",
				Method: "GET",
				Weight: 0.2, // 20%的健康检查请求
			},
			{
				Path:   "/api/v1/users",
				Method: "GET",
				Weight: 0.3, // 30%的用户查询请求
			},
			{
				Path:   "/api/v1/cases",
				Method: "GET",
				Weight: 0.3, // 30%的案件查询请求
			},
			{
				Path:   "/api/v1/clients",
				Method: "GET",
				Weight: 0.1, // 10%的客户查询请求
			},
			{
				Path:   "/performance/cache",
				Method: "GET",
				Weight: 0.1, // 10%的性能测试请求
			},
		},
	}

	log.Printf("开始PGO性能剖析工作负载...")
	log.Printf("配置: %+v", config)

	generator := NewWorkloadGenerator(config)

	// 执行工作负载
	err := generator.Run()
	if err != nil {
		log.Fatalf("工作负载执行失败: %v", err)
	}

	// 输出性能统计
	generator.PrintStats()
}

// 创建工作负载生成器
func NewWorkloadGenerator(config WorkloadConfig) *WorkloadGenerator {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			MaxConnsPerHost:     100,
			MaxIdleConnsPerHost: 10,
		},
	}

	return &WorkloadGenerator{
		config: config,
		stats: &PerformanceStats{
			MinResponseTime: time.Hour,
		},
		client: &httpClient{
			client: client,
			headers: map[string]string{
				"User-Agent": "PGO-Workload-Generator/1.0",
				"Accept":     "application/json",
			},
		},
	}
}

// 执行工作负载
func (wg *WorkloadGenerator) Run() error {
	log.Println("启动工作负载执行...")
	startTime := time.Now()
	wg.stats.TotalDuration = wg.config.Duration

	// 创建上下文用于控制执行时间
	ctx, cancel := context.WithTimeout(context.Background(), wg.config.Duration)
	defer cancel()

	// 使用WaitGroup等待所有用户完成
	var userWg sync.WaitGroup
	userWg.Add(wg.config.ConcurrentUsers)

	// 启动性能统计收集器
	go wg.collectStats(ctx)

	// 启动并发用户
	for i := 0; i < wg.config.ConcurrentUsers; i++ {
		go func(userID int) {
			defer userWg.Done()
			wg.runUserWorkload(ctx, userID)
		}(i)
	}

	// 等待所有用户完成或超时
	done := make(chan struct{})
	go func() {
		userWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("所有用户工作负载完成")
	case <-ctx.Done():
		log.Println("工作负载超时完成")
	}

	wg.stats.TotalDuration = time.Since(startTime)
	return nil
}

// 执行单个用户的工作负载
func (wg *WorkloadGenerator) runUserWorkload(ctx context.Context, userID int) {
	log.Printf("用户 %d 开始工作负载", userID)

	for i := 0; i < wg.config.RequestsPerUser; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			// 执行请求
			err := wg.executeRequest(ctx, userID, i)
			if err != nil {
				log.Printf("用户 %d 请求 %d 失败: %v", userID, i, err)
			}

			// 请求间隔
			if wg.config.RequestInterval > 0 {
				time.Sleep(time.Duration(wg.config.RequestInterval) * time.Millisecond)
			}
		}
	}
}

// 执行单个请求
func (wg *WorkloadGenerator) executeRequest(ctx context.Context, userID, requestID int) error {
	// 根据权重选择端点
	endpoint := wg.selectEndpoint()
	if endpoint == nil {
		return fmt.Errorf("没有可用的端点")
	}

	// 准备请求URL
	url := fmt.Sprintf("http://localhost:8080%s", endpoint.Path)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, endpoint.Method, url, nil)
	if err != nil {
		return err
	}

	// 设置请求头
	for key, value := range wg.client.headers {
		req.Header.Set(key, value)
	}
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	// 执行请求
	startTime := time.Now()
	resp, err := wg.client.client.Do(req)
	responseTime := time.Since(startTime)

	// 更新统计
	wg.mu.Lock()
	wg.stats.TotalRequests++
	if err == nil && resp.StatusCode < 400 {
		wg.stats.SuccessfulRequests++
	} else {
		wg.stats.FailedRequests++
	}

	if responseTime < wg.stats.MinResponseTime {
		wg.stats.MinResponseTime = responseTime
	}
	if responseTime > wg.stats.MaxResponseTime {
		wg.stats.MaxResponseTime = responseTime
	}
	wg.mu.Unlock()

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 记录响应时间用于性能分析
	log.Printf("用户 %d 请求 %d: %s %s - %d - %v",
		userID, requestID, endpoint.Method, endpoint.Path, resp.StatusCode, responseTime)

	return nil
}

// 根据权重选择端点
func (wg *WorkloadGenerator) selectEndpoint() *EndpointConfig {
	if len(wg.config.Endpoints) == 0 {
		return nil
	}

	// 计算总权重
	totalWeight := 0.0
	for _, endpoint := range wg.config.Endpoints {
		totalWeight += endpoint.Weight
	}

	// 生成随机数
	randomWeight := float64(time.Now().UnixNano()%1000000) / 1000000.0 * totalWeight

	// 选择端点
	currentWeight := 0.0
	for _, endpoint := range wg.config.Endpoints {
		currentWeight += endpoint.Weight
		if randomWeight <= currentWeight {
			return &endpoint
		}
	}

	return &wg.config.Endpoints[len(wg.config.Endpoints)-1]
}

// 收集性能统计
func (wg *WorkloadGenerator) collectStats(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			wg.printIntermediateStats()
		}
	}
}

// 打印中间统计
func (wg *WorkloadGenerator) printIntermediateStats() {
	wg.mu.RLock()
	defer wg.mu.RUnlock()

	if wg.stats.TotalRequests == 0 {
		return
	}

	avgResponseTime := time.Duration(0)
	if wg.stats.SuccessfulRequests > 0 {
		avgResponseTime = wg.stats.TotalDuration / time.Duration(wg.stats.SuccessfulRequests)
	}

	rps := float64(wg.stats.TotalRequests) / wg.stats.TotalDuration.Seconds()

	log.Printf("中间统计 - 总请求: %d, 成功: %d, 失败: %d, 平均响应时间: %v, RPS: %.2f",
		wg.stats.TotalRequests, wg.stats.SuccessfulRequests, wg.stats.FailedRequests,
		avgResponseTime, rps)
}

// 打印最终统计
func (wg *WorkloadGenerator) PrintStats() {
	wg.mu.RLock()
	defer wg.mu.RUnlock()

	if wg.stats.TotalRequests == 0 {
		log.Println("没有执行的请求")
		return
	}

	// 计算统计指标
	wg.stats.AverageResponseTime = wg.stats.TotalDuration / time.Duration(wg.stats.SuccessfulRequests)
	wg.stats.RequestsPerSecond = float64(wg.stats.TotalRequests) / wg.stats.TotalDuration.Seconds()

	// 打印详细统计
	fmt.Println("\n=== PGO工作负载性能统计 ===")
	fmt.Printf("总请求数: %d\n", wg.stats.TotalRequests)
	fmt.Printf("成功请求数: %d\n", wg.stats.SuccessfulRequests)
	fmt.Printf("失败请求数: %d\n", wg.stats.FailedRequests)
	fmt.Printf("成功率: %.2f%%\n", float64(wg.stats.SuccessfulRequests)/float64(wg.stats.TotalRequests)*100)
	fmt.Printf("总执行时间: %v\n", wg.stats.TotalDuration)
	fmt.Printf("平均响应时间: %v\n", wg.stats.AverageResponseTime)
	fmt.Printf("最小响应时间: %v\n", wg.stats.MinResponseTime)
	fmt.Printf("最大响应时间: %v\n", wg.stats.MaxResponseTime)
	fmt.Printf("请求/秒: %.2f\n", wg.stats.RequestsPerSecond)
	fmt.Println("================================")

	// 输出JSON格式的统计用于进一步分析
	statsJSON, _ := json.MarshalIndent(wg.stats, "", "  ")
	fmt.Printf("\nJSON统计:\n%s\n", statsJSON)
}
