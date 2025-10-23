package auth

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// ABACEngine 基于属性的访问控制引擎
type ABACEngine struct {
	policyRepo    repositories.PolicyRepository
	roleRepo      repositories.RoleRepository
	userRepo      repositories.UserRepository
	docRepo       repositories.DocumentRepository
	auditRepo     repositories.AuditRepository
	logger        *logrus.Logger
	ruleCache     map[string]*PolicyRule
	cacheExpiry   time.Duration
	lastCacheSync time.Time
}

// NewABACEngine 创建ABAC引擎
func NewABACEngine(
	policyRepo repositories.PolicyRepository,
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	docRepo repositories.DocumentRepository,
	auditRepo repositories.AuditRepository,
	logger *logrus.Logger,
) *ABACEngine {
	return &ABACEngine{
		policyRepo:  policyRepo,
		roleRepo:    roleRepo,
		userRepo:    userRepo,
		docRepo:     docRepo,
		auditRepo:   auditRepo,
		logger:      logger,
		ruleCache:   make(map[string]*PolicyRule),
		cacheExpiry: 5 * time.Minute,
	}
}

// PolicyRule 策略规则
type PolicyRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     int                    `json:"version"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Effect      string                 `json:"effect"` // allow, deny
	Subject     SubjectMatcher         `json:"subject"`
	Resource    ResourceMatcher        `json:"resource"`
	Action      ActionMatcher          `json:"action"`
	Environment EnvironmentMatcher     `json:"environment"`
	Conditions  []ConditionExpression  `json:"conditions"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	CreatedBy   uint                  `json:"created_by"`
	TenantID    string                `json:"tenant_id"`
}

// SubjectMatcher 主体匹配器
type SubjectMatcher struct {
	ID        []string          `json:"id,omitempty"`        // 用户ID列表
	Roles     []string          `json:"roles,omitempty"`     // 角色列表
	Groups    []string          `json:"groups,omitempty"`    // 用户组列表
	Email     []string          `json:"email,omitempty"`     // 邮箱模式
	Attributes map[string]interface{} `json:"attributes,omitempty"` // 自定义属性
}

// ResourceMatcher 资源匹配器
type ResourceMatcher struct {
	Type       []string          `json:"type,omitempty"`       // 资源类型
	ID         []string          `json:"id,omitempty"`         // 资源ID列表
	Owner      []string          `json:"owner,omitempty"`      // 所有者
	Tags       []string          `json:"tags,omitempty"`       // 标签
	Category   []string          `json:"category,omitempty"`   // 分类
	Sensitivity []string         `json:"sensitivity,omitempty"` // 敏感级别
	Attributes map[string]interface{} `json:"attributes,omitempty"` // 自定义属性
}

// ActionMatcher 动作匹配器
type ActionMatcher struct {
	Type     []string          `json:"type,omitempty"`     // 动作类型
	Method   []string          `json:"method,omitempty"`   // HTTP方法
	Attributes map[string]interface{} `json:"attributes,omitempty"` // 自定义属性
}

// EnvironmentMatcher 环境匹配器
type EnvironmentMatcher struct {
	Time       []TimeRange       `json:"time,omitempty"`       // 时间范围
	IP         []IPRange         `json:"ip,omitempty"`         // IP范围
	Device     []string          `json:"device,omitempty"`     // 设备类型
	Location   []string          `json:"location,omitempty"`   // 地理位置
	Attributes map[string]interface{} `json:"attributes,omitempty"` // 自定义属性
}

// ConditionExpression 条件表达式
type ConditionExpression struct {
	Variable   string      `json:"variable"`   // 变量名
	Operator   string      `json:"operator"`   // 操作符
	Value      interface{} `json:"value"`      // 值
	LogicalOp  string      `json:"logical_op"` // 逻辑操作符
}

// TimeRange 时间范围
type TimeRange struct {
	Start string `json:"start"` // HH:MM
	End   string `json:"end"`   // HH:MM
}

// IPRange IP范围
type IPRange struct {
	Start string `json:"start"` // 起始IP
	End   string `json:"end"`   // 结束IP
	Mask  int    `json:"mask"`  // 子网掩码
}

// AccessRequest 访问请求
type AccessRequest struct {
	Subject     UserContext      `json:"subject"`
	Resource    ResourceContext  `json:"resource"`
	Action      ActionContext    `json:"action"`
	Environment EnvironmentCtx   `json:"environment"`
	RequestID   string           `json:"request_id"`
	Timestamp   time.Time        `json:"timestamp"`
}

