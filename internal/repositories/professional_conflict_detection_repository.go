package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"

	"law-oa-go/internal/models"
)

// 专业冲突检测仓库实现
func (r *enhancedConflictRepository) CreateConflictCheckRequest(ctx context.Context, request *models.ProfessionalConflictCheckRequest) error {
	if err := r.db.WithContext(ctx).Create(request).Error; err != nil {
		return fmt.Errorf("创建冲突检查请求失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetConflictCheckRequest(ctx context.Context, id string) (*models.ProfessionalConflictCheckRequest, error) {
	var request models.ProfessionalConflictCheckRequest
	if err := r.db.WithContext(ctx).
		Preload("ConflictResults").
		Preload("RuleExecutions").
		First(&request, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("冲突检查请求不存在: %s", id)
		}
		return nil, fmt.Errorf("获取冲突检查请求失败: %w", err)
	}
	return &request, nil
}

func (r *enhancedConflictRepository) GetConflictCheckRequestByNumber(ctx context.Context, checkNumber string) (*models.ProfessionalConflictCheckRequest, error) {
	var request models.ProfessionalConflictCheckRequest
	if err := r.db.WithContext(ctx).
		Preload("ConflictResults").
		Preload("RuleExecutions").
		First(&request, "check_number = ?", checkNumber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("冲突检查请求不存在: %s", checkNumber)
		}
		return nil, fmt.Errorf("获取冲突检查请求失败: %w", err)
	}
	return &request, nil
}

func (r *enhancedConflictRepository) UpdateConflictCheckRequest(ctx context.Context, request *models.ProfessionalConflictCheckRequest) error {
	if err := r.db.WithContext(ctx).Save(request).Error; err != nil {
		return fmt.Errorf("更新冲突检查请求失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) DeleteConflictCheckRequest(ctx context.Context, id string) error {
	return fmt.Errorf("冲突检查请求属于审计证据，不允许删除；请通过追加更正或撤销记录处理")
}

func (r *enhancedConflictRepository) GetConflictCheckRequestsByStatus(ctx context.Context, status string) ([]*models.ProfessionalConflictCheckRequest, error) {
	var requests []*models.ProfessionalConflictCheckRequest
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("根据状态获取冲突检查请求失败: %w", err)
	}
	return requests, nil
}

func (r *enhancedConflictRepository) GetConflictCheckRequestsByPriority(ctx context.Context, priority string) ([]*models.ProfessionalConflictCheckRequest, error) {
	var requests []*models.ProfessionalConflictCheckRequest
	if err := r.db.WithContext(ctx).
		Where("priority = ?", priority).
		Order("created_at DESC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("根据优先级获取冲突检查请求失败: %w", err)
	}
	return requests, nil
}

func (r *enhancedConflictRepository) GetConflictCheckRequestsByAssignee(ctx context.Context, assigneeID string) ([]*models.ProfessionalConflictCheckRequest, error) {
	var requests []*models.ProfessionalConflictCheckRequest
	if err := r.db.WithContext(ctx).
		Where("assigned_to = ?", assigneeID).
		Order("priority DESC, created_at ASC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("根据分配人获取冲突检查请求失败: %w", err)
	}
	return requests, nil
}

func (r *enhancedConflictRepository) GetConflictCheckRequestsByDateRange(ctx context.Context, start, end time.Time) ([]*models.ProfessionalConflictCheckRequest, error) {
	var requests []*models.ProfessionalConflictCheckRequest
	if err := r.db.WithContext(ctx).
		Where("requested_date BETWEEN ? AND ?", start, end).
		Order("requested_date DESC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("根据日期范围获取冲突检查请求失败: %w", err)
	}
	return requests, nil
}

func (r *enhancedConflictRepository) GetOverdueConflictCheckRequests(ctx context.Context) ([]*models.ProfessionalConflictCheckRequest, error) {
	var requests []*models.ProfessionalConflictCheckRequest
	if err := r.db.WithContext(ctx).
		Where("status IN ('PENDING', 'IN_PROGRESS') AND required_by_date < ?", time.Now()).
		Order("required_by_date ASC").
		Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("获取逾期冲突检查请求失败: %w", err)
	}
	return requests, nil
}

// 多维度冲突结果管理
func (r *enhancedConflictRepository) CreateMultidimensionalConflictResult(ctx context.Context, result *models.MultidimensionalConflictResult) error {
	if err := r.db.WithContext(ctx).Create(result).Error; err != nil {
		return fmt.Errorf("创建多维度冲突结果失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetMultidimensionalConflictResults(ctx context.Context, checkRequestID string) ([]*models.MultidimensionalConflictResult, error) {
	var results []*models.MultidimensionalConflictResult
	if err := r.db.WithContext(ctx).
		Where("check_request_id = ?", checkRequestID).
		Order("severity_level DESC, conflict_type").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("获取多维度冲突结果失败: %w", err)
	}
	return results, nil
}

func (r *enhancedConflictRepository) GetMultidimensionalConflictResult(ctx context.Context, id string) (*models.MultidimensionalConflictResult, error) {
	var result models.MultidimensionalConflictResult
	if err := r.db.WithContext(ctx).
		First(&result, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("多维度冲突结果不存在: %s", id)
		}
		return nil, fmt.Errorf("获取多维度冲突结果失败: %w", err)
	}
	return &result, nil
}

