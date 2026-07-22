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
	"law-oa-go/internal/security"
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

	// P0 uses the complete system archive. The request field remains in the
	// audit payload, but a caller may not narrow the mandatory search window.
	request.SearchYears = 0
	// The formal P0 check always includes structured relations. A browser or
	// integration caller must not turn alias/related-party coverage off by
	// sending a legacy BASIC/false payload.
	request.IncludeCorporateRelations = true
	if strings.EqualFold(strings.TrimSpace(request.SearchDepth), "BASIC") || strings.TrimSpace(request.SearchDepth) == "" {
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

	// P0 deterministic policy: exact adverse-party/client matches block intake;
	// all candidate, relation and text hits require human review and cannot auto-escalate.
	decision, riskAssessment := s.applyP0ConflictPolicy(ctx, request, conflictCases)

	// 生成建议
	recommendations, err := s.riskAssessor.GenerateRecommendations(ctx, riskAssessment.OverallRisk, conflictCases)
	if err != nil {
		log.Printf("❌ 生成建议失败: %v", err)
		return nil, fmt.Errorf("生成建议失败: %w", err)
	}

	// 构建响应
	response := &models.ConflictCheckResponse{
		CheckID:            checkID,
		HasConflict:        decision.Status == "BLOCKED" || len(conflictCases) > 0 || decision.CoverageStatus != "COMPLETE",
		ConflictCases:      conflictCases,
		CheckStatistics:    s.buildCheckStatistics(ctx, request, conflictCases),
		RiskAssessment:     riskAssessment,
		Recommendations:    recommendations,
		CheckTime:          time.Now(),
		Duration:           time.Since(startTime).Milliseconds(),
		NormalizedSubjects: normalizeConflictSubjects(request),
		Decision:           decision,
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

	// Production P0 reads the immutable firm-wide subject index. The legacy
	// direct-query branches below remain available only when that index is not
	// deployed, which is useful for isolated compatibility tests but is blocked
	// by production readiness before the server can accept traffic.
	if indexer, ok := s.conflictRepo.(repositories.ConflictP0SubjectIndexer); ok && conflictP0SubjectIndexAvailable(s.caseRepo) {
		if err := indexer.SyncConflictSubjectIndex(ctx); err != nil {
			return nil, fmt.Errorf("同步全所冲突主体索引失败: %w", err)
		}
		subjects := normalizeConflictSubjects(request)
		hits, err := indexer.SearchConflictSubjectIndex(ctx, subjects, request.ClientID)
		if err != nil {
			return nil, fmt.Errorf("查询全所冲突主体索引失败: %w", err)
		}
		return s.buildIndexedConflictCases(hits), nil
	}

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
		return nil, fmt.Errorf("查找律所历史案件冲突失败: %w", err)
	}
	allConflicts = append(allConflicts, lawyerConflicts...)
	log.Printf("📋 找到 %d 个律所历史案件冲突案例", len(lawyerConflicts))

	// 2. 检查对方当事人冲突
	if len(request.OtherParties) > 0 {
		opponentConflicts, err := s.checkOpponentConflicts(ctx, request, since)
		if err != nil {
			// 对方当事人查询失败 = 冲突检查整体失败，禁止落库为"无冲突"成功记录
			log.Printf("❌ 对方当事人冲突检查失败: %v", err)
			return nil, fmt.Errorf("对方当事人冲突检查失败: %w", err)
		}
		allConflicts = append(allConflicts, opponentConflicts...)
		log.Printf("📋 找到 %d 个对方当事人冲突案例", len(opponentConflicts))
	}

	// Structured party/entity records are authoritative when present. A text
	// search alone cannot find aliases, historical names, or identity matches.
	// A missing structured source is a hard error so it cannot silently become a
	// clean result in a production database that claims to support P0 checks.
	structuredConflicts, err := s.checkStructuredPartyConflicts(ctx, request, since)
	if err != nil {
		return nil, fmt.Errorf("结构化主体冲突检查失败: %w", err)
	}
	allConflicts = append(allConflicts, structuredConflicts...)

	// 3. 检查客户关系冲突
	if request.IncludeCorporateRelations && request.SearchDepth != "BASIC" {
		relationConflicts, relationErr := s.checkClientRelationConflictsStrict(ctx, request, since)
		if relationErr != nil {
			return nil, fmt.Errorf("客户关联关系冲突检查失败: %w", relationErr)
		}
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
// 返回错误时调用方必须把整个冲突检查标记为失败——禁止吞掉错误伪造"无冲突"结果。
func (s *conflictDetectionService) checkOpponentConflicts(ctx context.Context, request *models.ConflictCheckRequest, since time.Time) ([]*models.ConflictCase, error) {
	var conflicts []*models.ConflictCase

	for _, opponent := range request.OtherParties {
		if strings.TrimSpace(opponent) == "" {
			continue
		}

		sinceFilter := ""
		args := []interface{}{
			"%" + opponent + "%",
			"%" + opponent + "%",
			"%" + opponent + "%",
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
					WHEN LOWER(c.title) LIKE LOWER(?) THEN 3
					WHEN LOWER(c.description) LIKE LOWER(?) THEN 2
					WHEN LOWER(cl.name) LIKE LOWER(?) THEN 1
					ELSE 0
				END as relevance_score
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			JOIN users u ON c.lawyer_id = u.id
			WHERE c.deleted_at IS NULL
			%s
			AND (LOWER(c.title) LIKE LOWER(?) OR LOWER(c.description) LIKE LOWER(?) OR LOWER(cl.name) LIKE LOWER(?))
			ORDER BY relevance_score DESC, c.created_at DESC
		`, sinceFilter)

		// 跨库通用大小写不敏感匹配：LOWER(col) LIKE LOWER(?)。
		// PostgreSQL 专属的 ILIKE 在 MySQL/SQLite 下直接报语法错误，错误被 continue 吞掉 → 全表漏报。
		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, args...).Rows()
		if err != nil {
			return nil, fmt.Errorf("查询对方当事人冲突失败 (opponent=%q): %w", opponent, err)
		}

		for rows.Next() {
			var caseID uint
			var caseNo, caseName, caseType, description, clientName, lawyerName string
			var lawyerID uint
			var createdAt time.Time
			var relevanceScore int

			if err := rows.Scan(&caseID, &caseNo, &caseName, &caseType, &description, &clientName, &lawyerName, &createdAt, &lawyerID, &relevanceScore); err != nil {
				rows.Close()
				return nil, fmt.Errorf("扫描对方当事人冲突行失败 (opponent=%q): %w", opponent, err)
			}

			// 非结构化文本和名称相似只能产生待核实候选；仅规范化名称完全一致可判直接冲突。
			conflictType := "文本提及待核实"
			riskLevel := "MEDIUM"
			kind := s.classifyPartyMatch(clientName, opponent)
			switch kind {
			case PartyExactNormalizedMatch:
				conflictType = "对方当事人直接冲突"
				riskLevel = "CRITICAL"
			case PartyCandidateMatch:
				conflictType = "名称相似待核实"
			}

			conflict := &models.ConflictCase{
				ID:              fmt.Sprintf("opponent_%d", caseID),
				CaseID:          fmt.Sprintf("%d", caseID),
				CaseName:        caseName,
				CaseNo:          caseNo,
				CaseType:        caseType,
				ConflictType:    conflictType,
				RiskLevel:       riskLevel,
				Description:     fmt.Sprintf("对方名称 %q 出现在历史案件 %q 的标题、摘要或客户名称中，需核实其当事人角色", opponent, caseName),
				CaseStatus:      "active",
				ClientID:        request.ClientID,
				OpposingParties: []string{opponent},
				ConflictDetails: "非结构化文本候选命中，不足以单独认定利益冲突",
				CreatedAt:       createdAt,
			}
			conflict.MatchType = "TEXT"
			conflict.RuleCode = "TEXT_MENTION_REVIEW"
			conflict.RequiresManualReview = true
			if kind == PartyExactNormalizedMatch {
				conflict.MatchType = "EXACT"
				conflict.RuleCode = "DIRECT_ADVERSE_CURRENT_CLIENT"
				conflict.Description = fmt.Sprintf("当前对方当事人 '%s' 与本所历史客户 '%s' 规范化后完全一致，存在直接利益冲突", opponent, clientName)
				conflict.ConflictDetails = "对方当事人与本所历史客户直接命中（Exact 规范化匹配）"
			} else if kind == PartyCandidateMatch {
				conflict.MatchType = "CANDIDATE"
				conflict.RuleCode = "SUBJECT_CANDIDATE_REVIEW"
				conflict.Description = fmt.Sprintf("对方当事人 '%s' 与本所历史客户 '%s' 名称相关，需人工复核是否构成利益冲突", opponent, clientName)
				conflict.ConflictDetails = "对方当事人与本所历史客户候选匹配（Candidate），禁止自动判定 CRITICAL"
			}
			conflict.Evidence = []models.ConflictEvidence{{
				EvidenceID: fmt.Sprintf("EV-CASE-%d-%s", caseID, conflict.MatchType), RuleCode: conflict.RuleCode,
				MatchType: conflict.MatchType, SourceType: "INTERNAL_CASE", RequestedParty: opponent,
				MatchedEntity: clientName, PartyRole: "OPPOSING_PARTY", HistoricalRole: "CLIENT",
				SourceCaseID: fmt.Sprintf("%d", caseID), SourceCaseNumber: caseNo, SourceCaseName: caseName,
				LawyerName: lawyerName, SourceUpdatedAt: createdAt, Summary: conflict.Description,
			}}

			conflicts = append(conflicts, conflict)
		}
		// rows.Err() 报告迭代期间发生的错误；Rows()+Next() 不会主动返回
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("迭代对方当事人冲突结果集失败 (opponent=%q): %w", opponent, err)
		}
		rows.Close()
	}

	return conflicts, nil
}

type structuredConflictSubject struct {
	Name        string
	Role        string
	EntityType  string
	Identifiers map[string]string
	Aliases     []string
}

// checkStructuredPartyConflicts searches the normalized entity/party index in
// addition to free text. The index is the only place where aliases, former
// names, and strong identity identifiers can be searched consistently across
// the whole firm archive.
func (s *conflictDetectionService) checkStructuredPartyConflicts(ctx context.Context, request *models.ConflictCheckRequest, since time.Time) ([]*models.ConflictCase, error) {
	if request == nil || s.caseRepo == nil || s.caseRepo.GetDB() == nil {
		return nil, fmt.Errorf("结构化主体检索数据库未初始化")
	}
	subjects := make([]structuredConflictSubject, 0, len(request.Parties)+len(request.OtherParties)+1)
	seen := map[string]struct{}{}
	if name := strings.TrimSpace(request.ClientName); name != "" {
		key := strings.ToLower(name) + "|CLIENT"
		seen[key] = struct{}{}
		subjects = append(subjects, structuredConflictSubject{
			Name:        name,
			Role:        conflictSubjectRoleClient,
			EntityType:  "ANY",
			Identifiers: request.ClientIdentifiers,
			Aliases:     request.ClientAliases,
		})
	}
	for _, party := range request.Parties {
		name := strings.TrimSpace(party.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name) + "|" + strings.ToUpper(strings.TrimSpace(party.Role))
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		subjects = append(subjects, structuredConflictSubject{
			Name: name, Role: normalizeConflictSubjectRole(party.Role),
			EntityType:  normalizeConflictSubjectEntityType(party.EntityType),
			Identifiers: party.Identifiers, Aliases: party.Aliases,
		})
	}
	for _, name := range request.OtherParties {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name) + "|OPPOSING_PARTY"
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		subjects = append(subjects, structuredConflictSubject{Name: name, Role: conflictSubjectRoleOpposingParty, EntityType: "ANY"})
	}
	if len(subjects) == 0 {
		return nil, nil
	}

	db := s.caseRepo.GetDB().WithContext(ctx)
	for _, table := range []string{"entities", "case_parties", "entity_name_history", "entity_relations"} {
		if !db.Migrator().HasTable(table) {
			return nil, fmt.Errorf("结构化主体索引未部署（需要 entities、case_parties、entity_name_history 和 entity_relations）")
		}
	}
	var clientID uint
	if _, err := fmt.Sscanf(strings.TrimSpace(request.ClientID), "%d", &clientID); err != nil || clientID == 0 {
		return nil, fmt.Errorf("客户ID格式错误，无法执行结构化主体检索")
	}

	entityConditions, entityArgs, err := structuredSubjectConditions("e.name", "e.alias", "e", subjects)
	if err != nil {
		return nil, err
	}
	clientArchiveConditions, clientArchiveArgs, err := clientArchiveSubjectConditions(subjects)
	if err != nil {
		return nil, err
	}
	historyConditions, historyArgs, err := structuredSubjectConditions("enh.old_name", "enh.new_name", "e", subjects)
	if err != nil {
		return nil, err
	}
	relationConditions, relationArgs, err := structuredSubjectConditions("re.name", "re.alias", "re", subjects)
	if err != nil {
		return nil, err
	}
	if entityConditions == "" && clientArchiveConditions == "" && historyConditions == "" && relationConditions == "" {
		return nil, nil
	}
	branchArgs := func(conditionArgs []interface{}) []interface{} {
		args := []interface{}{clientID}
		args = append(args, conditionArgs...)
		if !since.IsZero() {
			args = append(args, since)
		}
		return args
	}
	dateFilter := ""
	if !since.IsZero() {
		dateFilter = " AND c.created_at >= ?"
	}
	selectColumns := `
		SELECT c.id, c.case_number, c.title, c.case_type, c.description,
		       c.client_id, COALESCE(cl.name, ''), COALESCE(u.name, ''), c.created_at,
		       COALESCE(%s.name, ''), COALESCE(%s.alias, ''), COALESCE(%s.entity_type, ''),
		       COALESCE(%s.identity_type, ''), COALESCE(%s.identity_number_digest, ''), COALESCE(cp.role, ''),
		       COALESCE(%s, ''), '%s', '%s'
		FROM case_parties cp
		JOIN entities %s ON %s.id = cp.entity_id AND %s.deleted_at IS NULL
		JOIN cases c ON c.id = cp.case_id AND c.deleted_at IS NULL AND c.client_id <> ?
		LEFT JOIN clients cl ON cl.id = c.client_id
		LEFT JOIN users u ON u.id = c.lawyer_id
		WHERE cp.deleted_at IS NULL AND (%s)%s`
	branches := make([]string, 0, 3)
	args := make([]interface{}, 0)
	if entityConditions != "" {
		branches = append(branches, fmt.Sprintf(selectColumns, "e", "e", "e", "e", "e", "e.name", "ENTITY", "", "e", "e", "e", entityConditions, dateFilter))
		args = append(args, branchArgs(entityArgs)...)
	}
	if clientArchiveConditions != "" {
		branches = append(branches, fmt.Sprintf(`
		SELECT c.id, c.case_number, c.title, c.case_type, c.description,
		       c.client_id, COALESCE(cl.name, ''), COALESCE(u.name, ''), c.created_at,
		       COALESCE(cl.name, ''), COALESCE(cl.company, ''), COALESCE(cl.type, ''),
		       CASE WHEN COALESCE(cl.id_card_digest, '') <> '' THEN 'ID_CARD' ELSE '' END,
		       COALESCE(cl.id_card_digest, ''), 'CLIENT', COALESCE(cl.name, ''), 'CLIENT_ARCHIVE', ''
		FROM cases c
		JOIN clients cl ON cl.id = c.client_id AND cl.deleted_at IS NULL
		LEFT JOIN users u ON u.id = c.lawyer_id
		WHERE c.deleted_at IS NULL AND c.client_id <> ? AND (%s)%s`, clientArchiveConditions, dateFilter))
		args = append(args, branchArgs(clientArchiveArgs)...)
	}
	if historyConditions != "" {
		branches = append(branches, fmt.Sprintf(`
		SELECT c.id, c.case_number, c.title, c.case_type, c.description,
		       c.client_id, COALESCE(cl.name, ''), COALESCE(u.name, ''), c.created_at,
		       COALESCE(e.name, ''), COALESCE(e.alias, ''), COALESCE(e.entity_type, ''),
		       COALESCE(e.identity_type, ''), COALESCE(e.identity_number_digest, ''), COALESCE(cp.role, ''),
		       COALESCE(enh.old_name, enh.new_name), 'FORMER_NAME', ''
		FROM case_parties cp
		JOIN entities e ON e.id = cp.entity_id AND e.deleted_at IS NULL
		JOIN entity_name_history enh ON enh.entity_id = e.id AND enh.deleted_at IS NULL
		JOIN cases c ON c.id = cp.case_id AND c.deleted_at IS NULL AND c.client_id <> ?
		LEFT JOIN clients cl ON cl.id = c.client_id
		LEFT JOIN users u ON u.id = c.lawyer_id
		WHERE cp.deleted_at IS NULL AND (%s)%s`, historyConditions, dateFilter))
		args = append(args, branchArgs(historyArgs)...)
	}
	if relationConditions != "" {
		branches = append(branches, fmt.Sprintf(`
		SELECT c.id, c.case_number, c.title, c.case_type, c.description,
		       c.client_id, COALESCE(cl.name, ''), COALESCE(u.name, ''), c.created_at,
		       COALESCE(re.name, ''), COALESCE(re.alias, ''), COALESCE(re.entity_type, ''),
		       COALESCE(re.identity_type, ''), COALESCE(re.identity_number_digest, ''), COALESCE(cp.role, ''),
		       COALESCE(re.name, ''), 'RELATED_ENTITY', COALESCE(er.relation_type, '')
		FROM case_parties cp
		JOIN entities e ON e.id = cp.entity_id AND e.deleted_at IS NULL
		JOIN entity_relations er ON er.deleted_at IS NULL AND er.is_active = TRUE AND
		       (er.source_entity_id = e.id OR er.target_entity_id = e.id)
		JOIN entities re ON re.id = CASE WHEN er.source_entity_id = e.id THEN er.target_entity_id ELSE er.source_entity_id END
		       AND re.deleted_at IS NULL
		JOIN cases c ON c.id = cp.case_id AND c.deleted_at IS NULL AND c.client_id <> ?
		LEFT JOIN clients cl ON cl.id = c.client_id
		LEFT JOIN users u ON u.id = c.lawyer_id
		WHERE cp.deleted_at IS NULL AND (%s)%s`, relationConditions, dateFilter))
		args = append(args, branchArgs(relationArgs)...)
	}
	query := strings.Join(branches, " UNION ALL ") + " ORDER BY 9 DESC"

	rows, err := db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	conflicts := make([]*models.ConflictCase, 0)
	for rows.Next() {
		var caseID, historicalClientID uint
		var caseNo, caseName, caseType, description, clientName, lawyerName string
		var entityName, entityAlias, entityType, identityType, identityDigest, partyRole string
		var matchedName, matchSource, relationType string
		var createdAt time.Time
		if err := rows.Scan(&caseID, &caseNo, &caseName, &caseType, &description, &historicalClientID, &clientName, &lawyerName, &createdAt, &entityName, &entityAlias, &entityType, &identityType, &identityDigest, &partyRole, &matchedName, &matchSource, &relationType); err != nil {
			return nil, err
		}
		kind, matched, ruleCode, matchType := strongestStructuredMatch(s, entityName, entityAlias, entityType, identityType, identityDigest, matchedName, matchSource, subjects)
		if kind == PartyNoMatch {
			continue
		}
		riskLevel := "MEDIUM"
		conflictType := "结构化主体候选待核实"
		details := "结构化当事人记录命中，需独立核查主体身份、历史角色和保密信息接触范围"
		if ruleCode == "STRUCTURED_IDENTITY_EXACT" {
			conflictType = "主体身份标识完全匹配待复核"
			riskLevel = "HIGH"
			details = "结构化主体身份标识完全匹配，但仍须由独立核查人形成最终处置结论"
		} else if ruleCode == "CLIENT_ARCHIVE_NAME_CANDIDATE" {
			conflictType = "历史客户主体命中待复核"
			details = "历史客户主档名称或公司字段命中，需独立核查是否为同一主体及历史保密义务范围"
		} else if ruleCode == "RELATED_ENTITY_ADVERSE_REVIEW" {
			conflictType = "关联实体命中待复核"
			details = "历史案件当事人的关联实体命中，需核实关联方向、控制范围和历史保密信息接触范围"
		} else if ruleCode == "PERSON_NAME_ONLY_INSUFFICIENT" {
			conflictType = "自然人同名信息不足"
			details = "仅凭自然人姓名不足以认定为同一主体，请补充证件标识后复核"
		}
		conflict := &models.ConflictCase{
			ID: fmt.Sprintf("structured_%d_%s", caseID, matchType), CaseID: fmt.Sprint(caseID), CaseName: caseName, CaseNo: caseNo,
			CaseType: caseType, ConflictType: conflictType, RiskLevel: riskLevel,
			Description: fmt.Sprintf("当前主体 %q 命中历史案件中的结构化主体 %q，需独立核查", matched, matchedName),
			CaseStatus:  "active", ClientID: fmt.Sprint(historicalClientID), OpposingParties: []string{matched},
			ConflictDetails: details, CreatedAt: createdAt, MatchType: matchType, RuleCode: ruleCode, RequiresManualReview: true,
		}
		if relationType != "" {
			conflict.ConflictDetails = fmt.Sprintf("%s；关系类型：%s", conflict.ConflictDetails, relationType)
		}
		if matchSource == "RELATED_ENTITY" {
			conflict.MatchType = "RELATION"
			conflict.RuleCode = "RELATED_ENTITY_ADVERSE_REVIEW"
		} else if matchSource == "CLIENT_ARCHIVE" && conflict.RuleCode != "STRUCTURED_IDENTITY_EXACT" && conflict.RuleCode != "PERSON_NAME_ONLY_INSUFFICIENT" {
			conflict.RuleCode = "CLIENT_ARCHIVE_NAME_CANDIDATE"
		}
		conflict.Evidence = []models.ConflictEvidence{{
			EvidenceID: fmt.Sprintf("EV-STRUCTURED-%d-%s", caseID, matchType), RuleCode: conflict.RuleCode, MatchType: conflict.MatchType,
			SourceType: structuredEvidenceSource(matchSource), RequestedParty: matched, MatchedEntity: matchedName, PartyRole: structuredEvidencePartyRole(matchSource, partyRole),
			HistoricalRole: structuredEvidenceHistoricalRole(matchSource), SourceCaseID: fmt.Sprint(caseID), SourceCaseNumber: caseNo, SourceCaseName: caseName,
			LawyerName: lawyerName, SourceUpdatedAt: createdAt, Summary: details,
		}}
		conflicts = append(conflicts, conflict)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return conflicts, nil
}

func canonicalIdentityType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "id_card", "idcard", "身份证", "身份证号":
		return "ID_CARD"
	case "unified_social_credit_code", "social_credit_code", "统一社会信用代码":
		return "SOCIAL_CREDIT_CODE"
	case "business_license", "营业执照":
		return "BUSINESS_LICENSE"
	default:
		return strings.ToUpper(strings.TrimSpace(value))
	}
}

func structuredSubjectConditions(nameExpr, aliasExpr, identityAlias string, subjects []structuredConflictSubject) (string, []interface{}, error) {
	conditions := make([]string, 0, len(subjects)*3)
	args := make([]interface{}, 0, len(subjects)*3)
	for _, subject := range subjects {
		appendName := func(name string) {
			name = strings.TrimSpace(name)
			if name == "" {
				return
			}
			conditions = append(conditions, fmt.Sprintf("(LOWER(COALESCE(%s, '')) LIKE LOWER(?) OR LOWER(COALESCE(%s, '')) LIKE LOWER(?))", nameExpr, aliasExpr))
			args = append(args, "%"+name+"%", "%"+name+"%")
		}
		appendName(subject.Name)
		for _, alias := range subject.Aliases {
			appendName(alias)
		}
		for identifierType, identifierValue := range subject.Identifiers {
			if strings.TrimSpace(identifierType) == "" || strings.TrimSpace(identifierValue) == "" {
				continue
			}
			digest, err := security.IdentityDigest(identifierValue)
			if err != nil {
				return "", nil, fmt.Errorf("结构化主体身份检索不可用: %w", err)
			}
			conditions = append(conditions, fmt.Sprintf("(LOWER(COALESCE(%s.identity_type, '')) = LOWER(?) AND %s.identity_number_digest = ?)", identityAlias, identityAlias))
			args = append(args, canonicalIdentityType(identifierType), digest)
		}
	}
	return strings.Join(conditions, " OR "), args, nil
}

func clientArchiveSubjectConditions(subjects []structuredConflictSubject) (string, []interface{}, error) {
	conditions := make([]string, 0, len(subjects)*3)
	args := make([]interface{}, 0, len(subjects)*3)
	for _, subject := range subjects {
		appendName := func(name string) {
			name = strings.TrimSpace(name)
			if name == "" {
				return
			}
			conditions = append(conditions, "(LOWER(COALESCE(cl.name, '')) LIKE LOWER(?) OR LOWER(COALESCE(cl.company, '')) LIKE LOWER(?))")
			args = append(args, "%"+name+"%", "%"+name+"%")
		}
		appendName(subject.Name)
		for _, alias := range subject.Aliases {
			appendName(alias)
		}
		for identifierType, identifierValue := range subject.Identifiers {
			if canonicalIdentityType(identifierType) != "ID_CARD" || strings.TrimSpace(identifierValue) == "" {
				continue
			}
			digest, err := security.IdentityDigest(identifierValue)
			if err != nil {
				return "", nil, fmt.Errorf("历史客户身份检索不可用: %w", err)
			}
			conditions = append(conditions, "cl.id_card_digest = ?")
			args = append(args, digest)
		}
	}
	return strings.Join(conditions, " OR "), args, nil
}

func strongestStructuredMatch(s *conflictDetectionService, entityName, entityAlias, entityType, identityType, identityDigest, matchedName, matchSource string, subjects []structuredConflictSubject) (PartyMatchKind, string, string, string) {
	strongest := PartyNoMatch
	matched, ruleCode, matchType := "", "", ""
	for _, subject := range subjects {
		for key, value := range subject.Identifiers {
			digest, _ := security.IdentityDigest(value)
			if canonicalIdentityType(key) == canonicalIdentityType(identityType) && strings.TrimSpace(digest) != "" && strings.EqualFold(strings.TrimSpace(digest), strings.TrimSpace(identityDigest)) {
				if matchSource == "RELATED_ENTITY" {
					return PartyExactNormalizedMatch, subject.Name, "RELATED_ENTITY_ADVERSE_REVIEW", "RELATION"
				}
				return PartyExactNormalizedMatch, subject.Name, "STRUCTURED_IDENTITY_EXACT", "EXACT"
			}
		}
		kind := PartyNoMatch
		for _, candidate := range []string{matchedName, entityName, entityAlias} {
			if strings.TrimSpace(candidate) == "" {
				continue
			}
			if candidateKind := s.classifyPartyMatch(candidate, subject.Name); candidateKind > kind {
				kind = candidateKind
			}
		}
		if kind > strongest {
			strongest, matched, matchType = kind, subject.Name, "CANDIDATE"
			ruleCode = "SUBJECT_CANDIDATE_REVIEW"
			if matchSource == "FORMER_NAME" {
				ruleCode = "FORMER_NAME_CANDIDATE_REVIEW"
			} else if matchSource == "RELATED_ENTITY" {
				matchType = "RELATION"
				ruleCode = "RELATED_ENTITY_ADVERSE_REVIEW"
			}
			if strings.EqualFold(normalizeConflictEntityType(entityType), "PERSON") && len(subject.Identifiers) == 0 {
				ruleCode = "PERSON_NAME_ONLY_INSUFFICIENT"
			}
		}
	}
	return strongest, matched, ruleCode, matchType
}

func structuredEvidenceSource(matchSource string) string {
	switch matchSource {
	case "FORMER_NAME":
		return "ENTITY_NAME_HISTORY"
	case "RELATED_ENTITY":
		return "CLIENT_RELATION"
	case "CLIENT_ARCHIVE":
		return "CLIENT_ARCHIVE"
	default:
		return "STRUCTURED_CASE_PARTY"
	}
}

func structuredEvidencePartyRole(matchSource, partyRole string) string {
	if matchSource == "RELATED_ENTITY" {
		return "RELATED_PARTY"
	}
	return normalizeConflictSubjectRole(partyRole)
}

func structuredEvidenceHistoricalRole(matchSource string) string {
	if matchSource == "RELATED_ENTITY" {
		return "RELATED_ENTITY"
	}
	if matchSource == "FORMER_NAME" {
		return "HISTORICAL_PARTY_FORMER_NAME"
	}
	if matchSource == "CLIENT_ARCHIVE" {
		return "HISTORICAL_CLIENT"
	}
	return "HISTORICAL_PARTY"
}

func normalizeConflictEntityType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "INDIVIDUAL", "PERSON", "自然人":
		return "PERSON"
	case "LEGAL_PERSON", "ORGANIZATION", "COMPANY", "法人", "企业":
		return "COMPANY"
	default:
		return "ANY"
	}
}

// checkClientRelationConflicts 检查客户关系冲突
func (s *conflictDetectionService) checkClientRelationConflicts(ctx context.Context, request *models.ConflictCheckRequest, since time.Time) []*models.ConflictCase {
	conflicts, _ := s.checkClientRelationConflictsStrict(ctx, request, since)
	return conflicts
}

func (s *conflictDetectionService) checkClientRelationConflictsStrict(ctx context.Context, request *models.ConflictCheckRequest, since time.Time) ([]*models.ConflictCase, error) {
	var conflicts []*models.ConflictCase

	// 获取客户关系
	relations, err := s.conflictRepo.GetClientRelations(ctx, request.ClientID)
	if err != nil {
		return nil, fmt.Errorf("获取客户关系失败: %w", err)
	}

	for _, relation := range relations {
		sinceFilter := ""
		args := []interface{}{relation.RelatedClientID}
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
			%s
			ORDER BY c.created_at DESC
			LIMIT 5
		`, sinceFilter)

		rows, err := s.caseRepo.GetDB().WithContext(ctx).Raw(query, args...).Rows()
		if err != nil {
			return nil, fmt.Errorf("查询关联客户历史案件失败: %w", err)
		}

		for rows.Next() {
			var caseID uint
			var caseNo, caseName, caseType, description, clientName, lawyerName string
			var createdAt time.Time

			if err := rows.Scan(&caseID, &caseNo, &caseName, &caseType, &description, &clientName, &lawyerName, &createdAt); err != nil {
				continue
			}
			matchKind, matchedParty := PartyNoMatch, ""
			for _, party := range request.OtherParties {
				candidate := s.classifyPartyMatch(clientName, party)
				if candidate > matchKind {
					matchKind, matchedParty = candidate, party
				}
			}
			if matchKind == PartyNoMatch {
				continue
			}

			conflict := &models.ConflictCase{
				ID:                   fmt.Sprintf("relation_%d", caseID),
				CaseID:               fmt.Sprintf("%d", caseID),
				CaseName:             caseName,
				CaseNo:               caseNo,
				CaseType:             caseType,
				ConflictType:         "关联实体对立待复核",
				RiskLevel:            "MEDIUM",
				Description:          fmt.Sprintf("当前对方 %q 与客户关系网络中的实体 %q 匹配，需核实关系方向、控制范围和历史保密信息", matchedParty, clientName),
				CaseStatus:           "active",
				ClientID:             request.ClientID,
				ConflictDetails:      fmt.Sprintf("客户关系: %s - %s", relation.RelationType, relation.RelationDetail),
				CreatedAt:            createdAt,
				MatchType:            "RELATION",
				RuleCode:             "RELATED_ENTITY_ADVERSE_REVIEW",
				RequiresManualReview: true,
			}
			conflict.OpposingParties = []string{matchedParty}
			conflict.Evidence = []models.ConflictEvidence{{
				EvidenceID: fmt.Sprintf("EV-REL-%d", caseID), RuleCode: conflict.RuleCode, MatchType: "RELATION",
				SourceType: "CLIENT_RELATION", RequestedParty: matchedParty, MatchedEntity: clientName,
				PartyRole: "OPPOSING_PARTY", HistoricalRole: "RELATED_CLIENT", SourceCaseID: fmt.Sprintf("%d", caseID),
				SourceCaseNumber: caseNo, SourceCaseName: caseName, SourceUpdatedAt: createdAt,
				Summary: conflict.Description,
			}}

			conflicts = append(conflicts, conflict)
		}
		rows.Close()
	}

	return conflicts, nil
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

// PartyMatchKind 当事人名称匹配强度三态分类。
//
// 设计底层逻辑：一个 bool 承载不了"完全相等 vs 包含/简称"两种语义，
// 后者会误把短子串升级为 CRITICAL。分类后由 caller 按风险等级分流：
//   - Exact：可直接判 CRITICAL
//   - Candidate：最高 HIGH，必须人工复核
//   - NoMatch：不构成命中
type PartyMatchKind int

const (
	PartyNoMatch PartyMatchKind = iota
	PartyCandidateMatch
	PartyExactNormalizedMatch
)

// String 便于日志和测试输出
func (k PartyMatchKind) String() string {
	switch k {
	case PartyCandidateMatch:
		return "candidate"
	case PartyExactNormalizedMatch:
		return "exact"
	default:
		return "none"
	}
}

// classifyPartyMatch 把两个名称的匹配关系分成三态：
//   - PartyExactNormalizedMatch：去除大小写、空白、公司类型后缀后完全相等
//   - PartyCandidateMatch：单向/双向包含或简称——只能作为候选
//   - PartyNoMatch：完全无关，或任一输入过短/为空
//
// 短子串防护：任一侧规范化后长度 < 2 直接判 NoMatch，
// 避免 "华" 匹配 "华为技术有限公司" 之类短通用词误报。
func (s *conflictDetectionService) classifyPartyMatch(name1, name2 string) PartyMatchKind {
	n1 := s.cleanCompanyName(strings.ToLower(strings.TrimSpace(name1)))
	n2 := s.cleanCompanyName(strings.ToLower(strings.TrimSpace(name2)))
	if !isMeaningfulPartyName(n1) || !isMeaningfulPartyName(n2) {
		return PartyNoMatch
	}
	if n1 == n2 {
		return PartyExactNormalizedMatch
	}
	if strings.Contains(n1, n2) || strings.Contains(n2, n1) {
		return PartyCandidateMatch
	}
	return PartyNoMatch
}

// isMeaningfulPartyName 规范化后是否含有有效辨识内容。
// 太短（<2 字符）视为无意义，避免单字触发 CRITICAL。
func isMeaningfulPartyName(name string) bool {
	if len(strings.TrimSpace(name)) < 2 {
		return false
	}
	return true
}

// isPartyNameMatch 仅在 Exact 规范化相等时返回 true。
// 调用方需要区分 Candidate 时，请直接使用 classifyPartyMatch。
func (s *conflictDetectionService) isPartyNameMatch(name1, name2 string) bool {
	return s.classifyPartyMatch(name1, name2) == PartyExactNormalizedMatch
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
	timeRange := "系统已登记历史（覆盖完整性待确认，未登记档案需人工核查）"
	if s.conflictCoverageStatus(ctx) == ConflictCoverageComplete {
		timeRange = "已由律所确认完整覆盖的已登记历史"
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
	auditResponse := *response
	auditResponse.NormalizedSubjects = auditSafeNormalizedSubjects(response.NormalizedSubjects)
	record := &models.ConflictCheckRecord{
		CheckID:          response.CheckID,
		ClientID:         request.ClientID,
		ClientName:       request.ClientName,
		CaseName:         request.CaseName,
		CaseType:         request.CaseType,
		CheckStatus:      "COMPLETED",
		HasConflict:      response.HasConflict,
		RiskLevel:        riskLevel,
		SearchParameters: toConflictJSON(auditSafeConflictRequest(request)),
		CheckResult:      toConflictJSON(&auditResponse),
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
		return s.saveP0EvidenceIfAvailable(ctx, response.CheckID, response.NormalizedSubjects, response)
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

	if err := s.conflictRepo.SaveConflictCases(ctx, response.ConflictCases); err != nil {
		return err
	}
	return s.saveP0EvidenceIfAvailable(ctx, response.CheckID, response.NormalizedSubjects, response)
}

// saveP0EvidenceIfAvailable keeps legacy test doubles compatible while making
// the production schema contract observable. Production readiness rejects a
// database without these tables, so silently skipping is limited to legacy
// non-production repositories that cannot provide the optional writer.
func (s *conflictDetectionService) saveP0EvidenceIfAvailable(ctx context.Context, checkID string, subjects []models.ConflictNormalizedSubject, response *models.ConflictCheckResponse) error {
	writer, ok := s.conflictRepo.(repositories.ConflictP0EvidenceWriter)
	if !ok || s.caseRepo == nil || s.caseRepo.GetDB() == nil {
		return nil
	}
	if !s.caseRepo.GetDB().Migrator().HasTable((&models.ConflictSubjectVersion{}).TableName()) {
		return nil
	}
	if err := writer.SaveConflictP0Evidence(ctx, checkID, subjects, response); err != nil {
		return fmt.Errorf("保存 P0 主体与命中证据失败: %w", err)
	}
	return nil
}

// auditSafeConflictRequest keeps matching inputs useful for audit/replay while
// preventing raw identity values from being written to JSON search snapshots.
// The live request still contains the original values while the check runs.
func auditSafeConflictRequest(request *models.ConflictCheckRequest) *models.ConflictCheckRequest {
	if request == nil {
		return &models.ConflictCheckRequest{}
	}
	copyRequest := *request
	copyRequest.ClientIdentifiers = digestIdentifiers(request.ClientIdentifiers)
	copyRequest.Parties = append([]models.ConflictPartyInfo(nil), request.Parties...)
	for i := range copyRequest.Parties {
		copyRequest.Parties[i].Identifiers = digestIdentifiers(request.Parties[i].Identifiers)
		copyRequest.Parties[i].Aliases = append([]string(nil), request.Parties[i].Aliases...)
	}
	return &copyRequest
}

func digestIdentifiers(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	digests := make(map[string]string, len(values))
	for key, value := range values {
		digest, err := security.IdentityDigest(value)
		if err != nil {
			// Audit snapshots must never fall back to a guessable digest when
			// the dedicated subject-data key is unavailable.
			digests[strings.ToLower(strings.TrimSpace(key))] = "已保护标识"
			continue
		}
		digests[strings.ToLower(strings.TrimSpace(key))] = "hmac-sha256:" + digest
	}
	return digests
}

func auditSafeNormalizedSubjects(subjects []models.ConflictNormalizedSubject) []models.ConflictNormalizedSubject {
	if len(subjects) == 0 {
		return nil
	}
	copySubjects := make([]models.ConflictNormalizedSubject, len(subjects))
	copy(copySubjects, subjects)
	for i := range copySubjects {
		copySubjects[i].Identifiers = digestIdentifiers(subjects[i].Identifiers)
		copySubjects[i].Aliases = append([]string(nil), subjects[i].Aliases...)
	}
	return copySubjects
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
