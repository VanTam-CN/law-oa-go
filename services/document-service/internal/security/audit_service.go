package security

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuditEvent 审计事件结构（遵循ISO 27001和eIDAS标准）
type AuditEvent struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	Level       AuditLevel        `json:"level"`
	Category    AuditCategory     `json:"category"`
	Action      string            `json:"action"`
	Resource    string            `json:"resource"`
	UserID      string            `json:"user_id"`
	SessionID   string            `json:"session_id"`
	RequestID   string            `json:"request_id"`
	IPAddress   string            `json:"ip_address"`
	UserAgent   string            `json:"user_agent"`
	Result      AuditResult       `json:"result"`
	Message     string            `json:"message"`
	Details     map[string]string `json:"details"`
	Hash        string            `json:"hash"`
	Signature   string            `json:"signature"`
	PreviousHash string           `json:"previous_hash"`
}

// AuditLevel 审计级别
type AuditLevel string

const (
	AuditLevelInfo     AuditLevel = "INFO"
	AuditLevelWarn     AuditLevel = "WARN"
	AuditLevelError    AuditLevel = "ERROR"
	AuditLevelCritical AuditLevel = "CRITICAL"
	AuditLevelDebug    AuditLevel = "DEBUG"
)

// AuditCategory 审计分类
type AuditCategory string

const (
	CategoryAuthentication  AuditCategory = "AUTHENTICATION"
	CategoryAuthorization   AuditCategory = "AUTHORIZATION"
	CategoryDataAccess     AuditCategory = "DATA_ACCESS"
	CategorySystemOperation AuditCategory = "SYSTEM_OPERATION"
	CategorySecurityEvent   AuditCategory = "SECURITY_EVENT"
	CategoryCompliance      AuditCategory = "COMPLIANCE"
	CategorySignature      AuditCategory = "SIGNATURE"
	CategoryDocument        AuditCategory = "DOCUMENT"
	CategoryUserManagement  AuditCategory = "USER_MANAGEMENT"
)

// AuditResult 审计结果
type AuditResult string

const (
	ResultSuccess AuditResult = "SUCCESS"
	ResultFailure AuditResult = "FAILURE"
	ResultError   AuditResult = "ERROR"
	ResultPartial AuditResult = "PARTIAL"
)

// ComplianceRule 合规规则
type ComplianceRule struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Category    string                   `json:"category"`
	Enabled     bool                     `json:"enabled"`
	Severity    ComplianceSeverity       `json:"severity"`
	Conditions  []ComplianceCondition     `json:"conditions"`
	Actions     []ComplianceAction        `json:"actions"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// ComplianceCondition 合规条件
type ComplianceCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"` // eq, ne, gt, lt, contains, regex
	Value     interface{} `json:"value"`
	LogicalOp string      `json:"logical_op"` // and, or
}

// ComplianceSeverity 合规严重程度
type ComplianceSeverity string

const (
	SeverityLow      ComplianceSeverity = "LOW"
	SeverityMedium   ComplianceSeverity = "MEDIUM"
	SeverityHigh     ComplianceSeverity = "HIGH"
	SeverityCritical ComplianceSeverity = "CRITICAL"
)

// ComplianceAction 合规动作
type ComplianceAction struct {
	Type        string      `json:"type"` // alert, block, notify, log
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool        `json:"enabled"`
}

// ComplianceCheck 合规检查结果
type ComplianceCheck struct {
	RuleID      string            `json:"rule_id"`
	RuleName    string            `json:"rule_name"`
	Passed      bool              `json:"passed"`
	Severity    ComplianceSeverity `json:"severity"`
	Message     string            `json:"message"`
	CheckedAt   time.Time         `json:"checked_at"`
	Details     map[string]string `json:"details"`
}

// AuditReport 审计报告
type AuditReport struct {
	ID           string            `json:"id"`
	GeneratedAt  time.Time         `json:"generated_at"`
	Period       ReportPeriod      `json:"period"`
	Summary      ReportSummary     `json:"summary"`
	Events       []*AuditEvent     `json:"events"`
	Compliance   []*ComplianceCheck `json:"compliance"`
	Statistics   ReportStatistics  `json:"statistics"`
	Hash         string            `json:"hash"`
	Signature    string            `json:"signature"`
}

// ReportPeriod 报告周期
type ReportPeriod struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Type      string    `json:"type"` // daily, weekly, monthly, yearly
}

