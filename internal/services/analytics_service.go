package services

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// parseViewportSize 解析视口大小字符串 "1024x768" -> ViewportSize
func parseViewportSize(sizeStr string) *models.ViewportSize {
	if sizeStr == "" {
		return nil
	}

	parts := strings.Split(sizeStr, "x")
	if len(parts) != 2 {
		return nil
	}

	width, err1 := strconv.Atoi(parts[0])
	height, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return nil
	}

	return &models.ViewportSize{
		Width:  width,
		Height: height,
	}
}

// parseScreenSize 解析屏幕大小字符串 "1920x1080" -> ScreenSize
func parseScreenSize(sizeStr string) *models.ScreenSize {
	if sizeStr == "" {
		return nil
	}

	parts := strings.Split(sizeStr, "x")
	if len(parts) != 2 {
		return nil
	}

	width, err1 := strconv.Atoi(parts[0])
	height, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return nil
	}

	return &models.ScreenSize{
		Width:  width,
		Height: height,
	}
}

// parseInteraction 解析交互字符串为整数
func parseInteraction(interactionStr string) int {
	if interactionStr == "" {
		return 0
	}

	if count, err := strconv.Atoi(interactionStr); err == nil {
		return count
	}

	return 0
}

// AnalyticsService 用户行为分析服务
type AnalyticsService struct {
	analyticsRepo repositories.AnalyticsRepositoryInterface
	userRepo      repositories.UserRepository
}

// NewAnalyticsService 创建分析服务实例
func NewAnalyticsService(analyticsRepo repositories.AnalyticsRepositoryInterface, userRepo repositories.UserRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		userRepo:      userRepo,
	}
}

// SessionService 会话管理服务
type SessionService struct {
	analyticsRepo repositories.AnalyticsRepositoryInterface
}

// NewSessionService 创建会话服务实例
func NewSessionService(analyticsRepo repositories.AnalyticsRepositoryInterface) *SessionService {
	return &SessionService{
		analyticsRepo: analyticsRepo,
	}
}

// CreateSession 创建新的用户会话
func (s *SessionService) CreateSession(ctx context.Context, userID string, req *CreateSessionRequest) (*models.AnalyticsUserSession, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	// 解析用户代理信息
	deviceType, platform, browser := parseUserAgent(req.UserAgent)

	// 解析地理位置（如果有IP地址）
	var location *models.GeoLocation
	if req.IPAddress != "" {
		location = extractLocationFromIP(req.IPAddress)
	}

	// 解析来源信息
	source, campaign := parseReferrer(req.Referrer)

	session := &models.AnalyticsUserSession{
		ID:         sessionID,
		UserID:     userID,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		StartTime:  now,
		LastActive: now,
		IsActive:   true,
		Referrer:   req.Referrer,
		Source:     source,
		Campaign:   campaign,
		DeviceType: deviceType,
		Platform:   platform,
		Browser:    browser,
		Location:   location,
		Metadata:   req.Metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.analyticsRepo.CreateUserSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	return session, nil
}

// UpdateSession 更新用户会话
func (s *SessionService) UpdateSession(ctx context.Context, sessionID string, req *UpdateSessionRequest) error {
	session, err := s.analyticsRepo.GetUserSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get user session: %w", err)
	}

	// 更新会话信息
	if req.EndTime != nil {
		session.EndTime = req.EndTime
		session.Duration = int64(req.EndTime.Sub(session.StartTime).Milliseconds())
		session.IsActive = false
	}

	if req.Metadata != nil {
		session.Metadata = req.Metadata
	}

	session.UpdatedAt = time.Now()

	if err := s.analyticsRepo.UpdateUserSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update user session: %w", err)
	}

	return nil
}

// GetSession 获取用户会话信息
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*models.AnalyticsUserSession, error) {
	return s.analyticsRepo.GetUserSession(ctx, sessionID)
}

// TrackPageView 页面浏览追踪服务
type PageViewService struct {
	analyticsRepo repositories.AnalyticsRepositoryInterface
}

// NewPageViewService 创建页面浏览服务实例
func NewPageViewService(analyticsRepo repositories.AnalyticsRepositoryInterface) *PageViewService {
	return &PageViewService{
		analyticsRepo: analyticsRepo,
	}
}

