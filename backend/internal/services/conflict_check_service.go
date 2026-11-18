package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/security"

	"gorm.io/gorm"
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
	CaseID        uint              `json:"caseId" validate:"required"`
	CaseTitle     string            `json:"caseTitle" validate:"required"`
	CheckEntities []EntityCheckInfo `json:"checkEntities" validate:"required,min=1"`
	SearchDepth   int               `json:"searchDepth" validate:"min=1,max=5"` // 关联搜索深度，默认2
	RequestedBy   uint              `json:"requestedBy" validate:"required"`
}

// EntityCheckInfo 实体审查信息
type EntityCheckInfo struct {
	EntityID       uint   `json:"entityId"`
	EntityName     string `json:"entityName"`
	IdentityType   string `json:"identityType"`
	IdentityNumber string `json:"identityNumber"`
	PartyType      string `json:"partyType"` // CLIENT, OPPOSING, CO_DEFENDANT
}

// ConflictCheckResponse 冲突审查响应
type ConflictCheckResponse struct {
	CheckID        uint              `json:"checkId"`
	Status         string            `json:"status"`
	HasConflict    bool              `json:"hasConflict"`
	TotalConflicts int               `json:"totalConflicts"`
	Conflicts      []*ConflictDetail `json:"conflicts"`
	Summary        *ConflictSummary  `json:"summary"`
	CheckedAt      time.Time         `json:"checkedAt"`
	CoverageStatus string            `json:"coverageStatus"`
	CoverageNotice string            `json:"coverageNotice"`
}

// ConflictDetail 冲突详情
type ConflictDetail struct {
	ConflictID     uint                `json:"conflictId"`
	ConflictType   string              `json:"conflictType"` // IDENTITY_MATCH, NAME_SIMILAR, RELATIONSHIP, CASE_ASSOCIATION
	RiskLevel      string              `json:"riskLevel"`    // CRITICAL, HIGH, MEDIUM, LOW
	Description    string              `json:"description"`
	MatchedEntity  *ConflictEntityInfo `json:"matchedEntity"`
	RelatedCase    *ConflictCaseInfo   `json:"relatedCase,omitempty"`
	MatchReason    string              `json:"matchReason"`
	Recommendation string              `json:"recommendation"`
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
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	CaseType  string `json:"caseType"`
	Status    string `json:"status"`
	StartDate string `json:"startDate,omitempty"`
}

// ConflictSummary 冲突摘要
type ConflictSummary struct {
	TotalChecked     int            `json:"totalChecked"`
	ConflictCount    int            `json:"conflictCount"`
	RiskDistribution map[string]int `json:"riskDistribution"`
	ConflictTypes    map[string]int `json:"conflictTypes"`
	Recommendations  []string       `json:"recommendations"`
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
	entityRepo   repositories.EntityRepository
	conflictRepo repositories.ConflictCheckRepository
	userRepo     repositories.UserRepository
	db           *gorm.DB
}

