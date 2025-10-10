package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// AnalyticsRepository 用户行为分析数据访问层
type AnalyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository 创建分析数据访问层实例
func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// 用户会话相关操作

// CreateUserSession 创建用户会话
func (r *AnalyticsRepository) CreateUserSession(ctx context.Context, session *models.AnalyticsUserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetUserSession 获取用户会话
func (r *AnalyticsRepository) GetUserSession(ctx context.Context, sessionID string) (*models.AnalyticsUserSession, error) {
	var session models.AnalyticsUserSession
	err := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateUserSession 更新用户会话
func (r *AnalyticsRepository) UpdateUserSession(ctx context.Context, session *models.AnalyticsUserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// GetUserSessions 获取用户会话列表
func (r *AnalyticsRepository) GetUserSessions(ctx context.Context, userID string, page, pageSize int) ([]*models.AnalyticsUserSession, int64, error) {
	var sessions []*models.AnalyticsUserSession
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.AnalyticsUserSession{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("start_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&sessions).Error

	return sessions, total, err
}

// GetActiveUserSessions 获取活跃用户会话
func (r *AnalyticsRepository) GetActiveUserSessions(ctx context.Context) ([]*models.AnalyticsUserSession, error) {
	var sessions []*models.AnalyticsUserSession
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND start_time > ?", true, time.Now().Add(-24*time.Hour)).
		Find(&sessions).Error
	return sessions, err
}

// 页面浏览相关操作

// CreatePageView 创建页面浏览记录
func (r *AnalyticsRepository) CreatePageView(ctx context.Context, pageView *models.PageView) error {
	return r.db.WithContext(ctx).Create(pageView).Error
}

// GetPageViews 获取页面浏览记录
func (r *AnalyticsRepository) GetPageViews(ctx context.Context, sessionID string, page, pageSize int) ([]*models.PageView, int64, error) {
	var pageViews []*models.PageView
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&models.PageView{}).Where("session_id = ?", sessionID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("timestamp DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&pageViews).Error

	return pageViews, total, err
}

// GetPageViewsByURL 获取指定URL的页面浏览记录
func (r *AnalyticsRepository) GetPageViewsByURL(ctx context.Context, url string, startTime, endTime time.Time) ([]*models.PageView, error) {
	var pageViews []*models.PageView
	err := r.db.WithContext(ctx).
		Where("url = ? AND timestamp BETWEEN ? AND ?", url, startTime, endTime).
		Order("timestamp DESC").
		Find(&pageViews).Error
	return pageViews, err
}

// 用户事件相关操作

// CreateUserEvent 创建用户事件
func (r *AnalyticsRepository) CreateUserEvent(ctx context.Context, event *models.AnalyticsUserEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetUserEvents 获取用户事件
func (r *AnalyticsRepository) GetUserEvents(ctx context.Context, sessionID string, eventType string, page, pageSize int) ([]*models.AnalyticsUserEvent, int64, error) {
	var events []*models.AnalyticsUserEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AnalyticsUserEvent{}).Where("session_id = ?", sessionID)
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Order("timestamp DESC").Limit(pageSize).Offset(offset).Find(&events).Error

	return events, total, err
}

// GetEventsByType 获取指定类型的事件
func (r *AnalyticsRepository) GetEventsByType(ctx context.Context, eventType string, startTime, endTime time.Time) ([]*models.AnalyticsUserEvent, error) {
	var events []*models.AnalyticsUserEvent
	err := r.db.WithContext(ctx).
		Where("event_type = ? AND timestamp BETWEEN ? AND ?", eventType, startTime, endTime).
		Order("timestamp DESC").
		Find(&events).Error
	return events, err
}

// 用户旅程相关操作

// CreateUserJourney 创建用户旅程
func (r *AnalyticsRepository) CreateUserJourney(ctx context.Context, journey *models.UserJourney) error {
	return r.db.WithContext(ctx).Create(journey).Error
}

// GetUserJourneys 获取用户旅程
func (r *AnalyticsRepository) GetUserJourneys(ctx context.Context, userID string, journeyType string, page, pageSize int) ([]*models.UserJourney, int64, error) {
	var journeys []*models.UserJourney
	var total int64

	query := r.db.WithContext(ctx).Model(&models.UserJourney{}).Where("user_id = ?", userID)
	if journeyType != "" {
		query = query.Where("journey_type = ?", journeyType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Order("start_time DESC").Limit(pageSize).Offset(offset).Find(&journeys).Error

	return journeys, total, err
}

// 用户细分相关操作

// CreateUserSegment 创建用户细分
func (r *AnalyticsRepository) CreateUserSegment(ctx context.Context, segment *models.UserSegment) error {
	return r.db.WithContext(ctx).Create(segment).Error
}

// GetUserSegments 获取用户细分
func (r *AnalyticsRepository) GetUserSegments(ctx context.Context, isActive *bool, page, pageSize int) ([]*models.UserSegment, int64, error) {
	var segments []*models.UserSegment
	var total int64

	query := r.db.WithContext(ctx).Model(&models.UserSegment{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&segments).Error

	return segments, total, err
}

// 行为模式相关操作

// CreateBehaviorPattern 创建行为模式
func (r *AnalyticsRepository) CreateBehaviorPattern(ctx context.Context, pattern *models.BehaviorPattern) error {
	return r.db.WithContext(ctx).Create(pattern).Error
}

// GetBehaviorPatterns 获取行为模式
func (r *AnalyticsRepository) GetBehaviorPatterns(ctx context.Context, userID string, patternType string, page, pageSize int) ([]*models.BehaviorPattern, int64, error) {
	var patterns []*models.BehaviorPattern
	var total int64

	query := r.db.WithContext(ctx).Model(&models.BehaviorPattern{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if patternType != "" {
		query = query.Where("pattern_type = ?", patternType)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&patterns).Error

	return patterns, total, err
}

// 漏斗分析相关操作

// CreateFunnelAnalysis 创建漏斗分析
func (r *AnalyticsRepository) CreateFunnelAnalysis(ctx context.Context, funnel *models.FunnelAnalysis) error {
	return r.db.WithContext(ctx).Create(funnel).Error
}

// GetFunnelAnalyses 获取漏斗分析
func (r *AnalyticsRepository) GetFunnelAnalyses(ctx context.Context, funnelName string, page, pageSize int) ([]*models.FunnelAnalysis, int64, error) {
	var funnels []*models.FunnelAnalysis
	var total int64

	query := r.db.WithContext(ctx).Model(&models.FunnelAnalysis{})
	if funnelName != "" {
		query = query.Where("funnel_name = ?", funnelName)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&funnels).Error

	return funnels, total, err
}

// 留存分析相关操作

// CreateRetentionAnalysis 创建留存分析
func (r *AnalyticsRepository) CreateRetentionAnalysis(ctx context.Context, retention *models.RetentionAnalysis) error {
	return r.db.WithContext(ctx).Create(retention).Error
}

// GetRetentionAnalyses 获取留存分析
func (r *AnalyticsRepository) GetRetentionAnalyses(ctx context.Context, cohortDate time.Time, page, pageSize int) ([]*models.RetentionAnalysis, int64, error) {
	var retentions []*models.RetentionAnalysis
	var total int64

	query := r.db.WithContext(ctx).Model(&models.RetentionAnalysis{})
	if !cohortDate.IsZero() {
		query = query.Where("cohort_date = ?", cohortDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Order("cohort_date DESC, period_number ASC").Limit(pageSize).Offset(offset).Find(&retentions).Error

	return retentions, total, err
}

// 实时统计相关操作

// CreateRealTimeStats 创建实时统计
func (r *AnalyticsRepository) CreateRealTimeStats(ctx context.Context, stats *models.RealTimeStats) error {
	return r.db.WithContext(ctx).Create(stats).Error
}

// GetRealTimeStats 获取实时统计
func (r *AnalyticsRepository) GetRealTimeStats(ctx context.Context, metricName string, limit int) ([]*models.RealTimeStats, error) {
	var stats []*models.RealTimeStats
	query := r.db.WithContext(ctx).Order("timestamp DESC")

	if metricName != "" {
		query = query.Where("metric_name = ?", metricName)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&stats).Error
	return stats, err
}

// 聚合查询操作

// GetDailyActiveUsers 获取日活跃用户数
func (r *AnalyticsRepository) GetDailyActiveUsers(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.WithContext(ctx).
		Table("user_sessions").
		Select("DATE(start_time) as date, COUNT(DISTINCT user_id) as active_users").
		Where("start_time BETWEEN ? AND ?", startDate, endDate).
		Group("DATE(start_time)").
		Order("date ASC").
		Find(&results).Error

	return results, err
}

// GetPageViewStats 获取页面浏览统计
func (r *AnalyticsRepository) GetPageViewStats(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.WithContext(ctx).
		Table("page_views").
		Select("url, COUNT(*) as views, COUNT(DISTINCT session_id) as unique_sessions, AVG(duration) as avg_duration").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("url").
		Order("views DESC").
		Find(&results).Error

	return results, err
}

// GetEventStats 获取事件统计
func (r *AnalyticsRepository) GetEventStats(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.WithContext(ctx).
		Table("user_events").
		Select("event_type, COUNT(*) as count, COUNT(DISTINCT session_id) as unique_sessions").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("event_type").
		Order("count DESC").
		Find(&results).Error

	return results, err
}

// GetUserRetentionCohorts 获取用户留存队列
func (r *AnalyticsRepository) GetUserRetentionCohorts(ctx context.Context, cohortStartDate, cohortEndDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// 获取用户首次访问日期作为队列标识
	err := r.db.WithContext(ctx).
		Table("user_sessions").
		Select(`
			DATE(MIN(start_time)) as cohort_date,
			user_id,
			COUNT(*) as total_sessions,
			MAX(start_time) as last_session
		`).
		Where("start_time BETWEEN ? AND ?", cohortStartDate, cohortEndDate).
		Group("user_id").
		Find(&results).Error

	return results, err
}

// GetTopPages 获取热门页面
func (r *AnalyticsRepository) GetTopPages(ctx context.Context, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.WithContext(ctx).
		Table("page_views").
		Select("url, COUNT(*) as page_views, COUNT(DISTINCT user_id) as unique_users, AVG(duration) as avg_duration").
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Group("url").
		Order("page_views DESC").
		Limit(limit).
		Find(&results).Error

	return results, err
}

// GetUserSessionSummary 获取用户会话摘要
func (r *AnalyticsRepository) GetUserSessionSummary(ctx context.Context, userID string, startDate, endDate time.Time) (*models.UserSessionSummary, error) {
	var summary models.UserSessionSummary

	// 获取基础会话统计
	err := r.db.WithContext(ctx).
		Model(&models.AnalyticsUserSession{}).
		Select(`
			COUNT(*) as total_sessions,
			COUNT(CASE WHEN is_active = 1 THEN 1 END) as active_sessions,
			AVG(duration) as avg_duration,
			MAX(duration) as max_duration,
			MIN(start_time) as first_session,
			MAX(start_time) as last_session
		`).
		Where("user_id = ? AND start_time BETWEEN ? AND ?", userID, startDate, endDate).
		Scan(&summary).Error

	if err != nil {
		return nil, err
	}

	// 获取页面浏览统计
	var pageViewStats struct {
		TotalPageViews int64   `json:"total_page_views"`
		UniquePages    int64   `json:"unique_pages"`
		AvgDuration    float64 `json:"avg_page_duration"`
	}

	err = r.db.WithContext(ctx).
		Table("page_views pv").
		Joins("JOIN user_sessions us ON pv.session_id = us.id").
		Select(`
			COUNT(*) as total_page_views,
			COUNT(DISTINCT pv.url) as unique_pages,
			AVG(pv.duration) as avg_duration
		`).
		Where("us.user_id = ? AND pv.timestamp BETWEEN ? AND ?", userID, startDate, endDate).
		Scan(&pageViewStats).Error

	if err != nil {
		return nil, err
	}

	summary.TotalPageViews = pageViewStats.TotalPageViews
	summary.UniquePages = pageViewStats.UniquePages
	summary.AvgPageDuration = pageViewStats.AvgDuration

	// 获取事件统计
	var eventStats struct {
		TotalEvents int64 `json:"total_events"`
		ClickEvents int64 `json:"click_events"`
		FormEvents  int64 `json:"form_events"`
	}

	err = r.db.WithContext(ctx).
		Table("user_events ue").
		Joins("JOIN user_sessions us ON ue.session_id = us.id").
		Select(`
			COUNT(*) as total_events,
			COUNT(CASE WHEN ue.event_type = 'click' THEN 1 END) as click_events,
			COUNT(CASE WHEN ue.event_type = 'form_submit' THEN 1 END) as form_events
		`).
		Where("us.user_id = ? AND ue.timestamp BETWEEN ? AND ?", userID, startDate, endDate).
		Scan(&eventStats).Error

	if err != nil {
		return nil, err
	}

	summary.TotalEvents = eventStats.TotalEvents
	summary.ClickEvents = eventStats.ClickEvents
	summary.FormEvents = eventStats.FormEvents

	return &summary, nil
}

// GetUserBehaviorPatterns 获取用户行为模式
func (r *AnalyticsRepository) GetUserBehaviorPatterns(ctx context.Context, userID string, startDate, endDate time.Time) ([]*models.BehaviorPattern, error) {
	var patterns []*models.BehaviorPattern

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Order("created_at DESC").
		Find(&patterns).Error

	return patterns, err
}

// GetFunnelConversionRate 获取漏斗转化率
func (r *AnalyticsRepository) GetFunnelConversionRate(ctx context.Context, funnelName string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.WithContext(ctx).
		Table("funnel_analyses").
		Select(`
			step_name,
			step_order,
			COUNT(DISTINCT user_id) as users,
			SUM(conversion_rate) / COUNT(*) as avg_conversion_rate,
			AVG(time_to_convert) as avg_time_to_convert
		`).
		Where("funnel_name = ? AND created_at BETWEEN ? AND ?", funnelName, startDate, endDate).
		Group("step_name, step_order").
		Order("step_order ASC").
		Find(&results).Error

	return results, err
}

// CleanupOldAnalyticsData 清理旧的分析数据
func (r *AnalyticsRepository) CleanupOldAnalyticsData(ctx context.Context, retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// 清理旧的页面浏览记录
	if err := r.db.WithContext(ctx).Where("timestamp < ?", cutoffDate).Delete(&models.PageView{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old page views: %w", err)
	}

	// 清理旧的用户事件
	if err := r.db.WithContext(ctx).Where("timestamp < ?", cutoffDate).Delete(&models.AnalyticsUserEvent{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old user events: %w", err)
	}

	// 清理旧的会话记录（非活跃的）
	if err := r.db.WithContext(ctx).Where("start_time < ? AND is_active = ?", cutoffDate, false).Delete(&models.AnalyticsUserSession{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old user sessions: %w", err)
	}

	// 清理旧的实时统计数据
	if err := r.db.WithContext(ctx).Where("timestamp < ?", cutoffDate).Delete(&models.RealTimeStats{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old real-time stats: %w", err)
	}

	return nil
}

// AnalyticsRepositoryInterface 分析数据访问层接口
type AnalyticsRepositoryInterface interface {
	// 用户会话操作
	CreateUserSession(ctx context.Context, session *models.AnalyticsUserSession) error
	GetUserSession(ctx context.Context, sessionID string) (*models.AnalyticsUserSession, error)
	UpdateUserSession(ctx context.Context, session *models.AnalyticsUserSession) error
	GetUserSessions(ctx context.Context, userID string, page, pageSize int) ([]*models.AnalyticsUserSession, int64, error)
	GetActiveUserSessions(ctx context.Context) ([]*models.AnalyticsUserSession, error)

	// 页面浏览操作
	CreatePageView(ctx context.Context, pageView *models.PageView) error
	GetPageViews(ctx context.Context, sessionID string, page, pageSize int) ([]*models.PageView, int64, error)
	GetPageViewsByURL(ctx context.Context, url string, startTime, endTime time.Time) ([]*models.PageView, error)

	// 用户事件操作
	CreateUserEvent(ctx context.Context, event *models.AnalyticsUserEvent) error
	GetUserEvents(ctx context.Context, sessionID string, eventType string, page, pageSize int) ([]*models.AnalyticsUserEvent, int64, error)
	GetEventsByType(ctx context.Context, eventType string, startTime, endTime time.Time) ([]*models.AnalyticsUserEvent, error)

	// 用户旅程操作
	CreateUserJourney(ctx context.Context, journey *models.UserJourney) error
	GetUserJourneys(ctx context.Context, userID string, journeyType string, page, pageSize int) ([]*models.UserJourney, int64, error)

	// 用户细分操作
	CreateUserSegment(ctx context.Context, segment *models.UserSegment) error
	GetUserSegments(ctx context.Context, isActive *bool, page, pageSize int) ([]*models.UserSegment, int64, error)

	// 行为模式操作
	CreateBehaviorPattern(ctx context.Context, pattern *models.BehaviorPattern) error
	GetBehaviorPatterns(ctx context.Context, userID string, patternType string, page, pageSize int) ([]*models.BehaviorPattern, int64, error)

	// 漏斗分析操作
	CreateFunnelAnalysis(ctx context.Context, funnel *models.FunnelAnalysis) error
	GetFunnelAnalyses(ctx context.Context, funnelName string, page, pageSize int) ([]*models.FunnelAnalysis, int64, error)

	// 留存分析操作
	CreateRetentionAnalysis(ctx context.Context, retention *models.RetentionAnalysis) error
	GetRetentionAnalyses(ctx context.Context, cohortDate time.Time, page, pageSize int) ([]*models.RetentionAnalysis, int64, error)

	// 实时统计操作
	CreateRealTimeStats(ctx context.Context, stats *models.RealTimeStats) error
	GetRealTimeStats(ctx context.Context, metricName string, limit int) ([]*models.RealTimeStats, error)

	// 聚合查询操作
	GetDailyActiveUsers(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetPageViewStats(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetEventStats(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error)
	GetUserRetentionCohorts(ctx context.Context, cohortStartDate, cohortEndDate time.Time) ([]map[string]interface{}, error)
	GetTopPages(ctx context.Context, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error)
	GetUserSessionSummary(ctx context.Context, userID string, startDate, endDate time.Time) (*models.UserSessionSummary, error)
	GetUserBehaviorPatterns(ctx context.Context, userID string, startDate, endDate time.Time) ([]*models.BehaviorPattern, error)
	GetFunnelConversionRate(ctx context.Context, funnelName string, startDate, endDate time.Time) ([]map[string]interface{}, error)

	// 维护操作
	CleanupOldAnalyticsData(ctx context.Context, retentionDays int) error
}

// 确保AnalyticsRepository实现了接口
var _ AnalyticsRepositoryInterface = (*AnalyticsRepository)(nil)