// UserContext 用户上下文
type UserContext struct {
	ID          uint                   `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Roles       []string               `json:"roles"`
	Groups      []string               `json:"groups"`
	Attributes  map[string]interface{} `json:"attributes"`
	TenantID    string                 `json:"tenant_id"`
	Active      bool                   `json:"active"`
}

// ResourceContext 资源上下文
type ResourceContext struct {
	Type         string                 `json:"type"`
	ID           string                 `json:"id"`
	Owner        string                 `json:"owner"`
	TenantID     string                 `json:"tenant_id"`
	Attributes   map[string]interface{} `json:"attributes"`
	Sensitivity  string                 `json:"sensitivity"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ActionContext 动作上下文
type ActionContext struct {
	Type       string                 `json:"type"`
	Method     string                 `json:"method"`
	Attributes map[string]interface{} `json:"attributes"`
}

// EnvironmentCtx 环境上下文
type EnvironmentCtx struct {
	Time       time.Time              `json:"time"`
	IP         string                 `json:"ip"`
	UserAgent  string                 `json:"user_agent"`
	Device     string                 `json:"device"`
	Location   string                 `json:"location"`
	Attributes map[string]interface{} `json:"attributes"`
}

// AccessDecision 访问决策
type AccessDecision struct {
	Allowed     bool                   `json:"allowed"`
	Effect      string                 `json:"effect"`
	Reason      string                 `json:"reason"`
	PolicyID    string                 `json:"policy_id,omitempty"`
	PolicyName  string                 `json:"policy_name,omitempty"`
	Duration    time.Duration          `json:"duration"`
	TTL         time.Duration          `json:"ttl"`
	Attributes  map[string]interface{} `json:"attributes"`
	Obligations []Obligation           `json:"obligations"`
}

// Obligation 义务
type Obligation struct {
	Type       string                 `json:"type"`
	Attributes map[string]interface{} `json:"attributes"`
}

// Evaluate 评估访问请求
func (e *ABACEngine) Evaluate(ctx context.Context, req *AccessRequest) (*AccessDecision, error) {
	startTime := time.Now()

	e.logger.WithFields(logrus.Fields{
		"request_id": req.RequestID,
		"user_id":    req.Subject.ID,
		"resource":   fmt.Sprintf("%s:%s", req.Resource.Type, req.Resource.ID),
		"action":     req.Action.Type,
	}).Debug("Evaluating access request")

	// 刷新策略缓存
	if err := e.refreshPolicyCache(ctx, req.Subject.TenantID); err != nil {
		e.logger.WithError(err).Error("Failed to refresh policy cache")
		return nil, fmt.Errorf("failed to refresh policy cache: %w", err)
	}

	// 获取适用的策略
	applicablePolicies := e.getApplicablePolicies(req)

	if len(applicablePolicies) == 0 {
		e.logger.WithField("request_id", req.RequestID).Warn("No applicable policies found")
		return &AccessDecision{
			Allowed:    false,
			Effect:     "deny",
			Reason:     "No applicable policies found",
			Duration:   time.Since(startTime),
			Attributes: make(map[string]interface{}),
		}, nil
	}

	// 按优先级排序策略
	e.sortPoliciesByPriority(applicablePolicies)

	// 评估策略
	var lastAllowedPolicy *PolicyRule
	var deniedReasons []string

	for _, policy := range applicablePolicies {
		if !policy.Enabled {
			continue
		}

		match, err := e.evaluatePolicy(policy, req)
		if err != nil {
			e.logger.WithError(err).WithField("policy_id", policy.ID).Error("Failed to evaluate policy")
			continue
		}

		if match {
			if policy.Effect == "deny" {
				deniedReasons = append(deniedReasons, fmt.Sprintf("Policy '%s' denies access", policy.Name))
			} else {
				lastAllowedPolicy = policy
			}
		}
	}

	// 做出最终决策
	decision := &AccessDecision{
		Duration:   time.Since(startTime),
		Attributes: make(map[string]interface{}),
	}

	if lastAllowedPolicy != nil {
		decision.Allowed = true
		decision.Effect = "allow"
		decision.PolicyID = lastAllowedPolicy.ID
		decision.PolicyName = lastAllowedPolicy.Name
		decision.Reason = fmt.Sprintf("Allowed by policy '%s'", lastAllowedPolicy.Name)

		// 添加义务
		decision.Obligations = e.generateObligations(lastAllowedPolicy, req)
	} else {
		decision.Allowed = false
		decision.Effect = "deny"
		if len(deniedReasons) > 0 {
			decision.Reason = fmt.Sprintf("Access denied: %v", deniedReasons)
		} else {
			decision.Reason = "No matching allow policy found"
		}
	}

	// 记录审计日志
	e.logAccessDecision(ctx, req, decision)

	e.logger.WithFields(logrus.Fields{
		"request_id": req.RequestID,
		"allowed":    decision.Allowed,
		"effect":     decision.Effect,
		"reason":     decision.Reason,
		"duration":   decision.Duration.Milliseconds(),
	}).Info("Access decision made")

	return decision, nil
}

// getApplicablePolicies 获取适用的策略
func (e *ABACEngine) getApplicablePolicies(req *AccessRequest) []*PolicyRule {
	var policies []*PolicyRule

	for _, policy := range e.ruleCache {
		// 检查租户匹配
		if policy.TenantID != req.Subject.TenantID && policy.TenantID != "*" {
			continue
		}

		// 基本匹配检查
		if e.matchesSubject(policy.Subject, req.Subject) &&
			e.matchesResource(policy.Resource, req.Resource) &&
			e.matchesAction(policy.Action, req.Action) &&
			e.matchesEnvironment(policy.Environment, req.Environment) {
			policies = append(policies, policy)
		}
	}

	return policies
}

// matchesSubject 检查主体匹配
func (e *ABACEngine) matchesSubject(matcher SubjectMatcher, subject UserContext) bool {
	// 检查ID
	if len(matcher.ID) > 0 && !e.containsString(matcher.ID, fmt.Sprintf("%d", subject.ID)) {
		return false
	}

	// 检查角色
	if len(matcher.Roles) > 0 && !e.hasAnyRole(subject.Roles, matcher.Roles) {
		return false
	}

	// 检查用户组
	if len(matcher.Groups) > 0 && !e.hasAnyRole(subject.Groups, matcher.Groups) {
		return false
	}

	// 检查邮箱
	if len(matcher.Email) > 0 && !e.matchesAnyPattern(subject.Email, matcher.Email) {
		return false
	}

	// 检查自定义属性
	for key, expectedValue := range matcher.Attributes {
		if actualValue, exists := subject.Attributes[key]; !exists || !e.equals(expectedValue, actualValue) {
			return false
		}
	}

	return true
}

// matchesResource 检查资源匹配
func (e *ABACEngine) matchesResource(matcher ResourceMatcher, resource ResourceContext) bool {
	// 检查类型
	if len(matcher.Type) > 0 && !e.containsString(matcher.Type, resource.Type) {
		return false
	}

	// 检查ID
	if len(matcher.ID) > 0 && !e.containsString(matcher.ID, resource.ID) {
		return false
	}

	// 检查所有者
	if len(matcher.Owner) > 0 && !e.containsString(matcher.Owner, resource.Owner) {
		return false
	}

	// 检查标签
	if len(matcher.Tags) > 0 && !e.hasAnyTag(resource.Tags, matcher.Tags) {
		return false
	}

	// 检查分类
	if len(matcher.Category) > 0 && !e.containsString(matcher.Category, resource.Category) {
		return false
	}

	// 检查敏感级别
	if len(matcher.Sensitivity) > 0 && !e.containsString(matcher.Sensitivity, resource.Sensitivity) {
		return false
	}

	// 检查自定义属性
	for key, expectedValue := range matcher.Attributes {
		if actualValue, exists := resource.Attributes[key]; !exists || !e.equals(expectedValue, actualValue) {
			return false
		}
	}

	return true
}

// matchesAction 检查动作匹配
func (e *ABACEngine) matchesAction(matcher ActionMatcher, action ActionContext) bool {
	// 检查类型
	if len(matcher.Type) > 0 && !e.containsString(matcher.Type, action.Type) {
		return false
	}

	// 检查方法
	if len(matcher.Method) > 0 && !e.containsString(matcher.Method, action.Method) {
		return false
	}

	// 检查自定义属性
	for key, expectedValue := range matcher.Attributes {
		if actualValue, exists := action.Attributes[key]; !exists || !e.equals(expectedValue, actualValue) {
			return false
		}
	}

	return true
}

// matchesEnvironment 检查环境匹配
func (e *ABACEngine) matchesEnvironment(matcher EnvironmentMatcher, env EnvironmentCtx) bool {
	// 检查时间范围
	if len(matcher.Time) > 0 && !e.isInTimeRange(env.Time, matcher.Time) {
		return false
	}

	// 检查IP范围
	if len(matcher.IP) > 0 && !e.isInIPRange(env.IP, matcher.IP) {
		return false
	}

	// 检查设备类型
	if len(matcher.Device) > 0 && !e.containsString(matcher.Device, env.Device) {
		return false
	}

	// 检查地理位置
	if len(matcher.Location) > 0 && !e.containsString(matcher.Location, env.Location) {
		return false
	}

	// 检查自定义属性
	for key, expectedValue := range matcher.Attributes {
		if actualValue, exists := env.Attributes[key]; !exists || !e.equals(expectedValue, actualValue) {
			return false
		}
	}

	return true
}

// evaluatePolicy 评估单个策略
func (e *ABACEngine) evaluatePolicy(policy *PolicyRule, req *AccessRequest) (bool, error) {
	// 如果没有条件，直接返回匹配
	if len(policy.Conditions) == 0 {
		return true, nil
	}

	// 评估所有条件
	for i, condition := range policy.Conditions {
		match, err := e.evaluateCondition(condition, req)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate condition %d: %w", i, err)
		}

		if !match {
			// 如果是AND逻辑（默认），任何条件不匹配都返回false
			if condition.LogicalOp == "" || condition.LogicalOp == "and" {
				return false, nil
			}
		} else {
			// 如果是OR逻辑，任何条件匹配都返回true
			if condition.LogicalOp == "or" {
				return true, nil
			}
		}
	}

	return true, nil
}

