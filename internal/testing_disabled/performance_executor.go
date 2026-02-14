package testing

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceExecutor 性能测试执行器
type PerformanceExecutor struct {
	*BaseExecutor
	workers    int
	rampUpTime time.Duration
	duration   time.Duration
	thinkTime  time.Duration
	stats      *PerformanceStats
	mu         sync.RWMutex
}

// PerformanceStats 性能统计数据
type PerformanceStats struct {
	// 请求统计
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64

	// 响应时间统计
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	TotalResponseTime int64

	// 并发统计
	ConcurrentRequests int64
	PeakConcurrent    int64

	// 错误统计
	ErrorsByType map[string]int64
	StatusCodes  map[int]int64

	// 吞吐量统计
	StartTime time.Time
	EndTime   time.Time

	// 资源使用统计
	InitialMemory int64
	PeakMemory    int64
	CurrentMemory int64
	InitialGC     uint32
	CurrentGC     uint32

	mu sync.RWMutex
}

// PerformanceWorker 性能测试工作者
type PerformanceWorker struct {
	id         int
	executor   *PerformanceExecutor
	httpClient *http.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

// PerformanceScenario 性能测试场景
type PerformanceScenario struct {
	Name        string
	Weight      int
	Requests    []*TestCase
	Setup       []TestStep
	Teardown    []TestStep
	ThinkTime   time.Duration
	Pacing      time.Duration
}

// NewPerformanceExecutor 创建性能测试执行器
func NewPerformanceExecutor(options *TestExecutorOptions, logger TestLogger, metrics TestMetrics) *PerformanceExecutor {
	base := NewBaseExecutor(options, logger, metrics)

	return &PerformanceExecutor{
		BaseExecutor: base,
		workers:      options.MaxConcurrent,
		rampUpTime:   options.RampUpTime,
		duration:     options.Duration,
		thinkTime:    options.ThinkTime,
		stats: &PerformanceStats{
			ErrorsByType: make(map[string]int64),
			StatusCodes:  make(map[int]int64),
		},
	}
}

// GetExecutorType 获取执行器类型
func (e *PerformanceExecutor) GetExecutorType() TestType {
	return TestTypePerformance
}

// Setup 设置性能测试执行器
func (e *PerformanceExecutor) Setup(ctx context.Context, executionCtx *ExecutionContext) error {
	// 调用基础设置
	if err := e.BaseExecutor.Setup(ctx, executionCtx); err != nil {
		return err
	}

	// 初始化统计
	e.stats.mu.Lock()
	e.stats.StartTime = time.Now()
	e.stats.InitialMemory = e.getCurrentMemory()
	e.stats.InitialGC = e.getCurrentGC()
	e.stats.mu.Unlock()

	e.logger.Info("Performance executor setup completed",
		"workers", e.workers,
		"ramp_up_time", e.rampUpTime,
		"duration", e.duration,
		"think_time", e.thinkTime)

	return nil
}

// Teardown 清理性能测试执行器
func (e *PerformanceExecutor) Teardown(ctx context.Context, executionCtx *ExecutionContext) error {
	defer e.BaseExecutor.Teardown(ctx, executionCtx)

	// 记录结束时间
	e.stats.mu.Lock()
	e.stats.EndTime = time.Now()
	e.stats.CurrentMemory = e.getCurrentMemory()
	e.stats.CurrentGC = e.getCurrentGC()
	e.stats.mu.Unlock()

	// 输出性能报告
	e.generatePerformanceReport()

	e.logger.Info("Performance executor teardown completed")
	return nil
}

// executeMainTest 执行性能测试的主逻辑
func (e *PerformanceExecutor) executeMainTest(ctx context.Context, test *TestCase, result *TestResult, executionCtx *ExecutionContext) error {
	e.logger.Info("Executing performance test", "name", test.Name)

	// 重置统计
	e.resetStats()

	// 创建工作池
	workers := make([]*PerformanceWorker, e.workers)
	for i := 0; i < e.workers; i++ {
		workerCtx, cancel := context.WithCancel(ctx)
		workers[i] = &PerformanceWorker{
			id:       i + 1,
			executor: e,
			httpClient: &http.Client{
				Timeout: test.Timeout,
			},
			ctx:    workerCtx,
			cancel: cancel,
		}
	}

	// 启动工作池
	var wg sync.WaitGroup
	errChan := make(chan error, e.workers)

	// 渐进式启动工作者
	if e.rampUpTime > 0 {
		rampUpDelay := e.rampUpTime / time.Duration(e.workers)
		for i, worker := range workers {
			wg.Add(1)
			go func(w *PerformanceWorker, delay time.Duration) {
				defer wg.Done()
				time.Sleep(delay)
				if err := w.runTest(test, executionCtx); err != nil {
					errChan <- err
				}
			}(worker, time.Duration(i)*rampUpDelay)
		}
	} else {
		// 同时启动所有工作者
		for _, worker := range workers {
			wg.Add(1)
			go func(w *PerformanceWorker) {
				defer wg.Done()
				if err := w.runTest(test, executionCtx); err != nil {
					errChan <- err
				}
			}(worker)
		}
	}

	// 设置测试持续时间
	if e.duration > 0 {
		go func() {
			time.Sleep(e.duration)
			for _, worker := range workers {
				worker.cancel()
			}
		}()
	}

	// 等待所有工作者完成
	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			e.logger.Warn("Worker encountered error", "error", err)
		}
	}

	// 收集性能指标
	metrics := e.collectMetrics()
	result.Metadata["performance_metrics"] = metrics
	result.Metadata["performance_stats"] = e.stats

	return nil
}

