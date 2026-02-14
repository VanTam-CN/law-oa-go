package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// BadDebtService 坏账核销服务
type BadDebtService struct {
	badDebtRepo repositories.BadDebtRepository
	contractRepo repositories.ContractRepository
	invoiceRepo repositories.InvoiceRepository
	paymentRepo repositories.PaymentRepository
	userRepo    repositories.UserRepository
}

// NewBadDebtService 创建坏账核销服务实例
func NewBadDebtService(
	badDebtRepo repositories.BadDebtRepository,
	contractRepo repositories.ContractRepository,
	invoiceRepo repositories.InvoiceRepository,
	paymentRepo repositories.PaymentRepository,
	userRepo repositories.UserRepository,
) *BadDebtService {
	return &BadDebtService{
		badDebtRepo:  badDebtRepo,
		contractRepo: contractRepo,
		invoiceRepo:  invoiceRepo,
		paymentRepo:  paymentRepo,
		userRepo:     userRepo,
	}
}

// CreateBadDebtRequest 创建坏账核销请求
type CreateBadDebtRequest struct {
	ContractID  uint     `json:"contract_id" binding:"required"`
	InvoiceID   *uint    `json:"invoice_id,omitempty"`
	WriteOffAmount float64 `json:"write_off_amount" binding:"required,gt=0"`
	Reason      string   `json:"reason" binding:"required,min=1,max=1000"`
	ReasonType  string   `json:"reason_type" binding:"required,oneof=bankruptcy dispute uncollectible other"`
	AttachmentIDs []uint `json:"attachment_ids,omitempty"`
}

