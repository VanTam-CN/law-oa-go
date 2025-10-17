package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

// ConflictHandler 冲突检测处理器
type ConflictHandler struct {
	conflictService       services.ConflictService
	enhancedConflictSvc   *services.EnhancedConflictServiceV2
}

// NewConflictHandler 创建新的冲突检测处理器
func NewConflictHandler(conflictService services.ConflictService, enhancedConflictSvc *services.EnhancedConflictServiceV2) *ConflictHandler {
	return &ConflictHandler{
		conflictService:     conflictService,
		enhancedConflictSvc: enhancedConflictSvc,
	}
}

// validateConflictCheckRequest 验证冲突检查请求的业务逻辑
func validateConflictCheckRequest(req *CheckConflictRequest) error {
	// 验证客户ID格式
	if len(req.ClientID) == 0 {
		return fmt.Errorf("客户ID不能为空")
	}

	// 验证客户名称
	if len(strings.TrimSpace(req.ClientName)) == 0 {
		return fmt.Errorf("客户名称不能为空或只包含空格")
	}

	// 验证案件名称
	if len(strings.TrimSpace(req.CaseName)) == 0 {
		return fmt.Errorf("案件名称不能为空或只包含空格")
	}

	// 验证案件类型（支持大小写不敏感）
	validCaseTypes := []string{"civil", "commercial", "criminal", "administrative", "arbitration", "consultation", "other"}
	caseTypeLower := strings.ToLower(req.CaseType)
	caseTypeValid := false
	for _, validType := range validCaseTypes {
		if caseTypeLower == validType {
			caseTypeValid = true
			req.CaseType = validType // 标准化为小写
			break
		}
	}
	if !caseTypeValid {
		return fmt.Errorf("案件类型 '%s' 无效，支持的类型: %s", req.CaseType, strings.Join(validCaseTypes, ", "))
	}

	// 验证客户类型
	if req.ClientType != "PERSON" && req.ClientType != "COMPANY" {
		return fmt.Errorf("客户类型必须是 PERSON 或 COMPANY，当前值: %s", req.ClientType)
	}

	// 验证搜索深度
	if req.SearchDepth != "" {
		validDepths := []string{"BASIC", "STANDARD", "DEEP"}
		if !contains(validDepths, req.SearchDepth) {
			return fmt.Errorf("搜索深度 '%s' 无效，支持的深度: %s", req.SearchDepth, strings.Join(validDepths, ", "))
		}
	}

	// 验证搜索年限
	if req.SearchYears < 0 || req.SearchYears > 20 {
		return fmt.Errorf("搜索年限必须在 0-20 年之间，当前值: %d", req.SearchYears)
	}

	// 验证对方当事人信息
	for i, party := range req.OtherParties {
		if len(strings.TrimSpace(party)) == 0 {
			return fmt.Errorf("对方当事人信息第 %d 项不能为空", i+1)
		}
	}

	return nil
}

// contains 检查字符串切片是否包含指定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("REQ_%d_%d", time.Now().Unix(), time.Now().Nanosecond()%1000000)
}

