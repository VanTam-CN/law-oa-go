package services

import (
	"context"
	"fmt"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ConflictService 冲突检测服务接口
type ConflictService interface {
	// CheckConflict 检测利益冲突
	CheckConflict(ctx context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error)
	// GetCheckHistory 获取检查历史
	GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error)
	// GetCheckDetails 获取检查详情
	GetCheckDetails(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error)
	// SaveCheckResult 保存检查结果
	SaveCheckResult(ctx context.Context, result *models.ConflictCheckResponse) error
	// GetConflictRules 获取冲突检测规则
	GetConflictRules(ctx context.Context) ([]*models.ConflictRule, error)
	// UpdateConflictRule 更新冲突检测规则
	UpdateConflictRule(ctx context.Context, rule *models.ConflictRule) error
	// DeleteConflictRule 删除冲突检测规则
	DeleteConflictRule(ctx context.Context, ruleID string) error
	// GetMCPStandards 获取MCP标准
	GetMCPStandards(ctx context.Context) (*models.MCPStandards, error)
}

// ConflictServiceImpl 冲突检测服务实现
type ConflictServiceImpl struct {
	conflictRepo repositories.ConflictRepository
	mcpClient    interface{}
	ruleEngine   interface{}
	riskAssessor interface{}
}

// NewConflictService 创建新的冲突检测服务
func NewConflictService(
	conflictRepo repositories.ConflictRepository,
	mcpClient interface{},
	ruleEngine interface{},
	riskAssessor interface{},
) ConflictService {
	return &ConflictServiceImpl{
		conflictRepo: conflictRepo,
		mcpClient:    mcpClient,
		ruleEngine:   ruleEngine,
		riskAssessor: riskAssessor,
	}
}

