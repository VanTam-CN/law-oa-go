package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TestSuite 测试套件模型
type TestSuite struct {
	ID           string         `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	Name         string         `json:"name" gorm:"size:255;not null"`
	Type         TestType       `json:"type" gorm:"type:varchar(20);not null;default:'api'"`
	Description  string         `json:"description" gorm:"type:text"`
	Config       *TestConfig    `json:"config" gorm:"serializer:json"`
	CreatedBy    string         `json:"created_by" gorm:"size:36"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	ScheduleCron string         `json:"schedule_cron" gorm:"size:100"`
	Environment  string         `json:"environment" gorm:"size:50;default:'test'"`
}

// TestType 测试类型枚举
type TestType string

const (
	TestTypeAPI         TestType = "api"
	TestTypeUI          TestType = "ui"
	TestTypePerformance TestType = "performance"
	TestTypeIntegration TestType = "integration"
)

// Scan implements sql.Scanner interface
func (t *TestType) Scan(value interface{}) error {
	if value == nil {
		*t = TestTypeAPI
		return nil
	}
	switch v := value.(type) {
	case string:
		*t = TestType(v)
	case []byte:
		*t = TestType(string(v))
	default:
		*t = TestTypeAPI
	}
	return nil
}

// TestConfig 测试配置
type TestConfig struct {
	Timeout   int                    `json:"timeout,omitempty"`   // 超时时间（秒）
	Variables map[string]interface{} `json:"variables,omitempty"` // 环境变量
	Headers   map[string]string      `json:"headers,omitempty"`   // HTTP头
	Setup     []TestStep             `json:"setup,omitempty"`     // 设置步骤
	Teardown  []TestStep             `json:"teardown,omitempty"`  // 清理步骤
	Parallel  bool                   `json:"parallel,omitempty"`  // 是否并行执行
	Retries   int                    `json:"retries,omitempty"`   // 重试次数
}

// TestStep 测试步骤
type TestStep struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Action   string                 `json:"action"`
	Target   string                 `json:"target"`
	Value    interface{}            `json:"value,omitempty"`
	Expected interface{}            `json:"expected,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TestExecution 测试执行记录模型
type TestExecution struct {
	ID           string              `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt    time.Time           `json:"created_at"`
	SuiteID      string              `json:"suite_id" gorm:"type:varchar(36);not null;index"`
	Status       TestExecutionStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	StartedAt    *time.Time          `json:"started_at"`
	CompletedAt  *time.Time          `json:"completed_at"`
	DurationMs   int                 `json:"duration_ms" gorm:"default:0"`
	Environment  string              `json:"environment" gorm:"size:50"`
	TriggerType  TriggerType         `json:"trigger_type" gorm:"type:varchar(20);default:'manual'"`
	TriggeredBy  string              `json:"triggered_by" gorm:"size:36"`
	Result       *TestResult         `json:"result" gorm:"serializer:json"`
	ErrorMessage string              `json:"error_message" gorm:"type:text"`
	TestSuite    *TestSuite          `json:"test_suite,omitempty" gorm:"foreignKey:SuiteID"`
	TestResults  []TestResult        `json:"test_results,omitempty" gorm:"foreignKey:ExecutionID"`
}

// TestExecutionStatus 测试执行状态
type TestExecutionStatus string

const (
	TestStatusPending   TestExecutionStatus = "pending"
	TestStatusRunning   TestExecutionStatus = "running"
	TestStatusCompleted TestExecutionStatus = "completed"
	TestStatusFailed    TestExecutionStatus = "failed"
	TestStatusCancelled TestExecutionStatus = "cancelled"
)

// TriggerType 触发类型
type TriggerType string

const (
	TriggerManual    TriggerType = "manual"
	TriggerScheduled TriggerType = "scheduled"
	TriggerAPI       TriggerType = "api"
)

