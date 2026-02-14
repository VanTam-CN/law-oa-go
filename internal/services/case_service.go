package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// CaseService 案件服务
type CaseService struct {
	caseRepo  repositories.CaseRepository
	clientRepo repositories.ClientRepository
	userRepo  repositories.UserRepository
}

// CreateCaseRequest 创建案件请求
type CreateCaseRequest struct {
	Title              string                    `json:"title" binding:"required,min=1,max=200"`
	Description        string                    `json:"description" binding:"max=2000"`
	ClientID           uint                      `json:"client_id" binding:"required"`
	CaseType           string                    `json:"case_type" binding:"required"`
	Priority           string                    `json:"priority" binding:"required"`
	StartDate          string                    `json:"start_date,omitempty"`
	// 团队分配信息
	LawyerID           uint                      `json:"lawyer_id" binding:"required"`           // 主办律师
	AssistingLawyerID  *uint                     `json:"assisting_lawyer_id,omitempty"`        // 协办律师
	TeamMembers        []TeamMemberRequest       `json:"team_members,omitempty"`              // 其他团队成员
	BillingMethod      string                    `json:"billing_method" binding:"required"`
	IsMajorRisk        bool                      `json:"is_major_risk"`
	AssignedBy         uint                      `json:"assigned_by"`                         // 分配者ID
}

// TeamMemberRequest 团队成员请求
type TeamMemberRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Role     string `json:"role" binding:"required"`     // paralegal, assistant, intern
	Capacity int    `json:"capacity,omitempty"`         // 工作容量百分比
}

