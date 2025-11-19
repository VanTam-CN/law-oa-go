package services

import (
	"context"
	"fmt"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// EnhancedConflictDetectionService 增强冲突检测服务接口
type EnhancedConflictDetectionService interface {
	// 主要检测方法
	PerformMultiDimensionalDetection(ctx context.Context, requestID string, clientProfiles []*models.ClientProfile) ([]*ConflictDetectionResult, error)
	DetectConflictsForCase(ctx context.Context, caseID uint, clientIDs []string) ([]*ConflictDetectionResult, error)

	// 历史和分析
	GetConflictHistory(ctx context.Context, params *ConflictHistoryParams) ([]*ConflictHistoryRecord, error)
	AnalyzeConflictTrends(ctx context.Context, timeframe string) (*ConflictTrendAnalysis, error)
}

// ConflictDetectionResult 冲突检测结果
type ConflictDetectionResult struct {
	ID                string                 `json:"id"`
	ConflictType      string                 `json:"conflict_type"`
	SourceEntity      string                 `json:"source_entity"`
	RelatedEntity     string                 `json:"related_entity"`
	Description       string                 `json:"description"`
	SeverityLevel     string                 `json:"severity_level"`
	RiskScore         float64                `json:"risk_score"`
	DetectionDate     string                 `json:"detection_date"`
	Status            string                 `json:"status"`
	WaiverPossible    bool                   `json:"waiver_possible"`
	Recommendations   []string               `json:"recommendations"`
}

// ConflictHistoryParams 冲突历史查询参数
type ConflictHistoryParams struct {
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	ConflictType   string `json:"conflict_type"`
	Severity       string `json:"severity"`
	ClientID       string `json:"client_id"`
	LawyerID       string `json:"lawyer_id"`
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
}

// ConflictHistoryRecord 冲突历史记录
type ConflictHistoryRecord struct {
	ID              string `json:"id"`
	CheckRequestID  string `json:"check_request_id"`
	ConflictType    string `json:"conflict_type"`
	Description     string `json:"description"`
	Severity        string `json:"severity"`
	DetectionDate   string `json:"detection_date"`
	Status          string `json:"status"`
	ResolvedDate    string `json:"resolved_date,omitempty"`
}

// ConflictTrendAnalysis 冲突趋势分析
type ConflictTrendAnalysis struct {
	Timeframe     string                         `json:"timeframe"`
	TotalConflicts int                          `json:"total_conflicts"`
	Trends        []ConflictTrendData           `json:"trends"`
	RiskLevel     ConflictLevelDistribution      `json:"risk_level_distribution"`
	TopTypes      []ConflictTypeDistribution     `json:"top_types"`
}

// ConflictTrendData 冲突趋势数据
type ConflictTrendData struct {
	Date       string `json:"date"`
	Count      int    `json:"count"`
	Severity   string `json:"severity"`
}

// ConflictLevelDistribution 冲突级别分布
type ConflictLevelDistribution struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
}

// ConflictTypeDistribution 冲突类型分布
type ConflictTypeDistribution struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
	Percent float64 `json:"percent"`
}

// enhancedConflictDetectionService 增强冲突检测服务实现
type enhancedConflictDetectionService struct {
	repo repositories.EnhancedConflictRepository
}

// NewEnhancedConflictDetectionService 创建增强冲突检测服务
func NewEnhancedConflictDetectionService(repo repositories.EnhancedConflictRepository) EnhancedConflictDetectionService {
	return &enhancedConflictDetectionService{
		repo: repo,
	}
}

// PerformMultiDimensionalDetection 执行多维度冲突检测
func (s *enhancedConflictDetectionService) PerformMultiDimensionalDetection(ctx context.Context, requestID string, clientProfiles []*models.ClientProfile) ([]*ConflictDetectionResult, error) {
	var results []*ConflictDetectionResult

	// 对每个客户档案进行冲突检测
	for _, profile := range clientProfiles {
		conflicts := s.detectConflictsForProfile(ctx, profile)
		results = append(results, conflicts...)
	}

	// 去重和合并结果
	results = s.deduplicateResults(results)

	return results, nil
}

// DetectConflictsForCase 为案例检测冲突
func (s *enhancedConflictDetectionService) DetectConflictsForCase(ctx context.Context, caseID uint, clientIDs []string) ([]*ConflictDetectionResult, error) {
	var results []*ConflictDetectionResult

	// 临时实现：直接创建虚拟的客户档案进行检测
	// 在实际实现中，需要通过仓库接口获取真实的客户档案
	for _, clientID := range clientIDs {
		profile := &models.ClientProfile{
			BaseModel: models.BaseModel{
				ID: clientID,
			},
			ClientNumber:     fmt.Sprintf("CL-%s", clientID),
			ClientType:       "CORPORATE",
			ClientCategory:   "企业客户",
			ClientStatus:     "ACTIVE",
		}

		conflicts := s.detectConflictsForProfile(ctx, profile)
		results = append(results, conflicts...)
	}

	return results, nil
}

