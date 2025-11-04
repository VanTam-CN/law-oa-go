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

// EnhancedConflictDetectionService 增强版冲突检测服务
type EnhancedConflictDetectionService interface {
	// 核心功能
	PerformProfessionalConflictCheck(ctx context.Context, request *models.ProfessionalConflictCheckRequest) (*models.ProfessionalConflictCheckResponse, error)
	ValidateConflictCheckRequest(ctx context.Context, request *models.ProfessionalConflictCheckRequest) error

	// 规则引擎
	ExecuteDetectionRules(ctx context.Context, requestID string) ([]*models.ConflictRuleExecution, error)
	GetActiveRules(ctx context.Context) ([]*models.ConflictDetectionRule, error)

	// 客户信息管理
	EnrichClientInformation(ctx context.Context, clientID string) (*models.EnrichedClientProfile, error)
	SearchClientNameVariants(ctx context.Context, searchTerm string) ([]*models.ClientNameVariant, error)
	AnalyzeClientRelationships(ctx context.Context, clientID string) ([]*models.ClientRelationship, error)

	// 分类和风险评估
	ClassifyConflicts(ctx context.Context, conflicts []*models.MultiDimensionalConflict) ([]*models.ClassifiedConflict, error)
	AssessProfessionalRisk(ctx context.Context, classifiedConflicts []*models.ClassifiedConflict, request *models.ProfessionalConflictCheckRequest) (*models.ProfessionalRiskAssessment, error)

	// 豁免管理
	GenerateWaiverRecommendations(ctx context.Context, classifiedConflicts []*models.ClassifiedConflict) (*models.WaiverRecommendation, error)

	// 监控和审计
	GetConflictCheckHistory(ctx context.Context, params *models.ConflictCheckHistoryParams) ([]*models.ConflictCheckRecord, error)
	LogConflictCheckExecution(ctx context.Context, request *models.ProfessionalConflictCheckRequest, response *models.ProfessionalConflictCheckResponse) error

	// 统计和报告
	GenerateConflictStatistics(ctx context.Context, params *models.ConflictStatisticsParams) (*models.ConflictStatistics, error)
	GenerateComplianceReport(ctx context.Context, request *models.ComplianceReportRequest) (*models.ComplianceReport, error)
}

// enhancedConflictDetectionService 增强版冲突检测服务实现
type enhancedConflictDetectionService struct {
	// 核心仓库
	conflictRepo          repositories.EnhancedConflictRepository
	clientRepo           repositories.ClientRepository
	caseRepo             repositories.CaseRepository
	userRepo             repositories.UserRepository

	// 专业服务
	classificationService ConflictClassificationService
	riskAssessmentService ProfessionalRiskAssessmentService
	waiverService        WaiverManagementService
	auditLogger          AuditLogger

	// 配置
	config *ConflictDetectionConfig
}

// ConflictDetectionConfig 冲突检测配置
type ConflictDetectionConfig struct {
	DefaultSearchYears    int                    `json:"default_search_years"`
	DefaultSearchDepth    string                 `json:"default_search_depth"`
	EnableMLDetection     bool                  `json:"enable_ml_detection"`
	RiskTolerance         string                 `json:"risk_tolerance"`
	AutoApproveThreshold float64               `json:"auto_approve_threshold"`
	CacheTTLSecs          int                    `json:"cache_ttl_secs"`
}

// NewEnhancedConflictDetectionService 创建增强版冲突检测服务
func NewEnhancedConflictDetectionService(
	conflictRepo repositories.EnhancedConflictRepository,
	clientRepo repositories.ClientRepository,
	caseRepo repositories.CaseRepository,
	userRepo repositories.UserRepository,
	classificationService ConflictClassificationService,
	riskAssessmentService ProfessionalRiskAssessmentService,
	waiverService WaiverManagementService,
	auditLogger AuditLogger,
	config *ConflictDetectionConfig,
) EnhancedConflictDetectionService {
	return &enhancedConflictDetectionService{
		conflictRepo:          conflictRepo,
		clientRepo:           clientRepo,
		caseRepo:             caseRepo,
		userRepo:             userRepo,
		classificationService: classificationService,
		riskAssessmentService: riskAssessmentService,
		waiverService:        waiverService,
		auditLogger:          auditLogger,
		config:               config,
	}
}

