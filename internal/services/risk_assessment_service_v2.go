package services

import (
	"context"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// RiskAssessmentServiceV2 重构的风险评估服务
type RiskAssessmentServiceV2 struct {
	repo   repositories.EnhancedRepositoryInterface
	logger *log.Logger
}

// NewRiskAssessmentServiceV2 创建新的风险评估服务
func NewRiskAssessmentServiceV2(repo repositories.EnhancedRepositoryInterface) *RiskAssessmentServiceV2 {
	return &RiskAssessmentServiceV2{
		repo:   repo,
		logger: log.Default(),
	}
}

// AssessRisk 评估风险
func (s *RiskAssessmentServiceV2) AssessRisk(ctx context.Context, conflictCases []*models.Case, request *repositories.AdvancedConflictCheckRequest) (*repositories.RiskAssessment, error) {
	assessment := &repositories.RiskAssessment{
		OverallRisk:    "LOW",
		RiskScore:      0,
		RiskFactors:    make([]string, 0),
		Recommendations: make([]string, 0),
		AssessmentTime: time.Now(),
	}

	if len(conflictCases) == 0 {
		assessment.OverallRisk = "NONE"
		assessment.RiskScore = 0
		assessment.Recommendations = append(assessment.Recommendations, "✅ 未发现明显风险")
		return assessment, nil
	}

	// 计算基础风险分数
	baseScore := s.calculateBaseRiskScore(conflictCases)

	// 根据案件类型调整
	caseTypeMultiplier := s.getCaseTypeMultiplier(request.CaseType)

	// 根据搜索深度调整
	depthMultiplier := s.getSearchDepthMultiplier(request.SearchDepth)

	assessment.RiskScore = int(float64(baseScore) * caseTypeMultiplier * depthMultiplier)

	// 确定风险等级
	assessment.OverallRisk = s.determineRiskLevel(assessment.RiskScore)

	// 生成风险因素
	assessment.RiskFactors = s.generateRiskFactors(conflictCases, assessment.RiskScore)

	// 生成建议
	assessment.Recommendations = s.generateRiskRecommendations(assessment.RiskScore, assessment.OverallRisk)

	s.logger.Printf("风险评估完成 - 案件数: %d, 风险分数: %d, 风险等级: %s",
		len(conflictCases), assessment.RiskScore, assessment.OverallRisk)

	return assessment, nil
}

// calculateBaseRiskScore 计算基础风险分数
func (s *RiskAssessmentServiceV2) calculateBaseRiskScore(conflictCases []*models.Case) int {
	if len(conflictCases) == 0 {
		return 0
	}

	totalScore := 0
	for _, case_ := range conflictCases {
		// 根据案件状态计算分数
		switch case_.Status {
		case "in_progress":
			totalScore += 80
		case "completed":
			totalScore += 60
		case "pending":
			totalScore += 40
		default:
			totalScore += 30
		}
	}

	// 如果有多个冲突案件，增加额外风险
	if len(conflictCases) > 1 {
		multiplier := 1.0 + (float64(len(conflictCases)-1) * 0.2)
		totalScore = int(float64(totalScore) * multiplier)
	}

	return totalScore
}

// getCaseTypeMultiplier 获取案件类型乘数
func (s *RiskAssessmentServiceV2) getCaseTypeMultiplier(caseType string) float64 {
	switch caseType {
	case "criminal":
		return 1.5
	case "civil":
		return 1.2
	case "commercial":
		return 1.1
	case "administrative":
		return 1.0
	case "arbitration":
		return 1.1
	case "consultation":
		return 0.8
	default:
		return 1.0
	}
}

// getSearchDepthMultiplier 获取搜索深度乘数
func (s *RiskAssessmentServiceV2) getSearchDepthMultiplier(searchDepth string) float64 {
	switch searchDepth {
	case "comprehensive":
		return 1.2
	case "standard":
		return 1.0
	case "basic":
		return 0.8
	default:
		return 1.0
	}
}

// determineRiskLevel 确定风险等级
func (s *RiskAssessmentServiceV2) determineRiskLevel(score int) string {
	switch {
	case score >= 80:
		return "HIGH"
	case score >= 60:
		return "MEDIUM"
	case score >= 40:
		return "LOW"
	case score >= 20:
		return "MINIMAL"
	default:
		return "NONE"
	}
}

// generateRiskFactors 生成风险因素
func (s *RiskAssessmentServiceV2) generateRiskFactors(conflictCases []*models.Case, riskScore int) []string {
	factors := make([]string, 0)

	if len(conflictCases) > 0 {
		factors = append(factors, "发现潜在利益冲突")
	}

	if len(conflictCases) > 1 {
		factors = append(factors, "多个关联案件增加风险")
	}

	for _, case_ := range conflictCases {
		if case_.Status == "in_progress" {
			factors = append(factors, "存在进行中的相关案件")
			break
		}
	}

	if riskScore >= 80 {
		factors = append(factors, "高风险等级，需要详细审查")
	}

	return factors
}

// generateRiskRecommendations 生成风险建议
func (s *RiskAssessmentServiceV2) generateRiskRecommendations(riskScore int, riskLevel string) []string {
	recommendations := make([]string, 0)

	switch riskLevel {
	case "HIGH":
		recommendations = append(recommendations, "🚨 高风险案件，强烈建议避免代理")
		recommendations = append(recommendations, "📋 建议咨询律师事务所管理委员会")
		recommendations = append(recommendations, "🔍 需要进行详细的利益冲突审查")
	case "MEDIUM":
		recommendations = append(recommendations, "⚠️ 中等风险，建议谨慎考虑")
		recommendations = append(recommendations, "📝 建议建立信息墙机制")
		recommendations = append(recommendations, "👥 建议与相关方进行充分沟通")
	case "LOW":
		recommendations = append(recommendations, "✅ 风险较低，但仍需注意")
		recommendations = append(recommendations, "📊 建议定期监控案件进展")
	case "MINIMAL":
		recommendations = append(recommendations, "✅ 风险极小，可以正常进行")
		recommendations = append(recommendations, "📝 建议做好基本记录备案")
	default:
		recommendations = append(recommendations, "✅ 未发现明显风险")
	}

	return recommendations
}