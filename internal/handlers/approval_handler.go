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
	userName, _ := c.Get("user_name")
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

	// 处理用户名的多种类型
	var userNameStr string
	if name, ok := userName.(string); ok {
		userNameStr = name
	} else {
		userNameStr = "未知用户"
	}

	// 解析请求数据
	var req models.CreateApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("请求参数错误: %v", err)
		common.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 创建审批申请
	approval, err := h.approvalService.CreateApproval(userIDStr, userNameStr, &req)
	if err != nil {
		log.Printf("创建审批申请失败: %v", err)
		common.Error(c, http.StatusInternalServerError, "创建审批申请失败")
		return
	}

	common.APISuccess(c, approval)
}