// PerformProfessionalConflictCheck 执行专业冲突检查
func (s *enhancedConflictDetectionService) PerformProfessionalConflictCheck(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
) (*models.ProfessionalConflictCheckResponse, error) {
	startTime := time.Now()

	log.Printf("🚀 开始执行专业冲突检查: 检查ID=%s, 客户=%s, 案件=%s, 律师=%s",
		request.ID, request.ClientName, request.CaseName, request.PrimaryLawyerName)

	// 1. 验证请求
	if err := s.ValidateConflictCheckRequest(ctx, request); err != nil {
		log.Printf("❌ 冲突检查请求验证失败: %v", err)
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 2. 丰富客户信息
	enrichedClient, err := s.EnrichClientInformation(ctx, request.ClientID)
	if err != nil {
		log.Printf("⚠️ 丰富客户信息失败: %v", err)
		// 继续处理，但记录警告
		enrichedClient = &models.EnrichedClientProfile{
			ClientID:   request.ClientID,
			ClientName: request.ClientName,
		}
	}

	// 3. 执行多维度冲突检测
	conflicts, err := s.detectMultiDimensionalConflicts(ctx, request, enrichedClient)
	if err != nil {
		log.Printf("❌ 多维度冲突检测失败: %v", err)
		return nil, fmt.Errorf("冲突检测失败: %w", err)
	}

	log.Printf("📋 检测到 %d 个潜在冲突", len(conflicts))

	// 4. 执行检测规则
	ruleExecutions, err := s.ExecuteDetectionRules(ctx, request.ID)
	if err != nil {
		log.Printf("⚠️ 规则引擎执行失败: %v", err)
		// 继续处理，但记录警告
		ruleExecutions = []*models.ConflictRuleExecution{}
	}

	// 5. 冲突分类
	classifiedConflicts, err := s.ClassifyConflicts(ctx, conflicts)
	if err != nil {
		log.Printf("❌ 冲突分类失败: %v", err)
		return nil, fmt.Errorf("冲突分类失败: %w", err)
	}

	// 6. 专业风险评估
	riskAssessment, err := s.AssessProfessionalRisk(ctx, classifiedConflicts, request)
	if err != nil {
		log.Printf("❌ 风险评估失败: %v", err)
		return nil, fmt.Errorf("风险评估失败: %w", err)
	}

	// 7. 生成处理建议
	recommendations := s.generateRecommendations(ctx, classifiedConflicts, riskAssessment, request)

	// 8. 生成豁免建议
	waiverRecommendation, err := s.waiverService.GenerateWaiverRecommendations(ctx, classifiedConflicts)
	if err != nil {
		log.Printf("⚠️ 豁免建议生成失败: %v", err)
		waiverRecommendation = &models.WaiverRecommendation{}
	}

	// 9. 构建响应
	response := &models.ProfessionalConflictCheckResponse{
		RequestID:          request.ID,
		CheckNumber:         request.CheckNumber,
		ClientName:          request.ClientName,
		CaseName:            request.CaseName,
		CaseType:            request.CaseType,
		PrimaryLawyer:      request.PrimaryLawyerName,
		CheckStatus:         "COMPLETED",
		HasConflict:         len(classifiedConflicts) > 0,
		Conflicts:           classifiedConflicts,
		RiskAssessment:      riskAssessment,
		Recommendations:     recommendations,
		WaiverRecommendation: waiverRecommendation,
		RuleExecutions:      ruleExecutions,
		EnrichedClient:      enrichedClient,
		CheckDate:           time.Now(),
		ProcessingDuration:  time.Since(startTime).Milliseconds(),
		GeneratedAt:         time.Now(),
	}

	// 10. 保存检查记录
	if err := s.saveProfessionalCheckRecord(ctx, request, response); err != nil {
		log.Printf("⚠️ 保存检查记录失败: %v", err)
		// 不影响主流程
	}

	// 11. 审计记录
	if err := s.auditLogger.LogConflictCheckExecution(ctx, request, response); err != nil {
		log.Printf("⚠️ 审计记录失败: %v", err)
		// 不影响主流程
	}

	log.Printf("✅ 专业冲突检查完成: 检测ID=%s, 冲突数=%d, 风险等级=%s, 耗时=%dms",
		response.RequestID, len(response.Conflicts), response.RiskAssessment.OverallRisk, response.ProcessingDuration)

	return response, nil
}

// detectMultiDimensionalConflicts 检测多维度冲突
func (s *enhancedConflictDetectionService) detectMultiDimensionalConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
	enrichedClient *models.EnrichedClientProfile,
) ([]*models.MultiDimensionalConflict, error) {
	var allConflicts []*models.MultiDimensionalConflict

	// 1. 客户名称变体检测
	nameVariantConflicts, err := s.detectNameVariantConflicts(ctx, request, enrichedClient)
	if err != nil {
		log.Printf("⚠️ 名称变体冲突检测失败: %v", err)
	} else {
		allConflicts = append(allConflicts, nameVariantConflicts...)
	}

	// 2. 关联关系冲突检测
	relationshipConflicts, err := s.detectRelationshipConflicts(ctx, request, enrichedClient)
	if err != nil {
		log.Printf("⚠️ 关联关系冲突检测失败: %v", err)
	} else {
		allConflicts = append(allConflicts, relationshipConflicts...)
	}

	// 3. 律师代理历史冲突检测
	lawyerHistoryConflicts, err := s.detectLawyerHistoryConflicts(ctx, request)
	if err != nil {
		log.Printf("⚠️ 律师历史冲突检测失败: %v", err)
	} else {
		allConflicts = append(allConflicts, lawyerHistoryConflicts...)
	}

	// 4. 对方当事人冲突检测
	opposingPartyConflicts, err := s.detectOpposingPartyConflicts(ctx, request)
	if err != nil {
		log.Printf("⚠️ 对方当事人冲突检测失败: %v", err)
	} else {
		allConflicts = append(allConflicts, opposingPartyConflicts...)
	}

	// 5. 时间序列冲突检测
	temporalConflicts, err := s.detectTemporalConflicts(ctx, request)
	if err != nil {
		log.Printf("⚠️ 时间序列冲突检测失败: %v", err)
	} else {
		allConflicts = append(allConflicts, temporalConflicts...)
	}

	// 6. 财务利益冲突检测
	financialConflicts, err := s.detectFinancialConflicts(ctx, request)
	if err != nil {
		log.Printf("⚠️ 财务利益冲突检测失败: %v", err)
	} else {
		allConflicts = append(allConflicts, financialConflicts...)
	}

	// 7. 专业领域特定冲突检测
	professionalConflicts, err := s.detectProfessionalConflicts(ctx, request)
	if err != nil {
		log.Printf("⚠️ 专业领域冲突检测失败: %v", err)
	} else {
		allConflicts = append(allConflicts, professionalConflicts...)
	}

	// 8. 去重和优化
	optimizedConflicts := s.deduplicateConflicts(allConflicts)

	// 9. 计算置信度和风险评分
	for _, conflict := range optimizedConflicts {
		conflict.DetectionConfidence = s.calculateDetectionConfidence(conflict)
		conflict.RiskScore = s.calculateRiskScore(conflict)
	}

	// 10. 按风险评分排序
	s.sortConflictsByRiskScore(optimizedConflicts)

	log.Printf("📊 多维度冲突检测完成: 原始冲突=%d, 去重后=%d", len(allConflicts), len(optimizedConflicts))

	return optimizedConflicts, nil
}

