package services

import (
	"context"
	"encoding/json"
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
	userRepo     repositories.UserRepository
	clientRepo   repositories.ClientRepository
	caseRepo     repositories.CaseRepository
}

// 🔧 案件类型映射：前端英文 -> 数据库中文
// 🎯 基于数据库实际情况更新映射，确保冲突检测正常工作
var caseTypeMapping = map[string]string{
	"civil":          "民事",
	"commercial":     "商事",
	"criminal":       "刑事",
	"administrative": "行政",
	"labor":          "劳动",
	"intellectual":   "知识产权",
	"financial":      "金融",
	"arbitration":    "仲裁",
	"consultation":   "咨询",
	"other":          "其他",
}

// 🔧 反向映射：数据库中文 -> 前端英文
var reverseCaseTypeMapping = map[string]string{
	"民事":   "civil",
	"商事":   "commercial",
	"刑事":   "criminal",
	"行政":   "administrative",
	"劳动":   "labor",
	"知识产权": "intellectual",
	"金融":   "financial",
}

// 🔧 映射案件类型：将前端发送的英文转换为数据库查询用的中文
func (s *conflictDetectionService) mapCaseType(frontendCaseType string) string {
	if mapped, exists := caseTypeMapping[frontendCaseType]; exists {
		return mapped
	}
	// 如果找不到映射，返回原值（可能是中文）
	return frontendCaseType
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

	if request.SearchYears == 0 {
		request.SearchYears = 5
	}
	if request.SearchDepth == "" {
		request.SearchDepth = "STANDARD"
	}

	// 验证请求
	if err := request.Validate(); err != nil {
		log.Printf("❌ 冲突检测请求验证失败: %v", err)
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 生成检测ID
	checkID := request.CheckID
	if checkID == "" {
		checkID = fmt.Sprintf("CC_%d", time.Now().UnixNano())
	}

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
		log.Printf("❌ 保存检测记录失败: %v", err)
		return nil, fmt.Errorf("保存冲突检测审计记录失败: %w", err)
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

	log.Printf("🔍 开始冲突检测: 客户ID=%s, 用户ID=%s, 案件类型=%s, 律师ID=%d",
		request.ClientID, request.UserID, request.CaseType, uint(userIDUint))

	since := s.searchStartTime(request)

	lawyerConflicts, err := s.conflictRepo.GetPotentialConflicts(ctx, request.ClientID, uint(userIDUint), request.OtherParties, since)
	if err != nil {
		log.Printf("⚠️ 查找律师冲突失败: %v", err)
	} else {
		allConflicts = append(allConflicts, lawyerConflicts...)
		log.Printf("📋 找到 %d 个律师相关冲突案例", len(lawyerConflicts))
	}

	// 2. 检查对方当事人冲突
	if len(request.OtherParties) > 0 {
		opponentConflicts := s.checkOpponentConflicts(ctx, request, since)
		allConflicts = append(allConflicts, opponentConflicts...)
		log.Printf("📋 找到 %d 个对方当事人冲突案例", len(opponentConflicts))
	}

	// 3. 检查客户关系冲突
	if request.IncludeCorporateRelations && request.SearchDepth != "BASIC" {
		relationConflicts := s.checkClientRelationConflicts(ctx, request, since)
		allConflicts = append(allConflicts, relationConflicts...)
		log.Printf("📋 找到 %d 个客户关系冲突案例", len(relationConflicts))
	}

	// 4. 行业竞争冲突分析
	if request.SearchDepth == "DEEP" {
		industryConflicts := s.checkIndustryCompetitionConflicts(ctx, request, since)
		allConflicts = append(allConflicts, industryConflicts...)
		log.Printf("📋 找到 %d 个行业竞争冲突案例", len(industryConflicts))
	}

	// 去重并优化结果
	return s.deduplicateConflicts(allConflicts), nil
}

// checkOpponentConflicts 检查对方当事人冲突
func (s *conflictDetectionService) checkOpponentConflicts(ctx context.Context, request *models.ConflictCheckRequest, since time.Time) []*models.ConflictCase {
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

		sinceFilter := ""
		args := []interface{}{
			"%" + opponent + "%",
			"%" + opponent + "%",
			"%" + opponent + "%",
			uint(userIDUint),
		}
		if !since.IsZero() {
			sinceFilter = "AND c.created_at >= ?"
			args = append(args, since)
		}
		args = append(args, "%"+opponent+"%", "%"+opponent+"%", "%"+opponent+"%")

		query := fmt.Sprintf(`
			SELECT
				c.id,
				c.case_number,
				c.title as case_name,
				c.case_type,
				c.description,
				cl.name as client_name,
				u.name as lawyer_name,
				c.created_at,
				c.lawyer_id,
				-- 相关度计算：标题匹配权重最高
				CASE
					WHEN c.title ILIKE ? THEN 3
					WHEN c.description ILIKE ? THEN 2
					WHEN cl.name ILIKE ? THEN 1
					ELSE 0
				END as relevance_score
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			JOIN users u ON c.lawyer_id = u.id
			WHERE c.deleted_at IS NULL
			AND c.lawyer_id != ?
			%s
			AND (c.title ILIKE ? OR c.description ILIKE ? OR cl.name ILIKE ?)
			ORDER BY relevance_score DESC, c.created_at DESC
			LIMIT 10
		`, sinceFilter)

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, args...).Rows()
		if err != nil {
			log.Printf("⚠️ 查询对方当事人冲突失败: %v", err)
			continue
		}

		for rows.Next() {
			var caseID uint
			var caseNo, caseName, caseType, description, clientName, lawyerName string
			var lawyerID uint
			var createdAt time.Time
			var relevanceScore int

			if err := rows.Scan(&caseID, &caseNo, &caseName, &caseType, &description, &clientName, &lawyerName, &createdAt, &lawyerID, &relevanceScore); err != nil {
				continue
			}

			// 确定冲突类型和风险等级
			conflictType := s.determineConflictType(caseName, description, opponent)
			riskLevel := s.assessConflictRisk(conflictType, createdAt)
			if s.isPartyNameMatch(clientName, opponent) {
				conflictType = "对方当事人直接冲突"
				riskLevel = "CRITICAL"
			}

			conflict := &models.ConflictCase{
				ID:              fmt.Sprintf("opponent_%d", caseID),
				CaseID:          fmt.Sprintf("%d", caseID),
				CaseName:        caseName,
				CaseNo:          caseNo,
				CaseType:        caseType,
				ConflictType:    conflictType,
				RiskLevel:       riskLevel,
				Description:     fmt.Sprintf("对方当事人 '%s' 与案件 '%s' 存在关联", opponent, caseName),
				CaseStatus:      "active",
				ClientID:        request.ClientID,
				OpposingParties: []string{opponent},
				ConflictDetails: fmt.Sprintf("系统检测到对方当事人 '%s' 与此案件存在潜在冲突", opponent),
				CreatedAt:       createdAt,
			}
			if riskLevel == "CRITICAL" {
				conflict.Description = fmt.Sprintf("当前对方当事人 '%s' 是本所历史客户 '%s'，存在直接利益冲突", opponent, clientName)
				conflict.ConflictDetails = "对方当事人与本所历史客户直接命中"
			}

			conflicts = append(conflicts, conflict)
		}
		rows.Close()
	}

	return conflicts
}

