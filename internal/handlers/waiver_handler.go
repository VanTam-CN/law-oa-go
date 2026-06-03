package handlers

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type WaiverHandler struct {
	waiverService *services.WaiverWorkflowService
}

func NewWaiverHandler(waiverService *services.WaiverWorkflowService) *WaiverHandler {
	return &WaiverHandler{waiverService: waiverService}
}

func (h *WaiverHandler) CreateWaiverRequest(c *gin.Context) {
	approvalID := c.Param("id")
	var req services.CreateWaiverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	userID := contextUserID(c)
	userName := contextUserName(c)
	application, err := h.waiverService.CreateWaiver(c.Request.Context(), approvalID, userID, userName, &req)
	if err != nil {
		common.APIBadRequest(c, "创建豁免申请失败", err.Error())
		return
	}

	common.APISuccess(c, application)
}

func (h *WaiverHandler) GetWaiverRequest(c *gin.Context) {
	waiverID := c.Param("id")
	if waiverID == "" {
		common.APIBadRequest(c, "豁免申请ID不能为空")
		return
	}

	application, err := h.waiverService.GetWaiver(c.Request.Context(), waiverID)
	if err != nil {
		if errors.Is(err, services.ErrWaiverNotFound) {
			common.APINotFound(c, "豁免申请不存在")
			return
		}
		common.APIInternalServerError(c, "获取豁免申请失败", err.Error())
		return
	}

	common.APISuccess(c, application)
}

func (h *WaiverHandler) DecideWaiverRequest(c *gin.Context) {
	waiverID := c.Param("id")
	if waiverID == "" {
		common.APIBadRequest(c, "豁免申请ID不能为空")
		return
	}

	var req services.WaiverDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	application, err := h.waiverService.DecideWaiver(c.Request.Context(), waiverID, contextUserID(c), contextUserName(c), &req)
	if err != nil {
		if errors.Is(err, services.ErrWaiverNotFound) {
			common.APINotFound(c, "豁免申请不存在")
			return
		}
		common.APIBadRequest(c, "处理豁免决定失败", err.Error())
		return
	}

	common.APISuccess(c, application)
}

func contextUserID(c *gin.Context) string {
	value, exists := c.Get("user_id")
	if !exists {
		return "0"
	}
	switch v := value.(type) {
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case string:
		if v != "" {
			return v
		}
	}
	return "0"
}

func contextUserName(c *gin.Context) string {
	value, exists := c.Get("username")
	if !exists {
		return "未知用户"
	}
	if name, ok := value.(string); ok && name != "" {
		return name
	}
	return "未知用户"
}
