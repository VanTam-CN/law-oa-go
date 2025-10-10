package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// EnhancedConflictEngine 增强的利益冲突检测引擎
type EnhancedConflictEngine struct {
	conflictRepo repositories.ConflictRepository
	caseRepo     repositories.CaseRepository
	clientRepo   repositories.ClientRepository
}

// NewEnhancedConflictEngine 创建新的增强冲突检测引擎
func NewEnhancedConflictEngine(
	conflictRepo repositories.ConflictRepository,
	caseRepo repositories.CaseRepository,
	clientRepo repositories.ClientRepository,
) *EnhancedConflictEngine {
	return &EnhancedConflictEngine{
		conflictRepo: conflictRepo,
		caseRepo:     caseRepo,
		clientRepo:   clientRepo,
	}
}

// ConflictMatch 冲突匹配结果
type ConflictMatch struct {
	CaseID          string            `json:"caseId"`
	CaseName        string            `json:"caseName"`
	ConflictType    string            `json:"conflictType"`
	RiskLevel       string            `json:"riskLevel"`
	MatchScore      float64           `json:"matchScore"`
	MatchReasons    []string          `json:"matchReasons"`
	ConflictDetails string            `json:"conflictDetails"`
	CaseStatus      string            `json:"caseStatus"`
	ClientID        string            `json:"clientId"`
	OpposingParties []string          `json:"opposingParties"`
	MatchedEntities []MatchedEntity   `json:"matchedEntities"`
}

// MatchedEntity 匹配的实体
type MatchedEntity struct {
	EntityID     string  `json:"entityId"`
	EntityType   string  `json:"entityType"` // PERSON, COMPANY, etc.
	EntityName   string  `json:"entityName"`
	MatchType    string  `json:"matchType"`    // EXACT, FUZZY, PHONETIC, etc.
	MatchScore   float64 `json:"matchScore"`
	OriginalName string  `json:"originalName"`
	MatchedName  string  `json:"matchedName"`
}

// ConflictAnalysis 冲突分析结果
type ConflictAnalysis struct {
	TotalCasesChecked     int64            `json:"totalCasesChecked"`
	ConflictMatches       []*ConflictMatch  `json:"conflictMatches"`
	HighRiskMatches       []*ConflictMatch  `json:"highRiskMatches"`
	MediumRiskMatches     []*ConflictMatch  `json:"mediumRiskMatches"`
	LowRiskMatches        []*ConflictMatch  `json:"lowRiskMatches"`
	ClientHistoryCases    int64            `json:"clientHistoryCases"`
	RelatedPartiesChecked int64            `json:"relatedPartiesChecked"`
	CorporateRelations    int64            `json:"corporateRelationsChecked"`
	SearchTimeRange       string           `json:"searchTimeRange"`
	AnalysisDuration      int64            `json:"analysisDuration"`
	RiskAssessment        *RiskAssessment  `json:"riskAssessment"`
	Recommendations       []string         `json:"recommendations"`
}

// RiskAssessment 风险评估
type RiskAssessment struct {
	OverallRisk      string   `json:"overallRisk"`
	RiskScore        float64  `json:"riskScore"`
	RiskFactors      []string `json:"riskFactors"`
	RequiresApproval bool     `json:"requiresApproval"`
	ApprovalLevel    string   `json:"approvalLevel"`
	Mitigation       []string `json:"mitigation"`
}

