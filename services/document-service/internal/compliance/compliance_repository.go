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

// MemoryComplianceRepository 内存合规仓库实现
type MemoryComplianceRepository struct {
	results    map[string]*ComplianceCheckResult
	violations map[string][]Violation
	mutex      sync.RWMutex
	logger     *slog.Logger
}

// NewMemoryComplianceRepository 创建内存合规仓库
func NewMemoryComplianceRepository(logger *slog.Logger) ComplianceRepository {
	return &MemoryComplianceRepository{
		results:    make(map[string]*ComplianceCheckResult),
		violations: make(map[string][]Violation),
		logger:     logger,
	}
}

// SaveResult 保存检查结果
func (r *MemoryComplianceRepository) SaveResult(ctx context.Context, result *ComplianceCheckResult) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 保存检查结果
	r.results[result.RequestID] = result

	// 保存违规记录
	if len(result.Violations) > 0 {
		r.violations[result.SubjectID] = append(r.violations[result.SubjectID], result.Violations...)
	}

	r.logger.Info("合规检查结果已保存",
		"request_id", result.RequestID,
		"subject_id", result.SubjectID,
		"status", result.OverallStatus)

	return nil
}

// FindResult 查找检查结果
func (r *MemoryComplianceRepository) FindResult(ctx context.Context, requestID string) (*ComplianceCheckResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result, exists := r.results[requestID]
	if !exists {
		return nil, fmt.Errorf("检查结果不存在: %s", requestID)
	}

	// 返回深拷贝
	return r.cloneResult(result), nil
}

// FindHistory 查找历史结果
func (r *MemoryComplianceRepository) FindHistory(ctx context.Context, subjectID string, filter *HistoryFilter) ([]*ComplianceCheckResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var results []*ComplianceCheckResult

	for _, result := range r.results {
		if result.SubjectID != subjectID {
			continue
		}

		if r.matchesHistoryFilter(result, filter) {
			results = append(results, r.cloneResult(result))
		}
	}

	// 应用分页
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(results) {
			results = results[filter.Offset:]
		}

		if filter.Limit > 0 && filter.Limit < len(results) {
			results = results[:filter.Limit]
		}
	}

	return results, nil
}

// SaveViolation 保存违规记录
func (r *MemoryComplianceRepository) SaveViolation(ctx context.Context, violation *Violation) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	subjectID := r.extractSubjectIDFromViolation(violation)
	r.violations[subjectID] = append(r.violations[subjectID], *violation)

	r.logger.Info("违规记录已保存",
		"violation_id", violation.ViolationID,
		"rule_id", violation.RuleID,
		"severity", violation.Severity)

	return nil
}

// FindViolations 查找违规记录
func (r *MemoryComplianceRepository) FindViolations(ctx context.Context, filter *ViolationFilter) ([]Violation, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var violations []Violation

	for subjectID, subjectViolations := range r.violations {
		for _, violation := range subjectViolations {
			if r.matchesViolationFilter(violation, filter) {
				violations = append(violations, violation)
			}
		}
	}

	// 应用分页
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(violations) {
			violations = violations[filter.Offset:]
		}

		if filter.Limit > 0 && filter.Limit < len(violations) {
			violations = violations[:filter.Limit]
		}
	}

	return violations, nil
}

// matchesHistoryFilter 检查结果是否匹配历史过滤器
func (r *MemoryComplianceRepository) matchesHistoryFilter(result *ComplianceCheckResult, filter *HistoryFilter) bool {
	if filter == nil {
		return true
	}

	// 检查时间范围
	if filter.StartTime != nil && result.CheckTimestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && result.CheckTimestamp.After(*filter.EndTime) {
		return false
	}

	// 检查检查类型
	if filter.CheckType != "" && result.CheckType != filter.CheckType {
		return false
	}

	// 检查状态
	if filter.Status != "" && result.OverallStatus != filter.Status {
		return false
	}

	// 检查风险等级
	if filter.RiskLevel != "" && result.RiskLevel != filter.RiskLevel {
		return false
	}

	return true
}

