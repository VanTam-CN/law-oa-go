package database

import (
	"fmt"
	"gorm.io/gorm"
)

// Explicit additive upgrade, equivalent to migrations 000080-000084. Existing
// snapshots and inbox text are retained; old NULL modes are resolved at read time.
func ensureIntakeCollaborationSchema(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE case_intakes ADD COLUMN IF NOT EXISTS approval_claim_token VARCHAR(64)`,
		`ALTER TABLE approval_case_creation_tracking ADD COLUMN IF NOT EXISTS intake_id VARCHAR(36)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_case_creation_tracking_intake_unique ON approval_case_creation_tracking(intake_id)`,
		`ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS active_submission_snapshot_id UUID`,
		`ALTER TABLE approval_records ADD COLUMN IF NOT EXISTS submission_snapshot_id UUID`,
		`ALTER TABLE approval_records ADD COLUMN IF NOT EXISTS action_version VARCHAR(100) NOT NULL DEFAULT ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_intake_submission_round ON approval_snapshots(approval_request_id,snapshot_type,source_version) WHERE snapshot_type = 'intake_submission'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_intake_submission_return ON approval_snapshots(approval_request_id,(snapshot_data->>'return_record_id')) WHERE snapshot_type = 'intake_submission' AND COALESCE(snapshot_data->>'return_record_id','') <> ''`,
		`ALTER TABLE inbox_items ADD COLUMN IF NOT EXISTS source_ref VARCHAR(100), ADD COLUMN IF NOT EXISTS completion_mode VARCHAR(20), ADD COLUMN IF NOT EXISTS workflow_task_key VARCHAR(200), ADD COLUMN IF NOT EXISTS workflow_group VARCHAR(150), ADD COLUMN IF NOT EXISTS workflow_purpose VARCHAR(40), ADD COLUMN IF NOT EXISTS workflow_version VARCHAR(100), ADD COLUMN IF NOT EXISTS closed_reason VARCHAR(30), ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_inbox_workflow_task_key ON inbox_items(workflow_task_key)`,
		`CREATE INDEX IF NOT EXISTS idx_inbox_workflow_group ON inbox_items(workflow_group)`,
		`CREATE INDEX IF NOT EXISTS idx_inbox_owner_archive ON inbox_items(user_id,archived_at)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_client_creation_scope ON client_creation_requests(operator_id,operation,key_digest)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("接案协作结构升级失败: %w", err)
		}
	}
	constraints := []struct{ table, name, body string }{
		{"inbox_items", "ck_inbox_workflow_contract", `CHECK (completion_mode IS DISTINCT FROM 'workflow' OR (NULLIF(BTRIM(source_ref),'') IS NOT NULL AND NULLIF(BTRIM(workflow_task_key),'') IS NOT NULL AND NULLIF(BTRIM(workflow_group),'') IS NOT NULL AND NULLIF(BTRIM(workflow_purpose),'') IS NOT NULL AND NULLIF(BTRIM(workflow_version),'') IS NOT NULL))`},
		{"inbox_items", "ck_inbox_workflow_key", `CHECK (workflow_task_key IS NULL OR LENGTH(BTRIM(workflow_task_key)) > 0)`},
		{"client_creation_requests", "ck_client_creation_result", `CHECK ((status='pending' AND client_id IS NULL) OR (status='completed' AND client_id IS NOT NULL))`},
		{"client_creation_requests", "ck_client_creation_digests", `CHECK (length(key_digest)=64 AND length(request_digest)=64)`},
		{"client_creation_requests", "ck_client_creation_operator", `CHECK (operator_id > 0)`},
		{"client_creation_requests", "fk_client_creation_result", `FOREIGN KEY(client_id) REFERENCES clients(id)`},
	}
	for _, constraint := range constraints {
		var present bool
		if err := db.Raw(`SELECT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid = ?::regclass AND conname = ?)`, constraint.table, constraint.name).Scan(&present).Error; err != nil {
			return err
		}
		if !present {
			if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", constraint.table, constraint.name, constraint.body)).Error; err != nil {
				return err
			}
		}
	}
	guards := []string{`CREATE OR REPLACE FUNCTION guard_intake_submission_snapshot_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.snapshot_type = 'intake_submission' OR (TG_OP = 'UPDATE' AND NEW.snapshot_type = 'intake_submission') THEN
  RAISE EXCEPTION 'intake submission snapshots are append-only' USING ERRCODE = '55000';
 END IF;
 IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;`,
		`DROP TRIGGER IF EXISTS trg_intake_submission_snapshot_immutable ON approval_snapshots;`,
		`CREATE TRIGGER trg_intake_submission_snapshot_immutable BEFORE UPDATE OR DELETE ON approval_snapshots
 FOR EACH ROW EXECUTE FUNCTION guard_intake_submission_snapshot_immutable();`,
		`CREATE OR REPLACE FUNCTION guard_intake_submission_decision_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.submission_snapshot_id IS NOT NULL OR (TG_OP = 'UPDATE' AND NEW.submission_snapshot_id IS NOT NULL) THEN
  RAISE EXCEPTION 'versioned approval decisions are append-only' USING ERRCODE = '55000';
 END IF;
 IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;`,
		`DROP TRIGGER IF EXISTS trg_intake_submission_decision_immutable ON approval_records;`,
		`CREATE TRIGGER trg_intake_submission_decision_immutable BEFORE UPDATE OR DELETE ON approval_records
 FOR EACH ROW EXECUTE FUNCTION guard_intake_submission_decision_immutable();`}
	for _, guard := range guards {
		if err := db.Exec(guard).Error; err != nil {
			return fmt.Errorf("安装提交审计不可变保护失败: %w", err)
		}
	}
	return nil
}

// Do not let IF NOT EXISTS hide a malformed pre-existing claim column or slot
// index. This is also called by the runtime startup gate, before HTTP traffic.
func validateIntakeApprovalClaimContract(db *gorm.DB) error {
	var count int64
	if err := db.Raw(`SELECT count(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'case_intakes'
		AND column_name = 'approval_claim_token' AND data_type = 'character varying'
		AND character_maximum_length = 64 AND is_nullable = 'YES'`).Scan(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("case_intakes.approval_claim_token 必须为可空 VARCHAR(64)；请执行 bootstrap 或迁移 000080")
	}
	if err := db.Raw(`SELECT count(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'approval_case_creation_tracking'
		AND column_name = 'intake_id' AND is_nullable = 'YES'`).Scan(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("approval_case_creation_tracking.intake_id 必须可空，以保留普通审批语义")
	}
	if err := db.Raw(`SELECT count(*) FROM pg_index i
		JOIN pg_class idx ON idx.oid = i.indexrelid
		JOIN pg_class tbl ON tbl.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = tbl.relnamespace
		JOIN pg_attribute a ON a.attrelid = tbl.oid AND a.attnum = i.indkey[0]
		WHERE n.nspname = current_schema() AND tbl.relname = 'approval_case_creation_tracking'
		AND idx.relname = 'idx_case_creation_tracking_intake_unique'
		AND i.indisunique AND i.indisvalid AND i.indisready AND i.indimmediate
		AND i.indnatts = 1 AND i.indpred IS NULL AND i.indexprs IS NULL
		AND a.attname = 'intake_id'
		AND pg_get_indexdef(i.indexrelid) NOT LIKE '%NULLS NOT DISTINCT%'`).Scan(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("approval_case_creation_tracking 缺少完整 intake_id 唯一占槽索引；请执行 bootstrap 或迁移 000081")
	}
	return nil
}
