package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// CommissionService 提成计算服务
type CommissionService struct {
	commissionRepo     repositories.CommissionRepository
	commissionRuleRepo repositories.CommissionRuleRepository
	paymentRepo        repositories.PaymentRepository
	contractRepo       repositories.ContractRepository
	invoiceRepo        repositories.InvoiceRepository
	userRepo           repositories.UserRepository
	caseRepo           repositories.CaseRepository
}

// NewCommissionService 创建提成计算服务实例
func NewCommissionService(
	commissionRepo repositories.CommissionRepository,
	paymentRepo repositories.PaymentRepository,
	contractRepo repositories.ContractRepository,
	userRepo repositories.UserRepository,
	caseRepo repositories.CaseRepository,
) *CommissionService {
	return &CommissionService{
		commissionRepo: commissionRepo,
		paymentRepo:    paymentRepo,
		contractRepo:   contractRepo,
		userRepo:       userRepo,
		caseRepo:       caseRepo,
	}
}

// SetInvoiceRepository 设置发票仓储（延迟注入）
func (s *CommissionService) SetInvoiceRepository(invoiceRepo repositories.InvoiceRepository) {
	s.invoiceRepo = invoiceRepo
}

// SetCommissionRuleRepository 设置分成规则仓储（延迟注入）
func (s *CommissionService) SetCommissionRuleRepository(ruleRepo repositories.CommissionRuleRepository) {
	s.commissionRuleRepo = ruleRepo
}

// CommissionRule 提成规则
type CommissionRule struct {
	Role            string  // 角色: source/lawyer/assistant
	MinAmount       float64 // 最小金额
	MaxAmount       float64 // 最大金额
	BaseRate        float64 // 基础提成比例
	PerformanceRate float64 // 绩效提成比例（可选）
}

// CalculateCommissionRequest 计算提成请求
type CalculateCommissionRequest struct {
	PaymentID     uint    `json:"payment_id" binding:"required"`
	CostDeduction float64 `json:"cost_deduction"`
}

