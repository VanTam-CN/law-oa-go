package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// IntegrationHandler 集成处理器
type IntegrationHandler struct {
	integrationService services.ApprovalConflictIntegrationService
	conflictService    services.ConflictDetectionServiceInterface
	db                 *gorm.DB
	authz              *services.AuthorizationService
}

// NewIntegrationHandler 创建集成处理器
func NewIntegrationHandler(
	integrationService services.ApprovalConflictIntegrationService,
	conflictService services.ConflictDetectionServiceInterface,
	db ...*gorm.DB,
) *IntegrationHandler {
	var database *gorm.DB
	if len(db) > 0 {
		database = db[0]
	}
	return &IntegrationHandler{
		integrationService: integrationService,
		conflictService:    conflictService,
		db:                 database,
	}
}

func (h *IntegrationHandler) SetAuthorizationService(authz *services.AuthorizationService) {
	h.authz = authz
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

	userID, userName := currentUserForIntegration(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未授权访问",
		})
		return
	}
	if !canCreateIntegratedMatter(c, &req) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "只有执业律师或独立冲突核查岗可以发起接案冲突流程",
		})
		return
	}
	if !h.validateCaseIntakeApproval(c, &req, userID) {
		return
	}

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
		userIDStr = ""
	}
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未授权访问",
		})
		return
	}
	if !canCreateIntegratedMatter(c, &req) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "只有执业律师或独立冲突核查岗可以发起接案冲突流程",
		})
		return
	}
	if !h.validateCaseIntakeApproval(c, &req, userIDStr) {
		return
	}

	userNameStr := "未知用户"
	if userName, ok := c.Get("username"); ok {
		if name, ok := userName.(string); ok && strings.TrimSpace(name) != "" {
			userNameStr = name
		}
	}

	// 调用集成服务
	result, err := h.integrationService.CreateIntegratedApproval(c.Request.Context(), userIDStr, userNameStr, &req)
	if err != nil {
		log.Printf("创建带冲突检测的审批申请失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "创建审批失败，请稍后重试；已保存的接案草稿不会丢失",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// validateCaseIntakeApproval binds the approval request to the server-side
// intake state. The browser's conflict result and case configuration are
// presentation data, not authority to submit a stale or unrelated check.
func (h *IntegrationHandler) validateCaseIntakeApproval(c *gin.Context, req *services.IntegrationRequest, actorID string) bool {
	if req == nil {
		return false
	}
	typeCaseCreation := strings.EqualFold(strings.TrimSpace(req.Type), "case_creation") ||
		strings.EqualFold(strings.TrimSpace(req.WorkflowType), "case_creation")
	if req.ConflictCheckConfig != nil && req.CaseCreationConfig == nil {
		// The legacy integration service can otherwise run a browser-supplied
		// conflict check and create an approval without an intake, facts
		// confirmation, or a verifiable subject case. Conflict-only work must
		// use the controlled task endpoint and its frozen evidence context.
		common.NewAPIError(c, http.StatusConflict, "INTAKE_REQUIRED_FOR_CONFLICT_APPROVAL", "冲突审批必须关联接案工作台和已完成的案件上下文，不能从通用集成接口直接发起")
		return false
	}
	if !typeCaseCreation && req.CaseCreationConfig == nil {
		return true
	}
	if h.db == nil {
		common.NewAPIError(c, http.StatusServiceUnavailable, "INTAKE_APPROVAL_GATE_UNAVAILABLE", "接案审批门禁未初始化，已阻止提交")
		return false
	}
	intakeID := integrationMetadataString(req.Metadata, "intake_id")
	if intakeID == "" {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_REQUIRED_FOR_CASE_APPROVAL", "案件创建审批必须关联接案工作台记录")
		return false
	}
	if req.ConflictCheckConfig != nil {
		// A case-intake conflict result must be created by the controlled
		// facts-confirmation -> conflict-check endpoint. Allowing the legacy
		// integration payload to trigger another check would create an
		// unbound result after the approval has already been submitted.
		common.NewAPIError(c, http.StatusConflict, "INTAKE_CONFLICT_CHECK_MUST_PRECEDE_APPROVAL", "接案审批必须使用已绑定的冲突检测结果，不支持审批接口自动补跑检测")
		return false
	}
	var intake map[string]interface{}
	if err := h.db.Table("case_intakes").Where("id = ?", intakeID).Take(&intake).Error; err != nil {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_NOT_FOUND", "关联接案记录不存在，已阻止提交")
		return false
	}
	role, _ := currentAuthActor(c)
	if strings.EqualFold(strings.TrimSpace(role.Role), "lawyer") && strings.TrimSpace(fmt.Sprint(intake["created_by"])) != strings.TrimSpace(actorID) {
		forbidObjectAccess(c)
		return false
	}
	if strings.TrimSpace(fmt.Sprint(intake["status"])) != "conflict_ready" {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_CONFLICT_RESULT_STALE", "接案事实已变化或尚未完成检测，请重新确认并运行利益冲突检查")
		return false
	}
	if req.CaseCreationConfig != nil {
		intakeTitle := firstIntegrationMetadataString(intake, "title")
		configuredTitle := firstIntegrationMetadataString(req.CaseCreationConfig, "title", "case_title", "caseTitle")
		if configuredTitle != "" && intakeTitle != "" && configuredTitle != intakeTitle {
			common.NewAPIError(c, http.StatusConflict, "INTAKE_CASE_TITLE_MISMATCH", "提交的案件名称与已确认接案事实不一致，请返回接案工作台重新确认")
			return false
		}
		configuredClientID := firstIntegrationMetadataString(req.CaseCreationConfig, "client_id", "clientId")
		intakeClientID := strings.TrimSpace(fmt.Sprint(intake["client_id"]))
		if configuredClientID != "" && configuredClientID != "<nil>" && intakeClientID != "" && configuredClientID != intakeClientID {
			common.NewAPIError(c, http.StatusConflict, "INTAKE_CLIENT_MISMATCH", "提交的客户与已确认接案事实不一致，请返回接案工作台重新确认")
			return false
		}
	}
	checkID := integrationMetadataString(req.Metadata, "conflict_check_id")
	if checkID == "" {
		if result, ok := req.Metadata["conflict_result"].(map[string]interface{}); ok {
			checkID = firstIntegrationMetadataString(result, "checkId", "check_id", "id")
		}
	}
	intakeMetadata := integrationJSONMap(intake["metadata"])
	if checkID == "" || integrationMetadataString(intakeMetadata, "conflict_check_id") != checkID {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_CONFLICT_LINK_MISMATCH", "提交的冲突检测记录与接案记录不一致，已阻止提交")
		return false
	}
	var record struct {
		CheckID     string
		CheckStatus string
		CheckResult string
	}
	if err := h.db.Table("conflict_check_records").Select("check_id, check_status, check_result").Where("check_id = ?", checkID).Take(&record).Error; err != nil || !strings.EqualFold(strings.TrimSpace(record.CheckStatus), "COMPLETED") {
		common.NewAPIError(c, http.StatusConflict, "CONFLICT_RESULT_NOT_COMPLETED", "关联冲突检测尚未完成，已阻止提交")
		return false
	}
	var canonicalResult interface{}
	if strings.TrimSpace(record.CheckResult) == "" || json.Unmarshal([]byte(record.CheckResult), &canonicalResult) != nil {
		common.NewAPIError(c, http.StatusConflict, "CONFLICT_RESULT_UNAVAILABLE", "关联冲突检测缺少可追溯结果，已阻止提交")
		return false
	}
	if req.Metadata == nil {
		req.Metadata = map[string]interface{}{}
	}
	req.Metadata["conflict_check_id"] = record.CheckID
	req.Metadata["conflict_result"] = canonicalResult
	return true
}

func integrationJSONMap(value interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed
	case []byte:
		_ = json.Unmarshal(typed, &result)
	case string:
		_ = json.Unmarshal([]byte(typed), &result)
	}
	return result
}

func integrationMetadataString(metadata map[string]interface{}, key string) string {
	if metadata == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(metadata[key]))
}

