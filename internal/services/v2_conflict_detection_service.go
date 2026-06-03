package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"gorm.io/gorm"
)

// V2ConflictDetectionService v2 冲突检测服务接口
type V2ConflictDetectionService interface {
	// QuickCheck 快速冲突检测（基于冲突池）
	QuickCheck(ctx context.Context, req *ConflictCheckRequestV2) (*ConflictCheckResultV2, error)
	// DetailedCheck 详细冲突检测（包含 API 调用）
	DetailedCheck(ctx context.Context, req *ConflictCheckRequestV2) (*ConflictCheckResultV2, error)
	// BatchCheck 批量冲突检测
	BatchCheck(ctx context.Context, reqs []*ConflictCheckRequestV2) ([]*ConflictCheckResultV2, error)
	// GetRiskLevel 获取风险等级
	GetRiskLevel(ctx context.Context, req *ConflictCheckRequestV2) (string, error)
}

// ConflictCheckRequestV2 v2 冲突检测请求
type ConflictCheckRequestV2 struct {
	LawyerID       uint     `json:"lawyerId" validate:"required"`
	ClientName     string   `json:"clientName" validate:"required"`
	ClientTaxID    string   `json:"clientTaxId"`
	CaseID         uint     `json:"caseId"`
	OpposingParties []string `json:"opposingParties"`
	SearchDepth    string   `json:"searchDepth"` // basic/standard/deep
	IncludeRelated bool     `json:"includeRelated"`
}

// ConflictCheckResultV2 v2 冲突检测结果
type ConflictCheckResultV2 struct {
	CheckID        string                `json:"checkId"`
	RiskLevel      string                `json:"riskLevel"`
	RiskScore      float64               `json:"riskScore"`
	MatchCount     int                   `json:"matchCount"`
	Matches        []*ConflictMatchV2    `json:"matches"`
	CheckTime      time.Time             `json:"checkTime"`
	DurationMs     int64                 `json:"durationMs"`
	SearchScope    string                `json:"searchScope"`
	Recommendations []string             `json:"recommendations"`
}

// ConflictMatchV2 v2 冲突匹配项
type ConflictMatchV2 struct {
	MatchID        string                 `json:"matchId"`
	MatchType      string                 `json:"matchType"`     // direct/indirect/related
	LawyerID       uint                   `json:"lawyerId"`
	LawyerName     string                 `json:"lawyerName"`
	CaseID         uint                   `json:"caseId"`
	CaseTitle      string                 `json:"caseTitle"`
	CaseType       string                 `json:"caseType"`
	Relationship   string                 `json:"relationship"`  // client/opposing/witness
	MatchReason    string                 `json:"matchReason"`
	EntityInfo     *V2ConflictEntityInfo `json:"entityInfo"`
	RiskLevel      string                 `json:"riskLevel"`
	RiskFactors    []string               `json:"riskFactors"`
}

// EntityInfo 实体信息
type V2ConflictEntityInfo struct {
	Name           string   `json:"name"`
	StandardName   string   `json:"standardName"`
	TaxID          string   `json:"taxId"`
	Type           string   `json:"type"`
	Aliases        []string `json:"aliases"`
}

// v2ConflictDetectionService v2 冲突检测服务实现
type v2ConflictDetectionService struct {
	db              *gorm.DB
	poolService     ConflictPoolService
	caseRepo        CaseRepository
	clientRepo      ClientRepository
	userRepo        UserRepository
	companyAPI      CompanyAPIService
	config          *ConflictDetectionConfigV2
}

// UserRepository 用户仓库接口
type UserRepository interface {
	FindByID(ctx context.Context, id uint) (*models.User, error)
}

// ConflictDetectionConfigV2 v2 冲突检测配置
type ConflictDetectionConfigV2 struct {
	CacheEnabled        bool          `json:"cacheEnabled"`
	CacheTTL            time.Duration `json:"cacheTTL"`
	MaxConcurrency      int           `json:"maxConcurrency"`
	DefaultSearchDepth  string        `json:"defaultSearchDepth"`
	RiskThresholds      RiskThresholds `json:"riskThresholds"`
}

// RiskThresholds 风险阈值
type RiskThresholds struct {
	CriticalScore float64 `json:"criticalScore"` // >= 严重风险
	HighScore     float64 `json:"highScore"`     // >= 高风险
	MediumScore   float64 `json:"mediumScore"`   // >= 中风险
	LowScore      float64 `json:"lowScore"`      // >= 低风险
}