// CheckConflict 检测利益冲突
func (e *EnhancedConflictEngine) CheckConflict(
	ctx context.Context,
	request *models.ConflictCheckRequest,
) (*ConflictAnalysis, error) {
	startTime := time.Now()

	// 1. 数据预处理和标准化
	normalizedRequest := e.normalizeRequest(request)

	// 2. 获取历史案例数据
	historicalCases, err := e.getHistoricalCases(ctx, normalizedRequest)
	if err != nil {
		return nil, fmt.Errorf("获取历史案例失败: %w", err)
	}

	// 3. 获取相关方数据
	relatedParties, err := e.getRelatedParties(ctx, normalizedRequest)
	if err != nil {
		return nil, fmt.Errorf("获取相关方数据失败: %w", err)
	}

	// 4. 执行多层次匹配分析
	matches := e.performMultiLayerMatching(ctx, normalizedRequest, historicalCases, relatedParties)

	// 5. 风险评估
	riskAssessment := e.performRiskAssessment(matches)

	// 6. 生成建议
	recommendations := e.generateRecommendations(riskAssessment, matches)

	analysis := &ConflictAnalysis{
		TotalCasesChecked:     int64(len(historicalCases)),
		ConflictMatches:       matches,
		HighRiskMatches:       e.filterMatchesByRisk(matches, "HIGH"),
		MediumRiskMatches:     e.filterMatchesByRisk(matches, "MEDIUM"),
		LowRiskMatches:        e.filterMatchesByRisk(matches, "LOW"),
		ClientHistoryCases:    e.countClientHistoryCases(historicalCases, normalizedRequest.ClientID),
		RelatedPartiesChecked: int64(len(relatedParties)),
		CorporateRelations:    e.countCorporateRelations(ctx, normalizedRequest.ClientID),
		SearchTimeRange:       e.formatSearchTimeRange(request.SearchYears),
		AnalysisDuration:      time.Since(startTime).Milliseconds(),
		RiskAssessment:        riskAssessment,
		Recommendations:       recommendations,
	}

	return analysis, nil
}

// normalizeRequest 标准化请求数据
func (e *EnhancedConflictEngine) normalizeRequest(request *models.ConflictCheckRequest) *models.ConflictCheckRequest {
	normalized := *request

	// 标准化客户名称
	normalized.ClientName = e.normalizeEntityName(normalized.ClientName)

	// 标准化其他相关方名称
	for i, party := range normalized.OtherParties {
		normalized.OtherParties[i] = e.normalizeEntityName(party)
	}

	// 设置默认值
	if normalized.SearchYears == 0 {
		normalized.SearchYears = 5 // 默认5年
	}
	if normalized.SearchDepth == "" {
		normalized.SearchDepth = "STANDARD"
	}

	return &normalized
}

// normalizeEntityName 标准化实体名称
func (e *EnhancedConflictEngine) normalizeEntityName(name string) string {
	if name == "" {
		return ""
	}

	// 转换为小写
	name = strings.ToLower(name)

	// 移除多余的空格
	name = strings.Join(strings.Fields(name), " ")

	// 移除常见的公司后缀缀
	suffixes := []string{
		"co", "ltd", "inc", "corp", "llc", "limited", "company", "corporation",
		"公司", "有限", "股份", "集团", "控股", "投资", "实业", "科技",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(name, " "+suffix) || strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, " "+suffix)
			name = strings.TrimSuffix(name, suffix)
		}
	}

	// 移除标点符号
	var result strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}

// getHistoricalCases 获取历史案例
func (e *EnhancedConflictEngine) getHistoricalCases(
	ctx context.Context,
	request *models.ConflictCheckRequest,
) ([]*models.Case, error) {
	// 这里需要根据实际仓库接口实现
	// 暂时返回空切片，后续实现
	return []*models.Case{}, nil
}

// getRelatedParties 获取相关方数据
func (e *EnhancedConflictEngine) getRelatedParties(
	ctx context.Context,
	request *models.ConflictCheckRequest,
) ([]string, error) {
	parties := make([]string, 0)

	// 添加客户相关方
	if request.IncludeCorporateRelations {
		// 获取企业关联方
		corporateRelations, err := e.conflictRepo.GetClientRelations(ctx, request.ClientID)
		if err == nil {
			for _, relation := range corporateRelations {
				parties = append(parties, relation.RelatedClientID)
			}
		}
	}

	// 添加其他相关方
	parties = append(parties, request.OtherParties...)

	return parties, nil
}