// CheckConflict 检测利益冲突
func (s *ConflictServiceImpl) CheckConflict(ctx context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	startTime := time.Now()

	// 验证请求
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 生成检查ID
	checkID := fmt.Sprintf("CC_%s_%d", request.ClientID, time.Now().Unix())

	// 创建检查记录
	checkRecord := &models.ConflictCheckRecord{
		CheckID:         checkID,
		ClientID:        request.ClientID,
		ClientName:      request.ClientName,
		CaseName:        request.CaseName,
		CaseType:        request.CaseType,
		CheckStatus:     "PROCESSING",
		HasConflict:    false,
		RiskLevel:       "LOW", // 设置默认风险等级
		SearchParameters: models.FromMap(map[string]interface{}{
			"clientType":              request.ClientType,
			"otherParties":            request.OtherParties,
			"searchYears":             request.SearchYears,
			"includeCorporateRelations": request.IncludeCorporateRelations,
			"searchDepth":             request.SearchDepth,
		}),
		UserID:    request.UserID,
		CheckTime: startTime,
		CreatedAt: startTime,
		UpdatedAt: startTime,
	}

	// 保存检查记录
	if err := s.conflictRepo.SaveCheckRecord(ctx, checkRecord); err != nil {
		return nil, fmt.Errorf("保存检查记录失败: %w", err)
	}

	// 获取MCP标准
	var mcpStandards *models.MCPStandards
	if mcpClient, ok := s.mcpClient.(MCPClient); ok && mcpClient != nil {
		var err error
		mcpStandards, err = mcpClient.FetchLatestStandards(ctx)
		if err != nil {
			// MCP服务失败不影响主要功能，记录错误并继续
			fmt.Printf("获取MCP标准失败: %v\n", err)
		}
	}

	// 获取冲突检测规则
	rules, err := s.conflictRepo.GetConflictRules(ctx, true)
	if err != nil {
		return nil, fmt.Errorf("获取冲突检测规则失败: %w", err)
	}

	// 搜索冲突案例
	searchParams := &repositories.ConflictSearchParams{
		ClientID:    request.ClientID,
		CaseType:    request.CaseType,
		PageSize:    100,
	}

	conflictCases, err := s.conflictRepo.GetConflictCases(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("搜索冲突案例失败: %w", err)
	}

	// 执行规则评估
	var ruleEvaluationResult *RuleEvaluationResult
	if ruleEngine, ok := s.ruleEngine.(RuleEngine); ok && ruleEngine != nil {
		var err error
		ruleEvaluationResult, err = ruleEngine.EvaluateRules(ctx, request, rules)
		if err != nil {
			return nil, fmt.Errorf("规则评估失败: %w", err)
		}
	} else {
		// 创建空的评估结果
		ruleEvaluationResult = &RuleEvaluationResult{
			EvaluatedAt: time.Now(),
			RuleCount:   len(rules),
		}
	}

	// 执行风险评估
	var riskAssessment *models.RiskAssessment
	if riskAssessor, ok := s.riskAssessor.(RiskAssessor); ok && riskAssessor != nil {
		var err error
		riskAssessment, err = riskAssessor.AssessRisk(ctx, conflictCases, ruleEvaluationResult.Matches)
		if err != nil {
			return nil, fmt.Errorf("风险评估失败: %w", err)
		}
	} else {
		// 创建默认的风险评估
		riskAssessment = &models.RiskAssessment{
			OverallRisk:     "LOW",
			RiskScore:       0.0,
			RiskReason:      "未启用风险评估器",
			RequiresApproval: false,
			RiskFactors:     []string{},
			Mitigation:      []string{"未发现明显冲突，建议继续监控"},
		}
	}

	// 生成处理建议
	recommendations := riskAssessment.Mitigation
	if len(recommendations) == 0 {
		recommendations = []string{"未发现明显冲突，建议继续监控"}
	}

	// 统计信息
	totalCasesChecked := int64(len(conflictCases))
	relatedPartiesChecked := int64(len(request.OtherParties))
	corporateRelationsChecked := int64(0)

	if request.IncludeCorporateRelations {
		relations, err := s.conflictRepo.GetClientRelations(ctx, request.ClientID)
		if err == nil {
			corporateRelationsChecked = int64(len(relations))
		}
	}

	timeRange := fmt.Sprintf("%d年", request.SearchYears)
	if request.SearchYears == 0 {
		timeRange = "默认"
	}

	checkStatistics := &models.CheckStatistics{
		TotalCasesChecked:        totalCasesChecked,
		ClientHistoryCases:        totalCasesChecked,
		RelatedPartiesChecked:     relatedPartiesChecked,
		CorporateRelationsChecked: corporateRelationsChecked,
		TimeRange:                 timeRange,
		SearchScope:               request.SearchDepth,
		StartTime:                 startTime,
		EndTime:                   time.Now(),
	}

	// 构建响应
	response := &models.ConflictCheckResponse{
		CheckID:         checkID,
		HasConflict:     len(conflictCases) > 0 || len(ruleEvaluationResult.Matches) > 0,
		ConflictCases:   conflictCases,
		CheckStatistics: checkStatistics,
		RiskAssessment:  riskAssessment,
		Recommendations: recommendations,
		CheckTime:       time.Now(),
		Duration:        time.Since(startTime).Milliseconds(),
		MCPStandards:    mcpStandards,
	}

	// 更新检查记录
	checkRecord.CheckStatus = "COMPLETED"
	checkRecord.HasConflict = response.HasConflict
	checkRecord.RiskLevel = riskAssessment.OverallRisk
	checkRecord.CheckResult = models.FromMap(map[string]interface{}{
		"hasConflict":    response.HasConflict,
		"conflictCount":  len(conflictCases),
		"riskScore":      riskAssessment.RiskScore,
		"riskLevel":      riskAssessment.OverallRisk,
		"recommendations": recommendations,
	})
	checkRecord.Duration = response.Duration
	checkRecord.UpdatedAt = time.Now()

	// 注意：这里应该更新检查记录，但仓库接口暂不支持
	fmt.Printf("检查记录需要更新，但仓库接口暂不支持更新操作\n")

	return response, nil
}

