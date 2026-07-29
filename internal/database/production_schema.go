package database

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// ProductionSchemaVersion identifies the idempotent PostgreSQL bootstrap
// contract. A future breaking schema change must introduce a new version and
// an explicit migration instead of relying on application startup side effects.
const ProductionSchemaVersion = "postgres-mvp-2026-07-19-v4"

// productionCaseIntake and the following small records back tables that are
// intentionally written through Table(...) by the intake/approval workflow.
// Keeping their schema in this bootstrap prevents those raw writes from being
// silently absent in a fresh production database.
type productionCaseIntake struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IntakeCode  string    `gorm:"size:50;not null;uniqueIndex"`
	ClientID    *uint     `gorm:"index"`
	Title       string    `gorm:"size:255;not null"`
	CaseType    string    `gorm:"size:100"`
	Status      string    `gorm:"size:40;not null;default:'draft';index"`
	Priority    string    `gorm:"size:20;not null;default:'medium'"`
	Description string    `gorm:"type:text"`
	Metadata    string    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedBy   *uint     `gorm:"index"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (productionCaseIntake) TableName() string { return "case_intakes" }

type productionCaseIntakeParty struct {
	ID            uint      `gorm:"primaryKey"`
	CaseID        *uint     `gorm:"index"`
	IntakeID      string    `gorm:"type:uuid;index"`
	EntityName    string    `gorm:"size:255;not null"`
	EntityType    string    `gorm:"size:50;not null;default:'company'"`
	PartyRole     string    `gorm:"size:80;not null"`
	RelationDepth int       `gorm:"not null;default:0"`
	Metadata      string    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (productionCaseIntakeParty) TableName() string { return "case_intake_parties" }

type productionCaseMaterial struct {
	ID           uint      `gorm:"primaryKey"`
	CaseID       *uint     `gorm:"index"`
	IntakeID     string    `gorm:"type:uuid;index"`
	Name         string    `gorm:"size:255;not null"`
	MaterialType string    `gorm:"size:80"`
	Status       string    `gorm:"size:40;not null;default:'missing'"`
	Required     bool      `gorm:"not null;default:true"`
	StorageURL   string    `gorm:"type:text"`
	Metadata     string    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (productionCaseMaterial) TableName() string { return "case_materials" }

type productionApprovalSnapshot struct {
	ID                string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ApprovalRequestID string    `gorm:"size:36;not null;index"`
	SnapshotType      string    `gorm:"size:80;not null"`
	SnapshotData      string    `gorm:"type:jsonb;not null"`
	SourceVersion     int       `gorm:"not null;default:1"`
	CreatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (productionApprovalSnapshot) TableName() string { return "approval_snapshots" }

// productionApprovalTemplate is a union of the legacy template columns and
// the V2 template columns. Both services still use the shared table, so the
// fresh schema must expose both contracts without choosing incompatible ID
// types from two different Go models.
type productionApprovalTemplate struct {
	ID                    uint       `gorm:"primaryKey"`
	Name                  string     `gorm:"size:50;uniqueIndex"`
	DisplayName           string     `gorm:"size:100"`
	Description           string     `gorm:"type:text"`
	Steps                 string     `gorm:"type:jsonb"`
	Conditions            string     `gorm:"type:jsonb"`
	IsActive              bool       `gorm:"not null;default:true"`
	TemplateCode          string     `gorm:"size:100;uniqueIndex"`
	TemplateName          string     `gorm:"size:255"`
	TemplateType          string     `gorm:"size:100"`
	Category              string     `gorm:"size:100"`
	WorkflowType          string     `gorm:"size:100"`
	TemplateContent       string     `gorm:"type:jsonb"`
	FormSchema            string     `gorm:"type:jsonb"`
	ValidationRules       string     `gorm:"type:jsonb"`
	DefaultValues         string     `gorm:"type:jsonb"`
	RequiredFields        string     `gorm:"type:jsonb"`
	OptionalFields        string     `gorm:"type:jsonb"`
	ApplicableScenarios   string     `gorm:"type:jsonb"`
	ApplicableTypes       string     `gorm:"type:jsonb"`
	ApplicableDepartments string     `gorm:"type:jsonb"`
	ApplicableRoles       string     `gorm:"type:jsonb"`
	Status                string     `gorm:"size:20;not null;default:'active';index"`
	EffectiveDate         *time.Time `gorm:"type:date"`
	ExpiryDate            *time.Time `gorm:"type:date"`
	UsageCount            int        `gorm:"not null;default:0"`
	LastUsedDate          *time.Time
	CreatedBy             string         `gorm:"size:36"`
	UpdatedBy             string         `gorm:"size:36"`
	ApprovedBy            string         `gorm:"size:36"`
	Version               int            `gorm:"not null;default:1"`
	CreatedAt             time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt             time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt             gorm.DeletedAt `gorm:"index"`
}

func (productionApprovalTemplate) TableName() string { return "approval_templates" }

type productionSchemaState struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"size:120;not null"`
	AppliedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (productionSchemaState) TableName() string { return "schema_bootstrap_state" }

