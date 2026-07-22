package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// PaymentService 回款服务
type PaymentService struct {
	paymentRepo   repositories.PaymentRepository
	invoiceRepo   repositories.InvoiceRepository
	milestoneRepo repositories.PaymentMilestoneRepository
	userRepo      repositories.UserRepository
}

// NewPaymentService 创建回款服务实例
func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	invoiceRepo repositories.InvoiceRepository,
	milestoneRepo repositories.PaymentMilestoneRepository,
	userRepo repositories.UserRepository,
) *PaymentService {
	return &PaymentService{
		paymentRepo:   paymentRepo,
		invoiceRepo:   invoiceRepo,
		milestoneRepo: milestoneRepo,
		userRepo:      userRepo,
	}
}

// CreatePaymentRequest 创建回款请求
type CreatePaymentRequest struct {
	InvoiceID     uint    `json:"invoice_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentDate   string  `json:"payment_date" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=bank_transfer cash other"`
	ReferenceNo   string  `json:"reference_no" binding:"max=100"`
	PayerName     string  `json:"payer_name" binding:"max=200"`
	PayerAccount  string  `json:"payer_account" binding:"max=100"`
	AttachmentID  *uint   `json:"attachment_id,omitempty"`
	Remark        string  `json:"remark" binding:"max=500"`
}

// UpdatePaymentRequest 更新回款请求
type UpdatePaymentRequest struct {
	Amount        *float64 `json:"amount,omitempty" binding:"omitempty,gt=0"`
	PaymentDate   *string  `json:"payment_date,omitempty"`
	PaymentMethod *string  `json:"payment_method,omitempty" binding:"omitempty,oneof=bank_transfer cash other"`
	ReferenceNo   *string  `json:"reference_no,omitempty" binding:"omitempty,max=100"`
	PayerName     *string  `json:"payer_name,omitempty" binding:"omitempty,max=200"`
	PayerAccount  *string  `json:"payer_account,omitempty" binding:"omitempty,max=100"`
	AttachmentID  *uint    `json:"attachment_id,omitempty"`
	Remark        *string  `json:"remark,omitempty" binding:"omitempty,max=500"`
}

// PaymentResponse 回款响应
type PaymentResponse struct {
	ID            uint    `json:"id"`
	PaymentCode   string  `json:"payment_code"`
	InvoiceID     uint    `json:"invoice_id"`
	Amount        float64 `json:"amount"`
	PaymentDate   string  `json:"payment_date"`
	PaymentMethod string  `json:"payment_method"`
	ReferenceNo   string  `json:"reference_no"`
	PayerName     string  `json:"payer_name"`
	PayerAccount  string  `json:"payer_account"`
	AttachmentID  *uint   `json:"attachment_id,omitempty"`
	ConfirmedBy   uint    `json:"confirmed_by"`
	ConfirmedAt   *string `json:"confirmed_at,omitempty"`
	Status        string  `json:"status"`
	Remark        string  `json:"remark"`
	CreatedAt     string  `json:"created_at"`
	// 关联数据
	Invoice *InvoiceSummary `json:"invoice,omitempty"`
}

// InvoiceSummary 发票摘要
type InvoiceSummary struct {
	ID          uint    `json:"id"`
	InvoiceCode string  `json:"invoice_code"`
	TotalAmount float64 `json:"total_amount"`
	ClientName  string  `json:"client_name"`
}

// ListPaymentsRequest 回款列表请求
type ListPaymentsRequest struct {
	Page      int    `json:"page" form:"page" binding:"min=1"`
	PageSize  int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status    string `json:"status" form:"status" binding:"omitempty,oneof=pending confirmed rejected"`
	InvoiceID uint   `json:"invoice_id" form:"invoice_id"`
	ClientID  uint   `json:"client_id" form:"client_id"`
	Search    string `json:"search" form:"search"`
	DateFrom  string `json:"date_from" form:"date_from"`
	DateTo    string `json:"date_to" form:"date_to"`
}

// ListPaymentsResponse 回款列表响应
type ListPaymentsResponse struct {
	Payments   []*PaymentResponse `json:"payments"`
	Pagination Pagination         `json:"pagination"`
}

