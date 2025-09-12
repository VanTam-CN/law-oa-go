package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/common"
	"gorm.io/gorm"
)

type CaseService struct {
	db *gorm.DB
}

func NewCaseService(db *gorm.DB) *CaseService {
	return &CaseService{db: db}
}

type CreateCaseRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=2000"`
	ClientID    uint   `json:"client_id" binding:"required"`
	LawyerID    uint   `json:"lawyer_id" binding:"required"`
	CaseType    string `json:"case_type" binding:"required,oneof=civil criminal commercial administrative"`
	Priority    string `json:"priority" binding:"required,oneof=low medium high urgent"`
	Status      string `json:"status" binding:"omitempty,oneof=pending active closed suspended"`
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
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ClientID    uint      `json:"client_id"`
	ClientName  string    `json:"client_name"`
	LawyerID    uint      `json:"lawyer_id"`
	LawyerName  string    `json:"lawyer_name"`
	CaseType    string    `json:"case_type"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	if err := s.validateCaseRequest(req); err != nil {
		return nil, err
	}

	caseModel := &models.Case{
		Title:       req.Title,
		Description: req.Description,
		ClientID:    req.ClientID,
		LawyerID:    req.LawyerID,
		CaseType:    req.CaseType,
		Priority:    req.Priority,
		Status:      "pending",
	}

	if req.Status != "" {
		caseModel.Status = req.Status
	}

	if err := s.db.WithContext(ctx).Create(caseModel).Error; err != nil {
		return nil, fmt.Errorf("failed to create case: %w", err)
	}

	return s.GetCaseByID(context.Background(), caseModel.ID)
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

func (s *CaseService) UpdateCase(id uint, req *UpdateCaseRequest) (*CaseResponse, error) {
	var caseModel models.Case
	if err := s.db.First(&caseModel, id).Error; err != nil {
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
		if err := s.validateLawyerExists(*req.LawyerID); err != nil {
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
		if err := s.db.Model(&caseModel).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update case: %w", err)
		}
	}

	return s.GetCaseByID(context.Background(), id)
}

func (s *CaseService) DeleteCase(id uint) error {
	result := s.db.Delete(&models.Case{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete case: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("case not found")
	}
	return nil
}

func (s *CaseService) ListCases(req *CaseListRequest) ([]*CaseResponse, int64, error) {
	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	query := s.db.Model(&models.Case{}).Preload("Client").Preload("Lawyer")

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

func (s *CaseService) AssignLawyer(caseID, lawyerID uint) error {
	if err := s.validateLawyerExists(lawyerID); err != nil {
		return err
	}

	result := s.db.Model(&models.Case{}).Where("id = ?", caseID).Update("lawyer_id", lawyerID)
	if result.Error != nil {
		return fmt.Errorf("failed to assign lawyer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("case not found")
	}
	return nil
}

func (s *CaseService) UpdateCaseStatus(caseID uint, status string) error {
	validStatuses := map[string]bool{
		"pending":   true,
		"active":    true,
		"closed":    true,
		"suspended": true,
	}

	if !validStatuses[status] {
		return errors.New("invalid case status")
	}

	result := s.db.Model(&models.Case{}).Where("id = ?", caseID).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update case status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("case not found")
	}
	return nil
}

func (s *CaseService) validateCaseRequest(req *CreateCaseRequest) error {
	var client models.Client
	if err := s.db.First(&client, req.ClientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("client not found")
		}
		return fmt.Errorf("failed to validate client: %w", err)
	}

	return s.validateLawyerExists(req.LawyerID)
}

func (s *CaseService) validateLawyerExists(lawyerID uint) error {
	var user models.User
	err := s.db.Where("id = ? AND role = ?", lawyerID, "lawyer").First(&user).Error
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
		CreatedAt:   caseModel.CreatedAt,
		UpdatedAt:   caseModel.UpdatedAt,
	}

	if caseModel.Client != nil {
		response.ClientName = caseModel.Client.Name
	}

	if caseModel.Lawyer != nil {
		response.LawyerName = caseModel.Lawyer.Name
	}

	return response
}