// matchesViolationFilter 检查违规是否匹配过滤器
func (r *MemoryComplianceRepository) matchesViolationFilter(violation Violation, filter *ViolationFilter) bool {
	if filter == nil {
		return true
	}

	// 检查规则ID
	if filter.RuleID != "" && violation.RuleID != filter.RuleID {
		return false
	}

	// 检查严重级别
	if filter.Severity != "" && violation.Severity != filter.Severity {
		return false
	}

	// 检查状态
	if filter.Status != "" && violation.Status != filter.Status {
		return false
	}

	// 检查时间范围
	if filter.StartTime != nil && violation.DetectedAt.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && violation.DetectedAt.After(*filter.EndTime) {
		return false
	}

	// 检查分配人
	if filter.AssignedTo != "" {
		// 如果违规记录中有分配人信息
		if assignedTo, exists := violation.Evidence["assigned_to"]; exists {
			if assignedToStr, ok := assignedTo.(string); ok && assignedToStr != filter.AssignedTo {
				return false
			}
		}
	}

	return true
}

// extractSubjectIDFromViolation 从违规记录中提取主体ID
func (r *MemoryComplianceRepository) extractSubjectIDFromViolation(violation *Violation) string {
	// 简化实现：从证据中提取主体ID
	if subjectID, exists := violation.Evidence["subject_id"]; exists {
		if id, ok := subjectID.(string); ok {
			return id
		}
	}

	// 如果没有找到，返回一个默认值
	return "unknown"
}

// cloneResult 克隆检查结果
func (r *MemoryComplianceRepository) cloneResult(result *ComplianceCheckResult) *ComplianceCheckResult {
	clone := *result

	// 克隆违规记录
	clone.Violations = make([]Violation, len(result.Violations))
	copy(clone.Violations, result.Violations)

	// 克隆建议
	clone.Recommendations = make([]Recommendation, len(result.Recommendations))
	copy(clone.Recommendations, result.Recommendations)

	// 克隆必要行动
	clone.RequiredActions = make([]RequiredAction, len(result.RequiredActions))
	copy(clone.RequiredActions, result.RequiredActions)

	// 克隆元数据
	if result.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range result.Metadata {
			clone.Metadata[k] = v
		}
	}

	return &clone
}

// DatabaseComplianceRepository 数据库合规仓库实现
type DatabaseComplianceRepository struct {
	db     *sql.DB
	logger *slog.Logger
	mutex  sync.RWMutex
}

// NewDatabaseComplianceRepository 创建数据库合规仓库
func NewDatabaseComplianceRepository(db *sql.DB, logger *slog.Logger) ComplianceRepository {
	return &DatabaseComplianceRepository{
		db:     db,
		logger: logger,
	}
}

