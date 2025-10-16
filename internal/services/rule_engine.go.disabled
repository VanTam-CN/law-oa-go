package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// RuleEngine 规则引擎接口
type RuleEngine interface {
	// EvaluateRules 执行规则评估
	EvaluateRules(ctx context.Context, request *models.ConflictCheckRequest, rules []*models.ConflictRule) (*RuleEvaluationResult, error)
	// AddRule 添加规则
	AddRule(ctx context.Context, rule *models.ConflictRule) error
	// RemoveRule 删除规则
	RemoveRule(ctx context.Context, ruleID string) error
	// UpdateRule 更新规则
	UpdateRule(ctx context.Context, rule *models.ConflictRule) error
	// GetRule 获取规则
	GetRule(ctx context.Context, ruleID string) (*models.ConflictRule, error)
	// GetAllRules 获取所有规则
	GetAllRules(ctx context.Context) ([]*models.ConflictRule, error)
	// EvaluateSingleRule 评估单个规则
	EvaluateSingleRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error)
}

// RuleEvaluationResult 规则评估结果
type RuleEvaluationResult struct {
	Matches      []*RuleMatch `json:"matches"`
	RiskLevel    string       `json:"riskLevel"`
	TotalScore   float64      `json:"totalScore"`
	EvaluatedAt  time.Time    `json:"evaluatedAt"`
	RuleCount    int          `json:"ruleCount"`
	MatchedCount int          `json:"matchedCount"`
}


// ruleEngine 规则引擎实现
type ruleEngine struct {
	rules      map[string]*models.ConflictRule
	rulesMutex sync.RWMutex
	repository repositories.ConflictRepository
	mcpClient  interface{}
}

// NewRuleEngine 创建新的规则引擎
func NewRuleEngine(repository repositories.ConflictRepository, mcpClient interface{}) RuleEngine {
	return &ruleEngine{
		rules:      make(map[string]*models.ConflictRule),
		repository: repository,
		mcpClient:  mcpClient,
	}
}

// EvaluateRules 执行规则评估
func (e *ruleEngine) EvaluateRules(ctx context.Context, request *models.ConflictCheckRequest, rules []*models.ConflictRule) (*RuleEvaluationResult, error) {
	result := &RuleEvaluationResult{
		EvaluatedAt: time.Now(),
		RuleCount:   len(rules),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并发评估规则
	for _, rule := range rules {
		wg.Add(1)
		go func(r *models.ConflictRule) {
			defer wg.Done()

			match, err := e.EvaluateSingleRule(ctx, request, r)
			if err != nil {
				// 记录错误但继续执行其他规则
				fmt.Printf("规则评估失败: %v\n", err)
				return
			}

			if match.Matched {
				mu.Lock()
				result.Matches = append(result.Matches, match)
				result.TotalScore += match.RiskScore
				result.MatchedCount++
				mu.Unlock()
			}
		}(rule)
	}

	wg.Wait()

	// 计算风险等级
	result.RiskLevel = e.calculateRiskLevel(result.TotalScore, result.MatchedCount)

	return result, nil
}

// EvaluateSingleRule 评估单个规则
func (e *ruleEngine) EvaluateSingleRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error) {
	match := &RuleMatch{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		RuleType:    rule.Type,
		EvaluatedAt: time.Now(),
	}

	// 检查规则是否启用
	if !rule.Active {
		match.Matched = false
		return match, nil
	}

	// 根据规则类型执行不同的评估逻辑
	switch rule.Type {
	case "NAME_SIMILARITY":
		return e.evaluateNameSimilarityRule(ctx, request, rule)
	case "CORPORATE_RELATION":
		return e.evaluateCorporateRelationRule(ctx, request, rule)
	case "CASE_CONFLICT":
		return e.evaluateCaseConflictRule(ctx, request, rule)
	case "ADVERSE_PARTY":
		return e.evaluateAdversePartyRule(ctx, request, rule)
	case "TIME_OVERLAP":
		return e.evaluateTimeOverlapRule(ctx, request, rule)
	case "CUSTOM_PATTERN":
		return e.evaluateCustomPatternRule(ctx, request, rule)
	default:
		return nil, fmt.Errorf("不支持的规则类型: %s", rule.Type)
	}
}

