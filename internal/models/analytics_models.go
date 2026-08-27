package models

import (
	"time"
)

// AnalyticsUserSession 用户行为分析会话
type AnalyticsUserSession struct {
	ID         string                 `json:"id" gorm:"primaryKey;size:64"`
	UserID     string                 `json:"user_id" gorm:"size:64;not null;index"`
	IPAddress  string                 `json:"ip_address" gorm:"size:45"`
	UserAgent  string                 `json:"user_agent" gorm:"size:500"`
	StartTime  time.Time              `json:"start_time" gorm:"not null;index"`
	EndTime    *time.Time             `json:"end_time"`
	Duration   int64                  `json:"duration"` // 毫秒
	IsActive   bool                   `json:"is_active" gorm:"default:true;index"`
	PageViews  int                    `json:"page_views" gorm:"default:0"`
	LastActive time.Time              `json:"last_active" gorm:"not null"`
	Referrer   string                 `json:"referrer" gorm:"size:500"`
	Source     string                 `json:"source" gorm:"size:100"` // 来源：direct, search, social, email等
	Campaign   string                 `json:"campaign" gorm:"size:100"`
	DeviceType string                 `json:"device_type" gorm:"size:50"` // desktop, mobile, tablet
	Platform   string                 `json:"platform" gorm:"size:50"`    // windows, mac, linux, ios, android
	Browser    string                 `json:"browser" gorm:"size:100"`
	Location   *GeoLocation           `json:"location" gorm:"serializer:json"`
	Metadata   map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// GeoLocation 地理位置
type GeoLocation struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

// PageView 页面浏览
type PageView struct {
	ID            string                 `json:"id" gorm:"primaryKey;size:64"`
	SessionID     string                 `json:"session_id" gorm:"size:64;not null;index"`
	UserID        string                 `json:"user_id" gorm:"size:64;not null;index"`
	URL           string                 `json:"url" gorm:"size:2000;not null;index"`
	Path          string                 `json:"path" gorm:"size:500;not null;index"`
	Title         string                 `json:"title" gorm:"size:200"`
	Referrer      string                 `json:"referrer" gorm:"size:500"`
	Viewport      *ViewportSize          `json:"viewport" gorm:"serializer:json"`
	ScreenSize    *ScreenSize            `json:"screen_size" gorm:"serializer:json"`
	LoadTime      int64                  `json:"load_time"`    // 毫秒
	StayTime      int64                  `json:"stay_time"`    // 毫秒
	ScrollDepth   int                    `json:"scroll_depth"` // 像素
	Interactions  int                    `json:"interactions" gorm:"default:0"`
	IsBounce      bool                   `json:"is_bounce" gorm:"default:false"`
	IsExit        bool                   `json:"is_exit" gorm:"default:false"`
	EventCategory string                 `json:"event_category" gorm:"size:50"`
	EventAction   string                 `json:"event_action" gorm:"size:50"`
	EventLabel    string                 `json:"event_label" gorm:"size:100"`
	EventValue    float64                `json:"event_value"`
	CustomData    map[string]interface{} `json:"custom_data" gorm:"serializer:json"`
	CreatedAt     time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// ViewportSize 视口大小
type ViewportSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ScreenSize 屏幕大小
type ScreenSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// UserEvent 用户事件
type AnalyticsUserEvent struct {
	ID            string                 `json:"id" gorm:"primaryKey;size:64"`
	SessionID     string                 `json:"session_id" gorm:"size:64;not null;index"`
	UserID        string                 `json:"user_id" gorm:"size:64;not null;index"`
	PageID        string                 `json:"page_id" gorm:"size:64;index"`
	URL           string                 `json:"url" gorm:"size:2000;not null"`
	EventType     string                 `json:"event_type" gorm:"size:50;not null;index"`
	EventCategory string                 `json:"event_category" gorm:"size:50;not null;index"`
	EventAction   string                 `json:"event_action" gorm:"size:50;not null;index"`
	EventLabel    string                 `json:"event_label" gorm:"size:100"`
	EventValue    float64                `json:"event_value"`
	Element       string                 `json:"element" gorm:"size:200"` // CSS选择器或元素ID
	Properties    map[string]interface{} `json:"properties" gorm:"serializer:json"`
	Timestamp     time.Time              `json:"timestamp" gorm:"not null;index"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	CreatedAt     time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserJourney 用户旅程
type UserJourney struct {
	ID          string                 `json:"id" gorm:"primaryKey;size:64"`
	UserID      string                 `json:"user_id" gorm:"size:64;not null;index"`
	JourneyName string                 `json:"journey_name" gorm:"size:100;not null"`
	StartTime   time.Time              `json:"start_time" gorm:"not null;index"`
	EndTime     *time.Time             `json:"end_time"`
	Duration    int64                  `json:"duration"`                             // 秒
	Status      string                 `json:"status" gorm:"size:20;default:active"` // active, completed, abandoned
	Steps       int                    `json:"steps" gorm:"default:0"`
	StepDetails []JourneyStep          `json:"step_details" gorm:"serializer:json"`
	Completion  float64                `json:"completion" gorm:"default:0"` // 完成度百分比
	GoalType    string                 `json:"goal_type" gorm:"size:50"`    // conversion, retention, engagement
	GoalValue   float64                `json:"goal_value"`
	Tags        []string               `json:"tags" gorm:"serializer:json"`
	Properties  map[string]interface{} `json:"properties" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// JourneyStep 旅程步骤
type JourneyStep struct {
	StepID      int       `json:"step_id"`
	EventType   string    `json:"event_type"`
	EventAction string    `json:"event_action"`
	EventLabel  string    `json:"event_label"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    int       `json:"duration"` // 毫秒
}

// UserSegment 用户分段
type UserSegment struct {
	ID          string                 `json:"id" gorm:"primaryKey;size:64"`
	Name        string                 `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string                 `json:"description" gorm:"size:500"`
	Criteria    map[string]interface{} `json:"criteria" gorm:"serializer:json"` // 分段条件
	UserCount   int                    `json:"user_count" gorm:"default:0"`
	IsActive    bool                   `json:"is_active" gorm:"default:true"`
	Tags        []string               `json:"tags" gorm:"serializer:json"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserSegmentMembership 用户分段成员关系
type UserSegmentMembership struct {
	ID         string                 `json:"id" gorm:"primaryKey;size:64"`
	UserID     string                 `json:"user_id" gorm:"size:64;not null;index"`
	SegmentID  string                 `json:"segment_id" gorm:"size:64;not null;index"`
	AddedAt    time.Time              `json:"added_at" gorm:"not null"`
	AddedBy    string                 `json:"added_by" gorm:"size:64"` // 添加该成员的用户或系统
	IsActive   bool                   `json:"is_active" gorm:"default:true"`
	Properties map[string]interface{} `json:"properties" gorm:"serializer:json"`
}

// BehaviorPattern 行为模式
type BehaviorPattern struct {
	ID          string                 `json:"id" gorm:"primaryKey;size:64"`
	Name        string                 `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Type        string                 `json:"type" gorm:"size:50;not null"` // navigation, engagement, conversion, retention
	Description string                 `json:"description" gorm:"size:500"`
	Pattern     map[string]interface{} `json:"pattern" gorm:"serializer:json"` // 模式定义
	Threshold   float64                `json:"threshold"`                      // 阈值
	Confidence  float64                `json:"confidence"`                     // 置信度
	UserCount   int                    `json:"user_count"`
	IsPositive  bool                   `json:"is_positive"` // 是否为积极行为
	IsActive    bool                   `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserBehaviorRecord 用户行为记录
type UserBehaviorRecord struct {
	ID         string                 `json:"id" gorm:"primaryKey;size:64"`
	UserID     string                 `json:"user_id" gorm:"size:64;not null;index"`
	PatternID  string                 `json:"pattern_id" gorm:"size:64;index"`
	EventType  string                 `json:"event_type" gorm:"size:50;not null;index"`
	Score      float64                `json:"score"`      // 模式匹配分数
	Confidence float64                `json:"confidence"` // 置信度
	Timestamp  time.Time              `json:"timestamp" gorm:"not null;index"`
	Properties map[string]interface{} `json:"properties" gorm:"serializer:json"`
	IsMatch    bool                   `json:"is_match"`
	Feedback   string                 `json:"feedback"` // 用户反馈
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// FunnelAnalysis 漏斗分析
type FunnelAnalysis struct {
	ID             string       `json:"id" gorm:"primaryKey;size:64"`
	Name           string       `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description    string       `json:"description" gorm:"size:500"`
	Steps          []FunnelStep `json:"steps" gorm:"serializer:json"`
	Period         string       `json:"period" gorm:"size:20"` // daily, weekly, monthly
	StartDate      time.Time    `json:"start_date" gorm:"not null;index"`
	EndDate        time.Time    `json:"end_date" gorm:"not null;index"`
	TotalUsers     int64        `json:"total_users"`
	CompletionRate float64      `json:"completion_rate"`
	AvgTime        int64        `json:"avg_time"` // 平均时间（秒）
	CreatedAt      time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// FunnelStep 漏斗步骤
type FunnelStep struct {
	StepNumber     int     `json:"step_number"`
	StepName       string  `json:"step_name"`
	EventType      string  `json:"event_type"`
	EventAction    string  `json:"event_action"`
	EventLabel     string  `json:"event_label"`
	URL            string  `json:"url"`
	UserCount      int64   `json:"user_count"`
	DropoffCount   int64   `json:"dropoff_count"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgTime        int64   `json:"avg_time"` // 秒
	IsGoalStep     bool    `json:"is_goal_step"`
}

// RetentionAnalysis 留存分析
type RetentionAnalysis struct {
	ID            string          `json:"id" gorm:"primaryKey;size:64"`
	Name          string          `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description   string          `json:"description" gorm:"size:500"`
	PeriodType    string          `json:"period_type" gorm:"size:20"` // daily, weekly, monthly
	PeriodDays    int             `json:"period_days" gorm:"not null"`
	CohortDate    time.Time       `json:"cohort_date" gorm:"not null;index"`
	CohortSize    int64           `json:"cohort_size"`
	RetentionData []RetentionData `json:"retention_data" gorm:"serializer:json"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// RetentionData 留存数据
type RetentionData struct {
	PeriodNumber  int     `json:"period_number"`
	PeriodName    string  `json:"period_name"`
	UserCount     int64   `json:"user_count"`
	RetentionRate float64 `json:"retention_rate"`
}

// AnalyticsReport 分析报告
type AnalyticsReportData struct {
	ID          string                 `json:"id" gorm:"primaryKey;size:64"`
	Name        string                 `json:"name" gorm:"size:100;not null"`
	Type        string                 `json:"type" gorm:"size:50;not null"` // user_behavior, performance, conversion, retention
	Description string                 `json:"description" gorm:"size:500"`
	Period      string                 `json:"period" gorm:"size:20"` // daily, weekly, monthly
	StartDate   time.Time              `json:"start_date" gorm:"not null"`
	EndDate     time.Time              `json:"end_date" gorm:"not null"`
	Data        map[string]interface{} `json:"data" gorm:"serializer:json"`
	Insights    []ReportInsight        `json:"insights" gorm:"serializer:json"`
	Status      string                 `json:"status" gorm:"size:20;default:draft"` // draft, generating, completed, published
	CreatedBy   string                 `json:"created_by" gorm:"size:64"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	PublishedAt *time.Time             `json:"published_at"`
}

// ReportInsight 报告洞察
type ReportInsight struct {
	Type           string  `json:"type"` // trend, anomaly, opportunity, risk
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Value          float64 `json:"value"`
	Change         float64 `json:"change"` // 变化百分比
	Impact         string  `json:"impact"` // high, medium, low
	Recommendation string  `json:"recommendation"`
}

// HeatmapData 热力图数据
type HeatmapData struct {
	ID          string    `json:"id" gorm:"primaryKey;size:64"`
	PagePath    string    `json:"page_path" gorm:"size:500;not null;index"`
	ClickCount  int64     `json:"click_count"`
	DwellTime   int64     `json:"dwell_time"` // 停留时间（毫秒）
	ScrollDepth int       `json:"scroll_depth"`
	Date        time.Time `json:"date" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// ClickstreamRecord 点击流记录
type ClickstreamRecord struct {
	ID         string                 `json:"id" gorm:"primaryKey;size:64"`
	SessionID  string                 `json:"session_id" gorm:"size:64;not null;index"`
	UserID     string                 `json:"user_id" gorm:"size:64;not null;index"`
	X          int                    `json:"x"`                         // 点击X坐标
	Y          int                    `json:"y"`                         // 点击Y坐标
	Element    string                 `json:"element" gorm:"size:200"`   // 被点击元素
	Selector   string                 `json:"selector" gorm:"size:500"`  // CSS选择器
	EventType  string                 `json:"event_type" gorm:"size:50"` // click, scroll, form
	URL        string                 `json:"url" gorm:"size:2000;not null;index"`
	Timestamp  time.Time              `json:"timestamp" gorm:"not null;index"`
	Properties map[string]interface{} `json:"properties" gorm:"serializer:json"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// FormInteraction 表单交互
type FormInteraction struct {
	ID           string                 `json:"id" gorm:"primaryKey;size:64"`
	SessionID    string                 `json:"session_id" gorm:"size:64;not null;index"`
	UserID       string                 `json:"user_id" gorm:"size:64;not null;index"`
	FormID       string                 `json:"form_id" gorm:"size:64;index"`
	FormName     string                 `json:"form_name" gorm:"size:100"`
	FieldType    string                 `json:"field_type" gorm:"size:50"` // text, email, select, checkbox
	FieldName    string                 `json:"field_name" gorm:"size:100"`
	FieldValue   string                 `json:"field_value" gorm:"size:1000"`
	EventType    string                 `json:"event_type" gorm:"size:50"` // focus, blur, change, submit
	URL          string                 `json:"url" gorm:"size:2000;not null;index"`
	Timestamp    time.Time              `json:"timestamp" gorm:"not null;index"`
	IsValid      bool                   `json:"is_valid"`
	ErrorMessage string                 `json:"error_message"`
	Properties   map[string]interface{} `json:"properties" gorm:"serializer:json"`
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// PerformanceMetric 性能指标
type PerformanceMetric struct {
	ID           string    `json:"id" gorm:"primaryKey;size:64"`
	SessionID    string    `json:"session_id" gorm:"size:64;not null;index"`
	UserID       string    `json:"user_id" gorm:"size:64;not null;index"`
	URL          string    `json:"url" gorm:"size:2000;not null;index"`
	MetricType   string    `json:"metric_type" gorm:"size:50;not null;index"` // load_time, dom_complete, first_contentful_paint
	MetricValue  float64   `json:"metric_value" gorm:"not null"`
	Unit         string    `json:"unit" gorm:"size:20"`          // ms, percentage
	ResourceType string    `json:"resource_type" gorm:"size:50"` // html, css, js, image
	DeviceType   string    `json:"device_type" gorm:"size:50"`
	Browser      string    `json:"browser" gorm:"size:100"`
	Timestamp    time.Time `json:"timestamp" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SearchEvent 搜索事件
type SearchEvent struct {
	ID            string                 `json:"id" gorm:"primaryKey;size:64"`
	SessionID     string                 `json:"session_id" gorm:"size:64;not null;index"`
	UserID        string                 `json:"user_id" gorm:"size:64;not null;index"`
	SearchQuery   string                 `json:"search_query" gorm:"size:200;not null;index"`
	SearchType    string                 `json:"search_type" gorm:"size:50"` // site_search, page_search, form_search`
	ResultsCount  int                    `json:"results_count"`
	ClickedResult int                    `json:"clicked_result"`
	ClickPosition int                    `json:"click_position"` // 点击结果的序号
	URL           string                 `json:"url" gorm:"size:2000;not null;index"`
	Timestamp     time.Time              `json:"timestamp" gorm:"not null;index"`
	Properties    map[string]interface{} `json:"properties" gorm:"serializer:json"`
	CreatedAt     time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// ConversionEvent 转化事件
type ConversionEvent struct {
	ID            string                 `json:"id" gorm:"primaryKey;size:64"`
	SessionID     string                 `json:"session_id" gorm:"size:64;not null;index"`
	UserID        string                 `json:"user_id" gorm:"size:64;not null;index"`
	EventType     string                 `json:"event_type" gorm:"size:50;not null;index"` // signup, login, purchase, download
	EventCategory string                 `json:"event_category" gorm:"size:50;not null;index"`
	EventAction   string                 `json:"event_action" gorm:"size:50;not null;index"`
	EventLabel    string                 `json:"event_label" gorm:"size:100"`
	EventValue    float64                `json:"event_value"` // 转换值
	Currency      string                 `json:"currency" gorm:"size:10"`
	URL           string                 `json:"url" gorm:"size:2000;not null;index"`
	Referrer      string                 `json:"referrer" gorm:"size:500"`
	Source        string                 `json:"source" gorm:"size:100"`
	Campaign      string                 `json:"campaign" gorm:"size:100"`
	Properties    map[string]interface{} `json:"properties" gorm:"serializer:json"`
	Timestamp     time.Time              `json:"timestamp" gorm:"not null;index"`
	CreatedAt     time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// UserSessionSummary 用户会话摘要统计
type UserSessionSummary struct {
	TotalSessions   int64     `json:"total_sessions"`
	ActiveSessions  int64     `json:"active_sessions"`
	AvgDuration     int64     `json:"avg_duration"`
	MaxDuration     int64     `json:"max_duration"`
	FirstSession    time.Time `json:"first_session"`
	LastSession     time.Time `json:"last_session"`
	TotalPageViews  int64     `json:"total_page_views"`
	UniquePages     int64     `json:"unique_pages"`
	AvgPageDuration float64   `json:"avg_page_duration"`
	TotalEvents     int64     `json:"total_events"`
	ClickEvents     int64     `json:"click_events"`
	FormEvents      int64     `json:"form_events"`
}

// RealTimeStats 实时统计数据
type RealTimeStats struct {
	ID         string                 `json:"id" gorm:"primaryKey;size:64"`
	MetricName string                 `json:"metric_name" gorm:"size:100;not null;index"` // active_users, page_views, events_per_minute
	Value      float64                `json:"value"`                                      // 指标值
	Dimensions map[string]interface{} `json:"dimensions" gorm:"serializer:json"`          // 维度信息
	Timestamp  time.Time              `json:"timestamp" gorm:"not null;index"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// 表名映射
func (AnalyticsUserSession) TableName() string  { return "user_sessions" }
func (PageView) TableName() string              { return "page_views" }
func (AnalyticsUserEvent) TableName() string    { return "user_events" }
func (UserJourney) TableName() string           { return "user_journeys" }
func (UserSegment) TableName() string           { return "user_segments" }
func (UserSegmentMembership) TableName() string { return "user_segment_memberships" }
func (BehaviorPattern) TableName() string       { return "behavior_patterns" }
func (UserBehaviorRecord) TableName() string    { return "user_behavior_records" }
func (FunnelAnalysis) TableName() string        { return "funnel_analyses" }
func (RetentionAnalysis) TableName() string     { return "retention_analyses" }
func (AnalyticsReportData) TableName() string   { return "analytics_reports" }
func (HeatmapData) TableName() string           { return "heatmap_data" }
func (ClickstreamRecord) TableName() string     { return "clickstream_records" }
func (FormInteraction) TableName() string       { return "form_interactions" }
func (PerformanceMetric) TableName() string     { return "performance_metrics" }
func (SearchEvent) TableName() string           { return "search_events" }
func (ConversionEvent) TableName() string       { return "conversion_events" }
func (RealTimeStats) TableName() string         { return "real_time_stats" }