// SaveResult 保存检查结果
func (r *DatabaseComplianceRepository) SaveResult(ctx context.Context, result *ComplianceCheckResult) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 序列化数据
	violationsJSON, err := json.Marshal(result.Violations)
	if err != nil {
		return fmt.Errorf("序列化违规记录失败: %w", err)
	}

	recommendationsJSON, err := json.Marshal(result.Recommendations)
	if err != nil {
		return fmt.Errorf("序列化建议失败: %w", err)
	}

	actionsJSON, err := json.Marshal(result.RequiredActions)
	if err != nil {
		return fmt.Errorf("序列化必要行动失败: %w", err)
	}

	metadataJSON, err := json.Marshal(result.Metadata)
	if err != nil {
		return fmt.Errorf("序列化元数据失败: %w", err)
	}

	// 保存主记录
	query := `
		INSERT INTO compliance_check_results (
			request_id, check_type, subject_id, subject_type, overall_status,
			overall_score, risk_level, violations, recommendations, required_actions,
			check_timestamp, next_review_date, checked_by, processing_time, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(request_id) DO UPDATE SET
			overall_status = excluded.overall_status,
			overall_score = excluded.overall_score,
			risk_level = excluded.risk_level,
			violations = excluded.violations,
			recommendations = excluded.recommendations,
			required_actions = excluded.required_actions,
			next_review_date = excluded.next_review_date,
			processing_time = excluded.processing_time,
			metadata = excluded.metadata
	`

	_, err = tx.ExecContext(ctx, query,
		result.RequestID, result.CheckType, result.SubjectID, result.SubjectType,
		result.OverallStatus, result.OverallScore, result.RiskLevel,
		string(violationsJSON), string(recommendationsJSON), string(actionsJSON),
		result.CheckTimestamp, result.NextReviewDate, result.CheckedBy,
		result.ProcessingTime, string(metadataJSON),
	)

	if err != nil {
		return fmt.Errorf("保存检查结果失败: %w", err)
	}

	// 分别保存违规记录到专门的表
	for _, violation := range result.Violations {
		if err := r.saveViolationInTx(ctx, tx, result.RequestID, violation); err != nil {
			return fmt.Errorf("保存违规记录失败: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	r.logger.Info("合规检查结果已保存到数据库",
		"request_id", result.RequestID,
		"subject_id", result.SubjectID,
		"status", result.OverallStatus)

	return nil
}

// saveViolationInTx 在事务中保存违规记录
func (r *DatabaseComplianceRepository) saveViolationInTx(ctx context.Context, tx *sql.Tx, requestID string, violation Violation) error {
	evidenceJSON, err := json.Marshal(violation.Evidence)
	if err != nil {
		return fmt.Errorf("序列化证据失败: %w", err)
	}

	query := `
		INSERT INTO compliance_violations (
			violation_id, request_id, rule_id, rule_name, description,
			severity, detected_at, affected_resource, evidence, remediation, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(violation_id) DO UPDATE SET
			rule_name = excluded.rule_name,
			description = excluded.description,
			severity = excluded.severity,
			affected_resource = excluded.affected_resource,
			evidence = excluded.evidence,
			remediation = excluded.remediation,
			status = excluded.status
	`

	_, err = tx.ExecContext(ctx, query,
		violation.ViolationID, requestID, violation.RuleID, violation.RuleName,
		violation.Description, violation.Severity, violation.DetectedAt,
		violation.AffectedResource, string(evidenceJSON), violation.Remediation, violation.Status,
	)

	return err
}

// FindResult 查找检查结果
func (r *DatabaseComplianceRepository) FindResult(ctx context.Context, requestID string) (*ComplianceCheckResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT request_id, check_type, subject_id, subject_type, overall_status,
			   overall_score, risk_level, violations, recommendations, required_actions,
			   check_timestamp, next_review_date, checked_by, processing_time, metadata
		FROM compliance_check_results
		WHERE request_id = ?
	`

	var result ComplianceCheckResult
	var violationsJSON, recommendationsJSON, actionsJSON, metadataJSON string
	var nextReviewDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, requestID).Scan(
		&result.RequestID, &result.CheckType, &result.SubjectID, &result.SubjectType,
		&result.OverallStatus, &result.OverallScore, &result.RiskLevel,
		&violationsJSON, &recommendationsJSON, &actionsJSON,
		&result.CheckTimestamp, &nextReviewDate, &result.CheckedBy,
		&result.ProcessingTime, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("检查结果不存在: %s", requestID)
		}
		return nil, fmt.Errorf("查询检查结果失败: %w", err)
	}

	if nextReviewDate.Valid {
		result.NextReviewDate = &nextReviewDate.Time
	}

	// 反序列化数据
	if err := json.Unmarshal([]byte(violationsJSON), &result.Violations); err != nil {
		r.logger.Warn("反序列化违规记录失败", "request_id", requestID, "error", err)
		result.Violations = []Violation{}
	}

	if err := json.Unmarshal([]byte(recommendationsJSON), &result.Recommendations); err != nil {
		r.logger.Warn("反序列化建议失败", "request_id", requestID, "error", err)
		result.Recommendations = []Recommendation{}
	}

	if err := json.Unmarshal([]byte(actionsJSON), &result.RequiredActions); err != nil {
		r.logger.Warn("反序列化必要行动失败", "request_id", requestID, "error", err)
		result.RequiredActions = []RequiredAction{}
	}

	if err := json.Unmarshal([]byte(metadataJSON), &result.Metadata); err != nil {
		r.logger.Warn("反序列化元数据失败", "request_id", requestID, "error", err)
		result.Metadata = make(map[string]interface{})
	}

	return &result, nil
}

// FindHistory 查找历史结果
func (r *DatabaseComplianceRepository) FindHistory(ctx context.Context, subjectID string, filter *HistoryFilter) ([]*ComplianceCheckResult, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT request_id, check_type, subject_id, subject_type, overall_status,
			   overall_score, risk_level, violations, recommendations, required_actions,
			   check_timestamp, next_review_date, checked_by, processing_time, metadata
		FROM compliance_check_results
		WHERE subject_id = ?
	`

	args := []interface{}{subjectID}
	conditions := []string{"subject_id = ?"}

	// 构建查询条件
	if filter != nil {
		if filter.StartTime != nil {
			conditions = append(conditions, "check_timestamp >= ?")
			args = append(args, filter.StartTime)
		}

		if filter.EndTime != nil {
			conditions = append(conditions, "check_timestamp <= ?")
			args = append(args, filter.EndTime)
		}

		if filter.CheckType != "" {
			conditions = append(conditions, "check_type = ?")
			args = append(args, filter.CheckType)
		}

		if filter.Status != "" {
			conditions = append(conditions, "overall_status = ?")
			args = append(args, filter.Status)
		}

		if filter.RiskLevel != "" {
			conditions = append(conditions, "risk_level = ?")
			args = append(args, filter.RiskLevel)
		}
	}

	// 替换WHERE子句
	whereClause := strings.Join(conditions, " AND ")
	query = strings.Replace(query, "subject_id = ?", whereClause, 1)

	// 添加排序
	query += " ORDER BY check_timestamp DESC"

	// 添加分页
	if filter != nil {
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)

			if filter.Offset > 0 {
				query += " OFFSET ?"
				args = append(args, filter.Offset)
			}
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询历史结果失败: %w", err)
	}
	defer rows.Close()

	var results []*ComplianceCheckResult

	for rows.Next() {
		var result ComplianceCheckResult
		var violationsJSON, recommendationsJSON, actionsJSON, metadataJSON string
		var nextReviewDate sql.NullTime

		err := rows.Scan(
			&result.RequestID, &result.CheckType, &result.SubjectID, &result.SubjectType,
			&result.OverallStatus, &result.OverallScore, &result.RiskLevel,
			&violationsJSON, &recommendationsJSON, &actionsJSON,
			&result.CheckTimestamp, &nextReviewDate, &result.CheckedBy,
			&result.ProcessingTime, &metadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("扫描历史结果失败: %w", err)
		}

		if nextReviewDate.Valid {
			result.NextReviewDate = &nextReviewDate.Time
		}

		// 反序列化数据
		if err := json.Unmarshal([]byte(violationsJSON), &result.Violations); err != nil {
			r.logger.Warn("反序列化违规记录失败", "request_id", result.RequestID, "error", err)
			result.Violations = []Violation{}
		}

		if err := json.Unmarshal([]byte(recommendationsJSON), &result.Recommendations); err != nil {
			r.logger.Warn("反序列化建议失败", "request_id", result.RequestID, "error", err)
			result.Recommendations = []Recommendation{}
		}

		if err := json.Unmarshal([]byte(actionsJSON), &result.RequiredActions); err != nil {
			r.logger.Warn("反序列化必要行动失败", "request_id", result.RequestID, "error", err)
			result.RequiredActions = []RequiredAction{}
		}

		if err := json.Unmarshal([]byte(metadataJSON), &result.Metadata); err != nil {
			r.logger.Warn("反序列化元数据失败", "request_id", result.RequestID, "error", err)
			result.Metadata = make(map[string]interface{})
		}

		results = append(results, &result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历查询结果失败: %w", err)
	}

	return results, nil
}

// SaveViolation 保存违规记录
func (r *DatabaseComplianceRepository) SaveViolation(ctx context.Context, violation *Violation) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	evidenceJSON, err := json.Marshal(violation.Evidence)
	if err != nil {
		return fmt.Errorf("序列化证据失败: %w", err)
	}

	query := `
		INSERT INTO compliance_violations (
			violation_id, rule_id, rule_name, description, severity,
			detected_at, affected_resource, evidence, remediation, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(violation_id) DO UPDATE SET
			rule_name = excluded.rule_name,
			description = excluded.description,
			severity = excluded.severity,
			affected_resource = excluded.affected_resource,
			evidence = excluded.evidence,
			remediation = excluded.remediation,
			status = excluded.status
	`

	_, err = r.db.ExecContext(ctx, query,
		violation.ViolationID, violation.RuleID, violation.RuleName, violation.Description,
		violation.Severity, violation.DetectedAt, violation.AffectedResource,
		string(evidenceJSON), violation.Remediation, violation.Status,
	)

	if err != nil {
		return fmt.Errorf("保存违规记录失败: %w", err)
	}

	r.logger.Info("违规记录已保存到数据库",
		"violation_id", violation.ViolationID,
		"rule_id", violation.RuleID,
		"severity", violation.Severity)

	return nil
}