// evaluateNameSimilarityRule 评估姓名相似性规则
func (e *ruleEngine) evaluateNameSimilarityRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error) {
	match := &RuleMatch{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		RuleType:    rule.Type,
		EvaluatedAt: time.Now(),
	}

	// 获取规则配置
	threshold := e.getRuleConfigFloat(rule.Conditions, "threshold", 0.8)

	// 构建搜索条件
	searchTerms := []string{request.ClientName}
	searchTerms = append(searchTerms, request.OtherParties...)

	// 搜索相似名称
	conflictCases, err := e.repository.GetConflictCases(ctx, &repositories.ConflictSearchParams{
		ClientID: request.ClientID,
		PageSize: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索冲突案例失败: %w", err)
	}

	// 检查相似性
	for _, conflictCase := range conflictCases {
		for _, term := range searchTerms {
			similarity := e.calculateStringSimilarity(term, conflictCase.CaseName)
			if similarity >= threshold {
				match.Matched = true
				match.RiskScore = float64(rule.Priority) * similarity
				match.MatchDetails = &RuleMatchDetails{
					MatchType:      "NAME_SIMILARITY",
					MatchedFields:  []string{"caseName"},
					MatchedValues:  []string{term, conflictCase.CaseName},
					Confidence:     similarity,
					Description:    fmt.Sprintf("客户名称 '%s' 与历史案例 '%s' 相似度: %.2f", term, conflictCase.CaseName, similarity),
					Recommendation: "建议进行详细的背景调查以确认是否存在利益冲突",
				}
				return match, nil
			}
		}
	}

	match.Matched = false
	return match, nil
}

// evaluateCorporateRelationRule 评估企业关联规则
func (e *ruleEngine) evaluateCorporateRelationRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error) {
	match := &RuleMatch{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		RuleType:    rule.Type,
		EvaluatedAt: time.Now(),
	}

	// 只对企业客户执行此规则
	if request.ClientType != "COMPANY" {
		match.Matched = false
		return match, nil
	}

	// 获取客户关联关系
	relations, err := e.repository.GetClientRelations(ctx, request.ClientID)
	if err != nil {
		return nil, fmt.Errorf("获取客户关联关系失败: %w", err)
	}

	// 检查是否存在高风险关联
	for _, relation := range relations {
		if relation.RelationType == "COMPETITOR" || relation.RelationType == "ADVERSE" {
			match.Matched = true
			match.RiskScore = float64(rule.Priority)
			match.MatchDetails = &RuleMatchDetails{
				MatchType:      "CORPORATE_RELATION",
				MatchedFields:  []string{"corporateRelations"},
				MatchedValues:  []string{relation.RelatedClientID, relation.RelationType},
				Confidence:     0.9,
				Description:    fmt.Sprintf("客户与 '%s' 存在 '%s' 关系", relation.RelatedClientID, relation.RelationType),
				Recommendation: "建议审查与相关实体的业务关系，评估潜在冲突",
			}
			return match, nil
		}
	}

	match.Matched = false
	return match, nil
}

