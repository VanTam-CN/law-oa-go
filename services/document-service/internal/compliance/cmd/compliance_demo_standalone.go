package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
)

// 简化版合规检查系统实现

// 合规状态
type ComplianceStatus string

const (
	StatusCompliant    ComplianceStatus = "COMPLIANT"
	StatusNonCompliant ComplianceStatus = "NON_COMPLIANT"
	StatusPending      ComplianceStatus = "PENDING"
	StatusUnknown      ComplianceStatus = "UNKNOWN"
)

// 严重级别
type SeverityLevel string

const (
	SeverityLow    SeverityLevel = "LOW"
	SeverityMedium SeverityLevel = "MEDIUM"
	SeverityHigh   SeverityLevel = "HIGH"
	SeverityCritical SeverityLevel = "CRITICAL"
)

// 风险等级
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "LOW"
	RiskLevelMedium RiskLevel = "MEDIUM"
	RiskLevelHigh   RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

// 违规记录
type Violation struct {
	ViolationID     string                 `json:"violation_id"`
	RuleID          string                 `json:"rule_id"`
	RuleName        string                 `json:"rule_name"`
	Description     string                 `json:"description"`
	Severity        SeverityLevel          `json:"severity"`
	DetectedAt      time.Time              `json:"detected_at"`
	Evidence        map[string]interface{} `json:"evidence"`
	Remediation     string                 `json:"remediation"`
	Status          ComplianceStatus       `json:"status"`
}

// 建议
type Recommendation struct {
	RecommendationID string            `json:"recommendation_id"`
	Priority         int               `json:"priority"`
	Category         string            `json:"category"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	ActionItems      []string          `json:"action_items"`
	Status           ComplianceStatus   `json:"status"`
}

// 必要行动
type RequiredAction struct {
	ActionID    string          `json:"action_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Priority    int             `json:"priority"`
	Category    string          `json:"category"`
	Deadline    time.Time       `json:"deadline"`
	Status      ComplianceStatus `json:"status"`
}

// 合规检查请求
type ComplianceCheckRequest struct {
	RequestID   string                 `json:"request_id"`
	CheckType   string                 `json:"check_type"`
	SubjectID   string                 `json:"subject_id"`
	SubjectType string                 `json:"subject_type"`
	Data        map[string]interface{} `json:"data"`
	Priority    int                    `json:"priority"`
	RequestedBy string                 `json:"requested_by"`
}

// 合规检查结果
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
	CheckedBy          string               `json:"checked_by"`
	ProcessingTime     time.Duration        `json:"processing_time"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// 合规规则
type ComplianceRule struct {
	RuleID      string             `json:"rule_id"`
	RuleName    string             `json:"rule_name"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Version     string             `json:"version"`
	Enabled     bool               `json:"enabled"`
	Priority    int                `json:"priority"`
	Severity    SeverityLevel      `json:"severity"`
}

// 合规引擎接口
type ComplianceEngine interface {
	PerformCheck(ctx context.Context, req *ComplianceCheckRequest) (*ComplianceCheckResult, error)
	BatchCheck(ctx context.Context, requests []*ComplianceCheckRequest) ([]*ComplianceCheckResult, error)
	GetRules(ctx context.Context, category string) ([]*ComplianceRule, error)
	GetCheckHistory(ctx context.Context, subjectID string, limit int) ([]*ComplianceCheckResult, error)
	AddRule(rule *ComplianceRule)
}

// 简化的合规引擎实现
type complianceEngine struct {
	rules    map[string]*ComplianceRule
	results  map[string]*ComplianceCheckResult
	history  map[string][]*ComplianceCheckResult
	mutex    sync.RWMutex
	logger   *slog.Logger
}

// 创建新的合规引擎
func NewComplianceEngine(logger *slog.Logger) ComplianceEngine {
	return &complianceEngine{
		rules:   make(map[string]*ComplianceRule),
		results: make(map[string]*ComplianceCheckResult),
		history: make(map[string][]*ComplianceCheckResult),
		logger:  logger,
	}
}

