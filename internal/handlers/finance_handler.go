package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// FinanceHandler 财务处理器
type FinanceHandler struct {
	contractService    *services.ContractService
	milestoneService   *services.PaymentMilestoneService
	invoiceService     *services.InvoiceService
	paymentService     *services.PaymentService
	badDebtService     *services.BadDebtService
	commissionService  *services.CommissionService
}

// NewFinanceHandler 创建财务处理器实例
func NewFinanceHandler(
	contractService *services.ContractService,
	milestoneService *services.PaymentMilestoneService,
	invoiceService *services.InvoiceService,
	paymentService *services.PaymentService,
	badDebtService *services.BadDebtService,
	commissionService *services.CommissionService,
) *FinanceHandler {
	return &FinanceHandler{
		contractService:   contractService,
		milestoneService:  milestoneService,
		invoiceService:    invoiceService,
		paymentService:    paymentService,
		badDebtService:    badDebtService,
		commissionService: commissionService,
	}
}

// ============================================================================
// 合同管理
// ============================================================================

// CreateContract godoc
// @Summary 创建合同
// @Description 创建新的合同
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateContractRequest true "创建合同请求"
// @Success 200 {object} common.APIResponse{data=services.ContractResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts [post]
func (h *FinanceHandler) CreateContract(c *gin.Context) {
	var req services.CreateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	contract, err := h.contractService.CreateContract(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建合同失败", err.Error())
		return
	}

	common.APISuccess(c, contract)
}

// GetContract godoc
// @Summary 获取合同详情
// @Description 根据ID获取合同详细信息
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "合同ID"
// @Success 200 {object} common.APIResponse{data=services.ContractResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "合同不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/{id} [get]
func (h *FinanceHandler) GetContract(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "合同ID必须是有效数字")
		return
	}

	contract, err := h.contractService.GetContractByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "合同不存在", err.Error())
		return
	}

	common.APISuccess(c, contract)
}

// UpdateContract godoc
// @Summary 更新合同
// @Description 更新指定合同的信息
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "合同ID"
// @Param request body services.UpdateContractRequest true "更新合同请求"
// @Success 200 {object} common.APIResponse{data=services.ContractResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "合同不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/{id} [put]
func (h *FinanceHandler) UpdateContract(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "合同ID必须是有效数字")
		return
	}

	var req services.UpdateContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	contract, err := h.contractService.UpdateContract(c.Request.Context(), uint(id), &req)
	if err != nil {
		common.APIInternalServerError(c, "更新合同失败", err.Error())
		return
	}

	common.APISuccess(c, contract)
}

// DeleteContract godoc
// @Summary 删除合同
// @Description 删除指定的合同（仅草稿状态）
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "合同ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "合同不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/{id} [delete]
func (h *FinanceHandler) DeleteContract(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "合同ID必须是有效数字")
		return
	}

	err = h.contractService.DeleteContract(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "删除合同失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "合同删除成功"})
}

// ListContracts godoc
// @Summary 获取合同列表
// @Description 获取合同列表，支持分页和筛选
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态" Enums(draft, active, suspended, completed, cancelled)
// @Param contract_type query string false "合同类型" Enums(original, supplementary)
// @Param client_id query int false "客户ID"
// @Param case_id query int false "案件ID"
// @Param search query string false "搜索关键词"
// @Param start_date_from query string false "开始日期(起)"
// @Param start_date_to query string false "开始日期(止)"
// @Param end_date_from query string false "结束日期(起)"
// @Param end_date_to query string false "结束日期(止)"
// @Success 200 {object} common.APIResponse{data=services.ListContractsResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts [get]
func (h *FinanceHandler) ListContracts(c *gin.Context) {
	var req services.ListContractsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := h.contractService.ListContracts(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "查询合同列表失败", err.Error())
		return
	}

	common.APISuccessWithPage(c, result.Contracts, result.Pagination.Total, req.Page, req.PageSize)
}

// ActivateContract godoc
// @Summary 激活合同
// @Description 激活指定的合同
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "合同ID"
// @Success 200 {object} common.APIResponse{data=services.ContractResponse} "激活成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "合同不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/{id}/activate [post]
func (h *FinanceHandler) ActivateContract(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "合同ID必须是有效数字")
		return
	}

	contract, err := h.contractService.ActivateContract(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "激活合同失败", err.Error())
		return
	}

	common.APISuccess(c, contract)
}

