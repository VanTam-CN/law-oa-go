package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// TrustAccountService 代管款账户服务
type TrustAccountService struct {
	accountRepo     repositories.TrustAccountRepository
	transactionRepo repositories.TrustTransactionRepository
	clientRepo      repositories.ClientRepository
	caseRepo        repositories.CaseRepository
	userRepo        repositories.UserRepository
}

// NewTrustAccountService 创建代管款账户服务实例
func NewTrustAccountService(
	accountRepo repositories.TrustAccountRepository,
	transactionRepo repositories.TrustTransactionRepository,
	clientRepo repositories.ClientRepository,
	caseRepo repositories.CaseRepository,
	userRepo repositories.UserRepository,
) *TrustAccountService {
	return &TrustAccountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		clientRepo:      clientRepo,
		caseRepo:        caseRepo,
		userRepo:        userRepo,
	}
}

// CreateAccountRequest 创建账户请求
type CreateAccountRequest struct {
	ClientID           uint     `json:"client_id" binding:"required"`
	Currency           string   `json:"currency" binding:"omitempty,oneof=CNY USD EUR"`
	PurposeRestriction string   `json:"purpose_restriction" binding:"max=200"`
	AuthorizedUses     []string `json:"authorized_uses"`
}

// AccountResponse 账户响应
type AccountResponse struct {
	ID                 uint              `json:"id"`
	ClientID           uint              `json:"client_id"`
	AccountCode        string            `json:"account_code"`
	Balance            float64           `json:"balance"`
	Currency           string            `json:"currency"`
	FrozenAmount       float64           `json:"frozen_amount"`
	AvailableBalance   float64           `json:"available_balance"`
	PurposeRestriction string            `json:"purpose_restriction"`
	AuthorizedUses     []string          `json:"authorized_uses"`
	Status             string            `json:"status"`
	OpenedAt           *string           `json:"opened_at,omitempty"`
	ClosedAt           *string           `json:"closed_at,omitempty"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
	// 关联数据
	Client             *ClientSummary    `json:"client,omitempty"`
	RecentTransactions []*TransactionSummary `json:"recent_transactions,omitempty"`
}

// TransactionSummary 交易摘要
type TransactionSummary struct {
	ID              uint    `json:"id"`
	TransactionCode string  `json:"transaction_code"`
	TransactionType string  `json:"transaction_type"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"created_at"`
}

// ListAccountsRequest 账户列表请求
type ListAccountsRequest struct {
	Page      int    `json:"page" form:"page" binding:"min=1"`
	PageSize  int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	ClientID  uint   `json:"client_id" form:"client_id"`
	Status    string `json:"status" form:"status" binding:"omitempty,oneof=active frozen closed"`
	Currency  string `json:"currency" form:"currency" binding:"omitempty,oneof=CNY USD EUR"`
}

// ListAccountsResponse 账户列表响应
type ListAccountsResponse struct {
	Accounts  []*AccountResponse `json:"accounts"`
	Pagination Pagination         `json:"pagination"`
}

// CreateAccount 创建代管款账户
func (s *TrustAccountService) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*AccountResponse, error) {
	// 验证客户是否存在
	client, err := s.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("查询客户失败: %w", err)
	}
	if client == nil {
		return nil, errors.New("客户不存在")
	}

	// 检查客户是否已有账户
	existingAccounts, _, _ := s.accountRepo.GetByClientID(ctx, req.ClientID)
	for _, acc := range existingAccounts {
		if acc.Status != "closed" {
			return nil, errors.New("该客户已有有效账户")
		}
	}

	// 生成账户编号
	accountCode := fmt.Sprintf("TRUST-%d%s", req.ClientID, time.Now().Format("20060102"))

	now := time.Now()
	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	account := &models.ClientTrustAccount{
		ClientID:           req.ClientID,
		AccountCode:        accountCode,
		Balance:            0,
		Currency:           currency,
		FrozenAmount:       0,
		PurposeRestriction: req.PurposeRestriction,
		AuthorizedUses:     models.JSON{"uses": req.AuthorizedUses},
		Status:             "active",
		OpenedAt:           &now,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("创建账户失败: %w", err)
	}

	return s.GetAccountByID(ctx, account.ID)
}

// GetAccountByID 根据ID获取账户详情
func (s *TrustAccountService) GetAccountByID(ctx context.Context, id uint) (*AccountResponse, error) {
	account, err := s.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}
	if account == nil {
		return nil, errors.New("账户不存在")
	}

	return s.convertToResponse(ctx, account), nil
}

// ListAccounts 获取账户列表
func (s *TrustAccountService) ListAccounts(ctx context.Context, req *ListAccountsRequest) (*ListAccountsResponse, error) {
	params := &repositories.TrustAccountListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		ClientID: req.ClientID,
		Status:   req.Status,
		Currency: req.Currency,
	}

	accounts, total, err := s.accountRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询账户列表失败: %w", err)
	}

	response := &ListAccountsResponse{
		Accounts: make([]*AccountResponse, len(accounts)),
		Pagination: Pagination{
			Page:    req.Page,
			PageSize: req.PageSize,
			Total:   total,
		},
	}

	for i, a := range accounts {
		response.Accounts[i] = s.convertToResponse(ctx, a)
	}

	return response, nil
}

// FreezeAccount 冻结账户
func (s *TrustAccountService) FreezeAccount(ctx context.Context, id uint) (*AccountResponse, error) {
	account, err := s.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}
	if account == nil {
		return nil, errors.New("账户不存在")
	}

	if account.Status == "closed" {
		return nil, errors.New("已关闭的账户不能冻结")
	}

	account.Status = "frozen"

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("冻结账户失败: %w", err)
	}

	return s.GetAccountByID(ctx, id)
}

// UnfreezeAccount 解冻账户
func (s *TrustAccountService) UnfreezeAccount(ctx context.Context, id uint) (*AccountResponse, error) {
	account, err := s.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}
	if account == nil {
		return nil, errors.New("账户不存在")
	}

	if account.Status != "frozen" {
		return nil, errors.New("只有冻结状态的账户可以解冻")
	}

	account.Status = "active"

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("解冻账户失败: %w", err)
	}

	return s.GetAccountByID(ctx, id)
}

// CloseAccount 关闭账户
func (s *TrustAccountService) CloseAccount(ctx context.Context, id uint) (*AccountResponse, error) {
	account, err := s.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}
	if account == nil {
		return nil, errors.New("账户不存在")
	}

	if account.Status == "closed" {
		return nil, errors.New("账户已关闭")
	}

	// 检查余额
	if account.Balance > 0 {
		return nil, errors.New("账户有余额，请先处理余额后再关闭")
	}

	// 检查是否有未完成的交易
	pendingTransactions, _, _ := s.transactionRepo.GetByAccountID(ctx, id)
	for _, t := range pendingTransactions {
		if t.Status == "pending" {
			return nil, errors.New("账户有未完成的交易，请先处理后再关闭")
		}
	}

	now := time.Now()
	account.Status = "closed"
	account.ClosedAt = &now

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("关闭账户失败: %w", err)
	}

	return s.GetAccountByID(ctx, id)
}

// convertToResponse 转换为响应格式
func (s *TrustAccountService) convertToResponse(ctx context.Context, account *models.ClientTrustAccount) *AccountResponse {
	resp := &AccountResponse{
		ID:               account.ID,
		ClientID:         account.ClientID,
		AccountCode:      account.AccountCode,
		Balance:          account.Balance,
		Currency:         account.Currency,
		FrozenAmount:     account.FrozenAmount,
		AvailableBalance: account.Balance - account.FrozenAmount,
		PurposeRestriction: account.PurposeRestriction,
		Status:           account.Status,
		CreatedAt:        account.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        account.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	// 解析授权用途
	if account.AuthorizedUses != nil {
		if useList, ok := account.AuthorizedUses["uses"].([]interface{}); ok {
			for _, u := range useList {
				if str, ok := u.(string); ok {
					resp.AuthorizedUses = append(resp.AuthorizedUses, str)
				}
			}
		}
	}

	if account.OpenedAt != nil {
		formatted := account.OpenedAt.Format("2006-01-02")
		resp.OpenedAt = &formatted
	}
	if account.ClosedAt != nil {
		formatted := account.ClosedAt.Format("2006-01-02")
		resp.ClosedAt = &formatted
	}

	// 获取最近交易
	transactions, _, _ := s.transactionRepo.GetByAccountID(ctx, account.ID)
	if len(transactions) > 0 {
		resp.RecentTransactions = make([]*TransactionSummary, 0, min(len(transactions), 5))
		for i := 0; i < min(len(transactions), 5); i++ {
			t := transactions[i]
			resp.RecentTransactions = append(resp.RecentTransactions, &TransactionSummary{
				ID:              t.ID,
				TransactionCode: t.TransactionCode,
				TransactionType: t.TransactionType,
				Amount:          t.Amount,
				Description:     t.Description,
				Status:          t.Status,
				CreatedAt:       t.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	return resp
}

// ===================================
// 代管款交易服务
// ===================================

// TrustTransactionService 代管款交易服务
type TrustTransactionService struct {
	transactionRepo repositories.TrustTransactionRepository
	accountRepo     repositories.TrustAccountRepository
	caseRepo        repositories.CaseRepository
	userRepo        repositories.UserRepository
}

// NewTrustTransactionService 创建代管款交易服务实例
func NewTrustTransactionService(
	transactionRepo repositories.TrustTransactionRepository,
	accountRepo repositories.TrustAccountRepository,
	caseRepo repositories.CaseRepository,
	userRepo repositories.UserRepository,
) *TrustTransactionService {
	return &TrustTransactionService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		caseRepo:        caseRepo,
		userRepo:        userRepo,
	}
}

// CreateTransactionRequest 创建交易请求
type CreateTransactionRequest struct {
	AccountID           uint    `json:"account_id" binding:"required"`
	TransactionType     string  `json:"transaction_type" binding:"required,oneof=deposit deposit_refund withdraw transfer"`
	Amount              float64 `json:"amount" binding:"required,gt=0"`
	Description         string  `json:"description" binding:"required"`
	CaseID              *uint   `json:"case_id,omitempty"`
	PurposeCode         string  `json:"purpose_code" binding:"max=50"`
	RecipientName       string  `json:"recipient_name" binding:"max=200"`
	RecipientBankAccount string `json:"recipient_bank_account" binding:"max=50"`
	RecipientBankName   string  `json:"recipient_bank_name" binding:"max=100"`
	AttachmentID        *uint   `json:"attachment_id,omitempty"`
}

// TransactionResponse 交易响应
type TransactionResponse struct {
	ID                  uint              `json:"id"`
	AccountID           uint              `json:"account_id"`
	TransactionCode     string            `json:"transaction_code"`
	TransactionType     string            `json:"transaction_type"`
	Amount              float64           `json:"amount"`
	Description         string            `json:"description"`
	CaseID              *uint             `json:"case_id,omitempty"`
	PurposeCode         string            `json:"purpose_code"`
	RecipientName       string            `json:"recipient_name"`
	RecipientBankAccount string           `json:"recipient_bank_account"`
	RecipientBankName   string            `json:"recipient_bank_name"`
	Status              string            `json:"status"`
	CompletedAt         *string           `json:"completed_at,omitempty"`
	AttachmentID        *uint             `json:"attachment_id,omitempty"`
	CreatedBy           uint              `json:"created_by"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
	ApprovedBy          *uint             `json:"approved_by,omitempty"`
	ApprovedAt          *string           `json:"approved_at,omitempty"`
	// 关联数据
	Account             *AccountSummary   `json:"account,omitempty"`
	Case                *CaseSummary      `json:"case,omitempty"`
}