// GetConflictHistory 获取冲突历史
func (s *enhancedConflictDetectionService) GetConflictHistory(ctx context.Context, params *ConflictHistoryParams) ([]*ConflictHistoryRecord, error) {
	// 临时实现，返回空列表
	// 在实际实现中，这里会查询数据库
	return []*ConflictHistoryRecord{}, nil
}

// AnalyzeConflictTrends 分析冲突趋势
func (s *enhancedConflictDetectionService) AnalyzeConflictTrends(ctx context.Context, timeframe string) (*ConflictTrendAnalysis, error) {
	// 临时实现，返回基本分析结果
	analysis := &ConflictTrendAnalysis{
		Timeframe:     timeframe,
		TotalConflicts: 0,
		Trends:        []ConflictTrendData{},
		RiskLevel: ConflictLevelDistribution{
			High:   0,
			Medium: 0,
			Low:    0,
		},
		TopTypes: []ConflictTypeDistribution{},
	}

	return analysis, nil
}

// 内部辅助方法

// detectConflictsForProfile 为单个客户档案检测冲突
func (s *enhancedConflictDetectionService) detectConflictsForProfile(ctx context.Context, profile *models.ClientProfile) []*ConflictDetectionResult {
	var conflicts []*ConflictDetectionResult

	// 检测商业竞争冲突
	businessConflicts := s.detectBusinessCompetitionConflicts(profile)
	conflicts = append(conflicts, businessConflicts...)

	// 检测前客户冲突
	formerClientConflicts := s.detectFormerClientConflicts(profile)
	conflicts = append(conflicts, formerClientConflicts...)

	// 检测利益冲突
	interestConflicts := s.detectInterestConflicts(profile)
	conflicts = append(conflicts, interestConflicts...)

	// 为每个冲突添加详细信息
	for _, conflict := range conflicts {
		conflict.SourceEntity = profile.ClientNumber
		conflict.DetectionDate = fmt.Sprintf("%d", profile.CreatedAt.Unix())
		conflict.Status = "DETECTED"
		conflict.WaiverPossible = s.evaluateWaiverPossibility(conflict)
		conflict.Recommendations = s.generateRecommendations(conflict)
	}

	return conflicts
}

// detectBusinessCompetitionConflicts 检测商业竞争冲突
func (s *enhancedConflictDetectionService) detectBusinessCompetitionConflicts(profile *models.ClientProfile) []*ConflictDetectionResult {
	// 临时实现，返回空列表
	// 在实际实现中，这里会查询客户的竞争对手信息
	return []*ConflictDetectionResult{}
}

// detectFormerClientConflicts 检测前客户冲突
func (s *enhancedConflictDetectionService) detectFormerClientConflicts(profile *models.ClientProfile) []*ConflictDetectionResult {
	// 临时实现，返回空列表
	// 在实际实现中，这里会查询历史案件信息
	return []*ConflictDetectionResult{}
}

// detectInterestConflicts 检测利益冲突
func (s *enhancedConflictDetectionService) detectInterestConflicts(profile *models.ClientProfile) []*ConflictDetectionResult {
	// 临时实现，返回空列表
	// 在实际实现中，这里会分析利益关系网络
	return []*ConflictDetectionResult{}
}

// deduplicateResults 去重结果
func (s *enhancedConflictDetectionService) deduplicateResults(results []*ConflictDetectionResult) []*ConflictDetectionResult {
	seen := make(map[string]bool)
	var deduplicated []*ConflictDetectionResult

	for _, result := range results {
		key := fmt.Sprintf("%s_%s_%s", result.SourceEntity, result.RelatedEntity, result.ConflictType)
		if !seen[key] {
			seen[key] = true
			deduplicated = append(deduplicated, result)
		}
	}

	return deduplicated
}

// evaluateWaiverPossibility 评估豁免可能性
func (s *enhancedConflictDetectionService) evaluateWaiverPossibility(conflict *ConflictDetectionResult) bool {
	// 简化的豁免可能性评估逻辑
	if conflict.SeverityLevel == "HIGH" {
		return false
	}
	if conflict.SeverityLevel == "CRITICAL" {
		return false
	}
	return true
}

// generateRecommendations 生成建议
func (s *enhancedConflictDetectionService) generateRecommendations(conflict *ConflictDetectionResult) []string {
	var recommendations []string

	switch conflict.ConflictType {
	case "BUSINESS_COMPETITION":
		recommendations = append(recommendations, "建议避免代表直接竞争对手")
		recommendations = append(recommendations, "建立信息屏障隔离敏感信息")
	case "FORMER_CLIENT":
		recommendations = append(recommendations, "确认冷却期已过")
		recommendations = append(recommendations, "获得前客户的书面同意")
	case "INTEREST_CONFLICT":
		recommendations = append(recommendations, "披露潜在利益冲突")
		recommendations = append(recommendations, "获得客户知情同意")
	default:
		recommendations = append(recommendations, "详细评估冲突影响")
		recommendations = append(recommendations, "咨询合规部门")
	}

	return recommendations
}