// NewConflictCheckService 创建冲突审查服务
func NewConflictCheckService(
	entityRepo repositories.EntityRepository,
	conflictRepo repositories.ConflictCheckRepository,
	userRepo repositories.UserRepository,
	databases ...*gorm.DB,
) ConflictCheckService {
	var db *gorm.DB
	if len(databases) > 0 {
		db = databases[0]
	}
	return &conflictCheckService{
		entityRepo:   entityRepo,
		conflictRepo: conflictRepo,
		userRepo:     userRepo,
		db:           db,
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
		return nil, s.failConflictCheck(ctx, conflictCheck.ID, fmt.Errorf("查找冲突实体失败: %w", err))
	}

	// 执行冲突检测
	allConflicts := make([]*ConflictDetail, 0)
	checkedEntities := make(map[uint]bool)

	// 1. 证件号精确匹配检测
	for _, checkEntity := range req.CheckEntities {
		if checkEntity.IdentityNumber != "" {
			conflicts, err := s.detectByIdentityMatchChecked(ctx, checkEntity, conflictingEntities, req.CaseID)
			if err != nil {
				return nil, s.failConflictCheck(ctx, conflictCheck.ID, err)
			}
			allConflicts = append(allConflicts, conflicts...)
		}
		checkedEntities[checkEntity.EntityID] = true
	}

	// 2. 名称模糊匹配检测（包括曾用名）
	for _, checkEntity := range req.CheckEntities {
		conflicts, err := s.detectByNameSimilarityChecked(ctx, checkEntity, req.CaseID)
		if err != nil {
			return nil, s.failConflictCheck(ctx, conflictCheck.ID, err)
		}
		allConflicts = append(allConflicts, conflicts...)
	}

	// 3. 关联关系穿透检测
	searchDepth := req.SearchDepth
	if searchDepth <= 0 {
		searchDepth = 2 // 默认搜索深度为2
	}
	for _, checkEntity := range req.CheckEntities {
		conflicts, err := s.detectByRelationshipChecked(ctx, checkEntity, searchDepth, req.CaseID)
		if err != nil {
			return nil, s.failConflictCheck(ctx, conflictCheck.ID, err)
		}
		allConflicts = append(allConflicts, conflicts...)
	}

	// 4. 案件关联检测
	for _, checkEntity := range req.CheckEntities {
		conflicts, err := s.detectByCaseAssociationChecked(ctx, checkEntity, req.CaseID)
		if err != nil {
			return nil, s.failConflictCheck(ctx, conflictCheck.ID, err)
		}
		allConflicts = append(allConflicts, conflicts...)
	}

	// 去重和优先级排序
	allConflicts = s.deduplicateAndPrioritizeConflicts(allConflicts)
	coverageStatus := s.coverageStatus(ctx)
	coverageNotice := "本次检测尚未完成律所权威档案来源和关联主体覆盖确认，结果不得视为无冲突。"
	if coverageStatus == "COMPLETE" {
		coverageNotice = "本次检测覆盖已由律所登记并确认的冲突检索来源，仍需独立人工复核后形成处置结论。"
	}

	// 生成摘要
	summary := s.generateSummary(len(req.CheckEntities), allConflicts)
	if coverageStatus != "COMPLETE" {
		summary.Recommendations = []string{coverageNotice}
	}

	// 更新审查结果
	result := &models.CheckResult{
		HasConflict:    len(allConflicts) > 0 || coverageStatus != "COMPLETE",
		TotalConflicts: len(allConflicts),
		CompletedAt:    time.Now(),
		CoverageStatus: coverageStatus,
		CoverageNotice: coverageNotice,
	}

	status := models.CheckStatusCompleted
	if len(allConflicts) > 0 {
		status = models.CheckStatusCompletedWithConflict
		for _, conflict := range allConflicts {
			if conflict == nil {
				continue
			}
			switch conflict.ConflictType {
			case "NAME_CANDIDATE", "IDENTITY_INSUFFICIENT", "RELATION_CANDIDATE", "CASE_ASSOCIATION":
				status = "REVIEW_REQUIRED"
			}
		}
	}
	if coverageStatus != "COMPLETE" {
		status = "REVIEW_REQUIRED"
	}

	if err := s.conflictRepo.UpdateConflictCheckStatus(ctx, conflictCheck.ID, string(status), result); err != nil {
		return nil, fmt.Errorf("更新冲突审查状态失败: %w", err)
	}

	// 保存冲突详情
	for _, conflict := range allConflicts {
		if conflict == nil {
			continue
		}
		var matchedEntityID uint
		if conflict.MatchedEntity != nil {
			matchedEntityID = conflict.MatchedEntity.ID
		}
		var matchedCaseID *uint
		if conflict.RelatedCase != nil && conflict.RelatedCase.ID > 0 {
			caseID := conflict.RelatedCase.ID
			matchedCaseID = &caseID
		}
		detail := &models.ConflictDetail{
			ConflictCheckID: conflictCheck.ID,
			MatchedEntityID: matchedEntityID,
			MatchedCaseID:   matchedCaseID,
			ConflictType:    conflict.ConflictType,
			RiskLevel:       conflict.RiskLevel,
			Description:     conflict.Description,
			MatchReason:     conflict.MatchReason,
			Recommendation:  conflict.Recommendation,
		}
		if err := s.conflictRepo.CreateConflictDetail(ctx, detail); err != nil {
			return nil, s.failConflictCheck(ctx, conflictCheck.ID, fmt.Errorf("保存冲突证据失败: %w", err))
		}
	}

	return &ConflictCheckResponse{
		CheckID:        conflictCheck.ID,
		Status:         string(status),
		HasConflict:    len(allConflicts) > 0 || coverageStatus != "COMPLETE",
		TotalConflicts: len(allConflicts),
		Conflicts:      allConflicts,
		Summary:        summary,
		CheckedAt:      startTime,
		CoverageStatus: coverageStatus,
		CoverageNotice: coverageNotice,
	}, nil
}

