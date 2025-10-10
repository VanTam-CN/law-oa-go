package services

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"law-oa-go/internal/models"
)

// RiskAssessor 风险评估器接口
type RiskAssessor interface {
	// AssessRisk 评估冲突风险
	AssessRisk(ctx context.Context, conflicts []*models.ConflictCase, ruleMatches []*RuleMatch) (*models.RiskAssessment, error)
	// GenerateRecommendations 生成处理建议
	GenerateRecommendations(ctx context.Context, riskLevel string, conflicts []*models.ConflictCase) ([]string, error)
	// CalculateRiskScore 计算风险分数
	CalculateRiskScore(ctx context.Context, conflicts []*models.ConflictCase, ruleMatches []*RuleMatch) (float64, error)
	// GetRiskLevel 获取风险等级
	GetRiskLevel(score float64) string
	// AssessSingleConflict 评估单个冲突
	AssessSingleConflict(ctx context.Context, conflict *models.ConflictCase) (*SingleRiskResult, error)
}

// SingleRiskResult 单个冲突风险评估结果
type SingleRiskResult struct {
	ConflictID     string               `json:"conflictId"`
	RiskScore      float64              `json:"riskScore"`
	RiskLevel      string               `json:"riskLevel"`
	RiskFactors    []*RiskFactor       `json:"riskFactors"`
	Recommendation string               `json:"recommendation"`
	AssessedAt     time.Time            `json:"assessedAt"`
}

// RiskFactor 风险因素
type RiskFactor struct {
	Name         string  `json:"name"`
	Value        float64 `json:"value"`
	Weight       float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
	Description  string  `json:"description"`
}

// RiskConfig 风险评估配置
type RiskConfig struct {
	// 风险等级阈值
	HighRiskThreshold    float64 `json:"highRiskThreshold"    yaml:"highRiskThreshold"`
	MediumRiskThreshold  float64 `json:"mediumRiskThreshold"  yaml:"mediumRiskThreshold"`
	LowRiskThreshold     float64 `json:"lowRiskThreshold"     yaml:"lowRiskThreshold"`

	// 风险因素权重
	CaseTypeWeight      float64 `json:"caseTypeWeight"      yaml:"caseTypeWeight"`
	RiskLevelWeight     float64 `json:"riskLevelWeight"     yaml:"riskLevelWeight"`
	TimeRecencyWeight   float64 `json:"timeRecencyWeight"   yaml:"timeRecencyWeight"`
	RelationWeight      float64 `json:"relationWeight"      yaml:"relationWeight"`

	// 建议配置
	MaxRecommendations int    `json:"maxRecommendations" yaml:"maxRecommendations"`
	EnableAIRecommendations bool `json:"enableAIRecommendations" yaml:"enableAIRecommendations"`
}

// riskAssessor 风险评估器实现
type riskAssessor struct {
	config     *RiskConfig
	ruleEngine interface{}
}

// NewRiskAssessor 创建新的风险评估器
func NewRiskAssessor(config *RiskConfig, ruleEngine interface{}) RiskAssessor {
	if config == nil {
		config = getDefaultRiskConfig()
	}

	return &riskAssessor{
		config:     config,
		ruleEngine: ruleEngine,
	}
}

// AssessRisk 评估冲突风险
func (r *riskAssessor) AssessRisk(ctx context.Context, conflicts []*models.ConflictCase, ruleMatches []*RuleMatch) (*models.RiskAssessment, error) {
	if len(conflicts) == 0 {
		return &models.RiskAssessment{
			OverallRisk:     "LOW",
			RiskScore:       0.0,
			RiskReason:      "未发现冲突",
			RequiresApproval: false,
			RiskFactors:     []string{},
			Mitigation:      []string{},
		}, nil
	}

	// 计算总体风险分数
	totalScore, err := r.CalculateRiskScore(ctx, conflicts, ruleMatches)
	if err != nil {
		return nil, fmt.Errorf("计算风险分数失败: %w", err)
	}

	// 确定风险等级
	riskLevel := r.GetRiskLevel(totalScore)

	// 生成建议
	recommendations, err := r.GenerateRecommendations(ctx, riskLevel, conflicts)
	if err != nil {
		return nil, fmt.Errorf("生成建议失败: %w", err)
	}

	assessment := &models.RiskAssessment{
		OverallRisk:     riskLevel,
		RiskScore:       totalScore,
		RiskReason:      r.generateRiskReason(conflicts, riskLevel),
		RequiresApproval: totalScore > r.config.MediumRiskThreshold,
		ApprovalLevel:   r.getApprovalLevel(riskLevel),
		RiskFactors:     r.generateRiskFactors(conflicts),
		Mitigation:      recommendations,
	}

	return assessment, nil
}

