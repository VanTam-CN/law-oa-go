package testing

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TestScheduler 测试调度器
type TestScheduler struct {
	executorManager *ExecutorManager
	queue           *TestQueue
	runningTests    map[string]*RunningTest
	completedTests  []*TestExecutionResult
	logger          TestLogger
	metrics         TestMetrics
	config          *SchedulerConfig
	mu              sync.RWMutex
	stopCh          chan struct{}
	wg              sync.WaitGroup
	running         bool
}

// TestQueue 测试队列
type TestQueue struct {
	items    []*QueueItem
	priority bool
	mu       sync.Mutex
	notEmpty chan struct{}
}

// QueueItem 队列项
type QueueItem struct {
	Test       *TestCase
	ExecutionCtx *ExecutionContext
	Priority   int
	EnqueuedAt time.Time
	ScheduledAt time.Time
}

// RunningTest 运行中的测试
type RunningTest struct {
	ID         string
	Test       *TestCase
	ExecutionCtx *ExecutionContext
	Executor   TestExecutor
	StartTime  time.Time
	Status     TestStatus
	Progress   float64
	Cancel     context.CancelFunc
	Logger     TestLogger
	Mutex      sync.RWMutex
}

// TestExecutionResult 测试执行结果
type TestExecutionResult struct {
	TestID      string
	ExecutionID string
	StartTime   time.Time
	EndTime     time.Time
	Status      TestStatus
	Progress    float64
	Result      *TestResult
	Error       error
	Metadata    map[string]interface{}
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	MaxConcurrentTests int           `json:"max_concurrent_tests"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	QueueSize          int           `json:"queue_size"`
	PriorityQueue      bool          `json:"priority_queue"`
	EnableRetry        bool          `json:"enable_retry"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
	CleanupInterval    time.Duration `json:"cleanup_interval"`
	MetricsInterval    time.Duration `json:"metrics_interval"`
	EnableLogging      bool          `json:"enable_logging"`
	LogLevel           string        `json:"log_level"`
}

// NewTestScheduler 创建测试调度器
func NewTestScheduler(executorManager *ExecutorManager, logger TestLogger, metrics TestMetrics, config *SchedulerConfig) *TestScheduler {
	if config == nil {
		config = NewDefaultSchedulerConfig()
	}

	scheduler := &TestScheduler{
		executorManager: executorManager,
		queue:           NewTestQueue(config.PriorityQueue),
		runningTests:    make(map[string]*RunningTest),
		completedTests:  make([]*TestExecutionResult, 0),
		logger:          logger,
		metrics:         metrics,
		config:          config,
		stopCh:          make(chan struct{}),
	}

	return scheduler
}

// NewDefaultSchedulerConfig 创建默认调度器配置
func NewDefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		MaxConcurrentTests: 10,
		DefaultTimeout:     30 * time.Minute,
		QueueSize:          1000,
		PriorityQueue:      true,
		EnableRetry:        true,
		MaxRetries:         3,
		RetryDelay:         1 * time.Minute,
		CleanupInterval:    5 * time.Minute,
		MetricsInterval:    1 * time.Minute,
		EnableLogging:      true,
		LogLevel:           "info",
	}
}

// Start 启动调度器
func (s *TestScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	s.running = true

	// 启动主调度循环
	s.wg.Add(1)
	go s.scheduleLoop()

	// 启动清理任务
	s.wg.Add(1)
	go s.cleanupLoop()

	// 启动指标收集
	s.wg.Add(1)
	go s.metricsLoop()

	s.logger.Info("Test scheduler started", "config", s.config)
	return nil
}

// Stop 停止调度器
func (s *TestScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	close(s.stopCh)

	// 取消所有运行中的测试
	for _, runningTest := range s.runningTests {
		if runningTest.Cancel != nil {
			runningTest.Cancel()
		}
	}

	// 等待所有协程结束
	s.wg.Wait()

	s.logger.Info("Test scheduler stopped")
	return nil
}

