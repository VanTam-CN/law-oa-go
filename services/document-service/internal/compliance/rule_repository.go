package compliance

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// MemoryRuleRepository 内存规则仓库实现
type MemoryRuleRepository struct {
	rules map[string]*ComplianceRule
	mutex sync.RWMutex
	logger *slog.Logger
}

// NewMemoryRuleRepository 创建内存规则仓库
func NewMemoryRuleRepository(logger *slog.Logger) RuleRepository {
	return &MemoryRuleRepository{
		rules:  make(map[string]*ComplianceRule),
		logger: logger,
	}
}

// Save 保存规则
func (r *MemoryRuleRepository) Save(ctx context.Context, rule *ComplianceRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.rules[rule.RuleID] = rule
	r.logger.Info("规则已保存", "rule_id", rule.RuleID, "rule_name", rule.RuleName)
	return nil
}

// Find 查找规则
func (r *MemoryRuleRepository) Find(ctx context.Context, ruleID string) (*ComplianceRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	rule, exists := r.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("规则不存在: %s", ruleID)
	}

	// 返回深拷贝避免并发问题
	return r.cloneRule(rule), nil
}

// FindAll 查找规则列表
func (r *MemoryRuleRepository) FindAll(ctx context.Context, filter *RuleFilter) ([]*ComplianceRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var results []*ComplianceRule

	for _, rule := range r.rules {
		if r.matchesFilter(rule, filter) {
			results = append(results, r.cloneRule(rule))
		}
	}

	return results, nil
}

// Update 更新规则
func (r *MemoryRuleRepository) Update(ctx context.Context, rule *ComplianceRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.rules[rule.RuleID]; !exists {
		return fmt.Errorf("规则不存在: %s", rule.RuleID)
	}

	r.rules[rule.RuleID] = rule
	r.logger.Info("规则已更新", "rule_id", rule.RuleID, "rule_name", rule.RuleName)
	return nil
}

// Delete 删除规则
func (r *MemoryRuleRepository) Delete(ctx context.Context, ruleID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.rules[ruleID]; !exists {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	delete(r.rules, ruleID)
	r.logger.Info("规则已删除", "rule_id", ruleID)
	return nil
}

// matchesFilter 检查规则是否匹配过滤器
func (r *MemoryRuleRepository) matchesFilter(rule *ComplianceRule, filter *RuleFilter) bool {
	if filter == nil {
		return true
	}

	// 检查分类
	if filter.Category != "" && rule.Category != filter.Category {
		return false
	}

	// 检查启用状态
	if filter.Enabled != nil && rule.Enabled != *filter.Enabled {
		return false
	}

	// 检查优先级
	if filter.Priority != nil && rule.Priority != *filter.Priority {
		return false
	}

	// 检查标签（如果实现的话）
	if len(filter.Tags) > 0 {
		// 简化实现：如果需要标签支持，可以扩展规则结构
	}

	return true
}

// cloneRule 克隆规则
func (r *MemoryRuleRepository) cloneRule(rule *ComplianceRule) *ComplianceRule {
	// 深拷贝规则数据
	clone := *rule

	// 克隆条件
	clone.Conditions = make([]RuleCondition, len(rule.Conditions))
	copy(clone.Conditions, rule.Conditions)

	// 克隆动作
	clone.Actions = make([]RuleAction, len(rule.Actions))
	copy(clone.Actions, rule.Actions)

	// 克隆元数据
	if rule.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range rule.Metadata {
			clone.Metadata[k] = v
		}
	}

	return &clone
}

// DatabaseRuleRepository 数据库规则仓库实现
type DatabaseRuleRepository struct {
	db     *sql.DB
	logger *slog.Logger
	mutex  sync.RWMutex
}

// NewDatabaseRuleRepository 创建数据库规则仓库
func NewDatabaseRuleRepository(db *sql.DB, logger *slog.Logger) RuleRepository {
	return &DatabaseRuleRepository{
		db:     db,
		logger: logger,
	}
}