// TrackPageView 追踪页面浏览
func (p *PageViewService) TrackPageView(ctx context.Context, req *TrackPageViewRequest) error {
	// 验证会话是否存在
	session, err := p.analyticsRepo.GetUserSession(ctx, req.SessionID)
	if err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}

	// 解析URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	pageViewID := uuid.New().String()
	now := time.Now()

	pageView := &models.PageView{
		ID:            pageViewID,
		SessionID:     req.SessionID,
		UserID:        session.UserID,
		URL:           req.URL,
		Path:          parsedURL.Path,
		Title:         req.Title,
		Referrer:      req.Referrer,
		LoadTime:      req.Duration,
		StayTime:      req.Duration,
		ScrollDepth:   req.ScrollDepth,
		Viewport:      parseViewportSize(req.ViewportSize),
		ScreenSize:    parseScreenSize(req.ScreenSize),
		Interactions:  parseInteraction(req.Interaction),
		IsBounce:      req.IsBounce,
		IsExit:        req.ExitPage,
	}

	if err := p.analyticsRepo.CreatePageView(ctx, pageView); err != nil {
		return fmt.Errorf("failed to create page view: %w", err)
	}

	// 更新会话统计
	session.PageViews++
	session.LastActive = now
	if err := p.analyticsRepo.UpdateUserSession(ctx, session); err != nil {
		// 记录错误但不影响主流程
		fmt.Printf("Failed to update session stats: %v\n", err)
	}

	return nil
}

// EventTrackingService 事件追踪服务
type EventTrackingService struct {
	analyticsRepo repositories.AnalyticsRepositoryInterface
}

// NewEventTrackingService 创建事件追踪服务实例
func NewEventTrackingService(analyticsRepo repositories.AnalyticsRepositoryInterface) *EventTrackingService {
	return &EventTrackingService{
		analyticsRepo: analyticsRepo,
	}
}

// TrackEvent 追踪用户事件
func (e *EventTrackingService) TrackEvent(ctx context.Context, req *TrackEventRequest) error {
	// 验证会话是否存在
	session, err := e.analyticsRepo.GetUserSession(ctx, req.SessionID)
	if err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}

	eventID := uuid.New().String()
	now := time.Now()

	event := &models.AnalyticsUserEvent{
		ID:            eventID,
		SessionID:     req.SessionID,
		UserID:        session.UserID,
		URL:           req.URL,
		EventType:     req.EventType,
		EventCategory: req.EventCategory,
		EventAction:   req.EventAction,
		EventLabel:    req.EventLabel,
		EventValue:    req.EventValue,
		Element:       req.Element,
		Properties:    req.Properties,
		Timestamp:     now,
		CreatedAt:     now,
	}

	if err := e.analyticsRepo.CreateUserEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to create user event: %w", err)
	}

	// 更新会话最后活跃时间
	session.LastActive = now
	if err := e.analyticsRepo.UpdateUserSession(ctx, session); err != nil {
		fmt.Printf("Failed to update session last active time: %v\n", err)
	}

	return nil
}

// JourneyService 用户旅程服务
type JourneyService struct {
	analyticsRepo repositories.AnalyticsRepositoryInterface
}

// NewJourneyService 创建用户旅程服务实例
func NewJourneyService(analyticsRepo repositories.AnalyticsRepositoryInterface) *JourneyService {
	return &JourneyService{
		analyticsRepo: analyticsRepo,
	}
}

// CreateJourney 创建用户旅程
func (j *JourneyService) CreateJourney(ctx context.Context, req *CreateJourneyRequest) error {
	journeyID := uuid.New().String()
	now := time.Now()

	journey := &models.UserJourney{
		ID:          journeyID,
		UserID:      req.UserID,
		JourneyName: req.JourneyType,
		StartTime:   now,
		EndTime:     req.EndTime,
		Steps:       len(req.Steps),
		StepDetails: req.Steps,
		Properties:  req.Properties,
		CreatedAt:   now,
	}

	if err := j.analyticsRepo.CreateUserJourney(ctx, journey); err != nil {
		return fmt.Errorf("failed to create user journey: %w", err)
	}

	return nil
}

// BehaviorAnalysisService 行为分析服务
type BehaviorAnalysisService struct {
	analyticsRepo repositories.AnalyticsRepositoryInterface
}

// NewBehaviorAnalysisService 创建行为分析服务实例
func NewBehaviorAnalysisService(analyticsRepo repositories.AnalyticsRepositoryInterface) *BehaviorAnalysisService {
	return &BehaviorAnalysisService{
		analyticsRepo: analyticsRepo,
	}
}

