package services

import (
	"context"
	"fmt"

	"law-oa-go/internal/models"
	"gorm.io/gorm"
)

// BatchService 批量操作服务
type BatchService struct {
	db *gorm.DB
}

func NewBatchService(db *gorm.DB) *BatchService {
	return &BatchService{db: db}
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