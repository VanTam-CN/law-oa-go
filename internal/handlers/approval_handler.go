package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApprovalHandler struct {
	db                      *gorm.DB
	approvalService         *services.ApprovalService
	approvalTemplateService *services.ApprovalTemplateService
	integratedDecision      services.ApprovalConflictIntegrationService
	authorization           *services.AuthorizationService
}

func NewApprovalHandler(db *gorm.DB) *ApprovalHandler {
	return &ApprovalHandler{
		db:                      db,
		approvalService:         services.NewApprovalService(db),
		approvalTemplateService: services.NewApprovalTemplateService(db),
	}
}

// GetApprovalService 获取审批服务（供集成服务使用）
func (h *ApprovalHandler) GetApprovalService() services.ApprovalServiceInterface {
	// 创建一个适配器将ApprovalService转换为ApprovalServiceInterface
	return &approvalServiceAdapter{service: h.approvalService}
}

// SetIntegratedDecisionService routes case-creation and conflict approvals
// through the same preflight and idempotent post-approval gate used by the
// dedicated integration endpoint. Generic approvals keep the normal path.
func (h *ApprovalHandler) SetIntegratedDecisionService(service services.ApprovalConflictIntegrationService) {
	h.integratedDecision = service
}

// SetAuthorizationService installs the object-level gate used by controlled
// conflict and case-creation approvals. The constructor remains compatible
// with older tests and callers; the production router always sets this.
func (h *ApprovalHandler) SetAuthorizationService(authz *services.AuthorizationService) {
	h.authorization = authz
}

// approvalServiceAdapter ApprovalService适配器
type approvalServiceAdapter struct {
	service *services.ApprovalService
}

func (a *approvalServiceAdapter) CreateApproval(userID string, userName string, req *models.CreateApprovalRequest) (*models.ApprovalRequest, error) {
	return a.service.CreateApproval(userID, userName, req)
}

func (a *approvalServiceAdapter) GetApproval(userID string, id string) (*models.ApprovalRequest, error) {
	return a.service.GetApproval(userID, id)
}

func (a *approvalServiceAdapter) GetApprovalByID(id string) (*models.ApprovalRequest, error) {
	return a.service.GetApprovalByID(id)
}

func (a *approvalServiceAdapter) SubmitApproval(userID string, id string) (*models.ApprovalRequest, error) {
	return a.service.SubmitApproval(userID, id)
}

func (a *approvalServiceAdapter) ProcessApproval(userID string, userName string, id string, decisionReq *models.ApprovalDecisionRequest) (*models.ApprovalRequest, error) {
	return a.service.ProcessApprovalDecision(userID, id, decisionReq)
}

// GetApprovalStats 获取审批统计
func (h *ApprovalHandler) GetApprovalStats(c *gin.Context) {
	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64: // 修复：JSON解析的数字默认是float64类型
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}

	// Firm-wide approval counts are not safe until they are filtered by the
	// ethical-wall subject context. The wall-aware approval workbench is the
	// only endpoint allowed to expose management aggregates; this endpoint
	// remains user-scoped for every role.
	stats, err := h.approvalService.GetApprovalStats(userIDStr)
	if err != nil {
		log.Printf("获取审批统计失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取审批统计失败")
		return
	}

	common.APISuccess(c, stats)
}

// GetPendingApprovals 获取待审批列表
func (h *ApprovalHandler) GetPendingApprovals(c *gin.Context) {
	// The legacy repository query paginates before it can resolve a
	// conflict-bound approval's subject case. Returning its raw rows or total
	// would leak wall-protected titles and counts. The active workbench applies
	// object-level visibility before projecting rows.
	common.NewAPIError(c, http.StatusServiceUnavailable, "APPROVAL_PENDING_UNAVAILABLE", "旧版待审批列表未接入案件隔离墙过滤，请使用审批工作台")
}