// TestResult 测试结果模型
type TestResult struct {
	ID               string                 `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt        time.Time              `json:"created_at"`
	ExecutionID      string                 `json:"execution_id" gorm:"type:varchar(36);not null;index"`
	TestName         string                 `json:"test_name" gorm:"size:255;not null"`
	TestType         string                 `json:"test_type" gorm:"size:50;not null"`
	Status           TestResultStatus       `json:"status" gorm:"type:varchar(20);not null"`
	DurationMs       int                    `json:"duration_ms" gorm:"default:0"`
	ErrorMessage     string                 `json:"error_message" gorm:"type:text"`
	AssertionResults []Assertion            `json:"assertion_results" gorm:"serializer:json"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	TestExecution    *TestExecution         `json:"test_execution,omitempty" gorm:"foreignKey:ExecutionID"`
}

// TestResultStatus 测试结果状态
type TestResultStatus string

const (
	TestResultPassed  TestResultStatus = "passed"
	TestResultFailed  TestResultStatus = "failed"
	TestResultSkipped TestResultStatus = "skipped"
	TestResultError   TestResultStatus = "error"
)

// Assertion 断言结果
type Assertion struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Expected     interface{} `json:"expected"`
	Actual       interface{} `json:"actual"`
	Passed       bool        `json:"passed"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

// UserEvent 用户事件模型
type UserEvent struct {
	ID              string                 `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt       time.Time              `json:"created_at"`
	UserID          string                 `json:"user_id" gorm:"size:36;index"`
	SessionID       string                 `json:"session_id" gorm:"size:36;not null;index"`
	EventType       string                 `json:"event_type" gorm:"size:100;not null;index"`
	Element         string                 `json:"element" gorm:"size:255"`
	ElementSelector string                 `json:"element_selector" gorm:"size:500"`
	PageURL         string                 `json:"page_url" gorm:"size:500;index"`
	PageTitle       string                 `json:"page_title" gorm:"size:255"`
	Referrer        string                 `json:"referrer" gorm:"size:500"`
	UserAgent       string                 `json:"user_agent" gorm:"type:text"`
	IPAddress       string                 `json:"ip_address" gorm:"size:45"`
	Timestamp       time.Time              `json:"timestamp" gorm:"not null;index"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"serializer:json"`
}

// UserSession 用户会话模型
type UserSession struct {
	ID               string     `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt        time.Time  `json:"created_at"`
	UserID           string     `json:"user_id" gorm:"size:36;index"`
	StartedAt        time.Time  `json:"started_at" gorm:"not null;index"`
	EndedAt          *time.Time `json:"ended_at"`
	DurationMs       int        `json:"duration_ms" gorm:"default:0"`
	PageViews        int        `json:"page_views" gorm:"default:0"`
	EventsCount      int        `json:"events_count" gorm:"default:0"`
	Browser          string     `json:"browser" gorm:"size:100"`
	BrowserVersion   string     `json:"browser_version" gorm:"size:50"`
	OS               string     `json:"os" gorm:"size:100"`
	OSVersion        string     `json:"os_version" gorm:"size:50"`
	DeviceType       string     `json:"device_type" gorm:"size:50;index"`
	ScreenResolution string     `json:"screen_resolution" gorm:"size:20"`
	IPAddress        string     `json:"ip_address" gorm:"size:45;index"`
	UserAgent        string     `json:"user_agent" gorm:"type:text"`
}

// AnalyticsReport 分析报告模型
type AnalyticsReport struct {
	ID          string       `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt   time.Time    `json:"created_at"`
	ReportType  ReportType   `json:"report_type" gorm:"type:varchar(30);not null;index"`
	Title       string       `json:"title" gorm:"size:255;not null"`
	Description string       `json:"description" gorm:"type:text"`
	PeriodStart *time.Time   `json:"period_start"`
	PeriodEnd   *time.Time   `json:"period_end"`
	Data        interface{}  `json:"data" gorm:"serializer:json;not null"`
	FilePath    string       `json:"file_path" gorm:"size:500"`
	Status      ReportStatus `json:"status" gorm:"type:varchar(20);default:'generating';index"`
	GeneratedBy string       `json:"generated_by" gorm:"size:36;index"`
	GeneratedAt *time.Time   `json:"generated_at"`
}

