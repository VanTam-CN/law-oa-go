package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// ComplianceStatus 合规状态
type ComplianceStatus string

const (
	StatusCompliant    ComplianceStatus = "COMPLIANT"    // 合规
	StatusNonCompliant ComplianceStatus = "NON_COMPLIANT" // 不合规
	StatusPending      ComplianceStatus = "PENDING"      // 待处理
	StatusUnknown      ComplianceStatus = "UNKNOWN"      // 未知
)

// SeverityLevel 严重级别
type SeverityLevel string

const (
	SeverityLow    SeverityLevel = "LOW"    // 低
	SeverityMedium SeverityLevel = "MEDIUM" // 中
	SeverityHigh   SeverityLevel = "HIGH"   // 高
	SeverityCritical SeverityLevel = "CRITICAL" // 严重
)

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "LOW"    // 低风险
	RiskLevelMedium RiskLevel = "MEDIUM" // 中风险
	RiskLevelHigh   RiskLevel = "HIGH"   // 高风险
	RiskLevelCritical RiskLevel = "CRITICAL" // 严重风险
)

// Violation 违规记录
type Violation struct {
	ViolationID    string                 `json:"violation_id"`
	RuleID         string                 `json:"rule_id"`
	RuleName       string                 `json:"rule_name"`
	Description    string                 `json:"description"`
	Severity       SeverityLevel          `json:"severity"`
	DetectedAt     time.Time              `json:"detected_at"`
	AffectedResource string               `json:"affected_resource"`
	Evidence       map[string]interface{} `json:"evidence"`
	Remediation    string                 `json:"remediation"`
	Status         ComplianceStatus       `json:"status"`
}

