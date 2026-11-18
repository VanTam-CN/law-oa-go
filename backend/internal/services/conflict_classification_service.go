package services

import (
	"context"
	"fmt"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ConflictClassificationService 冲突分类服务接口
type ConflictClassificationService interface {
	// 主要功能
	ClassifyConflicts(ctx context.Context, conflictResults []interface{}) (*ClassificationResult, error)

	// 冲突类型管理
	GetConflictTypeDetails(ctx context.Context, conflictTypeID string) (*models.ConflictType, error)
	GetConflictTypes(ctx context.Context, filters *ConflictTypeFilters) ([]*models.ConflictType, error)

	// 风险评估
	AssessRiskLevel(ctx context.Context, conflictData *RiskAssessmentRequest) (*RiskAssessmentResult, error)

	// 豁免可能性评估
	EvaluateWaiverPossibility(ctx context.Context, conflictData *WaiverEvaluationRequest) (*WaiverEvaluationResult, error)
}

// ClassificationResult 分类结果
type ClassificationResult struct {
	TotalConflicts  int      `json:"total_conflicts"`
	HighRiskCount   int      `json:"high_risk_count"`
	MediumRiskCount int      `json:"medium_risk_count"`
	LowRiskCount    int      `json:"low_risk_count"`
	Summary         string   `json:"summary"`
	WaiverRequired  bool     `json:"waiver_required"`
	WaiverPossible  bool     `json:"waiver_possible"`
	Recommendations []string `json:"recommendations"`
	NextSteps       []string `json:"next_steps"`
}

// ConflictTypeFilters 冲突类型过滤器
type ConflictTypeFilters struct {
	Category       string `json:"category"`
	RiskLevel      string `json:"risk_level"`
	WaiverPossible *bool  `json:"waiver_possible"`
	Active         *bool  `json:"active"`
}

// RiskAssessmentRequest 风险评估请求
type RiskAssessmentRequest struct {
	ConflictType      string           `json:"conflict_type"`
	SeverityFactors   []SeverityFactor `json:"severity_factors"`
	ClientSensitivity string           `json:"client_sensitivity"`
	MatterImportance  string           `json:"matter_importance"`
}

// SeverityFactor 严重性因子
type SeverityFactor struct {
	Name        string  `json:"name"`
	Impact      string  `json:"impact"`
	Probability string  `json:"probability"`
	Score       float64 `json:"score"`
}

// RiskAssessmentResult 风险评估结果
type RiskAssessmentResult struct {
	OverallRiskLevel     string   `json:"overall_risk_level"`
	RiskScore            float64  `json:"risk_score"`
	RiskFactors          []string `json:"risk_factors"`
	MitigationStrategies []string `json:"mitigation_strategies"`
	PreventiveMeasures   []string `json:"preventive_measures"`
	ReviewSchedule       string   `json:"review_schedule"`
}

// WaiverEvaluationRequest 豁免评估请求
type WaiverEvaluationRequest struct {
	ConflictType       string `json:"conflict_type"`
	SeverityLevel      string `json:"severity_level"`
	ClientConsent      bool   `json:"client_consent"`
	DisclosureLevel    string `json:"disclosure_level"`
	MonitoringRequired bool   `json:"monitoring_required"`
	WaiverPossible     bool   `json:"waiver_possible"`
}

// WaiverEvaluationResult 豁免评估结果
type WaiverEvaluationResult struct {
	WaiverPossible    bool     `json:"waiver_possible"`
	WaiverType        string   `json:"waiver_type"`
	RequiredApprovals []string `json:"required_approvals"`
	MonitoringPlan    string   `json:"monitoring_plan"`
	ExpiryDate        string   `json:"expiry_date"`
	Conditions        []string `json:"conditions"`
}

// conflictClassificationService 冲突分类服务实现
type conflictClassificationService struct {
	repo repositories.EnhancedConflictRepository
}

// NewConflictClassificationService 创建冲突分类服务
func NewConflictClassificationService(repo repositories.EnhancedConflictRepository) ConflictClassificationService {
	return &conflictClassificationService{
		repo: repo,
	}
}

// ClassifyConflicts 分类冲突
func (s *conflictClassificationService) ClassifyConflicts(ctx context.Context, conflictResults []interface{}) (*ClassificationResult, error) {
	result := &ClassificationResult{
		TotalConflicts:  len(conflictResults),
		HighRiskCount:   0,
		MediumRiskCount: 0,
		LowRiskCount:    0,
		Summary:         "冲突检测分类完成",
		WaiverRequired:  false,
		WaiverPossible:  false,
		Recommendations: []string{},
		NextSteps:       []string{},
	}

	// 分析冲突类型和风险级别
	for _, conflict := range conflictResults {
		riskLevel := s.analyzeConflictRisk(conflict)
		switch riskLevel {
		case "HIGH":
			result.HighRiskCount++
			result.WaiverRequired = true
		case "MEDIUM":
			result.MediumRiskCount++
			if !result.WaiverRequired {
				result.WaiverPossible = true
			}
		case "LOW":
			result.LowRiskCount++
		}
	}

	// 生成建议和下一步行动
	if result.WaiverRequired || result.WaiverPossible {
		result.Recommendations = append(result.Recommendations, "建议申请豁免协议")
		result.NextSteps = append(result.NextSteps, "准备豁免申请材料")
	} else {
		result.Recommendations = append(result.Recommendations, "可以正常处理案件")
		result.NextSteps = append(result.NextSteps, "继续案件处理流程")
	}

	return result, nil
}

// GetConflictTypeDetails 获取冲突类型详情
func (s *conflictClassificationService) GetConflictTypeDetails(ctx context.Context, conflictTypeID string) (*models.ConflictType, error) {
	// 临时实现，需要根据实际仓库接口调整
	conflictTypes, err := s.repo.GetActiveConflictTypes(ctx)
	if err != nil {
		return nil, err
	}

	for _, ct := range conflictTypes {
		if ct.ID == conflictTypeID {
			return ct, nil
		}
	}

	return nil, fmt.Errorf("冲突类型不存在: %s", conflictTypeID)
}

// GetConflictTypes 获取冲突类型列表
func (s *conflictClassificationService) GetConflictTypes(ctx context.Context, filters *ConflictTypeFilters) ([]*models.ConflictType, error) {
	if filters != nil && (filters.Category != "" || filters.RiskLevel != "") {
		// 实现过滤逻辑
		return s.repo.GetActiveConflictTypes(ctx)
	}
	return s.repo.GetActiveConflictTypes(ctx)
}

// AssessRiskLevel 评估风险等级
func (s *conflictClassificationService) AssessRiskLevel(ctx context.Context, request *RiskAssessmentRequest) (*RiskAssessmentResult, error) {
	riskScore := s.calculateRiskScore(request)

	riskLevel := "LOW"
	if riskScore >= 8.0 {
		riskLevel = "HIGH"
	} else if riskScore >= 5.0 {
		riskLevel = "MEDIUM"
	}

	result := &RiskAssessmentResult{
		OverallRiskLevel:     riskLevel,
		RiskScore:            riskScore,
		RiskFactors:          []string{},
		MitigationStrategies: []string{"建立信息屏障", "限制信息访问"},
		PreventiveMeasures:   []string{"定期审查", "监控合规性"},
		ReviewSchedule:       "季度审查",
	}

	// 添加风险因子
	for _, factor := range request.SeverityFactors {
		result.RiskFactors = append(result.RiskFactors, factor.Name)
	}

	return result, nil
}

// EvaluateWaiverPossibility 评估豁免可能性
func (s *conflictClassificationService) EvaluateWaiverPossibility(ctx context.Context, request *WaiverEvaluationRequest) (*WaiverEvaluationResult, error) {
	waiverPossible := s.evaluateWaiverEligibility(request)

	waiverType := "INFORMED_CONSENT"
	if request.SeverityLevel == "HIGH" {
		waiverType = "WRITTEN_CONSENT"
	}

	result := &WaiverEvaluationResult{
		WaiverPossible:    waiverPossible,
		WaiverType:        waiverType,
		RequiredApprovals: []string{"主办律师", "合伙人"},
		MonitoringPlan:    "月度监控",
		ExpiryDate:        "1年后",
		Conditions:        []string{"完全信息披露", "定期审查"},
	}

	if !request.ClientConsent {
		result.RequiredApprovals = append(result.RequiredApprovals, "客户同意")
	}

	return result, nil
}

// 内部辅助方法

// analyzeConflictRisk 分析冲突风险
func (s *conflictClassificationService) analyzeConflictRisk(conflict interface{}) string {
	// 简化的风险分析逻辑
	// 在实际实现中，这里会根据冲突的具体内容进行详细分析
	return "MEDIUM"
}

// calculateRiskScore 计算风险分数
func (s *conflictClassificationService) calculateRiskScore(request *RiskAssessmentRequest) float64 {
	score := 3.0 // 基础分数

	// 根据严重性因子调整分数
	for _, factor := range request.SeverityFactors {
		score += factor.Score
	}

	// 根据客户敏感度调整
	if request.ClientSensitivity == "HIGH" {
		score += 2.0
	} else if request.ClientSensitivity == "MEDIUM" {
		score += 1.0
	}

	// 根据案件重要性调整
	if request.MatterImportance == "HIGH" {
		score += 1.5
	}

	if score > 10.0 {
		score = 10.0
	}

	return score
}

// evaluateWaiverEligibility 评估豁免资格
func (s *conflictClassificationService) evaluateWaiverEligibility(request *WaiverEvaluationRequest) bool {
	// 简化的豁免资格评估
	if request.SeverityLevel == "CRITICAL" {
		return false
	}

	if request.ClientConsent && request.DisclosureLevel == "FULL" {
		return true
	}

	return request.WaiverPossible // 假设请求中有这个字段
}
