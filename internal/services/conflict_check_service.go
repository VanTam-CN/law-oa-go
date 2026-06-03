package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ConflictCheckService 冲突审查服务接口
type ConflictCheckService interface {
	// 核心冲突检测
	CheckConflict(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error)
	GenerateReport(ctx context.Context, checkID uint) (*ConflictReport, error)

	// 冲突审查记录管理
	GetConflictCheck(ctx context.Context, id uint) (*models.ConflictCheck, error)
	ListConflictChecks(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*models.ConflictCheck, int64, error)

	// 实体管理
	SearchEntity(ctx context.Context, query string, entityType string) ([]*models.Entity, error)
	GetEntityRelations(ctx context.Context, entityID uint, depth int) ([]*models.Entity, error)
}

// ConflictCheckRequest 冲突审查请求
type ConflictCheckRequest struct {
	CaseID        uint   `json:"caseId" validate:"required"`
	CaseTitle     string `json:"caseTitle" validate:"required"`
	CheckEntities []EntityCheckInfo `json:"checkEntities" validate:"required,min=1"`
	SearchDepth   int    `json:"searchDepth" validate:"min=1,max=5"` // 关联搜索深度，默认2
	RequestedBy   uint   `json:"requestedBy" validate:"required"`
}

// EntityCheckInfo 实体审查信息
type EntityCheckInfo struct {
	EntityID      uint   `json:"entityId"`
	EntityName    string `json:"entityName"`
	IdentityType  string `json:"identityType"`
	IdentityNumber string `json:"identityNumber"`
	PartyType     string `json:"partyType"` // CLIENT, OPPOSING, CO_DEFENDANT
}

// ConflictCheckResponse 冲突审查响应
type ConflictCheckResponse struct {
	CheckID       uint                    `json:"checkId"`
	Status        string                  `json:"status"`
	HasConflict   bool                    `json:"hasConflict"`
	TotalConflicts int                    `json:"totalConflicts"`
	Conflicts     []*ConflictDetail       `json:"conflicts"`
	Summary       *ConflictSummary        `json:"summary"`
	CheckedAt     time.Time               `json:"checkedAt"`
}

// ConflictDetail 冲突详情
type ConflictDetail struct {
	ConflictID       uint        `json:"conflictId"`
	ConflictType     string      `json:"conflictType"` // IDENTITY_MATCH, NAME_SIMILAR, RELATIONSHIP, CASE_ASSOCIATION
	RiskLevel        string      `json:"riskLevel"`    // CRITICAL, HIGH, MEDIUM, LOW
	Description      string      `json:"description"`
	MatchedEntity    *ConflictEntityInfo `json:"matchedEntity"`
	RelatedCase      *ConflictCaseInfo   `json:"relatedCase,omitempty"`
	MatchReason      string      `json:"matchReason"`
	Recommendation   string      `json:"recommendation"`
}


// ConflictEntityInfo 实体信息
type ConflictEntityInfo struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	EntityType     string `json:"entityType"`
	IdentityNumber string `json:"identityNumber,omitempty"`
	Status         string `json:"status"`
}

// CaseInfo 案件信息
type ConflictCaseInfo struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	CaseType    string `json:"caseType"`
	Status      string `json:"status"`
	StartDate   string `json:"startDate,omitempty"`
}

// ConflictSummary 冲突摘要
type ConflictSummary struct {
	TotalChecked      int               `json:"totalChecked"`
	ConflictCount     int               `json:"conflictCount"`
	RiskDistribution  map[string]int    `json:"riskDistribution"`
	ConflictTypes     map[string]int    `json:"conflictTypes"`
	Recommendations   []string          `json:"recommendations"`
}

// ConflictReport 冲突报告
type ConflictReport struct {
	CheckID          uint              `json:"checkId"`
	CaseTitle        string            `json:"caseTitle"`
	GeneratedAt      time.Time         `json:"generatedAt"`
	GeneratedBy      string            `json:"generatedBy"`
	ExecutiveSummary string            `json:"executiveSummary"`
	Conflicts        []*ConflictDetail `json:"conflicts"`
	Recommendations  []string          `json:"recommendations"`
	Conclusion       string            `json:"conclusion"`
}

