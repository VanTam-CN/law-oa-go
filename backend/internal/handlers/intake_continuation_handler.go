package handlers

// GET /case-intakes/:id/continuation returns the server-verified
// continuation decision for one intake. The decision is computed fresh on
// every call from server-owned evidence (committed intake metadata, the
// bound conflict check record, the independent review ledger, and the
// active submission snapshot pointer); browser-restored validity flags are
// never trusted, so refresh and re-login simply re-read the same truth.
// Protected identity material never enters the response: subjects carry
// masked display values only, and lookup digests stay server-side.

import (
	"net/http"
	"strings"

	"law-oa-go/internal/common"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

// IntakeContinuation answers "may this user safely resume the same intake"
// for the workbench. Router registration stays with the main agent; the
// handler itself only translates transport concerns into service calls.
func (h *DemoAggregateHandler) IntakeContinuation(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		common.APIBadRequest(c, "接案记录ID不能为空", "路径缺少接案记录标识")
		return
	}
	var intakeRow struct {
		ClientID  *uint
		CreatedBy string
	}
	if err := h.db.Table("case_intakes").Select("client_id, COALESCE(CAST(created_by AS TEXT), '') AS created_by").
		Where("id = ?", id).Take(&intakeRow).Error; err != nil {
		common.APINotFound(c, "接案记录不存在", "指定接案记录不存在")
		return
	}
	// Same object boundary as the workbench read: management roles may open
	// any intake they may see, everyone else only their own intake rows.
	if !canViewAllMatterData(c) {
		actorID, ok := currentUserIDString(c)
		if !ok {
			return
		}
		if strings.TrimSpace(intakeRow.CreatedBy) != actorID {
			forbidObjectAccess(c)
			return
		}
	}
	if clientID := intakeRow.ClientID; clientID != nil && *clientID > 0 && !isAssistantRole(c) {
		if !h.authorizeIntakeClient(c, *clientID) {
			return
		}
	}
	service := services.NewIntakeContinuationService(h.db)
	decision, err := service.EvaluateIntakeContinuation(id)
	if err != nil {
		common.APIInternalServerError(c, "续办状态评估失败", err.Error())
		return
	}
	// Revision guard mirrors the write path: a stale X-Intake-Revision must
	// 409 with the server revision so the client reloads instead of acting on
	// an outdated page. Absent header means read-only inspection.
	if expected := strings.TrimSpace(c.GetHeader("X-Intake-Revision")); expected != "" && expected != decision.Revision {
		common.APIErrorWithContext(c, http.StatusConflict, "INTAKE_DRAFT_REVISION_CONFLICT",
			"接案版本已变化，请刷新后继续",
			map[string]interface{}{"current_revision": decision.Revision})
		return
	}
	common.APISuccess(c, decision)
}