// SuspendContract godoc
// @Summary 暂停合同
// @Description 暂停指定的合同
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "合同ID"
// @Success 200 {object} common.APIResponse{data=services.ContractResponse} "暂停成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "合同不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/{id}/suspend [post]
func (h *FinanceHandler) SuspendContract(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "合同ID必须是有效数字")
		return
	}

	contract, err := h.contractService.SuspendContract(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "暂停合同失败", err.Error())
		return
	}

	common.APISuccess(c, contract)
}

// CompleteContract godoc
// @Summary 完成合同
// @Description 完成指定的合同
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "合同ID"
// @Success 200 {object} common.APIResponse{data=services.ContractResponse} "完成成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "合同不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/{id}/complete [post]
func (h *FinanceHandler) CompleteContract(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "合同ID必须是有效数字")
		return
	}

	contract, err := h.contractService.CompleteContract(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "完成合同失败", err.Error())
		return
	}

	common.APISuccess(c, contract)
}

// GetContractStats godoc
// @Summary 获取合同统计信息
// @Description 获取合同统计数据
// @Tags 财务管理-合同
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.ContractStats} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/stats [get]
func (h *FinanceHandler) GetContractStats(c *gin.Context) {
	stats, err := h.contractService.GetContractStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取合同统计失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// ============================================================================
// 付款计划管理
// ============================================================================

// CreateMilestone godoc
// @Summary 创建付款计划
// @Description 为合同创建付款计划
// @Tags 财务管理-付款计划
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateMilestoneRequest true "创建付款计划请求"
// @Success 200 {object} common.APIResponse{data=services.PaymentMilestoneResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/milestones [post]
func (h *FinanceHandler) CreateMilestone(c *gin.Context) {
	var req struct {
		ContractID uint   `json:"contract_id" binding:"required"`
		Name       string `json:"name" binding:"required,min=1,max=200"`
		Amount     float64 `json:"amount" binding:"required,gt=0"`
		Percentage float64 `json:"percentage" binding:"required,gte=0,lte=100"`
		DueDate    *string `json:"due_date,omitempty"`
		Condition  string `json:"condition" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 转换为服务层请求类型
	serviceReq := &services.PaymentMilestoneCreateRequest{
		ContractID: req.ContractID,
		Name:       req.Name,
		Amount:     req.Amount,
		Percentage: req.Percentage,
		DueDate:    req.DueDate,
		Condition:  req.Condition,
	}

	milestone, err := h.milestoneService.CreateMilestone(c.Request.Context(), serviceReq)
	if err != nil {
		common.APIInternalServerError(c, "创建付款计划失败", err.Error())
		return
	}

	common.APISuccess(c, milestone)
}

// UpdateMilestone godoc
// @Summary 更新付款计划
// @Description 更新指定的付款计划
// @Tags 财务管理-付款计划
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "付款计划ID"
// @Param request body services.UpdateMilestoneRequest true "更新付款计划请求"
// @Success 200 {object} common.APIResponse{data=services.PaymentMilestoneResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "付款计划不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/milestones/{id} [put]
func (h *FinanceHandler) UpdateMilestone(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "付款计划ID必须是有效数字")
		return
	}

	var req services.UpdateMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	milestone, err := h.milestoneService.UpdateMilestone(c.Request.Context(), uint(id), &req)
	if err != nil {
		common.APIInternalServerError(c, "更新付款计划失败", err.Error())
		return
	}

	common.APISuccess(c, milestone)
}

// DeleteMilestone godoc
// @Summary 删除付款计划
// @Description 删除指定的付款计划
// @Tags 财务管理-付款计划
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "付款计划ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "付款计划不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/milestones/{id} [delete]
func (h *FinanceHandler) DeleteMilestone(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "付款计划ID必须是有效数字")
		return
	}

	err = h.milestoneService.DeleteMilestone(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "删除付款计划失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "付款计划删除成功"})
}

// GetMilestonesByContractID godoc
// @Summary 获取合同的付款计划
// @Description 获取指定合同的所有付款计划
// @Tags 财务管理-付款计划
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contract_id path int true "合同ID"
// @Success 200 {object} common.APIResponse{data=[]services.PaymentMilestoneResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/contracts/{contract_id}/milestones [get]
func (h *FinanceHandler) GetMilestonesByContractID(c *gin.Context) {
	contractIDStr := c.Param("contract_id")
	contractID, err := strconv.ParseUint(contractIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "合同ID必须是有效数字")
		return
	}

	milestones, err := h.milestoneService.GetMilestonesByContractID(c.Request.Context(), uint(contractID))
	if err != nil {
		common.APIInternalServerError(c, "查询付款计划失败", err.Error())
		return
	}

	common.APISuccess(c, milestones)
}

// ============================================================================
// 发票管理
// ============================================================================

// CreateInvoice godoc
// @Summary 创建发票
// @Description 创建新的发票
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateInvoiceRequest true "创建发票请求"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices [post]
func (h *FinanceHandler) CreateInvoice(c *gin.Context) {
	var req services.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 从上下文获取用户ID
	userID := c.GetUint("user_id")

	invoice, err := h.invoiceService.CreateInvoice(c.Request.Context(), &req, userID)
	if err != nil {
		common.APIInternalServerError(c, "创建发票失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// GetInvoice godoc
// @Summary 获取发票详情
// @Description 根据ID获取发票详细信息
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id} [get]
func (h *FinanceHandler) GetInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	invoice, err := h.invoiceService.GetInvoiceByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "发票不存在", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// UpdateInvoice godoc
// @Summary 更新发票
// @Description 更新指定发票的信息
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Param request body services.UpdateInvoiceRequest true "更新发票请求"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id} [put]
func (h *FinanceHandler) UpdateInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	var req services.UpdateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	invoice, err := h.invoiceService.UpdateInvoice(c.Request.Context(), uint(id), &req)
	if err != nil {
		common.APIInternalServerError(c, "更新发票失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// DeleteInvoice godoc
// @Summary 删除发票
// @Description 删除指定的发票（仅草稿状态）
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id} [delete]
func (h *FinanceHandler) DeleteInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	err = h.invoiceService.DeleteInvoice(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "删除发票失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "发票删除成功"})
}

// ListInvoices godoc
// @Summary 获取发票列表
// @Description 获取发票列表，支持分页和筛选
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态" Enums(draft, submitted, approved, issued, received, cancelled)
// @Param invoice_type query string false "发票类型" Enums(normal, credit)
// @Param client_id query int false "客户ID"
// @Param contract_id query int false "合同ID"
// @Param search query string false "搜索关键词"
// @Param date_from query string false "开始日期"
// @Param date_to query string false "结束日期"
// @Success 200 {object} common.APIResponse{data=services.ListInvoicesResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices [get]
func (h *FinanceHandler) ListInvoices(c *gin.Context) {
	var req services.ListInvoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := h.invoiceService.ListInvoices(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "查询发票列表失败", err.Error())
		return
	}

	common.APISuccessWithPage(c, result.Invoices, result.Pagination.Total, req.Page, req.PageSize)
}

// SubmitInvoice godoc
// @Summary 提交发票审批
// @Description 提交发票进行财务审批
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "提交成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id}/submit [post]
func (h *FinanceHandler) SubmitInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	// 从上下文获取用户ID
	userID := c.GetUint("user_id")

	invoice, err := h.invoiceService.SubmitInvoice(c.Request.Context(), uint(id), userID)
	if err != nil {
		common.APIInternalServerError(c, "提交发票失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// ApproveInvoice godoc
// @Summary 审批通过发票
// @Description 财务审批通过发票
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "审批成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id}/approve [post]
func (h *FinanceHandler) ApproveInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	// 从上下文获取用户ID
	userID := c.GetUint("user_id")

	invoice, err := h.invoiceService.ApproveInvoice(c.Request.Context(), uint(id), userID)
	if err != nil {
		common.APIInternalServerError(c, "审批发票失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// RejectInvoice godoc
// @Summary 审批拒绝发票
// @Description 财务审批拒绝发票
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "拒绝成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id}/reject [post]
func (h *FinanceHandler) RejectInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	invoice, err := h.invoiceService.RejectInvoice(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "拒绝发票失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// IssueInvoice godoc
// @Summary 开票
// @Description 发票开票
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Param request body object{electronic_url:string,code:string,number:string} true "开票信息"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "开票成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id}/issue [post]
func (h *FinanceHandler) IssueInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	var req struct {
		ElectronicURL string `json:"electronic_url"`
		Code          string `json:"code"`
		Number        string `json:"number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	invoice, err := h.invoiceService.IssueInvoice(c.Request.Context(), uint(id), req.ElectronicURL, req.Code, req.Number)
	if err != nil {
		common.APIInternalServerError(c, "开票失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// ConfirmInvoiceReceipt godoc
// @Summary 客户签收
// @Description 确认客户已签收发票
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "签收成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id}/confirm-receipt [post]
func (h *FinanceHandler) ConfirmInvoiceReceipt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	invoice, err := h.invoiceService.ConfirmReceipt(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "签收失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// CancelInvoice godoc
// @Summary 作废发票
// @Description 作废指定的发票
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "发票ID"
// @Success 200 {object} common.APIResponse{data=services.InvoiceResponse} "作废成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "发票不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{id}/cancel [post]
func (h *FinanceHandler) CancelInvoice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	invoice, err := h.invoiceService.CancelInvoice(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "作废发票失败", err.Error())
		return
	}

	common.APISuccess(c, invoice)
}

// GetInvoiceStats godoc
// @Summary 获取发票统计信息
// @Description 获取发票统计数据
// @Tags 财务管理-发票
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.InvoiceStats} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/stats [get]
func (h *FinanceHandler) GetInvoiceStats(c *gin.Context) {
	stats, err := h.invoiceService.GetInvoiceStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取发票统计失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// ============================================================================
// 回款管理
// ============================================================================

// CreatePayment godoc
// @Summary 创建回款记录
// @Description 登记新的回款记录
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreatePaymentRequest true "创建回款请求"
// @Success 200 {object} common.APIResponse{data=services.PaymentResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments [post]
func (h *FinanceHandler) CreatePayment(c *gin.Context) {
	var req services.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 从上下文获取用户ID
	userID := c.GetUint("user_id")

	payment, err := h.paymentService.CreatePayment(c.Request.Context(), &req, userID)
	if err != nil {
		common.APIInternalServerError(c, "创建回款记录失败", err.Error())
		return
	}

	common.APISuccess(c, payment)
}

// GetPayment godoc
// @Summary 获取回款详情
// @Description 根据ID获取回款详细信息
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "回款ID"
// @Success 200 {object} common.APIResponse{data=services.PaymentResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "回款记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments/{id} [get]
func (h *FinanceHandler) GetPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "回款ID必须是有效数字")
		return
	}

	payment, err := h.paymentService.GetPaymentByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "回款记录不存在", err.Error())
		return
	}

	common.APISuccess(c, payment)
}