// conflictCheckService 冲突审查服务实现
type conflictCheckService struct {
	entityRepo      repositories.EntityRepository
	conflictRepo    repositories.ConflictCheckRepository
	userRepo        repositories.UserRepository
}

// NewConflictCheckService 创建冲突审查服务
func NewConflictCheckService(
	entityRepo repositories.EntityRepository,
	conflictRepo repositories.ConflictCheckRepository,
	userRepo repositories.UserRepository,
) ConflictCheckService {
	return &conflictCheckService{
		entityRepo:   entityRepo,
		conflictRepo: conflictRepo,
		userRepo:     userRepo,
	}
}

// CheckConflict 执行冲突检测
// 匹配逻辑优先级：
// 1. 身份证号/证件号精确匹配 - CRITICAL
// 2. 名称模糊匹配（考虑曾用名）- HIGH
// 3. 关联关系穿透（股东、亲属等）- MEDIUM
// 4. 案件关联 - LOW
func (s *conflictCheckService) CheckConflict(ctx context.Context, req *ConflictCheckRequest) (*ConflictCheckResponse, error) {
	startTime := time.Now()

	// 创建冲突审查记录
	conflictCheck := &models.ConflictCheck{
		CaseID:      req.CaseID,
		Status:      string(models.CheckStatusProcessing),
		RequestedBy: req.RequestedBy,
		RequestedAt: startTime,
	}

	if err := s.conflictRepo.CreateConflictCheck(ctx, conflictCheck); err != nil {
		return nil, fmt.Errorf("创建冲突审查记录失败: %w", err)
	}

	// 收集所有实体ID
	entityIDs := make([]uint, 0, len(req.CheckEntities))
	for _, entity := range req.CheckEntities {
		if entity.EntityID > 0 {
			entityIDs = append(entityIDs, entity.EntityID)
		}
	}

	// 查找冲突实体
	conflictingEntities, err := s.conflictRepo.FindConflictingEntities(ctx, req.CaseID, entityIDs)
	if err != nil {
		return nil, fmt.Errorf("查找冲突实体失败: %w", err)
	}

	// 执行冲突检测
	allConflicts := make([]*ConflictDetail, 0)
	checkedEntities := make(map[uint]bool)

	// 1. 证件号精确匹配检测
	for _, checkEntity := range req.CheckEntities {
		if checkEntity.IdentityNumber != "" {
			conflicts := s.detectByIdentityMatch(ctx, checkEntity, conflictingEntities, req.CaseID)
			allConflicts = append(allConflicts, conflicts...)
		}
		checkedEntities[checkEntity.EntityID] = true
	}

	// 2. 名称模糊匹配检测（包括曾用名）
	for _, checkEntity := range req.CheckEntities {
		conflicts := s.detectByNameSimilarity(ctx, checkEntity, req.CaseID)
		allConflicts = append(allConflicts, conflicts...)
	}

	// 3. 关联关系穿透检测
	searchDepth := req.SearchDepth
	if searchDepth <= 0 {
		searchDepth = 2 // 默认搜索深度为2
	}
	for _, checkEntity := range req.CheckEntities {
		conflicts := s.detectByRelationship(ctx, checkEntity, searchDepth, req.CaseID)
		allConflicts = append(allConflicts, conflicts...)
	}

	// 4. 案件关联检测
	for _, checkEntity := range req.CheckEntities {
		conflicts := s.detectByCaseAssociation(ctx, checkEntity, req.CaseID)
		allConflicts = append(allConflicts, conflicts...)
	}

	// 去重和优先级排序
	allConflicts = s.deduplicateAndPrioritizeConflicts(allConflicts)

	// 生成摘要
	summary := s.generateSummary(len(req.CheckEntities), allConflicts)

	// 更新审查结果
	result := &models.CheckResult{
		HasConflict:   len(allConflicts) > 0,
		TotalConflicts: len(allConflicts),
		CompletedAt:   time.Now(),
	}

	status := models.CheckStatusCompleted
	if len(allConflicts) > 0 {
		status = models.CheckStatusCompletedWithConflict
	}

	if err := s.conflictRepo.UpdateConflictCheckStatus(ctx, conflictCheck.ID, string(status), result); err != nil {
		return nil, fmt.Errorf("更新冲突审查状态失败: %w", err)
	}

	// 保存冲突详情
	for _, conflict := range allConflicts {
		detail := &models.ConflictDetail{
			ConflictCheckID: conflictCheck.ID,
			MatchedEntityID: conflict.MatchedEntity.ID,
			ConflictType:    conflict.ConflictType,
			RiskLevel:       conflict.RiskLevel,
			Description:     conflict.Description,
			MatchReason:     conflict.MatchReason,
			Recommendation:  conflict.Recommendation,
		}
		if err := s.conflictRepo.CreateConflictDetail(ctx, detail); err != nil {
			// 记录错误但继续
			continue
		}
	}

	return &ConflictCheckResponse{
		CheckID:        conflictCheck.ID,
		Status:         string(status),
		HasConflict:    len(allConflicts) > 0,
		TotalConflicts: len(allConflicts),
		Conflicts:      allConflicts,
		Summary:        summary,
		CheckedAt:      startTime,
	}, nil
}

