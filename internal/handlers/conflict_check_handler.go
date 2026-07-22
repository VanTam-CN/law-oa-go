package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConflictCheckHandler struct {
	db                   *gorm.DB
	conflictCheckService services.ConflictCheckService
	entityRepo           repositories.EntityRepository
	authz                *services.AuthorizationService
}

func NewConflictCheckHandler(
	db *gorm.DB,
	conflictCheckService services.ConflictCheckService,
	entityRepo repositories.EntityRepository,
	authz ...*services.AuthorizationService,
) *ConflictCheckHandler {
	var authorizationService *services.AuthorizationService
	if len(authz) > 0 {
		authorizationService = authz[0]
	}
	return &ConflictCheckHandler{
		db:                   db,
		conflictCheckService: conflictCheckService,
		entityRepo:           entityRepo,
		authz:                authorizationService,
	}
}

type ConflictCheckCreateRequest struct {
	CaseID        uint                     `json:"caseId" binding:"required"`
	CaseTitle     string                   `json:"caseTitle"`
	CheckEntities []EntityCheckRequestItem `json:"checkEntities" binding:"required,min=1"`
	SearchDepth   int                      `json:"searchDepth" binding:"min=1,max=5"`
}

type EntityCheckRequestItem struct {
	EntityID       uint   `json:"entityId"`
	EntityName     string `json:"entityName" binding:"required"`
	IdentityType   string `json:"identityType"`
	IdentityNumber string `json:"identityNumber"`
	PartyType      string `json:"partyType" binding:"required,oneof=CLIENT OPPOSING CO_DEFENDANT CO_PLAINTIFF"`
}

func (h *ConflictCheckHandler) CreateConflictCheck(c *gin.Context) {
	var req ConflictCheckCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "参数验证失败", map[string]string{"error": err.Error()})
		return
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !h.requireAuthorization(c) {
		return
	}
	if !canCreateLegacyConflictCheck(actor.Role) {
		common.APIForbidden(c, "无权运行最终冲突检查", "助理只能整理待确认草稿，不能运行最终冲突检查")
		return
	}
	allowed, err := h.authz.CanReadConflictContext(c.Request.Context(), actor, req.CaseID)
	if err != nil {
		common.APIInternalServerError(c, "权限校验失败: "+err.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	// The case record is the source of truth. Do not allow callers to inject a
	// misleading title into the persisted conflict-check evidence.
	var caseRecord models.Case
	if err := h.db.WithContext(c.Request.Context()).Select("id", "title").First(&caseRecord, req.CaseID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			common.APINotFound(c, "案件不存在")
			return
		}
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_CASE_UNAVAILABLE", "案件上下文暂不可用，已阻止冲突检测")
		return
	}
	if strings.TrimSpace(caseRecord.Title) == "" {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_CASE_UNAVAILABLE", "案件缺少可追溯标题，已阻止冲突检测")
		return
	}

	checkEntities := make([]services.EntityCheckInfo, 0, len(req.CheckEntities))
	for _, entity := range req.CheckEntities {
		checkEntities = append(checkEntities, services.EntityCheckInfo{
			EntityID:       entity.EntityID,
			EntityName:     entity.EntityName,
			IdentityType:   entity.IdentityType,
			IdentityNumber: entity.IdentityNumber,
			PartyType:      entity.PartyType,
		})
	}

	checkReq := &services.ConflictCheckRequest{
		CaseID:        req.CaseID,
		CaseTitle:     caseRecord.Title,
		CheckEntities: checkEntities,
		SearchDepth:   req.SearchDepth,
		RequestedBy:   actor.UserID,
	}

	response, err := h.conflictCheckService.CheckConflict(c.Request.Context(), checkReq)
	if err != nil {
		common.APIInternalServerError(c, "冲突检测失败: "+err.Error())
		return
	}
	projectLegacyConflictResponseForViewer(c, response)
	common.APISuccess(c, response)
}

func canCreateLegacyConflictCheck(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "lawyer") || services.IsConflictReviewRole(role)
}

func (h *ConflictCheckHandler) GetConflictCheck(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID")
		return
	}
	check, err := h.conflictCheckService.GetConflictCheck(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "冲突审查记录不存在")
		return
	}
	if !h.canAccessConflictCheck(c, check) {
		return
	}
	if actor, ok := currentAuthActor(c); ok && !services.IsConflictReviewRole(actor.Role) {
		projectLegacyConflictCheckForViewer(check)
	}
	common.APISuccess(c, check)
}

