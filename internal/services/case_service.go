package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// CaseService 案件服务
type CaseService struct {
	caseRepo       repositories.CaseRepository
	clientRepo     repositories.ClientRepository
	userRepo       repositories.UserRepository
	subjectRecheck *SubjectRecheckService
}

// CreateCaseRequest 创建案件请求
type CreateCaseRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=200"`
	Description string `json:"description" binding:"max=2000"`
	ClientID    uint   `json:"client_id" binding:"required"`
	CaseType    string `json:"case_type" binding:"required"`
	Priority    string `json:"priority" binding:"required"`
	StartDate   string `json:"start_date,omitempty"`
	// 团队分配信息
	LawyerID          uint                `json:"lawyer_id" binding:"required"`  // 主办律师
	AssistingLawyerID *uint               `json:"assisting_lawyer_id,omitempty"` // 协办律师
	TeamMembers       []TeamMemberRequest `json:"team_members,omitempty"`        // 其他团队成员
	BillingMethod     string              `json:"billing_method" binding:"required"`
	IsMajorRisk       bool                `json:"is_major_risk"`
	AssignedBy        uint                `json:"assigned_by"` // 分配者ID
	ConflictCheckID   string              `json:"conflict_check_id,omitempty"`
	Approved          bool                `json:"-"` // 仅供审批通过后的内部成案流程使用
}

// TeamMemberRequest 团队成员请求
type TeamMemberRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Role     string `json:"role" binding:"required"` // paralegal, assistant, intern
	Capacity int    `json:"capacity,omitempty"`      // 工作容量百分比
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
	ID                       uint           `json:"id"`
	CaseNumber               string         `json:"case_number"`
	Title                    string         `json:"title"`
	Description              string         `json:"description"`
	ClientID                 uint           `json:"client_id"`
	Client                   *models.Client `json:"client,omitempty"`
	LawyerID                 uint           `json:"lawyer_id"`
	Lawyer                   *models.User   `json:"lawyer,omitempty"`
	CaseType                 string         `json:"case_type"`
	Priority                 string         `json:"priority"`
	Status                   string         `json:"status"`
	StartDate                *time.Time     `json:"start_date"`
	EndDate                  *time.Time     `json:"end_date"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	SubjectVersion           int            `json:"subject_version"`
	SubjectState             string         `json:"subject_state"`
	PendingSubjectRevisionID string         `json:"pending_subject_revision_id,omitempty"`
	ConflictCoverageStatus   string         `json:"conflict_coverage_status"`
}

// ListCasesRequest 案件列表请求
type ListCasesRequest struct {
	Page              int    `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize          int    `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=100"`
	Search            string `json:"search" form:"search"`
	Status            string `json:"status" form:"status"`
	CaseType          string `json:"case_type" form:"case_type"`
	ClientID          uint   `json:"client_id" form:"client_id"`
	LawyerID          uint   `json:"lawyer_id" form:"lawyer_id"`
	EthicalWallUserID uint   `json:"-" form:"-"`
}

// ListCasesResponse 案件列表响应
type ListCasesResponse struct {
	Cases      []CaseResponse          `json:"cases"`
	Pagination PaginationWithTotalPage `json:"pagination"`
}

// NewCaseService 创建案件服务
func NewCaseService(caseRepo repositories.CaseRepository, clientRepo repositories.ClientRepository, userRepo repositories.UserRepository) *CaseService {
	return &CaseService{
		caseRepo:   caseRepo,
		clientRepo: clientRepo,
		userRepo:   userRepo,
	}
}

// SetSubjectRecheckService installs the production server-side action gate.
func (s *CaseService) SetSubjectRecheckService(service *SubjectRecheckService) {
	s.subjectRecheck = service
}