// evaluateCondition 评估条件
func (e *ABACEngine) evaluateCondition(condition ConditionExpression, req *AccessRequest) (bool, error) {
	// 获取变量值
	value, err := e.getVariableValue(condition.Variable, req)
	if err != nil {
		return false, fmt.Errorf("failed to get variable value: %w", err)
	}

	// 根据操作符比较
	switch condition.Operator {
	case "eq", "equals":
		return e.equals(value, condition.Value), nil
	case "ne", "not_equals":
		return !e.equals(value, condition.Value), nil
	case "gt", "greater_than":
		return e.greaterThan(value, condition.Value), nil
	case "gte", "greater_than_equal":
		return e.greaterThanOrEqual(value, condition.Value), nil
	case "lt", "less_than":
		return e.lessThan(value, condition.Value), nil
	case "lte", "less_than_equal":
		return e.lessThanOrEqual(value, condition.Value), nil
	case "in":
		return e.isIn(value, condition.Value), nil
	case "not_in":
		return !e.isIn(value, condition.Value), nil
	case "contains":
		return e.contains(value, condition.Value), nil
	case "not_contains":
		return !e.contains(value, condition.Value), nil
	case "matches":
		return e.matches(value, condition.Value), nil
	case "not_matches":
		return !e.matches(value, condition.Value), nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", condition.Operator)
	}
}