// detectNameVariantConflicts 检测名称变体冲突
func (s *enhancedConflictDetectionService) detectNameVariantConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
	enrichedClient *models.EnrichedClientProfile,
) ([]*models.MultiDimensionalConflict, error) {
	var conflicts []*models.MultiDimensionalConflict

	// 解析对方当事人列表
	var opposingParties []string
	if err := json.Unmarshal(request.OpposingParties, &opposingParties); err != nil {
		log.Printf("⚠️ 解析对方当事人列表失败: %v", err)
		return conflicts, nil
	}

	// 检查所有名称变体
	nameVariants := []string{
		enrichedClient.ClientName,
	}

	// 添加名称变体
	for _, variant := range enrichedClient.NameVariants {
		nameVariants = append(nameVariants, variant.VariantName)
	}

	// 检查关联关系
	for _, relationship := range enrichedClient.Relationships {
		nameVariants = append(nameVariants, relationship.RelatedEntityName)
	}

	// 查询历史案件
	for _, variant := range nameVariants {
		if strings.TrimSpace(variant) == "" {
			continue
		}

		// 查询包含变体名称的案件
		query := `
			SELECT
				c.id,
				c.title as case_name,
				c.case_type,
				c.description,
				cl.name as client_name,
				cl.id as client_id,
				u.name as lawyer_name,
				u.id as lawyer_id,
				c.created_at
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			JOIN users u ON c.lawyer_id = u.id
			WHERE c.deleted_at IS NULL
			AND (c.title LIKE ? OR c.description LIKE ? OR cl.name LIKE ?)
			ORDER BY c.created_at DESC
			LIMIT 20
		`

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query,
			"%"+variant+"%", "%"+variant+"%", "%"+variant+"%").Rows()
		if err != nil {
			log.Printf("⚠️ 查询名称变体冲突失败: %v", err)
			continue
		}

		for rows.Next() {
			var caseID uint
			var caseName, caseType, description, clientName, lawyerName string
			var clientID uint
			var lawyerID uint
			var createdAt time.Time

			if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &clientID, &lawyerName, &lawyerID, &createdAt); err != nil {
				log.Printf("⚠️ 扫描案件数据失败: %v", err)
				continue
			}

			// 跳过当前律师的案件
			if lawyerID == s.parseLawyerID(request.PrimaryLawyerID) {
				continue
			}

			// 检查是否与对方当事人匹配
			var isRelatedToOpposingParty bool
			for _, opponent := range opposingParties {
				if s.isRelatedToOpposingParty(caseName, description, clientName, opponent) {
					isRelatedToOpposingParty = true
					break
				}
			}

			// 创建冲突记录
			conflict := &models.MultiDimensionalConflict{
				ID:                     s.generateConflictID(),
				ConflictTitle:          fmt.Sprintf("名称匹配冲突: %s vs %s", variant, clientName),
				ConflictDescription:    fmt.Sprintf("检测到客户名称变体 '%s' 与历史案件客户 '%s' 存在匹配", variant, clientName),
				ConflictType:           "RELATIONSHIP",
				ConflictSubtype:        "NAME_VARIANT",
				SourceEntityType:       "CLIENT",
				SourceEntityID:         enrichedClient.ClientID,
				SourceEntityName:       enrichedClient.ClientName,
				RelatedEntityID:         fmt.Sprintf("%d", clientID),
				RelatedEntityName:       clientName,
				ConflictTimeDimension:   "HISTORICAL",
				ConflictDate:            createdAt,
				BusinessDimension:       "CASE_RELATED",
				PracticeArea:           caseType,
				DetectionMethod:         "RULE_BASED",
				RecommendedActions:      []string{"详细审查关联关系", "评估冲突影响"},
				SupportingEvidence:       []string{"案件记录", "客户关系数据"},
				RiskFactors:             []string{"名称相似性", "历史关系"},
			}

			if isRelatedToOpposingParty {
				conflict.RiskLevel = "HIGH"
			} else {
				conflict.RiskLevel = "MEDIUM"
			}

			conflicts = append(conflicts, conflict)
		}
		rows.Close()
	}

	return conflicts, nil
}

