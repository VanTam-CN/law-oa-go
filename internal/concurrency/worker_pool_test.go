package concurrency

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkerPool(t *testing.T) {
	// 测试创建Worker Pool
	pool := NewWorkerPool(5, 100, 10*time.Second)

	assert.NotNil(t, pool)
	assert.False(t, pool.IsRunning())
	assert.Equal(t, 0, pool.GetActiveTasksCount())

	// 验证默认配置
	assert.NotNil(t, pool.retryPolicy)
	assert.Equal(t, 3, pool.retryPolicy.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, pool.retryPolicy.RetryDelay)
	assert.Equal(t, 2.0, pool.retryPolicy.BackoffFactor)
}

func TestWorkerPool_StartStop(t *testing.T) {
	pool := NewWorkerPool(3, 50, 5*time.Second)

	// 测试启动
	pool.Start()
	assert.True(t, pool.IsRunning())

	// 测试重复启动
	pool.Start() // 不应该panic
	assert.True(t, pool.IsRunning())

	// 测试停止
	pool.Stop()
	assert.False(t, pool.IsRunning())

	// 测试重复停止
	pool.Stop() // 不应该panic
	assert.False(t, pool.IsRunning())
}

func TestWorkerPool_SubmitTask(t *testing.T) {
	pool := NewWorkerPool(2, 10, 5*time.Second)
	pool.Start()
	defer pool.Stop()

	// 创建一个简单的任务
	task := &DatabaseTask{
		TaskID:       "test_task",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
		Context: context.Background(),
	}

	// 提交任务
	err := pool.Submit(task)
	assert.NoError(t, err)

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 0, pool.GetActiveTasksCount())
}

func TestWorkerPool_SubmitWithResult(t *testing.T) {
	pool := NewWorkerPool(2, 10, 5*time.Second)
	pool.Start()
	defer pool.Stop()

	task := &DatabaseTask{
		TaskID:       "test_result_task",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
		Context: context.Background(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 提交任务并等待结果
	result, err := pool.SubmitWithResult(ctx, task)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, task.TaskID, result.TaskID)
	assert.NoError(t, result.Error)
}

func TestWorkerPool_ConcurrentTasks(t *testing.T) {
	pool := NewWorkerPool(5, 20, 5*time.Second)
	pool.Start()
	defer pool.Stop()

	const numTasks = 10
	var wg sync.WaitGroup
	var mu sync.Mutex
	completedTasks := make([]string, 0, numTasks)

	// 创建并发任务
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		taskID := fmt.Sprintf("concurrent_task_%d", i)

		task := &DatabaseTask{
			TaskID:       taskID,
			TaskType:     "concurrent_test",
			TaskPriority: 1,
			Operation: func(ctx context.Context) error {
				defer wg.Done()
				time.Sleep(50 * time.Millisecond)

				mu.Lock()
				completedTasks = append(completedTasks, taskID)
				mu.Unlock()

				return nil
			},
			Context: context.Background(),
		}

		err := pool.Submit(task)
		assert.NoError(t, err)
	}

	// 等待所有任务完成
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, numTasks, len(completedTasks))
	assert.Equal(t, 0, pool.GetActiveTasksCount())
}

func TestWorkerPool_TaskErrorHandling(t *testing.T) {
	pool := NewWorkerPool(2, 10, 5*time.Second)
	pool.Start()
	defer pool.Stop()

	// 创建失败的任务
	task := &DatabaseTask{
		TaskID:       "error_task",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			return fmt.Errorf("task failed intentionally")
		},
		Context: context.Background(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 提交失败任务
	result, err := pool.SubmitWithResult(ctx, task)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, task.TaskID, result.TaskID)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "task failed intentionally")
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	pool := NewWorkerPool(2, 10, 1*time.Second)
	pool.Start()
	defer pool.Stop()

	// 创建长时间运行的任务
	task := &DatabaseTask{
		TaskID:       "long_task",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			select {
			case <-time.After(3 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		Context: context.Background(),
	}

	// 使用短超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 提交任务
	_, err := pool.SubmitWithResult(ctx, task)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestWorkerPool_RetryMechanism(t *testing.T) {
	pool := NewWorkerPool(2, 10, 5*time.Second)

	// 设置重试策略
	pool.retryPolicy = &RetryPolicy{
		MaxRetries:    2,
		RetryDelay:    10 * time.Millisecond,
		BackoffFactor: 1.5,
	}

	pool.Start()
	defer pool.Stop()

	attemptCount := 0

	task := &DatabaseTask{
		TaskID:       "retry_task",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			attemptCount++
			if attemptCount < 3 { // 前两次失败
				return fmt.Errorf("temporary failure")
			}
			return nil // 第三次成功
		},
		Context: context.Background(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 提交需要重试的任务
	result, err := pool.SubmitWithResult(ctx, task)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NoError(t, result.Error)
	assert.Equal(t, 3, attemptCount) // 应该重试3次
}

func TestWorkerPool_Metrics(t *testing.T) {
	pool := NewWorkerPool(3, 15, 5*time.Second)
	pool.Start()
	defer pool.Stop()

	// 获取初始指标
	metrics := pool.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalTasks)
	assert.Equal(t, int64(0), metrics.SuccessTasks)
	assert.Equal(t, int64(0), metrics.FailedTasks)

	// 提交一些任务
	for i := 0; i < 5; i++ {
		task := &DatabaseTask{
			TaskID:       fmt.Sprintf("metrics_task_%d", i),
			TaskType:     "test",
			TaskPriority: 1,
			Operation: func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			},
			Context: context.Background(),
		}
		pool.Submit(task)
	}

	// 等待任务完成
	time.Sleep(200 * time.Millisecond)

	// 检查更新后的指标
	metrics = pool.GetMetrics()
	assert.Equal(t, int64(5), metrics.TotalTasks)
	assert.Equal(t, int64(5), metrics.SuccessTasks)
	assert.Equal(t, int64(0), metrics.FailedTasks)
}

func TestWorkerPool_FullQueue(t *testing.T) {
	pool := NewWorkerPool(1, 2, 5*time.Second) // 小队列
	pool.Start()
	defer pool.Stop()

	// 填满队列
	task1 := &DatabaseTask{
		TaskID:       "queue_task_1",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			time.Sleep(200 * time.Millisecond)
			return nil
		},
		Context: context.Background(),
	}

	task2 := &DatabaseTask{
		TaskID:       "queue_task_2",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			time.Sleep(200 * time.Millisecond)
			return nil
		},
		Context: context.Background(),
	}

	// 提交前两个任务应该成功
	assert.NoError(t, pool.Submit(task1))
	assert.NoError(t, pool.Submit(task2))

	// 第三个任务应该失败（队列已满）
	task3 := &DatabaseTask{
		TaskID:       "queue_task_3",
		TaskType:     "test",
		TaskPriority: 1,
		Operation: func(ctx context.Context) error {
			return nil
		},
		Context: context.Background(),
	}

	err := pool.Submit(task3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}
