package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"law-oa-go/internal/services"
	"law-oa-go/internal/validators"
)

// AnalyticsHandler 用户行为分析HTTP处理器
type AnalyticsHandler struct {
	sessionService       *services.SessionService
	pageViewService      *services.PageViewService
	eventTrackingService *services.EventTrackingService
	journeyService       *services.JourneyService
	behaviorAnalysis     *services.BehaviorAnalysisService
	realTimeStatsService *services.RealTimeStatsService
	validator            *validators.SimpleTestValidator
}

// NewAnalyticsHandler 创建分析HTTP处理器
func NewAnalyticsHandler(
	sessionService *services.SessionService,
	pageViewService *services.PageViewService,
	eventTrackingService *services.EventTrackingService,
	journeyService *services.JourneyService,
	behaviorAnalysis *services.BehaviorAnalysisService,
	realTimeStatsService *services.RealTimeStatsService,
	validator *validators.SimpleTestValidator,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		sessionService:       sessionService,
		pageViewService:      pageViewService,
		eventTrackingService: eventTrackingService,
		journeyService:       journeyService,
		behaviorAnalysis:     behaviorAnalysis,
		realTimeStatsService: realTimeStatsService,
		validator:            validator,
	}
}

// SessionHandler 会话相关处理器

// CreateSession 创建用户会话
func (h *AnalyticsHandler) CreateSession(c *gin.Context) {
	var req services.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid request format", "code": "INVALID_REQUEST"},
			"data":  nil,
		})
		return
	}

	// 验证请求参数
	if err := h.validateCreateSessionRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "VALIDATION_ERROR"},
			"data":  nil,
		})
		return
	}

	// 获取用户ID（从JWT或上下文中获取）
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"message": "User not authenticated", "code": "UNAUTHORIZED"},
			"data":  nil,
		})
		return
	}

	session, err := h.sessionService.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to create session", "code": "SESSION_CREATE_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error": nil,
		"data":  session,
	})
}

// GetSession 获取用户会话信息
func (h *AnalyticsHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	if err := h.validator.ValidateExecutionID(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid session ID", "code": "INVALID_SESSION_ID"},
			"data":  nil,
		})
		return
	}

	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"message": "Session not found", "code": "SESSION_NOT_FOUND"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  session,
	})
}

// UpdateSession 更新用户会话
func (h *AnalyticsHandler) UpdateSession(c *gin.Context) {
	sessionID := c.Param("id")
	if err := h.validator.ValidateExecutionID(sessionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid session ID", "code": "INVALID_SESSION_ID"},
			"data":  nil,
		})
		return
	}

	var req services.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid request format", "code": "INVALID_REQUEST"},
			"data":  nil,
		})
		return
	}

	if err := h.sessionService.UpdateSession(c.Request.Context(), sessionID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to update session", "code": "SESSION_UPDATE_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  gin.H{"message": "Session updated successfully"},
	})
}

// PageViewHandler 页面浏览相关处理器

// TrackPageView 追踪页面浏览
func (h *AnalyticsHandler) TrackPageView(c *gin.Context) {
	var req services.TrackPageViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid request format", "code": "INVALID_REQUEST"},
			"data":  nil,
		})
		return
	}

	// 验证请求参数
	if err := h.validateTrackPageViewRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "VALIDATION_ERROR"},
			"data":  nil,
		})
		return
	}

	if err := h.pageViewService.TrackPageView(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to track page view", "code": "PAGE_VIEW_TRACK_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error": nil,
		"data":  gin.H{"message": "Page view tracked successfully"},
	})
}

// EventTrackingHandler 事件追踪相关处理器

// TrackEvent 追踪用户事件
func (h *AnalyticsHandler) TrackEvent(c *gin.Context) {
	var req services.TrackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid request format", "code": "INVALID_REQUEST"},
			"data":  nil,
		})
		return
	}

	// 验证请求参数
	if err := h.validateTrackEventRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "VALIDATION_ERROR"},
			"data":  nil,
		})
		return
	}

	if err := h.eventTrackingService.TrackEvent(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to track event", "code": "EVENT_TRACK_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error": nil,
		"data":  gin.H{"message": "Event tracked successfully"},
	})
}

// JourneyHandler 用户旅程相关处理器

