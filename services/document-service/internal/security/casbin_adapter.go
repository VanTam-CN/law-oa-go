package security

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

// GormAdapter GORM数据库适配器，用于Casbin持久化策略
type GormAdapter struct {
	db     *gorm.DB
	logger *slog.Logger
	table  string
}

// CasbinRule Casbin规则数据模型
type CasbinRule struct {
	ID    uint   `gorm:"primaryKey"`
	PType string `gorm:"size:100;index"`
	V0    string `gorm:"size:100;index"`
	V1    string `gorm:"size:100;index"`
	V2    string `gorm:"size:100;index"`
	V3    string `gorm:"size:100;index"`
	V4    string `gorm:"size:100;index"`
	V5    string `gorm:"size:100;index"`
}

// TableName 指定表名
func (CasbinRule) TableName() string {
	return "casbin_rules"
}

// NewGormAdapter 创建GORM适配器
func NewGormAdapter(db *gorm.DB) persist.Adapter {
	adapter := &GormAdapter{
		db:     db,
		logger: slog.Default(),
		table:  "casbin_rules",
	}

	// 自动迁移表结构
	if err := db.AutoMigrate(&CasbinRule{}); err != nil {
		adapter.logger.With("error", err).Error("Casbin表迁移失败")
	}

	return adapter
}

// LoadPolicy 从数据库加载策略
func (a *GormAdapter) LoadPolicy(model model.Model) error {
	var rules []CasbinRule

	if err := a.db.Find(&rules).Error; err != nil {
		return fmt.Errorf("加载策略失败: %w", err)
	}

	for _, rule := range rules {
		line := strings.TrimSpace(fmt.Sprintf("%s, %s, %s, %s, %s, %s",
			rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4))

		if err := persist.LoadPolicyLine(line, model); err != nil {
			a.logger.With("error", err, "rule", line).Warn("加载策略行失败")
		}
	}

	a.logger.Info("策略加载完成", "rules_count", len(rules))
	return nil
}

// SavePolicy 保存策略到数据库
func (a *GormAdapter) SavePolicy(model model.Model) error {
	// 清除现有规则
	if err := a.db.Exec("DELETE FROM casbin_rules").Error; err != nil {
		return fmt.Errorf("清除现有规则失败: %w", err)
	}

	// 获取所有策略
	var rules []CasbinRule

	// 处理请求定义
	if p := model["p"]; p != nil {
		for _, policy := range p.Policy {
			if len(policy) >= 3 {
				rule := CasbinRule{
					PType: "p",
					V0:    policy[0],
					V1:    policy[1],
					V2:    policy[2],
				}
				if len(policy) > 3 {
					rule.V3 = policy[3]
				}
				if len(policy) > 4 {
					rule.V4 = policy[4]
				}
				if len(policy) > 5 {
					rule.V5 = policy[5]
				}
				rules = append(rules, rule)
			}
		}
	}

	// 处理角色定义
	if g := model["g"]; g != nil {
		for _, policy := range g.Policy {
			if len(policy) >= 2 {
				rule := CasbinRule{
					PType: "g",
					V0:    policy[0],
					V1:    policy[1],
				}
				if len(policy) > 2 {
					rule.V2 = policy[2]
				}
				rules = append(rules, rule)
			}
		}
	}

	// 批量插入规则
	if len(rules) > 0 {
		if err := a.db.CreateInBatches(rules, 100).Error; err != nil {
			return fmt.Errorf("批量插入规则失败: %w", err)
		}
	}

	a.logger.Info("策略保存完成", "rules_count", len(rules))
	return nil
}

// AddPolicy 添加策略
func (a *GormAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	if len(rule) < 2 {
		return fmt.Errorf("规则参数不足")
	}

	casbinRule := CasbinRule{
		PType: ptype,
		V0:    rule[0],
		V1:    rule[1],
	}

	if len(rule) > 2 {
		casbinRule.V2 = rule[2]
	}
	if len(rule) > 3 {
		casbinRule.V3 = rule[3]
	}
	if len(rule) > 4 {
		casbinRule.V4 = rule[4]
	}
	if len(rule) > 5 {
		casbinRule.V5 = rule[5]
	}

	if err := a.db.Create(&casbinRule).Error; err != nil {
		return fmt.Errorf("添加策略失败: %w", err)
	}

	a.logger.Debug("策略添加成功", "ptype", ptype, "rule", rule)
	return nil
}

