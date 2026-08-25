package services

import (
	"context"
	stderrors "errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"
)

type ClientService struct {
	clientRepo repositories.ClientRepository
}

var ErrClientVersionConflict = stderrors.New("client version conflict")

func NewClientService(clientRepo repositories.ClientRepository) *ClientService {
	return &ClientService{clientRepo: clientRepo}
}

type CreateClientRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=100"`
	Type           string `json:"type" binding:"required,oneof=个人 企业"`
	Email          string `json:"email" binding:"omitempty,email"`
	Phone          string `json:"phone" binding:"omitempty,max=20"`
	Address        string `json:"address" binding:"omitempty,max=255"`
	IDCard         string `json:"id_card" binding:"omitempty,max=18"`
	IdentityType   string `json:"identity_type" binding:"omitempty,max=30"`
	IdentityNumber string `json:"identity_number" binding:"omitempty,max=100"`
	Aliases        string `json:"aliases" binding:"omitempty,max=1000"`
	CreatedBy      uint   `json:"-"`
	Company        string `json:"company" binding:"omitempty,max=100"`
	Industry       string `json:"industry" binding:"omitempty,max=50"`
	ContactPerson  string `json:"contact_person" binding:"omitempty,max=50"`
	ContactPhone   string `json:"contact_phone" binding:"omitempty,max=20"`
	Source         string `json:"source" binding:"omitempty,max=50"`
	Notes          string `json:"notes" binding:"omitempty,max=1000"`
}

type UpdateClientRequest struct {
	Version        *uint   `json:"version" binding:"required"`
	Name           *string `json:"name" binding:"omitempty,min=1,max=100"`
	Type           *string `json:"type" binding:"omitempty,oneof=个人 企业"`
	Email          *string `json:"email" binding:"omitempty,email"`
	Phone          *string `json:"phone" binding:"omitempty,max=20"`
	Address        *string `json:"address" binding:"omitempty,max=255"`
	IDCard         *string `json:"id_card" binding:"omitempty,max=18"`
	IdentityType   *string `json:"identity_type" binding:"omitempty,max=30"`
	IdentityNumber *string `json:"identity_number" binding:"omitempty,max=100"`
	Aliases        *string `json:"aliases" binding:"omitempty,max=1000"`
	Company        *string `json:"company" binding:"omitempty,max=100"`
	Industry       *string `json:"industry" binding:"omitempty,max=50"`
	ContactPerson  *string `json:"contact_person" binding:"omitempty,max=50"`
	ContactPhone   *string `json:"contact_phone" binding:"omitempty,max=20"`
	Source         *string `json:"source" binding:"omitempty,max=50"`
	Notes          *string `json:"notes" binding:"omitempty,max=1000"`
	Status         *string `json:"status" binding:"omitempty,oneof=active inactive"`
}