// 执行合规检查
func (ce *complianceEngine) PerformCheck(ctx context.Context, req *ComplianceCheckRequest) (*ComplianceCheckResult, error) {
	startTime := time.Now()

	ce.logger.Info("开始合规检查",
		"request_id", req.RequestID,
		"check_type", req.CheckType,
		"subject_id", req.SubjectID)

	// 获取适用规则
	rules := ce.getApplicableRules(req.CheckType)

	// 执行规则检查
	violations := make([]Violation, 0)
	recommendations := make([]Recommendation, 0)
	overallScore := 100.0

	for _, rule := range rules {
		if ce.evaluateRule(rule, req.Data) {
			// 规则匹配，创建违规
			violation := Violation{
				ViolationID: ce.generateID("violation"),
				RuleID:      rule.RuleID,
				RuleName:    rule.RuleName,
				Description: fmt.Sprintf("违反规则: %s", rule.Description),
				Severity:    rule.Severity,
				DetectedAt:  time.Now(),
				Evidence:    req.Data,
				Status:      StatusNonCompliant,
			}
			violations = append(violations, violation)

			// 生成建议
			recommendation := Recommendation{
				RecommendationID: ce.generateID("recommendation"),
				Priority:         rule.Priority,
				Category:         rule.Category,
				Title:            fmt.Sprintf("处理: %s", rule.RuleName),
				Description:      fmt.Sprintf("建议针对: %s", rule.Description),
				ActionItems:      []string{"分析违规原因", "制定纠正措施", "更新相关流程"},
				Status:           StatusPending,
			}
			recommendations = append(recommendations, recommendation)

			// 调整评分
			overallScore -= ce.getScorePenalty(rule.Severity)
		}
	}

	// 确定总体状态和风险等级
	overallStatus := ce.determineOverallStatus(overallScore, violations)
	riskLevel := ce.determineRiskLevel(overallScore, violations)

	// 生成必要行动
	requiredActions := ce.generateRequiredActions(violations, recommendations)

	// 创建检查结果
	result := &ComplianceCheckResult{
		RequestID:       req.RequestID,
		CheckType:       req.CheckType,
		SubjectID:       req.SubjectID,
		SubjectType:     req.SubjectType,
		OverallStatus:   overallStatus,
		OverallScore:    overallScore,
		RiskLevel:       riskLevel,
		Violations:      violations,
		Recommendations: recommendations,
		RequiredActions: requiredActions,
		CheckTimestamp:  time.Now(),
		CheckedBy:       "compliance_engine",
		ProcessingTime:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"rules_evaluated": len(rules),
			"violations_found": len(violations),
		},
	}

	// 保存结果
	ce.saveResult(result)

	ce.logger.Info("合规检查完成",
		"request_id", result.RequestID,
		"status", result.OverallStatus,
		"score", result.OverallScore,
		"violations", len(violations),
		"processing_time", result.ProcessingTime)

	return result, nil
}

// 批量检查
func (ce *complianceEngine) BatchCheck(ctx context.Context, requests []*ComplianceCheckRequest) ([]*ComplianceCheckResult, error) {
	results := make([]*ComplianceCheckResult, len(requests))

	for i, req := range requests {
		result, err := ce.PerformCheck(ctx, req)
		if err != nil {
			ce.logger.Error("批量检查失败", "request_id", req.RequestID, "error", err)
			// 创建失败结果
			result = &ComplianceCheckResult{
				RequestID:   req.RequestID,
				CheckType:   req.CheckType,
				SubjectID:   req.SubjectID,
				SubjectType: req.SubjectType,
				OverallStatus: StatusUnknown,
				OverallScore: 0.0,
				RiskLevel:   RiskLevelLow,
				CheckTimestamp: time.Now(),
				CheckedBy:   "compliance_engine",
				ProcessingTime: time.Duration(0),
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			}
		}
		results[i] = result
	}

	return results, nil
}