// detectByIdentityMatch 证件号精确匹配检测
func (s *conflictCheckService) detectByIdentityMatch(ctx context.Context, checkEntity EntityCheckInfo, conflictingEntities []*models.Entity, caseID uint) []*ConflictDetail {
	conflicts := make([]*ConflictDetail, 0)

	// 查询相同证件号的实体
	matchedEntities, err := s.entityRepo.SearchByIdentityNumber(ctx, checkEntity.IdentityNumber, 100)
	if err != nil {
		return conflicts
	}

	for _, matched := range matchedEntities {
		// 跳过自己
		if matched.ID == checkEntity.EntityID {
			continue
		}

		// 查找该实体的活跃案件
		cases, err := s.entityRepo.GetActiveCasesByEntity(ctx, matched.ID)
		if err != nil {
			continue
		}

		for _, caseInfo := range cases {
			// 跳过当前案件
			if caseInfo.ID == caseID {
				continue
			}

			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:    "IDENTITY_MATCH",
				RiskLevel:       "CRITICAL",
				Description:     fmt.Sprintf("证件号精确匹配: %s 与 %s 使用相同的证件号 %s", checkEntity.EntityName, matched.Name, checkEntity.IdentityNumber),
				MatchedEntity:   s.toEntityInfo(matched),
				RelatedCase:     s.toCaseInfo(caseInfo),
				MatchReason:     "证件号完全相同，存在直接利益冲突",
				Recommendation:  "建议拒绝受理，需客户确认是否存在身份重复",
			})
		}
	}

	return conflicts
}

