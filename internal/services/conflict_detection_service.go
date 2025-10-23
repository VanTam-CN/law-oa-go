package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ConflictDetectionService 冲突检测服务接口
type ConflictDetectionService interface {
	// PerformConflictCheck 执行冲突检测
	PerformConflictCheck(ctx context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error)
	// GetCheckHistory 获取检测历史
	GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error)
	// GetConflictStats 获取冲突统计
	GetConflictStats(ctx context.Context, clientID string) (*repositories.ConflictStats, error)
}

// conflictDetectionService 冲突检测服务实现
type conflictDetectionService struct {
	conflictRepo repositories.BasicConflictRepository
	riskAssessor RiskAssessor
	userRepo    repositories.UserRepository
	clientRepo  repositories.ClientRepository
	caseRepo    repositories.CaseRepository
}

// NewConflictDetectionService 创建新的冲突检测服务
func NewConflictDetectionService(
	conflictRepo repositories.BasicConflictRepository,
	riskAssessor RiskAssessor,
	userRepo repositories.UserRepository,
	clientRepo repositories.ClientRepository,
	caseRepo repositories.CaseRepository,
) ConflictDetectionService {
	return &conflictDetectionService{
		conflictRepo: conflictRepo,
		riskAssessor: riskAssessor,
		userRepo:     userRepo,
		clientRepo:   clientRepo,
		caseRepo:     caseRepo,
	}
}

