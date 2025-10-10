package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
)

type CaseService struct {
	db                  *gorm.DB
	conflictService     ConflictService
	enableConflictCheck bool
}

func NewCaseService(db *gorm.DB, conflictService ConflictService, enableConflictCheck bool) *CaseService {
	return &CaseService{
		db:                  db,
		conflictService:     conflictService,
		enableConflictCheck: enableConflictCheck,
	}
}

type CreateCaseRequest struct {
	Title             string `json:"title" binding:"required,min=1,max=200"`
	Description       string `json:"description" binding:"max=2000"`
	ClientID          uint   `json:"client_id" binding:"required"`
	LawyerID          uint   `json:"lawyer_id" binding:"required"`
	CaseType          string `json:"case_type" binding:"required,oneof=civil criminal commercial administrative"`
	Priority          string `json:"priority" binding:"required,oneof=low medium high urgent"`
	Status            string `json:"status" binding:"omitempty,oneof=pending active closed suspended"`
	SkipConflictCheck *bool  `json:"skip_conflict_check"` // 可选：跳过冲突检查
}

type UpdateCaseRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string `json:"description" binding:"omitempty,max=2000"`
	LawyerID    *uint   `json:"lawyer_id"`
	CaseType    *string `json:"case_type" binding:"omitempty,oneof=civil criminal commercial administrative"`
	Priority    *string `json:"priority" binding:"omitempty,oneof=low medium high urgent"`
	Status      *string `json:"status" binding:"omitempty,oneof=pending active closed suspended"`
}

type CaseResponse struct {
	ID                  uint                          `json:"id"`
	CaseNumber          string                        `json:"case_number,omitempty"`
	Title               string                        `json:"title"`
	Description         string                        `json:"description"`
	ClientID            uint                          `json:"client_id"`
	ClientName          string                        `json:"client_name,omitempty"`
	LawyerID            uint                          `json:"lawyer_id"`
	LawyerName          string                        `json:"lawyer_name,omitempty"`
	CaseType            string                        `json:"case_type"`
	Priority            string                        `json:"priority"`
	Status              string                        `json:"status"`
	StartDate           *time.Time                    `json:"start_date,omitempty"`
	EndDate             *time.Time                    `json:"end_date,omitempty"`
	CreatedAt           time.Time                     `json:"created_at"`
	UpdatedAt           time.Time                     `json:"updated_at"`
	CaseAmount          *float64                      `json:"case_amount,omitempty"`
	ExpectedEndDate     *time.Time                    `json:"expected_end_date,omitempty"`
	PrincipalInfo       string                        `json:"principal_info,omitempty"`
	OpponentInfo        string                        `json:"opponent_info,omitempty"`
	Client              *models.Client                `json:"client,omitempty"`
	Lawyer              *models.User                  `json:"lawyer,omitempty"`
	ConflictCheckResult *models.ConflictCheckResponse `json:"conflict_check_result,omitempty"`
}

type CaseListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=pending active closed suspended"`
	CaseType string `form:"case_type" binding:"omitempty,oneof=civil criminal commercial administrative"`
	Priority string `form:"priority" binding:"omitempty,oneof=low medium high urgent"`
	ClientID uint   `form:"client_id"`
	LawyerID uint   `form:"lawyer_id"`
	Search   string `form:"search"`
}

type CaseStatsResponse struct {
	TotalCases     int64 `json:"total_cases"`
	PendingCases   int64 `json:"pending_cases"`
	ActiveCases    int64 `json:"active_cases"`
	ClosedCases    int64 `json:"closed_cases"`
	SuspendedCases int64 `json:"suspended_cases"`
	HighPriority   int64 `json:"high_priority"`
	UrgentCases    int64 `json:"urgent_cases"`
}

func (s *CaseService) CreateCase(ctx context.Context, req *CreateCaseRequest) (*CaseResponse, error) {
	if err := s.validateCaseRequest(ctx, req); err != nil {
		return nil, err
	}

	// 执行冲突检测（如果需要）
	conflictResult, err := s.performConflictCheck(ctx, req)
	if err != nil {
		fmt.Printf("Conflict check failed: %v\n", err)
	}

	// 创建案件模型
	caseModel := s.buildCaseModel(req, conflictResult)

	// 保存案件
	if err := s.saveCase(ctx, caseModel); err != nil {
		return nil, err
	}

	// 获取完整的案件信息并添加冲突检测结果
	return s.getCaseWithConflictResult(context.Background(), caseModel.ID, conflictResult)
}

// performConflictCheck 执行冲突检测
func (s *CaseService) performConflictCheck(ctx context.Context, req *CreateCaseRequest) (*models.ConflictCheckResponse, error) {
	if !s.enableConflictCheck || (req.SkipConflictCheck != nil && *req.SkipConflictCheck) {
		return nil, nil
	}

	// 获取客户信息
	client, err := s.getClientForConflictCheck(ctx, req.ClientID)
	if err != nil {
		return nil, err
	}

	// 构建冲突检测请求
	conflictRequest := s.buildConflictCheckRequest(req, client)

	// 执行冲突检测
	return s.conflictService.CheckConflict(ctx, conflictRequest)
}

