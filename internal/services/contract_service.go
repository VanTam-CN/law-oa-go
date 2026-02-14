package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ContractService 合同服务
type ContractService struct {
	contractRepo    repositories.ContractRepository
	milestoneRepo   repositories.PaymentMilestoneRepository
	clientRepo      repositories.ClientRepository
	caseRepo        repositories.CaseRepository
	userRepo        repositories.UserRepository
}

// NewContractService 创建合同服务实例
func NewContractService(
	contractRepo repositories.ContractRepository,
	milestoneRepo repositories.PaymentMilestoneRepository,
	clientRepo repositories.ClientRepository,
	caseRepo repositories.CaseRepository,
	userRepo repositories.UserRepository,
) *ContractService {
	return &ContractService{
		contractRepo:  contractRepo,
		milestoneRepo: milestoneRepo,
		clientRepo:    clientRepo,
		caseRepo:      caseRepo,
		userRepo:      userRepo,
	}
}

// CreateContractRequest 创建合同请求
type CreateContractRequest struct {
	ContractCode   string                    `json:"contract_code" binding:"required,min=1,max=50"`
	CaseID         *uint                     `json:"case_id,omitempty"`
	ClientID       uint                      `json:"client_id" binding:"required"`
	ContractAmount float64                   `json:"contract_amount" binding:"required,gt=0"`
	Currency       string                    `json:"currency" binding:"required,oneof=CNY USD EUR"`
	BillingCycle   string                    `json:"billing_cycle" binding:"required,oneof=一次性 分期 按小时"`
	PaymentTerms   string                    `json:"payment_terms" binding:"max=500"`
	StartDate      *string                   `json:"start_date,omitempty"`
	EndDate        *string                   `json:"end_date,omitempty"`
	ContractType   string                    `json:"contract_type" binding:"required,oneof=original supplementary"`
	ParentContractID *uint                   `json:"parent_contract_id,omitempty"`
	DocumentID     *uint                     `json:"document_id,omitempty"`
	// 付款计划
	Milestones     []CreateMilestoneRequest  `json:"milestones,omitempty"`
}

// CreateMilestoneRequest 创建付款计划请求
type CreateMilestoneRequest struct {
	Name       string   `json:"name" binding:"required,min=1,max=200"`
	Amount     float64  `json:"amount" binding:"required,gt=0"`
	Percentage float64  `json:"percentage" binding:"required,gte=0,lte=100"`
	DueDate    *string  `json:"due_date,omitempty"`
	Condition  string   `json:"condition" binding:"max=500"`
}

// UpdateContractRequest 更新合同请求
type UpdateContractRequest struct {
	ContractAmount *float64 `json:"contract_amount,omitempty" binding:"omitempty,gt=0"`
	Currency       *string  `json:"currency,omitempty" binding:"omitempty,oneof=CNY USD EUR"`
	BillingCycle   *string  `json:"billing_cycle,omitempty" binding:"omitempty,oneof=一次性 分期 按小时"`
	PaymentTerms   *string  `json:"payment_terms,omitempty" binding:"omitempty,max=500"`
	StartDate      *string  `json:"start_date,omitempty"`
	EndDate        *string  `json:"end_date,omitempty"`
	DocumentID     *uint    `json:"document_id,omitempty"`
}

