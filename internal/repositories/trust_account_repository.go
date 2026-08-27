package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"law-oa-go/internal/models"
)

// TrustAccountRepositoryImpl 代管款账户仓储实现
type TrustAccountRepositoryImpl struct {
	db *gorm.DB
}

// NewTrustAccountRepository 创建代管款账户仓储实例
func NewTrustAccountRepository(db *gorm.DB) TrustAccountRepository {
	return &TrustAccountRepositoryImpl{db: db}
}

// Create 创建账户
func (r *TrustAccountRepositoryImpl) Create(ctx context.Context, account *models.ClientTrustAccount) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// FindByID 根据ID查找账户
func (r *TrustAccountRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.ClientTrustAccount, error) {
	var account models.ClientTrustAccount
	err := r.db.WithContext(ctx).First(&account, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// FindByIDForUpdate 加行锁读取账户（FOR UPDATE），必须在事务内调用
func (r *TrustAccountRepositoryImpl) FindByIDForUpdate(ctx context.Context, id uint) (*models.ClientTrustAccount, error) {
	var account models.ClientTrustAccount
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&account, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// FindByCode 根据账户编号查找账户
func (r *TrustAccountRepositoryImpl) FindByCode(ctx context.Context, code string) (*models.ClientTrustAccount, error) {
	var account models.ClientTrustAccount
	err := r.db.WithContext(ctx).Where("account_code = ?", code).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Update 更新账户
func (r *TrustAccountRepositoryImpl) Update(ctx context.Context, account *models.ClientTrustAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// UpdateBalance 直接按 ID 写入新余额，避免 Save 全字段覆盖
func (r *TrustAccountRepositoryImpl) UpdateBalance(ctx context.Context, id uint, newBalance float64) error {
	result := r.db.WithContext(ctx).
		Model(&models.ClientTrustAccount{}).
		Where("id = ?", id).
		Update("balance", newBalance)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除账户
func (r *TrustAccountRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ClientTrustAccount{}, id).Error
}

// List 账户列表查询
func (r *TrustAccountRepositoryImpl) List(ctx context.Context, params *TrustAccountListParams) ([]*models.ClientTrustAccount, int64, error) {
	var accounts []*models.ClientTrustAccount
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ClientTrustAccount{})

	// 筛选条件
	if params.ClientID > 0 {
		query = query.Where("client_id = ?", params.ClientID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Currency != "" {
		query = query.Where("currency = ?", params.Currency)
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (params.Page - 1) * params.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&accounts).Error
	if err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

// GetByClientID 获取客户的账户列表
func (r *TrustAccountRepositoryImpl) GetByClientID(ctx context.Context, clientID uint) ([]*models.ClientTrustAccount, int64, error) {
	var accounts []*models.ClientTrustAccount
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ClientTrustAccount{}).Where("client_id = ?", clientID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Find(&accounts).Error
	if err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

// TrustTransactionRepositoryImpl 代管款交易仓储实现
type TrustTransactionRepositoryImpl struct {
	db *gorm.DB
}

// NewTrustTransactionRepository 创建代管款交易仓储实例
func NewTrustTransactionRepository(db *gorm.DB) TrustTransactionRepository {
	return &TrustTransactionRepositoryImpl{db: db}
}

// Create 创建交易
func (r *TrustTransactionRepositoryImpl) Create(ctx context.Context, transaction *models.ClientTrustTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// FindByID 根据ID查找交易
func (r *TrustTransactionRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.ClientTrustTransaction, error) {
	var transaction models.ClientTrustTransaction
	err := r.db.WithContext(ctx).First(&transaction, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

// FindByIDForUpdate 加行锁读取交易（FOR UPDATE），必须在事务内调用
func (r *TrustTransactionRepositoryImpl) FindByIDForUpdate(ctx context.Context, id uint) (*models.ClientTrustTransaction, error) {
	var transaction models.ClientTrustTransaction
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&transaction, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &transaction, nil
}

// FindByCode 根据交易编号查找交易
func (r *TrustTransactionRepositoryImpl) FindByCode(ctx context.Context, code string) (*models.ClientTrustTransaction, error) {
	var transaction models.ClientTrustTransaction
	err := r.db.WithContext(ctx).Where("transaction_code = ?", code).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// Update 更新交易
func (r *TrustTransactionRepositoryImpl) Update(ctx context.Context, transaction *models.ClientTrustTransaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

// UpdateStatusIfPending 条件更新：仅当当前 status='pending' 时才写入新状态与审批信息
// 返回受影响行数：1=成功；0=已被并发改动（调用方据此判断幂等冲突）
func (r *TrustTransactionRepositoryImpl) UpdateStatusIfPending(ctx context.Context, id uint, newStatus string, approvedBy uint) (int64, error) {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      newStatus,
		"approved_by": approvedBy,
		"approved_at": &now,
	}
	if newStatus == "completed" {
		updates["completed_at"] = &now
	}

	result := r.db.WithContext(ctx).
		Model(&models.ClientTrustTransaction{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(updates)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// Delete 删除交易
func (r *TrustTransactionRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ClientTrustTransaction{}, id).Error
}

// List 交易列表查询
func (r *TrustTransactionRepositoryImpl) List(ctx context.Context, params *TrustTransactionListParams) ([]*models.ClientTrustTransaction, int64, error) {
	var transactions []*models.ClientTrustTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ClientTrustTransaction{})

	// 筛选条件
	if params.AccountID != nil {
		query = query.Where("account_id = ?", *params.AccountID)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Type != "" {
		query = query.Where("transaction_type = ?", params.Type)
	}
	if params.DateFrom != "" {
		if startTime, err := time.Parse("2006-01-02", params.DateFrom); err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}
	if params.DateTo != "" {
		if endTime, err := time.Parse("2006-01-02", params.DateTo); err == nil {
			// 添加一天以包含结束日期当天的所有记录
			endTime = endTime.AddDate(0, 0, 1)
			query = query.Where("created_at < ?", endTime)
		}
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (params.Page - 1) * params.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetByAccountID 获取账户的交易列表
func (r *TrustTransactionRepositoryImpl) GetByAccountID(ctx context.Context, accountID uint) ([]*models.ClientTrustTransaction, int64, error) {
	var transactions []*models.ClientTrustTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ClientTrustTransaction{}).Where("account_id = ?", accountID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// trustUnitOfWorkImpl 代管款事务工作单元实现
// 同一事务内的所有 repo 共享同一个 *gorm.DB（tx），保证原子性
type trustUnitOfWorkImpl struct {
	db *gorm.DB
}

// NewTrustUnitOfWork 创建代管款事务工作单元实例
func NewTrustUnitOfWork(db *gorm.DB) TrustUnitOfWork {
	return &trustUnitOfWorkImpl{db: db}
}

// WithinTransaction 在数据库事务内执行 fn；fn 返回 error 触发回滚
// fn 内拿到的 repo 已绑定到 tx，所有读写都在同一事务内
func (u *trustUnitOfWorkImpl) WithinTransaction(
	ctx context.Context,
	fn func(txTxnRepo TrustTransactionRepository, txAcctRepo TrustAccountRepository) error,
) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txTxnRepo := NewTrustTransactionRepository(tx)
		txAcctRepo := NewTrustAccountRepository(tx)
		return fn(txTxnRepo, txAcctRepo)
	})
}