// 获取规则列表
func (ce *complianceEngine) GetRules(ctx context.Context, category string) ([]*ComplianceRule, error) {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	var rules []*ComplianceRule
	for _, rule := range ce.rules {
		if category == "" || rule.Category == category {
			rules = append(rules, rule)
		}
	}

	return rules, nil
}

// 获取检查历史
func (ce *complianceEngine) GetCheckHistory(ctx context.Context, subjectID string, limit int) ([]*ComplianceCheckResult, error) {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	history, exists := ce.history[subjectID]
	if !exists {
		return []*ComplianceCheckResult{}, nil
	}

	if limit > 0 && limit < len(history) {
		return history[:limit], nil
	}

	return history, nil
}

// 获取适用规则
func (ce *complianceEngine) getApplicableRules(checkType string) []*ComplianceRule {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()

	var rules []*ComplianceRule
	for _, rule := range ce.rules {
		if rule.Enabled && (rule.Category == checkType || rule.Category == "general") {
			rules = append(rules, rule)
		}
	}

	// 按优先级排序
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority < rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	return rules
}

// 评估规则
func (ce *complianceEngine) evaluateRule(rule *ComplianceRule, data map[string]interface{}) bool {
	// 简化实现：基于规则ID进行基本检查
	switch rule.RuleID {
	case "aml_pep_check":
		if pep, exists := data["politically_exposed"]; exists {
			if pepVal, ok := pep.(bool); ok && pepVal {
				return true
			}
		}
	case "aml_sanction_check":
		if sanctioned, exists := data["sanctioned"]; exists {
			if sanctionedVal, ok := sanctioned.(bool); ok && sanctionedVal {
				return true
			}
		}
	case "aml_high_value_transaction":
		if hvTx, exists := data["high_value_transactions"]; exists {
			if hvTxVal, ok := hvTx.(int); ok && hvTxVal > 0 {
				return true
			}
		}
	case "kyc_document_expiry":
		if expired, exists := data["document_expired"]; exists {
			if expiredVal, ok := expired.(bool); ok && expiredVal {
				return true
			}
		}
	case "kyc_address_verification":
		if addrVerified, exists := data["address_verified"]; exists {
			if addrVerifiedVal, ok := addrVerified.(bool); ok && !addrVerifiedVal {
				return true
			}
		}
	}

	return false
}

// 确定总体状态
func (ce *complianceEngine) determineOverallStatus(score float64, violations []Violation) ComplianceStatus {
	if len(violations) > 0 {
		return StatusNonCompliant
	}

	if score >= 90 {
		return StatusCompliant
	} else if score >= 70 {
		return StatusPending
	}

	return StatusNonCompliant
}

// 确定风险等级
func (ce *complianceEngine) determineRiskLevel(score float64, violations []Violation) RiskLevel {
	if len(violations) > 0 {
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

	if score >= 90 {
		return RiskLevelLow
	} else if score >= 70 {
		return RiskLevelMedium
	}

	return RiskLevelHigh
}

// 生成必要行动
func (ce *complianceEngine) generateRequiredActions(violations []Violation, recommendations []Recommendation) []RequiredAction {
	var actions []RequiredAction

	// 为每个违规生成必要行动
	for _, violation := range violations {
		action := RequiredAction{
			ActionID:    ce.generateID("action"),
			Title:       fmt.Sprintf("处理违规: %s", violation.RuleName),
			Description: violation.Description,
			Priority:    ce.getPriorityFromSeverity(violation.Severity),
			Category:    "violation_handling",
			Deadline:    time.Now().AddDate(0, 0, ce.getDeadlineDays(violation.Severity)),
			Status:      StatusPending,
		}
		actions = append(actions, action)
	}

	return actions
}

// 获取评分惩罚
func (ce *complianceEngine) getScorePenalty(severity SeverityLevel) float64 {
	switch severity {
	case SeverityCritical:
		return 40
	case SeverityHigh:
		return 25
	case SeverityMedium:
		return 15
	case SeverityLow:
		return 5
	default:
		return 10
	}
}

// 从严重级别获取优先级
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

// 获取处理期限天数
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

// 保存结果
func (ce *complianceEngine) saveResult(result *ComplianceCheckResult) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()

	// 保存结果
	ce.results[result.RequestID] = result

	// 保存到历史记录
	if _, exists := ce.history[result.SubjectID]; !exists {
		ce.history[result.SubjectID] = []*ComplianceCheckResult{}
	}
	ce.history[result.SubjectID] = append([]*ComplianceCheckResult{result}, ce.history[result.SubjectID]...)

	// 保持历史记录数量限制
	if len(ce.history[result.SubjectID]) > 100 {
		ce.history[result.SubjectID] = ce.history[result.SubjectID][:100]
	}
}

