package handlers

import (
	"net/http"
	"strconv"

	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
	"github.com/gin-gonic/gin"
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
// @Success 200 {object} common.Response{data=services.CaseResponse} "创建成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部错误"
// @Router /cases [post]
func (h *CaseHandler) CreateCase(c *gin.Context) {
	var req services.CreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	caseResp, err := h.caseService.CreateCase(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "client not found" || err.Error() == "lawyer not found" {
			common.BadRequest(c, "Validation failed: "+err.Error())
			return
		}
		common.InternalServerError(c, "Failed to create case: "+err.Error())
		return
	}

	common.Success(c, caseResp)
}

// GetCase godoc
// @Summary 获取案件详情
// @Description 根据ID获取案件详细信息
// @Tags 案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.Response{data=services.CaseResponse} "获取成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 404 {object} common.Response "案件不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /cases/{id} [get]
func (h *CaseHandler) GetCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid case ID: Case ID must be a valid number")
		return
	}

	caseResp, err := h.caseService.GetCaseByID(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "case not found" {
			common.NotFound(c, "Case not found: The requested case does not exist")
			return
		}
		common.InternalServerError(c, "Failed to get case: "+err.Error())
		return
	}

	common.Success(c, caseResp)
}

func (h *CaseHandler) UpdateCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid case ID", "Case ID must be a valid number")
		return
	}

	var req services.UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	caseResp, err := h.caseService.UpdateCase(c.Request.Context(), uint(id), &req)
	if err != nil {
		if err.Error() == "case not found" {
			common.NotFound(c, "Case not found", "The requested case does not exist")
			return
		}
		if err.Error() == "lawyer not found" {
			common.BadRequest(c, "Validation failed", err.Error())
			return
		}
		common.InternalServerError(c, "Failed to update case", err.Error())
		return
	}

	common.Success(c, caseResp)
}

func (h *CaseHandler) DeleteCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid case ID", "Case ID must be a valid number")
		return
	}

	err = h.caseService.DeleteCase(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "case not found" {
			common.NotFound(c, "Case not found", "The requested case does not exist")
			return
		}
		common.InternalServerError(c, "Failed to delete case", err.Error())
		return
	}

	common.Success(c, gin.H{"message": "Case deleted successfully"})
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
// @Success 200 {object} common.PageResponse{data=[]services.CaseResponse} "获取成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部错误"
// @Router /cases [get]
func (h *CaseHandler) ListCases(c *gin.Context) {
	var req services.CaseListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	cases, total, err := h.caseService.ListCases(c.Request.Context(), &req)
	if err != nil {
		common.InternalServerError(c, "Failed to list cases", err.Error())
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

	response := common.PageResponse{
		Data:     cases,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	common.Success(c, response)
}

func (h *CaseHandler) GetCaseStats(c *gin.Context) {
	stats, err := h.caseService.GetCaseStats(c.Request.Context())
	if err != nil {
		common.InternalServerError(c, "Failed to get case statistics", err.Error())
		return
	}

	common.Success(c, stats)
}

func (h *CaseHandler) AssignLawyer(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid case ID", "Case ID must be a valid number")
		return
	}

	var req struct {
		LawyerID uint `json:"lawyer_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	err = h.caseService.AssignLawyer(c.Request.Context(), uint(caseID), req.LawyerID)
	if err != nil {
		if err.Error() == "case not found" {
			common.NotFound(c, "Case not found", "The requested case does not exist")
			return
		}
		if err.Error() == "lawyer not found" {
			common.BadRequest(c, "Validation failed", err.Error())
			return
		}
		common.InternalServerError(c, "Failed to assign lawyer", err.Error())
		return
	}

	common.Success(c, gin.H{"message": "Lawyer assigned successfully"})
}

func (h *CaseHandler) UpdateCaseStatus(c *gin.Context) {
	idStr := c.Param("id")
	caseID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid case ID", "Case ID must be a valid number")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending active closed suspended"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	err = h.caseService.UpdateCaseStatus(c.Request.Context(), uint(caseID), req.Status)
	if err != nil {
		if err.Error() == "case not found" {
			common.NotFound(c, "Case not found", "The requested case does not exist")
			return
		}
		if err.Error() == "invalid case status" {
			common.BadRequest(c, "Validation failed", err.Error())
			return
		}
		common.InternalServerError(c, "Failed to update case status", err.Error())
		return
	}

	common.Success(c, gin.H{"message": "Case status updated successfully"})
}