func firstIntegrationMetadataString(metadata map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := integrationMetadataString(metadata, key); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
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

	userID, _ := currentUserForIntegration(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未授权访问",
		})
		return
	}

	// 调用集成服务
	status, err := h.integrationService.GetIntegrationStatusForUser(c.Request.Context(), userID, approvalID)
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

	userID, _ := currentUserForIntegration(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未授权访问",
		})
		return
	}

	// 调用集成服务
	result, err := h.integrationService.AutoCreateCaseFromApprovalForUser(c.Request.Context(), userID, approvalID, req.CaseData)
	if err != nil {
		if isSubjectWorkflowError(err) {
			writeSubjectWorkflowError(c, err)
			return
		}
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
		userID, _ = currentUserForIntegration(c)
	}
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未授权访问",
		})
		return
	}
	if !canTriggerIntegratedConflictCheck(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "只有执业律师或独立冲突核查岗可以发起正式冲突检测",
		})
		return
	}
	if strings.TrimSpace(req.IntakeID) != "" {
		common.NewAPIError(c, http.StatusConflict, "INTAKE_CONFLICT_CHECK_REQUIRED", "接案冲突检查必须先由律师确认事实，再从接案工作台发起")
		return
	}
	// Keep this alternate transport on the exact same canonical case/client/
	// lawyer binding path as POST /conflict/check. Without this gate a caller
	// could submit browser-controlled subject labels and create a second,
	// unbound conflict conclusion.
	canonicalHandler := &ConflictHandlerSimple{authz: h.authz, db: h.db}
	if !canonicalHandler.prepareConflictRequest(c, &req) {
		return
	}
	// The integration endpoint must use the same full-history defaults as the
	// primary conflict-check endpoint; a client must not narrow the evidence
	// window by changing browser payloads.
	req.SearchYears = 0
	req.IncludeCorporateRelations = true
	if strings.EqualFold(strings.TrimSpace(req.SearchDepth), "BASIC") || strings.TrimSpace(req.SearchDepth) == "" {
		req.SearchDepth = "STANDARD"
	}
	req.RequestTime = time.Now()
	if h.conflictService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "conflict_service_unavailable",
			"message": "冲突检测服务未初始化",
		})
		return
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
	// This endpoint is an alternate transport for the same conflict check. It
	// must use the exact disclosure policy as /conflict/check; otherwise a
	// caller could bypass the ordinary lawyer result projection.
	projectConflictResponseForViewer(c, result)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetIntegrationStatistics 获取集成统计信息