// 生成ID
func (ce *complianceEngine) generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// 添加规则
func (ce *complianceEngine) AddRule(rule *ComplianceRule) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	ce.rules[rule.RuleID] = rule
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("合规检查系统演示启动")

	// 创建合规引擎
	engine := NewComplianceEngine(logger)

	// 初始化示例规则
	initSampleRules(engine, logger)

	// 运行演示
	if err := runComplianceDemo(engine, logger); err != nil {
		log.Fatalf("演示执行失败: %v", err)
	}

	logger.Info("合规检查系统演示完成")
}

// 初始化示例规则
func initSampleRules(engine ComplianceEngine, logger *slog.Logger) {
	rules := []*ComplianceRule{
		{
			RuleID:      "aml_pep_check",
			RuleName:    "PEP客户检查",
			Description: "检查客户是否为政治公众人物",
			Category:    "aml",
			Version:     "1.0.0",
			Enabled:     true,
			Priority:    1,
			Severity:    SeverityHigh,
		},
		{
			RuleID:      "aml_sanction_check",
			RuleName:    "制裁名单检查",
			Description: "检查客户是否在制裁名单上",
			Category:    "aml",
			Version:     "1.0.0",
			Enabled:     true,
			Priority:    1,
			Severity:    SeverityCritical,
		},
		{
			RuleID:      "aml_high_value_transaction",
			RuleName:    "大额交易检查",
			Description: "检查是否有大额交易",
			Category:    "aml",
			Version:     "1.0.0",
			Enabled:     true,
			Priority:    2,
			Severity:    SeverityMedium,
		},
		{
			RuleID:      "kyc_document_expiry",
			RuleName:    "文档有效期检查",
			Description: "检查客户文档是否在有效期内",
			Category:    "kyc",
			Version:     "1.0.0",
			Enabled:     true,
			Priority:    2,
			Severity:    SeverityMedium,
		},
		{
			RuleID:      "kyc_address_verification",
			RuleName:    "地址验证检查",
			Description: "检查客户地址是否已验证",
			Category:    "kyc",
			Version:     "1.0.0",
			Enabled:     true,
			Priority:    3,
			Severity:    SeverityLow,
		},
	}

	for _, rule := range rules {
		engine.AddRule(rule)
		logger.Info("规则已添加", "rule_id", rule.RuleID, "rule_name", rule.RuleName)
	}
}

