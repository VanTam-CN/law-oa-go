package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ConcurrentService 并发服务
type ConcurrentService struct {
	workerPool *WorkerPool
	config     *ConcurrentConfig
	mu         sync.RWMutex
}

// ConcurrentConfig 并发配置
type ConcurrentConfig struct {
	MaxWorkers     int           `json:"max_workers"`
	QueueSize      int           `json:"queue_size"`
	TaskTimeout    time.Duration `json:"task_timeout"`
	RetryPolicy    RetryPolicy   `json:"retry_policy"`
	EnableMetrics  bool          `json:"enable_metrics"`
	CircuitBreaker bool          `json:"circuit_breaker"`
	RateLimiter    bool          `json:"rate_limiter"`
}

// ConcurrentTask 并发任务接口
type ConcurrentTask interface {
	Task
	GetContext() context.Context
	GetData() interface{}
	SetResult(interface{})
	GetResult() interface{}
}

// BatchTask 批量任务
type BatchTask struct {
	TaskID       string
	TaskType     string
	TaskPriority int
	Tasks        []Task
	Context      context.Context
	results      []interface{}
	mu           sync.Mutex
}

func (t *BatchTask) Execute(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(t.Tasks))

	for _, task := range t.Tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			if err := task.Execute(ctx); err != nil {
				errChan <- err
			}
		}(task)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return fmt.Errorf("batch task failed: %w", err)
		}
	}

	return nil
}

func (t *BatchTask) ID() string {
	return t.TaskID
}

func (t *BatchTask) Type() string {
	return t.TaskType
}

func (t *BatchTask) Priority() int {
	return t.TaskPriority
}

// DatabaseTask 数据库任务
type DatabaseTask struct {
	TaskID       string
	TaskType     string
	TaskPriority int
	Operation    func(ctx context.Context) error
	Context      context.Context
	Data         interface{}
	Result       interface{}
}

func (t *DatabaseTask) Execute(ctx context.Context) error {
	if t.Operation == nil {
		return fmt.Errorf("database operation is nil")
	}
	return t.Operation(ctx)
}

func (t *DatabaseTask) ID() string {
	return t.TaskID
}

func (t *DatabaseTask) Type() string {
	return t.TaskType
}

func (t *DatabaseTask) Priority() int {
	return t.TaskPriority
}

// FileTask 文件处理任务
type FileTask struct {
	TaskID       string
	TaskType     string
	TaskPriority int
	FilePath     string
	Process      func(ctx context.Context, filePath string) error
	Context      context.Context
	Data         interface{}
	Result       interface{}
}

func (t *FileTask) Execute(ctx context.Context) error {
	if t.Process == nil {
		return fmt.Errorf("file process function is nil")
	}
	return t.Process(ctx, t.FilePath)
}

func (t *FileTask) ID() string {
	return t.TaskID
}

func (t *FileTask) Type() string {
	return t.TaskType
}

func (t *FileTask) Priority() int {
	return t.TaskPriority
}

// APIRequestTask API请求任务
type APIRequestTask struct {
	TaskID       string
	TaskType     string
	TaskPriority int
	URL          string
	Method       string
	Headers      map[string]string
	Body         interface{}
	Request      func(ctx context.Context, url, method string, headers map[string]string, body interface{}) (interface{}, error)
	Context      context.Context
	Data         interface{}
	Result       interface{}
	mu           sync.Mutex
}

func (t *APIRequestTask) Execute(ctx context.Context) error {
	if t.Request == nil {
		return fmt.Errorf("api request function is nil")
	}

	result, err := t.Request(ctx, t.URL, t.Method, t.Headers, t.Body)
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.Result = result
	t.mu.Unlock()
	return nil
}

func (t *APIRequestTask) ID() string {
	return t.TaskID
}

func (t *APIRequestTask) Type() string {
	return t.TaskType
}

func (t *APIRequestTask) Priority() int {
	return t.TaskPriority
}

// NewConcurrentService 创建并发服务
func NewConcurrentService(config *ConcurrentConfig) *ConcurrentService {
	if config == nil {
		config = &ConcurrentConfig{
			MaxWorkers:  10,
			QueueSize:   1000,
			TaskTimeout: 30 * time.Second,
			RetryPolicy: RetryPolicy{
				MaxRetries:    3,
				RetryDelay:    100 * time.Millisecond,
				BackoffFactor: 2.0,
			},
			EnableMetrics:  true,
			CircuitBreaker: true,
			RateLimiter:    true,
		}
	}

	pool := NewWorkerPool(
		config.MaxWorkers,
		config.QueueSize,
		config.TaskTimeout,
	)
	pool.retryPolicy = &config.RetryPolicy

	return &ConcurrentService{
		workerPool: pool,
		config:     config,
	}
}