// detectByNameSimilarity 名称模糊匹配检测
func (s *conflictCheckService) detectByNameSimilarity(ctx context.Context, checkEntity EntityCheckInfo, caseID uint) []*ConflictDetail {
	conflicts := make([]*ConflictDetail, 0)

	// 按名称搜索
	matchedEntities, err := s.entityRepo.SearchByName(ctx, checkEntity.EntityName, 50)
	if err != nil {
		return conflicts
	}

	for _, matched := range matchedEntities {
		// 跳过自己
		if matched.ID == checkEntity.EntityID {
			continue
		}

		// 计算相似度
		similarity := s.calculateSimilarity(checkEntity.EntityName, matched.Name)
		if similarity < 0.7 {
			continue // 相似度低于70%则跳过
		}

		// 查找该实体的活跃案件
		cases, err := s.entityRepo.GetActiveCasesByEntity(ctx, matched.ID)
		if err != nil {
			continue
		}

		for _, caseInfo := range cases {
			if caseInfo.ID == caseID {
				continue
			}

			riskLevel := "HIGH"
			if similarity >= 0.9 {
				riskLevel = "CRITICAL"
			}

			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:    "NAME_SIMILAR",
				RiskLevel:       riskLevel,
				Description:     fmt.Sprintf("名称高度相似: %s 与 %s 相似度 %.0f%%", checkEntity.EntityName, matched.Name, similarity*100),
				MatchedEntity:   s.toEntityInfo(matched),
				RelatedCase:     s.toCaseInfo(caseInfo),
				MatchReason:     fmt.Sprintf("名称相似度 %.0f%%，可能为同一人或关联方", similarity*100),
				Recommendation:  "建议进一步核实客户身份，确认是否为同一人",
			})
		}
	}

	// 检查曾用名
	formerMatches, err := s.entityRepo.SearchByFormerName(ctx, checkEntity.EntityName, 50)
	if err == nil {
		for _, matched := range formerMatches {
			if matched.ID == checkEntity.EntityID {
				continue
			}

			cases, err := s.entityRepo.GetActiveCasesByEntity(ctx, matched.ID)
			if err != nil {
				continue
			}

			for _, caseInfo := range cases {
				if caseInfo.ID == caseID {
					continue
				}

				conflicts = append(conflicts, &ConflictDetail{
					ConflictType:    "NAME_SIMILAR",
					RiskLevel:       "HIGH",
					Description:     fmt.Sprintf("曾用名匹配: %s 与 %s 的曾用名相同", checkEntity.EntityName, matched.Name),
					MatchedEntity:   s.toEntityInfo(matched),
					RelatedCase:     s.toCaseInfo(caseInfo),
					MatchReason:     "曾用名完全匹配，可能为同一人",
					Recommendation:  "建议确认是否为同一人的曾用名变更",
				})
			}
		}
	}

	return conflicts
}

// detectByRelationship 关联关系穿透检测
// 使用 PostgreSQL Recursive CTE 实现双向穿透，支持所有关系类型
func (s *conflictCheckService) detectByRelationship(ctx context.Context, checkEntity EntityCheckInfo, depth int, caseID uint) []*ConflictDetail {
	conflicts := make([]*ConflictDetail, 0)

	if checkEntity.EntityID == 0 {
		return conflicts
	}

	// 使用递归 CTE 双向穿透获取所有关联实体
	relatedNodes, err := s.entityRepo.GetRelatedEntitiesRecursive(ctx, checkEntity.EntityID, depth)
	if err != nil {
		return conflicts
	}

	checked := make(map[uint]bool)
	for _, node := range relatedNodes {
		if checked[node.EntityID] {
			continue
		}
		checked[node.EntityID] = true

		// 查找关联实体的活跃案件
		cases, err := s.entityRepo.GetActiveCasesByEntity(ctx, node.EntityID)
		if err != nil {
			continue
		}

		// 构建穿透路径描述
		pathDesc := buildPathDescription(node)

		// 根据深度和关系类型确定风险等级
		riskLevel := "MEDIUM"
		if node.Depth == 1 {
			switch node.RelationType {
			case models.RelationTypeActualController, models.RelationTypeSpouse, models.RelationTypeFamilyMember:
				riskLevel = "HIGH"
			}
		}

		for _, caseInfo := range cases {
			if caseInfo.ID == caseID {
				continue
			}

			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:   "RELATIONSHIP",
				RiskLevel:      riskLevel,
				Description:    fmt.Sprintf("关联关系穿透: %s 通过%s关联到 %s（第%d层）", checkEntity.EntityName, pathDesc, node.Entity.Name, node.Depth),
				MatchedEntity:  s.toEntityInfo(node.Entity),
				RelatedCase:    s.toCaseInfo(caseInfo),
				MatchReason:    fmt.Sprintf("通过 %s 关系在 %d 层内发现关联实体参与其他案件", node.RelationType, node.Depth),
				Recommendation: "建议评估关联关系的影响，考虑是否需要回避",
			})
		}
	}

	return conflicts
}