// NewV2ConflictDetectionService 创建新的 v2 冲突检测服务
func NewV2ConflictDetectionService(
	db *gorm.DB,
	poolService ConflictPoolService,
	caseRepo CaseRepository,
	clientRepo ClientRepository,
	userRepo UserRepository,
	companyAPI CompanyAPIService,
) V2ConflictDetectionService {
	return &v2ConflictDetectionService{
		db:          db,
		poolService: poolService,
		caseRepo:    caseRepo,
		clientRepo:  clientRepo,
		userRepo:    userRepo,
		companyAPI:  companyAPI,
		config:      DefaultConflictDetectionConfigV2(),
	}
}

// DefaultConflictDetectionConfigV2 默认配置
func DefaultConflictDetectionConfigV2() *ConflictDetectionConfigV2 {
	return &ConflictDetectionConfigV2{
		CacheEnabled:       true,
		CacheTTL:           10 * time.Minute,
		MaxConcurrency:     10,
		DefaultSearchDepth: "standard",
		RiskThresholds: RiskThresholds{
			CriticalScore: 0.8,
			HighScore:     0.6,
			MediumScore:   0.4,
			LowScore:      0.2,
		},
	}
}

// QuickCheck 快速冲突检测（基于冲突池）
func (s *v2ConflictDetectionService) QuickCheck(ctx context.Context, req *ConflictCheckRequestV2) (*ConflictCheckResultV2, error) {
	startTime := time.Now()

	log.Printf("🔍 快速冲突检测: lawyerID=%d, clientName=%s", req.LawyerID, req.ClientName)

	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 生成检测ID
	checkID := fmt.Sprintf("QC_%d_%d", req.LawyerID, time.Now().UnixNano())

	// 标准化客户名称
	standardName := s.standardizeName(req.ClientName)

	// 在冲突池中搜索
	matches, err := s.searchInPool(ctx, req, standardName)
	if err != nil {
		return nil, fmt.Errorf("搜索冲突池失败: %w", err)
	}

	// 如果需要，检查对方当事人
	if req.IncludeRelated && len(req.OpposingParties) > 0 {
		opposingMatches := s.searchOpposingParties(ctx, req)
		matches = append(matches, opposingMatches...)
	}

	// 计算风险等级和分数
	riskLevel, riskScore := s.calculateRisk(matches)

	// 生成建议
	recommendations := s.generateRecommendations(riskLevel, matches)

	result := &ConflictCheckResultV2{
		CheckID:        checkID,
		RiskLevel:      riskLevel,
		RiskScore:      riskScore,
		MatchCount:     len(matches),
		Matches:        matches,
		CheckTime:      startTime,
		DurationMs:     time.Since(startTime).Milliseconds(),
		SearchScope:    s.getSearchScope(req),
		Recommendations: recommendations,
	}

	log.Printf("✅ 快速冲突检测完成: 风险等级=%s, 匹配数=%d, 耗时=%dms",
		riskLevel, len(matches), result.DurationMs)

	return result, nil
}

// DetailedCheck 详细冲突检测（包含 API 调用）
func (s *v2ConflictDetectionService) DetailedCheck(ctx context.Context, req *ConflictCheckRequestV2) (*ConflictCheckResultV2, error) {
	startTime := time.Now()

	log.Printf("🔍 详细冲突检测: lawyerID=%d, clientName=%s", req.LawyerID, req.ClientName)

	// 先执行快速检测
	result, err := s.QuickCheck(ctx, req)
	if err != nil {
		return nil, err
	}

	// 如果需要深度搜索且有税号，调用 API 获取更多信息
	if req.SearchDepth == "deep" && req.ClientTaxID != "" {
		apiMatches := s.searchWithAPI(ctx, req)
		result.Matches = append(result.Matches, apiMatches...)
		result.MatchCount = len(result.Matches)

		// 重新计算风险
		riskLevel, riskScore := s.calculateRisk(result.Matches)
		result.RiskLevel = riskLevel
		result.RiskScore = riskScore
		result.Recommendations = s.generateRecommendations(riskLevel, result.Matches)
	}

	result.DurationMs = time.Since(startTime).Milliseconds()

	log.Printf("✅ 详细冲突检测完成: 风险等级=%s, 匹配数=%d, 耗时=%dms",
		result.RiskLevel, result.MatchCount, result.DurationMs)

	return result, nil
}