// runTest 运行测试
func (w *PerformanceWorker) runTest(test *TestCase, executionCtx *ExecutionContext) error {
	for {
		select {
		case <-w.ctx.Done():
			return nil
		default:
			startTime := time.Now()

			// 增加并发计数
			atomic.AddInt64(&w.executor.stats.ConcurrentRequests, 1)
			defer atomic.AddInt64(&w.executor.stats.ConcurrentRequests, -1)

			// 执行请求
			resp, err := w.executeRequest(test, executionCtx)
			if err != nil {
				w.executor.recordError(err)
				w.executor.recordRequest(false, 0, 0)
				continue
			}

			// 记录成功请求
			responseTime := time.Since(startTime)
			w.executor.recordRequest(true, resp.StatusCode, responseTime)

			// 思考时间
			if w.executor.thinkTime > 0 {
				select {
				case <-w.ctx.Done():
					return nil
				case <-time.After(w.executor.thinkTime):
					// 继续下一次请求
				}
			}
		}
	}
}

// executeRequest 执行单个请求
func (w *PerformanceWorker) executeRequest(test *TestCase, executionCtx *ExecutionContext) (*http.Response, error) {
	// 构建请求
	req, err := http.NewRequest(test.Method, test.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range test.Headers {
		req.Header.Set(key, value)
	}

	// 执行请求
	return w.httpClient.Do(req.WithContext(w.ctx))
}

// recordRequest 记录请求结果
func (e *PerformanceExecutor) recordRequest(success bool, statusCode int, responseTime time.Duration) {
	atomic.AddInt64(&e.stats.TotalRequests, 1)

	if success && statusCode >= 200 && statusCode < 400 {
		atomic.AddInt64(&e.stats.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&e.stats.FailedRequests, 1)
	}

	// 记录响应时间
	atomic.AddInt64(&e.stats.TotalResponseTime, responseTime.Nanoseconds())

	// 更新最小响应时间
	for {
		current := e.stats.MinResponseTime
		if current == 0 || responseTime < current {
			if atomic.CompareAndSwapInt64((*int64)(&e.stats.MinResponseTime), int64(current), int64(responseTime)) {
				break
			}
		} else {
			break
		}
	}

	// 更新最大响应时间
	for {
		current := e.stats.MaxResponseTime
		if responseTime > current {
			if atomic.CompareAndSwapInt64((*int64)(&e.stats.MaxResponseTime), int64(current), int64(responseTime)) {
				break
			}
		} else {
			break
		}
	}

	// 记录状态码
	e.stats.mu.Lock()
	e.stats.StatusCodes[statusCode]++
	e.stats.mu.Unlock()

	// 更新峰值并发
	current := atomic.LoadInt64(&e.stats.ConcurrentRequests)
	for {
		peak := atomic.LoadInt64(&e.stats.PeakConcurrent)
		if current > peak {
			if atomic.CompareAndSwapInt64(&e.stats.PeakConcurrent, peak, current) {
				break
			}
		} else {
			break
		}
	}

	// 更新内存使用
	currentMemory := e.getCurrentMemory()
	for {
		peak := atomic.LoadInt64(&e.stats.PeakMemory)
		if currentMemory > peak {
			if atomic.CompareAndSwapInt64(&e.stats.PeakMemory, peak, currentMemory) {
				break
			}
		} else {
			break
		}
	}
}

// recordError 记录错误
func (e *PerformanceExecutor) recordError(err error) {
	errorType := fmt.Sprintf("%T", err)
	e.stats.mu.Lock()
	e.stats.ErrorsByType[errorType]++
	e.stats.mu.Unlock()
}

// resetStats 重置统计
func (e *PerformanceExecutor) resetStats() {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()

	e.stats.TotalRequests = 0
	e.stats.SuccessRequests = 0
	e.stats.FailedRequests = 0
	e.stats.MinResponseTime = 0
	e.stats.MaxResponseTime = 0
	e.stats.TotalResponseTime = 0
	e.stats.ConcurrentRequests = 0
	e.stats.PeakConcurrent = 0
	e.stats.StatusCodes = make(map[int]int64)
	e.stats.ErrorsByType = make(map[string]int64)
}

// collectMetrics 收集性能指标
func (e *PerformanceExecutor) collectMetrics() map[string]interface{} {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()

	totalRequests := atomic.LoadInt64(&e.stats.TotalRequests)
	successRequests := atomic.LoadInt64(&e.stats.SuccessRequests)
	failedRequests := atomic.LoadInt64(&e.stats.FailedRequests)
	totalResponseTime := atomic.LoadInt64(&e.stats.TotalResponseTime)

	metrics := map[string]interface{}{
		"total_requests":      totalRequests,
		"success_requests":    successRequests,
		"failed_requests":     failedRequests,
		"success_rate":        0.0,
		"failure_rate":        0.0,
		"avg_response_time":   0.0,
		"min_response_time":   e.stats.MinResponseTime,
		"max_response_time":   e.stats.MaxResponseTime,
		"peak_concurrent":     atomic.LoadInt64(&e.stats.PeakConcurrent),
		"status_codes":        e.stats.StatusCodes,
		"errors_by_type":      e.stats.ErrorsByType,
		"initial_memory_kb":   e.stats.InitialMemory / 1024,
		"peak_memory_kb":      atomic.LoadInt64(&e.stats.PeakMemory) / 1024,
		"current_memory_kb":   e.getCurrentMemory() / 1024,
		"initial_gc_cycles":   e.stats.InitialGC,
		"current_gc_cycles":   e.getCurrentGC(),
		"duration_seconds":    e.stats.EndTime.Sub(e.stats.StartTime).Seconds(),
	}

	// 计算成功率
	if totalRequests > 0 {
		metrics["success_rate"] = float64(successRequests) / float64(totalRequests) * 100
		metrics["failure_rate"] = float64(failedRequests) / float64(totalRequests) * 100
	}

	// 计算平均响应时间
	if totalRequests > 0 {
		metrics["avg_response_time"] = time.Duration(totalResponseTime/totalRequests).Seconds()
	}

	// 计算吞吐量 (RPS)
	duration := e.stats.EndTime.Sub(e.stats.StartTime)
	if duration > 0 {
		metrics["requests_per_second"] = float64(totalRequests) / duration.Seconds()
	}

	return metrics
}

// generatePerformanceReport 生成性能报告
func (e *PerformanceExecutor) generatePerformanceReport() {
	metrics := e.collectMetrics()

	e.logger.Info("=== Performance Test Report ===")
	e.logger.Info("Total Requests", "count", metrics["total_requests"])
	e.logger.Info("Success Rate", "percent", metrics["success_rate"])
	e.logger.Info("Average Response Time", "seconds", metrics["avg_response_time"])
	e.logger.Info("Min/Max Response Time", "min", metrics["min_response_time"], "max", metrics["max_response_time"])
	e.logger.Info("Peak Concurrent", "count", metrics["peak_concurrent"])
	e.logger.Info("Requests Per Second", "rps", metrics["requests_per_second"])
	e.logger.Info("Memory Usage", "initial_kb", metrics["initial_memory_kb"], "peak_kb", metrics["peak_memory_kb"])
	e.logger.Info("GC Cycles", "initial", metrics["initial_gc_cycles"], "current", metrics["current_gc_cycles"])
	e.logger.Info("Status Codes", "codes", metrics["status_codes"])
	if len(metrics["errors_by_type"].(map[string]int64)) > 0 {
		e.logger.Info("Errors by Type", "errors", metrics["errors_by_type"])
	}
	e.logger.Info("=== End Performance Report ===")
}

// getCurrentMemory 获取当前内存使用量
func (e *PerformanceExecutor) getCurrentMemory() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// getCurrentGC 获取当前GC周期数
func (e *PerformanceExecutor) getCurrentGC() uint32 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.NumGC
}