// getClientForConflictCheck 获取客户信息用于冲突检测
func (s *CaseService) getClientForConflictCheck(ctx context.Context, clientID uint) (*models.Client, error) {
	var client models.Client
	if err := s.db.WithContext(ctx).First(&client, clientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, fmt.Errorf("failed to get client for conflict check: %w", err)
	}
	return &client, nil
}

// buildConflictCheckRequest 构建冲突检测请求
func (s *CaseService) buildConflictCheckRequest(req *CreateCaseRequest, client *models.Client) *models.ConflictCheckRequest {
	return &models.ConflictCheckRequest{
		ClientID:                  fmt.Sprintf("%d", req.ClientID),
		ClientName:                client.Name,
		CaseName:                  req.Title,
		CaseType:                  req.CaseType,
		ClientType:                "individual", // 默认值，根据实际情况调整
		UserID:                    0,            // TODO: 从上下文获取用户ID
		RequestTime:               time.Now(),
		SearchYears:               5, // 默认搜索5年
		SearchDepth:               "deep",
		IncludeCorporateRelations: true,
	}
}

// buildCaseModel 构建案件模型
func (s *CaseService) buildCaseModel(req *CreateCaseRequest, conflictResult *models.ConflictCheckResponse) *models.Case {
	caseModel := &models.Case{
		Title:       req.Title,
		Description: req.Description,
		ClientID:    req.ClientID,
		LawyerID:    req.LawyerID,
		CaseType:    req.CaseType,
		Priority:    req.Priority,
		Status:      "pending",
	}

	// 使用请求中的状态
	if req.Status != "" {
		caseModel.Status = req.Status
	}

	// 根据冲突检测结果调整状态
	s.adjustCaseStatusByConflict(caseModel, conflictResult)

	return caseModel
}

// adjustCaseStatusByConflict 根据冲突检测结果调整案件状态
func (s *CaseService) adjustCaseStatusByConflict(caseModel *models.Case, conflictResult *models.ConflictCheckResponse) {
	if conflictResult != nil && conflictResult.HasConflict {
		if conflictResult.RiskAssessment != nil &&
			(conflictResult.RiskAssessment.OverallRisk == "HIGH" || conflictResult.RiskAssessment.RequiresApproval) {
			caseModel.Status = "pending" // 强制设置为待审核状态
		}
	}
}

// saveCase 保存案件
func (s *CaseService) saveCase(ctx context.Context, caseModel *models.Case) error {
	if err := s.db.WithContext(ctx).Create(caseModel).Error; err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}
	return nil
}

// getCaseWithConflictResult 获取案件信息并添加冲突检测结果
func (s *CaseService) getCaseWithConflictResult(ctx context.Context, caseID uint, conflictResult *models.ConflictCheckResponse) (*CaseResponse, error) {
	caseResponse, err := s.GetCaseByID(ctx, caseID)
	if err != nil {
		return nil, err
	}

	// 添加冲突检测结果
	if conflictResult != nil {
		caseResponse.ConflictCheckResult = conflictResult
	}

	return caseResponse, nil
}

func (s *CaseService) GetCaseByID(ctx context.Context, id uint) (*CaseResponse, error) {
	var caseModel models.Case
	err := s.db.WithContext(ctx).Preload("Client").Preload("Lawyer").First(&caseModel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFoundError("case")
		}
		return nil, fmt.Errorf("failed to get case: %w", err)
	}

	return s.toCaseResponse(&caseModel), nil
}

func (s *CaseService) UpdateCase(ctx context.Context, id uint, req *UpdateCaseRequest) (*CaseResponse, error) {
	var caseModel models.Case
	if err := s.db.WithContext(ctx).First(&caseModel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("case not found")
		}
		return nil, fmt.Errorf("failed to find case: %w", err)
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.LawyerID != nil {
		if err := s.validateLawyerExists(ctx, *req.LawyerID); err != nil {
			return nil, err
		}
		updates["lawyer_id"] = *req.LawyerID
	}
	if req.CaseType != nil {
		updates["case_type"] = *req.CaseType
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&caseModel).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update case: %w", err)
		}
	}

	return s.GetCaseByID(context.Background(), id)
}

func (s *CaseService) DeleteCase(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&models.Case{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete case: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("case not found")
	}
	return nil
}

func (s *CaseService) ListCases(ctx context.Context, req *CaseListRequest) ([]*CaseResponse, int64, error) {
	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	query := s.db.WithContext(ctx).Model(&models.Case{}).Preload("Client").Preload("Lawyer")

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.CaseType != "" {
		query = query.Where("case_type = ?", req.CaseType)
	}
	if req.Priority != "" {
		query = query.Where("priority = ?", req.Priority)
	}
	if req.ClientID > 0 {
		query = query.Where("client_id = ?", req.ClientID)
	}
	if req.LawyerID > 0 {
		query = query.Where("lawyer_id = ?", req.LawyerID)
	}
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count cases: %w", err)
	}

	var cases []models.Case
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&cases).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list cases: %w", err)
	}

	responses := make([]*CaseResponse, len(cases))
	for i, caseModel := range cases {
		responses[i] = s.toCaseResponse(&caseModel)
	}

	return responses, total, nil
}

