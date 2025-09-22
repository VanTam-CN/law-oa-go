package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/services"
)

type CaseHandler struct {
	caseService *services.CaseService
}

func NewCaseHandler(caseService *services.CaseService) *CaseHandler {
	return &CaseHandler{
		caseService: caseService,
	}
}

// CreateCase godoc
// @Summary 创建案件
// @Description 创建新的法律案件
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateCaseRequest true "创建案件请求"
// @Success 200 {object} common.APIResponse{data=services.CaseResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases [post]
func (h *CaseHandler) CreateCase(c *gin.Context) {
	var req services.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format: "+err.Error(), "Invalid request format"))
		return
	}

	caseResp, err := h.caseService.CreateCase(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, caseResp)
}

// GetCase godoc
// @Summary 获取案件详情
// @Description 根据ID获取案件详细信息
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=services.CaseResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id} [get]
func (h *CaseHandler) GetCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid case ID: Case ID must be a valid number", "Invalid case ID: Case ID must be a valid number"))
		return
	}

	caseResp, err := h.caseService.GetCaseByID(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, caseResp)
}

// UpdateCase godoc
// @Summary 更新案件
// @Description 更新指定案件的信息
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param request body services.UpdateCaseRequest true "更新案件请求"
// @Success 200 {object} common.APIResponse{data=services.CaseResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id} [put]
func (h *CaseHandler) UpdateCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid case ID: Case ID must be a valid number", "Invalid case ID: Case ID must be a valid number"))
		return
	}

	var req services.UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format: "+err.Error(), "Invalid request format"))
		return
	}

	caseResp, err := h.caseService.UpdateCase(c.Request.Context(), uint(id), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, caseResp)
}

// DeleteCase godoc
// @Summary 删除案件
// @Description 删除指定的案件
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id} [delete]
func (h *CaseHandler) DeleteCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid case ID: Case ID must be a valid number", "Invalid case ID: Case ID must be a valid number"))
		return
	}

	err = h.caseService.DeleteCase(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, gin.H{"message": "Case deleted successfully"})
}

// ListCases godoc
// @Summary 获取案件列表
// @Description 分页获取案件列表，支持过滤和搜索
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "案件状态" Enums(pending,active,closed,suspended)
// @Param case_type query string false "案件类型" Enums(civil,criminal,commercial,administrative)
// @Param priority query string false "优先级" Enums(low,medium,high,urgent)
// @Param client_id query int false "客户ID"
// @Param lawyer_id query int false "律师ID"
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.APIResponse{data=[]services.CaseResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases [get]
func (h *CaseHandler) ListCases(c *gin.Context) {
	var req services.CaseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(errors.NewValidationError("query_binding", "query_binding", "Invalid query parameters: "+err.Error(), "Invalid query parameters"))
		return
	}

	cases, total, err := h.caseService.ListCases(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	common.APISuccessWithPage(c, cases, total, page, pageSize)
}

// GetCaseStats godoc
// @Summary 获取案件统计
// @Description 获取案件相关的统计数据
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.CaseStatsResponse} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/stats [get]
func (h *CaseHandler) GetCaseStats(c *gin.Context) {
	stats, err := h.caseService.GetCaseStats(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, stats)
}

// AssignLawyer godoc
// @Summary 分配律师
// @Description 为案件分配律师
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param request body object true "分配律师请求"
// @Success 200 {object} common.APIResponse "分配成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id}/assign-lawyer [post]
func (h *CaseHandler) AssignLawyer(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid case ID: Case ID must be a valid number", "Invalid case ID: Case ID must be a valid number"))
		return
	}

	var req struct {
		LawyerID uint `json:"lawyer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format: "+err.Error(), "Invalid request format"))
		return
	}

	err = h.caseService.AssignLawyer(c.Request.Context(), uint(caseID), req.LawyerID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, gin.H{"message": "Lawyer assigned successfully"})
}

// UpdateCaseStatus godoc
// @Summary 更新案件状态
// @Description 更新指定案件的状态
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param request body object true "状态更新请求"
// @Success 200 {object} common.APIResponse "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /cases/{id}/status [put]
func (h *CaseHandler) UpdateCaseStatus(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		_ = c.Error(errors.NewValidationError("id_validation", "id_validation", "Invalid case ID: Case ID must be a valid number", "Invalid case ID: Case ID must be a valid number"))
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending active closed suspended"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format: "+err.Error(), "Invalid request format"))
		return
	}

	err = h.caseService.UpdateCaseStatus(c.Request.Context(), uint(caseID), req.Status)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, gin.H{"message": "Case status updated successfully"})
}