// 运行合规检查演示
func runComplianceDemo(engine ComplianceEngine, logger *slog.Logger) error {
	ctx := context.Background()

	logger.Info("=== 合规检查演示开始 ===")

	// 1. 获取规则列表
	logger.Info("--- 获取规则列表 ---")
	amlRules, err := engine.GetRules(ctx, "aml")
	if err != nil {
		return fmt.Errorf("获取AML规则失败: %w", err)
	}
	logger.Info("AML规则数量", "count", len(amlRules))

	kycRules, err := engine.GetRules(ctx, "kyc")
	if err != nil {
		return fmt.Errorf("获取KYC规则失败: %w", err)
	}
	logger.Info("KYC规则数量", "count", len(kycRules))

	// 2. 执行AML检查演示
	logger.Info("--- AML检查演示 ---")
	if err := runAMLCheckDemo(ctx, engine, logger); err != nil {
		return fmt.Errorf("AML检查演示失败: %w", err)
	}

	// 3. 执行KYC检查演示
	logger.Info("--- KYC检查演示 ---")
	if err := runKYCCheckDemo(ctx, engine, logger); err != nil {
		return fmt.Errorf("KYC检查演示失败: %w", err)
	}

	// 4. 批量检查演示
	logger.Info("--- 批量检查演示 ---")
	if err := runBatchCheckDemo(ctx, engine, logger); err != nil {
		return fmt.Errorf("批量检查演示失败: %w", err)
	}

	// 5. 历史记录查询演示
	logger.Info("--- 历史记录查询演示 ---")
	if err := runHistoryQueryDemo(ctx, engine, logger); err != nil {
		return fmt.Errorf("历史记录查询演示失败: %w", err)
	}

	// 6. 性能统计演示
	logger.Info("--- 性能统计演示 ---")
	if err := runPerformanceStatsDemo(ctx, engine, logger); err != nil {
		return fmt.Errorf("性能统计演示失败: %w", err)
	}

	logger.Info("=== 合规检查演示完成 ===")
	return nil
}

// 运行AML检查演示
func runAMLCheckDemo(ctx context.Context, engine ComplianceEngine, logger *slog.Logger) error {
	// 正常客户检查
	req1 := &ComplianceCheckRequest{
		RequestID:   "aml_demo_001",
		CheckType:   "aml",
		SubjectID:   "client_001",
		SubjectType: "client",
		Data: map[string]interface{}{
			"politically_exposed":      false,
			"sanctioned":              false,
			"high_value_transactions": 0,
			"annual_income":           100000.0,
			"occupation":              "软件工程师",
			"industry":                "信息技术",
		},
		Priority:    2,
		RequestedBy: "demo_user",
	}

	logger.Info("执行AML检查 - 正常客户", "request_id", req1.RequestID)
	result1, err := engine.PerformCheck(ctx, req1)
	if err != nil {
		return fmt.Errorf("执行AML检查失败: %w", err)
	}

	logger.Info("AML检查结果 - 正常客户",
		"status", result1.OverallStatus,
		"score", result1.OverallScore,
		"risk_level", result1.RiskLevel,
		"violations", len(result1.Violations),
		"processing_time", result1.ProcessingTime)

	// PEP客户检查
	req2 := &ComplianceCheckRequest{
		RequestID:   "aml_demo_002",
		CheckType:   "aml",
		SubjectID:   "client_002",
		SubjectType: "client",
		Data: map[string]interface{}{
			"politically_exposed":      true,
			"sanctioned":              false,
			"high_value_transactions": 1,
			"annual_income":           500000.0,
			"occupation":              "政府官员",
			"industry":                "公共管理",
		},
		Priority:    1,
		RequestedBy: "demo_user",
	}

	logger.Info("执行AML检查 - PEP客户", "request_id", req2.RequestID)
	result2, err := engine.PerformCheck(ctx, req2)
	if err != nil {
		return fmt.Errorf("执行AML检查失败: %w", err)
	}

	logger.Info("AML检查结果 - PEP客户",
		"status", result2.OverallStatus,
		"score", result2.OverallScore,
		"risk_level", result2.RiskLevel,
		"violations", len(result2.Violations),
		"processing_time", result2.ProcessingTime)

	// 制裁名单客户检查
	req3 := &ComplianceCheckRequest{
		RequestID:   "aml_demo_003",
		CheckType:   "aml",
		SubjectID:   "client_003",
		SubjectType: "client",
		Data: map[string]interface{}{
			"politically_exposed":      false,
			"sanctioned":              true,
			"high_value_transactions": 3,
			"annual_income":           1000000.0,
			"occupation":              "企业主",
			"industry":                "贸易",
		},
		Priority:    1,
		RequestedBy: "demo_user",
	}

	logger.Info("执行AML检查 - 制裁名单客户", "request_id", req3.RequestID)
	result3, err := engine.PerformCheck(ctx, req3)
	if err != nil {
		return fmt.Errorf("执行AML检查失败: %w", err)
	}

	logger.Info("AML检查结果 - 制裁名单客户",
		"status", result3.OverallStatus,
		"score", result3.OverallScore,
		"risk_level", result3.RiskLevel,
		"violations", len(result3.Violations),
		"processing_time", result3.ProcessingTime)

	return nil
}

