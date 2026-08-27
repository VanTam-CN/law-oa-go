package repositories

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// ContractRepository 合同数据仓库接口
type ContractRepository interface {
	// Create 创建合同
	Create(ctx context.Context, contract *models.Contract) error
	// FindByID 根据ID查找合同
	FindByID(ctx context.Context, id uint) (*models.Contract, error)
	// FindByCode 根据合同编号查找合同
	FindByCode(ctx context.Context, code string) (*models.Contract, error)
	// Update 更新合同信息
	Update(ctx context.Context, contract *models.Contract) error
	// Delete 删除合同
	Delete(ctx context.Context, id uint) error
	// List 合同列表查询
	List(ctx context.Context, params *ContractListParams) ([]*models.Contract, int64, error)
	// GetByClientID 获取客户的合同列表
	GetByClientID(ctx context.Context, clientID uint) ([]*models.Contract, error)
	// GetByCaseID 获取案件的合同列表
	GetByCaseID(ctx context.Context, caseID uint) ([]*models.Contract, error)
	// GetSupplementaryContracts 获取主合同的补充协议列表
	GetSupplementaryContracts(ctx context.Context, parentContractID uint) ([]*models.Contract, error)
	// UpdateStatus 更新合同状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// GetDB 获取数据库连接
	GetDB() *gorm.DB
}

// ContractRepositoryImpl 合同数据仓库实现
type ContractRepositoryImpl struct {
	db *gorm.DB
}

// NewContractRepository 创建合同数据仓库实例
func NewContractRepository(db *gorm.DB) ContractRepository {
	return &ContractRepositoryImpl{db: db}
}

// Create 创建合同
func (r *ContractRepositoryImpl) Create(ctx context.Context, contract *models.Contract) error {
	return r.db.WithContext(ctx).Create(contract).Error
}

// FindByID 根据ID查找合同（包含关联数据）
func (r *ContractRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.WithContext(ctx).
		Preload("Client").
		Preload("Case").
		Preload("PaymentMilestones").
		Preload("Document").
		First(&contract, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &contract, nil
}

// FindByCode 根据合同编号查找合同
func (r *ContractRepositoryImpl) FindByCode(ctx context.Context, code string) (*models.Contract, error) {
	var contract models.Contract
	err := r.db.WithContext(ctx).
		Preload("Client").
		Preload("Case").
		Where("contract_code = ?", code).
		First(&contract).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &contract, nil
}

// Update 更新合同信息
func (r *ContractRepositoryImpl) Update(ctx context.Context, contract *models.Contract) error {
	return r.db.WithContext(ctx).Save(contract).Error
}

// Delete 删除合同
func (r *ContractRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Contract{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ContractListParams 合同列表查询参数
type ContractListParams struct {
	Page             int
	PageSize         int
	Status           string
	ContractType     string
	ClientID         uint
	CaseID           uint
	ParentContractID *uint
	Search           string
	StartDateFrom    string
	StartDateTo      string
	EndDateFrom      string
	EndDateTo        string
}

// List 合同列表查询
func (r *ContractRepositoryImpl) List(ctx context.Context, params *ContractListParams) ([]*models.Contract, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.Contract{}).
		Preload("Client").
		Preload("Case")

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 合同类型筛选
	if params.ContractType != "" {
		query = query.Where("contract_type = ?", params.ContractType)
	}

	// 客户筛选
	if params.ClientID > 0 {
		query = query.Where("client_id = ?", params.ClientID)
	}

	// 案件筛选
	if params.CaseID > 0 {
		query = query.Where("case_id = ?", params.CaseID)
	}

	// 主合同筛选（用于查询补充协议）
	if params.ParentContractID != nil {
		query = query.Where("parent_contract_id = ?", *params.ParentContractID)
	}

	// 搜索：合同编号或客户名称
	if params.Search != "" {
		searchLower := strings.ToLower(params.Search)
		searchTerm := "%" + searchLower + "%"
		query = query.Where("LOWER(contract_code) LIKE ?", searchTerm).
			Or("LOWER(contract_code) IN (SELECT LOWER(contract_code) FROM contracts WHERE id IN (SELECT id FROM contracts WHERE LOWER(contract_code) LIKE ?))", searchTerm)
	}

	// 日期范围筛选
	if params.StartDateFrom != "" {
		query = query.Where("start_date >= ?", params.StartDateFrom)
	}
	if params.StartDateTo != "" {
		query = query.Where("start_date <= ?", params.StartDateTo)
	}
	if params.EndDateFrom != "" {
		query = query.Where("end_date >= ?", params.EndDateFrom)
	}
	if params.EndDateTo != "" {
		query = query.Where("end_date <= ?", params.EndDateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var contracts []models.Contract
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&contracts).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.Contract, len(contracts))
	for i, c := range contracts {
		result[i] = &c
	}

	return result, total, nil
}

// GetByClientID 获取客户的合同列表
func (r *ContractRepositoryImpl) GetByClientID(ctx context.Context, clientID uint) ([]*models.Contract, error) {
	var contracts []models.Contract
	err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("created_at DESC").
		Find(&contracts).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Contract, len(contracts))
	for i, c := range contracts {
		result[i] = &c
	}
	return result, nil
}

// GetByCaseID 获取案件的合同列表
func (r *ContractRepositoryImpl) GetByCaseID(ctx context.Context, caseID uint) ([]*models.Contract, error) {
	var contracts []models.Contract
	err := r.db.WithContext(ctx).
		Where("case_id = ?", caseID).
		Order("created_at DESC").
		Find(&contracts).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Contract, len(contracts))
	for i, c := range contracts {
		result[i] = &c
	}
	return result, nil
}