// ListApprovals 获取审批列表
func (h *ApprovalHandler) ListApprovals(c *gin.Context) {
	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64: // 修复：JSON解析的数字默认是float64类型
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}
	if services.IsBusinessMatterManagementRole(c.GetString("role")) {
		// The legacy repository query cannot prove ethical-wall visibility for
		// conflict-bound rows. Management must use /approvals/workbench, which
		// resolves the subject case before returning a row or aggregate count.
		common.NewAPIError(c, http.StatusServiceUnavailable, "APPROVAL_LIST_UNAVAILABLE", "全所审批清单必须通过受隔离墙保护的审批工作台访问")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}

	req := models.ApprovalListRequest{
		Page:        page,
		PageSize:    pageSize,
		Status:      c.Query("status"),
		Type:        c.Query("type"),
		ApplicantID: c.Query("applicantId"), // 管理角色可按申请人筛选
		Keyword:     c.Query("keyword"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		SortBy:      c.DefaultQuery("sort_by", "created_at"),
		SortOrder:   c.DefaultQuery("sort_order", "desc"),
	}
	if !services.IsBusinessMatterManagementRole(c.GetString("role")) {
		req.ApplicantID = userIDStr
	}

	approvals, err := h.approvalService.ListApprovals(userIDStr, &req)
	if err != nil {
		log.Printf("获取审批列表失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取审批列表失败")
		return
	}

	common.APISuccess(c, approvals)
}

// GetApproval 获取单个审批详情
func (h *ApprovalHandler) GetApproval(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.Error(c, http.StatusBadRequest, "审批ID不能为空")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64: // 修复：JSON解析的数字默认是float64类型
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}

	approval, err := h.approvalService.GetApprovalForAuthorization(userIDStr, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound || strings.Contains(err.Error(), "record not found") {
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		} else if strings.Contains(err.Error(), "无权") {
			common.Error(c, http.StatusForbidden, "无权查看此审批记录")
		} else {
			log.Printf("获取审批详情失败: %v", err)
			common.Error(c, http.StatusInternalServerError, "获取审批详情失败")
		}
		return
	}
	if !h.authorizeApprovalConflictContext(c, approval) {
		return
	}
	if err := h.approvalService.LoadApprovalRecords(approval); err != nil {
		log.Printf("获取审批记录失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取审批记录失败")
		return
	}

	common.APISuccess(c, projectApprovalForViewer(approval, c.GetString("role")))
}

// authorizeApprovalMutation re-checks the object boundary immediately before
// a legacy approval mutation. An approval may outlive a case assignment or
// ethical-wall change, so the applicant check in ApprovalService is not a
// sufficient substitute for the underlying case/intake authorization.
func (h *ApprovalHandler) authorizeApprovalMutation(c *gin.Context, userID, approvalID string) bool {
	approval, err := h.approvalService.GetApprovalForAuthorization(userID, approvalID)
	if err != nil {
		if strings.Contains(err.Error(), "无权") {
			common.Error(c, http.StatusForbidden, "无权操作此审批记录")
		} else if err == gorm.ErrRecordNotFound || strings.Contains(err.Error(), "record not found") {
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		} else {
			log.Printf("获取审批上下文失败: %v", err)
			common.Error(c, http.StatusInternalServerError, "获取审批记录失败")
		}
		return false
	}
	return h.authorizeApprovalConflictContext(c, approval)
}

// GetApprovalWorkflows 获取审批工作流列表
func (h *ApprovalHandler) GetApprovalWorkflows(c *gin.Context) {
	workflows, err := h.approvalService.GetApprovalWorkflows()
	if err != nil {
		log.Printf("获取审批工作流失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取审批工作流失败")
		return
	}

	common.APISuccess(c, workflows)
}

// GetApprovalTemplates 获取审批模板列表
func (h *ApprovalHandler) GetApprovalTemplates(c *gin.Context) {
	templateType := c.Query("template_type")
	category := c.Query("category")

	templates, err := h.approvalService.GetApprovalTemplates(templateType, category)
	if err != nil {
		log.Printf("获取审批模板失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取审批模板失败")
		return
	}

	common.APISuccess(c, templates)
}