// Save 保存规则
func (r *DatabaseRuleRepository) Save(ctx context.Context, rule *ComplianceRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 序列化条件和动作
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("序列化条件失败: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}

	metadataJSON, err := json.Marshal(rule.Metadata)
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %w", err)
	}

	query := `
		INSERT INTO compliance_rules (
			rule_id, rule_name, description, category, version, enabled,
			conditions, actions, priority, severity, creation_date, last_updated, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(rule_id) DO UPDATE SET
			rule_name = excluded.rule_name,
			description = excluded.description,
			category = excluded.category,
			version = excluded.version,
			enabled = excluded.enabled,
			conditions = excluded.conditions,
			actions = excluded.actions,
			priority = excluded.priority,
			severity = excluded.severity,
			last_updated = excluded.last_updated,
			metadata = excluded.metadata
	`

	_, err = r.db.ExecContext(ctx, query,
		rule.RuleID, rule.RuleName, rule.Description, rule.Category, rule.Version,
		rule.Enabled, string(conditionsJSON), string(actionsJSON), rule.Priority,
		rule.Severity, rule.CreationDate, rule.LastUpdated, string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("保存规则失败: %w", err)
	}

	r.logger.Info("规则已保存到数据库", "rule_id", rule.RuleID, "rule_name", rule.RuleName)
	return nil
}

// Find 查找规则
func (r *DatabaseRuleRepository) Find(ctx context.Context, ruleID string) (*ComplianceRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT rule_id, rule_name, description, category, version, enabled,
			   conditions, actions, priority, severity, creation_date, last_updated, metadata
		FROM compliance_rules
		WHERE rule_id = ?
	`

	var rule ComplianceRule
	var conditionsJSON, actionsJSON, metadataJSON string

	err := r.db.QueryRowContext(ctx, query, ruleID).Scan(
		&rule.RuleID, &rule.RuleName, &rule.Description, &rule.Category,
		&rule.Version, &rule.Enabled, &conditionsJSON, &actionsJSON,
		&rule.Priority, &rule.Severity, &rule.CreationDate, &rule.LastUpdated, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("规则不存在: %s", ruleID)
		}
		return nil, fmt.Errorf("查询规则失败: %w", err)
	}

	// 反序列化条件和动作
	if err := json.Unmarshal([]byte(conditionsJSON), &rule.Conditions); err != nil {
		return nil, fmt.Errorf("反序列化条件失败: %w", err)
	}

	if err := json.Unmarshal([]byte(actionsJSON), &rule.Actions); err != nil {
		return nil, fmt.Errorf("反序列化动作失败: %w", err)
	}

	if err := json.Unmarshal([]byte(metadataJSON), &rule.Metadata); err != nil {
		return nil, fmt.Errorf("反序列化元数据失败: %w", err)
	}

	return &rule, nil
}

// FindAll 查找规则列表
func (r *DatabaseRuleRepository) FindAll(ctx context.Context, filter *RuleFilter) ([]*ComplianceRule, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT rule_id, rule_name, description, category, version, enabled,
			   conditions, actions, priority, severity, creation_date, last_updated, metadata
		FROM compliance_rules
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	conditions := make([]string, 0)

	// 构建查询条件
	if filter != nil {
		if filter.Category != "" {
			conditions = append(conditions, "category = ?")
			args = append(args, filter.Category)
		}

		if filter.Enabled != nil {
			conditions = append(conditions, "enabled = ?")
			args = append(args, *filter.Enabled)
		}

		if filter.Priority != nil {
			conditions = append(conditions, "priority = ?")
			args = append(args, *filter.Priority)
		}
	}

	// 添加查询条件
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// 添加排序
	query += " ORDER BY priority DESC, last_updated DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询规则列表失败: %w", err)
	}
	defer rows.Close()

	var rules []*ComplianceRule

	for rows.Next() {
		var rule ComplianceRule
		var conditionsJSON, actionsJSON, metadataJSON string

		err := rows.Scan(
			&rule.RuleID, &rule.RuleName, &rule.Description, &rule.Category,
			&rule.Version, &rule.Enabled, &conditionsJSON, &actionsJSON,
			&rule.Priority, &rule.Severity, &rule.CreationDate, &rule.LastUpdated, &metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("扫描规则数据失败: %w", err)
		}

		// 反序列化数据和动作
		if err := json.Unmarshal([]byte(conditionsJSON), &rule.Conditions); err != nil {
			r.logger.Warn("反序列化条件失败", "rule_id", rule.RuleID, "error", err)
			rule.Conditions = []RuleCondition{}
		}

		if err := json.Unmarshal([]byte(actionsJSON), &rule.Actions); err != nil {
			r.logger.Warn("反序列化动作失败", "rule_id", rule.RuleID, "error", err)
			rule.Actions = []RuleAction{}
		}

		if err := json.Unmarshal([]byte(metadataJSON), &rule.Metadata); err != nil {
			r.logger.Warn("反序列化元数据失败", "rule_id", rule.RuleID, "error", err)
			rule.Metadata = make(map[string]interface{})
		}

		rules = append(rules, &rule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历查询结果失败: %w", err)
	}

	return rules, nil
}

// Update 更新规则
func (r *DatabaseRuleRepository) Update(ctx context.Context, rule *ComplianceRule) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 序列化条件和动作
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("序列化条件失败: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("序列化动作失败: %w", err)
	}

	metadataJSON, err := json.Marshal(rule.Metadata)
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %w", err)
	}

	query := `
		UPDATE compliance_rules
		SET rule_name = ?, description = ?, category = ?, version = ?,
			enabled = ?, conditions = ?, actions = ?, priority = ?,
			severity = ?, last_updated = ?, metadata = ?
		WHERE rule_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		rule.RuleName, rule.Description, rule.Category, rule.Version,
		rule.Enabled, string(conditionsJSON), string(actionsJSON), rule.Priority,
		rule.Severity, rule.LastUpdated, string(metadataJSON), rule.RuleID,
	)

	if err != nil {
		return fmt.Errorf("更新规则失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取受影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("规则不存在: %s", rule.RuleID)
	}

	r.logger.Info("规则已更新到数据库", "rule_id", rule.RuleID, "rule_name", rule.RuleName)
	return nil
}