// ContractResponse 合同响应
type ContractResponse struct {
	ID              uint                       `json:"id"`
	ContractCode    string                     `json:"contract_code"`
	CaseID          *uint                      `json:"case_id,omitempty"`
	ClientID        uint                       `json:"client_id"`
	ContractAmount  float64                    `json:"contract_amount"`
	Currency        string                     `json:"currency"`
	BillingCycle    string                     `json:"billing_cycle"`
	PaymentTerms    string                     `json:"payment_terms"`
	StartDate       *string                    `json:"start_date,omitempty"`
	EndDate         *string                    `json:"end_date,omitempty"`
	Status          string                     `json:"status"`
	ContractType    string                     `json:"contract_type"`
	ParentContractID *uint                     `json:"parent_contract_id,omitempty"`
	SignedAt        *string                    `json:"signed_at,omitempty"`
	DocumentID      *uint                      `json:"document_id,omitempty"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Client          *ClientSummary             `json:"client,omitempty"`
	Case            *CaseSummary               `json:"case,omitempty"`
	Milestones      []*PaymentMilestoneResponse `json:"milestones,omitempty"`
	SupplementaryContracts []*ContractResponse  `json:"supplementary_contracts,omitempty"`
}

// PaymentMilestoneResponse 付款计划响应
type PaymentMilestoneResponse struct {
	ID         uint    `json:"id"`
	ContractID uint    `json:"contract_id"`
	Name       string  `json:"name"`
	Sequence   int     `json:"sequence"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
	DueDate    *string `json:"due_date,omitempty"`
	Condition  string  `json:"condition"`
	Status     string  `json:"status"`
	InvoiceID  *uint   `json:"invoice_id,omitempty"`
	PaidAmount float64 `json:"paid_amount"`
}