func (s *CaseService) GetCaseStats(ctx context.Context) (*CaseStatsResponse, error) {
	stats := &CaseStatsResponse{}

	// 优化：使用单次查询获取所有统计数据
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	type PriorityCount struct {
		Priority string `json:"priority"`
		Count    int64  `json:"count"`
	}

	// 获取总数
	if err := s.db.WithContext(ctx).Model(&models.Case{}).Count(&stats.TotalCases).Error; err != nil {
		return nil, common.NewDatabaseError("count total cases", err)
	}

	// 按状态统计（单次查询）
	var statusCounts []StatusCount
	err := s.db.WithContext(ctx).Model(&models.Case{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&statusCounts).Error
	if err != nil {
		return nil, common.NewDatabaseError("count cases by status", err)
	}

	// 按优先级统计（单次查询）
	var priorityCounts []PriorityCount
	err = s.db.WithContext(ctx).Model(&models.Case{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Find(&priorityCounts).Error
	if err != nil {
		return nil, common.NewDatabaseError("count cases by priority", err)
	}

	// 填充统计数据
	for _, sc := range statusCounts {
		switch sc.Status {
		case "pending":
			stats.PendingCases = sc.Count
		case "active":
			stats.ActiveCases = sc.Count
		case "closed":
			stats.ClosedCases = sc.Count
		case "suspended":
			stats.SuspendedCases = sc.Count
		}
	}

	for _, pc := range priorityCounts {
		switch pc.Priority {
		case "high":
			stats.HighPriority = pc.Count
		case "urgent":
			stats.UrgentCases = pc.Count
		}
	}

	return stats, nil
}

func (s *CaseService) AssignLawyer(ctx context.Context, caseID, lawyerID uint) error {
	if err := s.validateLawyerExists(ctx, lawyerID); err != nil {
		return err
	}

	result := s.db.WithContext(ctx).Model(&models.Case{}).Where("id = ?", caseID).Update("lawyer_id", lawyerID)
	if result.Error != nil {
		return fmt.Errorf("failed to assign lawyer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("case not found")
	}
	return nil
}

func (s *CaseService) UpdateCaseStatus(ctx context.Context, caseID uint, status string) error {
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"closed":    true,
		"suspended": true,
	}

	if !validStatuses[status] {
		return errors.New("invalid case status")
	}

	result := s.db.WithContext(ctx).Model(&models.Case{}).Where("id = ?", caseID).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update case status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("case not found")
	}
	return nil
}

func (s *CaseService) validateCaseRequest(ctx context.Context, req *CreateCaseRequest) error {
	var client models.Client
	if err := s.db.WithContext(ctx).First(&client, req.ClientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("client not found")
		}
		return fmt.Errorf("failed to validate client: %w", err)
	}

	return s.validateLawyerExists(ctx, req.LawyerID)
}

func (s *CaseService) validateLawyerExists(ctx context.Context, lawyerID uint) error {
	var user models.User
	err := s.db.WithContext(ctx).Where("id = ? AND role = ?", lawyerID, "lawyer").First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("lawyer not found")
		}
		return fmt.Errorf("failed to validate lawyer: %w", err)
	}
	return nil
}

func (s *CaseService) toCaseResponse(caseModel *models.Case) *CaseResponse {
	response := &CaseResponse{
		ID:          caseModel.ID,
		Title:       caseModel.Title,
		Description: caseModel.Description,
		ClientID:    caseModel.ClientID,
		LawyerID:    caseModel.LawyerID,
		CaseType:    caseModel.CaseType,
		Priority:    caseModel.Priority,
		Status:      caseModel.Status,
		StartDate:   caseModel.StartDate,
		EndDate:     caseModel.EndDate,
		CreatedAt:   caseModel.CreatedAt,
		UpdatedAt:   caseModel.UpdatedAt,
	}

	// 处理客户信息
	if caseModel.Client != nil {
		// 创建一个简化的客户对象用于序列化
		response.Client = &models.Client{
			ID:       caseModel.Client.ID,
			Name:     caseModel.Client.ClientName, // 使用ClientName字段
			Email:    caseModel.Client.Email,
			Phone:    caseModel.Client.Phone,
			Address:  caseModel.Client.Address,
			Company:  caseModel.Client.Company,
			Status:   caseModel.Client.Status,
		}
		// 优先使用company字段，如果为空则使用name字段
		if caseModel.Client.Company != "" {
			response.ClientName = caseModel.Client.Company
		} else {
			response.ClientName = caseModel.Client.ClientName
		}
	}

	// 处理律师信息
	if caseModel.Lawyer != nil {
		response.LawyerName = caseModel.Lawyer.Name
		// 包含完整的律师对象
		response.Lawyer = caseModel.Lawyer
	}

	return response
}