// UpdatePayment godoc
// @Summary 更新回款记录
// @Description 更新指定的回款记录
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "回款ID"
// @Param request body services.UpdatePaymentRequest true "更新回款请求"
// @Success 200 {object} common.APIResponse{data=services.PaymentResponse} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "回款记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments/{id} [put]
func (h *FinanceHandler) UpdatePayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "回款ID必须是有效数字")
		return
	}

	var req services.UpdatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	payment, err := h.paymentService.UpdatePayment(c.Request.Context(), uint(id), &req)
	if err != nil {
		common.APIInternalServerError(c, "更新回款记录失败", err.Error())
		return
	}

	common.APISuccess(c, payment)
}

// DeletePayment godoc
// @Summary 删除回款记录
// @Description 删除指定的回款记录（仅待确认状态）
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "回款ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "回款记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments/{id} [delete]
func (h *FinanceHandler) DeletePayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "回款ID必须是有效数字")
		return
	}

	err = h.paymentService.DeletePayment(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "删除回款记录失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "回款记录删除成功"})
}

// ListPayments godoc
// @Summary 获取回款列表
// @Description 获取回款列表，支持分页和筛选
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态" Enums(pending, confirmed, rejected)
// @Param invoice_id query int false "发票ID"
// @Param client_id query int false "客户ID"
// @Param search query string false "搜索关键词"
// @Param date_from query string false "开始日期"
// @Param date_to query string false "结束日期"
// @Success 200 {object} common.APIResponse{data=services.ListPaymentsResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments [get]
func (h *FinanceHandler) ListPayments(c *gin.Context) {
	var req services.ListPaymentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := h.paymentService.ListPayments(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "查询回款列表失败", err.Error())
		return
	}

	common.APISuccessWithPage(c, result.Payments, result.Pagination.Total, req.Page, req.PageSize)
}

