package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// CommissionRuleHandler 分成规则Handler
type CommissionRuleHandler struct {
	commissionService *services.CommissionService
}

// NewCommissionRuleHandler 创建分成规则Handler实例
func NewCommissionRuleHandler(commissionService *services.CommissionService) *CommissionRuleHandler {
	return &CommissionRuleHandler{
		commissionService: commissionService,
	}
}

// ListCommissionRules godoc
// @Summary 获取分成规则列表
// @Description 获取所有分成规则
// @Tags 财务管理-分成规则
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]models.CommissionRule} "获取成功"
// @Router /finance/commission-rules [get]
func (h *CommissionRuleHandler) ListCommissionRules(c *gin.Context) {
	rules, err := h.commissionService.ListCommissionRules(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取分成规则失败", err.Error())
		return
	}
	common.APISuccess(c, rules)
}

// CreateCommissionRule godoc
// @Summary 创建分成规则
// @Description 创建新的分成规则
// @Tags 财务管理-分成规则
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CommissionRule true "分成规则"
// @Success 200 {object} common.APIResponse "创建成功"
// @Router /finance/commission-rules [post]
func (h *CommissionRuleHandler) CreateCommissionRule(c *gin.Context) {
	var rule models.CommissionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	if rule.Name == "" || rule.Role == "" {
		common.APIBadRequest(c, "请求参数错误", "规则名称和角色不能为空")
		return
	}

	if err := h.commissionService.CreateCommissionRule(c.Request.Context(), &rule); err != nil {
		common.APIInternalServerError(c, "创建分成规则失败", err.Error())
		return
	}
	common.APISuccess(c, rule)
}

// UpdateCommissionRule godoc
// @Summary 更新分成规则
// @Description 更新指定的分成规则
// @Tags 财务管理-分成规则
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Param request body models.CommissionRule true "分成规则"
// @Success 200 {object} common.APIResponse "更新成功"
// @Router /finance/commission-rules/{id} [put]
func (h *CommissionRuleHandler) UpdateCommissionRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "ID必须是有效数字")
		return
	}

	var rule models.CommissionRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}
	rule.ID = uint(id)

	if err := h.commissionService.UpdateCommissionRule(c.Request.Context(), &rule); err != nil {
		common.APIInternalServerError(c, "更新分成规则失败", err.Error())
		return
	}
	common.APISuccess(c, rule)
}

// DeleteCommissionRule godoc
// @Summary 删除分成规则
// @Description 删除指定的分成规则
// @Tags 财务管理-分成规则
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "规则ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Router /finance/commission-rules/{id} [delete]
func (h *CommissionRuleHandler) DeleteCommissionRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "ID必须是有效数字")
		return
	}

	if err := h.commissionService.DeleteCommissionRule(c.Request.Context(), uint(id)); err != nil {
		common.APIInternalServerError(c, "删除分成规则失败", err.Error())
		return
	}
	common.APISuccess(c, nil)
}