func (s *conflictCheckService) coverageStatus(ctx context.Context) string {
	if s == nil || s.db == nil {
		return "COVERAGE_LIMITED"
	}
	if err := NewConflictScopeService(s.db).ValidateComplete(ctx); err != nil {
		return "COVERAGE_LIMITED"
	}
	return "COMPLETE"
}

// failConflictCheck records an incomplete search as review-required instead
// of allowing a repository outage to look like an empty result. The caller
// still receives an error so the HTTP request cannot continue into approval.
func (s *conflictCheckService) failConflictCheck(ctx context.Context, checkID uint, cause error) error {
	result := &models.CheckResult{
		HasConflict:    true,
		CompletedAt:    time.Now(),
		CoverageStatus: "COVERAGE_LIMITED",
		CoverageNotice: "冲突检索未完整执行，必须由独立冲突核查人进行人工复核。",
	}
	if err := s.conflictRepo.UpdateConflictCheckStatus(ctx, checkID, "REVIEW_REQUIRED", result); err != nil {
		return fmt.Errorf("冲突检索失败且无法记录待复核状态: %v; 状态更新失败: %w", cause, err)
	}
	return fmt.Errorf("冲突检索未完整执行，已转为待人工复核: %w", cause)
}

// detectByIdentityMatch 证件号精确匹配检测
func (s *conflictCheckService) detectByIdentityMatch(ctx context.Context, checkEntity EntityCheckInfo, conflictingEntities []*models.Entity, caseID uint) []*ConflictDetail {
	conflicts, _ := s.detectByIdentityMatchChecked(ctx, checkEntity, conflictingEntities, caseID)
	return conflicts
}

func (s *conflictCheckService) detectByIdentityMatchChecked(ctx context.Context, checkEntity EntityCheckInfo, conflictingEntities []*models.Entity, caseID uint) ([]*ConflictDetail, error) {
	conflicts := make([]*ConflictDetail, 0)

	// 查询相同证件号的实体
	matchedEntities, err := s.entityRepo.SearchByIdentityNumber(ctx, checkEntity.IdentityNumber, 100)
	if err != nil {
		return nil, err
	}

	for _, matched := range matchedEntities {
		// 跳过自己
		if matched.ID == checkEntity.EntityID {
			continue
		}

		// 查找该实体的活跃案件
		cases, err := s.entityRepo.GetAllCasesByEntity(ctx, matched.ID)
		if err != nil {
			return nil, err
		}

		for _, caseInfo := range cases {
			// 跳过当前案件
			if caseInfo.ID == caseID {
				continue
			}

			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:   "IDENTITY_MATCH",
				RiskLevel:      "CRITICAL",
				Description:    fmt.Sprintf("身份标识强一致: %s 与历史主体 %s 使用相同的受保护标识（%s）", checkEntity.EntityName, matched.Name, maskIdentityNumber(checkEntity.IdentityNumber)),
				MatchedEntity:  s.toEntityInfo(matched),
				RelatedCase:    s.toCaseInfo(caseInfo),
				MatchReason:    "身份标识摘要完全一致，确认同一主体仍需独立冲突复核",
				Recommendation: "暂停接案，提交独立冲突核查人判断是否回避、拒绝或按政策申请豁免；系统不自动作出法律结论。",
			})
		}
	}

	return conflicts, nil
}

// detectByNameSimilarity 名称模糊匹配检测
func (s *conflictCheckService) detectByNameSimilarity(ctx context.Context, checkEntity EntityCheckInfo, caseID uint) []*ConflictDetail {
	conflicts, _ := s.detectByNameSimilarityChecked(ctx, checkEntity, caseID)
	return conflicts
}