// AnalyzeUserBehavior 分析用户行为模式
func (b *BehaviorAnalysisService) AnalyzeUserBehavior(ctx context.Context, userID string, startDate, endDate time.Time) (*UserBehaviorAnalysis, error) {
	// 获取会话摘要
	sessionSummary, err := b.analyticsRepo.GetUserSessionSummary(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user session summary: %w", err)
	}

	// 获取行为模式
	patterns, err := b.analyticsRepo.GetUserBehaviorPatterns(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user behavior patterns: %w", err)
	}

	// 分析页面浏览统计
	pageViewStats, err := b.analyticsRepo.GetPageViewStats(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get page view stats: %w", err)
	}

	// 分析事件统计
	eventStats, err := b.analyticsRepo.GetEventStats(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	analysis := &UserBehaviorAnalysis{
		UserID:           userID,
		AnalysisPeriod:   AnalysisPeriod{StartDate: startDate, EndDate: endDate},
		SessionSummary:   sessionSummary,
		BehaviorPatterns: patterns,
		PageViewStats:    pageViewStats,
		EventStats:       eventStats,
		GeneratedAt:      time.Now(),
	}

	return analysis, nil
}

// DetectBehaviorPatterns 检测行为模式
func (b *BehaviorAnalysisService) DetectBehaviorPatterns(ctx context.Context, userID string) error {
	// 获取用户最近的会话数据
	sessions, _, err := b.analyticsRepo.GetUserSessions(ctx, userID, 1, 100)
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	// 分析访问时间模式
	timePattern := b.analyzeTimePattern(sessions)
	if timePattern != nil {
		pattern := &models.BehaviorPattern{
			ID:          uuid.New().String(),
			Name:        "访问时间偏好",
			Type:        "time_preference",
			Description: "用户在特定时间段内更活跃",
			Pattern:     timePattern.Data,
			Confidence:  timePattern.Confidence,
			UserCount:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := b.analyticsRepo.CreateBehaviorPattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to create time pattern: %w", err)
		}
	}

	// 分析页面浏览模式
	pagePattern := b.analyzePagePattern(sessions)
	if pagePattern != nil {
		pattern := &models.BehaviorPattern{
			ID:          uuid.New().String(),
			Name:        "页面浏览模式",
			Type:        "navigation_pattern",
			Description: "用户有特定的页面浏览偏好",
			Pattern:     pagePattern.Data,
			Confidence:  pagePattern.Confidence,
			UserCount:   1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := b.analyticsRepo.CreateBehaviorPattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to create page pattern: %w", err)
		}
	}

	return nil
}

// RealTimeStatsService 实时统计服务
type RealTimeStatsService struct {
	analyticsRepo repositories.AnalyticsRepositoryInterface
}

// NewRealTimeStatsService 创建实时统计服务实例
func NewRealTimeStatsService(analyticsRepo repositories.AnalyticsRepositoryInterface) *RealTimeStatsService {
	return &RealTimeStatsService{
		analyticsRepo: analyticsRepo,
	}
}

// UpdateRealTimeStats 更新实时统计数据
func (r *RealTimeStatsService) UpdateRealTimeStats(ctx context.Context) error {
	now := time.Now()

	// 统计当前活跃用户数
	activeSessions, err := r.analyticsRepo.GetActiveUserSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active sessions: %w", err)
	}

	activeUsersCount := int64(len(activeSessions))
	statsID := uuid.New().String()

	activeUsersStat := &models.RealTimeStats{
		ID:         statsID,
		MetricName: "active_users",
		Value:      float64(activeUsersCount),
		Dimensions: map[string]interface{}{
			"timestamp": now,
		},
		Timestamp: now,
		CreatedAt: now,
	}

	if err := r.analyticsRepo.CreateRealTimeStats(ctx, activeUsersStat); err != nil {
		return fmt.Errorf("failed to create active users stat: %w", err)
	}

	return nil
}