// ConfirmPayment godoc
// @Summary 确认回款
// @Description 确认回款记录
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "回款ID"
// @Success 200 {object} common.APIResponse{data=services.PaymentResponse} "确认成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "回款记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments/{id}/confirm [post]
func (h *FinanceHandler) ConfirmPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "回款ID必须是有效数字")
		return
	}

	// 从上下文获取用户ID
	userID := c.GetUint("user_id")

	payment, err := h.paymentService.ConfirmPayment(c.Request.Context(), uint(id), userID)
	if err != nil {
		common.APIInternalServerError(c, "确认回款失败", err.Error())
		return
	}

	common.APISuccess(c, payment)
}

// RejectPayment godoc
// @Summary 拒绝回款
// @Description 拒绝回款记录
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "回款ID"
// @Success 200 {object} common.APIResponse{data=services.PaymentResponse} "拒绝成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "回款记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments/{id}/reject [post]
func (h *FinanceHandler) RejectPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "回款ID必须是有效数字")
		return
	}

	payment, err := h.paymentService.RejectPayment(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "拒绝回款失败", err.Error())
		return
	}

	common.APISuccess(c, payment)
}

// GetPaymentsByInvoiceID godoc
// @Summary 获取发票的回款记录
// @Description 获取指定发票的所有回款记录
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param invoice_id path int true "发票ID"
// @Success 200 {object} common.APIResponse{data=[]services.PaymentResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/invoices/{invoice_id}/payments [get]
func (h *FinanceHandler) GetPaymentsByInvoiceID(c *gin.Context) {
	invoiceIDStr := c.Param("invoice_id")
	invoiceID, err := strconv.ParseUint(invoiceIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "发票ID必须是有效数字")
		return
	}

	payments, err := h.paymentService.GetPaymentsByInvoiceID(c.Request.Context(), uint(invoiceID))
	if err != nil {
		common.APIInternalServerError(c, "查询回款记录失败", err.Error())
		return
	}

	common.APISuccess(c, payments)
}