// GenerateRecommendations 生成处理建议
func (r *riskAssessor) GenerateRecommendations(ctx context.Context, riskLevel string, conflicts []*models.ConflictCase) ([]string, error) {
	var recommendations []string

	// 基于风险等级生成建议
	switch riskLevel {
	case "HIGH":
		recommendations = append(recommendations, r.getHighRiskRecommendations()...)
	case "MEDIUM":
		recommendations = append(recommendations, r.getMediumRiskRecommendations()...)
	case "LOW":
		recommendations = append(recommendations, r.getLowRiskRecommendations()...)
	}

	// 基于冲突类型生成建议
	conflictTypes := r.analyzeConflictTypes(conflicts)
	for conflictType, count := range conflictTypes {
		if count > 0 {
			recommendations = append(recommendations, r.getTypeSpecificRecommendations(conflictType)...)
		}
	}

	// 基于时间因素生成建议
	timeRecs := r.getTimeBasedRecommendations(conflicts)
	recommendations = append(recommendations, timeRecs...)

	// 去重并限制数量
	recommendations = r.deduplicateRecommendations(recommendations)
	if len(recommendations) > r.config.MaxRecommendations {
		recommendations = recommendations[:r.config.MaxRecommendations]
	}

	return recommendations, nil
}

// CalculateRiskScore 计算风险分数
func (r *riskAssessor) CalculateRiskScore(ctx context.Context, conflicts []*models.ConflictCase, ruleMatches []*RuleMatch) (float64, error) {
	if len(conflicts) == 0 {
		return 0.0, nil
	}

	var totalScore float64

	// 基于冲突案例计算风险分数
	for _, conflict := range conflicts {
		conflictScore := r.calculateConflictRiskScore(conflict)
		totalScore += conflictScore
	}

	// 基于规则匹配计算风险分数
	for _, match := range ruleMatches {
		if match.Matched {
			totalScore += match.RiskScore
		}
	}

	// 应用时间衰减因子
	timeDecay := r.calculateTimeDecay(conflicts)
	totalScore *= timeDecay

	// 归一化分数
	totalScore = math.Min(totalScore, 100.0) // 最大分数为100

	return totalScore, nil
}

// GetRiskLevel 获取风险等级
func (r *riskAssessor) GetRiskLevel(score float64) string {
	if score >= r.config.HighRiskThreshold {
		return "HIGH"
	} else if score >= r.config.MediumRiskThreshold {
		return "MEDIUM"
	} else if score >= r.config.LowRiskThreshold {
		return "LOW"
	}
	return "MINIMAL"
}

// AssessSingleConflict 评估单个冲突
func (r *riskAssessor) AssessSingleConflict(ctx context.Context, conflict *models.ConflictCase) (*SingleRiskResult, error) {
	// 计算冲突分数
	conflictScore := r.calculateConflictRiskScore(conflict)

	// 确定风险等级
	riskLevel := r.GetRiskLevel(conflictScore)

	// 分析风险因素
	riskFactors := r.analyzeConflictRiskFactors(conflict)

	// 生成建议
	recommendation := r.generateSingleConflictRecommendation(conflict, riskLevel)

	return &SingleRiskResult{
		ConflictID:     conflict.ID,
		RiskScore:      conflictScore,
		RiskLevel:      riskLevel,
		RiskFactors:    riskFactors,
		Recommendation: recommendation,
		AssessedAt:     time.Now(),
	}, nil
}