// getVariableValue 获取变量值
func (e *ABACEngine) getVariableValue(variable string, req *AccessRequest) (interface{}, error) {
	switch variable {
	case "subject.id":
		return req.Subject.ID, nil
	case "subject.username":
		return req.Subject.Username, nil
	case "subject.email":
		return req.Subject.Email, nil
	case "subject.roles":
		return req.Subject.Roles, nil
	case "subject.tenant_id":
		return req.Subject.TenantID, nil
	case "resource.id":
		return req.Resource.ID, nil
	case "resource.type":
		return req.Resource.Type, nil
	case "resource.owner":
		return req.Resource.Owner, nil
	case "resource.tenant_id":
		return req.Resource.TenantID, nil
	case "resource.sensitivity":
		return req.Resource.Sensitivity, nil
	case "resource.category":
		return req.Resource.Category, nil
	case "action.type":
		return req.Action.Type, nil
	case "action.method":
		return req.Action.Method, nil
	case "environment.time":
		return req.Environment.Time, nil
	case "environment.ip":
		return req.Environment.IP, nil
	case "environment.device":
		return req.Environment.Device, nil
	case "environment.location":
		return req.Environment.Location, nil
	default:
		// 检查是否是属性路径
		if len(variable) > 0 && variable[0] == '.' {
			return e.getAttributeValue(variable[1:], req)
		}
		return nil, fmt.Errorf("unknown variable: %s", variable)
	}
}