// productionDepartment avoids pulling the legacy Department.Users relation
// into GORM's PostgreSQL schema parser. The production runtime currently
// reads department data through repositories, so a flat table contract is
// sufficient and portable for bootstrap purposes.
type productionDepartment struct {
	ID          uint           `gorm:"primaryKey"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Name        string         `gorm:"size:100;not null"`
	Code        string         `gorm:"size:50;not null;uniqueIndex"`
	ParentID    uint           `gorm:"column:parent_id;default:0"`
	LeaderID    *uint          `gorm:"column:leader_id;index"`
	Description string         `gorm:"type:text"`
	SortOrder   int            `gorm:"column:sort_order;default:0"`
	Status      int            `gorm:"default:1"`
}

func (productionDepartment) TableName() string { return "departments" }

// BootstrapProductionSchema creates the PostgreSQL schema used by the
// production MVP. It is deliberately explicit and idempotent. The operation
// is transactional so a failed fresh install cannot leave a partially-created
// database that looks usable to the application.
func BootstrapProductionSchema(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	if db.Dialector.Name() != "postgres" {
		return fmt.Errorf("生产 schema bootstrap 仅支持 PostgreSQL，当前驱动为 %s", db.Dialector.Name())
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Every application replica may run the bootstrap init container. Hold a
		// transaction-scoped PostgreSQL advisory lock so concurrent first starts
		// cannot race while creating the same tables, indexes, or guards.
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtext(?))`, "law-oa-production-schema-bootstrap").Error; err != nil {
			return fmt.Errorf("获取生产 schema bootstrap 锁失败: %w", err)
		}
		if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
			return fmt.Errorf("启用 pgcrypto 扩展失败: %w", err)
		}
		if err := ensureProductionSchemaAdditiveColumns(tx); err != nil {
			return err
		}

		for _, model := range productionSchemaModels() {
			// Existing installations may have been created by an older SQL
			// migration set. GORM AutoMigrate is not a safe compatibility
			// mechanism for those tables: it can try to drop a legacy index or
			// constraint whose name was never present in PostgreSQL. Only create
			// missing tables here; upgrades of existing tables belong in reviewed
			// SQL migrations and are checked below by the contract validator.
			if tx.Migrator().HasTable(model) {
				continue
			}
			// CreateTable is intentionally used for missing models. AutoMigrate
			// recursively inspects legacy associations and may attempt to alter an
			// already-existing related table while creating a new one. That is
			// precisely the unsafe upgrade behavior this bootstrap must avoid.
			if err := tx.Migrator().CreateTable(model); err != nil {
				return fmt.Errorf("创建生产表 %T 失败: %w", model, err)
			}
		}
		if err := ensureProductionSchemaContractColumns(tx); err != nil {
			return err
		}
		if err := installProductionAppendOnlyGuards(tx); err != nil {
			return err
		}
		if err := validateProductionSchemaContract(tx); err != nil {
			return err
		}

		if err := tx.Exec(`
			INSERT INTO schema_bootstrap_state (id, version, applied_at)
			VALUES (1, ?, CURRENT_TIMESTAMP)
			ON CONFLICT (id) DO UPDATE
			SET version = EXCLUDED.version, applied_at = EXCLUDED.applied_at
		`, ProductionSchemaVersion).Error; err != nil {
			return fmt.Errorf("写入生产 schema 版本失败: %w", err)
		}
		return nil
	})
}

