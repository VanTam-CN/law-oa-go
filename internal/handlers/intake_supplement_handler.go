package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

func writeIntakeCollaborationError(c *gin.Context, err error) {
	if writeKnownIntakeCollaborationError(c, err) {
		return
	}
	var gate *services.ApprovalConflictGateError
	if errors.As(err, &gate) {
		common.NewAPIError(c, http.StatusConflict, gate.Code, gate.Message)
		return
	}
	if errors.Is(err, services.ErrApprovalForbidden) {
		forbidObjectAccess(c)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		common.APINotFound(c, "业务记录不存在", "请刷新后重试")
		return
	}
	if isSubjectWorkflowError(err) {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.APIInternalServerError(c, "处理接案补充失败", err.Error())
}

// Reply edits have a narrow contract. Mixing facts into the same request is
// rejected, so preserving a conflict result cannot hide a factual change.
func (h *DemoAggregateHandler) saveIntakeSupplement(c *gin.Context, intakeID, userID string, payload map[string]interface{}) {
	allowed := map[string]bool{"approval_id": true, "return_record_id": true, "expected_revision": true, "supplement_responses": true}
	for key := range payload {
		if !allowed[key] {
			common.APIBadRequest(c, "补充回复包含不允许的字段", "请先保存接案事实，再单独保存补充回复")
			return
		}
	}
	if isAssistantRole(c) {
		forbidObjectAccess(c)
		return
	}
	var intake struct {
		CreatedBy string
		ClientID  uint
	}
	if err := h.db.Table("case_intakes").Where("id = ?", intakeID).Take(&intake).Error; err != nil {
		common.APINotFound(c, "接案不存在", "指定接案不存在")
		return
	}
	if intake.CreatedBy != userID {
		forbidObjectAccess(c)
		return
	}
	if !h.authorizeIntakeClient(c, intake.ClientID) {
		return
	}
	var req struct {
		ApprovalID       string                      `json:"approval_id"`
		ReturnRecordID   string                      `json:"return_record_id"`
		ExpectedRevision string                      `json:"expected_revision"`
		Responses        []models.SupplementResponse `json:"supplement_responses"`
	}
	raw, err := json.Marshal(payload)
	if err != nil || json.Unmarshal(raw, &req) != nil {
		common.APIBadRequest(c, "补充回复格式错误", "请重新填写")
		return
	}
	var revision string
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var err error
		revision, err = services.SaveIntakeSupplementResponsesInTx(tx, intakeID, req.ApprovalID, userID, req.ExpectedRevision, req.ReturnRecordID, req.Responses)
		return err
	})
	if err != nil {
		writeIntakeCollaborationError(c, err)
		return
	}
	common.APISuccess(c, gin.H{"id": intakeID, "revision": revision, "return_record_id": req.ReturnRecordID, "supplement_responses": req.Responses})
}

func rejectControlledIntakeMetadata(c *gin.Context, value interface{}) bool {
	for key := range objectValue(value) {
		key = strings.ToLower(key)
		if key == "draft_revision" || key == "return_record_id" || key == "supplement_responses" || key == "supplement_return_record_id" || key == "active_submission_snapshot_id" || strings.HasPrefix(key, "lawyer_facts_") || strings.HasPrefix(key, "approval_snapshot") || strings.HasPrefix(key, "submission_") || strings.HasPrefix(key, "conflict_") {
			common.APIBadRequest(c, "不能修改系统维护的接案状态", "补充回复、事实版本与检测关联必须通过对应操作更新")
			return true
		}
	}
	return false
}

func writeKnownIntakeCollaborationError(c *gin.Context, err error) bool {
	var problem *services.IntakeCollaborationError
	if !errors.As(err, &problem) {
		return false
	}
	common.NewAPIError(c, problem.Status, problem.Code, problem.Message)
	return true
}