// CreateApproval 创建审批申请
func (h *ApprovalHandler) CreateApproval(c *gin.Context) {
	// 获取当前用户信息（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64: // 修复：JSON解析的数字默认是float64类型
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}

	// 解析请求数据
	var req models.CreateApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		common.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}
	typeName := strings.ToLower(strings.TrimSpace(req.Type))
	workflowType := strings.ToLower(strings.TrimSpace(req.WorkflowType))
	if typeName == "case_creation" || typeName == "conflict" || typeName == "conflict_approval" ||
		workflowType == "case_creation" || workflowType == "conflict_approval" {
		common.NewAPIError(c, http.StatusConflict, "APPROVAL_CONTROLLED_WORKFLOW_REQUIRED", "成案和冲突审批必须从接案工作台或带案件上下文的受控流程发起")
		return
	}

	// 创建审批申请（Service层会从数据库获取用户信息）
	approval, err := h.approvalService.CreateApproval(userIDStr, "", &req)
	if err != nil {
		log.Printf("创建审批申请失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "创建审批申请失败")
		return
	}

	common.APISuccess(c, approval)
}

// SubmitApproval 提交审批申请（从草稿状态提交）
func (h *ApprovalHandler) SubmitApproval(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.Error(c, http.StatusBadRequest, "审批ID不能为空")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}
	if !h.authorizeApprovalMutation(c, userIDStr, id) {
		return
	}

	// 调用服务提交审批
	approval, err := h.approvalService.SubmitApproval(userIDStr, id)
	if err != nil {
		log.Printf("提交审批失败: %v", err)
		switch {
		case err.Error() == "审批记录不存在":
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		case err.Error() == "只有申请人才能提交审批":
			common.Error(c, http.StatusForbidden, "只有申请人才能提交审批")
		case err.Error() == "草稿状态才能提交":
			common.Error(c, http.StatusBadRequest, "只有草稿状态的审批才能提交")
		default:
			common.Error(c, http.StatusInternalServerError, "提交审批失败")
		}
		return
	}

	common.APISuccess(c, approval)
}

// ProcessApprovalDecision 处理审批决定（批准/拒绝/要求修改/转派）
func (h *ApprovalHandler) ProcessApprovalDecision(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.Error(c, http.StatusBadRequest, "审批ID不能为空")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}
	// 解析请求数据
	var req models.ApprovalDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		common.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if h.integratedDecision != nil {
		existing, lookupErr := h.approvalService.GetApprovalForAuthorization(userIDStr, id)
		if lookupErr != nil {
			common.Error(c, http.StatusInternalServerError, "获取审批记录失败")
			return
		}
		if !h.authorizeApprovalConflictContext(c, existing) {
			return
		}
		if services.RequiresCaseCreationApproval(existing) {
			userName := c.GetString("username")
			if userName == "" {
				userName = "未知用户"
			}
			approval, err := h.integratedDecision.ProcessApprovalWithConflict(c.Request.Context(), userIDStr, userName, id, &req)
			if err != nil {
				log.Printf("处理受控成案审批决定失败: %v", err)
				h.writeApprovalDecisionError(c, err)
				return
			}
			common.APISuccess(c, approval)
			return
		}
	} else {
		existing, lookupErr := h.approvalService.GetApprovalForAuthorization(userIDStr, id)
		if lookupErr != nil {
			common.Error(c, http.StatusInternalServerError, "获取审批记录失败")
			return
		}
		if !h.authorizeApprovalConflictContext(c, existing) {
			return
		}
	}

	// 调用服务处理审批决定
	approval, err := h.approvalService.ProcessApprovalDecision(userIDStr, id, &req)
	if err != nil {
		log.Printf("处理审批决定失败: %v", err)
		switch {
		case err.Error() == "审批记录不存在":
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		case err.Error() == "无权审批此申请":
			common.Error(c, http.StatusForbidden, "无权审批此申请")
		case err.Error() == "审批状态不允许此操作":
			common.Error(c, http.StatusBadRequest, "当前状态不允许此操作")
		case err.Error() == "审批决定类型无效":
			common.Error(c, http.StatusBadRequest, "无效的审批决定类型")
		default:
			common.Error(c, http.StatusInternalServerError, "处理审批决定失败")
		}
		return
	}

	common.APISuccess(c, approval)
}

func (h *ApprovalHandler) writeApprovalDecisionError(c *gin.Context, err error) {
	switch err.Error() {
	case "审批记录不存在":
		common.Error(c, http.StatusNotFound, "审批记录不存在")
	case "无权审批此申请":
		common.Error(c, http.StatusForbidden, "无权审批此申请")
	case "审批状态不允许此操作", "当前状态不允许此操作":
		common.Error(c, http.StatusBadRequest, "当前状态不允许此操作")
	case "审批决定类型无效":
		common.Error(c, http.StatusBadRequest, "无效的审批决定类型")
	default:
		common.Error(c, http.StatusConflict, err.Error())
	}
}