// detectRelationshipConflicts 检测关联关系冲突
func (s *enhancedConflictDetectionService) detectRelationshipConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
	enrichedClient *models.EnrichedClientProfile,
) ([]*models.MultiDimensionalConflict, error) {
	var conflicts []*models.MultiDimensionalConflict

	// 解析对方当事人列表
	var opposingParties []string
	if err := json.Unmarshal(request.OpposingParties, &opposingParties); err != nil {
		log.Printf("⚠️ 解析对方当事人列表失败: %v", err)
		return conflicts, nil
	}

	// 检查客户关系
	for _, relationship := range enrichedClient.Relationships {
		// 检查是否与对方当事人存在关联
		for _, opponent := range opposingParties {
			if s.isRelatedEntity(relationship.RelatedEntityName, opponent) {
				conflict := &models.MultiDimensionalConflict{
					ID:                     s.generateConflictID(),
					ConflictTitle:          fmt.Sprintf("关联关系冲突: %s - %s", relationship.RelatedEntityName, opponent),
					ConflictDescription:    fmt.Sprintf("客户 '%s' 的关联实体 '%s' 与对方当事人 '%s' 存在 %s 关系",
						enrichedClient.ClientName, relationship.RelatedEntityName, opponent, relationship.RelationshipType),
					ConflictType:           "RELATIONSHIP",
					ConflictSubtype:        relationship.RelationshipType,
					SourceEntityType:       "CLIENT",
					SourceEntityID:         enrichedClient.ClientID,
					SourceEntityName:       enrichedClient.ClientName,
					RelatedEntityID:         relationship.RelatedEntityID,
					RelatedEntityName:       relationship.RelatedEntityName,
					ConflictTimeDimension:   "CURRENT",
					BusinessDimension:       "CLIENT_RELATED",
					PracticeArea:           request.PracticeArea,
					Jurisdiction:            request.Jurisdiction,
					DetectionMethod:         "RULE_BASED",
					RecommendedActions:      []string{"详细评估关联关系", "考虑代理限制"},
					SupportingEvidence:       []string{"客户关系记录", "关联关系证明"},
					RiskFactors:             []string{"关联关系强度", "利益冲突风险"},
				}

				// 根据关系类型和强度确定风险等级
				conflict.RiskLevel = s.determineRelationshipRiskLevel(relationship)

				conflicts = append(conflicts, conflict)
			}
		}

		// 检查与历史案件的关联
		historyConflicts, err := s.detectRelationshipHistoryConflicts(ctx, relationship, request)
		if err != nil {
			log.Printf("⚠️ 检查关联历史冲突失败: %v", err)
		} else {
			conflicts = append(conflicts, historyConflicts...)
		}
	}

	return conflicts, nil
}

// detectLawyerHistoryConflicts 检测律师代理历史冲突
func (s *enhancedConflictDetectionService) detectLawyerHistoryConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
) ([]*models.MultiDimensionalConflict, error) {
	var conflicts []*models.MultiDimensionalConflict

	lawyerID, err := s.parseLawyerID(request.PrimaryLawyerID)
	if err != nil {
		log.Printf("⚠️ 解析律师ID失败: %v", err)
		return conflicts, nil
	}

	clientID, err := s.parseClientID(request.ClientID)
	if err != nil {
		log.Printf("⚠️ 解析客户ID失败: %v", err)
		return conflicts, nil
	}

	// 查询律师代理的历史案件
	query := `
		SELECT
			c.id,
			c.title as case_name,
			c.case_type,
			c.description,
			cl.name as client_name,
			cl.id as client_id,
			c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = ?
		  AND c.client_id != ?
		  AND c.deleted_at IS NULL
		  AND c.created_at >= DATE_SUB(NOW(), INTERVAL ? YEAR)
		ORDER BY c.created_at DESC
		LIMIT 50
	`

	rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, lawyerID, clientID, request.SearchYears).Rows()
	if err != nil {
		return nil, fmt.Errorf("查询律师历史案件失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var caseID uint
		var caseName, caseType, description, clientName string
		var historyClientID uint
		var createdAt time.Time

		if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &historyClientID, &createdAt); err != nil {
			log.Printf("⚠️ 扫描律师历史案件失败: %v", err)
			continue
		}

		// 计算时间相关性
		timeRelevance := s.calculateTimeRelevance(createdAt)

		conflict := &models.MultiDimensionalConflict{
			ID:                     s.generateConflictID(),
			ConflictTitle:          fmt.Sprintf("律师历史代理冲突: %s", caseName),
			ConflictDescription:    fmt.Sprintf("律师 '%s' 历史代理案件 '%s'，客户 '%s'，存在潜在利益冲突", request.PrimaryLawyerName, caseName, clientName),
			ConflictType:           "CONCURRENT",
			ConflictSubtype:        "LAWYER_HISTORY",
			SourceEntityType:       "LAWYER",
			SourceEntityID:         request.PrimaryLawyerID,
			SourceEntityName:       request.PrimaryLawyerName,
			RelatedEntityID:         fmt.Sprintf("%d", historyClientID),
			RelatedEntityName:       clientName,
			ConflictTimeDimension:   "HISTORICAL",
			ConflictDate:            createdAt,
			TimeRelevance:          timeRelevance,
			BusinessDimension:       "LAWYER_RELATED",
			PracticeArea:           caseType,
			DetectionMethod:         "RULE_BASED",
			RecommendedActions:      []string{"评估保密义务", "考虑信息屏障", "获取客户同意"},
			SupportingEvidence:       []string{"律师代理记录", "案件档案"},
			RiskFactors:             []string{"代理历史", "保密义务", "时间相关性"},
		}

		// 根据时间相关性和案件性质确定风险等级
		conflict.RiskLevel = s.calculateLawyerHistoryRiskLevel(conflict, request)

		conflicts = append(conflicts, conflict)
	}

	return conflicts, nil
}

