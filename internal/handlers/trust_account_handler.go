package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// TrustAccountHandler 代管款账户处理器
type TrustAccountHandler struct {
	accountService     *services.TrustAccountService
	transactionService *services.TrustTransactionService
}

// NewTrustAccountHandler 创建代管款账户处理器
func NewTrustAccountHandler(
	accountService *services.TrustAccountService,
	transactionService *services.TrustTransactionService,
) *TrustAccountHandler {
	return &TrustAccountHandler{
		accountService:     accountService,
		transactionService: transactionService,
	}
}

// ============================================================================
// 账户相关处理器
// ============================================================================

// CreateAccount godoc
// @Summary 创建代管款账户
// @Description 为客户创建新的代管款账户
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body services.CreateAccountRequest true "创建账户请求"
// @Success 200 {object} common.APIResponse{data=services.AccountResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/accounts [post]
func (h *TrustAccountHandler) CreateAccount(c *gin.Context) {
	var req services.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), &req)
	if err != nil {
		common.APIInternalServerError(c, "创建账户失败", err.Error())
		return
	}

	common.APISuccess(c, account)
}

// GetAccount godoc
// @Summary 获取账户详情
// @Description 根据ID获取代管款账户详情
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账户ID"
// @Success 200 {object} common.APIResponse{data=services.AccountResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 404 {object} common.APIResponse "账户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/accounts/:id [get]
func (h *TrustAccountHandler) GetAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的账户ID", err.Error())
		return
	}

	account, err := h.accountService.GetAccountByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取账户失败", err.Error())
		return
	}

	common.APISuccess(c, account)
}

// ListAccounts godoc
// @Summary 获取账户列表
// @Description 分页获取代管款账户列表
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param client_id query int false "客户ID"
// @Param status query string false "账户状态" Enums(active, frozen, closed)
// @Param currency query string false "币种" Enums(CNY, USD, EUR)
// @Success 200 {object} common.APIResponse{data=services.ListAccountsResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/accounts [get]
func (h *TrustAccountHandler) ListAccounts(c *gin.Context) {
	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	var clientID uint
	if cid := c.Query("client_id"); cid != "" {
		if parsed, err := strconv.ParseUint(cid, 10, 32); err == nil {
			clientID = uint(parsed)
		}
	}

	req := &services.ListAccountsRequest{
		Page:     page,
		PageSize: pageSize,
		ClientID: clientID,
		Status:   c.Query("status"),
		Currency: c.Query("currency"),
	}

	accounts, err := h.accountService.ListAccounts(c.Request.Context(), req)
	if err != nil {
		common.APIInternalServerError(c, "获取账户列表失败", err.Error())
		return
	}

	common.APISuccess(c, accounts)
}

// FreezeAccount godoc
// @Summary 冻结账户
// @Description 冻结指定的代管款账户
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账户ID"
// @Success 200 {object} common.APIResponse{data=services.AccountResponse} "冻结成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 404 {object} common.APIResponse "账户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/accounts/:id/freeze [post]
func (h *TrustAccountHandler) FreezeAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的账户ID", err.Error())
		return
	}

	account, err := h.accountService.FreezeAccount(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "冻结账户失败", err.Error())
		return
	}

	common.APISuccess(c, account)
}

// UnfreezeAccount godoc
// @Summary 解冻账户
// @Description 解冻指定的代管款账户
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账户ID"
// @Success 200 {object} common.APIResponse{data=services.AccountResponse} "解冻成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 404 {object} common.APIResponse "账户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/accounts/:id/unfreeze [post]
func (h *TrustAccountHandler) UnfreezeAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的账户ID", err.Error())
		return
	}

	account, err := h.accountService.UnfreezeAccount(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "解冻账户失败", err.Error())
		return
	}

	common.APISuccess(c, account)
}

// CloseAccount godoc
// @Summary 关闭账户
// @Description 关闭指定的代管款账户
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账户ID"
// @Success 200 {object} common.APIResponse{data=services.AccountResponse} "关闭成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 404 {object} common.APIResponse "账户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/accounts/:id/close [post]
func (h *TrustAccountHandler) CloseAccount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的账户ID", err.Error())
		return
	}

	account, err := h.accountService.CloseAccount(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "关闭账户失败", err.Error())
		return
	}

	common.APISuccess(c, account)
}

// ============================================================================
// 交易相关处理器
// ============================================================================

// CreateTransaction godoc
// @Summary 创建交易
// @Description 创建新的代管款交易
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body services.CreateTransactionRequest true "创建交易请求"
// @Success 200 {object} common.APIResponse{data=services.TransactionResponse} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/transactions [post]
func (h *TrustAccountHandler) CreateTransaction(c *gin.Context) {
	var req services.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未认证", "用户信息无效")
		return
	}

	transaction, err := h.transactionService.CreateTransaction(
		c.Request.Context(),
		&req,
		userID.(uint),
	)
	if err != nil {
		common.APIInternalServerError(c, "创建交易失败", err.Error())
		return
	}

	common.APISuccess(c, transaction)
}