func (r *enhancedConflictRepository) UpdateMultidimensionalConflictResult(ctx context.Context, result *models.MultidimensionalConflictResult) error {
	if err := r.db.WithContext(ctx).Save(result).Error; err != nil {
		return fmt.Errorf("更新多维度冲突结果失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) DeleteMultidimensionalConflictResult(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.MultidimensionalConflictResult{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除多维度冲突结果失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetConflictsByType(ctx context.Context, conflictType string) ([]*models.MultidimensionalConflictResult, error) {
	var results []*models.MultidimensionalConflictResult
	if err := r.db.WithContext(ctx).
		Where("conflict_type = ?", conflictType).
		Order("created_at DESC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("根据类型获取冲突结果失败: %w", err)
	}
	return results, nil
}

func (r *enhancedConflictRepository) GetConflictsBySeverity(ctx context.Context, severity string) ([]*models.MultidimensionalConflictResult, error) {
	var results []*models.MultidimensionalConflictResult
	if err := r.db.WithContext(ctx).
		Where("severity_level = ?", severity).
		Order("created_at DESC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("根据严重程度获取冲突结果失败: %w", err)
	}
	return results, nil
}

func (r *enhancedConflictRepository) GetConflictsByStatus(ctx context.Context, status string) ([]*models.MultidimensionalConflictResult, error) {
	var results []*models.MultidimensionalConflictResult
	if err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("根据状态获取冲突结果失败: %w", err)
	}
	return results, nil
}

func (r *enhancedConflictRepository) GetConflictsByEntity(ctx context.Context, entityType, entityID string) ([]*models.MultidimensionalConflictResult, error) {
	var results []*models.MultidimensionalConflictResult
	if err := r.db.WithContext(ctx).
		Where("JSON_CONTAINS(primary_entity, ?, '$.type') AND JSON_CONTAINS(primary_entity, ?, '$.id')",
			fmt.Sprintf(`"%s"`, entityType), fmt.Sprintf(`"%s"`, entityID)).
		Or("JSON_CONTAINS(secondary_entity, ?, '$.type') AND JSON_CONTAINS(secondary_entity, ?, '$.id')",
			fmt.Sprintf(`"%s"`, entityType), fmt.Sprintf(`"%s"`, entityID)).
		Order("created_at DESC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("根据实体获取冲突结果失败: %w", err)
	}
	return results, nil
}

func (r *enhancedConflictRepository) GetConflictsRequiringMonitoring(ctx context.Context) ([]*models.MultidimensionalConflictResult, error) {
	var results []*models.MultidimensionalConflictResult
	if err := r.db.WithContext(ctx).
		Where("monitoring_required = ? AND status IN ('RESOLVED', 'WAVED')", true).
		Where("next_review_date IS NOT NULL AND next_review_date <= ?", time.Now().AddDate(0, 0, 7)).
		Order("next_review_date ASC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("获取需要监控的冲突结果失败: %w", err)
	}
	return results, nil
}

// 冲突检测规则管理
func (r *enhancedConflictRepository) CreateConflictDetectionRule(ctx context.Context, rule *models.ConflictDetectionRule) error {
	if err := r.db.WithContext(ctx).Create(rule).Error; err != nil {
		return fmt.Errorf("创建冲突检测规则失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetConflictDetectionRule(ctx context.Context, id string) (*models.ConflictDetectionRule, error) {
	var rule models.ConflictDetectionRule
	if err := r.db.WithContext(ctx).
		Preload("RuleExecutions").
		First(&rule, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("冲突检测规则不存在: %s", id)
		}
		return nil, fmt.Errorf("获取冲突检测规则失败: %w", err)
	}
	return &rule, nil
}

func (r *enhancedConflictRepository) GetConflictDetectionRuleByCode(ctx context.Context, ruleCode string) (*models.ConflictDetectionRule, error) {
	var rule models.ConflictDetectionRule
	if err := r.db.WithContext(ctx).
		First(&rule, "rule_code = ?", ruleCode).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("冲突检测规则不存在: %s", ruleCode)
		}
		return nil, fmt.Errorf("获取冲突检测规则失败: %w", err)
	}
	return &rule, nil
}