// PerformConflictCheck 执行冲突检测
func (s *conflictDetectionService) PerformConflictCheck(ctx context.Context, request *models.ConflictCheckRequest) (*models.ConflictCheckResponse, error) {
	startTime := time.Now()

	log.Printf("🔍 开始执行冲突检测，客户端ID: %s, 案件名称: %s", request.ClientID, request.CaseName)

	// 验证请求
	if err := request.Validate(); err != nil {
		log.Printf("❌ 冲突检测请求验证失败: %v", err)
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 生成检测ID
	checkID := fmt.Sprintf("CC_%d", time.Now().UnixNano())

	// 查找潜在冲突案例
	conflictCases, err := s.findPotentialConflicts(ctx, request)
	if err != nil {
		log.Printf("❌ 查找潜在冲突失败: %v", err)
		return nil, fmt.Errorf("查找潜在冲突失败: %w", err)
	}

	// 风险评估
	riskAssessment, err := s.riskAssessor.AssessRisk(ctx, conflictCases, []*RuleMatch{})
	if err != nil {
		log.Printf("❌ 风险评估失败: %v", err)
		return nil, fmt.Errorf("风险评估失败: %w", err)
	}

	// 生成建议
	recommendations, err := s.riskAssessor.GenerateRecommendations(ctx, riskAssessment.OverallRisk, conflictCases)
	if err != nil {
		log.Printf("❌ 生成建议失败: %v", err)
		return nil, fmt.Errorf("生成建议失败: %w", err)
	}

	// 构建响应
	response := &models.ConflictCheckResponse{
		CheckID:         checkID,
		HasConflict:     len(conflictCases) > 0,
		ConflictCases:   conflictCases,
		CheckStatistics: s.buildCheckStatistics(ctx, request, conflictCases),
		RiskAssessment:  riskAssessment,
		Recommendations: recommendations,
		CheckTime:       time.Now(),
		Duration:        time.Since(startTime).Milliseconds(),
	}

	// 保存检测记录
	if err := s.saveCheckRecord(ctx, request, response); err != nil {
		log.Printf("⚠️ 保存检测记录失败: %v", err)
		// 不影响主流程，只记录日志
	}

	log.Printf("✅ 冲突检测完成，检测到 %d 个冲突案例，风险等级: %s", len(conflictCases), riskAssessment.OverallRisk)

	return response, nil
}

// findPotentialConflicts 查找潜在冲突案例
func (s *conflictDetectionService) findPotentialConflicts(ctx context.Context, request *models.ConflictCheckRequest) ([]*models.ConflictCase, error) {
	var allConflicts []*models.ConflictCase

	// 1. 查找律师已代理的其他案件
	// 转换用户ID为uint用于数据库查询
	userIDUint, err := strconv.ParseUint(request.UserID, 10, 32)
	if err != nil {
		log.Printf("⚠️ 解析用户ID失败: %v", err)
		return nil, fmt.Errorf("用户ID格式错误: %w", err)
	}
	lawyerConflicts, err := s.conflictRepo.GetPotentialConflicts(ctx, request.ClientID, uint(userIDUint), request.OtherParties)
	if err != nil {
		log.Printf("⚠️ 查找律师冲突失败: %v", err)
	} else {
		allConflicts = append(allConflicts, lawyerConflicts...)
		log.Printf("📋 找到 %d 个律师相关冲突案例", len(lawyerConflicts))
	}

	// 2. 检查对方当事人冲突
	if len(request.OtherParties) > 0 {
		opponentConflicts := s.checkOpponentConflicts(ctx, request)
		allConflicts = append(allConflicts, opponentConflicts...)
		log.Printf("📋 找到 %d 个对方当事人冲突案例", len(opponentConflicts))
	}

	// 3. 检查客户关系冲突
	if request.IncludeCorporateRelations {
		relationConflicts := s.checkClientRelationConflicts(ctx, request)
		allConflicts = append(allConflicts, relationConflicts...)
		log.Printf("📋 找到 %d 个客户关系冲突案例", len(relationConflicts))
	}

	// 4. 行业竞争冲突分析
	industryConflicts := s.checkIndustryCompetitionConflicts(ctx, request)
	allConflicts = append(allConflicts, industryConflicts...)
	log.Printf("📋 找到 %d 个行业竞争冲突案例", len(industryConflicts))

	// 去重并优化结果
	return s.deduplicateConflicts(allConflicts), nil
}

// checkOpponentConflicts 检查对方当事人冲突
func (s *conflictDetectionService) checkOpponentConflicts(ctx context.Context, request *models.ConflictCheckRequest) []*models.ConflictCase {
	var conflicts []*models.ConflictCase

	for _, opponent := range request.OtherParties {
		if strings.TrimSpace(opponent) == "" {
			continue
		}

		// 转换用户ID为uint用于数据库查询
		userIDUint, err := strconv.ParseUint(request.UserID, 10, 32)
		if err != nil {
			log.Printf("⚠️ 解析用户ID失败: %v", err)
			continue
		}

		// 查询包含对方当事人名称的案件
		query := fmt.Sprintf(`
			SELECT
				c.id,
				c.title as case_name,
				c.case_type,
				c.description,
				cl.name as client_name,
				u.name as lawyer_name,
				c.created_at,
				c.lawyer_id
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			JOIN users u ON c.lawyer_id = u.id
			WHERE c.deleted_at IS NULL
			AND (c.title ILIKE ? OR c.description ILIKE ? OR cl.name ILIKE ?)
			AND c.lawyer_id != ?
			ORDER BY c.created_at DESC
			LIMIT 10
		`)

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query,
			"%"+opponent+"%", "%"+opponent+"%", "%"+opponent+"%", uint(userIDUint)).Rows()
		if err != nil {
			log.Printf("⚠️ 查询对方当事人冲突失败: %v", err)
			continue
		}

		for rows.Next() {
			var caseID uint
			var caseName, caseType, description, clientName, lawyerName string
			var lawyerID uint
			var createdAt time.Time

			if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &lawyerName, &createdAt, &lawyerID); err != nil {
				continue
			}

			// 确定冲突类型和风险等级
			conflictType := s.determineConflictType(caseName, description, opponent)
			riskLevel := s.assessConflictRisk(conflictType, createdAt)

			conflict := &models.ConflictCase{
				ID:              fmt.Sprintf("opponent_%d", caseID),
				CaseID:          fmt.Sprintf("%d", caseID),
				CaseName:        caseName,
				CaseNo:          fmt.Sprintf("CASE-%d", caseID),
				ConflictType:    conflictType,
				RiskLevel:       riskLevel,
				Description:      fmt.Sprintf("对方当事人 '%s' 与案件 '%s' 存在关联", opponent, caseName),
				CaseStatus:      "active",
				ClientID:        request.ClientID,
				OpposingParties: []string{opponent},
				ConflictDetails: fmt.Sprintf("系统检测到对方当事人 '%s' 与此案件存在潜在冲突", opponent),
				CreatedAt:       createdAt,
			}

			conflicts = append(conflicts, conflict)
		}
		rows.Close()
	}

	return conflicts
}