// performMultiLayerMatching 执行多层次匹配
func (e *EnhancedConflictEngine) performMultiLayerMatching(
	ctx context.Context,
	request *models.ConflictCheckRequest,
	historicalCases []*models.Case,
	relatedParties []string,
) []*ConflictMatch {
	var matches []*ConflictMatch

	// 第1层：精确匹配
	exactMatches := e.performExactMatching(request, historicalCases)
	matches = append(matches, exactMatches...)

	// 第2层：模糊匹配
	fuzzyMatches := e.performFuzzyMatching(request, historicalCases)
	matches = append(matches, fuzzyMatches...)

	// 第3层：语音匹配（处理音译名称）
	phoneticMatches := e.performPhoneticMatching(request, historicalCases)
	matches = append(matches, phoneticMatches...)

	// 第4层：关联实体匹配
	entityMatches := e.performEntityMatching(request, relatedParties, historicalCases)
	matches = append(matches, entityMatches...)

	// 去重和评分
	matches = e.deduplicateAndScore(matches)

	// 按风险等级排序
	matches = e.sortMatchesByRisk(matches)

	return matches
}

// performExactMatching 执行精确匹配
func (e *EnhancedConflictEngine) performExactMatching(
	request *models.ConflictCheckRequest,
	cases []*models.Case,
) []*ConflictMatch {
	var matches []*ConflictMatch

	for _, historicalCase := range cases {
		// 检查客户名称精确匹配
		clientName := ""
		if historicalCase.Client != nil {
			clientName = historicalCase.Client.Name
		}
		if e.isExactMatch(request.ClientName, clientName) {
			match := &ConflictMatch{
				CaseID:       fmt.Sprintf("%d", historicalCase.ID),
				CaseName:     historicalCase.Title,
				ConflictType: "CLIENT_NAME_EXACT",
				RiskLevel:    "HIGH",
				MatchScore:   1.0,
				MatchReasons: []string{"客户名称完全匹配"},
				CaseStatus:   historicalCase.Status,
				ClientID:     fmt.Sprintf("%d", historicalCase.ClientID),
				MatchedEntities: []MatchedEntity{
					{
						EntityID:     request.ClientID,
						EntityType:   request.ClientType,
						EntityName:   request.ClientName,
						MatchType:    "EXACT",
						MatchScore:   1.0,
						OriginalName: request.ClientName,
						MatchedName:  clientName,
					},
				},
			}
			matches = append(matches, match)
		}

		// 检查对方当事人精确匹配（从描述中提取）
		for _, party := range request.OtherParties {
			if e.isExactMatch(party, historicalCase.Description) {
				match := &ConflictMatch{
					CaseID:       fmt.Sprintf("%d", historicalCase.ID),
					CaseName:     historicalCase.Title,
					ConflictType: "OPPOSING_PARTY_EXACT",
					RiskLevel:    "CRITICAL",
					MatchScore:   1.0,
					MatchReasons: []string{"对方当事人完全匹配"},
					CaseStatus:   historicalCase.Status,
					ClientID:     fmt.Sprintf("%d", historicalCase.ClientID),
				}
				matches = append(matches, match)
			}
		}
	}

	return matches
}

