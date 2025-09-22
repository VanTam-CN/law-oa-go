package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/concurrency"
	"law-oa-go/internal/models"
)

// BatchService 批量操作服务
type BatchService struct {
	db             *gorm.DB
	concurrentSvc  *concurrency.ConcurrentService
	concurrentSafe *concurrency.ConcurrentSafe
	mu             sync.RWMutex
}

func NewBatchService(db *gorm.DB) *BatchService {
	config := &concurrency.ConcurrentConfig{
		MaxWorkers:     50, // 批量操作需要更多工作线程
		QueueSize:      5000,
		TaskTimeout:    60 * time.Second, // 批量操作可能需要更长时间
		EnableMetrics:  true,
		CircuitBreaker: true,
		RateLimiter:    true,
		RetryPolicy: concurrency.RetryPolicy{
			MaxRetries:    5, // 批量操作需要更多重试次数
			RetryDelay:    200 * time.Millisecond,
			BackoffFactor: 2.0,
		},
	}

	concurrentSvc := concurrency.NewConcurrentService(config)
	concurrentSafe := concurrency.NewConcurrentSafe(true, true, 200, 100) // 批量操作允许更高的速率

	return &BatchService{
		db:             db,
		concurrentSvc:  concurrentSvc,
		concurrentSafe: concurrentSafe,
	}
}

// BatchCreateClients 批量创建客户
func (s *BatchService) BatchCreateClients(ctx context.Context, clients []*models.Client, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100 // 默认批次大小
	}

	return s.db.WithContext(ctx).CreateInBatches(clients, batchSize).Error
}

// BatchUpdateClients 批量更新客户
func (s *BatchService) BatchUpdateClients(ctx context.Context, updates []map[string]interface{}, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	// 使用事务确保数据一致性
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(updates); i += batchSize {
			end := i + batchSize
			if end > len(updates) {
				end = len(updates)
			}

			batch := updates[i:end]
			for _, update := range batch {
				if id, ok := update["id"]; ok {
					delete(update, "id") // 移除ID字段
					if err := tx.Model(&models.Client{}).Where("id = ?", id).Updates(update).Error; err != nil {
						return fmt.Errorf("failed to update client %v: %w", id, err)
					}
				}
			}
		}
		return nil
	})
}

// BatchCreateCases 批量创建案件
func (s *BatchService) BatchCreateCases(ctx context.Context, cases []*models.Case, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	return s.db.WithContext(ctx).CreateInBatches(cases, batchSize).Error
}

// BatchDeleteByIDs 批量删除记录
func (s *BatchService) BatchDeleteByIDs(ctx context.Context, model interface{}, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Where("id IN ?", ids).Delete(model).Error
}

// BatchFindByIDs 批量查询记录
func (s *BatchService) BatchFindByIDs(ctx context.Context, model interface{}, ids []uint, result interface{}) error {
	if len(ids) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Where("id IN ?", ids).Find(result).Error
}

// BatchProcessWithCallback 批量处理数据（带回调）
func (s *BatchService) BatchProcessWithCallback(ctx context.Context,
	query *gorm.DB,
	batchSize int,
	callback func(batch interface{}) error) error {
	if batchSize <= 0 {
		batchSize = 1000
	}

	return query.WithContext(ctx).FindInBatches(&[]models.Client{}, batchSize, func(tx *gorm.DB, batch int) error {
		var records []models.Client
		if err := tx.Find(&records).Error; err != nil {
			return err
		}
		return callback(records)
	}).Error
}

// StartConcurrentService 启动并发服务
func (s *BatchService) StartConcurrentService() {
	s.concurrentSvc.Start()
}

// StopConcurrentService 停止并发服务
func (s *BatchService) StopConcurrentService() {
	s.concurrentSvc.Stop()
}

// GetConcurrentMetrics 获取并发指标
func (s *BatchService) GetConcurrentMetrics() *concurrency.PoolMetricsSnapshot {
	return s.concurrentSvc.GetMetrics()
}

// BatchCreateClientsConcurrent 并发批量创建客户
func (s *BatchService) BatchCreateClientsConcurrent(ctx context.Context, clients []*models.Client, maxConcurrency int) error {
	if len(clients) == 0 {
		return fmt.Errorf("no clients to create")
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 10 // 默认并发数
	}

	// 分批处理
	batchSize := len(clients) / maxConcurrency
	if batchSize < 1 {
		batchSize = 1
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(clients))
	semaphore := make(chan struct{}, maxConcurrency) // 信号量控制并发数

	for i := 0; i < len(clients); i += batchSize {
		end := i + batchSize
		if end > len(clients) {
			end = len(clients)
		}

		batch := clients[i:end]

		wg.Add(1)
		go func(batch []*models.Client) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			task := &concurrency.DatabaseTask{
				TaskID:       fmt.Sprintf("create_clients_batch_%d", i/batchSize),
				TaskType:     "create_clients",
				TaskPriority: 3,
				Operation: func(ctx context.Context) error {
					return s.concurrentSafe.Execute(ctx, func() error {
						return s.db.WithContext(ctx).CreateInBatches(batch, len(batch)).Error
					})
				},
				Context: ctx,
			}

			if err := task.Execute(ctx); err != nil {
				errChan <- fmt.Errorf("failed to create batch %d: %w", i/batchSize, err)
			}
		}(batch)
	}

	wg.Wait()
	close(errChan)

	// 检查错误
	for err := range errChan {
		if err != nil {
			return fmt.Errorf("batch create clients failed: %w", err)
		}
	}

	return nil
}