// CreateCase 创建案件
func (s *CaseService) CreateCase(ctx context.Context, req *CreateCaseRequest) (*CaseResponse, error) {
	if req == nil {
		return nil, errors.New("案件数据不能为空")
	}
	if !req.Approved {
		return nil, NewSubjectWorkflowError("CASE_INTAKE_REQUIRED", "正式案件必须通过立案工作台完成冲突检查和审批后成案")
	}
	if req.Approved {
		if err := s.ValidateApprovedCase(ctx, req); err != nil {
			return nil, err
		}
	}
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
	caseStatus := "pending"
	if req.Approved {
		caseStatus = "active"
	}
	case_ := &models.Case{
		CaseNumber:  fmt.Sprintf("CASE-%s", time.Now().Format("20060102150405")),
		Title:       req.Title,
		Description: req.Description,
		ClientID:    req.ClientID,
		LawyerID:    req.LawyerID,
		CaseType:    req.CaseType,
		Priority:    req.Priority,
		Status:      caseStatus,
		StartDate:   startDate,
		CreatedBy:   fmt.Sprintf("%d", req.AssignedBy),
	}
	if req.Approved {
		case_.SubjectVersion = 1
		case_.SubjectState = models.SubjectStateEffective
		case_.ConflictCheckID = req.ConflictCheckID
		case_.ConflictCoverageStatus = "COMPLETE"
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

// ValidateApprovedCase runs every server-side prerequisite without creating a
// case. Approval workflows use this preflight before changing an approval to
// approved, so a deterministic gate failure cannot leave a misleading
// "approved but no case" result.
func (s *CaseService) ValidateApprovedCase(ctx context.Context, req *CreateCaseRequest) error {
	if req == nil {
		return NewSubjectWorkflowError("CASE_DATA_INVALID", "正式成案数据不能为空")
	}
	if s == nil || s.clientRepo == nil || s.userRepo == nil {
		return NewSubjectWorkflowError("CASE_GATE_UNAVAILABLE", "正式成案依赖服务未初始化，已阻止成案")
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.CaseType) == "" || strings.TrimSpace(req.Priority) == "" || strings.TrimSpace(req.BillingMethod) == "" || req.ClientID == 0 || req.LawyerID == 0 {
		return NewSubjectWorkflowError("CASE_DATA_INVALID", "正式成案所需的案件、客户、承办律师和计费信息不完整")
	}
	if s.subjectRecheck == nil {
		return NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "正式成案门禁未初始化，已阻止成案")
	}
	if strings.TrimSpace(req.ConflictCheckID) == "" {
		return NewSubjectWorkflowError("CONFLICT_CHECK_REQUIRED", "正式成案前必须关联已复核的利益冲突检测记录")
	}
	if err := s.subjectRecheck.RequireConflictDispositionForCase(ctx, req.ConflictCheckID, req.ClientID, req.LawyerID, "case_creation"); err != nil {
		return err
	}
	client, err := s.clientRepo.FindByID(ctx, req.ClientID)
	if err != nil || client == nil {
		return NewSubjectWorkflowError("CLIENT_NOT_FOUND", "正式成案指定的客户不存在")
	}
	lawyer, err := s.userRepo.FindByID(ctx, req.LawyerID)
	if err != nil || lawyer == nil || strings.ToLower(strings.TrimSpace(lawyer.Status)) != "active" {
		return NewSubjectWorkflowError("LAWYER_NOT_AVAILABLE", "正式成案指定的承办律师不存在或已停用")
	}
	return nil
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
	if isControlledCaseStatus(req.Status) {
		if s.subjectRecheck == nil {
			return nil, NewSubjectWorkflowError("SUBJECT_GATE_UNAVAILABLE", "案件受控动作门禁未初始化，已阻止更新")
		}
		if err := s.subjectRecheck.RequireEffectiveSubject(ctx, id, "case_status_update"); err != nil {
			return nil, err
		}
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
		Page:              req.Page,
		PageSize:          req.PageSize,
		Search:            req.Search,
		Status:            req.Status,
		CaseType:          req.CaseType,
		ClientID:          req.ClientID,
		LawyerID:          req.LawyerID,
		EthicalWallUserID: req.EthicalWallUserID,
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
		ID:                       case_.ID,
		CaseNumber:               case_.CaseNumber,
		Title:                    case_.Title,
		Description:              case_.Description,
		ClientID:                 case_.ClientID,
		Client:                   case_.Client,
		LawyerID:                 case_.LawyerID,
		Lawyer:                   case_.Lawyer,
		CaseType:                 case_.CaseType,
		Priority:                 case_.Priority,
		Status:                   case_.Status,
		StartDate:                case_.StartDate,
		EndDate:                  case_.EndDate,
		CreatedAt:                case_.CreatedAt,
		UpdatedAt:                case_.UpdatedAt,
		SubjectVersion:           case_.SubjectVersion,
		SubjectState:             case_.SubjectState,
		PendingSubjectRevisionID: case_.PendingSubjectRevisionID,
		ConflictCoverageStatus:   case_.ConflictCoverageStatus,
	}
}

// GetLawyers 获取律师列表
func (s *CaseService) GetLawyers(ctx context.Context, page, pageSize int) ([]models.User, error) {
	return s.userRepo.GetLawyers(ctx, page, pageSize)
}

// GetLawyerByID 获取单个律师用户详情
func (s *CaseService) GetLawyerByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user.Role != "lawyer" && user.Role != "admin" {
		return nil, repositories.ErrUserNotFound
	}
	return user, nil
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
