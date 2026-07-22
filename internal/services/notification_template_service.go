package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// NotificationTemplateService 通知模板服务
type NotificationTemplateService struct {
	templateRepo repositories.NotificationTemplateRepository
	queueRepo    repositories.NotificationQueueRepository
}

// NewNotificationTemplateService 创建通知模板服务实例
func NewNotificationTemplateService(
	templateRepo repositories.NotificationTemplateRepository,
	queueRepo repositories.NotificationQueueRepository,
) *NotificationTemplateService {
	return &NotificationTemplateService{
		templateRepo: templateRepo,
		queueRepo:    queueRepo,
	}
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	TemplateCode     string   `json:"template_code" binding:"required,min=1,max=50"`
	TemplateName     string   `json:"template_name" binding:"required,min=1,max=100"`
	Channel          string   `json:"channel" binding:"required,oneof=email sms wechat"`
	RecipientType    string   `json:"recipient_type" binding:"required,oneof=client lawyer admin"`
	TriggerEvent     string   `json:"trigger_event" binding:"required,max=100"`
	SubjectTemplate  string   `json:"subject_template" binding:"max=200"`
	ContentTemplate  string   `json:"content_template" binding:"required"`
	Variables        []string `json:"variables"`
	AutoSend         bool     `json:"auto_send"`
	RequiresApproval bool     `json:"requires_approval"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	TemplateName     *string  `json:"template_name,omitempty" binding:"omitempty,min=1,max=100"`
	SubjectTemplate  *string  `json:"subject_template,omitempty" binding:"omitempty,max=200"`
	ContentTemplate  *string  `json:"content_template,omitempty" binding:"omitempty,min=1"`
	Variables        []string `json:"variables"`
	AutoSend         *bool    `json:"auto_send,omitempty"`
	RequiresApproval *bool    `json:"requires_approval,omitempty"`
	IsActive         *bool    `json:"is_active,omitempty"`
}

// TemplateResponse 模板响应
type TemplateResponse struct {
	ID               uint     `json:"id"`
	TemplateCode     string   `json:"template_code"`
	TemplateName     string   `json:"template_name"`
	Channel          string   `json:"channel"`
	RecipientType    string   `json:"recipient_type"`
	TriggerEvent     string   `json:"trigger_event"`
	SubjectTemplate  string   `json:"subject_template"`
	ContentTemplate  string   `json:"content_template"`
	Variables        []string `json:"variables"`
	AutoSend         bool     `json:"auto_send"`
	RequiresApproval bool     `json:"requires_approval"`
	IsActive         bool     `json:"is_active"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

// ListTemplatesRequest 模板列表请求
type ListTemplatesRequest struct {
	Page          int    `json:"page" form:"page" binding:"min=1"`
	PageSize      int    `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	Channel       string `json:"channel" form:"channel" binding:"omitempty,oneof=email sms wechat"`
	RecipientType string `json:"recipient_type" form:"recipient_type" binding:"omitempty,oneof=client lawyer admin"`
	TriggerEvent  string `json:"trigger_event" form:"trigger_event"`
	IsActive      *bool  `json:"is_active" form:"is_active"`
}

// ListTemplatesResponse 模板列表响应
type ListTemplatesResponse struct {
	Templates  []*TemplateResponse `json:"templates"`
	Pagination PaginationInfo      `json:"pagination"`
}

// CreateTemplate 创建通知模板
func (s *NotificationTemplateService) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (*TemplateResponse, error) {
	// 检查模板代码是否已存在
	existingTemplate, _ := s.templateRepo.FindByCode(ctx, req.TemplateCode)
	if existingTemplate != nil {
		return nil, errors.New("模板代码已存在")
	}

	// 验证模板内容中的变量
	contentVariables := s.extractVariables(req.ContentTemplate)
	subjectVariables := s.extractVariables(req.SubjectTemplate)

	// 合并变量
	allVariables := make(map[string]bool)
	for _, v := range contentVariables {
		allVariables[v] = true
	}
	for _, v := range subjectVariables {
		allVariables[v] = true
	}

	// 如果请求中指定了变量，进行验证
	if len(req.Variables) > 0 {
		for _, v := range req.Variables {
			if !allVariables[v] {
				return nil, fmt.Errorf("变量 %s 在模板中未使用", v)
			}
		}
	}

	template := &models.NotificationTemplate{
		TemplateCode:     req.TemplateCode,
		TemplateName:     req.TemplateName,
		Channel:          req.Channel,
		RecipientType:    req.RecipientType,
		TriggerEvent:     req.TriggerEvent,
		SubjectTemplate:  req.SubjectTemplate,
		ContentTemplate:  req.ContentTemplate,
		Variables:        models.JSON{"variables": req.Variables},
		AutoSend:         req.AutoSend,
		RequiresApproval: req.RequiresApproval,
		IsActive:         true,
	}

	if err := s.templateRepo.Create(ctx, template); err != nil {
		return nil, fmt.Errorf("创建模板失败: %w", err)
	}

	return s.GetTemplateByID(ctx, template.ID)
}

// extractVariables 从模板中提取变量
func (s *NotificationTemplateService) extractVariables(template string) []string {
	variables := make(map[string]bool)

	// 提取 {{variable}} 格式的变量
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			variables[match[1]] = true
		}
	}

	result := make([]string, 0, len(variables))
	for v := range variables {
		result = append(result, v)
	}

	return result
}

// GetTemplateByID 根据ID获取模板详情
func (s *NotificationTemplateService) GetTemplateByID(ctx context.Context, id uint) (*TemplateResponse, error) {
	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	if template == nil {
		return nil, errors.New("模板不存在")
	}

	return s.convertToResponse(template), nil
}

// GetTemplateByCode 根据代码获取模板详情
func (s *NotificationTemplateService) GetTemplateByCode(ctx context.Context, code string) (*TemplateResponse, error) {
	template, err := s.templateRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	if template == nil {
		return nil, errors.New("模板不存在")
	}

	return s.convertToResponse(template), nil
}

// ListTemplates 获取模板列表
func (s *NotificationTemplateService) ListTemplates(ctx context.Context, req *ListTemplatesRequest) (*ListTemplatesResponse, error) {
	params := &repositories.TemplateListParams{
		Page:          req.Page,
		PageSize:      req.PageSize,
		Channel:       req.Channel,
		RecipientType: req.RecipientType,
		TriggerEvent:  req.TriggerEvent,
		IsActive:      req.IsActive,
	}

	templates, total, err := s.templateRepo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("查询模板列表失败: %w", err)
	}

	response := &ListTemplatesResponse{
		Templates: make([]*TemplateResponse, len(templates)),
		Pagination: PaginationInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}

	for i, t := range templates {
		response.Templates[i] = s.convertToResponse(t)
	}

	return response, nil
}

// UpdateTemplate 更新模板
func (s *NotificationTemplateService) UpdateTemplate(ctx context.Context, id uint, req *UpdateTemplateRequest) (*TemplateResponse, error) {
	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	if template == nil {
		return nil, errors.New("模板不存在")
	}

	// 更新字段
	if req.TemplateName != nil {
		template.TemplateName = *req.TemplateName
	}
	if req.SubjectTemplate != nil {
		template.SubjectTemplate = *req.SubjectTemplate
	}
	if req.ContentTemplate != nil {
		template.ContentTemplate = *req.ContentTemplate
		// 重新提取变量
		contentVariables := s.extractVariables(*req.ContentTemplate)
		template.Variables = models.JSON{"variables": contentVariables}
	}
	if req.AutoSend != nil {
		template.AutoSend = *req.AutoSend
	}
	if req.RequiresApproval != nil {
		template.RequiresApproval = *req.RequiresApproval
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}
	if req.Variables != nil {
		template.Variables = models.JSON{"variables": req.Variables}
	}

	if err := s.templateRepo.Update(ctx, template); err != nil {
		return nil, fmt.Errorf("更新模板失败: %w", err)
	}

	return s.GetTemplateByID(ctx, id)
}

// DeleteTemplate 删除模板
func (s *NotificationTemplateService) DeleteTemplate(ctx context.Context, id uint) error {
	template, err := s.templateRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("查询模板失败: %w", err)
	}
	if template == nil {
		return errors.New("模板不存在")
	}

	// 检查是否有使用该模板的通知
	// TODO: 实现模板使用检查逻辑
	_ = ctx
	_ = s.queueRepo
	// usingNotifications, _, _ := s.queueRepo.List(ctx, &repositories.NotificationListParams{
	// 	Page:     1,
	// 	PageSize: 1,
	// })
	// if len(usingNotifications) > 0 {
	// 	return errors.New("该模板正在使用中，无法删除")
	// }

	if err := s.templateRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除模板失败: %w", err)
	}

	return nil
}

// PreviewTemplate 预览模板效果
func (s *NotificationTemplateService) PreviewTemplate(ctx context.Context, templateCode string, data map[string]interface{}) (*TemplatePreviewResponse, error) {
	template, err := s.templateRepo.FindByCode(ctx, templateCode)
	if err != nil {
		return nil, fmt.Errorf("查询模板失败: %w", err)
	}
	if template == nil {
		return nil, errors.New("模板不存在")
	}

	// 填充变量
	subject := template.SubjectTemplate
	content := template.ContentTemplate

	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		content = strings.ReplaceAll(content, placeholder, replacement)
		subject = strings.ReplaceAll(subject, placeholder, replacement)
	}

	return &TemplatePreviewResponse{
		TemplateCode:      template.TemplateCode,
		TemplateName:      template.TemplateName,
		Subject:           subject,
		Content:           content,
		UnfilledVariables: s.getUnfilledVariables(content),
	}, nil
}

// getUnfilledVariables 获取未填充的变量
func (s *NotificationTemplateService) getUnfilledVariables(content string) []string {
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(content, -1)

	variables := make([]string, 0)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			if !seen[match[1]] {
				variables = append(variables, match[1])
				seen[match[1]] = true
			}
		}
	}

	return variables
}

// TemplatePreviewResponse 模板预览响应
type TemplatePreviewResponse struct {
	TemplateCode      string   `json:"template_code"`
	TemplateName      string   `json:"template_name"`
	Subject           string   `json:"subject"`
	Content           string   `json:"content"`
	UnfilledVariables []string `json:"unfilled_variables"`
}

// convertToResponse 转换为响应格式
func (s *NotificationTemplateService) convertToResponse(template *models.NotificationTemplate) *TemplateResponse {
	// 解析变量
	var variables []string
	if len(template.Variables) > 0 {
		if v, ok := template.Variables["variables"].([]interface{}); ok {
			for _, v := range v {
				if str, ok := v.(string); ok {
					variables = append(variables, str)
				}
			}
		}
	}

	return &TemplateResponse{
		ID:               template.ID,
		TemplateCode:     template.TemplateCode,
		TemplateName:     template.TemplateName,
		Channel:          template.Channel,
		RecipientType:    template.RecipientType,
		TriggerEvent:     template.TriggerEvent,
		SubjectTemplate:  template.SubjectTemplate,
		ContentTemplate:  template.ContentTemplate,
		Variables:        variables,
		AutoSend:         template.AutoSend,
		RequiresApproval: template.RequiresApproval,
		IsActive:         template.IsActive,
		CreatedAt:        template.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        template.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
