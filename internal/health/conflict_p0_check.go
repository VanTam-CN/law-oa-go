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

const completeConflictScopePredicate = "s.status = 'ACTIVE' AND s.coverage_status = 'COMPLETE' AND TRIM(COALESCE(s.source_version, '')) <> '' AND TRIM(COALESCE(s.evidence_reference, '')) <> '' AND s.covered_from IS NOT NULL AND s.covered_to IS NOT NULL AND (s.missing_sources IS NULL OR TRIM(s.missing_sources) IN ('', '[]')) AND TRIM(COALESCE(s.index_run_id, '')) <> '' AND s.source_of_truth = TRUE AND UPPER(TRIM(COALESCE(s.sync_mode, ''))) IN ('REALTIME', 'BATCH', 'MANUAL_IMPORT') AND s.max_sync_lag_minutes > 0 AND s.last_successful_sync_at IS NOT NULL AND s.minimum_field_coverage_bps > 0 AND s.minimum_field_coverage_bps <= 10000 AND s.measured_field_coverage_bps >= s.minimum_field_coverage_bps AND s.measured_field_coverage_bps <= 10000 AND s.maximum_duplicate_rate_bps >= 0 AND s.maximum_duplicate_rate_bps <= 10000 AND s.measured_duplicate_rate_bps >= 0 AND s.measured_duplicate_rate_bps <= s.maximum_duplicate_rate_bps AND s.quality_owner_id IS NOT NULL AND s.quality_owner_id > 0 AND s.quality_reviewed_at IS NOT NULL AND s.max_quality_review_age_days > 0 AND TRIM(COALESCE(s.failure_alert_reference, '')) <> '' AND TRIM(COALESCE(s.correction_procedure_reference, '')) <> '' AND r.status = 'COMPLETED' AND r.scope_type = s.scope_type AND r.source_version = s.source_version AND r.missing_record_count = 0 AND r.indexed_record_count >= r.source_record_count AND TRIM(COALESCE(r.reconciliation_hash, '')) <> '' AND TRIM(COALESCE(r.evidence_reference, '')) <> ''"

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

	now := time.Now().UTC()
	var currentPolicies int64
	// A policy profile is only release evidence when it is traceable to the
	// immutable material package and to two role-separated endorsements of the
	// same integrity hash. Checking the profile alone would let a direct
	// database insert imitate an approved policy without proving either signer
	// saw the exact material package.
	policyQuery := "SELECT COUNT(*) FROM law_firm_compliance_policy_profiles p JOIN law_firm_compliance_policy_packages pkg ON pkg.id = p.id AND pkg.policy_version = p.policy_version AND pkg.integrity_hash = p.integrity_hash JOIN law_firm_compliance_policy_endorsements management_endorsement ON management_endorsement.policy_package_id = pkg.id AND management_endorsement.endorsement_type = 'MANAGEMENT' AND management_endorsement.endorsed_by = p.management_approved_by AND management_endorsement.package_integrity_hash = pkg.integrity_hash JOIN law_firm_compliance_policy_endorsements compliance_endorsement ON compliance_endorsement.policy_package_id = pkg.id AND compliance_endorsement.endorsement_type = 'COMPLIANCE' AND compliance_endorsement.endorsed_by = p.compliance_approved_by AND compliance_endorsement.package_integrity_hash = pkg.integrity_hash JOIN users management_approver ON management_approver.id = p.management_approved_by JOIN users compliance_approver ON compliance_approver.id = p.compliance_approved_by WHERE p.status = 'APPROVED' AND p.management_approved_by > 0 AND p.compliance_approved_by > 0 AND p.management_approved_by <> p.compliance_approved_by AND p.approved_at IS NOT NULL AND p.effective_at IS NOT NULL AND p.effective_at <= " + c.placeholder(1) + " AND p.next_review_at IS NOT NULL AND p.next_review_at > " + c.placeholder(2) + " AND (p.expires_at IS NULL OR p.expires_at > " + c.placeholder(3) + ") AND TRIM(COALESCE(p.jurisdiction, '')) <> '' AND TRIM(COALESCE(p.applicable_rule_name, '')) <> '' AND TRIM(COALESCE(p.applicable_rule_version, '')) <> '' AND TRIM(COALESCE(p.applicable_rule_authority, '')) <> '' AND TRIM(COALESCE(p.applicable_rule_reference, '')) <> '' AND TRIM(COALESCE(p.data_source_policy_reference, '')) <> '' AND TRIM(COALESCE(p.privacy_basis_matrix_reference, '')) <> '' AND TRIM(COALESCE(p.retention_policy_reference, '')) <> '' AND TRIM(COALESCE(p.waiver_policy_reference, '')) <> '' AND TRIM(COALESCE(p.controlled_actions_reference, '')) <> '' AND TRIM(COALESCE(p.external_review_reference, '')) <> '' AND LENGTH(TRIM(COALESCE(p.integrity_hash, ''))) = 64 AND management_approver.deleted_at IS NULL AND management_approver.status = 'active' AND LOWER(management_approver.role) IN ('director', 'partner', 'management') AND compliance_approver.deleted_at IS NULL AND compliance_approver.status = 'active' AND LOWER(compliance_approver.role) IN ('compliance', 'risk', 'risk_control')"
	if err := c.db.QueryRowContext(ctx, policyQuery, now, now, now).Scan(&currentPolicies); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("律所冲突合规政策状态不可用: %v", err))
	}
	if currentPolicies != 1 {
		return conflictP0Unhealthy(result, fmt.Sprintf("必须且只能有一份当前有效的双人批准冲突合规政策: current=%d", currentPolicies))
	}
	var currentOfficerAppointments int64
	appointmentQuery := "SELECT COUNT(*) FROM conflict_officer_appointments a JOIN users officer ON officer.id = a.officer_id JOIN users deputy ON deputy.id = a.deputy_id JOIN users appointer ON appointer.id = a.appointed_by WHERE a.effective_from <= " + c.placeholder(1) + " AND a.effective_to > " + c.placeholder(2) + " AND a.officer_id <> a.deputy_id AND a.officer_id <> a.appointed_by AND a.deputy_id <> a.appointed_by AND TRIM(COALESCE(a.recusal_declaration, '')) <> '' AND officer.deleted_at IS NULL AND officer.status = 'active' AND LOWER(officer.role) IN ('director','partner','compliance','risk','risk_control','management','conflict_officer') AND deputy.deleted_at IS NULL AND deputy.status = 'active' AND LOWER(deputy.role) IN ('director','partner','compliance','risk','risk_control','management','conflict_officer') AND appointer.deleted_at IS NULL AND appointer.status = 'active'"
	if err := c.db.QueryRowContext(ctx, appointmentQuery, now, now).Scan(&currentOfficerAppointments); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("冲突核查人任命状态不可用: %v", err))
	}
	if currentOfficerAppointments == 0 {
		return conflictP0Unhealthy(result, "缺少当前有效且包含独立代理人的冲突核查人任命记录")
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
	rows, err := c.db.QueryContext(ctx, "SELECT scope_type, last_successful_sync_at, max_sync_lag_minutes, quality_reviewed_at, max_quality_review_age_days FROM conflict_search_scopes WHERE status = 'ACTIVE'")
	if err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案同步时效不可用: %v", err))
	}
	defer rows.Close()
	for rows.Next() {
		var scopeType string
		var lastSync time.Time
		var maxLag int
		var qualityReviewedAt time.Time
		var maxQualityAgeDays int
		if err := rows.Scan(&scopeType, &lastSync, &maxLag, &qualityReviewedAt, &maxQualityAgeDays); err != nil {
			return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案同步时效读取失败: %v", err))
		}
		if maxLag <= 0 || lastSync.After(now.Add(5*time.Minute)) || now.After(lastSync.Add(time.Duration(maxLag)*time.Minute)) {
			return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案来源 %s 已超过允许同步延迟", scopeType))
		}
		if maxQualityAgeDays <= 0 || qualityReviewedAt.After(now.Add(5*time.Minute)) || now.After(qualityReviewedAt.AddDate(0, 0, maxQualityAgeDays)) {
			return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案来源 %s 已超过允许质量复核周期", scopeType))
		}
	}
	if err := rows.Err(); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("冲突档案同步时效读取失败: %v", err))
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
	var incompleteClientIdentityRows int64
	if err := c.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM clients
		WHERE deleted_at IS NULL
		  AND (
			(COALESCE(identity_number_digest, '') <> '' AND (COALESCE(identity_type, '') = '' OR COALESCE(identity_number_ciphertext, '') = ''))
			OR (COALESCE(id_card_digest, '') <> '' AND (COALESCE(identity_type, '') = '' OR COALESCE(identity_number_digest, '') = '' OR COALESCE(identity_number_ciphertext, '') = ''))
		  )
	`).Scan(&incompleteClientIdentityRows); err != nil {
		return conflictP0Unhealthy(result, fmt.Sprintf("客户通用身份迁移状态不可用: %v", err))
	}
	if incompleteClientIdentityRows > 0 {
		return conflictP0Unhealthy(result, fmt.Sprintf("仍有%d条客户身份未迁移到通用保护字段，请执行000071并核验回填", incompleteClientIdentityRows))
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

	for _, table := range []string{"users", "law_firm_compliance_policy_profiles", "case_subject_revisions", "compliance_audit_events", "conflict_subject_versions", "conflict_subject_identifiers", "conflict_match_evidence_v2", "conflict_index_build_runs", "entities", "clients", "case_parties", "entity_name_history", "entity_relations"} {
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
