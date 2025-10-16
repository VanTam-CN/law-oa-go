package repositories

import (
	"context"
	"time"

	"law-oa-go/internal/models"
)

// AdvancedConflictCheckRequest 高级冲突检查请求
type AdvancedConflictCheckRequest struct {
	LawyerID       int    `json:"lawyer_id" binding:"required"`
	ClientName     string `json:"client_name" binding:"required"`
	ClientType     string `json:"client_type" binding:"required"`
	ClientIndustry string `json:"client_industry"`
	OpposingParty  string `json:"opposing_party"`
	CaseType       string `json:"case_type" binding:"required"`
	SearchDepth    string `json:"search_depth"`
	IncludeRelated bool   `json:"include_related"`
	// 新增字段支持更复杂的冲突检测
	SearchYears     int       `json:"search_years"`
	TeamMembers     []string  `json:"team_members"`
	Description     string    `json:"description"`
	// 兼容旧字段
	IndustryID      int       `json:"industry_id"`
	IncludeIndustry bool      `json:"include_industry"`
	// 审计和追踪字段
	UserID         string    `json:"user_id"`
	RequestTime     time.Time `json:"request_time"`
}

// ConflictCaseDetail 冲突案件详情
type ConflictCaseDetail struct {
	Case            *models.Case `json:"case"`
	ConflictType    string     `json:"conflict_type"`
	RiskScore       int        `json:"risk_score"`
	ConflictReason  string     `json:"conflict_reason"`
	RelatedEntities []string   `json:"related_entities"`
	ImpactLevel     string     `json:"impact_level"`
}

// CompetitionAnalysis 竞争分析结果
type CompetitionAnalysis struct {
	HasCompetition    bool                   `json:"has_competition"`
	CompetitorInfo    []*CompetitorInfo       `json:"competitor_info"`
	IndustryAnalysis map[string]interface{} `json:"industry_analysis"`
	MarketPosition    string                 `json:"market_position"`
	RiskFactors       []string               `json:"risk_factors"`
}

// CompetitorInfo 竞争者信息
type CompetitorInfo struct {
	CompanyName     string   `json:"company_name"`
	Industry        string   `json:"industry"`
	ConflictLevel   string   `json:"conflict_level"`
	MatchedKeywords []string `json:"matched_keywords"`
	RelationType    string   `json:"relation_type"`
	RiskScore       int      `json:"risk_score"`
}

// AnalysisSummary 分析摘要
type AnalysisSummary struct {
	TotalCasesChecked      int `json:"total_cases_checked"`
	DirectConflicts        int `json:"direct_conflicts"`
	IndustryConflicts      int `json:"industry_conflicts"`
	NameSimilarityCases    int `json:"name_similarity_cases"`
	RelatedConflicts       int `json:"related_conflicts"`
	SearchScope            string `json:"search_scope"`
	SearchTimeRange        string `json:"search_time_range"`
}

// ConflictAnalysisResult 冲突分析结果
type ConflictAnalysisResult struct {
	RequestID        string                `json:"request_id"`
	HasConflicts     bool                  `json:"has_conflicts"`
	ConflictLevel   string                `json:"conflict_level"`
	RiskScore       int                   `json:"risk_score"`
	ConflictCases   []*ConflictCaseDetail  `json:"conflict_cases"`
	CompetitionAnalysis *CompetitionAnalysis `json:"competition_analysis"`
	Recommendations  []string              `json:"recommendations"`
	AnalysisSummary  *AnalysisSummary      `json:"analysis_summary"`
	DetectionTime    time.Time             `json:"detection_time"`
	Duration         int64                 `json:"duration_ms"`
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid   bool                    `json:"is_valid"`
	Errors    []ValidationError       `json:"errors"`
	Warnings  []ValidationError       `json:"warnings"`
	Request   *AdvancedConflictCheckRequest `json:"request"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// RiskAssessment 风险评估结果
type RiskAssessment struct {
	OverallRisk     string    `json:"overall_risk"`
	RiskScore       int       `json:"risk_score"`
	RiskFactors     []string  `json:"risk_factors"`
	Recommendations []string  `json:"recommendations"`
	AssessmentTime  time.Time `json:"assessment_time"`
}

// RuleApplication 规则应用结果
type RuleApplication struct {
	RuleID          string   `json:"rule_id"`
	RuleName        string   `json:"rule_name"`
	RuleType        string   `json:"rule_type"`
	Priority        int      `json:"priority"`
	Matched         bool     `json:"matched"`
	RiskFactors     []string `json:"risk_factors"`
	Recommendations []string `json:"recommendations"`
	Conflicts       bool     `json:"conflicts"`
}

// RuleApplicationResult 规则应用总结果
type RuleApplicationResult struct {
	AppliedRules    []*RuleApplication `json:"applied_rules"`
	AdditionalRisks []string           `json:"additional_risks"`
	Recommendations []string           `json:"recommendations"`
	RuleConflicts   []string           `json:"rule_conflicts"`
	ApplicationTime time.Time          `json:"application_time"`
}

// BuiltInRuleResult 内置规则结果
type BuiltInRuleResult struct {
	Risks          []string `json:"risks"`
	Recommendations []string `json:"recommendations"`
}

// EnhancedRepositoryInterface 增强仓储接口
type EnhancedRepositoryInterface interface {
	// 基础冲突检测操作
	GetPotentialConflicts(ctx context.Context, lawyerID int, clientName string, opposingParty string) ([]*models.Case, error)
	GetPotentialConflictsAdvanced(ctx context.Context, request *AdvancedConflictCheckRequest) ([]*models.Case, error)

	// 行业相关操作
	GetIndustryByClientName(ctx context.Context, clientName string) (*models.IndustryClassification, error)
	GetCompetitiveRelationsByIndustry(ctx context.Context, industryID int) ([]models.CompetitiveRelation, error)

	// 规则管理操作
	GetActiveConflictRules(ctx context.Context) ([]models.EnhancedConflictRule, error)
}

// ValidateConflictRequest 验证冲突检查请求
func ValidateConflictRequest(req *AdvancedConflictCheckRequest) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
		Request:  req,
	}

	// 必填字段验证
	if req.LawyerID <= 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "lawyer_id",
			Message: "律师ID必须大于0",
			Code:    "INVALID_LAWYER_ID",
		})
		result.IsValid = false
	}

	if req.ClientName == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "client_name",
			Message: "客户名称不能为空",
			Code:    "MISSING_CLIENT_NAME",
		})
		result.IsValid = false
	}

	if req.CaseType == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "case_type",
			Message: "案件类型不能为空",
			Code:    "MISSING_CASE_TYPE",
		})
		result.IsValid = false
	}

	// 业务逻辑验证
	if req.SearchYears <= 0 || req.SearchYears > 20 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "search_years",
			Message: "搜索年限建议在1-20年之间",
			Code:    "INVALID_SEARCH_YEARS",
		})
	}

	if req.SearchDepth == "" {
		req.SearchDepth = "comprehensive" // 默认值
	}

	return result
}