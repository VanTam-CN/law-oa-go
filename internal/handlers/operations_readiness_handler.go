package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type OperationsReadinessHandler struct {
	service *services.OperationsReadinessService
}

func NewOperationsReadinessHandler(service *services.OperationsReadinessService) *OperationsReadinessHandler {
	return &OperationsReadinessHandler{service: service}
}

func (h *OperationsReadinessHandler) Summary(c *gin.Context) {
	summary, err := h.service.Summary(c.Query("scope"))
	if err != nil {
		common.APIBadRequest(c, "读取运维准备度失败", err.Error())
		return
	}
	common.APISuccess(c, summary)
}

func (h *OperationsReadinessHandler) Register(c *gin.Context) {
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	var input services.OperationsReadinessEvidenceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		common.APIBadRequest(c, "运维证据登记无效", err.Error())
		return
	}
	evidence, err := h.service.Register(actor, input)
	if err != nil {
		common.NewAPIError(c, http.StatusBadRequest, "OPERATIONS_EVIDENCE_INVALID", "运维证据未登记", err.Error())
		return
	}
	common.APISuccess(c, evidence)
}