// UpdateCaseRequest 更新案件请求
type UpdateCaseRequest struct {
	Title       string `json:"title,omitempty" binding:"min=1,max=200"`
	Description string `json:"description,omitempty" binding:"max=2000"`
	CaseType    string `json:"case_type,omitempty"`
	Priority    string `json:"priority,omitempty"`
	Status      string `json:"status,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

// CaseResponse 案件响应
type CaseResponse struct {
	ID          uint                   `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	ClientID    uint                   `json:"client_id"`
	Client      *models.Client         `json:"client,omitempty"`
	LawyerID    uint                   `json:"lawyer_id"`
	Lawyer      *models.User           `json:"lawyer,omitempty"`
	CaseType    string                 `json:"case_type"`
	Priority    string                 `json:"priority"`
	Status      string                 `json:"status"`
	StartDate   *time.Time             `json:"start_date"`
	EndDate     *time.Time             `json:"end_date"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ListCasesRequest 案件列表请求
type ListCasesRequest struct {
	Page     int    `json:"page" form:"page" binding:"min=1"`
	PageSize int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Search   string `json:"search" form:"search"`
	Status   string `json:"status" form:"status"`
	CaseType string `json:"case_type" form:"case_type"`
	ClientID uint   `json:"client_id" form:"client_id"`
	LawyerID uint   `json:"lawyer_id" form:"lawyer_id"`
}

// ListCasesResponse 案件列表响应
type ListCasesResponse struct {
	Cases      []CaseResponse           `json:"cases"`
	Pagination PaginationWithTotalPage `json:"pagination"`
}

// NewCaseService 创建案件服务
func NewCaseService(caseRepo repositories.CaseRepository, clientRepo repositories.ClientRepository, userRepo repositories.UserRepository) *CaseService {
	return &CaseService{
		caseRepo:  caseRepo,
		clientRepo: clientRepo,
		userRepo:  userRepo,
	}
}

// CreateCase 创建案件
func (s *CaseService) CreateCase(ctx context.Context, req *CreateCaseRequest) (*CaseResponse, error) {
	// 验证客户是否存在
	client, err := s.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("客户不存在: %w", err)
	}
	if client == nil {
		return nil, errors.New("客户不存在")
	}

	// 解析开始日期
	var startDate *time.Time
	if req.StartDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		startDate = &parsedDate
	}

	// 创建案件
	case_ := &models.Case{
		Title:       req.Title,
		Description: req.Description,
		ClientID:    req.ClientID,
		CaseType:    req.CaseType,
		Priority:    req.Priority,
		Status:      "pending",
		StartDate:   startDate,
	}

	err = s.caseRepo.Create(ctx, case_)
	if err != nil {
		return nil, fmt.Errorf("创建案件失败: %w", err)
	}

	// 获取创建的完整案件信息
	caseWithDetails, err := s.caseRepo.FindByID(ctx, case_.ID)
	if err != nil {
		return nil, fmt.Errorf("获取案件详情失败: %w", err)
	}

	return s.convertToResponse(caseWithDetails), nil
}

// UpdateCase 更新案件
func (s *CaseService) UpdateCase(ctx context.Context, id uint, req *UpdateCaseRequest) (*CaseResponse, error) {
	// 获取现有案件
	existingCase, err := s.caseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取案件失败: %w", err)
	}
	if existingCase == nil {
		return nil, errors.New("案件不存在")
	}

	// 解析日期
	if req.StartDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("开始日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		existingCase.StartDate = &parsedDate
	}

	if req.EndDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("结束日期格式错误，应为 YYYY-MM-DD: %w", err)
		}
		existingCase.EndDate = &parsedDate
	}

	// 更新字段
	if req.Title != "" {
		existingCase.Title = req.Title
	}
	if req.Description != "" {
		existingCase.Description = req.Description
	}
	if req.CaseType != "" {
		existingCase.CaseType = req.CaseType
	}
	if req.Priority != "" {
		existingCase.Priority = req.Priority
	}
	if req.Status != "" {
		existingCase.Status = req.Status
	}

	// 保存更新
	err = s.caseRepo.Update(ctx, existingCase)
	if err != nil {
		return nil, fmt.Errorf("更新案件失败: %w", err)
	}

	return s.convertToResponse(existingCase), nil
}

// GetCaseByID 根据ID获取案件
func (s *CaseService) GetCaseByID(ctx context.Context, id uint) (*CaseResponse, error) {
	case_, err := s.caseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取案件失败: %w", err)
	}
	if case_ == nil {
		return nil, errors.New("案件不存在")
	}

	return s.convertToResponse(case_), nil
}

// ListCases 获取案件列表
func (s *CaseService) ListCases(ctx context.Context, req *ListCasesRequest) (*ListCasesResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 构建查询参数
	params := repositories.CaseListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Search:   req.Search,
		Status:   req.Status,
		CaseType: req.CaseType,
		ClientID: req.ClientID,
		LawyerID: req.LawyerID,
	}

	// 获取案件列表
	cases, total, err := s.caseRepo.List(ctx, &params)
	if err != nil {
		return nil, fmt.Errorf("获取案件列表失败: %w", err)
	}

	// 转换为响应格式
	caseResponses := make([]CaseResponse, len(cases))
	for i, case_ := range cases {
		caseResponses[i] = *s.convertToResponse(case_)
	}

	// 构建分页信息
	pagination := PaginationWithTotalPage{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Total:     total,
		TotalPage: (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	}

	return &ListCasesResponse{
		Cases:      caseResponses,
		Pagination: pagination,
	}, nil
}

// DeleteCase 删除案件
func (s *CaseService) DeleteCase(ctx context.Context, id uint) error {
	// 检查案件是否存在
	existingCase, err := s.caseRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取案件失败: %w", err)
	}
	if existingCase == nil {
		return errors.New("案件不存在")
	}

	// 软删除案件
	return s.caseRepo.Delete(ctx, id)
}

// convertToResponse 转换为响应格式
func (s *CaseService) convertToResponse(case_ *models.Case) *CaseResponse {
	return &CaseResponse{
		ID:          case_.ID,
		Title:       case_.Title,
		Description: case_.Description,
		ClientID:    case_.ClientID,
		Client:      case_.Client,
		LawyerID:    case_.LawyerID,
		Lawyer:      case_.Lawyer,
		CaseType:    case_.CaseType,
		Priority:    case_.Priority,
		Status:      case_.Status,
		StartDate:   case_.StartDate,
		EndDate:     case_.EndDate,
		CreatedAt:   case_.CreatedAt,
		UpdatedAt:   case_.UpdatedAt,
	}
}

// GetLawyers 获取律师列表
func (s *CaseService) GetLawyers(ctx context.Context, page, pageSize int) ([]models.User, error) {
	return s.userRepo.GetLawyers(ctx, page, pageSize)
}

// CaseTypeResponse 案件类型响应
type CaseTypeResponse struct {
	Value  string   `json:"value"`  // 案件类型值
	Label  string   `json:"label"`  // 显示标签
	Causes []string `json:"causes"` // 对应的案由列表
}

// GetCaseTypes 获取案件类型列表
func (s *CaseService) GetCaseTypes(ctx context.Context) ([]CaseTypeResponse, error) {
	// 返回标准的案件类型数据，基于前端CompactCaseFormWrapper中的fallback数据
	caseTypes := []CaseTypeResponse{
		{
			Value:  "民事案件",
			Label:  "民事案件",
			Causes: []string{"合同纠纷", "侵权责任", "婚姻家庭", "继承纠纷"},
		},
		{
			Value:  "商事案件",
			Label:  "商事案件",
			Causes: []string{"公司纠纷", "金融纠纷", "投资纠纷", "证券纠纷"},
		},
		{
			Value:  "刑事案件",
			Label:  "刑事案件",
			Causes: []string{"经济犯罪", "职务犯罪", "暴力犯罪", "网络犯罪"},
		},
		{
			Value:  "行政案件",
			Label:  "行政案件",
			Causes: []string{"行政处罚", "行政许可", "信息公开", "征收补偿"},
		},
		{
			Value:  "知识产权",
			Label:  "知识产权案件",
			Causes: []string{"商标侵权", "专利侵权", "著作权侵权", "商业秘密"},
		},
	}

	return caseTypes, nil
}