func (s *conflictCheckService) detectByNameSimilarityChecked(ctx context.Context, checkEntity EntityCheckInfo, caseID uint) ([]*ConflictDetail, error) {
	conflicts := make([]*ConflictDetail, 0)

	// 按名称搜索
	matchedEntities, err := s.entityRepo.SearchByName(ctx, checkEntity.EntityName, 50)
	if err != nil {
		return nil, err
	}

	for _, matched := range matchedEntities {
		// 跳过自己
		if matched.ID == checkEntity.EntityID {
			continue
		}

		// 计算名称、别名的最佳匹配。名称命中只能形成候选线索，不能
		// 覆盖同一主体的强身份标识判断。
		similarity := bestEntityNameSimilarity(checkEntity.EntityName, matched)
		if similarity < 0.7 {
			continue // 相似度低于70%则跳过
		}
		if sameIdentityDigest(checkEntity.IdentityNumber, matched) {
			// 该命中已经由证件号精确检查负责，避免重复生成两个结论。
			continue
		}

		// 查找该实体的全部历史案件
		cases, err := s.entityRepo.GetAllCasesByEntity(ctx, matched.ID)
		if err != nil {
			return nil, err
		}

		for _, caseInfo := range cases {
			if caseInfo.ID == caseID {
				continue
			}

			conflictType := "NAME_CANDIDATE"
			riskLevel := "MEDIUM"
			recommendation := "发现名称或别名关联命中，暂停接案并补充强身份标识，由独立冲突核查人复核；名称命中不等于确认同一主体。"
			matchReason := fmt.Sprintf("名称或别名相似度 %.0f%%，仅形成待核查候选", similarity*100)
			if matched.EntityType == models.EntityTypeIndividual && strings.TrimSpace(checkEntity.IdentityNumber) == "" {
				conflictType = "IDENTITY_INSUFFICIENT"
				riskLevel = "LOW"
				recommendation = "自然人仅凭姓名无法确认同一主体，请补充并核验身份证件或护照标识。"
				matchReason = "自然人姓名命中但缺少强身份标识，不能认定为同一主体"
			}

			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:   conflictType,
				RiskLevel:      riskLevel,
				Description:    fmt.Sprintf("发现名称关联候选: %s 与 %s 相似度 %.0f%%，尚不能确认同一主体", checkEntity.EntityName, matched.Name, similarity*100),
				MatchedEntity:  s.toEntityInfo(matched),
				RelatedCase:    s.toCaseInfo(caseInfo),
				MatchReason:    matchReason,
				Recommendation: recommendation,
			})
		}
	}

	// 检查曾用名
	formerMatches, err := s.entityRepo.SearchByFormerName(ctx, checkEntity.EntityName, 50)
	if err != nil {
		return nil, err
	}
	for _, matched := range formerMatches {
		if matched.ID == checkEntity.EntityID {
			continue
		}

		cases, err := s.entityRepo.GetAllCasesByEntity(ctx, matched.ID)
		if err != nil {
			return nil, err
		}

		for _, caseInfo := range cases {
			if caseInfo.ID == caseID {
				continue
			}

			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:   "NAME_CANDIDATE",
				RiskLevel:      "MEDIUM",
				Description:    fmt.Sprintf("发现曾用名关联候选: %s 与历史主体 %s 的曾用名相同，尚不能确认同一主体", checkEntity.EntityName, matched.Name),
				MatchedEntity:  s.toEntityInfo(matched),
				RelatedCase:    s.toCaseInfo(caseInfo),
				MatchReason:    "曾用名完全匹配，需核验身份标识和历史关系",
				Recommendation: "暂停接案并由独立冲突核查人核验曾用名对应主体，不得仅凭名称确认同一主体。",
			})
		}
	}

	return conflicts, nil
}

// detectByRelationship 关联关系穿透检测
// 使用 PostgreSQL Recursive CTE 实现双向穿透，支持所有关系类型
func (s *conflictCheckService) detectByRelationship(ctx context.Context, checkEntity EntityCheckInfo, depth int, caseID uint) []*ConflictDetail {
	conflicts, _ := s.detectByRelationshipChecked(ctx, checkEntity, depth, caseID)
	return conflicts
}