// Start 启动并发服务
func (s *ConcurrentService) Start() {
	s.workerPool.Start()
}

// Stop 停止并发服务
func (s *ConcurrentService) Stop() {
	s.workerPool.Stop()
}

// SubmitTask 提交任务
func (s *ConcurrentService) SubmitTask(task Task) error {
	return s.workerPool.Submit(task)
}

// SubmitTaskAndWait 提交任务并等待结果
func (s *ConcurrentService) SubmitTaskAndWait(ctx context.Context, task Task) (*TaskResult, error) {
	return s.workerPool.SubmitWithResult(ctx, task)
}

// SubmitBatchTasks 提交批量任务
func (s *ConcurrentService) SubmitBatchTasks(tasks []Task) (*TaskResult, error) {
	batchTask := &BatchTask{
		TaskID:       fmt.Sprintf("batch_%d", time.Now().UnixNano()),
		TaskType:     "batch",
		TaskPriority: 5,
		Tasks:        tasks,
		Context:      context.Background(),
	}

	// 创建等待结果的context
	ctx, cancel := context.WithTimeout(context.Background(), s.config.TaskTimeout)
	defer cancel()

	return s.SubmitTaskAndWait(ctx, batchTask)
}

// SubmitDatabaseTask 提交数据库任务
func (s *ConcurrentService) SubmitDatabaseTask(operation func(ctx context.Context) error) (*TaskResult, error) {
	task := &DatabaseTask{
		TaskID:       fmt.Sprintf("db_%d", time.Now().UnixNano()),
		TaskType:     "database",
		TaskPriority: 3,
		Operation:    operation,
		Context:      context.Background(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.TaskTimeout)
	defer cancel()

	return s.SubmitTaskAndWait(ctx, task)
}

// SubmitFileTask 提交文件处理任务
func (s *ConcurrentService) SubmitFileTask(filePath string, process func(ctx context.Context, filePath string) error) (*TaskResult, error) {
	task := &FileTask{
		TaskID:       fmt.Sprintf("file_%d", time.Now().UnixNano()),
		TaskType:     "file",
		TaskPriority: 4,
		FilePath:     filePath,
		Process:      process,
		Context:      context.Background(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.TaskTimeout)
	defer cancel()

	return s.SubmitTaskAndWait(ctx, task)
}

// SubmitAPITask 提交API请求任务
func (s *ConcurrentService) SubmitAPITask(url, method string, headers map[string]string, body interface{},
	request func(ctx context.Context, url, method string, headers map[string]string, body interface{}) (interface{}, error),
) (*TaskResult, error) {
	task := &APIRequestTask{
		TaskID:       fmt.Sprintf("api_%d", time.Now().UnixNano()),
		TaskType:     "api",
		TaskPriority: 2,
		URL:          url,
		Method:       method,
		Headers:      headers,
		Body:         body,
		Request:      request,
		Context:      context.Background(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.TaskTimeout)
	defer cancel()

	return s.SubmitTaskAndWait(ctx, task)
}

// GetMetrics 获取指标
func (s *ConcurrentService) GetMetrics() *PoolMetricsSnapshot {
	return s.workerPool.GetMetrics()
}

// IsRunning 检查服务是否在运行
func (s *ConcurrentService) IsRunning() bool {
	return s.workerPool.IsRunning()
}

// GetActiveTasksCount 获取活跃任务数量
func (s *ConcurrentService) GetActiveTasksCount() int {
	return s.workerPool.GetActiveTasksCount()
}

// UpdateConfig 更新配置
func (s *ConcurrentService) UpdateConfig(config *ConcurrentConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果服务正在运行，需要先停止
	if s.IsRunning() {
		s.Stop()
	}

	s.config = config

	// 重新创建worker pool
	s.workerPool = NewWorkerPool(
		config.MaxWorkers,
		config.QueueSize,
		config.TaskTimeout,
	)
	s.workerPool.retryPolicy = &config.RetryPolicy

	// 重新启动服务
	s.Start()
	return nil
}

// GetConfig 获取当前配置
func (s *ConcurrentService) GetConfig() *ConcurrentConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回配置的副本
	config := *s.config
	return &config
}