// calculateConflictRiskScore 计算单个冲突的风险分数
func (r *riskAssessor) calculateConflictRiskScore(conflict *models.ConflictCase) float64 {
	var score float64

	// 基于风险等级
	switch conflict.RiskLevel {
	case "HIGH":
		score += 50 * r.config.RiskLevelWeight
	case "MEDIUM":
		score += 30 * r.config.RiskLevelWeight
	case "LOW":
		score += 10 * r.config.RiskLevelWeight
	}

	// 基于案件类型
	caseTypeScore := r.getCaseTypeRiskScore(conflict.ConflictType)
	score += caseTypeScore * r.config.CaseTypeWeight

	// 基于时间衰减
	timeDecay := r.calculateSingleTimeDecay(conflict.CreatedAt)
	score *= timeDecay

	// 基于对立当事人数量
	opposingPartyCount := len(conflict.OpposingParties)
	score += float64(opposingPartyCount) * 5

	return score
}

// analyzeRiskDistribution 分析风险分布
func (r *riskAssessor) analyzeRiskDistribution(conflicts []*models.ConflictCase) map[string]int {
	distribution := make(map[string]int)

	for _, conflict := range conflicts {
		distribution[conflict.RiskLevel]++
	}

	// 确保所有等级都有计数
	for _, level := range []string{"HIGH", "MEDIUM", "LOW", "MINIMAL"} {
		if _, exists := distribution[level]; !exists {
			distribution[level] = 0
		}
	}

	return distribution
}

// countConflictsByLevel 按风险等级统计冲突数量
func (r *riskAssessor) countConflictsByLevel(conflicts []*models.ConflictCase, level string) int {
	count := 0
	for _, conflict := range conflicts {
		if conflict.RiskLevel == level {
			count++
		}
	}
	return count
}

// analyzeConflictTypes 分析冲突类型
func (r *riskAssessor) analyzeConflictTypes(conflicts []*models.ConflictCase) map[string]int {
	types := make(map[string]int)

	for _, conflict := range conflicts {
		types[conflict.ConflictType]++
	}

	return types
}

// calculateTimeDecay 计算时间衰减因子
func (r *riskAssessor) calculateTimeDecay(conflicts []*models.ConflictCase) float64 {
	if len(conflicts) == 0 {
		return 1.0
	}

	var totalDecay float64
	var count int

	for _, conflict := range conflicts {
		decay := r.calculateSingleTimeDecay(conflict.CreatedAt)
		totalDecay += decay
		count++
	}

	if count == 0 {
		return 1.0
	}

	return totalDecay / float64(count)
}

// calculateSingleTimeDecay 计算单个冲突的时间衰减
func (r *riskAssessor) calculateSingleTimeDecay(createdAt time.Time) float64 {
	now := time.Now()
	hoursPassed := now.Sub(createdAt).Hours()

	// 时间衰减：越近的冲突权重越高
	if hoursPassed < 24 {
		return 1.0 // 24小时内
	} else if hoursPassed < 168 { // 1周内
		return 0.8
	} else if hoursPassed < 720 { // 1个月内
		return 0.6
	} else if hoursPassed < 4320 { // 6个月内
		return 0.4
	} else {
		return 0.2 // 超过6个月
	}
}