func (s *conflictCheckService) detectByRelationshipChecked(ctx context.Context, checkEntity EntityCheckInfo, depth int, caseID uint) ([]*ConflictDetail, error) {
	conflicts := make([]*ConflictDetail, 0)

	if checkEntity.EntityID == 0 {
		return conflicts, nil
	}

	// 使用递归 CTE 双向穿透获取所有关联实体
	relatedNodes, err := s.entityRepo.GetRelatedEntitiesRecursive(ctx, checkEntity.EntityID, depth)
	if err != nil {
		return nil, err
	}

	checked := make(map[uint]bool)
	for _, node := range relatedNodes {
		if node == nil || node.Entity == nil {
			return nil, fmt.Errorf("关联主体证据不完整")
		}
		if checked[node.EntityID] {
			continue
		}
		checked[node.EntityID] = true

		// 查找关联实体的活跃案件
		cases, err := s.entityRepo.GetAllCasesByEntity(ctx, node.EntityID)
		if err != nil {
			return nil, err
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
				ConflictType:   "RELATION_CANDIDATE",
				RiskLevel:      riskLevel,
				Description:    fmt.Sprintf("关联关系穿透: %s 通过%s关联到 %s（第%d层）", checkEntity.EntityName, pathDesc, node.Entity.Name, node.Depth),
				MatchedEntity:  s.toEntityInfo(node.Entity),
				RelatedCase:    s.toCaseInfo(caseInfo),
				MatchReason:    fmt.Sprintf("通过 %s 关系在 %d 层内发现关联实体参与其他历史案件", node.RelationType, node.Depth),
				Recommendation: "发现关联方线索，暂停相关接案动作，由独立冲突核查人单独评估，不得当作同一主体。",
			})
		}
	}

	return conflicts, nil
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
	conflicts, _ := s.detectByCaseAssociationChecked(ctx, checkEntity, caseID)
	return conflicts
}

func (s *conflictCheckService) detectByCaseAssociationChecked(ctx context.Context, checkEntity EntityCheckInfo, caseID uint) ([]*ConflictDetail, error) {
	conflicts := make([]*ConflictDetail, 0)

	// 查找该实体参与的案件
	cases, err := s.entityRepo.GetAllCasesByEntity(ctx, checkEntity.EntityID)
	if err != nil {
		return nil, err
	}

	// 检查是否与已拒绝或撤案的案件有关联
	for _, caseInfo := range cases {
		if caseInfo.ID == caseID {
			continue
		}

		// 检查案件状态
		if caseInfo.Status == "rejected" || caseInfo.Status == "withdrawn" {
			conflicts = append(conflicts, &ConflictDetail{
				ConflictType:   "CASE_ASSOCIATION",
				RiskLevel:      "LOW",
				Description:    fmt.Sprintf("历史案件关联: %s 曾参与案件 %s (%s)", checkEntity.EntityName, caseInfo.Title, caseInfo.Status),
				MatchedEntity:  &ConflictEntityInfo{ID: checkEntity.EntityID, Name: checkEntity.EntityName},
				RelatedCase:    s.toCaseInfo(caseInfo),
				MatchReason:    "该实体曾参与其他案件",
				Recommendation: "建议了解历史案件详情，评估潜在影响",
			})
		}
	}

	return conflicts, nil
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

	// 生成建议。名称、身份信息不足和关系命中是待核查线索，不能在
	// 摘要中被写成已经确认的法律冲突。
	if len(conflicts) > 0 {
		candidateCount := 0
		for _, conflict := range conflicts {
			if conflict == nil {
				continue
			}
			switch conflict.ConflictType {
			case "NAME_CANDIDATE", "IDENTITY_INSUFFICIENT", "RELATION_CANDIDATE", "CASE_ASSOCIATION":
				candidateCount++
			}
		}
		if candidateCount > 0 {
			summary.Recommendations = append(summary.Recommendations, fmt.Sprintf("发现 %d 条待核查的名称、身份或关联方线索，必须补充证据并由独立冲突核查人复核。", candidateCount))
		}
		if summary.RiskDistribution["CRITICAL"] > 0 {
			summary.Recommendations = append(summary.Recommendations, "存在身份标识强一致命中，必须暂停接案并由独立冲突核查人作出正式处置")
		}
		if summary.RiskDistribution["HIGH"] > 0 {
			summary.Recommendations = append(summary.Recommendations, "存在高风险关联线索，需详细评估并考虑回避")
		}
		if summary.RiskDistribution["MEDIUM"] > 0 {
			summary.Recommendations = append(summary.Recommendations, "存在中等风险关联线索，需补充证据并留存复核记录")
		}
	} else {
		summary.Recommendations = append(summary.Recommendations, "未发现可识别匹配记录；仍须核对本次检索范围并完成规定的独立复核")
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
		if security.IdentityPresent(detail.MatchedEntity.IdentityNumber, detail.MatchedEntity.IdentityNumberCiphertext, detail.MatchedEntity.IdentityNumberDigest) {
			identityNumber, _ := entityIdentityNumberForDisplay(&detail.MatchedEntity)
			entityInfo.IdentityNumber = maskIdentityNumber(identityNumber)
			if entityInfo.IdentityNumber == "" {
				entityInfo.IdentityNumber = "已登记（受保护）"
			}
		}

		conflictDetails = append(conflictDetails, &ConflictDetail{
			ConflictID:     detail.ID,
			ConflictType:   detail.ConflictType,
			RiskLevel:      detail.RiskLevel,
			Description:    detail.Description,
			MatchedEntity:  entityInfo,
			MatchReason:    detail.MatchReason,
			Recommendation: detail.Recommendation,
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
	identityNumber := strings.TrimSpace(entity.IdentityNumber)
	if identityNumber == "" {
		identityNumber, _ = entityIdentityNumberForDisplay(entity)
	}
	return &ConflictEntityInfo{
		ID:             entity.ID,
		Name:           entity.Name,
		EntityType:     string(entity.EntityType),
		IdentityNumber: maskIdentityNumber(identityNumber),
		Status:         string(entity.Status),
	}
}

func entityIdentityNumberForDisplay(entity *models.Entity) (string, error) {
	if entity == nil {
		return "", nil
	}
	return security.DecryptIdentityNumber(entity.IdentityNumberCiphertext)
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

// bestEntityNameSimilarity compares the submitted name with the canonical
// name and stored aliases. A containment match is deliberately treated as a
// candidate rather than a confirmed identity; this covers inputs such as a
// shortened company name without turning suffix differences into certainty.
func bestEntityNameSimilarity(input string, entity *models.Entity) float64 {
	if entity == nil {
		return 0
	}
	best := calculateNameSimilarity(input, entity.Name)
	for _, alias := range splitEntityAliases(entity.Alias) {
		if score := calculateNameSimilarity(input, alias); score > best {
			best = score
		}
	}
	return best
}

func splitEntityAliases(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', '，', ';', '；', '/', '、', '|':
			return true
		default:
			return false
		}
	})
}

func normalizeEntityName(value string) []rune {
	normalized := make([]rune, 0, len(value))
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			normalized = append(normalized, r)
		}
	}
	return normalized
}