// GetStats 获取性能统计
func (e *PerformanceExecutor) GetStats() *PerformanceStats {
	return e.stats
}

// SetWorkers 设置工作者数量
func (e *PerformanceExecutor) SetWorkers(workers int) {
	e.workers = workers
}

// SetRampUpTime 设置渐进启动时间
func (e *PerformanceExecutor) SetRampUpTime(rampUpTime time.Duration) {
	e.rampUpTime = rampUpTime
}

// SetDuration 设置测试持续时间
func (e *PerformanceExecutor) SetDuration(duration time.Duration) {
	e.duration = duration
}

// SetThinkTime 设置思考时间
func (e *PerformanceExecutor) SetThinkTime(thinkTime time.Duration) {
	e.thinkTime = thinkTime
}

// ExecuteNavigate 执行导航步骤（性能测试中不适用）
func (e *PerformanceExecutor) ExecuteNavigate(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	return fmt.Errorf("navigate step not applicable for performance tests")
}

// ExecuteClick 执行点击步骤（性能测试中不适用）
func (e *PerformanceExecutor) ExecuteClick(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	return fmt.Errorf("click step not applicable for performance tests")
}

// ExecuteFill 执行填充步骤（性能测试中不适用）
func (e *PerformanceExecutor) ExecuteFill(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	return fmt.Errorf("fill step not applicable for performance tests")
}