// analyzeConflictRiskFactors 分析冲突风险因素
func (r *riskAssessor) analyzeConflictRiskFactors(conflict *models.ConflictCase) []*RiskFactor {
	var factors []*RiskFactor

	// 案件类型风险因素
	caseTypeScore := r.getCaseTypeRiskScore(conflict.ConflictType)
	factors = append(factors, &RiskFactor{
		Name:         "案件类型",
		Value:        caseTypeScore,
		Weight:       r.config.CaseTypeWeight,
		Contribution: caseTypeScore * r.config.CaseTypeWeight,
		Description:  fmt.Sprintf("案件类型 '%s' 的风险分数", conflict.ConflictType),
	})

	// 风险等级因素
	riskLevelScore := r.getRiskLevelScore(conflict.RiskLevel)
	factors = append(factors, &RiskFactor{
		Name:         "风险等级",
		Value:        riskLevelScore,
		Weight:       r.config.RiskLevelWeight,
		Contribution: riskLevelScore * r.config.RiskLevelWeight,
		Description:  fmt.Sprintf("风险等级 '%s' 的分数", conflict.RiskLevel),
	})

	// 时间因素
	timeDecay := r.calculateSingleTimeDecay(conflict.CreatedAt)
	factors = append(factors, &RiskFactor{
		Name:         "时间因素",
		Value:        timeDecay,
		Weight:       r.config.TimeRecencyWeight,
		Contribution: timeDecay * r.config.TimeRecencyWeight,
		Description:  fmt.Sprintf("时间衰减因子: %.2f", timeDecay),
	})

	// 对立当事人数量因素
	opposingPartyCount := len(conflict.OpposingParties)
	factors = append(factors, &RiskFactor{
		Name:         "对立当事人",
		Value:        float64(opposingPartyCount),
		Weight:       r.config.RelationWeight,
		Contribution: float64(opposingPartyCount) * r.config.RelationWeight,
		Description:  fmt.Sprintf("对立当事人数量: %d", opposingPartyCount),
	})

	return factors
}

// getCaseTypeRiskScore 获取案件类型风险分数
func (r *riskAssessor) getCaseTypeRiskScore(caseType string) float64 {
	switch strings.ToUpper(caseType) {
	case "CRIMINAL":
		return 80.0
	case "CIVIL":
		return 60.0
	case "ADMINISTRATIVE":
		return 40.0
	case "COMMERCIAL":
		return 70.0
	case "FAMILY":
		return 30.0
	case "LABOR":
		return 35.0
	case "REAL_ESTATE":
		return 50.0
	case "INTELLECTUAL_PROPERTY":
		return 65.0
	default:
		return 45.0
	}
}

// getRiskLevelScore 获取风险等级分数
func (r *riskAssessor) getRiskLevelScore(riskLevel string) float64 {
	switch riskLevel {
	case "HIGH":
		return 90.0
	case "MEDIUM":
		return 60.0
	case "LOW":
		return 30.0
	default:
		return 10.0
	}
}

// generateSingleConflictRecommendation 生成单个冲突的建议
func (r *riskAssessor) generateSingleConflictRecommendation(conflict *models.ConflictCase, riskLevel string) string {
	switch riskLevel {
	case "HIGH":
		return fmt.Sprintf("高度风险冲突：建议立即回避案件 '%s'，并咨询专业律师意见", conflict.CaseName)
	case "MEDIUM":
		return fmt.Sprintf("中度风险冲突：建议详细审查案件 '%s' 的具体情况，必要时采取回避措施", conflict.CaseName)
	case "LOW":
		return fmt.Sprintf("低度风险冲突：可以继续代理案件 '%s'，但需持续监控风险变化", conflict.CaseName)
	default:
		return fmt.Sprintf("冲突案件 '%s' 风险较低，按常规流程处理", conflict.CaseName)
	}
}

// 高风险建议
func (r *riskAssessor) getHighRiskRecommendations() []string {
	return []string{
		"建议立即回避相关案件",
		"咨询专业律师或律协意见",
		"向客户充分披露潜在冲突",
		"获取客户书面同意",
		"建立完善的隔离措施",
		"定期进行冲突审查",
	}
}

// 中风险建议
func (r *riskAssessor) getMediumRiskRecommendations() []string {
	return []string{
		"进行详细的冲突分析",
		"建立信息隔离墙",
		"限制相关人员接触敏感信息",
		"定期监控冲突情况",
		"向相关方披露冲突信息",
		"考虑是否需要回避",
	}
}

// 低风险建议
func (r *riskAssessor) getLowRiskRecommendations() []string {
	return []string{
		"记录冲突情况",
		"定期复查",
		"保持警惕",
		"建立预防措施",
	}
}