// detectOpposingPartyConflicts 检测对方当事人冲突
func (s *enhancedConflictDetectionService) detectOpposingPartyConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
) ([]*models.MultiDimensionalConflict, error) {
	var conflicts []*models.MultiDimensionalConflict

	// 解析对方当事人列表
	var opposingParties []string
	if err := json.Unmarshal(request.OpposingParties, &opposingParties); err != nil {
		log.Printf("⚠️ 解析对方当事人列表失败: %v", err)
		return conflicts, nil
	}

	lawyerID, err := s.parseLawyerID(request.PrimaryLawyerID)
	if err != nil {
		log.Printf("⚠️ 解析律师ID失败: %v", err)
		return conflicts, nil
	}

	// 检查每个对方当事人
	for _, opponent := range opposingParties {
		if strings.TrimSpace(opponent) == "" {
			continue
		}

		// 查询律师是否代理过对方当事人
		query := `
			SELECT
				c.id,
				c.title as case_name,
				c.case_type,
				c.description,
				cl.name as client_name,
				cl.id as client_id,
				c.created_at
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			WHERE c.lawyer_id = ?
			  AND (cl.name LIKE ? OR c.title LIKE ? OR c.description LIKE ?)
			  AND c.deleted_at IS NULL
			  AND c.created_at >= DATE_SUB(NOW(), INTERVAL ? YEAR)
			ORDER BY c.created_at DESC
			LIMIT 20
		`

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query,
			lawyerID, "%"+opponent+"%", "%"+opponent+"%", "%"+opponent+"%", request.SearchYears).Rows()
		if err != nil {
			log.Printf("⚠️ 查询对方当事人冲突失败: %v", err)
			continue
		}

		for rows.Next() {
			var caseID uint
			var caseName, caseType, description, clientName string
			var historyClientID uint
			var createdAt time.Time

			if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &historyClientID, &createdAt); err != nil {
				log.Printf("⚠️ 扫描对方当事人冲突数据失败: %v", err)
				continue
			}

			conflict := &models.MultiDimensionalConflict{
				ID:                     s.generateConflictID(),
				ConflictTitle:          fmt.Sprintf("对方当事人冲突: %s", opponent),
				ConflictDescription:    fmt.Sprintf("律师 '%s' 现代理客户 '%s' 的同时，历史上曾代理对方当事人 '%s' 的案件 '%s'",
					request.PrimaryLawyerName, request.ClientName, opponent, caseName),
				ConflictType:           "DIRECT_OPPOSITION",
				ConflictSubtype:        "OPPOSING_PARTY_HISTORY",
				SourceEntityType:       "LAWYER",
				SourceEntityID:         request.PrimaryLawyerID,
				SourceEntityName:       request.PrimaryLawyerName,
				RelatedEntityID:         fmt.Sprintf("%d", historyClientID),
				RelatedEntityName:       opponent,
				ConflictTimeDimension:   "HISTORICAL",
				ConflictDate:            createdAt,
				BusinessDimension:       "LAWYER_RELATED",
				PracticeArea:           caseType,
				DetectionMethod:         "RULE_BASED",
				RecommendedActions:      []string{"立即停止代理", "建议更换律师", "记录冲突详情"},
				SupportingEvidence:       []string{"历史案件记录", "律师代理档案"},
				RiskFactors:             []_array{"对方当事人代理历史", "直接利益冲突", "职业道德风险"},
			}

			conflict.RiskLevel = "CRITICAL"
			conflicts = append(conflicts, conflict)
		}
		rows.Close()

		// 检查当前代理的客户是否与对方当事人存在商业关系
		businessConflicts := s.detectBusinessRelationshipConflicts(ctx, request, opponent)
		conflicts = append(conflicts, businessConflicts...)
	}

	return conflicts, nil
}

// detectTemporalConflicts 检测时间序列冲突
func (s *enhancedConflictDetectionService) detectTemporalConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
) ([]*models.MultiDimensionalConflict, error) {
	var conflicts []*models.MultiDimensionalConflict

	// 解析团队成员
	var teamMembers []string
	if err := json.Unmarshal(request.TeamMembers, &teamMembers); err != nil {
		log.Printf("⚠️ 解析团队成员列表失败: %v", err)
		return conflicts, nil
	}

	// 检查每个团队成员的代理历史
	for _, member := range teamMembers {
		memberID, err := s.parseLawyerID(member)
		if err != nil {
			log.Printf("⚠️ 解析团队成员ID失败: %v", err)
			continue
		}

		// 查询团队成员在指定时间内的代理记录
		query := `
			SELECT
				c.id,
				c.title as case_name,
				c.case_type,
				c.description,
				cl.name as client_name,
				cl.id as client_id,
				c.created_at,
				c.completed_at
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			WHERE c.lawyer_id = ?
			  AND c.created_at BETWEEN DATE_SUB(NOW(), INTERVAL ? YEAR) AND NOW()
			  AND c.deleted_at IS NULL
			ORDER BY c.created_at DESC
			LIMIT 50
		`

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, memberID, request.SearchYears).Rows()
		if err != nil {
			log.Printf("⚠️ 查询团队成员代理历史失败: %v", err)
			continue
		}

		for rows.Next() {
			var caseID uint
			var caseName, caseType, description, clientName string
			var historyClientID uint
			var createdAt, completedAt time.Time

			if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &historyClientID, &createdAt, &completedAt); err != nil {
				log.Printf("⚠️ 扫描时间序列冲突数据失败: %v", err)
				continue
			}

			// 检查冷却期
			coolingPeriodRequired := s.calculateCoolingPeriodRequired(caseType, completedAt)
			daysSinceCompletion := int(time.Since(completedAt).Hours() / 24)

			if daysSinceCompletion < coolingPeriodRequired {
				conflict := &models.MultiDimensionalConflict{
					ID:                     s.generateConflictID(),
					ConflictTitle:          fmt.Sprintf("冷却期冲突: %s - %s", member, caseName),
					ConflictDescription:    fmt.Sprintf("团队成员 '%s' 的案件 '%s' 尚在 %d 天冷却期内（要求 %d 天）", member, caseName, daysSinceCompletion, coolingPeriodRequired),
					ConflictType:           "TEMPORAL",
					ConflictSubtype:        "COOLING_PERIOD",
					SourceEntityType:       "LAWYER",
					SourceEntityID:         member,
					SourceEntityName:       member,
					RelatedEntityID:         fmt.Sprintf("%d", historyClientID),
					RelatedEntityName:       clientName,
					ConflictTimeDimension:   "HISTORICAL",
					ConflictDate:            completedAt,
					CoolingPeriodDays:       coolingPeriodRequired,
					BusinessDimension:       "LAWYER_RELATED",
					PracticeArea:           caseType,
					DetectionMethod:         "RULE_BASED",
					RecommendedActions:      []string{"等待冷却期结束", "评估代理风险", "考虑使用信息屏障"},
					SupportingEvidence:       []string{"案件完成记录", "冷却期规定"},
					RiskFactors:             []_array{"冷却期限制", "信息保密义务", "职业责任风险"},
				}

				// 根据冷却期剩余时间确定风险等级
				if daysSinceCompletion < coolingPeriodRequired/4 {
					conflict.RiskLevel = "HIGH"
				} else {
					conflict.RiskLevel = "MEDIUM"
				}

				conflicts = append(conflicts, conflict)
			}
		}
		rows.Close()
	}

	return conflicts, nil
}