// ReportType 报告类型
type ReportType string

const (
	ReportTypeUserBehavior      ReportType = "user_behavior"
	ReportTypeSystemPerformance ReportType = "system_performance"
	ReportTypeTestSummary       ReportType = "test_summary"
	ReportTypeCustom            ReportType = "custom"
)

// ReportStatus 报告状态
type ReportStatus string

const (
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
)

// SystemMetric 系统指标模型
type SystemMetric struct {
	ID         string            `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt  time.Time         `json:"created_at"`
	MetricName string            `json:"metric_name" gorm:"size:100;not null;index"`
	MetricType MetricType        `json:"metric_type" gorm:"type:varchar(20);not null;index"`
	Value      float64           `json:"value" gorm:"type:decimal(15,4);not null"`
	Labels     map[string]string `json:"labels" gorm:"serializer:json"`
	Timestamp  time.Time         `json:"timestamp" gorm:"not null;index"`
}

// MetricType 指标类型
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Alert 预警模型
type Alert struct {
	ID             string                 `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt      time.Time              `json:"created_at"`
	Title          string                 `json:"title" gorm:"size:255;not null"`
	Description    string                 `json:"description" gorm:"type:text"`
	Severity       AlertSeverity          `json:"severity" gorm:"type:varchar(20);not null;index"`
	Source         string                 `json:"source" gorm:"size:100;not null;index"`
	Status         AlertStatus            `json:"status" gorm:"type:varchar(20);default:'active';index"`
	RuleName       string                 `json:"rule_name" gorm:"size:255"`
	RuleID         string                 `json:"rule_id" gorm:"size:36;index"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	TriggeredAt    time.Time              `json:"triggered_at" gorm:"not null;index"`
	AcknowledgedAt *time.Time             `json:"acknowledged_at"`
	AcknowledgedBy string                 `json:"acknowledged_by" gorm:"size:36"`
	ResolvedAt     *time.Time             `json:"resolved_at"`
	ResolvedBy     string                 `json:"resolved_by" gorm:"size:36"`
	Notifications  []AlertNotification    `json:"notifications,omitempty" gorm:"foreignKey:AlertID"`
}

// AlertSeverity 预警严重程度
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus 预警状态
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

// AlertNotification 预警通知模型
type AlertNotification struct {
	ID           string             `json:"id" gorm:"type:varchar(36);primaryKey"`
	CreatedAt    time.Time          `json:"created_at"`
	AlertID      string             `json:"alert_id" gorm:"type:varchar(36);not null;index"`
	ChannelType  NotificationType   `json:"channel_type" gorm:"type:varchar(20);not null;index"`
	Recipient    string             `json:"recipient" gorm:"size:255;not null"`
	Status       NotificationStatus `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	SentAt       *time.Time         `json:"sent_at"`
	ErrorMessage string             `json:"error_message" gorm:"type:text"`
	RetryCount   int                `json:"retry_count" gorm:"default:0"`
	Alert        *Alert             `json:"alert,omitempty" gorm:"foreignKey:AlertID"`
}

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeWebhook NotificationType = "webhook"
	NotificationTypeSMS     NotificationType = "sms"
	NotificationTypeInApp   NotificationType = "in_app"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

func ensureTestingModelID(id *string) {
	if *id == "" {
		*id = uuid.NewString()
	}
}

func (s *TestSuite) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&s.ID)
	return nil
}

func (e *TestExecution) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&e.ID)
	return nil
}

func (r *TestResult) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&r.ID)
	return nil
}

func (e *UserEvent) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&e.ID)
	return nil
}

func (s *UserSession) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&s.ID)
	return nil
}

func (r *AnalyticsReport) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&r.ID)
	return nil
}

func (m *SystemMetric) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&m.ID)
	return nil
}

func (a *Alert) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&a.ID)
	return nil
}

func (n *AlertNotification) BeforeCreate(tx *gorm.DB) error {
	ensureTestingModelID(&n.ID)
	return nil
}