// FindViolations 查找违规记录
func (r *DatabaseComplianceRepository) FindViolations(ctx context.Context, filter *ViolationFilter) ([]Violation, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT violation_id, rule_id, rule_name, description, severity,
			   detected_at, affected_resource, evidence, remediation, status
		FROM compliance_violations
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	conditions := make([]string, 0)

	// 构建查询条件
	if filter != nil {
		if filter.RuleID != "" {
			conditions = append(conditions, "rule_id = ?")
			args = append(args, filter.RuleID)
		}

		if filter.Severity != "" {
			conditions = append(conditions, "severity = ?")
			args = append(args, filter.Severity)
		}

		if filter.Status != "" {
			conditions = append(conditions, "status = ?")
			args = append(args, filter.Status)
		}

		if filter.StartTime != nil {
			conditions = append(conditions, "detected_at >= ?")
			args = append(args, filter.StartTime)
		}

		if filter.EndTime != nil {
			conditions = append(conditions, "detected_at <= ?")
			args = append(args, filter.EndTime)
		}

		if filter.AssignedTo != "" {
			conditions = append(conditions, "JSON_EXTRACT(evidence, '$.assigned_to') = ?")
			args = append(args, filter.AssignedTo)
		}
	}

	// 添加查询条件
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// 添加排序
	query += " ORDER BY detected_at DESC"

	// 添加分页
	if filter != nil {
		if filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)

			if filter.Offset > 0 {
				query += " OFFSET ?"
				args = append(args, filter.Offset)
			}
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询违规记录失败: %w", err)
	}
	defer rows.Close()

	var violations []Violation

	for rows.Next() {
		var violation Violation
		var evidenceJSON string

		err := rows.Scan(
			&violation.ViolationID, &violation.RuleID, &violation.RuleName,
			&violation.Description, &violation.Severity, &violation.DetectedAt,
			&violation.AffectedResource, &evidenceJSON, &violation.Remediation, &violation.Status,
		)

		if err != nil {
			return nil, fmt.Errorf("扫描违规记录失败: %w", err)
		}

		// 反序列化证据
		if err := json.Unmarshal([]byte(evidenceJSON), &violation.Evidence); err != nil {
			r.logger.Warn("反序列化证据失败", "violation_id", violation.ViolationID, "error", err)
			violation.Evidence = make(map[string]interface{})
		}

		violations = append(violations, violation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历查询结果失败: %w", err)
	}

	return violations, nil
}

