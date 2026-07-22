package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// authorizeApprovalConflictContext verifies the case subject behind a
// controlled approval before its detail or decision path is used. Approval
// applicant/approver IDs are not enough: a stale assignment must not bypass
// the ethical wall around the underlying case.
func (h *ApprovalHandler) authorizeApprovalConflictContext(c *gin.Context, approval *models.ApprovalRequest) bool {
	if approval == nil || !isConflictBoundApproval(approval) {
		return true
	}
	if h.authorization == nil || h.db == nil || !h.db.Migrator().HasTable("conflict_check_records") {
		common.NewAPIError(c, http.StatusServiceUnavailable, "APPROVAL_AUTHZ_UNAVAILABLE", "受控审批的案件权限服务未初始化，已阻止访问")
		return false
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return false
	}

	checkID := approvalConflictCheckID(approval)
	if checkID == "" {
		common.NewAPIError(c, http.StatusServiceUnavailable, "APPROVAL_CONTEXT_UNAVAILABLE", "审批缺少可验证的冲突检测上下文，已阻止访问")
		return false
	}

	var record models.ConflictCheckRecord
	if err := h.db.Table("conflict_check_records").
		Select("check_id, search_parameters, client_id, user_id").
		Where("check_id = ?", checkID).
		Take(&record).Error; err != nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "APPROVAL_CONTEXT_UNAVAILABLE", "审批关联的冲突检测记录不可用，已阻止访问")
		return false
	}

	subject, contextErr := services.ResolveConflictSubjectContext(c.Request.Context(), h.db, &record)
	if contextErr != nil {
		common.APIInternalServerError(c, "审批上下文解析失败", contextErr.Error())
		return false
	}
	if subject.CaseID == 0 && subject.IntakeID == "" {
		common.NewAPIError(c, http.StatusServiceUnavailable, "APPROVAL_CONTEXT_UNAVAILABLE", "审批缺少唯一案件或接案上下文，已阻止访问")
		return false
	}
	var allowed bool
	var authErr error
	if subject.CaseID > 0 {
		allowed, authErr = h.authorization.CanReadConflictContext(c.Request.Context(), actor, subject.CaseID)
	} else {
		allowed, authErr = h.authorization.CanReadConflictIntakeContext(c.Request.Context(), actor, subject.IntakeID, record.UserID, record.ClientID)
	}
	if authErr != nil {
		common.APIInternalServerError(c, "审批权限校验失败", authErr.Error())
		return false
	}
	if !allowed {
		forbidObjectAccess(c)
		return false
	}
	return true
}

func isConflictBoundApproval(approval *models.ApprovalRequest) bool {
	if approval == nil {
		return false
	}
	typeName := strings.ToLower(strings.TrimSpace(approval.Type))
	if typeName == "conflict" || typeName == "conflict_approval" || typeName == "case_creation" || typeName == "waiver" {
		return true
	}
	return approvalConflictCheckID(approval) != ""
}

func approvalConflictCheckID(approval *models.ApprovalRequest) string {
	if approval == nil {
		return ""
	}
	if checkID := strings.TrimSpace(approval.ConflictCheckID); checkID != "" {
		return checkID
	}
	metadata := integrationJSONMap(approval.Metadata)
	for _, key := range []string{"conflict_task_id", "conflict_check_id"} {
		if checkID := integrationMetadataString(metadata, key); checkID != "" && checkID != "<nil>" {
			return checkID
		}
	}
	return ""
}
