package health

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

const ConflictP0ReadinessCheckName = "conflict_p0_readiness"

var requiredProductionEvidenceGates = []string{"G0", "G1", "G2", "G3", "G4", "G5", "G6", "G7"}

var productionEvidenceGateActions = map[string]string{
	"G0": "完成并签署 PD-01 至 PD-07 决策物，登记书面签署凭证",
	"G1": "核验独立生产密钥和 TLS 数据库连接，并完成备份恢复演练登记",
	"G2": "完成目标生产库 bootstrap/迁移，并登记成功日志凭证",
	"G3": "完成敏感身份加密回填和零明文复核，并登记复核凭证",
	"G4": "完成案件、客户、主体、关系四类冲突索引回填、对账和凭证登记",
	"G5": "提交双人批准的冲突合规政策，登记独立核查人、代理人及范围质量凭证",
	"G6": "停用演示账号并完成 AT-01 至 AT-12 三角色隔离验收，登记验收记录",
	"G7": "由运维负责人和合规负责人复核前序凭证，登记最终技术放行记录",
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

func (c *ConflictP0ReadinessCheck) GetName() string { return ConflictP0ReadinessCheckName }

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
		result.Message = "QA技术就绪：生产数据库前置条件未执行；ready=true不代表生产可用"
		result.Details = conflictP0DatabasePrerequisitesDetails(false, nil)
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
	if err := c.verifyG5GovernanceAuditEvidence(ctx); err != nil {
		return conflictP0Unhealthy(result, err.Error())
	}
	if err := c.verifyExternalEvidenceRegistrations(ctx); err != nil {
		return conflictP0Unhealthy(result, err.Error())
	}
	result.Message = fmt.Sprintf("生产数据库前置条件和 G0-G7 外部证据登记复核通过：冲突档案覆盖完整，active=%d", activeScopes)
	result.Details = conflictP0DatabasePrerequisitesDetails(true, requiredProductionEvidenceGates)
	result.Duration = time.Since(started).Milliseconds()
	return result
}