func (h *ConflictCheckHandler) ListConflictChecks(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := make(map[string]interface{})
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if caseIDStr := c.Query("caseId"); caseIDStr != "" {
		caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
		if err != nil || caseID == 0 {
			common.APIBadRequest(c, "无效的案件ID")
			return
		}
		allowed, err := h.authz.CanReadConflictContext(c.Request.Context(), actor, uint(caseID))
		if err != nil {
			common.APIInternalServerError(c, "权限校验失败: "+err.Error())
			return
		}
		if !allowed {
			forbidObjectAccess(c)
			return
		}
		filters["case_id"] = uint(caseID)
	}
	// The repository applies this as a SQL predicate so pagination and total
	// counts cannot enumerate wall-protected conflict checks.
	filters["ethical_wall_user_id"] = actor.UserID
	if !services.IsConflictReviewRole(actor.Role) {
		filters["requested_by"] = actor.UserID
	}

	checks, total, err := h.conflictCheckService.ListConflictChecks(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		common.APIInternalServerError(c, "获取冲突审查列表失败: "+err.Error())
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		for _, check := range checks {
			projectLegacyConflictCheckForViewer(check)
		}
	}
	common.APISuccessWithPage(c, checks, total, page, pageSize)
}

func (h *ConflictCheckHandler) GenerateReport(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID")
		return
	}
	check, err := h.conflictCheckService.GetConflictCheck(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "冲突审查记录不存在")
		return
	}
	if !h.canAccessConflictCheck(c, check) {
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权生成完整冲突报告", "完整历史证据报告仅向独立冲突核查人开放")
		return
	}
	report, err := h.conflictCheckService.GenerateReport(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "生成报告失败: "+err.Error())
		return
	}
	common.APISuccess(c, report)
}

func (h *ConflictCheckHandler) SearchEntity(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	query := c.Query("query")
	if query == "" {
		common.APIBadRequest(c, "搜索关键词不能为空")
		return
	}
	entityType := c.Query("entityType")
	entities, err := h.conflictCheckService.SearchEntity(c.Request.Context(), query, entityType)
	if err != nil {
		common.APIInternalServerError(c, "搜索实体失败: "+err.Error())
		return
	}
	common.APISuccess(c, entities)
}

type EntityCreateRequest struct {
	EntityType          string  `json:"entityType" binding:"required,oneof=INDIVIDUAL LEGAL_PERSON ORGANIZATION"`
	Name                string  `json:"name" binding:"required,max=200"`
	Alias               string  `json:"alias" binding:"max=500"`
	IdentityType        string  `json:"identityType" binding:"required,oneof=ID_CARD PASSPORT BUSINESS_LICENSE ORGANIZATION_CODE SOCIAL_CREDIT_CODE OTHER"`
	IdentityNumber      string  `json:"identityNumber" binding:"max=100"`
	Gender              string  `json:"gender" binding:"omitempty,oneof=MALE FEMALE OTHER"`
	Nationality         string  `json:"nationality" binding:"max=50"`
	LegalRepresentative string  `json:"legalRepresentative" binding:"max=100"`
	RegisteredCapital   float64 `json:"registeredCapital"`
	Address             string  `json:"address"`
	Phone               string  `json:"phone" binding:"max=20"`
	Email               string  `json:"email" binding:"max=100,email"`
	ContactPerson       string  `json:"contactPerson" binding:"max=100"`
	Notes               string  `json:"notes"`
}

func (h *ConflictCheckHandler) CreateEntity(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	var req EntityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "参数验证失败", map[string]string{"error": err.Error()})
		return
	}
	entity := &models.Entity{
		EntityType:          models.EntityType(req.EntityType),
		Name:                req.Name,
		Alias:               req.Alias,
		IdentityType:        models.IdentityType(req.IdentityType),
		IdentityNumber:      req.IdentityNumber,
		Status:              models.EntityStatusActive,
		Gender:              models.Gender(req.Gender),
		Nationality:         req.Nationality,
		LegalRepresentative: req.LegalRepresentative,
		RegisteredCapital:   req.RegisteredCapital,
		Address:             req.Address,
		Phone:               req.Phone,
		Email:               req.Email,
		ContactPerson:       req.ContactPerson,
		Notes:               req.Notes,
	}
	if err := h.entityRepo.Create(c.Request.Context(), entity); err != nil {
		common.APIInternalServerError(c, "创建实体失败: "+err.Error())
		return
	}
	common.APISuccess(c, entity)
}

func (h *ConflictCheckHandler) GetEntity(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的实体ID")
		return
	}
	entity, err := h.entityRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "实体不存在")
		return
	}
	common.APISuccess(c, entity)
}