// BadDebtResponse 坏账核销响应
type BadDebtResponse struct {
	ID              uint             `json:"id"`
	ContractID      uint             `json:"contract_id"`
	InvoiceID       *uint            `json:"invoice_id,omitempty"`
	OriginalAmount  float64          `json:"original_amount"`
	WriteOffAmount  float64          `json:"write_off_amount"`
	RemainingAmount float64          `json:"remaining_amount"`
	Reason          string           `json:"reason"`
	ReasonType      string           `json:"reason_type"`
	Status          string           `json:"status"`
	ApprovedBy      *uint            `json:"approved_by,omitempty"`
	ApprovedAt      *string          `json:"approved_at,omitempty"`
	ApprovalNotes   string           `json:"approval_notes"`
	AttachmentIDs   []uint           `json:"attachment_ids"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
	// 关联数据
	Contract        *ContractSummary `json:"contract,omitempty"`
	Invoice         *InvoiceSummary  `json:"invoice,omitempty"`
}

// ListBadDebtsRequest 坏账核销列表请求
type ListBadDebtsRequest struct {
	Page       int    `json:"page" form:"page" binding:"min=1"`
	PageSize   int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status     string `json:"status" form:"status" binding:"omitempty,oneof=pending approved rejected"`
	ContractID uint   `json:"contract_id" form:"contract_id"`
	ReasonType string `json:"reason_type" form:"reason_type" binding:"omitempty,oneof=bankruptcy dispute uncollectible other"`
}

// ListBadDebtsResponse 坏账核销列表响应
type ListBadDebtsResponse struct {
	BadDebts  []*BadDebtResponse `json:"bad_debts"`
	Pagination Pagination        `json:"pagination"`
}

// CreateBadDebt 创建坏账核销申请
func (s *BadDebtService) CreateBadDebt(ctx context.Context, req *CreateBadDebtRequest) (*BadDebtResponse, error) {
	// 验证合同是否存在
	contract, err := s.contractRepo.FindByID(ctx, req.ContractID)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	// 计算原应收金额和剩余金额
	var originalAmount, remainingAmount float64

	if req.InvoiceID != nil {
		// 验证发票是否存在
		invoice, err := s.invoiceRepo.FindByID(ctx, *req.InvoiceID)
		if err != nil {
			return nil, fmt.Errorf("查询发票失败: %w", err)
		}
		if invoice == nil {
			return nil, errors.New("发票不存在")
		}
		if invoice.ContractID != nil && *invoice.ContractID != req.ContractID {
			return nil, errors.New("发票不属于该合同")
		}

		originalAmount = invoice.TotalAmount
		totalPaid, _ := s.paymentRepo.GetTotalPaidAmount(ctx, *req.InvoiceID)
		remainingAmount = originalAmount - totalPaid
	} else {
		// 获取合同所有发票
		invoices, _ := s.invoiceRepo.GetByContractID(ctx, req.ContractID)
		for _, inv := range invoices {
			originalAmount += inv.TotalAmount
			totalPaid, _ := s.paymentRepo.GetTotalPaidAmount(ctx, inv.ID)
			remainingAmount += inv.TotalAmount - totalPaid
		}
	}

	// 验证核销金额
	if req.WriteOffAmount > remainingAmount {
		return nil, fmt.Errorf("核销金额(%.2f)不能超过剩余应收金额(%.2f)", req.WriteOffAmount, remainingAmount)
	}

	badDebt := &models.BadDebtRecord{
		ContractID:      req.ContractID,
		InvoiceID:       req.InvoiceID,
		OriginalAmount:  originalAmount,
		WriteOffAmount:  req.WriteOffAmount,
		RemainingAmount: remainingAmount - req.WriteOffAmount,
		Reason:          req.Reason,
		ReasonType:      req.ReasonType,
		Status:          "pending",
		AttachmentIDs:   models.JSON{"attachments": req.AttachmentIDs},
	}

	if err := s.badDebtRepo.Create(ctx, badDebt); err != nil {
		return nil, fmt.Errorf("创建坏账核销申请失败: %w", err)
	}

	return s.GetBadDebtByID(ctx, badDebt.ID)
}

// GetBadDebtByID 根据ID获取坏账核销详情
func (s *BadDebtService) GetBadDebtByID(ctx context.Context, id uint) (*BadDebtResponse, error) {
	badDebt, err := s.badDebtRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询坏账核销记录失败: %w", err)
	}
	if badDebt == nil {
		return nil, errors.New("坏账核销记录不存在")
	}

	return s.convertToResponse(badDebt), nil
}

// ListBadDebts 获取坏账核销列表
func (s *BadDebtService) ListBadDebts(ctx context.Context, req *ListBadDebtsRequest) (*ListBadDebtsResponse, error) {
	params := &repositories.BadDebtListParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Status:     req.Status,
		ContractID: req.ContractID,
		ReasonType: req.ReasonType,
	}

	badDebts, total, err := s.badDebtRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询坏账核销列表失败: %w", err)
	}

	response := &ListBadDebtsResponse{
		BadDebts: make([]*BadDebtResponse, len(badDebts)),
		Pagination: Pagination{
			Page:    req.Page,
			PageSize: req.PageSize,
			Total:   total,
		},
	}

	for i, bd := range badDebts {
		response.BadDebts[i] = s.convertToResponse(bd)
	}

	return response, nil
}

// ApproveBadDebt 审批通过坏账核销
func (s *BadDebtService) ApproveBadDebt(ctx context.Context, id uint, approvedBy uint, notes string) (*BadDebtResponse, error) {
	badDebt, err := s.badDebtRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询坏账核销记录失败: %w", err)
	}
	if badDebt == nil {
		return nil, errors.New("坏账核销记录不存在")
	}

	if badDebt.Status != "pending" {
		return nil, errors.New("只有待审批状态的坏账核销可以审批")
	}

	now := time.Now()
	badDebt.Status = "approved"
	badDebt.ApprovedBy = &approvedBy
	badDebt.ApprovedAt = &now
	badDebt.ApprovalNotes = notes

	if err := s.badDebtRepo.Update(ctx, badDebt); err != nil {
		return nil, fmt.Errorf("审批坏账核销失败: %w", err)
	}

	// 如果是针对特定发票的坏账核销，更新发票状态
	if badDebt.InvoiceID != nil {
		// 检查是否全部核销
		if badDebt.RemainingAmount <= 0 {
			// 可以考虑将发票标记为已核销
		}
	}

	return s.GetBadDebtByID(ctx, id)
}

// RejectBadDebt 审批拒绝坏账核销
func (s *BadDebtService) RejectBadDebt(ctx context.Context, id uint, notes string) (*BadDebtResponse, error) {
	badDebt, err := s.badDebtRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询坏账核销记录失败: %w", err)
	}
	if badDebt == nil {
		return nil, errors.New("坏账核销记录不存在")
	}

	if badDebt.Status != "pending" {
		return nil, errors.New("只有待审批状态的坏账核销可以拒绝")
	}

	badDebt.Status = "rejected"
	badDebt.ApprovalNotes = notes

	if err := s.badDebtRepo.Update(ctx, badDebt); err != nil {
		return nil, fmt.Errorf("拒绝坏账核销失败: %w", err)
	}

	return s.GetBadDebtByID(ctx, id)
}

// GetPendingBadDebts 获取待审批的坏账核销列表
func (s *BadDebtService) GetPendingBadDebts(ctx context.Context) ([]*BadDebtResponse, error) {
	badDebts, err := s.badDebtRepo.GetPending(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询待审批坏账核销失败: %w", err)
	}

	result := make([]*BadDebtResponse, len(badDebts))
	for i, bd := range badDebts {
		result[i] = s.convertToResponse(bd)
	}

	return result, nil
}

// convertToResponse 转换为响应格式
func (s *BadDebtService) convertToResponse(badDebt *models.BadDebtRecord) *BadDebtResponse {
	resp := &BadDebtResponse{
		ID:              badDebt.ID,
		ContractID:      badDebt.ContractID,
		InvoiceID:       badDebt.InvoiceID,
		OriginalAmount:  badDebt.OriginalAmount,
		WriteOffAmount:  badDebt.WriteOffAmount,
		RemainingAmount: badDebt.RemainingAmount,
		Reason:          badDebt.Reason,
		ReasonType:      badDebt.ReasonType,
		Status:          badDebt.Status,
		ApprovalNotes:   badDebt.ApprovalNotes,
		CreatedAt:       badDebt.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       badDebt.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if badDebt.ApprovedBy != nil {
		resp.ApprovedBy = badDebt.ApprovedBy
	}
	if badDebt.ApprovedAt != nil {
		formatted := badDebt.ApprovedAt.Format("2006-01-02 15:04:05")
		resp.ApprovedAt = &formatted
	}
	// AttachmentIDs 是 JSON 类型，需要特殊处理
	// 如果需要，可以从 JSON 中解析出附件ID列表

	return resp
}