// SubmitTest 提交测试
func (s *TestScheduler) SubmitTest(test *TestCase, executionCtx *ExecutionContext) error {
	if executionCtx == nil {
		executionCtx = &ExecutionContext{
			ExecutionID: generateExecutionID(),
			Timeout:     s.config.DefaultTimeout,
		}
	}

	if executionCtx.Timeout == 0 {
		executionCtx.Timeout = s.config.DefaultTimeout
	}

	item := &QueueItem{
		Test:        test,
		ExecutionCtx: executionCtx,
		Priority:    0, // 默认优先级
		EnqueuedAt:  time.Now(),
	}

	if err := s.queue.Enqueue(item); err != nil {
		return fmt.Errorf("failed to enqueue test: %w", err)
	}

	s.logger.Info("Test submitted", "test_id", test.ID, "execution_id", executionCtx.ExecutionID)
	return nil
}

// SubmitTestWithPriority 提交带优先级的测试
func (s *TestScheduler) SubmitTestWithPriority(test *TestCase, executionCtx *ExecutionContext, priority int) error {
	if executionCtx == nil {
		executionCtx = &ExecutionContext{
			ExecutionID: generateExecutionID(),
			Timeout:     s.config.DefaultTimeout,
		}
	}

	item := &QueueItem{
		Test:        test,
		ExecutionCtx: executionCtx,
		Priority:    priority,
		EnqueuedAt:  time.Now(),
	}

	if err := s.queue.Enqueue(item); err != nil {
		return fmt.Errorf("failed to enqueue test with priority: %w", err)
	}

	s.logger.Info("Test submitted with priority", "test_id", test.ID, "priority", priority)
	return nil
}

// SubmitBatch 提交批量测试
func (s *TestScheduler) SubmitBatch(tests []*TestCase, executionCtx *ExecutionContext) error {
	for _, test := range tests {
		if err := s.SubmitTest(test, executionCtx); err != nil {
			return fmt.Errorf("failed to submit test %s: %w", test.ID, err)
		}
	}

	s.logger.Info("Batch tests submitted", "count", len(tests))
	return nil
}

// GetStatus 获取调度器状态
func (s *TestScheduler) GetStatus() *SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &SchedulerStatus{
		Running:           s.running,
		QueueSize:         s.queue.Size(),
		RunningTestsCount: len(s.runningTests),
		CompletedTestsCount: len(s.completedTests),
		MaxConcurrent:     s.config.MaxConcurrentTests,
	}

	// 计算队列等待时间
	if oldestItem := s.queue.GetOldest(); oldestItem != nil {
		status.OldestQueuedTime = time.Since(oldestItem.EnqueuedAt)
	}

	// 收集运行中测试的状态
	status.RunningTests = make([]*TestStatusInfo, 0, len(s.runningTests))
	for _, runningTest := range s.runningTests {
		runningTest.Mutex.RLock()
		status.RunningTests = append(status.RunningTests, &TestStatusInfo{
			TestID:       runningTest.Test.ID,
			ExecutionID:  runningTest.ExecutionCtx.ExecutionID,
			Status:       runningTest.Status,
			Progress:     runningTest.Progress,
			StartTime:    runningTest.StartTime,
			Duration:     time.Since(runningTest.StartTime),
		})
		runningTest.Mutex.RUnlock()
	}

	return status
}

// GetTestResult 获取测试结果
func (s *TestScheduler) GetTestResult(executionID string) (*TestExecutionResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 先检查运行中的测试
	if runningTest, exists := s.runningTests[executionID]; exists {
		runningTest.Mutex.RLock()
		defer runningTest.Mutex.RUnlock()

		return &TestExecutionResult{
			TestID:      runningTest.Test.ID,
			ExecutionID: runningTest.ExecutionCtx.ExecutionID,
			StartTime:   runningTest.StartTime,
			Status:      runningTest.Status,
			Progress:    runningTest.Progress,
		}, true
	}

	// 检查已完成的测试
	for _, result := range s.completedTests {
		if result.ExecutionID == executionID {
			return result, true
		}
	}

	return nil, false
}

// CancelTest 取消测试
func (s *TestScheduler) CancelTest(executionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查运行中的测试
	runningTest, exists := s.runningTests[executionID]
	if !exists {
		return fmt.Errorf("test execution not found: %s", executionID)
	}

	// 取消测试
	if runningTest.Cancel != nil {
		runningTest.Cancel()
	}

	s.logger.Info("Test cancelled", "execution_id", executionID)
	return nil
}

// scheduleLoop 主调度循环
func (s *TestScheduler) scheduleLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond) // 100ms 调度间隔
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.processQueue()
		}
	}
}