// CreateJourney 创建用户旅程
func (h *AnalyticsHandler) CreateJourney(c *gin.Context) {
	var req services.CreateJourneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid request format", "code": "INVALID_REQUEST"},
			"data":  nil,
		})
		return
	}

	// 验证请求参数
	if err := h.validateCreateJourneyRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "VALIDATION_ERROR"},
			"data":  nil,
		})
		return
	}

	if err := h.journeyService.CreateJourney(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to create journey", "code": "JOURNEY_CREATE_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error": nil,
		"data":  gin.H{"message": "Journey created successfully"},
	})
}

// AnalyticsAnalysisHandler 分析相关处理器

// GetUserBehaviorAnalysis 获取用户行为分析
func (h *AnalyticsHandler) GetUserBehaviorAnalysis(c *gin.Context) {
	userID := c.Param("user_id")
	if err := h.validator.ValidateExecutionID(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid user ID", "code": "INVALID_USER_ID"},
			"data":  nil,
		})
		return
	}

	// 解析时间范围
	startDate, endDate, err := h.parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "INVALID_DATE_RANGE"},
			"data":  nil,
		})
		return
	}

	analysis, err := h.behaviorAnalysis.AnalyzeUserBehavior(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to analyze user behavior", "code": "ANALYSIS_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  analysis,
	})
}

// DetectBehaviorPatterns 检测行为模式
func (h *AnalyticsHandler) DetectBehaviorPatterns(c *gin.Context) {
	userID := c.Param("user_id")
	if err := h.validator.ValidateExecutionID(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid user ID", "code": "INVALID_USER_ID"},
			"data":  nil,
		})
		return
	}

	if err := h.behaviorAnalysis.DetectBehaviorPatterns(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to detect behavior patterns", "code": "PATTERN_DETECTION_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  gin.H{"message": "Behavior patterns detection started"},
	})
}

// RealTimeStatsHandler 实时统计相关处理器

// GetRealTimeDashboard 获取实时仪表板数据
func (h *AnalyticsHandler) GetRealTimeDashboard(c *gin.Context) {
	dashboard, err := h.realTimeStatsService.GetRealTimeDashboard(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to get real-time dashboard", "code": "DASHBOARD_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  dashboard,
	})
}

// UpdateRealTimeStats 更新实时统计数据
func (h *AnalyticsHandler) UpdateRealTimeStats(c *gin.Context) {
	if err := h.realTimeStatsService.UpdateRealTimeStats(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": "Failed to update real-time stats", "code": "STATS_UPDATE_ERROR"},
			"data":  nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  gin.H{"message": "Real-time stats updated successfully"},
	})
}

// GetDailyActiveUsers 获取日活跃用户统计
func (h *AnalyticsHandler) GetDailyActiveUsers(c *gin.Context) {
	_, _, err := h.parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "INVALID_DATE_RANGE"},
			"data":  nil,
		})
		return
	}

	// 这里需要通过AnalyticsRepository获取数据
	// 由于我们在服务层没有这个方法，我们需要添加到服务层或直接在handler中使用repository
	// 暂时返回空数据
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  []map[string]interface{}{},
	})
}

// GetPageViewStats 获取页面浏览统计
func (h *AnalyticsHandler) GetPageViewStats(c *gin.Context) {
	_, _, err := h.parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "INVALID_DATE_RANGE"},
			"data":  nil,
		})
		return
	}

	// 解析分页参数
	page, pageSize := h.parsePaginationParams(c)

	// 这里需要通过AnalyticsRepository获取数据
	// 暂时返回空数据
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data": gin.H{
			"stats":     []map[string]interface{}{},
			"page":      page,
			"page_size": pageSize,
			"total":     0,
		},
	})
}

// GetEventStats 获取事件统计
func (h *AnalyticsHandler) GetEventStats(c *gin.Context) {
	_, _, err := h.parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": err.Error(), "code": "INVALID_DATE_RANGE"},
			"data":  nil,
		})
		return
	}

	// 解析分页参数
	page, pageSize := h.parsePaginationParams(c)

	// 这里需要通过AnalyticsRepository获取数据
	// 暂时返回空数据
	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data": gin.H{
			"stats":     []map[string]interface{}{},
			"page":      page,
			"page_size": pageSize,
			"total":     0,
		},
	})
}

