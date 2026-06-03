package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type HandoffHandler struct {
	handoffService *services.HandoffService
}

func NewHandoffHandler(handoffService *services.HandoffService) *HandoffHandler {
	return &HandoffHandler{handoffService: handoffService}
}

func (h *HandoffHandler) CreateClientHandoff(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "客户ID必须是有效数字")
		return
	}

	var req services.CreateClientHandoffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	response, err := h.handoffService.CreateClientHandoff(
		c.Request.Context(),
		uint(clientID),
		contextActorID(c),
		contextUserName(c),
		&req,
	)
	if err != nil {
		common.APIBadRequest(c, "创建移交通知失败", err.Error())
		return
	}

	common.APISuccess(c, response)
}

func contextActorID(c *gin.Context) uint {
	value, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	switch v := value.(type) {
	case uint:
		return v
	case int:
		if v > 0 {
			return uint(v)
		}
	case float64:
		if v > 0 {
			return uint(v)
		}
	case string:
		if parsed, err := strconv.ParseUint(v, 10, 32); err == nil {
			return uint(parsed)
		}
	}
	return 0
}