// BatchUpdateClientsConcurrent 并发批量更新客户
func (s *BatchService) BatchUpdateClientsConcurrent(ctx context.Context, updates []map[string]interface{}, maxConcurrency int) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates to process")
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	batchSize := len(updates) / maxConcurrency
	if batchSize < 1 {
		batchSize = 1
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(updates))
	semaphore := make(chan struct{}, maxConcurrency)

	for i := 0; i < len(updates); i += batchSize {
		end := i + batchSize
		if end > len(updates) {
			end = len(updates)
		}

		batch := updates[i:end]

		wg.Add(1)
		go func(batch []map[string]interface{}, batchIndex int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			task := &concurrency.DatabaseTask{
				TaskID:       fmt.Sprintf("update_clients_batch_%d", batchIndex),
				TaskType:     "update_clients",
				TaskPriority: 2,
				Operation: func(ctx context.Context) error {
					return s.concurrentSafe.Execute(ctx, func() error {
						return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
							for _, update := range batch {
								if id, ok := update["id"]; ok {
									delete(update, "id")
									if err := tx.Model(&models.Client{}).Where("id = ?", id).Updates(update).Error; err != nil {
										return fmt.Errorf("failed to update client %v: %w", id, err)
									}
								}
							}
							return nil
						})
					})
				},
				Context: ctx,
			}

			if err := task.Execute(ctx); err != nil {
				errChan <- fmt.Errorf("failed to update batch %d: %w", batchIndex, err)
			}
		}(batch, i/batchSize)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("batch update clients failed: %w", err)
		}
	}

	return nil
}

// BatchDeleteByIDsConcurrent 并发批量删除记录
func (s *BatchService) BatchDeleteByIDsConcurrent(ctx context.Context, model interface{}, ids []uint, maxConcurrency int) error {
	if len(ids) == 0 {
		return fmt.Errorf("no IDs to delete")
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	batchSize := len(ids) / maxConcurrency
	if batchSize < 1 {
		batchSize = 1
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(ids))
	semaphore := make(chan struct{}, maxConcurrency)

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batch := ids[i:end]

		wg.Add(1)
		go func(batch []uint, batchIndex int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			task := &concurrency.DatabaseTask{
				TaskID:       fmt.Sprintf("delete_records_batch_%d", batchIndex),
				TaskType:     "delete_records",
				TaskPriority: 1,
				Operation: func(ctx context.Context) error {
					return s.concurrentSafe.Execute(ctx, func() error {
						return s.db.WithContext(ctx).Where("id IN ?", batch).Delete(model).Error
					})
				},
				Context: ctx,
			}

			if err := task.Execute(ctx); err != nil {
				errChan <- fmt.Errorf("failed to delete batch %d: %w", batchIndex, err)
			}
		}(batch, i/batchSize)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return fmt.Errorf("batch delete records failed: %w", err)
		}
	}

	return nil
}

// BatchProcessWithCallbackConcurrent 并发批量处理数据（带回调）
func (s *BatchService) BatchProcessWithCallbackConcurrent(ctx context.Context,
	query *gorm.DB,
	batchSize int,
	maxConcurrency int,
	callback func(batch interface{}) error) error {
	if batchSize <= 0 {
		batchSize = 1000
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 5
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100) // 缓冲channel
	semaphore := make(chan struct{}, maxConcurrency)

	batchNumber := 0

	return query.WithContext(ctx).FindInBatches(&[]models.Client{}, batchSize, func(tx *gorm.DB, batch int) error {
		var records []models.Client
		if err := tx.Find(&records).Error; err != nil {
			return err
		}

		wg.Add(1)
		go func(batchData []models.Client, batchNum int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			task := &concurrency.DatabaseTask{
				TaskID:       fmt.Sprintf("process_batch_%d", batchNum),
				TaskType:     "process_batch",
				TaskPriority: 4,
				Operation: func(ctx context.Context) error {
					return s.concurrentSafe.Execute(ctx, func() error {
						return callback(batchData)
					})
				},
				Context: ctx,
			}

			if err := task.Execute(ctx); err != nil {
				errChan <- fmt.Errorf("failed to process batch %d: %w", batchNum, err)
			}
		}(records, batchNumber)

		batchNumber++
		return nil
	}).Error
}