func (h *IntegrationHandler) GetIntegrationStatistics(c *gin.Context) {
	common.NewAPIError(c, http.StatusServiceUnavailable, "INTEGRATION_STATISTICS_UNAVAILABLE", "集成统计尚未接入真实审计数据，当前不可用")
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

	userID, userName := currentUserForIntegration(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "未授权访问",
		})
		return
	}

	// 调用集成服务处理审批
	updatedApproval, err := h.integrationService.ProcessApprovalWithConflict(c.Request.Context(), userID, userName, approvalID, &req)
	if err != nil {
		var gateErr *services.ApprovalConflictGateError
		if errors.As(err, &gateErr) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   gateErr.Code,
				"message": gateErr.Message,
			})
			return
		}
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

func currentUserForIntegration(c *gin.Context) (string, string) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		return c.GetString("userID"), c.GetString("userName")
	}
	var userID string
	switch v := userIDValue.(type) {
	case uint:
		userID = strconv.FormatUint(uint64(v), 10)
	case int:
		userID = strconv.Itoa(v)
	case float64:
		userID = strconv.FormatInt(int64(v), 10)
	case string:
		userID = v
	default:
		userID = ""
	}
	userNameValue, ok := c.Get("username")
	if !ok || userNameValue == nil {
		return userID, "未知用户"
	}
	userName, ok := userNameValue.(string)
	if !ok || userName == "" {
		return userID, "未知用户"
	}
	return userID, userName
}

// canCreateIntegratedMatter protects the less visible integration API as well
// as the main intake workbench. Hiding a button for assistants is not an
// authorization boundary: a caller that submits conflict or case-creation
// configuration must still be a lawyer or an independent conflict reviewer.
func canCreateIntegratedMatter(c *gin.Context, req *services.IntegrationRequest) bool {
	if req == nil {
		return false
	}
	workflowType := strings.ToLower(strings.TrimSpace(req.WorkflowType))
	requestType := strings.ToLower(strings.TrimSpace(req.Type))
	requiresMatterRole := req.ConflictCheckConfig != nil || req.CaseCreationConfig != nil ||
		workflowType == "conflict_approval" || workflowType == "case_creation" ||
		requestType == "conflict_approval" || requestType == "case_creation"
	if !requiresMatterRole {
		return true
	}
	role := strings.ToLower(strings.TrimSpace(c.GetString("role")))
	return role == "lawyer" || services.IsConflictReviewRole(role)
}

func canTriggerIntegratedConflictCheck(c *gin.Context) bool {
	role := strings.ToLower(strings.TrimSpace(c.GetString("role")))
	return role == "lawyer" || services.IsConflictReviewRole(role)
}

// GetIntegrationHistory 获取集成历史记录
func (h *IntegrationHandler) GetIntegrationHistory(c *gin.Context) {
	common.NewAPIError(c, http.StatusServiceUnavailable, "INTEGRATION_HISTORY_UNAVAILABLE", "集成历史尚未接入真实审计数据，当前不可用")
}

// GetIntegrationLogs 获取集成日志
func (h *IntegrationHandler) GetIntegrationLogs(c *gin.Context) {
	common.NewAPIError(c, http.StatusServiceUnavailable, "INTEGRATION_LOGS_UNAVAILABLE", "集成日志尚未接入真实审计数据，当前不可用")
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
	if req.RetryType != "case_creation" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "unsupported_retry_type",
			"message": "当前只支持重试正式成案；冲突检索失败必须重新发起新的检测任务",
		})
		return
	}
	userID, _ := currentUserForIntegration(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "未授权访问"})
		return
	}
	result, err := h.integrationService.RetryCaseCreationForUser(c.Request.Context(), userID, approvalID)
	if err != nil {
		if isSubjectWorkflowError(err) {
			writeSubjectWorkflowError(c, err)
			return
		}
		c.JSON(http.StatusConflict, gin.H{
			"error":   "case_creation_retry_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// CancelIntegration 取消集成
func (h *IntegrationHandler) CancelIntegration(c *gin.Context) {
	common.NewAPIError(c, http.StatusServiceUnavailable, "INTEGRATION_CANCEL_UNAVAILABLE", "集成取消尚未接入真实状态机，当前不可用")
}