// GetPaymentStats godoc
// @Summary 获取回款统计信息
// @Description 获取回款统计数据
// @Tags 财务管理-回款
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.PaymentStats} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/payments/stats [get]
func (h *FinanceHandler) GetPaymentStats(c *gin.Context) {
	stats, err := h.paymentService.GetPaymentStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取回款统计失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// ============================================================================
// 坏账核销
// ============================================================================

// CreateBadDebt godoc
// @Summary 创建坏账核销申请
// @Description 创建新的坏账核销申请
// @Tags 财务管理-坏账
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateBadDebtRequest true "创建坏账核销请求"
// @Success 200 {object} common.APIResponse{data=services.BadDebtResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/bad-debts [post]
func (h *FinanceHandler) CreateBadDebt(c *gin.Context) {
	var req services.CreateBadDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	badDebt, err := h.badDebtService.CreateBadDebt(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建坏账核销申请失败", err.Error())
		return
	}

	common.APISuccess(c, badDebt)
}

// GetBadDebt godoc
// @Summary 获取坏账核销详情
// @Description 根据ID获取坏账核销详细信息
// @Tags 财务管理-坏账
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "坏账核销ID"
// @Success 200 {object} common.APIResponse{data=services.BadDebtResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "坏账核销记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/bad-debts/{id} [get]
func (h *FinanceHandler) GetBadDebt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "坏账核销ID必须是有效数字")
		return
	}

	badDebt, err := h.badDebtService.GetBadDebtByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "坏账核销记录不存在", err.Error())
		return
	}

	common.APISuccess(c, badDebt)
}

// ListBadDebts godoc
// @Summary 获取坏账核销列表
// @Description 获取坏账核销列表，支持分页和筛选
// @Tags 财务管理-坏账
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态" Enums(pending, approved, rejected)
// @Param contract_id query int false "合同ID"
// @Param reason_type query string false "原因类型" Enums(bankruptcy, dispute, uncollectible, other)
// @Success 200 {object} common.APIResponse{data=services.ListBadDebtsResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/bad-debts [get]
func (h *FinanceHandler) ListBadDebts(c *gin.Context) {
	var req services.ListBadDebtsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := h.badDebtService.ListBadDebts(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "查询坏账核销列表失败", err.Error())
		return
	}

	common.APISuccessWithPage(c, result.BadDebts, result.Pagination.Total, req.Page, req.PageSize)
}