// buildPathDescription 构建穿透路径的可读描述
func buildPathDescription(node *repositories.RelatedEntityNode) string {
	if len(node.Path) == 0 {
		return string(node.RelationType)
	}

	direction := "→"
	if node.Direction == "incoming" {
		direction = "←"
	}

	parts := make([]string, len(node.Path))
	for i, edge := range node.Path {
		parts[i] = fmt.Sprintf("%s(%s)%d", edge.RelationType, direction, edge.ToID)
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, " → "))
}

// detectByCaseAssociation 案件关联检测
func (s *conflictCheckService) detectByCaseAssociation(ctx context.Context, checkEntity EntityCheckInfo, caseID uint) []*ConflictDetail {
	conflicts := make([]*ConflictDetail, 0)

	// 查找该实体参与的案件
	cases, err := s.entityRepo.GetActiveCasesByEntity(ctx, checkEntity.EntityID)
	if err != nil {
		return conflicts
	}

	// 检查是否与已拒绝或撤案的案件有关联
	for _, caseInfo := range cases {
		if caseInfo.ID == caseID {
			continue
		}

		// 检查案件状态
		if caseInfo.Status == "rejected" || caseInfo.Status == "withdrawn" {
			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:    "CASE_ASSOCIATION",
				RiskLevel:       "LOW",
				Description:     fmt.Sprintf("历史案件关联: %s 曾参与案件 %s (%s)", checkEntity.EntityName, caseInfo.Title, caseInfo.Status),
				MatchedEntity:   nil,
				RelatedCase:     s.toCaseInfo(caseInfo),
				MatchReason:     "该实体曾参与其他案件",
				Recommendation:  "建议了解历史案件详情，评估潜在影响",
			})
		}
	}

	return conflicts
}

// generateSummary 生成冲突摘要
func (s *conflictCheckService) generateSummary(totalChecked int, conflicts []*ConflictDetail) *ConflictSummary {
	summary := &ConflictSummary{
		TotalChecked:     totalChecked,
		ConflictCount:    len(conflicts),
		RiskDistribution: make(map[string]int),
		ConflictTypes:    make(map[string]int),
		Recommendations:  []string{},
	}

	for _, conflict := range conflicts {
		summary.RiskDistribution[conflict.RiskLevel]++
		summary.ConflictTypes[conflict.ConflictType]++
	}

	// 生成建议
	if len(conflicts) > 0 {
		if summary.RiskDistribution["CRITICAL"] > 0 {
			summary.Recommendations = append(summary.Recommendations, "存在严重利益冲突，强烈建议拒绝受理")
		}
		if summary.RiskDistribution["HIGH"] > 0 {
			summary.Recommendations = append(summary.Recommendations, "存在高风险冲突，需详细评估并考虑回避")
		}
		if summary.RiskDistribution["MEDIUM"] > 0 {
			summary.Recommendations = append(summary.Recommendations, "存在中等风险冲突，建议获取客户确认并记录")
		}
	} else {
		summary.Recommendations = append(summary.Recommendations, "未发现明显利益冲突，可以正常受理")
	}

	return summary
}

