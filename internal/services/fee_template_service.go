package services

import (
	"context"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// FeeTemplateService 费率模板服务
type FeeTemplateService struct {
	feeTemplateRepo repositories.FeeTemplateRepository
	contractRepo    repositories.ContractRepository
}

// NewFeeTemplateService 创建费率模板服务
func NewFeeTemplateService(feeTemplateRepo repositories.FeeTemplateRepository, contractRepo repositories.ContractRepository) *FeeTemplateService {
	return &FeeTemplateService{
		feeTemplateRepo: feeTemplateRepo,
		contractRepo:    contractRepo,
	}
}

// CreateFeeTemplateRequest 创建费率模板请求
type CreateFeeTemplateRequest struct {
	Name                 string      `json:"name" binding:"required"`
	CaseType             string      `json:"case_type" binding:"required,oneof=litigation non_litigation consulting"`
	BillingType          string      `json:"billing_type" binding:"required,oneof=hourly fixed hybrid retainer"`
	BaseRates            models.JSON `json:"base_rates" binding:"required"`
	PerformanceBonusRate float64     `json:"performance_bonus_rate"`
	MinAmount            float64     `json:"min_amount"`
	MaxAmount            float64     `json:"max_amount"`
	CostRate             float64     `json:"cost_rate"`
}

// CreateFeeTemplate 创建费率模板
func (s *FeeTemplateService) CreateFeeTemplate(ctx context.Context, req *CreateFeeTemplateRequest) (*models.FeeTemplate, error) {
	template := &models.FeeTemplate{
		Name:                 req.Name,
		CaseType:             req.CaseType,
		BillingType:          req.BillingType,
		BaseRates:            req.BaseRates,
		PerformanceBonusRate: req.PerformanceBonusRate,
		MinAmount:            req.MinAmount,
		MaxAmount:            req.MaxAmount,
		CostRate:             req.CostRate,
		Active:               true,
	}
	if err := s.feeTemplateRepo.Create(ctx, template); err != nil {
		return nil, err
	}
	return template, nil
}

// GetFeeTemplate 获取费率模板
func (s *FeeTemplateService) GetFeeTemplate(ctx context.Context, id uint) (*models.FeeTemplate, error) {
	return s.feeTemplateRepo.GetByID(ctx, id)
}

// UpdateFeeTemplate 更新费率模板
func (s *FeeTemplateService) UpdateFeeTemplate(ctx context.Context, id uint, req *CreateFeeTemplateRequest) (*models.FeeTemplate, error) {
	template, err := s.feeTemplateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	template.Name = req.Name
	template.CaseType = req.CaseType
	template.BillingType = req.BillingType
	template.BaseRates = req.BaseRates
	template.PerformanceBonusRate = req.PerformanceBonusRate
	template.MinAmount = req.MinAmount
	template.MaxAmount = req.MaxAmount
	template.CostRate = req.CostRate
	if err := s.feeTemplateRepo.Update(ctx, template); err != nil {
		return nil, err
	}
	return template, nil
}

// DeleteFeeTemplate 删除费率模板
func (s *FeeTemplateService) DeleteFeeTemplate(ctx context.Context, id uint) error {
	return s.feeTemplateRepo.Delete(ctx, id)
}

// ListFeeTemplates 列出费率模板
func (s *FeeTemplateService) ListFeeTemplates(ctx context.Context, page, pageSize int) ([]*models.FeeTemplate, int64, error) {
	return s.feeTemplateRepo.List(ctx, page, pageSize)
}

// GetByCaseType 根据案件类型获取适用模板
func (s *FeeTemplateService) GetByCaseType(ctx context.Context, caseType string) ([]*models.FeeTemplate, error) {
	return s.feeTemplateRepo.GetByCaseType(ctx, caseType)
}

// CalculateFee 根据模板计算预估费用
func (s *FeeTemplateService) CalculateFee(ctx context.Context, templateID uint, amount float64) (*FeeCalculationResult, error) {
	template, err := s.feeTemplateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if amount < template.MinAmount {
		amount = template.MinAmount
	}

	totalBaseRate := 0.0
	for _, rate := range template.BaseRates {
		if r, ok := rate.(float64); ok {
			totalBaseRate += r
		}
	}

	baseFee := amount * totalBaseRate
	performanceBonus := amount * (template.PerformanceBonusRate / 100)
	totalFee := baseFee + performanceBonus
	costDeduction := totalFee * (template.CostRate / 100)

	return &FeeCalculationResult{
		TemplateID:      templateID,
		TemplateName:   template.Name,
		BillingType:     template.BillingType,
		BaseFee:         baseFee,
	PerformanceBonus: performanceBonus,
	TotalFee:         totalFee,
		CostDeduction:   costDeduction,
		NetFee:          totalFee - costDeduction,
	}, nil
}

// FeeCalculationResult 费用计算结果
type FeeCalculationResult struct {
	TemplateID      uint    `json:"template_id"`
	TemplateName   string  `json:"template_name"`
	BillingType     string  `json:"billing_type"`
	BaseFee         float64 `json:"base_fee"`
	PerformanceBonus float64 `json:"performance_bonus"`
	TotalFee         float64 `json:"total_fee"`
	CostDeduction   float64 `json:"cost_deduction"`
	NetFee          float64 `json:"net_fee"`
}
