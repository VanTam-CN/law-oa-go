package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApprovalHandler struct {
	db              *gorm.DB
	approvalService *services.ApprovalService
}

func NewApprovalHandler(db *gorm.DB) *ApprovalHandler {
	return &ApprovalHandler{
		db:              db,
		approvalService: services.NewApprovalService(db),
	}
}

// GetApprovalService 获取审批服务（供集成服务使用）
func (h *ApprovalHandler) GetApprovalService() services.ApprovalServiceInterface {
	// 创建一个适配器将ApprovalService转换为ApprovalServiceInterface
	return &approvalServiceAdapter{service: h.approvalService}
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
	fmt.Printf("🔍 调试 GetApprovalStats: userID类型=%T, 值=%v\n", userID, userID) // 临时调试日志
	var userIDStr string
	switch v := userID.(type) {
	case uint:
		fmt.Printf("✅ 匹配uint类型: %v\n", v)
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		fmt.Printf("✅ 匹配int类型: %v\n", v)
		userIDStr = strconv.Itoa(v)
	case float64: // 修复：JSON解析的数字默认是float64类型
		fmt.Printf("✅ 匹配float64类型: %v\n", v)
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		fmt.Printf("✅ 匹配string类型: %v\n", v)
		userIDStr = v
	default:
		fmt.Printf("❌ 未匹配的类型: %v (类型: %T)\n", v, v)
		common.Error(c, http.StatusUnauthorized, "用户ID格式错误")
		return
	}
	fmt.Printf("🔍 调试: 转换后的userIDStr=%s\n", userIDStr) // 临时调试日志

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

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize > 100 {
		pageSize = 100
	}

	req := models.ApprovalListRequest{
		Page:        page,
		PageSize:    pageSize,
		Status:      models.ApprovalStatusSubmitted, // 只获取已提交的
		ApplicantID: "", // 不限定申请人
	}

	approvals, err := h.approvalService.GetPendingApprovals(userIDStr, &req)
	if err != nil {
		log.Printf("获取待审批列表失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "获取待审批列表失败")
		return
	}

	common.APISuccess(c, approvals)
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
		ApplicantID: c.Query("applicantId"), // 修复参数名，使用前端传递的applicantId
		Keyword:     c.Query("keyword"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		SortBy:      c.DefaultQuery("sort_by", "created_at"),
		SortOrder:   c.DefaultQuery("sort_order", "desc"),
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

	approval, err := h.approvalService.GetApproval(userIDStr, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			common.Error(c, http.StatusNotFound, "审批记录不存在")
		} else {
			log.Printf("获取审批详情失败: %v", err)
			common.Error(c, http.StatusInternalServerError, "获取审批详情失败")
		}
		return
	}

	common.APISuccess(c, approval)
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