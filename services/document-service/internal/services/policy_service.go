package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/law-oa-go/document-service/internal/auth"
	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// PolicyService 策略服务接口
type PolicyService interface {
	// 策略CRUD操作
	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*PolicyResponse, error)
	GetPolicy(ctx context.Context, id uint) (*PolicyResponse, error)
	UpdatePolicy(ctx context.Context, id uint, req *UpdatePolicyRequest) (*PolicyResponse, error)
	DeletePolicy(ctx context.Context, id uint) error
	ListPolicies(ctx context.Context, filter *PolicyFilter) (*PolicyListResponse, error)

	// 策略启用/禁用
	EnablePolicy(ctx context.Context, id uint) error
	DisablePolicy(ctx context.Context, id uint) error

	// 策略评估
	EvaluatePolicy(ctx context.Context, req *EvaluatePolicyRequest) (*EvaluatePolicyResponse, error)
	BulkEvaluatePolicies(ctx context.Context, reqs []*EvaluatePolicyRequest) ([]*EvaluatePolicyResponse, error)

	// 策略模板管理
	CreatePolicyFromTemplate(ctx context.Context, req *CreatePolicyFromTemplateRequest) (*PolicyResponse, error)
	GetPolicyTemplates(ctx context.Context, filter *PolicyTemplateFilter) (*PolicyTemplateListResponse, error)
	CreatePolicyTemplate(ctx context.Context, req *CreatePolicyTemplateRequest) (*PolicyTemplateResponse, error)

	// 策略版本管理
	CreatePolicyVersion(ctx context.Context, policyID uint, req *CreatePolicyVersionRequest) (*PolicyVersionResponse, error)
	GetPolicyVersions(ctx context.Context, policyID uint) (*PolicyVersionListResponse, error)
	RollbackPolicyVersion(ctx context.Context, policyID uint, version int) error

	// 策略分析
	AnalyzePolicyConflicts(ctx context.Context, tenantID string) (*PolicyConflictAnalysis, error)
	GetPolicyRecommendations(ctx context.Context, tenantID string) (*PolicyRecommendationList, error)
	GetPolicyStatistics(ctx context.Context, tenantID string) (*PolicyStatistics, error)

	// 策略测试
	TestPolicy(ctx context.Context, req *TestPolicyRequest) (*TestPolicyResponse, error)
	ValidatePolicy(ctx context.Context, policy *auth.PolicyRule) (*PolicyValidationResult, error)

	// 策略导入导出
	ExportPolicies(ctx context.Context, tenantID string, format string) ([]byte, error)
	ImportPolicies(ctx context.Context, data []byte, format string) (*ImportResult, error)
}

// policyService 策略服务实现
type policyService struct {
	policyRepo    repositories.PolicyRepository
	roleRepo      repositories.RoleRepository
	userRepo      repositories.UserRepository
	docRepo       repositories.DocumentRepository
	auditRepo     repositories.AuditRepository
	abacEngine    *auth.ABACEngine
	logger        *logrus.Logger
}

// NewPolicyService 创建策略服务
func NewPolicyService(
	policyRepo repositories.PolicyRepository,
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	docRepo repositories.DocumentRepository,
	auditRepo repositories.AuditRepository,
	abacEngine *auth.ABACEngine,
	logger *logrus.Logger,
) PolicyService {
	return &policyService{
		policyRepo: policyRepo,
		roleRepo:   roleRepo,
		userRepo:   userRepo,
		docRepo:    docRepo,
		auditRepo:  auditRepo,
		abacEngine: abacEngine,
		logger:     logger,
	}
}