// processQueue 处理队列
func (s *TestScheduler) processQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否还有空闲槽位
	if len(s.runningTests) >= s.config.MaxConcurrentTests {
		return
	}

	// 从队列中获取测试
	for len(s.runningTests) < s.config.MaxConcurrentTests {
		item := s.queue.Dequeue()
		if item == nil {
			break
		}

		// 启动测试
		if err := s.startTest(item); err != nil {
			s.logger.Error("Failed to start test", "test_id", item.Test.ID, "error", err)

			// 记录失败结果
			result := &TestExecutionResult{
				TestID:      item.Test.ID,
				ExecutionID: item.ExecutionCtx.ExecutionID,
				StartTime:   time.Now(),
				EndTime:     time.Now(),
				Status:      TestStatusError,
				Error:       err,
			}
			s.completedTests = append(s.completedTests, result)
		}
	}
}

// startTest 启动测试
func (s *TestScheduler) startTest(item *QueueItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), item.ExecutionCtx.Timeout)

	// 获取执行器
	executor, err := s.executorManager.GetExecutor(item.Test.Type)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to get executor: %w", err)
	}

	// 设置执行器
	if err := executor.Setup(ctx, item.ExecutionCtx); err != nil {
		cancel()
		return fmt.Errorf("failed to setup executor: %w", err)
	}

	// 创建运行中测试
	runningTest := &RunningTest{
		ID:          item.ExecutionCtx.ExecutionID,
		Test:        item.Test,
		ExecutionCtx: item.ExecutionCtx,
		Executor:    executor,
		StartTime:   time.Now(),
		Status:      TestStatusRunning,
		Cancel:      cancel,
		Logger:      s.logger,
	}

	s.runningTests[item.ExecutionCtx.ExecutionID] = runningTest

	// 启动测试执行
	s.wg.Add(1)
	go s.executeTest(ctx, runningTest)

	s.logger.Info("Test started", "test_id", item.Test.ID, "execution_id", item.ExecutionCtx.ExecutionID)
	return nil
}

// executeTest 执行测试
func (s *TestScheduler) executeTest(ctx context.Context, runningTest *RunningTest) {
	defer s.wg.Done()

	// 更新状态
	runningTest.Mutex.Lock()
	runningTest.Status = TestStatusRunning
	runningTest.Mutex.Unlock()

	// 执行测试
	result, err := runningTest.Executor.Execute(ctx, runningTest.Test, runningTest.ExecutionCtx)

	// 清理执行器
	if err := runningTest.Executor.Teardown(ctx, runningTest.ExecutionCtx); err != nil {
		s.logger.Warn("Failed to teardown executor", "error", err)
	}

	// 处理结果
	executionResult := &TestExecutionResult{
		TestID:      runningTest.Test.ID,
		ExecutionID: runningTest.ExecutionCtx.ExecutionID,
		StartTime:   runningTest.StartTime,
		EndTime:     time.Now(),
		Status:      TestStatusFailed,
		Result:      result,
		Error:       err,
	}

	if err != nil {
		// 测试失败，检查是否需要重试
		if s.config.EnableRetry && s.shouldRetry(runningTest) {
			s.scheduleRetry(runningTest)
			return
		}
		executionResult.Status = TestStatusFailed
	} else if result != nil && result.Passed {
		executionResult.Status = TestStatusPassed
	} else {
		executionResult.Status = TestStatusFailed
	}

	// 记录结果
	s.mu.Lock()
	delete(s.runningTests, runningTest.ID)
	s.completedTests = append(s.completedTests, executionResult)
	s.mu.Unlock()

	// 记录指标
	s.metrics.RecordExecution(runningTest.Test.Type, executionResult.EndTime.Sub(executionResult.StartTime), executionResult.Status)

	s.logger.Info("Test completed", "test_id", runningTest.Test.ID, "execution_id", runningTest.ExecutionCtx.ExecutionID, "status", executionResult.Status)
}

// shouldRetry 检查是否应该重试
func (s *TestScheduler) shouldRetry(runningTest *RunningTest) bool {
	// 获取重试次数
	retryCount := 0
	if retryCountValue, exists := runningTest.ExecutionCtx.Variables["retry_count"]; exists {
		if count, ok := retryCountValue.(int); ok {
			retryCount = count
		}
	}

	return retryCount < s.config.MaxRetries
}

