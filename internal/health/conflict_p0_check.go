package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ConflictP0ReadinessCheck makes the production health endpoint reflect the
// conflict-control prerequisite, not only infrastructure availability. An
// empty or partially registered archive scope must keep the service out of
// rotation because the application intentionally returns COVERAGE_LIMITED.
type ConflictP0ReadinessCheck struct {
	db         *sql.DB
	production bool
	driver     string
}

var requiredConflictScopeTypes = []string{
	"CASE_ARCHIVE",
	"CLIENT_ARCHIVE",
	"SUBJECT_REGISTRY",
	"RELATION_ARCHIVE",
}

const completeConflictScopePredicate = "s.status = 'ACTIVE' AND s.coverage_status = 'COMPLETE' AND TRIM(COALESCE(s.source_version, '')) <> '' AND TRIM(COALESCE(s.evidence_reference, '')) <> '' AND s.covered_from IS NOT NULL AND s.covered_to IS NOT NULL AND (s.missing_sources IS NULL OR TRIM(s.missing_sources) IN ('', '[]')) AND TRIM(COALESCE(s.index_run_id, '')) <> '' AND r.status = 'COMPLETED' AND r.scope_type = s.scope_type AND r.source_version = s.source_version AND r.missing_record_count = 0 AND r.indexed_record_count >= r.source_record_count AND TRIM(COALESCE(r.reconciliation_hash, '')) <> '' AND TRIM(COALESCE(r.evidence_reference, '')) <> ''"

func NewConflictP0ReadinessCheck(db *sql.DB, production bool, drivers ...string) *ConflictP0ReadinessCheck {
	driver := "sqlite"
	if len(drivers) > 0 && drivers[0] != "" {
		driver = drivers[0]
	}
	return &ConflictP0ReadinessCheck{db: db, production: production, driver: driver}
}

func (c *ConflictP0ReadinessCheck) GetName() string { return "conflict_p0_readiness" }

func (c *ConflictP0ReadinessCheck) GetTimeout() time.Duration { return 3 * time.Second }

func (c *ConflictP0ReadinessCheck) Check(ctx context.Context) *HealthCheckResult {
	started := time.Now()
	result := &HealthCheckResult{
		Name:      c.GetName(),
		Status:    StatusHealthy,
		Timestamp: started,
	}
	result.Duration = time.Since(started).Milliseconds()

	if !c.production {
		result.Message = "生产冲突门禁仅在 production 环境强制执行"
		return result
	}
	if c.db == nil {
		return conflictP0Unhealthy(result, "生产冲突门禁数据库连接未初始化")
	}
	var demoAccounts int64
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND status = 'active' AND (LOWER(email) LIKE '%@example.test' OR LOWER(username) LIKE 'demo_%')").Scan(&demoAccounts); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("生产账号清理状态不可用: %v", err))
	}
	if demoAccounts > 0 {
		return conflictP0Unhealthy(result, fmt.Sprintf("仍有%d个演示账号处于启用状态，禁止进入生产", demoAccounts))
	}

	var activeScopes, completeScopes int64
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM conflict_search_scopes WHERE status = 'ACTIVE'").Scan(&activeScopes); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案覆盖表不可用: %v", err))
	}
	var duplicateActiveScopeTypes int64
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM (SELECT scope_type FROM conflict_search_scopes WHERE status = 'ACTIVE' GROUP BY scope_type HAVING COUNT(*) > 1) duplicates").Scan(&duplicateActiveScopeTypes); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案来源唯一性不可用: %v", err))
	}
	if duplicateActiveScopeTypes > 0 {
		return conflictP0Unhealthy(result, fmt.Sprintf("存在%d个重复的 ACTIVE 冲突档案来源类型，禁止进入生产", duplicateActiveScopeTypes))
	}
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM conflict_search_scopes s JOIN conflict_index_build_runs r ON r.id = s.index_run_id WHERE "+completeConflictScopePredicate).Scan(&completeScopes); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案覆盖状态不可用: %v", err))
	}
	if activeScopes == 0 || activeScopes != completeScopes {
		return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案覆盖未完成: active=%d complete=%d", activeScopes, completeScopes))
	}
	var plaintextIdentityRows int64
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM entities WHERE identity_number IS NOT NULL AND TRIM(identity_number) <> ''").Scan(&plaintextIdentityRows); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("主体身份保护状态不可用: %v", err))
	}
	if plaintextIdentityRows > 0 {
		return conflictP0Unhealthy(result, fmt.Sprintf("仍有%d条主体身份信息未完成加密回填", plaintextIdentityRows))
	}
	var plaintextClientIdentityRows int64
	if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM clients WHERE id_card IS NOT NULL AND TRIM(id_card) <> ''").Scan(&plaintextClientIdentityRows); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("客户身份保护状态不可用: %v", err))
	}
	if plaintextClientIdentityRows > 0 {
		return conflictP0Unhealthy(result, fmt.Sprintf("仍有%d条客户身份信息未完成加密回填", plaintextClientIdentityRows))
	}

	for _, scopeType := range requiredConflictScopeTypes {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM conflict_search_scopes s JOIN conflict_index_build_runs r ON r.id = s.index_run_id WHERE s.scope_type = %s AND "+completeConflictScopePredicate, c.placeholder(1))
		if err := c.db.QueryRowContext(ctx, query, scopeType).Scan(&count); err != nil {
			return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案来源 %s 不可用: %v", scopeType, err))
		}
		if count == 0 {
			return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案覆盖缺少必需来源: %s", scopeType))
		}
	}

	for _, table := range []string{"users", "case_subject_revisions", "compliance_audit_events", "conflict_subject_versions", "conflict_subject_identifiers", "conflict_match_evidence_v2", "conflict_index_build_runs", "entities", "clients", "case_parties", "entity_name_history", "entity_relations"} {
		var count int64
		if err := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			return conflictP0Unhealthy(result, fmt.Sprintf("P0证据表 %s 不可用: %v", table, err))
		}
	}
	result.Message = fmt.Sprintf("冲突档案覆盖完整，active=%d", activeScopes)
	result.Duration = time.Since(started).Milliseconds()
	return result
}

func (c *ConflictP0ReadinessCheck) placeholder(index int) string {
	if c != nil && (c.driver == "postgres" || c.driver == "postgresql") {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

func conflictP0Unhealthy(result *HealthCheckResult, message string) *HealthCheckResult {
	result.Status = StatusUnhealthy
	result.Message = message
	result.Duration = time.Since(result.Timestamp).Milliseconds()
	return result
}