func (c *ConflictP0ReadinessCheck) verifyG5GovernanceAuditEvidence(ctx context.Context) error {
	eventTypes := []string{
		"CONFLICT_POLICY_PACKAGE_CREATED",
		"CONFLICT_POLICY_ENDORSED",
		"CONFLICT_POLICY_APPROVED",
		"CONFLICT_OFFICER_APPOINTED",
		"CONFLICT_SCOPE_UPDATED",
	}
	placeholders := make([]string, 0, len(eventTypes))
	args := make([]interface{}, 0, len(eventTypes))
	for index, eventType := range eventTypes {
		placeholders = append(placeholders, c.placeholder(index+1))
		args = append(args, eventType)
	}
	query := fmt.Sprintf(
		"SELECT event_type, COUNT(*) FROM compliance_audit_events WHERE event_type IN (%s) AND integrity_hash IS NOT NULL AND LENGTH(TRIM(integrity_hash)) = 64 GROUP BY event_type",
		strings.Join(placeholders, ","),
	)
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("生产治理审计证据不可用: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int64, len(eventTypes))
	for rows.Next() {
		var eventType string
		var count int64
		if err := rows.Scan(&eventType, &count); err != nil {
			return fmt.Errorf("生产治理审计证据读取失败: %w", err)
		}
		counts[eventType] = count
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("生产治理审计证据读取失败: %w", err)
	}

	missing := make([]string, 0)
	for _, eventType := range eventTypes {
		if counts[eventType] == 0 {
			missing = append(missing, eventType)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("缺少带完整性哈希的生产治理审计证据: %s；请完成 G5 双人批准、任命登记和范围质量登记", strings.Join(missing, ","))
	}
	if err := ensureProductionExternalEvidenceTable(ctx, c.db); err != nil {
		return err
	}
	return nil
}

func (c *ConflictP0ReadinessCheck) verifyExternalEvidenceRegistrations(ctx context.Context) error {
	rows, err := c.db.QueryContext(ctx, `
		SELECT id, gate, evidence_reference, reviewed_by, reviewer_role, review_result,
		       reviewed_at, integrity_hash, created_at, updated_at
		FROM production_external_evidence
	`)
	if err != nil {
		return fmt.Errorf("生产外部证据登记不可用: %w", err)
	}
	defer rows.Close()

	registered := make(map[string]error)
	for rows.Next() {
		var registration productionExternalEvidenceRegistration
		if err := rows.Scan(
			&registration.ID, &registration.Gate, &registration.EvidenceReference, &registration.ReviewedBy,
			&registration.ReviewerRole, &registration.ReviewResult, &registration.ReviewedAt,
			&registration.IntegrityHash, &registration.CreatedAt, &registration.UpdatedAt,
		); err != nil {
			return fmt.Errorf("生产外部证据登记读取失败: %w", err)
		}
		validationError := registration.validate(time.Now().UTC())
		registered[registration.NormalizedGate()] = validationError
		if validationError == nil && registration.G7FinalReviewValid() {
			continue
		}
		if validationError == nil {
			registered[registration.NormalizedGate()] = fmt.Errorf("G7最终复核记录未由两名不同责任人共同签署")
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("生产外部证据登记读取失败: %w", err)
	}

	missing := make([]string, 0)
	for _, gate := range requiredProductionEvidenceGates {
		validationError, exists := registered[gate]
		if !exists || validationError != nil {
			missing = append(missing, formatProductionExternalEvidenceGap(gate, validationError))
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("生产外部证据门禁未完整登记或复核: %s", strings.Join(missing, "；"))
	}
	return nil
}

func ensureProductionExternalEvidenceTable(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("生产外部证据登记数据库连接未初始化")
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS production_external_evidence (
		id BIGSERIAL PRIMARY KEY,
		gate VARCHAR(8) NOT NULL UNIQUE,
		evidence_reference TEXT NOT NULL,
		reviewed_by VARCHAR(120) NOT NULL,
		reviewer_role VARCHAR(80) NOT NULL,
		review_result VARCHAR(20) NOT NULL,
		reviewed_at TEXT NOT NULL,
		integrity_hash VARCHAR(64) NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		CONSTRAINT chk_production_external_gate CHECK (gate IN ('G0','G1','G2','G3','G4','G5','G6','G7')),
		CONSTRAINT chk_production_external_result CHECK (review_result IN ('PASSED','FAILED'))
	)`); err != nil {
		return fmt.Errorf("生产外部证据登记表不可用: %w", err)
	}
	return nil
}

type productionExternalEvidenceRegistration struct {
	ID                int64
	Gate              string
	EvidenceReference string
	ReviewedBy        string
	ReviewerRole      string
	ReviewResult      string
	ReviewedAt        string
	IntegrityHash     string
	CreatedAt         string
	UpdatedAt         string
}

func (r productionExternalEvidenceRegistration) NormalizedGate() string {
	return strings.ToUpper(strings.TrimSpace(r.Gate))
}

func (r productionExternalEvidenceRegistration) validate(now time.Time) error {
	if _, supported := productionEvidenceGateActions[r.NormalizedGate()]; !supported {
		return fmt.Errorf("未知门禁 %s", r.Gate)
	}
	if strings.TrimSpace(r.EvidenceReference) == "" {
		return fmt.Errorf("缺少凭证引用")
	}
	if strings.TrimSpace(r.ReviewedBy) == "" || strings.TrimSpace(r.ReviewerRole) == "" {
		return fmt.Errorf("缺少复核责任人")
	}
	if strings.TrimSpace(r.ReviewResult) != "PASSED" {
		return fmt.Errorf("复核结论不是 PASSED")
	}
	reviewedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(r.ReviewedAt))
	if err != nil || reviewedAt.IsZero() || reviewedAt.After(now.Add(5*time.Minute)) {
		return fmt.Errorf("复核时间无效")
	}
	if len(strings.TrimSpace(r.IntegrityHash)) != 64 {
		return fmt.Errorf("完整性哈希无效")
	}
	createdAt, createdAtErr := time.Parse(time.RFC3339, strings.TrimSpace(r.CreatedAt))
	updatedAt, updatedAtErr := time.Parse(time.RFC3339, strings.TrimSpace(r.UpdatedAt))
	if createdAtErr != nil || updatedAtErr != nil || createdAt.IsZero() || updatedAt.Before(createdAt) {
		return fmt.Errorf("登记时间无效")
	}
	return nil
}

func (r productionExternalEvidenceRegistration) G7FinalReviewValid() bool {
	if r.NormalizedGate() != "G7" {
		return true
	}
	// The stored reviewed_by field carries two distinct reviewer identities
	// separated by '|'. Operator tooling must create this record only from two
	// separately authenticated acknowledgements; readiness never invents it.
	reviewers := strings.Split(r.ReviewedBy, "|")
	if len(reviewers) != 2 {
		return false
	}
	return strings.TrimSpace(reviewers[0]) != "" &&
		strings.TrimSpace(reviewers[1]) != "" &&
		strings.TrimSpace(reviewers[0]) != strings.TrimSpace(reviewers[1])
}

func formatProductionExternalEvidenceGap(gate string, err error) string {
	if err == nil {
		return gate
	}
	return fmt.Sprintf("%s(%v)", gate, err)
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
	result.Details = conflictP0DatabasePrerequisitesDetails(false, nil)
	result.Duration = time.Since(result.Timestamp).Milliseconds()
	return result
}

func conflictP0DatabasePrerequisitesDetails(ready bool, passedGates []string) map[string]interface{} {
	externalGates := make(map[string]bool, len(requiredProductionEvidenceGates))
	for _, gate := range requiredProductionEvidenceGates {
		externalGates[gate] = false
	}
	for _, gate := range passedGates {
		externalGates[strings.ToUpper(strings.TrimSpace(gate))] = true
	}
	return map[string]interface{}{
		"production_database_prerequisites_ready": ready,
		"external_evidence_gates":                 externalGates,
		"next_actions":                            productionEvidenceGateActions,
	}
}
