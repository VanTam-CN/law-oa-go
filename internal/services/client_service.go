package services

import (
	"context"
	"errors"
	"regexp"
	"time"

	"gorm.io/gorm"
	customErrors "law-oa-go/internal/errors"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

type ClientService struct {
	clientRepo repositories.ClientRepository
}

func NewClientService(clientRepo repositories.ClientRepository) *ClientService {
	return &ClientService{clientRepo: clientRepo}
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
	Company  string `form:"company"`
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
		existingClient, err := s.clientRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, customErrors.NewDatabaseError("check_email_existence", "Failed to check email existence", err)
		}
		if existingClient != nil {
			return nil, customErrors.NewBusinessError("email_exists", "Email already exists", nil)
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

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, customErrors.NewDatabaseError("create_client", "Failed to create client", err)
	}

	return s.toClientResponse(client), nil
}

func (s *ClientService) GetClientByID(ctx context.Context, id uint) (*ClientResponse, error) {
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, customErrors.NewDatabaseError("get_client", "Failed to get client", err)
	}
	if client == nil {
		return nil, customErrors.NewNotFoundError("client", "Client not found", nil)
	}

	return s.toClientResponse(client), nil
}

func (s *ClientService) UpdateClient(ctx context.Context, id uint, req *UpdateClientRequest) (*ClientResponse, error) {
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, customErrors.NewDatabaseError("find_client", "Failed to find client", err)
	}
	if client == nil {
		return nil, customErrors.NewNotFoundError("client", "Client not found", nil)
	}

	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.Email != nil {
		if *req.Email != client.Email && *req.Email != "" {
			existingClient, err := s.clientRepo.FindByEmail(ctx, *req.Email)
			if err != nil {
				return nil, customErrors.NewDatabaseError("check_email_existence", "Failed to check email existence", err)
			}
			if existingClient != nil && existingClient.ID != id {
				return nil, customErrors.NewBusinessError("email_exists", "Email already exists", nil)
			}
		}
		client.Email = *req.Email
	}
	if req.Phone != nil {
		client.Phone = *req.Phone
	}
	if req.Address != nil {
		client.Address = *req.Address
	}
	if req.Company != nil {
		client.Company = *req.Company
	}
	if req.Notes != nil {
		client.Notes = *req.Notes
	}
	if req.Status != nil {
		client.Status = *req.Status
	}

	if err := s.clientRepo.Update(ctx, client); err != nil {
		return nil, customErrors.NewDatabaseError("update_client", "Failed to update client", err)
	}

	return s.GetClientByID(ctx, id)
}

func (s *ClientService) DeleteClient(ctx context.Context, id uint) error {
	if err := s.clientRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customErrors.NewNotFoundError("client", "Client not found", nil)
		}
		return customErrors.NewDatabaseError("delete_client", "Failed to delete client", err)
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

	params := &repositories.ClientListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   req.Status,
		Search:   req.Search,
	}

	clients, total, err := s.clientRepo.List(ctx, params)
	if err != nil {
		return nil, 0, customErrors.NewDatabaseError("list_clients", "Failed to list clients", err)
	}

	responses := make([]*ClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = s.toClientResponse(client)
	}

	return responses, total, nil
}

func (s *ClientService) GetClientStats(ctx context.Context) (*ClientStatsResponse, error) {
	stats, err := s.clientRepo.GetStats(ctx)
	if err != nil {
		return nil, customErrors.NewDatabaseError("get_client_stats", "Failed to get client stats", err)
	}

	return &ClientStatsResponse{
		TotalClients:        stats.TotalClients,
		ActiveClients:       stats.ActiveClients,
		InactiveClients:     stats.InactiveClients,
		NewClientsThisMonth: stats.NewClientsThisMonth,
	}, nil
}

func (s *ClientService) validateClientRequest(req *CreateClientRequest) error {
	if req.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(req.Email) {
			return customErrors.NewValidationError("email", "invalid_email_format", "Invalid email format", "Please provide a valid email address")
		}
	}

	if req.Phone != "" {
		phoneRegex := regexp.MustCompile(`^[\d\s\-\+\(\)]+$`)
		if !phoneRegex.MatchString(req.Phone) {
			return customErrors.NewValidationError("phone", "invalid_phone_format", "Invalid phone format", "Please provide a valid phone number")
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