// 运行KYC检查演示
func runKYCCheckDemo(ctx context.Context, engine ComplianceEngine, logger *slog.Logger) error {
	// 正常客户检查
	req1 := &ComplianceCheckRequest{
		RequestID:   "kyc_demo_001",
		CheckType:   "kyc",
		SubjectID:   "client_004",
		SubjectType: "client",
		Data: map[string]interface{}{
			"document_expired":  false,
			"address_verified":  true,
			"biometric_verified": true,
			"documents_provided": []string{"身份证", "护照", "驾驶证"},
		},
		Priority:    2,
		RequestedBy: "demo_user",
	}

	logger.Info("执行KYC检查 - 正常客户", "request_id", req1.RequestID)
	result1, err := engine.PerformCheck(ctx, req1)
	if err != nil {
		return fmt.Errorf("执行KYC检查失败: %w", err)
	}

	logger.Info("KYC检查结果 - 正常客户",
		"status", result1.OverallStatus,
		"score", result1.OverallScore,
		"risk_level", result1.RiskLevel,
		"violations", len(result1.Violations),
		"processing_time", result1.ProcessingTime)

	// 文档过期客户检查
	req2 := &ComplianceCheckRequest{
		RequestID:   "kyc_demo_002",
		CheckType:   "kyc",
		SubjectID:   "client_005",
		SubjectType: "client",
		Data: map[string]interface{}{
			"document_expired":  true,
			"address_verified":  true,
			"biometric_verified": false,
			"documents_provided": []string{"身份证"},
		},
		Priority:    2,
		RequestedBy: "demo_user",
	}

	logger.Info("执行KYC检查 - 文档过期客户", "request_id", req2.RequestID)
	result2, err := engine.PerformCheck(ctx, req2)
	if err != nil {
		return fmt.Errorf("执行KYC检查失败: %w", err)
	}

	logger.Info("KYC检查结果 - 文档过期客户",
		"status", result2.OverallStatus,
		"score", result2.OverallScore,
		"risk_level", result2.RiskLevel,
		"violations", len(result2.Violations),
		"processing_time", result2.ProcessingTime)

	return nil
}

