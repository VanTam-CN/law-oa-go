package services

import (
	"context"
	"regexp"
	"time"

	stderrors "errors"
	"gorm.io/gorm"
	"law-oa-go/internal/errors"
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
	Name          string `json:"name" binding:"required,min=1,max=100"`
	Type          string `json:"type" binding:"required,oneof=个人 企业"`
	Email         string `json:"email" binding:"omitempty,email"`
	Phone         string `json:"phone" binding:"omitempty,max=20"`
	Address       string `json:"address" binding:"omitempty,max=255"`
	IDCard        string `json:"id_card" binding:"omitempty,max=18"`
	Company       string `json:"company" binding:"omitempty,max=100"`
	Industry      string `json:"industry" binding:"omitempty,max=50"`
	ContactPerson string `json:"contact_person" binding:"omitempty,max=50"`
	ContactPhone  string `json:"contact_phone" binding:"omitempty,max=20"`
	Source        string `json:"source" binding:"omitempty,max=50"`
	Notes         string `json:"notes" binding:"omitempty,max=1000"`
}

type UpdateClientRequest struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=100"`
	Type          *string `json:"type" binding:"omitempty,oneof=个人 企业"`
	Email         *string `json:"email" binding:"omitempty,email"`
	Phone         *string `json:"phone" binding:"omitempty,max=20"`
	Address       *string `json:"address" binding:"omitempty,max=255"`
	IDCard        *string `json:"id_card" binding:"omitempty,max=18"`
	Company       *string `json:"company" binding:"omitempty,max=100"`
	Industry      *string `json:"industry" binding:"omitempty,max=50"`
	ContactPerson *string `json:"contact_person" binding:"omitempty,max=50"`
	ContactPhone  *string `json:"contact_phone" binding:"omitempty,max=20"`
	Source        *string `json:"source" binding:"omitempty,max=50"`
	Notes         *string `json:"notes" binding:"omitempty,max=1000"`
	Status        *string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type ClientResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address"`
	IDCard        string    `json:"id_card"`
	Company       string    `json:"company"`
	Industry      string    `json:"industry"`
	ContactPerson string    `json:"contact_person"`
	ContactPhone  string    `json:"contact_phone"`
	Source        string    `json:"source"`
	Notes         string    `json:"notes"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ClientListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Name     string `form:"name" binding:"omitempty,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=个人 企业"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive"`
	Search   string `form:"search"`
	Company  string `form:"company"`
}

type ClientStatsResponse struct {
	Total          int64             `json:"total"`
	ActiveClients  int64             `json:"active_clients"`
	InactiveClients int64             `json:"inactive_clients"`
	MonthlyNew     int64             `json:"monthly_new"`
	TypeStats      map[string]int64  `json:"type_stats"`
	StatusStats    map[string]int64  `json:"status_stats"`
	SourceStats    map[string]int64  `json:"source_stats"`
}

func (s *ClientService) CreateClient(ctx context.Context, req *CreateClientRequest) (*ClientResponse, error) {
	if err := s.validateClientRequest(req); err != nil {
		return nil, err
	}

	if req.Email != "" {
		existingClient, err := s.clientRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, errors.DatabaseError("check_email_existence", "Failed to check email existence", err)
		}
		if existingClient != nil {
			return nil, errors.BusinessError("email", "exists", "Email already exists")
		}
	}

	client := &models.Client{
		Name:          req.Name,
		Type:          req.Type,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		IDCard:        req.IDCard,
		Company:       req.Company,
		Industry:      req.Industry,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		Source:        req.Source,
		Notes:         req.Notes,
		Status:        "active",
	}

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, errors.DatabaseError("create_client", "Failed to create client", err)
	}

	return s.toClientResponse(client), nil
}

func (s *ClientService) GetClientByID(ctx context.Context, id uint) (*ClientResponse, error) {
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.DatabaseError("get_client", "Failed to get client", err)
	}
	if client == nil {
		return nil, errors.NotFoundError("client", "Client not found", id)
	}

	return s.toClientResponse(client), nil
}

