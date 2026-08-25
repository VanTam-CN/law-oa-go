package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/security"
	"law-oa-go/internal/services"
)

// SubjectRecheckHandler exposes the only supported path for changing the
// effective subject set of an existing case.
type SubjectRecheckHandler struct {
	service       *services.SubjectRecheckService
	review        services.ConflictReviewService
	authorization *services.AuthorizationService
	db            *gorm.DB
}

func NewSubjectRecheckHandler(service *services.SubjectRecheckService, review services.ConflictReviewService, authorization *services.AuthorizationService, db *gorm.DB) *SubjectRecheckHandler {
	return &SubjectRecheckHandler{service: service, review: review, authorization: authorization, db: db}
}

// ListSubjectParties returns only the structured subjects attached to a case.
// Identity numbers are never returned; the masked hint only helps a lawyer
// distinguish two records with the same display name.
func (h *SubjectRecheckHandler) ListSubjectParties(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsPrivilegedRole(actor.Role) && !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权查看案件主体冲突信息", "技术管理员不自动获得受保护主体和历史冲突证据访问权")
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, false) {
		return
	}
	if h.db == nil {
		common.APIInternalServerError(c, "读取案件主体失败", "案件主体数据服务未初始化")
		return
	}
	var rows []models.CaseParty
	if err := h.db.WithContext(c.Request.Context()).Preload("Entity").Where("case_id = ? AND deleted_at IS NULL", caseID).Order("display_order ASC").Find(&rows).Error; err != nil {
		common.APIInternalServerError(c, "读取案件主体失败", err.Error())
		return
	}
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		result = append(result, subjectEntityView(row.Entity, string(row.Role), string(row.PartyType)))
	}
	common.APISuccess(c, result)
}

// SearchSubjectEntities is a case-scoped lookup for a lawyer preparing a
// subject revision. It does not expose the firm-wide conflict queue or full
// identity numbers, and it requires ownership of the target case.
func (h *SubjectRecheckHandler) SearchSubjectEntities(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsPrivilegedRole(actor.Role) && !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权搜索案件主体", "技术管理员不自动获得受保护主体和历史冲突证据访问权")
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, true) {
		return
	}
	query := strings.TrimSpace(c.Query("query"))
	if len([]rune(query)) < 2 {
		common.APIBadRequest(c, "搜索条件不足", "请输入至少两个字的主体名称")
		return
	}
	if h.db == nil {
		common.APIInternalServerError(c, "搜索主体失败", "主体数据服务未初始化")
		return
	}
	entities, err := h.service.SearchVisibleEntities(c.Request.Context(), caseID, actor.UserID, actor.Role, query, c.Query("entity_type"), 20)
	if err != nil {
		common.APIInternalServerError(c, "搜索主体失败", err.Error())
		return
	}
	result := make([]gin.H, 0, len(entities))
	for _, entity := range entities {
		result = append(result, subjectEntityView(entity, "", ""))
	}
	common.APISuccess(c, result)
}

func subjectEntityView(entity models.Entity, role, partyType string) gin.H {
	identityNumber := strings.TrimSpace(entity.IdentityNumber)
	if identityNumber == "" && strings.TrimSpace(entity.IdentityNumberCiphertext) != "" {
		identityNumber, _ = security.DecryptIdentityNumber(entity.IdentityNumberCiphertext)
	}
	identityPresent := security.IdentityPresent(entity.IdentityNumber, entity.IdentityNumberCiphertext, entity.IdentityNumberDigest)
	identityHint := security.MaskIdentityNumber(identityNumber)
	if identityHint == "" && identityPresent {
		identityHint = "已登记（受保护）"
	}
	return gin.H{
		"entity_id":        entity.ID,
		"name":             entity.Name,
		"entity_type":      entity.EntityType,
		"role":             role,
		"party_type":       partyType,
		"identity_type":    entity.IdentityType,
		"identity_present": identityPresent,
		"identity_hint":    identityHint,
	}
}

func maskSubjectIdentity(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(runes)-4) + string(runes[len(runes)-4:])
}

func (h *SubjectRecheckHandler) CreateRevision(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsIntakeAssistantRole(actor.Role) {
		common.APIForbidden(c, "无权提交主体变更", "助理只能整理草稿，不能修改正式案件主体")
		return
	}
	if !services.IsBusinessMatterManagementRole(actor.Role) && !strings.EqualFold(strings.TrimSpace(actor.Role), "lawyer") {
		common.APIForbidden(c, "无权提交主体变更", "独立冲突核查人负责复核，不直接修改正式案件主体")
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, true) {
		return
	}
	var request services.SubjectRevisionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.APIBadRequest(c, "主体变更请求无效", err.Error())
		return
	}
	result, err := h.service.CreateRevision(c.Request.Context(), caseID, actor.UserID, actor.Role, &request)
	if err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APISuccess(c, result)
}

func (h *SubjectRecheckHandler) CreateNewEntityRevision(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if services.IsIntakeAssistantRole(actor.Role) {
		common.APIForbidden(c, "无权提交新主体登记", "助理可以整理资料，但身份主体必须由承办律师确认后提交")
		return
	}
	if !services.IsBusinessMatterManagementRole(actor.Role) && !strings.EqualFold(strings.TrimSpace(actor.Role), "lawyer") {
		common.APIForbidden(c, "无权提交新主体登记", "只有案件承办律师或获授权业务负责人可以提交")
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, true) {
		return
	}
	var request services.NewSubjectEntityRevisionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.APIBadRequest(c, "新主体登记请求无效", err.Error())
		return
	}
	result, err := h.service.CreateNewEntityRevision(c.Request.Context(), caseID, actor.UserID, actor.Role, &request)
	if err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APISuccess(c, result)
}