func (h *ConflictCheckHandler) ListEntities(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	conditions := make(map[string]interface{})
	if entityType := c.Query("entityType"); entityType != "" {
		conditions["entity_type"] = entityType
	}
	if status := c.Query("status"); status != "" {
		conditions["status"] = status
	}
	entities, total, err := h.entityRepo.List(c.Request.Context(), (page-1)*pageSize, pageSize, conditions)
	if err != nil {
		common.APIInternalServerError(c, "获取实体列表失败: "+err.Error())
		return
	}
	common.APISuccessWithPage(c, entities, total, page, pageSize)
}

func (h *ConflictCheckHandler) GetEntityRelations(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的实体ID")
		return
	}
	depth, _ := strconv.Atoi(c.DefaultQuery("depth", "2"))
	if depth < 1 || depth > 5 {
		depth = 2
	}
	entities, err := h.conflictCheckService.GetEntityRelations(c.Request.Context(), uint(id), depth)
	if err != nil {
		common.APIInternalServerError(c, "获取关联关系失败: "+err.Error())
		return
	}
	common.APISuccess(c, entities)
}

func (h *ConflictCheckHandler) GetConflictCheckStats(c *gin.Context) {
	if !h.requireAuthorization(c) {
		return
	}
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 {
		days = 30
	}

	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}

	since := time.Now().AddDate(0, 0, -days)

	var totalChecks int64
	totalQuery := h.db.Model(&models.ConflictCheck{}).Where("created_at >= ?", since)
	totalQuery = h.applyEthicalWallScope(totalQuery, actor.UserID)
	if !services.IsConflictReviewRole(actor.Role) {
		totalQuery = totalQuery.Where("requested_by = ?", actor.UserID)
	}
	if err := totalQuery.Count(&totalChecks).Error; err != nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_STATS_UNAVAILABLE", "冲突检测统计暂不可用")
		return
	}

	var conflictChecks int64
	conflictQuery := h.db.Model(&models.ConflictCheck{}).
		Where("created_at >= ? AND status IN ?", since, []string{"COMPLETED_WITH_CONFLICT", "REJECTED"})
	conflictQuery = h.applyEthicalWallScope(conflictQuery, actor.UserID)
	if !services.IsConflictReviewRole(actor.Role) {
		conflictQuery = conflictQuery.Where("requested_by = ?", actor.UserID)
	}
	if err := conflictQuery.Count(&conflictChecks).Error; err != nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_STATS_UNAVAILABLE", "冲突检测统计暂不可用")
		return
	}

	// 按风险等级统计
	type RiskCount struct {
		Status string
		Count  int64
	}
	var riskCounts []RiskCount
	riskQuery := h.db.Model(&models.ConflictCheck{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("status")
	riskQuery = h.applyEthicalWallScope(riskQuery, actor.UserID)
	if !services.IsConflictReviewRole(actor.Role) {
		riskQuery = riskQuery.Where("requested_by = ?", actor.UserID)
	}
	if err := riskQuery.Find(&riskCounts).Error; err != nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_STATS_UNAVAILABLE", "冲突检测统计暂不可用")
		return
	}

	topConflictTypes := make(map[string]int64)
	for _, rc := range riskCounts {
		topConflictTypes[rc.Status] = rc.Count
	}

	stats := map[string]interface{}{
		"totalChecks":      totalChecks,
		"conflictChecks":   conflictChecks,
		"topConflictTypes": topConflictTypes,
		"period":           fmt.Sprintf("最近%d天", days),
	}
	common.APISuccess(c, stats)
}

func (h *ConflictCheckHandler) applyEthicalWallScope(query *gorm.DB, userID uint) *gorm.DB {
	if userID == 0 {
		return query
	}
	return query.Where(`conflict_checks.case_id IN (
		SELECT visible_cases.id FROM cases visible_cases
		WHERE visible_cases.deleted_at IS NULL
		  AND (
			  visible_cases.ethical_wall_enabled = ?
			  OR EXISTS (
				  SELECT 1 FROM case_ethical_wall_whitelist wall_access
				  WHERE wall_access.case_id = visible_cases.id
					AND wall_access.user_id = ?
			  )
		  )
	)`, false, userID)
}

type BatchImportEntitiesRequest struct {
	Entities []EntityCreateRequest `json:"entities" binding:"required,min=1,max=1000"`
}

