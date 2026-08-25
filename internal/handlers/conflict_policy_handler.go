package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type ConflictPolicyHandler struct {
	service *services.ConflictPolicyService
}

func NewConflictPolicyHandler(service *services.ConflictPolicyService) *ConflictPolicyHandler {
	return &ConflictPolicyHandler{service: service}
}

func (h *ConflictPolicyHandler) List(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	items, err := h.service.ListPackages(c.Request.Context(), actor)
	if err != nil {
		h.writeError(c, err, "读取冲突政策签署记录失败")
		return
	}
	common.APISuccess(c, items)
}

func (h *ConflictPolicyHandler) Create(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	var input services.ConflictPolicyPackageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		common.APIBadRequest(c, "政策材料包格式无效", err.Error())
		return
	}
	item, err := h.service.CreatePackage(c.Request.Context(), actor, input)
	if err != nil {
		h.writeError(c, err, "政策材料包未提交")
		return
	}
	common.APISuccess(c, item)
}

func (h *ConflictPolicyHandler) Endorse(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	item, err := h.service.Endorse(c.Request.Context(), actor, c.Param("id"))
	if err != nil {
		h.writeError(c, err, "政策确认未记录")
		return
	}
	common.APISuccess(c, item)
}

func (h *ConflictPolicyHandler) writeError(c *gin.Context, err error, message string) {
	var workflowErr *services.SubjectWorkflowError
	if errors.As(err, &workflowErr) {
		writeSubjectWorkflowError(c, err)
		return
	}
	common.NewAPIError(c, http.StatusBadRequest, "CONFLICT_POLICY_INVALID", message, err.Error())
}