// validateProductionSchemaContract verifies the columns used by the P0
// runtime after a bootstrap. It intentionally checks names, not ORM tags, so
// a legacy table can be upgraded by an explicit migration without allowing a
// partially compatible database to be marked ready.
func validateProductionSchemaContract(db *gorm.DB) error {
	required := map[string][]string{
		"users": {
			"id", "username", "email", "password", "role", "status", "department_id", "manager_id",
		},
		"clients": {
			"id", "name", "type", "id_card", "id_card_digest", "id_card_ciphertext", "status",
		},
		"cases": {
			"id", "case_number", "title", "client_id", "lawyer_id", "status",
			"subject_version", "subject_state", "subject_snapshot", "pending_subject_revision_id",
			"conflict_check_id", "conflict_coverage_status", "ethical_wall_enabled",
		},
		"case_ethical_wall_whitelist": {
			"id", "case_id", "user_id", "granted_by", "granted_at", "reason",
		},
		"ethical_wall_access_logs": {
			"id", "case_id", "user_id", "access_type", "access_result", "ip_address", "user_agent", "attempted_at",
		},
		"entities": {
			"id", "name", "entity_type", "identity_number", "identity_number_digest", "identity_number_ciphertext", "status",
		},
		"entity_relations":    {"id", "source_entity_id", "target_entity_id", "relation_type", "is_active"},
		"entity_name_history": {"id", "entity_id", "old_name", "new_name"},
		"case_parties":        {"id", "case_id", "entity_id", "role"},
		"conflict_subject_versions": {
			"id", "subject_key", "source_type", "source_id", "case_id", "client_id",
			"subject_role", "subject_type", "original_name", "normalized_name", "alias_snapshot",
			"source_version", "version_number", "verification_status", "snapshot", "created_at",
		},
		"conflict_subject_identifiers": {
			"id", "subject_version_id", "identifier_type", "digest", "ciphertext", "masked_value",
			"verification_status", "source_reference", "created_at",
		},
		"conflict_match_evidence_v2": {
			"id", "check_id", "subject_version_id", "match_type", "source_type", "source_object_id",
			"restricted", "evidence_snapshot", "evidence_hash", "created_at",
		},
		"conflict_search_scopes": {
			"id", "scope_type", "status", "coverage_status", "source_version", "evidence_reference",
			"covered_from", "covered_to", "missing_sources", "index_run_id", "approved_by", "approved_at",
		},
		"conflict_index_build_runs": {
			"id", "scope_type", "source_version", "status", "source_record_count", "indexed_record_count",
			"missing_record_count", "reconciliation_hash", "evidence_reference", "started_at", "completed_at",
			"created_by", "error_message", "created_at", "updated_at",
		},
		"case_subject_revisions": {
			"id", "case_id", "base_subject_version", "change_type", "status", "payload",
			"conflict_check_id", "requested_by", "reviewed_by", "review_decision", "review_notes",
		},
		"compliance_audit_events": {
			"id", "actor_id", "actor_role", "event_type", "object_type", "object_id", "request_id",
			"from_state", "to_state", "subject_version", "payload", "integrity_hash", "created_at",
		},
		"conflict_reviewer_assignments": {
			"id", "check_id", "reviewer_id", "assigned_by", "status", "recusal_declared",
			"independence_reason", "sla_due_at", "effective_from", "effective_to", "revoked_at",
		},
		"approval_requests": {
			"id", "request_number", "title", "type", "status", "applicant_id", "created_by",
			"conflict_check_id", "conflict_result", "case_created", "created_case_id", "case_creation_status",
		},
		"case_intakes":              {"id", "intake_code", "title", "status", "metadata", "created_by"},
		"case_intake_parties":       {"id", "intake_id", "entity_name", "party_role", "metadata"},
		"case_materials":            {"id", "intake_id", "name", "status", "metadata"},
		"approval_snapshots":        {"id", "approval_request_id", "snapshot_type", "snapshot_data", "source_version"},
		"waiver_applications":       {"id", "application_number", "conflict_check_id", "client_id", "lawyer_id", "status", "assigned_reviewer", "review_deadline", "created_by"},
		"waiver_approval_records":   {"id", "waiver_application_id", "approver_id", "decision", "approval_date", "effective_date", "expiry_date", "next_review_date", "status"},
		"waiver_signatures":         {"id", "waiver_application_id", "signer_type", "signature_content", "signature_timestamp", "status"},
		"waiver_monitoring_records": {"id", "waiver_application_id", "monitoring_type", "monitoring_date", "compliance_status", "monitored_by", "status"},
		"schema_bootstrap_state":    {"id", "version", "applied_at"},
	}
	for table, columns := range required {
		columnTypes, err := db.Migrator().ColumnTypes(table)
		if err != nil {
			return fmt.Errorf("读取生产表 %s 结构失败: %w", table, err)
		}
		available := make(map[string]struct{}, len(columnTypes))
		for _, column := range columnTypes {
			available[strings.ToLower(column.Name())] = struct{}{}
		}
		missing := make([]string, 0)
		for _, column := range columns {
			if _, ok := available[strings.ToLower(column)]; !ok {
				missing = append(missing, column)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("生产表 %s 缺少关键字段: %s；请执行对应的评审迁移后重试", table, strings.Join(missing, ", "))
		}
	}
	var uniqueReviewIndexCount int64
	if err := db.Raw(`
		SELECT COUNT(*)
		FROM pg_indexes
		WHERE schemaname = current_schema()
		  AND indexname = 'uq_conflict_reviews_check_id'
		  AND indexdef LIKE 'CREATE UNIQUE INDEX%'
	`).Scan(&uniqueReviewIndexCount).Error; err != nil {
		return fmt.Errorf("读取冲突复核唯一约束失败: %w", err)
	}
	if uniqueReviewIndexCount != 1 {
		return fmt.Errorf("conflict_reviews 缺少同一检测单次复核唯一约束")
	}
	for _, trigger := range []struct {
		name  string
		table string
	}{
		{name: "trg_compliance_audit_events_append_only", table: "compliance_audit_events"},
		{name: "trg_conflict_reviews_append_only", table: "conflict_reviews"},
		{name: "trg_case_subject_revisions_append_only", table: "case_subject_revisions"},
		{name: "trg_conflict_checks_no_delete", table: "conflict_checks"},
		{name: "trg_conflict_details_append_only", table: "conflict_details"},
		{name: "trg_conflict_check_records_no_delete", table: "conflict_check_records"},
		{name: "trg_conflict_cases_no_delete", table: "conflict_cases"},
		{name: "trg_conflict_reviewer_assignments_no_delete", table: "conflict_reviewer_assignments"},
		{name: "trg_conflict_subject_versions_append_only", table: "conflict_subject_versions"},
		{name: "trg_conflict_subject_identifiers_append_only", table: "conflict_subject_identifiers"},
		{name: "trg_conflict_match_evidence_v2_append_only", table: "conflict_match_evidence_v2"},
		{name: "trg_conflict_index_build_runs_no_delete", table: "conflict_index_build_runs"},
		{name: "trg_waiver_applications_no_delete", table: "waiver_applications"},
		{name: "trg_waiver_approval_records_append_only", table: "waiver_approval_records"},
		{name: "trg_waiver_signatures_append_only", table: "waiver_signatures"},
		{name: "trg_waiver_monitoring_records_append_only", table: "waiver_monitoring_records"},
	} {
		var triggerCount int64
		if err := db.Raw(`
			SELECT COUNT(*)
			FROM pg_trigger t
			JOIN pg_class c ON c.oid = t.tgrelid
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = current_schema()
			  AND c.relname = ?
			  AND t.tgname = ?
			  AND NOT t.tgisinternal
		`, trigger.table, trigger.name).Scan(&triggerCount).Error; err != nil {
			return fmt.Errorf("读取 %s 追加保护失败: %w", trigger.table, err)
		}
		if triggerCount != 1 {
			return fmt.Errorf("%s 缺少 append-only 保护触发器", trigger.table)
		}
	}
	return nil
}

// ValidateProductionSchemaContract is used by the application startup gate
// as well as the bootstrap command. Keeping one contract prevents a database
// from passing migration-time checks and then failing on the first P0 request.
func ValidateProductionSchemaContract(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	return validateProductionSchemaContract(db)
}

// ensureProductionSchemaAdditiveColumns upgrades only the reviewed, additive
// columns that are required by the P0 runtime. It deliberately does not use
// AutoMigrate: existing production tables may have legacy indexes and
// constraints that GORM cannot safely infer. Destructive changes belong in a
// separately reviewed migration and are never performed during startup.
func ensureProductionSchemaAdditiveColumns(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS department_id BIGINT`,
		`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS manager_id BIGINT`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("升级 users 复核关系字段失败: %w", err)
		}
	}
	return nil
}

// ensureProductionSchemaContractColumns applies only additive compatibility
// changes required by the P0 runtime. This covers databases created by older
// migration sets while keeping the bootstrap fail-closed for any field not in
// the reviewed contract. No column is dropped, rewritten, or data-cleared.
func ensureProductionSchemaContractColumns(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS department_id BIGINT`,
		`ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS manager_id BIGINT`,
		`CREATE INDEX IF NOT EXISTS idx_users_department_id ON users (department_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_manager_id ON users (manager_id)`,

		`ALTER TABLE IF EXISTS clients ADD COLUMN IF NOT EXISTS id_card_digest VARCHAR(64)`,
		`ALTER TABLE IF EXISTS clients ADD COLUMN IF NOT EXISTS id_card_ciphertext TEXT`,
		`CREATE INDEX IF NOT EXISTS idx_clients_id_card_digest ON clients (id_card_digest)`,
		`ALTER TABLE IF EXISTS entities ADD COLUMN IF NOT EXISTS identity_number_digest VARCHAR(64)`,
		`ALTER TABLE IF EXISTS entities ADD COLUMN IF NOT EXISTS identity_number_ciphertext TEXT`,
		`CREATE INDEX IF NOT EXISTS idx_entities_identity_number_digest ON entities (identity_number_digest)`,

		`ALTER TABLE IF EXISTS cases ADD COLUMN IF NOT EXISTS subject_version INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE IF EXISTS cases ADD COLUMN IF NOT EXISTS subject_state VARCHAR(50) NOT NULL DEFAULT 'EFFECTIVE'`,
		`ALTER TABLE IF EXISTS cases ADD COLUMN IF NOT EXISTS subject_snapshot TEXT`,
		`ALTER TABLE IF EXISTS cases ADD COLUMN IF NOT EXISTS pending_subject_revision_id VARCHAR(36)`,
		`ALTER TABLE IF EXISTS cases ADD COLUMN IF NOT EXISTS conflict_check_id VARCHAR(100)`,
		`ALTER TABLE IF EXISTS cases ADD COLUMN IF NOT EXISTS conflict_coverage_status VARCHAR(30) NOT NULL DEFAULT 'COVERAGE_LIMITED'`,

		`ALTER TABLE IF EXISTS conflict_search_scopes ADD COLUMN IF NOT EXISTS evidence_reference TEXT`,
		`ALTER TABLE IF EXISTS conflict_search_scopes ADD COLUMN IF NOT EXISTS index_run_id VARCHAR(100)`,
		`CREATE INDEX IF NOT EXISTS idx_conflict_search_scopes_index_run ON conflict_search_scopes (index_run_id)`,
		`ALTER TABLE IF EXISTS case_subject_revisions ADD COLUMN IF NOT EXISTS reason TEXT`,
		`ALTER TABLE IF EXISTS case_subject_revisions ADD COLUMN IF NOT EXISTS effective_at TIMESTAMP`,
		`ALTER TABLE IF EXISTS compliance_audit_events ADD COLUMN IF NOT EXISTS integrity_hash VARCHAR(64)`,
		`ALTER TABLE IF EXISTS approval_requests ADD COLUMN IF NOT EXISTS conflict_check_id VARCHAR(100)`,
		`ALTER TABLE IF EXISTS approval_requests ADD COLUMN IF NOT EXISTS conflict_result JSONB`,
		`ALTER TABLE IF EXISTS approval_requests ADD COLUMN IF NOT EXISTS case_created BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE IF EXISTS approval_requests ADD COLUMN IF NOT EXISTS created_case_id VARCHAR(36)`,
		`ALTER TABLE IF EXISTS approval_requests ADD COLUMN IF NOT EXISTS case_creation_status VARCHAR(30)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("升级生产 schema 兼容字段失败: %w", err)
		}
	}
	return nil
}

func installProductionAppendOnlyGuards(db *gorm.DB) error {
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS uq_conflict_reviews_check_id ON conflict_reviews (check_id)`).Error; err != nil {
		return fmt.Errorf("建立冲突复核单次结论约束失败: %w", err)
	}
	if err := db.Exec(`
CREATE OR REPLACE FUNCTION law_oa_reject_append_only_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
	RAISE EXCEPTION 'append-only evidence cannot be updated or deleted: %', TG_TABLE_NAME;
END;
$$`).Error; err != nil {
		return fmt.Errorf("安装 append-only 证据函数失败: %w", err)
	}
	for _, table := range []string{"compliance_audit_events", "conflict_reviews", "case_subject_revisions"} {
		triggerName := "trg_" + table + "_append_only"
		if err := db.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerName, table)).Error; err != nil {
			return fmt.Errorf("刷新 %s 保护触发器失败: %w", table, err)
		}
		statement := fmt.Sprintf("CREATE TRIGGER %s BEFORE UPDATE OR DELETE ON %s FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation()", triggerName, table)
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("安装 %s append-only 保护失败: %w", table, err)
		}
	}
	for _, table := range []string{"conflict_subject_versions", "conflict_subject_identifiers", "conflict_match_evidence_v2"} {
		triggerName := "trg_" + table + "_append_only"
		if err := db.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerName, table)).Error; err != nil {
			return fmt.Errorf("刷新 %s 保护触发器失败: %w", table, err)
		}
		statement := fmt.Sprintf("CREATE TRIGGER %s BEFORE UPDATE OR DELETE ON %s FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation()", triggerName, table)
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("安装 %s append-only 保护失败: %w", table, err)
		}
	}
	for _, table := range []struct {
		name      string
		table     string
		operation string
	}{
		{name: "trg_conflict_checks_no_delete", table: "conflict_checks", operation: "DELETE"},
		{name: "trg_conflict_details_append_only", table: "conflict_details", operation: "UPDATE OR DELETE"},
		{name: "trg_conflict_check_records_no_delete", table: "conflict_check_records", operation: "DELETE"},
		{name: "trg_conflict_cases_no_delete", table: "conflict_cases", operation: "DELETE"},
		{name: "trg_conflict_reviewer_assignments_no_delete", table: "conflict_reviewer_assignments", operation: "DELETE"},
		{name: "trg_waiver_applications_no_delete", table: "waiver_applications", operation: "DELETE"},
		{name: "trg_waiver_approval_records_append_only", table: "waiver_approval_records", operation: "UPDATE OR DELETE"},
		{name: "trg_waiver_signatures_append_only", table: "waiver_signatures", operation: "UPDATE OR DELETE"},
		{name: "trg_waiver_monitoring_records_append_only", table: "waiver_monitoring_records", operation: "UPDATE OR DELETE"},
		{name: "trg_conflict_index_build_runs_no_delete", table: "conflict_index_build_runs", operation: "DELETE"},
	} {
		if err := db.Exec(fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", table.name, table.table)).Error; err != nil {
			return fmt.Errorf("刷新 %s 保护触发器失败: %w", table.table, err)
		}
		statement := fmt.Sprintf("CREATE TRIGGER %s BEFORE %s ON %s FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation()", table.name, table.operation, table.table)
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("安装 %s append-only 保护失败: %w", table.table, err)
		}
	}
	// Older waiver migrations declared ON DELETE CASCADE. Replace those
	// constraints with RESTRICT so deleting a parent can never erase waiver
	// evidence. This is idempotent and scoped to known child tables.
	if err := db.Exec(`
DO $$
DECLARE
	constraint_row RECORD;
	constraint_definition TEXT;
BEGIN
	FOR constraint_row IN
		SELECT c.oid, c.conname, c.conrelid::regclass AS table_name
		FROM pg_constraint c
		WHERE c.contype = 'f'
		  AND c.confdeltype = 'c'
		  AND c.conrelid::regclass::text IN ('waiver_approval_records', 'waiver_signatures', 'waiver_monitoring_records')
	LOOP
		constraint_definition := replace(pg_get_constraintdef(constraint_row.oid), ' ON DELETE CASCADE', ' ON DELETE RESTRICT');
		EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', constraint_row.table_name, constraint_row.conname);
		EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I %s', constraint_row.table_name, constraint_row.conname, constraint_definition);
	END LOOP;
END $$`).Error; err != nil {
		return fmt.Errorf("修正豁免证据级联删除约束失败: %w", err)
	}
	return nil
}

