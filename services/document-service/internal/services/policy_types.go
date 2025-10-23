package services

import (
	"time"

	"github.com/law-oa-go/document-service/internal/auth"
	"github.com/law-oa-go/document-service/internal/repositories"
)

// CreatePolicyRequest 创建策略请求
type CreatePolicyRequest struct {
	Name        string                   `json:"name" validate:"required,min=1,max=100"`
	Description string                   `json:"description" validate:"max=500"`
	Enabled     bool                     `json:"enabled"`
	Priority    int                      `json:"priority" validate:"min=0,max=1000"`
	Effect      string                   `json:"effect" validate:"required,oneof=allow deny"`
	Subject     auth.SubjectMatcher       `json:"subject" validate:"required"`
	Resource    auth.ResourceMatcher      `json:"resource" validate:"required"`
	Action      auth.ActionMatcher        `json:"action" validate:"required"`
	Environment auth.EnvironmentMatcher   `json:"environment"`
	Conditions  []auth.ConditionExpression `json:"conditions"`
	Tags        []string                 `json:"tags"`
	TenantID    string                   `json:"tenant_id" validate:"required"`
	CreatedBy   uint                     `json:"created_by" validate:"required"`
}

// UpdatePolicyRequest 更新策略请求
type UpdatePolicyRequest struct {
	Name        string                        `json:"name" validate:"omitempty,min=1,max=100"`
	Description string                        `json:"description" validate:"omitempty,max=500"`
	Enabled     *bool                         `json:"enabled"`
	Priority    *int                          `json:"priority" validate:"omitempty,min=0,max=1000"`
	Effect      string                        `json:"effect" validate:"omitempty,oneof=allow deny"`
	Subject     *auth.SubjectMatcher           `json:"subject"`
	Resource    *auth.ResourceMatcher          `json:"resource"`
	Action      *auth.ActionMatcher            `json:"action"`
	Environment *auth.EnvironmentMatcher       `json:"environment"`
	Conditions  []auth.ConditionExpression     `json:"conditions"`
	Tags        []string                      `json:"tags"`
	UpdatedBy   uint                          `json:"updated_by" validate:"required"`
}

