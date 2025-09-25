package services

import (
	"context"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"time"
)

type LawyerService struct {
	lawyerRepo repositories.LawyerRepository
}

func NewLawyerService(lawyerRepo repositories.LawyerRepository) *LawyerService {
	return &LawyerService{lawyerRepo: lawyerRepo}
}

type LawyerResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Avatar    string    `json:"avatar"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type LawyerListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search"`
}

func (s *LawyerService) ListLawyers(ctx context.Context, req *LawyerListRequest) ([]*LawyerResponse, int64, error) {
	params := &repositories.LawyerListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Search:   req.Search,
	}
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}

	lawyers, total, err := s.lawyerRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*LawyerResponse, len(lawyers))
	for i, lawyer := range lawyers {
		responses[i] = s.toLawyerResponse(lawyer)
	}

	return responses, total, nil
}

func (s *LawyerService) toLawyerResponse(lawyer *models.User) *LawyerResponse {
	return &LawyerResponse{
		ID:        lawyer.ID,
		Name:      lawyer.Name,
		Email:     lawyer.Email,
		Phone:     lawyer.Phone,
		Avatar:    lawyer.Avatar,
		Status:    lawyer.Status,
		CreatedAt: lawyer.CreatedAt,
	}
}
