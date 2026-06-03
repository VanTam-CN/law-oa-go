package handlers

import (
	"fmt"
	"strconv"
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
}

func NewConflictCheckHandler(
	db *gorm.DB,
	conflictCheckService services.ConflictCheckService,
	entityRepo repositories.EntityRepository,
) *ConflictCheckHandler {
	return &ConflictCheckHandler{
		db:                   db,
		conflictCheckService: conflictCheckService,
		entityRepo:           entityRepo,
	}
}

type ConflictCheckCreateRequest struct {
	CaseID        uint                     `json:"caseId" binding:"required"`
	CaseTitle     string                   `json:"caseTitle" binding:"required"`
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

	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权访问")
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
		CaseTitle:     req.CaseTitle,
		CheckEntities: checkEntities,
		SearchDepth:   req.SearchDepth,
		RequestedBy:   userID.(uint),
	}

	response, err := h.conflictCheckService.CheckConflict(c.Request.Context(), checkReq)
	if err != nil {
		common.APIInternalServerError(c, "冲突检测失败: "+err.Error())
		return
	}
	common.APISuccess(c, response)
}

func (h *ConflictCheckHandler) GetConflictCheck(c *gin.Context) {
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
	common.APISuccess(c, check)
}

func (h *ConflictCheckHandler) ListConflictChecks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if caseIDStr := c.Query("caseId"); caseIDStr != "" {
		if caseID, err := strconv.ParseUint(caseIDStr, 10, 32); err == nil {
			filters["case_id"] = uint(caseID)
		}
	}

	checks, total, err := h.conflictCheckService.ListConflictChecks(c.Request.Context(), page, pageSize, filters)
	if err != nil {
		common.APIInternalServerError(c, "获取冲突审查列表失败: "+err.Error())
		return
	}
	common.APISuccessWithPage(c, checks, total, page, pageSize)
}

func (h *ConflictCheckHandler) GenerateReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID")
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }

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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的实体ID")
		return
	}
	depth, _ := strconv.Atoi(c.DefaultQuery("depth", "2"))
	if depth < 1 || depth > 5 { depth = 2 }
	entities, err := h.conflictCheckService.GetEntityRelations(c.Request.Context(), uint(id), depth)
	if err != nil {
		common.APIInternalServerError(c, "获取关联关系失败: "+err.Error())
		return
	}
	common.APISuccess(c, entities)
}

func (h *ConflictCheckHandler) GetConflictCheckStats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 { days = 30 }

	since := time.Now().AddDate(0, 0, -days)

	var totalChecks int64
	h.db.Model(&models.ConflictCheck{}).
		Where("created_at >= ?", since).
		Count(&totalChecks)

	var conflictChecks int64
	h.db.Model(&models.ConflictCheck{}).
		Where("created_at >= ? AND status IN ?", since, []string{"COMPLETED_WITH_CONFLICT", "REJECTED"}).
		Count(&conflictChecks)

	// 按风险等级统计
	type RiskCount struct {
		Status string
		Count  int64
	}
	var riskCounts []RiskCount
	h.db.Model(&models.ConflictCheck{}).
		Select("status, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("status").
		Find(&riskCounts)

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

type BatchImportEntitiesRequest struct {
	Entities []EntityCreateRequest `json:"entities" binding:"required,min=1,max=1000"`
}

func (h *ConflictCheckHandler) BatchImportEntities(c *gin.Context) {
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
	name := c.Query("name")
	if name == "" {
		common.APIBadRequest(c, "名称不能为空")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 { limit = 20 }
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

type AddEntityNameHistoryRequest struct {
	OldName      string    `json:"oldName" binding:"required,max=200"`
	NewName      string    `json:"newName" binding:"required,max=200"`
	ChangeDate   time.Time `json:"changeDate" binding:"required"`
	ChangeReason string    `json:"changeReason" binding:"max=500"`
}

func (h *ConflictCheckHandler) AddEntityNameHistory(c *gin.Context) {
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