// CreatePayment 创建回款记录
func (s *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*PaymentResponse, error) {
	// 验证发票是否存在
	invoice, err := s.invoiceRepo.FindByID(ctx, req.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	// 验证发票状态
	if invoice.Status != "issued" && invoice.Status != "received" {
		return nil, errors.New("只有已开票或已签收的发票可以登记回款")
	}

	// 验证红字发票
	if invoice.InvoiceType == "credit" {
		return nil, errors.New("红字发票不能登记回款")
	}

	// 解析付款日期
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		return nil, fmt.Errorf("付款日期格式错误，应为 YYYY-MM-DD: %w", err)
	}

	// 检查已回款金额
	totalPaid, _ := s.paymentRepo.GetTotalPaidAmount(ctx, req.InvoiceID)
	remainingAmount := invoice.TotalAmount - totalPaid
	if req.Amount > remainingAmount {
		return nil, fmt.Errorf("回款金额(%.2f)超过未回款金额(%.2f)", req.Amount, remainingAmount)
	}

	// 生成回款编号
	paymentCode := fmt.Sprintf("PAY-%d%s", time.Now().Unix(), generateRandomCode(4))

	payment := &models.Payment{
		PaymentCode:   paymentCode,
		InvoiceID:     req.InvoiceID,
		Amount:        req.Amount,
		PaymentDate:   paymentDate,
		PaymentMethod: req.PaymentMethod,
		ReferenceNo:   req.ReferenceNo,
		PayerName:     req.PayerName,
		PayerAccount:  req.PayerAccount,
		AttachmentID:  req.AttachmentID,
		Remark:        req.Remark,
		Status:        "pending",
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("创建回款记录失败: %w", err)
	}

	return s.GetPaymentByID(ctx, payment.ID)
}

// GetPaymentByID 根据ID获取回款详情
func (s *PaymentService) GetPaymentByID(ctx context.Context, id uint) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询回款记录失败: %w", err)
	}
	if payment == nil {
		return nil, errors.New("回款记录不存在")
	}

	return s.convertToResponse(ctx, payment), nil
}

// UpdatePayment 更新回款记录
func (s *PaymentService) UpdatePayment(ctx context.Context, id uint, req *UpdatePaymentRequest) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询回款记录失败: %w", err)
	}
	if payment == nil {
		return nil, errors.New("回款记录不存在")
	}

	// 只有待确认状态的回款可以编辑
	if payment.Status != "pending" {
		return nil, errors.New("只有待确认状态的回款可以编辑")
	}

	// 更新字段
	if req.Amount != nil {
		payment.Amount = *req.Amount
	}
	if req.PaymentDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.PaymentDate)
		if err != nil {
			return nil, fmt.Errorf("付款日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		payment.PaymentDate = parsedDate
	}
	if req.PaymentMethod != nil {
		payment.PaymentMethod = *req.PaymentMethod
	}
	if req.ReferenceNo != nil {
		payment.ReferenceNo = *req.ReferenceNo
	}
	if req.PayerName != nil {
		payment.PayerName = *req.PayerName
	}
	if req.PayerAccount != nil {
		payment.PayerAccount = *req.PayerAccount
	}
	if req.AttachmentID != nil {
		payment.AttachmentID = req.AttachmentID
	}
	if req.Remark != nil {
		payment.Remark = *req.Remark
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("更新回款记录失败: %w", err)
	}

	return s.GetPaymentByID(ctx, id)
}

// DeletePayment 删除回款记录
func (s *PaymentService) DeletePayment(ctx context.Context, id uint) error {
	payment, err := s.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询回款记录失败: %w", err)
	}
	if payment == nil {
		return errors.New("回款记录不存在")
	}

	// 只有待确认状态的回款可以删除
	if payment.Status != "pending" {
		return errors.New("只有待确认状态的回款可以删除")
	}

	if err := s.paymentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除回款记录失败: %w", err)
	}

	return nil
}

// ListPayments 获取回款列表
func (s *PaymentService) ListPayments(ctx context.Context, req *ListPaymentsRequest) (*ListPaymentsResponse, error) {
	params := &repositories.PaymentListParams{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Status:    req.Status,
		InvoiceID: req.InvoiceID,
		ClientID:  req.ClientID,
		Search:    req.Search,
		DateFrom:  req.DateFrom,
		DateTo:    req.DateTo,
	}

	payments, total, err := s.paymentRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询回款列表失败: %w", err)
	}

	response := &ListPaymentsResponse{
		Payments: make([]*PaymentResponse, len(payments)),
		Pagination: Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}

	for i, payment := range payments {
		response.Payments[i] = s.convertToResponse(ctx, payment)
	}

	return response, nil
}