// detectFinancialConflicts 检测财务利益冲突
func (s *enhancedConflictDetectionService) detectFinancialConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
) ([]*models.MultiDimensionalConflict, error) {
	var conflicts []*models.MultiDimensionalConflict

	// 解析团队成员
	var teamMembers []string
	if err := json.Unmarshal(request.TeamMembers, &teamMembers); err != nil {
		log.Printf("⚠️ 解析团队成员列表失败: %v", err)
		return conflicts, nil
	}

	// 检查律师与客户的财务关系
	for _, member := range teamMembers {
		// 这里需要实现财务关系检测逻辑
		// 由于需要外部数据源，这里提供框架

		// 检查直接财务利益
		directConflicts := s.detectDirectFinancialConflicts(ctx, member, request.ClientID)
		conflicts = append(conflicts, directConflicts...)

		// 检查间接财务利益
		indirectConflicts := s.detectIndirectFinancialConflicts(ctx, member, request.ClientID)
		conflicts = append(conflicts, indirectConflicts...)
	}

	return conflicts, nil
}

// detectProfessionalConflicts 检测专业领域特定冲突
func (s *enhancedConflictDetectionService) detectProfessionalConflicts(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
) ([]*models.MultiDimensionalConflict, error) {
	var conflicts []*models.MultiDimensionalConflict

	switch request.PracticeArea {
	case "INTELLECTUAL_PROPERTY":
		ipConflicts := s.detectIntellectualPropertyConflicts(ctx, request)
		conflicts = append(conflicts, ipConflicts...)

	case "SECURITIES":
		securitiesConflicts := s.detectSecuritiesConflicts(ctx, request)
		conflicts = append(conflicts, securitiesConflicts...)

	case "MERGERS_ACQUISITIONS":
		maConflicts := s.detectMergersAcquisitionsConflicts(ctx, request)
		conflicts = append(conflicts, maConflicts...)

	case "LITIGATION":
		litigationConflicts := s.detectLitigationConflicts(ctx, request)
		conflicts = append(conflicts, litigationConflicts...)

	case "FAMILY":
		familyConflicts := s.detectFamilyLawConflicts(ctx, request)
		conflicts = append(conflicts, familyConflicts...)
	}

	return conflicts, nil
}

// 辅助方法

// generateConflictID 生成冲突ID
func (s *enhancedConflictDetectionService) generateConflictID() string {
	return fmt.Sprintf("CONFLICT_%d", time.Now().UnixNano())
}

// parseLawyerID 解析律师ID
func (s *enhancedConflictDetectionService) parseLawyerID(lawyerIDStr string) (uint, error) {
	id, err := strconv.ParseUint(lawyerIDStr, 10, 32)
	return uint(id), err
}

// parseClientID 解析客户ID
func (s *enhancedConflictService) parseClientID(clientIDStr string) (uint, error) {
	id, err := strconv.ParseUint(clientIDStr, 10, 32)
	return uint(id), err
}

// isRelatedToOpposingParty 检查是否与对方当事人相关
func (s *enhancedConflictDetectionService) isRelatedToOpposingParty(caseName, description, clientName, opponent string) bool {
	caseName = strings.ToLower(caseName)
	description = strings.ToLower(description)
	clientName = strings.ToLower(clientName)
	opponent = strings.ToLower(opponent)

	// 检查完全匹配
	if strings.Contains(caseName, opponent) || strings.Contains(opponent, caseName) ||
		strings.Contains(description, opponent) || strings.Contains(clientName, opponent) {
		return true
	}

	// 检查变体匹配
	return false // 可以在这里添加更复杂的匹配逻辑
}

// isRelatedEntity 检查是否为关联实体
func (s *enhancedConflictDetectionService) isRelatedEntity(entity1, entity2 string) bool {
	entity1 = strings.ToLower(strings.TrimSpace(entity1))
	entity2 = strings.ToLower(strings.TrimSpace(entity2))

	// 检查常见公司变体
	companies := map[string][]string{
		"阿里巴巴": {"阿里", "淘宝", "天猫", "支付宝", "蚂蚁金服"},
		"腾讯":     {"微信", "qq", "财付通", "腾讯云", "腾讯游戏"},
		"字节跳动": {"抖音", "tiktok", "今日头条", "西瓜视频"},
	}

	for company, variants := range companies {
		if strings.Contains(entity1, company) || strings.Contains(entity2, company) {
			for _, variant := range variants {
				if strings.Contains(entity1, variant) || strings.Contains(entity2, variant) {
					return true
				}
			}
		}
	}

	return entity1 == entity2
}