// 基于冲突类型的建议
func (r *riskAssessor) getTypeSpecificRecommendations(conflictType string) []string {
	switch strings.ToUpper(conflictType) {
	case "CRIMINAL":
		return []string{
			"刑事案件冲突风险极高，建议严格回避",
			"注意避免违反职业道德规范",
		}
	case "COMMERCIAL":
		return []string{
			"商业案件需注意商业机密保护",
			"评估经济损失风险",
		}
	case "FAMILY":
		return []string{
			"家庭案件需特别注意情感因素",
			"保护当事人隐私",
		}
	default:
		return []string{
			fmt.Sprintf("针对%s类型案件制定专门处理流程", conflictType),
		}
	}
}

// 基于时间的建议
func (r *riskAssessor) getTimeBasedRecommendations(conflicts []*models.ConflictCase) []string {
	var recommendations []string

	now := time.Now()
	recentConflicts := 0
	oldConflicts := 0

	for _, conflict := range conflicts {
		if now.Sub(conflict.CreatedAt).Hours() < 168 { // 1周内
			recentConflicts++
		} else if now.Sub(conflict.CreatedAt).Hours() > 4320 { // 超过6个月
			oldConflicts++
		}
	}

	if recentConflicts > 0 {
		recommendations = append(recommendations, "近期发现新的冲突情况，建议加强监控")
	}

	if oldConflicts > 0 {
		recommendations = append(recommendations, "存在历史冲突，建议评估是否仍需关注")
	}

	return recommendations
}

// 去重建议
func (r *riskAssessor) deduplicateRecommendations(recommendations []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, rec := range recommendations {
		if !seen[rec] {
			seen[rec] = true
			result = append(result, rec)
		}
	}

	return result
}

// generateRiskReason 生成风险原因
func (r *riskAssessor) generateRiskReason(conflicts []*models.ConflictCase, riskLevel string) string {
	if len(conflicts) == 0 {
		return "未发现冲突"
	}

	conflictTypes := r.analyzeConflictTypes(conflicts)
	reasons := make([]string, 0)

	for conflictType, count := range conflictTypes {
		if count > 0 {
			reasons = append(reasons, fmt.Sprintf("%s类型冲突%d个", conflictType, count))
		}
	}

	return fmt.Sprintf("%s风险：%s", riskLevel, strings.Join(reasons, "，"))
}

// getApprovalLevel 获取审批级别
func (r *riskAssessor) getApprovalLevel(riskLevel string) string {
	switch riskLevel {
	case "HIGH":
		return "PARTNER"
	case "MEDIUM":
		return "SENIOR"
	default:
		return ""
	}
}

// generateRiskFactors 生成风险因素
func (r *riskAssessor) generateRiskFactors(conflicts []*models.ConflictCase) []string {
	factors := make([]string, 0)

	if len(conflicts) == 0 {
		return factors
	}

	// 统计风险因素
	highCount := r.countConflictsByLevel(conflicts, "HIGH")
	mediumCount := r.countConflictsByLevel(conflicts, "MEDIUM")
	lowCount := r.countConflictsByLevel(conflicts, "LOW")

	if highCount > 0 {
		factors = append(factors, fmt.Sprintf("高风险冲突%d个", highCount))
	}
	if mediumCount > 0 {
		factors = append(factors, fmt.Sprintf("中风险冲突%d个", mediumCount))
	}
	if lowCount > 0 {
		factors = append(factors, fmt.Sprintf("低风险冲突%d个", lowCount))
	}

	// 添加时间因素
	now := time.Now()
	recentConflicts := 0
	for _, conflict := range conflicts {
		if now.Sub(conflict.CreatedAt).Hours() < 168 { // 1周内
			recentConflicts++
		}
	}
	if recentConflicts > 0 {
		factors = append(factors, fmt.Sprintf("近期新增冲突%d个", recentConflicts))
	}

	return factors
}

// 获取默认风险配置
func getDefaultRiskConfig() *RiskConfig {
	return &RiskConfig{
		HighRiskThreshold:    70.0,
		MediumRiskThreshold:  40.0,
		LowRiskThreshold:     20.0,
		CaseTypeWeight:      0.3,
		RiskLevelWeight:     0.4,
		TimeRecencyWeight:   0.2,
		RelationWeight:      0.1,
		MaxRecommendations:  8,
		EnableAIRecommendations: false,
	}
}