// PolicyFilter 策略过滤器
type PolicyFilter struct {
	TenantID     string                        `json:"tenant_id" validate:"required"`
	Name         string                        `json:"name"`
	Description  string                        `json:"description"`
	Enabled      *bool                         `json:"enabled"`
	ResourceType string                        `json:"resource_type"`
	ActionType   string                        `json:"action_type"`
	SubjectType  string                        `json:"subject_type"`
	CreatorID    *uint                         `json:"creator_id"`
	Tags         []string                      `json:"tags"`
	CreatedFrom  *time.Time                    `json:"created_from"`
	CreatedTo    *time.Time                    `json:"created_to"`
	UpdatedFrom  *time.Time                    `json:"updated_from"`
	UpdatedTo    *time.Time                    `json:"updated_to"`
	Pagination   *repositories.Pagination      `json:"pagination"`
	SortBy       string                        `json:"sort_by"`
	SortOrder    string                        `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// PolicyResponse 策略响应
type PolicyResponse struct {
	ID          uint                        `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Version     int                         `json:"version"`
	Enabled     bool                        `json:"enabled"`
	Priority    int                         `json:"priority"`
	Effect      string                      `json:"effect"`
	Subject     auth.SubjectMatcher         `json:"subject"`
	Resource    auth.ResourceMatcher        `json:"resource"`
	Action      auth.ActionMatcher          `json:"action"`
	Environment auth.EnvironmentMatcher     `json:"environment"`
	Conditions  []auth.ConditionExpression  `json:"conditions"`
	Tags        []string                    `json:"tags"`
	TenantID    string                      `json:"tenant_id"`
	CreatedBy   uint                        `json:"created_by"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

// PolicyListResponse 策略列表响应
type PolicyListResponse struct {
	Policies []*PolicyResponse `json:"policies"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// EvaluatePolicyRequest 评估策略请求
type EvaluatePolicyRequest struct {
	RequestID   string                        `json:"request_id" validate:"required"`
	Subject     EvaluateSubjectRequest         `json:"subject" validate:"required"`
	Resource    EvaluateResourceRequest        `json:"resource" validate:"required"`
	Action      EvaluateActionRequest          `json:"action" validate:"required"`
	Environment EvaluateEnvironmentRequest     `json:"environment"`
}

// EvaluateSubjectRequest 评估主体请求
type EvaluateSubjectRequest struct {
	ID         uint                   `json:"id" validate:"required"`
	Username   string                 `json:"username" validate:"required"`
	Email      string                 `json:"email" validate:"required,email"`
	Roles      []string               `json:"roles"`
	Groups     []string               `json:"groups"`
	Attributes map[string]interface{} `json:"attributes"`
	TenantID   string                 `json:"tenant_id" validate:"required"`
	Active     bool                   `json:"active"`
}

// EvaluateResourceRequest 评估资源请求
type EvaluateResourceRequest struct {
	Type         string                 `json:"type" validate:"required"`
	ID           string                 `json:"id" validate:"required"`
	Owner        string                 `json:"owner"`
	TenantID     string                 `json:"tenant_id" validate:"required"`
	Attributes   map[string]interface{} `json:"attributes"`
	Sensitivity  string                 `json:"sensitivity"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// EvaluateActionRequest 评估动作请求
type EvaluateActionRequest struct {
	Type       string                 `json:"type" validate:"required"`
	Method     string                 `json:"method"`
	Attributes map[string]interface{} `json:"attributes"`
}

// EvaluateEnvironmentRequest 评估环境请求
type EvaluateEnvironmentRequest struct {
	Time       time.Time              `json:"time"`
	IP         string                 `json:"ip"`
	UserAgent  string                 `json:"user_agent"`
	Device     string                 `json:"device"`
	Location   string                 `json:"location"`
	Attributes map[string]interface{} `json:"attributes"`
}

// EvaluatePolicyResponse 评估策略响应
type EvaluatePolicyResponse struct {
	Allowed     bool                      `json:"allowed"`
	Effect      string                    `json:"effect"`
	Reason      string                    `json:"reason"`
	PolicyID    string                    `json:"policy_id,omitempty"`
	PolicyName  string                    `json:"policy_name,omitempty"`
	Duration    time.Duration             `json:"duration"`
	TTL         time.Duration             `json:"ttl"`
	Attributes  map[string]interface{}    `json:"attributes"`
	Obligations []auth.Obligation         `json:"obligations"`
}

// CreatePolicyFromTemplateRequest 从模板创建策略请求
type CreatePolicyFromTemplateRequest struct {
	TemplateID uint                        `json:"template_id" validate:"required"`
	Name       string                      `json:"name" validate:"required,min=1,max=100"`
	Parameters map[string]interface{}      `json:"parameters"`
	TenantID   string                      `json:"tenant_id" validate:"required"`
	CreatedBy  uint                        `json:"created_by" validate:"required"`
}

// PolicyTemplateFilter 策略模板过滤器
type PolicyTemplateFilter struct {
	Category    string                     `json:"category"`
	Name        string                     `json:"name"`
	Public      *bool                      `json:"public"`
	Enabled     *bool                      `json:"enabled"`
	Tags        []string                   `json:"tags"`
	Pagination  *repositories.Pagination   `json:"pagination"`
	SortBy      string                     `json:"sort_by"`
	SortOrder   string                     `json:"sort_order" validate:"omitempty,oneof=asc desc"`
}

// PolicyTemplateResponse 策略模板响应
type PolicyTemplateResponse struct {
	ID          uint                        `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Category    string                      `json:"category"`
	Subject     auth.SubjectMatcher         `json:"subject"`
	Resource    auth.ResourceMatcher        `json:"resource"`
	Action      auth.ActionMatcher          `json:"action"`
	Environment auth.EnvironmentMatcher     `json:"environment"`
	Conditions  []auth.ConditionExpression  `json:"conditions"`
	Tags        []string                    `json:"tags"`
	Public      bool                        `json:"public"`
	Enabled     bool                        `json:"enabled"`
	CreatedBy   uint                        `json:"created_by"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

// PolicyTemplateListResponse 策略模板列表响应
type PolicyTemplateListResponse struct {
	Templates []*PolicyTemplateResponse `json:"templates"`
	Total     int64                     `json:"total"`
	Page      int                       `json:"page"`
	PageSize  int                       `json:"page_size"`
}

// CreatePolicyTemplateRequest 创建策略模板请求
type CreatePolicyTemplateRequest struct {
	Name        string                        `json:"name" validate:"required,min=1,max=100"`
	Description string                        `json:"description" validate:"max=500"`
	Category    string                        `json:"category" validate:"required"`
	Subject     auth.SubjectMatcher           `json:"subject" validate:"required"`
	Resource    auth.ResourceMatcher          `json:"resource" validate:"required"`
	Action      auth.ActionMatcher            `json:"action" validate:"required"`
	Environment auth.EnvironmentMatcher       `json:"environment"`
	Conditions  []auth.ConditionExpression    `json:"conditions"`
	Tags        []string                      `json:"tags"`
	Public      bool                          `json:"public"`
	CreatedBy   uint                          `json:"created_by" validate:"required"`
}

// CreatePolicyVersionRequest 创建策略版本请求
type CreatePolicyVersionRequest struct {
	Description string                      `json:"description" validate:"max=500"`
	Enabled     bool                        `json:"enabled"`
	CreatedBy   uint                        `json:"created_by" validate:"required"`
}

// PolicyVersionResponse 策略版本响应
type PolicyVersionResponse struct {
	ID          uint      `json:"id"`
	PolicyID    uint      `json:"policy_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"`
	Effect      string    `json:"effect"`
	Subject     string    `json:"subject"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Environment string    `json:"environment"`
	Conditions  string    `json:"conditions"`
	Tags        []string  `json:"tags"`
	TenantID    string    `json:"tenant_id"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PolicyVersionListResponse 策略版本列表响应
type PolicyVersionListResponse struct {
	Versions []*PolicyVersionResponse `json:"versions"`
	Total    int64                    `json:"total"`
}

// PolicyConflictAnalysis 策略冲突分析
type PolicyConflictAnalysis struct {
	Conflicts      []PolicyConflictDetail `json:"conflicts"`
	TotalConflicts int                    `json:"total_conflicts"`
	AnalyzedAt     time.Time              `json:"analyzed_at"`
	Recommendations []string              `json:"recommendations"`
}

// PolicyConflictDetail 策略冲突详情
type PolicyConflictDetail struct {
	ConflictID   uint      `json:"conflict_id"`
	Policy1ID    uint      `json:"policy1_id"`
	Policy1Name  string    `json:"policy1_name"`
	Policy2ID    uint      `json:"policy2_id"`
	Policy2Name  string    `json:"policy2_name"`
	ConflictType string    `json:"conflict_type"`
	Description  string    `json:"description"`
	Severity     string    `json:"severity"`
	Status       string    `json:"status"`
	ResolvedBy   *uint     `json:"resolved_by"`
	ResolvedAt   *time.Time `json:"resolved_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// PolicyRecommendationList 策略推荐列表
type PolicyRecommendationList struct {
	Recommendations  []PolicyRecommendation `json:"recommendations"`
	TotalRecommendations int                `json:"total_recommendations"`
	GeneratedAt      time.Time             `json:"generated_at"`
}

// PolicyRecommendation 策略推荐
type PolicyRecommendation struct {
	ID            uint                   `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Subject       auth.SubjectMatcher    `json:"subject"`
	Resource      auth.ResourceMatcher   `json:"resource"`
	Action        auth.ActionMatcher     `json:"action"`
	Environment   auth.EnvironmentMatcher `json:"environment"`
	Conditions    []auth.ConditionExpression `json:"conditions"`
	Priority      int                    `json:"priority"`
	Effect        string                 `json:"effect"`
	Justification string                 `json:"justification"`
	Tags          []string               `json:"tags"`
	Enabled       bool                   `json:"enabled"`
	TenantID      string                 `json:"tenant_id"`
	CreatedBy     uint                   `json:"created_by"`
	CreatedAt     time.Time              `json:"created_at"`
}

// PolicyStatistics 策略统计
type PolicyStatistics struct {
	TotalPolicies       int64                     `json:"total_policies"`
	EnabledPolicies     int64                     `json:"enabled_policies"`
	DisabledPolicies    int64                     `json:"disabled_policies"`
	PoliciesByCategory  map[string]int64          `json:"policies_by_category"`
	PoliciesByEffect    map[string]int64          `json:"policies_by_effect"`
	PoliciesByResource  map[string]int64          `json:"policies_by_resource"`
	RecentEvaluations   []EvaluationStatistic     `json:"recent_evaluations"`
	EvaluationMetrics   EvaluationMetrics          `json:"evaluation_metrics"`
	GeneratedAt         time.Time                 `json:"generated_at"`
}

// EvaluationStatistic 评估统计
type EvaluationStatistic struct {
	Date           time.Time `json:"date"`
	TotalEvaluations int64    `json:"total_evaluations"`
	AllowedCount   int64     `json:"allowed_count"`
	DeniedCount    int64     `json:"denied_count"`
	AverageDuration float64  `json:"average_duration_ms"`
}

// EvaluationMetrics 评估指标
type EvaluationMetrics struct {
	TotalEvaluations     int64   `json:"total_evaluations"`
	AllowedEvaluations   int64   `json:"allowed_evaluations"`
	DeniedEvaluations    int64   `json:"denied_evaluations"`
	AllowanceRate        float64 `json:"allowance_rate"`
	AverageEvaluationTime float64 `json:"average_evaluation_time_ms"`
	CacheHitRate         float64 `json:"cache_hit_rate"`
	PolicyHitRate        map[string]float64 `json:"policy_hit_rate"`
}

// TestPolicyRequest 测试策略请求
type TestPolicyRequest struct {
	Policy       auth.PolicyRule               `json:"policy" validate:"required"`
	TestCases    []TestCase                    `json:"test_cases" validate:"required,min=1"`
	ValidateOnly bool                          `json:"validate_only"`
}

// TestCase 测试用例
type TestCase struct {
	Name        string                        `json:"name" validate:"required"`
	Description string                        `json:"description"`
	Subject     auth.UserContext              `json:"subject" validate:"required"`
	Resource    auth.ResourceContext          `json:"resource" validate:"required"`
	Action      auth.ActionContext            `json:"action" validate:"required"`
	Environment auth.EnvironmentCtx           `json:"environment"`
	Expected    TestExpectedResult            `json:"expected" validate:"required"`
}

// TestExpectedResult 测试期望结果
type TestExpectedResult struct {
	Allowed    bool                   `json:"allowed"`
	Effect     string                 `json:"effect"`
	Reason     string                 `json:"reason"`
	PolicyID   string                 `json:"policy_id,omitempty"`
	TTL        time.Duration          `json:"ttl"`
	Attributes map[string]interface{} `json:"attributes"`
}

// TestPolicyResponse 测试策略响应
type TestPolicyResponse struct {
	Validation  *PolicyValidationResult `json:"validation"`
	TestResults []TestResult            `json:"test_results"`
	Summary     TestSummary             `json:"summary"`
	TestedAt    time.Time              `json:"tested_at"`
}

// TestResult 测试结果
type TestResult struct {
	TestCase    TestCase               `json:"test_case"`
	Actual      auth.AccessDecision    `json:"actual"`
	Expected    TestExpectedResult     `json:"expected"`
	Passed      bool                   `json:"passed"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
}

// TestSummary 测试摘要
type TestSummary struct {
	TotalTests   int     `json:"total_tests"`
	PassedTests  int     `json:"passed_tests"`
	FailedTests  int     `json:"failed_tests"`
	PassRate     float64 `json:"pass_rate"`
	TotalDuration time.Duration `json:"total_duration"`
}

// PolicyValidationResult 策略验证结果
type PolicyValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Suggestions []string `json:"suggestions"`
}

// ImportResult 导入结果
type ImportResult struct {
	TotalPolicies   int                    `json:"total_policies"`
	ImportedPolicies int                    `json:"imported_policies"`
	SkippedPolicies  []SkipPolicyResult     `json:"skipped_policies"`
	ValidationErrors []ValidationError      `json:"validation_errors"`
	ImportedAt       time.Time             `json:"imported_at"`
}

// SkipPolicyResult 跳过策略结果
type SkipPolicyResult struct {
	PolicyName string `json:"policy_name"`
	Reason     string `json:"reason"`
}

// ValidationError 验证错误
type ValidationError struct {
	PolicyName string   `json:"policy_name"`
	Field      string   `json:"field"`
	Message    string   `json:"message"`
	Errors     []string `json:"errors"`
}