type ClientResponse struct {
	ID                uint                      `json:"id"`
	Version           uint                      `json:"version"`
	Name              string                    `json:"name"`
	Type              string                    `json:"type"`
	Email             string                    `json:"email"`
	Phone             string                    `json:"phone"`
	Address           string                    `json:"address"`
	IDCard            string                    `json:"id_card"`
	IdentityType      models.IdentityType       `json:"identity_type"`
	IdentityNumber    string                    `json:"identity_number"`
	IdentityStatus    string                    `json:"identity_status"`
	Aliases           string                    `json:"aliases"`
	Company           string                    `json:"company"`
	Industry          string                    `json:"industry"`
	ContactPerson     string                    `json:"contact_person"`
	ContactPhone      string                    `json:"contact_phone"`
	Source            string                    `json:"source"`
	Notes             string                    `json:"notes"`
	Status            string                    `json:"status"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	Completeness      ClientCompleteness        `json:"completeness"`
	RelatedParties    []RelatedPartySummary     `json:"related_parties"`
	HistoricalMatters []HistoricalMatterSummary `json:"historical_matters"`
}

type ClientCompleteness struct {
	Score                 int      `json:"score"`
	MissingFields         []string `json:"missing_fields"`
	ReadyForConflictCheck bool     `json:"ready_for_conflict_check"`
}

type RelatedPartySummary struct {
	Name             string `json:"name"`
	RelationshipType string `json:"relationship_type"`
	RiskImpact       string `json:"risk_impact"`
}

type HistoricalMatterSummary struct {
	CaseID     uint   `json:"case_id"`
	CaseNumber string `json:"case_number"`
	Title      string `json:"title"`
	Status     string `json:"status"`
}

type ClientListRequest struct {
	Page               int    `form:"page" binding:"omitempty,min=1"`
	PageSize           int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Name               string `form:"name" binding:"omitempty,max=100"`
	Type               string `form:"type" binding:"omitempty,oneof=个人 企业"`
	Status             string `form:"status" binding:"omitempty,oneof=active inactive"`
	Search             string `form:"search"`
	Company            string `form:"company"`
	AccessibleByUserID uint   `form:"-"`
	EthicalWallUserID  uint   `form:"-"`
}

type ClientStatsResponse struct {
	Total           int64            `json:"total"`
	ActiveClients   int64            `json:"active_clients"`
	InactiveClients int64            `json:"inactive_clients"`
	MonthlyNew      int64            `json:"monthly_new"`
	TypeStats       map[string]int64 `json:"type_stats"`
	StatusStats     map[string]int64 `json:"status_stats"`
	SourceStats     map[string]int64 `json:"source_stats"`
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

	identityNumber := strings.TrimSpace(req.IdentityNumber)
	if identityNumber == "" {
		identityNumber = strings.TrimSpace(req.IDCard)
	}
	identityType := normalizedClientIdentityType(req.Type, req.IdentityType)
	client := &models.Client{
		Name:           req.Name,
		Type:           req.Type,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		IdentityType:   identityType,
		IdentityNumber: identityNumber,
		Aliases:        strings.TrimSpace(req.Aliases),
		CreatedBy:      req.CreatedBy,
		Company:        req.Company,
		Industry:       req.Industry,
		ContactPerson:  req.ContactPerson,
		ContactPhone:   req.ContactPhone,
		Source:         req.Source,
		Notes:          req.Notes,
		Status:         "active",
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
	if req.Version == nil || *req.Version == 0 {
		return nil, errors.ValidationError("version", "Client version is required")
	}

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
	if req.IdentityType != nil {
		client.IdentityType = normalizedClientIdentityType(client.Type, *req.IdentityType)
	}
	if req.IdentityNumber != nil || req.IDCard != nil {
		identityNumber := ""
		if req.IdentityNumber != nil {
			identityNumber = strings.TrimSpace(*req.IdentityNumber)
		} else if req.IDCard != nil {
			identityNumber = strings.TrimSpace(*req.IDCard)
		}
		client.IdentityNumber = identityNumber
		client.IDCard = ""
		if identityNumber == "" {
			client.IdentityNumberDigest = ""
			client.IdentityNumberCiphertext = ""
			client.IDCardDigest = ""
			client.IDCardCiphertext = ""
		}
	}
	if req.Aliases != nil {
		client.Aliases = strings.TrimSpace(*req.Aliases)
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
	if client.HasIdentity() || strings.TrimSpace(client.IdentityNumber) != "" {
		if !validClientIdentityType(client.Type, client.EffectiveIdentityType()) {
			return nil, errors.ValidationError("identity_type", "identity type does not match client type")
		}
	}

	if err := s.clientRepo.UpdateWithVersion(ctx, client, *req.Version); err != nil {
		if stderrors.Is(err, repositories.ErrClientVersionConflict) {
			return nil, ErrClientVersionConflict
		}
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
		Page:               page,
		PageSize:           pageSize,
		Status:             req.Status,
		Search:             searchTerm,
		Type:               req.Type,
		Company:            req.Company,
		AccessibleByUserID: req.AccessibleByUserID,
		EthicalWallUserID:  req.EthicalWallUserID,
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
		Total:           stats.TotalClients,
		ActiveClients:   stats.ActiveClients,
		InactiveClients: stats.InactiveClients,
		MonthlyNew:      stats.NewClientsThisMonth,
	}, nil
}

func (s *ClientService) validateClientRequest(req *CreateClientRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return errors.ValidationError("name", "name is required")
	}
	if len([]rune(req.Name)) > 100 {
		return errors.ValidationError("name", "name is too long")
	}
	if req.Type != "个人" && req.Type != "企业" {
		return errors.ValidationError("type", "client type must be 个人 or 企业")
	}
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
	if len([]rune(req.Phone)) > 20 {
		return errors.ValidationError("phone", "phone is too long")
	}
	identityNumber := strings.TrimSpace(req.IdentityNumber)
	if identityNumber == "" {
		identityNumber = strings.TrimSpace(req.IDCard)
	}
	if identityNumber != "" {
		identityType := normalizedClientIdentityType(req.Type, req.IdentityType)
		if !validClientIdentityType(req.Type, identityType) {
			return errors.ValidationError("identity_type", "identity type does not match client type")
		}
		if len([]rune(security.NormalizeIdentityNumber(string(identityType), identityNumber))) < 4 {
			return errors.ValidationError("identity_number", "identity number is too short")
		}
	}

	return nil
}

func (s *ClientService) toClientResponse(client *models.Client) *ClientResponse {
	return &ClientResponse{
		ID:                client.ID,
		Version:           client.Version,
		Name:              client.Name,
		Type:              client.Type,
		Email:             client.Email,
		Phone:             models.MaskPhone(client.Phone),
		Address:           client.Address,
		IDCard:            client.ToSafeResponse().IDCard,
		IdentityType:      client.EffectiveIdentityType(),
		IdentityNumber:    client.MaskedIdentity(),
		IdentityStatus:    client.IdentityStatus(),
		Aliases:           client.Aliases,
		Company:           client.Company,
		Industry:          client.Industry,
		ContactPerson:     client.ContactPerson,
		ContactPhone:      models.MaskPhone(client.ContactPhone),
		Source:            client.Source,
		Notes:             client.Notes,
		Status:            client.Status,
		CreatedAt:         client.CreatedAt,
		UpdatedAt:         client.UpdatedAt,
		Completeness:      calculateClientCompleteness(client),
		RelatedParties:    []RelatedPartySummary{},
		HistoricalMatters: []HistoricalMatterSummary{},
	}
}

func calculateClientCompleteness(client *models.Client) ClientCompleteness {
	type requiredField struct {
		name   string
		filled bool
	}

	hasValue := func(value string) bool {
		return strings.TrimSpace(value) != ""
	}

	requiredFields := []requiredField{
		{name: "name", filled: hasValue(client.Name)},
		{name: "type", filled: hasValue(client.Type)},
		{name: "status", filled: hasValue(client.Status)},
		{name: "phone_or_email", filled: hasValue(client.Phone) || hasValue(client.Email)},
	}

	if client.Type == "企业" {
		requiredFields = append(requiredFields,
			requiredField{name: "unified_social_credit_code", filled: client.HasIdentity()},
			requiredField{name: "company", filled: hasValue(client.Company)},
			requiredField{name: "contact_person", filled: hasValue(client.ContactPerson)},
			requiredField{name: "contact_phone", filled: hasValue(client.ContactPhone)},
		)
	} else {
		requiredFields = append(requiredFields,
			requiredField{name: "id_card", filled: client.HasIdentity()},
		)
	}

	missingFields := make([]string, 0)
	filledCount := 0
	for _, field := range requiredFields {
		if field.filled {
			filledCount++
			continue
		}
		missingFields = append(missingFields, field.name)
	}

	score := 100
	if len(requiredFields) > 0 {
		score = filledCount * 100 / len(requiredFields)
	}

	return ClientCompleteness{
		Score:                 score,
		MissingFields:         missingFields,
		ReadyForConflictCheck: len(missingFields) == 0,
	}
}

func normalizedClientIdentityType(clientType, identityType string) models.IdentityType {
	value := models.IdentityType(strings.ToUpper(strings.TrimSpace(identityType)))
	if value != "" {
		return value
	}
	if strings.EqualFold(strings.TrimSpace(clientType), "企业") {
		return models.IdentityTypeSocialCredit
	}
	return models.IdentityTypeIDCard
}

func validClientIdentityType(clientType string, identityType models.IdentityType) bool {
	if strings.EqualFold(strings.TrimSpace(clientType), "个人") {
		return identityType == models.IdentityTypeIDCard || identityType == models.IdentityTypePassport || identityType == models.IdentityTypeOther
	}
	return identityType == models.IdentityTypeSocialCredit || identityType == models.IdentityTypeBusinessLicense || identityType == models.IdentityTypeOrgCode || identityType == models.IdentityTypeOther
}
