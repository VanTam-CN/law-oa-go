package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/gin-gonic/gin"
)

// ConflictHandlerSimple 冲突检测处理器
type ConflictHandlerSimple struct {
	conflictService      services.ConflictDetectionService
	asyncConflictService services.AsyncConflictCheckService
}

// NewConflictHandlerSimple 创建新的冲突处理器
func NewConflictHandlerSimple(conflictService services.ConflictDetectionService, asyncServices ...services.AsyncConflictCheckService) *ConflictHandlerSimple {
	var asyncService services.AsyncConflictCheckService
	if len(asyncServices) > 0 {
		asyncService = asyncServices[0]
	}
	return &ConflictHandlerSimple{
		conflictService:      conflictService,
		asyncConflictService: asyncService,
	}
}

// CheckConflict 执行冲突检查
func (h *ConflictHandlerSimple) CheckConflict(c *gin.Context) {
	log.Println("🔍 处理冲突检查请求")

	// 解析请求
	var request models.ConflictCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ 解析请求失败: %v", err)
		common.APIBadRequest(c, "请求参数错误", "请求体解析失败")
		return
	}

	if !h.prepareConflictRequest(c, &request) {
		return
	}

	// 设置默认值
	if request.SearchYears == 0 {
		request.SearchYears = 5
	}
	if request.SearchDepth == "" {
		request.SearchDepth = "STANDARD"
	}
	request.RequestTime = time.Now()

	log.Printf("📋 冲突检测请求: 客户ID=%s, 案件=%s, 律师ID=%s",
		request.ClientID, request.CaseName, request.UserID)

	// 执行冲突检测
	var result *models.ConflictCheckResponse
	var err error

	if h.conflictService != nil {
		result, err = h.conflictService.PerformConflictCheck(c.Request.Context(), &request)
		if err != nil {
			log.Printf("❌ 冲突检测失败: %v", err)
			var conflictErr *models.ConflictError
			if errors.As(err, &conflictErr) {
				common.APIBadRequest(c, conflictErr.Message, conflictErr.Code)
				return
			}
			common.APIInternalServerError(c, "冲突检测失败", err.Error())
			return
		}
	} else {
		log.Printf("❌ 冲突检测服务未初始化")
		common.NewAPIError(c, http.StatusServiceUnavailable, "CONFLICT_SERVICE_UNAVAILABLE", "冲突检测服务未初始化")
		return
	}

	// 构建统一响应格式
	data := gin.H{
		"checkId":         result.CheckID,
		"hasConflict":     result.HasConflict,
		"conflictCases":   result.ConflictCases,
		"checkStatistics": result.CheckStatistics,
		"riskAssessment":  result.RiskAssessment,
		"recommendations": result.Recommendations,
		"checkTime":       result.CheckTime,
		"duration":        result.Duration,
	}

	log.Printf("✅ 冲突检查完成，检测到 %d 个冲突案例，风险等级: %s",
		len(result.ConflictCases), result.RiskAssessment.OverallRisk)

	common.APISuccess(c, data)
}

// CreateConflictTask 创建异步冲突检测任务
func (h *ConflictHandlerSimple) CreateConflictTask(c *gin.Context) {
	var request models.ConflictCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请求体解析失败")
		return
	}

	if !h.prepareConflictRequest(c, &request) {
		return
	}
	if h.asyncConflictService == nil {
		common.APIInternalServerError(c, "异步冲突检测服务未初始化")
		return
	}

	task, err := h.asyncConflictService.CreateTask(c.Request.Context(), &request)
	if err != nil {
		common.APIBadRequest(c, "创建冲突检测任务失败", err.Error())
		return
	}

	common.APISuccess(c, task)
}

// GetConflictTaskStatus 获取异步冲突检测任务状态
func (h *ConflictHandlerSimple) GetConflictTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		common.APIBadRequest(c, "任务ID不能为空")
		return
	}
	if h.asyncConflictService == nil {
		common.APIInternalServerError(c, "异步冲突检测服务未初始化")
		return
	}

	task, err := h.asyncConflictService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, services.ErrConflictTaskNotFound) {
			common.APINotFound(c, "任务不存在", "指定的冲突检测任务不存在")
			return
		}
		common.APIInternalServerError(c, "获取任务状态失败", err.Error())
		return
	}

	common.APISuccess(c, task)
}