// CommissionResponse 提成响应
type CommissionResponse struct {
	ID               uint    `json:"id"`
	CommissionCode   string  `json:"commission_code"`
	ContractID       uint    `json:"contract_id"`
	PaymentID        uint    `json:"payment_id"`
	CaseID           *uint   `json:"case_id,omitempty"`
	BeneficiaryID    uint    `json:"beneficiary_id"`
	BeneficiaryRole  string  `json:"beneficiary_role"`
	PaymentAmount    float64 `json:"payment_amount"`
	CostDeduction    float64 `json:"cost_deduction"`
	CommissionBase   float64 `json:"commission_base"`
	CommissionRate   float64 `json:"commission_rate"`
	CommissionAmount float64 `json:"commission_amount"`
	Status           string  `json:"status"`
	PaidDate         *string `json:"paid_date,omitempty"`
	PaymentVoucher   string  `json:"payment_voucher"`
	CalculatedAt     *string `json:"calculated_at,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	// 关联数据
	Contract    *ContractSummary `json:"contract,omitempty"`
	Payment     *PaymentSummary2 `json:"payment,omitempty"`
	Case        *CaseSummary     `json:"case,omitempty"`
	Beneficiary *UserSummary     `json:"beneficiary,omitempty"`
}

// PaymentSummary2 回款摘要（简化版）
type PaymentSummary2 struct {
	ID          uint    `json:"id"`
	PaymentCode string  `json:"payment_code"`
	Amount      float64 `json:"amount"`
}

// UserSummary 用户摘要
type UserSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ListCommissionsRequest 提成列表请求
type ListCommissionsRequest struct {
	Page            int    `json:"page" form:"page" binding:"min=1"`
	PageSize        int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status          string `json:"status" form:"status" binding:"omitempty,oneof=pending calculated paid cancelled"`
	BeneficiaryID   uint   `json:"beneficiary_id" form:"beneficiary_id"`
	BeneficiaryRole string `json:"beneficiary_role" form:"beneficiary_role" binding:"omitempty,oneof=source lawyer assistant"`
	ContractID      uint   `json:"contract_id" form:"contract_id"`
	CaseID          *uint  `json:"case_id" form:"case_id"`
	DateFrom        string `json:"date_from" form:"date_from"`
	DateTo          string `json:"date_to" form:"date_to"`
}

// ListCommissionsResponse 提成列表响应
type ListCommissionsResponse struct {
	Commissions []*CommissionResponse `json:"commissions"`
	Pagination  Pagination            `json:"pagination"`
}

// CalculateCommissions 计算提成
func (s *CommissionService) CalculateCommissions(ctx context.Context, req *CalculateCommissionRequest) ([]*CommissionResponse, error) {
	// 获取回款记录
	payment, err := s.paymentRepo.FindByID(ctx, req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("查询回款记录失败: %w", err)
	}
	if payment == nil {
		return nil, errors.New("回款记录不存在")
	}

	if payment.Status != "confirmed" {
		return nil, errors.New("只有已确认的回款可以计算提成")
	}

	// 获取发票
	invoice, err := s.invoiceRepo.FindByID(ctx, payment.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("查询发票失败: %w", err)
	}
	if invoice == nil {
		return nil, errors.New("发票不存在")
	}

	// 获取合同
	if invoice.ContractID == nil {
		return nil, errors.New("发票未关联合同")
	}
	contract, err := s.contractRepo.FindByID(ctx, *invoice.ContractID)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	// 检查是否已计算过提成
	existingCommissions, _ := s.commissionRepo.GetByPaymentID(ctx, req.PaymentID)
	if len(existingCommissions) > 0 {
		return nil, errors.New("该回款已计算过提成")
	}

	// 获取案件信息
	var case_ *models.Case
	if contract.CaseID != nil {
		case_, _ = s.caseRepo.FindByID(ctx, *contract.CaseID)
	}

	// 获取团队成员（案源、主办律师、协办律师）
	// 这里假设从案件或合同中获取团队信息
	// 实际实现需要根据业务逻辑调整
	beneficiaries := s.getBeneficiaries(ctx, contract, case_)

	// 计算提成基数
	commissionBase := payment.Amount - req.CostDeduction
	if commissionBase < 0 {
		return nil, errors.New("成本扣除不能超过回款金额")
	}

	now := time.Now()
	var result []*CommissionResponse

	// 为每个受益人计算提成
	for _, beneficiary := range beneficiaries {
		rate := s.getCommissionRate(beneficiary.Role, payment.Amount)
		commissionAmount := commissionBase * (rate / 100)

		if commissionAmount <= 0 {
			continue
		}

		// 生成提成编号
		commissionCode := fmt.Sprintf("COM-%d%s", time.Now().Unix(), generateRandomCode(4))

		var caseID *uint
		if case_ != nil {
			caseID = &case_.ID
		}

		commission := &models.CommissionRecord{
			CommissionCode:   commissionCode,
			ContractID:       contract.ID,
			PaymentID:        req.PaymentID,
			CaseID:           caseID,
			BeneficiaryID:    beneficiary.UserID,
			BeneficiaryRole:  beneficiary.Role,
			PaymentAmount:    payment.Amount,
			CostDeduction:    req.CostDeduction,
			CommissionBase:   commissionBase,
			CommissionRate:   rate,
			CommissionAmount: commissionAmount,
			Status:           "calculated",
			CalculatedAt:     &now,
		}

		if err := s.commissionRepo.Create(ctx, commission); err != nil {
			return nil, fmt.Errorf("创建提成记录失败: %w", err)
		}

		result = append(result, s.convertToResponse(commission))
	}

	return result, nil
}

// Beneficiary 受益人
type Beneficiary struct {
	UserID uint
	Role   string // source/lawyer/assistant
}

// getBeneficiaries 获取受益人列表
func (s *CommissionService) getBeneficiaries(ctx context.Context, contract *models.Contract, case_ *models.Case) []Beneficiary {
	var beneficiaries []Beneficiary

	// 如果关联案件，从案件获取团队信息
	if case_ != nil {
		// 主办律师
		if case_.LawyerID > 0 {
			beneficiaries = append(beneficiaries, Beneficiary{
				UserID: case_.LawyerID,
				Role:   "lawyer",
			})
		}

		// 案源律师 - 如果有合同签署人，可以作为案源
		// 这里简化处理，假设主办律师就是案源
		// 实际业务中可能需要额外的字段来区分
	}

	return beneficiaries
}

// getCommissionRate 获取提成比例
func (s *CommissionService) getCommissionRate(role string, amount float64) float64 {
	// 优先从数据库读取规则
	if s.commissionRuleRepo != nil {
		rules, err := s.commissionRuleRepo.FindActiveByRole(context.Background(), role)
		if err == nil && len(rules) > 0 {
			for _, rule := range rules {
				if amount >= rule.MinAmount && (rule.MaxAmount <= 0 || amount < rule.MaxAmount) {
					return rule.BaseRate
				}
			}
			// 没有匹配的金额区间，返回最低规则的基础比例
			return rules[len(rules)-1].BaseRate
		}
	}

	// fallback: 硬编码默认规则
	switch role {
	case "source":
		if amount < 10000 {
			return 10
		} else if amount < 50000 {
			return 15
		} else if amount < 100000 {
			return 20
		}
		return 30
	case "lawyer":
		if amount < 10000 {
			return 20
		} else if amount < 50000 {
			return 30
		} else if amount < 100000 {
			return 40
		}
		return 50
	case "assistant":
		if amount < 10000 {
			return 5
		} else if amount < 50000 {
			return 8
		} else if amount < 100000 {
			return 12
		}
		return 15
	default:
		return 0
	}
}

// GetCommissionByID 根据ID获取提成详情
func (s *CommissionService) GetCommissionByID(ctx context.Context, id uint) (*CommissionResponse, error) {
	commission, err := s.commissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询提成记录失败: %w", err)
	}
	if commission == nil {
		return nil, errors.New("提成记录不存在")
	}

	return s.convertToResponse(commission), nil
}

// ListCommissions 获取提成列表
func (s *CommissionService) ListCommissions(ctx context.Context, req *ListCommissionsRequest) (*ListCommissionsResponse, error) {
	params := &repositories.CommissionListParams{
		Page:            req.Page,
		PageSize:        req.PageSize,
		Status:          req.Status,
		BeneficiaryID:   req.BeneficiaryID,
		BeneficiaryRole: req.BeneficiaryRole,
		ContractID:      req.ContractID,
		CaseID:          req.CaseID,
		DateFrom:        req.DateFrom,
		DateTo:          req.DateTo,
	}

	commissions, total, err := s.commissionRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询提成列表失败: %w", err)
	}

	response := &ListCommissionsResponse{
		Commissions: make([]*CommissionResponse, len(commissions)),
		Pagination: Pagination{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}

	for i, c := range commissions {
		response.Commissions[i] = s.convertToResponse(c)
	}

	return response, nil
}

// MarkAsPaid 标记提成已支付
func (s *CommissionService) MarkAsPaid(ctx context.Context, id uint, paidDate string, voucher string) (*CommissionResponse, error) {
	commission, err := s.commissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询提成记录失败: %w", err)
	}
	if commission == nil {
		return nil, errors.New("提成记录不存在")
	}

	if commission.Status != "calculated" {
		return nil, errors.New("只有已计算状态的提成可以标记为已支付")
	}

	// 解析支付日期
	parsedDate, err := time.Parse("2006-01-02", paidDate)
	if err != nil {
		return nil, fmt.Errorf("支付日期格式错误，应为 YYYY-MM-DD: %w", err)
	}

	commission.Status = "paid"
	commission.PaidDate = &parsedDate
	commission.PaymentVoucher = voucher

	if err := s.commissionRepo.Update(ctx, commission); err != nil {
		return nil, fmt.Errorf("更新提成记录失败: %w", err)
	}

	return s.GetCommissionByID(ctx, id)
}

// CancelCommission 取消提成
func (s *CommissionService) CancelCommission(ctx context.Context, id uint) (*CommissionResponse, error) {
	commission, err := s.commissionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询提成记录失败: %w", err)
	}
	if commission == nil {
		return nil, errors.New("提成记录不存在")
	}

	if commission.Status == "paid" {
		return nil, errors.New("已支付的提成不能取消")
	}

	commission.Status = "cancelled"

	if err := s.commissionRepo.Update(ctx, commission); err != nil {
		return nil, fmt.Errorf("取消提成记录失败: %w", err)
	}

	return s.GetCommissionByID(ctx, id)
}

// GetCommissionsByBeneficiary 获取受益人的提成记录
func (s *CommissionService) GetCommissionsByBeneficiary(ctx context.Context, beneficiaryID uint) ([]*CommissionResponse, error) {
	commissions, err := s.commissionRepo.GetByBeneficiaryID(ctx, beneficiaryID)
	if err != nil {
		return nil, fmt.Errorf("查询受益人提成记录失败: %w", err)
	}

	result := make([]*CommissionResponse, len(commissions))
	for i, c := range commissions {
		result[i] = s.convertToResponse(c)
	}

	return result, nil
}

// convertToResponse 转换为响应格式
func (s *CommissionService) convertToResponse(commission *models.CommissionRecord) *CommissionResponse {
	resp := &CommissionResponse{
		ID:               commission.ID,
		CommissionCode:   commission.CommissionCode,
		ContractID:       commission.ContractID,
		PaymentID:        commission.PaymentID,
		CaseID:           commission.CaseID,
		BeneficiaryID:    commission.BeneficiaryID,
		BeneficiaryRole:  commission.BeneficiaryRole,
		PaymentAmount:    commission.PaymentAmount,
		CostDeduction:    commission.CostDeduction,
		CommissionBase:   commission.CommissionBase,
		CommissionRate:   commission.CommissionRate,
		CommissionAmount: commission.CommissionAmount,
		Status:           commission.Status,
		PaymentVoucher:   commission.PaymentVoucher,
		CreatedAt:        commission.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        commission.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if commission.CalculatedAt != nil {
		formatted := commission.CalculatedAt.Format("2006-01-02 15:04:05")
		resp.CalculatedAt = &formatted
	}
	if commission.PaidDate != nil {
		formatted := commission.PaidDate.Format("2006-01-02")
		resp.PaidDate = &formatted
	}

	// 关联数据需要额外查询
	// 这里简化处理，实际使用时可以添加预加载

	return resp
}

// GetCommissionStats 获取提成统计信息
func (s *CommissionService) GetCommissionStats(ctx context.Context) (*CommissionStats, error) {
	var stats CommissionStats

	// 获取所有提成记录进行统计
	// 注意：实际项目中应该在 repository 层实现专门的统计方法
	allCommissions, total, err := s.commissionRepo.List(ctx, &repositories.CommissionListParams{
		Page:     1,
		PageSize: 10000, // 获取所有记录进行统计
	})
	if err != nil {
		return nil, fmt.Errorf("查询提成记录失败: %w", err)
	}

	stats.TotalCommissions = total

	for _, c := range allCommissions {
		// 按状态统计
		switch c.Status {
		case "pending":
			stats.PendingCommissions++
		case "calculated":
			stats.CalculatedCommissions++
		case "paid":
			stats.PaidCommissions++
		case "cancelled":
			stats.CancelledCommissions++
		}

		// 累计总金额
		stats.TotalCommissionAmount += c.CommissionAmount

		// 待支付金额
		if c.Status == "calculated" {
			stats.PendingCommissionAmount += c.CommissionAmount
		}

		// 本月提成金额
		if c.CreatedAt.Month() == time.Now().Month() && c.CreatedAt.Year() == time.Now().Year() {
			stats.MonthCommissionAmount += c.CommissionAmount
		}
	}

	return &stats, nil
}

// CommissionStats 提成统计信息
type CommissionStats struct {
	TotalCommissions        int64   `json:"total_commissions"`
	PendingCommissions      int64   `json:"pending_commissions"`
	CalculatedCommissions   int64   `json:"calculated_commissions"`
	PaidCommissions         int64   `json:"paid_commissions"`
	CancelledCommissions    int64   `json:"cancelled_commissions"`
	TotalCommissionAmount   float64 `json:"total_commission_amount"`
	PendingCommissionAmount float64 `json:"pending_commission_amount"`
	MonthCommissionAmount   float64 `json:"month_commission_amount"`
}

// ==================== 分成规则 CRUD ====================

// ListCommissionRules 获取所有分成规则
func (s *CommissionService) ListCommissionRules(ctx context.Context) ([]*models.CommissionRule, error) {
	return s.commissionRuleRepo.FindAll(ctx)
}

// CreateCommissionRule 创建分成规则
func (s *CommissionService) CreateCommissionRule(ctx context.Context, rule *models.CommissionRule) error {
	return s.commissionRuleRepo.Create(ctx, rule)
}

// UpdateCommissionRule 更新分成规则
func (s *CommissionService) UpdateCommissionRule(ctx context.Context, rule *models.CommissionRule) error {
	existing, err := s.commissionRuleRepo.FindByID(ctx, rule.ID)
	if err != nil {
		return fmt.Errorf("规则不存在: %w", err)
	}
	existing.Name = rule.Name
	existing.Role = rule.Role
	existing.MinAmount = rule.MinAmount
	existing.MaxAmount = rule.MaxAmount
	existing.BaseRate = rule.BaseRate
	existing.PerformanceRate = rule.PerformanceRate
	existing.Priority = rule.Priority
	existing.Active = rule.Active
	existing.EffectiveDate = rule.EffectiveDate
	existing.ExpiryDate = rule.ExpiryDate
	return s.commissionRuleRepo.Update(ctx, existing)
}

// DeleteCommissionRule 删除分成规则
func (s *CommissionService) DeleteCommissionRule(ctx context.Context, id uint) error {
	return s.commissionRuleRepo.Delete(ctx, id)
}