// RemovePolicy 删除策略
func (a *GormAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	query := a.db.Where("ptype = ?", ptype)

	if len(rule) > 0 {
		query = query.Where("v0 = ?", rule[0])
	}
	if len(rule) > 1 {
		query = query.Where("v1 = ?", rule[1])
	}
	if len(rule) > 2 {
		query = query.Where("v2 = ?", rule[2])
	}
	if len(rule) > 3 {
		query = query.Where("v3 = ?", rule[3])
	}
	if len(rule) > 4 {
		query = query.Where("v4 = ?", rule[4])
	}
	if len(rule) > 5 {
		query = query.Where("v5 = ?", rule[5])
	}

	result := query.Delete(&CasbinRule{})
	if result.Error != nil {
		return fmt.Errorf("删除策略失败: %w", result.Error)
	}

	a.logger.Debug("策略删除成功", "ptype", ptype, "rule", rule, "affected_rows", result.RowsAffected)
	return nil
}

// RemoveFilteredPolicy 删除过滤的策略
func (a *GormAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	query := a.db.Where("ptype = ?", ptype)

	// 根据字段索引和值构建查询条件
	for i, value := range fieldValues {
		if value != "" {
			fieldName := a.getFieldName(fieldIndex + i)
			query = query.Where(fmt.Sprintf("%s = ?", fieldName), value)
		}
	}

	result := query.Delete(&CasbinRule{})
	if result.Error != nil {
		return fmt.Errorf("删除过滤策略失败: %w", result.Error)
	}

	a.logger.Debug("过滤策略删除成功", "ptype", ptype, "field_index", fieldIndex,
		"field_values", fieldValues, "affected_rows", result.RowsAffected)
	return nil
}

// getFieldName 根据字段索引获取字段名
func (a *GormAdapter) getFieldName(index int) string {
	fieldNames := []string{"v0", "v1", "v2", "v3", "v4", "v5"}
	if index >= 0 && index < len(fieldNames) {
		return fieldNames[index]
	}
	return "v0"
}

// GetFilteredPolicy 获取过滤的策略
func (a *GormAdapter) GetFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	query := a.db.Where("ptype = ?", ptype)

	// 根据字段索引和值构建查询条件
	for i, value := range fieldValues {
		if value != "" {
			fieldName := a.getFieldName(fieldIndex + i)
			query = query.Where(fmt.Sprintf("%s = ?", fieldName), value)
		}
	}

	var rules []CasbinRule
	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("查询过滤策略失败: %w", err)
	}

	// 转换为字符串切片
	var result [][]string
	for _, rule := range rules {
		policy := []string{rule.PType}
		if rule.V0 != "" {
			policy = append(policy, rule.V0)
		}
		if rule.V1 != "" {
			policy = append(policy, rule.V1)
		}
		if rule.V2 != "" {
			policy = append(policy, rule.V2)
		}
		if rule.V3 != "" {
			policy = append(policy, rule.V3)
		}
		if rule.V4 != "" {
			policy = append(policy, rule.V4)
		}
		if rule.V5 != "" {
			policy = append(policy, rule.V5)
		}
		result = append(result, policy)
	}

	return result, nil
}

// UpdatePolicy 更新策略
func (a *GormAdapter) UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	// 查找现有规则
	query := a.db.Where("ptype = ?", ptype)

	if len(oldRule) > 0 {
		query = query.Where("v0 = ?", oldRule[0])
	}
	if len(oldRule) > 1 {
		query = query.Where("v1 = ?", oldRule[1])
	}
	if len(oldRule) > 2 {
		query = query.Where("v2 = ?", oldRule[2])
	}

	var rule CasbinRule
	if err := query.First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("策略不存在")
		}
		return fmt.Errorf("查找策略失败: %w", err)
	}

	// 更新规则
	if len(newRule) > 0 {
		rule.V0 = newRule[0]
	}
	if len(newRule) > 1 {
		rule.V1 = newRule[1]
	}
	if len(newRule) > 2 {
		rule.V2 = newRule[2]
	}
	if len(newRule) > 3 {
		rule.V3 = newRule[3]
	}
	if len(newRule) > 4 {
		rule.V4 = newRule[4]
	}
	if len(newRule) > 5 {
		rule.V5 = newRule[5]
	}

	if err := a.db.Save(&rule).Error; err != nil {
		return fmt.Errorf("更新策略失败: %w", err)
	}

	a.logger.Debug("策略更新成功", "ptype", ptype, "old_rule", oldRule, "new_rule", newRule)
	return nil
}

