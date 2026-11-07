-- Temporary databases only. Never discard live submission history to downgrade.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM approval_snapshots WHERE snapshot_type = 'intake_submission')
 OR EXISTS (SELECT 1 FROM approval_records WHERE submission_snapshot_id IS NOT NULL) THEN
  RAISE EXCEPTION 'intake submission history exists; use a compatible reader or forward repair';
 END IF;
END $$;
DROP TRIGGER IF EXISTS trg_intake_submission_decision_immutable ON approval_records;
DROP FUNCTION IF EXISTS guard_intake_submission_decision_immutable();
DROP TRIGGER IF EXISTS trg_intake_submission_snapshot_immutable ON approval_snapshots;
DROP FUNCTION IF EXISTS guard_intake_submission_snapshot_immutable();
DROP INDEX IF EXISTS idx_intake_submission_return;
DROP INDEX IF EXISTS idx_intake_submission_round;
ALTER TABLE approval_records DROP COLUMN IF EXISTS action_version;
ALTER TABLE approval_records DROP COLUMN IF EXISTS submission_snapshot_id;
ALTER TABLE approval_requests DROP COLUMN IF EXISTS active_submission_snapshot_id;