// determineRelationshipRiskLevel 确定关系风险等级
func (s *enhancedConflictDetectionService) determineRelationshipRiskLevel(relationship *models.ClientRelationship) string {
	switch relationship.RelationshipType {
	case "PARENT", "SUBSIDIARY":
		return "HIGH"
	case "COMPETITOR", "ADVERSE":
		return "HIGH"
	case "SISTER", "AFFILIATE":
		return "MEDIUM"
	case "SUPPLIER", "CUSTOMER":
		return "LOW"
	default:
		return "MEDIUM"
	}
}

// calculateTimeRelevance 计算时间相关性
func (s *enhancedConflictDetectionService) calculateTimeRelevance(date time.Time) string {
	days := int(time.Since(date).Hours() / 24)

	switch {
	case days < 30:
		return "RECENT"
	case days < 90:
		return "MEDIUM_TERM"
	case days < 365:
		return "LONG_TERM"
	default:
		return "HISTORICAL"
	}
}

// calculateLawyerHistoryRiskLevel 计算律师历史风险等级
func (s *enhancedConflictDetectionService) calculateLawyerHistoryRiskLevel(conflict *models.MultiDimensionalConflict, request *models.ProfessionalConflictCheckRequest) string {
	// 基础风险等级
	baseRisk := "MEDIUM"

	// 根据时间相关性调整
	if conflict.TimeRelevance == "RECENT" {
		baseRisk = "HIGH"
	} else if conflict.TimeRelevance == "MEDIUM_TERM" {
		baseRisk = "MEDIUM"
	} else {
		baseRisk = "LOW"
	}

	// 根据案件性质调整
	if strings.Contains(conflict.PracticeArea, "CRIMINAL") ||
		strings.Contains(conflict.PracticeArea, "FAMILY") {
		if baseRisk != "LOW" {
			baseRisk = "HIGH"
		}
	}

	return baseRisk
}