// 运行批量检查演示
func runBatchCheckDemo(ctx context.Context, engine ComplianceEngine, logger *slog.Logger) error {
	requests := []*ComplianceCheckRequest{
		{
			RequestID:   "batch_demo_001",
			CheckType:   "aml",
			SubjectID:   "client_006",
			SubjectType: "client",
			Data: map[string]interface{}{
				"politically_exposed": false,
				"sanctioned":           false,
				"high_value_transactions": 2,
			},
			Priority:    2,
			RequestedBy: "demo_user",
		},
		{
			RequestID:   "batch_demo_002",
			CheckType:   "kyc",
			SubjectID:   "client_007",
			SubjectType: "client",
			Data: map[string]interface{}{
				"document_expired":  false,
				"address_verified":  false,
			},
			Priority:    2,
			RequestedBy: "demo_user",
		},
		{
			RequestID:   "batch_demo_003",
			CheckType:   "aml",
			SubjectID:   "client_008",
			SubjectType: "client",
			Data: map[string]interface{}{
				"politically_exposed": true,
				"high_value_transactions": 5,
			},
			Priority:    1,
			RequestedBy: "demo_user",
		},
	}

	logger.Info("执行批量检查", "request_count", len(requests))
	startTime := time.Now()

	results, err := engine.BatchCheck(ctx, requests)
	if err != nil {
		return fmt.Errorf("执行批量检查失败: %w", err)
	}

	totalTime := time.Since(startTime)
	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.OverallStatus != StatusUnknown {
			successCount++
		} else {
			failureCount++
		}
	}

	logger.Info("批量检查结果",
		"total_requests", len(requests),
		"success_count", successCount,
		"failure_count", failureCount,
		"total_processing_time", totalTime,
		"avg_processing_time", totalTime/time.Duration(len(requests)))

	// 输出每个结果
	for i, result := range results {
		logger.Info("批量检查项目结果",
			"index", i+1,
			"request_id", result.RequestID,
			"check_type", result.CheckType,
			"status", result.OverallStatus,
			"score", result.OverallScore,
			"violations", len(result.Violations))
	}

	return nil
}

// 运行历史记录查询演示
func runHistoryQueryDemo(ctx context.Context, engine ComplianceEngine, logger *slog.Logger) error {
	// 查询客户的历史检查记录
	subjectID := "client_001"
	limit := 5

	logger.Info("查询历史检查记录", "subject_id", subjectID, "limit", limit)

	history, err := engine.GetCheckHistory(ctx, subjectID, limit)
	if err != nil {
		return fmt.Errorf("查询历史记录失败: %w", err)
	}

	logger.Info("历史记录查询结果",
		"subject_id", subjectID,
		"record_count", len(history))

	for i, record := range history {
		logger.Info("历史记录",
			"index", i+1,
			"request_id", record.RequestID,
			"check_type", record.CheckType,
			"status", record.OverallStatus,
			"score", record.OverallScore,
			"check_timestamp", record.CheckTimestamp)
	}

	return nil
}

// 运行性能统计演示
func runPerformanceStatsDemo(ctx context.Context, engine ComplianceEngine, logger *slog.Logger) error {
	// 模拟多个检查请求来测试性能
	requestCount := 100
	logger.Info("开始性能测试", "request_count", requestCount)

	startTime := time.Now()
	successCount := 0

	for i := 0; i < requestCount; i++ {
		req := &ComplianceCheckRequest{
			RequestID:   fmt.Sprintf("perf_test_%d", i),
			CheckType:   "aml",
			SubjectID:   fmt.Sprintf("perf_client_%d", i%10), // 10个不同客户
			SubjectType: "client",
			Data: map[string]interface{}{
				"politically_exposed":      i%5 == 0, // 每5个中1个PEP
				"sanctioned":              i%20 == 0, // 每20个中1个制裁
				"high_value_transactions": i % 3,
				"annual_income":           float64(50000 + i*1000),
			},
			Priority:    2,
			RequestedBy: "perf_test",
		}

		result, err := engine.PerformCheck(ctx, req)
		if err != nil {
			logger.Warn("性能测试请求失败", "request_id", req.RequestID, "error", err)
		} else {
			if result.OverallStatus != StatusUnknown {
				successCount++
			}
		}
	}

	totalTime := time.Since(startTime)
	avgTime := totalTime / time.Duration(requestCount)
	throughput := float64(requestCount) / totalTime.Seconds()

	logger.Info("性能测试结果",
		"request_count", requestCount,
		"success_count", successCount,
		"total_time", totalTime,
		"avg_time_per_request", avgTime,
		"throughput_requests_per_second", throughput,
		"success_rate", float64(successCount)/float64(requestCount)*100)

	return nil
}