func calculateNameSimilarity(s1, s2 string) float64 {
	a := normalizeEntityName(s1)
	b := normalizeEntityName(s2)
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	if string(a) == string(b) {
		return 1
	}
	if (len(a) >= 3 && strings.Contains(string(b), string(a))) ||
		(len(b) >= 3 && strings.Contains(string(a), string(b))) {
		return 0.8
	}

	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = minInt(current[j-1]+1, previous[j]+1, previous[j-1]+cost)
		}
		previous, current = current, previous
	}

	maxLength := len(a)
	if len(b) > maxLength {
		maxLength = len(b)
	}
	return 1 - float64(previous[len(b)])/float64(maxLength)
}

func minInt(values ...int) int {
	minimum := values[0]
	for _, value := range values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum
}

func sameIdentityDigest(identityNumber string, entity *models.Entity) bool {
	if strings.TrimSpace(identityNumber) == "" || entity == nil || strings.TrimSpace(entity.IdentityNumberDigest) == "" {
		return false
	}
	digest, err := security.IdentityDigest(identityNumber)
	return err == nil && digest == entity.IdentityNumberDigest
}

// calculateSimilarity remains as a compatibility helper for package-local
// callers; the production path uses canonical names and aliases above.
func (s *conflictCheckService) calculateSimilarity(s1, s2 string) float64 {
	return calculateNameSimilarity(s1, s2)
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
	if check != nil && check.Result != nil && check.Result.CoverageStatus != "" && check.Result.CoverageStatus != "COMPLETE" {
		return "本次利益冲突审查已完成检索，但档案覆盖范围尚未完成律所确认，不能据此认定无冲突或正常受理。"
	}
	if len(conflicts) == 0 {
		return "本次检索未发现可识别匹配记录；结果仍需依据检索范围完成规定的独立复核。"
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
	if check != nil && check.Result != nil && check.Result.CoverageStatus != "" && check.Result.CoverageStatus != "COMPLETE" {
		return "结论：档案覆盖范围尚未完成律所确认，当前结果仅为待复核，不得视为无冲突。"
	}
	if len(conflicts) == 0 {
		return "结论：未发现可识别匹配记录，不构成无冲突或可正常受理的法律结论。"
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
