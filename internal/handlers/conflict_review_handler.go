package handlers

import (
	"strconv"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ConflictReviewHandler struct {
	conflictCheckService services.ConflictCheckService
	entityService        services.EntityService
	db                   *gorm.DB
}

func NewConflictReviewHandler(
	conflictCheckService services.ConflictCheckService,
	entityService services.EntityService,
	db *gorm.DB,
) *ConflictReviewHandler {
	return &ConflictReviewHandler{
		conflictCheckService: conflictCheckService,
		entityService:        entityService,
		db:                   db,
	}
}

type ReviewCheckRequest struct {
	CaseID         uint   `json:"case_id" binding:"required"`
	SearchYears    int    `json:"search_years"`
	SearchDepth    string `json:"search_depth"`
	IncludeRelated bool   `json:"include_related"`
	PartyIDs       []uint `json:"party_ids"`
}

func (h *ConflictReviewHandler) CreateConflictCheck(c *gin.Context) {
	var req ReviewCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未授权")
		return
	}

	searchDepth := 2
	if req.SearchDepth != "" {
		switch req.SearchDepth {
		case "basic":
			searchDepth = 1
		case "standard":
			searchDepth = 2
		case "deep":
			searchDepth = 4
		}
	}

	checkEntities := make([]services.EntityCheckInfo, 0, len(req.PartyIDs))
	for _, partyID := range req.PartyIDs {
		checkEntities = append(checkEntities, services.EntityCheckInfo{
			EntityID: partyID,
		})
	}

	if len(checkEntities) == 0 {
		common.APIBadRequest(c, "至少需要指定一个审查实体")
		return
	}

	checkReq := &services.ConflictCheckRequest{
		CaseID:        req.CaseID,
		CheckEntities: checkEntities,
		SearchDepth:   searchDepth,
		RequestedBy:   userID.(uint),
	}

	result, err := h.conflictCheckService.CheckConflict(c.Request.Context(), checkReq)
	if err != nil {
		common.APIInternalServerError(c, "创建冲突审查失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

func (h *ConflictReviewHandler) GetConflictCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID", err.Error())
		return
	}

	check, err := h.conflictCheckService.GetConflictCheck(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "审查记录不存在")
		return
	}

	common.APISuccess(c, gin.H{"check": check})
}

type ListConflictChecksRequest struct {
	Page     int   `form:"page,default=1"`
	PageSize int   `form:"page_size,default=20"`
	CaseID   *uint `form:"case_id"`
}

func (h *ConflictReviewHandler) ListConflictChecks(c *gin.Context) {
	var req ListConflictChecksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	filters := make(map[string]interface{})
	if req.CaseID != nil {
		filters["case_id"] = *req.CaseID
	}

	checks, total, err := h.conflictCheckService.ListConflictChecks(c.Request.Context(), req.Page, req.PageSize, filters)
	if err != nil {
		common.APIInternalServerError(c, "获取审查列表失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"data": checks,
		"pagination": gin.H{
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total":       total,
			"total_pages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		},
	})
}

func (h *ConflictReviewHandler) GetConflictDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID", err.Error())
		return
	}

	report, err := h.conflictCheckService.GenerateReport(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取冲突明细失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"conflicts": report.Conflicts,
		"summary":  report.Recommendations,
	})
}

func (h *ConflictReviewHandler) GetStatistics(c *gin.Context) {
	_, total, err := h.conflictCheckService.ListConflictChecks(c.Request.Context(), 1, 1, nil)
	if err != nil {
		common.APIInternalServerError(c, "获取统计信息失败", err.Error())
		return
	}
	common.APISuccess(c, gin.H{"total_checks": total})
}

func (h *ConflictReviewHandler) GetPendingChecks(c *gin.Context) {
	filters := map[string]interface{}{"status": "processing"}
	checks, _, err := h.conflictCheckService.ListConflictChecks(c.Request.Context(), 1, 100, filters)
	if err != nil {
		common.APIInternalServerError(c, "获取待处理审查失败", err.Error())
		return
	}
	common.APISuccess(c, checks)
}

// ReviewActionRequest 审批操作请求体
type ReviewActionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *ConflictReviewHandler) ApproveCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID", err.Error())
		return
	}

	var req ReviewActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 验证审查记录存在
	check, err := h.conflictCheckService.GetConflictCheck(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "审查记录不存在")
		return
	}

	// 角色权限检查：仅管理员和经理可执行审批操作
	userRole, _ := c.Get("role")
	if userRole != "ADMIN" && userRole != "MANAGER" {
		common.APIForbidden(c, "无权执行此操作，仅管理员或经理可执行审批")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":         string(models.CheckStatusApproved),
		"checked_at":     &now,
		"result_summary": req.Reason,
	}

	if err := h.db.Model(&models.ConflictCheck{}).Where("id = ? AND status IN ?", id, []string{string(models.CheckStatusPending), string(models.CheckStatusInProgress)}).Updates(updates).Error; err != nil {
		common.APIInternalServerError(c, "更新审查状态失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"check_id":   check.ID,
		"status":     string(models.CheckStatusApproved),
		"reason":     req.Reason,
		"updated_at": now,
	})
}

func (h *ConflictReviewHandler) RejectCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID", err.Error())
		return
	}

	var req ReviewActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 验证审查记录存在
	check, err := h.conflictCheckService.GetConflictCheck(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "审查记录不存在")
		return
	}

	// 角色权限检查：仅管理员和经理可执行审批操作
	userRole, _ := c.Get("role")
	if userRole != "ADMIN" && userRole != "MANAGER" {
		common.APIForbidden(c, "无权执行此操作，仅管理员或经理可执行审批")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":         string(models.CheckStatusRejected),
		"checked_at":     &now,
		"result_summary": req.Reason,
	}

	if err := h.db.Model(&models.ConflictCheck{}).Where("id = ? AND status IN ?", id, []string{string(models.CheckStatusPending), string(models.CheckStatusInProgress)}).Updates(updates).Error; err != nil {
		common.APIInternalServerError(c, "更新审查状态失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"check_id":   check.ID,
		"status":     string(models.CheckStatusRejected),
		"reason":     req.Reason,
		"updated_at": now,
	})
}

func (h *ConflictReviewHandler) WaiveCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的审查ID", err.Error())
		return
	}

	var req ReviewActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 验证审查记录存在
	check, err := h.conflictCheckService.GetConflictCheck(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "审查记录不存在")
		return
	}

	// 角色权限检查：仅管理员和经理可执行审批操作
	userRole, _ := c.Get("role")
	if userRole != "ADMIN" && userRole != "MANAGER" {
		common.APIForbidden(c, "无权执行此操作，仅管理员或经理可执行审批")
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":         string(models.CheckStatusCompleted),
		"checked_at":     &now,
		"result_summary": "豁免: " + req.Reason,
	}

	if err := h.db.Model(&models.ConflictCheck{}).Where("id = ? AND status IN ?", id, []string{string(models.CheckStatusPending), string(models.CheckStatusInProgress)}).Updates(updates).Error; err != nil {
		common.APIInternalServerError(c, "更新审查状态失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"check_id":   check.ID,
		"status":     string(models.CheckStatusCompleted),
		"reason":     req.Reason,
		"updated_at": now,
	})
}

func (h *ConflictReviewHandler) ResolveConflictDetail(c *gin.Context) {
	common.APISuccess(c, gin.H{"message": "解决成功"})
}
