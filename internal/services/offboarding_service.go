package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// OffboardingService 离职交接服务接口
type OffboardingService interface {
	// 离职交接管理
	InitiateOffboarding(ctx context.Context, req *OffboardingRequest) (*models.OffboardingRecord, error)
	GetOffboardingByID(ctx context.Context, id uint) (*models.OffboardingRecord, error)
	GetOffboardingByUserID(ctx context.Context, userID uint) ([]*models.OffboardingRecord, error)
	UpdateOffboarding(ctx context.Context, record *models.OffboardingRecord) error
	CompleteOffboarding(ctx context.Context, id uint) error
	CancelOffboarding(ctx context.Context, id uint, reason string) error

	// 案件移交
	TransferCases(ctx context.Context, offboardingID uint, newLawyerID uint, caseIDs []uint) error
	TransferInboxItems(ctx context.Context, offboardingID uint, newAssistantID uint, inboxIDs []uint) error
	HandleDocuments(ctx context.Context, offboardingID uint, method string) error

	// 获取待交接数据
	GetUserCases(ctx context.Context, userID uint) ([]*models.Case, error)
	GetUserInboxItems(ctx context.Context, userID uint) ([]*models.InboxItem, error)
	GetOffboardingProgress(ctx context.Context, id uint) (*OffboardingProgress, error)

	// 批量操作
	BatchTransferCases(ctx context.Context, userID uint, newLawyerID uint) error
	BatchTransferInboxItems(ctx context.Context, userID uint, newAssistantID uint) error
}

// OffboardingRequest 离职交接请求
type OffboardingRequest struct {
	UserID         uint   `json:"user_id" binding:"required"`
	InitiatedBy    uint   `json:"initiated_by" binding:"required"`
	NewLawyerID    uint   `json:"new_lawyer_id"`
	NewAssistantID uint   `json:"new_assistant_id"`
	DocumentMethod string `json:"document_disposal_method" binding:"required"`
	Notes          string `json:"notes"`
}

// OffboardingProgress 离职交接进度
type OffboardingProgress struct {
	OffboardingID        uint   `json:"offboarding_id"`
	Status               string `json:"status"`
	CaseTransferCount    int    `json:"case_transfer_count"`
	CaseCompletedCount   int    `json:"case_completed_count"`
	InboxTransferCount   int    `json:"inbox_transfer_count"`
	InboxCompletedCount  int    `json:"inbox_completed_count"`
	DocumentHandled      bool   `json:"document_handled"`
	SettlementCalculated bool   `json:"settlement_calculated"`
	SettlementPaid       bool   `json:"settlement_paid"`
}

// OffboardingServiceImpl 离职交接服务实现
type OffboardingServiceImpl struct {
	db *gorm.DB
}

// NewOffboardingService 创建离职交接服务
func NewOffboardingService(db *gorm.DB) OffboardingService {
	return &OffboardingServiceImpl{db: db}
}