// GetConflictTaskResult 获取异步冲突检测任务结果
func (h *ConflictHandlerSimple) GetConflictTaskResult(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		common.APIBadRequest(c, "任务ID不能为空")
		return
	}
	if h.asyncConflictService == nil {
		common.APIInternalServerError(c, "异步冲突检测服务未初始化")
		return
	}

	result, err := h.asyncConflictService.GetTaskResult(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, services.ErrConflictTaskNotFound) {
			common.APINotFound(c, "任务不存在", "指定的冲突检测任务不存在")
			return
		}
		common.APIInternalServerError(c, "获取任务结果失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

func (h *ConflictHandlerSimple) prepareConflictRequest(c *gin.Context, request *models.ConflictCheckRequest) bool {
	jwtUserID, exists := c.Get("user_id")
	if !exists {
		if request.UserID == "" {
			common.APIUnauthorized(c, "未授权访问", "缺少用户认证信息")
			return false
		}
	} else {
		var jwtUserIDStr string
		if uid, ok := jwtUserID.(float64); ok {
			jwtUserIDStr = strconv.FormatUint(uint64(uid), 10)
		} else if uid, ok := jwtUserID.(uint); ok {
			jwtUserIDStr = strconv.FormatUint(uint64(uid), 10)
		} else if uid, ok := jwtUserID.(string); ok {
			jwtUserIDStr = uid
		}

		if request.UserID == "" {
			request.UserID = jwtUserIDStr
		} else if request.UserID != jwtUserIDStr && !canRunConflictCheckForOthers(c) {
			log.Printf("🚫 冲突检测被拒绝: 登录用户=%s, 请求检查律师=%s, 角色=%v", jwtUserIDStr, request.UserID, c.GetString("role"))
			common.NewAPIError(c, http.StatusForbidden, "CONFLICT_LAWYER_SCOPE_FORBIDDEN", "普通律师只能以本人作为承办律师执行冲突检查")
			return false
		}
	}

	if request.SearchYears == 0 {
		request.SearchYears = 5
	}
	if request.SearchDepth == "" {
		request.SearchDepth = "STANDARD"
	}
	request.RequestTime = time.Now()
	return true
}

func canRunConflictCheckForOthers(c *gin.Context) bool {
	role := strings.ToLower(strings.TrimSpace(c.GetString("role")))
	switch role {
	case "admin", "super_admin", "director", "partner", "compliance", "risk", "risk_control", "management":
		return true
	default:
		return false
	}
}

// GetCheckHistory 获取冲突检测历史
func (h *ConflictHandlerSimple) GetCheckHistory(c *gin.Context) {
	clientID := c.Query("clientId")
	if clientID == "" {
		common.APIBadRequest(c, "客户ID不能为空", "查询参数clientId是必需的")
		return
	}

	// 获取历史记录
	history, err := h.conflictService.GetCheckHistory(c.Request.Context(), clientID, 10)
	if err != nil {
		log.Printf("❌ 获取冲突检测历史失败: %v", err)
		common.APIInternalServerError(c, "获取历史记录失败", err.Error())
		return
	}

	common.APISuccess(c, history)
}

// GetConflictStats 获取冲突检测统计
func (h *ConflictHandlerSimple) GetConflictStats(c *gin.Context) {
	clientID := c.Query("clientId")

	stats, err := h.conflictService.GetConflictStats(c.Request.Context(), clientID)
	if err != nil {
		log.Printf("❌ 获取冲突检测统计失败: %v", err)
		common.APIInternalServerError(c, "获取统计信息失败", err.Error())
		return
	}

	common.APISuccess(c, stats)
}

// HealthCheck 冲突检查服务健康检查
func (h *ConflictHandlerSimple) HealthCheck(c *gin.Context) {
	result := gin.H{
		"status":    "healthy",
		"service":   "conflict-check",
		"timestamp": time.Now(),
		"version":   "v1.0.0",
	}

	common.APISuccess(c, result)
}