// checkClientRelationConflicts 检查客户关系冲突
func (s *conflictDetectionService) checkClientRelationConflicts(ctx context.Context, request *models.ConflictCheckRequest) []*models.ConflictCase {
	var conflicts []*models.ConflictCase

	// 获取客户关系
	relations, err := s.conflictRepo.GetClientRelations(ctx, request.ClientID)
	if err != nil {
		log.Printf("⚠️ 获取客户关系失败: %v", err)
		return conflicts
	}

	for _, relation := range relations {
		// 转换用户ID为uint用于数据库查询
		userIDUint, err := strconv.ParseUint(request.UserID, 10, 32)
		if err != nil {
			log.Printf("⚠️ 解析用户ID失败: %v", err)
			continue
		}

		// 查询关联客户的相关案件
		query := fmt.Sprintf(`
			SELECT
				c.id,
				c.title as case_name,
				c.case_type,
				c.description,
				cl.name as client_name,
				u.name as lawyer_name,
				c.created_at
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			JOIN users u ON c.lawyer_id = u.id
			WHERE c.deleted_at IS NULL
			AND c.client_id = ?
			AND c.lawyer_id != ?
			ORDER BY c.created_at DESC
			LIMIT 5
		`)

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, relation.RelatedClientID, uint(userIDUint)).Rows()
		if err != nil {
			continue
		}

		for rows.Next() {
			var caseID uint
			var caseName, caseType, description, clientName, lawyerName string
			var createdAt time.Time

			if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &lawyerName, &createdAt); err != nil {
				continue
			}

			conflict := &models.ConflictCase{
				ID:              fmt.Sprintf("relation_%d", caseID),
				CaseID:          fmt.Sprintf("%d", caseID),
				CaseName:        caseName,
				CaseNo:          fmt.Sprintf("CASE-%d", caseID),
				ConflictType:    "客户关系冲突",
				RiskLevel:       "MEDIUM",
				Description:      fmt.Sprintf("关联客户 '%s' 的案件 '%s' 存在潜在冲突", clientName, caseName),
				CaseStatus:      "active",
				ClientID:        request.ClientID,
				ConflictDetails: fmt.Sprintf("客户关系: %s - %s", relation.RelationType, relation.RelationDetail),
				CreatedAt:       createdAt,
			}

			conflicts = append(conflicts, conflict)
		}
		rows.Close()
	}

	return conflicts
}

// checkIndustryCompetitionConflicts 检查行业竞争冲突
func (s *conflictDetectionService) checkIndustryCompetitionConflicts(ctx context.Context, request *models.ConflictCheckRequest) []*models.ConflictCase {
	var conflicts []*models.ConflictCase

	// 获取当前客户行业信息
	clientIDUint, err := strconv.ParseUint(request.ClientID, 10, 32)
	if err != nil {
		log.Printf("⚠️ 解析客户ID失败: %v", err)
		return conflicts
	}
	client, err := s.clientRepo.FindByID(ctx, uint(clientIDUint))
	if err != nil {
		log.Printf("⚠️ 获取客户信息失败: %v", err)
		return conflicts
	}

	// 基于客户名称和案件类型推断行业
	currentIndustry := s.inferIndustry(client.Name, request.CaseType)
	if currentIndustry == "" {
		return conflicts
	}

	// 转换用户ID为uint用于数据库查询
	userIDUint, err := strconv.ParseUint(request.UserID, 10, 32)
	if err != nil {
		log.Printf("⚠️ 解析用户ID失败: %v", err)
		return conflicts
	}

	// 查找同行业的竞争案件
	query := fmt.Sprintf(`
		SELECT
			c.id,
			c.title as case_name,
			c.case_type,
			c.description,
			cl.name as client_name,
			u.name as lawyer_name,
			c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.deleted_at IS NULL
		AND c.lawyer_id = ?
		AND c.client_id != ?
		AND (cl.name ILIKE ? OR cl.name ILIKE ? OR cl.name ILIKE ?)
		ORDER BY c.created_at DESC
		LIMIT 10
	`)

	industryKeywords := s.getIndustryKeywords(currentIndustry)
	clientIDUint, err2 := strconv.ParseUint(request.ClientID, 10, 32)
	if err2 != nil {
		log.Printf("⚠️ 解析客户ID失败: %v", err2)
		return conflicts
	}

	rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, uint(userIDUint), uint(clientIDUint),
		"%"+industryKeywords[0]+"%", "%"+industryKeywords[1]+"%", "%"+industryKeywords[2]+"%").Rows()
	if err != nil {
		log.Printf("⚠️ 查询行业竞争冲突失败: %v", err)
		return conflicts
	}

	for rows.Next() {
		var caseID uint
		var caseName, caseType, description, clientName, lawyerName string
		var createdAt time.Time

		if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &lawyerName, &createdAt); err != nil {
			continue
		}

		conflict := &models.ConflictCase{
			ID:              fmt.Sprintf("industry_%d", caseID),
			CaseID:          fmt.Sprintf("%d", caseID),
			CaseName:        caseName,
			CaseNo:          fmt.Sprintf("CASE-%d", caseID),
			ConflictType:    "行业竞争冲突",
			RiskLevel:       "HIGH",
			Description:      fmt.Sprintf("同行业竞争: '%s' 与 '%s' 存在业务竞争", client.Name, clientName),
			CaseStatus:      "active",
			ClientID:        request.ClientID,
			ConflictDetails: fmt.Sprintf("行业: %s, 可能存在商业利益冲突", currentIndustry),
			CreatedAt:       createdAt,
		}

		conflicts = append(conflicts, conflict)
	}
	rows.Close()

	return conflicts
}