// ReportSummary 报告摘要
type ReportSummary struct {
	TotalEvents      int               `json:"total_events"`
	EventsByLevel    map[AuditLevel]int `json:"events_by_level"`
	EventsByCategory map[AuditCategory]int `json:"events_by_category"`
	SecurityEvents   int               `json:"security_events"`
	ComplianceScore  float64           `json:"compliance_score"`
	RiskLevel        string            `json:"risk_level"`
}

// ReportStatistics 报告统计
type ReportStatistics struct {
	UserActivity       map[string]int `json:"user_activity"`
	ResourceAccess     map[string]int `json:"resource_access"`
	TrafficPatterns    map[string]int `json:"traffic_patterns"`
	ErrorRate          float64       `json:"error_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TopActions         []ActionStat  `json:"top_actions"`
}

// ActionStat 动作统计
type ActionStat struct {
	Action  string `json:"action"`
	Count   int    `json:"count"`
	Percent float64 `json:"percent"`
}

// AuditStorage 审计存储接口
type AuditStorage interface {
	StoreEvent(event *AuditEvent) error
	StoreReport(report *AuditReport) error
	GetEvents(filter *EventFilter) ([]*AuditEvent, error)
	GetReports(filter *ReportFilter) ([]*AuditReport, error)
	DeleteEventsBefore(t time.Time) error
	GetEventCount(filter *EventFilter) (int64, error)
}

// EventFilter 事件过滤器
type EventFilter struct {
	StartTime   *time.Time       `json:"start_time,omitempty"`
	EndTime     *time.Time       `json:"end_time,omitempty"`
	Level       *AuditLevel      `json:"level,omitempty"`
	Category    *AuditCategory   `json:"category,omitempty"`
	UserID      *string          `json:"user_id,omitempty"`
	Action      *string          `json:"action,omitempty"`
	Resource    *string          `json:"resource,omitempty"`
	SessionID   *string          `json:"session_id,omitempty"`
	RequestID   *string          `json:"request_id,omitempty"`
	IPAddress   *string          `json:"ip_address,omitempty"`
	Limit       *int             `json:"limit,omitempty"`
	Offset      *int             `json:"offset,omitempty"`
}

// ReportFilter 报告过滤器
type ReportFilter struct {
	StartTime *time.Time  `json:"start_time,omitempty"`
	EndTime   *time.Time  `json:"end_time,omitempty"`
	Type      *string     `json:"type,omitempty"`
	UserID    *string     `json:"user_id,omitempty"`
	Limit     *int        `json:"limit,omitempty"`
	Offset    *int        `json:"offset,omitempty"`
}

// AuditConfiguration 审计配置
type AuditConfiguration struct {
	Enabled              bool               `json:"enabled"`
	LogLevel             AuditLevel         `json:"log_level"`
	MaxEventsPerSecond   int                `json:"max_events_per_second"`
	RetentionPeriod      time.Duration      `json:"retention_period"`
	EncryptionEnabled    bool               `json:"encryption_enabled"`
	CompressionEnabled   bool               `json:"compression_enabled"`
	StorageBackend       string             `json:"storage_backend"` // memory, file, database
	ComplianceEnabled    bool               `json:"compliance_enabled"`
	RealTimeMonitoring   bool               `json:"real_time_monitoring"`
	AlertThresholds      map[string]int     `json:"alert_thresholds"`
	RequiredFields       []string           `json:"required_fields"`
	SensitiveDataFields  []string           `json:"sensitive_data_fields"`
	DataRetentionPolicies map[string]time.Duration `json:"data_retention_policies"`
}

// AuditService 审计服务
type AuditService struct {
	config         *AuditConfiguration
	storage        AuditStorage
	logger         *slog.Logger
	complianceRules map[string]*ComplianceRule
	signingKey     []byte
	previousHash   string
	mutex          sync.RWMutex
	eventChan      chan *AuditEvent
	reportChan     chan *AuditReport
	shutdownChan   chan struct{}
	wg             sync.WaitGroup
}

// NewAuditService 创建审计服务
func NewAuditService(config *AuditConfiguration, storage AuditStorage, logger *slog.Logger) (*AuditService, error) {
	if config == nil {
		config = &AuditConfiguration{
			Enabled:            true,
			LogLevel:           AuditLevelInfo,
			MaxEventsPerSecond: 10000,
			RetentionPeriod:    7 * 365 * 24 * time.Hour, // 7年
			EncryptionEnabled:  true,
			CompressionEnabled: true,
			StorageBackend:     "memory",
			ComplianceEnabled:  true,
			RealTimeMonitoring: true,
			AlertThresholds: map[string]int{
				"error_rate":        10,
				"failed_login":      5,
				"unauthorized_access": 3,
			},
			RequiredFields: []string{
				"timestamp", "level", "category", "action", "user_id", "result",
			},
			SensitiveDataFields: []string{
				"password", "ssn", "credit_card", "api_key", "token",
			},
			DataRetentionPolicies: map[string]time.Duration{
				"authentication_events": 7 * 365 * 24 * time.Hour,
				"security_events":     10 * 365 * 24 * time.Hour,
				"system_events":       2 * 365 * 24 * time.Hour,
				"error_events":         6 * 30 * 24 * time.Hour,
			},
		}
	}

	// 生成签名密钥
	signingKey := make([]byte, 32)
	if _, err := rand.Read(signingKey); err != nil {
		return nil, fmt.Errorf("生成签名密钥失败: %w", err)
	}

	service := &AuditService{
		config:          config,
		storage:         storage,
		logger:          logger,
		complianceRules: make(map[string]*ComplianceRule),
		signingKey:      signingKey,
		eventChan:       make(chan *AuditEvent, 10000),
		reportChan:      make(chan *AuditReport, 100),
		shutdownChan:    make(chan struct{}),
	}

	// 初始化默认合规规则
	if err := service.initializeDefaultRules(); err != nil {
		return nil, fmt.Errorf("初始化合规规则失败: %w", err)
	}

	// 启动后台处理协程
	if config.Enabled {
		service.wg.Add(2)
		go service.eventProcessor()
		go service.reportGenerator()
	}

	logger.Info("审计服务创建成功",
		"enabled", config.Enabled,
		"log_level", config.LogLevel,
		"retention_period", config.RetentionPeriod,
		"compliance_enabled", config.ComplianceEnabled,
		"real_time_monitoring", config.RealTimeMonitoring,
	)

	return service, nil
}

// initializeDefaultRules 初始化默认合规规则
func (as *AuditService) initializeDefaultRules() error {
	defaultRules := []*ComplianceRule{
		// GDPR合规规则
		{
			ID:          "gdpr_data_access",
			Name:        "GDPR数据访问合规",
			Description: "确保数据访问符合GDPR要求",
			Category:    "GDPR",
			Enabled:     true,
			Severity:    SeverityHigh,
			Conditions: []ComplianceCondition{
				{
					Field:    "category",
					Operator: "eq",
					Value:    CategoryDataAccess,
				},
				{
					Field:    "result",
					Operator: "eq",
					Value:    ResultSuccess,
				},
			},
			Actions: []ComplianceAction{
				{
					Type:   "log",
					Parameters: map[string]interface{}{
						"message": "GDPR合规数据访问记录",
					},
					Enabled: true,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// SOX合规规则
		{
			ID:          "sox_financial_operations",
			Name:        "SOX财务操作合规",
			Description: "确保财务操作符合SOX要求",
			Category:    "SOX",
			Enabled:     true,
			Severity:    SeverityCritical,
			Conditions: []ComplianceCondition{
				{
					Field:    "category",
					Operator: "eq",
					Value:    CategorySystemOperation,
				},
				{
					Field:    "action",
					Operator: "contains",
					Value:    "financial",
				},
			},
			Actions: []ComplianceAction{
				{
					Type:   "alert",
					Parameters: map[string]interface{}{
						"recipients": []string{"compliance@lawfirm.com"},
						"subject":   "SOX财务操作合规警报",
					},
					Enabled: true,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// 安全事件规则
		{
			ID:          "security_event_monitoring",
			Name:        "安全事件监控",
			Description: "监控和报告安全事件",
			Category:    "SECURITY",
			Enabled:     true,
			Severity:    SeverityMedium,
			Conditions: []ComplianceCondition{
				{
					Field:    "level",
					Operator: "eq",
					Value:    AuditLevelError,
				},
			},
			Actions: []ComplianceAction{
				{
					Type:   "notify",
					Parameters: map[string]interface{}{
						"channels": []string{"email", "slack"},
					},
					Enabled: true,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		// 数字签名规则
		{
			ID:          "digital_signature_compliance",
			Name:        "数字签名合规检查",
			Description: "确保数字签名操作符合合规要求",
			Category:    "SIGNATURE",
			Enabled:     true,
			Severity:    SeverityHigh,
			Conditions: []ComplianceCondition{
				{
					Field:    "category",
					Operator: "eq",
					Value:    CategorySignature,
				},
			},
			Actions: []ComplianceAction{
				{
					Type:   "log",
					Parameters: map[string]interface{}{
						"message": "数字签名合规检查通过",
					},
					Enabled: true,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, rule := range defaultRules {
		as.complianceRules[rule.ID] = rule
	}

	as.logger.Info("默认合规规则初始化完成", "count", len(defaultRules))
	return nil
}

// LogEvent 记录审计事件
func (as *AuditService) LogEvent(ctx context.Context, event *AuditEvent) error {
	if !as.config.Enabled {
		return nil
	}

	// 生成事件ID
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// 设置时间戳
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// 确保必填字段存在
	if err := as.validateEvent(event); err != nil {
		return fmt.Errorf("事件验证失败: %w", err)
	}

	// 敏感数据脱敏
	if err := as.sanitizeEvent(event); err != nil {
		return fmt.Errorf("事件脱敏失败: %w", err)
	}

	// 计算事件哈希
	event.Hash = as.calculateEventHash(event)

	// 数字签名事件
	if as.config.EncryptionEnabled {
		signature, err := as.signEvent(event)
		if err != nil {
			return fmt.Errorf("事件签名失败: %w", err)
		}
		event.Signature = signature
	}

	// 设置前一个哈希（用于构建区块链）
	event.PreviousHash = as.previousHash
	as.previousHash = event.Hash

	// 检查合规性
	if as.config.ComplianceEnabled {
		go as.checkCompliance(event)
	}

	// 异步处理事件
	select {
	case as.eventChan <- event:
	default:
		// 事件队列满了，记录警告
		as.logger.Warn("审计事件队列已满，丢弃事件", "event_id", event.ID)
		return fmt.Errorf("审计事件队列已满")
	}

	return nil
}

// validateEvent 验证事件
func (as *AuditService) validateEvent(event *AuditEvent) error {
	// 检查必填字段
	for _, field := range as.config.RequiredFields {
		switch field {
		case "timestamp":
			if event.Timestamp.IsZero() {
				return fmt.Errorf("缺少必填字段: %s", field)
			}
		case "level":
			if event.Level == "" {
				return fmt.Errorf("缺少必填字段: %s", field)
			}
		case "category":
			if event.Category == "" {
				return fmt.Errorf("缺少必填字段: %s", field)
			}
		case "action":
			if event.Action == "" {
				return fmt.Errorf("缺少必填字段: %s", field)
			}
		case "user_id":
			if event.UserID == "" {
				return fmt.Errorf("缺少必填字段: %s", field)
			}
		case "result":
			if event.Result == "" {
				return fmt.Errorf("缺少必填字段: %s", field)
			}
		}
	}

	// 检查事件级别
	if !isValidLevel(event.Level) {
		return fmt.Errorf("无效的事件级别: %s", event.Level)
	}

	// 检查事件分类
	if !isValidCategory(event.Category) {
		return fmt.Errorf("无效的事件分类: %s", event.Category)
	}

	// 检查事件结果
	if !isValidResult(event.Result) {
		return fmt.Errorf("无效的事件结果: %s", event.Result)
	}

	return nil
}

// sanitizeEvent 脱敏事件
func (as *AuditService) sanitizeEvent(event *AuditEvent) error {
	if event.Details == nil {
		event.Details = make(map[string]string)
	}

	// 脱敏敏感数据字段
	for _, field := range as.config.SensitiveDataFields {
		if value, exists := event.Details[field]; exists {
			event.Details[field] = as.maskSensitiveData(value)
		}
	}

	return nil
}

// maskSensitiveData 掩码敏感数据
func (as *AuditService) maskSensitiveData(data string) string {
	if len(data) <= 4 {
		return "****"
	}
	return data[:2] + "****" + data[len(data)-2:]
}

// calculateEventHash 计算事件哈希
func (as *AuditService) calculateEventHash(event *AuditEvent) string {
	// 序列化事件（排除哈希和签名字段）
	eventCopy := *event
	eventCopy.Hash = ""
	eventCopy.Signature = ""
	eventCopy.PreviousHash = ""

	data, err := json.Marshal(eventCopy)
	if err != nil {
		as.logger.Error("序列化事件失败", "error", err)
		return ""
	}

	// 使用SHA-256计算哈希
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// signEvent 签名事件
func (as *AuditService) signEvent(event *AuditEvent) (string, error) {
	data := []byte(event.Hash)

	// 简化签名实现（实际应用中应使用更强的签名算法）
	signature := hmac.New(sha256.New, as.signingKey)
	signature.Write(data)
	signatureBytes := signature.Sum(nil)

	return hex.EncodeToString(signatureBytes), nil
}

// checkCompliance 检查合规性
func (as *AuditService) checkCompliance(event *AuditEvent) {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	for _, rule := range as.complianceRules {
		if !rule.Enabled {
			continue
		}

		passed := as.evaluateConditions(rule.Conditions, event)

		check := &ComplianceCheck{
			RuleID:    rule.ID,
			RuleName:  rule.Name,
			Passed:    passed,
			Severity:  rule.Severity,
			CheckedAt: time.Now(),
			Details:   map[string]string{},
		}

		if !passed {
			check.Message = "合规检查失败"
			as.executeComplianceActions(rule.Actions, event, check)
		} else {
			check.Message = "合规检查通过"
		}

		// 记录合规检查结果
		as.logger.Info("合规检查完成",
			"rule_id", check.RuleID,
			"rule_name", check.RuleName,
			"passed", check.Passed,
			"severity", check.Severity,
			"message", check.Message,
		)
	}
}

// evaluateConditions 评估条件
func (as *AuditService) evaluateConditions(conditions []ComplianceCondition, event *AuditEvent) bool {
	if len(conditions) == 0 {
		return true
	}

	result := true
	currentLogicalOp := "and"

	for _, condition := range conditions {
		conditionResult := as.evaluateCondition(condition, event)

		switch currentLogicalOp {
		case "and":
			result = result && conditionResult
		case "or":
			result = result || conditionResult
		}

		// 下一个条件的逻辑操作符
		if condition.LogicalOp != "" {
			currentLogicalOp = condition.LogicalOp
		}
	}

	return result
}

// evaluateCondition 评估单个条件
func (as *AuditService) evaluateCondition(condition ComplianceCondition, event *AuditEvent) bool {
	var fieldValue interface{}

	switch condition.Field {
	case "level":
		fieldValue = event.Level
	case "category":
		fieldValue = event.Category
	case "action":
		fieldValue = event.Action
	case "resource":
		fieldValue = event.Resource
	case "result":
		fieldValue = event.Result
	case "user_id":
		fieldValue = event.UserID
	default:
		// 从details中获取字段值
		if event.Details != nil {
			fieldValue = event.Details[condition.Field]
		}
	}

	return as.compareValues(fieldValue, condition.Operator, condition.Value)
}

// compareValues 比较值
func (as *AuditService) compareValues(fieldValue interface{}, operator string, conditionValue interface{}) bool {
	switch operator {
	case "eq":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", conditionValue)
	case "ne":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", conditionValue)
	case "contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		conditionStr := fmt.Sprintf("%v", conditionValue)
		return fieldStr != "" && len(fieldStr) >= len(conditionStr) &&
			   fieldStr[:len(conditionStr)] == conditionStr
	case "regex":
		// 简化实现，实际应用中应使用正则表达式
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", conditionValue)
	default:
		return false
	}
}

// executeComplianceActions 执行合规动作
func (as *AuditService) executeComplianceActions(actions []ComplianceAction, event *AuditEvent, check *ComplianceCheck) {
	for _, action := range actions {
		if !action.Enabled {
			continue
		}

		switch action.Type {
		case "log":
			as.logger.Info("合规动作执行",
				"action_type", action.Type,
				"rule_id", check.RuleID,
				"event_id", event.ID,
				"parameters", action.Parameters,
			)
		case "alert":
			as.logger.Warn("合规警报触发",
				"action_type", action.Type,
				"rule_id", check.RuleID,
				"event_id", event.ID,
				"severity", check.Severity,
				"parameters", action.Parameters,
			)
		case "notify":
			as.logger.Info("发送合规通知",
				"action_type", action.Type,
				"rule_id", check.RuleID,
				"event_id", event.ID,
				"parameters", action.Parameters,
			)
		case "block":
			as.logger.Error("合规阻止动作执行",
				"action_type", action.Type,
				"rule_id", check.RuleID,
				"event_id", event.ID,
				"parameters", action.Parameters,
			)
		}
	}
}

// eventProcessor 事件处理器
func (as *AuditService) eventProcessor() {
	defer as.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	batch := make([]*AuditEvent, 0, 1000)
	maxBatchSize := 100

	for {
		select {
		case event := <-as.eventChan:
			batch = append(batch, event)

			// 批量处理
			if len(batch) >= maxBatchSize {
				as.processBatch(batch)
				batch = batch[:0] // 清空批次
			}

		case <-ticker.C:
			// 定期处理批次
			if len(batch) > 0 {
				as.processBatch(batch)
				batch = batch[:0]
			}

		case <-as.shutdownChan:
			// 处理剩余批次
			if len(batch) > 0 {
				as.processBatch(batch)
			}
			return
		}
	}
}

// processBatch 批量处理事件
func (as *AuditService) processBatch(batch []*AuditEvent) {
	startTime := time.Now()

	// 存储事件
	for _, event := range batch {
		if err := as.storage.StoreEvent(event); err != nil {
			as.logger.Error("存储审计事件失败",
				"event_id", event.ID,
				"error", err,
			)
		}
	}

	duration := time.Since(startTime)
	as.logger.Debug("批量处理审计事件完成",
		"batch_size", len(batch),
		"duration", duration,
		"throughput", float64(len(batch))/duration.Seconds(),
	)
}

// reportGenerator 报告生成器
func (as *AuditService) reportGenerator() {
	defer as.wg.Done()

	// 每天生成一次报告
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := as.generateDailyReport(); err != nil {
				as.logger.Error("生成日报失败", "error", err)
			}

		case <-as.shutdownChan:
			return
		}
	}
}

// generateDailyReport 生成日报
func (as *AuditService) generateDailyReport() error {
	endTime := time.Now().UTC()
	startTime := endTime.Add(-24 * time.Hour)

	period := ReportPeriod{
		StartTime: startTime,
		EndTime:   endTime,
		Type:      "daily",
	}

	report, err := as.generateReport(period)
	if err != nil {
		return fmt.Errorf("生成报告失败: %w", err)
	}

	// 签名报告
	report.Hash = as.calculateReportHash(report)
	if as.config.EncryptionEnabled {
		signature, err := as.signReport(report)
		if err != nil {
			return fmt.Errorf("签名报告失败: %w", err)
		}
		report.Signature = signature
	}

	// 存储报告
	if err := as.storage.StoreReport(report); err != nil {
		return fmt.Errorf("存储报告失败: %w", err)
	}

	as.logger.Info("日报生成完成",
		"report_id", report.ID,
		"period_start", startTime.Format("2006-01-02"),
		"period_end", endTime.Format("2006-01-02"),
		"total_events", len(report.Events),
		"compliance_score", report.Summary.ComplianceScore,
	)

	return nil
}

// generateReport 生成报告
func (as *AuditService) generateReport(period ReportPeriod) (*AuditReport, error) {
	// 获取事件数据
	filter := &EventFilter{
		StartTime: &period.StartTime,
		EndTime:   &period.EndTime,
	}

	events, err := as.storage.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("获取事件失败: %w", err)
	}

	// 生成报告ID
	reportID := fmt.Sprintf("report_%s_%s",
		period.Type,
		period.StartTime.Format("20060102"))

	report := &AuditReport{
		ID:          reportID,
		GeneratedAt: time.Now(),
		Period:      period,
		Events:      events,
		Statistics:  as.calculateStatistics(events),
	}

	// 生成摘要
	report.Summary = as.generateSummary(events)

	// 生成合规检查结果
	report.Compliance = as.generateComplianceSummary(events)

	return report, nil
}

// generateSummary 生成摘要
func (as *AuditService) generateSummary(events []*AuditEvent) ReportSummary {
	summary := ReportSummary{
		TotalEvents:      len(events),
		EventsByLevel:    make(map[AuditLevel]int),
		EventsByCategory: make(map[AuditCategory]int),
		SecurityEvents:   0,
	}

	for _, event := range events {
		summary.EventsByLevel[event.Level]++
		summary.EventsByCategory[event.Category]++

		if event.Category == CategorySecurityEvent {
			summary.SecurityEvents++
		}
	}

	// 计算合规分数
	if summary.TotalEvents > 0 {
		criticalCount := summary.EventsByLevel[AuditLevelCritical]
		errorCount := summary.EventsByLevel[AuditLevelError]

		// 简单的合规分数计算
		baseScore := 100.0
		deduction := float64(criticalCount*5 + errorCount*2)
		summary.ComplianceScore = baseScore - deduction

		if summary.ComplianceScore < 0 {
			summary.ComplianceScore = 0
		}
	}

	// 确定风险等级
	if summary.ComplianceScore >= 90 {
		summary.RiskLevel = "LOW"
	} else if summary.ComplianceScore >= 70 {
		summary.RiskLevel = "MEDIUM"
	} else if summary.ComplianceScore >= 50 {
		summary.RiskLevel = "HIGH"
	} else {
		summary.RiskLevel = "CRITICAL"
	}

	return summary
}

// calculateStatistics 计算统计信息
func (as *AuditService) calculateStatistics(events []*AuditEvent) ReportStatistics {
	stats := ReportStatistics{
		UserActivity:    make(map[string]int),
		ResourceAccess:  make(map[string]int),
		TrafficPatterns: make(map[string]int),
		TopActions:      make([]ActionStat, 0),
	}

	errorCount := 0
	totalResponseTime := time.Duration(0)

	for _, event := range events {
		// 用户活动统计
		if event.UserID != "" {
			stats.UserActivity[event.UserID]++
		}

		// 资源访问统计
		if event.Resource != "" {
			stats.ResourceAccess[event.Resource]++
		}

		// 流量模式统计
		stats.TrafficPatterns[string(event.Action)]++

		// 错误率计算
		if event.Result == ResultError || event.Result == ResultFailure {
			errorCount++
		}

		// 响应时间（从details中获取）
		if duration, exists := event.Details["response_time"]; exists {
			if ms, err := time.ParseDuration(duration + "ms"); err == nil {
				totalResponseTime += ms
			}
		}
	}

	// 计算错误率
	if len(events) > 0 {
		stats.ErrorRate = float64(errorCount) / float64(len(events))
	}

	// 计算平均响应时间
	if len(events) > 0 {
		stats.AverageResponseTime = totalResponseTime / time.Duration(len(events))
	}

	// 生成Top动作
	for action, count := range stats.TrafficPatterns {
		percent := float64(count) / float64(len(events)) * 100
		stats.TopActions = append(stats.TopActions, ActionStat{
			Action:  action,
			Count:   count,
			Percent: percent,
		})
	}

	// 按次数排序
	for i := 0; i < len(stats.TopActions)-1; i++ {
		for j := i + 1; j < len(stats.TopActions); j++ {
			if stats.TopActions[i].Count < stats.TopActions[j].Count {
				stats.TopActions[i], stats.TopActions[j] = stats.TopActions[j], stats.TopActions[i]
			}
		}
	}

	// 只保留前10个
	if len(stats.TopActions) > 10 {
		stats.TopActions = stats.TopActions[:10]
	}

	return stats
}

// generateComplianceSummary 生成合规摘要
func (as *AuditService) generateComplianceSummary(events []*AuditEvent) []*ComplianceCheck {
	// 简化实现，实际应用中应该基于事件数据生成合规检查结果
	checks := make([]*ComplianceCheck, 0)

	// 检查数据访问合规性
	dataAccessCount := 0
	authorizedCount := 0
	for _, event := range events {
		if event.Category == CategoryDataAccess {
			dataAccessCount++
			if event.Result == ResultSuccess {
				authorizedCount++
			}
		}
	}

	if dataAccessCount > 0 {
		complianceRate := float64(authorizedCount) / float64(dataAccessCount) * 100
		checks = append(checks, &ComplianceCheck{
			RuleID:    "data_access_compliance",
			RuleName:  "数据访问合规",
			Passed:    complianceRate >= 95,
			Severity:  SeverityMedium,
			Message:   fmt.Sprintf("数据访问合规率: %.2f%%", complianceRate),
			CheckedAt: time.Now(),
			Details: map[string]string{
				"total_accesses":     fmt.Sprintf("%d", dataAccessCount),
				"authorized_accesses": fmt.Sprintf("%d", authorizedCount),
				"compliance_rate":    fmt.Sprintf("%.2f%%", complianceRate),
			},
		})
	}

	return checks
}

// calculateReportHash 计算报告哈希
func (as *AuditService) calculateReportHash(report *AuditReport) string {
	reportCopy := *report
	reportCopy.Hash = ""
	reportCopy.Signature = ""

	data, err := json.Marshal(reportCopy)
	if err != nil {
		as.logger.Error("序列化报告失败", "error", err)
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// signReport 签名报告
func (as *AuditService) signReport(report *AuditReport) (string, error) {
	data := []byte(report.Hash)

	signature := hmac.New(sha256.New, as.signingKey)
	signature.Write(data)
	signatureBytes := signature.Sum(nil)

	return hex.EncodeToString(signatureBytes), nil
}

// AddComplianceRule 添加合规规则
func (as *AuditService) AddComplianceRule(rule *ComplianceRule) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if _, exists := as.complianceRules[rule.ID]; exists {
		return fmt.Errorf("合规规则已存在: %s", rule.ID)
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	as.complianceRules[rule.ID] = rule

	as.logger.Info("添加合规规则成功",
		"rule_id", rule.ID,
		"rule_name", rule.Name,
		"category", rule.Category,
		"severity", rule.Severity,
	)

	return nil
}

// UpdateComplianceRule 更新合规规则
func (as *AuditService) UpdateComplianceRule(rule *ComplianceRule) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if _, exists := as.complianceRules[rule.ID]; !exists {
		return fmt.Errorf("合规规则不存在: %s", rule.ID)
	}

	rule.UpdatedAt = time.Now()
	as.complianceRules[rule.ID] = rule

	as.logger.Info("更新合规规则成功",
		"rule_id", rule.ID,
		"rule_name", rule.Name,
	)

	return nil
}

// DeleteComplianceRule 删除合规规则
func (as *AuditService) DeleteComplianceRule(ruleID string) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	if _, exists := as.complianceRules[ruleID]; !exists {
		return fmt.Errorf("合规规则不存在: %s", ruleID)
	}

	delete(as.complianceRules, ruleID)

	as.logger.Info("删除合规规则成功", "rule_id", ruleID)

	return nil
}

// GetComplianceRules 获取合规规则列表
func (as *AuditService) GetComplianceRules() []*ComplianceRule {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	rules := make([]*ComplianceRule, 0, len(as.complianceRules))
	for _, rule := range as.complianceRules {
		rules = append(rules, rule)
	}

	return rules
}

// GetEvents 获取审计事件
func (as *AuditService) GetEvents(filter *EventFilter) ([]*AuditEvent, error) {
	return as.storage.GetEvents(filter)
}

// GetReports 获取审计报告
func (as *AuditService) GetReports(filter *ReportFilter) ([]*AuditReport, error) {
	return as.storage.GetReports(filter)
}

// GetEventCount 获取事件数量
func (as *AuditService) GetEventCount(filter *EventFilter) (int64, error) {
	return as.storage.GetEventCount(filter)
}

// CleanupExpiredEvents 清理过期事件
func (as *AuditService) CleanupExpiredEvents() error {
	expiredTime := time.Now().Add(-as.config.RetentionPeriod)

	if err := as.storage.DeleteEventsBefore(expiredTime); err != nil {
		return fmt.Errorf("清理过期事件失败: %w", err)
	}

	as.logger.Info("清理过期事件完成", "expired_before", expiredTime)
	return nil
}

// Close 关闭审计服务
func (as *AuditService) Close() error {
	as.logger.Info("关闭审计服务")

	close(as.shutdownChan)
	as.wg.Wait()

	// 清理资源
	close(as.eventChan)
	close(as.reportChan)

	return nil
}

// 辅助函数

func isValidLevel(level AuditLevel) bool {
	switch level {
	case AuditLevelInfo, AuditLevelWarn, AuditLevelError, AuditLevelCritical, AuditLevelDebug:
		return true
	default:
		return false
	}
}

func isValidCategory(category AuditCategory) bool {
	switch category {
	case CategoryAuthentication, CategoryAuthorization, CategoryDataAccess,
		 CategorySystemOperation, CategorySecurityEvent, CategoryCompliance,
		 CategorySignature, CategoryDocument, CategoryUserManagement:
		return true
	default:
		return false
	}
}

func isValidResult(result AuditResult) bool {
	switch result {
	case ResultSuccess, ResultFailure, ResultError, ResultPartial:
		return true
	default:
		return false
	}
}