// CreateComplianceTables 创建合规相关表
func CreateComplianceTables(ctx context.Context, db *sql.DB) error {
	// 创建检查结果表
	resultQuery := `
	CREATE TABLE IF NOT EXISTS compliance_check_results (
		request_id TEXT PRIMARY KEY,
		check_type TEXT NOT NULL,
		subject_id TEXT NOT NULL,
		subject_type TEXT NOT NULL,
		overall_status TEXT NOT NULL,
		overall_score REAL,
		risk_level TEXT,
		violations TEXT,
		recommendations TEXT,
		required_actions TEXT,
		check_timestamp DATETIME NOT NULL,
		next_review_date DATETIME,
		checked_by TEXT,
		processing_time INTEGER,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_compliance_results_subject ON compliance_check_results(subject_id, subject_type);
	CREATE INDEX IF NOT EXISTS idx_compliance_results_type ON compliance_check_results(check_type);
	CREATE INDEX IF NOT EXISTS idx_compliance_results_status ON compliance_check_results(overall_status);
	CREATE INDEX IF NOT EXISTS idx_compliance_results_risk ON compliance_check_results(risk_level);
	CREATE INDEX IF NOT EXISTS idx_compliance_results_timestamp ON compliance_check_results(check_timestamp);
	`

	// 创建违规记录表
	violationQuery := `
	CREATE TABLE IF NOT EXISTS compliance_violations (
		violation_id TEXT PRIMARY KEY,
		request_id TEXT,
		rule_id TEXT NOT NULL,
		rule_name TEXT NOT NULL,
		description TEXT,
		severity TEXT NOT NULL,
		detected_at DATETIME NOT NULL,
		affected_resource TEXT,
		evidence TEXT,
		remediation TEXT,
		status TEXT NOT NULL DEFAULT 'NON_COMPLIANT',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (request_id) REFERENCES compliance_check_results(request_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_compliance_violations_rule ON compliance_violations(rule_id);
	CREATE INDEX IF NOT EXISTS idx_compliance_violations_severity ON compliance_violations(severity);
	CREATE INDEX IF NOT EXISTS idx_compliance_violations_status ON compliance_violations(status);
	CREATE INDEX IF NOT EXISTS idx_compliance_violations_detected ON compliance_violations(detected_at);
	CREATE INDEX IF NOT EXISTS idx_compliance_violations_request ON compliance_violations(request_id);
	`

	// 执行创建表的SQL
	for _, query := range []string{resultQuery, violationQuery} {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("创建合规表失败: %w", err)
		}
	}

	return nil
}