// performFuzzyMatching 执行模糊匹配
func (e *EnhancedConflictEngine) performFuzzyMatching(
	request *models.ConflictCheckRequest,
	cases []*models.Case,
) []*ConflictMatch {
	var matches []*ConflictMatch

	for _, historicalCase := range cases {
		// 获取客户名称
		clientName := ""
		if historicalCase.Client != nil {
			clientName = historicalCase.Client.Name
		}

		// 使用Levenshtein距离进行模糊匹配
		clientScore := e.calculateLevenshteinSimilarity(request.ClientName, clientName)
		if clientScore >= 0.8 { // 80%相似度阈值
			match := &ConflictMatch{
				CaseID:       fmt.Sprintf("%d", historicalCase.ID),
				CaseName:     historicalCase.Title,
				ConflictType: "CLIENT_NAME_FUZZY",
				RiskLevel:    "MEDIUM",
				MatchScore:   clientScore,
				MatchReasons: []string{fmt.Sprintf("客户名称相似度: %.1f%%", clientScore*100)},
				CaseStatus:   historicalCase.Status,
				ClientID:     fmt.Sprintf("%d", historicalCase.ClientID),
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// performPhoneticMatching 执行语音匹配
func (e *EnhancedConflictEngine) performPhoneticMatching(
	request *models.ConflictCheckRequest,
	cases []*models.Case,
) []*ConflictMatch {
	var matches []*ConflictMatch

	for _, historicalCase := range cases {
		// 获取客户名称
		clientName := ""
		if historicalCase.Client != nil {
			clientName = historicalCase.Client.Name
		}

		// 使用Soundex算法进行语音匹配
		if e.isPhoneticMatch(request.ClientName, clientName) {
			match := &ConflictMatch{
				CaseID:       fmt.Sprintf("%d", historicalCase.ID),
				CaseName:     historicalCase.Title,
				ConflictType: "CLIENT_NAME_PHONETIC",
				RiskLevel:    "LOW",
				MatchScore:   0.7,
				MatchReasons: []string{"客户名称语音相似"},
				CaseStatus:   historicalCase.Status,
				ClientID:     fmt.Sprintf("%d", historicalCase.ClientID),
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// performEntityMatching 执行实体匹配
func (e *EnhancedConflictEngine) performEntityMatching(
	request *models.ConflictCheckRequest,
	relatedParties []string,
	cases []*models.Case,
) []*ConflictMatch {
	var matches []*ConflictMatch

	for _, party := range relatedParties {
		for _, historicalCase := range cases {
			// 获取客户名称
			clientName := ""
			if historicalCase.Client != nil {
				clientName = historicalCase.Client.Name
			}

			// 检查相关方与历史案例的匹配
			if e.isExactMatch(party, clientName) ||
				e.isExactMatch(party, historicalCase.Description) {

				match := &ConflictMatch{
					CaseID:       fmt.Sprintf("%d", historicalCase.ID),
					CaseName:     historicalCase.Title,
					ConflictType: "RELATED_PARTY_MATCH",
					RiskLevel:    "MEDIUM",
					MatchScore:   0.8,
					MatchReasons: []string{fmt.Sprintf("相关方 '%s' 匹配", party)},
					CaseStatus:   historicalCase.Status,
					ClientID:     fmt.Sprintf("%d", historicalCase.ClientID),
				}
				matches = append(matches, match)
			}
		}
	}

	return matches
}

// deduplicateAndScore 去重和评分
func (e *EnhancedConflictEngine) deduplicateAndScore(matches []*ConflictMatch) []*ConflictMatch {
	// 使用map去重，以CaseID+ConflictType为键
	matchMap := make(map[string]*ConflictMatch)

	for _, match := range matches {
		key := fmt.Sprintf("%s_%s", match.CaseID, match.ConflictType)
		if existing, exists := matchMap[key]; exists {
			// 保留分数更高的匹配
			if match.MatchScore > existing.MatchScore {
				matchMap[key] = match
			}
		} else {
			matchMap[key] = match
		}
	}

	// 转换回切片
	var result []*ConflictMatch
	for _, match := range matchMap {
		result = append(result, match)
	}

	return result
}

// sortMatchesByRisk 按风险等级排序
func (e *EnhancedConflictEngine) sortMatchesByRisk(matches []*ConflictMatch) []*ConflictMatch {
	riskOrder := map[string]int{
		"CRITICAL": 0,
		"HIGH":     1,
		"MEDIUM":   2,
		"LOW":      3,
	}

	sorted := make([]*ConflictMatch, len(matches))
	copy(sorted, matches)

	// 简单的冒泡排序按风险等级排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if riskOrder[sorted[j].RiskLevel] > riskOrder[sorted[j+1].RiskLevel] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// 辅助函数

func (e *EnhancedConflictEngine) isExactMatch(name1, name2 string) bool {
	return e.normalizeEntityName(name1) == e.normalizeEntityName(name2)
}

func (e *EnhancedConflictEngine) calculateLevenshteinSimilarity(s1, s2 string) float64 {
	// 简化的Levenshtein距离计算
	s1 = e.normalizeEntityName(s1)
	s2 = e.normalizeEntityName(s2)

	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 {
		return 0.0
	}
	if len2 == 0 {
		return 0.0
	}

	// 这里使用简化的编辑距离计算
	// 实际实现应该使用完整的Levenshtein算法
	distance := e.simpleEditDistance(s1, s2)
	maxLen := len1
	if len2 > len1 {
		maxLen = len2
	}

	return 1.0 - float64(distance)/float64(maxLen)
}

func (e *EnhancedConflictEngine) simpleEditDistance(s1, s2 string) int {
	// 简化的编辑距离计算
	if len(s1) < len(s2) {
		s1, s2 = s2, s1
	}

	if len(s2) == 0 {
		return len(s1)
	}

	distance := 0
	for i := 0; i < len(s2) && i < len(s1); i++ {
		if s1[i] != s2[i] {
			distance++
		}
	}

	distance += len(s1) - len(s2)
	return distance
}

func (e *EnhancedConflictEngine) isPhoneticMatch(name1, name2 string) bool {
	// 简化的语音匹配，实际应该使用Soundex或Metaphone算法
	soundex1 := e.simpleSoundex(name1)
	soundex2 := e.simpleSoundex(name2)
	return soundex1 == soundex2
}

func (e *EnhancedConflictEngine) simpleSoundex(name string) string {
	if name == "" {
		return "0000"
	}

	name = e.normalizeEntityName(name)
	if len(name) == 0 {
		return "0000"
	}

	// 简化的Soundex算法
	result := string(unicode.ToUpper(rune(name[0])))
	code := ""

	for i := 1; i < len(name) && len(code) < 3; i++ {
		char := name[i]
		switch char {
		case 'b', 'f', 'p', 'v':
			if len(code) == 0 || code[len(code)-1] != '1' {
				code += "1"
			}
		case 'c', 'g', 'j', 'k', 'q', 's', 'x', 'z':
			if len(code) == 0 || code[len(code)-1] != '2' {
				code += "2"
			}
		case 'd', 't':
			if len(code) == 0 || code[len(code)-1] != '3' {
				code += "3"
			}
		case 'l':
			if len(code) == 0 || code[len(code)-1] != '4' {
				code += "4"
			}
		case 'm', 'n':
			if len(code) == 0 || code[len(code)-1] != '5' {
				code += "5"
			}
		case 'r':
			if len(code) == 0 || code[len(code)-1] != '6' {
				code += "6"
			}
		}
	}

	// 填充到4位
	for len(code) < 3 {
		code += "0"
	}

	return result + code
}

func (e *EnhancedConflictEngine) filterMatchesByRisk(matches []*ConflictMatch, riskLevel string) []*ConflictMatch {
	var filtered []*ConflictMatch
	for _, match := range matches {
		if match.RiskLevel == riskLevel {
			filtered = append(filtered, match)
		}
	}
	return filtered
}

func (e *EnhancedConflictEngine) countClientHistoryCases(cases []*models.Case, clientID string) int64 {
	count := int64(0)
	for _, case_ := range cases {
		if fmt.Sprintf("%d", case_.ClientID) == clientID {
			count++
		}
	}
	return count
}

func (e *EnhancedConflictEngine) countCorporateRelations(ctx context.Context, clientID string) int64 {
	relations, err := e.conflictRepo.GetClientRelations(ctx, clientID)
	if err != nil {
		return 0
	}
	return int64(len(relations))
}

func (e *EnhancedConflictEngine) formatSearchTimeRange(years int) string {
	if years == 0 {
		return "全部历史"
	}
	endYear := time.Now().Year()
	startYear := endYear - years
	return fmt.Sprintf("%d年 - %d年", startYear, endYear)
}

func (e *EnhancedConflictEngine) performRiskAssessment(matches []*ConflictMatch) *RiskAssessment {
	if len(matches) == 0 {
		return &RiskAssessment{
			OverallRisk:      "LOW",
			RiskScore:        0.0,
			RiskFactors:      []string{},
			RequiresApproval: false,
			Mitigation:       []string{"未发现明显冲突，建议正常处理"},
		}
	}

	riskScore := 0.0
	riskFactors := []string{}

	for _, match := range matches {
		switch match.RiskLevel {
		case "CRITICAL":
			riskScore += 0.4
			riskFactors = append(riskFactors, fmt.Sprintf("关键冲突: %s", match.CaseName))
		case "HIGH":
			riskScore += 0.3
			riskFactors = append(riskFactors, fmt.Sprintf("高风险冲突: %s", match.CaseName))
		case "MEDIUM":
			riskScore += 0.2
			riskFactors = append(riskFactors, fmt.Sprintf("中等风险冲突: %s", match.CaseName))
		case "LOW":
			riskScore += 0.1
		}
	}

	// 限制分数在0-1之间
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	overallRisk := "LOW"
	requiresApproval := false
	approvalLevel := ""

	if riskScore >= 0.7 {
		overallRisk = "CRITICAL"
		requiresApproval = true
		approvalLevel = "SENIOR_PARTNER"
	} else if riskScore >= 0.5 {
		overallRisk = "HIGH"
		requiresApproval = true
		approvalLevel = "PARTNER"
	} else if riskScore >= 0.3 {
		overallRisk = "MEDIUM"
		requiresApproval = false
	}

	return &RiskAssessment{
		OverallRisk:      overallRisk,
		RiskScore:        riskScore,
		RiskFactors:      riskFactors,
		RequiresApproval: requiresApproval,
		ApprovalLevel:    approvalLevel,
		Mitigation:       e.generateMitigation(overallRisk, matches),
	}
}

func (e *EnhancedConflictEngine) generateMitigation(riskLevel string, matches []*ConflictMatch) []string {
	var mitigation []string

	switch riskLevel {
	case "CRITICAL":
		mitigation = []string{
			"立即停止案件受理",
			"要求高级合伙人审查",
			"考虑是否需要拒绝代理",
			"详细记录冲突情况",
		}
	case "HIGH":
		mitigation = []string{
			"要求合伙人级别审查",
			"获取客户书面同意",
			"建立信息隔离墙",
			"持续监控潜在冲突",
		}
	case "MEDIUM":
		mitigation = []string{
			"要求主管律师审查",
			"加强内部信息管理",
			"定期更新冲突检查",
		}
	case "LOW":
		mitigation = []string{
			"正常案件处理流程",
			"保持基本的冲突监控",
		}
	default:
		mitigation = []string{"未发现明显冲突，建议正常处理"}
	}

	return mitigation
}

func (e *EnhancedConflictEngine) generateRecommendations(
	assessment *RiskAssessment,
	matches []*ConflictMatch,
) []string {
	var recommendations []string

	if len(matches) == 0 {
		recommendations = []string{
			"未发现明显利益冲突",
			"可以正常受理案件",
			"建议在案件处理过程中持续监控",
		}
	} else {
		recommendations = append(recommendations, assessment.Mitigation...)

		if assessment.RequiresApproval {
			recommendations = append(recommendations,
				fmt.Sprintf("需要%s级别批准", assessment.ApprovalLevel))
		}

		if assessment.RiskScore >= 0.5 {
			recommendations = append(recommendations,
				"建议咨询风险管理委员会")
		}
	}

	return recommendations
}