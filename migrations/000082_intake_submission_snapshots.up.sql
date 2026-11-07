-- Additive: historical snapshots are neither selected by time nor rewritten.
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS active_submission_snapshot_id UUID;
ALTER TABLE approval_records ADD COLUMN IF NOT EXISTS submission_snapshot_id UUID;
ALTER TABLE approval_records ADD COLUMN IF NOT EXISTS action_version VARCHAR(100) NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_intake_submission_round
 ON approval_snapshots (approval_request_id, snapshot_type, source_version)
 WHERE snapshot_type = 'intake_submission';
-- A return can produce only one new submission, even across racing retries.
CREATE UNIQUE INDEX IF NOT EXISTS idx_intake_submission_return
 ON approval_snapshots (approval_request_id, (snapshot_data->>'return_record_id'))
 WHERE snapshot_type = 'intake_submission' AND COALESCE(snapshot_data->>'return_record_id', '') <> '';

-- SQL-level guards are required because Table(...).Updates bypasses GORM hooks.
CREATE OR REPLACE FUNCTION guard_intake_submission_snapshot_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.snapshot_type = 'intake_submission' OR (TG_OP = 'UPDATE' AND NEW.snapshot_type = 'intake_submission') THEN
  RAISE EXCEPTION 'intake submission snapshots are append-only' USING ERRCODE = '55000';
 END IF;
 IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS trg_intake_submission_snapshot_immutable ON approval_snapshots;
CREATE TRIGGER trg_intake_submission_snapshot_immutable BEFORE UPDATE OR DELETE ON approval_snapshots
 FOR EACH ROW EXECUTE FUNCTION guard_intake_submission_snapshot_immutable();
CREATE OR REPLACE FUNCTION guard_intake_submission_decision_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.submission_snapshot_id IS NOT NULL OR (TG_OP = 'UPDATE' AND NEW.submission_snapshot_id IS NOT NULL) THEN
  RAISE EXCEPTION 'versioned approval decisions are append-only' USING ERRCODE = '55000';
 END IF;
 IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
 RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS trg_intake_submission_decision_immutable ON approval_records;
CREATE TRIGGER trg_intake_submission_decision_immutable BEFORE UPDATE OR DELETE ON approval_records
 FOR EACH ROW EXECUTE FUNCTION guard_intake_submission_decision_immutable();