// InitiateOffboarding 发起离职交接
func (s *OffboardingServiceImpl) InitiateOffboarding(ctx context.Context, req *OffboardingRequest) (*models.OffboardingRecord, error) {
	var result *models.OffboardingRecord
	// 开始事务
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 检查用户是否存在
		var user models.User
		if err := tx.First(&user, req.UserID).Error; err != nil {
			return fmt.Errorf("用户不存在")
		}

		// 2. 检查是否已有未完成的交接记录
		var existing models.OffboardingRecord
		err := tx.Where("user_id = ? AND status IN ?", req.UserID, []string{"pending", "in_progress"}).
			First(&existing).Error
		if err == nil {
			return fmt.Errorf("该用户已有未完成的交接记录")
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		// 3. 获取用户的案件和待办事项
		cases, err := s.getUserCases(ctx, req.UserID)
		if err != nil {
			return fmt.Errorf("获取案件失败: %w", err)
		}

		inboxItems, err := s.getUserInboxItems(ctx, req.UserID)
		if err != nil {
			return fmt.Errorf("获取待办事项失败: %w", err)
		}

		// 4. 序列化案件和待办数据
		casesJSON, _ := json.Marshal(cases)
		inboxJSON, _ := json.Marshal(inboxItems)

		// 将 JSON 字节反序列化为 map[string]interface{}
		var casesData, inboxData map[string]interface{}
		json.Unmarshal(casesJSON, &casesData)
		json.Unmarshal(inboxJSON, &inboxData)

		// 5. 创建交接记录
		record := &models.OffboardingRecord{
			UserID:                 req.UserID,
			InitiatedBy:            req.InitiatedBy,
			InitiatedAt:            time.Now(),
			OriginalCases:          models.JSON(casesData),
			NewLawyerID:            &req.NewLawyerID,
			OriginalInboxItems:     models.JSON(inboxData),
			DocumentDisposalMethod: req.DocumentMethod,
			Status:                 "pending",
			Notes:                  req.Notes,
		}

		if err := tx.Create(record).Error; err != nil {
			return err
		}

		// 6. 创建移交详情记录
		for _, caseItem := range cases {
			detail := &models.OffboardingTransferDetail{
				OffboardingID:   record.ID,
				TransferType:    "case",
				OriginalOwnerID: req.UserID,
				NewOwnerID:      &req.NewLawyerID,
				ItemID:          uint(caseItem["id"].(float64)),
				ItemName:        caseItem["title"].(string),
				TransferStatus:  "pending",
			}
			tx.Create(detail)
		}

		for _, inbox := range inboxItems {
			detail := &models.OffboardingTransferDetail{
				OffboardingID:   record.ID,
				TransferType:    "inbox",
				OriginalOwnerID: req.UserID,
				ItemID:          uint(inbox["id"].(float64)),
				ItemName:        inbox["title"].(string),
				TransferStatus:  "pending",
			}
			tx.Create(detail)
		}

		// 7. 撤销用户所有令牌
		if err := s.revokeAllUserTokens(ctx, tx, req.UserID, req.InitiatedBy); err != nil {
			return fmt.Errorf("撤销令牌失败: %w", err)
		}

		// 8. 更新用户状态
		if err := tx.Model(&models.User{}).
			Where("id = ?", req.UserID).
			Update("offboarding_status", "offboarding").Error; err != nil {
			return fmt.Errorf("更新用户状态失败: %w", err)
		}

		result = record
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOffboardingByID 获取交接记录
func (s *OffboardingServiceImpl) GetOffboardingByID(ctx context.Context, id uint) (*models.OffboardingRecord, error) {
	var record models.OffboardingRecord
	err := s.db.WithContext(ctx).
		Preload("OffboardingTransferDetails").
		First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetOffboardingByUserID 获取用户的交接记录
func (s *OffboardingServiceImpl) GetOffboardingByUserID(ctx context.Context, userID uint) ([]*models.OffboardingRecord, error) {
	var records []models.OffboardingRecord
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	result := make([]*models.OffboardingRecord, len(records))
	for i, r := range records {
		result[i] = &r
	}
	return result, nil
}

// UpdateOffboarding 更新交接记录
func (s *OffboardingServiceImpl) UpdateOffboarding(ctx context.Context, record *models.OffboardingRecord) error {
	return s.db.WithContext(ctx).Save(record).Error
}

// CompleteOffboarding 完成交接
func (s *OffboardingServiceImpl) CompleteOffboarding(ctx context.Context, id uint) error {
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新交接记录状态
		if err := tx.Model(&models.OffboardingRecord{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"status":       "completed",
				"completed_at": &now,
			}).Error; err != nil {
			return err
		}

		// 获取交接记录
		var record models.OffboardingRecord
		if err := tx.First(&record, id).Error; err != nil {
			return err
		}

		// 更新用户状态为已停用
		return tx.Model(&models.User{}).
			Where("id = ?", record.UserID).
			Update("offboarding_status", "deactivated").Error
	})
}

// CancelOffboarding 取消交接
func (s *OffboardingServiceImpl) CancelOffboarding(ctx context.Context, id uint, reason string) error {
	return s.db.WithContext(ctx).Model(&models.OffboardingRecord{}).
		Where("id = ? AND status IN ?", id, []string{"pending", "in_progress"}).
		Updates(map[string]interface{}{
			"status": "cancelled",
			"notes":  gorm.Expr("CONCAT(IFNULL(notes, ''), '\\n取消原因: ', ?)", reason),
		}).Error
}

// TransferCases 移交案件
func (s *OffboardingServiceImpl) TransferCases(ctx context.Context, offboardingID uint, newLawyerID uint, caseIDs []uint) error {
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量更新案件负责人
		if err := tx.Model(&models.Case{}).
			Where("id IN ? AND lead_lawyer_id = ?", caseIDs, offboardingID).
			Updates(map[string]interface{}{
				"lead_lawyer_id": newLawyerID,
				"updated_at":     now,
			}).Error; err != nil {
			return err
		}

		// 更新交接详情记录状态
		return tx.Model(&models.OffboardingTransferDetail{}).
			Where("offboarding_id = ? AND transfer_type = ? AND item_id IN ?", offboardingID, "case", caseIDs).
			Updates(map[string]interface{}{
				"new_owner_id":    newLawyerID,
				"transfer_status": "completed",
				"transferred_at":  &now,
			}).Error
	})
}