func (r *enhancedConflictRepository) UpdateConflictDetectionRule(ctx context.Context, rule *models.ConflictDetectionRule) error {
	if err := r.db.WithContext(ctx).Save(rule).Error; err != nil {
		return fmt.Errorf("更新冲突检测规则失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) DeleteConflictDetectionRule(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.ConflictDetectionRule{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("删除冲突检测规则失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetActiveConflictDetectionRules(ctx context.Context) ([]*models.ConflictDetectionRule, error) {
	var rules []*models.ConflictDetectionRule
	if err := r.db.WithContext(ctx).
		Where("status = 'ACTIVE'").
		Order("priority DESC, rule_category, rule_name").
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("获取活跃冲突检测规则失败: %w", err)
	}
	return rules, nil
}

func (r *enhancedConflictRepository) GetConflictDetectionRulesByType(ctx context.Context, ruleType string) ([]*models.ConflictDetectionRule, error) {
	var rules []*models.ConflictDetectionRule
	if err := r.db.WithContext(ctx).
		Where("rule_type = ? AND status = 'ACTIVE'", ruleType).
		Order("priority DESC, rule_name").
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("根据类型获取冲突检测规则失败: %w", err)
	}
	return rules, nil
}

// 规则执行记录管理
func (r *enhancedConflictRepository) CreateConflictRuleExecution(ctx context.Context, execution *models.ConflictRuleExecution) error {
	if err := r.db.WithContext(ctx).Create(execution).Error; err != nil {
		return fmt.Errorf("创建冲突规则执行记录失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetConflictRuleExecutions(ctx context.Context, checkRequestID string) ([]*models.ConflictRuleExecution, error) {
	var executions []*models.ConflictRuleExecution
	if err := r.db.WithContext(ctx).
		Preload("Rule").
		Where("check_request_id = ?", checkRequestID).
		Order("execution_sequence").
		Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("获取冲突规则执行记录失败: %w", err)
	}
	return executions, nil
}

func (r *enhancedConflictRepository) GetConflictRuleExecutionsByRule(ctx context.Context, ruleID string) ([]*models.ConflictRuleExecution, error) {
	var executions []*models.ConflictRuleExecution
	if err := r.db.WithContext(ctx).
		Preload("CheckRequest").
		Where("rule_id = ?", ruleID).
		Order("executed_at DESC").
		Limit(100).
		Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("根据规则获取执行记录失败: %w", err)
	}
	return executions, nil
}

func (r *enhancedConflictRepository) UpdateConflictRuleExecution(ctx context.Context, execution *models.ConflictRuleExecution) error {
	if err := r.db.WithContext(ctx).Save(execution).Error; err != nil {
		return fmt.Errorf("更新冲突规则执行记录失败: %w", err)
	}
	return nil
}

func (r *enhancedConflictRepository) GetRuleExecutionStatistics(ctx context.Context, ruleID string) (*RuleExecutionStats, error) {
	var stats RuleExecutionStats

	// 获取总执行次数
	if err := r.db.WithContext(ctx).
		Model(&models.ConflictRuleExecution{}).
		Where("rule_id = ?", ruleID).
		Count(&stats.TotalExecutions).Error; err != nil {
		return nil, fmt.Errorf("获取规则执行统计失败: %w", err)
	}

	// 获取成功率
	var successCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.ConflictRuleExecution{}).
		Where("rule_id = ? AND execution_result = 'SUCCESS'", ruleID).
		Count(&successCount).Error; err != nil {
		return nil, fmt.Errorf("获取规则成功执行次数失败: %w", err)
	}

	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalExecutions) * 100
	}

	// 获取平均执行时间
	var avgTime sql.NullInt64
	if err := r.db.WithContext(ctx).
		Model(&models.ConflictRuleExecution{}).
		Where("rule_id = ? AND execution_result = 'SUCCESS'", ruleID).
		Select("AVG(execution_time)").
		Scan(&avgTime).Error; err != nil {
		return nil, fmt.Errorf("获取规则平均执行时间失败: %w", err)
	}

	if avgTime.Valid {
		stats.AvgExecutionTime = avgTime.Int64
	}

	// 获取最后执行时间
	var lastExecution models.ConflictRuleExecution
	if err := r.db.WithContext(ctx).
		Where("rule_id = ?", ruleID).
		Order("executed_at DESC").
		First(&lastExecution).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("获取规则最后执行时间失败: %w", err)
	}

	if lastExecution.ID != "" {
		stats.LastExecuted = &lastExecution.ExecutedAt
	}

	// 获取错误率
	var errorCount int64
	if err := r.db.WithContext(ctx).
		Model(&models.ConflictRuleExecution{}).
		Where("rule_id = ? AND execution_result = 'ERROR'", ruleID).
		Count(&errorCount).Error; err != nil {
		return nil, fmt.Errorf("获取规则错误执行次数失败: %w", err)
	}

	if stats.TotalExecutions > 0 {
		stats.ErrorRate = float64(errorCount) / float64(stats.TotalExecutions) * 100
	}

	return &stats, nil
}