// BatchCheck 批量冲突检测
func (s *v2ConflictDetectionService) BatchCheck(ctx context.Context, reqs []*ConflictCheckRequestV2) ([]*ConflictCheckResultV2, error) {
	log.Printf("🔍 批量冲突检测: 数量=%d", len(reqs))

	results := make([]*ConflictCheckResultV2, len(reqs))

	for i, req := range reqs {
		result, err := s.QuickCheck(ctx, req)
		if err != nil {
			log.Printf("⚠️ 批量检测中第 %d 项失败: %v", i, err)
			result = &ConflictCheckResultV2{
				CheckID: fmt.Sprintf("ERR_%d", i),
				RiskLevel: "ERROR",
				CheckTime: time.Now(),
			}
		}
		results[i] = result
	}

	return results, nil
}

// GetRiskLevel 获取风险等级
func (s *v2ConflictDetectionService) GetRiskLevel(ctx context.Context, req *ConflictCheckRequestV2) (string, error) {
	result, err := s.QuickCheck(ctx, req)
	if err != nil {
		return "UNKNOWN", err
	}
	return result.RiskLevel, nil
}

// ============================================================================
// 私有方法
// ============================================================================

// validateRequest 验证请求
func (s *v2ConflictDetectionService) validateRequest(req *ConflictCheckRequestV2) error {
	if req.LawyerID == 0 {
		return fmt.Errorf("律师ID不能为空")
	}
	if strings.TrimSpace(req.ClientName) == "" {
		return fmt.Errorf("客户名称不能为空")
	}
	return nil
}

// standardizeName 标准化名称
func (s *v2ConflictDetectionService) standardizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	// 移除常见后缀
	suffixes := []string{"有限公司", "股份有限公司", "有限责任公司", "集团"}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

// searchInPool 在冲突池中搜索
func (s *v2ConflictDetectionService) searchInPool(ctx context.Context, req *ConflictCheckRequestV2, standardName string) ([]*ConflictMatchV2, error) {
	// 构建搜索请求
	poolReq := &PoolSearchRequest{
		SearchTerm: standardName,
		SearchType: "fuzzy",
		LawyerID:   &req.LawyerID,
	}

	// 调用冲突池服务搜索
	poolResults, err := s.poolService.SearchInPool(ctx, poolReq)
	if err != nil {
		return nil, err
	}

	// 转换为 ConflictMatchV2
	matches := make([]*ConflictMatchV2, 0, len(poolResults))
	for _, pr := range poolResults {
		// 获取律师信息
		lawyer, _ := s.userRepo.FindByID(ctx, pr.PoolEntry.LawyerID)
		lawyerName := ""
		if lawyer != nil {
			lawyerName = lawyer.Name
		}

		match := &ConflictMatchV2{
			MatchID:     fmt.Sprintf("POOL_%d", pr.PoolEntry.ID),
			MatchType:   s.determineMatchType(pr),
			LawyerID:    pr.PoolEntry.LawyerID,
			LawyerName:  lawyerName,
			CaseID:      pr.PoolEntry.CaseID,
			CaseTitle:   pr.PoolEntry.CaseTitle,
			Relationship: pr.PoolEntry.RelationshipType,
			MatchReason: pr.MatchReason,
			EntityInfo: &V2ConflictEntityInfo{
				Name:         pr.PoolEntry.EntityName,
				StandardName: pr.PoolEntry.EntityNameStandard,
				TaxID:        pr.PoolEntry.EntityTaxID,
				Type:         pr.PoolEntry.EntityType,
			},
			RiskLevel:   pr.RiskLevel,
			RiskFactors: s.extractRiskFactors(pr.PoolEntry),
		}

		matches = append(matches, match)
	}

	return matches, nil
}

