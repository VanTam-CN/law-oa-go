package handlers

import (
	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type LawyerHandler struct {
	lawyerService *services.LawyerService
}

func NewLawyerHandler(lawyerService *services.LawyerService) *LawyerHandler {
	return &LawyerHandler{lawyerService: lawyerService}
}

// ListLawyers godoc
// @Summary 获取律师列表
// @Description 分页获取律师列表，支持搜索
// @Tags 律师管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.APIResponse{data=[]services.LawyerResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /lawfirm/lawyers [get]
func (h *LawyerHandler) ListLawyers(c *gin.Context) {
	handler := APIListHandler(func(c *gin.Context, req *services.LawyerListRequest) ([]*services.LawyerResponse, int64, error) {
		return h.lawyerService.ListLawyers(c.Request.Context(), req)
	}, "lawyers")
	handler(c)
}

// GetLawyerStats 获取律师统计
func (h *LawyerHandler) GetLawyerStats(c *gin.Context) {
	// 模拟律师统计数据
	stats := map[string]interface{}{
		"total_lawyers": 6,
		"active_lawyers": 6,
		"inactive_lawyers": 0,
		"new_lawyers_this_month": 1,
		"specialty_stats": map[string]interface{}{
			"民商事": 2,
			"刑事": 1,
			"知识产权": 1,
			"劳动法": 1,
			"其他": 1,
		},
		"department_stats": map[string]interface{}{
			"诉讼部": 3,
			"非诉部": 2,
			"行政部门": 1,
		},
	}

	common.APISuccess(c, stats)
}

// DeleteLawyer godoc
// @Summary 删除律师
// @Description 根据ID删除律师
// @Tags 律师管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "律师ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "律师不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /lawfirm/lawyers/{id} [delete]
func (h *LawyerHandler) DeleteLawyer(c *gin.Context) {
	handler := APIDeleteHandler(func(c *gin.Context, id uint) error {
		return h.lawyerService.DeleteLawyer(c.Request.Context(), id)
	}, "lawyer")
	handler(c)
}