// evaluateCaseConflictRule 评估案件冲突规则
func (e *ruleEngine) evaluateCaseConflictRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error) {
	match := &RuleMatch{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		RuleType:    rule.Type,
		EvaluatedAt: time.Now(),
	}

	// 搜索同一客户的案件
	conflictCases, err := e.repository.GetConflictCases(ctx, &repositories.ConflictSearchParams{
		ClientID: request.ClientID,
		CaseType: request.CaseType,
		PageSize: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索冲突案件失败: %w", err)
	}

	// 检查是否存在案件冲突
	for _, conflictCase := range conflictCases {
		if e.isCaseConflict(conflictCase, request.CaseName, request.OtherParties) {
			match.Matched = true
			match.RiskScore = float64(rule.Priority)
			match.MatchDetails = &RuleMatchDetails{
				MatchType:      "CASE_CONFLICT",
				MatchedFields:  []string{"caseName", "otherParties"},
				MatchedValues:  []string{conflictCase.CaseName, request.CaseName},
				Confidence:     0.95,
				Description:    fmt.Sprintf("案件 '%s' 与现有案件 '%s' 存在潜在冲突", request.CaseName, conflictCase.CaseName),
				Recommendation: "建议详细审查两个案件的具体情况，确定是否能够同时代理",
			}
			return match, nil
		}
	}

	match.Matched = false
	return match, nil
}

// evaluateAdversePartyRule 评估对立当事人规则
func (e *ruleEngine) evaluateAdversePartyRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error) {
	match := &RuleMatch{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		RuleType:    rule.Type,
		EvaluatedAt: time.Now(),
	}

	// 搜索包含对立当事人的案件
	conflictCases, err := e.repository.GetConflictCases(ctx, &repositories.ConflictSearchParams{
		ClientID: request.ClientID,
		PageSize: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索对立当事人案件失败: %w", err)
	}

	// 检查对立当事人
	for _, conflictCase := range conflictCases {
		for _, party := range request.OtherParties {
			if e.isAdverseParty(conflictCase, party) {
				match.Matched = true
				match.RiskScore = float64(rule.Priority) * 1.2 // 对立当事人风险更高
				match.MatchDetails = &RuleMatchDetails{
					MatchType:      "ADVERSE_PARTY",
					MatchedFields:  []string{"otherParties"},
					MatchedValues:  []string{party, conflictCase.CaseName},
					Confidence:     0.9,
					Description:    fmt.Sprintf("当事人 '%s' 与案件 '%s' 存在对立关系", party, conflictCase.CaseName),
					Recommendation: "严格审查是否存在利益冲突，可能需要回避",
				}
				return match, nil
			}
		}
	}

	match.Matched = false
	return match, nil
}

// evaluateTimeOverlapRule 评估时间重叠规则
func (e *ruleEngine) evaluateTimeOverlapRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error) {
	match := &RuleMatch{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		RuleType:    rule.Type,
		EvaluatedAt: time.Now(),
	}

	// 获取时间范围内的案件
	years := request.SearchYears
	if years == 0 {
		years = 5 // 默认搜索5年
	}

	conflictCases, err := e.repository.GetConflictCases(ctx, &repositories.ConflictSearchParams{
		ClientID: request.ClientID,
		PageSize: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索时间重叠案件失败: %w", err)
	}

	// 检查时间重叠
	for _, conflictCase := range conflictCases {
		if e.isTimeOverlap(conflictCase, request.RequestTime, years) {
			match.Matched = true
			match.RiskScore = float64(rule.Priority) * 0.8 // 时间重叠风险相对较低
			match.MatchDetails = &RuleMatchDetails{
				MatchType:      "TIME_OVERLAP",
				MatchedFields:  []string{"requestTime"},
				MatchedValues:  []string{conflictCase.CaseName},
				Confidence:     0.7,
				Description:    fmt.Sprintf("与案件 '%s' 存在时间重叠", conflictCase.CaseName),
				Recommendation: "确认两个案件的时间安排是否存在冲突",
			}
			return match, nil
		}
	}

	match.Matched = false
	return match, nil
}

// evaluateCustomPatternRule 评估自定义模式规则
func (e *ruleEngine) evaluateCustomPatternRule(ctx context.Context, request *models.ConflictCheckRequest, rule *models.ConflictRule) (*RuleMatch, error) {
	match := &RuleMatch{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		RuleType:    rule.Type,
		EvaluatedAt: time.Now(),
	}

	// 获取自定义模式配置
	pattern := e.getRuleConfigString(rule.Conditions, "pattern", "")
	if pattern == "" {
		return nil, fmt.Errorf("自定义模式规则缺少模式配置")
	}

	// 编译正则表达式
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("无效的正则表达式模式: %w", err)
	}

	// 检查匹配
	testString := fmt.Sprintf("%s|%s|%s", request.ClientName, request.CaseName, strings.Join(request.OtherParties, "|"))
	if regex.MatchString(testString) {
		match.Matched = true
		match.RiskScore = float64(rule.Priority)
		match.MatchDetails = &RuleMatchDetails{
			MatchType:      "CUSTOM_PATTERN",
			MatchedFields:  []string{"customPattern"},
			MatchedValues:  []string{pattern},
			Confidence:     0.85,
			Description:    "匹配自定义模式规则",
			Recommendation: "根据自定义模式规则的建议处理",
		}
	} else {
		match.Matched = false
	}

	return match, nil
}

// AddRule 添加规则
func (e *ruleEngine) AddRule(ctx context.Context, rule *models.ConflictRule) error {
	e.rulesMutex.Lock()
	defer e.rulesMutex.Unlock()

	// 验证规则
	if err := e.validateRule(rule); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 保存到内存
	e.rules[rule.ID] = rule

	// 保存到数据库
	if err := e.repository.SaveConflictRule(ctx, rule); err != nil {
		return fmt.Errorf("保存规则失败: %w", err)
	}

	return nil
}

// RemoveRule 删除规则
func (e *ruleEngine) RemoveRule(ctx context.Context, ruleID string) error {
	e.rulesMutex.Lock()
	defer e.rulesMutex.Unlock()

	// 从内存删除
	delete(e.rules, ruleID)

	// 从数据库删除 (这里需要repository实现删除方法)
	// 暂时只从内存删除

	return nil
}