// GetRealTimeDashboard 获取实时仪表板数据
func (r *RealTimeStatsService) GetRealTimeDashboard(ctx context.Context) (*RealTimeDashboard, error) {
	now := time.Now()

	// 获取活跃用户数
	activeUsersStats, err := r.analyticsRepo.GetRealTimeStats(ctx, "active_users", 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users stats: %w", err)
	}

	var activeUsers int64
	if len(activeUsersStats) > 0 {
		activeUsers = int64(activeUsersStats[0].Value)
	}

	// 获取今日页面浏览量
	pageViewStats, err := r.analyticsRepo.GetPageViewStats(ctx, now.Truncate(24*time.Hour), now)
	if err != nil {
		return nil, fmt.Errorf("failed to get page view stats: %w", err)
	}

	totalPageViews := int64(0)
	for _, stat := range pageViewStats {
		if views, ok := stat["views"].(int64); ok {
			totalPageViews += views
		}
	}

	// 获取今日事件数
	eventStats, err := r.analyticsRepo.GetEventStats(ctx, now.Truncate(24*time.Hour), now)
	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	totalEvents := int64(0)
	for _, stat := range eventStats {
		if count, ok := stat["count"].(int64); ok {
			totalEvents += count
		}
	}

	dashboard := &RealTimeDashboard{
		ActiveUsers:    activeUsers,
		PageViews:      totalPageViews,
		Events:         totalEvents,
		LastUpdated:    now,
		TimeRange:      "Last 24 hours",
	}

	return dashboard, nil
}

// 辅助函数

// parseUserAgent 解析用户代理字符串
func parseUserAgent(userAgent string) (deviceType, platform, browser string) {
	userAgent = strings.ToLower(userAgent)

	// 检测设备类型
	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") {
		deviceType = "mobile"
	} else if strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "ipad") {
		deviceType = "tablet"
	} else {
		deviceType = "desktop"
	}

	// 检测操作系统
	if strings.Contains(userAgent, "windows") {
		platform = "windows"
	} else if strings.Contains(userAgent, "mac") || strings.Contains(userAgent, "os x") {
		platform = "mac"
	} else if strings.Contains(userAgent, "linux") {
		platform = "linux"
	} else if strings.Contains(userAgent, "android") {
		platform = "android"
	} else if strings.Contains(userAgent, "ios") || strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		platform = "ios"
	} else {
		platform = "unknown"
	}

	// 检测浏览器
	if strings.Contains(userAgent, "chrome") && !strings.Contains(userAgent, "edg") {
		browser = "chrome"
	} else if strings.Contains(userAgent, "firefox") {
		browser = "firefox"
	} else if strings.Contains(userAgent, "safari") && !strings.Contains(userAgent, "chrome") {
		browser = "safari"
	} else if strings.Contains(userAgent, "edg") {
		browser = "edge"
	} else if strings.Contains(userAgent, "opera") {
		browser = "opera"
	} else {
		browser = "unknown"
	}

	return deviceType, platform, browser
}

// extractLocationFromIP 从IP地址提取地理位置信息
func extractLocationFromIP(ipAddress string) *models.GeoLocation {
	// 这里应该调用真实的地理位置服务API
	// 目前返回默认值
	return &models.GeoLocation{
		Country:     "Unknown",
		CountryCode: "XX",
		Region:      "Unknown",
		City:        "Unknown",
		Latitude:    0,
		Longitude:   0,
		Timezone:    "UTC",
	}
}

// parseReferrer 解析来源信息
func parseReferrer(referrer string) (source, campaign string) {
	if referrer == "" {
		return "direct", ""
	}

	parsedURL, err := url.Parse(referrer)
	if err != nil {
		return "unknown", ""
	}

	domain := parsedURL.Host
	if domain == "" {
		return "unknown", ""
	}

	// 识别常见搜索引擎
	if strings.Contains(domain, "google") {
		return "search", "google"
	} else if strings.Contains(domain, "baidu") {
		return "search", "baidu"
	} else if strings.Contains(domain, "bing") {
		return "search", "bing"
	} else if strings.Contains(domain, "facebook") || strings.Contains(domain, "fb.") {
		return "social", "facebook"
	} else if strings.Contains(domain, "twitter") || strings.Contains(domain, "t.co") {
		return "social", "twitter"
	} else if strings.Contains(domain, "linkedin") {
		return "social", "linkedin"
	}

	return "referral", domain
}

// TimePattern 时间模式
type TimePattern struct {
	Confidence float64                `json:"confidence"`
	Frequency  int                    `json:"frequency"`
	Data       map[string]interface{} `json:"data"`
}

// PagePattern 页面浏览模式
type PagePattern struct {
	Confidence float64                `json:"confidence"`
	Frequency  int                    `json:"frequency"`
	Data       map[string]interface{} `json:"data"`
}