// ConfirmPayment 确认回款
func (s *PaymentService) ConfirmPayment(ctx context.Context, id uint, confirmedBy uint) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询回款记录失败: %w", err)
	}
	if payment == nil {
		return nil, errors.New("回款记录不存在")
	}

	if payment.Status != "pending" {
		return nil, errors.New("只有待确认状态的回款可以确认")
	}

	now := time.Now()
	payment.Status = "confirmed"
	payment.ConfirmedAt = &now
	payment.ConfirmedBy = confirmedBy

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("确认回款失败: %w", err)
	}

	// 更新付款计划状态
	invoice, _ := s.invoiceRepo.FindByID(ctx, payment.InvoiceID)
	if invoice != nil && invoice.MilestoneID != nil {
		milestone, _ := s.milestoneRepo.FindByID(ctx, *invoice.MilestoneID)
		if milestone != nil {
			// 获取总回款金额
			totalPaid, _ := s.paymentRepo.GetTotalPaidAmount(ctx, payment.InvoiceID)
			milestone.PaidAmount = totalPaid

			if totalPaid >= milestone.Amount {
				milestone.Status = "paid"
			} else if totalPaid > 0 {
				milestone.Status = "partial_paid"
			}

			s.milestoneRepo.Update(ctx, milestone)
		}
	}

	return s.GetPaymentByID(ctx, id)
}

// RejectPayment 拒绝回款
func (s *PaymentService) RejectPayment(ctx context.Context, id uint) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询回款记录失败: %w", err)
	}
	if payment == nil {
		return nil, errors.New("回款记录不存在")
	}

	if payment.Status != "pending" {
		return nil, errors.New("只有待确认状态的回款可以拒绝")
	}

	payment.Status = "rejected"

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("拒绝回款失败: %w", err)
	}

	return s.GetPaymentByID(ctx, id)
}

// GetPaymentsByInvoiceID 获取发票的回款记录
func (s *PaymentService) GetPaymentsByInvoiceID(ctx context.Context, invoiceID uint) ([]*PaymentResponse, error) {
	payments, err := s.paymentRepo.GetByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("查询发票回款记录失败: %w", err)
	}

	result := make([]*PaymentResponse, len(payments))
	for i, p := range payments {
		result[i] = s.convertToResponse(ctx, p)
	}

	return result, nil
}

// convertToResponse 转换为响应格式
func (s *PaymentService) convertToResponse(ctx context.Context, payment *models.Payment) *PaymentResponse {
	resp := &PaymentResponse{
		ID:            payment.ID,
		PaymentCode:   payment.PaymentCode,
		InvoiceID:     payment.InvoiceID,
		Amount:        payment.Amount,
		PaymentDate:   payment.PaymentDate.Format("2006-01-02"),
		PaymentMethod: payment.PaymentMethod,
		ReferenceNo:   payment.ReferenceNo,
		PayerName:     payment.PayerName,
		PayerAccount:  payment.PayerAccount,
		AttachmentID:  payment.AttachmentID,
		ConfirmedBy:   payment.ConfirmedBy,
		Status:        payment.Status,
		Remark:        payment.Remark,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if payment.ConfirmedAt != nil {
		formatted := payment.ConfirmedAt.Format("2006-01-02 15:04:05")
		resp.ConfirmedAt = &formatted
	}

	// 发票信息 - 通过仓储获取
	if invoice, err := s.invoiceRepo.FindByID(ctx, payment.InvoiceID); err == nil && invoice != nil {
		resp.Invoice = &InvoiceSummary{
			ID:          invoice.ID,
			InvoiceCode: invoice.InvoiceCode,
			TotalAmount: invoice.TotalAmount,
			ClientName:  invoice.ClientName,
		}
	}

	return resp
}

// GetPaymentStats 获取回款统计信息
func (s *PaymentService) GetPaymentStats(ctx context.Context) (*PaymentStats, error) {
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	nextMonth := currentMonth.AddDate(0, 1, 0)

	total, pendingCount, confirmedCount, rejectedCount, totalAmount, monthAmount, pendingAmount, err :=
		s.paymentRepo.GetPaymentAggregation(ctx, currentMonth.Format("2006-01-02"), nextMonth.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("查询回款统计失败: %w", err)
	}

	return &PaymentStats{
		TotalPayments:      total,
		PendingPayments:    pendingCount,
		ConfirmedPayments:  confirmedCount,
		RejectedPayments:   rejectedCount,
		TotalPaymentAmount: totalAmount,
		MonthPaymentAmount: monthAmount,
		PendingAmount:      pendingAmount,
	}, nil
}

// PaymentStats 回款统计信息
type PaymentStats struct {
	TotalPayments      int64   `json:"total_payments"`
	PendingPayments    int64   `json:"pending_payments"`
	ConfirmedPayments  int64   `json:"confirmed_payments"`
	RejectedPayments   int64   `json:"rejected_payments"`
	TotalPaymentAmount float64 `json:"total_payment_amount"`
	MonthPaymentAmount float64 `json:"month_payment_amount"`
	PendingAmount      float64 `json:"pending_amount"`
}

// generateRandomCode 生成随机码
func generateRandomCode(length int) string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