// calculateDetectionConfidence 计算检测置信度
func (s *enhancedConflictDetectionService) calculateDetectionConfidence(conflict *models.MultiDimensionalConflict) float64 {
	confidence := 0.80 // 基础置信度

	// 根据检测方法调整
	switch conflict.DetectionMethod {
	case "RULE_BASED":
		confidence += 0.10
	case "MACHINE_LEARNING":
		confidence += 0.15
	case "HUMAN_JUDGMENT":
		confidence += 0.05
	}

	// 根据证据充分性调整
	if len(conflict.SupportingEvidence) > 2 {
		confidence += 0.10
	}

	// 确保不超过1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateRiskScore 计算风险评分
func (s *enhancedConflictDetectionService) calculateRiskScore(conflict *models.MultiDimensionalConflict) float64 {
	score := 50.0 // 基础评分

	// 根据风险等级调整
	switch conflict.RiskLevel {
	case "CRITICAL":
		score += 40.0
	case "HIGH":
		score += 25.0
	case "MEDIUM":
		score += 10.0
	case "LOW":
		score += 0.0
	}

	// 根据时间相关性调整
	switch conflict.TimeRelevance {
	case "RECENT":
		score += 15.0
	case "MEDIUM_TERM":
		score += 10.0
	case "LONG_TERM":
		score += 5.0
	case "HISTORICAL":
		score += 0.0
	}

	// 根据冲突类型调整
	switch conflict.ConflictType {
	case "DIRECT_OPPOSITION":
		score += 20.0
	case "FINANCIAL":
		score += 15.0
	case "CONCURRENT":
		score += 10.0
	case "RELATIONSHIP":
		score += 5.0
	}

	// 确保不超过100
	if score > 100.0 {
		score = 100.0
	}

	return score
}

// sortConflictsByRiskScore 按风险评分排序冲突
func (enhancedConflictDetectionService) sortConflictsByRiskScore(conflicts []*models.MultiDimensionalConflict) {
	// 简单的冒泡排序，实际生产环境建议使用更高效的排序算法
	for i := 0; i < len(conflicts); i++ {
		for j := i + 1; j < len(conflicts); j++ {
			if conflicts[i].RiskScore < conflicts[j].RiskScore {
				conflicts[i], conflicts[j] = conflicts[j], conflicts[i]
			}
		}
	}
}

// deduplicateConflicts 去重冲突记录
func (enhancedConflictDetectionService) deduplicateConflicts(conflicts []*models.MultiDimensionalConflict) []*models.MultiDimensionalConflict {
	seen := make(map[string]bool)
	var result []*models.MultiDimensionalConflict

	for _, conflict := conflicts {
		key := fmt.Sprintf("%s_%s_%s", conflict.SourceEntityName, conflict.RelatedEntityName, conflict.ConflictType)
		if !seen[key] {
			seen[key] = true
			result = append(result, conflict)
		}
	}

	return result
}

// generateRecommendations 生成处理建议
func (s *enhancedConflictDetectionService) generateRecommendations(
	ctx context.Context,
	classifiedConflicts []*models.ClassifiedConflict,
	riskAssessment *models.ProfessionalRiskAssessment,
	request *models.ProfessionalConflictCheckRequest,
) []string {
	var recommendations []string

	if len(classifiedConflicts) == 0 {
		recommendations = append(recommendations, "未发现明显利益冲突，可以继续代理")
		return recommendations
	}

	// 根据风险等级生成建议
	switch riskAssessment.OverallRisk {
	case "CRITICAL":
		recommendations = append(recommendations, "检测到严重利益冲突，强烈建议拒绝代理")
		recommendations = append(recommendations, "如必须代理，需要寻求伦理委员会审批")

	case "HIGH":
		recommendations = append(recommendations, "检测到高风险利益冲突，需要谨慎处理")
		recommendations = append(recommendations, "建议获取客户知情同意并实施信息屏障")

	case "MEDIUM":
		recommendations = append(recommendations, "检测到中等风险利益冲突，需要加强监控")
		recommendations = append(recommendations, "建议定期重新评估冲突状况")

	case "LOW":
		recommendations = append(recommendations, "检测到低风险利益冲突，可以正常代理")
		recommendations = append(recommendations, "建议保持必要的监控措施")
	}

	// 根据冲突类型生成具体建议
	conflictTypes := make(map[string]bool)
	for _, conflict := range classifiedConflicts {
		conflictTypes[conflict.ConflictType] = true
	}

	if conflictTypes["DIRECT_OPPOSITION"] {
		recommendations = append(recommendations, "存在直接对立关系，建议考虑更换律师或拒绝代理")
	}

	if conflictTypes["FINANCIAL"] {
		recommendations = append(recommendations, "存在财务利益冲突，需要完全披露并考虑回避")
	}

	if conflictTypes["TEMPORAL"] {
		recommendations = append(recommendations, "存在时间相关冲突，需要检查冷却期要求")
	}

	return recommendations
}

// 保存专业检查记录
func (s *enhancedConflictDetectionService) saveProfessionalCheckRecord(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
	response *models.ProfessionalConflictCheckResponse,
) error {
	record := &models.ProfessionalConflictCheckRecord{
		ID:                    response.RequestID,
		CheckNumber:          response.CheckNumber,
		ClientID:              request.ClientID,
		ClientName:            request.ClientName,
		CaseName:              request.CaseName,
		CaseType:              request.CaseType,
		PrimaryLawyer:        request.PrimaryLawyerName,
		TeamMembers:           request.TeamMembers,
		OpposingParties:       request.OpposingParties,
		SearchDepth:           request.SearchDepth,
		SearchYears:           request.SearchYears,
		Status:                response.CheckStatus,
		HasConflict:            response.HasConflict,
		ConflictCount:         len(response.Conflicts),
		OverallRiskLevel:      response.RiskAssessment.OverallRisk,
		ProcessingDuration:     response.ProcessingDuration,
		GeneratedAt:           response.GeneratedAt,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	return s.conflictRepo.SaveProfessionalCheckRecord(ctx, record)
}

// ValidateConflictCheckRequest 验证冲突检查请求
func (s *enhancedConflictDetectionService) ValidateConflictCheckRequest(
	ctx context.Context,
	request *models.ProfessionalConflictCheckRequest,
) error {
	if request.ClientID == "" {
		return fmt.Errorf("客户ID不能为空")
	}

	if request.ClientName == "" {
		return fmt.Errorf("客户名称不能为空")
	}

	if request.CaseName == "" {
		return fmt.Errorf("案件名称不能为空")
	}

	if request.CaseType == "" {
		return fmt.Errorf("案件类型不能为空")
	}

	if request.PrimaryLawyerID == "" {
		return fmt.Errorf("主办律师ID不能为空")
	}

	if request.SearchYears <= 0 {
		request.SearchYears = s.config.DefaultSearchYears
	}

	if request.SearchDepth == "" {
		request.SearchDepth = s.config.DefaultSearchDepth
	}

	return nil
}

// 以下是一些占位符方法，实际实现需要根据具体需求完善

func (s *enhancedConflictDetectionService) detectDirectFinancialConflicts(ctx context.Context, lawyerID string, clientID string) []*models.MultiDimensionalConflict {
	// TODO: 实现直接财务利益冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectIndirectFinancialConflicts(ctx context.Context, lawyerID string, clientID string) []*models.MultiDimensionalConflict {
	// TODO: 实现间接财务利益冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectIntellectualPropertyConflicts(ctx context.Context, request *models.ProfessionalConflictCheckRequest) []*models.MultiDimensionalConflict {
	// TODO: 实现知识产权冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectSecuritiesConflicts(ctx context.Context, request *models.ProfessionalConflictCheckRequest) []*models.MultiDimensionalConflict {
	// TODO: 实现证券冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectMergersAcquisitionsConflicts(ctx context.Context, request *models.ProfessionalConflictCheckRequest) []*models.MultiDimensionalConflict {
	// TODO: 实现并购冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectLitigationConflicts(ctx context.Context, request *models.ProfessionalConflictCheckRequest) []*models.MultiDimensionalConflict {
	// TODO: 实现诉讼冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectFamilyLawConflicts(ctx context.Context, request *models.ProfessionalConflictCheckRequest) []*models.MultiDimensionalConflict {
	// TODO: 实现家事法冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectRelationshipHistoryConflicts(ctx context.Context, relationship *models.ClientRelationship, request *models.ProfessionalConflictCheckRequest) []*models.MultiDimensionalConflict {
	// TODO: 实现关联历史冲突检测
	return []*models.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) detectBusinessRelationshipConflicts(ctx context.Context, request *models.ProfessionalConflictCheckRequest, opponent string) []*models.MultiDimensionalConflict {
	// TODO: 实现商业关系冲突检测
	return []*.MultiDimensionalConflict{}
}

func (s *enhancedConflictDetectionService) calculateCoolingPeriodRequired(caseType string, completionDate time.Time) int {
	// 根据案件类型确定冷却期（天）
	switch caseType {
	case "CRIMINAL":
		return 365
	case "FAMILY":
		return 180
	case "INTELLECTUAL_PROPERTY":
		return 90
	default:
		return 365
	}
}