func (s *ClientService) UpdateClient(ctx context.Context, id uint, req *UpdateClientRequest) (*ClientResponse, error) {
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.DatabaseError("find_client", "Failed to find client", err)
	}
	if client == nil {
		return nil, errors.NotFoundError("client", "Client not found", id)
	}

	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.Type != nil {
		client.Type = *req.Type
	}
	if req.Email != nil {
		if *req.Email != client.Email && *req.Email != "" {
			existingClient, err := s.clientRepo.FindByEmail(ctx, *req.Email)
			if err != nil {
				return nil, errors.DatabaseError("check_email_existence", "Failed to check email existence", err)
			}
			if existingClient != nil && existingClient.ID != id {
				return nil, errors.BusinessError("email", "exists", "Email already exists")
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
	if req.IDCard != nil {
		client.IDCard = *req.IDCard
	}
	if req.Company != nil {
		client.Company = *req.Company
	}
	if req.Industry != nil {
		client.Industry = *req.Industry
	}
	if req.ContactPerson != nil {
		client.ContactPerson = *req.ContactPerson
	}
	if req.ContactPhone != nil {
		client.ContactPhone = *req.ContactPhone
	}
	if req.Source != nil {
		client.Source = *req.Source
	}
	if req.Notes != nil {
		client.Notes = *req.Notes
	}
	if req.Status != nil {
		client.Status = *req.Status
	}

	if err := s.clientRepo.Update(ctx, client); err != nil {
		return nil, errors.DatabaseError("update_client", "Failed to update client", err)
	}

	return s.GetClientByID(ctx, id)
}

func (s *ClientService) DeleteClient(ctx context.Context, id uint) error {
	if err := s.clientRepo.Delete(ctx, id); err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundError("client", "Client not found", id)
		}
		return errors.DatabaseError("delete_client", "Failed to delete client", err)
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

	// 如果有name参数但没有search参数，则使用name作为搜索条件
	searchTerm := req.Search
	if searchTerm == "" && req.Name != "" {
		searchTerm = req.Name
	}

	params := &repositories.ClientListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   req.Status,
		Search:   searchTerm,
		Type:     req.Type,
		Company:  req.Company,
	}

	clients, total, err := s.clientRepo.List(ctx, params)
	if err != nil {
		return nil, 0, errors.DatabaseError("list_clients", "Failed to list clients", err)
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
		return nil, errors.DatabaseError("get_client_stats", "Failed to get client stats", err)
	}

	return &ClientStatsResponse{
		Total:          stats.TotalClients,
		ActiveClients:  stats.ActiveClients,
		InactiveClients: stats.InactiveClients,
		MonthlyNew:     stats.NewClientsThisMonth,
	}, nil
}

func (s *ClientService) validateClientRequest(req *CreateClientRequest) error {
	if req.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(req.Email) {
			return errors.ValidationErrorWithDetails("email", "Invalid email format", "Please provide a valid email address", []string{"must be valid email", "cannot contain special chars"})
		}
	}

	if req.Phone != "" {
		phoneRegex := regexp.MustCompile(`^[\d\s\-\+\(\)]+$`)
		if !phoneRegex.MatchString(req.Phone) {
			return errors.ValidationErrorWithDetails("phone", "Invalid phone format", "Please provide a valid phone number", []string{"must be valid phone number", "only digits, spaces, and symbols allowed"})
		}
	}

	return nil
}

func (s *ClientService) toClientResponse(client *models.Client) *ClientResponse {
	return &ClientResponse{
		ID:            client.ID,
		Name:          client.Name,
		Type:          client.Type,
		Email:         client.Email,
		Phone:         client.Phone,
		Address:       client.Address,
		IDCard:        client.IDCard,
		Company:       client.Company,
		Industry:      client.Industry,
		ContactPerson: client.ContactPerson,
		ContactPhone:  client.ContactPhone,
		Source:        client.Source,
		Notes:         client.Notes,
		Status:        client.Status,
		CreatedAt:     client.CreatedAt,
		UpdatedAt:     client.UpdatedAt,
	}
}
