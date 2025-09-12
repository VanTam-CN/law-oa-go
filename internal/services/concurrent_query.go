package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"law-oa-go/internal/models"

	"gorm.io/gorm"
	"log/slog"
)

// QueryResult 查询结果
type QueryResult[T any] struct {
	Data T
	Err  error
}

// QueryExecutor 并发查询执行器
type QueryExecutor struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewQueryExecutor(db *gorm.DB, logger *slog.Logger) *QueryExecutor {
	return &QueryExecutor{
		db:     db,
		logger: logger,
	}
}

// ExecuteConcurrentQueries 并发执行多个查询
func (qe *QueryExecutor) ExecuteConcurrentQueries(ctx context.Context, queries ...func(context.Context) QueryResult[any]) ([]any, error) {
	var wg sync.WaitGroup
	results := make([]any, len(queries))
	errChan := make(chan error, len(queries))
	done := make(chan struct{})

	// 创建带超时的context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for i, query := range queries {
		wg.Add(1)
		go func(idx int, q func(context.Context) QueryResult[any]) {
			defer wg.Done()
			
			result := q(ctx)
			if result.Err != nil {
				select {
				case errChan <- result.Err:
				case <-done:
				}
				return
			}
			results[idx] = result.Data
		}(i, query)
	}

	// 等待所有查询完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		close(done)
		return nil, errors.Join(errs...)
	}

	return results, nil
}

// 优化后的案件统计查询
func (s *CaseService) GetCaseStatsOptimized(ctx context.Context) (*CaseStatsResponse, error) {
	start := time.Now()
	defer func() {
		// TODO: 记录数据库指标，避免循环依赖
		_ = time.Since(start)
	}()

	executor := NewQueryExecutor(s.db, nil) // 暂时传入nil logger

	queries := []func(context.Context) QueryResult[any]{
		func(ctx context.Context) QueryResult[any] {
			var count int64
			err := s.db.WithContext(ctx).Model(&models.Case{}).Count(&count).Error
			return QueryResult[any]{Data: count, Err: err}
		},
		func(ctx context.Context) QueryResult[any] {
			var count int64
			err := s.db.WithContext(ctx).Model(&models.Case{}).Where("status = ?", "active").Count(&count).Error
			return QueryResult[any]{Data: count, Err: err}
		},
		func(ctx context.Context) QueryResult[any] {
			var count int64
			err := s.db.WithContext(ctx).Model(&models.Case{}).Where("status = ?", "pending").Count(&count).Error
			return QueryResult[any]{Data: count, Err: err}
		},
		func(ctx context.Context) QueryResult[any] {
			var count int64
			err := s.db.WithContext(ctx).Model(&models.Case{}).Where("status = ?", "closed").Count(&count).Error
			return QueryResult[any]{Data: count, Err: err}
		},
		func(ctx context.Context) QueryResult[any] {
			var count int64
			err := s.db.WithContext(ctx).Model(&models.Case{}).Where("priority = ?", "urgent").Count(&count).Error
			return QueryResult[any]{Data: count, Err: err}
		},
	}

	results, err := executor.ExecuteConcurrentQueries(ctx, queries...)
	if err != nil {
		return nil, err
	}

	stats := &CaseStatsResponse{
		TotalCases:      results[0].(int64),
		ActiveCases:     results[1].(int64),
		PendingCases:    results[2].(int64),
		ClosedCases:     results[3].(int64),
		UrgentCases:     results[4].(int64),
	}

	return stats, nil
}