// ExecuteWait 执行等待步骤
func (e *PerformanceExecutor) ExecuteWait(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	if step.Wait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(step.Wait):
			// 等待完成
		}
	}
	return nil
}

// ExecuteScreenshot 执行截图步骤（性能测试中不适用）
func (e *PerformanceExecutor) ExecuteScreenshot(ctx context.Context, step TestStep, result *TestResult) error {
	return fmt.Errorf("screenshot step not applicable for performance tests")
}

// ExecuteJavaScript 执行JavaScript步骤（性能测试中不适用）
func (e *PerformanceExecutor) ExecuteJavaScript(ctx context.Context, step TestStep, result *TestResult, executionCtx *ExecutionContext) error {
	return fmt.Errorf("javascript step not applicable for performance tests")
}

// ExecuteStep 执行单个步骤
func (e *PerformanceExecutor) ExecuteStep(ctx context.Context, step TestStep, executionCtx *ExecutionContext) error {
	switch step.Type {
	case "navigate":
		return e.ExecuteNavigate(ctx, step, executionCtx)
	case "click":
		return e.ExecuteClick(ctx, step, executionCtx)
	case "fill":
		return e.ExecuteFill(ctx, step, executionCtx)
	case "wait":
		return e.ExecuteWait(ctx, step, executionCtx)
	case "assert":
		// 断言在 BaseExecutor 的 executeSteps 中处理
		return nil
	case "screenshot":
		return fmt.Errorf("screenshot step not applicable for performance tests")
	case "javascript":
		return fmt.Errorf("javascript step not applicable for performance tests")
	default:
		e.logger.Warn("Unknown step type for performance executor", "type", step.Type)
		return nil
	}
}