// GetTransaction godoc
// @Summary 获取交易详情
// @Description 根据ID获取交易详情
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "交易ID"
// @Success 200 {object} common.APIResponse{data=services.TransactionResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 404 {object} common.APIResponse "交易不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/transactions/:id [get]
func (h *TrustAccountHandler) GetTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的交易ID", err.Error())
		return
	}

	transaction, err := h.transactionService.GetTransactionByID(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "获取交易失败", err.Error())
		return
	}

	common.APISuccess(c, transaction)
}

// ListTransactions godoc
// @Summary 获取交易列表
// @Description 分页获取交易列表
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param account_id query int false "账户ID"
// @Success 200 {object} common.APIResponse{data=[]services.TransactionResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/transactions [get]
func (h *TrustAccountHandler) ListTransactions(c *gin.Context) {
	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	var accountID *uint
	if aid := c.Query("account_id"); aid != "" {
		if parsed, err := strconv.ParseUint(aid, 10, 32); err == nil {
			id := uint(parsed)
			accountID = &id
		}
	}

	transactions, total, err := h.transactionService.ListTransactions(
		c.Request.Context(),
		page,
		pageSize,
		accountID,
	)
	if err != nil {
		common.APIInternalServerError(c, "获取交易列表失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"transactions": transactions,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// ApproveTransaction godoc
// @Summary 审批通过交易
// @Description 审批通过指定的交易
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "交易ID"
// @Success 200 {object} common.APIResponse{data=services.TransactionResponse} "审批成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 404 {object} common.APIResponse "交易不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/transactions/:id/approve [post]
func (h *TrustAccountHandler) ApproveTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的交易ID", err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未认证", "用户信息无效")
		return
	}

	transaction, err := h.transactionService.ApproveTransaction(
		c.Request.Context(),
		uint(id),
		userID.(uint),
	)
	if err != nil {
		common.APIInternalServerError(c, "审批交易失败", err.Error())
		return
	}

	common.APISuccess(c, transaction)
}

// RejectTransaction godoc
// @Summary 审批拒绝交易
// @Description 审批拒绝指定的交易
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "交易ID"
// @Success 200 {object} common.APIResponse{data=services.TransactionResponse} "拒绝成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 404 {object} common.APIResponse "交易不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/transactions/:id/reject [post]
func (h *TrustAccountHandler) RejectTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的交易ID", err.Error())
		return
	}

	transaction, err := h.transactionService.RejectTransaction(
		c.Request.Context(),
		uint(id),
	)
	if err != nil {
		common.APIInternalServerError(c, "拒绝交易失败", err.Error())
		return
	}

	common.APISuccess(c, transaction)
}

// GetAccountTransactions godoc
// @Summary 获取账户交易记录
// @Description 获取指定账户的交易记录
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "账户ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} common.APIResponse{data=[]services.TransactionSummary} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/accounts/:id/transactions [get]
func (h *TrustAccountHandler) GetAccountTransactions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.APIBadRequest(c, "无效的账户ID", err.Error())
		return
	}

	page := 1
	pageSize := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	accountID := uint(id)
	transactions, total, err := h.transactionService.ListTransactions(
		c.Request.Context(),
		page,
		pageSize,
		&accountID,
	)
	if err != nil {
		common.APIInternalServerError(c, "获取交易记录失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"transactions": transactions,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// GetAccountStats godoc
// @Summary 获取账户统计
// @Description 获取代管款账户的统计数据
// @Tags 代管款管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.APIResponse{data=map[string]interface{}} "获取成功"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /api/v1/trust/stats [get]
func (h *TrustAccountHandler) GetAccountStats(c *gin.Context) {
	// 获取所有账户进行统计
	allAccounts, err := h.accountService.ListAccounts(
		c.Request.Context(),
		&services.ListAccountsRequest{
			Page:     1,
			PageSize: 1000,
		},
	)
	if err != nil {
		common.APIInternalServerError(c, "获取统计数据失败", err.Error())
		return
	}

	totalAccounts := len(allAccounts.Accounts)
	totalBalance := 0.0
	totalFrozen := 0.0
	activeAccounts := 0

	for _, acc := range allAccounts.Accounts {
		totalBalance += acc.Balance
		totalFrozen += acc.FrozenAmount
		if acc.Status == "active" {
			activeAccounts++
		}
	}

	common.APISuccess(c, gin.H{
		"total_accounts":  totalAccounts,
		"total_balance":   totalBalance,
		"total_frozen":    totalFrozen,
		"active_accounts": activeAccounts,
	})
}
