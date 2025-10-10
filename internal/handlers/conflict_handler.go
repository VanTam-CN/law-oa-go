package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
	"github.com/gin-gonic/gin"
)

// ConflictHandler 冲突检测处理器
type ConflictHandler struct {
	conflictService services.ConflictService
}

// NewConflictHandler 创建新的冲突检测处理器
func NewConflictHandler(conflictService services.ConflictService) *ConflictHandler {
	return &ConflictHandler{
		conflictService: conflictService,
	}
}

// CheckConflictRequest 冲突检测请求
type CheckConflictRequest struct {
	ClientID                 string   `json:"clientId" binding:"required"`
	ClientName               string   `json:"clientName" binding:"required"`
	CaseName                 string   `json:"caseName" binding:"required"`
	CaseType                 string   `json:"caseType" binding:"required"`
	ClientType               string   `json:"clientType" binding:"required"`
	OtherParties             []string `json:"otherParties"`
	SearchYears              int      `json:"searchYears"`
	IncludeCorporateRelations bool     `json:"includeCorporateRelations"`
	SearchDepth              string   `json:"searchDepth"`
	UserID                   string   `json:"userId"`
	RequestTime              time.Time `json:"requestTime"`
}

// CheckConflictResponse 冲突检测响应
type CheckConflictResponse struct {
	Success    bool                         `json:"success"`
	Message    string                       `json:"message"`
	Data       *models.ConflictCheckResponse `json:"data,omitempty"`
	Error      string                       `json:"error,omitempty"`
	Timestamp  time.Time                    `json:"timestamp"`
}

// RuleCreateRequest 创建规则请求
type RuleCreateRequest struct {
	Name        string        `json:"name" binding:"required"`
	Type        string        `json:"type" binding:"required"`
	Description string        `json:"description"`
	Priority    int           `json:"priority" binding:"required"`
	Active      bool          `json:"active"`
	Conditions  models.JSON   `json:"conditions"`
}

// RuleUpdateRequest 更新规则请求
type RuleUpdateRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Priority    int           `json:"priority"`
	Active      bool          `json:"active"`
	Conditions  models.JSON   `json:"conditions"`
}

// CheckConflict 执行利益冲突检测
// @Summary 执行利益冲突检测
// @Description 对新案件进行利益冲突检测，返回检测结果和建议
// @Tags conflict
// @Accept json
// @Produce json
// @Param request body CheckConflictRequest true "冲突检测请求"
// @Success 200 {object} CheckConflictResponse
// @Failure 400 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/check [post]
func (h *ConflictHandler) CheckConflict(c *gin.Context) {
	startTime := time.Now()

	var req CheckConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "请求参数错误",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 转换为服务层请求格式
	request := &models.ConflictCheckRequest{
		ClientID:                 req.ClientID,
		ClientName:               req.ClientName,
		CaseName:                 req.CaseName,
		CaseType:                 req.CaseType,
		ClientType:               req.ClientType,
		OtherParties:             req.OtherParties,
		SearchYears:              req.SearchYears,
		IncludeCorporateRelations: req.IncludeCorporateRelations,
		SearchDepth:              req.SearchDepth,
		UserID:                   0, // TODO: 转换字符串到uint
		RequestTime:              req.RequestTime,
	}

	// 执行冲突检测
	result, err := h.conflictService.CheckConflict(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckConflictResponse{
			Success:   false,
			Message:   "冲突检测失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 记录检测完成日志
	duration := time.Since(startTime)
	fmt.Printf("冲突检测完成 - 耗时: %vms, 客户: %s, 案件: %s\n",
		duration.Milliseconds(), req.ClientName, req.CaseName)

	c.JSON(http.StatusOK, CheckConflictResponse{
		Success:   true,
		Message:   "冲突检测完成",
		Data:      result,
		Timestamp: time.Now(),
	})
}

// GetCheckHistory 获取检查历史
// @Summary 获取检查历史
// @Description 获取指定客户的冲突检查历史记录
// @Tags conflict
// @Accept json
// @Produce json
// @Param clientId path string true "客户ID"
// @Param limit query int false "返回记录数量限制" default(10)
// @Success 200 {object} CheckConflictResponse
// @Failure 400 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/history/{clientId} [get]
func (h *ConflictHandler) GetCheckHistory(c *gin.Context) {
	clientID := c.Param("clientId")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "客户ID不能为空",
			Timestamp: time.Now(),
		})
		return
	}

	// 获取limit参数
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // 限制最大返回数量
	}

	// 获取历史记录
	records, err := h.conflictService.GetCheckHistory(c.Request.Context(), clientID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckConflictResponse{
			Success:   false,
			Message:   "获取检查历史失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "获取检查历史成功",
		"data":      records,
		"timestamp": time.Now(),
	})
}

// GetCheckDetails 获取检查详情
// @Summary 获取检查详情
// @Description 获取指定冲突检查的详细信息
// @Tags conflict
// @Accept json
// @Produce json
// @Param checkId path string true "检查ID"
// @Success 200 {object} CheckConflictResponse
// @Failure 400 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/details/{checkId} [get]
func (h *ConflictHandler) GetCheckDetails(c *gin.Context) {
	checkID := c.Param("checkId")
	if checkID == "" {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "检查ID不能为空",
			Timestamp: time.Now(),
		})
		return
	}

	// 获取检查详情
	record, err := h.conflictService.GetCheckDetails(c.Request.Context(), checkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckConflictResponse{
			Success:   false,
			Message:   "获取检查详情失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "获取检查详情成功",
		"data":      record,
		"timestamp": time.Now(),
	})
}