// AccountSummary 账户摘要
type AccountSummary struct {
	ID          uint    `json:"id"`
	AccountCode string  `json:"account_code"`
	Balance     float64 `json:"balance"`
}

// CreateTransaction 创建交易
func (s *TrustTransactionService) CreateTransaction(ctx context.Context, req *CreateTransactionRequest, createdBy uint) (*TransactionResponse, error) {
	// 验证账户是否存在
	account, err := s.accountRepo.FindByID(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}
	if account == nil {
		return nil, errors.New("账户不存在")
	}

	if account.Status != "active" {
		return nil, errors.New("账户状态不正确，无法进行交易")
	}

	// 验证金额（对于取款和转账，检查余额）
	if req.TransactionType == "withdraw" || req.TransactionType == "transfer" {
		if req.Amount > account.Balance-account.FrozenAmount {
			return nil, errors.New("可用余额不足")
		}
	}

	// 生成交易编号
	transactionCode := fmt.Sprintf("TXN-%d%s", req.AccountID, time.Now().Format("20060102150405"))

	transaction := &models.ClientTrustTransaction{
		AccountID:           req.AccountID,
		TransactionCode:     transactionCode,
		TransactionType:     req.TransactionType,
		Amount:              req.Amount,
		Description:         req.Description,
		CaseID:              req.CaseID,
		PurposeCode:         req.PurposeCode,
		RecipientName:       req.RecipientName,
		RecipientBankAccount: req.RecipientBankAccount,
		RecipientBankName:   req.RecipientBankName,
		Status:              "pending",
		CreatedBy:           createdBy,
		AttachmentID:        req.AttachmentID,
	}

	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("创建交易失败: %w", err)
	}

	return s.GetTransactionByID(ctx, transaction.ID)
}