func (h *ConflictCheckHandler) BatchImportEntities(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	var req BatchImportEntitiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "参数验证失败", map[string]string{"error": err.Error()})
		return
	}
	entities := make([]*models.Entity, 0, len(req.Entities))
	for _, item := range req.Entities {
		entity := &models.Entity{
			EntityType:          models.EntityType(item.EntityType),
			Name:                item.Name,
			Alias:               item.Alias,
			IdentityType:        models.IdentityType(item.IdentityType),
			IdentityNumber:      item.IdentityNumber,
			Status:              models.EntityStatusActive,
			Gender:              models.Gender(item.Gender),
			Nationality:         item.Nationality,
			LegalRepresentative: item.LegalRepresentative,
			RegisteredCapital:   item.RegisteredCapital,
			Address:             item.Address,
			Phone:               item.Phone,
			Email:               item.Email,
			ContactPerson:       item.ContactPerson,
			Notes:               item.Notes,
		}
		entities = append(entities, entity)
	}
	if err := h.entityRepo.BatchCreate(c.Request.Context(), entities); err != nil {
		common.APIInternalServerError(c, "批量导入实体失败: "+err.Error())
		return
	}
	common.APISuccess(c, gin.H{
		"imported": len(entities),
		"message":  fmt.Sprintf("成功导入 %d 个实体", len(entities)),
	})
}

func (h *ConflictCheckHandler) GetEntityByName(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	name := c.Query("name")
	if name == "" {
		common.APIBadRequest(c, "名称不能为空")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	entities, err := h.entityRepo.SearchByName(c.Request.Context(), name, limit)
	if err != nil {
		common.APIInternalServerError(c, "搜索实体失败: "+err.Error())
		return
	}
	common.APISuccess(c, entities)
}

type AddEntityRelationRequest struct {
	TargetEntityID    uint    `json:"targetEntityId" binding:"required"`
	RelationType      string  `json:"relationType" binding:"required,oneof=PARENT_COMPANY SUBSIDIARY ACTUAL_CONTROLLER MAJOR_SHAREHOLDER NOMINEE_SHAREHOLDER BRANCH JOINT_VENTURE RELATED_PARTY FAMILY_MEMBER SPOUSE"`
	ShareholdingRatio float64 `json:"shareholdingRatio" binding:"min=0,max=100"`
	Description       string  `json:"description"`
}

func (h *ConflictCheckHandler) AddEntityRelation(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	idStr := c.Param("id")
	sourceEntityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的源实体ID")
		return
	}
	var req AddEntityRelationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "参数验证失败", map[string]string{"error": err.Error()})
		return
	}
	now := time.Now()
	relation := &models.EntityRelation{
		SourceEntityID:    uint(sourceEntityID),
		TargetEntityID:    req.TargetEntityID,
		RelationType:      models.RelationType(req.RelationType),
		ShareholdingRatio: req.ShareholdingRatio,
		Description:       req.Description,
		IsActive:          true,
		StartDate:         &now,
	}
	if err := h.entityRepo.BatchCreateRelations(c.Request.Context(), []*models.EntityRelation{relation}); err != nil {
		common.APIInternalServerError(c, "添加关联关系失败: "+err.Error())
		return
	}
	common.APISuccess(c, relation)
}

func (h *ConflictCheckHandler) GetEntityNameHistory(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的实体ID")
		return
	}
	history, err := h.entityRepo.GetNameHistory(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取名称历史失败: "+err.Error())
		return
	}
	common.APISuccess(c, history)
}