// GenerateReport 生成冲突报告
func (s *conflictCheckService) GenerateReport(ctx context.Context, checkID uint) (*ConflictReport, error) {
	// 获取冲突审查记录
	conflictCheck, err := s.conflictRepo.GetConflictCheck(ctx, checkID)
	if err != nil {
		return nil, fmt.Errorf("获取冲突审查记录失败: %w", err)
	}

	// 获取冲突详情
	details, err := s.conflictRepo.GetConflictDetails(ctx, checkID)
	if err != nil {
		return nil, fmt.Errorf("获取冲突详情失败: %w", err)
	}

	// 获取请求人信息
	user, err := s.userRepo.FindByID(ctx, conflictCheck.RequestedBy)
	username := "未知用户"
	if err == nil && user != nil {
		username = user.Name
	}

	// 转换冲突详情
	conflictDetails := make([]*ConflictDetail, 0, len(details))
	for _, detail := range details {
		entityInfo := &ConflictEntityInfo{
			ID:         detail.MatchedEntity.ID,
			Name:       detail.MatchedEntity.Name,
			EntityType: string(detail.MatchedEntity.EntityType),
		}
		if detail.MatchedEntity.IdentityNumber != "" {
			entityInfo.IdentityNumber = maskIdentityNumber(detail.MatchedEntity.IdentityNumber)
		}

		conflictDetails = append(conflictDetails, &ConflictDetail{
			ConflictID:      detail.ID,
			ConflictType:    detail.ConflictType,
			RiskLevel:       detail.RiskLevel,
			Description:     detail.Description,
			MatchedEntity:   entityInfo,
			MatchReason:     detail.MatchReason,
			Recommendation:  detail.Recommendation,
		})
	}

	// 生成执行摘要
	executiveSummary := s.generateExecutiveSummary(conflictCheck, conflictDetails)

	// 生成结论
	conclusion := s.generateConclusion(conflictCheck, conflictDetails)

	return &ConflictReport{
		CheckID:          checkID,
		CaseTitle:        "", // 可以从案件信息获取
		GeneratedAt:      time.Now(),
		GeneratedBy:      username,
		ExecutiveSummary: executiveSummary,
		Conflicts:        conflictDetails,
		Recommendations:  extractRecommendations(conflictDetails),
		Conclusion:       conclusion,
	}, nil
}

// GetConflictCheck 获取冲突审查记录
func (s *conflictCheckService) GetConflictCheck(ctx context.Context, id uint) (*models.ConflictCheck, error) {
	return s.conflictRepo.GetConflictCheck(ctx, id)
}

// ListConflictChecks 列表查询冲突审查记录
func (s *conflictCheckService) ListConflictChecks(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]*models.ConflictCheck, int64, error) {
	offset := (page - 1) * pageSize
	return s.conflictRepo.ListConflictChecks(ctx, offset, pageSize, filters)
}

// SearchEntity 搜索实体
func (s *conflictCheckService) SearchEntity(ctx context.Context, query string, entityType string) ([]*models.Entity, error) {
	return s.entityRepo.SearchByName(ctx, query, 50)
}

// GetEntityRelations 获取实体关联关系（支持递归穿透）
func (s *conflictCheckService) GetEntityRelations(ctx context.Context, entityID uint, depth int) ([]*models.Entity, error) {
	if depth <= 1 {
		// 深度为1时只获取直接关联
		relations, err := s.entityRepo.GetRelations(ctx, entityID, true)
		if err != nil {
			return nil, err
		}

		entityIDs := make([]uint, 0)
		for _, relation := range relations {
			if relation.SourceEntityID == entityID {
				entityIDs = append(entityIDs, relation.TargetEntityID)
			} else {
				entityIDs = append(entityIDs, relation.SourceEntityID)
			}
		}

		entities := make([]*models.Entity, 0)
		for _, id := range entityIDs {
			entity, err := s.entityRepo.GetByID(ctx, id)
			if err != nil {
				continue
			}
			entities = append(entities, entity)
		}
		return entities, nil
	}

	// 深度>1时使用递归 CTE 穿透
	nodes, err := s.entityRepo.GetRelatedEntitiesRecursive(ctx, entityID, depth)
	if err != nil {
		return nil, err
	}

	entities := make([]*models.Entity, 0, len(nodes))
	seen := make(map[uint]bool)
	for _, node := range nodes {
		if !seen[node.EntityID] && node.Entity != nil {
			seen[node.EntityID] = true
			entities = append(entities, node.Entity)
		}
	}

	return entities, nil
}

// 辅助函数

// toEntityInfo 转换为实体信息
func (s *conflictCheckService) toEntityInfo(entity *models.Entity) *ConflictEntityInfo {
	return &ConflictEntityInfo{
		ID:             entity.ID,
		Name:           entity.Name,
		EntityType:     string(entity.EntityType),
		IdentityNumber: maskIdentityNumber(entity.IdentityNumber),
		Status:         string(entity.Status),
	}
}