// searchOpposingParties 搜索对方当事人
func (s *v2ConflictDetectionService) searchOpposingParties(ctx context.Context, req *ConflictCheckRequestV2) []*ConflictMatchV2 {
	matches := make([]*ConflictMatchV2, 0)

	for _, opposingParty := range req.OpposingParties {
		if strings.TrimSpace(opposingParty) == "" {
			continue
		}

		// 在冲突池中搜索对方当事人
		poolReq := &PoolSearchRequest{
			SearchTerm: s.standardizeName(opposingParty),
			SearchType: "fuzzy",
			LawyerID:   &req.LawyerID,
		}

		poolResults, err := s.poolService.SearchInPool(ctx, poolReq)
		if err != nil {
			log.Printf("⚠️ 搜索对方当事人失败: %s, error: %v", opposingParty, err)
			continue
		}

		for _, pr := range poolResults {
			// 对方当事人冲突风险等级提升
			riskLevel := pr.RiskLevel
			if riskLevel != "CRITICAL" {
				riskLevel = "HIGH"
			}

			lawyer, _ := s.userRepo.FindByID(ctx, pr.PoolEntry.LawyerID)
			lawyerName := ""
			if lawyer != nil {
				lawyerName = lawyer.Name
			}

			match := &ConflictMatchV2{
				MatchID:     fmt.Sprintf("OPP_%d", pr.PoolEntry.ID),
				MatchType:   "opposing",
				LawyerID:    pr.PoolEntry.LawyerID,
				LawyerName:  lawyerName,
				CaseID:      pr.PoolEntry.CaseID,
				CaseTitle:   pr.PoolEntry.CaseTitle,
				Relationship: "opposing",
				MatchReason: fmt.Sprintf("对方当事人 '%s' 与案件关联", opposingParty),
				EntityInfo: &V2ConflictEntityInfo{
					Name:         pr.PoolEntry.EntityName,
					StandardName: pr.PoolEntry.EntityNameStandard,
					TaxID:        pr.PoolEntry.EntityTaxID,
					Type:         pr.PoolEntry.EntityType,
				},
				RiskLevel:   riskLevel,
				RiskFactors: []string{"对方当事人冲突", "法律对立风险"},
			}

			matches = append(matches, match)
		}
	}

	return matches
}

// searchWithAPI 使用 API 搜索
func (s *v2ConflictDetectionService) searchWithAPI(ctx context.Context, req *ConflictCheckRequestV2) []*ConflictMatchV2 {
	matches := make([]*ConflictMatchV2, 0)

	// 调用 API 搜索公司
	searchResult, err := s.companyAPI.SearchCompany(ctx, req.ClientName, ProviderMock)
	if err != nil {
		log.Printf("⚠️ API 搜索失败: %v", err)
		return matches
	}

	// 查找匹配的冲突池条目
	for _, company := range searchResult.Companies {
		poolReq := &PoolSearchRequest{
			SearchTerm: company.TaxID,
			SearchType: "standard",
		}

		poolResults, err := s.poolService.SearchInPool(ctx, poolReq)
		if err != nil {
			continue
		}

		for _, pr := range poolResults {
			lawyer, _ := s.userRepo.FindByID(ctx, pr.PoolEntry.LawyerID)
			lawyerName := ""
			if lawyer != nil {
				lawyerName = lawyer.Name
			}

			match := &ConflictMatchV2{
				MatchID:     fmt.Sprintf("API_%d", pr.PoolEntry.ID),
				MatchType:   "api",
				LawyerID:    pr.PoolEntry.LawyerID,
				LawyerName:  lawyerName,
				CaseID:      pr.PoolEntry.CaseID,
				CaseTitle:   pr.PoolEntry.CaseTitle,
				Relationship: pr.PoolEntry.RelationshipType,
				MatchReason: fmt.Sprintf("通过 API 匹配到税号 %s", company.TaxID),
				EntityInfo: &V2ConflictEntityInfo{
					Name:         company.Name,
					StandardName: s.standardizeName(company.Name),
					TaxID:        company.TaxID,
					Type:         "company",
				},
				RiskLevel:   pr.RiskLevel,
				RiskFactors: []string{"API 匹配", "税号完全匹配"},
			}

			matches = append(matches, match)
		}
	}

	return matches
}

// calculateRisk 计算风险等级和分数
func (s *v2ConflictDetectionService) calculateRisk(matches []*ConflictMatchV2) (string, float64) {
	if len(matches) == 0 {
		return "PASS", 0.0
	}

	// 计算加权风险分数
	totalScore := 0.0
	weights := map[string]float64{
		"CRITICAL": 1.0,
		"HIGH":     0.7,
		"MEDIUM":   0.4,
		"LOW":      0.2,
	}

	for _, match := range matches {
		if weight, ok := weights[match.RiskLevel]; ok {
			totalScore += weight
		}
	}

	// 归一化到 0-1
	normalizedScore := totalScore / float64(len(matches))

	// 确定风险等级
	riskLevel := s.scoreToLevel(normalizedScore)

	return riskLevel, normalizedScore
}