// GetCheckHistory 获取检查历史
func (s *ConflictServiceImpl) GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error) {
	return s.conflictRepo.GetCheckHistory(ctx, clientID, limit)
}

// GetCheckDetails 获取检查详情
func (s *ConflictServiceImpl) GetCheckDetails(ctx context.Context, checkID string) (*models.ConflictCheckRecord, error) {
	// TODO: 仓库接口暂不支持获取单个检查记录
	// return s.conflictRepo.GetCheckRecord(ctx, checkID)
	return nil, fmt.Errorf("获取检查详情功能尚未实现")
}

// SaveCheckResult 保存检查结果
func (s *ConflictServiceImpl) SaveCheckResult(ctx context.Context, result *models.ConflictCheckResponse) error {
	// TODO: 仓库接口暂不支持更新检查记录
	// 获取现有记录
	// record, err := s.conflictRepo.GetCheckRecord(ctx, result.CheckID)
	// if err != nil {
	// 	return fmt.Errorf("获取检查记录失败: %w", err)
	// }
	//
	// // 更新记录
	// record.CheckResult = map[string]interface{}{
	// 	"hasConflict":    result.HasConflict,
	// 	"conflictCount":  len(result.ConflictCases),
	// 	"riskScore":      result.RiskAssessment.RiskScore,
	// 	"riskLevel":      result.RiskAssessment.OverallRisk,
	// 	"recommendations": result.Recommendations,
	// }
	// record.UpdatedAt = time.Now()
	//
	// return s.conflictRepo.UpdateCheckRecord(ctx, record)
	return fmt.Errorf("保存检查结果功能尚未实现")
}

// GetConflictRules 获取冲突检测规则
func (s *ConflictServiceImpl) GetConflictRules(ctx context.Context) ([]*models.ConflictRule, error) {
	return s.conflictRepo.GetConflictRules(ctx, false)
}

// UpdateConflictRule 更新冲突检测规则
func (s *ConflictServiceImpl) UpdateConflictRule(ctx context.Context, rule *models.ConflictRule) error {
	// 验证规则
	if err := rule.Validate(); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 更新规则
	if err := s.conflictRepo.UpdateConflictRule(ctx, rule); err != nil {
		return fmt.Errorf("更新规则失败: %w", err)
	}

	// 更新规则引擎中的规则
	if ruleEngine, ok := s.ruleEngine.(RuleEngine); ok && ruleEngine != nil {
		if err := ruleEngine.UpdateRule(ctx, rule); err != nil {
			fmt.Printf("更新规则引擎失败: %v\n", err)
		}
	}

	return nil
}

// DeleteConflictRule 删除冲突检测规则
func (s *ConflictServiceImpl) DeleteConflictRule(ctx context.Context, ruleID string) error {
	// TODO: 仓库接口暂不支持删除规则
	// if err := s.conflictRepo.DeleteConflictRule(ctx, ruleID); err != nil {
	// 	return fmt.Errorf("删除规则失败: %w", err)
	// }

	// 从规则引擎中删除规则
	if ruleEngine, ok := s.ruleEngine.(RuleEngine); ok && ruleEngine != nil {
		if err := ruleEngine.RemoveRule(ctx, ruleID); err != nil {
			fmt.Printf("从规则引擎删除规则失败: %v\n", err)
		}
	}

	return nil
}

// GetMCPStandards 获取MCP标准
func (s *ConflictServiceImpl) GetMCPStandards(ctx context.Context) (*models.MCPStandards, error) {
	if s.mcpClient == nil {
		return nil, models.ErrMCPServiceUnavailable
	}

	if mcpClient, ok := s.mcpClient.(MCPClient); ok {
		return mcpClient.FetchLatestStandards(ctx)
	}
	return nil, models.ErrMCPServiceUnavailable
}