// ApproveBadDebt godoc
// @Summary 审批通过坏账核销
// @Description 审批通过坏账核销申请
// @Tags 财务管理-坏账
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "坏账核销ID"
// @Param request body object{notes:string} true "审批意见"
// @Success 200 {object} common.APIResponse{data=services.BadDebtResponse} "审批成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "坏账核销记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/bad-debts/{id}/approve [post]
func (h *FinanceHandler) ApproveBadDebt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "坏账核销ID必须是有效数字")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// notes 是可选的
		req.Notes = ""
	}

	// 从上下文获取用户ID
	userID := c.GetUint("user_id")

	badDebt, err := h.badDebtService.ApproveBadDebt(c.Request.Context(), uint(id), userID, req.Notes)
	if err != nil {
		common.APIInternalServerError(c, "审批坏账核销失败", err.Error())
		return
	}

	common.APISuccess(c, badDebt)
}

// RejectBadDebt godoc
// @Summary 审批拒绝坏账核销
// @Description 审批拒绝坏账核销申请
// @Tags 财务管理-坏账
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "坏账核销ID"
// @Param request body object{notes:string} true "拒绝原因"
// @Success 200 {object} common.APIResponse{data=services.BadDebtResponse} "拒绝成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "坏账核销记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/bad-debts/{id}/reject [post]
func (h *FinanceHandler) RejectBadDebt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "坏账核销ID必须是有效数字")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	badDebt, err := h.badDebtService.RejectBadDebt(c.Request.Context(), uint(id), req.Notes)
	if err != nil {
		common.APIInternalServerError(c, "拒绝坏账核销失败", err.Error())
		return
	}

	common.APISuccess(c, badDebt)
}

// GetPendingBadDebts godoc
// @Summary 获取待审批的坏账核销列表
// @Description 获取所有待审批的坏账核销申请
// @Tags 财务管理-坏账
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]services.BadDebtResponse} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/bad-debts/pending [get]
func (h *FinanceHandler) GetPendingBadDebts(c *gin.Context) {
	badDebts, err := h.badDebtService.GetPendingBadDebts(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "查询待审批坏账核销失败", err.Error())
		return
	}

	common.APISuccess(c, badDebts)
}

// ============================================================================
// 提成管理
// ============================================================================

// CalculateCommissions godoc
// @Summary 计算提成
// @Description 根据回款计算提成
// @Tags 财务管理-提成
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CalculateCommissionRequest true "计算提成请求"
// @Success 200 {object} common.APIResponse{data=[]services.CommissionResponse} "计算成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/commissions/calculate [post]
func (h *FinanceHandler) CalculateCommissions(c *gin.Context) {
	var req services.CalculateCommissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	commissions, err := h.commissionService.CalculateCommissions(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "计算提成失败", err.Error())
		return
	}

	common.APISuccess(c, commissions)
}

// GetCommission godoc
// @Summary 获取提成详情
// @Description 根据ID获取提成详细信息
// @Tags 财务管理-提成
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提成ID"
// @Success 200 {object} common.APIResponse{data=services.CommissionResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "提成记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/commissions/{id} [get]
func (h *FinanceHandler) GetCommission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "提成ID必须是有效数字")
		return
	}

	commission, err := h.commissionService.GetCommissionByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "提成记录不存在", err.Error())
		return
	}

	common.APISuccess(c, commission)
}

// ListCommissions godoc
// @Summary 获取提成列表
// @Description 获取提成列表，支持分页和筛选
// @Tags 财务管理-提成
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "状态" Enums(pending, calculated, paid, cancelled)
// @Param beneficiary_id query int false "受益人ID"
// @Param beneficiary_role query string false "受益人角色" Enums(source, lawyer, assistant)
// @Param contract_id query int false "合同ID"
// @Param case_id query int false "案件ID"
// @Param date_from query string false "开始日期"
// @Param date_to query string false "结束日期"
// @Success 200 {object} common.APIResponse{data=services.ListCommissionsResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/commissions [get]
func (h *FinanceHandler) ListCommissions(c *gin.Context) {
	var req services.ListCommissionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	result, err := h.commissionService.ListCommissions(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "查询提成列表失败", err.Error())
		return
	}

	common.APISuccessWithPage(c, result.Commissions, result.Pagination.Total, req.Page, req.PageSize)
}