// GetTransactionByID 根据ID获取交易详情
func (s *TrustTransactionService) GetTransactionByID(ctx context.Context, id uint) (*TransactionResponse, error) {
	transaction, err := s.transactionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询交易失败: %w", err)
	}
	if transaction == nil {
		return nil, errors.New("交易不存在")
	}

	return s.convertToResponse(transaction), nil
}

// ApproveTransaction 审批通过交易
func (s *TrustTransactionService) ApproveTransaction(ctx context.Context, id uint, approvedBy uint) (*TransactionResponse, error) {
	transaction, err := s.transactionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询交易失败: %w", err)
	}
	if transaction == nil {
		return nil, errors.New("交易不存在")
	}

	if transaction.Status != "pending" {
		return nil, errors.New("只有待审批状态的交易可以审批")
	}

	// 获取账户
	account, err := s.accountRepo.FindByID(ctx, transaction.AccountID)
	if err != nil {
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}

	// 执行交易
	now := time.Now()

	switch transaction.TransactionType {
	case "deposit":
		// 存入：增加余额
		account.Balance += transaction.Amount
	case "deposit_refund", "withdraw", "transfer":
		// 取出：减少余额
		if transaction.Amount > account.Balance-account.FrozenAmount {
			return nil, errors.New("可用余额不足")
		}
		account.Balance -= transaction.Amount
	}

	transaction.Status = "completed"
	transaction.CompletedAt = &now
	transaction.ApprovedBy = &approvedBy
	transaction.ApprovedAt = &now

	// 更新账户余额
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("更新账户余额失败: %w", err)
	}

	// 更新交易状态
	if err := s.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("审批交易失败: %w", err)
	}

	return s.GetTransactionByID(ctx, id)
}

