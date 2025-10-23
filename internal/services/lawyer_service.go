package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	stderrors "errors"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"gorm.io/gorm"
)

type LawyerService struct {
	lawyerRepo repositories.LawyerRepository
}

func NewLawyerService(lawyerRepo repositories.LawyerRepository) *LawyerService {
	return &LawyerService{lawyerRepo: lawyerRepo}
}

type LawyerResponse struct {
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Avatar         string    `json:"avatar"`
	Status         string    `json:"status"`
	LicenseNumber  string    `json:"licenseNumber"`  // 执业证号
	Specialty      string    `json:"specialty"`      // 专业领域
	Department     string    `json:"department"`     // 部门
	Position       string    `json:"position"`       // 职位
	Gender         string    `json:"gender"`         // 性别
	Experience     int       `json:"experience"`     // 从业年限
	JoinDate       string    `json:"joinDate"`       // 入职日期
	Profile        string    `json:"profile"`        // 个人简介
	CreatedAt      time.Time `json:"created_at"`
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
	// 🔧 修复：完善律师响应数据，为律师特有字段提供默认值或推导值
	response := &LawyerResponse{
		ID:        lawyer.ID,
		Name:      lawyer.Name,
		Email:     lawyer.Email,
		Phone:     lawyer.Phone,
		Avatar:    lawyer.Avatar,
		Status:    lawyer.Status,
		CreatedAt: lawyer.CreatedAt,
	}

	// 根据用户名或邮箱推导执业证号
	if lawyer.Username != "" {
		response.LicenseNumber = fmt.Sprintf("LAW%s", lawyer.Username)
	} else {
		response.LicenseNumber = fmt.Sprintf("LAW%d", lawyer.ID)
	}

	// 根据用户姓名或邮箱推导专业领域
	response.Specialty = "综合法律服务"
	if strings.Contains(lawyer.Name, "刑事") || strings.Contains(lawyer.Name, "刑法") {
		response.Specialty = "刑事辩护"
	} else if strings.Contains(lawyer.Name, "民事") || strings.Contains(lawyer.Name, "合同") {
		response.Specialty = "民事诉讼"
	} else if strings.Contains(lawyer.Name, "商事") || strings.Contains(lawyer.Name, "公司") {
		response.Specialty = "商事法律"
	}

	// 默认部门
	response.Department = "法务部"

	// 根据名称或邮箱推导职位
	if strings.Contains(lawyer.Name, "主任") || strings.Contains(lawyer.Name, "高级") {
		response.Position = "高级律师"
	} else {
		response.Position = "执业律师"
	}

	// 计算从业年限（基于创建时间）
	yearsOfExperience := time.Since(lawyer.CreatedAt).Hours() / 24 / 365
	response.Experience = int(yearsOfExperience)

	// 入职日期
	response.JoinDate = lawyer.CreatedAt.Format("2006-01-02")

	// 默认个人简介
	response.Profile = fmt.Sprintf("%s，专业从事法律服务%d年，擅长%s。",
		lawyer.Name, response.Experience, response.Specialty)

	return response
}

func (s *LawyerService) DeleteLawyer(ctx context.Context, id uint) error {
	if err := s.lawyerRepo.Delete(ctx, id); err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFoundError("lawyer", "Lawyer not found", id)
		}
		return err
	}
	return nil
}