// ResubmitApproval 重新提交被驳回的审批申请
func (h *ApprovalHandler) ResubmitApproval(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.Error(c, http.StatusBadRequest, "审批ID不能为空")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}
	if !h.authorizeApprovalMutation(c, userIDStr, id) {
		return
	}

	// 解析请求数据
	var req struct {
		RevisionNote string `json:"revision_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		common.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 调用服务重新提交
	approval, err := h.approvalService.ResubmitApproval(userIDStr, id, req.RevisionNote)
	if err != nil {
		log.Printf("重新提交审批失败: %v", err)
		switch {
		case err.Error() == "审批记录不存在":
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		case err.Error() == "只有申请人才能重新提交":
			common.Error(c, http.StatusForbidden, "只有申请人才能重新提交")
		case err.Error() == "只有被驳回或需要修改的审批才能重新提交":
			common.Error(c, http.StatusBadRequest, "当前状态不允许重新提交")
		default:
			common.Error(c, http.StatusInternalServerError, "重新提交审批失败")
		}
		return
	}

	common.APISuccess(c, approval)
}

// CancelApproval 取消审批申请
func (h *ApprovalHandler) CancelApproval(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.Error(c, http.StatusBadRequest, "审批ID不能为空")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}
	if !h.authorizeApprovalMutation(c, userIDStr, id) {
		return
	}

	// 调用服务取消审批
	err := h.approvalService.CancelApproval(userIDStr, id)
	if err != nil {
		log.Printf("取消审批失败: %v", err)
		switch {
		case err.Error() == "审批记录不存在":
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		case err.Error() == "只有申请人才能取消审批":
			common.Error(c, http.StatusForbidden, "只有申请人才能取消审批")
		case err.Error() == "已有审批记录，无法取消":
			common.Error(c, http.StatusBadRequest, "已有审批记录，无法取消")
		case err.Error() == "当前状态不允许取消":
			common.Error(c, http.StatusBadRequest, "当前状态不允许取消")
		default:
			common.Error(c, http.StatusInternalServerError, "取消审批失败")
		}
		return
	}

	common.APISuccess(c, gin.H{"message": "审批已取消"})
}

// UpdateApproval 更新审批申请（仅草稿或需要修改状态）
func (h *ApprovalHandler) UpdateApproval(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.Error(c, http.StatusBadRequest, "审批ID不能为空")
		return
	}

	// 获取当前用户ID（从JWT中获取）
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}
	if !h.authorizeApprovalMutation(c, userIDStr, id) {
		return
	}

	// 解析请求数据
	var req models.UpdateApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		common.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 调用服务更新审批
	approval, err := h.approvalService.UpdateApproval(userIDStr, id, &req)
	if err != nil {
		log.Printf("更新审批失败: %v", err)
		switch {
		case err.Error() == "审批记录不存在":
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		case err.Error() == "只有申请人才能更新审批":
			common.Error(c, http.StatusForbidden, "只有申请人才能更新审批")
		case err.Error() == "只有草稿或需要修改状态的审批才能更新":
			common.Error(c, http.StatusBadRequest, "当前状态不允许更新")
		default:
			common.Error(c, http.StatusInternalServerError, "更新审批失败")
		}
		return
	}

	common.APISuccess(c, approval)
}

// CreateFromTemplate 从模板创建审批
func (h *ApprovalHandler) CreateFromTemplate(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}

	// 解析请求数据
	var req struct {
		TemplateName string                   `json:"template_name" binding:"required"`
		Title        string                   `json:"title" binding:"required"`
		Type         string                   `json:"type" binding:"required"`
		Category     string                   `json:"category"`
		Content      string                   `json:"content" binding:"required"`
		Metadata     map[string]interface{}   `json:"metadata"`
		Attachments  []map[string]interface{} `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		common.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 构建创建请求
	createReq := &models.CreateApprovalRequest{
		Title:        req.Title,
		Type:         req.Type,
		Category:     req.Category,
		Content:      req.Content,
		WorkflowType: req.TemplateName,
		Metadata:     req.Metadata,
		Attachments:  req.Attachments,
	}

	// 从模板创建审批
	approval, err := h.approvalTemplateService.CreateFromTemplate(req.TemplateName, userIDStr, "", createReq, req.Metadata)
	if err != nil {
		log.Printf("从模板创建审批失败: %v", err)
		common.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.APISuccess(c, approval)
}

// GetApprovalFlow 获取审批流程
func (h *ApprovalHandler) GetApprovalFlow(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.Error(c, http.StatusBadRequest, "审批ID不能为空")
		return
	}

	flow, err := h.approvalTemplateService.GetApprovalFlow(id)
	if err != nil {
		log.Printf("获取审批流程失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取审批流程失败")
		return
	}

	common.APISuccess(c, flow)
}

