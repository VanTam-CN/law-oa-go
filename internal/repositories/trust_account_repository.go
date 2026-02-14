package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
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
