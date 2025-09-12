package services

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
)

type ClientService struct {
	db *gorm.DB
}

func NewClientService(db *gorm.DB) *ClientService {
	return &ClientService{db: db}
}

type CreateClientRequest struct {
	Name    string `json:"name" binding:"required,min=1,max=100"`
	Email   string `json:"email" binding:"omitempty,email"`
	Phone   string `json:"phone" binding:"omitempty,max=20"`
	Address string `json:"address" binding:"omitempty,max=255"`
	Company string `json:"company" binding:"omitempty,max=100"`
	Notes   string `json:"notes" binding:"omitempty,max=1000"`
}

type UpdateClientRequest struct {
	Name    *string `json:"name" binding:"omitempty,min=1,max=100"`
	Email   *string `json:"email" binding:"omitempty,email"`
	Phone   *string `json:"phone" binding:"omitempty,max=20"`
	Address *string `json:"address" binding:"omitempty,max=255"`
	Company *string `json:"company" binding:"omitempty,max=100"`
	Notes   *string `json:"notes" binding:"omitempty,max=1000"`
	Status  *string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type ClientResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Company   string    `json:"company"`
	Notes     string    `json:"notes"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClientListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive"`
	Search   string `form:"search"`
}

type ClientStatsResponse struct {
	TotalClients        int64 `json:"total_clients"`
	ActiveClients       int64 `json:"active_clients"`
	InactiveClients     int64 `json:"inactive_clients"`
	NewClientsThisMonth int64 `json:"new_clients_this_month"`
}

func (s *ClientService) CreateClient(ctx context.Context, req *CreateClientRequest) (*ClientResponse, error) {
	if err := s.validateClientRequest(req); err != nil {
		return nil, err
	}

	if req.Email != "" {
		var existingClient models.Client
		if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingClient).Error; err == nil {
			return nil, common.NewValidationError("email already exists", "The email address is already registered")
		}
	}

	client := &models.Client{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		Company: req.Company,
		Notes:   req.Notes,
		Status:  "active",
	}

	if err := s.db.WithContext(ctx).Create(client).Error; err != nil {
		return nil, common.NewDatabaseError("create client", err)
	}

	return s.toClientResponse(client), nil
}

func (s *ClientService) GetClientByID(ctx context.Context, id uint) (*ClientResponse, error) {
	var client models.Client
	err := s.db.WithContext(ctx).First(&client, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFoundError("client")
		}
		return nil, common.NewDatabaseError("get client", err)
	}

	return s.toClientResponse(&client), nil
}

func (s *ClientService) UpdateClient(ctx context.Context, id uint, req *UpdateClientRequest) (*ClientResponse, error) {
	var client models.Client
	if err := s.db.WithContext(ctx).First(&client, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewNotFoundError("client")
		}
		return nil, common.NewDatabaseError("find client", err)
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		if *req.Email != client.Email && *req.Email != "" {
			var existingClient models.Client
			if err := s.db.WithContext(ctx).Where("email = ? AND id != ?", *req.Email, id).First(&existingClient).Error; err == nil {
				return nil, common.NewValidationError("email already exists", "The email address is already in use")
			}
		}
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Company != nil {
		updates["company"] = *req.Company
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&client).Updates(updates).Error; err != nil {
			return nil, common.NewDatabaseError("update client", err)
		}
	}

	return s.GetClientByID(ctx, id)
}

func (s *ClientService) DeleteClient(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Delete(&models.Client{}, id)
	if result.Error != nil {
		return common.NewDatabaseError("delete client", result.Error)
	}
	if result.RowsAffected == 0 {
		return common.NewNotFoundError("client")
	}
	return nil
}

func (s *ClientService) ListClients(ctx context.Context, req *ClientListRequest) ([]*ClientResponse, int64, error) {
	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	query := s.db.WithContext(ctx).Model(&models.Client{})

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone) LIKE ?", searchTerm, searchTerm, searchTerm)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, common.NewDatabaseError("count clients", err)
	}

	var clients []models.Client
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&clients).Error; err != nil {
		return nil, 0, common.NewDatabaseError("list clients", err)
	}

	responses := make([]*ClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = s.toClientResponse(&client)
	}

	return responses, total, nil
}

func (s *ClientService) GetClientStats(ctx context.Context) (*ClientStatsResponse, error) {
	stats := &ClientStatsResponse{}

	if err := s.db.WithContext(ctx).Model(&models.Client{}).Count(&stats.TotalClients).Error; err != nil {
		return nil, common.NewDatabaseError("count total clients", err)
	}

	if err := s.db.WithContext(ctx).Model(&models.Client{}).Where("status = ?", "active").Count(&stats.ActiveClients).Error; err != nil {
		return nil, common.NewDatabaseError("count active clients", err)
	}

	if err := s.db.WithContext(ctx).Model(&models.Client{}).Where("status = ?", "inactive").Count(&stats.InactiveClients).Error; err != nil {
		return nil, common.NewDatabaseError("count inactive clients", err)
	}

	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1).Truncate(24 * time.Hour)
	if err := s.db.WithContext(ctx).Model(&models.Client{}).Where("created_at >= ?", startOfMonth).Count(&stats.NewClientsThisMonth).Error; err != nil {
		return nil, common.NewDatabaseError("count new clients this month", err)
	}

	return stats, nil
}

func (s *ClientService) validateClientRequest(req *CreateClientRequest) error {
	if req.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(req.Email) {
			return common.NewValidationError("invalid email format", "Please provide a valid email address")
		}
	}

	if req.Phone != "" {
		phoneRegex := regexp.MustCompile(`^[\d\s\-\+\(\)]+$`)
		if !phoneRegex.MatchString(req.Phone) {
			return common.NewValidationError("invalid phone format", "Please provide a valid phone number")
		}
	}

	return nil
}

func (s *ClientService) toClientResponse(client *models.Client) *ClientResponse {
	return &ClientResponse{
		ID:        client.ID,
		Name:      client.Name,
		Email:     client.Email,
		Phone:     client.Phone,
		Address:   client.Address,
		Company:   client.Company,
		Notes:     client.Notes,
		Status:    client.Status,
		CreatedAt: client.CreatedAt,
		UpdatedAt: client.UpdatedAt,
	}
}
