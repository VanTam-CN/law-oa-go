package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task 表示一个可执行的任务
type Task interface {
	Execute(ctx context.Context) error
	ID() string
	Type() string
	Priority() int
}

// TaskResult 表示任务执行结果
type TaskResult struct {
	TaskID string
	Error  error
	Data   interface{}
}

// WorkerPool 工作池
type WorkerPool struct {
	workers       int
	taskQueue     chan Task
	resultQueue   chan TaskResult
	workerWg      sync.WaitGroup
	stopChan      chan struct{}
	isRunning     bool
	mu            sync.RWMutex
	activeTasks   map[string]context.CancelFunc
	taskTimeout   time.Duration
	retryPolicy   *RetryPolicy
	metrics       *PoolMetrics
	taskResults   sync.Map      // 存储任务结果
	resultTTL     time.Duration // 结果存活时间
	cleanupTicker *time.Ticker  // 清理定时器
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries    int
	RetryDelay    time.Duration
	BackoffFactor float64
}

// PoolMetrics 池指标
type PoolMetrics struct {
	TotalTasks         int64
	SuccessTasks       int64
	FailedTasks        int64
	RetriedTasks       int64
	ActiveWorkers      int
	QueueSize          int
	AverageProcessTime time.Duration
	mu                 sync.RWMutex
}

// PoolMetricsSnapshot 用于安全返回指标的快照
type PoolMetricsSnapshot struct {
	TotalTasks         int64
	SuccessTasks       int64
	FailedTasks        int64
	RetriedTasks       int64
	ActiveWorkers      int
	QueueSize          int
	AverageProcessTime time.Duration
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool(workers int, queueSize int, timeout time.Duration) *WorkerPool {
	if workers <= 0 {
		workers = 10 // 默认工作线程数
	}
	if queueSize <= 0 {
		queueSize = 1000 // 默认队列大小
	}

	pool := &WorkerPool{
		workers:     workers,
		taskQueue:   make(chan Task, queueSize),
		resultQueue: make(chan TaskResult, queueSize),
		stopChan:    make(chan struct{}),
		activeTasks: make(map[string]context.CancelFunc),
		taskTimeout: timeout,
		retryPolicy: &RetryPolicy{
			MaxRetries:    3,
			RetryDelay:    100 * time.Millisecond,
			BackoffFactor: 2.0,
		},
		metrics:   &PoolMetrics{},
		resultTTL: 5 * time.Minute, // 结果存活5分钟
	}

	return pool
}

// Start 启动工作池
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return
	}

	p.isRunning = true

	// 启动工作线程
	for i := 0; i < p.workers; i++ {
		p.workerWg.Add(1)
		go p.worker(i)
	}

	// 启动结果处理器
	go p.resultProcessor()

	// 启动指标收集器
	go p.metricsCollector()

	// 启动结果清理器
	p.cleanupTicker = time.NewTicker(1 * time.Minute)
	go p.resultCleaner()
}

// Stop 停止工作池
func (p *WorkerPool) Stop() {
	p.mu.Lock()

	if !p.isRunning {
		p.mu.Unlock()
		return
	}

	p.isRunning = false

	// 取消所有活跃任务
	for _, cancel := range p.activeTasks {
		cancel()
	}
	p.activeTasks = make(map[string]context.CancelFunc)

	p.mu.Unlock()

	// 停止清理定时器
	if p.cleanupTicker != nil {
		p.cleanupTicker.Stop()
	}

	// 发送停止信号并关闭队列
	close(p.stopChan)
	close(p.taskQueue)

	// 等待所有工作线程完成
	p.workerWg.Wait()

	// 关闭结果队列
	close(p.resultQueue)

	// 清理所有结果
	p.cleanupAllResults()
}

// Submit 提交任务到工作池
func (p *WorkerPool) Submit(task Task) error {
	p.mu.RLock()
	if !p.isRunning {
		p.mu.RUnlock()
		return fmt.Errorf("worker pool is not running")
	}
	p.mu.RUnlock()

	select {
	case p.taskQueue <- task:
		p.metrics.mu.Lock()
		p.metrics.TotalTasks++
		p.metrics.QueueSize = len(p.taskQueue)
		p.metrics.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// SubmitWithResult 提交任务并等待结果
func (p *WorkerPool) SubmitWithResult(ctx context.Context, task Task) (*TaskResult, error) {
	if err := p.Submit(task); err != nil {
		return nil, err
	}

	// 轮询等待任务结果
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			// 检查任务是否完成
			if result, ok := p.taskResults.Load(task.ID()); ok {
				taskResult := result.(TaskResult)
				p.taskResults.Delete(task.ID()) // 清理结果
				return &taskResult, nil
			}

			// 检查任务是否还在活跃
			p.mu.RLock()
			_, exists := p.activeTasks[task.ID()]
			p.mu.RUnlock()

			if !exists {
				// 任务已完成但没有结果存储，可能是被取消了
				return nil, fmt.Errorf("task completed but no result found")
			}
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for result")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// worker 工作线程
func (p *WorkerPool) worker(id int) {
	defer p.workerWg.Done()

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				return // 任务队列已关闭
			}

			p.processTask(task)

		case <-p.stopChan:
			return // 收到停止信号
		}
	}
}