// 辅助方法

// determineConflictType 确定冲突类型
func (s *conflictDetectionService) determineConflictType(caseName, description, opponent string) string {
	caseName = strings.ToLower(caseName)
	description = strings.ToLower(description)
	opponent = strings.ToLower(opponent)

	if strings.Contains(caseName, "离婚") || strings.Contains(description, "离婚") ||
		strings.Contains(caseName, "抚养") || strings.Contains(description, "抚养") {
		return "法律对立冲突"
	}

	if strings.Contains(caseName, "股权") || strings.Contains(description, "股权") ||
		strings.Contains(caseName, "收购") || strings.Contains(description, "收购") {
		return "股权纠纷冲突"
	}

	if strings.Contains(caseName, "商标") || strings.Contains(description, "商标") ||
		strings.Contains(caseName, "专利") || strings.Contains(description, "专利") ||
		strings.Contains(caseName, "版权") || strings.Contains(description, "版权") {
		return "知识产权冲突"
	}

	if strings.Contains(caseName, "合同") || strings.Contains(description, "合同") ||
		strings.Contains(caseName, "服务") || strings.Contains(description, "服务") {
		return "服务纠纷冲突"
	}

	return "商业竞争冲突"
}

// assessConflictRisk 评估冲突风险
func (s *conflictDetectionService) assessConflictRisk(conflictType string, createdAt time.Time) string {
	// 基于冲突类型的基础风险
	baseRisk := map[string]string{
		"法律对立冲突":    "CRITICAL",
		"股权纠纷冲突":    "HIGH",
		"知识产权冲突":    "HIGH",
		"服务纠纷冲突":    "MEDIUM",
		"商业竞争冲突":    "HIGH",
		"客户关系冲突":    "MEDIUM",
		"行业竞争冲突":    "HIGH",
	}

	riskLevel := baseRisk[conflictType]
	if riskLevel == "" {
		riskLevel = "MEDIUM"
	}

	// 时间衰减：越旧的案件风险越低
	hoursPassed := time.Since(createdAt).Hours()
	if hoursPassed > 2160 { // 3个月
		if riskLevel == "CRITICAL" {
			riskLevel = "HIGH"
		} else if riskLevel == "HIGH" {
			riskLevel = "MEDIUM"
		} else if riskLevel == "MEDIUM" {
			riskLevel = "LOW"
		}
	}

	return riskLevel
}

// deduplicateConflicts 去重冲突案例
func (s *conflictDetectionService) deduplicateConflicts(conflicts []*models.ConflictCase) []*models.ConflictCase {
	seen := make(map[string]bool)
	result := make([]*models.ConflictCase, 0)

	for _, conflict := range conflicts {
		key := fmt.Sprintf("%s_%s", conflict.CaseID, conflict.ConflictType)
		if !seen[key] {
			seen[key] = true
			result = append(result, conflict)
		}
	}

	return result
}