// getAttributeValue 获取属性值
func (e *ABACEngine) getAttributeValue(path string, req *AccessRequest) (interface{}, error) {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid attribute path: %s", path)
	}

	contextType := parts[0]
	attributePath := strings.Join(parts[1:], ".")

	switch contextType {
	case "subject":
		return e.getNestedValue(req.Subject.Attributes, attributePath)
	case "resource":
		return e.getNestedValue(req.Resource.Attributes, attributePath)
	case "action":
		return e.getNestedValue(req.Action.Attributes, attributePath)
	case "environment":
		return e.getNestedValue(req.Environment.Attributes, attributePath)
	default:
		return nil, fmt.Errorf("unknown context type: %s", contextType)
	}
}

// 辅助方法
func (e *ABACEngine) containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str || s == "*" {
			return true
		}
	}
	return false
}

func (e *ABACEngine) hasAnyRole(roles, requiredRoles []string) bool {
	for _, role := range roles {
		for _, required := range requiredRoles {
			if role == required || required == "*" {
				return true
			}
		}
	}
	return false
}

func (e *ABACEngine) hasAnyTag(tags, requiredTags []string) bool {
	for _, tag := range tags {
		for _, required := range requiredTags {
			if tag == required || required == "*" {
				return true
			}
		}
	}
	return false
}

func (e *ABACEngine) matchesAnyPattern(str string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, str); matched {
			return true
		}
	}
	return false
}

func (e *ABACEngine) equals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (e *ABACEngine) greaterThan(a, b interface{}) bool {
	// 实现数值比较逻辑
	return false
}

func (e *ABACEngine) greaterThanOrEqual(a, b interface{}) bool {
	// 实现数值比较逻辑
	return false
}

func (e *ABACEngine) lessThan(a, b interface{}) bool {
	// 实现数值比较逻辑
	return false
}

func (e *ABACEngine) lessThanOrEqual(a, b interface{}) bool {
	// 实现数值比较逻辑
	return false
}

func (e *ABACEngine) isIn(a, b interface{}) bool {
	// 实现包含逻辑
	return false
}

func (e *ABACEngine) contains(a, b interface{}) bool {
	// 实现包含逻辑
	return false
}

func (e *ABACEngine) matches(a, b interface{}) bool {
	// 实现正则匹配逻辑
	return false
}

func (e *ABACEngine) isInTimeRange(t time.Time, ranges []TimeRange) bool {
	// 实现时间范围检查
	return true
}

func (e *ABACEngine) isInIPRange(ip string, ranges []IPRange) bool {
	// 实现IP范围检查
	return true
}

func (e *ABACEngine) getNestedValue(obj map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := obj

	for _, part := range parts {
		if next, exists := current[part]; exists {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				return next, nil
			}
		} else {
			return nil, fmt.Errorf("path not found: %s", path)
		}
	}

	return current, nil
}

func (e *ABACEngine) sortPoliciesByPriority(policies []*PolicyRule) {
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority < policies[j].Priority
	})
}

func (e *ABACEngine) generateObligations(policy *PolicyRule, req *AccessRequest) []Obligation {
	// 生成策略义务
	return []Obligation{}
}

func (e *ABACEngine) logAccessDecision(ctx context.Context, req *AccessRequest, decision *AccessDecision) {
	// 记录访问决策审计日志
	auditLog := &models.AuditLog{
		UserID:     req.Subject.ID,
		Action:     req.Action.Type,
		Resource:   fmt.Sprintf("%s:%s", req.Resource.Type, req.Resource.ID),
		Result:     decision.Effect,
		Reason:     decision.Reason,
		IPAddress:  req.Environment.IP,
		UserAgent:  req.Environment.UserAgent,
		TenantID:   req.Subject.TenantID,
		CreatedAt:  time.Now(),
	}

	if err := e.auditRepo.Create(ctx, auditLog); err != nil {
		e.logger.WithError(err).Error("Failed to log access decision")
	}
}

func (e *ABACEngine) refreshPolicyCache(ctx context.Context, tenantID string) error {
	// 检查缓存是否过期
	if time.Since(e.lastCacheSync) < e.cacheExpiry {
		return nil
	}

	e.logger.Debug("Refreshing policy cache")

	// 从数据库加载策略
	policies, err := e.policyRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}

	// 更新缓存
	e.ruleCache = make(map[string]*PolicyRule)
	for _, policy := range policies {
		rule := &PolicyRule{
			ID:          fmt.Sprintf("%d", policy.ID),
			Name:        policy.Name,
			Description: policy.Description,
			Enabled:     policy.Enabled,
			// 设置其他字段...
		}
		e.ruleCache[rule.ID] = rule
	}

	e.lastCacheSync = time.Now()
	return nil
}