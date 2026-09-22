package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// InvoiceService 发票服务
type InvoiceService struct {
	invoiceRepo   repositories.InvoiceRepository
	contractRepo  repositories.ContractRepository
	milestoneRepo repositories.PaymentMilestoneRepository
	clientRepo    repositories.ClientRepository
	paymentRepo   repositories.PaymentRepository
}

// NewInvoiceService 创建发票服务实例
func NewInvoiceService(
	invoiceRepo repositories.InvoiceRepository,
	contractRepo repositories.ContractRepository,
	milestoneRepo repositories.PaymentMilestoneRepository,
	clientRepo repositories.ClientRepository,
	paymentRepo repositories.PaymentRepository,
) *InvoiceService {
	return &InvoiceService{
		invoiceRepo:   invoiceRepo,
		contractRepo:  contractRepo,
		milestoneRepo: milestoneRepo,
		clientRepo:    clientRepo,
		paymentRepo:   paymentRepo,
	}
}

// CreateInvoiceRequest 创建发票请求
type CreateInvoiceRequest struct {
	InvoiceCode string  `json:"invoice_code" binding:"required,min=1,max=50"`
	ContractID  *uint   `json:"contract_id,omitempty"`
	MilestoneID *uint   `json:"milestone_id,omitempty"`
	ClientID    uint    `json:"client_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	TaxRate     float64 `json:"tax_rate" binding:"required,gte=0,lte=100"`
	InvoiceType string  `json:"invoice_type" binding:"required,oneof=normal credit"`
	// 客户开票信息
	ClientName        string `json:"client_name" binding:"required,min=1,max=200"`
	ClientTaxID       string `json:"client_tax_id" binding:"max=50"`
	ClientAddress     string `json:"client_address" binding:"max=500"`
	ClientBankName    string `json:"client_bank_name" binding:"max=100"`
	ClientBankAccount string `json:"client_bank_account" binding:"max=50"`
	// 红冲信息
	OriginalInvoiceID *uint  `json:"original_invoice_id,omitempty"`
	RefundReason      string `json:"refund_reason,omitempty" binding:"max=500"`
}

// UpdateInvoiceRequest 更新发票请求
type UpdateInvoiceRequest struct {
	Amount            *float64 `json:"amount,omitempty" binding:"omitempty,gt=0"`
	TaxRate           *float64 `json:"tax_rate,omitempty" binding:"omitempty,gte=0,lte=100"`
	ClientName        *string  `json:"client_name,omitempty" binding:"omitempty,min=1,max=200"`
	ClientTaxID       *string  `json:"client_tax_id,omitempty" binding:"omitempty,max=50"`
	ClientAddress     *string  `json:"client_address,omitempty" binding:"omitempty,max=500"`
	ClientBankName    *string  `json:"client_bank_name,omitempty" binding:"omitempty,max=100"`
	ClientBankAccount *string  `json:"client_bank_account,omitempty" binding:"omitempty,max=50"`
}

// InvoiceResponse 发票响应
type InvoiceResponse struct {
	ID          uint    `json:"id"`
	InvoiceCode string  `json:"invoice_code"`
	ContractID  *uint   `json:"contract_id,omitempty"`
	MilestoneID *uint   `json:"milestone_id,omitempty"`
	ClientID    uint    `json:"client_id"`
	Amount      float64 `json:"amount"`
	TaxRate     float64 `json:"tax_rate"`
	TaxAmount   float64 `json:"tax_amount"`
	TotalAmount float64 `json:"total_amount"`
	// 客户开票信息
	ClientName        string `json:"client_name"`
	ClientTaxID       string `json:"client_tax_id"`
	ClientAddress     string `json:"client_address"`
	ClientBankName    string `json:"client_bank_name"`
	ClientBankAccount string `json:"client_bank_account"`
	// 发票类型
	InvoiceType       string  `json:"invoice_type"`
	OriginalInvoiceID *uint   `json:"original_invoice_id,omitempty"`
	RefundReason      string  `json:"refund_reason,omitempty"`
	WriteOffAmount    float64 `json:"write_off_amount"`
	// 状态
	Status              string  `json:"status"`
	SubmittedAt         *string `json:"submitted_at,omitempty"`
	ApprovedByFinanceAt *string `json:"approved_by_finance_at,omitempty"`
	IssuedAt            *string `json:"issued_at,omitempty"`
	ReceivedAt          *string `json:"received_at,omitempty"`
	// 电子发票
	ElectronicInvoiceURL    string `json:"electronic_invoice_url"`
	ElectronicInvoiceCode   string `json:"electronic_invoice_code"`
	ElectronicInvoiceNumber string `json:"electronic_invoice_number"`
	// 审批信息
	CreatedBy   uint   `json:"created_by"`
	SubmittedBy *uint  `json:"submitted_by,omitempty"`
	ApprovedBy  *uint  `json:"approved_by,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	// 关联数据
	Client          *ClientSummary    `json:"client,omitempty"`
	Contract        *ContractSummary  `json:"contract,omitempty"`
	Milestone       *MilestoneSummary `json:"milestone,omitempty"`
	Payments        []*PaymentSummary `json:"payments,omitempty"`
	TotalPaidAmount float64           `json:"total_paid_amount"`
	RemainingAmount float64           `json:"remaining_amount"`
}

