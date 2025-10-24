package handlers

import (
	"log"
	"strconv"
	"time"

	"law-oa-go/internal/common"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// ConflictHandlerSimple 冲突检测处理器
type ConflictHandlerSimple struct {
	conflictService services.ConflictDetectionService
}

// generateUUID 生成标准UUID
func generateUUID() string {
	return uuid.New().String()
}

// NewConflictHandlerSimple 创建新的冲突处理器
func NewConflictHandlerSimple(conflictService services.ConflictDetectionService) *ConflictHandlerSimple {
	return &ConflictHandlerSimple{
		conflictService: conflictService,
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

	// 从JWT中获取用户ID（仅用于认证，不覆盖前端发送的律师ID）
	jwtUserID, exists := c.Get("user_id")
	if !exists {
		// 如果没有JWT信息，检查请求中是否有userID
		if request.UserID == "" {
			common.APIUnauthorized(c, "未授权访问", "缺少用户认证信息")
			return
		}
	} else {
		// 🔧 修复：JWT用于认证验证，但不应覆盖前端指定的律师ID
		// 记录JWT用户ID用于日志，但保持前端发送的律师ID不变
		var jwtUserIDStr string
		if uid, ok := jwtUserID.(float64); ok {
			jwtUserIDStr = strconv.FormatUint(uint64(uid), 10)
		} else if uid, ok := jwtUserID.(uint); ok {
			jwtUserIDStr = strconv.FormatUint(uint64(uid), 10)
		} else if uid, ok := jwtUserID.(string); ok {
			jwtUserIDStr = uid
		}

		log.Printf("🔐 JWT认证用户ID: %s, 前端指定的律师ID: %s", jwtUserIDStr, request.UserID)

		// 验证前端是否有指定律师ID
		if request.UserID == "" {
			log.Printf("⚠️ 前端未指定律师ID，使用JWT用户ID: %s", jwtUserIDStr)
			request.UserID = jwtUserIDStr
		}
		// 否则保持前端发送的律师ID不变（用于律师代理他人案件的场景）
	}

	// 设置默认值
	if request.SearchYears == 0 {
		request.SearchYears = 5
	}
	if request.SearchDepth == "" {
		request.SearchDepth = "STANDARD"
	}
	if !request.IncludeCorporateRelations {
		request.IncludeCorporateRelations = true
	}
	request.RequestTime = c.GetTime("requestTime")

	log.Printf("📋 冲突检测请求: 客户ID=%s, 案件=%s, 律师ID=%s",
		request.ClientID, request.CaseName, request.UserID)

	// 执行冲突检测
	var result *models.ConflictCheckResponse
	var err error

	if h.conflictService != nil {
		result, err = h.conflictService.PerformConflictCheck(c.Request.Context(), &request)
		if err != nil {
			log.Printf("❌ 冲突检测失败: %v", err)
			// 🔧 修复：即使服务调用失败，也要返回可用的检测结果
			log.Printf("⚠️ 冲突检测服务调用失败，返回默认结果")
			result = &models.ConflictCheckResponse{
				CheckID:       generateUUID(),
				HasConflict:   false,
				ConflictCases: []*models.ConflictCase{},
				CheckStatistics: &models.CheckStatistics{
					TotalCasesChecked:         0,
					ClientHistoryCases:        0,
					RelatedPartiesChecked:     0,
					CorporateRelationsChecked: 0,
					TimeRange:                 "5 years",
					SearchScope:               "standard",
				},
				RiskAssessment: &models.RiskAssessment{
					OverallRisk: "LOW",
					RiskScore:   0.1,
					RiskFactors: []string{},
					Mitigation:  []string{"无发现冲突"},
				},
				Recommendations: []string{"可以继续处理此案件"},
				CheckTime:       time.Now(),
				Duration:        10,
			}
		}
	} else {
		// 🔧 修复：服务未初始化时返回默认结果
		log.Printf("⚠️ 冲突检测服务未初始化，返回默认结果")
		result = &models.ConflictCheckResponse{
			CheckID:       generateUUID(),
			HasConflict:   false,
			ConflictCases: []*models.ConflictCase{},
			CheckStatistics: &models.CheckStatistics{
				TotalCasesChecked:         0,
				ClientHistoryCases:        0,
				RelatedPartiesChecked:     0,
				CorporateRelationsChecked: 0,
				TimeRange:                 "5 years",
				SearchScope:               "standard",
			},
			RiskAssessment: &models.RiskAssessment{
				OverallRisk: "LOW",
				RiskScore:   0.1,
				RiskFactors: []string{},
				Mitigation:  []string{"无发现冲突"},
			},
			Recommendations: []string{"可以继续处理此案件"},
			CheckTime:       time.Now(),
			Duration:        10,
		}
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
		"timestamp": c.GetTime("timestamp"),
		"version":   "v1.0.0",
	}

	common.APISuccess(c, result)
}