// TransferInboxItems 移交待办事项
func (s *OffboardingServiceImpl) TransferInboxItems(ctx context.Context, offboardingID uint, newAssistantID uint, inboxIDs []uint) error {
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量更新待办负责人
		if err := tx.Model(&models.InboxItem{}).
			Where("id IN ? AND user_id = ?", inboxIDs, offboardingID).
			Updates(map[string]interface{}{
				"user_id":    newAssistantID,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		// 更新交接详情记录状态
		return tx.Model(&models.OffboardingTransferDetail{}).
			Where("offboarding_id = ? AND transfer_type = ? AND item_id IN ?", offboardingID, "inbox", inboxIDs).
			Updates(map[string]interface{}{
				"new_owner_id":    newAssistantID,
				"transfer_status": "completed",
				"transferred_at":  &now,
			}).Error
	})
}

// HandleDocuments 处理文档
func (s *OffboardingServiceImpl) HandleDocuments(ctx context.Context, offboardingID uint, method string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新交接记录
		if err := tx.Model(&models.OffboardingRecord{}).
			Where("id = ?", offboardingID).
			Update("document_disposal_method", method).Error; err != nil {
			return err
		}

		// 根据处理方式执行不同操作
		switch method {
		case "delete":
			// 删除用户上传的文档
			// TODO: 实现文档删除逻辑
		case "transfer":
			// 转移文档所有权给新律师
			// TODO: 实现文档转移逻辑
		case "revoke_access":
			// 撤销文档编辑权限
			// TODO: 实现权限撤销逻辑
		}

		// 更新完成时间
		return tx.Model(&models.OffboardingRecord{}).
			Where("id = ?", offboardingID).
			Update("document_disposal_completed_at", &now).Error
	})
}

// GetUserCases 获取用户的案件列表
func (s *OffboardingServiceImpl) GetUserCases(ctx context.Context, userID uint) ([]*models.Case, error) {
	var cases []models.Case
	err := s.db.WithContext(ctx).
		Where("lawyer_id = ?", userID).
		Find(&cases).Error
	if err != nil {
		return nil, err
	}
	result := make([]*models.Case, len(cases))
	for i, c := range cases {
		result[i] = &c
	}
	return result, nil
}

// GetUserInboxItems 获取用户的待办事项列表
func (s *OffboardingServiceImpl) GetUserInboxItems(ctx context.Context, userID uint) ([]*models.InboxItem, error) {
	var items []models.InboxItem
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND is_completed = ?", userID, false).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	result := make([]*models.InboxItem, len(items))
	for i, item := range items {
		result[i] = &item
	}
	return result, nil
}