// ContractSummary 合同摘要
type ContractSummary struct {
	ID             uint    `json:"id"`
	ContractCode   string  `json:"contract_code"`
	ContractAmount float64 `json:"contract_amount"`
}

// MilestoneSummary 付款计划摘要
type MilestoneSummary struct {
	ID     uint    `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

// PaymentSummary 回款摘要
type PaymentSummary struct {
	ID          uint    `json:"id"`
	PaymentCode string  `json:"payment_code"`
	Amount      float64 `json:"amount"`
	PaymentDate string  `json:"payment_date"`
	Status      string  `json:"status"`
}

// ListInvoicesRequest 发票列表请求
type ListInvoicesRequest struct {
	Page        int    `json:"page" form:"page" binding:"min=1"`
	PageSize    int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status      string `json:"status" form:"status" binding:"omitempty,oneof=draft submitted approved issued received cancelled"`
	InvoiceType string `json:"invoice_type" form:"invoice_type" binding:"omitempty,oneof=normal credit"`
	ClientID    uint   `json:"client_id" form:"client_id"`
	ContractID  uint   `json:"contract_id" form:"contract_id"`
	Search      string `json:"search" form:"search"`
	DateFrom    string `json:"date_from" form:"date_from"`
	DateTo      string `json:"date_to" form:"date_to"`
}

// ListInvoicesResponse 发票列表响应
type ListInvoicesResponse struct {
	Invoices   []*InvoiceResponse `json:"invoices"`
	Pagination Pagination         `json:"pagination"`
}

// CreateInvoice 创建发票
func (s *InvoiceService) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest, createdBy uint) (*InvoiceResponse, error) {
	// 验证客户是否存在
	client, err := s.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("查询客户失败: %w", err)
	}
	if client == nil {
		return nil, errors.New("客户不存在")
	}

	// 验证发票编号是否已存在
	existingInvoice, err := s.invoiceRepo.FindByCode(ctx, req.InvoiceCode)
	if err != nil {
		return nil, fmt.Errorf("查询发票编号失败: %w", err)
	}
	if existingInvoice != nil {
		return nil, errors.New("发票编号已存在")
	}

	// 如果是红字发票，验证原发票是否存在
	if req.InvoiceType == "credit" {
		if req.OriginalInvoiceID == nil {
			return nil, errors.New("红字发票必须指定原发票")
		}
		originalInvoice, err := s.invoiceRepo.FindByID(ctx, *req.OriginalInvoiceID)
		if err != nil {
			return nil, fmt.Errorf("查询原发票失败: %w", err)
		}
		if originalInvoice == nil {
			return nil, errors.New("原发票不存在")
		}
		if originalInvoice.InvoiceType == "credit" {
			return nil, errors.New("不能对红字发票进行红冲")
		}

		// 红冲金额不能超过原发票金额
		totalAmount := req.Amount * (1 + req.TaxRate/100)
		if totalAmount > originalInvoice.TotalAmount {
			return nil, fmt.Errorf("红冲金额(%.2f)不能超过原发票金额(%.2f)", totalAmount, originalInvoice.TotalAmount)
		}
	}

	// 如果关联付款计划，验证是否存在
	if req.MilestoneID != nil {
		milestone, err := s.milestoneRepo.FindByID(ctx, *req.MilestoneID)
		if err != nil {
			return nil, fmt.Errorf("查询付款计划失败: %w", err)
		}
		if milestone == nil {
			return nil, errors.New("付款计划不存在")
		}
		if milestone.Status == "paid" {
			return nil, errors.New("该付款计划已付清，无需开票")
		}
	}

	// 红字发票以负数金额入账，保持应收、税额和价税合计的方向一致。
	amount := req.Amount
	if req.InvoiceType == "credit" {
		amount = -math.Abs(amount)
	}
	taxAmount := amount * (req.TaxRate / 100)
	totalAmount := amount + taxAmount

	invoice := &models.Invoice{
		InvoiceCode: req.InvoiceCode,
		ContractID:  req.ContractID,
		MilestoneID: req.MilestoneID,
		ClientID:    req.ClientID,
		Amount:      amount,
		TaxRate:     req.TaxRate,
		TaxAmount:   taxAmount,
		TotalAmount: totalAmount,
		// 客户开票信息
		ClientName:        req.ClientName,
		ClientTaxID:       req.ClientTaxID,
		ClientAddress:     req.ClientAddress,
		ClientBankName:    req.ClientBankName,
		ClientBankAccount: req.ClientBankAccount,
		// 发票类型
		InvoiceType:       req.InvoiceType,
		OriginalInvoiceID: req.OriginalInvoiceID,
		RefundReason:      req.RefundReason,
		// 状态
		Status:    "draft",
		CreatedBy: createdBy,
	}

	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, fmt.Errorf("创建发票失败: %w", err)
	}

	return s.GetInvoiceByID(ctx, invoice.ID)
}

// GetInvoiceByID 根据ID获取发票详情
func (s *InvoiceService) GetInvoiceByID(ctx context.Context, id uint) (*InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	return s.convertToResponse(ctx, invoice), nil
}

// UpdateInvoice 更新发票
func (s *InvoiceService) UpdateInvoice(ctx context.Context, id uint, req *UpdateInvoiceRequest) (*InvoiceResponse, error) {
	// 获取现有发票
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	// 只有草稿状态的发票可以编辑
	if invoice.Status != "draft" {
		return nil, errors.New("只有草稿状态的发票可以编辑")
	}

	// 更新字段
	needRecalculate := false
	if req.Amount != nil {
		invoice.Amount = *req.Amount
		needRecalculate = true
	}
	if req.TaxRate != nil {
		invoice.TaxRate = *req.TaxRate
		needRecalculate = true
	}
	if req.ClientName != nil {
		invoice.ClientName = *req.ClientName
	}
	if req.ClientTaxID != nil {
		invoice.ClientTaxID = *req.ClientTaxID
	}
	if req.ClientAddress != nil {
		invoice.ClientAddress = *req.ClientAddress
	}
	if req.ClientBankName != nil {
		invoice.ClientBankName = *req.ClientBankName
	}
	if req.ClientBankAccount != nil {
		invoice.ClientBankAccount = *req.ClientBankAccount
	}

	// 重新计算税额和价税合计
	if needRecalculate {
		invoice.TaxAmount = invoice.Amount * (invoice.TaxRate / 100)
		invoice.TotalAmount = invoice.Amount + invoice.TaxAmount
	}

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("更新发票失败: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// DeleteInvoice 删除发票
func (s *InvoiceService) DeleteInvoice(ctx context.Context, id uint) error {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return errors.New("发票不存在")
	}

	// 只有草稿状态的发票可以删除
	if invoice.Status != "draft" {
		return errors.New("只有草稿状态的发票可以删除")
	}

	if err := s.invoiceRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除发票失败: %w", err)
	}

	return nil
}

// ListInvoices 获取发票列表
func (s *InvoiceService) ListInvoices(ctx context.Context, req *ListInvoicesRequest) (*ListInvoicesResponse, error) {
	params := &repositories.InvoiceListParams{
		Page:        req.Page,
		PageSize:    req.PageSize,
		Status:      req.Status,
		InvoiceType: req.InvoiceType,
		ClientID:    req.ClientID,
		ContractID:  req.ContractID,
		Search:      req.Search,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
	}

	invoices, total, err := s.invoiceRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询发票列表失败: %w", err)
	}

	response := &ListInvoicesResponse{
		Invoices: make([]*InvoiceResponse, len(invoices)),
		Pagination: Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}

	for i, invoice := range invoices {
		response.Invoices[i] = s.convertToResponse(ctx, invoice)
	}

	return response, nil
}

// SubmitInvoice 提交发票审批
func (s *InvoiceService) SubmitInvoice(ctx context.Context, id uint, submittedBy uint) (*InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	if invoice.Status != "draft" {
		return nil, errors.New("只有草稿状态的发票可以提交审批")
	}

	now := time.Now()
	invoice.Status = "submitted"
	invoice.SubmittedAt = &now
	invoice.SubmittedBy = &submittedBy

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("提交发票失败: %w", err)
	}

	// 如果关联付款计划，更新状态为billed
	if invoice.MilestoneID != nil {
		s.milestoneRepo.UpdateStatus(ctx, *invoice.MilestoneID, "billed")
	}

	return s.GetInvoiceByID(ctx, id)
}

// ApproveInvoice 财务审批通过
func (s *InvoiceService) ApproveInvoice(ctx context.Context, id uint, approvedBy uint) (*InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	if invoice.Status != "submitted" {
		return nil, errors.New("只有已提交状态的发票可以审批")
	}

	now := time.Now()
	invoice.Status = "approved"
	invoice.ApprovedByFinanceAt = &now
	invoice.ApprovedBy = &approvedBy

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("审批发票失败: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// RejectInvoice 财务审批拒绝
func (s *InvoiceService) RejectInvoice(ctx context.Context, id uint) (*InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	if invoice.Status != "submitted" {
		return nil, errors.New("只有已提交状态的发票可以拒绝")
	}

	invoice.Status = "draft"
	invoice.SubmittedAt = nil
	invoice.SubmittedBy = nil

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("拒绝发票失败: %w", err)
	}

	// 如果关联付款计划，恢复状态
	if invoice.MilestoneID != nil {
		s.milestoneRepo.UpdateStatus(ctx, *invoice.MilestoneID, "pending")
	}

	return s.GetInvoiceByID(ctx, id)
}

// IssueInvoice 开票
func (s *InvoiceService) IssueInvoice(ctx context.Context, id uint, electronicURL, code, number string) (*InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	if invoice.Status != "approved" {
		return nil, errors.New("只有审批通过的发票可以开票")
	}

	now := time.Now()
	invoice.Status = "issued"
	invoice.IssuedAt = &now
	invoice.ElectronicInvoiceURL = electronicURL
	invoice.ElectronicInvoiceCode = code
	invoice.ElectronicInvoiceNumber = number

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("开票失败: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// ConfirmReceipt 客户签收
func (s *InvoiceService) ConfirmReceipt(ctx context.Context, id uint) (*InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	if invoice.Status != "issued" {
		return nil, errors.New("只有已开票状态的发票可以签收")
	}

	now := time.Now()
	invoice.Status = "received"
	invoice.ReceivedAt = &now

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("签收失败: %w", err)
	}

	return s.GetInvoiceByID(ctx, id)
}

// CancelInvoice 作废发票
func (s *InvoiceService) CancelInvoice(ctx context.Context, id uint) (*InvoiceResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	if invoice.Status == "cancelled" {
		return nil, errors.New("发票已作废")
	}

	// 已开票的发票需要先红冲
	if invoice.Status == "issued" || invoice.Status == "received" {
		return nil, errors.New("已开票的发票需要先红冲才能作废")
	}

	invoice.Status = "cancelled"

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		return nil, fmt.Errorf("作废发票失败: %w", err)
	}

	// 如果关联付款计划，恢复状态
	if invoice.MilestoneID != nil {
		s.milestoneRepo.UpdateStatus(ctx, *invoice.MilestoneID, "pending")
	}

	return s.GetInvoiceByID(ctx, id)
}

// convertToResponse 转换为响应格式
func (s *InvoiceService) convertToResponse(ctx context.Context, invoice *models.Invoice) *InvoiceResponse {
	resp := &InvoiceResponse{
		ID:                      invoice.ID,
		InvoiceCode:             invoice.InvoiceCode,
		ContractID:              invoice.ContractID,
		MilestoneID:             invoice.MilestoneID,
		ClientID:                invoice.ClientID,
		Amount:                  invoice.Amount,
		TaxRate:                 invoice.TaxRate,
		TaxAmount:               invoice.TaxAmount,
		TotalAmount:             invoice.TotalAmount,
		ClientName:              invoice.ClientName,
		ClientTaxID:             invoice.ClientTaxID,
		ClientAddress:           invoice.ClientAddress,
		ClientBankName:          invoice.ClientBankName,
		ClientBankAccount:       invoice.ClientBankAccount,
		InvoiceType:             invoice.InvoiceType,
		OriginalInvoiceID:       invoice.OriginalInvoiceID,
		RefundReason:            invoice.RefundReason,
		WriteOffAmount:          invoice.WriteOffAmount,
		Status:                  invoice.Status,
		ElectronicInvoiceURL:    invoice.ElectronicInvoiceURL,
		ElectronicInvoiceCode:   invoice.ElectronicInvoiceCode,
		ElectronicInvoiceNumber: invoice.ElectronicInvoiceNumber,
		CreatedBy:               invoice.CreatedBy,
		CreatedAt:               invoice.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:               invoice.UpdatedAt.Format("2006-01-02 15:04:05"),
		RemainingAmount:         invoice.TotalAmount,
	}

	if invoice.SubmittedAt != nil {
		formatted := invoice.SubmittedAt.Format("2006-01-02 15:04:05")
		resp.SubmittedAt = &formatted
	}
	if invoice.ApprovedByFinanceAt != nil {
		formatted := invoice.ApprovedByFinanceAt.Format("2006-01-02 15:04:05")
		resp.ApprovedByFinanceAt = &formatted
	}
	if invoice.IssuedAt != nil {
		formatted := invoice.IssuedAt.Format("2006-01-02")
		resp.IssuedAt = &formatted
	}
	if invoice.ReceivedAt != nil {
		formatted := invoice.ReceivedAt.Format("2006-01-02")
		resp.ReceivedAt = &formatted
	}

	// 客户信息 - 通过仓储获取
	if client, err := s.clientRepo.FindByID(ctx, invoice.ClientID); err == nil && client != nil {
		resp.Client = &ClientSummary{
			ID:   client.ID,
			Name: client.Name,
		}
	}

	// 合同信息 - 通过仓储获取
	if invoice.ContractID != nil {
		if contract, err := s.contractRepo.FindByID(ctx, *invoice.ContractID); err == nil && contract != nil {
			resp.Contract = &ContractSummary{
				ID:             contract.ID,
				ContractCode:   contract.ContractCode,
				ContractAmount: contract.ContractAmount,
			}
		}
	}

	// 付款计划信息 - 通过仓储获取
	if invoice.MilestoneID != nil {
		milestones, err := s.milestoneRepo.GetByContractID(ctx, 0) // 使用空合同ID获取所有
		if err == nil {
			for _, milestone := range milestones {
				if milestone.ID == *invoice.MilestoneID {
					resp.Milestone = &MilestoneSummary{
						ID:     milestone.ID,
						Name:   milestone.Name,
						Amount: milestone.Amount,
					}
					break
				}
			}
		}
	}

	// 回款信息
	if invoice.ID > 0 {
		payments, _ := s.paymentRepo.GetByInvoiceID(ctx, invoice.ID)
		if len(payments) > 0 {
			resp.Payments = make([]*PaymentSummary, len(payments))
			for i, p := range payments {
				resp.Payments[i] = &PaymentSummary{
					ID:          p.ID,
					PaymentCode: p.PaymentCode,
					Amount:      p.Amount,
					PaymentDate: p.PaymentDate.Format("2006-01-02"),
					Status:      p.Status,
				}
				if p.Status == "confirmed" {
					resp.TotalPaidAmount += p.Amount
				}
			}
		}
	}

	resp.RemainingAmount = resp.TotalAmount - resp.TotalPaidAmount

	return resp
}

// GetInvoiceStats 获取发票统计信息
func (s *InvoiceService) GetInvoiceStats(ctx context.Context) (*InvoiceStats, error) {
	var stats InvoiceStats

	// 获取所有发票进行统计
	invoices, total, err := s.invoiceRepo.List(ctx, &repositories.InvoiceListParams{
		Page:     1,
		PageSize: 10000, // 获取所有记录进行统计
	})
	if err != nil {
		return nil, fmt.Errorf("查询发票列表失败: %w", err)
	}

	stats.TotalInvoices = total
	deadline := time.Now().AddDate(0, 0, -30)

	for _, invoice := range invoices {
		// 按状态统计
		switch invoice.Status {
		case "draft":
			stats.DraftInvoices++
		case "submitted":
			stats.SubmittedInvoices++
			stats.PendingInvoiceAmount += invoice.TotalAmount
		case "approved":
			stats.ApprovedInvoices++
			if invoice.InvoiceType == "normal" {
				stats.TotalInvoiceAmount += invoice.TotalAmount
			}
			if invoice.CreatedAt.Before(deadline) {
				stats.OverdueAmount += invoice.TotalAmount
			}
		case "issued":
			stats.IssuedInvoices++
			if invoice.InvoiceType == "normal" {
				stats.TotalInvoiceAmount += invoice.TotalAmount
			}
			if invoice.CreatedAt.Before(deadline) {
				stats.OverdueAmount += invoice.TotalAmount
			}
		case "received":
			stats.ReceivedInvoices++
			if invoice.InvoiceType == "normal" {
				stats.TotalInvoiceAmount += invoice.TotalAmount
			}
		}
	}

	return &stats, nil
}

// InvoiceStats 发票统计信息
type InvoiceStats struct {
	TotalInvoices        int64   `json:"total_invoices"`
	DraftInvoices        int64   `json:"draft_invoices"`
	SubmittedInvoices    int64   `json:"submitted_invoices"`
	ApprovedInvoices     int64   `json:"approved_invoices"`
	IssuedInvoices       int64   `json:"issued_invoices"`
	ReceivedInvoices     int64   `json:"received_invoices"`
	TotalInvoiceAmount   float64 `json:"total_invoice_amount"`
	PendingInvoiceAmount float64 `json:"pending_invoice_amount"`
	OverdueAmount        float64 `json:"overdue_amount"`
}