// Recommendation 建议
type Recommendation struct {
	RecommendationID string            `json:"recommendation_id"`
	Priority         int               `json:"priority"`
	Category         string            `json:"category"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	ActionItems      []string          `json:"action_items"`
	Deadline         *time.Time        `json:"deadline,omitempty"`
	AssignedTo       string            `json:"assigned_to,omitempty"`
	Status           ComplianceStatus   `json:"status"`
}

// RequiredAction 必要行动
type RequiredAction struct {
	ActionID       string                 `json:"action_id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Priority       int                    `json:"priority"`
	Deadline       time.Time              `json:"deadline"`
	AssignedTo     string                 `json:"assigned_to"`
	Status         ComplianceStatus       `json:"status"`
	Dependencies   []string               `json:"dependencies,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceCheckRequest 合规检查请求
type ComplianceCheckRequest struct {
	RequestID       string                 `json:"request_id"`
	CheckType       string                 `json:"check_type"`
	SubjectID       string                 `json:"subject_id"`
	SubjectType     string                 `json:"subject_type"`
	Data            map[string]interface{} `json:"data"`
	Context         map[string]interface{} `json:"context,omitempty"`
	Priority        int                    `json:"priority"`
	ScheduledTime   *time.Time             `json:"scheduled_time,omitempty"`
	RequestedBy     string                 `json:"requested_by"`
}

// ComplianceCheckResult 合规检查结果
type ComplianceCheckResult struct {
	RequestID          string               `json:"request_id"`
	CheckType          string               `json:"check_type"`
	SubjectID          string               `json:"subject_id"`
	SubjectType        string               `json:"subject_type"`
	OverallStatus      ComplianceStatus     `json:"overall_status"`
	OverallScore       float64              `json:"overall_score"`
	RiskLevel          RiskLevel            `json:"risk_level"`
	Violations         []Violation          `json:"violations"`
	Recommendations    []Recommendation     `json:"recommendations"`
	RequiredActions    []RequiredAction     `json:"required_actions"`
	CheckTimestamp     time.Time            `json:"check_timestamp"`
	NextReviewDate     *time.Time           `json:"next_review_date,omitempty"`
	CheckedBy          string               `json:"checked_by"`
	ProcessingTime     time.Duration        `json:"processing_time"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceRule 合规规则
type ComplianceRule struct {
	RuleID           string                 `json:"rule_id"`
	RuleName         string                 `json:"rule_name"`
	Description      string                 `json:"description"`
	Category         string                 `json:"category"`
	Version          string                 `json:"version"`
	Enabled          bool                   `json:"enabled"`
	Conditions       []RuleCondition        `json:"conditions"`
	Actions          []RuleAction           `json:"actions"`
	Priority         int                    `json:"priority"`
	Severity         SeverityLevel          `json:"severity"`
	CreationDate     time.Time              `json:"creation_date"`
	LastUpdated      time.Time              `json:"last_updated"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// RuleCondition 规则条件
type RuleCondition struct {
	Type       string        `json:"type"`        // "field", "expression", "custom"
	Field      string        `json:"field,omitempty"`
	Operator   string        `json:"operator,omitempty"`  // "==", "!=", ">", "<", "contains", "matches", "in"
	Value      interface{}   `json:"value,omitempty"`
	Expression string        `json:"expression,omitempty"`
	Conditions []RuleCondition `json:"conditions,omitempty"` // 嵌套条件
	Logic      string        `json:"logic,omitempty"`       // "AND", "OR", "NOT"
}

// RuleAction 规则动作
type RuleAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Condition  *RuleCondition         `json:"condition,omitempty"`  // 条件执行
}

// RuleExecutionResult 规则执行结果
type RuleExecutionResult struct {
	RuleID        string                 `json:"rule_id"`
	RuleName      string                 `json:"rule_name"`
	Matched       bool                   `json:"matched"`
	Executed      bool                   `json:"executed"`
	ExecutionTime time.Duration         `json:"execution_time"`
	Output        map[string]interface{} `json:"output,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ComplianceEngine 合规引擎接口
type ComplianceEngine interface {
	// 执行合规检查
	PerformCheck(ctx context.Context, req *ComplianceCheckRequest) (*ComplianceCheckResult, error)
	// 批量检查
	BatchCheck(ctx context.Context, requests []*ComplianceCheckRequest) ([]*ComplianceCheckResult, error)
	// 获取规则列表
	GetRules(ctx context.Context, category string) ([]*ComplianceRule, error)
	// 添加规则
	AddRule(ctx context.Context, rule *ComplianceRule) error
	// 更新规则
	UpdateRule(ctx context.Context, rule *ComplianceRule) error
	// 删除规则
	DeleteRule(ctx context.Context, ruleID string) error
	// 获取检查历史
	GetCheckHistory(ctx context.Context, subjectID string, limit int) ([]*ComplianceCheckResult, error)
}

// RuleRepository 规则仓库接口
type RuleRepository interface {
	// 保存规则
	Save(ctx context.Context, rule *ComplianceRule) error
	// 查找规则
	Find(ctx context.Context, ruleID string) (*ComplianceRule, error)
	// 查找规则列表
	FindAll(ctx context.Context, filter *RuleFilter) ([]*ComplianceRule, error)
	// 更新规则
	Update(ctx context.Context, rule *ComplianceRule) error
	// 删除规则
	Delete(ctx context.Context, ruleID string) error
}

// RuleFilter 规则过滤器
type RuleFilter struct {
	Category string            `json:"category,omitempty"`
	Enabled  *bool             `json:"enabled,omitempty"`
	Priority *int              `json:"priority,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceRepository 合规仓库接口
type ComplianceRepository interface {
	// 保存检查结果
	SaveResult(ctx context.Context, result *ComplianceCheckResult) error
	// 查找检查结果
	FindResult(ctx context.Context, requestID string) (*ComplianceCheckResult, error)
	// 查找历史结果
	FindHistory(ctx context.Context, subjectID string, filter *HistoryFilter) ([]*ComplianceCheckResult, error)
	// 保存违规记录
	SaveViolation(ctx context.Context, violation *Violation) error
	// 查找违规记录
	FindViolations(ctx context.Context, filter *ViolationFilter) ([]*Violation, error)
}

// HistoryFilter 历史记录过滤器
type HistoryFilter struct {
	StartTime   *time.Time         `json:"start_time,omitempty"`
	EndTime     *time.Time         `json:"end_time,omitempty"`
	CheckType   string             `json:"check_type,omitempty"`
	Status      ComplianceStatus   `json:"status,omitempty"`
	RiskLevel   RiskLevel          `json:"risk_level,omitempty"`
	Limit       int                `json:"limit,omitempty"`
	Offset      int                `json:"offset,omitempty"`
}

// ViolationFilter 违规记录过滤器
type ViolationFilter struct {
	RuleID       string            `json:"rule_id,omitempty"`
	Severity     SeverityLevel     `json:"severity,omitempty"`
	Status       ComplianceStatus  `json:"status,omitempty"`
	StartTime    *time.Time        `json:"start_time,omitempty"`
	EndTime      *time.Time        `json:"end_time,omitempty"`
	AssignedTo   string            `json:"assigned_to,omitempty"`
	Limit        int               `json:"limit,omitempty"`
	Offset       int               `json:"offset,omitempty"`
}

// ComplianceStatistics 合规统计
type ComplianceStatistics struct {
	TotalChecks       int64                    `json:"total_checks"`
	PassedChecks      int64                    `json:"passed_checks"`
	FailedChecks      int64                    `json:"failed_checks"`
	PendingChecks     int64                    `json:"pending_checks"`
	OverallScore      float64                  `json:"overall_score"`
	RiskDistribution  map[RiskLevel]int64     `json:"risk_distribution"`
	ViolationCount    int64                    `json:"violation_count"`
	Recommendations   int64                    `json:"recommendations"`
	LastCheckTime     time.Time                `json:"last_check_time"`
	CategoryStats     map[string]interface{}   `json:"category_stats"`
}

// ComplianceMetrics 合规指标
type ComplianceMetrics struct {
	CheckPerformance    map[string]time.Duration `json:"check_performance"`
	RuleEffectiveness   map[string]float64       `json:"rule_effectiveness"`
	TrendAnalysis       map[string]interface{}   `json:"trend_analysis"`
	RiskMetrics         map[string]interface{}   `json:"risk_metrics"`
	ComplianceRate      map[string]float64       `json:"compliance_rate"`
}

// AuditEvent 审计事件
type AuditEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	Timestamp     time.Time              `json:"timestamp"`
	UserID        string                 `json:"user_id,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	ResourceType  string                 `json:"resource_type"`
	ResourceID    string                 `json:"resource_id,omitempty"`
	Action        string                 `json:"action"`
	Outcome       string                 `json:"outcome"`
	Details       map[string]interface{} `json:"details,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
}

// ComplianceConfig 合规配置
type ComplianceConfig struct {
	EngineType           string                 `json:"engine_type"`
	RuleUpdateInterval   time.Duration          `json:"rule_update_interval"`
	CheckTimeout         time.Duration          `json:"check_timeout"`
	MaxConcurrentChecks  int                    `json:"max_concurrent_checks"`
	RetryAttempts        int                    `json:"retry_attempts"`
	NotificationSettings map[string]interface{} `json:"notification_settings"`
	StorageSettings      map[string]interface{} `json:"storage_settings"`
	SecuritySettings     map[string]interface{} `json:"security_settings"`
}

// DefaultComplianceConfig 默认配置
func DefaultComplianceConfig() *ComplianceConfig {
	return &ComplianceConfig{
		EngineType:           "rule_engine",
		RuleUpdateInterval:   time.Hour * 24,
		CheckTimeout:         time.Minute * 30,
		MaxConcurrentChecks:  100,
		RetryAttempts:        3,
		NotificationSettings: map[string]interface{}{
			"email_enabled": true,
			"sms_enabled":  false,
			"webhook_url":   "",
		},
		StorageSettings: map[string]interface{}{
			"retention_days": 2555, // 7年
			"archive_days":  3650, // 10年
		},
		SecuritySettings: map[string]interface{}{
			"encryption_enabled": true,
			"access_log_enabled": true,
			"digital_signature": true,
		},
	}
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error 实现error接口
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// ValidationErrors 多个验证错误
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error 实现error接口
func (ve ValidationErrors) Error() string {
	messages := make([]string, len(ve.Errors))
	for i, err := range ve.Errors {
		messages[i] = err.Error()
	}
	return fmt.Sprintf("validation errors: %s", messages)
}

// 合规引擎实现
type complianceEngine struct {
	config           *ComplianceConfig
	ruleRepo         RuleRepository
	complianceRepo   ComplianceRepository
	logger           *slog.Logger
	mutex            sync.RWMutex
	activeChecks     map[string]context.CancelFunc
	checkCounter     int64
}

// NewComplianceEngine 创建合规引擎
func NewComplianceEngine(config *ComplianceConfig, ruleRepo RuleRepository, complianceRepo ComplianceRepository, logger *slog.Logger) ComplianceEngine {
	if config == nil {
		config = DefaultComplianceConfig()
	}

	return &complianceEngine{
		config:         config,
		ruleRepo:       ruleRepo,
		complianceRepo: complianceRepo,
		logger:         logger,
		activeChecks:   make(map[string]context.CancelFunc),
	}
}

// PerformCheck 执行合规检查
func (ce *complianceEngine) PerformCheck(ctx context.Context, req *ComplianceCheckRequest) (*ComplianceCheckResult, error) {
	startTime := time.Now()

	// 1. 验证请求
	if err := ce.validateRequest(req); err != nil {
		return nil, fmt.Errorf("请求验证失败: %w", err)
	}

	// 2. 检查并发限制
	if err := ce.checkConcurrencyLimit(); err != nil {
		return nil, fmt.Errorf("超过并发限制: %w", err)
	}

	// 3. 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, ce.config.CheckTimeout)
	defer cancel()

	// 4. 注册检查任务
	ce.registerCheck(req.RequestID, cancel)
	defer ce.unregisterCheck(req.RequestID)

	// 5. 获取适用规则
	rules, err := ce.getApplicableRules(checkCtx, req.CheckType)
	if err != nil {
		return nil, fmt.Errorf("获取适用规则失败: %w", err)
	}

	// 6. 执行规则检查
	results := make([]*RuleExecutionResult, 0, len(rules))
	violations := make([]Violation, 0)
	recommendations := make([]Recommendation, 0)

	for _, rule := range rules {
		result, err := ce.executeRule(checkCtx, rule, req)
		if err != nil {
			ce.logger.Error("规则执行失败",
				"rule_id", rule.RuleID,
				"request_id", req.RequestID,
				"error", err)
			continue
		}
		results = append(results, result)

		// 处理违规和建议
		if result.Matched && result.Executed {
			ruleViolations, ruleRecommendations := ce.processRuleOutput(rule, result.Output)
			violations = append(violations, ruleViolations...)
			recommendations = append(recommendations, ruleRecommendations...)
		}
	}

	// 7. 计算总体状态和评分
	overallStatus, overallScore := ce.calculateOverallStatus(results)
	riskLevel := ce.determineRiskLevel(overallScore, violations)

	// 8. 生成必要行动
	requiredActions := ce.generateRequiredActions(violations, recommendations)

	// 9. 创建检查结果
	result := &ComplianceCheckResult{
		RequestID:      req.RequestID,
		CheckType:      req.CheckType,
		SubjectID:      req.SubjectID,
		SubjectType:    req.SubjectType,
		OverallStatus:  overallStatus,
		OverallScore:   overallScore,
		RiskLevel:      riskLevel,
		Violations:     violations,
		Recommendations: recommendations,
		RequiredActions: requiredActions,
		CheckTimestamp: time.Now(),
		NextReviewDate: ce.calculateNextReviewDate(riskLevel),
		CheckedBy:      "compliance_engine",
		ProcessingTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"rules_executed": len(results),
			"rules_matched":  ce.countMatchedRules(results),
		},
	}

	// 10. 保存结果
	if err := ce.complianceRepo.SaveResult(ctx, result); err != nil {
		ce.logger.Error("保存检查结果失败",
			"request_id", req.RequestID,
			"error", err)
	}

	ce.logger.Info("合规检查完成",
		"request_id", req.RequestID,
		"subject_id", req.SubjectID,
		"status", overallStatus,
		"score", overallScore,
		"violations", len(violations),
		"processing_time", result.ProcessingTime)

	return result, nil
}

// BatchCheck 批量检查
func (ce *complianceEngine) BatchCheck(ctx context.Context, requests []*ComplianceCheckRequest) ([]*ComplianceCheckResult, error) {
	results := make([]*ComplianceCheckResult, len(requests))

	// 使用goroutine池并发处理
	semaphore := make(chan struct{}, ce.config.MaxConcurrentChecks)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *ComplianceCheckRequest) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := ce.PerformCheck(ctx, request)
			if err != nil {
				ce.logger.Error("批量检查失败",
					"request_id", request.RequestID,
					"error", err)
				// 创建失败结果
				result = &ComplianceCheckResult{
					RequestID:   request.RequestID,
					CheckType:   request.CheckType,
					SubjectID:   request.SubjectID,
					SubjectType: request.SubjectType,
					OverallStatus: StatusUnknown,
					OverallScore: 0.0,
					RiskLevel:   RiskLevelLow,
					CheckTimestamp: time.Now(),
					CheckedBy:   "compliance_engine",
					Metadata: map[string]interface{}{
						"error": err.Error(),
					},
				}
			}

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, req)
	}

	wg.Wait()
	return results, nil
}

// GetRules 获取规则列表
func (ce *complianceEngine) GetRules(ctx context.Context, category string) ([]*ComplianceRule, error) {
	filter := &RuleFilter{
		Category: category,
		Enabled:  &[]bool{true}[0], // 只返回启用的规则
	}
	return ce.ruleRepo.FindAll(ctx, filter)
}

// AddRule 添加规则
func (ce *complianceEngine) AddRule(ctx context.Context, rule *ComplianceRule) error {
	// 验证规则
	if err := ce.validateRule(rule); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 设置默认值
	if rule.CreationDate.IsZero() {
		rule.CreationDate = time.Now()
	}
	rule.LastUpdated = time.Now()
	if rule.Version == "" {
		rule.Version = "1.0.0"
	}

	return ce.ruleRepo.Save(ctx, rule)
}

// UpdateRule 更新规则
func (ce *complianceEngine) UpdateRule(ctx context.Context, rule *ComplianceRule) error {
	// 验证规则存在
	existing, err := ce.ruleRepo.Find(ctx, rule.RuleID)
	if err != nil {
		return fmt.Errorf("规则不存在: %w", err)
	}

	// 验证规则
	if err := ce.validateRule(rule); err != nil {
		return fmt.Errorf("规则验证失败: %w", err)
	}

	// 更新版本号
	rule.LastUpdated = time.Now()
	rule.Version = ce.incrementVersion(existing.Version)

	return ce.ruleRepo.Update(ctx, rule)
}

// DeleteRule 删除规则
func (ce *complianceEngine) DeleteRule(ctx context.Context, ruleID string) error {
	return ce.ruleRepo.Delete(ctx, ruleID)
}

// GetCheckHistory 获取检查历史
func (ce *complianceEngine) GetCheckHistory(ctx context.Context, subjectID string, limit int) ([]*ComplianceCheckResult, error) {
	filter := &HistoryFilter{
		Limit: limit,
	}
	return ce.complianceRepo.FindHistory(ctx, subjectID, filter)
}

// 以下是私有辅助方法

// validateRequest 验证请求
func (ce *complianceEngine) validateRequest(req *ComplianceCheckRequest) error {
	var errors []ValidationError

	if req.RequestID == "" {
		errors = append(errors, ValidationError{
			Field:   "request_id",
			Message: "请求ID不能为空",
			Code:    "REQUIRED",
		})
	}

	if req.CheckType == "" {
		errors = append(errors, ValidationError{
			Field:   "check_type",
			Message: "检查类型不能为空",
			Code:    "REQUIRED",
		})
	}

	if req.SubjectID == "" {
		errors = append(errors, ValidationError{
			Field:   "subject_id",
			Message: "主体ID不能为空",
			Code:    "REQUIRED",
		})
	}

	if req.SubjectType == "" {
		errors = append(errors, ValidationError{
			Field:   "subject_type",
			Message: "主体类型不能为空",
			Code:    "REQUIRED",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// validateRule 验证规则
func (ce *complianceEngine) validateRule(rule *ComplianceRule) error {
	var errors []ValidationError

	if rule.RuleID == "" {
		errors = append(errors, ValidationError{
			Field:   "rule_id",
			Message: "规则ID不能为空",
			Code:    "REQUIRED",
		})
	}

	if rule.RuleName == "" {
		errors = append(errors, ValidationError{
			Field:   "rule_name",
			Message: "规则名称不能为空",
			Code:    "REQUIRED",
		})
	}

	if rule.Category == "" {
		errors = append(errors, ValidationError{
			Field:   "category",
			Message: "规则分类不能为空",
			Code:    "REQUIRED",
		})
	}

	if len(rule.Conditions) == 0 {
		errors = append(errors, ValidationError{
			Field:   "conditions",
			Message: "规则条件不能为空",
			Code:    "REQUIRED",
		})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// checkConcurrencyLimit 检查并发限制
func (ce *complianceEngine) checkConcurrencyLimit() error {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	if len(ce.activeChecks) >= ce.config.MaxConcurrentChecks {
		return fmt.Errorf("当前活动检查数量: %d, 超过最大并发限制: %d",
			len(ce.activeChecks), ce.config.MaxConcurrentChecks)
	}

	return nil
}

// registerCheck 注册检查任务
func (ce *complianceEngine) registerCheck(requestID string, cancel context.CancelFunc) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	ce.activeChecks[requestID] = cancel
}

// unregisterCheck 取消注册检查任务
func (ce *complianceEngine) unregisterCheck(requestID string) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	delete(ce.activeChecks, requestID)
}

// getApplicableRules 获取适用规则
func (ce *complianceEngine) getApplicableRules(ctx context.Context, checkType string) ([]*ComplianceRule, error) {
	rules, err := ce.ruleRepo.FindAll(ctx, &RuleFilter{
		Category: checkType,
		Enabled:  &[]bool{true}[0],
	})
	if err != nil {
		return nil, err
	}

	// 按优先级排序
	sortedRules := make([]*ComplianceRule, len(rules))
	copy(sortedRules, rules)

	// 简单的优先级排序（降序）
	for i := 0; i < len(sortedRules)-1; i++ {
		for j := i + 1; j < len(sortedRules); j++ {
			if sortedRules[i].Priority < sortedRules[j].Priority {
				sortedRules[i], sortedRules[j] = sortedRules[j], sortedRules[i]
			}
		}
	}

	return sortedRules, nil
}

// executeRule 执行规则
func (ce *complianceEngine) executeRule(ctx context.Context, rule *ComplianceRule, req *ComplianceCheckRequest) (*RuleExecutionResult, error) {
	startTime := time.Now()

	result := &RuleExecutionResult{
		RuleID:    rule.RuleID,
		RuleName:  rule.RuleName,
		Timestamp: time.Now(),
	}

	// 评估条件
	matched, err := ce.evaluateConditions(ctx, rule.Conditions, req.Data)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Matched = matched

	if matched {
		// 执行动作
		output, err := ce.executeActions(ctx, rule.Actions, req.Data)
		if err != nil {
			result.Error = err.Error()
			return result, err
		}

		result.Executed = true
		result.Output = output
	}

	result.ExecutionTime = time.Since(startTime)
	return result, nil
}

// evaluateConditions 评估条件
func (ce *complianceEngine) evaluateConditions(ctx context.Context, conditions []RuleCondition, data map[string]interface{}) (bool, error) {
	if len(conditions) == 0 {
		return true, nil
	}

	// 简化实现：支持AND逻辑
	for _, condition := range conditions {
		result, err := ce.evaluateCondition(ctx, condition, data)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil
		}
	}

	return true, nil
}

// evaluateCondition 评估单个条件
func (ce *complianceEngine) evaluateCondition(ctx context.Context, condition RuleCondition, data map[string]interface{}) (bool, error) {
	switch condition.Type {
	case "field":
		return ce.evaluateFieldCondition(condition, data)
	case "expression":
		return ce.evaluateExpressionCondition(condition, data)
	default:
		return false, fmt.Errorf("不支持的条件类型: %s", condition.Type)
	}
}

// evaluateFieldCondition 评估字段条件
func (ce *complianceEngine) evaluateFieldCondition(condition RuleCondition, data map[string]interface{}) (bool, error) {
	fieldValue, exists := data[condition.Field]
	if !exists {
		return false, nil
	}

	switch condition.Operator {
	case "==":
		return ce.compareValues(fieldValue, condition.Value, "==")
	case "!=":
		return ce.compareValues(fieldValue, condition.Value, "!=")
	case ">":
		return ce.compareValues(fieldValue, condition.Value, ">")
	case "<":
		return ce.compareValues(fieldValue, condition.Value, "<")
	case "contains":
		return ce.stringContains(fieldValue, condition.Value)
	case "in":
		return ce.valueInSlice(fieldValue, condition.Value)
	default:
		return false, fmt.Errorf("不支持的操作符: %s", condition.Operator)
	}
}

// evaluateExpressionCondition 评估表达式条件
func (ce *complianceEngine) evaluateExpressionCondition(condition RuleCondition, data map[string]interface{}) (bool, error) {
	// 简化实现：这里应该集成真正的表达式引擎
	// 例如: github.com/Knetic/govaluate
	return false, fmt.Errorf("表达式条件暂未实现")
}

// compareValues 比较值
func (ce *complianceEngine) compareValues(value1, value2 interface{}, operator string) (bool, error) {
	// 简化实现
	switch v1 := value1.(type) {
	case string:
		if v2, ok := value2.(string); ok {
			switch operator {
			case "==":
				return v1 == v2, nil
			case "!=":
				return v1 != v2, nil
			case "contains":
				return ce.stringContains(value1, value2)
			default:
				return false, fmt.Errorf("字符串不支持操作符: %s", operator)
			}
		}
	case int, int64, float64:
		// 转换为float64比较
		var f1, f2 float64
		switch vt := v1.(type) {
		case int:
			f1 = float64(vt)
		case int64:
			f1 = float64(vt)
		case float64:
			f1 = vt
		}

		switch vt := value2.(type) {
		case int:
			f2 = float64(vt)
		case int64:
			f2 = float64(vt)
		case float64:
			f2 = vt
		default:
			return false, fmt.Errorf("类型不匹配")
		}

		switch operator {
		case "==":
			return f1 == f2, nil
		case "!=":
			return f1 != f2, nil
		case ">":
			return f1 > f2, nil
		case "<":
			return f1 < f2, nil
		default:
			return false, fmt.Errorf("数字不支持操作符: %s", operator)
		}
	case bool:
		if v2, ok := value2.(bool); ok {
			switch operator {
			case "==":
				return v1 == v2, nil
			case "!=":
				return v1 != v2, nil
			default:
				return false, fmt.Errorf("布尔值不支持操作符: %s", operator)
			}
		}
	}

	return false, fmt.Errorf("不支持的值类型: %T", value1)
}

// stringContains 字符串包含检查
func (ce *complianceEngine) stringContains(value, substr interface{}) (bool, error) {
	str, ok := value.(string)
	if !ok {
		return false, fmt.Errorf("值不是字符串")
	}

	sub, ok := substr.(string)
	if !ok {
		return false, fmt.Errorf("子字符串不是字符串")
	}

	return ce.containsString(str, sub), nil
}

// containsString 检查字符串包含（忽略大小写）
func (ce *complianceEngine) containsString(str, substr string) bool {
	// 简化实现
	return len(str) >= len(substr) && str[:len(substr)] == substr
}

// valueInSlice 检查值是否在切片中
func (ce *complianceEngine) valueInSlice(value, slice interface{}) (bool, error) {
	s, ok := slice.([]interface{})
	if !ok {
		return false, fmt.Errorf("目标不是切片")
	}

	for _, item := range s {
		if item == value {
			return true, nil
		}
	}

	return false, nil
}

// executeActions 执行动作
func (ce *complianceEngine) executeActions(ctx context.Context, actions []RuleAction, data map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	for _, action := range actions {
		result, err := ce.executeAction(ctx, action, data)
		if err != nil {
			return nil, fmt.Errorf("执行动作失败 [%s]: %w", action.Type, err)
		}

		// 合并结果
		for k, v := range result {
			output[k] = v
		}
	}

	return output, nil
}

// executeAction 执行单个动作
func (ce *complianceEngine) executeAction(ctx context.Context, action RuleAction, data map[string]interface{}) (map[string]interface{}, error) {
	switch action.Type {
	case "set_violation":
		return ce.executeSetViolationAction(action, data)
	case "set_recommendation":
		return ce.executeSetRecommendationAction(action, data)
	case "calculate_score":
		return ce.executeCalculateScoreAction(action, data)
	default:
		return nil, fmt.Errorf("不支持的动作类型: %s", action.Type)
	}
}

// executeSetViolationAction 执行设置违规动作
func (ce *complianceEngine) executeSetViolationAction(action RuleAction, data map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"violation": true,
		"severity":  ce.getActionParameter(action, "severity", "MEDIUM"),
		"message":   ce.getActionParameter(action, "message", "合规检查发现违规"),
	}, nil
}

// executeSetRecommendationAction 执行设置建议动作
func (ce *complianceEngine) executeSetRecommendationAction(action RuleAction, data map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"recommendation": true,
		"priority":       ce.getActionParameter(action, "priority", 1),
		"title":          ce.getActionParameter(action, "title", "合规建议"),
		"description":    ce.getActionParameter(action, "description", "建议采取相应措施"),
	}, nil
}

// executeCalculateScoreAction 执行计算评分动作
func (ce *complianceEngine) executeCalculateScoreAction(action RuleAction, data map[string]interface{}) (map[string]interface{}, error) {
	// 简化实现：固定返回评分
	return map[string]interface{}{
		"score": ce.getActionParameter(action, "score", 50.0),
	}, nil
}

// getActionParameter 获取动作参数
func (ce *complianceEngine) getActionParameter(action RuleAction, key string, defaultValue interface{}) interface{} {
	if action.Parameters == nil {
		return defaultValue
	}

	if value, exists := action.Parameters[key]; exists {
		return value
	}

	return defaultValue
}

// processRuleOutput 处理规则输出
func (ce *complianceEngine) processRuleOutput(rule *ComplianceRule, output map[string]interface{}) ([]Violation, []Recommendation) {
	var violations []Violation
	var recommendations []Recommendation

	// 处理违规
	if violation, ok := output["violation"].(bool); ok && violation {
		v := Violation{
			ViolationID:      ce.generateID("violation"),
			RuleID:          rule.RuleID,
			RuleName:        rule.RuleName,
			Description:     ce.getStringFromOutput(output, "message", "规则违规"),
			Severity:        SeverityLevel(ce.getStringFromOutput(output, "severity", "MEDIUM")),
			DetectedAt:      time.Now(),
			Evidence:        output,
			Status:          StatusNonCompliant,
		}
		violations = append(violations, v)
	}

	// 处理建议
	if recommendation, ok := output["recommendation"].(bool); ok && recommendation {
		r := Recommendation{
			RecommendationID: ce.generateID("recommendation"),
			Priority:         ce.getIntFromOutput(output, "priority", 1),
			Category:         rule.Category,
			Title:            ce.getStringFromOutput(output, "title", "合规建议"),
			Description:      ce.getStringFromOutput(output, "description", "建议采取相应措施"),
			Status:           StatusPending,
		}
		recommendations = append(recommendations, r)
	}

	return violations, recommendations
}

// calculateOverallStatus 计算总体状态
func (ce *complianceEngine) calculateOverallStatus(results []*RuleExecutionResult) (ComplianceStatus, float64) {
	if len(results) == 0 {
		return StatusUnknown, 0.0
	}

	totalScore := 0.0
	hasViolations := false

	for _, result := range results {
		if result.Matched && result.Executed {
			if _, hasViolation := result.Output["violation"]; hasViolation {
				hasViolations = true
			}

			if score, hasScore := result.Output["score"]; hasScore {
				if s, ok := score.(float64); ok {
					totalScore += s
				}
			}
		}
	}

	averageScore := totalScore / float64(len(results))

	if hasViolations {
		return StatusNonCompliant, averageScore
	}

	if averageScore >= 80.0 {
		return StatusCompliant, averageScore
	}

	return StatusPending, averageScore
}

// determineRiskLevel 确定风险等级
func (ce *complianceEngine) determineRiskLevel(score float64, violations []Violation) RiskLevel {
	if len(violations) == 0 {
		if score >= 90.0 {
			return RiskLevelLow
		}
		return RiskLevelMedium
	}

	// 检查是否有严重违规
	for _, violation := range violations {
		if violation.Severity == SeverityCritical {
			return RiskLevelCritical
		}
		if violation.Severity == SeverityHigh {
			return RiskLevelHigh
		}
	}

	return RiskLevelMedium
}

// generateRequiredActions 生成必要行动
func (ce *complianceEngine) generateRequiredActions(violations []Violation, recommendations []Recommendation) []RequiredAction {
	var actions []RequiredAction

	// 为违规生成必要行动
	for _, violation := range violations {
		action := RequiredAction{
			ActionID:    ce.generateID("action"),
			Title:       "处理违规: " + violation.RuleName,
			Description: violation.Description,
			Priority:    ce.getPriorityFromSeverity(violation.Severity),
			Deadline:    time.Now().AddDate(0, 0, ce.getDeadlineDays(violation.Severity)),
			Status:      StatusPending,
			Metadata: map[string]interface{}{
				"violation_id": violation.ViolationID,
				"rule_id":      violation.RuleID,
			},
		}
		actions = append(actions, action)
	}

	return actions
}

// getPriorityFromSeverity 从严重级别获取优先级
func (ce *complianceEngine) getPriorityFromSeverity(severity SeverityLevel) int {
	switch severity {
	case SeverityCritical:
		return 1
	case SeverityHigh:
		return 2
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 4
	default:
		return 5
	}
}

// getDeadlineDays 获取处理期限天数
func (ce *complianceEngine) getDeadlineDays(severity SeverityLevel) int {
	switch severity {
	case SeverityCritical:
		return 1
	case SeverityHigh:
		return 7
	case SeverityMedium:
		return 14
	case SeverityLow:
		return 30
	default:
		return 30
	}
}

// calculateNextReviewDate 计算下次审查日期
func (ce *complianceEngine) calculateNextReviewDate(riskLevel RiskLevel) *time.Time {
	var days int
	switch riskLevel {
	case RiskLevelCritical:
		days = 30
	case RiskLevelHigh:
		days = 60
	case RiskLevelMedium:
		days = 90
	case RiskLevelLow:
		days = 180
	default:
		days = 90
	}

	nextReview := time.Now().AddDate(0, 0, days)
	return &nextReview
}

// countMatchedRules 统计匹配的规则数量
func (ce *complianceEngine) countMatchedRules(results []*RuleExecutionResult) int {
	count := 0
	for _, result := range results {
		if result.Matched {
			count++
		}
	}
	return count
}

// incrementVersion 增加版本号
func (ce *complianceEngine) incrementVersion(version string) string {
	// 简化实现：只增加最后一位
	return version + ".1"
}

// generateID 生成ID
func (ce *complianceEngine) generateID(prefix string) string {
	// 简化实现：使用时间戳和计数器
	ce.checkCounter++
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), ce.checkCounter)
}

// getStringFromOutput 从输出中获取字符串值
func (ce *complianceEngine) getStringFromOutput(output map[string]interface{}, key, defaultValue string) string {
	if value, exists := output[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getIntFromOutput 从输出中获取整数值
func (ce *complianceEngine) getIntFromOutput(output map[string]interface{}, key string, defaultValue int) int {
	if value, exists := output[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			// 尝试解析字符串为整数
			if i, err := json.Number(v).Int64(); err == nil {
				return int(i)
			}
		}
	}
	return defaultValue
}