// Delete 删除规则
func (r *DatabaseRuleRepository) Delete(ctx context.Context, ruleID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	query := `DELETE FROM compliance_rules WHERE rule_id = ?`

	result, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return fmt.Errorf("删除规则失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取受影响行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("规则不存在: %s", ruleID)
	}

	r.logger.Info("规则已从数据库删除", "rule_id", ruleID)
	return nil
}

// CreateComplianceRulesTable 创建合规规则表
func CreateComplianceRulesTable(ctx context.Context, db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS compliance_rules (
		rule_id TEXT PRIMARY KEY,
		rule_name TEXT NOT NULL,
		description TEXT,
		category TEXT NOT NULL,
		version TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT TRUE,
		conditions TEXT NOT NULL,
		actions TEXT NOT NULL,
		priority INTEGER NOT NULL DEFAULT 0,
		severity TEXT NOT NULL DEFAULT 'MEDIUM',
		creation_date DATETIME NOT NULL,
		last_updated DATETIME NOT NULL,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_compliance_rules_category ON compliance_rules(category);
	CREATE INDEX IF NOT EXISTS idx_compliance_rules_enabled ON compliance_rules(enabled);
	CREATE INDEX IF NOT EXISTS idx_compliance_rules_priority ON compliance_rules(priority);
	CREATE INDEX IF NOT EXISTS idx_compliance_rules_severity ON compliance_rules(severity);
	CREATE INDEX IF NOT EXISTS idx_compliance_rules_updated_at ON compliance_rules(last_updated);
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("创建合规规则表失败: %w", err)
	}

	return nil
}