// CreatePolicy 创建策略
func (s *policyService) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*PolicyResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"name":      req.Name,
		"tenant_id": req.TenantID,
		"creator":   req.CreatedBy,
	}).Info("Creating new policy")

	// 验证策略规则
	policyRule := &auth.PolicyRule{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		Effect:      req.Effect,
		Subject:     req.Subject,
		Resource:    req.Resource,
		Action:      req.Action,
		Environment: req.Environment,
		Conditions:  req.Conditions,
		TenantID:    req.TenantID,
		CreatedBy:   req.CreatedBy,
	}

	validation, err := s.abacEngine.ValidatePolicy(ctx, policyRule)
	if err != nil {
		s.logger.WithError(err).Error("Failed to validate policy")
		return nil, fmt.Errorf("policy validation failed: %w", err)
	}

	if !validation.Valid {
		s.logger.WithField("errors", validation.Errors).Error("Policy validation failed")
		return nil, fmt.Errorf("policy validation failed: %v", validation.Errors)
	}

	// 序列化策略组件
	subjectJSON, err := json.Marshal(req.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subject: %w", err)
	}

	resourceJSON, err := json.Marshal(req.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}

	actionJSON, err := json.Marshal(req.Action)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal action: %w", err)
	}

	environmentJSON, err := json.Marshal(req.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal environment: %w", err)
	}

	conditionsJSON, err := json.Marshal(req.Conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	// 创建策略模型
	policy := &models.Policy{
		Name:        req.Name,
		Description: req.Description,
		Version:     1,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		Effect:      req.Effect,
		Subject:     string(subjectJSON),
		Resource:    string(resourceJSON),
		Action:      string(actionJSON),
		Environment: string(environmentJSON),
		Conditions:  string(conditionsJSON),
		TenantID:    req.TenantID,
		CreatedBy:   req.CreatedBy,
		Tags:        req.Tags,
	}

	if err := s.policyRepo.Create(ctx, policy); err != nil {
		s.logger.WithError(err).Error("Failed to create policy")
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	// 记录审计日志
	s.logPolicyOperation(ctx, "create", policy.ID, req.CreatedBy, req.TenantID, "Policy created")

	s.logger.WithField("policy_id", policy.ID).Info("Policy created successfully")

	return s.buildPolicyResponse(policy), nil
}

// GetPolicy 获取策略
func (s *policyService) GetPolicy(ctx context.Context, id uint) (*PolicyResponse, error) {
	policy, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("policy_id", id).Error("Failed to get policy")
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	return s.buildPolicyResponse(policy), nil
}

// UpdatePolicy 更新策略
func (s *policyService) UpdatePolicy(ctx context.Context, id uint, req *UpdatePolicyRequest) (*PolicyResponse, error) {
	policy, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	// 记录更新前的版本
	oldVersion := *policy

	// 更新字段
	if req.Name != "" {
		policy.Name = req.Name
	}
	if req.Description != "" {
		policy.Description = req.Description
	}
	if req.Enabled != nil {
		policy.Enabled = *req.Enabled
	}
	if req.Priority != nil {
		policy.Priority = *req.Priority
	}
	if req.Effect != "" {
		policy.Effect = req.Effect
	}
	if req.Subject != nil {
		subjectJSON, _ := json.Marshal(req.Subject)
		policy.Subject = string(subjectJSON)
	}
	if req.Resource != nil {
		resourceJSON, _ := json.Marshal(req.Resource)
		policy.Resource = string(resourceJSON)
	}
	if req.Action != nil {
		actionJSON, _ := json.Marshal(req.Action)
		policy.Action = string(actionJSON)
	}
	if req.Environment != nil {
		environmentJSON, _ := json.Marshal(req.Environment)
		policy.Environment = string(environmentJSON)
	}
	if req.Conditions != nil {
		conditionsJSON, _ := json.Marshal(req.Conditions)
		policy.Conditions = string(conditionsJSON)
	}
	if req.Tags != nil {
		policy.Tags = req.Tags
	}

	if err := s.policyRepo.Update(ctx, policy); err != nil {
		s.logger.WithError(err).WithField("policy_id", id).Error("Failed to update policy")
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	// 记录审计日志
	s.logPolicyOperation(ctx, "update", policy.ID, req.UpdatedBy, policy.TenantID,
		fmt.Sprintf("Policy updated: %v", s.getPolicyChanges(&oldVersion, policy)))

	s.logger.WithField("policy_id", policy.ID).Info("Policy updated successfully")

	return s.buildPolicyResponse(policy), nil
}

// DeletePolicy 删除策略
func (s *policyService) DeletePolicy(ctx context.Context, id uint) error {
	policy, err := s.policyRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	if err := s.policyRepo.Delete(ctx, id); err != nil {
		s.logger.WithError(err).WithField("policy_id", id).Error("Failed to delete policy")
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	// 记录审计日志
	s.logPolicyOperation(ctx, "delete", id, 0, policy.TenantID, "Policy deleted")

	s.logger.WithField("policy_id", id).Info("Policy deleted successfully")

	return nil
}

// ListPolicies 列出策略
func (s *policyService) ListPolicies(ctx context.Context, filter *PolicyFilter) (*PolicyListResponse, error) {
	policies, total, err := s.policyRepo.List(ctx, &repositories.PolicyFilter{
		TenantID:     filter.TenantID,
		Name:         filter.Name,
		Description:  filter.Description,
		Enabled:      filter.Enabled,
		ResourceType: filter.ResourceType,
		ActionType:   filter.ActionType,
		SubjectType:  filter.SubjectType,
		CreatorID:    filter.CreatorID,
		Tags:         filter.Tags,
		CreatedFrom:  filter.CreatedFrom,
		CreatedTo:    filter.CreatedTo,
		UpdatedFrom:  filter.UpdatedFrom,
		UpdatedTo:    filter.UpdatedTo,
		Pagination:   filter.Pagination,
		SortBy:       filter.SortBy,
		SortOrder:    filter.SortOrder,
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to list policies")
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	var policyResponses []*PolicyResponse
	for _, policy := range policies {
		policyResponses = append(policyResponses, s.buildPolicyResponse(policy))
	}

	return &PolicyListResponse{
		Policies: policyResponses,
		Total:    total,
		Page:     filter.Pagination.Page,
		PageSize: filter.Pagination.PageSize,
	}, nil
}

// EnablePolicy 启用策略
func (s *policyService) EnablePolicy(ctx context.Context, id uint) error {
	if err := s.policyRepo.EnablePolicy(ctx, id); err != nil {
		s.logger.WithError(err).WithField("policy_id", id).Error("Failed to enable policy")
		return fmt.Errorf("failed to enable policy: %w", err)
	}

	s.logger.WithField("policy_id", id).Info("Policy enabled successfully")
	return nil
}

// DisablePolicy 禁用策略
func (s *policyService) DisablePolicy(ctx context.Context, id uint) error {
	if err := s.policyRepo.DisablePolicy(ctx, id); err != nil {
		s.logger.WithError(err).WithField("policy_id", id).Error("Failed to disable policy")
		return fmt.Errorf("failed to disable policy: %w", err)
	}

	s.logger.WithField("policy_id", id).Info("Policy disabled successfully")
	return nil
}

// EvaluatePolicy 评估策略
func (s *policyService) EvaluatePolicy(ctx context.Context, req *EvaluatePolicyRequest) (*EvaluatePolicyResponse, error) {
	// 构建访问请求
	accessRequest := &auth.AccessRequest{
		Subject: auth.UserContext{
			ID:         req.SubjectID,
			Username:   req.Subject.Username,
			Email:      req.Subject.Email,
			Roles:      req.Subject.Roles,
			Groups:     req.Subject.Groups,
			Attributes: req.Subject.Attributes,
			TenantID:   req.Subject.TenantID,
			Active:     req.Subject.Active,
		},
		Resource: auth.ResourceContext{
			Type:        req.Resource.Type,
			ID:          req.Resource.ID,
			Owner:       req.Resource.Owner,
			TenantID:    req.Resource.TenantID,
			Attributes:  req.Resource.Attributes,
			Sensitivity: req.Resource.Sensitivity,
			Category:    req.Resource.Category,
			Tags:        req.Resource.Tags,
			CreatedAt:   req.Resource.CreatedAt,
			UpdatedAt:   req.Resource.UpdatedAt,
		},
		Action: auth.ActionContext{
			Type:       req.Action.Type,
			Method:     req.Action.Method,
			Attributes: req.Action.Attributes,
		},
		Environment: auth.EnvironmentCtx{
			Time:       req.Environment.Time,
			IP:         req.Environment.IP,
			UserAgent:  req.Environment.UserAgent,
			Device:     req.Environment.Device,
			Location:   req.Environment.Location,
			Attributes: req.Environment.Attributes,
		},
		RequestID: req.RequestID,
		Timestamp: time.Now(),
	}

	// 使用ABAC引擎评估
	decision, err := s.abacEngine.Evaluate(ctx, accessRequest)
	if err != nil {
		s.logger.WithError(err).WithField("request_id", req.RequestID).Error("Failed to evaluate policy")
		return nil, fmt.Errorf("failed to evaluate policy: %w", err)
	}

	return &EvaluatePolicyResponse{
		Allowed:     decision.Allowed,
		Effect:      decision.Effect,
		Reason:      decision.Reason,
		PolicyID:    decision.PolicyID,
		PolicyName:  decision.PolicyName,
		Duration:    decision.Duration,
		TTL:         decision.TTL,
		Attributes:  decision.Attributes,
		Obligations: decision.Obligations,
	}, nil
}

// BulkEvaluatePolicies 批量评估策略
func (s *policyService) BulkEvaluatePolicies(ctx context.Context, reqs []*EvaluatePolicyRequest) ([]*EvaluatePolicyResponse, error) {
	var responses []*EvaluatePolicyResponse
	var errors []error

	for _, req := range reqs {
		response, err := s.EvaluatePolicy(ctx, req)
		if err != nil {
			errors = append(errors, err)
			responses = append(responses, &EvaluatePolicyResponse{
				Allowed: false,
				Effect:  "deny",
				Reason:  fmt.Sprintf("Evaluation error: %v", err),
			})
		} else {
			responses = append(responses, response)
		}
	}

	if len(errors) > 0 {
		s.logger.WithField("error_count", len(errors)).Warn("Some policy evaluations failed")
	}

	return responses, nil
}

// 辅助方法
func (s *policyService) buildPolicyResponse(policy *models.Policy) *PolicyResponse {
	response := &PolicyResponse{
		ID:          policy.ID,
		Name:        policy.Name,
		Description: policy.Description,
		Version:     policy.Version,
		Enabled:     policy.Enabled,
		Priority:    policy.Priority,
		Effect:      policy.Effect,
		TenantID:    policy.TenantID,
		Tags:        policy.Tags,
		CreatedBy:   policy.CreatedBy,
		CreatedAt:   policy.CreatedAt,
		UpdatedAt:   policy.UpdatedAt,
	}

	// 解析JSON字段
	if policy.Subject != "" {
		json.Unmarshal([]byte(policy.Subject), &response.Subject)
	}
	if policy.Resource != "" {
		json.Unmarshal([]byte(policy.Resource), &response.Resource)
	}
	if policy.Action != "" {
		json.Unmarshal([]byte(policy.Action), &response.Action)
	}
	if policy.Environment != "" {
		json.Unmarshal([]byte(policy.Environment), &response.Environment)
	}
	if policy.Conditions != "" {
		json.Unmarshal([]byte(policy.Conditions), &response.Conditions)
	}

	return response
}

func (s *policyService) logPolicyOperation(ctx context.Context, operation string, policyID uint, userID uint, tenantID string, details string) {
	auditLog := &models.AuditLog{
		UserID:    userID,
		Action:    fmt.Sprintf("policy_%s", operation),
		Resource:  fmt.Sprintf("policy:%d", policyID),
		Result:    "success",
		Details:   details,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}

	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		s.logger.WithError(err).Error("Failed to log policy operation")
	}
}

func (s *policyService) getPolicyChanges(oldPolicy, newPolicy *models.Policy) []string {
	var changes []string

	if oldPolicy.Name != newPolicy.Name {
		changes = append(changes, fmt.Sprintf("name: %s -> %s", oldPolicy.Name, newPolicy.Name))
	}
	if oldPolicy.Description != newPolicy.Description {
		changes = append(changes, "description changed")
	}
	if oldPolicy.Enabled != newPolicy.Enabled {
		changes = append(changes, fmt.Sprintf("enabled: %v -> %v", oldPolicy.Enabled, newPolicy.Enabled))
	}
	if oldPolicy.Priority != newPolicy.Priority {
		changes = append(changes, fmt.Sprintf("priority: %d -> %d", oldPolicy.Priority, newPolicy.Priority))
	}
	if oldPolicy.Effect != newPolicy.Effect {
		changes = append(changes, fmt.Sprintf("effect: %s -> %s", oldPolicy.Effect, newPolicy.Effect))
	}

	return changes
}

// 占位符方法，需要进一步实现
func (s *policyService) CreatePolicyFromTemplate(ctx context.Context, req *CreatePolicyFromTemplateRequest) (*PolicyResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) GetPolicyTemplates(ctx context.Context, filter *PolicyTemplateFilter) (*PolicyTemplateListResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) CreatePolicyTemplate(ctx context.Context, req *CreatePolicyTemplateRequest) (*PolicyTemplateResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) CreatePolicyVersion(ctx context.Context, policyID uint, req *CreatePolicyVersionRequest) (*PolicyVersionResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) GetPolicyVersions(ctx context.Context, policyID uint) (*PolicyVersionListResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) RollbackPolicyVersion(ctx context.Context, policyID uint, version int) error {
	return fmt.Errorf("not implemented")
}

func (s *policyService) AnalyzePolicyConflicts(ctx context.Context, tenantID string) (*PolicyConflictAnalysis, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) GetPolicyRecommendations(ctx context.Context, tenantID string) (*PolicyRecommendationList, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) GetPolicyStatistics(ctx context.Context, tenantID string) (*PolicyStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) TestPolicy(ctx context.Context, req *TestPolicyRequest) (*TestPolicyResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) ValidatePolicy(ctx context.Context, policy *auth.PolicyRule) (*PolicyValidationResult, error) {
	return &PolicyValidationResult{
		Valid: true,
		Errors: []string{},
		Warnings: []string{},
	}, nil
}

func (s *policyService) ExportPolicies(ctx context.Context, tenantID string, format string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *policyService) ImportPolicies(ctx context.Context, data []byte, format string) (*ImportResult, error) {
	return nil, fmt.Errorf("not implemented")
}