// ProcessNode 处理审批节点
func (h *ApprovalHandler) ProcessNode(c *gin.Context) {
	approvalID := c.Param("id")
	nodeIDStr := c.Param("nodeId")

	if approvalID == "" || nodeIDStr == "" {
		common.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		common.Error(c, http.StatusUnauthorized, "未授权访问")
		return
	}

	var userIDStr string
	switch v := userID.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}

	nodeID, err := strconv.ParseUint(nodeIDStr, 10, 64)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "节点ID格式错误")
		return
	}

	// 解析请求
	var req struct {
		Action  string `json:"action" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	node, err := h.approvalTemplateService.ProcessNode(uint(nodeID), req.Action, req.Comment, userIDStr)
	if err != nil {
		log.Printf("处理审批节点失败: %v", err)
		common.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	common.APISuccess(c, node)
}

// GetAllTemplates 获取所有审批模板
func (h *ApprovalHandler) GetAllTemplates(c *gin.Context) {
	templates, err := h.approvalTemplateService.GetAllTemplates()
	if err != nil {
		log.Printf("获取审批模板失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取审批模板失败")
		return
	}

	common.APISuccess(c, templates)
}

// InitializeTemplates 初始化默认模板
func (h *ApprovalHandler) InitializeTemplates(c *gin.Context) {
	if err := h.approvalTemplateService.InitializeDefaultTemplates(); err != nil {
		log.Printf("初始化默认模板失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "初始化默认模板失败")
		return
	}

	common.APISuccess(c, gin.H{"message": "默认模板初始化成功"})
}

// SupportCountersign 支持会签
func (h *ApprovalHandler) SupportCountersign(c *gin.Context) {
	approvalID := c.Param("id")
	stepOrderStr := c.Param("stepOrder")

	if approvalID == "" || stepOrderStr == "" {
		common.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	stepOrder, err := strconv.Atoi(stepOrderStr)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "步骤序号格式错误")
		return
	}

	if err := h.approvalTemplateService.SupportCountersign(approvalID, stepOrder); err != nil {
		log.Printf("设置会签失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "设置会签失败")
		return
	}

	common.APISuccess(c, gin.H{"message": "会签设置成功"})
}

// SupportOrSign 支持或签
func (h *ApprovalHandler) SupportOrSign(c *gin.Context) {
	approvalID := c.Param("id")
	stepOrderStr := c.Param("stepOrder")

	if approvalID == "" || stepOrderStr == "" {
		common.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	stepOrder, err := strconv.Atoi(stepOrderStr)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "步骤序号格式错误")
		return
	}

	if err := h.approvalTemplateService.SupportOrSign(approvalID, stepOrder); err != nil {
		log.Printf("设置或签失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "设置或签失败")
		return
	}

	common.APISuccess(c, gin.H{"message": "或签设置成功"})
}

// ReturnToPrevious 退回到上一步
func (h *ApprovalHandler) ReturnToPrevious(c *gin.Context) {
	approvalID := c.Param("id")
	stepOrderStr := c.Param("stepOrder")

	if approvalID == "" || stepOrderStr == "" {
		common.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	stepOrder, err := strconv.Atoi(stepOrderStr)
	if err != nil {
		common.Error(c, http.StatusBadRequest, "步骤序号格式错误")
		return
	}

	if err := h.approvalTemplateService.ReturnToPrevious(approvalID, stepOrder); err != nil {
		log.Printf("退回失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "退回失败")
		return
	}

	common.APISuccess(c, gin.H{"message": "退回成功"})
}