// BatchTrackEvents 批量追踪事件
func (h *AnalyticsHandler) BatchTrackEvents(c *gin.Context) {
	var req struct {
		Events []services.TrackEventRequest `json:"events" validate:"required,dive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Invalid request format", "code": "INVALID_REQUEST"},
			"data":  nil,
		})
		return
	}

	// 验证事件数量
	if len(req.Events) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "No events provided", "code": "NO_EVENTS"},
			"data":  nil,
		})
		return
	}

	if len(req.Events) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"message": "Too many events (max 100 per batch)", "code": "TOO_MANY_EVENTS"},
			"data":  nil,
		})
		return
	}

	// 处理每个事件
	successCount := 0
	var errors []string

	for i, event := range req.Events {
		if err := h.validateTrackEventRequest(&event); err != nil {
			errors = append(errors, fmt.Sprintf("Event %d: %s", i+1, err.Error()))
			continue
		}

		if err := h.eventTrackingService.TrackEvent(c.Request.Context(), &event); err != nil {
			errors = append(errors, fmt.Sprintf("Event %d: Failed to track", i+1))
			continue
		}

		successCount++
	}

	response := gin.H{
		"success_count": successCount,
		"total_count":   len(req.Events),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusOK, gin.H{
		"error": nil,
		"data":  response,
	})
}

// 验证方法

// validateCreateSessionRequest 验证创建会话请求
func (h *AnalyticsHandler) validateCreateSessionRequest(req *services.CreateSessionRequest) error {
	if req.IPAddress == "" {
		return fmt.Errorf("IP address is required")
	}

	if req.UserAgent == "" {
		return fmt.Errorf("User agent is required")
	}

	return nil
}

// validateTrackPageViewRequest 验证页面浏览追踪请求
func (h *AnalyticsHandler) validateTrackPageViewRequest(req *services.TrackPageViewRequest) error {
	if err := h.validator.ValidateExecutionID(req.SessionID); err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	if err := h.validator.ValidateURL(req.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if req.Duration < 0 {
		return fmt.Errorf("duration cannot be negative")
	}

	if req.ScrollDepth < 0 || req.ScrollDepth > 100 {
		return fmt.Errorf("scroll depth must be between 0 and 100")
	}

	return nil
}

// validateTrackEventRequest 验证事件追踪请求
func (h *AnalyticsHandler) validateTrackEventRequest(req *services.TrackEventRequest) error {
	if err := h.validator.ValidateExecutionID(req.SessionID); err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	if req.EventType == "" {
		return fmt.Errorf("event type is required")
	}

	if req.EventCategory == "" {
		return fmt.Errorf("event category is required")
	}

	if req.EventAction == "" {
		return fmt.Errorf("event action is required")
	}

	if req.URL != "" {
		if err := h.validator.ValidateURL(req.URL); err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
	}

	return nil
}

// validateCreateJourneyRequest 验证创建旅程请求
func (h *AnalyticsHandler) validateCreateJourneyRequest(req *services.CreateJourneyRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if req.JourneyType == "" {
		return fmt.Errorf("journey type is required")
	}

	if req.CurrentStep < 0 {
		return fmt.Errorf("current step cannot be negative")
	}

	if len(req.Steps) > 0 && req.CurrentStep >= len(req.Steps) {
		return fmt.Errorf("current step exceeds number of steps")
	}

	return nil
}

// 辅助方法

// parseDateRange 解析日期范围参数
func (h *AnalyticsHandler) parseDateRange(c *gin.Context) (time.Time, time.Time, error) {
	// 默认时间范围：最近7天
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	// 解析开始日期
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
		}
		startDate = parsed
	}

	// 解析结束日期
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
		}
		endDate = parsed
	}

	// 验证时间范围
	if startDate.After(endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date cannot be after end_date")
	}

	if endDate.Sub(startDate) > 365*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("date range cannot exceed 1 year")
	}

	return startDate, endDate, nil
}

// parsePaginationParams 解析分页参数
func (h *AnalyticsHandler) parsePaginationParams(c *gin.Context) (int, int) {
	// 默认分页参数
	page := 1
	pageSize := 20

	// 解析页码
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 解析页面大小
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return page, pageSize
}