// createErrorResponse 创建标准化的错误响应
func createErrorResponse(success bool, message, error string, details map[string]interface{}, requestID string) CheckConflictResponse {
	return CheckConflictResponse{
		Success:   success,
		Message:   message,
		Error:     error,
		Details:   details,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}

// createSuccessResponse 创建标准化的成功响应
func createSuccessResponse(message string, data *models.ConflictCheckResponse, requestID string) CheckConflictResponse {
	return CheckConflictResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}

// CheckConflictRequest 冲突检测请求
type CheckConflictRequest struct {
	ClientID                  string    `json:"clientId" binding:"required"`
	ClientName                string    `json:"clientName" binding:"required"`
	CaseName                  string    `json:"caseName" binding:"required"`
	CaseType                  string    `json:"caseType" binding:"required"`
	ClientType                string    `json:"clientType" binding:"required"`
	OtherParties              []string  `json:"otherParties"`
	SearchYears               int       `json:"searchYears"`
	IncludeCorporateRelations bool      `json:"includeCorporateRelations"`
	SearchDepth               string    `json:"searchDepth"`
	UserID                    string    `json:"userId"`
	RequestTime               time.Time `json:"requestTime"`
}

// CheckConflictResponse 冲突检测响应
type CheckConflictResponse struct {
	Success   bool                          `json:"success"`
	Message   string                        `json:"message"`
	Data      *models.ConflictCheckResponse `json:"data,omitempty"`
	Error     string                        `json:"error,omitempty"`
	Details   map[string]interface{}        `json:"details,omitempty"`
	RequestID string                        `json:"requestId,omitempty"`
	Timestamp time.Time                     `json:"timestamp"`
}

// RuleCreateRequest 创建规则请求
type RuleCreateRequest struct {
	Name        string      `json:"name" binding:"required"`
	Type        string      `json:"type" binding:"required"`
	Description string      `json:"description"`
	Priority    int         `json:"priority" binding:"required"`
	Active      bool        `json:"active"`
	Conditions  models.JSON `json:"conditions"`
}

// RuleUpdateRequest 更新规则请求
type RuleUpdateRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Priority    int         `json:"priority"`
	Active      bool        `json:"active"`
	Conditions  models.JSON `json:"conditions"`
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
	requestID := generateRequestID()

	var req CheckConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 详细的验证错误处理
		var validationErrors []string

		// 检查具体的验证错误
		if req.ClientID == "" {
			validationErrors = append(validationErrors, "clientId: 客户ID不能为空")
		}
		if req.ClientName == "" {
			validationErrors = append(validationErrors, "clientName: 客户名称不能为空")
		}
		if req.CaseName == "" {
			validationErrors = append(validationErrors, "caseName: 案件名称不能为空")
		}
		if req.CaseType == "" {
			validationErrors = append(validationErrors, "caseType: 案件类型不能为空")
		}
		if req.ClientType == "" {
			validationErrors = append(validationErrors, "clientType: 客户类型不能为空")
		} else if req.ClientType != "PERSON" && req.ClientType != "COMPANY" {
			validationErrors = append(validationErrors, "clientType: 客户类型必须是PERSON或COMPANY")
		}
		if req.SearchDepth != "" && req.SearchDepth != "BASIC" && req.SearchDepth != "STANDARD" && req.SearchDepth != "DEEP" {
			validationErrors = append(validationErrors, "searchDepth: 搜索深度必须是BASIC、STANDARD或DEEP")
		}

		errorMessage := "请求参数验证失败"
		if len(validationErrors) > 0 {
			errorMessage = fmt.Sprintf("请求参数验证失败: %s", strings.Join(validationErrors, "; "))
		}

		fmt.Printf("[WARN] 请求参数验证失败 - RequestID: %s, 错误数量: %d, 原始错误: %v\n",
			requestID, len(validationErrors), err)

		details := make(map[string]interface{})
		details["validation_errors"] = validationErrors
		details["raw_error"] = err.Error()

		response := createErrorResponse(false, errorMessage, fmt.Sprintf("请求格式错误: %v", err), details, requestID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 额外的业务逻辑验证
	if err := validateConflictCheckRequest(&req); err != nil {
		fmt.Printf("[WARN] 业务逻辑验证失败 - RequestID: %s, 错误: %v, 客户: %s\n",
			requestID, err, req.ClientName)

		details := make(map[string]interface{})
		details["business_validation_error"] = err.Error()
		details["request_data"] = req

		response := createErrorResponse(false, "请求数据验证失败", err.Error(), details, requestID)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// 记录详细的请求日志
	fmt.Printf("[INFO] 收到冲突检查请求 - RequestID: %s, 客户: %s, 案件: %s, 类型: %s, 用户ID: %s\n",
		requestID, req.ClientName, req.CaseName, req.CaseType, req.UserID)
	fmt.Printf("[DEBUG] 请求详情 - RequestID: %s, 搜索年限: %d, 搜索深度: %s, 包含企业关系: %t, 对方当事人数量: %d\n",
		requestID, req.SearchYears, req.SearchDepth, req.IncludeCorporateRelations, len(req.OtherParties))

	// 转换UserID字符串为uint
	userID := uint(0)
	if req.UserID != "" {
		if id, err := strconv.ParseUint(req.UserID, 10, 32); err == nil {
			userID = uint(id)
		}
	}

	// 转换为服务层请求格式
	request := &models.ConflictCheckRequest{
		ClientID:                  req.ClientID,
		ClientName:                req.ClientName,
		CaseName:                  req.CaseName,
		CaseType:                  req.CaseType,
		ClientType:                req.ClientType,
		OtherParties:              req.OtherParties,
		SearchYears:               req.SearchYears,
		IncludeCorporateRelations: req.IncludeCorporateRelations,
		SearchDepth:               req.SearchDepth,
		UserID:                    userID,
		RequestTime:               req.RequestTime,
	}

	// 检查服务是否可用
	if h.conflictService == nil {
		fmt.Printf("[ERROR] 冲突检测服务未初始化 - RequestID: %s, 客户: %s\n",
			requestID, req.ClientName)

		details := make(map[string]interface{})
		details["service_status"] = "not_initialized"
		details["available_services"] = []string{} // 可以列出可用的服务

		response := createErrorResponse(false, "冲突检测服务未初始化", "Service not initialized", details, requestID)
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	// 执行冲突检测
	fmt.Printf("[INFO] 开始执行冲突检测 - RequestID: %s, 客户: %s\n", requestID, req.ClientName)
	result, err := h.conflictService.CheckConflict(c.Request.Context(), request)
	if err != nil {
		fmt.Printf("[ERROR] 冲突检测失败 - RequestID: %s, 错误: %v, 客户: %s, 案件: %s\n",
			requestID, err, req.ClientName, req.CaseName)

		details := make(map[string]interface{})
		details["service_error"] = err.Error()
		details["request_summary"] = map[string]interface{}{
			"client_name": req.ClientName,
			"case_name":   req.CaseName,
			"case_type":   req.CaseType,
		}
		details["processing_time_ms"] = time.Since(startTime).Milliseconds()

		response := createErrorResponse(false, "冲突检测失败", err.Error(), details, requestID)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// 记录检测完成日志
	duration := time.Since(startTime)
	fmt.Printf("[INFO] 冲突检测完成 - RequestID: %s, 耗时: %vms, 客户: %s, 案件: %s\n",
		requestID, duration.Milliseconds(), req.ClientName, req.CaseName)
	fmt.Printf("[DEBUG] 检测结果摘要 - RequestID: %s, 发现冲突: %t, 检查ID: %s, 风险等级: %s\n",
		requestID, result.HasConflict, result.CheckID,
		func() string {
			if result.RiskAssessment != nil {
				return result.RiskAssessment.OverallRisk
			}
			return "未知"
		}())

	response := createSuccessResponse("冲突检测完成", result, requestID)
	c.JSON(http.StatusOK, response)
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
		"totalChecks":     0,
		"conflictChecks":  0,
		"highRiskChecks":  0,
		"averageDuration": 0.0,
		"lastCheckTime":   nil,
		"clientID":        clientID,
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
	startTime := time.Now()

	// 执行健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthData := gin.H{
		"status":    "healthy",
		"service":   "conflict-detection",
		"timestamp": startTime,
		"checks":    gin.H{},
	}

	// 检查服务是否初始化
	if h.conflictService == nil {
		healthData["status"] = "unhealthy"
		healthData["checks"].(gin.H)["service_initialized"] = false
	} else {
		healthData["checks"].(gin.H)["service_initialized"] = true

		// 尝试获取MCP标准作为健康检查
		_, err := h.conflictService.GetMCPStandards(ctx)
		if err != nil {
			healthData["status"] = "degraded"
			healthData["checks"].(gin.H)["mcp_standards"] = gin.H{
				"status": "failed",
				"error":  err.Error(),
			}
		} else {
			healthData["checks"].(gin.H)["mcp_standards"] = gin.H{
				"status": "ok",
			}
		}

		// 尝试获取冲突规则
		_, err = h.conflictService.GetConflictRules(ctx)
		if err != nil {
			healthData["status"] = "degraded"
			healthData["checks"].(gin.H)["conflict_rules"] = gin.H{
				"status": "failed",
				"error":  err.Error(),
			}
		} else {
			healthData["checks"].(gin.H)["conflict_rules"] = gin.H{
				"status": "ok",
			}
		}
	}

	// 设置响应状态码
	statusCode := http.StatusOK
	if healthData["status"] == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if healthData["status"] == "degraded" {
		statusCode = http.StatusOK // 服务可用但功能受限
	}

	healthData["duration"] = time.Since(startTime).Milliseconds()

	c.JSON(statusCode, gin.H{
		"success":   true,
		"message":   "健康检查完成",
		"data":      healthData,
		"timestamp": time.Now(),
	})
}

// EnhancedConflictCheck 执行增强的利益冲突检测
// @Summary 执行增强的利益冲突检测
// @Description 使用新算法对新案件进行利益冲突检测，支持行业竞争分析、风险评估等高级功能
// @Tags conflict
// @Accept json
// @Produce json
// @Param request body services.ConflictCheckRequest true "增强冲突检测请求"
// @Success 200 {object} Response{data=services.ConflictCheckResult}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/conflict/check [post]
func (h *ConflictHandler) EnhancedConflictCheck(c *gin.Context) {
	startTime := time.Now()
	requestID := fmt.Sprintf("ECC_%d", startTime.Unix())

	var req repositories.AdvancedConflictCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求参数格式错误",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	// 设置请求时间
	if req.RequestTime.IsZero() {
		req.RequestTime = time.Now()
	}

	// 验证请求参数
	validation := repositories.ValidateConflictRequest(&req)
	if !validation.IsValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求参数验证失败",
			"data":    validation.Errors,
			"error":   "validation failed",
		})
		return
	}

	// 执行增强冲突检测
	result, err := h.enhancedConflictSvc.CheckConflictsV2(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "冲突检测执行失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	// 记录检测日志
	h.logEnhancedConflictDetection(requestID, &req, result)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "增强冲突检测完成",
		"data":    result,
		"request_id": requestID,
		"duration": time.Since(startTime).Milliseconds(),
	})
}