func (h *ConflictCheckHandler) canAccessConflictCheck(c *gin.Context, check *models.ConflictCheck) bool {
	if check == nil {
		common.APINotFound(c, "冲突审查记录不存在")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	if h.authz == nil {
		h.requireAuthorization(c)
		return false
	}
	allowed, err := h.authz.CanReadConflictContext(c.Request.Context(), actor, check.CaseID)
	if err != nil {
		common.APIInternalServerError(c, "权限校验失败: "+err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

func (h *ConflictCheckHandler) requireAuthorization(c *gin.Context) bool {
	if h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_AUTHZ_UNAVAILABLE", "冲突检查权限服务未初始化，当前已阻止冲突数据操作")
		return false
	}
	return true
}

// projectLegacyConflictCheckForViewer protects the compatibility conflict-v2
// record API. The current task owner may see the task and aggregate count, but
// only an independent conflict reviewer may see historical matter details.
func projectLegacyConflictCheckForViewer(check *models.ConflictCheck) {
	if check == nil {
		return
	}
	check.Status = "REVIEW_REQUIRED"
	check.ResultSummary = "检测已完成，但仍需独立冲突核查人确认。"
	check.CriticalCount = 0
	check.HighCount = 0
	check.MediumCount = 0
	check.LowCount = 0
	check.CheckParams = nil
	check.ReportPath = ""
	check.ReportGeneratedAt = nil
	if check.Result == nil {
		check.Result = &models.CheckResult{}
	}
	coverageIncomplete := !strings.EqualFold(strings.TrimSpace(check.Result.CoverageStatus), "COMPLETE")
	check.Result.HasConflict = check.TotalConflicts > 0 || coverageIncomplete
	check.Result.CoverageNotice = "检测结果需要独立冲突核查人确认，不能据此认定无冲突。"
	if check.TotalConflicts > 0 {
		check.ResultSummary = "发现历史匹配记录，但详情仅向独立冲突核查人披露。"
	}
	for index := range check.ConflictDetails {
		detail := &check.ConflictDetails[index]
		detail.MatchedEntityID = 0
		detail.MatchedEntity = models.Entity{}
		detail.MatchedCaseID = nil
		detail.MatchedCase = models.Case{}
		detail.ConflictType = "受限历史记录"
		detail.RiskLevel = ""
		detail.Description = "存在受隔离记录，请联系独立冲突核查人。"
		detail.Evidence = ""
		detail.Recommendation = "请联系独立冲突核查人。"
		detail.MatchReason = ""
		detail.IsWaived = false
		detail.WaivedBy = nil
		detail.WaivedAt = nil
		detail.WaiveReason = ""
	}
}

// projectLegacyConflictResponseForViewer protects the compatibility create
// endpoint before it serializes the in-memory result to the caller.
func projectLegacyConflictResponseForViewer(c *gin.Context, response *services.ConflictCheckResponse) {
	if response == nil || services.IsConflictReviewRole(c.GetString("role")) {
		return
	}
	response.Status = "REVIEW_REQUIRED"
	coverageIncomplete := !strings.EqualFold(strings.TrimSpace(response.CoverageStatus), "COMPLETE")
	response.HasConflict = response.TotalConflicts > 0 || coverageIncomplete
	response.CoverageNotice = "检测结果需要独立冲突核查人确认，不能据此认定无冲突。"
	recommendation := "检索已完成，但仍需独立冲突核查人确认档案覆盖和主体信息完整性。"
	if response.TotalConflicts > 0 {
		recommendation = "存在历史匹配记录，详情仅向独立冲突核查人披露。"
	}
	totalChecked := 0
	if response.Summary != nil {
		totalChecked = response.Summary.TotalChecked
	}
	response.Summary = &services.ConflictSummary{
		TotalChecked:     totalChecked,
		ConflictCount:    response.TotalConflicts,
		RiskDistribution: map[string]int{},
		ConflictTypes:    map[string]int{"REVIEW_REQUIRED": response.TotalConflicts},
		Recommendations:  []string{recommendation},
	}
	for _, conflict := range response.Conflicts {
		if conflict == nil {
			continue
		}
		conflict.ConflictType = "受限历史记录"
		conflict.RiskLevel = ""
		conflict.Description = "存在受隔离记录，请联系独立冲突核查人。"
		conflict.MatchedEntity = nil
		conflict.RelatedCase = nil
		conflict.MatchReason = ""
		conflict.Recommendation = "请联系独立冲突核查人。"
	}
}

type AddEntityNameHistoryRequest struct {
	OldName      string    `json:"oldName" binding:"required,max=200"`
	NewName      string    `json:"newName" binding:"required,max=200"`
	ChangeDate   time.Time `json:"changeDate" binding:"required"`
	ChangeReason string    `json:"changeReason" binding:"max=500"`
}

func (h *ConflictCheckHandler) AddEntityNameHistory(c *gin.Context) {
	if !h.requireConflictRegistryManager(c) {
		return
	}
	idStr := c.Param("id")
	entityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的实体ID")
		return
	}
	var req AddEntityNameHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIValidationError(c, "参数验证失败", map[string]string{"error": err.Error()})
		return
	}
	history := &models.EntityNameHistory{
		EntityID:     uint(entityID),
		OldName:      req.OldName,
		NewName:      req.NewName,
		ChangeDate:   req.ChangeDate,
		ChangeReason: req.ChangeReason,
	}
	if err := h.entityRepo.AddNameHistory(c.Request.Context(), history); err != nil {
		common.APIInternalServerError(c, "添加名称历史失败: "+err.Error())
		return
	}
	common.APISuccess(c, history)
}

// requireConflictRegistryManager protects the firm-wide subject registry.
// Running a conflict check is a normal lawyer workflow, but browsing or
// editing the underlying registry exposes historical identities, aliases and
// relationships across the firm. That access belongs to the independent
// conflict-review function or an explicitly authorized business manager.
func (h *ConflictCheckHandler) requireConflictRegistryManager(c *gin.Context) bool {
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权访问主体库", "主体库详情和维护仅限独立冲突核查或获授权业务管理角色")
		return false
	}
	return true
}