// buildCheckStatistics 构建检查统计
func (s *conflictDetectionService) buildCheckStatistics(ctx context.Context, request *models.ConflictCheckRequest, conflicts []*models.ConflictCase) *models.CheckStatistics {
	endTime := time.Now()
	timeRange := fmt.Sprintf("%d年", request.SearchYears)
	if request.SearchYears == 0 {
		timeRange = "默认范围"
	}

	return &models.CheckStatistics{
		TotalCasesChecked:        int64(len(conflicts)),
		ClientHistoryCases:        int64(len(conflicts)), // 简化处理
		RelatedPartiesChecked:     int64(len(request.OtherParties) + 1),
		CorporateRelationsChecked: 0, // 简化处理
		TimeRange:                 timeRange,
		SearchScope:               request.SearchDepth,
		StartTime:                 endTime,
		EndTime:                   endTime,
	}
}

// saveCheckRecord 保存检测记录
func (s *conflictDetectionService) saveCheckRecord(ctx context.Context, request *models.ConflictCheckRequest, response *models.ConflictCheckResponse) error {
	// 转换用户ID为uint用于数据库存储
	userIDUint, err := strconv.ParseUint(request.UserID, 10, 32)
	if err != nil {
		log.Printf("⚠️ 保存记录时解析用户ID失败: %v", err)
		// 如果转换失败，使用默认值0
		userIDUint = 0
	}

	record := &models.ConflictCheckRecord{
		CheckID:       response.CheckID,
		ClientID:      request.ClientID,
		ClientName:    request.ClientName,
		CaseName:      request.CaseName,
		CaseType:      request.CaseType,
		CheckStatus:   "COMPLETED",
		HasConflict:   response.HasConflict,
		RiskLevel:     response.RiskAssessment.OverallRisk,
		UserID:        uint(userIDUint),
		Duration:      response.Duration,
		CheckTime:     response.CheckTime,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return s.conflictRepo.SaveCheckRecord(ctx, record)
}

// inferIndustry 推断行业
func (s *conflictDetectionService) inferIndustry(clientName, caseType string) string {
	clientName = strings.ToLower(clientName)
	caseType = strings.ToLower(caseType)

	// 互联网科技
	if strings.Contains(clientName, "科技") || strings.Contains(clientName, "网络") ||
		strings.Contains(clientName, "软件") || strings.Contains(clientName, "信息") ||
		strings.Contains(caseType, "互联网") || strings.Contains(caseType, "电商") {
		return "互联网科技"
	}

	// 金融
	if strings.Contains(clientName, "银行") || strings.Contains(clientName, "保险") ||
		strings.Contains(clientName, "证券") || strings.Contains(clientName, "基金") ||
		strings.Contains(caseType, "金融") || strings.Contains(caseType, "投资") {
		return "金融"
	}

	// 房地产
	if strings.Contains(clientName, "地产") || strings.Contains(clientName, "置业") ||
		strings.Contains(clientName, "建设") || strings.Contains(clientName, "房地产") ||
		strings.Contains(caseType, "房地产") || strings.Contains(caseType, "建设") {
		return "房地产"
	}

	// 制造业
	if strings.Contains(clientName, "制造") || strings.Contains(clientName, "生产") ||
		strings.Contains(clientName, "工厂") || strings.Contains(caseType, "制造") {
		return "制造业"
	}

	return ""
}

// getIndustryKeywords 获取行业关键词
func (s *conflictDetectionService) getIndustryKeywords(industry string) []string {
	keywords := map[string][]string{
		"互联网科技": {"科技", "网络", "软件", "信息", "互联网"},
		"金融":     {"银行", "保险", "证券", "基金", "金融", "投资"},
		"房地产":   {"地产", "置业", "建设", "房地产", "物业"},
		"制造业":   {"制造", "生产", "工厂", "加工", "制造"},
	}

	if kw, exists := keywords[industry]; exists {
		return kw
	}

	return []string{industry}
}

// GetCheckHistory 获取检测历史
func (s *conflictDetectionService) GetCheckHistory(ctx context.Context, clientID string, limit int) ([]*models.ConflictCheckRecord, error) {
	return s.conflictRepo.GetCheckHistory(ctx, clientID, limit)
}

// GetConflictStats 获取冲突统计
func (s *conflictDetectionService) GetConflictStats(ctx context.Context, clientID string) (*repositories.ConflictStats, error) {
	return s.conflictRepo.GetConflictStats(ctx, clientID)
}