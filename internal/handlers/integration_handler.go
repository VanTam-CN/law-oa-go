package handlers

import (
	"net/http"
	"strconv"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

// IntegrationHandler 集成处理器
type IntegrationHandler struct {
	integrationService services.ApprovalConflictIntegrationService
	conflictService    services.ConflictDetectionServiceInterface
}

// NewIntegrationHandler 创建集成处理器
func NewIntegrationHandler(
	integrationService services.ApprovalConflictIntegrationService,
	conflictService services.ConflictDetectionServiceInterface,
) *IntegrationHandler {
	return &IntegrationHandler{
		integrationService: integrationService,
		conflictService:    conflictService,
	}
}

// CreateIntegratedApproval 创建集成的审批申请
func (h *IntegrationHandler) CreateIntegratedApproval(c *gin.Context) {
	var req services.IntegrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 从上下文获取用户信息
	userID := c.GetString("userID")
	userName := c.GetString("userName")

	// 调用集成服务
	result, err := h.integrationService.CreateIntegratedApproval(c.Request.Context(), userID, userName, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "创建集成审批申请失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// CreateApprovalWithConflict 创建带有冲突检测的审批申请
func (h *IntegrationHandler) CreateApprovalWithConflict(c *gin.Context) {
	var req services.IntegrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 从上下文获取用户信息（JWT中间件设置的是user_id和username）
	userIDInt, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未授权访问",
		})
		return
	}

	// 处理用户ID的多种类型
	var userIDStr string
	switch v := userIDInt.(type) {
	case uint:
		userIDStr = strconv.FormatUint(uint64(v), 10)
	case int:
		userIDStr = strconv.Itoa(v)
	case float64:
		userIDStr = strconv.FormatInt(int64(v), 10)
	case string:
		userIDStr = v
	default:
		userIDStr = "1"
	}

	userName, _ := c.Get("username")
	if userName == nil {
		userName = "未知用户"
	}
	userNameStr := userName.(string)

	// 调用集成服务
	result, err := h.integrationService.CreateIntegratedApproval(c.Request.Context(), userIDStr, userNameStr, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "创建带冲突检测的审批申请失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetApprovalIntegrationStatus 获取审批集成状态
func (h *IntegrationHandler) GetApprovalIntegrationStatus(c *gin.Context) {
	approvalID := c.Param("id")
	if approvalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "审批ID不能为空",
		})
		return
	}

	// 调用集成服务
	status, err := h.integrationService.GetIntegrationStatus(c.Request.Context(), approvalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "获取集成状态失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// PerformCaseCreation 执行案件创建
func (h *IntegrationHandler) PerformCaseCreation(c *gin.Context) {
	approvalID := c.Param("id")
	if approvalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "审批ID不能为空",
		})
		return
	}

	var req struct {
		CaseData map[string]interface{} `json:"case_data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 调用集成服务
	result, err := h.integrationService.AutoCreateCaseFromApproval(c.Request.Context(), approvalID, req.CaseData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "创建案件失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// TriggerConflictCheck 触发冲突检测
func (h *IntegrationHandler) TriggerConflictCheck(c *gin.Context) {
	var req models.ConflictCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 设置用户ID
	userID := c.GetString("userID")
	if userID == "" {
		userID = "unknown"
	}

	// 执行冲突检测
	result, err := h.conflictService.PerformConflictCheck(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "冲突检测失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetIntegrationStatistics 获取集成统计信息
func (h *IntegrationHandler) GetIntegrationStatistics(c *gin.Context) {
	// 这里应该调用统计服务获取集成相关的统计信息
	// 由于统计服务还没有完全实现，先返回模拟数据

	statistics := gin.H{
		"total_integrations": 0,
		"conflict_checks": map[string]interface{}{
			"total":    0,
			"has_conflict": 0,
			"no_conflict": 0,
		},
		"case_creations": map[string]interface{}{
			"total":      0,
			"successful": 0,
			"failed":     0,
		},
		"approval_types": map[string]int{},
		"success_rate": 0.0,
		"average_processing_time": "0s",
		"last_updated": time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    statistics,
	})
}

// ProcessApprovalWithConflict 处理包含冲突检测的审批申请
func (h *IntegrationHandler) ProcessApprovalWithConflict(c *gin.Context) {
	approvalID := c.Param("id")
	if approvalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "审批ID不能为空",
		})
		return
	}

	var req models.ApprovalDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 从上下文获取用户信息
	userID := c.GetString("userID")
	userName := c.GetString("userName")

	// 调用集成服务处理审批
	updatedApproval, err := h.integrationService.ProcessApprovalWithConflict(c.Request.Context(), userID, userName, approvalID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "处理审批申请失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    updatedApproval,
	})
}

// GetIntegrationHistory 获取集成历史记录
func (h *IntegrationHandler) GetIntegrationHistory(c *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	approvalType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 这里应该调用集成服务获取历史记录
	// 由于服务还没有完全实现，先返回模拟数据

	history := gin.H{
		"records": []gin.H{},
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       0,
			"total_pages": 0,
		},
		"filters": gin.H{
			"status": status,
			"type":   approvalType,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
	})
}

// GetIntegrationLogs 获取集成日志
func (h *IntegrationHandler) GetIntegrationLogs(c *gin.Context) {
	approvalID := c.Query("approval_id")
	action := c.Query("action")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 这里应该调用日志服务获取集成日志
	// 由于日志服务还没有完全实现，先返回模拟数据

	logs := gin.H{
		"logs": []gin.H{},
		"filters": gin.H{
			"approval_id": approvalID,
			"action":      action,
			"start_date":  startDate,
			"end_date":    endDate,
		},
		"total": 0,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// RetryFailedIntegration 重试失败的集成
func (h *IntegrationHandler) RetryFailedIntegration(c *gin.Context) {
	approvalID := c.Param("id")
	if approvalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "审批ID不能为空",
		})
		return
	}

	var req struct {
		RetryType string `json:"retry_type" binding:"required"` // "conflict_check" or "case_creation"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 这里应该调用集成服务重试失败的集成
	// 由于服务还没有完全实现，先返回模拟响应

	result := gin.H{
		"approval_id": approvalID,
		"retry_type":  req.RetryType,
		"status":      "retried",
		"message":     "重试请求已提交",
		"retried_at":  time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// CancelIntegration 取消集成
func (h *IntegrationHandler) CancelIntegration(c *gin.Context) {
	approvalID := c.Param("id")
	if approvalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "审批ID不能为空",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "请求参数无效: " + err.Error(),
		})
		return
	}

	// 这里应该调用集成服务取消集成
	// 由于服务还没有完全实现，先返回模拟响应

	result := gin.H{
		"approval_id": approvalID,
		"status":      "cancelled",
		"reason":      req.Reason,
		"cancelled_at": time.Now(),
		"cancelled_by": c.GetString("userID"),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}