// GetSupplementaryContracts 获取主合同的补充协议列表
func (r *ContractRepositoryImpl) GetSupplementaryContracts(ctx context.Context, parentContractID uint) ([]*models.Contract, error) {
	var contracts []models.Contract
	err := r.db.WithContext(ctx).
		Where("parent_contract_id = ?", parentContractID).
		Order("created_at ASC").
		Find(&contracts).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Contract, len(contracts))
	for i, c := range contracts {
		result[i] = &c
	}
	return result, nil
}

// UpdateStatus 更新合同状态
func (r *ContractRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Contract{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetDB 获取数据库连接
func (r *ContractRepositoryImpl) GetDB() *gorm.DB {
	return r.db
}

// PaymentMilestoneRepository 付款计划数据仓库接口
type PaymentMilestoneRepository interface {
	// Create 创建付款计划
	Create(ctx context.Context, milestone *models.PaymentMilestone) error
	// FindByID 根据ID查找付款计划
	FindByID(ctx context.Context, id uint) (*models.PaymentMilestone, error)
	// Update 更新付款计划
	Update(ctx context.Context, milestone *models.PaymentMilestone) error
	// Delete 删除付款计划
	Delete(ctx context.Context, id uint) error
	// GetByContractID 获取合同的付款计划列表
	GetByContractID(ctx context.Context, contractID uint) ([]*models.PaymentMilestone, error)
	// DeleteByContractID 删除合同的所有付款计划
	DeleteByContractID(ctx context.Context, contractID uint) error
	// UpdateStatus 更新付款计划状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// UpdateInvoiceID 更新关联的发票ID
	UpdateInvoiceID(ctx context.Context, id uint, invoiceID *uint) error
	// GetPendingMilestones 获取待开票的付款计划
	GetPendingMilestones(ctx context.Context) ([]*models.PaymentMilestone, error)
	// GetOverdueMilestones 获取已过期的付款计划
	GetOverdueMilestones(ctx context.Context) ([]*models.PaymentMilestone, error)
}

// PaymentMilestoneRepositoryImpl 付款计划数据仓库实现
type PaymentMilestoneRepositoryImpl struct {
	db *gorm.DB
}

// NewPaymentMilestoneRepository 创建付款计划数据仓库实例
func NewPaymentMilestoneRepository(db *gorm.DB) PaymentMilestoneRepository {
	return &PaymentMilestoneRepositoryImpl{db: db}
}

// Create 创建付款计划
func (r *PaymentMilestoneRepositoryImpl) Create(ctx context.Context, milestone *models.PaymentMilestone) error {
	return r.db.WithContext(ctx).Create(milestone).Error
}

// FindByID 根据ID查找付款计划
func (r *PaymentMilestoneRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.PaymentMilestone, error) {
	var milestone models.PaymentMilestone
	err := r.db.WithContext(ctx).First(&milestone, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &milestone, nil
}

// Update 更新付款计划
func (r *PaymentMilestoneRepositoryImpl) Update(ctx context.Context, milestone *models.PaymentMilestone) error {
	return r.db.WithContext(ctx).Save(milestone).Error
}

// Delete 删除付款计划
func (r *PaymentMilestoneRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.PaymentMilestone{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetByContractID 获取合同的付款计划列表
func (r *PaymentMilestoneRepositoryImpl) GetByContractID(ctx context.Context, contractID uint) ([]*models.PaymentMilestone, error) {
	var milestones []models.PaymentMilestone
	err := r.db.WithContext(ctx).
		Where("contract_id = ?", contractID).
		Order("sequence ASC").
		Find(&milestones).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.PaymentMilestone, len(milestones))
	for i, m := range milestones {
		result[i] = &m
	}
	return result, nil
}

// DeleteByContractID 删除合同的所有付款计划
func (r *PaymentMilestoneRepositoryImpl) DeleteByContractID(ctx context.Context, contractID uint) error {
	return r.db.WithContext(ctx).
		Where("contract_id = ?", contractID).
		Delete(&models.PaymentMilestone{}).Error
}

// UpdateStatus 更新付款计划状态
func (r *PaymentMilestoneRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).
		Model(&models.PaymentMilestone{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateInvoiceID 更新关联的发票ID
func (r *PaymentMilestoneRepositoryImpl) UpdateInvoiceID(ctx context.Context, id uint, invoiceID *uint) error {
	result := r.db.WithContext(ctx).
		Model(&models.PaymentMilestone{}).
		Where("id = ?", id).
		Update("invoice_id", invoiceID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetPendingMilestones 获取待开票的付款计划
func (r *PaymentMilestoneRepositoryImpl) GetPendingMilestones(ctx context.Context) ([]*models.PaymentMilestone, error) {
	var milestones []models.PaymentMilestone
	err := r.db.WithContext(ctx).
		Where("status = ? AND due_date <= NOW()", "pending").
		Preload("Contract").
		Preload("Contract.Client").
		Order("due_date ASC").
		Find(&milestones).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.PaymentMilestone, len(milestones))
	for i, m := range milestones {
		result[i] = &m
	}
	return result, nil
}

// GetOverdueMilestones 获取已过期的付款计划
func (r *PaymentMilestoneRepositoryImpl) GetOverdueMilestones(ctx context.Context) ([]*models.PaymentMilestone, error) {
	var milestones []models.PaymentMilestone
	err := r.db.WithContext(ctx).
		Where("status IN (?, ?) AND due_date < NOW()", "pending", "billed").
		Preload("Contract").
		Preload("Contract.Client").
		Order("due_date ASC").
		Find(&milestones).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.PaymentMilestone, len(milestones))
	for i, m := range milestones {
		result[i] = &m
	}
	return result, nil
}

// InvoiceRepository 发票数据仓库接口
type InvoiceRepository interface {
	// Create 创建发票
	Create(ctx context.Context, invoice *models.Invoice) error
	// FindByID 根据ID查找发票
	FindByID(ctx context.Context, id uint) (*models.Invoice, error)
	// FindByCode 根据发票编号查找发票
	FindByCode(ctx context.Context, code string) (*models.Invoice, error)
	// Update 更新发票
	Update(ctx context.Context, invoice *models.Invoice) error
	// Delete 删除发票
	Delete(ctx context.Context, id uint) error
	// List 发票列表查询
	List(ctx context.Context, params *InvoiceListParams) ([]*models.Invoice, int64, error)
	// GetByContractID 获取合同的发票列表
	GetByContractID(ctx context.Context, contractID uint) ([]*models.Invoice, error)
	// GetByClientID 获取客户的发票列表
	GetByClientID(ctx context.Context, clientID uint) ([]*models.Invoice, error)
	// UpdateStatus 更新发票状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// GetPendingInvoices 获取待审批的发票列表
	GetPendingInvoices(ctx context.Context) ([]*models.Invoice, error)
	// GetOverdueInvoices 获取逾期未回款的发票
	GetOverdueInvoices(ctx context.Context) ([]*models.Invoice, error)
}

// InvoiceRepositoryImpl 发票数据仓库实现
type InvoiceRepositoryImpl struct {
	db *gorm.DB
}

// NewInvoiceRepository 创建发票数据仓库实例
func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &InvoiceRepositoryImpl{db: db}
}

// Create 创建发票
func (r *InvoiceRepositoryImpl) Create(ctx context.Context, invoice *models.Invoice) error {
	return r.db.WithContext(ctx).Create(invoice).Error
}

// FindByID 根据ID查找发票
func (r *InvoiceRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).
		Preload("Contract").
		Preload("Milestone").
		Preload("Client").
		First(&invoice, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

// FindByCode 根据发票编号查找发票
func (r *InvoiceRepositoryImpl) FindByCode(ctx context.Context, code string) (*models.Invoice, error) {
	var invoice models.Invoice
	err := r.db.WithContext(ctx).
		Preload("Contract").
		Preload("Milestone").
		Where("invoice_code = ?", code).
		First(&invoice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invoice, nil
}

// Update 更新发票
func (r *InvoiceRepositoryImpl) Update(ctx context.Context, invoice *models.Invoice) error {
	return r.db.WithContext(ctx).Save(invoice).Error
}

// Delete 删除发票
func (r *InvoiceRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Invoice{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// InvoiceListParams 发票列表查询参数
type InvoiceListParams struct {
	Page        int
	PageSize    int
	Status      string
	InvoiceType string
	ClientID    uint
	ContractID  uint
	Search      string
	DateFrom    string
	DateTo      string
}

// List 发票列表查询
func (r *InvoiceRepositoryImpl) List(ctx context.Context, params *InvoiceListParams) ([]*models.Invoice, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.Invoice{}).
		Preload("Client").
		Preload("Contract")

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 发票类型筛选
	if params.InvoiceType != "" {
		query = query.Where("invoice_type = ?", params.InvoiceType)
	}

	// 客户筛选
	if params.ClientID > 0 {
		query = query.Where("client_id = ?", params.ClientID)
	}

	// 合同筛选
	if params.ContractID > 0 {
		query = query.Where("contract_id = ?", params.ContractID)
	}

	// 搜索：发票编号或客户名称
	if params.Search != "" {
		searchLower := strings.ToLower(params.Search)
		searchTerm := "%" + searchLower + "%"
		query = query.Where("LOWER(invoice_code) LIKE ? OR LOWER(client_name) LIKE ?", searchTerm, searchTerm)
	}

	// 日期范围筛选
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var invoices []models.Invoice
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = &inv
	}

	return result, total, nil
}

// GetByContractID 获取合同的发票列表
func (r *InvoiceRepositoryImpl) GetByContractID(ctx context.Context, contractID uint) ([]*models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.WithContext(ctx).
		Where("contract_id = ?", contractID).
		Order("created_at DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = &inv
	}
	return result, nil
}

// GetByClientID 获取客户的发票列表
func (r *InvoiceRepositoryImpl) GetByClientID(ctx context.Context, clientID uint) ([]*models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("created_at DESC").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = &inv
	}
	return result, nil
}

// UpdateStatus 更新发票状态
func (r *InvoiceRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetPendingInvoices 获取待审批的发票列表
func (r *InvoiceRepositoryImpl) GetPendingInvoices(ctx context.Context) ([]*models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.WithContext(ctx).
		Where("status = ?", "submitted").
		Preload("Client").
		Preload("Contract").
		Order("created_at ASC").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = &inv
	}
	return result, nil
}

// GetOverdueInvoices 获取逾期未回款的发票
func (r *InvoiceRepositoryImpl) GetOverdueInvoices(ctx context.Context) ([]*models.Invoice, error) {
	var invoices []models.Invoice
	err := r.db.WithContext(ctx).
		Where("status IN (?, ?, ?) AND created_at < NOW() - INTERVAL '30 days'", "approved", "issued", "received").
		Preload("Client").
		Preload("Contract").
		Order("created_at ASC").
		Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = &inv
	}
	return result, nil
}

// PaymentRepository 回款记录数据仓库接口
type PaymentRepository interface {
	// Create 创建回款记录
	Create(ctx context.Context, payment *models.Payment) error
	// FindByID 根据ID查找回款记录
	FindByID(ctx context.Context, id uint) (*models.Payment, error)
	// FindByCode 根据回款编号查找回款记录
	FindByCode(ctx context.Context, code string) (*models.Payment, error)
	// Update 更新回款记录
	Update(ctx context.Context, payment *models.Payment) error
	// Delete 删除回款记录
	Delete(ctx context.Context, id uint) error
	// List 回款记录列表查询
	List(ctx context.Context, params *PaymentListParams) ([]*models.Payment, int64, error)
	// GetByInvoiceID 获取发票的回款记录
	GetByInvoiceID(ctx context.Context, invoiceID uint) ([]*models.Payment, error)
	// GetTotalPaidAmount 获取发票的总回款金额
	GetTotalPaidAmount(ctx context.Context, invoiceID uint) (float64, error)
	// UpdateStatus 更新回款状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// GetByPaymentDateRange 获取日期范围内的回款记录
	GetByPaymentDateRange(ctx context.Context, startDate, endDate string) ([]*models.Payment, error)
	// GetPaymentAggregation 获取回款聚合统计（按状态分组求和）
	GetPaymentAggregation(ctx context.Context, monthStart, monthEnd string) (total int64, pendingCount int64, confirmedCount int64, rejectedCount int64, totalAmount float64, monthAmount float64, pendingAmount float64, err error)
}

// PaymentRepositoryImpl 回款记录数据仓库实现
type PaymentRepositoryImpl struct {
	db *gorm.DB
}

// NewPaymentRepository 创建回款记录数据仓库实例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &PaymentRepositoryImpl{db: db}
}

// Create 创建回款记录
func (r *PaymentRepositoryImpl) Create(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// FindByID 根据ID查找回款记录
func (r *PaymentRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Preload("Invoice").
		First(&payment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// FindByCode 根据回款编号查找回款记录
func (r *PaymentRepositoryImpl) FindByCode(ctx context.Context, code string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Preload("Invoice").
		Where("payment_code = ?", code).
		First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// Update 更新回款记录
func (r *PaymentRepositoryImpl) Update(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// Delete 删除回款记录
func (r *PaymentRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Payment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// PaymentListParams 回款记录列表查询参数
type PaymentListParams struct {
	Page      int
	PageSize  int
	Status    string
	InvoiceID uint
	ClientID  uint
	Search    string
	DateFrom  string
	DateTo    string
}

// List 回款记录列表查询
func (r *PaymentRepositoryImpl) List(ctx context.Context, params *PaymentListParams) ([]*models.Payment, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Preload("Invoice").
		Preload("Invoice.Client")

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 发票筛选
	if params.InvoiceID > 0 {
		query = query.Where("invoice_id = ?", params.InvoiceID)
	}

	// 客户筛选（通过发票关联）
	if params.ClientID > 0 {
		query = query.Joins("JOIN invoices ON invoices.id = payments.invoice_id").
			Where("invoices.client_id = ?", params.ClientID)
	}

	// 搜索：回款编号或付款人
	if params.Search != "" {
		searchLower := strings.ToLower(params.Search)
		searchTerm := "%" + searchLower + "%"
		query = query.Where("LOWER(payment_code) LIKE ? OR LOWER(payer_name) LIKE ?", searchTerm, searchTerm)
	}

	// 日期范围筛选
	if params.DateFrom != "" {
		query = query.Where("payment_date >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("payment_date <= ?", params.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var payments []models.Payment
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("payment_date DESC").Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.Payment, len(payments))
	for i, p := range payments {
		result[i] = &p
	}

	return result, total, nil
}

// GetByInvoiceID 获取发票的回款记录
func (r *PaymentRepositoryImpl) GetByInvoiceID(ctx context.Context, invoiceID uint) ([]*models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("invoice_id = ?", invoiceID).
		Order("payment_date DESC").
		Find(&payments).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Payment, len(payments))
	for i, p := range payments {
		result[i] = &p
	}
	return result, nil
}

// GetTotalPaidAmount 获取发票的总回款金额
func (r *PaymentRepositoryImpl) GetTotalPaidAmount(ctx context.Context, invoiceID uint) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("invoice_id = ? AND status = ?", invoiceID, "confirmed").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

// UpdateStatus 更新回款状态
func (r *PaymentRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Payment{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetByPaymentDateRange 获取日期范围内的回款记录
func (r *PaymentRepositoryImpl) GetByPaymentDateRange(ctx context.Context, startDate, endDate string) ([]*models.Payment, error) {
	var payments []models.Payment
	err := r.db.WithContext(ctx).
		Where("payment_date >= ? AND payment_date <= ?", startDate, endDate).
		Preload("Invoice").
		Preload("Invoice.Client").
		Order("payment_date DESC").
		Find(&payments).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.Payment, len(payments))
	for i, p := range payments {
		result[i] = &p
	}
	return result, nil
}

// BadDebtRepository 坏账核销记录数据仓库接口
type BadDebtRepository interface {
	// Create 创建坏账核销记录
	Create(ctx context.Context, badDebt *models.BadDebtRecord) error
	// FindByID 根据ID查找坏账核销记录
	FindByID(ctx context.Context, id uint) (*models.BadDebtRecord, error)
	// Update 更新坏账核销记录
	Update(ctx context.Context, badDebt *models.BadDebtRecord) error
	// Delete 删除坏账核销记录
	Delete(ctx context.Context, id uint) error
	// List 坏账核销记录列表查询
	List(ctx context.Context, params *BadDebtListParams) ([]*models.BadDebtRecord, int64, error)
	// GetByContractID 获取合同的坏账核销记录
	GetByContractID(ctx context.Context, contractID uint) ([]*models.BadDebtRecord, error)
	// GetPending 获取待审批的坏账核销记录
	GetPending(ctx context.Context) ([]*models.BadDebtRecord, error)
	// UpdateStatus 更新坏账核销状态
	UpdateStatus(ctx context.Context, id uint, status string) error
}

// BadDebtRepositoryImpl 坏账核销记录数据仓库实现
type BadDebtRepositoryImpl struct {
	db *gorm.DB
}

// NewBadDebtRepository 创建坏账核销记录数据仓库实例
func NewBadDebtRepository(db *gorm.DB) BadDebtRepository {
	return &BadDebtRepositoryImpl{db: db}
}

// Create 创建坏账核销记录
func (r *BadDebtRepositoryImpl) Create(ctx context.Context, badDebt *models.BadDebtRecord) error {
	return r.db.WithContext(ctx).Create(badDebt).Error
}

// FindByID 根据ID查找坏账核销记录
func (r *BadDebtRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.BadDebtRecord, error) {
	var badDebt models.BadDebtRecord
	err := r.db.WithContext(ctx).
		Preload("Contract").
		Preload("Invoice").
		First(&badDebt, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &badDebt, nil
}

// Update 更新坏账核销记录
func (r *BadDebtRepositoryImpl) Update(ctx context.Context, badDebt *models.BadDebtRecord) error {
	return r.db.WithContext(ctx).Save(badDebt).Error
}

// Delete 删除坏账核销记录
func (r *BadDebtRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.BadDebtRecord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// BadDebtListParams 坏账核销记录列表查询参数
type BadDebtListParams struct {
	Page       int
	PageSize   int
	Status     string
	ContractID uint
	ReasonType string
}

// List 坏账核销记录列表查询
func (r *BadDebtRepositoryImpl) List(ctx context.Context, params *BadDebtListParams) ([]*models.BadDebtRecord, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.BadDebtRecord{}).
		Preload("Contract").
		Preload("Contract.Client")

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 合同筛选
	if params.ContractID > 0 {
		query = query.Where("contract_id = ?", params.ContractID)
	}

	// 原因类型筛选
	if params.ReasonType != "" {
		query = query.Where("reason_type = ?", params.ReasonType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var badDebts []models.BadDebtRecord
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&badDebts).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.BadDebtRecord, len(badDebts))
	for i, bd := range badDebts {
		result[i] = &bd
	}

	return result, total, nil
}

// GetByContractID 获取合同的坏账核销记录
func (r *BadDebtRepositoryImpl) GetByContractID(ctx context.Context, contractID uint) ([]*models.BadDebtRecord, error) {
	var badDebts []models.BadDebtRecord
	err := r.db.WithContext(ctx).
		Where("contract_id = ?", contractID).
		Order("created_at DESC").
		Find(&badDebts).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.BadDebtRecord, len(badDebts))
	for i, bd := range badDebts {
		result[i] = &bd
	}
	return result, nil
}

// GetPending 获取待审批的坏账核销记录
func (r *BadDebtRepositoryImpl) GetPending(ctx context.Context) ([]*models.BadDebtRecord, error) {
	var badDebts []models.BadDebtRecord
	err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Preload("Contract").
		Preload("Contract.Client").
		Order("created_at ASC").
		Find(&badDebts).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.BadDebtRecord, len(badDebts))
	for i, bd := range badDebts {
		result[i] = &bd
	}
	return result, nil
}

// UpdateStatus 更新坏账核销状态
func (r *BadDebtRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.BadDebtRecord{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CommissionRepository 提成记录数据仓库接口
type CommissionRepository interface {
	// Create 创建提成记录
	Create(ctx context.Context, commission *models.CommissionRecord) error
	// FindByID 根据ID查找提成记录
	FindByID(ctx context.Context, id uint) (*models.CommissionRecord, error)
	// FindByCode 根据提成编号查找提成记录
	FindByCode(ctx context.Context, code string) (*models.CommissionRecord, error)
	// Update 更新提成记录
	Update(ctx context.Context, commission *models.CommissionRecord) error
	// Delete 删除提成记录
	Delete(ctx context.Context, id uint) error
	// List 提成记录列表查询
	List(ctx context.Context, params *CommissionListParams) ([]*models.CommissionRecord, int64, error)
	// GetByPaymentID 获取回款的提成记录
	GetByPaymentID(ctx context.Context, paymentID uint) ([]*models.CommissionRecord, error)
	// GetByContractID 获取合同的提成记录
	GetByContractID(ctx context.Context, contractID uint) ([]*models.CommissionRecord, error)
	// GetByBeneficiaryID 获取受益人的提成记录
	GetByBeneficiaryID(ctx context.Context, beneficiaryID uint) ([]*models.CommissionRecord, error)
	// UpdateStatus 更新提成状态
	UpdateStatus(ctx context.Context, id uint, status string) error
	// GetPending 获取待支付的提成记录
	GetPending(ctx context.Context) ([]*models.CommissionRecord, error)
	// GetTotalCommission 获取合同的总提成金额
	GetTotalCommission(ctx context.Context, contractID uint) (float64, error)
}

// CommissionRepositoryImpl 提成记录数据仓库实现
type CommissionRepositoryImpl struct {
	db *gorm.DB
}

// NewCommissionRepository 创建提成记录数据仓库实例
func NewCommissionRepository(db *gorm.DB) CommissionRepository {
	return &CommissionRepositoryImpl{db: db}
}

// Create 创建提成记录
func (r *CommissionRepositoryImpl) Create(ctx context.Context, commission *models.CommissionRecord) error {
	return r.db.WithContext(ctx).Create(commission).Error
}

// FindByID 根据ID查找提成记录
func (r *CommissionRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.CommissionRecord, error) {
	var commission models.CommissionRecord
	err := r.db.WithContext(ctx).
		Preload("Contract").
		Preload("Payment").
		Preload("Case").
		First(&commission, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &commission, nil
}

// FindByCode 根据提成编号查找提成记录
func (r *CommissionRepositoryImpl) FindByCode(ctx context.Context, code string) (*models.CommissionRecord, error) {
	var commission models.CommissionRecord
	err := r.db.WithContext(ctx).
		Preload("Contract").
		Where("commission_code = ?", code).
		First(&commission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &commission, nil
}

// Update 更新提成记录
func (r *CommissionRepositoryImpl) Update(ctx context.Context, commission *models.CommissionRecord) error {
	return r.db.WithContext(ctx).Save(commission).Error
}

// Delete 删除提成记录
func (r *CommissionRepositoryImpl) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.CommissionRecord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CommissionListParams 提成记录列表查询参数
type CommissionListParams struct {
	Page            int
	PageSize        int
	Status          string
	BeneficiaryID   uint
	BeneficiaryRole string
	ContractID      uint
	CaseID          *uint
	DateFrom        string
	DateTo          string
}

// List 提成记录列表查询
func (r *CommissionRepositoryImpl) List(ctx context.Context, params *CommissionListParams) ([]*models.CommissionRecord, int64, error) {
	page := 1
	pageSize := 20
	if params.Page > 0 {
		page = params.Page
	}
	if params.PageSize > 0 && params.PageSize <= 100 {
		pageSize = params.PageSize
	}

	query := r.db.WithContext(ctx).Model(&models.CommissionRecord{}).
		Preload("Contract").
		Preload("Contract.Client").
		Preload("Payment").
		Preload("Case")

	// 状态筛选
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// 受益人筛选
	if params.BeneficiaryID > 0 {
		query = query.Where("beneficiary_id = ?", params.BeneficiaryID)
	}

	// 受益人角色筛选
	if params.BeneficiaryRole != "" {
		query = query.Where("beneficiary_role = ?", params.BeneficiaryRole)
	}

	// 合同筛选
	if params.ContractID > 0 {
		query = query.Where("contract_id = ?", params.ContractID)
	}

	// 案件筛选
	if params.CaseID != nil {
		query = query.Where("case_id = ?", *params.CaseID)
	}

	// 日期范围筛选
	if params.DateFrom != "" {
		query = query.Where("created_at >= ?", params.DateFrom)
	}
	if params.DateTo != "" {
		query = query.Where("created_at <= ?", params.DateTo)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var commissions []models.CommissionRecord
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&commissions).Error; err != nil {
		return nil, 0, err
	}

	result := make([]*models.CommissionRecord, len(commissions))
	for i, c := range commissions {
		result[i] = &c
	}

	return result, total, nil
}

// GetByPaymentID 获取回款的提成记录
func (r *CommissionRepositoryImpl) GetByPaymentID(ctx context.Context, paymentID uint) ([]*models.CommissionRecord, error) {
	var commissions []models.CommissionRecord
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&commissions).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.CommissionRecord, len(commissions))
	for i, c := range commissions {
		result[i] = &c
	}
	return result, nil
}

// GetByContractID 获取合同的提成记录
func (r *CommissionRepositoryImpl) GetByContractID(ctx context.Context, contractID uint) ([]*models.CommissionRecord, error) {
	var commissions []models.CommissionRecord
	err := r.db.WithContext(ctx).
		Where("contract_id = ?", contractID).
		Preload("Payment").
		Order("created_at DESC").
		Find(&commissions).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.CommissionRecord, len(commissions))
	for i, c := range commissions {
		result[i] = &c
	}
	return result, nil
}

// GetByBeneficiaryID 获取受益人的提成记录
func (r *CommissionRepositoryImpl) GetByBeneficiaryID(ctx context.Context, beneficiaryID uint) ([]*models.CommissionRecord, error) {
	var commissions []models.CommissionRecord
	err := r.db.WithContext(ctx).
		Where("beneficiary_id = ?", beneficiaryID).
		Preload("Contract").
		Preload("Contract.Client").
		Preload("Payment").
		Order("created_at DESC").
		Find(&commissions).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.CommissionRecord, len(commissions))
	for i, c := range commissions {
		result[i] = &c
	}
	return result, nil
}

// UpdateStatus 更新提成状态
func (r *CommissionRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.CommissionRecord{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetPending 获取待支付的提成记录
func (r *CommissionRepositoryImpl) GetPending(ctx context.Context) ([]*models.CommissionRecord, error) {
	var commissions []models.CommissionRecord
	err := r.db.WithContext(ctx).
		Where("status = ?", "calculated").
		Preload("Contract").
		Preload("Contract.Client").
		Order("created_at ASC").
		Find(&commissions).Error
	if err != nil {
		return nil, err
	}

	result := make([]*models.CommissionRecord, len(commissions))
	for i, c := range commissions {
		result[i] = &c
	}
	return result, nil
}

// GetTotalCommission 获取合同的总提成金额
func (r *CommissionRepositoryImpl) GetTotalCommission(ctx context.Context, contractID uint) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&models.CommissionRecord{}).
		Where("contract_id = ?", contractID).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&total).Error
	return total, err
}
