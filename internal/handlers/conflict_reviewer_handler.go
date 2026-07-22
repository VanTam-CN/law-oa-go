package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// ConflictReviewerHandler exposes the explicit assignment step required
// before a professional conflict conclusion can be submitted.
type ConflictReviewerHandler struct {
	db    *gorm.DB
	authz *services.AuthorizationService
}

func NewConflictReviewerHandler(db *gorm.DB) *ConflictReviewerHandler {
	return &ConflictReviewerHandler{db: db}
}

func (h *ConflictReviewerHandler) SetAuthorizationService(authz *services.AuthorizationService) {
	h.authz = authz
}

func (h *ConflictReviewerHandler) Assign(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	var input services.ConflictReviewerAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		common.APIBadRequest(c, "复核人指定信息无效", err.Error())
		return
	}
	if !h.canAccessConflictTask(c, c.Param("task_id")) {
		return
	}
	assignment, err := services.AssignConflictReviewer(c.Request.Context(), h.db, services.AuthActor{UserID: actor.UserID, Role: actor.Role}, c.Param("task_id"), input)
	if err != nil {
		writeConflictReviewerError(c, err)
		return
	}
	common.APISuccess(c, assignment)
}

func (h *ConflictReviewerHandler) GetAssignment(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权查看复核指定", "只有业务冲突复核角色可以查看复核指定")
		return
	}
	if !h.canAccessConflictTask(c, c.Param("task_id")) {
		return
	}
	assignment, err := services.GetActiveConflictReviewerAssignment(c.Request.Context(), h.db, c.Param("task_id"))
	if err != nil {
		writeConflictReviewerError(c, err)
		return
	}
	common.APISuccess(c, assignment)
}

func (h *ConflictReviewerHandler) canAccessConflictTask(c *gin.Context, taskID string) bool {
	if h.db == nil || h.authz == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "REVIEWER_GATE_UNAVAILABLE", "独立复核对象权限未初始化")
		return false
	}
	var record models.ConflictCheckRecord
	err := h.db.WithContext(c.Request.Context()).Where("check_id = ?", strings.TrimSpace(taskID)).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			common.APINotFound(c, "冲突检测记录不存在", "指定的冲突检测记录不存在")
			return false
		}
		common.APIInternalServerError(c, "读取冲突检测上下文失败", err.Error())
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}
	subject, contextErr := services.ResolveConflictSubjectContext(c.Request.Context(), h.db, &record)
	if contextErr != nil {
		common.APIInternalServerError(c, "读取冲突检测上下文失败", contextErr.Error())
		return false
	}
	if subject.CaseID == 0 && subject.IntakeID == "" {
		common.NewAPIError(c, http.StatusConflict, "CASE_CONTEXT_REQUIRED", "冲突检测缺少可验证的案件或接案上下文，已阻止复核操作")
		return false
	}
	var allowed bool
	var authErr error
	if subject.CaseID > 0 {
		allowed, authErr = h.authz.CanReadConflictContext(c.Request.Context(), actor, subject.CaseID)
	} else {
		allowed, authErr = h.authz.CanReadConflictIntakeContext(c.Request.Context(), actor, subject.IntakeID, record.UserID, record.ClientID)
	}
	if authErr != nil {
		common.APIInternalServerError(c, "案件权限校验失败", authErr.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

func (h *ConflictReviewerHandler) Candidates(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) && !services.IsBusinessMatterManagementRole(actor.Role) {
		common.APIForbidden(c, "无权查看复核人候选名单", "只有冲突核查岗或业务管理合伙人可以指定复核人")
		return
	}
	if h.db == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "REVIEWER_GATE_UNAVAILABLE", "独立复核服务未初始化")
		return
	}
	var users []struct {
		ID         uint   `json:"id"`
		Username   string `json:"username"`
		Name       string `json:"name"`
		Role       string `json:"role"`
		Department string `json:"department"`
	}
	roles := []string{"director", "partner", "compliance", "risk", "risk_control", "management", "conflict_officer"}
	if err := h.db.WithContext(c.Request.Context()).
		Table("users").Select("id, username, name, role, department").
		Where("status = ? AND deleted_at IS NULL AND LOWER(role) IN ?", "active", roles).
		Order("name ASC, id ASC").Find(&users).Error; err != nil {
		common.APIInternalServerError(c, "读取复核人候选名单失败", err.Error())
		return
	}
	common.APISuccess(c, users)
}

func writeConflictReviewerError(c *gin.Context, err error) {
	var reviewerErr *services.ConflictReviewerError
	if !errors.As(err, &reviewerErr) {
		common.APIInternalServerError(c, "处理复核人指定失败", err.Error())
		return
	}
	status := http.StatusConflict
	if strings.HasSuffix(reviewerErr.Code, "FORBIDDEN") || reviewerErr.Code == "REVIEWER_ROLE_FORBIDDEN" || reviewerErr.Code == "REVIEWER_INACTIVE" || reviewerErr.Code == "REVIEWER_NOT_FOUND" {
		status = http.StatusForbidden
	}
	common.NewAPIError(c, status, reviewerErr.Code, reviewerErr.Message)
}
