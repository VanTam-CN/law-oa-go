package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type ConflictScopeHandler struct {
	service *services.ConflictScopeService
}

func NewConflictScopeHandler(service *services.ConflictScopeService) *ConflictScopeHandler {
	return &ConflictScopeHandler{service: service}
}

func (h *ConflictScopeHandler) List(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权查看档案覆盖配置", "只有冲突核查岗或获授权管理合伙人可以查看档案覆盖配置")
		return
	}
	scopes, err := h.service.List(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "读取档案覆盖配置失败", err.Error())
		return
	}
	common.APISuccess(c, scopes)
}

func (h *ConflictScopeHandler) Upsert(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	if !services.IsConflictReviewRole(actor.Role) {
		common.APIForbidden(c, "无权维护档案覆盖配置", "只有冲突核查岗或获授权管理合伙人可以维护档案覆盖配置")
		return
	}
	var input services.ConflictSearchScopeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		common.APIBadRequest(c, "档案覆盖配置无效", err.Error())
		return
	}
	input.ID = c.Param("id")
	scope, err := h.service.Upsert(c.Request.Context(), actor, input)
	if err != nil {
		var workflowErr *services.SubjectWorkflowError
		if errors.As(err, &workflowErr) {
			writeSubjectWorkflowError(c, err)
			return
		}
		common.NewAPIError(c, http.StatusBadRequest, "CONFLICT_SCOPE_INVALID", "档案覆盖配置未保存", err.Error())
		return
	}
	common.APISuccess(c, scope)
}