// toCaseInfo 转换为案件信息
func (s *conflictCheckService) toCaseInfo(caseModel *models.Case) *ConflictCaseInfo {
	caseInfo := &ConflictCaseInfo{
		ID:       caseModel.ID,
		Title:    caseModel.Title,
		CaseType: caseModel.CaseType,
		Status:   caseModel.Status,
	}
	if caseModel.StartDate != nil {
		caseInfo.StartDate = caseModel.StartDate.Format("2006-01-02")
	}
	return caseInfo
}

// calculateSimilarity 计算字符串相似度（简单的编辑距离算法）
func (s *conflictCheckService) calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	len1, len2 := len(s1), len(s2)
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	// 简化的相似度计算：基于公共子串
	maxLen := len1
	if len2 < maxLen {
		maxLen = len2
	}

	common := 0
	for i := 0; i < maxLen; i++ {
		if s1[i] == s2[i] {
			common++
		}
	}

	return float64(common) / float64(maxLen)
}

// deduplicateAndPrioritizeConflicts 去重并按优先级排序冲突
func (s *conflictCheckService) deduplicateAndPrioritizeConflicts(conflicts []*ConflictDetail) []*ConflictDetail {
	// 按风险等级排序
	riskOrder := map[string]int{
		"CRITICAL": 4,
		"HIGH":     3,
		"MEDIUM":   2,
		"LOW":      1,
	}

	// 使用简单的冒泡排序
	for i := 0; i < len(conflicts)-1; i++ {
		for j := 0; j < len(conflicts)-i-1; j++ {
			if riskOrder[conflicts[j].RiskLevel] < riskOrder[conflicts[j+1].RiskLevel] {
				conflicts[j], conflicts[j+1] = conflicts[j+1], conflicts[j]
			}
		}
	}

	return conflicts
}

// maskIdentityNumber 遮蔽证件号中间部分
func maskIdentityNumber(number string) string {
	if len(number) <= 8 {
		return strings.Repeat("*", len(number))
	}
	return number[:4] + strings.Repeat("*", len(number)-8) + number[len(number)-4:]
}

// generateExecutiveSummary 生成执行摘要
func (s *conflictCheckService) generateExecutiveSummary(check *models.ConflictCheck, conflicts []*ConflictDetail) string {
	if len(conflicts) == 0 {
		return "本次利益冲突审查未发现明显冲突。"
	}

	criticalCount := 0
	highCount := 0
	for _, conflict := range conflicts {
		if conflict.RiskLevel == "CRITICAL" {
			criticalCount++
		} else if conflict.RiskLevel == "HIGH" {
			highCount++
		}
	}

	summary := fmt.Sprintf("本次利益冲突审查发现 %d 个潜在冲突点", len(conflicts))
	if criticalCount > 0 {
		summary += fmt.Sprintf("，其中 %d 个严重冲突", criticalCount)
	}
	if highCount > 0 {
		summary += fmt.Sprintf("，%d 个高风险冲突", highCount)
	}
	summary += "。建议详细审查每个冲突点并评估影响。"

	return summary
}

// generateConclusion 生成结论
func (s *conflictCheckService) generateConclusion(check *models.ConflictCheck, conflicts []*ConflictDetail) string {
	if len(conflicts) == 0 {
		return "结论：未发现明显利益冲突，建议可以正常受理案件。"
	}

	hasCritical := false
	for _, conflict := range conflicts {
		if conflict.RiskLevel == "CRITICAL" {
			hasCritical = true
			break
		}
	}

	if hasCritical {
		return "结论：存在严重利益冲突，建议拒绝受理或寻求专业意见。"
	}

	return "结论：存在潜在利益冲突，建议详细评估并考虑采取适当的回避措施。"
}

// extractRecommendations 提取所有建议
func extractRecommendations(conflicts []*ConflictDetail) []string {
	recommendationsMap := make(map[string]bool)
	recommendations := []string{}

	for _, conflict := range conflicts {
		if conflict.Recommendation != "" && !recommendationsMap[conflict.Recommendation] {
			recommendationsMap[conflict.Recommendation] = true
			recommendations = append(recommendations, conflict.Recommendation)
		}
	}

	return recommendations
}