// ComplianceStatistics 合规统计
func (r *DatabaseComplianceRepository) GetStatistics(ctx context.Context, subjectID string) (*ComplianceStatistics, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query := `
		SELECT
			COUNT(*) as total_checks,
			COUNT(CASE WHEN overall_status = 'COMPLIANT' THEN 1 END) as passed_checks,
			COUNT(CASE WHEN overall_status = 'NON_COMPLIANT' THEN 1 END) as failed_checks,
			COUNT(CASE WHEN overall_status = 'PENDING' THEN 1 END) as pending_checks,
			COALESCE(AVG(overall_score), 0) as overall_score,
			COUNT(CASE WHEN risk_level = 'LOW' THEN 1 END) as low_risk_count,
			COUNT(CASE WHEN risk_level = 'MEDIUM' THEN 1 END) as medium_risk_count,
			COUNT(CASE WHEN risk_level = 'HIGH' THEN 1 END) as high_risk_count,
			COUNT(CASE WHEN risk_level = 'CRITICAL' THEN 1 END) as critical_risk_count,
			MAX(check_timestamp) as last_check_time
		FROM compliance_check_results
		WHERE subject_id = ?
	`

	var stats ComplianceStatistics
	var lowRisk, mediumRisk, highRisk, criticalRisk int64

	err := r.db.QueryRowContext(ctx, query, subjectID).Scan(
		&stats.TotalChecks, &stats.PassedChecks, &stats.FailedChecks,
		&stats.PendingChecks, &stats.OverallScore, &lowRisk, &mediumRisk,
		&highRisk, &criticalRisk, &stats.LastCheckTime,
	)

	if err != nil {
		return nil, fmt.Errorf("查询合规统计失败: %w", err)
	}

	// 构建风险分布
	stats.RiskDistribution = map[RiskLevel]int64{
		RiskLevelLow:      lowRisk,
		RiskLevelMedium:    mediumRisk,
		RiskLevelHigh:      highRisk,
		RiskLevelCritical:  criticalRisk,
	}

	// 查询违规数量
	violationQuery := `SELECT COUNT(*) FROM compliance_violations WHERE request_id IN (
		SELECT request_id FROM compliance_check_results WHERE subject_id = ?
	)`

	err = r.db.QueryRowContext(ctx, violationQuery, subjectID).Scan(&stats.ViolationCount)
	if err != nil {
		r.logger.Warn("查询违规数量失败", "subject_id", subjectID, "error", err)
		stats.ViolationCount = 0
	}

	// 查询建议数量（简化实现）
	stats.Recommendations = stats.FailedChecks // 暂时用失败检查数代替

	return &stats, nil
}