// UpdateRule 更新规则
func (e *ruleEngine) UpdateRule(ctx context.Context, rule *models.ConflictRule) error {
	e.rulesMutex.Lock()
	defer e.rulesMutex.Unlock()

	// 验证规则
	if err := e.validateRule(rule); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 更新内存
	e.rules[rule.ID] = rule

	// 更新数据库
	if err := e.repository.UpdateConflictRule(ctx, rule); err != nil {
		return fmt.Errorf("更新规则失败: %w", err)
	}

	return nil
}

// GetRule 获取规则
func (e *ruleEngine) GetRule(ctx context.Context, ruleID string) (*models.ConflictRule, error) {
	e.rulesMutex.RLock()
	defer e.rulesMutex.RUnlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("规则不存在: %s", ruleID)
	}

	return rule, nil
}

// GetAllRules 获取所有规则
func (e *ruleEngine) GetAllRules(ctx context.Context) ([]*models.ConflictRule, error) {
	e.rulesMutex.RLock()
	defer e.rulesMutex.RUnlock()

	rules := make([]*models.ConflictRule, 0, len(e.rules))
	for _, rule := range e.rules {
		rules = append(rules, rule)
	}

	return rules, nil
}

// 辅助方法

// calculateRiskLevel 计算风险等级
func (e *ruleEngine) calculateRiskLevel(totalScore float64, matchedCount int) string {
	if matchedCount == 0 {
		return "LOW"
	}

	avgScore := totalScore / float64(matchedCount)
	if avgScore >= 0.8 {
		return "HIGH"
	} else if avgScore >= 0.5 {
		return "MEDIUM"
	} else {
		return "LOW"
	}
}

// calculateStringSimilarity 计算字符串相似度
func (e *ruleEngine) calculateStringSimilarity(s1, s2 string) float64 {
	// 简单的相似度计算，可以替换为更复杂的算法
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 1.0
	}

	// 计算编辑距离
	distance := e.levenshteinDistance(s1, s2)
	maxLen := max(len(s1), len(s2))
	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance 计算编辑距离
func (e *ruleEngine) levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// 创建距离矩阵
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	// 初始化第一行和第一列
	for i := 0; i <= n; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j
	}

	// 填充矩阵
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			dp[i][j] = min3(
				dp[i-1][j]+1,      // 删除
				dp[i][j-1]+1,      // 插入
				dp[i-1][j-1]+cost, // 替换
			)
		}
	}

	return dp[n][m]
}

// isCaseConflict 检查案件冲突
func (e *ruleEngine) isCaseConflict(conflictCase *models.ConflictCase, caseName string, otherParties []string) bool {
	// 简单的案件冲突检查逻辑
	if conflictCase.CaseName == caseName {
		return true
	}

	// 检查当事人是否重复
	for _, party := range otherParties {
		for _, existingParty := range conflictCase.OpposingParties {
			if party == existingParty {
				return true
			}
		}
	}

	return false
}

// isAdverseParty 检查是否为对立当事人
func (e *ruleEngine) isAdverseParty(conflictCase *models.ConflictCase, party string) bool {
	// 简单的对立当事人检查
	for _, adverseParty := range conflictCase.OpposingParties {
		if adverseParty == party {
			return true
		}
	}
	return false
}

// isTimeOverlap 检查时间重叠
func (e *ruleEngine) isTimeOverlap(conflictCase *models.ConflictCase, requestTime time.Time, years int) bool {
	// 简单的时间重叠检查
	timeThreshold := requestTime.AddDate(-years, 0, 0)
	return conflictCase.CreatedAt.After(timeThreshold)
}

// validateRule 验证规则
func (e *ruleEngine) validateRule(rule *models.ConflictRule) error {
	if rule.ID == "" {
		return fmt.Errorf("规则ID不能为空")
	}
	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	if rule.Type == "" {
		return fmt.Errorf("规则类型不能为空")
	}
	if rule.Priority <= 0 {
		return fmt.Errorf("规则优先级必须大于0")
	}

	return nil
}

// 规则配置辅助方法
func (e *ruleEngine) getRuleConfigFloat(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, exists := config[key]; exists {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}

func (e *ruleEngine) getRuleConfigString(config map[string]interface{}, key string, defaultValue string) string {
	if val, exists := config[key]; exists {
		if stringVal, ok := val.(string); ok {
			return stringVal
		}
	}
	return defaultValue
}

func (e *ruleEngine) getRuleConfigStringSlice(config map[string]interface{}, key string, defaultValue []string) []string {
	if val, exists := config[key]; exists {
		if sliceVal, ok := val.([]string); ok {
			return sliceVal
		}
	}
	return defaultValue
}

// minInt 返回两个整数的最小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// min3 返回三个整数的最小值
func min3(a, b, c int) int {
	return minInt(minInt(a, b), c)
}

// max 返回两个整数的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}