// scoreToLevel 分数转等级
func (s *v2ConflictDetectionService) scoreToLevel(score float64) string {
	switch {
	case score >= s.config.RiskThresholds.CriticalScore:
		return "CRITICAL"
	case score >= s.config.RiskThresholds.HighScore:
		return "HIGH"
	case score >= s.config.RiskThresholds.MediumScore:
		return "MEDIUM"
	case score >= s.config.RiskThresholds.LowScore:
		return "LOW"
	default:
		return "PASS"
	}
}

// generateRecommendations 生成建议
func (s *v2ConflictDetectionService) generateRecommendations(riskLevel string, matches []*ConflictMatchV2) []string {
	recommendations := make([]string, 0)

	switch riskLevel {
	case "CRITICAL":
		recommendations = append(recommendations,
			"检测到严重利益冲突，强烈建议拒绝代理此案",
			"冲突涉及对方当事人或关联实体，存在法律职业道德风险",
			"建议与律所风险控制部门讨论",
		)
	case "HIGH":
		recommendations = append(recommendations,
			"检测到高风险利益冲突，建议谨慎评估",
			"可能存在客户关系冲突，建议详细审查",
			"建议与客户充分披露潜在冲突",
		)
	case "MEDIUM":
		recommendations = append(recommendations,
			"检测到中等风险冲突，建议进一步调查",
			"可能存在间接关联关系，建议获取更多相关信息",
		)
	case "LOW":
		recommendations = append(recommendations,
			"检测到低风险冲突，一般可以接受",
			"建议保持关注，如有变化及时更新",
		)
	case "PASS":
		recommendations = append(recommendations,
			"未检测到明显利益冲突",
			"可以正常处理此案件",
		)
	}

	// 根据匹配项添加具体建议
	for _, match := range matches {
		if match.MatchType == "opposing" {
			recommendations = append(recommendations,
				fmt.Sprintf("与案件 %s 存在对方当事人关联", match.CaseTitle))
		}
	}

	return recommendations
}

// determineMatchType 确定匹配类型
func (s *v2ConflictDetectionService) determineMatchType(pr *PoolMatchResult) string {
	if pr.MatchScore >= 0.9 {
		return "direct"
	} else if pr.MatchScore >= 0.5 {
		return "indirect"
	} else {
		return "related"
	}
}

// extractRiskFactors 提取风险因素
func (s *v2ConflictDetectionService) extractRiskFactors(entry *models.LawyerConflictPool) []string {
	factors := make([]string, 0)

	switch entry.RelationshipType {
	case "client":
		factors = append(factors, "客户关系冲突")
	case "opposing":
		factors = append(factors, "对方当事人冲突")
	case "witness":
		factors = append(factors, "证人冲突")
	}

	if entry.EntityType == "company" {
		factors = append(factors, "企业客户（可能有股权穿透）")
	}

	// 检查股权穿透信息
	if entry.ShareholdingInfo != nil {
		if data, ok := entry.ShareholdingInfo["directShareholders"].([]interface{}); ok {
			if len(data) > 0 {
				factors = append(factors, "存在多层级股权结构")
			}
		}
	}

	return factors
}

// getSearchScope 获取搜索范围
func (s *v2ConflictDetectionService) getSearchScope(req *ConflictCheckRequestV2) string {
	scope := req.SearchDepth
	if scope == "" {
		scope = s.config.DefaultSearchDepth
	}

	switch scope {
	case "basic":
		return "基础搜索（仅直接匹配）"
	case "standard":
		return "标准搜索（直接+模糊匹配）"
	case "deep":
		return "深度搜索（标准+关联实体+API查询）"
	default:
		return "标准搜索"
	}
}

// ============================================================================
// 性能优化辅助方法
// ============================================================================

// GetPerformanceMetrics 获取性能指标
func (s *v2ConflictDetectionService) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{}

	// 统计冲突池大小
	var poolSize int64
	if err := s.db.WithContext(ctx).
		Model(&models.LawyerConflictPool{}).
		Count(&poolSize).Error; err != nil {
		return nil, err
	}
	metrics.PoolSize = poolSize

	// 统计平均查询时间（这里使用模拟值）
	metrics.AvgQueryTimeMs = 50 // 目标 < 100ms
	metrics.P95QueryTimeMs = 90

	// 缓存命中率（模拟值）
	metrics.CacheHitRate = 0.75

	return metrics, nil
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	PoolSize        int64   `json:"poolSize"`
	AvgQueryTimeMs  int64   `json:"avgQueryTimeMs"`
	P95QueryTimeMs  int64   `json:"p95QueryTimeMs"`
	CacheHitRate    float64 `json:"cacheHitRate"`
}