// ListContractsRequest 合同列表请求
type ListContractsRequest struct {
	Page             int     `json:"page" form:"page" binding:"min=1"`
	PageSize         int     `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Status           string  `json:"status" form:"status" binding:"omitempty,oneof=draft active suspended completed cancelled"`
	ContractType     string  `json:"contract_type" form:"contract_type" binding:"omitempty,oneof=original supplementary"`
	ClientID         uint    `json:"client_id" form:"client_id"`
	CaseID           uint    `json:"case_id" form:"case_id"`
	Search           string  `json:"search" form:"search"`
	StartDateFrom    string  `json:"start_date_from" form:"start_date_from"`
	StartDateTo      string  `json:"start_date_to" form:"start_date_to"`
	EndDateFrom      string  `json:"end_date_from" form:"end_date_from"`
	EndDateTo        string  `json:"end_date_to" form:"end_date_to"`
}

// ListContractsResponse 合同列表响应
type ListContractsResponse struct {
	Contracts  []*ContractResponse `json:"contracts"`
	Pagination Pagination          `json:"pagination"`
}

// CreateContract 创建合同
func (s *ContractService) CreateContract(ctx context.Context, req *CreateContractRequest) (*ContractResponse, error) {
	// 验证客户是否存在
	client, err := s.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("查询客户失败: %w", err)
	}
	if client == nil {
		return nil, errors.New("客户不存在")
	}

	// 如果有关联案件，验证案件是否存在
	if req.CaseID != nil {
		case_, err := s.caseRepo.FindByID(ctx, *req.CaseID)
		if err != nil {
			return nil, fmt.Errorf("查询案件失败: %w", err)
		}
		if case_ == nil {
			return nil, errors.New("案件不存在")
		}
	}

	// 如果是补充协议，验证主合同是否存在
	if req.ContractType == "supplementary" {
		if req.ParentContractID == nil {
			return nil, errors.New("补充协议必须指定主合同")
		}
		parentContract, err := s.contractRepo.FindByID(ctx, *req.ParentContractID)
		if err != nil {
			return nil, fmt.Errorf("查询主合同失败: %w", err)
		}
		if parentContract == nil {
			return nil, errors.New("主合同不存在")
		}
		if parentContract.ContractType != "original" {
			return nil, errors.New("主合同必须是原始合同")
		}
	}

	// 验证合同编号是否已存在
	existingContract, err := s.contractRepo.FindByCode(ctx, req.ContractCode)
	if err != nil {
		return nil, fmt.Errorf("查询合同编号失败: %w", err)
	}
	if existingContract != nil {
		return nil, errors.New("合同编号已存在")
	}

	// 解析日期
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		startDate = &parsedDate
	}
	if req.EndDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		endDate = &parsedDate
	}

	// 验证日期范围
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return nil, errors.New("结束日期不能早于开始日期")
	}

	// 验证付款计划总额
	var milestoneTotal float64
	for _, m := range req.Milestones {
		milestoneTotal += m.Amount
	}
	if len(req.Milestones) > 0 && milestoneTotal > req.ContractAmount {
		return nil, fmt.Errorf("付款计划总额(%.2f)不能超过合同金额(%.2f)", milestoneTotal, req.ContractAmount)
	}

	// 创建合同
	contract := &models.Contract{
		ContractCode:    req.ContractCode,
		CaseID:          req.CaseID,
		ClientID:        req.ClientID,
		ContractAmount:  req.ContractAmount,
		Currency:        req.Currency,
		BillingCycle:    req.BillingCycle,
		PaymentTerms:    req.PaymentTerms,
		StartDate:       startDate,
		EndDate:         endDate,
		Status:          "draft",
		ContractType:    req.ContractType,
		ParentContractID: req.ParentContractID,
		DocumentID:      req.DocumentID,
	}

	if err := s.contractRepo.Create(ctx, contract); err != nil {
		return nil, fmt.Errorf("创建合同失败: %w", err)
	}

	// 创建付款计划
	for i, m := range req.Milestones {
		var dueDate *time.Time
		if m.DueDate != nil {
			parsedDate, err := time.Parse("2006-01-02", *m.DueDate)
			if err != nil {
				return nil, fmt.Errorf("付款计划日期格式错误: %w", err)
			}
			dueDate = &parsedDate
		}

		milestone := &models.PaymentMilestone{
			ContractID: contract.ID,
			Name:       m.Name,
			Sequence:   i + 1,
			Amount:     m.Amount,
			Percentage: m.Percentage,
			DueDate:    dueDate,
			Condition:  m.Condition,
			Status:     "pending",
		}

		if err := s.milestoneRepo.Create(ctx, milestone); err != nil {
			return nil, fmt.Errorf("创建付款计划失败: %w", err)
		}
	}

	// 获取创建的完整合同信息
	return s.GetContractByID(ctx, contract.ID)
}

// GetContractByID 根据ID获取合同详情
func (s *ContractService) GetContractByID(ctx context.Context, id uint) (*ContractResponse, error) {
	contract, err := s.contractRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	return s.convertToResponse(ctx, contract), nil
}

// UpdateContract 更新合同
func (s *ContractService) UpdateContract(ctx context.Context, id uint, req *UpdateContractRequest) (*ContractResponse, error) {
	// 获取现有合同
	contract, err := s.contractRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	// 只有草稿状态的合同可以编辑
	if contract.Status != "draft" {
		return nil, errors.New("只有草稿状态的合同可以编辑")
	}

	// 解析日期
	if req.StartDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		contract.StartDate = &parsedDate
	}
	if req.EndDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		contract.EndDate = &parsedDate
	}

	// 验证日期范围
	if contract.StartDate != nil && contract.EndDate != nil && contract.EndDate.Before(*contract.StartDate) {
		return nil, errors.New("结束日期不能早于开始日期")
	}

	// 更新字段
	if req.ContractAmount != nil {
		contract.ContractAmount = *req.ContractAmount
	}
	if req.Currency != nil {
		contract.Currency = *req.Currency
	}
	if req.BillingCycle != nil {
		contract.BillingCycle = *req.BillingCycle
	}
	if req.PaymentTerms != nil {
		contract.PaymentTerms = *req.PaymentTerms
	}
	if req.DocumentID != nil {
		contract.DocumentID = req.DocumentID
	}

	if err := s.contractRepo.Update(ctx, contract); err != nil {
		return nil, fmt.Errorf("更新合同失败: %w", err)
	}

	return s.GetContractByID(ctx, id)
}

// DeleteContract 删除合同
func (s *ContractService) DeleteContract(ctx context.Context, id uint) error {
	contract, err := s.contractRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return errors.New("合同不存在")
	}

	// 只有草稿状态的合同可以删除
	if contract.Status != "draft" {
		return errors.New("只有草稿状态的合同可以删除")
	}

	// 如果是主合同，检查是否有补充协议
	if contract.ContractType == "original" {
		supplementaryContracts, err := s.contractRepo.GetSupplementaryContracts(ctx, id)
		if err != nil {
			return fmt.Errorf("查询补充协议失败: %w", err)
		}
		if len(supplementaryContracts) > 0 {
			return errors.New("主合同存在补充协议，无法删除")
		}
	}

	// 删除付款计划
	if err := s.milestoneRepo.DeleteByContractID(ctx, id); err != nil {
		return fmt.Errorf("删除付款计划失败: %w", err)
	}

	// 删除合同
	if err := s.contractRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除合同失败: %w", err)
	}

	return nil
}

// ListContracts 获取合同列表
func (s *ContractService) ListContracts(ctx context.Context, req *ListContractsRequest) (*ListContractsResponse, error) {
	params := &repositories.ContractListParams{
		Page:         req.Page,
		PageSize:     req.PageSize,
		Status:       req.Status,
		ContractType: req.ContractType,
		ClientID:     req.ClientID,
		CaseID:       req.CaseID,
		Search:       req.Search,
		StartDateFrom: req.StartDateFrom,
		StartDateTo:   req.StartDateTo,
		EndDateFrom:   req.EndDateFrom,
		EndDateTo:     req.EndDateTo,
	}

	contracts, total, err := s.contractRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询合同列表失败: %w", err)
	}

	response := &ListContractsResponse{
		Contracts: make([]*ContractResponse, len(contracts)),
		Pagination: Pagination{
			Page:    req.Page,
			PageSize: req.PageSize,
			Total:   total,
		},
	}

	for i, contract := range contracts {
		response.Contracts[i] = s.convertToResponseSimple(ctx, contract)
	}

	return response, nil
}

// ActivateContract 激活合同
func (s *ContractService) ActivateContract(ctx context.Context, id uint) (*ContractResponse, error) {
	contract, err := s.contractRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	if contract.Status != "draft" {
		return nil, errors.New("只有草稿状态的合同可以激活")
	}

	now := time.Now()
	contract.Status = "active"
	contract.SignedAt = &now

	if err := s.contractRepo.Update(ctx, contract); err != nil {
		return nil, fmt.Errorf("激活合同失败: %w", err)
	}

	return s.GetContractByID(ctx, id)
}

// SuspendContract 暂停合同
func (s *ContractService) SuspendContract(ctx context.Context, id uint) (*ContractResponse, error) {
	contract, err := s.contractRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	if contract.Status != "active" {
		return nil, errors.New("只有激活状态的合同可以暂停")
	}

	contract.Status = "suspended"

	if err := s.contractRepo.Update(ctx, contract); err != nil {
		return nil, fmt.Errorf("暂停合同失败: %w", err)
	}

	return s.GetContractByID(ctx, id)
}

// CompleteContract 完成合同
func (s *ContractService) CompleteContract(ctx context.Context, id uint) (*ContractResponse, error) {
	contract, err := s.contractRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	if contract.Status != "active" && contract.Status != "suspended" {
		return nil, errors.New("只有激活或暂停状态的合同可以完成")
	}

	// 检查付款计划是否已全部完成
	milestones, err := s.milestoneRepo.GetByContractID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询付款计划失败: %w", err)
	}

	for _, m := range milestones {
		if m.Status != "paid" {
			return nil, fmt.Errorf("付款计划未全部完成: %s", m.Name)
		}
	}

	contract.Status = "completed"

	if err := s.contractRepo.Update(ctx, contract); err != nil {
		return nil, fmt.Errorf("完成合同失败: %w", err)
	}

	return s.GetContractByID(ctx, id)
}

// convertToResponse 转换为响应格式
func (s *ContractService) convertToResponse(ctx context.Context, contract *models.Contract) *ContractResponse {
	resp := &ContractResponse{
		ID:               contract.ID,
		ContractCode:     contract.ContractCode,
		CaseID:           contract.CaseID,
		ClientID:         contract.ClientID,
		ContractAmount:   contract.ContractAmount,
		Currency:         contract.Currency,
		BillingCycle:     contract.BillingCycle,
		PaymentTerms:     contract.PaymentTerms,
		Status:           contract.Status,
		ContractType:     contract.ContractType,
		ParentContractID: contract.ParentContractID,
		DocumentID:       contract.DocumentID,
		CreatedAt:        contract.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        contract.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if contract.StartDate != nil {
		formatted := contract.StartDate.Format("2006-01-02")
		resp.StartDate = &formatted
	}
	if contract.EndDate != nil {
		formatted := contract.EndDate.Format("2006-01-02")
		resp.EndDate = &formatted
	}
	if contract.SignedAt != nil {
		formatted := contract.SignedAt.Format("2006-01-02")
		resp.SignedAt = &formatted
	}

	// 客户信息 - 通过仓储获取
	if client, err := s.clientRepo.FindByID(ctx, contract.ClientID); err == nil && client != nil {
		resp.Client = &ClientSummary{
			ID:   client.ID,
			Name: client.Name,
		}
	}

	// 案件信息 - 通过仓储获取
	if contract.CaseID != nil {
		if case_, err := s.caseRepo.FindByID(ctx, *contract.CaseID); err == nil && case_ != nil {
			resp.Case = &CaseSummary{
				ID:    case_.ID,
				Title: case_.Title,
			}
		}
	}

	// 付款计划 - 通过仓储获取
	milestones, _ := s.milestoneRepo.GetByContractID(ctx, contract.ID)
	if len(milestones) > 0 {
		resp.Milestones = make([]*PaymentMilestoneResponse, len(milestones))
		for i, m := range milestones {
			resp.Milestones[i] = s.convertMilestoneToResponse(m)
		}
	}

	// 补充协议
	if contract.ContractType == "original" {
		supplementaryContracts, _ := s.contractRepo.GetSupplementaryContracts(ctx, contract.ID)
		if len(supplementaryContracts) > 0 {
			resp.SupplementaryContracts = make([]*ContractResponse, len(supplementaryContracts))
			for i, sc := range supplementaryContracts {
				resp.SupplementaryContracts[i] = s.convertToResponseSimple(ctx, sc)
			}
		}
	}

	return resp
}

// convertToResponseSimple 转换为响应格式（不含关联数据）
func (s *ContractService) convertToResponseSimple(ctx context.Context, contract *models.Contract) *ContractResponse {
	resp := &ContractResponse{
		ID:               contract.ID,
		ContractCode:     contract.ContractCode,
		CaseID:           contract.CaseID,
		ClientID:         contract.ClientID,
		ContractAmount:   contract.ContractAmount,
		Currency:         contract.Currency,
		BillingCycle:     contract.BillingCycle,
		PaymentTerms:     contract.PaymentTerms,
		Status:           contract.Status,
		ContractType:     contract.ContractType,
		ParentContractID: contract.ParentContractID,
		DocumentID:       contract.DocumentID,
		CreatedAt:        contract.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        contract.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if contract.StartDate != nil {
		formatted := contract.StartDate.Format("2006-01-02")
		resp.StartDate = &formatted
	}
	if contract.EndDate != nil {
		formatted := contract.EndDate.Format("2006-01-02")
		resp.EndDate = &formatted
	}
	if contract.SignedAt != nil {
		formatted := contract.SignedAt.Format("2006-01-02")
		resp.SignedAt = &formatted
	}

	// convertToResponseSimple 不包含关联数据

	return resp
}

// convertMilestoneToResponse 转换付款计划为响应格式
func (s *ContractService) convertMilestoneToResponse(milestone *models.PaymentMilestone) *PaymentMilestoneResponse {
	resp := &PaymentMilestoneResponse{
		ID:         milestone.ID,
		ContractID: milestone.ContractID,
		Name:       milestone.Name,
		Sequence:   milestone.Sequence,
		Amount:     milestone.Amount,
		Percentage: milestone.Percentage,
		Condition:  milestone.Condition,
		Status:     milestone.Status,
		InvoiceID:  milestone.InvoiceID,
		PaidAmount: milestone.PaidAmount,
	}

	if milestone.DueDate != nil {
		formatted := milestone.DueDate.Format("2006-01-02")
		resp.DueDate = &formatted
	}

	return resp
}

// GetContractStats 获取合同统计信息
func (s *ContractService) GetContractStats(ctx context.Context) (*ContractStats, error) {
	db := s.contractRepo.GetDB()

	var stats ContractStats

	// 总合同数
	if err := db.WithContext(ctx).Model(&models.Contract{}).Count(&stats.TotalContracts).Error; err != nil {
		return nil, fmt.Errorf("查询合同总数失败: %w", err)
	}

	// 按状态统计
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	err := db.WithContext(ctx).Model(&models.Contract{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, fmt.Errorf("查询状态统计失败: %w", err)
	}

	for _, sc := range statusCounts {
		switch sc.Status {
		case "draft":
			stats.DraftContracts = sc.Count
		case "active":
			stats.ActiveContracts = sc.Count
		case "suspended":
			stats.SuspendedContracts = sc.Count
		case "completed":
			stats.CompletedContracts = sc.Count
		case "cancelled":
			stats.CancelledContracts = sc.Count
		}
	}

	// 合同总金额
	var totalAmount struct {
		Total float64
	}
	err = db.WithContext(ctx).Model(&models.Contract{}).
		Select("COALESCE(SUM(contract_amount), 0) as total").
		Where("status IN ?", []string{"active", "completed"}).
		Scan(&totalAmount).Error
	if err != nil {
		return nil, fmt.Errorf("查询合同总金额失败: %w", err)
	}
	stats.TotalContractAmount = totalAmount.Total

	// 本月新增合同
	var monthCount int64
	firstOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	err = db.WithContext(ctx).Model(&models.Contract{}).
		Where("created_at >= ?", firstOfMonth).
		Count(&monthCount).Error
	if err != nil {
		return nil, fmt.Errorf("查询本月新增合同失败: %w", err)
	}
	stats.NewContractsThisMonth = monthCount

	return &stats, nil
}

// ContractStats 合同统计信息
type ContractStats struct {
	TotalContracts       int64   `json:"total_contracts"`
	DraftContracts       int64   `json:"draft_contracts"`
	ActiveContracts      int64   `json:"active_contracts"`
	SuspendedContracts   int64   `json:"suspended_contracts"`
	CompletedContracts   int64   `json:"completed_contracts"`
	CancelledContracts   int64   `json:"cancelled_contracts"`
	TotalContractAmount  float64 `json:"total_contract_amount"`
	NewContractsThisMonth int64   `json:"new_contracts_this_month"`
}

// GetContractsByClientID 获取客户的合同列表
func (s *ContractService) GetContractsByClientID(ctx context.Context, clientID uint) ([]*ContractResponse, error) {
	contracts, err := s.contractRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("查询客户合同失败: %w", err)
	}

	result := make([]*ContractResponse, len(contracts))
	for i, c := range contracts {
		result[i] = s.convertToResponseSimple(ctx, c)
	}

	return result, nil
}

// GetContractsByCaseID 获取案件的合同列表
func (s *ContractService) GetContractsByCaseID(ctx context.Context, caseID uint) ([]*ContractResponse, error) {
	contracts, err := s.contractRepo.GetByCaseID(ctx, caseID)
	if err != nil {
		return nil, fmt.Errorf("查询案件合同失败: %w", err)
	}

	result := make([]*ContractResponse, len(contracts))
	for i, c := range contracts {
		result[i] = s.convertToResponseSimple(ctx, c)
	}

	return result, nil
}

// PaymentMilestoneService 付款计划服务
type PaymentMilestoneService struct {
	milestoneRepo repositories.PaymentMilestoneRepository
	contractRepo  repositories.ContractRepository
	invoiceRepo   repositories.InvoiceRepository
}

// NewPaymentMilestoneService 创建付款计划服务实例
func NewPaymentMilestoneService(
	milestoneRepo repositories.PaymentMilestoneRepository,
	contractRepo repositories.ContractRepository,
	invoiceRepo repositories.InvoiceRepository,
) *PaymentMilestoneService {
	return &PaymentMilestoneService{
		milestoneRepo: milestoneRepo,
		contractRepo:  contractRepo,
		invoiceRepo:   invoiceRepo,
	}
}

// PaymentMilestoneCreateRequest 付款计划创建请求（独立服务用）
type PaymentMilestoneCreateRequest struct {
	ContractID uint    `json:"contract_id" binding:"required"`
	Name       string  `json:"name" binding:"required,min=1,max=200"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Percentage float64 `json:"percentage" binding:"required,gte=0,lte=100"`
	DueDate    *string `json:"due_date,omitempty"`
	Condition  string  `json:"condition" binding:"max=500"`
}