// UpdatePolicies 批量更新策略
func (a *GormAdapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	// 开始事务
	tx := a.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, oldRule := range oldRules {
		if i >= len(newRules) {
			break
		}
		newRule := newRules[i]

		// 查找现有规则
		query := tx.Where("ptype = ?", ptype)
		if len(oldRule) > 0 {
			query = query.Where("v0 = ?", oldRule[0])
		}
		if len(oldRule) > 1 {
			query = query.Where("v1 = ?", oldRule[1])
		}
		if len(oldRule) > 2 {
			query = query.Where("v2 = ?", oldRule[2])
		}

		var rule CasbinRule
		if err := query.First(&rule).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue // 跳过不存在的规则
			}
			tx.Rollback()
			return fmt.Errorf("查找策略失败: %w", err)
		}

		// 更新规则
		if len(newRule) > 0 {
			rule.V0 = newRule[0]
		}
		if len(newRule) > 1 {
			rule.V1 = newRule[1]
		}
		if len(newRule) > 2 {
			rule.V2 = newRule[2]
		}
		if len(newRule) > 3 {
			rule.V3 = newRule[3]
		}
		if len(newRule) > 4 {
			rule.V4 = newRule[4]
		}
		if len(newRule) > 5 {
			rule.V5 = newRule[5]
		}

		if err := tx.Save(&rule).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("更新策略失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	a.logger.Debug("批量策略更新成功", "ptype", ptype, "rules_count", len(oldRules))
	return nil
}

// GetValuesForFieldInPolicy 获取策略字段的所有值
func (a *GormAdapter) GetValuesForFieldInPolicy(sec string, ptype string, fieldIndex int) ([]string, error) {
	fieldName := a.getFieldName(fieldIndex)

	var values []string
	err := a.db.Model(&CasbinRule{}).
		Where("ptype = ?", ptype).
		Where(fmt.Sprintf("%s != ''", fieldName)).
		Pluck(fmt.Sprintf("DISTINCT %s", fieldName), &values).Error

	if err != nil {
		return nil, fmt.Errorf("获取字段值失败: %w", err)
	}

	return values, nil
}

// GetPolicy 获取所有策略
func (a *GormAdapter) GetPolicy(sec string) ([][]string, error) {
	var rules []CasbinRule
	if err := a.db.Where("ptype IN ?", []string{"p", "g"}).Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("获取策略失败: %w", err)
	}

	var result [][]string
	for _, rule := range rules {
		policy := []string{rule.PType}
		if rule.V0 != "" {
			policy = append(policy, rule.V0)
		}
		if rule.V1 != "" {
			policy = append(policy, rule.V1)
		}
		if rule.V2 != "" {
			policy = append(policy, rule.V2)
		}
		if rule.V3 != "" {
			policy = append(policy, rule.V3)
		}
		if rule.V4 != "" {
			policy = append(policy, rule.V4)
		}
		if rule.V5 != "" {
			policy = append(policy, rule.V5)
		}
		result = append(result, policy)
	}

	return result, nil
}

// GetGroupingPolicy 获取分组策略
func (a *GormAdapter) GetGroupingPolicy(sec string) ([][]string, error) {
	return a.GetFilteredGroupingPolicy(sec, "g", 0, "")
}

// GetFilteredGroupingPolicy 获取过滤的分组策略
func (a *GormAdapter) GetFilteredGroupingPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	query := a.db.Where("ptype = ?", ptype)

	// 根据字段索引和值构建查询条件
	for i, value := range fieldValues {
		if value != "" {
			fieldName := a.getFieldName(fieldIndex + i)
			query = query.Where(fmt.Sprintf("%s = ?", fieldName), value)
		}
	}

	var rules []CasbinRule
	if err := query.Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("查询过滤分组策略失败: %w", err)
	}

	// 转换为字符串切片
	var result [][]string
	for _, rule := range rules {
		policy := []string{rule.PType}
		if rule.V0 != "" {
			policy = append(policy, rule.V0)
		}
		if rule.V1 != "" {
			policy = append(policy, rule.V1)
		}
		if rule.V2 != "" {
			policy = append(policy, rule.V2)
		}
		result = append(result, policy)
	}

	return result, nil
}

// AddGroupingPolicy 添加分组策略
func (a *GormAdapter) AddGroupingPolicy(sec string, ptype string, rule []string) error {
	if len(rule) < 2 {
		return fmt.Errorf("分组策略参数不足")
	}

	casbinRule := CasbinRule{
		PType: ptype,
		V0:    rule[0],
		V1:    rule[1],
	}

	if len(rule) > 2 {
		casbinRule.V2 = rule[2]
	}

	if err := a.db.Create(&casbinRule).Error; err != nil {
		return fmt.Errorf("添加分组策略失败: %w", err)
	}

	a.logger.Debug("分组策略添加成功", "ptype", ptype, "rule", rule)
	return nil
}