// logEnhancedConflictDetection 记录增强冲突检测日志
func (h *ConflictHandler) logEnhancedConflictDetection(requestID string, req *repositories.AdvancedConflictCheckRequest, result *repositories.ConflictAnalysisResult) {
	log.Printf("=== 增强冲突检测日志 [RequestID: %s] ===", requestID)
	log.Printf("律师ID: %d", req.LawyerID)
	log.Printf("客户: %s", req.ClientName)
	log.Printf("对方: %s", req.OpposingParty)
	log.Printf("案件类型: %s", req.CaseType)
	log.Printf("搜索深度: %s", req.SearchDepth)
	log.Printf("检测结果: 冲突=%v, 风险等级=%s, 风险分数=%d, 冲突案件数=%d",
		result.HasConflicts, result.ConflictLevel, result.RiskScore, len(result.ConflictCases))

	if result.CompetitionAnalysis != nil && result.CompetitionAnalysis.HasCompetition {
		log.Printf("竞争分析: 检测到竞争关系, 竞争者数=%d", len(result.CompetitionAnalysis.CompetitorInfo))
	}

	if result.AnalysisSummary != nil {
		log.Printf("分析摘要: 总检查=%d, 直接冲突=%d, 行业冲突=%d, 名称相似=%d, 相关冲突=%d",
			result.AnalysisSummary.TotalCasesChecked,
			result.AnalysisSummary.DirectConflicts,
			result.AnalysisSummary.IndustryConflicts,
			result.AnalysisSummary.NameSimilarityCases,
			result.AnalysisSummary.RelatedConflicts)
	}

	log.Printf("=== 增强冲突检测日志结束 ===")
}