func productionSchemaModels() []interface{} {
	return []interface{}{
		// Identity, client, case and RBAC core.
		&models.User{}, &models.Client{}, &models.Case{},
		&models.Role{}, &models.Permission{}, &models.RolePermission{}, &models.UserRole{},
		&models.Entity{}, &models.EntityRelation{}, &models.EntityNameHistory{}, &models.CaseParty{},
		&productionDepartment{},
		&productionSchemaState{},

		// Conflict detection and P0 review/recheck evidence.
		&models.ConflictCheck{}, &models.ConflictDetail{}, &models.ConflictReview{},
		&models.ConflictCase{}, &models.ConflictRule{}, &models.ConflictCheckRecord{}, &models.ClientRelation{},
		&models.ConflictSearchScope{}, &models.CaseSubjectRevision{}, &models.ComplianceAuditEvent{},
		&models.ConflictIndexBuildRun{},
		&models.ConflictReviewerAssignment{},
		&models.ConflictSubjectVersion{}, &models.ConflictSubjectIdentifier{}, &models.ConflictMatchEvidenceV2{},
		&models.WaiverApplication{}, &models.WaiverApprovalRecord{}, &models.WaiverSignature{}, &models.WaiverMonitoringRecord{},
		&models.LawyerConflictPool{},
		&models.ConflictReport{}, &models.CompanyAPICall{}, &models.ConflictScanJob{},
		// ConflictRuleExecution belongs to the legacy enhanced model set and
		// still declares a MySQL-only enum tag. Keep it out of the PostgreSQL
		// production bootstrap until that model is ported to portable checks.

		// Approval and controlled case creation.
		&models.ApprovalRequest{}, &models.ApprovalWorkflow{}, &models.ApprovalRecord{},
		&models.ApprovalNode{}, &models.ApprovalNotification{}, &models.ApprovalDelegation{},
		&productionApprovalTemplate{}, &models.ConflictCheckAssociation{}, &models.CaseCreationAssociation{},
		&models.ApprovalIntegrationMetadata{},
		&repositories.ConflictAssociation{}, &repositories.CaseCreationTracking{}, &repositories.IntegrationConfig{},

		// Intake, snapshots, documents and isolation evidence.
		&productionCaseIntake{}, &productionCaseIntakeParty{}, &productionCaseMaterial{},
		&productionApprovalSnapshot{}, &models.Document{}, &models.DocumentLock{},
		&models.DocumentVersionNew{}, &models.DocumentIndexQueue{}, &models.CaseFolder{},
		&models.CaseFolderTemplate{}, &models.InboxItem{}, &models.InboxReminderRule{},
		&models.CaseEthicalWallWhitelist{}, &models.EthicalWallAccessLog{},
		&models.NotificationQueue{}, &models.NotificationTemplate{}, &models.Notification{},
		&models.SensitiveWord{}, &models.ContentFilterLog{},
		&models.ClientTrustAccount{}, &models.ClientTrustTransaction{},
		&models.OffboardingRecord{}, &models.OffboardingTransferDetail{}, &models.TokenRevocationLog{},
		&models.DataImportTask{}, &models.DataImportError{}, &models.SystemConfig{},
		&models.OperationLog{},

		// Legal search and analytics are PostgreSQL-compatible standard models;
		// they are included because their routes are enabled in the production
		// router even when the conflict MVP is the primary workflow.
		&models.LegalStatute{}, &models.LegalCategory{}, &models.LegalHierarchy{},
		&models.LegalStatuteVersion{}, &models.UserLegalFavorite{}, &models.LegalSearchHistory{},
		&models.LegalTag{}, &models.LegalStatuteTag{},
		&models.AnalyticsUserSession{}, &models.PageView{}, &models.AnalyticsUserEvent{},
		&models.UserJourney{}, &models.UserSegment{}, &models.UserSegmentMembership{},
		&models.BehaviorPattern{}, &models.UserBehaviorRecord{}, &models.FunnelAnalysis{},
		&models.RetentionAnalysis{}, &models.AnalyticsReportData{}, &models.HeatmapData{},
		&models.ClickstreamRecord{}, &models.FormInteraction{}, &models.PerformanceMetric{},
		&models.SearchEvent{}, &models.ConversionEvent{}, &models.RealTimeStats{},

		// Finance tables are part of the current router surface and are harmless
		// to create even when the firm keeps finance disabled for the MVP.
		&models.Contract{}, &models.PaymentMilestone{}, &models.Invoice{}, &models.Payment{},
		&models.BadDebtRecord{}, &models.CommissionRecord{}, &models.CommissionRule{}, &models.FeeTemplate{},
	}
}