// RemoveGroupingPolicy 删除分组策略
func (a *GormAdapter) RemoveGroupingPolicy(sec string, ptype string, rule []string) error {
	query := a.db.Where("ptype = ?", ptype)

	if len(rule) > 0 {
		query = query.Where("v0 = ?", rule[0])
	}
	if len(rule) > 1 {
		query = query.Where("v1 = ?", rule[1])
	}
	if len(rule) > 2 {
		query = query.Where("v2 = ?", rule[2])
	}

	result := query.Delete(&CasbinRule{})
	if result.Error != nil {
		return fmt.Errorf("删除分组策略失败: %w", result.Error)
	}

	a.logger.Debug("分组策略删除成功", "ptype", ptype, "rule", rule, "affected_rows", result.RowsAffected)
	return nil
}

// RemoveFilteredGroupingPolicy 删除过滤的分组策略
func (a *GormAdapter) RemoveFilteredGroupingPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return a.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...)
}

// UpdateGroupingPolicy 更新分组策略
func (a *GormAdapter) UpdateGroupingPolicy(sec string, ptype string, oldRule, newRule []string) error {
	return a.UpdatePolicy(sec, ptype, oldRule, newRule)
}

// UpdateGroupingPolicies 批量更新分组策略
func (a *GormAdapter) UpdateGroupingPolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	return a.UpdatePolicies(sec, ptype, oldRules, newRules)
}

// AddPolicies 批量添加策略
func (a *GormAdapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}

	var casbinRules []CasbinRule
	for _, rule := range rules {
		if len(rule) < 2 {
			continue
		}

		casbinRule := CasbinRule{
			PType: ptype,
			V0:    rule[0],
			V1:    rule[1],
		}

		if len(rule) > 2 {
			casbinRule.V2 = rule[2]
		}
		if len(rule) > 3 {
			casbinRule.V3 = rule[3]
		}
		if len(rule) > 4 {
			casbinRule.V4 = rule[4]
		}
		if len(rule) > 5 {
			casbinRule.V5 = rule[5]
		}

		casbinRules = append(casbinRules, casbinRule)
	}

	if len(casbinRules) > 0 {
		if err := a.db.CreateInBatches(casbinRules, 100).Error; err != nil {
			return fmt.Errorf("批量添加策略失败: %w", err)
		}
	}

	a.logger.Debug("批量策略添加成功", "ptype", ptype, "rules_count", len(casbinRules))
	return nil
}

// RemovePolicies 批量删除策略
func (a *GormAdapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}

	// 开始事务
	tx := a.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, rule := range rules {
		query := tx.Where("ptype = ?", ptype)

		if len(rule) > 0 {
			query = query.Where("v0 = ?", rule[0])
		}
		if len(rule) > 1 {
			query = query.Where("v1 = ?", rule[1])
		}
		if len(rule) > 2 {
			query = query.Where("v2 = ?", rule[2])
		}

		if err := query.Delete(&CasbinRule{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除策略失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	a.logger.Debug("批量策略删除成功", "ptype", ptype, "rules_count", len(rules))
	return nil
}

// AddGroupingPolicies 批量添加分组策略
func (a *GormAdapter) AddGroupingPolicies(sec string, ptype string, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}

	var casbinRules []CasbinRule
	for _, rule := range rules {
		if len(rule) < 2 {
			continue
		}

		casbinRule := CasbinRule{
			PType: ptype,
			V0:    rule[0],
			V1:    rule[1],
		}

		if len(rule) > 2 {
			casbinRule.V2 = rule[2]
		}

		casbinRules = append(casbinRules, casbinRule)
	}

	if len(casbinRules) > 0 {
		if err := a.db.CreateInBatches(casbinRules, 100).Error; err != nil {
			return fmt.Errorf("批量添加分组策略失败: %w", err)
		}
	}

	a.logger.Debug("批量分组策略添加成功", "ptype", ptype, "rules_count", len(casbinRules))
	return nil
}

// RemoveGroupingPolicies 批量删除分组策略
func (a *GormAdapter) RemoveGroupingPolicies(sec string, ptype string, rules [][]string) error {
	if len(rules) == 0 {
		return nil
	}

	// 开始事务
	tx := a.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, rule := range rules {
		query := tx.Where("ptype = ?", ptype)

		if len(rule) > 0 {
			query = query.Where("v0 = ?", rule[0])
		}
		if len(rule) > 1 {
			query = query.Where("v1 = ?", rule[1])
		}
		if len(rule) > 2 {
			query = query.Where("v2 = ?", rule[2])
		}

		if err := query.Delete(&CasbinRule{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("删除分组策略失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	a.logger.Debug("批量分组策略删除成功", "ptype", ptype, "rules_count", len(rules))
	return nil
}