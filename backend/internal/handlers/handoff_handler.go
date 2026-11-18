package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type HandoffHandler struct {
	handoffService *services.HandoffService
	authz          *services.AuthorizationService
}

func NewHandoffHandler(handoffService *services.HandoffService, authz ...*services.AuthorizationService) *HandoffHandler {
	var authorizationService *services.AuthorizationService
	if len(authz) > 0 {
		authorizationService = authz[0]
	}
	return &HandoffHandler{handoffService: handoffService, authz: authorizationService}
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
	if h.authz == nil {
		common.NewAPIError(c, 503, "CLIENT_AUTHZ_UNAVAILABLE", "客户权限服务未初始化，当前已阻止客户移交")
		return
	}
	actor, ok := currentAuthActor(c)
	if !ok {
		return
	}
	allowed, authErr := h.authz.CanReadClient(c.Request.Context(), actor, uint(clientID))
	if authErr != nil {
		common.APIInternalServerError(c, "客户权限校验失败", authErr.Error())
		return
	}
	if !allowed {
		forbidObjectAccess(c)
		return
	}

	response, err := h.handoffService.CreateClientHandoff(
		c.Request.Context(),
		uint(clientID),
		actor.UserID,
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