// checkClientRelationConflicts 检查客户关系冲突
func (s *conflictDetectionService) checkClientRelationConflicts(ctx context.Context, request *models.ConflictCheckRequest, since time.Time) []*models.ConflictCase {
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

		sinceFilter := ""
		args := []interface{}{relation.RelatedClientID, uint(userIDUint)}
		if !since.IsZero() {
			sinceFilter = "AND c.created_at >= ?"
			args = append(args, since)
		}

		// 查询关联客户的相关案件
		query := fmt.Sprintf(`
			SELECT
				c.id,
				c.case_number,
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
			%s
			ORDER BY c.created_at DESC
			LIMIT 5
		`, sinceFilter)

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, args...).Rows()
		if err != nil {
			continue
		}

		for rows.Next() {
			var caseID uint
			var caseNo, caseName, caseType, description, clientName, lawyerName string
			var createdAt time.Time

			if err := rows.Scan(&caseID, &caseNo, &caseName, &caseType, &description, &clientName, &lawyerName, &createdAt); err != nil {
				continue
			}

			conflict := &models.ConflictCase{
				ID:              fmt.Sprintf("relation_%d", caseID),
				CaseID:          fmt.Sprintf("%d", caseID),
				CaseName:        caseName,
				CaseNo:          caseNo,
				CaseType:        caseType,
				ConflictType:    "客户关系冲突",
				RiskLevel:       "MEDIUM",
				Description:     fmt.Sprintf("关联客户 '%s' 的案件 '%s' 存在潜在冲突", clientName, caseName),
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
func (s *conflictDetectionService) checkIndustryCompetitionConflicts(ctx context.Context, request *models.ConflictCheckRequest, since time.Time) []*models.ConflictCase {
	var conflicts []*models.ConflictCase

	// 转换用户ID为uint用于数据库查询
	userIDUint, err := strconv.ParseUint(request.UserID, 10, 32)
	if err != nil {
		log.Printf("⚠️ 解析用户ID失败: %v", err)
		return conflicts
	}

	sinceFilter := ""
	args := []interface{}{uint(userIDUint)}
	if !since.IsZero() {
		sinceFilter = "AND c.created_at >= ?"
		args = append(args, since)
	}

	// 1. 查找同一律师代理的所有案件
	query := fmt.Sprintf(`
		SELECT
			c.id,
			c.case_number,
			c.title as case_name,
			c.case_type,
			c.description,
			cl.name as client_name,
			cl.type as client_type,
			u.name as lawyer_name,
			c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.deleted_at IS NULL
		AND c.lawyer_id = ?
		%s
		ORDER BY c.created_at DESC
		LIMIT 50
	`, sinceFilter)

	rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		log.Printf("⚠️ 查询律师案件失败: %v", err)
		return conflicts
	}
	defer rows.Close()

	// 收集律师的所有客户
	lawyerClients := make(map[string]string) // clientName -> clientType
	var lawyerCases []struct {
		ID          uint
		CaseNumber  string
		Title       string
		CaseType    string
		Description string
		ClientName  string
		ClientType  string
		CreatedAt   time.Time
	}

	for rows.Next() {
		var caseModel struct {
			ID          uint
			CaseNumber  string
			Title       string
			CaseType    string
			Description string
			ClientName  string
			ClientType  string
			CreatedAt   time.Time
		}

		if err := rows.Scan(&caseModel.ID, &caseModel.CaseNumber, &caseModel.Title, &caseModel.CaseType, &caseModel.Description, &caseModel.ClientName, &caseModel.ClientType, &caseModel.CreatedAt); err != nil {
			continue
		}

		lawyerCases = append(lawyerCases, caseModel)
		lawyerClients[caseModel.ClientName] = caseModel.ClientType
	}

	log.Printf("🔍 律师代理的客户数量: %d", len(lawyerClients))

	// 2. 获取当前客户信息
	clientIDUint, err := strconv.ParseUint(request.ClientID, 10, 32)
	if err != nil {
		log.Printf("⚠️ 解析客户ID失败: %v", err)
		return conflicts
	}
	currentClient, err := s.clientRepo.FindByID(ctx, uint(clientIDUint))
	if err != nil {
		log.Printf("⚠️ 获取客户信息失败: %v", err)
		return conflicts
	}

	// 3. 检查行业竞争冲突
	for _, case_ := range lawyerCases {
		// 跳过当前客户自身的案件
		if case_.ClientName == currentClient.Name {
			continue
		}

		// 检查是否存在行业竞争关系
		conflictType, riskLevel := s.analyzeCompetitionConflict(currentClient.Name, case_.ClientName, request.OtherParties)

		if conflictType != "" {
			description := fmt.Sprintf("律师同时代理存在竞争关系的客户: '%s' 与 '%s'", currentClient.Name, case_.ClientName)
			if riskLevel == "CRITICAL" {
				description += "，存在严重利益冲突"
			} else if riskLevel == "HIGH" {
				description += "，存在商业竞争冲突"
			}

			conflict := &models.ConflictCase{
				ID:              fmt.Sprintf("industry_%d", case_.ID),
				CaseID:          fmt.Sprintf("%d", case_.ID),
				CaseName:        case_.Title,
				CaseNo:          case_.CaseNumber,
				CaseType:        case_.CaseType,
				ConflictType:    conflictType,
				RiskLevel:       riskLevel,
				Description:     description,
				CaseStatus:      "active",
				ClientID:        request.ClientID,
				ConflictDetails: fmt.Sprintf("竞争关系: %s vs %s", currentClient.Name, case_.ClientName),
				CreatedAt:       case_.CreatedAt,
			}

			conflicts = append(conflicts, conflict)
			log.Printf("⚠️ 发现行业竞争冲突: %s (风险等级: %s)", conflictType, riskLevel)
		}
	}

	// 4. 检查与对方当事人的直接冲突
	for _, opponent := range request.OtherParties {
		if strings.TrimSpace(opponent) == "" {
			continue
		}

		// 检查律师是否已经代理了对方当事人
		if _, exists := lawyerClients[opponent]; exists {
			conflict := &models.ConflictCase{
				ID:              fmt.Sprintf("direct_opponent_%d", time.Now().UnixNano()),
				CaseID:          fmt.Sprintf("OPPONENT_%d", time.Now().UnixNano()),
				CaseName:        "对方当事人冲突",
				CaseNo:          fmt.Sprintf("OPP-%d", time.Now().UnixNano()),
				CaseType:        request.CaseType,
				ConflictType:    "对方当事人冲突",
				RiskLevel:       "CRITICAL",
				Description:     fmt.Sprintf("律师同时代理当前客户 '%s' 和对方当事人 '%s'，存在直接利益冲突", currentClient.Name, opponent),
				CaseStatus:      "active",
				ClientID:        request.ClientID,
				ConflictDetails: fmt.Sprintf("禁止代理对方当事人: %s", opponent),
				CreatedAt:       time.Now(),
			}

			conflicts = append(conflicts, conflict)
			log.Printf("🚨 发现对方当事人冲突: %s", opponent)
		}

		// 检查律师代理的客户是否与对方当事人存在竞争关系
		for clientName := range lawyerClients {
			if s.hasCompetitionRelationship(clientName, opponent) {
				conflict := &models.ConflictCase{
					ID:              fmt.Sprintf("competition_%d", time.Now().UnixNano()),
					CaseID:          fmt.Sprintf("COMP-%d", time.Now().UnixNano()),
					CaseName:        "商业竞争冲突",
					CaseNo:          fmt.Sprintf("COMP-%d", time.Now().UnixNano()),
					CaseType:        request.CaseType,
					ConflictType:    "商业竞争冲突",
					RiskLevel:       "HIGH",
					Description:     fmt.Sprintf("律师代理的客户 '%s' 与对方当事人 '%s' 存在商业竞争关系", clientName, opponent),
					CaseStatus:      "active",
					ClientID:        request.ClientID,
					ConflictDetails: fmt.Sprintf("商业竞争: %s vs %s", clientName, opponent),
					CreatedAt:       time.Now(),
				}

				conflicts = append(conflicts, conflict)
				log.Printf("⚠️ 发现商业竞争冲突: %s vs %s", clientName, opponent)
			}
		}
	}

	return conflicts
}

// 辅助方法

// analyzeCompetitionConflict 分析竞争冲突
func (s *conflictDetectionService) analyzeCompetitionConflict(currentClient, otherClient string, otherParties []string) (string, string) {
	// 标准化公司名称
	currentClient = strings.ToLower(currentClient)
	otherClient = strings.ToLower(otherClient)

	// 检查是否为同一公司的关联实体
	if s.isSameCompany(currentClient, otherClient) {
		return "同一实体冲突", "CRITICAL"
	}

	// 检查是否为直接竞争对手
	if s.isDirectCompetitor(currentClient, otherClient) {
		return "商业竞争冲突", "HIGH"
	}

	// 检查是否为相关行业冲突
	if s.isRelatedIndustry(currentClient, otherClient) {
		return "行业关联冲突", "MEDIUM"
	}

	// 检查是否与对方当事人存在关联
	for _, opponent := range otherParties {
		opponent = strings.ToLower(opponent)
		if strings.Contains(currentClient, opponent) || strings.Contains(opponent, currentClient) {
			return "对方当事人关联", "HIGH"
		}
		if s.isDirectCompetitor(currentClient, opponent) || s.isDirectCompetitor(otherClient, opponent) {
			return "竞争对手关联", "HIGH"
		}
	}

	return "", "LOW"
}

// isSameCompany 检查是否为同一公司
func (s *conflictDetectionService) isSameCompany(name1, name2 string) bool {
	// 移除常见的公司后缀
	name1 = s.cleanCompanyName(name1)
	name2 = s.cleanCompanyName(name2)

	// 完全匹配
	if name1 == name2 {
		return true
	}

	// 包含关系
	if strings.Contains(name1, name2) || strings.Contains(name2, name1) {
		return true
	}

	// 检查是否为同一公司的不同实体
	sameCompanyMapping := map[string][]string{
		"阿里巴巴": {"阿里", "淘宝", "天猫", "支付宝", "蚂蚁金服", "蚂蚁集团"},
		"腾讯":   {"微信", "qq", "财付通", "腾讯云", "腾讯游戏"},
		"字节跳动": {"抖音", "tiktok", "今日头条", "西瓜视频", "飞书"},
		"百度":   {"百度网盘", "百度地图", "百度云", "小度"},
		"京东":   {"京东物流", "京东数科", "京东健康"},
		"美团":   {"美团外卖", "大众点评", "美团买菜", "美团优选"},
	}

	for company, entities := range sameCompanyMapping {
		if strings.Contains(name1, company) || strings.Contains(name2, company) {
			for _, entity := range entities {
				if (strings.Contains(name1, entity) && strings.Contains(name2, company)) ||
					(strings.Contains(name2, entity) && strings.Contains(name1, company)) {
					return true
				}
			}
		}
	}

	return false
}

func (s *conflictDetectionService) isPartyNameMatch(name1, name2 string) bool {
	name1 = s.cleanCompanyName(strings.ToLower(strings.TrimSpace(name1)))
	name2 = s.cleanCompanyName(strings.ToLower(strings.TrimSpace(name2)))
	if name1 == "" || name2 == "" {
		return false
	}
	return name1 == name2 || strings.Contains(name1, name2) || strings.Contains(name2, name1)
}

// isDirectCompetitor 检查是否为直接竞争对手
func (s *conflictDetectionService) isDirectCompetitor(name1, name2 string) bool {
	// 互联网行业主要竞争关系
	competitors := map[string][]string{
		"阿里巴巴": {"腾讯", "字节跳动", "京东", "拼多多", "百度"},
		"腾讯":   {"阿里巴巴", "字节跳动", "京东", "拼多多", "百度"},
		"字节跳动": {"阿里巴巴", "腾讯", "快手", "小红书"},
		"京东":   {"阿里巴巴", "腾讯", "拼多多", "抖音电商"},
		"拼多多":  {"阿里巴巴", "京东", "淘宝", "天猫"},
		"百度":   {"阿里巴巴", "腾讯", "字节跳动", "搜狗"},
	}

	for company, rivals := range competitors {
		if strings.Contains(name1, company) {
			for _, rival := range rivals {
				if strings.Contains(name2, rival) {
					return true
				}
			}
		}
		if strings.Contains(name2, company) {
			for _, rival := range rivals {
				if strings.Contains(name1, rival) {
					return true
				}
			}
		}
	}

	return false
}

// isRelatedIndustry 检查是否为相关行业
func (s *conflictDetectionService) isRelatedIndustry(name1, name2 string) bool {
	// 金融行业
	financeKeywords := []string{"银行", "保险", "证券", "基金", "信托", "投资", "金融"}
	// 互联网行业
	internetKeywords := []string{"科技", "网络", "软件", "信息", "互联网", "数据", "云服务"}
	// 房地产行业
	realEstateKeywords := []string{"地产", "置业", "建设", "建筑", "房地产", "物业"}
	// 制造业
	manufacturingKeywords := []string{"制造", "生产", "工厂", "加工", "设备", "机械"}

	name1InFinance := s.containsAny(name1, financeKeywords)
	name2InFinance := s.containsAny(name2, financeKeywords)
	name1InInternet := s.containsAny(name1, internetKeywords)
	name2InInternet := s.containsAny(name2, internetKeywords)
	name1InRealEstate := s.containsAny(name1, realEstateKeywords)
	name2InRealEstate := s.containsAny(name2, realEstateKeywords)
	name1InManufacturing := s.containsAny(name1, manufacturingKeywords)
	name2InManufacturing := s.containsAny(name2, manufacturingKeywords)

	// 同行业
	if (name1InFinance && name2InFinance) ||
		(name1InInternet && name2InInternet) ||
		(name1InRealEstate && name2InRealEstate) ||
		(name1InManufacturing && name2InManufacturing) {
		return true
	}

	// 相关行业（金融+互联网，房地产+建筑等）
	if (name1InFinance && name2InInternet) || (name1InInternet && name2InFinance) ||
		(name1InRealEstate && name2InManufacturing) || (name1InManufacturing && name2InRealEstate) {
		return true
	}

	return false
}

// hasCompetitionRelationship 检查竞争关系
func (s *conflictDetectionService) hasCompetitionRelationship(company1, company2 string) bool {
	company1 = strings.ToLower(company1)
	company2 = strings.ToLower(company2)

	// 使用 isDirectCompetitor 方法
	return s.isDirectCompetitor(company1, company2)
}

// cleanCompanyName 清理公司名称
func (s *conflictDetectionService) cleanCompanyName(name string) string {
	// 移除公司后缀
	suffixes := []string{"有限公司", "股份有限公司", "集团", "控股", "科技", "网络", "信息", "数据", "云", "服务", "投资", "管理", "咨询"}

	for _, suffix := range suffixes {
		name = strings.Replace(name, suffix, "", -1)
	}

	return strings.TrimSpace(name)
}

// containsAny 检查字符串是否包含任意关键词
func (s *conflictDetectionService) containsAny(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// determineConflictType 确定冲突类型（改进版本）
func (s *conflictDetectionService) determineConflictType(caseName, description, opponent string) string {
	caseName = strings.ToLower(caseName)
	description = strings.ToLower(description)
	opponent = strings.ToLower(opponent)

	// 1. 法律对立冲突 - 最高风险
	if strings.Contains(caseName, "离婚") || strings.Contains(description, "离婚") ||
		strings.Contains(caseName, "抚养") || strings.Contains(description, "抚养") ||
		strings.Contains(caseName, "继承") || strings.Contains(description, "继承") ||
		strings.Contains(caseName, "监护") || strings.Contains(description, "监护") {
		return "法律对立冲突"
	}

	// 2. 股权纠纷冲突 - 高风险
	if strings.Contains(caseName, "股权") || strings.Contains(description, "股权") ||
		strings.Contains(caseName, "收购") || strings.Contains(description, "收购") ||
		strings.Contains(caseName, "并购") || strings.Contains(description, "并购") ||
		strings.Contains(caseName, "投资") || strings.Contains(description, "投资") ||
		strings.Contains(caseName, "持股") || strings.Contains(description, "持股") {
		return "股权纠纷冲突"
	}

	// 3. 知识产权冲突 - 高风险
	if strings.Contains(caseName, "商标") || strings.Contains(description, "商标") ||
		strings.Contains(caseName, "专利") || strings.Contains(description, "专利") ||
		strings.Contains(caseName, "版权") || strings.Contains(description, "版权") ||
		strings.Contains(caseName, "知识产权") || strings.Contains(description, "知识产权") ||
		strings.Contains(caseName, "侵权") || strings.Contains(description, "侵权") {
		return "知识产权冲突"
	}

	// 4. 服务纠纷冲突 - 中等风险
	if strings.Contains(caseName, "合同") || strings.Contains(description, "合同") ||
		strings.Contains(caseName, "服务") || strings.Contains(description, "服务") ||
		strings.Contains(caseName, "劳务") || strings.Contains(description, "劳务") ||
		strings.Contains(caseName, "咨询") || strings.Contains(description, "咨询") {
		return "服务纠纷冲突"
	}

	// 5. 建筑工程纠纷 - 高风险
	if strings.Contains(caseName, "工程") || strings.Contains(description, "工程") ||
		strings.Contains(caseName, "建设") || strings.Contains(description, "建设") ||
		strings.Contains(caseName, "施工") || strings.Contains(description, "施工") ||
		strings.Contains(caseName, "建筑") || strings.Contains(description, "建筑") {
		return "服务纠纷冲突"
	}

	// 6. 劳动纠纷 - 中等风险
	if strings.Contains(caseName, "劳动") || strings.Contains(description, "劳动") ||
		strings.Contains(caseName, "工伤") || strings.Contains(description, "工伤") ||
		strings.Contains(caseName, "解雇") || strings.Contains(description, "解雇") {
		return "服务纠纷冲突"
	}

	// 7. 客户关系冲突 - 中等风险
	if strings.Contains(caseName, "客户") || strings.Contains(description, "客户") ||
		strings.Contains(caseName, "合作") || strings.Contains(description, "合作") ||
		strings.Contains(caseName, "伙伴") || strings.Contains(description, "伙伴") {
		return "客户关系冲突"
	}

	// 8. 默认为商业竞争冲突 - 高风险
	return "商业竞争冲突"
}

// analyzeConflictWithBestPractices 基于最佳实践的冲突分析
func (s *conflictDetectionService) analyzeConflictWithBestPractices(caseName, description, opponent string, createdAt time.Time, clientID string) (string, string) {
	// 基于最佳实践的多维度冲突分析
	caseNameLower := strings.ToLower(caseName)
	descriptionLower := strings.ToLower(description)
	opponentLower := strings.ToLower(opponent)

	// 1. 分析法律关系冲突
	if s.containsAny(caseNameLower, []string{"离婚", "抚养", "继承", "监护"}) ||
		s.containsAny(descriptionLower, []string{"离婚", "抚养", "继承", "监护"}) {
		return "法律对立冲突", "CRITICAL"
	}

	// 2. 分析股权纠纷
	if s.containsAny(caseNameLower, []string{"股权", "收购", "并购", "投资", "持股"}) ||
		s.containsAny(descriptionLower, []string{"股权", "收购", "并购", "投资", "持股"}) {
		return "股权纠纷冲突", "HIGH"
	}

	// 3. 分析知识产权冲突
	if s.containsAny(caseNameLower, []string{"商标", "专利", "版权", "知识产权", "侵权"}) ||
		s.containsAny(descriptionLower, []string{"商标", "专利", "版权", "知识产权", "侵权"}) {
		return "知识产权冲突", "HIGH"
	}

	// 4. 分析服务合同冲突
	if s.containsAny(caseNameLower, []string{"合同", "服务", "协议", "合作"}) ||
		s.containsAny(descriptionLower, []string{"合同", "服务", "协议", "合作"}) {
		return "服务纠纷冲突", "MEDIUM"
	}

	// 5. 基于时间的风险衰减
	hoursPassed := time.Since(createdAt).Hours()
	timeRiskFactor := s.calculateTimeRiskFactor(hoursPassed)

	// 6. 综合评估确定最终风险等级
	return s.determineFinalRiskLevel(caseNameLower, opponentLower, timeRiskFactor)
}

// calculateTimeRiskFactor 计算时间风险因子
func (s *conflictDetectionService) calculateTimeRiskFactor(hoursPassed float64) string {
	if hoursPassed > 8760 { // 1年
		return "LOW_TIME_FACTOR"
	} else if hoursPassed > 4380 { // 6个月
		return "MEDIUM_TIME_FACTOR"
	} else {
		return "HIGH_TIME_FACTOR"
	}
}

// determineFinalRiskLevel 确定最终风险等级
func (s *conflictDetectionService) determineFinalRiskLevel(caseName, opponent string, timeRiskFactor string) (string, string) {
	// 检查是否存在直接竞争关系
	if s.isDirectCompetitor(caseName, opponent) {
		return "商业竞争冲突", "CRITICAL"
	}

	// 检查关联公司关系
	if s.isSameCompany(caseName, opponent) {
		return "同一实体冲突", "CRITICAL"
	}

	// 基于时间因子调整风险等级
	baseRisk := "HIGH"
	if timeRiskFactor == "LOW_TIME_FACTOR" {
		baseRisk = "MEDIUM"
	} else if timeRiskFactor == "LOW_TIME_FACTOR" && s.isRelatedIndustry(caseName, opponent) {
		baseRisk = "MEDIUM" // 时间长且行业关联，降低风险
	}

	return baseRisk, baseRisk
}

// generateDetailedConflictDescription 生成详细的冲突描述
func (s *conflictDetectionService) generateDetailedConflictDescription(conflictType, caseName, opponent, conflictCategory string, caseID string) string {
	switch conflictType {
	case "法律对立冲突":
		return fmt.Sprintf("⚠️ 严重法律冲突：案件 '%s' 与对方当事人 '%s' 存在直接法律对立关系，案件ID: %s。建议：立即拒绝代理或采取隔离措施。", caseName, opponent, caseID)
	case "股权纠纷冲突":
		return fmt.Sprintf("⚠️ 股权利益冲突：案件 '%s' 与对方当事人 '%s' 存在股权纠纷关系，案件ID: %s。建议：详细审查股权结构，评估潜在影响。", caseName, opponent, caseID)
	case "知识产权冲突":
		return fmt.Sprintf("⚠️ 知识产权冲突：案件 '%s' 与对方当事人 '%s' 存在知识产权纠纷，案件ID: %s。建议：进行知识产权交叉检查，避免侵权风险。", caseName, opponent, caseID)
	case "服务纠纷冲突":
		return fmt.Sprintf("⚠️ 服务合同冲突：案件 '%s' 与对方当事人 '%s' 存在服务合同冲突，案件ID: %s。建议：审查合同条款，评估履约冲突。", caseName, opponent, caseID)
	case "商业竞争冲突":
		return fmt.Sprintf("⚠️ 商业竞争冲突：案件 '%s' 与对方当事人 '%s' 存在直接商业竞争关系，案件ID: %s。建议：评估市场竞争影响，考虑利益冲突范围。", caseName, opponent, caseID)
	case "行业关联冲突":
		return fmt.Sprintf("⚠️ 行业关联冲突：案件 '%s' 与对方当事人 '%s' 存在行业关联关系，案件ID: %s。建议：评估行业影响范围，适度限制代理。", caseName, opponent, caseID)
	default:
		return fmt.Sprintf("ℹ️ 一般利益冲突：案件 '%s' 与对方当事人 '%s' 存在潜在利益冲突，案件ID: %s。建议：进一步调查具体冲突情况。", caseName, opponent, caseID)
	}
}

// assessConflictRisk 评估冲突风险（改进版本）
func (s *conflictDetectionService) assessConflictRisk(conflictType string, createdAt time.Time) string {
	// 基于冲突类型的基础风险
	baseRisk := map[string]string{
		"法律对立冲突": "CRITICAL",
		"股权纠纷冲突": "HIGH",
		"知识产权冲突": "HIGH",
		"服务纠纷冲突": "MEDIUM",
		"商业竞争冲突": "HIGH", // 修改：从HIGH改为HIGH，保持商业竞争的高风险
		"客户关系冲突": "MEDIUM",
		"行业竞争冲突": "HIGH",
	}

	riskLevel := baseRisk[conflictType]
	if riskLevel == "" {
		riskLevel = "MEDIUM"
	}

	// 时间衰减：适度降低风险，但保持合理的风险评估
	hoursPassed := time.Since(createdAt).Hours()
	if hoursPassed > 4320 { // 6个月
		if riskLevel == "CRITICAL" {
			riskLevel = "HIGH"
		} else if riskLevel == "HIGH" {
			riskLevel = "MEDIUM"
		}
		// 注意：MEDIUM级别不再降级到LOW，保持合理的风险评估
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
		TotalCasesChecked:         int64(len(conflicts)),
		ClientHistoryCases:        int64(len(conflicts)), // 简化处理
		RelatedPartiesChecked:     int64(len(request.OtherParties) + 1),
		CorporateRelationsChecked: 0, // 简化处理
		TimeRange:                 timeRange,
		SearchScope:               request.SearchDepth,
		StartTime:                 endTime,
		EndTime:                   endTime,
	}
}

func (s *conflictDetectionService) searchStartTime(request *models.ConflictCheckRequest) time.Time {
	if request.SearchYears <= 0 {
		return time.Time{}
	}
	return time.Now().AddDate(-request.SearchYears, 0, 0)
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

	riskLevel := "LOW"
	if response.RiskAssessment != nil {
		riskLevel = response.RiskAssessment.OverallRisk
	}

	now := time.Now()
	record := &models.ConflictCheckRecord{
		CheckID:          response.CheckID,
		ClientID:         request.ClientID,
		ClientName:       request.ClientName,
		CaseName:         request.CaseName,
		CaseType:         request.CaseType,
		CheckStatus:      "COMPLETED",
		HasConflict:      response.HasConflict,
		RiskLevel:        riskLevel,
		SearchParameters: toConflictJSON(request),
		CheckResult:      toConflictJSON(response),
		UserID:           uint(userIDUint),
		Duration:         response.Duration,
		CheckTime:        response.CheckTime,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.conflictRepo.SaveCheckRecord(ctx, record); err != nil {
		return err
	}

	if len(response.ConflictCases) == 0 {
		return nil
	}

	for _, conflict := range response.ConflictCases {
		if conflict == nil {
			continue
		}
		conflict.CheckID = response.CheckID
		conflict.ID = fmt.Sprintf("%s_%s", response.CheckID, conflict.ID)
		if conflict.CaseType == "" {
			conflict.CaseType = request.CaseType
		}
		if conflict.RiskLevel == "" {
			conflict.RiskLevel = "LOW"
		}
		if conflict.OpposingParties == nil {
			conflict.OpposingParties = models.JSONStringArray{}
		}
		if conflict.CreatedAt.IsZero() {
			conflict.CreatedAt = now
		}
	}

	return s.conflictRepo.SaveConflictCases(ctx, response.ConflictCases)
}

func toConflictJSON(value interface{}) models.JSON {
	if value == nil {
		return models.JSON{}
	}
	if data, ok := value.(models.JSON); ok {
		return data
	}
	if data, ok := value.(map[string]interface{}); ok {
		return models.JSON(data)
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return models.JSON{"marshal_error": err.Error()}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(raw, &result); err == nil {
		return models.JSON(result)
	}

	var generic interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		return models.JSON{"unmarshal_error": err.Error(), "raw": string(raw)}
	}
	return models.JSON{"items": generic}
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
		"金融":    {"银行", "保险", "证券", "基金", "金融", "投资"},
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