// analyzeTimePattern 分析时间模式
func (b *BehaviorAnalysisService) analyzeTimePattern(sessions []*models.AnalyticsUserSession) *TimePattern {
	if len(sessions) < 5 {
		return nil
	}

	// 统计访问时间分布
	hourCount := make(map[int]int)
	for _, session := range sessions {
		hour := session.StartTime.Hour()
		hourCount[hour]++
	}

	// 找出最活跃的时间段
	maxCount := 0
	mostActiveHour := 0
	for hour, count := range hourCount {
		if count > maxCount {
			maxCount = count
			mostActiveHour = hour
		}
	}

	// 计算置信度
	confidence := float64(maxCount) / float64(len(sessions))
	if confidence < 0.6 {
		return nil
	}

	return &TimePattern{
		Confidence: confidence,
		Frequency:  maxCount,
		Data: map[string]interface{}{
			"most_active_hour": mostActiveHour,
			"hour_distribution": hourCount,
		},
	}
}

// analyzePagePattern 分析页面浏览模式
func (b *BehaviorAnalysisService) analyzePagePattern(sessions []*models.AnalyticsUserSession) *PagePattern {
	// 这里可以实现页面浏览模式分析逻辑
	// 目前返回nil，表示未检测到明显模式
	return nil
}

// 请求结构体

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	IPAddress string                 `json:"ip_address" validate:"required"`
	UserAgent string                 `json:"user_agent" validate:"required"`
	Referrer  string                 `json:"referrer"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	EndTime  *time.Time             `json:"end_time"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TrackPageViewRequest 页面浏览追踪请求
type TrackPageViewRequest struct {
	SessionID    string                 `json:"session_id" validate:"required"`
	URL          string                 `json:"url" validate:"required,url"`
	Title        string                 `json:"title"`
	Referrer     string                 `json:"referrer"`
	Duration     int64                  `json:"duration"`        // 停留时间（毫秒）
	ScrollDepth  int                    `json:"scroll_depth"`    // 滚动深度百分比
	ViewportSize string                 `json:"viewport_size"`   // 视口大小 "1024x768"
	ScreenSize   string                 `json:"screen_size"`     // 屏幕大小 "1920x1080"
	Interaction  string                 `json:"interaction"`     // 交互类型
	IsBounce     bool                   `json:"is_bounce"`       // 是否跳出
	ExitPage     bool                   `json:"exit_page"`       // 是否退出页
	EntryPage    bool                   `json:"entry_page"`      // 是否入口页
	Properties   map[string]interface{} `json:"properties"`
}

// TrackEventRequest 事件追踪请求
type TrackEventRequest struct {
	SessionID     string                 `json:"session_id" validate:"required"`
	EventType     string                 `json:"event_type" validate:"required"`
	EventCategory string                 `json:"event_category" validate:"required"`
	EventAction   string                 `json:"event_action" validate:"required"`
	EventLabel    string                 `json:"event_label"`
	EventValue    float64                `json:"event_value"`
	URL           string                 `json:"url"`
	Element       string                 `json:"element"`        // 触发事件的元素
	Properties    map[string]interface{} `json:"properties"`
}

// CreateJourneyRequest 创建用户旅程请求
type CreateJourneyRequest struct {
	UserID       string                 `json:"user_id" validate:"required"`
	JourneyType  string                 `json:"journey_type" validate:"required"`
	EndTime      *time.Time             `json:"end_time"`
	Steps        []models.JourneyStep   `json:"steps"`
	CurrentStep  int                    `json:"current_step"`
	IsCompleted  bool                   `json:"is_completed"`
	Properties   map[string]interface{} `json:"properties"`
}

// 响应结构体

// AnalysisPeriod 分析周期
type AnalysisPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// UserBehaviorAnalysis 用户行为分析结果
type UserBehaviorAnalysis struct {
	UserID           string                    `json:"user_id"`
	AnalysisPeriod   AnalysisPeriod            `json:"analysis_period"`
	SessionSummary   *models.UserSessionSummary `json:"session_summary"`
	BehaviorPatterns []*models.BehaviorPattern `json:"behavior_patterns"`
	PageViewStats    []map[string]interface{}  `json:"page_view_stats"`
	EventStats       []map[string]interface{}  `json:"event_stats"`
	GeneratedAt      time.Time                 `json:"generated_at"`
}

// RealTimeDashboard 实时仪表板数据
type RealTimeDashboard struct {
	ActiveUsers int64     `json:"active_users"`
	PageViews   int64     `json:"page_views"`
	Events      int64     `json:"events"`
	LastUpdated time.Time `json:"last_updated"`
	TimeRange   string    `json:"time_range"`
}