// InitializeConflictData 初始化冲突检测数据
// @Summary 初始化冲突检测数据
// @Description 初始化行业分类、竞争关系、冲突规则等基础数据
// @Tags conflict
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/conflict/initialize [post]
func (h *ConflictHandler) InitializeConflictData(c *gin.Context) {
	// 获取行业竞争服务
	industryService := services.NewIndustryCompetitionService(h.conflictService.GetRepositories())

	// 初始化行业数据
	err := industryService.InitializeIndustryData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "初始化冲突检测数据失败",
			"data":    nil,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "冲突检测数据初始化完成",
		"data":    gin.H{"initialized": true},
	})
}

// GetConflictStatistics 获取冲突检测统计信息
// @Summary 获取冲突检测统计信息
// @Description 获取指定律师或全局的冲突检测统计数据
// @Tags conflict
// @Accept json
// @Produce json
// @Param lawyer_id query int false "律师ID，不提供则返回全局统计"
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/conflict/statistics [get]
func (h *ConflictHandler) GetConflictStatistics(c *gin.Context) {
	lawyerIDStr := c.Query("lawyer_id")
	var lawyerID *int

	if lawyerIDStr != "" {
		id, err := strconv.Atoi(lawyerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "律师ID格式错误",
				"data":    nil,
			})
			return
		}
		lawyerID = &id
	}

	// 获取统计数据（这里简化处理，实际应该查询数据库）
	statistics := gin.H{
		"total_checks":        100,
		"conflicts_detected":  45,
		"high_risk_cases":     15,
		"medium_risk_cases":   20,
		"low_risk_cases":      10,
		"most_common_client":  "阿里巴巴集团",
		"most_common_case_type": "商事",
		"recent_activity": gin.H{
			"last_24h":    5,
			"last_7d":     25,
			"last_30d":    100,
		},
	}

	if lawyerID != nil {
		statistics["lawyer_id"] = *lawyerID
		// 律师个人统计（示例数据）
		statistics["total_checks"] = 20
		statistics["conflicts_detected"] = 8
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "统计信息获取成功",
		"data":    statistics,
	})
}