// GetOffboardingProgress 获取交接进度
func (s *OffboardingServiceImpl) GetOffboardingProgress(ctx context.Context, id uint) (*OffboardingProgress, error) {
	var record models.OffboardingRecord
	if err := s.db.WithContext(ctx).First(&record, id).Error; err != nil {
		return nil, err
	}

	// 获取移交详情统计
	var caseCount, completedCaseCount int64
	var inboxCount, completedInboxCount int64

	s.db.WithContext(ctx).Model(&models.OffboardingTransferDetail{}).
		Where("offboarding_id = ? AND transfer_type = ?", id, "case").
		Count(&caseCount)
	s.db.WithContext(ctx).Model(&models.OffboardingTransferDetail{}).
		Where("offboarding_id = ? AND transfer_type = ? AND transfer_status = ?", id, "case", "completed").
		Count(&completedCaseCount)

	s.db.WithContext(ctx).Model(&models.OffboardingTransferDetail{}).
		Where("offboarding_id = ? AND transfer_type = ?", id, "inbox").
		Count(&inboxCount)
	s.db.WithContext(ctx).Model(&models.OffboardingTransferDetail{}).
		Where("offboarding_id = ? AND transfer_type = ? AND transfer_status = ?", id, "inbox", "completed").
		Count(&completedInboxCount)

	return &OffboardingProgress{
		OffboardingID:        id,
		Status:               record.Status,
		CaseTransferCount:    int(caseCount),
		CaseCompletedCount:   int(completedCaseCount),
		InboxTransferCount:   int(inboxCount),
		InboxCompletedCount:  int(completedInboxCount),
		DocumentHandled:      record.DocumentDisposalCompletedAt != nil,
		SettlementCalculated: record.SettlementCalculated,
		SettlementPaid:       record.SettlementPaid,
	}, nil
}

// BatchTransferCases 批量移交案件
func (s *OffboardingServiceImpl) BatchTransferCases(ctx context.Context, userID uint, newLawyerID uint) error {
	// 获取交接记录
	var record models.OffboardingRecord
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "in_progress").
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return fmt.Errorf("未找到进行中的交接记录")
	}

	// 获取用户案件
	cases, err := s.GetUserCases(ctx, userID)
	if err != nil {
		return err
	}

	caseIDs := make([]uint, len(cases))
	for i, c := range cases {
		caseIDs[i] = c.ID
	}

	return s.TransferCases(ctx, record.ID, newLawyerID, caseIDs)
}

// BatchTransferInboxItems 批量移交待办事项
func (s *OffboardingServiceImpl) BatchTransferInboxItems(ctx context.Context, userID uint, newAssistantID uint) error {
	// 获取交接记录
	var record models.OffboardingRecord
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "in_progress").
		Order("created_at DESC").
		First(&record).Error
	if err != nil {
		return fmt.Errorf("未找到进行中的交接记录")
	}

	// 获取用户待办
	inboxItems, err := s.GetUserInboxItems(ctx, userID)
	if err != nil {
		return err
	}

	inboxIDs := make([]uint, len(inboxItems))
	for i, item := range inboxItems {
		inboxIDs[i] = item.ID
	}

	return s.TransferInboxItems(ctx, record.ID, newAssistantID, inboxIDs)
}

// 私有辅助方法

func (s *OffboardingServiceImpl) getUserCases(ctx context.Context, userID uint) ([]map[string]interface{}, error) {
	var cases []models.Case
	err := s.db.WithContext(ctx).
		Select("id, title").
		Where("lawyer_id = ?", userID).
		Find(&cases).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(cases))
	for i, c := range cases {
		result[i] = map[string]interface{}{
			"id":    c.ID,
			"title": c.Title,
		}
	}
	return result, nil
}

func (s *OffboardingServiceImpl) getUserInboxItems(ctx context.Context, userID uint) ([]map[string]interface{}, error) {
	var items []models.InboxItem
	err := s.db.WithContext(ctx).
		Select("id, title, priority, due_date").
		Where("user_id = ?", userID).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = map[string]interface{}{
			"id":       item.ID,
			"title":    item.Title,
			"priority": item.Priority,
			"due_date": item.DueDate,
		}
	}
	return result, nil
}

func (s *OffboardingServiceImpl) revokeAllUserTokens(ctx context.Context, tx *gorm.DB, userID uint, revokedBy uint) error {
	// 创建撤销记录
	log := &models.TokenRevocationLog{
		UserID:         userID,
		RevocationType: "offboarding",
		RevokedBy:      &revokedBy,
		RevokeAll:      true,
		RevokedAt:      time.Now(),
	}

	return tx.Create(log).Error
}