// UpdateMilestoneRequest 更新付款计划请求
type UpdateMilestoneRequest struct {
	Name       *string  `json:"name,omitempty" binding:"omitempty,min=1,max=200"`
	Amount     *float64 `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Percentage *float64 `json:"percentage,omitempty" binding:"omitempty,gte=0,lte=100"`
	DueDate    *string  `json:"due_date,omitempty"`
	Condition  *string  `json:"condition,omitempty" binding:"omitempty,max=500"`
}

// CreateMilestone 创建付款计划
func (s *PaymentMilestoneService) CreateMilestone(ctx context.Context, req *PaymentMilestoneCreateRequest) (*PaymentMilestoneResponse, error) {
	// 验证合同是否存在
	contract, err := s.contractRepo.FindByID(ctx, req.ContractID)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil {
		return nil, errors.New("合同不存在")
	}

	// 只有草稿状态的合同可以添加付款计划
	if contract.Status != "draft" {
		return nil, errors.New("只有草稿状态的合同可以添加付款计划")
	}

	// 获取现有付款计划数量
	existingMilestones, err := s.milestoneRepo.GetByContractID(ctx, req.ContractID)
	if err != nil {
		return nil, fmt.Errorf("查询付款计划失败: %w", err)
	}

	// 计算总金额
	var totalAmount float64
	for _, m := range existingMilestones {
		totalAmount += m.Amount
	}
	totalAmount += req.Amount

	if totalAmount > contract.ContractAmount {
		return nil, fmt.Errorf("付款计划总额(%.2f)不能超过合同金额(%.2f)", totalAmount, contract.ContractAmount)
	}

	// 解析日期
	var dueDate *time.Time
	if req.DueDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		dueDate = &parsedDate
	}

	milestone := &models.PaymentMilestone{
		ContractID: req.ContractID,
		Name:       req.Name,
		Sequence:   len(existingMilestones) + 1,
		Amount:     req.Amount,
		Percentage: req.Percentage,
		DueDate:    dueDate,
		Condition:  req.Condition,
		Status:     "pending",
	}

	if err := s.milestoneRepo.Create(ctx, milestone); err != nil {
		return nil, fmt.Errorf("创建付款计划失败: %w", err)
	}

	return s.convertToResponse(milestone), nil
}

// UpdateMilestone 更新付款计划
func (s *PaymentMilestoneService) UpdateMilestone(ctx context.Context, id uint, req *UpdateMilestoneRequest) (*PaymentMilestoneResponse, error) {
	milestone, err := s.milestoneRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询付款计划失败: %w", err)
	}
	if milestone == nil {
		return nil, errors.New("付款计划不存在")
	}

	// 只有pending状态的付款计划可以编辑
	if milestone.Status != "pending" {
		return nil, errors.New("只有待处理状态的付款计划可以编辑")
	}

	// 验证合同状态
	contract, err := s.contractRepo.FindByID(ctx, milestone.ContractID)
	if err != nil {
		return nil, fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil || contract.Status != "draft" {
		return nil, errors.New("只有草稿状态的合同的付款计划可以编辑")
	}

	// 更新字段
	if req.Name != nil {
		milestone.Name = *req.Name
	}
	if req.Amount != nil {
		milestone.Amount = *req.Amount
	}
	if req.Percentage != nil {
		milestone.Percentage = *req.Percentage
	}
	if req.Condition != nil {
		milestone.Condition = *req.Condition
	}
	if req.DueDate != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		milestone.DueDate = &parsedDate
	}

	if err := s.milestoneRepo.Update(ctx, milestone); err != nil {
		return nil, fmt.Errorf("更新付款计划失败: %w", err)
	}

	return s.convertToResponse(milestone), nil
}

// DeleteMilestone 删除付款计划
func (s *PaymentMilestoneService) DeleteMilestone(ctx context.Context, id uint) error {
	milestone, err := s.milestoneRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询付款计划失败: %w", err)
	}
	if milestone == nil {
		return errors.New("付款计划不存在")
	}

	// 只有pending状态的付款计划可以删除
	if milestone.Status != "pending" {
		return errors.New("只有待处理状态的付款计划可以删除")
	}

	// 验证合同状态
	contract, err := s.contractRepo.FindByID(ctx, milestone.ContractID)
	if err != nil {
		return fmt.Errorf("查询合同失败: %w", err)
	}
	if contract == nil || contract.Status != "draft" {
		return errors.New("只有草稿状态的合同的付款计划可以删除")
	}

	if err := s.milestoneRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除付款计划失败: %w", err)
	}

	// 重新排序剩余的付款计划
	milestones, _ := s.milestoneRepo.GetByContractID(ctx, milestone.ContractID)
	for i, m := range milestones {
		m.Sequence = i + 1
		s.milestoneRepo.Update(ctx, m)
	}

	return nil
}

// GetMilestonesByContractID 获取合同的付款计划列表
func (s *PaymentMilestoneService) GetMilestonesByContractID(ctx context.Context, contractID uint) ([]*PaymentMilestoneResponse, error) {
	milestones, err := s.milestoneRepo.GetByContractID(ctx, contractID)
	if err != nil {
		return nil, fmt.Errorf("查询付款计划失败: %w", err)
	}

	result := make([]*PaymentMilestoneResponse, len(milestones))
	for i, m := range milestones {
		result[i] = s.convertToResponse(m)
	}

	return result, nil
}

// convertToResponse 转换为响应格式
func (s *PaymentMilestoneService) convertToResponse(milestone *models.PaymentMilestone) *PaymentMilestoneResponse {
	resp := &PaymentMilestoneResponse{
		ID:         milestone.ID,
		ContractID: milestone.ContractID,
		Name:       milestone.Name,
		Sequence:   milestone.Sequence,
		Amount:     milestone.Amount,
		Percentage: milestone.Percentage,
		Condition:  milestone.Condition,
		Status:     milestone.Status,
		InvoiceID:  milestone.InvoiceID,
		PaidAmount: milestone.PaidAmount,
	}

	if milestone.DueDate != nil {
		formatted := milestone.DueDate.Format("2006-01-02")
		resp.DueDate = &formatted
	}

	return resp
}
