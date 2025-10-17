package handlers

import (
	"strconv"
	"strings"

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
		// 提供更详细的验证错误信息
		fieldErrors := make(map[string]string)

		// 手动检查常见字段错误
		if err.Error() == "EOF" {
			fieldErrors["request_body"] = "请求体不能为空"
		} else {
			// 解析具体的字段验证错误
			errMsg := err.Error()
			if strings.Contains(errMsg, "title") {
				fieldErrors["title"] = "案件名称不能为空且长度不超过200字符"
			}
			if strings.Contains(errMsg, "client_id") {
				fieldErrors["client_id"] = "请选择有效的委托客户"
			}
			if strings.Contains(errMsg, "lawyer_id") {
				fieldErrors["lawyer_id"] = "请选择有效的负责律师"
			}
			if strings.Contains(errMsg, "case_type") {
				fieldErrors["case_type"] = "请选择有效的案件类型"
			}
			if strings.Contains(errMsg, "priority") {
				fieldErrors["priority"] = "请选择有效的优先级"
			}
			if strings.Contains(errMsg, "required") {
				fieldErrors["required_fields"] = "请确保所有必填字段都已填写"
			}
		}

		common.APIValidationError(c, "请求参数验证失败", fieldErrors)
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
		// 提供更详细的验证错误信息
		fieldErrors := make(map[string]string)

		errMsg := err.Error()
		if strings.Contains(errMsg, "title") {
			fieldErrors["title"] = "案件名称长度不能超过200字符"
		}
		if strings.Contains(errMsg, "case_type") {
			fieldErrors["case_type"] = "请选择有效的案件类型"
		}
		if strings.Contains(errMsg, "priority") {
			fieldErrors["priority"] = "请选择有效的优先级"
		}
		if strings.Contains(errMsg, "status") {
			fieldErrors["status"] = "请选择有效的案件状态"
		}

		common.APIValidationError(c, "更新参数验证失败", fieldErrors)
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
		// 提供更详细的查询参数验证错误信息
		fieldErrors := make(map[string]string)

		errMsg := err.Error()
		if strings.Contains(errMsg, "page") {
			fieldErrors["page"] = "页码必须是大于0的整数"
		}
		if strings.Contains(errMsg, "page_size") {
			fieldErrors["page_size"] = "每页数量必须是1-100之间的整数"
		}
		if strings.Contains(errMsg, "status") {
			fieldErrors["status"] = "案件状态无效，有效值：pending,active,closed,suspended"
		}
		if strings.Contains(errMsg, "case_type") {
			fieldErrors["case_type"] = "案件类型无效，有效值：civil,criminal,commercial,administrative"
		}
		if strings.Contains(errMsg, "priority") {
			fieldErrors["priority"] = "优先级无效，有效值：low,medium,high,urgent"
		}
		if strings.Contains(errMsg, "client_id") {
			fieldErrors["client_id"] = "客户ID必须是有效的数字"
		}
		if strings.Contains(errMsg, "lawyer_id") {
			fieldErrors["lawyer_id"] = "律师ID必须是有效的数字"
		}

		common.APIValidationError(c, "查询参数验证失败", fieldErrors)
		return
	}

	// 验证筛选参数的组合合理性
	if req.Status != "" && req.Status != "pending" && req.Status != "active" &&
	   req.Status != "closed" && req.Status != "suspended" {
		common.APIBadRequest(c, "无效的案件状态筛选条件", "案件状态必须是：pending,active,closed,suspended 之一")
		return
	}

	if req.CaseType != "" && req.CaseType != "civil" && req.CaseType != "criminal" &&
	   req.CaseType != "commercial" && req.CaseType != "administrative" {
		common.APIBadRequest(c, "无效的案件类型筛选条件", "案件类型必须是：civil,criminal,commercial,administrative 之一")
		return
	}

	if req.Priority != "" && req.Priority != "low" && req.Priority != "medium" &&
	   req.Priority != "high" && req.Priority != "urgent" {
		common.APIBadRequest(c, "无效的优先级筛选条件", "优先级必须是：low,medium,high,urgent 之一")
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

	// 如果没有找到任何案件，返回友好的提示信息
	if len(cases) == 0 {
		common.APISuccessWithPage(c, cases, total, page, pageSize)
		return
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