func (h *SubjectRecheckHandler) ReviewEntityRegistration(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	revisionID := strings.TrimSpace(c.Param("revision_id"))
	if revisionID == "" {
		common.APIBadRequest(c, "主体登记复核请求无效", "revision_id不能为空")
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权处理主体登记", "只有独立冲突核查人或获授权业务负责人可以处理")
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, false) {
		return
	}
	var request services.SubjectEntityRegistrationReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.APIBadRequest(c, "主体登记复核请求无效", err.Error())
		return
	}
	result, err := h.service.ReviewEntityRegistration(c.Request.Context(), caseID, revisionID, actor.UserID, actor.Role, &request)
	if err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APISuccess(c, result)
}

func (h *SubjectRecheckHandler) ListPendingEntityRegistrations(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权查看主体登记队列", "只有独立冲突核查人或获授权业务负责人可以查看")
		return
	}
	result, err := h.service.ListPendingEntityRegistrations(c.Request.Context(), actor.UserID, actor.Role)
	if err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APISuccess(c, result)
}

func (h *SubjectRecheckHandler) GetSubjectRevisionStatus(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	revisionID := strings.TrimSpace(c.Param("revision_id"))
	if revisionID == "" {
		common.APIBadRequest(c, "主体变更记录编号无效", "revision_id不能为空")
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, false) {
		return
	}
	result, err := h.service.GetSubjectRevisionStatus(c.Request.Context(), caseID, revisionID)
	if err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APISuccess(c, result)
}

func (h *SubjectRecheckHandler) RunRecheck(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	revisionID := strings.TrimSpace(c.Param("revision_id"))
	if revisionID == "" {
		common.APIBadRequest(c, "主体变更请求无效", "revision_id不能为空")
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsBusinessMatterManagementRole(actor.Role) && !strings.EqualFold(strings.TrimSpace(actor.Role), "lawyer") {
		common.APIForbidden(c, "无权执行主体重检", "独立冲突核查人负责复核，不直接修改正式案件主体")
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, true) {
		return
	}
	var request models.ConflictCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		common.APIBadRequest(c, "重检请求无效", err.Error())
		return
	}
	result, err := h.service.RunRecheck(c.Request.Context(), caseID, revisionID, actor.UserID, actor.Role, &request)
	if err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APISuccess(c, result)
}

func (h *SubjectRecheckHandler) Review(c *gin.Context) {
	caseID, ok := parseSubjectCaseID(c)
	if !ok {
		return
	}
	revisionID := strings.TrimSpace(c.Param("revision_id"))
	if revisionID == "" {
		common.APIBadRequest(c, "主体复核请求无效", "revision_id不能为空")
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权提交独立复核", "只有冲突核查人或获授权管理合伙人可以形成复核结论")
		return
	}
	if !h.authorizeSubjectCase(c, actor, caseID, false) {
		return
	}
	var request struct {
		CheckID  string `json:"check_id"`
		Decision string `json:"decision"`
		Notes    string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		common.APIBadRequest(c, "主体复核请求无效", err.Error())
		return
	}
	checkID := strings.TrimSpace(request.CheckID)
	if checkID == "" {
		common.APIBadRequest(c, "主体复核请求无效", "check_id不能为空")
		return
	}
	result, err := h.service.ReviewAndApply(c.Request.Context(), caseID, revisionID, checkID, actor.UserID, actor.Role, c.GetString("username"), request.Decision, request.Notes, nil)
	if err != nil {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APISuccess(c, result)
}

func parseSubjectCaseID(c *gin.Context) (uint, bool) {
	caseID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || caseID == 0 {
		common.APIBadRequest(c, "案件编号无效", "案件ID必须是正整数")
		return 0, false
	}
	return uint(caseID), true
}

// authorizeSubjectCase keeps subject-version endpoints behind a real case
// authorization dependency. The old handler treated a missing authorization
// service as an implicit allow for lawyers and reviewers, which made a guessed
// case ID enough to enumerate subjects or submit a revision.
func (h *SubjectRecheckHandler) authorizeSubjectCase(c *gin.Context, actor services.AuthActor, caseID uint, manage bool) bool {
	if h.authorization == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "CASE_AUTHZ_UNAVAILABLE", "案件权限服务未初始化，已阻止主体操作")
		return false
	}
	var (
		allowed bool
		err     error
	)
	switch {
	case services.IsConflictReviewRole(actor.Role) && !services.IsBusinessMatterManagementRole(actor.Role):
		allowed, err = h.authorization.CanReadConflictContext(c.Request.Context(), actor, caseID)
	case manage:
		allowed, err = h.authorization.CanManageCase(c.Request.Context(), actor, caseID)
	default:
		allowed, err = h.authorization.CanReadCase(c.Request.Context(), actor, caseID)
	}
	if err != nil {
		common.APIInternalServerError(c, "案件权限检查失败", err.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

func writeSubjectWorkflowError(c *gin.Context, err error) {
	var workflowErr *services.SubjectWorkflowError
	if errors.As(err, &workflowErr) {
		status := http.StatusConflict
		if strings.HasSuffix(workflowErr.Code, "FORBIDDEN") || workflowErr.Code == "REVIEWER_FORBIDDEN" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"success": false, "error": workflowErr.Code, "message": workflowErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "SUBJECT_WORKFLOW_ERROR", "message": err.Error()})
}

func isSubjectWorkflowError(err error) bool {
	var workflowErr *services.SubjectWorkflowError
	return errors.As(err, &workflowErr)
}