// GetConflictRules 获取冲突规则
// @Summary 获取冲突规则
// @Description 获取所有的利益冲突检测规则
// @Tags conflict
// @Accept json
// @Produce json
// @Success 200 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/rules [get]
func (h *ConflictHandler) GetConflictRules(c *gin.Context) {
	// 获取规则列表
	rules, err := h.conflictService.GetConflictRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckConflictResponse{
			Success:   false,
			Message:   "获取冲突规则失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "获取冲突规则成功",
		"data":      rules,
		"timestamp": time.Now(),
	})
}

// CreateConflictRule 创建冲突规则
// @Summary 创建冲突规则
// @Description 创建新的利益冲突检测规则
// @Tags conflict
// @Accept json
// @Produce json
// @Param request body RuleCreateRequest true "创建规则请求"
// @Success 200 {object} CheckConflictResponse
// @Failure 400 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/rules [post]
func (h *ConflictHandler) CreateConflictRule(c *gin.Context) {
	var req RuleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "请求参数错误",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 创建规则对象
	rule := &models.ConflictRule{
		ID:          fmt.Sprintf("RULE_%d", time.Now().Unix()),
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Priority:    req.Priority,
		Active:      req.Active,
		Conditions:  req.Conditions,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 验证规则
	if err := rule.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "规则验证失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 注意：这里需要调用服务层的SaveConflictRule方法
	// 由于服务层接口暂不支持，返回占位符响应
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "创建规则功能待实现",
		"data":      rule,
		"timestamp": time.Now(),
	})
}

// UpdateConflictRule 更新冲突规则
// @Summary 更新冲突规则
// @Description 更新现有的利益冲突检测规则
// @Tags conflict
// @Accept json
// @Produce json
// @Param ruleId path string true "规则ID"
// @Param request body RuleUpdateRequest true "更新规则请求"
// @Success 200 {object} CheckConflictResponse
// @Failure 400 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/rules/{ruleId} [put]
func (h *ConflictHandler) UpdateConflictRule(c *gin.Context) {
	ruleID := c.Param("ruleId")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "规则ID不能为空",
			Timestamp: time.Now(),
		})
		return
	}

	var req RuleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "请求参数错误",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// 创建规则对象
	rule := &models.ConflictRule{
		ID:          ruleID,
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		Active:      req.Active,
		Conditions:  req.Conditions,
		UpdatedAt:   time.Now(),
	}

	// 更新规则
	err := h.conflictService.UpdateConflictRule(c.Request.Context(), rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckConflictResponse{
			Success:   false,
			Message:   "更新规则失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "更新规则成功",
		"data":      rule,
		"timestamp": time.Now(),
	})
}

// DeleteConflictRule 删除冲突规则
// @Summary 删除冲突规则
// @Description 删除指定的利益冲突检测规则
// @Tags conflict
// @Accept json
// @Produce json
// @Param ruleId path string true "规则ID"
// @Success 200 {object} CheckConflictResponse
// @Failure 400 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/rules/{ruleId} [delete]
func (h *ConflictHandler) DeleteConflictRule(c *gin.Context) {
	ruleID := c.Param("ruleId")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, CheckConflictResponse{
			Success:   false,
			Message:   "规则ID不能为空",
			Timestamp: time.Now(),
		})
		return
	}

	// 删除规则
	err := h.conflictService.DeleteConflictRule(c.Request.Context(), ruleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckConflictResponse{
			Success:   false,
			Message:   "删除规则失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "删除规则成功",
		"timestamp": time.Now(),
	})
}

// GetMCPStandards 获取MCP标准
// @Summary 获取MCP标准
// @Description 获取最新的MCP利益冲突检测标准
// @Tags conflict
// @Accept json
// @Produce json
// @Success 200 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/standards [get]
func (h *ConflictHandler) GetMCPStandards(c *gin.Context) {
	// 获取MCP标准
	standards, err := h.conflictService.GetMCPStandards(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, CheckConflictResponse{
			Success:   false,
			Message:   "获取MCP标准失败",
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "获取MCP标准成功",
		"data":      standards,
		"timestamp": time.Now(),
	})
}

// GetConflictStats 获取冲突统计
// @Summary 获取冲突统计
// @Description 获取冲突检测的统计信息
// @Tags conflict
// @Accept json
// @Produce json
// @Param clientId query string false "客户ID，不提供则返回全局统计"
// @Success 200 {object} CheckConflictResponse
// @Failure 500 {object} CheckConflictResponse
// @Router /api/v1/conflict/stats [get]
func (h *ConflictHandler) GetConflictStats(c *gin.Context) {
	clientID := c.Query("clientId")

	// TODO: 实现统计功能，需要扩展服务层接口
	// 目前返回模拟数据
	stats := gin.H{
		"totalChecks":        0,
		"conflictChecks":     0,
		"highRiskChecks":     0,
		"averageDuration":    0.0,
		"lastCheckTime":      nil,
		"clientID":          clientID,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "获取统计信息成功",
		"data":      stats,
		"timestamp": time.Now(),
	})
}

// HealthCheck 健康检查
// @Summary 冲突检测服务健康检查
// @Description 检查冲突检测服务的运行状态
// @Tags conflict
// @Accept json
// @Produce json
// @Success 200 {object} CheckConflictResponse
// @Router /api/v1/conflict/health [get]
func (h *ConflictHandler) HealthCheck(c *gin.Context) {
	// 执行简单的服务检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试获取MCP标准作为健康检查
	_, err := h.conflictService.GetMCPStandards(ctx)

	status := "healthy"
	if err != nil {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "健康检查完成",
		"data": gin.H{
			"status":    status,
			"service":   "conflict-detection",
			"timestamp": time.Now(),
		},
		"timestamp": time.Now(),
	})
}