// RuleTemplate 规则模板
type RuleTemplate struct {
	TemplateID   string                 `json:"template_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Template     *ComplianceRule        `json:"template"`
	Parameters   []RuleTemplateParameter `json:"parameters"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// RuleTemplateParameter 规则模板参数
type RuleTemplateParameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`     // "string", "number", "boolean", "select"
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Options      []string    `json:"options,omitempty"` // 用于select类型
	Validation   map[string]interface{} `json:"validation,omitempty"`
}

// RuleTemplateRepository 规则模板仓库接口
type RuleTemplateRepository interface {
	Save(ctx context.Context, template *RuleTemplate) error
	Find(ctx context.Context, templateID string) (*RuleTemplate, error)
	FindAll(ctx context.Context, category string) ([]*RuleTemplate, error)
	Update(ctx context.Context, template *RuleTemplate) error
	Delete(ctx context.Context, templateID string) error
}

// MemoryRuleTemplateRepository 内存规则模板仓库
type MemoryRuleTemplateRepository struct {
	templates map[string]*RuleTemplate
	mutex     sync.RWMutex
	logger    *slog.Logger
}

// NewMemoryRuleTemplateRepository 创建内存规则模板仓库
func NewMemoryRuleTemplateRepository(logger *slog.Logger) RuleTemplateRepository {
	return &MemoryRuleTemplateRepository{
		templates: make(map[string]*RuleTemplate),
		logger:    logger,
	}
}

// Save 保存模板
func (r *MemoryRuleTemplateRepository) Save(ctx context.Context, template *RuleTemplate) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}
	template.UpdatedAt = time.Now()

	r.templates[template.TemplateID] = template
	r.logger.Info("规则模板已保存", "template_id", template.TemplateID, "name", template.Name)
	return nil
}

// Find 查找模板
func (r *MemoryRuleTemplateRepository) Find(ctx context.Context, templateID string) (*RuleTemplate, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	template, exists := r.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("规则模板不存在: %s", templateID)
	}

	// 返回深拷贝
	return r.cloneTemplate(template), nil
}

// FindAll 查找模板列表
func (r *MemoryRuleTemplateRepository) FindAll(ctx context.Context, category string) ([]*RuleTemplate, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var results []*RuleTemplate

	for _, template := range r.templates {
		if category == "" || template.Category == category {
			results = append(results, r.cloneTemplate(template))
		}
	}

	return results, nil
}

// Update 更新模板
func (r *MemoryRuleTemplateRepository) Update(ctx context.Context, template *RuleTemplate) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.templates[template.TemplateID]; !exists {
		return fmt.Errorf("规则模板不存在: %s", template.TemplateID)
	}

	template.UpdatedAt = time.Now()
	r.templates[template.TemplateID] = template
	r.logger.Info("规则模板已更新", "template_id", template.TemplateID, "name", template.Name)
	return nil
}

// Delete 删除模板
func (r *MemoryRuleTemplateRepository) Delete(ctx context.Context, templateID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.templates[templateID]; !exists {
		return fmt.Errorf("规则模板不存在: %s", templateID)
	}

	delete(r.templates, templateID)
	r.logger.Info("规则模板已删除", "template_id", templateID)
	return nil
}

// cloneTemplate 克隆模板
func (r *MemoryRuleTemplateRepository) cloneTemplate(template *RuleTemplate) *RuleTemplate {
	clone := *template

	// 克隆模板规则
	if template.Template != nil {
		clone.Template = r.cloneRule(template.Template)
	}

	// 克隆参数
	clone.Parameters = make([]RuleTemplateParameter, len(template.Parameters))
	copy(clone.Parameters, template.Parameters)

	return &clone
}

// cloneRule 克隆规则（模板使用）
func (r *MemoryRuleTemplateRepository) cloneRule(rule *ComplianceRule) *ComplianceRule {
	clone := *rule

	// 克隆条件
	clone.Conditions = make([]RuleCondition, len(rule.Conditions))
	copy(clone.Conditions, rule.Conditions)

	// 克隆动作
	clone.Actions = make([]RuleAction, len(rule.Actions))
	copy(clone.Actions, rule.Actions)

	// 克隆元数据
	if rule.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range rule.Metadata {
			clone.Metadata[k] = v
		}
	}

	return &clone
}