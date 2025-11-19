package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type EnhancedCaseHandler struct {
	enhancedCaseService services.EnhancedCaseService
}

func NewEnhancedCaseHandler(enhancedCaseService services.EnhancedCaseService) *EnhancedCaseHandler {
	return &EnhancedCaseHandler{
		enhancedCaseService: enhancedCaseService,
	}
}

// CreateEnhancedCase godoc
// @Summary 创建增强案件
// @Description 创建新的增强法律案件，支持多客户和冲突检测
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.EnhancedCreateCaseRequest true "创建增强案件请求"
// @Success 200 {object} common.APIResponse{data=services.EnhancedCaseResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases [post]
func (h *EnhancedCaseHandler) CreateEnhancedCase(c *gin.Context) {
	var req services.EnhancedCreateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}

	caseResp, err := h.enhancedCaseService.CreateEnhancedCase(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建增强案件失败", err.Error())
		return
	}

	common.APISuccess(c, caseResp)
}

// GetEnhancedCase godoc
// @Summary 获取增强案件详情
// @Description 根据ID获取增强案件详细信息
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=services.EnhancedCaseResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases/{id} [get]
func (h *EnhancedCaseHandler) GetEnhancedCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的案件ID", "ID必须是正整数")
		return
	}

	caseResp, err := h.enhancedCaseService.GetEnhancedCaseByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "案件不存在", err.Error())
		return
	}

	common.APISuccess(c, caseResp)
}

// UpdateEnhancedCase godoc
// @Summary 更新增强案件
// @Description 更新增强案件信息
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param request body services.UpdateEnhancedCaseRequest true "更新案件请求"
// @Success 200 {object} common.APIResponse{data=services.EnhancedCaseResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases/{id} [put]
func (h *EnhancedCaseHandler) UpdateEnhancedCase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的案件ID", "ID必须是正整数")
		return
	}

	var req services.UpdateEnhancedCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}

	caseResp, err := h.enhancedCaseService.UpdateEnhancedCase(c.Request.Context(), uint(id), &req)
	if err != nil {
		common.APIInternalServerError(c, "更新增强案件失败", err.Error())
		return
	}

	common.APISuccess(c, caseResp)
}

// ListEnhancedCases godoc
// @Summary 获取增强案件列表
// @Description 分页获取增强案件列表，支持搜索和过滤
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param search query string false "搜索关键词"
// @Param status query string false "案件状态"
// @Param priority query string false "优先级"
// @Param lawyer_id query int false "主办律师ID"
// @Success 200 {object} common.APIResponse{data=services.EnhancedCaseListResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases [get]
func (h *EnhancedCaseHandler) ListEnhancedCases(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	status := c.Query("status")
	priority := c.Query("priority")
	lawyerIDStr := c.Query("lawyer_id")

	listReq := &services.ListEnhancedCasesRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
		Status:   status,
		Priority: priority,
		LawyerID: lawyerIDStr,
	}

	listResp, err := h.enhancedCaseService.ListEnhancedCases(c.Request.Context(), listReq)
	if err != nil {
		common.APIInternalServerError(c, "获取增强案件列表失败", err.Error())
		return
	}

	common.APISuccess(c, listResp)
}

// DeleteEnhancedCase godoc
// @Summary 删除增强案件
// @Description 删除指定的增强案件（暂未实现）
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases/{id} [delete]
func (h *EnhancedCaseHandler) DeleteEnhancedCase(c *gin.Context) {
	// 暂时返回未实现错误
	common.APIInternalServerError(c, "功能暂未实现", "删除功能正在开发中")
}

// PerformConflictCheck godoc
// @Summary 执行冲突检测
// @Description 为指定案件执行冲突检测
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Success 200 {object} common.APIResponse{data=services.ConflictCheckResult} "检测成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases/{id}/conflict-check [post]
func (h *EnhancedCaseHandler) PerformConflictCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的案件ID", "ID必须是正整数")
		return
	}

	// 使用现有的方法触发冲突检测
	err = h.enhancedCaseService.TriggerConflictDetection(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "执行冲突检测失败", err.Error())
		return
	}

	// 获取检测状态
	status, err := h.enhancedCaseService.GetConflictDetectionStatus(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取冲突检测状态失败", err.Error())
		return
	}

	common.APISuccess(c, status)
}

// AddClientToCase godoc
// @Summary 添加客户到案件
// @Description 向现有案件添加新客户
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param request body services.AddClientToCaseRequest true "添加客户请求"
// @Success 200 {object} common.APIResponse{data=services.EnhancedCaseResponse} "添加成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases/{id}/clients [post]
func (h *EnhancedCaseHandler) AddClientToCase(c *gin.Context) {
	// 暂时返回未实现错误
	common.APIInternalServerError(c, "功能暂未实现", "添加客户功能正在开发中")
}

// RemoveClientFromCase godoc
// @Summary 从案件移除客户
// @Description 从案件中移除指定客户
// @Tags 增强案件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "案件ID"
// @Param client_id path string true "客户档案ID"
// @Success 200 {object} common.APIResponse{data=services.EnhancedCaseResponse} "移除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "案件不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /enhanced-cases/{id}/clients/{client_id} [delete]
func (h *EnhancedCaseHandler) RemoveClientFromCase(c *gin.Context) {
	// 暂时返回未实现错误
	common.APIInternalServerError(c, "功能暂未实现", "移除客户功能正在开发中")
}