// RejectTransaction 审批拒绝交易
func (s *TrustTransactionService) RejectTransaction(ctx context.Context, id uint) (*TransactionResponse, error) {
	transaction, err := s.transactionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询交易失败: %w", err)
	}
	if transaction == nil {
		return nil, errors.New("交易不存在")
	}

	if transaction.Status != "pending" {
		return nil, errors.New("只有待审批状态的交易可以拒绝")
	}

	transaction.Status = "cancelled"

	if err := s.transactionRepo.Update(ctx, transaction); err != nil {
		return nil, fmt.Errorf("拒绝交易失败: %w", err)
	}

	return s.GetTransactionByID(ctx, id)
}

// ListTransactions 获取交易列表
func (s *TrustTransactionService) ListTransactions(ctx context.Context, page, pageSize int, accountID *uint) ([]*TransactionResponse, int64, error) {
	params := &repositories.TrustTransactionListParams{
		Page:     page,
		PageSize: pageSize,
		AccountID: accountID,
	}

	transactions, total, err := s.transactionRepo.List(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("查询交易列表失败: %w", err)
	}

	response := make([]*TransactionResponse, len(transactions))
	for i, t := range transactions {
		response[i] = s.convertToResponse(t)
	}

	return response, total, nil
}

// convertToResponse 转换为响应格式
func (s *TrustTransactionService) convertToResponse(transaction *models.ClientTrustTransaction) *TransactionResponse {
	resp := &TransactionResponse{
		ID:                  transaction.ID,
		AccountID:           transaction.AccountID,
		TransactionCode:     transaction.TransactionCode,
		TransactionType:     transaction.TransactionType,
		Amount:              transaction.Amount,
		Description:         transaction.Description,
		CaseID:              transaction.CaseID,
		PurposeCode:         transaction.PurposeCode,
		RecipientName:       transaction.RecipientName,
		RecipientBankAccount: transaction.RecipientBankAccount,
		RecipientBankName:   transaction.RecipientBankName,
		Status:              transaction.Status,
		AttachmentID:        transaction.AttachmentID,
		CreatedBy:           transaction.CreatedBy,
		CreatedAt:           transaction.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           transaction.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if transaction.CompletedAt != nil {
		formatted := transaction.CompletedAt.Format("2006-01-02 15:04:05")
		resp.CompletedAt = &formatted
	}
	if transaction.ApprovedBy != nil {
		resp.ApprovedBy = transaction.ApprovedBy
	}
	if transaction.ApprovedAt != nil {
		formatted := transaction.ApprovedAt.Format("2006-01-02 15:04:05")
		resp.ApprovedAt = &formatted
	}

	return resp
}