// MarkCommissionAsPaid godoc
// @Summary 标记提成已支付
// @Description 标记提成为已支付状态
// @Tags 财务管理-提成
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提成ID"
// @Param request body object{paid_date:string,voucher:string} true "支付信息"
// @Success 200 {object} common.APIResponse{data=services.CommissionResponse} "标记成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "提成记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/commissions/{id}/mark-paid [post]
func (h *FinanceHandler) MarkCommissionAsPaid(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "提成ID必须是有效数字")
		return
	}

	var req struct {
		PaidDate string `json:"paid_date" binding:"required"`
		Voucher  string `json:"voucher" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	commission, err := h.commissionService.MarkAsPaid(c.Request.Context(), uint(id), req.PaidDate, req.Voucher)
	if err != nil {
		common.APIInternalServerError(c, "标记提成失败", err.Error())
		return
	}

	common.APISuccess(c, commission)
}

// CancelCommission godoc
// @Summary 取消提成
// @Description 取消指定的提成记录
// @Tags 财务管理-提成
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "提成ID"
// @Success 200 {object} common.APIResponse{data=services.CommissionResponse} "取消成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "提成记录不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/commissions/{id}/cancel [post]
func (h *FinanceHandler) CancelCommission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "提成ID必须是有效数字")
		return
	}

	commission, err := h.commissionService.CancelCommission(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "取消提成失败", err.Error())
		return
	}

	common.APISuccess(c, commission)
}

// GetCommissionsByBeneficiary godoc
// @Summary 获取受益人的提成记录
// @Description 获取指定受益人的所有提成记录
// @Tags 财务管理-提成
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param beneficiary_id path int true "受益人ID"
// @Success 200 {object} common.APIResponse{data=[]services.CommissionResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/commissions/beneficiary/{beneficiary_id} [get]
func (h *FinanceHandler) GetCommissionsByBeneficiary(c *gin.Context) {
	beneficiaryIDStr := c.Param("beneficiary_id")
	beneficiaryID, err := strconv.ParseUint(beneficiaryIDStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "请求参数错误", "受益人ID必须是有效数字")
		return
	}

	commissions, err := h.commissionService.GetCommissionsByBeneficiary(c.Request.Context(), uint(beneficiaryID))
	if err != nil {
		common.APIInternalServerError(c, "查询受益人提成记录失败", err.Error())
		return
	}

	common.APISuccess(c, commissions)
}

// GetCommissionStats godoc
// @Summary 获取提成统计信息
// @Description 获取提成统计数据
// @Tags 财务管理-提成
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.CommissionStats} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/commissions/stats [get]
func (h *FinanceHandler) GetCommissionStats(c *gin.Context) {
	stats, err := h.commissionService.GetCommissionStats(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "获取提成统计失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// ============================================================================
// 财务统计概览
// ============================================================================

// GetFinanceOverview godoc
// @Summary 获取财务概览
// @Description 获取财务数据概览
// @Tags 财务管理-统计
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.FinanceOverview} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /finance/overview [get]
func (h *FinanceHandler) GetFinanceOverview(c *gin.Context) {
	ctx := c.Request.Context()

	// 并行获取各模块统计数据
	type result struct {
		contractStats *services.ContractStats
		invoiceStats  *services.InvoiceStats
		paymentStats  *services.PaymentStats
		commissionStats *services.CommissionStats
		contractErr   error
		invoiceErr    error
		paymentErr    error
		commissionErr error
	}

	ch := make(chan result, 1)

	go func() {
		contractStats, contractErr := h.contractService.GetContractStats(ctx)
		invoiceStats, invoiceErr := h.invoiceService.GetInvoiceStats(ctx)
		paymentStats, paymentErr := h.paymentService.GetPaymentStats(ctx)
		commissionStats, commissionErr := h.commissionService.GetCommissionStats(ctx)
		ch <- result{contractStats, invoiceStats, paymentStats, commissionStats, contractErr, invoiceErr, paymentErr, commissionErr}
	}()

	r := <-ch

	// 处理错误
	if r.contractErr != nil {
		common.APIInternalServerError(c, "获取合同统计失败", r.contractErr.Error())
		return
	}
	if r.invoiceErr != nil {
		common.APIInternalServerError(c, "获取发票统计失败", r.invoiceErr.Error())
		return
	}
	if r.paymentErr != nil {
		common.APIInternalServerError(c, "获取回款统计失败", r.paymentErr.Error())
		return
	}
	if r.commissionErr != nil {
		common.APIInternalServerError(c, "获取提成统计失败", r.commissionErr.Error())
		return
	}

	overview := &services.FinanceOverview{
		ContractStats:   r.contractStats,
		InvoiceStats:    r.invoiceStats,
		PaymentStats:    r.paymentStats,
		CommissionStats: r.commissionStats,
	}

	common.APISuccess(c, overview)
}