// processTask 处理单个任务
func (p *WorkerPool) processTask(task Task) {
	startTime := time.Now()

	// 创建任务上下文
	ctx, cancel := context.WithTimeout(context.Background(), p.taskTimeout)
	defer cancel()

	// 注册活跃任务
	p.registerActiveTask(task.ID(), cancel)
	defer p.unregisterActiveTask(task.ID())

	// 执行任务（带重试机制）
	var result TaskResult
	var err error

	for attempt := 0; attempt <= p.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			p.metrics.mu.Lock()
			p.metrics.RetriedTasks++
			p.metrics.mu.Unlock()

			// 指数退避
			delay := p.retryPolicy.RetryDelay * time.Duration(float64(attempt)*p.retryPolicy.BackoffFactor)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return
			}
		}

		err = task.Execute(ctx)
		if err == nil {
			break
		}
	}

	result = TaskResult{
		TaskID: task.ID(),
		Error:  err,
		Data:   nil, // 可以在具体任务实现中设置
	}

	// 更新指标
	processingTime := time.Since(startTime)
	p.updateMetrics(result.Error == nil, processingTime)

	// 存储结果
	p.taskResults.Store(task.ID(), result)

	// 发送结果
	select {
	case p.resultQueue <- result:
	case <-ctx.Done():
	}
}

// resultProcessor 处理任务结果
func (p *WorkerPool) resultProcessor() {
	for result := range p.resultQueue {
		// 这里可以根据任务类型执行不同的后处理逻辑
		// 例如：记录日志、触发回调、更新缓存等
		p.handleResult(result)
	}
}

// handleResult 处理单个结果
func (p *WorkerPool) handleResult(result TaskResult) {
	if result.Error != nil {
		// 记录错误日志或触发错误处理流程
		fmt.Printf("Task %s failed: %v\n", result.TaskID, result.Error)
	} else {
		// 记录成功日志或触发成功回调
		fmt.Printf("Task %s completed successfully\n", result.TaskID)
	}
}

// registerActiveTask 注册活跃任务
func (p *WorkerPool) registerActiveTask(taskID string, cancel context.CancelFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.activeTasks[taskID] = cancel
}

// unregisterActiveTask 注销活跃任务
func (p *WorkerPool) unregisterActiveTask(taskID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.activeTasks, taskID)
}

// updateMetrics 更新指标
func (p *WorkerPool) updateMetrics(success bool, processingTime time.Duration) {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	if success {
		p.metrics.SuccessTasks++
	} else {
		p.metrics.FailedTasks++
	}

	// 更新平均处理时间
	totalTasks := p.metrics.SuccessTasks + p.metrics.FailedTasks
	if totalTasks > 0 {
		currentAvg := p.metrics.AverageProcessTime
		newAvg := time.Duration((int64(currentAvg)*(totalTasks-1) + int64(processingTime)) / totalTasks)
		p.metrics.AverageProcessTime = newAvg
	}
}

// metricsCollector 指标收集器
func (p *WorkerPool) metricsCollector() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.collectMetrics()
		case <-p.stopChan:
			return
		}
	}
}

// collectMetrics 收集指标
func (p *WorkerPool) collectMetrics() {
	p.metrics.mu.Lock()
	defer p.metrics.mu.Unlock()

	p.metrics.ActiveWorkers = p.workers
	p.metrics.QueueSize = len(p.taskQueue)

	// 可以在这里将指标发送到监控系统
	fmt.Printf("WorkerPool Metrics: Active=%d, Queue=%d, Success=%d, Failed=%d, AvgTime=%v\n",
		p.metrics.ActiveWorkers,
		p.metrics.QueueSize,
		p.metrics.SuccessTasks,
		p.metrics.FailedTasks,
		p.metrics.AverageProcessTime)
}

// GetMetrics 获取当前指标
func (p *WorkerPool) GetMetrics() *PoolMetricsSnapshot {
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// 返回安全的快照，避免锁拷贝
	return &PoolMetricsSnapshot{
		TotalTasks:         p.metrics.TotalTasks,
		SuccessTasks:       p.metrics.SuccessTasks,
		FailedTasks:        p.metrics.FailedTasks,
		RetriedTasks:       p.metrics.RetriedTasks,
		ActiveWorkers:      p.metrics.ActiveWorkers,
		QueueSize:          p.metrics.QueueSize,
		AverageProcessTime: p.metrics.AverageProcessTime,
	}
}

// IsRunning 检查工作池是否在运行
func (p *WorkerPool) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isRunning
}

// GetActiveTasksCount 获取活跃任务数量
func (p *WorkerPool) GetActiveTasksCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.activeTasks)
}

// resultCleaner 定期清理过期的任务结果
func (p *WorkerPool) resultCleaner() {
	for {
		select {
		case <-p.cleanupTicker.C:
			p.cleanupExpiredResults()
		case <-p.stopChan:
			return
		}
	}
}

// cleanupExpiredResults 清理过期的任务结果
func (p *WorkerPool) cleanupExpiredResults() {
	p.taskResults.Range(func(key, value interface{}) bool {
		if _, ok := value.(TaskResult); ok {
			// 简单的清理策略：定期清理一部分结果以避免内存泄漏
			// 在实际应用中，可以根据任务时间戳进行更精确的清理
			if time.Now().UnixNano()%100 == 0 { // 随机清理约1%的结果
				p.taskResults.Delete(key)
			}
		}
		return true
	})
}

// cleanupAllResults 清理所有任务结果
func (p *WorkerPool) cleanupAllResults() {
	p.taskResults.Range(func(key, value interface{}) bool {
		p.taskResults.Delete(key)
		return true
	})
}