// scheduleRetry 调度重试
func (s *TestScheduler) scheduleRetry(runningTest *RunningTest) {
	// 增加重试次数
	retryCount := 0
	if retryCountValue, exists := runningTest.ExecutionCtx.Variables["retry_count"]; exists {
		if count, ok := retryCountValue.(int); ok {
			retryCount = count
		}
	}
	retryCount++

	// 更新执行上下文
	if runningTest.ExecutionCtx.Variables == nil {
		runningTest.ExecutionCtx.Variables = make(map[string]interface{})
	}
	runningTest.ExecutionCtx.Variables["retry_count"] = retryCount

	// 延迟重试
	go func() {
		time.Sleep(s.config.RetryDelay)
		if err := s.SubmitTest(runningTest.Test, runningTest.ExecutionCtx); err != nil {
			s.logger.Error("Failed to schedule retry", "test_id", runningTest.Test.ID, "error", err)
		}
	}()

	s.logger.Info("Test scheduled for retry", "test_id", runningTest.Test.ID, "retry_count", retryCount)
}

// cleanupLoop 清理循环
func (s *TestScheduler) cleanupLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

// cleanup 清理过期数据
func (s *TestScheduler) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 限制完成测试的历史记录数量
	maxCompletedTests := 1000
	if len(s.completedTests) > maxCompletedTests {
		// 保留最新的结果
		s.completedTests = s.completedTests[len(s.completedTests)-maxCompletedTests:]
	}

	s.logger.Debug("Cleanup completed", "completed_tests_count", len(s.completedTests))
}

// metricsLoop 指标收集循环
func (s *TestScheduler) metricsLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

// collectMetrics 收集指标
func (s *TestScheduler) collectMetrics() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 记录队列大小
	s.metrics.RecordCustom("scheduler_queue_size", float64(s.queue.Size()), map[string]string{
		"scheduler": "main",
	})

	// 记录运行中测试数量
	s.metrics.RecordCustom("scheduler_running_tests", float64(len(s.runningTests)), map[string]string{
		"scheduler": "main",
	})

	// 记录已完成测试数量
	s.metrics.RecordCustom("scheduler_completed_tests", float64(len(s.completedTests)), map[string]string{
		"scheduler": "main",
	})
}

// SchedulerStatus 调度器状态
type SchedulerStatus struct {
	Running            bool              `json:"running"`
	QueueSize          int               `json:"queue_size"`
	RunningTestsCount  int               `json:"running_tests_count"`
	CompletedTestsCount int               `json:"completed_tests_count"`
	MaxConcurrent      int               `json:"max_concurrent"`
	OldestQueuedTime   time.Duration     `json:"oldest_queued_time"`
	RunningTests       []*TestStatusInfo `json:"running_tests"`
}

// TestStatusInfo 测试状态信息
type TestStatusInfo struct {
	TestID      string        `json:"test_id"`
	ExecutionID string        `json:"execution_id"`
	Status      TestStatus    `json:"status"`
	Progress    float64       `json:"progress"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration"`
}

// NewTestQueue 创建测试队列
func NewTestQueue(priority bool) *TestQueue {
	return &TestQueue{
		items:    make([]*QueueItem, 0),
		priority: priority,
		notEmpty: make(chan struct{}, 1),
	}
}

// Enqueue 入队
func (q *TestQueue) Enqueue(item *QueueItem) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 检查队列大小限制
	if len(q.items) >= 1000 { // 默认最大队列大小
		return fmt.Errorf("queue is full")
	}

	q.items = append(q.items, item)

	// 通知非空
	select {
	case q.notEmpty <- struct{}{}:
	default:
	}

	return nil
}

// Dequeue 出队
func (q *TestQueue) Dequeue() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	if q.priority {
		// 优先级队列：找到优先级最高的项目
		maxIndex := 0
		for i := 1; i < len(q.items); i++ {
			if q.items[i].Priority > q.items[maxIndex].Priority {
				maxIndex = i
			}
		}
		item := q.items[maxIndex]
		q.items = append(q.items[:maxIndex], q.items[maxIndex+1:]...)
		return item
	} else {
		// FIFO队列
		item := q.items[0]
		q.items = q.items[1:]
		return item
	}
}

// Size 获取队列大小
func (q *TestQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// GetOldest 获取最旧的队列项
func (q *TestQueue) GetOldest() *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	oldest := q.items[0]
	for _, item := range q.items {
		if item.EnqueuedAt.Before(oldest.EnqueuedAt) {
			oldest = item
		}
	}

	return oldest
}

// generateExecutionID 生成执行ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}